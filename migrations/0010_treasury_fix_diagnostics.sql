-- =============================================================================
-- 0010_treasury_fix_diagnostics.sql
-- Fix treasury tables schema mismatch + System Diagnostics & Error Tracking
-- Idempotent — safe to run multiple times
-- =============================================================================

-- ─────────────────────────────────────────────────────────────────────────────
-- PART 1: FIX CHEQUES TABLE
-- The original schema uses cheque_type/payee_payer/deposited_at/bounced_at
-- The handler uses type/partner_name/deposited_date/bounced_date
-- Strategy: add missing columns as aliases via new columns
-- ─────────────────────────────────────────────────────────────────────────────

-- Add type column (alias for cheque_type)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='cheques' AND column_name='type') THEN
    ALTER TABLE cheques ADD COLUMN type VARCHAR(20) NOT NULL DEFAULT 'received';
    UPDATE cheques SET type = cheque_type WHERE type = 'received' AND cheque_type IS NOT NULL;
  END IF;
END $$;

-- Add partner_name column (alias for payee_payer)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='cheques' AND column_name='partner_name') THEN
    ALTER TABLE cheques ADD COLUMN partner_name VARCHAR(300);
    UPDATE cheques SET partner_name = payee_payer WHERE partner_name IS NULL AND payee_payer IS NOT NULL;
  END IF;
END $$;

-- Add partner_id column
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='cheques' AND column_name='partner_id') THEN
    ALTER TABLE cheques ADD COLUMN partner_id UUID;
  END IF;
END $$;

-- Add deposited_date column (alias for deposited_at as DATE)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='cheques' AND column_name='deposited_date') THEN
    ALTER TABLE cheques ADD COLUMN deposited_date DATE;
    UPDATE cheques SET deposited_date = deposited_at::DATE WHERE deposited_at IS NOT NULL;
  END IF;
END $$;

-- Add bounced_date column (alias for bounced_at as DATE)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='cheques' AND column_name='bounced_date') THEN
    ALTER TABLE cheques ADD COLUMN bounced_date DATE;
    UPDATE cheques SET bounced_date = bounced_at::DATE WHERE bounced_at IS NOT NULL;
  END IF;
END $$;

-- Add notes column if not exists
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='cheques' AND column_name='notes') THEN
    ALTER TABLE cheques ADD COLUMN notes TEXT;
  END IF;
END $$;

-- ─────────────────────────────────────────────────────────────────────────────
-- PART 2: FIX PAYMENTS TABLE
-- Original: number, payment_type, direction, date, amount, currency,
--           cash_account_id, bank_account_id, reference, notes
-- Handler uses: type, partner_name, partner_id, allocated_amount, method, status
-- ─────────────────────────────────────────────────────────────────────────────

-- Add type column (handler's name for payment category)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='payments' AND column_name='type') THEN
    ALTER TABLE payments ADD COLUMN type VARCHAR(50) NOT NULL DEFAULT 'supplier';
  END IF;
END $$;

-- Add partner_name
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='payments' AND column_name='partner_name') THEN
    ALTER TABLE payments ADD COLUMN partner_name VARCHAR(300);
  END IF;
END $$;

-- Add allocated_amount
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='payments' AND column_name='allocated_amount') THEN
    ALTER TABLE payments ADD COLUMN allocated_amount NUMERIC(18,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

-- Add method (handler uses method, original uses payment_type)
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='payments' AND column_name='method') THEN
    ALTER TABLE payments ADD COLUMN method VARCHAR(50);
    UPDATE payments SET method = payment_type::TEXT WHERE method IS NULL;
  END IF;
END $$;

-- Add status
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='payments' AND column_name='status') THEN
    ALTER TABLE payments ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'draft';
  END IF;
END $$;

-- Add date column (handler uses 'date', original also has 'date' - should be fine)
-- Ensure updated_at exists
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='payments' AND column_name='updated_at') THEN
    ALTER TABLE payments ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
  END IF;
END $$;

