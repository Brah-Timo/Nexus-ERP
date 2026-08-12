<template>
  <div class="p-6 space-y-6" :class="app.darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <BarChart2 class="w-7 h-7 text-blue-500" />
          Budget Dashboard
        </h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">
          Fiscal Year {{ currentYear }} — Real-time budget performance overview
        </p>
      </div>
      <button @click="load" class="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700 text-sm font-medium">
        <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
        Refresh
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div v-for="kpi in kpiCards" :key="kpi.label"
        class="rounded-xl p-4 border"
        :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <div class="flex items-center justify-between">
          <span class="text-xs font-medium uppercase tracking-wide" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ kpi.label }}</span>
          <component :is="kpi.icon" class="w-5 h-5" :class="kpi.color" />
        </div>
        <p class="text-2xl font-bold mt-2">{{ kpi.value }}</p>
        <p class="text-xs mt-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ kpi.sub }}</p>
      </div>
    </div>

    <!-- Utilisation Bar -->
    <div class="rounded-xl p-5 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
      <h2 class="font-semibold mb-3 flex items-center gap-2">
        <TrendingUp class="w-5 h-5 text-blue-500" />
        Overall Budget Utilisation
      </h2>
      <div class="w-full rounded-full h-5 overflow-hidden" :class="app.darkMode ? 'bg-gray-700' : 'bg-gray-200'">
        <div class="h-5 rounded-full transition-all duration-700 flex items-center justify-end pr-2"
          :style="{ width: Math.min(data.kpi.utilization_pct || 0, 100) + '%' }"
          :class="utilizationClass">
          <span class="text-xs font-bold text-white">{{ (data.kpi.utilization_pct || 0).toFixed(1) }}%</span>
        </div>
      </div>
      <div class="flex justify-between text-xs mt-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">
        <span>0%</span>
        <span>Spent: {{ fmt(data.kpi.total_spent) }} / Allocated: {{ fmt(data.kpi.total_allocated) }}</span>
        <span>100%</span>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- By Year Bar Chart -->
      <div class="rounded-xl p-5 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <h2 class="font-semibold mb-4 flex items-center gap-2">
          <CalendarDays class="w-5 h-5 text-indigo-500" />
          Budget by Fiscal Year
        </h2>
        <div class="space-y-3">
          <div v-for="yr in data.by_year" :key="yr.fiscal_year">
            <div class="flex justify-between text-xs mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-600'">
              <span class="font-medium">FY {{ yr.fiscal_year }}</span>
              <span>{{ fmt(yr.spent) }} / {{ fmt(yr.allocated) }}</span>
            </div>
            <div class="w-full rounded-full h-3 overflow-hidden" :class="app.darkMode ? 'bg-gray-700' : 'bg-gray-200'">
              <div class="h-3 rounded-full bg-indigo-500 transition-all"
                :style="{ width: yr.allocated > 0 ? Math.min((yr.spent / yr.allocated) * 100, 100) + '%' : '0%' }">
              </div>
            </div>
          </div>
          <p v-if="!data.by_year?.length" class="text-center text-sm py-4" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">No data</p>
        </div>
      </div>

      <!-- Monthly Trend -->
      <div class="rounded-xl p-5 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <h2 class="font-semibold mb-4 flex items-center gap-2">
          <TrendingUp class="w-5 h-5 text-emerald-500" />
          Monthly Spend Trend ({{ currentYear }})
        </h2>
        <div class="flex items-end gap-1 h-32">
          <div v-for="m in data.monthly_trend" :key="m.month"
            class="flex-1 flex flex-col items-center gap-1">
            <div class="w-full rounded-t transition-all duration-500 bg-emerald-500 opacity-80 hover:opacity-100"
              :style="{ height: maxMonthly > 0 ? (m.amount / maxMonthly * 100) + '%' : '4px', minHeight: '4px' }"
              :title="`${m.month}: ${fmt(m.amount)}`">
            </div>
            <span class="text-xs" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">{{ m.month.slice(0,1) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Bottom Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Top Departments -->
      <div class="rounded-xl p-5 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <h2 class="font-semibold mb-4 flex items-center gap-2">
          <Building2 class="w-5 h-5 text-violet-500" />
          Top Departments by Spend
        </h2>
        <div class="space-y-3">
          <div v-for="dept in data.top_departments" :key="dept.department_name">
            <div class="flex justify-between text-sm mb-1">
              <span class="font-medium truncate max-w-xs">{{ dept.department_name }}</span>
              <span :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ fmt(dept.spent) }}</span>
            </div>
            <div class="w-full rounded-full h-2 overflow-hidden" :class="app.darkMode ? 'bg-gray-700' : 'bg-gray-200'">
              <div class="h-2 rounded-full bg-violet-500"
                :style="{ width: dept.allocated > 0 ? Math.min((dept.spent / dept.allocated) * 100, 100) + '%' : '0%' }">
              </div>
            </div>
          </div>
          <p v-if="!data.top_departments?.length" class="text-center text-sm py-4" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">No data</p>
        </div>
      </div>

      <!-- Recent Commitments -->
      <div class="rounded-xl p-5 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <h2 class="font-semibold mb-4 flex items-center gap-2">
          <Handshake class="w-5 h-5 text-amber-500" />
          Recent Commitments
        </h2>
        <div class="space-y-2">
          <div v-for="c in data.recent_commitments" :key="c.commitment_number"
            class="flex items-center justify-between py-2 border-b text-sm"
            :class="app.darkMode ? 'border-gray-700' : 'border-gray-100'">
            <div class="min-w-0">
              <p class="font-medium truncate">{{ c.commitment_number }}</p>
              <p class="text-xs truncate" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ c.vendor_name || c.description }}</p>
            </div>
            <div class="text-right flex-shrink-0 ml-3">
              <p class="font-semibold">{{ fmt(c.committed_amount) }}</p>
              <span class="text-xs px-2 py-0.5 rounded-full" :class="statusClass(c.status)">{{ c.status }}</span>
            </div>
          </div>
          <p v-if="!data.recent_commitments?.length" class="text-center text-sm py-4" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">No commitments</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  BarChart2, RefreshCw, TrendingUp, CalendarDays,
  Building2, Handshake, DollarSign, Target, AlertTriangle, CheckCircle2
} from 'lucide-vue-next'
import { budgetingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const loading = ref(false)
const currentYear = new Date().getFullYear()

interface DashData {
  kpi: Record<string, number>
  by_year: Array<{ fiscal_year: number; allocated: number; spent: number; committed: number }>
  top_departments: Array<{ department_name: string; allocated: number; spent: number }>
  by_category: Array<{ category_name: string; budget_amount: number; actual_amount: number }>
  recent_commitments: Array<Record<string, unknown>>
  monthly_trend: Array<{ month: string; amount: number }>
}

const data = ref<DashData>({
  kpi: {},
  by_year: [],
  top_departments: [],
  by_category: [],
  recent_commitments: [],
  monthly_trend: Array.from({ length: 12 }, (_, i) => ({
    month: ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'][i],
    amount: 0
  }))
})

const fmt = (v: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(v || 0)

const maxMonthly = computed(() => Math.max(...(data.value.monthly_trend?.map(m => m.amount) || [1]), 1))

const utilizationClass = computed(() => {
  const pct = data.value.kpi.utilization_pct || 0
  if (pct > 90) return 'bg-red-500'
  if (pct > 75) return 'bg-amber-500'
  return 'bg-emerald-500'
})

const kpiCards = computed(() => [
  { label: 'Total Budgets', value: data.value.kpi.total_budgets || 0, sub: `${data.value.kpi.active_budgets || 0} Active`, icon: BarChart2, color: 'text-blue-500' },
  { label: 'Total Allocated', value: fmt(data.value.kpi.total_allocated), sub: `${data.value.kpi.draft_budgets || 0} Drafts`, icon: Target, color: 'text-indigo-500' },
  { label: 'Total Spent', value: fmt(data.value.kpi.total_spent), sub: `${(data.value.kpi.utilization_pct || 0).toFixed(1)}% Utilised`, icon: DollarSign, color: 'text-emerald-500' },
  { label: 'Available Balance', value: fmt(data.value.kpi.available_balance), sub: `Committed: ${fmt(data.value.kpi.total_committed)}`, icon: CheckCircle2, color: 'text-amber-500' },
])

const statusClass = (s: string) => ({
  pending:   'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
  approved:  'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
  fulfilled: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
  cancelled: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
}[s] || 'bg-gray-100 text-gray-800')

async function load() {
  loading.value = true
  try {
    const r = await budgetingAPI.getDashboard()
    data.value = r.data
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
