-- =============================================================================
-- Nexus ERP — Quality Management Module Migration
-- Version: 0015 | Full Quality schema (idempotent)
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- ENUMS
-- =============================================================================

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

-- =============================================================================
-- QUALITY CONTROL PLANS
-- =============================================================================

CREATE TABLE IF NOT EXISTS quality_control_plans (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code            VARCHAR(30)  NOT NULL,
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    version         VARCHAR(20)  NOT NULL DEFAULT '1.0',
    item_id         UUID         REFERENCES items(id),
    item_category_id UUID        REFERENCES item_categories(id),
    applies_to      VARCHAR(50)  NOT NULL DEFAULT 'all'
                        CHECK (applies_to IN ('all','incoming','in_process','final','item','category')),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_by      UUID         REFERENCES users(id),
    approved_by     UUID         REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE INDEX IF NOT EXISTS idx_qcp_company ON quality_control_plans(company_id);
CREATE INDEX IF NOT EXISTS idx_qcp_item    ON quality_control_plans(item_id);

-- =============================================================================
-- QUALITY CHECK TEMPLATES (criteria in a plan)
-- =============================================================================

CREATE TABLE IF NOT EXISTS quality_check_templates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    plan_id         UUID         NOT NULL REFERENCES quality_control_plans(id) ON DELETE CASCADE,
    sequence        INT          NOT NULL DEFAULT 1,
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    check_type      VARCHAR(30)  NOT NULL DEFAULT 'visual'
                        CHECK (check_type IN ('visual','measurement','functional','document','count')),
    unit            VARCHAR(30),
    min_value       NUMERIC(18,4),
    max_value       NUMERIC(18,4),
    norm_reference  VARCHAR(100),
    is_mandatory    BOOLEAN      NOT NULL DEFAULT TRUE,
    instructions    TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_qct_plan ON quality_check_templates(plan_id);

-- =============================================================================
-- QUALITY INSPECTIONS
-- =============================================================================

CREATE TABLE IF NOT EXISTS quality_inspections (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    reference       VARCHAR(50)  NOT NULL,
    inspection_type inspection_type NOT NULL DEFAULT 'incoming',
    status          inspection_status NOT NULL DEFAULT 'pending',
    plan_id         UUID         REFERENCES quality_control_plans(id),
    item_id         UUID         REFERENCES items(id),
    lot_number      VARCHAR(50),
    qty_to_inspect  NUMERIC(18,4) NOT NULL DEFAULT 0,
    qty_passed      NUMERIC(18,4) NOT NULL DEFAULT 0,
    qty_failed      NUMERIC(18,4) NOT NULL DEFAULT 0,
    -- Source document links
    source_type     VARCHAR(50),   -- 'purchase_order','manufacturing_order','stock_movement','manual'
    source_id       UUID,
    source_ref      VARCHAR(100),
    -- Dates
    scheduled_date  DATE,
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    -- Inspector
    inspector_id    UUID         REFERENCES users(id),
    -- Results
    overall_result  check_result,
    notes           TEXT,
    -- Signature
    approved_by     UUID         REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, reference)
);

