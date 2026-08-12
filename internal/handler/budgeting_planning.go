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

// BudgetingHandler handles all budgeting & planning endpoints
type BudgetingHandler struct {
	db *pgxpool.Pool
}

// ─── local helpers (budget-prefixed to avoid redeclaration conflicts) ────────

func budgetFloatVal(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok && v != nil {
		switch t := v.(type) {
		case float64:
			return t
		case float32:
			return float64(t)
		case int:
			return float64(t)
		case int64:
			return float64(t)
		}
	}
	return 0
}

func budgetFloatValDef(m map[string]interface{}, key string, def float64) float64 {
	if v, ok := m[key]; ok && v != nil {
		switch t := v.(type) {
		case float64:
			return t
		case float32:
			return float64(t)
		case int:
			return float64(t)
		case int64:
			return float64(t)
		}
	}
	return def
}

func budgetBoolVal(m map[string]interface{}, key string, def bool) bool {
	if v, ok := m[key]; ok && v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return def
}

func budgetNullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func budgetNullableUUID(s string) interface{} {
	if s == "" || s == "null" {
		return nil
	}
	return s
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

func (h *BudgetingHandler) GetDashboard(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	// KPI: total budgets, total allocated, total spent, total committed
	kpiRow := h.db.QueryRow(ctx, `
		SELECT
			COUNT(*)                                               AS total_budgets,
			COALESCE(SUM(total_amount),0)                         AS total_allocated,
			COALESCE((SELECT SUM(actual_amount) FROM budget_line_items WHERE company_id=$1),0) AS total_spent,
			COALESCE((SELECT SUM(committed_amount) FROM budget_commitments
				WHERE company_id=$1 AND status IN ('pending','approved')),0) AS total_committed,
			COUNT(*) FILTER (WHERE status='active')               AS active_budgets,
			COUNT(*) FILTER (WHERE status='draft')                AS draft_budgets
		FROM annual_budgets WHERE company_id=$1`, companyID)

	var totalBudgets, activeBudgets, draftBudgets int64
	var totalAllocated, totalSpent, totalCommitted float64
	if err := kpiRow.Scan(&totalBudgets, &totalAllocated, &totalSpent, &totalCommitted, &activeBudgets, &draftBudgets); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Budget utilisation by fiscal year (last 3 years)
	byYearRows, err := h.db.Query(ctx, `
		SELECT
			ab.fiscal_year,
			COALESCE(SUM(ab.total_amount),0)                AS allocated,
			COALESCE(SUM(bli.actual_sum),0)                 AS spent,
			COALESCE(SUM(bc.committed_sum),0)               AS committed
		FROM annual_budgets ab
		LEFT JOIN (
			SELECT annual_budget_id, SUM(actual_amount) AS actual_sum
			FROM budget_line_items GROUP BY annual_budget_id
		) bli ON bli.annual_budget_id = ab.id
		LEFT JOIN (
			SELECT annual_budget_id, SUM(committed_amount) AS committed_sum
			FROM budget_commitments WHERE status IN ('pending','approved')
			GROUP BY annual_budget_id
		) bc ON bc.annual_budget_id = ab.id
		WHERE ab.company_id=$1
		GROUP BY ab.fiscal_year
		ORDER BY ab.fiscal_year DESC
		LIMIT 5`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer byYearRows.Close()
	byYear := []map[string]interface{}{}
	for byYearRows.Next() {
		var yr int
		var alloc, spent2, comm float64
		byYearRows.Scan(&yr, &alloc, &spent2, &comm)
		byYear = append(byYear, map[string]interface{}{
			"fiscal_year": yr, "allocated": alloc, "spent": spent2, "committed": comm,
		})
	}

	// Top departments by spend (current year)
	currentYear := time.Now().Year()
	deptRows, err := h.db.Query(ctx, `
		SELECT
			db.department_name,
			COALESCE(SUM(db.allocated_amount),0) AS allocated,
			COALESCE(SUM(db.spent_amount),0)     AS spent
		FROM department_budgets db
		JOIN annual_budgets ab ON ab.id = db.annual_budget_id
		WHERE db.company_id=$1 AND ab.fiscal_year=$2
		GROUP BY db.department_name
		ORDER BY spent DESC
		LIMIT 8`, companyID, currentYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer deptRows.Close()
	topDepts := []map[string]interface{}{}
	for deptRows.Next() {
		var name string
		var alloc, spent2 float64
		deptRows.Scan(&name, &alloc, &spent2)
		topDepts = append(topDepts, map[string]interface{}{
			"department_name": name, "allocated": alloc, "spent": spent2,
		})
	}

	// Budget by category (current year)
	catRows, err := h.db.Query(ctx, `
		SELECT
			COALESCE(bc.name,'Uncategorized') AS category_name,
			COALESCE(SUM(bli.budget_amount),0) AS budget_amount,
			COALESCE(SUM(bli.actual_amount),0) AS actual_amount
		FROM budget_line_items bli
		JOIN annual_budgets ab ON ab.id = bli.annual_budget_id
		LEFT JOIN budget_categories bc ON bc.id = bli.category_id
		WHERE bli.company_id=$1 AND ab.fiscal_year=$2
		GROUP BY bc.name
		ORDER BY budget_amount DESC
		LIMIT 8`, companyID, currentYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer catRows.Close()
	byCategory := []map[string]interface{}{}
	for catRows.Next() {
		var name string
		var budgAmt, actAmt float64
		catRows.Scan(&name, &budgAmt, &actAmt)
		byCategory = append(byCategory, map[string]interface{}{
			"category_name": name, "budget_amount": budgAmt, "actual_amount": actAmt,
		})
	}

	// Recent commitments
	recentCommRows, err := h.db.Query(ctx, `
		SELECT
			commitment_number,
			COALESCE(vendor_name,''),
			COALESCE(description,''),
			committed_amount,
			status::text,
			commitment_date
		FROM budget_commitments
		WHERE company_id=$1
		ORDER BY created_at DESC
		LIMIT 5`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer recentCommRows.Close()
	recentComm := []map[string]interface{}{}
	for recentCommRows.Next() {
		var num, vendor, desc, status string
		var amt float64
		var dt time.Time
		recentCommRows.Scan(&num, &vendor, &desc, &amt, &status, &dt)
		recentComm = append(recentComm, map[string]interface{}{
			"commitment_number": num, "vendor_name": vendor, "description": desc,
			"committed_amount": amt, "status": status, "commitment_date": dt.Format("2006-01-02"),
		})
	}

	// Monthly spend trend (current year)
	monthRows, err := h.db.Query(ctx, `
		SELECT
			EXTRACT(MONTH FROM transaction_date)::int AS month,
			COALESCE(SUM(amount),0) AS total
		FROM budget_actuals
		WHERE company_id=$1
		  AND EXTRACT(YEAR FROM transaction_date) = $2
		  AND posted = true
		GROUP BY month
		ORDER BY month`, companyID, currentYear)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer monthRows.Close()
	monthlyTrend := make([]map[string]interface{}, 12)
	monthNames := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	for i := 0; i < 12; i++ {
		monthlyTrend[i] = map[string]interface{}{"month": monthNames[i], "amount": 0.0}
	}
	for monthRows.Next() {
		var m int
		var amt float64
		monthRows.Scan(&m, &amt)
		if m >= 1 && m <= 12 {
			monthlyTrend[m-1]["amount"] = amt
		}
	}

	utilizationPct := 0.0
	if totalAllocated > 0 {
		utilizationPct = (totalSpent / totalAllocated) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"kpi": map[string]interface{}{
			"total_budgets":     totalBudgets,
			"active_budgets":    activeBudgets,
			"draft_budgets":     draftBudgets,
			"total_allocated":   totalAllocated,
			"total_spent":       totalSpent,
			"total_committed":   totalCommitted,
			"utilization_pct":   utilizationPct,
			"available_balance": totalAllocated - totalSpent - totalCommitted,
		},
		"by_year":        byYear,
		"top_departments": topDepts,
		"by_category":    byCategory,
		"recent_commitments": recentComm,
		"monthly_trend":  monthlyTrend,
	})
}

// ─── Budget Categories ─────────────────────────────────────────────────────────

func (h *BudgetingHandler) ListBudgetCategories(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	rows, err := h.db.Query(ctx, `
		SELECT
			bc.id,
			bc.code,
			bc.name,
			COALESCE(bc.description,''),
			COALESCE(bc.parent_id::text,''),
			COALESCE(p.name,'')     AS parent_name,
			bc.is_active,
			bc.sort_order,
			bc.created_at
		FROM budget_categories bc
		LEFT JOIN budget_categories p ON p.id = bc.parent_id
		WHERE bc.company_id=$1
		ORDER BY bc.sort_order, bc.name`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var id, code, name, desc, parentID, parentName string
		var isActive bool
		var sortOrder int
		var createdAt time.Time
		rows.Scan(&id, &code, &name, &desc, &parentID, &parentName, &isActive, &sortOrder, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "code": code, "name": name, "description": desc,
			"parent_id": parentID, "parent_name": parentName,
			"is_active": isActive, "sort_order": sortOrder,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *BudgetingHandler) CreateBudgetCategory(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO budget_categories (company_id,code,name,description,parent_id,is_active,sort_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		companyID,
		strVal(body, "code"),
		strVal(body, "name"),
		budgetNullableStr(strVal(body, "description")),
		budgetNullableUUID(strVal(body, "parent_id")),
		budgetBoolVal(body, "is_active", true),
		int(budgetFloatValDef(body, "sort_order", 0)),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *BudgetingHandler) UpdateBudgetCategory(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(ctx, `
		UPDATE budget_categories SET
			code=$3, name=$4, description=$5, parent_id=$6,
			is_active=$7, sort_order=$8
		WHERE id=$1 AND company_id=$2`,
		id, companyID,
		strVal(body, "code"),
		strVal(body, "name"),
		budgetNullableStr(strVal(body, "description")),
		budgetNullableUUID(strVal(body, "parent_id")),
		budgetBoolVal(body, "is_active", true),
		int(budgetFloatValDef(body, "sort_order", 0)),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) DeleteBudgetCategory(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	_, err := h.db.Exec(ctx, `DELETE FROM budget_categories WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Annual Budgets ────────────────────────────────────────────────────────────

func (h *BudgetingHandler) ListAnnualBudgets(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	where := []string{"company_id=$1"}
	args := []interface{}{companyID}
	n := 2

	if fy := c.Query("fiscal_year"); fy != "" {
		where = append(where, fmt.Sprintf("fiscal_year=$%d", n))
		args = append(args, fy)
		n++
	}
	if st := c.Query("status"); st != "" {
		where = append(where, fmt.Sprintf("status=$%d::budget_status", n))
		args = append(args, st)
		n++
	}
	if bt := c.Query("budget_type"); bt != "" {
		where = append(where, fmt.Sprintf("budget_type=$%d::budget_type", n))
		args = append(args, bt)
		n++
	}
	_ = n

	q := fmt.Sprintf(`
		SELECT
			id, budget_number, fiscal_year, name,
			COALESCE(description,''),
			budget_type::text, status::text,
			start_date, end_date, total_amount,
			COALESCE(notes,''),
			created_at
		FROM annual_budgets
		WHERE %s
		ORDER BY fiscal_year DESC, created_at DESC`, strings.Join(where, " AND "))

	rows, err := h.db.Query(ctx, q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var id, num, name, desc, btype, status, notes string
		var fy int
		var startDt, endDt time.Time
		var totalAmt float64
		var createdAt time.Time
		rows.Scan(&id, &num, &fy, &name, &desc, &btype, &status, &startDt, &endDt, &totalAmt, &notes, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "budget_number": num, "fiscal_year": fy, "name": name,
			"description": desc, "budget_type": btype, "status": status,
			"start_date": startDt.Format("2006-01-02"), "end_date": endDt.Format("2006-01-02"),
			"total_amount": totalAmt, "notes": notes,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *BudgetingHandler) GetAnnualBudget(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")

	row := h.db.QueryRow(ctx, `
		SELECT
			id, budget_number, fiscal_year, name,
			COALESCE(description,''),
			budget_type::text, status::text,
			start_date, end_date, total_amount,
			COALESCE(notes,''),
			COALESCE(approved_by::text,''),
			COALESCE(approved_at::text,''),
			created_at
		FROM annual_budgets
		WHERE id=$1 AND company_id=$2`, id, companyID)

	var budgetID, num, name, desc, btype, status, notes, approvedBy, approvedAt string
	var fy int
	var startDt, endDt, createdAt time.Time
	var totalAmt float64
	if err := row.Scan(&budgetID, &num, &fy, &name, &desc, &btype, &status, &startDt, &endDt, &totalAmt, &notes, &approvedBy, &approvedAt, &createdAt); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// line items
	liRows, err := h.db.Query(ctx, `
		SELECT
			bli.id,
			COALESCE(bc.name,'Uncategorized'),
			COALESCE(bli.account_code,''),
			COALESCE(bli.account_name,''),
			COALESCE(bli.description,''),
			bli.budget_amount, bli.q1_amount, bli.q2_amount, bli.q3_amount, bli.q4_amount,
			bli.actual_amount, bli.committed_amount
		FROM budget_line_items bli
		LEFT JOIN budget_categories bc ON bc.id = bli.category_id
		WHERE bli.annual_budget_id=$1 AND bli.company_id=$2
		ORDER BY bc.name, bli.account_code`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer liRows.Close()
	lineItems := []map[string]interface{}{}
	for liRows.Next() {
		var liID, catName, accCode, accName, liDesc string
		var budgAmt, q1, q2, q3, q4, actAmt, commAmt float64
		liRows.Scan(&liID, &catName, &accCode, &accName, &liDesc, &budgAmt, &q1, &q2, &q3, &q4, &actAmt, &commAmt)
		lineItems = append(lineItems, map[string]interface{}{
			"id": liID, "category_name": catName, "account_code": accCode,
			"account_name": accName, "description": liDesc,
			"budget_amount": budgAmt, "q1_amount": q1, "q2_amount": q2,
			"q3_amount": q3, "q4_amount": q4,
			"actual_amount": actAmt, "committed_amount": commAmt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id": budgetID, "budget_number": num, "fiscal_year": fy, "name": name,
		"description": desc, "budget_type": btype, "status": status,
		"start_date": startDt.Format("2006-01-02"), "end_date": endDt.Format("2006-01-02"),
		"total_amount": totalAmt, "notes": notes,
		"approved_by": approvedBy, "approved_at": approvedAt,
		"created_at": createdAt.Format(time.RFC3339),
		"line_items": lineItems,
	})
}

func (h *BudgetingHandler) CreateAnnualBudget(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var nextSeq int64
	h.db.QueryRow(ctx, `SELECT nextval('budget_number_seq')`).Scan(&nextSeq)
	budgetNumber := fmt.Sprintf("BGT-%04d", nextSeq)

	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO annual_budgets
			(company_id, budget_number, fiscal_year, name, description,
			 budget_type, status, start_date, end_date, total_amount, notes)
		VALUES ($1,$2,$3,$4,$5,$6::budget_type,$7::budget_status,$8::date,$9::date,$10,$11)
		RETURNING id`,
		companyID, budgetNumber,
		int(budgetFloatValDef(body, "fiscal_year", float64(time.Now().Year()))),
		strVal(body, "name"),
		budgetNullableStr(strVal(body, "description")),
		strValDefault(body, "budget_type", "operational"),
		strValDefault(body, "status", "draft"),
		strVal(body, "start_date"),
		strVal(body, "end_date"),
		budgetFloatVal(body, "total_amount"),
		budgetNullableStr(strVal(body, "notes")),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "budget_number": budgetNumber})
}

func (h *BudgetingHandler) UpdateAnnualBudget(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(ctx, `
		UPDATE annual_budgets SET
			fiscal_year=$3, name=$4, description=$5,
			budget_type=$6::budget_type, status=$7::budget_status,
			start_date=$8::date, end_date=$9::date,
			total_amount=$10, notes=$11
		WHERE id=$1 AND company_id=$2`,
		id, companyID,
		int(budgetFloatValDef(body, "fiscal_year", float64(time.Now().Year()))),
		strVal(body, "name"),
		budgetNullableStr(strVal(body, "description")),
		strValDefault(body, "budget_type", "operational"),
		strValDefault(body, "status", "draft"),
		strVal(body, "start_date"),
		strVal(body, "end_date"),
		budgetFloatVal(body, "total_amount"),
		budgetNullableStr(strVal(body, "notes")),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) DeleteAnnualBudget(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	_, err := h.db.Exec(ctx, `DELETE FROM annual_budgets WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) ApproveBudget(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	_, err := h.db.Exec(ctx, `
		UPDATE annual_budgets SET status='active'::budget_status, approved_at=NOW()
		WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) LockBudget(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	_, err := h.db.Exec(ctx, `
		UPDATE annual_budgets SET status='locked'::budget_status
		WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Budget Line Items ─────────────────────────────────────────────────────────

func (h *BudgetingHandler) ListLineItems(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	budgetID := c.Query("annual_budget_id")

	where := []string{"bli.company_id=$1"}
	args := []interface{}{companyID}
	n := 2
	if budgetID != "" {
		where = append(where, fmt.Sprintf("bli.annual_budget_id=$%d", n))
		args = append(args, budgetID)
		n++
	}
	_ = n

	q := fmt.Sprintf(`
		SELECT
			bli.id,
			COALESCE(bc.name,'Uncategorized'),
			COALESCE(bli.category_id::text,''),
			COALESCE(bli.account_code,''),
			COALESCE(bli.account_name,''),
			COALESCE(bli.description,''),
			bli.budget_amount, bli.q1_amount, bli.q2_amount, bli.q3_amount, bli.q4_amount,
			bli.actual_amount, bli.committed_amount,
			COALESCE(bli.notes,''),
			bli.created_at
		FROM budget_line_items bli
		LEFT JOIN budget_categories bc ON bc.id = bli.category_id
		WHERE %s
		ORDER BY bc.name, bli.account_code`, strings.Join(where, " AND "))

	rows, err := h.db.Query(ctx, q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var id, catName, catID, accCode, accName, desc, notes string
		var budgAmt, q1, q2, q3, q4, actAmt, commAmt float64
		var createdAt time.Time
		rows.Scan(&id, &catName, &catID, &accCode, &accName, &desc,
			&budgAmt, &q1, &q2, &q3, &q4, &actAmt, &commAmt, &notes, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "category_name": catName, "category_id": catID,
			"account_code": accCode, "account_name": accName, "description": desc,
			"budget_amount": budgAmt, "q1_amount": q1, "q2_amount": q2,
			"q3_amount": q3, "q4_amount": q4,
			"actual_amount": actAmt, "committed_amount": commAmt,
			"notes": notes, "created_at": createdAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *BudgetingHandler) CreateLineItem(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO budget_line_items
			(company_id, annual_budget_id, department_budget_id, category_id,
			 account_code, account_name, description,
			 budget_amount, q1_amount, q2_amount, q3_amount, q4_amount, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		RETURNING id`,
		companyID,
		strVal(body, "annual_budget_id"),
		budgetNullableUUID(strVal(body, "department_budget_id")),
		budgetNullableUUID(strVal(body, "category_id")),
		budgetNullableStr(strVal(body, "account_code")),
		budgetNullableStr(strVal(body, "account_name")),
		budgetNullableStr(strVal(body, "description")),
		budgetFloatVal(body, "budget_amount"),
		budgetFloatVal(body, "q1_amount"),
		budgetFloatVal(body, "q2_amount"),
		budgetFloatVal(body, "q3_amount"),
		budgetFloatVal(body, "q4_amount"),
		budgetNullableStr(strVal(body, "notes")),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *BudgetingHandler) UpdateLineItem(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(ctx, `
		UPDATE budget_line_items SET
			category_id=$3, account_code=$4, account_name=$5,
			description=$6, budget_amount=$7,
			q1_amount=$8, q2_amount=$9, q3_amount=$10, q4_amount=$11, notes=$12
		WHERE id=$1 AND company_id=$2`,
		id, companyID,
		budgetNullableUUID(strVal(body, "category_id")),
		budgetNullableStr(strVal(body, "account_code")),
		budgetNullableStr(strVal(body, "account_name")),
		budgetNullableStr(strVal(body, "description")),
		budgetFloatVal(body, "budget_amount"),
		budgetFloatVal(body, "q1_amount"),
		budgetFloatVal(body, "q2_amount"),
		budgetFloatVal(body, "q3_amount"),
		budgetFloatVal(body, "q4_amount"),
		budgetNullableStr(strVal(body, "notes")),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) DeleteLineItem(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	_, err := h.db.Exec(ctx, `DELETE FROM budget_line_items WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Department Budgets ────────────────────────────────────────────────────────

func (h *BudgetingHandler) ListDepartmentBudgets(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	where := []string{"db.company_id=$1"}
	args := []interface{}{companyID}
	n := 2
	if bid := c.Query("annual_budget_id"); bid != "" {
		where = append(where, fmt.Sprintf("db.annual_budget_id=$%d", n))
		args = append(args, bid)
		n++
	}
	_ = n

	q := fmt.Sprintf(`
		SELECT
			db.id,
			ab.budget_number,
			ab.fiscal_year,
			ab.name         AS budget_name,
			db.annual_budget_id::text,
			db.department_name,
			COALESCE(db.department_code,''),
			db.allocated_amount,
			db.spent_amount,
			db.committed_amount,
			COALESCE(db.notes,''),
			db.created_at
		FROM department_budgets db
		JOIN annual_budgets ab ON ab.id = db.annual_budget_id
		WHERE %s
		ORDER BY ab.fiscal_year DESC, db.department_name`, strings.Join(where, " AND "))

	rows, err := h.db.Query(ctx, q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var id, budgetNum, budgetName, budgetID, deptName, deptCode, notes string
		var fy int
		var allocated, spent, committed float64
		var createdAt time.Time
		rows.Scan(&id, &budgetNum, &fy, &budgetName, &budgetID,
			&deptName, &deptCode, &allocated, &spent, &committed, &notes, &createdAt)
		available := allocated - spent - committed
		utilizationPct := 0.0
		if allocated > 0 {
			utilizationPct = (spent / allocated) * 100
		}
		list = append(list, map[string]interface{}{
			"id": id, "budget_number": budgetNum, "fiscal_year": fy, "budget_name": budgetName,
			"annual_budget_id": budgetID, "department_name": deptName, "department_code": deptCode,
			"allocated_amount": allocated, "spent_amount": spent, "committed_amount": committed,
			"available_amount": available, "utilization_pct": utilizationPct,
			"notes": notes, "created_at": createdAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *BudgetingHandler) CreateDepartmentBudget(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO department_budgets
			(company_id, annual_budget_id, department_name, department_code,
			 allocated_amount, notes)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id`,
		companyID,
		strVal(body, "annual_budget_id"),
		strVal(body, "department_name"),
		budgetNullableStr(strVal(body, "department_code")),
		budgetFloatVal(body, "allocated_amount"),
		budgetNullableStr(strVal(body, "notes")),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *BudgetingHandler) UpdateDepartmentBudget(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(ctx, `
		UPDATE department_budgets SET
			department_name=$3, department_code=$4,
			allocated_amount=$5, notes=$6
		WHERE id=$1 AND company_id=$2`,
		id, companyID,
		strVal(body, "department_name"),
		budgetNullableStr(strVal(body, "department_code")),
		budgetFloatVal(body, "allocated_amount"),
		budgetNullableStr(strVal(body, "notes")),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) DeleteDepartmentBudget(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	_, err := h.db.Exec(ctx, `DELETE FROM department_budgets WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Budget vs Actual ──────────────────────────────────────────────────────────

func (h *BudgetingHandler) GetBudgetVsActual(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	budgetID := c.Query("annual_budget_id")
	deptID := c.Query("department_budget_id")

	// summary totals
	whereSum := []string{"bli.company_id=$1"}
	argsSum := []interface{}{companyID}
	nSum := 2
	if budgetID != "" {
		whereSum = append(whereSum, fmt.Sprintf("bli.annual_budget_id=$%d", nSum))
		argsSum = append(argsSum, budgetID)
		nSum++
	}
	if deptID != "" {
		whereSum = append(whereSum, fmt.Sprintf("bli.department_budget_id=$%d", nSum))
		argsSum = append(argsSum, deptID)
		nSum++
	}
	_ = nSum

	sumRow := h.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT
			COALESCE(SUM(budget_amount),0),
			COALESCE(SUM(actual_amount),0),
			COALESCE(SUM(committed_amount),0)
		FROM budget_line_items bli
		WHERE %s`, strings.Join(whereSum, " AND ")), argsSum...)

	var totalBudget, totalActual, totalCommitted float64
	sumRow.Scan(&totalBudget, &totalActual, &totalCommitted)

	// by category
	whereCat := []string{"bli.company_id=$1"}
	argsCat := []interface{}{companyID}
	nCat := 2
	if budgetID != "" {
		whereCat = append(whereCat, fmt.Sprintf("bli.annual_budget_id=$%d", nCat))
		argsCat = append(argsCat, budgetID)
		nCat++
	}
	if deptID != "" {
		whereCat = append(whereCat, fmt.Sprintf("bli.department_budget_id=$%d", nCat))
		argsCat = append(argsCat, deptID)
		nCat++
	}
	_ = nCat

	catRows, err := h.db.Query(ctx, fmt.Sprintf(`
		SELECT
			COALESCE(bc.name,'Uncategorized') AS category_name,
			COALESCE(SUM(bli.budget_amount),0)    AS budget_amt,
			COALESCE(SUM(bli.actual_amount),0)    AS actual_amt,
			COALESCE(SUM(bli.committed_amount),0) AS committed_amt
		FROM budget_line_items bli
		LEFT JOIN budget_categories bc ON bc.id = bli.category_id
		WHERE %s
		GROUP BY bc.name
		ORDER BY budget_amt DESC`, strings.Join(whereCat, " AND ")), argsCat...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer catRows.Close()
	byCategory := []map[string]interface{}{}
	for catRows.Next() {
		var catName string
		var budgAmt, actAmt, commAmt float64
		catRows.Scan(&catName, &budgAmt, &actAmt, &commAmt)
		variance := budgAmt - actAmt
		pct := 0.0
		if budgAmt > 0 {
			pct = (actAmt / budgAmt) * 100
		}
		byCategory = append(byCategory, map[string]interface{}{
			"category_name":    catName,
			"budget_amount":    budgAmt,
			"actual_amount":    actAmt,
			"committed_amount": commAmt,
			"variance":         variance,
			"utilization_pct":  pct,
		})
	}

	// monthly actuals
	whereMonth := []string{"ba.company_id=$1"}
	argsMonth := []interface{}{companyID}
	nMonth := 2
	if budgetID != "" {
		whereMonth = append(whereMonth, fmt.Sprintf("ba.annual_budget_id=$%d", nMonth))
		argsMonth = append(argsMonth, budgetID)
		nMonth++
	}
	_ = nMonth

	monthRows, err := h.db.Query(ctx, fmt.Sprintf(`
		SELECT
			EXTRACT(MONTH FROM transaction_date)::int AS m,
			EXTRACT(YEAR FROM transaction_date)::int  AS y,
			COALESCE(SUM(amount),0) AS total
		FROM budget_actuals ba
		WHERE %s AND posted=true
		GROUP BY m, y
		ORDER BY y, m`, strings.Join(whereMonth, " AND ")), argsMonth...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer monthRows.Close()
	monthlyActuals := []map[string]interface{}{}
	for monthRows.Next() {
		var m, y int
		var amt float64
		monthRows.Scan(&m, &y, &amt)
		monthNames := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
		label := fmt.Sprintf("%s %d", monthNames[m], y)
		monthlyActuals = append(monthlyActuals, map[string]interface{}{
			"month": label, "amount": amt, "month_num": m, "year": y,
		})
	}

	// line-level detail
	whereDetail := []string{"bli.company_id=$1"}
	argsDetail := []interface{}{companyID}
	nDetail := 2
	if budgetID != "" {
		whereDetail = append(whereDetail, fmt.Sprintf("bli.annual_budget_id=$%d", nDetail))
		argsDetail = append(argsDetail, budgetID)
		nDetail++
	}
	if deptID != "" {
		whereDetail = append(whereDetail, fmt.Sprintf("bli.department_budget_id=$%d", nDetail))
		argsDetail = append(argsDetail, deptID)
		nDetail++
	}
	_ = nDetail

	detailRows, err := h.db.Query(ctx, fmt.Sprintf(`
		SELECT
			bli.id,
			COALESCE(bc.name,'Uncategorized'),
			COALESCE(bli.account_code,''),
			COALESCE(bli.account_name,''),
			bli.budget_amount,
			bli.actual_amount,
			bli.committed_amount,
			bli.q1_amount, bli.q2_amount, bli.q3_amount, bli.q4_amount
		FROM budget_line_items bli
		LEFT JOIN budget_categories bc ON bc.id = bli.category_id
		WHERE %s
		ORDER BY bc.name, bli.account_code`, strings.Join(whereDetail, " AND ")), argsDetail...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer detailRows.Close()
	lineDetails := []map[string]interface{}{}
	for detailRows.Next() {
		var id, catName, accCode, accName string
		var budgAmt, actAmt, commAmt, q1, q2, q3, q4 float64
		detailRows.Scan(&id, &catName, &accCode, &accName, &budgAmt, &actAmt, &commAmt, &q1, &q2, &q3, &q4)
		variance := budgAmt - actAmt
		pct := 0.0
		if budgAmt > 0 {
			pct = (actAmt / budgAmt) * 100
		}
		lineDetails = append(lineDetails, map[string]interface{}{
			"id": id, "category_name": catName,
			"account_code": accCode, "account_name": accName,
			"budget_amount": budgAmt, "actual_amount": actAmt,
			"committed_amount": commAmt, "variance": variance, "utilization_pct": pct,
			"q1_amount": q1, "q2_amount": q2, "q3_amount": q3, "q4_amount": q4,
		})
	}

	overallPct := 0.0
	if totalBudget > 0 {
		overallPct = (totalActual / totalBudget) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": map[string]interface{}{
			"total_budget":    totalBudget,
			"total_actual":    totalActual,
			"total_committed": totalCommitted,
			"variance":        totalBudget - totalActual,
			"utilization_pct": overallPct,
		},
		"by_category":     byCategory,
		"monthly_actuals": monthlyActuals,
		"line_items":      lineDetails,
	})
}

// ─── Budget Revisions ──────────────────────────────────────────────────────────

func (h *BudgetingHandler) ListRevisions(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	where := []string{"br.company_id=$1"}
	args := []interface{}{companyID}
	n := 2
	if bid := c.Query("annual_budget_id"); bid != "" {
		where = append(where, fmt.Sprintf("br.annual_budget_id=$%d", n))
		args = append(args, bid)
		n++
	}
	if st := c.Query("status"); st != "" {
		where = append(where, fmt.Sprintf("br.status=$%d::budget_status", n))
		args = append(args, st)
		n++
	}
	_ = n

	q := fmt.Sprintf(`
		SELECT
			br.id,
			br.revision_number,
			ab.budget_number,
			ab.name              AS budget_name,
			br.annual_budget_id::text,
			br.revision_type::text,
			br.original_amount,
			br.revised_amount,
			br.change_amount,
			br.reason,
			br.status::text,
			COALESCE(br.effective_date::text,''),
			COALESCE(br.notes,''),
			br.created_at
		FROM budget_revisions br
		JOIN annual_budgets ab ON ab.id = br.annual_budget_id
		WHERE %s
		ORDER BY br.created_at DESC`, strings.Join(where, " AND "))

	rows, err := h.db.Query(ctx, q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var id, revNum, budgetNum, budgetName, budgetID, revType, reason, status, effDate, notes string
		var origAmt, revisedAmt, changeAmt float64
		var createdAt time.Time
		rows.Scan(&id, &revNum, &budgetNum, &budgetName, &budgetID, &revType,
			&origAmt, &revisedAmt, &changeAmt, &reason, &status, &effDate, &notes, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "revision_number": revNum, "budget_number": budgetNum,
			"budget_name": budgetName, "annual_budget_id": budgetID,
			"revision_type": revType, "original_amount": origAmt,
			"revised_amount": revisedAmt, "change_amount": changeAmt,
			"reason": reason, "status": status, "effective_date": effDate,
			"notes": notes, "created_at": createdAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *BudgetingHandler) CreateRevision(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var nextSeq int64
	h.db.QueryRow(ctx, `SELECT nextval('revision_number_seq')`).Scan(&nextSeq)
	revNumber := fmt.Sprintf("REV-%04d", nextSeq)

	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO budget_revisions
			(company_id, revision_number, annual_budget_id, line_item_id,
			 department_budget_id, revision_type, original_amount, revised_amount,
			 reason, status, effective_date, notes)
		VALUES ($1,$2,$3,$4,$5,$6::revision_type,$7,$8,$9,$10::budget_status,$11::date,$12)
		RETURNING id`,
		companyID, revNumber,
		strVal(body, "annual_budget_id"),
		budgetNullableUUID(strVal(body, "line_item_id")),
		budgetNullableUUID(strVal(body, "department_budget_id")),
		strValDefault(body, "revision_type", "increase"),
		budgetFloatVal(body, "original_amount"),
		budgetFloatVal(body, "revised_amount"),
		strVal(body, "reason"),
		strValDefault(body, "status", "draft"),
		budgetNullableStr(strVal(body, "effective_date")),
		budgetNullableStr(strVal(body, "notes")),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "revision_number": revNumber})
}

func (h *BudgetingHandler) UpdateRevision(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(ctx, `
		UPDATE budget_revisions SET
			revision_type=$3::revision_type, original_amount=$4, revised_amount=$5,
			reason=$6, status=$7::budget_status, effective_date=$8::date, notes=$9
		WHERE id=$1 AND company_id=$2`,
		id, companyID,
		strValDefault(body, "revision_type", "increase"),
		budgetFloatVal(body, "original_amount"),
		budgetFloatVal(body, "revised_amount"),
		strVal(body, "reason"),
		strValDefault(body, "status", "draft"),
		budgetNullableStr(strVal(body, "effective_date")),
		budgetNullableStr(strVal(body, "notes")),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) ApproveRevision(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")

	// Get revision details to apply change to line item
	row := h.db.QueryRow(ctx, `
		SELECT line_item_id, revised_amount, department_budget_id
		FROM budget_revisions WHERE id=$1 AND company_id=$2`, id, companyID)
	var lineItemID, deptBudgetID *string
	var revisedAmt float64
	row.Scan(&lineItemID, &revisedAmt, &deptBudgetID)

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE budget_revisions SET status='active'::budget_status, approved_at=NOW()
		WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Apply revision to line item if linked
	if lineItemID != nil && *lineItemID != "" {
		_, err = tx.Exec(ctx, `
			UPDATE budget_line_items SET budget_amount=$2
			WHERE id=$1`, *lineItemID, revisedAmt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	tx.Commit(ctx)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) DeleteRevision(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	_, err := h.db.Exec(ctx, `DELETE FROM budget_revisions WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Budget Commitments ────────────────────────────────────────────────────────

func (h *BudgetingHandler) ListCommitments(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	where := []string{"bc.company_id=$1"}
	args := []interface{}{companyID}
	n := 2
	if bid := c.Query("annual_budget_id"); bid != "" {
		where = append(where, fmt.Sprintf("bc.annual_budget_id=$%d", n))
		args = append(args, bid)
		n++
	}
	if st := c.Query("status"); st != "" {
		where = append(where, fmt.Sprintf("bc.status=$%d::commitment_status", n))
		args = append(args, st)
		n++
	}
	if ct := c.Query("commitment_type"); ct != "" {
		where = append(where, fmt.Sprintf("bc.commitment_type=$%d::commitment_type", n))
		args = append(args, ct)
		n++
	}
	_ = n

	q := fmt.Sprintf(`
		SELECT
			bc.id,
			bc.commitment_number,
			ab.budget_number,
			ab.name              AS budget_name,
			bc.annual_budget_id::text,
			bc.commitment_type::text,
			bc.status::text,
			COALESCE(bc.reference_number,''),
			COALESCE(bc.vendor_name,''),
			bc.description,
			bc.committed_amount,
			bc.fulfilled_amount,
			bc.remaining_amount,
			bc.commitment_date,
			COALESCE(bc.expected_fulfillment::text,''),
			COALESCE(bc.notes,''),
			bc.created_at
		FROM budget_commitments bc
		JOIN annual_budgets ab ON ab.id = bc.annual_budget_id
		WHERE %s
		ORDER BY bc.created_at DESC`, strings.Join(where, " AND "))

	rows, err := h.db.Query(ctx, q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var id, commNum, budgetNum, budgetName, budgetID, commType, status string
		var refNum, vendor, desc, notes, expFulfill string
		var committedAmt, fulfilledAmt, remainingAmt float64
		var commDt, createdAt time.Time
		rows.Scan(&id, &commNum, &budgetNum, &budgetName, &budgetID, &commType, &status,
			&refNum, &vendor, &desc, &committedAmt, &fulfilledAmt, &remainingAmt,
			&commDt, &expFulfill, &notes, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "commitment_number": commNum, "budget_number": budgetNum,
			"budget_name": budgetName, "annual_budget_id": budgetID,
			"commitment_type": commType, "status": status,
			"reference_number": refNum, "vendor_name": vendor, "description": desc,
			"committed_amount": committedAmt, "fulfilled_amount": fulfilledAmt,
			"remaining_amount": remainingAmt,
			"commitment_date":     commDt.Format("2006-01-02"),
			"expected_fulfillment": expFulfill,
			"notes": notes, "created_at": createdAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *BudgetingHandler) CreateCommitment(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var nextSeq int64
	h.db.QueryRow(ctx, `SELECT nextval('commitment_number_seq')`).Scan(&nextSeq)
	commNumber := fmt.Sprintf("CMT-%04d", nextSeq)

	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO budget_commitments
			(company_id, commitment_number, annual_budget_id, department_budget_id,
			 line_item_id, commitment_type, status, reference_number, vendor_name,
			 description, committed_amount, commitment_date, expected_fulfillment, notes)
		VALUES ($1,$2,$3,$4,$5,$6::commitment_type,$7::commitment_status,$8,$9,$10,$11,$12::date,$13::date,$14)
		RETURNING id`,
		companyID, commNumber,
		strVal(body, "annual_budget_id"),
		budgetNullableUUID(strVal(body, "department_budget_id")),
		budgetNullableUUID(strVal(body, "line_item_id")),
		strValDefault(body, "commitment_type", "purchase_order"),
		strValDefault(body, "status", "pending"),
		budgetNullableStr(strVal(body, "reference_number")),
		budgetNullableStr(strVal(body, "vendor_name")),
		strVal(body, "description"),
		budgetFloatVal(body, "committed_amount"),
		strVal(body, "commitment_date"),
		budgetNullableStr(strVal(body, "expected_fulfillment")),
		budgetNullableStr(strVal(body, "notes")),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "commitment_number": commNumber})
}

func (h *BudgetingHandler) UpdateCommitment(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err := h.db.Exec(ctx, `
		UPDATE budget_commitments SET
			commitment_type=$3::commitment_type, status=$4::commitment_status,
			reference_number=$5, vendor_name=$6, description=$7,
			committed_amount=$8, commitment_date=$9::date,
			expected_fulfillment=$10::date, notes=$11
		WHERE id=$1 AND company_id=$2`,
		id, companyID,
		strValDefault(body, "commitment_type", "purchase_order"),
		strValDefault(body, "status", "pending"),
		budgetNullableStr(strVal(body, "reference_number")),
		budgetNullableStr(strVal(body, "vendor_name")),
		strVal(body, "description"),
		budgetFloatVal(body, "committed_amount"),
		strVal(body, "commitment_date"),
		budgetNullableStr(strVal(body, "expected_fulfillment")),
		budgetNullableStr(strVal(body, "notes")),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) ApproveCommitment(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	_, err := h.db.Exec(ctx, `
		UPDATE budget_commitments SET status='approved'::commitment_status, approved_at=NOW()
		WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) FulfillCommitment(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fulfilledAmt := budgetFloatVal(body, "fulfilled_amount")
	_, err := h.db.Exec(ctx, `
		UPDATE budget_commitments
		SET fulfilled_amount=$3,
		    status = CASE WHEN $3 >= committed_amount THEN 'fulfilled'::commitment_status
			             ELSE status END
		WHERE id=$1 AND company_id=$2`, id, companyID, fulfilledAmt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) CancelCommitment(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	_, err := h.db.Exec(ctx, `
		UPDATE budget_commitments SET status='cancelled'::commitment_status
		WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BudgetingHandler) DeleteCommitment(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	id := c.Param("id")
	_, err := h.db.Exec(ctx, `DELETE FROM budget_commitments WHERE id=$1 AND company_id=$2`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ─── Budget Actuals ────────────────────────────────────────────────────────────

func (h *BudgetingHandler) ListActuals(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()

	where := []string{"company_id=$1"}
	args := []interface{}{companyID}
	n := 2
	if bid := c.Query("annual_budget_id"); bid != "" {
		where = append(where, fmt.Sprintf("annual_budget_id=$%d", n))
		args = append(args, bid)
		n++
	}
	_ = n

	q := fmt.Sprintf(`
		SELECT
			id,
			annual_budget_id::text,
			transaction_date,
			COALESCE(reference_type,''),
			COALESCE(reference_number,''),
			COALESCE(description,''),
			amount,
			posted,
			COALESCE(posted_at::text,''),
			created_at
		FROM budget_actuals
		WHERE %s
		ORDER BY transaction_date DESC, created_at DESC`, strings.Join(where, " AND "))

	rows, err := h.db.Query(ctx, q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	list := []map[string]interface{}{}
	for rows.Next() {
		var id, budgetID, refType, refNum, desc, postedAt string
		var txDt, createdAt time.Time
		var amount float64
		var posted bool
		rows.Scan(&id, &budgetID, &txDt, &refType, &refNum, &desc, &amount, &posted, &postedAt, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "annual_budget_id": budgetID,
			"transaction_date": txDt.Format("2006-01-02"),
			"reference_type": refType, "reference_number": refNum,
			"description": desc, "amount": amount,
			"posted": posted, "posted_at": postedAt,
			"created_at": createdAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": list})
}

func (h *BudgetingHandler) CreateActual(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id string
	err := h.db.QueryRow(ctx, `
		INSERT INTO budget_actuals
			(company_id, annual_budget_id, department_budget_id, line_item_id,
			 transaction_date, reference_type, reference_number, description, amount)
		VALUES ($1,$2,$3,$4,$5::date,$6,$7,$8,$9)
		RETURNING id`,
		companyID,
		strVal(body, "annual_budget_id"),
		budgetNullableUUID(strVal(body, "department_budget_id")),
		budgetNullableUUID(strVal(body, "line_item_id")),
		strVal(body, "transaction_date"),
		budgetNullableStr(strVal(body, "reference_type")),
		budgetNullableStr(strVal(body, "reference_number")),
		budgetNullableStr(strVal(body, "description")),
		budgetFloatVal(body, "amount"),
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *BudgetingHandler) PostActuals(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ids is array of strings
	idsRaw, _ := body["ids"].([]interface{})
	ids := make([]string, 0, len(idsRaw))
	for _, v := range idsRaw {
		if s, ok := v.(string); ok {
			ids = append(ids, s)
		}
	}
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no ids provided"})
		return
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+1)
	args[0] = companyID
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	_, err := h.db.Exec(ctx, fmt.Sprintf(`
		UPDATE budget_actuals
		SET posted=true, posted_at=NOW()
		WHERE company_id=$1 AND id IN (%s) AND posted=false`, strings.Join(placeholders, ",")), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update actual_amount on line items
	_, _ = h.db.Exec(ctx, `
		UPDATE budget_line_items bli
		SET actual_amount = (
			SELECT COALESCE(SUM(amount),0)
			FROM budget_actuals ba
			WHERE ba.line_item_id = bli.id AND ba.posted = true
		)
		WHERE bli.company_id=$1`, companyID)

	c.JSON(http.StatusOK, gin.H{"success": true, "posted_count": len(ids)})
}

// ─── Budget Reports ────────────────────────────────────────────────────────────

func (h *BudgetingHandler) GetReports(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := c.Request.Context()
	reportType := c.Query("type")
	fiscalYear := c.Query("fiscal_year")
	if fiscalYear == "" {
		fiscalYear = fmt.Sprintf("%d", time.Now().Year())
	}

	switch reportType {
	case "budget_summary":
		rows, err := h.db.Query(ctx, `
			SELECT
				ab.budget_number,
				ab.name,
				ab.fiscal_year,
				ab.budget_type::text,
				ab.status::text,
				ab.total_amount,
				COALESCE(SUM(bli.actual_amount),0)    AS total_spent,
				COALESCE(SUM(bli.committed_amount),0) AS total_committed,
				ab.start_date,
				ab.end_date
			FROM annual_budgets ab
			LEFT JOIN budget_line_items bli ON bli.annual_budget_id = ab.id
			WHERE ab.company_id=$1 AND ab.fiscal_year=$2
			GROUP BY ab.id
			ORDER BY ab.budget_number`, companyID, fiscalYear)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var num, name, btype, status string
			var fy int
			var total, spent, committed float64
			var startDt, endDt time.Time
			rows.Scan(&num, &name, &fy, &btype, &status, &total, &spent, &committed, &startDt, &endDt)
			pct := 0.0
			if total > 0 {
				pct = (spent / total) * 100
			}
			list = append(list, map[string]interface{}{
				"budget_number": num, "name": name, "fiscal_year": fy,
				"budget_type": btype, "status": status, "total_amount": total,
				"spent_amount": spent, "committed_amount": committed,
				"available_amount": total - spent - committed,
				"utilization_pct":  pct,
				"start_date": startDt.Format("2006-01-02"), "end_date": endDt.Format("2006-01-02"),
			})
		}
		c.JSON(http.StatusOK, gin.H{"type": "budget_summary", "fiscal_year": fiscalYear, "data": list})

	case "department_performance":
		rows, err := h.db.Query(ctx, `
			SELECT
				db.department_name,
				COALESCE(db.department_code,''),
				ab.budget_number,
				ab.name         AS budget_name,
				db.allocated_amount,
				db.spent_amount,
				db.committed_amount,
				ab.fiscal_year
			FROM department_budgets db
			JOIN annual_budgets ab ON ab.id = db.annual_budget_id
			WHERE db.company_id=$1 AND ab.fiscal_year=$2
			ORDER BY db.spent_amount DESC`, companyID, fiscalYear)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var deptName, deptCode, budgetNum, budgetName string
			var fy int
			var allocated, spent, committed float64
			rows.Scan(&deptName, &deptCode, &budgetNum, &budgetName, &allocated, &spent, &committed, &fy)
			pct := 0.0
			if allocated > 0 {
				pct = (spent / allocated) * 100
			}
			list = append(list, map[string]interface{}{
				"department_name": deptName, "department_code": deptCode,
				"budget_number": budgetNum, "budget_name": budgetName,
				"fiscal_year": fy, "allocated_amount": allocated,
				"spent_amount": spent, "committed_amount": committed,
				"available_amount": allocated - spent - committed,
				"utilization_pct":  pct,
			})
		}
		c.JSON(http.StatusOK, gin.H{"type": "department_performance", "fiscal_year": fiscalYear, "data": list})

	case "variance_analysis":
		rows, err := h.db.Query(ctx, `
			SELECT
				COALESCE(bc.name,'Uncategorized') AS category_name,
				COALESCE(SUM(bli.budget_amount),0)    AS budget_amount,
				COALESCE(SUM(bli.actual_amount),0)    AS actual_amount,
				COALESCE(SUM(bli.committed_amount),0) AS committed_amount,
				COALESCE(SUM(bli.q1_amount),0)        AS q1_budget,
				COALESCE(SUM(bli.q2_amount),0)        AS q2_budget,
				COALESCE(SUM(bli.q3_amount),0)        AS q3_budget,
				COALESCE(SUM(bli.q4_amount),0)        AS q4_budget
			FROM budget_line_items bli
			JOIN annual_budgets ab ON ab.id = bli.annual_budget_id
			LEFT JOIN budget_categories bc ON bc.id = bli.category_id
			WHERE bli.company_id=$1 AND ab.fiscal_year=$2
			GROUP BY bc.name
			ORDER BY budget_amount DESC`, companyID, fiscalYear)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var catName string
			var budgAmt, actAmt, commAmt, q1, q2, q3, q4 float64
			rows.Scan(&catName, &budgAmt, &actAmt, &commAmt, &q1, &q2, &q3, &q4)
			variance := budgAmt - actAmt
			pct := 0.0
			if budgAmt > 0 {
				pct = (actAmt / budgAmt) * 100
			}
			list = append(list, map[string]interface{}{
				"category_name": catName, "budget_amount": budgAmt,
				"actual_amount": actAmt, "committed_amount": commAmt,
				"variance": variance, "utilization_pct": pct,
				"q1_budget": q1, "q2_budget": q2, "q3_budget": q3, "q4_budget": q4,
			})
		}
		c.JSON(http.StatusOK, gin.H{"type": "variance_analysis", "fiscal_year": fiscalYear, "data": list})

	case "commitment_report":
		rows, err := h.db.Query(ctx, `
			SELECT
				bc.commitment_number,
				bc.commitment_type::text,
				bc.status::text,
				COALESCE(bc.vendor_name,''),
				bc.description,
				bc.committed_amount,
				bc.fulfilled_amount,
				bc.remaining_amount,
				bc.commitment_date,
				COALESCE(bc.expected_fulfillment::text,''),
				ab.budget_number,
				ab.fiscal_year
			FROM budget_commitments bc
			JOIN annual_budgets ab ON ab.id = bc.annual_budget_id
			WHERE bc.company_id=$1 AND ab.fiscal_year=$2
			ORDER BY bc.commitment_date DESC`, companyID, fiscalYear)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var commNum, commType, status, vendor, desc, expFulfill, budgetNum string
			var fy int
			var committed, fulfilled, remaining float64
			var commDt time.Time
			rows.Scan(&commNum, &commType, &status, &vendor, &desc,
				&committed, &fulfilled, &remaining, &commDt, &expFulfill, &budgetNum, &fy)
			list = append(list, map[string]interface{}{
				"commitment_number": commNum, "commitment_type": commType, "status": status,
				"vendor_name": vendor, "description": desc,
				"committed_amount": committed, "fulfilled_amount": fulfilled, "remaining_amount": remaining,
				"commitment_date": commDt.Format("2006-01-02"), "expected_fulfillment": expFulfill,
				"budget_number": budgetNum, "fiscal_year": fy,
			})
		}
		c.JSON(http.StatusOK, gin.H{"type": "commitment_report", "fiscal_year": fiscalYear, "data": list})

	case "revision_history":
		rows, err := h.db.Query(ctx, `
			SELECT
				br.revision_number,
				ab.budget_number,
				ab.name          AS budget_name,
				br.revision_type::text,
				br.original_amount,
				br.revised_amount,
				br.change_amount,
				br.reason,
				br.status::text,
				COALESCE(br.effective_date::text,''),
				br.created_at,
				ab.fiscal_year
			FROM budget_revisions br
			JOIN annual_budgets ab ON ab.id = br.annual_budget_id
			WHERE br.company_id=$1 AND ab.fiscal_year=$2
			ORDER BY br.created_at DESC`, companyID, fiscalYear)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		list := []map[string]interface{}{}
		for rows.Next() {
			var revNum, budgetNum, budgetName, revType, reason, status, effDate string
			var fy int
			var origAmt, revisedAmt, changeAmt float64
			var createdAt time.Time
			rows.Scan(&revNum, &budgetNum, &budgetName, &revType, &origAmt, &revisedAmt, &changeAmt,
				&reason, &status, &effDate, &createdAt, &fy)
			list = append(list, map[string]interface{}{
				"revision_number": revNum, "budget_number": budgetNum, "budget_name": budgetName,
				"revision_type": revType, "original_amount": origAmt,
				"revised_amount": revisedAmt, "change_amount": changeAmt,
				"reason": reason, "status": status, "effective_date": effDate,
				"fiscal_year": fy, "created_at": createdAt.Format(time.RFC3339),
			})
		}
		c.JSON(http.StatusOK, gin.H{"type": "revision_history", "fiscal_year": fiscalYear, "data": list})

	default:
		// Overview report
		row := h.db.QueryRow(ctx, `
			SELECT
				COUNT(DISTINCT ab.id)                       AS total_budgets,
				COALESCE(SUM(ab.total_amount),0)            AS total_allocated,
				COALESCE(SUM(bli.actual_sum),0)             AS total_spent,
				COALESCE(SUM(comm.committed_sum),0)         AS total_committed,
				COUNT(DISTINCT ab.id) FILTER (WHERE ab.status='active') AS active_budgets
			FROM annual_budgets ab
			LEFT JOIN (
				SELECT annual_budget_id, SUM(actual_amount) AS actual_sum
				FROM budget_line_items GROUP BY annual_budget_id
			) bli ON bli.annual_budget_id = ab.id
			LEFT JOIN (
				SELECT annual_budget_id, SUM(committed_amount) AS committed_sum
				FROM budget_commitments WHERE status IN ('pending','approved')
				GROUP BY annual_budget_id
			) comm ON comm.annual_budget_id = ab.id
			WHERE ab.company_id=$1 AND ab.fiscal_year=$2`, companyID, fiscalYear)

		var totalBudgets, activeBudgets int64
		var totalAllocated, totalSpent, totalCommitted float64
		row.Scan(&totalBudgets, &totalAllocated, &totalSpent, &totalCommitted, &activeBudgets)
		pct := 0.0
		if totalAllocated > 0 {
			pct = (totalSpent / totalAllocated) * 100
		}
		c.JSON(http.StatusOK, gin.H{
			"type": "overview", "fiscal_year": fiscalYear,
			"data": map[string]interface{}{
				"total_budgets": totalBudgets, "active_budgets": activeBudgets,
				"total_allocated": totalAllocated, "total_spent": totalSpent,
				"total_committed": totalCommitted,
				"available": totalAllocated - totalSpent - totalCommitted,
				"utilization_pct": pct,
			},
		})
	}
}
