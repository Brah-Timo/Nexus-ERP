<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { salesAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  ArrowLeft, Printer, Send, CheckCircle, XCircle, CreditCard,
  ReceiptText, Building2, Calendar, Clock, AlertTriangle,
  FileText, DollarSign, TrendingUp, ChevronRight, Wallet,
  BarChart3, User, Phone, Mail, Globe, Hash, Package
} from 'lucide-vue-next'

// ─── Types ───────────────────────────────────────────────────────────────────

interface DocumentLine {
  id: string
  description: string
  quantity: number
  unit_price: number
  discount_pct: number
  tva_rate: number
  subtotal: number
  tva_amount: number
  total: number
  sort_order: number
}

interface SalesInvoice {
  id: string
  number: string
  customer_id: string
  customer_name: string
  order_id?: string
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
  lines: DocumentLine[]
  created_at: string
  updated_at: string
}

interface Customer {
  id: string
  name: string
  code: string
  nif: string
  nis: string
  rc: string
  art: string
  address: string
  city: string
  wilaya: string
  phone: string
  email: string
  website: string
  payment_terms: number
  balance: number
  tax_regime: string
}

// ─── Constants ────────────────────────────────────────────────────────────────

const STATUS_MAP: Record<string, { label: string; cls: string }> = {
  draft:          { label: 'Draft',          cls: 'bg-slate-100 text-slate-700 dark:bg-slate-700 dark:text-slate-200' },
  confirmed:      { label: 'Confirmed',      cls: 'bg-blue-100  text-blue-700  dark:bg-blue-900/50 dark:text-blue-200' },
  partially_paid: { label: 'Partial',        cls: 'bg-amber-100 text-amber-700 dark:bg-amber-900/50 dark:text-amber-200' },
  paid:           { label: 'Paid',           cls: 'bg-green-100 text-green-700 dark:bg-green-900/50 dark:text-green-200' },
  overdue:        { label: 'Overdue',        cls: 'bg-red-100   text-red-700   dark:bg-red-900/50 dark:text-red-200' },
  cancelled:      { label: 'Cancelled',      cls: 'bg-gray-100  text-gray-500  dark:bg-gray-800 dark:text-gray-400' },
}

// ─── State ───────────────────────────────────────────────────────────────────

const route  = useRoute()
const router = useRouter()
const app    = useAppStore()

const invoice    = ref<SalesInvoice | null>(null)
const customer   = ref<Customer | null>(null)
const loading    = ref(true)
const actionBusy = ref(false)

// Payment modal
const showPayment   = ref(false)
const paymentForm   = ref({ amount: 0, method: 'bank', payment_date: '', notes: '' })
const paymentBusy   = ref(false)

// ─── Computed ────────────────────────────────────────────────────────────────

const paidPct = computed(() => {
  if (!invoice.value || !invoice.value.total_amount) return 0
  return Math.min(100, Math.round((invoice.value.paid_amount / invoice.value.total_amount) * 100))
})

const isOverdue = computed(() => {
  if (!invoice.value?.due_date) return false
  return new Date(invoice.value.due_date) < new Date() &&
    !['paid', 'cancelled'].includes(invoice.value.status)
})

const overdueDays = computed(() => {
  if (!invoice.value?.due_date || !isOverdue.value) return 0
  const diff = Date.now() - new Date(invoice.value.due_date).getTime()
  return Math.floor(diff / (1000 * 60 * 60 * 24))
})

const canConfirm  = computed(() => invoice.value?.status === 'draft')
const canCancel   = computed(() => !['paid', 'cancelled'].includes(invoice.value?.status ?? ''))
const canPay      = computed(() => ['confirmed', 'partially_paid', 'overdue'].includes(invoice.value?.status ?? ''))

const lineSubtotal  = (l: DocumentLine) => l.quantity * l.unit_price * (1 - l.discount_pct / 100)
const lineTVA       = (l: DocumentLine) => lineSubtotal(l) * (l.tva_rate / 100)
const lineTotal     = (l: DocumentLine) => lineSubtotal(l) + lineTVA(l)

// ─── Methods ─────────────────────────────────────────────────────────────────

function fmt(v: number | undefined, decimals = 2): string {
  return (v ?? 0).toLocaleString('fr-DZ', { minimumFractionDigits: decimals, maximumFractionDigits: decimals })
}

