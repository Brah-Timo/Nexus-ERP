-- =============================================================================
-- 0014_fix_schema.sql
-- Fix schema mismatches discovered at runtime
-- Idempotent — safe to run multiple times
-- =============================================================================

-- ─────────────────────────────────────────────────────────────────────────────
-- 1. BANK ACCOUNTS — add missing columns
-- ─────────────────────────────────────────────────────────────────────────────

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='bank_accounts' AND column_name='swift_code') THEN
    ALTER TABLE bank_accounts ADD COLUMN swift_code VARCHAR(20);
  END IF;
END$$;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='bank_accounts' AND column_name='swift')
  AND EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='bank_accounts' AND column_name='swift_code') THEN
    UPDATE bank_accounts SET swift_code = swift WHERE swift_code IS NULL AND swift IS NOT NULL;
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='bank_accounts' AND column_name='branch') THEN
    ALTER TABLE bank_accounts ADD COLUMN branch VARCHAR(200);
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='bank_accounts' AND column_name='opening_balance') THEN
    ALTER TABLE bank_accounts ADD COLUMN opening_balance NUMERIC(18,2) NOT NULL DEFAULT 0;
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='bank_accounts' AND column_name='last_reconciled') THEN
    ALTER TABLE bank_accounts ADD COLUMN last_reconciled DATE;
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='bank_accounts' AND column_name='updated_at') THEN
    ALTER TABLE bank_accounts ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
  END IF;
END$$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='bank_accounts' AND column_name='notes') THEN
    ALTER TABLE bank_accounts ADD COLUMN notes TEXT;
  END IF;
END$$;

-- ─────────────────────────────────────────────────────────────────────────────
-- 2. TAX_PAYMENTS — ensure balance is a GENERATED column
-- ─────────────────────────────────────────────────────────────────────────────

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name='tax_payments') THEN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
      WHERE table_name='tax_payments' AND column_name='balance') THEN
      ALTER TABLE tax_payments
        ADD COLUMN balance NUMERIC(18,2) GENERATED ALWAYS AS (amount_due - amount_paid) STORED;
    END IF;
  END IF;
END$$;

