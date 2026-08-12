package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"nexus-erp/internal/middleware"
)

// ─── Handler ─────────────────────────────────────────────────────────────────

type ReportsBIHandler struct{ db *pgxpool.Pool }

// ─── helpers ─────────────────────────────────────────────────────────────────

func biYear(c *gin.Context) int {
	if v, err := strconv.Atoi(c.Query("year")); err == nil && v > 2000 {
		return v
	}
	return time.Now().Year()
}

func biF(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func biI(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func scanF(db *pgxpool.Pool, ctx context.Context, q string, args ...interface{}) float64 {
	var v float64
	_ = db.QueryRow(ctx, q, args...).Scan(&v)
	return v
}

func scanI(db *pgxpool.Pool, ctx context.Context, q string, args ...interface{}) int {
	var v int
	_ = db.QueryRow(ctx, q, args...).Scan(&v)
	return v
}

// ─────────────────────────────────────────────────────────────────────────────
// BI DASHBOARD — GET /reports/bi-dashboard
// ─────────────────────────────────────────────────────────────────────────────

func (h *ReportsBIHandler) BIDashboard(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	ctx := context.Background()
	year := biYear(c)

	// Revenue (confirmed sales invoices)
	revenue := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount),0) FROM sales_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year)

	// Expenses (purchase invoices)
	expenses := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount),0) FROM purchase_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year)

	// Net profit (revenue - expenses)
	netProfit := revenue - expenses

	// Accounts Receivable (unpaid sales invoices)
	ar := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount - paid_amount),0) FROM sales_invoices
		 WHERE company_id=$1 AND status IN ('confirmed','partial')`, cid)

	// Accounts Payable (unpaid purchase invoices)
	ap := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount - paid_amount),0) FROM purchase_invoices
		 WHERE company_id=$1 AND status IN ('confirmed','partial')`, cid)

	// Cash position
	cashBalance := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(current_balance),0) FROM cash_accounts WHERE company_id=$1 AND is_active=true`, cid)
	bankBalance := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(current_balance),0) FROM bank_accounts WHERE company_id=$1 AND is_active=true`, cid)

	// Active orders
	activeOrders := scanI(h.db, ctx,
		`SELECT COUNT(*) FROM sales_orders WHERE company_id=$1 AND status IN ('confirmed','processing')`, cid)

	// Inventory value
	inventoryValue := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(sl.quantity * COALESCE(i.cost_price, i.sale_price, 0)),0)
		 FROM stock_levels sl JOIN items i ON i.id=sl.item_id
		 WHERE i.company_id=$1`, cid)

	// Overdue invoices count
	overdueCount := scanI(h.db, ctx,
		`SELECT COUNT(*) FROM sales_invoices
		 WHERE company_id=$1 AND due_date < CURRENT_DATE
		   AND status IN ('confirmed','partial')`, cid)

	// Monthly revenue trend (12 months)
	rows, _ := h.db.Query(ctx,
		`SELECT EXTRACT(MONTH FROM date)::int AS mo, COALESCE(SUM(total_amount),0)
		 FROM sales_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')
		 GROUP BY mo ORDER BY mo`, cid, year)
	monthlyRevenue := make([]map[string]interface{}, 12)
	for i := 0; i < 12; i++ {
		monthlyRevenue[i] = map[string]interface{}{"month": i + 1, "revenue": 0.0, "expenses": 0.0}
	}
	if rows != nil {
		for rows.Next() {
			var mo int
			var amt float64
			_ = rows.Scan(&mo, &amt)
			if mo >= 1 && mo <= 12 {
				monthlyRevenue[mo-1]["revenue"] = amt
			}
		}
		rows.Close()
	}

	// Monthly expenses trend
	rows2, _ := h.db.Query(ctx,
		`SELECT EXTRACT(MONTH FROM date)::int AS mo, COALESCE(SUM(total_amount),0)
		 FROM purchase_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')
		 GROUP BY mo ORDER BY mo`, cid, year)
	if rows2 != nil {
		for rows2.Next() {
			var mo int
			var amt float64
			_ = rows2.Scan(&mo, &amt)
			if mo >= 1 && mo <= 12 {
				monthlyRevenue[mo-1]["expenses"] = amt
			}
		}
		rows2.Close()
	}

	// Top 5 customers by revenue
	custRows, _ := h.db.Query(ctx,
		`SELECT COALESCE(c.name,'Unknown'), COALESCE(SUM(si.total_amount),0) AS rev
		 FROM sales_invoices si
		 LEFT JOIN customers c ON c.id=si.customer_id
		 WHERE si.company_id=$1 AND EXTRACT(YEAR FROM si.date)=$2
		   AND si.status NOT IN ('cancelled','draft')
		 GROUP BY c.name ORDER BY rev DESC LIMIT 5`, cid, year)
	topCustomers := []map[string]interface{}{}
	if custRows != nil {
		for custRows.Next() {
			var name string
			var rev float64
			_ = custRows.Scan(&name, &rev)
			topCustomers = append(topCustomers, map[string]interface{}{"name": name, "revenue": rev})
		}
		custRows.Close()
	}

	// Top 5 products by revenue
	prodRows, _ := h.db.Query(ctx,
		`SELECT COALESCE(il.description,'Unknown'), COALESCE(SUM(il.total_price),0) AS rev
		 FROM invoice_lines il
		 JOIN sales_invoices si ON si.id=il.invoice_id
		 WHERE si.company_id=$1 AND EXTRACT(YEAR FROM si.date)=$2
		   AND si.status NOT IN ('cancelled','draft')
		 GROUP BY il.description ORDER BY rev DESC LIMIT 5`, cid, year)
	topProducts := []map[string]interface{}{}
	if prodRows != nil {
		for prodRows.Next() {
			var name string
			var rev float64
			_ = prodRows.Scan(&name, &rev)
			topProducts = append(topProducts, map[string]interface{}{"name": name, "revenue": rev})
		}
		prodRows.Close()
	}

	c.JSON(http.StatusOK, gin.H{
		"year":             year,
		"revenue":          revenue,
		"expenses":         expenses,
		"net_profit":       netProfit,
		"profit_margin":    func() float64 { if revenue > 0 { return netProfit / revenue * 100 }; return 0 }(),
		"accounts_receivable": ar,
		"accounts_payable":    ap,
		"cash_balance":        cashBalance,
		"bank_balance":        bankBalance,
		"total_cash":          cashBalance + bankBalance,
		"inventory_value":     inventoryValue,
		"active_orders":       activeOrders,
		"overdue_invoices":    overdueCount,
		"monthly_trend":       monthlyRevenue,
		"top_customers":       topCustomers,
		"top_products":        topProducts,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// FINANCIAL REPORTS — GET /reports/financial
// ─────────────────────────────────────────────────────────────────────────────

func (h *ReportsBIHandler) FinancialReports(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	ctx := context.Background()
	year := biYear(c)

	// Income Statement
	revenue := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(sub_total),0) FROM sales_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year)
	tvaCollected := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(tva_amount),0) FROM sales_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year)
	cogs := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(pi.total_amount),0) FROM purchase_invoices pi
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year)
	grossProfit := revenue - cogs
	gpMargin := 0.0
	if revenue > 0 {
		gpMargin = grossProfit / revenue * 100
	}

	// Operating expenses (project expenses + payroll)
	opex := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(gross_salary),0) FROM payroll_slips ps
		 JOIN payroll_runs pr ON pr.id=ps.payroll_run_id
		 WHERE pr.company_id=$1 AND EXTRACT(YEAR FROM pr.pay_date)=$2`, cid, year)
	ebitda := grossProfit - opex
	netProfit := ebitda

	// Balance Sheet items
	totalAssets := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(current_balance),0) FROM cash_accounts WHERE company_id=$1 AND is_active=true`, cid) +
		scanF(h.db, ctx, `SELECT COALESCE(SUM(current_balance),0) FROM bank_accounts WHERE company_id=$1 AND is_active=true`, cid) +
		scanF(h.db, ctx, `SELECT COALESCE(SUM(total_amount-paid_amount),0) FROM sales_invoices WHERE company_id=$1 AND status IN ('confirmed','partial')`, cid)

	totalLiabilities := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount-paid_amount),0) FROM purchase_invoices WHERE company_id=$1 AND status IN ('confirmed','partial')`, cid)

	equity := totalAssets - totalLiabilities

	// Financial Ratios
	currentRatio := 0.0
	if totalLiabilities > 0 {
		currentRatio = totalAssets / totalLiabilities
	}
	debtRatio := 0.0
	if totalAssets > 0 {
		debtRatio = totalLiabilities / totalAssets * 100
	}
	roe := 0.0
	if equity > 0 {
		roe = netProfit / equity * 100
	}
	roa := 0.0
	if totalAssets > 0 {
		roa = netProfit / totalAssets * 100
	}

	// Quarterly breakdown
	quarters := []map[string]interface{}{}
	for q := 1; q <= 4; q++ {
		startM := (q-1)*3 + 1
		endM := q * 3
		qRev := scanF(h.db, ctx,
			`SELECT COALESCE(SUM(sub_total),0) FROM sales_invoices
			 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
			   AND EXTRACT(MONTH FROM date) BETWEEN $3 AND $4
			   AND status NOT IN ('cancelled','draft')`, cid, year, startM, endM)
		qCost := scanF(h.db, ctx,
			`SELECT COALESCE(SUM(total_amount),0) FROM purchase_invoices
			 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
			   AND EXTRACT(MONTH FROM date) BETWEEN $3 AND $4
			   AND status NOT IN ('cancelled','draft')`, cid, year, startM, endM)
		qGP := qRev - qCost
		qMargin := 0.0
		if qRev > 0 {
			qMargin = qGP / qRev * 100
		}
		quarters = append(quarters, map[string]interface{}{
			"quarter": q, "revenue": qRev, "cost": qCost,
			"gross_profit": qGP, "margin_pct": qMargin,
		})
	}

	// Monthly cash-flow
	cashFlowRows, _ := h.db.Query(ctx,
		`SELECT m, inflow, outflow FROM (
		   SELECT EXTRACT(MONTH FROM date)::int AS m, COALESCE(SUM(total_amount),0) AS inflow, 0 AS outflow
		   FROM sales_invoices
		   WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2 AND status NOT IN ('cancelled','draft')
		   GROUP BY m
		   UNION ALL
		   SELECT EXTRACT(MONTH FROM date)::int, 0, COALESCE(SUM(total_amount),0)
		   FROM purchase_invoices
		   WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2 AND status NOT IN ('cancelled','draft')
		   GROUP BY m
		) t GROUP BY m, inflow, outflow ORDER BY m`, cid, year)
	cashFlowMap := map[int]map[string]float64{}
	for i := 1; i <= 12; i++ {
		cashFlowMap[i] = map[string]float64{"inflow": 0, "outflow": 0}
	}
	if cashFlowRows != nil {
		for cashFlowRows.Next() {
			var mo int
			var inf, out float64
			_ = cashFlowRows.Scan(&mo, &inf, &out)
			if mo >= 1 && mo <= 12 {
				cashFlowMap[mo]["inflow"] += inf
				cashFlowMap[mo]["outflow"] += out
			}
		}
		cashFlowRows.Close()
	}
	cashFlow := []map[string]interface{}{}
	for i := 1; i <= 12; i++ {
		net := cashFlowMap[i]["inflow"] - cashFlowMap[i]["outflow"]
		cashFlow = append(cashFlow, map[string]interface{}{
			"month": i, "inflow": cashFlowMap[i]["inflow"],
			"outflow": cashFlowMap[i]["outflow"], "net": net,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"year":              year,
		"revenue":           revenue,
		"tva_collected":     tvaCollected,
		"cogs":              cogs,
		"gross_profit":      grossProfit,
		"gross_margin_pct":  gpMargin,
		"opex":              opex,
		"ebitda":            ebitda,
		"net_profit":        netProfit,
		"net_margin_pct":    func() float64 { if revenue > 0 { return netProfit / revenue * 100 }; return 0 }(),
		"total_assets":      totalAssets,
		"total_liabilities": totalLiabilities,
		"equity":            equity,
		"current_ratio":     currentRatio,
		"debt_ratio":        debtRatio,
		"roe":               roe,
		"roa":               roa,
		"quarters":          quarters,
		"cash_flow":         cashFlow,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// SALES REPORTS — GET /reports/sales
// ─────────────────────────────────────────────────────────────────────────────

func (h *ReportsBIHandler) SalesReports(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	ctx := context.Background()
	year := biYear(c)

	totalRevenue := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount),0) FROM sales_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year)
	totalOrders := scanI(h.db, ctx,
		`SELECT COUNT(*) FROM sales_orders WHERE company_id=$1 AND EXTRACT(YEAR FROM created_at)=$2`, cid, year)
	invoiceCount := scanI(h.db, ctx,
		`SELECT COUNT(*) FROM sales_invoices WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year)
	avgOrderValue := 0.0
	if invoiceCount > 0 {
		avgOrderValue = totalRevenue / float64(invoiceCount)
	}
	totalAR := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount-paid_amount),0) FROM sales_invoices
		 WHERE company_id=$1 AND status IN ('confirmed','partial')`, cid)
	overdueAR := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount-paid_amount),0) FROM sales_invoices
		 WHERE company_id=$1 AND status IN ('confirmed','partial') AND due_date < CURRENT_DATE`, cid)

	// Monthly sales
	monthlyRows, _ := h.db.Query(ctx,
		`SELECT EXTRACT(MONTH FROM date)::int, COALESCE(SUM(total_amount),0), COUNT(*)
		 FROM sales_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')
		 GROUP BY 1 ORDER BY 1`, cid, year)
	monthly := make([]map[string]interface{}, 12)
	for i := 0; i < 12; i++ {
		monthly[i] = map[string]interface{}{"month": i + 1, "revenue": 0.0, "count": 0}
	}
	if monthlyRows != nil {
		for monthlyRows.Next() {
			var mo, cnt int
			var amt float64
			_ = monthlyRows.Scan(&mo, &amt, &cnt)
			if mo >= 1 && mo <= 12 {
				monthly[mo-1] = map[string]interface{}{"month": mo, "revenue": amt, "count": cnt}
			}
		}
		monthlyRows.Close()
	}

	// Top customers
	topCustRows, _ := h.db.Query(ctx,
		`SELECT COALESCE(c.name,'Unknown'), COALESCE(SUM(si.total_amount),0),
		        COUNT(si.id), COALESCE(SUM(si.total_amount - si.paid_amount),0)
		 FROM sales_invoices si
		 LEFT JOIN customers c ON c.id=si.customer_id
		 WHERE si.company_id=$1 AND EXTRACT(YEAR FROM si.date)=$2
		   AND si.status NOT IN ('cancelled','draft')
		 GROUP BY c.name ORDER BY 2 DESC LIMIT 10`, cid, year)
	topCustomers := []map[string]interface{}{}
	if topCustRows != nil {
		for topCustRows.Next() {
			var name string
			var rev, ar float64
			var cnt int
			_ = topCustRows.Scan(&name, &rev, &cnt, &ar)
			topCustomers = append(topCustomers, map[string]interface{}{
				"name": name, "revenue": rev, "invoice_count": cnt, "outstanding": ar,
			})
		}
		topCustRows.Close()
	}

	// Top products
	topProdRows, _ := h.db.Query(ctx,
		`SELECT il.description, COALESCE(SUM(il.quantity),0), COALESCE(SUM(il.total_price),0)
		 FROM invoice_lines il
		 JOIN sales_invoices si ON si.id=il.invoice_id
		 WHERE si.company_id=$1 AND EXTRACT(YEAR FROM si.date)=$2
		   AND si.status NOT IN ('cancelled','draft')
		 GROUP BY il.description ORDER BY 3 DESC LIMIT 10`, cid, year)
	topProducts := []map[string]interface{}{}
	if topProdRows != nil {
		for topProdRows.Next() {
			var name string
			var qty, rev float64
			_ = topProdRows.Scan(&name, &qty, &rev)
			topProducts = append(topProducts, map[string]interface{}{
				"name": name, "quantity": qty, "revenue": rev,
			})
		}
		topProdRows.Close()
	}

	// Sales by status
	statusRows, _ := h.db.Query(ctx,
		`SELECT status, COUNT(*), COALESCE(SUM(total_amount),0)
		 FROM sales_invoices WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		 GROUP BY status ORDER BY 3 DESC`, cid, year)
	byStatus := []map[string]interface{}{}
	if statusRows != nil {
		for statusRows.Next() {
			var st string
			var cnt int
			var amt float64
			_ = statusRows.Scan(&st, &cnt, &amt)
			byStatus = append(byStatus, map[string]interface{}{"status": st, "count": cnt, "amount": amt})
		}
		statusRows.Close()
	}

	// Pipeline summary (opportunities)
	pipelineTotal := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(amount),0) FROM sales_opportunities WHERE company_id=$1 AND status NOT IN ('lost','cancelled')`, cid)
	pipelineCount := scanI(h.db, ctx,
		`SELECT COUNT(*) FROM sales_opportunities WHERE company_id=$1 AND status NOT IN ('lost','cancelled')`, cid)

	c.JSON(http.StatusOK, gin.H{
		"year":            year,
		"total_revenue":   totalRevenue,
		"total_orders":    totalOrders,
		"invoice_count":   invoiceCount,
		"avg_order_value": avgOrderValue,
		"total_ar":        totalAR,
		"overdue_ar":      overdueAR,
		"collection_rate": func() float64 {
			total := totalRevenue
			if total > 0 {
				return (total - totalAR) / total * 100
			}
			return 0
		}(),
		"pipeline_total":  pipelineTotal,
		"pipeline_count":  pipelineCount,
		"monthly":         monthly,
		"top_customers":   topCustomers,
		"top_products":    topProducts,
		"by_status":       byStatus,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// PURCHASE REPORTS — GET /reports/purchase
// ─────────────────────────────────────────────────────────────────────────────

func (h *ReportsBIHandler) PurchaseReports(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	ctx := context.Background()
	year := biYear(c)

	totalSpend := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount),0) FROM purchase_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year)
	orderCount := scanI(h.db, ctx,
		`SELECT COUNT(*) FROM purchase_orders WHERE company_id=$1 AND EXTRACT(YEAR FROM created_at)=$2`, cid, year)
	supplierCount := scanI(h.db, ctx,
		`SELECT COUNT(DISTINCT supplier_id) FROM purchase_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2`, cid, year)
	totalAP := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount-paid_amount),0) FROM purchase_invoices
		 WHERE company_id=$1 AND status IN ('confirmed','partial')`, cid)
	overdueAP := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount-paid_amount),0) FROM purchase_invoices
		 WHERE company_id=$1 AND status IN ('confirmed','partial') AND due_date < CURRENT_DATE`, cid)

	// Monthly spend
	monthlyRows, _ := h.db.Query(ctx,
		`SELECT EXTRACT(MONTH FROM date)::int, COALESCE(SUM(total_amount),0), COUNT(*)
		 FROM purchase_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')
		 GROUP BY 1 ORDER BY 1`, cid, year)
	monthly := make([]map[string]interface{}, 12)
	for i := 0; i < 12; i++ {
		monthly[i] = map[string]interface{}{"month": i + 1, "spend": 0.0, "count": 0}
	}
	if monthlyRows != nil {
		for monthlyRows.Next() {
			var mo, cnt int
			var amt float64
			_ = monthlyRows.Scan(&mo, &amt, &cnt)
			if mo >= 1 && mo <= 12 {
				monthly[mo-1] = map[string]interface{}{"month": mo, "spend": amt, "count": cnt}
			}
		}
		monthlyRows.Close()
	}

	// Top suppliers
	topSupRows, _ := h.db.Query(ctx,
		`SELECT COALESCE(s.name,'Unknown'), COALESCE(SUM(pi.total_amount),0),
		        COUNT(pi.id), COALESCE(SUM(pi.total_amount-pi.paid_amount),0)
		 FROM purchase_invoices pi
		 LEFT JOIN suppliers s ON s.id=pi.supplier_id
		 WHERE pi.company_id=$1 AND EXTRACT(YEAR FROM pi.date)=$2
		   AND pi.status NOT IN ('cancelled','draft')
		 GROUP BY s.name ORDER BY 2 DESC LIMIT 10`, cid, year)
	topSuppliers := []map[string]interface{}{}
	if topSupRows != nil {
		for topSupRows.Next() {
			var name string
			var spend, ap float64
			var cnt int
			_ = topSupRows.Scan(&name, &spend, &cnt, &ap)
			topSuppliers = append(topSuppliers, map[string]interface{}{
				"name": name, "spend": spend, "invoice_count": cnt, "outstanding": ap,
			})
		}
		topSupRows.Close()
	}

	// Top purchased items
	topItemRows, _ := h.db.Query(ctx,
		`SELECT pil.description, COALESCE(SUM(pil.quantity),0), COALESCE(SUM(pil.total_price),0)
		 FROM purchase_invoice_lines pil
		 JOIN purchase_invoices pi ON pi.id=pil.invoice_id
		 WHERE pi.company_id=$1 AND EXTRACT(YEAR FROM pi.date)=$2
		   AND pi.status NOT IN ('cancelled','draft')
		 GROUP BY pil.description ORDER BY 3 DESC LIMIT 10`, cid, year)
	topItems := []map[string]interface{}{}
	if topItemRows != nil {
		for topItemRows.Next() {
			var name string
			var qty, spend float64
			_ = topItemRows.Scan(&name, &qty, &spend)
			topItems = append(topItems, map[string]interface{}{
				"name": name, "quantity": qty, "spend": spend,
			})
		}
		topItemRows.Close()
	}

	// Order status breakdown
	statusRows, _ := h.db.Query(ctx,
		`SELECT status, COUNT(*), COALESCE(SUM(total_amount),0)
		 FROM purchase_orders WHERE company_id=$1 AND EXTRACT(YEAR FROM created_at)=$2
		 GROUP BY status ORDER BY 3 DESC`, cid, year)
	byStatus := []map[string]interface{}{}
	if statusRows != nil {
		for statusRows.Next() {
			var st string
			var cnt int
			var amt float64
			_ = statusRows.Scan(&st, &cnt, &amt)
			byStatus = append(byStatus, map[string]interface{}{"status": st, "count": cnt, "amount": amt})
		}
		statusRows.Close()
	}

	c.JSON(http.StatusOK, gin.H{
		"year":            year,
		"total_spend":     totalSpend,
		"order_count":     orderCount,
		"supplier_count":  supplierCount,
		"total_ap":        totalAP,
		"overdue_ap":      overdueAP,
		"monthly":         monthly,
		"top_suppliers":   topSuppliers,
		"top_items":       topItems,
		"by_status":       byStatus,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// INVENTORY REPORTS — GET /reports/inventory
// ─────────────────────────────────────────────────────────────────────────────

func (h *ReportsBIHandler) InventoryReports(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	ctx := context.Background()

	totalItems := scanI(h.db, ctx,
		`SELECT COUNT(*) FROM items WHERE company_id=$1 AND is_active=true`, cid)
	totalValue := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(sl.quantity * COALESCE(i.cost_price, i.sale_price, 0)),0)
		 FROM stock_levels sl JOIN items i ON i.id=sl.item_id WHERE i.company_id=$1`, cid)
	lowStockCount := scanI(h.db, ctx,
		`SELECT COUNT(*) FROM (
		   SELECT i.id FROM items i
		   JOIN stock_levels sl ON sl.item_id=i.id
		   WHERE i.company_id=$1 AND i.is_active=true
		   GROUP BY i.id, i.reorder_point
		   HAVING SUM(sl.quantity) <= i.reorder_point AND i.reorder_point > 0
		 ) t`, cid)
	outOfStockCount := scanI(h.db, ctx,
		`SELECT COUNT(*) FROM (
		   SELECT i.id FROM items i
		   LEFT JOIN stock_levels sl ON sl.item_id=i.id
		   WHERE i.company_id=$1 AND i.is_active=true
		   GROUP BY i.id
		   HAVING COALESCE(SUM(sl.quantity),0) = 0
		 ) t`, cid)
	warehouseCount := scanI(h.db, ctx,
		`SELECT COUNT(*) FROM warehouses WHERE company_id=$1 AND is_active=true`, cid)
	movementCount := scanI(h.db, ctx,
		`SELECT COUNT(*) FROM inventory_movements im
		 JOIN items i ON i.id=im.item_id
		 WHERE i.company_id=$1 AND im.created_at >= CURRENT_DATE - INTERVAL '30 days'`, cid)

	// Stock by category
	catRows, _ := h.db.Query(ctx,
		`SELECT COALESCE(cat.name,'Uncategorized'),
		        COUNT(i.id), COALESCE(SUM(sl.quantity),0),
		        COALESCE(SUM(sl.quantity * COALESCE(i.cost_price, i.sale_price, 0)),0)
		 FROM items i
		 LEFT JOIN item_categories cat ON cat.id=i.category_id
		 LEFT JOIN stock_levels sl ON sl.item_id=i.id
		 WHERE i.company_id=$1 AND i.is_active=true
		 GROUP BY cat.name ORDER BY 4 DESC LIMIT 10`, cid)
	byCategory := []map[string]interface{}{}
	if catRows != nil {
		for catRows.Next() {
			var name string
			var cnt int
			var qty, val float64
			_ = catRows.Scan(&name, &cnt, &qty, &val)
			byCategory = append(byCategory, map[string]interface{}{
				"category": name, "item_count": cnt, "quantity": qty, "value": val,
			})
		}
		catRows.Close()
	}

	// Stock by warehouse
	whRows, _ := h.db.Query(ctx,
		`SELECT w.name, COUNT(DISTINCT sl.item_id), COALESCE(SUM(sl.quantity),0),
		        COALESCE(SUM(sl.quantity * COALESCE(i.cost_price, i.sale_price, 0)),0)
		 FROM warehouses w
		 LEFT JOIN stock_levels sl ON sl.warehouse_id=w.id
		 LEFT JOIN items i ON i.id=sl.item_id
		 WHERE w.company_id=$1 AND w.is_active=true
		 GROUP BY w.name ORDER BY 4 DESC`, cid)
	byWarehouse := []map[string]interface{}{}
	if whRows != nil {
		for whRows.Next() {
			var name string
			var cnt int
			var qty, val float64
			_ = whRows.Scan(&name, &cnt, &qty, &val)
			byWarehouse = append(byWarehouse, map[string]interface{}{
				"warehouse": name, "item_count": cnt, "quantity": qty, "value": val,
			})
		}
		whRows.Close()
	}

	// Low-stock items
	lowStockRows, _ := h.db.Query(ctx,
		`SELECT i.code, i.name, COALESCE(SUM(sl.quantity),0), i.reorder_point
		 FROM items i
		 JOIN stock_levels sl ON sl.item_id=i.id
		 WHERE i.company_id=$1 AND i.is_active=true AND i.reorder_point > 0
		 GROUP BY i.id, i.code, i.name, i.reorder_point
		 HAVING SUM(sl.quantity) <= i.reorder_point
		 ORDER BY (SUM(sl.quantity) / NULLIF(i.reorder_point,0)) ASC LIMIT 20`, cid)
	lowStockItems := []map[string]interface{}{}
	if lowStockRows != nil {
		for lowStockRows.Next() {
			var code, name string
			var qty, rp float64
			_ = lowStockRows.Scan(&code, &name, &qty, &rp)
			lowStockItems = append(lowStockItems, map[string]interface{}{
				"code": code, "name": name, "quantity": qty, "reorder_point": rp,
			})
		}
		lowStockRows.Close()
	}

	// Recent movements (last 30 days by type)
	mvtRows, _ := h.db.Query(ctx,
		`SELECT movement_type, COUNT(*), COALESCE(SUM(quantity),0)
		 FROM inventory_movements im
		 JOIN items i ON i.id=im.item_id
		 WHERE i.company_id=$1 AND im.created_at >= CURRENT_DATE - INTERVAL '30 days'
		 GROUP BY movement_type ORDER BY 3 DESC`, cid)
	byMovementType := []map[string]interface{}{}
	if mvtRows != nil {
		for mvtRows.Next() {
			var mvt string
			var cnt int
			var qty float64
			_ = mvtRows.Scan(&mvt, &cnt, &qty)
			byMovementType = append(byMovementType, map[string]interface{}{
				"type": mvt, "count": cnt, "quantity": qty,
			})
		}
		mvtRows.Close()
	}

	c.JSON(http.StatusOK, gin.H{
		"total_items":       totalItems,
		"total_value":       totalValue,
		"low_stock_count":   lowStockCount,
		"out_of_stock":      outOfStockCount,
		"warehouse_count":   warehouseCount,
		"movements_30d":     movementCount,
		"by_category":       byCategory,
		"by_warehouse":      byWarehouse,
		"low_stock_items":   lowStockItems,
		"by_movement_type":  byMovementType,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// PROJECT REPORTS — GET /reports/projects
// ─────────────────────────────────────────────────────────────────────────────

func (h *ReportsBIHandler) ProjectReports(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	ctx := context.Background()
	year := biYear(c)

	totalProjects := scanI(h.db, ctx, `SELECT COUNT(*) FROM projects WHERE company_id=$1`, cid)
	activeProjects := scanI(h.db, ctx, `SELECT COUNT(*) FROM projects WHERE company_id=$1 AND status='in_progress'`, cid)
	completedProjects := scanI(h.db, ctx, `SELECT COUNT(*) FROM projects WHERE company_id=$1 AND status='completed'`, cid)
	totalBudget := scanF(h.db, ctx, `SELECT COALESCE(SUM(budget),0) FROM projects WHERE company_id=$1`, cid)
	totalActual := scanF(h.db, ctx, `SELECT COALESCE(SUM(actual_cost),0) FROM projects WHERE company_id=$1`, cid)
	totalHours := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(t.hours),0) FROM timesheets t
		 JOIN projects p ON p.id=t.project_id
		 WHERE p.company_id=$1 AND EXTRACT(YEAR FROM t.date)=$2`, cid, year)
	billableHours := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(t.hours),0) FROM timesheets t
		 JOIN projects p ON p.id=t.project_id
		 WHERE p.company_id=$1 AND t.billable=true AND EXTRACT(YEAR FROM t.date)=$2`, cid, year)
	totalExpenses := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(pe.amount),0) FROM project_expenses pe
		 JOIN projects p ON p.id=pe.project_id
		 WHERE p.company_id=$1 AND pe.status IN ('approved','paid')
		   AND EXTRACT(YEAR FROM pe.date)=$2`, cid, year)

	// Project list with KPIs
	projRows, _ := h.db.Query(ctx,
		`SELECT p.id, p.code, p.name, p.status, p.progress_pct,
		        p.budget, p.actual_cost,
		        COALESCE(ts.hrs,0), COALESCE(ts.bhrs,0), COALESCE(ex.amt,0),
		        COALESCE(tk.total,0), COALESCE(tk.done,0)
		 FROM projects p
		 LEFT JOIN LATERAL (
		   SELECT COALESCE(SUM(hours),0) hrs, COALESCE(SUM(hours) FILTER (WHERE billable),0) bhrs
		   FROM timesheets WHERE project_id=p.id AND EXTRACT(YEAR FROM date)=$2
		 ) ts ON TRUE
		 LEFT JOIN LATERAL (
		   SELECT COALESCE(SUM(amount),0) amt
		   FROM project_expenses WHERE project_id=p.id AND status IN ('approved','paid')
		 ) ex ON TRUE
		 LEFT JOIN LATERAL (
		   SELECT COUNT(*) total, COUNT(*) FILTER (WHERE status='done') done
		   FROM project_tasks WHERE project_id=p.id
		 ) tk ON TRUE
		 WHERE p.company_id=$1
		 ORDER BY p.created_at DESC`, cid, year)
	projects := []map[string]interface{}{}
	if projRows != nil {
		for projRows.Next() {
			var id, code, name, status string
			var prog, tkTotal, tkDone int
			var budget, actual, hrs, bhrs, exAmt float64
			if projRows.Scan(&id, &code, &name, &status, &prog,
				&budget, &actual, &hrs, &bhrs, &exAmt, &tkTotal, &tkDone) != nil {
				continue
			}
			budgetUsed := 0.0
			if budget > 0 {
				budgetUsed = actual / budget * 100
			}
			projects = append(projects, map[string]interface{}{
				"id": id, "code": code, "name": name, "status": status,
				"progress_pct": prog, "budget": budget, "actual_cost": actual,
				"budget_used_pct": budgetUsed, "variance": budget - actual,
				"total_hours": hrs, "billable_hours": bhrs, "expenses": exAmt,
				"total_tasks": tkTotal, "done_tasks": tkDone,
			})
		}
		projRows.Close()
	}

	// By status breakdown
	statusRows, _ := h.db.Query(ctx,
		`SELECT status, COUNT(*), COALESCE(SUM(budget),0), COALESCE(SUM(actual_cost),0)
		 FROM projects WHERE company_id=$1 GROUP BY status ORDER BY 2 DESC`, cid)
	byStatus := []map[string]interface{}{}
	if statusRows != nil {
		for statusRows.Next() {
			var st string
			var cnt int
			var bud, act float64
			_ = statusRows.Scan(&st, &cnt, &bud, &act)
			byStatus = append(byStatus, map[string]interface{}{
				"status": st, "count": cnt, "budget": bud, "actual": act,
			})
		}
		statusRows.Close()
	}

	c.JSON(http.StatusOK, gin.H{
		"year":               year,
		"total_projects":     totalProjects,
		"active_projects":    activeProjects,
		"completed_projects": completedProjects,
		"total_budget":       totalBudget,
		"total_actual":       totalActual,
		"budget_variance":    totalBudget - totalActual,
		"total_hours":        totalHours,
		"billable_hours":     billableHours,
		"billable_rate":      func() float64 { if totalHours > 0 { return billableHours / totalHours * 100 }; return 0 }(),
		"total_expenses":     totalExpenses,
		"projects":           projects,
		"by_status":          byStatus,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// MANAGEMENT REPORTS — GET /reports/management
// ─────────────────────────────────────────────────────────────────────────────

func (h *ReportsBIHandler) ManagementReports(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	ctx := context.Background()
	year := biYear(c)

	// Core KPIs
	revenue := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount),0) FROM sales_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year)
	prevRevenue := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount),0) FROM sales_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year-1)
	revenueGrowth := 0.0
	if prevRevenue > 0 {
		revenueGrowth = (revenue - prevRevenue) / prevRevenue * 100
	}

	expenses := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount),0) FROM purchase_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year)
	netProfit := revenue - expenses
	netMargin := 0.0
	if revenue > 0 {
		netMargin = netProfit / revenue * 100
	}

	employees := scanI(h.db, ctx, `SELECT COUNT(*) FROM employees WHERE company_id=$1 AND status='active'`, cid)
	payroll := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(net_salary),0) FROM payroll_slips ps
		 JOIN payroll_runs pr ON pr.id=ps.payroll_run_id
		 WHERE pr.company_id=$1 AND EXTRACT(YEAR FROM pr.pay_date)=$2`, cid, year)

	revenuePerEmployee := 0.0
	if employees > 0 {
		revenuePerEmployee = revenue / float64(employees)
	}

	// Quarterly P&L
	quarterlyPL := []map[string]interface{}{}
	for q := 1; q <= 4; q++ {
		startM := (q-1)*3 + 1
		endM := q * 3
		qRev := scanF(h.db, ctx,
			`SELECT COALESCE(SUM(total_amount),0) FROM sales_invoices
			 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
			   AND EXTRACT(MONTH FROM date) BETWEEN $3 AND $4
			   AND status NOT IN ('cancelled','draft')`, cid, year, startM, endM)
		qExp := scanF(h.db, ctx,
			`SELECT COALESCE(SUM(total_amount),0) FROM purchase_invoices
			 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
			   AND EXTRACT(MONTH FROM date) BETWEEN $3 AND $4
			   AND status NOT IN ('cancelled','draft')`, cid, year, startM, endM)
		qPayroll := scanF(h.db, ctx,
			`SELECT COALESCE(SUM(net_salary),0) FROM payroll_slips ps
			 JOIN payroll_runs pr ON pr.id=ps.payroll_run_id
			 WHERE pr.company_id=$1 AND EXTRACT(YEAR FROM pr.pay_date)=$2
			   AND EXTRACT(MONTH FROM pr.pay_date) BETWEEN $3 AND $4`, cid, year, startM, endM)
		qNet := qRev - qExp - qPayroll
		qMargin := 0.0
		if qRev > 0 {
			qMargin = qNet / qRev * 100
		}
		quarterlyPL = append(quarterlyPL, map[string]interface{}{
			"quarter": q, "revenue": qRev, "expenses": qExp + qPayroll,
			"net_profit": qNet, "margin_pct": qMargin,
		})
	}

	// Department costs
	deptRows, _ := h.db.Query(ctx,
		`SELECT d.name, COUNT(e.id), COALESCE(SUM(ps.gross_salary),0)
		 FROM departments d
		 LEFT JOIN employees e ON e.department_id=d.id AND e.status='active'
		 LEFT JOIN payroll_slips ps ON ps.employee_id=e.id
		   AND ps.payroll_run_id IN (
		     SELECT id FROM payroll_runs WHERE company_id=$1 AND EXTRACT(YEAR FROM pay_date)=$2
		   )
		 WHERE d.company_id=$1
		 GROUP BY d.name ORDER BY 3 DESC`, cid, year)
	deptCosts := []map[string]interface{}{}
	if deptRows != nil {
		for deptRows.Next() {
			var name string
			var cnt int
			var cost float64
			_ = deptRows.Scan(&name, &cnt, &cost)
			deptCosts = append(deptCosts, map[string]interface{}{
				"department": name, "headcount": cnt, "payroll_cost": cost,
			})
		}
		deptRows.Close()
	}

	// Scorecard
	arDays := scanF(h.db, ctx,
		`SELECT COALESCE(AVG(EXTRACT(DAY FROM CURRENT_DATE - date)),0)
		 FROM sales_invoices
		 WHERE company_id=$1 AND status IN ('confirmed','partial')`, cid)
	apDays := scanF(h.db, ctx,
		`SELECT COALESCE(AVG(EXTRACT(DAY FROM CURRENT_DATE - date)),0)
		 FROM purchase_invoices
		 WHERE company_id=$1 AND status IN ('confirmed','partial')`, cid)
	inventoryTurnover := 0.0
	invValue := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(sl.quantity * COALESCE(i.cost_price, i.sale_price, 0)),0)
		 FROM stock_levels sl JOIN items i ON i.id=sl.item_id WHERE i.company_id=$1`, cid)
	if invValue > 0 {
		inventoryTurnover = expenses / invValue
	}

	c.JSON(http.StatusOK, gin.H{
		"year":                 year,
		"revenue":              revenue,
		"prev_revenue":         prevRevenue,
		"revenue_growth_pct":   revenueGrowth,
		"expenses":             expenses,
		"net_profit":           netProfit,
		"net_margin_pct":       netMargin,
		"employees":            employees,
		"payroll":              payroll,
		"revenue_per_employee": revenuePerEmployee,
		"ar_days":              arDays,
		"ap_days":              apDays,
		"inventory_turnover":   inventoryTurnover,
		"quarterly_pl":         quarterlyPL,
		"department_costs":     deptCosts,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// ANALYTICS — GET /reports/analytics
// ─────────────────────────────────────────────────────────────────────────────

func (h *ReportsBIHandler) Analytics(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	ctx := context.Background()
	year := biYear(c)

	// YoY revenue comparison (last 3 years)
	yoyRows, _ := h.db.Query(ctx,
		`SELECT EXTRACT(YEAR FROM date)::int, COALESCE(SUM(total_amount),0)
		 FROM sales_invoices
		 WHERE company_id=$1
		   AND EXTRACT(YEAR FROM date) BETWEEN $2 AND $3
		   AND status NOT IN ('cancelled','draft')
		 GROUP BY 1 ORDER BY 1`, cid, year-2, year)
	yoyRevenue := []map[string]interface{}{}
	if yoyRows != nil {
		for yoyRows.Next() {
			var yr int
			var rev float64
			_ = yoyRows.Scan(&yr, &rev)
			yoyRevenue = append(yoyRevenue, map[string]interface{}{"year": yr, "revenue": rev})
		}
		yoyRows.Close()
	}

	// Revenue by month (current year vs previous year)
	prevMonthlyRows, _ := h.db.Query(ctx,
		`SELECT EXTRACT(MONTH FROM date)::int, COALESCE(SUM(total_amount),0)
		 FROM sales_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')
		 GROUP BY 1 ORDER BY 1`, cid, year-1)
	prevMonthly := make([]float64, 12)
	if prevMonthlyRows != nil {
		for prevMonthlyRows.Next() {
			var mo int
			var amt float64
			_ = prevMonthlyRows.Scan(&mo, &amt)
			if mo >= 1 && mo <= 12 {
				prevMonthly[mo-1] = amt
			}
		}
		prevMonthlyRows.Close()
	}
	currMonthlyRows, _ := h.db.Query(ctx,
		`SELECT EXTRACT(MONTH FROM date)::int, COALESCE(SUM(total_amount),0)
		 FROM sales_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')
		 GROUP BY 1 ORDER BY 1`, cid, year)
	currMonthly := make([]float64, 12)
	if currMonthlyRows != nil {
		for currMonthlyRows.Next() {
			var mo int
			var amt float64
			_ = currMonthlyRows.Scan(&mo, &amt)
			if mo >= 1 && mo <= 12 {
				currMonthly[mo-1] = amt
			}
		}
		currMonthlyRows.Close()
	}
	monthlyComparison := []map[string]interface{}{}
	for i := 0; i < 12; i++ {
		growth := 0.0
		if prevMonthly[i] > 0 {
			growth = (currMonthly[i] - prevMonthly[i]) / prevMonthly[i] * 100
		}
		monthlyComparison = append(monthlyComparison, map[string]interface{}{
			"month": i + 1, "current": currMonthly[i], "previous": prevMonthly[i], "growth_pct": growth,
		})
	}

	// Customer segmentation (by revenue brackets)
	custSegRows, _ := h.db.Query(ctx,
		`SELECT
		   CASE
		     WHEN rev >= 1000000 THEN 'Enterprise (>1M)'
		     WHEN rev >= 100000  THEN 'Large (100K-1M)'
		     WHEN rev >= 10000   THEN 'Medium (10K-100K)'
		     ELSE 'Small (<10K)'
		   END AS segment,
		   COUNT(*) AS customers,
		   SUM(rev) AS total_revenue
		 FROM (
		   SELECT customer_id, SUM(total_amount) AS rev
		   FROM sales_invoices
		   WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		     AND status NOT IN ('cancelled','draft')
		   GROUP BY customer_id
		 ) t
		 GROUP BY segment ORDER BY total_revenue DESC`, cid, year)
	customerSegments := []map[string]interface{}{}
	if custSegRows != nil {
		for custSegRows.Next() {
			var seg string
			var cnt int
			var rev float64
			_ = custSegRows.Scan(&seg, &cnt, &rev)
			customerSegments = append(customerSegments, map[string]interface{}{
				"segment": seg, "customers": cnt, "revenue": rev,
			})
		}
		custSegRows.Close()
	}

	// Product performance (top 10 by contribution)
	prodPerfRows, _ := h.db.Query(ctx,
		`SELECT il.description, COALESCE(SUM(il.quantity),0) AS qty,
		        COALESCE(SUM(il.total_price),0) AS rev,
		        COALESCE(SUM(il.total_price) / NULLIF(SUM(SUM(il.total_price)) OVER(),0) * 100, 0) AS contribution_pct
		 FROM invoice_lines il
		 JOIN sales_invoices si ON si.id=il.invoice_id
		 WHERE si.company_id=$1 AND EXTRACT(YEAR FROM si.date)=$2
		   AND si.status NOT IN ('cancelled','draft')
		 GROUP BY il.description ORDER BY rev DESC LIMIT 10`, cid, year)
	productPerformance := []map[string]interface{}{}
	if prodPerfRows != nil {
		for prodPerfRows.Next() {
			var name string
			var qty, rev, pct float64
			_ = prodPerfRows.Scan(&name, &qty, &rev, &pct)
			productPerformance = append(productPerformance, map[string]interface{}{
				"product": name, "quantity": qty, "revenue": rev, "contribution_pct": pct,
			})
		}
		prodPerfRows.Close()
	}

	// Tax burden analysis
	totalTaxDue := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(amount_due),0) FROM tax_payments
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM due_date)=$2`, cid, year)
	totalTaxPaid := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(amount_paid),0) FROM tax_payments
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM due_date)=$2`, cid, year)
	revenue2 := scanF(h.db, ctx,
		`SELECT COALESCE(SUM(total_amount),0) FROM sales_invoices
		 WHERE company_id=$1 AND EXTRACT(YEAR FROM date)=$2
		   AND status NOT IN ('cancelled','draft')`, cid, year)
	effectiveTaxRate := 0.0
	if revenue2 > 0 {
		effectiveTaxRate = totalTaxDue / revenue2 * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"year":                year,
		"yoy_revenue":         yoyRevenue,
		"monthly_comparison":  monthlyComparison,
		"customer_segments":   customerSegments,
		"product_performance": productPerformance,
		"total_tax_due":       totalTaxDue,
		"total_tax_paid":      totalTaxPaid,
		"effective_tax_rate":  effectiveTaxRate,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// REPORT DEFINITIONS CRUD
// ─────────────────────────────────────────────────────────────────────────────

func (h *ReportsBIHandler) ListReportDefinitions(c *gin.Context) {
	cid := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, err := h.db.Query(ctx,
		`SELECT id, name, description, report_type, category, schedule, is_public, created_at
		 FROM report_definitions WHERE company_id=$1 ORDER BY created_at DESC`, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	list := []map[string]interface{}{}
	for rows.Next() {
		var id, name, desc, rt, cat, sched string
		var isPub bool
		var createdAt interface{}
		_ = rows.Scan(&id, &name, &desc, &rt, &cat, &sched, &isPub, &createdAt)
		list = append(list, map[string]interface{}{
			"id": id, "name": name, "description": desc,
			"report_type": rt, "category": cat, "schedule": sched,
			"is_public": isPub, "created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, list)
}
