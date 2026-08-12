<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  BookOpen, Plus, Search, RefreshCw, Filter, Loader2, CheckCircle,
  X, ChevronLeft, ChevronRight, AlertCircle, Eye, Check, Ban,
  ArrowUpDown, Calendar, Hash, FileText, Tag, ArrowRight,
  TrendingUp, DollarSign, Clock, Send, AlertTriangle,
  Minus, SlidersHorizontal, ChevronDown
} from 'lucide-vue-next'
import { accountingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ─────────────────────────────────────────────────────────────────────
interface JournalEntry {
  id: string
  number: string
  date: string
  description: string
  reference: string
  source_type: string
  status: string
  total_debit: number
  total_credit: number
  created_at: string
  lines?: JournalLine[]
}

interface JournalLine {
  id: string
  account_id: string
  account_code: string
  account_name: string
  description: string
  debit: number
  credit: number
  cost_center_id?: string
}

interface Account {
  id: string
  code: string
  name: string
  type: string
  is_group: boolean
}

// ─── State ─────────────────────────────────────────────────────────────────────
const entries     = ref<JournalEntry[]>([])
const accounts    = ref<Account[]>([])
const loading     = ref(true)
const saving      = ref(false)
const postingId   = ref<string | null>(null)
const cancellingId = ref<string | null>(null)
const search      = ref('')
const statusFilter = ref('')
const journalFilter = ref('')
const dateFrom    = ref('')
const dateTo      = ref('')
const page        = ref(1)
const perPage     = 25
const showModal   = ref(false)
const showDrawer  = ref(false)
const selectedEntry = ref<JournalEntry | null>(null)
const drawerLoading = ref(false)

// new entry form
const newEntry = ref({
  date: new Date().toISOString().split('T')[0],
  description: '',
  reference: '',
  source_type: 'manual',
  lines: [
    { account_id: '', account_code: '', account_name: '', description: '', debit: 0, credit: 0 },
    { account_id: '', account_code: '', account_name: '', description: '', debit: 0, credit: 0 },
  ] as JournalLine[]
})

// ─── Constants ─────────────────────────────────────────────────────────────────
const journals = ['general', 'sales', 'purchase', 'cash', 'bank', 'payroll', 'asset', 'opening']
const statusConfig: Record<string, { label: string; color: string; bg: string; border: string; dot: string }> = {
  draft:     { label: 'Draft',     color: 'text-slate-400',   bg: 'bg-slate-500/10',   border: 'border-slate-500/20',   dot: 'bg-slate-400' },
  posted:    { label: 'Posted',    color: 'text-emerald-400', bg: 'bg-emerald-500/10', border: 'border-emerald-500/20', dot: 'bg-emerald-400' },
  cancelled: { label: 'Cancelled', color: 'text-rose-400',    bg: 'bg-rose-500/10',    border: 'border-rose-500/20',    dot: 'bg-rose-400' },
}

// ─── Computed ──────────────────────────────────────────────────────────────────
const stats = computed(() => {
  const total = entries.value.length
  const draft = entries.value.filter(e => e.status === 'draft').length
  const posted = entries.value.filter(e => e.status === 'posted').length
  const postedVol = entries.value.filter(e => e.status === 'posted').reduce((s, e) => s + (e.total_debit ?? 0), 0)
  return { total, draft, posted, postedVol }
})

const filteredEntries = computed(() => {
  let list = [...entries.value]
  if (statusFilter.value) list = list.filter(e => e.status === statusFilter.value)
  if (journalFilter.value) list = list.filter(e => e.source_type === journalFilter.value || e.reference === journalFilter.value)
  if (dateFrom.value) list = list.filter(e => e.date >= dateFrom.value)
  if (dateTo.value) list = list.filter(e => e.date <= dateTo.value)
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(e =>
      e.number.toLowerCase().includes(q) ||
      e.description.toLowerCase().includes(q) ||
      e.reference.toLowerCase().includes(q)
    )
  }
  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filteredEntries.value.length / perPage)))
const pagedEntries = computed(() => {
  const start = (page.value - 1) * perPage
  return filteredEntries.value.slice(start, start + perPage)
})

const newBalance = computed(() => {
  const debit = newEntry.value.lines.reduce((s, l) => s + (Number(l.debit) || 0), 0)
  const credit = newEntry.value.lines.reduce((s, l) => s + (Number(l.credit) || 0), 0)
  return { debit, credit, diff: debit - credit, balanced: Math.abs(debit - credit) < 0.001 }
})

