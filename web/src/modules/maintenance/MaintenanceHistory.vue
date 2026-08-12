<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  History, Search, RefreshCw, Eye, X, Filter,
  Wrench, CheckCircle, DollarSign, Clock,
  ChevronLeft, ChevronRight, User, Calendar, FileText
} from 'lucide-vue-next'
import { maintenanceAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── types ────────────────────────────────────────────────────────────────────
interface HistoryRecord {
  id: string
  order_id?: string
  order_number?: string
  equipment_id?: string
  equipment_name?: string
  equipment_code?: string
  history_type: string
  performed_date: string
  technician_name?: string
  work_performed?: string
  findings?: string
  downtime_hours?: number
  labor_cost?: number
  parts_cost?: number
  other_cost?: number
  total_cost?: number
  next_service_date?: string
  created_at: string
}

// ─── state ────────────────────────────────────────────────────────────────────
const history    = ref<HistoryRecord[]>([])
const loading    = ref(false)
const search     = ref('')
const typeFilter = ref('all')

const page    = ref(1)
const perPage = 25

const showView = ref(false)
const selected = ref<HistoryRecord | null>(null)

// date range
const dateFrom = ref('')
const dateTo   = ref('')

// ─── computed ────────────────────────────────────────────────────────────────
const dk = (a: string, b: string) => app.darkMode ? a : b

const filtered = computed(() => {
  let list = history.value
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(h =>
      (h.equipment_name || '').toLowerCase().includes(q) ||
      (h.order_number || '').toLowerCase().includes(q) ||
      (h.technician_name || '').toLowerCase().includes(q) ||
      (h.work_performed || '').toLowerCase().includes(q)
    )
  }
  if (typeFilter.value !== 'all') list = list.filter(h => h.history_type === typeFilter.value)
  if (dateFrom.value) list = list.filter(h => h.performed_date >= dateFrom.value)
  if (dateTo.value)   list = list.filter(h => h.performed_date <= dateTo.value + 'T23:59:59')
  return list
})

const paginated   = computed(() => filtered.value.slice((page.value-1)*perPage, page.value*perPage))
const totalPages  = computed(() => Math.max(1, Math.ceil(filtered.value.length / perPage)))

const kpis = computed(() => {
  const all = history.value
  const totalCost    = all.reduce((s,h) => s + (h.total_cost || 0), 0)
  const totalHours   = all.reduce((s,h) => s + (h.downtime_hours || 0), 0)
  const corrective   = all.filter(h => h.history_type === 'corrective').length
  const preventive   = all.filter(h => h.history_type === 'preventive').length
  return [
    { label: 'Total Records',  value: all.length,    sub: 'all time',        icon: History,      color: 'text-violet-400', bg: 'bg-violet-500/10' },
    { label: 'Total Cost',     value: totalCost,      sub: 'all maintenance', icon: DollarSign,   color: 'text-teal-400',   bg: 'bg-teal-500/10',   money: true },
    { label: 'Downtime Hours', value: totalHours,     sub: 'h total',         icon: Clock,        color: 'text-amber-400',  bg: 'bg-amber-500/10',  hours: true },
    { label: 'Corrective',     value: corrective,     sub: 'unplanned',       icon: Wrench,       color: 'text-rose-400',   bg: 'bg-rose-500/10' },
    { label: 'Preventive',     value: preventive,     sub: 'scheduled',       icon: CheckCircle,  color: 'text-emerald-400',bg: 'bg-emerald-500/10' },
  ]
})

const typeCounts = computed(() => {
  const types = ['corrective','preventive','inspection','emergency','upgrade','other']
  return types.map(t => ({ type: t, count: history.value.filter(h => h.history_type === t).length }))
    .filter(t => t.count > 0)
})

// ─── helpers ─────────────────────────────────────────────────────────────────
const fmt = (n: number) =>
  new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'

const fmtDate = (s?: string) => {
  if (!s) return '—'
  return new Date(s).toLocaleDateString('fr-DZ', { day:'2-digit', month:'short', year:'numeric' })
}

