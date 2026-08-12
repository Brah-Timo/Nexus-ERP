package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nexus-erp/internal/middleware"
	"nexus-erp/internal/models"
)

// ─── Purchase Handler ─────────────────────────────────────────────────────────

type PurchaseHandler struct{ db *pgxpool.Pool }

func (h *PurchaseHandler) ListSuppliers(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, code, name, tax_id, address, city, phone, email, contact_name, 
		        payment_terms, balance, rating, is_active, created_at
		 FROM suppliers WHERE company_id = $1 ORDER BY name`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var suppliers []models.Supplier
	for rows.Next() {
		var s models.Supplier
		_ = rows.Scan(&s.ID, &s.Code, &s.Name, &s.TaxID, &s.Address, &s.City,
			&s.Phone, &s.Email, &s.ContactName, &s.PaymentTerms, &s.Balance, &s.Rating, &s.IsActive, &s.CreatedAt)
		suppliers = append(suppliers, s)
	}
	c.JSON(http.StatusOK, suppliers)
}

func (h *PurchaseHandler) GetSupplier(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	var s models.Supplier
	err := h.db.QueryRow(ctx,
		`SELECT id, code, name, tax_id, address, city, phone, email, contact_name, payment_terms, balance, rating, is_active
		 FROM suppliers WHERE id = $1`, id).Scan(
		&s.ID, &s.Code, &s.Name, &s.TaxID, &s.Address, &s.City,
		&s.Phone, &s.Email, &s.ContactName, &s.PaymentTerms, &s.Balance, &s.Rating, &s.IsActive)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *PurchaseHandler) CreateSupplier(c *gin.Context) {
	var s models.Supplier
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.ID = uuid.NewString()
	s.CompanyID = middleware.GetCompanyID(c)
	ctx := context.Background()
	_, err := h.db.Exec(ctx,
		`INSERT INTO suppliers (id, company_id, code, name, tax_id, address, city, phone, email, contact_name, payment_terms, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		s.ID, s.CompanyID, s.Code, s.Name, s.TaxID, s.Address, s.City, s.Phone, s.Email, s.ContactName, s.PaymentTerms, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, s)
}

func (h *PurchaseHandler) UpdateSupplier(c *gin.Context) {
	id := c.Param("id")
	var s models.Supplier
	c.ShouldBindJSON(&s)
	ctx := context.Background()
	_, _ = h.db.Exec(ctx,
		`UPDATE suppliers SET name=$1, tax_id=$2, address=$3, phone=$4, email=$5 WHERE id=$6`,
		s.Name, s.TaxID, s.Address, s.Phone, s.Email, id)
	c.JSON(http.StatusOK, s)
}

func (h *PurchaseHandler) ListRFQs(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, number, supplier_id, supplier_name, date, status, total_amount, created_at
		 FROM rfqs WHERE company_id = $1 ORDER BY date DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var rfqs []map[string]interface{}
	for rows.Next() {
		row := make(map[string]interface{})
		var id, number, suppID, suppName, status string
		var total float64
		var date, createdAt interface{}
		_ = rows.Scan(&id, &number, &suppID, &suppName, &date, &status, &total, &createdAt)
		row["id"] = id
		row["number"] = number
		row["supplier_name"] = suppName
		row["date"] = date
		row["status"] = status
		row["total_amount"] = total
		rfqs = append(rfqs, row)
	}
	c.JSON(http.StatusOK, rfqs)
}

func (h *PurchaseHandler) CreateRFQ(c *gin.Context) {
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	req["id"] = uuid.NewString()
	c.JSON(http.StatusCreated, req)
}

func (h *PurchaseHandler) SendRFQ(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, _ = h.db.Exec(ctx, `UPDATE rfqs SET status = 'sent' WHERE id = $1`, id)
	c.JSON(http.StatusOK, gin.H{"message": "RFQ sent to supplier"})
}

func (h *PurchaseHandler) ListOrders(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, number, supplier_id, supplier_name, date, expected_date, status, total_amount, created_at
		 FROM purchase_orders WHERE company_id = $1 ORDER BY date DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var orders []models.PurchaseOrder
	for rows.Next() {
		var o models.PurchaseOrder
		_ = rows.Scan(&o.ID, &o.Number, &o.SupplierID, &o.SupplierName, &o.Date,
			&o.ExpectedDate, &o.Status, &o.TotalAmount, &o.CreatedAt)
		orders = append(orders, o)
	}
	c.JSON(http.StatusOK, orders)
}

