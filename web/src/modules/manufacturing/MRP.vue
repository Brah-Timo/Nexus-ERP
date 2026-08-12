<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  RefreshCw, TrendingUp, AlertTriangle, CheckCircle, Package,
  Cog, CalendarClock, ShoppingCart, ChevronDown, ChevronUp,
  BarChart2, Search, Download, Info
} from 'lucide-vue-next'
import { manufacturingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ────────────────────────────────────────────────────────────────────
interface MRPLine {
  mo_id: string
  mo_number: string
  product_id: string
  product_code: string
  product_name: string
  planned_qty: number
  planned_start: string | null
  component_id: string
  component_code: string
  component_name: string
  required_qty: number
  available_qty: number
  shortage_qty: number
  needs_to_purchase: boolean
}

interface MRPResult {
  lines: MRPLine[]
  total_lines: number
  total_shortages: number
  generated_at: string
}

interface DashStats {
  total_orders: number
  draft_orders: number
  in_progress_orders: number
  completed_orders: number
  total_boms: number
  active_work_centers: number
  total_material_cost: number
}

// ─── State ────────────────────────────────────────────────────────────────────
const mrpResult = ref<MRPResult | null>(null)
const dashStats = ref<DashStats | null>(null)
const recentOrders = ref<any[]>([])

const loading = ref(false)
const loadingDash = ref(false)
const running = ref(false)

const search = ref('')
const filterShortages = ref(false)
const sortField = ref('mo_number')
const sortDir = ref<'asc' | 'desc'>('asc')

// Group by MO
const groupByMO = ref(true)

// ─── Computed ─────────────────────────────────────────────────────────────────
const filteredLines = computed(() => {
  if (!mrpResult.value) return []
  let lines = [...mrpResult.value.lines]
  if (filterShortages.value) lines = lines.filter(l => l.shortage_qty > 0)
  if (search.value) {
    const q = search.value.toLowerCase()
    lines = lines.filter(l =>
      l.mo_number.toLowerCase().includes(q) ||
      l.product_name.toLowerCase().includes(q) ||
      l.component_name.toLowerCase().includes(q) ||
      l.component_code.toLowerCase().includes(q)
    )
  }
  lines.sort((a, b) => {
    const av = (a as any)[sortField.value] ?? ''
    const bv = (b as any)[sortField.value] ?? ''
    const cmp = String(av).localeCompare(String(bv))
    return sortDir.value === 'asc' ? cmp : -cmp
  })
  return lines
})

// Group lines by MO for grouped view
const groupedLines = computed(() => {
  const groups = new Map<string, { moNumber: string; productName: string; productCode: string; plannedQty: number; plannedStart: string | null; lines: MRPLine[]; hasShortage: boolean }>()
  for (const line of filteredLines.value) {
    if (!groups.has(line.mo_id)) {
      groups.set(line.mo_id, {
        moNumber: line.mo_number,
        productName: line.product_name,
        productCode: line.product_code,
        plannedQty: line.planned_qty,
        plannedStart: line.planned_start,
        lines: [],
        hasShortage: false
      })
    }
    const group = groups.get(line.mo_id)!
    group.lines.push(line)
    if (line.shortage_qty > 0) group.hasShortage = true
  }
  return Array.from(groups.values())
})

const shortageLines = computed(() => filteredLines.value.filter(l => l.shortage_qty > 0))
const okLines = computed(() => filteredLines.value.filter(l => l.shortage_qty === 0))

const expandedGroups = ref(new Set<string>())

function toggleGroup(moId: string) {
  if (expandedGroups.value.has(moId)) expandedGroups.value.delete(moId)
  else expandedGroups.value.add(moId)
}

function expandAll() {
  const groups = groupedLines.value
  groups.forEach(g => expandedGroups.value.add(g.moNumber))
}

function collapseAll() {
  expandedGroups.value.clear()
}

// ─── Load ─────────────────────────────────────────────────────────────────────
async function loadDashboard() {
  loadingDash.value = true
  try {
    const res = await manufacturingAPI.getDashboard()
    dashStats.value = res.data.stats
    recentOrders.value = res.data.recent_orders || []
  } catch (e: any) {
    // non-critical
  } finally {
    loadingDash.value = false
  }
}

async function runMRP() {
  running.value = true
  try {
    const res = await manufacturingAPI.runMRP()
    mrpResult.value = res.data
    app.addToast(`MRP computed: ${res.data.total_shortages} shortages found`, 'success')
    // auto-expand groups with shortages
    expandedGroups.value.clear()
    if (res.data.lines) {
      for (const line of res.data.lines) {
        if (line.shortage_qty > 0) expandedGroups.value.add(line.mo_number)
      }
    }
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'MRP run failed', 'error')
  } finally {
    running.value = false
  }
}