-- ─────────────────────────────────────────────────────────────────────────────
-- PART 3: CREATE RECEIPTS TABLE (referenced in handler but didn't exist)
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS receipts (
    id                UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id        UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    number            VARCHAR(50)  NOT NULL,
    receipt_type      VARCHAR(30)  NOT NULL DEFAULT 'customer',  -- customer/other
    status            VARCHAR(20)  NOT NULL DEFAULT 'draft',
    partner_id        UUID,
    partner_name      VARCHAR(300),
    amount            NUMERIC(18,2) NOT NULL DEFAULT 0,
    allocated_amount  NUMERIC(18,2) NOT NULL DEFAULT 0,
    unallocated_amt   NUMERIC(18,2) GENERATED ALWAYS AS (amount - allocated_amount) STORED,
    receipt_date      DATE         NOT NULL DEFAULT CURRENT_DATE,
    reference         VARCHAR(100),
    payment_method    VARCHAR(50)  NOT NULL DEFAULT 'bank_transfer',
    currency          VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    bank_account_id   UUID         REFERENCES bank_accounts(id) ON DELETE SET NULL,
    cash_account_id   UUID         REFERENCES cash_accounts(id) ON DELETE SET NULL,
    description       TEXT,
    invoice_id        UUID,
    invoice_number    VARCHAR(50),
    confirmed_at      TIMESTAMPTZ,
    notes             TEXT,
    created_by        UUID         REFERENCES users(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE INDEX IF NOT EXISTS idx_receipts_company   ON receipts(company_id);
CREATE INDEX IF NOT EXISTS idx_receipts_date      ON receipts(receipt_date DESC);
CREATE INDEX IF NOT EXISTS idx_receipts_status    ON receipts(status);
CREATE INDEX IF NOT EXISTS idx_receipts_partner   ON receipts(partner_id);

-- ─────────────────────────────────────────────────────────────────────────────
-- PART 4: FIX BANK_RECONCILIATIONS TABLE
-- Original: bank_account_id, period_date, statement_balance, book_balance,
--           difference(generated), is_reconciled, reconciled_at, notes, created_at
-- Handler uses: reference, status, period_start, period_end, company_id,
--               opening_balance, closing_balance, system_balance, matched_items,
--               unmatched_items, notes, created_at
-- ─────────────────────────────────────────────────────────────────────────────

-- Add company_id
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='bank_reconciliations' AND column_name='company_id') THEN
    ALTER TABLE bank_reconciliations ADD COLUMN company_id UUID REFERENCES companies(id) ON DELETE CASCADE;
    -- Populate company_id from bank_accounts
    UPDATE bank_reconciliations br
    SET company_id = ba.company_id
    FROM bank_accounts ba
    WHERE ba.id = br.bank_account_id AND br.company_id IS NULL;
  END IF;
END $$;

-- Add reference
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='bank_reconciliations' AND column_name='reference') THEN
    ALTER TABLE bank_reconciliations ADD COLUMN reference VARCHAR(50);
    -- Generate references for existing rows
    UPDATE bank_reconciliations SET reference = 'RECON-' || EXTRACT(YEAR FROM COALESCE(period_date, NOW()))::TEXT || '-' || (EXTRACT(MONTH FROM COALESCE(period_date, NOW()))::TEXT) WHERE reference IS NULL;
  END IF;
END $$;

-- Add status
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='bank_reconciliations' AND column_name='status') THEN
    ALTER TABLE bank_reconciliations ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'draft';
    UPDATE bank_reconciliations SET status = CASE WHEN is_reconciled THEN 'reconciled' ELSE 'draft' END;
  END IF;
END $$;

-- Add period_start
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='bank_reconciliations' AND column_name='period_start') THEN
    ALTER TABLE bank_reconciliations ADD COLUMN period_start DATE;
    UPDATE bank_reconciliations SET period_start = date_trunc('month', COALESCE(period_date, NOW()))::DATE WHERE period_start IS NULL;
  END IF;
END $$;

-- Add period_end
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='bank_reconciliations' AND column_name='period_end') THEN
    ALTER TABLE bank_reconciliations ADD COLUMN period_end DATE;
    UPDATE bank_reconciliations SET period_end = COALESCE(period_date, NOW()::DATE) WHERE period_end IS NULL;
  END IF;
END $$;

-- Add opening_balance
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='bank_reconciliations' AND column_name='opening_balance') THEN
    ALTER TABLE bank_reconciliations ADD COLUMN opening_balance NUMERIC(18,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

-- Add closing_balance
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='bank_reconciliations' AND column_name='closing_balance') THEN
    ALTER TABLE bank_reconciliations ADD COLUMN closing_balance NUMERIC(18,2) NOT NULL DEFAULT 0;
    UPDATE bank_reconciliations SET closing_balance = book_balance WHERE closing_balance = 0 AND book_balance IS NOT NULL;
  END IF;
END $$;

-- Add system_balance
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='bank_reconciliations' AND column_name='system_balance') THEN
    ALTER TABLE bank_reconciliations ADD COLUMN system_balance NUMERIC(18,2) NOT NULL DEFAULT 0;
    UPDATE bank_reconciliations SET system_balance = book_balance WHERE system_balance = 0 AND book_balance IS NOT NULL;
  END IF;
