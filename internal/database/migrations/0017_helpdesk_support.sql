-- ============================================================
-- Migration 0017: Helpdesk / Support Module
-- Idempotent: safe to run multiple times
-- ============================================================

-- ── ENUMs ────────────────────────────────────────────────────────────────────

DO $$ BEGIN
  CREATE TYPE ticket_status AS ENUM ('open','pending','in_progress','resolved','closed','cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE ticket_priority AS ENUM ('low','medium','high','critical');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE ticket_source AS ENUM ('email','phone','chat','portal','api','internal');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE sla_priority AS ENUM ('low','medium','high','critical');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE escalation_status AS ENUM ('pending','active','resolved','closed');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE csat_rating AS ENUM ('very_dissatisfied','dissatisfied','neutral','satisfied','very_satisfied');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TYPE agent_status AS ENUM ('active','inactive','on_leave','busy');
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

-- ── Tables ───────────────────────────────────────────────────────────────────

-- Ticket Categories
CREATE TABLE IF NOT EXISTS helpdesk_categories (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id    UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  description   TEXT NOT NULL DEFAULT '',
  parent_id     UUID REFERENCES helpdesk_categories(id) ON DELETE SET NULL,
  color         TEXT NOT NULL DEFAULT '#6366f1',
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  sort_order    INTEGER NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_hd_categories_company ON helpdesk_categories(company_id);
CREATE INDEX IF NOT EXISTS idx_hd_categories_parent  ON helpdesk_categories(parent_id);

-- SLA Policies
CREATE TABLE IF NOT EXISTS helpdesk_sla_policies (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id            UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  name                  TEXT NOT NULL,
  description           TEXT NOT NULL DEFAULT '',
  priority              sla_priority NOT NULL DEFAULT 'medium',
  first_response_hours  INTEGER NOT NULL DEFAULT 4,
  resolution_hours      INTEGER NOT NULL DEFAULT 24,
  business_hours_only   BOOLEAN NOT NULL DEFAULT TRUE,
  is_active             BOOLEAN NOT NULL DEFAULT TRUE,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_hd_sla_company ON helpdesk_sla_policies(company_id);

-- Helpdesk Agents
CREATE TABLE IF NOT EXISTS helpdesk_agents (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id    UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
  name          TEXT NOT NULL,
  email         TEXT NOT NULL,
  phone         TEXT NOT NULL DEFAULT '',
  department    TEXT NOT NULL DEFAULT '',
  specialization TEXT NOT NULL DEFAULT '',
  status        agent_status NOT NULL DEFAULT 'active',
  max_tickets   INTEGER NOT NULL DEFAULT 20,
  is_active     BOOLEAN NOT NULL DEFAULT TRUE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_hd_agents_company ON helpdesk_agents(company_id);
CREATE INDEX IF NOT EXISTS idx_hd_agents_user    ON helpdesk_agents(user_id);

-- Tickets
CREATE TABLE IF NOT EXISTS helpdesk_tickets (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  ticket_number   TEXT NOT NULL,
  subject         TEXT NOT NULL,
  description     TEXT NOT NULL DEFAULT '',
  status          ticket_status NOT NULL DEFAULT 'open',
  priority        ticket_priority NOT NULL DEFAULT 'medium',
  source          ticket_source NOT NULL DEFAULT 'portal',
  category_id     UUID REFERENCES helpdesk_categories(id) ON DELETE SET NULL,
  sla_policy_id   UUID REFERENCES helpdesk_sla_policies(id) ON DELETE SET NULL,
  assigned_agent_id UUID REFERENCES helpdesk_agents(id) ON DELETE SET NULL,
  requester_name  TEXT NOT NULL DEFAULT '',
  requester_email TEXT NOT NULL DEFAULT '',
  requester_phone TEXT NOT NULL DEFAULT '',
  company_name    TEXT NOT NULL DEFAULT '',
  first_response_at TIMESTAMPTZ,
  resolved_at     TIMESTAMPTZ,
  closed_at       TIMESTAMPTZ,
  due_date        TIMESTAMPTZ,
  tags            TEXT[] NOT NULL DEFAULT '{}',
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_hd_tickets_company  ON helpdesk_tickets(company_id);
CREATE INDEX IF NOT EXISTS idx_hd_tickets_status   ON helpdesk_tickets(company_id, status);
CREATE INDEX IF NOT EXISTS idx_hd_tickets_priority ON helpdesk_tickets(company_id, priority);
CREATE INDEX IF NOT EXISTS idx_hd_tickets_agent    ON helpdesk_tickets(assigned_agent_id);
CREATE INDEX IF NOT EXISTS idx_hd_tickets_category ON helpdesk_tickets(category_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_hd_tickets_number ON helpdesk_tickets(company_id, ticket_number);

-- Ticket Comments
CREATE TABLE IF NOT EXISTS ticket_comments (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  ticket_id   UUID NOT NULL REFERENCES helpdesk_tickets(id) ON DELETE CASCADE,
  agent_id    UUID REFERENCES helpdesk_agents(id) ON DELETE SET NULL,
  author_name TEXT NOT NULL DEFAULT '',
  body        TEXT NOT NULL,
  is_internal BOOLEAN NOT NULL DEFAULT FALSE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_hd_comments_ticket  ON ticket_comments(ticket_id);
CREATE INDEX IF NOT EXISTS idx_hd_comments_company ON ticket_comments(company_id);

-- Ticket Assignments (assignment history)
CREATE TABLE IF NOT EXISTS ticket_assignments (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  ticket_id       UUID NOT NULL REFERENCES helpdesk_tickets(id) ON DELETE CASCADE,
  agent_id        UUID REFERENCES helpdesk_agents(id) ON DELETE SET NULL,
  assigned_by     TEXT NOT NULL DEFAULT '',
  reason          TEXT NOT NULL DEFAULT '',
  assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  unassigned_at   TIMESTAMPTZ,
  is_current      BOOLEAN NOT NULL DEFAULT TRUE
);
CREATE INDEX IF NOT EXISTS idx_hd_assignments_ticket  ON ticket_assignments(ticket_id);
CREATE INDEX IF NOT EXISTS idx_hd_assignments_agent   ON ticket_assignments(agent_id);
CREATE INDEX IF NOT EXISTS idx_hd_assignments_company ON ticket_assignments(company_id);

-- Ticket Escalations
CREATE TABLE IF NOT EXISTS ticket_escalations (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  ticket_id       UUID NOT NULL REFERENCES helpdesk_tickets(id) ON DELETE CASCADE,
  escalated_by    TEXT NOT NULL DEFAULT '',
  escalated_to    TEXT NOT NULL DEFAULT '',
  reason          TEXT NOT NULL DEFAULT '',
  status          escalation_status NOT NULL DEFAULT 'active',
  resolution_note TEXT NOT NULL DEFAULT '',
  escalated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  resolved_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_hd_escalations_ticket  ON ticket_escalations(ticket_id);
CREATE INDEX IF NOT EXISTS idx_hd_escalations_company ON ticket_escalations(company_id);
CREATE INDEX IF NOT EXISTS idx_hd_escalations_status  ON ticket_escalations(company_id, status);

-- CSAT Surveys
CREATE TABLE IF NOT EXISTS csat_surveys (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id      UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  ticket_id       UUID NOT NULL REFERENCES helpdesk_tickets(id) ON DELETE CASCADE,
  agent_id        UUID REFERENCES helpdesk_agents(id) ON DELETE SET NULL,
  rating          csat_rating NOT NULL,
  comment         TEXT NOT NULL DEFAULT '',
  requester_name  TEXT NOT NULL DEFAULT '',
  requester_email TEXT NOT NULL DEFAULT '',
  submitted_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_hd_csat_company ON csat_surveys(company_id);
CREATE INDEX IF NOT EXISTS idx_hd_csat_ticket  ON csat_surveys(ticket_id);
CREATE INDEX IF NOT EXISTS idx_hd_csat_agent   ON csat_surveys(agent_id);

-- Helpdesk Metrics (daily aggregated)
CREATE TABLE IF NOT EXISTS helpdesk_metrics (
  id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id              UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  metric_date             DATE NOT NULL,
  tickets_opened          INTEGER NOT NULL DEFAULT 0,
  tickets_resolved        INTEGER NOT NULL DEFAULT 0,
  tickets_closed          INTEGER NOT NULL DEFAULT 0,
  avg_first_response_mins NUMERIC(10,2) NOT NULL DEFAULT 0,
  avg_resolution_hours    NUMERIC(10,2) NOT NULL DEFAULT 0,
  sla_breached_count      INTEGER NOT NULL DEFAULT 0,
  csat_responses          INTEGER NOT NULL DEFAULT 0,
  csat_score_sum          NUMERIC(10,2) NOT NULL DEFAULT 0,
  created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_hd_metrics_company ON helpdesk_metrics(company_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_hd_metrics_date ON helpdesk_metrics(company_id, metric_date);

-- ── Ticket number sequence ───────────────────────────────────────────────────
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_sequences WHERE schemaname='public' AND sequencename='helpdesk_ticket_seq') THEN
    CREATE SEQUENCE helpdesk_ticket_seq START 1000 INCREMENT 1;
  END IF;
END$$;

-- ── Seed default SLA policies (per-company will be created by app) ────────────
-- No global seeds needed; each company creates their own via the UI.

-- ── updated_at trigger function (reuse if exists) ────────────────────────────
CREATE OR REPLACE FUNCTION update_helpdesk_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$;

DO $$ BEGIN
  CREATE TRIGGER trg_hd_categories_updated
    BEFORE UPDATE ON helpdesk_categories
    FOR EACH ROW EXECUTE FUNCTION update_helpdesk_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TRIGGER trg_hd_sla_updated
    BEFORE UPDATE ON helpdesk_sla_policies
    FOR EACH ROW EXECUTE FUNCTION update_helpdesk_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TRIGGER trg_hd_agents_updated
    BEFORE UPDATE ON helpdesk_agents
    FOR EACH ROW EXECUTE FUNCTION update_helpdesk_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TRIGGER trg_hd_tickets_updated
    BEFORE UPDATE ON helpdesk_tickets
    FOR EACH ROW EXECUTE FUNCTION update_helpdesk_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL; END$$;

DO $$ BEGIN
  CREATE TRIGGER trg_hd_escalations_updated
    BEFORE UPDATE ON ticket_escalations
    FOR EACH ROW EXECUTE FUNCTION update_helpdesk_updated_at();
EXCEPTION WHEN duplicate_object THEN NULL; END$$;
