<template>
  <div :class="dk('min-h-screen bg-gray-950 text-gray-100','min-h-screen bg-gray-50 text-gray-900')">
    <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border-b px-6 py-4'">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-lg bg-orange-600/20">
            <Wrench class="w-5 h-5 text-orange-400" />
          </div>
          <div>
            <h1 class="text-lg font-semibold">Fleet Maintenance</h1>
            <p :class="dk('text-gray-400','text-gray-500')" class="text-xs">Vehicle service and repair records</p>
          </div>
        </div>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 bg-orange-600 hover:bg-orange-700 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" />
          Add Service Record
        </button>
      </div>
    </div>

    <div class="p-6 space-y-6">
      <!-- KPI Cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-5 gap-4">
        <div v-for="kpi in kpis" :key="kpi.label" :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-4'">
          <div class="flex items-center justify-between mb-2">
            <span :class="dk('text-gray-400','text-gray-500')" class="text-xs">{{ kpi.label }}</span>
            <component :is="kpi.icon" :class="'w-4 h-4 ' + kpi.color" />
          </div>
          <div class="text-2xl font-bold">{{ kpi.value }}</div>
        </div>
      </div>

      <!-- Filters -->
      <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-4 flex flex-wrap gap-3 items-center'">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input v-model="search" placeholder="Search vehicle, service type, technician…"
            :class="dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500','bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400')+ ' border rounded-lg pl-9 pr-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-orange-500'" />
        </div>
        <div class="flex gap-2 flex-wrap">
          <button v-for="s in statusFilters" :key="s.value"
            @click="statusFilter = s.value"
            :class="statusFilter === s.value ? 'bg-orange-600 text-white' : dk('bg-gray-800 text-gray-300 hover:bg-gray-700','bg-gray-100 text-gray-600 hover:bg-gray-200')"
            class="px-3 py-1.5 rounded-lg text-xs font-medium transition-colors">
            {{ s.label }}
          </button>
        </div>
      </div>

      <!-- Table -->
      <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl overflow-hidden'">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead :class="dk('bg-gray-800/50 text-gray-400','bg-gray-50 text-gray-500')">
              <tr>
                <th class="px-4 py-3 text-left font-medium">Reference</th>
                <th class="px-4 py-3 text-left font-medium">Vehicle</th>
                <th class="px-4 py-3 text-left font-medium">Service Type</th>
                <th class="px-4 py-3 text-left font-medium">Status</th>
                <th class="px-4 py-3 text-left font-medium">Scheduled</th>
                <th class="px-4 py-3 text-left font-medium">Completed</th>
                <th class="px-4 py-3 text-left font-medium">Technician</th>
                <th class="px-4 py-3 text-right font-medium">Cost</th>
                <th class="px-4 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody :class="dk('divide-gray-800','divide-gray-100')+ ' divide-y'">
              <tr v-if="loading">
                <td colspan="9" class="py-12 text-center text-gray-400">
                  <Loader2 class="w-6 h-6 animate-spin mx-auto mb-2" />Loading…
                </td>
              </tr>
              <tr v-else-if="filtered.length === 0">
                <td colspan="9" class="py-12 text-center text-gray-400">
                  <Wrench class="w-8 h-8 mx-auto mb-2 opacity-30" />No service records found
                </td>
              </tr>
              <tr v-for="m in filtered" :key="m.id"
                :class="[dk('hover:bg-gray-800/50','hover:bg-gray-50'), isOverdue(m) ? 'border-l-2 border-l-red-500' : '']"
                class="transition-colors">
                <td class="px-4 py-3 font-mono text-xs">{{ m.service_number }}</td>
                <td class="px-4 py-3 font-medium text-sm">{{ m.vehicle_plate || m.vehicle_id }}</td>
                <td class="px-4 py-3">
                  <span :class="serviceTypeColor(m.service_type)" class="px-2 py-0.5 rounded text-xs">{{ formatEnum(m.service_type) }}</span>
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-1.5">
                    <span v-if="isOverdue(m)" class="text-red-400 text-xs font-medium">Overdue</span>
                    <span v-else :class="statusColor(m.status)" class="px-2 py-0.5 rounded-full text-xs font-medium">{{ formatEnum(m.status) }}</span>
                  </div>
                </td>
                <td class="px-4 py-3 text-xs">{{ m.scheduled_date || '—' }}</td>
                <td class="px-4 py-3 text-xs">{{ m.completed_date || '—' }}</td>
                <td class="px-4 py-3 text-xs">{{ m.technician_name || '—' }}</td>
                <td class="px-4 py-3 text-right font-medium text-orange-400">{{ m.total_cost != null ? fmtDZD(m.total_cost) : '—' }}</td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-end gap-1">
                    <button v-if="m.status !== 'completed'" @click="openComplete(m)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors" title="Mark Complete">
                      <CheckCircle2 class="w-4 h-4 text-green-400" />
                    </button>
                    <button @click="openEdit(m)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors">
                      <Pencil class="w-4 h-4 text-amber-400" />
                    </button>
                    <button @click="openDelete(m)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors">
                      <Trash2 class="w-4 h-4 text-red-400" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showForm" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showForm = false" />
        <div :class="dk('bg-gray-900 border-gray-700','bg-white border-gray-200')+ ' relative border rounded-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto shadow-2xl'">
          <div :class="dk('border-gray-800','border-gray-100')+ ' flex items-center justify-between p-6 border-b'">
            <h2 class="text-lg font-semibold">{{ editing ? 'Edit Service Record' : 'Add Service Record' }}</h2>
            <button @click="showForm = false" :class="dk('hover:bg-gray-800','hover:bg-gray-100')" class="p-2 rounded-lg transition-colors"><X class="w-5 h-5" /></button>
          </div>
          <div class="p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Vehicle ID <span class="text-red-400">*</span></label>
              <input v-model="form.vehicle_id" :class="inputCls" placeholder="Vehicle UUID" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Service Type <span class="text-red-400">*</span></label>
              <select v-model="form.service_type" :class="inputCls">
                <option value="preventive">Preventive</option>
                <option value="corrective">Corrective</option>
                <option value="inspection">Inspection</option>
                <option value="tire_change">Tire Change</option>
                <option value="oil_change">Oil Change</option>
                <option value="brake_service">Brake Service</option>
                <option value="other">Other</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Title <span class="text-red-400">*</span></label>
              <input v-model="form.title" :class="inputCls" placeholder="Service description" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Status</label>
              <select v-model="form.status" :class="inputCls">
                <option value="scheduled">Scheduled</option>
                <option value="in_progress">In Progress</option>
                <option value="completed">Completed</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Scheduled Date</label>
              <input v-model="form.scheduled_date" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Completed Date</label>
              <input v-model="form.completed_date" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Mileage at Service</label>
              <input v-model.number="form.mileage_at_service" type="number" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Technician</label>
              <input v-model="form.technician_name" :class="inputCls" placeholder="Technician name" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Labor Cost (DZD)</label>
              <input v-model.number="form.labor_cost" type="number" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Parts Cost (DZD)</label>
              <input v-model.number="form.parts_cost" type="number" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Next Service Date</label>
              <input v-model="form.next_service_date" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Next Service Mileage</label>
              <input v-model.number="form.next_service_mileage" type="number" :class="inputCls" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium mb-1 text-gray-400">Notes</label>
              <textarea v-model="form.notes" :class="inputCls" rows="2" />
            </div>
          </div>
          <div :class="dk('border-gray-800 bg-gray-900/50','border-gray-100 bg-gray-50')+ ' flex justify-end gap-3 p-6 border-t'">
            <button @click="showForm = false" :class="dk('bg-gray-800 hover:bg-gray-700 text-gray-300','bg-gray-100 hover:bg-gray-200 text-gray-700')" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="save" :disabled="saving" class="px-4 py-2 bg-orange-600 hover:bg-orange-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              {{ editing ? 'Save Changes' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Complete Modal -->
    <Teleport to="body">
      <div v-if="showComplete && completeItem" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showComplete = false" />
        <div :class="dk('bg-gray-900 border-gray-700','bg-white border-gray-200')+ ' relative border rounded-2xl w-full max-w-md shadow-2xl p-6'">
          <h2 class="text-lg font-semibold mb-4 flex items-center gap-2"><CheckCircle2 class="w-5 h-5 text-green-400" /> Complete Service</h2>
          <p :class="dk('text-gray-400','text-gray-600')" class="text-sm mb-4">{{ completeItem.title }}</p>
          <div class="space-y-3">
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Completed Date</label>
              <input v-model="completeForm.completed_date" type="date" :class="inputCls" />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-medium mb-1 text-gray-400">Labor Cost</label>
                <input v-model.number="completeForm.labor_cost" type="number" :class="inputCls" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1 text-gray-400">Parts Cost</label>
                <input v-model.number="completeForm.parts_cost" type="number" :class="inputCls" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Work Performed</label>
              <textarea v-model="completeForm.notes" :class="inputCls" rows="2" />
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showComplete = false" :class="dk('bg-gray-800 hover:bg-gray-700 text-gray-300','bg-gray-100 hover:bg-gray-200 text-gray-700')" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="confirmComplete" :disabled="saving" class="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />Mark Complete
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Delete Modal -->
    <Teleport to="body">
      <div v-if="showDelete && deleteItem" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showDelete = false" />
        <div :class="dk('bg-gray-900 border-gray-700','bg-white border-gray-200')+ ' relative border rounded-2xl w-full max-w-md shadow-2xl p-6'">
          <div class="flex items-center gap-3 mb-4">
            <div class="p-2 rounded-full bg-red-500/20"><AlertTriangle class="w-5 h-5 text-red-400" /></div>
            <h2 class="text-lg font-semibold">Delete Service Record</h2>
          </div>
          <p :class="dk('text-gray-400','text-gray-600')" class="text-sm mb-6">Delete <strong>{{ deleteItem.title }}</strong>?</p>
          <div class="flex justify-end gap-3">
            <button @click="showDelete = false" :class="dk('bg-gray-800 hover:bg-gray-700 text-gray-300','bg-gray-100 hover:bg-gray-200 text-gray-700')" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="confirmDelete" :disabled="saving" class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />Delete
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Wrench, Plus, Search, Pencil, Trash2, X, Loader2, AlertTriangle, CheckCircle2, Clock, DollarSign, AlertCircle, LayoutList } from 'lucide-vue-next'
import { fleetAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const dk = (dark: string, light: string) => app.darkMode ? dark : light

const records = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const search = ref('')
const statusFilter = ref('all')
const showForm = ref(false)
const showDelete = ref(false)
const showComplete = ref(false)
const editing = ref<any>(null)
const deleteItem = ref<any>(null)
const completeItem = ref<any>(null)

const statusFilters = [
  { value: 'all', label: 'All' },
  { value: 'scheduled', label: 'Scheduled' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'completed', label: 'Completed' },
]

const defaultForm = () => ({
  vehicle_id: '', service_type: 'preventive', title: '', status: 'scheduled',
  scheduled_date: '', completed_date: '', mileage_at_service: null,
  technician_name: '', labor_cost: null, parts_cost: null,
  next_service_date: '', next_service_mileage: null, notes: '',
})
const form = ref(defaultForm())
const completeForm = ref({ completed_date: new Date().toISOString().split('T')[0], labor_cost: 0, parts_cost: 0, notes: '' })

const inputCls = computed(() =>
  dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500', 'bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400') +
  ' border rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-orange-500')

const filtered = computed(() => {
  let list = records.value
  if (statusFilter.value !== 'all') list = list.filter(m => m.status === statusFilter.value)
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(m => [m.vehicle_plate, m.title, m.technician_name, m.service_type].some(f => f?.toLowerCase().includes(q)))
  }
  return list
})

