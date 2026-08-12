<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { dashboardAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  TrendingUp, DollarSign, Users, ShoppingCart, Package, Building2,
  FolderOpen, CheckCircle, AlertTriangle, BarChart3, Briefcase,
  ClipboardList, Truck, ArrowUpRight, ArrowDownRight, Clock,
  FileText, Activity, RefreshCw, ChevronRight, Layers,
  ReceiptText, Factory, CreditCard, Wallet
} from 'lucide-vue-next'

// ─── Types ────────────────────────────────────────────────────────────────────

interface DashboardSummary {
  monthly_sales: number
  receivables: number
  pipeline_value: number
  won_this_month: number
  treasury_balance: number
  stock_value: number
  payables: number
  monthly_payroll: number
  overdue_invoices: number
  customer_count: number
  open_opportunities: number
  draft_quotations: number
  open_orders: number
  employee_count: number
  active_projects: number
  active_mfg_orders: number
  pending_approvals: number
}

interface RecentActivity {
  id: number
  action: string
  entity_type: string
  entity_id: string
  user_name: string
  module: string
  created_at: string
}

interface CashFlowEntry {
  period: string
  inflow: number
  outflow: number
  net: number
}

// ─── State ────────────────────────────────────────────────────────────────────

const router = useRouter()
const app = useAppStore()

const summary = ref<DashboardSummary | null>(null)
const cashFlow = ref<CashFlowEntry[]>([])
const recentActivity = ref<RecentActivity[]>([])
const approvals = ref<unknown[]>([])
const loading = ref(true)
const refreshing = ref(false)

// ─── Data Loading ─────────────────────────────────────────────────────────────

