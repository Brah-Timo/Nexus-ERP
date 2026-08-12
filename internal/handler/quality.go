package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"nexus-erp/internal/middleware"
)

// QualityHandler handles all quality management HTTP requests.
type QualityHandler struct {
	db *pgxpool.Pool
}

// ── helper functions (q-prefixed to avoid conflicts) ──────────────────────────

func qStr(m map[string]interface{}, k string) string {
	if v, ok := m[k]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func qStrD(m map[string]interface{}, k, def string) string {
	s := qStr(m, k)
	if s == "" {
		return def
	}
	return s
}

func qStrN(m map[string]interface{}, k string) *string {
	s := qStr(m, k)
	if s == "" {
		return nil
	}
	return &s
}

func qF64(m map[string]interface{}, k string) float64 {
	if v, ok := m[k]; ok && v != nil {
		switch t := v.(type) {
		case float64:
			return t
		case string:
			if f, err := strconv.ParseFloat(t, 64); err == nil {
				return f
			}
		}
	}
	return 0
}

func qInt(m map[string]interface{}, k string) int {
	if v, ok := m[k]; ok && v != nil {
		switch t := v.(type) {
		case float64:
			return int(t)
		case int:
			return t
		case string:
			if i, err := strconv.Atoi(t); err == nil {
				return i
			}
		}
	}
	return 0
}

func qBool(m map[string]interface{}, k string) bool {
	if v, ok := m[k]; ok && v != nil {
		switch t := v.(type) {
		case bool:
			return t
		case string:
			return strings.EqualFold(t, "true") || t == "1"
		}
	}
	return false
}

// ─────────────────────────────────────────────────────────────────────────────
// DASHBOARD
// ─────────────────────────────────────────────────────────────────────────────

func (h *QualityHandler) GetDashboard(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	type Stats struct {
		TotalInspections   int     `json:"total_inspections"`
		PendingInspections int     `json:"pending_inspections"`
		PassedInspections  int     `json:"passed_inspections"`
		FailedInspections  int     `json:"failed_inspections"`
		FirstPassRate      float64 `json:"first_pass_rate"`
		TotalChecks        int     `json:"total_checks"`
		FailedChecks       int     `json:"failed_checks"`
		TotalNC            int     `json:"total_nc"`
		OpenNC             int     `json:"open_nc"`
		CriticalNC         int     `json:"critical_nc"`
		TotalCA            int     `json:"total_ca"`
		OpenCA             int     `json:"open_ca"`
		OverdueCA          int     `json:"overdue_ca"`
		DefectRate         float64 `json:"defect_rate"`
	}

	var s Stats

	// Inspection stats
	h.db.QueryRow(ctx, `
		SELECT
			COUNT(*)                                                              AS total,
			COUNT(*) FILTER (WHERE status IN ('pending','in_progress'))          AS pending,
			COUNT(*) FILTER (WHERE status = 'passed')                            AS passed,
			COUNT(*) FILTER (WHERE status = 'failed')                            AS failed,
			COALESCE(
				ROUND(COUNT(*) FILTER (WHERE status='passed')::NUMERIC /
				NULLIF(COUNT(*) FILTER (WHERE status IN ('passed','failed')),0)*100, 2)
			, 0)                                                                  AS pass_rate
		FROM quality_inspections
		WHERE company_id = $1`,
		companyID,
	).Scan(&s.TotalInspections, &s.PendingInspections, &s.PassedInspections, &s.FailedInspections, &s.FirstPassRate)

	// Check stats
	h.db.QueryRow(ctx, `
		SELECT
			COUNT(*)                                          AS total_checks,
			COUNT(*) FILTER (WHERE result = 'fail')          AS failed_checks
		FROM quality_checks qc
		JOIN quality_inspections qi ON qi.id = qc.inspection_id
		WHERE qi.company_id = $1`,
		companyID,
	).Scan(&s.TotalChecks, &s.FailedChecks)

	// NC stats
	h.db.QueryRow(ctx, `
		SELECT
			COUNT(*)                                                       AS total,
			COUNT(*) FILTER (WHERE status NOT IN ('closed','cancelled'))   AS open,
			COUNT(*) FILTER (WHERE severity IN ('critical','critical_safety')
			                  AND status NOT IN ('closed','cancelled'))    AS critical
		FROM non_conformities
		WHERE company_id = $1`,
		companyID,
	).Scan(&s.TotalNC, &s.OpenNC, &s.CriticalNC)

	// CA stats
	h.db.QueryRow(ctx, `
		SELECT
			COUNT(*)                                                                    AS total,
			COUNT(*) FILTER (WHERE status NOT IN ('closed','cancelled','verified'))    AS open,
			COUNT(*) FILTER (WHERE due_date < CURRENT_DATE
			                  AND status NOT IN ('closed','cancelled','verified'))     AS overdue
		FROM corrective_actions
		WHERE company_id = $1`,
		companyID,
	).Scan(&s.TotalCA, &s.OpenCA, &s.OverdueCA)

	// Defect rate
	if s.TotalChecks > 0 {
		s.DefectRate = float64(s.FailedChecks) / float64(s.TotalChecks) * 100
	}

	// Recent inspections
	type RecentInspection struct {
		ID             string  `json:"id"`
		Reference      string  `json:"reference"`
		InspectionType string  `json:"inspection_type"`
		Status         string  `json:"status"`
		ItemName       string  `json:"item_name"`
		QtyToInspect   float64 `json:"qty_to_inspect"`
		QtyPassed      float64 `json:"qty_passed"`
		QtyFailed      float64 `json:"qty_failed"`
		InspectorName  string  `json:"inspector_name"`
		CreatedAt      string  `json:"created_at"`
	}

	var recentInspections []RecentInspection
	rows, err := h.db.Query(ctx, `
		SELECT
			qi.id, qi.reference, qi.inspection_type::text, qi.status::text,
			COALESCE(i.name,''),
			qi.qty_to_inspect, qi.qty_passed, qi.qty_failed,
			COALESCE(u.full_name,''),
			qi.created_at
		FROM quality_inspections qi
		LEFT JOIN items i ON i.id = qi.item_id
		LEFT JOIN users u ON u.id = qi.inspector_id
		WHERE qi.company_id = $1
		ORDER BY qi.created_at DESC
		LIMIT 10`,
		companyID,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var r RecentInspection
			var createdAt time.Time
			rows.Scan(&r.ID, &r.Reference, &r.InspectionType, &r.Status,
				&r.ItemName, &r.QtyToInspect, &r.QtyPassed, &r.QtyFailed,
				&r.InspectorName, &createdAt)
			r.CreatedAt = createdAt.Format(time.RFC3339)
			recentInspections = append(recentInspections, r)
		}
	}
	if recentInspections == nil {
		recentInspections = []RecentInspection{}
	}

	// Recent NCs
	type RecentNC struct {
		ID           string `json:"id"`
		Reference    string `json:"reference"`
		Title        string `json:"title"`
		Status       string `json:"status"`
		Severity     string `json:"severity"`
		DetectedDate string `json:"detected_date"`
		AssigneeName string `json:"assignee_name"`
	}
	var recentNCs []RecentNC
	rows2, err2 := h.db.Query(ctx, `
		SELECT
			nc.id, nc.reference, nc.title, nc.status::text, nc.severity::text,
			nc.detected_date,
			COALESCE(u.full_name,'')
		FROM non_conformities nc
		LEFT JOIN users u ON u.id = nc.assigned_to
		WHERE nc.company_id = $1 AND nc.status NOT IN ('closed','cancelled')
		ORDER BY
			CASE nc.severity
				WHEN 'critical_safety' THEN 1
				WHEN 'critical'        THEN 2
				WHEN 'major'           THEN 3
				ELSE 4
			END,
			nc.created_at DESC
		LIMIT 10`,
		companyID,
	)
	if err2 == nil {
		defer rows2.Close()
		for rows2.Next() {
			var r RecentNC
			var dDate time.Time
			rows2.Scan(&r.ID, &r.Reference, &r.Title, &r.Status, &r.Severity,
				&dDate, &r.AssigneeName)
			r.DetectedDate = dDate.Format("2006-01-02")
			recentNCs = append(recentNCs, r)
		}
	}
	if recentNCs == nil {
		recentNCs = []RecentNC{}
	}

	// Monthly trend (last 6 months)
	type MonthTrend struct {
		Month      string  `json:"month"`
		Inspections int    `json:"inspections"`
		Passed      int    `json:"passed"`
		Failed      int    `json:"failed"`
		NCCount     int    `json:"nc_count"`
		PassRate    float64 `json:"pass_rate"`
	}
	var trend []MonthTrend
	rows3, err3 := h.db.Query(ctx, `
		SELECT
			TO_CHAR(DATE_TRUNC('month', qi.created_at), 'YYYY-MM') AS month,
			COUNT(*)                                                AS inspections,
			COUNT(*) FILTER (WHERE qi.status = 'passed')           AS passed,
			COUNT(*) FILTER (WHERE qi.status = 'failed')           AS failed,
			COALESCE(
				ROUND(COUNT(*) FILTER (WHERE qi.status='passed')::NUMERIC /
				NULLIF(COUNT(*) FILTER (WHERE qi.status IN ('passed','failed')),0)*100, 2)
			,0)                                                     AS pass_rate
		FROM quality_inspections qi
		WHERE qi.company_id = $1
		  AND qi.created_at >= NOW() - INTERVAL '6 months'
		GROUP BY DATE_TRUNC('month', qi.created_at)
		ORDER BY 1`,
		companyID,
	)
	if err3 == nil {
		defer rows3.Close()
		for rows3.Next() {
			var t MonthTrend
			rows3.Scan(&t.Month, &t.Inspections, &t.Passed, &t.Failed, &t.PassRate)
			trend = append(trend, t)
		}
	}
	if trend == nil {
		trend = []MonthTrend{}
	}

	// NC by severity
	type SeverityCount struct {
		Severity string `json:"severity"`
		Count    int    `json:"count"`
	}
	var ncBySeverity []SeverityCount
	rows4, err4 := h.db.Query(ctx, `
		SELECT severity::text, COUNT(*) AS cnt
		FROM non_conformities
		WHERE company_id = $1 AND status NOT IN ('closed','cancelled')
		GROUP BY severity
		ORDER BY cnt DESC`,
		companyID,
	)
	if err4 == nil {
		defer rows4.Close()
		for rows4.Next() {
			var sc SeverityCount
			rows4.Scan(&sc.Severity, &sc.Count)
			ncBySeverity = append(ncBySeverity, sc)
		}
	}
	if ncBySeverity == nil {
		ncBySeverity = []SeverityCount{}
	}

	c.JSON(http.StatusOK, gin.H{
		"stats":              s,
		"recent_inspections": recentInspections,
		"recent_nc":          recentNCs,
		"monthly_trend":      trend,
		"nc_by_severity":     ncBySeverity,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// QUALITY CONTROL PLANS
// ─────────────────────────────────────────────────────────────────────────────

func (h *QualityHandler) ListPlans(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			qcp.id, qcp.code, qcp.name, COALESCE(qcp.description,''),
			qcp.version,
			COALESCE(qcp.item_id::text,''), COALESCE(i.name,''),
			qcp.applies_to, qcp.is_active,
			COALESCE(u.full_name,''),
			qcp.created_at,
			(SELECT COUNT(*) FROM quality_check_templates WHERE plan_id = qcp.id) AS check_count
		FROM quality_control_plans qcp
		LEFT JOIN items i  ON i.id  = qcp.item_id
		LEFT JOIN users u  ON u.id  = qcp.created_by
		WHERE qcp.company_id = $1
		ORDER BY qcp.name`,
		companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type Plan struct {
		ID          string    `json:"id"`
		Code        string    `json:"code"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
		Version     string    `json:"version"`
		ItemID      string    `json:"item_id"`
		ItemName    string    `json:"item_name"`
		AppliesTo   string    `json:"applies_to"`
		IsActive    bool      `json:"is_active"`
		CreatedBy   string    `json:"created_by"`
		CreatedAt   time.Time `json:"created_at"`
		CheckCount  int       `json:"check_count"`
	}

	var plans []Plan
	for rows.Next() {
		var p Plan
		rows.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.Version,
			&p.ItemID, &p.ItemName, &p.AppliesTo, &p.IsActive,
			&p.CreatedBy, &p.CreatedAt, &p.CheckCount)
		plans = append(plans, p)
	}
	if plans == nil {
		plans = []Plan{}
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

func (h *QualityHandler) CreatePlan(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID, _ := c.Get("user_id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	id := uuid.New().String()
	code := qStr(body, "code")
	if code == "" {
		var seq int
		h.db.QueryRow(ctx, `SELECT COUNT(*)+1 FROM quality_control_plans WHERE company_id=$1`, companyID).Scan(&seq)
		code = fmt.Sprintf("QCP-%04d", seq)
	}

	_, err := h.db.Exec(ctx, `
		INSERT INTO quality_control_plans
			(id, company_id, code, name, description, version, item_id, applies_to, is_active, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, companyID, code,
		qStr(body, "name"),
		qStr(body, "description"),
		qStrD(body, "version", "1.0"),
		qStrN(body, "item_id"),
		qStrD(body, "applies_to", "all"),
		qBool(body, "is_active"),
		fmt.Sprintf("%v", userID),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "code": code, "message": "Plan created"})
}

func (h *QualityHandler) UpdatePlan(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err := h.db.Exec(ctx, `
		UPDATE quality_control_plans
		SET name=$1, description=$2, version=$3, item_id=$4, applies_to=$5, is_active=$6, updated_at=NOW()
		WHERE id=$7 AND company_id=$8`,
		qStr(body, "name"), qStr(body, "description"),
		qStrD(body, "version", "1.0"),
		qStrN(body, "item_id"),
		qStrD(body, "applies_to", "all"),
		qBool(body, "is_active"),
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Plan updated"})
}

func (h *QualityHandler) DeletePlan(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `DELETE FROM quality_control_plans WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Plan deleted"})
}

// ─────────────────────────────────────────────────────────────────────────────
// QUALITY INSPECTIONS
// ─────────────────────────────────────────────────────────────────────────────

func (h *QualityHandler) ListInspections(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 { page = 1 }
	if perPage < 1 || perPage > 100 { perPage = 20 }
	offset := (page - 1) * perPage

	// Filters
	statusF := c.Query("status")
	typeF   := c.Query("inspection_type")
	search  := c.Query("search")

	args := []interface{}{companyID}
	where := []string{"qi.company_id = $1"}
	argN := 2

	if statusF != "" {
		where = append(where, fmt.Sprintf("qi.status::text = $%d", argN))
		args = append(args, statusF)
		argN++
	}
	if typeF != "" {
		where = append(where, fmt.Sprintf("qi.inspection_type::text = $%d", argN))
		args = append(args, typeF)
		argN++
	}
	if search != "" {
		where = append(where, fmt.Sprintf("(qi.reference ILIKE $%d OR COALESCE(i.name,'') ILIKE $%d)", argN, argN))
		args = append(args, "%"+search+"%")
		argN++
	}

	whereSQL := strings.Join(where, " AND ")

	var total int
	h.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM quality_inspections qi
		LEFT JOIN items i ON i.id = qi.item_id
		WHERE %s`, whereSQL), args...,
	).Scan(&total)

	// Main query
	args = append(args, perPage, offset)
	rows, err := h.db.Query(ctx, fmt.Sprintf(`
		SELECT
			qi.id, qi.reference, qi.inspection_type::text, qi.status::text,
			COALESCE(qcp.name,''), COALESCE(i.name,''),
			COALESCE(qi.lot_number,''),
			qi.qty_to_inspect, qi.qty_passed, qi.qty_failed,
			COALESCE(qi.source_type,''), COALESCE(qi.source_ref,''),
			qi.scheduled_date, qi.completed_at,
			COALESCE(u.full_name,''), COALESCE(qi.overall_result::text,''),
			COALESCE(qi.notes,''), qi.created_at
		FROM quality_inspections qi
		LEFT JOIN quality_control_plans qcp ON qcp.id = qi.plan_id
		LEFT JOIN items i ON i.id = qi.item_id
		LEFT JOIN users u ON u.id = qi.inspector_id
		WHERE %s
		ORDER BY qi.created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereSQL, argN, argN+1,
	), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type Inspection struct {
		ID             string      `json:"id"`
		Reference      string      `json:"reference"`
		InspectionType string      `json:"inspection_type"`
		Status         string      `json:"status"`
		PlanName       string      `json:"plan_name"`
		ItemName       string      `json:"item_name"`
		LotNumber      string      `json:"lot_number"`
		QtyToInspect   float64     `json:"qty_to_inspect"`
		QtyPassed      float64     `json:"qty_passed"`
		QtyFailed      float64     `json:"qty_failed"`
		SourceType     string      `json:"source_type"`
		SourceRef      string      `json:"source_ref"`
		ScheduledDate  interface{} `json:"scheduled_date"`
		CompletedAt    interface{} `json:"completed_at"`
		InspectorName  string      `json:"inspector_name"`
		OverallResult  string      `json:"overall_result"`
		Notes          string      `json:"notes"`
		CreatedAt      time.Time   `json:"created_at"`
	}

	var inspections []Inspection
	for rows.Next() {
		var r Inspection
		var schedDate, completedAt interface{}
		rows.Scan(
			&r.ID, &r.Reference, &r.InspectionType, &r.Status,
			&r.PlanName, &r.ItemName, &r.LotNumber,
			&r.QtyToInspect, &r.QtyPassed, &r.QtyFailed,
			&r.SourceType, &r.SourceRef,
			&schedDate, &completedAt,
			&r.InspectorName, &r.OverallResult, &r.Notes, &r.CreatedAt,
		)
		r.ScheduledDate = schedDate
		r.CompletedAt   = completedAt
		inspections = append(inspections, r)
	}
	if inspections == nil {
		inspections = []Inspection{}
	}

	c.JSON(http.StatusOK, gin.H{
		"inspections": inspections,
		"total":       total,
		"page":        page,
		"per_page":    perPage,
	})
}

func (h *QualityHandler) GetInspection(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	var insp struct {
		ID             string      `json:"id"`
		Reference      string      `json:"reference"`
		InspectionType string      `json:"inspection_type"`
		Status         string      `json:"status"`
		PlanID         string      `json:"plan_id"`
		PlanName       string      `json:"plan_name"`
		ItemID         string      `json:"item_id"`
		ItemName       string      `json:"item_name"`
		LotNumber      string      `json:"lot_number"`
		QtyToInspect   float64     `json:"qty_to_inspect"`
		QtyPassed      float64     `json:"qty_passed"`
		QtyFailed      float64     `json:"qty_failed"`
		SourceType     string      `json:"source_type"`
		SourceID       string      `json:"source_id"`
		SourceRef      string      `json:"source_ref"`
		ScheduledDate  interface{} `json:"scheduled_date"`
		StartedAt      interface{} `json:"started_at"`
		CompletedAt    interface{} `json:"completed_at"`
		InspectorID    string      `json:"inspector_id"`
		InspectorName  string      `json:"inspector_name"`
		OverallResult  string      `json:"overall_result"`
		Notes          string      `json:"notes"`
		CreatedAt      time.Time   `json:"created_at"`
	}

	var schedDate, startedAt, completedAt interface{}
	err := h.db.QueryRow(ctx, `
		SELECT
			qi.id, qi.reference, qi.inspection_type::text, qi.status::text,
			COALESCE(qi.plan_id::text,''), COALESCE(qcp.name,''),
			COALESCE(qi.item_id::text,''), COALESCE(i.name,''),
			COALESCE(qi.lot_number,''),
			qi.qty_to_inspect, qi.qty_passed, qi.qty_failed,
			COALESCE(qi.source_type,''), COALESCE(qi.source_id::text,''), COALESCE(qi.source_ref,''),
			qi.scheduled_date, qi.started_at, qi.completed_at,
			COALESCE(qi.inspector_id::text,''), COALESCE(u.full_name,''),
			COALESCE(qi.overall_result::text,''), COALESCE(qi.notes,''),
			qi.created_at
		FROM quality_inspections qi
		LEFT JOIN quality_control_plans qcp ON qcp.id = qi.plan_id
		LEFT JOIN items i ON i.id = qi.item_id
		LEFT JOIN users u ON u.id = qi.inspector_id
		WHERE qi.id = $1 AND qi.company_id = $2`,
		id, companyID,
	).Scan(
		&insp.ID, &insp.Reference, &insp.InspectionType, &insp.Status,
		&insp.PlanID, &insp.PlanName,
		&insp.ItemID, &insp.ItemName,
		&insp.LotNumber,
		&insp.QtyToInspect, &insp.QtyPassed, &insp.QtyFailed,
		&insp.SourceType, &insp.SourceID, &insp.SourceRef,
		&schedDate, &startedAt, &completedAt,
		&insp.InspectorID, &insp.InspectorName,
		&insp.OverallResult, &insp.Notes, &insp.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inspection not found"})
		return
	}
	insp.ScheduledDate = schedDate
	insp.StartedAt     = startedAt
	insp.CompletedAt   = completedAt

	// Load checks
	type Check struct {
		ID            string      `json:"id"`
		Sequence      int         `json:"sequence"`
		Name          string      `json:"name"`
		CheckType     string      `json:"check_type"`
		Unit          string      `json:"unit"`
		MinValue      interface{} `json:"min_value"`
		MaxValue      interface{} `json:"max_value"`
		MeasuredValue interface{} `json:"measured_value"`
		Result        string      `json:"result"`
		IsMandatory   bool        `json:"is_mandatory"`
		NormReference string      `json:"norm_reference"`
		Notes         string      `json:"notes"`
		CheckedBy     string      `json:"checked_by"`
		CheckedAt     interface{} `json:"checked_at"`
	}
	var checks []Check
	chkRows, err2 := h.db.Query(ctx, `
		SELECT
			qc.id, qc.sequence, qc.name, qc.check_type,
			COALESCE(qc.unit,''), qc.min_value, qc.max_value, qc.measured_value,
			COALESCE(qc.result::text,''), qc.is_mandatory,
			COALESCE(qc.norm_reference,''), COALESCE(qc.notes,''),
			COALESCE(u.full_name,''), qc.checked_at
		FROM quality_checks qc
		LEFT JOIN users u ON u.id = qc.checked_by
		WHERE qc.inspection_id = $1
		ORDER BY qc.sequence`, id,
	)
	if err2 == nil {
		defer chkRows.Close()
		for chkRows.Next() {
			var ch Check
			chkRows.Scan(
				&ch.ID, &ch.Sequence, &ch.Name, &ch.CheckType,
				&ch.Unit, &ch.MinValue, &ch.MaxValue, &ch.MeasuredValue,
				&ch.Result, &ch.IsMandatory,
				&ch.NormReference, &ch.Notes,
				&ch.CheckedBy, &ch.CheckedAt,
			)
			checks = append(checks, ch)
		}
	}
	if checks == nil {
		checks = []Check{}
	}

	c.JSON(http.StatusOK, gin.H{"inspection": insp, "checks": checks})
}

func (h *QualityHandler) CreateInspection(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID, _ := c.Get("user_id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	id := uuid.New().String()

	// Generate reference
	ref := qStr(body, "reference")
	if ref == "" {
		var seq int
		h.db.QueryRow(ctx, `SELECT COUNT(*)+1 FROM quality_inspections WHERE company_id=$1`, companyID).Scan(&seq)
		ref = fmt.Sprintf("QI-%s-%05d", time.Now().Format("2006"), seq)
	}

	inspType := qStrD(body, "inspection_type", "incoming")

	_, err := h.db.Exec(ctx, `
		INSERT INTO quality_inspections
			(id, company_id, reference, inspection_type, status, plan_id, item_id,
			 lot_number, qty_to_inspect, qty_passed, qty_failed,
			 source_type, source_id, source_ref,
			 scheduled_date, inspector_id, notes, created_by)
		VALUES ($1,$2,$3,$4::inspection_type,'pending',$5,$6,$7,$8,0,0,$9,$10,$11,$12,$13,$14,$15)`,
		id, companyID, ref, inspType,
		qStrN(body, "plan_id"),
		qStrN(body, "item_id"),
		qStr(body, "lot_number"),
		qF64(body, "qty_to_inspect"),
		qStr(body, "source_type"),
		qStrN(body, "source_id"),
		qStr(body, "source_ref"),
		qStrN(body, "scheduled_date"),
		qStrN(body, "inspector_id"),
		qStr(body, "notes"),
		fmt.Sprintf("%v", userID),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If plan_id provided, auto-create checks from templates
	if planID := qStr(body, "plan_id"); planID != "" {
		tplRows, tplErr := h.db.Query(ctx, `
			SELECT id, sequence, name, COALESCE(description,''), check_type,
			       COALESCE(unit,''), min_value, max_value,
			       COALESCE(norm_reference,''), COALESCE(instructions,''), is_mandatory
			FROM quality_check_templates
			WHERE plan_id = $1
			ORDER BY sequence`, planID,
		)
		if tplErr == nil {
			defer tplRows.Close()
			for tplRows.Next() {
				var tplID, name, desc, chkType, unit, normRef, instructions string
				var seq int
				var isMandatory bool
				var minVal, maxVal interface{}
				tplRows.Scan(&tplID, &seq, &name, &desc, &chkType, &unit, &minVal, &maxVal, &normRef, &instructions, &isMandatory)
				chkID := uuid.New().String()
				h.db.Exec(ctx, `
					INSERT INTO quality_checks
						(id, inspection_id, template_id, sequence, name, description,
						 check_type, unit, min_value, max_value, norm_reference,
						 instructions, is_mandatory)
					VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
					chkID, id, tplID, seq, name, desc, chkType, unit, minVal, maxVal, normRef, instructions, isMandatory,
				)
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "reference": ref, "message": "Inspection created"})
}

func (h *QualityHandler) UpdateInspection(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err := h.db.Exec(ctx, `
		UPDATE quality_inspections SET
			inspection_type = $1::inspection_type,
			plan_id         = $2,
			item_id         = $3,
			lot_number      = $4,
			qty_to_inspect  = $5,
			source_type     = $6,
			source_ref      = $7,
			scheduled_date  = $8,
			inspector_id    = $9,
			notes           = $10,
			updated_at      = NOW()
		WHERE id = $11 AND company_id = $12`,
		qStrD(body, "inspection_type", "incoming"),
		qStrN(body, "plan_id"),
		qStrN(body, "item_id"),
		qStr(body, "lot_number"),
		qF64(body, "qty_to_inspect"),
		qStr(body, "source_type"),
		qStr(body, "source_ref"),
		qStrN(body, "scheduled_date"),
		qStrN(body, "inspector_id"),
		qStr(body, "notes"),
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Inspection updated"})
}

func (h *QualityHandler) StartInspection(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		UPDATE quality_inspections
		SET status='in_progress'::inspection_status, started_at=NOW(), updated_at=NOW()
		WHERE id=$1 AND company_id=$2 AND status='pending'::inspection_status`,
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Inspection started"})
}

func (h *QualityHandler) CompleteInspection(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result := qStrD(body, "overall_result", "pass")
	var finalStatus string
	switch result {
	case "pass":
		finalStatus = "passed"
	case "fail":
		finalStatus = "failed"
	default:
		finalStatus = "passed"
	}

	_, err := h.db.Exec(ctx, `
		UPDATE quality_inspections SET
			status         = $1::inspection_status,
			overall_result = $2::check_result,
			qty_passed     = $3,
			qty_failed     = $4,
			completed_at   = NOW(),
			approved_by    = $5,
			approved_at    = NOW(),
			notes          = COALESCE(NULLIF($6,''), notes),
			updated_at     = NOW()
		WHERE id=$7 AND company_id=$8`,
		finalStatus, result,
		qF64(body, "qty_passed"),
		qF64(body, "qty_failed"),
		fmt.Sprintf("%v", userID),
		qStr(body, "notes"),
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Inspection completed", "status": finalStatus})
}

func (h *QualityHandler) DeleteInspection(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	// Only allow deleting pending inspections
	_, err := h.db.Exec(ctx, `
		DELETE FROM quality_inspections
		WHERE id=$1 AND company_id=$2 AND status='pending'::inspection_status`,
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Inspection deleted"})
}

// ─────────────────────────────────────────────────────────────────────────────
// QUALITY CHECKS
// ─────────────────────────────────────────────────────────────────────────────

func (h *QualityHandler) ListChecks(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "25"))
	if page < 1 { page = 1 }
	if perPage < 1 || perPage > 100 { perPage = 25 }
	offset := (page - 1) * perPage

	inspectionID := c.Query("inspection_id")
	resultF      := c.Query("result")

	args := []interface{}{companyID}
	where := []string{"qi.company_id = $1"}
	argN := 2

	if inspectionID != "" {
		where = append(where, fmt.Sprintf("qc.inspection_id = $%d", argN))
		args = append(args, inspectionID)
		argN++
	}
	if resultF != "" {
		where = append(where, fmt.Sprintf("qc.result::text = $%d", argN))
		args = append(args, resultF)
		argN++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	h.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) FROM quality_checks qc
		JOIN quality_inspections qi ON qi.id = qc.inspection_id
		WHERE %s`, whereSQL), args...,
	).Scan(&total)

	args = append(args, perPage, offset)
	rows, err := h.db.Query(ctx, fmt.Sprintf(`
		SELECT
			qc.id, qc.inspection_id, qi.reference,
			qc.sequence, qc.name, qc.check_type,
			COALESCE(qc.unit,''),
			qc.min_value, qc.max_value, qc.measured_value,
			COALESCE(qc.result::text,''), qc.is_mandatory,
			COALESCE(qc.norm_reference,''), COALESCE(qc.notes,''),
			COALESCE(u.full_name,''), qc.checked_at,
			qc.created_at
		FROM quality_checks qc
		JOIN quality_inspections qi ON qi.id = qc.inspection_id
		LEFT JOIN users u ON u.id = qc.checked_by
		WHERE %s
		ORDER BY qi.created_at DESC, qc.sequence
		LIMIT $%d OFFSET $%d`,
		whereSQL, argN, argN+1,
	), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type Check struct {
		ID             string      `json:"id"`
		InspectionID   string      `json:"inspection_id"`
		InspectionRef  string      `json:"inspection_ref"`
		Sequence       int         `json:"sequence"`
		Name           string      `json:"name"`
		CheckType      string      `json:"check_type"`
		Unit           string      `json:"unit"`
		MinValue       interface{} `json:"min_value"`
		MaxValue       interface{} `json:"max_value"`
		MeasuredValue  interface{} `json:"measured_value"`
		Result         string      `json:"result"`
		IsMandatory    bool        `json:"is_mandatory"`
		NormReference  string      `json:"norm_reference"`
		Notes          string      `json:"notes"`
		CheckedBy      string      `json:"checked_by"`
		CheckedAt      interface{} `json:"checked_at"`
		CreatedAt      time.Time   `json:"created_at"`
	}

	var checks []Check
	for rows.Next() {
		var ch Check
		rows.Scan(
			&ch.ID, &ch.InspectionID, &ch.InspectionRef,
			&ch.Sequence, &ch.Name, &ch.CheckType,
			&ch.Unit, &ch.MinValue, &ch.MaxValue, &ch.MeasuredValue,
			&ch.Result, &ch.IsMandatory, &ch.NormReference, &ch.Notes,
			&ch.CheckedBy, &ch.CheckedAt, &ch.CreatedAt,
		)
		checks = append(checks, ch)
	}
	if checks == nil {
		checks = []Check{}
	}

	c.JSON(http.StatusOK, gin.H{
		"checks":   checks,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func (h *QualityHandler) CreateCheck(c *gin.Context) {
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	id := uuid.New().String()
	_, err := h.db.Exec(ctx, `
		INSERT INTO quality_checks
			(id, inspection_id, sequence, name, description, check_type,
			 unit, min_value, max_value, norm_reference, instructions, is_mandatory)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id,
		qStr(body, "inspection_id"),
		qInt(body, "sequence"),
		qStr(body, "name"),
		qStr(body, "description"),
		qStrD(body, "check_type", "visual"),
		qStr(body, "unit"),
		qStrN(body, "min_value"),
		qStrN(body, "max_value"),
		qStr(body, "norm_reference"),
		qStr(body, "instructions"),
		qBool(body, "is_mandatory"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Check created"})
}

func (h *QualityHandler) RecordCheckResult(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result := qStrD(body, "result", "pass")

	_, err := h.db.Exec(ctx, `
		UPDATE quality_checks SET
			result         = $1::check_result,
			measured_value = $2,
			notes          = $3,
			checked_by     = $4,
			checked_at     = NOW(),
			updated_at     = NOW()
		WHERE id = $5`,
		result,
		qStrN(body, "measured_value"),
		qStr(body, "notes"),
		fmt.Sprintf("%v", userID),
		id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Recalculate inspection overall result
	var inspID string
	h.db.QueryRow(ctx, `SELECT inspection_id FROM quality_checks WHERE id=$1`, id).Scan(&inspID)
	if inspID != "" {
		var passCount, failCount, pendingCount int
		h.db.QueryRow(ctx, `
			SELECT
				COUNT(*) FILTER (WHERE result='pass'),
				COUNT(*) FILTER (WHERE result='fail'),
				COUNT(*) FILTER (WHERE result IS NULL AND is_mandatory)
			FROM quality_checks
			WHERE inspection_id=$1`, inspID,
		).Scan(&passCount, &failCount, &pendingCount)

		if failCount > 0 {
			h.db.Exec(ctx, `UPDATE quality_inspections SET overall_result='fail'::check_result, updated_at=NOW() WHERE id=$1`, inspID)
		} else if pendingCount == 0 {
			h.db.Exec(ctx, `UPDATE quality_inspections SET overall_result='pass'::check_result, updated_at=NOW() WHERE id=$1`, inspID)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Check result recorded"})
}

func (h *QualityHandler) DeleteCheck(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `DELETE FROM quality_checks WHERE id=$1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Check deleted"})
}

// ─────────────────────────────────────────────────────────────────────────────
// NON-CONFORMITIES
// ─────────────────────────────────────────────────────────────────────────────

func (h *QualityHandler) ListNonConformities(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 { page = 1 }
	if perPage < 1 || perPage > 100 { perPage = 20 }
	offset := (page - 1) * perPage

	statusF   := c.Query("status")
	severityF := c.Query("severity")
	search    := c.Query("search")

	args := []interface{}{companyID}
	where := []string{"nc.company_id = $1"}
	argN := 2

	if statusF != "" {
		where = append(where, fmt.Sprintf("nc.status::text = $%d", argN))
		args = append(args, statusF)
		argN++
	}
	if severityF != "" {
		where = append(where, fmt.Sprintf("nc.severity::text = $%d", argN))
		args = append(args, severityF)
		argN++
	}
	if search != "" {
		where = append(where, fmt.Sprintf("(nc.reference ILIKE $%d OR nc.title ILIKE $%d)", argN, argN))
		args = append(args, "%"+search+"%")
		argN++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	h.db.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM non_conformities nc WHERE %s`, whereSQL), args...).Scan(&total)

	args = append(args, perPage, offset)
	rows, err := h.db.Query(ctx, fmt.Sprintf(`
		SELECT
			nc.id, nc.reference, nc.title, nc.status::text, nc.severity::text,
			COALESCE(nc.source_type,''), COALESCE(nc.source_ref,''),
			COALESCE(i.name,''), COALESCE(nc.lot_number,''),
			nc.qty_affected,
			COALESCE(d.name,''),
			COALESCE(ub.full_name,'') AS detected_by_name,
			nc.detected_date,
			COALESCE(ua.full_name,'') AS assigned_to_name,
			nc.target_date,
			COALESCE(nc.root_cause,''),
			COALESCE(nc.immediate_action,''),
			nc.closed_at,
			COALESCE(nc.closure_notes,''),
			nc.created_at,
			(SELECT COUNT(*) FROM corrective_actions WHERE nc_id = nc.id) AS ca_count
		FROM non_conformities nc
		LEFT JOIN items i        ON i.id  = nc.item_id
		LEFT JOIN departments d  ON d.id  = nc.department_id
		LEFT JOIN users ub       ON ub.id = nc.detected_by
		LEFT JOIN users ua       ON ua.id = nc.assigned_to
		WHERE %s
		ORDER BY
			CASE nc.severity
				WHEN 'critical_safety' THEN 1
				WHEN 'critical' THEN 2
				WHEN 'major' THEN 3
				ELSE 4
			END,
			nc.created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereSQL, argN, argN+1,
	), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type NC struct {
		ID               string      `json:"id"`
		Reference        string      `json:"reference"`
		Title            string      `json:"title"`
		Status           string      `json:"status"`
		Severity         string      `json:"severity"`
		SourceType       string      `json:"source_type"`
		SourceRef        string      `json:"source_ref"`
		ItemName         string      `json:"item_name"`
		LotNumber        string      `json:"lot_number"`
		QtyAffected      interface{} `json:"qty_affected"`
		DepartmentName   string      `json:"department_name"`
		DetectedByName   string      `json:"detected_by_name"`
		DetectedDate     time.Time   `json:"detected_date"`
		AssignedToName   string      `json:"assigned_to_name"`
		TargetDate       interface{} `json:"target_date"`
		RootCause        string      `json:"root_cause"`
		ImmediateAction  string      `json:"immediate_action"`
		ClosedAt         interface{} `json:"closed_at"`
		ClosureNotes     string      `json:"closure_notes"`
		CreatedAt        time.Time   `json:"created_at"`
		CACount          int         `json:"ca_count"`
	}

	var ncs []NC
	for rows.Next() {
		var r NC
		rows.Scan(
			&r.ID, &r.Reference, &r.Title, &r.Status, &r.Severity,
			&r.SourceType, &r.SourceRef,
			&r.ItemName, &r.LotNumber, &r.QtyAffected,
			&r.DepartmentName,
			&r.DetectedByName, &r.DetectedDate,
			&r.AssignedToName, &r.TargetDate,
			&r.RootCause, &r.ImmediateAction,
			&r.ClosedAt, &r.ClosureNotes,
			&r.CreatedAt, &r.CACount,
		)
		ncs = append(ncs, r)
	}
	if ncs == nil {
		ncs = []NC{}
	}

	c.JSON(http.StatusOK, gin.H{
		"non_conformities": ncs,
		"total":            total,
		"page":             page,
		"per_page":         perPage,
	})
}

func (h *QualityHandler) GetNonConformity(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	type NC struct {
		ID              string      `json:"id"`
		Reference       string      `json:"reference"`
		Title           string      `json:"title"`
		Description     string      `json:"description"`
		Status          string      `json:"status"`
		Severity        string      `json:"severity"`
		SourceType      string      `json:"source_type"`
		SourceID        string      `json:"source_id"`
		SourceRef       string      `json:"source_ref"`
		InspectionID    string      `json:"inspection_id"`
		ItemID          string      `json:"item_id"`
		ItemName        string      `json:"item_name"`
		LotNumber       string      `json:"lot_number"`
		QtyAffected     interface{} `json:"qty_affected"`
		DepartmentID    string      `json:"department_id"`
		DepartmentName  string      `json:"department_name"`
		Process         string      `json:"process"`
		DetectedBy      string      `json:"detected_by"`
		DetectedByName  string      `json:"detected_by_name"`
		DetectedDate    time.Time   `json:"detected_date"`
		AssignedTo      string      `json:"assigned_to"`
		AssignedToName  string      `json:"assigned_to_name"`
		TargetDate      interface{} `json:"target_date"`
		RootCause       string      `json:"root_cause"`
		ImmediateAction string      `json:"immediate_action"`
		ClosedAt        interface{} `json:"closed_at"`
		ClosureNotes    string      `json:"closure_notes"`
		CreatedAt       time.Time   `json:"created_at"`
	}

	var nc NC
	err := h.db.QueryRow(ctx, `
		SELECT
			nc.id, nc.reference, nc.title, COALESCE(nc.description,''),
			nc.status::text, nc.severity::text,
			COALESCE(nc.source_type,''), COALESCE(nc.source_id::text,''), COALESCE(nc.source_ref,''),
			COALESCE(nc.inspection_id::text,''),
			COALESCE(nc.item_id::text,''), COALESCE(i.name,''),
			COALESCE(nc.lot_number,''), nc.qty_affected,
			COALESCE(nc.department_id::text,''), COALESCE(d.name,''),
			COALESCE(nc.process,''),
			COALESCE(nc.detected_by::text,''), COALESCE(ub.full_name,''),
			nc.detected_date,
			COALESCE(nc.assigned_to::text,''), COALESCE(ua.full_name,''),
			nc.target_date,
			COALESCE(nc.root_cause,''), COALESCE(nc.immediate_action,''),
			nc.closed_at, COALESCE(nc.closure_notes,''),
			nc.created_at
		FROM non_conformities nc
		LEFT JOIN items i       ON i.id  = nc.item_id
		LEFT JOIN departments d ON d.id  = nc.department_id
		LEFT JOIN users ub      ON ub.id = nc.detected_by
		LEFT JOIN users ua      ON ua.id = nc.assigned_to
		WHERE nc.id=$1 AND nc.company_id=$2`,
		id, companyID,
	).Scan(
		&nc.ID, &nc.Reference, &nc.Title, &nc.Description,
		&nc.Status, &nc.Severity,
		&nc.SourceType, &nc.SourceID, &nc.SourceRef,
		&nc.InspectionID,
		&nc.ItemID, &nc.ItemName, &nc.LotNumber, &nc.QtyAffected,
		&nc.DepartmentID, &nc.DepartmentName, &nc.Process,
		&nc.DetectedBy, &nc.DetectedByName, &nc.DetectedDate,
		&nc.AssignedTo, &nc.AssignedToName, &nc.TargetDate,
		&nc.RootCause, &nc.ImmediateAction,
		&nc.ClosedAt, &nc.ClosureNotes, &nc.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Non-conformity not found"})
		return
	}

	// Load linked corrective actions
	type CA struct {
		ID        string `json:"id"`
		Reference string `json:"reference"`
		Title     string `json:"title"`
		Status    string `json:"status"`
		CAType    string `json:"ca_type"`
	}
	var cas []CA
	caRows, caErr := h.db.Query(ctx, `
		SELECT id, reference, title, status::text, ca_type::text
		FROM corrective_actions
		WHERE nc_id = $1
		ORDER BY created_at`, id,
	)
	if caErr == nil {
		defer caRows.Close()
		for caRows.Next() {
			var ca CA
			caRows.Scan(&ca.ID, &ca.Reference, &ca.Title, &ca.Status, &ca.CAType)
			cas = append(cas, ca)
		}
	}
	if cas == nil {
		cas = []CA{}
	}

	c.JSON(http.StatusOK, gin.H{"non_conformity": nc, "corrective_actions": cas})
}

func (h *QualityHandler) CreateNonConformity(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID, _ := c.Get("user_id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	id := uuid.New().String()
	ref := qStr(body, "reference")
	if ref == "" {
		var seq int
		h.db.QueryRow(ctx, `SELECT COUNT(*)+1 FROM non_conformities WHERE company_id=$1`, companyID).Scan(&seq)
		ref = fmt.Sprintf("NC-%s-%05d", time.Now().Format("2006"), seq)
	}

	_, err := h.db.Exec(ctx, `
		INSERT INTO non_conformities
			(id, company_id, reference, title, description, status, severity,
			 source_type, source_id, source_ref, inspection_id,
			 item_id, lot_number, qty_affected,
			 department_id, process,
			 detected_by, detected_date,
			 assigned_to, target_date,
			 root_cause, immediate_action, created_by)
		VALUES ($1,$2,$3,$4,$5,'open'::nc_status,$6::nc_severity,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
		id, companyID, ref,
		qStr(body, "title"),
		qStr(body, "description"),
		qStrD(body, "severity", "minor"),
		qStr(body, "source_type"),
		qStrN(body, "source_id"),
		qStr(body, "source_ref"),
		qStrN(body, "inspection_id"),
		qStrN(body, "item_id"),
		qStr(body, "lot_number"),
		qStrN(body, "qty_affected"),
		qStrN(body, "department_id"),
		qStr(body, "process"),
		qStrN(body, "detected_by"),
		qStrD(body, "detected_date", time.Now().Format("2006-01-02")),
		qStrN(body, "assigned_to"),
		qStrN(body, "target_date"),
		qStr(body, "root_cause"),
		qStr(body, "immediate_action"),
		fmt.Sprintf("%v", userID),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "reference": ref, "message": "Non-conformity created"})
}

func (h *QualityHandler) UpdateNonConformity(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err := h.db.Exec(ctx, `
		UPDATE non_conformities SET
			title            = $1,
			description      = $2,
			severity         = $3::nc_severity,
			source_type      = $4,
			source_ref       = $5,
			item_id          = $6,
			lot_number       = $7,
			qty_affected     = $8,
			department_id    = $9,
			process          = $10,
			assigned_to      = $11,
			target_date      = $12,
			root_cause       = $13,
			immediate_action = $14,
			updated_at       = NOW()
		WHERE id=$15 AND company_id=$16`,
		qStr(body, "title"),
		qStr(body, "description"),
		qStrD(body, "severity", "minor"),
		qStr(body, "source_type"),
		qStr(body, "source_ref"),
		qStrN(body, "item_id"),
		qStr(body, "lot_number"),
		qStrN(body, "qty_affected"),
		qStrN(body, "department_id"),
		qStr(body, "process"),
		qStrN(body, "assigned_to"),
		qStrN(body, "target_date"),
		qStr(body, "root_cause"),
		qStr(body, "immediate_action"),
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Non-conformity updated"})
}

func (h *QualityHandler) UpdateNCStatus(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	newStatus := qStr(body, "status")

	var execErr error
	if newStatus == "closed" {
		_, execErr = h.db.Exec(ctx, `
			UPDATE non_conformities SET
				status        = $1::nc_status,
				closed_by     = $2,
				closed_at     = NOW(),
				closure_notes = $3,
				updated_at    = NOW()
			WHERE id=$4 AND company_id=$5`,
			newStatus,
			fmt.Sprintf("%v", userID),
			qStr(body, "closure_notes"),
			id, companyID,
		)
	} else {
		_, execErr = h.db.Exec(ctx, `
			UPDATE non_conformities SET status=$1::nc_status, updated_at=NOW()
			WHERE id=$2 AND company_id=$3`,
			newStatus, id, companyID,
		)
	}
	if execErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": execErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status updated", "status": newStatus})
}

func (h *QualityHandler) DeleteNonConformity(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		DELETE FROM non_conformities WHERE id=$1 AND company_id=$2 AND status='open'::nc_status`,
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Non-conformity deleted"})
}

// ─────────────────────────────────────────────────────────────────────────────
// CORRECTIVE ACTIONS
// ─────────────────────────────────────────────────────────────────────────────

func (h *QualityHandler) ListCorrectiveActions(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 { page = 1 }
	if perPage < 1 || perPage > 100 { perPage = 20 }
	offset := (page - 1) * perPage

	statusF   := c.Query("status")
	caTypeF   := c.Query("ca_type")
	priorityF := c.Query("priority")
	search    := c.Query("search")
	ncID      := c.Query("nc_id")

	args := []interface{}{companyID}
	where := []string{"ca.company_id = $1"}
	argN := 2

	if statusF != "" {
		where = append(where, fmt.Sprintf("ca.status::text = $%d", argN))
		args = append(args, statusF)
		argN++
	}
	if caTypeF != "" {
		where = append(where, fmt.Sprintf("ca.ca_type::text = $%d", argN))
		args = append(args, caTypeF)
		argN++
	}
	if priorityF != "" {
		where = append(where, fmt.Sprintf("ca.priority = $%d", argN))
		args = append(args, priorityF)
		argN++
	}
	if ncID != "" {
		where = append(where, fmt.Sprintf("ca.nc_id = $%d", argN))
		args = append(args, ncID)
		argN++
	}
	if search != "" {
		where = append(where, fmt.Sprintf("(ca.reference ILIKE $%d OR ca.title ILIKE $%d)", argN, argN))
		args = append(args, "%"+search+"%")
		argN++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	h.db.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM corrective_actions ca WHERE %s`, whereSQL), args...).Scan(&total)

	args = append(args, perPage, offset)
	rows, err := h.db.Query(ctx, fmt.Sprintf(`
		SELECT
			ca.id, ca.reference, ca.title, ca.ca_type::text, ca.status::text, ca.priority,
			COALESCE(nc.reference,'') AS nc_ref, COALESCE(nc.title,'') AS nc_title,
			COALESCE(nc.severity::text,'') AS nc_severity,
			COALESCE(ur.full_name,'') AS responsible_name,
			COALESCE(d.name,'') AS dept_name,
			ca.due_date, ca.implementation_date,
			COALESCE(ca.proposed_action,''),
			COALESCE(ca.implemented_action,''),
			ca.effectiveness_rating,
			ca.estimated_cost, ca.actual_cost,
			ca.closed_at,
			ca.created_at,
			(ca.due_date < CURRENT_DATE AND ca.status NOT IN ('closed','cancelled','verified')) AS is_overdue,
			(SELECT COUNT(*) FROM ca_tasks WHERE ca_id = ca.id AND completed = false) AS pending_tasks
		FROM corrective_actions ca
		LEFT JOIN non_conformities nc ON nc.id  = ca.nc_id
		LEFT JOIN users ur            ON ur.id  = ca.responsible_id
		LEFT JOIN departments d       ON d.id   = ca.department_id
		WHERE %s
		ORDER BY
			CASE ca.priority WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 ELSE 4 END,
			ca.due_date ASC NULLS LAST,
			ca.created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereSQL, argN, argN+1,
	), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type CA struct {
		ID                 string      `json:"id"`
		Reference          string      `json:"reference"`
		Title              string      `json:"title"`
		CAType             string      `json:"ca_type"`
		Status             string      `json:"status"`
		Priority           string      `json:"priority"`
		NCRef              string      `json:"nc_ref"`
		NCTitle            string      `json:"nc_title"`
		NCSeverity         string      `json:"nc_severity"`
		ResponsibleName    string      `json:"responsible_name"`
		DeptName           string      `json:"dept_name"`
		DueDate            interface{} `json:"due_date"`
		ImplementationDate interface{} `json:"implementation_date"`
		ProposedAction     string      `json:"proposed_action"`
		ImplementedAction  string      `json:"implemented_action"`
		EffectivenessRating interface{} `json:"effectiveness_rating"`
		EstimatedCost      float64     `json:"estimated_cost"`
		ActualCost         float64     `json:"actual_cost"`
		ClosedAt           interface{} `json:"closed_at"`
		CreatedAt          time.Time   `json:"created_at"`
		IsOverdue          bool        `json:"is_overdue"`
		PendingTasks       int         `json:"pending_tasks"`
	}

	var cas []CA
	for rows.Next() {
		var ca CA
		rows.Scan(
			&ca.ID, &ca.Reference, &ca.Title, &ca.CAType, &ca.Status, &ca.Priority,
			&ca.NCRef, &ca.NCTitle, &ca.NCSeverity,
			&ca.ResponsibleName, &ca.DeptName,
			&ca.DueDate, &ca.ImplementationDate,
			&ca.ProposedAction, &ca.ImplementedAction,
			&ca.EffectivenessRating,
			&ca.EstimatedCost, &ca.ActualCost,
			&ca.ClosedAt, &ca.CreatedAt, &ca.IsOverdue, &ca.PendingTasks,
		)
		cas = append(cas, ca)
	}
	if cas == nil {
		cas = []CA{}
	}

	c.JSON(http.StatusOK, gin.H{
		"corrective_actions": cas,
		"total":              total,
		"page":               page,
		"per_page":           perPage,
	})
}

func (h *QualityHandler) GetCorrectiveAction(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	type CA struct {
		ID                  string      `json:"id"`
		Reference           string      `json:"reference"`
		Title               string      `json:"title"`
		Description         string      `json:"description"`
		CAType              string      `json:"ca_type"`
		Status              string      `json:"status"`
		Priority            string      `json:"priority"`
		NCID                string      `json:"nc_id"`
		NCRef               string      `json:"nc_ref"`
		RootCause           string      `json:"root_cause"`
		RootCauseMethod     string      `json:"root_cause_method"`
		ProposedAction      string      `json:"proposed_action"`
		ImplementedAction   string      `json:"implemented_action"`
		ResponsibleID       string      `json:"responsible_id"`
		ResponsibleName     string      `json:"responsible_name"`
		DepartmentID        string      `json:"department_id"`
		DeptName            string      `json:"dept_name"`
		DueDate             interface{} `json:"due_date"`
		ImplementationDate  interface{} `json:"implementation_date"`
		VerificationDate    interface{} `json:"verification_date"`
		EffectivenessRating interface{} `json:"effectiveness_rating"`
		EffectivenessNotes  string      `json:"effectiveness_notes"`
		VerifiedBy          string      `json:"verified_by"`
		EstimatedCost       float64     `json:"estimated_cost"`
		ActualCost          float64     `json:"actual_cost"`
		ClosedAt            interface{} `json:"closed_at"`
		CreatedAt           time.Time   `json:"created_at"`
	}

	var ca CA
	err := h.db.QueryRow(ctx, `
		SELECT
			ca.id, ca.reference, ca.title, COALESCE(ca.description,''),
			ca.ca_type::text, ca.status::text, ca.priority,
			COALESCE(ca.nc_id::text,''), COALESCE(nc.reference,''),
			COALESCE(ca.root_cause,''), COALESCE(ca.root_cause_method,'5why'),
			COALESCE(ca.proposed_action,''), COALESCE(ca.implemented_action,''),
			COALESCE(ca.responsible_id::text,''), COALESCE(ur.full_name,''),
			COALESCE(ca.department_id::text,''), COALESCE(d.name,''),
			ca.due_date, ca.implementation_date, ca.verification_date,
			ca.effectiveness_rating, COALESCE(ca.effectiveness_notes,''),
			COALESCE(uv.full_name,''),
			ca.estimated_cost, ca.actual_cost,
			ca.closed_at, ca.created_at
		FROM corrective_actions ca
		LEFT JOIN non_conformities nc ON nc.id = ca.nc_id
		LEFT JOIN users ur            ON ur.id = ca.responsible_id
		LEFT JOIN departments d       ON d.id  = ca.department_id
		LEFT JOIN users uv            ON uv.id = ca.verified_by
		WHERE ca.id=$1 AND ca.company_id=$2`,
		id, companyID,
	).Scan(
		&ca.ID, &ca.Reference, &ca.Title, &ca.Description,
		&ca.CAType, &ca.Status, &ca.Priority,
		&ca.NCID, &ca.NCRef,
		&ca.RootCause, &ca.RootCauseMethod,
		&ca.ProposedAction, &ca.ImplementedAction,
		&ca.ResponsibleID, &ca.ResponsibleName,
		&ca.DepartmentID, &ca.DeptName,
		&ca.DueDate, &ca.ImplementationDate, &ca.VerificationDate,
		&ca.EffectivenessRating, &ca.EffectivenessNotes,
		&ca.VerifiedBy,
		&ca.EstimatedCost, &ca.ActualCost,
		&ca.ClosedAt, &ca.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Corrective action not found"})
		return
	}

	// Load tasks
	type Task struct {
		ID          string      `json:"id"`
		Sequence    int         `json:"sequence"`
		Description string      `json:"description"`
		AssignedTo  string      `json:"assigned_to"`
		DueDate     interface{} `json:"due_date"`
		Completed   bool        `json:"completed"`
		CompletedAt interface{} `json:"completed_at"`
		Notes       string      `json:"notes"`
	}
	var tasks []Task
	taskRows, taskErr := h.db.Query(ctx, `
		SELECT t.id, t.sequence, t.description,
		       COALESCE(u.full_name,''), t.due_date, t.completed, t.completed_at,
		       COALESCE(t.notes,'')
		FROM ca_tasks t
		LEFT JOIN users u ON u.id = t.assigned_to
		WHERE t.ca_id = $1
		ORDER BY t.sequence`, id,
	)
	if taskErr == nil {
		defer taskRows.Close()
		for taskRows.Next() {
			var t Task
			taskRows.Scan(&t.ID, &t.Sequence, &t.Description,
				&t.AssignedTo, &t.DueDate, &t.Completed, &t.CompletedAt, &t.Notes)
			tasks = append(tasks, t)
		}
	}
	if tasks == nil {
		tasks = []Task{}
	}

	c.JSON(http.StatusOK, gin.H{"corrective_action": ca, "tasks": tasks})
}

func (h *QualityHandler) CreateCorrectiveAction(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID, _ := c.Get("user_id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	id := uuid.New().String()
	ref := qStr(body, "reference")
	if ref == "" {
		var seq int
		h.db.QueryRow(ctx, `SELECT COUNT(*)+1 FROM corrective_actions WHERE company_id=$1`, companyID).Scan(&seq)
		ref = fmt.Sprintf("CA-%s-%05d", time.Now().Format("2006"), seq)
	}

	_, err := h.db.Exec(ctx, `
		INSERT INTO corrective_actions
			(id, company_id, reference, title, description,
			 ca_type, status, priority, nc_id,
			 root_cause, root_cause_method,
			 proposed_action,
			 responsible_id, department_id,
			 due_date, estimated_cost, created_by)
		VALUES ($1,$2,$3,$4,$5,$6::ca_type,'open'::ca_status,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		id, companyID, ref,
		qStr(body, "title"),
		qStr(body, "description"),
		qStrD(body, "ca_type", "corrective"),
		qStrD(body, "priority", "medium"),
		qStrN(body, "nc_id"),
		qStr(body, "root_cause"),
		qStrD(body, "root_cause_method", "5why"),
		qStr(body, "proposed_action"),
		qStrN(body, "responsible_id"),
		qStrN(body, "department_id"),
		qStrN(body, "due_date"),
		qF64(body, "estimated_cost"),
		fmt.Sprintf("%v", userID),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update NC status to corrective_action if nc_id provided
	if ncID := qStr(body, "nc_id"); ncID != "" {
		h.db.Exec(ctx, `
			UPDATE non_conformities
			SET status='corrective_action'::nc_status, updated_at=NOW()
			WHERE id=$1 AND company_id=$2 AND status='open'::nc_status`,
			ncID, companyID,
		)
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "reference": ref, "message": "Corrective action created"})
}

func (h *QualityHandler) UpdateCorrectiveAction(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err := h.db.Exec(ctx, `
		UPDATE corrective_actions SET
			title               = $1,
			description         = $2,
			ca_type             = $3::ca_type,
			priority            = $4,
			root_cause          = $5,
			root_cause_method   = $6,
			proposed_action     = $7,
			implemented_action  = $8,
			responsible_id      = $9,
			department_id       = $10,
			due_date            = $11,
			implementation_date = $12,
			estimated_cost      = $13,
			actual_cost         = $14,
			updated_at          = NOW()
		WHERE id=$15 AND company_id=$16`,
		qStr(body, "title"),
		qStr(body, "description"),
		qStrD(body, "ca_type", "corrective"),
		qStrD(body, "priority", "medium"),
		qStr(body, "root_cause"),
		qStrD(body, "root_cause_method", "5why"),
		qStr(body, "proposed_action"),
		qStr(body, "implemented_action"),
		qStrN(body, "responsible_id"),
		qStrN(body, "department_id"),
		qStrN(body, "due_date"),
		qStrN(body, "implementation_date"),
		qF64(body, "estimated_cost"),
		qF64(body, "actual_cost"),
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Corrective action updated"})
}

func (h *QualityHandler) UpdateCAStatus(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	userID, _ := c.Get("user_id")
	ctx := context.Background()

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	newStatus := qStr(body, "status")
	var execErr error

	switch newStatus {
	case "closed":
		_, execErr = h.db.Exec(ctx, `
			UPDATE corrective_actions SET
				status    = 'closed'::ca_status,
				closed_by = $1,
				closed_at = NOW(),
				updated_at = NOW()
			WHERE id=$2 AND company_id=$3`,
			fmt.Sprintf("%v", userID), id, companyID,
		)
	case "verified":
		_, execErr = h.db.Exec(ctx, `
			UPDATE corrective_actions SET
				status               = 'verified'::ca_status,
				verified_by          = $1,
				verification_date    = CURRENT_DATE,
				effectiveness_rating = $2,
				effectiveness_notes  = $3,
				updated_at           = NOW()
			WHERE id=$4 AND company_id=$5`,
			fmt.Sprintf("%v", userID),
			qStrN(body, "effectiveness_rating"),
			qStr(body, "effectiveness_notes"),
			id, companyID,
		)
	default:
		_, execErr = h.db.Exec(ctx, `
			UPDATE corrective_actions SET status=$1::ca_status, updated_at=NOW()
			WHERE id=$2 AND company_id=$3`,
			newStatus, id, companyID,
		)
	}

	if execErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": execErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status updated", "status": newStatus})
}

func (h *QualityHandler) DeleteCorrectiveAction(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		DELETE FROM corrective_actions WHERE id=$1 AND company_id=$2 AND status='open'::ca_status`,
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Corrective action deleted"})
}

// ─────────────────────────────────────────────────────────────────────────────
// QUALITY REPORTS
// ─────────────────────────────────────────────────────────────────────────────

func (h *QualityHandler) GetReports(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	fromDate := c.DefaultQuery("from", time.Now().AddDate(0, -3, 0).Format("2006-01-02"))
	toDate   := c.DefaultQuery("to",   time.Now().Format("2006-01-02"))

	// Inspection summary by type
	type InspByType struct {
		InspType string  `json:"inspection_type"`
		Total    int     `json:"total"`
		Passed   int     `json:"passed"`
		Failed   int     `json:"failed"`
		PassRate float64 `json:"pass_rate"`
	}
	var inspByType []InspByType
	rows1, e1 := h.db.Query(ctx, `
		SELECT
			inspection_type::text,
			COUNT(*),
			COUNT(*) FILTER (WHERE status='passed'),
			COUNT(*) FILTER (WHERE status='failed'),
			COALESCE(ROUND(
				COUNT(*) FILTER (WHERE status='passed')::NUMERIC /
				NULLIF(COUNT(*) FILTER (WHERE status IN ('passed','failed')),0)*100
			,2),0)
		FROM quality_inspections
		WHERE company_id=$1
		  AND created_at BETWEEN $2 AND $3::date + INTERVAL '1 day'
		GROUP BY inspection_type
		ORDER BY total DESC`,
		companyID, fromDate, toDate,
	)
	if e1 == nil {
		defer rows1.Close()
		for rows1.Next() {
			var r InspByType
			rows1.Scan(&r.InspType, &r.Total, &r.Passed, &r.Failed, &r.PassRate)
			inspByType = append(inspByType, r)
		}
	}
	if inspByType == nil { inspByType = []InspByType{} }

	// Top defective items
	type DefectItem struct {
		ItemName    string  `json:"item_name"`
		TotalChecks int     `json:"total_checks"`
		FailedChecks int    `json:"failed_checks"`
		DefectRate  float64 `json:"defect_rate"`
	}
	var defectItems []DefectItem
	rows2, e2 := h.db.Query(ctx, `
		SELECT
			COALESCE(i.name, 'Unknown') AS item_name,
			COUNT(qc.id)                AS total_checks,
			COUNT(qc.id) FILTER (WHERE qc.result='fail') AS failed_checks,
			COALESCE(ROUND(
				COUNT(qc.id) FILTER (WHERE qc.result='fail')::NUMERIC /
				NULLIF(COUNT(qc.id),0)*100
			,2),0)                      AS defect_rate
		FROM quality_checks qc
		JOIN quality_inspections qi ON qi.id = qc.inspection_id
		LEFT JOIN items i ON i.id = qi.item_id
		WHERE qi.company_id=$1
		  AND qi.created_at BETWEEN $2 AND $3::date + INTERVAL '1 day'
		GROUP BY i.name
		HAVING COUNT(qc.id) FILTER (WHERE qc.result='fail') > 0
		ORDER BY failed_checks DESC
		LIMIT 10`,
		companyID, fromDate, toDate,
	)
	if e2 == nil {
		defer rows2.Close()
		for rows2.Next() {
			var r DefectItem
			rows2.Scan(&r.ItemName, &r.TotalChecks, &r.FailedChecks, &r.DefectRate)
			defectItems = append(defectItems, r)
		}
	}
	if defectItems == nil { defectItems = []DefectItem{} }

	// NC trend by month
	type NCTrend struct {
		Month    string `json:"month"`
		Total    int    `json:"total"`
		Closed   int    `json:"closed"`
		Critical int    `json:"critical"`
	}
	var ncTrend []NCTrend
	rows3, e3 := h.db.Query(ctx, `
		SELECT
			TO_CHAR(DATE_TRUNC('month', created_at),'YYYY-MM') AS month,
			COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('closed')),
			COUNT(*) FILTER (WHERE severity IN ('critical','critical_safety'))
		FROM non_conformities
		WHERE company_id=$1
		  AND created_at BETWEEN $2 AND $3::date + INTERVAL '1 day'
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY 1`,
		companyID, fromDate, toDate,
	)
	if e3 == nil {
		defer rows3.Close()
		for rows3.Next() {
			var r NCTrend
			rows3.Scan(&r.Month, &r.Total, &r.Closed, &r.Critical)
			ncTrend = append(ncTrend, r)
		}
	}
	if ncTrend == nil { ncTrend = []NCTrend{} }

	// CA effectiveness summary
	type CAEffectiveness struct {
		Status           string  `json:"status"`
		Count            int     `json:"count"`
		AvgClosureDays   float64 `json:"avg_closure_days"`
		AvgEffectiveness float64 `json:"avg_effectiveness"`
	}
	var caEffectiveness []CAEffectiveness
	rows4, e4 := h.db.Query(ctx, `
		SELECT
			status::text,
			COUNT(*),
			COALESCE(ROUND(AVG(EXTRACT(DAY FROM closed_at - created_at)),1),0),
			COALESCE(ROUND(AVG(effectiveness_rating::NUMERIC),2),0)
		FROM corrective_actions
		WHERE company_id=$1
		  AND created_at BETWEEN $2 AND $3::date + INTERVAL '1 day'
		GROUP BY status
		ORDER BY count DESC`,
		companyID, fromDate, toDate,
	)
	if e4 == nil {
		defer rows4.Close()
		for rows4.Next() {
			var r CAEffectiveness
			rows4.Scan(&r.Status, &r.Count, &r.AvgClosureDays, &r.AvgEffectiveness)
			caEffectiveness = append(caEffectiveness, r)
		}
	}
	if caEffectiveness == nil { caEffectiveness = []CAEffectiveness{} }

	// Overall KPIs for period
	type KPIs struct {
		TotalInspections   int     `json:"total_inspections"`
		FirstPassRate      float64 `json:"first_pass_rate"`
		TotalNC            int     `json:"total_nc"`
		NCClosureRate      float64 `json:"nc_closure_rate"`
		TotalCA            int     `json:"total_ca"`
		CAClosureRate      float64 `json:"ca_closure_rate"`
		OverdueCA          int     `json:"overdue_ca"`
		TotalDefects       int     `json:"total_defects"`
		DefectRate         float64 `json:"defect_rate"`
	}
	var kpis KPIs
	h.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(ROUND(COUNT(*) FILTER (WHERE status='passed')::NUMERIC /
				NULLIF(COUNT(*) FILTER (WHERE status IN ('passed','failed')),0)*100,2),0)
		FROM quality_inspections
		WHERE company_id=$1 AND created_at BETWEEN $2 AND $3::date + INTERVAL '1 day'`,
		companyID, fromDate, toDate,
	).Scan(&kpis.TotalInspections, &kpis.FirstPassRate)

	h.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(ROUND(COUNT(*) FILTER (WHERE status='closed')::NUMERIC /
				NULLIF(COUNT(*),0)*100,2),0)
		FROM non_conformities
		WHERE company_id=$1 AND created_at BETWEEN $2 AND $3::date + INTERVAL '1 day'`,
		companyID, fromDate, toDate,
	).Scan(&kpis.TotalNC, &kpis.NCClosureRate)

	h.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(ROUND(COUNT(*) FILTER (WHERE status IN ('closed','verified'))::NUMERIC /
				NULLIF(COUNT(*),0)*100,2),0),
			COUNT(*) FILTER (WHERE due_date < CURRENT_DATE AND status NOT IN ('closed','cancelled','verified'))
		FROM corrective_actions
		WHERE company_id=$1 AND created_at BETWEEN $2 AND $3::date + INTERVAL '1 day'`,
		companyID, fromDate, toDate,
	).Scan(&kpis.TotalCA, &kpis.CAClosureRate, &kpis.OverdueCA)

	h.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE qc.result='fail'),
			COALESCE(ROUND(COUNT(*) FILTER (WHERE qc.result='fail')::NUMERIC /
				NULLIF(COUNT(qc.id),0)*100,2),0)
		FROM quality_checks qc
		JOIN quality_inspections qi ON qi.id = qc.inspection_id
		WHERE qi.company_id=$1 AND qi.created_at BETWEEN $2 AND $3::date + INTERVAL '1 day'`,
		companyID, fromDate, toDate,
	).Scan(&kpis.TotalDefects, &kpis.DefectRate)

	c.JSON(http.StatusOK, gin.H{
		"period":          gin.H{"from": fromDate, "to": toDate},
		"kpis":            kpis,
		"insp_by_type":    inspByType,
		"top_defect_items": defectItems,
		"nc_trend":        ncTrend,
		"ca_effectiveness": caEffectiveness,
	})
}
