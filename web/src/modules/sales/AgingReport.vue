<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { salesAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  RefreshCw, BarChart3, AlertTriangle, CheckCircle,
  TrendingUp, Clock, ChevronDown, ChevronUp,
  Search, Phone, Mail, Download
} from 'lucide-vue-next'

// ─── Types ───────────────────────────────────────────────────────────────────

interface CustomerAging {
  customer_id: string
  customer_name: string
  phone: string
  email: string
  invoice_count: number
  current_amount: number     // not yet due
  days_1_to_30: number
  days_31_to_60: number
  days_61_to_90: number
  days_over_90: number
  total_outstanding: number
}

// ─── State ───────────────────────────────────────────────────────────────────

const app = useAppStore()
const aging = ref<CustomerAging[]>([])
const loading = ref(true)
const search = ref('')
const sortBy = ref<keyof CustomerAging>('total_outstanding')
const sortDir = ref<'asc' | 'desc'>('desc')

// ─── Computed ────────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = [...aging.value]
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(r => r.customer_name.toLowerCase().includes(q) || r.phone?.includes(q) || r.email?.toLowerCase().includes(q))
  }
  list.sort((a, b) => {
    const av = a[sortBy.value] as number | string
    const bv = b[sortBy.value] as number | string
    if (typeof av === 'number' && typeof bv === 'number') {
      return sortDir.value === 'asc' ? av - bv : bv - av
    }
    return sortDir.value === 'asc' ? String(av).localeCompare(String(bv)) : String(bv).localeCompare(String(av))
  })
  return list
})

const totals = computed(() => ({
  invoice_count:    filtered.value.reduce((s, r) => s + r.invoice_count, 0),
  current_amount:   filtered.value.reduce((s, r) => s + r.current_amount, 0),
  days_1_to_30:     filtered.value.reduce((s, r) => s + r.days_1_to_30, 0),
  days_31_to_60:    filtered.value.reduce((s, r) => s + r.days_31_to_60, 0),
  days_61_to_90:    filtered.value.reduce((s, r) => s + r.days_61_to_90, 0),
  days_over_90:     filtered.value.reduce((s, r) => s + r.days_over_90, 0),
  total_outstanding:filtered.value.reduce((s, r) => s + r.total_outstanding, 0),
}))

const maxTotal = computed(() => Math.max(...filtered.value.map(r => r.total_outstanding), 1))

const criticalCount = computed(() => filtered.value.filter(r => r.days_over_90 > 0).length)
const overdueCount  = computed(() => filtered.value.filter(r => r.days_1_to_30 + r.days_31_to_60 + r.days_61_to_90 + r.days_over_90 > 0).length)

// ─── Data ─────────────────────────────────────────────────────────────────────

async function loadData() {
  loading.value = true
  try {
    const res = await salesAPI.getAgingReport()
    aging.value = res.data || []
  } catch {
    app.addToast('Failed to load aging report', 'error')
  } finally {
    loading.value = false
  }
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

function fmtCurrency(n?: number) {
  if (n == null || n === 0) return '—'
  if (Math.abs(n) >= 1_000_000) return (n / 1_000_000).toFixed(2) + ' M DZD'
  if (Math.abs(n) >= 1_000) return (n / 1_000).toFixed(1) + ' k DZD'
  return n.toLocaleString('fr-DZ', { minimumFractionDigits: 2 }) + ' DZD'
}

function toggleSort(col: keyof CustomerAging) {
  if (sortBy.value === col) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortBy.value = col; sortDir.value = 'desc' }
}

// Aging bucket severity
function bucketColor(amount: number): string {
  if (amount === 0) return 'text-gray-300 dark:text-gray-700'
  return 'text-gray-900 dark:text-white font-medium tabular-nums'
}

function overdueSeverity(row: CustomerAging): string {
  if (row.days_over_90 > 0) return 'bg-red-50 dark:bg-red-900/10'
  if (row.days_61_to_90 > 0) return 'bg-orange-50 dark:bg-orange-900/10'
  if (row.days_31_to_60 > 0) return 'bg-amber-50 dark:bg-amber-900/10'
  return ''
}