async function loadData(silent = false) {
  if (!silent) loading.value = true
  else refreshing.value = true
  try {
    const [summaryRes, cfRes, actRes, apprRes] = await Promise.allSettled([
      dashboardAPI.getSummary(),
      dashboardAPI.getCashFlow(),
      dashboardAPI.getActivity(),
      dashboardAPI.getApprovals(),
    ])
    if (summaryRes.status === 'fulfilled') summary.value = summaryRes.value.data
    if (cfRes.status === 'fulfilled') cashFlow.value = cfRes.value.data || []
    if (actRes.status === 'fulfilled') recentActivity.value = actRes.value.data || []
    if (apprRes.status === 'fulfilled') approvals.value = apprRes.value.data || []
  } catch {
    app.addToast('Failed to load dashboard data', 'error')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

// ─── Formatters ───────────────────────────────────────────────────────────────

function fmtCurrency(n: number | undefined | null): string {
  if (n == null) return '—'
  if (Math.abs(n) >= 1_000_000) return (n / 1_000_000).toFixed(2) + ' M DZD'
  if (Math.abs(n) >= 1_000) return (n / 1_000).toFixed(1) + ' k DZD'
  return n.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) + ' DZD'
}

function fmtInt(n: number | undefined | null): string {
  if (n == null) return '—'
  return n.toLocaleString('fr-DZ')
}

function fmtDate(iso: string): string {
  return new Date(iso).toLocaleString('fr-DZ', {
    day: '2-digit', month: '2-digit', year: 'numeric',
    hour: '2-digit', minute: '2-digit',
  })
}

// ─── KPI Cards Config ─────────────────────────────────────────────────────────

const kpiCards = computed(() => {
  const s = summary.value
  return [
    {
      label: 'Monthly Sales',
      value: fmtCurrency(s?.monthly_sales),
      icon: TrendingUp,
      color: 'green',
      to: '/sales/invoices',
      sub: s?.open_orders != null ? `${s.open_orders} open orders` : null,
      trend: 'up',
    },
    {
      label: 'Treasury Balance',
      value: fmtCurrency(s?.treasury_balance),
      icon: Wallet,
      color: 'amber',
      to: '/treasury/cash-position',
      sub: null,
      trend: null,
    },
    {
      label: 'Receivables',
      value: fmtCurrency(s?.receivables),
      icon: ReceiptText,
      color: 'blue',
      to: '/sales/reports/aging',
      sub: s?.overdue_invoices != null ? `${s.overdue_invoices} overdue` : null,
      trend: s?.overdue_invoices ? 'down' : null,
    },
    {
      label: 'Payables',
      value: fmtCurrency(s?.payables),
      icon: CreditCard,
      color: 'purple',
      to: '/purchase/invoices',
      sub: null,
      trend: null,
    },
    {
      label: 'Stock Value',
      value: fmtCurrency(s?.stock_value),
      icon: Package,
      color: 'indigo',
      to: '/inventory/stock-levels',
      sub: null,
      trend: null,
    },
    {
      label: 'Pipeline Value',
      value: fmtCurrency(s?.pipeline_value),
      icon: BarChart3,
      color: 'cyan',
      to: '/sales/pipeline',
      sub: s?.open_opportunities != null ? `${s.open_opportunities} opportunities` : null,
      trend: 'up',
    },
    {
      label: 'Customers',
      value: fmtInt(s?.customer_count),
      icon: Users,
      color: 'teal',
      to: '/sales/customers',
      sub: null,
      trend: null,
    },
    {
      label: 'Employees',
      value: fmtInt(s?.employee_count),
      icon: Briefcase,
      color: 'slate',
      to: '/hr/employees',
      sub: s?.monthly_payroll != null ? fmtCurrency(s.monthly_payroll) + ' payroll' : null,
      trend: null,
    },
    {
      label: 'Active Projects',
      value: fmtInt(s?.active_projects),
      icon: FolderOpen,
      color: 'orange',
      to: '/projects',
      sub: null,
      trend: null,
    },
    {
      label: 'Mfg Orders',
      value: fmtInt(s?.active_mfg_orders),
      icon: Factory,
      color: 'rose',
      to: '/manufacturing/orders',
      sub: null,
      trend: null,
    },
    {
      label: 'Draft Quotations',
      value: fmtInt(s?.draft_quotations),
      icon: FileText,
      color: 'sky',
      to: '/sales/quotations',
      sub: null,
      trend: null,
    },
    {
      label: 'Pending Approvals',
      value: fmtInt(s?.pending_approvals),
      icon: CheckCircle,
      color: 'red',
      to: '/settings/workflow',
      sub: null,
      trend: s?.pending_approvals ? 'warn' : null,
    },
  ]
})

const colorClasses: Record<string, { card: string; icon: string; badge: string }> = {
  green:  { card: 'border-green-200 dark:border-green-800',  icon: 'bg-green-100 dark:bg-green-900/40 text-green-700 dark:text-green-400',  badge: 'text-green-600 dark:text-green-400' },
  amber:  { card: 'border-amber-200 dark:border-amber-800',  icon: 'bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400',  badge: 'text-amber-600 dark:text-amber-400' },
  blue:   { card: 'border-blue-200 dark:border-blue-800',    icon: 'bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-400',    badge: 'text-blue-600 dark:text-blue-400' },
  purple: { card: 'border-purple-200 dark:border-purple-800',icon: 'bg-purple-100 dark:bg-purple-900/40 text-purple-700 dark:text-purple-400',badge: 'text-purple-600 dark:text-purple-400' },
  indigo: { card: 'border-indigo-200 dark:border-indigo-800',icon: 'bg-indigo-100 dark:bg-indigo-900/40 text-indigo-700 dark:text-indigo-400',badge: 'text-indigo-600 dark:text-indigo-400' },
  cyan:   { card: 'border-cyan-200 dark:border-cyan-800',    icon: 'bg-cyan-100 dark:bg-cyan-900/40 text-cyan-700 dark:text-cyan-400',    badge: 'text-cyan-600 dark:text-cyan-400' },
  teal:   { card: 'border-teal-200 dark:border-teal-800',    icon: 'bg-teal-100 dark:bg-teal-900/40 text-teal-700 dark:text-teal-400',    badge: 'text-teal-600 dark:text-teal-400' },
  slate:  { card: 'border-slate-200 dark:border-slate-700',  icon: 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400',   badge: 'text-slate-600 dark:text-slate-400' },
  orange: { card: 'border-orange-200 dark:border-orange-800',icon: 'bg-orange-100 dark:bg-orange-900/40 text-orange-700 dark:text-orange-400',badge: 'text-orange-600 dark:text-orange-400' },
  rose:   { card: 'border-rose-200 dark:border-rose-800',    icon: 'bg-rose-100 dark:bg-rose-900/40 text-rose-700 dark:text-rose-400',    badge: 'text-rose-600 dark:text-rose-400' },
  sky:    { card: 'border-sky-200 dark:border-sky-800',      icon: 'bg-sky-100 dark:bg-sky-900/40 text-sky-700 dark:text-sky-400',       badge: 'text-sky-600 dark:text-sky-400' },
  red:    { card: 'border-red-200 dark:border-red-800',      icon: 'bg-red-100 dark:bg-red-900/40 text-red-700 dark:text-red-400',       badge: 'text-red-600 dark:text-red-400' },
}

// ─── Quick Actions ────────────────────────────────────────────────────────────

const quickActions = [
  { label: 'New Quotation',    icon: FileText,     to: '/sales/quotations',  color: 'blue' },
  { label: 'New Invoice',      icon: ReceiptText,  to: '/sales/invoices',    color: 'green' },
  { label: 'New Customer',     icon: Users,        to: '/sales/customers',   color: 'teal' },
  { label: 'New Purchase',     icon: ShoppingCart, to: '/purchase/orders',   color: 'purple' },
  { label: 'Journal Entry',    icon: ClipboardList,to: '/accounting/journal-entries', color: 'amber' },
  { label: 'Stock Transfer',   icon: Truck,        to: '/inventory/movements', color: 'indigo' },
  { label: 'New Employee',     icon: Briefcase,    to: '/hr/employees',      color: 'slate' },
  { label: 'New Project',      icon: FolderOpen,   to: '/projects',          color: 'orange' },
]

const qaColor: Record<string, string> = {
  blue:   'bg-blue-50 hover:bg-blue-100 text-blue-700 border-blue-200 dark:bg-blue-900/20 dark:hover:bg-blue-900/40 dark:text-blue-400 dark:border-blue-800',
  green:  'bg-green-50 hover:bg-green-100 text-green-700 border-green-200 dark:bg-green-900/20 dark:hover:bg-green-900/40 dark:text-green-400 dark:border-green-800',
  teal:   'bg-teal-50 hover:bg-teal-100 text-teal-700 border-teal-200 dark:bg-teal-900/20 dark:hover:bg-teal-900/40 dark:text-teal-400 dark:border-teal-800',
  purple: 'bg-purple-50 hover:bg-purple-100 text-purple-700 border-purple-200 dark:bg-purple-900/20 dark:hover:bg-purple-900/40 dark:text-purple-400 dark:border-purple-800',
  amber:  'bg-amber-50 hover:bg-amber-100 text-amber-700 border-amber-200 dark:bg-amber-900/20 dark:hover:bg-amber-900/40 dark:text-amber-400 dark:border-amber-800',
  indigo: 'bg-indigo-50 hover:bg-indigo-100 text-indigo-700 border-indigo-200 dark:bg-indigo-900/20 dark:hover:bg-indigo-900/40 dark:text-indigo-400 dark:border-indigo-800',
  slate:  'bg-slate-50 hover:bg-slate-100 text-slate-700 border-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 dark:border-slate-600',
  orange: 'bg-orange-50 hover:bg-orange-100 text-orange-700 border-orange-200 dark:bg-orange-900/20 dark:hover:bg-orange-900/40 dark:text-orange-400 dark:border-orange-800',
}

// ─── Activity helpers ─────────────────────────────────────────────────────────

function activityIcon(module: string) {
  const m: Record<string, unknown> = {
    sales: TrendingUp, purchase: ShoppingCart, accounting: ClipboardList,
    hr: Briefcase, inventory: Package, manufacturing: Factory,
    projects: FolderOpen, treasury: Wallet,
  }
  return m[module] ?? Activity
}

function activityColor(module: string): string {
  const m: Record<string, string> = {
    sales: 'text-blue-600 dark:text-blue-400 bg-blue-100 dark:bg-blue-900/40',
    purchase: 'text-purple-600 dark:text-purple-400 bg-purple-100 dark:bg-purple-900/40',
    accounting: 'text-amber-600 dark:text-amber-400 bg-amber-100 dark:bg-amber-900/40',
    hr: 'text-green-600 dark:text-green-400 bg-green-100 dark:bg-green-900/40',
    inventory: 'text-indigo-600 dark:text-indigo-400 bg-indigo-100 dark:bg-indigo-900/40',
    manufacturing: 'text-rose-600 dark:text-rose-400 bg-rose-100 dark:bg-rose-900/40',
    projects: 'text-orange-600 dark:text-orange-400 bg-orange-100 dark:bg-orange-900/40',
    treasury: 'text-teal-600 dark:text-teal-400 bg-teal-100 dark:bg-teal-900/40',
  }
  return m[module] ?? 'text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800'
}

onMounted(() => loadData())
</script>

<template>
  <div class="space-y-6">

    <!-- ─── Header ──────────────────────────────────────────────────────────── -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Dashboard</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
          Executive overview — real-time data
        </p>
      </div>
      <button
        @click="loadData(true)"
        :disabled="refreshing"
        class="inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-gray-200 dark:border-gray-700
               bg-white dark:bg-gray-800 text-sm font-medium text-gray-700 dark:text-gray-200
               hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors disabled:opacity-60"
      >
        <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': refreshing }" />
        Refresh
      </button>
    </div>

    <!-- ─── Loading skeleton ─────────────────────────────────────────────────── -->
    <div v-if="loading" class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
      <div
        v-for="i in 12" :key="i"
        class="h-28 rounded-xl bg-gray-200 dark:bg-gray-800 animate-pulse"
      />
    </div>

    <!-- ─── KPI Grid ──────────────────────────────────────────────────────────── -->
    <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-4">
      <button
        v-for="kpi in kpiCards"
        :key="kpi.label"
        @click="router.push(kpi.to)"
        class="group relative flex flex-col gap-2 p-4 rounded-xl border bg-white dark:bg-gray-900
               shadow-sm hover:shadow-md transition-all duration-200 text-left cursor-pointer"
        :class="colorClasses[kpi.color].card"
      >
        <!-- Icon + trend arrow -->
        <div class="flex items-start justify-between">
          <div class="p-2 rounded-lg" :class="colorClasses[kpi.color].icon">
            <component :is="kpi.icon" class="w-5 h-5" />
          </div>
          <div class="flex items-center gap-1 text-xs font-medium" v-if="kpi.trend">
            <ArrowUpRight v-if="kpi.trend === 'up'" class="w-4 h-4 text-green-500" />
            <ArrowDownRight v-else-if="kpi.trend === 'down'" class="w-4 h-4 text-red-500" />
            <AlertTriangle v-else-if="kpi.trend === 'warn'" class="w-4 h-4 text-amber-500" />
          </div>
          <ChevronRight v-else class="w-4 h-4 text-gray-300 dark:text-gray-600 group-hover:translate-x-0.5 transition-transform" />
        </div>
        <!-- Value -->
        <div>
          <p class="text-lg font-bold text-gray-900 dark:text-white leading-tight truncate">
            {{ kpi.value }}
          </p>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ kpi.label }}</p>
        </div>
        <!-- Sub-label -->
        <p v-if="kpi.sub" class="text-xs font-medium" :class="colorClasses[kpi.color].badge">
          {{ kpi.sub }}
        </p>
      </button>
    </div>

    <!-- ─── Row 2: Quick Actions + Cash Flow + Pending Approvals ─────────────── -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">

      <!-- Quick Actions -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 p-5 shadow-sm">
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4 flex items-center gap-2">
          <Layers class="w-4 h-4 text-gray-400" />
          Quick Actions
        </h2>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="action in quickActions"
            :key="action.label"
            @click="router.push(action.to)"
            class="flex items-center gap-2 px-3 py-2.5 rounded-lg border text-xs font-medium transition-colors"
            :class="qaColor[action.color]"
          >
            <component :is="action.icon" class="w-4 h-4 flex-shrink-0" />
            <span class="truncate">{{ action.label }}</span>
          </button>
        </div>
      </div>

      <!-- Cash Flow Chart (table representation) -->
      <div class="lg:col-span-2 bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 p-5 shadow-sm">
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4 flex items-center gap-2">
          <BarChart3 class="w-4 h-4 text-gray-400" />
          Cash Flow — Last 6 Months
        </h2>
        <div v-if="cashFlow.length === 0" class="flex flex-col items-center justify-center h-32 gap-2 text-gray-400 dark:text-gray-600">
          <BarChart3 class="w-8 h-8" />
          <p class="text-sm">No cash flow data available</p>
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-xs text-gray-500 dark:text-gray-400 border-b border-gray-100 dark:border-gray-800">
                <th class="pb-2 font-medium">Period</th>
                <th class="pb-2 font-medium text-right">Inflow</th>
                <th class="pb-2 font-medium text-right">Outflow</th>
                <th class="pb-2 font-medium text-right">Net</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-50 dark:divide-gray-800">
              <tr v-for="row in cashFlow" :key="row.period" class="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                <td class="py-2 text-gray-700 dark:text-gray-300 font-medium">{{ row.period }}</td>
                <td class="py-2 text-right text-green-600 dark:text-green-400">{{ fmtCurrency(row.inflow) }}</td>
                <td class="py-2 text-right text-red-600 dark:text-red-400">{{ fmtCurrency(row.outflow) }}</td>
                <td class="py-2 text-right font-semibold" :class="row.net >= 0 ? 'text-green-700 dark:text-green-400' : 'text-red-700 dark:text-red-400'">
                  {{ fmtCurrency(row.net) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ─── Row 3: Recent Activity + Pending Approvals ─────────────────────── -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

      <!-- Recent Activity -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 p-5 shadow-sm">
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4 flex items-center gap-2">
          <Activity class="w-4 h-4 text-gray-400" />
          Recent Activity
        </h2>
        <div v-if="recentActivity.length === 0" class="flex flex-col items-center justify-center h-32 gap-2 text-gray-400 dark:text-gray-600">
          <Activity class="w-8 h-8" />
          <p class="text-sm">No recent activity</p>
        </div>
        <ul v-else class="space-y-3">
          <li
            v-for="item in recentActivity.slice(0, 10)"
            :key="item.id"
            class="flex items-start gap-3"
          >
            <div class="mt-0.5 p-1.5 rounded-lg flex-shrink-0" :class="activityColor(item.module)">
              <component :is="activityIcon(item.module)" class="w-3.5 h-3.5" />
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm text-gray-800 dark:text-gray-200 truncate">
                <span class="font-medium capitalize">{{ item.action }}</span>
                <span class="text-gray-500 dark:text-gray-400"> · {{ item.entity_type }}</span>
              </p>
              <p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">
                {{ item.user_name }} · {{ fmtDate(item.created_at) }}
              </p>
            </div>
          </li>
        </ul>
      </div>

      <!-- Pending Approvals -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 p-5 shadow-sm">
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-4 flex items-center gap-2">
          <CheckCircle class="w-4 h-4 text-gray-400" />
          Pending Approvals
          <span
            v-if="approvals.length > 0"
            class="ml-auto inline-flex items-center justify-center w-5 h-5 rounded-full bg-red-100 dark:bg-red-900/40 text-red-700 dark:text-red-400 text-xs font-bold"
          >
            {{ approvals.length }}
          </span>
        </h2>
        <div v-if="approvals.length === 0" class="flex flex-col items-center justify-center h-32 gap-2 text-gray-400 dark:text-gray-600">
          <CheckCircle class="w-8 h-8" />
          <p class="text-sm">No pending approvals</p>
        </div>
        <ul v-else class="space-y-2">
          <li
            v-for="(appr, idx) in (approvals as any[]).slice(0, 8)"
            :key="idx"
            class="flex items-center justify-between gap-2 p-3 rounded-lg bg-gray-50 dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer transition-colors"
            @click="router.push('/settings/workflow')"
          >
            <div class="flex items-center gap-2 min-w-0">
              <Clock class="w-4 h-4 text-amber-500 flex-shrink-0" />
              <div class="min-w-0">
                <p class="text-sm font-medium text-gray-800 dark:text-gray-200 truncate">
                  {{ appr.document_type || appr.entity_type || 'Approval request' }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400 truncate">
                  {{ appr.requested_by || appr.user_name || '' }}
                </p>
              </div>
            </div>
            <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400 flex-shrink-0">
              Pending
            </span>
          </li>
        </ul>
        <button
          v-if="approvals.length > 0"
          @click="router.push('/settings/workflow')"
          class="mt-3 w-full text-sm text-blue-600 dark:text-blue-400 hover:underline flex items-center justify-center gap-1"
        >
          View all approvals
          <ChevronRight class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- ─── Won This Month banner ─────────────────────────────────────────────── -->
    <div
      v-if="summary?.won_this_month != null && summary.won_this_month > 0"
      class="flex items-center gap-3 px-5 py-4 rounded-xl
             bg-gradient-to-r from-green-50 to-emerald-50
             dark:from-green-900/20 dark:to-emerald-900/20
             border border-green-200 dark:border-green-800"
    >
      <TrendingUp class="w-5 h-5 text-green-600 dark:text-green-400 flex-shrink-0" />
      <p class="text-sm font-medium text-green-800 dark:text-green-300">
        Won this month:
        <span class="font-bold text-green-700 dark:text-green-400 ml-1">
          {{ fmtCurrency(summary.won_this_month) }}
        </span>
      </p>
      <button
        @click="router.push('/sales/pipeline')"
        class="ml-auto text-xs text-green-700 dark:text-green-400 hover:underline flex items-center gap-1"
      >
        View pipeline <ChevronRight class="w-3 h-3" />
      </button>
    </div>

  </div>
</template>
