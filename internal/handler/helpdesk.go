package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"nexus-erp/internal/middleware"
)

// ── Struct ────────────────────────────────────────────────────────────────────

type HelpdeskHandler struct{ db *pgxpool.Pool }

// ── Dashboard ─────────────────────────────────────────────────────────────────

func (h *HelpdeskHandler) GetDashboard(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	// KPI summary
	type KPI struct {
		TotalOpen        int     `json:"total_open"`
		TotalPending     int     `json:"total_pending"`
		TotalInProgress  int     `json:"total_in_progress"`
		TotalResolved    int     `json:"total_resolved"`
		TotalClosed      int     `json:"total_closed"`
		OpenedToday      int     `json:"opened_today"`
		ResolvedToday    int     `json:"resolved_today"`
		OverdueSLA       int     `json:"overdue_sla"`
		AvgResolutionHrs float64 `json:"avg_resolution_hrs"`
		CSATScore        float64 `json:"csat_score"`
		CSATResponses    int     `json:"csat_responses"`
		ActiveAgents     int     `json:"active_agents"`
	}

	var kpi KPI
	err := h.db.QueryRow(ctx, `
		SELECT
		  COALESCE(SUM(CASE WHEN status='open'        THEN 1 ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN status='pending'     THEN 1 ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN status='in_progress' THEN 1 ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN status='resolved'    THEN 1 ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN status='closed'      THEN 1 ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN DATE(created_at)=CURRENT_DATE THEN 1 ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN DATE(resolved_at)=CURRENT_DATE THEN 1 ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN due_date < NOW() AND status NOT IN ('resolved','closed','cancelled') THEN 1 ELSE 0 END),0),
		  COALESCE(AVG(CASE WHEN resolved_at IS NOT NULL
		                    THEN EXTRACT(EPOCH FROM (resolved_at - created_at))/3600
		               END),0),
		  0,0,0
		FROM helpdesk_tickets
		WHERE company_id=$1`, companyID).Scan(
		&kpi.TotalOpen, &kpi.TotalPending, &kpi.TotalInProgress,
		&kpi.TotalResolved, &kpi.TotalClosed,
		&kpi.OpenedToday, &kpi.ResolvedToday, &kpi.OverdueSLA,
		&kpi.AvgResolutionHrs,
		&kpi.CSATScore, &kpi.CSATResponses, &kpi.ActiveAgents,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// CSAT score
	_ = h.db.QueryRow(ctx, `
		SELECT
		  COALESCE(COUNT(*),0),
		  COALESCE(AVG(CASE rating
		    WHEN 'very_dissatisfied' THEN 1
		    WHEN 'dissatisfied'      THEN 2
		    WHEN 'neutral'           THEN 3
		    WHEN 'satisfied'         THEN 4
		    WHEN 'very_satisfied'    THEN 5
		  END),0)
		FROM csat_surveys WHERE company_id=$1`, companyID).Scan(&kpi.CSATResponses, &kpi.CSATScore)

	// Active agents
	_ = h.db.QueryRow(ctx, `SELECT COUNT(*) FROM helpdesk_agents WHERE company_id=$1 AND status='active' AND is_active=TRUE`, companyID).Scan(&kpi.ActiveAgents)

	// Tickets by priority (last 30 days)
	type PriorityCount struct {
		Priority string `json:"priority"`
		Count    int    `json:"count"`
	}
	priorityCounts := []PriorityCount{}
	rows, err := h.db.Query(ctx, `
		SELECT priority::TEXT, COUNT(*) FROM helpdesk_tickets
		WHERE company_id=$1 AND created_at >= NOW()-INTERVAL '30 days'
		GROUP BY priority ORDER BY COUNT(*) DESC`, companyID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p PriorityCount
			if rows.Scan(&p.Priority, &p.Count) == nil {
				priorityCounts = append(priorityCounts, p)
			}
		}
	}

	// Tickets by status
	type StatusCount struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}
	statusCounts := []StatusCount{}
	rows2, err2 := h.db.Query(ctx, `
		SELECT status::TEXT, COUNT(*) FROM helpdesk_tickets
		WHERE company_id=$1 GROUP BY status ORDER BY COUNT(*) DESC`, companyID)
	if err2 == nil {
		defer rows2.Close()
		for rows2.Next() {
			var s StatusCount
			if rows2.Scan(&s.Status, &s.Count) == nil {
				statusCounts = append(statusCounts, s)
			}
		}
	}

	// Daily tickets trend (last 14 days)
	type DailyTrend struct {
		Date     string `json:"date"`
		Opened   int    `json:"opened"`
		Resolved int    `json:"resolved"`
	}
	trend := []DailyTrend{}
	rows3, _ := h.db.Query(ctx, `
		WITH days AS (
		  SELECT generate_series(CURRENT_DATE-13, CURRENT_DATE, '1 day'::interval)::date AS d
		)
		SELECT
		  d::TEXT,
		  COALESCE(SUM(CASE WHEN DATE(t.created_at)=d THEN 1 ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN DATE(t.resolved_at)=d THEN 1 ELSE 0 END),0)
		FROM days
		LEFT JOIN helpdesk_tickets t ON t.company_id=$1
		GROUP BY d ORDER BY d`, companyID)
	if rows3 != nil {
		defer rows3.Close()
		for rows3.Next() {
			var dt DailyTrend
			if rows3.Scan(&dt.Date, &dt.Opened, &dt.Resolved) == nil {
				trend = append(trend, dt)
			}
		}
	}

	// Top agents by resolved tickets
	type AgentStat struct {
		Name     string  `json:"name"`
		Resolved int     `json:"resolved"`
		AvgHrs   float64 `json:"avg_resolution_hrs"`
	}
	agentStats := []AgentStat{}
	rows4, _ := h.db.Query(ctx, `
		SELECT a.name,
		  COUNT(CASE WHEN t.status IN ('resolved','closed') THEN 1 END) AS resolved,
		  COALESCE(AVG(CASE WHEN t.resolved_at IS NOT NULL
		    THEN EXTRACT(EPOCH FROM (t.resolved_at - t.created_at))/3600 END),0)
		FROM helpdesk_agents a
		LEFT JOIN helpdesk_tickets t ON t.assigned_agent_id=a.id AND t.company_id=$1
		WHERE a.company_id=$1 AND a.is_active=TRUE
		GROUP BY a.id, a.name
		ORDER BY resolved DESC LIMIT 10`, companyID)
	if rows4 != nil {
		defer rows4.Close()
		for rows4.Next() {
			var as_ AgentStat
			if rows4.Scan(&as_.Name, &as_.Resolved, &as_.AvgHrs) == nil {
				agentStats = append(agentStats, as_)
			}
		}
	}

	// Recent tickets
	type RecentTicket struct {
		ID             string `json:"id"`
		TicketNumber   string `json:"ticket_number"`
		Subject        string `json:"subject"`
		Status         string `json:"status"`
		Priority       string `json:"priority"`
		RequesterName  string `json:"requester_name"`
		AssignedAgent  string `json:"assigned_agent"`
		CreatedAt      string `json:"created_at"`
	}
	recentTickets := []RecentTicket{}
	rows5, _ := h.db.Query(ctx, `
		SELECT t.id, t.ticket_number, t.subject, t.status::TEXT, t.priority::TEXT,
		  COALESCE(t.requester_name,''), COALESCE(a.name,'Unassigned'),
		  t.created_at::TEXT
		FROM helpdesk_tickets t
		LEFT JOIN helpdesk_agents a ON a.id=t.assigned_agent_id
		WHERE t.company_id=$1
		ORDER BY t.created_at DESC LIMIT 10`, companyID)
	if rows5 != nil {
		defer rows5.Close()
		for rows5.Next() {
			var rt RecentTicket
			if rows5.Scan(&rt.ID, &rt.TicketNumber, &rt.Subject, &rt.Status, &rt.Priority,
				&rt.RequesterName, &rt.AssignedAgent, &rt.CreatedAt) == nil {
				recentTickets = append(recentTickets, rt)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"kpi":            kpi,
		"priority_counts": priorityCounts,
		"status_counts":  statusCounts,
		"trend":          trend,
		"agent_stats":    agentStats,
		"recent_tickets": recentTickets,
	})
}

