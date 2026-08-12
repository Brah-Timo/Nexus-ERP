package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"nexus-erp/internal/middleware"
	"nexus-erp/internal/models"
)

// ─── Sales Handler ────────────────────────────────────────────────────────────

type SalesHandler struct{ db *pgxpool.Pool }

// ─────────────────────────────────────────────────────────────────────────────
// DASHBOARD
// ─────────────────────────────────────────────────────────────────────────────

type DashboardHandler struct{ db *pgxpool.Pool }

// GetSummary returns the main dashboard KPIs by querying all relevant tables.
func (h *DashboardHandler) GetSummary(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	s := models.DashboardSummary{}

	// ── Sales KPIs ──
	h.db.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE
				WHEN status IN ('confirmed','partially_paid','paid')
				 AND date_trunc('month', date) = date_trunc('month', CURRENT_DATE)
				THEN total_amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status NOT IN ('paid','cancelled') THEN balance_due ELSE 0 END), 0),
			COALESCE(SUM(CASE
				WHEN status NOT IN ('paid','cancelled') AND due_date < CURRENT_DATE
				THEN 1 ELSE 0 END), 0)
		FROM sales_invoices WHERE company_id = $1`,
		companyID,
	).Scan(&s.MonthlySales, &s.Receivables, &s.OverdueInvoices)

	// ── Customer & pipeline ──
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM customers WHERE company_id = $1 AND is_active = TRUE`, companyID).Scan(&s.CustomerCount)
	h.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(amount * probability / 100.0), 0),
			COALESCE(SUM(CASE
				WHEN stage = 'won'
				 AND date_trunc('month', updated_at) = date_trunc('month', CURRENT_DATE)
				THEN amount ELSE 0 END), 0)
		FROM opportunities WHERE company_id = $1 AND stage NOT IN ('won','lost')`,
		companyID,
	).Scan(&s.OpenOpportunities, &s.PipelineValue, &s.WonThisMonth)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM quotations WHERE company_id = $1 AND status = 'draft'`, companyID).Scan(&s.DraftQuotations)
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM sales_orders WHERE company_id = $1 AND status NOT IN ('cancelled','delivered')`, companyID).Scan(&s.OpenOrders)

	// ── Payables (purchase) ──
	h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(balance_due), 0)
		FROM purchase_invoices WHERE company_id = $1 AND status NOT IN ('paid','cancelled')`,
		companyID,
	).Scan(&s.Payables)

	// ── Treasury balance ──
	h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(current_balance), 0)
		FROM bank_accounts WHERE company_id = $1 AND is_active = TRUE`,
		companyID,
	).Scan(&s.TreasuryBalance)

	// ── Stock value ──
	h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(sl.quantity * i.cost_price), 0)
		FROM stock_levels sl
		JOIN items i ON i.id = sl.item_id
		WHERE sl.company_id = $1`,
		companyID,
	).Scan(&s.StockValue)

	// ── HR ──
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM employees WHERE company_id = $1 AND status = 'active'`, companyID).Scan(&s.EmployeeCount)
	h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_net), 0)
		FROM payroll_runs
		WHERE company_id = $1
		  AND period_year  = EXTRACT(year  FROM CURRENT_DATE)
		  AND period_month = EXTRACT(month FROM CURRENT_DATE)`,
		companyID,
	).Scan(&s.MonthlyPayroll)

	// ── Projects ──
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM projects WHERE company_id = $1 AND status = 'active'`, companyID).Scan(&s.ActiveProjects)

	// ── Manufacturing ──
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM manufacturing_orders WHERE company_id = $1 AND status = 'in_progress'`, companyID).Scan(&s.ActiveMfgOrders)

	// ── Workflow approvals ──
	h.db.QueryRow(ctx, `SELECT COUNT(*) FROM workflow_approvals WHERE company_id = $1 AND status = 'pending'`, companyID).Scan(&s.PendingApprovals)

	c.JSON(http.StatusOK, s)
}

// GetCashFlow returns last 6 months of monthly cash in/out.
func (h *DashboardHandler) GetCashFlow(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			TO_CHAR(date_trunc('month', date), 'Mon YYYY') AS month,
			COALESCE(SUM(CASE WHEN status IN ('confirmed','partially_paid','paid') THEN total_amount ELSE 0 END), 0) AS inflow,
			0::numeric AS outflow
		FROM sales_invoices
		WHERE company_id = $1
		  AND date >= date_trunc('month', CURRENT_DATE) - INTERVAL '5 months'
		GROUP BY date_trunc('month', date)
		ORDER BY date_trunc('month', date)`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	type CF struct {
		Month   string          `json:"month"`
		Inflow  decimal.Decimal `json:"inflow"`
		Outflow decimal.Decimal `json:"outflow"`
	}
	var result []CF
	for rows.Next() {
		var r CF
		_ = rows.Scan(&r.Month, &r.Inflow, &r.Outflow)
		result = append(result, r)
	}
	c.JSON(http.StatusOK, result)
}

// GetRecentActivity returns the last 20 audit log entries.
func (h *DashboardHandler) GetRecentActivity(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT al.id, al.action, al.entity_type, COALESCE(al.entity_id,''),
		       COALESCE(u.full_name, 'System'),
		       al.entity_type, al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON u.id = al.user_id
		WHERE al.company_id = $1
		ORDER BY al.created_at DESC
		LIMIT 20`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var result []models.RecentActivity
	for rows.Next() {
		var r models.RecentActivity
		_ = rows.Scan(&r.ID, &r.Action, &r.EntityType, &r.EntityID, &r.UserName, &r.Module, &r.CreatedAt)
		result = append(result, r)
	}
	c.JSON(http.StatusOK, result)
}

