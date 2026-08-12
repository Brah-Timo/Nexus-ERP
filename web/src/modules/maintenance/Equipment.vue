<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { maintenanceAPI } from '@/api/client'
import {
  Search, Plus, Edit2, Trash2, Eye, X, ChevronDown, ChevronUp,
  Settings, MapPin, Building2, Calendar, DollarSign, AlertTriangle,
  CheckCircle, Clock, Archive, Tag, Wrench, RefreshCw, Filter
} from 'lucide-vue-next'

const app = useAppStore()

interface Equipment {
  id: string
  code: string
  name: string
  category: string | null
  subcategory: string | null
  location: string | null
  department: string | null
  status: string
  purchase_date: string | null
  purchase_cost: number
  current_value: number
  warranty_expiry: string | null
  manufacturer: string | null
  model: string | null
  serial_number: string | null
  asset_tag: string | null
  last_maintenance_date: string | null
  next_maintenance_date: string | null
  maintenance_interval_days: number | null
  expected_life_years: number | null
  notes: string | null
  is_active: boolean
  created_at: string
  open_orders: number
  open_requests: number
  overdue: boolean
}

interface StatusSummary { status: string; count: number; value: number }
interface CategorySummary { category: string; count: number }

const items = ref<Equipment[]>([])
const total = ref(0)
const loading = ref(false)
const statusSummary = ref<StatusSummary[]>([])
const categorySummary = ref<CategorySummary[]>([])
const equipmentList = ref<{ id: string; name: string; code: string }[]>([])

const search = ref('')
const filterStatus = ref('')
const filterCategory = ref('')
const page = ref(1)
const limit = 50

const showCreateModal = ref(false)
const showEditModal = ref(false)
const showViewModal = ref(false)
const showDeleteModal = ref(false)

const selectedItem = ref<Equipment | null>(null)
const viewItem = ref<any>(null)
const saving = ref(false)
const deleting = ref(false)

const form = ref({
  code: '', name: '', category: '', subcategory: '',
  location: '', department: '', status: 'active',
  purchase_date: '', purchase_cost: 0, current_value: 0,
  warranty_expiry: '', manufacturer: '', model: '',
  serial_number: '', asset_tag: '',
  last_maintenance_date: '', next_maintenance_date: '',
  maintenance_interval_days: 90, expected_life_years: null as number | null,
  notes: '', is_active: true
})

const statusOptions = [
  { value: 'active', label: 'Active' },
  { value: 'inactive', label: 'Inactive' },
  { value: 'under_maintenance', label: 'Under Maintenance' },
  { value: 'decommissioned', label: 'Decommissioned' },
  { value: 'reserved', label: 'Reserved' },
]

const kpis = computed(() => {
  const active = statusSummary.value.find(s => s.status === 'active')
  const underMaint = statusSummary.value.find(s => s.status === 'under_maintenance')
  const overdue = items.value.filter(i => i.overdue).length
  const totalValue = statusSummary.value.reduce((sum, s) => sum + (s.value || 0), 0)
  return [
    { label: 'Total Equipment', value: total.value, icon: Settings, color: 'indigo' },
    { label: 'Active', value: active?.count ?? 0, icon: CheckCircle, color: 'green' },
    { label: 'Under Maintenance', value: underMaint?.count ?? 0, icon: Wrench, color: 'yellow' },
    { label: 'Overdue Maintenance', value: overdue, icon: AlertTriangle, color: 'red' },
    { label: 'Total Asset Value', value: fmtCurrency(totalValue), icon: DollarSign, color: 'blue', raw: true },
  ]
})

function statusColor(status: string) {
  const map: Record<string, string> = {
    active: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-300',
    inactive: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
    under_maintenance: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300',
    decommissioned: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
    reserved: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-300',
  }
  return map[status] ?? 'bg-slate-100 text-slate-600'
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    active: 'Active', inactive: 'Inactive',
    under_maintenance: 'Under Maintenance',
    decommissioned: 'Decommissioned', reserved: 'Reserved',
  }
  return map[status] ?? status
}

function kpiColor(color: string) {
  const map: Record<string, string> = {
    indigo: 'bg-indigo-50 text-indigo-600 dark:bg-indigo-900/30 dark:text-indigo-300',
    green: 'bg-green-50 text-green-600 dark:bg-green-900/30 dark:text-green-300',
    yellow: 'bg-yellow-50 text-yellow-600 dark:bg-yellow-900/30 dark:text-yellow-300',
    red: 'bg-red-50 text-red-600 dark:bg-red-900/30 dark:text-red-300',
    blue: 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-300',
  }
  return map[color] ?? 'bg-slate-100 text-slate-600'
}

