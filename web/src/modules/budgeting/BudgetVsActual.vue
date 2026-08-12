<template>
  <div class="p-6" :class="app.darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <TrendingUp class="w-7 h-7 text-emerald-500" />
          Budget vs Actual
        </h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Track actual spending against approved budgets</p>
      </div>
      <button @click="load" :disabled="loading" class="flex items-center gap-2 px-4 py-2 rounded-lg bg-emerald-600 text-white hover:bg-emerald-700 text-sm font-medium">
        <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
        Refresh
      </button>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-5">
      <select v-model="filterBudget" @change="load" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option value="">All Budgets</option>
        <option v-for="b in annualBudgets" :key="b.id as string" :value="b.id">{{ b.budget_number }} — {{ b.name }}</option>
      </select>
      <select v-model="filterDept" @change="load" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option value="">All Departments</option>
        <option v-for="d in deptBudgets" :key="d.id as string" :value="d.id">{{ d.department_name }}</option>
      </select>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
      <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Total Budget</p>
        <p class="text-xl font-bold text-blue-500">{{ fmt(data.summary?.total_budget) }}</p>
      </div>
      <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Actual Spent</p>
        <p class="text-xl font-bold text-emerald-500">{{ fmt(data.summary?.total_actual) }}</p>
      </div>
      <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Committed</p>
        <p class="text-xl font-bold text-amber-500">{{ fmt(data.summary?.total_committed) }}</p>
      </div>
      <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Variance</p>
        <p class="text-xl font-bold" :class="(data.summary?.variance || 0) >= 0 ? 'text-green-500' : 'text-red-500'">
          {{ fmt(data.summary?.variance) }}
        </p>
      </div>
      <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Utilisation</p>
        <p class="text-xl font-bold" :class="(data.summary?.utilization_pct || 0) > 90 ? 'text-red-500' : 'text-emerald-500'">
          {{ (data.summary?.utilization_pct || 0).toFixed(1) }}%
        </p>
      </div>
    </div>

    <!-- By Category Chart -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
      <div class="rounded-xl p-5 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <h2 class="font-semibold mb-4 flex items-center gap-2">
          <Tag class="w-5 h-5 text-indigo-500" />
          By Category
        </h2>
        <div class="space-y-3">
          <div v-for="cat in data.by_category" :key="cat.category_name as string">
            <div class="flex justify-between text-sm mb-1">
              <span class="font-medium truncate max-w-xs">{{ cat.category_name }}</span>
              <span :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ fmt(cat.actual_amount as number) }} / {{ fmt(cat.budget_amount as number) }}</span>
            </div>
            <div class="w-full rounded-full h-2.5 overflow-hidden" :class="app.darkMode ? 'bg-gray-700' : 'bg-gray-200'">
              <div class="h-2.5 rounded-full transition-all"
                :class="(cat.utilization_pct as number) > 100 ? 'bg-red-500' : (cat.utilization_pct as number) > 80 ? 'bg-amber-500' : 'bg-indigo-500'"
                :style="{ width: Math.min(cat.utilization_pct as number, 100) + '%' }">
              </div>
            </div>
          </div>
          <p v-if="!data.by_category?.length" class="text-center py-4" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">No data</p>
        </div>
      </div>

      <!-- Monthly Actuals Bar -->
      <div class="rounded-xl p-5 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <h2 class="font-semibold mb-4 flex items-center gap-2">
          <BarChart2 class="w-5 h-5 text-emerald-500" />
          Monthly Actuals
        </h2>
        <div v-if="data.monthly_actuals?.length" class="flex items-end gap-1 h-36">
          <div v-for="m in data.monthly_actuals" :key="m.month as string"
            class="flex-1 flex flex-col items-center gap-1">
            <div class="w-full rounded-t bg-emerald-500 opacity-80 hover:opacity-100 transition-all"
              :style="{ height: maxMonthly > 0 ? (m.amount as number / maxMonthly * 100) + '%' : '4px', minHeight: '4px' }"
              :title="`${m.month}: ${fmt(m.amount as number)}`">
            </div>
            <span class="text-xs truncate w-full text-center" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">{{ String(m.month).slice(0, 3) }}</span>
          </div>
        </div>
        <p v-else class="text-center py-8" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">No actuals data</p>
      </div>
    </div>

    <!-- Line-level Detail Table -->
    <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
      <div class="px-4 py-3 border-b flex items-center gap-2" :class="app.darkMode ? 'border-gray-700' : 'border-gray-100'">
        <List class="w-4 h-4 text-blue-500" />
        <h2 class="font-semibold">Line Item Detail</h2>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'bg-gray-750 text-gray-400' : 'bg-gray-50 text-gray-500'">
              <th class="px-4 py-3 text-left font-medium">Category</th>
              <th class="px-4 py-3 text-left font-medium">Account</th>
              <th class="px-4 py-3 text-right font-medium">Budget</th>
              <th class="px-4 py-3 text-right font-medium">Q1</th>
              <th class="px-4 py-3 text-right font-medium">Q2</th>
              <th class="px-4 py-3 text-right font-medium">Q3</th>
              <th class="px-4 py-3 text-right font-medium">Q4</th>
              <th class="px-4 py-3 text-right font-medium">Actual</th>
              <th class="px-4 py-3 text-right font-medium">Committed</th>
              <th class="px-4 py-3 text-right font-medium">Variance</th>
              <th class="px-4 py-3 text-center font-medium">%</th>
            </tr>
          </thead>
          <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
            <tr v-for="li in data.line_items" :key="li.id as string"
              class="transition-colors" :class="app.darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-50'">
              <td class="px-4 py-2">{{ li.category_name }}</td>
              <td class="px-4 py-2 text-xs" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ li.account_code }} {{ li.account_name }}</td>
              <td class="px-4 py-2 text-right">{{ fmt(li.budget_amount as number) }}</td>
              <td class="px-4 py-2 text-right text-xs">{{ fmt(li.q1_amount as number) }}</td>
              <td class="px-4 py-2 text-right text-xs">{{ fmt(li.q2_amount as number) }}</td>
              <td class="px-4 py-2 text-right text-xs">{{ fmt(li.q3_amount as number) }}</td>
              <td class="px-4 py-2 text-right text-xs">{{ fmt(li.q4_amount as number) }}</td>
              <td class="px-4 py-2 text-right">{{ fmt(li.actual_amount as number) }}</td>
              <td class="px-4 py-2 text-right text-amber-500">{{ fmt(li.committed_amount as number) }}</td>
              <td class="px-4 py-2 text-right font-medium" :class="(li.variance as number) >= 0 ? 'text-green-500' : 'text-red-500'">
                {{ fmt(li.variance as number) }}
              </td>
              <td class="px-4 py-2 text-center text-xs">
                <span class="px-1.5 py-0.5 rounded text-xs"
                  :class="(li.utilization_pct as number) > 100 ? 'bg-red-100 text-red-700' : (li.utilization_pct as number) > 80 ? 'bg-amber-100 text-amber-700' : 'bg-green-100 text-green-700'">
                  {{ (li.utilization_pct as number).toFixed(0) }}%
                </span>
              </td>
            </tr>
            <tr v-if="!data.line_items?.length">
              <td colspan="11" class="px-4 py-12 text-center" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">No data</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { TrendingUp, RefreshCw, Tag, BarChart2, List } from 'lucide-vue-next'
import { budgetingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const loading = ref(false)
const filterBudget = ref('')
const filterDept = ref('')
const annualBudgets = ref<Record<string, unknown>[]>([])
const deptBudgets = ref<Record<string, unknown>[]>([])

interface VsActualData {
  summary: Record<string, number>
  by_category: Record<string, unknown>[]
  monthly_actuals: Record<string, unknown>[]
  line_items: Record<string, unknown>[]
}

const data = ref<VsActualData>({ summary: {}, by_category: [], monthly_actuals: [], line_items: [] })

const fmt = (v: number | undefined) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(v || 0)
const maxMonthly = computed(() => Math.max(...(data.value.monthly_actuals?.map(m => m.amount as number) || [1]), 1))

const inputClass = computed(() => app.darkMode
  ? 'bg-gray-700 border-gray-600 text-white focus:border-emerald-500 focus:outline-none'
  : 'bg-white border-gray-300 text-gray-900 focus:border-emerald-500 focus:ring-1 focus:ring-emerald-500 focus:outline-none')

async function load() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (filterBudget.value) params.annual_budget_id = filterBudget.value
    if (filterDept.value) params.department_budget_id = filterDept.value
    const r = await budgetingAPI.getBudgetVsActual(params)
    data.value = r.data
  } catch (e) { console.error(e) } finally { loading.value = false }
}

async function loadRefs() {
  const [bRes, dRes] = await Promise.all([budgetingAPI.listAnnualBudgets(), budgetingAPI.listDepartmentBudgets()])
  annualBudgets.value = bRes.data.data || []
  deptBudgets.value = dRes.data.data || []
}

onMounted(async () => { await loadRefs(); await load() })
</script>
