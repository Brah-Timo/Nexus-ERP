-- Migration: 0020_budgeting_planning.sql
-- Budgeting & Planning Module
-- Idempotent: safe to run multiple times

-- ============================================================
-- ENUMS
-- ============================================================

DO $$ BEGIN
  CREATE TYPE budget_status AS ENUM ('draft','active','locked','closed','cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE budget_type AS ENUM ('operational','capital','project','department');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE revision_type AS ENUM ('increase','decrease','reallocation','correction');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE commitment_status AS ENUM ('pending','approved','fulfilled','cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE commitment_type AS ENUM ('purchase_order','contract','reservation','other');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

-- ============================================================
-- SEQUENCES
-- ============================================================

CREATE SEQUENCE IF NOT EXISTS budget_number_seq START 1000;
CREATE SEQUENCE IF NOT EXISTS revision_number_seq START 100;
CREATE SEQUENCE IF NOT EXISTS commitment_number_seq START 100;

-- ============================================================
-- TABLES
-- ============================================================

-- Budget Categories
CREATE TABLE IF NOT EXISTS budget_categories (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id    UUID NOT NULL,
  code          VARCHAR(50) NOT NULL,
  name          VARCHAR(255) NOT NULL,
  description   TEXT,
  parent_id     UUID REFERENCES budget_categories(id) ON DELETE SET NULL,
  is_active     BOOLEAN NOT NULL DEFAULT true,
  sort_order    INTEGER NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, code)
);

-- Annual Budgets
CREATE TABLE IF NOT EXISTS annual_budgets (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id      UUID NOT NULL,
  budget_number   VARCHAR(50) NOT NULL,
  fiscal_year     INTEGER NOT NULL,
  name            VARCHAR(255) NOT NULL,
  description     TEXT,
  budget_type     budget_type NOT NULL DEFAULT 'operational',
  status          budget_status NOT NULL DEFAULT 'draft',
  start_date      DATE NOT NULL,
  end_date        DATE NOT NULL,
  total_amount    NUMERIC(20,4) NOT NULL DEFAULT 0,
  approved_by     UUID,
  approved_at     TIMESTAMPTZ,
  notes           TEXT,
  created_by      UUID,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, budget_number)
);

-- Department Budgets (allocations per department per annual budget)
CREATE TABLE IF NOT EXISTS department_budgets (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id        UUID NOT NULL,
  annual_budget_id  UUID NOT NULL REFERENCES annual_budgets(id) ON DELETE CASCADE,
  department_id     UUID,
  department_name   VARCHAR(255) NOT NULL,
  department_code   VARCHAR(50),
  allocated_amount  NUMERIC(20,4) NOT NULL DEFAULT 0,
  spent_amount      NUMERIC(20,4) NOT NULL DEFAULT 0,
  committed_amount  NUMERIC(20,4) NOT NULL DEFAULT 0,
  notes             TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Budget Line Items (per category per annual budget or department budget)
CREATE TABLE IF NOT EXISTS budget_line_items (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id            UUID NOT NULL,
  annual_budget_id      UUID NOT NULL REFERENCES annual_budgets(id) ON DELETE CASCADE,
  department_budget_id  UUID REFERENCES department_budgets(id) ON DELETE SET NULL,
  category_id           UUID REFERENCES budget_categories(id) ON DELETE SET NULL,
  account_code          VARCHAR(50),
  account_name          VARCHAR(255),
  description           TEXT,
  budget_amount         NUMERIC(20,4) NOT NULL DEFAULT 0,
  q1_amount             NUMERIC(20,4) NOT NULL DEFAULT 0,
  q2_amount             NUMERIC(20,4) NOT NULL DEFAULT 0,
  q3_amount             NUMERIC(20,4) NOT NULL DEFAULT 0,
  q4_amount             NUMERIC(20,4) NOT NULL DEFAULT 0,
  actual_amount         NUMERIC(20,4) NOT NULL DEFAULT 0,
  committed_amount      NUMERIC(20,4) NOT NULL DEFAULT 0,
  notes                 TEXT,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Budget Revisions
CREATE TABLE IF NOT EXISTS budget_revisions (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id        UUID NOT NULL,
  revision_number   VARCHAR(50) NOT NULL,
  annual_budget_id  UUID NOT NULL REFERENCES annual_budgets(id) ON DELETE CASCADE,
  line_item_id      UUID REFERENCES budget_line_items(id) ON DELETE SET NULL,
  department_budget_id UUID REFERENCES department_budgets(id) ON DELETE SET NULL,
  revision_type     revision_type NOT NULL,
  original_amount   NUMERIC(20,4) NOT NULL DEFAULT 0,
  revised_amount    NUMERIC(20,4) NOT NULL DEFAULT 0,
  change_amount     NUMERIC(20,4) GENERATED ALWAYS AS (revised_amount - original_amount) STORED,
  reason            TEXT NOT NULL,
  status            budget_status NOT NULL DEFAULT 'draft',
  requested_by      UUID,
  approved_by       UUID,
  approved_at       TIMESTAMPTZ,
  effective_date    DATE,
  notes             TEXT,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, revision_number)
);

-- Budget Commitments (encumbrances / pre-commitments)
CREATE TABLE IF NOT EXISTS budget_commitments (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id            UUID NOT NULL,
  commitment_number     VARCHAR(50) NOT NULL,
  annual_budget_id      UUID NOT NULL REFERENCES annual_budgets(id) ON DELETE CASCADE,
  department_budget_id  UUID REFERENCES department_budgets(id) ON DELETE SET NULL,
  line_item_id          UUID REFERENCES budget_line_items(id) ON DELETE SET NULL,
  commitment_type       commitment_type NOT NULL DEFAULT 'purchase_order',
  status                commitment_status NOT NULL DEFAULT 'pending',
  reference_number      VARCHAR(100),
  vendor_name           VARCHAR(255),
  description           TEXT NOT NULL,
  committed_amount      NUMERIC(20,4) NOT NULL DEFAULT 0,
  fulfilled_amount      NUMERIC(20,4) NOT NULL DEFAULT 0,
  remaining_amount      NUMERIC(20,4) GENERATED ALWAYS AS (committed_amount - fulfilled_amount) STORED,
  commitment_date       DATE NOT NULL,
  expected_fulfillment  DATE,
  approved_by           UUID,
  approved_at           TIMESTAMPTZ,
  notes                 TEXT,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, commitment_number)
);

