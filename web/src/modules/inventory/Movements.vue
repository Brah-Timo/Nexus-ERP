<template>
  <div class="flex flex-col h-full bg-gray-50 dark:bg-gray-950 min-h-screen">

    <!-- Header -->
    <div class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Stock Movements</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Complete audit trail of all inventory movements</p>
        </div>
        <div class="flex items-center gap-3">
          <button @click="openAdjustment"
            class="inline-flex items-center gap-2 bg-amber-600 hover:bg-amber-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            <SlidersHorizontal :size="15" />
            Adjustment
          </button>
          <button @click="openTransfer"
            class="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            <ArrowLeftRight :size="15" />
            Transfer
          </button>
        </div>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="px-6 py-4 grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Total Movements</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ movements.length }}</p>
          </div>
          <div class="w-10 h-10 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg flex items-center justify-center">
            <ArrowUpDown :size="20" class="text-indigo-600 dark:text-indigo-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Purchases</p>
            <p class="text-2xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">{{ countByType('purchase') }}</p>
          </div>
          <div class="w-10 h-10 bg-emerald-100 dark:bg-emerald-900/40 rounded-lg flex items-center justify-center">
            <ShoppingCart :size="20" class="text-emerald-600 dark:text-emerald-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Transfers</p>
            <p class="text-2xl font-bold text-blue-600 dark:text-blue-400 mt-1">{{ countByType('transfer') }}</p>
          </div>
          <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900/40 rounded-lg flex items-center justify-center">
            <ArrowLeftRight :size="20" class="text-blue-600 dark:text-blue-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Adjustments</p>
            <p class="text-2xl font-bold text-amber-600 dark:text-amber-400 mt-1">{{ countByType('adjustment') }}</p>
          </div>
          <div class="w-10 h-10 bg-amber-100 dark:bg-amber-900/40 rounded-lg flex items-center justify-center">
            <SlidersHorizontal :size="20" class="text-amber-600 dark:text-amber-400" />
          </div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="px-6 pb-4 flex flex-wrap gap-3">
      <div class="relative flex-1 min-w-64">
        <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
        <input v-model="search" placeholder="Search by number, item, reference..." type="text"
          class="w-full pl-9 pr-4 py-2 text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
      </div>
      <select v-model="filterType" @change="load"
        class="text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
        <option value="">All Types</option>
        <option value="purchase">Purchase</option>
        <option value="sale">Sale</option>
        <option value="transfer">Transfer</option>
        <option value="adjustment">Adjustment</option>
        <option value="return_in">Return In</option>
        <option value="return_out">Return Out</option>
        <option value="production_in">Production In</option>
        <option value="production_out">Production Out</option>
      </select>
      <select v-model="filterWarehouse" @change="load"
        class="text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
        <option value="">All Warehouses</option>
        <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
      </select>
    </div>

    <!-- Table -->
    <div class="flex-1 px-6 pb-6">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div v-if="loading" class="flex items-center justify-center py-16">
          <Loader2 :size="28" class="animate-spin text-indigo-500" />
          <span class="ml-3 text-gray-500 dark:text-gray-400">Loading movements...</span>
        </div>
        <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center py-16">
          <ArrowUpDown :size="48" class="text-gray-300 dark:text-gray-600 mb-3" />
          <p class="text-gray-500 dark:text-gray-400 font-medium">No movements found</p>
          <p class="text-sm text-gray-400 dark:text-gray-500 mt-1">Record your first adjustment or transfer</p>
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
              <tr v-for="m in paginated" :key="m.id"
                class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                <td class="px-4 py-3">
                  <span class="font-mono text-xs font-semibold text-indigo-600 dark:text-indigo-400">{{ m.number }}</span>
                </td>
                <td class="px-4 py-3 text-gray-600 dark:text-gray-300 text-xs">{{ m.date ? m.date.slice(0, 10) : '-' }}</td>
                <td class="px-4 py-3">
                  <span :class="typeBadge(m.type)" class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium whitespace-nowrap">
                    <component :is="typeIcon(m.type)" :size="10" />
                    {{ typeLabel(m.type) }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <div class="font-mono text-xs font-semibold text-gray-700 dark:text-gray-200">{{ m.item_code }}</div>
                  <div class="text-xs text-gray-400 mt-0.5 max-w-[140px] truncate">{{ m.item_name }}</div>
                </td>
                <td class="px-4 py-3 text-xs text-gray-600 dark:text-gray-300">{{ m.warehouse_name }}</td>
                <td class="px-4 py-3 text-xs text-gray-500 dark:text-gray-400">
                  <span v-if="m.to_warehouse_name">{{ m.to_warehouse_name }}</span>
                  <span v-else class="text-gray-300 dark:text-gray-600">—</span>
                </td>
                <td class="px-4 py-3 text-right">
                  <span :class="m.quantity < 0 ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'" class="font-semibold">
                    {{ m.quantity > 0 ? '+' : '' }}{{ fmtQty(m.quantity) }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-300">{{ fmt(m.unit_cost) }}</td>
                <td class="px-4 py-3 text-right font-medium text-gray-900 dark:text-white">{{ fmt(m.total_cost) }}</td>
                <td class="px-4 py-3 text-xs text-gray-500 dark:text-gray-400">{{ m.reference || '-' }}</td>
                <td class="px-4 py-3 text-xs text-gray-500 dark:text-gray-400">{{ m.created_by_name || '-' }}</td>
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

    <!-- Adjustment Modal -->
    <Teleport to="body">
      <div v-if="showAdjModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showAdjModal = false"></div>
        <div class="relative bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-lg">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 bg-amber-100 dark:bg-amber-900/40 rounded-lg flex items-center justify-center">
                <SlidersHorizontal :size="16" class="text-amber-600 dark:text-amber-400" />
              </div>
              <h2 class="text-lg font-bold text-gray-900 dark:text-white">Stock Adjustment</h2>
            </div>
            <button @click="showAdjModal = false" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400">
              <X :size="18" />
            </button>
          </div>
          <div class="px-6 py-4 space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Item <span class="text-red-500">*</span></label>
              <select v-model="adjForm.item_id"
                class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="">Select item...</option>
                <option v-for="item in items" :key="item.id" :value="item.id">
                  {{ item.code }} – {{ item.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Warehouse <span class="text-red-500">*</span></label>
              <select v-model="adjForm.warehouse_id"
                class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="">Select warehouse...</option>
                <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
              </select>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Quantity <span class="text-red-500">*</span></label>
                <input v-model.number="adjForm.quantity" type="number" step="0.001"
                  placeholder="Use negative for reduction"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Unit Cost</label>
                <input v-model.number="adjForm.unit_cost" type="number" step="0.0001" min="0"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Reference</label>
              <input v-model="adjForm.reference" type="text" placeholder="Optional reference number"
                class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Notes</label>
              <textarea v-model="adjForm.notes" rows="2" placeholder="Reason for adjustment"
                class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"></textarea>
            </div>
          </div>
          <div class="border-t border-gray-200 dark:border-gray-800 px-6 py-4 flex items-center justify-end gap-3">
            <button @click="showAdjModal = false" class="text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 px-4 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
              Cancel
            </button>
            <button @click="saveAdjustment" :disabled="saving"
              class="inline-flex items-center gap-2 bg-amber-600 hover:bg-amber-700 disabled:opacity-60 text-white text-sm font-medium px-5 py-2 rounded-lg transition-colors">
              <Loader2 v-if="saving" :size="14" class="animate-spin" />
              <SlidersHorizontal v-else :size="14" />
              Record Adjustment
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Transfer Modal -->
    <Teleport to="body">
      <div v-if="showTransferModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showTransferModal = false"></div>
        <div class="relative bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-lg">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg flex items-center justify-center">
                <ArrowLeftRight :size="16" class="text-indigo-600 dark:text-indigo-400" />
              </div>
              <h2 class="text-lg font-bold text-gray-900 dark:text-white">Stock Transfer</h2>
            </div>
            <button @click="showTransferModal = false" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400">
              <X :size="18" />
            </button>
          </div>
          <div class="px-6 py-4 space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Item <span class="text-red-500">*</span></label>
              <select v-model="trForm.item_id"
                class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="">Select item...</option>
                <option v-for="item in items" :key="item.id" :value="item.id">
                  {{ item.code }} – {{ item.name }}
                </option>
              </select>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">From Warehouse <span class="text-red-500">*</span></label>
                <select v-model="trForm.warehouse_id"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                  <option value="">Select...</option>
                  <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">To Warehouse <span class="text-red-500">*</span></label>
                <select v-model="trForm.to_warehouse_id"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                  <option value="">Select...</option>
                  <option v-for="w in warehouses.filter(x => x.id !== trForm.warehouse_id)" :key="w.id" :value="w.id">{{ w.name }}</option>
                </select>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Quantity <span class="text-red-500">*</span></label>
                <input v-model.number="trForm.quantity" type="number" min="0.001" step="0.001"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Reference</label>
                <input v-model="trForm.reference" type="text" placeholder="Optional"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Notes</label>
              <textarea v-model="trForm.notes" rows="2"
                class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"></textarea>
            </div>
          </div>
          <div class="border-t border-gray-200 dark:border-gray-800 px-6 py-4 flex items-center justify-end gap-3">
            <button @click="showTransferModal = false" class="text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 px-4 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
              Cancel
            </button>
            <button @click="saveTransfer" :disabled="saving"
              class="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 text-white text-sm font-medium px-5 py-2 rounded-lg transition-colors">
              <Loader2 v-if="saving" :size="14" class="animate-spin" />
              <ArrowLeftRight v-else :size="14" />
              Confirm Transfer
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  ArrowUpDown, ShoppingCart, ArrowLeftRight, SlidersHorizontal, Search, Loader2,
  X, ChevronLeft, ChevronRight, ChevronsUpDown, ChevronUp, ChevronDown,
  TrendingUp, TrendingDown, RefreshCcw, RotateCcw, Settings, Factory
} from 'lucide-vue-next'
import { inventoryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ────────────────────────────────────────────────────────────────────

interface Movement {
  id: string
  company_id: string
  number: string
  date: string
  type: string
  item_id: string
  item_code: string
  item_name: string
  warehouse_id: string
  warehouse_name: string
  to_warehouse_id?: string
  to_warehouse_name?: string
  quantity: number
  unit_cost: number
  total_cost: number
  reference?: string
  source_type?: string
  notes?: string
  created_by?: string
  created_by_name?: string
  created_at: string
}

interface SimpleItem { id: string; code: string; name: string }
interface SimpleWarehouse { id: string; code: string; name: string }

// ─── State ────────────────────────────────────────────────────────────────────

const movements = ref<Movement[]>([])
const items = ref<SimpleItem[]>([])
const warehouses = ref<SimpleWarehouse[]>([])
const loading = ref(false)
const saving = ref(false)

const search = ref('')
const filterType = ref('')
const filterWarehouse = ref('')

const sortKey = ref('date')
const sortDir = ref<'asc' | 'desc'>('desc')
const currentPage = ref(1)
const pageSize = 25

const showAdjModal = ref(false)
const showTransferModal = ref(false)

const adjForm = ref({
  item_id: '', warehouse_id: '', quantity: 0, unit_cost: 0, reference: '', notes: ''
})
const trForm = ref({
  item_id: '', warehouse_id: '', to_warehouse_id: '', quantity: 0, reference: '', notes: ''
})

// ─── Columns ──────────────────────────────────────────────────────────────────

const columns = [
  { key: 'number', label: 'Number' },
  { key: 'date', label: 'Date' },
  { key: 'type', label: 'Type' },
  { key: 'item_code', label: 'Item' },
  { key: 'warehouse_name', label: 'From' },
  { key: 'to_warehouse_name', label: 'To' },
  { key: 'quantity', label: 'Qty' },
  { key: 'unit_cost', label: 'Unit Cost' },
  { key: 'total_cost', label: 'Total' },
  { key: 'reference', label: 'Reference' },
  { key: 'created_by_name', label: 'By' },
]

// ─── Computed ─────────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = [...movements.value]
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(m =>
      m.number.toLowerCase().includes(q) ||
      m.item_code.toLowerCase().includes(q) ||
      m.item_name.toLowerCase().includes(q) ||
      (m.reference || '').toLowerCase().includes(q)
    )
  }
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

const countByType = (t: string) => movements.value.filter(m => m.type === t).length

// ─── Helpers ──────────────────────────────────────────────────────────────────

const fmt = (v: number) =>
  new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0)

const fmtQty = (v: number) =>
  new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 4 }).format(v || 0)

