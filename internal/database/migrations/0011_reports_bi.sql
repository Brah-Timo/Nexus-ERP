-- ============================================================
-- Migration 0011 — Reports & BI
-- Idempotent: safe to run multiple times
-- ============================================================

-- ─────────────────────────────────────────────────────────────
-- PART 1 — Report Definitions (saved report configurations)
-- ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS report_definitions (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    report_type     VARCHAR(50)  NOT NULL,  -- financial, sales, purchase, inventory, project, management, analytics
    category        VARCHAR(50)  NOT NULL DEFAULT 'custom',
    query_config    JSONB        NOT NULL DEFAULT '{}',
    filters         JSONB        NOT NULL DEFAULT '{}',
    columns         JSONB        NOT NULL DEFAULT '[]',
    chart_config    JSONB        NOT NULL DEFAULT '{}',
    schedule        VARCHAR(20),            -- daily, weekly, monthly, none
    is_public       BOOLEAN      NOT NULL DEFAULT FALSE,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────
-- PART 2 — Dashboard Widgets
-- ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS dashboard_widgets (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id         UUID        REFERENCES users(id) ON DELETE CASCADE,
    widget_type     VARCHAR(50)  NOT NULL,   -- kpi, bar_chart, line_chart, pie_chart, table, gauge
    title           VARCHAR(200) NOT NULL,
    data_source     VARCHAR(100) NOT NULL,   -- endpoint key
    config          JSONB        NOT NULL DEFAULT '{}',
    position_x      INT          NOT NULL DEFAULT 0,
    position_y      INT          NOT NULL DEFAULT 0,
    width           INT          NOT NULL DEFAULT 4,
    height          INT          NOT NULL DEFAULT 3,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ─────────────────────────────────────────────────────────────
-- PART 3 — Report Cache (for expensive aggregations)
-- ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS report_cache (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    cache_key       VARCHAR(500) NOT NULL,
    report_type     VARCHAR(50)  NOT NULL,
    period_start    DATE,
    period_end      DATE,
    data            JSONB        NOT NULL DEFAULT '{}',
    generated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ  NOT NULL DEFAULT (NOW() + INTERVAL '1 hour'),
    UNIQUE (company_id, cache_key)
);

-- ─────────────────────────────────────────────────────────────
-- PART 4 — KPI Snapshots (daily snapshots for trend analysis)
-- ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS kpi_snapshots (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    snapshot_date   DATE        NOT NULL,
    kpi_key         VARCHAR(100) NOT NULL,
    kpi_value       NUMERIC(20,4),
    kpi_unit        VARCHAR(20)  DEFAULT '',
    dimension       VARCHAR(100) DEFAULT '',  -- e.g., product category, region, department
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (company_id, snapshot_date, kpi_key, dimension)
);

-- ─────────────────────────────────────────────────────────────
-- PART 5 — Add indexes
-- ─────────────────────────────────────────────────────────────
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_report_definitions_company') THEN
    CREATE INDEX idx_report_definitions_company ON report_definitions(company_id, report_type);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_dashboard_widgets_company_user') THEN
    CREATE INDEX idx_dashboard_widgets_company_user ON dashboard_widgets(company_id, user_id);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_report_cache_key') THEN
    CREATE INDEX idx_report_cache_key ON report_cache(company_id, cache_key, expires_at);
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_kpi_snapshots_date') THEN
    CREATE INDEX idx_kpi_snapshots_date ON kpi_snapshots(company_id, snapshot_date DESC, kpi_key);
  END IF;
END $$;

-- ─────────────────────────────────────────────────────────────
-- PART 6 — Ensure invoice_lines has all needed columns
-- ─────────────────────────────────────────────────────────────
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='invoice_lines' AND column_name='discount_pct'
  ) THEN
    ALTER TABLE invoice_lines ADD COLUMN discount_pct NUMERIC(5,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='invoice_lines' AND column_name='tax_amount'
  ) THEN
    ALTER TABLE invoice_lines ADD COLUMN tax_amount NUMERIC(18,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

-- ─────────────────────────────────────────────────────────────
-- PART 7 — Ensure purchase_invoice_lines has needed columns
-- ─────────────────────────────────────────────────────────────
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='purchase_invoice_lines' AND column_name='discount_pct'
  ) THEN
    ALTER TABLE purchase_invoice_lines ADD COLUMN discount_pct NUMERIC(5,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='purchase_invoice_lines' AND column_name='tax_amount'
  ) THEN
    ALTER TABLE purchase_invoice_lines ADD COLUMN tax_amount NUMERIC(18,2) NOT NULL DEFAULT 0;
  END IF;
END $$;

-- ─────────────────────────────────────────────────────────────
-- PART 8 — Ensure inventory items have cost_price
-- ─────────────────────────────────────────────────────────────
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='items' AND column_name='cost_price'
  ) THEN
    ALTER TABLE items ADD COLUMN cost_price NUMERIC(18,4) NOT NULL DEFAULT 0;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='items' AND column_name='reorder_point'
  ) THEN
    ALTER TABLE items ADD COLUMN reorder_point NUMERIC(18,4) NOT NULL DEFAULT 0;
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name='items' AND column_name='max_stock'
  ) THEN
    ALTER TABLE items ADD COLUMN max_stock NUMERIC(18,4) NOT NULL DEFAULT 0;
  END IF;
END $$;

-- ─────────────────────────────────────────────────────────────
-- PART 9 — Seed default widgets for new installs
-- ─────────────────────────────────────────────────────────────
-- (No seed data required — widgets are user-configurable)

-- ─────────────────────────────────────────────────────────────
-- Done
-- ─────────────────────────────────────────────────────────────
