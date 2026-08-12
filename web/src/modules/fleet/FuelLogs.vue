<template>
  <div :class="dk('min-h-screen bg-gray-950 text-gray-100','min-h-screen bg-gray-50 text-gray-900')">
    <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border-b px-6 py-4'">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-lg bg-yellow-600/20">
            <Fuel class="w-5 h-5 text-yellow-400" />
          </div>
          <div>
            <h1 class="text-lg font-semibold">Fuel Logs</h1>
            <p :class="dk('text-gray-400','text-gray-500')" class="text-xs">Track fuel consumption and costs</p>
          </div>
        </div>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 bg-yellow-600 hover:bg-yellow-700 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" />
          Log Fuel
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
          <div v-if="kpi.sub" :class="dk('text-gray-400','text-gray-500')" class="text-xs mt-1">{{ kpi.sub }}</div>
        </div>
      </div>

      <!-- Filters -->
      <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-4 flex flex-wrap gap-3 items-center'">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input v-model="search" placeholder="Search vehicle, station, driver…"
            :class="dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500','bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400')+ ' border rounded-lg pl-9 pr-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-yellow-500'" />
        </div>
        <div class="flex gap-2">
          <input v-model="dateFrom" type="date" :class="dk('bg-gray-800 border-gray-700 text-gray-100','bg-gray-50 border-gray-300 text-gray-900')+ ' border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-yellow-500'" />
          <input v-model="dateTo" type="date" :class="dk('bg-gray-800 border-gray-700 text-gray-100','bg-gray-50 border-gray-300 text-gray-900')+ ' border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-yellow-500'" />
        </div>
      </div>

      <!-- Table -->
      <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl overflow-hidden'">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead :class="dk('bg-gray-800/50 text-gray-400','bg-gray-50 text-gray-500')">
              <tr>
                <th class="px-4 py-3 text-left font-medium">Date</th>
                <th class="px-4 py-3 text-left font-medium">Vehicle</th>
                <th class="px-4 py-3 text-left font-medium">Driver</th>
                <th class="px-4 py-3 text-left font-medium">Fuel Type</th>
                <th class="px-4 py-3 text-right font-medium">Liters</th>
                <th class="px-4 py-3 text-right font-medium">Price/L</th>
                <th class="px-4 py-3 text-right font-medium">Total Cost</th>
                <th class="px-4 py-3 text-right font-medium">Mileage</th>
                <th class="px-4 py-3 text-left font-medium">Station</th>
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
                  <Fuel class="w-8 h-8 mx-auto mb-2 opacity-30" />No fuel logs found
                </td>
              </tr>
              <tr v-for="l in filtered" :key="l.id" :class="dk('hover:bg-gray-800/50','hover:bg-gray-50')" class="transition-colors">
                <td class="px-4 py-3 text-xs">{{ l.fill_date }}</td>
                <td class="px-4 py-3">
                  <div class="font-medium text-sm">{{ l.vehicle_plate || l.vehicle_id }}</div>
                </td>
                <td class="px-4 py-3 text-xs">{{ l.driver_name || '—' }}</td>
                <td class="px-4 py-3">
                  <span :class="fuelTypeColor(l.fuel_type)" class="px-2 py-0.5 rounded text-xs">{{ formatEnum(l.fuel_type) }}</span>
                </td>
                <td class="px-4 py-3 text-right font-medium">{{ l.liters?.toFixed(2) }} L</td>
                <td class="px-4 py-3 text-right text-xs" :class="dk('text-gray-300','text-gray-600')">{{ l.price_per_liter?.toFixed(2) }}</td>
                <td class="px-4 py-3 text-right font-semibold text-yellow-400">{{ fmtDZD(l.total_cost || l.liters * l.price_per_liter) }}</td>
                <td class="px-4 py-3 text-right text-xs">{{ l.mileage_at_fill != null ? l.mileage_at_fill.toLocaleString() + ' km' : '—' }}</td>
                <td class="px-4 py-3 text-xs">{{ l.fuel_station || '—' }}</td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-end gap-1">
                    <button @click="openEdit(l)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors">
                      <Pencil class="w-4 h-4 text-amber-400" />
                    </button>
                    <button @click="openDelete(l)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors">
                      <Trash2 class="w-4 h-4 text-red-400" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
            <tfoot v-if="filtered.length > 0" :class="dk('bg-gray-800/30 text-gray-300','bg-gray-50 text-gray-700')">
              <tr>
                <td colspan="4" class="px-4 py-3 font-medium text-sm">Totals ({{ filtered.length }} records)</td>
                <td class="px-4 py-3 text-right font-bold">{{ totalLiters.toFixed(2) }} L</td>
                <td class="px-4 py-3" />
                <td class="px-4 py-3 text-right font-bold text-yellow-400">{{ fmtDZD(totalCost) }}</td>
                <td colspan="3" />
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showForm" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showForm = false" />
        <div :class="dk('bg-gray-900 border-gray-700','bg-white border-gray-200')+ ' relative border rounded-2xl w-full max-w-xl max-h-[90vh] overflow-y-auto shadow-2xl'">
          <div :class="dk('border-gray-800','border-gray-100')+ ' flex items-center justify-between p-6 border-b'">
            <h2 class="text-lg font-semibold">{{ editing ? 'Edit Fuel Log' : 'Log Fuel Fill' }}</h2>
            <button @click="showForm = false" :class="dk('hover:bg-gray-800','hover:bg-gray-100')" class="p-2 rounded-lg transition-colors"><X class="w-5 h-5" /></button>
          </div>
          <div class="p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Vehicle ID <span class="text-red-400">*</span></label>
              <input v-model="form.vehicle_id" :class="inputCls" placeholder="Vehicle UUID" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Driver ID</label>
              <input v-model="form.driver_id" :class="inputCls" placeholder="Driver UUID" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Fill Date <span class="text-red-400">*</span></label>
              <input v-model="form.fill_date" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Fuel Type</label>
              <select v-model="form.fuel_type" :class="inputCls">
                <option value="gasoline">Gasoline</option>
                <option value="diesel">Diesel</option>
                <option value="lpg">LPG</option>
                <option value="electric">Electric</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Liters <span class="text-red-400">*</span></label>
              <input v-model.number="form.liters" type="number" step="0.01" :class="inputCls" placeholder="50.00" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Price per Liter (DZD) <span class="text-red-400">*</span></label>
              <input v-model.number="form.price_per_liter" type="number" step="0.01" :class="inputCls" placeholder="22.00" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Mileage at Fill (km)</label>
              <input v-model.number="form.mileage_at_fill" type="number" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Is Full Tank?</label>
              <select v-model="form.is_full_tank" :class="inputCls">
                <option :value="true">Yes</option>
                <option :value="false">No</option>
              </select>
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium mb-1 text-gray-400">Fuel Station</label>
              <input v-model="form.fuel_station" :class="inputCls" placeholder="Station name or location" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Receipt Number</label>
              <input v-model="form.receipt_number" :class="inputCls" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium mb-1 text-gray-400">Notes</label>
              <textarea v-model="form.notes" :class="inputCls" rows="2" />
            </div>
          </div>
          <!-- Calculated total preview -->
          <div v-if="form.liters && form.price_per_liter" :class="dk('bg-yellow-500/10 border-yellow-500/30','bg-yellow-50 border-yellow-200')+ ' mx-6 mb-4 border rounded-lg p-3 flex items-center justify-between'">
            <span :class="dk('text-yellow-300','text-yellow-700')" class="text-sm">Calculated Total Cost</span>
            <span class="text-lg font-bold text-yellow-400">{{ fmtDZD(form.liters * form.price_per_liter) }}</span>
          </div>
          <div :class="dk('border-gray-800 bg-gray-900/50','border-gray-100 bg-gray-50')+ ' flex justify-end gap-3 p-6 border-t'">
            <button @click="showForm = false" :class="dk('bg-gray-800 hover:bg-gray-700 text-gray-300','bg-gray-100 hover:bg-gray-200 text-gray-700')" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="save" :disabled="saving" class="px-4 py-2 bg-yellow-600 hover:bg-yellow-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              {{ editing ? 'Save Changes' : 'Log Fuel' }}
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
            <h2 class="text-lg font-semibold">Delete Fuel Log</h2>
          </div>
          <p :class="dk('text-gray-400','text-gray-600')" class="text-sm mb-6">Delete fuel log dated <strong>{{ deleteItem.fill_date }}</strong>?</p>
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
import { Fuel, Plus, Search, Pencil, Trash2, X, Loader2, AlertTriangle, DropletIcon, DollarSign, BarChart3, TrendingUp } from 'lucide-vue-next'
import { fleetAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const dk = (dark: string, light: string) => app.darkMode ? dark : light

const logs = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const search = ref('')
const dateFrom = ref('')
const dateTo = ref('')
const showForm = ref(false)
const showDelete = ref(false)
const editing = ref<any>(null)
const deleteItem = ref<any>(null)

const defaultForm = () => ({
  vehicle_id: '', driver_id: '', fill_date: new Date().toISOString().split('T')[0],
  fuel_type: 'diesel', liters: null, price_per_liter: null, mileage_at_fill: null,
  is_full_tank: true, fuel_station: '', receipt_number: '', notes: '',
})
const form = ref(defaultForm())

const inputCls = computed(() =>
  dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500', 'bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400') +
  ' border rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-yellow-500')

const filtered = computed(() => {
  let list = logs.value
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(l => [l.vehicle_plate, l.driver_name, l.fuel_station].some(f => f?.toLowerCase().includes(q)))
  }
  if (dateFrom.value) list = list.filter(l => l.fill_date >= dateFrom.value)
  if (dateTo.value) list = list.filter(l => l.fill_date <= dateTo.value)
  return list
})

