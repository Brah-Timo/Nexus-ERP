<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { assetsAPI } from '@/api/client'
import {
  BarChart3, RefreshCw, Download, FileText,
  TrendingDown, Wrench, ArrowLeftRight, Package
} from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(false)
const reportType = ref('summary')
const dateFrom = ref(new Date(new Date().getFullYear(), 0, 1).toISOString().slice(0, 10))
const dateTo = ref(new Date().toISOString().slice(0, 10))
const year = ref(new Date().getFullYear().toString())
const data = ref<any>({})

const reportTypes = [
  { value: 'summary', label: 'Executive Summary', icon: BarChart3 },
  { value: 'asset_register', label: 'Asset Register', icon: Package },
  { value: 'depreciation_summary', label: 'Depreciation Summary', icon: TrendingDown },
  { value: 'maintenance_costs', label: 'Maintenance Costs', icon: Wrench },
  { value: 'transfer_history', label: 'Transfer History', icon: ArrowLeftRight },
]

async function load() {
  loading.value = true
  try {
    const params: Record<string, string> = {
      type: reportType.value,
      date_from: dateFrom.value,
      date_to: dateTo.value,
    }
    if (reportType.value === 'depreciation_summary') params.year = year.value
    const res = await assetsAPI.getReports(params)
    data.value = res.data
  } finally {
    loading.value = false
  }
}

onMounted(load)

function fmtCurrency(v: number) {
  return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0)
}
function fmtDate(d: any) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString()
}

const cardCls = computed(() =>
  app.darkMode ? 'bg-slate-800/60 border-slate-700' : 'bg-white border-slate-200 shadow-sm'
)
const inputCls = computed(() =>
  app.darkMode
    ? 'bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400 focus:border-indigo-500'
    : 'bg-white border-slate-300 text-slate-900 placeholder-slate-400 focus:border-indigo-500'
)
const thCls = computed(() =>
  app.darkMode ? 'text-slate-400 border-slate-700 bg-slate-700/40' : 'text-slate-500 border-slate-200 bg-slate-50'
)
const tdCls = computed(() =>
  app.darkMode ? 'text-slate-300 border-slate-700' : 'text-slate-700 border-slate-200'
)

const maxTrendDep = computed(() =>
  Math.max(1, ...(data.value.depreciation_trend?.map((r: any) => r.depreciation) ?? [0]))
)
const maxMainCost = computed(() =>
  Math.max(1, ...(data.value.maintenance_costs?.map((r: any) => r.cost) ?? [0]))
)
</script>

