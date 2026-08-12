<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Wrench, Plus, Search, Eye, Edit2, Trash2, CheckCircle, X,
  ChevronDown, AlertTriangle, Clock, Package, DollarSign,
  Filter, RefreshCw, ChevronLeft, ChevronRight, User, Calendar,
  FileText, BarChart2, Layers
} from 'lucide-vue-next'
import { maintenanceAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── types ────────────────────────────────────────────────────────────────────
interface OrderLine {
  id: string
  part_name: string
  part_number?: string
  quantity: number
  unit_cost: number
  total_cost: number
}

interface Order {
  id: string
  order_number: string
  title: string
  description?: string
  order_type: string
  status: string
  priority: string
  equipment_id?: string
  equipment_name?: string
  equipment_code?: string
  request_id?: string
  request_number?: string
  assigned_to?: string
  technician_name?: string
  scheduled_date?: string
  actual_start?: string
  actual_end?: string
  estimated_hours?: number
  actual_hours?: number
  estimated_cost?: number
  actual_cost?: number
  labor_cost?: number
  parts_cost?: number
  other_cost?: number
  work_performed?: string
  findings?: string
  next_service_date?: string
  color?: string
  lines?: OrderLine[]
  created_at: string
}

interface KPI {
  label: string
  value: number
  sub?: string
  icon: unknown
  color: string
  bg: string
}

// ─── state ────────────────────────────────────────────────────────────────────
const orders       = ref<Order[]>([])
const equipment    = ref<{ id: string; name: string; code: string }[]>([])
const loading      = ref(false)
const search       = ref('')
const statusFilter = ref('all')
const typeFilter   = ref('all')

const page         = ref(1)
const perPage      = 20

// modal flags
const showCreate   = ref(false)
const showEdit     = ref(false)
const showView     = ref(false)
const showComplete = ref(false)
const showDelete   = ref(false)
const deleting     = ref(false)
const saving       = ref(false)
const completing   = ref(false)

const selected     = ref<Order | null>(null)

// form data
const form = ref({
  title: '', description: '', order_type: 'corrective', priority: 'medium',
  equipment_id: '', request_id: '',
  assigned_to: '', technician_name: '',
  scheduled_date: '', estimated_hours: '', estimated_cost: '',
  status: 'pending'
})

const completeForm = ref({
  work_performed: '', findings: '',
  actual_hours: '', labor_cost: '', parts_cost: '', other_cost: '',
  technician_name: '', next_service_date: ''
})

// ─── computed ────────────────────────────────────────────────────────────────
const dk = (a: string, b: string) => app.darkMode ? a : b

const filtered = computed(() => {
  let list = orders.value
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(o =>
      o.order_number.toLowerCase().includes(q) ||
      o.title.toLowerCase().includes(q) ||
      (o.equipment_name || '').toLowerCase().includes(q) ||
      (o.technician_name || '').toLowerCase().includes(q)
    )
  }
  if (statusFilter.value !== 'all') list = list.filter(o => o.status === statusFilter.value)
  if (typeFilter.value   !== 'all') list = list.filter(o => o.order_type === typeFilter.value)
  return list
})

const paginated = computed(() => {
  const start = (page.value - 1) * perPage
  return filtered.value.slice(start, start + perPage)
})
const totalPages = computed(() => Math.max(1, Math.ceil(filtered.value.length / perPage)))

const kpis = computed<KPI[]>(() => {
  const all       = orders.value
  const inProg    = all.filter(o => o.status === 'in_progress').length
  const done      = all.filter(o => o.status === 'completed').length
  const overdue   = all.filter(o => isOverdue(o)).length
  const pending   = all.filter(o => o.status === 'pending').length
  const totalCost = all.reduce((s, o) => s + (o.actual_cost || o.estimated_cost || 0), 0)
  return [
    { label: 'Total Orders',    value: all.length,  sub: 'all time',          icon: Wrench,       color: 'text-violet-500', bg: 'bg-violet-500/10' },
    { label: 'In Progress',     value: inProg,       sub: 'active now',        icon: Clock,        color: 'text-blue-500',   bg: 'bg-blue-500/10' },
    { label: 'Completed',       value: done,         sub: 'finished',          icon: CheckCircle,  color: 'text-emerald-500',bg: 'bg-emerald-500/10' },
    { label: 'Overdue',         value: overdue,      sub: 'need attention',    icon: AlertTriangle,color: 'text-rose-500',   bg: 'bg-rose-500/10' },
    { label: 'Pending',         value: pending,      sub: 'not started',       icon: Filter,       color: 'text-amber-500',  bg: 'bg-amber-500/10' },
    { label: 'Total Cost',      value: totalCost,    sub: 'all orders',        icon: DollarSign,   color: 'text-teal-500',   bg: 'bg-teal-500/10' },
  ]
})

const typeCounts = computed(() => {
  const types = ['corrective','preventive','inspection','emergency','upgrade']
  return types.map(t => ({ type: t, count: orders.value.filter(o => o.order_type === t).length }))
})

