<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { salesAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  Plus, X, Save, Search, RefreshCw, Edit2, Eye, ReceiptText,
  ChevronDown, ChevronUp, Trash2, CheckCircle, XCircle,
  CreditCard, DollarSign, AlertTriangle, Clock, Calendar,
  Building2, Wallet
} from 'lucide-vue-next'

// ─── Types ───────────────────────────────────────────────────────────────────

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
  sort_order: number
}

interface SalesInvoice {
  id: string
  company_id: string
  number: string
  order_id?: string
  customer_id: string
  customer_name?: string
  date: string
  due_date?: string
  status: string
  subtotal: number
  discount_amount: number
  tva_amount: number
  stamp_tax: number
  total_amount: number
  paid_amount: number
  balance_due: number
  currency: string
  notes: string
  journal_entry_id?: string
  lines?: DocumentLine[]
  created_at: string
  updated_at: string
}

interface Customer { id: string; name: string; code: string; payment_terms: number }

const EMPTY_LINE = (sort = 0): DocumentLine => ({
  description: '', quantity: 1, unit_price: 0, discount_pct: 0, tva_rate: 19, sort_order: sort,
})

const STATUS_BADGE: Record<string, string> = {
  draft:          'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  confirmed:      'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  partially_paid: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  paid:           'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300',
  cancelled:      'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300',
  overdue:        'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300',
}

// ─── State ───────────────────────────────────────────────────────────────────

const app = useAppStore()
const invoices = ref<SalesInvoice[]>([])
const customers = ref<Customer[]>([])
const loading = ref(true)
const saving = ref(false)

const showForm = ref(false)
const showDetail = ref(false)
const showPaymentModal = ref(false)
const isEdit = ref(false)

const editingInvoice = ref<Partial<SalesInvoice> & { lines: DocumentLine[] }>({
  number: '', customer_id: '', date: today(), due_date: '', status: 'draft',
  notes: '', currency: 'DZD', stamp_tax: 0, lines: [EMPTY_LINE()],
})

const detailInvoice = ref<SalesInvoice | null>(null)

const paymentForm = ref({ amount: 0, method: 'bank', payment_date: today(), notes: '' })
const payingInvoice = ref<SalesInvoice | null>(null)

const search = ref('')
const filterStatus = ref('')
const sortBy = ref<keyof SalesInvoice>('date')
const sortDir = ref<'asc' | 'desc'>('desc')

// ─── Computed ────────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = [...invoices.value]
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(i => i.number.toLowerCase().includes(q) || i.customer_name?.toLowerCase().includes(q))
  }
  if (filterStatus.value) list = list.filter(i => i.status === filterStatus.value)
  list.sort((a, b) => {
    const av = String(a[sortBy.value] ?? ''); const bv = String(b[sortBy.value] ?? '')
    return sortDir.value === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
  })
  return list
})

const totalReceivables = computed(() => invoices.value.filter(i => !['paid','cancelled'].includes(i.status)).reduce((s, i) => s + (i.balance_due || 0), 0))
const totalOverdue = computed(() => invoices.value.filter(i => i.status === 'overdue').length)

const lineTotals = computed(() => {
  let subtotal = 0, tva = 0
  for (const l of editingInvoice.value.lines) {
    const sub = l.quantity * l.unit_price * (1 - l.discount_pct / 100)
    subtotal += sub; tva += sub * (l.tva_rate / 100)
  }
  const total = subtotal + tva + (editingInvoice.value.stamp_tax || 0)
  return { subtotal: round2(subtotal), tva: round2(tva), total: round2(total) }
})

// ─── Helpers ─────────────────────────────────────────────────────────────────

function today() { return new Date().toISOString().slice(0, 10) }
function round2(n: number) { return Math.round(n * 100) / 100 }
function fmtCurrency(n?: number | null) {
  if (n == null) return '—'
  return n.toLocaleString('fr-DZ', { minimumFractionDigits: 2 }) + ' DZD'
}
function fmtDate(d?: string) { if (!d) return '—'; return new Date(d).toLocaleDateString('fr-DZ') }
function lineSubtotal(l: DocumentLine) { return round2(l.quantity * l.unit_price * (1 - l.discount_pct / 100)) }
function lineTVA(l: DocumentLine) { return round2(lineSubtotal(l) * (l.tva_rate / 100)) }
function lineTotal(l: DocumentLine) { return round2(lineSubtotal(l) + lineTVA(l)) }