const typeColor = (t: string) => ({
  corrective:'#f43f5e', preventive:'#10b981',
  inspection:'#8b5cf6', emergency:'#f97316',
  upgrade:'#f59e0b', other:'#64748b',
}[t] ?? '#64748b')

const typeBadge = (t: string) => ({
  corrective:'bg-rose-500/15 text-rose-400',
  preventive:'bg-emerald-500/15 text-emerald-400',
  inspection:'bg-violet-500/15 text-violet-400',
  emergency:'bg-orange-500/15 text-orange-400',
  upgrade:'bg-amber-500/15 text-amber-400',
  other:'bg-slate-500/15 text-slate-400',
}[t] ?? 'bg-slate-500/15 text-slate-400')

const typeLabel = (t: string) => ({
  corrective:'Corrective', preventive:'Preventive',
  inspection:'Inspection', emergency:'Emergency',
  upgrade:'Upgrade', other:'Other',
}[t] ?? t)

// ─── data loading ─────────────────────────────────────────────────────────────
const load = async () => {
  loading.value = true
  try {
    const params: Record<string,string> = {}
    if (typeFilter.value !== 'all') params.type = typeFilter.value
    if (dateFrom.value) params.date_from = dateFrom.value
    if (dateTo.value)   params.date_to   = dateTo.value
    const res = await maintenanceAPI.listHistory(params)
    history.value = res.data.history ?? res.data ?? []
  } catch {
    app.addToast('Failed to load maintenance history', 'error')
  } finally {
    loading.value = false
  }
}

const openView = (h: HistoryRecord) => {
  selected.value = h
  showView.value = true
}

onMounted(load)
</script>

