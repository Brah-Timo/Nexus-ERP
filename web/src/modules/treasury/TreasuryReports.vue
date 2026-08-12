<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Treasury Reports</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Cash flow, payments and reconciliation analytics</p>
      </div>
      <div class="flex items-center gap-3">
        <select v-model="year" @change="load" class="px-3 py-2 text-sm border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-gray-900 text-gray-700 dark:text-gray-200 outline-none focus:ring-2 focus:ring-indigo-500">
          <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
        </select>
        <button @click="load" class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors">
          <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          Refresh
        </button>
      </div>
    </div>

    <div v-if="loading" class="flex items-center justify-center py-20">
      <Loader2 class="w-10 h-10 text-indigo-500 animate-spin" />
    </div>

    <template v-else>
      <!-- Top KPIs -->
      <div class="grid grid-cols-2 xl:grid-cols-4 gap-4">
        <div v-for="kpi in topKpis" :key="kpi.label" class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-start gap-4">
          <div :class="kpi.iconBg" class="w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0">
            <component :is="kpi.icon" :class="kpi.iconColor" class="w-6 h-6" />
          </div>
          <div>
            <p class="text-xs font-medium text-gray-400 uppercase tracking-wide">{{ kpi.label }}</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(kpi.value) }}</p>
            <p class="text-xs text-gray-400">{{ kpi.sub }}</p>
          </div>
        </div>
      </div>

      <!-- Monthly Cash Flow Chart (bar-style) -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
        <h3 class="font-semibold text-gray-900 dark:text-white mb-5 flex items-center gap-2">
          <BarChart2 class="w-4 h-4 text-indigo-500" />
          Monthly Cash Flow — {{ year }}
        </h3>
        <div v-if="!cashFlow.length" class="flex items-center justify-center py-10 text-gray-400 text-sm">No data for {{ year }}</div>
        <div v-else class="overflow-x-auto">
          <div class="flex items-end gap-2 min-w-max" style="height:160px;">
            <div v-for="cf in cashFlow" :key="cf.month" class="flex flex-col items-center gap-1 w-16">
              <div class="w-full flex items-end justify-center gap-0.5" style="height:130px;">
                <div class="w-5 rounded-t-sm bg-emerald-500/80 dark:bg-emerald-600 transition-all"
                  :style="{ height: `${Math.max(3, (cf.inflow / maxFlow) * 120)}px` }"
                  :title="`Inflow: ${fmtCurrency(cf.inflow)}`"></div>
                <div class="w-5 rounded-t-sm bg-red-400/80 dark:bg-red-500 transition-all"
                  :style="{ height: `${Math.max(3, (cf.outflow / maxFlow) * 120)}px` }"
                  :title="`Outflow: ${fmtCurrency(cf.outflow)}`"></div>
              </div>
              <span class="text-xs text-gray-400">{{ cf.month.slice(5) }}</span>
            </div>
          </div>
          <div class="flex items-center gap-4 mt-3 text-xs text-gray-500">
            <span class="flex items-center gap-1"><span class="w-3 h-3 rounded-sm bg-emerald-500 inline-block"></span> Inflow</span>
            <span class="flex items-center gap-1"><span class="w-3 h-3 rounded-sm bg-red-400 inline-block"></span> Outflow</span>
          </div>
        </div>
      </div>

      <!-- Two column: Payment Summary + Cheque Stats -->
      <div class="grid grid-cols-1 xl:grid-cols-2 gap-4">

        <!-- Payment Summary -->
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
          <h3 class="font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
            <CreditCard class="w-4 h-4 text-indigo-500" />
            Payment Summary
          </h3>
          <div v-if="!paymentSummary.length" class="text-gray-400 text-sm text-center py-6">No payment data</div>
          <div v-else class="space-y-3">
            <div v-for="ps in paymentSummary" :key="ps.type" class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
              <div class="flex items-center gap-3">
                <div :class="ps.type === 'outgoing' ? 'bg-red-100 dark:bg-red-900/30' : 'bg-emerald-100 dark:bg-emerald-900/30'" class="w-8 h-8 rounded-lg flex items-center justify-center">
                  <ArrowUpRight v-if="ps.type === 'outgoing'" class="w-4 h-4 text-red-500" />
                  <ArrowDownLeft v-else class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
                </div>
                <div>
                  <p class="font-medium text-sm text-gray-900 dark:text-white capitalize">{{ ps.type }}</p>
                  <p class="text-xs text-gray-400">{{ ps.count }} payments</p>
                </div>
              </div>
              <span class="font-bold text-gray-900 dark:text-white">{{ fmtCurrency(ps.amount) }}</span>
            </div>
          </div>
        </div>

        <!-- Cheque Stats -->
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
          <h3 class="font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
            <FileText class="w-4 h-4 text-indigo-500" />
            Cheque Statistics
          </h3>
          <div v-if="!data?.cheques" class="text-gray-400 text-sm text-center py-6">No data</div>
          <div v-else class="grid grid-cols-3 gap-3">
            <div class="bg-amber-50 dark:bg-amber-900/20 rounded-xl p-3 text-center">
              <p class="text-2xl font-bold text-amber-600 dark:text-amber-400">{{ data.cheques.pending }}</p>
              <p class="text-xs text-gray-400 mt-0.5">Pending</p>
              <p class="text-xs font-medium text-amber-600 dark:text-amber-400">{{ fmtCurrencyShort(data.cheques.pending_amt) }}</p>
            </div>
            <div class="bg-emerald-50 dark:bg-emerald-900/20 rounded-xl p-3 text-center">
              <p class="text-2xl font-bold text-emerald-600 dark:text-emerald-400">{{ data.cheques.deposited }}</p>
              <p class="text-xs text-gray-400 mt-0.5">Deposited</p>
              <p class="text-xs font-medium text-emerald-600 dark:text-emerald-400">{{ fmtCurrencyShort(data.cheques.deposited_amt) }}</p>
            </div>
            <div class="bg-red-50 dark:bg-red-900/20 rounded-xl p-3 text-center">
              <p class="text-2xl font-bold text-red-500">{{ data.cheques.bounced }}</p>
              <p class="text-xs text-gray-400 mt-0.5">Bounced</p>
              <p class="text-xs font-medium text-red-500">—</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Monthly Cash Flow Table -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-800">
          <h3 class="font-semibold text-gray-900 dark:text-white">Monthly Breakdown</h3>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
                <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Month</th>
                <th class="text-right px-4 py-3 text-xs font-semibold text-emerald-600 uppercase tracking-wide">Inflow</th>
                <th class="text-right px-4 py-3 text-xs font-semibold text-red-500 uppercase tracking-wide">Outflow</th>
                <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Net</th>
                <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Margin</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
              <tr v-if="!cashFlow.length">
                <td colspan="5" class="text-center py-8 text-gray-400">No data for {{ year }}</td>
              </tr>
              <tr v-for="cf in cashFlow" :key="cf.month" class="hover:bg-gray-50 dark:hover:bg-gray-800/40">
                <td class="px-4 py-3 font-medium text-gray-900 dark:text-white">{{ monthName(cf.month) }}</td>
                <td class="px-4 py-3 text-right text-emerald-600 dark:text-emerald-400 font-medium">{{ fmtCurrency(cf.inflow) }}</td>
                <td class="px-4 py-3 text-right text-red-500 font-medium">{{ fmtCurrency(cf.outflow) }}</td>
                <td class="px-4 py-3 text-right font-bold" :class="cf.net >= 0 ? 'text-indigo-600 dark:text-indigo-400' : 'text-red-500'">{{ fmtCurrency(cf.net) }}</td>
                <td class="px-4 py-3 text-right text-gray-500 text-xs">
                  <span v-if="cf.inflow > 0">{{ Math.round((cf.net / cf.inflow) * 100) }}%</span>
                  <span v-else>—</span>
                </td>
              </tr>
              <!-- Totals row -->
              <tr v-if="cashFlow.length" class="bg-indigo-50 dark:bg-indigo-900/20 font-bold border-t-2 border-indigo-200 dark:border-indigo-800">
                <td class="px-4 py-3 text-indigo-700 dark:text-indigo-300">Total {{ year }}</td>
                <td class="px-4 py-3 text-right text-emerald-600 dark:text-emerald-400">{{ fmtCurrency(data?.total_inflow) }}</td>
                <td class="px-4 py-3 text-right text-red-500">{{ fmtCurrency(data?.total_outflow) }}</td>
                <td class="px-4 py-3 text-right" :class="(data?.net_flow ?? 0) >= 0 ? 'text-indigo-600 dark:text-indigo-400' : 'text-red-500'">{{ fmtCurrency(data?.net_flow) }}</td>
                <td class="px-4 py-3 text-right text-gray-500 text-xs">
                  <span v-if="data?.total_inflow > 0">{{ Math.round((data.net_flow / data.total_inflow) * 100) }}%</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Receipts Summary -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
        <h3 class="font-semibold text-gray-900 dark:text-white mb-3 flex items-center gap-2">
          <ArrowDownToLine class="w-4 h-4 text-indigo-500" />
          Receipts Summary — {{ year }}
        </h3>
        <div v-if="data?.receipts" class="grid grid-cols-2 sm:grid-cols-3 gap-4">
          <div class="bg-indigo-50 dark:bg-indigo-900/20 rounded-xl p-4 text-center">
            <p class="text-3xl font-bold text-indigo-600 dark:text-indigo-400">{{ data.receipts.count }}</p>
            <p class="text-sm text-gray-500 mt-1">Confirmed Receipts</p>
          </div>
          <div class="bg-emerald-50 dark:bg-emerald-900/20 rounded-xl p-4 text-center col-span-2 sm:col-span-1">
            <p class="text-2xl font-bold text-emerald-600 dark:text-emerald-400">{{ fmtCurrencyShort(data.receipts.amount) }}</p>
            <p class="text-sm text-gray-500 mt-1">Total Received</p>
          </div>
        </div>
        <div v-else class="text-gray-400 text-sm text-center py-4">No receipt data</div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  RefreshCw, Loader2, BarChart2, CreditCard, FileText,
  ArrowUpRight, ArrowDownLeft, ArrowDownToLine,
  TrendingUp, TrendingDown, Landmark, DollarSign
} from 'lucide-vue-next'
import { treasuryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()
const loading = ref(true)
const data = ref<any>(null)
const currentYear = new Date().getFullYear()
const year = ref(String(currentYear))
const years = Array.from({ length: 5 }, (_, i) => String(currentYear - i))

const cashFlow = computed(() => data.value?.cash_flow || [])
const paymentSummary = computed(() => data.value?.payment_summary || [])
const maxFlow = computed(() => Math.max(...cashFlow.value.flatMap((cf: any) => [cf.inflow, cf.outflow]), 1))

const topKpis = computed(() => [
  {
    label: 'Total Inflow',
    value: data.value?.total_inflow ?? 0,
    sub: `FY ${year.value}`,
    icon: TrendingUp,
    iconBg: 'bg-emerald-100 dark:bg-emerald-900/30',
    iconColor: 'text-emerald-600 dark:text-emerald-400'
  },
  {
    label: 'Total Outflow',
    value: data.value?.total_outflow ?? 0,
    sub: `FY ${year.value}`,
    icon: TrendingDown,
    iconBg: 'bg-red-100 dark:bg-red-900/30',
    iconColor: 'text-red-500'
  },
  {
    label: 'Net Flow',
    value: data.value?.net_flow ?? 0,
    sub: 'Inflow - Outflow',
    icon: Landmark,
    iconBg: 'bg-indigo-100 dark:bg-indigo-900/30',
    iconColor: 'text-indigo-600 dark:text-indigo-400'
  },
  {
    label: 'Receipts Total',
    value: data.value?.receipts?.amount ?? 0,
    sub: `${data.value?.receipts?.count ?? 0} confirmed`,
    icon: DollarSign,
    iconBg: 'bg-violet-100 dark:bg-violet-900/30',
    iconColor: 'text-violet-600 dark:text-violet-400'
  },
])

function fmtCurrency(n?: number) {
  if (!n && n !== 0) return '—'
  return new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 2 }).format(n) + ' DZD'
}
function fmtCurrencyShort(n?: number) {
  if (!n) return '0 DZD'
  if (n >= 1e9) return (n / 1e9).toFixed(2) + 'B DZD'
  if (n >= 1e6) return (n / 1e6).toFixed(2) + 'M DZD'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K DZD'
  return n.toFixed(2) + ' DZD'
}
function monthName(m: string) {
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']
  const idx = parseInt(m.slice(5)) - 1
  return months[idx] || m
}

async function load() {
  loading.value = true
  try {
    const res = await treasuryAPI.getTreasuryReport({ year: year.value })
    data.value = res.data
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load treasury report', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