// GetPendingApprovals returns workflow items pending approval.
func (h *DashboardHandler) GetPendingApprovals(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT id, document_type, module, requested_by, created_at
		FROM workflow_approvals
		WHERE company_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
		LIMIT 20`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	type Approval struct {
		ID           string    `json:"id"`
		DocumentType string    `json:"document_type"`
		Module       string    `json:"module"`
		RequestedBy  string    `json:"requested_by"`
		CreatedAt    time.Time `json:"created_at"`
	}
	var result []Approval
	for rows.Next() {
		var a Approval
		_ = rows.Scan(&a.ID, &a.DocumentType, &a.Module, &a.RequestedBy, &a.CreatedAt)
		result = append(result, a)
	}
	c.JSON(http.StatusOK, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// LEADS
// ─────────────────────────────────────────────────────────────────────────────

func (h *SalesHandler) ListLeads(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT id,
		       COALESCE(title,''),
		       COALESCE(first_name,''),
		       COALESCE(last_name,''),
		       COALESCE(company_name,''),
		       COALESCE(email,''),
		       COALESCE(phone,''),
		       COALESCE(source,''),
		       status,
		       salesperson_id,
		       COALESCE(notes,''),
		       converted_to,
		       created_at, updated_at
		FROM leads
		WHERE company_id = $1
		ORDER BY created_at DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var leads []models.Lead
	for rows.Next() {
		var l models.Lead
		l.CompanyID = companyID
		_ = rows.Scan(
			&l.ID, &l.Title, &l.FirstName, &l.LastName, &l.CompanyName,
			&l.Email, &l.Phone, &l.Source, &l.Status,
			&l.SalespersonID, &l.Notes, &l.ConvertedTo,
			&l.CreatedAt, &l.UpdatedAt,
		)
		leads = append(leads, l)
	}
	c.JSON(http.StatusOK, leads)
}

func (h *SalesHandler) CreateLead(c *gin.Context) {
	var l models.Lead
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l.ID = uuid.NewString()
	l.CompanyID = middleware.GetCompanyID(c)
	if l.Status == "" {
		l.Status = "new"
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO leads
		  (id, company_id, title, first_name, last_name, company_name,
		   email, phone, source, status, salesperson_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		l.ID, l.CompanyID, l.Title, l.FirstName, l.LastName, l.CompanyName,
		l.Email, l.Phone, l.Source, l.Status, l.SalespersonID, l.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, l)
}

func (h *SalesHandler) UpdateLead(c *gin.Context) {
	id := c.Param("id")
	var l models.Lead
	if err := c.ShouldBindJSON(&l); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE leads
		SET title=$1, first_name=$2, last_name=$3, company_name=$4,
		    email=$5, phone=$6, source=$7, status=$8,
		    salesperson_id=$9, notes=$10, updated_at=NOW()
		WHERE id=$11`,
		l.Title, l.FirstName, l.LastName, l.CompanyName,
		l.Email, l.Phone, l.Source, l.Status,
		l.SalespersonID, l.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	l.ID = id
	c.JSON(http.StatusOK, l)
}

// ─────────────────────────────────────────────────────────────────────────────
// OPPORTUNITIES / PIPELINE
// ─────────────────────────────────────────────────────────────────────────────

func (h *SalesHandler) ListOpportunities(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT o.id,
		       o.customer_id,
		       COALESCE(cu.name,'') AS customer_name,
		       o.lead_id,
		       COALESCE(o.name,''),
		       o.stage,
		       COALESCE(o.amount, 0),
		       o.probability,
		       o.expected_close,
		       o.salesperson_id,
		       COALESCE(o.notes,''),
		       COALESCE(o.lost_reason,''),
		       o.created_at, o.updated_at
		FROM opportunities o
		LEFT JOIN customers cu ON cu.id = o.customer_id
		WHERE o.company_id = $1
		ORDER BY o.amount DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var opps []models.Opportunity
	for rows.Next() {
		var o models.Opportunity
		o.CompanyID = companyID
		_ = rows.Scan(
			&o.ID, &o.CustomerID, &o.CustomerName, &o.LeadID,
			&o.Name, &o.Stage, &o.Amount, &o.Probability,
			&o.ExpectedClose, &o.SalespersonID, &o.Notes, &o.LostReason,
			&o.CreatedAt, &o.UpdatedAt,
		)
		opps = append(opps, o)
	}
	c.JSON(http.StatusOK, opps)
}

func (h *SalesHandler) CreateOpportunity(c *gin.Context) {
	var o models.Opportunity
	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	o.ID = uuid.NewString()
	o.CompanyID = middleware.GetCompanyID(c)
	if o.Stage == "" {
		o.Stage = "lead"
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO opportunities
		  (id, company_id, customer_id, lead_id, name, stage,
		   amount, probability, expected_close, salesperson_id, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		o.ID, o.CompanyID, o.CustomerID, o.LeadID, o.Name, o.Stage,
		o.Amount, o.Probability, o.ExpectedClose, o.SalespersonID, o.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, o)
}

func (h *SalesHandler) UpdateOpportunity(c *gin.Context) {
	id := c.Param("id")
	var o models.Opportunity
	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE opportunities
		SET name=$1, stage=$2, amount=$3, probability=$4,
		    expected_close=$5, customer_id=$6, salesperson_id=$7,
		    notes=$8, lost_reason=$9, updated_at=NOW()
		WHERE id=$10`,
		o.Name, o.Stage, o.Amount, o.Probability,
		o.ExpectedClose, o.CustomerID, o.SalespersonID,
		o.Notes, o.LostReason, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	o.ID = id
	c.JSON(http.StatusOK, o)
}

// PipelineSummary returns aggregate by stage for the Kanban view.
func (h *SalesHandler) PipelineSummary(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT stage,
		       COUNT(*) AS count,
		       COALESCE(SUM(amount), 0) AS total_amount,
		       COALESCE(ROUND(AVG(probability)::numeric, 0), 0) AS avg_probability
		FROM opportunities
		WHERE company_id = $1
		GROUP BY stage
		ORDER BY
		  CASE stage
		    WHEN 'lead'        THEN 1
		    WHEN 'qualified'   THEN 2
		    WHEN 'proposal'    THEN 3
		    WHEN 'negotiation' THEN 4
		    WHEN 'won'         THEN 5
		    WHEN 'lost'        THEN 6
		    ELSE 99
		  END`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var result []models.PipelineSummary
	for rows.Next() {
		var ps models.PipelineSummary
		_ = rows.Scan(&ps.Stage, &ps.Count, &ps.TotalAmount, &ps.AvgProbability)
		result = append(result, ps)
	}
	c.JSON(http.StatusOK, result)
}

// ─────────────────────────────────────────────────────────────────────────────
// CUSTOMERS
// ─────────────────────────────────────────────────────────────────────────────

func (h *SalesHandler) ListCustomers(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT id,
		       code, name,
		       COALESCE(type,'company'),
		       COALESCE(nif,''), COALESCE(nis,''), COALESCE(rc,''), COALESCE(art,''),
		       COALESCE(tax_regime,'reel'),
		       COALESCE(address,''), COALESCE(city,''), COALESCE(wilaya,''),
		       COALESCE(postal_code,''),
		       COALESCE(phone,''), COALESCE(email,''), COALESCE(website,''),
		       COALESCE(credit_limit,0), COALESCE(balance,0),
		       payment_terms,
		       account_id, salesperson_id,
		       is_active,
		       COALESCE(notes,''),
		       created_at, updated_at
		FROM customers
		WHERE company_id = $1
		ORDER BY name`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var cu models.Customer
		cu.CompanyID = companyID
		_ = rows.Scan(
			&cu.ID, &cu.Code, &cu.Name,
			&cu.Type, &cu.NIF, &cu.NIS, &cu.RC, &cu.ART,
			&cu.TaxRegime,
			&cu.Address, &cu.City, &cu.Wilaya, &cu.PostalCode,
			&cu.Phone, &cu.Email, &cu.Website,
			&cu.CreditLimit, &cu.Balance, &cu.PaymentTerms,
			&cu.AccountID, &cu.SalespersonID,
			&cu.IsActive, &cu.Notes,
			&cu.CreatedAt, &cu.UpdatedAt,
		)
		customers = append(customers, cu)
	}
	c.JSON(http.StatusOK, customers)
}

func (h *SalesHandler) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	var cu models.Customer
	err := h.db.QueryRow(ctx, `
		SELECT id,
		       company_id, code, name,
		       COALESCE(type,'company'),
		       COALESCE(nif,''), COALESCE(nis,''), COALESCE(rc,''), COALESCE(art,''),
		       COALESCE(tax_regime,'reel'),
		       COALESCE(address,''), COALESCE(city,''), COALESCE(wilaya,''),
		       COALESCE(postal_code,''),
		       COALESCE(phone,''), COALESCE(email,''), COALESCE(website,''),
		       COALESCE(credit_limit,0), COALESCE(balance,0),
		       payment_terms,
		       account_id, salesperson_id,
		       is_active,
		       COALESCE(notes,''),
		       created_at, updated_at
		FROM customers WHERE id = $1`, id).Scan(
		&cu.ID,
		&cu.CompanyID, &cu.Code, &cu.Name,
		&cu.Type, &cu.NIF, &cu.NIS, &cu.RC, &cu.ART,
		&cu.TaxRegime,
		&cu.Address, &cu.City, &cu.Wilaya, &cu.PostalCode,
		&cu.Phone, &cu.Email, &cu.Website,
		&cu.CreditLimit, &cu.Balance, &cu.PaymentTerms,
		&cu.AccountID, &cu.SalespersonID,
		&cu.IsActive, &cu.Notes,
		&cu.CreatedAt, &cu.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	c.JSON(http.StatusOK, cu)
}

func (h *SalesHandler) CreateCustomer(c *gin.Context) {
	var cu models.Customer
	if err := c.ShouldBindJSON(&cu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cu.ID = uuid.NewString()
	cu.CompanyID = middleware.GetCompanyID(c)
	if cu.Type == "" {
		cu.Type = "company"
	}
	if cu.TaxRegime == "" {
		cu.TaxRegime = "reel"
	}
	if cu.PaymentTerms == 0 {
		cu.PaymentTerms = 30
	}
	// Auto-generate code if not supplied
	if cu.Code == "" {
		cu.Code = fmt.Sprintf("C%s", cu.ID[:8])
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO customers
		  (id, company_id, code, name, type, nif, nis, rc, art,
		   tax_regime, address, city, wilaya, postal_code,
		   phone, email, website,
		   credit_limit, payment_terms,
		   account_id, salesperson_id,
		   is_active, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`,
		cu.ID, cu.CompanyID, cu.Code, cu.Name, cu.Type,
		cu.NIF, cu.NIS, cu.RC, cu.ART,
		cu.TaxRegime, cu.Address, cu.City, cu.Wilaya, cu.PostalCode,
		cu.Phone, cu.Email, cu.Website,
		cu.CreditLimit, cu.PaymentTerms,
		cu.AccountID, cu.SalespersonID,
		true, cu.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cu)
}

