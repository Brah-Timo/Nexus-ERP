package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"nexus-erp/internal/middleware"
)

// ─── Diagnostics Handler ─────────────────────────────────────────────────────

type DiagnosticsHandler struct{ db *pgxpool.Pool }

// ─────────────────────────────────────────────────────────────────────────────
// List Logs — GET /diagnostics/logs
// Supports: severity, source, module, resolved, start_date, end_date, search,
//           page, limit
// ─────────────────────────────────────────────────────────────────────────────

func (h *DiagnosticsHandler) ListLogs(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	severity  := c.Query("severity")
	source    := c.Query("source")
	module    := c.Query("module")
	resolved  := c.Query("resolved")
	search    := c.Query("search")
	startDate := c.Query("start_date")
	endDate   := c.Query("end_date")
	pageStr   := c.DefaultQuery("page", "1")
	limitStr  := c.DefaultQuery("limit", "50")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 { page = 1 }
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 200 { limit = 50 }
	offset := (page - 1) * limit

	args := []interface{}{companyID}
	idx := 2
	where := " WHERE company_id = $1 "

	if severity != "" {
		where += " AND severity = $" + strconv.Itoa(idx)
		args = append(args, severity); idx++
	}
	if source != "" {
		where += " AND source = $" + strconv.Itoa(idx)
		args = append(args, source); idx++
	}
	if module != "" {
		where += " AND module ILIKE $" + strconv.Itoa(idx)
		args = append(args, "%"+module+"%"); idx++
	}
	if resolved == "true" {
		where += " AND is_resolved = TRUE "
	} else if resolved == "false" {
		where += " AND is_resolved = FALSE "
	}
	if startDate != "" {
		where += " AND created_at >= $" + strconv.Itoa(idx)
		args = append(args, startDate); idx++
	}
	if endDate != "" {
		where += " AND created_at < ($" + strconv.Itoa(idx) + "::DATE + INTERVAL '1 day')"
		args = append(args, endDate); idx++
	}
	if search != "" {
		where += " AND (message ILIKE $" + strconv.Itoa(idx) +
			" OR endpoint ILIKE $" + strconv.Itoa(idx) +
			" OR error_code ILIKE $" + strconv.Itoa(idx) + ")"
		args = append(args, "%"+search+"%"); idx++
	}

	// Count total
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	h.db.QueryRow(ctx, "SELECT COUNT(*) FROM system_logs"+where, countArgs...).Scan(&total)

	// Fetch page
	args = append(args, limit, offset)
	query := `
		SELECT
			id, severity::TEXT, source::TEXT,
			COALESCE(module,''),
			COALESCE(endpoint,''),
			COALESCE(method,''),
			COALESCE(http_status, 0),
			request_id::TEXT,
			COALESCE(correlation_id,''),
			message,
			COALESCE(error_code,''),
			COALESCE(sql_state,''),
			COALESCE(stack_trace,''),
			COALESCE(ip_address::TEXT,''),
			COALESCE(user_agent,''),
			COALESCE(page_url,''),
			COALESCE(duration_ms, 0),
			is_resolved,
			resolved_at,
			COALESCE(resolution_note,''),
			created_at
		FROM system_logs` + where +
		` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(idx) + ` OFFSET $` + strconv.Itoa(idx+1)

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type LogRow struct {
		ID             string     `json:"id"`
		Severity       string     `json:"severity"`
		Source         string     `json:"source"`
		Module         string     `json:"module"`
		Endpoint       string     `json:"endpoint"`
		Method         string     `json:"method"`
		HTTPStatus     int        `json:"http_status"`
		RequestID      string     `json:"request_id"`
		CorrelationID  string     `json:"correlation_id"`
		Message        string     `json:"message"`
		ErrorCode      string     `json:"error_code"`
		SQLState       string     `json:"sql_state"`
		StackTrace     string     `json:"stack_trace"`
		IPAddress      string     `json:"ip_address"`
		UserAgent      string     `json:"user_agent"`
		PageURL        string     `json:"page_url"`
		DurationMs     int        `json:"duration_ms"`
		IsResolved     bool       `json:"is_resolved"`
		ResolvedAt     *time.Time `json:"resolved_at"`
		ResolutionNote string     `json:"resolution_note"`
		CreatedAt      time.Time  `json:"created_at"`
	}

	var logs []LogRow
	for rows.Next() {
		var l LogRow
		if err := rows.Scan(
			&l.ID, &l.Severity, &l.Source,
			&l.Module, &l.Endpoint, &l.Method,
			&l.HTTPStatus, &l.RequestID, &l.CorrelationID,
			&l.Message, &l.ErrorCode, &l.SQLState,
			&l.StackTrace, &l.IPAddress, &l.UserAgent,
			&l.PageURL, &l.DurationMs,
			&l.IsResolved, &l.ResolvedAt, &l.ResolutionNote,
			&l.CreatedAt,
		); err == nil {
			logs = append(logs, l)
		}
	}
	if logs == nil {
		logs = []LogRow{}
	}
	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
		"pages": (total + limit - 1) / limit,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Create Log — POST /diagnostics/logs
// ─────────────────────────────────────────────────────────────────────────────

func (h *DiagnosticsHandler) CreateLog(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		Severity      string                 `json:"severity"`
		Source        string                 `json:"source"`
		Module        string                 `json:"module"`
		Endpoint      string                 `json:"endpoint"`
		Method        string                 `json:"method"`
		HTTPStatus    int                    `json:"http_status"`
		CorrelationID string                 `json:"correlation_id"`
		Message       string                 `json:"message" binding:"required"`
		ErrorCode     string                 `json:"error_code"`
		SQLState      string                 `json:"sql_state"`
		StackTrace    string                 `json:"stack_trace"`
		RequestBody   map[string]interface{} `json:"request_body"`
		ResponseBody  map[string]interface{} `json:"response_body"`
		IPAddress     string                 `json:"ip_address"`
		UserAgent     string                 `json:"user_agent"`
		PageURL       string                 `json:"page_url"`
		DurationMs    int                    `json:"duration_ms"`
		Tags          []string               `json:"tags"`
		Extra         map[string]interface{} `json:"extra"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Severity == "" { req.Severity = "error" }
	if req.Source == "" { req.Source = "frontend_js" }

	// Use client IP if not provided
	ip := req.IPAddress
	if ip == "" {
		ip = c.ClientIP()
	}
	ua := req.UserAgent
	if ua == "" {
		ua = c.GetHeader("User-Agent")
	}

	id := uuid.NewString()
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO system_logs
		  (id, company_id, severity, source, module, endpoint, method, http_status,
		   correlation_id, message, error_code, sql_state, stack_trace,
		   request_body, response_body, ip_address, user_agent, page_url,
		   duration_ms, tags, extra)
		VALUES ($1,$2,$3::error_severity,$4::error_source,
		        NULLIF($5,''), NULLIF($6,''), NULLIF($7,''),
		        CASE WHEN $8=0 THEN NULL ELSE $8 END,
		        NULLIF($9,''), $10, NULLIF($11,''), NULLIF($12,''),
		        NULLIF($13,''), $14, $15,
		        NULLIF($16,'')::INET, NULLIF($17,''), NULLIF($18,''),
		        CASE WHEN $19=0 THEN NULL ELSE $19 END,
		        $20, $21)
	`, id, companyID, req.Severity, req.Source,
		req.Module, req.Endpoint, req.Method, req.HTTPStatus,
		req.CorrelationID, req.Message, req.ErrorCode, req.SQLState,
		req.StackTrace, req.RequestBody, req.ResponseBody,
		ip, ua, req.PageURL, req.DurationMs,
		req.Tags, req.Extra)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Log created"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Get Log — GET /diagnostics/logs/:id
// ─────────────────────────────────────────────────────────────────────────────

func (h *DiagnosticsHandler) GetLog(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	type FullLog struct {
		ID             string                 `json:"id"`
		Severity       string                 `json:"severity"`
		Source         string                 `json:"source"`
		Module         string                 `json:"module"`
		Endpoint       string                 `json:"endpoint"`
		Method         string                 `json:"method"`
		HTTPStatus     int                    `json:"http_status"`
		RequestID      string                 `json:"request_id"`
		CorrelationID  string                 `json:"correlation_id"`
		Message        string                 `json:"message"`
		ErrorCode      string                 `json:"error_code"`
		SQLState       string                 `json:"sql_state"`
		StackTrace     string                 `json:"stack_trace"`
		RequestBody    map[string]interface{} `json:"request_body"`
		ResponseBody   map[string]interface{} `json:"response_body"`
		IPAddress      string                 `json:"ip_address"`
		UserAgent      string                 `json:"user_agent"`
		PageURL        string                 `json:"page_url"`
		DurationMs     int                    `json:"duration_ms"`
		IsResolved     bool                   `json:"is_resolved"`
		ResolvedAt     *time.Time             `json:"resolved_at"`
		ResolutionNote string                 `json:"resolution_note"`
		Tags           []string               `json:"tags"`
		Extra          map[string]interface{} `json:"extra"`
		CreatedAt      time.Time              `json:"created_at"`
	}
	var l FullLog
	err := h.db.QueryRow(ctx, `
		SELECT id, severity::TEXT, source::TEXT,
		       COALESCE(module,''), COALESCE(endpoint,''), COALESCE(method,''),
		       COALESCE(http_status,0), request_id::TEXT, COALESCE(correlation_id,''),
		       message, COALESCE(error_code,''), COALESCE(sql_state,''),
		       COALESCE(stack_trace,''), request_body, response_body,
		       COALESCE(ip_address::TEXT,''), COALESCE(user_agent,''), COALESCE(page_url,''),
		       COALESCE(duration_ms,0), is_resolved, resolved_at, COALESCE(resolution_note,''),
		       COALESCE(tags, ARRAY[]::TEXT[]), extra, created_at
		FROM system_logs WHERE id = $1 AND company_id = $2
	`, id, companyID).Scan(
		&l.ID, &l.Severity, &l.Source,
		&l.Module, &l.Endpoint, &l.Method,
		&l.HTTPStatus, &l.RequestID, &l.CorrelationID,
		&l.Message, &l.ErrorCode, &l.SQLState,
		&l.StackTrace, &l.RequestBody, &l.ResponseBody,
		&l.IPAddress, &l.UserAgent, &l.PageURL,
		&l.DurationMs, &l.IsResolved, &l.ResolvedAt, &l.ResolutionNote,
		&l.Tags, &l.Extra, &l.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
		return
	}
	c.JSON(http.StatusOK, l)
}

// ─────────────────────────────────────────────────────────────────────────────
// Resolve Log — POST /diagnostics/logs/:id/resolve
// ─────────────────────────────────────────────────────────────────────────────

func (h *DiagnosticsHandler) ResolveLog(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	var req struct {
		ResolutionNote string `json:"resolution_note"`
	}
	c.ShouldBindJSON(&req)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE system_logs
		SET is_resolved = TRUE, resolved_at = NOW(), resolved_by = $1,
		    resolution_note = NULLIF($2,'')
		WHERE id = $3 AND company_id = $4
	`, userID, req.ResolutionNote, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Log resolved"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Bulk Resolve — POST /diagnostics/logs/bulk-resolve
// ─────────────────────────────────────────────────────────────────────────────

func (h *DiagnosticsHandler) BulkResolveLogs(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	userID := middleware.GetUserID(c)
	var req struct {
		IDs            []string `json:"ids" binding:"required"`
		ResolutionNote string   `json:"resolution_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE system_logs
		SET is_resolved = TRUE, resolved_at = NOW(), resolved_by = $1,
		    resolution_note = NULLIF($2,'')
		WHERE id = ANY($3::UUID[]) AND company_id = $4
	`, userID, req.ResolutionNote, req.IDs, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logs resolved", "count": len(req.IDs)})
}

// ─────────────────────────────────────────────────────────────────────────────
// Delete Log — DELETE /diagnostics/logs/:id
// ─────────────────────────────────────────────────────────────────────────────

func (h *DiagnosticsHandler) DeleteLog(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `DELETE FROM system_logs WHERE id = $1 AND company_id = $2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Log deleted"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Get Stats — GET /diagnostics/stats
// ─────────────────────────────────────────────────────────────────────────────

func (h *DiagnosticsHandler) GetStats(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	period := c.DefaultQuery("period", "7d")

	var days int
	switch period {
	case "24h": days = 1
	case "7d": days = 7
	case "30d": days = 30
	case "90d": days = 90
	default: days = 7
	}

	type SeverityStat struct {
		Severity string `json:"severity"`
		Count    int    `json:"count"`
		Unresolved int  `json:"unresolved"`
	}
	type SourceStat struct {
		Source  string `json:"source"`
		Count   int    `json:"count"`
	}
	type ModuleStat struct {
		Module string `json:"module"`
		Count  int    `json:"count"`
		ErrorRate float64 `json:"error_rate"`
	}
	type TopError struct {
		Message       string    `json:"message"`
		Module        string    `json:"module"`
		Count         int       `json:"count"`
		LastSeen      time.Time `json:"last_seen"`
		HTTPStatus    int       `json:"http_status"`
	}
	type HourlyStat struct {
		Hour  time.Time `json:"hour"`
		Count int       `json:"count"`
		Errors int      `json:"errors"`
	}

	// Summary counts
	var (
		totalLogs, criticalCount, errorCount, warningCount, infoCount int
		unresolvedCount, resolvedToday int
	)
	h.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE severity = 'critical'),
			COUNT(*) FILTER (WHERE severity = 'error'),
			COUNT(*) FILTER (WHERE severity = 'warning'),
			COUNT(*) FILTER (WHERE severity = 'info'),
			COUNT(*) FILTER (WHERE is_resolved = FALSE AND severity IN ('error','critical')),
			COUNT(*) FILTER (WHERE is_resolved = TRUE AND resolved_at >= CURRENT_DATE)
		FROM system_logs
		WHERE company_id = $1 AND created_at >= NOW() - ($2 || ' days')::INTERVAL
	`, companyID, days).Scan(&totalLogs, &criticalCount, &errorCount, &warningCount, &infoCount, &unresolvedCount, &resolvedToday)

	// By severity
	sevRows, _ := h.db.Query(ctx, `
		SELECT severity::TEXT, COUNT(*), COUNT(*) FILTER (WHERE is_resolved = FALSE)
		FROM system_logs
		WHERE company_id = $1 AND created_at >= NOW() - ($2 || ' days')::INTERVAL
		GROUP BY severity ORDER BY COUNT(*) DESC
	`, companyID, days)
	var bySeverity []SeverityStat
	if sevRows != nil {
		defer sevRows.Close()
		for sevRows.Next() {
			var s SeverityStat
			sevRows.Scan(&s.Severity, &s.Count, &s.Unresolved)
			bySeverity = append(bySeverity, s)
		}
	}

	// By source
	srcRows, _ := h.db.Query(ctx, `
		SELECT source::TEXT, COUNT(*)
		FROM system_logs
		WHERE company_id = $1 AND created_at >= NOW() - ($2 || ' days')::INTERVAL
		GROUP BY source ORDER BY COUNT(*) DESC
	`, companyID, days)
	var bySource []SourceStat
	if srcRows != nil {
		defer srcRows.Close()
		for srcRows.Next() {
			var s SourceStat
			srcRows.Scan(&s.Source, &s.Count)
			bySource = append(bySource, s)
		}
	}

	// By module (top 10)
	modRows, _ := h.db.Query(ctx, `
		SELECT COALESCE(module,'unknown'), COUNT(*),
		       ROUND(100.0 * COUNT(*) FILTER (WHERE severity IN ('error','critical')) / COUNT(*), 1)
		FROM system_logs
		WHERE company_id = $1 AND created_at >= NOW() - ($2 || ' days')::INTERVAL
		GROUP BY module ORDER BY COUNT(*) DESC LIMIT 10
	`, companyID, days)
	var byModule []ModuleStat
	if modRows != nil {
		defer modRows.Close()
		for modRows.Next() {
			var m ModuleStat
			modRows.Scan(&m.Module, &m.Count, &m.ErrorRate)
			byModule = append(byModule, m)
		}
	}

	// Top 10 recurring errors
	errRows, _ := h.db.Query(ctx, `
		SELECT message, COALESCE(module,''), COUNT(*), MAX(created_at), COALESCE(http_status,0)
		FROM system_logs
		WHERE company_id = $1 AND severity IN ('error','critical')
		  AND created_at >= NOW() - ($2 || ' days')::INTERVAL
		GROUP BY message, module, http_status
		ORDER BY COUNT(*) DESC LIMIT 10
	`, companyID, days)
	var topErrors []TopError
	if errRows != nil {
		defer errRows.Close()
		for errRows.Next() {
			var e TopError
			errRows.Scan(&e.Message, &e.Module, &e.Count, &e.LastSeen, &e.HTTPStatus)
			topErrors = append(topErrors, e)
		}
	}

	// Hourly trend (last 24h)
	hourRows, _ := h.db.Query(ctx, `
		SELECT date_trunc('hour', created_at), COUNT(*),
		       COUNT(*) FILTER (WHERE severity IN ('error','critical'))
		FROM system_logs
		WHERE company_id = $1 AND created_at >= NOW() - INTERVAL '24 hours'
		GROUP BY date_trunc('hour', created_at) ORDER BY 1
	`, companyID)
	var hourlyTrend []HourlyStat
	if hourRows != nil {
		defer hourRows.Close()
		for hourRows.Next() {
			var h HourlyStat
			hourRows.Scan(&h.Hour, &h.Count, &h.Errors)
			hourlyTrend = append(hourlyTrend, h)
		}
	}

	if bySeverity == nil { bySeverity = []SeverityStat{} }
	if bySource == nil { bySource = []SourceStat{} }
	if byModule == nil { byModule = []ModuleStat{} }
	if topErrors == nil { topErrors = []TopError{} }
	if hourlyTrend == nil { hourlyTrend = []HourlyStat{} }

	c.JSON(http.StatusOK, gin.H{
		"period":          period,
		"total":           totalLogs,
		"critical":        criticalCount,
		"errors":          errorCount,
		"warnings":        warningCount,
		"info":            infoCount,
		"unresolved":      unresolvedCount,
		"resolved_today":  resolvedToday,
		"by_severity":     bySeverity,
		"by_source":       bySource,
		"by_module":       byModule,
		"top_errors":      topErrors,
		"hourly_trend":    hourlyTrend,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Purge Old Logs — DELETE /diagnostics/logs/purge
// ─────────────────────────────────────────────────────────────────────────────

func (h *DiagnosticsHandler) PurgeLogs(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	var req struct {
		OlderThanDays int    `json:"older_than_days"`
		Severity      string `json:"severity"`
		OnlyResolved  bool   `json:"only_resolved"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.OlderThanDays < 1 { req.OlderThanDays = 90 }

	ctx := context.Background()
	q := `DELETE FROM system_logs WHERE company_id = $1 AND created_at < NOW() - ($2 || ' days')::INTERVAL`
	args := []interface{}{companyID, req.OlderThanDays}
	idx := 3
	if req.Severity != "" {
		q += ` AND severity = $` + strconv.Itoa(idx)
		args = append(args, req.Severity); idx++
	}
	if req.OnlyResolved {
		q += ` AND is_resolved = TRUE`
	}

	res, err := h.db.Exec(ctx, q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": res.RowsAffected(), "message": "Logs purged"})
}
