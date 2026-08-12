-- ============================================================
-- Migration 0019: Assets Management Module
-- Idempotent: safe to run multiple times
-- ============================================================

-- ── ENUMs ────────────────────────────────────────────────────────────────────

DO $$ BEGIN
  CREATE TYPE asset_status AS ENUM (
    'active','in_use','in_storage','under_maintenance','disposed','sold','written_off'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE asset_condition AS ENUM ('excellent','good','fair','poor','damaged');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE depreciation_method AS ENUM (
    'straight_line','declining_balance','double_declining','sum_of_years','units_of_production'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE transfer_status AS ENUM ('pending','approved','in_transit','completed','cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE maintenance_type AS ENUM (
    'preventive','corrective','inspection','upgrade','repair','calibration'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE maintenance_status AS ENUM (
    'scheduled','in_progress','completed','cancelled','overdue'
  );
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

-- ── Asset Categories ──────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS asset_categories (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id           UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name                 TEXT NOT NULL,
  description          TEXT NOT NULL DEFAULT '',
  parent_id            UUID REFERENCES asset_categories(id) ON DELETE SET NULL,
  depreciation_method  depreciation_method NOT NULL DEFAULT 'straight_line',
  useful_life_years    NUMERIC(6,2) NOT NULL DEFAULT 5,
  salvage_value_pct    NUMERIC(5,2) NOT NULL DEFAULT 10,
  depreciation_rate    NUMERIC(7,4) NOT NULL DEFAULT 20,
  gl_asset_account     TEXT NOT NULL DEFAULT '',
  gl_depreciation_account TEXT NOT NULL DEFAULT '',
  gl_accumulated_account  TEXT NOT NULL DEFAULT '',
  color                TEXT NOT NULL DEFAULT '#6366f1',
  is_active            BOOLEAN NOT NULL DEFAULT TRUE,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_asset_cat_company ON asset_categories(company_id);
CREATE INDEX IF NOT EXISTS idx_asset_cat_parent  ON asset_categories(parent_id);

-- ── Asset Locations ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS asset_locations (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  address     TEXT NOT NULL DEFAULT '',
  city        TEXT NOT NULL DEFAULT '',
  country     TEXT NOT NULL DEFAULT '',
  parent_id   UUID REFERENCES asset_locations(id) ON DELETE SET NULL,
  manager     TEXT NOT NULL DEFAULT '',
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_asset_loc_company ON asset_locations(company_id);

-- ── Fixed Assets ──────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS fixed_assets (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id           UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  asset_number         TEXT NOT NULL,
  name                 TEXT NOT NULL,
  description          TEXT NOT NULL DEFAULT '',
  category_id          UUID REFERENCES asset_categories(id) ON DELETE SET NULL,
  location_id          UUID REFERENCES asset_locations(id) ON DELETE SET NULL,
  status               asset_status NOT NULL DEFAULT 'active',
  condition            asset_condition NOT NULL DEFAULT 'good',
  serial_number        TEXT NOT NULL DEFAULT '',
  barcode              TEXT NOT NULL DEFAULT '',
  brand                TEXT NOT NULL DEFAULT '',
  model                TEXT NOT NULL DEFAULT '',
  supplier             TEXT NOT NULL DEFAULT '',
  purchase_date        DATE,
  purchase_cost        NUMERIC(18,4) NOT NULL DEFAULT 0,
  salvage_value        NUMERIC(18,4) NOT NULL DEFAULT 0,
  useful_life_years    NUMERIC(6,2) NOT NULL DEFAULT 5,
  depreciation_method  depreciation_method NOT NULL DEFAULT 'straight_line',
  depreciation_rate    NUMERIC(7,4) NOT NULL DEFAULT 20,
  accumulated_depreciation NUMERIC(18,4) NOT NULL DEFAULT 0,
  net_book_value       NUMERIC(18,4) GENERATED ALWAYS AS (purchase_cost - accumulated_depreciation) STORED,
  depreciation_start   DATE,
  last_depreciation_date DATE,
  disposal_date        DATE,
  disposal_value       NUMERIC(18,4) NOT NULL DEFAULT 0,
  disposal_reason      TEXT NOT NULL DEFAULT '',
  assigned_to          TEXT NOT NULL DEFAULT '',
  warranty_expiry      DATE,
  insurance_policy     TEXT NOT NULL DEFAULT '',
  insurance_expiry     DATE,
  notes                TEXT NOT NULL DEFAULT '',
  tags                 TEXT[] NOT NULL DEFAULT '{}',
  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_fa_company    ON fixed_assets(company_id);
CREATE INDEX IF NOT EXISTS idx_fa_category   ON fixed_assets(category_id);
CREATE INDEX IF NOT EXISTS idx_fa_location   ON fixed_assets(location_id);
CREATE INDEX IF NOT EXISTS idx_fa_status     ON fixed_assets(company_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_fa_number ON fixed_assets(company_id, asset_number);

-- ── Depreciation Schedules ────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS asset_depreciation_schedules (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id           UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  asset_id             UUID NOT NULL REFERENCES fixed_assets(id) ON DELETE CASCADE,
  period_year          INTEGER NOT NULL,
  period_month         INTEGER NOT NULL,
  period_label         TEXT NOT NULL DEFAULT '',
  opening_book_value   NUMERIC(18,4) NOT NULL DEFAULT 0,
  depreciation_amount  NUMERIC(18,4) NOT NULL DEFAULT 0,
  accumulated_depreciation NUMERIC(18,4) NOT NULL DEFAULT 0,
  closing_book_value   NUMERIC(18,4) NOT NULL DEFAULT 0,
  is_posted            BOOLEAN NOT NULL DEFAULT FALSE,
  posted_at            TIMESTAMPTZ,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_dep_sched_asset   ON asset_depreciation_schedules(asset_id);
CREATE INDEX IF NOT EXISTS idx_dep_sched_company ON asset_depreciation_schedules(company_id);
CREATE INDEX IF NOT EXISTS idx_dep_sched_period  ON asset_depreciation_schedules(company_id, period_year, period_month);

-- ── Asset Transfers ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS asset_transfers (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id       UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  transfer_number  TEXT NOT NULL,
  asset_id         UUID NOT NULL REFERENCES fixed_assets(id) ON DELETE CASCADE,
  from_location_id UUID REFERENCES asset_locations(id) ON DELETE SET NULL,
  to_location_id   UUID REFERENCES asset_locations(id) ON DELETE SET NULL,
  from_custodian   TEXT NOT NULL DEFAULT '',
  to_custodian     TEXT NOT NULL DEFAULT '',
  transfer_date    DATE NOT NULL DEFAULT CURRENT_DATE,
  reason           TEXT NOT NULL DEFAULT '',
  status           transfer_status NOT NULL DEFAULT 'pending',
  approved_by      TEXT NOT NULL DEFAULT '',
  approved_at      TIMESTAMPTZ,
  completed_at     TIMESTAMPTZ,
  notes            TEXT NOT NULL DEFAULT '',
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_at_company ON asset_transfers(company_id);
CREATE INDEX IF NOT EXISTS idx_at_asset   ON asset_transfers(asset_id);
CREATE INDEX IF NOT EXISTS idx_at_status  ON asset_transfers(company_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_at_number ON asset_transfers(company_id, transfer_number);

-- ── Asset Maintenance ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS asset_maintenance_records (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id       UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  asset_id         UUID NOT NULL REFERENCES fixed_assets(id) ON DELETE CASCADE,
  maintenance_type maintenance_type NOT NULL DEFAULT 'preventive',
  status           maintenance_status NOT NULL DEFAULT 'scheduled',
  title            TEXT NOT NULL,
  description      TEXT NOT NULL DEFAULT '',
  scheduled_date   DATE,
  started_at       TIMESTAMPTZ,
  completed_at     TIMESTAMPTZ,
  performed_by     TEXT NOT NULL DEFAULT '',
  vendor           TEXT NOT NULL DEFAULT '',
  cost             NUMERIC(18,4) NOT NULL DEFAULT 0,
  downtime_hours   NUMERIC(8,2) NOT NULL DEFAULT 0,
  next_maintenance_date DATE,
  findings         TEXT NOT NULL DEFAULT '',
  actions_taken    TEXT NOT NULL DEFAULT '',
  parts_replaced   TEXT NOT NULL DEFAULT '',
  warranty_claim   BOOLEAN NOT NULL DEFAULT FALSE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_am_company ON asset_maintenance_records(company_id);
CREATE INDEX IF NOT EXISTS idx_am_asset   ON asset_maintenance_records(asset_id);
CREATE INDEX IF NOT EXISTS idx_am_status  ON asset_maintenance_records(company_id, status);

-- ── Asset sequence ────────────────────────────────────────────────────────────

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_sequences WHERE schemaname='public' AND sequencename='asset_number_seq'
  ) THEN
    CREATE SEQUENCE asset_number_seq START 1000 INCREMENT 1;
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_sequences WHERE schemaname='public' AND sequencename='asset_transfer_seq'
  ) THEN
    CREATE SEQUENCE asset_transfer_seq START 100 INCREMENT 1;
  END IF;
END$$;

-- ── updated_at trigger ────────────────────────────────────────────────────────

CREATE OR REPLACE FUNCTION update_assets_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$;

DO $$ BEGIN
  CREATE TRIGGER trg_asset_cat_updated
    BEFORE UPDATE ON asset_categories
    FOR EACH ROW EXECUTE FUNCTION update_assets_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TRIGGER trg_asset_loc_updated
    BEFORE UPDATE ON asset_locations
    FOR EACH ROW EXECUTE FUNCTION update_assets_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TRIGGER trg_fixed_assets_updated
    BEFORE UPDATE ON fixed_assets
    FOR EACH ROW EXECUTE FUNCTION update_assets_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TRIGGER trg_asset_transfers_updated
    BEFORE UPDATE ON asset_transfers
    FOR EACH ROW EXECUTE FUNCTION update_assets_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TRIGGER trg_asset_maint_updated
    BEFORE UPDATE ON asset_maintenance_records
    FOR EACH ROW EXECUTE FUNCTION update_assets_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL; END$$;
