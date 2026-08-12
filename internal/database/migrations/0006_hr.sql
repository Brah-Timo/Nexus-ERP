-- =============================================================================
-- Migration 0006 — HR & Payroll extended schema + Recruitment module
-- (بدون تسجيل يدوي - النظام يسجل تلقائياً)
-- =============================================================================

-- ── 1. حذف الجداول والأنواع القديمة ──────────────────────────────────────────

DROP TABLE IF EXISTS job_applications CASCADE;
DROP TABLE IF EXISTS job_postings CASCADE;
DROP TABLE IF EXISTS employee_deductions CASCADE;
DROP TABLE IF EXISTS employee_allowances CASCADE;

DROP TYPE IF EXISTS application_status CASCADE;
DROP TYPE IF EXISTS job_status CASCADE;

-- ── 2. إنشاء الأنواع ──────────────────────────────────────────────────────────

CREATE TYPE job_status AS ENUM ('draft','open','closed','on_hold');

CREATE TYPE application_status AS ENUM (
    'new','screening','interview','offer','hired','rejected','withdrawn'
);

-- ── 3. إنشاء الجداول ──────────────────────────────────────────────────────────

CREATE TABLE job_postings (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    department_id   UUID         REFERENCES departments(id),
    position_id     UUID         REFERENCES positions(id),
    title           VARCHAR(200) NOT NULL,
    description     TEXT,
    requirements    TEXT,
    location        VARCHAR(200),
    employment_type VARCHAR(30)  NOT NULL DEFAULT 'permanent',
    vacancies       INT          NOT NULL DEFAULT 1,
    status          job_status   NOT NULL DEFAULT 'draft',
    published_at    TIMESTAMPTZ,
    deadline_date   DATE,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE job_applications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_posting_id  UUID         NOT NULL REFERENCES job_postings(id) ON DELETE CASCADE,
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    email           VARCHAR(200),
    phone           VARCHAR(30),
    cv_url          TEXT,
    cover_letter    TEXT,
    source          VARCHAR(50)  DEFAULT 'direct',
    status          application_status NOT NULL DEFAULT 'new',
    expected_salary NUMERIC(18,2),
    interview_date  TIMESTAMPTZ,
    interview_notes TEXT,
    rejection_reason TEXT,
    hired_as_employee_id UUID REFERENCES employees(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE employee_allowances (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID         NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    amount          NUMERIC(18,2) NOT NULL DEFAULT 0,
    is_taxable      BOOLEAN      NOT NULL DEFAULT TRUE,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE employee_deductions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id     UUID         NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    amount          NUMERIC(18,2) NOT NULL DEFAULT 0,
    is_recurring    BOOLEAN      NOT NULL DEFAULT FALSE,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ── 4. الفهارس ──────────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_departments_company ON departments(company_id);
CREATE INDEX IF NOT EXISTS idx_positions_company ON positions(company_id, department_id);
CREATE INDEX IF NOT EXISTS idx_attendance_employee ON attendance(employee_id, date);
CREATE INDEX IF NOT EXISTS idx_attendance_date ON attendance(date);
CREATE INDEX IF NOT EXISTS idx_leave_requests_emp ON leave_requests(employee_id, status);
CREATE INDEX IF NOT EXISTS idx_leave_requests_comp ON leave_requests(company_id, status);
CREATE INDEX IF NOT EXISTS idx_payroll_runs_company ON payroll_runs(company_id, period_year, period_month);
CREATE INDEX IF NOT EXISTS idx_payslips_run ON payslips(payroll_run_id);
CREATE INDEX IF NOT EXISTS idx_payslips_employee ON payslips(employee_id);
CREATE INDEX IF NOT EXISTS idx_job_postings_company ON job_postings(company_id, status);
CREATE INDEX IF NOT EXISTS idx_job_applications_job ON job_applications(job_posting_id, status);
CREATE INDEX IF NOT EXISTS idx_job_applications_comp ON job_applications(company_id);

-- ── 5. المشاهدات ──────────────────────────────────────────────────────────────────

CREATE OR REPLACE VIEW v_employees_detail AS
SELECT
    e.id,
    e.company_id,
    e.employee_number,
    e.first_name,
    e.last_name,
    e.first_name || ' ' || e.last_name AS full_name,
    e.gender,
    e.birth_date,
    e.hire_date,
    e.termination_date,
    e.national_id,
    e.cnas_number,
    e.nif,
    e.department_id,
    d.name AS department_name,
    d.code AS department_code,
    e.position_id,
    p.title AS position_title,
    p.grade AS position_grade,
    e.manager_id,
    m.first_name || ' ' || m.last_name AS manager_name,
    e.employment_type,
    e.status,
    e.base_salary,
    e.bank_account,
    e.bank_name,
    e.email,
    e.phone,
    e.address,
    e.city,
    e.wilaya,
    e.notes,
    e.created_at,
    e.updated_at
FROM employees e
LEFT JOIN departments d ON d.id = e.department_id
LEFT JOIN positions p ON p.id = e.position_id
LEFT JOIN employees m ON m.id = e.manager_id;

CREATE OR REPLACE VIEW v_attendance_detail AS
SELECT
    a.id,
    a.employee_id,
    e.first_name || ' ' || e.last_name AS employee_name,
    e.employee_number,
    d.name AS department_name,
    a.date,
    a.check_in,
    a.check_out,
    a.hours_worked,
    a.overtime_hours,
    a.status,
    a.notes
FROM attendance a
JOIN employees e ON e.id = a.employee_id
LEFT JOIN departments d ON d.id = e.department_id;

CREATE OR REPLACE VIEW v_leave_requests_detail AS
SELECT
    lr.id,
    lr.company_id,
    lr.employee_id,
    e.first_name || ' ' || e.last_name AS employee_name,
    e.employee_number,
    d.name AS department_name,
    lt.name AS leave_type_name,
    lt.is_paid,
    lr.start_date,
    lr.end_date,
    lr.days_count,
    lr.reason,
    lr.status,
    lr.rejection_reason,
    lr.approved_at,
    lr.created_at
FROM leave_requests lr
JOIN employees e ON e.id = lr.employee_id
LEFT JOIN departments d ON d.id = e.department_id
LEFT JOIN leave_types lt ON lt.id = lr.leave_type_id;

CREATE OR REPLACE VIEW v_payroll_summary AS
SELECT
    pr.id,
    pr.company_id,
    pr.period_month,
    pr.period_year,
    pr.status,
    pr.total_gross,
    pr.total_irg,
    pr.total_cnas_employee,
    pr.total_cnas_employer,
    pr.total_net,
    pr.total_employees,
    pr.approved_at,
    pr.paid_at,
    pr.created_at
FROM payroll_runs pr;

-- ── 6. الدالة ──────────────────────────────────────────────────────────────────

CREATE OR REPLACE FUNCTION next_employee_number(p_company_id UUID)
RETURNS VARCHAR AS $$
BEGIN
    RETURN 'EMP-' || LPAD(((SELECT COUNT(*) FROM employees WHERE company_id = p_company_id) + 1)::TEXT, 4, '0');
END;
$$ LANGUAGE plpgsql;

-- ── 7. إضافة قيد فريد لجدول leave_types ──────────────────────────────────────

ALTER TABLE leave_types DROP CONSTRAINT IF EXISTS leave_types_company_id_name_key;
ALTER TABLE leave_types ADD CONSTRAINT leave_types_company_id_name_key UNIQUE (company_id, name);

-- ── 8. إدخال البيانات الافتراضية ──────────────────────────────────────────────

INSERT INTO leave_types (id, company_id, name, days_allowed, is_paid, color)
SELECT
    gen_random_uuid(), c.id, lt.name, lt.days, lt.paid, lt.color
FROM companies c
CROSS JOIN (
    VALUES
        ('Congé annuel', 30, true, '#22c55e'),
        ('Congé maladie', 15, true, '#f59e0b'),
        ('Congé maternité', 98, true, '#ec4899'),
        ('Congé sans solde', 30, false, '#94a3b8'),
        ('Congé exceptionnel', 3, true, '#6366f1')
) AS lt(name, days, paid, color)
WHERE NOT EXISTS (
    SELECT 1 FROM leave_types x 
    WHERE x.company_id = c.id AND x.name = lt.name
);

-- ── ❌ تم حذف جزء تسجيل الميجريشن ──────────────────────────────────────────────
-- النظام يقوم بتسجيل الميجريشن تلقائياً، لذلك لا نحتاج إلى إدراج يدوي