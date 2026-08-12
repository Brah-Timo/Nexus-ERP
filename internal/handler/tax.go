package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"nexus-erp/internal/middleware"
)

// ─── Tax Handler ──────────────────────────────────────────────────────────────

type TaxHandler struct{ db *pgxpool.Pool }

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func taxItoa(n int) string { return strconv.Itoa(n) }

func taxParsePeriod(c *gin.Context) (year, month int) {
	now := time.Now()
	year = now.Year()
	month = int(now.Month())
	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}
	if m := c.Query("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil {
			month = v
		}
	}
	return
}

// ─────────────────────────────────────────────────────────────────────────────
// G50 Declarations
// ─────────────────────────────────────────────────────────────────────────────

// ListDeclarations — GET /tax/declarations
func (h *TaxHandler) ListDeclarations(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	year := c.Query("year")
	dtype := c.Query("type")
	status := c.Query("status")

	query := `
		SELECT
			d.id, d.reference, d.declaration_type, d.period_type,
			d.period_year, d.period_month, d.period_quarter, d.status,
			d.tva_collected, d.tva_deductible, d.tva_credit_bf,
			d.tva_net_due, d.tva_credit_carry,
			d.tap_base, d.tap_rate, d.tap_net_due,
			d.ibs_taxable_income, d.ibs_rate, d.ibs_net_due,
			d.stamp_tax_amount, d.irg_wages_amount, d.irg_fees_amount,
			d.total_tax_due,
			COALESCE(SUM(tp.amount_paid), 0) AS total_paid,
			d.total_tax_due - COALESCE(SUM(tp.amount_paid), 0) AS balance_due,
			d.submitted_at, d.submission_ref, d.notes, d.created_at
		FROM tax_declarations d
		LEFT JOIN tax_payments tp ON tp.declaration_id = d.id AND tp.status != 'cancelled'
		WHERE d.company_id = $1
	`
	args := []interface{}{companyID}
	n := 2
	if year != "" {
		query += fmt.Sprintf(" AND d.period_year = $%d", n)
		args = append(args, year)
		n++
	}
	if dtype != "" {
		query += fmt.Sprintf(" AND d.declaration_type = $%d", n)
		args = append(args, dtype)
		n++
	}
	if status != "" {
		query += fmt.Sprintf(" AND d.status = $%d", n)
		args = append(args, status)
		n++
	}
	query += " GROUP BY d.id ORDER BY d.period_year DESC, d.period_month DESC, d.created_at DESC"

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var (
			id, reference, declType, periodType, status, submRef, notes string
			periodYear, periodMonth, periodQuarter                        int
			periodMonthPtr, periodQtrPtr                                  *int
			tvaCollected, tvaDeductible, tvaCreditBF                      float64
			tvaNetDue, tvaCreditCarry                                      float64
			tapBase, tapRate, tapNetDue                                    float64
			ibsTaxable, ibsRate, ibsNetDue                                 float64
			stampTax, irgWages, irgFees                                    float64
			totalDue, totalPaid, balanceDue                                float64
			submittedAt                                                    *time.Time
			createdAt                                                      time.Time
		)
		_ = periodMonth
		_ = periodQuarter
		err = rows.Scan(
			&id, &reference, &declType, &periodType,
			&periodYear, &periodMonthPtr, &periodQtrPtr, &status,
			&tvaCollected, &tvaDeductible, &tvaCreditBF,
			&tvaNetDue, &tvaCreditCarry,
			&tapBase, &tapRate, &tapNetDue,
			&ibsTaxable, &ibsRate, &ibsNetDue,
			&stampTax, &irgWages, &irgFees,
			&totalDue, &totalPaid, &balanceDue,
			&submittedAt, &submRef, &notes, &createdAt,
		)
		if err != nil {
			continue
		}
		rec := map[string]interface{}{
			"id": id, "reference": reference,
			"declaration_type": declType, "period_type": periodType,
			"period_year": periodYear, "period_month": periodMonthPtr,
			"period_quarter": periodQtrPtr, "status": status,
			"tva_collected": tvaCollected, "tva_deductible": tvaDeductible,
			"tva_credit_bf": tvaCreditBF, "tva_net_due": tvaNetDue,
			"tva_credit_carry": tvaCreditCarry,
			"tap_base": tapBase, "tap_rate": tapRate, "tap_net_due": tapNetDue,
			"ibs_taxable_income": ibsTaxable, "ibs_rate": ibsRate, "ibs_net_due": ibsNetDue,
			"stamp_tax_amount": stampTax, "irg_wages_amount": irgWages, "irg_fees_amount": irgFees,
			"total_tax_due": totalDue, "total_paid": totalPaid, "balance_due": balanceDue,
			"submitted_at": submittedAt, "submission_ref": submRef,
			"notes": notes, "created_at": createdAt,
		}
		list = append(list, rec)
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

// GetDeclaration — GET /tax/declarations/:id
func (h *TaxHandler) GetDeclaration(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	var rec map[string]interface{}
	row := h.db.QueryRow(ctx, `
		SELECT
			d.id, d.reference, d.declaration_type, d.period_type,
			d.period_year, d.period_month, d.period_quarter, d.status,
			d.tva_collected, d.tva_deductible, d.tva_credit_bf,
			d.tva_net_due, d.tva_credit_carry,
			d.tap_base, d.tap_rate, d.tap_net_due,
			d.ibs_taxable_income, d.ibs_rate, d.ibs_net_due,
			d.stamp_tax_amount, d.irg_wages_amount, d.irg_fees_amount,
			d.total_tax_due,
			COALESCE(SUM(tp.amount_paid), 0)                             AS total_paid,
			d.total_tax_due - COALESCE(SUM(tp.amount_paid), 0)          AS balance_due,
			d.submitted_at, d.submission_ref, d.notes, d.created_at, d.updated_at
		FROM tax_declarations d
		LEFT JOIN tax_payments tp ON tp.declaration_id = d.id AND tp.status != 'cancelled'
		WHERE d.id = $1 AND d.company_id = $2
		GROUP BY d.id
	`, id, companyID)

	var (
		rid, reference, declType, periodType, status, submRef, notes string
		periodYear                                                      int
		periodMonthPtr, periodQtrPtr                                    *int
		tvaCollected, tvaDeductible, tvaCreditBF                        float64
		tvaNetDue, tvaCreditCarry                                        float64
		tapBase, tapRate, tapNetDue                                      float64
		ibsTaxable, ibsRate, ibsNetDue                                   float64
		stampTax, irgWages, irgFees                                      float64
		totalDue, totalPaid, balanceDue                                  float64
		submittedAt                                                      *time.Time
		createdAt, updatedAt                                             time.Time
	)
	err := row.Scan(
		&rid, &reference, &declType, &periodType,
		&periodYear, &periodMonthPtr, &periodQtrPtr, &status,
		&tvaCollected, &tvaDeductible, &tvaCreditBF,
		&tvaNetDue, &tvaCreditCarry,
		&tapBase, &tapRate, &tapNetDue,
		&ibsTaxable, &ibsRate, &ibsNetDue,
		&stampTax, &irgWages, &irgFees,
		&totalDue, &totalPaid, &balanceDue,
		&submittedAt, &submRef, &notes, &createdAt, &updatedAt,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Declaration not found"})
		return
	}
	rec = map[string]interface{}{
		"id": rid, "reference": reference,
		"declaration_type": declType, "period_type": periodType,
		"period_year": periodYear, "period_month": periodMonthPtr,
		"period_quarter": periodQtrPtr, "status": status,
		"tva_collected": tvaCollected, "tva_deductible": tvaDeductible,
		"tva_credit_bf": tvaCreditBF, "tva_net_due": tvaNetDue,
		"tva_credit_carry": tvaCreditCarry,
		"tap_base": tapBase, "tap_rate": tapRate, "tap_net_due": tapNetDue,
		"ibs_taxable_income": ibsTaxable, "ibs_rate": ibsRate, "ibs_net_due": ibsNetDue,
		"stamp_tax_amount": stampTax, "irg_wages_amount": irgWages, "irg_fees_amount": irgFees,
		"total_tax_due": totalDue, "total_paid": totalPaid, "balance_due": balanceDue,
		"submitted_at": submittedAt, "submission_ref": submRef,
		"notes": notes, "created_at": createdAt, "updated_at": updatedAt,
	}
	c.JSON(http.StatusOK, rec)
}

// CreateDeclaration — POST /tax/declarations
func (h *TaxHandler) CreateDeclaration(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.NewString()
	declType := stringVal(req, "declaration_type", "g50")
	periodType := stringVal(req, "period_type", "monthly")
	periodYear := intVal(req, "period_year", time.Now().Year())
	periodMonth := intPtrVal(req, "period_month")
	periodQtr := intPtrVal(req, "period_quarter")

	tvaCollected := floatVal(req, "tva_collected", 0)
	tvaDeductible := floatVal(req, "tva_deductible", 0)
	tvaCreditBF := floatVal(req, "tva_credit_bf", 0)
	tapBase := floatVal(req, "tap_base", 0)
	tapRate := floatVal(req, "tap_rate", 0.02)
	tapReduction := floatVal(req, "tap_reduction", 0)
	ibsTaxable := floatVal(req, "ibs_taxable_income", 0)
	ibsRate := floatVal(req, "ibs_rate", 0.23)
	ibsPrepay := floatVal(req, "ibs_prepayments", 0)
	stampTax := floatVal(req, "stamp_tax_amount", 0)
	irgWages := floatVal(req, "irg_wages_amount", 0)
	irgFees := floatVal(req, "irg_fees_amount", 0)
	notes := stringVal(req, "notes", "")

	_, err := h.db.Exec(ctx, `
		INSERT INTO tax_declarations (
			id, company_id, declaration_type, period_type,
			period_year, period_month, period_quarter,
			tva_collected, tva_deductible, tva_credit_bf,
			tap_base, tap_rate, tap_reduction,
			ibs_taxable_income, ibs_rate, ibs_prepayments,
			stamp_tax_amount, irg_wages_amount, irg_fees_amount, notes
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20
		)`,
		id, companyID, declType, periodType,
		periodYear, periodMonth, periodQtr,
		tvaCollected, tvaDeductible, tvaCreditBF,
		tapBase, tapRate, tapReduction,
		ibsTaxable, ibsRate, ibsPrepay,
		stampTax, irgWages, irgFees, notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Declaration created"})
}

// UpdateDeclaration — PUT /tax/declarations/:id
func (h *TaxHandler) UpdateDeclaration(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(ctx, `
		UPDATE tax_declarations SET
			tva_collected      = COALESCE($3, tva_collected),
			tva_deductible     = COALESCE($4, tva_deductible),
			tva_credit_bf      = COALESCE($5, tva_credit_bf),
			tap_base           = COALESCE($6, tap_base),
			tap_rate           = COALESCE($7, tap_rate),
			tap_reduction      = COALESCE($8, tap_reduction),
			ibs_taxable_income = COALESCE($9, ibs_taxable_income),
			ibs_rate           = COALESCE($10, ibs_rate),
			ibs_prepayments    = COALESCE($11, ibs_prepayments),
			stamp_tax_amount   = COALESCE($12, stamp_tax_amount),
			irg_wages_amount   = COALESCE($13, irg_wages_amount),
			irg_fees_amount    = COALESCE($14, irg_fees_amount),
			notes              = COALESCE($15, notes),
			updated_at         = NOW()
		WHERE id = $1 AND company_id = $2`,
		id, companyID,
		floatPtrVal(req, "tva_collected"),
		floatPtrVal(req, "tva_deductible"),
		floatPtrVal(req, "tva_credit_bf"),
		floatPtrVal(req, "tap_base"),
		floatPtrVal(req, "tap_rate"),
		floatPtrVal(req, "tap_reduction"),
		floatPtrVal(req, "ibs_taxable_income"),
		floatPtrVal(req, "ibs_rate"),
		floatPtrVal(req, "ibs_prepayments"),
		floatPtrVal(req, "stamp_tax_amount"),
		floatPtrVal(req, "irg_wages_amount"),
		floatPtrVal(req, "irg_fees_amount"),
		strPtrVal(req, "notes"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Declaration updated"})
}

// SubmitDeclaration — POST /tax/declarations/:id/submit
func (h *TaxHandler) SubmitDeclaration(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	var req struct {
		SubmissionRef string `json:"submission_ref"`
	}
	_ = c.ShouldBindJSON(&req)

	_, err := h.db.Exec(ctx, `
		UPDATE tax_declarations
		SET status = 'submitted', submitted_at = NOW(),
		    submission_ref = NULLIF($3, ''), updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND status = 'draft'`,
		id, companyID, req.SubmissionRef,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Declaration submitted"})
}

// AmendDeclaration — POST /tax/declarations/:id/amend
func (h *TaxHandler) AmendDeclaration(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		UPDATE tax_declarations
		SET status = 'amended', updated_at = NOW()
		WHERE id = $1 AND company_id = $2 AND status IN ('submitted','accepted')`,
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Declaration amended"})
}

// DeleteDeclaration — DELETE /tax/declarations/:id
func (h *TaxHandler) DeleteDeclaration(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	_, err := h.db.Exec(ctx,
		`DELETE FROM tax_declarations WHERE id = $1 AND company_id = $2 AND status = 'draft'`,
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Declaration deleted"})
}

// ─────────────────────────────────────────────────────────────────────────────
// G50 Auto-compute (build from actual invoices)
// ─────────────────────────────────────────────────────────────────────────────

// GetG50 — GET /tax/declarations/g50  (compute on the fly from invoices)
func (h *TaxHandler) GetG50(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	year, month := taxParsePeriod(c)

	var tvaCollected, tvaDeductible, tapSales, stampTax float64

	// TVA collected from sales invoices
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(tva_amount), 0)
		FROM sales_invoices
		WHERE company_id = $1
		  AND EXTRACT(YEAR  FROM date) = $2
		  AND EXTRACT(MONTH FROM date) = $3
		  AND status NOT IN ('cancelled','draft')
	`, companyID, year, month).Scan(&tvaCollected)

	// TVA deductible from purchase invoices
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(tva_amount), 0)
		FROM purchase_invoices
		WHERE company_id = $1
		  AND EXTRACT(YEAR  FROM date) = $2
		  AND EXTRACT(MONTH FROM date) = $3
		  AND status NOT IN ('cancelled','draft')
	`, companyID, year, month).Scan(&tvaDeductible)

	// TAP from sales invoices (2% on sub_total)
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(sub_total) * 0.02, 0)
		FROM sales_invoices
		WHERE company_id = $1
		  AND EXTRACT(YEAR  FROM date) = $2
		  AND EXTRACT(MONTH FROM date) = $3
		  AND status NOT IN ('cancelled','draft')
	`, companyID, year, month).Scan(&tapSales)

	// Stamp tax from sales invoices
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(stamp_tax), 0)
		FROM sales_invoices
		WHERE company_id = $1
		  AND EXTRACT(YEAR  FROM date) = $2
		  AND EXTRACT(MONTH FROM date) = $3
		  AND status NOT IN ('cancelled','draft')
	`, companyID, year, month).Scan(&stampTax)

	// Credit BF from previous period declaration
	var creditBF float64
	prevMonth := month - 1
	prevYear := year
	if prevMonth == 0 {
		prevMonth = 12
		prevYear--
	}
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(tva_credit_carry, 0)
		FROM tax_declarations
		WHERE company_id = $1
		  AND declaration_type = 'g50'
		  AND period_year = $2 AND period_month = $3
		  AND status != 'draft'
		ORDER BY created_at DESC LIMIT 1
	`, companyID, prevYear, prevMonth).Scan(&creditBF)

	tvaNet := tvaCollected - tvaDeductible - creditBF
	tvaCreditCarry := 0.0
	if tvaNet < 0 {
		tvaCreditCarry = -tvaNet
		tvaNet = 0
	}
	totalDue := tvaNet + tapSales + stampTax

	c.JSON(http.StatusOK, gin.H{
		"period_year":      year,
		"period_month":     month,
		"tva_collected":    tvaCollected,
		"tva_deductible":   tvaDeductible,
		"tva_credit_bf":    creditBF,
		"tva_net_due":      tvaNet,
		"tva_credit_carry": tvaCreditCarry,
		"tap_base":         tvaCollected / 1.19 * 19,
		"tap_rate":         0.02,
		"tap_amount":       tapSales,
		"stamp_tax_amount": stampTax,
		"total_tax_due":    totalDue,
	})
}