// ─── Methods ───────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const params: Record<string, string> = {}
    if (statusFilter.value) params.status = statusFilter.value
    const [entriesRes, accountsRes] = await Promise.all([
      accountingAPI.getJournalEntries(params),
      accountingAPI.getChartOfAccounts()
    ])
    entries.value = entriesRes.data ?? []
    accounts.value = (accountsRes.data ?? []).filter((a: Account) => !a.is_group)
    page.value = 1
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Failed to load journal entries', 'error')
  } finally {
    loading.value = false
  }
}

async function openDrawer(entry: JournalEntry) {
  selectedEntry.value = { ...entry }
  showDrawer.value = true
  if (!entry.lines) {
    drawerLoading.value = true
    try {
      const res = await accountingAPI.getJournalEntry(entry.id)
      selectedEntry.value = res.data
    } catch { /* ignore */ } finally {
      drawerLoading.value = false
    }
  }
}

async function postEntry(id: string) {
  postingId.value = id
  try {
    await accountingAPI.postJournalEntry(id)
    app.addToast('Journal entry posted successfully', 'success')
    await load()
    if (selectedEntry.value?.id === id) showDrawer.value = false
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Failed to post entry', 'error')
  } finally {
    postingId.value = null
  }
}

async function cancelEntry(id: string) {
  cancellingId.value = id
  try {
    await accountingAPI.cancelJournalEntry(id)
    app.addToast('Journal entry cancelled', 'success')
    await load()
    if (selectedEntry.value?.id === id) showDrawer.value = false
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Failed to cancel entry', 'error')
  } finally {
    cancellingId.value = null
  }
}

function addLine() {
  newEntry.value.lines.push({ account_id: '', account_code: '', account_name: '', description: '', debit: 0, credit: 0 })
}

function removeLine(i: number) {
  if (newEntry.value.lines.length > 2) newEntry.value.lines.splice(i, 1)
}

function setLineAccount(i: number, accountId: string) {
  const acc = accounts.value.find(a => a.id === accountId)
  if (acc) {
    newEntry.value.lines[i].account_id = acc.id
    newEntry.value.lines[i].account_code = acc.code
    newEntry.value.lines[i].account_name = acc.name
  }
}

async function createEntry() {
  if (!newBalance.value.balanced) {
    app.addToast('Entry must be balanced (debit = credit)', 'error')
    return
  }
  const validLines = newEntry.value.lines.filter(l => l.account_id && (Number(l.debit) > 0 || Number(l.credit) > 0))
  if (validLines.length < 2) {
    app.addToast('At least 2 lines with amounts are required', 'error')
    return
  }
  saving.value = true
  try {
    await accountingAPI.createJournalEntry({
      date: newEntry.value.date,
      description: newEntry.value.description,
      reference: newEntry.value.reference,
        source_type: newEntry.value.source_type,
      lines: validLines.map(l => ({ ...l, debit: Number(l.debit), credit: Number(l.credit) }))
    })
    app.addToast('Journal entry created', 'success')
    showModal.value = false
    newEntry.value = {
      date: new Date().toISOString().split('T')[0],
      description: '',
      reference: '',
  source_type: 'manual',
      lines: [
        { account_id: '', account_code: '', account_name: '', description: '', debit: 0, credit: 0 },
        { account_id: '', account_code: '', account_name: '', description: '', debit: 0, credit: 0 },
      ]
    }
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Failed to create entry', 'error')
  } finally {
    saving.value = false
  }
}

