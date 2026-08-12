-- =============================================================================
-- Nexus ERP — Sales & CRM Schema Patch  v1.7.0
-- Migration: 0002_sales_crm.sql
-- Safe to run on top of 0001_init_schema.sql — all statements are idempotent.
-- =============================================================================

-- ─── Extension guard ─────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- 1. LINE TABLE FIXES
-- =============================================================================

-- quotation_lines: ensure PK exists
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'quotation_lines_pkey'
      AND conrelid = 'quotation_lines'::regclass
  ) THEN
    ALTER TABLE quotation_lines ADD PRIMARY KEY (id);
  END IF;
END $$;

ALTER TABLE quotation_lines ALTER COLUMN id SET DEFAULT uuid_generate_v4();

-- quotation_lines: add sort_order if missing (some DB versions lack it)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'quotation_lines' AND column_name = 'sort_order'
  ) THEN
    ALTER TABLE quotation_lines ADD COLUMN sort_order INT NOT NULL DEFAULT 0;
  END IF;
END $$;

ALTER TABLE quotation_lines ALTER COLUMN sort_order SET DEFAULT 0;

-- sales_order_lines: ensure PK exists
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'sales_order_lines_pkey'
      AND conrelid = 'sales_order_lines'::regclass
  ) THEN
    ALTER TABLE sales_order_lines ADD PRIMARY KEY (id);
  END IF;
END $$;

ALTER TABLE sales_order_lines ALTER COLUMN id SET DEFAULT uuid_generate_v4();

-- sales_order_lines: add sort_order if missing (critical fix)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_order_lines' AND column_name = 'sort_order'
  ) THEN
    ALTER TABLE sales_order_lines ADD COLUMN sort_order INT NOT NULL DEFAULT 0;
  END IF;
END $$;

ALTER TABLE sales_order_lines ALTER COLUMN sort_order SET DEFAULT 0;

-- sales_order_lines: add subtotal if missing (generated or plain numeric)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_order_lines' AND column_name = 'subtotal'
  ) THEN
    ALTER TABLE sales_order_lines ADD COLUMN subtotal NUMERIC(18,2);
  END IF;
END $$;

-- sales_order_lines: add tva_amount if missing
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_order_lines' AND column_name = 'tva_amount'
  ) THEN
    ALTER TABLE sales_order_lines ADD COLUMN tva_amount NUMERIC(18,2);
  END IF;
END $$;

-- sales_order_lines: add total if missing
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_order_lines' AND column_name = 'total'
  ) THEN
    ALTER TABLE sales_order_lines ADD COLUMN total NUMERIC(18,2);
  END IF;
END $$;

-- sales_order_lines: add discount_pct if missing
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_order_lines' AND column_name = 'discount_pct'
  ) THEN
    ALTER TABLE sales_order_lines ADD COLUMN discount_pct NUMERIC(5,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

-- sales_order_lines: add tva_rate if missing
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_order_lines' AND column_name = 'tva_rate'
  ) THEN
    ALTER TABLE sales_order_lines ADD COLUMN tva_rate NUMERIC(5,2) NOT NULL DEFAULT 19;
  END IF;
END $$;

-- sales_order_lines: add account_id if missing
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_order_lines' AND column_name = 'account_id'
  ) THEN
    ALTER TABLE sales_order_lines ADD COLUMN account_id UUID;
  END IF;
END $$;

-- quotation_lines: same safety guards
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'quotation_lines' AND column_name = 'subtotal'
  ) THEN
    ALTER TABLE quotation_lines ADD COLUMN subtotal NUMERIC(18,2);
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'quotation_lines' AND column_name = 'tva_amount'
  ) THEN
    ALTER TABLE quotation_lines ADD COLUMN tva_amount NUMERIC(18,2);
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'quotation_lines' AND column_name = 'total'
  ) THEN
    ALTER TABLE quotation_lines ADD COLUMN total NUMERIC(18,2);
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'quotation_lines' AND column_name = 'discount_pct'
  ) THEN
    ALTER TABLE quotation_lines ADD COLUMN discount_pct NUMERIC(5,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'quotation_lines' AND column_name = 'tva_rate'
  ) THEN
    ALTER TABLE quotation_lines ADD COLUMN tva_rate NUMERIC(5,2) NOT NULL DEFAULT 19;
  END IF;
