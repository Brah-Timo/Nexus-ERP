<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Search, Plus, RefreshCw, Edit2, ChevronDown, ChevronUp,
  Cog, Play, CheckCircle, XCircle, Clock, AlertCircle,
  X, Check, Package, Factory, TrendingUp, BarChart2,
  Calendar, Wrench
} from 'lucide-vue-next'
import { manufacturingAPI, inventoryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ────────────────────────────────────────────────────────────────────
interface MOComponentLine {
  id?: string
  component_id: string
  component_code: string
  component_name: string
  required_qty: number
  consumed_qty: number
  unit_cost: number
}

interface ManufacturingOrder {
  id: string
  number: string
  bom_id: string
  bom_code: string
  bom_version: string
  product_id: string
  product_code: string
  product_name: string
  warehouse_id: string | null
  warehouse_name: string
  planned_qty: number
  produced_qty: number
  status: string
  planned_start: string | null
  planned_end: string | null
  actual_start: string | null
  actual_end: string | null
  material_cost: number
  labor_cost: number
  overhead_cost: number
  total_cost: number
  progress_pct: number
  notes: string | null
  component_lines?: MOComponentLine[]
  created_at: string
  updated_at: string
}

// ─── State ────────────────────────────────────────────────────────────────────
const orders = ref<ManufacturingOrder[]>([])
const boms = ref<any[]>([])
const warehouses = ref<any[]>([])

const loading = ref(false)
const search = ref('')
const filterStatus = ref('')
const sortField = ref('created_at')
const sortDir = ref<'asc' | 'desc'>('desc')

// Modal
const showModal = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const saving = ref(false)

// Drawer
const drawerOrder = ref<ManufacturingOrder | null>(null)
const drawerLoading = ref(false)

// Complete modal
const showCompleteModal = ref(false)
const completingOrder = ref<ManufacturingOrder | null>(null)
const completeForm = ref({ produced_qty: 0, material_cost: 0, labor_cost: 0, overhead_cost: 0 })
const completing = ref(false)

// Cancel confirm
const confirmCancel = ref<ManufacturingOrder | null>(null)
const cancelling = ref(false)

const form = ref({
  id: '',
  bom_id: '',
  product_id: '',
  warehouse_id: null as string | null,
  planned_qty: 1,
  planned_start: '',
  planned_end: '',
  notes: ''
})

// ─── Computed ─────────────────────────────────────────────────────────────────
const filtered = computed(() => {
  let list = [...orders.value]
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(o =>
      o.number.toLowerCase().includes(q) ||
      o.product_name.toLowerCase().includes(q) ||
      o.product_code.toLowerCase().includes(q) ||
      o.bom_code.toLowerCase().includes(q)
    )
  }
  if (filterStatus.value) list = list.filter(o => o.status === filterStatus.value)
  list.sort((a, b) => {
    const av = (a as any)[sortField.value] ?? ''
    const bv = (b as any)[sortField.value] ?? ''
    const cmp = String(av).localeCompare(String(bv))
    return sortDir.value === 'asc' ? cmp : -cmp
  })
  return list
})

const statuses = ['draft', 'planned', 'in_progress', 'completed', 'cancelled']

const kpis = computed(() => ({
  total: orders.value.length,
  inProgress: orders.value.filter(o => o.status === 'in_progress').length,
  completed: orders.value.filter(o => o.status === 'completed').length,
  draft: orders.value.filter(o => o.status === 'draft' || o.status === 'planned').length,
  totalCost: orders.value.filter(o => o.status === 'completed').reduce((s, o) => s + o.total_cost, 0)
}))

// ─── Load ─────────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (filterStatus.value) params.status = filterStatus.value
    const [ordersRes, bomsRes, whRes] = await Promise.all([
      manufacturingAPI.getOrders(params),
      manufacturingAPI.getBOMs({ active: 'true' }),
      inventoryAPI.getWarehouses()
    ])
    orders.value = ordersRes.data || []
    boms.value = bomsRes.data || []
    warehouses.value = whRes.data || []
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to load orders', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(load)

function setSort(field: string) {
  if (sortField.value === field) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortField.value = field; sortDir.value = 'asc' }
}

// ─── BOM selection ────────────────────────────────────────────────────────────
function onBOMSelect(bomId: string) {
  const bom = boms.value.find(b => b.id === bomId)
  if (bom) {
    form.value.bom_id = bom.id
    form.value.product_id = bom.product_id
  }
}