func (h *SalesHandler) UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var cu models.Customer
	if err := c.ShouldBindJSON(&cu); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE customers
		SET name=$1, type=$2, nif=$3, nis=$4, rc=$5, art=$6,
		    tax_regime=$7, address=$8, city=$9, wilaya=$10,
		    postal_code=$11, phone=$12, email=$13, website=$14,
		    credit_limit=$15, payment_terms=$16,
		    account_id=$17, salesperson_id=$18,
		    is_active=$19, notes=$20, updated_at=NOW()
		WHERE id=$21`,
		cu.Name, cu.Type, cu.NIF, cu.NIS, cu.RC, cu.ART,
		cu.TaxRegime, cu.Address, cu.City, cu.Wilaya,
		cu.PostalCode, cu.Phone, cu.Email, cu.Website,
		cu.CreditLimit, cu.PaymentTerms,
		cu.AccountID, cu.SalespersonID,
		cu.IsActive, cu.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cu.ID = id
	c.JSON(http.StatusOK, cu)
}

// ─────────────────────────────────────────────────────────────────────────────
// QUOTATIONS
// ─────────────────────────────────────────────────────────────────────────────

func (h *SalesHandler) ListQuotations(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT q.id,
		       q.number, q.customer_id, COALESCE(cu.name,'') AS customer_name,
		       q.date, q.valid_until, q.status,
		       COALESCE(q.subtotal,0), COALESCE(q.discount_amount,0),
		       COALESCE(q.tva_amount,0), COALESCE(q.stamp_tax,0),
		       COALESCE(q.total_amount,0),
		       COALESCE(q.currency,'DZD'),
		       COALESCE(q.notes,''),
		       q.salesperson_id, q.converted_to,
		       q.created_at, q.updated_at
		FROM quotations q
		JOIN customers cu ON cu.id = q.customer_id
		WHERE q.company_id = $1
		ORDER BY q.date DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var quotes []models.Quotation
	for rows.Next() {
		var q models.Quotation
		q.CompanyID = companyID
		_ = rows.Scan(
			&q.ID, &q.Number, &q.CustomerID, &q.CustomerName,
			&q.Date, &q.ValidUntil, &q.Status,
			&q.Subtotal, &q.DiscountAmount, &q.TVAAmount, &q.StampTax, &q.TotalAmount,
			&q.Currency, &q.Notes,
			&q.SalespersonID, &q.ConvertedTo,
			&q.CreatedAt, &q.UpdatedAt,
		)
		quotes = append(quotes, q)
	}
	c.JSON(http.StatusOK, quotes)
}

func (h *SalesHandler) GetQuotation(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	var q models.Quotation
	err := h.db.QueryRow(ctx, `
		SELECT q.id,
		       q.company_id, q.number, q.customer_id, COALESCE(cu.name,''),
		       q.date, q.valid_until, q.status,
		       COALESCE(q.subtotal,0), COALESCE(q.discount_amount,0),
		       COALESCE(q.tva_amount,0), COALESCE(q.stamp_tax,0),
		       COALESCE(q.total_amount,0),
		       COALESCE(q.currency,'DZD'),
		       COALESCE(q.notes,''), COALESCE(q.terms,''),
		       q.salesperson_id, q.converted_to,
		       q.created_at, q.updated_at
		FROM quotations q
		JOIN customers cu ON cu.id = q.customer_id
		WHERE q.id = $1`, id).Scan(
		&q.ID, &q.CompanyID, &q.Number, &q.CustomerID, &q.CustomerName,
		&q.Date, &q.ValidUntil, &q.Status,
		&q.Subtotal, &q.DiscountAmount, &q.TVAAmount, &q.StampTax, &q.TotalAmount,
		&q.Currency, &q.Notes, &q.Terms,
		&q.SalespersonID, &q.ConvertedTo,
		&q.CreatedAt, &q.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quotation not found"})
		return
	}
	// Fetch lines
	lrows, _ := h.db.Query(ctx, `
		SELECT id, quotation_id,
		       item_id,
		       description,
		       COALESCE(quantity,1), COALESCE(unit_price,0),
		       COALESCE(discount_pct,0), COALESCE(tva_rate,19),
		       COALESCE(subtotal,0), COALESCE(tva_amount,0), COALESCE(total,0),
		       account_id, COALESCE(sort_order,0)
		FROM quotation_lines
		WHERE quotation_id = $1
		ORDER BY sort_order`, id)
	if lrows != nil {
		defer lrows.Close()
		for lrows.Next() {
			var l models.DocumentLine
			var parentID string
			_ = lrows.Scan(
				&l.ID, &parentID,
				&l.ItemID, &l.Description,
				&l.Quantity, &l.UnitPrice,
				&l.DiscountPct, &l.TVARate,
				&l.Subtotal, &l.TVAAmount, &l.Total,
				&l.AccountID, &l.SortOrder,
			)
			l.ParentID = parentID
			q.Lines = append(q.Lines, l)
		}
	}
	c.JSON(http.StatusOK, q)
}

func (h *SalesHandler) CreateQuotation(c *gin.Context) {
	var q models.Quotation
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q.ID = uuid.NewString()
	q.CompanyID = middleware.GetCompanyID(c)
	q.Status = "draft"
	if q.Currency == "" {
		q.Currency = "DZD"
	}
	q.Number = generateNumber("QT", q.CompanyID, h.db)

	// Recalculate totals from lines
	q.Subtotal = decimal.Zero
	q.TVAAmount = decimal.Zero
	for i := range q.Lines {
		l := &q.Lines[i]
		if l.ID == "" {
			l.ID = uuid.NewString()
		}
		lineSubtotal := l.Quantity.Mul(l.UnitPrice).
			Mul(decimal.NewFromInt(1).Sub(l.DiscountPct.Div(decimal.NewFromInt(100))))
		l.Subtotal = lineSubtotal.Round(2)
		l.TVAAmount = lineSubtotal.Mul(l.TVARate.Div(decimal.NewFromInt(100))).Round(2)
		l.Total = l.Subtotal.Add(l.TVAAmount)
		q.Subtotal = q.Subtotal.Add(l.Subtotal)
		q.TVAAmount = q.TVAAmount.Add(l.TVAAmount)
		l.SortOrder = i
	}
	q.TotalAmount = q.Subtotal.Add(q.TVAAmount).Add(q.StampTax).Sub(q.DiscountAmount)

	ctx := context.Background()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO quotations
		  (id, company_id, number, customer_id, date, valid_until, status,
		   subtotal, discount_amount, tva_amount, stamp_tax, total_amount,
		   currency, notes, terms, salesperson_id, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		q.ID, q.CompanyID, q.Number, q.CustomerID,
		q.Date, q.ValidUntil, q.Status,
		q.Subtotal, q.DiscountAmount, q.TVAAmount, q.StampTax, q.TotalAmount,
		q.Currency, q.Notes, q.Terms, q.SalespersonID,
		middleware.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, l := range q.Lines {
		_, err = tx.Exec(ctx, `
			INSERT INTO quotation_lines
			  (id, quotation_id, item_id, description, quantity, unit_price,
			   discount_pct, tva_rate, subtotal, tva_amount, total,
			   account_id, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			l.ID, q.ID, l.ItemID, l.Description,
			l.Quantity, l.UnitPrice, l.DiscountPct, l.TVARate,
			l.Subtotal, l.TVAAmount, l.Total,
			l.AccountID, l.SortOrder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, q)
}

