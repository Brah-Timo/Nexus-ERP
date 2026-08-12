<template>
  <div :class="dk('min-h-screen bg-gray-950 text-gray-100','min-h-screen bg-gray-50 text-gray-900')">
    <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border-b px-6 py-4'">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-lg bg-emerald-600/20">
            <Link class="w-5 h-5 text-emerald-400" />
          </div>
          <div>
            <h1 class="text-lg font-semibold">Vehicle Assignments</h1>
            <p :class="dk('text-gray-400','text-gray-500')" class="text-xs">Assign drivers to vehicles</p>
          </div>
        </div>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" />
          New Assignment
        </button>
      </div>
    </div>

    <div class="p-6 space-y-6">
      <!-- KPI Cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
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
          <input v-model="search" placeholder="Search vehicle, driver, purpose…"
            :class="dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500','bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400')+ ' border rounded-lg pl-9 pr-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-emerald-500'" />
        </div>
        <div class="flex gap-2 flex-wrap">
          <button v-for="s in statusFilters" :key="s.value"
            @click="statusFilter = s.value"
            :class="statusFilter === s.value ? 'bg-emerald-600 text-white' : dk('bg-gray-800 text-gray-300 hover:bg-gray-700','bg-gray-100 text-gray-600 hover:bg-gray-200')"
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
                <th class="px-4 py-3 text-left font-medium">Driver</th>
                <th class="px-4 py-3 text-left font-medium">Status</th>
                <th class="px-4 py-3 text-left font-medium">Start Date</th>
                <th class="px-4 py-3 text-left font-medium">End Date</th>
                <th class="px-4 py-3 text-left font-medium">Purpose</th>
                <th class="px-4 py-3 text-left font-medium">Km Start</th>
                <th class="px-4 py-3 text-left font-medium">Km End</th>
                <th class="px-4 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody :class="dk('divide-gray-800','divide-gray-100')+ ' divide-y'">
              <tr v-if="loading">
                <td colspan="10" class="py-12 text-center text-gray-400">
                  <Loader2 class="w-6 h-6 animate-spin mx-auto mb-2" />Loading…
                </td>
              </tr>
              <tr v-else-if="filtered.length === 0">
                <td colspan="10" class="py-12 text-center text-gray-400">
                  <Link class="w-8 h-8 mx-auto mb-2 opacity-30" />No assignments found
                </td>
              </tr>
              <tr v-for="a in filtered" :key="a.id" :class="dk('hover:bg-gray-800/50','hover:bg-gray-50')" class="transition-colors">
                <td class="px-4 py-3 font-mono text-xs">{{ a.assignment_number }}</td>
                <td class="px-4 py-3">
                  <div class="font-medium text-sm">{{ a.vehicle_plate || a.vehicle_id }}</div>
                </td>
                <td class="px-4 py-3">{{ a.driver_name || a.driver_id }}</td>
                <td class="px-4 py-3">
                  <span :class="statusColor(a.status)" class="px-2 py-0.5 rounded-full text-xs font-medium">{{ formatEnum(a.status) }}</span>
                </td>
                <td class="px-4 py-3 text-xs">{{ a.start_date }}</td>
                <td class="px-4 py-3 text-xs">{{ a.end_date || '—' }}</td>
                <td class="px-4 py-3 text-xs max-w-32 truncate">{{ a.purpose || '—' }}</td>
                <td class="px-4 py-3 text-xs">{{ a.start_mileage != null ? a.start_mileage.toLocaleString() : '—' }}</td>
                <td class="px-4 py-3 text-xs">{{ a.end_mileage != null ? a.end_mileage.toLocaleString() : '—' }}</td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-end gap-1">
                    <button @click="openEdit(a)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors">
                      <Pencil class="w-4 h-4 text-amber-400" />
                    </button>
                    <button @click="openDelete(a)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors">
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
            <h2 class="text-lg font-semibold">{{ editing ? 'Edit Assignment' : 'New Assignment' }}</h2>
            <button @click="showForm = false" :class="dk('hover:bg-gray-800','hover:bg-gray-100')" class="p-2 rounded-lg transition-colors"><X class="w-5 h-5" /></button>
          </div>
          <div class="p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Vehicle ID <span class="text-red-400">*</span></label>
              <input v-model="form.vehicle_id" :class="inputCls" placeholder="Vehicle UUID" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Driver ID <span class="text-red-400">*</span></label>
              <input v-model="form.driver_id" :class="inputCls" placeholder="Driver UUID" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Start Date <span class="text-red-400">*</span></label>
              <input v-model="form.start_date" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">End Date</label>
              <input v-model="form.end_date" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Status</label>
              <select v-model="form.status" :class="inputCls">
                <option value="active">Active</option>
                <option value="completed">Completed</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Purpose</label>
              <input v-model="form.purpose" :class="inputCls" placeholder="Business trip, delivery…" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Start Mileage (km)</label>
              <input v-model.number="form.start_mileage" type="number" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">End Mileage (km)</label>
              <input v-model.number="form.end_mileage" type="number" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Destination</label>
              <input v-model="form.destination" :class="inputCls" placeholder="City or location" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Approved By</label>
              <input v-model="form.approved_by" :class="inputCls" placeholder="Manager name" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium mb-1 text-gray-400">Notes</label>
              <textarea v-model="form.notes" :class="inputCls" rows="2" />
            </div>
          </div>
          <div :class="dk('border-gray-800 bg-gray-900/50','border-gray-100 bg-gray-50')+ ' flex justify-end gap-3 p-6 border-t'">
            <button @click="showForm = false" :class="dk('bg-gray-800 hover:bg-gray-700 text-gray-300','bg-gray-100 hover:bg-gray-200 text-gray-700')" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="save" :disabled="saving" class="px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              {{ editing ? 'Save Changes' : 'Create' }}
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
            <h2 class="text-lg font-semibold">Delete Assignment</h2>
          </div>
          <p :class="dk('text-gray-400','text-gray-600')" class="text-sm mb-6">Delete assignment <strong>{{ deleteItem.assignment_number }}</strong>?</p>
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
import { Link, Plus, Search, Pencil, Trash2, X, Loader2, AlertTriangle, CheckCircle2, Clock, XCircle, LayoutList } from 'lucide-vue-next'
import { fleetAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const dk = (dark: string, light: string) => app.darkMode ? dark : light

const assignments = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const search = ref('')
const statusFilter = ref('all')
const showForm = ref(false)
const showDelete = ref(false)
const editing = ref<any>(null)
const deleteItem = ref<any>(null)

const statusFilters = [
  { value: 'all', label: 'All' },
  { value: 'active', label: 'Active' },
  { value: 'completed', label: 'Completed' },
  { value: 'cancelled', label: 'Cancelled' },
]

const defaultForm = () => ({
  vehicle_id: '', driver_id: '', start_date: '', end_date: '', status: 'active',
  purpose: '', start_mileage: null, end_mileage: null, destination: '', approved_by: '', notes: '',
})
const form = ref(defaultForm())

const inputCls = computed(() =>
  dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500', 'bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400') +
  ' border rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-emerald-500')

const filtered = computed(() => {
  let list = assignments.value
  if (statusFilter.value !== 'all') list = list.filter(a => a.status === statusFilter.value)
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(a => [a.vehicle_plate, a.driver_name, a.purpose, a.assignment_number].some(f => f?.toLowerCase().includes(q)))
  }
  return list
})

