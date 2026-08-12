<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { maintenanceAPI } from '@/api/client'
import {
  Search, Plus, Edit2, Trash2, Eye, X, RefreshCw,
  ClipboardList, AlertTriangle, CheckCircle, Clock, Ban
} from 'lucide-vue-next'

const app = useAppStore()

interface MaintenanceRequest {
  id: string
  request_number: string
  title: string
  description: string | null
  priority: string
  status: string
  failure_type: string | null
  symptoms: string | null
  requested_by_name: string | null
  assigned_to_name: string | null
  submitted_at: string | null
  approved_at: string | null
  completed_at: string | null
  estimated_cost: number | null
  actual_cost: number | null
  notes: string | null
  created_at: string
  equipment_id: string | null
  equipment_name: string | null
  equipment_code: string | null
  equipment_location: string | null
}

const items = ref<MaintenanceRequest[]>([])
const total = ref(0)
const loading = ref(false)
const summary = ref<{ status: string; count: number }[]>([])

const equipmentOptions = ref<{ id: string; name: string; code: string }[]>([])

const search = ref('')
const filterStatus = ref('')
const filterPriority = ref('')
const page = ref(1)
const limit = 50

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showViewModal = ref(false)
const showDeleteModal = ref(false)

const selectedItem = ref<MaintenanceRequest | null>(null)
const saving = ref(false)
const deleting = ref(false)

const form = ref({
  equipment_id: '',
  title: '',
  description: '',
  priority: 'medium',
  status: 'draft',
  failure_type: '',
  symptoms: '',
  requested_by_name: '',
  assigned_to_name: '',
  estimated_cost: null as number | null,
  notes: '',
})

const priorityOptions = [
  { value: 'low', label: 'Low' },
  { value: 'medium', label: 'Medium' },
  { value: 'high', label: 'High' },
  { value: 'critical', label: 'Critical' },
]

const statusOptions = [
  { value: 'draft', label: 'Draft' },
  { value: 'submitted', label: 'Submitted' },
  { value: 'approved', label: 'Approved' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'completed', label: 'Completed' },
  { value: 'rejected', label: 'Rejected' },
  { value: 'cancelled', label: 'Cancelled' },
]

const kpis = computed(() => {
  const getCount = (s: string) => summary.value.find(x => x.status === s)?.count ?? 0
  return [
    { label: 'Total Requests', value: total.value, icon: ClipboardList, color: 'indigo' },
    { label: 'Open', value: getCount('submitted') + getCount('approved') + getCount('in_progress'), icon: Clock, color: 'yellow' },
    { label: 'Pending Approval', value: getCount('submitted'), icon: AlertTriangle, color: 'orange' },
    { label: 'Completed', value: getCount('completed'), icon: CheckCircle, color: 'green' },
  ]
})

function priorityColor(p: string) {
  const map: Record<string, string> = {
    low: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
    medium: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
    high: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300',
    critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
  }
  return map[p] ?? 'bg-slate-100 text-slate-600'
}

function statusColor(s: string) {
  const map: Record<string, string> = {
    draft: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
    submitted: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
    approved: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-300',
    in_progress: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300',
    completed: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
    rejected: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
    cancelled: 'bg-slate-100 text-slate-500 dark:bg-slate-700 dark:text-slate-400',
  }
  return map[s] ?? 'bg-slate-100 text-slate-600'
}

function kpiColor(c: string) {
  const map: Record<string, string> = {
    indigo: 'bg-indigo-50 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-300',
    yellow: 'bg-yellow-50 text-yellow-600 dark:bg-yellow-900/30 dark:text-yellow-300',
    orange: 'bg-orange-50 text-orange-600 dark:bg-orange-900/30 dark:text-orange-300',
    green: 'bg-green-50 text-green-600 dark:bg-green-900/30 dark:text-green-300',
  }
  return map[c] ?? 'bg-slate-100 text-slate-600'
}

function fmtCurrency(n: number | null) {
  if (n == null) return '-'
  return new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'
}

