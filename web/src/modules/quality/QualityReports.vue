<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { qualityAPI } from '@/api/client'
import {
  BarChart3, RefreshCw, Calendar, Download,
  TrendingUp, TrendingDown, Target, ShieldCheck,
  AlertTriangle, Wrench, ClipboardList, Loader2,
  CheckCircle2, XCircle, Activity, Percent
} from 'lucide-vue-next'

const app = useAppStore()
const dk = (d: string, l: string) => app.darkMode ? d : l

// ─── state ────────────────────────────────────────────────────────────────────
const loading   = ref(false)
const kpis      = ref<any>({})
const period    = ref<any>({})
const inspByType  = ref<any[]>([])
const defectItems = ref<any[]>([])
const ncTrend     = ref<any[]>([])
const caEffectiveness = ref<any[]>([])

const fromDate = ref(new Date(Date.now() - 90 * 24 * 60 * 60 * 1000).toISOString().split('T')[0])
const toDate   = ref(new Date().toISOString().split('T')[0])

// ─── load ─────────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const res = await qualityAPI.getReports({ from: fromDate.value, to: toDate.value })
    const d = res.data
    kpis.value           = d.kpis || {}
    period.value         = d.period || {}
    inspByType.value     = d.inspections_by_type || []
    defectItems.value    = d.top_defect_items || []
    ncTrend.value        = d.nc_trend || []
    caEffectiveness.value = d.ca_effectiveness || []
  } catch {}
  loading.value = false
}

onMounted(load)