// Export CSV
function exportCSV() {
  const headers = ['Customer', 'Phone', 'Email', 'Invoices', 'Current', '1-30d', '31-60d', '61-90d', '90+d', 'Total Outstanding']
  const rows = filtered.value.map(r => [
    r.customer_name, r.phone || '', r.email || '',
    r.invoice_count,
    r.current_amount, r.days_1_to_30, r.days_31_to_60, r.days_61_to_90, r.days_over_90,
    r.total_outstanding,
  ])
  const csv = [headers, ...rows].map(r => r.join(',')).join('\n')
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url; a.download = `aging-report-${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(loadData)
</script>

<template>
  <div class="space-y-5">

    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Customer Aging Report</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
          Outstanding receivables by aging bucket
          <span v-if="overdueCount > 0" class="ml-1 text-amber-600 dark:text-amber-400">· {{ overdueCount }} customers with overdue</span>
          <span v-if="criticalCount > 0" class="ml-1 text-red-600 dark:text-red-400">· {{ criticalCount }} critical (90+ days)</span>
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="exportCSV" class="inline-flex items-center gap-2 px-4 py-2 border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 text-sm font-medium rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
          <Download class="w-4 h-4" /> Export CSV
        </button>
        <button @click="loadData" :disabled="loading" class="p-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors disabled:opacity-50">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
      </div>
    </div>

    <!-- KPI Cards -->
    <div v-if="!loading && aging.length > 0" class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
      <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 p-3 shadow-sm">
        <p class="text-xs text-gray-500 dark:text-gray-400 font-medium">Current (not due)</p>
        <p class="text-sm font-bold text-gray-900 dark:text-white mt-1 truncate">{{ fmtCurrency(totals.current_amount) }}</p>
      </div>
      <div class="rounded-xl border border-amber-200 dark:border-amber-800 bg-amber-50 dark:bg-amber-900/10 p-3 shadow-sm">
        <p class="text-xs text-amber-600 dark:text-amber-400 font-medium">1 – 30 Days</p>
        <p class="text-sm font-bold text-amber-800 dark:text-amber-300 mt-1 truncate">{{ fmtCurrency(totals.days_1_to_30) }}</p>
      </div>
      <div class="rounded-xl border border-orange-200 dark:border-orange-800 bg-orange-50 dark:bg-orange-900/10 p-3 shadow-sm">
        <p class="text-xs text-orange-600 dark:text-orange-400 font-medium">31 – 60 Days</p>
        <p class="text-sm font-bold text-orange-800 dark:text-orange-300 mt-1 truncate">{{ fmtCurrency(totals.days_31_to_60) }}</p>
      </div>
      <div class="rounded-xl border border-red-300 dark:border-red-800 bg-red-50 dark:bg-red-900/10 p-3 shadow-sm">
        <p class="text-xs text-red-600 dark:text-red-400 font-medium">61 – 90 Days</p>
        <p class="text-sm font-bold text-red-800 dark:text-red-300 mt-1 truncate">{{ fmtCurrency(totals.days_61_to_90) }}</p>
      </div>
      <div class="rounded-xl border border-red-400 dark:border-red-700 bg-red-100 dark:bg-red-900/20 p-3 shadow-sm">
        <p class="text-xs text-red-700 dark:text-red-300 font-medium">Over 90 Days</p>
        <p class="text-sm font-bold text-red-900 dark:text-red-200 mt-1 truncate">{{ fmtCurrency(totals.days_over_90) }}</p>
      </div>
      <div class="rounded-xl border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 p-3 shadow-sm">
        <p class="text-xs text-blue-600 dark:text-blue-400 font-medium">Total Outstanding</p>
        <p class="text-sm font-bold text-blue-800 dark:text-blue-300 mt-1 truncate">{{ fmtCurrency(totals.total_outstanding) }}</p>
      </div>
    </div>

    <!-- Search -->
    <div class="relative max-w-xs">
      <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
      <input v-model="search" placeholder="Search customer…" class="w-full pl-9 pr-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 outline-none" />
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center h-40">
        <RefreshCw class="w-6 h-6 text-gray-400 animate-spin" />
      </div>
      <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center h-40 gap-2 text-gray-400 dark:text-gray-600">
        <CheckCircle class="w-8 h-8 text-green-400" />
        <p class="text-sm font-medium text-green-600 dark:text-green-400">No outstanding receivables</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <tr>
              <th @click="toggleSort('customer_name')" class="th cursor-pointer min-w-48">
                <div class="flex items-center gap-1">Customer <ChevronUp v-if="sortBy==='customer_name'&&sortDir==='asc'" class="w-3 h-3"/><ChevronDown v-else-if="sortBy==='customer_name'" class="w-3 h-3"/></div>
              </th>
              <th class="th">Contact</th>
              <th @click="toggleSort('invoice_count')" class="th text-right cursor-pointer">
                <div class="flex items-center justify-end gap-1">Inv. <ChevronUp v-if="sortBy==='invoice_count'&&sortDir==='asc'" class="w-3 h-3"/><ChevronDown v-else-if="sortBy==='invoice_count'" class="w-3 h-3"/></div>
              </th>
              <th @click="toggleSort('current_amount')" class="th text-right cursor-pointer whitespace-nowrap">
                <div class="flex items-center justify-end gap-1">Current <ChevronUp v-if="sortBy==='current_amount'&&sortDir==='asc'" class="w-3 h-3"/><ChevronDown v-else-if="sortBy==='current_amount'" class="w-3 h-3"/></div>
              </th>
              <th @click="toggleSort('days_1_to_30')" class="th text-right cursor-pointer whitespace-nowrap">
                <div class="flex items-center justify-end gap-1">1–30d <ChevronUp v-if="sortBy==='days_1_to_30'&&sortDir==='asc'" class="w-3 h-3"/><ChevronDown v-else-if="sortBy==='days_1_to_30'" class="w-3 h-3"/></div>
              </th>
              <th @click="toggleSort('days_31_to_60')" class="th text-right cursor-pointer whitespace-nowrap">
                <div class="flex items-center justify-end gap-1">31–60d <ChevronUp v-if="sortBy==='days_31_to_60'&&sortDir==='asc'" class="w-3 h-3"/><ChevronDown v-else-if="sortBy==='days_31_to_60'" class="w-3 h-3"/></div>
              </th>
              <th @click="toggleSort('days_61_to_90')" class="th text-right cursor-pointer whitespace-nowrap">
                <div class="flex items-center justify-end gap-1">61–90d <ChevronUp v-if="sortBy==='days_61_to_90'&&sortDir==='asc'" class="w-3 h-3"/><ChevronDown v-else-if="sortBy==='days_61_to_90'" class="w-3 h-3"/></div>
              </th>
              <th @click="toggleSort('days_over_90')" class="th text-right cursor-pointer whitespace-nowrap">
                <div class="flex items-center justify-end gap-1 text-red-600 dark:text-red-400">90+d <ChevronUp v-if="sortBy==='days_over_90'&&sortDir==='asc'" class="w-3 h-3"/><ChevronDown v-else-if="sortBy==='days_over_90'" class="w-3 h-3"/></div>
              </th>
              <th @click="toggleSort('total_outstanding')" class="th text-right cursor-pointer whitespace-nowrap">
                <div class="flex items-center justify-end gap-1 text-blue-600 dark:text-blue-400">Total <ChevronUp v-if="sortBy==='total_outstanding'&&sortDir==='asc'" class="w-3 h-3"/><ChevronDown v-else-if="sortBy==='total_outstanding'" class="w-3 h-3"/></div>
              </th>
              <th class="th min-w-24">Risk Bar</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr
              v-for="row in filtered"
              :key="row.customer_id"
              class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
              :class="overdueSeverity(row)"
            >
              <!-- Customer Name -->
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <AlertTriangle v-if="row.days_over_90 > 0" class="w-3.5 h-3.5 text-red-500 flex-shrink-0" />
                  <Clock v-else-if="row.days_61_to_90 > 0" class="w-3.5 h-3.5 text-orange-500 flex-shrink-0" />
                  <span class="font-semibold text-gray-900 dark:text-white truncate max-w-44">{{ row.customer_name }}</span>
                </div>
              </td>
              <!-- Contact -->
              <td class="px-4 py-3">
                <div class="space-y-0.5">
                  <div v-if="row.phone" class="flex items-center gap-1 text-xs text-gray-500 dark:text-gray-400">
                    <Phone class="w-3 h-3 flex-shrink-0" />{{ row.phone }}
                  </div>
                  <div v-if="row.email" class="flex items-center gap-1 text-xs text-gray-500 dark:text-gray-400">
                    <Mail class="w-3 h-3 flex-shrink-0" /><span class="truncate max-w-36">{{ row.email }}</span>
                  </div>
                </div>
              </td>
              <!-- Invoice count -->
              <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-400 tabular-nums">{{ row.invoice_count }}</td>
              <!-- Buckets -->
              <td class="px-4 py-3 text-right text-xs" :class="bucketColor(row.current_amount)">
                {{ row.current_amount ? row.current_amount.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) : '—' }}
              </td>
              <td class="px-4 py-3 text-right text-xs" :class="row.days_1_to_30 > 0 ? 'text-amber-700 dark:text-amber-400 font-medium tabular-nums' : 'text-gray-300 dark:text-gray-700'">
                {{ row.days_1_to_30 ? row.days_1_to_30.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) : '—' }}
              </td>
              <td class="px-4 py-3 text-right text-xs" :class="row.days_31_to_60 > 0 ? 'text-orange-700 dark:text-orange-400 font-medium tabular-nums' : 'text-gray-300 dark:text-gray-700'">
                {{ row.days_31_to_60 ? row.days_31_to_60.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) : '—' }}
              </td>
              <td class="px-4 py-3 text-right text-xs" :class="row.days_61_to_90 > 0 ? 'text-red-600 dark:text-red-400 font-semibold tabular-nums' : 'text-gray-300 dark:text-gray-700'">
                {{ row.days_61_to_90 ? row.days_61_to_90.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) : '—' }}
              </td>
              <td class="px-4 py-3 text-right text-xs" :class="row.days_over_90 > 0 ? 'text-red-700 dark:text-red-300 font-bold tabular-nums' : 'text-gray-300 dark:text-gray-700'">
                <div class="flex items-center justify-end gap-1">
                  <AlertTriangle v-if="row.days_over_90 > 0" class="w-3 h-3 text-red-500 flex-shrink-0" />
                  {{ row.days_over_90 ? row.days_over_90.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) : '—' }}
                </div>
              </td>
              <!-- Total -->
              <td class="px-4 py-3 text-right font-bold text-blue-700 dark:text-blue-400 tabular-nums text-sm">
                {{ row.total_outstanding.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) }}
              </td>
              <!-- Risk Bar -->
              <td class="px-4 py-3">
                <div class="h-2 rounded-full bg-gray-100 dark:bg-gray-800 overflow-hidden w-20">
                  <div class="h-full flex">
                    <div v-if="row.current_amount" class="bg-green-400 dark:bg-green-500" :style="{ width: (row.current_amount / row.total_outstanding * 100) + '%' }" />
                    <div v-if="row.days_1_to_30" class="bg-amber-400 dark:bg-amber-500" :style="{ width: (row.days_1_to_30 / row.total_outstanding * 100) + '%' }" />
                    <div v-if="row.days_31_to_60" class="bg-orange-400 dark:bg-orange-500" :style="{ width: (row.days_31_to_60 / row.total_outstanding * 100) + '%' }" />
                    <div v-if="row.days_61_to_90" class="bg-red-400 dark:bg-red-500" :style="{ width: (row.days_61_to_90 / row.total_outstanding * 100) + '%' }" />
                    <div v-if="row.days_over_90" class="bg-red-700 dark:bg-red-800" :style="{ width: (row.days_over_90 / row.total_outstanding * 100) + '%' }" />
                  </div>
                </div>
              </td>
            </tr>
          </tbody>
          <!-- Totals Row -->
          <tfoot class="bg-gray-100 dark:bg-gray-800 border-t-2 border-gray-300 dark:border-gray-600">
            <tr>
              <td class="px-4 py-3 font-bold text-gray-900 dark:text-white" colspan="2">TOTALS</td>
              <td class="px-4 py-3 text-right font-bold text-gray-700 dark:text-gray-300 tabular-nums">{{ totals.invoice_count }}</td>
              <td class="px-4 py-3 text-right font-semibold text-gray-700 dark:text-gray-300 tabular-nums text-xs">{{ totals.current_amount.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) }}</td>
              <td class="px-4 py-3 text-right font-semibold text-amber-700 dark:text-amber-400 tabular-nums text-xs">{{ totals.days_1_to_30.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) }}</td>
              <td class="px-4 py-3 text-right font-semibold text-orange-700 dark:text-orange-400 tabular-nums text-xs">{{ totals.days_31_to_60.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) }}</td>
              <td class="px-4 py-3 text-right font-semibold text-red-600 dark:text-red-400 tabular-nums text-xs">{{ totals.days_61_to_90.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) }}</td>
              <td class="px-4 py-3 text-right font-bold text-red-700 dark:text-red-300 tabular-nums text-xs">{{ totals.days_over_90.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) }}</td>
              <td class="px-4 py-3 text-right font-bold text-blue-700 dark:text-blue-400 tabular-nums text-sm">{{ totals.total_outstanding.toLocaleString('fr-DZ', { minimumFractionDigits: 0 }) }}</td>
              <td class="px-4 py-3"/>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>

    <!-- Legend -->
    <div class="flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400 px-1">
      <div class="flex items-center gap-1.5"><div class="w-3 h-3 rounded-full bg-green-400"></div>Current (not due)</div>
      <div class="flex items-center gap-1.5"><div class="w-3 h-3 rounded-full bg-amber-400"></div>1–30 days overdue</div>
      <div class="flex items-center gap-1.5"><div class="w-3 h-3 rounded-full bg-orange-400"></div>31–60 days overdue</div>
      <div class="flex items-center gap-1.5"><div class="w-3 h-3 rounded-full bg-red-400"></div>61–90 days overdue</div>
      <div class="flex items-center gap-1.5"><div class="w-3 h-3 rounded-full bg-red-700"></div>Over 90 days (critical)</div>
    </div>

  </div>
</template>

<style scoped>
.th { @apply px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide whitespace-nowrap; }
</style>