const kpis = computed(() => {
  const totalCost = records.value.reduce((s, m) => s + (m.total_cost || 0), 0)
  return [
    { label: 'Total Records', value: records.value.length, icon: LayoutList, color: 'text-orange-400' },
    { label: 'Scheduled', value: records.value.filter(m => m.status === 'scheduled').length, icon: Clock, color: 'text-blue-400' },
    { label: 'In Progress', value: records.value.filter(m => m.status === 'in_progress').length, icon: Wrench, color: 'text-amber-400' },
    { label: 'Overdue', value: records.value.filter(isOverdue).length, icon: AlertCircle, color: 'text-red-400' },
    { label: 'Total Cost', value: fmtDZD(totalCost), icon: DollarSign, color: 'text-green-400' },
  ]
})

function formatEnum(s: string) { return s ? s.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()) : '—' }
function fmtDZD(n: number) { return new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n || 0) + ' DZD' }
function isOverdue(m: any) {
  return m.status !== 'completed' && m.status !== 'cancelled' && m.scheduled_date && new Date(m.scheduled_date) < new Date()
}
function statusColor(s: string) {
  const m: Record<string, string> = {
    scheduled: 'bg-blue-500/15 text-blue-400',
    in_progress: 'bg-amber-500/15 text-amber-400',
    completed: 'bg-green-500/15 text-green-400',
    cancelled: 'bg-red-500/15 text-red-400',
  }
  return m[s] || 'bg-gray-500/15 text-gray-400'
}
function serviceTypeColor(t: string) {
  const m: Record<string, string> = {
    preventive: 'bg-blue-500/15 text-blue-400',
    corrective: 'bg-red-500/15 text-red-400',
    inspection: 'bg-purple-500/15 text-purple-400',
    oil_change: 'bg-yellow-500/15 text-yellow-400',
    tire_change: 'bg-green-500/15 text-green-400',
  }
  return m[t] || 'bg-gray-500/15 text-gray-400'
}