// ── Tickets ───────────────────────────────────────────────────────────────────

func (h *HelpdeskHandler) ListTickets(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	where := []string{"t.company_id=$1"}
	args := []interface{}{companyID}
	n := 2

	if s := c.Query("status"); s != "" {
		where = append(where, fmt.Sprintf("t.status=$%d::ticket_status", n))
		args = append(args, s); n++
	}
	if p := c.Query("priority"); p != "" {
		where = append(where, fmt.Sprintf("t.priority=$%d::ticket_priority", n))
		args = append(args, p); n++
	}
	if cat := c.Query("category_id"); cat != "" {
		where = append(where, fmt.Sprintf("t.category_id=$%d", n))
		args = append(args, cat); n++
	}
	if agent := c.Query("agent_id"); agent != "" {
		where = append(where, fmt.Sprintf("t.assigned_agent_id=$%d", n))
		args = append(args, agent); n++
	}
	if search := c.Query("search"); search != "" {
		where = append(where, fmt.Sprintf("(t.subject ILIKE $%d OR t.ticket_number ILIKE $%d OR t.requester_name ILIKE $%d)", n, n, n))
		args = append(args, "%"+search+"%"); n++
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		where = append(where, fmt.Sprintf("t.created_at >= $%d", n))
		args = append(args, dateFrom); n++
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		where = append(where, fmt.Sprintf("t.created_at <= $%d", n))
		args = append(args, dateTo); n++
	}
	_ = n

	sql := `
		SELECT t.id, t.ticket_number, t.subject, t.status::TEXT, t.priority::TEXT,
		  t.source::TEXT, COALESCE(c.name,''), COALESCE(a.name,'Unassigned'),
		  COALESCE(t.requester_name,''), COALESCE(t.requester_email,''),
		  COALESCE(t.company_name,''), t.created_at::TEXT, t.updated_at::TEXT,
		  COALESCE(t.due_date::TEXT,''), COALESCE(t.resolved_at::TEXT,''),
		  COALESCE(sp.name,'')
		FROM helpdesk_tickets t
		LEFT JOIN helpdesk_categories c ON c.id=t.category_id
		LEFT JOIN helpdesk_agents a ON a.id=t.assigned_agent_id
		LEFT JOIN helpdesk_sla_policies sp ON sp.id=t.sla_policy_id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY t.created_at DESC`

	rows, err := h.db.Query(ctx, sql, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type Ticket struct {
		ID            string `json:"id"`
		TicketNumber  string `json:"ticket_number"`
		Subject       string `json:"subject"`
		Status        string `json:"status"`
		Priority      string `json:"priority"`
		Source        string `json:"source"`
		Category      string `json:"category"`
		AssignedAgent string `json:"assigned_agent"`
		RequesterName string `json:"requester_name"`
		RequesterEmail string `json:"requester_email"`
		CompanyName   string `json:"company_name"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
		DueDate       string `json:"due_date"`
		ResolvedAt    string `json:"resolved_at"`
		SLAPolicy     string `json:"sla_policy"`
	}

	tickets := []Ticket{}
	for rows.Next() {
		var t Ticket
		if err := rows.Scan(&t.ID, &t.TicketNumber, &t.Subject, &t.Status, &t.Priority,
			&t.Source, &t.Category, &t.AssignedAgent, &t.RequesterName, &t.RequesterEmail,
			&t.CompanyName, &t.CreatedAt, &t.UpdatedAt, &t.DueDate, &t.ResolvedAt, &t.SLAPolicy); err == nil {
			tickets = append(tickets, t)
		}
	}
	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}

func (h *HelpdeskHandler) GetTicket(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	type Ticket struct {
		ID             string   `json:"id"`
		TicketNumber   string   `json:"ticket_number"`
		Subject        string   `json:"subject"`
		Description    string   `json:"description"`
		Status         string   `json:"status"`
		Priority       string   `json:"priority"`
		Source         string   `json:"source"`
		CategoryID     string   `json:"category_id"`
		Category       string   `json:"category"`
		SLAPolicyID    string   `json:"sla_policy_id"`
		SLAPolicy      string   `json:"sla_policy"`
		AssignedAgentID string  `json:"assigned_agent_id"`
		AssignedAgent  string   `json:"assigned_agent"`
		RequesterName  string   `json:"requester_name"`
		RequesterEmail string   `json:"requester_email"`
		RequesterPhone string   `json:"requester_phone"`
		CompanyName    string   `json:"company_name"`
		DueDate        string   `json:"due_date"`
		FirstResponseAt string  `json:"first_response_at"`
		ResolvedAt     string   `json:"resolved_at"`
		ClosedAt       string   `json:"closed_at"`
		Tags           []string `json:"tags"`
		CreatedAt      string   `json:"created_at"`
		UpdatedAt      string   `json:"updated_at"`
	}

	var t Ticket
	err := h.db.QueryRow(ctx, `
		SELECT t.id, t.ticket_number, t.subject, COALESCE(t.description,''),
		  t.status::TEXT, t.priority::TEXT, t.source::TEXT,
		  COALESCE(t.category_id::TEXT,''), COALESCE(cat.name,''),
		  COALESCE(t.sla_policy_id::TEXT,''), COALESCE(sp.name,''),
		  COALESCE(t.assigned_agent_id::TEXT,''), COALESCE(a.name,''),
		  COALESCE(t.requester_name,''), COALESCE(t.requester_email,''),
		  COALESCE(t.requester_phone,''), COALESCE(t.company_name,''),
		  COALESCE(t.due_date::TEXT,''), COALESCE(t.first_response_at::TEXT,''),
		  COALESCE(t.resolved_at::TEXT,''), COALESCE(t.closed_at::TEXT,''),
		  t.tags, t.created_at::TEXT, t.updated_at::TEXT
		FROM helpdesk_tickets t
		LEFT JOIN helpdesk_categories cat ON cat.id=t.category_id
		LEFT JOIN helpdesk_sla_policies sp ON sp.id=t.sla_policy_id
		LEFT JOIN helpdesk_agents a ON a.id=t.assigned_agent_id
		WHERE t.id=$1 AND t.company_id=$2`, id, companyID).Scan(
		&t.ID, &t.TicketNumber, &t.Subject, &t.Description,
		&t.Status, &t.Priority, &t.Source,
		&t.CategoryID, &t.Category, &t.SLAPolicyID, &t.SLAPolicy,
		&t.AssignedAgentID, &t.AssignedAgent,
		&t.RequesterName, &t.RequesterEmail, &t.RequesterPhone, &t.CompanyName,
		&t.DueDate, &t.FirstResponseAt, &t.ResolvedAt, &t.ClosedAt,
		&t.Tags, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	// Load comments
	type Comment struct {
		ID         string `json:"id"`
		AuthorName string `json:"author_name"`
		Body       string `json:"body"`
		IsInternal bool   `json:"is_internal"`
		CreatedAt  string `json:"created_at"`
	}
	comments := []Comment{}
	cr, _ := h.db.Query(ctx, `
		SELECT id, COALESCE(author_name,''), body, is_internal, created_at::TEXT
		FROM ticket_comments WHERE ticket_id=$1 ORDER BY created_at ASC`, id)
	if cr != nil {
		defer cr.Close()
		for cr.Next() {
			var cm Comment
			if cr.Scan(&cm.ID, &cm.AuthorName, &cm.Body, &cm.IsInternal, &cm.CreatedAt) == nil {
				comments = append(comments, cm)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"ticket": t, "comments": comments})
}

func (h *HelpdeskHandler) CreateTicket(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate ticket number
	var seq int64
	_ = h.db.QueryRow(ctx, `SELECT nextval('helpdesk_ticket_seq')`).Scan(&seq)
	ticketNumber := fmt.Sprintf("TKT-%06d", seq)

	subject        := strVal(body, "subject")
	description    := strVal(body, "description")
	status         := strValDefault(body, "status", "open")
	priority       := strValDefault(body, "priority", "medium")
	source         := strValDefault(body, "source", "portal")
	requesterName  := strVal(body, "requester_name")
	requesterEmail := strVal(body, "requester_email")
	requesterPhone := strVal(body, "requester_phone")
	companyName    := strVal(body, "company_name")

	// Nullable UUIDs
	categoryID    := nullUUID(strVal(body, "category_id"))
	slaPolicyID   := nullUUID(strVal(body, "sla_policy_id"))
	assignedAgentID := nullUUID(strVal(body, "assigned_agent_id"))

	// Compute due_date from SLA
	var dueDate *time.Time
	if slaPolicyID != nil {
		var resHours int
		_ = h.db.QueryRow(ctx, `SELECT resolution_hours FROM helpdesk_sla_policies WHERE id=$1 AND company_id=$2`,
			slaPolicyID, companyID).Scan(&resHours)
		if resHours > 0 {
			d := time.Now().Add(time.Duration(resHours) * time.Hour)
			dueDate = &d
		}
	}

	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO helpdesk_tickets
		  (company_id, ticket_number, subject, description, status, priority, source,
		   category_id, sla_policy_id, assigned_agent_id,
		   requester_name, requester_email, requester_phone, company_name, due_date)
		VALUES ($1,$2,$3,$4,$5::ticket_status,$6::ticket_priority,$7::ticket_source,
		        $8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id`,
		companyID, ticketNumber, subject, description, status, priority, source,
		categoryID, slaPolicyID, assignedAgentID,
		requesterName, requesterEmail, requesterPhone, companyName, dueDate,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If agent assigned, record assignment
	if assignedAgentID != nil {
		_, _ = h.db.Exec(ctx, `
			INSERT INTO ticket_assignments (company_id, ticket_id, agent_id, assigned_by, is_current)
			VALUES ($1,$2,$3,'system',TRUE)`, companyID, id, assignedAgentID)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "ticket_number": ticketNumber})
}

func (h *HelpdeskHandler) UpdateTicket(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subject         := strVal(body, "subject")
	description     := strVal(body, "description")
	status          := strValDefault(body, "status", "open")
	priority        := strValDefault(body, "priority", "medium")
	source          := strValDefault(body, "source", "portal")
	requesterName   := strVal(body, "requester_name")
	requesterEmail  := strVal(body, "requester_email")
	requesterPhone  := strVal(body, "requester_phone")
	companyName     := strVal(body, "company_name")
	categoryID      := nullUUID(strVal(body, "category_id"))
	slaPolicyID     := nullUUID(strVal(body, "sla_policy_id"))
	assignedAgentID := nullUUID(strVal(body, "assigned_agent_id"))

	// resolved_at / closed_at based on status
	var resolvedAt, closedAt *time.Time
	if status == "resolved" {
		now := time.Now()
		resolvedAt = &now
	} else if status == "closed" {
		now := time.Now()
		closedAt = &now
	}

	_, err := h.db.Exec(ctx, `
		UPDATE helpdesk_tickets SET
		  subject=$3, description=$4, status=$5::ticket_status, priority=$6::ticket_priority,
		  source=$7::ticket_source, category_id=$8, sla_policy_id=$9, assigned_agent_id=$10,
		  requester_name=$11, requester_email=$12, requester_phone=$13, company_name=$14,
		  resolved_at=COALESCE($15,resolved_at), closed_at=COALESCE($16,closed_at),
		  updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, companyID, subject, description, status, priority, source,
		categoryID, slaPolicyID, assignedAgentID,
		requesterName, requesterEmail, requesterPhone, companyName,
		resolvedAt, closedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *HelpdeskHandler) UpdateTicketStatus(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var resolvedAt, closedAt *time.Time
	now := time.Now()
	if body.Status == "resolved" {
		resolvedAt = &now
	} else if body.Status == "closed" {
		closedAt = &now
	}

	_, err := h.db.Exec(ctx, `
		UPDATE helpdesk_tickets SET
		  status=$3::ticket_status,
		  resolved_at=COALESCE($4,resolved_at),
		  closed_at=COALESCE($5,closed_at),
		  updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, companyID, body.Status, resolvedAt, closedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *HelpdeskHandler) DeleteTicket(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	_, err := h.db.Exec(ctx, `DELETE FROM helpdesk_tickets WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Ticket Comments ───────────────────────────────────────────────────────────

func (h *HelpdeskHandler) AddComment(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ticketID := c.Param("id")
	ctx := c.Request.Context()

	var body struct {
		AuthorName string `json:"author_name"`
		Body       string `json:"body"`
		IsInternal bool   `json:"is_internal"`
		AgentID    string `json:"agent_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID := nullUUID(body.AgentID)

	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO ticket_comments (company_id, ticket_id, agent_id, author_name, body, is_internal)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		companyID, ticketID, agentID, body.AuthorName, body.Body, body.IsInternal,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Mark first response if this is the first agent comment
	if agentID != nil {
		_, _ = h.db.Exec(ctx, `
			UPDATE helpdesk_tickets SET first_response_at=COALESCE(first_response_at,NOW()), updated_at=NOW()
			WHERE id=$1 AND company_id=$2`, ticketID, companyID)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ── Categories ────────────────────────────────────────────────────────────────

func (h *HelpdeskHandler) ListCategories(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	rows, err := h.db.Query(ctx, `
		SELECT c.id, c.name, COALESCE(c.description,''), COALESCE(c.color,'#6366f1'),
		  COALESCE(c.parent_id::TEXT,''), COALESCE(p.name,''), c.is_active, c.sort_order,
		  c.created_at::TEXT,
		  (SELECT COUNT(*) FROM helpdesk_tickets t WHERE t.category_id=c.id) AS ticket_count
		FROM helpdesk_categories c
		LEFT JOIN helpdesk_categories p ON p.id=c.parent_id
		WHERE c.company_id=$1
		ORDER BY c.sort_order, c.name`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type Category struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		ParentID    string `json:"parent_id"`
		ParentName  string `json:"parent_name"`
		IsActive    bool   `json:"is_active"`
		SortOrder   int    `json:"sort_order"`
		CreatedAt   string `json:"created_at"`
		TicketCount int    `json:"ticket_count"`
	}

	cats := []Category{}
	for rows.Next() {
		var ct Category
		if rows.Scan(&ct.ID, &ct.Name, &ct.Description, &ct.Color, &ct.ParentID, &ct.ParentName,
			&ct.IsActive, &ct.SortOrder, &ct.CreatedAt, &ct.TicketCount) == nil {
			cats = append(cats, ct)
		}
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

func (h *HelpdeskHandler) CreateCategory(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name        := strVal(body, "name")
	description := strVal(body, "description")
	color       := strValDefault(body, "color", "#6366f1")
	parentID    := nullUUID(strVal(body, "parent_id"))
	isActive    := true
	if v, ok := body["is_active"].(bool); ok { isActive = v }
	sortOrder := 0
	if v, ok := body["sort_order"].(float64); ok { sortOrder = int(v) }

	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO helpdesk_categories (company_id, name, description, color, parent_id, is_active, sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		companyID, name, description, color, parentID, isActive, sortOrder,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *HelpdeskHandler) UpdateCategory(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name        := strVal(body, "name")
	description := strVal(body, "description")
	color       := strValDefault(body, "color", "#6366f1")
	parentID    := nullUUID(strVal(body, "parent_id"))
	isActive    := true
	if v, ok := body["is_active"].(bool); ok { isActive = v }
	sortOrder := 0
	if v, ok := body["sort_order"].(float64); ok { sortOrder = int(v) }

	_, err := h.db.Exec(ctx, `
		UPDATE helpdesk_categories SET name=$3, description=$4, color=$5, parent_id=$6,
		  is_active=$7, sort_order=$8, updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, companyID, name, description, color, parentID, isActive, sortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *HelpdeskHandler) DeleteCategory(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	_, err := h.db.Exec(ctx, `DELETE FROM helpdesk_categories WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Agents ────────────────────────────────────────────────────────────────────

func (h *HelpdeskHandler) ListAgents(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	where := []string{"a.company_id=$1"}
	args := []interface{}{companyID}
	n := 2

	if s := c.Query("status"); s != "" {
		where = append(where, fmt.Sprintf("a.status=$%d::agent_status", n))
		args = append(args, s); n++
	}
	if search := c.Query("search"); search != "" {
		where = append(where, fmt.Sprintf("(a.name ILIKE $%d OR a.email ILIKE $%d)", n, n))
		args = append(args, "%"+search+"%"); n++
	}
	_ = n

	sql := `
		SELECT a.id, a.name, COALESCE(a.email,''), COALESCE(a.phone,''),
		  COALESCE(a.department,''), COALESCE(a.specialization,''),
		  a.status::TEXT, a.max_tickets, a.is_active, a.created_at::TEXT,
		  COUNT(CASE WHEN t.status NOT IN ('resolved','closed','cancelled') THEN 1 END) AS open_tickets,
		  COUNT(CASE WHEN t.status IN ('resolved','closed') THEN 1 END) AS resolved_tickets
		FROM helpdesk_agents a
		LEFT JOIN helpdesk_tickets t ON t.assigned_agent_id=a.id AND t.company_id=$1
		WHERE ` + strings.Join(where, " AND ") + `
		GROUP BY a.id
		ORDER BY a.name`

	rows, err := h.db.Query(ctx, sql, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type Agent struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Email          string `json:"email"`
		Phone          string `json:"phone"`
		Department     string `json:"department"`
		Specialization string `json:"specialization"`
		Status         string `json:"status"`
		MaxTickets     int    `json:"max_tickets"`
		IsActive       bool   `json:"is_active"`
		CreatedAt      string `json:"created_at"`
		OpenTickets    int    `json:"open_tickets"`
		ResolvedTickets int   `json:"resolved_tickets"`
	}

	agents := []Agent{}
	for rows.Next() {
		var a Agent
		if rows.Scan(&a.ID, &a.Name, &a.Email, &a.Phone, &a.Department, &a.Specialization,
			&a.Status, &a.MaxTickets, &a.IsActive, &a.CreatedAt, &a.OpenTickets, &a.ResolvedTickets) == nil {
			agents = append(agents, a)
		}
	}
	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

func (h *HelpdeskHandler) CreateAgent(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name           := strVal(body, "name")
	email          := strVal(body, "email")
	phone          := strVal(body, "phone")
	department     := strVal(body, "department")
	specialization := strVal(body, "specialization")
	status         := strValDefault(body, "status", "active")
	maxTickets := 20
	if v, ok := body["max_tickets"].(float64); ok { maxTickets = int(v) }

	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO helpdesk_agents (company_id, name, email, phone, department, specialization, status, max_tickets)
		VALUES ($1,$2,$3,$4,$5,$6,$7::agent_status,$8) RETURNING id`,
		companyID, name, email, phone, department, specialization, status, maxTickets,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *HelpdeskHandler) UpdateAgent(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name           := strVal(body, "name")
	email          := strVal(body, "email")
	phone          := strVal(body, "phone")
	department     := strVal(body, "department")
	specialization := strVal(body, "specialization")
	status         := strValDefault(body, "status", "active")
	isActive       := true
	if v, ok := body["is_active"].(bool); ok { isActive = v }
	maxTickets := 20
	if v, ok := body["max_tickets"].(float64); ok { maxTickets = int(v) }

	_, err := h.db.Exec(ctx, `
		UPDATE helpdesk_agents SET name=$3, email=$4, phone=$5, department=$6,
		  specialization=$7, status=$8::agent_status, max_tickets=$9, is_active=$10, updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, companyID, name, email, phone, department, specialization, status, maxTickets, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *HelpdeskHandler) DeleteAgent(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	_, err := h.db.Exec(ctx, `DELETE FROM helpdesk_agents WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Ticket Assignments ────────────────────────────────────────────────────────

func (h *HelpdeskHandler) ListAssignments(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	where := []string{"ta.company_id=$1"}
	args := []interface{}{companyID}
	n := 2

	if ticketID := c.Query("ticket_id"); ticketID != "" {
		where = append(where, fmt.Sprintf("ta.ticket_id=$%d", n))
		args = append(args, ticketID); n++
	}
	if agentID := c.Query("agent_id"); agentID != "" {
		where = append(where, fmt.Sprintf("ta.agent_id=$%d", n))
		args = append(args, agentID); n++
	}
	_ = n

	sql := `
		SELECT ta.id, ta.ticket_id, t.ticket_number, t.subject,
		  COALESCE(ta.agent_id::TEXT,''), COALESCE(a.name,'Unassigned'),
		  COALESCE(ta.assigned_by,''), COALESCE(ta.reason,''),
		  ta.assigned_at::TEXT, COALESCE(ta.unassigned_at::TEXT,''), ta.is_current
		FROM ticket_assignments ta
		LEFT JOIN helpdesk_tickets t ON t.id=ta.ticket_id
		LEFT JOIN helpdesk_agents a ON a.id=ta.agent_id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY ta.assigned_at DESC`

	rows, err := h.db.Query(ctx, sql, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type Assignment struct {
		ID           string `json:"id"`
		TicketID     string `json:"ticket_id"`
		TicketNumber string `json:"ticket_number"`
		Subject      string `json:"subject"`
		AgentID      string `json:"agent_id"`
		AgentName    string `json:"agent_name"`
		AssignedBy   string `json:"assigned_by"`
		Reason       string `json:"reason"`
		AssignedAt   string `json:"assigned_at"`
		UnassignedAt string `json:"unassigned_at"`
		IsCurrent    bool   `json:"is_current"`
	}

	assignments := []Assignment{}
	for rows.Next() {
		var a Assignment
		if rows.Scan(&a.ID, &a.TicketID, &a.TicketNumber, &a.Subject,
			&a.AgentID, &a.AgentName, &a.AssignedBy, &a.Reason,
			&a.AssignedAt, &a.UnassignedAt, &a.IsCurrent) == nil {
			assignments = append(assignments, a)
		}
	}
	c.JSON(http.StatusOK, gin.H{"assignments": assignments})
}

func (h *HelpdeskHandler) AssignTicket(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ticketID := c.Param("id")
	ctx := c.Request.Context()

	var body struct {
		AgentID    string `json:"agent_id"`
		AssignedBy string `json:"assigned_by"`
		Reason     string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mark previous assignments as not current
	_, _ = h.db.Exec(ctx, `
		UPDATE ticket_assignments SET is_current=FALSE, unassigned_at=NOW()
		WHERE ticket_id=$1 AND is_current=TRUE`, ticketID)

	agentID := nullUUID(body.AgentID)

	// Insert new assignment
	_, err := h.db.Exec(ctx, `
		INSERT INTO ticket_assignments (company_id, ticket_id, agent_id, assigned_by, reason, is_current)
		VALUES ($1,$2,$3,$4,$5,TRUE)`,
		companyID, ticketID, agentID, body.AssignedBy, body.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update ticket's assigned agent
	_, _ = h.db.Exec(ctx, `
		UPDATE helpdesk_tickets SET assigned_agent_id=$2, status='in_progress'::ticket_status, updated_at=NOW()
		WHERE id=$1 AND company_id=$3`, ticketID, agentID, companyID)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── Escalations ───────────────────────────────────────────────────────────────

func (h *HelpdeskHandler) ListEscalations(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	where := []string{"e.company_id=$1"}
	args := []interface{}{companyID}
	n := 2

	if s := c.Query("status"); s != "" {
		where = append(where, fmt.Sprintf("e.status=$%d::escalation_status", n))
		args = append(args, s); n++
	}
	if tid := c.Query("ticket_id"); tid != "" {
		where = append(where, fmt.Sprintf("e.ticket_id=$%d", n))
		args = append(args, tid); n++
	}
	_ = n

	sql := `
		SELECT e.id, e.ticket_id, t.ticket_number, t.subject, t.priority::TEXT,
		  COALESCE(e.escalated_by,''), COALESCE(e.escalated_to,''),
		  COALESCE(e.reason,''), e.status::TEXT,
		  COALESCE(e.resolution_note,''), e.escalated_at::TEXT,
		  COALESCE(e.resolved_at::TEXT,'')
		FROM ticket_escalations e
		JOIN helpdesk_tickets t ON t.id=e.ticket_id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY e.escalated_at DESC`

	rows, err := h.db.Query(ctx, sql, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type Escalation struct {
		ID             string `json:"id"`
		TicketID       string `json:"ticket_id"`
		TicketNumber   string `json:"ticket_number"`
		Subject        string `json:"subject"`
		Priority       string `json:"priority"`
		EscalatedBy    string `json:"escalated_by"`
		EscalatedTo    string `json:"escalated_to"`
		Reason         string `json:"reason"`
		Status         string `json:"status"`
		ResolutionNote string `json:"resolution_note"`
		EscalatedAt    string `json:"escalated_at"`
		ResolvedAt     string `json:"resolved_at"`
	}

	escalations := []Escalation{}
	for rows.Next() {
		var e Escalation
		if rows.Scan(&e.ID, &e.TicketID, &e.TicketNumber, &e.Subject, &e.Priority,
			&e.EscalatedBy, &e.EscalatedTo, &e.Reason, &e.Status,
			&e.ResolutionNote, &e.EscalatedAt, &e.ResolvedAt) == nil {
			escalations = append(escalations, e)
		}
	}
	c.JSON(http.StatusOK, gin.H{"escalations": escalations})
}

func (h *HelpdeskHandler) CreateEscalation(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketID    := strVal(body, "ticket_id")
	escalatedBy := strVal(body, "escalated_by")
	escalatedTo := strVal(body, "escalated_to")
	reason      := strVal(body, "reason")

	// Update ticket priority to critical
	_, _ = h.db.Exec(ctx, `
		UPDATE helpdesk_tickets SET priority='critical'::ticket_priority, updated_at=NOW()
		WHERE id=$1 AND company_id=$2`, ticketID, companyID)

	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO ticket_escalations (company_id, ticket_id, escalated_by, escalated_to, reason)
		VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		companyID, ticketID, escalatedBy, escalatedTo, reason,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *HelpdeskHandler) UpdateEscalationStatus(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	var body struct {
		Status         string `json:"status" binding:"required"`
		ResolutionNote string `json:"resolution_note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var resolvedAt *time.Time
	if body.Status == "resolved" || body.Status == "closed" {
		now := time.Now()
		resolvedAt = &now
	}

	_, err := h.db.Exec(ctx, `
		UPDATE ticket_escalations SET
		  status=$3::escalation_status, resolution_note=$4,
		  resolved_at=COALESCE($5,resolved_at), updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, companyID, body.Status, body.ResolutionNote, resolvedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *HelpdeskHandler) DeleteEscalation(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	_, err := h.db.Exec(ctx, `DELETE FROM ticket_escalations WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ── SLA Policies ──────────────────────────────────────────────────────────────

func (h *HelpdeskHandler) ListSLAPolicies(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	rows, err := h.db.Query(ctx, `
		SELECT id, name, COALESCE(description,''), priority::TEXT,
		  first_response_hours, resolution_hours, business_hours_only, is_active,
		  created_at::TEXT,
		  (SELECT COUNT(*) FROM helpdesk_tickets t WHERE t.sla_policy_id=sp.id) AS ticket_count
		FROM helpdesk_sla_policies sp
		WHERE company_id=$1 ORDER BY priority, name`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type SLAPolicy struct {
		ID                  string `json:"id"`
		Name                string `json:"name"`
		Description         string `json:"description"`
		Priority            string `json:"priority"`
		FirstResponseHours  int    `json:"first_response_hours"`
		ResolutionHours     int    `json:"resolution_hours"`
		BusinessHoursOnly   bool   `json:"business_hours_only"`
		IsActive            bool   `json:"is_active"`
		CreatedAt           string `json:"created_at"`
		TicketCount         int    `json:"ticket_count"`
	}

	policies := []SLAPolicy{}
	for rows.Next() {
		var p SLAPolicy
		if rows.Scan(&p.ID, &p.Name, &p.Description, &p.Priority,
			&p.FirstResponseHours, &p.ResolutionHours, &p.BusinessHoursOnly, &p.IsActive,
			&p.CreatedAt, &p.TicketCount) == nil {
			policies = append(policies, p)
		}
	}
	c.JSON(http.StatusOK, gin.H{"policies": policies})
}

func (h *HelpdeskHandler) CreateSLAPolicy(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name        := strVal(body, "name")
	description := strVal(body, "description")
	priority    := strValDefault(body, "priority", "medium")
	firstResponseHours := 4
	if v, ok := body["first_response_hours"].(float64); ok { firstResponseHours = int(v) }
	resolutionHours := 24
	if v, ok := body["resolution_hours"].(float64); ok { resolutionHours = int(v) }
	businessHoursOnly := true
	if v, ok := body["business_hours_only"].(bool); ok { businessHoursOnly = v }

	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO helpdesk_sla_policies
		  (company_id, name, description, priority, first_response_hours, resolution_hours, business_hours_only)
		VALUES ($1,$2,$3,$4::sla_priority,$5,$6,$7) RETURNING id`,
		companyID, name, description, priority, firstResponseHours, resolutionHours, businessHoursOnly,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *HelpdeskHandler) UpdateSLAPolicy(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name        := strVal(body, "name")
	description := strVal(body, "description")
	priority    := strValDefault(body, "priority", "medium")
	firstResponseHours := 4
	if v, ok := body["first_response_hours"].(float64); ok { firstResponseHours = int(v) }
	resolutionHours := 24
	if v, ok := body["resolution_hours"].(float64); ok { resolutionHours = int(v) }
	businessHoursOnly := true
	if v, ok := body["business_hours_only"].(bool); ok { businessHoursOnly = v }
	isActive := true
	if v, ok := body["is_active"].(bool); ok { isActive = v }

	_, err := h.db.Exec(ctx, `
		UPDATE helpdesk_sla_policies SET name=$3, description=$4, priority=$5::sla_priority,
		  first_response_hours=$6, resolution_hours=$7, business_hours_only=$8, is_active=$9, updated_at=NOW()
		WHERE id=$1 AND company_id=$2`,
		id, companyID, name, description, priority, firstResponseHours, resolutionHours, businessHoursOnly, isActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *HelpdeskHandler) DeleteSLAPolicy(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := c.Request.Context()

	_, err := h.db.Exec(ctx, `DELETE FROM helpdesk_sla_policies WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetSLATracking returns SLA compliance overview
func (h *HelpdeskHandler) GetSLATracking(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	type SLASummary struct {
		PolicyID           string  `json:"policy_id"`
		PolicyName         string  `json:"policy_name"`
		Priority           string  `json:"priority"`
		TotalTickets       int     `json:"total_tickets"`
		WithinSLA          int     `json:"within_sla"`
		Breached           int     `json:"breached"`
		ComplianceRate     float64 `json:"compliance_rate"`
		AvgFirstResponseHrs float64 `json:"avg_first_response_hrs"`
		AvgResolutionHrs   float64 `json:"avg_resolution_hrs"`
		FirstResponseHoursTarget int `json:"first_response_hours_target"`
		ResolutionHoursTarget    int `json:"resolution_hours_target"`
	}

	rows, err := h.db.Query(ctx, `
		SELECT sp.id, sp.name, sp.priority::TEXT,
		  COUNT(t.id) AS total,
		  COUNT(CASE WHEN t.due_date IS NULL OR t.due_date >= NOW()
		             OR t.status IN ('resolved','closed') THEN 1 END) AS within_sla,
		  COUNT(CASE WHEN t.due_date < NOW() AND t.status NOT IN ('resolved','closed','cancelled') THEN 1 END) AS breached,
		  COALESCE(AVG(CASE WHEN t.first_response_at IS NOT NULL
		    THEN EXTRACT(EPOCH FROM (t.first_response_at - t.created_at))/3600 END),0),
		  COALESCE(AVG(CASE WHEN t.resolved_at IS NOT NULL
		    THEN EXTRACT(EPOCH FROM (t.resolved_at - t.created_at))/3600 END),0),
		  sp.first_response_hours, sp.resolution_hours
		FROM helpdesk_sla_policies sp
		LEFT JOIN helpdesk_tickets t ON t.sla_policy_id=sp.id AND t.company_id=$1
		WHERE sp.company_id=$1
		GROUP BY sp.id, sp.name, sp.priority, sp.first_response_hours, sp.resolution_hours
		ORDER BY sp.priority`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	summaries := []SLASummary{}
	for rows.Next() {
		var s SLASummary
		if rows.Scan(&s.PolicyID, &s.PolicyName, &s.Priority, &s.TotalTickets,
			&s.WithinSLA, &s.Breached, &s.AvgFirstResponseHrs, &s.AvgResolutionHrs,
			&s.FirstResponseHoursTarget, &s.ResolutionHoursTarget) == nil {
			if s.TotalTickets > 0 {
				s.ComplianceRate = float64(s.WithinSLA) / float64(s.TotalTickets) * 100
			}
			summaries = append(summaries, s)
		}
	}

	// Overdue tickets detail
	type OverdueTicket struct {
		ID             string `json:"id"`
		TicketNumber   string `json:"ticket_number"`
		Subject        string `json:"subject"`
		Priority       string `json:"priority"`
		PolicyName     string `json:"policy_name"`
		DueDate        string `json:"due_date"`
		HoursOverdue   float64 `json:"hours_overdue"`
		AssignedAgent  string `json:"assigned_agent"`
	}
	overdueTickets := []OverdueTicket{}
	rows2, _ := h.db.Query(ctx, `
		SELECT t.id, t.ticket_number, t.subject, t.priority::TEXT,
		  COALESCE(sp.name,'No SLA'), t.due_date::TEXT,
		  EXTRACT(EPOCH FROM (NOW()-t.due_date))/3600 AS hours_overdue,
		  COALESCE(a.name,'Unassigned')
		FROM helpdesk_tickets t
		LEFT JOIN helpdesk_sla_policies sp ON sp.id=t.sla_policy_id
		LEFT JOIN helpdesk_agents a ON a.id=t.assigned_agent_id
		WHERE t.company_id=$1 AND t.due_date < NOW()
		  AND t.status NOT IN ('resolved','closed','cancelled')
		ORDER BY t.due_date ASC LIMIT 50`, companyID)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var ot OverdueTicket
			if rows2.Scan(&ot.ID, &ot.TicketNumber, &ot.Subject, &ot.Priority,
				&ot.PolicyName, &ot.DueDate, &ot.HoursOverdue, &ot.AssignedAgent) == nil {
				overdueTickets = append(overdueTickets, ot)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"summaries":      summaries,
		"overdue_tickets": overdueTickets,
	})
}

// ── CSAT Surveys ──────────────────────────────────────────────────────────────

func (h *HelpdeskHandler) ListCSAT(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	where := []string{"cs.company_id=$1"}
	args := []interface{}{companyID}
	n := 2

	if rating := c.Query("rating"); rating != "" {
		where = append(where, fmt.Sprintf("cs.rating=$%d::csat_rating", n))
		args = append(args, rating); n++
	}
	if agentID := c.Query("agent_id"); agentID != "" {
		where = append(where, fmt.Sprintf("cs.agent_id=$%d", n))
		args = append(args, agentID); n++
	}
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		where = append(where, fmt.Sprintf("cs.submitted_at>=$%d", n))
		args = append(args, dateFrom); n++
	}
	if dateTo := c.Query("date_to"); dateTo != "" {
		where = append(where, fmt.Sprintf("cs.submitted_at<=$%d", n))
		args = append(args, dateTo); n++
	}
	_ = n

	sql := `
		SELECT cs.id, cs.ticket_id, t.ticket_number, cs.rating::TEXT,
		  COALESCE(cs.comment,''), COALESCE(cs.requester_name,''),
		  COALESCE(cs.requester_email,''), COALESCE(a.name,''),
		  cs.submitted_at::TEXT
		FROM csat_surveys cs
		JOIN helpdesk_tickets t ON t.id=cs.ticket_id
		LEFT JOIN helpdesk_agents a ON a.id=cs.agent_id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY cs.submitted_at DESC`

	rows, err := h.db.Query(ctx, sql, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type Survey struct {
		ID             string `json:"id"`
		TicketID       string `json:"ticket_id"`
		TicketNumber   string `json:"ticket_number"`
		Rating         string `json:"rating"`
		Comment        string `json:"comment"`
		RequesterName  string `json:"requester_name"`
		RequesterEmail string `json:"requester_email"`
		AgentName      string `json:"agent_name"`
		SubmittedAt    string `json:"submitted_at"`
	}

	surveys := []Survey{}
	for rows.Next() {
		var s Survey
		if rows.Scan(&s.ID, &s.TicketID, &s.TicketNumber, &s.Rating, &s.Comment,
			&s.RequesterName, &s.RequesterEmail, &s.AgentName, &s.SubmittedAt) == nil {
			surveys = append(surveys, s)
		}
	}

	// CSAT aggregates
	type CSATAggregate struct {
		TotalResponses     int     `json:"total_responses"`
		AvgScore           float64 `json:"avg_score"`
		VeryDissatisfied   int     `json:"very_dissatisfied"`
		Dissatisfied       int     `json:"dissatisfied"`
		Neutral            int     `json:"neutral"`
		Satisfied          int     `json:"satisfied"`
		VerySatisfied      int     `json:"very_satisfied"`
		SatisfactionRate   float64 `json:"satisfaction_rate"`
	}
	var agg CSATAggregate
	_ = h.db.QueryRow(ctx, `
		SELECT
		  COUNT(*),
		  COALESCE(AVG(CASE rating WHEN 'very_dissatisfied' THEN 1 WHEN 'dissatisfied' THEN 2
		    WHEN 'neutral' THEN 3 WHEN 'satisfied' THEN 4 WHEN 'very_satisfied' THEN 5 END),0),
		  COUNT(CASE WHEN rating='very_dissatisfied' THEN 1 END),
		  COUNT(CASE WHEN rating='dissatisfied' THEN 1 END),
		  COUNT(CASE WHEN rating='neutral' THEN 1 END),
		  COUNT(CASE WHEN rating='satisfied' THEN 1 END),
		  COUNT(CASE WHEN rating='very_satisfied' THEN 1 END)
		FROM csat_surveys WHERE company_id=$1`, companyID).Scan(
		&agg.TotalResponses, &agg.AvgScore,
		&agg.VeryDissatisfied, &agg.Dissatisfied, &agg.Neutral,
		&agg.Satisfied, &agg.VerySatisfied,
	)
	if agg.TotalResponses > 0 {
		agg.SatisfactionRate = float64(agg.Satisfied+agg.VerySatisfied) / float64(agg.TotalResponses) * 100
	}

	c.JSON(http.StatusOK, gin.H{"surveys": surveys, "aggregate": agg})
}

func (h *HelpdeskHandler) CreateCSAT(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticketID       := strVal(body, "ticket_id")
	rating         := strValDefault(body, "rating", "neutral")
	comment        := strVal(body, "comment")
	requesterName  := strVal(body, "requester_name")
	requesterEmail := strVal(body, "requester_email")
	agentID        := nullUUID(strVal(body, "agent_id"))

	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO csat_surveys (company_id, ticket_id, agent_id, rating, comment, requester_name, requester_email)
		VALUES ($1,$2,$3,$4::csat_rating,$5,$6,$7) RETURNING id`,
		companyID, ticketID, agentID, rating, comment, requesterName, requesterEmail,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ── Reports ───────────────────────────────────────────────────────────────────

func (h *HelpdeskHandler) GetReports(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	dateFrom := c.Query("date_from")
	dateTo   := c.Query("date_to")
	if dateFrom == "" { dateFrom = time.Now().AddDate(0, -1, 0).Format("2006-01-02") }
	if dateTo == ""   { dateTo = time.Now().Format("2006-01-02") }

	// Overall stats
	type Stats struct {
		TotalTickets        int     `json:"total_tickets"`
		Resolved            int     `json:"resolved"`
		Closed              int     `json:"closed"`
		Open                int     `json:"open"`
		AvgFirstResponseHrs float64 `json:"avg_first_response_hrs"`
		AvgResolutionHrs    float64 `json:"avg_resolution_hrs"`
		SLABreached         int     `json:"sla_breached"`
		SLACompliance       float64 `json:"sla_compliance"`
		CSATAvg             float64 `json:"csat_avg"`
	}
	var stats Stats
	_ = h.db.QueryRow(ctx, `
		SELECT COUNT(*),
		  COUNT(CASE WHEN status='resolved' THEN 1 END),
		  COUNT(CASE WHEN status='closed' THEN 1 END),
		  COUNT(CASE WHEN status='open' THEN 1 END),
		  COALESCE(AVG(CASE WHEN first_response_at IS NOT NULL
		    THEN EXTRACT(EPOCH FROM (first_response_at-created_at))/3600 END),0),
		  COALESCE(AVG(CASE WHEN resolved_at IS NOT NULL
		    THEN EXTRACT(EPOCH FROM (resolved_at-created_at))/3600 END),0),
		  COUNT(CASE WHEN due_date < NOW() AND status NOT IN ('resolved','closed','cancelled') THEN 1 END),
		  0
		FROM helpdesk_tickets
		WHERE company_id=$1 AND created_at BETWEEN $2::timestamptz AND $3::timestamptz+'1 day'::interval`,
		companyID, dateFrom, dateTo).Scan(
		&stats.TotalTickets, &stats.Resolved, &stats.Closed, &stats.Open,
		&stats.AvgFirstResponseHrs, &stats.AvgResolutionHrs, &stats.SLABreached,
		&stats.SLACompliance,
	)
	if stats.TotalTickets > 0 {
		stats.SLACompliance = float64(stats.TotalTickets-stats.SLABreached) / float64(stats.TotalTickets) * 100
	}
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(CASE rating WHEN 'very_dissatisfied' THEN 1 WHEN 'dissatisfied' THEN 2
		  WHEN 'neutral' THEN 3 WHEN 'satisfied' THEN 4 WHEN 'very_satisfied' THEN 5 END),0)
		FROM csat_surveys WHERE company_id=$1 AND submitted_at BETWEEN $2::timestamptz AND $3::timestamptz+'1 day'::interval`,
		companyID, dateFrom, dateTo).Scan(&stats.CSATAvg)

	// By category
	type CatReport struct {
		Category string `json:"category"`
		Total    int    `json:"total"`
		Resolved int    `json:"resolved"`
		AvgHrs   float64 `json:"avg_resolution_hrs"`
	}
	catReports := []CatReport{}
	rows, _ := h.db.Query(ctx, `
		SELECT COALESCE(cat.name,'Uncategorized'), COUNT(*),
		  COUNT(CASE WHEN t.status IN ('resolved','closed') THEN 1 END),
		  COALESCE(AVG(CASE WHEN t.resolved_at IS NOT NULL
		    THEN EXTRACT(EPOCH FROM (t.resolved_at-t.created_at))/3600 END),0)
		FROM helpdesk_tickets t
		LEFT JOIN helpdesk_categories cat ON cat.id=t.category_id
		WHERE t.company_id=$1 AND t.created_at BETWEEN $2::timestamptz AND $3::timestamptz+'1 day'::interval
		GROUP BY cat.name ORDER BY COUNT(*) DESC`, companyID, dateFrom, dateTo)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var cr CatReport
			if rows.Scan(&cr.Category, &cr.Total, &cr.Resolved, &cr.AvgHrs) == nil {
				catReports = append(catReports, cr)
			}
		}
	}

	// By agent
	type AgentReport struct {
		AgentName  string  `json:"agent_name"`
		Total      int     `json:"total"`
		Resolved   int     `json:"resolved"`
		AvgHrs     float64 `json:"avg_resolution_hrs"`
		CSATAvg    float64 `json:"csat_avg"`
	}
	agentReports := []AgentReport{}
	rows2, _ := h.db.Query(ctx, `
		SELECT COALESCE(a.name,'Unassigned'), COUNT(t.id),
		  COUNT(CASE WHEN t.status IN ('resolved','closed') THEN 1 END),
		  COALESCE(AVG(CASE WHEN t.resolved_at IS NOT NULL
		    THEN EXTRACT(EPOCH FROM (t.resolved_at-t.created_at))/3600 END),0),
		  COALESCE(AVG(CASE cs.rating WHEN 'very_dissatisfied' THEN 1 WHEN 'dissatisfied' THEN 2
		    WHEN 'neutral' THEN 3 WHEN 'satisfied' THEN 4 WHEN 'very_satisfied' THEN 5 END),0)
		FROM helpdesk_tickets t
		LEFT JOIN helpdesk_agents a ON a.id=t.assigned_agent_id
		LEFT JOIN csat_surveys cs ON cs.ticket_id=t.id
		WHERE t.company_id=$1 AND t.created_at BETWEEN $2::timestamptz AND $3::timestamptz+'1 day'::interval
		GROUP BY a.name ORDER BY COUNT(t.id) DESC LIMIT 20`, companyID, dateFrom, dateTo)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var ar AgentReport
			if rows2.Scan(&ar.AgentName, &ar.Total, &ar.Resolved, &ar.AvgHrs, &ar.CSATAvg) == nil {
				agentReports = append(agentReports, ar)
			}
		}
	}

	// By source
	type SourceReport struct {
		Source string `json:"source"`
		Count  int    `json:"count"`
	}
	sourceReports := []SourceReport{}
	rows3, _ := h.db.Query(ctx, `
		SELECT source::TEXT, COUNT(*) FROM helpdesk_tickets
		WHERE company_id=$1 AND created_at BETWEEN $2::timestamptz AND $3::timestamptz+'1 day'::interval
		GROUP BY source ORDER BY COUNT(*) DESC`, companyID, dateFrom, dateTo)
	if rows3 != nil {
		defer rows3.Close()
		for rows3.Next() {
			var sr SourceReport
			if rows3.Scan(&sr.Source, &sr.Count) == nil {
				sourceReports = append(sourceReports, sr)
			}
		}
	}

	// Daily trend
	type DailyReport struct {
		Date     string `json:"date"`
		Opened   int    `json:"opened"`
		Resolved int    `json:"resolved"`
		Closed   int    `json:"closed"`
	}
	dailyReports := []DailyReport{}
	rows4, _ := h.db.Query(ctx, `
		WITH days AS (
		  SELECT generate_series($2::date, $3::date, '1 day'::interval)::date AS d
		)
		SELECT d::TEXT,
		  COALESCE(SUM(CASE WHEN DATE(t.created_at)=d THEN 1 ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN DATE(t.resolved_at)=d THEN 1 ELSE 0 END),0),
		  COALESCE(SUM(CASE WHEN DATE(t.closed_at)=d THEN 1 ELSE 0 END),0)
		FROM days
		LEFT JOIN helpdesk_tickets t ON t.company_id=$1
		GROUP BY d ORDER BY d`, companyID, dateFrom, dateTo)
	if rows4 != nil {
		defer rows4.Close()
		for rows4.Next() {
			var dr DailyReport
			if rows4.Scan(&dr.Date, &dr.Opened, &dr.Resolved, &dr.Closed) == nil {
				dailyReports = append(dailyReports, dr)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":          stats,
		"by_category":    catReports,
		"by_agent":       agentReports,
		"by_source":      sourceReports,
		"daily_trend":    dailyReports,
		"date_from":      dateFrom,
		"date_to":        dateTo,
	})
}

// ── Helper: nullUUID ──────────────────────────────────────────────────────────

func nullUUID(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" || s == "null" || s == "undefined" {
		return nil
	}
	return s
}
