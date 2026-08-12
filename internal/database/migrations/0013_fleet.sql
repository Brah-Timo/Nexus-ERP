-- =============================================================================
-- Migration 0013: Fleet Management Module
-- Idempotent: safe to run multiple times
-- =============================================================================

-- ── ENUMs ────────────────────────────────────────────────────────────────────

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'vehicle_status') THEN
    CREATE TYPE vehicle_status AS ENUM (
      'active', 'inactive', 'in_service', 'under_maintenance',
      'out_of_service', 'sold', 'retired'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'vehicle_type') THEN
    CREATE TYPE vehicle_type AS ENUM (
      'car', 'truck', 'van', 'bus', 'motorcycle',
      'heavy_equipment', 'trailer', 'other'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'fuel_type') THEN
    CREATE TYPE fuel_type AS ENUM (
      'petrol', 'diesel', 'electric', 'hybrid', 'lpg', 'cng', 'other'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'driver_status') THEN
    CREATE TYPE driver_status AS ENUM (
      'active', 'inactive', 'on_leave', 'suspended', 'terminated'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'license_class') THEN
    CREATE TYPE license_class AS ENUM (
      'A', 'B', 'C', 'D', 'E', 'B1', 'C1', 'D1', 'other'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'assignment_status') THEN
    CREATE TYPE assignment_status AS ENUM (
      'active', 'completed', 'cancelled'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'fleet_expense_type') THEN
    CREATE TYPE fleet_expense_type AS ENUM (
      'insurance', 'registration', 'tax', 'fine', 'parking',
      'toll', 'repair', 'parts', 'cleaning', 'other'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'fleet_maintenance_status') THEN
    CREATE TYPE fleet_maintenance_status AS ENUM (
      'scheduled', 'in_progress', 'completed', 'overdue', 'cancelled'
    );
  END IF;
END $$;

-- ── TABLES ───────────────────────────────────────────────────────────────────

-- vehicles
CREATE TABLE IF NOT EXISTS fleet_vehicles (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id            UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  plate_number          VARCHAR(30) NOT NULL,
  vin                   VARCHAR(50),
  make                  VARCHAR(100) NOT NULL,
  model                 VARCHAR(100) NOT NULL,
  year                  INTEGER NOT NULL,
  color                 VARCHAR(50),
  vehicle_type          vehicle_type NOT NULL DEFAULT 'car',
  fuel_type             fuel_type NOT NULL DEFAULT 'petrol',
  status                vehicle_status NOT NULL DEFAULT 'active',
  odometer_km           INTEGER NOT NULL DEFAULT 0,
  fuel_tank_capacity    NUMERIC(10,2),
  seating_capacity      INTEGER,
  purchase_date         DATE,
  purchase_price        NUMERIC(15,2) DEFAULT 0,
  current_value         NUMERIC(15,2) DEFAULT 0,
  insurance_policy      VARCHAR(100),
  insurance_expiry      DATE,
  registration_expiry   DATE,
  technical_visit_expiry DATE,
  assigned_driver_id    UUID,
  department            VARCHAR(100),
  notes                 TEXT,
  image_url             TEXT,
  is_active             BOOLEAN NOT NULL DEFAULT TRUE,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, plate_number)
);

-- drivers
CREATE TABLE IF NOT EXISTS fleet_drivers (
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id          UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  employee_id         UUID REFERENCES employees(id) ON DELETE SET NULL,
  first_name          VARCHAR(100) NOT NULL,
  last_name           VARCHAR(100) NOT NULL,
  phone               VARCHAR(30),
  email               VARCHAR(200),
  national_id         VARCHAR(50),
  license_number      VARCHAR(50) NOT NULL,
  license_class       license_class NOT NULL DEFAULT 'B',
  license_expiry      DATE,
  license_issue_date  DATE,
  status              driver_status NOT NULL DEFAULT 'active',
  hire_date           DATE,
  address             TEXT,
  emergency_contact   VARCHAR(200),
  emergency_phone     VARCHAR(30),
  notes               TEXT,
  photo_url           TEXT,
  is_active           BOOLEAN NOT NULL DEFAULT TRUE,
  created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, license_number)
);

-- Add FK for assigned_driver after drivers table exists
DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.table_constraints
    WHERE constraint_name = 'fk_vehicle_assigned_driver'
  ) THEN
    ALTER TABLE fleet_vehicles
      ADD CONSTRAINT fk_vehicle_assigned_driver
      FOREIGN KEY (assigned_driver_id) REFERENCES fleet_drivers(id) ON DELETE SET NULL;
  END IF;
END $$;