END $$;

-- =============================================================================
-- 2. EXTRA COLUMNS ON SALES INVOICES
-- =============================================================================

-- payment_method + payment_date (used by RecordPayment handler)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_invoices' AND column_name = 'payment_method'
  ) THEN
    ALTER TABLE sales_invoices ADD COLUMN payment_method VARCHAR(30);
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_invoices' AND column_name = 'payment_date'
  ) THEN
    ALTER TABLE sales_invoices ADD COLUMN payment_date DATE;
  END IF;
END $$;

-- confirmed_at / confirmed_by (used by ConfirmInvoice handler)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_invoices' AND column_name = 'confirmed_at'
  ) THEN
    ALTER TABLE sales_invoices ADD COLUMN confirmed_at TIMESTAMPTZ;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'sales_invoices' AND column_name = 'confirmed_by'
  ) THEN
    ALTER TABLE sales_invoices ADD COLUMN confirmed_by UUID;
  END IF;
END $$;

-- =============================================================================
-- 3. EXTRA COLUMNS ON QUOTATIONS
-- =============================================================================

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'quotations' AND column_name = 'converted_to'
  ) THEN
    ALTER TABLE quotations ADD COLUMN converted_to UUID;
  END IF;
END $$;

-- =============================================================================
-- 4. INDEXES
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_leads_company       ON leads(company_id, status);
CREATE INDEX IF NOT EXISTS idx_leads_salesperson   ON leads(salesperson_id);

