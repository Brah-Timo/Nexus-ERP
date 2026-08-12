<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { assetsAPI } from '@/api/client'
import {
  Package, Archive, TrendingDown, DollarSign, Wrench,
  ArrowLeftRight, AlertCircle, RefreshCw, BarChart3,
  CheckCircle, Clock, XCircle, MapPin, FolderOpen
} from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(false)
const dashboard = ref<Record<string, any>>({})

async function load() {
  loading.value = true
  try {
    const res = await assetsAPI.getDashboard()
    dashboard.value = res.data
  } finally {
    loading.value = false
  }
}

onMounted(load)

const kpis = computed(() => [
  {
    label: 'Total Assets',
    value: dashboard.value.total_assets ?? 0,
    unit: '',
    icon: Package,
    color: 'text-indigo-400',
    bg: 'bg-indigo-500/10',
  },
  {
    label: 'Total Cost',
    value: fmtCurrency(dashboard.value.total_cost ?? 0),
    unit: '',
    icon: DollarSign,
    color: 'text-emerald-400',
    bg: 'bg-emerald-500/10',
  },
  {
    label: 'Net Book Value',
    value: fmtCurrency(dashboard.value.total_net_book_value ?? 0),
    unit: '',
    icon: TrendingDown,
    color: 'text-sky-400',
    bg: 'bg-sky-500/10',
  },
  {
    label: 'Accum. Depreciation',
    value: fmtCurrency(dashboard.value.total_accum_dep ?? 0),
    unit: '',
    icon: BarChart3,
    color: 'text-amber-400',
    bg: 'bg-amber-500/10',
  },
  {
    label: 'Due Maintenance',
    value: dashboard.value.due_maintenance ?? 0,
    unit: '',
    icon: Wrench,
    color: 'text-red-400',
    bg: 'bg-red-500/10',
  },
  {
    label: 'Pending Transfers',
    value: dashboard.value.pending_transfers ?? 0,
    unit: '',
    icon: ArrowLeftRight,
    color: 'text-purple-400',
    bg: 'bg-purple-500/10',
  },
])

function fmtCurrency(v: number) {
  return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)
}

function statusColor(s: string) {
  const m: Record<string, string> = {
    active: 'bg-emerald-500/20 text-emerald-300',
    in_use: 'bg-sky-500/20 text-sky-300',
    in_storage: 'bg-slate-500/20 text-slate-300',
    under_maintenance: 'bg-amber-500/20 text-amber-300',
    disposed: 'bg-red-500/20 text-red-300',
    sold: 'bg-purple-500/20 text-purple-300',
    written_off: 'bg-gray-500/20 text-gray-400',
  }
  return m[s] ?? 'bg-slate-500/20 text-slate-300'
}

function statusIcon(s: string) {
  const m: Record<string, any> = {
    active: CheckCircle,
    in_use: CheckCircle,
    in_storage: Archive,
    under_maintenance: Wrench,
    disposed: XCircle,
    sold: XCircle,
    written_off: AlertCircle,
  }
  return m[s] ?? Package
}

const byStatus = computed(() => dashboard.value.by_status ?? [])
const byCategory = computed(() => dashboard.value.by_category ?? [])
const byLocation = computed(() => dashboard.value.by_location ?? [])
const recentAssets = computed(() => dashboard.value.recent_assets ?? [])

