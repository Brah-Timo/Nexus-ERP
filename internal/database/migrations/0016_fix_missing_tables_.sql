-- =============================================================================
-- Nexus ERP — Fix Missing Tables Migration
-- Version: 0016 | Fleet + Maintenance + Quality tables (idempotent)
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- SECTION 1 : FLEET MANAGEMENT
-- =============================================================================

-- ── ENUMs ────────────────────────────────────────────────────────────────────
DO $$ BEGIN
    CREATE TYPE vehicle_status AS ENUM ('active','inactive','maintenance','sold','scrapped');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
    CREATE TYPE driver_status AS ENUM ('active','inactive','suspended','terminated');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
    CREATE TYPE assignment_status AS ENUM ('active','completed','cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
    CREATE TYPE fleet_fuel_type AS ENUM ('diesel','gasoline','electric','hybrid','lpg','cng');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
    CREATE TYPE fleet_maintenance_status AS ENUM ('scheduled','in_progress','completed','cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
    CREATE TYPE fleet_expense_type AS ENUM (
        'fuel','maintenance','insurance','registration','fine','toll','parking',
        'washing','tire','other'
    );
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

-- ── fleet_drivers ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS fleet_drivers (
    id                  UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id          UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    first_name          VARCHAR(100) NOT NULL,
    last_name           VARCHAR(100) NOT NULL,
    phone               VARCHAR(30),
    email               VARCHAR(150),
    national_id         VARCHAR(50),
    license_number      VARCHAR(50),
    license_class       VARCHAR(20)  NOT NULL DEFAULT 'B',
    license_expiry      DATE,
    license_issue_date  DATE,
    status              VARCHAR(30)  NOT NULL DEFAULT 'active',
    hire_date           DATE,
    address             TEXT,
    emergency_contact   VARCHAR(150),
    emergency_phone     VARCHAR(30),
    notes               TEXT,
    is_active           BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_fleet_drivers_company ON fleet_drivers(company_id);
CREATE INDEX IF NOT EXISTS idx_fleet_drivers_status  ON fleet_drivers(company_id, status);

-- ── fleet_vehicles ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS fleet_vehicles (
    id                      UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id              UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    plate_number            VARCHAR(30)  NOT NULL,
    vin                     VARCHAR(50),
    make                    VARCHAR(100) NOT NULL DEFAULT 'Unknown',
    model                   VARCHAR(100) NOT NULL DEFAULT 'Unknown',
    year                    INT          NOT NULL DEFAULT EXTRACT(YEAR FROM NOW())::INT,
    color                   VARCHAR(50),
    vehicle_type            VARCHAR(30)  NOT NULL DEFAULT 'car'
                                CHECK (vehicle_type IN ('car','truck','van','bus','motorcycle','heavy_equipment','other')),
    fuel_type               VARCHAR(30)  NOT NULL DEFAULT 'diesel',
    status                  VARCHAR(30)  NOT NULL DEFAULT 'active',
    mileage_at_fill         INT          NOT NULL DEFAULT 0,
    current_mileage         INT          NOT NULL DEFAULT 0,
    fuel_tank_capacity      NUMERIC(8,2),
    seating_capacity        INT,
    purchase_date           DATE,
    purchase_price          NUMERIC(18,2),
    current_value           NUMERIC(18,2),
    insurance_policy        VARCHAR(100),
    insurance_expiry        DATE,
    registration_expiry     DATE,
    technical_visit_expiry  DATE,
    department              VARCHAR(100),
    notes                   TEXT,
    image_url               TEXT,
    assigned_driver_id      UUID         REFERENCES fleet_drivers(id),
    is_active               BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, plate_number)
);
CREATE INDEX IF NOT EXISTS idx_fleet_vehicles_company ON fleet_vehicles(company_id);
CREATE INDEX IF NOT EXISTS idx_fleet_vehicles_status  ON fleet_vehicles(company_id, status);
CREATE INDEX IF NOT EXISTS idx_fleet_vehicles_driver  ON fleet_vehicles(assigned_driver_id);

-- ── fleet_assignments ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS fleet_assignments (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    vehicle_id      UUID         NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
    driver_id       UUID         NOT NULL REFERENCES fleet_drivers(id)  ON DELETE CASCADE,
    start_date      DATE         NOT NULL DEFAULT CURRENT_DATE,
    end_date        DATE,
    start_odometer  INT          NOT NULL DEFAULT 0,
    end_odometer    INT,
    purpose         VARCHAR(200),
    destination     VARCHAR(200),
    notes           TEXT,
    status          VARCHAR(30)  NOT NULL DEFAULT 'active',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_fleet_assign_company ON fleet_assignments(company_id);
CREATE INDEX IF NOT EXISTS idx_fleet_assign_vehicle ON fleet_assignments(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fleet_assign_driver  ON fleet_assignments(driver_id);

-- ── fleet_fuel_logs ───────────────────────────────────────────────────────────
-- 🔧 FIX: Add missing columns if they don't exist
DO $$ 
BEGIN
    -- Check if table exists before altering
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'fleet_fuel_logs') THEN
        ALTER TABLE fleet_fuel_logs ADD COLUMN IF NOT EXISTS fill_date DATE NOT NULL DEFAULT CURRENT_DATE;
        ALTER TABLE fleet_fuel_logs ADD COLUMN IF NOT EXISTS mileage_at_fill INT NOT NULL DEFAULT 0;
        ALTER TABLE fleet_fuel_logs ADD COLUMN IF NOT EXISTS liters NUMERIC(10,3) NOT NULL DEFAULT 0;
        ALTER TABLE fleet_fuel_logs ADD COLUMN IF NOT EXISTS price_per_liter NUMERIC(10,4) NOT NULL DEFAULT 0;
        ALTER TABLE fleet_fuel_logs ADD COLUMN IF NOT EXISTS fuel_type VARCHAR(30) NOT NULL DEFAULT 'diesel';
        ALTER TABLE fleet_fuel_logs ADD COLUMN IF NOT EXISTS fuel_station VARCHAR(200);
        ALTER TABLE fleet_fuel_logs ADD COLUMN IF NOT EXISTS is_full_tank BOOLEAN NOT NULL DEFAULT TRUE;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS fleet_fuel_logs (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    vehicle_id      UUID         NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
    driver_id       UUID         REFERENCES fleet_drivers(id),
    fill_date       DATE         NOT NULL DEFAULT CURRENT_DATE,
    mileage_at_fill INT          NOT NULL DEFAULT 0,
    liters          NUMERIC(10,3) NOT NULL DEFAULT 0,
    price_per_liter NUMERIC(10,4) NOT NULL DEFAULT 0,
    total_cost      NUMERIC(18,2) GENERATED ALWAYS AS (liters * price_per_liter) STORED,
    fuel_type       VARCHAR(30)  NOT NULL DEFAULT 'diesel',
    fuel_station    VARCHAR(200),
    is_full_tank    BOOLEAN      NOT NULL DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_fleet_fuel_company ON fleet_fuel_logs(company_id);
CREATE INDEX IF NOT EXISTS idx_fleet_fuel_vehicle ON fleet_fuel_logs(vehicle_id, fill_date DESC);

-- ── fleet_maintenance ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS fleet_maintenance (
    id                  UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id          UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    vehicle_id          UUID         NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
    title               VARCHAR(200) NOT NULL,
    description         TEXT,
    maintenance_type    VARCHAR(50)  NOT NULL DEFAULT 'preventive'
                            CHECK (maintenance_type IN ('preventive','corrective','emergency','inspection','tire','oil_change','other')),
    status              VARCHAR(30)  NOT NULL DEFAULT 'scheduled',
    scheduled_date      DATE,
    completed_date      DATE,
    mileage_at_fill     INT          NOT NULL DEFAULT 0,
    next_service_km     INT,
    next_service_date   DATE,
    technician          VARCHAR(150),
    garage_name         VARCHAR(200),
    labor_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    parts_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    work_performed      TEXT,
    notes               TEXT,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_fleet_maint_company ON fleet_maintenance(company_id);
CREATE INDEX IF NOT EXISTS idx_fleet_maint_vehicle ON fleet_maintenance(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_fleet_maint_status  ON fleet_maintenance(company_id, status);

-- ── fleet_expenses ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS fleet_expenses (
    id               UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id       UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    vehicle_id       UUID         NOT NULL REFERENCES fleet_vehicles(id) ON DELETE CASCADE,
    driver_id        UUID         REFERENCES fleet_drivers(id),
    expense_type     VARCHAR(50)  NOT NULL DEFAULT 'other',
    expense_date     DATE         NOT NULL DEFAULT CURRENT_DATE,
    amount           NUMERIC(18,2) NOT NULL DEFAULT 0,
    description      TEXT,
    reference_number VARCHAR(100),
    notes            TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_fleet_exp_company ON fleet_expenses(company_id);
CREATE INDEX IF NOT EXISTS idx_fleet_exp_vehicle ON fleet_expenses(vehicle_id, expense_date DESC);

-- =============================================================================
-- SECTION 2 : MAINTENANCE MANAGEMENT
-- =============================================================================

-- ── equipment ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS equipment (
    id                          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id                  UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code                        VARCHAR(50)  NOT NULL,
    name                        VARCHAR(200) NOT NULL,
    category                    VARCHAR(100),
    subcategory                 VARCHAR(100),
    location                    VARCHAR(200),
    department                  VARCHAR(100),
    status                      VARCHAR(30)  NOT NULL DEFAULT 'active'
                                    CHECK (status IN ('active','inactive','under_maintenance','retired','disposed')),
    purchase_date               DATE,
    purchase_cost               NUMERIC(18,2) NOT NULL DEFAULT 0,
    current_value               NUMERIC(18,2) NOT NULL DEFAULT 0,
    warranty_expiry             DATE,
    manufacturer                VARCHAR(150),
    model                       VARCHAR(150),
    serial_number               VARCHAR(100),
    asset_tag                   VARCHAR(100),
    last_maintenance_date       DATE,
    next_maintenance_date       DATE,
    maintenance_interval_days   INT          NOT NULL DEFAULT 90,
    expected_life_years         INT,
    notes                       TEXT,
    is_active                   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at                  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);
CREATE INDEX IF NOT EXISTS idx_equipment_company   ON equipment(company_id);
CREATE INDEX IF NOT EXISTS idx_equipment_status    ON equipment(company_id, status);
CREATE INDEX IF NOT EXISTS idx_equipment_next_maint ON equipment(company_id, next_maintenance_date);

-- ── maintenance_requests ──────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS maintenance_requests (
    id                  UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id          UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    equipment_id        UUID         REFERENCES equipment(id),
    request_number      VARCHAR(50)  NOT NULL,
    title               VARCHAR(200) NOT NULL,
    description         TEXT,
    priority            VARCHAR(20)  NOT NULL DEFAULT 'medium'
                            CHECK (priority IN ('low','medium','high','critical')),
    status              VARCHAR(30)  NOT NULL DEFAULT 'draft'
                            CHECK (status IN ('draft','open','assigned','in_progress','completed','rejected','cancelled')),
    failure_type        VARCHAR(50),
    symptoms            TEXT,
    requested_by_name   VARCHAR(150),
    assigned_to_name    VARCHAR(150),
    estimated_cost      NUMERIC(18,2),
    actual_cost         NUMERIC(18,2),
    resolution_notes    TEXT,
    completed_at        TIMESTAMPTZ,
    notes               TEXT,
    is_active           BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, request_number)
);
CREATE INDEX IF NOT EXISTS idx_mr_company   ON maintenance_requests(company_id);
CREATE INDEX IF NOT EXISTS idx_mr_equipment ON maintenance_requests(equipment_id);
CREATE INDEX IF NOT EXISTS idx_mr_status    ON maintenance_requests(company_id, status);

-- ── maintenance_orders ────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS maintenance_orders (
    id                  UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id          UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    request_id          UUID         REFERENCES maintenance_requests(id),
    equipment_id        UUID         REFERENCES equipment(id),
    order_number        VARCHAR(50)  NOT NULL,
    order_type          VARCHAR(30)  NOT NULL DEFAULT 'corrective'
                            CHECK (order_type IN ('corrective','preventive','inspection','emergency','improvement')),
    status              VARCHAR(30)  NOT NULL DEFAULT 'draft'
                            CHECK (status IN ('draft','open','in_progress','on_hold','completed','cancelled')),
    priority            VARCHAR(20)  NOT NULL DEFAULT 'medium'
                            CHECK (priority IN ('low','medium','high','critical')),
    title               VARCHAR(200) NOT NULL,
    description         TEXT,
    assigned_technician VARCHAR(150),
    planned_start_date  DATE,
    planned_end_date    DATE,
    actual_start_date   DATE,
    actual_end_date     DATE,
    estimated_hours     NUMERIC(8,2) NOT NULL DEFAULT 0,
    actual_hours        NUMERIC(8,2) NOT NULL DEFAULT 0,
    labor_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    parts_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    other_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    color               VARCHAR(20),
    completion_notes    TEXT,
    is_active           BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, order_number)
);
CREATE INDEX IF NOT EXISTS idx_mo_company   ON maintenance_orders(company_id);
CREATE INDEX IF NOT EXISTS idx_mo_equipment ON maintenance_orders(equipment_id);
CREATE INDEX IF NOT EXISTS idx_mo_status    ON maintenance_orders(company_id, status);
CREATE INDEX IF NOT EXISTS idx_mo_request   ON maintenance_orders(request_id);

-- ── maintenance_order_lines ───────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS maintenance_order_lines (
    id           UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id     UUID         NOT NULL REFERENCES maintenance_orders(id) ON DELETE CASCADE,
    part_name    VARCHAR(200) NOT NULL,
    part_number  VARCHAR(100),
    quantity     NUMERIC(14,4) NOT NULL DEFAULT 1,
    unit         VARCHAR(30),
    unit_cost    NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost   NUMERIC(18,2) NOT NULL DEFAULT 0,
    notes        TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mol_order ON maintenance_order_lines(order_id);

-- ── preventive_maintenance_plans ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS preventive_maintenance_plans (
    id                  UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id          UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    equipment_id        UUID         REFERENCES equipment(id),
    name                VARCHAR(200) NOT NULL,
    description         TEXT,
    frequency_type      VARCHAR(20)  NOT NULL DEFAULT 'monthly'
                            CHECK (frequency_type IN ('daily','weekly','monthly','quarterly','yearly','hours','km')),
    frequency_value     INT          NOT NULL DEFAULT 1,
    estimated_hours     NUMERIC(8,2) NOT NULL DEFAULT 0,
    estimated_cost      NUMERIC(18,2) NOT NULL DEFAULT 0,
    last_performed      DATE,
    next_due            DATE,
    lead_days           INT          NOT NULL DEFAULT 7,
    auto_create_order   BOOLEAN      NOT NULL DEFAULT FALSE,
    assigned_to         VARCHAR(150),
    is_active           BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_pmp_company   ON preventive_maintenance_plans(company_id);
CREATE INDEX IF NOT EXISTS idx_pmp_equipment ON preventive_maintenance_plans(equipment_id);
CREATE INDEX IF NOT EXISTS idx_pmp_next_due  ON preventive_maintenance_plans(company_id, next_due);

-- ── maintenance_history ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS maintenance_history (
    id               UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id       UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    equipment_id     UUID         REFERENCES equipment(id),
    order_id         UUID         REFERENCES maintenance_orders(id),
    history_type     VARCHAR(30)  NOT NULL DEFAULT 'corrective',
    title            VARCHAR(200) NOT NULL,
    description      TEXT,
    work_performed   TEXT,
    findings         TEXT,
    technician_name  VARCHAR(150),
    performed_date   DATE         NOT NULL DEFAULT CURRENT_DATE,
    duration_hours   NUMERIC(8,2) NOT NULL DEFAULT 0,
    downtime_hours   NUMERIC(8,2) NOT NULL DEFAULT 0,
    labor_cost       NUMERIC(18,2) NOT NULL DEFAULT 0,
    parts_cost       NUMERIC(18,2) NOT NULL DEFAULT 0,
    other_cost       NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_cost       NUMERIC(18,2) NOT NULL DEFAULT 0,
    next_service_date DATE,
    is_active        BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_mh_company   ON maintenance_history(company_id);
CREATE INDEX IF NOT EXISTS idx_mh_equipment ON maintenance_history(equipment_id, performed_date DESC);
CREATE INDEX IF NOT EXISTS idx_mh_order     ON maintenance_history(order_id);

-- =============================================================================
-- SECTION 3 : QUALITY MANAGEMENT
-- =============================================================================

-- ── ENUMs ─────────────────────────────────────────────────────────────────────
DO $$ BEGIN
    CREATE TYPE inspection_status AS ENUM ('pending','in_progress','passed','failed','cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
    CREATE TYPE inspection_type AS ENUM ('incoming','in_process','final','audit','periodic');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
    CREATE TYPE check_result AS ENUM ('pass','fail','na','observation');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
    CREATE TYPE nc_status AS ENUM ('open','under_review','corrective_action','closed','cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
    CREATE TYPE nc_severity AS ENUM ('minor','major','critical','critical_safety');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
    CREATE TYPE ca_status AS ENUM ('open','in_progress','pending_verification','verified','closed','cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
    CREATE TYPE ca_type AS ENUM ('corrective','preventive','improvement');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

-- ── quality_control_plans ─────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS quality_control_plans (
    id               UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id       UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code             VARCHAR(30)  NOT NULL,
    name             VARCHAR(200) NOT NULL,
    description      TEXT,
    version          VARCHAR(20)  NOT NULL DEFAULT '1.0',
    item_id          UUID         REFERENCES items(id),
    item_category_id UUID         REFERENCES item_categories(id),
    applies_to       VARCHAR(50)  NOT NULL DEFAULT 'all'
                         CHECK (applies_to IN ('all','incoming','in_process','final','item','category')),
    is_active        BOOLEAN      NOT NULL DEFAULT TRUE,
    created_by       UUID         REFERENCES users(id),
    approved_by      UUID         REFERENCES users(id),
    approved_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);
CREATE INDEX IF NOT EXISTS idx_qcp_company ON quality_control_plans(company_id);
CREATE INDEX IF NOT EXISTS idx_qcp_item    ON quality_control_plans(item_id);

-- ── quality_check_templates ───────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS quality_check_templates (
    id             UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    plan_id        UUID         NOT NULL REFERENCES quality_control_plans(id) ON DELETE CASCADE,
    sequence       INT          NOT NULL DEFAULT 1,
    name           VARCHAR(200) NOT NULL,
    description    TEXT,
    check_type     VARCHAR(30)  NOT NULL DEFAULT 'visual'
                       CHECK (check_type IN ('visual','measurement','functional','document','count')),
    unit           VARCHAR(30),
    min_value      NUMERIC(18,4),
    max_value      NUMERIC(18,4),
    norm_reference VARCHAR(100),
    is_mandatory   BOOLEAN      NOT NULL DEFAULT TRUE,
    instructions   TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_qct_plan ON quality_check_templates(plan_id);

-- ── quality_inspections ───────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS quality_inspections (
    id              UUID              PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID              NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    reference       VARCHAR(50)       NOT NULL,
    inspection_type inspection_type   NOT NULL DEFAULT 'incoming',
    status          inspection_status NOT NULL DEFAULT 'pending',
    plan_id         UUID              REFERENCES quality_control_plans(id),
    item_id         UUID              REFERENCES items(id),
    lot_number      VARCHAR(100),
    qty_to_inspect  NUMERIC(14,4)     NOT NULL DEFAULT 0,
    qty_passed      NUMERIC(14,4)     NOT NULL DEFAULT 0,
    qty_failed      NUMERIC(14,4)     NOT NULL DEFAULT 0,
    overall_result  VARCHAR(20)       CHECK (overall_result IN ('passed','failed','conditional')),
    source_type     VARCHAR(50),
    source_id       UUID,
    source_ref      VARCHAR(100),
    scheduled_date  DATE,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    inspector_id    UUID              REFERENCES users(id),
    notes           TEXT,
    closure_notes   TEXT,
    created_by      UUID              REFERENCES users(id),
    created_at      TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, reference)
);
CREATE INDEX IF NOT EXISTS idx_qi_company ON quality_inspections(company_id);
CREATE INDEX IF NOT EXISTS idx_qi_status  ON quality_inspections(company_id, status);
CREATE INDEX IF NOT EXISTS idx_qi_item    ON quality_inspections(item_id);
CREATE INDEX IF NOT EXISTS idx_qi_date    ON quality_inspections(company_id, scheduled_date);

-- ── quality_checks ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS quality_checks (
    id             UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    inspection_id  UUID         NOT NULL REFERENCES quality_inspections(id) ON DELETE CASCADE,
    template_id    UUID         REFERENCES quality_check_templates(id),
    sequence       INT          NOT NULL DEFAULT 1,
    name           VARCHAR(200) NOT NULL,
    description    TEXT,
    check_type     VARCHAR(30)  NOT NULL DEFAULT 'visual'
                       CHECK (check_type IN ('visual','measurement','functional','document','count')),
    unit           VARCHAR(30),
    min_value      NUMERIC(18,4),
    max_value      NUMERIC(18,4),
    norm_reference VARCHAR(100),
    instructions   TEXT,
    is_mandatory   BOOLEAN      NOT NULL DEFAULT TRUE,
    result         check_result,
    measured_value NUMERIC(18,4),
    notes          TEXT,
    checked_by     UUID         REFERENCES users(id),
    checked_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_qchk_inspection ON quality_checks(inspection_id);
CREATE INDEX IF NOT EXISTS idx_qchk_result     ON quality_checks(inspection_id, result);

-- ── non_conformities ──────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS non_conformities (
    id               UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id       UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    reference        VARCHAR(50)  NOT NULL,
    title            VARCHAR(200) NOT NULL,
    description      TEXT,
    status           nc_status    NOT NULL DEFAULT 'open',
    severity         nc_severity  NOT NULL DEFAULT 'minor',
    source_type      VARCHAR(50),
    source_id        UUID,
    source_ref       VARCHAR(100),
    inspection_id    UUID         REFERENCES quality_inspections(id),
    item_id          UUID         REFERENCES items(id),
    lot_number       VARCHAR(100),
    qty_affected     NUMERIC(14,4),
    department_id    UUID         REFERENCES departments(id),
    process          VARCHAR(200),
    detected_by      UUID         REFERENCES users(id),
    detected_date    DATE         NOT NULL DEFAULT CURRENT_DATE,
    assigned_to      UUID         REFERENCES users(id),
    target_date      DATE,
    closed_date      DATE,
    root_cause       TEXT,
    immediate_action TEXT,
    closure_notes    TEXT,
    created_by       UUID         REFERENCES users(id),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, reference)
);
CREATE INDEX IF NOT EXISTS idx_nc_company  ON non_conformities(company_id);
CREATE INDEX IF NOT EXISTS idx_nc_status   ON non_conformities(company_id, status);
CREATE INDEX IF NOT EXISTS idx_nc_severity ON non_conformities(company_id, severity);
CREATE INDEX IF NOT EXISTS idx_nc_inspect  ON non_conformities(inspection_id);

-- ── corrective_actions ────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS corrective_actions (
    id                   UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id           UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    reference            VARCHAR(50)  NOT NULL,
    title                VARCHAR(200) NOT NULL,
    description          TEXT,
    ca_type              ca_type      NOT NULL DEFAULT 'corrective',
    status               ca_status    NOT NULL DEFAULT 'open',
    priority             VARCHAR(20)  NOT NULL DEFAULT 'medium'
                             CHECK (priority IN ('low','medium','high','critical')),
    nc_id                UUID         REFERENCES non_conformities(id),
    root_cause           TEXT,
    root_cause_method    VARCHAR(50)
                             CHECK (root_cause_method IN ('5why','fishbone','fmea','brainstorming','other')),
    proposed_action      TEXT,
    implemented_action   TEXT,
    responsible_id       UUID         REFERENCES users(id),
    department_id        UUID         REFERENCES departments(id),
    due_date             DATE,
    implementation_date  DATE,
    verified_by          UUID         REFERENCES users(id),
    verified_date        DATE,
    effectiveness_rating INT          CHECK (effectiveness_rating BETWEEN 1 AND 5),
    effectiveness_notes  TEXT,
    estimated_cost       NUMERIC(18,2) NOT NULL DEFAULT 0,
    actual_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    closed_date          DATE,
    created_by           UUID         REFERENCES users(id),
    created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, reference)
);
CREATE INDEX IF NOT EXISTS idx_ca_company  ON corrective_actions(company_id);
CREATE INDEX IF NOT EXISTS idx_ca_status   ON corrective_actions(company_id, status);
CREATE INDEX IF NOT EXISTS idx_ca_nc       ON corrective_actions(nc_id);
CREATE INDEX IF NOT EXISTS idx_ca_due_date ON corrective_actions(company_id, due_date);

-- ── ca_tasks ──────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS ca_tasks (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    ca_id           UUID         NOT NULL REFERENCES corrective_actions(id) ON DELETE CASCADE,
    title           VARCHAR(200) NOT NULL,
    description     TEXT,
    assigned_to     UUID         REFERENCES users(id),
    due_date        DATE,
    completed_date  DATE,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','in_progress','completed','cancelled')),
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ca_tasks_ca ON ca_tasks(ca_id);

-- ── quality_metrics ───────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS quality_metrics (
    id              UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    metric_date     DATE         NOT NULL DEFAULT CURRENT_DATE,
    metric_type     VARCHAR(50)  NOT NULL,
    value           NUMERIC(18,4) NOT NULL DEFAULT 0,
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, metric_date, metric_type)
);
CREATE INDEX IF NOT EXISTS idx_qm_company ON quality_metrics(company_id, metric_date DESC);

-- =============================================================================
-- SECTION 4 : NUMBERING CONFIG SEEDS
-- =============================================================================
DO $$
DECLARE
    col_name TEXT;
BEGIN
    SELECT column_name INTO col_name
    FROM information_schema.columns
    WHERE table_name = 'numbering_config'
      AND column_name IN ('doc_type', 'document_type')
    LIMIT 1;

    IF col_name = 'doc_type' THEN
        INSERT INTO numbering_config (company_id, doc_type, prefix, next_number, padding, suffix)
        SELECT id, 'fleet_vehicle',     'VH', 1, 4, '' FROM companies ON CONFLICT (company_id, doc_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, doc_type, prefix, next_number, padding, suffix)
        SELECT id, 'fleet_driver',      'DR', 1, 4, '' FROM companies ON CONFLICT (company_id, doc_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, doc_type, prefix, next_number, padding, suffix)
        SELECT id, 'fleet_maintenance', 'FM', 1, 4, '' FROM companies ON CONFLICT (company_id, doc_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, doc_type, prefix, next_number, padding, suffix)
        SELECT id, 'maintenance_request', 'MR', 1, 4, '' FROM companies ON CONFLICT (company_id, doc_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, doc_type, prefix, next_number, padding, suffix)
        SELECT id, 'maintenance_order',   'MO', 1, 4, '' FROM companies ON CONFLICT (company_id, doc_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, doc_type, prefix, next_number, padding, suffix)
        SELECT id, 'quality_inspection', 'QI', 1, 5, '' FROM companies ON CONFLICT (company_id, doc_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, doc_type, prefix, next_number, padding, suffix)
        SELECT id, 'non_conformity',     'NC', 1, 5, '' FROM companies ON CONFLICT (company_id, doc_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, doc_type, prefix, next_number, padding, suffix)
        SELECT id, 'corrective_action',  'CA', 1, 5, '' FROM companies ON CONFLICT (company_id, doc_type) DO NOTHING;
    ELSIF col_name = 'document_type' THEN
        INSERT INTO numbering_config (company_id, document_type, prefix, next_number, padding, suffix)
        SELECT id, 'fleet_vehicle',     'VH', 1, 4, '' FROM companies ON CONFLICT (company_id, document_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, document_type, prefix, next_number, padding, suffix)
        SELECT id, 'fleet_driver',      'DR', 1, 4, '' FROM companies ON CONFLICT (company_id, document_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, document_type, prefix, next_number, padding, suffix)
        SELECT id, 'fleet_maintenance', 'FM', 1, 4, '' FROM companies ON CONFLICT (company_id, document_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, document_type, prefix, next_number, padding, suffix)
        SELECT id, 'maintenance_request', 'MR', 1, 4, '' FROM companies ON CONFLICT (company_id, document_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, document_type, prefix, next_number, padding, suffix)
        SELECT id, 'maintenance_order',   'MO', 1, 4, '' FROM companies ON CONFLICT (company_id, document_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, document_type, prefix, next_number, padding, suffix)
        SELECT id, 'quality_inspection', 'QI', 1, 5, '' FROM companies ON CONFLICT (company_id, document_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, document_type, prefix, next_number, padding, suffix)
        SELECT id, 'non_conformity',     'NC', 1, 5, '' FROM companies ON CONFLICT (company_id, document_type) DO NOTHING;
        INSERT INTO numbering_config (company_id, document_type, prefix, next_number, padding, suffix)
        SELECT id, 'corrective_action',  'CA', 1, 5, '' FROM companies ON CONFLICT (company_id, document_type) DO NOTHING;
    END IF;
END$$;

-- =============================================================================
-- SECTION 5 : RECORD MIGRATION VERSION
-- =============================================================================

-- Insert migration record only if it doesn't exist
INSERT INTO schema_migrations (version) 
SELECT '0016_fix_missing_tables'
WHERE NOT EXISTS (
    SELECT 1 FROM schema_migrations WHERE version = '0016_fix_missing_tables'
);

-- Also back-fill previous migrations if not recorded
INSERT INTO schema_migrations (version) 
SELECT '0001_init_schema'
WHERE NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0001_init_schema');

INSERT INTO schema_migrations (version) 
SELECT '0002_sales_crm'
WHERE NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0002_sales_crm');

INSERT INTO schema_migrations (version) 
SELECT '0003_purchase'
WHERE NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0003_purchase');

INSERT INTO schema_migrations (version) 
SELECT '0004_settings_ext'
WHERE NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0004_settings_ext');

INSERT INTO schema_migrations (version) 
SELECT '0014_settings'
WHERE NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0014_settings');

INSERT INTO schema_migrations (version) 
SELECT '0015_quality'
WHERE NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0015_quality');