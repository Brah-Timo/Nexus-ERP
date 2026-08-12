<template>
  <div class="flex flex-col h-full bg-gray-50 dark:bg-gray-950 min-h-screen">

    <!-- Header -->
    <div class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Stock Levels</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Real-time inventory positions across all warehouses</p>
        </div>
        <div class="flex items-center gap-3">
          <button @click="load" class="inline-flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
            <RefreshCw :size="14" :class="loading ? 'animate-spin' : ''" />
            Refresh
          </button>
        </div>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="px-6 py-4 grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Total SKUs</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ levels.length }}</p>
          </div>
          <div class="w-10 h-10 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg flex items-center justify-center">
            <Package :size="20" class="text-indigo-600 dark:text-indigo-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Total Value</p>
            <p class="text-xl font-bold text-gray-900 dark:text-white mt-1">{{ fmt(totalValue) }}</p>
          </div>
          <div class="w-10 h-10 bg-emerald-100 dark:bg-emerald-900/40 rounded-lg flex items-center justify-center">
            <TrendingUp :size="20" class="text-emerald-600 dark:text-emerald-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Low Stock</p>
            <p class="text-2xl font-bold text-amber-600 dark:text-amber-400 mt-1">{{ lowStockCount }}</p>
          </div>
          <div class="w-10 h-10 bg-amber-100 dark:bg-amber-900/40 rounded-lg flex items-center justify-center">
            <AlertTriangle :size="20" class="text-amber-600 dark:text-amber-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Out of Stock</p>
            <p class="text-2xl font-bold text-red-600 dark:text-red-400 mt-1">{{ outOfStockCount }}</p>
          </div>
          <div class="w-10 h-10 bg-red-100 dark:bg-red-900/40 rounded-lg flex items-center justify-center">
            <XCircle :size="20" class="text-red-600 dark:text-red-400" />
          </div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="px-6 pb-4 flex flex-wrap gap-3">
      <div class="relative flex-1 min-w-64">
        <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
        <input v-model="search" placeholder="Search by item code, name, warehouse..." type="text"
          class="w-full pl-9 pr-4 py-2 text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
      </div>
      <select v-model="filterWarehouse"
        class="text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
        <option value="">All Warehouses</option>
        <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
      </select>
      <select v-model="filterStatus"
        class="text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
        <option value="">All Status</option>
        <option value="out_of_stock">Out of Stock</option>
        <option value="low_stock">Low Stock</option>
        <option value="normal">Normal</option>
        <option value="over_stock">Over Stock</option>
      </select>
    </div>

    <!-- Table -->
    <div class="flex-1 px-6 pb-6">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div v-if="loading" class="flex items-center justify-center py-16">
          <Loader2 :size="28" class="animate-spin text-indigo-500" />
          <span class="ml-3 text-gray-500 dark:text-gray-400">Loading stock levels...</span>
        </div>
        <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center py-16">
          <BarChart3 :size="48" class="text-gray-300 dark:text-gray-600 mb-3" />
          <p class="text-gray-500 dark:text-gray-400 font-medium">No stock data found</p>
          <p class="text-sm text-gray-400 dark:text-gray-500 mt-1">Stock levels appear after first inventory movements</p>
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th v-for="col in columns" :key="col.key"
                  @click="sortBy(col.key)"
                  class="px-4 py-3 text-left font-semibold text-gray-600 dark:text-gray-300 cursor-pointer select-none hover:text-gray-900 dark:hover:text-white transition-colors whitespace-nowrap">
                  <div class="flex items-center gap-1">
                    {{ col.label }}
                    <component :is="sortIcon(col.key)" :size="12" class="text-gray-400" />
                  </div>
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
              <tr v-for="row in paginated" :key="row.id"
                :class="rowClass(row)"
                class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                <td class="px-4 py-3">
                  <div class="font-mono text-xs font-semibold text-indigo-600 dark:text-indigo-400">{{ row.item_code }}</div>
                  <div class="text-xs text-gray-400 mt-0.5 max-w-[140px] truncate">{{ row.item_name }}</div>
                </td>
                <td class="px-4 py-3 text-gray-600 dark:text-gray-300 text-xs">{{ row.warehouse_name }}</td>
                <td class="px-4 py-3">
                  <span :class="row.item_type === 'storable' ? 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400' : 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'"
                    class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium">
                    {{ row.item_type }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right">
                  <span class="font-semibold text-gray-900 dark:text-white">{{ fmtQty(row.qty_on_hand) }}</span>
                  <span v-if="row.uom_code" class="text-xs text-gray-400 ml-1">{{ row.uom_code }}</span>
                </td>
                <td class="px-4 py-3 text-right">
                  <span class="text-gray-600 dark:text-gray-300">{{ fmtQty(row.qty_reserved) }}</span>
                </td>
                <td class="px-4 py-3 text-right">
                  <span :class="row.qty_available <= 0 ? 'text-red-600 dark:text-red-400 font-semibold' : row.qty_available < row.min_stock_qty ? 'text-amber-600 dark:text-amber-400 font-semibold' : 'text-emerald-600 dark:text-emerald-400 font-semibold'">
                    {{ fmtQty(row.qty_available) }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-300">{{ fmt(row.cmup_cost) }}</td>
                <td class="px-4 py-3 text-right font-medium text-gray-900 dark:text-white">{{ fmt(row.total_value) }}</td>
                <td class="px-4 py-3 text-right text-gray-500 dark:text-gray-400">{{ fmtQty(row.min_stock_qty) }}</td>
                <td class="px-4 py-3 text-right text-gray-500 dark:text-gray-400">{{ fmtQty(row.reorder_qty) }}</td>
                <td class="px-4 py-3">
                  <span :class="statusBadge(row.stock_status)" class="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium whitespace-nowrap">
                    <component :is="statusIcon(row.stock_status)" :size="10" class="mr-1" />
                    {{ statusLabel(row.stock_status) }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right text-xs text-gray-400">
                  {{ row.updated_at ? row.updated_at.slice(0, 10) : '-' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="!loading && filtered.length > 0" class="border-t border-gray-200 dark:border-gray-700 px-4 py-3 flex items-center justify-between">
          <p class="text-sm text-gray-500 dark:text-gray-400">
            Showing {{ (currentPage - 1) * pageSize + 1 }}–{{ Math.min(currentPage * pageSize, filtered.length) }} of {{ filtered.length }}
          </p>
          <div class="flex items-center gap-1">
            <button @click="currentPage--" :disabled="currentPage === 1"
              class="p-1.5 rounded-lg text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
              <ChevronLeft :size="16" />
            </button>
            <span class="text-sm text-gray-600 dark:text-gray-300 px-2">{{ currentPage }} / {{ totalPages }}</span>
            <button @click="currentPage++" :disabled="currentPage === totalPages"
              class="p-1.5 rounded-lg text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
              <ChevronRight :size="16" />
            </button>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Package, TrendingUp, AlertTriangle, XCircle, Search, RefreshCw, Loader2,
  BarChart3, ChevronLeft, ChevronRight, ChevronsUpDown, ChevronUp, ChevronDown,
  CheckCircle, MinusCircle
} from 'lucide-vue-next'
import { inventoryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── State ────────────────────────────────────────────────────────────────────

interface StockLevel {
  id: string
  company_id: string
  item_id: string
  item_code: string
  item_name: string
  item_type: string
  category_name?: string
  uom_code?: string
  warehouse_id: string
  warehouse_name: string
  location_id?: string
  qty_on_hand: number
  qty_reserved: number
  qty_available: number
  cmup_cost: number
  total_value: number
  min_stock_qty: number
  reorder_qty: number
  max_stock_qty: number
  stock_status: string
  updated_at?: string
}

interface Warehouse { id: string; name: string; code: string }

const levels = ref<StockLevel[]>([])
const warehouses = ref<Warehouse[]>([])
const loading = ref(false)

const search = ref('')
const filterWarehouse = ref('')
const filterStatus = ref('')

const sortKey = ref('item_code')
const sortDir = ref<'asc' | 'desc'>('asc')
const currentPage = ref(1)
const pageSize = 25

// ─── Columns ──────────────────────────────────────────────────────────────────

const columns = [
  { key: 'item_code', label: 'Item Code' },
  { key: 'warehouse_name', label: 'Warehouse' },
  { key: 'item_type', label: 'Type' },
  { key: 'qty_on_hand', label: 'On Hand' },
  { key: 'qty_reserved', label: 'Reserved' },
  { key: 'qty_available', label: 'Available' },
  { key: 'cmup_cost', label: 'CMUP' },
  { key: 'total_value', label: 'Value' },
  { key: 'min_stock_qty', label: 'Min Stock' },
  { key: 'reorder_qty', label: 'Reorder' },
  { key: 'stock_status', label: 'Status' },
  { key: 'updated_at', label: 'Updated' },
]

// ─── Computed ─────────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = [...levels.value]
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(l =>
      l.item_code.toLowerCase().includes(q) ||
      l.item_name.toLowerCase().includes(q) ||
      l.warehouse_name.toLowerCase().includes(q)
    )
  }
  if (filterWarehouse.value) list = list.filter(l => l.warehouse_id === filterWarehouse.value)
  if (filterStatus.value) list = list.filter(l => l.stock_status === filterStatus.value)

  list.sort((a, b) => {
    const va = (a as Record<string, unknown>)[sortKey.value]
    const vb = (b as Record<string, unknown>)[sortKey.value]
    const cmp = String(va ?? '').localeCompare(String(vb ?? ''), undefined, { numeric: true })
    return sortDir.value === 'asc' ? cmp : -cmp
  })
  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filtered.value.length / pageSize)))
const paginated = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filtered.value.slice(start, start + pageSize)
})