// SubmitG50 — POST /tax/declarations/g50 (save computed G50 as declaration)
func (h *TaxHandler) SubmitG50(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	year := intVal(req, "period_year", time.Now().Year())
	month := intVal(req, "period_month", int(time.Now().Month()))

	// Check for existing draft
	var existID string
	_ = h.db.QueryRow(ctx, `
		SELECT id FROM tax_declarations
		WHERE company_id = $1 AND declaration_type = 'g50'
		  AND period_year = $2 AND period_month = $3 AND status = 'draft'
		LIMIT 1
	`, companyID, year, month).Scan(&existID)

	id := existID
	if id == "" {
		id = uuid.NewString()
		_, err := h.db.Exec(ctx, `
			INSERT INTO tax_declarations (
				id, company_id, declaration_type, period_type,
				period_year, period_month,
				tva_collected, tva_deductible, tva_credit_bf,
				tap_base, tap_rate, tap_reduction,
				stamp_tax_amount, irg_wages_amount, irg_fees_amount, notes,
				status
			) VALUES ($1,$2,'g50','monthly',$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,'submitted')`,
			id, companyID, year, month,
			floatVal(req, "tva_collected", 0),
			floatVal(req, "tva_deductible", 0),
			floatVal(req, "tva_credit_bf", 0),
			floatVal(req, "tap_base", 0),
			floatVal(req, "tap_rate", 0.02),
			0.0,
			floatVal(req, "stamp_tax_amount", 0),
			floatVal(req, "irg_wages_amount", 0),
			floatVal(req, "irg_fees_amount", 0),
			stringVal(req, "notes", ""),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		_, _ = h.db.Exec(ctx, `
			UPDATE tax_declarations SET
				tva_collected    = $3, tva_deductible = $4, tva_credit_bf = $5,
				tap_base         = $6, tap_rate = $7,
				stamp_tax_amount = $8, irg_wages_amount = $9, irg_fees_amount = $10,
				notes            = $11,
				status           = 'submitted', submitted_at = NOW(), updated_at = NOW()
			WHERE id = $1 AND company_id = $2`,
			id, companyID,
			floatVal(req, "tva_collected", 0),
			floatVal(req, "tva_deductible", 0),
			floatVal(req, "tva_credit_bf", 0),
			floatVal(req, "tap_base", 0),
			floatVal(req, "tap_rate", 0.02),
			floatVal(req, "stamp_tax_amount", 0),
			floatVal(req, "irg_wages_amount", 0),
			floatVal(req, "irg_fees_amount", 0),
			stringVal(req, "notes", ""),
		)
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "G50 declaration saved and submitted"})
}

