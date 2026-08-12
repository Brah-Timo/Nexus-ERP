<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { assetsAPI } from '@/api/client'
import {
  Wrench, Plus, Edit, Trash2, RefreshCw, XCircle,
  CheckCircle, Clock, AlertCircle, Calendar
} from 'lucide-vue-next'

const app = useAppStore()
const records = ref<any[]>([])
const assets = ref<any[]>([])
const loading = ref(false)
const filterStatus = ref('')
const filterType = ref('')
const showModal = ref(false)
const showCompleteModal = ref(false)
const editId = ref('')
const completeId = ref('')
const form = ref<Record<string, any>>({})
const completeForm = ref({ completed_date: '', cost: 0, notes: '', next_maintenance_date: '' })

const maintenanceTypes = ['preventive', 'corrective', 'inspection', 'upgrade', 'repair', 'calibration']
const statusOptions = [
  { value: '', label: 'All' },
  { value: 'scheduled', label: 'Scheduled' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'completed', label: 'Completed' },
  { value: 'cancelled', label: 'Cancelled' },
  { value: 'overdue', label: 'Overdue' },
]

async function load() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (filterStatus.value) params.status = filterStatus.value
    if (filterType.value) params.type = filterType.value
    const [mRes, aRes] = await Promise.all([
      assetsAPI.listMaintenance(params),
      assetsAPI.listAssets(),
    ])
    records.value = mRes.data
    assets.value = aRes.data
  } finally {
    loading.value = false
  }
}

onMounted(load)

function openCreate() {
  editId.value = ''
  form.value = {
    asset_id: '', maintenance_type: 'preventive', status: 'scheduled',
    scheduled_date: new Date().toISOString().slice(0, 10),
    completed_date: '', description: '', technician: '', vendor: '',
    cost: 0, parts_used: '', notes: '', next_maintenance_date: '',
  }
  showModal.value = true
}

function openEdit(r: any) {
  editId.value = r.id
  form.value = { ...r }
  showModal.value = true
}

function openComplete(r: any) {
  completeId.value = r.id
  completeForm.value = {
    completed_date: new Date().toISOString().slice(0, 10),
    cost: r.cost ?? 0,
    notes: '',
    next_maintenance_date: '',
  }
  showCompleteModal.value = true
}

async function save() {
  if (editId.value) {
    await assetsAPI.updateMaintenance(editId.value, form.value)
  } else {
    await assetsAPI.createMaintenance(form.value)
  }
  showModal.value = false
  load()
}

async function complete() {
  await assetsAPI.completeMaintenance(completeId.value, completeForm.value)
  showCompleteModal.value = false
  load()
}

async function remove(id: string) {
  if (!confirm('Delete this maintenance record?')) return
  await assetsAPI.deleteMaintenance(id)
  load()
}

function statusBadge(s: string) {
  const m: Record<string, string> = {
    scheduled: 'bg-sky-500/20 text-sky-300',
    in_progress: 'bg-amber-500/20 text-amber-300',
    completed: 'bg-emerald-500/20 text-emerald-300',
    cancelled: 'bg-red-500/20 text-red-300',
    overdue: 'bg-red-700/30 text-red-400',
  }
  return m[s] ?? 'bg-slate-500/20 text-slate-300'
}

function typeBadge(t: string) {
  const m: Record<string, string> = {
    preventive: 'bg-indigo-500/20 text-indigo-300',
    corrective: 'bg-orange-500/20 text-orange-300',
    inspection: 'bg-blue-500/20 text-blue-300',
    upgrade: 'bg-purple-500/20 text-purple-300',
    repair: 'bg-red-500/20 text-red-300',
    calibration: 'bg-teal-500/20 text-teal-300',
  }
  return m[t] ?? 'bg-slate-500/20 text-slate-300'
}

function fmtDate(d: any) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString()
}
function fmtCurrency(v: number) {
  return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0)
}

const cardCls = computed(() =>
  app.darkMode ? 'bg-slate-800/60 border-slate-700' : 'bg-white border-slate-200 shadow-sm'
)
const inputCls = computed(() =>
  app.darkMode
    ? 'bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400 focus:border-indigo-500'
    : 'bg-white border-slate-300 text-slate-900 placeholder-slate-400 focus:border-indigo-500'
)
const thCls = computed(() =>
  app.darkMode ? 'text-slate-400 border-slate-700' : 'text-slate-500 border-slate-200'
)
const tdCls = computed(() =>
  app.darkMode ? 'text-slate-300 border-slate-700' : 'text-slate-700 border-slate-200'
)