// ─── Modal helpers ────────────────────────────────────────────────────────────
function openCreate() {
  modalMode.value = 'create'
  form.value = { id: '', bom_id: '', product_id: '', warehouse_id: null, planned_qty: 1, planned_start: '', planned_end: '', notes: '' }
  showModal.value = true
}

function openEdit(order: ManufacturingOrder) {
  modalMode.value = 'edit'
  form.value = {
    id: order.id,
    bom_id: order.bom_id,
    product_id: order.product_id,
    warehouse_id: order.warehouse_id,
    planned_qty: order.planned_qty,
    planned_start: order.planned_start || '',
    planned_end: order.planned_end || '',
    notes: order.notes || ''
  }
  showModal.value = true
}

async function openDrawer(order: ManufacturingOrder) {
  drawerOrder.value = order
  drawerLoading.value = true
  try {
    const res = await manufacturingAPI.getOrder(order.id)
    drawerOrder.value = res.data
  } catch { /* ignore */ } finally {
    drawerLoading.value = false
  }
}

// ─── Save ─────────────────────────────────────────────────────────────────────
async function save() {
  if (!form.value.bom_id) {
    app.addToast('Please select a BOM', 'error')
    return
  }
  saving.value = true
  try {
    const payload = {
      bom_id: form.value.bom_id,
      product_id: form.value.product_id,
      warehouse_id: form.value.warehouse_id,
      planned_qty: form.value.planned_qty,
      planned_start: form.value.planned_start || null,
      planned_end: form.value.planned_end || null,
      notes: form.value.notes || null
    }
    if (modalMode.value === 'create') {
      await manufacturingAPI.createOrder(payload)
      app.addToast('Manufacturing order created', 'success')
    } else {
      await manufacturingAPI.updateOrder(form.value.id, payload)
      app.addToast('Order updated', 'success')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to save order', 'error')
  } finally {
    saving.value = false
  }
}

// ─── Status transitions ───────────────────────────────────────────────────────
async function startOrder(order: ManufacturingOrder) {
  try {
    await manufacturingAPI.startOrder(order.id)
    app.addToast(`Order ${order.number} started`, 'success')
    await load()
    if (drawerOrder.value?.id === order.id) openDrawer(order)
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to start order', 'error')
  }
}

function openCompleteModal(order: ManufacturingOrder) {
  completingOrder.value = order
  completeForm.value = {
    produced_qty: order.planned_qty,
    material_cost: order.material_cost,
    labor_cost: order.labor_cost,
    overhead_cost: order.overhead_cost
  }
  showCompleteModal.value = true
}

async function completeOrder() {
  if (!completingOrder.value) return
  completing.value = true
  try {
    await manufacturingAPI.completeOrder(completingOrder.value.id, completeForm.value)
    app.addToast(`Order ${completingOrder.value.number} completed`, 'success')
    showCompleteModal.value = false
    completingOrder.value = null
    await load()
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to complete order', 'error')
  } finally {
    completing.value = false
  }
}

async function cancelOrder() {
  if (!confirmCancel.value) return
  cancelling.value = true
  try {
    await manufacturingAPI.cancelOrder(confirmCancel.value.id)
    app.addToast(`Order ${confirmCancel.value.number} cancelled`, 'success')
    confirmCancel.value = null
    await load()
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to cancel order', 'error')
  } finally {
    cancelling.value = false
  }
}

// ─── Helpers ──────────────────────────────────────────────────────────────────
function fmtDate(d: string | null) {
  return d ? new Date(d).toLocaleDateString('en-GB') : '—'
}
function fmtNum(n: number) {
  return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n)
}

