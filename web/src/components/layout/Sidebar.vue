<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app'
import {
  LayoutDashboard,
  Calculator,
  ListChecks,
  Tags,
  Building2,
  Banknote,
  LineChart,
  Target,
  Filter,
  Users,
  FileText,
  ShoppingBag,
  Receipt,
  Clock,
  Truck,
  Mail,
  Package,
  Warehouse,
  ArrowLeftRight,
  ClipboardList,
  Cog,
  Hammer,
  Factory,
  CalendarDays,
  User,
  IdCard,
  Globe,
  Timer,
  Calendar,
  CreditCard,
  UserPlus,
  FolderKanban,
  Landmark,
  ChevronRight,
  ChevronDown,
  Percent,
  FileSearch,
  Settings,
  Building,
  Workflow,
  DollarSign,
  Hash,
  ChevronLeft,
  BookOpen,
  Scale,
  PiggyBank,
  BarChart3,
  TrendingUp,
  CheckSquare,
  Flag,
  Wallet,
  ArrowDownToLine,
  GitMerge,
  BarChart2,
  FileCheck,
  FileBarChart,
  ScrollText,
  Activity,
  Monitor,
  Wrench,
  Shield,
  History,
  Fuel,
  UserCheck,
  Link,
  Receipt as ReceiptIcon,
  BarChart3 as BarChart3Icon,
  ShieldCheck,
  ClipboardCheck,
  AlertOctagon
} from 'lucide-vue-next'

const props = defineProps<{ collapsed: boolean }>()
const route = useRoute()
const app = useAppStore()

const expandedGroups = ref<Record<string, boolean>>({
  Finance: true,
  Sales: false,
  Purchase: false,
  Inventory: false,
  'Human Resources': false,
  Production: false,
  Projects: false,
  Tax: false,
  Reports: false,
  Maintenance: false,
  Fleet: false,
  Quality: false,
  System: false,
})

interface NavItem {
  label: string
  icon: any
  to?: string
  children?: NavItem[]
  group?: string
}