-- vehicle assignments
CREATE TABLE IF NOT EXISTS fleet_assignments (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  vehicle_id      UUID NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
  driver_id       UUID NOT NULL REFERENCES fleet_drivers(id) ON DELETE CASCADE,
  start_date      DATE NOT NULL,
  end_date        DATE,
  start_odometer  INTEGER DEFAULT 0,
  end_odometer    INTEGER,
  purpose         VARCHAR(200),
  destination     VARCHAR(200),
  notes           TEXT,
  status          assignment_status NOT NULL DEFAULT 'active',
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- fuel logs
CREATE TABLE IF NOT EXISTS fleet_fuel_logs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  vehicle_id      UUID NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
  driver_id       UUID REFERENCES fleet_drivers(id) ON DELETE SET NULL,
  log_date        DATE NOT NULL DEFAULT CURRENT_DATE,
  odometer_km     INTEGER NOT NULL DEFAULT 0,
  liters          NUMERIC(10,3) NOT NULL DEFAULT 0,
  price_per_liter NUMERIC(10,3) NOT NULL DEFAULT 0,
  total_cost      NUMERIC(15,2) GENERATED ALWAYS AS (liters * price_per_liter) STORED,
  fuel_type       fuel_type NOT NULL DEFAULT 'diesel',
  station_name    VARCHAR(200),
  full_tank       BOOLEAN NOT NULL DEFAULT TRUE,
  notes           TEXT,
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- fleet maintenance
CREATE TABLE IF NOT EXISTS fleet_maintenance (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id        UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  vehicle_id        UUID NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
  title             VARCHAR(200) NOT NULL,
  description       TEXT,
  maintenance_type  VARCHAR(50) NOT NULL DEFAULT 'routine',
  status            fleet_maintenance_status NOT NULL DEFAULT 'scheduled',
  scheduled_date    DATE,
  completed_date    DATE,
  odometer_km       INTEGER DEFAULT 0,
  next_service_km   INTEGER,
  next_service_date DATE,
  technician        VARCHAR(200),
  garage_name       VARCHAR(200),
  labor_cost        NUMERIC(15,2) DEFAULT 0,
  parts_cost        NUMERIC(15,2) DEFAULT 0,
  total_cost        NUMERIC(15,2) DEFAULT 0,
  work_performed    TEXT,
  notes             TEXT,
  is_active         BOOLEAN NOT NULL DEFAULT TRUE,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- vehicle expenses
CREATE TABLE IF NOT EXISTS fleet_expenses (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  vehicle_id      UUID NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
  driver_id       UUID REFERENCES fleet_drivers(id) ON DELETE SET NULL,
  expense_type    fleet_expense_type NOT NULL DEFAULT 'other',
  expense_date    DATE NOT NULL DEFAULT CURRENT_DATE,
  amount          NUMERIC(15,2) NOT NULL DEFAULT 0,
  description     VARCHAR(500),
  reference_number VARCHAR(100),
  notes           TEXT,
  is_active       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── INDEXES ──────────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_fleet_vehicles_company    ON fleet_vehicles(company_id);
CREATE INDEX IF NOT EXISTS idx_fleet_vehicles_status     ON fleet_vehicles(status);
CREATE INDEX IF NOT EXISTS idx_fleet_vehicles_plate      ON fleet_vehicles(company_id, plate_number);
CREATE INDEX IF NOT EXISTS idx_fleet_drivers_company     ON fleet_drivers(company_id);
CREATE INDEX IF NOT EXISTS idx_fleet_drivers_status      ON fleet_drivers(status);
CREATE INDEX IF NOT EXISTS idx_fleet_assignments_vehicle ON fleet_assignments(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fleet_assignments_driver  ON fleet_assignments(driver_id);
CREATE INDEX IF NOT EXISTS idx_fleet_assignments_status  ON fleet_assignments(status);
CREATE INDEX IF NOT EXISTS idx_fleet_fuel_vehicle        ON fleet_fuel_logs(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fleet_fuel_date           ON fleet_fuel_logs(log_date);
CREATE INDEX IF NOT EXISTS idx_fleet_maintenance_vehicle ON fleet_maintenance(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fleet_maintenance_status  ON fleet_maintenance(status);
CREATE INDEX IF NOT EXISTS idx_fleet_expenses_vehicle    ON fleet_expenses(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fleet_expenses_type       ON fleet_expenses(expense_type);
CREATE INDEX IF NOT EXISTS idx_fleet_expenses_date       ON fleet_expenses(expense_date);

-- ── TRIGGERS ─────────────────────────────────────────────────────────────────

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_fleet_vehicles_updated_at') THEN
    CREATE TRIGGER trg_fleet_vehicles_updated_at
      BEFORE UPDATE ON fleet_vehicles
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_fleet_drivers_updated_at') THEN
    CREATE TRIGGER trg_fleet_drivers_updated_at
      BEFORE UPDATE ON fleet_drivers
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_fleet_assignments_updated_at') THEN
    CREATE TRIGGER trg_fleet_assignments_updated_at
      BEFORE UPDATE ON fleet_assignments
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_fleet_fuel_logs_updated_at') THEN
    CREATE TRIGGER trg_fleet_fuel_logs_updated_at
      BEFORE UPDATE ON fleet_fuel_logs
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_fleet_maintenance_updated_at') THEN
    CREATE TRIGGER trg_fleet_maintenance_updated_at
      BEFORE UPDATE ON fleet_maintenance
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_fleet_expenses_updated_at') THEN
    CREATE TRIGGER trg_fleet_expenses_updated_at
      BEFORE UPDATE ON fleet_expenses
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;
