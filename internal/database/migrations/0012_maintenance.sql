-- ============================================================
-- 0012_maintenance.sql
-- Idempotent migration for Maintenance module
-- Nexus ERP v2.1.0
-- ============================================================

-- ─── 1. ENUMs ────────────────────────────────────────────────────────────────

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'equipment_status') THEN
    CREATE TYPE equipment_status AS ENUM (
      'active', 'inactive', 'under_maintenance', 'decommissioned', 'reserved'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'maintenance_priority') THEN
    CREATE TYPE maintenance_priority AS ENUM ('low', 'medium', 'high', 'critical');
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'maintenance_request_status') THEN
    CREATE TYPE maintenance_request_status AS ENUM (
      'draft', 'submitted', 'approved', 'in_progress', 'completed', 'rejected', 'cancelled'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'maintenance_order_status') THEN
    CREATE TYPE maintenance_order_status AS ENUM (
      'draft', 'planned', 'in_progress', 'on_hold', 'completed', 'cancelled'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'maintenance_order_type') THEN
    CREATE TYPE maintenance_order_type AS ENUM (
      'corrective', 'preventive', 'inspection', 'emergency', 'upgrade'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'frequency_type') THEN
    CREATE TYPE frequency_type AS ENUM (
      'daily', 'weekly', 'monthly', 'quarterly', 'semi_annual', 'annual'
    );
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'maintenance_history_type') THEN
    CREATE TYPE maintenance_history_type AS ENUM (
      'corrective', 'preventive', 'inspection', 'emergency', 'upgrade', 'other'
    );
  END IF;
END $$;

-- ─── 2. EQUIPMENT ─────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS equipment (
  id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id                UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  code                      VARCHAR(50) NOT NULL,
  name                      VARCHAR(255) NOT NULL,
  category                  VARCHAR(100),
  subcategory               VARCHAR(100),
  location                  VARCHAR(255),
  department                VARCHAR(100),
  status                    equipment_status NOT NULL DEFAULT 'active',
  purchase_date             DATE,
  purchase_cost             NUMERIC(18,2) DEFAULT 0,
  current_value             NUMERIC(18,2) DEFAULT 0,
  warranty_expiry           DATE,
  manufacturer              VARCHAR(100),
  model                     VARCHAR(100),
  serial_number             VARCHAR(100),
  asset_tag                 VARCHAR(100),
  specifications            JSONB DEFAULT '{}',
  last_maintenance_date     DATE,
  next_maintenance_date     DATE,
  maintenance_interval_days INTEGER DEFAULT 90,
  expected_life_years       INTEGER,
  notes                     TEXT,
  image_url                 TEXT,
  is_active                 BOOLEAN NOT NULL DEFAULT TRUE,
  created_by                UUID,
  created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, code)
);

CREATE INDEX IF NOT EXISTS idx_equipment_company_id      ON equipment(company_id);
CREATE INDEX IF NOT EXISTS idx_equipment_status          ON equipment(status);
CREATE INDEX IF NOT EXISTS idx_equipment_category        ON equipment(category);
CREATE INDEX IF NOT EXISTS idx_equipment_next_maint      ON equipment(next_maintenance_date);
CREATE INDEX IF NOT EXISTS idx_equipment_is_active       ON equipment(is_active);

-- ─── 3. MAINTENANCE REQUESTS ─────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_requests (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id        UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  equipment_id      UUID REFERENCES equipment(id),
  request_number    VARCHAR(50) NOT NULL,
  title             VARCHAR(255) NOT NULL,
  description       TEXT,
  priority          maintenance_priority NOT NULL DEFAULT 'medium',
  status            maintenance_request_status NOT NULL DEFAULT 'draft',
  failure_type      VARCHAR(100),
  symptoms          TEXT,
  requested_by      UUID,
  requested_by_name VARCHAR(255),
  assigned_to       UUID,
  assigned_to_name  VARCHAR(255),
  submitted_at      TIMESTAMPTZ,
  approved_at       TIMESTAMPTZ,
  approved_by       UUID,
  completed_at      TIMESTAMPTZ,
  rejection_reason  TEXT,
  estimated_cost    NUMERIC(18,2),
  actual_cost       NUMERIC(18,2),
  notes             TEXT,
  is_active         BOOLEAN NOT NULL DEFAULT TRUE,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, request_number)
);

