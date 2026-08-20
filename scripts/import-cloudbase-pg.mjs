#!/usr/bin/env node

import { spawn, spawnSync } from "node:child_process";
import { createInterface } from "node:readline";
import { resolve } from "node:path";

const envId = process.env.CLOUDBASE_ENV_ID || "chrms-d9gywgbw57877b4fb";
const databasePath = resolve(process.cwd(), "backend/data/hrms.db");
const tables = [
  "certificate_catalogs",
  "talents",
  "certificates",
  "companies",
  "company_requirements",
  "delivery_orders",
  "delivery_order_talents",
  "contracts",
  "reminders",
  "system_settings",
  "audit_logs",
];
const nullableColumns = new Set(["years_of_experience", "handled_at"]);
const booleanColumns = new Set(["is_enabled", "is_available"]);
const numericColumns = new Set([
  "sort_order",
  "quantity",
  "talent_quote",
  "performance_amount",
  "received_amount",
  "paid_amount",
  "company_rebate",
  "direct_payment",
  "performance_total",
  "received_total",
  "paid_total",
  "direct_payment_total",
]);

function loadRows(table) {
  const result = spawnSync("sqlite3", ["-json", databasePath, `SELECT * FROM ${table}`], {
    encoding: "utf8",
  });
  if (result.status !== 0) {
    throw new Error(`read ${table}: ${result.stderr.trim()}`);
  }
  return JSON.parse(result.stdout || "[]");
}

function sqlValue(column, value) {
  if (value === null || value === undefined) {
    return nullableColumns.has(column) ? "NULL" : "''";
  }
  if (booleanColumns.has(column)) {
    return Number(value) === 0 ? "false" : "true";
  }
  if (numericColumns.has(column)) {
    const numeric = Number(value);
    return Number.isFinite(numeric) ? String(numeric) : "0";
  }
  if (typeof value === "number") {
    return Number.isFinite(value) ? String(value) : "0";
  }
  return `'${String(value).replaceAll("'", "''")}'`;
}

function insertSQL(table, row) {
  const columns = Object.keys(row);
  const quotedColumns = columns.map((column) => `"${column}"`).join(", ");
  const values = columns.map((column) => sqlValue(column, row[column])).join(", ");
  const conflictTarget = table === "system_settings" ? "key" : "id";
  return `INSERT INTO public.${table} (${quotedColumns}) VALUES (${values}) ON CONFLICT (${conflictTarget}) DO NOTHING`;
}

class CloudBaseMCP {
  constructor() {
    this.nextID = 1;
    this.pending = new Map();
    this.process = spawn("npx", ["-y", "@cloudbase/cloudbase-mcp@latest"], {
      stdio: ["pipe", "pipe", "inherit"],
    });
    this.process.once("error", (error) => this.rejectAll(error));
    this.process.once("exit", (code) => this.rejectAll(new Error(`CloudBase MCP exited with code ${code}`)));
    const output = createInterface({ input: this.process.stdout });
    output.on("line", (line) => this.handleLine(line));
  }

  rejectAll(error) {
    for (const { reject } of this.pending.values()) reject(error);
    this.pending.clear();
  }

  handleLine(line) {
    let message;
    try {
      message = JSON.parse(line);
    } catch {
      return;
    }
    const waiter = this.pending.get(message.id);
    if (!waiter) return;
    this.pending.delete(message.id);
    if (message.error) {
      waiter.reject(new Error(message.error.message || JSON.stringify(message.error)));
      return;
    }
    waiter.resolve(message.result);
  }

  request(method, params) {
    const id = this.nextID++;
    const message = { jsonrpc: "2.0", id, method, params };
    return new Promise((resolveRequest, reject) => {
      this.pending.set(id, { resolve: resolveRequest, reject });
      this.process.stdin.write(`${JSON.stringify(message)}\n`);
    });
  }

  async start() {
    await this.request("initialize", {
      protocolVersion: "2024-11-05",
      capabilities: {},
      clientInfo: { name: "hrms-pg-import", version: "1.0" },
    });
  }

  async execute(sql) {
    const result = await this.request("tools/call", {
      name: "managePgDatabase",
      arguments: { action: "execute", envId, sql, confirm: true },
    });
    const text = result.content?.find((item) => item.type === "text")?.text || "";
    if (!text.includes('"success": true')) {
      throw new Error(text || "CloudBase rejected database mutation");
    }
  }

  close() {
    this.process.kill();
  }
}

const client = new CloudBaseMCP();
try {
  await client.start();
  for (const table of tables) {
    const rows = loadRows(table);
    for (const row of rows) await client.execute(insertSQL(table, row));
    console.log(`${table}: ${rows.length} rows processed`);
  }
} finally {
  client.close();
}