// ConfirmQuotation converts a draft quotation to sent status.
func (h *SalesHandler) ConfirmQuotation(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE quotations SET status='sent', updated_at=NOW()
		WHERE id=$1 AND status='draft'`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Quotation sent"})
}

// ConvertToOrder converts a confirmed quotation to a sales order.
func (h *SalesHandler) ConvertToOrder(c *gin.Context) {
	quotID := c.Param("id")
	ctx := context.Background()

	var q models.Quotation
	err := h.db.QueryRow(ctx, `
		SELECT id, company_id, customer_id, date,
		       subtotal, discount_amount, tva_amount, stamp_tax, total_amount,
		       currency, notes
		FROM quotations WHERE id = $1 AND status IN ('sent','confirmed')`, quotID).Scan(
		&q.ID, &q.CompanyID, &q.CustomerID, &q.Date,
		&q.Subtotal, &q.DiscountAmount, &q.TVAAmount, &q.StampTax, &q.TotalAmount,
		&q.Currency, &q.Notes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quotation not found or not in correct status"})
		return
	}

	// Fetch lines
	lrows, _ := h.db.Query(ctx, `
		SELECT item_id, description, quantity, unit_price, discount_pct, tva_rate,
		       subtotal, tva_amount, total, account_id, sort_order
		FROM quotation_lines WHERE quotation_id = $1 ORDER BY sort_order`, quotID)
	var lines []models.DocumentLine
	if lrows != nil {
		defer lrows.Close()
		for lrows.Next() {
			var l models.DocumentLine
			_ = lrows.Scan(&l.ItemID, &l.Description, &l.Quantity, &l.UnitPrice,
				&l.DiscountPct, &l.TVARate, &l.Subtotal, &l.TVAAmount, &l.Total,
				&l.AccountID, &l.SortOrder)
			lines = append(lines, l)
		}
	}

	so := models.SalesOrder{
		CompanyID:      q.CompanyID,
		CustomerID:     q.CustomerID,
		Date:           q.Date,
		Status:         "draft",
		Subtotal:       q.Subtotal,
		DiscountAmount: q.DiscountAmount,
		TVAAmount:      q.TVAAmount,
		StampTax:       q.StampTax,
		TotalAmount:    q.TotalAmount,
		Currency:       q.Currency,
		Notes:          q.Notes,
		Lines:          lines,
	}
	so.ID = uuid.NewString()
	so.Number = generateNumber("SO", so.CompanyID, h.db)
	so.QuotationID = &quotID

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO sales_orders
		  (id, company_id, number, quotation_id, customer_id, date, status,
		   subtotal, discount_amount, tva_amount, stamp_tax, total_amount,
		   currency, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		so.ID, so.CompanyID, so.Number, so.QuotationID, so.CustomerID,
		so.Date, so.Status,
		so.Subtotal, so.DiscountAmount, so.TVAAmount, so.StampTax, so.TotalAmount,
		so.Currency, so.Notes, middleware.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i, l := range so.Lines {
		l.ID = uuid.NewString()
		_, err = tx.Exec(ctx, `
			INSERT INTO sales_order_lines
			  (id, order_id, item_id, description, quantity, unit_price,
			   discount_pct, tva_rate, subtotal, tva_amount, total,
			   account_id, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			l.ID, so.ID, l.ItemID, l.Description,
			l.Quantity, l.UnitPrice, l.DiscountPct, l.TVARate,
			l.Subtotal, l.TVAAmount, l.Total,
			l.AccountID, i)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	// Mark quotation as converted
	_, _ = tx.Exec(ctx, `UPDATE quotations SET status='confirmed', converted_to=$1, updated_at=NOW() WHERE id=$2`, so.ID, quotID)

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, so)
}