-- Budget Actuals (actual spend transactions linked to budget)
CREATE TABLE IF NOT EXISTS budget_actuals (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id            UUID NOT NULL,
  annual_budget_id      UUID NOT NULL REFERENCES annual_budgets(id) ON DELETE CASCADE,
  department_budget_id  UUID REFERENCES department_budgets(id) ON DELETE SET NULL,
  line_item_id          UUID REFERENCES budget_line_items(id) ON DELETE SET NULL,
  commitment_id         UUID REFERENCES budget_commitments(id) ON DELETE SET NULL,
  transaction_date      DATE NOT NULL,
  reference_type        VARCHAR(50),
  reference_id          UUID,
  reference_number      VARCHAR(100),
  description           TEXT,
  amount                NUMERIC(20,4) NOT NULL DEFAULT 0,
  posted               BOOLEAN NOT NULL DEFAULT false,
  posted_at             TIMESTAMPTZ,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- INDEXES
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_budget_categories_company ON budget_categories(company_id);
CREATE INDEX IF NOT EXISTS idx_budget_categories_parent ON budget_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_annual_budgets_company ON annual_budgets(company_id);
CREATE INDEX IF NOT EXISTS idx_annual_budgets_fiscal_year ON annual_budgets(company_id, fiscal_year);
CREATE INDEX IF NOT EXISTS idx_annual_budgets_status ON annual_budgets(company_id, status);
CREATE INDEX IF NOT EXISTS idx_department_budgets_company ON department_budgets(company_id);
CREATE INDEX IF NOT EXISTS idx_department_budgets_annual ON department_budgets(annual_budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_line_items_company ON budget_line_items(company_id);
CREATE INDEX IF NOT EXISTS idx_budget_line_items_annual ON budget_line_items(annual_budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_line_items_dept ON budget_line_items(department_budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_line_items_category ON budget_line_items(category_id);
CREATE INDEX IF NOT EXISTS idx_budget_revisions_company ON budget_revisions(company_id);
CREATE INDEX IF NOT EXISTS idx_budget_revisions_annual ON budget_revisions(annual_budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_commitments_company ON budget_commitments(company_id);
CREATE INDEX IF NOT EXISTS idx_budget_commitments_annual ON budget_commitments(annual_budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_commitments_status ON budget_commitments(company_id, status);
CREATE INDEX IF NOT EXISTS idx_budget_actuals_company ON budget_actuals(company_id);
CREATE INDEX IF NOT EXISTS idx_budget_actuals_annual ON budget_actuals(annual_budget_id);
CREATE INDEX IF NOT EXISTS idx_budget_actuals_date ON budget_actuals(company_id, transaction_date);

-- ============================================================
-- TRIGGERS: updated_at
-- ============================================================

CREATE OR REPLACE FUNCTION set_budgeting_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_budget_categories_updated_at ON budget_categories;
CREATE TRIGGER trg_budget_categories_updated_at
  BEFORE UPDATE ON budget_categories
  FOR EACH ROW EXECUTE FUNCTION set_budgeting_updated_at();

DROP TRIGGER IF EXISTS trg_annual_budgets_updated_at ON annual_budgets;
CREATE TRIGGER trg_annual_budgets_updated_at
  BEFORE UPDATE ON annual_budgets
  FOR EACH ROW EXECUTE FUNCTION set_budgeting_updated_at();

DROP TRIGGER IF EXISTS trg_department_budgets_updated_at ON department_budgets;
CREATE TRIGGER trg_department_budgets_updated_at
  BEFORE UPDATE ON department_budgets
  FOR EACH ROW EXECUTE FUNCTION set_budgeting_updated_at();

DROP TRIGGER IF EXISTS trg_budget_line_items_updated_at ON budget_line_items;
CREATE TRIGGER trg_budget_line_items_updated_at
  BEFORE UPDATE ON budget_line_items
  FOR EACH ROW EXECUTE FUNCTION set_budgeting_updated_at();

DROP TRIGGER IF EXISTS trg_budget_revisions_updated_at ON budget_revisions;
CREATE TRIGGER trg_budget_revisions_updated_at
  BEFORE UPDATE ON budget_revisions
  FOR EACH ROW EXECUTE FUNCTION set_budgeting_updated_at();

DROP TRIGGER IF EXISTS trg_budget_commitments_updated_at ON budget_commitments;
CREATE TRIGGER trg_budget_commitments_updated_at
  BEFORE UPDATE ON budget_commitments
  FOR EACH ROW EXECUTE FUNCTION set_budgeting_updated_at();

DROP TRIGGER IF EXISTS trg_budget_actuals_updated_at ON budget_actuals;
CREATE TRIGGER trg_budget_actuals_updated_at
  BEFORE UPDATE ON budget_actuals
  FOR EACH ROW EXECUTE FUNCTION set_budgeting_updated_at();