CREATE INDEX IF NOT EXISTS idx_qi_company       ON quality_inspections(company_id, status);
CREATE INDEX IF NOT EXISTS idx_qi_type_date     ON quality_inspections(company_id, inspection_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_qi_item          ON quality_inspections(item_id);
CREATE INDEX IF NOT EXISTS idx_qi_source        ON quality_inspections(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_qi_inspector     ON quality_inspections(inspector_id);

-- =============================================================================
-- QUALITY CHECKS (per inspection line)
-- =============================================================================

CREATE TABLE IF NOT EXISTS quality_checks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    inspection_id   UUID         NOT NULL REFERENCES quality_inspections(id) ON DELETE CASCADE,
    template_id     UUID         REFERENCES quality_check_templates(id),
    sequence        INT          NOT NULL DEFAULT 1,
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    check_type      VARCHAR(30)  NOT NULL DEFAULT 'visual',
    unit            VARCHAR(30),
    min_value       NUMERIC(18,4),
    max_value       NUMERIC(18,4),
    measured_value  NUMERIC(18,4),
    result          check_result,
    is_mandatory    BOOLEAN      NOT NULL DEFAULT TRUE,
    norm_reference  VARCHAR(100),
    instructions    TEXT,
    notes           TEXT,
    image_url       TEXT,
    checked_by      UUID         REFERENCES users(id),
    checked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_qc_inspection ON quality_checks(inspection_id);
CREATE INDEX IF NOT EXISTS idx_qc_result     ON quality_checks(result);

-- =============================================================================
-- NON-CONFORMITIES
-- =============================================================================

CREATE TABLE IF NOT EXISTS non_conformities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    reference       VARCHAR(50)  NOT NULL,
    title           VARCHAR(300) NOT NULL,
    description     TEXT,
    status          nc_status    NOT NULL DEFAULT 'open',
    severity        nc_severity  NOT NULL DEFAULT 'minor',
    -- Source
    source_type     VARCHAR(50),  -- 'inspection','customer_complaint','internal_audit','supplier','production'
    source_id       UUID,
    source_ref      VARCHAR(100),
    -- Linked inspection
    inspection_id   UUID         REFERENCES quality_inspections(id),
    check_id        UUID         REFERENCES quality_checks(id),
    -- Item/lot
    item_id         UUID         REFERENCES items(id),
    lot_number      VARCHAR(50),
    qty_affected    NUMERIC(18,4),
    -- Department / process
    department_id   UUID         REFERENCES departments(id),
    process         VARCHAR(100),
    -- Detected by
    detected_by     UUID         REFERENCES users(id),
    detected_date   DATE         NOT NULL DEFAULT CURRENT_DATE,
    -- Assigned to
    assigned_to     UUID         REFERENCES users(id),
    target_date     DATE,
    -- Closure
    root_cause      TEXT,
    immediate_action TEXT,
    closed_by       UUID         REFERENCES users(id),
    closed_at       TIMESTAMPTZ,
    closure_notes   TEXT,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, reference)
);

CREATE INDEX IF NOT EXISTS idx_nc_company     ON non_conformities(company_id, status);
CREATE INDEX IF NOT EXISTS idx_nc_severity    ON non_conformities(company_id, severity);
CREATE INDEX IF NOT EXISTS idx_nc_item        ON non_conformities(item_id);
CREATE INDEX IF NOT EXISTS idx_nc_inspection  ON non_conformities(inspection_id);
CREATE INDEX IF NOT EXISTS idx_nc_assigned    ON non_conformities(assigned_to);

-- =============================================================================
-- CORRECTIVE ACTIONS
-- =============================================================================

CREATE TABLE IF NOT EXISTS corrective_actions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    reference       VARCHAR(50)  NOT NULL,
    title           VARCHAR(300) NOT NULL,
    description     TEXT,
    ca_type         ca_type      NOT NULL DEFAULT 'corrective',
    status          ca_status    NOT NULL DEFAULT 'open',
    priority        VARCHAR(20)  NOT NULL DEFAULT 'medium'
                        CHECK (priority IN ('low','medium','high','critical')),
    -- Linked NC
    nc_id           UUID         REFERENCES non_conformities(id),
    -- Root cause & analysis
    root_cause      TEXT,
    root_cause_method VARCHAR(50) DEFAULT '5why'
                        CHECK (root_cause_method IN ('5why','fishbone','fmea','brainstorming','other')),
    -- Actions
    proposed_action TEXT,
    implemented_action TEXT,
    -- Responsibility
    responsible_id  UUID         REFERENCES users(id),
    department_id   UUID         REFERENCES departments(id),
    -- Dates
    due_date        DATE,
    implementation_date DATE,
    verification_date   DATE,
    -- Effectiveness
    effectiveness_rating INT      CHECK (effectiveness_rating BETWEEN 1 AND 5),
    effectiveness_notes  TEXT,
    verified_by     UUID         REFERENCES users(id),
    -- Closure
    closed_by       UUID         REFERENCES users(id),
    closed_at       TIMESTAMPTZ,
    -- Costs
    estimated_cost  NUMERIC(18,2) NOT NULL DEFAULT 0,
    actual_cost     NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, reference)
);