function fmtDate(d: string) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function fmtAmt(v: number) {
  if (v == null || v === 0) return '—'
  return new Intl.NumberFormat('en-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)
}

function fmtAmtRaw(v: number) {
  return new Intl.NumberFormat('en-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v ?? 0)
}

watch([statusFilter, journalFilter], () => { page.value = 1 })

onMounted(load)
</script>

<template>
  <div class="flex flex-col h-full transition-colors duration-200"
     :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">

    <!-- ── Header ─────────────────────────────────────────────────────────── -->
    <div class="border-b border-slate-800/60 bg-slate-900/50 backdrop-blur-sm px-6 py-4 flex-shrink-0">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-lg bg-violet-500/15 border border-violet-500/25 flex items-center justify-center">
            <BookOpen class="w-4.5 h-4.5 text-violet-400" />
          </div>
          <div>
            <h1 class="text-[15px] font-semibold text-slate-100">Journal Entries</h1>
            <p class="text-[11px] text-slate-500">Double-entry bookkeeping ledger</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            class="h-8 px-3 rounded-lg border border-slate-700/60 bg-slate-800/60 text-slate-400 hover:text-slate-200 hover:border-slate-600 text-xs font-medium inline-flex items-center gap-1.5 transition-all disabled:opacity-50">
            <RefreshCw class="w-3.5 h-3.5" :class="loading ? 'animate-spin' : ''" />
            Refresh
          </button>
          <button @click="showModal=true"
            class="h-8 px-3 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-xs font-medium inline-flex items-center gap-1.5 transition-all shadow-lg shadow-violet-900/30">
            <Plus class="w-3.5 h-3.5" />
            New Entry
          </button>
        </div>
      </div>
    </div>

    <!-- ── Stats row ──────────────────────────────────────────────────────── -->
    <div class="px-6 py-4 flex-shrink-0 grid grid-cols-4 gap-3">
      <div class="rounded-xl bg-slate-900/70 border border-slate-800/50 p-4">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider text-slate-500">Total Entries</span>
          <Hash class="w-3.5 h-3.5 text-slate-600" />
        </div>
        <div class="text-2xl font-bold text-slate-100">{{ stats.total }}</div>
        <div class="text-[11px] text-slate-500 mt-0.5">All journals combined</div>
      </div>
      <div class="rounded-xl bg-slate-900/70 border border-slate-800/50 p-4">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider text-slate-500">Draft</span>
          <Clock class="w-3.5 h-3.5 text-amber-500/60" />
        </div>
        <div class="text-2xl font-bold text-amber-400">{{ stats.draft }}</div>
        <div class="text-[11px] text-amber-500/60 mt-0.5">Awaiting posting</div>
      </div>
      <div class="rounded-xl bg-slate-900/70 border border-slate-800/50 p-4">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider text-slate-500">Posted</span>
          <CheckCircle class="w-3.5 h-3.5 text-emerald-500/60" />
        </div>
        <div class="text-2xl font-bold text-emerald-400">{{ stats.posted }}</div>
        <div class="text-[11px] text-emerald-500/60 mt-0.5">Confirmed entries</div>
      </div>
      <div class="rounded-xl bg-slate-900/70 border border-slate-800/50 p-4">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider text-slate-500">Posted Volume</span>
          <TrendingUp class="w-3.5 h-3.5 text-violet-500/60" />
        </div>
        <div class="text-xl font-bold text-violet-400 truncate">{{ fmtAmtRaw(stats.postedVol) }}</div>
        <div class="text-[11px] text-violet-500/60 mt-0.5">DZD total debits</div>
      </div>
    </div>

    <!-- ── Filters ────────────────────────────────────────────────────────── -->
    <div class="px-6 pb-3 flex-shrink-0 flex items-center gap-3 flex-wrap">
      <!-- Search -->
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500" />
        <input v-model="search" type="text" placeholder="Search number, description..."
          class="h-8 w-56 pl-8 pr-3 rounded-lg bg-slate-900 border border-slate-700/60 text-sm text-slate-200 placeholder-slate-600 focus:outline-none focus:border-violet-500/60 focus:ring-1 focus:ring-violet-500/20 transition-all" />
        <button v-if="search" @click="search=''" class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-600 hover:text-slate-400">
          <X class="w-3 h-3" />
        </button>
      </div>

      <!-- Status -->
      <select v-model="statusFilter"
        class="h-8 px-3 rounded-lg bg-slate-900 border border-slate-700/60 text-xs text-slate-300 focus:outline-none focus:border-violet-500/60 transition-all">
        <option value="">All Statuses</option>
        <option v-for="(cfg, s) in statusConfig" :key="s" :value="s">{{ cfg.label }}</option>
      </select>

      <!-- Journal -->
      <select v-model="journalFilter"
        class="h-8 px-3 rounded-lg bg-slate-900 border border-slate-700/60 text-xs text-slate-300 focus:outline-none focus:border-violet-500/60 transition-all capitalize">
        <option value="">All Journals</option>
        <option v-for="j in journals" :key="j" :value="j">{{ j.charAt(0).toUpperCase() + j.slice(1) }}</option>
      </select>

      <!-- Date range -->
      <input v-model="dateFrom" type="date"
        class="h-8 px-3 rounded-lg bg-slate-900 border border-slate-700/60 text-xs text-slate-300 focus:outline-none focus:border-violet-500/60 transition-all" />
      <span class="text-slate-600 text-xs">to</span>
      <input v-model="dateTo" type="date"
        class="h-8 px-3 rounded-lg bg-slate-900 border border-slate-700/60 text-xs text-slate-300 focus:outline-none focus:border-violet-500/60 transition-all" />

      <button v-if="search||statusFilter||journalFilter||dateFrom||dateTo"
        @click="search='';statusFilter='';journalFilter='';dateFrom='';dateTo=''"
        class="h-8 px-3 rounded-lg border border-rose-500/30 bg-rose-500/10 text-rose-400 text-xs hover:bg-rose-500/20 transition-all inline-flex items-center gap-1.5">
        <X class="w-3 h-3" /> Clear
      </button>

      <div class="ml-auto text-xs text-slate-600">
        {{ filteredEntries.length }} entries &middot; page {{ page }} / {{ totalPages }}
      </div>
    </div>

    <!-- ── Table ──────────────────────────────────────────────────────────── -->
    <div class="flex-1 overflow-hidden px-6 pb-4">
      <div class="h-full flex flex-col rounded-xl border border-slate-800/60 overflow-hidden bg-slate-900/40">
        <div class="flex-1 overflow-auto">
          <table class="w-full text-sm border-collapse">
            <thead class="sticky top-0 z-10">
              <tr class="bg-slate-900/90 backdrop-blur border-b border-slate-800/60">
                <th class="text-left px-4 py-3 w-32 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Number</th>
                <th class="text-left px-4 py-3 w-28 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Date</th>
                <th class="text-left px-4 py-3 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Description</th>
                <th class="text-left px-4 py-3 w-24 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Reference</th>
                <th class="text-left px-4 py-3 w-24 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Status</th>
                <th class="text-right px-4 py-3 w-36 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Debit</th>
                <th class="text-right px-4 py-3 w-36 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Credit</th>
                <th class="px-4 py-3 w-28 text-[11px] font-semibold uppercase tracking-wider text-slate-500 text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              <template v-if="loading">
                <tr><td colspan="8" class="py-20 text-center"><Loader2 class="w-6 h-6 text-violet-400 animate-spin mx-auto" /></td></tr>
              </template>
              <template v-else-if="pagedEntries.length === 0">
                <tr><td colspan="8" class="py-20 text-center">
                  <BookOpen class="w-10 h-10 mx-auto mb-3 text-slate-700" />
                  <p class="text-sm text-slate-600">No journal entries found</p>
                </td></tr>
              </template>
              <tr v-for="(entry, i) in pagedEntries" :key="entry.id"
                :class="['border-b border-slate-800/30 hover:bg-slate-800/30 transition-colors group cursor-pointer', i%2===0?'':'bg-slate-900/20']"
                @click="openDrawer(entry)">
                <!-- Number -->
                <td class="px-4 py-3">
                  <span class="font-mono text-[12px] font-semibold text-violet-300">{{ entry.number }}</span>
                </td>
                <!-- Date -->
                <td class="px-4 py-3">
                  <span class="text-[12px] text-slate-400">{{ fmtDate(entry.date) }}</span>
                </td>
                <!-- Description -->
                <td class="px-4 py-3">
                  <span class="text-[13px] text-slate-200 line-clamp-1">{{ entry.description || '—' }}</span>
                </td>
                <!-- Journal -->
                <td class="px-4 py-3">
                  <span class="text-[11px] bg-slate-800/80 text-slate-400 px-1.5 py-0.5 rounded capitalize">{{ entry.reference || entry.source_type || "—" }}</span>
                </td>
                <!-- Status -->
                <td class="px-4 py-3" @click.stop>
                  <span :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold border', statusConfig[entry.status]?.bg, statusConfig[entry.status]?.color, statusConfig[entry.status]?.border]">
                    <span class="w-1.5 h-1.5 rounded-full" :class="statusConfig[entry.status]?.dot" />
                    {{ statusConfig[entry.status]?.label }}
                  </span>
                </td>
                <!-- Debit -->
                <td class="px-4 py-3 text-right">
                  <span class="font-mono text-[12px] text-emerald-400">{{ fmtAmt(entry.total_debit) }}</span>
                </td>
                <!-- Credit -->
                <td class="px-4 py-3 text-right">
                  <span class="font-mono text-[12px] text-rose-400">{{ fmtAmt(entry.total_credit) }}</span>
                </td>
                <!-- Actions -->
                <td class="px-4 py-3 text-right" @click.stop>
                  <div class="flex items-center justify-end gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button @click="openDrawer(entry)" title="View"
                      class="w-7 h-7 rounded-lg bg-slate-700/60 hover:bg-violet-500/20 hover:text-violet-400 text-slate-400 inline-flex items-center justify-center transition-all">
                      <Eye class="w-3.5 h-3.5" />
                    </button>
                    <button v-if="entry.status==='draft'" @click="postEntry(entry.id)" :disabled="postingId===entry.id" title="Post"
                      class="w-7 h-7 rounded-lg bg-slate-700/60 hover:bg-emerald-500/20 hover:text-emerald-400 text-slate-400 inline-flex items-center justify-center transition-all disabled:opacity-50">
                      <Loader2 v-if="postingId===entry.id" class="w-3.5 h-3.5 animate-spin" />
                      <Send v-else class="w-3.5 h-3.5" />
                    </button>
                    <button v-if="entry.status==='draft'" @click="cancelEntry(entry.id)" :disabled="cancellingId===entry.id" title="Cancel"
                      class="w-7 h-7 rounded-lg bg-slate-700/60 hover:bg-rose-500/20 hover:text-rose-400 text-slate-400 inline-flex items-center justify-center transition-all disabled:opacity-50">
                      <Ban v-if="cancellingId!==entry.id" class="w-3.5 h-3.5" />
                      <Loader2 v-else class="w-3.5 h-3.5 animate-spin" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div class="flex-shrink-0 border-t border-slate-800/60 px-4 py-2.5 flex items-center justify-between bg-slate-900/60">
          <span class="text-xs text-slate-600">
            Showing {{ ((page-1)*perPage)+1 }}–{{ Math.min(page*perPage, filteredEntries.length) }} of {{ filteredEntries.length }}
          </span>
          <div class="flex items-center gap-1">
            <button @click="page=Math.max(1,page-1)" :disabled="page===1"
              class="w-7 h-7 rounded-lg border border-slate-700/60 bg-slate-800/40 text-slate-400 hover:text-slate-200 disabled:opacity-30 inline-flex items-center justify-center transition-all">
              <ChevronLeft class="w-3.5 h-3.5" />
            </button>
            <span v-for="p in Math.min(7, totalPages)" :key="p">
              <button @click="page=p" :class="['w-7 h-7 rounded-lg text-xs font-medium transition-all', page===p?'bg-violet-600 text-white':'border border-slate-700/60 bg-slate-800/40 text-slate-500 hover:text-slate-300']">{{ p }}</button>
            </span>
            <span v-if="totalPages > 7" class="text-slate-600 text-xs px-1">...</span>
            <button @click="page=Math.min(totalPages,page+1)" :disabled="page===totalPages"
              class="w-7 h-7 rounded-lg border border-slate-700/60 bg-slate-800/40 text-slate-400 hover:text-slate-200 disabled:opacity-30 inline-flex items-center justify-center transition-all">
              <ChevronRight class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ── Detail Drawer ──────────────────────────────────────────────────── -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-all duration-300"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-all duration-200"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0">
        <div v-if="showDrawer" class="fixed inset-0 z-50 flex">
          <div class="flex-1 bg-slate-950/60 backdrop-blur-sm" @click="showDrawer=false" />
          <Transition
            enter-active-class="transition-all duration-300"
            enter-from-class="translate-x-full"
            enter-to-class="translate-x-0"
            leave-active-class="transition-all duration-200"
            leave-from-class="translate-x-0"
            leave-to-class="translate-x-full">
            <div v-if="showDrawer" class="w-[600px] max-w-full flex flex-col bg-slate-900 border-l border-slate-800/60 shadow-2xl">
              <!-- Drawer header -->
              <div class="px-6 py-4 border-b border-slate-800/60 flex items-center justify-between flex-shrink-0">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-violet-500/15 border border-violet-500/25 flex items-center justify-center">
                    <FileText class="w-4 h-4 text-violet-400" />
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-slate-100">{{ selectedEntry?.number }}</h3>
                    <p class="text-[11px] text-slate-500">{{ fmtDate(selectedEntry?.date ?? '') }}</p>
                  </div>
                  <span v-if="selectedEntry" :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold border', statusConfig[selectedEntry.status]?.bg, statusConfig[selectedEntry.status]?.color, statusConfig[selectedEntry.status]?.border]">
                    <span class="w-1.5 h-1.5 rounded-full" :class="statusConfig[selectedEntry?.status]?.dot" />
                    {{ statusConfig[selectedEntry?.status]?.label }}
                  </span>
                </div>
                <button @click="showDrawer=false" class="w-7 h-7 rounded-lg flex items-center justify-center text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-all">
                  <X class="w-4 h-4" />
                </button>
              </div>

              <!-- Drawer body -->
              <div class="flex-1 overflow-auto px-6 py-5">
                <div v-if="drawerLoading" class="flex items-center justify-center py-20">
                  <Loader2 class="w-6 h-6 text-violet-400 animate-spin" />
                </div>
                <template v-else-if="selectedEntry">
                  <!-- Meta grid -->
                  <div class="grid grid-cols-2 gap-3 mb-5">
                    <div class="rounded-lg bg-slate-800/40 border border-slate-700/40 p-3">
                      <div class="text-[10px] text-slate-600 mb-1 uppercase tracking-wider font-semibold">Journal</div>
                      <div class="text-sm text-slate-200 capitalize">{{ selectedEntry.reference || selectedEntry.source_type || "—" }}</div>
                    </div>
                    <div class="rounded-lg bg-slate-800/40 border border-slate-700/40 p-3">
                      <div class="text-[10px] text-slate-600 mb-1 uppercase tracking-wider font-semibold">Description</div>
                      <div class="text-sm text-slate-200">{{ selectedEntry.description || '—' }}</div>
                    </div>
                    <div class="rounded-lg bg-slate-800/40 border border-slate-700/40 p-3">
                      <div class="text-[10px] text-slate-600 mb-1 uppercase tracking-wider font-semibold">Total Debit</div>
                      <div class="text-sm font-mono font-semibold text-emerald-400">{{ fmtAmtRaw(selectedEntry.total_debit) }}</div>
                    </div>
                    <div class="rounded-lg bg-slate-800/40 border border-slate-700/40 p-3">
                      <div class="text-[10px] text-slate-600 mb-1 uppercase tracking-wider font-semibold">Total Credit</div>
                      <div class="text-sm font-mono font-semibold text-rose-400">{{ fmtAmtRaw(selectedEntry.total_credit) }}</div>
                    </div>
                  </div>

                  <!-- Balance indicator -->
                  <div :class="['rounded-lg px-3 py-2 mb-5 flex items-center gap-2 text-xs font-medium', Math.abs((selectedEntry.total_debit??0)-(selectedEntry.total_credit??0)) < 0.001 ? 'bg-emerald-500/10 border border-emerald-500/20 text-emerald-400' : 'bg-rose-500/10 border border-rose-500/20 text-rose-400']">
                    <CheckCircle v-if="Math.abs((selectedEntry.total_debit??0)-(selectedEntry.total_credit??0)) < 0.001" class="w-3.5 h-3.5" />
                    <AlertCircle v-else class="w-3.5 h-3.5" />
                    {{ Math.abs((selectedEntry.total_debit??0)-(selectedEntry.total_credit??0)) < 0.001 ? 'Entry is balanced' : 'Entry is NOT balanced' }}
                  </div>

                  <!-- Lines table -->
                  <div v-if="selectedEntry.lines && selectedEntry.lines.length > 0">
                    <h4 class="text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-2">Journal Lines</h4>
                    <div class="rounded-lg border border-slate-800/60 overflow-hidden">
                      <table class="w-full text-xs">
                        <thead>
                          <tr class="bg-slate-800/60 border-b border-slate-700/40">
                            <th class="text-left px-3 py-2 text-slate-500 font-semibold">Account</th>
                            <th class="text-left px-3 py-2 text-slate-500 font-semibold">Description</th>
                            <th class="text-right px-3 py-2 text-slate-500 font-semibold">Debit</th>
                            <th class="text-right px-3 py-2 text-slate-500 font-semibold">Credit</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr v-for="line in selectedEntry.lines" :key="line.id"
                            class="border-b border-slate-800/30 hover:bg-slate-800/20">
                            <td class="px-3 py-2">
                              <div class="font-mono text-violet-300 text-[11px]">{{ line.account_code }}</div>
                              <div class="text-slate-400 text-[10px]">{{ line.account_name }}</div>
                            </td>
                            <td class="px-3 py-2 text-slate-500">{{ line.description || '—' }}</td>
                            <td class="px-3 py-2 text-right font-mono">
                              <span v-if="line.debit > 0" class="text-emerald-400">{{ fmtAmtRaw(line.debit) }}</span>
                              <span v-else class="text-slate-700">—</span>
                            </td>
                            <td class="px-3 py-2 text-right font-mono">
                              <span v-if="line.credit > 0" class="text-rose-400">{{ fmtAmtRaw(line.credit) }}</span>
                              <span v-else class="text-slate-700">—</span>
                            </td>
                          </tr>
                        </tbody>
                        <tfoot>
                          <tr class="bg-slate-800/40 border-t border-slate-700/40">
                            <td colspan="2" class="px-3 py-2 text-[11px] font-semibold text-slate-400">Totals</td>
                            <td class="px-3 py-2 text-right font-mono text-[11px] font-semibold text-emerald-400">{{ fmtAmtRaw(selectedEntry.total_debit) }}</td>
                            <td class="px-3 py-2 text-right font-mono text-[11px] font-semibold text-rose-400">{{ fmtAmtRaw(selectedEntry.total_credit) }}</td>
                          </tr>
                        </tfoot>
                      </table>
                    </div>
                  </div>
                </template>
              </div>

              <!-- Drawer footer -->
              <div v-if="selectedEntry" class="flex-shrink-0 px-6 py-4 border-t border-slate-800/60 flex items-center gap-3">
                <button v-if="selectedEntry.status==='draft'" @click="postEntry(selectedEntry.id)" :disabled="!!postingId"
                  class="h-9 px-4 rounded-lg bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white text-sm font-medium inline-flex items-center gap-2 transition-all">
                  <Loader2 v-if="postingId" class="w-3.5 h-3.5 animate-spin" />
                  <Send v-else class="w-3.5 h-3.5" />
                  Post Entry
                </button>
                <button v-if="selectedEntry.status==='draft'" @click="cancelEntry(selectedEntry.id)" :disabled="!!cancellingId"
                  class="h-9 px-4 rounded-lg border border-rose-500/30 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 text-sm font-medium inline-flex items-center gap-2 transition-all disabled:opacity-50">
                  <Ban class="w-3.5 h-3.5" />
                  Cancel
                </button>
                <button @click="showDrawer=false" class="ml-auto h-9 px-4 rounded-lg border border-slate-700 bg-slate-800/40 text-slate-400 hover:text-slate-200 text-sm font-medium transition-all">
                  Close
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>

    <!-- ── New Entry Modal ────────────────────────────────────────────────── -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-all duration-200"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-all duration-150"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-start justify-center p-4 pt-12 overflow-y-auto">
          <div class="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" @click="showModal=false" />
          <div class="relative w-full max-w-3xl bg-slate-900 border border-slate-700/60 rounded-2xl shadow-2xl overflow-hidden mb-8">

            <!-- Header -->
            <div class="px-6 py-4 border-b border-slate-800/60 flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-lg bg-violet-500/15 border border-violet-500/25 flex items-center justify-center">
                  <BookOpen class="w-4 h-4 text-violet-400" />
                </div>
                <div>
                  <h3 class="text-sm font-semibold text-slate-100">New Journal Entry</h3>
                  <p class="text-[11px] text-slate-500">Create a balanced double-entry transaction</p>
                </div>
              </div>
              <button @click="showModal=false" class="w-7 h-7 rounded-lg flex items-center justify-center text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-all">
                <X class="w-4 h-4" />
              </button>
            </div>

            <!-- Body -->
            <div class="px-6 py-5 space-y-4">
              <!-- Entry meta -->
              <div class="grid grid-cols-3 gap-4">
                <div>
                  <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Date *</label>
                  <input v-model="newEntry.date" type="date"
                    class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-violet-500/60 focus:ring-1 focus:ring-violet-500/20" />
                </div>
                <div>
                  <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Journal *</label>
                  <select v-model="newEntry.reference"
                    class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-violet-500/60 focus:ring-1 focus:ring-violet-500/20 capitalize">
                    <option v-for="j in journals" :key="j" :value="j">{{ j.charAt(0).toUpperCase() + j.slice(1) }}</option>
                  </select>
                </div>
                <div class="col-span-1">
                  <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Description</label>
                  <input v-model="newEntry.description" type="text" placeholder="Transaction description..."
                    class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500/60 focus:ring-1 focus:ring-violet-500/20" />
                </div>
              </div>

              <!-- Balance indicator -->
              <div :class="['rounded-lg px-4 py-3 flex items-center gap-3 text-sm', newBalance.balanced ? 'bg-emerald-500/10 border border-emerald-500/20' : 'bg-amber-500/10 border border-amber-500/20']">
                <CheckCircle v-if="newBalance.balanced" class="w-4 h-4 text-emerald-400 flex-shrink-0" />
                <AlertTriangle v-else class="w-4 h-4 text-amber-400 flex-shrink-0" />
                <div class="flex-1 flex items-center gap-4">
                  <span :class="newBalance.balanced ? 'text-emerald-400' : 'text-amber-400'">
                    {{ newBalance.balanced ? 'Balanced' : 'Not balanced' }}
                  </span>
                  <span class="text-slate-500 text-xs">Debit: <span class="font-mono text-emerald-400">{{ fmtAmtRaw(newBalance.debit) }}</span></span>
                  <span class="text-slate-500 text-xs">Credit: <span class="font-mono text-rose-400">{{ fmtAmtRaw(newBalance.credit) }}</span></span>
                  <span v-if="!newBalance.balanced" class="text-xs text-amber-400">Diff: <span class="font-mono">{{ fmtAmtRaw(Math.abs(newBalance.diff)) }}</span></span>
                </div>
              </div>

              <!-- Lines -->
              <div>
                <div class="flex items-center justify-between mb-2">
                  <label class="text-[11px] font-semibold uppercase tracking-wider text-slate-500">Journal Lines</label>
                  <button @click="addLine" class="h-7 px-2.5 rounded-lg bg-slate-800 border border-slate-700/60 text-slate-400 hover:text-slate-200 text-xs inline-flex items-center gap-1 transition-all">
                    <Plus class="w-3 h-3" /> Add Line
                  </button>
                </div>
                <div class="rounded-lg border border-slate-800/60 overflow-hidden">
                  <table class="w-full text-sm">
                    <thead>
                      <tr class="bg-slate-800/60 border-b border-slate-700/40">
                        <th class="text-left px-3 py-2 text-[10px] font-semibold uppercase tracking-wider text-slate-500 w-56">Account</th>
                        <th class="text-left px-3 py-2 text-[10px] font-semibold uppercase tracking-wider text-slate-500">Description</th>
                        <th class="text-left px-3 py-2 text-[10px] font-semibold uppercase tracking-wider text-slate-500 w-32">Debit</th>
                        <th class="text-left px-3 py-2 text-[10px] font-semibold uppercase tracking-wider text-slate-500 w-32">Credit</th>
                        <th class="px-3 py-2 w-8" />
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(line, i) in newEntry.lines" :key="i" class="border-b border-slate-800/30">
                        <td class="px-2 py-1.5">
                          <select :value="line.account_id" @change="setLineAccount(i, ($event.target as HTMLSelectElement).value)"
                            class="w-full h-8 px-2 rounded-lg bg-slate-800/80 border border-slate-700/60 text-xs text-slate-200 focus:outline-none focus:border-violet-500/60 transition-all">
                            <option value="">Select account...</option>
                            <option v-for="acc in accounts" :key="acc.id" :value="acc.id">{{ acc.code }} — {{ acc.name }}</option>
                          </select>
                        </td>
                        <td class="px-2 py-1.5">
                          <input v-model="line.description" type="text" placeholder="Note..."
                            class="w-full h-8 px-2 rounded-lg bg-slate-800/80 border border-slate-700/60 text-xs text-slate-200 placeholder-slate-600 focus:outline-none focus:border-violet-500/60 transition-all" />
                        </td>
                        <td class="px-2 py-1.5">
                          <input v-model.number="line.debit" type="number" min="0" step="0.01" placeholder="0.00"
                            @input="Number(line.debit) > 0 ? (line.credit = 0) : null"
                            class="w-full h-8 px-2 rounded-lg bg-slate-800/80 border border-slate-700/60 text-xs font-mono text-emerald-400 placeholder-slate-700 focus:outline-none focus:border-emerald-500/40 transition-all" />
                        </td>
                        <td class="px-2 py-1.5">
                          <input v-model.number="line.credit" type="number" min="0" step="0.01" placeholder="0.00"
                            @input="Number(line.credit) > 0 ? (line.debit = 0) : null"
                            class="w-full h-8 px-2 rounded-lg bg-slate-800/80 border border-slate-700/60 text-xs font-mono text-rose-400 placeholder-slate-700 focus:outline-none focus:border-rose-500/40 transition-all" />
                        </td>
                        <td class="px-2 py-1.5">
                          <button @click="removeLine(i)" :disabled="newEntry.lines.length <= 2"
                            class="w-7 h-7 rounded-lg flex items-center justify-center text-slate-600 hover:text-rose-400 hover:bg-rose-500/10 disabled:opacity-20 transition-all">
                            <X class="w-3.5 h-3.5" />
                          </button>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

            <!-- Footer -->
            <div class="px-6 py-4 border-t border-slate-800/60 flex items-center justify-end gap-3">
              <button @click="showModal=false" class="h-9 px-4 rounded-lg border border-slate-700 bg-slate-800/40 text-slate-400 hover:text-slate-200 text-sm font-medium transition-all">Cancel</button>
              <button @click="createEntry" :disabled="saving||!newBalance.balanced"
                class="h-9 px-5 rounded-lg bg-violet-600 hover:bg-violet-500 disabled:opacity-50 text-white text-sm font-medium inline-flex items-center gap-2 transition-all shadow-lg shadow-violet-900/30">
                <Loader2 v-if="saving" class="w-3.5 h-3.5 animate-spin" />
                <CheckCircle v-else class="w-3.5 h-3.5" />
                Create Entry
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
