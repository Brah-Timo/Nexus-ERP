<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  BarChart3, FileText, Scale, TrendingUp, TrendingDown,
  RefreshCw, Loader2, Printer, Calendar, AlertCircle,
  CheckCircle, Download, ArrowRight, Minus, DollarSign,
  ChevronRight, BookOpen, Building2
} from 'lucide-vue-next'
import { accountingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ─────────────────────────────────────────────────────────────────────
interface TrialBalanceLine {
  code: string
  name: string
  type: string
  total_debit: number
  total_credit: number
  balance: number
}

interface BalanceSheetSection {
  name: string
  accounts: Array<{ code: string; name: string; balance: number }>
  total: number
}

// ─── State ─────────────────────────────────────────────────────────────────────
const activeTab = ref<'trial' | 'balance' | 'income'>('trial')
const loadingTrial   = ref(false)
const loadingBalance = ref(false)
const loadingIncome  = ref(false)

const trialData   = ref<{ date: string; lines: TrialBalanceLine[] } | null>(null)
const balanceData = ref<{ date: string; assets: BalanceSheetSection; liabilities: BalanceSheetSection; equity: BalanceSheetSection } | null>(null)
const incomeData  = ref<{ period: string; revenues: any[]; expenses: any[]; total_revenue: number; total_expense: number; net_profit: number } | null>(null)

// ─── Computed ──────────────────────────────────────────────────────────────────
const trialTotals = computed(() => {
  if (!trialData.value?.lines) return { debit: 0, credit: 0, balance: 0 }
  const debit   = trialData.value.lines.reduce((s, l) => s + (l.total_debit ?? 0), 0)
  const credit  = trialData.value.lines.reduce((s, l) => s + (l.total_credit ?? 0), 0)
  const balance = trialData.value.lines.reduce((s, l) => s + (l.balance ?? 0), 0)
  return { debit, credit, balance }
})

const trialBalanced = computed(() =>
  Math.abs(trialTotals.value.debit - trialTotals.value.credit) < 0.01
)

const balanceBalanced = computed(() => {
  if (!balanceData.value) return false
  const assets = balanceData.value.assets?.total ?? 0
  const liab   = balanceData.value.liabilities?.total ?? 0
  const equity = balanceData.value.equity?.total ?? 0
  return Math.abs(assets - (liab + equity)) < 0.01
})

const netProfitPositive = computed(() => (incomeData.value?.net_profit ?? 0) >= 0)

const typeConfig: Record<string, string> = {
  asset: 'text-blue-400', liability: 'text-rose-400',
  equity: 'text-violet-400', revenue: 'text-emerald-400', expense: 'text-amber-400'
}

// ─── Methods ───────────────────────────────────────────────────────────────────
async function loadTrialBalance() {
  loadingTrial.value = true
  try {
    const res = await accountingAPI.getTrialBalance()
    trialData.value = res.data
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Failed to load trial balance', 'error')
  } finally {
    loadingTrial.value = false
  }
}

async function loadBalanceSheet() {
  loadingBalance.value = true
  try {
    const res = await accountingAPI.getBalanceSheet()
    balanceData.value = res.data
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Failed to load balance sheet', 'error')
  } finally {
    loadingBalance.value = false
  }
}

async function loadIncomeStatement() {
  loadingIncome.value = true
  try {
    const res = await accountingAPI.getIncomeStatement()
    incomeData.value = res.data
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Failed to load income statement', 'error')
  } finally {
    loadingIncome.value = false
  }
}

async function loadAll() {
  await Promise.all([loadTrialBalance(), loadBalanceSheet(), loadIncomeStatement()])
}

function refreshTab() {
  if (activeTab.value === 'trial') loadTrialBalance()
  else if (activeTab.value === 'balance') loadBalanceSheet()
  else loadIncomeStatement()
}

function printReport() {
  window.print()
}

function fmtAmt(v: number) {
  if (v == null) return '0.00'
  return new Intl.NumberFormat('en-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)
}

function fmtAmtK(v: number) {
  if (Math.abs(v) >= 1_000_000) return (v / 1_000_000).toFixed(1) + 'M'
  if (Math.abs(v) >= 1_000) return (v / 1_000).toFixed(0) + 'K'
  return v?.toFixed(0) ?? '0'
}

onMounted(loadAll)
</script>

<template>
  <div class="flex flex-col h-full bg-slate-950 text-slate-100 print:bg-white print:text-black">

    <!-- ── Header ─────────────────────────────────────────────────────────── -->
    <div class="border-b border-slate-800/60 bg-slate-900/50 backdrop-blur-sm px-6 py-4 flex-shrink-0 print:hidden">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-lg bg-indigo-500/15 border border-indigo-500/25 flex items-center justify-center">
            <BarChart3 class="w-4.5 h-4.5 text-indigo-400" />
          </div>
          <div>
            <h1 class="text-[15px] font-semibold text-slate-100">Financial Reports</h1>
            <p class="text-[11px] text-slate-500">SCF — Accounting & Financial Statements</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="refreshTab" :disabled="loadingTrial||loadingBalance||loadingIncome"
            class="h-8 px-3 rounded-lg border border-slate-700/60 bg-slate-800/60 text-slate-400 hover:text-slate-200 text-xs font-medium inline-flex items-center gap-1.5 transition-all disabled:opacity-50">
            <RefreshCw class="w-3.5 h-3.5" :class="(loadingTrial||loadingBalance||loadingIncome) ? 'animate-spin' : ''" />
            Refresh
          </button>
          <button @click="printReport"
            class="h-8 px-3 rounded-lg border border-slate-700/60 bg-slate-800/60 text-slate-400 hover:text-slate-200 text-xs font-medium inline-flex items-center gap-1.5 transition-all">
            <Printer class="w-3.5 h-3.5" />
            Print
          </button>
        </div>
      </div>
    </div>

    <!-- ── Tab nav ────────────────────────────────────────────────────────── -->
    <div class="px-6 pt-4 pb-0 flex-shrink-0 print:hidden">
      <div class="flex items-end gap-1 border-b border-slate-800/60">
        <button @click="activeTab='trial'"
          :class="['px-4 py-2.5 text-sm font-medium inline-flex items-center gap-2 border-b-2 transition-all -mb-px', activeTab==='trial' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-slate-500 hover:text-slate-300']">
          <Scale class="w-3.5 h-3.5" />
          Trial Balance
        </button>
        <button @click="activeTab='balance'"
          :class="['px-4 py-2.5 text-sm font-medium inline-flex items-center gap-2 border-b-2 transition-all -mb-px', activeTab==='balance' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-slate-500 hover:text-slate-300']">
          <Building2 class="w-3.5 h-3.5" />
          Balance Sheet
        </button>
        <button @click="activeTab='income'"
          :class="['px-4 py-2.5 text-sm font-medium inline-flex items-center gap-2 border-b-2 transition-all -mb-px', activeTab==='income' ? 'border-indigo-500 text-indigo-400' : 'border-transparent text-slate-500 hover:text-slate-300']">
          <TrendingUp class="w-3.5 h-3.5" />
          Income Statement
        </button>
      </div>
    </div>

    <!-- ── Content ────────────────────────────────────────────────────────── -->
    <div class="flex-1 overflow-hidden">

      <!-- TRIAL BALANCE ─────────────────────────────────────────────────── -->
      <div v-if="activeTab==='trial'" class="h-full flex flex-col">
        <div v-if="loadingTrial" class="flex items-center justify-center flex-1">
          <Loader2 class="w-7 h-7 text-indigo-400 animate-spin" />
        </div>
        <template v-else-if="trialData">
          <!-- Report header -->
          <div class="px-6 py-3 flex-shrink-0 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <h2 class="text-sm font-semibold text-slate-300">Trial Balance</h2>
              <span class="text-xs text-slate-600">As of {{ trialData.date }}</span>
              <span :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold border', trialBalanced ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-rose-500/10 text-rose-400 border-rose-500/20']">
                <CheckCircle v-if="trialBalanced" class="w-3 h-3" />
                <AlertCircle v-else class="w-3 h-3" />
                {{ trialBalanced ? 'Balanced' : 'Unbalanced' }}
              </span>
            </div>
          </div>
          <div class="flex-1 overflow-auto px-6 pb-6">
            <div class="rounded-xl border border-slate-800/60 overflow-hidden bg-slate-900/40">
              <table class="w-full text-sm border-collapse">
                <thead class="sticky top-0 z-10">
                  <tr class="bg-slate-900/90 backdrop-blur border-b border-slate-800/60">
                    <th class="text-left px-4 py-3 w-24 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Code</th>
                    <th class="text-left px-4 py-3 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Account Name</th>
                    <th class="text-left px-4 py-3 w-24 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Type</th>
                    <th class="text-right px-4 py-3 w-36 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Debit</th>
                    <th class="text-right px-4 py-3 w-36 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Credit</th>
                    <th class="text-right px-4 py-3 w-36 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Balance</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-if="!trialData.lines?.length">
                    <td colspan="6" class="py-16 text-center text-slate-600 text-sm">No data available</td>
                  </tr>
                  <tr v-for="(line, i) in trialData.lines" :key="line.code"
                    :class="['border-b border-slate-800/30 hover:bg-slate-800/20 transition-colors', i%2===0?'':'bg-slate-900/20']">
                    <td class="px-4 py-2.5">
                      <span class="font-mono text-[12px] font-semibold text-indigo-300">{{ line.code }}</span>
                    </td>
                    <td class="px-4 py-2.5">
                      <span class="text-[13px] text-slate-200">{{ line.name }}</span>
                    </td>
                    <td class="px-4 py-2.5">
                      <span :class="['text-[11px] font-semibold capitalize', typeConfig[line.type] ?? 'text-slate-400']">{{ line.type }}</span>
                    </td>
                    <td class="px-4 py-2.5 text-right">
                      <span v-if="line.total_debit" class="font-mono text-[12px] text-emerald-400">{{ fmtAmt(line.total_debit) }}</span>
                      <span v-else class="text-slate-700">—</span>
                    </td>
                    <td class="px-4 py-2.5 text-right">
                      <span v-if="line.total_credit" class="font-mono text-[12px] text-rose-400">{{ fmtAmt(line.total_credit) }}</span>
                      <span v-else class="text-slate-700">—</span>
                    </td>
                    <td class="px-4 py-2.5 text-right">
                      <span :class="['font-mono text-[12px] font-semibold', (line.balance??0) >= 0 ? 'text-slate-200' : 'text-rose-400']">
                        {{ fmtAmt(line.balance) }}
                      </span>
                    </td>
                  </tr>
                </tbody>
                <tfoot>
                  <tr class="border-t-2 border-slate-700/60 bg-slate-900/80">
                    <td colspan="3" class="px-4 py-3 text-[12px] font-bold text-slate-300 uppercase tracking-wider">Totals</td>
                    <td class="px-4 py-3 text-right">
                      <span class="font-mono text-sm font-bold text-emerald-400">{{ fmtAmt(trialTotals.debit) }}</span>
                    </td>
                    <td class="px-4 py-3 text-right">
                      <span class="font-mono text-sm font-bold text-rose-400">{{ fmtAmt(trialTotals.credit) }}</span>
                    </td>
                    <td class="px-4 py-3 text-right">
                      <span :class="['font-mono text-sm font-bold', trialBalanced ? 'text-emerald-400' : 'text-rose-400']">{{ fmtAmt(trialTotals.balance) }}</span>
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>
        </template>
      </div>

      <!-- BALANCE SHEET ─────────────────────────────────────────────────── -->
      <div v-else-if="activeTab==='balance'" class="h-full flex flex-col">
        <div v-if="loadingBalance" class="flex items-center justify-center flex-1">
          <Loader2 class="w-7 h-7 text-indigo-400 animate-spin" />
        </div>
        <template v-else-if="balanceData">
          <div class="px-6 py-3 flex-shrink-0 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <h2 class="text-sm font-semibold text-slate-300">Balance Sheet</h2>
              <span class="text-xs text-slate-600">As of {{ balanceData.date }}</span>
              <span :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold border', balanceBalanced ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-rose-500/10 text-rose-400 border-rose-500/20']">
                <CheckCircle v-if="balanceBalanced" class="w-3 h-3" />
                <AlertCircle v-else class="w-3 h-3" />
                {{ balanceBalanced ? 'Assets = Liabilities + Equity' : 'Balance mismatch' }}
              </span>
            </div>
          </div>
          <div class="flex-1 overflow-auto px-6 pb-6">
            <div class="grid grid-cols-2 gap-4">
              <!-- Assets -->
              <div class="rounded-xl border border-blue-500/20 overflow-hidden bg-slate-900/40">
                <div class="px-4 py-3 bg-blue-500/10 border-b border-blue-500/15 flex items-center justify-between">
                  <h3 class="text-sm font-semibold text-blue-300">Assets</h3>
                  <span class="font-mono text-sm font-bold text-blue-400">{{ fmtAmt(balanceData.assets?.total ?? 0) }}</span>
                </div>
                <table class="w-full text-sm">
                  <tbody>
                    <tr v-for="(acc, i) in (balanceData.assets?.accounts ?? [])" :key="acc.code"
                      :class="['border-b border-slate-800/20 hover:bg-slate-800/20', i%2===0?'':'bg-slate-900/20']">
                      <td class="px-4 py-2 text-left">
                        <div class="flex items-center gap-2">
                          <span class="font-mono text-[11px] text-blue-300/60">{{ acc.code }}</span>
                          <span class="text-[12px] text-slate-300">{{ acc.name }}</span>
                        </div>
                      </td>
                      <td class="px-4 py-2 text-right font-mono text-[12px]"
                        :class="(acc.balance??0) >= 0 ? 'text-slate-300' : 'text-rose-400'">
                        {{ fmtAmt(acc.balance) }}
                      </td>
                    </tr>
                    <tr v-if="!(balanceData.assets?.accounts?.length)">
                      <td colspan="2" class="py-8 text-center text-slate-600 text-xs">No asset accounts</td>
                    </tr>
                  </tbody>
                  <tfoot>
                    <tr class="border-t-2 border-blue-500/30 bg-blue-500/5">
                      <td class="px-4 py-2.5 text-[11px] font-bold text-blue-300 uppercase tracking-wider">Total Assets</td>
                      <td class="px-4 py-2.5 text-right font-mono text-sm font-bold text-blue-400">{{ fmtAmt(balanceData.assets?.total ?? 0) }}</td>
                    </tr>
                  </tfoot>
                </table>
              </div>

              <!-- Liabilities + Equity -->
              <div class="space-y-4">
                <!-- Liabilities -->
                <div class="rounded-xl border border-rose-500/20 overflow-hidden bg-slate-900/40">
                  <div class="px-4 py-3 bg-rose-500/10 border-b border-rose-500/15 flex items-center justify-between">
                    <h3 class="text-sm font-semibold text-rose-300">Liabilities</h3>
                    <span class="font-mono text-sm font-bold text-rose-400">{{ fmtAmt(balanceData.liabilities?.total ?? 0) }}</span>
                  </div>
                  <table class="w-full text-sm">
                    <tbody>
                      <tr v-for="(acc, i) in (balanceData.liabilities?.accounts ?? [])" :key="acc.code"
                        :class="['border-b border-slate-800/20 hover:bg-slate-800/20', i%2===0?'':'bg-slate-900/20']">
                        <td class="px-4 py-2">
                          <div class="flex items-center gap-2">
                            <span class="font-mono text-[11px] text-rose-300/60">{{ acc.code }}</span>
                            <span class="text-[12px] text-slate-300">{{ acc.name }}</span>
                          </div>
                        </td>
                        <td class="px-4 py-2 text-right font-mono text-[12px] text-slate-300">{{ fmtAmt(acc.balance) }}</td>
                      </tr>
                      <tr v-if="!(balanceData.liabilities?.accounts?.length)">
                        <td colspan="2" class="py-6 text-center text-slate-600 text-xs">No liability accounts</td>
                      </tr>
                    </tbody>
                    <tfoot>
                      <tr class="border-t-2 border-rose-500/30 bg-rose-500/5">
                        <td class="px-4 py-2 text-[11px] font-bold text-rose-300 uppercase tracking-wider">Total Liabilities</td>
                        <td class="px-4 py-2 text-right font-mono text-sm font-bold text-rose-400">{{ fmtAmt(balanceData.liabilities?.total ?? 0) }}</td>
                      </tr>
                    </tfoot>
                  </table>
                </div>

                <!-- Equity -->
                <div class="rounded-xl border border-violet-500/20 overflow-hidden bg-slate-900/40">
                  <div class="px-4 py-3 bg-violet-500/10 border-b border-violet-500/15 flex items-center justify-between">
                    <h3 class="text-sm font-semibold text-violet-300">Equity</h3>
                    <span class="font-mono text-sm font-bold text-violet-400">{{ fmtAmt(balanceData.equity?.total ?? 0) }}</span>
                  </div>
                  <table class="w-full text-sm">
                    <tbody>
                      <tr v-for="(acc, i) in (balanceData.equity?.accounts ?? [])" :key="acc.code"
                        :class="['border-b border-slate-800/20 hover:bg-slate-800/20', i%2===0?'':'bg-slate-900/20']">
                        <td class="px-4 py-2">
                          <div class="flex items-center gap-2">
                            <span class="font-mono text-[11px] text-violet-300/60">{{ acc.code }}</span>
                            <span class="text-[12px] text-slate-300">{{ acc.name }}</span>
                          </div>
                        </td>
                        <td class="px-4 py-2 text-right font-mono text-[12px] text-slate-300">{{ fmtAmt(acc.balance) }}</td>
                      </tr>
                      <tr v-if="!(balanceData.equity?.accounts?.length)">
                        <td colspan="2" class="py-6 text-center text-slate-600 text-xs">No equity accounts</td>
                      </tr>
                    </tbody>
                    <tfoot>
                      <tr class="border-t-2 border-violet-500/30 bg-violet-500/5">
                        <td class="px-4 py-2 text-[11px] font-bold text-violet-300 uppercase tracking-wider">Total Equity</td>
                        <td class="px-4 py-2 text-right font-mono text-sm font-bold text-violet-400">{{ fmtAmt(balanceData.equity?.total ?? 0) }}</td>
                      </tr>
                    </tfoot>
                  </table>
                </div>
              </div>
            </div>

            <!-- Balance check -->
            <div class="mt-4 rounded-xl px-5 py-3 flex items-center gap-3"
              :class="balanceBalanced ? 'bg-emerald-500/10 border border-emerald-500/20' : 'bg-rose-500/10 border border-rose-500/20'">
              <CheckCircle v-if="balanceBalanced" class="w-4 h-4 text-emerald-400" />
              <AlertCircle v-else class="w-4 h-4 text-rose-400" />
              <div class="flex-1 text-sm" :class="balanceBalanced ? 'text-emerald-400' : 'text-rose-400'">
                Assets ({{ fmtAmt(balanceData.assets?.total??0) }}) = Liabilities ({{ fmtAmt(balanceData.liabilities?.total??0) }}) + Equity ({{ fmtAmt(balanceData.equity?.total??0) }})
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- INCOME STATEMENT ──────────────────────────────────────────────── -->
      <div v-else-if="activeTab==='income'" class="h-full flex flex-col">
        <div v-if="loadingIncome" class="flex items-center justify-center flex-1">
          <Loader2 class="w-7 h-7 text-indigo-400 animate-spin" />
        </div>
        <template v-else-if="incomeData">
          <div class="px-6 py-3 flex-shrink-0 flex items-center gap-3">
            <h2 class="text-sm font-semibold text-slate-300">Income Statement</h2>
            <span class="text-xs text-slate-600">FY {{ incomeData.period }}</span>
          </div>
          <div class="flex-1 overflow-auto px-6 pb-6 space-y-4">
            <!-- Net Profit banner -->
            <div :class="['rounded-xl px-6 py-4 flex items-center gap-4', netProfitPositive ? 'bg-emerald-500/10 border border-emerald-500/20' : 'bg-rose-500/10 border border-rose-500/20']">
              <div class="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0"
                :class="netProfitPositive ? 'bg-emerald-500/20' : 'bg-rose-500/20'">
                <TrendingUp v-if="netProfitPositive" :class="netProfitPositive ? 'w-5 h-5 text-emerald-400' : 'w-5 h-5 text-rose-400'" />
                <TrendingDown v-else class="w-5 h-5 text-rose-400" />
              </div>
              <div>
                <div class="text-[11px] font-semibold uppercase tracking-wider"
                  :class="netProfitPositive ? 'text-emerald-500/70' : 'text-rose-500/70'">
                  {{ netProfitPositive ? 'Net Profit' : 'Net Loss' }}
                </div>
                <div class="text-2xl font-mono font-bold"
                  :class="netProfitPositive ? 'text-emerald-400' : 'text-rose-400'">
                  {{ fmtAmt(Math.abs(incomeData.net_profit)) }} DZD
                </div>
              </div>
              <div class="ml-auto flex items-center gap-8">
                <div class="text-right">
                  <div class="text-[10px] text-emerald-500/60 font-semibold uppercase tracking-wider mb-0.5">Revenue</div>
                  <div class="text-sm font-mono font-semibold text-emerald-400">{{ fmtAmt(incomeData.total_revenue) }}</div>
                </div>
                <div class="text-[10px] text-slate-700 font-bold">–</div>
                <div class="text-right">
                  <div class="text-[10px] text-amber-500/60 font-semibold uppercase tracking-wider mb-0.5">Expenses</div>
                  <div class="text-sm font-mono font-semibold text-amber-400">{{ fmtAmt(incomeData.total_expense) }}</div>
                </div>
              </div>
            </div>

            <!-- Two-column: Revenue vs Expenses -->
            <div class="grid grid-cols-2 gap-4">
              <!-- Revenue -->
              <div class="rounded-xl border border-emerald-500/20 overflow-hidden bg-slate-900/40">
                <div class="px-4 py-3 bg-emerald-500/10 border-b border-emerald-500/15 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <TrendingUp class="w-3.5 h-3.5 text-emerald-400" />
                    <h3 class="text-sm font-semibold text-emerald-300">Revenue</h3>
                  </div>
                  <span class="font-mono text-sm font-bold text-emerald-400">{{ fmtAmt(incomeData.total_revenue) }}</span>
                </div>
                <table class="w-full text-sm">
                  <tbody>
                    <tr v-for="(rev, i) in (incomeData.revenues ?? [])" :key="rev.code"
                      :class="['border-b border-slate-800/20 hover:bg-slate-800/20', i%2===0?'':'bg-slate-900/20']">
                      <td class="px-4 py-2">
                        <div class="flex items-center gap-2">
                          <span class="font-mono text-[11px] text-emerald-300/50">{{ rev.code }}</span>
                          <span class="text-[12px] text-slate-300">{{ rev.name }}</span>
                        </div>
                      </td>
                      <td class="px-4 py-2 text-right font-mono text-[12px] text-emerald-400">{{ fmtAmt(rev.balance) }}</td>
                    </tr>
                    <tr v-if="!(incomeData.revenues?.length)">
                      <td colspan="2" class="py-8 text-center text-slate-600 text-xs">No revenue accounts</td>
                    </tr>
                  </tbody>
                  <tfoot>
                    <tr class="border-t-2 border-emerald-500/30 bg-emerald-500/5">
                      <td class="px-4 py-2 text-[11px] font-bold text-emerald-300 uppercase tracking-wider">Total Revenue</td>
                      <td class="px-4 py-2 text-right font-mono text-sm font-bold text-emerald-400">{{ fmtAmt(incomeData.total_revenue) }}</td>
                    </tr>
                  </tfoot>
                </table>
              </div>

              <!-- Expenses -->
              <div class="rounded-xl border border-amber-500/20 overflow-hidden bg-slate-900/40">
                <div class="px-4 py-3 bg-amber-500/10 border-b border-amber-500/15 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <TrendingDown class="w-3.5 h-3.5 text-amber-400" />
                    <h3 class="text-sm font-semibold text-amber-300">Expenses</h3>
                  </div>
                  <span class="font-mono text-sm font-bold text-amber-400">{{ fmtAmt(incomeData.total_expense) }}</span>
                </div>
                <table class="w-full text-sm">
                  <tbody>
                    <tr v-for="(exp, i) in (incomeData.expenses ?? [])" :key="exp.code"
                      :class="['border-b border-slate-800/20 hover:bg-slate-800/20', i%2===0?'':'bg-slate-900/20']">
                      <td class="px-4 py-2">
                        <div class="flex items-center gap-2">
                          <span class="font-mono text-[11px] text-amber-300/50">{{ exp.code }}</span>
                          <span class="text-[12px] text-slate-300">{{ exp.name }}</span>
                        </div>
                      </td>
                      <td class="px-4 py-2 text-right font-mono text-[12px] text-amber-400">{{ fmtAmt(exp.balance) }}</td>
                    </tr>
                    <tr v-if="!(incomeData.expenses?.length)">
                      <td colspan="2" class="py-8 text-center text-slate-600 text-xs">No expense accounts</td>
                    </tr>
                  </tbody>
                  <tfoot>
                    <tr class="border-t-2 border-amber-500/30 bg-amber-500/5">
                      <td class="px-4 py-2 text-[11px] font-bold text-amber-300 uppercase tracking-wider">Total Expenses</td>
                      <td class="px-4 py-2 text-right font-mono text-sm font-bold text-amber-400">{{ fmtAmt(incomeData.total_expense) }}</td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </div>
          </div>
        </template>
      </div>

    </div>
  </div>
</template>