CREATE INDEX IF NOT EXISTS idx_ca_company  ON corrective_actions(company_id, status);
CREATE INDEX IF NOT EXISTS idx_ca_nc       ON corrective_actions(nc_id);
CREATE INDEX IF NOT EXISTS idx_ca_resp     ON corrective_actions(responsible_id);
CREATE INDEX IF NOT EXISTS idx_ca_due      ON corrective_actions(due_date);

-- =============================================================================
-- CORRECTIVE ACTION TASKS (sub-steps)
-- =============================================================================

CREATE TABLE IF NOT EXISTS ca_tasks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ca_id           UUID         NOT NULL REFERENCES corrective_actions(id) ON DELETE CASCADE,
    sequence        INT          NOT NULL DEFAULT 1,
    description     TEXT         NOT NULL,
    assigned_to     UUID         REFERENCES users(id),
    due_date        DATE,
    completed       BOOLEAN      NOT NULL DEFAULT FALSE,
    completed_at    TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ca_tasks_ca ON ca_tasks(ca_id);

-- =============================================================================
-- QUALITY KPIs / METRICS (stored aggregates for dashboard)
-- =============================================================================

CREATE TABLE IF NOT EXISTS quality_metrics (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    period_year     INT          NOT NULL,
    period_month    INT          NOT NULL,
    total_inspections   INT      NOT NULL DEFAULT 0,
    passed_inspections  INT      NOT NULL DEFAULT 0,
    failed_inspections  INT      NOT NULL DEFAULT 0,
    total_checks        INT      NOT NULL DEFAULT 0,
    failed_checks       INT      NOT NULL DEFAULT 0,
    total_nc            INT      NOT NULL DEFAULT 0,
    open_nc             INT      NOT NULL DEFAULT 0,
    closed_nc           INT      NOT NULL DEFAULT 0,
    critical_nc         INT      NOT NULL DEFAULT 0,
    total_ca            INT      NOT NULL DEFAULT 0,
    open_ca             INT      NOT NULL DEFAULT 0,
    overdue_ca          INT      NOT NULL DEFAULT 0,
    avg_closure_days    NUMERIC(8,2),
    first_pass_rate     NUMERIC(8,4),   -- percentage
    defect_rate         NUMERIC(8,4),   -- percentage
    computed_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, period_year, period_month)
);

CREATE INDEX IF NOT EXISTS idx_qm_company ON quality_metrics(company_id, period_year DESC, period_month DESC);

-- =============================================================================
-- NUMBERING: add quality doc types
-- =============================================================================

INSERT INTO numbering_config (company_id, doc_type, prefix, padding, reset_yearly)
SELECT id, 'quality_inspection', 'QI-', 5, true FROM companies
ON CONFLICT DO NOTHING;

INSERT INTO numbering_config (company_id, doc_type, prefix, padding, reset_yearly)
SELECT id, 'non_conformity', 'NC-', 5, true FROM companies
ON CONFLICT DO NOTHING;

INSERT INTO numbering_config (company_id, doc_type, prefix, padding, reset_yearly)
SELECT id, 'corrective_action', 'CA-', 5, true FROM companies
ON CONFLICT DO NOTHING;

-- =============================================================================
-- Migration tracking
-- =============================================================================

INSERT INTO schema_migrations (version) VALUES ('0015_quality') ON CONFLICT DO NOTHING;
