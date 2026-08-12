<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { salesAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  Plus, X, Save, Search, RefreshCw, Edit2, Eye, FileText,
  ChevronDown, ChevronUp, Trash2, Send, ArrowRight,
  CheckCircle, XCircle, Clock, Hash, Calendar, User, Building2,
  DollarSign, Percent, Package
} from 'lucide-vue-next'

// ─── Types ──────────────────────────────────────────────────────────────────

interface DocumentLine {
  id?: string
  item_id?: string
  description: string
  quantity: number
  unit_price: number
  discount_pct: number
  tva_rate: number
  tva_amount?: number
  total?: number
  account_id?: string
  sort_order: number
}

interface Quotation {
  id: string
  company_id: string
  number: string
  customer_id: string
  customer_name?: string
  date: string
  valid_until?: string
  status: string
  subtotal: number
  discount_amount: number
  tva_amount: number
  stamp_tax: number
  total_amount: number
  currency: string
  notes: string
  terms: string
  salesperson_id?: string
  converted_to?: string
  lines?: DocumentLine[]
  created_at: string
  updated_at: string
}

interface Customer { id: string; name: string; code: string; payment_terms: number }

const EMPTY_LINE = (sort = 0): DocumentLine => ({
  description: '', quantity: 1, unit_price: 0, discount_pct: 0, tva_rate: 19, sort_order: sort,
})

const STATUS_BADGE: Record<string, string> = {
  draft:     'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  sent:      'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  confirmed: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/40 dark:text-indigo-300',
  converted: 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300',
  cancelled: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300',
  expired:   'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300',
}

// ─── State ───────────────────────────────────────────────────────────────────

const app = useAppStore()
const quotations = ref<Quotation[]>([])
const customers = ref<Customer[]>([])
const loading = ref(true)
const saving = ref(false)

// Modal — list vs form
const showForm = ref(false)
const showDetail = ref(false)
const isEdit = ref(false)
const editingQuot = ref<Partial<Quotation> & { lines: DocumentLine[] }>({
  number: '', customer_id: '', date: today(), valid_until: '', status: 'draft',
  notes: '', terms: '', currency: 'DZD', lines: [EMPTY_LINE(0)],
})
const detailQuot = ref<Quotation | null>(null)

const search = ref('')
const filterStatus = ref('')
const sortBy = ref<keyof Quotation>('date')
const sortDir = ref<'asc' | 'desc'>('desc')

// ─── Computed ────────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = [...quotations.value]
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(qt =>
      qt.number.toLowerCase().includes(q) ||
      qt.customer_name?.toLowerCase().includes(q)
    )
  }
  if (filterStatus.value) list = list.filter(qt => qt.status === filterStatus.value)
  list.sort((a, b) => {
    const av = String(a[sortBy.value] ?? '')
    const bv = String(b[sortBy.value] ?? '')
    return sortDir.value === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
  })
  return list
})

const lineTotals = computed(() => {
  let subtotal = 0, tva = 0
  for (const l of editingQuot.value.lines) {
    const sub = l.quantity * l.unit_price * (1 - l.discount_pct / 100)
    const tvAmt = sub * (l.tva_rate / 100)
    subtotal += sub
    tva += tvAmt
  }
  const total = subtotal + tva + (editingQuot.value.stamp_tax || 0)
  return { subtotal: round2(subtotal), tva: round2(tva), total: round2(total) }
})

// ─── Helpers ─────────────────────────────────────────────────────────────────

function today() {
  return new Date().toISOString().slice(0, 10)
}

function round2(n: number) {
  return Math.round(n * 100) / 100
}

function fmtCurrency(n?: number) {
  if (n == null) return '—'
  return n.toLocaleString('fr-DZ', { minimumFractionDigits: 2 }) + ' DZD'
}

function fmtDate(d?: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('fr-DZ')
}

// ─── Data ─────────────────────────────────────────────────────────────────────

async function loadData() {
  loading.value = true
  try {
    const [qRes, cRes] = await Promise.allSettled([
      salesAPI.getQuotations(),
      salesAPI.getCustomers(),
    ])
    if (qRes.status === 'fulfilled') quotations.value = qRes.value.data || []
    if (cRes.status === 'fulfilled') customers.value = cRes.value.data || []
  } catch {
    app.addToast('Failed to load quotations', 'error')
  } finally {
    loading.value = false
  }
}

