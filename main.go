package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"nexus-erp/internal/database"
	"nexus-erp/internal/handler"
	"nexus-erp/internal/middleware"
)

//go:embed web/dist
var frontend embed.FS

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db, err := database.NewPool(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize handlers
	h := handler.NewHandler(db)

	// Setup Gin
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Serve embedded frontend static files
	dist, err := fs.Sub(frontend, "web/dist")
	if err != nil {
		log.Fatalf("Failed to create sub filesystem: %v", err)
	}

	// Static assets (JS, CSS, images)
	r.StaticFS("/assets", http.FS(must(fs.Sub(dist, "assets"))))

	// SPA fallback — return index.html for all unmatched routes
	r.NoRoute(func(c *gin.Context) {
		// Only serve index.html for non-API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		index, err := fs.ReadFile(dist, "index.html")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})

	// ─── Public routes ────────────────────────────────────────────────────────
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", h.Auth.Login)
		auth.POST("/logout", h.Auth.Logout)
		auth.POST("/refresh", h.Auth.RefreshToken)
		auth.POST("/forgot-password", h.Auth.ForgotPassword)
		auth.POST("/reset-password", h.Auth.ResetPassword)
	}

	// ─── Protected routes ─────────────────────────────────────────────────────
	api := r.Group("/api", middleware.JWTAuth())
	{
		// Dashboard
		api.GET("/dashboard/summary", h.Dashboard.GetSummary)
		api.GET("/dashboard/cashflow", h.Dashboard.GetCashFlow)
		api.GET("/dashboard/activity", h.Dashboard.GetRecentActivity)
		api.GET("/dashboard/approvals", h.Dashboard.GetPendingApprovals)

		// ── Settings ──────────────────────────────────────────────────────────
		settings := api.Group("/settings")
		{
			settings.GET("/companies", h.Settings.ListCompanies)
			settings.POST("/companies", h.Settings.CreateCompany)
			settings.PUT("/companies/:id", h.Settings.UpdateCompany)

			settings.GET("/users", h.Settings.ListUsers)
			settings.POST("/users", h.Settings.CreateUser)
			settings.PUT("/users/:id", h.Settings.UpdateUser)
			settings.DELETE("/users/:id", h.Settings.DeleteUser)

			settings.GET("/roles", h.Settings.ListRoles)
			settings.POST("/roles", h.Settings.CreateRole)
			settings.PUT("/roles/:id", h.Settings.UpdateRole)
			settings.DELETE("/roles/:id", h.Settings.DeleteRole)

			settings.GET("/fiscal-years", h.Settings.ListFiscalYears)
			settings.POST("/fiscal-years", h.Settings.CreateFiscalYear)
			settings.PUT("/fiscal-years/:id/close", h.Settings.CloseFiscalYear)

			settings.GET("/currencies", h.Settings.ListCurrencies)
			settings.POST("/currencies", h.Settings.CreateCurrency)
			settings.PUT("/currencies/:id", h.Settings.UpdateExchangeRate)
			settings.DELETE("/currencies/:id", h.Settings.DeleteCurrency)

			settings.GET("/numbering", h.Settings.GetNumberingConfig)
			settings.PUT("/numbering", h.Settings.UpdateNumberingConfig)

			settings.GET("/taxes", h.Settings.ListTaxes)
			settings.POST("/taxes", h.Settings.CreateTax)
			settings.PUT("/taxes/:id", h.Settings.UpdateTax)
			settings.DELETE("/taxes/:id", h.Settings.DeleteTax)

			settings.GET("/workflow-rules", h.Settings.ListWorkflowRules)
			settings.POST("/workflow-rules", h.Settings.CreateWorkflowRule)
			settings.PUT("/workflow-rules/:id", h.Settings.UpdateWorkflowRule)
			settings.DELETE("/workflow-rules/:id", h.Settings.DeleteWorkflowRule)

			settings.GET("/audit-log", h.Settings.GetAuditLog)
		}

		// ── Accounting ─────────────────────────────────────────────────────────
		accounting := api.Group("/accounting")
		{
			accounting.GET("/chart-of-accounts", h.Accounting.ListChartOfAccounts)
			accounting.POST("/chart-of-accounts", h.Accounting.CreateAccount)
			accounting.PUT("/chart-of-accounts/:id", h.Accounting.UpdateAccount)

			accounting.GET("/journal-entries", h.Accounting.ListJournalEntries)
			accounting.GET("/journal-entries/:id", h.Accounting.GetJournalEntry)
			accounting.POST("/journal-entries", h.Accounting.CreateJournalEntry)
			accounting.PUT("/journal-entries/:id/post", h.Accounting.PostJournalEntry)
			accounting.PUT("/journal-entries/:id/cancel", h.Accounting.CancelJournalEntry)

			accounting.GET("/cost-centers", h.Accounting.ListCostCenters)
			accounting.POST("/cost-centers", h.Accounting.CreateCostCenter)

			accounting.GET("/fixed-assets", h.Accounting.ListFixedAssets)
			accounting.POST("/fixed-assets", h.Accounting.CreateFixedAsset)
			accounting.POST("/fixed-assets/depreciate", h.Accounting.RunDepreciation)

			accounting.GET("/bank-reconciliations", h.Accounting.ListBankReconciliations)
			accounting.POST("/bank-reconciliations", h.Accounting.CreateBankReconciliation)

			accounting.GET("/reports/trial-balance", h.Accounting.TrialBalance)
			accounting.GET("/reports/balance-sheet", h.Accounting.BalanceSheet)
			accounting.GET("/reports/income-statement", h.Accounting.IncomeStatement)
			accounting.GET("/reports/cash-flow", h.Accounting.CashFlowStatement)

			accounting.GET("/budgets", h.Accounting.ListBudgets)
			accounting.POST("/budgets", h.Accounting.CreateBudget)
		}

		// ── HR & Payroll ───────────────────────────────────────────────────────
		hr := api.Group("/hr")
		{
			// Dashboard
			hr.GET("/dashboard", h.HR.GetHRDashboard)

			// Employees
			hr.GET("/employees", h.HR.ListEmployees)
			hr.GET("/employees/:id", h.HR.GetEmployee)
			hr.POST("/employees", h.HR.CreateEmployee)
			hr.PUT("/employees/:id", h.HR.UpdateEmployee)
			hr.DELETE("/employees/:id", h.HR.DeleteEmployee)

			// Departments
			hr.GET("/departments", h.HR.ListDepartments)
			hr.POST("/departments", h.HR.CreateDepartment)
			hr.PUT("/departments/:id", h.HR.UpdateDepartment)
			hr.DELETE("/departments/:id", h.HR.DeleteDepartment)

			// Positions
			hr.GET("/positions", h.HR.ListPositions)
			hr.POST("/positions", h.HR.CreatePosition)

			// Attendance
			hr.GET("/attendance", h.HR.ListAttendance)
			hr.POST("/attendance", h.HR.RecordAttendance)
			hr.PUT("/attendance/:id", h.HR.UpdateAttendance)
			hr.GET("/attendance/summary", h.HR.GetAttendanceSummary)

			// Leave Types
			hr.GET("/leave-types", h.HR.ListLeaveTypes)
			hr.POST("/leave-types", h.HR.CreateLeaveType)

			// Leave Requests
			hr.GET("/leave-requests", h.HR.ListLeaveRequests)
			hr.POST("/leave-requests", h.HR.CreateLeaveRequest)
			hr.PUT("/leave-requests/:id/approve", h.HR.ApproveLeaveRequest)
			hr.PUT("/leave-requests/:id/reject", h.HR.RejectLeaveRequest)
			hr.PUT("/leave-requests/:id/cancel", h.HR.CancelLeaveRequest)

			// Payroll
			hr.GET("/payroll/runs", h.HR.ListPayrollRuns)
			hr.POST("/payroll/runs", h.HR.RunPayroll)
			hr.PUT("/payroll/runs/:id/approve", h.HR.ApprovePayrollRun)
			hr.PUT("/payroll/runs/:id/pay", h.HR.PayPayrollRun)
			hr.GET("/payroll/runs/:id/payslips", h.HR.GetPayslips)
			hr.GET("/payroll/runs/:id/g29", h.HR.ExportG29)

			// Recruitment — Job Postings
			hr.GET("/recruitment/jobs", h.HR.ListJobPostings)
			hr.GET("/recruitment/jobs/:id", h.HR.GetJobPosting)
			hr.POST("/recruitment/jobs", h.HR.CreateJobPosting)
			hr.PUT("/recruitment/jobs/:id", h.HR.UpdateJobPosting)
			hr.DELETE("/recruitment/jobs/:id", h.HR.DeleteJobPosting)

			// Recruitment — Applications
			hr.GET("/recruitment/applications", h.HR.ListApplications)
			hr.POST("/recruitment/applications", h.HR.CreateApplication)
			hr.PUT("/recruitment/applications/:id/status", h.HR.UpdateApplicationStatus)

			// Legacy compat
			hr.GET("/recruitment", h.HR.ListCandidates)
			hr.POST("/recruitment", h.HR.CreateCandidate)
			hr.GET("/performance-reviews", h.HR.ListPerformanceReviews)
			hr.POST("/performance-reviews", h.HR.CreatePerformanceReview)
		}

		// ── Sales & CRM ────────────────────────────────────────────────────────
		sales := api.Group("/sales")
		{
			// Leads
			sales.GET("/leads", h.Sales.ListLeads)
			sales.POST("/leads", h.Sales.CreateLead)
			sales.PUT("/leads/:id", h.Sales.UpdateLead)
			sales.DELETE("/leads/:id", h.Sales.DeleteLead)

			// Opportunities / Pipeline
			sales.GET("/opportunities", h.Sales.ListOpportunities)
			sales.POST("/opportunities", h.Sales.CreateOpportunity)
			sales.PUT("/opportunities/:id", h.Sales.UpdateOpportunity)
			sales.DELETE("/opportunities/:id", h.Sales.DeleteOpportunity)
			sales.GET("/pipeline/summary", h.Sales.PipelineSummary)

			// Customers
			sales.GET("/customers", h.Sales.ListCustomers)
			sales.GET("/customers/:id", h.Sales.GetCustomer)
			sales.POST("/customers", h.Sales.CreateCustomer)
			sales.PUT("/customers/:id", h.Sales.UpdateCustomer)
			sales.DELETE("/customers/:id", h.Sales.DeleteCustomer)

			// Quotations
			sales.GET("/quotations", h.Sales.ListQuotations)
			sales.GET("/quotations/:id", h.Sales.GetQuotation)
			sales.POST("/quotations", h.Sales.CreateQuotation)
			sales.PUT("/quotations/:id", h.Sales.UpdateQuotation)
			sales.PUT("/quotations/:id/confirm", h.Sales.ConfirmQuotation)
			sales.PUT("/quotations/:id/cancel", h.Sales.CancelQuotation)
			sales.POST("/quotations/:id/convert", h.Sales.ConvertToOrder)

			// Sales Orders
			sales.GET("/orders", h.Sales.ListOrders)
			sales.GET("/orders/:id", h.Sales.GetOrder)
			sales.POST("/orders", h.Sales.CreateOrder)
			sales.PUT("/orders/:id", h.Sales.UpdateOrder)
			sales.PUT("/orders/:id/confirm", h.Sales.ConfirmOrder)
			sales.PUT("/orders/:id/deliver", h.Sales.DeliverOrder)
			sales.PUT("/orders/:id/cancel", h.Sales.CancelOrder)

			// Sales Invoices
			sales.GET("/invoices", h.Sales.ListInvoices)
			sales.GET("/invoices/:id", h.Sales.GetInvoice)
			sales.POST("/invoices", h.Sales.CreateInvoice)
			sales.PUT("/invoices/:id/confirm", h.Sales.ConfirmInvoice)
			sales.PUT("/invoices/:id/cancel", h.Sales.CancelInvoice)
			sales.POST("/invoices/:id/payment", h.Sales.RecordPayment)

			sales.GET("/credit-notes", h.Sales.ListCreditNotes)
			sales.POST("/credit-notes", h.Sales.CreateCreditNote)

			sales.GET("/commissions", h.Sales.ListCommissions)
			sales.GET("/reports/aging", h.Sales.AgingReport)
		}

		// ── Purchase ───────────────────────────────────────────────────────────
		purchase := api.Group("/purchase")
		{
			purchase.GET("/suppliers", h.Purchase.ListSuppliers)
			purchase.GET("/suppliers/:id", h.Purchase.GetSupplier)
			purchase.POST("/suppliers", h.Purchase.CreateSupplier)
			purchase.PUT("/suppliers/:id", h.Purchase.UpdateSupplier)

			purchase.GET("/rfqs", h.Purchase.ListRFQs)
			purchase.POST("/rfqs", h.Purchase.CreateRFQ)
			purchase.PUT("/rfqs/:id/send", h.Purchase.SendRFQ)

			purchase.GET("/orders", h.Purchase.ListOrders)
			purchase.GET("/orders/:id", h.Purchase.GetOrder)
			purchase.POST("/orders", h.Purchase.CreateOrder)
			purchase.PUT("/orders/:id/approve", h.Purchase.ApprovePurchaseOrder)
			purchase.PUT("/orders/:id/confirm", h.Purchase.ConfirmPurchaseOrder)

			purchase.GET("/receipts", h.Purchase.ListGoodsReceipts)
			purchase.POST("/receipts", h.Purchase.CreateGoodsReceipt)

			purchase.GET("/invoices", h.Purchase.ListInvoices)
			purchase.POST("/invoices", h.Purchase.CreateInvoice)
			purchase.PUT("/invoices/:id/match", h.Purchase.ThreeWayMatch)
			purchase.POST("/invoices/:id/payment", h.Purchase.RecordPayment)

			purchase.GET("/supplier-evaluations", h.Purchase.ListEvaluations)
			purchase.POST("/supplier-evaluations", h.Purchase.CreateEvaluation)

			purchase.GET("/reports/aging", h.Purchase.AgingReport)
		}

		// ── Inventory ──────────────────────────────────────────────────────────
		inventory := api.Group("/inventory")
		{
			inventory.GET("/items", h.Inventory.ListItems)
			inventory.GET("/items/:id", h.Inventory.GetItem)
			inventory.POST("/items", h.Inventory.CreateItem)
			inventory.PUT("/items/:id", h.Inventory.UpdateItem)
			inventory.DELETE("/items/:id", h.Inventory.DeactivateItem)

			inventory.GET("/categories", h.Inventory.ListCategories)
			inventory.POST("/categories", h.Inventory.CreateCategory)

			inventory.GET("/units", h.Inventory.ListUnits)
			inventory.POST("/units", h.Inventory.CreateUnit)

			inventory.GET("/warehouses", h.Inventory.ListWarehouses)
			inventory.GET("/warehouses/:id", h.Inventory.GetWarehouse)
			inventory.POST("/warehouses", h.Inventory.CreateWarehouse)
			inventory.PUT("/warehouses/:id", h.Inventory.UpdateWarehouse)

			inventory.GET("/locations", h.Inventory.ListLocations)
			inventory.POST("/locations", h.Inventory.CreateLocation)

			inventory.GET("/movements", h.Inventory.ListMovements)
			inventory.POST("/movements/transfer", h.Inventory.TransferStock)
			inventory.POST("/movements/adjustment", h.Inventory.AdjustStock)

			inventory.GET("/stock-levels", h.Inventory.GetStockLevels)
			inventory.GET("/lots", h.Inventory.ListLots)

			inventory.GET("/inventory-counts", h.Inventory.ListInventoryCounts)
			inventory.GET("/inventory-counts/:id", h.Inventory.GetInventoryCount)
			inventory.POST("/inventory-counts", h.Inventory.CreateInventoryCount)
			inventory.PUT("/inventory-counts/:id/validate", h.Inventory.ValidateInventoryCount)

			inventory.GET("/reports/valuation", h.Inventory.ValuationReport)
			inventory.GET("/dashboard", h.Inventory.GetInventoryDashboard)
		}

		// ── Manufacturing ──────────────────────────────────────────────────────
		mfg := api.Group("/manufacturing")
		{
			// BOM
			mfg.GET("/bom", h.Manufacturing.ListBOMs)
			mfg.GET("/bom/:id", h.Manufacturing.GetBOM)
			mfg.POST("/bom", h.Manufacturing.CreateBOM)
			mfg.PUT("/bom/:id", h.Manufacturing.UpdateBOM)
			mfg.DELETE("/bom/:id", h.Manufacturing.DeactivateBOM)

			// Work Centers
			mfg.GET("/work-centers", h.Manufacturing.ListWorkCenters)
			mfg.POST("/work-centers", h.Manufacturing.CreateWorkCenter)
			mfg.PUT("/work-centers/:id", h.Manufacturing.UpdateWorkCenter)
			mfg.DELETE("/work-centers/:id", h.Manufacturing.DeactivateWorkCenter)

			// Manufacturing Orders
			mfg.GET("/orders", h.Manufacturing.ListOrders)
			mfg.GET("/orders/:id", h.Manufacturing.GetOrder)
			mfg.POST("/orders", h.Manufacturing.CreateOrder)
			mfg.PUT("/orders/:id", h.Manufacturing.UpdateOrder)
			mfg.PUT("/orders/:id/start", h.Manufacturing.StartOrder)
			mfg.PUT("/orders/:id/complete", h.Manufacturing.CompleteOrder)
			mfg.PUT("/orders/:id/cancel", h.Manufacturing.CancelOrder)

			// Quality Inspections
			mfg.GET("/quality-inspections", h.Manufacturing.ListQualityInspections)
			mfg.POST("/quality-inspections", h.Manufacturing.CreateQualityInspection)

			// MRP & Dashboard
			mfg.POST("/mrp/suggest", h.Manufacturing.RunMRP)
			mfg.GET("/dashboard", h.Manufacturing.GetManufacturingDashboard)
		}

		// ── Projects ───────────────────────────────────────────────────────────
		projects := api.Group("/projects")
		{
			// Dashboard & Reports
			projects.GET("/dashboard", h.Projects.GetProjectDashboard)
			projects.GET("/report", h.Projects.GetProjectsReport)

			// Projects CRUD
			projects.GET("", h.Projects.ListProjects)
			projects.POST("", h.Projects.CreateProject)
			projects.GET("/:id", h.Projects.GetProject)
			projects.PUT("/:id", h.Projects.UpdateProject)
			projects.DELETE("/:id", h.Projects.DeleteProject)
			projects.GET("/:id/costs", h.Projects.GetProjectCosts)

			// Tasks (cross-project + per-project)
			projects.GET("/tasks/all", h.Projects.ListAllTasks)
			projects.GET("/:id/tasks", h.Projects.ListTasks)
			projects.POST("/:id/tasks", h.Projects.CreateTask)
			projects.GET("/tasks/:taskId", h.Projects.GetTask)
			projects.PUT("/tasks/:taskId", h.Projects.UpdateTask)
			projects.DELETE("/tasks/:taskId", h.Projects.DeleteTask)

			// Milestones
			projects.GET("/:id/milestones", h.Projects.ListMilestones)
			projects.POST("/:id/milestones", h.Projects.CreateMilestone)
			projects.PUT("/milestones/:milestoneId", h.Projects.UpdateMilestone)
			projects.DELETE("/milestones/:milestoneId", h.Projects.DeleteMilestone)

			// Timesheets
			projects.GET("/timesheets", h.Projects.ListTimesheets)
			projects.POST("/timesheets", h.Projects.CreateTimesheet)
			projects.PUT("/timesheets/:timesheetId", h.Projects.UpdateTimesheet)
			projects.DELETE("/timesheets/:timesheetId", h.Projects.DeleteTimesheet)

			// Expenses
			projects.GET("/expenses", h.Projects.ListExpenses)
			projects.POST("/expenses", h.Projects.CreateExpense)
			projects.PUT("/expenses/:expenseId", h.Projects.UpdateExpense)
			projects.DELETE("/expenses/:expenseId", h.Projects.DeleteExpense)

			// Planning slots
			projects.GET("/planning", h.Projects.ListPlanningSlots)
			projects.POST("/planning", h.Projects.UpsertPlanningSlot)
			projects.DELETE("/planning/:slotId", h.Projects.DeletePlanningSlot)
		}

		// ── Treasury ───────────────────────────────────────────────────────────
		treasury := api.Group("/treasury")
		{
			// Cash Accounts
			treasury.GET("/cash-accounts", h.Treasury.ListCashAccounts)
			treasury.GET("/cash-accounts/:id", h.Treasury.GetCashAccount)
			treasury.POST("/cash-accounts", h.Treasury.CreateCashAccount)
			treasury.PUT("/cash-accounts/:id", h.Treasury.UpdateCashAccount)

			// Bank Accounts
			treasury.GET("/bank-accounts", h.Treasury.ListBankAccounts)
			treasury.POST("/bank-accounts", h.Treasury.CreateBankAccount)
			treasury.PUT("/bank-accounts/:id", h.Treasury.UpdateBankAccount)

			// Cheques
			treasury.GET("/cheques", h.Treasury.ListCheques)
			treasury.POST("/cheques", h.Treasury.CreateCheque)
			treasury.PUT("/cheques/:id", h.Treasury.UpdateCheque)
			treasury.PUT("/cheques/:id/deposit", h.Treasury.DepositCheque)
			treasury.PUT("/cheques/:id/bounce", h.Treasury.BounceCheque)
			treasury.PUT("/cheques/:id/cancel", h.Treasury.CancelCheque)

			// Movements
			treasury.GET("/movements", h.Treasury.ListMovements)
			treasury.POST("/movements", h.Treasury.CreateMovement)

			// Payments
			treasury.GET("/payments", h.Treasury.ListPayments)
			treasury.POST("/payments", h.Treasury.CreatePayment)
			treasury.PUT("/payments/:id", h.Treasury.UpdatePayment)
			treasury.PUT("/payments/:id/confirm", h.Treasury.ConfirmPayment)
			treasury.PUT("/payments/:id/allocate", h.Treasury.AllocatePayment)

			// Receipts
			treasury.GET("/receipts", h.Treasury.ListReceipts)
			treasury.POST("/receipts", h.Treasury.CreateReceipt)
			treasury.PUT("/receipts/:id", h.Treasury.UpdateReceipt)
			treasury.PUT("/receipts/:id/confirm", h.Treasury.ConfirmReceipt)
			treasury.PUT("/receipts/:id/cancel", h.Treasury.CancelReceipt)
			treasury.DELETE("/receipts/:id", h.Treasury.DeleteReceipt)

			// Bank Reconciliation
			treasury.GET("/reconciliations", h.Treasury.ListReconciliations)
			treasury.GET("/reconciliations/:id", h.Treasury.GetReconciliation)
			treasury.POST("/reconciliations", h.Treasury.CreateReconciliation)
			treasury.PUT("/reconciliations/:id", h.Treasury.UpdateReconciliation)
			treasury.PUT("/reconciliations/:id/complete", h.Treasury.CompleteReconciliation)
			treasury.POST("/reconciliations/:id/lines", h.Treasury.AddReconciliationLine)
			treasury.POST("/reconciliations/match-lines", h.Treasury.MatchReconciliationLines)

			// Reports
			treasury.GET("/reports/aging", h.Treasury.AgingReport)
			treasury.GET("/reports/cash-position", h.Treasury.CashPositionReport)
			treasury.GET("/reports/treasury", h.Treasury.TreasuryReport)
		}

		// ── Tax Compliance ─────────────────────────────────────────────────────
		tax := api.Group("/tax")
		{
			// Declarations (G50, IBS, etc.)
			tax.GET("/declarations", h.Tax.ListDeclarations)
			tax.GET("/declarations/:id", h.Tax.GetDeclaration)
			tax.POST("/declarations", h.Tax.CreateDeclaration)
			tax.PUT("/declarations/:id", h.Tax.UpdateDeclaration)
			tax.DELETE("/declarations/:id", h.Tax.DeleteDeclaration)
			tax.POST("/declarations/:id/submit", h.Tax.SubmitDeclaration)
			tax.POST("/declarations/:id/amend", h.Tax.AmendDeclaration)

			// G50 auto-compute
			tax.GET("/declarations/g50", h.Tax.GetG50)
			tax.POST("/declarations/g50", h.Tax.SubmitG50)

			// IBS
			tax.GET("/declarations/ibs", h.Tax.GetIBS)

			// VAT Register
			tax.GET("/vat-register", h.Tax.GetVATRegister)
			tax.POST("/vat-register", h.Tax.CreateVATEntry)

			// VAT Returns
			tax.GET("/vat-returns", h.Tax.ListVATReturns)
			tax.POST("/vat-returns", h.Tax.CreateVATReturn)
			tax.PUT("/vat-returns/:id", h.Tax.UpdateVATReturn)
			tax.POST("/vat-returns/:id/submit", h.Tax.SubmitVATReturn)
			tax.GET("/vat-returns/compute", h.Tax.ComputeVATReturn)

			// Tax Payments
			tax.GET("/payments", h.Tax.ListTaxPayments)
			tax.POST("/payments", h.Tax.CreateTaxPayment)
			tax.PUT("/payments/:id", h.Tax.UpdateTaxPayment)
			tax.DELETE("/payments/:id", h.Tax.DeleteTaxPayment)

			// Reports
			tax.GET("/reports", h.Tax.GetTaxReport)
			tax.GET("/rates", h.Tax.GetTaxRateConfig)
		}

		// ── Workflow & Approvals ───────────────────────────────────────────────
		workflow := api.Group("/workflow")
		{
			workflow.GET("/rules", h.Workflow.ListRules)
			workflow.POST("/rules", h.Workflow.CreateRule)
			workflow.PUT("/rules/:id", h.Workflow.UpdateRule)

			workflow.GET("/approvals/inbox", h.Workflow.GetApprovalInbox)
			workflow.PUT("/approvals/:id/approve", h.Workflow.Approve)
			workflow.PUT("/approvals/:id/reject", h.Workflow.Reject)
		}

		// ── Reports & BI ───────────────────────────────────────────────────────
		reports := api.Group("/reports")
		{
			reports.GET("/financial-ratios", h.Reports.FinancialRatios)
			reports.GET("/kpi-summary", h.Reports.KPISummary)
			reports.POST("/custom", h.Reports.RunCustomReport)
			reports.GET("/export/:type", h.Reports.Export)

			// Reports & BI extended
			reports.GET("/bi-dashboard", h.ReportsBI.BIDashboard)
			reports.GET("/financial", h.ReportsBI.FinancialReports)
			reports.GET("/sales", h.ReportsBI.SalesReports)
			reports.GET("/purchase", h.ReportsBI.PurchaseReports)
			reports.GET("/inventory", h.ReportsBI.InventoryReports)
			reports.GET("/projects", h.ReportsBI.ProjectReports)
			reports.GET("/management", h.ReportsBI.ManagementReports)
			reports.GET("/analytics", h.ReportsBI.Analytics)
			reports.GET("/definitions", h.ReportsBI.ListReportDefinitions)
		}

		// Diagnostics routes
		diagnostics := api.Group("/diagnostics")
		{
			diagnostics.GET("/logs", h.Diagnostics.ListLogs)
			diagnostics.POST("/logs", h.Diagnostics.CreateLog)
			diagnostics.GET("/logs/:id", h.Diagnostics.GetLog)
			diagnostics.POST("/logs/:id/resolve", h.Diagnostics.ResolveLog)
			diagnostics.POST("/logs/bulk-resolve", h.Diagnostics.BulkResolveLogs)
			diagnostics.DELETE("/logs/:id", h.Diagnostics.DeleteLog)
			diagnostics.GET("/stats", h.Diagnostics.GetStats)
			diagnostics.DELETE("/logs/purge", h.Diagnostics.PurgeLogs)
		}
	}
		// Maintenance routes
		maintenance := api.Group("/maintenance")
		{
			// Equipment
			maintenance.GET("/equipment", h.Maintenance.ListEquipment)
			maintenance.GET("/equipment/categories", h.Maintenance.GetEquipmentCategories)
			maintenance.GET("/equipment/:id", h.Maintenance.GetEquipment)
			maintenance.POST("/equipment", h.Maintenance.CreateEquipment)
			maintenance.PUT("/equipment/:id", h.Maintenance.UpdateEquipment)
			maintenance.DELETE("/equipment/:id", h.Maintenance.DeleteEquipment)

			// Requests
			maintenance.GET("/requests", h.Maintenance.ListRequests)
			maintenance.GET("/requests/:id", h.Maintenance.GetRequest)
			maintenance.POST("/requests", h.Maintenance.CreateRequest)
			maintenance.PUT("/requests/:id", h.Maintenance.UpdateRequest)
			maintenance.DELETE("/requests/:id", h.Maintenance.DeleteRequest)

			// Orders
			maintenance.GET("/orders", h.Maintenance.ListOrders)
			maintenance.GET("/orders/:id", h.Maintenance.GetOrder)
			maintenance.POST("/orders", h.Maintenance.CreateOrder)
			maintenance.PUT("/orders/:id", h.Maintenance.UpdateOrder)
			maintenance.PUT("/orders/:id/complete", h.Maintenance.CompleteOrder)
			maintenance.DELETE("/orders/:id", h.Maintenance.DeleteOrder)

			// Preventive Plans
			maintenance.GET("/preventive-plans", h.Maintenance.ListPreventivePlans)
			maintenance.POST("/preventive-plans", h.Maintenance.CreatePreventivePlan)
			maintenance.PUT("/preventive-plans/:id", h.Maintenance.UpdatePreventivePlan)
			maintenance.DELETE("/preventive-plans/:id", h.Maintenance.DeletePreventivePlan)

			// Calendar
			maintenance.GET("/calendar", h.Maintenance.GetCalendar)

			// History
			maintenance.GET("/history", h.Maintenance.ListHistory)
			maintenance.POST("/history", h.Maintenance.CreateHistory)

			// Dashboard & Reports
			maintenance.GET("/dashboard", h.Maintenance.GetDashboard)
			maintenance.GET("/reports", h.Maintenance.GetReports)
		}

		// Fleet routes
		fleet := api.Group("/fleet")
		{
			// Vehicles
			fleet.GET("/vehicles", h.Fleet.ListVehicles)
			fleet.GET("/vehicles/:id", h.Fleet.GetVehicle)
			fleet.POST("/vehicles", h.Fleet.CreateVehicle)
			fleet.PUT("/vehicles/:id", h.Fleet.UpdateVehicle)
			fleet.DELETE("/vehicles/:id", h.Fleet.DeleteVehicle)

			// Drivers
			fleet.GET("/drivers", h.Fleet.ListDrivers)
			fleet.POST("/drivers", h.Fleet.CreateDriver)
			fleet.PUT("/drivers/:id", h.Fleet.UpdateDriver)
			fleet.DELETE("/drivers/:id", h.Fleet.DeleteDriver)

			// Assignments
			fleet.GET("/assignments", h.Fleet.ListAssignments)
			fleet.POST("/assignments", h.Fleet.CreateAssignment)
			fleet.PUT("/assignments/:id", h.Fleet.UpdateAssignment)
			fleet.DELETE("/assignments/:id", h.Fleet.DeleteAssignment)

			// Fuel Logs
			fleet.GET("/fuel", h.Fleet.ListFuelLogs)
			fleet.POST("/fuel", h.Fleet.CreateFuelLog)
			fleet.PUT("/fuel/:id", h.Fleet.UpdateFuelLog)
			fleet.DELETE("/fuel/:id", h.Fleet.DeleteFuelLog)

			// Fleet Maintenance
			fleet.GET("/maintenance", h.Fleet.ListFleetMaintenance)
			fleet.POST("/maintenance", h.Fleet.CreateFleetMaintenance)
			fleet.PUT("/maintenance/:id", h.Fleet.UpdateFleetMaintenance)
			fleet.DELETE("/maintenance/:id", h.Fleet.DeleteFleetMaintenance)

			// Expenses
			fleet.GET("/expenses", h.Fleet.ListExpenses)
			fleet.POST("/expenses", h.Fleet.CreateExpense)
			fleet.PUT("/expenses/:id", h.Fleet.UpdateExpense)
			fleet.DELETE("/expenses/:id", h.Fleet.DeleteExpense)

			// Dashboard & Reports
			fleet.GET("/dashboard", h.Fleet.GetFleetDashboard)
			fleet.GET("/reports", h.Fleet.GetFleetReports)
		}

		// ── Quality ───────────────────────────────────────────────────────────
		quality := api.Group("/quality")
		{
			// Dashboard
			quality.GET("/dashboard", h.Quality.GetDashboard)

			// Control Plans
			quality.GET("/plans", h.Quality.ListPlans)
			quality.POST("/plans", h.Quality.CreatePlan)
			quality.PUT("/plans/:id", h.Quality.UpdatePlan)
			quality.DELETE("/plans/:id", h.Quality.DeletePlan)

			// Inspections
			quality.GET("/inspections", h.Quality.ListInspections)
			quality.GET("/inspections/:id", h.Quality.GetInspection)
			quality.POST("/inspections", h.Quality.CreateInspection)
			quality.PUT("/inspections/:id", h.Quality.UpdateInspection)
			quality.POST("/inspections/:id/start", h.Quality.StartInspection)
			quality.POST("/inspections/:id/complete", h.Quality.CompleteInspection)
			quality.DELETE("/inspections/:id", h.Quality.DeleteInspection)

			// Checks
			quality.GET("/checks", h.Quality.ListChecks)
			quality.POST("/checks", h.Quality.CreateCheck)
			quality.PUT("/checks/:id/result", h.Quality.RecordCheckResult)
			quality.DELETE("/checks/:id", h.Quality.DeleteCheck)

			// Non-Conformities
			quality.GET("/non-conformities", h.Quality.ListNonConformities)
			quality.GET("/non-conformities/:id", h.Quality.GetNonConformity)
			quality.POST("/non-conformities", h.Quality.CreateNonConformity)
			quality.PUT("/non-conformities/:id", h.Quality.UpdateNonConformity)
			quality.PUT("/non-conformities/:id/status", h.Quality.UpdateNCStatus)
			quality.DELETE("/non-conformities/:id", h.Quality.DeleteNonConformity)

			// Corrective Actions
			quality.GET("/corrective-actions", h.Quality.ListCorrectiveActions)
			quality.GET("/corrective-actions/:id", h.Quality.GetCorrectiveAction)
			quality.POST("/corrective-actions", h.Quality.CreateCorrectiveAction)
			quality.PUT("/corrective-actions/:id", h.Quality.UpdateCorrectiveAction)
			quality.PUT("/corrective-actions/:id/status", h.Quality.UpdateCAStatus)
			quality.DELETE("/corrective-actions/:id", h.Quality.DeleteCorrectiveAction)

			// Reports
			quality.GET("/reports", h.Quality.GetReports)
		}

		// ── Helpdesk / Support ────────────────────────────────────────────────
		helpdesk := api.Group("/helpdesk")
		{
			// Dashboard
			helpdesk.GET("/dashboard", h.Helpdesk.GetDashboard)

			// Tickets
			helpdesk.GET("/tickets", h.Helpdesk.ListTickets)
			helpdesk.GET("/tickets/:id", h.Helpdesk.GetTicket)
			helpdesk.POST("/tickets", h.Helpdesk.CreateTicket)
			helpdesk.PUT("/tickets/:id", h.Helpdesk.UpdateTicket)
			helpdesk.PUT("/tickets/:id/status", h.Helpdesk.UpdateTicketStatus)
			helpdesk.DELETE("/tickets/:id", h.Helpdesk.DeleteTicket)

			// Ticket Comments
			helpdesk.POST("/tickets/:id/comments", h.Helpdesk.AddComment)

			// Ticket Assignments
			helpdesk.GET("/assignments", h.Helpdesk.ListAssignments)
			helpdesk.POST("/tickets/:id/assign", h.Helpdesk.AssignTicket)

			// Categories
			helpdesk.GET("/categories", h.Helpdesk.ListCategories)
			helpdesk.POST("/categories", h.Helpdesk.CreateCategory)
			helpdesk.PUT("/categories/:id", h.Helpdesk.UpdateCategory)
			helpdesk.DELETE("/categories/:id", h.Helpdesk.DeleteCategory)

			// Agents
			helpdesk.GET("/agents", h.Helpdesk.ListAgents)
			helpdesk.POST("/agents", h.Helpdesk.CreateAgent)
			helpdesk.PUT("/agents/:id", h.Helpdesk.UpdateAgent)
			helpdesk.DELETE("/agents/:id", h.Helpdesk.DeleteAgent)

			// Escalations
			helpdesk.GET("/escalations", h.Helpdesk.ListEscalations)
			helpdesk.POST("/escalations", h.Helpdesk.CreateEscalation)
			helpdesk.PUT("/escalations/:id/status", h.Helpdesk.UpdateEscalationStatus)
			helpdesk.DELETE("/escalations/:id", h.Helpdesk.DeleteEscalation)

			// SLA Policies
			helpdesk.GET("/sla-policies", h.Helpdesk.ListSLAPolicies)
			helpdesk.POST("/sla-policies", h.Helpdesk.CreateSLAPolicy)
			helpdesk.PUT("/sla-policies/:id", h.Helpdesk.UpdateSLAPolicy)
			helpdesk.DELETE("/sla-policies/:id", h.Helpdesk.DeleteSLAPolicy)
			helpdesk.GET("/sla-tracking", h.Helpdesk.GetSLATracking)

			// CSAT Surveys
			helpdesk.GET("/csat", h.Helpdesk.ListCSAT)
			helpdesk.POST("/csat", h.Helpdesk.CreateCSAT)

			// Reports
			helpdesk.GET("/reports", h.Helpdesk.GetReports)
		}

		// ── Assets Management ─────────────────────────────────────────────────────
		assets := api.Group("/assets")
		{
			// Dashboard
			assets.GET("/dashboard", h.Assets.GetDashboard)

			// Fixed Assets
			assets.GET("/assets", h.Assets.ListAssets)
			assets.GET("/assets/:id", h.Assets.GetAsset)
			assets.POST("/assets", h.Assets.CreateAsset)
			assets.PUT("/assets/:id", h.Assets.UpdateAsset)
			assets.DELETE("/assets/:id", h.Assets.DeleteAsset)
			assets.POST("/assets/:id/dispose", h.Assets.DisposeAsset)

			// Categories
			assets.GET("/categories", h.Assets.ListCategories)
			assets.POST("/categories", h.Assets.CreateCategory)
			assets.PUT("/categories/:id", h.Assets.UpdateCategory)
			assets.DELETE("/categories/:id", h.Assets.DeleteCategory)

			// Locations
			assets.GET("/locations", h.Assets.ListLocations)
			assets.POST("/locations", h.Assets.CreateLocation)
			assets.PUT("/locations/:id", h.Assets.UpdateLocation)
			assets.DELETE("/locations/:id", h.Assets.DeleteLocation)

			// Transfers
			assets.GET("/transfers", h.Assets.ListTransfers)
			assets.POST("/transfers", h.Assets.CreateTransfer)
			assets.PUT("/transfers/:id/approve", h.Assets.ApproveTransfer)
			assets.PUT("/transfers/:id/complete", h.Assets.CompleteTransfer)
			assets.DELETE("/transfers/:id", h.Assets.DeleteTransfer)

			// Depreciation
			assets.GET("/depreciation", h.Assets.ListDepreciation)
			assets.POST("/depreciation/generate", h.Assets.GenerateDepreciation)
			assets.POST("/depreciation/post", h.Assets.PostDepreciation)

			// Maintenance
			assets.GET("/maintenance", h.Assets.ListMaintenance)
			assets.POST("/maintenance", h.Assets.CreateMaintenance)
			assets.PUT("/maintenance/:id", h.Assets.UpdateMaintenance)
			assets.PUT("/maintenance/:id/complete", h.Assets.CompleteMaintenance)
			assets.DELETE("/maintenance/:id", h.Assets.DeleteMaintenance)

			// Reports
			assets.GET("/reports", h.Assets.GetReports)
		}

		// ── Budgeting & Planning ──────────────────────────────────────
		budgeting := api.Group("/budgeting")
		{
			budgeting.GET("/dashboard", h.Budgeting.GetDashboard)

			// Categories
			budgeting.GET("/categories", h.Budgeting.ListBudgetCategories)
			budgeting.POST("/categories", h.Budgeting.CreateBudgetCategory)
			budgeting.PUT("/categories/:id", h.Budgeting.UpdateBudgetCategory)
			budgeting.DELETE("/categories/:id", h.Budgeting.DeleteBudgetCategory)

			// Annual Budgets
			budgeting.GET("/annual", h.Budgeting.ListAnnualBudgets)
			budgeting.GET("/annual/:id", h.Budgeting.GetAnnualBudget)
			budgeting.POST("/annual", h.Budgeting.CreateAnnualBudget)
			budgeting.PUT("/annual/:id", h.Budgeting.UpdateAnnualBudget)
			budgeting.DELETE("/annual/:id", h.Budgeting.DeleteAnnualBudget)
			budgeting.PUT("/annual/:id/approve", h.Budgeting.ApproveBudget)
			budgeting.PUT("/annual/:id/lock", h.Budgeting.LockBudget)

			// Line Items
			budgeting.GET("/line-items", h.Budgeting.ListLineItems)
			budgeting.POST("/line-items", h.Budgeting.CreateLineItem)
			budgeting.PUT("/line-items/:id", h.Budgeting.UpdateLineItem)
			budgeting.DELETE("/line-items/:id", h.Budgeting.DeleteLineItem)

			// Department Budgets
			budgeting.GET("/departments", h.Budgeting.ListDepartmentBudgets)
			budgeting.POST("/departments", h.Budgeting.CreateDepartmentBudget)
			budgeting.PUT("/departments/:id", h.Budgeting.UpdateDepartmentBudget)
			budgeting.DELETE("/departments/:id", h.Budgeting.DeleteDepartmentBudget)

			// Budget vs Actual
			budgeting.GET("/vs-actual", h.Budgeting.GetBudgetVsActual)

			// Revisions
			budgeting.GET("/revisions", h.Budgeting.ListRevisions)
			budgeting.POST("/revisions", h.Budgeting.CreateRevision)
			budgeting.PUT("/revisions/:id", h.Budgeting.UpdateRevision)
			budgeting.PUT("/revisions/:id/approve", h.Budgeting.ApproveRevision)
			budgeting.DELETE("/revisions/:id", h.Budgeting.DeleteRevision)

			// Commitments
			budgeting.GET("/commitments", h.Budgeting.ListCommitments)
			budgeting.POST("/commitments", h.Budgeting.CreateCommitment)
			budgeting.PUT("/commitments/:id", h.Budgeting.UpdateCommitment)
			budgeting.PUT("/commitments/:id/approve", h.Budgeting.ApproveCommitment)
			budgeting.PUT("/commitments/:id/fulfill", h.Budgeting.FulfillCommitment)
			budgeting.PUT("/commitments/:id/cancel", h.Budgeting.CancelCommitment)
			budgeting.DELETE("/commitments/:id", h.Budgeting.DeleteCommitment)

			// Actuals
			budgeting.GET("/actuals", h.Budgeting.ListActuals)
			budgeting.POST("/actuals", h.Budgeting.CreateActual)
			budgeting.POST("/actuals/post", h.Budgeting.PostActuals)

			// Reports
			budgeting.GET("/reports", h.Budgeting.GetReports)
		}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Nexus ERP server running on http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// must panics if err is not nil — helper for startup-time filesystem ops
func must(f fs.FS, err error) fs.FS {
	if err != nil {
		panic(err)
	}
	return f
}
