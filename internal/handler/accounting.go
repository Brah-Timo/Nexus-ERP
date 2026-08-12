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

// ─── Accounting Handler ────────────────────────────────────────────────────────

type AccountingHandler struct{ db *pgxpool.Pool }

// ── Chart of Accounts ──────────────────────────────────────────────────────────
// Schema: chart_of_accounts(id, company_id, code, name, type, nature, parent_id,
//   is_group, is_reconcilable, currency, balance, debit_balance, credit_balance,
//   level, description, is_active, created_at, updated_at, created_by)

func (h *AccountingHandler) ListChartOfAccounts(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT id, code, name, type, nature, parent_id,
		       is_group, is_reconcilable, COALESCE(currency,'DZD'),
		       COALESCE(balance,0), COALESCE(debit_balance,0), COALESCE(credit_balance,0),
		       COALESCE(level,0), COALESCE(description,''), COALESCE(is_active,true), created_at
		FROM chart_of_accounts
		WHERE company_id = $1
		ORDER BY code
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var accounts []models.ChartOfAccount
	for rows.Next() {
		var a models.ChartOfAccount
		err := rows.Scan(
			&a.ID, &a.Code, &a.Name, &a.Type, &a.Nature, &a.ParentID,
			&a.IsGroup, &a.IsReconcilable, &a.Currency,
			&a.Balance, &a.DebitBalance, &a.CreditBalance,
			&a.Level, &a.Description, &a.IsActive, &a.CreatedAt,
		)
		if err != nil {
			continue
		}
		accounts = append(accounts, a)
	}
	if accounts == nil {
		accounts = []models.ChartOfAccount{}
	}
	c.JSON(http.StatusOK, accounts)
}