const totalLiters = computed(() => filtered.value.reduce((s, l) => s + (l.liters || 0), 0))
const totalCost = computed(() => filtered.value.reduce((s, l) => s + (l.total_cost || l.liters * l.price_per_liter || 0), 0))

const kpis = computed(() => {
  const avgPL = logs.value.length > 0
    ? logs.value.reduce((s, l) => s + (l.price_per_liter || 0), 0) / logs.value.length : 0
  return [
    { label: 'Total Records', value: logs.value.length, icon: DropletIcon, color: 'text-yellow-400', sub: null },
    { label: 'Total Liters', value: totalLiters.value.toFixed(0) + ' L', icon: Fuel, color: 'text-blue-400', sub: null },
    { label: 'Total Cost', value: fmtDZD(totalCost.value), icon: DollarSign, color: 'text-green-400', sub: null },
    { label: 'Avg Price/L', value: avgPL.toFixed(2) + ' DZD', icon: TrendingUp, color: 'text-purple-400', sub: null },
  ]
})

function formatEnum(s: string) { return s ? s.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()) : '—' }
function fmtDZD(n: number) { return new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n || 0) + ' DZD' }
function fuelTypeColor(t: string) {
  const m: Record<string, string> = {
    diesel: 'bg-amber-500/15 text-amber-400',
    gasoline: 'bg-blue-500/15 text-blue-400',
    lpg: 'bg-green-500/15 text-green-400',
    electric: 'bg-purple-500/15 text-purple-400',
  }
  return m[t] || 'bg-gray-500/15 text-gray-400'
}