// ─── helpers ─────────────────────────────────────────────────────────────────
const fmt = (n: number) =>
  new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'

const fmtDate = (s?: string) => {
  if (!s) return '—'
  return new Date(s).toLocaleDateString('fr-DZ', { day:'2-digit', month:'short', year:'numeric' })
}

const isOverdue = (o: Order) => {
  if (['completed','cancelled'].includes(o.status)) return false
  if (!o.scheduled_date) return false
  return new Date(o.scheduled_date) < new Date()
}

const typeLabel = (t: string) => ({
  corrective:'Corrective', preventive:'Preventive',
  inspection:'Inspection', emergency:'Emergency', upgrade:'Upgrade'
}[t] ?? t)

const typeBadge = (t: string) => ({
  corrective: 'bg-blue-500/15 text-blue-400',
  preventive: 'bg-emerald-500/15 text-emerald-400',
  inspection: 'bg-violet-500/15 text-violet-400',
  emergency:  'bg-rose-500/15 text-rose-400',
  upgrade:    'bg-amber-500/15 text-amber-400',
}[t] ?? 'bg-slate-500/15 text-slate-400')

const typeColor = (t: string) => ({
  corrective:'#3b82f6', preventive:'#10b981',
  inspection:'#8b5cf6', emergency:'#f43f5e', upgrade:'#f59e0b'
}[t] ?? '#64748b')

const statusBadge = (s: string) => ({
  pending:     'bg-slate-500/15 text-slate-400',
  in_progress: 'bg-blue-500/15 text-blue-400',
  on_hold:     'bg-amber-500/15 text-amber-400',
  completed:   'bg-emerald-500/15 text-emerald-400',
  cancelled:   'bg-rose-500/15 text-rose-400',
}[s] ?? 'bg-slate-500/15 text-slate-400')

const statusLabel = (s: string) => ({
  pending:'Pending', in_progress:'In Progress',
  on_hold:'On Hold', completed:'Completed', cancelled:'Cancelled'
}[s] ?? s)

const priorityBadge = (p: string) => ({
  low:'bg-slate-500/15 text-slate-400',
  medium:'bg-amber-500/15 text-amber-400',
  high:'bg-orange-500/15 text-orange-400',
  critical:'bg-rose-500/15 text-rose-400',
}[p] ?? 'bg-slate-500/15 text-slate-400')

// ─── data loading ────────────────────────────────────────────────────────────
const load = async () => {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (statusFilter.value !== 'all') params.status = statusFilter.value
    if (typeFilter.value   !== 'all') params.type   = typeFilter.value
    const res = await maintenanceAPI.listOrders(params)
    orders.value = res.data.orders ?? res.data ?? []
  } catch {
    app.addToast('Failed to load maintenance orders', 'error')
  } finally {
    loading.value = false
  }
}

const loadEquipment = async () => {
  try {
    const res = await maintenanceAPI.listEquipment({ limit: '500' })
    equipment.value = (res.data.equipment ?? res.data ?? []).map((e: any) => ({
      id: e.id, name: e.name, code: e.code
    }))
  } catch { /* silent */ }
}

// ─── CRUD ─────────────────────────────────────────────────────────────────────
const openCreate = () => {
  form.value = {
    title:'', description:'', order_type:'corrective', priority:'medium',
    equipment_id:'', request_id:'', assigned_to:'', technician_name:'',
    scheduled_date:'', estimated_hours:'', estimated_cost:'', status:'pending'
  }
  showCreate.value = true
}

const openEdit = (o: Order) => {
  selected.value = o
  form.value = {
    title: o.title, description: o.description ?? '',
    order_type: o.order_type, priority: o.priority,
    equipment_id: o.equipment_id ?? '', request_id: o.request_id ?? '',
    assigned_to: o.assigned_to ?? '', technician_name: o.technician_name ?? '',
    scheduled_date: o.scheduled_date ? o.scheduled_date.substring(0,10) : '',
    estimated_hours: o.estimated_hours?.toString() ?? '',
    estimated_cost: o.estimated_cost?.toString() ?? '',
    status: o.status
  }
  showEdit.value = true
}

const openView = async (o: Order) => {
  selected.value = o
  showView.value = true
  try {
    const res = await maintenanceAPI.getOrder(o.id)
    selected.value = res.data
  } catch { /* keep existing */ }
}

const openComplete = (o: Order) => {
  selected.value = o
  completeForm.value = {
    work_performed:'', findings:'',
    actual_hours: o.estimated_hours?.toString() ?? '',
    labor_cost:'', parts_cost:'', other_cost:'',
    technician_name: o.technician_name ?? '',
    next_service_date:''
  }
  showComplete.value = true
}

const openDelete = (o: Order) => {
  selected.value = o
  showDelete.value = true
}