const navItems: NavItem[] = [
  { label: 'Dashboard', icon: LayoutDashboard, to: '/dashboard' },

  { label: 'Accounting', icon: Calculator, group: 'Finance', children: [
    { label: 'Chart of Accounts',    icon: BookOpen,        to: '/accounting/chart-of-accounts' },
    { label: 'Journal Entries',      icon: FileText,        to: '/accounting/journal-entries' },
    { label: 'Cost Centers',         icon: Tags,            to: '/accounting/cost-centers' },
    { label: 'Fixed Assets',         icon: Building2,       to: '/accounting/fixed-assets' },
    { label: 'Bank Reconciliation',  icon: Scale,           to: '/accounting/bank-reconciliation' },
    { label: 'Financial Reports',    icon: BarChart3,       to: '/accounting/reports' },
    { label: 'Budgets',              icon: TrendingUp,      to: '/accounting/budgets' },
  ]},

  { label: 'Sales & CRM', icon: Filter, group: 'Sales', children: [
    { label: 'Pipeline',       icon: Filter,      to: '/sales/pipeline' },
    { label: 'Customers',      icon: Users,       to: '/sales/customers' },
    { label: 'Quotations',     icon: FileText,    to: '/sales/quotations' },
    { label: 'Sales Orders',   icon: ShoppingBag, to: '/sales/orders' },
    { label: 'Invoices',       icon: Receipt,     to: '/sales/invoices' },
    { label: 'Customer Aging', icon: Clock,       to: '/sales/reports/aging' },
  ]},

  { label: 'Purchase', icon: ShoppingBag, group: 'Purchase', children: [
    { label: 'Suppliers',         icon: Truck,      to: '/purchase/suppliers' },
    { label: 'RFQs',              icon: Mail,       to: '/purchase/rfqs' },
    { label: 'Purchase Orders',   icon: ShoppingBag,to: '/purchase/orders' },
    { label: 'Goods Receipts',    icon: Package,    to: '/purchase/receipts' },
    { label: 'Supplier Invoices', icon: Receipt,    to: '/purchase/invoices' },
  ]},

  { label: 'Inventory', icon: Package, group: 'Inventory', children: [
    { label: 'Products & Items',   icon: Tags,          to: '/inventory/items' },
    { label: 'Stock Levels',       icon: LineChart,     to: '/inventory/stock-levels' },
    { label: 'Warehouses',         icon: Warehouse,     to: '/inventory/warehouses' },
    { label: 'Movements',          icon: ArrowLeftRight,to: '/inventory/movements' },
    { label: 'Inventory Counts',   icon: ClipboardList, to: '/inventory/counts' },
  ]},

  { label: 'Manufacturing', icon: Cog, group: 'Production', children: [
    { label: 'Bill of Materials', icon: ClipboardList, to: '/manufacturing/bom' },
    { label: 'Mfg Orders',        icon: Hammer,        to: '/manufacturing/orders' },
    { label: 'Work Centers',      icon: Factory,       to: '/manufacturing/work-centers' },
    { label: 'MRP Planning',      icon: CalendarDays,  to: '/manufacturing/mrp' },
  ]},

  { label: 'HR & Payroll', icon: User, group: 'Human Resources', children: [
    { label: 'Employees',     icon: IdCard,    to: '/hr/employees' },
    { label: 'Departments',   icon: Globe,     to: '/hr/departments' },
    { label: 'Attendance',    icon: Timer,     to: '/hr/attendance' },
    { label: 'Leave Requests',icon: Calendar,  to: '/hr/leave-requests' },
    { label: 'Payroll',       icon: CreditCard,to: '/hr/payroll' },
    { label: 'Recruitment',   icon: UserPlus,  to: '/hr/recruitment' },
  ]},

  { label: 'Projects', icon: FolderKanban, group: 'Projects', children: [
    { label: 'Projects',   icon: FolderKanban, to: '/projects' },
    { label: 'Tasks',      icon: CheckSquare,  to: '/projects/tasks' },
    { label: 'Timesheets', icon: Timer,        to: '/projects/timesheets' },
    { label: 'Planning',   icon: CalendarDays, to: '/projects/planning' },
    { label: 'Milestones', icon: Flag,         to: '/projects/milestones' },
    { label: 'Expenses',   icon: Wallet,       to: '/projects/expenses' },
    { label: 'Reports',    icon: BarChart3,    to: '/projects/reports' },
  ]},

  { label: 'Treasury', icon: Landmark, group: 'Finance', children: [
    { label: 'Cash Position',      icon: Landmark,        to: '/treasury/cash-position' },
    { label: 'Bank Accounts',      icon: CreditCard,      to: '/treasury/bank-accounts' },
    { label: 'Cheques',            icon: Receipt,         to: '/treasury/cheques' },
    { label: 'Payments',           icon: Banknote,        to: '/treasury/payments' },
    { label: 'Receipts',           icon: ArrowDownToLine, to: '/treasury/receipts' },
    { label: 'Bank Reconciliation',icon: GitMerge,        to: '/treasury/reconciliation' },
    { label: 'Reports',            icon: BarChart2,       to: '/treasury/reports' },
  ]},

  { label: 'Tax', icon: Percent, group: 'Tax', children: [
    { label: 'G50 Declaration', icon: ScrollText,   to: '/tax/g50' },
    { label: 'VAT Register',    icon: FileText,     to: '/tax/vat-register' },
    { label: 'VAT Returns',     icon: FileCheck,    to: '/tax/vat-returns' },
    { label: 'Tax Payments',    icon: Banknote,     to: '/tax/payments' },
    { label: 'Tax Reports',     icon: FileBarChart, to: '/tax/reports' },
  ]},

  { label: 'Reports & BI', icon: FileSearch, group: 'Reports', children: [
    { label: 'BI Dashboard',       icon: BarChart3,    to: '/reports/bi-dashboard' },
    { label: 'Financial Reports',  icon: FileBarChart, to: '/reports/financial' },
    { label: 'Sales Reports',      icon: TrendingUp,   to: '/reports/sales' },
    { label: 'Purchase Reports',   icon: ShoppingBag,  to: '/reports/purchase' },
    { label: 'Inventory Reports',  icon: Package,      to: '/reports/inventory' },
    { label: 'Project Reports',    icon: FolderKanban, to: '/reports/projects' },
    { label: 'Management Reports', icon: BarChart2,    to: '/reports/management' },
    { label: 'Analytics',          icon: LineChart,    to: '/reports/analytics' },
  ]},

  { label: 'Maintenance', icon: Wrench, group: 'Maintenance', children: [
    { label: 'Dashboard',             icon: BarChart2,    to: '/maintenance' },
    { label: 'Equipment',             icon: Settings,     to: '/maintenance/equipment' },
    { label: 'Maintenance Requests',  icon: ClipboardList,to: '/maintenance/requests' },
    { label: 'Maintenance Orders',    icon: Wrench,       to: '/maintenance/orders' },
    { label: 'Preventive Maintenance',icon: Shield,       to: '/maintenance/preventive' },
    { label: 'Maintenance Calendar',  icon: Calendar,     to: '/maintenance/calendar' },
    { label: 'Maintenance History',   icon: History,      to: '/maintenance/history' },
    { label: 'Maintenance Reports',   icon: BarChart2,    to: '/maintenance/reports' },
  ]},

  { label: 'Fleet', icon: Truck, group: 'Fleet', children: [
    { label: 'Fleet Dashboard',       icon: BarChart3Icon, to: '/fleet' },
    { label: 'Vehicles',              icon: Truck,         to: '/fleet/vehicles' },
    { label: 'Drivers',               icon: UserCheck,     to: '/fleet/drivers' },
    { label: 'Assignments',           icon: Link,          to: '/fleet/assignments' },
    { label: 'Fuel Logs',             icon: Fuel,          to: '/fleet/fuel' },
    { label: 'Fleet Maintenance',     icon: Wrench,        to: '/fleet/maintenance' },
    { label: 'Vehicle Expenses',      icon: ReceiptIcon,   to: '/fleet/expenses' },
    { label: 'Fleet Reports',         icon: BarChart3Icon, to: '/fleet/reports' },
  ]},

  { label: 'Quality', icon: ShieldCheck, group: 'Quality', children: [
    { label: 'Dashboard',          icon: LayoutDashboard, to: '/quality' },
    { label: 'Inspections',        icon: ClipboardList,   to: '/quality/inspections' },
    { label: 'Quality Checks',     icon: ClipboardCheck,  to: '/quality/checks' },
    { label: 'Non-Conformities',   icon: AlertOctagon,    to: '/quality/non-conformities' },
    { label: 'Corrective Actions', icon: Wrench,          to: '/quality/corrective-actions' },
    { label: 'Reports',            icon: BarChart3,       to: '/quality/reports' },
  ]},

  { label: 'Settings', icon: Settings, group: 'System', children: [
    { label: 'Companies',          icon: Building,   to: '/settings/companies' },
    { label: 'Users & Roles',      icon: Users,      to: '/settings/users' },
    { label: 'Fiscal Years',       icon: Calendar,   to: '/settings/fiscal-years' },
    { label: 'Workflow',           icon: Workflow,   to: '/settings/workflow' },
    { label: 'Currencies',         icon: DollarSign, to: '/settings/currencies' },
    { label: 'Numbering',          icon: Hash,       to: '/settings/numbering' },
    { label: 'Taxes',              icon: Percent,    to: '/settings/taxes' },
    { label: 'Audit Log',          icon: FileText,   to: '/settings/audit-log' },
    { label: 'System Diagnostics', icon: Activity,   to: '/settings/diagnostics' },
  ]},
]