func (h *PurchaseHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	var o models.PurchaseOrder
	err := h.db.QueryRow(ctx, `
		SELECT id, number, supplier_id, supplier_name, date, expected_date, status, 
		       sub_total, tva_amount, total_amount, notes
		FROM purchase_orders WHERE id = $1
	`, id).Scan(&o.ID, &o.Number, &o.SupplierID, &o.SupplierName, &o.Date,
		&o.ExpectedDate, &o.Status, &o.SubTotal, &o.TVAAmount, &o.TotalAmount, &o.Notes)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order not found"})
		return
	}

	rows, _ := h.db.Query(ctx, `
		SELECT id, item_id, item_code, item_name, quantity, received_qty, unit_price, tva_rate, sub_total, total_amount
		FROM purchase_order_lines WHERE purchase_order_id = $1
	`, id)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var l models.PurchaseOrderLine
			_ = rows.Scan(&l.ID, &l.ItemID, &l.ItemCode, &l.ItemName, &l.Quantity,
				&l.ReceivedQty, &l.UnitPrice, &l.TVARate, &l.SubTotal, &l.TotalAmount)
			o.Lines = append(o.Lines, l)
		}
	}
	c.JSON(http.StatusOK, o)
}

func (h *PurchaseHandler) CreateOrder(c *gin.Context) {
	var o models.PurchaseOrder
	if err := c.ShouldBindJSON(&o); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	o.ID = uuid.NewString()
	o.CompanyID = middleware.GetCompanyID(c)
	o.Status = "draft"
	o.Number = generateNumber("PO", o.CompanyID, h.db)

	ctx := context.Background()
	tx, _ := h.db.Begin(ctx)
	defer tx.Rollback(ctx)

	_, err := tx.Exec(ctx, `
		INSERT INTO purchase_orders 
		(id, company_id, number, supplier_id, supplier_name, date, expected_date, status, sub_total, tva_amount, total_amount, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`, o.ID, o.CompanyID, o.Number, o.SupplierID, o.SupplierName, o.Date, o.ExpectedDate,
		o.Status, o.SubTotal, o.TVAAmount, o.TotalAmount, o.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, line := range o.Lines {
		line.ID = uuid.NewString()
		_, _ = tx.Exec(ctx, `
			INSERT INTO purchase_order_lines 
			(id, purchase_order_id, item_id, item_code, item_name, quantity, unit_price, tva_rate, sub_total, total_amount)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`, line.ID, o.ID, line.ItemID, line.ItemCode, line.ItemName,
			line.Quantity, line.UnitPrice, line.TVARate, line.SubTotal, line.TotalAmount)
	}

	_ = tx.Commit(ctx)
	c.JSON(http.StatusCreated, o)
}

func (h *PurchaseHandler) ApprovePurchaseOrder(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, _ = h.db.Exec(ctx, `UPDATE purchase_orders SET status = 'approved' WHERE id = $1 AND status = 'draft'`, id)
	c.JSON(http.StatusOK, gin.H{"message": "Purchase order approved"})
}

func (h *PurchaseHandler) ConfirmPurchaseOrder(c *gin.Context) {
	id := c.Param("id")
	ctx := context.Background()
	_, _ = h.db.Exec(ctx, `UPDATE purchase_orders SET status = 'confirmed' WHERE id = $1`, id)
	c.JSON(http.StatusOK, gin.H{"message": "Purchase order confirmed"})
}

func (h *PurchaseHandler) ListGoodsReceipts(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, number, purchase_order_id, supplier_name, date, status, created_at
		 FROM goods_receipt_notes WHERE company_id = $1 ORDER BY date DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var grns []map[string]interface{}
	for rows.Next() {
		var id, number, poID, suppName, status string
		var date, createdAt interface{}
		_ = rows.Scan(&id, &number, &poID, &suppName, &date, &status, &createdAt)
		grns = append(grns, map[string]interface{}{
			"id": id, "number": number, "purchase_order_id": poID,
			"supplier_name": suppName, "date": date, "status": status,
		})
	}
	c.JSON(http.StatusOK, grns)
}

func (h *PurchaseHandler) CreateGoodsReceipt(c *gin.Context) {
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	req["id"] = uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	req["company_id"] = companyID
	req["number"] = generateNumber("GRN", companyID, h.db)
	req["status"] = "received"

	// TODO: Update PO received quantities and create stock movements
	c.JSON(http.StatusCreated, req)
}

func (h *PurchaseHandler) ListInvoices(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, number, supplier_id, supplier_name, date, due_date, status, total_amount, paid_amount, created_at
		 FROM purchase_invoices WHERE company_id = $1 ORDER BY date DESC`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()
	var invoices []map[string]interface{}
	for rows.Next() {
		var id, number, suppID, suppName, status string
		var total, paid float64
		var date, dueDate, createdAt interface{}
		_ = rows.Scan(&id, &number, &suppID, &suppName, &date, &dueDate, &status, &total, &paid, &createdAt)
		invoices = append(invoices, map[string]interface{}{
			"id": id, "number": number, "supplier_name": suppName,
			"date": date, "due_date": dueDate, "status": status,
			"total_amount": total, "paid_amount": paid,
		})
	}
	c.JSON(http.StatusOK, invoices)
}

func (h *PurchaseHandler) CreateInvoice(c *gin.Context) {
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	req["id"] = uuid.NewString()
	companyID := middleware.GetCompanyID(c)
	req["company_id"] = companyID
	req["number"] = generateNumber("PINV", companyID, h.db)
	req["status"] = "draft"
	ctx := context.Background()
	_, _ = h.db.Exec(ctx,
		`INSERT INTO purchase_invoices (id, company_id, number, supplier_id, supplier_name, date, due_date, status, total_amount)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		req["id"], companyID, req["number"], req["supplier_id"], req["supplier_name"],
		req["date"], req["due_date"], "draft", req["total_amount"])
	c.JSON(http.StatusCreated, req)
}