function fmtDate(d: string | null) {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('fr-DZ')
}

function fmtDateTime(d: string | null) {
  if (!d) return '-'
  return new Date(d).toLocaleString('fr-DZ')
}

async function load() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (search.value) params.search = search.value
    if (filterStatus.value) params.status = filterStatus.value
    if (filterPriority.value) params.priority = filterPriority.value
    params.page = String(page.value)
    params.limit = String(limit)
    const { data } = await maintenanceAPI.listRequests(params)
    items.value = data.items ?? []
    total.value = data.total ?? 0
    summary.value = data.summary ?? []
  } catch (e: any) {
    app.addToast(e.response?.data?.error ?? 'Failed to load requests', 'error')
  } finally {
    loading.value = false
  }
}

async function loadEquipment() {
  try {
    const { data } = await maintenanceAPI.listEquipment({ limit: '200' })
    equipmentOptions.value = (data.items ?? []).map((e: any) => ({ id: e.id, name: e.name, code: e.code }))
  } catch { /* ignore */ }
}

function resetForm() {
  form.value = {
    equipment_id: '', title: '', description: '',
    priority: 'medium', status: 'draft',
    failure_type: '', symptoms: '',
    requested_by_name: '', assigned_to_name: '',
    estimated_cost: null, notes: '',
  }
}

function openCreate() {
  resetForm()
  showCreateModal.value = true
}

function openEdit(item: MaintenanceRequest) {
  selectedItem.value = item
  form.value = {
    equipment_id: item.equipment_id ?? '',
    title: item.title,
    description: item.description ?? '',
    priority: item.priority,
    status: item.status,
    failure_type: item.failure_type ?? '',
    symptoms: item.symptoms ?? '',
    requested_by_name: item.requested_by_name ?? '',
    assigned_to_name: item.assigned_to_name ?? '',
    estimated_cost: item.estimated_cost ?? null,
    notes: item.notes ?? '',
  }
  showEditModal.value = true
}

function openView(item: MaintenanceRequest) {
  selectedItem.value = item
  showViewModal.value = true
}

function openDelete(item: MaintenanceRequest) {
  selectedItem.value = item
  showDeleteModal.value = true
}

async function save() {
  if (!form.value.title) { app.addToast('Title is required', 'error'); return }
  saving.value = true
  try {
    if (showCreateModal.value) {
      await maintenanceAPI.createRequest({ ...form.value })
      app.addToast('Request created', 'success')
      showCreateModal.value = false
    } else {
      await maintenanceAPI.updateRequest(selectedItem.value!.id, { ...form.value })
      app.addToast('Request updated', 'success')
      showEditModal.value = false
    }
    await load()
  } catch (e: any) {
    app.addToast(e.response?.data?.error ?? 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

async function deleteItem() {
  if (!selectedItem.value) return
  deleting.value = true
  try {
    await maintenanceAPI.deleteRequest(selectedItem.value.id)
    app.addToast('Request deleted', 'success')
    showDeleteModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e.response?.data?.error ?? 'Delete failed', 'error')
  } finally {
    deleting.value = false
  }
}

let searchTimer: ReturnType<typeof setTimeout>
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; load() }, 400)
}

onMounted(() => { load(); loadEquipment() })
</script>