const maxCatValue = computed(() =>
  Math.max(1, ...byCategory.value.map((r: any) => r.value ?? 0))
)
const maxLocCount = computed(() =>
  Math.max(1, ...byLocation.value.map((r: any) => r.count ?? 0))
)
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Assets Dashboard</h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
          Fixed assets overview and performance metrics
        </p>
      </div>
      <button
        @click="load"
        class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors bg-indigo-600 hover:bg-indigo-700 text-white"
      >
        <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        Refresh
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-4">
      <div
        v-for="kpi in kpis"
        :key="kpi.label"
        class="rounded-xl p-4 border"
        :class="app.darkMode ? 'bg-slate-800/50 border-slate-700' : 'bg-white border-slate-200 shadow-sm'"
      >
        <div class="flex items-center gap-3 mb-3">
          <div :class="['w-9 h-9 rounded-lg flex items-center justify-center', kpi.bg]">
            <component :is="kpi.icon" :class="['w-5 h-5', kpi.color]" />
          </div>
        </div>
        <div class="text-2xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
          {{ kpi.value }}
        </div>
        <div class="text-xs mt-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ kpi.label }}</div>
      </div>
    </div>

    <!-- Row: Status breakdown + Category chart -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Assets by Status -->
      <div class="rounded-xl border p-5"
        :class="app.darkMode ? 'bg-slate-800/50 border-slate-700' : 'bg-white border-slate-200 shadow-sm'">
        <h2 class="text-base font-semibold mb-4 flex items-center gap-2"
          :class="app.darkMode ? 'text-white' : 'text-slate-900'">
          <Package class="w-4 h-4 text-indigo-400" /> Assets by Status
        </h2>
        <div v-if="byStatus.length === 0" class="text-center py-8 text-slate-400 text-sm">No data</div>
        <div v-else class="space-y-3">
          <div v-for="row in byStatus" :key="row.status" class="flex items-center gap-3">
            <component :is="statusIcon(row.status)" class="w-4 h-4 flex-shrink-0" :class="statusColor(row.status).split(' ')[1]" />
            <div class="flex-1 min-w-0">
              <div class="flex justify-between items-center mb-1">
                <span class="text-sm font-medium capitalize" :class="app.darkMode ? 'text-slate-200' : 'text-slate-700'">
                  {{ row.status.replace(/_/g,' ') }}
                </span>
                <span class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
                  {{ row.count }} assets
                </span>
              </div>
              <div class="h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-700' : 'bg-slate-100'">
                <div
                  class="h-full rounded-full bg-indigo-500 transition-all"
                  :style="{ width: Math.min(100, (row.count / (dashboard.total_assets || 1)) * 100) + '%' }"
                />
              </div>
            </div>
            <span class="text-xs font-medium w-20 text-right" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
              {{ fmtCurrency(row.value) }}
            </span>
          </div>
        </div>
      </div>

      <!-- Assets by Category -->
      <div class="rounded-xl border p-5"
        :class="app.darkMode ? 'bg-slate-800/50 border-slate-700' : 'bg-white border-slate-200 shadow-sm'">
        <h2 class="text-base font-semibold mb-4 flex items-center gap-2"
          :class="app.darkMode ? 'text-white' : 'text-slate-900'">
          <FolderOpen class="w-4 h-4 text-amber-400" /> Assets by Category
        </h2>
        <div v-if="byCategory.length === 0" class="text-center py-8 text-slate-400 text-sm">No data</div>
        <div v-else class="space-y-3">
          <div v-for="row in byCategory.slice(0,8)" :key="row.category" class="flex items-center gap-3">
            <div class="flex-1 min-w-0">
              <div class="flex justify-between items-center mb-1">
                <span class="text-sm font-medium truncate" :class="app.darkMode ? 'text-slate-200' : 'text-slate-700'">
                  {{ row.category }}
                </span>
                <span class="text-xs ml-2 flex-shrink-0" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
                  {{ row.count }}
                </span>
              </div>
              <div class="h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-700' : 'bg-slate-100'">
                <div
                  class="h-full rounded-full bg-amber-500 transition-all"
                  :style="{ width: Math.min(100, (row.value / maxCatValue) * 100) + '%' }"
                />
              </div>
            </div>
            <span class="text-xs font-medium w-24 text-right" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
              {{ fmtCurrency(row.value) }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Row: Locations + Recent assets -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Assets by Location -->
      <div class="rounded-xl border p-5"
        :class="app.darkMode ? 'bg-slate-800/50 border-slate-700' : 'bg-white border-slate-200 shadow-sm'">
        <h2 class="text-base font-semibold mb-4 flex items-center gap-2"
          :class="app.darkMode ? 'text-white' : 'text-slate-900'">
          <MapPin class="w-4 h-4 text-sky-400" /> Assets by Location
        </h2>
        <div v-if="byLocation.length === 0" class="text-center py-8 text-slate-400 text-sm">No data</div>
        <div v-else class="space-y-3">
          <div v-for="row in byLocation.slice(0,8)" :key="row.location" class="flex items-center gap-3">
            <div class="flex-1 min-w-0">
              <div class="flex justify-between items-center mb-1">
                <span class="text-sm font-medium truncate" :class="app.darkMode ? 'text-slate-200' : 'text-slate-700'">
                  {{ row.location }}
                </span>
                <span class="text-xs ml-2 flex-shrink-0" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
                  {{ row.count }}
                </span>
              </div>
              <div class="h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-700' : 'bg-slate-100'">
                <div
                  class="h-full rounded-full bg-sky-500 transition-all"
                  :style="{ width: Math.min(100, (row.count / maxLocCount) * 100) + '%' }"
                />
              </div>
            </div>
            <span class="text-xs font-medium w-24 text-right" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
              {{ fmtCurrency(row.value) }}
            </span>
          </div>
        </div>
      </div>

      <!-- Recent Assets -->
      <div class="rounded-xl border p-5"
        :class="app.darkMode ? 'bg-slate-800/50 border-slate-700' : 'bg-white border-slate-200 shadow-sm'">
        <h2 class="text-base font-semibold mb-4 flex items-center gap-2"
          :class="app.darkMode ? 'text-white' : 'text-slate-900'">
          <Clock class="w-4 h-4 text-purple-400" /> Recently Added Assets
        </h2>
        <div v-if="recentAssets.length === 0" class="text-center py-8 text-slate-400 text-sm">No assets found</div>
        <div v-else class="space-y-2">
          <div
            v-for="a in recentAssets"
            :key="a.id"
            class="flex items-center gap-3 p-2 rounded-lg transition-colors"
            :class="app.darkMode ? 'hover:bg-slate-700/50' : 'hover:bg-slate-50'"
          >
            <div class="w-8 h-8 rounded-lg bg-indigo-500/20 flex items-center justify-center flex-shrink-0">
              <Package class="w-4 h-4 text-indigo-400" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-sm font-medium truncate" :class="app.darkMode ? 'text-slate-200' : 'text-slate-800'">
                {{ a.name }}
              </div>
              <div class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
                {{ a.asset_number }} · {{ a.category || 'Uncategorized' }}
              </div>
            </div>
            <div class="text-right flex-shrink-0">
              <div class="text-sm font-medium" :class="app.darkMode ? 'text-slate-200' : 'text-slate-800'">
                {{ fmtCurrency(a.net_book_value) }}
              </div>
              <span class="text-xs px-1.5 py-0.5 rounded-full" :class="statusColor(a.status)">
                {{ a.status.replace(/_/g,' ') }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Depreciation Summary row -->
    <div class="rounded-xl border p-5"
      :class="app.darkMode ? 'bg-slate-800/50 border-slate-700' : 'bg-white border-slate-200 shadow-sm'">
      <h2 class="text-base font-semibold mb-4 flex items-center gap-2"
        :class="app.darkMode ? 'text-white' : 'text-slate-900'">
        <TrendingDown class="w-4 h-4 text-red-400" /> Depreciation Overview
      </h2>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="rounded-lg p-4" :class="app.darkMode ? 'bg-slate-700/50' : 'bg-slate-50'">
          <div class="text-xs mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Monthly Depreciation</div>
          <div class="text-xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
            {{ fmtCurrency(dashboard.monthly_depreciation ?? 0) }}
          </div>
        </div>
        <div class="rounded-lg p-4" :class="app.darkMode ? 'bg-slate-700/50' : 'bg-slate-50'">
          <div class="text-xs mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Total Accumulated</div>
          <div class="text-xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
            {{ fmtCurrency(dashboard.total_accum_dep ?? 0) }}
          </div>
        </div>
        <div class="rounded-lg p-4" :class="app.darkMode ? 'bg-slate-700/50' : 'bg-slate-50'">
          <div class="text-xs mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Depreciation Rate</div>
          <div class="text-xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
            {{ dashboard.depreciation_rate ?? 0 }}%
          </div>
        </div>
        <div class="rounded-lg p-4" :class="app.darkMode ? 'bg-slate-700/50' : 'bg-slate-50'">
          <div class="text-xs mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Remaining Value</div>
          <div class="text-xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
            {{ fmtCurrency(dashboard.total_net_book_value ?? 0) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