async function load() {
  loading.value = true
  try {
    const r = await fleetAPI.listFleetMaintenance()
    records.value = r.data.items || r.data || []
  } catch { app.addToast('Failed to load maintenance records', 'error') }
  finally { loading.value = false }
}

function openCreate() { editing.value = null; form.value = defaultForm(); showForm.value = true }
function openEdit(m: any) { editing.value = m; form.value = { ...defaultForm(), ...m }; showForm.value = true }
function openDelete(m: any) { deleteItem.value = m; showDelete.value = true }
function openComplete(m: any) {
  completeItem.value = m
  completeForm.value = { completed_date: new Date().toISOString().split('T')[0], labor_cost: 0, parts_cost: 0, notes: '' }
  showComplete.value = true
}

async function save() {
  if (!form.value.vehicle_id || !form.value.title) { app.addToast('Vehicle and title are required', 'error'); return }
  saving.value = true
  try {
    if (editing.value) {
      await fleetAPI.updateFleetMaintenance(editing.value.id, form.value)
      app.addToast('Service record updated', 'success')
    } else {
      await fleetAPI.createFleetMaintenance(form.value)
      app.addToast('Service record created', 'success')
    }
    showForm.value = false; await load()
  } catch (e: any) { app.addToast(e?.response?.data?.error || 'Save failed', 'error') }
  finally { saving.value = false }
}

async function confirmComplete() {
  if (!completeItem.value) return
  saving.value = true
  try {
    await fleetAPI.updateFleetMaintenance(completeItem.value.id, { ...completeItem.value, status: 'completed', ...completeForm.value })
    app.addToast('Service completed', 'success')
    showComplete.value = false; await load()
  } catch { app.addToast('Update failed', 'error') }
  finally { saving.value = false }
}

async function confirmDelete() {
  if (!deleteItem.value) return
  saving.value = true
  try {
    await fleetAPI.deleteFleetMaintenance(deleteItem.value.id)
    app.addToast('Record deleted', 'success')
    showDelete.value = false; await load()
  } catch { app.addToast('Delete failed', 'error') }
  finally { saving.value = false }
}

onMounted(load)
</script>
