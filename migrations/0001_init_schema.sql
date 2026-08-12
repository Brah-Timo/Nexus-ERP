-- =============================================================================
-- Nexus ERP — Full PostgreSQL Schema
-- Version: 1.0.0  |  Algerian SCF/IRG/CNAS Compliant
-- =============================================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- =============================================================================
-- ENUMS
-- =============================================================================

CREATE TYPE user_role AS ENUM (
    'super_admin','admin','accountant','hr_manager','sales_manager',
    'purchase_manager','inventory_manager','project_manager','viewer','approver'
);

CREATE TYPE account_type AS ENUM (
    'asset','liability','equity','revenue','expense','contra'
);

CREATE TYPE account_nature AS ENUM ('debit','credit');

CREATE TYPE journal_status AS ENUM ('draft','posted','cancelled');

CREATE TYPE invoice_status AS ENUM (
    'draft','confirmed','partially_paid','paid','cancelled','overdue'
);

CREATE TYPE po_status AS ENUM (
    'draft','pending_approval','approved','partially_received','received','cancelled'
);

CREATE TYPE employee_status AS ENUM ('active','inactive','on_leave','terminated');

CREATE TYPE payroll_status AS ENUM ('draft','calculated','approved','paid');

CREATE TYPE leave_status AS ENUM ('pending','approved','rejected','cancelled');

CREATE TYPE mo_status AS ENUM (
    'draft','planned','in_progress','completed','cancelled'
);

CREATE TYPE project_status AS ENUM (
    'planning','active','on_hold','completed','cancelled'
);

CREATE TYPE task_status AS ENUM (
    'backlog','todo','in_progress','review','done'
);

CREATE TYPE task_priority AS ENUM ('low','medium','high','critical');

CREATE TYPE cheque_status AS ENUM (
    'pending','deposited','cleared','bounced','cancelled'
);

CREATE TYPE payment_type AS ENUM (
    'cash','bank_transfer','cheque','card','other'
);

CREATE TYPE movement_type AS ENUM (
    'purchase','sale','transfer','adjustment','production_in','production_out','return_in','return_out'
);

CREATE TYPE approval_status AS ENUM ('pending','approved','rejected','cancelled');

CREATE TYPE asset_status AS ENUM ('active','fully_depreciated','disposed','idle');

CREATE TYPE depreciation_method AS ENUM ('linear','diminishing_balance');

CREATE TYPE opportunity_stage AS ENUM (
    'lead','qualified','proposal','negotiation','won','lost'
);

-- =============================================================================
-- CORE / MULTI-TENANT
-- =============================================================================

CREATE TABLE tenants (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code        VARCHAR(20)  NOT NULL UNIQUE,
    name        VARCHAR(200) NOT NULL,
    plan        VARCHAR(50)  NOT NULL DEFAULT 'standard',
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    settings    JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE companies (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code            VARCHAR(20)  NOT NULL,
    name            VARCHAR(200) NOT NULL,
    legal_name      VARCHAR(300),
    nif             VARCHAR(20),   -- Numéro d'Identification Fiscale
    nis             VARCHAR(20),   -- Numéro d'Identification Statistique
    rc              VARCHAR(30),   -- Registre de Commerce
    art             VARCHAR(30),   -- Article d'Imposition
    address         TEXT,
    city            VARCHAR(100),
    wilaya          VARCHAR(100),
    postal_code     VARCHAR(10),
    phone           VARCHAR(30),
    email           VARCHAR(200),
    website         VARCHAR(200),
    logo_url        TEXT,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    fiscal_year_start INT        NOT NULL DEFAULT 1,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    settings        JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, code)
);

CREATE TABLE branches (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code        VARCHAR(20)  NOT NULL,
    name        VARCHAR(200) NOT NULL,
    address     TEXT,
    city        VARCHAR(100),
    phone       VARCHAR(30),
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE fiscal_years (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name            VARCHAR(50) NOT NULL,
    start_date      DATE        NOT NULL,
    end_date        DATE        NOT NULL,
    is_closed       BOOLEAN     NOT NULL DEFAULT FALSE,
    closed_at       TIMESTAMPTZ,
    closed_by       UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, name)
);

CREATE TABLE currencies (
    code        VARCHAR(10)  PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    symbol      VARCHAR(10)  NOT NULL,
    rate_to_dzd NUMERIC(18,6) NOT NULL DEFAULT 1,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO currencies (code, name, symbol, rate_to_dzd) VALUES
    ('DZD', 'Algerian Dinar',    'DA',  1.0),
    ('EUR', 'Euro',               '€',   145.0),
    ('USD', 'US Dollar',          '$',   134.0),
    ('GBP', 'British Pound',      '£',   169.0);

-- =============================================================================
-- USERS & AUTH
-- =============================================================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID         NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    company_id      UUID         REFERENCES companies(id),
    branch_id       UUID         REFERENCES branches(id),
    username        VARCHAR(100) NOT NULL,
    email           VARCHAR(200) NOT NULL,
    password_hash   TEXT         NOT NULL,
    full_name       VARCHAR(200) NOT NULL,
    role            user_role    NOT NULL DEFAULT 'viewer',
    permissions     JSONB        NOT NULL DEFAULT '{}',
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    refresh_token   TEXT,
    reset_token     TEXT,
    reset_token_expires TIMESTAMPTZ,
    avatar_url      TEXT,
    preferences     JSONB        NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, username),
    UNIQUE(tenant_id, email)
);

CREATE TABLE audit_logs (
    id          BIGSERIAL    PRIMARY KEY,
    tenant_id   UUID         NOT NULL,
    company_id  UUID,
    user_id     UUID         REFERENCES users(id),
    action      VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id   TEXT,
    old_data    JSONB,
    new_data    JSONB,
    ip_address  INET,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant_created  ON audit_logs(tenant_id, created_at DESC);
CREATE INDEX idx_audit_logs_user_id         ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity          ON audit_logs(entity_type, entity_id);

-- =============================================================================
-- NUMBERING SEQUENCES
-- =============================================================================

CREATE TABLE numbering_sequences (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    doc_type    VARCHAR(50)  NOT NULL,
    prefix      VARCHAR(20)  NOT NULL DEFAULT '',
    suffix      VARCHAR(20)  NOT NULL DEFAULT '',
    next_number INT          NOT NULL DEFAULT 1,
    padding     INT          NOT NULL DEFAULT 5,
    reset_yearly BOOLEAN     NOT NULL DEFAULT TRUE,
    current_year INT,
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, doc_type)
);

-- =============================================================================
-- ACCOUNTING — CHART OF ACCOUNTS (SCF)
-- =============================================================================

