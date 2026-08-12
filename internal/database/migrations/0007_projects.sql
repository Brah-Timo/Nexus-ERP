-- =============================================================================
-- Migration 0007: Projects Extended Schema
-- Idempotent — safe to run multiple times
-- =============================================================================

-- ─── ENUM types ──────────────────────────────────────────────────────────────

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'milestone_status') THEN
        CREATE TYPE milestone_status AS ENUM ('pending','in_progress','completed','cancelled');
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'expense_category') THEN
        CREATE TYPE expense_category AS ENUM (
            'travel','accommodation','equipment','software','consulting',
            'materials','utilities','communication','other'
        );
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'expense_status') THEN
        CREATE TYPE expense_status AS ENUM ('draft','submitted','approved','rejected','paid');
    END IF;
END $$;

-- ─── 🔥 تحويل عمود status في project_tasks من ENUM إلى VARCHAR ─────────────

DO $$
BEGIN
    -- التحقق من وجود الجدول project_tasks
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'project_tasks') THEN
        -- تغيير نوع العمود status من ENUM إلى VARCHAR
        ALTER TABLE project_tasks ALTER COLUMN status TYPE VARCHAR(50);
        -- إضافة CHECK CONSTRAINT للقيم المسموحة
        ALTER TABLE project_tasks DROP CONSTRAINT IF EXISTS project_tasks_status_check;
        ALTER TABLE project_tasks ADD CONSTRAINT project_tasks_status_check 
            CHECK (status IN ('todo','in_progress','review','done','cancelled'));
    END IF;
END $$;

-- ─── project_milestones ───────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS project_milestones (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title           VARCHAR(300) NOT NULL,
    description     TEXT,
    due_date        DATE         NOT NULL,
    completed_at    TIMESTAMPTZ,
    status          milestone_status NOT NULL DEFAULT 'pending',
    owner_id        UUID         REFERENCES employees(id),
    progress_pct    INT          NOT NULL DEFAULT 0
                       CONSTRAINT milestone_progress_check CHECK (progress_pct BETWEEN 0 AND 100),
    sort_order      INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ─── project_expenses ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS project_expenses (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_id         UUID         REFERENCES project_tasks(id),
    employee_id     UUID         NOT NULL REFERENCES employees(id),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    category        expense_category NOT NULL DEFAULT 'other',
    description     VARCHAR(500) NOT NULL,
    amount          NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    expense_date    DATE         NOT NULL DEFAULT CURRENT_DATE,
    receipt_url     TEXT,
    status          expense_status NOT NULL DEFAULT 'draft',
    approved_by     UUID         REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    rejection_note  TEXT,
    is_billable     BOOLEAN      NOT NULL DEFAULT FALSE,
    billed          BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ─── project_comments ─────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS project_comments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_id         UUID         REFERENCES project_tasks(id),
    author_id       UUID         NOT NULL REFERENCES users(id),
    content         TEXT         NOT NULL,
    parent_id       UUID         REFERENCES project_comments(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ─── planning_slots ───────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS planning_slots (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_id         UUID         REFERENCES project_tasks(id),
    employee_id     UUID         NOT NULL REFERENCES employees(id),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    planned_date    DATE         NOT NULL,
    planned_hours   NUMERIC(5,2) NOT NULL DEFAULT 0,
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(task_id, employee_id, planned_date)
);

-- ─── Extend projects: missing columns ────────────────────────────────────────

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='projects' AND column_name='notes') THEN
        ALTER TABLE projects ADD COLUMN notes TEXT;
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='projects' AND column_name='color') THEN
        ALTER TABLE projects ADD COLUMN color VARCHAR(20) DEFAULT '#6366f1';
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='project_tasks' AND column_name='color') THEN
        ALTER TABLE project_tasks ADD COLUMN color VARCHAR(20) DEFAULT NULL;
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='project_tasks' AND column_name='tags') THEN
        ALTER TABLE project_tasks ADD COLUMN tags TEXT[] DEFAULT '{}';
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='project_tasks' AND column_name='start_date') THEN
        ALTER TABLE project_tasks ADD COLUMN start_date DATE;
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='timesheets' AND column_name='hourly_rate') THEN
        ALTER TABLE timesheets ADD COLUMN hourly_rate NUMERIC(10,2) NOT NULL DEFAULT 0;
    END IF;
END $$;