function fmtCurrency(n: number) {
  return new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'
}

function fmtDate(d: string | null) {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('fr-DZ')
}

function isOverdueDate(d: string | null) {
  if (!d) return false
  return new Date(d) < new Date()
}

function daysUntil(d: string | null) {
  if (!d) return null
  const diff = Math.ceil((new Date(d).getTime() - Date.now()) / 86400000)
  return diff
}

async function load() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (search.value) params.search = search.value
    if (filterStatus.value) params.status = filterStatus.value
    if (filterCategory.value) params.category = filterCategory.value
    params.page = String(page.value)
    params.limit = String(limit)
    const { data } = await maintenanceAPI.listEquipment(params)
    items.value = data.items ?? []
    total.value = data.total ?? 0
    statusSummary.value = data.status_summary ?? []
    categorySummary.value = data.category_summary ?? []
  } catch (e: any) {
    app.addToast(e.response?.data?.error ?? 'Failed to load equipment', 'error')
  } finally {
    loading.value = false
  }
}

function resetForm() {
  form.value = {
    code: '', name: '', category: '', subcategory: '',
    location: '', department: '', status: 'active',
    purchase_date: '', purchase_cost: 0, current_value: 0,
    warranty_expiry: '', manufacturer: '', model: '',
    serial_number: '', asset_tag: '',
    last_maintenance_date: '', next_maintenance_date: '',
    maintenance_interval_days: 90, expected_life_years: null,
    notes: '', is_active: true
  }
}

function openCreate() {
  resetForm()
  showCreateModal.value = true
}

function openEdit(item: Equipment) {
  selectedItem.value = item
  form.value = {
    code: item.code,
    name: item.name,
    category: item.category ?? '',
    subcategory: item.subcategory ?? '',
    location: item.location ?? '',
    department: item.department ?? '',
    status: item.status,
    purchase_date: item.purchase_date ?? '',
    purchase_cost: item.purchase_cost,
    current_value: item.current_value,
    warranty_expiry: item.warranty_expiry ?? '',
    manufacturer: item.manufacturer ?? '',
    model: item.model ?? '',
    serial_number: item.serial_number ?? '',
    asset_tag: item.asset_tag ?? '',
    last_maintenance_date: item.last_maintenance_date ?? '',
    next_maintenance_date: item.next_maintenance_date ?? '',
    maintenance_interval_days: item.maintenance_interval_days ?? 90,
    expected_life_years: item.expected_life_years ?? null,
    notes: item.notes ?? '',
    is_active: item.is_active,
  }
  showEditModal.value = true
}

async function openView(item: Equipment) {
  try {
    const { data } = await maintenanceAPI.getEquipment(item.id)
    viewItem.value = data
    showViewModal.value = true
  } catch {
    viewItem.value = item
    showViewModal.value = true
  }
}

function openDelete(item: Equipment) {
  selectedItem.value = item
  showDeleteModal.value = true
}

