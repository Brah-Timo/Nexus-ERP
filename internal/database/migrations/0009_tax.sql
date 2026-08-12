-- ============================================================
-- Migration 0009 – Tax Compliance (Algerian DGI)
-- Idempotent: safe to run multiple times
-- Covers: G50, G50B, VAT register, VAT returns,
--         tax payments, IBS, TAP, withholding tax
-- ============================================================

-- ─── ENUMs ───────────────────────────────────────────────────────────────────

DO $$ BEGIN
  CREATE TYPE tax_declaration_type AS ENUM (
    'g50', 'g50b', 'ibs', 'tap', 'irg', 'tva_return', 'other'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE tax_declaration_status AS ENUM (
    'draft', 'submitted', 'accepted', 'rejected', 'amended'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE tax_period_type AS ENUM (
    'monthly', 'quarterly', 'annual'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE tax_payment_status AS ENUM (
    'pending', 'paid', 'partial', 'overdue', 'cancelled'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE tax_payment_method AS ENUM (
    'bank_transfer', 'cheque', 'cash', 'direct_debit', 'online'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE vat_register_type AS ENUM (
    'sales', 'purchase', 'credit_note', 'debit_note'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE withholding_tax_type AS ENUM (
    'irg_wages', 'irg_fees', 'irg_dividends', 'tva_withholding', 'other'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ─── SEQUENCES ───────────────────────────────────────────────────────────────

CREATE SEQUENCE IF NOT EXISTS tax_declaration_seq START 1000 INCREMENT 1;
CREATE SEQUENCE IF NOT EXISTS tax_payment_seq     START 2000 INCREMENT 1;

-- ─── CORE TABLES ─────────────────────────────────────────────────────────────

-- G50 / Tax Declarations master table
CREATE TABLE IF NOT EXISTS tax_declarations (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id       UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  declaration_type tax_declaration_type NOT NULL DEFAULT 'g50',
  period_type      tax_period_type      NOT NULL DEFAULT 'monthly',
  period_year      INTEGER NOT NULL CHECK (period_year BETWEEN 2000 AND 2100),
  period_month     INTEGER CHECK (period_month BETWEEN 1 AND 12),
  period_quarter   INTEGER CHECK (period_quarter BETWEEN 1 AND 4),
  reference        VARCHAR(50) UNIQUE NOT NULL DEFAULT ('DEC-' || nextval('tax_declaration_seq')::TEXT),
  status           tax_declaration_status NOT NULL DEFAULT 'draft',

  -- TVA (VAT) fields
  tva_collected         NUMERIC(18,2) NOT NULL DEFAULT 0,   -- TVA sur CA
  tva_deductible        NUMERIC(18,2) NOT NULL DEFAULT 0,   -- TVA déductible
  tva_credit_bf         NUMERIC(18,2) NOT NULL DEFAULT 0,   -- Crédit reporté N-1
  tva_net_due           NUMERIC(18,2) GENERATED ALWAYS AS (
                          GREATEST(0, tva_collected - tva_deductible - tva_credit_bf)
                        ) STORED,
  tva_credit_carry      NUMERIC(18,2) GENERATED ALWAYS AS (
                          GREATEST(0, tva_deductible + tva_credit_bf - tva_collected)
                        ) STORED,

  -- TAP (Taxe sur l'Activité Professionnelle)
  tap_base              NUMERIC(18,2) NOT NULL DEFAULT 0,
  tap_rate              NUMERIC(5,4)  NOT NULL DEFAULT 0.0200,  -- 2% default
  tap_amount            NUMERIC(18,2) GENERATED ALWAYS AS (tap_base * tap_rate) STORED,
  tap_reduction         NUMERIC(18,2) NOT NULL DEFAULT 0,
  tap_net_due           NUMERIC(18,2) GENERATED ALWAYS AS (
                          GREATEST(0, tap_base * tap_rate - tap_reduction)
                        ) STORED,

  -- IBS (Impôt sur les Bénéfices des Sociétés) — annual
  ibs_taxable_income    NUMERIC(18,2) NOT NULL DEFAULT 0,
  ibs_rate              NUMERIC(5,4)  NOT NULL DEFAULT 0.2300,  -- 23% default
  ibs_amount            NUMERIC(18,2) GENERATED ALWAYS AS (ibs_taxable_income * ibs_rate) STORED,
  ibs_prepayments       NUMERIC(18,2) NOT NULL DEFAULT 0,
  ibs_net_due           NUMERIC(18,2) GENERATED ALWAYS AS (
                          GREATEST(0, ibs_taxable_income * ibs_rate - ibs_prepayments)
                        ) STORED,

  -- Stamp tax / Timbre fiscal
  stamp_tax_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,

  -- Withholding taxes
  irg_wages_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
  irg_fees_amount       NUMERIC(18,2) NOT NULL DEFAULT 0,

  -- Totals
  total_tax_due         NUMERIC(18,2) GENERATED ALWAYS AS (
                          GREATEST(0, tva_collected - tva_deductible - tva_credit_bf)
                          + GREATEST(0, tap_base * tap_rate - tap_reduction)
                          + stamp_tax_amount
                          + irg_wages_amount
                          + irg_fees_amount
                        ) STORED,

  -- Meta
  submitted_at     TIMESTAMPTZ,
  accepted_at      TIMESTAMPTZ,
  submission_ref   VARCHAR(100),
  notes            TEXT,
  created_by       UUID,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- VAT Register entries (linked to invoices)
CREATE TABLE IF NOT EXISTS vat_register (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id       UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  register_type    vat_register_type NOT NULL,
  period_year      INTEGER NOT NULL,
  period_month     INTEGER NOT NULL CHECK (period_month BETWEEN 1 AND 12),
  document_date    DATE NOT NULL,
  document_number  VARCHAR(50) NOT NULL,
  document_ref     UUID,                          -- FK to invoice (nullable, polymorphic)
  partner_name     VARCHAR(200) NOT NULL DEFAULT '',
  partner_tax_id   VARCHAR(50),                   -- NIF du partenaire
  taxable_base     NUMERIC(18,2) NOT NULL DEFAULT 0,
  vat_rate         NUMERIC(5,4)  NOT NULL DEFAULT 0.1900, -- 19% standard
  vat_amount       NUMERIC(18,2) NOT NULL DEFAULT 0,
  total_amount     NUMERIC(18,2) NOT NULL DEFAULT 0,
  is_exported      BOOLEAN NOT NULL DEFAULT FALSE,
  declaration_id   UUID REFERENCES tax_declarations(id) ON DELETE SET NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- VAT Returns (récapitulatif TVA)
CREATE TABLE IF NOT EXISTS vat_returns (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id       UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  period_year      INTEGER NOT NULL,
  period_month     INTEGER NOT NULL CHECK (period_month BETWEEN 1 AND 12),
  declaration_id   UUID REFERENCES tax_declarations(id) ON DELETE SET NULL,

  -- Sales VAT summary by rate
  sales_base_0     NUMERIC(18,2) NOT NULL DEFAULT 0,
  sales_base_9     NUMERIC(18,2) NOT NULL DEFAULT 0,  -- reduced rate 9%
  sales_base_19    NUMERIC(18,2) NOT NULL DEFAULT 0,  -- standard 19%
  sales_vat_9      NUMERIC(18,2) NOT NULL DEFAULT 0,
  sales_vat_19     NUMERIC(18,2) NOT NULL DEFAULT 0,
  total_sales_base NUMERIC(18,2) GENERATED ALWAYS AS (sales_base_0 + sales_base_9 + sales_base_19) STORED,
  total_sales_vat  NUMERIC(18,2) GENERATED ALWAYS AS (sales_vat_9 + sales_vat_19) STORED,

  -- Purchase VAT summary by rate
  purch_base_9     NUMERIC(18,2) NOT NULL DEFAULT 0,
  purch_base_19    NUMERIC(18,2) NOT NULL DEFAULT 0,
  purch_vat_9      NUMERIC(18,2) NOT NULL DEFAULT 0,
  purch_vat_19     NUMERIC(18,2) NOT NULL DEFAULT 0,
  total_purch_base NUMERIC(18,2) GENERATED ALWAYS AS (purch_base_9 + purch_base_19) STORED,
  total_purch_vat  NUMERIC(18,2) GENERATED ALWAYS AS (purch_vat_9 + purch_vat_19) STORED,

  -- Credit from previous period
  credit_bf        NUMERIC(18,2) NOT NULL DEFAULT 0,

  -- Computed dues
  vat_net_due      NUMERIC(18,2) GENERATED ALWAYS AS (
                     GREATEST(0, (sales_vat_9 + sales_vat_19) - (purch_vat_9 + purch_vat_19) - credit_bf)
                   ) STORED,
  credit_cf        NUMERIC(18,2) GENERATED ALWAYS AS (
                     GREATEST(0, (purch_vat_9 + purch_vat_19) + credit_bf - (sales_vat_9 + sales_vat_19))
                   ) STORED,

  status           tax_declaration_status NOT NULL DEFAULT 'draft',
  notes            TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tax Payments
CREATE TABLE IF NOT EXISTS tax_payments (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id       UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  declaration_id   UUID REFERENCES tax_declarations(id) ON DELETE SET NULL,
  payment_number   VARCHAR(50) UNIQUE NOT NULL DEFAULT ('TXPAY-' || nextval('tax_payment_seq')::TEXT),
  payment_date     DATE NOT NULL,
  due_date         DATE NOT NULL,
  declaration_type tax_declaration_type NOT NULL DEFAULT 'g50',
  period_year      INTEGER NOT NULL,
  period_month     INTEGER,
  period_quarter   INTEGER,
  amount_due       NUMERIC(18,2) NOT NULL DEFAULT 0,
  amount_paid      NUMERIC(18,2) NOT NULL DEFAULT 0,
  balance          NUMERIC(18,2) GENERATED ALWAYS AS (amount_due - amount_paid) STORED,
  status           tax_payment_status   NOT NULL DEFAULT 'pending',
  payment_method   tax_payment_method   NOT NULL DEFAULT 'bank_transfer',
  bank_account_id  UUID REFERENCES bank_accounts(id) ON DELETE SET NULL,
  reference        VARCHAR(100),
  receipt_number   VARCHAR(100),
  notes            TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Withholding Tax (retenues à la source)
CREATE TABLE IF NOT EXISTS withholding_taxes (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id       UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  declaration_id   UUID REFERENCES tax_declarations(id) ON DELETE SET NULL,
  withholding_type withholding_tax_type NOT NULL DEFAULT 'irg_fees',
  period_year      INTEGER NOT NULL,
  period_month     INTEGER NOT NULL CHECK (period_month BETWEEN 1 AND 12),
  document_date    DATE NOT NULL,
  beneficiary_name VARCHAR(200) NOT NULL,
  beneficiary_nif  VARCHAR(50),
  gross_amount     NUMERIC(18,2) NOT NULL DEFAULT 0,
  wht_rate         NUMERIC(5,4)  NOT NULL DEFAULT 0,
  wht_amount       NUMERIC(18,2) GENERATED ALWAYS AS (gross_amount * wht_rate) STORED,
  net_amount       NUMERIC(18,2) GENERATED ALWAYS AS (gross_amount - gross_amount * wht_rate) STORED,
  is_remitted      BOOLEAN NOT NULL DEFAULT FALSE,
  remittance_ref   VARCHAR(100),
  notes            TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tax Rates configuration (per company)
CREATE TABLE IF NOT EXISTS tax_rate_config (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id       UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  tax_type         VARCHAR(20) NOT NULL,   -- 'tva', 'tap', 'ibs', 'irg'
  rate_name        VARCHAR(100) NOT NULL,
  rate_value       NUMERIC(8,6) NOT NULL,
  effective_from   DATE NOT NULL DEFAULT CURRENT_DATE,
  effective_to     DATE,
  is_active        BOOLEAN NOT NULL DEFAULT TRUE,
  notes            TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (company_id, tax_type, rate_name, effective_from)
);

-- ─── EXTEND EXISTING TABLES ───────────────────────────────────────────────────

-- Add NIF to companies if missing
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='companies' AND column_name='nif') THEN
    ALTER TABLE companies ADD COLUMN nif VARCHAR(20);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='companies' AND column_name='nis') THEN
    ALTER TABLE companies ADD COLUMN nis VARCHAR(20);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='companies' AND column_name='rc') THEN
    ALTER TABLE companies ADD COLUMN rc VARCHAR(30);   -- Registre de commerce
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='companies' AND column_name='art') THEN
    ALTER TABLE companies ADD COLUMN art VARCHAR(30);  -- Article d'imposition
  END IF;
END $$;

-- Add TAP base column to sales_invoices if missing
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='sales_invoices' AND column_name='tap_amount') THEN
    ALTER TABLE sales_invoices ADD COLUMN tap_amount NUMERIC(18,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

-- ─── INDEXES ─────────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_tax_decl_company      ON tax_declarations(company_id);
CREATE INDEX IF NOT EXISTS idx_tax_decl_type         ON tax_declarations(declaration_type);
CREATE INDEX IF NOT EXISTS idx_tax_decl_period       ON tax_declarations(period_year, period_month);
CREATE INDEX IF NOT EXISTS idx_tax_decl_status       ON tax_declarations(status);
CREATE INDEX IF NOT EXISTS idx_tax_decl_ref          ON tax_declarations(reference);

CREATE INDEX IF NOT EXISTS idx_vat_reg_company       ON vat_register(company_id);
CREATE INDEX IF NOT EXISTS idx_vat_reg_type          ON vat_register(register_type);
CREATE INDEX IF NOT EXISTS idx_vat_reg_period        ON vat_register(period_year, period_month);
CREATE INDEX IF NOT EXISTS idx_vat_reg_docnum        ON vat_register(document_number);
CREATE INDEX IF NOT EXISTS idx_vat_reg_decl          ON vat_register(declaration_id);

CREATE INDEX IF NOT EXISTS idx_vat_returns_company   ON vat_returns(company_id);
CREATE INDEX IF NOT EXISTS idx_vat_returns_period    ON vat_returns(period_year, period_month);

CREATE INDEX IF NOT EXISTS idx_tax_pay_company       ON tax_payments(company_id);
CREATE INDEX IF NOT EXISTS idx_tax_pay_status        ON tax_payments(status);
CREATE INDEX IF NOT EXISTS idx_tax_pay_duedate       ON tax_payments(due_date);
CREATE INDEX IF NOT EXISTS idx_tax_pay_type          ON tax_payments(declaration_type);
CREATE INDEX IF NOT EXISTS idx_tax_pay_period        ON tax_payments(period_year, period_month);
CREATE INDEX IF NOT EXISTS idx_tax_pay_decl          ON tax_payments(declaration_id);

CREATE INDEX IF NOT EXISTS idx_wht_company           ON withholding_taxes(company_id);
CREATE INDEX IF NOT EXISTS idx_wht_period            ON withholding_taxes(period_year, period_month);
CREATE INDEX IF NOT EXISTS idx_wht_type              ON withholding_taxes(withholding_type);

-- ─── TRIGGERS ─────────────────────────────────────────────────────────────────

-- updated_at trigger function (reuse if exists)
CREATE OR REPLACE FUNCTION set_updated_at_tax()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$;

DO $$ BEGIN
  CREATE TRIGGER tax_declarations_updated_at
    BEFORE UPDATE ON tax_declarations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_tax();
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TRIGGER vat_returns_updated_at
    BEFORE UPDATE ON vat_returns
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_tax();
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TRIGGER tax_payments_updated_at
    BEFORE UPDATE ON tax_payments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at_tax();
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ─── VIEWS ────────────────────────────────────────────────────────────────────

-- G50 Summary view
CREATE OR REPLACE VIEW v_g50_summary AS
SELECT
  d.id,
  d.company_id,
  d.reference,
  d.declaration_type,
  d.period_year,
  d.period_month,
  d.period_quarter,
  d.status,
  d.tva_collected,
  d.tva_deductible,
  d.tva_credit_bf,
  d.tva_net_due,
  d.tva_credit_carry,
  d.tap_base,
  d.tap_rate,
  d.tap_net_due,
  d.ibs_taxable_income,
  d.ibs_rate,
  d.ibs_net_due,
  d.stamp_tax_amount,
  d.irg_wages_amount,
  d.irg_fees_amount,
  d.total_tax_due,
  d.submitted_at,
  d.submission_ref,
  d.notes,
  d.created_at,
  COALESCE(SUM(tp.amount_paid), 0)  AS total_paid,
  d.total_tax_due - COALESCE(SUM(tp.amount_paid), 0) AS balance_due
FROM tax_declarations d
LEFT JOIN tax_payments tp ON tp.declaration_id = d.id AND tp.status != 'cancelled'
GROUP BY d.id;

-- VAT Register summary view
CREATE OR REPLACE VIEW v_vat_register_summary AS
SELECT
  company_id,
  period_year,
  period_month,
  register_type,
  COUNT(*)                  AS entry_count,
  SUM(taxable_base)         AS total_base,
  SUM(vat_amount)           AS total_vat,
  SUM(total_amount)         AS total_amount,
  COUNT(*) FILTER (WHERE is_exported) AS exported_count
FROM vat_register
GROUP BY company_id, period_year, period_month, register_type;

-- Tax payments aging view
CREATE OR REPLACE VIEW v_tax_payment_aging AS
SELECT
  tp.id,
  tp.company_id,
  tp.payment_number,
  tp.declaration_type,
  tp.period_year,
  tp.period_month,
  tp.due_date,
  tp.amount_due,
  tp.amount_paid,
  tp.balance,
  tp.status,
  CASE
    WHEN tp.status = 'paid'      THEN 'Paid'
    WHEN tp.due_date >= CURRENT_DATE THEN 'Current'
    WHEN CURRENT_DATE - tp.due_date <= 30  THEN '1-30 days'
    WHEN CURRENT_DATE - tp.due_date <= 60  THEN '31-60 days'
    WHEN CURRENT_DATE - tp.due_date <= 90  THEN '61-90 days'
    ELSE 'Over 90 days'
  END AS aging_bucket,
  CURRENT_DATE - tp.due_date AS days_overdue
FROM tax_payments tp
WHERE tp.status != 'cancelled';

-- Annual tax summary view
CREATE OR REPLACE VIEW v_annual_tax_summary AS
SELECT
  d.company_id,
  d.period_year,
  SUM(d.tva_net_due)         AS total_vat_due,
  SUM(d.tap_net_due)         AS total_tap_due,
  SUM(d.ibs_net_due)         AS total_ibs_due,
  SUM(d.stamp_tax_amount)    AS total_stamp_due,
  SUM(d.irg_wages_amount)    AS total_irg_wages,
  SUM(d.irg_fees_amount)     AS total_irg_fees,
  SUM(d.total_tax_due)       AS grand_total_due,
  COALESCE(SUM(tp.amount_paid), 0)  AS grand_total_paid,
  SUM(d.total_tax_due) - COALESCE(SUM(tp.amount_paid), 0) AS grand_balance
FROM tax_declarations d
LEFT JOIN tax_payments tp ON tp.declaration_id = d.id AND tp.status != 'cancelled'
GROUP BY d.company_id, d.period_year;

-- ─── DEFAULT TAX RATE CONFIG ─────────────────────────────────────────────────

-- Insert standard Algerian tax rates (non-company-specific placeholder)
-- These would be seeded per company on first use; here we check existence before insert

INSERT INTO tax_rate_config (id, company_id, tax_type, rate_name, rate_value, effective_from)
SELECT
  gen_random_uuid(),
  c.id,
  rates.tax_type,
  rates.rate_name,
  rates.rate_value,
  '2024-01-01'::DATE
FROM companies c
CROSS JOIN (VALUES
  ('tva',  'Standard 19%',     0.190000),
  ('tva',  'Reduced 9%',       0.090000),
  ('tva',  'Zero-rated 0%',    0.000000),
  ('tap',  'Standard 2%',      0.020000),
  ('tap',  'Export 0%',        0.000000),
  ('ibs',  'Standard 23%',     0.230000),
  ('ibs',  'SME 26%',          0.260000),
  ('irg',  'Band 0-240K 0%',   0.000000),
  ('irg',  'Band 240K-480K 23%', 0.230000),
  ('irg',  'Band 480K-960K 27%', 0.270000),
  ('irg',  'Band 960K-1920K 30%', 0.300000),
  ('irg',  'Band 1920K-3840K 33%', 0.330000),
  ('irg',  'Over 3840K 35%',   0.350000)
) AS rates(tax_type, rate_name, rate_value)
WHERE NOT EXISTS (
  SELECT 1 FROM tax_rate_config rc
  WHERE rc.company_id = c.id AND rc.tax_type = rates.tax_type AND rc.rate_name = rates.rate_name
);
