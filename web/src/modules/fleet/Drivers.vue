<template>
  <div :class="dk('min-h-screen bg-gray-950 text-gray-100','min-h-screen bg-gray-50 text-gray-900')">
    <!-- Header -->
    <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border-b px-6 py-4'">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-lg bg-indigo-600/20">
            <UserCheck class="w-5 h-5 text-indigo-400" />
          </div>
          <div>
            <h1 class="text-lg font-semibold">Drivers</h1>
            <p :class="dk('text-gray-400','text-gray-500')" class="text-xs">Fleet driver management</p>
          </div>
        </div>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" />
          Add Driver
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
          <div v-if="kpi.sub" :class="'text-xs mt-1 ' + kpi.subColor">{{ kpi.sub }}</div>
        </div>
      </div>

      <!-- Filters -->
      <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-4'">
        <div class="flex flex-wrap gap-3 items-center">
          <div class="relative flex-1 min-w-48">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input v-model="search" placeholder="Search name, license, phone…"
              :class="dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500','bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400')+ ' border rounded-lg pl-9 pr-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500'" />
          </div>
          <div class="flex gap-2 flex-wrap">
            <button v-for="s in statusFilters" :key="s.value"
              @click="statusFilter = s.value"
              :class="statusFilter === s.value ? 'bg-indigo-600 text-white' : dk('bg-gray-800 text-gray-300 hover:bg-gray-700','bg-gray-100 text-gray-600 hover:bg-gray-200')"
              class="px-3 py-1.5 rounded-lg text-xs font-medium transition-colors">
              {{ s.label }}
            </button>
          </div>
        </div>
      </div>

      <!-- Table -->
      <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl overflow-hidden'">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead :class="dk('bg-gray-800/50 text-gray-400','bg-gray-50 text-gray-500')">
              <tr>
                <th class="px-4 py-3 text-left font-medium">Driver</th>
                <th class="px-4 py-3 text-left font-medium">License</th>
                <th class="px-4 py-3 text-left font-medium">Class</th>
                <th class="px-4 py-3 text-left font-medium">Status</th>
                <th class="px-4 py-3 text-left font-medium">Phone</th>
                <th class="px-4 py-3 text-left font-medium">License Expiry</th>
                <th class="px-4 py-3 text-left font-medium">Medical Expiry</th>
                <th class="px-4 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody :class="dk('divide-gray-800','divide-gray-100')+ ' divide-y'">
              <tr v-if="loading">
                <td colspan="8" class="py-12 text-center text-gray-400">
                  <Loader2 class="w-6 h-6 animate-spin mx-auto mb-2" />Loading drivers…
                </td>
              </tr>
              <tr v-else-if="filtered.length === 0">
                <td colspan="8" class="py-12 text-center text-gray-400">
                  <UserCheck class="w-8 h-8 mx-auto mb-2 opacity-30" />No drivers found
                </td>
              </tr>
              <tr v-for="d in filtered" :key="d.id"
                :class="[dk('hover:bg-gray-800/50','hover:bg-gray-50'), isExpiringSoon(d.license_expiry) ? 'border-l-2 border-l-amber-500' : '']"
                class="transition-colors">
                <td class="px-4 py-3">
                  <div class="font-medium">{{ d.full_name }}</div>
                  <div :class="dk('text-gray-400','text-gray-500')" class="text-xs">{{ d.employee_id || d.code }}</div>
                </td>
                <td class="px-4 py-3 font-mono text-xs">{{ d.license_number }}</td>
                <td class="px-4 py-3">
                  <span :class="dk('bg-purple-500/15 text-purple-400','bg-purple-100 text-purple-700')" class="px-2 py-0.5 rounded text-xs">{{ d.license_class }}</span>
                </td>
                <td class="px-4 py-3">
                  <span :class="statusColor(d.status)" class="px-2 py-0.5 rounded-full text-xs font-medium">{{ formatEnum(d.status) }}</span>
                </td>
                <td class="px-4 py-3 text-xs">{{ d.phone || '—' }}</td>
                <td class="px-4 py-3">
                  <span :class="expiryClass(d.license_expiry)" class="text-xs">{{ d.license_expiry || '—' }}</span>
                </td>
                <td class="px-4 py-3">
                  <span :class="expiryClass(d.medical_expiry)" class="text-xs">{{ d.medical_expiry || '—' }}</span>
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-end gap-1">
                    <button @click="openEdit(d)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors">
                      <Pencil class="w-4 h-4 text-amber-400" />
                    </button>
                    <button @click="openDelete(d)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors">
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
            <h2 class="text-lg font-semibold">{{ editing ? 'Edit Driver' : 'Add Driver' }}</h2>
            <button @click="showForm = false" :class="dk('hover:bg-gray-800','hover:bg-gray-100')" class="p-2 rounded-lg transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Full Name <span class="text-red-400">*</span></label>
              <input v-model="form.full_name" :class="inputCls" placeholder="Ahmed Ben Ali" required />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Employee ID</label>
              <input v-model="form.employee_id" :class="inputCls" placeholder="EMP-001" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">License Number <span class="text-red-400">*</span></label>
              <input v-model="form.license_number" :class="inputCls" placeholder="DZ-123456" required />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">License Class</label>
              <select v-model="form.license_class" :class="inputCls">
                <option value="B">B — Car</option>
                <option value="C">C — Truck</option>
                <option value="D">D — Bus</option>
                <option value="E">E — Heavy</option>
                <option value="A">A — Motorcycle</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">License Expiry</label>
              <input v-model="form.license_expiry" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Status</label>
              <select v-model="form.status" :class="inputCls">
                <option value="available">Available</option>
                <option value="on_duty">On Duty</option>
                <option value="off_duty">Off Duty</option>
                <option value="on_leave">On Leave</option>
                <option value="terminated">Terminated</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Phone</label>
              <input v-model="form.phone" :class="inputCls" placeholder="+213 5XX XXX XXX" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Email</label>
              <input v-model="form.email" type="email" :class="inputCls" placeholder="driver@example.com" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Date of Birth</label>
              <input v-model="form.date_of_birth" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">National ID</label>
              <input v-model="form.national_id" :class="inputCls" placeholder="National ID number" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Medical Expiry</label>
              <input v-model="form.medical_expiry" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Hire Date</label>
              <input v-model="form.hire_date" type="date" :class="inputCls" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium mb-1 text-gray-400">Address</label>
              <input v-model="form.address" :class="inputCls" placeholder="Street address" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium mb-1 text-gray-400">Notes</label>
              <textarea v-model="form.notes" :class="inputCls" rows="2" />
            </div>
          </div>
          <div :class="dk('border-gray-800 bg-gray-900/50','border-gray-100 bg-gray-50')+ ' flex justify-end gap-3 p-6 border-t'">
            <button @click="showForm = false" :class="dk('bg-gray-800 hover:bg-gray-700 text-gray-300','bg-gray-100 hover:bg-gray-200 text-gray-700')" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="saveDriver" :disabled="saving" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              {{ editing ? 'Save Changes' : 'Create Driver' }}
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
            <div class="p-2 rounded-full bg-red-500/20">
              <AlertTriangle class="w-5 h-5 text-red-400" />
            </div>
            <h2 class="text-lg font-semibold">Delete Driver</h2>
          </div>
          <p :class="dk('text-gray-400','text-gray-600')" class="text-sm mb-6">
            Delete driver <strong>{{ deleteItem.full_name }}</strong>? This action cannot be undone.
          </p>
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
import { UserCheck, Plus, Search, Pencil, Trash2, X, Loader2, AlertTriangle, Users, ShieldAlert, Calendar, BadgeCheck } from 'lucide-vue-next'
import { fleetAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const dk = (dark: string, light: string) => app.darkMode ? dark : light

const drivers = ref<any[]>([])
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
  { value: 'available', label: 'Available' },
  { value: 'on_duty', label: 'On Duty' },
  { value: 'off_duty', label: 'Off Duty' },
  { value: 'on_leave', label: 'On Leave' },
]