const kpis = computed(() => [
  { label: 'Total', value: assignments.value.length, icon: LayoutList, color: 'text-emerald-400' },
  { label: 'Active', value: assignments.value.filter(a => a.status === 'active').length, icon: CheckCircle2, color: 'text-green-400' },
  { label: 'Completed', value: assignments.value.filter(a => a.status === 'completed').length, icon: Clock, color: 'text-blue-400' },
  { label: 'Cancelled', value: assignments.value.filter(a => a.status === 'cancelled').length, icon: XCircle, color: 'text-red-400' },
])

function formatEnum(s: string) { return s ? s.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()) : '—' }
function statusColor(s: string) {
  const m: Record<string, string> = {
    active: 'bg-green-500/15 text-green-400',
    completed: 'bg-blue-500/15 text-blue-400',
    cancelled: 'bg-red-500/15 text-red-400',
  }
  return m[s] || 'bg-gray-500/15 text-gray-400'
}

async function load() {
  loading.value = true
  try {
    const r = await fleetAPI.listAssignments()
    assignments.value = r.data.items || r.data || []
  } catch { app.addToast('Failed to load assignments', 'error') }
  finally { loading.value = false }
}

function openCreate() { editing.value = null; form.value = defaultForm(); showForm.value = true }
function openEdit(a: any) { editing.value = a; form.value = { ...defaultForm(), ...a }; showForm.value = true }
function openDelete(a: any) { deleteItem.value = a; showDelete.value = true }

async function save() {
  if (!form.value.vehicle_id || !form.value.driver_id || !form.value.start_date) {
    app.addToast('Vehicle, driver, and start date are required', 'error'); return
  }
  saving.value = true
  try {
    if (editing.value) {
      await fleetAPI.updateAssignment(editing.value.id, form.value)
      app.addToast('Assignment updated', 'success')
    } else {
      await fleetAPI.createAssignment(form.value)
      app.addToast('Assignment created', 'success')
    }
    showForm.value = false; await load()
  } catch (e: any) { app.addToast(e?.response?.data?.error || 'Save failed', 'error') }
  finally { saving.value = false }
}

async function confirmDelete() {
  if (!deleteItem.value) return
  saving.value = true
  try {
    await fleetAPI.deleteAssignment(deleteItem.value.id)
    app.addToast('Assignment deleted', 'success')
    showDelete.value = false; await load()
  } catch { app.addToast('Delete failed', 'error') }
  finally { saving.value = false }
}

onMounted(load)
</script>
