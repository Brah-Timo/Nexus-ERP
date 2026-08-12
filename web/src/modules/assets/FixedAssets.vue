<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { assetsAPI } from '@/api/client'
import {
  Package, Plus, Search, RefreshCw, Edit, Trash2, Eye,
  XCircle, CheckCircle, Archive, Wrench, ChevronDown
} from 'lucide-vue-next'

const app = useAppStore()

const assets = ref<any[]>([])
const categories = ref<any[]>([])
const locations = ref<any[]>([])
const loading = ref(false)
const searchQuery = ref('')
const filterStatus = ref('')
const filterCategory = ref('')

const showModal = ref(false)
const showDetailModal = ref(false)
const showDisposeModal = ref(false)
const editId = ref('')
const detailAsset = ref<any>(null)
const form = ref<Record<string, any>>({})
const disposeForm = ref({ status: 'disposed', disposal_date: '', disposal_amount: 0, reason: '' })

const depMethods = [
  { value: 'straight_line', label: 'Straight Line' },
  { value: 'declining_balance', label: 'Declining Balance' },
  { value: 'double_declining', label: 'Double Declining' },
  { value: 'sum_of_years', label: 'Sum of Years' },
  { value: 'units_of_production', label: 'Units of Production' },
]
const statusOptions = [
  { value: '', label: 'All Statuses' },
  { value: 'active', label: 'Active' },
  { value: 'in_use', label: 'In Use' },
  { value: 'in_storage', label: 'In Storage' },
  { value: 'under_maintenance', label: 'Under Maintenance' },
  { value: 'disposed', label: 'Disposed' },
  { value: 'sold', label: 'Sold' },
  { value: 'written_off', label: 'Written Off' },
]
const conditionOptions = ['excellent', 'good', 'fair', 'poor', 'damaged']

async function load() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (filterStatus.value) params.status = filterStatus.value
    if (filterCategory.value) params.category_id = filterCategory.value
    if (searchQuery.value) params.q = searchQuery.value
    const [assRes, catRes, locRes] = await Promise.all([
      assetsAPI.listAssets(params),
      assetsAPI.listCategories(),
      assetsAPI.listLocations(),
    ])
    assets.value = assRes.data
    categories.value = catRes.data
    locations.value = locRes.data
  } finally {
    loading.value = false
  }
}

onMounted(load)

function openCreate() {
  editId.value = ''
  form.value = {
    name: '', description: '', status: 'active', condition: 'good',
    category_id: '', location_id: '',
    purchase_cost: 0, salvage_value: 0, useful_life_years: 5,
    depreciation_method: 'straight_line', depreciation_rate: 20,
    purchase_date: new Date().toISOString().slice(0, 10),
    in_service_date: '', serial_number: '', model: '', manufacturer: '',
    warranty_expiry_date: '', notes: '', is_depreciable: true,
  }
  showModal.value = true
}

function openEdit(a: any) {
  editId.value = a.id
  form.value = { ...a }
  showModal.value = true
}

async function openDetail(a: any) {
  const res = await assetsAPI.getAsset(a.id)
  detailAsset.value = res.data
  showDetailModal.value = true
}

function openDispose(a: any) {
  editId.value = a.id
  disposeForm.value = {
    status: 'disposed',
    disposal_date: new Date().toISOString().slice(0, 10),
    disposal_amount: a.net_book_value ?? 0,
    reason: '',
  }
  showDisposeModal.value = true
}

async function save() {
  if (editId.value) {
    await assetsAPI.updateAsset(editId.value, form.value)
  } else {
    await assetsAPI.createAsset(form.value)
  }
  showModal.value = false
  load()
}

async function remove(id: string) {
  if (!confirm('Delete this asset?')) return
  await assetsAPI.deleteAsset(id)
  load()
}

async function dispose() {
  await assetsAPI.disposeAsset(editId.value, disposeForm.value)
  showDisposeModal.value = false
  load()
}

function fmtCurrency(v: number | string) {
  const n = typeof v === 'string' ? parseFloat(v) : v
  return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n || 0)
}