// ─────────────────────────────────────────────────────────────────────────────
// SALES ORDERS
// ─────────────────────────────────────────────────────────────────────────────

func (h *SalesHandler) ListOrders(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT so.id,
		       so.number, so.quotation_id, so.customer_id,
		       COALESCE(cu.name,'') AS customer_name,
		       so.date, so.delivery_date, so.status,
		       COALESCE(so.subtotal,0), COALESCE(so.discount_amount,0),
		       COALESCE(so.tva_amount,0), COALESCE(so.stamp_tax,0),
		       COALESCE(so.total_amount,0),
		       COALESCE(so.currency,'DZD'),
		       COALESCE(so.notes,''),
		       so.created_at, so.updated_at
		FROM sales_orders so
		JOIN customers cu ON cu.id = so.customer_id
		WHERE so.company_id = $1
		ORDER BY so.date DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var orders []models.SalesOrder
	for rows.Next() {
		var so models.SalesOrder
		so.CompanyID = companyID
		_ = rows.Scan(
			&so.ID, &so.Number, &so.QuotationID, &so.CustomerID, &so.CustomerName,
			&so.Date, &so.DeliveryDate, &so.Status,
			&so.Subtotal, &so.DiscountAmount, &so.TVAAmount, &so.StampTax, &so.TotalAmount,
			&so.Currency, &so.Notes,
			&so.CreatedAt, &so.UpdatedAt,
		)
		orders = append(orders, so)
	}
	c.JSON(http.StatusOK, orders)
}