func (h *AccountingHandler) CreateAccount(c *gin.Context) {
	var a models.ChartOfAccount
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a.ID = uuid.NewString()
	a.CompanyID = middleware.GetCompanyID(c)
	if a.Currency == "" {
		a.Currency = "DZD"
	}
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		INSERT INTO chart_of_accounts
		(id, company_id, code, name, type, nature, parent_id, is_group, is_reconcilable, currency, description, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`, a.ID, a.CompanyID, a.Code, a.Name, a.Type, a.Nature,
		a.ParentID, a.IsGroup, a.IsReconcilable, a.Currency, a.Description, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *AccountingHandler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	var a models.ChartOfAccount
	if err := c.ShouldBindJSON(&a); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx,
		`UPDATE chart_of_accounts SET name=$1, is_group=$2, description=$3 WHERE id=$4`,
		a.Name, a.IsGroup, a.Description, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

// ── Journal Entries ────────────────────────────────────────────────────────────
// Schema: journal_entries(id, company_id, branch_id, fiscal_year_id, number, date,
//   reference, description, status, total_debit, total_credit, currency,
//   source_type, source_id, cost_center_id, created_at, updated_at, created_by)
// Schema: journal_lines(id, journal_entry_id, account_id, cost_center_id,
//   description, debit, credit, currency, created_at)

func (h *AccountingHandler) ListJournalEntries(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	query := `
		SELECT id, number, date, COALESCE(reference,''), COALESCE(description,''),
		       status, COALESCE(total_debit,0), COALESCE(total_credit,0),
		       COALESCE(currency,'DZD'), COALESCE(source_type,''), created_at
		FROM journal_entries
		WHERE company_id = $1
	`
	params := []interface{}{companyID}

	if status := c.Query("status"); status != "" {
		query += ` AND status = $2`
		params = append(params, status)
	}
	query += ` ORDER BY date DESC, number DESC LIMIT 100`

	rows, err := h.db.Query(ctx, query, params...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var entries []models.JournalEntry
	for rows.Next() {
		var e models.JournalEntry
		err := rows.Scan(
			&e.ID, &e.Number, &e.Date, &e.Reference, &e.Description,
			&e.Status, &e.TotalDebit, &e.TotalCredit, &e.Currency, &e.SourceType, &e.CreatedAt,
		)
		if err != nil {
			continue
		}
		entries = append(entries, e)
	}
	if entries == nil {
		entries = []models.JournalEntry{}
	}
	c.JSON(http.StatusOK, entries)
}

func (h *AccountingHandler) GetJournalEntry(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var e models.JournalEntry
	err := h.db.QueryRow(ctx, `
		SELECT id, number, date, COALESCE(reference,''), COALESCE(description,''),
		       fiscal_year_id, status, COALESCE(total_debit,0), COALESCE(total_credit,0),
		       COALESCE(currency,'DZD'), COALESCE(source_type,''), source_id, created_at
		FROM journal_entries WHERE id = $1
	`, id).Scan(
		&e.ID, &e.Number, &e.Date, &e.Reference, &e.Description,
		&e.FiscalYearID, &e.Status, &e.TotalDebit, &e.TotalCredit,
		&e.Currency, &e.SourceType, &e.SourceID, &e.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Journal entry not found"})
		return
	}

	// Fetch lines using correct FK: journal_entry_id
	rows, _ := h.db.Query(ctx, `
		SELECT jl.id, jl.journal_entry_id, jl.account_id,
		       COALESCE(ca.code,''), COALESCE(ca.name,''),
		       COALESCE(jl.description,''),
		       COALESCE(jl.debit,0), COALESCE(jl.credit,0),
		       COALESCE(jl.currency,'DZD'), jl.cost_center_id
		FROM journal_lines jl
		LEFT JOIN chart_of_accounts ca ON ca.id = jl.account_id
		WHERE jl.journal_entry_id = $1
		ORDER BY jl.id
	`, id)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var l models.JournalLine
			_ = rows.Scan(
				&l.ID, &l.JournalEntryID, &l.AccountID,
				&l.AccountCode, &l.AccountName,
				&l.Description, &l.Debit, &l.Credit,
				&l.Currency, &l.CostCenterID,
			)
			e.Lines = append(e.Lines, l)
		}
	}

	c.JSON(http.StatusOK, e)
}

func (h *AccountingHandler) CreateJournalEntry(c *gin.Context) {
	var e models.JournalEntry
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate balance
	var totalDebit, totalCredit decimal.Decimal
	for _, l := range e.Lines {
		totalDebit = totalDebit.Add(l.Debit)
		totalCredit = totalCredit.Add(l.Credit)
	}
	if !totalDebit.Equal(totalCredit) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Journal entry is not balanced (debit != credit)"})
		return
	}

	e.ID = uuid.NewString()
	e.CompanyID = middleware.GetCompanyID(c)
	e.Status = models.JEStatusDraft
	e.TotalDebit = totalDebit
	e.TotalCredit = totalCredit
	e.Number = generateNumber("JE", e.CompanyID, h.db)
	if e.Currency == "" {
		e.Currency = "DZD"
	}

	ctx := context.Background()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO journal_entries
		(id, company_id, number, date, reference, description, fiscal_year_id,
		 status, total_debit, total_credit, currency, source_type, source_id, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, e.ID, e.CompanyID, e.Number, e.Date, e.Reference, e.Description, e.FiscalYearID,
		e.Status, e.TotalDebit, e.TotalCredit, e.Currency, e.SourceType, e.SourceID,
		middleware.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, l := range e.Lines {
		l.ID = uuid.NewString()
		if l.Currency == "" {
			l.Currency = e.Currency
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO journal_lines (id, journal_entry_id, account_id, cost_center_id, description, debit, credit, currency)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		`, l.ID, e.ID, l.AccountID, l.CostCenterID, l.Description, l.Debit, l.Credit, l.Currency)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit"})
		return
	}

	c.JSON(http.StatusCreated, e)
}

func (h *AccountingHandler) PostJournalEntry(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE journal_entries SET status = 'posted' WHERE id = $1 AND status = 'draft'`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update account balances using correct FK journal_entry_id
	_, err = tx.Exec(ctx, `
		UPDATE chart_of_accounts ca
		SET balance = balance + (
			SELECT COALESCE(SUM(jl.debit - jl.credit), 0)
			FROM journal_lines jl
			WHERE jl.journal_entry_id = $1 AND jl.account_id = ca.id
		)
		WHERE id IN (SELECT account_id FROM journal_lines WHERE journal_entry_id = $1)
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Journal entry posted"})
}