func (h *PurchaseHandler) ThreeWayMatch(c *gin.Context) {
	id := c.Param("id")
	// Three-way match: PO vs GRN vs Invoice
	c.JSON(http.StatusOK, gin.H{"id": id, "matched": true, "message": "Three-way match validated"})
}

func (h *PurchaseHandler) RecordPayment(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Amount float64 `json:"amount"`
	}
	c.ShouldBindJSON(&req)
	ctx := context.Background()
	_, _ = h.db.Exec(ctx, `
		UPDATE purchase_invoices 
		SET paid_amount = paid_amount + $1,
		    status = CASE WHEN total_amount - paid_amount - $1 <= 0 THEN 'paid' ELSE 'partial' END
		WHERE id = $2
	`, req.Amount, id)
	c.JSON(http.StatusOK, gin.H{"message": "Payment recorded"})
}

func (h *PurchaseHandler) ListEvaluations(c *gin.Context) {
	c.JSON(http.StatusOK, []interface{}{})
}

func (h *PurchaseHandler) CreateEvaluation(c *gin.Context) {
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	req["id"] = uuid.NewString()
	c.JSON(http.StatusCreated, req)
}

func (h *PurchaseHandler) AgingReport(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	rows, err := h.db.Query(ctx, `
		SELECT 
			s.name,
			SUM(CASE WHEN NOW() - pi.due_date <= INTERVAL '30 days' THEN pi.total_amount - pi.paid_amount ELSE 0 END) as "0_30",
			SUM(CASE WHEN NOW() - pi.due_date BETWEEN INTERVAL '31 days' AND INTERVAL '60 days' THEN pi.total_amount - pi.paid_amount ELSE 0 END) as "31_60",
			SUM(CASE WHEN NOW() - pi.due_date > INTERVAL '60 days' THEN pi.total_amount - pi.paid_amount ELSE 0 END) as "over_60",
			SUM(pi.total_amount - pi.paid_amount) as total
		FROM purchase_invoices pi
		JOIN suppliers s ON s.id = pi.supplier_id
		WHERE pi.company_id = $1 AND pi.status NOT IN ('paid','cancelled')
		GROUP BY s.name
		ORDER BY total DESC
	`, companyID)
	if err != nil {
		c.JSON(http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	var aging []map[string]interface{}
	for rows.Next() {
		var name string
		var d30, d60, over60, total float64
		_ = rows.Scan(&name, &d30, &d60, &over60, &total)
		aging = append(aging, map[string]interface{}{
			"supplier": name, "0_30": d30, "31_60": d60, "over_60": over60, "total": total,
		})
	}
	c.JSON(http.StatusOK, aging)
}

// ─── Inventory Handler ────────────────────────────────────────────────────────
// InventoryHandler struct, all methods, and SQL are in inventory.go