// ─── Form ─────────────────────────────────────────────────────────────────────

function openCreate() {
  isEdit.value = false
  editingQuot.value = {
    number: '', customer_id: '', date: today(), valid_until: '',
    status: 'draft', notes: '', terms: '', currency: 'DZD',
    stamp_tax: 0, lines: [EMPTY_LINE(0)],
  }
  showForm.value = true
}

function openEdit(q: Quotation) {
  isEdit.value = true
  editingQuot.value = {
    ...q,
    lines: q.lines?.length ? [...q.lines] : [EMPTY_LINE(0)],
  }
  showForm.value = true
}

async function openDetail(q: Quotation) {
  try {
    const res = await salesAPI.getQuotation(q.id)
    detailQuot.value = res.data
    showDetail.value = true
  } catch {
    app.addToast('Failed to load quotation detail', 'error')
  }
}

function closeForm() { showForm.value = false }
function closeDetail() { showDetail.value = false }

// Lines management
function addLine() {
  editingQuot.value.lines.push(EMPTY_LINE(editingQuot.value.lines.length))
}

function removeLine(idx: number) {
  if (editingQuot.value.lines.length === 1) return
  editingQuot.value.lines.splice(idx, 1)
}

function lineSubtotal(l: DocumentLine) {
  return round2(l.quantity * l.unit_price * (1 - l.discount_pct / 100))
}

function lineTVA(l: DocumentLine) {
  return round2(lineSubtotal(l) * (l.tva_rate / 100))
}

function lineTotal(l: DocumentLine) {
  return round2(lineSubtotal(l) + lineTVA(l))
}

async function save() {
  const q = editingQuot.value
  if (!q.customer_id) { app.addToast('Customer is required', 'error'); return }
  if (!q.date) { app.addToast('Date is required', 'error'); return }
  if (!q.lines.length || !q.lines[0].description) { app.addToast('At least one line required', 'error'); return }

  // Inject computed totals into lines
  const lines = q.lines.map((l, i) => ({
    ...l, sort_order: i,
    tva_amount: lineTVA(l),
    total: lineTotal(l),
  }))

  const payload = {
    ...q,
    subtotal: lineTotals.value.subtotal,
    tva_amount: lineTotals.value.tva,
    total_amount: lineTotals.value.total,
    lines,
  }

  saving.value = true
  try {
    if (isEdit.value && q.id) {
      await salesAPI.updateQuotation(q.id, payload)
      app.addToast('Quotation updated', 'success')
    } else {
      await salesAPI.createQuotation(payload)
      app.addToast('Quotation created', 'success')
    }
    closeForm()
    await loadData()
  } catch {
    app.addToast('Failed to save quotation', 'error')
  } finally {
    saving.value = false
  }
}

async function confirmQuot(q: Quotation) {
  try {
    await salesAPI.confirmQuotation(q.id)
    app.addToast('Quotation confirmed (sent)', 'success')
    await loadData()
  } catch {
    app.addToast('Failed to confirm quotation', 'error')
  }
}

async function convertToOrder(q: Quotation) {
  if (!confirm(`Convert quotation ${q.number} to a Sales Order?`)) return
  try {
    await salesAPI.convertToOrder(q.id)
    app.addToast('Quotation converted to Sales Order', 'success')
    await loadData()
  } catch {
    app.addToast('Failed to convert quotation', 'error')
  }
}

async function cancelQuot(q: Quotation) {
  if (!confirm(`Cancel quotation ${q.number}?`)) return
  try {
    await salesAPI.cancelQuotation(q.id)
    app.addToast('Quotation cancelled', 'success')
    await loadData()
  } catch {
    app.addToast('Failed to cancel quotation', 'error')
  }
}

function toggleSort(col: keyof Quotation) {
  if (sortBy.value === col) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortBy.value = col; sortDir.value = 'asc' }
}

onMounted(loadData)
</script>