// GetIBS — GET /tax/declarations/ibs
func (h *TaxHandler) GetIBS(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	year, _ := taxParsePeriod(c)

	var totalRevenue, totalExpenses float64

	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(sub_total), 0)
		FROM sales_invoices
		WHERE company_id = $1
		  AND EXTRACT(YEAR FROM date) = $2
		  AND status NOT IN ('cancelled','draft')
	`, companyID, year).Scan(&totalRevenue)

	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(sub_total), 0)
		FROM purchase_invoices
		WHERE company_id = $1
		  AND EXTRACT(YEAR FROM date) = $2
		  AND status NOT IN ('cancelled','draft')
	`, companyID, year).Scan(&totalExpenses)

	taxableIncome := totalRevenue - totalExpenses
	ibsRate := 0.23
	if taxableIncome < 0 {
		taxableIncome = 0
	}
	ibsAmount := taxableIncome * ibsRate

	// IBS prepayments already paid in quarterly installments
	var prepayments float64
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(tp.amount_paid), 0)
		FROM tax_payments tp
		JOIN tax_declarations d ON tp.declaration_id = d.id
		WHERE d.company_id = $1
		  AND d.declaration_type = 'ibs'
		  AND d.period_year = $2
		  AND tp.status != 'cancelled'
	`, companyID, year).Scan(&prepayments)

	netDue := ibsAmount - prepayments
	if netDue < 0 {
		netDue = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"period_year":       year,
		"total_revenue":     totalRevenue,
		"total_expenses":    totalExpenses,
		"taxable_income":    taxableIncome,
		"ibs_rate":          ibsRate,
		"ibs_amount":        ibsAmount,
		"prepayments":       prepayments,
		"net_due":           netDue,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// VAT Register
// ─────────────────────────────────────────────────────────────────────────────

// GetVATRegister — GET /tax/vat-register
func (h *TaxHandler) GetVATRegister(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	year, month := taxParsePeriod(c)
	regType := c.Query("type") // 'sales' | 'purchase' | ''

	query := `
		SELECT
			vr.id, vr.register_type, vr.document_date, vr.document_number,
			vr.partner_name, vr.partner_tax_id,
			vr.taxable_base, vr.vat_rate, vr.vat_amount, vr.total_amount,
			vr.is_exported, vr.created_at
		FROM vat_register vr
		WHERE vr.company_id = $1
		  AND vr.period_year = $2
		  AND vr.period_month = $3
	`
	args := []interface{}{companyID, year, month}
	n := 4
	if regType != "" {
		query += fmt.Sprintf(" AND vr.register_type = $%d", n)
		args = append(args, regType)
		n++
	}
	query += " ORDER BY vr.document_date DESC, vr.document_number DESC"

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var entries []map[string]interface{}
	var totalBase, totalVAT, totalAmount float64

	for rows.Next() {
		var (
			id, regTypeStr, docNum, partnerName string
			partnerTaxID                         *string
			docDate                              time.Time
			taxableBase, vatRate, vatAmt, total  float64
			isExported                           bool
			createdAt                            time.Time
		)
		err = rows.Scan(
			&id, &regTypeStr, &docDate, &docNum,
			&partnerName, &partnerTaxID,
			&taxableBase, &vatRate, &vatAmt, &total,
			&isExported, &createdAt,
		)
		if err != nil {
			continue
		}
		totalBase += taxableBase
		totalVAT += vatAmt
		totalAmount += total
		entries = append(entries, map[string]interface{}{
			"id": id, "register_type": regTypeStr,
			"document_date": docDate, "document_number": docNum,
			"partner_name": partnerName, "partner_tax_id": partnerTaxID,
			"taxable_base": taxableBase, "vat_rate": vatRate,
			"vat_amount": vatAmt, "total_amount": total,
			"is_exported": isExported, "created_at": createdAt,
		})
	}
	if entries == nil {
		entries = []map[string]interface{}{}
	}

	// Also pull from actual invoices if no manual entries exist for this period
	if len(entries) == 0 {
		entries = h.buildVATRegisterFromInvoices(ctx, companyID, year, month, regType)
		for _, e := range entries {
			totalBase += e["taxable_base"].(float64)
			totalVAT += e["vat_amount"].(float64)
			totalAmount += e["total_amount"].(float64)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"period_year":  year,
		"period_month": month,
		"entries":      entries,
		"totals": gin.H{
			"total_base":   totalBase,
			"total_vat":    totalVAT,
			"total_amount": totalAmount,
		},
	})
}

func (h *TaxHandler) buildVATRegisterFromInvoices(
	ctx context.Context,
	companyID string, year, month int, regType string,
) []map[string]interface{} {
	var result []map[string]interface{}

	if regType == "" || regType == "sales" {
		rows, err := h.db.Query(ctx, `
			SELECT
				id, date, number, customer_name,
				COALESCE(tax_id, '') AS tax_id,
				sub_total, tva_amount, total_amount
			FROM sales_invoices
			WHERE company_id = $1
			  AND EXTRACT(YEAR  FROM date) = $2
			  AND EXTRACT(MONTH FROM date) = $3
			  AND status NOT IN ('cancelled','draft')
			ORDER BY date DESC, number DESC
		`, companyID, year, month)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var (
					id, num, custName, taxID string
					docDate                  time.Time
					base, vat, total         float64
				)
				vatRate := 0.19
				if rows.Scan(&id, &docDate, &num, &custName, &taxID, &base, &vat, &total) == nil {
					if base > 0 {
						vatRate = vat / base
					}
					result = append(result, map[string]interface{}{
						"id": id, "register_type": "sales",
						"document_date": docDate, "document_number": num,
						"partner_name": custName, "partner_tax_id": taxID,
						"taxable_base": base, "vat_rate": vatRate,
						"vat_amount": vat, "total_amount": total,
						"is_exported": false,
					})
				}
			}
		}
	}

	if regType == "" || regType == "purchase" {
		rows, err := h.db.Query(ctx, `
			SELECT
				pi.id, pi.date, pi.number, s.name AS supplier_name,
				COALESCE(s.tax_id, '') AS tax_id,
				pi.sub_total, pi.tva_amount, pi.total_amount
			FROM purchase_invoices pi
			LEFT JOIN suppliers s ON pi.supplier_id = s.id
			WHERE pi.company_id = $1
			  AND EXTRACT(YEAR  FROM pi.date) = $2
			  AND EXTRACT(MONTH FROM pi.date) = $3
			  AND pi.status NOT IN ('cancelled','draft')
			ORDER BY pi.date DESC, pi.number DESC
		`, companyID, year, month)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var (
					id, num, supplierName, taxID string
					docDate                      time.Time
					base, vat, total             float64
				)
				vatRate := 0.19
				if rows.Scan(&id, &docDate, &num, &supplierName, &taxID, &base, &vat, &total) == nil {
					if base > 0 {
						vatRate = vat / base
					}
					result = append(result, map[string]interface{}{
						"id": id, "register_type": "purchase",
						"document_date": docDate, "document_number": num,
						"partner_name": supplierName, "partner_tax_id": taxID,
						"taxable_base": base, "vat_rate": vatRate,
						"vat_amount": vat, "total_amount": total,
						"is_exported": false,
					})
				}
			}
		}
	}

	if result == nil {
		return []map[string]interface{}{}
	}
	return result
}

// CreateVATEntry — POST /tax/vat-register
func (h *TaxHandler) CreateVATEntry(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.NewString()
	docDate := stringVal(req, "document_date", time.Now().Format("2006-01-02"))
	base := floatVal(req, "taxable_base", 0)
	vatRate := floatVal(req, "vat_rate", 0.19)
	vatAmt := floatVal(req, "vat_amount", base*vatRate)
	total := floatVal(req, "total_amount", base+vatAmt)

	var periodYear, periodMonth int
	if t, err := time.Parse("2006-01-02", docDate); err == nil {
		periodYear = t.Year()
		periodMonth = int(t.Month())
	} else {
		periodYear = time.Now().Year()
		periodMonth = int(time.Now().Month())
	}

	_, err := h.db.Exec(ctx, `
		INSERT INTO vat_register (
			id, company_id, register_type, period_year, period_month,
			document_date, document_number, partner_name, partner_tax_id,
			taxable_base, vat_rate, vat_amount, total_amount
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		id, companyID,
		stringVal(req, "register_type", "sales"),
		periodYear, periodMonth,
		docDate,
		stringVal(req, "document_number", ""),
		stringVal(req, "partner_name", ""),
		strPtrVal(req, "partner_tax_id"),
		base, vatRate, vatAmt, total,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "VAT entry created"})
}

