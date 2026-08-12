-- =============================================================================
-- Nexus ERP — Purchase Schema Patch  v1.7.0
-- Migration: 0003_purchase.sql
-- Safe to run on top of 0001_init_schema.sql — all statements are idempotent.
-- Root cause fixed: old migration referenced customer_id which does NOT exist
-- on purchase tables (suppliers use supplier_id, not customer_id).
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- 1. SUPPLIERS TABLE — add missing columns safely
-- =============================================================================

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='suppliers' AND column_name='type') THEN
    ALTER TABLE suppliers ADD COLUMN type VARCHAR(20) NOT NULL DEFAULT 'company';
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='suppliers' AND column_name='nif') THEN
    ALTER TABLE suppliers ADD COLUMN nif VARCHAR(20);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='suppliers' AND column_name='nis') THEN
    ALTER TABLE suppliers ADD COLUMN nis VARCHAR(20);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='suppliers' AND column_name='rc') THEN
    ALTER TABLE suppliers ADD COLUMN rc VARCHAR(30);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='suppliers' AND column_name='art') THEN
    ALTER TABLE suppliers ADD COLUMN art VARCHAR(30);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='suppliers' AND column_name='wilaya') THEN
    ALTER TABLE suppliers ADD COLUMN wilaya VARCHAR(100);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='suppliers' AND column_name='contact_name') THEN
    ALTER TABLE suppliers ADD COLUMN contact_name VARCHAR(200);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='suppliers' AND column_name='credit_limit') THEN
    ALTER TABLE suppliers ADD COLUMN credit_limit NUMERIC(18,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='suppliers' AND column_name='rating') THEN
    ALTER TABLE suppliers ADD COLUMN rating INT NOT NULL DEFAULT 3;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='suppliers' AND column_name='notes') THEN
    ALTER TABLE suppliers ADD COLUMN notes TEXT;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='suppliers' AND column_name='updated_at') THEN
    ALTER TABLE suppliers ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
  END IF;
END $$;

-- =============================================================================
-- 2. RFQs TABLE — add missing columns safely
-- NOTE: rfqs does NOT have customer_id — it has supplier_id
-- =============================================================================

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='rfqs' AND column_name='total_amount') THEN
    ALTER TABLE rfqs ADD COLUMN total_amount NUMERIC(18,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='rfqs' AND column_name='subtotal') THEN
    ALTER TABLE rfqs ADD COLUMN subtotal NUMERIC(18,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='rfqs' AND column_name='tva_amount') THEN
    ALTER TABLE rfqs ADD COLUMN tva_amount NUMERIC(18,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='rfqs' AND column_name='currency') THEN
    ALTER TABLE rfqs ADD COLUMN currency VARCHAR(10) NOT NULL DEFAULT 'DZD';
  END IF;
END $$;

-- Create rfq_lines table if it doesn't exist
CREATE TABLE IF NOT EXISTS rfq_lines (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rfq_id          UUID         NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
    item_id         UUID,
    description     VARCHAR(500) NOT NULL DEFAULT '',
    quantity        NUMERIC(18,4) NOT NULL DEFAULT 1,
    unit_price      NUMERIC(18,4) NOT NULL DEFAULT 0,
    discount_pct    NUMERIC(5,2)  NOT NULL DEFAULT 0,
    tva_rate        NUMERIC(5,2)  NOT NULL DEFAULT 19,
    subtotal        NUMERIC(18,2),
    tva_amount      NUMERIC(18,2),
    total           NUMERIC(18,2),
    sort_order      INT          NOT NULL DEFAULT 0
);

-- =============================================================================
-- 3. PURCHASE_ORDERS TABLE — fix column names (0001 uses subtotal, not sub_total)
-- =============================================================================