CREATE INDEX IF NOT EXISTS idx_mreq_company_id    ON maintenance_requests(company_id);
CREATE INDEX IF NOT EXISTS idx_mreq_equipment_id  ON maintenance_requests(equipment_id);
CREATE INDEX IF NOT EXISTS idx_mreq_status        ON maintenance_requests(status);
CREATE INDEX IF NOT EXISTS idx_mreq_priority      ON maintenance_requests(priority);
CREATE INDEX IF NOT EXISTS idx_mreq_created_at    ON maintenance_requests(created_at);

-- ─── 4. MAINTENANCE ORDERS ───────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_orders (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id           UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  request_id           UUID REFERENCES maintenance_requests(id),
  equipment_id         UUID REFERENCES equipment(id),
  order_number         VARCHAR(50) NOT NULL,
  order_type           maintenance_order_type NOT NULL DEFAULT 'corrective',
  status               maintenance_order_status NOT NULL DEFAULT 'draft',
  priority             maintenance_priority NOT NULL DEFAULT 'medium',
  title                VARCHAR(255) NOT NULL,
  description          TEXT,
  work_performed       TEXT,
  findings             TEXT,
  assigned_technician  VARCHAR(255),
  team_members         JSONB DEFAULT '[]',
  planned_start_date   DATE,
  planned_end_date     DATE,
  actual_start_date    DATE,
  actual_end_date      DATE,
  estimated_hours      NUMERIC(10,2) DEFAULT 0,
  actual_hours         NUMERIC(10,2) DEFAULT 0,
  labor_cost           NUMERIC(18,2) DEFAULT 0,
  parts_cost           NUMERIC(18,2) DEFAULT 0,
  other_cost           NUMERIC(18,2) DEFAULT 0,
  total_cost           NUMERIC(18,2) DEFAULT 0,
  next_service_date    DATE,
  color                VARCHAR(20) DEFAULT '#6366f1',
  is_active            BOOLEAN NOT NULL DEFAULT TRUE,
  created_by           UUID,
  completed_by         UUID,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, order_number)
);

CREATE INDEX IF NOT EXISTS idx_morder_company_id    ON maintenance_orders(company_id);
CREATE INDEX IF NOT EXISTS idx_morder_equipment_id  ON maintenance_orders(equipment_id);
CREATE INDEX IF NOT EXISTS idx_morder_status        ON maintenance_orders(status);
CREATE INDEX IF NOT EXISTS idx_morder_order_type    ON maintenance_orders(order_type);
CREATE INDEX IF NOT EXISTS idx_morder_planned_start ON maintenance_orders(planned_start_date);
CREATE INDEX IF NOT EXISTS idx_morder_created_at    ON maintenance_orders(created_at);

