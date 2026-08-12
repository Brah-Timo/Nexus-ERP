<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Cpu, Plus, Search, RefreshCw, Loader2, CheckCircle, X,
  TrendingDown, Wrench, Play, DollarSign, BarChart3,
  Calendar, Tag, ArrowUpDown, AlertCircle, ChevronDown,
  ChevronUp, Activity, Layers, Hash, Clock
} from 'lucide-vue-next'
import { accountingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ─────────────────────────────────────────────────────────────────────
interface FixedAsset {
  id: string
  name: string
  category: string
  purchase_date: string
  purchase_value: number
  depreciation_method: string
  useful_life_years: number
  residual_value: number
  current_value: number
  status: string
  created_at: string
}

// ─── State ─────────────────────────────────────────────────────────────────────
const assets     = ref<FixedAsset[]>([])
const accounts   = ref<any[]>([])
const loading    = ref(true)
const saving     = ref(false)
const deprecating = ref(false)
const search     = ref('')
const statusFilter = ref('')
const showModal  = ref(false)
const expandedId = ref<string | null>(null)

const form = ref({
  name: '',
  category: 'tangible',
  purchase_date: new Date().toISOString().split('T')[0],
  purchase_value: 0,
  depreciation_method: 'linear',
  useful_life_years: 5,
  residual_value: 0,
  account_id: ''
})

// ─── Config ────────────────────────────────────────────────────────────────────
const statusConfig: Record<string, { label: string; color: string; bg: string; border: string; dot: string }> = {
  active:   { label: 'Active',   color: 'text-emerald-400', bg: 'bg-emerald-500/10', border: 'border-emerald-500/20', dot: 'bg-emerald-400' },
  disposed: { label: 'Disposed', color: 'text-rose-400',    bg: 'bg-rose-500/10',    border: 'border-rose-500/20',    dot: 'bg-rose-400' },
  repaired: { label: 'Repaired', color: 'text-amber-400',   bg: 'bg-amber-500/10',   border: 'border-amber-500/20',   dot: 'bg-amber-400' },
}

const categories = ['tangible', 'intangible', 'financial', 'land', 'building', 'equipment', 'vehicle', 'furniture', 'computer']

// ─── Computed ──────────────────────────────────────────────────────────────────
const stats = computed(() => {
  const total        = assets.value.length
  const active       = assets.value.filter(a => a.status === 'active').length
  const disposed     = assets.value.filter(a => a.status === 'disposed').length
  const originalVal  = assets.value.reduce((s, a) => s + (Number(a.purchase_value) || 0), 0)
  const bookVal      = assets.value.reduce((s, a) => s + (Number(a.current_value) || 0), 0)
  const depreciated  = originalVal - bookVal
  return { total, active, disposed, originalVal, bookVal, depreciated }
})

const filtered = computed(() => {
  let list = [...assets.value]
  if (statusFilter.value) list = list.filter(a => a.status === statusFilter.value)
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(a => a.name.toLowerCase().includes(q) || a.category.toLowerCase().includes(q))
  }
  return list
})

const depreciationPreview = computed(() => {
  const base = Number(form.value.purchase_value) - Number(form.value.residual_value)
  if (base <= 0 || form.value.useful_life_years <= 0) return { annual: 0, monthly: 0 }
  const annual = base / form.value.useful_life_years
  const monthly = annual / 12
  return { annual, monthly }
})

// ─── Methods ───────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const [assetsRes, accountsRes] = await Promise.all([
      accountingAPI.getFixedAssets(),
      accountingAPI.getChartOfAccounts()
    ])
    assets.value = assetsRes.data ?? []
    accounts.value = accountsRes.data ?? []
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Failed to load assets', 'error')
  } finally {
    loading.value = false
  }
}

async function runDepreciation() {
  deprecating.value = true
  try {
    const res = await accountingAPI.runDepreciation()
    const count = res.data?.results?.length ?? 0
    app.addToast(`Depreciation run complete — ${count} assets updated`, 'success')
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Depreciation run failed', 'error')
  } finally {
    deprecating.value = false
  }
}