async function save() {
  if (!form.value.name) { app.addToast('Name is required', 'error'); return }
  saving.value = true
  try {
    const payload = { ...form.value }
    if (showCreateModal.value) {
      await maintenanceAPI.createEquipment(payload)
      app.addToast('Equipment created', 'success')
      showCreateModal.value = false
    } else {
      await maintenanceAPI.updateEquipment(selectedItem.value!.id, payload)
      app.addToast('Equipment updated', 'success')
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
    await maintenanceAPI.deleteEquipment(selectedItem.value.id)
    app.addToast('Equipment deleted', 'success')
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

onMounted(load)
</script>

<template>
  <div :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'" class="min-h-screen">
    <div class="max-w-screen-2xl mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">

      <!-- Header -->
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 class="text-2xl font-bold">Equipment Registry</h1>
          <p :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'" class="text-sm mt-1">
            Manage and track all company equipment and assets
          </p>
        </div>
        <button @click="openCreate"
          class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" />
          Add Equipment
        </button>
      </div>

      <!-- KPI Cards -->
      <div class="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-5 gap-4">
        <div v-for="kpi in kpis" :key="kpi.label"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'"
          class="rounded-xl border p-4 space-y-2">
          <div class="flex items-center justify-between">
            <p :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'" class="text-xs font-medium truncate">{{ kpi.label }}</p>
            <div :class="kpiColor(kpi.color)" class="p-1.5 rounded-lg">
              <component :is="kpi.icon" class="w-4 h-4" />
            </div>
          </div>
          <p class="text-xl font-bold truncate">{{ kpi.raw ? kpi.value : kpi.value }}</p>
        </div>
      </div>

      <!-- Category Cards -->
      <div v-if="categorySummary.length" class="flex gap-3 flex-wrap">
        <button
          v-for="cat in categorySummary" :key="cat.category"
          @click="filterCategory = filterCategory === cat.category ? '' : cat.category; load()"
          :class="[
            filterCategory === cat.category
              ? 'bg-indigo-600 text-white border-indigo-600'
              : app.darkMode
                ? 'bg-slate-900 border-slate-700 text-slate-300 hover:border-indigo-500'
                : 'bg-white border-slate-200 text-slate-700 hover:border-indigo-400'
          ]"
          class="flex items-center gap-2 px-3 py-1.5 rounded-lg border text-sm transition-all">
          <Tag class="w-3.5 h-3.5" />
          {{ cat.category }}
          <span class="font-semibold">{{ cat.count }}</span>
        </button>
      </div>

      <!-- Filters -->
      <div :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'"
        class="rounded-xl border p-4 flex flex-col sm:flex-row gap-3">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="search" @input="onSearch" placeholder="Search by name, code, serial..."
            :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500' : 'bg-slate-50 border-slate-200 text-slate-900 placeholder-slate-400'"
            class="w-full pl-9 pr-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
        </div>
        <select v-model="filterStatus" @change="page = 1; load()"
          :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-slate-50 border-slate-200 text-slate-900'"
          class="px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 min-w-[160px]">
          <option value="">All Statuses</option>
          <option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
        </select>
        <button @click="search = ''; filterStatus = ''; filterCategory = ''; page = 1; load()"
          :class="app.darkMode ? 'bg-slate-800 hover:bg-slate-700 text-slate-300' : 'bg-slate-100 hover:bg-slate-200 text-slate-600'"
          class="px-3 py-2 rounded-lg text-sm transition-colors flex items-center gap-2">
          <RefreshCw class="w-4 h-4" /> Reset
        </button>
      </div>

      <!-- Table -->
      <div :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'"
        class="rounded-xl border overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr :class="app.darkMode ? 'bg-slate-800 text-slate-300' : 'bg-slate-50 text-slate-600'"
                class="text-xs uppercase tracking-wider">
                <th class="px-4 py-3 text-left">Code / Name</th>
                <th class="px-4 py-3 text-left">Category</th>
                <th class="px-4 py-3 text-left">Location</th>
                <th class="px-4 py-3 text-left">Status</th>
                <th class="px-4 py-3 text-left">Next Maintenance</th>
                <th class="px-4 py-3 text-right">Value (DZD)</th>
                <th class="px-4 py-3 text-center">Orders</th>
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
                  <Settings class="w-10 h-10 mx-auto mb-2 opacity-20" />
                  <p :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">No equipment found</p>
                </td>
              </tr>
              <tr v-for="item in items" :key="item.id"
                :class="[
                  item.overdue ? (app.darkMode ? 'bg-red-950/20' : 'bg-red-50/50') : '',
                  app.darkMode ? 'hover:bg-slate-800' : 'hover:bg-slate-50'
                ]"
                class="transition-colors">
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2">
                    <AlertTriangle v-if="item.overdue" class="w-4 h-4 text-red-500 flex-shrink-0" />
                    <div>
                      <p class="font-semibold">{{ item.code }}</p>
                      <p :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'" class="text-xs">{{ item.name }}</p>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <span :class="app.darkMode ? 'text-slate-300' : 'text-slate-600'" class="text-xs">
                    {{ item.category ?? '-' }}
                    <span v-if="item.subcategory" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">/ {{ item.subcategory }}</span>
                  </span>
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-1 text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
                    <MapPin class="w-3 h-3" />
                    {{ item.location ?? '-' }}
                  </div>
                </td>
                <td class="px-4 py-3">
                  <span :class="statusColor(item.status)" class="px-2 py-0.5 rounded-full text-xs font-medium">
                    {{ statusLabel(item.status) }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span v-if="item.next_maintenance_date"
                    :class="isOverdueDate(item.next_maintenance_date) ? 'text-red-500 font-semibold' : (app.darkMode ? 'text-slate-300' : 'text-slate-700')"
                    class="text-xs">
                    {{ fmtDate(item.next_maintenance_date) }}
                    <span v-if="daysUntil(item.next_maintenance_date) !== null"
                      :class="daysUntil(item.next_maintenance_date)! < 0 ? 'text-red-500' : 'text-slate-400'"
                      class="ml-1">
                      ({{ daysUntil(item.next_maintenance_date)! < 0 ? 'overdue' : daysUntil(item.next_maintenance_date) + 'd' }})
                    </span>
                  </span>
                  <span v-else :class="app.darkMode ? 'text-slate-600' : 'text-slate-400'" class="text-xs">-</span>
                </td>
                <td class="px-4 py-3 text-right">
                  <span :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'" class="text-xs font-mono">
                    {{ fmtCurrency(item.current_value) }}
                  </span>
                </td>
                <td class="px-4 py-3 text-center">
                  <div class="flex items-center justify-center gap-2">
                    <span v-if="item.open_orders" class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300 text-xs">
                      <Wrench class="w-3 h-3" /> {{ item.open_orders }}
                    </span>
                    <span v-if="item.open_requests" class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300 text-xs">
                      {{ item.open_requests }} req
                    </span>
                    <span v-if="!item.open_orders && !item.open_requests" :class="app.darkMode ? 'text-slate-600' : 'text-slate-400'" class="text-xs">-</span>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-center gap-1">
                    <button @click="openView(item)" :class="app.darkMode ? 'hover:bg-slate-700 text-slate-400' : 'hover:bg-slate-100 text-slate-500'" class="p-1.5 rounded-lg transition-colors"><Eye class="w-4 h-4" /></button>
                    <button @click="openEdit(item)" :class="app.darkMode ? 'hover:bg-slate-700 text-slate-400' : 'hover:bg-slate-100 text-slate-500'" class="p-1.5 rounded-lg transition-colors"><Edit2 class="w-4 h-4" /></button>
                    <button @click="openDelete(item)" class="p-1.5 rounded-lg transition-colors hover:bg-red-100 dark:hover:bg-red-900/30 text-red-500"><Trash2 class="w-4 h-4" /></button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <!-- Pagination -->
        <div v-if="total > limit" :class="app.darkMode ? 'border-slate-800 text-slate-400' : 'border-slate-100 text-slate-500'"
          class="px-4 py-3 border-t flex items-center justify-between text-sm">
          <span>Showing {{ (page - 1) * limit + 1 }}–{{ Math.min(page * limit, total) }} of {{ total }}</span>
          <div class="flex gap-2">
            <button @click="page--; load()" :disabled="page === 1"
              class="px-3 py-1 rounded border disabled:opacity-40"
              :class="app.darkMode ? 'border-slate-700 hover:bg-slate-800' : 'border-slate-200 hover:bg-slate-50'">Prev</button>
            <button @click="page++; load()" :disabled="page * limit >= total"
              class="px-3 py-1 rounded border disabled:opacity-40"
              :class="app.darkMode ? 'border-slate-700 hover:bg-slate-800' : 'border-slate-200 hover:bg-slate-50'">Next</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showCreateModal || showEditModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showCreateModal = false; showEditModal = false" />
        <div :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'"
          class="relative z-10 w-full max-w-2xl max-h-[90vh] overflow-y-auto rounded-2xl border shadow-2xl">
          <div class="sticky top-0 z-10 flex items-center justify-between px-6 py-4 border-b"
            :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'">
            <h2 class="text-lg font-bold">{{ showCreateModal ? 'Add Equipment' : 'Edit Equipment' }}</h2>
            <button @click="showCreateModal = false; showEditModal = false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Code *</label>
                <input v-model="form.code" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Name *</label>
                <input v-model="form.name" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Category</label>
                <input v-model="form.category" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Subcategory</label>
                <input v-model="form.subcategory" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Location</label>
                <input v-model="form.location" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Department</label>
                <input v-model="form.department" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Status</label>
                <select v-model="form.status" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                  <option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
                </select>
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Manufacturer</label>
                <input v-model="form.manufacturer" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Model</label>
                <input v-model="form.model" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Serial Number</label>
                <input v-model="form.serial_number" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Asset Tag</label>
                <input v-model="form.asset_tag" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Purchase Date</label>
                <input type="date" v-model="form.purchase_date" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Purchase Cost (DZD)</label>
                <input type="number" v-model="form.purchase_cost" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Current Value (DZD)</label>
                <input type="number" v-model="form.current_value" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Warranty Expiry</label>
                <input type="date" v-model="form.warranty_expiry" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Last Maintenance</label>
                <input type="date" v-model="form.last_maintenance_date" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Next Maintenance</label>
                <input type="date" v-model="form.next_maintenance_date" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="space-y-1">
                <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Maintenance Interval (days)</label>
                <input type="number" v-model="form.maintenance_interval_days" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                  class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
            </div>
            <div class="space-y-1">
              <label class="text-xs font-medium" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Notes</label>
              <textarea v-model="form.notes" rows="2" :class="app.darkMode ? 'bg-slate-800 border-slate-700 text-slate-100' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 rounded-lg border text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none" />
            </div>
          </div>
          <div class="sticky bottom-0 flex items-center justify-end gap-3 px-6 py-4 border-t"
            :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'">
            <button @click="showCreateModal = false; showEditModal = false"
              :class="app.darkMode ? 'bg-slate-800 hover:bg-slate-700 text-slate-300' : 'bg-slate-100 hover:bg-slate-200 text-slate-700'"
              class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="save" :disabled="saving"
              class="px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium transition-colors disabled:opacity-60 flex items-center gap-2">
              <RefreshCw v-if="saving" class="w-4 h-4 animate-spin" />
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- View Modal -->
    <Teleport to="body">
      <div v-if="showViewModal && viewItem" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showViewModal = false" />
        <div :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'"
          class="relative z-10 w-full max-w-xl max-h-[85vh] overflow-y-auto rounded-2xl border shadow-2xl">
          <div class="flex items-center justify-between px-6 py-4 border-b"
            :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
            <div>
              <h2 class="text-lg font-bold">{{ viewItem.code }} — {{ viewItem.name }}</h2>
              <span :class="statusColor(viewItem.status)" class="text-xs px-2 py-0.5 rounded-full font-medium">{{ statusLabel(viewItem.status) }}</span>
            </div>
            <button @click="showViewModal = false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800"><X class="w-5 h-5" /></button>
          </div>
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-3 text-sm">
              <div v-for="field in [
                { label: 'Category', value: viewItem.category ?? '-' },
                { label: 'Location', value: viewItem.location ?? '-' },
                { label: 'Manufacturer', value: viewItem.manufacturer ?? '-' },
                { label: 'Model', value: viewItem.model ?? '-' },
                { label: 'Serial Number', value: viewItem.serial_number ?? '-' },
                { label: 'Asset Tag', value: viewItem.asset_tag ?? '-' },
                { label: 'Purchase Cost', value: fmtCurrency(viewItem.purchase_cost) },
                { label: 'Current Value', value: fmtCurrency(viewItem.current_value) },
                { label: 'Last Maintenance', value: fmtDate(viewItem.last_maintenance_date) },
                { label: 'Next Maintenance', value: fmtDate(viewItem.next_maintenance_date) },
              ]" :key="field.label">
                <div>
                  <p :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'" class="text-xs">{{ field.label }}</p>
                  <p class="font-medium truncate">{{ field.value }}</p>
                </div>
              </div>
            </div>
            <div v-if="viewItem.recent_history?.length">
              <p class="text-xs font-semibold uppercase tracking-wider mb-2" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Recent History</p>
              <div class="space-y-2">
                <div v-for="h in viewItem.recent_history" :key="h.id"
                  :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-50'"
                  class="flex items-center justify-between rounded-lg px-3 py-2 text-xs">
                  <div>
                    <p class="font-medium">{{ h.title }}</p>
                    <p :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ fmtDate(h.performed_date) }}</p>
                  </div>
                  <span class="font-mono">{{ fmtCurrency(h.cost) }}</span>
                </div>
              </div>
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
          <h3 class="text-lg font-bold">Delete Equipment</h3>
          <p :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'" class="text-sm">
            Are you sure you want to delete <strong>{{ selectedItem?.name }}</strong>?
          </p>
          <div class="flex gap-3 justify-end">
            <button @click="showDeleteModal = false"
              :class="app.darkMode ? 'bg-slate-800 text-slate-300' : 'bg-slate-100 text-slate-700'"
              class="px-4 py-2 rounded-lg text-sm font-medium">Cancel</button>
            <button @click="deleteItem" :disabled="deleting"
              class="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-700 text-white text-sm font-medium disabled:opacity-60 flex items-center gap-2">
              <RefreshCw v-if="deleting" class="w-4 h-4 animate-spin" />
              Delete
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