CREATE TABLE chart_of_accounts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code            VARCHAR(20)  NOT NULL,
    name            VARCHAR(300) NOT NULL,
    type            account_type NOT NULL,
    nature          account_nature NOT NULL,
    parent_id       UUID         REFERENCES chart_of_accounts(id),
    is_group        BOOLEAN      NOT NULL DEFAULT FALSE,
    is_reconcilable BOOLEAN      NOT NULL DEFAULT FALSE,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    balance         NUMERIC(18,2) NOT NULL DEFAULT 0,
    debit_balance   NUMERIC(18,2) NOT NULL DEFAULT 0,
    credit_balance  NUMERIC(18,2) NOT NULL DEFAULT 0,
    level           INT          NOT NULL DEFAULT 1,
    description     TEXT,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE INDEX idx_coa_company_type ON chart_of_accounts(company_id, type);
CREATE INDEX idx_coa_parent       ON chart_of_accounts(parent_id);
CREATE INDEX idx_coa_code         ON chart_of_accounts(company_id, code);

-- Default SCF Chart of Accounts (Class 1-7 + Tax accounts)
-- These are inserted when a company is created (handled in code), but seeding here for reference
INSERT INTO currencies(code,name,symbol,rate_to_dzd) VALUES ('XOF','CFA Franc','FCFA',0.22) ON CONFLICT DO NOTHING;