-- ─── 5. MAINTENANCE ORDER LINES (PARTS USED) ─────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_order_lines (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id     UUID NOT NULL REFERENCES maintenance_orders(id) ON DELETE CASCADE,
  company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  part_name    VARCHAR(255) NOT NULL,
  part_number  VARCHAR(100),
  quantity     NUMERIC(10,3) DEFAULT 1,
  unit         VARCHAR(50),
  unit_cost    NUMERIC(18,2) DEFAULT 0,
  total_cost   NUMERIC(18,2) DEFAULT 0,
  notes        TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mol_order_id    ON maintenance_order_lines(order_id);
CREATE INDEX IF NOT EXISTS idx_mol_company_id  ON maintenance_order_lines(company_id);

-- ─── 6. PREVENTIVE MAINTENANCE PLANS ─────────────────────────────────────────

CREATE TABLE IF NOT EXISTS preventive_maintenance_plans (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id        UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  equipment_id      UUID REFERENCES equipment(id),
  name              VARCHAR(255) NOT NULL,
  description       TEXT,
  frequency_type    frequency_type NOT NULL DEFAULT 'monthly',
  frequency_value   INTEGER NOT NULL DEFAULT 1,
  estimated_hours   NUMERIC(10,2) DEFAULT 0,
  estimated_cost    NUMERIC(18,2) DEFAULT 0,
  tasks             JSONB DEFAULT '[]',
  checklist         JSONB DEFAULT '[]',
  last_performed    DATE,
  next_due          DATE,
  lead_days         INTEGER DEFAULT 7,
  auto_create_order BOOLEAN NOT NULL DEFAULT FALSE,
  assigned_to       VARCHAR(255),
  is_active         BOOLEAN NOT NULL DEFAULT TRUE,
  created_by        UUID,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_pmp_company_id    ON preventive_maintenance_plans(company_id);
CREATE INDEX IF NOT EXISTS idx_pmp_equipment_id  ON preventive_maintenance_plans(equipment_id);
CREATE INDEX IF NOT EXISTS idx_pmp_next_due      ON preventive_maintenance_plans(next_due);
CREATE INDEX IF NOT EXISTS idx_pmp_is_active     ON preventive_maintenance_plans(is_active);

-- ─── 7. MAINTENANCE SCHEDULE ─────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_schedule (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  plan_id      UUID REFERENCES preventive_maintenance_plans(id),
  order_id     UUID REFERENCES maintenance_orders(id),
  equipment_id UUID REFERENCES equipment(id),
  title        VARCHAR(255) NOT NULL,
  event_type   VARCHAR(50) NOT NULL DEFAULT 'preventive',
  scheduled_date DATE NOT NULL,
  end_date       DATE,
  status       VARCHAR(50) NOT NULL DEFAULT 'scheduled',
  color        VARCHAR(20) DEFAULT '#6366f1',
  notes        TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ms_company_id      ON maintenance_schedule(company_id);
CREATE INDEX IF NOT EXISTS idx_ms_scheduled_date  ON maintenance_schedule(scheduled_date);
CREATE INDEX IF NOT EXISTS idx_ms_equipment_id    ON maintenance_schedule(equipment_id);

-- ─── 8. MAINTENANCE HISTORY ───────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_history (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id        UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  equipment_id      UUID REFERENCES equipment(id),
  order_id          UUID REFERENCES maintenance_orders(id),
  history_type      maintenance_history_type NOT NULL DEFAULT 'corrective',
  title             VARCHAR(255) NOT NULL,
  description       TEXT,
  work_performed    TEXT,
  findings          TEXT,
  technician_name   VARCHAR(255),
  performed_date    DATE NOT NULL,
  duration_hours    NUMERIC(10,2) DEFAULT 0,
  downtime_hours    NUMERIC(10,2) DEFAULT 0,
  labor_cost        NUMERIC(18,2) DEFAULT 0,
  parts_cost        NUMERIC(18,2) DEFAULT 0,
  other_cost        NUMERIC(18,2) DEFAULT 0,
  total_cost        NUMERIC(18,2) DEFAULT 0,
  next_service_date DATE,
  is_active         BOOLEAN NOT NULL DEFAULT TRUE,
  created_by        UUID,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mh_company_id    ON maintenance_history(company_id);
CREATE INDEX IF NOT EXISTS idx_mh_equipment_id  ON maintenance_history(equipment_id);
CREATE INDEX IF NOT EXISTS idx_mh_performed_date ON maintenance_history(performed_date);
CREATE INDEX IF NOT EXISTS idx_mh_history_type  ON maintenance_history(history_type);

-- ─── 9. MAINTENANCE PARTS USED (per history record) ──────────────────────────

CREATE TABLE IF NOT EXISTS maintenance_parts_used (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  history_id   UUID NOT NULL REFERENCES maintenance_history(id) ON DELETE CASCADE,
  company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  part_name    VARCHAR(255) NOT NULL,
  part_number  VARCHAR(100),
  quantity     NUMERIC(10,3) DEFAULT 1,
  unit         VARCHAR(50),
  unit_cost    NUMERIC(18,2) DEFAULT 0,
  total_cost   NUMERIC(18,2) DEFAULT 0,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mpu_history_id   ON maintenance_parts_used(history_id);
CREATE INDEX IF NOT EXISTS idx_mpu_company_id   ON maintenance_parts_used(company_id);

-- ─── 10. AUTO-UPDATE updated_at TRIGGERS ─────────────────────────────────────

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_equipment_updated_at'
  ) THEN
    CREATE TRIGGER trg_equipment_updated_at
      BEFORE UPDATE ON equipment
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_mreq_updated_at'
  ) THEN
    CREATE TRIGGER trg_mreq_updated_at
      BEFORE UPDATE ON maintenance_requests
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_morder_updated_at'
  ) THEN
    CREATE TRIGGER trg_morder_updated_at
      BEFORE UPDATE ON maintenance_orders
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_pmp_updated_at'
  ) THEN
    CREATE TRIGGER trg_pmp_updated_at
      BEFORE UPDATE ON preventive_maintenance_plans
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_mschedule_updated_at'
  ) THEN
    CREATE TRIGGER trg_mschedule_updated_at
      BEFORE UPDATE ON maintenance_schedule
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_mhistory_updated_at'
  ) THEN
    CREATE TRIGGER trg_mhistory_updated_at
      BEFORE UPDATE ON maintenance_history
      FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
  END IF;
END $$;

-- ─── END OF 0012_maintenance.sql ─────────────────────────────────────────────
