package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"nexus-erp/internal/middleware"
)

// ─── Treasury Handler ─────────────────────────────────────────────────────────

type TreasuryHandler struct{ db *pgxpool.Pool }

// ─────────────────────────────────────────────────────────────────────────────
// Cash Accounts
// ─────────────────────────────────────────────────────────────────────────────

func (h *TreasuryHandler) ListCashAccounts(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			id, name, COALESCE(account_type, 'petty_cash'),
			COALESCE(account_number, ''), currency,
			balance, COALESCE(opening_balance, 0),
			is_active, COALESCE(notes, ''),
			created_at, COALESCE(updated_at, created_at)
		FROM cash_accounts
		WHERE company_id = $1
		ORDER BY name
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type CashAccountRow struct {
		ID             string    `json:"id"`
		Name           string    `json:"name"`
		AccountType    string    `json:"account_type"`
		AccountNumber  string    `json:"account_number"`
		Currency       string    `json:"currency"`
		Balance        float64   `json:"balance"`
		OpeningBalance float64   `json:"opening_balance"`
		IsActive       bool      `json:"is_active"`
		Notes          string    `json:"notes"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}

	var accounts []CashAccountRow
	for rows.Next() {
		var a CashAccountRow
		if err := rows.Scan(
			&a.ID, &a.Name, &a.AccountType, &a.AccountNumber,
			&a.Currency, &a.Balance, &a.OpeningBalance,
			&a.IsActive, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
		); err == nil {
			accounts = append(accounts, a)
		}
	}
	if accounts == nil {
		accounts = []CashAccountRow{}
	}
	c.JSON(http.StatusOK, accounts)
}

func (h *TreasuryHandler) GetCashAccount(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var (
		accID, name, accType, accNum, currency, notes string
		balance, opening                              float64
		isActive                                      bool
		createdAt, updatedAt                          time.Time
	)
	err := h.db.QueryRow(ctx, `
		SELECT id, name, COALESCE(account_type,'petty_cash'),
		       COALESCE(account_number,''), currency, balance,
		       COALESCE(opening_balance,0), is_active, COALESCE(notes,''),
		       created_at, COALESCE(updated_at, created_at)
		FROM cash_accounts WHERE id = $1 AND company_id = $2
	`, id, companyID).Scan(
		&accID, &name, &accType, &accNum, &currency,
		&balance, &opening, &isActive, &notes, &createdAt, &updatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cash account not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": accID, "name": name, "account_type": accType,
		"account_number": accNum, "currency": currency,
		"balance": balance, "opening_balance": opening,
		"is_active": isActive, "notes": notes,
		"created_at": createdAt, "updated_at": updatedAt,
	})
}

func (h *TreasuryHandler) CreateCashAccount(c *gin.Context) {
	var req struct {
		Name          string  `json:"name" binding:"required"`
		AccountType   string  `json:"account_type"`
		AccountNumber string  `json:"account_number"`
		Currency      string  `json:"currency"`
		OpeningBalance float64 `json:"opening_balance"`
		Notes         string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Currency == "" {
		req.Currency = "DZD"
	}
	if req.AccountType == "" {
		req.AccountType = "petty_cash"
	}

	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		INSERT INTO cash_accounts
		  (id, company_id, name, account_type, account_number,
		   currency, balance, opening_balance, is_active, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$7,$8,$9)
	`, id, companyID, req.Name, req.AccountType, req.AccountNumber,
		req.Currency, req.OpeningBalance, true, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "name": req.Name, "message": "Cash account created"})
}

