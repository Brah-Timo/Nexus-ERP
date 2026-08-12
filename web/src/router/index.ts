import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  // ── Auth ──────────────────────────────────────────────────────────────────
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/modules/auth/Login.vue'),
    meta: { public: true, title: 'Sign In' }
  },
  {
    path: '/forgot-password',
    name: 'ForgotPassword',
    component: () => import('@/modules/auth/ForgotPassword.vue'),
    meta: { public: true, title: 'Reset Password' }
  },

  // ── Main App ──────────────────────────────────────────────────────────────
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/modules/dashboard/Dashboard.vue'),
    meta: { title: 'Dashboard', icon: 'LayoutDashboard' }
  },

  // ── Accounting ────────────────────────────────────────────────────────────
  {
    path: '/accounting',
    name: 'Accounting',
    redirect: '/accounting/journal-entries',
    meta: { title: 'Accounting', icon: 'Calculator', group: 'Finance' }
  },
  {
    path: '/accounting/chart-of-accounts',
    name: 'ChartOfAccounts',
    component: () => import('@/modules/accounting/ChartOfAccounts.vue'),
    meta: { title: 'Chart of Accounts', icon: 'ListTree', group: 'Finance' }
  },
  {
    path: '/accounting/journal-entries',
    name: 'JournalEntries',
    component: () => import('@/modules/accounting/JournalEntries.vue'),
    meta: { title: 'Journal Entries', icon: 'BookOpen', group: 'Finance' }
  },
  {
    path: '/accounting/journal-entries/:id',
    name: 'JournalEntryDetail',
    component: () => import('@/modules/accounting/JournalEntryDetail.vue'),
    meta: { title: 'Journal Entry', group: 'Finance' }
  },
  {
    path: '/accounting/cost-centers',
    name: 'CostCenters',
    component: () => import('@/modules/accounting/CostCenters.vue'),
    meta: { title: 'Cost Centers', icon: 'Layers', group: 'Finance' }
  },
  {
    path: '/accounting/fixed-assets',
    name: 'FixedAssets',
    component: () => import('@/modules/accounting/FixedAssets.vue'),
    meta: { title: 'Fixed Assets', icon: 'Building', group: 'Finance' }
  },
  {
    path: '/accounting/bank-reconciliation',
    name: 'BankReconciliation',
    component: () => import('@/modules/accounting/BankReconciliation.vue'),
    meta: { title: 'Bank Reconciliation', icon: 'Landmark', group: 'Finance' }
  },
  {
    path: '/accounting/reports',
    name: 'FinancialReports',
    component: () => import('@/modules/accounting/FinancialReports.vue'),
    meta: { title: 'Financial Reports', icon: 'FileBarChart', group: 'Finance' }
  },
  {
    path: '/accounting/budgets',
    name: 'Budgets',
    component: () => import('@/modules/accounting/Budgets.vue'),
    meta: { title: 'Budgets', icon: 'Target', group: 'Finance' }
  },

  // ── HR & Payroll ──────────────────────────────────────────────────────────
  {
    path: '/hr',
    redirect: '/hr/employees',
    meta: { title: 'HR', group: 'Human Resources' }
  },
  {
    path: '/hr/employees',
    name: 'Employees',
    component: () => import('@/modules/hr/Employees.vue'),
    meta: { title: 'Employees', icon: 'IdCard', group: 'Human Resources' }
  },
  {
    path: '/hr/employees/:id',
    name: 'EmployeeDetail',
    component: () => import('@/modules/hr/EmployeeDetail.vue'),
    meta: { title: 'Employee Profile', group: 'Human Resources' }
  },
  {
    path: '/hr/departments',
    name: 'Departments',
    component: () => import('@/modules/hr/Departments.vue'),
    meta: { title: 'Departments', icon: 'Network', group: 'Human Resources' }
  },
  {
    path: '/hr/attendance',
    name: 'Attendance',
    component: () => import('@/modules/hr/Attendance.vue'),
    meta: { title: 'Attendance', icon: 'Clock4', group: 'Human Resources' }
  },
  {
    path: '/hr/leave-requests',
    name: 'LeaveRequests',
    component: () => import('@/modules/hr/LeaveRequests.vue'),
    meta: { title: 'Leave Requests', icon: 'CalendarOff', group: 'Human Resources' }
  },
  {
    path: '/hr/payroll',
    name: 'Payroll',
    component: () => import('@/modules/hr/Payroll.vue'),
    meta: { title: 'Payroll', icon: 'Wallet', group: 'Human Resources' }
  },
  {
    path: '/hr/recruitment',
    name: 'Recruitment',
    component: () => import('@/modules/hr/Recruitment.vue'),
    meta: { title: 'Recruitment', icon: 'UserPlus', group: 'Human Resources' }
  },

  // ── Sales & CRM ───────────────────────────────────────────────────────────
  {
    path: '/sales',
    redirect: '/sales/invoices',
    meta: { title: 'Sales', group: 'Sales' }
  },
  {
    path: '/sales/pipeline',
    name: 'SalesPipeline',
    component: () => import('@/modules/sales/Pipeline.vue'),
    meta: { title: 'Sales Pipeline', icon: 'Filter', group: 'Sales' }
  },
  {
    path: '/sales/customers',
    name: 'Customers',
    component: () => import('@/modules/sales/Customers.vue'),
    meta: { title: 'Customers', icon: 'Users', group: 'Sales' }
  },
  {
    path: '/sales/quotations',
    name: 'Quotations',
    component: () => import('@/modules/sales/Quotations.vue'),
    meta: { title: 'Quotations', icon: 'FileText', group: 'Sales' }
  },
  {
    path: '/sales/orders',
    name: 'SalesOrders',
    component: () => import('@/modules/sales/SalesOrders.vue'),
    meta: { title: 'Sales Orders', icon: 'ShoppingBag', group: 'Sales' }
  },
  {
    path: '/sales/invoices',
    name: 'SalesInvoices',
    component: () => import('@/modules/sales/SalesInvoices.vue'),
    meta: { title: 'Invoices', icon: 'ReceiptText', group: 'Sales' }
  },
  {
    path: '/sales/invoices/:id',
    name: 'SalesInvoiceDetail',
    component: () => import('@/modules/sales/SalesInvoiceDetail.vue'),
    meta: { title: 'Invoice Detail', group: 'Sales' }
  },
  {
    path: '/sales/reports/aging',
    name: 'CustomerAging',
    component: () => import('@/modules/sales/AgingReport.vue'),
    meta: { title: 'Customer Aging', icon: 'AlarmClock', group: 'Sales' }
  },

  // ── Purchase ──────────────────────────────────────────────────────────────
  {
    path: '/purchase',
    redirect: '/purchase/orders',
    meta: { title: 'Purchase', group: 'Purchase' }
  },
  {
    path: '/purchase/suppliers',
    name: 'Suppliers',
    component: () => import('@/modules/purchase/Suppliers.vue'),
    meta: { title: 'Suppliers', icon: 'Truck', group: 'Purchase' }
  },
  {
    path: '/purchase/rfqs',
    name: 'RFQs',
    component: () => import('@/modules/purchase/RFQs.vue'),
    meta: { title: 'Request for Quotes', icon: 'Mail', group: 'Purchase' }
  },
  {
    path: '/purchase/orders',
    name: 'PurchaseOrders',
    component: () => import('@/modules/purchase/PurchaseOrders.vue'),
    meta: { title: 'Purchase Orders', icon: 'ShoppingCart', group: 'Purchase' }
  },
  {
    path: '/purchase/receipts',
    name: 'GoodsReceipts',
    component: () => import('@/modules/purchase/GoodsReceipts.vue'),
    meta: { title: 'Goods Receipts', icon: 'PackageCheck', group: 'Purchase' }
  },
  {
    path: '/purchase/invoices',
    name: 'PurchaseInvoices',
    component: () => import('@/modules/purchase/PurchaseInvoices.vue'),
    meta: { title: 'Supplier Invoices', icon: 'FileInput', group: 'Purchase' }
  },

  // ── Inventory ─────────────────────────────────────────────────────────────
  {
    path: '/inventory',
    redirect: '/inventory/items',
    meta: { title: 'Inventory', group: 'Inventory' }
  },
  {
    path: '/inventory/items',
    name: 'Items',
    component: () => import('@/modules/inventory/Items.vue'),
    meta: { title: 'Products & Items', icon: 'Package', group: 'Inventory' }
  },
  {
    path: '/inventory/stock-levels',
    name: 'StockLevels',
    component: () => import('@/modules/inventory/StockLevels.vue'),
    meta: { title: 'Stock Levels', icon: 'BarChart2', group: 'Inventory' }
  },
  {
    path: '/inventory/warehouses',
    name: 'Warehouses',
    component: () => import('@/modules/inventory/Warehouses.vue'),
    meta: { title: 'Warehouses', icon: 'Warehouse', group: 'Inventory' }
  },
  {
    path: '/inventory/movements',
    name: 'StockMovements',
    component: () => import('@/modules/inventory/Movements.vue'),
    meta: { title: 'Stock Movements', icon: 'ArrowLeftRight', group: 'Inventory' }
  },
  {
    path: '/inventory/counts',
    name: 'InventoryCounts',
    component: () => import('@/modules/inventory/InventoryCounts.vue'),
    meta: { title: 'Inventory Counts', icon: 'ClipboardList', group: 'Inventory' }
  },

  // ── Manufacturing ─────────────────────────────────────────────────────────
  {
    path: '/manufacturing',
    redirect: '/manufacturing/orders',
    meta: { title: 'Manufacturing', group: 'Production' }
  },
  {
    path: '/manufacturing/bom',
    name: 'BOM',
    component: () => import('@/modules/manufacturing/BOM.vue'),
    meta: { title: 'Bill of Materials', icon: 'Layers3', group: 'Production' }
  },
  {
    path: '/manufacturing/orders',
    name: 'ManufacturingOrders',
    component: () => import('@/modules/manufacturing/ManufacturingOrders.vue'),
    meta: { title: 'Manufacturing Orders', icon: 'Cog', group: 'Production' }
  },
  {
    path: '/manufacturing/work-centers',
    name: 'WorkCenters',
    component: () => import('@/modules/manufacturing/WorkCenters.vue'),
    meta: { title: 'Work Centers', icon: 'Factory', group: 'Production' }
  },
  {
    path: '/manufacturing/mrp',
    name: 'MRP',
    component: () => import('@/modules/manufacturing/MRP.vue'),
    meta: { title: 'MRP Planning', icon: 'CalendarClock', group: 'Production' }
  },

  // ── Projects ──────────────────────────────────────────────────────────────
  {
    path: '/projects',
    name: 'Projects',
    component: () => import('@/modules/projects/Projects.vue'),
    meta: { title: 'Projects', icon: 'FolderKanban', group: 'Projects' }
  },
  {
    path: '/projects/tasks',
    name: 'Tasks',
    component: () => import('@/modules/projects/Tasks.vue'),
    meta: { title: 'Tasks', icon: 'CheckSquare', group: 'Projects' }
  },
  {
    path: '/projects/timesheets',
    name: 'Timesheets',
    component: () => import('@/modules/projects/Timesheets.vue'),
    meta: { title: 'Timesheets', icon: 'Clock', group: 'Projects' }
  },
  {
    path: '/projects/planning',
    name: 'Planning',
    component: () => import('@/modules/projects/Planning.vue'),
    meta: { title: 'Planning', icon: 'CalendarDays', group: 'Projects' }
  },
  {
    path: '/projects/milestones',
    name: 'Milestones',
    component: () => import('@/modules/projects/Milestones.vue'),
    meta: { title: 'Milestones', icon: 'Flag', group: 'Projects' }
  },
  {
    path: '/projects/expenses',
    name: 'ProjectExpenses',
    component: () => import('@/modules/projects/Expenses.vue'),
    meta: { title: 'Expenses', icon: 'Receipt', group: 'Projects' }
  },
  {
    path: '/projects/reports',
    name: 'ProjectReports',
    component: () => import('@/modules/projects/ProjectReports.vue'),
    meta: { title: 'Project Reports', icon: 'BarChart3', group: 'Projects' }
  },
  {
    path: '/projects/:id',
    name: 'ProjectDetail',
    component: () => import('@/modules/projects/ProjectDetail.vue'),
    meta: { title: 'Project Detail', group: 'Projects' }
  },

  // ── Treasury ──────────────────────────────────────────────────────────────
  {
    path: '/treasury',
    redirect: '/treasury/cash-position',
    meta: { title: 'Treasury', group: 'Finance' }
  },
  {
    path: '/treasury/cash-position',
    name: 'CashPosition',
    component: () => import('@/modules/treasury/CashPosition.vue'),
    meta: { title: 'Cash Position', icon: 'Landmark', group: 'Finance' }
  },
  {
    path: '/treasury/bank-accounts',
    name: 'BankAccounts',
    component: () => import('@/modules/treasury/BankAccounts.vue'),
    meta: { title: 'Bank Accounts', icon: 'CreditCard', group: 'Finance' }
  },
  {
    path: '/treasury/cheques',
    name: 'Cheques',
    component: () => import('@/modules/treasury/Cheques.vue'),
    meta: { title: 'Cheques', icon: 'Receipt', group: 'Finance' }
  },
  {
    path: '/treasury/payments',
    name: 'Payments',
    component: () => import('@/modules/treasury/Payments.vue'),
    meta: { title: 'Payments', icon: 'CreditCard', group: 'Finance' }
  },
  {
    path: '/treasury/receipts',
    name: 'Receipts',
    component: () => import('@/modules/treasury/Receipts.vue'),
    meta: { title: 'Receipts', icon: 'ArrowDownToLine', group: 'Finance' }
  },
  {
    path: '/treasury/reconciliation',
    name: 'TreasuryBankReconciliation',
    component: () => import('@/modules/treasury/BankReconciliation.vue'),
    meta: { title: 'Bank Reconciliation', icon: 'GitMerge', group: 'Finance' }
  },
  {
    path: '/treasury/reports',
    name: 'TreasuryReports',
    component: () => import('@/modules/treasury/TreasuryReports.vue'),
    meta: { title: 'Treasury Reports', icon: 'BarChart2', group: 'Finance' }
  },

  // ── Tax ───────────────────────────────────────────────────────────────────
  {
    path: '/tax/g50',
    name: 'G50',
    component: () => import('@/modules/tax/G50Declaration.vue'),
    meta: { title: 'G50 Declaration', icon: 'FileSpreadsheet', group: 'Tax' }
  },
  {
    path: '/tax/vat-register',
    name: 'VATRegister',
    component: () => import('@/modules/tax/VATRegister.vue'),
    meta: { title: 'VAT Register', icon: 'Percent', group: 'Tax' }
  },
  {
    path: '/tax/vat-returns',
    name: 'VATReturns',
    component: () => import('@/modules/tax/VATReturns.vue'),
    meta: { title: 'VAT Returns', icon: 'FileCheck', group: 'Tax' }
  },
  {
    path: '/tax/payments',
    name: 'TaxPayments',
    component: () => import('@/modules/tax/TaxPayments.vue'),
    meta: { title: 'Tax Payments', icon: 'Banknote', group: 'Tax' }
  },
  {
    path: '/tax/reports',
    name: 'TaxReports',
    component: () => import('@/modules/tax/TaxReports.vue'),
    meta: { title: 'Tax Reports', icon: 'BarChart2', group: 'Tax' }
  },

  // ── Reports & BI ──────────────────────────────────────────────────────────
  {
    path: '/reports',
    name: 'Reports',
    component: () => import('@/modules/reports/Reports.vue'),
    meta: { title: 'Reports & BI', icon: 'BarChart3', group: 'Reports' }
  },
  {
    path: '/reports/bi-dashboard',
    name: 'BIDashboard',
    component: () => import('@/modules/reports/BIDashboard.vue'),
    meta: { title: 'BI Dashboard', icon: 'BarChart3', group: 'Reports' }
  },
  {
    path: '/reports/financial',
    name: 'ReportsFinancial',
    component: () => import('@/modules/reports/FinancialReports.vue'),
    meta: { title: 'Financial Reports', icon: 'FileBarChart', group: 'Reports' }
  },
  {
    path: '/reports/sales',
    name: 'ReportsSales',
    component: () => import('@/modules/reports/SalesReports.vue'),
    meta: { title: 'Sales Reports', icon: 'TrendingUp', group: 'Reports' }
  },
  {
    path: '/reports/purchase',
    name: 'ReportsPurchase',
    component: () => import('@/modules/reports/PurchaseReports.vue'),
    meta: { title: 'Purchase Reports', icon: 'ShoppingBag', group: 'Reports' }
  },
  {
    path: '/reports/inventory',
    name: 'ReportsInventory',
    component: () => import('@/modules/reports/InventoryReports.vue'),
    meta: { title: 'Inventory Reports', icon: 'Package', group: 'Reports' }
  },
  {
    path: '/reports/projects',
    name: 'ReportsProjects',
    component: () => import('@/modules/reports/ProjectReports.vue'),
    meta: { title: 'Project Reports', icon: 'FolderKanban', group: 'Reports' }
  },
  {
    path: '/reports/management',
    name: 'ReportsManagement',
    component: () => import('@/modules/reports/ManagementReports.vue'),
    meta: { title: 'Management Reports', icon: 'BarChart3', group: 'Reports' }
  },
  {
    path: '/reports/analytics',
    name: 'ReportsAnalytics',
    component: () => import('@/modules/reports/Analytics.vue'),
    meta: { title: 'Analytics', icon: 'LineChart', group: 'Reports' }
  },
  // ── System Diagnostics ────────────────────────────────────────────────────
  {
    path: '/settings/diagnostics',
    name: 'SystemDiagnostics',
    component: () => import('@/modules/settings/SystemDiagnostics.vue'),
    meta: { title: 'System Diagnostics', icon: 'Activity', group: 'System' }
  },

  // ── Settings ──────────────────────────────────────────────────────────────
  {
    path: '/settings',
    redirect: '/settings/companies',
    meta: { title: 'Settings', group: 'System' }
  },
  {
    path: '/settings/companies',
    name: 'Companies',
    component: () => import('@/modules/settings/Companies.vue'),
    meta: { title: 'Companies', icon: 'Building2', group: 'System' }
  },
  {
    path: '/settings/users',
    name: 'Users',
    component: () => import('@/modules/settings/Users.vue'),
    meta: { title: 'Users & Roles', icon: 'Users', group: 'System' }
  },
  {
    path: '/settings/fiscal-years',
    name: 'FiscalYears',
    component: () => import('@/modules/settings/FiscalYears.vue'),
    meta: { title: 'Fiscal Years', icon: 'CalendarDays', group: 'System' }
  },
  {
    path: '/settings/workflow',
    name: 'WorkflowSettings',
    component: () => import('@/modules/settings/Workflow.vue'),
    meta: { title: 'Workflow & Approvals', icon: 'GitBranch', group: 'System' }
  },
  {
    path: '/settings/currencies',
    name: 'Currencies',
    component: () => import('@/modules/settings/Currencies.vue'),
    meta: { title: 'Currencies', icon: 'DollarSign', group: 'System' }
  },
  {
    path: '/settings/numbering',
    name: 'Numbering',
    component: () => import('@/modules/settings/Numbering.vue'),
    meta: { title: 'Document Numbering', icon: 'Hash', group: 'System' }
  },
  {
    path: '/settings/taxes',
    name: 'TaxSettings',
    component: () => import('@/modules/settings/Taxes.vue'),
    meta: { title: 'Taxes', icon: 'Percent', group: 'System' }
  },
  {
    path: '/settings/audit-log',
    name: 'AuditLog',
    component: () => import('@/modules/settings/AuditLog.vue'),
    meta: { title: 'Audit Log', icon: 'ScrollText', group: 'System' }
  },

  // ── Maintenance ───────────────────────────────────────────────────────────
  {
    path: '/maintenance',
    name: 'MaintenanceDashboard',
    component: () => import('@/modules/maintenance/MaintenanceReports.vue'),
    meta: { title: 'Maintenance', icon: 'Wrench', group: 'Maintenance' }
  },
  {
    path: '/maintenance/equipment',
    name: 'Equipment',
    component: () => import('@/modules/maintenance/Equipment.vue'),
    meta: { title: 'Equipment', icon: 'Settings', group: 'Maintenance' }
  },
  {
    path: '/maintenance/requests',
    name: 'MaintenanceRequests',
    component: () => import('@/modules/maintenance/MaintenanceRequests.vue'),
    meta: { title: 'Maintenance Requests', icon: 'ClipboardList', group: 'Maintenance' }
  },
  {
    path: '/maintenance/orders',
    name: 'MaintenanceOrders',
    component: () => import('@/modules/maintenance/MaintenanceOrders.vue'),
    meta: { title: 'Maintenance Orders', icon: 'Wrench', group: 'Maintenance' }
  },
  {
    path: '/maintenance/preventive',
    name: 'PreventiveMaintenance',
    component: () => import('@/modules/maintenance/PreventiveMaintenance.vue'),
    meta: { title: 'Preventive Maintenance', icon: 'Shield', group: 'Maintenance' }
  },
  {
    path: '/maintenance/calendar',
    name: 'MaintenanceCalendar',
    component: () => import('@/modules/maintenance/MaintenanceCalendar.vue'),
    meta: { title: 'Maintenance Calendar', icon: 'Calendar', group: 'Maintenance' }
  },
  {
    path: '/maintenance/history',
    name: 'MaintenanceHistory',
    component: () => import('@/modules/maintenance/MaintenanceHistory.vue'),
    meta: { title: 'Maintenance History', icon: 'History', group: 'Maintenance' }
  },
  {
    path: '/maintenance/reports',
    name: 'MaintenanceReports',
    component: () => import('@/modules/maintenance/MaintenanceReports.vue'),
    meta: { title: 'Maintenance Reports', icon: 'BarChart2', group: 'Maintenance' }
  },

  // ── Fleet ─────────────────────────────────────────────────────────────────
  {
    path: '/fleet',
    name: 'FleetDashboard',
    component: () => import('@/modules/fleet/FleetReports.vue'),
    meta: { title: 'Fleet Dashboard', icon: 'BarChart3', group: 'Fleet' }
  },
  {
    path: '/fleet/vehicles',
    name: 'FleetVehicles',
    component: () => import('@/modules/fleet/Vehicles.vue'),
    meta: { title: 'Fleet Vehicles', icon: 'Truck', group: 'Fleet' }
  },
  {
    path: '/fleet/drivers',
    name: 'FleetDrivers',
    component: () => import('@/modules/fleet/Drivers.vue'),
    meta: { title: 'Drivers', icon: 'UserCheck', group: 'Fleet' }
  },
  {
    path: '/fleet/assignments',
    name: 'VehicleAssignments',
    component: () => import('@/modules/fleet/VehicleAssignments.vue'),
    meta: { title: 'Vehicle Assignments', icon: 'Link', group: 'Fleet' }
  },
  {
    path: '/fleet/fuel',
    name: 'FuelLogs',
    component: () => import('@/modules/fleet/FuelLogs.vue'),
    meta: { title: 'Fuel Logs', icon: 'Fuel', group: 'Fleet' }
  },
  {
    path: '/fleet/maintenance',
    name: 'FleetMaintenance',
    component: () => import('@/modules/fleet/FleetMaintenance.vue'),
    meta: { title: 'Fleet Maintenance', icon: 'Wrench', group: 'Fleet' }
  },
  {
    path: '/fleet/expenses',
    name: 'VehicleExpenses',
    component: () => import('@/modules/fleet/VehicleExpenses.vue'),
    meta: { title: 'Vehicle Expenses', icon: 'Receipt', group: 'Fleet' }
  },
  {
    path: '/fleet/reports',
    name: 'FleetReports',
    component: () => import('@/modules/fleet/FleetReports.vue'),
    meta: { title: 'Fleet Reports', icon: 'BarChart3', group: 'Fleet' }
  },

  // ── Quality ───────────────────────────────────────────────────────────────
  {
    path: '/quality',
    name: 'QualityDashboard',
    component: () => import('@/modules/quality/QualityDashboard.vue'),
    meta: { title: 'Quality Dashboard', icon: 'ShieldCheck', group: 'Quality' }
  },
  {
    path: '/quality/inspections',
    name: 'QualityInspections',
    component: () => import('@/modules/quality/Inspections.vue'),
    meta: { title: 'Quality Inspections', icon: 'ClipboardCheck', group: 'Quality' }
  },
  {
    path: '/quality/checks',
    name: 'QualityChecks',
    component: () => import('@/modules/quality/QualityChecks.vue'),
    meta: { title: 'Quality Checks', icon: 'CheckSquare', group: 'Quality' }
  },
  {
    path: '/quality/non-conformities',
    name: 'NonConformities',
    component: () => import('@/modules/quality/NonConformities.vue'),
    meta: { title: 'Non-Conformities', icon: 'AlertTriangle', group: 'Quality' }
  },
  {
    path: '/quality/corrective-actions',
    name: 'CorrectiveActions',
    component: () => import('@/modules/quality/CorrectiveActions.vue'),
    meta: { title: 'Corrective Actions', icon: 'Wrench', group: 'Quality' }
  },
  {
    path: '/quality/reports',
    name: 'QualityReports',
    component: () => import('@/modules/quality/QualityReports.vue'),
    meta: { title: 'Quality Reports', icon: 'BarChart3', group: 'Quality' }
  },

  // ── Catch-all ─────────────────────────────────────────────────────────────
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/modules/NotFound.vue'),
    meta: { public: true, title: '404 Not Found' }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(_, __, savedPosition) {
    if (savedPosition) return savedPosition
    return { top: 0 }
  }
})

// Navigation guard — redirect unauthenticated users to login
router.beforeEach((to, _, next) => {
  const auth = useAuthStore()

  // Update document title
  document.title = `${to.meta?.title ?? 'Nexus ERP'} — Nexus ERP`

  if (to.meta?.public) {
    next()
    return
  }

  if (!auth.token) {
    next({ name: 'Login', query: { redirect: to.fullPath } })
    return
  }

  next()
})

export default router
