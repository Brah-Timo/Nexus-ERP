<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { purchaseAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  Plus, X, Save, Search, RefreshCw, Eye, CheckCircle,
  ChevronDown, ChevronUp, Trash2, Package, Truck,
  Calendar, FileText, Building2, AlertTriangle, ClipboardList
} from 'lucide-vue-next'

// ─── Types ───────────────────────────────────────────────────────────────────

interface GRLine {
  po_line_id: string
  item_id: string
  description: string
  expected_qty: number
  received_qty: number
  unit_cost: number
}

interface GoodsReceipt {
  id: string
  number: string
  po_id: string
  po_number?: string
  supplier_id: string
  supplier_name: string
  date: string
  warehouse_id: string
  status: string
  total_amount: number
  notes: string
  created_at: string
  lines?: GRLine[]
}

interface PurchaseOrder {
  id: string
  number: string
  supplier_name: string
  supplier_id: string
  total_amount: number
  status: string
  lines?: POLine[]
}

interface POLine {
  id: string
  description: string
  quantity: number
  received_qty: number
  unit_price: number
  unit_cost?: number
}

const EMPTY_LINE = (): GRLine => ({
  po_line_id: '',
  item_id: '',
  description: '',
  expected_qty: 0,
  received_qty: 0,
  unit_cost: 0,
})

const STATUS_BADGE: Record<string, string> = {
  draft:    'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  received: 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300',
}

const STATUS_LABEL: Record<string, string> = {
  draft:    'Draft',
  received: 'Received',
}

// ─── State ───────────────────────────────────────────────────────────────────

const app = useAppStore()
const receipts  = ref<GoodsReceipt[]>([])
const orders    = ref<PurchaseOrder[]>([])
const loading   = ref(true)
const saving    = ref(false)

const showForm   = ref(false)
const showDetail = ref(false)

const form = ref<{
  po_id: string
  date: string
  warehouse_id: string
  notes: string
  lines: GRLine[]
}>({
  po_id: '',
  date: today(),
  warehouse_id: '',
  notes: '',
  lines: [EMPTY_LINE()],
})

const detailReceipt  = ref<GoodsReceipt | null>(null)
const selectedOrder  = ref<PurchaseOrder | null>(null)

const search      = ref('')
const filterStatus = ref('')
const sortBy      = ref<keyof GoodsReceipt>('date')
const sortDir     = ref<'asc' | 'desc'>('desc')

// ─── Computed ────────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = [...receipts.value]
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(r =>
      r.number.toLowerCase().includes(q) ||
      r.supplier_name.toLowerCase().includes(q) ||
      (r.po_number ?? '').toLowerCase().includes(q)
    )
  }
  if (filterStatus.value) {
    list = list.filter(r => r.status === filterStatus.value)
  }
  list.sort((a, b) => {
    const av = a[sortBy.value] ?? ''
    const bv = b[sortBy.value] ?? ''
    const cmp = String(av).localeCompare(String(bv), undefined, { numeric: true })
    return sortDir.value === 'asc' ? cmp : -cmp
  })
  return list
})

const kpis = computed(() => {
  const all = receipts.value
  return {
    total:     all.length,
    draft:     all.filter(r => r.status === 'draft').length,
    received:  all.filter(r => r.status === 'received').length,
    totalValue: all.reduce((s, r) => s + (r.total_amount ?? 0), 0),
  }
})

const formTotal = computed(() =>
  form.value.lines.reduce((s, l) => s + l.received_qty * l.unit_cost, 0)
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

function today() {
  return new Date().toISOString().slice(0, 10)
}

function fmt(n: number) {
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n)
}

function fmtDate(d: string | null | undefined) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('fr-DZ')
}

function toggleSort(col: keyof GoodsReceipt) {
  if (sortBy.value === col) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = col
    sortDir.value = 'desc'
  }
}

// ─── Data Loading ─────────────────────────────────────────────────────────────

async function loadReceipts() {
  loading.value = true
  try {
    const r = await purchaseAPI.getReceipts()
    receipts.value = r.data ?? []
  } catch {
    app.addToast('Failed to load goods receipts', 'error')
  } finally {
    loading.value = false
  }
}

async function loadOrders() {
  try {
    const r = await purchaseAPI.getOrders()
    // Only show approved orders that can still receive
    orders.value = (r.data ?? []).filter((o: PurchaseOrder) =>
      ['approved', 'partially_received'].includes(o.status)
    )
  } catch {
    orders.value = []
  }
}

