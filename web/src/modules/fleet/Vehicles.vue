<template>
  <div :class="dk('min-h-screen bg-gray-950 text-gray-100','min-h-screen bg-gray-50 text-gray-900')">
    <!-- Header -->
    <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border-b px-6 py-4'">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-lg bg-blue-600/20">
            <Truck class="w-5 h-5 text-blue-400" />
          </div>
          <div>
            <h1 class="text-lg font-semibold">Fleet Vehicles</h1>
            <p :class="dk('text-gray-400','text-gray-500')" class="text-xs">Manage your vehicle fleet</p>
          </div>
        </div>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" />
          Add Vehicle
        </button>
      </div>
    </div>

    <div class="p-6 space-y-6">
      <!-- KPI Cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 xl:grid-cols-6 gap-4">
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
            <input v-model="search" placeholder="Search plate, make, model…"
              :class="dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500','bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400')+ ' border rounded-lg pl-9 pr-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-blue-500'" />
          </div>
          <div class="flex gap-2 flex-wrap">
            <button v-for="s in statusFilters" :key="s.value"
              @click="statusFilter = s.value"
              :class="statusFilter === s.value ? 'bg-blue-600 text-white' : dk('bg-gray-800 text-gray-300 hover:bg-gray-700','bg-gray-100 text-gray-600 hover:bg-gray-200')"
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
                <th class="px-4 py-3 text-left font-medium">Vehicle</th>
                <th class="px-4 py-3 text-left font-medium">Plate</th>
                <th class="px-4 py-3 text-left font-medium">Type</th>
                <th class="px-4 py-3 text-left font-medium">Status</th>
                <th class="px-4 py-3 text-left font-medium">Fuel</th>
                <th class="px-4 py-3 text-left font-medium">Year</th>
                <th class="px-4 py-3 text-left font-medium">Mileage</th>
                <th class="px-4 py-3 text-left font-medium">Insurance</th>
                <th class="px-4 py-3 text-left font-medium">Inspection</th>
                <th class="px-4 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody :class="dk('divide-gray-800','divide-gray-100')+ ' divide-y'">
              <tr v-if="loading" class="text-center">
                <td colspan="10" class="py-12 text-gray-400">
                  <Loader2 class="w-6 h-6 animate-spin mx-auto mb-2" />
                  Loading vehicles…
                </td>
              </tr>
              <tr v-else-if="filtered.length === 0">
                <td colspan="10" class="py-12 text-center text-gray-400">
                  <Truck class="w-8 h-8 mx-auto mb-2 opacity-30" />
                  No vehicles found
                </td>
              </tr>
              <tr v-for="v in filtered" :key="v.id"
                :class="[dk('hover:bg-gray-800/50','hover:bg-gray-50'), isExpiringSoon(v.insurance_expiry) ? dk('border-l-2 border-l-amber-500','border-l-2 border-l-amber-500') : '']"
                class="transition-colors">
                <td class="px-4 py-3">
                  <div class="font-medium">{{ v.make }} {{ v.model }}</div>
                  <div :class="dk('text-gray-400','text-gray-500')" class="text-xs">{{ v.code }}</div>
                </td>
                <td class="px-4 py-3 font-mono text-sm">{{ v.plate_number }}</td>
                <td class="px-4 py-3">
                  <span :class="typeColor(v.vehicle_type)" class="px-2 py-0.5 rounded text-xs font-medium">{{ formatEnum(v.vehicle_type) }}</span>
                </td>
                <td class="px-4 py-3">
                  <span :class="statusColor(v.status)" class="px-2 py-0.5 rounded-full text-xs font-medium">{{ formatEnum(v.status) }}</span>
                </td>
                <td class="px-4 py-3 text-xs">{{ formatEnum(v.fuel_type) }}</td>
                <td class="px-4 py-3">{{ v.year || '—' }}</td>
                <td class="px-4 py-3">{{ v.mileage != null ? v.mileage.toLocaleString() + ' km' : '—' }}</td>
                <td class="px-4 py-3">
                  <span v-if="v.insurance_expiry" :class="expiryClass(v.insurance_expiry)" class="text-xs">{{ v.insurance_expiry }}</span>
                  <span v-else class="text-gray-400 text-xs">—</span>
                </td>
                <td class="px-4 py-3">
                  <span v-if="v.next_inspection_date" :class="expiryClass(v.next_inspection_date)" class="text-xs">{{ v.next_inspection_date }}</span>
                  <span v-else class="text-gray-400 text-xs">—</span>
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-end gap-1">
                    <button @click="openView(v)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors" title="View">
                      <Eye class="w-4 h-4 text-blue-400" />
                    </button>
                    <button @click="openEdit(v)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors" title="Edit">
                      <Pencil class="w-4 h-4 text-amber-400" />
                    </button>
                    <button @click="openDelete(v)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors" title="Delete">
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
        <div :class="dk('bg-gray-900 border-gray-700','bg-white border-gray-200')+ ' relative border rounded-2xl w-full max-w-3xl max-h-[90vh] overflow-y-auto shadow-2xl'">
          <div :class="dk('border-gray-800','border-gray-100')+ ' flex items-center justify-between p-6 border-b'">
            <h2 class="text-lg font-semibold">{{ editing ? 'Edit Vehicle' : 'Add Vehicle' }}</h2>
            <button @click="showForm = false" :class="dk('hover:bg-gray-800','hover:bg-gray-100')" class="p-2 rounded-lg transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Code</label>
              <input v-model="form.code" :class="inputCls" placeholder="VEH-0001 (auto)" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Plate Number <span class="text-red-400">*</span></label>
              <input v-model="form.plate_number" :class="inputCls" placeholder="123 TUN 16" required />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Make <span class="text-red-400">*</span></label>
              <input v-model="form.make" :class="inputCls" placeholder="Toyota" required />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Model <span class="text-red-400">*</span></label>
              <input v-model="form.model" :class="inputCls" placeholder="Land Cruiser" required />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Year</label>
              <input v-model.number="form.year" type="number" :class="inputCls" placeholder="2022" min="1990" max="2030" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">VIN</label>
              <input v-model="form.vin" :class="inputCls" placeholder="VIN number" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Color</label>
              <input v-model="form.color" :class="inputCls" placeholder="White" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Vehicle Type</label>
              <select v-model="form.vehicle_type" :class="inputCls">
                <option value="">— Select —</option>
                <option v-for="t in vehicleTypes" :key="t" :value="t">{{ formatEnum(t) }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Fuel Type</label>
              <select v-model="form.fuel_type" :class="inputCls">
                <option value="">— Select —</option>
                <option v-for="t in fuelTypes" :key="t" :value="t">{{ formatEnum(t) }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Status</label>
              <select v-model="form.status" :class="inputCls">
                <option value="active">Active</option>
                <option value="in_use">In Use</option>
                <option value="maintenance">Maintenance</option>
                <option value="inactive">Inactive</option>
                <option value="retired">Retired</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Mileage (km)</label>
              <input v-model.number="form.mileage" type="number" :class="inputCls" placeholder="0" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Fuel Capacity (L)</label>
              <input v-model.number="form.fuel_capacity_liters" type="number" :class="inputCls" placeholder="60" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Purchase Date</label>
              <input v-model="form.purchase_date" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Purchase Cost (DZD)</label>
              <input v-model.number="form.purchase_cost" type="number" :class="inputCls" placeholder="0" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Insurance Expiry</label>
              <input v-model="form.insurance_expiry" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Registration Expiry</label>
              <input v-model="form.registration_expiry" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Next Inspection</label>
              <input v-model="form.next_inspection_date" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Department</label>
              <input v-model="form.department" :class="inputCls" placeholder="Operations" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium mb-1 text-gray-400">Notes</label>
              <textarea v-model="form.notes" :class="inputCls" rows="2" placeholder="Additional notes…" />
            </div>
          </div>
          <div :class="dk('border-gray-800 bg-gray-900/50','border-gray-100 bg-gray-50')+ ' flex justify-end gap-3 p-6 border-t'">
            <button @click="showForm = false" :class="dk('bg-gray-800 hover:bg-gray-700 text-gray-300','bg-gray-100 hover:bg-gray-200 text-gray-700')" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="saveVehicle" :disabled="saving" class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              {{ editing ? 'Save Changes' : 'Create Vehicle' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- View Modal -->
    <Teleport to="body">
      <div v-if="showView && viewItem" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showView = false" />
        <div :class="dk('bg-gray-900 border-gray-700','bg-white border-gray-200')+ ' relative border rounded-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto shadow-2xl'">
          <div :class="dk('border-gray-800','border-gray-100')+ ' flex items-center justify-between p-6 border-b'">
            <div>
              <h2 class="text-lg font-semibold">{{ viewItem.make }} {{ viewItem.model }}</h2>
              <p :class="dk('text-gray-400','text-gray-500')" class="text-sm">{{ viewItem.plate_number }}</p>
            </div>
            <button @click="showView = false" :class="dk('hover:bg-gray-800','hover:bg-gray-100')" class="p-2 rounded-lg transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="p-6">
            <div class="grid grid-cols-2 gap-4">
              <div v-for="f in viewFields" :key="f.label" :class="dk('bg-gray-800/50','bg-gray-50')+ ' rounded-lg p-3'">
                <div :class="dk('text-gray-400','text-gray-500')" class="text-xs mb-1">{{ f.label }}</div>
                <div class="text-sm font-medium">{{ f.value || '—' }}</div>
              </div>
            </div>
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
            <h2 class="text-lg font-semibold">Delete Vehicle</h2>
          </div>
          <p :class="dk('text-gray-400','text-gray-600')" class="text-sm mb-6">
            Delete <strong>{{ deleteItem.make }} {{ deleteItem.model }}</strong> ({{ deleteItem.plate_number }})? This action cannot be undone.
          </p>
          <div class="flex justify-end gap-3">
            <button @click="showDelete = false" :class="dk('bg-gray-800 hover:bg-gray-700 text-gray-300','bg-gray-100 hover:bg-gray-200 text-gray-700')" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="confirmDelete" :disabled="saving" class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              Delete
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Truck, Plus, Search, Eye, Pencil, Trash2, X, Loader2, AlertTriangle, Gauge, ShieldCheck, Wrench, CircleOff } from 'lucide-vue-next'
import { fleetAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const dk = (dark: string, light: string) => app.darkMode ? dark : light

const vehicles = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const search = ref('')
const statusFilter = ref('all')

const showForm = ref(false)
const showView = ref(false)
const showDelete = ref(false)
const editing = ref<any>(null)
const viewItem = ref<any>(null)
const deleteItem = ref<any>(null)

const vehicleTypes = ['car','van','truck','bus','motorcycle','heavy_equipment','trailer','other']
const fuelTypes = ['gasoline','diesel','electric','hybrid','lpg','cng']

const statusFilters = [
  { value: 'all', label: 'All' },
  { value: 'active', label: 'Active' },
  { value: 'in_use', label: 'In Use' },
  { value: 'maintenance', label: 'Maintenance' },
  { value: 'inactive', label: 'Inactive' },
]

const defaultForm = () => ({
  code: '', plate_number: '', make: '', model: '', year: null, vin: '', color: '',
  vehicle_type: 'car', fuel_type: 'diesel', status: 'active', mileage: 0,
  fuel_capacity_liters: null, purchase_date: '', purchase_cost: null,
  insurance_expiry: '', registration_expiry: '', next_inspection_date: '',
  department: '', notes: '',
})
const form = ref(defaultForm())

const inputCls = computed(() =>
  dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500',
    'bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400') +
  ' border rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-blue-500')

const filtered = computed(() => {
  let list = vehicles.value
  if (statusFilter.value !== 'all') list = list.filter(v => v.status === statusFilter.value)
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(v => [v.plate_number, v.make, v.model, v.code].some(f => f?.toLowerCase().includes(q)))
  }
  return list
})

const kpis = computed(() => {
  const total = vehicles.value.length
  const active = vehicles.value.filter(v => v.status === 'active' || v.status === 'in_use').length
  const inMaint = vehicles.value.filter(v => v.status === 'maintenance').length
  const expiringSoon = vehicles.value.filter(v => isExpiringSoon(v.insurance_expiry)).length
  const expired = vehicles.value.filter(v => isExpired(v.insurance_expiry)).length
  return [
    { label: 'Total', value: total, icon: Truck, color: 'text-blue-400', sub: null, subColor: '' },
    { label: 'Active', value: active, icon: Gauge, color: 'text-green-400', sub: null, subColor: '' },
    { label: 'In Maintenance', value: inMaint, icon: Wrench, color: 'text-amber-400', sub: null, subColor: '' },
    { label: 'Insurance Expiring', value: expiringSoon, icon: ShieldCheck, color: 'text-orange-400', sub: expiringSoon > 0 ? 'Within 30 days' : null, subColor: 'text-orange-400' },
    { label: 'Insurance Expired', value: expired, icon: CircleOff, color: 'text-red-400', sub: null, subColor: '' },
    { label: 'Inactive', value: vehicles.value.filter(v => v.status === 'inactive' || v.status === 'retired').length, icon: CircleOff, color: 'text-gray-400', sub: null, subColor: '' },
  ]
})

const viewFields = computed(() => {
  if (!viewItem.value) return []
  const v = viewItem.value
  return [
    { label: 'Code', value: v.code },
    { label: 'Plate Number', value: v.plate_number },
    { label: 'Make / Model', value: `${v.make} ${v.model}` },
    { label: 'Year', value: v.year },
    { label: 'VIN', value: v.vin },
    { label: 'Color', value: v.color },
    { label: 'Type', value: formatEnum(v.vehicle_type) },
    { label: 'Fuel Type', value: formatEnum(v.fuel_type) },
    { label: 'Status', value: formatEnum(v.status) },
    { label: 'Mileage', value: v.mileage != null ? v.mileage.toLocaleString() + ' km' : null },
    { label: 'Fuel Capacity', value: v.fuel_capacity_liters ? v.fuel_capacity_liters + ' L' : null },
    { label: 'Purchase Date', value: v.purchase_date },
    { label: 'Purchase Cost', value: v.purchase_cost != null ? fmtDZD(v.purchase_cost) : null },
    { label: 'Insurance Expiry', value: v.insurance_expiry },
    { label: 'Registration Expiry', value: v.registration_expiry },
    { label: 'Next Inspection', value: v.next_inspection_date },
    { label: 'Department', value: v.department },
    { label: 'Notes', value: v.notes },
  ]
})

function formatEnum(s: string) {
  return s ? s.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()) : '—'
}
function fmtDZD(n: number) {
  return new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'
}
function isExpired(date: string) {
  if (!date) return false
  return new Date(date) < new Date()
}
function isExpiringSoon(date: string) {
  if (!date) return false
  const d = new Date(date)
  const now = new Date()
  const soon = new Date()
  soon.setDate(soon.getDate() + 30)
  return d >= now && d <= soon
}
function expiryClass(date: string) {
  if (isExpired(date)) return 'text-red-400 font-medium'
  if (isExpiringSoon(date)) return 'text-amber-400 font-medium'
  return dk('text-gray-300','text-gray-700')
}
function statusColor(s: string) {
  const m: Record<string, string> = {
    active: 'bg-green-500/15 text-green-400',
    in_use: 'bg-blue-500/15 text-blue-400',
    maintenance: 'bg-amber-500/15 text-amber-400',
    inactive: 'bg-gray-500/15 text-gray-400',
    retired: 'bg-red-500/15 text-red-400',
  }
  return m[s] || 'bg-gray-500/15 text-gray-400'
}
function typeColor(t: string) {
  const m: Record<string, string> = {
    car: 'bg-blue-500/15 text-blue-400',
    van: 'bg-purple-500/15 text-purple-400',
    truck: 'bg-orange-500/15 text-orange-400',
    bus: 'bg-green-500/15 text-green-400',
    motorcycle: 'bg-pink-500/15 text-pink-400',
    heavy_equipment: 'bg-red-500/15 text-red-400',
  }
  return m[t] || 'bg-gray-500/15 text-gray-400'
}

async function load() {
  loading.value = true
  try {
    const r = await fleetAPI.listVehicles()
    vehicles.value = r.data.items || r.data || []
  } catch { app.addToast('Failed to load vehicles', 'error') }
  finally { loading.value = false }
}

function openCreate() {
  editing.value = null
  form.value = defaultForm()
  showForm.value = true
}
function openEdit(v: any) {
  editing.value = v
  form.value = { ...defaultForm(), ...v }
  showForm.value = true
}
function openView(v: any) {
  viewItem.value = v
  showView.value = true
}
function openDelete(v: any) {
  deleteItem.value = v
  showDelete.value = true
}

async function saveVehicle() {
  if (!form.value.plate_number || !form.value.make || !form.value.model) {
    app.addToast('Plate number, make, and model are required', 'error'); return
  }
  saving.value = true
  try {
    if (editing.value) {
      await fleetAPI.updateVehicle(editing.value.id, form.value)
      app.addToast('Vehicle updated', 'success')
    } else {
      await fleetAPI.createVehicle(form.value)
      app.addToast('Vehicle created', 'success')
    }
    showForm.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally { saving.value = false }
}

async function confirmDelete() {
  if (!deleteItem.value) return
  saving.value = true
  try {
    await fleetAPI.deleteVehicle(deleteItem.value.id)
    app.addToast('Vehicle deleted', 'success')
    showDelete.value = false
    await load()
  } catch { app.addToast('Delete failed', 'error') }
  finally { saving.value = false }
}

onMounted(load)
</script>