-- The real schema in 0001 uses: subtotal, discount_amount, tva_amount, total_amount
-- Old Go handler used sub_total — we add alias columns safely
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='purchase_orders' AND column_name='supplier_name') THEN
    ALTER TABLE purchase_orders ADD COLUMN supplier_name VARCHAR(300);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='purchase_orders' AND column_name='confirmed_at') THEN
    ALTER TABLE purchase_orders ADD COLUMN confirmed_at TIMESTAMPTZ;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='purchase_orders' AND column_name='cancelled_at') THEN
    ALTER TABLE purchase_orders ADD COLUMN cancelled_at TIMESTAMPTZ;
  END IF;
END $$;

-- =============================================================================
-- 4. PURCHASE_ORDER_LINES TABLE — fix column names used by old handler
-- Old handler used: item_code, item_name, sub_total, total_amount
-- Real schema uses:  description, subtotal, tva_amount, total
-- =============================================================================

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='purchase_order_lines' AND column_name='item_code') THEN
    ALTER TABLE purchase_order_lines ADD COLUMN item_code VARCHAR(50);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='purchase_order_lines' AND column_name='item_name') THEN
    ALTER TABLE purchase_order_lines ADD COLUMN item_name VARCHAR(300);
  END IF;
END $$;

-- Ensure sort_order exists (may be missing on some installs)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='purchase_order_lines' AND column_name='sort_order') THEN
    ALTER TABLE purchase_order_lines ADD COLUMN sort_order INT NOT NULL DEFAULT 0;
  END IF;
END $$;

-- =============================================================================
-- 5. GOODS_RECEIPTS TABLE — add missing columns
-- NOTE: no customer_id here — schema uses supplier_id
-- =============================================================================

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='goods_receipts' AND column_name='supplier_name') THEN
    ALTER TABLE goods_receipts ADD COLUMN supplier_name VARCHAR(300);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='goods_receipts' AND column_name='total_amount') THEN
    ALTER TABLE goods_receipts ADD COLUMN total_amount NUMERIC(18,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

-- =============================================================================
-- 6. GOODS_RECEIPT_LINES TABLE
-- =============================================================================

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='goods_receipt_lines' AND column_name='item_code') THEN
    ALTER TABLE goods_receipt_lines ADD COLUMN item_code VARCHAR(50);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='goods_receipt_lines' AND column_name='item_name') THEN
    ALTER TABLE goods_receipt_lines ADD COLUMN item_name VARCHAR(300);
  END IF;
END $$;

-- =============================================================================
-- 7. PURCHASE_INVOICES TABLE — add missing columns
-- NOTE: no customer_id here — uses supplier_id
-- =============================================================================

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='purchase_invoices' AND column_name='supplier_name') THEN
    ALTER TABLE purchase_invoices ADD COLUMN supplier_name VARCHAR(300);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='purchase_invoices' AND column_name='balance_due') THEN
    ALTER TABLE purchase_invoices ADD COLUMN balance_due NUMERIC(18,2)
        GENERATED ALWAYS AS (total_amount - paid_amount) STORED;
  END IF;
END $$;

-- Create purchase_invoice_lines if missing
CREATE TABLE IF NOT EXISTS purchase_invoice_lines (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id          UUID         NOT NULL REFERENCES purchase_invoices(id) ON DELETE CASCADE,
    po_line_id          UUID         REFERENCES purchase_order_lines(id),
    item_id             UUID,
    description         VARCHAR(500) NOT NULL DEFAULT '',
    quantity            NUMERIC(18,4) NOT NULL DEFAULT 1,
    unit_price          NUMERIC(18,4) NOT NULL DEFAULT 0,
    discount_pct        NUMERIC(5,2)  NOT NULL DEFAULT 0,
    tva_rate            NUMERIC(5,2)  NOT NULL DEFAULT 19,
    subtotal            NUMERIC(18,2),
    tva_amount          NUMERIC(18,2),
    total               NUMERIC(18,2),
    account_id          UUID         REFERENCES chart_of_accounts(id),
    sort_order          INT          NOT NULL DEFAULT 0
);

-- =============================================================================
-- 8. SUPPLIER_EVALUATIONS — add overall_score default
-- =============================================================================

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='supplier_evaluations' AND column_name='quality_score') THEN
    ALTER TABLE supplier_evaluations ADD COLUMN quality_score INT NOT NULL DEFAULT 3;
  END IF;
