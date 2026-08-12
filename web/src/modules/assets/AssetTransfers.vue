<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { assetsAPI } from '@/api/client'
import {
  ArrowLeftRight, Plus, RefreshCw, XCircle,
  CheckCircle, Clock, Truck, Trash2, Eye
} from 'lucide-vue-next'

const app = useAppStore()
const transfers = ref<any[]>([])
const assets = ref<any[]>([])
const locations = ref<any[]>([])
const loading = ref(false)
const filterStatus = ref('')
const showModal = ref(false)
const form = ref<Record<string, any>>({})

const statusOptions = [
  { value: '', label: 'All' },
  { value: 'pending', label: 'Pending' },
  { value: 'approved', label: 'Approved' },
  { value: 'in_transit', label: 'In Transit' },
  { value: 'completed', label: 'Completed' },
  { value: 'cancelled', label: 'Cancelled' },
]

async function load() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (filterStatus.value) params.status = filterStatus.value
    const [tRes, aRes, lRes] = await Promise.all([
      assetsAPI.listTransfers(params),
      assetsAPI.listAssets(),
      assetsAPI.listLocations(),
    ])
    transfers.value = tRes.data
    assets.value = aRes.data
    locations.value = lRes.data
  } finally {
    loading.value = false
  }
}

onMounted(load)

function openCreate() {
  form.value = {
    asset_id: '', from_location_id: '', to_location_id: '',
    transfer_date: new Date().toISOString().slice(0, 10),
    reason: '', notes: '',
  }
  showModal.value = true
}

async function save() {
  await assetsAPI.createTransfer(form.value)
  showModal.value = false
  load()
}

async function approve(id: string) {
  if (!confirm('Approve this transfer?')) return
  await assetsAPI.approveTransfer(id)
  load()
}

async function complete(id: string) {
  if (!confirm('Mark transfer as completed? This will move the asset to the new location.')) return
  await assetsAPI.completeTransfer(id)
  load()
}

async function remove(id: string) {
  if (!confirm('Delete this transfer?')) return
  await assetsAPI.deleteTransfer(id)
  load()
}

function statusBadge(s: string) {
  const m: Record<string, string> = {
    pending: 'bg-amber-500/20 text-amber-300',
    approved: 'bg-sky-500/20 text-sky-300',
    in_transit: 'bg-indigo-500/20 text-indigo-300',
    completed: 'bg-emerald-500/20 text-emerald-300',
    cancelled: 'bg-red-500/20 text-red-300',
  }
  return m[s] ?? 'bg-slate-500/20 text-slate-300'
}

function fmtDate(d: any) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString()
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

const kpis = computed(() => [
  { label: 'Total', value: transfers.value.length, color: 'text-indigo-400' },
  { label: 'Pending', value: transfers.value.filter(t => t.status === 'pending').length, color: 'text-amber-400' },
  { label: 'Approved', value: transfers.value.filter(t => t.status === 'approved').length, color: 'text-sky-400' },
  { label: 'Completed', value: transfers.value.filter(t => t.status === 'completed').length, color: 'text-emerald-400' },
])
</script>

