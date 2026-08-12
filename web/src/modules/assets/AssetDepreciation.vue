<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { assetsAPI } from '@/api/client'
import {
  TrendingDown, RefreshCw, Play, CheckSquare,
  XCircle, AlertCircle, Calendar
} from 'lucide-vue-next'

const app = useAppStore()
const schedules = ref<any[]>([])
const assets = ref<any[]>([])
const loading = ref(false)
const generating = ref(false)
const posting = ref(false)
const filterYear = ref(new Date().getFullYear().toString())
const filterPosted = ref('')
const filterAsset = ref('')
const selectedIds = ref<string[]>([])
const genMonth = ref(new Date().getMonth() + 1)
const genYear = ref(new Date().getFullYear())

async function load() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (filterYear.value) params.year = filterYear.value
    if (filterPosted.value !== '') params.posted = filterPosted.value
    if (filterAsset.value) params.asset_id = filterAsset.value
    const [sRes, aRes] = await Promise.all([
      assetsAPI.listDepreciation(params),
      assetsAPI.listAssets(),
    ])
    schedules.value = sRes.data
    assets.value = aRes.data
  } finally {
    loading.value = false
  }
}

onMounted(load)

async function generate() {
  generating.value = true
  try {
    const res = await assetsAPI.generateDepreciation({ year: genYear.value, month: genMonth.value })
    alert(res.data.message)
    load()
  } finally {
    generating.value = false
  }
}

async function postSelected() {
  if (selectedIds.value.length === 0) {
    alert('Select schedules to post')
    return
  }
  posting.value = true
  try {
    const res = await assetsAPI.postDepreciation({ ids: selectedIds.value })
    alert(res.data.message)
    selectedIds.value = []
    load()
  } finally {
    posting.value = false
  }
}

async function postAll() {
  if (!confirm(`Post all unposted depreciation for ${genMonth.value}/${genYear.value}?`)) return
  posting.value = true
  try {
    const res = await assetsAPI.postDepreciation({ year: genYear.value, month: genMonth.value })
    alert(res.data.message)
    load()
  } finally {
    posting.value = false
  }
}