async function load() {
  loading.value = true
  try {
    const r = await fleetAPI.listFuelLogs()
    logs.value = r.data.items || r.data || []
  } catch { app.addToast('Failed to load fuel logs', 'error') }
  finally { loading.value = false }
}

function openCreate() { editing.value = null; form.value = defaultForm(); showForm.value = true }
function openEdit(l: any) { editing.value = l; form.value = { ...defaultForm(), ...l }; showForm.value = true }
function openDelete(l: any) { deleteItem.value = l; showDelete.value = true }

async function save() {
  if (!form.value.vehicle_id || !form.value.fill_date || !form.value.liters || !form.value.price_per_liter) {
    app.addToast('Vehicle, date, liters, and price are required', 'error'); return
  }
  saving.value = true
  try {
    if (editing.value) {
      await fleetAPI.updateFuelLog(editing.value.id, form.value)
      app.addToast('Fuel log updated', 'success')
    } else {
      await fleetAPI.createFuelLog(form.value)
      app.addToast('Fuel log created', 'success')
    }
    showForm.value = false; await load()
  } catch (e: any) { app.addToast(e?.response?.data?.error || 'Save failed', 'error') }
  finally { saving.value = false }
}

async function confirmDelete() {
  if (!deleteItem.value) return
  saving.value = true
  try {
    await fleetAPI.deleteFuelLog(deleteItem.value.id)
    app.addToast('Fuel log deleted', 'success')
    showDelete.value = false; await load()
  } catch { app.addToast('Delete failed', 'error') }
  finally { saving.value = false }
}

onMounted(load)
</script>