<template>
  <div class="p-6 space-y-5">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Assets Reports</h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Comprehensive reports and analytics for fixed assets</p>
      </div>
      <button @click="load" class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors"
        :class="app.darkMode ? 'border-slate-600 text-slate-300 hover:bg-slate-700' : 'border-slate-300 text-slate-600 hover:bg-slate-50'">
        <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
      </button>
    </div>

    <!-- Report Type Tabs -->
    <div class="flex gap-2 flex-wrap">
      <button
        v-for="rt in reportTypes"
        :key="rt.value"
        @click="reportType = rt.value; load()"
        class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors border"
        :class="reportType === rt.value
          ? 'bg-indigo-600 border-indigo-600 text-white'
          : app.darkMode
            ? 'border-slate-600 text-slate-300 hover:bg-slate-700'
            : 'border-slate-200 text-slate-600 hover:bg-slate-50'">
        <component :is="rt.icon" class="w-4 h-4" />
        {{ rt.label }}
      </button>
    </div>

    <!-- Date Filters -->
    <div class="flex flex-wrap gap-3 items-end">
      <div v-if="reportType === 'depreciation_summary'">
        <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Year</label>
        <select v-model="year" @change="load" class="px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls">
          <option v-for="y in [new Date().getFullYear()-2, new Date().getFullYear()-1, new Date().getFullYear()]" :key="y" :value="y.toString()">{{ y }}</option>
        </select>
      </div>
      <template v-else>
        <div>
          <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">From</label>
          <input v-model="dateFrom" type="date" class="px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
        </div>
        <div>
          <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">To</label>
          <input v-model="dateTo" type="date" class="px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
        </div>
        <button @click="load" class="px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white">Apply</button>
      </template>
    </div>

    <!-- ── Summary Report ── -->
    <template v-if="reportType === 'summary'">
      <!-- KPI Row -->
      <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
        <div v-for="kpi in [
          { label: 'Total Assets', value: data.total_assets ?? 0, color: 'text-indigo-400' },
          { label: 'Active Assets', value: data.active_assets ?? 0, color: 'text-emerald-400' },
          { label: 'Disposed Assets', value: data.disposed_assets ?? 0, color: 'text-red-400' },
          { label: 'Total Cost', value: fmtCurrency(data.total_cost ?? 0), color: 'text-amber-400' },
          { label: 'Net Book Value', value: fmtCurrency(data.total_nbv ?? 0), color: 'text-sky-400' },
          { label: 'Accum. Depreciation', value: fmtCurrency(data.total_accum_dep ?? 0), color: 'text-purple-400' },
        ]" :key="kpi.label" class="rounded-xl border p-4" :class="cardCls">
          <div class="text-xs mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ kpi.label }}</div>
          <div class="text-xl font-bold" :class="kpi.color">{{ kpi.value }}</div>
        </div>
      </div>

      <!-- Depreciation Trend -->
      <div class="rounded-xl border p-5" :class="cardCls">
        <h2 class="text-sm font-semibold mb-4" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Monthly Depreciation Trend</h2>
        <div v-if="!data.depreciation_trend?.length" class="text-center py-6 text-slate-400 text-sm">No data</div>
        <div v-else class="space-y-2">
          <div v-for="t in data.depreciation_trend" :key="t.label" class="flex items-center gap-3">
            <span class="text-xs w-14 flex-shrink-0 font-mono" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ t.label }}</span>
            <div class="flex-1 h-5 rounded" :class="app.darkMode ? 'bg-slate-700' : 'bg-slate-100'">
              <div class="h-full rounded bg-indigo-500 flex items-center justify-end pr-2 transition-all"
                :style="{ width: Math.max(2, (t.depreciation / maxTrendDep) * 100) + '%' }">
                <span class="text-xs text-white font-medium">{{ fmtCurrency(t.depreciation) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Maintenance Cost Trend -->
      <div class="rounded-xl border p-5" :class="cardCls">
        <h2 class="text-sm font-semibold mb-4" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Maintenance Costs by Period</h2>
        <div v-if="!data.maintenance_costs?.length" class="text-center py-6 text-slate-400 text-sm">No data</div>
        <div v-else class="space-y-2">
          <div v-for="m in data.maintenance_costs" :key="m.period" class="flex items-center gap-3">
            <span class="text-xs w-14 flex-shrink-0 font-mono" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ m.period }}</span>
            <div class="flex-1 h-5 rounded" :class="app.darkMode ? 'bg-slate-700' : 'bg-slate-100'">
              <div class="h-full rounded bg-emerald-500 flex items-center justify-end pr-2 transition-all"
                :style="{ width: Math.max(2, (m.cost / maxMainCost) * 100) + '%' }">
                <span class="text-xs text-white font-medium">{{ fmtCurrency(m.cost) }}</span>
              </div>
            </div>
            <span class="text-xs w-6 text-right" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ m.count }}</span>
          </div>
        </div>
      </div>
    </template>

    <!-- ── Asset Register ── -->
    <template v-else-if="reportType === 'asset_register'">
      <div class="rounded-xl border overflow-hidden" :class="cardCls">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr :class="thCls">
                <th class="text-left px-4 py-3 font-medium">Asset #</th>
                <th class="text-left px-4 py-3 font-medium">Name</th>
                <th class="text-left px-4 py-3 font-medium">Category</th>
                <th class="text-left px-4 py-3 font-medium">Location</th>
                <th class="text-left px-4 py-3 font-medium">Status</th>
                <th class="text-right px-4 py-3 font-medium">Cost</th>
                <th class="text-right px-4 py-3 font-medium">Accum. Dep.</th>
                <th class="text-right px-4 py-3 font-medium">Net Book Value</th>
                <th class="text-left px-4 py-3 font-medium">Method</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="loading"><td colspan="9" class="text-center py-8 text-slate-400">Loading...</td></tr>
              <tr v-else-if="!data.data?.length"><td colspan="9" class="text-center py-8 text-slate-400">No data</td></tr>
              <tr v-for="row in data.data" :key="row.asset_number"
                class="border-t" :class="tdCls">
                <td class="px-4 py-2.5 font-mono text-indigo-400 text-xs">{{ row.asset_number }}</td>
                <td class="px-4 py-2.5 font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">{{ row.name }}</td>
                <td class="px-4 py-2.5 text-xs">{{ row.category }}</td>
                <td class="px-4 py-2.5 text-xs">{{ row.location }}</td>
                <td class="px-4 py-2.5"><span class="text-xs capitalize">{{ row.status?.replace(/_/g,' ') }}</span></td>
                <td class="px-4 py-2.5 text-right font-mono text-xs">{{ fmtCurrency(row.purchase_cost) }}</td>
                <td class="px-4 py-2.5 text-right font-mono text-xs text-amber-400">{{ fmtCurrency(row.accumulated_depreciation) }}</td>
                <td class="px-4 py-2.5 text-right font-mono text-xs text-emerald-400 font-medium">{{ fmtCurrency(row.net_book_value) }}</td>
                <td class="px-4 py-2.5 text-xs capitalize">{{ row.depreciation_method?.replace(/_/g,' ') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <!-- ── Depreciation Summary ── -->
    <template v-else-if="reportType === 'depreciation_summary'">
      <div class="rounded-xl border overflow-hidden" :class="cardCls">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr :class="thCls">
                <th class="text-left px-4 py-3 font-medium">Category</th>
                <th class="text-right px-4 py-3 font-medium">Assets</th>
                <th class="text-right px-4 py-3 font-medium">Purchase Cost</th>
                <th class="text-right px-4 py-3 font-medium">Accum. Dep.</th>
                <th class="text-right px-4 py-3 font-medium">Net Book Value</th>
                <th class="text-right px-4 py-3 font-medium">Year Depreciation</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="loading"><td colspan="6" class="text-center py-8 text-slate-400">Loading...</td></tr>
              <tr v-else-if="!data.data?.length"><td colspan="6" class="text-center py-8 text-slate-400">No data</td></tr>
              <tr v-for="row in data.data" :key="row.category"
                class="border-t" :class="tdCls">
                <td class="px-4 py-2.5 font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">{{ row.category }}</td>
                <td class="px-4 py-2.5 text-right">{{ row.asset_count }}</td>
                <td class="px-4 py-2.5 text-right font-mono text-xs">{{ fmtCurrency(row.purchase_cost) }}</td>
                <td class="px-4 py-2.5 text-right font-mono text-xs text-amber-400">{{ fmtCurrency(row.accumulated_depreciation) }}</td>
                <td class="px-4 py-2.5 text-right font-mono text-xs text-emerald-400">{{ fmtCurrency(row.net_book_value) }}</td>
                <td class="px-4 py-2.5 text-right font-mono text-xs text-red-400">{{ fmtCurrency(row.year_depreciation) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <!-- ── Maintenance Costs ── -->
    <template v-else-if="reportType === 'maintenance_costs'">
      <div class="rounded-xl border overflow-hidden" :class="cardCls">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr :class="thCls">
                <th class="text-left px-4 py-3 font-medium">Asset #</th>
                <th class="text-left px-4 py-3 font-medium">Name</th>
                <th class="text-left px-4 py-3 font-medium">Category</th>
                <th class="text-right px-4 py-3 font-medium">Records</th>
                <th class="text-right px-4 py-3 font-medium">Completed</th>
                <th class="text-right px-4 py-3 font-medium">Pending</th>
                <th class="text-right px-4 py-3 font-medium">Total Cost</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="loading"><td colspan="7" class="text-center py-8 text-slate-400">Loading...</td></tr>
              <tr v-else-if="!data.data?.length"><td colspan="7" class="text-center py-8 text-slate-400">No data</td></tr>
              <tr v-for="row in data.data" :key="row.asset_number"
                class="border-t" :class="tdCls">
                <td class="px-4 py-2.5 font-mono text-indigo-400 text-xs">{{ row.asset_number }}</td>
                <td class="px-4 py-2.5 font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">{{ row.name }}</td>
                <td class="px-4 py-2.5 text-xs">{{ row.category }}</td>
                <td class="px-4 py-2.5 text-right">{{ row.total_records }}</td>
                <td class="px-4 py-2.5 text-right text-emerald-400">{{ row.completed }}</td>
                <td class="px-4 py-2.5 text-right text-amber-400">{{ row.pending }}</td>
                <td class="px-4 py-2.5 text-right font-mono text-sm font-medium text-red-400">{{ fmtCurrency(row.total_cost) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <!-- ── Transfer History ── -->
    <template v-else-if="reportType === 'transfer_history'">
      <div class="rounded-xl border overflow-hidden" :class="cardCls">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr :class="thCls">
                <th class="text-left px-4 py-3 font-medium">Transfer #</th>
                <th class="text-left px-4 py-3 font-medium">Asset</th>
                <th class="text-left px-4 py-3 font-medium">From</th>
                <th class="text-left px-4 py-3 font-medium">To</th>
                <th class="text-left px-4 py-3 font-medium">Status</th>
                <th class="text-left px-4 py-3 font-medium">Date</th>
                <th class="text-left px-4 py-3 font-medium">Reason</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="loading"><td colspan="7" class="text-center py-8 text-slate-400">Loading...</td></tr>
              <tr v-else-if="!data.data?.length"><td colspan="7" class="text-center py-8 text-slate-400">No data</td></tr>
              <tr v-for="row in data.data" :key="row.transfer_number"
                class="border-t" :class="tdCls">
                <td class="px-4 py-2.5 font-mono text-indigo-400 text-xs">{{ row.transfer_number }}</td>
                <td class="px-4 py-2.5 text-xs">
                  <div class="font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">{{ row.asset_name }}</div>
                  <div class="text-slate-400">{{ row.asset_number }}</div>
                </td>
                <td class="px-4 py-2.5 text-xs">{{ row.from_location || '—' }}</td>
                <td class="px-4 py-2.5 text-xs text-sky-400">{{ row.to_location }}</td>
                <td class="px-4 py-2.5"><span class="text-xs capitalize">{{ row.status?.replace(/_/g,' ') }}</span></td>
                <td class="px-4 py-2.5 text-xs">{{ fmtDate(row.transfer_date) }}</td>
                <td class="px-4 py-2.5 text-xs max-w-[150px] truncate">{{ row.reason || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>
