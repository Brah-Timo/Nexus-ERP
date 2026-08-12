<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { purchaseAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  Plus, X, Save, Search, RefreshCw, Edit2, Eye,
  CheckCircle, XCircle, ChevronDown, ChevronUp, Trash2,
  ShoppingCart, Package, TrendingUp, Clock, Send
} from 'lucide-vue-next'

interface POLine {
  item_id?: string
  description: string
  quantity: number
  unit_price: number
  discount_pct: number
  tva_rate: number
  sort_order: number
}

interface PurchaseOrder {
  id: string
  number: string
  supplier_id: string
  supplier_name: string
  rfq_id: string
  date: string
  expected_date: string
  status: string
  subtotal: number
  discount_amount: number
  tva_amount: number
  total_amount: number
  received_amount: number
  currency: string
  notes: string
  created_at: string
  lines?: any[]
}

interface Supplier { id: string; name: string; code: string }

const EMPTY_LINE = (i = 0): POLine => ({
  description: '', quantity: 1, unit_price: 0, discount_pct: 0, tva_rate: 19, sort_order: i
})

const STATUS_CFG: Record<string, { label: string; cls: string }> = {
  draft:              { label: 'Draft',               cls: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300' },
  pending_approval:   { label: 'Pending Approval',    cls: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300' },
  approved:           { label: 'Approved',            cls: 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300' },
  partially_received: { label: 'Partial Receipt',     cls: 'bg-purple-100 text-purple-700 dark:bg-purple-900/40 dark:text-purple-300' },
  received:           { label: 'Received',            cls: 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300' },
  cancelled:          { label: 'Cancelled',           cls: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300' },
}

const app = useAppStore()
const orders = ref<PurchaseOrder[]>([])
const suppliers = ref<Supplier[]>([])
const loading = ref(true)
const saving = ref(false)
const showForm = ref(false)
const isEdit = ref(false)
const detailOrder = ref<PurchaseOrder | null>(null)
const searchQ = ref('')
const filterStatus = ref('all')
const sortField = ref<keyof PurchaseOrder>('date')
const sortDir = ref<'asc' | 'desc'>('desc')

const form = ref<{
  id?: string; supplier_id: string; date: string; expected_date: string
  notes: string; currency: string; lines: POLine[]
}>({
  supplier_id: '', date: new Date().toISOString().slice(0,10),
  expected_date: '', notes: '', currency: 'DZD', lines: [EMPTY_LINE()]
})

const filtered = computed(() => {
  let list = orders.value
  if (searchQ.value.trim()) {
    const q = searchQ.value.toLowerCase()
    list = list.filter(o => o.number.toLowerCase().includes(q) || o.supplier_name.toLowerCase().includes(q))
  }
  if (filterStatus.value !== 'all') list = list.filter(o => o.status === filterStatus.value)
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
    sub += base; tva += base * l.tva_rate / 100
  }
  return { subtotal: sub, tva_amount: tva, total: sub + tva }
})

const stats = computed(() => ({
  total: orders.value.length,
  draft: orders.value.filter(o => o.status === 'draft').length,
  approved: orders.value.filter(o => o.status === 'approved').length,
  totalValue: orders.value.filter(o => o.status !== 'cancelled')
    .reduce((s, o) => s + (o.total_amount || 0), 0)
}))

async function load() {
  loading.value = true
  try {
    const [ordRes, suppRes] = await Promise.all([purchaseAPI.getOrders(), purchaseAPI.getSuppliers()])
    orders.value = ordRes.data || []
    suppliers.value = suppRes.data || []
  } catch { app.addToast('Failed to load orders', 'error') }
  finally { loading.value = false }
}

function openCreate() {
  form.value = {
    supplier_id: '', date: new Date().toISOString().slice(0,10),
    expected_date: '', notes: '', currency: 'DZD', lines: [EMPTY_LINE()]
  }
  isEdit.value = false
  showForm.value = true
}

async function openEdit(o: PurchaseOrder) {
  try {
    const res = await purchaseAPI.getOrder(o.id)
    const data = res.data
    form.value = {
      id: data.id,
      supplier_id: data.supplier_id || '',
      date: data.date ? String(data.date).slice(0,10) : '',
      expected_date: data.expected_date || '',
      notes: data.notes || '', currency: data.currency || 'DZD',
      lines: (data.lines && data.lines.length) ? data.lines.map((l: any, i: number) => ({
        item_id: l.item_id, description: l.description || l.item_name || '',
        quantity: Number(l.quantity) || 1, unit_price: Number(l.unit_price) || 0,
        discount_pct: Number(l.discount_pct) || 0, tva_rate: Number(l.tva_rate) || 19, sort_order: i
      })) : [EMPTY_LINE()]
    }
    isEdit.value = true
    showForm.value = true
  } catch { app.addToast('Failed to load order details', 'error') }
}

async function viewDetail(o: PurchaseOrder) {
  try {
    const res = await purchaseAPI.getOrder(o.id)
    detailOrder.value = res.data
  } catch { detailOrder.value = o }
}

async function save() {
  if (!form.value.supplier_id) { app.addToast('Supplier is required', 'error'); return }
  saving.value = true
  try {
    const payload = { ...form.value, ...formTotals.value }
    if (isEdit.value && form.value.id) {
      await purchaseAPI.updateOrder(form.value.id, payload)
      app.addToast('Order updated', 'success')
    } else {
      await purchaseAPI.createOrder(payload)
      app.addToast('Order created', 'success')
    }
    showForm.value = false
    await load()
  } catch (e: any) { app.addToast(e?.response?.data?.error || 'Save failed', 'error') }
  finally { saving.value = false }
}

async function approve(o: PurchaseOrder) {
  if (!confirm(`Approve order ${o.number}?`)) return
  try {
    await purchaseAPI.approveOrder(o.id)
    app.addToast('Order approved', 'success')
    await load()
  } catch (e: any) { app.addToast(e?.response?.data?.error || 'Approve failed', 'error') }
}

async function cancelOrder(o: PurchaseOrder) {
  if (!confirm(`Cancel order ${o.number}?`)) return
  try {
    await purchaseAPI.cancelOrder(o.id)
    app.addToast('Order cancelled', 'success')
    await load()
  } catch (e: any) { app.addToast(e?.response?.data?.error || 'Cancel failed', 'error') }
}

function addLine() { form.value.lines.push(EMPTY_LINE(form.value.lines.length)) }
function removeLine(i: number) { if (form.value.lines.length > 1) form.value.lines.splice(i, 1) }
function toggleSort(f: keyof PurchaseOrder) {
  if (sortField.value === f) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortField.value = f; sortDir.value = 'asc' }
}
function statusCfg(s: string) { return STATUS_CFG[s] || { label: s, cls: 'bg-gray-100 text-gray-700' } }
function fmt(n: number) { return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2 }).format(n) }

onMounted(load)
</script>

<template>
  <div class="space-y-5">

    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white">Purchase Orders</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Manage procurement purchase orders</p>
      </div>
      <button @click="openCreate"
        class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary-600 text-white text-sm font-semibold hover:bg-primary-700 transition-colors shadow-sm">
        <Plus :size="16" /> New Order
      </button>
    </div>

    <!-- KPI -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total Orders</p>
        <p class="text-2xl font-bold text-slate-900 dark:text-white mt-1">{{ stats.total }}</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Draft</p>
        <p class="text-2xl font-bold text-gray-600 dark:text-gray-300 mt-1">{{ stats.draft }}</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Approved</p>
        <p class="text-2xl font-bold text-blue-600 dark:text-blue-400 mt-1">{{ stats.approved }}</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total Value</p>
        <p class="text-2xl font-bold text-primary-600 dark:text-primary-400 mt-1">{{ fmt(stats.totalValue) }}</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
      <div class="flex flex-wrap gap-3 items-center">
        <div class="relative flex-1 min-w-[200px]">
          <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input v-model="searchQ" placeholder="Search number, supplier..."
            class="w-full pl-9 pr-4 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700/50 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
        </div>
        <div class="flex rounded-lg overflow-hidden border border-slate-200 dark:border-slate-600">
          <button v-for="opt in ['all','draft','approved','received','cancelled']" :key="opt"
            @click="filterStatus = opt"
            :class="['px-3 py-2 text-xs font-medium transition-colors capitalize',
              filterStatus === opt ? 'bg-primary-600 text-white' : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700']">
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
        <ShoppingCart :size="40" class="mb-3 opacity-30" />
        <p class="font-medium">No purchase orders found</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-slate-50 dark:bg-slate-700/50 border-b border-slate-200 dark:border-slate-600">
            <tr>
              <th v-for="[k,l] in [['number','Number'],['supplier_name','Supplier'],['date','Date'],['expected_date','Expected'],['status','Status'],['total_amount','Total TTC']]"
                  :key="k" @click="toggleSort(k as keyof PurchaseOrder)"
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
            <tr v-for="o in filtered" :key="o.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
              <td class="px-4 py-3 font-semibold text-slate-900 dark:text-white">{{ o.number }}</td>
              <td class="px-4 py-3 text-slate-700 dark:text-slate-300">{{ o.supplier_name || '—' }}</td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400">{{ o.date ? String(o.date).slice(0,10) : '—' }}</td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400">{{ o.expected_date || '—' }}</td>
              <td class="px-4 py-3">
                <span :class="['inline-flex px-2 py-0.5 rounded-full text-xs font-semibold', statusCfg(o.status).cls]">
                  {{ statusCfg(o.status).label }}
                </span>
              </td>
              <td class="px-4 py-3 font-semibold text-slate-900 dark:text-white">
                {{ fmt(o.total_amount || 0) }} <span class="text-xs font-normal text-slate-400">DZD</span>
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <button @click="viewDetail(o)" title="View" class="p-1.5 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"><Eye :size="15" /></button>
                  <button v-if="o.status === 'draft'" @click="openEdit(o)" title="Edit" class="p-1.5 rounded-lg text-slate-500 hover:bg-primary-50 dark:hover:bg-primary-900/30 hover:text-primary-600 transition-colors"><Edit2 :size="15" /></button>
                  <button v-if="o.status === 'draft'" @click="approve(o)" title="Approve" class="p-1.5 rounded-lg text-slate-500 hover:bg-green-50 dark:hover:bg-green-900/30 hover:text-green-600 transition-colors"><CheckCircle :size="15" /></button>
                  <button v-if="!['received','cancelled'].includes(o.status)" @click="cancelOrder(o)" title="Cancel" class="p-1.5 rounded-lg text-slate-500 hover:bg-red-50 dark:hover:bg-red-900/30 hover:text-red-600 transition-colors"><XCircle :size="15" /></button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-if="!loading" class="px-4 py-2 border-t border-slate-100 dark:border-slate-700 text-xs text-slate-500 dark:text-slate-400">
        {{ filtered.length }} of {{ orders.length }} orders
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showForm" class="fixed inset-0 z-50 flex items-start justify-center bg-black/50 backdrop-blur-sm p-4 overflow-y-auto">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-3xl my-8">
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700">
            <h2 class="text-lg font-bold text-slate-900 dark:text-white">{{ isEdit ? 'Edit Purchase Order' : 'New Purchase Order' }}</h2>
            <button @click="showForm = false" class="p-2 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700"><X :size="18" /></button>
          </div>
          <div class="p-6 space-y-5">
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div class="col-span-2">
                <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Supplier <span class="text-red-500">*</span></label>
                <select v-model="form.supplier_id"
                  class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500">
                  <option value="">— Select Supplier —</option>
                  <option v-for="s in suppliers" :key="s.id" :value="s.id">{{ s.name }}</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Order Date <span class="text-red-500">*</span></label>
                <input v-model="form.date" type="date"
                  class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Expected Delivery</label>
                <input v-model="form.expected_date" type="date"
                  class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
              </div>
            </div>

            <!-- Lines -->
            <div>
              <div class="flex items-center justify-between mb-2">
                <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300">Order Lines</h3>
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
                        <input v-model.number="line.quantity" type="number" min="0" step="0.001"
                          class="w-full px-2 py-1 rounded border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-right text-slate-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-primary-400" />
                      </td>
                      <td class="px-3 py-1.5">
                        <input v-model.number="line.unit_price" type="number" min="0" step="0.01"
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
              <div class="flex justify-end mt-3">
                <div class="w-64 space-y-1.5 text-sm">
                  <div class="flex justify-between text-slate-600 dark:text-slate-400">
                    <span>Subtotal HT</span><span class="font-semibold text-slate-900 dark:text-white">{{ fmt(formTotals.subtotal) }}</span>
                  </div>
                  <div class="flex justify-between text-slate-600 dark:text-slate-400">
                    <span>TVA</span><span class="font-semibold text-slate-900 dark:text-white">{{ fmt(formTotals.tva_amount) }}</span>
                  </div>
                  <div class="flex justify-between pt-2 border-t border-slate-200 dark:border-slate-600 font-bold">
                    <span class="text-slate-900 dark:text-white">Total TTC</span>
                    <span class="text-primary-600 dark:text-primary-400">{{ fmt(formTotals.total) }} DZD</span>
                  </div>
                </div>
              </div>
            </div>

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
              {{ saving ? 'Saving...' : 'Save Order' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Detail Modal -->
    <Teleport to="body">
      <div v-if="detailOrder" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700 sticky top-0 bg-white dark:bg-slate-800 z-10">
            <div>
              <h2 class="text-lg font-bold text-slate-900 dark:text-white">{{ detailOrder.number }}</h2>
              <span :class="['inline-flex px-2 py-0.5 rounded-full text-xs font-semibold mt-1', statusCfg(detailOrder.status).cls]">
                {{ statusCfg(detailOrder.status).label }}
              </span>
            </div>
            <button @click="detailOrder = null" class="p-2 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700"><X :size="18" /></button>
          </div>
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-3 text-sm">
              <div><p class="text-xs text-slate-500 dark:text-slate-400">Supplier</p><p class="font-semibold text-slate-900 dark:text-white">{{ detailOrder.supplier_name || '—' }}</p></div>
              <div><p class="text-xs text-slate-500 dark:text-slate-400">Order Date</p><p class="font-semibold text-slate-900 dark:text-white">{{ String(detailOrder.date).slice(0,10) }}</p></div>
              <div><p class="text-xs text-slate-500 dark:text-slate-400">Expected Delivery</p><p class="font-semibold text-slate-900 dark:text-white">{{ detailOrder.expected_date || '—' }}</p></div>
              <div><p class="text-xs text-slate-500 dark:text-slate-400">Currency</p><p class="font-semibold text-slate-900 dark:text-white">{{ detailOrder.currency }}</p></div>
            </div>
            <div v-if="detailOrder.lines && detailOrder.lines.length" class="overflow-x-auto rounded-lg border border-slate-200 dark:border-slate-600">
              <table class="w-full text-xs">
                <thead class="bg-slate-50 dark:bg-slate-700/50">
                  <tr>
                    <th class="px-3 py-2 text-left font-semibold text-slate-500 dark:text-slate-400">#</th>
                    <th class="px-3 py-2 text-left font-semibold text-slate-500 dark:text-slate-400">Description</th>
                    <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400">Qty</th>
                    <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400">Rcvd</th>
                    <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400">Unit Price</th>
                    <th class="px-3 py-2 text-right font-semibold text-slate-500 dark:text-slate-400">Total</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
                  <tr v-for="(l, i) in detailOrder.lines" :key="i">
                    <td class="px-3 py-2 text-slate-500 dark:text-slate-400">{{ i+1 }}</td>
                    <td class="px-3 py-2 text-slate-900 dark:text-white">{{ l.description || l.item_name }}</td>
                    <td class="px-3 py-2 text-right">{{ l.quantity }}</td>
                    <td class="px-3 py-2 text-right">{{ l.received_qty || 0 }}</td>
                    <td class="px-3 py-2 text-right">{{ fmt(l.unit_price) }}</td>
                    <td class="px-3 py-2 text-right font-semibold text-slate-900 dark:text-white">{{ fmt(l.total || 0) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div class="flex justify-end">
              <div class="w-56 space-y-1 text-sm">
                <div class="flex justify-between text-slate-600 dark:text-slate-400"><span>HT</span><span>{{ fmt(detailOrder.subtotal) }}</span></div>
                <div class="flex justify-between text-slate-600 dark:text-slate-400"><span>TVA</span><span>{{ fmt(detailOrder.tva_amount) }}</span></div>
                <div class="flex justify-between pt-2 border-t border-slate-200 dark:border-slate-600 font-bold text-slate-900 dark:text-white">
                  <span>Total TTC</span><span class="text-primary-600 dark:text-primary-400">{{ fmt(detailOrder.total_amount) }} DZD</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