function isActive(item: NavItem): boolean {
  if (item.to) return route.path === item.to || route.path.startsWith(item.to + '/')
  if (item.children) return item.children.some(child => child.to && (route.path === child.to || route.path.startsWith(child.to + '/')))
  return false
}

function toggleGroup(item: NavItem) {
  const key = item.group || item.label
  expandedGroups.value[key] = !expandedGroups.value[key]
}

function isGroupExpanded(item: NavItem): boolean {
  const key = item.group || item.label
  return !!expandedGroups.value[key]
}
</script>

<template>
  <aside
    class="flex flex-col transition-all duration-300 ease-in-out"
    :class="[
      collapsed ? 'w-16' : 'w-64',
      app.darkMode
        ? 'bg-slate-950 text-slate-100 border-r border-slate-800'
        : 'bg-slate-900 text-slate-100 border-r border-slate-800'
    ]"
  >
    <!-- Logo / Brand -->
    <div class="flex items-center gap-3 px-4 py-4 border-b border-slate-800 min-h-[60px]">
      <div class="w-8 h-8 bg-indigo-500 rounded-lg flex items-center justify-center font-bold text-white text-sm flex-shrink-0">
        N
      </div>
      <transition name="fade-text">
        <span v-if="!collapsed" class="font-bold text-white text-sm truncate">Nexus ERP</span>
      </transition>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 overflow-y-auto py-2 scrollbar-thin">
      <template v-for="item in navItems" :key="item.label">
        <!-- Simple link -->
        <template v-if="!item.children">
          <router-link
            :to="item.to!"
            class="flex items-center gap-3 px-4 py-2.5 mx-2 rounded-lg text-sm transition-colors cursor-pointer"
            :class="isActive(item)
              ? 'bg-indigo-600 text-white'
              : 'text-slate-300 hover:bg-slate-800 hover:text-white'"
          >
            <component :is="item.icon" class="w-5 h-5 flex-shrink-0" />
            <span v-if="!collapsed" class="truncate">{{ item.label }}</span>
          </router-link>
        </template>

        <!-- Expandable group -->
        <template v-else>
          <button
            class="w-full flex items-center gap-3 px-4 py-2.5 mx-0 text-sm transition-colors"
            :class="isActive(item)
              ? 'text-indigo-300'
              : 'text-slate-400 hover:text-slate-200'"
            @click="!collapsed && toggleGroup(item)"
          >
            <component :is="item.icon" class="w-5 h-5 flex-shrink-0 ml-2" />
            <span v-if="!collapsed" class="truncate flex-1 text-left font-medium">{{ item.label }}</span>
            <span v-if="!collapsed" class="text-xs opacity-60">
              <ChevronDown v-if="isGroupExpanded(item)" class="w-4 h-4" />
              <ChevronRight v-else class="w-4 h-4" />
            </span>
          </button>

          <transition name="expand">
            <div v-if="!collapsed && isGroupExpanded(item)" class="ml-4 mb-1">
              <router-link
                v-for="child in item.children"
                :key="child.to"
                :to="child.to!"
                class="flex items-center gap-2 px-4 py-2 mx-2 rounded-lg text-sm transition-colors"
                :class="isActive(child)
                  ? 'bg-indigo-600 text-white'
                  : 'text-slate-400 hover:bg-slate-800 hover:text-slate-200'"
              >
                <component :is="child.icon" class="w-4 h-4 flex-shrink-0" />
                <span class="truncate">{{ child.label }}</span>
              </router-link>
            </div>
          </transition>
        </template>
      </template>
    </nav>

    <!-- Collapse Toggle -->
    <button
      class="flex items-center justify-center py-3 border-t border-slate-800 text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
      @click="app.toggleSidebar()"
    >
      <ChevronRight v-if="collapsed" class="w-5 h-5" />
      <ChevronLeft v-else class="w-5 h-5" />
    </button>
  </aside>
</template>

<style scoped>
.expand-enter-active,
.expand-leave-active {
  transition: all 0.2s ease;
  overflow: hidden;
}
.expand-enter-from,
.expand-leave-to {
  max-height: 0;
  opacity: 0;
}
.expand-enter-to,
.expand-leave-from {
  max-height: 600px;
  opacity: 1;
}
.fade-text-enter-active,
.fade-text-leave-active {
  transition: opacity 0.2s;
}
.fade-text-enter-from,
.fade-text-leave-to {
  opacity: 0;
}
</style>
