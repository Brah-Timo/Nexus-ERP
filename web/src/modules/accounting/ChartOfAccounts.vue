<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  LayoutList, Plus, Search, X, ChevronDown, ChevronRight,
  Loader2, CheckCircle, AlertCircle, Pencil, RefreshCw,
  TrendingUp, TrendingDown, Landmark, Scale, DollarSign,
  BarChart3, Network, Filter, ArrowUpDown, SlidersHorizontal,
  Layers, BookOpen, Hash, Globe, Check, TreePine
} from 'lucide-vue-next'
import { accountingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ─────────────────────────────────────────────────────────────────────
interface Account {
  id: string
  code: string
  name: string
  type: string
  parent_id: string | null
  is_group: boolean
  is_reconcilable: boolean
  currency: string
  balance: number
  children?: Account[]
}

// ─── State ─────────────────────────────────────────────────────────────────────
const accounts    = ref<Account[]>([])
const loading     = ref(true)
const saving      = ref(false)
const error       = ref('')
const search      = ref('')
const typeFilter  = ref('')
const viewMode    = ref<'flat' | 'tree'>('flat')
const showModal   = ref(false)
const editTarget  = ref<Account | null>(null)
const expandedRows = ref<Set<string>>(new Set())
const sortField   = ref<'code' | 'name' | 'balance'>('code')
const sortDir     = ref<'asc' | 'desc'>('asc')

const form = ref({
  code: '',
  name: '',
  
  type: 'asset',
  parent_id: null as string | null,
  is_group: false,
  is_reconcilable: false,
  currency: 'DZD'
})

// ─── Account type config ───────────────────────────────────────────────────────
const typeConfig: Record<string, { label: string; color: string; bg: string; border: string; dot: string }> = {
  asset:     { label: 'Asset',     color: 'text-blue-400',   bg: 'bg-blue-500/10',   border: 'border-blue-500/20',   dot: 'bg-blue-400' },
  liability: { label: 'Liability', color: 'text-rose-400',   bg: 'bg-rose-500/10',   border: 'border-rose-500/20',   dot: 'bg-rose-400' },
  equity:    { label: 'Equity',    color: 'text-violet-400', bg: 'bg-violet-500/10', border: 'border-violet-500/20', dot: 'bg-violet-400' },
  revenue:   { label: 'Revenue',   color: 'text-emerald-400',bg: 'bg-emerald-500/10',border: 'border-emerald-500/20',dot: 'bg-emerald-400' },
  expense:   { label: 'Expense',   color: 'text-amber-400',  bg: 'bg-amber-500/10',  border: 'border-amber-500/20',  dot: 'bg-amber-400' },
}

// ─── Computed ──────────────────────────────────────────────────────────────────
const stats = computed(() => {
  const result: Record<string, { count: number; balance: number }> = {}
  for (const t of Object.keys(typeConfig)) {
    const subset = accounts.value.filter(a => a.type === t)
    result[t] = {
      count: subset.length,
      balance: subset.reduce((s, a) => s + (a.balance ?? 0), 0)
    }
  }
  return result
})

const totalAccounts = computed(() => accounts.value.length)

const filteredAccounts = computed(() => {
  let list = [...accounts.value]
  if (typeFilter.value) list = list.filter(a => a.type === typeFilter.value)
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(a =>
      a.code.toLowerCase().includes(q) ||
      a.name.toLowerCase().includes(q) ||
      a.name_ar?.includes(q)
    )
  }
  list.sort((a, b) => {
    let va: string | number = a[sortField.value]
    let vb: string | number = b[sortField.value]
    if (typeof va === 'string') va = va.toLowerCase()
    if (typeof vb === 'string') vb = vb.toLowerCase()
    if (va < vb) return sortDir.value === 'asc' ? -1 : 1
    if (va > vb) return sortDir.value === 'asc' ? 1 : -1
    return 0
  })
  return list
})

const treeData = computed(() => {
  const rootItems: Account[] = []
  const map: Record<string, Account> = {}
  const src = typeFilter.value
    ? accounts.value.filter(a => a.type === typeFilter.value)
    : accounts.value
  src.forEach(a => (map[a.id] = { ...a, children: [] }))
  src.forEach(a => {
    if (a.parent_id && map[a.parent_id]) {
      map[a.parent_id].children!.push(map[a.id])
    } else {
      rootItems.push(map[a.id])
    }
  })
  return rootItems
})

