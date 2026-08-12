package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// ─── Base ──────────────────────────────────────────────────────────────────────

type Base struct {
	ID        string    `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy string    `json:"created_by,omitempty" db:"created_by"`
}

// ─── System / Auth ─────────────────────────────────────────────────────────────

type Tenant struct {
	Base
	Name     string `json:"name" db:"name"`
	Code     string `json:"code" db:"code"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

type Company struct {
	Base
	TenantID    string `json:"tenant_id" db:"tenant_id"`
	Name        string `json:"name" db:"name"`
	TaxID       string `json:"tax_id" db:"tax_id"`
	Address     string `json:"address" db:"address"`
	City        string `json:"city" db:"city"`
	Phone       string `json:"phone" db:"phone"`
	Email       string `json:"email" db:"email"`
	LogoURL     string `json:"logo_url" db:"logo_url"`
	Currency    string `json:"currency" db:"currency"`
	IsActive    bool   `json:"is_active" db:"is_active"`
}

type Branch struct {
	Base
	CompanyID string `json:"company_id" db:"company_id"`
	Name      string `json:"name" db:"name"`
	Code      string `json:"code" db:"code"`
	Address   string `json:"address" db:"address"`
	IsActive  bool   `json:"is_active" db:"is_active"`
}

type FiscalYear struct {
	Base
	CompanyID  string     `json:"company_id" db:"company_id"`
	Name       string     `json:"name" db:"name"`
	StartDate  time.Time  `json:"start_date" db:"start_date"`
	EndDate    time.Time  `json:"end_date" db:"end_date"`
	IsActive   bool       `json:"is_active" db:"is_active"`
	IsClosed   bool       `json:"is_closed" db:"is_closed"`
	ClosedAt   *time.Time `json:"closed_at,omitempty" db:"closed_at"`
}

type Role struct {
	Base
	Name        string   `json:"name" db:"name"`
	Description string   `json:"description" db:"description"`
	Permissions []string `json:"permissions" db:"permissions"`
}

type User struct {
	Base
	TenantID     string     `json:"tenant_id" db:"tenant_id"`
	CompanyID    string     `json:"company_id" db:"company_id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FullName     string     `json:"full_name" db:"full_name"`
	Role         string     `json:"role" db:"role"`
	BranchID     string     `json:"branch_id" db:"branch_id"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	AvatarURL    string     `json:"avatar_url,omitempty" db:"avatar_url"`
}

type AuditLog struct {
	ID        string      `json:"id" db:"id"`
	UserID    string      `json:"user_id" db:"user_id"`
	Username  string      `json:"username" db:"username"`
	Action    string      `json:"action" db:"action"` // CREATE, UPDATE, DELETE, LOGIN
	Module    string      `json:"module" db:"module"`
	RecordID  string      `json:"record_id" db:"record_id"`
	OldValues interface{} `json:"old_values,omitempty" db:"old_values"`
	NewValues interface{} `json:"new_values,omitempty" db:"new_values"`
	IPAddress string      `json:"ip_address" db:"ip_address"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

// ─── Accounting ────────────────────────────────────────────────────────────────

type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"
	AccountTypeLiability AccountType = "liability"
	AccountTypeEquity    AccountType = "equity"
	AccountTypeRevenue   AccountType = "revenue"
	AccountTypeExpense   AccountType = "expense"
	AccountTypeContra    AccountType = "contra"
)

// ChartOfAccount maps to chart_of_accounts table.
// Actual schema: is_group (NOT is_postable), currency (NOT currency_code), no name_ar column.
type ChartOfAccount struct {
	Base
	CompanyID      string           `json:"company_id" db:"company_id"`
	Code           string           `json:"code" db:"code"`
	Name           string           `json:"name" db:"name"`
	Type           AccountType      `json:"type" db:"type"`
	Nature         string           `json:"nature" db:"nature"` // debit | credit
	ParentID       *string          `json:"parent_id,omitempty" db:"parent_id"`
	IsGroup        bool             `json:"is_group" db:"is_group"`
	IsReconcilable bool             `json:"is_reconcilable" db:"is_reconcilable"`
	Currency       string           `json:"currency" db:"currency"` // default DZD
	Balance        decimal.Decimal  `json:"balance" db:"balance"`
	DebitBalance   decimal.Decimal  `json:"debit_balance" db:"debit_balance"`
	CreditBalance  decimal.Decimal  `json:"credit_balance" db:"credit_balance"`
	Level          int              `json:"level" db:"level"`
	Description    string           `json:"description" db:"description"`
	IsActive       bool             `json:"is_active" db:"is_active"`
	Children       []ChartOfAccount `json:"children,omitempty" db:"-"`
}

type JournalEntryStatus string

const (
	JEStatusDraft     JournalEntryStatus = "draft"
	JEStatusPosted    JournalEntryStatus = "posted"
	JEStatusCancelled JournalEntryStatus = "cancelled"
)

// JournalEntry maps to journal_entries table.
// Actual schema: reference (NOT journal), source_type (NOT source_module).
type JournalEntry struct {
	Base
	CompanyID    string             `json:"company_id" db:"company_id"`
	BranchID     *string            `json:"branch_id,omitempty" db:"branch_id"`
	FiscalYearID *string            `json:"fiscal_year_id,omitempty" db:"fiscal_year_id"`
	Number       string             `json:"number" db:"number"`
	Date         time.Time          `json:"date" db:"date"`
	Reference    string             `json:"reference" db:"reference"`
	Description  string             `json:"description" db:"description"`
	Status       JournalEntryStatus `json:"status" db:"status"`
	TotalDebit   decimal.Decimal    `json:"total_debit" db:"total_debit"`
	TotalCredit  decimal.Decimal    `json:"total_credit" db:"total_credit"`
	Currency     string             `json:"currency" db:"currency"`
	SourceType   string             `json:"source_type,omitempty" db:"source_type"`
	SourceID     *string            `json:"source_id,omitempty" db:"source_id"`
	CostCenterID *string            `json:"cost_center_id,omitempty" db:"cost_center_id"`
	Lines        []JournalLine      `json:"lines,omitempty" db:"-"`
}

// JournalLine maps to journal_lines table.
// Actual schema: journal_entry_id (NOT entry_id).
type JournalLine struct {
	ID             string          `json:"id" db:"id"`
	JournalEntryID string          `json:"journal_entry_id" db:"journal_entry_id"`
	AccountID      string          `json:"account_id" db:"account_id"`
	AccountCode    string          `json:"account_code" db:"account_code"`
	AccountName    string          `json:"account_name" db:"account_name"`
	CostCenterID   *string         `json:"cost_center_id,omitempty" db:"cost_center_id"`
	Description    string          `json:"description" db:"description"`
	Debit          decimal.Decimal `json:"debit" db:"debit"`
	Credit         decimal.Decimal `json:"credit" db:"credit"`
	Currency       string          `json:"currency" db:"currency"`
}

// CostCenter maps to cost_centers table.
// Actual schema: NO type column.
type CostCenter struct {
	Base
	CompanyID string  `json:"company_id" db:"company_id"`
	Code      string  `json:"code" db:"code"`
	Name      string  `json:"name" db:"name"`
	ParentID  *string `json:"parent_id,omitempty" db:"parent_id"`
	Budget    decimal.Decimal `json:"budget" db:"budget"`
	Actual    decimal.Decimal `json:"actual" db:"actual"`
	IsActive  bool    `json:"is_active" db:"is_active"`
}

// FixedAsset maps to fixed_assets table.
// Actual schema: purchase_date (NOT acquisition_date), purchase_value (NOT original_value),
// current_value (NOT book_value), depreciation_method enum: linear/diminishing_balance.
type FixedAsset struct {
	Base
	CompanyID               string          `json:"company_id" db:"company_id"`
	Code                    string          `json:"code" db:"code"`
	Name                    string          `json:"name" db:"name"`
	Category                string          `json:"category" db:"category"`
	PurchaseDate            time.Time       `json:"purchase_date" db:"purchase_date"`
	PurchaseValue           decimal.Decimal `json:"purchase_value" db:"purchase_value"`
	ResidualValue           decimal.Decimal `json:"residual_value" db:"residual_value"`
	CurrentValue            decimal.Decimal `json:"current_value" db:"current_value"`
	AccumulatedDepreciation decimal.Decimal `json:"accumulated_depreciation" db:"accumulated_depreciation"`
	DepreciationMethod      string          `json:"depreciation_method" db:"depreciation_method"` // linear | diminishing_balance
	UsefulLifeYears         int             `json:"useful_life_years" db:"useful_life_years"`
	DepreciationRate        decimal.Decimal `json:"depreciation_rate" db:"depreciation_rate"`
	Status                  string          `json:"status" db:"status"` // active | fully_depreciated | disposed | idle
	AssetAccountID          *string         `json:"asset_account_id,omitempty" db:"asset_account_id"`
	DepExpenseAccountID     *string         `json:"dep_expense_account_id,omitempty" db:"dep_expense_account_id"`
	AccDepAccountID         *string         `json:"acc_dep_account_id,omitempty" db:"acc_dep_account_id"`
}

// Budget maps to budgets header table.
// Lines are in budget_lines with per-month columns jan-dec.
type Budget struct {
	Base
	CompanyID    string          `json:"company_id" db:"company_id"`
	FiscalYearID *string         `json:"fiscal_year_id,omitempty" db:"fiscal_year_id"`
	Name         string          `json:"name" db:"name"`
	Description  string          `json:"description" db:"description"`
	Status       string          `json:"status" db:"status"`
	TotalBudget  decimal.Decimal `json:"total_budget" db:"total_budget"`
	TotalActual  decimal.Decimal `json:"total_actual" db:"total_actual"`
	Lines        []BudgetLine    `json:"lines,omitempty" db:"-"`
}

// BudgetLine maps to budget_lines table with per-month columns.
type BudgetLine struct {
	ID           string          `json:"id" db:"id"`
	BudgetID     string          `json:"budget_id" db:"budget_id"`
	AccountID    *string         `json:"account_id,omitempty" db:"account_id"`
	AccountCode  string          `json:"account_code" db:"account_code"`
	AccountName  string          `json:"account_name" db:"account_name"`
	CostCenterID *string         `json:"cost_center_id,omitempty" db:"cost_center_id"`
	Jan          decimal.Decimal `json:"jan" db:"jan"`
	Feb          decimal.Decimal `json:"feb" db:"feb"`
	Mar          decimal.Decimal `json:"mar" db:"mar"`
	Apr          decimal.Decimal `json:"apr" db:"apr"`
	May          decimal.Decimal `json:"may" db:"may"`
	Jun          decimal.Decimal `json:"jun" db:"jun"`
	Jul          decimal.Decimal `json:"jul" db:"jul"`
	Aug          decimal.Decimal `json:"aug" db:"aug"`
	Sep          decimal.Decimal `json:"sep" db:"sep"`
	Oct          decimal.Decimal `json:"oct" db:"oct"`
	Nov          decimal.Decimal `json:"nov" db:"nov"`
	Dec          decimal.Decimal `json:"dec" db:"dec"`
	TotalBudget  decimal.Decimal `json:"total_budget" db:"total_budget"`
	TotalActual  decimal.Decimal `json:"total_actual" db:"total_actual"`
}

// BankReconciliation maps to bank_reconciliations table.
// Actual schema: period_date (single DATE), statement_balance (NOT bank_balance),
// is_reconciled BOOLEAN (NOT status string).
type BankReconciliation struct {
	ID              string          `json:"id" db:"id"`
	BankAccountID   string          `json:"bank_account_id" db:"bank_account_id"`
	BankAccountName string          `json:"bank_account_name" db:"bank_account_name"`
	PeriodDate      time.Time       `json:"period_date" db:"period_date"`
	StatementBalance decimal.Decimal `json:"statement_balance" db:"statement_balance"`
	BookBalance     decimal.Decimal `json:"book_balance" db:"book_balance"`
	Difference      decimal.Decimal `json:"difference" db:"difference"`
	IsReconciled    bool            `json:"is_reconciled" db:"is_reconciled"`
	ReconciledAt    *time.Time      `json:"reconciled_at,omitempty" db:"reconciled_at"`
	Notes           string          `json:"notes" db:"notes"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

// ─── HR & Payroll ──────────────────────────────────────────────────────────────

type Employee struct {
	Base
	CompanyID      string          `json:"company_id" db:"company_id"`
	EmployeeNumber string          `json:"employee_number" db:"employee_number"`
	FirstName      string          `json:"first_name" db:"first_name"`
	LastName       string          `json:"last_name" db:"last_name"`
	FullName       string          `json:"full_name" db:"full_name"`
	Email          string          `json:"email" db:"email"`
	Phone          string          `json:"phone" db:"phone"`
	NationalID     string          `json:"national_id" db:"national_id"`
	CNASNumber     string          `json:"cnas_number" db:"cnas_number"`
	BirthDate      time.Time       `json:"birth_date" db:"birth_date"`
	HireDate       time.Time       `json:"hire_date" db:"hire_date"`
	DepartmentID   string          `json:"department_id" db:"department_id"`
	PositionID     string          `json:"position_id" db:"position_id"`
	ManagerID      *string         `json:"manager_id,omitempty" db:"manager_id"`
	ContractType   string          `json:"contract_type" db:"contract_type"`
	BasicSalary    decimal.Decimal `json:"basic_salary" db:"basic_salary"`
	BankName       string          `json:"bank_name" db:"bank_name"`
	RIB            string          `json:"rib" db:"rib"`
	IsActive       bool            `json:"is_active" db:"is_active"`
	ProfilePicture string          `json:"profile_picture" db:"profile_picture"`
}

type Department struct {
	Base
	CompanyID string  `json:"company_id" db:"company_id"`
	Name      string  `json:"name" db:"name"`
	Code      string  `json:"code" db:"code"`
	ManagerID *string `json:"manager_id,omitempty" db:"manager_id"`
	ParentID  *string `json:"parent_id,omitempty" db:"parent_id"`
}

type Position struct {
	Base
	CompanyID    string `json:"company_id" db:"company_id"`
	Title        string `json:"title" db:"title"`
	DepartmentID string `json:"department_id" db:"department_id"`
	Level        int    `json:"level" db:"level"`
}

type AttendanceRecord struct {
	Base
	EmployeeID    string     `json:"employee_id" db:"employee_id"`
	Date          time.Time  `json:"date" db:"date"`
	CheckIn       *time.Time `json:"check_in,omitempty" db:"check_in"`
	CheckOut      *time.Time `json:"check_out,omitempty" db:"check_out"`
	WorkHours     float64    `json:"work_hours" db:"work_hours"`
	OvertimeHours float64    `json:"overtime_hours" db:"overtime_hours"`
	Status        string     `json:"status" db:"status"` // present, absent, late, half_day
	Notes         string     `json:"notes" db:"notes"`
}

type LeaveRequest struct {
	Base
	EmployeeID string     `json:"employee_id" db:"employee_id"`
	LeaveType  string     `json:"leave_type" db:"leave_type"` // annual, sick, exceptional, maternity
	StartDate  time.Time  `json:"start_date" db:"start_date"`
	EndDate    time.Time  `json:"end_date" db:"end_date"`
	Days       int        `json:"days" db:"days"`
	Reason     string     `json:"reason" db:"reason"`
	Status     string     `json:"status" db:"status"` // pending, approved, rejected
	ApprovedBy *string    `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt *time.Time `json:"approved_at,omitempty" db:"approved_at"`
}

type PayrollRun struct {
	Base
	CompanyID      string          `json:"company_id" db:"company_id"`
	Month          int             `json:"month" db:"month"`
	Year           int             `json:"year" db:"year"`
	Status         string          `json:"status" db:"status"` // draft, validated, paid
	TotalGross     decimal.Decimal `json:"total_gross" db:"total_gross"`
	TotalCNAS      decimal.Decimal `json:"total_cnas" db:"total_cnas"`
	TotalIRG       decimal.Decimal `json:"total_irg" db:"total_irg"`
	TotalNet       decimal.Decimal `json:"total_net" db:"total_net"`
	JournalEntryID *string         `json:"journal_entry_id,omitempty" db:"journal_entry_id"`
}

type Payslip struct {
	Base
	PayrollRunID       string          `json:"payroll_run_id" db:"payroll_run_id"`
	EmployeeID         string          `json:"employee_id" db:"employee_id"`
	BasicSalary        decimal.Decimal `json:"basic_salary" db:"basic_salary"`
	TransportAllowance decimal.Decimal `json:"transport_allowance" db:"transport_allowance"`
	PerformanceBonus   decimal.Decimal `json:"performance_bonus" db:"performance_bonus"`
	SeniorityBonus     decimal.Decimal `json:"seniority_bonus" db:"seniority_bonus"`
	GrossSalary        decimal.Decimal `json:"gross_salary" db:"gross_salary"`
	CNASEmployee       decimal.Decimal `json:"cnas_employee" db:"cnas_employee"` // 9%
	CNASEmployer       decimal.Decimal `json:"cnas_employer" db:"cnas_employer"` // 26%
	IRG                decimal.Decimal `json:"irg" db:"irg"`
	OtherDeductions    decimal.Decimal `json:"other_deductions" db:"other_deductions"`
	NetSalary          decimal.Decimal `json:"net_salary" db:"net_salary"`
	WorkingDays        int             `json:"working_days" db:"working_days"`
	AbsentDays         int             `json:"absent_days" db:"absent_days"`
	OvertimeHours      float64         `json:"overtime_hours" db:"overtime_hours"`
}

// ─── Sales & CRM ───────────────────────────────────────────────────────────────

// Lead maps exactly to the `leads` table in 0001_init_schema.sql
type Lead struct {
	Base
	CompanyID     string  `json:"company_id" db:"company_id"`
	Title         string  `json:"title" db:"title"`
	FirstName     string  `json:"first_name" db:"first_name"`
	LastName      string  `json:"last_name" db:"last_name"`
	CompanyName   string  `json:"company_name" db:"company_name"`
	Email         string  `json:"email" db:"email"`
	Phone         string  `json:"phone" db:"phone"`
	Source        string  `json:"source" db:"source"`
	Status        string  `json:"status" db:"status"` // new, contacted, qualified, lost
	SalespersonID *string `json:"salesperson_id,omitempty" db:"salesperson_id"`
	Notes         string  `json:"notes" db:"notes"`
	ConvertedTo   *string `json:"converted_to,omitempty" db:"converted_to"`
}

// Opportunity maps exactly to the `opportunities` table
type Opportunity struct {
	Base
	CompanyID     string          `json:"company_id" db:"company_id"`
	CustomerID    *string         `json:"customer_id,omitempty" db:"customer_id"`
	CustomerName  string          `json:"customer_name,omitempty" db:"-"`
	LeadID        *string         `json:"lead_id,omitempty" db:"lead_id"`
	Name          string          `json:"name" db:"name"`
	Stage         string          `json:"stage" db:"stage"` // lead,qualified,proposal,negotiation,won,lost
	Amount        decimal.Decimal `json:"amount" db:"amount"`
	Probability   int             `json:"probability" db:"probability"` // 0-100
	ExpectedClose *time.Time      `json:"expected_close,omitempty" db:"expected_close"`
	SalespersonID *string         `json:"salesperson_id,omitempty" db:"salesperson_id"`
	Notes         string          `json:"notes" db:"notes"`
	LostReason    string          `json:"lost_reason" db:"lost_reason"`
}

// Customer maps exactly to the `customers` table
type Customer struct {
	Base
	CompanyID     string          `json:"company_id" db:"company_id"`
	Code          string          `json:"code" db:"code"`
	Name          string          `json:"name" db:"name"`
	Type          string          `json:"type" db:"type"` // company/individual
	NIF           string          `json:"nif" db:"nif"`
	NIS           string          `json:"nis" db:"nis"`
	RC            string          `json:"rc" db:"rc"`
	ART           string          `json:"art" db:"art"`
	TaxRegime     string          `json:"tax_regime" db:"tax_regime"` // reel/forfait/exonere
	Address       string          `json:"address" db:"address"`
	City          string          `json:"city" db:"city"`
	Wilaya        string          `json:"wilaya" db:"wilaya"`
	PostalCode    string          `json:"postal_code" db:"postal_code"`
	Phone         string          `json:"phone" db:"phone"`
	Email         string          `json:"email" db:"email"`
	Website       string          `json:"website" db:"website"`
	CreditLimit   decimal.Decimal `json:"credit_limit" db:"credit_limit"`
	Balance       decimal.Decimal `json:"balance" db:"balance"`
	PaymentTerms  int             `json:"payment_terms" db:"payment_terms"` // days
	AccountID     *string         `json:"account_id,omitempty" db:"account_id"`
	SalespersonID *string         `json:"salesperson_id,omitempty" db:"salesperson_id"`
	IsActive      bool            `json:"is_active" db:"is_active"`
	Notes         string          `json:"notes" db:"notes"`
}

// Quotation maps exactly to the `quotations` table
type Quotation struct {
	Base
	CompanyID      string          `json:"company_id" db:"company_id"`
	BranchID       *string         `json:"branch_id,omitempty" db:"branch_id"`
	Number         string          `json:"number" db:"number"`
	CustomerID     string          `json:"customer_id" db:"customer_id"`
	CustomerName   string          `json:"customer_name" db:"-"`
	Date           time.Time       `json:"date" db:"date"`
	ValidUntil     *time.Time      `json:"valid_until,omitempty" db:"valid_until"`
	Status         string          `json:"status" db:"status"` // draft,sent,confirmed,cancelled,expired
	Subtotal       decimal.Decimal `json:"subtotal" db:"subtotal"`
	DiscountAmount decimal.Decimal `json:"discount_amount" db:"discount_amount"`
	TVAAmount      decimal.Decimal `json:"tva_amount" db:"tva_amount"`
	StampTax       decimal.Decimal `json:"stamp_tax" db:"stamp_tax"`
	TotalAmount    decimal.Decimal `json:"total_amount" db:"total_amount"`
	Currency       string          `json:"currency" db:"currency"`
	Notes          string          `json:"notes" db:"notes"`
	Terms          string          `json:"terms" db:"terms"`
	SalespersonID  *string         `json:"salesperson_id,omitempty" db:"salesperson_id"`
	ConvertedTo    *string         `json:"converted_to,omitempty" db:"converted_to"` // sales_order UUID
	Lines          []DocumentLine  `json:"lines,omitempty" db:"-"`
}

// SalesOrder maps exactly to the `sales_orders` table
type SalesOrder struct {
	Base
	CompanyID      string          `json:"company_id" db:"company_id"`
	BranchID       *string         `json:"branch_id,omitempty" db:"branch_id"`
	Number         string          `json:"number" db:"number"`
	QuotationID    *string         `json:"quotation_id,omitempty" db:"quotation_id"`
	CustomerID     string          `json:"customer_id" db:"customer_id"`
	CustomerName   string          `json:"customer_name" db:"-"`
	Date           time.Time       `json:"date" db:"date"`
	DeliveryDate   *time.Time      `json:"delivery_date,omitempty" db:"delivery_date"`
	Status         string          `json:"status" db:"status"` // draft,confirmed,delivered,cancelled
	Subtotal       decimal.Decimal `json:"subtotal" db:"subtotal"`
	DiscountAmount decimal.Decimal `json:"discount_amount" db:"discount_amount"`
	TVAAmount      decimal.Decimal `json:"tva_amount" db:"tva_amount"`
	StampTax       decimal.Decimal `json:"stamp_tax" db:"stamp_tax"`
	TotalAmount    decimal.Decimal `json:"total_amount" db:"total_amount"`
	Currency       string          `json:"currency" db:"currency"`
	Notes          string          `json:"notes" db:"notes"`
	Lines          []DocumentLine  `json:"lines,omitempty" db:"-"`
}

// DocumentLine is used for quotation_lines, sales_order_lines, and sales_invoice_lines
// The parent FK (invoice_id / quotation_id / order_id) is stored separately per table
type DocumentLine struct {
	ID          string          `json:"id" db:"id"`
	ParentID    string          `json:"parent_id" db:"-"` // set at runtime
	ItemID      *string         `json:"item_id,omitempty" db:"item_id"`
	Description string          `json:"description" db:"description"`
	Quantity    decimal.Decimal `json:"quantity" db:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price" db:"unit_price"`
	DiscountPct decimal.Decimal `json:"discount_pct" db:"discount_pct"`
	TVARate     decimal.Decimal `json:"tva_rate" db:"tva_rate"`
	Subtotal    decimal.Decimal `json:"subtotal" db:"subtotal"`
	TVAAmount   decimal.Decimal `json:"tva_amount" db:"tva_amount"`
	Total       decimal.Decimal `json:"total" db:"total"`
	AccountID   *string         `json:"account_id,omitempty" db:"account_id"`
	SortOrder   int             `json:"sort_order" db:"sort_order"`
}

type InvoiceStatus string

const (
	InvoiceStatusDraft          InvoiceStatus = "draft"
	InvoiceStatusConfirmed      InvoiceStatus = "confirmed"
	InvoiceStatusPartiallyPaid  InvoiceStatus = "partially_paid"
	InvoiceStatusPaid           InvoiceStatus = "paid"
	InvoiceStatusCancelled      InvoiceStatus = "cancelled"
	InvoiceStatusOverdue        InvoiceStatus = "overdue"
)

// SalesInvoice maps exactly to the `sales_invoices` table
type SalesInvoice struct {
	Base
	CompanyID      string          `json:"company_id" db:"company_id"`
	BranchID       *string         `json:"branch_id,omitempty" db:"branch_id"`
	Number         string          `json:"number" db:"number"`
	OrderID        *string         `json:"order_id,omitempty" db:"order_id"`
	CustomerID     string          `json:"customer_id" db:"customer_id"`
	CustomerName   string          `json:"customer_name" db:"-"`
	Date           time.Time       `json:"date" db:"date"`
	DueDate        *time.Time      `json:"due_date,omitempty" db:"due_date"`
	Status         InvoiceStatus   `json:"status" db:"status"`
	Subtotal       decimal.Decimal `json:"subtotal" db:"subtotal"`
	DiscountAmount decimal.Decimal `json:"discount_amount" db:"discount_amount"`
	TVAAmount      decimal.Decimal `json:"tva_amount" db:"tva_amount"`
	StampTax       decimal.Decimal `json:"stamp_tax" db:"stamp_tax"`
	TotalAmount    decimal.Decimal `json:"total_amount" db:"total_amount"`
	PaidAmount     decimal.Decimal `json:"paid_amount" db:"paid_amount"`
	BalanceDue     decimal.Decimal `json:"balance_due" db:"balance_due"`
	Currency       string          `json:"currency" db:"currency"`
	Notes          string          `json:"notes" db:"notes"`
	JournalEntryID *string         `json:"journal_entry_id,omitempty" db:"journal_entry_id"`
	Lines          []DocumentLine  `json:"lines,omitempty" db:"-"`
}

// SalesInvoiceLine (kept for backward compat in GetInvoice handler)
type SalesInvoiceLine struct {
	ID          string          `json:"id" db:"id"`
	InvoiceID   string          `json:"invoice_id" db:"invoice_id"`
	ItemID      *string         `json:"item_id,omitempty" db:"item_id"`
	Description string          `json:"description" db:"description"`
	Quantity    decimal.Decimal `json:"quantity" db:"quantity"`
	UnitPrice   decimal.Decimal `json:"unit_price" db:"unit_price"`
	DiscountPct decimal.Decimal `json:"discount_pct" db:"discount_pct"`
	TVARate     decimal.Decimal `json:"tva_rate" db:"tva_rate"`
	Subtotal    decimal.Decimal `json:"subtotal" db:"subtotal"`
	TVAAmount   decimal.Decimal `json:"tva_amount" db:"tva_amount"`
	Total       decimal.Decimal `json:"total" db:"total"`
	AccountID   *string         `json:"account_id,omitempty" db:"account_id"`
	SortOrder   int             `json:"sort_order" db:"sort_order"`
}

// CustomerAging aggregates open invoice balances by age bucket
type CustomerAging struct {
	CustomerID       string          `json:"customer_id"`
	CustomerName     string          `json:"customer_name"`
	Phone            string          `json:"phone"`
	Email            string          `json:"email"`
	InvoiceCount     int             `json:"invoice_count"`
	CurrentAmount    decimal.Decimal `json:"current_amount"`
	Days1to30        decimal.Decimal `json:"days_1_30"`
	Days31to60       decimal.Decimal `json:"days_31_60"`
	Days61to90       decimal.Decimal `json:"days_61_90"`
	DaysOver90       decimal.Decimal `json:"days_over_90"`
	TotalOutstanding decimal.Decimal `json:"total_outstanding"`
}

// DashboardSummary aggregates KPIs for the main dashboard
type DashboardSummary struct {
	MonthlySales      decimal.Decimal `json:"monthly_sales"`
	Receivables       decimal.Decimal `json:"receivables"`
	OverdueInvoices   int             `json:"overdue_invoices"`
	CustomerCount     int             `json:"customer_count"`
	OpenOpportunities int             `json:"open_opportunities"`
	PipelineValue     decimal.Decimal `json:"pipeline_value"`
	WonThisMonth      decimal.Decimal `json:"won_this_month"`
	DraftQuotations   int             `json:"draft_quotations"`
	OpenOrders        int             `json:"open_orders"`
	// From other modules (fetched separately and merged)
	TreasuryBalance  decimal.Decimal `json:"treasury_balance"`
	StockValue       decimal.Decimal `json:"stock_value"`
	Payables         decimal.Decimal `json:"payables"`
	EmployeeCount    int             `json:"employee_count"`
	MonthlyPayroll   decimal.Decimal `json:"monthly_payroll"`
	ActiveProjects   int             `json:"active_projects"`
	ActiveMfgOrders  int             `json:"active_mfg_orders"`
	PendingApprovals int             `json:"pending_approvals"`
}

// PipelineSummary aggregates the CRM pipeline by stage
type PipelineSummary struct {
	Stage           string          `json:"stage"`
	Count           int             `json:"count"`
	TotalAmount     decimal.Decimal `json:"total_amount"`
	AvgProbability  int             `json:"avg_probability"`
}

// RecentActivity for the dashboard feed
type RecentActivity struct {
	ID         int64     `json:"id"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	UserName   string    `json:"user_name"`
	Module     string    `json:"module"`
	CreatedAt  time.Time `json:"created_at"`
}

// ─── Purchase ──────────────────────────────────────────────────────────────────

type Supplier struct {
	Base
	CompanyID    string          `json:"company_id" db:"company_id"`
	Code         string          `json:"code" db:"code"`
	Name         string          `json:"name" db:"name"`
	TaxID        string          `json:"tax_id" db:"tax_id"`
	Address      string          `json:"address" db:"address"`
	City         string          `json:"city" db:"city"`
	Phone        string          `json:"phone" db:"phone"`
	Email        string          `json:"email" db:"email"`
	ContactName  string          `json:"contact_name" db:"contact_name"`
	PaymentTerms int             `json:"payment_terms" db:"payment_terms"`
	Balance      decimal.Decimal `json:"balance" db:"balance"`
	AccountID    string          `json:"account_id" db:"account_id"`
	IsActive     bool            `json:"is_active" db:"is_active"`
	Rating       int             `json:"rating" db:"rating"` // 1-5
}

type PurchaseOrder struct {
	Base
	CompanyID    string          `json:"company_id" db:"company_id"`
	Number       string          `json:"number" db:"number"`
	SupplierID   string          `json:"supplier_id" db:"supplier_id"`
	SupplierName string          `json:"supplier_name" db:"supplier_name"`
	Date         time.Time       `json:"date" db:"date"`
	ExpectedDate time.Time       `json:"expected_date" db:"expected_date"`
	Status       string          `json:"status" db:"status"`
	SubTotal     decimal.Decimal `json:"sub_total" db:"sub_total"`
	TVAAmount    decimal.Decimal `json:"tva_amount" db:"tva_amount"`
	TotalAmount  decimal.Decimal `json:"total_amount" db:"total_amount"`
	Notes        string          `json:"notes" db:"notes"`
	Lines        []PurchaseOrderLine `json:"lines,omitempty" db:"-"`
}

type PurchaseOrderLine struct {
	ID              string          `json:"id" db:"id"`
	PurchaseOrderID string          `json:"purchase_order_id" db:"purchase_order_id"`
	ItemID          string          `json:"item_id" db:"item_id"`
	ItemCode        string          `json:"item_code" db:"item_code"`
	ItemName        string          `json:"item_name" db:"item_name"`
	Quantity        decimal.Decimal `json:"quantity" db:"quantity"`
	ReceivedQty     decimal.Decimal `json:"received_qty" db:"received_qty"`
	UnitPrice       decimal.Decimal `json:"unit_price" db:"unit_price"`
	TVARate         decimal.Decimal `json:"tva_rate" db:"tva_rate"`
	SubTotal        decimal.Decimal `json:"sub_total" db:"sub_total"`
	TotalAmount     decimal.Decimal `json:"total_amount" db:"total_amount"`
}

// ─── Inventory ─────────────────────────────────────────────────────────────────

type Item struct {
	Base
	CompanyID     string          `json:"company_id" db:"company_id"`
	Code          string          `json:"code" db:"code"`
	Name          string          `json:"name" db:"name"`
	Description   string          `json:"description" db:"description"`
	CategoryID    string          `json:"category_id" db:"category_id"`
	CategoryName  string          `json:"category_name" db:"category_name"`
	UnitID        string          `json:"unit_id" db:"unit_id"`
	UnitName      string          `json:"unit_name" db:"unit_name"`
	Type          string          `json:"type" db:"type"` // product, service, consumable
	SalePrice     decimal.Decimal `json:"sale_price" db:"sale_price"`
	PurchasePrice decimal.Decimal `json:"purchase_price" db:"purchase_price"`
	CMUP          decimal.Decimal `json:"cmup" db:"cmup"`
	TVARateS      decimal.Decimal `json:"tva_rate_sale" db:"tva_rate_sale"`
	TVARateP      decimal.Decimal `json:"tva_rate_purchase" db:"tva_rate_purchase"`
	ReorderPoint  decimal.Decimal `json:"reorder_point" db:"reorder_point"`
	MinStock      decimal.Decimal `json:"min_stock" db:"min_stock"`
	MaxStock      decimal.Decimal `json:"max_stock" db:"max_stock"`
	IsActive      bool            `json:"is_active" db:"is_active"`
	TrackLots     bool            `json:"track_lots" db:"track_lots"`
	TrackSerial   bool            `json:"track_serial" db:"track_serial"`
	SalesAccountID string         `json:"sales_account_id" db:"sales_account_id"`
	StockAccountID string         `json:"stock_account_id" db:"stock_account_id"`
	COGSAccountID  string         `json:"cogs_account_id" db:"cogs_account_id"`
}

type Warehouse struct {
	Base
	CompanyID string `json:"company_id" db:"company_id"`
	Name      string `json:"name" db:"name"`
	Code      string `json:"code" db:"code"`
	Address   string `json:"address" db:"address"`
	IsActive  bool   `json:"is_active" db:"is_active"`
}

type StockMovement struct {
	Base
	CompanyID    string          `json:"company_id" db:"company_id"`
	ItemID       string          `json:"item_id" db:"item_id"`
	ItemCode     string          `json:"item_code" db:"item_code"`
	ItemName     string          `json:"item_name" db:"item_name"`
	WarehouseID  string          `json:"warehouse_id" db:"warehouse_id"`
	Type         string          `json:"type" db:"type"` // in, out, transfer, adjustment
	Quantity     decimal.Decimal `json:"quantity" db:"quantity"`
	UnitCost     decimal.Decimal `json:"unit_cost" db:"unit_cost"`
	TotalCost    decimal.Decimal `json:"total_cost" db:"total_cost"`
	LotID        *string         `json:"lot_id,omitempty" db:"lot_id"`
	SourceModule string          `json:"source_module" db:"source_module"`
	SourceID     string          `json:"source_id" db:"source_id"`
	Notes        string          `json:"notes" db:"notes"`
}

// ─── Manufacturing ─────────────────────────────────────────────────────────────

type BOM struct {
	Base
	CompanyID   string          `json:"company_id" db:"company_id"`
	ProductID   string          `json:"product_id" db:"product_id"`
	ProductName string          `json:"product_name" db:"product_name"`
	Quantity    decimal.Decimal `json:"quantity" db:"quantity"`
	Version     string          `json:"version" db:"version"`
	IsActive    bool            `json:"is_active" db:"is_active"`
	Lines       []BOMLine       `json:"lines,omitempty" db:"-"`
}

type BOMLine struct {
	ID            string          `json:"id" db:"id"`
	BOMID         string          `json:"bom_id" db:"bom_id"`
	ComponentID   string          `json:"component_id" db:"component_id"`
	ComponentName string          `json:"component_name" db:"component_name"`
	Quantity      decimal.Decimal `json:"quantity" db:"quantity"`
	UnitID        string          `json:"unit_id" db:"unit_id"`
	Scrap         decimal.Decimal `json:"scrap" db:"scrap"`
}

type ManufacturingOrder struct {
	Base
	CompanyID     string          `json:"company_id" db:"company_id"`
	Number        string          `json:"number" db:"number"`
	ProductID     string          `json:"product_id" db:"product_id"`
	ProductName   string          `json:"product_name" db:"product_name"`
	BOMID         string          `json:"bom_id" db:"bom_id"`
	WorkCenterID  string          `json:"work_center_id" db:"work_center_id"`
	PlannedQty    decimal.Decimal `json:"planned_qty" db:"planned_qty"`
	ProducedQty   decimal.Decimal `json:"produced_qty" db:"produced_qty"`
	ScheduledDate time.Time       `json:"scheduled_date" db:"scheduled_date"`
	Status        string          `json:"status" db:"status"`
	ActualCost    decimal.Decimal `json:"actual_cost" db:"actual_cost"`
	Notes         string          `json:"notes" db:"notes"`
}

// ─── Projects ──────────────────────────────────────────────────────────────────

type Project struct {
	Base
	CompanyID   string          `json:"company_id" db:"company_id"`
	Code        string          `json:"code" db:"code"`
	Name        string          `json:"name" db:"name"`
	CustomerID  *string         `json:"customer_id,omitempty" db:"customer_id"`
	ManagerID   string          `json:"manager_id" db:"manager_id"`
	StartDate   time.Time       `json:"start_date" db:"start_date"`
	EndDate     time.Time       `json:"end_date" db:"end_date"`
	Status      string          `json:"status" db:"status"`
	Budget      decimal.Decimal `json:"budget" db:"budget"`
	ActualCost  decimal.Decimal `json:"actual_cost" db:"actual_cost"`
	Progress    int             `json:"progress" db:"progress"` // 0-100
	Description string          `json:"description" db:"description"`
}

type ProjectTask struct {
	Base
	ProjectID   string    `json:"project_id" db:"project_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	AssigneeID  *string   `json:"assignee_id,omitempty" db:"assignee_id"`
	Status      string    `json:"status" db:"status"`
	Priority    string    `json:"priority" db:"priority"`
	StartDate   time.Time `json:"start_date" db:"start_date"`
	DueDate     time.Time `json:"due_date" db:"due_date"`
	Progress    int       `json:"progress" db:"progress"`
	EstimatedH  float64   `json:"estimated_hours" db:"estimated_hours"`
	ActualH     float64   `json:"actual_hours" db:"actual_hours"`
}

type Timesheet struct {
	Base
	ProjectID   string    `json:"project_id" db:"project_id"`
	TaskID      *string   `json:"task_id,omitempty" db:"task_id"`
	EmployeeID  string    `json:"employee_id" db:"employee_id"`
	Date        time.Time `json:"date" db:"date"`
	Hours       float64   `json:"hours" db:"hours"`
	Description string    `json:"description" db:"description"`
	IsBillable  bool      `json:"is_billable" db:"is_billable"`
	HourlyRate  float64   `json:"hourly_rate" db:"hourly_rate"`
}

// ─── Treasury ──────────────────────────────────────────────────────────────────

type CashAccount struct {
	Base
	CompanyID string          `json:"company_id" db:"company_id"`
	BranchID  string          `json:"branch_id" db:"branch_id"`
	Name      string          `json:"name" db:"name"`
	Currency  string          `json:"currency" db:"currency"`
	Balance   decimal.Decimal `json:"balance" db:"balance"`
	AccountID string          `json:"account_id" db:"account_id"`
	IsActive  bool            `json:"is_active" db:"is_active"`
}

type BankAccount struct {
	Base
	CompanyID     string          `json:"company_id" db:"company_id"`
	BankName      string          `json:"bank_name" db:"bank_name"`
	AccountNumber string          `json:"account_number" db:"account_number"`
	RIB           string          `json:"rib" db:"rib"`
	IBAN          string          `json:"iban" db:"iban"`
	Currency      string          `json:"currency" db:"currency"`
	Balance       decimal.Decimal `json:"balance" db:"balance"`
	AccountID     string          `json:"account_id" db:"account_id"`
	IsActive      bool            `json:"is_active" db:"is_active"`
}

type Cheque struct {
	Base
	CompanyID     string          `json:"company_id" db:"company_id"`
	Number        string          `json:"number" db:"number"`
	Type          string          `json:"type" db:"type"` // received, issued
	BankAccountID string          `json:"bank_account_id" db:"bank_account_id"`
	PartnerID     string          `json:"partner_id" db:"partner_id"`
	PartnerName   string          `json:"partner_name" db:"partner_name"`
	Amount        decimal.Decimal `json:"amount" db:"amount"`
	IssueDate     time.Time       `json:"issue_date" db:"issue_date"`
	DueDate       time.Time       `json:"due_date" db:"due_date"`
	Status        string          `json:"status" db:"status"`
	PaymentID     *string         `json:"payment_id,omitempty" db:"payment_id"`
}

type Payment struct {
	Base
	CompanyID      string          `json:"company_id" db:"company_id"`
	Type           string          `json:"type" db:"type"` // customer, supplier
	PartnerID      string          `json:"partner_id" db:"partner_id"`
	PartnerName    string          `json:"partner_name" db:"partner_name"`
	Date           time.Time       `json:"date" db:"date"`
	Amount         decimal.Decimal `json:"amount" db:"amount"`
	AllocatedAmt   decimal.Decimal `json:"allocated_amount" db:"allocated_amount"`
	Method         string          `json:"method" db:"method"`
	Reference      string          `json:"reference" db:"reference"`
	Notes          string          `json:"notes" db:"notes"`
	JournalEntryID *string         `json:"journal_entry_id,omitempty" db:"journal_entry_id"`
}

// ─── Workflow ──────────────────────────────────────────────────────────────────

type WorkflowRule struct {
	Base
	CompanyID    string         `json:"company_id" db:"company_id"`
	Name         string         `json:"name" db:"name"`
	Module       string         `json:"module" db:"module"`
	DocumentType string         `json:"document_type" db:"document_type"`
	Condition    string         `json:"condition" db:"condition"`
	Steps        []WorkflowStep `json:"steps" db:"steps"`
	IsActive     bool           `json:"is_active" db:"is_active"`
}

type WorkflowStep struct {
	StepNumber   int    `json:"step_number"`
	ApproverID   string `json:"approver_id"`
	ApproverName string `json:"approver_name"`
	Role         string `json:"role"`
}

type ApprovalRequest struct {
	Base
	CompanyID    string     `json:"company_id" db:"company_id"`
	RuleID       string     `json:"rule_id" db:"rule_id"`
	Module       string     `json:"module" db:"module"`
	DocumentType string     `json:"document_type" db:"document_type"`
	DocumentID   string     `json:"document_id" db:"document_id"`
	DocumentRef  string     `json:"document_ref" db:"document_ref"`
	RequestedBy  string     `json:"requested_by" db:"requested_by"`
	CurrentStep  int        `json:"current_step" db:"current_step"`
	Status       string     `json:"status" db:"status"` // pending, approved, rejected
	ApproverID   string     `json:"approver_id" db:"approver_id"`
	ApproverName string     `json:"approver_name" db:"approver_name"`
	Notes        string     `json:"notes" db:"notes"`
	ApprovedAt   *time.Time `json:"approved_at,omitempty" db:"approved_at"`
}