func (h *TreasuryHandler) UpdateCashAccount(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	var req struct {
		Name          string  `json:"name"`
		AccountType   string  `json:"account_type"`
		AccountNumber string  `json:"account_number"`
		Currency      string  `json:"currency"`
		Notes         string  `json:"notes"`
		IsActive      *bool   `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE cash_accounts
		SET name = COALESCE(NULLIF($1,''), name),
		    account_type = COALESCE(NULLIF($2,''), account_type),
		    account_number = COALESCE(NULLIF($3,''), account_number),
		    currency = COALESCE(NULLIF($4,''), currency),
		    notes = $5,
		    is_active = COALESCE($6, is_active),
		    updated_at = NOW()
		WHERE id = $7 AND company_id = $8
	`, req.Name, req.AccountType, req.AccountNumber, req.Currency,
		req.Notes, req.IsActive, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cash account updated"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Bank Accounts
// ─────────────────────────────────────────────────────────────────────────────

func (h *TreasuryHandler) ListBankAccounts(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			id, bank_name, account_number, COALESCE(rib,''),
			currency, balance, COALESCE(opening_balance,0),
			COALESCE(swift_code,''), COALESCE(branch,''),
			is_active, COALESCE(notes,''),
			COALESCE(last_reconciled::TEXT,''),
			created_at, COALESCE(updated_at, created_at)
		FROM bank_accounts
		WHERE company_id = $1
		ORDER BY bank_name
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type BankAccountRow struct {
		ID             string    `json:"id"`
		BankName       string    `json:"bank_name"`
		AccountNumber  string    `json:"account_number"`
		RIB            string    `json:"rib"`
		Currency       string    `json:"currency"`
		Balance        float64   `json:"balance"`
		OpeningBalance float64   `json:"opening_balance"`
		SwiftCode      string    `json:"swift_code"`
		Branch         string    `json:"branch"`
		IsActive       bool      `json:"is_active"`
		Notes          string    `json:"notes"`
		LastReconciled string    `json:"last_reconciled"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
	}

	var accounts []BankAccountRow
	for rows.Next() {
		var a BankAccountRow
		if err := rows.Scan(
			&a.ID, &a.BankName, &a.AccountNumber, &a.RIB,
			&a.Currency, &a.Balance, &a.OpeningBalance,
			&a.SwiftCode, &a.Branch, &a.IsActive, &a.Notes,
			&a.LastReconciled, &a.CreatedAt, &a.UpdatedAt,
		); err == nil {
			accounts = append(accounts, a)
		}
	}
	if accounts == nil {
		accounts = []BankAccountRow{}
	}
	c.JSON(http.StatusOK, accounts)
}

func (h *TreasuryHandler) CreateBankAccount(c *gin.Context) {
	var req struct {
		BankName      string  `json:"bank_name" binding:"required"`
		AccountNumber string  `json:"account_number" binding:"required"`
		RIB           string  `json:"rib"`
		Currency      string  `json:"currency"`
		SwiftCode     string  `json:"swift_code"`
		Branch        string  `json:"branch"`
		OpeningBalance float64 `json:"opening_balance"`
		Notes         string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Currency == "" {
		req.Currency = "DZD"
	}

	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		INSERT INTO bank_accounts
		  (id, company_id, bank_name, account_number, rib, currency,
		   balance, opening_balance, swift_code, branch, is_active, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$7,$8,$9,$10,$11)
	`, id, companyID, req.BankName, req.AccountNumber, req.RIB,
		req.Currency, req.OpeningBalance, req.SwiftCode, req.Branch, true, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "bank_name": req.BankName, "message": "Bank account created"})
}

func (h *TreasuryHandler) UpdateBankAccount(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	var req struct {
		BankName      string `json:"bank_name"`
		AccountNumber string `json:"account_number"`
		RIB           string `json:"rib"`
		Currency      string `json:"currency"`
		SwiftCode     string `json:"swift_code"`
		Branch        string `json:"branch"`
		Notes         string `json:"notes"`
		IsActive      *bool  `json:"is_active"`
	}
	c.ShouldBindJSON(&req)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE bank_accounts
		SET bank_name      = COALESCE(NULLIF($1,''), bank_name),
		    account_number = COALESCE(NULLIF($2,''), account_number),
		    rib            = COALESCE(NULLIF($3,''), rib),
		    currency       = COALESCE(NULLIF($4,''), currency),
		    swift_code     = COALESCE(NULLIF($5,''), swift_code),
		    branch         = COALESCE(NULLIF($6,''), branch),
		    notes          = $7,
		    is_active      = COALESCE($8, is_active),
		    updated_at     = NOW()
		WHERE id = $9 AND company_id = $10
	`, req.BankName, req.AccountNumber, req.RIB, req.Currency,
		req.SwiftCode, req.Branch, req.Notes, req.IsActive, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bank account updated"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Cheques
// ─────────────────────────────────────────────────────────────────────────────

func (h *TreasuryHandler) ListCheques(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	status := c.Query("status")
	chequeType := c.Query("type")

	query := `
		SELECT
			c.id, c.number, c.type, c.partner_name,
			c.amount, c.issue_date, c.due_date, c.status,
			COALESCE(c.bank_account_id::TEXT, ''),
			COALESCE(ba.bank_name, ''),
			COALESCE(c.notes, ''),
			COALESCE(c.deposited_date::TEXT, ''),
			COALESCE(c.bounced_date::TEXT, ''),
			c.created_at
		FROM cheques c
		LEFT JOIN bank_accounts ba ON ba.id = c.bank_account_id
		WHERE c.company_id = $1
	`
	args := []interface{}{companyID}
	argIdx := 2

	if status != "" {
		query += " AND c.status = $" + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if chequeType != "" {
		query += " AND c.type = $" + itoa(argIdx)
		args = append(args, chequeType)
		argIdx++
	}
	query += " ORDER BY c.due_date DESC"

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type ChequeRow struct {
		ID             string    `json:"id"`
		Number         string    `json:"number"`
		Type           string    `json:"type"`
		PartnerName    string    `json:"partner_name"`
		Amount         float64   `json:"amount"`
		IssueDate      time.Time `json:"issue_date"`
		DueDate        time.Time `json:"due_date"`
		Status         string    `json:"status"`
		BankAccountID  string    `json:"bank_account_id"`
		BankName       string    `json:"bank_name"`
		Notes          string    `json:"notes"`
		DepositedDate  string    `json:"deposited_date"`
		BouncedDate    string    `json:"bounced_date"`
		CreatedAt      time.Time `json:"created_at"`
	}

	var cheques []ChequeRow
	for rows.Next() {
		var ch ChequeRow
		if err := rows.Scan(
			&ch.ID, &ch.Number, &ch.Type, &ch.PartnerName,
			&ch.Amount, &ch.IssueDate, &ch.DueDate, &ch.Status,
			&ch.BankAccountID, &ch.BankName, &ch.Notes,
			&ch.DepositedDate, &ch.BouncedDate, &ch.CreatedAt,
		); err == nil {
			cheques = append(cheques, ch)
		}
	}
	if cheques == nil {
		cheques = []ChequeRow{}
	}
	c.JSON(http.StatusOK, cheques)
}

func (h *TreasuryHandler) CreateCheque(c *gin.Context) {
	var req struct {
		Number        string    `json:"number" binding:"required"`
		Type          string    `json:"type" binding:"required"`
		BankAccountID string    `json:"bank_account_id"`
		PartnerID     string    `json:"partner_id"`
		PartnerName   string    `json:"partner_name"`
		Amount        float64   `json:"amount" binding:"required"`
		IssueDate     time.Time `json:"issue_date"`
		DueDate       time.Time `json:"due_date"`
		Notes         string    `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		INSERT INTO cheques
		  (id, company_id, number, type, bank_account_id, partner_id,
		   partner_name, amount, issue_date, due_date, status, notes)
		VALUES ($1,$2,$3,$4,
		  CASE WHEN $5='' THEN NULL ELSE $5::UUID END,
		  CASE WHEN $6='' THEN NULL ELSE $6::UUID END,
		  $7,$8,$9,$10,'pending',$11)
	`, id, companyID, req.Number, req.Type,
		req.BankAccountID, req.PartnerID, req.PartnerName,
		req.Amount, req.IssueDate, req.DueDate, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "number": req.Number, "message": "Cheque created"})
}

func (h *TreasuryHandler) UpdateCheque(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	var req struct {
		PartnerName   string    `json:"partner_name"`
		Amount        float64   `json:"amount"`
		DueDate       time.Time `json:"due_date"`
		Notes         string    `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE cheques
		SET partner_name = COALESCE(NULLIF($1,''), partner_name),
		    amount       = CASE WHEN $2 > 0 THEN $2 ELSE amount END,
		    due_date     = CASE WHEN $3 > '0001-01-01'::DATE THEN $3 ELSE due_date END,
		    notes        = $4,
		    updated_at   = NOW()
		WHERE id = $5 AND company_id = $6
	`, req.PartnerName, req.Amount, req.DueDate, req.Notes, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cheque updated"})
}

func (h *TreasuryHandler) DepositCheque(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		UPDATE cheques
		SET status = 'deposited',
		    deposited_date = CURRENT_DATE,
		    updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND status = 'pending'
	`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cheque deposited"})
}

func (h *TreasuryHandler) BounceCheque(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		UPDATE cheques
		SET status = 'bounced',
		    bounced_date = CURRENT_DATE,
		    updated_at = NOW()
		WHERE id = $1 AND company_id = $2
	`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cheque marked as bounced"})
}

func (h *TreasuryHandler) CancelCheque(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE cheques SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND status = 'pending'
	`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cheque cancelled"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Treasury Movements
// ─────────────────────────────────────────────────────────────────────────────

func (h *TreasuryHandler) ListMovements(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	movType := c.Query("type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := `
		SELECT
			tm.id,
			tm.type,
			tm.amount,
			tm.date,
			COALESCE(tm.reference, ''),
			COALESCE(tm.notes, ''),
			COALESCE(tm.category, ''),
			COALESCE(tm.reconciled, false),
			COALESCE(ca.name, ''),
			COALESCE(ba.bank_name, ''),
			tm.created_at
		FROM treasury_movements tm
		LEFT JOIN cash_accounts ca ON ca.id = tm.cash_account_id
		LEFT JOIN bank_accounts  ba ON ba.id = tm.bank_account_id
		WHERE tm.company_id = $1
	`
	args := []interface{}{companyID}
	argIdx := 2

	if movType != "" {
		query += " AND tm.type = $" + itoa(argIdx)
		args = append(args, movType)
		argIdx++
	}
	if startDate != "" {
		query += " AND tm.date >= $" + itoa(argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		query += " AND tm.date <= $" + itoa(argIdx)
		args = append(args, endDate)
		argIdx++
	}
	query += " ORDER BY tm.date DESC LIMIT 200"

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type MovRow struct {
		ID              string    `json:"id"`
		Type            string    `json:"type"`
		Amount          float64   `json:"amount"`
		Date            time.Time `json:"date"`
		Reference       string    `json:"reference"`
		Notes           string    `json:"notes"`
		Category        string    `json:"category"`
		Reconciled      bool      `json:"reconciled"`
		CashAccountName string    `json:"cash_account_name"`
		BankAccountName string    `json:"bank_account_name"`
		CreatedAt       time.Time `json:"created_at"`
	}

	var movements []MovRow
	for rows.Next() {
		var m MovRow
		if err := rows.Scan(
			&m.ID, &m.Type, &m.Amount, &m.Date,
			&m.Reference, &m.Notes, &m.Category, &m.Reconciled,
			&m.CashAccountName, &m.BankAccountName, &m.CreatedAt,
		); err == nil {
			movements = append(movements, m)
		}
	}
	if movements == nil {
		movements = []MovRow{}
	}
	c.JSON(http.StatusOK, movements)
}

func (h *TreasuryHandler) CreateMovement(c *gin.Context) {
	var req struct {
		Type          string    `json:"type" binding:"required"`
		Amount        float64   `json:"amount" binding:"required"`
		Date          time.Time `json:"date"`
		Reference     string    `json:"reference"`
		Notes         string    `json:"notes"`
		Category      string    `json:"category"`
		CashAccountID string    `json:"cash_account_id"`
		BankAccountID string    `json:"bank_account_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Date.IsZero() {
		req.Date = time.Now()
	}

	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		INSERT INTO treasury_movements
		  (id, company_id, type, amount, date, reference, notes, category,
		   cash_account_id, bank_account_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,
		  CASE WHEN $9='' THEN NULL ELSE $9::UUID END,
		  CASE WHEN $10='' THEN NULL ELSE $10::UUID END)
	`, id, companyID, req.Type, req.Amount, req.Date,
		req.Reference, req.Notes, req.Category,
		req.CashAccountID, req.BankAccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Movement created"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Payments
// ─────────────────────────────────────────────────────────────────────────────

func (h *TreasuryHandler) ListPayments(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	payType := c.Query("type")
	status := c.Query("status")

	query := `
		SELECT
			p.id,
			p.type,
			COALESCE(p.partner_name, ''),
			p.date,
			p.amount,
			COALESCE(p.allocated_amount, 0),
			p.amount - COALESCE(p.allocated_amount, 0) AS unallocated,
			COALESCE(p.method, ''),
			COALESCE(p.reference, ''),
			COALESCE(p.status, 'draft'),
			COALESCE(ba.bank_name, ''),
			COALESCE(p.notes, ''),
			p.created_at
		FROM payments p
		LEFT JOIN bank_accounts ba ON ba.id = p.bank_account_id
		WHERE p.company_id = $1
	`
	args := []interface{}{companyID}
	argIdx := 2

	if payType != "" {
		query += " AND p.type = $" + itoa(argIdx)
		args = append(args, payType)
		argIdx++
	}
	if status != "" {
		query += " AND p.status = $" + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	query += " ORDER BY p.date DESC LIMIT 200"

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type PayRow struct {
		ID              string    `json:"id"`
		Type            string    `json:"type"`
		PartnerName     string    `json:"partner_name"`
		Date            time.Time `json:"date"`
		Amount          float64   `json:"amount"`
		AllocatedAmt    float64   `json:"allocated_amount"`
		UnallocatedAmt  float64   `json:"unallocated_amount"`
		Method          string    `json:"method"`
		Reference       string    `json:"reference"`
		Status          string    `json:"status"`
		BankName        string    `json:"bank_name"`
		Notes           string    `json:"notes"`
		CreatedAt       time.Time `json:"created_at"`
	}

	var payments []PayRow
	for rows.Next() {
		var p PayRow
		if err := rows.Scan(
			&p.ID, &p.Type, &p.PartnerName, &p.Date,
			&p.Amount, &p.AllocatedAmt, &p.UnallocatedAmt,
			&p.Method, &p.Reference, &p.Status,
			&p.BankName, &p.Notes, &p.CreatedAt,
		); err == nil {
			payments = append(payments, p)
		}
	}
	if payments == nil {
		payments = []PayRow{}
	}
	c.JSON(http.StatusOK, payments)
}

func (h *TreasuryHandler) CreatePayment(c *gin.Context) {
	var req struct {
		Type          string    `json:"type" binding:"required"`
		PartnerID     string    `json:"partner_id"`
		PartnerName   string    `json:"partner_name"`
		Date          time.Time `json:"date"`
		Amount        float64   `json:"amount" binding:"required"`
		Method        string    `json:"method"`
		Reference     string    `json:"reference"`
		BankAccountID string    `json:"bank_account_id"`
		CashAccountID string    `json:"cash_account_id"`
		Notes         string    `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Date.IsZero() {
		req.Date = time.Now()
	}

	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		INSERT INTO payments
		  (id, company_id, type, partner_id, partner_name, date, amount,
		   method, reference, bank_account_id, cash_account_id, notes, status,
		   allocated_amount)
		VALUES ($1,$2,$3,
		  CASE WHEN $4='' THEN NULL ELSE $4::UUID END,
		  $5,$6,$7,$8,$9,
		  CASE WHEN $10='' THEN NULL ELSE $10::UUID END,
		  CASE WHEN $11='' THEN NULL ELSE $11::UUID END,
		  $12,'draft',0)
	`, id, companyID, req.Type,
		req.PartnerID, req.PartnerName, req.Date, req.Amount,
		req.Method, req.Reference, req.BankAccountID, req.CashAccountID, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Payment created"})
}

func (h *TreasuryHandler) UpdatePayment(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	var req struct {
		PartnerName   string    `json:"partner_name"`
		Date          time.Time `json:"date"`
		Amount        float64   `json:"amount"`
		Method        string    `json:"method"`
		Reference     string    `json:"reference"`
		Notes         string    `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE payments
		SET partner_name = COALESCE(NULLIF($1,''), partner_name),
		    date         = CASE WHEN $2 > '0001-01-01'::DATE THEN $2 ELSE date END,
		    amount       = CASE WHEN $3 > 0 THEN $3 ELSE amount END,
		    method       = COALESCE(NULLIF($4,''), method),
		    reference    = COALESCE(NULLIF($5,''), reference),
		    notes        = $6,
		    updated_at   = NOW()
		WHERE id = $7 AND company_id = $8 AND status = 'draft'
	`, req.PartnerName, req.Date, req.Amount, req.Method, req.Reference, req.Notes, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Payment updated"})
}

func (h *TreasuryHandler) ConfirmPayment(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE payments SET status = 'confirmed', updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND status = 'draft'
	`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Payment confirmed"})
}

func (h *TreasuryHandler) AllocatePayment(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	var req struct {
		AllocatedAmount float64 `json:"allocated_amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE payments
		SET allocated_amount = $1, updated_at = NOW()
		WHERE id = $2 AND company_id = $3
	`, req.AllocatedAmount, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Payment allocated"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Receipts
// ─────────────────────────────────────────────────────────────────────────────

func (h *TreasuryHandler) ListReceipts(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	status := c.Query("status")
	receiptType := c.Query("type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := `
		SELECT
			r.id,
			r.number,
			r.receipt_type,
			r.status,
			COALESCE(r.partner_name, ''),
			r.amount,
			COALESCE(r.allocated_amount, 0),
			r.unallocated_amt,
			r.receipt_date,
			COALESCE(r.reference, ''),
			COALESCE(r.payment_method, 'bank_transfer'),
			COALESCE(r.currency, 'DZD'),
			COALESCE(ba.bank_name, ''),
			COALESCE(r.description, ''),
			r.created_at
		FROM receipts r
		LEFT JOIN bank_accounts ba ON ba.id = r.bank_account_id
		WHERE r.company_id = $1
	`
	args := []interface{}{companyID}
	argIdx := 2

	if status != "" {
		query += " AND r.status = $" + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if receiptType != "" {
		query += " AND r.receipt_type = $" + itoa(argIdx)
		args = append(args, receiptType)
		argIdx++
	}
	if startDate != "" {
		query += " AND r.receipt_date >= $" + itoa(argIdx)
		args = append(args, startDate)
		argIdx++
	}
	if endDate != "" {
		query += " AND r.receipt_date <= $" + itoa(argIdx)
		args = append(args, endDate)
		argIdx++
	}
	query += " ORDER BY r.receipt_date DESC LIMIT 200"

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type ReceiptRow struct {
		ID             string    `json:"id"`
		Number         string    `json:"number"`
		ReceiptType    string    `json:"receipt_type"`
		Status         string    `json:"status"`
		PartnerName    string    `json:"partner_name"`
		Amount         float64   `json:"amount"`
		AllocatedAmt   float64   `json:"allocated_amount"`
		UnallocatedAmt float64   `json:"unallocated_amount"`
		ReceiptDate    time.Time `json:"receipt_date"`
		Reference      string    `json:"reference"`
		PaymentMethod  string    `json:"payment_method"`
		Currency       string    `json:"currency"`
		BankName       string    `json:"bank_name"`
		Description    string    `json:"description"`
		CreatedAt      time.Time `json:"created_at"`
	}

	var receipts []ReceiptRow
	for rows.Next() {
		var r ReceiptRow
		if err := rows.Scan(
			&r.ID, &r.Number, &r.ReceiptType, &r.Status,
			&r.PartnerName, &r.Amount, &r.AllocatedAmt, &r.UnallocatedAmt,
			&r.ReceiptDate, &r.Reference, &r.PaymentMethod,
			&r.Currency, &r.BankName, &r.Description, &r.CreatedAt,
		); err == nil {
			receipts = append(receipts, r)
		}
	}
	if receipts == nil {
		receipts = []ReceiptRow{}
	}
	c.JSON(http.StatusOK, receipts)
}

func (h *TreasuryHandler) CreateReceipt(c *gin.Context) {
	var req struct {
		ReceiptType   string    `json:"receipt_type"`
		PartnerID     string    `json:"partner_id"`
		PartnerName   string    `json:"partner_name"`
		Amount        float64   `json:"amount" binding:"required"`
		ReceiptDate   time.Time `json:"receipt_date"`
		Reference     string    `json:"reference"`
		PaymentMethod string    `json:"payment_method"`
		BankAccountID string    `json:"bank_account_id"`
		CashAccountID string    `json:"cash_account_id"`
		Currency      string    `json:"currency"`
		Description   string    `json:"description"`
		InvoiceID     string    `json:"invoice_id"`
		InvoiceNumber string    `json:"invoice_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ReceiptDate.IsZero() {
		req.ReceiptDate = time.Now()
	}
	if req.ReceiptType == "" {
		req.ReceiptType = "customer"
	}
	if req.Currency == "" {
		req.Currency = "DZD"
	}
	if req.PaymentMethod == "" {
		req.PaymentMethod = "bank_transfer"
	}

	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// Generate receipt number
	var number string
	h.db.QueryRow(ctx, `SELECT next_receipt_number($1)`, companyID).Scan(&number)
	if number == "" {
		number = "RCT-" + time.Now().Format("20060102") + "-" + id[:6]
	}

	_, err := h.db.Exec(ctx, `
		INSERT INTO receipts
		  (id, company_id, number, receipt_type, status,
		   partner_id, partner_name, amount, allocated_amount,
		   receipt_date, reference, payment_method,
		   bank_account_id, cash_account_id,
		   currency, description, invoice_id, invoice_number)
		VALUES ($1,$2,$3,$4,'draft',
		  CASE WHEN $5='' THEN NULL ELSE $5::UUID END,
		  $6,$7,0,$8,$9,$10,
		  CASE WHEN $11='' THEN NULL ELSE $11::UUID END,
		  CASE WHEN $12='' THEN NULL ELSE $12::UUID END,
		  $13,$14,
		  CASE WHEN $15='' THEN NULL ELSE $15::UUID END,
		  NULLIF($16,''))
	`, id, companyID, number, req.ReceiptType,
		req.PartnerID, req.PartnerName, req.Amount,
		req.ReceiptDate, req.Reference, req.PaymentMethod,
		req.BankAccountID, req.CashAccountID,
		req.Currency, req.Description,
		req.InvoiceID, req.InvoiceNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "number": number, "message": "Receipt created"})
}

func (h *TreasuryHandler) UpdateReceipt(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	var req struct {
		PartnerName   string    `json:"partner_name"`
		Amount        float64   `json:"amount"`
		ReceiptDate   time.Time `json:"receipt_date"`
		Reference     string    `json:"reference"`
		Description   string    `json:"description"`
	}
	c.ShouldBindJSON(&req)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE receipts
		SET partner_name  = COALESCE(NULLIF($1,''), partner_name),
		    amount        = CASE WHEN $2 > 0 THEN $2 ELSE amount END,
		    receipt_date  = CASE WHEN $3 > '0001-01-01'::DATE THEN $3 ELSE receipt_date END,
		    reference     = COALESCE(NULLIF($4,''), reference),
		    description   = $5,
		    updated_at    = NOW()
		WHERE id = $6 AND company_id = $7 AND status = 'draft'
	`, req.PartnerName, req.Amount, req.ReceiptDate, req.Reference, req.Description, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Receipt updated"})
}

func (h *TreasuryHandler) ConfirmReceipt(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE receipts
		SET status = 'confirmed', confirmed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND status = 'draft'
	`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Receipt confirmed"})
}

func (h *TreasuryHandler) CancelReceipt(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE receipts SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND status IN ('draft','confirmed')
	`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Receipt cancelled"})
}

func (h *TreasuryHandler) DeleteReceipt(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		DELETE FROM receipts WHERE id = $1 AND company_id = $2 AND status = 'draft'
	`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Receipt deleted"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Bank Reconciliation
// ─────────────────────────────────────────────────────────────────────────────

func (h *TreasuryHandler) ListReconciliations(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			br.id,
			br.reference,
			br.status,
			br.period_start,
			br.period_end,
			ba.id::TEXT,
			ba.bank_name,
			ba.account_number,
			COALESCE(br.opening_balance, 0),
			COALESCE(br.closing_balance, 0),
			COALESCE(br.statement_balance, 0),
			COALESCE(br.system_balance, 0),
			br.difference,
			COALESCE(br.matched_items, 0),
			COALESCE(br.unmatched_items, 0),
			COALESCE(br.notes, ''),
			br.created_at
		FROM bank_reconciliations br
		JOIN bank_accounts ba ON ba.id = br.bank_account_id
		WHERE br.company_id = $1
		ORDER BY br.period_end DESC
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type ReconRow struct {
		ID               string    `json:"id"`
		Reference        string    `json:"reference"`
		Status           string    `json:"status"`
		PeriodStart      time.Time `json:"period_start"`
		PeriodEnd        time.Time `json:"period_end"`
		BankAccountID    string    `json:"bank_account_id"`
		BankName         string    `json:"bank_name"`
		AccountNumber    string    `json:"account_number"`
		OpeningBalance   float64   `json:"opening_balance"`
		ClosingBalance   float64   `json:"closing_balance"`
		StatementBalance float64   `json:"statement_balance"`
		SystemBalance    float64   `json:"system_balance"`
		Difference       float64   `json:"difference"`
		MatchedItems     int       `json:"matched_items"`
		UnmatchedItems   int       `json:"unmatched_items"`
		Notes            string    `json:"notes"`
		CreatedAt        time.Time `json:"created_at"`
	}

	var recons []ReconRow
	for rows.Next() {
		var r ReconRow
		if err := rows.Scan(
			&r.ID, &r.Reference, &r.Status,
			&r.PeriodStart, &r.PeriodEnd,
			&r.BankAccountID, &r.BankName, &r.AccountNumber,
			&r.OpeningBalance, &r.ClosingBalance,
			&r.StatementBalance, &r.SystemBalance, &r.Difference,
			&r.MatchedItems, &r.UnmatchedItems,
			&r.Notes, &r.CreatedAt,
		); err == nil {
			recons = append(recons, r)
		}
	}
	if recons == nil {
		recons = []ReconRow{}
	}
	c.JSON(http.StatusOK, recons)
}

func (h *TreasuryHandler) GetReconciliation(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var (
		recID, ref, status, bankAccID, bankName, accNum, notes string
		periodStart, periodEnd                                  time.Time
		openBal, closeBal, stmtBal, sysBal, diff               float64
		matchedItems, unmatchedItems                           int
		createdAt                                              time.Time
	)
	err := h.db.QueryRow(ctx, `
		SELECT br.id, br.reference, br.status,
		       br.period_start, br.period_end,
		       ba.id::TEXT, ba.bank_name, ba.account_number,
		       COALESCE(br.opening_balance,0), COALESCE(br.closing_balance,0),
		       COALESCE(br.statement_balance,0), COALESCE(br.system_balance,0),
		       br.difference,
		       COALESCE(br.matched_items,0), COALESCE(br.unmatched_items,0),
		       COALESCE(br.notes,''), br.created_at
		FROM bank_reconciliations br
		JOIN bank_accounts ba ON ba.id = br.bank_account_id
		WHERE br.id = $1 AND br.company_id = $2
	`, id, companyID).Scan(
		&recID, &ref, &status, &periodStart, &periodEnd,
		&bankAccID, &bankName, &accNum,
		&openBal, &closeBal, &stmtBal, &sysBal, &diff,
		&matchedItems, &unmatchedItems, &notes, &createdAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reconciliation not found"})
		return
	}

	// Fetch lines
	lineRows, _ := h.db.Query(ctx, `
		SELECT id, line_type, status, transaction_date,
		       COALESCE(description,''), COALESCE(reference,''),
		       COALESCE(debit_amount,0), COALESCE(credit_amount,0),
		       COALESCE(source_type,''), COALESCE(source_ref,'')
		FROM bank_reconciliation_lines
		WHERE reconciliation_id = $1
		ORDER BY transaction_date DESC
	`, id)

	type Line struct {
		ID              string    `json:"id"`
		LineType        string    `json:"line_type"`
		Status          string    `json:"status"`
		Date            time.Time `json:"transaction_date"`
		Description     string    `json:"description"`
		Reference       string    `json:"reference"`
		DebitAmount     float64   `json:"debit_amount"`
		CreditAmount    float64   `json:"credit_amount"`
		SourceType      string    `json:"source_type"`
		SourceRef       string    `json:"source_ref"`
	}

	var lines []Line
	if lineRows != nil {
		defer lineRows.Close()
		for lineRows.Next() {
			var l Line
			if err := lineRows.Scan(
				&l.ID, &l.LineType, &l.Status, &l.Date,
				&l.Description, &l.Reference,
				&l.DebitAmount, &l.CreditAmount,
				&l.SourceType, &l.SourceRef,
			); err == nil {
				lines = append(lines, l)
			}
		}
	}
	if lines == nil {
		lines = []Line{}
	}

	c.JSON(http.StatusOK, gin.H{
		"id": recID, "reference": ref, "status": status,
		"period_start": periodStart, "period_end": periodEnd,
		"bank_account_id": bankAccID, "bank_name": bankName, "account_number": accNum,
		"opening_balance": openBal, "closing_balance": closeBal,
		"statement_balance": stmtBal, "system_balance": sysBal, "difference": diff,
		"matched_items": matchedItems, "unmatched_items": unmatchedItems,
		"notes": notes, "created_at": createdAt, "lines": lines,
	})
}

func (h *TreasuryHandler) CreateReconciliation(c *gin.Context) {
	var req struct {
		BankAccountID    string    `json:"bank_account_id" binding:"required"`
		PeriodStart      time.Time `json:"period_start" binding:"required"`
		PeriodEnd        time.Time `json:"period_end" binding:"required"`
		OpeningBalance   float64   `json:"opening_balance"`
		StatementBalance float64   `json:"statement_balance"`
		Notes            string    `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// Generate reference
	var reference string
	h.db.QueryRow(ctx, `SELECT next_bank_recon_reference($1)`, companyID).Scan(&reference)
	if reference == "" {
		reference = "BRC-" + time.Now().Format("200601") + "-" + id[:4]
	}

	// Calculate system balance from bank movements
	var systemBalance float64
	h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(
		  CASE WHEN type IN ('inflow','transfer') THEN amount
		       ELSE -amount END
		), 0)
		FROM treasury_movements
		WHERE bank_account_id = $1
		  AND date BETWEEN $2 AND $3
		  AND company_id = $4
	`, req.BankAccountID, req.PeriodStart, req.PeriodEnd, companyID).Scan(&systemBalance)

	_, err := h.db.Exec(ctx, `
		INSERT INTO bank_reconciliations
		  (id, company_id, bank_account_id, reference, status,
		   period_start, period_end, opening_balance, statement_balance,
		   system_balance, notes)
		VALUES ($1,$2,$3::UUID,$4,'draft',$5,$6,$7,$8,$9,$10)
	`, id, companyID, req.BankAccountID, reference,
		req.PeriodStart, req.PeriodEnd,
		req.OpeningBalance, req.StatementBalance, systemBalance, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "reference": reference, "message": "Reconciliation created"})
}

func (h *TreasuryHandler) UpdateReconciliation(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	var req struct {
		StatementBalance float64 `json:"statement_balance"`
		Notes            string  `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE bank_reconciliations
		SET statement_balance = CASE WHEN $1 != 0 THEN $1 ELSE statement_balance END,
		    notes = $2,
		    updated_at = NOW()
		WHERE id = $3 AND company_id = $4 AND status IN ('draft','in_progress')
	`, req.StatementBalance, req.Notes, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Reconciliation updated"})
}

func (h *TreasuryHandler) CompleteReconciliation(c *gin.Context) {
	id := c.Param("id")
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE bank_reconciliations
		SET status = 'completed', completed_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND status IN ('draft','in_progress')
	`, id, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Update bank account last_reconciled date
	h.db.Exec(ctx, `
		UPDATE bank_accounts ba
		SET last_reconciled = (
		  SELECT period_end FROM bank_reconciliations WHERE id = $1
		), updated_at = NOW()
		WHERE ba.id = (SELECT bank_account_id FROM bank_reconciliations WHERE id = $1)
	`, id)
	c.JSON(http.StatusOK, gin.H{"message": "Reconciliation completed"})
}

func (h *TreasuryHandler) AddReconciliationLine(c *gin.Context) {
	reconID := c.Param("id")
	var req struct {
		LineType        string    `json:"line_type"`
		TransactionDate time.Time `json:"transaction_date"`
		Description     string    `json:"description"`
		Reference       string    `json:"reference"`
		DebitAmount     float64   `json:"debit_amount"`
		CreditAmount    float64   `json:"credit_amount"`
		SourceType      string    `json:"source_type"`
		SourceID        string    `json:"source_id"`
		SourceRef       string    `json:"source_ref"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.LineType == "" {
		req.LineType = "bank_statement"
	}
	if req.TransactionDate.IsZero() {
		req.TransactionDate = time.Now()
	}

	id := uuid.NewString()
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		INSERT INTO bank_reconciliation_lines
		  (id, reconciliation_id, line_type, status,
		   transaction_date, description, reference,
		   debit_amount, credit_amount,
		   source_type, source_id, source_ref)
		VALUES ($1,$2,$3,'unmatched',$4,$5,$6,$7,$8,$9,
		  CASE WHEN $10='' THEN NULL ELSE $10::UUID END,$11)
	`, id, reconID, req.LineType, req.TransactionDate,
		req.Description, req.Reference,
		req.DebitAmount, req.CreditAmount,
		req.SourceType, req.SourceID, req.SourceRef)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Line added"})
}

func (h *TreasuryHandler) MatchReconciliationLines(c *gin.Context) {
	var req struct {
		LineID1 string `json:"line_id_1" binding:"required"`
		LineID2 string `json:"line_id_2" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx := context.Background()
	_, err := h.db.Exec(ctx, `
		UPDATE bank_reconciliation_lines
		SET status = 'matched',
		    matched_line_id = CASE WHEN id = $1 THEN $2::UUID ELSE $1::UUID END,
		    matched_at = NOW()
		WHERE id IN ($1, $2)
	`, req.LineID1, req.LineID2)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Lines matched"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Reports
// ─────────────────────────────────────────────────────────────────────────────

func (h *TreasuryHandler) AgingReport(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT
			p.id, p.type, COALESCE(p.partner_name,''),
			p.date,
			p.amount,
			COALESCE(p.allocated_amount, 0),
			p.amount - COALESCE(p.allocated_amount, 0) AS outstanding,
			CURRENT_DATE - p.date::date AS days_out,
			CASE
			  WHEN CURRENT_DATE - p.date::date <= 30 THEN '0-30'
			  WHEN CURRENT_DATE - p.date::date <= 60 THEN '31-60'
			  WHEN CURRENT_DATE - p.date::date <= 90 THEN '61-90'
			  ELSE '90+'
			END AS bucket
		FROM payments p
		WHERE p.company_id = $1
		  AND p.amount > COALESCE(p.allocated_amount, 0)
		ORDER BY days_out DESC
		LIMIT 200
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type AgingRow struct {
		ID          string    `json:"id"`
		Type        string    `json:"type"`
		PartnerName string    `json:"partner_name"`
		Date        time.Time `json:"date"`
		Amount      float64   `json:"amount"`
		Allocated   float64   `json:"allocated"`
		Outstanding float64   `json:"outstanding"`
		DaysOut     int       `json:"days_outstanding"`
		Bucket      string    `json:"aging_bucket"`
	}

	var items []AgingRow
	for rows.Next() {
		var a AgingRow
		if err := rows.Scan(
			&a.ID, &a.Type, &a.PartnerName, &a.Date,
			&a.Amount, &a.Allocated, &a.Outstanding,
			&a.DaysOut, &a.Bucket,
		); err == nil {
			items = append(items, a)
		}
	}
	if items == nil {
		items = []AgingRow{}
	}

	// Bucket totals
	buckets := map[string]float64{"0-30": 0, "31-60": 0, "61-90": 0, "90+": 0}
	for _, i := range items {
		buckets[i.Bucket] += i.Outstanding
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "buckets": buckets})
}

func (h *TreasuryHandler) CashPositionReport(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	// Cash accounts
	cashRows, _ := h.db.Query(ctx, `
		SELECT id, name, currency, balance FROM cash_accounts
		WHERE company_id = $1 AND is_active = true ORDER BY name
	`, companyID)

	type AccBalance struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		Currency string  `json:"currency"`
		Balance  float64 `json:"balance"`
	}

	var cashAccounts []AccBalance
	var totalCash float64
	if cashRows != nil {
		defer cashRows.Close()
		for cashRows.Next() {
			var a AccBalance
			if err := cashRows.Scan(&a.ID, &a.Name, &a.Currency, &a.Balance); err == nil {
				cashAccounts = append(cashAccounts, a)
				totalCash += a.Balance
			}
		}
	}

	// Bank accounts
	bankRows, _ := h.db.Query(ctx, `
		SELECT id, bank_name, account_number, currency, balance
		FROM bank_accounts WHERE company_id = $1 AND is_active = true ORDER BY bank_name
	`, companyID)

	type BankBalance struct {
		ID            string  `json:"id"`
		BankName      string  `json:"bank_name"`
		AccountNumber string  `json:"account_number"`
		Currency      string  `json:"currency"`
		Balance       float64 `json:"balance"`
	}

	var bankAccounts []BankBalance
	var totalBank float64
	if bankRows != nil {
		defer bankRows.Close()
		for bankRows.Next() {
			var b BankBalance
			if err := bankRows.Scan(&b.ID, &b.BankName, &b.AccountNumber, &b.Currency, &b.Balance); err == nil {
				bankAccounts = append(bankAccounts, b)
				totalBank += b.Balance
			}
		}
	}

	// Pending cheques total
	var pendingCheques float64
	h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0)
		FROM cheques WHERE company_id = $1 AND status = 'pending'
	`, companyID).Scan(&pendingCheques)

	// This month inflows/outflows
	var monthInflow, monthOutflow float64
	h.db.QueryRow(ctx, `
		SELECT
		  COALESCE(SUM(CASE WHEN type = 'inflow' THEN amount ELSE 0 END), 0),
		  COALESCE(SUM(CASE WHEN type = 'outflow' THEN amount ELSE 0 END), 0)
		FROM treasury_movements
		WHERE company_id = $1
		  AND date >= DATE_TRUNC('month', CURRENT_DATE)
	`, companyID).Scan(&monthInflow, &monthOutflow)

	// Recent movements (last 10)
	movRows, _ := h.db.Query(ctx, `
		SELECT type, amount, date, COALESCE(reference,''), COALESCE(notes,'')
		FROM treasury_movements
		WHERE company_id = $1
		ORDER BY date DESC LIMIT 10
	`, companyID)

	type RecentMov struct {
		Type      string    `json:"type"`
		Amount    float64   `json:"amount"`
		Date      time.Time `json:"date"`
		Reference string    `json:"reference"`
		Notes     string    `json:"notes"`
	}

	var recentMovements []RecentMov
	if movRows != nil {
		defer movRows.Close()
		for movRows.Next() {
			var m RecentMov
			if err := movRows.Scan(&m.Type, &m.Amount, &m.Date, &m.Reference, &m.Notes); err == nil {
				recentMovements = append(recentMovements, m)
			}
		}
	}

	if cashAccounts == nil {
		cashAccounts = []AccBalance{}
	}
	if bankAccounts == nil {
		bankAccounts = []BankBalance{}
	}
	if recentMovements == nil {
		recentMovements = []RecentMov{}
	}

	c.JSON(http.StatusOK, gin.H{
		"cash_accounts":    cashAccounts,
		"bank_accounts":    bankAccounts,
		"total_cash":       totalCash,
		"total_bank":       totalBank,
		"total_position":   totalCash + totalBank,
		"pending_cheques":  pendingCheques,
		"month_inflow":     monthInflow,
		"month_outflow":    monthOutflow,
		"net_month":        monthInflow - monthOutflow,
		"recent_movements": recentMovements,
		"date":             time.Now().Format("2006-01-02"),
	})
}

func (h *TreasuryHandler) TreasuryReport(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	year := c.Query("year")
	if year == "" {
		year = time.Now().Format("2006")
	}

	// Monthly cash flow
	cfRows, _ := h.db.Query(ctx, `
		SELECT
			TO_CHAR(date, 'YYYY-MM') AS month,
			COALESCE(SUM(CASE WHEN type = 'inflow' THEN amount ELSE 0 END), 0)  AS inflow,
			COALESCE(SUM(CASE WHEN type = 'outflow' THEN amount ELSE 0 END), 0) AS outflow
		FROM treasury_movements
		WHERE company_id = $1
		  AND EXTRACT(YEAR FROM date) = $2::INT
		GROUP BY TO_CHAR(date, 'YYYY-MM')
		ORDER BY month
	`, companyID, year)

	type MonthlyCF struct {
		Month   string  `json:"month"`
		Inflow  float64 `json:"inflow"`
		Outflow float64 `json:"outflow"`
		Net     float64 `json:"net"`
	}

	var cashFlow []MonthlyCF
	var totalInflow, totalOutflow float64
	if cfRows != nil {
		defer cfRows.Close()
		for cfRows.Next() {
			var cf MonthlyCF
			if err := cfRows.Scan(&cf.Month, &cf.Inflow, &cf.Outflow); err == nil {
				cf.Net = cf.Inflow - cf.Outflow
				totalInflow += cf.Inflow
				totalOutflow += cf.Outflow
				cashFlow = append(cashFlow, cf)
			}
		}
	}

	// Payment totals by type
	payRows, _ := h.db.Query(ctx, `
		SELECT type, COUNT(*), COALESCE(SUM(amount),0)
		FROM payments WHERE company_id = $1
		  AND EXTRACT(YEAR FROM date) = $2::INT
		GROUP BY type
	`, companyID, year)

	type PaySummary struct {
		Type   string  `json:"type"`
		Count  int     `json:"count"`
		Amount float64 `json:"amount"`
	}

	var paymentSummary []PaySummary
	if payRows != nil {
		defer payRows.Close()
		for payRows.Next() {
			var ps PaySummary
			if err := payRows.Scan(&ps.Type, &ps.Count, &ps.Amount); err == nil {
				paymentSummary = append(paymentSummary, ps)
			}
		}
	}

	// Cheque stats
	var chequePending, chequeDeposited, chequeBounced int
	var chequePendingAmt, chequeDepositedAmt float64
	h.db.QueryRow(ctx, `
		SELECT
		  COUNT(*) FILTER (WHERE status='pending'),
		  COUNT(*) FILTER (WHERE status='deposited'),
		  COUNT(*) FILTER (WHERE status='bounced'),
		  COALESCE(SUM(amount) FILTER (WHERE status='pending'), 0),
		  COALESCE(SUM(amount) FILTER (WHERE status='deposited'), 0)
		FROM cheques WHERE company_id = $1
		  AND EXTRACT(YEAR FROM issue_date) = $2::INT
	`, companyID, year).Scan(
		&chequePending, &chequeDeposited, &chequeBounced,
		&chequePendingAmt, &chequeDepositedAmt,
	)

	// Receipt stats
	var receiptCount int
	var receiptTotal float64
	h.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(amount),0)
		FROM receipts WHERE company_id = $1
		  AND EXTRACT(YEAR FROM receipt_date) = $2::INT
		  AND status = 'confirmed'
	`, companyID, year).Scan(&receiptCount, &receiptTotal)

	if cashFlow == nil {
		cashFlow = []MonthlyCF{}
	}
	if paymentSummary == nil {
		paymentSummary = []PaySummary{}
	}

	c.JSON(http.StatusOK, gin.H{
		"year":          year,
		"cash_flow":     cashFlow,
		"total_inflow":  totalInflow,
		"total_outflow": totalOutflow,
		"net_flow":      totalInflow - totalOutflow,
		"payment_summary": paymentSummary,
		"cheques": gin.H{
			"pending":       chequePending,
			"deposited":     chequeDeposited,
			"bounced":       chequeBounced,
			"pending_amt":   chequePendingAmt,
			"deposited_amt": chequeDepositedAmt,
		},
		"receipts": gin.H{
			"count":  receiptCount,
			"amount": receiptTotal,
		},
	})
}

// treasuryItoa is defined in projects.go as itoa(); reuse it
// (Go allows cross-file helper usage within same package)