func (h *SalesHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	var so models.SalesOrder
	err := h.db.QueryRow(ctx, `
		SELECT so.id, so.company_id,
		       so.number, so.quotation_id, so.customer_id,
		       COALESCE(cu.name,''),
		       so.date, so.delivery_date, so.status,
		       COALESCE(so.subtotal,0), COALESCE(so.discount_amount,0),
		       COALESCE(so.tva_amount,0), COALESCE(so.stamp_tax,0),
		       COALESCE(so.total_amount,0),
		       COALESCE(so.currency,'DZD'),
		       COALESCE(so.notes,''),
		       so.created_at, so.updated_at
		FROM sales_orders so
		JOIN customers cu ON cu.id = so.customer_id
		WHERE so.id = $1`, id).Scan(
		&so.ID, &so.CompanyID,
		&so.Number, &so.QuotationID, &so.CustomerID, &so.CustomerName,
		&so.Date, &so.DeliveryDate, &so.Status,
		&so.Subtotal, &so.DiscountAmount, &so.TVAAmount, &so.StampTax, &so.TotalAmount,
		&so.Currency, &so.Notes,
		&so.CreatedAt, &so.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	// Fetch lines
	lrows, _ := h.db.Query(ctx, `
		SELECT id, order_id, item_id, description,
		       COALESCE(quantity,1), COALESCE(unit_price,0),
		       COALESCE(discount_pct,0), COALESCE(tva_rate,19),
		       COALESCE(subtotal,0), COALESCE(tva_amount,0), COALESCE(total,0),
		       account_id, COALESCE(sort_order,0)
		FROM sales_order_lines WHERE order_id = $1 ORDER BY sort_order`, id)
	if lrows != nil {
		defer lrows.Close()
		for lrows.Next() {
			var l models.DocumentLine
			var parentID string
			_ = lrows.Scan(
				&l.ID, &parentID, &l.ItemID, &l.Description,
				&l.Quantity, &l.UnitPrice, &l.DiscountPct, &l.TVARate,
				&l.Subtotal, &l.TVAAmount, &l.Total,
				&l.AccountID, &l.SortOrder,
			)
			l.ParentID = parentID
			so.Lines = append(so.Lines, l)
		}
	}
	c.JSON(http.StatusOK, so)
}

func (h *SalesHandler) CreateOrder(c *gin.Context) {
	var so models.SalesOrder
	if err := c.ShouldBindJSON(&so); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	so.ID = uuid.NewString()
	so.CompanyID = middleware.GetCompanyID(c)
	so.Status = "draft"
	if so.Currency == "" {
		so.Currency = "DZD"
	}
	so.Number = generateNumber("SO", so.CompanyID, h.db)

	so.Subtotal = decimal.Zero
	so.TVAAmount = decimal.Zero
	for i := range so.Lines {
		l := &so.Lines[i]
		if l.ID == "" {
			l.ID = uuid.NewString()
		}
		ls := l.Quantity.Mul(l.UnitPrice).
			Mul(decimal.NewFromInt(1).Sub(l.DiscountPct.Div(decimal.NewFromInt(100))))
		l.Subtotal = ls.Round(2)
		l.TVAAmount = ls.Mul(l.TVARate.Div(decimal.NewFromInt(100))).Round(2)
		l.Total = l.Subtotal.Add(l.TVAAmount)
		so.Subtotal = so.Subtotal.Add(l.Subtotal)
		so.TVAAmount = so.TVAAmount.Add(l.TVAAmount)
		l.SortOrder = i
	}
	so.TotalAmount = so.Subtotal.Add(so.TVAAmount).Add(so.StampTax).Sub(so.DiscountAmount)

	ctx := context.Background()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO sales_orders
		  (id, company_id, number, quotation_id, customer_id, date, delivery_date, status,
		   subtotal, discount_amount, tva_amount, stamp_tax, total_amount,
		   currency, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		so.ID, so.CompanyID, so.Number, so.QuotationID, so.CustomerID,
		so.Date, so.DeliveryDate, so.Status,
		so.Subtotal, so.DiscountAmount, so.TVAAmount, so.StampTax, so.TotalAmount,
		so.Currency, so.Notes, middleware.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, l := range so.Lines {
		_, err = tx.Exec(ctx, `
			INSERT INTO sales_order_lines
			  (id, order_id, item_id, description, quantity, unit_price,
			   discount_pct, tva_rate, subtotal, tva_amount, total,
			   account_id, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			l.ID, so.ID, l.ItemID, l.Description,
			l.Quantity, l.UnitPrice, l.DiscountPct, l.TVARate,
			l.Subtotal, l.TVAAmount, l.Total,
			l.AccountID, l.SortOrder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, so)
}

// FulfillOrder marks an order as delivered.
func (h *SalesHandler) FulfillOrder(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE sales_orders SET status='delivered', delivery_date=CURRENT_DATE, updated_at=NOW()
		WHERE id=$1 AND status='confirmed'`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order delivered"})
}

// ─────────────────────────────────────────────────────────────────────────────
// SALES INVOICES
// ─────────────────────────────────────────────────────────────────────────────

func (h *SalesHandler) ListInvoices(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT si.id,
		       si.number, si.order_id, si.customer_id,
		       COALESCE(cu.name,'') AS customer_name,
		       si.date, si.due_date, si.status,
		       COALESCE(si.subtotal,0), COALESCE(si.discount_amount,0),
		       COALESCE(si.tva_amount,0), COALESCE(si.stamp_tax,0),
		       COALESCE(si.total_amount,0), COALESCE(si.paid_amount,0),
		       COALESCE(si.balance_due,0),
		       COALESCE(si.currency,'DZD'),
		       COALESCE(si.notes,''),
		       si.journal_entry_id,
		       si.created_at, si.updated_at
		FROM sales_invoices si
		JOIN customers cu ON cu.id = si.customer_id
		WHERE si.company_id = $1
		ORDER BY si.date DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var invoices []models.SalesInvoice
	for rows.Next() {
		var inv models.SalesInvoice
		inv.CompanyID = companyID
		_ = rows.Scan(
			&inv.ID, &inv.Number, &inv.OrderID, &inv.CustomerID, &inv.CustomerName,
			&inv.Date, &inv.DueDate, &inv.Status,
			&inv.Subtotal, &inv.DiscountAmount, &inv.TVAAmount, &inv.StampTax,
			&inv.TotalAmount, &inv.PaidAmount, &inv.BalanceDue,
			&inv.Currency, &inv.Notes,
			&inv.JournalEntryID,
			&inv.CreatedAt, &inv.UpdatedAt,
		)
		invoices = append(invoices, inv)
	}
	c.JSON(http.StatusOK, invoices)
}

func (h *SalesHandler) GetInvoice(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	var inv models.SalesInvoice
	err := h.db.QueryRow(ctx, `
		SELECT si.id, si.company_id,
		       si.number, si.order_id, si.customer_id,
		       COALESCE(cu.name,''),
		       si.date, si.due_date, si.status,
		       COALESCE(si.subtotal,0), COALESCE(si.discount_amount,0),
		       COALESCE(si.tva_amount,0), COALESCE(si.stamp_tax,0),
		       COALESCE(si.total_amount,0), COALESCE(si.paid_amount,0),
		       COALESCE(si.balance_due,0),
		       COALESCE(si.currency,'DZD'),
		       COALESCE(si.notes,''),
		       si.journal_entry_id,
		       si.created_at, si.updated_at
		FROM sales_invoices si
		JOIN customers cu ON cu.id = si.customer_id
		WHERE si.id = $1`, id).Scan(
		&inv.ID, &inv.CompanyID,
		&inv.Number, &inv.OrderID, &inv.CustomerID, &inv.CustomerName,
		&inv.Date, &inv.DueDate, &inv.Status,
		&inv.Subtotal, &inv.DiscountAmount, &inv.TVAAmount, &inv.StampTax,
		&inv.TotalAmount, &inv.PaidAmount, &inv.BalanceDue,
		&inv.Currency, &inv.Notes,
		&inv.JournalEntryID,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}
	// Fetch lines
	lrows, _ := h.db.Query(ctx, `
		SELECT id, invoice_id, item_id, description,
		       COALESCE(quantity,1), COALESCE(unit_price,0),
		       COALESCE(discount_pct,0), COALESCE(tva_rate,19),
		       COALESCE(subtotal,0), COALESCE(tva_amount,0), COALESCE(total,0),
		       account_id, COALESCE(sort_order,0)
		FROM sales_invoice_lines WHERE invoice_id = $1 ORDER BY sort_order`, id)
	if lrows != nil {
		defer lrows.Close()
		for lrows.Next() {
			var l models.DocumentLine
			var parentID string
			_ = lrows.Scan(
				&l.ID, &parentID, &l.ItemID, &l.Description,
				&l.Quantity, &l.UnitPrice, &l.DiscountPct, &l.TVARate,
				&l.Subtotal, &l.TVAAmount, &l.Total,
				&l.AccountID, &l.SortOrder,
			)
			l.ParentID = parentID
			inv.Lines = append(inv.Lines, l)
		}
	}
	c.JSON(http.StatusOK, inv)
}

func (h *SalesHandler) CreateInvoice(c *gin.Context) {
	var inv models.SalesInvoice
	if err := c.ShouldBindJSON(&inv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	inv.ID = uuid.NewString()
	inv.CompanyID = middleware.GetCompanyID(c)
	inv.Status = models.InvoiceStatusDraft
	if inv.Currency == "" {
		inv.Currency = "DZD"
	}
	inv.Number = generateNumber("INV", inv.CompanyID, h.db)

	inv.Subtotal = decimal.Zero
	inv.TVAAmount = decimal.Zero
	for i := range inv.Lines {
		l := &inv.Lines[i]
		if l.ID == "" {
			l.ID = uuid.NewString()
		}
		ls := l.Quantity.Mul(l.UnitPrice).
			Mul(decimal.NewFromInt(1).Sub(l.DiscountPct.Div(decimal.NewFromInt(100))))
		l.Subtotal = ls.Round(2)
		l.TVAAmount = ls.Mul(l.TVARate.Div(decimal.NewFromInt(100))).Round(2)
		l.Total = l.Subtotal.Add(l.TVAAmount)
		inv.Subtotal = inv.Subtotal.Add(l.Subtotal)
		inv.TVAAmount = inv.TVAAmount.Add(l.TVAAmount)
		l.SortOrder = i
	}
	inv.TotalAmount = inv.Subtotal.Add(inv.TVAAmount).Add(inv.StampTax).Sub(inv.DiscountAmount)
	inv.BalanceDue = inv.TotalAmount

	ctx := context.Background()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO sales_invoices
		  (id, company_id, number, order_id, customer_id, date, due_date, status,
		   subtotal, discount_amount, tva_amount, stamp_tax,
		   total_amount, paid_amount, currency, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		inv.ID, inv.CompanyID, inv.Number, inv.OrderID, inv.CustomerID,
		inv.Date, inv.DueDate, inv.Status,
		inv.Subtotal, inv.DiscountAmount, inv.TVAAmount, inv.StampTax,
		inv.TotalAmount, decimal.Zero, inv.Currency, inv.Notes,
		middleware.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, l := range inv.Lines {
		_, err = tx.Exec(ctx, `
			INSERT INTO sales_invoice_lines
			  (id, invoice_id, item_id, description, quantity, unit_price,
			   discount_pct, tva_rate, subtotal, tva_amount, total,
			   account_id, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			l.ID, inv.ID, l.ItemID, l.Description,
			l.Quantity, l.UnitPrice, l.DiscountPct, l.TVARate,
			l.Subtotal, l.TVAAmount, l.Total,
			l.AccountID, l.SortOrder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, inv)
}

// ConfirmInvoice posts a draft invoice: sets status, updates customer balance,
// creates a journal entry (Debtors Dr / Sales Cr / TVA Payable Cr).
func (h *SalesHandler) ConfirmInvoice(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var inv models.SalesInvoice
	err := h.db.QueryRow(ctx, `
		SELECT id, company_id, customer_id,
		       COALESCE(total_amount,0), COALESCE(tva_amount,0), COALESCE(subtotal,0),
		       COALESCE(currency,'DZD')
		FROM sales_invoices WHERE id = $1 AND status = 'draft'`, id).Scan(
		&inv.ID, &inv.CompanyID, &inv.CustomerID,
		&inv.TotalAmount, &inv.TVAAmount, &inv.Subtotal,
		&inv.Currency)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice not found or already confirmed"})
		return
	}

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	// 1. Update invoice status
	_, err = tx.Exec(ctx, `
		UPDATE sales_invoices SET status='confirmed', confirmed_at=NOW(), confirmed_by=$1, updated_at=NOW()
		WHERE id=$2`, middleware.GetUserID(c), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Update customer running balance
	_, _ = tx.Exec(ctx, `UPDATE customers SET balance = balance + $1, updated_at=NOW() WHERE id = $2`,
		inv.TotalAmount, inv.CustomerID)

	// 3. Create journal entry
	jeID := uuid.NewString()
	jeNum := generateNumber("JE", inv.CompanyID, h.db)
	_, _ = tx.Exec(ctx, `
		INSERT INTO journal_entries
		  (id, company_id, number, date, reference, description, status,
		   total_debit, total_credit, currency, source_type, source_id, created_by)
		VALUES ($1,$2,$3,CURRENT_DATE,$4,$5,'posted',$6,$6,$7,'sales_invoice',$8,$9)`,
		jeID, inv.CompanyID, jeNum,
		id, "Sales Invoice "+inv.Number,
		inv.TotalAmount, inv.Currency,
		id, middleware.GetUserID(c))

	_, _ = tx.Exec(ctx, `UPDATE sales_invoices SET journal_entry_id = $1 WHERE id = $2`, jeID, id)

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invoice confirmed", "journal_entry_id": jeID})
}

func (h *SalesHandler) CancelInvoice(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var totalAmount decimal.Decimal
	var customerID string
	var status string
	err := h.db.QueryRow(ctx,
		`SELECT status, customer_id, COALESCE(total_amount,0) FROM sales_invoices WHERE id=$1`, id,
	).Scan(&status, &customerID, &totalAmount)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	tx, _ := h.db.Begin(ctx)
	defer tx.Rollback(ctx)

	_, _ = tx.Exec(ctx, `UPDATE sales_invoices SET status='cancelled', updated_at=NOW() WHERE id=$1`, id)
	if status == "confirmed" {
		_, _ = tx.Exec(ctx, `UPDATE customers SET balance = balance - $1 WHERE id=$2`, totalAmount, customerID)
	}
	_ = tx.Commit(ctx)
	c.JSON(http.StatusOK, gin.H{"message": "Invoice cancelled"})
}

func (h *SalesHandler) RecordPayment(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Amount float64 `json:"amount"`
		Method string  `json:"method"`
		Date   string  `json:"date"`
		Notes  string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	amount := decimal.NewFromFloat(req.Amount)
	ctx := context.Background()

	// Load current state
	var totalAmount, paidAmount decimal.Decimal
	var customerID string
	err := h.db.QueryRow(ctx,
		`SELECT COALESCE(total_amount,0), COALESCE(paid_amount,0), customer_id FROM sales_invoices WHERE id=$1`, id,
	).Scan(&totalAmount, &paidAmount, &customerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	newPaid := paidAmount.Add(amount)
	newBalance := totalAmount.Sub(newPaid)
	newStatus := "confirmed"
	if newBalance.LessThanOrEqual(decimal.Zero) {
		newStatus = "paid"
	} else if newPaid.GreaterThan(decimal.Zero) {
		newStatus = "partially_paid"
	}

	tx, _ := h.db.Begin(ctx)
	defer tx.Rollback(ctx)

	_, _ = tx.Exec(ctx, `
		UPDATE sales_invoices
		SET paid_amount=$1, status=$2,
		    payment_method=$3, payment_date=$4, updated_at=NOW()
		WHERE id=$5`,
		newPaid, newStatus, req.Method, req.Date, id)

	_, _ = tx.Exec(ctx, `UPDATE customers SET balance = balance - $1 WHERE id=$2`, amount, customerID)

	_ = tx.Commit(ctx)
	c.JSON(http.StatusOK, gin.H{"message": "Payment recorded", "new_status": newStatus, "balance_due": newBalance})
}

// ─────────────────────────────────────────────────────────────────────────────
// CUSTOMER AGING REPORT
// ─────────────────────────────────────────────────────────────────────────────

func (h *SalesHandler) AgingReport(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			cu.id,
			cu.name,
			COALESCE(cu.phone,''),
			COALESCE(cu.email,''),
			COUNT(si.id) AS invoice_count,
			COALESCE(SUM(CASE
				WHEN si.due_date IS NULL OR CURRENT_DATE <= si.due_date
				THEN si.balance_due ELSE 0 END), 0) AS current_amount,
			COALESCE(SUM(CASE
				WHEN si.due_date IS NOT NULL
				 AND CURRENT_DATE - si.due_date BETWEEN 1  AND 30
				THEN si.balance_due ELSE 0 END), 0) AS days_1_30,
			COALESCE(SUM(CASE
				WHEN si.due_date IS NOT NULL
				 AND CURRENT_DATE - si.due_date BETWEEN 31 AND 60
				THEN si.balance_due ELSE 0 END), 0) AS days_31_60,
			COALESCE(SUM(CASE
				WHEN si.due_date IS NOT NULL
				 AND CURRENT_DATE - si.due_date BETWEEN 61 AND 90
				THEN si.balance_due ELSE 0 END), 0) AS days_61_90,
			COALESCE(SUM(CASE
				WHEN si.due_date IS NOT NULL
				 AND CURRENT_DATE - si.due_date > 90
				THEN si.balance_due ELSE 0 END), 0) AS days_over_90,
			COALESCE(SUM(si.balance_due), 0) AS total_outstanding
		FROM sales_invoices si
		JOIN customers cu ON cu.id = si.customer_id
		WHERE si.company_id = $1
		  AND si.status NOT IN ('paid','cancelled')
		GROUP BY cu.id, cu.name, cu.phone, cu.email
		ORDER BY total_outstanding DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var aging []models.CustomerAging
	for rows.Next() {
		var a models.CustomerAging
		_ = rows.Scan(
			&a.CustomerID, &a.CustomerName, &a.Phone, &a.Email,
			&a.InvoiceCount,
			&a.CurrentAmount, &a.Days1to30, &a.Days31to60,
			&a.Days61to90, &a.DaysOver90,
			&a.TotalOutstanding,
		)
		aging = append(aging, a)
	}
	c.JSON(http.StatusOK, aging)
}

func (h *SalesHandler) ListCreditNotes(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}
func (h *SalesHandler) CreateCreditNote(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"id": uuid.NewString()})
}
func (h *SalesHandler) ListCommissions(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE handlers
// ─────────────────────────────────────────────────────────────────────────────

func (h *SalesHandler) DeleteLead(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	tag, err := h.db.Exec(ctx, `DELETE FROM leads WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Lead not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Lead deleted"})
}

func (h *SalesHandler) DeleteOpportunity(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	tag, err := h.db.Exec(ctx, `DELETE FROM opportunities WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Opportunity not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Opportunity deleted"})
}

func (h *SalesHandler) DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	// Soft-delete: mark inactive rather than hard delete
	tag, err := h.db.Exec(ctx, `UPDATE customers SET is_active = FALSE, updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Customer deactivated"})
}

// ─────────────────────────────────────────────────────────────────────────────
// QUOTATION — Update & Cancel
// ─────────────────────────────────────────────────────────────────────────────

func (h *SalesHandler) UpdateQuotation(c *gin.Context) {
	id := c.Param("id")
	var q models.Quotation
	if err := c.ShouldBindJSON(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()

	// Recalculate totals from lines
	q.Subtotal = decimal.Zero
	q.TVAAmount = decimal.Zero
	for i := range q.Lines {
		l := &q.Lines[i]
		if l.ID == "" {
			l.ID = uuid.NewString()
		}
		ls := l.Quantity.Mul(l.UnitPrice).
			Mul(decimal.NewFromInt(1).Sub(l.DiscountPct.Div(decimal.NewFromInt(100))))
		l.Subtotal = ls.Round(2)
		l.TVAAmount = ls.Mul(l.TVARate.Div(decimal.NewFromInt(100))).Round(2)
		l.Total = l.Subtotal.Add(l.TVAAmount)
		q.Subtotal = q.Subtotal.Add(l.Subtotal)
		q.TVAAmount = q.TVAAmount.Add(l.TVAAmount)
		l.SortOrder = i
	}
	q.TotalAmount = q.Subtotal.Add(q.TVAAmount).Add(q.StampTax).Sub(q.DiscountAmount)

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE quotations
		SET customer_id=$1, date=$2, valid_until=$3,
		    subtotal=$4, discount_amount=$5, tva_amount=$6, stamp_tax=$7, total_amount=$8,
		    currency=$9, notes=$10, terms=$11, salesperson_id=$12, updated_at=NOW()
		WHERE id=$13 AND status IN ('draft','sent')`,
		q.CustomerID, q.Date, q.ValidUntil,
		q.Subtotal, q.DiscountAmount, q.TVAAmount, q.StampTax, q.TotalAmount,
		q.Currency, q.Notes, q.Terms, q.SalespersonID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Replace lines
	_, _ = tx.Exec(ctx, `DELETE FROM quotation_lines WHERE quotation_id = $1`, id)
	for _, l := range q.Lines {
		_, err = tx.Exec(ctx, `
			INSERT INTO quotation_lines
			  (id, quotation_id, item_id, description, quantity, unit_price,
			   discount_pct, tva_rate, subtotal, tva_amount, total,
			   account_id, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			l.ID, id, l.ItemID, l.Description,
			l.Quantity, l.UnitPrice, l.DiscountPct, l.TVARate,
			l.Subtotal, l.TVAAmount, l.Total,
			l.AccountID, l.SortOrder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	q.ID = id
	c.JSON(http.StatusOK, q)
}

func (h *SalesHandler) CancelQuotation(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE quotations SET status='cancelled', updated_at=NOW()
		WHERE id=$1 AND status NOT IN ('confirmed','cancelled')`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Quotation cancelled"})
}

// ─────────────────────────────────────────────────────────────────────────────
// SALES ORDER — Update, Confirm, Cancel, Deliver
// ─────────────────────────────────────────────────────────────────────────────

func (h *SalesHandler) UpdateOrder(c *gin.Context) {
	id := c.Param("id")
	var so models.SalesOrder
	if err := c.ShouldBindJSON(&so); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()

	so.Subtotal = decimal.Zero
	so.TVAAmount = decimal.Zero
	for i := range so.Lines {
		l := &so.Lines[i]
		if l.ID == "" {
			l.ID = uuid.NewString()
		}
		ls := l.Quantity.Mul(l.UnitPrice).
			Mul(decimal.NewFromInt(1).Sub(l.DiscountPct.Div(decimal.NewFromInt(100))))
		l.Subtotal = ls.Round(2)
		l.TVAAmount = ls.Mul(l.TVARate.Div(decimal.NewFromInt(100))).Round(2)
		l.Total = l.Subtotal.Add(l.TVAAmount)
		so.Subtotal = so.Subtotal.Add(l.Subtotal)
		so.TVAAmount = so.TVAAmount.Add(l.TVAAmount)
		l.SortOrder = i
	}
	so.TotalAmount = so.Subtotal.Add(so.TVAAmount).Add(so.StampTax).Sub(so.DiscountAmount)

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE sales_orders
		SET customer_id=$1, date=$2, delivery_date=$3,
		    subtotal=$4, discount_amount=$5, tva_amount=$6, stamp_tax=$7, total_amount=$8,
		    currency=$9, notes=$10, updated_at=NOW()
		WHERE id=$11 AND status IN ('draft','confirmed')`,
		so.CustomerID, so.Date, so.DeliveryDate,
		so.Subtotal, so.DiscountAmount, so.TVAAmount, so.StampTax, so.TotalAmount,
		so.Currency, so.Notes, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_, _ = tx.Exec(ctx, `DELETE FROM sales_order_lines WHERE order_id = $1`, id)
	for _, l := range so.Lines {
		_, err = tx.Exec(ctx, `
			INSERT INTO sales_order_lines
			  (id, order_id, item_id, description, quantity, unit_price,
			   discount_pct, tva_rate, subtotal, tva_amount, total,
			   account_id, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			l.ID, id, l.ItemID, l.Description,
			l.Quantity, l.UnitPrice, l.DiscountPct, l.TVARate,
			l.Subtotal, l.TVAAmount, l.Total,
			l.AccountID, l.SortOrder)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	so.ID = id
	c.JSON(http.StatusOK, so)
}

func (h *SalesHandler) ConfirmOrder(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	tag, err := h.db.Exec(ctx, `
		UPDATE sales_orders SET status='confirmed', updated_at=NOW()
		WHERE id=$1 AND status='draft'`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order not found or not in draft status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order confirmed"})
}

func (h *SalesHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE sales_orders SET status='cancelled', updated_at=NOW()
		WHERE id=$1 AND status NOT IN ('delivered','cancelled')`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled"})
}

// DeliverOrder is the renamed FulfillOrder endpoint (/orders/:id/deliver).
func (h *SalesHandler) DeliverOrder(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	tag, err := h.db.Exec(ctx, `
		UPDATE sales_orders SET status='delivered', delivery_date=CURRENT_DATE, updated_at=NOW()
		WHERE id=$1 AND status='confirmed'`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order not found or not confirmed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order delivered"})
}