const defaultForm = () => ({
  full_name: '', employee_id: '', license_number: '', license_class: 'B',
  license_expiry: '', status: 'available', phone: '', email: '',
  date_of_birth: '', national_id: '', medical_expiry: '', hire_date: '', address: '', notes: '',
})
const form = ref(defaultForm())

const inputCls = computed(() =>
  dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500',
    'bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400') +
  ' border rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-indigo-500')

const filtered = computed(() => {
  let list = drivers.value
  if (statusFilter.value !== 'all') list = list.filter(d => d.status === statusFilter.value)
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(d => [d.full_name, d.license_number, d.phone, d.employee_id].some(f => f?.toLowerCase().includes(q)))
  }
  return list
})

const kpis = computed(() => {
  const total = drivers.value.length
  const available = drivers.value.filter(d => d.status === 'available').length
  const onDuty = drivers.value.filter(d => d.status === 'on_duty').length
  const expiring = drivers.value.filter(d => isExpiringSoon(d.license_expiry)).length
  return [
    { label: 'Total Drivers', value: total, icon: Users, color: 'text-indigo-400', sub: null, subColor: '' },
    { label: 'Available', value: available, icon: BadgeCheck, color: 'text-green-400', sub: null, subColor: '' },
    { label: 'On Duty', value: onDuty, icon: UserCheck, color: 'text-blue-400', sub: null, subColor: '' },
    { label: 'License Expiring', value: expiring, icon: ShieldAlert, color: 'text-amber-400', sub: expiring > 0 ? 'Within 30 days' : null, subColor: 'text-amber-400' },
  ]
})