const totalValue = computed(() => levels.value.reduce((s, l) => s + (l.total_value || 0), 0))
const lowStockCount = computed(() => levels.value.filter(l => l.stock_status === 'low_stock').length)
const outOfStockCount = computed(() => levels.value.filter(l => l.stock_status === 'out_of_stock').length)

// ─── Helpers ──────────────────────────────────────────────────────────────────

const fmt = (v: number) =>
  new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0)

const fmtQty = (v: number) =>
  new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 4 }).format(v || 0)

const rowClass = (row: StockLevel) => {
  if (row.stock_status === 'out_of_stock') return 'bg-red-50/40 dark:bg-red-900/10'
  if (row.stock_status === 'low_stock') return 'bg-amber-50/40 dark:bg-amber-900/10'
  return ''
}

const statusLabel = (s: string) => {
  if (s === 'out_of_stock') return 'Out of Stock'
  if (s === 'low_stock') return 'Low Stock'
  if (s === 'over_stock') return 'Over Stock'
  return 'Normal'
}

const statusBadge = (s: string) => {
  if (s === 'out_of_stock') return 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
  if (s === 'low_stock') return 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400'
  if (s === 'over_stock') return 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400'
  return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400'
}

const statusIcon = (s: string) => {
  if (s === 'out_of_stock') return XCircle
  if (s === 'low_stock') return AlertTriangle
  if (s === 'over_stock') return MinusCircle
  return CheckCircle
}

const sortBy = (key: string) => {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'asc'
  }
  currentPage.value = 1
}

const sortIcon = (key: string) => {
  if (sortKey.value !== key) return ChevronsUpDown
  return sortDir.value === 'asc' ? ChevronUp : ChevronDown
}

// ─── Load ─────────────────────────────────────────────────────────────────────

const load = async () => {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (filterWarehouse.value) params.warehouse_id = filterWarehouse.value
    const [levelsRes, whRes] = await Promise.all([
      inventoryAPI.getStockLevels(params),
      inventoryAPI.getWarehouses(),
    ])
    levels.value = levelsRes.data || []
    warehouses.value = whRes.data || []
    currentPage.value = 1
  } catch {
    app.addToast('Failed to load stock levels', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
