<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { purchaseAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  Plus, X, Save, Search, RefreshCw, Eye, CheckCircle,
  XCircle, CreditCard, ChevronDown, ChevronUp, Trash2,
  FileText, Calendar, Building2, DollarSign, AlertTriangle,
  Clock, Wallet, ReceiptText, Link2
} from 'lucide-vue-next'

// ─── Types ───────────────────────────────────────────────────────────────────

interface InvoiceLine {
  po_line_id: string
  item_id: string
  description: string
  quantity: number
  unit_price: number
  discount_pct: number
  tva_rate: number
  sort_order: number
}

interface PurchaseInvoice {
  id: string
  number: string
  supplier_ref: string
  po_id: string
  grn_id: string
  supplier_id: string
  supplier_name: string
  date: string
  due_date: string
  status: string
  subtotal: number
  tva_amount: number
  total_amount: number
  paid_amount: number
  balance_due: number
  currency: string
  match_status: string
  notes: string
  created_at: string
  lines?: InvoiceLine[]
}

interface Supplier {
  id: string
  name: string
  code: string
  payment_terms: number
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
  unit_price: number
  discount_pct: number
  tva_rate: number
}

const EMPTY_LINE = (sort = 0): InvoiceLine => ({
  po_line_id: '',
  item_id: '',
  description: '',
  quantity: 1,
  unit_price: 0,
  discount_pct: 0,
  tva_rate: 19,
  sort_order: sort,
})

const STATUS_BADGE: Record<string, string> = {
  draft:           'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  confirmed:       'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  partially_paid:  'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  paid:            'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300',
  cancelled:       'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300',
  overdue:         'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300',
}

const STATUS_LABEL: Record<string, string> = {
  draft:          'Draft',
  confirmed:      'Confirmed',
  partially_paid: 'Partially Paid',
  paid:           'Paid',
  cancelled:      'Cancelled',
  overdue:        'Overdue',
}

const MATCH_BADGE: Record<string, string> = {
  unmatched: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400',
  '2way':    'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/40 dark:text-yellow-300',
  '3way':    'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300',
}

// ─── State ───────────────────────────────────────────────────────────────────

const app = useAppStore()
const invoices   = ref<PurchaseInvoice[]>([])
const suppliers  = ref<Supplier[]>([])
const orders     = ref<PurchaseOrder[]>([])
const loading    = ref(true)
const saving     = ref(false)

const showForm          = ref(false)
const showDetail        = ref(false)
const showPaymentModal  = ref(false)

const form = ref<{
  supplier_id: string
  supplier_ref: string
  po_id: string
  grn_id: string
  date: string
  due_date: string
  currency: string
  notes: string
  lines: InvoiceLine[]
}>({
  supplier_id: '',
  supplier_ref: '',
  po_id: '',
  grn_id: '',
  date: today(),
  due_date: '',
  currency: 'DZD',
  notes: '',
  lines: [EMPTY_LINE()],
})

const detailInvoice   = ref<PurchaseInvoice | null>(null)
const payingInvoice   = ref<PurchaseInvoice | null>(null)
const paymentForm     = ref({ amount: 0, payment_date: today(), notes: '' })

const search       = ref('')
const filterStatus = ref('')
const sortBy       = ref<keyof PurchaseInvoice>('date')
const sortDir      = ref<'asc' | 'desc'>('desc')

// ─── Computed ────────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = [...invoices.value]
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(i =>
      i.number.toLowerCase().includes(q) ||
      i.supplier_name.toLowerCase().includes(q) ||
      (i.supplier_ref ?? '').toLowerCase().includes(q)
    )
  }
  if (filterStatus.value) {
    list = list.filter(i => i.status === filterStatus.value)
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
  const all = invoices.value
  const unpaid = all.filter(i => !['paid', 'cancelled'].includes(i.status))
  return {
    total:        all.length,
    draft:        all.filter(i => i.status === 'draft').length,
    confirmed:    all.filter(i => i.status === 'confirmed').length,
    paid:         all.filter(i => i.status === 'paid').length,
    outstanding:  unpaid.reduce((s, i) => s + (i.balance_due ?? (i.total_amount - i.paid_amount)), 0),
    overdue:      all.filter(i => i.status === 'overdue').length,
  }
})

const formTotals = computed(() => {
  let sub = 0, tva = 0
  for (const l of form.value.lines) {
    const base = l.quantity * l.unit_price * (1 - l.discount_pct / 100)
    sub += base
    tva += base * l.tva_rate / 100
  }
  return { subtotal: sub, tva, total: sub + tva }
})

// ─── Helpers ─────────────────────────────────────────────────────────────────

function today() {
  return new Date().toISOString().slice(0, 10)
}