function formatEnum(s: string) { return s ? s.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()) : '—' }
function isExpiringSoon(date: string) {
  if (!date) return false
  const d = new Date(date), now = new Date(), soon = new Date()
  soon.setDate(soon.getDate() + 30)
  return d >= now && d <= soon
}
function isExpired(date: string) { return date ? new Date(date) < new Date() : false }
function expiryClass(date: string) {
  if (isExpired(date)) return 'text-red-400 font-medium'
  if (isExpiringSoon(date)) return 'text-amber-400 font-medium'
  return dk('text-gray-300','text-gray-700')
}
function statusColor(s: string) {
  const m: Record<string, string> = {
    available: 'bg-green-500/15 text-green-400',
    on_duty: 'bg-blue-500/15 text-blue-400',
    off_duty: 'bg-gray-500/15 text-gray-400',
    on_leave: 'bg-amber-500/15 text-amber-400',
    terminated: 'bg-red-500/15 text-red-400',
  }
  return m[s] || 'bg-gray-500/15 text-gray-400'
}

async function load() {
  loading.value = true
  try {
    const r = await fleetAPI.listDrivers()
    drivers.value = r.data.items || r.data || []
  } catch { app.addToast('Failed to load drivers', 'error') }
  finally { loading.value = false }
}

function openCreate() { editing.value = null; form.value = defaultForm(); showForm.value = true }
function openEdit(d: any) { editing.value = d; form.value = { ...defaultForm(), ...d }; showForm.value = true }
function openDelete(d: any) { deleteItem.value = d; showDelete.value = true }

async function saveDriver() {
  if (!form.value.full_name || !form.value.license_number) {
    app.addToast('Full name and license number are required', 'error'); return
  }
  saving.value = true
  try {
    if (editing.value) {
      await fleetAPI.updateDriver(editing.value.id, form.value)
      app.addToast('Driver updated', 'success')
    } else {
      await fleetAPI.createDriver(form.value)
      app.addToast('Driver created', 'success')
    }
    showForm.value = false; await load()
  } catch (e: any) { app.addToast(e?.response?.data?.error || 'Save failed', 'error') }
  finally { saving.value = false }
}

async function confirmDelete() {
  if (!deleteItem.value) return
  saving.value = true
  try {
    await fleetAPI.deleteDriver(deleteItem.value.id)
    app.addToast('Driver deleted', 'success')
    showDelete.value = false; await load()
  } catch { app.addToast('Delete failed', 'error') }
  finally { saving.value = false }
}

onMounted(load)
</script>