const typeLabel = (t: string) => {
  const labels: Record<string, string> = {
    purchase: 'Purchase', sale: 'Sale', transfer: 'Transfer',
    adjustment: 'Adjustment', return_in: 'Return In', return_out: 'Return Out',
    production_in: 'Prod. In', production_out: 'Prod. Out',
  }
  return labels[t] || t
}

const typeBadge = (t: string) => {
  const badges: Record<string, string> = {
    purchase: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400',
    sale: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
    transfer: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400',
    adjustment: 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400',
    return_in: 'bg-teal-100 text-teal-800 dark:bg-teal-900/30 dark:text-teal-400',
    return_out: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
    production_in: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
    production_out: 'bg-rose-100 text-rose-800 dark:bg-rose-900/30 dark:text-rose-400',
  }
  return badges[t] || 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
}

const typeIcon = (t: string) => {
  if (t === 'purchase') return ShoppingCart
  if (t === 'sale') return TrendingDown
  if (t === 'transfer') return ArrowLeftRight
  if (t === 'adjustment') return SlidersHorizontal
  if (t === 'return_in' || t === 'return_out') return RotateCcw
  if (t === 'production_in' || t === 'production_out') return Factory
  return RefreshCcw
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

// ─── Actions ──────────────────────────────────────────────────────────────────

const openAdjustment = () => {
  adjForm.value = { item_id: '', warehouse_id: '', quantity: 0, unit_cost: 0, reference: '', notes: '' }
  showAdjModal.value = true
}

const openTransfer = () => {
  trForm.value = { item_id: '', warehouse_id: '', to_warehouse_id: '', quantity: 0, reference: '', notes: '' }
  showTransferModal.value = true
}

const saveAdjustment = async () => {
  if (!adjForm.value.item_id || !adjForm.value.warehouse_id || adjForm.value.quantity === 0) {
    app.addToast('Item, warehouse, and quantity are required', 'error')
    return
  }
  saving.value = true
  try {
    await inventoryAPI.createMovement({
      item_id: adjForm.value.item_id,
      warehouse_id: adjForm.value.warehouse_id,
      quantity: adjForm.value.quantity,
      unit_cost: adjForm.value.unit_cost || 0,
      reference: adjForm.value.reference || null,
      notes: adjForm.value.notes || null,
    })
    app.addToast('Adjustment recorded', 'success')
    showAdjModal.value = false
    await load()
  } catch {
    app.addToast('Failed to record adjustment', 'error')
  } finally {
    saving.value = false
  }
}

const saveTransfer = async () => {
  if (!trForm.value.item_id || !trForm.value.warehouse_id || !trForm.value.to_warehouse_id || trForm.value.quantity <= 0) {
    app.addToast('Item, both warehouses, and quantity are required', 'error')
    return
  }
  if (trForm.value.warehouse_id === trForm.value.to_warehouse_id) {
    app.addToast('Source and destination warehouses must be different', 'error')
    return
  }
  saving.value = true
  try {
    await inventoryAPI.transferStock({
      item_id: trForm.value.item_id,
      warehouse_id: trForm.value.warehouse_id,
      to_warehouse_id: trForm.value.to_warehouse_id,
      quantity: trForm.value.quantity,
      reference: trForm.value.reference || null,
      notes: trForm.value.notes || null,
    })
    app.addToast('Transfer recorded', 'success')
    showTransferModal.value = false
    await load()
  } catch {
    app.addToast('Failed to record transfer', 'error')
  } finally {
    saving.value = false
  }
}

// ─── Load ─────────────────────────────────────────────────────────────────────

const load = async () => {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (filterType.value) params.type = filterType.value
    if (filterWarehouse.value) params.warehouse_id = filterWarehouse.value

    const [movRes, itemsRes, whRes] = await Promise.all([
      inventoryAPI.getMovements(params),
      inventoryAPI.getItems(),
      inventoryAPI.getWarehouses(),
    ])
    movements.value = movRes.data || []
    items.value = itemsRes.data || []
    warehouses.value = whRes.data || []
    currentPage.value = 1
  } catch {
    app.addToast('Failed to load movements', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