function addDays(days: number) {
  const d = new Date()
  d.setDate(d.getDate() + days)
  return d.toISOString().slice(0, 10)
}

function fmt(n: number) {
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n ?? 0)
}

function fmtDate(d: string | null | undefined) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('fr-DZ')
}

function isOverdue(inv: PurchaseInvoice) {
  return inv.due_date && new Date(inv.due_date) < new Date() && !['paid', 'cancelled'].includes(inv.status)
}

function toggleSort(col: keyof PurchaseInvoice) {
  if (sortBy.value === col) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = col
    sortDir.value = 'desc'
  }
}

function lineTotal(l: InvoiceLine) {
  const base = l.quantity * l.unit_price * (1 - l.discount_pct / 100)
  return base + base * l.tva_rate / 100
}

// ─── Data Loading ─────────────────────────────────────────────────────────────

async function loadInvoices() {
  loading.value = true
  try {
    const r = await purchaseAPI.getInvoices()
    invoices.value = r.data ?? []
  } catch {
    app.addToast('Failed to load invoices', 'error')
  } finally {
    loading.value = false
  }
}

async function loadSuppliers() {
  try {
    const r = await purchaseAPI.getSuppliers()
    suppliers.value = r.data ?? []
  } catch {
    suppliers.value = []
  }
}

async function loadOrders() {
  try {
    const r = await purchaseAPI.getOrders()
    orders.value = (r.data ?? []).filter((o: PurchaseOrder) =>
      ['approved', 'partially_received', 'received'].includes(o.status)
    )
  } catch {
    orders.value = []
  }
}

// ─── Form Actions ─────────────────────────────────────────────────────────────

function openCreate() {
  form.value = {
    supplier_id:  '',
    supplier_ref: '',
    po_id:        '',
    grn_id:       '',
    date:         today(),
    due_date:     addDays(30),
    currency:     'DZD',
    notes:        '',
    lines:        [EMPTY_LINE()],
  }
  showForm.value = true
}

async function onPOSelect() {
  if (!form.value.po_id) return
  try {
    const r = await purchaseAPI.getOrder(form.value.po_id)
    const po: PurchaseOrder = r.data
    if (po) {
      form.value.supplier_id = po.supplier_id
      // Pre-fill lines
      if (po.lines?.length) {
        form.value.lines = po.lines.map((l: POLine, i: number) => ({
          po_line_id:  l.id,
          item_id:     '',
          description: l.description,
          quantity:    l.quantity,
          unit_price:  l.unit_price,
          discount_pct: l.discount_pct ?? 0,
          tva_rate:    l.tva_rate ?? 19,
          sort_order:  i,
        }))
      }
    }
  } catch {
    /* silent */
  }
}

function onSupplierSelect() {
  const sup = suppliers.value.find(s => s.id === form.value.supplier_id)
  if (sup && sup.payment_terms) {
    form.value.due_date = addDays(sup.payment_terms)
  }
}

function addLine() {
  form.value.lines.push(EMPTY_LINE(form.value.lines.length))
}

function removeLine(i: number) {
  if (form.value.lines.length > 1) form.value.lines.splice(i, 1)
}

async function saveInvoice() {
  if (!form.value.supplier_id) {
    app.addToast('Please select a supplier', 'error')
    return
  }
  const validLines = form.value.lines.filter(l => l.description.trim())
  if (!validLines.length) {
    app.addToast('Add at least one line with a description', 'error')
    return
  }
  saving.value = true
  try {
    await purchaseAPI.createInvoice({
      supplier_id:  form.value.supplier_id,
      supplier_ref: form.value.supplier_ref,
      po_id:        form.value.po_id,
      grn_id:       form.value.grn_id,
      date:         form.value.date,
      due_date:     form.value.due_date,
      currency:     form.value.currency,
      notes:        form.value.notes,
      lines:        validLines,
    })
    app.addToast('Invoice created successfully', 'success')
    showForm.value = false
    await loadInvoices()
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Save failed'
    app.addToast(msg, 'error')
  } finally {
    saving.value = false
  }
}

// ─── Confirm / Cancel ─────────────────────────────────────────────────────────

async function confirmInvoice(id: string) {
  try {
    await purchaseAPI.confirmInvoice(id)
    app.addToast('Invoice confirmed', 'success')
    await loadInvoices()
    if (detailInvoice.value?.id === id) await refreshDetail(id)
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Confirm failed'
    app.addToast(msg, 'error')
  }
}

async function cancelInvoice(id: string) {
  if (!confirm('Cancel this invoice? This action cannot be undone.')) return
  try {
    await purchaseAPI.cancelInvoice(id)
    app.addToast('Invoice cancelled', 'success')
    await loadInvoices()
    if (detailInvoice.value?.id === id) await refreshDetail(id)
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Cancel failed'
    app.addToast(msg, 'error')
  }
}

