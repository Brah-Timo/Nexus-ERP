import axios, { type AxiosInstance, type AxiosResponse, type AxiosError } from 'axios'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import router from '@/router'

// ─── Axios Instance ────────────────────────────────────────────────────────────

const api: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// ─── Request Interceptor ───────────────────────────────────────────────────────

api.interceptors.request.use(
  (config) => {
    const auth = useAuthStore()
    if (auth.token) {
      config.headers.Authorization = `Bearer ${auth.token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// ─── Response Interceptor ──────────────────────────────────────────────────────

api.interceptors.response.use(
  (response: AxiosResponse) => response,
  async (error: AxiosError) => {
    const appStore = useAppStore()

    if (error.response?.status === 401) {
      const auth = useAuthStore()
      // Try to refresh token
      if (auth.refreshToken) {
        try {
          await auth.refresh()
          // Retry original request
          return api(error.config!)
        } catch {
          auth.logout()
          router.push({ name: 'Login' })
        }
      } else {
        auth.logout()
        router.push({ name: 'Login' })
      }
    }

    if (error.response?.status === 403) {
      appStore.addToast('You do not have permission to perform this action', 'error')
    }

    if (error.response?.status && error.response.status >= 500) {
      appStore.addToast('A server error occurred. Please try again later.', 'error')
    }

    return Promise.reject(error)
  }
)

// ─── Typed API Methods ─────────────────────────────────────────────────────────

export const authAPI = {
  login: (username: string, password: string, companyId?: string) =>
    api.post('/auth/login', { username, password, company_id: companyId }),
  logout: () => api.post('/auth/logout'),
  refresh: (refreshToken: string) => api.post('/auth/refresh', { refresh_token: refreshToken }),
  forgotPassword: (email: string) => api.post('/auth/forgot-password', { email }),
}

export const dashboardAPI = {
  getSummary: () => api.get('/dashboard/summary'),
  getCashFlow: () => api.get('/dashboard/cashflow'),
  getActivity: () => api.get('/dashboard/activity'),
  getApprovals: () => api.get('/dashboard/approvals'),
}

export const accountingAPI = {
  getChartOfAccounts: () => api.get('/accounting/chart-of-accounts'),
  createAccount: (data: unknown) => api.post('/accounting/chart-of-accounts', data),

  getJournalEntries: (params?: Record<string, string>) =>
    api.get('/accounting/journal-entries', { params }),
  getJournalEntry: (id: string) => api.get(`/accounting/journal-entries/${id}`),
  createJournalEntry: (data: unknown) => api.post('/accounting/journal-entries', data),
  postJournalEntry: (id: string) => api.put(`/accounting/journal-entries/${id}/post`),
  cancelJournalEntry: (id: string) => api.put(`/accounting/journal-entries/${id}/cancel`),

  getCostCenters: () => api.get('/accounting/cost-centers'),
  createCostCenter: (data: unknown) => api.post('/accounting/cost-centers', data),

  getFixedAssets: () => api.get('/accounting/fixed-assets'),
  createFixedAsset: (data: unknown) => api.post('/accounting/fixed-assets', data),
  runDepreciation: () => api.post('/accounting/fixed-assets/depreciate'),

  getTrialBalance: () => api.get('/accounting/reports/trial-balance'),
  getBalanceSheet: () => api.get('/accounting/reports/balance-sheet'),
  getIncomeStatement: () => api.get('/accounting/reports/income-statement'),

  getBudgets: () => api.get('/accounting/budgets'),
  createBudget: (data: unknown) => api.post('/accounting/budgets', data),

  getBankReconciliations: () => api.get('/accounting/bank-reconciliations'),
  createBankReconciliation: (data: unknown) => api.post('/accounting/bank-reconciliations', data),
}

export const hrAPI = {
  // Dashboard
  getDashboard: () => api.get('/hr/dashboard'),

  // Employees
  getEmployees: (params?: Record<string, string>) => api.get('/hr/employees', { params }),
  getEmployee: (id: string) => api.get(`/hr/employees/${id}`),
  createEmployee: (data: unknown) => api.post('/hr/employees', data),
  updateEmployee: (id: string, data: unknown) => api.put(`/hr/employees/${id}`, data),
  deleteEmployee: (id: string) => api.delete(`/hr/employees/${id}`),

  // Departments
  getDepartments: () => api.get('/hr/departments'),
  createDepartment: (data: unknown) => api.post('/hr/departments', data),
  updateDepartment: (id: string, data: unknown) => api.put(`/hr/departments/${id}`, data),
  deleteDepartment: (id: string) => api.delete(`/hr/departments/${id}`),

  // Positions
  getPositions: () => api.get('/hr/positions'),
  createPosition: (data: unknown) => api.post('/hr/positions', data),

  // Attendance
  getAttendance: (params?: Record<string, string>) => api.get('/hr/attendance', { params }),
  recordAttendance: (data: unknown) => api.post('/hr/attendance', data),
  updateAttendance: (id: string, data: unknown) => api.put(`/hr/attendance/${id}`, data),
  getAttendanceSummary: (params?: Record<string, string>) => api.get('/hr/attendance/summary', { params }),

  // Leave Types
  getLeaveTypes: () => api.get('/hr/leave-types'),
  createLeaveType: (data: unknown) => api.post('/hr/leave-types', data),

  // Leave Requests
  getLeaveRequests: (params?: Record<string, string>) => api.get('/hr/leave-requests', { params }),
  createLeaveRequest: (data: unknown) => api.post('/hr/leave-requests', data),
  approveLeave: (id: string) => api.put(`/hr/leave-requests/${id}/approve`),
  rejectLeave: (id: string, data?: unknown) => api.put(`/hr/leave-requests/${id}/reject`, data),
  cancelLeave: (id: string) => api.put(`/hr/leave-requests/${id}/cancel`),

  // Payroll
  getPayrollRuns: () => api.get('/hr/payroll/runs'),
  runPayroll: (month: number, year: number) => api.post('/hr/payroll/runs', { month, year }),
  approvePayrollRun: (id: string) => api.put(`/hr/payroll/runs/${id}/approve`),
  payPayrollRun: (id: string) => api.put(`/hr/payroll/runs/${id}/pay`),
  getPayslips: (runId: string) => api.get(`/hr/payroll/runs/${runId}/payslips`),
  exportG29: (runId: string) => api.get(`/hr/payroll/runs/${runId}/g29`),

  // Recruitment — Job Postings
  getJobPostings: (params?: Record<string, string>) => api.get('/hr/recruitment/jobs', { params }),
  getJobPosting: (id: string) => api.get(`/hr/recruitment/jobs/${id}`),
  createJobPosting: (data: unknown) => api.post('/hr/recruitment/jobs', data),
  updateJobPosting: (id: string, data: unknown) => api.put(`/hr/recruitment/jobs/${id}`, data),
  deleteJobPosting: (id: string) => api.delete(`/hr/recruitment/jobs/${id}`),

  // Recruitment — Applications
  getApplications: (params?: Record<string, string>) => api.get('/hr/recruitment/applications', { params }),
  createApplication: (data: unknown) => api.post('/hr/recruitment/applications', data),
  updateApplicationStatus: (id: string, data: unknown) => api.put(`/hr/recruitment/applications/${id}/status`, data),
}

export const salesAPI = {
  // ── Leads ──────────────────────────────────────────────────────────────────
  getLeads: (params?: Record<string, string>) => api.get('/sales/leads', { params }),
  createLead: (data: unknown) => api.post('/sales/leads', data),
  updateLead: (id: string, data: unknown) => api.put(`/sales/leads/${id}`, data),
  deleteLead: (id: string) => api.delete(`/sales/leads/${id}`),

  // ── Opportunities / Pipeline ───────────────────────────────────────────────
  getOpportunities: (params?: Record<string, string>) => api.get('/sales/opportunities', { params }),
  createOpportunity: (data: unknown) => api.post('/sales/opportunities', data),
  updateOpportunity: (id: string, data: unknown) => api.put(`/sales/opportunities/${id}`, data),
  deleteOpportunity: (id: string) => api.delete(`/sales/opportunities/${id}`),
  getPipelineSummary: () => api.get('/sales/pipeline/summary'),

  // ── Customers ──────────────────────────────────────────────────────────────
  getCustomers: (params?: Record<string, string>) => api.get('/sales/customers', { params }),
  getCustomer: (id: string) => api.get(`/sales/customers/${id}`),
  createCustomer: (data: unknown) => api.post('/sales/customers', data),
  updateCustomer: (id: string, data: unknown) => api.put(`/sales/customers/${id}`, data),
  deleteCustomer: (id: string) => api.delete(`/sales/customers/${id}`),

  // ── Quotations ─────────────────────────────────────────────────────────────
  getQuotations: (params?: Record<string, string>) => api.get('/sales/quotations', { params }),
  getQuotation: (id: string) => api.get(`/sales/quotations/${id}`),
  createQuotation: (data: unknown) => api.post('/sales/quotations', data),
  updateQuotation: (id: string, data: unknown) => api.put(`/sales/quotations/${id}`, data),
  confirmQuotation: (id: string) => api.put(`/sales/quotations/${id}/confirm`),
  convertToOrder: (id: string) => api.post(`/sales/quotations/${id}/convert`),
  cancelQuotation: (id: string) => api.put(`/sales/quotations/${id}/cancel`),

  // ── Sales Orders ───────────────────────────────────────────────────────────
  getOrders: (params?: Record<string, string>) => api.get('/sales/orders', { params }),
  getOrder: (id: string) => api.get(`/sales/orders/${id}`),
  createOrder: (data: unknown) => api.post('/sales/orders', data),
  updateOrder: (id: string, data: unknown) => api.put(`/sales/orders/${id}`, data),
  confirmOrder: (id: string) => api.put(`/sales/orders/${id}/confirm`),
  fulfillOrder: (id: string) => api.put(`/sales/orders/${id}/deliver`),
  cancelOrder: (id: string) => api.put(`/sales/orders/${id}/cancel`),

  // ── Sales Invoices ─────────────────────────────────────────────────────────
  getInvoices: (params?: Record<string, string>) => api.get('/sales/invoices', { params }),
  getInvoice: (id: string) => api.get(`/sales/invoices/${id}`),
  createInvoice: (data: unknown) => api.post('/sales/invoices', data),
  confirmInvoice: (id: string) => api.put(`/sales/invoices/${id}/confirm`),
  cancelInvoice: (id: string) => api.put(`/sales/invoices/${id}/cancel`),
  recordPayment: (id: string, data: unknown) => api.post(`/sales/invoices/${id}/payment`, data),

  // ── Reports ────────────────────────────────────────────────────────────────
  getAgingReport: () => api.get('/sales/reports/aging'),
}

export const purchaseAPI = {
  // Suppliers
  getSuppliers: () => api.get('/purchase/suppliers'),
  getSupplier: (id: string) => api.get(`/purchase/suppliers/${id}`),
  createSupplier: (data: unknown) => api.post('/purchase/suppliers', data),
  updateSupplier: (id: string, data: unknown) => api.put(`/purchase/suppliers/${id}`, data),
  deleteSupplier: (id: string) => api.delete(`/purchase/suppliers/${id}`),

  // RFQs
  getRFQs: () => api.get('/purchase/rfqs'),
  getRFQ: (id: string) => api.get(`/purchase/rfqs/${id}`),
  createRFQ: (data: unknown) => api.post('/purchase/rfqs', data),
  updateRFQ: (id: string, data: unknown) => api.put(`/purchase/rfqs/${id}`, data),
  sendRFQ: (id: string) => api.put(`/purchase/rfqs/${id}/send`),
  cancelRFQ: (id: string) => api.put(`/purchase/rfqs/${id}/cancel`),
  convertRFQToOrder: (id: string) => api.post(`/purchase/rfqs/${id}/convert`),

  // Purchase Orders
  getOrders: () => api.get('/purchase/orders'),
  getOrder: (id: string) => api.get(`/purchase/orders/${id}`),
  createOrder: (data: unknown) => api.post('/purchase/orders', data),
  updateOrder: (id: string, data: unknown) => api.put(`/purchase/orders/${id}`, data),
  approveOrder: (id: string) => api.put(`/purchase/orders/${id}/approve`),
  confirmOrder: (id: string) => api.put(`/purchase/orders/${id}/confirm`),
  cancelOrder: (id: string) => api.put(`/purchase/orders/${id}/cancel`),

  // Goods Receipts
  getReceipts: () => api.get('/purchase/receipts'),
  getReceipt: (id: string) => api.get(`/purchase/receipts/${id}`),
  createReceipt: (data: unknown) => api.post('/purchase/receipts', data),
  validateReceipt: (id: string) => api.put(`/purchase/receipts/${id}/validate`),

  // Invoices
  getInvoices: () => api.get('/purchase/invoices'),
  getInvoice: (id: string) => api.get(`/purchase/invoices/${id}`),
  createInvoice: (data: unknown) => api.post('/purchase/invoices', data),
  confirmInvoice: (id: string) => api.put(`/purchase/invoices/${id}/confirm`),
  cancelInvoice: (id: string) => api.put(`/purchase/invoices/${id}/cancel`),
  matchInvoice: (id: string) => api.put(`/purchase/invoices/${id}/match`),
  recordPayment: (id: string, data: unknown) => api.post(`/purchase/invoices/${id}/payment`, data),

  // Evaluations & Reports
  getEvaluations: () => api.get('/purchase/supplier-evaluations'),
  createEvaluation: (data: unknown) => api.post('/purchase/supplier-evaluations', data),
  getAgingReport: () => api.get('/purchase/reports/aging'),
  getDashboard: () => api.get('/purchase/dashboard'),
}

export const inventoryAPI = {
  // Items
  getItems: () => api.get('/inventory/items'),
  getItem: (id: string) => api.get(`/inventory/items/${id}`),
  createItem: (data: unknown) => api.post('/inventory/items', data),
  updateItem: (id: string, data: unknown) => api.put(`/inventory/items/${id}`, data),
  deleteItem: (id: string) => api.delete(`/inventory/items/${id}`),

  // Categories & Units
  getCategories: () => api.get('/inventory/categories'),
  createCategory: (data: unknown) => api.post('/inventory/categories', data),
  getUnits: () => api.get('/inventory/units'),
  createUnit: (data: unknown) => api.post('/inventory/units', data),

  // Warehouses
  getWarehouses: () => api.get('/inventory/warehouses'),
  getWarehouse: (id: string) => api.get(`/inventory/warehouses/${id}`),
  createWarehouse: (data: unknown) => api.post('/inventory/warehouses', data),
  updateWarehouse: (id: string, data: unknown) => api.put(`/inventory/warehouses/${id}`, data),

  // Locations
  getLocations: (warehouseId?: string) =>
    api.get('/inventory/locations', { params: warehouseId ? { warehouse_id: warehouseId } : {} }),
  createLocation: (data: unknown) => api.post('/inventory/locations', data),

  // Stock Levels
  getStockLevels: (params?: { warehouse_id?: string; item_id?: string }) =>
    api.get('/inventory/stock-levels', { params }),

  // Movements
  getMovements: (params?: { type?: string; warehouse_id?: string; item_id?: string }) =>
    api.get('/inventory/movements', { params }),
  createMovement: (data: unknown) => api.post('/inventory/movements/adjustment', data),
  transferStock: (data: unknown) => api.post('/inventory/movements/transfer', data),

  // Inventory Counts
  getInventoryCounts: () => api.get('/inventory/inventory-counts'),
  getInventoryCount: (id: string) => api.get(`/inventory/inventory-counts/${id}`),
  createInventoryCount: (data: unknown) => api.post('/inventory/inventory-counts', data),
  validateInventoryCount: (id: string) => api.put(`/inventory/inventory-counts/${id}/validate`),

  // Reports & Dashboard
  getValuationReport: () => api.get('/inventory/reports/valuation'),
  getDashboard: () => api.get('/inventory/dashboard'),
}

export const manufacturingAPI = {
  // Bill of Materials
  getBOMs: (params?: Record<string, string>) =>
    api.get('/manufacturing/bom', { params }),
  getBOM: (id: string) => api.get(`/manufacturing/bom/${id}`),
  createBOM: (data: unknown) => api.post('/manufacturing/bom', data),
  updateBOM: (id: string, data: unknown) => api.put(`/manufacturing/bom/${id}`, data),
  deleteBOM: (id: string) => api.delete(`/manufacturing/bom/${id}`),

  // Work Centers
  getWorkCenters: (params?: Record<string, string>) =>
    api.get('/manufacturing/work-centers', { params }),
  createWorkCenter: (data: unknown) => api.post('/manufacturing/work-centers', data),
  updateWorkCenter: (id: string, data: unknown) => api.put(`/manufacturing/work-centers/${id}`, data),
  deleteWorkCenter: (id: string) => api.delete(`/manufacturing/work-centers/${id}`),

  // Manufacturing Orders
  getOrders: (params?: Record<string, string>) =>
    api.get('/manufacturing/orders', { params }),
  getOrder: (id: string) => api.get(`/manufacturing/orders/${id}`),
  createOrder: (data: unknown) => api.post('/manufacturing/orders', data),
  updateOrder: (id: string, data: unknown) => api.put(`/manufacturing/orders/${id}`, data),
  startOrder: (id: string) => api.put(`/manufacturing/orders/${id}/start`),
  completeOrder: (id: string, data: unknown) => api.put(`/manufacturing/orders/${id}/complete`, data),
  cancelOrder: (id: string) => api.put(`/manufacturing/orders/${id}/cancel`),

  // MRP & Dashboard
  runMRP: (data?: unknown) => api.post('/manufacturing/mrp/suggest', data ?? {}),
  getDashboard: () => api.get('/manufacturing/dashboard'),
}

export const projectsAPI = {
  // Dashboard & Reports
  getDashboard: () => api.get('/projects/dashboard'),
  getProjectsReport: (params?: unknown) => api.get('/projects/report', { params }),

  // Projects CRUD
  getProjects: (params?: unknown) => api.get('/projects', { params }),
  getProject: (id: string) => api.get(`/projects/${id}`),
  createProject: (data: unknown) => api.post('/projects', data),
  updateProject: (id: string, data: unknown) => api.put(`/projects/${id}`, data),
  deleteProject: (id: string) => api.delete(`/projects/${id}`),
  getProjectCosts: (id: string) => api.get(`/projects/${id}/costs`),

  // Tasks
  getAllTasks: (params?: unknown) => api.get('/projects/tasks/all', { params }),
  getTasks: (projectId: string, params?: unknown) => api.get(`/projects/${projectId}/tasks`, { params }),
  getTask: (taskId: string) => api.get(`/projects/tasks/${taskId}`),
  createTask: (projectId: string, data: unknown) => api.post(`/projects/${projectId}/tasks`, data),
  updateTask: (taskId: string, data: unknown) => api.put(`/projects/tasks/${taskId}`, data),
  deleteTask: (taskId: string) => api.delete(`/projects/tasks/${taskId}`),

  // Milestones
  getMilestones: (projectId: string) => api.get(`/projects/${projectId}/milestones`),
  createMilestone: (projectId: string, data: unknown) => api.post(`/projects/${projectId}/milestones`, data),
  updateMilestone: (milestoneId: string, data: unknown) => api.put(`/projects/milestones/${milestoneId}`, data),
  deleteMilestone: (milestoneId: string) => api.delete(`/projects/milestones/${milestoneId}`),

  // Timesheets
  getTimesheets: (params?: unknown) => api.get('/projects/timesheets', { params }),
  createTimesheet: (data: unknown) => api.post('/projects/timesheets', data),
  updateTimesheet: (timesheetId: string, data: unknown) => api.put(`/projects/timesheets/${timesheetId}`, data),
  deleteTimesheet: (timesheetId: string) => api.delete(`/projects/timesheets/${timesheetId}`),

  // Expenses
  getExpenses: (params?: unknown) => api.get('/projects/expenses', { params }),
  createExpense: (data: unknown) => api.post('/projects/expenses', data),
  updateExpense: (expenseId: string, data: unknown) => api.put(`/projects/expenses/${expenseId}`, data),
  deleteExpense: (expenseId: string) => api.delete(`/projects/expenses/${expenseId}`),

  // Planning
  getPlanningSlots: (params?: unknown) => api.get('/projects/planning', { params }),
  upsertPlanningSlot: (data: unknown) => api.post('/projects/planning', data),
  deletePlanningSlot: (slotId: string) => api.delete(`/projects/planning/${slotId}`),
}

export const treasuryAPI = {
  // Cash Accounts
  getCashAccounts: () => api.get('/treasury/cash-accounts'),
  getCashAccount: (id: string) => api.get(`/treasury/cash-accounts/${id}`),
  createCashAccount: (data: unknown) => api.post('/treasury/cash-accounts', data),
  updateCashAccount: (id: string, data: unknown) => api.put(`/treasury/cash-accounts/${id}`, data),

  // Bank Accounts
  getBankAccounts: () => api.get('/treasury/bank-accounts'),
  createBankAccount: (data: unknown) => api.post('/treasury/bank-accounts', data),
  updateBankAccount: (id: string, data: unknown) => api.put(`/treasury/bank-accounts/${id}`, data),

  // Cheques
  getCheques: (params?: unknown) => api.get('/treasury/cheques', { params }),
  createCheque: (data: unknown) => api.post('/treasury/cheques', data),
  updateCheque: (id: string, data: unknown) => api.put(`/treasury/cheques/${id}`, data),
  depositCheque: (id: string) => api.put(`/treasury/cheques/${id}/deposit`),
  bounceCheque: (id: string) => api.put(`/treasury/cheques/${id}/bounce`),
  cancelCheque: (id: string) => api.put(`/treasury/cheques/${id}/cancel`),

  // Movements
  getMovements: (params?: unknown) => api.get('/treasury/movements', { params }),
  createMovement: (data: unknown) => api.post('/treasury/movements', data),

  // Payments
  getPayments: (params?: unknown) => api.get('/treasury/payments', { params }),
  createPayment: (data: unknown) => api.post('/treasury/payments', data),
  updatePayment: (id: string, data: unknown) => api.put(`/treasury/payments/${id}`, data),
  confirmPayment: (id: string) => api.put(`/treasury/payments/${id}/confirm`),
  allocatePayment: (id: string, data: unknown) => api.put(`/treasury/payments/${id}/allocate`, data),

  // Receipts
  getReceipts: (params?: unknown) => api.get('/treasury/receipts', { params }),
  createReceipt: (data: unknown) => api.post('/treasury/receipts', data),
  updateReceipt: (id: string, data: unknown) => api.put(`/treasury/receipts/${id}`, data),
  confirmReceipt: (id: string) => api.put(`/treasury/receipts/${id}/confirm`),
  cancelReceipt: (id: string) => api.put(`/treasury/receipts/${id}/cancel`),
  deleteReceipt: (id: string) => api.delete(`/treasury/receipts/${id}`),

  // Bank Reconciliation
  getReconciliations: () => api.get('/treasury/reconciliations'),
  getReconciliation: (id: string) => api.get(`/treasury/reconciliations/${id}`),
  createReconciliation: (data: unknown) => api.post('/treasury/reconciliations', data),
  updateReconciliation: (id: string, data: unknown) => api.put(`/treasury/reconciliations/${id}`, data),
  completeReconciliation: (id: string) => api.put(`/treasury/reconciliations/${id}/complete`),
  addReconciliationLine: (id: string, data: unknown) => api.post(`/treasury/reconciliations/${id}/lines`, data),
  matchReconciliationLines: (data: unknown) => api.post('/treasury/reconciliations/match-lines', data),

  // Reports
  getCashPosition: () => api.get('/treasury/reports/cash-position'),
  getAgingReport: () => api.get('/treasury/reports/aging'),
  getTreasuryReport: (params?: unknown) => api.get('/treasury/reports/treasury', { params }),
}

export const settingsAPI = {
  // Companies
  getCompanies: () => api.get('/settings/companies'),
  createCompany: (data: unknown) => api.post('/settings/companies', data),
  updateCompany: (id: string, data: unknown) => api.put(`/settings/companies/${id}`, data),

  // Users
  getUsers: () => api.get('/settings/users'),
  createUser: (data: unknown) => api.post('/settings/users', data),
  updateUser: (id: string, data: unknown) => api.put(`/settings/users/${id}`, data),
  deleteUser: (id: string) => api.delete(`/settings/users/${id}`),

  // Roles
  getRoles: () => api.get('/settings/roles'),
  createRole: (data: unknown) => api.post('/settings/roles', data),
  updateRole: (id: string, data: unknown) => api.put(`/settings/roles/${id}`, data),
  deleteRole: (id: string) => api.delete(`/settings/roles/${id}`),

  // Fiscal Years
  getFiscalYears: () => api.get('/settings/fiscal-years'),
  createFiscalYear: (data: unknown) => api.post('/settings/fiscal-years', data),
  closeFiscalYear: (id: string) => api.put(`/settings/fiscal-years/${id}/close`),

  // Currencies
  getCurrencies: () => api.get('/settings/currencies'),
  createCurrency: (data: unknown) => api.post('/settings/currencies', data),
  updateCurrency: (id: string, data: unknown) => api.put(`/settings/currencies/${id}`, data),
  deleteCurrency: (id: string) => api.delete(`/settings/currencies/${id}`),

  // Numbering
  getNumbering: () => api.get('/settings/numbering'),
  updateNumbering: (data: unknown) => api.put('/settings/numbering', data),

  // Taxes
  getTaxes: () => api.get('/settings/taxes'),
  createTax: (data: unknown) => api.post('/settings/taxes', data),
  updateTax: (id: string, data: unknown) => api.put(`/settings/taxes/${id}`, data),
  deleteTax: (id: string) => api.delete(`/settings/taxes/${id}`),

  // Workflow Rules
  getWorkflowRules: () => api.get('/settings/workflow-rules'),
  createWorkflowRule: (data: unknown) => api.post('/settings/workflow-rules', data),
  updateWorkflowRule: (id: string, data: unknown) => api.put(`/settings/workflow-rules/${id}`, data),
  deleteWorkflowRule: (id: string) => api.delete(`/settings/workflow-rules/${id}`),

  // Audit Log
  getAuditLog: (params?: Record<string, string | number>) => api.get('/settings/audit-log', { params }),
}

export const workflowAPI = {
  getApprovalInbox: () => api.get('/workflow/approvals/inbox'),
  approve: (id: string, notes: string) => api.put(`/workflow/approvals/${id}/approve`, { notes }),
  reject: (id: string, notes: string) => api.put(`/workflow/approvals/${id}/reject`, { notes }),
  getRules: () => api.get('/workflow/rules'),
}

export const taxAPI = {
  // ── Declarations ──────────────────────────────────────────────────────────
  listDeclarations: (params?: Record<string, string>) => api.get('/tax/declarations', { params }),
  getDeclaration:   (id: string) => api.get(`/tax/declarations/${id}`),
  createDeclaration: (data: unknown) => api.post('/tax/declarations', data),
  updateDeclaration: (id: string, data: unknown) => api.put(`/tax/declarations/${id}`, data),
  deleteDeclaration: (id: string) => api.delete(`/tax/declarations/${id}`),
  submitDeclaration: (id: string, data?: unknown) => api.post(`/tax/declarations/${id}/submit`, data ?? {}),
  amendDeclaration:  (id: string) => api.post(`/tax/declarations/${id}/amend`, {}),

  // ── G50 auto-compute & submit ─────────────────────────────────────────────
  getG50:    (year: string, month: string) => api.get('/tax/declarations/g50', { params: { year, month } }),
  submitG50: (data: unknown) => api.post('/tax/declarations/g50', data),

  // ── IBS ───────────────────────────────────────────────────────────────────
  getIBS: (year: string) => api.get('/tax/declarations/ibs', { params: { year } }),

  // ── VAT Register ──────────────────────────────────────────────────────────
  getVATRegister:  (year: string, month: string, type?: string) =>
    api.get('/tax/vat-register', { params: { year, month, type } }),
  createVATEntry:  (data: unknown) => api.post('/tax/vat-register', data),

  // ── VAT Returns ───────────────────────────────────────────────────────────
  listVATReturns:   (params?: Record<string, string>) => api.get('/tax/vat-returns', { params }),
  createVATReturn:  (data: unknown) => api.post('/tax/vat-returns', data),
  updateVATReturn:  (id: string, data: unknown) => api.put(`/tax/vat-returns/${id}`, data),
  submitVATReturn:  (id: string) => api.post(`/tax/vat-returns/${id}/submit`, {}),
  computeVATReturn: (year: string, month: string) =>
    api.get('/tax/vat-returns/compute', { params: { year, month } }),

  // ── Tax Payments ──────────────────────────────────────────────────────────
  listTaxPayments:  (params?: Record<string, string>) => api.get('/tax/payments', { params }),
  createTaxPayment: (data: unknown) => api.post('/tax/payments', data),
  updateTaxPayment: (id: string, data: unknown) => api.put(`/tax/payments/${id}`, data),
  deleteTaxPayment: (id: string) => api.delete(`/tax/payments/${id}`),

  // ── Reports ───────────────────────────────────────────────────────────────
  getTaxReport:     (year: string) => api.get('/tax/reports', { params: { year } }),
  getTaxRates:      () => api.get('/tax/rates'),
}

export const reportsAPI = {
  getFinancialRatios: () => api.get('/reports/financial-ratios'),
  getKPISummary: () => api.get('/reports/kpi-summary'),

  // ── Reports & BI extended ────────────────────────────────────────────────
  getBIDashboard:       (year: string) => api.get('/reports/bi-dashboard', { params: { year } }),
  getFinancialReports:  (year: string) => api.get('/reports/financial', { params: { year } }),
  getSalesReports:      (year: string) => api.get('/reports/sales', { params: { year } }),
  getPurchaseReports:   (year: string) => api.get('/reports/purchase', { params: { year } }),
  getInventoryReports:  () => api.get('/reports/inventory'),
  getProjectReports:    (year: string) => api.get('/reports/projects', { params: { year } }),
  getManagementReports: (year: string) => api.get('/reports/management', { params: { year } }),
  getAnalytics:         (year: string) => api.get('/reports/analytics', { params: { year } }),
  listReportDefinitions: () => api.get('/reports/definitions'),
}

// ── Diagnostics ─────────────────────────────────────────────────────────────
export const diagnosticsAPI = {
  listLogs:       (params?: Record<string, string | number>) => api.get('/diagnostics/logs', { params }),
  createLog:      (data: unknown) => api.post('/diagnostics/logs', data),
  getLog:         (id: string) => api.get(`/diagnostics/logs/${id}`),
  resolveLog:     (id: string, note?: string) => api.post(`/diagnostics/logs/${id}/resolve`, { resolution_note: note }),
  bulkResolve:    (ids: string[], note?: string) => api.post('/diagnostics/logs/bulk-resolve', { ids, resolution_note: note }),
  deleteLog:      (id: string) => api.delete(`/diagnostics/logs/${id}`),
  getStats:       () => api.get('/diagnostics/stats'),
  purgeLogs:      (params: { severity?: string; before_date?: string; resolved_only?: boolean }) =>
    api.delete('/diagnostics/logs/purge', { params }),
}

// ── Maintenance ──────────────────────────────────────────────────────────────
export const maintenanceAPI = {
  // Equipment
  listEquipment:            (params?: Record<string, string>) => api.get('/maintenance/equipment', { params }),
  getEquipment:             (id: string) => api.get(`/maintenance/equipment/${id}`),
  getEquipmentCategories:   () => api.get('/maintenance/equipment/categories'),
  createEquipment:          (data: unknown) => api.post('/maintenance/equipment', data),
  updateEquipment:          (id: string, data: unknown) => api.put(`/maintenance/equipment/${id}`, data),
  deleteEquipment:          (id: string) => api.delete(`/maintenance/equipment/${id}`),

  // Maintenance Requests
  listRequests:     (params?: Record<string, string>) => api.get('/maintenance/requests', { params }),
  getRequest:       (id: string) => api.get(`/maintenance/requests/${id}`),
  createRequest:    (data: unknown) => api.post('/maintenance/requests', data),
  updateRequest:    (id: string, data: unknown) => api.put(`/maintenance/requests/${id}`, data),
  deleteRequest:    (id: string) => api.delete(`/maintenance/requests/${id}`),

  // Maintenance Orders
  listOrders:       (params?: Record<string, string>) => api.get('/maintenance/orders', { params }),
  getOrder:         (id: string) => api.get(`/maintenance/orders/${id}`),
  createOrder:      (data: unknown) => api.post('/maintenance/orders', data),
  updateOrder:      (id: string, data: unknown) => api.put(`/maintenance/orders/${id}`, data),
  completeOrder:    (id: string, data: unknown) => api.put(`/maintenance/orders/${id}/complete`, data),
  deleteOrder:      (id: string) => api.delete(`/maintenance/orders/${id}`),

  // Preventive Plans
  listPreventivePlans:    () => api.get('/maintenance/preventive-plans'),
  createPreventivePlan:   (data: unknown) => api.post('/maintenance/preventive-plans', data),
  updatePreventivePlan:   (id: string, data: unknown) => api.put(`/maintenance/preventive-plans/${id}`, data),
  deletePreventivePlan:   (id: string) => api.delete(`/maintenance/preventive-plans/${id}`),

  // Calendar
  getCalendar: (params?: Record<string, string>) => api.get('/maintenance/calendar', { params }),

  // History
  listHistory:    (params?: Record<string, string>) => api.get('/maintenance/history', { params }),
  createHistory:  (data: unknown) => api.post('/maintenance/history', data),

  // Dashboard & Reports
  getDashboard: () => api.get('/maintenance/dashboard'),
  getReports:   (params?: Record<string, string>) => api.get('/maintenance/reports', { params }),
}

// ── Fleet ─────────────────────────────────────────────────────────────────────
export const fleetAPI = {
  // Vehicles
  listVehicles:   (params?: Record<string, string>) => api.get('/fleet/vehicles', { params }),
  getVehicle:     (id: string) => api.get(`/fleet/vehicles/${id}`),
  createVehicle:  (data: unknown) => api.post('/fleet/vehicles', data),
  updateVehicle:  (id: string, data: unknown) => api.put(`/fleet/vehicles/${id}`, data),
  deleteVehicle:  (id: string) => api.delete(`/fleet/vehicles/${id}`),

  // Drivers
  listDrivers:    (params?: Record<string, string>) => api.get('/fleet/drivers', { params }),
  createDriver:   (data: unknown) => api.post('/fleet/drivers', data),
  updateDriver:   (id: string, data: unknown) => api.put(`/fleet/drivers/${id}`, data),
  deleteDriver:   (id: string) => api.delete(`/fleet/drivers/${id}`),

  // Assignments
  listAssignments:    (params?: Record<string, string>) => api.get('/fleet/assignments', { params }),
  createAssignment:   (data: unknown) => api.post('/fleet/assignments', data),
  updateAssignment:   (id: string, data: unknown) => api.put(`/fleet/assignments/${id}`, data),
  deleteAssignment:   (id: string) => api.delete(`/fleet/assignments/${id}`),

  // Fuel Logs
  listFuelLogs:   (params?: Record<string, string>) => api.get('/fleet/fuel', { params }),
  createFuelLog:  (data: unknown) => api.post('/fleet/fuel', data),
  updateFuelLog:  (id: string, data: unknown) => api.put(`/fleet/fuel/${id}`, data),
  deleteFuelLog:  (id: string) => api.delete(`/fleet/fuel/${id}`),

  // Fleet Maintenance
  listFleetMaintenance:   (params?: Record<string, string>) => api.get('/fleet/maintenance', { params }),
  createFleetMaintenance: (data: unknown) => api.post('/fleet/maintenance', data),
  updateFleetMaintenance: (id: string, data: unknown) => api.put(`/fleet/maintenance/${id}`, data),
  deleteFleetMaintenance: (id: string) => api.delete(`/fleet/maintenance/${id}`),

  // Expenses
  listExpenses:   (params?: Record<string, string>) => api.get('/fleet/expenses', { params }),
  createExpense:  (data: unknown) => api.post('/fleet/expenses', data),
  updateExpense:  (id: string, data: unknown) => api.put(`/fleet/expenses/${id}`, data),
  deleteExpense:  (id: string) => api.delete(`/fleet/expenses/${id}`),

  // Dashboard & Reports
  getDashboard:   () => api.get('/fleet/dashboard'),
  getReports:     (params?: Record<string, string>) => api.get('/fleet/reports', { params }),
}

// ── Quality ───────────────────────────────────────────────────────────────────
export const qualityAPI = {
  // Dashboard
  getDashboard: () => api.get('/quality/dashboard'),

  // Control Plans
  listPlans:    () => api.get('/quality/plans'),
  createPlan:   (data: unknown) => api.post('/quality/plans', data),
  updatePlan:   (id: string, data: unknown) => api.put(`/quality/plans/${id}`, data),
  deletePlan:   (id: string) => api.delete(`/quality/plans/${id}`),

  // Inspections
  listInspections:   (params?: Record<string, string>) => api.get('/quality/inspections', { params }),
  getInspection:     (id: string) => api.get(`/quality/inspections/${id}`),
  createInspection:  (data: unknown) => api.post('/quality/inspections', data),
  updateInspection:  (id: string, data: unknown) => api.put(`/quality/inspections/${id}`, data),
  startInspection:   (id: string) => api.post(`/quality/inspections/${id}/start`, {}),
  completeInspection:(id: string, data: unknown) => api.post(`/quality/inspections/${id}/complete`, data),
  deleteInspection:  (id: string) => api.delete(`/quality/inspections/${id}`),

  // Checks
  listChecks:        (params?: Record<string, string>) => api.get('/quality/checks', { params }),
  createCheck:       (data: unknown) => api.post('/quality/checks', data),
  recordCheckResult: (id: string, data: unknown) => api.put(`/quality/checks/${id}/result`, data),
  deleteCheck:       (id: string) => api.delete(`/quality/checks/${id}`),

  // Non-Conformities
  listNonConformities:  (params?: Record<string, string>) => api.get('/quality/non-conformities', { params }),
  getNonConformity:     (id: string) => api.get(`/quality/non-conformities/${id}`),
  createNonConformity:  (data: unknown) => api.post('/quality/non-conformities', data),
  updateNonConformity:  (id: string, data: unknown) => api.put(`/quality/non-conformities/${id}`, data),
  updateNCStatus:       (id: string, data: unknown) => api.put(`/quality/non-conformities/${id}/status`, data),
  deleteNonConformity:  (id: string) => api.delete(`/quality/non-conformities/${id}`),

  // Corrective Actions
  listCorrectiveActions:   (params?: Record<string, string>) => api.get('/quality/corrective-actions', { params }),
  getCorrectiveAction:     (id: string) => api.get(`/quality/corrective-actions/${id}`),
  createCorrectiveAction:  (data: unknown) => api.post('/quality/corrective-actions', data),
  updateCorrectiveAction:  (id: string, data: unknown) => api.put(`/quality/corrective-actions/${id}`, data),
  updateCAStatus:          (id: string, data: unknown) => api.put(`/quality/corrective-actions/${id}/status`, data),
  deleteCorrectiveAction:  (id: string) => api.delete(`/quality/corrective-actions/${id}`),

  // Reports
  getReports: (params?: Record<string, string>) => api.get('/quality/reports', { params }),
}

export default api