const parentAccounts = computed(() =>
  accounts.value.filter(a => a.is_group)
)

// ─── Methods ───────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await accountingAPI.getChartOfAccounts()
    accounts.value = res.data ?? []
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? 'Failed to load accounts'
    app.addToast(error.value, 'error')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editTarget.value = null
  form.value = { code: '', name: '',  type: 'asset', parent_id: null, is_group: false, is_reconcilable: false, currency: 'DZD' }
  showModal.value = true
}

function openEdit(a: Account) {
  editTarget.value = a
  form.value = { code: a.code, name: a.name,  type: a.type, parent_id: a.parent_id, is_group: a.is_group, is_reconcilable: a.is_reconcilable, currency: a.currency }
  showModal.value = true
}

async function save() {
  if (!form.value.code || !form.value.name) {
    app.addToast('Code and name are required', 'error')
    return
  }
  saving.value = true
  try {
    if (editTarget.value) {
      await accountingAPI.createAccount({ ...form.value, id: editTarget.value.id })
      app.addToast('Account updated', 'success')
    } else {
      await accountingAPI.createAccount(form.value)
      app.addToast('Account created', 'success')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

function toggleRow(id: string) {
  expandedRows.value.has(id) ? expandedRows.value.delete(id) : expandedRows.value.add(id)
  expandedRows.value = new Set(expandedRows.value)
}

function setSort(field: 'code' | 'name' | 'balance') {
  if (sortField.value === field) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortField.value = field; sortDir.value = 'asc' }
}

function fmtBalance(v: number) {
  if (v == null) return '0.00'
  return new Intl.NumberFormat('en-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)
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
          <div class="w-9 h-9 rounded-lg bg-indigo-500/15 border border-indigo-500/25 flex items-center justify-center">
            <LayoutList class="w-4.5 h-4.5 text-indigo-400" />
          </div>
          <div>
            <h1 class="text-[15px] font-semibold text-slate-100 leading-tight">Chart of Accounts</h1>
            <p class="text-[11px] text-slate-500 leading-tight">SCF — Système Comptable Financier</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            class="h-8 px-3 rounded-lg border border-slate-700/60 bg-slate-800/60 text-slate-400 hover:text-slate-200 hover:border-slate-600 text-xs font-medium inline-flex items-center gap-1.5 transition-all disabled:opacity-50">
            <RefreshCw class="w-3.5 h-3.5" :class="loading ? 'animate-spin' : ''" />
            Refresh
          </button>
          <button @click="openCreate"
            class="h-8 px-3 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium inline-flex items-center gap-1.5 transition-all shadow-lg shadow-indigo-900/30">
            <Plus class="w-3.5 h-3.5" />
            New Account
          </button>
        </div>
      </div>
    </div>

    <!-- ── KPI Stats ──────────────────────────────────────────────────────── -->
    <div class="px-6 py-4 flex-shrink-0 grid grid-cols-5 gap-3">
      <button v-for="(cfg, type) in typeConfig" :key="type"
        @click="typeFilter = typeFilter === type ? '' : type"
        :class="[
          'group relative rounded-xl p-3.5 border transition-all text-left overflow-hidden',
          typeFilter === type
            ? `${cfg.bg} ${cfg.border} border-2`
            : 'bg-slate-900/70 border-slate-800/50 hover:border-slate-700'
        ]">
        <div class="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity"
          :class="cfg.bg" style="opacity:0.04;" />
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider" :class="cfg.color">{{ cfg.label }}</span>
          <div class="w-1.5 h-1.5 rounded-full" :class="cfg.dot" />
        </div>
        <div class="text-xl font-bold text-slate-100">{{ stats[type]?.count ?? 0 }}</div>
        <div class="text-[11px] mt-0.5" :class="cfg.color">
          DZD {{ fmtBalance(stats[type]?.balance ?? 0) }}
        </div>
        <div v-if="typeFilter === type"
          class="absolute bottom-0 left-0 right-0 h-0.5 rounded-b-xl" :class="cfg.dot" />
      </button>
    </div>

    <!-- ── Toolbar ────────────────────────────────────────────────────────── -->
    <div class="px-6 pb-3 flex-shrink-0 flex items-center gap-3">
      <!-- Search -->
      <div class="relative flex-1 max-w-xs">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500" />
        <input v-model="search" type="text" placeholder="Search code, name..."
          class="w-full h-8 pl-8 pr-3 rounded-lg bg-slate-900 border border-slate-700/60 text-sm text-slate-200 placeholder-slate-600 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/20 transition-all" />
        <button v-if="search" @click="search=''" class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-600 hover:text-slate-400">
          <X class="w-3 h-3" />
        </button>
      </div>

      <!-- View toggle -->
      <div class="flex items-center bg-slate-900 border border-slate-700/60 rounded-lg p-0.5">
        <button @click="viewMode='flat'"
          :class="['h-7 px-3 rounded-md text-xs font-medium transition-all', viewMode==='flat' ? 'bg-indigo-600 text-white shadow' : 'text-slate-500 hover:text-slate-300']">
          <span class="flex items-center gap-1.5"><LayoutList class="w-3 h-3" /> Flat</span>
        </button>
        <button @click="viewMode='tree'"
          :class="['h-7 px-3 rounded-md text-xs font-medium transition-all', viewMode==='tree' ? 'bg-indigo-600 text-white shadow' : 'text-slate-500 hover:text-slate-300']">
          <span class="flex items-center gap-1.5"><TreePine class="w-3 h-3" /> Tree</span>
        </button>
      </div>

      <div class="ml-auto flex items-center gap-2">
        <span v-if="typeFilter" class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[11px] font-medium"
          :class="[typeConfig[typeFilter]?.bg, typeConfig[typeFilter]?.color, typeConfig[typeFilter]?.border, 'border']">
          {{ typeConfig[typeFilter]?.label }}
          <button @click="typeFilter=''" class="ml-0.5 hover:opacity-70"><X class="w-2.5 h-2.5" /></button>
        </span>
        <span class="text-xs text-slate-600">{{ filteredAccounts.length }} / {{ totalAccounts }}</span>
      </div>
    </div>

    <!-- ── Table / Tree ────────────────────────────────────────────────────── -->
    <div class="flex-1 overflow-hidden px-6 pb-6">
      <div class="h-full rounded-xl border border-slate-800/60 overflow-hidden bg-slate-900/40">

        <!-- Loading state -->
        <div v-if="loading" class="flex flex-col items-center justify-center h-full gap-3">
          <Loader2 class="w-8 h-8 text-indigo-400 animate-spin" />
          <span class="text-sm text-slate-500">Loading accounts...</span>
        </div>

        <!-- Error state -->
        <div v-else-if="error" class="flex flex-col items-center justify-center h-full gap-3">
          <AlertCircle class="w-8 h-8 text-rose-400" />
          <span class="text-sm text-slate-400">{{ error }}</span>
          <button @click="load" class="px-4 py-2 rounded-lg bg-slate-800 text-slate-300 text-sm hover:bg-slate-700 transition-all">Retry</button>
        </div>

        <!-- FLAT VIEW ──────────────────────────────────────────────────────── -->
        <template v-else-if="viewMode === 'flat'">
          <div class="overflow-auto h-full">
            <table class="w-full text-sm border-collapse">
              <thead class="sticky top-0 z-10">
                <tr class="bg-slate-900/90 backdrop-blur border-b border-slate-800/60">
                  <th class="text-left px-4 py-3 w-32">
                    <button @click="setSort('code')" class="flex items-center gap-1 text-[11px] font-semibold uppercase tracking-wider text-slate-500 hover:text-slate-300 transition-colors">
                      Code
                      <ArrowUpDown class="w-3 h-3" :class="sortField==='code' ? 'text-indigo-400' : ''" />
                    </button>
                  </th>
                  <th class="text-left px-4 py-3">
                    <button @click="setSort('name')" class="flex items-center gap-1 text-[11px] font-semibold uppercase tracking-wider text-slate-500 hover:text-slate-300 transition-colors">
                      Account Name
                      <ArrowUpDown class="w-3 h-3" :class="sortField==='name' ? 'text-indigo-400' : ''" />
                    </button>
                  </th>
                  <th class="text-left px-4 py-3 w-28 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Type</th>
                  <th class="text-left px-4 py-3 w-20 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Currency</th>
                  <th class="text-left px-4 py-3 w-20 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Flags</th>
                  <th class="text-right px-4 py-3 w-40">
                    <button @click="setSort('balance')" class="flex items-center gap-1 ml-auto text-[11px] font-semibold uppercase tracking-wider text-slate-500 hover:text-slate-300 transition-colors">
                      Balance
                      <ArrowUpDown class="w-3 h-3" :class="sortField==='balance' ? 'text-indigo-400' : ''" />
                    </button>
                  </th>
                  <th class="px-4 py-3 w-20 text-[11px] font-semibold uppercase tracking-wider text-slate-500 text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                <template v-if="filteredAccounts.length === 0">
                  <tr>
                    <td colspan="7" class="py-20 text-center text-slate-600">
                      <LayoutList class="w-10 h-10 mx-auto mb-3 opacity-30" />
                      <p class="text-sm">No accounts found</p>
                    </td>
                  </tr>
                </template>
                <tr v-for="(account, i) in filteredAccounts" :key="account.id"
                  :class="['border-b border-slate-800/30 hover:bg-slate-800/30 transition-colors group', i % 2 === 0 ? '' : 'bg-slate-900/20']">
                  <!-- Code -->
                  <td class="px-4 py-3">
                    <span class="font-mono text-[13px] font-semibold text-indigo-300">{{ account.code }}</span>
                  </td>
                  <!-- Name -->
                  <td class="px-4 py-3">
                    <div class="font-medium text-slate-200 text-[13px]">{{ account.name }}</div>
                    <div v-if="account.name_ar" class="text-[11px] text-slate-500 mt-0.5 text-right font-arabic">{{ account.name_ar }}</div>
                  </td>
                  <!-- Type badge -->
                  <td class="px-4 py-3">
                    <span :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold border', typeConfig[account.type]?.bg, typeConfig[account.type]?.color, typeConfig[account.type]?.border]">
                      <span class="w-1 h-1 rounded-full" :class="typeConfig[account.type]?.dot" />
                      {{ typeConfig[account.type]?.label }}
                    </span>
                  </td>
                  <!-- Currency -->
                  <td class="px-4 py-3">
                    <span class="text-[11px] font-mono text-slate-400 bg-slate-800/60 px-1.5 py-0.5 rounded">{{ account.currency }}</span>
                  </td>
                  <!-- Flags -->
                  <td class="px-4 py-3">
                    <div class="flex items-center gap-1">
                      <span v-if="account.is_group" title="Postable" class="w-4 h-4 rounded bg-emerald-500/15 border border-emerald-500/30 flex items-center justify-center">
                        <Check class="w-2.5 h-2.5 text-emerald-400" />
                      </span>
                      <span v-if="account.is_reconcilable" title="Reconcilable" class="w-4 h-4 rounded bg-blue-500/15 border border-blue-500/30 flex items-center justify-center">
                        <Scale class="w-2.5 h-2.5 text-blue-400" />
                      </span>
                    </div>
                  </td>
                  <!-- Balance -->
                  <td class="px-4 py-3 text-right">
                    <span :class="['font-mono text-[13px] font-semibold', (account.balance ?? 0) >= 0 ? 'text-emerald-400' : 'text-rose-400']">
                      {{ fmtBalance(account.balance) }}
                    </span>
                  </td>
                  <!-- Actions -->
                  <td class="px-4 py-3 text-right">
                    <button @click="openEdit(account)"
                      class="opacity-0 group-hover:opacity-100 inline-flex items-center justify-center w-7 h-7 rounded-lg bg-slate-700/60 hover:bg-indigo-500/20 hover:text-indigo-400 text-slate-400 transition-all">
                      <Pencil class="w-3.5 h-3.5" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>

        <!-- TREE VIEW ───────────────────────────────────────────────────────── -->
        <template v-else>
          <div class="overflow-auto h-full">
            <table class="w-full text-sm border-collapse">
              <thead class="sticky top-0 z-10">
                <tr class="bg-slate-900/90 backdrop-blur border-b border-slate-800/60">
                  <th class="text-left px-4 py-3 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Code & Name</th>
                  <th class="text-left px-4 py-3 w-28 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Type</th>
                  <th class="text-right px-4 py-3 w-40 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Balance</th>
                  <th class="px-4 py-3 w-16 text-[11px] font-semibold uppercase tracking-wider text-slate-500 text-right">Edit</th>
                </tr>
              </thead>
              <tbody>
                <template v-if="treeData.length === 0">
                  <tr><td colspan="4" class="py-20 text-center text-slate-600 text-sm">No accounts found</td></tr>
                </template>
                <TreeAccountRow
                  v-for="node in treeData" :key="node.id"
                  :account="node" :depth="0"
                  :expanded-rows="expandedRows"
                  :type-config="typeConfig"
                  @toggle="toggleRow"
                  @edit="openEdit"
                  :fmt-balance="fmtBalance"
                />
              </tbody>
            </table>
          </div>
        </template>
      </div>
    </div>

    <!-- ── Create / Edit Modal ────────────────────────────────────────────── -->
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
            enter-from-class="opacity-0 scale-95 translate-y-2"
            enter-to-class="opacity-100 scale-100 translate-y-0"
            leave-active-class="transition-all duration-150"
            leave-from-class="opacity-100 scale-100"
            leave-to-class="opacity-0 scale-95">
            <div v-if="showModal" class="relative w-full max-w-lg bg-slate-900 border border-slate-700/60 rounded-2xl shadow-2xl overflow-hidden">

              <!-- Modal header -->
              <div class="px-6 py-4 border-b border-slate-800/60 flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-indigo-500/15 border border-indigo-500/25 flex items-center justify-center">
                    <BookOpen class="w-4 h-4 text-indigo-400" />
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-slate-100">{{ editTarget ? 'Edit Account' : 'New Account' }}</h3>
                    <p class="text-[11px] text-slate-500">{{ editTarget ? `Editing ${editTarget.code}` : 'Add to chart of accounts' }}</p>
                  </div>
                </div>
                <button @click="showModal=false" class="w-7 h-7 rounded-lg flex items-center justify-center text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-all">
                  <X class="w-4 h-4" />
                </button>
              </div>

              <!-- Modal body -->
              <div class="px-6 py-5 space-y-4">
                <div class="grid grid-cols-2 gap-4">
                  <!-- Code -->
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Account Code *</label>
                    <input v-model="form.code" type="text" placeholder="e.g. 5121"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/20 font-mono" />
                  </div>
                  <!-- Currency -->
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Currency</label>
                    <select v-model="form.currency"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/20">
                      <option value="DZD">DZD — Dinar</option>
                      <option value="EUR">EUR — Euro</option>
                      <option value="USD">USD — Dollar</option>
                    </select>
                  </div>
                </div>

                <!-- Name -->
                <div>
                  <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Account Name (French) *</label>
                  <input v-model="form.name" type="text" placeholder="e.g. Banques locales en monnaies nationales"
                    class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/20" />
                </div>

                <!-- Arabic name -->
                <div>
                  <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Account Name (Arabic)</label>
                  <input v-model="form.name_ar" type="text" placeholder="اسم الحساب بالعربية" dir="rtl"
                    class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/20 text-right" />
                </div>

                <!-- Type + Parent -->
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Account Type *</label>
                    <select v-model="form.type"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/20">
                      <option v-for="(cfg, t) in typeConfig" :key="t" :value="t">{{ cfg.label }}</option>
                    </select>
                  </div>
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Parent Account</label>
                    <select v-model="form.parent_id"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-indigo-500/60 focus:ring-1 focus:ring-indigo-500/20">
                      <option :value="null">-- None (root) --</option>
                      <option v-for="p in accounts.filter(a => a.type === form.type && a.id !== editTarget?.id)" :key="p.id" :value="p.id">
                        {{ p.code }} — {{ p.name }}
                      </option>
                    </select>
                  </div>
                </div>

                <!-- Toggles -->
                <div class="flex items-center gap-6 pt-1">
                  <label class="flex items-center gap-2.5 cursor-pointer group">
                    <button type="button" @click="form.is_group = !form.is_group"
                      :class="['w-9 h-5 rounded-full relative transition-all', form.is_group ? 'bg-indigo-600' : 'bg-slate-700']">
                      <span :class="['absolute top-0.5 w-4 h-4 rounded-full bg-white shadow transition-all', form.is_group ? 'left-4.5' : 'left-0.5']" />
                    </button>
                    <span class="text-xs text-slate-400 group-hover:text-slate-200 transition-colors">Group Account</span>
                  </label>
                  <label class="flex items-center gap-2.5 cursor-pointer group">
                    <button type="button" @click="form.is_reconcilable = !form.is_reconcilable"
                      :class="['w-9 h-5 rounded-full relative transition-all', form.is_reconcilable ? 'bg-indigo-600' : 'bg-slate-700']">
                      <span :class="['absolute top-0.5 w-4 h-4 rounded-full bg-white shadow transition-all', form.is_reconcilable ? 'left-4.5' : 'left-0.5']" />
                    </button>
                    <span class="text-xs text-slate-400 group-hover:text-slate-200 transition-colors">Reconcilable</span>
                  </label>
                </div>
              </div>

              <!-- Modal footer -->
              <div class="px-6 py-4 border-t border-slate-800/60 flex items-center justify-end gap-3">
                <button @click="showModal=false"
                  class="h-9 px-4 rounded-lg border border-slate-700 bg-slate-800/40 text-slate-400 hover:text-slate-200 text-sm font-medium transition-all">
                  Cancel
                </button>
                <button @click="save" :disabled="saving"
                  class="h-9 px-5 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-sm font-medium inline-flex items-center gap-2 transition-all shadow-lg shadow-indigo-900/30">
                  <Loader2 v-if="saving" class="w-3.5 h-3.5 animate-spin" />
                  <CheckCircle v-else class="w-3.5 h-3.5" />
                  {{ editTarget ? 'Update' : 'Create Account' }}
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<!-- ── Recursive tree row component ──────────────────────────────────────────── -->
<script lang="ts">
import { defineComponent, h, type PropType } from 'vue'
import { ChevronRight, ChevronDown, Pencil, Check, Scale } from 'lucide-vue-next'

const TreeAccountRow = defineComponent({
  name: 'TreeAccountRow',
  props: {
    account: { type: Object as PropType<any>, required: true },
    depth: { type: Number, default: 0 },
    expandedRows: { type: Object as PropType<Set<string>>, required: true },
    typeConfig: { type: Object as PropType<Record<string, any>>, required: true },
    fmtBalance: { type: Function as PropType<(v: number) => string>, required: true },
  },
  emits: ['toggle', 'edit'],
  render() {
    const { account, depth, expandedRows, typeConfig, fmtBalance } = this
    const hasChildren = account.children && account.children.length > 0
    const isExpanded = expandedRows.has(account.id)
    const cfg = typeConfig[account.type] ?? {}
    const pad = depth * 20 + 16

    const rows: any[] = [
      h('tr', {
        class: 'border-b border-slate-800/30 hover:bg-slate-800/20 transition-colors group'
      }, [
        h('td', { class: 'px-4 py-2.5' }, [
          h('div', { class: 'flex items-center', style: `padding-left:${pad}px` }, [
            hasChildren
              ? h('button', {
                  class: 'w-5 h-5 rounded flex items-center justify-center text-slate-500 hover:text-slate-300 mr-1.5 flex-shrink-0',
                  onClick: () => this.$emit('toggle', account.id)
                }, [h(isExpanded ? ChevronDown : ChevronRight, { class: 'w-3.5 h-3.5' })])
              : h('span', { class: 'w-5 mr-1.5' }),
            h('span', { class: 'font-mono text-[12px] font-semibold text-indigo-300 mr-2.5' }, account.code),
            h('span', { class: 'text-[13px] text-slate-200' }, account.name)
          ])
        ]),
        h('td', { class: 'px-4 py-2.5' }, [
          h('span', { class: `inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold ${cfg.bg} ${cfg.color}` }, cfg.label)
        ]),
        h('td', { class: 'px-4 py-2.5 text-right' }, [
          h('span', {
            class: `font-mono text-[12px] font-semibold ${(account.balance ?? 0) >= 0 ? 'text-emerald-400' : 'text-rose-400'}`
          }, fmtBalance(account.balance))
        ]),
        h('td', { class: 'px-4 py-2.5 text-right' }, [
          h('button', {
            class: 'opacity-0 group-hover:opacity-100 inline-flex items-center justify-center w-6 h-6 rounded bg-slate-700/60 hover:bg-indigo-500/20 hover:text-indigo-400 text-slate-400 transition-all',
            onClick: () => this.$emit('edit', account)
          }, [h(Pencil, { class: 'w-3 h-3' })])
        ])
      ])
    ]

    if (isExpanded && hasChildren) {
      for (const child of account.children) {
        rows.push(h(TreeAccountRow, {
          account: child,
          depth: depth + 1,
          expandedRows,
          typeConfig,
          fmtBalance,
          onToggle: (id: string) => this.$emit('toggle', id),
          onEdit: (a: any) => this.$emit('edit', a)
        }))
      }
    }

    return rows
  }
})

export { TreeAccountRow }
</script>
