<template>
  <div class="p-6" :class="app.darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <BarChart3 class="w-7 h-7 text-cyan-500" />
          Budget Reports
        </h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Comprehensive budgeting analytics and reports</p>
      </div>
    </div>

    <!-- Controls -->
    <div class="flex flex-wrap gap-3 mb-6">
      <select v-model="reportType" @change="load" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option value="">Overview</option>
        <option value="budget_summary">Budget Summary</option>
        <option value="department_performance">Department Performance</option>
        <option value="variance_analysis">Variance Analysis</option>
        <option value="commitment_report">Commitment Report</option>
        <option value="revision_history">Revision History</option>
      </select>
      <select v-model="fiscalYear" @change="load" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option v-for="y in yearOptions" :key="y" :value="y">FY {{ y }}</option>
      </select>
      <button @click="load" :disabled="loading" class="flex items-center gap-2 px-4 py-2 rounded-lg bg-cyan-600 text-white hover:bg-cyan-700 text-sm font-medium">
        <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
        Run Report
      </button>
    </div>

    <!-- Overview -->
    <div v-if="reportType === '' && data" class="space-y-5">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
          <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Total Budgets</p>
          <p class="text-2xl font-bold text-cyan-500">{{ overviewData.total_budgets }}</p>
          <p class="text-xs mt-1" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">{{ overviewData.active_budgets }} Active</p>
        </div>
        <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
          <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Total Allocated</p>
          <p class="text-2xl font-bold text-blue-500">{{ fmt(overviewData.total_allocated) }}</p>
        </div>
        <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
          <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Total Spent</p>
          <p class="text-2xl font-bold text-emerald-500">{{ fmt(overviewData.total_spent) }}</p>
          <p class="text-xs mt-1" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">{{ (overviewData.utilization_pct || 0).toFixed(1) }}% utilised</p>
        </div>
        <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
          <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Available</p>
          <p class="text-2xl font-bold" :class="(overviewData.available || 0) >= 0 ? 'text-green-500' : 'text-red-500'">{{ fmt(overviewData.available) }}</p>
          <p class="text-xs mt-1" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">Committed: {{ fmt(overviewData.total_committed) }}</p>
        </div>
      </div>

      <!-- Utilisation bar -->
      <div class="rounded-xl p-5 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <h3 class="font-semibold mb-3">Budget Utilisation — FY {{ fiscalYear }}</h3>
        <div class="w-full rounded-full h-6 overflow-hidden" :class="app.darkMode ? 'bg-gray-700' : 'bg-gray-200'">
          <div class="h-6 rounded-full flex items-center justify-end pr-3 transition-all duration-700 font-semibold text-white text-sm"
            :class="(overviewData.utilization_pct || 0) > 90 ? 'bg-red-500' : (overviewData.utilization_pct || 0) > 75 ? 'bg-amber-500' : 'bg-emerald-500'"
            :style="{ width: Math.min(overviewData.utilization_pct || 0, 100) + '%' }">
            {{ (overviewData.utilization_pct || 0).toFixed(1) }}%
          </div>
        </div>
      </div>
    </div>

    <!-- Table Reports -->
    <div v-if="tableData.length > 0" class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
      <!-- Budget Summary -->
      <template v-if="reportType === 'budget_summary'">
        <div class="px-4 py-3 border-b" :class="app.darkMode ? 'border-gray-700' : 'border-gray-100'">
          <h2 class="font-semibold">Budget Summary — FY {{ fiscalYear }}</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead><tr :class="app.darkMode ? 'bg-gray-700 text-gray-400' : 'bg-gray-50 text-gray-500'">
              <th class="px-4 py-3 text-left font-medium">Budget #</th>
              <th class="px-4 py-3 text-left font-medium">Name</th>
              <th class="px-4 py-3 text-center font-medium">Type</th>
              <th class="px-4 py-3 text-center font-medium">Status</th>
              <th class="px-4 py-3 text-right font-medium">Allocated</th>
              <th class="px-4 py-3 text-right font-medium">Spent</th>
              <th class="px-4 py-3 text-right font-medium">Available</th>
              <th class="px-4 py-3 text-right font-medium">Utilisation</th>
            </tr></thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
              <tr v-for="r in tableData" :key="r.budget_number as string" :class="app.darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-50'">
                <td class="px-4 py-3 font-mono font-medium text-blue-500">{{ r.budget_number }}</td>
                <td class="px-4 py-3 font-medium">{{ r.name }}</td>
                <td class="px-4 py-3 text-center capitalize text-xs">{{ r.budget_type }}</td>
                <td class="px-4 py-3 text-center"><span class="px-2 py-0.5 rounded-full text-xs capitalize" :class="statusClass(r.status as string)">{{ r.status }}</span></td>
                <td class="px-4 py-3 text-right">{{ fmt(r.total_amount as number) }}</td>
                <td class="px-4 py-3 text-right">{{ fmt(r.spent_amount as number) }}</td>
                <td class="px-4 py-3 text-right" :class="(r.available_amount as number) >= 0 ? 'text-green-500' : 'text-red-500'">{{ fmt(r.available_amount as number) }}</td>
                <td class="px-4 py-3 text-right">{{ (r.utilization_pct as number).toFixed(1) }}%</td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>

      <!-- Department Performance -->
      <template v-if="reportType === 'department_performance'">
        <div class="px-4 py-3 border-b" :class="app.darkMode ? 'border-gray-700' : 'border-gray-100'">
          <h2 class="font-semibold">Department Performance — FY {{ fiscalYear }}</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead><tr :class="app.darkMode ? 'bg-gray-700 text-gray-400' : 'bg-gray-50 text-gray-500'">
              <th class="px-4 py-3 text-left font-medium">Department</th>
              <th class="px-4 py-3 text-left font-medium">Budget</th>
              <th class="px-4 py-3 text-right font-medium">Allocated</th>
              <th class="px-4 py-3 text-right font-medium">Spent</th>
              <th class="px-4 py-3 text-right font-medium">Available</th>
              <th class="px-4 py-3 text-center font-medium">Utilisation</th>
            </tr></thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
              <tr v-for="r in tableData" :key="r.department_name as string" :class="app.darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-50'">
                <td class="px-4 py-3"><p class="font-medium">{{ r.department_name }}</p><p class="text-xs" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">{{ r.department_code }}</p></td>
                <td class="px-4 py-3 text-xs">{{ r.budget_number }}</td>
                <td class="px-4 py-3 text-right">{{ fmt(r.allocated_amount as number) }}</td>
                <td class="px-4 py-3 text-right">{{ fmt(r.spent_amount as number) }}</td>
                <td class="px-4 py-3 text-right" :class="(r.available_amount as number) >= 0 ? 'text-green-500' : 'text-red-500'">{{ fmt(r.available_amount as number) }}</td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2">
                    <div class="flex-1 rounded-full h-2 overflow-hidden" :class="app.darkMode ? 'bg-gray-700' : 'bg-gray-200'">
                      <div class="h-2 rounded-full" :class="(r.utilization_pct as number) > 90 ? 'bg-red-500' : 'bg-violet-500'" :style="{ width: Math.min(r.utilization_pct as number, 100) + '%' }"></div>
                    </div>
                    <span class="text-xs w-10">{{ (r.utilization_pct as number).toFixed(0) }}%</span>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>

      <!-- Variance Analysis -->
      <template v-if="reportType === 'variance_analysis'">
        <div class="px-4 py-3 border-b" :class="app.darkMode ? 'border-gray-700' : 'border-gray-100'">
          <h2 class="font-semibold">Variance Analysis — FY {{ fiscalYear }}</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead><tr :class="app.darkMode ? 'bg-gray-700 text-gray-400' : 'bg-gray-50 text-gray-500'">
              <th class="px-4 py-3 text-left font-medium">Category</th>
              <th class="px-4 py-3 text-right font-medium">Budget</th>
              <th class="px-4 py-3 text-right font-medium">Actual</th>
              <th class="px-4 py-3 text-right font-medium">Variance</th>
              <th class="px-4 py-3 text-right font-medium">Q1</th>
              <th class="px-4 py-3 text-right font-medium">Q2</th>
              <th class="px-4 py-3 text-right font-medium">Q3</th>
              <th class="px-4 py-3 text-right font-medium">Q4</th>
              <th class="px-4 py-3 text-center font-medium">%</th>
            </tr></thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
              <tr v-for="r in tableData" :key="r.category_name as string" :class="app.darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-50'">
                <td class="px-4 py-3 font-medium">{{ r.category_name }}</td>
                <td class="px-4 py-3 text-right">{{ fmt(r.budget_amount as number) }}</td>
                <td class="px-4 py-3 text-right">{{ fmt(r.actual_amount as number) }}</td>
                <td class="px-4 py-3 text-right font-semibold" :class="(r.variance as number) >= 0 ? 'text-green-500' : 'text-red-500'">{{ fmt(r.variance as number) }}</td>
                <td class="px-4 py-3 text-right text-xs">{{ fmt(r.q1_budget as number) }}</td>
                <td class="px-4 py-3 text-right text-xs">{{ fmt(r.q2_budget as number) }}</td>
                <td class="px-4 py-3 text-right text-xs">{{ fmt(r.q3_budget as number) }}</td>
                <td class="px-4 py-3 text-right text-xs">{{ fmt(r.q4_budget as number) }}</td>
                <td class="px-4 py-3 text-center"><span class="px-1.5 py-0.5 rounded text-xs" :class="(r.utilization_pct as number) > 100 ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'">{{ (r.utilization_pct as number).toFixed(0) }}%</span></td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>

      <!-- Commitment Report -->
      <template v-if="reportType === 'commitment_report'">
        <div class="px-4 py-3 border-b" :class="app.darkMode ? 'border-gray-700' : 'border-gray-100'">
          <h2 class="font-semibold">Commitment Report — FY {{ fiscalYear }}</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead><tr :class="app.darkMode ? 'bg-gray-700 text-gray-400' : 'bg-gray-50 text-gray-500'">
              <th class="px-4 py-3 text-left font-medium">Commitment #</th>
              <th class="px-4 py-3 text-left font-medium">Vendor</th>
              <th class="px-4 py-3 text-center font-medium">Type</th>
              <th class="px-4 py-3 text-right font-medium">Committed</th>
              <th class="px-4 py-3 text-right font-medium">Fulfilled</th>
              <th class="px-4 py-3 text-right font-medium">Remaining</th>
              <th class="px-4 py-3 text-center font-medium">Date</th>
              <th class="px-4 py-3 text-center font-medium">Status</th>
            </tr></thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
              <tr v-for="r in tableData" :key="r.commitment_number as string" :class="app.darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-50'">
                <td class="px-4 py-3 font-mono text-orange-500">{{ r.commitment_number }}</td>
                <td class="px-4 py-3"><p class="font-medium truncate max-w-xs">{{ r.vendor_name || r.description }}</p></td>
                <td class="px-4 py-3 text-center capitalize text-xs">{{ String(r.commitment_type).replace('_', ' ') }}</td>
                <td class="px-4 py-3 text-right">{{ fmt(r.committed_amount as number) }}</td>
                <td class="px-4 py-3 text-right text-emerald-500">{{ fmt(r.fulfilled_amount as number) }}</td>
                <td class="px-4 py-3 text-right text-amber-500">{{ fmt(r.remaining_amount as number) }}</td>
                <td class="px-4 py-3 text-center text-xs">{{ r.commitment_date }}</td>
                <td class="px-4 py-3 text-center"><span class="px-2 py-0.5 rounded-full text-xs capitalize" :class="statusClass(r.status as string)">{{ r.status }}</span></td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>

      <!-- Revision History -->
      <template v-if="reportType === 'revision_history'">
        <div class="px-4 py-3 border-b" :class="app.darkMode ? 'border-gray-700' : 'border-gray-100'">
          <h2 class="font-semibold">Revision History — FY {{ fiscalYear }}</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead><tr :class="app.darkMode ? 'bg-gray-700 text-gray-400' : 'bg-gray-50 text-gray-500'">
              <th class="px-4 py-3 text-left font-medium">Revision #</th>
              <th class="px-4 py-3 text-left font-medium">Budget</th>
              <th class="px-4 py-3 text-center font-medium">Type</th>
              <th class="px-4 py-3 text-right font-medium">Original</th>
              <th class="px-4 py-3 text-right font-medium">Revised</th>
              <th class="px-4 py-3 text-right font-medium">Change</th>
              <th class="px-4 py-3 text-left font-medium">Reason</th>
              <th class="px-4 py-3 text-center font-medium">Status</th>
              <th class="px-4 py-3 text-center font-medium">Date</th>
            </tr></thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
              <tr v-for="r in tableData" :key="r.revision_number as string" :class="app.darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-50'">
                <td class="px-4 py-3 font-mono text-amber-500">{{ r.revision_number }}</td>
                <td class="px-4 py-3 text-xs">{{ r.budget_number }}</td>
                <td class="px-4 py-3 text-center capitalize text-xs">{{ r.revision_type }}</td>
                <td class="px-4 py-3 text-right">{{ fmt(r.original_amount as number) }}</td>
                <td class="px-4 py-3 text-right">{{ fmt(r.revised_amount as number) }}</td>
                <td class="px-4 py-3 text-right font-semibold" :class="(r.change_amount as number) >= 0 ? 'text-green-500' : 'text-red-500'">{{ fmt(r.change_amount as number) }}</td>
                <td class="px-4 py-3 max-w-xs truncate text-xs" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ r.reason }}</td>
                <td class="px-4 py-3 text-center"><span class="px-2 py-0.5 rounded-full text-xs" :class="statusClass(r.status as string)">{{ r.status === 'active' ? 'Approved' : r.status }}</span></td>
                <td class="px-4 py-3 text-center text-xs">{{ String(r.created_at).slice(0,10) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <div v-if="reportType !== '' && !tableData.length && !loading" class="rounded-xl border p-12 text-center" :class="app.darkMode ? 'bg-gray-800 border-gray-700 text-gray-500' : 'bg-white border-gray-200 text-gray-400'">
      No data for the selected report
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { BarChart3, RefreshCw } from 'lucide-vue-next'
import { budgetingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const loading = ref(false)
const reportType = ref('')
const currentYear = new Date().getFullYear()
const fiscalYear = ref(String(currentYear))
const yearOptions = Array.from({ length: 5 }, (_, i) => currentYear - 2 + i)

const data = ref<Record<string, unknown> | null>(null)
const tableData = ref<Record<string, unknown>[]>([])
const overviewData = computed<Record<string, number>>(() => {
  if (data.value?.data && typeof data.value.data === 'object' && !Array.isArray(data.value.data)) {
    return data.value.data as Record<string, number>
  }
  return {}
})

const fmt = (v: number | undefined) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(v || 0)

const inputClass = computed(() => app.darkMode
  ? 'bg-gray-700 border-gray-600 text-white focus:border-cyan-500 focus:outline-none'
  : 'bg-white border-gray-300 text-gray-900 focus:border-cyan-500 focus:ring-1 focus:ring-cyan-500 focus:outline-none')

const statusClass = (s: string) => ({
  draft:     'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
  active:    'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  locked:    'bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300',
  closed:    'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  cancelled: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
  pending:   'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300',
  approved:  'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  fulfilled: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
}[s] || 'bg-gray-100 text-gray-700')

async function load() {
  loading.value = true
  try {
    const params: Record<string, string> = { fiscal_year: fiscalYear.value }
    if (reportType.value) params.type = reportType.value
    const r = await budgetingAPI.getReports(params)
    data.value = r.data
    tableData.value = Array.isArray(r.data.data) ? r.data.data : []
  } catch (e) { console.error(e) } finally { loading.value = false }
}

onMounted(load)
</script>