const saveOrder = async (isEdit: boolean) => {
  saving.value = true
  try {
    const payload = {
      ...form.value,
      estimated_hours: form.value.estimated_hours ? parseFloat(form.value.estimated_hours) : null,
      estimated_cost:  form.value.estimated_cost  ? parseFloat(form.value.estimated_cost)  : null,
    }
    if (isEdit && selected.value) {
      await maintenanceAPI.updateOrder(selected.value.id, payload)
      app.addToast('Order updated successfully', 'success')
    } else {
      await maintenanceAPI.createOrder(payload)
      app.addToast('Order created successfully', 'success')
    }
    showCreate.value = false
    showEdit.value   = false
    await load()
  } catch {
    app.addToast('Failed to save order', 'error')
  } finally {
    saving.value = false
  }
}

const completeOrder = async () => {
  if (!selected.value) return
  completing.value = true
  try {
    const payload = {
      ...completeForm.value,
      actual_hours:  completeForm.value.actual_hours  ? parseFloat(completeForm.value.actual_hours)  : null,
      labor_cost:    completeForm.value.labor_cost    ? parseFloat(completeForm.value.labor_cost)    : null,
      parts_cost:    completeForm.value.parts_cost    ? parseFloat(completeForm.value.parts_cost)    : null,
      other_cost:    completeForm.value.other_cost    ? parseFloat(completeForm.value.other_cost)    : null,
    }
    await maintenanceAPI.completeOrder(selected.value.id, payload)
    app.addToast('Order completed successfully', 'success')
    showComplete.value = false
    await load()
  } catch {
    app.addToast('Failed to complete order', 'error')
  } finally {
    completing.value = false
  }
}

const deleteOrder = async () => {
  if (!selected.value) return
  deleting.value = true
  try {
    await maintenanceAPI.deleteOrder(selected.value.id)
    app.addToast('Order deleted', 'success')
    showDelete.value = false
    await load()
  } catch {
    app.addToast('Failed to delete order', 'error')
  } finally {
    deleting.value = false
  }
}

onMounted(() => { load(); loadEquipment() })
</script>

