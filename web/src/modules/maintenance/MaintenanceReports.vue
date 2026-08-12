<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  BarChart2, RefreshCw, Calendar, DollarSign,
  Wrench, CheckCircle, Clock, AlertTriangle,
  TrendingUp, TrendingDown, Package, Shield,
  ChevronDown, Settings, FileText
} from 'lucide-vue-next'
import { maintenanceAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── types ────────────────────────────────────────────────────────────────────
interface DashboardData {
  equipment_summary: {
    total: number
    operational: number
    under_maintenance: number
    out_of_service: number
    retired: number
  }
  orders_summary: {
    total: number
    pending: number
    in_progress: number
    completed: number
    cancelled: number
    overdue: number
  }
  cost_summary: {
    total_cost: number
    labor_cost: number
    parts_cost: number
    other_cost: number
    avg_cost_per_order: number
  }
  mttr: number
  uptime_rate: number
  pm_compliance: number
  by_type: Array<{ type: string; count: number; cost: number }>
  by_equipment: Array<{ equipment_name: string; orders: number; total_cost: number }>
  monthly_costs: Array<{ month: string; cost: number; orders: number }>
  recent_orders: Array<{
    id: string; order_number: string; title: string
    status: string; order_type: string; scheduled_date?: string
    equipment_name?: string
  }>
}

// ─── state ────────────────────────────────────────────────────────────────────
const data    = ref<DashboardData | null>(null)
const loading = ref(false)
const year    = ref(new Date().getFullYear())

const years = computed(() => {
  const y = new Date().getFullYear()
  return [y, y-1, y-2, y-3]
})

// ─── computed ────────────────────────────────────────────────────────────────
const dk = (a: string, b: string) => app.darkMode ? a : b

const equipSummary = computed(() => data.value?.equipment_summary ?? {
  total:0, operational:0, under_maintenance:0, out_of_service:0, retired:0
})

const ordersSummary = computed(() => data.value?.orders_summary ?? {
  total:0, pending:0, in_progress:0, completed:0, cancelled:0, overdue:0
})

const costSummary = computed(() => data.value?.cost_summary ?? {
  total_cost:0, labor_cost:0, parts_cost:0, other_cost:0, avg_cost_per_order:0
})

const byType = computed(() => data.value?.by_type ?? [])

const byEquipment = computed(() => (data.value?.by_equipment ?? []).slice(0, 10))

const monthlyCosts = computed(() => data.value?.monthly_costs ?? [])

const maxMonthlyCost = computed(() =>
  Math.max(...monthlyCosts.value.map(m => m.cost), 1)
)

const recentOrders = computed(() => data.value?.recent_orders ?? [])

// ─── helpers ─────────────────────────────────────────────────────────────────
const fmt = (n: number) =>
  new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'

const fmtDate = (s?: string) => {
  if (!s) return '—'
  return new Date(s).toLocaleDateString('fr-DZ', { day:'2-digit', month:'short' })
}

const pct = (n: number, total: number) =>
  total > 0 ? Math.round((n / total) * 100) : 0

const typeColor = (t: string) => ({
  corrective:'#f43f5e', preventive:'#10b981',
  inspection:'#8b5cf6', emergency:'#f97316',
  upgrade:'#f59e0b', other:'#64748b',
}[t] ?? '#64748b')

const typeLabel = (t: string) => ({
  corrective:'Corrective', preventive:'Preventive',
  inspection:'Inspection', emergency:'Emergency',
  upgrade:'Upgrade', other:'Other',
}[t] ?? t)

const statusBadge = (s: string) => ({
  pending:'bg-slate-500/15 text-slate-400',
  in_progress:'bg-blue-500/15 text-blue-400',
  on_hold:'bg-amber-500/15 text-amber-400',
  completed:'bg-emerald-500/15 text-emerald-400',
  cancelled:'bg-rose-500/15 text-rose-400',
}[s] ?? 'bg-slate-500/15 text-slate-400')

const statusLabel = (s: string) => ({
  pending:'Pending', in_progress:'In Progress', on_hold:'On Hold',
  completed:'Completed', cancelled:'Cancelled',
}[s] ?? s)

const monthShort = (m: string) => {
  const months = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec']
  const idx = parseInt(m.split('-')[1] || '1') - 1
  return months[idx] ?? m
}

// ─── data loading ─────────────────────────────────────────────────────────────
const load = async () => {
  loading.value = true
  try {
    const res = await maintenanceAPI.getReports({ year: String(year.value) })
    data.value = res.data
  } catch {
    app.addToast('Failed to load maintenance reports', 'error')
  } finally {
    loading.value = false
  }
}

watch(year, load)
onMounted(load)
</script>

<template>
  <div :class="['min-h-screen p-6 space-y-6', dk('bg-slate-950 text-slate-100','bg-slate-50 text-slate-900')]">

    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-4">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-violet-500/15 flex items-center justify-center">
          <BarChart2 class="w-5 h-5 text-violet-400" />
        </div>
        <div>
          <h1 class="text-xl font-bold">Maintenance Reports</h1>
          <p :class="['text-sm', dk('text-slate-400','text-slate-500')]">KPIs, costs, and performance analytics</p>
        </div>
      </div>
      <div class="flex items-center gap-3">
        <select v-model="year"
          :class="['px-3 py-2 rounded-lg border text-sm',
            dk('bg-slate-900 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')]">
          <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
        </select>
        <button @click="load"
          :class="['p-2 rounded-lg border transition-colors',
            dk('bg-slate-900 border-slate-700 text-slate-300 hover:bg-slate-800',
               'bg-white border-slate-200 text-slate-700 hover:bg-slate-50')]">
          <RefreshCw :class="['w-4 h-4', loading && 'animate-spin']" />
        </button>
      </div>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-20">
      <RefreshCw class="w-8 h-8 animate-spin text-violet-400" />
    </div>

    <template v-else-if="data">

      <!-- ── Top KPI Row ─────────────────────────────────────────────────── -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <div :class="['rounded-xl border p-4', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
          <div class="flex items-center justify-between mb-3">
            <span :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]">Total Cost</span>
            <div class="w-8 h-8 rounded-lg bg-teal-500/10 flex items-center justify-center">
              <DollarSign class="w-4 h-4 text-teal-400" />
            </div>
          </div>
          <div class="text-xl font-bold">{{ fmt(costSummary.total_cost) }}</div>
          <div :class="['text-xs mt-1', dk('text-slate-500','text-slate-400')]">Avg {{ fmt(costSummary.avg_cost_per_order) }} / order</div>
        </div>

        <div :class="['rounded-xl border p-4', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
          <div class="flex items-center justify-between mb-3">
            <span :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]">MTTR</span>
            <div class="w-8 h-8 rounded-lg bg-amber-500/10 flex items-center justify-center">
              <Clock class="w-4 h-4 text-amber-400" />
            </div>
          </div>
          <div class="text-xl font-bold">{{ data.mttr?.toFixed(1) ?? '—' }}h</div>
          <div :class="['text-xs mt-1', dk('text-slate-500','text-slate-400')]">Mean time to repair</div>
        </div>

        <div :class="['rounded-xl border p-4', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
          <div class="flex items-center justify-between mb-3">
            <span :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]">Uptime Rate</span>
            <div class="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center">
              <TrendingUp class="w-4 h-4 text-emerald-400" />
            </div>
          </div>
          <div class="text-xl font-bold">{{ data.uptime_rate?.toFixed(1) ?? '—' }}%</div>
          <div class="mt-2 h-1.5 rounded-full" :class="dk('bg-slate-700','bg-slate-200')">
            <div class="h-full rounded-full bg-emerald-400 transition-all"
              :style="{ width: (data.uptime_rate || 0) + '%' }"></div>
          </div>
        </div>

        <div :class="['rounded-xl border p-4', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
          <div class="flex items-center justify-between mb-3">
            <span :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]">PM Compliance</span>
            <div class="w-8 h-8 rounded-lg bg-violet-500/10 flex items-center justify-center">
              <Shield class="w-4 h-4 text-violet-400" />
            </div>
          </div>
          <div class="text-xl font-bold">{{ data.pm_compliance?.toFixed(1) ?? '—' }}%</div>
          <div class="mt-2 h-1.5 rounded-full" :class="dk('bg-slate-700','bg-slate-200')">
            <div class="h-full rounded-full bg-violet-400 transition-all"
              :style="{ width: (data.pm_compliance || 0) + '%' }"></div>
          </div>
        </div>
      </div>

      <!-- ── Equipment + Orders Summaries ───────────────────────────────── -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">

        <!-- Equipment Status -->
        <div :class="['rounded-xl border p-5', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
          <div class="flex items-center gap-2 mb-4">
            <Settings class="w-4 h-4 text-slate-400" />
            <h2 class="font-semibold text-sm">Equipment Status</h2>
            <span :class="['ml-auto text-xs font-bold', dk('text-slate-400','text-slate-500')]">
              {{ equipSummary.total }} total
            </span>
          </div>
          <div class="space-y-3">
            <div v-for="item in [
              { label: 'Operational',       value: equipSummary.operational,        color: 'bg-emerald-400', text: 'text-emerald-400' },
              { label: 'Under Maintenance', value: equipSummary.under_maintenance,  color: 'bg-amber-400',   text: 'text-amber-400' },
              { label: 'Out of Service',    value: equipSummary.out_of_service,     color: 'bg-rose-400',    text: 'text-rose-400' },
              { label: 'Retired',           value: equipSummary.retired,            color: 'bg-slate-500',   text: 'text-slate-400' },
            ]" :key="item.label">
              <div class="flex items-center gap-3">
                <span :class="['text-xs w-36 flex-shrink-0', dk('text-slate-400','text-slate-500')]">{{ item.label }}</span>
                <div class="flex-1 h-2 rounded-full" :class="dk('bg-slate-800','bg-slate-200')">
                  <div :class="['h-full rounded-full transition-all', item.color]"
                    :style="{ width: pct(item.value, equipSummary.total) + '%' }"></div>
                </div>
                <span :class="['text-xs font-semibold w-8 text-right', item.text]">{{ item.value }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Orders Status -->
        <div :class="['rounded-xl border p-5', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
          <div class="flex items-center gap-2 mb-4">
            <Wrench class="w-4 h-4 text-slate-400" />
            <h2 class="font-semibold text-sm">Work Orders — {{ year }}</h2>
            <span :class="['ml-auto text-xs font-bold', dk('text-slate-400','text-slate-500')]">
              {{ ordersSummary.total }} total
            </span>
          </div>
          <div class="grid grid-cols-3 gap-3">
            <div v-for="item in [
              { label: 'Pending',     value: ordersSummary.pending,     color: 'text-slate-400',   bg: 'bg-slate-500/10' },
              { label: 'In Progress', value: ordersSummary.in_progress, color: 'text-blue-400',    bg: 'bg-blue-500/10' },
              { label: 'Completed',   value: ordersSummary.completed,   color: 'text-emerald-400', bg: 'bg-emerald-500/10' },
              { label: 'Overdue',     value: ordersSummary.overdue,     color: 'text-rose-400',    bg: 'bg-rose-500/10' },
              { label: 'Cancelled',   value: ordersSummary.cancelled,   color: 'text-slate-400',   bg: 'bg-slate-500/10' },
              { label: 'Total',       value: ordersSummary.total,       color: 'text-violet-400',  bg: 'bg-violet-500/10' },
            ]" :key="item.label"
              :class="['rounded-xl p-3 text-center', item.bg]">
              <div :class="['text-xl font-bold', item.color]">{{ item.value }}</div>
              <div :class="['text-xs mt-0.5', dk('text-slate-400','text-slate-500')]">{{ item.label }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- ── Cost Breakdown + Type Distribution ─────────────────────────── -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">

        <!-- Cost Breakdown -->
        <div :class="['rounded-xl border p-5', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
          <div class="flex items-center gap-2 mb-4">
            <DollarSign class="w-4 h-4 text-teal-400" />
            <h2 class="font-semibold text-sm">Cost Breakdown — {{ year }}</h2>
          </div>
          <div class="space-y-4">
            <div v-for="item in [
              { label: 'Labor Cost',  value: costSummary.labor_cost,  color: 'bg-blue-400',    text: 'text-blue-400' },
              { label: 'Parts Cost',  value: costSummary.parts_cost,  color: 'bg-amber-400',   text: 'text-amber-400' },
              { label: 'Other Cost',  value: costSummary.other_cost,  color: 'bg-violet-400',  text: 'text-violet-400' },
            ]" :key="item.label">
              <div class="flex items-center justify-between text-xs mb-1">
                <span :class="dk('text-slate-400','text-slate-500')">{{ item.label }}</span>
                <span :class="['font-semibold', item.text]">{{ fmt(item.value) }}</span>
              </div>
              <div class="h-2 rounded-full" :class="dk('bg-slate-800','bg-slate-200')">
                <div :class="['h-full rounded-full transition-all', item.color]"
                  :style="{ width: pct(item.value, costSummary.total_cost) + '%' }"></div>
              </div>
              <div :class="['text-xs mt-0.5 text-right', dk('text-slate-500','text-slate-400')]">
                {{ pct(item.value, costSummary.total_cost) }}%
              </div>
            </div>
            <div :class="['pt-3 border-t flex items-center justify-between', dk('border-slate-700','border-slate-200')]">
              <span :class="['text-xs font-medium', dk('text-slate-300','text-slate-700')]">Total</span>
              <span class="font-bold text-teal-400">{{ fmt(costSummary.total_cost) }}</span>
            </div>
          </div>
        </div>

        <!-- By Type -->
        <div :class="['rounded-xl border p-5', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
          <div class="flex items-center gap-2 mb-4">
            <BarChart2 class="w-4 h-4 text-violet-400" />
            <h2 class="font-semibold text-sm">Orders by Type</h2>
          </div>
          <div v-if="!byType.length" class="text-center py-8">
            <p :class="['text-sm', dk('text-slate-500','text-slate-400')]">No data for selected year</p>
          </div>
          <div v-else class="space-y-3">
            <div v-for="t in byType" :key="t.type">
              <div class="flex items-center justify-between text-xs mb-1">
                <div class="flex items-center gap-2">
                  <div class="w-2 h-2 rounded-full" :style="{ backgroundColor: typeColor(t.type) }"></div>
                  <span :class="dk('text-slate-300','text-slate-700')">{{ typeLabel(t.type) }}</span>
                </div>
                <div class="flex items-center gap-3">
                  <span :class="['font-medium', dk('text-slate-400','text-slate-500')]">{{ t.count }} orders</span>
                  <span class="font-semibold" :style="{ color: typeColor(t.type) }">{{ fmt(t.cost) }}</span>
                </div>
              </div>
              <div class="h-1.5 rounded-full" :class="dk('bg-slate-800','bg-slate-200')">
                <div class="h-full rounded-full transition-all"
                  :style="{ backgroundColor: typeColor(t.type), width: pct(t.count, ordersSummary.total) + '%' }"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ── Monthly Cost Chart ─────────────────────────────────────────── -->
      <div :class="['rounded-xl border p-5', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
        <div class="flex items-center gap-2 mb-5">
          <TrendingUp class="w-4 h-4 text-violet-400" />
          <h2 class="font-semibold text-sm">Monthly Cost Trend — {{ year }}</h2>
        </div>
        <div v-if="!monthlyCosts.length" class="text-center py-8">
          <p :class="['text-sm', dk('text-slate-500','text-slate-400')]">No monthly data available</p>
        </div>
        <div v-else class="flex items-end gap-2 h-40">
          <div v-for="m in monthlyCosts" :key="m.month"
            class="flex-1 flex flex-col items-center gap-1 min-w-0">
            <div :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]" style="writing-mode:initial">
              {{ m.orders }}
            </div>
            <div class="w-full relative flex items-end" style="height:100px">
              <div class="w-full rounded-t-md transition-all bg-violet-500/70 hover:bg-violet-500 cursor-default"
                :style="{ height: Math.max(4, (m.cost / maxMonthlyCost) * 100) + 'px' }"
                :title="`${monthShort(m.month)}: ${m.orders} orders, ${m.cost.toLocaleString('fr-DZ')} DZD`">
              </div>
            </div>
            <div :class="['text-xs', dk('text-slate-400','text-slate-500')]">{{ monthShort(m.month) }}</div>
          </div>
        </div>
      </div>

      <!-- ── Top Equipment + Recent Orders ──────────────────────────────── -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">

        <!-- Top Equipment by Cost -->
        <div :class="['rounded-xl border p-5', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
          <div class="flex items-center gap-2 mb-4">
            <Package class="w-4 h-4 text-amber-400" />
            <h2 class="font-semibold text-sm">Top Equipment by Cost</h2>
          </div>
          <div v-if="!byEquipment.length" class="text-center py-8">
            <p :class="['text-sm', dk('text-slate-500','text-slate-400')]">No equipment data</p>
          </div>
          <div v-else class="space-y-2">
            <div v-for="(e, idx) in byEquipment" :key="e.equipment_name"
              :class="['flex items-center gap-3 p-2.5 rounded-xl text-sm',
                dk('hover:bg-slate-800','hover:bg-slate-50')]">
              <div :class="['w-6 h-6 rounded-lg flex items-center justify-center text-xs font-bold flex-shrink-0',
                idx === 0 ? 'bg-amber-500/20 text-amber-400' :
                idx === 1 ? 'bg-slate-600/40 text-slate-300' :
                idx === 2 ? 'bg-orange-700/20 text-orange-400' :
                dk('bg-slate-800 text-slate-500','bg-slate-100 text-slate-400')]">
                {{ idx+1 }}
              </div>
              <div class="flex-1 min-w-0">
                <div class="font-medium text-sm truncate">{{ e.equipment_name }}</div>
                <div :class="['text-xs', dk('text-slate-500','text-slate-400')]">{{ e.orders }} orders</div>
              </div>
              <div class="text-right flex-shrink-0">
                <div class="font-semibold text-teal-400 text-sm">{{ fmt(e.total_cost) }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Recent Orders -->
        <div :class="['rounded-xl border p-5', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
          <div class="flex items-center gap-2 mb-4">
            <FileText class="w-4 h-4 text-blue-400" />
            <h2 class="font-semibold text-sm">Recent Orders</h2>
          </div>
          <div v-if="!recentOrders.length" class="text-center py-8">
            <p :class="['text-sm', dk('text-slate-500','text-slate-400')]">No recent orders</p>
          </div>
          <div v-else class="space-y-2">
            <div v-for="o in recentOrders.slice(0,8)" :key="o.id"
              :class="['flex items-start gap-3 p-2.5 rounded-xl text-sm',
                dk('hover:bg-slate-800','hover:bg-slate-50')]">
              <div class="w-1.5 h-1.5 rounded-full mt-2 flex-shrink-0"
                :style="{ backgroundColor: typeColor(o.order_type) }"></div>
              <div class="flex-1 min-w-0">
                <div class="font-medium text-sm truncate">{{ o.title }}</div>
                <div class="flex items-center gap-2 mt-0.5">
                  <span class="font-mono text-xs text-violet-400">{{ o.order_number }}</span>
                  <span v-if="o.equipment_name" :class="['text-xs', dk('text-slate-500','text-slate-400')]">
                    • {{ o.equipment_name }}
                  </span>
                </div>
              </div>
              <div class="flex-shrink-0 text-right">
                <span :class="['px-2 py-0.5 rounded-md text-xs font-medium', statusBadge(o.status)]">
                  {{ statusLabel(o.status) }}
                </span>
                <div :class="['text-xs mt-1', dk('text-slate-500','text-slate-400')]">{{ fmtDate(o.scheduled_date) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Empty state -->
    <div v-else-if="!loading" class="text-center py-20">
      <BarChart2 class="w-12 h-12 mx-auto mb-3 text-slate-500" />
      <p :class="['font-medium', dk('text-slate-300','text-slate-700')]">No report data available</p>
      <button @click="load" class="mt-3 text-sm text-violet-400 hover:underline">Retry</button>
    </div>

  </div>
</template>
