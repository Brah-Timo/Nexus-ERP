<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Target, Plus, RefreshCw, Loader2, CheckCircle, X,
  TrendingUp, TrendingDown, AlertTriangle, BarChart3,
  Calendar, Pencil, DollarSign, Filter, ChevronDown,
  Activity, Hash, FileText
} from 'lucide-vue-next'
import { accountingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ─────────────────────────────────────────────────────────────────────
// Matches actual DB schema: budgets header + budget_lines with per-month columns
interface BudgetLine {
  id?: string
  budget_id?: string
  account_id: string | null
  account_code: string
  account_name: string
  cost_center_id: string | null
  jan: number; feb: number; mar: number; apr: number
  may: number; jun: number; jul: number; aug: number
  sep: number; oct: number; nov: number; dec: number
  total_budget: number
  total_actual: number
}

interface Budget {
  id: string
  fiscal_year_id: string | null
  name: string
  description: string
  status: string
  total_budget: number
  total_actual: number
  lines?: BudgetLine[]
}

interface Account {
  id: string
  code: string
  name: string
  type: string
}

interface CostCenter {
  id: string
  code: string
  name: string
}

// ─── State ─────────────────────────────────────────────────────────────────────
const budgets     = ref<Budget[]>([])
const accounts    = ref<Account[]>([])
const costCenters = ref<CostCenter[]>([])
const loading     = ref(true)
const saving      = ref(false)
const showModal   = ref(false)
const selectedBudget = ref<Budget | null>(null)
const viewTab     = ref<'overview' | 'lines'>('overview')

const MONTHS = ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec']
const MONTH_KEYS = ['jan','feb','mar','apr','may','jun','jul','aug','sep','oct','nov','dec'] as const

const form = ref({
  name: '',
  description: '',
  fiscal_year_id: '',
  status: 'draft',
  lines: [] as BudgetLine[]
})

// ─── Computed ──────────────────────────────────────────────────────────────────
const stats = computed(() => {
  const totalBudget  = budgets.value.reduce((s, b) => s + (Number(b.total_budget) || 0), 0)
  const totalActual  = budgets.value.reduce((s, b) => s + (Number(b.total_actual) || 0), 0)
  const overBudget   = budgets.value.filter(b => Number(b.total_actual) > Number(b.total_budget)).length
  const utilization  = totalBudget > 0 ? (totalActual / totalBudget) * 100 : 0
  return { totalBudget, totalActual, overBudget, utilization }
})

// Monthly chart data based on selected budget or aggregate
const monthlyData = computed(() => {
  const src = selectedBudget.value?.lines ?? []
  return MONTHS.map((label, i) => {
    const key = MONTH_KEYS[i]
    const planned = src.reduce((s, l) => s + (Number((l as any)[key]) || 0), 0)
    return { label, planned, actual: 0 }
  })
})

const maxMonthlyValue = computed(() =>
  Math.max(...monthlyData.value.map(m => m.planned), 1)
)

// ─── Methods ───────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const [bRes, aRes, ccRes] = await Promise.all([
      accountingAPI.getBudgets(),
      accountingAPI.getChartOfAccounts(),
      accountingAPI.getCostCenters()
    ])
    budgets.value = bRes.data ?? []
    accounts.value = aRes.data ?? []
    costCenters.value = ccRes.data ?? []
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Failed to load budgets', 'error')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = {
    name: '',
    description: '',
    fiscal_year_id: '',
    status: 'draft',
    lines: [newBlankLine()]
  }
  showModal.value = true
}

function newBlankLine(): BudgetLine {
  return {
    account_id: null, account_code: '', account_name: '',
    cost_center_id: null,
    jan: 0, feb: 0, mar: 0, apr: 0,
    may: 0, jun: 0, jul: 0, aug: 0,
    sep: 0, oct: 0, nov: 0, dec: 0,
    total_budget: 0, total_actual: 0
  }
}