func (h *AccountingHandler) CancelJournalEntry(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, err := h.db.Exec(ctx,
		`UPDATE journal_entries SET status = 'cancelled' WHERE id = $1 AND status = 'draft'`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Journal entry cancelled"})
}

// ── Cost Centers ───────────────────────────────────────────────────────────────
// Schema: cost_centers(id, company_id, code, name, parent_id, budget, actual, is_active, ...)
// NOTE: NO type column in cost_centers table.

func (h *AccountingHandler) ListCostCenters(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, code, name, parent_id,
		        COALESCE(budget,0), COALESCE(actual,0), COALESCE(is_active,true), created_at
		 FROM cost_centers WHERE company_id = $1 ORDER BY code`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var ccs []models.CostCenter
	for rows.Next() {
		var cc models.CostCenter
		err := rows.Scan(&cc.ID, &cc.Code, &cc.Name, &cc.ParentID,
			&cc.Budget, &cc.Actual, &cc.IsActive, &cc.CreatedAt)
		if err != nil {
			continue
		}
		ccs = append(ccs, cc)
	}
	if ccs == nil {
		ccs = []models.CostCenter{}
	}
	c.JSON(http.StatusOK, ccs)
}

func (h *AccountingHandler) CreateCostCenter(c *gin.Context) {
	var cc models.CostCenter
	if err := c.ShouldBindJSON(&cc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cc.ID = uuid.NewString()
	cc.CompanyID = middleware.GetCompanyID(c)
	ctx := context.Background()
	// No type column — insert only existing columns
	_, err := h.db.Exec(ctx,
		`INSERT INTO cost_centers (id, company_id, code, name, parent_id, budget, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		cc.ID, cc.CompanyID, cc.Code, cc.Name, cc.ParentID, cc.Budget, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cc)
}

// ── Fixed Assets ───────────────────────────────────────────────────────────────
// Schema: fixed_assets(id, company_id, code, name, category,
//   purchase_date, purchase_value, residual_value, current_value,
//   accumulated_depreciation, depreciation_method, useful_life_years,
//   depreciation_rate, status, asset_account_id, dep_expense_account_id,
//   acc_dep_account_id, created_at, ...)

func (h *AccountingHandler) ListFixedAssets(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, COALESCE(code,''), name, COALESCE(category,''),
		        purchase_date, COALESCE(purchase_value,0),
		        COALESCE(residual_value,0), COALESCE(current_value,0),
		        COALESCE(accumulated_depreciation,0),
		        COALESCE(depreciation_method,'linear'),
		        COALESCE(useful_life_years,0), COALESCE(depreciation_rate,0),
		        COALESCE(status,'active'), created_at
		 FROM fixed_assets WHERE company_id = $1 ORDER BY name`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	var assets []models.FixedAsset
	for rows.Next() {
		var fa models.FixedAsset
		err := rows.Scan(
			&fa.ID, &fa.Code, &fa.Name, &fa.Category,
			&fa.PurchaseDate, &fa.PurchaseValue,
			&fa.ResidualValue, &fa.CurrentValue,
			&fa.AccumulatedDepreciation,
			&fa.DepreciationMethod, &fa.UsefulLifeYears,
			&fa.DepreciationRate, &fa.Status, &fa.CreatedAt,
		)
		if err != nil {
			continue
		}
		assets = append(assets, fa)
	}
	if assets == nil {
		assets = []models.FixedAsset{}
	}
	c.JSON(http.StatusOK, assets)
}

func (h *AccountingHandler) CreateFixedAsset(c *gin.Context) {
	var fa models.FixedAsset
	if err := c.ShouldBindJSON(&fa); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fa.ID = uuid.NewString()
	fa.CompanyID = middleware.GetCompanyID(c)
	fa.CurrentValue = fa.PurchaseValue // initial book value = purchase value
	if fa.DepreciationMethod == "" {
		fa.DepreciationMethod = "linear"
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO fixed_assets
		(id, company_id, code, name, category, purchase_date, purchase_value,
		 residual_value, current_value, depreciation_method, useful_life_years,
		 depreciation_rate, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, fa.ID, fa.CompanyID, fa.Code, fa.Name, fa.Category,
		fa.PurchaseDate, fa.PurchaseValue,
		fa.ResidualValue, fa.CurrentValue,
		fa.DepreciationMethod, fa.UsefulLifeYears,
		fa.DepreciationRate, "active")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, fa)
}

func (h *AccountingHandler) RunDepreciation(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// Use correct column names: purchase_value, current_value
	rows, err := h.db.Query(ctx, `
		SELECT id, COALESCE(purchase_value,0), COALESCE(residual_value,0),
		       COALESCE(current_value,0), COALESCE(useful_life_years,1),
		       COALESCE(depreciation_method,'linear')
		FROM fixed_assets
		WHERE company_id = $1 AND status = 'active' AND current_value > residual_value
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type DepResult struct {
		AssetID string          `json:"asset_id"`
		Amount  decimal.Decimal `json:"amount"`
	}

	var results []DepResult
	for rows.Next() {
		var assetID, method string
		var purchaseVal, residual, currentVal decimal.Decimal
		var lifeYears int
		_ = rows.Scan(&assetID, &purchaseVal, &residual, &currentVal, &lifeYears, &method)

		if lifeYears <= 0 {
			lifeYears = 1
		}

		var monthlyDep decimal.Decimal
		if method == "linear" {
			depBase := purchaseVal.Sub(residual)
			monthlyDep = depBase.Div(decimal.NewFromInt(int64(lifeYears * 12)))
		} else {
			// diminishing_balance
			rate := decimal.NewFromFloat(2.0).Div(decimal.NewFromInt(int64(lifeYears)))
			monthlyDep = currentVal.Mul(rate).Div(decimal.NewFromInt(12))
		}

		remaining := currentVal.Sub(residual)
		if monthlyDep.GreaterThan(remaining) {
			monthlyDep = remaining
		}

		if monthlyDep.IsZero() {
			continue
		}

		// Update current_value and accumulated_depreciation
		_, _ = h.db.Exec(ctx, `
			UPDATE fixed_assets
			SET current_value = current_value - $1,
			    accumulated_depreciation = accumulated_depreciation + $1
			WHERE id = $2
		`, monthlyDep, assetID)

		results = append(results, DepResult{AssetID: assetID, Amount: monthlyDep})
	}

	if results == nil {
		results = []DepResult{}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Depreciation run for %d assets", len(results)),
		"results": results,
	})
}