// ─── Form Actions ─────────────────────────────────────────────────────────────

function openCreate() {
  form.value = {
    po_id: '',
    date: today(),
    warehouse_id: '',
    notes: '',
    lines: [EMPTY_LINE()],
  }
  selectedOrder.value = null
  showForm.value = true
}

async function onPOSelect() {
  if (!form.value.po_id) {
    selectedOrder.value = null
    form.value.lines = [EMPTY_LINE()]
    return
  }
  try {
    const r = await purchaseAPI.getOrder(form.value.po_id)
    selectedOrder.value = r.data
    // Pre-fill lines from PO lines
    if (r.data?.lines?.length) {
      form.value.lines = r.data.lines.map((l: POLine) => ({
        po_line_id:  l.id,
        item_id:     '',
        description: l.description,
        expected_qty: l.quantity - (l.received_qty ?? 0),
        received_qty: l.quantity - (l.received_qty ?? 0),
        unit_cost:   l.unit_price,
      }))
    } else {
      form.value.lines = [EMPTY_LINE()]
    }
  } catch {
    selectedOrder.value = null
    form.value.lines = [EMPTY_LINE()]
  }
}

function addLine() {
  form.value.lines.push(EMPTY_LINE())
}

function removeLine(i: number) {
  if (form.value.lines.length > 1) form.value.lines.splice(i, 1)
}

async function saveReceipt() {
  if (!form.value.po_id) {
    app.addToast('Please select a Purchase Order', 'error')
    return
  }
  const validLines = form.value.lines.filter(l => l.description.trim())
  if (!validLines.length) {
    app.addToast('Add at least one line with a description', 'error')
    return
  }
  saving.value = true
  try {
    await purchaseAPI.createReceipt({
      po_id:        form.value.po_id,
      date:         form.value.date,
      warehouse_id: form.value.warehouse_id,
      notes:        form.value.notes,
      lines:        validLines,
    })
    app.addToast('Goods receipt created', 'success')
    showForm.value = false
    await loadReceipts()
    await loadOrders()
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Save failed'
    app.addToast(msg, 'error')
  } finally {
    saving.value = false
  }
}

// ─── Detail ───────────────────────────────────────────────────────────────────

async function viewDetail(id: string) {
  try {
    const r = await purchaseAPI.getReceipt(id)
    detailReceipt.value = r.data
    showDetail.value = true
  } catch {
    app.addToast('Failed to load receipt details', 'error')
  }
}

// ─── Validate ────────────────────────────────────────────────────────────────

async function validateReceipt(id: string) {
  if (!confirm('Validate this goods receipt? This will update PO received quantities.')) return
  try {
    await purchaseAPI.validateReceipt(id)
    app.addToast('Goods receipt validated', 'success')
    await loadReceipts()
    if (detailReceipt.value?.id === id) {
      const r = await purchaseAPI.getReceipt(id)
      detailReceipt.value = r.data
    }
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Validation failed'
    app.addToast(msg, 'error')
  }
}

// ─── Lifecycle ───────────────────────────────────────────────────────────────

onMounted(async () => {
  await Promise.all([loadReceipts(), loadOrders()])
})
</script>