<template>
  <div class="space-y-5">

    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Quotations</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">{{ quotations.length }} quotations</p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="loadData" :disabled="loading" class="p-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors disabled:opacity-50">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button @click="openCreate" class="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg shadow-sm transition-colors">
          <Plus class="w-4 h-4" /> New Quotation
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3">
      <div class="relative flex-1 min-w-48">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <input v-model="search" placeholder="Search number, customer…" class="w-full pl-9 pr-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 outline-none" />
      </div>
      <select v-model="filterStatus" class="px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 outline-none">
        <option value="">All Statuses</option>
        <option value="draft">Draft</option>
        <option value="sent">Sent</option>
        <option value="confirmed">Confirmed</option>
        <option value="converted">Converted</option>
        <option value="cancelled">Cancelled</option>
        <option value="expired">Expired</option>
      </select>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center h-40">
        <RefreshCw class="w-6 h-6 text-gray-400 animate-spin" />
      </div>
      <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center h-40 gap-2 text-gray-400 dark:text-gray-600">
        <FileText class="w-8 h-8" />
        <p class="text-sm">No quotations found</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <tr>
              <th @click="toggleSort('number')" class="th cursor-pointer">
                <div class="flex items-center gap-1">Number <ChevronUp v-if="sortBy==='number'&&sortDir==='asc'" class="w-3 h-3" /><ChevronDown v-else-if="sortBy==='number'" class="w-3 h-3" /></div>
              </th>
              <th class="th">Customer</th>
              <th @click="toggleSort('date')" class="th cursor-pointer">
                <div class="flex items-center gap-1">Date <ChevronUp v-if="sortBy==='date'&&sortDir==='asc'" class="w-3 h-3" /><ChevronDown v-else-if="sortBy==='date'" class="w-3 h-3" /></div>
              </th>
              <th class="th">Valid Until</th>
              <th @click="toggleSort('total_amount')" class="th text-right cursor-pointer">
                <div class="flex items-center justify-end gap-1">Total <ChevronUp v-if="sortBy==='total_amount'&&sortDir==='asc'" class="w-3 h-3" /><ChevronDown v-else-if="sortBy==='total_amount'" class="w-3 h-3" /></div>
              </th>
              <th class="th text-center">Status</th>
              <th class="th" />
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr v-for="q in filtered" :key="q.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
              <td class="px-4 py-3 font-mono text-sm font-semibold text-gray-800 dark:text-gray-200">{{ q.number }}</td>
              <td class="px-4 py-3 text-gray-700 dark:text-gray-300">{{ q.customer_name || '—' }}</td>
              <td class="px-4 py-3 text-gray-600 dark:text-gray-400">{{ fmtDate(q.date) }}</td>
              <td class="px-4 py-3 text-gray-600 dark:text-gray-400">{{ fmtDate(q.valid_until) }}</td>
              <td class="px-4 py-3 text-right font-semibold text-gray-900 dark:text-white">{{ fmtCurrency(q.total_amount) }}</td>
              <td class="px-4 py-3 text-center">
                <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium capitalize" :class="STATUS_BADGE[q.status]">
                  {{ q.status }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1 justify-end">
                  <button @click="openDetail(q)" class="icon-btn" title="View"><Eye class="w-3.5 h-3.5" /></button>
                  <button v-if="q.status === 'draft'" @click="openEdit(q)" class="icon-btn text-blue-600 dark:text-blue-400" title="Edit"><Edit2 class="w-3.5 h-3.5" /></button>
                  <button v-if="q.status === 'draft'" @click="confirmQuot(q)" class="icon-btn text-indigo-600 dark:text-indigo-400" title="Send / Confirm"><Send class="w-3.5 h-3.5" /></button>
                  <button v-if="['draft','sent','confirmed'].includes(q.status)" @click="convertToOrder(q)" class="icon-btn text-green-600 dark:text-green-400" title="Convert to Order"><ArrowRight class="w-3.5 h-3.5" /></button>
                  <button v-if="['draft','sent'].includes(q.status)" @click="cancelQuot(q)" class="icon-btn text-red-500 dark:text-red-400" title="Cancel"><XCircle class="w-3.5 h-3.5" /></button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ─── Create / Edit Form Modal ─────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showForm" class="fixed inset-0 z-50 flex items-start justify-center p-4 bg-black/50 backdrop-blur-sm overflow-y-auto" @click.self="closeForm">
        <div class="w-full max-w-4xl bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 my-8">

          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div class="flex items-center gap-2">
              <FileText class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ isEdit ? 'Edit Quotation' : 'New Quotation' }}</h2>
            </div>
            <button @click="closeForm" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"><X class="w-5 h-5" /></button>
          </div>

          <div class="p-6 space-y-6">
            <!-- Header fields -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div class="md:col-span-2">
                <label class="label">Customer <span class="text-red-500">*</span></label>
                <select v-model="editingQuot.customer_id" class="input">
                  <option value="">— Select Customer —</option>
                  <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
              </div>
              <div>
                <label class="label">Number</label>
                <input v-model="editingQuot.number" class="input font-mono" placeholder="QT-2025-001" />
              </div>
              <div>
                <label class="label">Currency</label>
                <select v-model="editingQuot.currency" class="input">
                  <option value="DZD">DZD</option>
                  <option value="EUR">EUR</option>
                  <option value="USD">USD</option>
                </select>
              </div>
              <div>
                <label class="label">Date <span class="text-red-500">*</span></label>
                <input v-model="editingQuot.date" type="date" class="input" />
              </div>
              <div>
                <label class="label">Valid Until</label>
                <input v-model="editingQuot.valid_until" type="date" class="input" />
              </div>
              <div>
                <label class="label">Stamp Tax (DZD)</label>
                <input v-model.number="editingQuot.stamp_tax" type="number" min="0" step="100" class="input" />
              </div>
            </div>

            <!-- Lines -->
            <div>
              <div class="flex items-center justify-between mb-3">
                <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Line Items</h3>
                <button @click="addLine" class="inline-flex items-center gap-1.5 text-xs text-blue-600 dark:text-blue-400 hover:underline">
                  <Plus class="w-3.5 h-3.5" /> Add Line
                </button>
              </div>
              <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
                <table class="w-full text-xs">
                  <thead class="bg-gray-50 dark:bg-gray-800">
                    <tr>
                      <th class="px-3 py-2 text-left font-semibold text-gray-500 dark:text-gray-400 w-64">Description</th>
                      <th class="px-3 py-2 text-right font-semibold text-gray-500 dark:text-gray-400 w-20">Qty</th>
                      <th class="px-3 py-2 text-right font-semibold text-gray-500 dark:text-gray-400 w-28">Unit Price</th>
                      <th class="px-3 py-2 text-right font-semibold text-gray-500 dark:text-gray-400 w-20">Disc %</th>
                      <th class="px-3 py-2 text-right font-semibold text-gray-500 dark:text-gray-400 w-20">TVA %</th>
                      <th class="px-3 py-2 text-right font-semibold text-gray-500 dark:text-gray-400 w-28">Subtotal HT</th>
                      <th class="px-3 py-2 text-right font-semibold text-gray-500 dark:text-gray-400 w-24">TVA</th>
                      <th class="px-3 py-2 text-right font-semibold text-gray-500 dark:text-gray-400 w-28">Total TTC</th>
                      <th class="px-3 py-2 w-8" />
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                    <tr v-for="(line, idx) in editingQuot.lines" :key="idx">
                      <td class="px-2 py-1.5">
                        <input v-model="line.description" class="line-input" placeholder="Product / Service description" />
                      </td>
                      <td class="px-2 py-1.5">
                        <input v-model.number="line.quantity" type="number" min="0.001" step="1" class="line-input text-right" />
                      </td>
                      <td class="px-2 py-1.5">
                        <input v-model.number="line.unit_price" type="number" min="0" step="100" class="line-input text-right" />
                      </td>
                      <td class="px-2 py-1.5">
                        <input v-model.number="line.discount_pct" type="number" min="0" max="100" step="1" class="line-input text-right" />
                      </td>
                      <td class="px-2 py-1.5">
                        <input v-model.number="line.tva_rate" type="number" min="0" max="100" step="1" class="line-input text-right" />
                      </td>
                      <td class="px-2 py-1.5 text-right font-medium text-gray-800 dark:text-gray-200 tabular-nums">
                        {{ lineSubtotal(line).toLocaleString('fr-DZ', { minimumFractionDigits: 2 }) }}
                      </td>
                      <td class="px-2 py-1.5 text-right text-gray-600 dark:text-gray-400 tabular-nums">
                        {{ lineTVA(line).toLocaleString('fr-DZ', { minimumFractionDigits: 2 }) }}
                      </td>
                      <td class="px-2 py-1.5 text-right font-bold text-gray-900 dark:text-white tabular-nums">
                        {{ lineTotal(line).toLocaleString('fr-DZ', { minimumFractionDigits: 2 }) }}
                      </td>
                      <td class="px-2 py-1.5 text-center">
                        <button @click="removeLine(idx)" class="text-red-400 hover:text-red-600 dark:hover:text-red-300 transition-colors" :disabled="editingQuot.lines.length === 1">
                          <Trash2 class="w-3.5 h-3.5" />
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Totals + Notes -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <label class="label">Notes</label>
                <textarea v-model="editingQuot.notes" rows="3" class="input resize-none" placeholder="Customer-facing notes…" />
                <label class="label mt-3">Terms & Conditions</label>
                <textarea v-model="editingQuot.terms" rows="2" class="input resize-none" placeholder="Payment terms, validity…" />
              </div>
              <div class="bg-gray-50 dark:bg-gray-800 rounded-xl p-4 space-y-2">
                <div class="flex justify-between text-sm text-gray-600 dark:text-gray-400">
                  <span>Subtotal HT</span>
                  <span class="font-medium text-gray-800 dark:text-gray-200 tabular-nums">{{ fmtCurrency(lineTotals.subtotal) }}</span>
                </div>
                <div class="flex justify-between text-sm text-gray-600 dark:text-gray-400">
                  <span>TVA (19%)</span>
                  <span class="font-medium text-gray-800 dark:text-gray-200 tabular-nums">{{ fmtCurrency(lineTotals.tva) }}</span>
                </div>
                <div v-if="editingQuot.stamp_tax" class="flex justify-between text-sm text-gray-600 dark:text-gray-400">
                  <span>Stamp Tax</span>
                  <span class="font-medium text-gray-800 dark:text-gray-200 tabular-nums">{{ fmtCurrency(editingQuot.stamp_tax) }}</span>
                </div>
                <div class="flex justify-between text-base font-bold text-gray-900 dark:text-white border-t border-gray-200 dark:border-gray-700 pt-2">
                  <span>Total TTC</span>
                  <span class="tabular-nums">{{ fmtCurrency(lineTotals.total) }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
            <button @click="closeForm" class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
            <button @click="save" :disabled="saving" class="inline-flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg shadow-sm transition-colors disabled:opacity-60">
              <Save class="w-4 h-4" />{{ saving ? 'Saving…' : (isEdit ? 'Update' : 'Create Quotation') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ─── Detail Modal ──────────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showDetail && detailQuot" class="fixed inset-0 z-50 flex items-start justify-center p-4 bg-black/50 backdrop-blur-sm overflow-y-auto" @click.self="closeDetail">
        <div class="w-full max-w-3xl bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 my-8">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div class="flex items-center gap-3">
              <FileText class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white font-mono">{{ detailQuot.number }}</h2>
              <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium capitalize" :class="STATUS_BADGE[detailQuot.status]">{{ detailQuot.status }}</span>
            </div>
            <button @click="closeDetail" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"><X class="w-5 h-5" /></button>
          </div>
          <div class="p-6 space-y-5">
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div><p class="text-gray-400 dark:text-gray-500 text-xs">Customer</p><p class="font-semibold text-gray-800 dark:text-gray-200">{{ detailQuot.customer_name || '—' }}</p></div>
              <div><p class="text-gray-400 dark:text-gray-500 text-xs">Date</p><p class="font-semibold text-gray-800 dark:text-gray-200">{{ fmtDate(detailQuot.date) }}</p></div>
              <div><p class="text-gray-400 dark:text-gray-500 text-xs">Valid Until</p><p class="font-semibold text-gray-800 dark:text-gray-200">{{ fmtDate(detailQuot.valid_until) }}</p></div>
              <div><p class="text-gray-400 dark:text-gray-500 text-xs">Currency</p><p class="font-semibold text-gray-800 dark:text-gray-200">{{ detailQuot.currency }}</p></div>
            </div>
            <!-- Lines table -->
            <div v-if="detailQuot.lines?.length" class="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
              <table class="w-full text-xs">
                <thead class="bg-gray-50 dark:bg-gray-800">
                  <tr>
                    <th class="px-3 py-2 text-left font-semibold text-gray-500 dark:text-gray-400">Description</th>
                    <th class="px-3 py-2 text-right font-semibold text-gray-500 dark:text-gray-400">Qty</th>
                    <th class="px-3 py-2 text-right font-semibold text-gray-500 dark:text-gray-400">Unit Price</th>
                    <th class="px-3 py-2 text-right font-semibold text-gray-500 dark:text-gray-400">TVA</th>
                    <th class="px-3 py-2 text-right font-semibold text-gray-500 dark:text-gray-400">Total TTC</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                  <tr v-for="l in detailQuot.lines" :key="l.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                    <td class="px-3 py-2 text-gray-700 dark:text-gray-300">{{ l.description }}</td>
                    <td class="px-3 py-2 text-right text-gray-700 dark:text-gray-300 tabular-nums">{{ l.quantity }}</td>
                    <td class="px-3 py-2 text-right text-gray-700 dark:text-gray-300 tabular-nums">{{ fmtCurrency(l.unit_price) }}</td>
                    <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-400 tabular-nums">{{ l.tva_rate }}%</td>
                    <td class="px-3 py-2 text-right font-semibold text-gray-900 dark:text-white tabular-nums">{{ fmtCurrency(l.total) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <!-- Totals -->
            <div class="flex justify-end">
              <div class="w-64 space-y-1.5 text-sm">
                <div class="flex justify-between text-gray-600 dark:text-gray-400"><span>Subtotal HT</span><span class="tabular-nums">{{ fmtCurrency(detailQuot.subtotal) }}</span></div>
                <div class="flex justify-between text-gray-600 dark:text-gray-400"><span>TVA</span><span class="tabular-nums">{{ fmtCurrency(detailQuot.tva_amount) }}</span></div>
                <div v-if="detailQuot.stamp_tax" class="flex justify-between text-gray-600 dark:text-gray-400"><span>Stamp Tax</span><span class="tabular-nums">{{ fmtCurrency(detailQuot.stamp_tax) }}</span></div>
                <div class="flex justify-between font-bold text-gray-900 dark:text-white border-t border-gray-200 dark:border-gray-700 pt-1.5"><span>Total TTC</span><span class="tabular-nums">{{ fmtCurrency(detailQuot.total_amount) }}</span></div>
              </div>
            </div>
            <!-- Actions -->
            <div class="flex flex-wrap gap-2 pt-2 border-t border-gray-200 dark:border-gray-700">
              <button v-if="detailQuot.status === 'draft'" @click="confirmQuot(detailQuot); closeDetail()" class="btn-secondary inline-flex items-center gap-1.5"><Send class="w-4 h-4" /> Confirm / Send</button>
              <button v-if="['draft','sent','confirmed'].includes(detailQuot.status)" @click="convertToOrder(detailQuot); closeDetail()" class="btn-primary inline-flex items-center gap-1.5"><ArrowRight class="w-4 h-4" /> Convert to Order</button>
              <button v-if="['draft','sent'].includes(detailQuot.status)" @click="cancelQuot(detailQuot); closeDetail()" class="btn-danger inline-flex items-center gap-1.5"><XCircle class="w-4 h-4" /> Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<style scoped>
.label { @apply block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5; }
.input { @apply w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500 outline-none; }
.line-input { @apply w-full px-2 py-1 rounded border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-xs focus:ring-1 focus:ring-blue-500 outline-none; }
.th { @apply px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide whitespace-nowrap; }
.icon-btn { @apply p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 transition-colors; }
.btn-primary { @apply px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors; }
.btn-secondary { @apply px-4 py-2 border border-indigo-200 dark:border-indigo-700 text-indigo-700 dark:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 text-sm font-medium rounded-lg transition-colors; }
.btn-danger { @apply px-4 py-2 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 text-sm font-medium rounded-lg transition-colors; }
</style>