onMounted(async () => {
  await loadDashboard()
  await runMRP()
})

function setSort(field: string) {
  if (sortField.value === field) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortField.value = field; sortDir.value = 'asc' }
}

// ─── Helpers ──────────────────────────────────────────────────────────────────
function fmtNum(n: number, dec = 2) {
  return new Intl.NumberFormat('en-US', { minimumFractionDigits: dec, maximumFractionDigits: dec }).format(n)
}
function fmtDate(d: string | null) {
  return d ? new Date(d).toLocaleDateString('en-GB') : '—'
}
function fmtDateTime(d: string) {
  return d ? new Date(d).toLocaleString('en-GB', { dateStyle: 'medium', timeStyle: 'short' }) : '—'
}

const statusConfig: Record<string, string> = {
  draft:       'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300',
  planned:     'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400',
  in_progress: 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400',
  completed:   'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400',
  cancelled:   'bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400'
}
</script>

<template>
  <div class="flex flex-col h-full gap-4 p-4 bg-slate-50 dark:bg-slate-950 min-h-screen">

    <!-- ── Header ──────────────────────────────────────────────────────────── -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
          <CalendarClock class="w-6 h-6 text-indigo-600" />
          MRP Planning
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Material Requirements Planning — component shortage analysis
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="loadDashboard" :disabled="loadingDash"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300
                 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loadingDash }" />
        </button>
        <button @click="runMRP" :disabled="running"
          class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50
                 text-white rounded-lg text-sm font-medium transition-colors shadow-sm">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': running }" />
          {{ running ? 'Running MRP...' : 'Run MRP' }}
        </button>
      </div>
    </div>

    <!-- ── Dashboard KPI Cards ─────────────────────────────────────────────── -->
    <div v-if="dashStats" class="grid grid-cols-2 lg:grid-cols-6 gap-3">
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total MOs</p>
        <p class="text-2xl font-bold text-slate-900 dark:text-white mt-1">{{ dashStats.total_orders }}</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Draft</p>
        <p class="text-2xl font-bold text-slate-600 dark:text-slate-400 mt-1">{{ dashStats.draft_orders }}</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">In Progress</p>
        <p class="text-2xl font-bold text-amber-600 mt-1">{{ dashStats.in_progress_orders }}</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Completed</p>
        <p class="text-2xl font-bold text-emerald-600 mt-1">{{ dashStats.completed_orders }}</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Active BOMs</p>
        <p class="text-2xl font-bold text-indigo-600 mt-1">{{ dashStats.total_boms }}</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Work Centers</p>
        <p class="text-2xl font-bold text-violet-600 mt-1">{{ dashStats.active_work_centers }}</p>
      </div>
    </div>

    <!-- ── MRP Summary ─────────────────────────────────────────────────────── -->
    <div v-if="mrpResult" class="grid grid-cols-3 gap-4">
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4 flex items-center gap-4">
        <div class="p-2.5 bg-indigo-100 dark:bg-indigo-900/30 rounded-xl">
          <BarChart2 class="w-5 h-5 text-indigo-600" />
        </div>
        <div>
          <p class="text-xs text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total Lines</p>
          <p class="text-xl font-bold text-slate-900 dark:text-white">{{ mrpResult.total_lines }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-red-200 dark:border-red-800/50 p-4 flex items-center gap-4">
        <div class="p-2.5 bg-red-100 dark:bg-red-900/30 rounded-xl">
          <AlertTriangle class="w-5 h-5 text-red-600" />
        </div>
        <div>
          <p class="text-xs text-slate-500 dark:text-slate-400 uppercase tracking-wide">Shortages</p>
          <p class="text-xl font-bold text-red-600">{{ mrpResult.total_shortages }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-emerald-200 dark:border-emerald-800/50 p-4 flex items-center gap-4">
        <div class="p-2.5 bg-emerald-100 dark:bg-emerald-900/30 rounded-xl">
          <CheckCircle class="w-5 h-5 text-emerald-600" />
        </div>
        <div>
          <p class="text-xs text-slate-500 dark:text-slate-400 uppercase tracking-wide">Available</p>
          <p class="text-xl font-bold text-emerald-600">{{ mrpResult.total_lines - mrpResult.total_shortages }}</p>
        </div>
      </div>
    </div>

    <!-- ── Generated at ────────────────────────────────────────────────────── -->
    <div v-if="mrpResult" class="flex items-center gap-2 text-xs text-slate-500 dark:text-slate-400">
      <Info class="w-3.5 h-3.5" />
      MRP last run: {{ fmtDateTime(mrpResult.generated_at) }}
    </div>

    <!-- ── MRP Lines Table ─────────────────────────────────────────────────── -->
    <div v-if="mrpResult"
      class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 flex-1 flex flex-col overflow-hidden">

      <!-- Toolbar -->
      <div class="flex flex-wrap items-center gap-3 p-4 border-b border-slate-200 dark:border-slate-700">
        <div class="relative flex-1 min-w-[200px]">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="search" type="text" placeholder="Search MO, product, component..."
            class="w-full pl-9 pr-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                   bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-indigo-500 outline-none" />
        </div>
        <label class="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-300 cursor-pointer">
          <input type="checkbox" v-model="filterShortages"
            class="w-4 h-4 rounded border-slate-300 text-red-600 focus:ring-red-500" />
          Shortages only
        </label>
        <label class="flex items-center gap-2 text-sm text-slate-600 dark:text-slate-300 cursor-pointer">
          <input type="checkbox" v-model="groupByMO"
            class="w-4 h-4 rounded border-slate-300 text-indigo-600 focus:ring-indigo-500" />
          Group by MO
        </label>
        <div v-if="groupByMO" class="flex items-center gap-1">
          <button @click="expandAll"
            class="px-2.5 py-1.5 text-xs text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
            Expand All
          </button>
          <button @click="collapseAll"
            class="px-2.5 py-1.5 text-xs text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
            Collapse
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="running" class="flex-1 flex items-center justify-center py-12">
        <div class="text-center">
          <RefreshCw class="w-8 h-8 animate-spin text-indigo-600 mx-auto mb-3" />
          <p class="text-sm text-slate-500">Running Material Requirements Planning...</p>
        </div>
      </div>

      <!-- Empty -->
      <div v-else-if="filteredLines.length === 0" class="flex-1 flex items-center justify-center py-12">
        <div class="text-center">
          <CalendarClock class="w-10 h-10 mx-auto mb-3 text-slate-300 dark:text-slate-600" />
          <p class="text-sm font-medium text-slate-500 dark:text-slate-400">
            {{ mrpResult.total_lines === 0 ? 'No planned or draft orders found' : 'No results match your filters' }}
          </p>
          <p v-if="mrpResult.total_lines === 0" class="text-xs text-slate-400 dark:text-slate-500 mt-1">
            Create Manufacturing Orders in draft or planned status to see MRP data
          </p>
        </div>
      </div>

      <!-- Grouped View -->
      <div v-else-if="groupByMO" class="flex-1 overflow-auto">
        <div v-for="group in groupedLines" :key="group.moNumber" class="border-b border-slate-100 dark:border-slate-800">
          <!-- Group header -->
          <button
            class="w-full flex items-center gap-3 px-4 py-3 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors text-left"
            @click="toggleGroup(group.moNumber)">
            <ChevronDown v-if="expandedGroups.has(group.moNumber)" class="w-4 h-4 text-slate-400 shrink-0" />
            <ChevronUp v-else class="w-4 h-4 text-slate-400 shrink-0 rotate-180" />

            <div class="flex-1 flex items-center gap-4">
              <span class="font-mono font-semibold text-indigo-700 dark:text-indigo-400">{{ group.moNumber }}</span>
              <div>
                <span class="text-sm font-medium text-slate-900 dark:text-white">{{ group.productName }}</span>
                <span class="text-xs text-slate-400 ml-2">{{ group.productCode }}</span>
              </div>
              <div class="flex items-center gap-1 text-xs text-slate-500 dark:text-slate-400">
                <Cog class="w-3.5 h-3.5" />
                Qty: {{ fmtNum(group.plannedQty, 0) }}
              </div>
              <div v-if="group.plannedStart" class="text-xs text-slate-500 dark:text-slate-400">
                Start: {{ fmtDate(group.plannedStart) }}
              </div>
            </div>

            <div class="flex items-center gap-2 shrink-0">
              <span v-if="group.hasShortage"
                class="flex items-center gap-1 px-2 py-0.5 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded-full text-xs font-medium">
                <AlertTriangle class="w-3 h-3" />
                Shortage
              </span>
              <span v-else
                class="flex items-center gap-1 px-2 py-0.5 bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 rounded-full text-xs font-medium">
                <CheckCircle class="w-3 h-3" />
                OK
              </span>
              <span class="text-xs text-slate-500 dark:text-slate-400">{{ group.lines.length }} components</span>
            </div>
          </button>

          <!-- Group lines -->
          <div v-if="expandedGroups.has(group.moNumber)" class="border-t border-slate-100 dark:border-slate-800">
            <table class="w-full text-sm">
              <thead>
                <tr class="bg-slate-50 dark:bg-slate-800/70">
                  <th class="px-8 py-2 text-left text-xs font-medium text-slate-400 uppercase">Component</th>
                  <th class="px-4 py-2 text-right text-xs font-medium text-slate-400 uppercase">Required</th>
                  <th class="px-4 py-2 text-right text-xs font-medium text-slate-400 uppercase">Available</th>
                  <th class="px-4 py-2 text-right text-xs font-medium text-slate-400 uppercase">Shortage</th>
                  <th class="px-4 py-2 text-center text-xs font-medium text-slate-400 uppercase">Action</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="line in group.lines" :key="line.component_id + line.mo_id"
                  class="border-t border-slate-100 dark:border-slate-800"
                  :class="line.shortage_qty > 0 ? 'bg-red-50/50 dark:bg-red-950/20' : ''">
                  <td class="px-8 py-2.5">
                    <p class="font-medium text-slate-900 dark:text-white">{{ line.component_name }}</p>
                    <p class="text-xs text-slate-400">{{ line.component_code }}</p>
                  </td>
                  <td class="px-4 py-2.5 text-right font-mono text-slate-700 dark:text-slate-300">
                    {{ fmtNum(line.required_qty) }}
                  </td>
                  <td class="px-4 py-2.5 text-right font-mono"
                    :class="line.available_qty >= line.required_qty
                      ? 'text-emerald-600 dark:text-emerald-400'
                      : 'text-amber-600 dark:text-amber-400'">
                    {{ fmtNum(line.available_qty) }}
                  </td>
                  <td class="px-4 py-2.5 text-right font-mono"
                    :class="line.shortage_qty > 0 ? 'text-red-600 dark:text-red-400 font-semibold' : 'text-slate-400'">
                    {{ line.shortage_qty > 0 ? fmtNum(line.shortage_qty) : '—' }}
                  </td>
                  <td class="px-4 py-2.5 text-center">
                    <span v-if="line.needs_to_purchase"
                      class="inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 rounded-full font-medium">
                      <ShoppingCart class="w-3 h-3" />
                      Purchase
                    </span>
                    <span v-else
                      class="inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 rounded-full font-medium">
                      <CheckCircle class="w-3 h-3" />
                      OK
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Flat View -->
      <div v-else class="flex-1 overflow-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/50">
              <th v-for="col in [
                { key: 'mo_number', label: 'MO Number' },
                { key: 'product_name', label: 'Product' },
                { key: 'component_name', label: 'Component' },
                { key: 'required_qty', label: 'Required' },
                { key: 'available_qty', label: 'Available' },
                { key: 'shortage_qty', label: 'Shortage' },
                { key: 'planned_start', label: 'Start Date' },
                { key: '', label: 'Action' }
              ]" :key="col.label"
                class="px-4 py-3 text-left font-medium text-slate-500 dark:text-slate-400 whitespace-nowrap"
                :class="col.key ? 'cursor-pointer hover:text-slate-700 dark:hover:text-slate-200 select-none' : ''"
                @click="col.key ? setSort(col.key) : null">
                <span class="flex items-center gap-1">
                  {{ col.label }}
                  <template v-if="col.key && sortField === col.key">
                    <ChevronUp v-if="sortDir === 'asc'" class="w-3 h-3" />
                    <ChevronDown v-else class="w-3 h-3" />
                  </template>
                </span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="line in filteredLines" :key="line.component_id + line.mo_id"
              class="border-b border-slate-100 dark:border-slate-800 transition-colors"
              :class="line.shortage_qty > 0
                ? 'bg-red-50/50 dark:bg-red-950/20 hover:bg-red-50 dark:hover:bg-red-950/30'
                : 'hover:bg-slate-50 dark:hover:bg-slate-800/50'">
              <td class="px-4 py-3">
                <span class="font-mono font-semibold text-indigo-700 dark:text-indigo-400">{{ line.mo_number }}</span>
              </td>
              <td class="px-4 py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-white">{{ line.product_name }}</p>
                  <p class="text-xs text-slate-400">{{ line.product_code }}</p>
                </div>
              </td>
              <td class="px-4 py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-white">{{ line.component_name }}</p>
                  <p class="text-xs text-slate-400">{{ line.component_code }}</p>
                </div>
              </td>
              <td class="px-4 py-3 text-right font-mono text-slate-700 dark:text-slate-300">
                {{ fmtNum(line.required_qty) }}
              </td>
              <td class="px-4 py-3 text-right font-mono"
                :class="line.available_qty >= line.required_qty
                  ? 'text-emerald-600 dark:text-emerald-400'
                  : 'text-amber-600 dark:text-amber-400'">
                {{ fmtNum(line.available_qty) }}
              </td>
              <td class="px-4 py-3 text-right font-mono font-semibold"
                :class="line.shortage_qty > 0 ? 'text-red-600 dark:text-red-400' : 'text-slate-400'">
                {{ line.shortage_qty > 0 ? fmtNum(line.shortage_qty) : '—' }}
              </td>
              <td class="px-4 py-3 text-xs text-slate-500 dark:text-slate-400">
                {{ fmtDate(line.planned_start) }}
              </td>
              <td class="px-4 py-3">
                <span v-if="line.needs_to_purchase"
                  class="inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 rounded-full font-medium">
                  <ShoppingCart class="w-3 h-3" />
                  Purchase
                </span>
                <span v-else
                  class="inline-flex items-center gap-1 px-2 py-0.5 text-xs bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 rounded-full font-medium">
                  <CheckCircle class="w-3 h-3" />
                  OK
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="px-4 py-2 border-t border-slate-200 dark:border-slate-700 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
        <span>{{ filteredLines.length }} component lines — {{ shortageLines.length }} shortages</span>
        <span v-if="mrpResult">
          Generated: {{ fmtDateTime(mrpResult.generated_at) }}
        </span>
      </div>
    </div>

    <!-- ── Recent Orders ───────────────────────────────────────────────────── -->
    <div v-if="recentOrders.length > 0"
      class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-5">
      <h3 class="text-sm font-semibold text-slate-900 dark:text-white mb-4 flex items-center gap-2">
        <Cog class="w-4 h-4 text-amber-600" />
        Recent Manufacturing Orders
      </h3>
      <div class="overflow-hidden rounded-lg border border-slate-200 dark:border-slate-700">
        <table class="w-full text-sm">
          <thead class="bg-slate-50 dark:bg-slate-800">
            <tr>
              <th class="px-4 py-2 text-left text-xs font-medium text-slate-500 dark:text-slate-400">Number</th>
              <th class="px-4 py-2 text-left text-xs font-medium text-slate-500 dark:text-slate-400">Product</th>
              <th class="px-4 py-2 text-right text-xs font-medium text-slate-500 dark:text-slate-400">Planned</th>
              <th class="px-4 py-2 text-right text-xs font-medium text-slate-500 dark:text-slate-400">Produced</th>
              <th class="px-4 py-2 text-center text-xs font-medium text-slate-500 dark:text-slate-400">Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in recentOrders" :key="order.number"
              class="border-t border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors">
              <td class="px-4 py-2.5 font-mono text-amber-700 dark:text-amber-400 font-semibold">{{ order.number }}</td>
              <td class="px-4 py-2.5 text-slate-900 dark:text-white">{{ order.product }}</td>
              <td class="px-4 py-2.5 text-right font-mono text-slate-700 dark:text-slate-300">{{ fmtNum(order.planned_qty) }}</td>
              <td class="px-4 py-2.5 text-right font-mono text-slate-700 dark:text-slate-300">{{ fmtNum(order.produced_qty) }}</td>
              <td class="px-4 py-2.5 text-center">
                <span :class="statusConfig[order.status]" class="px-2 py-0.5 rounded-full text-xs font-medium">
                  {{ order.status?.replace('_', ' ') }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ── Empty state (no MRP result yet) ────────────────────────────────── -->
    <div v-if="!mrpResult && !running"
      class="flex-1 flex items-center justify-center bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 py-16">
      <div class="text-center">
        <CalendarClock class="w-12 h-12 mx-auto mb-4 text-slate-300 dark:text-slate-600" />
        <h3 class="text-lg font-semibold text-slate-700 dark:text-slate-300">Run MRP to Get Started</h3>
        <p class="text-sm text-slate-400 dark:text-slate-500 mt-1 max-w-sm">
          Click "Run MRP" to analyze component requirements for all planned and draft manufacturing orders.
        </p>
        <button @click="runMRP" :disabled="running"
          class="mt-4 flex items-center gap-2 px-5 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors mx-auto">
          <CalendarClock class="w-4 h-4" />
          Run MRP Now
        </button>
      </div>
    </div>

  </div>
</template>