function fmtDate(d: string | undefined): string {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('fr-DZ', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

async function load() {
  loading.value = true
  try {
    const id = route.params.id as string
    const [invRes, custsRes] = await Promise.all([
      salesAPI.getInvoice(id),
      salesAPI.getCustomers({ per_page: 500 })
    ])
    invoice.value = invRes.data
    const custList: Customer[] = custsRes.data || []
    customer.value = custList.find(c => c.id === invoice.value?.customer_id) ?? null
    if (!customer.value && invoice.value?.customer_id) {
      try {
        const cr = await salesAPI.getCustomer(invoice.value.customer_id)
        customer.value = cr.data
      } catch (_) {}
    }
  } catch (e: any) {
    app.error('Failed to load invoice: ' + (e?.response?.data?.error ?? e.message))
  } finally {
    loading.value = false
  }
}

async function confirmInvoice() {
  if (!invoice.value) return
  actionBusy.value = true
  try {
    await salesAPI.confirmInvoice(invoice.value.id)
    app.success('Invoice confirmed and posted')
    await load()
  } catch (e: any) {
    app.error(e?.response?.data?.error ?? 'Confirm failed')
  } finally {
    actionBusy.value = false
  }
}

async function cancelInvoice() {
  if (!invoice.value || !confirm('Cancel this invoice?')) return
  actionBusy.value = true
  try {
    await salesAPI.cancelInvoice(invoice.value.id)
    app.success('Invoice cancelled')
    await load()
  } catch (e: any) {
    app.error(e?.response?.data?.error ?? 'Cancel failed')
  } finally {
    actionBusy.value = false
  }
}

function openPayment() {
  if (!invoice.value) return
  paymentForm.value = {
    amount: invoice.value.balance_due,
    method: 'bank',
    payment_date: new Date().toISOString().slice(0, 10),
    notes: ''
  }
  showPayment.value = true
}

async function submitPayment() {
  if (!invoice.value) return
  paymentBusy.value = true
  try {
    await salesAPI.recordPayment(invoice.value.id, paymentForm.value)
    app.success('Payment recorded successfully')
    showPayment.value = false
    await load()
  } catch (e: any) {
    app.error(e?.response?.data?.error ?? 'Payment failed')
  } finally {
    paymentBusy.value = false
  }
}

function printInvoice() {
  window.print()
}

onMounted(load)
</script>

<template>
  <!-- ─── Loading ─────────────────────────────────────────────────────────── -->
  <div v-if="loading" class="flex items-center justify-center h-64">
    <div class="w-8 h-8 border-4 border-primary-500 border-t-transparent rounded-full animate-spin"></div>
  </div>

  <!-- ─── Not Found ──────────────────────────────────────────────────────── -->
  <div v-else-if="!invoice"
    class="flex flex-col items-center justify-center h-64 gap-3"
    :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
    <FileText class="w-12 h-12 opacity-30" />
    <p class="text-lg font-medium">Invoice not found</p>
    <button @click="router.back()"
      class="mt-2 px-4 py-2 rounded-lg text-sm font-medium bg-primary-600 text-white hover:bg-primary-700 transition-colors">
      Go back
    </button>
  </div>

  <!-- ─── Main Content ───────────────────────────────────────────────────── -->
  <div v-else class="space-y-6 animate-fade-in print:space-y-4">

    <!-- Header bar -->
    <div class="flex items-start justify-between gap-4 flex-wrap">
      <div class="flex items-center gap-3">
        <button @click="router.back()"
          class="p-2 rounded-lg transition-colors print:hidden"
          :class="app.darkMode ? 'hover:bg-slate-700 text-slate-400' : 'hover:bg-slate-100 text-slate-500'">
          <ArrowLeft class="w-5 h-5" />
        </button>
        <div>
          <div class="flex items-center gap-2 flex-wrap">
            <h1 class="text-2xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
              {{ invoice.number }}
            </h1>
            <span class="px-2.5 py-0.5 rounded-full text-xs font-semibold"
              :class="STATUS_MAP[invoice.status]?.cls ?? STATUS_MAP.draft.cls">
              {{ STATUS_MAP[invoice.status]?.label ?? invoice.status }}
            </span>
            <span v-if="isOverdue"
              class="flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-300">
              <AlertTriangle class="w-3 h-3" />
              {{ overdueDays }}d overdue
            </span>
          </div>
          <p class="text-sm mt-0.5" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
            Issued {{ fmtDate(invoice.date) }}
            <span v-if="invoice.due_date"> · Due {{ fmtDate(invoice.due_date) }}</span>
          </p>
        </div>
      </div>

      <!-- Action buttons -->
      <div class="flex items-center gap-2 flex-wrap print:hidden">
        <button @click="printInvoice"
          class="flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm font-medium border transition-colors"
          :class="app.darkMode
            ? 'border-slate-600 text-slate-300 hover:bg-slate-700'
            : 'border-slate-200 text-slate-600 hover:bg-slate-50'">
          <Printer class="w-4 h-4" /> Print
        </button>
        <button v-if="canConfirm" @click="confirmInvoice" :disabled="actionBusy"
          class="flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm font-medium bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50 transition-colors">
          <CheckCircle class="w-4 h-4" />
          {{ actionBusy ? 'Confirming…' : 'Confirm & Post' }}
        </button>
        <button v-if="canPay" @click="openPayment"
          class="flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm font-medium bg-green-600 text-white hover:bg-green-700 transition-colors">
          <CreditCard class="w-4 h-4" /> Record Payment
        </button>
        <button v-if="canCancel" @click="cancelInvoice" :disabled="actionBusy"
          class="flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm font-medium bg-red-600 text-white hover:bg-red-700 disabled:opacity-50 transition-colors">
          <XCircle class="w-4 h-4" /> Cancel
        </button>
      </div>
    </div>

    <!-- ─── KPI bar ──────────────────────────────────────────────────────── -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
      <!-- Total TTC -->
      <div class="rounded-xl p-4 border"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200 shadow-card'">
        <div class="flex items-center gap-2 mb-1">
          <ReceiptText class="w-4 h-4 text-primary-500" />
          <span class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Total TTC</span>
        </div>
        <p class="text-xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
          {{ fmt(invoice.total_amount) }}
        </p>
        <p class="text-xs mt-0.5" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ invoice.currency }}</p>
      </div>
      <!-- Paid -->
      <div class="rounded-xl p-4 border"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200 shadow-card'">
        <div class="flex items-center gap-2 mb-1">
          <CheckCircle class="w-4 h-4 text-green-500" />
          <span class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Paid</span>
        </div>
        <p class="text-xl font-bold text-green-600 dark:text-green-400">
          {{ fmt(invoice.paid_amount) }}
        </p>
        <p class="text-xs mt-0.5" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ paidPct }}% of total</p>
      </div>
      <!-- Balance Due -->
      <div class="rounded-xl p-4 border"
        :class="[app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200 shadow-card',
          invoice.balance_due > 0 ? 'ring-1 ring-orange-400/50' : '']">
        <div class="flex items-center gap-2 mb-1">
          <Wallet class="w-4 h-4" :class="invoice.balance_due > 0 ? 'text-orange-500' : 'text-slate-400'" />
          <span class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Balance Due</span>
        </div>
        <p class="text-xl font-bold"
          :class="invoice.balance_due > 0
            ? (isOverdue ? 'text-red-600 dark:text-red-400' : 'text-orange-600 dark:text-orange-400')
            : 'text-green-600 dark:text-green-400'">
          {{ fmt(invoice.balance_due) }}
        </p>
        <p class="text-xs mt-0.5" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ invoice.currency }}</p>
      </div>
      <!-- TVA -->
      <div class="rounded-xl p-4 border"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200 shadow-card'">
        <div class="flex items-center gap-2 mb-1">
          <BarChart3 class="w-4 h-4 text-violet-500" />
          <span class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">TVA (19%)</span>
        </div>
        <p class="text-xl font-bold text-violet-600 dark:text-violet-400">
          {{ fmt(invoice.tva_amount) }}
        </p>
        <p class="text-xs mt-0.5" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">HT: {{ fmt(invoice.subtotal) }}</p>
      </div>
    </div>

    <!-- ─── Payment progress bar ─────────────────────────────────────────── -->
    <div v-if="invoice.status !== 'draft' && invoice.status !== 'cancelled'"
      class="rounded-xl p-4 border"
      :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200 shadow-card'">
      <div class="flex justify-between items-center mb-2">
        <span class="text-sm font-medium" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">
          Payment Progress
        </span>
        <span class="text-sm font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
          {{ paidPct }}%
        </span>
      </div>
      <div class="w-full h-3 rounded-full overflow-hidden"
        :class="app.darkMode ? 'bg-slate-700' : 'bg-slate-100'">
        <div class="h-full rounded-full transition-all duration-500"
          :style="{ width: paidPct + '%' }"
          :class="paidPct >= 100 ? 'bg-green-500' : isOverdue ? 'bg-red-500' : 'bg-primary-500'">
        </div>
      </div>
      <div class="flex justify-between text-xs mt-1.5"
        :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">
        <span>Paid: {{ fmt(invoice.paid_amount) }} {{ invoice.currency }}</span>
        <span>Remaining: {{ fmt(invoice.balance_due) }} {{ invoice.currency }}</span>
      </div>
    </div>

    <!-- ─── Two-column layout: invoice + customer ─────────────────────────── -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">

      <!-- Left: invoice details (2/3) -->
      <div class="lg:col-span-2 space-y-6">

        <!-- Line items table -->
        <div class="rounded-xl border overflow-hidden"
          :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200 shadow-card'">
          <div class="px-5 py-3 border-b flex items-center gap-2"
            :class="app.darkMode ? 'border-slate-700' : 'border-slate-100'">
            <Package class="w-4 h-4 text-primary-500" />
            <h2 class="text-sm font-semibold" :class="app.darkMode ? 'text-white' : 'text-slate-800'">
              Line Items
            </h2>
            <span class="ml-auto text-xs" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">
              {{ invoice.lines?.length ?? 0 }} line{{ (invoice.lines?.length ?? 0) !== 1 ? 's' : '' }}
            </span>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr :class="app.darkMode ? 'bg-slate-700/50 text-slate-400' : 'bg-slate-50 text-slate-500'">
                  <th class="text-left px-4 py-2.5 font-medium">#</th>
                  <th class="text-left px-4 py-2.5 font-medium">Description</th>
                  <th class="text-right px-4 py-2.5 font-medium">Qty</th>
                  <th class="text-right px-4 py-2.5 font-medium">Unit Price</th>
                  <th class="text-right px-4 py-2.5 font-medium">Disc%</th>
                  <th class="text-right px-4 py-2.5 font-medium">TVA%</th>
                  <th class="text-right px-4 py-2.5 font-medium">HT</th>
                  <th class="text-right px-4 py-2.5 font-medium">TTC</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(l, i) in (invoice.lines ?? [])" :key="l.id"
                  class="border-t transition-colors"
                  :class="[
                    app.darkMode ? 'border-slate-700 hover:bg-slate-700/30' : 'border-slate-100 hover:bg-slate-50/70'
                  ]">
                  <td class="px-4 py-3" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ i + 1 }}</td>
                  <td class="px-4 py-3 max-w-[200px]">
                    <span class="font-medium block truncate" :class="app.darkMode ? 'text-slate-200' : 'text-slate-800'">
                      {{ l.description || '—' }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-right" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">
                    {{ l.quantity }}
                  </td>
                  <td class="px-4 py-3 text-right tabular-nums" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">
                    {{ fmt(l.unit_price) }}
                  </td>
                  <td class="px-4 py-3 text-right" :class="l.discount_pct > 0 ? 'text-amber-500' : (app.darkMode ? 'text-slate-500' : 'text-slate-400')">
                    {{ l.discount_pct > 0 ? l.discount_pct + '%' : '—' }}
                  </td>
                  <td class="px-4 py-3 text-right" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
                    {{ l.tva_rate }}%
                  </td>
                  <td class="px-4 py-3 text-right tabular-nums" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">
                    {{ fmt(lineSubtotal(l)) }}
                  </td>
                  <td class="px-4 py-3 text-right tabular-nums font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
                    {{ fmt(lineTotal(l)) }}
                  </td>
                </tr>
                <tr v-if="!invoice.lines?.length">
                  <td colspan="8" class="px-4 py-8 text-center text-sm"
                    :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">
                    No line items
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Totals footer -->
          <div class="px-5 py-4 border-t space-y-1.5"
            :class="app.darkMode ? 'border-slate-700 bg-slate-700/20' : 'border-slate-100 bg-slate-50/60'">
            <div class="flex justify-between text-sm"
              :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
              <span>Subtotal HT</span>
              <span class="tabular-nums">{{ fmt(invoice.subtotal) }} {{ invoice.currency }}</span>
            </div>
            <div v-if="invoice.discount_amount > 0" class="flex justify-between text-sm text-amber-600 dark:text-amber-400">
              <span>Discount</span>
              <span class="tabular-nums">- {{ fmt(invoice.discount_amount) }} {{ invoice.currency }}</span>
            </div>
            <div class="flex justify-between text-sm"
              :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
              <span>TVA (19%)</span>
              <span class="tabular-nums">{{ fmt(invoice.tva_amount) }} {{ invoice.currency }}</span>
            </div>
            <div v-if="invoice.stamp_tax > 0" class="flex justify-between text-sm"
              :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
              <span>Stamp Tax</span>
              <span class="tabular-nums">{{ fmt(invoice.stamp_tax) }} {{ invoice.currency }}</span>
            </div>
            <div class="flex justify-between text-base font-bold pt-2 border-t"
              :class="[
                app.darkMode ? 'border-slate-600 text-white' : 'border-slate-200 text-slate-900'
              ]">
              <span>Total TTC</span>
              <span class="tabular-nums text-primary-600 dark:text-primary-400">
                {{ fmt(invoice.total_amount) }} {{ invoice.currency }}
              </span>
            </div>
          </div>
        </div>

        <!-- Notes -->
        <div v-if="invoice.notes" class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200 shadow-card'">
          <div class="flex items-center gap-2 mb-2">
            <FileText class="w-4 h-4" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'" />
            <h3 class="text-sm font-semibold" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Notes</h3>
          </div>
          <p class="text-sm leading-relaxed" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
            {{ invoice.notes }}
          </p>
        </div>

        <!-- Journal Entry link -->
        <div v-if="invoice.journal_entry_id" class="rounded-xl border p-4 flex items-center gap-3"
          :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200 shadow-card'">
          <div class="w-8 h-8 rounded-lg bg-violet-100 dark:bg-violet-900/30 flex items-center justify-center flex-shrink-0">
            <Hash class="w-4 h-4 text-violet-600 dark:text-violet-400" />
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-xs" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Journal Entry</p>
            <p class="text-sm font-mono font-medium truncate" :class="app.darkMode ? 'text-slate-200' : 'text-slate-700'">
              {{ invoice.journal_entry_id }}
            </p>
          </div>
          <ChevronRight class="w-4 h-4 flex-shrink-0" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'" />
        </div>
      </div>

      <!-- Right: meta info (1/3) -->
      <div class="space-y-5">

        <!-- Invoice metadata -->
        <div class="rounded-xl border divide-y"
          :class="app.darkMode
            ? 'bg-slate-800 border-slate-700 divide-slate-700'
            : 'bg-white border-slate-200 shadow-card divide-slate-100'">
          <div class="px-4 py-3">
            <p class="text-xs font-semibold uppercase tracking-wider mb-1"
              :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Invoice Info</p>
          </div>
          <div class="px-4 py-3 flex justify-between items-center">
            <span class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Number</span>
            <span class="text-sm font-mono font-medium" :class="app.darkMode ? 'text-slate-200' : 'text-slate-800'">
              {{ invoice.number }}
            </span>
          </div>
          <div class="px-4 py-3 flex justify-between items-center">
            <span class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Issue Date</span>
            <span class="text-sm" :class="app.darkMode ? 'text-slate-200' : 'text-slate-800'">{{ fmtDate(invoice.date) }}</span>
          </div>
          <div class="px-4 py-3 flex justify-between items-center">
            <span class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Due Date</span>
            <span class="text-sm font-medium"
              :class="isOverdue
                ? 'text-red-600 dark:text-red-400'
                : (app.darkMode ? 'text-slate-200' : 'text-slate-800')">
              {{ fmtDate(invoice.due_date) }}
            </span>
          </div>
          <div v-if="invoice.order_id" class="px-4 py-3 flex justify-between items-center">
            <span class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Sales Order</span>
            <span class="text-xs font-mono" :class="app.darkMode ? 'text-slate-300' : 'text-slate-600'">
              {{ invoice.order_id.slice(0, 8) }}…
            </span>
          </div>
          <div class="px-4 py-3 flex justify-between items-center">
            <span class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Currency</span>
            <span class="text-sm font-medium" :class="app.darkMode ? 'text-slate-200' : 'text-slate-800'">
              {{ invoice.currency }}
            </span>
          </div>
          <div class="px-4 py-3 flex justify-between items-center">
            <span class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Created</span>
            <span class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
              {{ fmtDate(invoice.created_at) }}
            </span>
          </div>
        </div>

        <!-- Customer card -->
        <div class="rounded-xl border"
          :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200 shadow-card'">
          <div class="px-4 py-3 border-b flex items-center gap-2"
            :class="app.darkMode ? 'border-slate-700' : 'border-slate-100'">
            <Building2 class="w-4 h-4 text-primary-500" />
            <h3 class="text-sm font-semibold" :class="app.darkMode ? 'text-white' : 'text-slate-800'">
              Customer
            </h3>
          </div>
          <div class="p-4 space-y-3">
            <div>
              <p class="font-semibold text-base" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
                {{ invoice.customer_name || customer?.name || '—' }}
              </p>
              <p v-if="customer?.code" class="text-xs font-mono mt-0.5"
                :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">
                {{ customer.code }}
              </p>
            </div>
            <div v-if="customer" class="space-y-2 text-sm">
              <div v-if="customer.address || customer.city" class="flex items-start gap-2"
                :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
                <Globe class="w-3.5 h-3.5 mt-0.5 flex-shrink-0" />
                <span>{{ [customer.address, customer.city, customer.wilaya].filter(Boolean).join(', ') }}</span>
              </div>
              <div v-if="customer.phone" class="flex items-center gap-2"
                :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
                <Phone class="w-3.5 h-3.5 flex-shrink-0" />
                <span>{{ customer.phone }}</span>
              </div>
              <div v-if="customer.email" class="flex items-center gap-2"
                :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
                <Mail class="w-3.5 h-3.5 flex-shrink-0" />
                <span class="truncate">{{ customer.email }}</span>
              </div>
            </div>

            <!-- Fiscal IDs -->
            <div v-if="customer && (customer.nif || customer.nis || customer.rc)"
              class="pt-3 border-t space-y-1.5"
              :class="app.darkMode ? 'border-slate-700' : 'border-slate-100'">
              <p class="text-xs font-semibold uppercase tracking-wider"
                :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Fiscal IDs</p>
              <div v-if="customer.nif" class="flex justify-between text-xs"
                :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
                <span>NIF</span><span class="font-mono">{{ customer.nif }}</span>
              </div>
              <div v-if="customer.nis" class="flex justify-between text-xs"
                :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
                <span>NIS</span><span class="font-mono">{{ customer.nis }}</span>
              </div>
              <div v-if="customer.rc" class="flex justify-between text-xs"
                :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
                <span>RC</span><span class="font-mono">{{ customer.rc }}</span>
              </div>
              <div v-if="customer.art" class="flex justify-between text-xs"
                :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">
                <span>ART</span><span class="font-mono">{{ customer.art }}</span>
              </div>
            </div>

            <!-- Balance -->
            <div v-if="customer"
              class="pt-3 border-t"
              :class="app.darkMode ? 'border-slate-700' : 'border-slate-100'">
              <div class="flex justify-between items-center">
                <span class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Account Balance</span>
                <span class="text-sm font-bold"
                  :class="(customer.balance || 0) > 0
                    ? 'text-orange-600 dark:text-orange-400'
                    : 'text-green-600 dark:text-green-400'">
                  {{ fmt(customer.balance) }} DZD
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Quick actions (mobile print, etc.) -->
        <div class="rounded-xl border p-4 print:hidden"
          :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200 shadow-card'">
          <p class="text-xs font-semibold uppercase tracking-wider mb-3"
            :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Quick Actions</p>
          <div class="space-y-2">
            <button v-if="canPay" @click="openPayment"
              class="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium bg-green-600/10 text-green-700 dark:text-green-400 hover:bg-green-600/20 transition-colors">
              <CreditCard class="w-4 h-4" /> Record Payment
            </button>
            <button v-if="canConfirm" @click="confirmInvoice"
              class="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium bg-blue-600/10 text-blue-700 dark:text-blue-400 hover:bg-blue-600/20 transition-colors">
              <CheckCircle class="w-4 h-4" /> Confirm & Post
            </button>
            <button @click="printInvoice"
              class="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-colors"
              :class="app.darkMode
                ? 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                : 'bg-slate-100 text-slate-600 hover:bg-slate-200'">
              <Printer class="w-4 h-4" /> Print Invoice
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- ─── Payment Modal ─────────────────────────────────────────────────────── -->
  <Teleport to="body">
    <div v-if="showPayment" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <!-- Backdrop -->
      <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showPayment = false"></div>
      <!-- Panel -->
      <div class="relative w-full max-w-md rounded-2xl shadow-2xl animate-scale-in"
        :class="app.darkMode ? 'bg-slate-800 border border-slate-700' : 'bg-white'">

        <!-- Header -->
        <div class="flex items-center justify-between px-6 py-4 border-b"
          :class="app.darkMode ? 'border-slate-700' : 'border-slate-100'">
          <div class="flex items-center gap-2">
            <CreditCard class="w-5 h-5 text-green-500" />
            <h2 class="text-base font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
              Record Payment
            </h2>
          </div>
          <button @click="showPayment = false"
            class="p-1.5 rounded-lg transition-colors"
            :class="app.darkMode ? 'hover:bg-slate-700 text-slate-400' : 'hover:bg-slate-100 text-slate-500'">
            <XCircle class="w-4 h-4" />
          </button>
        </div>

        <!-- Body -->
        <div class="px-6 py-5 space-y-4">
          <!-- Invoice summary -->
          <div class="rounded-lg p-3 text-sm"
            :class="app.darkMode ? 'bg-slate-700/50' : 'bg-slate-50'">
            <div class="flex justify-between">
              <span :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Invoice</span>
              <span class="font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">{{ invoice?.number }}</span>
            </div>
            <div class="flex justify-between mt-1">
              <span :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Balance Due</span>
              <span class="font-bold text-orange-600 dark:text-orange-400">{{ fmt(invoice?.balance_due) }} {{ invoice?.currency }}</span>
            </div>
          </div>

          <!-- Amount -->
          <div>
            <label class="block text-xs font-medium mb-1.5"
              :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Amount</label>
            <input v-model.number="paymentForm.amount" type="number" min="0.01" step="0.01"
              class="w-full px-3 py-2 rounded-lg border text-sm transition-colors focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              :class="app.darkMode
                ? 'bg-slate-700 border-slate-600 text-white placeholder-slate-500'
                : 'bg-white border-slate-200 text-slate-900'" />
          </div>

          <!-- Method -->
          <div>
            <label class="block text-xs font-medium mb-1.5"
              :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Payment Method</label>
            <select v-model="paymentForm.method"
              class="w-full px-3 py-2 rounded-lg border text-sm transition-colors focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              :class="app.darkMode
                ? 'bg-slate-700 border-slate-600 text-white'
                : 'bg-white border-slate-200 text-slate-900'">
              <option value="bank">Bank Transfer (Virement)</option>
              <option value="cash">Cash (Especes)</option>
              <option value="cheque">Cheque</option>
              <option value="card">Card</option>
            </select>
          </div>

          <!-- Date -->
          <div>
            <label class="block text-xs font-medium mb-1.5"
              :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Payment Date</label>
            <input v-model="paymentForm.payment_date" type="date"
              class="w-full px-3 py-2 rounded-lg border text-sm transition-colors focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              :class="app.darkMode
                ? 'bg-slate-700 border-slate-600 text-white'
                : 'bg-white border-slate-200 text-slate-900'" />
          </div>

          <!-- Notes -->
          <div>
            <label class="block text-xs font-medium mb-1.5"
              :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Notes (optional)</label>
            <input v-model="paymentForm.notes" type="text" placeholder="Reference, cheque #, etc."
              class="w-full px-3 py-2 rounded-lg border text-sm transition-colors focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              :class="app.darkMode
                ? 'bg-slate-700 border-slate-600 text-white placeholder-slate-500'
                : 'bg-white border-slate-200 text-slate-900'" />
          </div>
        </div>

        <!-- Footer -->
        <div class="flex justify-end gap-3 px-6 py-4 border-t"
          :class="app.darkMode ? 'border-slate-700' : 'border-slate-100'">
          <button @click="showPayment = false"
            class="px-4 py-2 rounded-lg text-sm font-medium border transition-colors"
            :class="app.darkMode
              ? 'border-slate-600 text-slate-300 hover:bg-slate-700'
              : 'border-slate-200 text-slate-600 hover:bg-slate-50'">
            Cancel
          </button>
          <button @click="submitPayment" :disabled="paymentBusy || paymentForm.amount <= 0"
            class="flex items-center gap-2 px-5 py-2 rounded-lg text-sm font-semibold bg-green-600 text-white hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors">
            <span v-if="paymentBusy" class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
            <CreditCard v-else class="w-4 h-4" />
            {{ paymentBusy ? 'Recording…' : 'Record Payment' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
@media print {
  .print\:hidden { display: none !important; }
  .print\:space-y-4 > * + * { margin-top: 1rem; }
}
</style>