<template>
  <div :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'" class="min-h-screen">
    <div class="max-w-screen-2xl mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">

      <!-- Header -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold">Maintenance Requests</h1>
          <p :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'" class="text-sm mt-1">Track and manage all maintenance requests</p>
        </div>
        <button @click="openCreate"
          class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" /> New Request
        </button>
      </div>

      <!-- KPIs -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div v-for="kpi in kpis" :key="kpi.label"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'"
          class="rounded-xl border p-4 space-y-2">
          <div class="flex items-center justify-between">
            <p :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'" class="text-xs font-medium">{{ kpi.label }}</p>
            <div :class="kpiColor(kpi.color)" class="p-1.5 rounded-lg"><component :is="kpi.icon" class="w-4 h-4" /></div>
          </div>
          <p class="text-2xl font-bold">{{ kpi.value }}</p>
        </div>
      </div>

      <!-- Filters -->
      <div :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'"
        class="rounded-xl border p-4 flex flex-col sm:flex-row gap-3">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="search" @input="onSearch" placeholder="Search by number or title..."
            :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500' : 'bg-slate-50 border-slate-200 text-slate-900 placeholder-slate-400'"
            class="w-full pl-9 pr-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
        </div>
        <select v-model="filterStatus" @change="page = 1; load()"
          :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-slate-50 border-slate-200 text-slate-900'"
          class="px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 min-w-[160px]">
          <option value="">All Statuses</option>
          <option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
        </select>
        <select v-model="filterPriority" @change="page = 1; load()"
          :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-slate-50 border-slate-200 text-slate-900'"
          class="px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 min-w-[140px]">
          <option value="">All Priorities</option>
          <option v-for="p in priorityOptions" :key="p.value" :value="p.value">{{ p.label }}</option>
        </select>
        <button @click="search = ''; filterStatus = ''; filterPriority = ''; page = 1; load()"
          :class="app.darkMode ? 'bg-slate-800 hover:bg-slate-700 text-slate-300' : 'bg-slate-100 hover:bg-slate-200 text-slate-600'"
          class="px-3 py-2 rounded-lg text-sm transition-colors flex items-center gap-2">
          <RefreshCw class="w-4 h-4" /> Reset
        </button>
      </div>

      <!-- Table -->
      <div :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'" class="rounded-xl border overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr :class="app.darkMode ? 'bg-slate-800 text-slate-300' : 'bg-slate-50 text-slate-600'" class="text-xs uppercase tracking-wider">
                <th class="px-4 py-3 text-left">Number / Title</th>
                <th class="px-4 py-3 text-left">Equipment</th>
                <th class="px-4 py-3 text-left">Priority</th>
                <th class="px-4 py-3 text-left">Status</th>
                <th class="px-4 py-3 text-left">Requested By</th>
                <th class="px-4 py-3 text-right">Est. Cost</th>
                <th class="px-4 py-3 text-left">Date</th>
                <th class="px-4 py-3 text-center">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
              <tr v-if="loading">
                <td colspan="8" class="text-center py-12">
                  <RefreshCw class="w-6 h-6 animate-spin mx-auto text-indigo-500" />
                </td>
              </tr>
              <tr v-else-if="!items.length">
                <td colspan="8" class="text-center py-12">
                  <ClipboardList class="w-10 h-10 mx-auto mb-2 opacity-20" />
                  <p :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">No requests found</p>
                </td>
              </tr>
              <tr v-for="item in items" :key="item.id"
                :class="app.darkMode ? 'hover:bg-slate-800' : 'hover:bg-slate-50'" class="transition-colors">
                <td class="px-4 py-3">
                  <p class="font-mono font-semibold text-xs">{{ item.request_number }}</p>
                  <p :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'" class="text-sm truncate max-w-[200px]">{{ item.title }}</p>
                </td>
                <td class="px-4 py-3">
                  <p v-if="item.equipment_name" class="text-sm truncate max-w-[140px]">{{ item.equipment_name }}</p>
                  <p v-else :class="app.darkMode ? 'text-slate-600' : 'text-slate-400'" class="text-xs">-</p>
                </td>
                <td class="px-4 py-3">
                  <span :class="priorityColor(item.priority)" class="px-2 py-0.5 rounded-full text-xs font-medium capitalize">
                    {{ item.priority }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span :class="statusColor(item.status)" class="px-2 py-0.5 rounded-full text-xs font-medium capitalize">
                    {{ item.status.replace('_', ' ') }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <p :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'" class="text-xs truncate max-w-[120px]">
                    {{ item.requested_by_name ?? '-' }}
                  </p>
                </td>
                <td class="px-4 py-3 text-right font-mono text-xs">{{ fmtCurrency(item.estimated_cost) }}</td>
                <td class="px-4 py-3 text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ fmtDate(item.created_at) }}</td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-center gap-1">
                    <button @click="openView(item)" :class="app.darkMode ? 'hover:bg-slate-700 text-slate-400' : 'hover:bg-slate-100 text-slate-500'" class="p-1.5 rounded-lg transition-colors"><Eye class="w-4 h-4" /></button>
                    <button @click="openEdit(item)" :class="app.darkMode ? 'hover:bg-slate-700 text-slate-400' : 'hover:bg-slate-100 text-slate-500'" class="p-1.5 rounded-lg transition-colors"><Edit2 class="w-4 h-4" /></button>
                    <button @click="openDelete(item)" class="p-1.5 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/30 text-red-500 transition-colors"><Trash2 class="w-4 h-4" /></button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-if="total > limit" :class="app.darkMode ? 'border-slate-800 text-slate-400' : 'border-slate-100 text-slate-500'"
          class="px-4 py-3 border-t flex items-center justify-between text-sm">
          <span>{{ (page - 1) * limit + 1 }}–{{ Math.min(page * limit, total) }} of {{ total }}</span>
          <div class="flex gap-2">
            <button @click="page--; load()" :disabled="page === 1" class="px-3 py-1 rounded border disabled:opacity-40" :class="app.darkMode ? 'border-slate-700 hover:bg-slate-800' : 'border-slate-200 hover:bg-slate-50'">Prev</button>
            <button @click="page++; load()" :disabled="page * limit >= total" class="px-3 py-1 rounded border disabled:opacity-40" :class="app.darkMode ? 'border-slate-700 hover:bg-slate-800' : 'border-slate-200 hover:bg-slate-50'">Next</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showCreateModal || showEditModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showCreateModal = false; showEditModal = false" />
        <div :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'"
          class="relative z-10 w-full max-w-xl max-h-[90vh] overflow-y-auto rounded-2xl border shadow-2xl">
          <div class="sticky top-0 z-10 flex items-center justify-between px-6 py-4 border-b"
            :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'">
            <h2 class="text-lg font-bold">{{ showCreateModal ? 'New Request' : 'Edit Request' }}</h2>
            <button @click="showCreateModal = false; showEditModal = false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800"><X class="w-5 h-5" /></button>
          </div>
          <div class="p-6 space-y-4">
            <div class="space-y-1">
              <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Equipment</label>
              <select v-model="form.equipment_id" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="">Select equipment...</option>
                <option v-for="e in equipmentOptions" :key="e.id" :value="e.id">{{ e.code }} — {{ e.name }}</option>
              </select>
            </div>
            <div class="space-y-1">
              <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Title *</label>
              <input v-model="form.title" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Priority</label>
                <select v-model="form.priority" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                  <option v-for="p in priorityOptions" :key="p.value" :value="p.value">{{ p.label }}</option>
                </select>
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Status</label>
                <select v-model="form.status" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                  <option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
                </select>
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Failure Type</label>
                <input v-model="form.failure_type" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Est. Cost (DZD)</label>
                <input type="number" v-model="form.estimated_cost" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Requested By</label>
                <input v-model="form.requested_by_name" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Assigned To</label>
                <input v-model="form.assigned_to_name" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
            </div>
            <div class="space-y-1">
              <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Symptoms</label>
              <textarea v-model="form.symptoms" rows="2" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none" />
            </div>
            <div class="space-y-1">
              <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Description</label>
              <textarea v-model="form.description" rows="2" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none" />
            </div>
          </div>
          <div class="sticky bottom-0 flex items-center justify-end gap-3 px-6 py-4 border-t"
            :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'">
            <button @click="showCreateModal = false; showEditModal = false"
              :class="app.darkMode ? 'bg-slate-800 text-slate-300' : 'bg-slate-100 text-slate-700'"
              class="px-4 py-2 rounded-lg text-sm font-medium">Cancel</button>
            <button @click="save" :disabled="saving"
              class="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium disabled:opacity-60 flex items-center gap-2">
              <RefreshCw v-if="saving" class="w-4 h-4 animate-spin" />
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- View Modal -->
    <Teleport to="body">
      <div v-if="showViewModal && selectedItem" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showViewModal = false" />
        <div :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'"
          class="relative z-10 w-full max-w-lg max-h-[85vh] overflow-y-auto rounded-2xl border shadow-2xl">
          <div class="flex items-center justify-between px-6 py-4 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
            <div>
              <p class="font-mono text-xs text-indigo-400">{{ selectedItem.request_number }}</p>
              <h2 class="text-lg font-bold">{{ selectedItem.title }}</h2>
            </div>
            <button @click="showViewModal = false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800"><X class="w-5 h-5" /></button>
          </div>
          <div class="p-6 space-y-4 text-sm">
            <div class="flex gap-2 flex-wrap">
              <span :class="priorityColor(selectedItem.priority)" class="px-2 py-0.5 rounded-full text-xs font-medium capitalize">{{ selectedItem.priority }}</span>
              <span :class="statusColor(selectedItem.status)" class="px-2 py-0.5 rounded-full text-xs font-medium">{{ selectedItem.status.replace('_', ' ') }}</span>
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div><p class="text-xs text-slate-500">Equipment</p><p class="font-medium">{{ selectedItem.equipment_name ?? '-' }}</p></div>
              <div><p class="text-xs text-slate-500">Failure Type</p><p class="font-medium">{{ selectedItem.failure_type ?? '-' }}</p></div>
              <div><p class="text-xs text-slate-500">Requested By</p><p class="font-medium">{{ selectedItem.requested_by_name ?? '-' }}</p></div>
              <div><p class="text-xs text-slate-500">Assigned To</p><p class="font-medium">{{ selectedItem.assigned_to_name ?? '-' }}</p></div>
              <div><p class="text-xs text-slate-500">Est. Cost</p><p class="font-medium">{{ fmtCurrency(selectedItem.estimated_cost) }}</p></div>
              <div><p class="text-xs text-slate-500">Actual Cost</p><p class="font-medium">{{ fmtCurrency(selectedItem.actual_cost) }}</p></div>
              <div><p class="text-xs text-slate-500">Submitted</p><p class="font-medium">{{ fmtDateTime(selectedItem.submitted_at) }}</p></div>
              <div><p class="text-xs text-slate-500">Completed</p><p class="font-medium">{{ fmtDateTime(selectedItem.completed_at) }}</p></div>
            </div>
            <div v-if="selectedItem.symptoms">
              <p class="text-xs text-slate-500 mb-1">Symptoms</p>
              <p :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-50'" class="p-3 rounded-lg text-sm">{{ selectedItem.symptoms }}</p>
            </div>
            <div v-if="selectedItem.description">
              <p class="text-xs text-slate-500 mb-1">Description</p>
              <p :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-50'" class="p-3 rounded-lg text-sm">{{ selectedItem.description }}</p>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Delete Modal -->
    <Teleport to="body">
      <div v-if="showDeleteModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showDeleteModal = false" />
        <div :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'"
          class="relative z-10 w-full max-w-sm rounded-2xl border shadow-2xl p-6 space-y-4">
          <h3 class="text-lg font-bold">Delete Request</h3>
          <p :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'" class="text-sm">Delete <strong>{{ selectedItem?.request_number }}</strong>?</p>
          <div class="flex gap-3 justify-end">
            <button @click="showDeleteModal = false" :class="app.darkMode ? 'bg-slate-800 text-slate-300' : 'bg-slate-100 text-slate-700'" class="px-4 py-2 rounded-lg text-sm font-medium">Cancel</button>
            <button @click="deleteItem" :disabled="deleting" class="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-700 text-white text-sm font-medium disabled:opacity-60 flex items-center gap-2">
              <RefreshCw v-if="deleting" class="w-4 h-4 animate-spin" /> Delete
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