async function matchInvoice(id: string) {
  try {
    await purchaseAPI.matchInvoice(id)
    app.addToast('3-way match performed', 'success')
    await loadInvoices()
    if (detailInvoice.value?.id === id) await refreshDetail(id)
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Match failed'
    app.addToast(msg, 'error')
  }
}

// ─── Detail ───────────────────────────────────────────────────────────────────

async function viewDetail(id: string) {
  try {
    const r = await purchaseAPI.getInvoice(id)
    detailInvoice.value = r.data
    showDetail.value = true
  } catch {
    app.addToast('Failed to load invoice details', 'error')
  }
}

async function refreshDetail(id: string) {
  try {
    const r = await purchaseAPI.getInvoice(id)
    detailInvoice.value = r.data
  } catch { /* silent */ }
}

// ─── Payment ──────────────────────────────────────────────────────────────────

function openPayment(inv: PurchaseInvoice) {
  payingInvoice.value = inv
  paymentForm.value = {
    amount:       inv.balance_due ?? (inv.total_amount - inv.paid_amount),
    payment_date: today(),
    notes:        '',
  }
  showPaymentModal.value = true
}

async function submitPayment() {
  if (!payingInvoice.value) return
  saving.value = true
  try {
    await purchaseAPI.recordPayment(payingInvoice.value.id, paymentForm.value)
    app.addToast('Payment recorded successfully', 'success')
    showPaymentModal.value = false
    await loadInvoices()
    if (detailInvoice.value?.id === payingInvoice.value.id) await refreshDetail(payingInvoice.value.id)
  } catch (e: unknown) {
    const msg = (e as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Payment failed'
    app.addToast(msg, 'error')
  } finally {
    saving.value = false
  }
}

// ─── Lifecycle ───────────────────────────────────────────────────────────────

onMounted(async () => {
  await Promise.all([loadInvoices(), loadSuppliers(), loadOrders()])
})
</script>

<template>
  <div class="space-y-6">

    <!-- ── Header ──────────────────────────────────────────────────────────── -->
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white">Supplier Invoices</h1>
        <p class="mt-0.5 text-sm text-slate-500 dark:text-slate-400">
          Manage payables, confirm invoices, record payments
        </p>
      </div>
      <button
        @click="openCreate"
        class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-700 transition-colors"
      >
        <Plus class="h-4 w-4" />
        New Invoice
      </button>
    </div>

    <!-- ── KPI Cards ───────────────────────────────────────────────────────── -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-5">
      <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
        <div class="flex items-center justify-between">
          <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Total</p>
          <ReceiptText class="h-4 w-4 text-slate-400" />
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
          <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Confirmed</p>
          <CheckCircle class="h-4 w-4 text-blue-500" />
        </div>
        <p class="mt-2 text-2xl font-bold text-blue-600 dark:text-blue-400">{{ kpis.confirmed }}</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-800">
        <div class="flex items-center justify-between">
          <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">Paid</p>
          <Wallet class="h-4 w-4 text-green-500" />
        </div>
        <p class="mt-2 text-2xl font-bold text-green-600 dark:text-green-400">{{ kpis.paid }}</p>
      </div>
      <div class="col-span-2 lg:col-span-1 rounded-xl border border-orange-200 bg-orange-50 p-4 dark:border-orange-800/40 dark:bg-orange-900/20">
        <div class="flex items-center justify-between">
          <p class="text-xs font-medium text-orange-600 dark:text-orange-400 uppercase tracking-wider">Outstanding</p>
          <AlertTriangle class="h-4 w-4 text-orange-500" />
        </div>
        <p class="mt-2 text-xl font-bold text-orange-700 dark:text-orange-300">{{ fmt(kpis.outstanding) }}</p>
        <p class="text-xs text-orange-500">DZD payable</p>
      </div>
    </div>

    <!-- ── Toolbar ─────────────────────────────────────────────────────────── -->
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
      <div class="relative flex-1">
        <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
        <input
          v-model="search"
          type="text"
          placeholder="Search by invoice#, supplier, ref..."
          class="w-full rounded-lg border border-slate-200 bg-white py-2 pl-9 pr-3 text-sm text-slate-900 placeholder-slate-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-800 dark:text-white dark:placeholder-slate-500"
        />
      </div>
      <select
        v-model="filterStatus"
        class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200"
      >
        <option value="">All Statuses</option>
        <option value="draft">Draft</option>
        <option value="confirmed">Confirmed</option>
        <option value="partially_paid">Partially Paid</option>
        <option value="paid">Paid</option>
        <option value="overdue">Overdue</option>
        <option value="cancelled">Cancelled</option>
      </select>
      <button
        @click="loadInvoices"
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
        <ReceiptText class="mb-3 h-12 w-12 text-slate-300 dark:text-slate-600" />
        <p class="text-sm font-medium text-slate-500 dark:text-slate-400">No invoices found</p>
        <p class="mt-1 text-xs text-slate-400 dark:text-slate-500">Create your first supplier invoice</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-slate-200 bg-slate-50 dark:border-slate-700 dark:bg-slate-900/50">
              <th
                v-for="col in [
                  { key: 'number',       label: 'Invoice#' },
                  { key: 'supplier_name', label: 'Supplier' },
                  { key: 'date',         label: 'Date' },
                  { key: 'due_date',     label: 'Due Date' },
                  { key: 'status',       label: 'Status' },
                  { key: 'total_amount', label: 'Total' },
                  { key: 'paid_amount',  label: 'Paid' },
                  { key: 'balance_due',  label: 'Balance Due' },
                ]"
                :key="col.key"
                @click="toggleSort(col.key as keyof PurchaseInvoice)"
                class="cursor-pointer select-none px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 whitespace-nowrap"
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
              v-for="inv in filtered"
              :key="inv.id"
              :class="['group hover:bg-slate-50 dark:hover:bg-slate-700/50 transition-colors', isOverdue(inv) ? 'bg-red-50/40 dark:bg-red-900/10' : '']"
            >
              <td class="px-4 py-3">
                <div>
                  <span class="font-mono text-xs font-semibold text-blue-600 dark:text-blue-400">{{ inv.number }}</span>
                  <span v-if="inv.supplier_ref" class="ml-2 text-xs text-slate-400">({{ inv.supplier_ref }})</span>
                </div>
                <div class="mt-0.5">
                  <span :class="['inline-flex items-center rounded px-1.5 py-0.5 text-xs font-medium', MATCH_BADGE[inv.match_status] ?? MATCH_BADGE.unmatched]">
                    {{ inv.match_status === '3way' ? '3-way match' : inv.match_status === '2way' ? '2-way' : 'unmatched' }}
                  </span>
                </div>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900/40">
                    <Building2 class="h-3.5 w-3.5 text-blue-600 dark:text-blue-400" />
                  </div>
                  <span class="font-medium text-slate-900 dark:text-white">{{ inv.supplier_name }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400">
                <span class="inline-flex items-center gap-1.5">
                  <Calendar class="h-3.5 w-3.5 text-slate-400" />
                  {{ fmtDate(inv.date) }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span
                  :class="[
                    'inline-flex items-center gap-1 text-sm',
                    isOverdue(inv) ? 'font-semibold text-red-600 dark:text-red-400' : 'text-slate-600 dark:text-slate-400'
                  ]"
                >
                  <AlertTriangle v-if="isOverdue(inv)" class="h-3.5 w-3.5" />
                  {{ fmtDate(inv.due_date) }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span :class="['inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium', STATUS_BADGE[inv.status] ?? STATUS_BADGE.draft]">
                  {{ STATUS_LABEL[inv.status] ?? inv.status }}
                </span>
              </td>
              <td class="px-4 py-3 text-right font-semibold text-slate-900 dark:text-white">
                {{ fmt(inv.total_amount) }}
                <span class="ml-0.5 text-xs font-normal text-slate-400">DZD</span>
              </td>
              <td class="px-4 py-3 text-right text-green-600 dark:text-green-400">
                {{ fmt(inv.paid_amount) }}
              </td>
              <td class="px-4 py-3 text-right">
                <span :class="[
                  'font-bold',
                  (inv.balance_due ?? (inv.total_amount - inv.paid_amount)) > 0
                    ? 'text-red-600 dark:text-red-400'
                    : 'text-slate-400'
                ]">
                  {{ fmt(inv.balance_due ?? (inv.total_amount - inv.paid_amount)) }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center justify-end gap-1">
                  <button
                    @click="viewDetail(inv.id)"
                    class="rounded p-1.5 text-slate-400 hover:bg-slate-100 hover:text-blue-600 dark:hover:bg-slate-700 dark:hover:text-blue-400 transition-colors"
                    title="View Details"
                  >
                    <Eye class="h-4 w-4" />
                  </button>
                  <button
                    v-if="inv.status === 'draft'"
                    @click="confirmInvoice(inv.id)"
                    class="rounded p-1.5 text-slate-400 hover:bg-blue-50 hover:text-blue-600 dark:hover:bg-blue-900/30 dark:hover:text-blue-400 transition-colors"
                    title="Confirm Invoice"
                  >
                    <CheckCircle class="h-4 w-4" />
                  </button>
                  <button
                    v-if="['confirmed', 'partially_paid'].includes(inv.status)"
                    @click="openPayment(inv)"
                    class="rounded p-1.5 text-slate-400 hover:bg-green-50 hover:text-green-600 dark:hover:bg-green-900/30 dark:hover:text-green-400 transition-colors"
                    title="Record Payment"
                  >
                    <CreditCard class="h-4 w-4" />
                  </button>
                  <button
                    v-if="['draft', 'confirmed'].includes(inv.status)"
                    @click="cancelInvoice(inv.id)"
                    class="rounded p-1.5 text-slate-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/30 dark:hover:text-red-400 transition-colors"
                    title="Cancel Invoice"
                  >
                    <XCircle class="h-4 w-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-if="!loading && filtered.length" class="border-t border-slate-100 px-4 py-2 dark:border-slate-700">
        <p class="text-xs text-slate-400">
          Showing {{ filtered.length }} of {{ invoices.length }} invoices
        </p>
      </div>
    </div>

    <!-- ── Create Invoice Modal ────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showForm" class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/60 p-4 backdrop-blur-sm">
        <div class="my-6 w-full max-w-5xl rounded-2xl bg-white shadow-2xl dark:bg-slate-800">

          <!-- Header -->
          <div class="flex items-center justify-between border-b border-slate-200 px-6 py-4 dark:border-slate-700">
            <div class="flex items-center gap-3">
              <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-600">
                <ReceiptText class="h-5 w-5 text-white" />
              </div>
              <div>
                <h2 class="text-lg font-semibold text-slate-900 dark:text-white">New Supplier Invoice</h2>
                <p class="text-xs text-slate-500 dark:text-slate-400">Enter invoice details and line items</p>
              </div>
            </div>
            <button @click="showForm = false" class="rounded-lg p-2 text-slate-400 hover:bg-slate-100 transition-colors dark:hover:bg-slate-700">
              <X class="h-5 w-5" />
            </button>
          </div>

          <div class="space-y-6 p-6">

            <!-- Row 1: Supplier + Supplier Ref -->
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  Supplier <span class="text-red-500">*</span>
                </label>
                <select
                  v-model="form.supplier_id"
                  @change="onSupplierSelect"
                  class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                >
                  <option value="">Select supplier...</option>
                  <option v-for="s in suppliers" :key="s.id" :value="s.id">{{ s.name }}</option>
                </select>
              </div>
              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  Supplier Invoice Reference
                </label>
                <input
                  v-model="form.supplier_ref"
                  type="text"
                  placeholder="Supplier's invoice number"
                  class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white dark:placeholder-slate-500"
                />
              </div>
            </div>

            <!-- Row 2: PO Link + Date + Due Date -->
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  Link to Purchase Order
                </label>
                <select
                  v-model="form.po_id"
                  @change="onPOSelect"
                  class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                >
                  <option value="">No PO linked</option>
                  <option v-for="o in orders" :key="o.id" :value="o.id">
                    {{ o.number }} — {{ o.supplier_name }}
                  </option>
                </select>
              </div>
              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  Invoice Date <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="form.date"
                  type="date"
                  class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                />
              </div>
              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  Due Date
                </label>
                <input
                  v-model="form.due_date"
                  type="date"
                  class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                />
              </div>
            </div>

            <!-- Notes + Currency -->
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div class="sm:col-span-2">
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">Notes</label>
                <input
                  v-model="form.notes"
                  type="text"
                  placeholder="Optional internal notes"
                  class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white dark:placeholder-slate-500"
                />
              </div>
              <div>
                <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">Currency</label>
                <select
                  v-model="form.currency"
                  class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                >
                  <option value="DZD">DZD</option>
                  <option value="EUR">EUR</option>
                  <option value="USD">USD</option>
                </select>
              </div>
            </div>

            <!-- Lines Table -->
            <div>
              <div class="mb-3 flex items-center justify-between">
                <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300">Invoice Lines</h3>
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
                      <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 w-20">Qty</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 w-28">Unit Price</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 w-20">Disc %</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 w-20">TVA %</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 w-32">Total TTC</th>
                      <th class="w-10"></th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
                    <tr v-for="(line, i) in form.lines" :key="i">
                      <td class="px-3 py-2">
                        <input
                          v-model="line.description"
                          type="text"
                          placeholder="Description"
                          class="w-full min-w-[160px] rounded border border-slate-200 bg-white px-2 py-1.5 text-xs text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                        />
                      </td>
                      <td class="px-3 py-2">
                        <input
                          v-model.number="line.quantity"
                          type="number"
                          min="0"
                          step="0.001"
                          class="w-full rounded border border-slate-200 bg-white px-2 py-1.5 text-right text-xs text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                        />
                      </td>
                      <td class="px-3 py-2">
                        <input
                          v-model.number="line.unit_price"
                          type="number"
                          min="0"
                          step="0.01"
                          class="w-full rounded border border-slate-200 bg-white px-2 py-1.5 text-right text-xs text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                        />
                      </td>
                      <td class="px-3 py-2">
                        <input
                          v-model.number="line.discount_pct"
                          type="number"
                          min="0"
                          max="100"
                          step="0.01"
                          class="w-full rounded border border-slate-200 bg-white px-2 py-1.5 text-right text-xs text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                        />
                      </td>
                      <td class="px-3 py-2">
                        <input
                          v-model.number="line.tva_rate"
                          type="number"
                          min="0"
                          max="99"
                          step="0.01"
                          class="w-full rounded border border-slate-200 bg-white px-2 py-1.5 text-right text-xs text-slate-900 focus:border-blue-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
                        />
                      </td>
                      <td class="px-3 py-2 text-right text-xs font-semibold text-slate-700 dark:text-slate-300">
                        {{ fmt(lineTotal(line)) }}
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
                    <tr class="border-t border-slate-200 dark:border-slate-600">
                      <td colspan="5" class="px-3 py-1.5 text-right text-xs text-slate-400 uppercase tracking-wider">Subtotal HT</td>
                      <td class="px-3 py-1.5 text-right text-xs text-slate-600 dark:text-slate-400">{{ fmt(formTotals.subtotal) }}</td>
                      <td></td>
                    </tr>
                    <tr class="border-t border-slate-100 dark:border-slate-700">
                      <td colspan="5" class="px-3 py-1.5 text-right text-xs text-slate-400 uppercase tracking-wider">TVA</td>
                      <td class="px-3 py-1.5 text-right text-xs text-slate-600 dark:text-slate-400">{{ fmt(formTotals.tva) }}</td>
                      <td></td>
                    </tr>
                    <tr class="border-t-2 border-slate-200 dark:border-slate-600">
                      <td colspan="5" class="px-3 py-2.5 text-right text-xs font-semibold text-slate-500 uppercase tracking-wider">Total TTC</td>
                      <td class="px-3 py-2.5 text-right text-sm font-bold text-slate-900 dark:text-white">
                        {{ fmt(formTotals.total) }} <span class="text-xs font-normal text-slate-400">{{ form.currency }}</span>
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
              @click="saveInvoice"
              :disabled="saving || !form.supplier_id"
              class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-5 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-700 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
            >
              <Save v-if="!saving" class="h-4 w-4" />
              <div v-else class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></div>
              {{ saving ? 'Saving...' : 'Create Invoice' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Detail Modal ────────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showDetail && detailInvoice" class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/60 p-4 backdrop-blur-sm">
        <div class="my-6 w-full max-w-4xl rounded-2xl bg-white shadow-2xl dark:bg-slate-800">

          <!-- Header -->
          <div class="flex items-start justify-between border-b border-slate-200 px-6 py-4 dark:border-slate-700">
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <span class="font-mono text-lg font-bold text-blue-600 dark:text-blue-400">{{ detailInvoice.number }}</span>
                <span :class="['inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium', STATUS_BADGE[detailInvoice.status] ?? STATUS_BADGE.draft]">
                  {{ STATUS_LABEL[detailInvoice.status] ?? detailInvoice.status }}
                </span>
                <span :class="['inline-flex items-center rounded px-2 py-0.5 text-xs font-medium', MATCH_BADGE[detailInvoice.match_status] ?? MATCH_BADGE.unmatched]">
                  {{ detailInvoice.match_status }}
                </span>
              </div>
              <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
                {{ detailInvoice.supplier_name }}
                <span v-if="detailInvoice.supplier_ref"> &mdash; Ref: {{ detailInvoice.supplier_ref }}</span>
              </p>
            </div>
            <button @click="showDetail = false" class="rounded-lg p-2 text-slate-400 hover:bg-slate-100 transition-colors dark:hover:bg-slate-700">
              <X class="h-5 w-5" />
            </button>
          </div>

          <!-- Info Grid -->
          <div class="grid grid-cols-2 gap-4 border-b border-slate-100 p-6 sm:grid-cols-4 dark:border-slate-700">
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Invoice Date</p>
              <p class="mt-1 text-sm text-slate-700 dark:text-slate-300">{{ fmtDate(detailInvoice.date) }}</p>
            </div>
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Due Date</p>
              <p :class="['mt-1 text-sm font-medium', isOverdue(detailInvoice) ? 'text-red-600 dark:text-red-400' : 'text-slate-700 dark:text-slate-300']">
                {{ fmtDate(detailInvoice.due_date) }}
              </p>
            </div>
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Purchase Order</p>
              <p class="mt-1 font-mono text-sm text-slate-700 dark:text-slate-300">{{ detailInvoice.po_id ? detailInvoice.po_id.slice(0,8) + '...' : '—' }}</p>
            </div>
            <div>
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Currency</p>
              <p class="mt-1 text-sm text-slate-700 dark:text-slate-300">{{ detailInvoice.currency }}</p>
            </div>
          </div>

          <!-- Financials -->
          <div class="grid grid-cols-2 gap-4 border-b border-slate-100 px-6 py-4 sm:grid-cols-4 dark:border-slate-700">
            <div class="rounded-lg bg-slate-50 p-3 dark:bg-slate-700/50">
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">Subtotal HT</p>
              <p class="mt-1 text-base font-bold text-slate-900 dark:text-white">{{ fmt(detailInvoice.subtotal) }}</p>
            </div>
            <div class="rounded-lg bg-slate-50 p-3 dark:bg-slate-700/50">
              <p class="text-xs font-semibold uppercase tracking-wider text-slate-400">TVA</p>
              <p class="mt-1 text-base font-bold text-slate-900 dark:text-white">{{ fmt(detailInvoice.tva_amount) }}</p>
            </div>
            <div class="rounded-lg bg-blue-50 p-3 dark:bg-blue-900/30">
              <p class="text-xs font-semibold uppercase tracking-wider text-blue-600 dark:text-blue-400">Total TTC</p>
              <p class="mt-1 text-base font-bold text-blue-700 dark:text-blue-300">{{ fmt(detailInvoice.total_amount) }}</p>
            </div>
            <div :class="['rounded-lg p-3', (detailInvoice.balance_due ?? 0) > 0 ? 'bg-red-50 dark:bg-red-900/20' : 'bg-green-50 dark:bg-green-900/20']">
              <p :class="['text-xs font-semibold uppercase tracking-wider', (detailInvoice.balance_due ?? 0) > 0 ? 'text-red-500' : 'text-green-500']">Balance Due</p>
              <p :class="['mt-1 text-base font-bold', (detailInvoice.balance_due ?? 0) > 0 ? 'text-red-700 dark:text-red-300' : 'text-green-700 dark:text-green-300']">
                {{ fmt(detailInvoice.balance_due ?? (detailInvoice.total_amount - detailInvoice.paid_amount)) }}
              </p>
            </div>
          </div>

          <!-- Lines -->
          <div v-if="detailInvoice.lines?.length" class="p-6">
            <h3 class="mb-3 text-sm font-semibold text-slate-700 dark:text-slate-300">Invoice Lines</h3>
            <div class="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-700">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-slate-200 bg-slate-50 dark:border-slate-700 dark:bg-slate-900/50">
                    <th class="px-3 py-2 text-left text-xs font-semibold text-slate-500">Description</th>
                    <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500">Qty</th>
                    <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500">Unit Price</th>
                    <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500">Disc %</th>
                    <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500">TVA %</th>
                    <th class="px-3 py-2 text-right text-xs font-semibold text-slate-500">Total TTC</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
                  <tr v-for="(line, i) in detailInvoice.lines" :key="i" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
                    <td class="px-3 py-2 text-slate-700 dark:text-slate-300">{{ line.description }}</td>
                    <td class="px-3 py-2 text-right text-slate-600">{{ line.quantity }}</td>
                    <td class="px-3 py-2 text-right text-slate-600">{{ fmt(line.unit_price) }}</td>
                    <td class="px-3 py-2 text-right text-slate-600">{{ line.discount_pct }}%</td>
                    <td class="px-3 py-2 text-right text-slate-600">{{ line.tva_rate }}%</td>
                    <td class="px-3 py-2 text-right font-semibold text-slate-900 dark:text-white">{{ fmt(lineTotal(line)) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Footer Actions -->
          <div class="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-6 py-4 dark:border-slate-700">
            <div class="flex flex-wrap gap-2">
              <button
                v-if="detailInvoice.status === 'draft'"
                @click="confirmInvoice(detailInvoice.id)"
                class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-700 transition-colors"
              >
                <CheckCircle class="h-4 w-4" />
                Confirm
              </button>
              <button
                v-if="['confirmed', 'partially_paid'].includes(detailInvoice.status)"
                @click="openPayment(detailInvoice)"
                class="inline-flex items-center gap-2 rounded-lg bg-green-600 px-4 py-2 text-sm font-semibold text-white hover:bg-green-700 transition-colors"
              >
                <CreditCard class="h-4 w-4" />
                Record Payment
              </button>
              <button
                v-if="detailInvoice.po_id && detailInvoice.status !== 'cancelled'"
                @click="matchInvoice(detailInvoice.id)"
                class="inline-flex items-center gap-2 rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50 transition-colors dark:border-slate-600 dark:text-slate-300"
              >
                <Link2 class="h-4 w-4" />
                3-Way Match
              </button>
              <button
                v-if="['draft', 'confirmed'].includes(detailInvoice.status)"
                @click="cancelInvoice(detailInvoice.id)"
                class="inline-flex items-center gap-2 rounded-lg border border-red-200 px-4 py-2 text-sm font-medium text-red-600 hover:bg-red-50 transition-colors dark:border-red-800/40 dark:hover:bg-red-900/20"
              >
                <XCircle class="h-4 w-4" />
                Cancel Invoice
              </button>
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

    <!-- ── Payment Modal ───────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showPaymentModal && payingInvoice" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4 backdrop-blur-sm">
        <div class="w-full max-w-md rounded-2xl bg-white shadow-2xl dark:bg-slate-800">

          <div class="flex items-center justify-between border-b border-slate-200 px-6 py-4 dark:border-slate-700">
            <div class="flex items-center gap-3">
              <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-green-600">
                <CreditCard class="h-5 w-5 text-white" />
              </div>
              <div>
                <h2 class="text-lg font-semibold text-slate-900 dark:text-white">Record Payment</h2>
                <p class="text-xs text-slate-500">{{ payingInvoice.number }} &mdash; {{ payingInvoice.supplier_name }}</p>
              </div>
            </div>
            <button @click="showPaymentModal = false" class="rounded-lg p-2 text-slate-400 hover:bg-slate-100 transition-colors dark:hover:bg-slate-700">
              <X class="h-5 w-5" />
            </button>
          </div>

          <div class="space-y-4 p-6">
            <div class="rounded-lg bg-slate-50 p-4 dark:bg-slate-700/50">
              <div class="flex items-center justify-between text-sm">
                <span class="text-slate-500 dark:text-slate-400">Invoice Total</span>
                <span class="font-semibold text-slate-900 dark:text-white">{{ fmt(payingInvoice.total_amount) }} {{ payingInvoice.currency }}</span>
              </div>
              <div class="mt-1 flex items-center justify-between text-sm">
                <span class="text-slate-500 dark:text-slate-400">Already Paid</span>
                <span class="text-green-600">{{ fmt(payingInvoice.paid_amount) }}</span>
              </div>
              <div class="mt-1 flex items-center justify-between border-t border-slate-200 pt-1 text-sm dark:border-slate-600">
                <span class="font-semibold text-slate-700 dark:text-slate-300">Balance Due</span>
                <span class="font-bold text-red-600 dark:text-red-400">
                  {{ fmt(payingInvoice.balance_due ?? (payingInvoice.total_amount - payingInvoice.paid_amount)) }}
                </span>
              </div>
            </div>

            <div>
              <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                Payment Amount <span class="text-red-500">*</span>
              </label>
              <input
                v-model.number="paymentForm.amount"
                type="number"
                min="0.01"
                step="0.01"
                class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 focus:border-green-500 focus:outline-none focus:ring-1 focus:ring-green-500 dark:border-slate-600 dark:bg-slate-700 dark:text-white"
              />
            </div>
            <div>
              <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">Payment Date</label>
              <input
                v-model="paymentForm.payment_date"
                type="date"
                class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 focus:border-green-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white"
              />
            </div>
            <div>
              <label class="mb-1.5 block text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">Notes</label>
              <input
                v-model="paymentForm.notes"
                type="text"
                placeholder="Bank transfer, cheque no., etc."
                class="w-full rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 placeholder-slate-400 focus:border-green-500 focus:outline-none dark:border-slate-600 dark:bg-slate-700 dark:text-white dark:placeholder-slate-500"
              />
            </div>
          </div>

          <div class="flex items-center justify-end gap-3 border-t border-slate-200 px-6 py-4 dark:border-slate-700">
            <button
              @click="showPaymentModal = false"
              class="rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50 transition-colors dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
            >
              Cancel
            </button>
            <button
              @click="submitPayment"
              :disabled="saving || paymentForm.amount <= 0"
              class="inline-flex items-center gap-2 rounded-lg bg-green-600 px-5 py-2 text-sm font-semibold text-white shadow-sm hover:bg-green-700 disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
            >
              <DollarSign v-if="!saving" class="h-4 w-4" />
              <div v-else class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></div>
              {{ saving ? 'Processing...' : 'Record Payment' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
