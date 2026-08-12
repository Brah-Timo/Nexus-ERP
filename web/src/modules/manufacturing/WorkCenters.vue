<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Search, Plus, RefreshCw, Edit2, Trash2, ChevronDown, ChevronUp,
  Factory, X, Check, AlertCircle, DollarSign, Clock, Activity,
  ToggleLeft, ToggleRight
} from 'lucide-vue-next'
import { manufacturingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ────────────────────────────────────────────────────────────────────
interface WorkCenter {
  id: string
  code: string
  name: string
  capacity: number
  cost_per_hour: number
  account_id: string | null
  is_active: boolean
  created_at: string
}

// ─── State ────────────────────────────────────────────────────────────────────
const workCenters = ref<WorkCenter[]>([])
const loading = ref(false)
const search = ref('')
const filterActive = ref('')
const sortField = ref('name')
const sortDir = ref<'asc' | 'desc'>('asc')

const showModal = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const saving = ref(false)

const confirmDelete = ref<WorkCenter | null>(null)
const deleting = ref(false)

const form = ref({
  id: '',
  code: '',
  name: '',
  capacity: 8,
  cost_per_hour: 0,
  account_id: null as string | null,
  is_active: true
})

// ─── Computed ─────────────────────────────────────────────────────────────────
const filtered = computed(() => {
  let list = [...workCenters.value]
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(wc =>
      wc.code.toLowerCase().includes(q) ||
      wc.name.toLowerCase().includes(q)
    )
  }
  if (filterActive.value === 'true') list = list.filter(wc => wc.is_active)
  if (filterActive.value === 'false') list = list.filter(wc => !wc.is_active)
  list.sort((a, b) => {
    const av = (a as any)[sortField.value] ?? ''
    const bv = (b as any)[sortField.value] ?? ''
    const cmp = String(av).localeCompare(String(bv))
    return sortDir.value === 'asc' ? cmp : -cmp
  })
  return list
})

const kpis = computed(() => ({
  total: workCenters.value.length,
  active: workCenters.value.filter(wc => wc.is_active).length,
  totalCapacity: workCenters.value.filter(wc => wc.is_active).reduce((s, wc) => s + wc.capacity, 0),
  avgCostPerHour: workCenters.value.filter(wc => wc.is_active).length > 0
    ? workCenters.value.filter(wc => wc.is_active).reduce((s, wc) => s + wc.cost_per_hour, 0) /
      workCenters.value.filter(wc => wc.is_active).length
    : 0
}))

// ─── Load ─────────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const res = await manufacturingAPI.getWorkCenters()
    workCenters.value = res.data || []
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to load work centers', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(load)

function setSort(field: string) {
  if (sortField.value === field) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortField.value = field; sortDir.value = 'asc' }
}

// ─── Modal helpers ────────────────────────────────────────────────────────────
function openCreate() {
  modalMode.value = 'create'
  form.value = { id: '', code: '', name: '', capacity: 8, cost_per_hour: 0, account_id: null, is_active: true }
  showModal.value = true
}

function openEdit(wc: WorkCenter) {
  modalMode.value = 'edit'
  form.value = {
    id: wc.id, code: wc.code, name: wc.name,
    capacity: wc.capacity, cost_per_hour: wc.cost_per_hour,
    account_id: wc.account_id, is_active: wc.is_active
  }
  showModal.value = true
}