const totalCost = computed(() => records.value.reduce((s, r) => s + (r.cost || 0), 0))
</script>

<template>
  <div class="p-6 space-y-5">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Asset Maintenance</h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Track maintenance records and schedule for all assets</p>
      </div>
      <div class="flex gap-2">
        <button @click="load" class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors"
          :class="app.darkMode ? 'border-slate-600 text-slate-300 hover:bg-slate-700' : 'border-slate-300 text-slate-600 hover:bg-slate-50'">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white">
          <Plus class="w-4 h-4" /> New Record
        </button>
      </div>
    </div>

    <!-- KPIs -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div v-for="stat in [
        { label: 'Total Records', value: records.length, color: 'text-indigo-400' },
        { label: 'Scheduled', value: records.filter(r=>r.status==='scheduled').length, color: 'text-sky-400' },
        { label: 'In Progress', value: records.filter(r=>r.status==='in_progress').length, color: 'text-amber-400' },
        { label: 'Total Cost', value: fmtCurrency(totalCost), color: 'text-emerald-400' },
      ]" :key="stat.label" class="rounded-xl border p-4" :class="cardCls">
        <div class="text-xs mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ stat.label }}</div>
        <div class="text-xl font-bold" :class="stat.color">{{ stat.value }}</div>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex gap-3 flex-wrap">
      <select v-model="filterStatus" @change="load" class="px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls">
        <option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
      </select>
      <select v-model="filterType" @change="load" class="px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls">
        <option value="">All Types</option>
        <option v-for="t in maintenanceTypes" :key="t" :value="t">{{ t }}</option>
      </select>
    </div>

    <!-- Table -->
    <div class="rounded-xl border overflow-hidden" :class="cardCls">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b" :class="thCls">
              <th class="text-left px-4 py-3 font-medium">Asset</th>
              <th class="text-left px-4 py-3 font-medium">Type</th>
              <th class="text-left px-4 py-3 font-medium">Status</th>
              <th class="text-left px-4 py-3 font-medium">Scheduled</th>
              <th class="text-left px-4 py-3 font-medium">Completed</th>
              <th class="text-left px-4 py-3 font-medium">Technician</th>
              <th class="text-right px-4 py-3 font-medium">Cost</th>
              <th class="text-left px-4 py-3 font-medium">Next Service</th>
              <th class="text-right px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading"><td colspan="9" class="text-center py-10 text-slate-400">Loading...</td></tr>
            <tr v-else-if="records.length === 0"><td colspan="9" class="text-center py-10 text-slate-400">No records found</td></tr>
            <tr v-for="r in records" :key="r.id"
              class="border-t transition-colors"
              :class="[tdCls, app.darkMode ? 'hover:bg-slate-700/30' : 'hover:bg-slate-50']">
              <td class="px-4 py-3">
                <div class="font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">{{ r.asset_name }}</div>
                <div class="text-xs text-slate-400">{{ r.asset_number }}</div>
              </td>
              <td class="px-4 py-3">
                <span class="px-2 py-0.5 rounded-full text-xs font-medium capitalize" :class="typeBadge(r.maintenance_type)">
                  {{ r.maintenance_type?.replace(/_/g,' ') }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="statusBadge(r.status)">
                  {{ r.status?.replace(/_/g,' ') }}
                </span>
              </td>
              <td class="px-4 py-3 text-xs">{{ fmtDate(r.scheduled_date) }}</td>
              <td class="px-4 py-3 text-xs">{{ fmtDate(r.completed_date) }}</td>
              <td class="px-4 py-3 text-xs">{{ r.technician || '—' }}</td>
              <td class="px-4 py-3 text-right font-mono text-sm">{{ fmtCurrency(r.cost) }}</td>
              <td class="px-4 py-3 text-xs text-amber-400">{{ fmtDate(r.next_maintenance_date) }}</td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <button v-if="r.status !== 'completed' && r.status !== 'cancelled'"
                    @click="openComplete(r)"
                    class="p-1.5 rounded hover:bg-emerald-500/20 text-emerald-400 transition-colors" title="Complete">
                    <CheckCircle class="w-4 h-4" />
                  </button>
                  <button @click="openEdit(r)" class="p-1.5 rounded hover:bg-amber-500/20 text-amber-400 transition-colors" title="Edit">
                    <Edit class="w-4 h-4" />
                  </button>
                  <button @click="remove(r.id)" class="p-1.5 rounded hover:bg-red-500/20 text-red-400 transition-colors" title="Delete">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60">
      <div class="w-full max-w-2xl rounded-2xl border max-h-[90vh] overflow-y-auto"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'">
        <div class="flex items-center justify-between p-5 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
            {{ editId ? 'Edit Record' : 'New Maintenance Record' }}
          </h2>
          <button @click="showModal = false" class="text-slate-400 hover:text-slate-200"><XCircle class="w-5 h-5" /></button>
        </div>
        <div class="p-5 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="col-span-2">
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Asset *</label>
              <select v-model="form.asset_id" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option value="">Select asset...</option>
                <option v-for="a in assets" :key="a.id" :value="a.id">{{ a.asset_number }} — {{ a.name }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Type</label>
              <select v-model="form.maintenance_type" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option v-for="t in maintenanceTypes" :key="t" :value="t">{{ t }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Status</label>
              <select v-model="form.status" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option v-for="s in statusOptions.slice(1)" :key="s.value" :value="s.value">{{ s.label }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Scheduled Date</label>
              <input v-model="form.scheduled_date" type="date" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Completed Date</label>
              <input v-model="form.completed_date" type="date" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Technician</label>
              <input v-model="form.technician" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Vendor</label>
              <input v-model="form.vendor" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Cost</label>
              <input v-model.number="form.cost" type="number" step="0.01" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Next Maintenance Date</label>
              <input v-model="form.next_maintenance_date" type="date" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
            </div>
            <div class="col-span-2">
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Description</label>
              <textarea v-model="form.description" rows="2" class="w-full px-3 py-2 rounded-lg border text-sm outline-none resize-none" :class="inputCls" />
            </div>
            <div class="col-span-2">
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Parts Used</label>
              <textarea v-model="form.parts_used" rows="2" class="w-full px-3 py-2 rounded-lg border text-sm outline-none resize-none" :class="inputCls" />
            </div>
          </div>
        </div>
        <div class="flex justify-end gap-3 p-5 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showModal = false" class="px-4 py-2 rounded-lg text-sm border" :class="app.darkMode ? 'border-slate-600 text-slate-300' : 'border-slate-300 text-slate-600'">Cancel</button>
          <button @click="save" class="px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white">
            {{ editId ? 'Save' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Complete Modal -->
    <div v-if="showCompleteModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60">
      <div class="w-full max-w-md rounded-2xl border"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'">
        <div class="flex items-center justify-between p-5 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Complete Maintenance</h2>
          <button @click="showCompleteModal = false" class="text-slate-400 hover:text-slate-200"><XCircle class="w-5 h-5" /></button>
        </div>
        <div class="p-5 space-y-4">
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Completion Date</label>
            <input v-model="completeForm.completed_date" type="date" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
          </div>
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Actual Cost</label>
            <input v-model.number="completeForm.cost" type="number" step="0.01" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
          </div>
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Next Maintenance Date</label>
            <input v-model="completeForm.next_maintenance_date" type="date" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
          </div>
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Notes</label>
            <textarea v-model="completeForm.notes" rows="3" class="w-full px-3 py-2 rounded-lg border text-sm outline-none resize-none" :class="inputCls" />
          </div>
        </div>
        <div class="flex justify-end gap-3 p-5 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showCompleteModal = false" class="px-4 py-2 rounded-lg text-sm border" :class="app.darkMode ? 'border-slate-600 text-slate-300' : 'border-slate-300 text-slate-600'">Cancel</button>
          <button @click="complete" class="px-4 py-2 rounded-lg text-sm font-medium bg-emerald-600 hover:bg-emerald-700 text-white">Mark Complete</button>
        </div>
      </div>
    </div>
  </div>
</template>