-- ─────────────────────────────────────────────────────────────────────────────
-- 3. STOCK_LEVELS — ensure table exists
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS stock_levels (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id    UUID          NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    item_id       UUID          NOT NULL,
    warehouse_id  UUID,
    location_id   UUID,
    qty_on_hand   NUMERIC(18,4) NOT NULL DEFAULT 0,
    qty_reserved  NUMERIC(18,4) NOT NULL DEFAULT 0,
    qty_available NUMERIC(18,4) GENERATED ALWAYS AS (qty_on_hand - qty_reserved) STORED,
    unit_cost     NUMERIC(18,4) NOT NULL DEFAULT 0,
    reorder_point NUMERIC(18,4) NOT NULL DEFAULT 0,
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_stock_levels_unique 
ON stock_levels (company_id, item_id, COALESCE(warehouse_id, '00000000-0000-0000-0000-000000000000'));

CREATE INDEX IF NOT EXISTS idx_stock_levels_company ON stock_levels(company_id);
CREATE INDEX IF NOT EXISTS idx_stock_levels_item ON stock_levels(item_id);
CREATE INDEX IF NOT EXISTS idx_stock_levels_warehouse ON stock_levels(warehouse_id);

-- ─────────────────────────────────────────────────────────────────────────────
-- 4. MO_COMPONENT_LINES — ensure it exists for MRP planning
-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS mo_component_lines (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    mo_id          UUID          NOT NULL,
    component_id   UUID          NOT NULL,
    required_qty   NUMERIC(18,4) NOT NULL DEFAULT 0,
    consumed_qty   NUMERIC(18,4) NOT NULL DEFAULT 0,
    unit           VARCHAR(20)   NOT NULL DEFAULT 'unit',
    notes          TEXT,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mo_comp_mo ON mo_component_lines(mo_id);
CREATE INDEX IF NOT EXISTS idx_mo_comp_item ON mo_component_lines(component_id);

-- ─────────────────────────────────────────────────────────────────────────────
-- 5. FLEET — ensure fleet_vehicles & related tables exist
-- ─────────────────────────────────────────────────────────────────────────────

DO $$
BEGIN
  CREATE TYPE vehicle_status AS ENUM ('active','in_use','maintenance','inactive','retired');
EXCEPTION WHEN duplicate_object THEN NULL;
END$$;

DO $$
BEGIN
  CREATE TYPE vehicle_type AS ENUM ('car','truck','van','bus','motorcycle','other');
EXCEPTION WHEN duplicate_object THEN NULL;
END$$;

DO $$
BEGIN
  CREATE TYPE fuel_type AS ENUM ('gasoline','diesel','lpg','electric','hybrid');
EXCEPTION WHEN duplicate_object THEN NULL;
END$$;

DO $$
BEGIN
  CREATE TYPE driver_status AS ENUM ('available','on_duty','off_duty','on_leave','inactive');
EXCEPTION WHEN duplicate_object THEN NULL;
END$$;

DO $$
BEGIN
  CREATE TYPE license_class AS ENUM ('B','C','D','E','A');
EXCEPTION WHEN duplicate_object THEN NULL;
END$$;

DO $$
BEGIN
  CREATE TYPE assignment_status AS ENUM ('active','completed','cancelled');
EXCEPTION WHEN duplicate_object THEN NULL;
END$$;

DO $$
BEGIN
  CREATE TYPE fleet_expense_type AS ENUM ('fuel','maintenance','insurance','registration','repair','toll','parking','other');
EXCEPTION WHEN duplicate_object THEN NULL;
END$$;

DO $$
BEGIN
  CREATE TYPE fleet_maintenance_status AS ENUM ('scheduled','in_progress','completed','cancelled','overdue');
EXCEPTION WHEN duplicate_object THEN NULL;
END$$;

CREATE TABLE IF NOT EXISTS fleet_vehicles (
    id                 UUID               PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id         UUID               NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code               VARCHAR(30)        UNIQUE,
    plate_number       VARCHAR(30)        NOT NULL,
    make               VARCHAR(100),
    model              VARCHAR(100),
    year               INTEGER,
    vehicle_type       vehicle_type       NOT NULL DEFAULT 'car',
    status             vehicle_status     NOT NULL DEFAULT 'active',
    fuel_type          fuel_type          NOT NULL DEFAULT 'diesel',
    color              VARCHAR(50),
    vin                VARCHAR(50),
    engine_number      VARCHAR(50),
    chassis_number     VARCHAR(50),
    seats              INTEGER,
    max_load_kg        NUMERIC(10,2),
    current_mileage    NUMERIC(12,2)      NOT NULL DEFAULT 0,
    insurance_policy   VARCHAR(100),
    insurance_expiry   DATE,
    inspection_expiry  DATE,
    registration_expiry DATE,
    purchase_date      DATE,
    purchase_price     NUMERIC(18,2)      NOT NULL DEFAULT 0,
    assigned_driver_id UUID,
    notes              TEXT,
    created_at         TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fleet_drivers (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID            NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code            VARCHAR(30)     UNIQUE,
    first_name      VARCHAR(100)    NOT NULL,
    last_name       VARCHAR(100)    NOT NULL,
    phone           VARCHAR(50),
    email           VARCHAR(200),
    license_number  VARCHAR(50),
    license_class   license_class   NOT NULL DEFAULT 'B',
    license_expiry  DATE,
    medical_expiry  DATE,
    hire_date       DATE,
    status          driver_status   NOT NULL DEFAULT 'available',
    employee_id     UUID,
    notes           TEXT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fleet_assignments (
    id               UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID              NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    assignment_number VARCHAR(30)      UNIQUE,
    vehicle_id       UUID              NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
    driver_id        UUID              NOT NULL REFERENCES fleet_drivers(id) ON DELETE CASCADE,
    start_date       DATE              NOT NULL,
    end_date         DATE,
    purpose          VARCHAR(200),
    destination      VARCHAR(200),
    mileage_start    NUMERIC(12,2)     NOT NULL DEFAULT 0,
    mileage_end      NUMERIC(12,2),
    status           assignment_status NOT NULL DEFAULT 'active',
    notes            TEXT,
    created_at       TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ       NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fleet_fuel_logs (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    vehicle_id       UUID        NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
    driver_id        UUID        REFERENCES fleet_drivers(id) ON DELETE SET NULL,
    fill_date        DATE        NOT NULL,
    fuel_type        fuel_type   NOT NULL DEFAULT 'diesel',
    liters           NUMERIC(10,3) NOT NULL,
    price_per_liter  NUMERIC(10,4) NOT NULL,
    total_cost       NUMERIC(18,2) GENERATED ALWAYS AS (liters * price_per_liter) STORED,
    mileage_at_fill  NUMERIC(12,2),
    is_full_tank     BOOLEAN     NOT NULL DEFAULT TRUE,
    fuel_station     VARCHAR(200),
    receipt_number   VARCHAR(100),
    notes            TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fleet_maintenance (
    id               UUID                    PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID                    NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    vehicle_id       UUID                    NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
    service_type     VARCHAR(100)            NOT NULL,
    status           fleet_maintenance_status NOT NULL DEFAULT 'scheduled',
    scheduled_date   DATE,
    completed_date   DATE,
    mileage_at_service NUMERIC(12,2),
    description      TEXT,
    garage_name      VARCHAR(200),
    labor_cost       NUMERIC(18,2)           NOT NULL DEFAULT 0,
    parts_cost       NUMERIC(18,2)           NOT NULL DEFAULT 0,
    total_cost       NUMERIC(18,2)           GENERATED ALWAYS AS (labor_cost + parts_cost) STORED,
    invoice_number   VARCHAR(100),
    notes            TEXT,
    created_at       TIMESTAMPTZ             NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ             NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fleet_expenses (
    id            UUID               PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id    UUID               NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    vehicle_id    UUID               NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
    driver_id     UUID               REFERENCES fleet_drivers(id) ON DELETE SET NULL,
    expense_type  fleet_expense_type NOT NULL DEFAULT 'other',
    expense_date  DATE               NOT NULL,
    amount        NUMERIC(18,2)      NOT NULL DEFAULT 0,
    description   TEXT,
    reference     VARCHAR(100),
    notes         TEXT,
    created_at    TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fv_company ON fleet_vehicles(company_id);
CREATE INDEX IF NOT EXISTS idx_fv_status ON fleet_vehicles(status);
CREATE INDEX IF NOT EXISTS idx_fd_company ON fleet_drivers(company_id);
CREATE INDEX IF NOT EXISTS idx_fa_company ON fleet_assignments(company_id);
CREATE INDEX IF NOT EXISTS idx_fa_vehicle ON fleet_assignments(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fa_driver ON fleet_assignments(driver_id);
CREATE INDEX IF NOT EXISTS idx_ffl_company ON fleet_fuel_logs(company_id);
CREATE INDEX IF NOT EXISTS idx_ffl_vehicle ON fleet_fuel_logs(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fm_company ON fleet_maintenance(company_id);
CREATE INDEX IF NOT EXISTS idx_fm_vehicle ON fleet_maintenance(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fe_company ON fleet_expenses(company_id);
CREATE INDEX IF NOT EXISTS idx_fe_vehicle ON fleet_expenses(vehicle_id);

-- ─────────────────────────────────────────────────────────────────────────────
-- 6. ITEMS TABLE — ensure it has required columns
-- ─────────────────────────────────────────────────────────────────────────────

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
    WHERE table_name='items' AND column_name='code') THEN
    ALTER TABLE items ADD COLUMN code VARCHAR(50);
  END IF;
END$$;