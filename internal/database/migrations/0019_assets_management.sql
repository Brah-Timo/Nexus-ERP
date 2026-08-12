-- ============================================================
-- Migration 0019: Assets Management Module (Complete Version with Drop & Recreate)
-- Idempotent - Safe to run multiple times
-- ============================================================

-- ── Drop all related tables and types ─────────────────────────────────────────
-- Tables are dropped first, then types to avoid conflicts

-- Drop tables (in reverse dependency order)
DROP TABLE IF EXISTS asset_maintenance_records CASCADE;
DROP TABLE IF EXISTS asset_transfers CASCADE;
DROP TABLE IF EXISTS asset_depreciation_schedules CASCADE;
DROP TABLE IF EXISTS fixed_assets CASCADE;
DROP TABLE IF EXISTS asset_locations CASCADE;
DROP TABLE IF EXISTS asset_categories CASCADE;

-- Drop types
DROP TYPE IF EXISTS asset_status CASCADE;
DROP TYPE IF EXISTS asset_condition CASCADE;
DROP TYPE IF EXISTS depreciation_method CASCADE;
DROP TYPE IF EXISTS transfer_status CASCADE;
DROP TYPE IF EXISTS maintenance_type CASCADE;
DROP TYPE IF EXISTS maintenance_status CASCADE;

-- Drop functions and triggers
DROP FUNCTION IF EXISTS update_assets_updated_at() CASCADE;

-- Drop sequences
DROP SEQUENCE IF EXISTS asset_number_seq CASCADE;
DROP SEQUENCE IF EXISTS asset_transfer_seq CASCADE;

-- ── ENUMs ──────────────────────────────────────────────────────────────────────

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

-- ── Asset Categories ───────────────────────────────────────────────────────────