// ─────────────────────────────────────────────────────────────────────────────
// VAT Returns
// ─────────────────────────────────────────────────────────────────────────────

// ListVATReturns — GET /tax/vat-returns
func (h *TaxHandler) ListVATReturns(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	year := c.Query("year")

	query := `
		SELECT
			r.id, r.period_year, r.period_month, r.status,
			r.sales_base_0, r.sales_base_9, r.sales_base_19,
			r.sales_vat_9, r.sales_vat_19,
			r.total_sales_base, r.total_sales_vat,
			r.purch_base_9, r.purch_base_19,
			r.purch_vat_9, r.purch_vat_19,
			r.total_purch_base, r.total_purch_vat,
			r.credit_bf, r.vat_net_due, r.credit_cf,
			r.notes, r.created_at, r.updated_at
		FROM vat_returns r
		WHERE r.company_id = $1
	`
	args := []interface{}{companyID}
	if year != "" {
		query += " AND r.period_year = $2"
		args = append(args, year)
	}
	query += " ORDER BY r.period_year DESC, r.period_month DESC"

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var (
			id, status, notes                                    string
			periodYear, periodMonth                              int
			sb0, sb9, sb19, sv9, sv19, tsBase, tsVAT            float64
			pb9, pb19, pv9, pv19, tpBase, tpVAT                 float64
			creditBF, vatNetDue, creditCF                        float64
			createdAt, updatedAt                                 time.Time
		)
		err = rows.Scan(
			&id, &periodYear, &periodMonth, &status,
			&sb0, &sb9, &sb19, &sv9, &sv19, &tsBase, &tsVAT,
			&pb9, &pb19, &pv9, &pv19, &tpBase, &tpVAT,
			&creditBF, &vatNetDue, &creditCF,
			&notes, &createdAt, &updatedAt,
		)
		if err != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id": id, "period_year": periodYear, "period_month": periodMonth, "status": status,
			"sales_base_0": sb0, "sales_base_9": sb9, "sales_base_19": sb19,
			"sales_vat_9": sv9, "sales_vat_19": sv19,
			"total_sales_base": tsBase, "total_sales_vat": tsVAT,
			"purch_base_9": pb9, "purch_base_19": pb19,
			"purch_vat_9": pv9, "purch_vat_19": pv19,
			"total_purch_base": tpBase, "total_purch_vat": tpVAT,
			"credit_bf": creditBF, "vat_net_due": vatNetDue, "credit_cf": creditCF,
			"notes": notes, "created_at": createdAt, "updated_at": updatedAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

// CreateVATReturn — POST /tax/vat-returns
func (h *TaxHandler) CreateVATReturn(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.NewString()
	_, err := h.db.Exec(ctx, `
		INSERT INTO vat_returns (
			id, company_id, period_year, period_month,
			sales_base_0, sales_base_9, sales_base_19, sales_vat_9, sales_vat_19,
			purch_base_9, purch_base_19, purch_vat_9, purch_vat_19,
			credit_bf, notes
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		id, companyID,
		intVal(req, "period_year", time.Now().Year()),
		intVal(req, "period_month", int(time.Now().Month())),
		floatVal(req, "sales_base_0", 0),
		floatVal(req, "sales_base_9", 0),
		floatVal(req, "sales_base_19", 0),
		floatVal(req, "sales_vat_9", 0),
		floatVal(req, "sales_vat_19", 0),
		floatVal(req, "purch_base_9", 0),
		floatVal(req, "purch_base_19", 0),
		floatVal(req, "purch_vat_9", 0),
		floatVal(req, "purch_vat_19", 0),
		floatVal(req, "credit_bf", 0),
		stringVal(req, "notes", ""),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "VAT return created"})
}

// UpdateVATReturn — PUT /tax/vat-returns/:id
func (h *TaxHandler) UpdateVATReturn(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(ctx, `
		UPDATE vat_returns SET
			sales_base_0  = COALESCE($3, sales_base_0),
			sales_base_9  = COALESCE($4, sales_base_9),
			sales_base_19 = COALESCE($5, sales_base_19),
			sales_vat_9   = COALESCE($6, sales_vat_9),
			sales_vat_19  = COALESCE($7, sales_vat_19),
			purch_base_9  = COALESCE($8, purch_base_9),
			purch_base_19 = COALESCE($9, purch_base_19),
			purch_vat_9   = COALESCE($10, purch_vat_9),
			purch_vat_19  = COALESCE($11, purch_vat_19),
			credit_bf     = COALESCE($12, credit_bf),
			notes         = COALESCE($13, notes),
			updated_at    = NOW()
		WHERE id = $1 AND company_id = $2`,
		id, companyID,
		floatPtrVal(req, "sales_base_0"),
		floatPtrVal(req, "sales_base_9"),
		floatPtrVal(req, "sales_base_19"),
		floatPtrVal(req, "sales_vat_9"),
		floatPtrVal(req, "sales_vat_19"),
		floatPtrVal(req, "purch_base_9"),
		floatPtrVal(req, "purch_base_19"),
		floatPtrVal(req, "purch_vat_9"),
		floatPtrVal(req, "purch_vat_19"),
		floatPtrVal(req, "credit_bf"),
		strPtrVal(req, "notes"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "VAT return updated"})
}

// SubmitVATReturn — POST /tax/vat-returns/:id/submit
func (h *TaxHandler) SubmitVATReturn(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	_, err := h.db.Exec(ctx, `
		UPDATE vat_returns SET status='submitted', updated_at=NOW()
		WHERE id=$1 AND company_id=$2 AND status='draft'`,
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "VAT return submitted"})
}

// ComputeVATReturn — GET /tax/vat-returns/compute
// Auto-compute VAT return from invoice data for a given period
func (h *TaxHandler) ComputeVATReturn(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	year, month := taxParsePeriod(c)

	type vatBucket struct {
		base9, base19, vat9, vat19 float64
	}
	var salesBucket, purchBucket vatBucket
	var salesBase0 float64

	// Sales breakdown by VAT rate
	salesRows, err := h.db.Query(ctx, `
		SELECT
			COALESCE(tva_rate, 0.19)   AS rate,
			COALESCE(SUM(sub_total), 0) AS base,
			COALESCE(SUM(tva_amount),0) AS vat
		FROM sales_invoices
		WHERE company_id = $1
		  AND EXTRACT(YEAR  FROM date) = $2
		  AND EXTRACT(MONTH FROM date) = $3
		  AND status NOT IN ('cancelled','draft')
		GROUP BY tva_rate
	`, companyID, year, month)
	if err == nil {
		defer salesRows.Close()
		for salesRows.Next() {
			var rate, base, vat float64
			if salesRows.Scan(&rate, &base, &vat) == nil {
				switch {
				case rate < 0.01:
					salesBase0 += base
				case rate < 0.15:
					salesBucket.base9 += base
					salesBucket.vat9 += vat
				default:
					salesBucket.base19 += base
					salesBucket.vat19 += vat
				}
			}
		}
	}

	// Purchase breakdown
	purchRows, err := h.db.Query(ctx, `
		SELECT
			COALESCE(tva_rate, 0.19)    AS rate,
			COALESCE(SUM(sub_total), 0) AS base,
			COALESCE(SUM(tva_amount),0) AS vat
		FROM purchase_invoices
		WHERE company_id = $1
		  AND EXTRACT(YEAR  FROM date) = $2
		  AND EXTRACT(MONTH FROM date) = $3
		  AND status NOT IN ('cancelled','draft')
		GROUP BY tva_rate
	`, companyID, year, month)
	if err == nil {
		defer purchRows.Close()
		for purchRows.Next() {
			var rate, base, vat float64
			if purchRows.Scan(&rate, &base, &vat) == nil {
				if rate < 0.15 {
					purchBucket.base9 += base
					purchBucket.vat9 += vat
				} else {
					purchBucket.base19 += base
					purchBucket.vat19 += vat
				}
			}
		}
	}

	// Credit BF from previous VAT return
	var creditBF float64
	prevM, prevY := month-1, year
	if prevM == 0 {
		prevM = 12
		prevY--
	}
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(credit_cf, 0) FROM vat_returns
		WHERE company_id=$1 AND period_year=$2 AND period_month=$3
		AND status != 'draft'
		ORDER BY created_at DESC LIMIT 1
	`, companyID, prevY, prevM).Scan(&creditBF)

	totalSalesVAT := salesBucket.vat9 + salesBucket.vat19
	totalPurchVAT := purchBucket.vat9 + purchBucket.vat19
	vatNetDue := totalSalesVAT - totalPurchVAT - creditBF
	creditCF := 0.0
	if vatNetDue < 0 {
		creditCF = -vatNetDue
		vatNetDue = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"period_year": year, "period_month": month,
		"sales_base_0": salesBase0,
		"sales_base_9": salesBucket.base9, "sales_base_19": salesBucket.base19,
		"sales_vat_9": salesBucket.vat9, "sales_vat_19": salesBucket.vat19,
		"total_sales_base": salesBase0 + salesBucket.base9 + salesBucket.base19,
		"total_sales_vat": totalSalesVAT,
		"purch_base_9": purchBucket.base9, "purch_base_19": purchBucket.base19,
		"purch_vat_9": purchBucket.vat9, "purch_vat_19": purchBucket.vat19,
		"total_purch_base": purchBucket.base9 + purchBucket.base19,
		"total_purch_vat": totalPurchVAT,
		"credit_bf": creditBF, "vat_net_due": vatNetDue, "credit_cf": creditCF,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Tax Payments
// ─────────────────────────────────────────────────────────────────────────────

// ListTaxPayments — GET /tax/payments
func (h *TaxHandler) ListTaxPayments(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	year := c.Query("year")
	status := c.Query("status")
	dtype := c.Query("type")

	query := `
		SELECT
			tp.id, tp.payment_number, tp.payment_date, tp.due_date,
			tp.declaration_type, tp.period_year, tp.period_month, tp.period_quarter,
			tp.amount_due, tp.amount_paid,
			(tp.amount_due - tp.amount_paid) AS balance,
			tp.status,
			tp.payment_method, tp.reference, tp.receipt_number, tp.notes,
			tp.created_at, tp.updated_at,
			CASE
				WHEN tp.status = 'paid' THEN 'Paid'
				WHEN tp.due_date >= CURRENT_DATE THEN 'Current'
				WHEN CURRENT_DATE - tp.due_date <= 30  THEN '1-30 days'
				WHEN CURRENT_DATE - tp.due_date <= 60  THEN '31-60 days'
				WHEN CURRENT_DATE - tp.due_date <= 90  THEN '61-90 days'
				ELSE 'Over 90 days'
			END AS aging_bucket,
			GREATEST(0, CURRENT_DATE - tp.due_date) AS days_overdue
		FROM tax_payments tp
		WHERE tp.company_id = $1
	`
	args := []interface{}{companyID}
	n := 2
	if year != "" {
		query += fmt.Sprintf(" AND tp.period_year = $%d", n)
		args = append(args, year)
		n++
	}
	if status != "" {
		query += fmt.Sprintf(" AND tp.status = $%d", n)
		args = append(args, status)
		n++
	}
	if dtype != "" {
		query += fmt.Sprintf(" AND tp.declaration_type = $%d", n)
		args = append(args, dtype)
		n++
	}
	query += " ORDER BY tp.due_date DESC, tp.created_at DESC"

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var (
			id, payNum, declType, status, payMethod string
			pmRef, receiptNum, notes, agingBucket   string
			payDate, dueDate                         time.Time
			periodYear, daysOverdue                  int
			periodMonthPtr, periodQtrPtr             *int
			amtDue, amtPaid, balance                 float64
			createdAt, updatedAt                     time.Time
		)
		err = rows.Scan(
			&id, &payNum, &payDate, &dueDate,
			&declType, &periodYear, &periodMonthPtr, &periodQtrPtr,
			&amtDue, &amtPaid, &balance, &status,
			&payMethod, &pmRef, &receiptNum, &notes,
			&createdAt, &updatedAt,
			&agingBucket, &daysOverdue,
		)
		if err != nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id": id, "payment_number": payNum,
			"payment_date": payDate, "due_date": dueDate,
			"declaration_type": declType,
			"period_year": periodYear, "period_month": periodMonthPtr, "period_quarter": periodQtrPtr,
			"amount_due": amtDue, "amount_paid": amtPaid, "balance": balance,
			"status": status, "payment_method": payMethod,
			"reference": pmRef, "receipt_number": receiptNum, "notes": notes,
			"aging_bucket": agingBucket, "days_overdue": daysOverdue,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

// CreateTaxPayment — POST /tax/payments
func (h *TaxHandler) CreateTaxPayment(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.NewString()
	payDate := stringVal(req, "payment_date", time.Now().Format("2006-01-02"))
	dueDate := stringVal(req, "due_date", time.Now().Format("2006-01-02"))
	amtDue := floatVal(req, "amount_due", 0)
	amtPaid := floatVal(req, "amount_paid", 0)
	status := "pending"
	if amtPaid >= amtDue {
		status = "paid"
	} else if amtPaid > 0 {
		status = "partial"
	}

	_, err := h.db.Exec(ctx, `
		INSERT INTO tax_payments (
			id, company_id, declaration_id,
			payment_date, due_date, declaration_type,
			period_year, period_month, period_quarter,
			amount_due, amount_paid, status,
			payment_method, reference, receipt_number, notes,
			bank_account_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		id, companyID,
		strPtrVal(req, "declaration_id"),
		payDate, dueDate,
		stringVal(req, "declaration_type", "g50"),
		intVal(req, "period_year", time.Now().Year()),
		intPtrVal(req, "period_month"),
		intPtrVal(req, "period_quarter"),
		amtDue, amtPaid, status,
		stringVal(req, "payment_method", "bank_transfer"),
		stringVal(req, "reference", ""),
		stringVal(req, "receipt_number", ""),
		stringVal(req, "notes", ""),
		strPtrVal(req, "bank_account_id"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Tax payment created"})
}

// UpdateTaxPayment — PUT /tax/payments/:id
func (h *TaxHandler) UpdateTaxPayment(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Recompute status if amounts changed
	var amtDue, amtPaid float64
	_ = h.db.QueryRow(ctx,
		`SELECT amount_due, amount_paid FROM tax_payments WHERE id=$1 AND company_id=$2`,
		id, companyID,
	).Scan(&amtDue, &amtPaid)

	if v, ok := req["amount_due"]; ok {
		amtDue = toFloat64(v)
	}
	if v, ok := req["amount_paid"]; ok {
		amtPaid = toFloat64(v)
	}
	newStatus := "pending"
	if amtPaid >= amtDue {
		newStatus = "paid"
	} else if amtPaid > 0 {
		newStatus = "partial"
	}

	_, err := h.db.Exec(ctx, `
		UPDATE tax_payments SET
			payment_date     = COALESCE($3::DATE, payment_date),
			due_date         = COALESCE($4::DATE, due_date),
			amount_due       = COALESCE($5, amount_due),
			amount_paid      = COALESCE($6, amount_paid),
			status           = $7,
			payment_method   = COALESCE($8, payment_method),
			reference        = COALESCE($9, reference),
			receipt_number   = COALESCE($10, receipt_number),
			notes            = COALESCE($11, notes),
			updated_at       = NOW()
		WHERE id = $1 AND company_id = $2`,
		id, companyID,
		strPtrVal(req, "payment_date"),
		strPtrVal(req, "due_date"),
		floatPtrVal(req, "amount_due"),
		floatPtrVal(req, "amount_paid"),
		newStatus,
		strPtrVal(req, "payment_method"),
		strPtrVal(req, "reference"),
		strPtrVal(req, "receipt_number"),
		strPtrVal(req, "notes"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tax payment updated"})
}

// DeleteTaxPayment — DELETE /tax/payments/:id
func (h *TaxHandler) DeleteTaxPayment(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	id := c.Param("id")
	ctx := context.Background()

	_, err := h.db.Exec(ctx,
		`DELETE FROM tax_payments WHERE id=$1 AND company_id=$2 AND status='pending'`,
		id, companyID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tax payment deleted"})
}

// ─────────────────────────────────────────────────────────────────────────────
// Tax Reports
// ─────────────────────────────────────────────────────────────────────────────

// GetTaxReport — GET /tax/reports
func (h *TaxHandler) GetTaxReport(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	year, _ := taxParsePeriod(c)

	// Annual totals from declarations
	var totalVATDue, totalTAPDue, totalIBSDue, totalStampDue float64
	var totalIRGWages, totalIRGFees, grandTotalDue, grandTotalPaid float64
	_ = h.db.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(d.tva_net_due), 0),
			COALESCE(SUM(d.tap_net_due), 0),
			COALESCE(SUM(d.ibs_net_due), 0),
			COALESCE(SUM(d.stamp_tax_amount), 0),
			COALESCE(SUM(d.irg_wages_amount), 0),
			COALESCE(SUM(d.irg_fees_amount), 0),
			COALESCE(SUM(d.total_tax_due), 0),
			COALESCE(SUM(tp.amount_paid), 0)
		FROM tax_declarations d
		LEFT JOIN tax_payments tp ON tp.declaration_id = d.id AND tp.status != 'cancelled'
		WHERE d.company_id = $1 AND d.period_year = $2
	`, companyID, year).Scan(
		&totalVATDue, &totalTAPDue, &totalIBSDue, &totalStampDue,
		&totalIRGWages, &totalIRGFees, &grandTotalDue, &grandTotalPaid,
	)

	// Monthly breakdown
	monthRows, err := h.db.Query(ctx, `
		SELECT
			LPAD(d.period_month::TEXT, 2, '0') AS mon,
			COALESCE(SUM(d.tva_net_due), 0)    AS vat_due,
			COALESCE(SUM(d.tap_net_due), 0)    AS tap_due,
			COALESCE(SUM(d.total_tax_due), 0)  AS total_due,
			COALESCE(SUM(tp.amount_paid), 0)   AS total_paid
		FROM tax_declarations d
		LEFT JOIN tax_payments tp ON tp.declaration_id = d.id AND tp.status != 'cancelled'
		WHERE d.company_id = $1
		  AND d.period_year = $2
		  AND d.period_month IS NOT NULL
		GROUP BY d.period_month
		ORDER BY d.period_month
	`, companyID, year)

	var monthly []map[string]interface{}
	if err == nil {
		defer monthRows.Close()
		for monthRows.Next() {
			var mon string
			var vatDue, tapDue, totalDue, totalPaid float64
			if monthRows.Scan(&mon, &vatDue, &tapDue, &totalDue, &totalPaid) == nil {
				monthly = append(monthly, map[string]interface{}{
					"month": taxItoa(year) + "-" + mon,
					"vat_due": vatDue, "tap_due": tapDue,
					"total_due": totalDue, "total_paid": totalPaid,
					"balance": totalDue - totalPaid,
				})
			}
		}
	}
	if monthly == nil {
		monthly = []map[string]interface{}{}
	}

	// Payment status breakdown
	var pendingCount, paidCount, overdueCount int
	var pendingAmt, paidAmt, overdueAmt float64
	_ = h.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'paid'),
			COUNT(*) FILTER (WHERE status IN ('pending','partial') AND due_date < CURRENT_DATE),
			COALESCE(SUM(amount_due - amount_paid) FILTER (WHERE status = 'pending'), 0),
			COALESCE(SUM(amount_paid) FILTER (WHERE status = 'paid'), 0),
			COALESCE(SUM(amount_due - amount_paid) FILTER (WHERE status IN ('pending','partial') AND due_date < CURRENT_DATE), 0)
		FROM tax_payments
		WHERE company_id = $1 AND period_year = $2
	`, companyID, year).Scan(
		&pendingCount, &paidCount, &overdueCount,
		&pendingAmt, &paidAmt, &overdueAmt,
	)

	c.JSON(http.StatusOK, gin.H{
		"period_year": year,
		"summary": gin.H{
			"total_vat_due":    totalVATDue,
			"total_tap_due":    totalTAPDue,
			"total_ibs_due":    totalIBSDue,
			"total_stamp_due":  totalStampDue,
			"total_irg_wages":  totalIRGWages,
			"total_irg_fees":   totalIRGFees,
			"grand_total_due":  grandTotalDue,
			"grand_total_paid": grandTotalPaid,
			"grand_balance":    grandTotalDue - grandTotalPaid,
		},
		"monthly": monthly,
		"payments": gin.H{
			"pending_count": pendingCount, "pending_amount": pendingAmt,
			"paid_count": paidCount, "paid_amount": paidAmt,
			"overdue_count": overdueCount, "overdue_amount": overdueAmt,
		},
	})
}

// GetTaxRateConfig — GET /tax/rates
func (h *TaxHandler) GetTaxRateConfig(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT id, tax_type, rate_name, rate_value, effective_from, effective_to, is_active, notes
		FROM tax_rate_config
		WHERE company_id = $1
		ORDER BY tax_type, rate_name
	`, companyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var list []map[string]interface{}
	for rows.Next() {
		var (
			id, taxType, rateName, notes string
			rateValue                     float64
			effectiveFrom                 time.Time
			effectiveTo                   *time.Time
			isActive                      bool
		)
		if rows.Scan(&id, &taxType, &rateName, &rateValue, &effectiveFrom, &effectiveTo, &isActive, &notes) == nil {
			list = append(list, map[string]interface{}{
				"id": id, "tax_type": taxType, "rate_name": rateName,
				"rate_value": rateValue, "effective_from": effectiveFrom,
				"effective_to": effectiveTo, "is_active": isActive, "notes": notes,
			})
		}
	}
	if list == nil {
		list = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, list)
}

// ─────────────────────────────────────────────────────────────────────────────
// Private helpers (shared with projects.go pattern)
// ─────────────────────────────────────────────────────────────────────────────

func stringVal(m map[string]interface{}, k, def string) string {
	if v, ok := m[k]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func floatVal(m map[string]interface{}, k string, def float64) float64 {
	if v, ok := m[k]; ok && v != nil {
		return toFloat64(v)
	}
	return def
}

func intVal(m map[string]interface{}, k string, def int) int {
	if v, ok := m[k]; ok && v != nil {
		switch x := v.(type) {
		case float64:
			return int(x)
		case int:
			return x
		case int64:
			return int(x)
		}
	}
	return def
}

func toFloat64(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	}
	return 0
}

func floatPtrVal(m map[string]interface{}, k string) *float64 {
	if v, ok := m[k]; ok && v != nil {
		f := toFloat64(v)
		return &f
	}
	return nil
}

func strPtrVal(m map[string]interface{}, k string) *string {
	if v, ok := m[k]; ok && v != nil {
		if s, ok := v.(string); ok && s != "" {
			return &s
		}
	}
	return nil
}

func intPtrVal(m map[string]interface{}, k string) *int {
	if v, ok := m[k]; ok && v != nil {
		i := intVal(m, k, 0)
		return &i
	}
	return nil
}