const statusConfig: Record<string, { label: string; classes: string }> = {
  draft:       { label: 'Draft',       classes: 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300' },
  planned:     { label: 'Planned',     classes: 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400' },
  in_progress: { label: 'In Progress', classes: 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400' },
  completed:   { label: 'Completed',   classes: 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400' },
  cancelled:   { label: 'Cancelled',   classes: 'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400' }
}
</script>

<template>
  <div class="flex flex-col h-full gap-4 p-4 bg-slate-50 dark:bg-slate-950 min-h-screen">

    <!-- ── Header ──────────────────────────────────────────────────────────── -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
          <Cog class="w-6 h-6 text-amber-600" />
          Manufacturing Orders
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Plan and track production orders</p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="load" :disabled="loading"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300
                 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button @click="openCreate"
          class="flex items-center gap-2 px-4 py-2 bg-amber-600 hover:bg-amber-700 text-white
                 rounded-lg text-sm font-medium transition-colors shadow-sm">
          <Plus class="w-4 h-4" />
          New Order
        </button>
      </div>
    </div>

    <!-- ── KPI Cards ───────────────────────────────────────────────────────── -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total Orders</p>
        <p class="text-2xl font-bold text-slate-900 dark:text-white mt-1">{{ kpis.total }}</p>
        <p class="text-xs text-slate-400 mt-0.5">All time</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">In Progress</p>
        <p class="text-2xl font-bold text-amber-600 mt-1">{{ kpis.inProgress }}</p>
        <p class="text-xs text-slate-400 mt-0.5">Active production</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Completed</p>
        <p class="text-2xl font-bold text-emerald-600 mt-1">{{ kpis.completed }}</p>
        <p class="text-xs text-slate-400 mt-0.5">Finished orders</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Production Cost</p>
        <p class="text-2xl font-bold text-blue-600 mt-1">{{ fmtNum(kpis.totalCost) }}</p>
        <p class="text-xs text-slate-400 mt-0.5">Completed orders</p>
      </div>
    </div>

    <!-- ── Table ───────────────────────────────────────────────────────────── -->
    <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 flex-1 flex flex-col overflow-hidden">
      <!-- Toolbar -->
      <div class="flex flex-wrap items-center gap-3 p-4 border-b border-slate-200 dark:border-slate-700">
        <div class="relative flex-1 min-w-[200px]">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="search" type="text" placeholder="Search number, product..."
            class="w-full pl-9 pr-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                   bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-amber-500 outline-none" />
        </div>
        <select v-model="filterStatus" @change="load"
          class="px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-amber-500 outline-none">
          <option value="">All Status</option>
          <option v-for="s in statuses" :key="s" :value="s">
            {{ statusConfig[s]?.label ?? s }}
          </option>
        </select>
      </div>

      <!-- Table body -->
      <div class="flex-1 overflow-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/50">
              <th v-for="col in [
                { key: 'number', label: 'MO Number' },
                { key: 'product_name', label: 'Product' },
                { key: 'bom_code', label: 'BOM' },
                { key: 'planned_qty', label: 'Planned' },
                { key: 'produced_qty', label: 'Produced' },
                { key: 'status', label: 'Status' },
                { key: 'planned_start', label: 'Start Date' },
                { key: 'planned_end', label: 'End Date' },
                { key: '', label: 'Progress' },
                { key: '', label: '' }
              ]" :key="col.label"
                class="px-4 py-3 text-left font-medium text-slate-500 dark:text-slate-400 whitespace-nowrap"
                :class="col.key ? 'cursor-pointer hover:text-slate-700 dark:hover:text-slate-200 select-none' : ''"
                @click="col.key ? setSort(col.key) : null">
                <span class="flex items-center gap-1">
                  {{ col.label }}
                  <template v-if="col.key && sortField === col.key">
                    <ChevronUp v-if="sortDir === 'asc'" class="w-3 h-3" />
                    <ChevronDown v-else class="w-3 h-3" />
                  </template>
                </span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="10" class="px-4 py-12 text-center text-slate-400">
                <RefreshCw class="w-5 h-5 animate-spin mx-auto mb-2" /> Loading...
              </td>
            </tr>
            <tr v-else-if="filtered.length === 0">
              <td colspan="10" class="px-4 py-12 text-center text-slate-400">
                <Cog class="w-8 h-8 mx-auto mb-2 opacity-30" />
                No manufacturing orders found
              </td>
            </tr>
            <tr v-for="order in filtered" :key="order.id"
              class="border-b border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50
                     cursor-pointer transition-colors"
              @click="openDrawer(order)">
              <td class="px-4 py-3">
                <span class="font-mono font-semibold text-amber-700 dark:text-amber-400">{{ order.number }}</span>
              </td>
              <td class="px-4 py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-white">{{ order.product_name }}</p>
                  <p class="text-xs text-slate-400">{{ order.product_code }}</p>
                </div>
              </td>
              <td class="px-4 py-3">
                <span class="font-mono text-xs text-slate-600 dark:text-slate-300">
                  {{ order.bom_code }} v{{ order.bom_version }}
                </span>
              </td>
              <td class="px-4 py-3 font-mono text-slate-700 dark:text-slate-300">
                {{ fmtNum(order.planned_qty) }}
              </td>
              <td class="px-4 py-3 font-mono text-slate-700 dark:text-slate-300">
                {{ fmtNum(order.produced_qty) }}
              </td>
              <td class="px-4 py-3">
                <span :class="statusConfig[order.status]?.classes"
                  class="px-2 py-0.5 rounded-full text-xs font-medium">
                  {{ statusConfig[order.status]?.label ?? order.status }}
                </span>
              </td>
              <td class="px-4 py-3 text-xs text-slate-500 dark:text-slate-400">
                {{ fmtDate(order.planned_start) }}
              </td>
              <td class="px-4 py-3 text-xs text-slate-500 dark:text-slate-400">
                {{ fmtDate(order.planned_end) }}
              </td>
              <td class="px-4 py-3 min-w-[100px]">
                <div class="flex items-center gap-2">
                  <div class="flex-1 bg-slate-100 dark:bg-slate-700 rounded-full h-1.5 overflow-hidden">
                    <div class="h-full rounded-full transition-all duration-300"
                      :class="order.status === 'completed' ? 'bg-emerald-500' : 'bg-amber-500'"
                      :style="{ width: `${Math.min(order.progress_pct, 100)}%` }" />
                  </div>
                  <span class="text-xs text-slate-500 dark:text-slate-400 font-mono w-8 shrink-0">
                    {{ Math.round(order.progress_pct) }}%
                  </span>
                </div>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1" @click.stop>
                  <!-- Start -->
                  <button v-if="['draft','planned'].includes(order.status)"
                    @click="startOrder(order)"
                    title="Start Order"
                    class="p-1.5 rounded-lg hover:bg-amber-50 dark:hover:bg-amber-900/20 text-slate-400 hover:text-amber-600 transition-colors">
                    <Play class="w-3.5 h-3.5" />
                  </button>
                  <!-- Complete -->
                  <button v-if="order.status === 'in_progress'"
                    @click="openCompleteModal(order)"
                    title="Complete Order"
                    class="p-1.5 rounded-lg hover:bg-emerald-50 dark:hover:bg-emerald-900/20 text-slate-400 hover:text-emerald-600 transition-colors">
                    <CheckCircle class="w-3.5 h-3.5" />
                  </button>
                  <!-- Edit -->
                  <button v-if="['draft','planned'].includes(order.status)"
                    @click="openEdit(order)"
                    title="Edit"
                    class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-400 hover:text-slate-700 dark:hover:text-white transition-colors">
                    <Edit2 class="w-3.5 h-3.5" />
                  </button>
                  <!-- Cancel -->
                  <button v-if="['draft','planned'].includes(order.status)"
                    @click="confirmCancel = order"
                    title="Cancel"
                    class="p-1.5 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 text-slate-400 hover:text-red-600 transition-colors">
                    <XCircle class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="px-4 py-2 border-t border-slate-200 dark:border-slate-700 text-xs text-slate-500 dark:text-slate-400">
        {{ filtered.length }} of {{ orders.length }} orders
      </div>
    </div>

    <!-- ── Order Detail Drawer ──────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="drawerOrder" class="fixed inset-0 z-40 flex justify-end">
        <div class="absolute inset-0 bg-black/30 dark:bg-black/50" @click="drawerOrder = null" />
        <div class="relative z-10 w-full max-w-xl bg-white dark:bg-slate-900 shadow-2xl flex flex-col overflow-hidden">
          <div class="flex items-center justify-between p-5 border-b border-slate-200 dark:border-slate-700">
            <div>
              <h2 class="text-lg font-bold text-slate-900 dark:text-white font-mono">{{ drawerOrder.number }}</h2>
              <p class="text-sm text-slate-500 mt-0.5">{{ drawerOrder.product_name }}</p>
            </div>
            <div class="flex items-center gap-2">
              <span :class="statusConfig[drawerOrder.status]?.classes"
                class="px-2.5 py-1 rounded-full text-xs font-medium">
                {{ statusConfig[drawerOrder.status]?.label }}
              </span>
              <button @click="drawerOrder = null" class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500">
                <X class="w-4 h-4" />
              </button>
            </div>
          </div>

          <div class="flex-1 overflow-y-auto p-5 space-y-5">
            <div v-if="drawerLoading" class="flex justify-center py-8">
              <RefreshCw class="w-5 h-5 animate-spin text-amber-600" />
            </div>
            <template v-else>
              <!-- Progress bar -->
              <div class="p-4 bg-slate-50 dark:bg-slate-800 rounded-xl">
                <div class="flex justify-between items-center mb-2">
                  <span class="text-sm font-medium text-slate-700 dark:text-slate-300">Production Progress</span>
                  <span class="text-sm font-bold text-amber-600">{{ Math.round(drawerOrder.progress_pct) }}%</span>
                </div>
                <div class="bg-slate-200 dark:bg-slate-700 rounded-full h-3 overflow-hidden">
                  <div class="h-full rounded-full transition-all duration-500"
                    :class="drawerOrder.status === 'completed' ? 'bg-emerald-500' : 'bg-amber-500'"
                    :style="{ width: `${Math.min(drawerOrder.progress_pct, 100)}%` }" />
                </div>
                <div class="flex justify-between text-xs text-slate-500 mt-1">
                  <span>Produced: {{ fmtNum(drawerOrder.produced_qty) }}</span>
                  <span>Planned: {{ fmtNum(drawerOrder.planned_qty) }}</span>
                </div>
              </div>

              <!-- Info grid -->
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">BOM</p>
                  <p class="text-sm font-mono text-slate-700 dark:text-slate-300 mt-0.5">{{ drawerOrder.bom_code }} v{{ drawerOrder.bom_version }}</p>
                </div>
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Warehouse</p>
                  <p class="text-sm text-slate-700 dark:text-slate-300 mt-0.5">{{ drawerOrder.warehouse_name || '—' }}</p>
                </div>
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Planned Start</p>
                  <p class="text-sm text-slate-700 dark:text-slate-300 mt-0.5">{{ fmtDate(drawerOrder.planned_start) }}</p>
                </div>
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Planned End</p>
                  <p class="text-sm text-slate-700 dark:text-slate-300 mt-0.5">{{ fmtDate(drawerOrder.planned_end) }}</p>
                </div>
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Actual Start</p>
                  <p class="text-sm text-slate-700 dark:text-slate-300 mt-0.5">{{ fmtDate(drawerOrder.actual_start) }}</p>
                </div>
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Actual End</p>
                  <p class="text-sm text-slate-700 dark:text-slate-300 mt-0.5">{{ fmtDate(drawerOrder.actual_end) }}</p>
                </div>
              </div>

              <!-- Costs -->
              <div class="border border-slate-200 dark:border-slate-700 rounded-xl overflow-hidden">
                <div class="bg-slate-50 dark:bg-slate-800 px-4 py-2.5 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">
                  Cost Breakdown
                </div>
                <div class="divide-y divide-slate-100 dark:divide-slate-800">
                  <div class="flex justify-between px-4 py-2.5 text-sm">
                    <span class="text-slate-600 dark:text-slate-400">Material Cost</span>
                    <span class="font-mono text-slate-900 dark:text-white">{{ fmtNum(drawerOrder.material_cost) }}</span>
                  </div>
                  <div class="flex justify-between px-4 py-2.5 text-sm">
                    <span class="text-slate-600 dark:text-slate-400">Labor Cost</span>
                    <span class="font-mono text-slate-900 dark:text-white">{{ fmtNum(drawerOrder.labor_cost) }}</span>
                  </div>
                  <div class="flex justify-between px-4 py-2.5 text-sm">
                    <span class="text-slate-600 dark:text-slate-400">Overhead Cost</span>
                    <span class="font-mono text-slate-900 dark:text-white">{{ fmtNum(drawerOrder.overhead_cost) }}</span>
                  </div>
                  <div class="flex justify-between px-4 py-2.5 text-sm font-semibold bg-slate-50 dark:bg-slate-800">
                    <span class="text-slate-700 dark:text-slate-300">Total Cost</span>
                    <span class="font-mono text-amber-700 dark:text-amber-400">{{ fmtNum(drawerOrder.total_cost) }}</span>
                  </div>
                </div>
              </div>

              <!-- Components -->
              <div>
                <h3 class="text-sm font-semibold text-slate-900 dark:text-white flex items-center gap-2 mb-3">
                  <Package class="w-4 h-4 text-blue-600" />
                  Components ({{ drawerOrder.component_lines?.length ?? 0 }})
                </h3>
                <div v-if="drawerOrder.component_lines && drawerOrder.component_lines.length > 0"
                  class="rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
                  <table class="w-full text-xs">
                    <thead class="bg-slate-50 dark:bg-slate-800">
                      <tr>
                        <th class="px-3 py-2 text-left text-slate-500 dark:text-slate-400 font-medium">Component</th>
                        <th class="px-3 py-2 text-right text-slate-500 dark:text-slate-400 font-medium">Required</th>
                        <th class="px-3 py-2 text-right text-slate-500 dark:text-slate-400 font-medium">Consumed</th>
                        <th class="px-3 py-2 text-right text-slate-500 dark:text-slate-400 font-medium">Unit Cost</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="line in drawerOrder.component_lines" :key="line.id"
                        class="border-t border-slate-100 dark:border-slate-800">
                        <td class="px-3 py-2">
                          <p class="font-medium text-slate-900 dark:text-white">{{ line.component_name }}</p>
                          <p class="text-slate-400">{{ line.component_code }}</p>
                        </td>
                        <td class="px-3 py-2 text-right font-mono text-slate-700 dark:text-slate-300">{{ fmtNum(line.required_qty) }}</td>
                        <td class="px-3 py-2 text-right font-mono"
                          :class="line.consumed_qty >= line.required_qty
                            ? 'text-emerald-600 dark:text-emerald-400'
                            : 'text-slate-700 dark:text-slate-300'">
                          {{ fmtNum(line.consumed_qty) }}
                        </td>
                        <td class="px-3 py-2 text-right font-mono text-slate-500 dark:text-slate-400">{{ fmtNum(line.unit_cost) }}</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <p v-else class="text-xs text-slate-400 italic">No component lines</p>
              </div>

              <!-- Notes -->
              <div v-if="drawerOrder.notes">
                <p class="text-xs text-slate-400 uppercase tracking-wide mb-1">Notes</p>
                <p class="text-sm text-slate-600 dark:text-slate-300">{{ drawerOrder.notes }}</p>
              </div>

              <!-- Quick actions -->
              <div v-if="['draft','planned','in_progress'].includes(drawerOrder.status)"
                class="flex gap-2 pt-2 border-t border-slate-100 dark:border-slate-800">
                <button v-if="['draft','planned'].includes(drawerOrder.status)"
                  @click="startOrder(drawerOrder)"
                  class="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-amber-600 hover:bg-amber-700 text-white rounded-lg transition-colors">
                  <Play class="w-3.5 h-3.5" /> Start
                </button>
                <button v-if="drawerOrder.status === 'in_progress'"
                  @click="openCompleteModal(drawerOrder)"
                  class="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg transition-colors">
                  <CheckCircle class="w-3.5 h-3.5" /> Complete
                </button>
              </div>
            </template>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Create / Edit Modal ─────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40" @click="showModal = false" />
        <div class="relative z-10 w-full max-w-lg bg-white dark:bg-slate-900 rounded-2xl shadow-2xl">
          <div class="flex items-center justify-between p-6 border-b border-slate-200 dark:border-slate-700">
            <h2 class="text-lg font-bold text-slate-900 dark:text-white">
              {{ modalMode === 'create' ? 'New Manufacturing Order' : 'Edit Order' }}
            </h2>
            <button @click="showModal = false" class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500">
              <X class="w-5 h-5" />
            </button>
          </div>

          <div class="p-6 space-y-4">
            <div>
              <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                Bill of Materials <span class="text-red-500">*</span>
              </label>
              <select :value="form.bom_id"
                @change="onBOMSelect(($event.target as HTMLSelectElement).value)"
                class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                       bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-amber-500 outline-none">
                <option value="">Select BOM...</option>
                <option v-for="bom in boms" :key="bom.id" :value="bom.id">
                  {{ bom.code }} v{{ bom.version }} — {{ bom.product_name }}
                </option>
              </select>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Planned Quantity</label>
                <input v-model.number="form.planned_qty" type="number" min="0.0001" step="0.0001"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-amber-500 outline-none" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Warehouse</label>
                <select v-model="form.warehouse_id"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-amber-500 outline-none">
                  <option :value="null">None</option>
                  <option v-for="wh in warehouses" :key="wh.id" :value="wh.id">{{ wh.name }}</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Planned Start</label>
                <input v-model="form.planned_start" type="date"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-amber-500 outline-none" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Planned End</label>
                <input v-model="form.planned_end" type="date"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-amber-500 outline-none" />
              </div>
            </div>

            <div>
              <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Notes</label>
              <textarea v-model="form.notes" rows="2"
                class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                       bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-amber-500 outline-none resize-none" />
            </div>
          </div>

          <div class="flex justify-end gap-3 p-6 border-t border-slate-200 dark:border-slate-700">
            <button @click="showModal = false"
              class="px-4 py-2 text-sm text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
              Cancel
            </button>
            <button @click="save" :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 text-sm font-medium bg-amber-600 hover:bg-amber-700
                     disabled:opacity-50 text-white rounded-lg transition-colors">
              <RefreshCw v-if="saving" class="w-3.5 h-3.5 animate-spin" />
              <Check v-else class="w-3.5 h-3.5" />
              {{ modalMode === 'create' ? 'Create Order' : 'Save Changes' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Complete Modal ──────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showCompleteModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40" @click="showCompleteModal = false" />
        <div class="relative z-10 w-full max-w-md bg-white dark:bg-slate-900 rounded-2xl shadow-2xl">
          <div class="flex items-center justify-between p-5 border-b border-slate-200 dark:border-slate-700">
            <h2 class="text-base font-bold text-slate-900 dark:text-white">Complete Order</h2>
            <button @click="showCompleteModal = false" class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500">
              <X class="w-4 h-4" />
            </button>
          </div>
          <div class="p-5 space-y-4">
            <div>
              <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Produced Quantity</label>
              <input v-model.number="completeForm.produced_qty" type="number" min="0" step="0.0001"
                class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                       bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none" />
            </div>
            <div class="grid grid-cols-3 gap-3">
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Material Cost</label>
                <input v-model.number="completeForm.material_cost" type="number" min="0" step="0.01"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Labor Cost</label>
                <input v-model.number="completeForm.labor_cost" type="number" min="0" step="0.01"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Overhead</label>
                <input v-model.number="completeForm.overhead_cost" type="number" min="0" step="0.01"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none" />
              </div>
            </div>
            <div class="p-3 bg-emerald-50 dark:bg-emerald-900/20 rounded-lg flex justify-between items-center text-sm">
              <span class="text-emerald-700 dark:text-emerald-400 font-medium">Total Cost</span>
              <span class="font-mono font-bold text-emerald-700 dark:text-emerald-400">
                {{ fmtNum(completeForm.material_cost + completeForm.labor_cost + completeForm.overhead_cost) }}
              </span>
            </div>
          </div>
          <div class="flex justify-end gap-3 p-5 border-t border-slate-200 dark:border-slate-700">
            <button @click="showCompleteModal = false"
              class="px-4 py-2 text-sm text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
              Cancel
            </button>
            <button @click="completeOrder" :disabled="completing"
              class="flex items-center gap-2 px-4 py-2 text-sm font-medium bg-emerald-600 hover:bg-emerald-700
                     disabled:opacity-50 text-white rounded-lg transition-colors">
              <RefreshCw v-if="completing" class="w-3.5 h-3.5 animate-spin" />
              <CheckCircle v-else class="w-3.5 h-3.5" />
              Mark Complete
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Cancel Confirm ──────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="confirmCancel" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40" @click="confirmCancel = null" />
        <div class="relative z-10 w-full max-w-sm bg-white dark:bg-slate-900 rounded-xl shadow-2xl p-6">
          <div class="flex items-start gap-4 mb-4">
            <div class="p-2 bg-red-100 dark:bg-red-900/30 rounded-lg">
              <AlertCircle class="w-5 h-5 text-red-600" />
            </div>
            <div>
              <h3 class="font-semibold text-slate-900 dark:text-white">Cancel Order?</h3>
              <p class="text-sm text-slate-500 mt-1">
                Order <strong>{{ confirmCancel.number }}</strong> will be cancelled.
              </p>
            </div>
          </div>
          <div class="flex gap-3 justify-end">
            <button @click="confirmCancel = null"
              class="px-4 py-2 text-sm text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
              Keep
            </button>
            <button @click="cancelOrder" :disabled="cancelling"
              class="px-4 py-2 text-sm font-medium bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white rounded-lg transition-colors">
              Cancel Order
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