-- ─── Indexes ──────────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_projects_company_status ON projects(company_id, status);
CREATE INDEX IF NOT EXISTS idx_projects_manager ON projects(manager_id);
CREATE INDEX IF NOT EXISTS idx_project_tasks_project ON project_tasks(project_id, status);
CREATE INDEX IF NOT EXISTS idx_project_tasks_assignee ON project_tasks(assignee_id);
CREATE INDEX IF NOT EXISTS idx_project_tasks_due_date ON project_tasks(due_date);
CREATE INDEX IF NOT EXISTS idx_project_milestones_project ON project_milestones(project_id, status);
CREATE INDEX IF NOT EXISTS idx_project_milestones_due_date ON project_milestones(due_date);
CREATE INDEX IF NOT EXISTS idx_project_expenses_project ON project_expenses(project_id, status);
CREATE INDEX IF NOT EXISTS idx_project_expenses_employee ON project_expenses(employee_id);
CREATE INDEX IF NOT EXISTS idx_project_expenses_company ON project_expenses(company_id, expense_date);
CREATE INDEX IF NOT EXISTS idx_timesheets_company_date ON timesheets(company_id, date);
CREATE INDEX IF NOT EXISTS idx_timesheets_employee ON timesheets(employee_id);
CREATE INDEX IF NOT EXISTS idx_timesheets_task ON timesheets(task_id);
CREATE INDEX IF NOT EXISTS idx_planning_slots_project ON planning_slots(project_id, planned_date);
CREATE INDEX IF NOT EXISTS idx_planning_slots_employee ON planning_slots(employee_id, planned_date);
CREATE INDEX IF NOT EXISTS idx_project_comments_project ON project_comments(project_id);

-- ─── Views ───────────────────────────────────────────────────────────────────

CREATE OR REPLACE VIEW v_projects_detail AS
SELECT
    p.id,
    p.company_id,
    p.code,
    p.name,
    p.status,
    p.start_date,
    p.end_date,
    p.budget,
    p.actual_cost,
    p.progress_pct,
    p.description,
    p.notes,
    p.color,
    p.account_id,
    p.cost_center_id,
    p.created_at,
    p.updated_at,
    p.customer_id,
    c.name AS customer_name,
    p.manager_id,
    e.first_name || ' ' || e.last_name AS manager_name,
    COALESCE(task_stats.total_tasks, 0) AS total_tasks,
    COALESCE(task_stats.completed_tasks, 0) AS completed_tasks,
    COALESCE(task_stats.overdue_tasks, 0) AS overdue_tasks,
    COALESCE(ts_stats.total_hours, 0) AS total_hours,
    COALESCE(ts_stats.billable_hours, 0) AS billable_hours,
    COALESCE(exp_stats.total_expenses, 0) AS total_expenses,
    COALESCE(mile_stats.total_milestones, 0) AS total_milestones,
    COALESCE(mile_stats.completed_milestones, 0) AS completed_milestones
FROM projects p
LEFT JOIN customers c ON c.id = p.customer_id
LEFT JOIN employees e ON e.id = p.manager_id
LEFT JOIN LATERAL (
    SELECT
        COUNT(*) AS total_tasks,
        COUNT(*) FILTER (WHERE status = 'done') AS completed_tasks,
        COUNT(*) FILTER (WHERE due_date < CURRENT_DATE AND status NOT IN ('done','cancelled')) AS overdue_tasks
    FROM project_tasks WHERE project_id = p.id
) task_stats ON TRUE
LEFT JOIN LATERAL (
    SELECT
        COALESCE(SUM(hours), 0) AS total_hours,
        COALESCE(SUM(hours) FILTER (WHERE billable), 0) AS billable_hours
    FROM timesheets WHERE project_id = p.id
) ts_stats ON TRUE
LEFT JOIN LATERAL (
    SELECT COALESCE(SUM(amount), 0) AS total_expenses
    FROM project_expenses WHERE project_id = p.id AND status IN ('approved','paid')
) exp_stats ON TRUE
LEFT JOIN LATERAL (
    SELECT
        COUNT(*) AS total_milestones,
        COUNT(*) FILTER (WHERE status = 'completed') AS completed_milestones
    FROM project_milestones WHERE project_id = p.id
) mile_stats ON TRUE;