CREATE INDEX IF NOT EXISTS idx_opp_company_stage   ON opportunities(company_id, stage);
CREATE INDEX IF NOT EXISTS idx_opp_customer        ON opportunities(customer_id);
CREATE INDEX IF NOT EXISTS idx_opp_updated         ON opportunities(company_id, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_cust_company        ON customers(company_id, is_active);
CREATE INDEX IF NOT EXISTS idx_cust_name           ON customers(company_id, name);

CREATE INDEX IF NOT EXISTS idx_quot_company        ON quotations(company_id, status);
CREATE INDEX IF NOT EXISTS idx_quot_customer       ON quotations(customer_id);
CREATE INDEX IF NOT EXISTS idx_quot_date           ON quotations(company_id, date DESC);

CREATE INDEX IF NOT EXISTS idx_so_company          ON sales_orders(company_id, status);
CREATE INDEX IF NOT EXISTS idx_so_customer         ON sales_orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_so_date             ON sales_orders(company_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_so_quotation        ON sales_orders(quotation_id);

CREATE INDEX IF NOT EXISTS idx_si_company          ON sales_invoices(company_id, status);
CREATE INDEX IF NOT EXISTS idx_si_customer         ON sales_invoices(customer_id);
CREATE INDEX IF NOT EXISTS idx_si_date             ON sales_invoices(company_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_si_due_date         ON sales_invoices(company_id, due_date);
CREATE INDEX IF NOT EXISTS idx_si_order            ON sales_invoices(order_id);

CREATE INDEX IF NOT EXISTS idx_sil_invoice         ON sales_invoice_lines(invoice_id);
CREATE INDEX IF NOT EXISTS idx_ql_quotation        ON quotation_lines(quotation_id);
CREATE INDEX IF NOT EXISTS idx_sol_order           ON sales_order_lines(order_id);

-- =============================================================================
-- 5. VIEWS
-- =============================================================================

CREATE OR REPLACE VIEW v_pipeline_summary AS
SELECT
    company_id,
    stage,
    COUNT(*)                                        AS count,
    COALESCE(SUM(amount), 0)                        AS total_amount,
    COALESCE(ROUND(AVG(probability)::numeric, 0), 0) AS avg_probability
FROM opportunities
GROUP BY company_id, stage;

CREATE OR REPLACE VIEW v_customer_aging AS
SELECT
    si.company_id,
    cu.id                                            AS customer_id,
    cu.name                                          AS customer_name,
    COALESCE(cu.phone,  '')                          AS phone,
    COALESCE(cu.email,  '')                          AS email,
    COUNT(si.id)                                     AS invoice_count,
    COALESCE(SUM(CASE
        WHEN si.due_date IS NULL OR CURRENT_DATE <= si.due_date
        THEN si.balance_due ELSE 0 END), 0)          AS current_amount,
    COALESCE(SUM(CASE
        WHEN si.due_date IS NOT NULL
         AND CURRENT_DATE - si.due_date BETWEEN 1  AND 30
        THEN si.balance_due ELSE 0 END), 0)          AS days_1_30,
    COALESCE(SUM(CASE
        WHEN si.due_date IS NOT NULL
         AND CURRENT_DATE - si.due_date BETWEEN 31 AND 60
        THEN si.balance_due ELSE 0 END), 0)          AS days_31_60,
    COALESCE(SUM(CASE
        WHEN si.due_date IS NOT NULL
         AND CURRENT_DATE - si.due_date BETWEEN 61 AND 90
        THEN si.balance_due ELSE 0 END), 0)          AS days_61_90,
    COALESCE(SUM(CASE
        WHEN si.due_date IS NOT NULL
         AND CURRENT_DATE - si.due_date > 90
        THEN si.balance_due ELSE 0 END), 0)          AS days_over_90,
    COALESCE(SUM(si.balance_due), 0)                 AS total_outstanding
FROM sales_invoices si
JOIN customers cu ON cu.id = si.customer_id
WHERE si.status NOT IN ('paid', 'cancelled')
GROUP BY si.company_id, cu.id, cu.name, cu.phone, cu.email;

CREATE OR REPLACE VIEW v_dashboard_summary AS
SELECT
    co.id AS company_id,
    COALESCE((
        SELECT SUM(total_amount) FROM sales_invoices
        WHERE company_id = co.id
          AND status IN ('confirmed','partially_paid','paid')
          AND date_trunc('month', date) = date_trunc('month', CURRENT_DATE)
    ), 0) AS monthly_sales,
    COALESCE((
        SELECT SUM(balance_due) FROM sales_invoices
        WHERE company_id = co.id AND status NOT IN ('paid','cancelled')
    ), 0) AS receivables,
    COALESCE((
        SELECT COUNT(*) FROM sales_invoices
        WHERE company_id = co.id
          AND status NOT IN ('paid','cancelled')
          AND due_date < CURRENT_DATE
    ), 0) AS overdue_invoices,
    COALESCE((
        SELECT COUNT(*) FROM customers
        WHERE company_id = co.id AND is_active = TRUE
    ), 0) AS customer_count,
    COALESCE((
        SELECT COUNT(*) FROM opportunities
        WHERE company_id = co.id AND stage NOT IN ('won','lost')
    ), 0) AS open_opportunities,
    COALESCE((
        SELECT SUM(amount * probability / 100.0) FROM opportunities
        WHERE company_id = co.id AND stage NOT IN ('won','lost')
    ), 0) AS pipeline_value,
    COALESCE((
        SELECT SUM(amount) FROM opportunities
        WHERE company_id = co.id AND stage = 'won'
          AND date_trunc('month', updated_at) = date_trunc('month', CURRENT_DATE)
    ), 0) AS won_this_month,
    COALESCE((
        SELECT COUNT(*) FROM quotations
        WHERE company_id = co.id AND status = 'draft'
    ), 0) AS draft_quotations,
    COALESCE((
        SELECT COUNT(*) FROM sales_orders
        WHERE company_id = co.id AND status NOT IN ('cancelled','delivered')
    ), 0) AS open_orders
FROM companies co;

-- =============================================================================
-- End of 0002_sales_crm.sql
-- =============================================================================