<template>
  <div class="p-6 space-y-5">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Asset Transfers</h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Track and manage asset transfers between locations</p>
      </div>
      <div class="flex gap-2">
        <button @click="load" class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors"
          :class="app.darkMode ? 'border-slate-600 text-slate-300 hover:bg-slate-700' : 'border-slate-300 text-slate-600 hover:bg-slate-50'">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white">
          <Plus class="w-4 h-4" /> New Transfer
        </button>
      </div>
    </div>

    <!-- KPIs -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div v-for="k in kpis" :key="k.label" class="rounded-xl border p-4" :class="cardCls">
        <div class="text-xs mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ k.label }}</div>
        <div class="text-2xl font-bold" :class="k.color">{{ k.value }}</div>
      </div>
    </div>

    <!-- Filter -->
    <div class="flex gap-3">
      <select v-model="filterStatus" @change="load"
        class="px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500"
        :class="inputCls">
        <option v-for="s in statusOptions" :key="s.value" :value="s.value">{{ s.label }}</option>
      </select>
    </div>

    <!-- Table -->
    <div class="rounded-xl border overflow-hidden" :class="cardCls">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b" :class="thCls">
              <th class="text-left px-4 py-3 font-medium">Transfer #</th>
              <th class="text-left px-4 py-3 font-medium">Asset</th>
              <th class="text-left px-4 py-3 font-medium">From</th>
              <th class="text-left px-4 py-3 font-medium">To</th>
              <th class="text-left px-4 py-3 font-medium">Date</th>
              <th class="text-left px-4 py-3 font-medium">Status</th>
              <th class="text-left px-4 py-3 font-medium">Reason</th>
              <th class="text-right px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading"><td colspan="8" class="text-center py-10 text-slate-400">Loading...</td></tr>
            <tr v-else-if="transfers.length === 0"><td colspan="8" class="text-center py-10 text-slate-400">No transfers found</td></tr>
            <tr v-for="t in transfers" :key="t.id"
              class="border-t transition-colors"
              :class="[tdCls, app.darkMode ? 'hover:bg-slate-700/30' : 'hover:bg-slate-50']">
              <td class="px-4 py-3 font-mono text-indigo-400 font-medium">{{ t.transfer_number }}</td>
              <td class="px-4 py-3">
                <div class="font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">{{ t.asset_name }}</div>
                <div class="text-xs text-slate-400">{{ t.asset_number }}</div>
              </td>
              <td class="px-4 py-3 text-xs">{{ t.from_location_name || '—' }}</td>
              <td class="px-4 py-3 text-xs text-sky-400">{{ t.to_location_name || '—' }}</td>
              <td class="px-4 py-3 text-xs">{{ fmtDate(t.transfer_date) }}</td>
              <td class="px-4 py-3">
                <span class="px-2 py-0.5 rounded-full text-xs font-medium" :class="statusBadge(t.status)">
                  {{ t.status?.replace(/_/g,' ') }}
                </span>
              </td>
              <td class="px-4 py-3 text-xs max-w-[150px] truncate">{{ t.reason || '—' }}</td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <button v-if="t.status === 'pending'" @click="approve(t.id)"
                    class="p-1.5 rounded hover:bg-sky-500/20 text-sky-400 transition-colors" title="Approve">
                    <CheckCircle class="w-4 h-4" />
                  </button>
                  <button v-if="t.status === 'approved'" @click="complete(t.id)"
                    class="p-1.5 rounded hover:bg-emerald-500/20 text-emerald-400 transition-colors" title="Complete">
                    <Truck class="w-4 h-4" />
                  </button>
                  <button v-if="t.status === 'pending'" @click="remove(t.id)"
                    class="p-1.5 rounded hover:bg-red-500/20 text-red-400 transition-colors" title="Delete">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60">
      <div class="w-full max-w-lg rounded-2xl border"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'">
        <div class="flex items-center justify-between p-5 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">New Transfer Request</h2>
          <button @click="showModal = false" class="text-slate-400 hover:text-slate-200"><XCircle class="w-5 h-5" /></button>
        </div>
        <div class="p-5 space-y-4">
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Asset *</label>
            <select v-model="form.asset_id" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
              <option value="">Select asset...</option>
              <option v-for="a in assets" :key="a.id" :value="a.id">{{ a.asset_number }} — {{ a.name }}</option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">From Location</label>
              <select v-model="form.from_location_id" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option value="">Current location</option>
                <option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">To Location *</label>
              <select v-model="form.to_location_id" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option value="">Select location...</option>
                <option v-for="l in locations" :key="l.id" :value="l.id">{{ l.name }}</option>
              </select>
            </div>
          </div>
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Transfer Date</label>
            <input v-model="form.transfer_date" type="date" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
          </div>
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Reason</label>
            <input v-model="form.reason" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" placeholder="Transfer reason..." />
          </div>
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Notes</label>
            <textarea v-model="form.notes" rows="2" class="w-full px-3 py-2 rounded-lg border text-sm outline-none resize-none" :class="inputCls" />
          </div>
        </div>
        <div class="flex justify-end gap-3 p-5 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showModal = false" class="px-4 py-2 rounded-lg text-sm border" :class="app.darkMode ? 'border-slate-600 text-slate-300' : 'border-slate-300 text-slate-600'">Cancel</button>
          <button @click="save" class="px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white">Create Transfer</button>
        </div>
      </div>
    </div>
  </div>
</template>