async function createAsset() {
  if (!form.value.name || !form.value.purchase_value) {
    app.addToast('Name and original value are required', 'error')
    return
  }
  saving.value = true
  try {
    await accountingAPI.createFixedAsset({ ...form.value })
    app.addToast('Fixed asset created', 'success')
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Create failed', 'error')
  } finally {
    saving.value = false
  }
}

function depPct(asset: FixedAsset) {
  const orig = Number(asset.purchase_value)
  const book = Number(asset.current_value)
  const residual = Number(asset.residual_value)
  if (!orig) return 0
  const depreciableBase = orig - residual
  if (!depreciableBase) return 0
  const depreciated = orig - book
  return Math.min(100, Math.max(0, (depreciated / depreciableBase) * 100))
}

function depPctColor(pct: number) {
  if (pct < 40) return 'bg-emerald-500'
  if (pct < 75) return 'bg-amber-500'
  return 'bg-rose-500'
}

function fmtDate(d: string) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function fmtAmt(v: number) {
  return new Intl.NumberFormat('en-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(v ?? 0)
}

function fmtAmtDec(v: number) {
  return new Intl.NumberFormat('en-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v ?? 0)
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col h-full transition-colors duration-200"
     :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">

    <!-- ── Header ─────────────────────────────────────────────────────────── -->
    <div class="border-b border-slate-800/60 bg-slate-900/50 backdrop-blur-sm px-6 py-4 flex-shrink-0">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-lg bg-amber-500/15 border border-amber-500/25 flex items-center justify-center">
            <Cpu class="w-4.5 h-4.5 text-amber-400" />
          </div>
          <div>
            <h1 class="text-[15px] font-semibold text-slate-100">Fixed Assets</h1>
            <p class="text-[11px] text-slate-500">Asset register with depreciation tracking</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            class="h-8 px-3 rounded-lg border border-slate-700/60 bg-slate-800/60 text-slate-400 hover:text-slate-200 text-xs font-medium inline-flex items-center gap-1.5 transition-all disabled:opacity-50">
            <RefreshCw class="w-3.5 h-3.5" :class="loading ? 'animate-spin' : ''" />
            Refresh
          </button>
          <button @click="runDepreciation" :disabled="deprecating"
            class="h-8 px-3 rounded-lg border border-amber-500/30 bg-amber-500/10 hover:bg-amber-500/20 text-amber-400 text-xs font-medium inline-flex items-center gap-1.5 transition-all disabled:opacity-50">
            <Loader2 v-if="deprecating" class="w-3.5 h-3.5 animate-spin" />
            <Play v-else class="w-3.5 h-3.5" />
            Run Depreciation
          </button>
          <button @click="showModal=true"
            class="h-8 px-3 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-xs font-medium inline-flex items-center gap-1.5 transition-all shadow-lg shadow-amber-900/30">
            <Plus class="w-3.5 h-3.5" />
            New Asset
          </button>
        </div>
      </div>
    </div>

    <!-- ── Summary bar ────────────────────────────────────────────────────── -->
    <div class="px-6 py-4 flex-shrink-0 grid grid-cols-6 gap-3">
      <div class="rounded-xl bg-slate-900/70 border border-slate-800/50 p-3.5">
        <div class="text-[10px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Total Assets</div>
        <div class="text-xl font-bold text-slate-100">{{ stats.total }}</div>
      </div>
      <div class="rounded-xl bg-slate-900/70 border border-emerald-500/20 bg-emerald-500/5 p-3.5">
        <div class="text-[10px] font-semibold uppercase tracking-wider text-emerald-500/70 mb-1.5">Active</div>
        <div class="text-xl font-bold text-emerald-400">{{ stats.active }}</div>
      </div>
      <div class="rounded-xl bg-slate-900/70 border border-rose-500/20 bg-rose-500/5 p-3.5">
        <div class="text-[10px] font-semibold uppercase tracking-wider text-rose-500/70 mb-1.5">Disposed</div>
        <div class="text-xl font-bold text-rose-400">{{ stats.disposed }}</div>
      </div>
      <div class="rounded-xl bg-slate-900/70 border border-slate-800/50 p-3.5">
        <div class="text-[10px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Purchase Value</div>
        <div class="text-base font-bold text-slate-100 truncate">{{ fmtAmt(stats.originalVal) }}</div>
        <div class="text-[10px] text-slate-600">DZD</div>
      </div>
      <div class="rounded-xl bg-slate-900/70 border border-amber-500/20 bg-amber-500/5 p-3.5">
        <div class="text-[10px] font-semibold uppercase tracking-wider text-amber-500/70 mb-1.5">Current Value</div>
        <div class="text-base font-bold text-amber-400 truncate">{{ fmtAmt(stats.bookVal) }}</div>
        <div class="text-[10px] text-amber-500/40">DZD</div>
      </div>
      <div class="rounded-xl bg-slate-900/70 border border-slate-800/50 p-3.5">
        <div class="text-[10px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Depreciated</div>
        <div class="text-base font-bold text-rose-400 truncate">{{ fmtAmt(stats.depreciated) }}</div>
        <div class="text-[10px] text-slate-600">DZD</div>
      </div>
    </div>

    <!-- ── Filters ────────────────────────────────────────────────────────── -->
    <div class="px-6 pb-3 flex-shrink-0 flex items-center gap-3">
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500" />
        <input v-model="search" type="text" placeholder="Search assets..."
          class="h-8 w-56 pl-8 pr-3 rounded-lg bg-slate-900 border border-slate-700/60 text-sm text-slate-200 placeholder-slate-600 focus:outline-none focus:border-amber-500/60 focus:ring-1 focus:ring-amber-500/20 transition-all" />
        <button v-if="search" @click="search=''" class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-600 hover:text-slate-400">
          <X class="w-3 h-3" />
        </button>
      </div>
      <div class="flex items-center gap-1">
        <button @click="statusFilter=''"
          :class="['h-7 px-3 rounded-full text-xs font-medium transition-all', !statusFilter ? 'bg-amber-600 text-white' : 'bg-slate-900 border border-slate-700/60 text-slate-500 hover:text-slate-300']">
          All
        </button>
        <button v-for="(cfg, s) in statusConfig" :key="s" @click="statusFilter = statusFilter===s?'':s"
          :class="['h-7 px-3 rounded-full text-xs font-medium transition-all border', statusFilter===s ? `${cfg.bg} ${cfg.color} ${cfg.border}` : 'bg-slate-900 border-slate-700/60 text-slate-500 hover:text-slate-300']">
          {{ cfg.label }}
        </button>
      </div>
      <div class="ml-auto text-xs text-slate-600">{{ filtered.length }} assets</div>
    </div>

    <!-- ── Asset Cards ────────────────────────────────────────────────────── -->
    <div class="flex-1 overflow-auto px-6 pb-6">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <Loader2 class="w-7 h-7 text-amber-400 animate-spin" />
      </div>
      <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center py-20 gap-3">
        <Cpu class="w-12 h-12 text-slate-700" />
        <p class="text-slate-500 text-sm">No fixed assets found</p>
        <button @click="showModal=true" class="px-4 py-2 rounded-lg bg-amber-600 text-white text-sm hover:bg-amber-500 transition-all">Register First Asset</button>
      </div>
      <div v-else class="space-y-2">
        <div v-for="asset in filtered" :key="asset.id"
          class="rounded-xl border border-slate-800/60 bg-slate-900/40 hover:border-slate-700/80 transition-all overflow-hidden">
          <!-- Main row -->
          <div class="px-4 py-3.5 flex items-center gap-4 cursor-pointer" @click="expandedId = expandedId===asset.id ? null : asset.id">
            <!-- Name + category -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-1">
                <Cpu class="w-3.5 h-3.5 text-amber-400/60 flex-shrink-0" />
                <span class="font-semibold text-[13px] text-slate-100 truncate">{{ asset.name }}</span>
                <span :class="['inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-semibold', statusConfig[asset.status]?.bg, statusConfig[asset.status]?.color, statusConfig[asset.status]?.border, 'border']">
                  <span class="w-1 h-1 rounded-full" :class="statusConfig[asset.status]?.dot" />
                  {{ statusConfig[asset.status]?.label ?? asset.status }}
                </span>
              </div>
              <div class="flex items-center gap-3">
                <span class="text-[11px] text-slate-500 capitalize">{{ asset.category }}</span>
                <span class="text-slate-700 text-[10px]">•</span>
                <span class="text-[11px] text-slate-500 capitalize">{{ asset.depreciation_method }} method</span>
                <span class="text-slate-700 text-[10px]">•</span>
                <span class="text-[11px] text-slate-500">{{ asset.useful_life_years }}y life</span>
              </div>
            </div>

            <!-- Depreciation bar section -->
            <div class="w-48 flex-shrink-0">
              <div class="flex items-center justify-between mb-1">
                <span class="text-[10px] text-slate-600">Depreciation</span>
                <span class="text-[11px] font-semibold" :class="depPct(asset) < 40 ? 'text-emerald-400' : depPct(asset) < 75 ? 'text-amber-400' : 'text-rose-400'">
                  {{ depPct(asset).toFixed(0) }}%
                </span>
              </div>
              <div class="h-1.5 rounded-full bg-slate-800 overflow-hidden">
                <div class="h-full rounded-full transition-all" :class="depPctColor(depPct(asset))" :style="`width:${depPct(asset)}%`" />
              </div>
            </div>

            <!-- Financials -->
            <div class="flex items-center gap-6 flex-shrink-0">
              <div class="text-right">
                <div class="text-[10px] text-slate-600 mb-0.5">Current Value</div>
                <div class="text-sm font-mono font-semibold text-amber-400">{{ fmtAmt(asset.current_value) }}</div>
              </div>
              <div class="text-right">
                <div class="text-[10px] text-slate-600 mb-0.5">Original</div>
                <div class="text-sm font-mono text-slate-400">{{ fmtAmt(asset.purchase_value) }}</div>
              </div>
            </div>

            <!-- Chevron -->
            <ChevronDown class="w-4 h-4 text-slate-600 flex-shrink-0 transition-transform" :class="expandedId===asset.id ? 'rotate-180' : ''" />
          </div>

          <!-- Expanded details -->
          <Transition
            enter-active-class="transition-all duration-200"
            enter-from-class="opacity-0 max-h-0"
            enter-to-class="opacity-100 max-h-40"
            leave-active-class="transition-all duration-150"
            leave-from-class="opacity-100 max-h-40"
            leave-to-class="opacity-0 max-h-0">
            <div v-if="expandedId===asset.id" class="border-t border-slate-800/60 px-4 py-3 bg-slate-900/60">
              <div class="grid grid-cols-4 gap-4">
                <div>
                  <div class="text-[10px] text-slate-600 mb-0.5 uppercase tracking-wider font-semibold">Purchase Date</div>
                  <div class="text-xs text-slate-300">{{ fmtDate(asset.purchase_date) }}</div>
                </div>
                <div>
                  <div class="text-[10px] text-slate-600 mb-0.5 uppercase tracking-wider font-semibold">Residual Value</div>
                  <div class="text-xs font-mono text-slate-300">{{ fmtAmtDec(asset.residual_value) }}</div>
                </div>
                <div>
                  <div class="text-[10px] text-slate-600 mb-0.5 uppercase tracking-wider font-semibold">Accumulated Dep.</div>
                  <div class="text-xs font-mono text-rose-400">{{ fmtAmtDec(Number(asset.purchase_value) - Number(asset.current_value)) }}</div>
                </div>
                <div>
                  <div class="text-[10px] text-slate-600 mb-0.5 uppercase tracking-wider font-semibold">Monthly Dep.</div>
                  <div class="text-xs font-mono text-amber-400">
                    {{ fmtAmtDec((Number(asset.purchase_value) - Number(asset.residual_value)) / (asset.useful_life_years * 12)) }}
                  </div>
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </div>

    <!-- ── Create Modal ──────────────────────────────────────────────────── -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-all duration-200"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-all duration-150"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" @click="showModal=false" />
          <Transition
            enter-active-class="transition-all duration-200"
            enter-from-class="opacity-0 scale-95"
            enter-to-class="opacity-100 scale-100">
            <div v-if="showModal" class="relative w-full max-w-lg bg-slate-900 border border-slate-700/60 rounded-2xl shadow-2xl overflow-hidden">
              <!-- Header -->
              <div class="px-6 py-4 border-b border-slate-800/60 flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-amber-500/15 border border-amber-500/25 flex items-center justify-center">
                    <Cpu class="w-4 h-4 text-amber-400" />
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-slate-100">Register Fixed Asset</h3>
                    <p class="text-[11px] text-slate-500">Add asset to the register</p>
                  </div>
                </div>
                <button @click="showModal=false" class="w-7 h-7 rounded-lg flex items-center justify-center text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-all">
                  <X class="w-4 h-4" />
                </button>
              </div>
              <!-- Body -->
              <div class="px-6 py-5 space-y-4">
                <!-- Name -->
                <div>
                  <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Asset Name *</label>
                  <input v-model="form.name" type="text" placeholder="e.g. Server Rack Unit A"
                    class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-amber-500/60 focus:ring-1 focus:ring-amber-500/20" />
                </div>
                <div class="grid grid-cols-2 gap-4">
                  <!-- Category -->
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Category</label>
                    <select v-model="form.category"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-amber-500/60 capitalize">
                      <option v-for="c in categories" :key="c" :value="c">{{ c.charAt(0).toUpperCase()+c.slice(1) }}</option>
                    </select>
                  </div>
                  <!-- Acquisition date -->
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Purchase Date</label>
                    <input v-model="form.purchase_date" type="date"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-amber-500/60" />
                  </div>
                </div>
                <div class="grid grid-cols-2 gap-4">
                  <!-- Original value -->
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Original Value (DZD) *</label>
                    <input v-model.number="form.purchase_value" type="number" min="0" step="1000" placeholder="0"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm font-mono text-slate-100 placeholder-slate-600 focus:outline-none focus:border-amber-500/60" />
                  </div>
                  <!-- Residual value -->
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Residual Value (DZD)</label>
                    <input v-model.number="form.residual_value" type="number" min="0" step="100" placeholder="0"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm font-mono text-slate-100 placeholder-slate-600 focus:outline-none focus:border-amber-500/60" />
                  </div>
                </div>
                <div class="grid grid-cols-2 gap-4">
                  <!-- Method -->
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Depreciation Method</label>
                    <select v-model="form.depreciation_method"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-amber-500/60">
                      <option value="linear">Linear (Straight-line)</option>
                      <option value="diminishing">Diminishing Balance</option>
                    </select>
                  </div>
                  <!-- Useful life -->
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Useful Life (Years)</label>
                    <input v-model.number="form.useful_life_years" type="number" min="1" max="50"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm font-mono text-slate-100 focus:outline-none focus:border-amber-500/60" />
                  </div>
                </div>
                <!-- Depreciation preview -->
                <div v-if="form.purchase_value > 0 && form.useful_life_years > 0"
                  class="rounded-lg bg-amber-500/5 border border-amber-500/20 px-4 py-3">
                  <div class="text-[10px] font-semibold uppercase tracking-wider text-amber-500/70 mb-2">Depreciation Preview</div>
                  <div class="grid grid-cols-2 gap-3">
                    <div>
                      <div class="text-[11px] text-slate-500">Annual</div>
                      <div class="text-sm font-mono font-semibold text-amber-400">{{ fmtAmtDec(depreciationPreview.annual) }} DZD</div>
                    </div>
                    <div>
                      <div class="text-[11px] text-slate-500">Monthly</div>
                      <div class="text-sm font-mono font-semibold text-amber-400">{{ fmtAmtDec(depreciationPreview.monthly) }} DZD</div>
                    </div>
                  </div>
                </div>
              </div>
              <!-- Footer -->
              <div class="px-6 py-4 border-t border-slate-800/60 flex items-center justify-end gap-3">
                <button @click="showModal=false" class="h-9 px-4 rounded-lg border border-slate-700 bg-slate-800/40 text-slate-400 hover:text-slate-200 text-sm font-medium transition-all">Cancel</button>
                <button @click="createAsset" :disabled="saving"
                  class="h-9 px-5 rounded-lg bg-amber-600 hover:bg-amber-500 disabled:opacity-50 text-white text-sm font-medium inline-flex items-center gap-2 transition-all shadow-lg shadow-amber-900/30">
                  <Loader2 v-if="saving" class="w-3.5 h-3.5 animate-spin" />
                  <CheckCircle v-else class="w-3.5 h-3.5" />
                  Register Asset
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
