-- =============================================================================
-- Nexus ERP — Settings Extensions Migration
-- Version: 0004 | Extends base schema with roles, taxes, workflow, audit tables
-- =============================================================================

-- ── Roles ─────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS roles (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB        NOT NULL DEFAULT '[]',
    is_system   BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, name)
);

-- Add role_id FK to users if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='users' AND column_name='role_id'
    ) THEN
        ALTER TABLE users ADD COLUMN role_id UUID REFERENCES roles(id);
    END IF;
END$$;

-- Add phone to users if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='users' AND column_name='phone'
    ) THEN
        ALTER TABLE users ADD COLUMN phone VARCHAR(30);
    END IF;
END$$;

-- ── Currencies per company ────────────────────────────────────────────────────
-- The global currencies table exists; we add a company-level currency config table
CREATE TABLE IF NOT EXISTS company_currencies (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id    UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code          VARCHAR(10)  NOT NULL,
    name          VARCHAR(100) NOT NULL,
    symbol        VARCHAR(10)  NOT NULL DEFAULT '',
    exchange_rate NUMERIC(18,6) NOT NULL DEFAULT 1,
    is_base       BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

-- Seed default currency for default company
INSERT INTO company_currencies (company_id, code, name, symbol, exchange_rate, is_base, is_active)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'DZD', 'Algerian Dinar', 'DA', 1.0, true, true
) ON CONFLICT DO NOTHING;

-- ── Taxes ─────────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS taxes (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    code        VARCHAR(20)  NOT NULL,
    tax_type    VARCHAR(20)  NOT NULL DEFAULT 'percentage' CHECK (tax_type IN ('percentage','fixed')),
    rate        NUMERIC(10,4) NOT NULL DEFAULT 0,
    description TEXT,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

-- Seed default taxes for default company
INSERT INTO taxes (company_id, name, code, tax_type, rate, is_active)
VALUES
    ('00000000-0000-0000-0000-000000000002', 'TVA 19%', 'TVA19', 'percentage', 19.00, true),
    ('00000000-0000-0000-0000-000000000002', 'TVA 9%', 'TVA9', 'percentage', 9.00, true),
    ('00000000-0000-0000-0000-000000000002', 'TVA Exonéré', 'TVA0', 'percentage', 0.00, true),
    ('00000000-0000-0000-0000-000000000002', 'TAP 2%', 'TAP2', 'percentage', 2.00, true)
ON CONFLICT DO NOTHING;

-- ── Numbering Config ──────────────────────────────────────────────────────────
-- Alias table name used by settings handler (maps to numbering_sequences)
-- We keep the original table numbering_sequences and add a view alias
CREATE TABLE IF NOT EXISTS numbering_config (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id   UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    doc_type     VARCHAR(50)  NOT NULL,
    prefix       VARCHAR(20)  NOT NULL DEFAULT '',
    suffix       VARCHAR(20)  NOT NULL DEFAULT '',
    next_number  INT          NOT NULL DEFAULT 1,
    padding      INT          NOT NULL DEFAULT 4,
    reset_yearly BOOLEAN      NOT NULL DEFAULT TRUE,
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, doc_type)
);

-- Seed default doc types
INSERT INTO numbering_config (company_id, doc_type, prefix, padding, reset_yearly)
VALUES
    ('00000000-0000-0000-0000-000000000002', 'sales_invoice',    'INV-',  5, true),
    ('00000000-0000-0000-0000-000000000002', 'sales_order',      'SO-',   5, true),
    ('00000000-0000-0000-0000-000000000002', 'quotation',        'QUO-',  5, true),
    ('00000000-0000-0000-0000-000000000002', 'purchase_order',   'PO-',   5, true),
    ('00000000-0000-0000-0000-000000000002', 'purchase_invoice', 'PI-',   5, true),
    ('00000000-0000-0000-0000-000000000002', 'goods_receipt',    'GRN-',  5, true),
    ('00000000-0000-0000-0000-000000000002', 'payment',          'PAY-',  5, true),
    ('00000000-0000-0000-0000-000000000002', 'receipt',          'REC-',  5, true),
    ('00000000-0000-0000-0000-000000000002', 'journal_entry',    'JE-',   5, true),
    ('00000000-0000-0000-0000-000000000002', 'manufacturing_order','MO-',  5, true),
    ('00000000-0000-0000-0000-000000000002', 'stock_movement',   'SM-',   5, true),
    ('00000000-0000-0000-0000-000000000002', 'inventory_count',  'IC-',   5, true)
ON CONFLICT DO NOTHING;

-- ── Extended Workflow Rules ───────────────────────────────────────────────────
-- Add extra columns to existing workflow_rules table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_rules' AND column_name='doc_type') THEN
        ALTER TABLE workflow_rules ADD COLUMN doc_type VARCHAR(50);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_rules' AND column_name='trigger_event') THEN
        ALTER TABLE workflow_rules ADD COLUMN trigger_event VARCHAR(50) DEFAULT 'on_create';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_rules' AND column_name='conditions') THEN
        ALTER TABLE workflow_rules ADD COLUMN conditions JSONB DEFAULT '[]';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_rules' AND column_name='actions') THEN
        ALTER TABLE workflow_rules ADD COLUMN actions JSONB DEFAULT '[]';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='workflow_rules' AND column_name='priority') THEN
        ALTER TABLE workflow_rules ADD COLUMN priority INT DEFAULT 10;
    END IF;
END$$;

-- ── Fiscal Years: ensure status & is_current columns exist ────────────────────
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='fiscal_years' AND column_name='status') THEN
        ALTER TABLE fiscal_years ADD COLUMN status VARCHAR(10) NOT NULL DEFAULT 'open'
            CHECK (status IN ('open','closed','locked'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='fiscal_years' AND column_name='is_current') THEN
        ALTER TABLE fiscal_years ADD COLUMN is_current BOOLEAN NOT NULL DEFAULT FALSE;
    END IF;
END$$;

-- Update existing fiscal_years: closed → status='closed', else status='open'
UPDATE fiscal_years SET
    status = CASE WHEN is_closed THEN 'closed' ELSE 'open' END,
    is_current = CASE WHEN id='00000000-0000-0000-0000-000000000011' THEN TRUE ELSE FALSE END
WHERE status IS NULL OR status = 'open';

-- ── Companies: ensure extra columns exist ────────────────────────────────────
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='companies' AND column_name='tax_id') THEN
        ALTER TABLE companies ADD COLUMN tax_id VARCHAR(50);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='companies' AND column_name='country') THEN
        ALTER TABLE companies ADD COLUMN country VARCHAR(50) DEFAULT 'DZ';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='companies' AND column_name='timezone') THEN
        ALTER TABLE companies ADD COLUMN timezone VARCHAR(50) DEFAULT 'Africa/Algiers';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='companies' AND column_name='default_currency') THEN
        ALTER TABLE companies ADD COLUMN default_currency VARCHAR(10) DEFAULT 'DZD';
    END IF;
END$$;

-- ── Migration tracking ────────────────────────────────────────────────────────
INSERT INTO schema_migrations (version) VALUES ('0004_settings_ext') ON CONFLICT DO NOTHING;