// ── Bank Reconciliation ────────────────────────────────────────────────────────
// Schema: bank_reconciliations(id, bank_account_id, period_date, statement_balance,
//   book_balance, difference GENERATED, is_reconciled, reconciled_at, notes, created_at)
// NOTE: NO company_id directly; joined via bank_accounts.

func (h *AccountingHandler) ListBankReconciliations(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT br.id, br.bank_account_id, COALESCE(ba.bank_name,'') || ' - ' || COALESCE(ba.account_number,''),
		        br.period_date,
		        COALESCE(br.statement_balance,0), COALESCE(br.book_balance,0),
		        COALESCE(br.difference,0),
		        COALESCE(br.is_reconciled,false),
		        br.reconciled_at,
		        COALESCE(br.notes,''),
		        br.created_at
		 FROM bank_reconciliations br
		 JOIN bank_accounts ba ON ba.id = br.bank_account_id
		 WHERE ba.company_id = $1
		 ORDER BY br.period_date DESC`, companyID)
	if err != nil {
		// Return empty array on error (table might be empty)
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var recs []models.BankReconciliation
	for rows.Next() {
		var r models.BankReconciliation
		err := rows.Scan(
			&r.ID, &r.BankAccountID, &r.BankAccountName,
			&r.PeriodDate, &r.StatementBalance, &r.BookBalance,
			&r.Difference, &r.IsReconciled, &r.ReconciledAt,
			&r.Notes, &r.CreatedAt,
		)
		if err != nil {
			continue
		}
		recs = append(recs, r)
	}
	if recs == nil {
		recs = []models.BankReconciliation{}
	}
	c.JSON(http.StatusOK, recs)
}

func (h *AccountingHandler) CreateBankReconciliation(c *gin.Context) {
	var req struct {
		BankAccountID    string          `json:"bank_account_id"`
		PeriodDate       string          `json:"period_date"`
		StatementBalance decimal.Decimal `json:"statement_balance"`
		BookBalance      decimal.Decimal `json:"book_balance"`
		Notes            string          `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.NewString()
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		INSERT INTO bank_reconciliations
		(id, bank_account_id, period_date, statement_balance, book_balance, is_reconciled, notes)
		VALUES ($1,$2,$3::date,$4,$5,$6,$7)
	`, id, req.BankAccountID, req.PeriodDate,
		req.StatementBalance, req.BookBalance, false, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Bank reconciliation created"})
}

// ── Financial Reports ──────────────────────────────────────────────────────────
// chart_of_accounts.type column IS correct (asset/liability/equity/revenue/expense/contra)
// journal_lines FK is journal_entry_id

func (h *AccountingHandler) TrialBalance(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT ca.code, ca.name, ca.type,
		       COALESCE(SUM(CASE WHEN jl.debit > 0 THEN jl.debit ELSE 0 END),0) AS total_debit,
		       COALESCE(SUM(CASE WHEN jl.credit > 0 THEN jl.credit ELSE 0 END),0) AS total_credit,
		       COALESCE(ca.balance,0)
		FROM chart_of_accounts ca
		LEFT JOIN journal_lines jl ON jl.account_id = ca.id
		LEFT JOIN journal_entries je ON je.id = jl.journal_entry_id AND je.status = 'posted'
		WHERE ca.company_id = $1
		GROUP BY ca.code, ca.name, ca.type, ca.balance
		ORDER BY ca.code
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var lines []map[string]interface{}
	for rows.Next() {
		var code, name, accType string
		var debit, credit, balance float64
		_ = rows.Scan(&code, &name, &accType, &debit, &credit, &balance)
		lines = append(lines, map[string]interface{}{
			"code": code, "name": name, "type": accType,
			"total_debit": debit, "total_credit": credit, "balance": balance,
		})
	}
	if lines == nil {
		lines = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"report": "trial_balance",
		"date":   time.Now().Format("2006-01-02"),
		"lines":  lines,
	})
}

func (h *AccountingHandler) BalanceSheet(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	type Section struct {
		Name     string                   `json:"name"`
		Accounts []map[string]interface{} `json:"accounts"`
		Total    float64                  `json:"total"`
	}

	sections := map[string]*Section{
		"asset":     {Name: "Assets"},
		"liability": {Name: "Liabilities"},
		"equity":    {Name: "Equity"},
	}

	rows, err := h.db.Query(ctx, `
		SELECT type, code, name, COALESCE(balance,0)
		FROM chart_of_accounts
		WHERE company_id = $1 AND type IN ('asset','liability','equity') AND is_active = true
		ORDER BY code
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var accType, code, name string
		var balance float64
		_ = rows.Scan(&accType, &code, &name, &balance)
		if s, ok := sections[accType]; ok {
			s.Accounts = append(s.Accounts, map[string]interface{}{
				"code": code, "name": name, "balance": balance,
			})
			s.Total += balance
		}
	}

	for _, s := range sections {
		if s.Accounts == nil {
			s.Accounts = []map[string]interface{}{}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"report":      "balance_sheet",
		"date":        time.Now().Format("2006-01-02"),
		"assets":      sections["asset"],
		"liabilities": sections["liability"],
		"equity":      sections["equity"],
	})
}

