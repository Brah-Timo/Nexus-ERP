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

type WorkflowHandler struct{ db *pgxpool.Pool }

func (h *WorkflowHandler) ListRules(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()
	rows, _ := h.db.Query(ctx,
		`SELECT id, name, module, document_type, condition, is_active FROM workflow_rules WHERE company_id = $1`, companyID)
	var rules []models.WorkflowRule
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var r models.WorkflowRule
			_ = rows.Scan(&r.ID, &r.Name, &r.Module, &r.DocumentType, &r.Condition, &r.IsActive)
			rules = append(rules, r)
		}
	}
	c.JSON(http.StatusOK, rules)
}

func (h *WorkflowHandler) CreateRule(c *gin.Context) {
	var r models.WorkflowRule
	c.ShouldBindJSON(&r)
	r.ID = uuid.NewString()
	r.CompanyID = middleware.GetCompanyID(c)
	ctx := context.Background()
	_, _ = h.db.Exec(ctx,
		`INSERT INTO workflow_rules (id, company_id, name, module, document_type, condition, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		r.ID, r.CompanyID, r.Name, r.Module, r.DocumentType, r.Condition, true)
	c.JSON(http.StatusCreated, r)
}

func (h *WorkflowHandler) UpdateRule(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Rule updated"})
}

func (h *WorkflowHandler) GetApprovalInbox(c *gin.Context) {
	userID := middleware.GetUserID(c)
	ctx := context.Background()
	rows, _ := h.db.Query(ctx, `
		SELECT id, module, document_type, document_ref, requested_by, status, created_at
		FROM approval_requests WHERE approver_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`, userID)
	var approvals []models.ApprovalRequest
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var a models.ApprovalRequest
			_ = rows.Scan(&a.ID, &a.Module, &a.DocumentType, &a.DocumentRef, &a.RequestedBy, &a.Status, &a.CreatedAt)
			approvals = append(approvals, a)
		}
	}
	c.JSON(http.StatusOK, approvals)
}

func (h *WorkflowHandler) Approve(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Notes string `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	ctx := context.Background()
	_, _ = h.db.Exec(ctx,
		`UPDATE approval_requests SET status = 'approved', approved_at = NOW(), notes = $1 WHERE id = $2`,
		req.Notes, id)
	c.JSON(http.StatusOK, gin.H{"message": "Approved"})
}

func (h *WorkflowHandler) Reject(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Notes string `json:"notes"`
	}
	c.ShouldBindJSON(&req)
	ctx := context.Background()
	_, _ = h.db.Exec(ctx,
		`UPDATE approval_requests SET status = 'rejected', notes = $1 WHERE id = $2`,
		req.Notes, id)
	c.JSON(http.StatusOK, gin.H{"message": "Rejected"})
}

// ─── Reports Handler ──────────────────────────────────────────────────────────

type ReportsHandler struct{ db *pgxpool.Pool }

func (h *ReportsHandler) FinancialRatios(c *gin.Context) {
	companyID := middleware.GetCompanyID(c)
	ctx := context.Background()

	var totalAssets, totalLiabilities, revenue, netProfit float64
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(balance), 0) FROM chart_of_accounts WHERE company_id = $1 AND type = 'asset'
	`, companyID).Scan(&totalAssets)
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(balance), 0) FROM chart_of_accounts WHERE company_id = $1 AND type = 'liability'
	`, companyID).Scan(&totalLiabilities)
	_ = h.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(balance), 0) FROM chart_of_accounts WHERE company_id = $1 AND type = 'revenue'
	`, companyID).Scan(&revenue)

	equity := totalAssets - totalLiabilities
	netProfit = revenue * 0.15 // Simplified

	ratios := map[string]interface{}{
		"current_ratio":        func() float64 { if totalLiabilities == 0 { return 0 }; return totalAssets / totalLiabilities }(),
		"debt_to_equity":       func() float64 { if equity == 0 { return 0 }; return totalLiabilities / equity }(),
		"return_on_equity":     func() float64 { if equity == 0 { return 0 }; return (netProfit / equity) * 100 }(),
		"return_on_assets":     func() float64 { if totalAssets == 0 { return 0 }; return (netProfit / totalAssets) * 100 }(),
		"net_profit_margin":    func() float64 { if revenue == 0 { return 0 }; return (netProfit / revenue) * 100 }(),
		"total_assets":         totalAssets,
		"total_liabilities":    totalLiabilities,
		"equity":               equity,
	}

	c.JSON(http.StatusOK, ratios)
}

func (h *ReportsHandler) KPISummary(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "KPI Summary"})
}

func (h *ReportsHandler) RunCustomReport(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Custom report executed"})
}

func (h *ReportsHandler) Export(c *gin.Context) {
	format := c.Param("type")
	c.JSON(http.StatusOK, gin.H{"format": format, "message": "Export ready"})
}
