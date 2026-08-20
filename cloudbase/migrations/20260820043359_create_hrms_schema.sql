
CREATE TABLE public.admins (
  id text PRIMARY KEY,
  username varchar(64) NOT NULL UNIQUE,
  password_hash varchar(255) NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE TABLE public.talents (
  id text PRIMARY KEY,
  code varchar(64) NOT NULL UNIQUE,
  name varchar(50) NOT NULL,
  id_card_number varchar(18) NOT NULL,
  gender varchar(16) NOT NULL DEFAULT '',
  birth_date varchar(10) NOT NULL DEFAULT '',
  phone varchar(32) NOT NULL,
  social_insurance_status varchar(64) NOT NULL DEFAULT '',
  native_place varchar(255) NOT NULL DEFAULT '',
  current_location varchar(255) NOT NULL DEFAULT '',
  education varchar(32) NOT NULL DEFAULT '',
  major varchar(255) NOT NULL DEFAULT '',
  years_of_experience integer,
  primary_certificate varchar(100) NOT NULL DEFAULT '',
  compensation varchar(64) NOT NULL DEFAULT '',
  bi_expires_on varchar(10) NOT NULL DEFAULT '',
  certificate_renewal_note text NOT NULL DEFAULT '',
  note text NOT NULL DEFAULT '',
  status varchar(32) NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE INDEX idx_talents_id_card_number ON public.talents (id_card_number);
CREATE INDEX idx_talents_current_location ON public.talents (current_location);
CREATE INDEX idx_talents_bi_expires_on ON public.talents (bi_expires_on);
CREATE INDEX idx_talents_status ON public.talents (status);
CREATE INDEX idx_talents_updated_at ON public.talents (updated_at DESC);
CREATE TABLE public.certificate_catalogs (
  id text PRIMARY KEY,
  name varchar(100) NOT NULL UNIQUE,
  is_enabled boolean NOT NULL DEFAULT true,
  sort_order integer NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE TABLE public.certificates (
  id text PRIMARY KEY,
  talent_id text NOT NULL UNIQUE,
  catalog_id text NOT NULL DEFAULT '',
  certificate_name_snapshot varchar(100) NOT NULL,
  category varchar(32) NOT NULL,
  specialty varchar(100) NOT NULL DEFAULT '',
  certificate_number varchar(100) NOT NULL DEFAULT '',
  issuer varchar(255) NOT NULL DEFAULT '',
  issued_date varchar(10) NOT NULL DEFAULT '',
  expires_on varchar(10) NOT NULL DEFAULT '',
  registration_status varchar(32) NOT NULL,
  registered_company varchar(255) NOT NULL DEFAULT '',
  is_available boolean NOT NULL DEFAULT true,
  note text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE INDEX idx_certificates_catalog_id ON public.certificates (catalog_id);
CREATE INDEX idx_certificates_category ON public.certificates (category);
CREATE INDEX idx_certificates_expires_on ON public.certificates (expires_on);
CREATE INDEX idx_certificates_is_available ON public.certificates (is_available);
CREATE TABLE public.contracts (
  id text PRIMARY KEY,
  talent_id text NOT NULL,
  contract_number varchar(100) NOT NULL UNIQUE,
  company_name varchar(255) NOT NULL,
  contract_type varchar(32) NOT NULL,
  start_date varchar(10) NOT NULL,
  end_date varchar(10) NOT NULL,
  status varchar(32) NOT NULL,
  note text NOT NULL DEFAULT '',
  terminated_on varchar(10) NOT NULL DEFAULT '',
  termination_reason text NOT NULL DEFAULT '',
  renewed_from_contract_id text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE INDEX idx_contracts_talent_id ON public.contracts (talent_id);
CREATE INDEX idx_contracts_end_date ON public.contracts (end_date);
CREATE INDEX idx_contracts_status ON public.contracts (status);
CREATE TABLE public.reminders (
  id text PRIMARY KEY,
  reminder_type varchar(32) NOT NULL,
  source_id text NOT NULL,
  talent_id text NOT NULL,
  due_date varchar(10) NOT NULL,
  status varchar(32) NOT NULL DEFAULT 'pending',
  handled_at timestamptz,
  handled_by_admin_id text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE (reminder_type, source_id)
);
CREATE INDEX idx_reminders_talent_id ON public.reminders (talent_id);
CREATE INDEX idx_reminders_due_date ON public.reminders (due_date);
CREATE INDEX idx_reminders_status ON public.reminders (status);
CREATE TABLE public.system_settings (
  key varchar(100) PRIMARY KEY,
  value varchar(255) NOT NULL,
  updated_by_admin_id text NOT NULL DEFAULT '',
  updated_at timestamptz NOT NULL
);
CREATE TABLE public.companies (
  id text PRIMARY KEY,
  code varchar(64) NOT NULL UNIQUE,
  name varchar(255) NOT NULL,
  contact_name varchar(64) NOT NULL DEFAULT '',
  contact_phone varchar(32) NOT NULL DEFAULT '',
  owner_name varchar(64) NOT NULL DEFAULT '',
  client_type varchar(32) NOT NULL DEFAULT '',
  note text NOT NULL DEFAULT '',
  contract_attachment_name varchar(255) NOT NULL DEFAULT '',
  contract_attachment_path varchar(255) NOT NULL DEFAULT '',
  contract_expires_on varchar(10) NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE INDEX idx_companies_name ON public.companies (name);
CREATE INDEX idx_companies_contract_expires_on ON public.companies (contract_expires_on);
CREATE TABLE public.company_requirements (
  id text PRIMARY KEY,
  company_id text NOT NULL,
  specialty varchar(100) NOT NULL,
  conditions varchar(255) NOT NULL DEFAULT '',
  quantity integer NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE INDEX idx_company_requirements_company_id ON public.company_requirements (company_id);
CREATE TABLE public.delivery_orders (
  id text PRIMARY KEY,
  code varchar(64) NOT NULL UNIQUE,
  company_id text NOT NULL,
  registration_unit_name varchar(255) NOT NULL DEFAULT '',
  unit_nature varchar(64) NOT NULL DEFAULT '',
  status varchar(32) NOT NULL,
  approval_status varchar(32) NOT NULL DEFAULT 'pending',
  contract_expires_on varchar(10) NOT NULL DEFAULT '',
  performance_total double precision NOT NULL DEFAULT 0,
  received_total double precision NOT NULL DEFAULT 0,
  paid_total double precision NOT NULL DEFAULT 0,
  direct_payment_total double precision NOT NULL DEFAULT 0,
  note text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE INDEX idx_delivery_orders_company_id ON public.delivery_orders (company_id);
CREATE INDEX idx_delivery_orders_status ON public.delivery_orders (status);
CREATE INDEX idx_delivery_orders_contract_expires_on ON public.delivery_orders (contract_expires_on);
CREATE TABLE public.delivery_order_talents (
  id text PRIMARY KEY,
  delivery_order_id text NOT NULL,
  talent_id text NOT NULL,
  certificate_id text NOT NULL DEFAULT '',
  specialty varchar(100) NOT NULL DEFAULT '',
  talent_quote double precision NOT NULL DEFAULT 0,
  performance_amount double precision NOT NULL DEFAULT 0,
  received_amount double precision NOT NULL DEFAULT 0,
  paid_amount double precision NOT NULL DEFAULT 0,
  company_rebate double precision NOT NULL DEFAULT 0,
  direct_payment double precision NOT NULL DEFAULT 0,
  note text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);
CREATE INDEX idx_delivery_order_talents_order_id ON public.delivery_order_talents (delivery_order_id);
CREATE INDEX idx_delivery_order_talents_talent_id ON public.delivery_order_talents (talent_id);
CREATE INDEX idx_delivery_order_talents_certificate_id ON public.delivery_order_talents (certificate_id);
CREATE TABLE public.audit_logs (
  id text PRIMARY KEY,
  admin_id text NOT NULL DEFAULT '',
  action varchar(64) NOT NULL,
  resource_type varchar(64) NOT NULL,
  resource_id text NOT NULL DEFAULT '',
  display_name varchar(255) NOT NULL DEFAULT '',
  summary text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL
);
CREATE INDEX idx_audit_logs_admin_id ON public.audit_logs (admin_id);
CREATE INDEX idx_audit_logs_action ON public.audit_logs (action);
CREATE INDEX idx_audit_logs_resource_type ON public.audit_logs (resource_type);
CREATE INDEX idx_audit_logs_resource_id ON public.audit_logs (resource_id);
CREATE INDEX idx_audit_logs_created_at ON public.audit_logs (created_at DESC);
ALTER TABLE public.admins ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.talents ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.certificate_catalogs ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.certificates ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.contracts ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.reminders ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.system_settings ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.companies ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.company_requirements ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.delivery_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.delivery_order_talents ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.audit_logs ENABLE ROW LEVEL SECURITY;