func (h *AccountingHandler) IncomeStatement(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT type, code, name, COALESCE(balance,0)
		FROM chart_of_accounts
		WHERE company_id = $1 AND type IN ('revenue','expense') AND is_active = true
		ORDER BY code
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var revenues, expenses []map[string]interface{}
	var totalRevenue, totalExpense float64

	for rows.Next() {
		var accType, code, name string
		var balance float64
		_ = rows.Scan(&accType, &code, &name, &balance)
		item := map[string]interface{}{"code": code, "name": name, "balance": balance}
		if accType == "revenue" {
			revenues = append(revenues, item)
			totalRevenue += balance
		} else {
			expenses = append(expenses, item)
			totalExpense += balance
		}
	}

	if revenues == nil {
		revenues = []map[string]interface{}{}
	}
	if expenses == nil {
		expenses = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"report":        "income_statement",
		"period":        time.Now().Format("2006"),
		"revenues":      revenues,
		"expenses":      expenses,
		"total_revenue": totalRevenue,
		"total_expense": totalExpense,
		"net_profit":    totalRevenue - totalExpense,
	})
}

func (h *AccountingHandler) CashFlowStatement(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"report":  "cash_flow",
		"message": "Cash flow statement",
		"operating": []interface{}{},
		"investing": []interface{}{},
		"financing": []interface{}{},
	})
}

// ── Budgets ────────────────────────────────────────────────────────────────────
// Schema: budgets(id, company_id, fiscal_year_id, name, description, status,
//   total_budget, total_actual, created_at, ...)
// Schema: budget_lines(id, budget_id, account_id, cost_center_id,
//   jan,feb,mar,apr,may,jun,jul,aug,sep,oct,nov,dec, total_budget GENERATED, total_actual)