function dueDateColor(inv: SalesInvoice) {
  if (['paid', 'cancelled'].includes(inv.status)) return ''
  if (inv.status === 'overdue') return 'text-red-600 dark:text-red-400'
  if (!inv.due_date) return ''
  const daysLeft = (new Date(inv.due_date).getTime() - Date.now()) / 86400000
  if (daysLeft < 7) return 'text-orange-600 dark:text-orange-400'
  return ''
}

// ─── Data ─────────────────────────────────────────────────────────────────────

async function loadData() {
  loading.value = true
  try {
    const [iRes, cRes] = await Promise.allSettled([salesAPI.getInvoices(), salesAPI.getCustomers()])
    if (iRes.status === 'fulfilled') invoices.value = iRes.value.data || []
    if (cRes.status === 'fulfilled') customers.value = cRes.value.data || []
  } catch { app.addToast('Failed to load invoices', 'error') } finally { loading.value = false }
}

// ─── Form ─────────────────────────────────────────────────────────────────────

function openCreate() {
  isEdit.value = false
  editingInvoice.value = { number: '', customer_id: '', date: today(), due_date: dueDate30(), status: 'draft', notes: '', currency: 'DZD', stamp_tax: 0, lines: [EMPTY_LINE()] }
  showForm.value = true
}

function dueDate30() {
  const d = new Date(); d.setDate(d.getDate() + 30); return d.toISOString().slice(0, 10)
}

function closeForm() { showForm.value = false }
function closeDetail() { showDetail.value = false }
function closePayment() { showPaymentModal.value = false; payingInvoice.value = null }

async function openDetail(inv: SalesInvoice) {
  try {
    const res = await salesAPI.getInvoice(inv.id)
    detailInvoice.value = res.data
    showDetail.value = true
  } catch { app.addToast('Failed to load invoice detail', 'error') }
}

function openPayment(inv: SalesInvoice) {
  payingInvoice.value = inv
  paymentForm.value = { amount: inv.balance_due, method: 'bank', payment_date: today(), notes: '' }
  showPaymentModal.value = true
}

function addLine() { editingInvoice.value.lines.push(EMPTY_LINE(editingInvoice.value.lines.length)) }
function removeLine(idx: number) { if (editingInvoice.value.lines.length > 1) editingInvoice.value.lines.splice(idx, 1) }

async function save() {
  const inv = editingInvoice.value
  if (!inv.customer_id) { app.addToast('Customer is required', 'error'); return }
  const lines = inv.lines.map((l, i) => ({ ...l, sort_order: i, tva_amount: lineTVA(l), total: lineTotal(l) }))
  const payload = { ...inv, subtotal: lineTotals.value.subtotal, tva_amount: lineTotals.value.tva, total_amount: lineTotals.value.total, lines }
  saving.value = true
  try {
    await salesAPI.createInvoice(payload)
    app.addToast('Invoice created', 'success')
    closeForm(); await loadData()
  } catch { app.addToast('Failed to save invoice', 'error') } finally { saving.value = false }
}

async function confirmInvoice(inv: SalesInvoice) {
  try { await salesAPI.confirmInvoice(inv.id); app.addToast('Invoice confirmed and posted', 'success'); await loadData() }
  catch { app.addToast('Failed to confirm invoice', 'error') }
}

async function cancelInvoice(inv: SalesInvoice) {
  if (!confirm(`Cancel invoice ${inv.number}?`)) return
  try { await salesAPI.cancelInvoice(inv.id); app.addToast('Invoice cancelled', 'success'); await loadData() }
  catch { app.addToast('Failed to cancel invoice', 'error') }
}

async function recordPayment() {
  const inv = payingInvoice.value
  if (!inv) return
  if (paymentForm.value.amount <= 0) { app.addToast('Amount must be positive', 'error'); return }
  saving.value = true
  try {
    await salesAPI.recordPayment(inv.id, paymentForm.value)
    app.addToast('Payment recorded', 'success')
    closePayment(); await loadData()
  } catch { app.addToast('Failed to record payment', 'error') } finally { saving.value = false }
}