// ─── Save ─────────────────────────────────────────────────────────────────────
async function save() {
  if (!form.value.code || !form.value.name) {
    app.addToast('Code and name are required', 'error')
    return
  }
  saving.value = true
  try {
    const payload = {
      code: form.value.code,
      name: form.value.name,
      capacity: form.value.capacity,
      cost_per_hour: form.value.cost_per_hour,
      account_id: form.value.account_id,
      is_active: form.value.is_active
    }
    if (modalMode.value === 'create') {
      await manufacturingAPI.createWorkCenter(payload)
      app.addToast('Work center created', 'success')
    } else {
      await manufacturingAPI.updateWorkCenter(form.value.id, payload)
      app.addToast('Work center updated', 'success')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to save work center', 'error')
  } finally {
    saving.value = false
  }
}

// ─── Toggle active ────────────────────────────────────────────────────────────
async function toggleActive(wc: WorkCenter) {
  try {
    await manufacturingAPI.updateWorkCenter(wc.id, { ...wc, is_active: !wc.is_active })
    app.addToast(`Work center ${!wc.is_active ? 'activated' : 'deactivated'}`, 'success')
    await load()
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to update', 'error')
  }
}

// ─── Delete / Deactivate ──────────────────────────────────────────────────────
async function deactivate() {
  if (!confirmDelete.value) return
  deleting.value = true
  try {
    await manufacturingAPI.deleteWorkCenter(confirmDelete.value.id)
    app.addToast('Work center deactivated', 'success')
    confirmDelete.value = null
    await load()
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to deactivate', 'error')
  } finally {
    deleting.value = false
  }
}

// ─── Helpers ──────────────────────────────────────────────────────────────────
function fmtNum(n: number, dec = 2) {
  return new Intl.NumberFormat('en-US', { minimumFractionDigits: dec, maximumFractionDigits: dec }).format(n)
}
function fmtDate(d: string) {
  return d ? new Date(d).toLocaleDateString('en-GB') : '—'
}
</script>

<template>
  <div class="flex flex-col h-full gap-4 p-4 bg-slate-50 dark:bg-slate-950 min-h-screen">

    <!-- ── Header ──────────────────────────────────────────────────────────── -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
          <Factory class="w-6 h-6 text-emerald-600" />
          Work Centers
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Manage production work centers and capacity
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="load" :disabled="loading"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300
                 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button @click="openCreate"
          class="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white
                 rounded-lg text-sm font-medium transition-colors shadow-sm">
          <Plus class="w-4 h-4" />
          New Work Center
        </button>
      </div>
    </div>

    <!-- ── KPI Cards ───────────────────────────────────────────────────────── -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total</p>
        <p class="text-2xl font-bold text-slate-900 dark:text-white mt-1">{{ kpis.total }}</p>
        <p class="text-xs text-slate-400 mt-0.5">All work centers</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Active</p>
        <p class="text-2xl font-bold text-emerald-600 mt-1">{{ kpis.active }}</p>
        <p class="text-xs text-slate-400 mt-0.5">Operational centers</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total Capacity</p>
        <p class="text-2xl font-bold text-blue-600 mt-1">{{ fmtNum(kpis.totalCapacity, 0) }}h</p>
        <p class="text-xs text-slate-400 mt-0.5">Per day (active)</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Avg Cost/Hour</p>
        <p class="text-2xl font-bold text-amber-600 mt-1">{{ fmtNum(kpis.avgCostPerHour) }}</p>
        <p class="text-xs text-slate-400 mt-0.5">Active centers avg</p>
      </div>
    </div>

    <!-- ── Table ───────────────────────────────────────────────────────────── -->
    <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 flex-1 flex flex-col overflow-hidden">
      <!-- Toolbar -->
      <div class="flex flex-wrap items-center gap-3 p-4 border-b border-slate-200 dark:border-slate-700">
        <div class="relative flex-1 min-w-[200px]">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="search" type="text" placeholder="Search code or name..."
            class="w-full pl-9 pr-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                   bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none" />
        </div>
        <select v-model="filterActive"
          class="px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none">
          <option value="">All</option>
          <option value="true">Active</option>
          <option value="false">Inactive</option>
        </select>
      </div>

      <!-- Table -->
      <div class="flex-1 overflow-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/50">
              <th v-for="col in [
                { key: 'code', label: 'Code' },
                { key: 'name', label: 'Name' },
                { key: 'capacity', label: 'Capacity (h/day)' },
                { key: 'cost_per_hour', label: 'Cost/Hour' },
                { key: 'is_active', label: 'Status' },
                { key: 'created_at', label: 'Created' },
                { key: '', label: '' }
              ]" :key="col.key"
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
              <td colspan="7" class="px-4 py-12 text-center text-slate-400">
                <RefreshCw class="w-5 h-5 animate-spin mx-auto mb-2" /> Loading...
              </td>
            </tr>
            <tr v-else-if="filtered.length === 0">
              <td colspan="7" class="px-4 py-12 text-center text-slate-400">
                <Factory class="w-8 h-8 mx-auto mb-2 opacity-30" />
                No work centers found
              </td>
            </tr>
            <tr v-for="wc in filtered" :key="wc.id"
              class="border-b border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
              :class="!wc.is_active ? 'opacity-60' : ''">
              <td class="px-4 py-3">
                <span class="font-mono font-semibold text-emerald-700 dark:text-emerald-400">{{ wc.code }}</span>
              </td>
              <td class="px-4 py-3">
                <p class="font-medium text-slate-900 dark:text-white">{{ wc.name }}</p>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1.5">
                  <Clock class="w-3.5 h-3.5 text-slate-400" />
                  <span class="font-mono text-slate-700 dark:text-slate-300">{{ fmtNum(wc.capacity, 1) }}h</span>
                </div>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1.5">
                  <DollarSign class="w-3.5 h-3.5 text-slate-400" />
                  <span class="font-mono text-slate-700 dark:text-slate-300">{{ fmtNum(wc.cost_per_hour) }}/h</span>
                </div>
              </td>
              <td class="px-4 py-3">
                <span :class="wc.is_active
                  ? 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400'
                  : 'bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400'"
                  class="px-2 py-0.5 rounded-full text-xs font-medium">
                  {{ wc.is_active ? 'Active' : 'Inactive' }}
                </span>
              </td>
              <td class="px-4 py-3 text-xs text-slate-500 dark:text-slate-400">
                {{ fmtDate(wc.created_at) }}
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1">
                  <button @click="toggleActive(wc)" :title="wc.is_active ? 'Deactivate' : 'Activate'"
                    class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
                    :class="wc.is_active ? 'text-emerald-600' : 'text-slate-400'">
                    <Activity class="w-3.5 h-3.5" />
                  </button>
                  <button @click="openEdit(wc)"
                    class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-500 hover:text-slate-700 dark:hover:text-white transition-colors">
                    <Edit2 class="w-3.5 h-3.5" />
                  </button>
                  <button @click="confirmDelete = wc"
                    class="p-1.5 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 text-slate-500 hover:text-red-600 transition-colors">
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="px-4 py-2 border-t border-slate-200 dark:border-slate-700 text-xs text-slate-500 dark:text-slate-400">
        {{ filtered.length }} of {{ workCenters.length }} work centers
      </div>
    </div>

    <!-- ── Capacity visual ─────────────────────────────────────────────────── -->
    <div v-if="workCenters.filter(wc => wc.is_active).length > 0"
      class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-5">
      <h3 class="text-sm font-semibold text-slate-900 dark:text-white mb-4 flex items-center gap-2">
        <Activity class="w-4 h-4 text-emerald-600" />
        Capacity Overview (Active Centers)
      </h3>
      <div class="space-y-3">
        <div v-for="wc in workCenters.filter(wc => wc.is_active)" :key="wc.id + 'bar'"
          class="flex items-center gap-3">
          <div class="w-36 shrink-0">
            <p class="text-sm font-medium text-slate-700 dark:text-slate-300 truncate">{{ wc.name }}</p>
            <p class="text-xs text-slate-400">{{ wc.code }}</p>
          </div>
          <div class="flex-1 bg-slate-100 dark:bg-slate-800 rounded-full h-2 overflow-hidden">
            <div class="h-full bg-emerald-500 rounded-full"
              :style="{ width: `${Math.min((wc.capacity / Math.max(...workCenters.filter(w=>w.is_active).map(w=>w.capacity), 1)) * 100, 100)}%` }" />
          </div>
          <div class="w-20 text-right">
            <span class="text-xs font-mono text-slate-700 dark:text-slate-300">{{ fmtNum(wc.capacity, 1) }}h/day</span>
          </div>
          <div class="w-24 text-right">
            <span class="text-xs font-mono text-slate-500 dark:text-slate-400">{{ fmtNum(wc.cost_per_hour) }}/h</span>
          </div>
        </div>
      </div>
    </div>

    <!-- ── Create / Edit Modal ─────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40 dark:bg-black/60" @click="showModal = false" />
        <div class="relative z-10 w-full max-w-md bg-white dark:bg-slate-900 rounded-2xl shadow-2xl">
          <div class="flex items-center justify-between p-6 border-b border-slate-200 dark:border-slate-700">
            <h2 class="text-lg font-bold text-slate-900 dark:text-white">
              {{ modalMode === 'create' ? 'New Work Center' : 'Edit Work Center' }}
            </h2>
            <button @click="showModal = false" class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500">
              <X class="w-5 h-5" />
            </button>
          </div>

          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                  Code <span class="text-red-500">*</span>
                </label>
                <input v-model="form.code" type="text" placeholder="WC-001"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none" />
              </div>
              <div class="flex items-end pb-0.5">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" v-model="form.is_active"
                    class="w-4 h-4 rounded border-slate-300 text-emerald-600 focus:ring-emerald-500" />
                  <span class="text-sm text-slate-700 dark:text-slate-300">Active</span>
                </label>
              </div>
            </div>

            <div>
              <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                Name <span class="text-red-500">*</span>
              </label>
              <input v-model="form.name" type="text" placeholder="Assembly Line 1"
                class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                       bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none" />
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                  Capacity (hours/day)
                </label>
                <input v-model.number="form.capacity" type="number" min="0" step="0.5"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                  Cost per Hour
                </label>
                <input v-model.number="form.cost_per_hour" type="number" min="0" step="0.01"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-emerald-500 outline-none" />
              </div>
            </div>

            <!-- Summary box -->
            <div class="p-3 bg-emerald-50 dark:bg-emerald-900/20 rounded-lg grid grid-cols-2 gap-3 text-sm">
              <div>
                <p class="text-xs text-emerald-700 dark:text-emerald-400">Daily Capacity Cost</p>
                <p class="font-mono font-semibold text-emerald-800 dark:text-emerald-300 mt-0.5">
                  {{ fmtNum(form.capacity * form.cost_per_hour) }}
                </p>
              </div>
              <div>
                <p class="text-xs text-emerald-700 dark:text-emerald-400">Monthly (22 days)</p>
                <p class="font-mono font-semibold text-emerald-800 dark:text-emerald-300 mt-0.5">
                  {{ fmtNum(form.capacity * form.cost_per_hour * 22) }}
                </p>
              </div>
            </div>
          </div>

          <div class="flex justify-end gap-3 p-6 border-t border-slate-200 dark:border-slate-700">
            <button @click="showModal = false"
              class="px-4 py-2 text-sm text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
              Cancel
            </button>
            <button @click="save" :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 text-sm font-medium bg-emerald-600 hover:bg-emerald-700
                     disabled:opacity-50 text-white rounded-lg transition-colors">
              <RefreshCw v-if="saving" class="w-3.5 h-3.5 animate-spin" />
              <Check v-else class="w-3.5 h-3.5" />
              {{ modalMode === 'create' ? 'Create' : 'Save Changes' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Delete Confirm ──────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="confirmDelete" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40" @click="confirmDelete = null" />
        <div class="relative z-10 w-full max-w-sm bg-white dark:bg-slate-900 rounded-xl shadow-2xl p-6">
          <div class="flex items-start gap-4 mb-4">
            <div class="p-2 bg-red-100 dark:bg-red-900/30 rounded-lg">
              <AlertCircle class="w-5 h-5 text-red-600" />
            </div>
            <div>
              <h3 class="font-semibold text-slate-900 dark:text-white">Deactivate Work Center?</h3>
              <p class="text-sm text-slate-500 mt-1">
                <strong>{{ confirmDelete.name }}</strong> will be marked inactive.
              </p>
            </div>
          </div>
          <div class="flex gap-3 justify-end">
            <button @click="confirmDelete = null"
              class="px-4 py-2 text-sm text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
              Cancel
            </button>
            <button @click="deactivate" :disabled="deleting"
              class="px-4 py-2 text-sm font-medium bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white rounded-lg transition-colors">
              Deactivate
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