function toggleSelect(id: string) {
  const idx = selectedIds.value.indexOf(id)
  if (idx > -1) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

function toggleAll() {
  const unposted = schedules.value.filter(s => !s.is_posted)
  if (selectedIds.value.length === unposted.length) {
    selectedIds.value = []
  } else {
    selectedIds.value = unposted.map(s => s.id)
  }
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

const totalDep = computed(() => schedules.value.reduce((s, r) => s + (r.depreciation_amount || 0), 0))
const postedCount = computed(() => schedules.value.filter(s => s.is_posted).length)
const unpostedCount = computed(() => schedules.value.filter(s => !s.is_posted).length)
const years = computed(() => {
  const yr = new Date().getFullYear()
  return Array.from({ length: 5 }, (_, i) => yr - 2 + i)
})
const months = [
  { v: 1, l: 'January' }, { v: 2, l: 'February' }, { v: 3, l: 'March' },
  { v: 4, l: 'April' }, { v: 5, l: 'May' }, { v: 6, l: 'June' },
  { v: 7, l: 'July' }, { v: 8, l: 'August' }, { v: 9, l: 'September' },
  { v: 10, l: 'October' }, { v: 11, l: 'November' }, { v: 12, l: 'December' },
]
</script>

<template>
  <div class="p-6 space-y-5">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Asset Depreciation</h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Generate and post monthly depreciation schedules</p>
      </div>
      <button @click="load" class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors"
        :class="app.darkMode ? 'border-slate-600 text-slate-300 hover:bg-slate-700' : 'border-slate-300 text-slate-600 hover:bg-slate-50'">
        <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
      </button>
    </div>

    <!-- KPIs -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div v-for="stat in [
        { label: 'Total Entries', value: schedules.length, color: 'text-indigo-400', unit: '' },
        { label: 'Posted', value: postedCount, color: 'text-emerald-400', unit: '' },
        { label: 'Unposted', value: unpostedCount, color: 'text-amber-400', unit: '' },
        { label: 'Total Dep. Amount', value: fmtCurrency(totalDep), color: 'text-red-400', unit: '' },
      ]" :key="stat.label" class="rounded-xl border p-4" :class="cardCls">
        <div class="text-xs mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ stat.label }}</div>
        <div class="text-xl font-bold" :class="stat.color">{{ stat.value }}</div>
      </div>
    </div>

    <!-- Generate + Post panel -->
    <div class="rounded-xl border p-5" :class="cardCls">
      <h2 class="text-sm font-semibold mb-4" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Generate / Post Depreciation</h2>
      <div class="flex flex-wrap items-end gap-4">
        <div>
          <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Year</label>
          <select v-model.number="genYear" class="px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls">
            <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
          </select>
        </div>
        <div>
          <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Month</label>
          <select v-model.number="genMonth" class="px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls">
            <option v-for="m in months" :key="m.v" :value="m.v">{{ m.l }}</option>
          </select>
        </div>
        <button @click="generate" :disabled="generating"
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white disabled:opacity-50">
          <Play class="w-4 h-4" />
          {{ generating ? 'Generating...' : 'Generate' }}
        </button>
        <button @click="postAll" :disabled="posting"
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-emerald-600 hover:bg-emerald-700 text-white disabled:opacity-50">
          <CheckSquare class="w-4 h-4" />
          {{ posting ? 'Posting...' : 'Post All for Period' }}
        </button>
        <button v-if="selectedIds.length > 0" @click="postSelected" :disabled="posting"
          class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-sky-600 hover:bg-sky-700 text-white disabled:opacity-50">
          <CheckSquare class="w-4 h-4" />
          Post Selected ({{ selectedIds.length }})
        </button>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3">
      <select v-model="filterYear" @change="load" class="px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls">
        <option value="">All Years</option>
        <option v-for="y in years" :key="y" :value="y.toString()">{{ y }}</option>
      </select>
      <select v-model="filterPosted" @change="load" class="px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls">
        <option value="">All</option>
        <option value="false">Unposted</option>
        <option value="true">Posted</option>
      </select>
      <select v-model="filterAsset" @change="load" class="px-3 py-2 rounded-lg border text-sm outline-none" :class="inputCls">
        <option value="">All Assets</option>
        <option v-for="a in assets" :key="a.id" :value="a.id">{{ a.asset_number }} — {{ a.name }}</option>
      </select>
    </div>

    <!-- Table -->
    <div class="rounded-xl border overflow-hidden" :class="cardCls">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b" :class="thCls">
              <th class="px-4 py-3">
                <input type="checkbox" @change="toggleAll" class="rounded" />
              </th>
              <th class="text-left px-4 py-3 font-medium">Asset</th>
              <th class="text-center px-4 py-3 font-medium">Period</th>
              <th class="text-right px-4 py-3 font-medium">Opening NBV</th>
              <th class="text-right px-4 py-3 font-medium">Dep. Amount</th>
              <th class="text-right px-4 py-3 font-medium">Accum. Dep.</th>
              <th class="text-right px-4 py-3 font-medium">Closing NBV</th>
              <th class="text-center px-4 py-3 font-medium">Posted</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading"><td colspan="8" class="text-center py-10 text-slate-400">Loading...</td></tr>
            <tr v-else-if="schedules.length === 0"><td colspan="8" class="text-center py-10 text-slate-400">No depreciation schedules found</td></tr>
            <tr v-for="s in schedules" :key="s.id"
              class="border-t transition-colors"
              :class="[tdCls, app.darkMode ? 'hover:bg-slate-700/30' : 'hover:bg-slate-50', s.is_posted ? 'opacity-60' : '']">
              <td class="px-4 py-3 text-center">
                <input v-if="!s.is_posted" type="checkbox"
                  :checked="selectedIds.includes(s.id)"
                  @change="toggleSelect(s.id)"
                  class="rounded" />
              </td>
              <td class="px-4 py-3">
                <div class="font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">{{ s.asset_name }}</div>
                <div class="text-xs text-slate-400">{{ s.asset_number }}</div>
              </td>
              <td class="px-4 py-3 text-center font-mono text-xs">
                {{ s.period_year }}-{{ String(s.period_month).padStart(2,'0') }}
              </td>
              <td class="px-4 py-3 text-right font-mono">{{ fmtCurrency(s.opening_book_value) }}</td>
              <td class="px-4 py-3 text-right font-mono text-amber-400 font-medium">{{ fmtCurrency(s.depreciation_amount) }}</td>
              <td class="px-4 py-3 text-right font-mono">{{ fmtCurrency(s.accumulated_depreciation) }}</td>
              <td class="px-4 py-3 text-right font-mono text-emerald-400">{{ fmtCurrency(s.closing_book_value) }}</td>
              <td class="px-4 py-3 text-center">
                <span v-if="s.is_posted" class="px-2 py-0.5 rounded-full text-xs bg-emerald-500/20 text-emerald-300">Posted</span>
                <span v-else class="px-2 py-0.5 rounded-full text-xs bg-amber-500/20 text-amber-300">Pending</span>
              </td>
            </tr>
          </tbody>
          <tfoot v-if="schedules.length > 0">
            <tr class="border-t font-medium" :class="thCls">
              <td colspan="3" class="px-4 py-3 text-right text-sm">Totals</td>
              <td class="px-4 py-3 text-right font-mono text-sm">—</td>
              <td class="px-4 py-3 text-right font-mono text-amber-400 text-sm">{{ fmtCurrency(totalDep) }}</td>
              <td colspan="3" />
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  </div>
</template>