<template>
  <div :class="['min-h-screen p-6 space-y-6', dk('bg-slate-950 text-slate-100','bg-slate-50 text-slate-900')]">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-violet-500/15 flex items-center justify-center">
          <History class="w-5 h-5 text-violet-400" />
        </div>
        <div>
          <h1 class="text-xl font-bold">Maintenance History</h1>
          <p :class="['text-sm', dk('text-slate-400','text-slate-500')]">Complete log of all maintenance activities</p>
        </div>
      </div>
      <button @click="load"
        :class="['flex items-center gap-2 px-4 py-2 rounded-xl border text-sm transition-colors',
          dk('bg-slate-900 border-slate-700 text-slate-300 hover:bg-slate-800',
             'bg-white border-slate-200 text-slate-700 hover:bg-slate-50')]">
        <RefreshCw :class="['w-4 h-4', loading && 'animate-spin']" /> Refresh
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
      <div v-for="k in kpis" :key="k.label"
        :class="['rounded-xl border p-4', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
        <div class="flex items-center justify-between mb-2">
          <span :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]">{{ k.label }}</span>
          <div :class="['w-8 h-8 rounded-lg flex items-center justify-center', k.bg]">
            <component :is="k.icon" :class="['w-4 h-4', k.color]" />
          </div>
        </div>
        <div class="text-xl font-bold">
          <template v-if="k.money">{{ fmt(k.value) }}</template>
          <template v-else-if="k.hours">{{ k.value.toFixed(1) }}h</template>
          <template v-else>{{ k.value }}</template>
        </div>
        <div :class="['text-xs mt-0.5', dk('text-slate-500','text-slate-400')]">{{ k.sub }}</div>
      </div>
    </div>

    <!-- Type Pills -->
    <div class="flex flex-wrap gap-2">
      <button @click="typeFilter='all'; page=1"
        :class="['px-3 py-1.5 rounded-lg text-xs font-medium border transition-colors',
          typeFilter==='all'
            ? 'bg-violet-600 text-white border-violet-600'
            : dk('bg-slate-900 border-slate-700 text-slate-400 hover:border-slate-500',
                 'bg-white border-slate-200 text-slate-500 hover:border-slate-400')]">
        All ({{ history.length }})
      </button>
      <button v-for="t in typeCounts" :key="t.type"
        @click="typeFilter=t.type; page=1"
        :class="['px-3 py-1.5 rounded-lg text-xs font-medium border transition-colors',
          typeFilter===t.type
            ? 'text-white border-transparent'
            : dk('bg-slate-900 border-slate-700 text-slate-400 hover:border-slate-500',
                 'bg-white border-slate-200 text-slate-500 hover:border-slate-400')]"
        :style="typeFilter===t.type ? { backgroundColor: typeColor(t.type) } : {}">
        {{ typeLabel(t.type) }} ({{ t.count }})
      </button>
    </div>

    <!-- Filters -->
    <div :class="['rounded-xl border p-4', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
      <div class="flex flex-col sm:flex-row gap-3">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="search" placeholder="Search equipment, technician, work performed…"
            :class="['w-full pl-9 pr-4 py-2 rounded-lg border text-sm',
              dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500 focus:border-violet-500',
                 'bg-slate-50 border-slate-200 text-slate-900 placeholder-slate-400 focus:border-violet-500')]"
            @input="page=1" />
        </div>
        <div class="flex gap-2 items-center">
          <Calendar class="w-4 h-4 text-slate-400 flex-shrink-0" />
          <input v-model="dateFrom" type="date"
            :class="['px-3 py-2 rounded-lg border text-sm',
              dk('bg-slate-800 border-slate-700 text-slate-100','bg-slate-50 border-slate-200 text-slate-900')]"
            @change="page=1; load()" />
          <span :class="['text-xs', dk('text-slate-500','text-slate-400')]">to</span>
          <input v-model="dateTo" type="date"
            :class="['px-3 py-2 rounded-lg border text-sm',
              dk('bg-slate-800 border-slate-700 text-slate-100','bg-slate-50 border-slate-200 text-slate-900')]"
            @change="page=1; load()" />
        </div>
      </div>
    </div>

    <!-- Table -->
    <div :class="['rounded-xl border overflow-hidden', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
      <div v-if="loading" class="p-12 text-center">
        <RefreshCw class="w-8 h-8 animate-spin text-violet-400 mx-auto mb-3" />
        <p :class="['text-sm', dk('text-slate-400','text-slate-500')]">Loading history…</p>
      </div>
      <div v-else-if="!filtered.length" class="p-12 text-center">
        <History class="w-10 h-10 mx-auto mb-3 text-slate-500" />
        <p :class="['font-medium', dk('text-slate-300','text-slate-700')]">No history records found</p>
        <p :class="['text-sm mt-1', dk('text-slate-500','text-slate-400')]">Complete maintenance orders to build history</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="['border-b', dk('border-slate-800 bg-slate-800/50','border-slate-100 bg-slate-50')]">
              <th class="text-left px-4 py-3 font-medium text-slate-400 w-4"></th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Date</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Equipment</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Type</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Work Performed</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Technician</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Downtime</th>
              <th class="text-left px-4 py-3 font-medium text-slate-400">Cost</th>
              <th class="text-right px-4 py-3 font-medium text-slate-400"></th>
            </tr>
          </thead>
          <tbody class="divide-y" :class="dk('divide-slate-800','divide-slate-100')">
            <tr v-for="h in paginated" :key="h.id"
              :class="['transition-colors', dk('hover:bg-slate-800/40','hover:bg-slate-50')]">
              <td class="px-4 py-3">
                <div class="w-1 h-8 rounded-full" :style="{ backgroundColor: typeColor(h.history_type) }"></div>
              </td>
              <td class="px-4 py-3">
                <div class="font-medium text-sm">{{ fmtDate(h.performed_date) }}</div>
                <div v-if="h.order_number" class="font-mono text-xs text-violet-400 mt-0.5">{{ h.order_number }}</div>
              </td>
              <td class="px-4 py-3">
                <div v-if="h.equipment_name" class="font-medium text-sm">{{ h.equipment_name }}</div>
                <div v-if="h.equipment_code" :class="['text-xs', dk('text-slate-500','text-slate-400')]">{{ h.equipment_code }}</div>
                <div v-if="!h.equipment_name" :class="['text-xs', dk('text-slate-600','text-slate-400')]">—</div>
              </td>
              <td class="px-4 py-3">
                <span :class="['px-2 py-0.5 rounded-md text-xs font-medium', typeBadge(h.history_type)]">
                  {{ typeLabel(h.history_type) }}
                </span>
              </td>
              <td class="px-4 py-3 max-w-xs">
                <p :class="['text-sm line-clamp-2', dk('text-slate-300','text-slate-700')]">
                  {{ h.work_performed || '—' }}
                </p>
              </td>
              <td class="px-4 py-3">
                <div v-if="h.technician_name" class="flex items-center gap-1.5 text-sm">
                  <User class="w-3.5 h-3.5 text-slate-400" />
                  {{ h.technician_name }}
                </div>
                <div v-else :class="['text-xs', dk('text-slate-600','text-slate-400')]">—</div>
              </td>
              <td class="px-4 py-3">
                <span v-if="h.downtime_hours" :class="['text-sm', h.downtime_hours > 8 ? 'text-rose-400' : '']">
                  {{ h.downtime_hours }}h
                </span>
                <span v-else :class="['text-xs', dk('text-slate-600','text-slate-400')]">—</span>
              </td>
              <td class="px-4 py-3">
                <span class="text-sm font-medium">{{ h.total_cost ? fmt(h.total_cost) : '—' }}</span>
              </td>
              <td class="px-4 py-3 text-right">
                <button @click="openView(h)"
                  :class="['p-1.5 rounded-lg transition-colors', dk('hover:bg-slate-700 text-slate-400','hover:bg-slate-100 text-slate-500')]">
                  <Eye class="w-3.5 h-3.5" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="!loading && filtered.length > perPage"
        :class="['flex items-center justify-between px-4 py-3 border-t text-sm', dk('border-slate-800','border-slate-200')]">
        <span :class="dk('text-slate-400','text-slate-500')">
          {{ (page-1)*perPage+1 }}–{{ Math.min(page*perPage, filtered.length) }} of {{ filtered.length }}
        </span>
        <div class="flex items-center gap-1">
          <button @click="page--" :disabled="page===1"
            :class="['p-1.5 rounded-lg disabled:opacity-40 transition-colors',
              dk('hover:bg-slate-700 text-slate-400','hover:bg-slate-100 text-slate-500')]">
            <ChevronLeft class="w-4 h-4" />
          </button>
          <span :class="['px-3 py-1 rounded-lg text-xs', dk('bg-slate-800','bg-slate-100')]">{{ page }}/{{ totalPages }}</span>
          <button @click="page++" :disabled="page===totalPages"
            :class="['p-1.5 rounded-lg disabled:opacity-40 transition-colors',
              dk('hover:bg-slate-700 text-slate-400','hover:bg-slate-100 text-slate-500')]">
            <ChevronRight class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>

    <!-- ── View Modal ──────────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showView && selected"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
        @click.self="showView=false">
        <div :class="['w-full max-w-lg rounded-2xl border shadow-2xl', dk('bg-slate-900 border-slate-700','bg-white border-slate-200')]">
          <div class="h-1.5 rounded-t-2xl" :style="{ backgroundColor: typeColor(selected.history_type) }"></div>
          <div :class="['flex items-center justify-between px-6 py-4 border-b', dk('border-slate-800','border-slate-200')]">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl flex items-center justify-center"
                :style="{ backgroundColor: typeColor(selected.history_type) + '22' }">
                <History class="w-4 h-4" :style="{ color: typeColor(selected.history_type) }" />
              </div>
              <div>
                <h2 class="font-semibold text-sm">Maintenance Record</h2>
                <p :class="['text-xs', dk('text-slate-400','text-slate-500')]">{{ fmtDate(selected.performed_date) }}</p>
              </div>
            </div>
            <button @click="showView=false"
              :class="['p-1.5 rounded-lg', dk('hover:bg-slate-800 text-slate-400','hover:bg-slate-100 text-slate-500')]">
              <X class="w-4 h-4" />
            </button>
          </div>
          <div class="p-6 space-y-4 max-h-[65vh] overflow-y-auto">
            <!-- Badges -->
            <div class="flex flex-wrap gap-2">
              <span :class="['px-2.5 py-1 rounded-lg text-xs font-medium', typeBadge(selected.history_type)]">
                {{ typeLabel(selected.history_type) }}
              </span>
              <span v-if="selected.order_number" class="font-mono text-xs px-2.5 py-1 rounded-lg bg-violet-500/15 text-violet-400">
                {{ selected.order_number }}
              </span>
            </div>
            <!-- Grid details -->
            <div class="grid grid-cols-2 gap-4 text-sm">
              <div v-if="selected.equipment_name">
                <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Equipment</span>
                <p class="font-medium mt-0.5">{{ selected.equipment_name }}</p>
              </div>
              <div v-if="selected.technician_name">
                <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Technician</span>
                <p class="font-medium mt-0.5">{{ selected.technician_name }}</p>
              </div>
              <div v-if="selected.downtime_hours">
                <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Downtime</span>
                <p class="font-medium mt-0.5">{{ selected.downtime_hours }}h</p>
              </div>
              <div v-if="selected.next_service_date">
                <span :class="['text-xs', dk('text-slate-400','text-slate-500')]">Next Service</span>
                <p class="font-medium mt-0.5">{{ fmtDate(selected.next_service_date) }}</p>
              </div>
            </div>
            <!-- Cost -->
            <div :class="['rounded-xl border p-4', dk('border-slate-700 bg-slate-800/50','border-slate-200 bg-slate-50')]">
              <div class="text-xs font-medium mb-3 flex items-center gap-2">
                <DollarSign class="w-3.5 h-3.5 text-teal-400" /> Cost Breakdown
              </div>
              <div class="grid grid-cols-4 gap-3 text-xs">
                <div>
                  <span :class="dk('text-slate-400','text-slate-500')">Labor</span>
                  <p class="font-medium mt-0.5">{{ selected.labor_cost ? fmt(selected.labor_cost) : '—' }}</p>
                </div>
                <div>
                  <span :class="dk('text-slate-400','text-slate-500')">Parts</span>
                  <p class="font-medium mt-0.5">{{ selected.parts_cost ? fmt(selected.parts_cost) : '—' }}</p>
                </div>
                <div>
                  <span :class="dk('text-slate-400','text-slate-500')">Other</span>
                  <p class="font-medium mt-0.5">{{ selected.other_cost ? fmt(selected.other_cost) : '—' }}</p>
                </div>
                <div>
                  <span :class="dk('text-slate-400','text-slate-500')">Total</span>
                  <p class="font-semibold mt-0.5 text-teal-400">{{ selected.total_cost ? fmt(selected.total_cost) : '—' }}</p>
                </div>
              </div>
            </div>
            <!-- Work performed -->
            <div v-if="selected.work_performed">
              <span :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]">Work Performed</span>
              <p :class="['mt-1.5 text-sm leading-relaxed p-3 rounded-lg', dk('bg-slate-800 text-slate-300','bg-slate-50 text-slate-700')]">
                {{ selected.work_performed }}
              </p>
            </div>
            <div v-if="selected.findings">
              <span :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]">Findings</span>
              <p :class="['mt-1.5 text-sm leading-relaxed p-3 rounded-lg', dk('bg-slate-800 text-slate-300','bg-slate-50 text-slate-700')]">
                {{ selected.findings }}
              </p>
            </div>
          </div>
          <div :class="['px-6 py-4 border-t flex justify-end', dk('border-slate-800','border-slate-200')]">
            <button @click="showView=false"
              :class="['px-4 py-2 rounded-lg border text-sm',
                dk('border-slate-700 text-slate-300 hover:bg-slate-800','border-slate-200 text-slate-600 hover:bg-slate-50')]">
              Close
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