function addFormLine() {
  form.value.lines.push(newBlankLine())
}

function removeFormLine(i: number) {
  if (form.value.lines.length > 1) form.value.lines.splice(i, 1)
}

function onLineAccountChange(i: number, accountId: string) {
  const acc = accounts.value.find(a => a.id === accountId)
  if (acc) {
    form.value.lines[i].account_id = acc.id
    form.value.lines[i].account_code = acc.code
    form.value.lines[i].account_name = acc.name
  }
}

function lineTotalBudget(l: BudgetLine): number {
  return MONTH_KEYS.reduce((s, k) => s + (Number((l as any)[k]) || 0), 0)
}

async function save() {
  if (!form.value.name) {
    app.addToast('Budget name is required', 'error'); return
  }
  saving.value = true
  try {
    const payload = {
      ...form.value,
      lines: form.value.lines.filter(l => l.account_id)
    }
    await accountingAPI.createBudget(payload)
    app.addToast('Budget created', 'success')
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

async function selectBudget(b: Budget) {
  selectedBudget.value = b
  viewTab.value = 'lines'
  // Could load full budget with lines here if needed
}

function getVarianceClass(budget: number, actual: number): string {
  if (actual <= budget * 0.8)  return 'text-emerald-400'
  if (actual <= budget)        return 'text-amber-400'
  return 'text-rose-400'
}

function fmt(v: number | undefined) {
  return new Intl.NumberFormat('en-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(v ?? 0)
}

function statusBadge(s: string) {
  const map: Record<string, string> = {
    draft: 'text-slate-400 bg-slate-500/10 border-slate-500/20',
    active: 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20',
    closed: 'text-rose-400 bg-rose-500/10 border-rose-500/20'
  }
  return map[s] ?? map.draft
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col h-full"
       :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">

    <!-- ── Header ──────────────────────────────────────────────────────────── -->
    <div class="border-b px-6 py-4 flex-shrink-0"
         :class="app.darkMode ? 'border-slate-800/60 bg-slate-900/50' : 'border-slate-200 bg-white'">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-lg flex items-center justify-center"
               :class="app.darkMode ? 'bg-rose-500/15 border border-rose-500/25' : 'bg-rose-100 border border-rose-200'">
            <Target class="w-4 h-4" :class="app.darkMode ? 'text-rose-400' : 'text-rose-600'" />
          </div>
          <div>
            <h1 class="text-[15px] font-semibold leading-tight">Budgets</h1>
            <p class="text-[11px] leading-tight" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">
              Budget planning &amp; variance tracking
            </p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            class="h-8 px-3 rounded-lg border text-xs font-medium inline-flex items-center gap-1.5 transition-all disabled:opacity-50"
            :class="app.darkMode
              ? 'border-slate-700/60 bg-slate-800/60 text-slate-400 hover:text-slate-200'
              : 'border-slate-300 bg-white text-slate-500 hover:text-slate-700'">
            <RefreshCw class="w-3.5 h-3.5" :class="loading ? 'animate-spin' : ''" />
            Refresh
          </button>
          <button @click="openCreate"
            class="h-8 px-3 rounded-lg bg-rose-600 hover:bg-rose-500 text-white text-xs font-medium inline-flex items-center gap-1.5 transition-all shadow-lg shadow-rose-900/30">
            <Plus class="w-3.5 h-3.5" />
            New Budget
          </button>
        </div>
      </div>
    </div>

    <!-- ── KPI Cards ────────────────────────────────────────────────────────── -->
    <div class="px-6 py-4 flex-shrink-0 grid grid-cols-4 gap-3">
      <div class="rounded-xl p-4 border"
           :class="app.darkMode ? 'bg-slate-900/70 border-slate-800/50' : 'bg-white border-slate-200 shadow-sm'">
        <p class="text-[11px] font-semibold uppercase tracking-wider mb-2"
           :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Total Budget</p>
        <p class="text-2xl font-bold" :class="app.darkMode ? 'text-slate-100' : 'text-slate-900'">
          {{ fmt(stats.totalBudget) }}
        </p>
        <p class="text-[11px] mt-1" :class="app.darkMode ? 'text-slate-500' : 'text-slate-500'">DZD</p>
      </div>
      <div class="rounded-xl p-4 border"
           :class="app.darkMode ? 'bg-slate-900/70 border-slate-800/50' : 'bg-white border-slate-200 shadow-sm'">
        <p class="text-[11px] font-semibold uppercase tracking-wider mb-2"
           :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Actual Spent</p>
        <p class="text-2xl font-bold" :class="app.darkMode ? 'text-slate-100' : 'text-slate-900'">
          {{ fmt(stats.totalActual) }}
        </p>
        <p class="text-[11px] mt-1" :class="app.darkMode ? 'text-slate-500' : 'text-slate-500'">DZD</p>
      </div>
      <div class="rounded-xl p-4 border"
           :class="app.darkMode ? 'bg-slate-900/70 border-slate-800/50' : 'bg-white border-slate-200 shadow-sm'">
        <p class="text-[11px] font-semibold uppercase tracking-wider mb-2 text-amber-400">Utilization</p>
        <p class="text-2xl font-bold" :class="stats.utilization > 100 ? 'text-rose-400' : 'text-amber-400'">
          {{ stats.utilization.toFixed(1) }}%
        </p>
        <div class="mt-2 h-1.5 rounded-full overflow-hidden"
             :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
          <div class="h-full rounded-full transition-all"
               :class="stats.utilization > 100 ? 'bg-rose-500' : 'bg-amber-500'"
               :style="{ width: Math.min(stats.utilization, 100) + '%' }" />
        </div>
      </div>
      <div class="rounded-xl p-4 border"
           :class="app.darkMode ? 'bg-rose-500/10 border-rose-500/20' : 'bg-rose-50 border-rose-200 shadow-sm'">
        <p class="text-[11px] font-semibold uppercase tracking-wider mb-2 text-rose-400">Over Budget</p>
        <p class="text-2xl font-bold text-rose-400">{{ stats.overBudget }}</p>
        <p class="text-[11px] mt-1" :class="app.darkMode ? 'text-slate-500' : 'text-slate-500'">budgets exceeded</p>
      </div>
    </div>

    <!-- ── Main Content ──────────────────────────────────────────────────────── -->
    <div class="flex-1 overflow-hidden px-6 pb-6 flex gap-4">

      <!-- Budget List -->
      <div class="w-80 flex-shrink-0 rounded-xl border overflow-hidden"
           :class="app.darkMode ? 'border-slate-800/60 bg-slate-900/40' : 'border-slate-200 bg-white shadow-sm'">
        <div class="px-4 py-3 border-b flex items-center justify-between"
             :class="app.darkMode ? 'border-slate-800/60' : 'border-slate-100'">
          <span class="text-xs font-semibold uppercase tracking-wider"
                :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">All Budgets</span>
          <span class="text-xs" :class="app.darkMode ? 'text-slate-600' : 'text-slate-400'">{{ budgets.length }}</span>
        </div>
        <div class="overflow-y-auto" style="max-height: calc(100% - 48px);">
          <div v-if="loading" class="flex items-center justify-center py-10">
            <Loader2 class="w-6 h-6 text-rose-400 animate-spin" />
          </div>
          <div v-else-if="budgets.length === 0" class="py-10 text-center px-4">
            <Target class="w-8 h-8 mx-auto mb-2 opacity-30" :class="app.darkMode ? 'text-slate-600' : 'text-slate-300'" />
            <p class="text-sm" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">No budgets yet</p>
          </div>
          <div v-for="b in budgets" :key="b.id"
               @click="selectBudget(b)"
               class="px-4 py-3 border-b cursor-pointer transition-colors"
               :class="[
                 app.darkMode ? 'border-slate-800/30 hover:bg-slate-800/30' : 'border-slate-100 hover:bg-slate-50',
                 selectedBudget?.id === b.id
                   ? (app.darkMode ? 'bg-rose-500/10 border-l-2 border-l-rose-500' : 'bg-rose-50 border-l-2 border-l-rose-500')
                   : ''
               ]">
            <div class="flex items-center justify-between mb-1">
              <span class="text-sm font-medium truncate"
                    :class="app.darkMode ? 'text-slate-200' : 'text-slate-800'">{{ b.name }}</span>
              <span :class="['text-[10px] px-1.5 py-0.5 rounded-full border font-medium', statusBadge(b.status)]">
                {{ b.status }}
              </span>
            </div>
            <div class="text-[11px] flex items-center justify-between"
                 :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">
              <span>{{ fmt(b.total_budget) }} DZD</span>
              <span :class="getVarianceClass(b.total_budget, b.total_actual)">
                {{ fmt(b.total_actual) }} actual
              </span>
            </div>
            <div v-if="b.total_budget > 0" class="mt-2 h-1 rounded-full overflow-hidden"
                 :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
              <div class="h-full rounded-full transition-all"
                   :class="b.total_actual > b.total_budget ? 'bg-rose-500' : 'bg-rose-400'"
                   :style="{ width: Math.min(100, (b.total_actual / b.total_budget) * 100) + '%' }" />
            </div>
          </div>
        </div>
      </div>

      <!-- Budget Detail / Monthly Chart -->
      <div class="flex-1 rounded-xl border overflow-hidden flex flex-col"
           :class="app.darkMode ? 'border-slate-800/60 bg-slate-900/40' : 'border-slate-200 bg-white shadow-sm'">

        <!-- No selection state -->
        <div v-if="!selectedBudget" class="flex-1 flex flex-col items-center justify-center gap-3 py-10">
          <Target class="w-12 h-12 opacity-20" :class="app.darkMode ? 'text-slate-600' : 'text-slate-300'" />
          <p class="text-sm" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Select a budget to view details</p>
        </div>

        <template v-else>
          <!-- Detail header -->
          <div class="px-5 py-4 border-b flex items-center justify-between flex-shrink-0"
               :class="app.darkMode ? 'border-slate-800/60' : 'border-slate-100'">
            <div>
              <h2 class="text-sm font-semibold" :class="app.darkMode ? 'text-slate-100' : 'text-slate-800'">
                {{ selectedBudget.name }}
              </h2>
              <p class="text-[11px] mt-0.5" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">
                {{ selectedBudget.description || 'No description' }}
              </p>
            </div>
            <div class="flex items-center gap-2">
              <div class="flex rounded-lg overflow-hidden border"
                   :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
                <button @click="viewTab='overview'"
                  class="h-7 px-3 text-xs font-medium transition-colors"
                  :class="viewTab === 'overview'
                    ? 'bg-rose-600 text-white'
                    : (app.darkMode ? 'bg-slate-800 text-slate-400 hover:text-slate-200' : 'bg-white text-slate-500 hover:text-slate-700')">
                  Overview
                </button>
                <button @click="viewTab='lines'"
                  class="h-7 px-3 text-xs font-medium transition-colors"
                  :class="viewTab === 'lines'
                    ? 'bg-rose-600 text-white'
                    : (app.darkMode ? 'bg-slate-800 text-slate-400 hover:text-slate-200' : 'bg-white text-slate-500 hover:text-slate-700')">
                  Lines
                </button>
              </div>
            </div>
          </div>

          <!-- Overview Tab: Monthly Bar Chart -->
          <div v-if="viewTab === 'overview'" class="flex-1 overflow-auto px-5 py-4">
            <div class="mb-4">
              <h3 class="text-xs font-semibold uppercase tracking-wider mb-3"
                  :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Monthly Budget Distribution</h3>
              <div class="flex items-end gap-1.5 h-32">
                <div v-for="(m, i) in monthlyData" :key="i" class="flex-1 flex flex-col items-center gap-1">
                  <div class="w-full relative flex-1 flex items-end">
                    <div class="w-full rounded-t transition-all"
                         :class="app.darkMode ? 'bg-rose-500/70' : 'bg-rose-500'"
                         :style="{ height: maxMonthlyValue > 0 ? (m.planned / maxMonthlyValue * 100) + '%' : '4px', minHeight: '2px' }" />
                  </div>
                  <span class="text-[9px] font-medium"
                        :class="app.darkMode ? 'text-slate-600' : 'text-slate-400'">{{ m.label }}</span>
                </div>
              </div>
            </div>

            <!-- Summary cards -->
            <div class="grid grid-cols-2 gap-3 mt-4">
              <div class="rounded-lg p-3 border"
                   :class="app.darkMode ? 'bg-slate-800/40 border-slate-700/40' : 'bg-slate-50 border-slate-200'">
                <p class="text-[10px] font-semibold uppercase tracking-wider mb-1"
                   :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Total Budget</p>
                <p class="text-lg font-bold text-rose-400">{{ fmt(selectedBudget.total_budget) }}</p>
                <p class="text-[10px] mt-0.5" :class="app.darkMode ? 'text-slate-600' : 'text-slate-500'">DZD planned</p>
              </div>
              <div class="rounded-lg p-3 border"
                   :class="app.darkMode ? 'bg-slate-800/40 border-slate-700/40' : 'bg-slate-50 border-slate-200'">
                <p class="text-[10px] font-semibold uppercase tracking-wider mb-1"
                   :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Actual</p>
                <p class="text-lg font-bold" :class="getVarianceClass(selectedBudget.total_budget, selectedBudget.total_actual)">
                  {{ fmt(selectedBudget.total_actual) }}
                </p>
                <p class="text-[10px] mt-0.5" :class="app.darkMode ? 'text-slate-600' : 'text-slate-500'">DZD spent</p>
              </div>
            </div>
          </div>

          <!-- Lines Tab -->
          <div v-else class="flex-1 overflow-auto">
            <div v-if="!selectedBudget.lines || selectedBudget.lines.length === 0"
                 class="flex flex-col items-center justify-center py-10">
              <FileText class="w-8 h-8 opacity-20 mb-2" :class="app.darkMode ? 'text-slate-600' : 'text-slate-300'" />
              <p class="text-sm" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">No budget lines</p>
            </div>
            <table v-else class="w-full text-xs border-collapse">
              <thead class="sticky top-0 z-10">
                <tr :class="app.darkMode ? 'bg-slate-900 border-b border-slate-800' : 'bg-slate-50 border-b border-slate-200'">
                  <th class="text-left px-3 py-2 font-semibold uppercase tracking-wider"
                      :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Account</th>
                  <th v-for="m in MONTHS" :key="m" class="text-right px-2 py-2 font-semibold uppercase tracking-wider"
                      :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ m }}</th>
                  <th class="text-right px-3 py-2 font-semibold uppercase tracking-wider"
                      :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Total</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="line in selectedBudget.lines" :key="line.id"
                    class="border-b transition-colors"
                    :class="app.darkMode ? 'border-slate-800/30 hover:bg-slate-800/20' : 'border-slate-100 hover:bg-slate-50'">
                  <td class="px-3 py-2">
                    <span class="font-mono" :class="app.darkMode ? 'text-rose-300' : 'text-rose-600'">{{ line.account_code }}</span>
                    <span class="ml-1.5" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">{{ line.account_name }}</span>
                  </td>
                  <td v-for="(key, i) in MONTH_KEYS" :key="key" class="text-right px-2 py-2 font-mono"
                      :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
                    {{ fmt((line as any)[key]) }}
                  </td>
                  <td class="text-right px-3 py-2 font-mono font-semibold"
                      :class="app.darkMode ? 'text-rose-300' : 'text-rose-600'">
                    {{ fmt(line.total_budget) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </div>
    </div>

    <!-- ── Create Budget Modal ──────────────────────────────────────────────── -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-all duration-200"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-all duration-150"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" @click="showModal=false" />
          <div class="relative w-full max-w-3xl rounded-2xl shadow-2xl overflow-hidden border"
               :class="app.darkMode ? 'bg-slate-900 border-slate-700/60' : 'bg-white border-slate-200'">
            <!-- Header -->
            <div class="px-6 py-4 border-b flex items-center justify-between"
                 :class="app.darkMode ? 'border-slate-800/60' : 'border-slate-100'">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-lg flex items-center justify-center"
                     :class="app.darkMode ? 'bg-rose-500/15 border border-rose-500/25' : 'bg-rose-100 border border-rose-200'">
                  <Target class="w-4 h-4" :class="app.darkMode ? 'text-rose-400' : 'text-rose-600'" />
                </div>
                <div>
                  <h3 class="text-sm font-semibold" :class="app.darkMode ? 'text-slate-100' : 'text-slate-800'">New Budget</h3>
                  <p class="text-[11px]" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Create budget with per-month allocation</p>
                </div>
              </div>
              <button @click="showModal=false"
                class="w-7 h-7 rounded-lg flex items-center justify-center transition-all"
                :class="app.darkMode ? 'text-slate-500 hover:text-slate-300 hover:bg-slate-800' : 'text-slate-400 hover:text-slate-600 hover:bg-slate-100'">
                <X class="w-4 h-4" />
              </button>
            </div>

            <!-- Body -->
            <div class="px-6 py-5 space-y-4 max-h-[70vh] overflow-y-auto">
              <!-- Name & Description -->
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-[11px] font-semibold uppercase tracking-wider mb-1.5"
                         :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Budget Name *</label>
                  <input v-model="form.name" type="text" placeholder="e.g. Budget 2025"
                    class="w-full h-9 px-3 rounded-lg border text-sm focus:outline-none transition-all"
                    :class="app.darkMode
                      ? 'bg-slate-800/60 border-slate-700/60 text-slate-100 placeholder-slate-600 focus:border-rose-500/60'
                      : 'bg-white border-slate-300 text-slate-800 placeholder-slate-400 focus:border-rose-400'" />
                </div>
                <div>
                  <label class="block text-[11px] font-semibold uppercase tracking-wider mb-1.5"
                         :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Status</label>
                  <select v-model="form.status"
                    class="w-full h-9 px-3 rounded-lg border text-sm focus:outline-none transition-all"
                    :class="app.darkMode
                      ? 'bg-slate-800/60 border-slate-700/60 text-slate-100 focus:border-rose-500/60'
                      : 'bg-white border-slate-300 text-slate-800 focus:border-rose-400'">
                    <option value="draft">Draft</option>
                    <option value="active">Active</option>
                    <option value="closed">Closed</option>
                  </select>
                </div>
              </div>
              <div>
                <label class="block text-[11px] font-semibold uppercase tracking-wider mb-1.5"
                       :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Description</label>
                <input v-model="form.description" type="text" placeholder="Optional description..."
                  class="w-full h-9 px-3 rounded-lg border text-sm focus:outline-none transition-all"
                  :class="app.darkMode
                    ? 'bg-slate-800/60 border-slate-700/60 text-slate-100 placeholder-slate-600 focus:border-rose-500/60'
                    : 'bg-white border-slate-300 text-slate-800 placeholder-slate-400 focus:border-rose-400'" />
              </div>

              <!-- Budget Lines -->
              <div>
                <div class="flex items-center justify-between mb-2">
                  <label class="text-[11px] font-semibold uppercase tracking-wider"
                         :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Budget Lines (per month)</label>
                  <button @click="addFormLine"
                    class="h-6 px-2.5 rounded text-[11px] font-medium inline-flex items-center gap-1 transition-all"
                    :class="app.darkMode ? 'bg-slate-800 text-slate-300 hover:bg-slate-700' : 'bg-slate-100 text-slate-600 hover:bg-slate-200'">
                    <Plus class="w-3 h-3" />
                    Add Line
                  </button>
                </div>
                <div class="overflow-x-auto rounded-lg border"
                     :class="app.darkMode ? 'border-slate-800/60' : 'border-slate-200'">
                  <table class="w-full text-xs">
                    <thead>
                      <tr :class="app.darkMode ? 'bg-slate-800/60 border-b border-slate-800' : 'bg-slate-50 border-b border-slate-200'">
                        <th class="text-left px-3 py-2 font-semibold uppercase tracking-wider"
                            :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'" style="min-width:160px">Account</th>
                        <th v-for="m in MONTHS" :key="m" class="text-center px-1 py-2 font-semibold uppercase tracking-wider w-16"
                            :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ m }}</th>
                        <th class="px-2 py-2 w-8"></th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(line, i) in form.lines" :key="i"
                          class="border-b"
                          :class="app.darkMode ? 'border-slate-800/30' : 'border-slate-100'">
                        <td class="px-2 py-1.5">
                          <select @change="onLineAccountChange(i, ($event.target as HTMLSelectElement).value)"
                            class="w-full h-7 px-2 rounded border text-[11px] focus:outline-none"
                            :class="app.darkMode
                              ? 'bg-slate-800 border-slate-700 text-slate-200'
                              : 'bg-white border-slate-200 text-slate-700'">
                            <option value="">-- Account --</option>
                            <option v-for="a in accounts" :key="a.id" :value="a.id">
                              {{ a.code }} — {{ a.name }}
                            </option>
                          </select>
                        </td>
                        <td v-for="(key, mi) in MONTH_KEYS" :key="key" class="px-1 py-1.5">
                          <input type="number" v-model.number="(line as any)[key]" min="0"
                            class="w-16 h-7 px-1.5 rounded border text-right text-[11px] font-mono focus:outline-none"
                            :class="app.darkMode
                              ? 'bg-slate-800 border-slate-700 text-slate-200'
                              : 'bg-white border-slate-200 text-slate-700'" />
                        </td>
                        <td class="px-2 py-1.5 text-center">
                          <button @click="removeFormLine(i)"
                            class="w-6 h-6 rounded flex items-center justify-center transition-colors"
                            :class="app.darkMode ? 'text-slate-600 hover:text-rose-400 hover:bg-rose-500/10' : 'text-slate-400 hover:text-rose-500 hover:bg-rose-50'">
                            <X class="w-3 h-3" />
                          </button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

            <!-- Footer -->
            <div class="px-6 py-4 border-t flex items-center justify-end gap-3"
                 :class="app.darkMode ? 'border-slate-800/60' : 'border-slate-100'">
              <button @click="showModal=false"
                class="h-9 px-4 rounded-lg border text-sm font-medium transition-all"
                :class="app.darkMode
                  ? 'border-slate-700 bg-slate-800/40 text-slate-400 hover:text-slate-200'
                  : 'border-slate-300 bg-white text-slate-500 hover:text-slate-700'">
                Cancel
              </button>
              <button @click="save" :disabled="saving"
                class="h-9 px-5 rounded-lg bg-rose-600 hover:bg-rose-500 disabled:opacity-50 text-white text-sm font-medium inline-flex items-center gap-2 transition-all shadow-lg shadow-rose-900/30">
                <Loader2 v-if="saving" class="w-3.5 h-3.5 animate-spin" />
                <CheckCircle v-else class="w-3.5 h-3.5" />
                Create Budget
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