END $$;

-- =============================================================================
-- 9. INDEXES
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_suppliers_company          ON suppliers(company_id);
CREATE INDEX IF NOT EXISTS idx_suppliers_name             ON suppliers(company_id, name);
CREATE INDEX IF NOT EXISTS idx_rfqs_company_status        ON rfqs(company_id, status);
CREATE INDEX IF NOT EXISTS idx_rfqs_supplier              ON rfqs(supplier_id);
CREATE INDEX IF NOT EXISTS idx_po_supplier                ON purchase_orders(supplier_id);
CREATE INDEX IF NOT EXISTS idx_po_rfq                     ON purchase_orders(rfq_id);
CREATE INDEX IF NOT EXISTS idx_pol_po                     ON purchase_order_lines(po_id);
CREATE INDEX IF NOT EXISTS idx_gr_company                 ON goods_receipts(company_id, status);
CREATE INDEX IF NOT EXISTS idx_gr_po                      ON goods_receipts(po_id);
CREATE INDEX IF NOT EXISTS idx_grl_grn                    ON goods_receipt_lines(grn_id);
CREATE INDEX IF NOT EXISTS idx_pi_company_status          ON purchase_invoices(company_id, status);
CREATE INDEX IF NOT EXISTS idx_pi_supplier                ON purchase_invoices(supplier_id);
CREATE INDEX IF NOT EXISTS idx_pi_due_date                ON purchase_invoices(due_date);
CREATE INDEX IF NOT EXISTS idx_rfq_lines_rfq              ON rfq_lines(rfq_id);
CREATE INDEX IF NOT EXISTS idx_pinv_lines_invoice         ON purchase_invoice_lines(invoice_id);

-- =============================================================================
-- 10. VIEWS
-- =============================================================================

CREATE OR REPLACE VIEW v_purchase_dashboard AS
SELECT
    c.id AS company_id,
    COUNT(DISTINCT s.id)  FILTER (WHERE s.is_active)                         AS active_suppliers,
    COUNT(DISTINCT po.id) FILTER (WHERE po.status = 'draft')                 AS draft_orders,
    COUNT(DISTINCT po.id) FILTER (WHERE po.status = 'approved')              AS approved_orders,
    COALESCE(SUM(po.total_amount) FILTER (WHERE po.status NOT IN ('cancelled')), 0) AS total_po_amount,
    COUNT(DISTINCT pi.id) FILTER (WHERE pi.status NOT IN ('paid','cancelled')) AS open_invoices,
    COALESCE(SUM(pi.total_amount - pi.paid_amount)
        FILTER (WHERE pi.status NOT IN ('paid','cancelled')), 0)              AS outstanding_payables,
    COUNT(DISTINCT gr.id) FILTER (WHERE gr.status = 'draft')                 AS pending_receipts
FROM companies c
LEFT JOIN suppliers s    ON s.company_id = c.id
LEFT JOIN purchase_orders po ON po.company_id = c.id
LEFT JOIN purchase_invoices pi ON pi.company_id = c.id
LEFT JOIN goods_receipts gr ON gr.company_id = c.id
GROUP BY c.id;

CREATE OR REPLACE VIEW v_supplier_balance AS
SELECT
    s.company_id,
    s.id AS supplier_id,
    s.name AS supplier_name,
    s.code,
    COALESCE(SUM(pi.total_amount - pi.paid_amount)
        FILTER (WHERE pi.status NOT IN ('paid','cancelled')), 0) AS balance_due,
    COUNT(pi.id) FILTER (WHERE pi.status NOT IN ('paid','cancelled'))         AS open_invoices,
    COUNT(pi.id) FILTER (WHERE pi.due_date < CURRENT_DATE
                          AND pi.status NOT IN ('paid','cancelled'))          AS overdue_invoices
FROM suppliers s
LEFT JOIN purchase_invoices pi ON pi.supplier_id = s.id
GROUP BY s.company_id, s.id, s.name, s.code;
