<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { purchaseAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  Plus, X, Save, Search, RefreshCw, Edit2, Eye,
  Send, XCircle, ChevronDown, ChevronUp, Trash2,
  FileText, ArrowRightCircle, Calendar, Hash, Building2
} from 'lucide-vue-next'

// ─── Types ────────────────────────────────────────────────────────────────────

interface RFQLine {
  item_id?: string
  description: string
  quantity: number
  unit_price: number
  discount_pct: number
  tva_rate: number
  sort_order: number
}

interface RFQ {
  id: string
  number: string
  supplier_id: string
  supplier_name: string
  date: string
  deadline: string
  status: string
  subtotal: number
  tva_amount: number
  total_amount: number
  currency: string
  notes: string
  created_at: string
  lines?: RFQLine[]
}

interface Supplier { id: string; name: string; code: string }

const EMPTY_LINE = (i = 0): RFQLine => ({
  description: '', quantity: 1, unit_price: 0, discount_pct: 0, tva_rate: 19, sort_order: i
})

const STATUS_CFG: Record<string, { label: string; cls: string }> = {
  draft:     { label: 'Draft',     cls: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300' },
  sent:      { label: 'Sent',      cls: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300' },
  converted: { label: 'Converted', cls: 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300' },
  cancelled: { label: 'Cancelled', cls: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300' },
}

// ─── State ────────────────────────────────────────────────────────────────────

const app = useAppStore()
const rfqs = ref<RFQ[]>([])
const suppliers = ref<Supplier[]>([])
const loading = ref(true)
const saving = ref(false)
const showForm = ref(false)
const isEdit = ref(false)
const detailRFQ = ref<RFQ | null>(null)
const searchQ = ref('')
const filterStatus = ref('all')
const sortField = ref<keyof RFQ>('date')
const sortDir = ref<'asc' | 'desc'>('desc')

const form = ref<{
  id?: string; supplier_id: string; date: string; deadline: string
  notes: string; currency: string; lines: RFQLine[]
}>({
  supplier_id: '', date: new Date().toISOString().slice(0,10),
  deadline: '', notes: '', currency: 'DZD',
  lines: [EMPTY_LINE()]
})

// ─── Computed ─────────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = rfqs.value
  if (searchQ.value.trim()) {
    const q = searchQ.value.toLowerCase()
    list = list.filter(r => r.number.toLowerCase().includes(q) || r.supplier_name.toLowerCase().includes(q))
  }
  if (filterStatus.value !== 'all') list = list.filter(r => r.status === filterStatus.value)
  return [...list].sort((a, b) => {
    const av = (a as any)[sortField.value] ?? ''
    const bv = (b as any)[sortField.value] ?? ''
    const r = String(av).localeCompare(String(bv))
    return sortDir.value === 'asc' ? r : -r
  })
})

const formTotals = computed(() => {
  let sub = 0, tva = 0
  for (const l of form.value.lines) {
    const base = l.quantity * l.unit_price * (1 - l.discount_pct / 100)
    sub += base
    tva += base * l.tva_rate / 100
  }
  return { subtotal: sub, tva_amount: tva, total: sub + tva }
})

// ─── Methods ─────────────────────────────────────────────────────────────────

async function load() {
  loading.value = true
  try {
    const [rfqRes, suppRes] = await Promise.all([purchaseAPI.getRFQs(), purchaseAPI.getSuppliers()])
    rfqs.value = rfqRes.data || []
    suppliers.value = suppRes.data || []
  } catch {
    app.addToast('Failed to load RFQs', 'error')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = {
    supplier_id: '', date: new Date().toISOString().slice(0,10),
    deadline: '', notes: '', currency: 'DZD', lines: [EMPTY_LINE()]
  }
  isEdit.value = false
  showForm.value = true
}

async function openEdit(r: RFQ) {
  try {
    const res = await purchaseAPI.getRFQ(r.id)
    const data = res.data
    form.value = {
      id: data.id,
      supplier_id: data.supplier_id || '',
      date: data.date ? String(data.date).slice(0,10) : '',
      deadline: data.deadline || '',
      notes: data.notes || '',
      currency: data.currency || 'DZD',
      lines: (data.lines && data.lines.length) ? data.lines.map((l: any, i: number) => ({
        item_id: l.item_id, description: l.description || '',
        quantity: l.quantity || 1, unit_price: l.unit_price || 0,
        discount_pct: l.discount_pct || 0, tva_rate: l.tva_rate || 19, sort_order: i
      })) : [EMPTY_LINE()]
    }
    isEdit.value = true
    showForm.value = true
  } catch {
    app.addToast('Failed to load RFQ details', 'error')
  }
}

async function viewDetail(r: RFQ) {
  try {
    const res = await purchaseAPI.getRFQ(r.id)
    detailRFQ.value = res.data
  } catch {
    detailRFQ.value = r
  }
}

async function save() {
  if (!form.value.date) { app.addToast('Date is required', 'error'); return }
  saving.value = true
  try {
    const payload = { ...form.value, ...formTotals.value }
    if (isEdit.value && form.value.id) {
      await purchaseAPI.updateRFQ(form.value.id, payload)
      app.addToast('RFQ updated', 'success')
    } else {
      await purchaseAPI.createRFQ(payload)
      app.addToast('RFQ created', 'success')
    }
    showForm.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

async function sendRFQ(r: RFQ) {
  if (!confirm(`Send RFQ ${r.number} to supplier?`)) return
  try {
    await purchaseAPI.sendRFQ(r.id)
    app.addToast('RFQ sent', 'success')
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Failed to send', 'error')
  }
}

async function cancelRFQ(r: RFQ) {
  if (!confirm(`Cancel RFQ ${r.number}?`)) return
  try {
    await purchaseAPI.cancelRFQ(r.id)
    app.addToast('RFQ cancelled', 'success')
    await load()
  } catch {
    app.addToast('Failed to cancel', 'error')
  }
}

async function convertToOrder(r: RFQ) {
  if (!confirm(`Convert RFQ ${r.number} to Purchase Order?`)) return
  try {
    const res = await purchaseAPI.convertRFQToOrder(r.id)
    app.addToast(`Purchase Order ${res.data.number} created`, 'success')
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Conversion failed', 'error')
  }
}

function addLine() {
  form.value.lines.push(EMPTY_LINE(form.value.lines.length))
}

function removeLine(i: number) {
  if (form.value.lines.length > 1) form.value.lines.splice(i, 1)
}

function recalcLine(l: RFQLine) {
  // totals are all computed in formTotals
}

function toggleSort(f: keyof RFQ) {
  if (sortField.value === f) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortField.value = f; sortDir.value = 'asc' }
}

function statusCfg(s: string) {
  return STATUS_CFG[s] || { label: s, cls: 'bg-gray-100 text-gray-700' }
}

function fmt(n: number) {
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2 }).format(n)
}

onMounted(load)
</script>

<template>
  <div class="space-y-5">

    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white">Requests for Quotation</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Create and manage supplier RFQs</p>
      </div>
      <button @click="openCreate"
        class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary-600 text-white
               text-sm font-semibold hover:bg-primary-700 transition-colors shadow-sm">
        <Plus :size="16" /> New RFQ
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div v-for="stat in [
        { label:'Total',     val: rfqs.length,                           cls:'text-slate-900 dark:text-white' },
        { label:'Draft',     val: rfqs.filter(r=>r.status==='draft').length,     cls:'text-gray-700 dark:text-gray-300' },
        { label:'Sent',      val: rfqs.filter(r=>r.status==='sent').length,      cls:'text-blue-700 dark:text-blue-300' },
        { label:'Converted', val: rfqs.filter(r=>r.status==='converted').length, cls:'text-green-700 dark:text-green-300' },
      ]" :key="stat.label"
        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">{{ stat.label }}</p>
        <p class="text-2xl font-bold mt-1" :class="stat.cls">{{ stat.val }}</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
      <div class="flex flex-wrap gap-3 items-center">
        <div class="relative flex-1 min-w-[200px]">
          <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input v-model="searchQ" placeholder="Search number, supplier..."
            class="w-full pl-9 pr-4 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                   bg-slate-50 dark:bg-slate-700/50 text-slate-900 dark:text-white
                   focus:outline-none focus:ring-2 focus:ring-primary-500" />
        </div>
        <div class="flex rounded-lg overflow-hidden border border-slate-200 dark:border-slate-600">
          <button v-for="opt in ['all','draft','sent','converted','cancelled']" :key="opt"
            @click="filterStatus = opt"
            :class="['px-3 py-2 text-xs font-medium transition-colors capitalize',
              filterStatus === opt
                ? 'bg-primary-600 text-white'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700']">
            {{ opt }}
          </button>
        </div>
        <button @click="load" class="p-2 rounded-lg border border-slate-200 dark:border-slate-600 text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
          <RefreshCw :size="16" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center h-48">
        <RefreshCw :size="24" class="animate-spin text-primary-500" />
      </div>
      <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center h-48 text-slate-400">
        <FileText :size="40" class="mb-3 opacity-30" />
        <p class="font-medium">No RFQs found</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-slate-50 dark:bg-slate-700/50 border-b border-slate-200 dark:border-slate-600">
            <tr>
              <th v-for="[k,l] in [['number','Number'],['supplier_name','Supplier'],['date','Date'],['deadline','Deadline'],['status','Status'],['total_amount','Total']]"
                  :key="k"
                  @click="toggleSort(k as keyof RFQ)"
                  class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide cursor-pointer hover:text-slate-900 dark:hover:text-white select-none">
                <span class="inline-flex items-center gap-1">{{ l }}
                  <ChevronUp v-if="sortField===k && sortDir==='asc'" :size="11" />
                  <ChevronDown v-else-if="sortField===k" :size="11" />
                </span>
              </th>
              <th class="px-4 py-3 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-for="r in filtered" :key="r.id"
              class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
              <td class="px-4 py-3 font-semibold text-slate-900 dark:text-white">{{ r.number }}</td>
              <td class="px-4 py-3 text-slate-700 dark:text-slate-300">{{ r.supplier_name || '—' }}</td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400">{{ r.date ? String(r.date).slice(0,10) : '—' }}</td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400">{{ r.deadline || '—' }}</td>
              <td class="px-4 py-3">
                <span :class="['inline-flex px-2 py-0.5 rounded-full text-xs font-semibold', statusCfg(r.status).cls]">
                  {{ statusCfg(r.status).label }}
                </span>
              </td>
              <td class="px-4 py-3 font-semibold text-slate-900 dark:text-white">
                {{ fmt(r.total_amount || 0) }} <span class="text-xs font-normal text-slate-400">DZD</span>
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <button @click="viewDetail(r)" title="View" class="p-1.5 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"><Eye :size="15" /></button>
                  <button v-if="r.status === 'draft'" @click="openEdit(r)" title="Edit" class="p-1.5 rounded-lg text-slate-500 hover:bg-primary-50 dark:hover:bg-primary-900/30 hover:text-primary-600 transition-colors"><Edit2 :size="15" /></button>
                  <button v-if="r.status === 'draft'" @click="sendRFQ(r)" title="Send to supplier" class="p-1.5 rounded-lg text-slate-500 hover:bg-blue-50 dark:hover:bg-blue-900/30 hover:text-blue-600 transition-colors"><Send :size="15" /></button>
                  <button v-if="r.status === 'sent'" @click="convertToOrder(r)" title="Convert to PO" class="p-1.5 rounded-lg text-slate-500 hover:bg-green-50 dark:hover:bg-green-900/30 hover:text-green-600 transition-colors"><ArrowRightCircle :size="15" /></button>
                  <button v-if="['draft','sent'].includes(r.status)" @click="cancelRFQ(r)" title="Cancel" class="p-1.5 rounded-lg text-slate-500 hover:bg-red-50 dark:hover:bg-red-900/30 hover:text-red-600 transition-colors"><XCircle :size="15" /></button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-if="!loading" class="px-4 py-2 border-t border-slate-100 dark:border-slate-700 text-xs text-slate-500 dark:text-slate-400">
        {{ filtered.length }} of {{ rfqs.length }} RFQs
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showForm" class="fixed inset-0 z-50 flex items-start justify-center bg-black/50 backdrop-blur-sm p-4 overflow-y-auto">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-3xl my-8">
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700">
            <h2 class="text-lg font-bold text-slate-900 dark:text-white">{{ isEdit ? 'Edit RFQ' : 'New RFQ' }}</h2>
            <button @click="showForm = false" class="p-2 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700">
              <X :size="18" />
            </button>
          </div>

          <div class="p-6 space-y-5">
            <!-- Header fields -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div class="col-span-2">
                <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Supplier</label>
                <select v-model="form.supplier_id"
                  class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500">
                  <option value="">— No Supplier —</option>
                  <option v-for="s in suppliers" :key="s.id" :value="s.id">{{ s.name }}</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Date <span class="text-red-500">*</span></label>
                <input v-model="form.date" type="date"
                  class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Deadline</label>
                <input v-model="form.deadline" type="date"
                  class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
              </div>
            </div>

            <!-- Lines -->
            <div>
              <div class="flex items-center justify-between mb-2">
                <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300">Items / Services</h3>
                <button @click="addLine"
                  class="inline-flex items-center gap-1 px-3 py-1.5 text-xs font-semibold rounded-lg bg-primary-50 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300 hover:bg-primary-100 transition-colors">
                  <Plus :size="13" /> Add Line
                </button>
              </div>
              <div class="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-600">
                <table class="w-full text-xs">
                  <thead class="bg-slate-50 dark:bg-slate-700/50">
                    <tr>
                      <th class="px-3 py-2 text-left font-semibold text-slate-500 dark:text-slate-400">Description</th>
                      <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400 w-20">Qty</th>
                      <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400 w-28">Unit Price</th>
                      <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400 w-20">Disc%</th>
                      <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400 w-20">TVA%</th>
                      <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400 w-28">Total HT</th>
                      <th class="w-8"></th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
                    <tr v-for="(line, i) in form.lines" :key="i">
                      <td class="px-3 py-1.5">
                        <input v-model="line.description" placeholder="Description"
                          class="w-full px-2 py-1 rounded border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-primary-400" />
                      </td>
                      <td class="px-3 py-1.5">
                        <input v-model.number="line.quantity" type="number" min="0" step="0.001" @input="recalcLine(line)"
                          class="w-full px-2 py-1 rounded border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-right text-slate-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-primary-400" />
                      </td>
                      <td class="px-3 py-1.5">
                        <input v-model.number="line.unit_price" type="number" min="0" step="0.01" @input="recalcLine(line)"
                          class="w-full px-2 py-1 rounded border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-right text-slate-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-primary-400" />
                      </td>
                      <td class="px-3 py-1.5">
                        <input v-model.number="line.discount_pct" type="number" min="0" max="100" step="0.01"
                          class="w-full px-2 py-1 rounded border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-right text-slate-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-primary-400" />
                      </td>
                      <td class="px-3 py-1.5">
                        <select v-model.number="line.tva_rate"
                          class="w-full px-2 py-1 rounded border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-primary-400">
                          <option v-for="t in [0, 9, 19]" :key="t" :value="t">{{ t }}%</option>
                        </select>
                      </td>
                      <td class="px-3 py-1.5 text-right font-semibold text-slate-700 dark:text-slate-300">
                        {{ fmt(line.quantity * line.unit_price * (1 - line.discount_pct/100)) }}
                      </td>
                      <td class="px-1 py-1.5">
                        <button @click="removeLine(i)" :disabled="form.lines.length === 1"
                          class="p-1 rounded text-slate-400 hover:text-red-500 disabled:opacity-30 transition-colors">
                          <Trash2 :size="13" />
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <!-- Totals -->
              <div class="flex justify-end mt-3">
                <div class="w-64 space-y-1.5 text-sm">
                  <div class="flex justify-between text-slate-600 dark:text-slate-400">
                    <span>Subtotal HT</span>
                    <span class="font-semibold text-slate-900 dark:text-white">{{ fmt(formTotals.subtotal) }}</span>
                  </div>
                  <div class="flex justify-between text-slate-600 dark:text-slate-400">
                    <span>TVA</span>
                    <span class="font-semibold text-slate-900 dark:text-white">{{ fmt(formTotals.tva_amount) }}</span>
                  </div>
                  <div class="flex justify-between pt-2 border-t border-slate-200 dark:border-slate-600 font-bold">
                    <span class="text-slate-900 dark:text-white">Total TTC</span>
                    <span class="text-primary-600 dark:text-primary-400">{{ fmt(formTotals.total) }} DZD</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Notes -->
            <div>
              <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Notes</label>
              <textarea v-model="form.notes" rows="2"
                class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none" />
            </div>
          </div>

          <div class="flex justify-end gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-700">
            <button @click="showForm = false"
              class="px-4 py-2 text-sm font-medium rounded-lg border border-slate-200 dark:border-slate-600 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
            <button @click="save" :disabled="saving"
              class="inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-lg bg-primary-600 text-white hover:bg-primary-700 disabled:opacity-50 transition-colors shadow-sm">
              <RefreshCw v-if="saving" :size="15" class="animate-spin" />
              <Save v-else :size="15" />
              {{ saving ? 'Saving...' : 'Save RFQ' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Detail Modal -->
    <Teleport to="body">
      <div v-if="detailRFQ" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700 sticky top-0 bg-white dark:bg-slate-800 z-10">
            <div>
              <h2 class="text-lg font-bold text-slate-900 dark:text-white">{{ detailRFQ.number }}</h2>
              <span :class="['inline-flex px-2 py-0.5 rounded-full text-xs font-semibold mt-1', statusCfg(detailRFQ.status).cls]">
                {{ statusCfg(detailRFQ.status).label }}
              </span>
            </div>
            <button @click="detailRFQ = null" class="p-2 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700">
              <X :size="18" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-4 text-sm">
              <div><p class="text-xs text-slate-500 dark:text-slate-400">Supplier</p><p class="font-semibold text-slate-900 dark:text-white">{{ detailRFQ.supplier_name || '—' }}</p></div>
              <div><p class="text-xs text-slate-500 dark:text-slate-400">Date</p><p class="font-semibold text-slate-900 dark:text-white">{{ String(detailRFQ.date).slice(0,10) }}</p></div>
              <div><p class="text-xs text-slate-500 dark:text-slate-400">Deadline</p><p class="font-semibold text-slate-900 dark:text-white">{{ detailRFQ.deadline || '—' }}</p></div>
              <div><p class="text-xs text-slate-500 dark:text-slate-400">Currency</p><p class="font-semibold text-slate-900 dark:text-white">{{ detailRFQ.currency }}</p></div>
            </div>
            <div v-if="detailRFQ.lines && detailRFQ.lines.length" class="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-600">
              <table class="w-full text-xs">
                <thead class="bg-slate-50 dark:bg-slate-700/50">
                  <tr>
                    <th class="px-3 py-2 text-left font-semibold text-slate-500 dark:text-slate-400">#</th>
                    <th class="px-3 py-2 text-left font-semibold text-slate-500 dark:text-slate-400">Description</th>
                    <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400">Qty</th>
                    <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400">Unit Price</th>
                    <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400">TVA%</th>
                    <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400">Total</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
                  <tr v-for="(l, i) in detailRFQ.lines" :key="i">
                    <td class="px-3 py-2 text-slate-500 dark:text-slate-400">{{ i+1 }}</td>
                    <td class="px-3 py-2 text-slate-900 dark:text-white">{{ (l as any).description }}</td>
                    <td class="px-3 py-2 text-right">{{ (l as any).quantity }}</td>
                    <td class="px-3 py-2 text-right">{{ fmt((l as any).unit_price) }}</td>
                    <td class="px-3 py-2 text-right">{{ (l as any).tva_rate }}%</td>
                    <td class="px-3 py-2 text-right font-semibold text-slate-900 dark:text-white">{{ fmt((l as any).total || 0) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div class="flex justify-end">
              <div class="w-56 space-y-1 text-sm">
                <div class="flex justify-between text-slate-600 dark:text-slate-400"><span>HT</span><span>{{ fmt(detailRFQ.subtotal) }}</span></div>
                <div class="flex justify-between text-slate-600 dark:text-slate-400"><span>TVA</span><span>{{ fmt(detailRFQ.tva_amount) }}</span></div>
                <div class="flex justify-between pt-2 border-t border-slate-200 dark:border-slate-600 font-bold text-slate-900 dark:text-white">
                  <span>Total TTC</span><span class="text-primary-600 dark:text-primary-400">{{ fmt(detailRFQ.total_amount) }} DZD</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
