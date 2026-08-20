// Package cloudbasepg keeps CloudBase PostgreSQL as the durable store for the API.
package cloudbasepg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
)

var tableOrder = []string{
	"admins", "certificate_catalogs", "talents", "certificates", "companies",
	"company_requirements", "delivery_orders", "delivery_order_talents", "contracts",
	"reminders", "system_settings", "audit_logs",
}

// Client accesses the documented CloudBase PG PostgREST endpoint with a server-only API key.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New builds a client for a CloudBase environment. The API key must never be exposed to browsers.
func New(envID, apiKey, baseURL string) (*Client, error) {
	if strings.TrimSpace(envID) == "" {
		return nil, errors.New("CloudBase environment ID is required")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://" + strings.TrimSpace(envID) + ".api.tcloudbasegateway.com/v1/rdb/rest"
	}
	return NewWithBaseURL(baseURL, apiKey, nil)
}

// NewWithBaseURL exists for tests and private-gateway deployments.
func NewWithBaseURL(baseURL, apiKey string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("CloudBase API key is required")
	}
	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New("CloudBase PG REST base URL is required")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), apiKey: apiKey, httpClient: httpClient}, nil
}

// FetchTable reads all rows from one known business table.
func (c *Client) FetchTable(ctx context.Context, table string) ([]map[string]any, error) {
	endpoint, err := c.tableURL(table, url.Values{"select": {"*"}})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	if err := c.doJSON(request, &rows); err != nil {
		return nil, fmt.Errorf("fetch %s: %w", table, err)
	}
	return rows, nil
}

// Create inserts a single row. One row per request avoids the mixed-key batch limitation.
func (c *Client) Create(ctx context.Context, table string, row map[string]any) error {
	return c.write(ctx, http.MethodPost, table, nil, row)
}

// Update replaces the supplied fields of one row selected by its primary key.
func (c *Client) Update(ctx context.Context, table, key string, row map[string]any) error {
	return c.write(ctx, http.MethodPatch, table, keyQuery(table, key), row)
}

// Delete removes one row selected by its primary key.
func (c *Client) Delete(ctx context.Context, table, key string) error {
	return c.write(ctx, http.MethodDelete, table, keyQuery(table, key), nil)
}

// ReplaceTable is retained for controlled imports and tests. Runtime writes use ApplyChanges.
func (c *Client) ReplaceTable(ctx context.Context, table string, rows []map[string]any) error {
	endpoint, err := c.tableURL(table, nil)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	if err := c.doJSON(request, nil); err != nil {
		return fmt.Errorf("clear %s: %w", table, err)
	}
	for _, row := range rows {
		if err := c.Create(ctx, table, row); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) write(ctx context.Context, method, table string, query url.Values, body any) error {
	endpoint, err := c.tableURL(table, query)
	if err != nil {
		return err
	}
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode %s payload: %w", table, err)
		}
		payload = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, payload)
	if err != nil {
		return err
	}
	request.Header.Set("Prefer", "return=minimal")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if err := c.doJSON(request, nil); err != nil {
		return fmt.Errorf("%s %s: %w", method, table, err)
	}
	return nil
}

func (c *Client) doJSON(request *http.Request, target any) error {
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return err
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("CloudBase PG returned HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	if target != nil && len(body) > 0 {
		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("decode CloudBase PG response: %w", err)
		}
	}
	return nil
}

func (c *Client) tableURL(table string, query url.Values) (string, error) {
	if !knownTable(table) {
		return "", fmt.Errorf("unknown CloudBase table %q", table)
	}
	endpoint := c.baseURL + "/" + table
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	return endpoint, nil
}

func knownTable(table string) bool {
	for _, candidate := range tableOrder {
		if candidate == table {
			return true
		}
	}
	return false
}

func primaryKey(table string) string {
	if table == "system_settings" {
		return "key"
	}
	return "id"
}

func keyQuery(table, key string) url.Values {
	return url.Values{primaryKey(table): {"eq." + key}}
}

// Snapshot is a database-independent view used to calculate per-row mutations.
type Snapshot map[string]map[string]map[string]any

// TakeSnapshot reads the in-memory working set without relying on model JSON tags.
func TakeSnapshot(db *gorm.DB) (Snapshot, error) {
	snapshot := make(Snapshot, len(tableOrder))
	for _, table := range tableOrder {
		var rows []map[string]any
		if err := db.Table(table).Find(&rows).Error; err != nil {
			return nil, fmt.Errorf("snapshot %s: %w", table, err)
		}
		byKey := make(map[string]map[string]any, len(rows))
		keyName := primaryKey(table)
		for _, row := range rows {
			normalized := normalizeRow(row)
			key, ok := normalized[keyName].(string)
			if !ok || key == "" {
				return nil, fmt.Errorf("snapshot %s row has no %s", table, keyName)
			}
			byKey[key] = normalized
		}
		snapshot[table] = byKey
	}
	return snapshot, nil
}

// Load replaces the ephemeral SQLite working set with CloudBase PG data at process start.
func (c *Client) Load(ctx context.Context, db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for index := len(tableOrder) - 1; index >= 0; index-- {
			if err := tx.Exec("DELETE FROM " + tableOrder[index]).Error; err != nil {
				return err
			}
		}
		for _, table := range tableOrder {
			rows, err := c.FetchTable(ctx, table)
			if err != nil {
				return err
			}
			for _, row := range rows {
				if err := tx.Table(table).Create(normalizeRow(row)).Error; err != nil {
					return fmt.Errorf("load %s: %w", table, err)
				}
			}
		}
		return nil
	})
}

// ApplyChanges persists only rows changed by the current successful request.
func (c *Client) ApplyChanges(ctx context.Context, before, after Snapshot) error {
	for _, table := range tableOrder {
		if err := c.applyCreatesAndUpdates(ctx, table, before[table], after[table]); err != nil {
			return err
		}
	}
	for index := len(tableOrder) - 1; index >= 0; index-- {
		table := tableOrder[index]
		for key := range before[table] {
			if _, exists := after[table][key]; !exists {
				if err := c.Delete(ctx, table, key); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *Client) applyCreatesAndUpdates(ctx context.Context, table string, before, after map[string]map[string]any) error {
	for key, row := range after {
		previous, exists := before[key]
		if !exists {
			if err := c.Create(ctx, table, row); err != nil {
				return err
			}
			continue
		}
		if !rowsEqual(previous, row) {
			if err := c.Update(ctx, table, key, row); err != nil {
				return err
			}
		}
	}
	return nil
}

func rowsEqual(left, right map[string]any) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return bytes.Equal(leftJSON, rightJSON)
}

func normalizeRow(row map[string]any) map[string]any {
	normalized := make(map[string]any, len(row))
	for key, value := range row {
		switch typed := value.(type) {
		case []byte:
			normalized[key] = string(typed)
		case time.Time:
			normalized[key] = typed.UTC().Format(time.RFC3339Nano)
		case *time.Time:
			if typed != nil {
				normalized[key] = typed.UTC().Format(time.RFC3339Nano)
			} else {
				normalized[key] = nil
			}
		case int64:
			if key == "is_enabled" || key == "is_available" {
				normalized[key] = typed != 0
			} else {
				normalized[key] = typed
			}
		case int:
			if key == "is_enabled" || key == "is_available" {
				normalized[key] = typed != 0
			} else {
				normalized[key] = typed
			}
		default:
			normalized[key] = value
		}
	}
	return normalized
}