<template>
  <div class="space-y-6">

    <!-- ── Header ──────────────────────────────────────────────────────────── -->
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white">Goods Receipts</h1>
        <p class="mt-0.5 text-sm text-slate-500 dark:text-slate-400">
          Receive and validate incoming shipments against purchase orders
        </p>
      </div>
      <button
        @click="openCreate"
        class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-700 transition-colors"
      >
        <Plus class="h-4 w-4" />
        New Receipt
      </button>
    </div>

    <!-- ── KPI Cards ───────────────────────────────────────────────────────── -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
        <div class="flex items-center justify-between">
          <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Total Receipts</p>
          <ClipboardList class="h-4 w-4 text-slate-400" />
        </div>
        <p class="mt-2 text-2xl font-bold text-slate-900 dark:text-white">{{ kpis.total }}</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
        <div class="flex items-center justify-between">
          <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Draft</p>
          <FileText class="h-4 w-4 text-gray-400" />
        </div>
        <p class="mt-2 text-2xl font-bold text-gray-700 dark:text-gray-300">{{ kpis.draft }}</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
        <div class="flex items-center justify-between">
          <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Validated</p>
          <CheckCircle class="h-4 w-4 text-green-500" />
        </div>
        <p class="mt-2 text-2xl font-bold text-green-600 dark:text-green-400">{{ kpis.received }}</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
        <div class="flex items-center justify-between">
          <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Total Value</p>
          <Package class="h-4 w-4 text-blue-500" />
        </div>
        <p class="mt-2 text-xl font-bold text-slate-900 dark:text-white">{{ fmt(kpis.totalValue) }}</p>
        <p class="text-xs text-slate-400">DZD</p>
      </div>
    </div>

    <!-- ── Toolbar ─────────────────────────────────────────────────────────── -->
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
      <div class="relative flex-1">
        <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
        <input
          v-model="search"
          type="text"
          placeholder="Search by GRN#, supplier, PO..."
          class="w-full rounded-lg border border-slate-200 bg-white py-2 pl-9 pr-3 text-sm text-slate-900 placeholder-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-800 dark:text-white dark:placeholder-slate-500"
        />
      </div>
      <select
        v-model="filterStatus"
        class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200"
      >
        <option value="">All Statuses</option>
        <option value="draft">Draft</option>
        <option value="received">Received</option>
      </select>
      <button
        @click="loadReceipts"
        class="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-600 hover:bg-slate-50 transition-colors dark:border-slate-600 dark:bg-slate-800 dark:text-slate-300"
      >
        <RefreshCw class="h-4 w-4" />
      </button>
    </div>

    <!-- ── Table ───────────────────────────────────────────────────────────── -->
    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent"></div>
      </div>
      <div v-else-if="!filtered.length" class="flex flex-col items-center justify-center py-20 text-center">
        <Package class="mb-3 h-12 w-12 text-slate-300 dark:text-slate-600" />
        <p class="text-sm font-medium text-slate-500 dark:text-slate-400">No goods receipts found</p>
        <p class="mt-1 text-xs text-slate-400 dark:text-slate-500">Create a receipt from an approved purchase order</p>
      </div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="border-b border-slate-200 bg-slate-50 dark:border-slate-700 dark:bg-slate-900/50">
            <th
              v-for="col in [
                { key: 'number', label: 'GRN#' },
                { key: 'po_id', label: 'PO#' },
                { key: 'supplier_name', label: 'Supplier' },
                { key: 'date', label: 'Date' },
                { key: 'status', label: 'Status' },
                { key: 'total_amount', label: 'Total Amount' },
              ]"
              :key="col.key"
              @click="toggleSort(col.key as keyof GoodsReceipt)"
              class="cursor-pointer select-none px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200"
            >
              <span class="inline-flex items-center gap-1">
                {{ col.label }}
                <component
                  :is="sortBy === col.key ? (sortDir === 'asc' ? ChevronUp : ChevronDown) : ChevronDown"
                  :class="['h-3 w-3', sortBy === col.key ? 'text-blue-500' : 'opacity-30']"
                />
              </span>
            </th>
            <th class="px-4 py-3 text-right text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">Actions</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
          <tr
            v-for="rec in filtered"
            :key="rec.id"
            class="group hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors"
          >
            <td class="px-4 py-3 font-mono text-xs font-semibold text-blue-600 dark:text-blue-400">
              {{ rec.number }}
            </td>
            <td class="px-4 py-3 font-mono text-xs text-slate-600 dark:text-slate-400">
              {{ rec.po_number || rec.po_id.slice(0, 8) + '...' }}
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-purple-100 dark:bg-purple-900/40">
                  <Truck class="h-3.5 w-3.5 text-purple-600 dark:text-purple-400" />
                </div>
                <span class="font-medium text-slate-900 dark:text-white">{{ rec.supplier_name }}</span>
              </div>
            </td>
            <td class="px-4 py-3 text-slate-600 dark:text-slate-400">
              <span class="inline-flex items-center gap-1.5">
                <Calendar class="h-3.5 w-3.5 text-slate-400" />
                {{ fmtDate(rec.date) }}
              </span>
            </td>
            <td class="px-4 py-3">
              <span
                :class="['inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium', STATUS_BADGE[rec.status] ?? STATUS_BADGE.draft]"
              >
                {{ STATUS_LABEL[rec.status] ?? rec.status }}
              </span>
            </td>
            <td class="px-4 py-3 text-right font-semibold text-slate-900 dark:text-white">
              {{ fmt(rec.total_amount) }}
              <span class="ml-1 text-xs font-normal text-slate-400">DZD</span>
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center justify-end gap-1">
                <button
                  @click="viewDetail(rec.id)"
                  class="rounded p-1.5 text-slate-400 hover:bg-slate-100 hover:text-blue-600 dark:hover:bg-slate-700 dark:hover:text-blue-400 transition-colors"
                  title="View Details"
                >
                  <Eye class="h-4 w-4" />
                </button>
                <button
                  v-if="rec.status === 'draft'"
                  @click="validateReceipt(rec.id)"
                  class="rounded p-1.5 text-slate-400 hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/30 dark:hover:text-green-400 transition-colors"
                  title="Validate Receipt"
                >
                  <CheckCircle class="h-4 w-4" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="!loading && filtered.length" class="border-t border-slate-100 px-4 py-2 dark:border-slate-700">
        <p class="text-xs text-slate-400">
          Showing {{ filtered.length }} of {{ receipts.length }} receipts
        </p>
      </div>
    </div>

    <!-- ── Create Form Modal ───────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showForm" class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/60 p-4 backdrop-blur-sm">
        <div class="my-6 w-full max-w-4xl rounded-2xl bg-white shadow-2xl dark:bg-slate-800">

          <!-- Header -->
          <div class="flex items-center justify-between border-b border-slate-200 px-6 py-4 dark:border-slate-700">
            <div class="flex items-center gap-3">
              <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-600">
                <Package class="h-5 w-5 text-white" />
              </div>
              <div>
                <h2 class="text-lg font-semibold text-slate-900 dark:text-white">New Goods Receipt</h2>
                <p class="text-xs text-slate-500 dark:text-slate-400">Record incoming shipment against a purchase order</p>
              </div>
            </div>
            <button
              @click="showForm = false"
              class="rounded-lg p-2 text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors dark:hover:bg-slate-700"
            >
              <X class="h-5 w-5" />
            </button>
          </div>

          <!-- Body -->
          <div class="space-y-6 p-6">

            <!-- PO + Date row -->
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  Purchase Order <span class="text-red-500">*</span>
                </label>
                <select
                  v-model="form.po_id"
                  @change="onPOSelect"
                  class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                >
                  <option value="">Select an approved PO...</option>
                  <option v-for="o in orders" :key="o.id" :value="o.id">
                    {{ o.number }} — {{ o.supplier_name }}
                  </option>
                </select>
                <p v-if="!orders.length" class="mt-1 text-xs text-amber-600 dark:text-amber-400">
                  No approved purchase orders available
                </p>
              </div>
              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  Receipt Date <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="form.date"
                  type="date"
                  class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                />
              </div>
            </div>

            <!-- Warehouse + Notes -->
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  Destination Warehouse
                </label>
                <input
                  v-model="form.warehouse_id"
                  type="text"
                  placeholder="Warehouse ID (optional)"
                  class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-700 dark:text-white dark:placeholder-slate-500"
                />
              </div>
              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  Notes
                </label>
                <input
                  v-model="form.notes"
                  type="text"
                  placeholder="Optional remarks"
                  class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-700 dark:text-white dark:placeholder-slate-500"
                />
              </div>
            </div>

            <!-- Selected PO info -->
            <div v-if="selectedOrder" class="rounded-lg bg-blue-50 px-4 py-3 dark:bg-blue-900/20">
              <div class="flex items-start gap-2">
                <Building2 class="mt-0.5 h-4 w-4 text-blue-600 dark:text-blue-400 shrink-0" />
                <div class="text-sm text-blue-800 dark:text-blue-300">
                  <span class="font-semibold">PO {{ selectedOrder.number }}</span>
                  &mdash; {{ selectedOrder.supplier_name }}
                  &mdash; Total: <span class="font-semibold">{{ fmt(selectedOrder.total_amount) }} DZD</span>
                </div>
              </div>
            </div>

            <!-- Lines Table -->
            <div>
              <div class="mb-3 flex items-center justify-between">
                <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300">Receipt Lines</h3>
                <button
                  @click="addLine"
                  class="inline-flex items-center gap-1 rounded-lg border border-slate-200 px-3 py-1.5 text-xs font-medium text-slate-600 hover:bg-slate-50 transition-colors dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
                >
                  <Plus class="h-3.5 w-3.5" /> Add Line
                </button>
              </div>

              <div class="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-700">
                <table class="w-full text-sm">
                  <thead>
                    <tr class="border-b border-slate-200 bg-slate-50 dark:border-slate-700 dark:bg-slate-900/50">
                      <th class="px-3 py-2 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">Description</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 w-24">Exp. Qty</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 w-28">Received Qty</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 w-32">Unit Cost</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 w-32">Line Total</th>
                      <th class="w-10"></th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
                    <tr v-for="(line, i) in form.lines" :key="i">
                      <td class="px-3 py-2">
                        <input
                          v-model="line.description"
                          type="text"
                          placeholder="Item / service description"
                          class="w-full min-w-[180px] rounded border border-slate-200 bg-white px-2 py-1.5 text-xs text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                        />
                      </td>
                      <td class="px-3 py-2">
                        <input
                          v-model.number="line.expected_qty"
                          type="number"
                          min="0"
                          step="0.001"
                          class="w-full rounded border border-slate-200 bg-white px-2 py-1.5 text-right text-xs text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                        />
                      </td>
                      <td class="px-3 py-2">
                        <input
                          v-model.number="line.received_qty"
                          type="number"
                          min="0"
                          step="0.001"
                          class="w-full rounded border border-slate-200 bg-white px-2 py-1.5 text-right text-xs font-semibold text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                        />
                      </td>
                      <td class="px-3 py-2">
                        <input
                          v-model.number="line.unit_cost"
                          type="number"
                          min="0"
                          step="0.01"
                          class="w-full rounded border border-slate-200 bg-white px-2 py-1.5 text-right text-xs text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                        />
                      </td>
                      <td class="px-3 py-2 text-right text-xs font-semibold text-slate-700 dark:text-slate-300">
                        {{ fmt(line.received_qty * line.unit_cost) }}
                      </td>
                      <td class="px-3 py-2 text-center">
                        <button
                          @click="removeLine(i)"
                          :disabled="form.lines.length === 1"
                          class="rounded p-1 text-slate-300 hover:text-red-500 disabled:opacity-30 transition-colors"
                        >
                          <Trash2 class="h-3.5 w-3.5" />
                        </button>
                      </td>
                    </tr>
                  </tbody>
                  <tfoot>
                    <tr class="border-t-2 border-slate-200 dark:border-slate-600">
                      <td colspan="4" class="px-3 py-2.5 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                        Total Received Value
                      </td>
                      <td class="px-3 py-2.5 text-right text-sm font-bold text-slate-900 dark:text-white">
                        {{ fmt(formTotal) }} <span class="text-xs font-normal text-slate-400">DZD</span>
                      </td>
                      <td></td>
                    </tr>
                  </tfoot>
                </table>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="flex items-center justify-end gap-3 border-t border-slate-200 px-6 py-4 dark:border-slate-700">
            <button
              @click="showForm = false"
              class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50 transition-colors dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
            >
              Cancel
            </button>
            <button
              @click="saveReceipt"
              :disabled="saving || !form.po_id"
              class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-700 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
            >
              <Save v-if="!saving" class="h-4 w-4" />
              <div v-else class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></div>
              {{ saving ? 'Saving...' : 'Create Receipt' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Detail Modal ────────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showDetail && detailReceipt" class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/60 p-4 backdrop-blur-sm">
        <div class="my-6 w-full max-w-3xl rounded-2xl bg-white shadow-2xl dark:bg-slate-800">

          <!-- Header -->
          <div class="flex items-start justify-between border-b border-slate-200 px-6 py-4 dark:border-slate-700">
            <div>
              <div class="flex items-center gap-3">
                <span class="font-mono text-lg font-bold text-blue-600 dark:text-blue-400">{{ detailReceipt.number }}</span>
                <span :class="['inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium', STATUS_BADGE[detailReceipt.status] ?? STATUS_BADGE.draft]">
                  {{ STATUS_LABEL[detailReceipt.status] ?? detailReceipt.status }}
                </span>
              </div>
              <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
                {{ detailReceipt.supplier_name }} &mdash; {{ fmtDate(detailReceipt.date) }}
              </p>
            </div>
            <button
              @click="showDetail = false"
              class="rounded-lg p-2 text-slate-400 hover:bg-slate-100 transition-colors dark:hover:bg-slate-700"
            >
              <X class="h-5 w-5" />
            </button>
          </div>

          <!-- Info Grid -->
          <div class="grid grid-cols-2 gap-4 border-b border-slate-100 p-6 sm:grid-cols-3 dark:border-slate-700">
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Purchase Order</p>
              <p class="mt-1 font-mono text-sm font-semibold text-slate-700 dark:text-slate-300">
                {{ detailReceipt.po_number || '—' }}
              </p>
            </div>
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Supplier</p>
              <p class="mt-1 text-sm font-semibold text-slate-700 dark:text-slate-300">{{ detailReceipt.supplier_name }}</p>
            </div>
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Date</p>
              <p class="mt-1 text-sm text-slate-700 dark:text-slate-300">{{ fmtDate(detailReceipt.date) }}</p>
            </div>
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Warehouse</p>
              <p class="mt-1 text-sm text-slate-700 dark:text-slate-300">{{ detailReceipt.warehouse_id || '—' }}</p>
            </div>
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Total Value</p>
              <p class="mt-1 text-base font-bold text-slate-900 dark:text-white">{{ fmt(detailReceipt.total_amount) }} <span class="text-xs font-normal text-slate-400">DZD</span></p>
            </div>
            <div v-if="detailReceipt.notes">
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Notes</p>
              <p class="mt-1 text-sm text-slate-700 dark:text-slate-300">{{ detailReceipt.notes }}</p>
            </div>
          </div>

          <!-- Lines -->
          <div v-if="detailReceipt.lines?.length" class="p-6">
            <h3 class="mb-3 text-sm font-semibold text-slate-700 dark:text-slate-300">Receipt Lines</h3>
            <div class="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-700">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-slate-200 bg-slate-50 dark:border-slate-700 dark:bg-slate-900/50">
                    <th class="px-3 py-2 text-left text-xs font-semibold text-slate-500 dark:text-slate-400">Description</th>
                    <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400">Expected</th>
                    <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400">Received</th>
                    <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400">Unit Cost</th>
                    <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400">Line Total</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
                  <tr v-for="(line, i) in detailReceipt.lines" :key="i" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
                    <td class="px-3 py-2 text-slate-700 dark:text-slate-300">{{ line.description }}</td>
                    <td class="px-3 py-2 text-right text-slate-500 dark:text-slate-400">{{ line.expected_qty }}</td>
                    <td class="px-3 py-2 text-right">
                      <span :class="[
                        'font-semibold',
                        line.received_qty >= line.expected_qty
                          ? 'text-green-600 dark:text-green-400'
                          : 'text-amber-600 dark:text-amber-400'
                      ]">
                        {{ line.received_qty }}
                      </span>
                    </td>
                    <td class="px-3 py-2 text-right text-slate-600 dark:text-slate-400">{{ fmt(line.unit_cost) }}</td>
                    <td class="px-3 py-2 text-right font-semibold text-slate-900 dark:text-white">
                      {{ fmt(line.received_qty * line.unit_cost) }}
                    </td>
                  </tr>
                </tbody>
                <tfoot>
                  <tr class="border-t-2 border-slate-200 dark:border-slate-600">
                    <td colspan="4" class="px-3 py-2.5 text-right text-xs font-semibold text-slate-500 uppercase tracking-wider">
                      Total
                    </td>
                    <td class="px-3 py-2.5 text-right text-sm font-bold text-slate-900 dark:text-white">
                      {{ fmt(detailReceipt.total_amount) }} <span class="text-xs font-normal text-slate-400">DZD</span>
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </div>

          <!-- Footer Actions -->
          <div class="flex items-center justify-between border-t border-slate-200 px-6 py-4 dark:border-slate-700">
            <div>
              <button
                v-if="detailReceipt.status === 'draft'"
                @click="validateReceipt(detailReceipt.id)"
                class="inline-flex items-center gap-2 rounded-lg bg-green-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-green-700 transition-colors"
              >
                <CheckCircle class="h-4 w-4" />
                Validate Receipt
              </button>
              <div v-else class="flex items-center gap-2 text-sm text-green-600 dark:text-green-400">
                <CheckCircle class="h-4 w-4" />
                <span class="font-medium">Receipt validated</span>
              </div>
            </div>
            <button
              @click="showDetail = false"
              class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50 transition-colors dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