END $$;

-- Add matched_items
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='bank_reconciliations' AND column_name='matched_items') THEN
    ALTER TABLE bank_reconciliations ADD COLUMN matched_items INT NOT NULL DEFAULT 0;
  END IF;
END $$;

-- Add unmatched_items
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='bank_reconciliations' AND column_name='unmatched_items') THEN
    ALTER TABLE bank_reconciliations ADD COLUMN unmatched_items INT NOT NULL DEFAULT 0;
  END IF;
END $$;

-- Add updated_at
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                 WHERE table_name='bank_reconciliations' AND column_name='updated_at') THEN
    ALTER TABLE bank_reconciliations ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
  END IF;
END $$;

-- Create bank_reconciliation_lines if not exists (handler uses this table)
CREATE TABLE IF NOT EXISTS bank_reconciliation_lines (
    id                    UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    reconciliation_id     UUID         NOT NULL REFERENCES bank_reconciliations(id) ON DELETE CASCADE,
    line_type             VARCHAR(20)  NOT NULL DEFAULT 'movement',
    amount                NUMERIC(18,2) NOT NULL DEFAULT 0,
    description           TEXT,
    reference             VARCHAR(100),
    transaction_date      DATE,
    is_matched            BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bank_recon_lines_recon ON bank_reconciliation_lines(reconciliation_id);

-- Add indexes for bank_reconciliations
CREATE INDEX IF NOT EXISTS idx_bank_recon_company    ON bank_reconciliations(company_id);
CREATE INDEX IF NOT EXISTS idx_bank_recon_account    ON bank_reconciliations(bank_account_id);
CREATE INDEX IF NOT EXISTS idx_bank_recon_status     ON bank_reconciliations(status);

-- ─────────────────────────────────────────────────────────────────────────────
-- PART 5: HELPER FUNCTIONS
-- ─────────────────────────────────────────────────────────────────────────────

-- next_receipt_number function
CREATE OR REPLACE FUNCTION next_receipt_number(p_company_id UUID)
RETURNS TEXT AS $$
DECLARE
    v_count  INT;
    v_prefix TEXT := 'RCT';
    v_year   TEXT := TO_CHAR(NOW(), 'YYYY');
BEGIN
    SELECT COUNT(*) + 1 INTO v_count
    FROM receipts
    WHERE company_id = p_company_id
      AND EXTRACT(YEAR FROM receipt_date) = EXTRACT(YEAR FROM NOW());
    RETURN v_prefix || '-' || v_year || '-' || LPAD(v_count::TEXT, 5, '0');
END;
$$ LANGUAGE plpgsql;

-- next_bank_recon_reference function
CREATE OR REPLACE FUNCTION next_bank_recon_reference(p_company_id UUID)
RETURNS TEXT AS $$
DECLARE
    v_count  INT;
    v_year   TEXT := TO_CHAR(NOW(), 'YYYY');
BEGIN
    SELECT COUNT(*) + 1 INTO v_count
    FROM bank_reconciliations
    WHERE company_id = p_company_id
      AND EXTRACT(YEAR FROM COALESCE(period_end, created_at::DATE)) = EXTRACT(YEAR FROM NOW());
    RETURN 'RECON-' || v_year || '-' || LPAD(v_count::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;

-- ─────────────────────────────────────────────────────────────────────────────
-- PART 6: SYSTEM DIAGNOSTICS & ERROR TRACKING
-- ─────────────────────────────────────────────────────────────────────────────

-- Error severity enum
DO $$ BEGIN
  CREATE TYPE error_severity AS ENUM ('debug', 'info', 'warning', 'error', 'critical');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Error source enum
DO $$ BEGIN
  CREATE TYPE error_source AS ENUM (
    'frontend_js', 'backend_go', 'database_sql',
    'api_http', 'auth', 'background_job', 'migration', 'system'
  );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Main system_logs table
CREATE TABLE IF NOT EXISTS system_logs (
    id              UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID            REFERENCES companies(id) ON DELETE SET NULL,
    user_id         UUID            REFERENCES users(id) ON DELETE SET NULL,
    severity        error_severity  NOT NULL DEFAULT 'error',
    source          error_source    NOT NULL DEFAULT 'backend_go',
    module          VARCHAR(100),
    endpoint        VARCHAR(500),
    method          VARCHAR(10),
    http_status     INT,
    request_id      UUID            DEFAULT uuid_generate_v4(),
    correlation_id  VARCHAR(100),
    message         TEXT            NOT NULL,
    error_code      VARCHAR(100),
    sql_state       VARCHAR(10),
    stack_trace     TEXT,
    request_body    JSONB,
    response_body   JSONB,
    ip_address      INET,
    user_agent      VARCHAR(500),
    page_url        VARCHAR(1000),
    duration_ms     INT,
    is_resolved     BOOLEAN         NOT NULL DEFAULT FALSE,
    resolved_at     TIMESTAMPTZ,
    resolved_by     UUID            REFERENCES users(id) ON DELETE SET NULL,
    resolution_note TEXT,
    tags            TEXT[],
    extra           JSONB,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_syslog_company    ON system_logs(company_id);
CREATE INDEX IF NOT EXISTS idx_syslog_severity   ON system_logs(severity);
CREATE INDEX IF NOT EXISTS idx_syslog_source     ON system_logs(source);
CREATE INDEX IF NOT EXISTS idx_syslog_module     ON system_logs(module);
CREATE INDEX IF NOT EXISTS idx_syslog_created    ON system_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_syslog_status     ON system_logs(http_status);
CREATE INDEX IF NOT EXISTS idx_syslog_resolved   ON system_logs(is_resolved);
CREATE INDEX IF NOT EXISTS idx_syslog_request_id ON system_logs(request_id);
CREATE INDEX IF NOT EXISTS idx_syslog_user       ON system_logs(user_id);

-- Error stats view
CREATE OR REPLACE VIEW v_error_stats AS
SELECT
    company_id,
    date_trunc('hour', created_at)                          AS hour,
    severity,
    source,
    module,
    COUNT(*)                                                 AS count,
    COUNT(*) FILTER (WHERE is_resolved = FALSE)             AS unresolved_count,
    AVG(duration_ms)                                         AS avg_duration_ms,
    MAX(created_at)                                          AS last_occurrence
FROM system_logs
GROUP BY company_id, date_trunc('hour', created_at), severity, source, module;

-- Daily error summary view
CREATE OR REPLACE VIEW v_daily_error_summary AS
SELECT
    company_id,
    created_at::DATE                                         AS log_date,
    COUNT(*)                                                 AS total,
    COUNT(*) FILTER (WHERE severity = 'critical')            AS critical_count,
    COUNT(*) FILTER (WHERE severity = 'error')               AS error_count,
    COUNT(*) FILTER (WHERE severity = 'warning')             AS warning_count,
    COUNT(*) FILTER (WHERE severity = 'info')                AS info_count,
    COUNT(*) FILTER (WHERE is_resolved = TRUE)               AS resolved_count,
    COUNT(*) FILTER (WHERE is_resolved = FALSE
                       AND severity IN ('error','critical'))  AS open_errors
FROM system_logs
GROUP BY company_id, created_at::DATE;

-- Top errors view
CREATE OR REPLACE VIEW v_top_errors AS
SELECT
    company_id,
    module,
    message,
    error_code,
    http_status,
    source,
    severity,
    COUNT(*)                AS occurrence_count,
    MIN(created_at)         AS first_seen,
    MAX(created_at)         AS last_seen,
    COUNT(*) FILTER (WHERE is_resolved = FALSE) AS unresolved
FROM system_logs
WHERE severity IN ('error', 'critical')
GROUP BY company_id, module, message, error_code, http_status, source, severity
ORDER BY occurrence_count DESC;

-- ─────────────────────────────────────────────────────────────────────────────
-- PART 7: TREASURY MOVEMENTS TABLE (if not exists — handler also uses it)
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS treasury_movements (
    id              UUID            PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID            NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    type            VARCHAR(30)     NOT NULL DEFAULT 'debit',
    amount          NUMERIC(18,2)   NOT NULL DEFAULT 0,
    date            DATE            NOT NULL DEFAULT CURRENT_DATE,
    reference       VARCHAR(100),
    notes           TEXT,
    category        VARCHAR(100),
    reconciled      BOOLEAN         NOT NULL DEFAULT FALSE,
    cash_account_id UUID            REFERENCES cash_accounts(id) ON DELETE SET NULL,
    bank_account_id UUID            REFERENCES bank_accounts(id)  ON DELETE SET NULL,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_treas_mov_company ON treasury_movements(company_id);
CREATE INDEX IF NOT EXISTS idx_treas_mov_date    ON treasury_movements(date DESC);

-- ─────────────────────────────────────────────────────────────────────────────
-- TRACK MIGRATION
-- ─────────────────────────────────────────────────────────────────────────────
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0010_treasury_fix_diagnostics') THEN
        INSERT INTO schema_migrations (version) VALUES ('0010_treasury_fix_diagnostics');
    END IF;
END $$;