CREATE TABLE asset_categories (
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

-- ── Asset Locations ────────────────────────────────────────────────────────────

CREATE TABLE asset_locations (
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

CREATE TABLE fixed_assets (
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

CREATE TABLE asset_depreciation_schedules (
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

-- ── Asset Transfers ────────────────────────────────────────────────────────────

CREATE TABLE asset_transfers (
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

-- ── Asset Maintenance Records ─────────────────────────────────────────────────

CREATE TABLE asset_maintenance_records (
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

-- ── Sequences ──────────────────────────────────────────────────────────────────

CREATE SEQUENCE IF NOT EXISTS asset_number_seq START 1000 INCREMENT 1;
CREATE SEQUENCE IF NOT EXISTS asset_transfer_seq START 100 INCREMENT 1;

-- ── Updated At Trigger Function ──────────────────────────────────────────────

CREATE OR REPLACE FUNCTION update_assets_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN 
    NEW.updated_at = NOW(); 
    RETURN NEW; 
END;
$$;

-- ── Create Triggers ────────────────────────────────────────────────────────────

CREATE TRIGGER trg_asset_cat_updated
    BEFORE UPDATE ON asset_categories
    FOR EACH ROW EXECUTE FUNCTION update_assets_updated_at();

CREATE TRIGGER trg_asset_loc_updated
    BEFORE UPDATE ON asset_locations
    FOR EACH ROW EXECUTE FUNCTION update_assets_updated_at();

CREATE TRIGGER trg_fixed_assets_updated
    BEFORE UPDATE ON fixed_assets
    FOR EACH ROW EXECUTE FUNCTION update_assets_updated_at();

CREATE TRIGGER trg_asset_transfers_updated
    BEFORE UPDATE ON asset_transfers
    FOR EACH ROW EXECUTE FUNCTION update_assets_updated_at();

CREATE TRIGGER trg_asset_maint_updated
    BEFORE UPDATE ON asset_maintenance_records
    FOR EACH ROW EXECUTE FUNCTION update_assets_updated_at();

-- ── Seed Data ──────────────────────────────────────────────────────────────────

-- Add default asset categories
INSERT INTO asset_categories (company_id, name, description, depreciation_method, useful_life_years, salvage_value_pct, gl_asset_account, gl_depreciation_account, gl_accumulated_account, color)
SELECT 
    id,
    'Buildings & Structures',
    'Buildings and structural constructions',
    'straight_line',
    30,
    5,
    '211000',
    '681100',
    '281100',
    '#ef4444'
FROM companies LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO asset_categories (company_id, name, description, depreciation_method, useful_life_years, salvage_value_pct, gl_asset_account, gl_depreciation_account, gl_accumulated_account, color)
SELECT 
    id,
    'Equipment & Machinery',
    'Industrial equipment and machinery',
    'straight_line',
    10,
    10,
    '215000',
    '681200',
    '281200',
    '#3b82f6'
FROM companies LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO asset_categories (company_id, name, description, depreciation_method, useful_life_years, salvage_value_pct, gl_asset_account, gl_depreciation_account, gl_accumulated_account, color)
SELECT 
    id,
    'Vehicles',
    'Vehicles and cars',
    'declining_balance',
    5,
    15,
    '218000',
    '681300',
    '281300',
    '#22c55e'
FROM companies LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO asset_categories (company_id, name, description, depreciation_method, useful_life_years, salvage_value_pct, gl_asset_account, gl_depreciation_account, gl_accumulated_account, color)
SELECT 
    id,
    'Office Equipment',
    'Office equipment and furniture',
    'straight_line',
    5,
    10,
    '214000',
    '681400',
    '281400',
    '#f59e0b'
FROM companies LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO asset_categories (company_id, name, description, depreciation_method, useful_life_years, salvage_value_pct, gl_asset_account, gl_depreciation_account, gl_accumulated_account, color)
SELECT 
    id,
    'Software',
    'Software and licenses',
    'straight_line',
    3,
    0,
    '208000',
    '681500',
    '281500',
    '#8b5cf6'
FROM companies LIMIT 1
ON CONFLICT DO NOTHING;

-- Add default locations
INSERT INTO asset_locations (company_id, name, description, city, country)
SELECT 
    id,
    'Head Office',
    'Company headquarters',
    'Algiers',
    'Algeria'
FROM companies LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO asset_locations (company_id, name, description, city, country)
SELECT 
    id,
    'Central Warehouse',
    'Central storage warehouse',
    'Algiers',
    'Algeria'
FROM companies LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO asset_locations (company_id, name, description, city, country)
SELECT 
    id,
    'Oran Branch',
    'Company branch in Oran',
    'Oran',
    'Algeria'
FROM companies LIMIT 1
ON CONFLICT DO NOTHING;

INSERT INTO asset_locations (company_id, name, description, city, country)
SELECT 
    id,
    'Constantine Branch',
    'Company branch in Constantine',
    'Constantine',
    'Algeria'
FROM companies LIMIT 1
ON CONFLICT DO NOTHING;

-- ── Display Results ───────────────────────────────────────────────────────────

DO $$
DECLARE
    table_count INTEGER;
BEGIN
    RAISE NOTICE '═══════════════════════════════════════════════════════════';
    RAISE NOTICE '✅ Assets Management Module Created Successfully!';
    RAISE NOTICE '═══════════════════════════════════════════════════════════';
    
    SELECT COUNT(*) INTO table_count FROM asset_categories;
    RAISE NOTICE '📊 Asset Categories: %', table_count;
    
    SELECT COUNT(*) INTO table_count FROM asset_locations;
    RAISE NOTICE '📊 Asset Locations: %', table_count;
    
    SELECT COUNT(*) INTO table_count FROM fixed_assets;
    RAISE NOTICE '📊 Fixed Assets: %', table_count;
    
    SELECT COUNT(*) INTO table_count FROM asset_depreciation_schedules;
    RAISE NOTICE '📊 Depreciation Schedules: %', table_count;
    
    SELECT COUNT(*) INTO table_count FROM asset_transfers;
    RAISE NOTICE '📊 Asset Transfers: %', table_count;
    
    -- SELECT COUNT(*) INTO table_count FROM asset_maintenance_records;
    -- RAISE NOTICE '📊 Maintenance Records: %', table_count;
    
    RAISE NOTICE '═══════════════════════════════════════════════════════════';
    RAISE NOTICE '✅ ENUMs: asset_status, asset_condition, depreciation_method,';
    RAISE NOTICE '           transfer_status, maintenance_type, maintenance_status';
    RAISE NOTICE '✅ Triggers: 5 triggers for automatic updated_at updates';
    RAISE NOTICE '✅ Sequences: asset_number_seq, asset_transfer_seq';
    RAISE NOTICE '═══════════════════════════════════════════════════════════';
END$$;