CREATE TABLE cost_centers (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code        VARCHAR(20)  NOT NULL,
    name        VARCHAR(200) NOT NULL,
    parent_id   UUID         REFERENCES cost_centers(id),
    budget      NUMERIC(18,2) NOT NULL DEFAULT 0,
    actual      NUMERIC(18,2) NOT NULL DEFAULT 0,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE journal_entries (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id       UUID         REFERENCES branches(id),
    fiscal_year_id  UUID         REFERENCES fiscal_years(id),
    number          VARCHAR(50)  NOT NULL,
    date            DATE         NOT NULL,
    reference       VARCHAR(100),
    description     TEXT,
    status          journal_status NOT NULL DEFAULT 'draft',
    total_debit     NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_credit    NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    exchange_rate   NUMERIC(18,6) NOT NULL DEFAULT 1,
    source_type     VARCHAR(50),  -- 'manual','sale','purchase','payroll','depreciation'
    source_id       UUID,
    cost_center_id  UUID         REFERENCES cost_centers(id),
    posted_at       TIMESTAMPTZ,
    posted_by       UUID         REFERENCES users(id),
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE INDEX idx_je_company_date   ON journal_entries(company_id, date DESC);
CREATE INDEX idx_je_status         ON journal_entries(company_id, status);
CREATE INDEX idx_je_source         ON journal_entries(source_type, source_id);
CREATE INDEX idx_je_fiscal_year    ON journal_entries(fiscal_year_id);

CREATE TABLE journal_lines (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    journal_entry_id UUID        NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id      UUID         NOT NULL REFERENCES chart_of_accounts(id),
    cost_center_id  UUID         REFERENCES cost_centers(id),
    description     TEXT,
    debit           NUMERIC(18,2) NOT NULL DEFAULT 0,
    credit          NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    amount_currency NUMERIC(18,2) NOT NULL DEFAULT 0,
    reconciled      BOOLEAN      NOT NULL DEFAULT FALSE,
    reconciled_at   TIMESTAMPTZ,
    partner_id      UUID,
    partner_type    VARCHAR(50),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jl_entry         ON journal_lines(journal_entry_id);
CREATE INDEX idx_jl_account       ON journal_lines(account_id);

-- =============================================================================
-- FIXED ASSETS
-- =============================================================================

CREATE TABLE fixed_assets (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id          UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code                VARCHAR(30)  NOT NULL,
    name                VARCHAR(200) NOT NULL,
    category            VARCHAR(100),
    account_id          UUID         REFERENCES chart_of_accounts(id),
    depreciation_account_id UUID     REFERENCES chart_of_accounts(id),
    accumulated_account_id  UUID     REFERENCES chart_of_accounts(id),
    purchase_date       DATE         NOT NULL,
    in_service_date     DATE,
    purchase_value      NUMERIC(18,2) NOT NULL DEFAULT 0,
    residual_value      NUMERIC(18,2) NOT NULL DEFAULT 0,
    current_value       NUMERIC(18,2) NOT NULL DEFAULT 0,
    accumulated_depreciation NUMERIC(18,2) NOT NULL DEFAULT 0,
    depreciation_method depreciation_method NOT NULL DEFAULT 'linear',
    useful_life_years   INT          NOT NULL DEFAULT 5,
    depreciation_rate   NUMERIC(8,4) NOT NULL DEFAULT 0.20,
    last_depreciation_date DATE,
    disposal_date       DATE,
    disposal_value      NUMERIC(18,2),
    status              asset_status NOT NULL DEFAULT 'active',
    description         TEXT,
    location            VARCHAR(200),
    serial_number       VARCHAR(100),
    supplier_id         UUID,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE asset_depreciation_lines (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    asset_id    UUID         NOT NULL REFERENCES fixed_assets(id) ON DELETE CASCADE,
    period_date DATE         NOT NULL,
    amount      NUMERIC(18,2) NOT NULL,
    journal_entry_id UUID    REFERENCES journal_entries(id),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- BUDGETS
-- =============================================================================

CREATE TABLE budgets (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    fiscal_year_id  UUID         NOT NULL REFERENCES fiscal_years(id),
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft',
    total_budget    NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_actual    NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE budget_lines (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    budget_id       UUID         NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    account_id      UUID         NOT NULL REFERENCES chart_of_accounts(id),
    cost_center_id  UUID         REFERENCES cost_centers(id),
    jan NUMERIC(18,2) NOT NULL DEFAULT 0,
    feb NUMERIC(18,2) NOT NULL DEFAULT 0,
    mar NUMERIC(18,2) NOT NULL DEFAULT 0,
    apr NUMERIC(18,2) NOT NULL DEFAULT 0,
    may NUMERIC(18,2) NOT NULL DEFAULT 0,
    jun NUMERIC(18,2) NOT NULL DEFAULT 0,
    jul NUMERIC(18,2) NOT NULL DEFAULT 0,
    aug NUMERIC(18,2) NOT NULL DEFAULT 0,
    sep NUMERIC(18,2) NOT NULL DEFAULT 0,
    oct NUMERIC(18,2) NOT NULL DEFAULT 0,
    nov NUMERIC(18,2) NOT NULL DEFAULT 0,
    dec NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_budget    NUMERIC(18,2) GENERATED ALWAYS AS (
        jan+feb+mar+apr+may+jun+jul+aug+sep+oct+nov+dec
    ) STORED,
    total_actual    NUMERIC(18,2) NOT NULL DEFAULT 0
);

-- =============================================================================
-- HR / PAYROLL
-- =============================================================================

CREATE TABLE departments (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code        VARCHAR(20)  NOT NULL,
    name        VARCHAR(200) NOT NULL,
    parent_id   UUID         REFERENCES departments(id),
    manager_id  UUID,
    cost_center_id UUID      REFERENCES cost_centers(id),
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE positions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    department_id   UUID         REFERENCES departments(id),
    code            VARCHAR(20)  NOT NULL,
    title           VARCHAR(200) NOT NULL,
    grade           VARCHAR(50),
    min_salary      NUMERIC(18,2),
    max_salary      NUMERIC(18,2),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE employees (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id          UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id           UUID         REFERENCES branches(id),
    employee_number     VARCHAR(20)  NOT NULL,
    first_name          VARCHAR(100) NOT NULL,
    last_name           VARCHAR(100) NOT NULL,
    gender              VARCHAR(10),
    birth_date          DATE,
    hire_date           DATE         NOT NULL,
    termination_date    DATE,
    national_id         VARCHAR(30),
    cnas_number         VARCHAR(30),   -- Numéro SS CNAS
    nif                 VARCHAR(20),
    department_id       UUID         REFERENCES departments(id),
    position_id         UUID         REFERENCES positions(id),
    manager_id          UUID         REFERENCES employees(id),
    employment_type     VARCHAR(30)  NOT NULL DEFAULT 'permanent', -- permanent/contract/part_time
    status              employee_status NOT NULL DEFAULT 'active',
    base_salary         NUMERIC(18,2) NOT NULL DEFAULT 0,
    bank_account        VARCHAR(50),
    bank_name           VARCHAR(100),
    email               VARCHAR(200),
    phone               VARCHAR(30),
    address             TEXT,
    city                VARCHAR(100),
    wilaya              VARCHAR(100),
    avatar_url          TEXT,
    user_id             UUID         REFERENCES users(id),
    notes               TEXT,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, employee_number)
);

CREATE INDEX idx_employees_company    ON employees(company_id, status);
CREATE INDEX idx_employees_department ON employees(department_id);

CREATE TABLE leave_types (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID         NOT NULL REFERENCES companies(id),
    name        VARCHAR(100) NOT NULL,
    days_allowed INT         NOT NULL DEFAULT 30,
    is_paid     BOOLEAN      NOT NULL DEFAULT TRUE,
    color       VARCHAR(20),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE leave_requests (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    employee_id     UUID         NOT NULL REFERENCES employees(id),
    leave_type_id   UUID         REFERENCES leave_types(id),
    start_date      DATE         NOT NULL,
    end_date        DATE         NOT NULL,
    days_count      INT          NOT NULL DEFAULT 1,
    reason          TEXT,
    status          leave_status NOT NULL DEFAULT 'pending',
    approved_by     UUID         REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    rejection_reason TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE attendance (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID         NOT NULL REFERENCES employees(id),
    date        DATE         NOT NULL,
    check_in    TIMESTAMPTZ,
    check_out   TIMESTAMPTZ,
    hours_worked NUMERIC(5,2),
    overtime_hours NUMERIC(5,2) NOT NULL DEFAULT 0,
    status      VARCHAR(20)  NOT NULL DEFAULT 'present',
    notes       TEXT,
    UNIQUE(employee_id, date)
);

CREATE TABLE payroll_runs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    period_month    INT          NOT NULL,
    period_year     INT          NOT NULL,
    status          payroll_status NOT NULL DEFAULT 'draft',
    total_gross     NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_irg       NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_cnas_employee NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_cnas_employer NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_net       NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_employees INT          NOT NULL DEFAULT 0,
    approved_by     UUID         REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    paid_at         TIMESTAMPTZ,
    journal_entry_id UUID        REFERENCES journal_entries(id),
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, period_year, period_month)
);

CREATE TABLE payslips (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payroll_run_id      UUID         NOT NULL REFERENCES payroll_runs(id) ON DELETE CASCADE,
    employee_id         UUID         NOT NULL REFERENCES employees(id),
    period_month        INT          NOT NULL,
    period_year         INT          NOT NULL,
    days_worked         NUMERIC(5,2) NOT NULL DEFAULT 26,
    overtime_hours      NUMERIC(5,2) NOT NULL DEFAULT 0,
    base_salary         NUMERIC(18,2) NOT NULL DEFAULT 0,
    overtime_amount     NUMERIC(18,2) NOT NULL DEFAULT 0,
    transport_allowance NUMERIC(18,2) NOT NULL DEFAULT 0,
    meal_allowance      NUMERIC(18,2) NOT NULL DEFAULT 0,
    housing_allowance   NUMERIC(18,2) NOT NULL DEFAULT 0,
    other_allowances    NUMERIC(18,2) NOT NULL DEFAULT 0,
    gross_salary        NUMERIC(18,2) NOT NULL DEFAULT 0,
    cnas_employee       NUMERIC(18,2) NOT NULL DEFAULT 0,   -- 9%
    cnas_employer       NUMERIC(18,2) NOT NULL DEFAULT 0,   -- 26%
    taxable_income      NUMERIC(18,2) NOT NULL DEFAULT 0,
    irg_amount          NUMERIC(18,2) NOT NULL DEFAULT 0,
    other_deductions    NUMERIC(18,2) NOT NULL DEFAULT 0,
    advance_deduction   NUMERIC(18,2) NOT NULL DEFAULT 0,
    net_salary          NUMERIC(18,2) NOT NULL DEFAULT 0,
    irg_bracket         JSONB,         -- IRG calculation details
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE salary_advances (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID         NOT NULL REFERENCES employees(id),
    amount      NUMERIC(18,2) NOT NULL,
    date        DATE         NOT NULL,
    reason      TEXT,
    deducted    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- CRM / SALES
-- =============================================================================

CREATE TABLE customers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code            VARCHAR(20)  NOT NULL,
    name            VARCHAR(300) NOT NULL,
    type            VARCHAR(20)  NOT NULL DEFAULT 'company', -- company/individual
    nif             VARCHAR(20),
    nis             VARCHAR(20),
    rc              VARCHAR(30),
    art             VARCHAR(30),
    tax_regime      VARCHAR(20)  NOT NULL DEFAULT 'reel', -- reel/forfait/exonere
    address         TEXT,
    city            VARCHAR(100),
    wilaya          VARCHAR(100),
    postal_code     VARCHAR(10),
    phone           VARCHAR(30),
    email           VARCHAR(200),
    website         VARCHAR(200),
    credit_limit    NUMERIC(18,2) NOT NULL DEFAULT 0,
    balance         NUMERIC(18,2) NOT NULL DEFAULT 0,
    payment_terms   INT          NOT NULL DEFAULT 30,
    account_id      UUID         REFERENCES chart_of_accounts(id),
    salesperson_id  UUID         REFERENCES employees(id),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE INDEX idx_customers_company ON customers(company_id, is_active);

CREATE TABLE leads (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    title           VARCHAR(200) NOT NULL,
    first_name      VARCHAR(100),
    last_name       VARCHAR(100),
    company_name    VARCHAR(200),
    email           VARCHAR(200),
    phone           VARCHAR(30),
    source          VARCHAR(50),
    status          VARCHAR(20)  NOT NULL DEFAULT 'new',
    salesperson_id  UUID         REFERENCES employees(id),
    notes           TEXT,
    converted_at    TIMESTAMPTZ,
    converted_to    UUID         REFERENCES customers(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE opportunities (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    customer_id     UUID         REFERENCES customers(id),
    lead_id         UUID         REFERENCES leads(id),
    name            VARCHAR(300) NOT NULL,
    stage           opportunity_stage NOT NULL DEFAULT 'lead',
    amount          NUMERIC(18,2) NOT NULL DEFAULT 0,
    probability     INT          NOT NULL DEFAULT 10,
    expected_close  DATE,
    salesperson_id  UUID         REFERENCES employees(id),
    notes           TEXT,
    lost_reason     TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE quotations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id       UUID         REFERENCES branches(id),
    number          VARCHAR(50)  NOT NULL,
    customer_id     UUID         NOT NULL REFERENCES customers(id),
    date            DATE         NOT NULL,
    valid_until     DATE,
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft',
    subtotal        NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    tva_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
    stamp_tax       NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    notes           TEXT,
    terms           TEXT,
    salesperson_id  UUID         REFERENCES employees(id),
    converted_to    UUID,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE TABLE sales_orders (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id       UUID         REFERENCES branches(id),
    number          VARCHAR(50)  NOT NULL,
    quotation_id    UUID         REFERENCES quotations(id),
    customer_id     UUID         NOT NULL REFERENCES customers(id),
    date            DATE         NOT NULL,
    delivery_date   DATE,
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft',
    subtotal        NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    tva_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
    stamp_tax       NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    notes           TEXT,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE TABLE sales_invoices (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id       UUID         REFERENCES branches(id),
    number          VARCHAR(50)  NOT NULL,
    order_id        UUID         REFERENCES sales_orders(id),
    customer_id     UUID         NOT NULL REFERENCES customers(id),
    date            DATE         NOT NULL,
    due_date        DATE,
    status          invoice_status NOT NULL DEFAULT 'draft',
    subtotal        NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    tva_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
    stamp_tax       NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    paid_amount     NUMERIC(18,2) NOT NULL DEFAULT 0,
    balance_due     NUMERIC(18,2) GENERATED ALWAYS AS (total_amount - paid_amount) STORED,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    notes           TEXT,
    journal_entry_id UUID        REFERENCES journal_entries(id),
    confirmed_at    TIMESTAMPTZ,
    confirmed_by    UUID         REFERENCES users(id),
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE INDEX idx_si_company_status ON sales_invoices(company_id, status);
CREATE INDEX idx_si_customer       ON sales_invoices(customer_id);
CREATE INDEX idx_si_due_date       ON sales_invoices(due_date);

CREATE TABLE sales_invoice_lines (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id      UUID         NOT NULL REFERENCES sales_invoices(id) ON DELETE CASCADE,
    item_id         UUID,
    description     VARCHAR(500) NOT NULL,
    quantity        NUMERIC(18,4) NOT NULL DEFAULT 1,
    unit_price      NUMERIC(18,4) NOT NULL DEFAULT 0,
    discount_pct    NUMERIC(5,2)  NOT NULL DEFAULT 0,
    tva_rate        NUMERIC(5,2)  NOT NULL DEFAULT 19,
    subtotal        NUMERIC(18,2) GENERATED ALWAYS AS (ROUND(quantity * unit_price * (1 - discount_pct/100), 2)) STORED,
    tva_amount      NUMERIC(18,2),
    total           NUMERIC(18,2),
    account_id      UUID         REFERENCES chart_of_accounts(id),
    sort_order      INT          NOT NULL DEFAULT 0
);

-- Same structure for quotation_lines and sales_order_lines
CREATE TABLE quotation_lines (LIKE sales_invoice_lines INCLUDING ALL EXCLUDING CONSTRAINTS);
ALTER TABLE quotation_lines ADD COLUMN quotation_id UUID REFERENCES quotations(id) ON DELETE CASCADE;
ALTER TABLE quotation_lines DROP COLUMN IF EXISTS invoice_id;

CREATE TABLE sales_order_lines (LIKE sales_invoice_lines INCLUDING ALL EXCLUDING CONSTRAINTS);
ALTER TABLE sales_order_lines ADD COLUMN order_id UUID REFERENCES sales_orders(id) ON DELETE CASCADE;
ALTER TABLE sales_order_lines DROP COLUMN IF EXISTS invoice_id;

-- =============================================================================
-- PURCHASE
-- =============================================================================

CREATE TABLE suppliers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code            VARCHAR(20)  NOT NULL,
    name            VARCHAR(300) NOT NULL,
    type            VARCHAR(20)  NOT NULL DEFAULT 'company',
    nif             VARCHAR(20),
    nis             VARCHAR(20),
    rc              VARCHAR(30),
    art             VARCHAR(30),
    address         TEXT,
    city            VARCHAR(100),
    wilaya          VARCHAR(100),
    phone           VARCHAR(30),
    email           VARCHAR(200),
    payment_terms   INT          NOT NULL DEFAULT 30,
    credit_limit    NUMERIC(18,2) NOT NULL DEFAULT 0,
    balance         NUMERIC(18,2) NOT NULL DEFAULT 0,
    account_id      UUID         REFERENCES chart_of_accounts(id),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE rfqs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    number          VARCHAR(50)  NOT NULL,
    supplier_id     UUID         REFERENCES suppliers(id),
    date            DATE         NOT NULL,
    deadline        DATE,
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft',
    notes           TEXT,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE TABLE purchase_orders (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id       UUID         REFERENCES branches(id),
    number          VARCHAR(50)  NOT NULL,
    rfq_id          UUID         REFERENCES rfqs(id),
    supplier_id     UUID         NOT NULL REFERENCES suppliers(id),
    date            DATE         NOT NULL,
    expected_date   DATE,
    status          po_status    NOT NULL DEFAULT 'draft',
    subtotal        NUMERIC(18,2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    tva_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    received_amount NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    notes           TEXT,
    approved_by     UUID         REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    journal_entry_id UUID        REFERENCES journal_entries(id),
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE INDEX idx_po_company_status ON purchase_orders(company_id, status);

CREATE TABLE purchase_order_lines (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    po_id           UUID         NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    item_id         UUID,
    description     VARCHAR(500) NOT NULL,
    quantity        NUMERIC(18,4) NOT NULL DEFAULT 1,
    received_qty    NUMERIC(18,4) NOT NULL DEFAULT 0,
    unit_price      NUMERIC(18,4) NOT NULL DEFAULT 0,
    discount_pct    NUMERIC(5,2)  NOT NULL DEFAULT 0,
    tva_rate        NUMERIC(5,2)  NOT NULL DEFAULT 19,
    subtotal        NUMERIC(18,2),
    tva_amount      NUMERIC(18,2),
    total           NUMERIC(18,2),
    account_id      UUID         REFERENCES chart_of_accounts(id),
    sort_order      INT          NOT NULL DEFAULT 0
);

CREATE TABLE goods_receipts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    number          VARCHAR(50)  NOT NULL,
    po_id           UUID         NOT NULL REFERENCES purchase_orders(id),
    supplier_id     UUID         NOT NULL REFERENCES suppliers(id),
    date            DATE         NOT NULL,
    warehouse_id    UUID,
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft',
    notes           TEXT,
    validated_by    UUID         REFERENCES users(id),
    validated_at    TIMESTAMPTZ,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE TABLE goods_receipt_lines (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    grn_id          UUID         NOT NULL REFERENCES goods_receipts(id) ON DELETE CASCADE,
    po_line_id      UUID         REFERENCES purchase_order_lines(id),
    item_id         UUID,
    description     VARCHAR(500) NOT NULL,
    expected_qty    NUMERIC(18,4) NOT NULL DEFAULT 0,
    received_qty    NUMERIC(18,4) NOT NULL DEFAULT 0,
    unit_cost       NUMERIC(18,4) NOT NULL DEFAULT 0
);

CREATE TABLE purchase_invoices (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    number          VARCHAR(50)  NOT NULL,
    supplier_ref    VARCHAR(100),
    grn_id          UUID         REFERENCES goods_receipts(id),
    po_id           UUID         REFERENCES purchase_orders(id),
    supplier_id     UUID         NOT NULL REFERENCES suppliers(id),
    date            DATE         NOT NULL,
    due_date        DATE,
    status          invoice_status NOT NULL DEFAULT 'draft',
    subtotal        NUMERIC(18,2) NOT NULL DEFAULT 0,
    tva_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    paid_amount     NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    match_status    VARCHAR(20)  NOT NULL DEFAULT 'unmatched', -- 2way/3way/unmatched
    notes           TEXT,
    journal_entry_id UUID        REFERENCES journal_entries(id),
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE TABLE supplier_evaluations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    supplier_id     UUID         NOT NULL REFERENCES suppliers(id),
    po_id           UUID         REFERENCES purchase_orders(id),
    date            DATE         NOT NULL DEFAULT CURRENT_DATE,
    quality_score   INT          NOT NULL DEFAULT 3, -- 1-5
    delivery_score  INT          NOT NULL DEFAULT 3,
    price_score     INT          NOT NULL DEFAULT 3,
    service_score   INT          NOT NULL DEFAULT 3,
    overall_score   NUMERIC(3,1),
    notes           TEXT,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- INVENTORY
-- =============================================================================

CREATE TABLE item_categories (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID         NOT NULL REFERENCES companies(id),
    code        VARCHAR(20)  NOT NULL,
    name        VARCHAR(200) NOT NULL,
    parent_id   UUID         REFERENCES item_categories(id),
    account_id  UUID         REFERENCES chart_of_accounts(id),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE units_of_measure (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID         NOT NULL REFERENCES companies(id),
    code        VARCHAR(20)  NOT NULL,
    name        VARCHAR(100) NOT NULL,
    category    VARCHAR(50),
    factor      NUMERIC(18,6) NOT NULL DEFAULT 1,
    UNIQUE(company_id, code)
);

CREATE TABLE items (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id          UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code                VARCHAR(50)  NOT NULL,
    name                VARCHAR(300) NOT NULL,
    description         TEXT,
    category_id         UUID         REFERENCES item_categories(id),
    uom_id              UUID         REFERENCES units_of_measure(id),
    item_type           VARCHAR(20)  NOT NULL DEFAULT 'storable', -- storable/consumable/service
    track_inventory     BOOLEAN      NOT NULL DEFAULT TRUE,
    tva_rate            NUMERIC(5,2) NOT NULL DEFAULT 19,
    cost_method         VARCHAR(10)  NOT NULL DEFAULT 'cmup', -- cmup/fifo/lifo
    standard_cost       NUMERIC(18,4) NOT NULL DEFAULT 0,
    cmup_cost           NUMERIC(18,4) NOT NULL DEFAULT 0,
    sale_price          NUMERIC(18,4) NOT NULL DEFAULT 0,
    min_stock_qty       NUMERIC(18,4) NOT NULL DEFAULT 0,
    reorder_qty         NUMERIC(18,4) NOT NULL DEFAULT 0,
    max_stock_qty       NUMERIC(18,4) NOT NULL DEFAULT 0,
    barcode             VARCHAR(100),
    internal_ref        VARCHAR(100),
    hs_code             VARCHAR(20),
    inventory_account_id UUID        REFERENCES chart_of_accounts(id),
    cogs_account_id     UUID         REFERENCES chart_of_accounts(id),
    revenue_account_id  UUID         REFERENCES chart_of_accounts(id),
    is_active           BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE INDEX idx_items_company ON items(company_id, is_active);
CREATE INDEX idx_items_code    ON items(company_id, code);

CREATE TABLE warehouses (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id  UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id   UUID         REFERENCES branches(id),
    code        VARCHAR(20)  NOT NULL,
    name        VARCHAR(200) NOT NULL,
    address     TEXT,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE warehouse_locations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    warehouse_id    UUID         NOT NULL REFERENCES warehouses(id) ON DELETE CASCADE,
    code            VARCHAR(30)  NOT NULL,
    name            VARCHAR(100) NOT NULL,
    parent_id       UUID         REFERENCES warehouse_locations(id),
    UNIQUE(warehouse_id, code)
);

CREATE TABLE stock_levels (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    item_id         UUID         NOT NULL REFERENCES items(id),
    warehouse_id    UUID         NOT NULL REFERENCES warehouses(id),
    location_id     UUID         REFERENCES warehouse_locations(id),
    qty_on_hand     NUMERIC(18,4) NOT NULL DEFAULT 0,
    qty_reserved    NUMERIC(18,4) NOT NULL DEFAULT 0,
    qty_available   NUMERIC(18,4) GENERATED ALWAYS AS (qty_on_hand - qty_reserved) STORED,
    cmup_cost       NUMERIC(18,4) NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(item_id, warehouse_id, COALESCE(location_id, '00000000-0000-0000-0000-000000000000'::UUID))
);

CREATE INDEX idx_stock_levels_item      ON stock_levels(item_id);
CREATE INDEX idx_stock_levels_warehouse ON stock_levels(warehouse_id);
CREATE INDEX idx_stock_levels_company   ON stock_levels(company_id);

CREATE TABLE stock_movements (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    number          VARCHAR(50)  NOT NULL,
    date            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    type            movement_type NOT NULL,
    item_id         UUID         NOT NULL REFERENCES items(id),
    warehouse_id    UUID         NOT NULL REFERENCES warehouses(id),
    from_location_id UUID        REFERENCES warehouse_locations(id),
    to_warehouse_id UUID         REFERENCES warehouses(id),
    to_location_id  UUID         REFERENCES warehouse_locations(id),
    quantity        NUMERIC(18,4) NOT NULL,
    unit_cost       NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_cost      NUMERIC(18,2) GENERATED ALWAYS AS (ROUND(quantity * unit_cost, 2)) STORED,
    reference       VARCHAR(100),
    source_type     VARCHAR(50),
    source_id       UUID,
    notes           TEXT,
    journal_entry_id UUID        REFERENCES journal_entries(id),
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE INDEX idx_sm_company_date ON stock_movements(company_id, date DESC);
CREATE INDEX idx_sm_item         ON stock_movements(item_id);
CREATE INDEX idx_sm_source       ON stock_movements(source_type, source_id);

CREATE TABLE inventory_counts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    number          VARCHAR(50)  NOT NULL,
    date            DATE         NOT NULL,
    warehouse_id    UUID         NOT NULL REFERENCES warehouses(id),
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft',
    notes           TEXT,
    validated_by    UUID         REFERENCES users(id),
    validated_at    TIMESTAMPTZ,
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE TABLE inventory_count_lines (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    count_id        UUID         NOT NULL REFERENCES inventory_counts(id) ON DELETE CASCADE,
    item_id         UUID         NOT NULL REFERENCES items(id),
    location_id     UUID         REFERENCES warehouse_locations(id),
    book_qty        NUMERIC(18,4) NOT NULL DEFAULT 0,
    counted_qty     NUMERIC(18,4),
    difference      NUMERIC(18,4) GENERATED ALWAYS AS (COALESCE(counted_qty,0) - book_qty) STORED,
    unit_cost       NUMERIC(18,4) NOT NULL DEFAULT 0
);

-- =============================================================================
-- MANUFACTURING
-- =============================================================================

CREATE TABLE work_centers (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    code            VARCHAR(20)  NOT NULL,
    name            VARCHAR(200) NOT NULL,
    capacity        NUMERIC(8,2) NOT NULL DEFAULT 8,
    cost_per_hour   NUMERIC(18,2) NOT NULL DEFAULT 0,
    account_id      UUID         REFERENCES chart_of_accounts(id),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE bill_of_materials (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code            VARCHAR(30)  NOT NULL,
    product_id      UUID         NOT NULL REFERENCES items(id),
    version         VARCHAR(20)  NOT NULL DEFAULT '1.0',
    quantity        NUMERIC(18,4) NOT NULL DEFAULT 1,
    uom_id          UUID         REFERENCES units_of_measure(id),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE bom_components (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bom_id          UUID         NOT NULL REFERENCES bill_of_materials(id) ON DELETE CASCADE,
    component_id    UUID         NOT NULL REFERENCES items(id),
    quantity        NUMERIC(18,4) NOT NULL DEFAULT 1,
    uom_id          UUID         REFERENCES units_of_measure(id),
    scrap_pct       NUMERIC(5,2) NOT NULL DEFAULT 0,
    sort_order      INT          NOT NULL DEFAULT 0
);

CREATE TABLE bom_operations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bom_id          UUID         NOT NULL REFERENCES bill_of_materials(id) ON DELETE CASCADE,
    work_center_id  UUID         NOT NULL REFERENCES work_centers(id),
    name            VARCHAR(200) NOT NULL,
    duration_hours  NUMERIC(8,2) NOT NULL DEFAULT 1,
    sort_order      INT          NOT NULL DEFAULT 0
);

CREATE TABLE manufacturing_orders (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id          UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    number              VARCHAR(50)  NOT NULL,
    bom_id              UUID         NOT NULL REFERENCES bill_of_materials(id),
    product_id          UUID         NOT NULL REFERENCES items(id),
    warehouse_id        UUID         REFERENCES warehouses(id),
    planned_qty         NUMERIC(18,4) NOT NULL DEFAULT 1,
    produced_qty        NUMERIC(18,4) NOT NULL DEFAULT 0,
    status              mo_status    NOT NULL DEFAULT 'draft',
    planned_start       DATE,
    planned_end         DATE,
    actual_start        TIMESTAMPTZ,
    actual_end          TIMESTAMPTZ,
    material_cost       NUMERIC(18,2) NOT NULL DEFAULT 0,
    labor_cost          NUMERIC(18,2) NOT NULL DEFAULT 0,
    overhead_cost       NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_cost          NUMERIC(18,2) GENERATED ALWAYS AS (material_cost + labor_cost + overhead_cost) STORED,
    notes               TEXT,
    journal_entry_id    UUID         REFERENCES journal_entries(id),
    created_by          UUID         REFERENCES users(id),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE TABLE mo_component_lines (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    mo_id           UUID         NOT NULL REFERENCES manufacturing_orders(id) ON DELETE CASCADE,
    component_id    UUID         NOT NULL REFERENCES items(id),
    required_qty    NUMERIC(18,4) NOT NULL DEFAULT 0,
    consumed_qty    NUMERIC(18,4) NOT NULL DEFAULT 0,
    unit_cost       NUMERIC(18,4) NOT NULL DEFAULT 0
);

-- =============================================================================
-- PROJECTS & TIMESHEETS
-- =============================================================================

CREATE TABLE projects (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code            VARCHAR(30)  NOT NULL,
    name            VARCHAR(300) NOT NULL,
    customer_id     UUID         REFERENCES customers(id),
    manager_id      UUID         REFERENCES employees(id),
    start_date      DATE,
    end_date        DATE,
    status          project_status NOT NULL DEFAULT 'planning',
    budget          NUMERIC(18,2) NOT NULL DEFAULT 0,
    actual_cost     NUMERIC(18,2) NOT NULL DEFAULT 0,
    progress_pct    INT          NOT NULL DEFAULT 0,
    description     TEXT,
    account_id      UUID         REFERENCES chart_of_accounts(id),
    cost_center_id  UUID         REFERENCES cost_centers(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE project_tasks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID         NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    parent_id       UUID         REFERENCES project_tasks(id),
    title           VARCHAR(300) NOT NULL,
    description     TEXT,
    assignee_id     UUID         REFERENCES employees(id),
    status          task_status  NOT NULL DEFAULT 'todo',
    priority        task_priority NOT NULL DEFAULT 'medium',
    estimated_hours NUMERIC(8,2) NOT NULL DEFAULT 0,
    actual_hours    NUMERIC(8,2) NOT NULL DEFAULT 0,
    due_date        DATE,
    completed_at    TIMESTAMPTZ,
    sort_order      INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE timesheets (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    employee_id     UUID         NOT NULL REFERENCES employees(id),
    project_id      UUID         NOT NULL REFERENCES projects(id),
    task_id         UUID         REFERENCES project_tasks(id),
    date            DATE         NOT NULL,
    hours           NUMERIC(5,2) NOT NULL,
    description     TEXT,
    billable        BOOLEAN      NOT NULL DEFAULT TRUE,
    billed          BOOLEAN      NOT NULL DEFAULT FALSE,
    approved        BOOLEAN      NOT NULL DEFAULT FALSE,
    approved_by     UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- TREASURY
-- =============================================================================

CREATE TABLE cash_accounts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id       UUID         REFERENCES branches(id),
    code            VARCHAR(20)  NOT NULL,
    name            VARCHAR(200) NOT NULL,
    account_id      UUID         REFERENCES chart_of_accounts(id),
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    balance         NUMERIC(18,2) NOT NULL DEFAULT 0,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE bank_accounts (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    branch_id       UUID         REFERENCES branches(id),
    code            VARCHAR(20)  NOT NULL,
    name            VARCHAR(200) NOT NULL,
    bank_name       VARCHAR(200) NOT NULL,
    account_number  VARCHAR(50)  NOT NULL,
    rib             VARCHAR(30),
    iban            VARCHAR(50),
    swift           VARCHAR(20),
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    balance         NUMERIC(18,2) NOT NULL DEFAULT 0,
    account_id      UUID         REFERENCES chart_of_accounts(id),
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE TABLE bank_reconciliations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bank_account_id UUID         NOT NULL REFERENCES bank_accounts(id),
    period_date     DATE         NOT NULL,
    statement_balance NUMERIC(18,2) NOT NULL DEFAULT 0,
    book_balance    NUMERIC(18,2) NOT NULL DEFAULT 0,
    difference      NUMERIC(18,2) GENERATED ALWAYS AS (statement_balance - book_balance) STORED,
    is_reconciled   BOOLEAN      NOT NULL DEFAULT FALSE,
    reconciled_at   TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE cheques (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    number          VARCHAR(50)  NOT NULL,
    bank_account_id UUID         NOT NULL REFERENCES bank_accounts(id),
    cheque_type     VARCHAR(20)  NOT NULL DEFAULT 'received', -- received/issued
    amount          NUMERIC(18,2) NOT NULL,
    issue_date      DATE         NOT NULL,
    due_date        DATE,
    payee_payer     VARCHAR(300),
    status          cheque_status NOT NULL DEFAULT 'pending',
    deposited_at    TIMESTAMPTZ,
    cleared_at      TIMESTAMPTZ,
    bounced_at      TIMESTAMPTZ,
    bounce_reason   TEXT,
    journal_entry_id UUID        REFERENCES journal_entries(id),
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    number          VARCHAR(50)  NOT NULL,
    payment_type    payment_type NOT NULL DEFAULT 'bank_transfer',
    direction       VARCHAR(20)  NOT NULL DEFAULT 'outgoing', -- incoming/outgoing
    date            DATE         NOT NULL,
    amount          NUMERIC(18,2) NOT NULL,
    currency        VARCHAR(10)  NOT NULL DEFAULT 'DZD',
    cash_account_id UUID         REFERENCES cash_accounts(id),
    bank_account_id UUID         REFERENCES bank_accounts(id),
    cheque_id       UUID         REFERENCES cheques(id),
    partner_id      UUID,
    partner_type    VARCHAR(20),  -- customer/supplier
    invoice_id      UUID,
    invoice_type    VARCHAR(20),  -- sales/purchase
    reference       VARCHAR(100),
    notes           TEXT,
    journal_entry_id UUID        REFERENCES journal_entries(id),
    created_by      UUID         REFERENCES users(id),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, number)
);

-- =============================================================================
-- TAX DECLARATIONS (G50/G29)
-- =============================================================================

CREATE TABLE tax_declarations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    declaration_type VARCHAR(10) NOT NULL, -- G50/G29/IBS
    period_month    INT          NOT NULL,
    period_year     INT          NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'draft',
    tva_collected   NUMERIC(18,2) NOT NULL DEFAULT 0,
    tva_deductible  NUMERIC(18,2) NOT NULL DEFAULT 0,
    tva_net         NUMERIC(18,2) GENERATED ALWAYS AS (tva_collected - tva_deductible) STORED,
    irs_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
    tap_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,  -- Taxe sur l'Activité Professionnelle
    other_taxes     NUMERIC(18,2) NOT NULL DEFAULT 0,
    total_payable   NUMERIC(18,2),
    filing_date     DATE,
    payment_date    DATE,
    reference       VARCHAR(100),
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE(company_id, declaration_type, period_year, period_month)
);

CREATE TABLE vat_entries (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    date            DATE         NOT NULL,
    entry_type      VARCHAR(20)  NOT NULL, -- collected/deductible
    partner_name    VARCHAR(300),
    partner_nif     VARCHAR(20),
    invoice_ref     VARCHAR(100),
    tva_rate        NUMERIC(5,2) NOT NULL DEFAULT 19,
    taxable_amount  NUMERIC(18,2) NOT NULL DEFAULT 0,
    tva_amount      NUMERIC(18,2) NOT NULL DEFAULT 0,
    source_type     VARCHAR(50),
    source_id       UUID,
    period_month    INT,
    period_year     INT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- WORKFLOW & APPROVALS
-- =============================================================================

CREATE TABLE workflow_rules (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    name            VARCHAR(200) NOT NULL,
    entity_type     VARCHAR(50)  NOT NULL, -- purchase_order/sales_invoice/leave_request/payroll
    condition_field VARCHAR(100),
    condition_op    VARCHAR(20)  NOT NULL DEFAULT 'gt',
    condition_value NUMERIC(18,2),
    approver_id     UUID         REFERENCES users(id),
    approver_role   user_role,
    sequence        INT          NOT NULL DEFAULT 1,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE approval_requests (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id      UUID         NOT NULL REFERENCES companies(id),
    entity_type     VARCHAR(50)  NOT NULL,
    entity_id       UUID         NOT NULL,
    entity_number   VARCHAR(50),
    entity_amount   NUMERIC(18,2),
    workflow_rule_id UUID        REFERENCES workflow_rules(id),
    status          approval_status NOT NULL DEFAULT 'pending',
    requested_by    UUID         NOT NULL REFERENCES users(id),
    approver_id     UUID         REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    rejection_reason TEXT,
    deadline        TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_approval_company_status ON approval_requests(company_id, status);
CREATE INDEX idx_approval_approver       ON approval_requests(approver_id, status);

-- =============================================================================
-- NOTIFICATIONS
-- =============================================================================

CREATE TABLE notifications (
    id          BIGSERIAL    PRIMARY KEY,
    user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        VARCHAR(50)  NOT NULL,
    title       VARCHAR(300) NOT NULL,
    message     TEXT,
    link        TEXT,
    is_read     BOOLEAN      NOT NULL DEFAULT FALSE,
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);

-- =============================================================================
-- MIGRATION TRACKING
-- =============================================================================

CREATE TABLE IF NOT EXISTS schema_migrations (
    version     VARCHAR(50) PRIMARY KEY,
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO schema_migrations (version) VALUES ('0001_init_schema') ON CONFLICT DO NOTHING;

-- =============================================================================
-- DEFAULT DATA — Super Tenant + Admin User
-- =============================================================================

INSERT INTO tenants (id, code, name, plan) VALUES
    ('00000000-0000-0000-0000-000000000001', 'NEXUS', 'Nexus Default Tenant', 'enterprise')
ON CONFLICT DO NOTHING;

INSERT INTO companies (id, tenant_id, code, name, legal_name, currency) VALUES
    ('00000000-0000-0000-0000-000000000002',
     '00000000-0000-0000-0000-000000000001',
     'COMP01', 'My Company', 'My Company SARL', 'DZD')
ON CONFLICT DO NOTHING;

-- Default admin user: admin / Admin@123456
-- Password hash for "Admin@123456" using bcrypt cost=12
INSERT INTO users (id, tenant_id, company_id, username, email, password_hash, full_name, role) VALUES
    ('00000000-0000-0000-0000-000000000003',
     '00000000-0000-0000-0000-000000000001',
     '00000000-0000-0000-0000-000000000002',
     'admin',
     'admin@nexus-erp.local',
     '$2b$12$7ohOxG.AkzuDrYs.YjkAZeT782oQJPqF8767Ay3/laNUzBUDbeFhC',
     'System Administrator',
     'super_admin')
ON CONFLICT DO NOTHING;

-- Default SCF Chart of Accounts Class 1-7
INSERT INTO chart_of_accounts (id, company_id, code, name, type, nature, is_group, level) VALUES
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '1', 'CAPITAUX', 'equity', 'credit', TRUE, 1),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '10', 'Capital et réserves', 'equity', 'credit', TRUE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '101', 'Capital social', 'equity', 'credit', FALSE, 3),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '106', 'Réserves', 'equity', 'credit', FALSE, 3),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '12', 'Résultat de l''exercice', 'equity', 'credit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '2', 'IMMOBILISATIONS', 'asset', 'debit', TRUE, 1),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '21', 'Immobilisations corporelles', 'asset', 'debit', TRUE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '213', 'Constructions', 'asset', 'debit', FALSE, 3),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '215', 'Installations techniques', 'asset', 'debit', FALSE, 3),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '218', 'Autres immobilisations corporelles', 'asset', 'debit', FALSE, 3),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '22', 'Immobilisations incorporelles', 'asset', 'debit', TRUE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '3', 'STOCKS', 'asset', 'debit', TRUE, 1),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '30', 'Stocks de marchandises', 'asset', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '31', 'Matières premières', 'asset', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '35', 'Produits finis', 'asset', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '4', 'COMPTES DE TIERS', 'asset', 'debit', TRUE, 1),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '40', 'Fournisseurs', 'liability', 'credit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '41', 'Clients', 'asset', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '42', 'Personnel', 'liability', 'credit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '43', 'Organismes sociaux (CNAS)', 'liability', 'credit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '44', 'Etat et collectivités', 'liability', 'credit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '4451', 'TVA collectée', 'liability', 'credit', FALSE, 3),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '4456', 'TVA déductible', 'asset', 'debit', FALSE, 3),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '5', 'COMPTES FINANCIERS', 'asset', 'debit', TRUE, 1),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '51', 'Banques', 'asset', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '53', 'Caisse', 'asset', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '6', 'CHARGES', 'expense', 'debit', TRUE, 1),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '60', 'Achats consommés', 'expense', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '61', 'Services extérieurs', 'expense', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '63', 'Charges de personnel', 'expense', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '64', 'Impôts et taxes', 'expense', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '65', 'Autres charges opérationnelles', 'expense', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '68', 'Dotations amortissements', 'expense', 'debit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '7', 'PRODUITS', 'revenue', 'credit', TRUE, 1),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '70', 'Ventes de produits et marchandises', 'revenue', 'credit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '74', 'Autres produits opérationnels', 'revenue', 'credit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '75', 'Produits financiers', 'revenue', 'credit', FALSE, 2),
    (uuid_generate_v4(), '00000000-0000-0000-0000-000000000002', '76', 'Produits hors activité ordinaire', 'revenue', 'credit', FALSE, 2)
ON CONFLICT (company_id, code) DO NOTHING;

-- Default warehouse
INSERT INTO warehouses (id, company_id, code, name) VALUES
    ('00000000-0000-0000-0000-000000000010',
     '00000000-0000-0000-0000-000000000002',
     'WH01', 'Main Warehouse')
ON CONFLICT DO NOTHING;

-- Default fiscal year
INSERT INTO fiscal_years (id, company_id, name, start_date, end_date) VALUES
    ('00000000-0000-0000-0000-000000000011',
     '00000000-0000-0000-0000-000000000002',
     '2025', '2025-01-01', '2025-12-31')
ON CONFLICT DO NOTHING;

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

CREATE INDEX IF NOT EXISTS idx_je_company_number      ON journal_entries(company_id, number);
CREATE INDEX IF NOT EXISTS idx_si_number              ON sales_invoices(company_id, number);
CREATE INDEX IF NOT EXISTS idx_po_number              ON purchase_orders(company_id, number);
CREATE INDEX IF NOT EXISTS idx_employees_number       ON employees(company_id, employee_number);
CREATE INDEX IF NOT EXISTS idx_customers_name_trgm    ON customers USING gin(name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_suppliers_name_trgm    ON suppliers USING gin(name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_items_name_trgm        ON items USING gin(name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_sm_warehouse           ON stock_movements(warehouse_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_payslips_run           ON payslips(payroll_run_id);
CREATE INDEX IF NOT EXISTS idx_payslips_employee      ON payslips(employee_id);
CREATE INDEX IF NOT EXISTS idx_timesheets_project     ON timesheets(project_id, date);
CREATE INDEX IF NOT EXISTS idx_vat_entries_period     ON vat_entries(company_id, period_year, period_month);