function statusBadge(s: string) {
  const m: Record<string, string> = {
    active: 'bg-emerald-500/20 text-emerald-300',
    in_use: 'bg-sky-500/20 text-sky-300',
    in_storage: 'bg-slate-500/20 text-slate-300',
    under_maintenance: 'bg-amber-500/20 text-amber-300',
    disposed: 'bg-red-500/20 text-red-300',
    sold: 'bg-purple-500/20 text-purple-300',
    written_off: 'bg-gray-500/20 text-gray-400',
  }
  return m[s] ?? 'bg-slate-500/20 text-slate-300'
}

const filtered = computed(() => {
  let list = assets.value
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(a =>
      a.name?.toLowerCase().includes(q) ||
      a.asset_number?.toLowerCase().includes(q) ||
      a.serial_number?.toLowerCase().includes(q)
    )
  }
  return list
})

const cardCls = computed(() =>
  app.darkMode
    ? 'bg-slate-800/60 border-slate-700'
    : 'bg-white border-slate-200 shadow-sm'
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
</script>

<template>
  <div class="p-6 space-y-5">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Fixed Assets</h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
          Manage all fixed assets, depreciation, and lifecycle
        </p>
      </div>
      <div class="flex gap-2">
        <button @click="load" class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors"
          :class="app.darkMode ? 'border-slate-600 text-slate-300 hover:bg-slate-700' : 'border-slate-300 text-slate-600 hover:bg-slate-50'">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button @click="openCreate"
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white transition-colors">
          <Plus class="w-4 h-4" /> New Asset
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3">
      <div class="relative flex-1 min-w-[200px]">
        <Search class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <input v-model="searchQuery" placeholder="Search assets..." @input="load"
          class="w-full pl-9 pr-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500"
          :class="inputCls" />
      </div>
      <select v-model="filterStatus" @change="load"
        class="px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500"
        :class="inputCls">
        <option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
      </select>
      <select v-model="filterCategory" @change="load"
        class="px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500"
        :class="inputCls">
        <option value="">All Categories</option>
        <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
      </select>
    </div>

    <!-- Table -->
    <div class="rounded-xl border overflow-hidden" :class="cardCls">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b" :class="thCls">
              <th class="text-left px-4 py-3 font-medium">Asset Number</th>
              <th class="text-left px-4 py-3 font-medium">Name</th>
              <th class="text-left px-4 py-3 font-medium">Category</th>
              <th class="text-left px-4 py-3 font-medium">Location</th>
              <th class="text-left px-4 py-3 font-medium">Status</th>
              <th class="text-right px-4 py-3 font-medium">Cost</th>
              <th class="text-right px-4 py-3 font-medium">Accum. Dep.</th>
              <th class="text-right px-4 py-3 font-medium">Net Book Value</th>
              <th class="text-right px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="9" class="text-center py-10 text-slate-400">Loading...</td>
            </tr>
            <tr v-else-if="filtered.length === 0">
              <td colspan="9" class="text-center py-10 text-slate-400">No assets found</td>
            </tr>
            <tr v-for="a in filtered" :key="a.id"
              class="border-t transition-colors"
              :class="[tdCls, app.darkMode ? 'hover:bg-slate-700/30' : 'hover:bg-slate-50']">
              <td class="px-4 py-3 font-mono text-indigo-400 font-medium">{{ a.asset_number }}</td>
              <td class="px-4 py-3 font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
                {{ a.name }}
                <div class="text-xs font-normal text-slate-400">{{ a.serial_number }}</div>
              </td>
              <td class="px-4 py-3">{{ a.category_name || '—' }}</td>
              <td class="px-4 py-3">{{ a.location_name || '—' }}</td>
              <td class="px-4 py-3">
                <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="statusBadge(a.status)">
                  {{ a.status?.replace(/_/g,' ') }}
                </span>
              </td>
              <td class="px-4 py-3 text-right font-mono text-sm">{{ fmtCurrency(a.purchase_cost) }}</td>
              <td class="px-4 py-3 text-right font-mono text-sm text-amber-400">{{ fmtCurrency(a.accumulated_depreciation) }}</td>
              <td class="px-4 py-3 text-right font-mono text-sm text-emerald-400 font-medium">{{ fmtCurrency(a.net_book_value) }}</td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <button @click="openDetail(a)" class="p-1.5 rounded hover:bg-indigo-500/20 text-indigo-400 transition-colors" title="View Detail">
                    <Eye class="w-4 h-4" />
                  </button>
                  <button @click="openEdit(a)" class="p-1.5 rounded hover:bg-amber-500/20 text-amber-400 transition-colors" title="Edit">
                    <Edit class="w-4 h-4" />
                  </button>
                  <button @click="openDispose(a)" class="p-1.5 rounded hover:bg-purple-500/20 text-purple-400 transition-colors" title="Dispose">
                    <Archive class="w-4 h-4" />
                  </button>
                  <button @click="remove(a.id)" class="p-1.5 rounded hover:bg-red-500/20 text-red-400 transition-colors" title="Delete">
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
      <div class="w-full max-w-3xl rounded-2xl border max-h-[90vh] overflow-y-auto"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'">
        <div class="flex items-center justify-between p-5 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
            {{ editId ? 'Edit Asset' : 'New Asset' }}
          </h2>
          <button @click="showModal = false" class="text-slate-400 hover:text-slate-200"><XCircle class="w-5 h-5" /></button>
        </div>
        <div class="p-5 space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="md:col-span-2">
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Name *</label>
              <input v-model="form.name" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" placeholder="Asset name" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Category</label>
              <select v-model="form.category_id" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option value="">None</option>
                <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Location</label>
              <select v-model="form.location_id" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option value="">None</option>
                <option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Status</label>
              <select v-model="form.status" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option v-for="s in statusOptions.slice(1)" :key="s.value" :value="s.value">{{ s.label }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Condition</label>
              <select v-model="form.condition" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option v-for="c in conditionOptions" :key="c" :value="c">{{ c }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Purchase Cost</label>
              <input v-model.number="form.purchase_cost" type="number" step="0.01" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Salvage Value</label>
              <input v-model.number="form.salvage_value" type="number" step="0.01" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Useful Life (Years)</label>
              <input v-model.number="form.useful_life_years" type="number" step="0.5" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Depreciation Method</label>
              <select v-model="form.depreciation_method" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option v-for="m in depMethods" :key="m.value" :value="m.value">{{ m.label }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Depreciation Rate (%)</label>
              <input v-model.number="form.depreciation_rate" type="number" step="0.01" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Purchase Date</label>
              <input v-model="form.purchase_date" type="date" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">In Service Date</label>
              <input v-model="form.in_service_date" type="date" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Serial Number</label>
              <input v-model="form.serial_number" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Model</label>
              <input v-model="form.model" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Manufacturer</label>
              <input v-model="form.manufacturer" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Warranty Expiry</label>
              <input v-model="form.warranty_expiry_date" type="date" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Notes</label>
              <textarea v-model="form.notes" rows="2" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500 resize-none" :class="inputCls" />
            </div>
            <div class="flex items-center gap-2">
              <input v-model="form.is_depreciable" type="checkbox" id="isDepreciable" class="rounded" />
              <label for="isDepreciable" class="text-sm" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Is Depreciable</label>
            </div>
          </div>
        </div>
        <div class="flex justify-end gap-3 p-5 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showModal = false" class="px-4 py-2 rounded-lg text-sm border transition-colors"
            :class="app.darkMode ? 'border-slate-600 text-slate-300 hover:bg-slate-700' : 'border-slate-300 text-slate-600 hover:bg-slate-50'">
            Cancel
          </button>
          <button @click="save" class="px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white transition-colors">
            {{ editId ? 'Save Changes' : 'Create Asset' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Detail Modal -->
    <div v-if="showDetailModal && detailAsset" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60">
      <div class="w-full max-w-3xl rounded-2xl border max-h-[90vh] overflow-y-auto"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'">
        <div class="flex items-center justify-between p-5 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
            Asset Detail — {{ detailAsset.asset_number }}
          </h2>
          <button @click="showDetailModal = false" class="text-slate-400 hover:text-slate-200"><XCircle class="w-5 h-5" /></button>
        </div>
        <div class="p-5 space-y-4">
          <div class="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm">
            <div v-for="[k,v] in [
              ['Name', detailAsset.name],
              ['Asset Number', detailAsset.asset_number],
              ['Category', detailAsset.category_name],
              ['Location', detailAsset.location_name],
              ['Status', detailAsset.status?.replace(/_/g,' ')],
              ['Condition', detailAsset.condition],
              ['Purchase Cost', fmtCurrency(detailAsset.purchase_cost)],
              ['Salvage Value', fmtCurrency(detailAsset.salvage_value)],
              ['Accumulated Dep.', fmtCurrency(detailAsset.accumulated_depreciation)],
              ['Net Book Value', fmtCurrency(detailAsset.net_book_value)],
              ['Dep. Method', detailAsset.depreciation_method?.replace(/_/g,' ')],
              ['Useful Life', detailAsset.useful_life_years + ' yrs'],
              ['Serial Number', detailAsset.serial_number || '—'],
              ['Model', detailAsset.model || '—'],
              ['Manufacturer', detailAsset.manufacturer || '—'],
            ]" :key="k">
              <div>
                <div class="text-xs mb-0.5" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ k }}</div>
                <div class="font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">{{ v || '—' }}</div>
              </div>
            </div>
          </div>

          <!-- Depreciation Schedule -->
          <div v-if="detailAsset.depreciation_schedule?.length">
            <h3 class="text-sm font-semibold mb-2" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Depreciation Schedule</h3>
            <div class="overflow-x-auto rounded-lg border" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
              <table class="w-full text-xs">
                <thead :class="app.darkMode ? 'bg-slate-700/50' : 'bg-slate-50'">
                  <tr>
                    <th class="px-3 py-2 text-left font-medium">Period</th>
                    <th class="px-3 py-2 text-right font-medium">Opening</th>
                    <th class="px-3 py-2 text-right font-medium">Dep. Amount</th>
                    <th class="px-3 py-2 text-right font-medium">Closing</th>
                    <th class="px-3 py-2 text-center font-medium">Posted</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="s in detailAsset.depreciation_schedule.slice(0, 24)" :key="s.id"
                    class="border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
                    <td class="px-3 py-1.5">{{ s.period_year }}-{{ String(s.period_month).padStart(2,'0') }}</td>
                    <td class="px-3 py-1.5 text-right font-mono">{{ fmtCurrency(s.opening_book_value) }}</td>
                    <td class="px-3 py-1.5 text-right font-mono text-amber-400">{{ fmtCurrency(s.depreciation_amount) }}</td>
                    <td class="px-3 py-1.5 text-right font-mono text-emerald-400">{{ fmtCurrency(s.closing_book_value) }}</td>
                    <td class="px-3 py-1.5 text-center">
                      <CheckCircle v-if="s.is_posted" class="w-3.5 h-3.5 text-emerald-400 inline" />
                      <XCircle v-else class="w-3.5 h-3.5 text-slate-500 inline" />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Dispose Modal -->
    <div v-if="showDisposeModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60">
      <div class="w-full max-w-md rounded-2xl border"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'">
        <div class="flex items-center justify-between p-5 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Dispose Asset</h2>
          <button @click="showDisposeModal = false" class="text-slate-400 hover:text-slate-200"><XCircle class="w-5 h-5" /></button>
        </div>
        <div class="p-5 space-y-3">
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Disposal Type</label>
            <select v-model="disposeForm.status" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls">
              <option value="disposed">Disposed</option>
              <option value="sold">Sold</option>
              <option value="written_off">Written Off</option>
            </select>
          </div>
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Disposal Date</label>
            <input v-model="disposeForm.disposal_date" type="date" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
          </div>
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Disposal Amount</label>
            <input v-model.number="disposeForm.disposal_amount" type="number" step="0.01" class="w-full px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls" />
          </div>
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Reason</label>
            <textarea v-model="disposeForm.reason" rows="2" class="w-full px-3 py-2 rounded-lg border text-sm outline-none resize-none" :class="inputCls" />
          </div>
        </div>
        <div class="flex justify-end gap-3 p-5 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showDisposeModal = false" class="px-4 py-2 rounded-lg text-sm border" :class="app.darkMode ? 'border-slate-600 text-slate-300' : 'border-slate-300 text-slate-600'">Cancel</button>
          <button @click="dispose" class="px-4 py-2 rounded-lg text-sm font-medium bg-red-600 hover:bg-red-700 text-white">Confirm Disposal</button>
        </div>
      </div>
    </div>
  </div>
</template>