CREATE OR REPLACE VIEW v_project_tasks_detail AS
SELECT
    t.id,
    t.project_id,
    t.parent_id,
    t.title,
    t.description,
    t.assignee_id,
    t.status,
    t.priority,
    t.estimated_hours,
    t.actual_hours,
    t.start_date,
    t.due_date,
    t.completed_at,
    t.sort_order,
    t.color,
    t.tags,
    t.created_at,
    t.updated_at,
    p.name AS project_name,
    p.code AS project_code,
    p.color AS project_color,
    e.first_name || ' ' || e.last_name AS assignee_name,
    CASE WHEN t.due_date < CURRENT_DATE AND t.status NOT IN ('done','cancelled') THEN TRUE ELSE FALSE END AS is_overdue,
    COALESCE(sub.sub_count, 0) AS sub_task_count,
    COALESCE(sub.sub_done, 0) AS sub_task_done
FROM project_tasks t
LEFT JOIN projects p ON p.id = t.project_id
LEFT JOIN employees e ON e.id = t.assignee_id
LEFT JOIN LATERAL (
    SELECT COUNT(*) AS sub_count, COUNT(*) FILTER (WHERE status='done') AS sub_done
    FROM project_tasks WHERE parent_id = t.id
) sub ON TRUE;

CREATE OR REPLACE VIEW v_timesheets_detail AS
SELECT
    ts.id,
    ts.company_id,
    ts.employee_id,
    ts.project_id,
    ts.task_id,
    ts.date,
    ts.hours,
    ts.hourly_rate,
    ts.hours * ts.hourly_rate AS line_amount,
    ts.description,
    ts.billable,
    ts.billed,
    ts.approved,
    ts.approved_by,
    ts.created_at,
    e.first_name || ' ' || e.last_name AS employee_name,
    e.employee_number,
    p.name AS project_name,
    p.code AS project_code,
    t.title AS task_title
FROM timesheets ts
LEFT JOIN employees e ON e.id = ts.employee_id
LEFT JOIN projects p ON p.id = ts.project_id
LEFT JOIN project_tasks t ON t.id = ts.task_id;

CREATE OR REPLACE VIEW v_project_expenses_detail AS
SELECT
    ex.id,
    ex.project_id,
    ex.task_id,
    ex.employee_id,
    ex.company_id,
    ex.category,
    ex.description,
    ex.amount,
    ex.currency,
    ex.expense_date,
    ex.receipt_url,
    ex.status,
    ex.approved_by,
    ex.approved_at,
    ex.is_billable,
    ex.billed,
    ex.created_at,
    e.first_name || ' ' || e.last_name AS employee_name,
    p.name AS project_name,
    p.code AS project_code,
    t.title AS task_title
FROM project_expenses ex
LEFT JOIN employees e ON e.id = ex.employee_id
LEFT JOIN projects p ON p.id = ex.project_id
LEFT JOIN project_tasks t ON t.id = ex.task_id;

CREATE OR REPLACE VIEW v_project_planning AS
SELECT
    ps.id,
    ps.project_id,
    ps.task_id,
    ps.employee_id,
    ps.company_id,
    ps.planned_date,
    ps.planned_hours,
    ps.notes,
    ps.created_at,
    e.first_name || ' ' || e.last_name AS employee_name,
    e.employee_number,
    p.name AS project_name,
    p.code AS project_code,
    t.title AS task_title
FROM planning_slots ps
LEFT JOIN employees e ON e.id = ps.employee_id
LEFT JOIN projects p ON p.id = ps.project_id
LEFT JOIN project_tasks t ON t.id = ps.task_id;

-- ─── Functions ────────────────────────────────────────────────────────────────

CREATE OR REPLACE FUNCTION next_project_code(p_company_id UUID)
RETURNS VARCHAR AS $$
DECLARE
    v_seq INT;
BEGIN
    SELECT COALESCE(MAX(CAST(REGEXP_REPLACE(code, '[^0-9]', '', 'g') AS INT)), 0) + 1
    INTO v_seq
    FROM projects
    WHERE company_id = p_company_id
      AND code ~ '^PRJ-[0-9]+$';
    RETURN 'PRJ-' || LPAD(v_seq::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_project_progress(p_project_id UUID)
RETURNS VOID AS $$
DECLARE
    v_total INT;
    v_done  INT;
    v_pct   INT;
BEGIN
    SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'done')
    INTO v_total, v_done
    FROM project_tasks
    WHERE project_id = p_project_id AND parent_id IS NULL;

    IF v_total = 0 THEN
        v_pct := 0;
    ELSE
        v_pct := (v_done * 100 / v_total);
    END IF;

    UPDATE projects SET progress_pct = v_pct, updated_at = NOW()
    WHERE id = p_project_id;
END;
$$ LANGUAGE plpgsql;

-- ─── ❌ تم حذف جزء تسجيل الميجريشن ──────────────────────────────────────────────
-- النظام يقوم بالتسجيل تلقائياً