func (h *AccountingHandler) ListBudgets(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, fiscal_year_id, COALESCE(name,''), COALESCE(description,''),
		        COALESCE(status,'draft'), COALESCE(total_budget,0), COALESCE(total_actual,0),
		        created_at
		 FROM budgets WHERE company_id = $1 ORDER BY created_at DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var budgets []models.Budget
	for rows.Next() {
		var b models.Budget
		err := rows.Scan(
			&b.ID, &b.FiscalYearID, &b.Name, &b.Description,
			&b.Status, &b.TotalBudget, &b.TotalActual, &b.CreatedAt,
		)
		if err != nil {
			continue
		}
		budgets = append(budgets, b)
	}
	if budgets == nil {
		budgets = []models.Budget{}
	}
	c.JSON(http.StatusOK, budgets)
}

func (h *AccountingHandler) GetBudget(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()

	var b models.Budget
	err := h.db.QueryRow(ctx,
		`SELECT id, fiscal_year_id, COALESCE(name,''), COALESCE(description,''),
		        COALESCE(status,'draft'), COALESCE(total_budget,0), COALESCE(total_actual,0), created_at
		 FROM budgets WHERE id = $1`, id).Scan(
		&b.ID, &b.FiscalYearID, &b.Name, &b.Description,
		&b.Status, &b.TotalBudget, &b.TotalActual, &b.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Budget not found"})
		return
	}

	// Fetch lines
	lrows, _ := h.db.Query(ctx, `
		SELECT bl.id, bl.budget_id, bl.account_id,
		       COALESCE(ca.code,''), COALESCE(ca.name,''),
		       bl.cost_center_id,
		       COALESCE(bl.jan,0),COALESCE(bl.feb,0),COALESCE(bl.mar,0),COALESCE(bl.apr,0),
		       COALESCE(bl.may,0),COALESCE(bl.jun,0),COALESCE(bl.jul,0),COALESCE(bl.aug,0),
		       COALESCE(bl.sep,0),COALESCE(bl.oct,0),COALESCE(bl.nov,0),COALESCE(bl.dec,0),
		       COALESCE(bl.total_budget,0), COALESCE(bl.total_actual,0)
		FROM budget_lines bl
		LEFT JOIN chart_of_accounts ca ON ca.id = bl.account_id
		WHERE bl.budget_id = $1
		ORDER BY ca.code
	`, id)
	if lrows != nil {
		defer lrows.Close()
		for lrows.Next() {
			var l models.BudgetLine
			_ = lrows.Scan(
				&l.ID, &l.BudgetID, &l.AccountID,
				&l.AccountCode, &l.AccountName, &l.CostCenterID,
				&l.Jan, &l.Feb, &l.Mar, &l.Apr,
				&l.May, &l.Jun, &l.Jul, &l.Aug,
				&l.Sep, &l.Oct, &l.Nov, &l.Dec,
				&l.TotalBudget, &l.TotalActual,
			)
			b.Lines = append(b.Lines, l)
		}
	}
	if b.Lines == nil {
		b.Lines = []models.BudgetLine{}
	}

	c.JSON(http.StatusOK, b)
}

func (h *AccountingHandler) CreateBudget(c *gin.Context) {
	var b models.Budget
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	b.ID = uuid.NewString()
	b.CompanyID = middleware.GetCompanyID(c)
	if b.Status == "" {
		b.Status = "draft"
	}
	ctx := context.Background()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO budgets (id, company_id, fiscal_year_id, name, description, status)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		b.ID, b.CompanyID, b.FiscalYearID, b.Name, b.Description, b.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert budget lines with per-month columns
	for _, l := range b.Lines {
		l.ID = uuid.NewString()
		_, err = tx.Exec(ctx, `
			INSERT INTO budget_lines
			(id, budget_id, account_id, cost_center_id,
			 jan,feb,mar,apr,may,jun,jul,aug,sep,oct,nov,dec)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		`, l.ID, b.ID, l.AccountID, l.CostCenterID,
			l.Jan, l.Feb, l.Mar, l.Apr,
			l.May, l.Jun, l.Jul, l.Aug,
			l.Sep, l.Oct, l.Nov, l.Dec)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, b)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func generateNumber(prefix string, companyID string, db *pgxpool.Pool) string {
	ctx := context.Background()
	var next int
	_ = db.QueryRow(ctx, `
		UPDATE numbering_config SET next_number = next_number + 1
		WHERE company_id = $1 AND document_type = $2
		RETURNING next_number
	`, companyID, prefix).Scan(&next)
	if next == 0 {
		next = 1
	}
	return fmt.Sprintf("%s-%05d", prefix, next)
}
