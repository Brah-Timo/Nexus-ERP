-- 0014_settings.sql — idempotent Settings module migration
-- NOTE: NO schema_migrations entry - handled by the application

-- ─── Extend companies ─────────────────────────────────────────────────────────
ALTER TABLE companies ADD COLUMN IF NOT EXISTS logo_url       TEXT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS website        TEXT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS email          TEXT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS phone          TEXT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS address        TEXT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS city           TEXT;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS country        TEXT DEFAULT 'DZ';
ALTER TABLE companies ADD COLUMN IF NOT EXISTS timezone       TEXT DEFAULT 'Africa/Algiers';
ALTER TABLE companies ADD COLUMN IF NOT EXISTS default_currency TEXT DEFAULT 'DZD';
ALTER TABLE companies ADD COLUMN IF NOT EXISTS fiscal_year_start INT DEFAULT 1;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS is_active      BOOLEAN DEFAULT TRUE;
ALTER TABLE companies ADD COLUMN IF NOT EXISTS updated_at     TIMESTAMPTZ;

-- ─── Roles ────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS roles (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name         TEXT NOT NULL,
  description  TEXT,
  permissions  JSONB NOT NULL DEFAULT '[]',
  is_system    BOOLEAN NOT NULL DEFAULT FALSE,
  is_active    BOOLEAN NOT NULL DEFAULT TRUE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ,
  UNIQUE(company_id, name)
);

-- ─── Extend users ─────────────────────────────────────────────────────────────
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id       UUID REFERENCES roles(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS full_name     TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone         TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url    TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login    TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active     BOOLEAN DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at    TIMESTAMPTZ;

-- ─── Fiscal Years ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS fiscal_years (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name         TEXT NOT NULL,
  start_date   DATE NOT NULL,
  end_date     DATE NOT NULL,
  status       TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open','closed','locked')),
  is_current   BOOLEAN NOT NULL DEFAULT FALSE,
  closed_at    TIMESTAMPTZ,
  closed_by    UUID REFERENCES users(id),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Currencies ───────────────────────────────────────────────────────────────
-- Drop and recreate currencies table with correct schema
DROP TABLE IF EXISTS currencies CASCADE;

CREATE TABLE currencies (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id    UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  code          TEXT NOT NULL,
  name          TEXT NOT NULL,
  symbol        TEXT NOT NULL DEFAULT '',
  exchange_rate NUMERIC(18,6) NOT NULL DEFAULT 1.0,
  is_base       BOOLEAN NOT NULL DEFAULT FALSE,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  updated_at    TIMESTAMPTZ,
  UNIQUE(company_id, code)
);

-- Seed DZD for every existing company
INSERT INTO currencies (company_id, code, name, symbol, exchange_rate, is_base, is_active)
SELECT id, 'DZD', 'Dinar Algérien', 'DA', 1.0, TRUE, TRUE
FROM companies
ON CONFLICT (company_id, code) DO NOTHING;

-- ─── Numbering Config ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS numbering_config (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  doc_type     TEXT NOT NULL,
  prefix       TEXT NOT NULL DEFAULT '',
  suffix       TEXT NOT NULL DEFAULT '',
  next_number  INT  NOT NULL DEFAULT 1,
  padding      INT  NOT NULL DEFAULT 4,
  reset_yearly BOOLEAN NOT NULL DEFAULT TRUE,
  UNIQUE(company_id, doc_type)
);

-- seed defaults for existing companies
INSERT INTO numbering_config (company_id, doc_type, prefix, next_number)
SELECT id, unnest(ARRAY['invoice','purchase_order','quotation','receipt','payment','cheque','expense']),
       unnest(ARRAY['INV-','PO-','QT-','RCT-','PAY-','CHQ-','EXP-']), 1
FROM companies
ON CONFLICT (company_id, doc_type) DO NOTHING;

-- ─── Taxes ────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS taxes (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name         TEXT NOT NULL,
  code         TEXT NOT NULL,
  tax_type     TEXT NOT NULL DEFAULT 'percentage' CHECK (tax_type IN ('percentage','fixed')),
  rate         NUMERIC(10,4) NOT NULL DEFAULT 0,
  account_id   UUID,
  is_active    BOOLEAN NOT NULL DEFAULT TRUE,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, code)
);

-- ─── Audit Log ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS audit_log (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id   UUID REFERENCES companies(id) ON DELETE CASCADE,
  user_id      UUID REFERENCES users(id) ON DELETE SET NULL,
  action       TEXT NOT NULL,
  table_name   TEXT NOT NULL,
  record_id    TEXT,
  old_values   JSONB,
  new_values   JSONB,
  ip_address   TEXT,
  user_agent   TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_company ON audit_log(company_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_user    ON audit_log(user_id, created_at DESC);

-- ─── Workflow Rules ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS workflow_rules (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id    UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  doc_type      TEXT NOT NULL,
  trigger_event TEXT NOT NULL DEFAULT 'on_create',
  conditions    JSONB NOT NULL DEFAULT '[]',
  actions       JSONB NOT NULL DEFAULT '[]',
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  priority      INT NOT NULL DEFAULT 10,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ
);

-- ─── System Config ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS system_config (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  config_key   TEXT NOT NULL,
  config_value TEXT,
  description  TEXT,
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, config_key)
);

-- ─── Auto-update updated_at function ─────────────────────────────────────────
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

-- ============================================================
-- END 0014_settings.sql
-- ============================================================