function toggleSort(col: keyof SalesInvoice) {
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
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Sales Invoices</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
          {{ invoices.length }} invoices · Receivables: {{ fmtCurrency(totalReceivables) }}
          <span v-if="totalOverdue > 0" class="text-red-500 dark:text-red-400 ml-1">· {{ totalOverdue }} overdue</span>
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="loadData" :disabled="loading" class="p-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors disabled:opacity-50">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button @click="openCreate" class="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg shadow-sm transition-colors">
          <Plus class="w-4 h-4" /> New Invoice
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
        <option value="confirmed">Confirmed</option>
        <option value="partially_paid">Partially Paid</option>
        <option value="paid">Paid</option>
        <option value="overdue">Overdue</option>
        <option value="cancelled">Cancelled</option>
      </select>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center h-40"><RefreshCw class="w-6 h-6 text-gray-400 animate-spin" /></div>
      <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center h-40 gap-2 text-gray-400 dark:text-gray-600">
        <ReceiptText class="w-8 h-8" />
        <p class="text-sm">No invoices found</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <tr>
              <th @click="toggleSort('number')" class="th cursor-pointer"><div class="flex items-center gap-1">Invoice # <ChevronUp v-if="sortBy==='number'&&sortDir==='asc'" class="w-3 h-3"/><ChevronDown v-else-if="sortBy==='number'" class="w-3 h-3"/></div></th>
              <th class="th">Customer</th>
              <th @click="toggleSort('date')" class="th cursor-pointer"><div class="flex items-center gap-1">Date <ChevronUp v-if="sortBy==='date'&&sortDir==='asc'" class="w-3 h-3"/><ChevronDown v-else-if="sortBy==='date'" class="w-3 h-3"/></div></th>
              <th class="th">Due Date</th>
              <th @click="toggleSort('total_amount')" class="th text-right cursor-pointer"><div class="flex items-center justify-end gap-1">Total TTC <ChevronUp v-if="sortBy==='total_amount'&&sortDir==='asc'" class="w-3 h-3"/><ChevronDown v-else-if="sortBy==='total_amount'" class="w-3 h-3"/></div></th>
              <th class="th text-right">Balance Due</th>
              <th class="th text-center">Status</th>
              <th class="th"/>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr v-for="inv in filtered" :key="inv.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
              <td class="px-4 py-3 font-mono text-sm font-semibold text-gray-800 dark:text-gray-200">{{ inv.number }}</td>
              <td class="px-4 py-3 text-gray-700 dark:text-gray-300">{{ inv.customer_name || '—' }}</td>
              <td class="px-4 py-3 text-gray-600 dark:text-gray-400">{{ fmtDate(inv.date) }}</td>
              <td class="px-4 py-3" :class="dueDateColor(inv)">
                <div class="flex items-center gap-1">
                  <AlertTriangle v-if="inv.status==='overdue'" class="w-3.5 h-3.5 flex-shrink-0" />
                  {{ fmtDate(inv.due_date) }}
                </div>
              </td>
              <td class="px-4 py-3 text-right font-semibold tabular-nums text-gray-900 dark:text-white">{{ fmtCurrency(inv.total_amount) }}</td>
              <td class="px-4 py-3 text-right tabular-nums" :class="inv.balance_due > 0 ? 'font-bold text-blue-700 dark:text-blue-400' : 'text-gray-400'">
                {{ fmtCurrency(inv.balance_due) }}
              </td>
              <td class="px-4 py-3 text-center">
                <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium" :class="STATUS_BADGE[inv.status]">{{ inv.status.replace('_', ' ') }}</span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1 justify-end">
                  <button @click="openDetail(inv)" class="icon-btn" title="View"><Eye class="w-3.5 h-3.5"/></button>
                  <button v-if="inv.status==='draft'" @click="confirmInvoice(inv)" class="icon-btn text-blue-600 dark:text-blue-400" title="Confirm & Post"><CheckCircle class="w-3.5 h-3.5"/></button>
                  <button v-if="['confirmed','partially_paid','overdue'].includes(inv.status)" @click="openPayment(inv)" class="icon-btn text-green-600 dark:text-green-400" title="Record Payment"><Wallet class="w-3.5 h-3.5"/></button>
                  <button v-if="inv.status==='draft'" @click="cancelInvoice(inv)" class="icon-btn text-red-500 dark:text-red-400" title="Cancel"><XCircle class="w-3.5 h-3.5"/></button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ─── Create Invoice Modal ──────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showForm" class="fixed inset-0 z-50 flex items-start justify-center p-4 bg-black/50 backdrop-blur-sm overflow-y-auto" @click.self="closeForm">
        <div class="w-full max-w-4xl bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 my-8">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div class="flex items-center gap-2">
              <ReceiptText class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">New Sales Invoice</h2>
            </div>
            <button @click="closeForm" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"><X class="w-5 h-5"/></button>
          </div>
          <div class="p-6 space-y-6">
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div class="md:col-span-2">
                <label class="label">Customer <span class="text-red-500">*</span></label>
                <select v-model="editingInvoice.customer_id" class="input">
                  <option value="">— Select Customer —</option>
                  <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
              </div>
              <div>
                <label class="label">Invoice Number</label>
                <input v-model="editingInvoice.number" class="input font-mono" placeholder="FAC-2025-001" />
              </div>
              <div>
                <label class="label">Currency</label>
                <select v-model="editingInvoice.currency" class="input">
                  <option value="DZD">DZD</option>
                  <option value="EUR">EUR</option>
                  <option value="USD">USD</option>
                </select>
              </div>
              <div>
                <label class="label">Invoice Date <span class="text-red-500">*</span></label>
                <input v-model="editingInvoice.date" type="date" class="input" />
              </div>
              <div>
                <label class="label">Due Date</label>
                <input v-model="editingInvoice.due_date" type="date" class="input" />
              </div>
              <div>
                <label class="label">Stamp Tax (DZD)</label>
                <input v-model.number="editingInvoice.stamp_tax" type="number" min="0" step="100" class="input" />
              </div>
            </div>

            <!-- Lines -->
            <div>
              <div class="flex items-center justify-between mb-3">
                <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Invoice Lines</h3>
                <button @click="addLine" class="inline-flex items-center gap-1.5 text-xs text-blue-600 dark:text-blue-400 hover:underline"><Plus class="w-3.5 h-3.5"/> Add Line</button>
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
                      <th class="px-3 py-2 w-8"/>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                    <tr v-for="(line, idx) in editingInvoice.lines" :key="idx">
                      <td class="px-2 py-1.5"><input v-model="line.description" class="line-input" placeholder="Item or service description" /></td>
                      <td class="px-2 py-1.5"><input v-model.number="line.quantity" type="number" min="0.001" step="1" class="line-input text-right" /></td>
                      <td class="px-2 py-1.5"><input v-model.number="line.unit_price" type="number" min="0" step="100" class="line-input text-right" /></td>
                      <td class="px-2 py-1.5"><input v-model.number="line.discount_pct" type="number" min="0" max="100" class="line-input text-right" /></td>
                      <td class="px-2 py-1.5"><input v-model.number="line.tva_rate" type="number" min="0" max="100" class="line-input text-right" /></td>
                      <td class="px-2 py-1.5 text-right font-medium text-gray-800 dark:text-gray-200 tabular-nums">{{ lineSubtotal(line).toLocaleString('fr-DZ', { minimumFractionDigits: 2 }) }}</td>
                      <td class="px-2 py-1.5 text-right text-gray-600 dark:text-gray-400 tabular-nums">{{ lineTVA(line).toLocaleString('fr-DZ', { minimumFractionDigits: 2 }) }}</td>
                      <td class="px-2 py-1.5 text-right font-bold text-gray-900 dark:text-white tabular-nums">{{ lineTotal(line).toLocaleString('fr-DZ', { minimumFractionDigits: 2 }) }}</td>
                      <td class="px-2 py-1.5 text-center">
                        <button @click="removeLine(idx)" class="text-red-400 hover:text-red-600 dark:hover:text-red-300 transition-colors" :disabled="editingInvoice.lines.length===1"><Trash2 class="w-3.5 h-3.5"/></button>
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
                <textarea v-model="editingInvoice.notes" rows="4" class="input resize-none" placeholder="Invoice notes…" />
              </div>
              <div class="bg-gray-50 dark:bg-gray-800 rounded-xl p-4 space-y-2">
                <div class="flex justify-between text-sm text-gray-600 dark:text-gray-400"><span>Subtotal HT</span><span class="tabular-nums">{{ fmtCurrency(lineTotals.subtotal) }}</span></div>
                <div class="flex justify-between text-sm text-gray-600 dark:text-gray-400"><span>TVA</span><span class="tabular-nums">{{ fmtCurrency(lineTotals.tva) }}</span></div>
                <div v-if="editingInvoice.stamp_tax" class="flex justify-between text-sm text-gray-600 dark:text-gray-400"><span>Stamp Tax</span><span class="tabular-nums">{{ fmtCurrency(editingInvoice.stamp_tax) }}</span></div>
                <div class="flex justify-between font-bold text-gray-900 dark:text-white border-t border-gray-200 dark:border-gray-700 pt-2 text-base"><span>Total TTC</span><span class="tabular-nums">{{ fmtCurrency(lineTotals.total) }}</span></div>
              </div>
            </div>
          </div>
          <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
            <button @click="closeForm" class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
            <button @click="save" :disabled="saving" class="inline-flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg shadow-sm transition-colors disabled:opacity-60">
              <Save class="w-4 h-4"/>{{ saving ? 'Saving…' : 'Create Invoice' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ─── Invoice Detail Modal ──────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showDetail && detailInvoice" class="fixed inset-0 z-50 flex items-start justify-center p-4 bg-black/50 backdrop-blur-sm overflow-y-auto" @click.self="closeDetail">
        <div class="w-full max-w-3xl bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 my-8">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div class="flex items-center gap-3">
              <ReceiptText class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white font-mono">{{ detailInvoice.number }}</h2>
              <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium" :class="STATUS_BADGE[detailInvoice.status]">{{ detailInvoice.status.replace('_', ' ') }}</span>
            </div>
            <button @click="closeDetail" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"><X class="w-5 h-5"/></button>
          </div>
          <div class="p-6 space-y-5">
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
              <div><p class="text-gray-400 text-xs">Customer</p><p class="font-semibold text-gray-800 dark:text-gray-200">{{ detailInvoice.customer_name || '—' }}</p></div>
              <div><p class="text-gray-400 text-xs">Date</p><p class="font-semibold text-gray-800 dark:text-gray-200">{{ fmtDate(detailInvoice.date) }}</p></div>
              <div><p class="text-gray-400 text-xs">Due Date</p><p class="font-semibold" :class="dueDateColor(detailInvoice) || 'text-gray-800 dark:text-gray-200'">{{ fmtDate(detailInvoice.due_date) }}</p></div>
              <div><p class="text-gray-400 text-xs">Currency</p><p class="font-semibold text-gray-800 dark:text-gray-200">{{ detailInvoice.currency }}</p></div>
            </div>
            <!-- Payment Progress -->
            <div v-if="detailInvoice.total_amount > 0" class="bg-gray-50 dark:bg-gray-800 rounded-xl p-4">
              <div class="flex items-center justify-between mb-2 text-sm">
                <span class="text-gray-600 dark:text-gray-400">Payment Progress</span>
                <span class="font-semibold text-gray-900 dark:text-white">{{ Math.round((detailInvoice.paid_amount / detailInvoice.total_amount) * 100) }}%</span>
              </div>
              <div class="h-2 rounded-full bg-gray-200 dark:bg-gray-700">
                <div class="h-2 rounded-full bg-green-500 dark:bg-green-400 transition-all" :style="{ width: Math.min(100, (detailInvoice.paid_amount / detailInvoice.total_amount) * 100) + '%' }" />
              </div>
              <div class="flex justify-between mt-2 text-xs text-gray-500 dark:text-gray-400">
                <span>Paid: {{ fmtCurrency(detailInvoice.paid_amount) }}</span>
                <span class="font-semibold text-blue-600 dark:text-blue-400">Balance: {{ fmtCurrency(detailInvoice.balance_due) }}</span>
              </div>
            </div>
            <!-- Lines -->
            <div v-if="detailInvoice.lines?.length" class="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-700">
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
                  <tr v-for="l in detailInvoice.lines" :key="l.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                    <td class="px-3 py-2 text-gray-700 dark:text-gray-300">{{ l.description }}</td>
                    <td class="px-3 py-2 text-right tabular-nums text-gray-700 dark:text-gray-300">{{ l.quantity }}</td>
                    <td class="px-3 py-2 text-right tabular-nums text-gray-700 dark:text-gray-300">{{ fmtCurrency(l.unit_price) }}</td>
                    <td class="px-3 py-2 text-right tabular-nums text-gray-600 dark:text-gray-400">{{ l.tva_rate }}%</td>
                    <td class="px-3 py-2 text-right tabular-nums font-semibold text-gray-900 dark:text-white">{{ fmtCurrency(l.total) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <!-- Totals -->
            <div class="flex justify-end">
              <div class="w-72 space-y-1.5 text-sm">
                <div class="flex justify-between text-gray-600 dark:text-gray-400"><span>Subtotal HT</span><span class="tabular-nums">{{ fmtCurrency(detailInvoice.subtotal) }}</span></div>
                <div class="flex justify-between text-gray-600 dark:text-gray-400"><span>TVA</span><span class="tabular-nums">{{ fmtCurrency(detailInvoice.tva_amount) }}</span></div>
                <div v-if="detailInvoice.stamp_tax" class="flex justify-between text-gray-600 dark:text-gray-400"><span>Stamp Tax</span><span class="tabular-nums">{{ fmtCurrency(detailInvoice.stamp_tax) }}</span></div>
                <div class="flex justify-between font-bold text-gray-900 dark:text-white border-t border-gray-200 dark:border-gray-700 pt-1.5 text-base"><span>Total TTC</span><span class="tabular-nums">{{ fmtCurrency(detailInvoice.total_amount) }}</span></div>
                <div class="flex justify-between text-green-700 dark:text-green-400"><span>Paid</span><span class="tabular-nums">{{ fmtCurrency(detailInvoice.paid_amount) }}</span></div>
                <div class="flex justify-between font-bold text-blue-700 dark:text-blue-400 text-base"><span>Balance Due</span><span class="tabular-nums">{{ fmtCurrency(detailInvoice.balance_due) }}</span></div>
              </div>
            </div>
            <!-- Actions -->
            <div class="flex flex-wrap gap-2 pt-2 border-t border-gray-200 dark:border-gray-700">
              <button v-if="detailInvoice.status==='draft'" @click="confirmInvoice(detailInvoice); closeDetail()" class="btn-secondary inline-flex items-center gap-1.5"><CheckCircle class="w-4 h-4"/> Confirm & Post</button>
              <button v-if="['confirmed','partially_paid','overdue'].includes(detailInvoice.status)" @click="openPayment(detailInvoice); closeDetail()" class="btn-primary inline-flex items-center gap-1.5"><Wallet class="w-4 h-4"/> Record Payment</button>
              <button v-if="detailInvoice.status==='draft'" @click="cancelInvoice(detailInvoice); closeDetail()" class="btn-danger inline-flex items-center gap-1.5"><XCircle class="w-4 h-4"/> Cancel</button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ─── Payment Modal ─────────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showPaymentModal && payingInvoice" class="fixed inset-0 z-60 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" @click.self="closePayment">
        <div class="w-full max-w-md bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div class="flex items-center gap-2">
              <Wallet class="w-5 h-5 text-green-600 dark:text-green-400" />
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Record Payment</h2>
            </div>
            <button @click="closePayment" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"><X class="w-5 h-5"/></button>
          </div>
          <div class="p-6 space-y-4">
            <div class="bg-blue-50 dark:bg-blue-900/20 rounded-lg px-4 py-3 text-sm">
              <p class="text-blue-700 dark:text-blue-300"><span class="font-semibold">Invoice:</span> {{ payingInvoice.number }}</p>
              <p class="text-blue-700 dark:text-blue-300 mt-0.5"><span class="font-semibold">Balance Due:</span> {{ fmtCurrency(payingInvoice.balance_due) }}</p>
            </div>
            <div>
              <label class="label">Payment Amount (DZD) <span class="text-red-500">*</span></label>
              <input v-model.number="paymentForm.amount" type="number" min="0.01" :max="payingInvoice.balance_due" step="1000" class="input" />
            </div>
            <div>
              <label class="label">Payment Method</label>
              <select v-model="paymentForm.method" class="input">
                <option value="bank">Bank Transfer</option>
                <option value="cash">Cash</option>
                <option value="cheque">Cheque</option>
                <option value="card">Card</option>
              </select>
            </div>
            <div>
              <label class="label">Payment Date</label>
              <input v-model="paymentForm.payment_date" type="date" class="input" />
            </div>
            <div>
              <label class="label">Notes</label>
              <input v-model="paymentForm.notes" class="input" placeholder="Reference, check number…" />
            </div>
          </div>
          <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
            <button @click="closePayment" class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
            <button @click="recordPayment" :disabled="saving" class="inline-flex items-center gap-2 px-5 py-2 bg-green-600 hover:bg-green-700 text-white text-sm font-medium rounded-lg transition-colors disabled:opacity-60">
              <Wallet class="w-4 h-4"/>{{ saving ? 'Saving…' : 'Record Payment' }}
            </button>
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
.btn-primary { @apply px-4 py-2 bg-green-600 hover:bg-green-700 text-white text-sm font-medium rounded-lg transition-colors; }
.btn-secondary { @apply px-4 py-2 border border-blue-200 dark:border-blue-700 text-blue-700 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 text-sm font-medium rounded-lg transition-colors; }
.btn-danger { @apply px-4 py-2 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 text-sm font-medium rounded-lg transition-colors; }
</style>