<template>
  <div :class="['min-h-screen p-6 space-y-6', dk('bg-slate-950 text-slate-100','bg-slate-50 text-slate-900')]">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-violet-500/15 flex items-center justify-center">
          <Wrench class="w-5 h-5 text-violet-400" />
        </div>
        <div>
          <h1 class="text-xl font-bold">Maintenance Orders</h1>
          <p :class="['text-sm', dk('text-slate-400','text-slate-500')]">Manage and track work orders</p>
        </div>
      </div>
      <button @click="openCreate"
        class="flex items-center gap-2 px-4 py-2 rounded-xl bg-violet-600 hover:bg-violet-500 text-white text-sm font-medium transition-colors">
        <Plus class="w-4 h-4" /> New Order
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
      <div v-for="k in kpis" :key="k.label"
        :class="['rounded-xl p-4 border', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
        <div class="flex items-center justify-between mb-2">
          <span :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]">{{ k.label }}</span>
          <div :class="['w-8 h-8 rounded-lg flex items-center justify-center', k.bg]">
            <component :is="k.icon" :class="['w-4 h-4', k.color]" />
          </div>
        </div>
        <div class="text-xl font-bold">
          {{ k.label === 'Total Cost' ? fmt(k.value) : k.value }}
        </div>
        <div :class="['text-xs mt-0.5', dk('text-slate-500','text-slate-400')]">{{ k.sub }}</div>
      </div>
    </div>

    <!-- Type Filter Pills -->
    <div class="flex flex-wrap gap-2">
      <button @click="typeFilter='all'; page=1"
        :class="['px-3 py-1.5 rounded-lg text-sm font-medium transition-colors border',
          typeFilter==='all'
            ? 'bg-violet-600 text-white border-violet-600'
            : dk('bg-slate-900 border-slate-700 text-slate-400 hover:border-slate-500','bg-white border-slate-200 text-slate-500 hover:border-slate-400')]">
        All ({{ orders.length }})
      </button>
      <button v-for="t in typeCounts" :key="t.type"
        @click="typeFilter=t.type; page=1"
        :class="['px-3 py-1.5 rounded-lg text-sm font-medium transition-colors border',
          typeFilter===t.type
            ? 'bg-violet-600 text-white border-violet-600'
            : dk('bg-slate-900 border-slate-700 text-slate-400 hover:border-slate-500','bg-white border-slate-200 text-slate-500 hover:border-slate-400')]">
        {{ typeLabel(t.type) }} ({{ t.count }})
      </button>
    </div>

    <!-- Search + Status -->
    <div :class="['rounded-xl border p-4', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
      <div class="flex flex-col sm:flex-row gap-3">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="search" placeholder="Search orders, equipment, technician…"
            :class="['w-full pl-9 pr-4 py-2 rounded-lg border text-sm',
              dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500 focus:border-violet-500',
                 'bg-slate-50 border-slate-200 text-slate-900 placeholder-slate-400 focus:border-violet-500')]"
            @input="page=1" />
        </div>
        <select v-model="statusFilter" @change="page=1; load()"
          :class="['px-3 py-2 rounded-lg border text-sm',
            dk('bg-slate-800 border-slate-700 text-slate-100','bg-slate-50 border-slate-200 text-slate-900')]">
          <option value="all">All Statuses</option>
          <option value="pending">Pending</option>
          <option value="in_progress">In Progress</option>
          <option value="on_hold">On Hold</option>
          <option value="completed">Completed</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <button @click="load" :class="['px-3 py-2 rounded-lg border text-sm flex items-center gap-2',
          dk('bg-slate-800 border-slate-700 text-slate-300 hover:border-slate-600','bg-slate-50 border-slate-200 text-slate-600 hover:bg-slate-100')]">
          <RefreshCw :class="['w-4 h-4', loading && 'animate-spin']" />
        </button>
      </div>
    </div>

    <!-- Table -->
    <div :class="['rounded-xl border overflow-hidden', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
      <div v-if="loading" class="p-12 text-center">
        <RefreshCw class="w-8 h-8 animate-spin text-violet-400 mx-auto mb-3" />
        <p :class="['text-sm', dk('text-slate-400','text-slate-500')]">Loading orders…</p>
      </div>
      <div v-else-if="!filtered.length" class="p-12 text-center">
        <Wrench class="w-10 h-10 mx-auto mb-3 text-slate-500" />
        <p :class="['font-medium', dk('text-slate-300','text-slate-700')]">No orders found</p>
        <p :class="['text-sm mt-1', dk('text-slate-500','text-slate-400')]">Try adjusting your filters or create a new order</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="['border-b', dk('border-slate-800 bg-slate-800/50','border-slate-100 bg-slate-50')]">
              <th class="text-left px-4 py-3 font-medium text-slate-400 w-8"></th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Order</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Equipment</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Type</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Priority</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Status</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Scheduled</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Technician</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Cost</th>
              <th class="text-right px-4 py-3 font-medium text-slate-400">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y" :class="dk('divide-slate-800','divide-slate-100')">
            <tr v-for="o in paginated" :key="o.id"
              :class="['transition-colors group',
                isOverdue(o)
                  ? dk('bg-rose-950/20 hover:bg-rose-950/40','bg-rose-50 hover:bg-rose-100')
                  : dk('hover:bg-slate-800/50','hover:bg-slate-50')]">
              <!-- color bar -->
              <td class="px-4 py-3 w-8">
                <div class="w-1 h-8 rounded-full mx-auto"
                  :style="{ backgroundColor: o.color || typeColor(o.order_type) }"></div>
              </td>
              <td class="px-4 py-3">
                <div class="font-mono text-xs text-violet-400 mb-0.5">{{ o.order_number }}</div>
                <div class="font-medium text-sm line-clamp-1">{{ o.title }}</div>
                <div v-if="isOverdue(o)" class="flex items-center gap-1 text-rose-400 text-xs mt-0.5">
                  <AlertTriangle class="w-3 h-3" /> Overdue
                </div>
              </td>
              <td class="px-4 py-3">
                <div v-if="o.equipment_name" class="text-sm">{{ o.equipment_name }}</div>
                <div v-if="o.equipment_code" :class="['text-xs', dk('text-slate-500','text-slate-400')]">{{ o.equipment_code }}</div>
                <div v-if="!o.equipment_name" :class="['text-xs', dk('text-slate-600','text-slate-400')]">—</div>
              </td>
              <td class="px-4 py-3">
                <span :class="['px-2 py-0.5 rounded-md text-xs font-medium', typeBadge(o.order_type)]">
                  {{ typeLabel(o.order_type) }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span :class="['px-2 py-0.5 rounded-md text-xs font-medium capitalize', priorityBadge(o.priority)]">
                  {{ o.priority }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span :class="['px-2 py-0.5 rounded-md text-xs font-medium', statusBadge(o.status)]">
                  {{ statusLabel(o.status) }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div :class="['text-sm', isOverdue(o) && 'text-rose-400']">{{ fmtDate(o.scheduled_date) }}</div>
              </td>
              <td class="px-4 py-3">
                <div v-if="o.technician_name" class="flex items-center gap-1.5 text-sm">
                  <User class="w-3.5 h-3.5 text-slate-400" />
                  {{ o.technician_name }}
                </div>
                <div v-else :class="['text-xs', dk('text-slate-600','text-slate-400')]">—</div>
              </td>
              <td class="px-4 py-3">
                <div class="text-sm font-medium">
                  {{ o.actual_cost ? fmt(o.actual_cost) : o.estimated_cost ? fmt(o.estimated_cost) : '—' }}
                </div>
                <div v-if="o.actual_cost && o.estimated_cost" :class="['text-xs', dk('text-slate-500','text-slate-400')]">
                  est. {{ fmt(o.estimated_cost) }}
                </div>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center justify-end gap-1">
                  <button @click="openView(o)"
                    :class="['p-1.5 rounded-lg transition-colors', dk('hover:bg-slate-700 text-slate-400','hover:bg-slate-100 text-slate-500')]"
                    title="View">
                    <Eye class="w-3.5 h-3.5" />
                  </button>
                  <button v-if="!['completed','cancelled'].includes(o.status)"
                    @click="openComplete(o)"
                    :class="['p-1.5 rounded-lg transition-colors', dk('hover:bg-emerald-900/40 text-emerald-400','hover:bg-emerald-50 text-emerald-600')]"
                    title="Complete">
                    <CheckCircle class="w-3.5 h-3.5" />
                  </button>
                  <button @click="openEdit(o)"
                    :class="['p-1.5 rounded-lg transition-colors', dk('hover:bg-slate-700 text-slate-400','hover:bg-slate-100 text-slate-500')]"
                    title="Edit">
                    <Edit2 class="w-3.5 h-3.5" />
                  </button>
                  <button @click="openDelete(o)"
                    :class="['p-1.5 rounded-lg transition-colors', dk('hover:bg-rose-900/40 text-rose-400','hover:bg-rose-50 text-rose-500')]"
                    title="Delete">
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="!loading && filtered.length > perPage"
        :class="['flex items-center justify-between px-4 py-3 border-t text-sm', dk('border-slate-800','border-slate-200')]">
        <span :class="dk('text-slate-400','text-slate-500')">
          {{ (page-1)*perPage+1 }}–{{ Math.min(page*perPage, filtered.length) }} of {{ filtered.length }}
        </span>
        <div class="flex items-center gap-1">
          <button @click="page--" :disabled="page===1"
            :class="['p-1.5 rounded-lg transition-colors disabled:opacity-40',
              dk('hover:bg-slate-700 text-slate-400','hover:bg-slate-100 text-slate-500')]">
            <ChevronLeft class="w-4 h-4" />
          </button>
          <span :class="['px-3 py-1 rounded-lg text-xs', dk('bg-slate-800','bg-slate-100')]">{{ page }}/{{ totalPages }}</span>
          <button @click="page++" :disabled="page===totalPages"
            :class="['p-1.5 rounded-lg transition-colors disabled:opacity-40',
              dk('hover:bg-slate-700 text-slate-400','hover:bg-slate-100 text-slate-500')]">
            <ChevronRight class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>

    <!-- ── Create / Edit Modal ─────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showCreate || showEdit"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
        @click.self="showCreate=false; showEdit=false">
        <div :class="['w-full max-w-2xl rounded-2xl border shadow-2xl', dk('bg-slate-900 border-slate-700','bg-white border-slate-200')]">
          <div :class="['flex items-center justify-between px-6 py-4 border-b', dk('border-slate-800','border-slate-200')]">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-violet-500/15 flex items-center justify-center">
                <Wrench class="w-4 h-4 text-violet-400" />
              </div>
              <h2 class="font-semibold text-base">{{ showEdit ? 'Edit Order' : 'New Maintenance Order' }}</h2>
            </div>
            <button @click="showCreate=false; showEdit=false"
              :class="['p-1.5 rounded-lg transition-colors', dk('hover:bg-slate-800 text-slate-400','hover:bg-slate-100 text-slate-500')]">
              <X class="w-4 h-4" />
            </button>
          </div>
          <div class="p-6 space-y-4 max-h-[70vh] overflow-y-auto">
            <!-- Title -->
            <div>
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Title *</label>
              <input v-model="form.title" placeholder="e.g. Replace hydraulic filter"
                :class="['w-full px-3 py-2 rounded-lg border text-sm',
                  dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500 focus:border-violet-500',
                     'bg-white border-slate-200 text-slate-900 placeholder-slate-400 focus:border-violet-500')]" />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <!-- Type -->
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Type *</label>
                <select v-model="form.order_type"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')]">
                  <option value="corrective">Corrective</option>
                  <option value="preventive">Preventive</option>
                  <option value="inspection">Inspection</option>
                  <option value="emergency">Emergency</option>
                  <option value="upgrade">Upgrade</option>
                </select>
              </div>
              <!-- Priority -->
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Priority *</label>
                <select v-model="form.priority"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')]">
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                  <option value="critical">Critical</option>
                </select>
              </div>
            </div>
            <!-- Equipment -->
            <div>
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Equipment</label>
              <select v-model="form.equipment_id"
                :class="['w-full px-3 py-2 rounded-lg border text-sm',
                  dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')]">
                <option value="">— None —</option>
                <option v-for="e in equipment" :key="e.id" :value="e.id">{{ e.name }} ({{ e.code }})</option>
              </select>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <!-- Technician -->
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Technician Name</label>
                <input v-model="form.technician_name" placeholder="Full name"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                       'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
              </div>
              <!-- Scheduled Date -->
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Scheduled Date</label>
                <input v-model="form.scheduled_date" type="date"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')]" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <!-- Estimated Hours -->
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Estimated Hours</label>
                <input v-model="form.estimated_hours" type="number" step="0.5" min="0" placeholder="0"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                       'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
              </div>
              <!-- Estimated Cost -->
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Estimated Cost (DZD)</label>
                <input v-model="form.estimated_cost" type="number" min="0" placeholder="0"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                       'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
              </div>
            </div>
            <!-- Status (edit only) -->
            <div v-if="showEdit">
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Status</label>
              <select v-model="form.status"
                :class="['w-full px-3 py-2 rounded-lg border text-sm',
                  dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')]">
                <option value="pending">Pending</option>
                <option value="in_progress">In Progress</option>
                <option value="on_hold">On Hold</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </div>
            <!-- Description -->
            <div>
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Description</label>
              <textarea v-model="form.description" rows="3" placeholder="Describe the work required…"
                :class="['w-full px-3 py-2 rounded-lg border text-sm resize-none',
                  dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                     'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
            </div>
          </div>
          <div :class="['flex items-center justify-end gap-3 px-6 py-4 border-t', dk('border-slate-800','border-slate-200')]">
            <button @click="showCreate=false; showEdit=false"
              :class="['px-4 py-2 rounded-lg border text-sm transition-colors',
                dk('border-slate-700 text-slate-300 hover:bg-slate-800','border-slate-200 text-slate-600 hover:bg-slate-50')]">
              Cancel
            </button>
            <button @click="saveOrder(showEdit)" :disabled="saving || !form.title"
              class="px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-500 disabled:opacity-50 text-white text-sm font-medium transition-colors">
              {{ saving ? 'Saving…' : showEdit ? 'Save Changes' : 'Create Order' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Complete Order Modal ────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showComplete && selected"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
        @click.self="showComplete=false">
        <div :class="['w-full max-w-lg rounded-2xl border shadow-2xl', dk('bg-slate-900 border-slate-700','bg-white border-slate-200')]">
          <div :class="['flex items-center justify-between px-6 py-4 border-b', dk('border-slate-800','border-slate-200')]">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-emerald-500/15 flex items-center justify-center">
                <CheckCircle class="w-4 h-4 text-emerald-400" />
              </div>
              <div>
                <h2 class="font-semibold text-base">Complete Order</h2>
                <p :class="['text-xs', dk('text-slate-400','text-slate-500')]">{{ selected.order_number }}</p>
              </div>
            </div>
            <button @click="showComplete=false"
              :class="['p-1.5 rounded-lg', dk('hover:bg-slate-800 text-slate-400','hover:bg-slate-100 text-slate-500')]">
              <X class="w-4 h-4" />
            </button>
          </div>
          <div class="p-6 space-y-4 max-h-[70vh] overflow-y-auto">
            <!-- Work Performed -->
            <div>
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Work Performed *</label>
              <textarea v-model="completeForm.work_performed" rows="3"
                placeholder="Describe what was done…"
                :class="['w-full px-3 py-2 rounded-lg border text-sm resize-none',
                  dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                     'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
            </div>
            <!-- Findings -->
            <div>
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Findings / Notes</label>
              <textarea v-model="completeForm.findings" rows="2"
                placeholder="Observations, issues found…"
                :class="['w-full px-3 py-2 rounded-lg border text-sm resize-none',
                  dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                     'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <!-- Actual Hours -->
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Actual Hours</label>
                <input v-model="completeForm.actual_hours" type="number" step="0.5" min="0"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')]" />
              </div>
              <!-- Technician -->
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Technician Name</label>
                <input v-model="completeForm.technician_name" placeholder="Full name"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                       'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
              </div>
            </div>
            <!-- Cost Breakdown -->
            <div :class="['rounded-xl border p-4 space-y-3', dk('border-slate-700 bg-slate-800/50','border-slate-200 bg-slate-50')]">
              <div class="flex items-center gap-2 text-sm font-medium">
                <DollarSign class="w-4 h-4 text-emerald-400" />
                Cost Breakdown (DZD)
              </div>
              <div class="grid grid-cols-3 gap-3">
                <div>
                  <label :class="['block text-xs mb-1', dk('text-slate-400','text-slate-500')]">Labor</label>
                  <input v-model="completeForm.labor_cost" type="number" min="0" placeholder="0"
                    :class="['w-full px-3 py-2 rounded-lg border text-sm',
                      dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-500',
                         'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
                </div>
                <div>
                  <label :class="['block text-xs mb-1', dk('text-slate-400','text-slate-500')]">Parts</label>
                  <input v-model="completeForm.parts_cost" type="number" min="0" placeholder="0"
                    :class="['w-full px-3 py-2 rounded-lg border text-sm',
                      dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-500',
                         'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
                </div>
                <div>
                  <label :class="['block text-xs mb-1', dk('text-slate-400','text-slate-500')]">Other</label>
                  <input v-model="completeForm.other_cost" type="number" min="0" placeholder="0"
                    :class="['w-full px-3 py-2 rounded-lg border text-sm',
                      dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-500',
                         'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
                </div>
              </div>
              <div :class="['text-sm font-semibold text-right pt-1 border-t', dk('border-slate-700','border-slate-200')]">
                Total: {{ fmt(
                  (parseFloat(completeForm.labor_cost||'0')||0) +
                  (parseFloat(completeForm.parts_cost||'0')||0) +
                  (parseFloat(completeForm.other_cost||'0')||0)
                ) }}
              </div>
            </div>
            <!-- Next Service Date -->
            <div>
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Next Service Date</label>
              <input v-model="completeForm.next_service_date" type="date"
                :class="['w-full px-3 py-2 rounded-lg border text-sm',
                  dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')]" />
            </div>
          </div>
          <div :class="['flex items-center justify-end gap-3 px-6 py-4 border-t', dk('border-slate-800','border-slate-200')]">
            <button @click="showComplete=false"
              :class="['px-4 py-2 rounded-lg border text-sm',
                dk('border-slate-700 text-slate-300 hover:bg-slate-800','border-slate-200 text-slate-600 hover:bg-slate-50')]">
              Cancel
            </button>
            <button @click="completeOrder" :disabled="completing || !completeForm.work_performed"
              class="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white text-sm font-medium transition-colors">
              {{ completing ? 'Completing…' : 'Mark Complete' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── View Modal ──────────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showView && selected"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
        @click.self="showView=false">
        <div :class="['w-full max-w-2xl rounded-2xl border shadow-2xl', dk('bg-slate-900 border-slate-700','bg-white border-slate-200')]">
          <div :class="['flex items-center justify-between px-6 py-4 border-b', dk('border-slate-800','border-slate-200')]">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl flex items-center justify-center"
                :style="{ backgroundColor: (selected.color || typeColor(selected.order_type)) + '22' }">
                <Wrench class="w-4 h-4" :style="{ color: selected.color || typeColor(selected.order_type) }" />
              </div>
              <div>
                <h2 class="font-semibold text-base">{{ selected.title }}</h2>
                <p :class="['text-xs font-mono', dk('text-slate-400','text-slate-500')]">{{ selected.order_number }}</p>
              </div>
            </div>
            <button @click="showView=false"
              :class="['p-1.5 rounded-lg', dk('hover:bg-slate-800 text-slate-400','hover:bg-slate-100 text-slate-500')]">
              <X class="w-4 h-4" />
            </button>
          </div>
          <div class="p-6 space-y-5 max-h-[70vh] overflow-y-auto">
            <!-- Status badges -->
            <div class="flex flex-wrap gap-2">
              <span :class="['px-2.5 py-1 rounded-lg text-xs font-medium', statusBadge(selected.status)]">
                {{ statusLabel(selected.status) }}
              </span>
              <span :class="['px-2.5 py-1 rounded-lg text-xs font-medium', typeBadge(selected.order_type)]">
                {{ typeLabel(selected.order_type) }}
              </span>
              <span :class="['px-2.5 py-1 rounded-lg text-xs font-medium capitalize', priorityBadge(selected.priority)]">
                {{ selected.priority }}
              </span>
            </div>
            <!-- Details Grid -->
            <div class="grid grid-cols-2 gap-4 text-sm">
              <div v-if="selected.equipment_name">
                <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Equipment</span>
                <p class="font-medium mt-0.5">{{ selected.equipment_name }}</p>
              </div>
              <div v-if="selected.technician_name">
                <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Technician</span>
                <p class="font-medium mt-0.5">{{ selected.technician_name }}</p>
              </div>
              <div>
                <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Scheduled</span>
                <p class="font-medium mt-0.5">{{ fmtDate(selected.scheduled_date) }}</p>
              </div>
              <div v-if="selected.actual_end">
                <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Completed</span>
                <p class="font-medium mt-0.5">{{ fmtDate(selected.actual_end) }}</p>
              </div>
              <div>
                <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Est. Hours</span>
                <p class="font-medium mt-0.5">{{ selected.estimated_hours ?? '—' }}</p>
              </div>
              <div v-if="selected.actual_hours">
                <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Actual Hours</span>
                <p class="font-medium mt-0.5">{{ selected.actual_hours }}</p>
              </div>
            </div>
            <!-- Cost -->
            <div :class="['rounded-xl border p-4', dk('border-slate-700 bg-slate-800/50','border-slate-200 bg-slate-50')]">
              <div class="text-xs font-medium mb-3 flex items-center gap-2">
                <DollarSign class="w-3.5 h-3.5 text-teal-400" /> Cost Summary
              </div>
              <div class="grid grid-cols-4 gap-3 text-sm">
                <div>
                  <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Labor</span>
                  <p class="font-medium mt-0.5 text-xs">{{ selected.labor_cost ? fmt(selected.labor_cost) : '—' }}</p>
                </div>
                <div>
                  <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Parts</span>
                  <p class="font-medium mt-0.5 text-xs">{{ selected.parts_cost ? fmt(selected.parts_cost) : '—' }}</p>
                </div>
                <div>
                  <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Other</span>
                  <p class="font-medium mt-0.5 text-xs">{{ selected.other_cost ? fmt(selected.other_cost) : '—' }}</p>
                </div>
                <div>
                  <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Total</span>
                  <p class="font-semibold mt-0.5 text-xs text-emerald-400">
                    {{ selected.actual_cost ? fmt(selected.actual_cost) : selected.estimated_cost ? fmt(selected.estimated_cost) : '—' }}
                  </p>
                </div>
              </div>
            </div>
            <!-- Work performed -->
            <div v-if="selected.work_performed">
              <span :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]">Work Performed</span>
              <p :class="['mt-1.5 text-sm leading-relaxed p-3 rounded-lg', dk('bg-slate-800 text-slate-300','bg-slate-50 text-slate-700')]">
                {{ selected.work_performed }}
              </p>
            </div>
            <div v-if="selected.description">
              <span :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]">Description</span>
              <p :class="['mt-1.5 text-sm leading-relaxed', dk('text-slate-300','text-slate-700')]">{{ selected.description }}</p>
            </div>
            <!-- Parts Lines -->
            <div v-if="selected.lines && selected.lines.length">
              <span :class="['text-xs font-medium flex items-center gap-1.5 mb-2', dk('text-slate-400','text-slate-500')]">
                <Package class="w-3.5 h-3.5" /> Parts Used ({{ selected.lines.length }})
              </span>
              <div :class="['rounded-xl border overflow-hidden', dk('border-slate-700','border-slate-200')]">
                <table class="w-full text-xs">
                  <thead :class="dk('bg-slate-800 text-slate-400','bg-slate-50 text-slate-500')">
                    <tr>
                      <th class="text-left px-3 py-2">Part</th>
                      <th class="text-right px-3 py-2">Qty</th>
                      <th class="text-right px-3 py-2">Unit</th>
                      <th class="text-right px-3 py-2">Total</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y" :class="dk('divide-slate-700','divide-slate-100')">
                    <tr v-for="l in selected.lines" :key="l.id">
                      <td class="px-3 py-2">
                        <div>{{ l.part_name }}</div>
                        <div v-if="l.part_number" :class="dk('text-slate-500','text-slate-400')">{{ l.part_number }}</div>
                      </td>
                      <td class="px-3 py-2 text-right">{{ l.quantity }}</td>
                      <td class="px-3 py-2 text-right">{{ fmt(l.unit_cost) }}</td>
                      <td class="px-3 py-2 text-right font-medium">{{ fmt(l.total_cost) }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
          <div :class="['flex items-center justify-end gap-3 px-6 py-4 border-t', dk('border-slate-800','border-slate-200')]">
            <button v-if="!['completed','cancelled'].includes(selected.status)"
              @click="showView=false; openComplete(selected)"
              class="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium transition-colors flex items-center gap-2">
              <CheckCircle class="w-4 h-4" /> Complete
            </button>
            <button @click="showView=false"
              :class="['px-4 py-2 rounded-lg border text-sm',
                dk('border-slate-700 text-slate-300 hover:bg-slate-800','border-slate-200 text-slate-600 hover:bg-slate-50')]">
              Close
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Delete Confirm Modal ────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showDelete && selected"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
        @click.self="showDelete=false">
        <div :class="['w-full max-w-sm rounded-2xl border shadow-2xl p-6', dk('bg-slate-900 border-slate-700','bg-white border-slate-200')]">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 rounded-xl bg-rose-500/15 flex items-center justify-center">
              <Trash2 class="w-5 h-5 text-rose-400" />
            </div>
            <div>
              <h2 class="font-semibold">Delete Order</h2>
              <p :class="['text-xs', dk('text-slate-400','text-slate-500')]">This action cannot be undone</p>
            </div>
          </div>
          <p :class="['text-sm mb-5', dk('text-slate-300','text-slate-600')]">
            Delete order <span class="font-mono font-semibold text-rose-400">{{ selected.order_number }}</span>?
            All associated records will be removed.
          </p>
          <div class="flex gap-3">
            <button @click="showDelete=false"
              :class="['flex-1 py-2 rounded-lg border text-sm',
                dk('border-slate-700 text-slate-300 hover:bg-slate-800','border-slate-200 text-slate-600 hover:bg-slate-50')]">
              Cancel
            </button>
            <button @click="deleteOrder" :disabled="deleting"
              class="flex-1 py-2 rounded-lg bg-rose-600 hover:bg-rose-500 disabled:opacity-50 text-white text-sm font-medium transition-colors">
              {{ deleting ? 'Deleting…' : 'Delete' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
