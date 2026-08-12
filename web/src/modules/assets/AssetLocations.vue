<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { assetsAPI } from '@/api/client'
import { MapPin, Plus, Edit, Trash2, RefreshCw, XCircle, ChevronRight, Building2 } from 'lucide-vue-next'

const app = useAppStore()
const locations = ref<any[]>([])
const loading = ref(false)
const showModal = ref(false)
const editId = ref('')
const form = ref<Record<string, any>>({})

async function load() {
  loading.value = true
  try {
    const res = await assetsAPI.listLocations()
    locations.value = res.data
  } finally {
    loading.value = false
  }
}

onMounted(load)

function openCreate() {
  editId.value = ''
  form.value = {
    name: '', code: '', description: '',
    address: '', city: '', country: '',
    parent_id: '', is_active: true,
  }
  showModal.value = true
}

function openEdit(loc: any) {
  editId.value = loc.id
  form.value = { ...loc }
  showModal.value = true
}

async function save() {
  if (editId.value) {
    await assetsAPI.updateLocation(editId.value, form.value)
  } else {
    await assetsAPI.createLocation(form.value)
  }
  showModal.value = false
  load()
}

async function remove(id: string) {
  if (!confirm('Delete this location?')) return
  try {
    await assetsAPI.deleteLocation(id)
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Delete failed')
  }
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

function fmtCurrency(v: number) {
  return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0)
}

const parentOptions = computed(() => locations.value.filter(l => l.id !== editId.value))
const totalAssets = computed(() => locations.value.reduce((s, l) => s + l.asset_count, 0))
</script>

<template>
  <div class="p-6 space-y-5">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Asset Locations</h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
          Manage physical locations and departments for asset tracking
        </p>
      </div>
      <div class="flex gap-2">
        <button @click="load" class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors"
          :class="app.darkMode ? 'border-slate-600 text-slate-300 hover:bg-slate-700' : 'border-slate-300 text-slate-600 hover:bg-slate-50'">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white">
          <Plus class="w-4 h-4" /> New Location
        </button>
      </div>
    </div>

    <!-- Summary -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div v-for="stat in [
        { label: 'Total Locations', value: locations.length, color: 'text-indigo-400' },
        { label: 'Active', value: locations.filter(l=>l.is_active).length, color: 'text-emerald-400' },
        { label: 'Top-Level', value: locations.filter(l=>!l.parent_id).length, color: 'text-amber-400' },
        { label: 'Total Assets', value: totalAssets, color: 'text-sky-400' },
      ]" :key="stat.label" class="rounded-xl border p-4" :class="cardCls">
        <div class="text-xs mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ stat.label }}</div>
        <div class="text-2xl font-bold" :class="stat.color">{{ stat.value }}</div>
      </div>
    </div>

    <!-- Table -->
    <div class="rounded-xl border overflow-hidden" :class="cardCls">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b" :class="thCls">
              <th class="text-left px-4 py-3 font-medium">Location</th>
              <th class="text-left px-4 py-3 font-medium">Code</th>
              <th class="text-left px-4 py-3 font-medium">Parent</th>
              <th class="text-left px-4 py-3 font-medium">City / Country</th>
              <th class="text-right px-4 py-3 font-medium">Assets</th>
              <th class="text-right px-4 py-3 font-medium">Net Book Value</th>
              <th class="text-center px-4 py-3 font-medium">Active</th>
              <th class="text-right px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="8" class="text-center py-10 text-slate-400">Loading...</td>
            </tr>
            <tr v-else-if="locations.length === 0">
              <td colspan="8" class="text-center py-10 text-slate-400">No locations found</td>
            </tr>
            <tr v-for="loc in locations" :key="loc.id"
              class="border-t transition-colors"
              :class="[tdCls, app.darkMode ? 'hover:bg-slate-700/30' : 'hover:bg-slate-50']">
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <MapPin class="w-4 h-4 text-sky-400 flex-shrink-0" />
                  <div>
                    <div class="font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">{{ loc.name }}</div>
                    <div v-if="loc.address" class="text-xs text-slate-400">{{ loc.address }}</div>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3 font-mono text-xs text-indigo-400">{{ loc.code || '—' }}</td>
              <td class="px-4 py-3">
                <span v-if="loc.parent_name" class="flex items-center gap-1 text-xs text-slate-400">
                  <ChevronRight class="w-3 h-3" /> {{ loc.parent_name }}
                </span>
                <span v-else class="text-xs text-slate-500">—</span>
              </td>
              <td class="px-4 py-3 text-xs">
                <span v-if="loc.city || loc.country">{{ [loc.city, loc.country].filter(Boolean).join(', ') }}</span>
                <span v-else class="text-slate-500">—</span>
              </td>
              <td class="px-4 py-3 text-right font-medium">{{ loc.asset_count }}</td>
              <td class="px-4 py-3 text-right font-mono text-emerald-400">{{ fmtCurrency(loc.total_net_book_value) }}</td>
              <td class="px-4 py-3 text-center">
                <span class="px-2 py-0.5 rounded-full text-xs font-medium"
                  :class="loc.is_active ? 'bg-emerald-500/20 text-emerald-300' : 'bg-slate-500/20 text-slate-400'">
                  {{ loc.is_active ? 'Yes' : 'No' }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <button @click="openEdit(loc)" class="p-1.5 rounded hover:bg-amber-500/20 text-amber-400 transition-colors">
                    <Edit class="w-4 h-4" />
                  </button>
                  <button @click="remove(loc.id)" class="p-1.5 rounded hover:bg-red-500/20 text-red-400 transition-colors">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60">
      <div class="w-full max-w-lg rounded-2xl border"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'">
        <div class="flex items-center justify-between p-5 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
            {{ editId ? 'Edit Location' : 'New Location' }}
          </h2>
          <button @click="showModal = false" class="text-slate-400 hover:text-slate-200"><XCircle class="w-5 h-5" /></button>
        </div>
        <div class="p-5 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="col-span-2">
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Name *</label>
              <input v-model="form.name" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Code</label>
              <input v-model="form.code" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Parent Location</label>
              <select v-model="form.parent_id" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option value="">None</option>
                <option v-for="l in parentOptions" :key="l.id" :value="l.id">{{ l.name }}</option>
              </select>
            </div>
            <div class="col-span-2">
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Address</label>
              <input v-model="form.address" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">City</label>
              <input v-model="form.city" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Country</label>
              <input v-model="form.country" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div class="col-span-2">
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Description</label>
              <textarea v-model="form.description" rows="2" class="w-full px-3 py-2 rounded-lg border text-sm outline-none resize-none" :class="inputCls" />
            </div>
            <div class="flex items-center gap-2">
              <input v-model="form.is_active" type="checkbox" id="locActive" class="rounded" />
              <label for="locActive" class="text-sm" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Active</label>
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
  </div>
</template>