// ─── quick presets ────────────────────────────────────────────────────────────
function setPreset(days: number) {
  toDate.value   = new Date().toISOString().split('T')[0]
  fromDate.value = new Date(Date.now() - days * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
  load()
}

// ─── computed ─────────────────────────────────────────────────────────────────
const kpiCards = computed(() => [
  {
    label: 'Total Inspections',
    value: kpis.value.total_inspections ?? 0,
    icon: ClipboardList,
    color: 'indigo',
    sub: null,
  },
  {
    label: 'First Pass Rate',
    value: `${(kpis.value.first_pass_rate ?? 0).toFixed(1)}%`,
    icon: Target,
    color: (kpis.value.first_pass_rate ?? 0) >= 95 ? 'emerald' : (kpis.value.first_pass_rate ?? 0) >= 80 ? 'amber' : 'red',
    sub: null,
  },
  {
    label: 'Defect Rate',
    value: `${(kpis.value.defect_rate ?? 0).toFixed(2)}%`,
    icon: Percent,
    color: (kpis.value.defect_rate ?? 0) <= 2 ? 'emerald' : (kpis.value.defect_rate ?? 0) <= 5 ? 'amber' : 'red',
    sub: `${kpis.value.total_defects ?? 0} total defects`,
  },
  {
    label: 'Non-Conformities',
    value: kpis.value.total_nc ?? 0,
    icon: AlertTriangle,
    color: 'orange',
    sub: `${(kpis.value.nc_closure_rate ?? 0).toFixed(0)}% closed`,
  },
  {
    label: 'Corrective Actions',
    value: kpis.value.total_ca ?? 0,
    icon: Wrench,
    color: 'teal',
    sub: `${(kpis.value.ca_closure_rate ?? 0).toFixed(0)}% closed`,
  },
  {
    label: 'Overdue CAs',
    value: kpis.value.overdue_ca ?? 0,
    icon: Activity,
    color: kpis.value.overdue_ca > 0 ? 'red' : 'emerald',
    sub: null,
  },
])

const maxNCTrend = computed(() => {
  if (!ncTrend.value.length) return 1
  return Math.max(...ncTrend.value.map((t: any) => t.total || 0)) || 1
})

const maxInspections = computed(() => {
  if (!inspByType.value.length) return 1
  return Math.max(...inspByType.value.map((t: any) => t.total || 0)) || 1
})

const maxDefects = computed(() => {
  if (!defectItems.value.length) return 1
  return Math.max(...defectItems.value.map((d: any) => d.failed_checks || 0)) || 1
})

// ─── helpers ─────────────────────────────────────────────────────────────────
function caStatusColor(s: string) {
  const m: Record<string, string> = {
    open: 'bg-blue-500',
    in_progress: 'bg-indigo-500',
    pending_verification: 'bg-purple-500',
    verified: 'bg-teal-500',
    closed: 'bg-emerald-500',
    cancelled: 'bg-slate-500',
  }
  return m[s] || 'bg-slate-500'
}

function caStatusLabel(s: string) {
  const m: Record<string, string> = {
    open: 'Open', in_progress: 'In Progress',
    pending_verification: 'Pending Verification',
    verified: 'Verified', closed: 'Closed', cancelled: 'Cancelled',
  }
  return m[s] || s
}

function inspTypeColor(t: string) {
  const m: Record<string, string> = {
    incoming: 'bg-indigo-500', in_process: 'bg-purple-500',
    final: 'bg-cyan-500', audit: 'bg-orange-500', periodic: 'bg-teal-500',
  }
  return m[t] || 'bg-slate-500'
}

function barWidth(val: number, max: number) {
  if (!max) return '0%'
  return `${Math.round((val / max) * 100)}%`
}

function fmtMonth(m: string) {
  if (!m) return ''
  const [y, mo] = m.split('-')
  const months = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec']
  return `${months[parseInt(mo) - 1]} ${y}`
}
</script>

<template>
  <div :class="dk('bg-slate-900 text-slate-100','bg-slate-50 text-slate-900')" class="min-h-screen">
    <div class="max-w-screen-2xl mx-auto px-6 py-6">

      <!-- Header -->
      <div class="flex items-center justify-between mb-6 flex-wrap gap-3">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-violet-600 flex items-center justify-center">
            <BarChart3 class="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 class="text-2xl font-bold">Quality Reports</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">
              {{ period.from }} to {{ period.to }}
            </p>
          </div>
        </div>
        <div class="flex items-center gap-2 flex-wrap">
          <!-- Period presets -->
          <div class="flex items-center gap-1">
            <button v-for="preset in [
              { label: '30d', days: 30 },
              { label: '90d', days: 90 },
              { label: '180d', days: 180 },
              { label: '1y', days: 365 },
            ]" :key="preset.days"
              @click="setPreset(preset.days)"
              :class="dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-100')"
              class="px-3 py-1.5 text-xs border rounded-lg transition-colors">
              {{ preset.label }}
            </button>
          </div>
          <!-- Date range -->
          <div class="flex items-center gap-2">
            <input v-model="fromDate" type="date"
              :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
              class="px-3 py-1.5 text-xs border rounded-lg outline-none focus:ring-2 focus:ring-violet-500" />
            <span :class="dk('text-slate-400','text-slate-400')" class="text-xs">to</span>
            <input v-model="toDate" type="date"
              :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
              class="px-3 py-1.5 text-xs border rounded-lg outline-none focus:ring-2 focus:ring-violet-500" />
          </div>
          <button @click="load"
            class="flex items-center gap-2 px-4 py-1.5 text-sm rounded-lg bg-violet-600 hover:bg-violet-700 text-white transition-colors">
            <RefreshCw :class="{'animate-spin': loading}" class="w-4 h-4" />
            Apply
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-20">
        <Loader2 class="w-8 h-8 text-violet-500 animate-spin" />
      </div>

      <template v-else>

        <!-- ─── KPI Cards ────────────────────────────────────────────────── -->
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4 mb-6">
          <div v-for="card in kpiCards" :key="card.label"
            :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
            class="rounded-xl border p-4 flex flex-col gap-2">
            <div class="flex items-center justify-between">
              <p :class="dk('text-slate-400','text-slate-500')" class="text-xs font-medium">{{ card.label }}</p>
              <div :class="`bg-${card.color}-500/10 text-${card.color}-500`"
                class="w-7 h-7 rounded-lg flex items-center justify-center">
                <component :is="card.icon" class="w-4 h-4" />
              </div>
            </div>
            <p class="text-2xl font-bold" :class="`text-${card.color}-500`">{{ card.value }}</p>
            <p v-if="card.sub" :class="dk('text-slate-500','text-slate-400')" class="text-xs">{{ card.sub }}</p>
          </div>
        </div>

        <!-- ─── Main grid ─────────────────────────────────────────────────── -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-5 mb-5">

          <!-- Inspections by Type -->
          <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
            class="rounded-xl border p-5">
            <h3 class="font-semibold mb-4" :class="dk('text-slate-100','text-slate-800')">
              Inspections by Type
            </h3>
            <div v-if="inspByType.length === 0"
              :class="dk('text-slate-500','text-slate-400')" class="text-sm text-center py-8">
              No data for selected period
            </div>
            <div v-else class="space-y-3">
              <div v-for="row in inspByType" :key="row.inspection_type">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-sm font-medium capitalize" :class="dk('text-slate-200','text-slate-700')">
                    {{ row.inspection_type.replace('_', ' ') }}
                  </span>
                  <div class="flex items-center gap-3 text-xs">
                    <span class="text-emerald-500 font-semibold">{{ row.passed }} passed</span>
                    <span class="text-red-500 font-semibold">{{ row.failed }} failed</span>
                    <span :class="dk('text-slate-400','text-slate-500')">
                      {{ row.pass_rate.toFixed(1) }}%
                    </span>
                  </div>
                </div>
                <div :class="dk('bg-slate-700','bg-slate-100')" class="h-2 rounded-full overflow-hidden">
                  <div class="h-full flex">
                    <div class="h-full bg-emerald-500 transition-all"
                      :style="{ width: barWidth(row.passed, row.total) }" />
                    <div class="h-full bg-red-500 transition-all"
                      :style="{ width: barWidth(row.failed, row.total) }" />
                  </div>
                </div>
                <div class="flex justify-between text-xs mt-0.5"
                  :class="dk('text-slate-500','text-slate-400')">
                  <span>0</span>
                  <span>Total: {{ row.total }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- NC Trend by Month -->
          <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
            class="rounded-xl border p-5">
            <h3 class="font-semibold mb-4" :class="dk('text-slate-100','text-slate-800')">
              Non-Conformity Trend
            </h3>
            <div v-if="ncTrend.length === 0"
              :class="dk('text-slate-500','text-slate-400')" class="text-sm text-center py-8">
              No data for selected period
            </div>
            <div v-else>
              <!-- Bar chart -->
              <div class="flex items-end gap-2 h-36 mb-2">
                <div v-for="row in ncTrend" :key="row.month"
                  class="flex-1 flex flex-col items-center gap-1">
                  <div class="w-full flex flex-col justify-end" style="height: 100%">
                    <!-- Total bar -->
                    <div class="w-full rounded-t transition-all relative"
                      :style="{ height: barWidth(row.total, maxNCTrend) }"
                      :class="dk('bg-orange-600/60','bg-orange-400/60')">
                      <!-- Critical overlay -->
                      <div class="absolute bottom-0 left-0 right-0 rounded-t"
                        :style="{ height: barWidth(row.critical, row.total || 1) }"
                        :class="dk('bg-red-500','bg-red-500')" />
                    </div>
                  </div>
                  <span class="text-xs" :class="dk('text-slate-500','text-slate-400')" style="font-size:10px">
                    {{ fmtMonth(row.month).slice(0, 3) }}
                  </span>
                </div>
              </div>
              <!-- Legend -->
              <div class="flex items-center gap-4 text-xs" :class="dk('text-slate-400','text-slate-500')">
                <div class="flex items-center gap-1.5">
                  <div class="w-3 h-3 rounded" :class="dk('bg-orange-600/60','bg-orange-400/60')" />
                  <span>Total</span>
                </div>
                <div class="flex items-center gap-1.5">
                  <div class="w-3 h-3 rounded bg-red-500" />
                  <span>Critical</span>
                </div>
              </div>
              <!-- Table -->
              <table class="w-full mt-4 text-xs">
                <thead>
                  <tr :class="dk('text-slate-400','text-slate-500')">
                    <th class="text-left py-1 font-medium">Month</th>
                    <th class="text-right py-1 font-medium">Total</th>
                    <th class="text-right py-1 font-medium">Closed</th>
                    <th class="text-right py-1 font-medium">Critical</th>
                    <th class="text-right py-1 font-medium">Rate</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in ncTrend" :key="row.month"
                    :class="dk('border-t border-slate-700/50 text-slate-300','border-t border-slate-100 text-slate-600')">
                    <td class="py-1.5">{{ fmtMonth(row.month) }}</td>
                    <td class="text-right font-semibold">{{ row.total }}</td>
                    <td class="text-right text-emerald-500">{{ row.closed }}</td>
                    <td class="text-right text-red-500">{{ row.critical }}</td>
                    <td class="text-right">
                      {{ row.total > 0 ? ((row.closed / row.total) * 100).toFixed(0) : 0 }}%
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- ─── Second row ─────────────────────────────────────────────────── -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">

          <!-- Top Defect Items -->
          <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
            class="rounded-xl border p-5">
            <h3 class="font-semibold mb-4" :class="dk('text-slate-100','text-slate-800')">
              Top Defect Items
            </h3>
            <div v-if="defectItems.length === 0"
              :class="dk('text-slate-500','text-slate-400')" class="text-sm text-center py-8">
              No defects recorded for selected period
            </div>
            <div v-else class="space-y-3">
              <div v-for="(item, idx) in defectItems" :key="item.item_name">
                <div class="flex items-center justify-between mb-1">
                  <div class="flex items-center gap-2">
                    <span class="text-xs font-bold w-5 text-right"
                      :class="idx === 0 ? 'text-red-500' : idx === 1 ? 'text-orange-500' : idx === 2 ? 'text-amber-500' : dk('text-slate-500','text-slate-400')">
                      {{ idx + 1 }}
                    </span>
                    <span class="text-sm font-medium truncate max-w-[160px]"
                      :class="dk('text-slate-200','text-slate-700')">
                      {{ item.item_name }}
                    </span>
                  </div>
                  <div class="flex items-center gap-3 text-xs">
                    <span class="text-red-500 font-semibold">{{ item.failed_checks }} fails</span>
                    <span :class="dk('text-slate-400','text-slate-500')">{{ item.defect_rate }}%</span>
                  </div>
                </div>
                <div :class="dk('bg-slate-700','bg-slate-100')" class="h-1.5 rounded-full overflow-hidden">
                  <div class="h-full bg-red-500 transition-all"
                    :style="{ width: barWidth(item.failed_checks, maxDefects) }" />
                </div>
              </div>
            </div>
          </div>

          <!-- CA Effectiveness Summary -->
          <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
            class="rounded-xl border p-5">
            <h3 class="font-semibold mb-4" :class="dk('text-slate-100','text-slate-800')">
              Corrective Action Effectiveness
            </h3>
            <div v-if="caEffectiveness.length === 0"
              :class="dk('text-slate-500','text-slate-400')" class="text-sm text-center py-8">
              No corrective actions for selected period
            </div>
            <div v-else>
              <table class="w-full text-sm">
                <thead>
                  <tr :class="dk('text-slate-400','text-slate-500')">
                    <th class="text-left py-2 font-medium">Status</th>
                    <th class="text-right py-2 font-medium">Count</th>
                    <th class="text-right py-2 font-medium">Avg Days</th>
                    <th class="text-right py-2 font-medium">Avg Rating</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in caEffectiveness" :key="row.status"
                    :class="dk('border-t border-slate-700/60','border-t border-slate-100')">
                    <td class="py-3">
                      <div class="flex items-center gap-2">
                        <div :class="caStatusColor(row.status)" class="w-2 h-2 rounded-full" />
                        <span :class="dk('text-slate-200','text-slate-700')">
                          {{ caStatusLabel(row.status) }}
                        </span>
                      </div>
                    </td>
                    <td class="text-right py-3 font-semibold" :class="dk('text-slate-200','text-slate-700')">
                      {{ row.count }}
                    </td>
                    <td class="text-right py-3" :class="dk('text-slate-300','text-slate-600')">
                      {{ row.avg_closure_days > 0 ? row.avg_closure_days + ' days' : '—' }}
                    </td>
                    <td class="text-right py-3">
                      <div v-if="row.avg_effectiveness > 0" class="flex items-center justify-end gap-1">
                        <div class="flex">
                          <svg v-for="i in 5" :key="i" viewBox="0 0 20 20"
                            :class="i <= Math.round(row.avg_effectiveness)
                              ? 'text-amber-400 fill-amber-400'
                              : dk('text-slate-600 fill-slate-600','text-slate-300 fill-slate-300')"
                            class="w-3.5 h-3.5">
                            <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                          </svg>
                        </div>
                        <span class="text-xs ml-1" :class="dk('text-slate-400','text-slate-500')">
                          {{ row.avg_effectiveness.toFixed(1) }}
                        </span>
                      </div>
                      <span v-else :class="dk('text-slate-600','text-slate-400')" class="text-xs">—</span>
                    </td>
                  </tr>
                </tbody>
              </table>

              <!-- Overall stats -->
              <div class="mt-4 pt-4" :class="dk('border-t border-slate-700','border-t border-slate-200')">
                <div class="grid grid-cols-3 gap-4">
                  <div class="text-center">
                    <p :class="dk('text-slate-400','text-slate-500')" class="text-xs mb-1">Total CAs</p>
                    <p class="text-xl font-bold text-teal-500">{{ kpis.total_ca ?? 0 }}</p>
                  </div>
                  <div class="text-center">
                    <p :class="dk('text-slate-400','text-slate-500')" class="text-xs mb-1">Closure Rate</p>
                    <p class="text-xl font-bold"
                      :class="(kpis.ca_closure_rate ?? 0) >= 80 ? 'text-emerald-500' : 'text-amber-500'">
                      {{ (kpis.ca_closure_rate ?? 0).toFixed(1) }}%
                    </p>
                  </div>
                  <div class="text-center">
                    <p :class="dk('text-slate-400','text-slate-500')" class="text-xs mb-1">Overdue</p>
                    <p class="text-xl font-bold"
                      :class="(kpis.overdue_ca ?? 0) > 0 ? 'text-red-500' : 'text-emerald-500'">
                      {{ kpis.overdue_ca ?? 0 }}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- ─── Summary Table ─────────────────────────────────────────────── -->
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="rounded-xl border p-5 mt-5">
          <h3 class="font-semibold mb-4" :class="dk('text-slate-100','text-slate-800')">
            Period Summary
          </h3>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-6">
            <div>
              <h4 class="text-xs font-semibold uppercase tracking-wider mb-3"
                :class="dk('text-slate-400','text-slate-500')">Inspection KPIs</h4>
              <div class="space-y-2">
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">Total</span>
                  <span class="font-semibold" :class="dk('text-slate-200','text-slate-700')">
                    {{ kpis.total_inspections ?? 0 }}
                  </span>
                </div>
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">First Pass Rate</span>
                  <span class="font-semibold"
                    :class="(kpis.first_pass_rate ?? 0) >= 95 ? 'text-emerald-500' : (kpis.first_pass_rate ?? 0) >= 80 ? 'text-amber-500' : 'text-red-500'">
                    {{ (kpis.first_pass_rate ?? 0).toFixed(1) }}%
                  </span>
                </div>
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">Defect Rate</span>
                  <span class="font-semibold"
                    :class="(kpis.defect_rate ?? 0) <= 2 ? 'text-emerald-500' : (kpis.defect_rate ?? 0) <= 5 ? 'text-amber-500' : 'text-red-500'">
                    {{ (kpis.defect_rate ?? 0).toFixed(2) }}%
                  </span>
                </div>
              </div>
            </div>
            <div>
              <h4 class="text-xs font-semibold uppercase tracking-wider mb-3"
                :class="dk('text-slate-400','text-slate-500')">NC KPIs</h4>
              <div class="space-y-2">
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">Total NCs</span>
                  <span class="font-semibold" :class="dk('text-slate-200','text-slate-700')">
                    {{ kpis.total_nc ?? 0 }}
                  </span>
                </div>
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">NC Closure Rate</span>
                  <span class="font-semibold"
                    :class="(kpis.nc_closure_rate ?? 0) >= 80 ? 'text-emerald-500' : 'text-amber-500'">
                    {{ (kpis.nc_closure_rate ?? 0).toFixed(1) }}%
                  </span>
                </div>
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">Total Defects</span>
                  <span class="font-semibold text-red-500">{{ kpis.total_defects ?? 0 }}</span>
                </div>
              </div>
            </div>
            <div>
              <h4 class="text-xs font-semibold uppercase tracking-wider mb-3"
                :class="dk('text-slate-400','text-slate-500')">CA KPIs</h4>
              <div class="space-y-2">
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">Total CAs</span>
                  <span class="font-semibold" :class="dk('text-slate-200','text-slate-700')">
                    {{ kpis.total_ca ?? 0 }}
                  </span>
                </div>
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">CA Closure Rate</span>
                  <span class="font-semibold"
                    :class="(kpis.ca_closure_rate ?? 0) >= 80 ? 'text-emerald-500' : 'text-amber-500'">
                    {{ (kpis.ca_closure_rate ?? 0).toFixed(1) }}%
                  </span>
                </div>
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">Overdue CAs</span>
                  <span class="font-semibold" :class="(kpis.overdue_ca ?? 0) > 0 ? 'text-red-500' : 'text-emerald-500'">
                    {{ kpis.overdue_ca ?? 0 }}
                  </span>
                </div>
              </div>
            </div>
            <div>
              <h4 class="text-xs font-semibold uppercase tracking-wider mb-3"
                :class="dk('text-slate-400','text-slate-500')">Period</h4>
              <div class="space-y-2">
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">From</span>
                  <span class="font-semibold" :class="dk('text-slate-200','text-slate-700')">
                    {{ period.from || '—' }}
                  </span>
                </div>
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">To</span>
                  <span class="font-semibold" :class="dk('text-slate-200','text-slate-700')">
                    {{ period.to || '—' }}
                  </span>
                </div>
                <div class="flex justify-between text-sm">
                  <span :class="dk('text-slate-400','text-slate-500')">Insp. Types</span>
                  <span class="font-semibold" :class="dk('text-slate-200','text-slate-700')">
                    {{ inspByType.length }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

      </template>
    </div>
  </div>
</template>
