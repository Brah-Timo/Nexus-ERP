<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Landmark, Plus, RefreshCw, Loader2, CheckCircle, X,
  Calendar, AlertCircle, TrendingUp, TrendingDown, Scale,
  CheckSquare, Square, DollarSign, Clock, Search, ArrowRight
} from 'lucide-vue-next'
import { accountingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ─────────────────────────────────────────────────────────────────────
interface Reconciliation {
  id: string
  bank_account_id: string
  period_date: string
  statement_balance: number
  book_balance: number
  difference: number
  is_reconciled: boolean
  created_at: string
}

// ─── State ─────────────────────────────────────────────────────────────────────
const recs     = ref<Reconciliation[]>([])
const loading  = ref(true)
const saving   = ref(false)
const search   = ref('')
const showModal = ref(false)

const form = ref({
  bank_account_id: '',
  period_date: new Date(new Date().getFullYear(), new Date().getMonth(), 1).toISOString().split('T')[0],
  // period_end: new Date(new Date().getFullYear(), new Date().getMonth() + 1, 0).toISOString().split('T')[0],
  statement_balance: 0,
  book_balance: 0,
  status: 'draft',
  notes: ''
})

// ─── Config ────────────────────────────────────────────────────────────────────
const statusConfig: Record<string, { label: string; color: string; bg: string; border: string; dot: string }> = {
  reconciled:  { label: 'Reconciled',  color: 'text-emerald-400', bg: 'bg-emerald-500/10', border: 'border-emerald-500/20', dot: 'bg-emerald-400' },
  in_progress: { label: 'In Progress', color: 'text-amber-400',   bg: 'bg-amber-500/10',   border: 'border-amber-500/20',   dot: 'bg-amber-400' },
  draft:       { label: 'Draft',       color: 'text-slate-400',   bg: 'bg-slate-500/10',   border: 'border-slate-500/20',   dot: 'bg-slate-400' },
}

// ─── Computed ──────────────────────────────────────────────────────────────────
const stats = computed(() => {
  const total       = recs.value.length
  const reconciled  = recs.value.filter(r => r.is_reconciled).length
  const pending     = recs.value.filter(r => r.is_reconciled ? "Reconciled" : "Open" !== 'reconciled').length
  const totalDiff   = recs.value.reduce((s, r) => s + Math.abs(Number(r.difference) ?? 0), 0)
  return { total, reconciled, pending, totalDiff }
})

const filtered = computed(() => {
  if (!search.value.trim()) return recs.value
  const q = search.value.toLowerCase()
  return recs.value.filter(r =>
    r.bank_account_id?.toLowerCase().includes(q) ||
    r.is_reconciled ? "Reconciled" : "Open"?.toLowerCase().includes(q)
  )
})

const liveDiff = computed(() => {
  const diff = Number(form.value.statement_balance) - Number(form.value.book_balance)
  return diff
})

// ─── Methods ───────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const res = await accountingAPI.getBankReconciliations()
    recs.value = res.data ?? []
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Failed to load reconciliations', 'error')
  } finally {
    loading.value = false
  }
}

async function create() {
  if (!form.value.period_date || !form.value.period_date) {
    app.addToast('Period dates are required', 'error')
    return
  }
  saving.value = true
  try {
    await accountingAPI.createBankReconciliation({
      ...form.value,
      difference: liveDiff.value
    })
    app.addToast('Bank reconciliation created', 'success')
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Create failed', 'error')
  } finally {
    saving.value = false
  }
}

function fmtDate(d: string) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function fmtAmt(v: number) {
  return new Intl.NumberFormat('en-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v ?? 0)
}

function monthName(d: string) {
  if (!d) return ''
  return new Date(d).toLocaleDateString('en-GB', { month: 'long', year: 'numeric' })
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
          <div class="w-9 h-9 rounded-lg bg-sky-500/15 border border-sky-500/25 flex items-center justify-center">
            <Landmark class="w-4.5 h-4.5 text-sky-400" />
          </div>
          <div>
            <h1 class="text-[15px] font-semibold text-slate-100">Bank Reconciliation</h1>
            <p class="text-[11px] text-slate-500">Match bank statements to book records</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            class="h-8 px-3 rounded-lg border border-slate-700/60 bg-slate-800/60 text-slate-400 hover:text-slate-200 text-xs font-medium inline-flex items-center gap-1.5 transition-all disabled:opacity-50">
            <RefreshCw class="w-3.5 h-3.5" :class="loading ? 'animate-spin' : ''" />
            Refresh
          </button>
          <button @click="showModal=true"
            class="h-8 px-3 rounded-lg bg-sky-600 hover:bg-sky-500 text-white text-xs font-medium inline-flex items-center gap-1.5 transition-all shadow-lg shadow-sky-900/30">
            <Plus class="w-3.5 h-3.5" />
            New Reconciliation
          </button>
        </div>
      </div>
    </div>

    <!-- ── Stats ──────────────────────────────────────────────────────────── -->
    <div class="px-6 py-4 flex-shrink-0 grid grid-cols-4 gap-3">
      <div class="rounded-xl bg-slate-900/70 border border-slate-800/50 p-4">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider text-slate-500">Total</span>
          <Landmark class="w-3.5 h-3.5 text-slate-600" />
        </div>
        <div class="text-2xl font-bold text-slate-100">{{ stats.total }}</div>
        <div class="text-[11px] text-slate-500 mt-0.5">All reconciliations</div>
      </div>
      <div class="rounded-xl bg-emerald-500/5 border border-emerald-500/20 p-4">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider text-emerald-500/70">Reconciled</span>
          <CheckSquare class="w-3.5 h-3.5 text-emerald-500/50" />
        </div>
        <div class="text-2xl font-bold text-emerald-400">{{ stats.reconciled }}</div>
        <div class="text-[11px] text-emerald-500/60 mt-0.5">Balanced periods</div>
      </div>
      <div class="rounded-xl bg-amber-500/5 border border-amber-500/20 p-4">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider text-amber-500/70">Pending</span>
          <Clock class="w-3.5 h-3.5 text-amber-500/50" />
        </div>
        <div class="text-2xl font-bold text-amber-400">{{ stats.pending }}</div>
        <div class="text-[11px] text-amber-500/60 mt-0.5">Need review</div>
      </div>
      <div class="rounded-xl bg-rose-500/5 border border-rose-500/20 p-4">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider text-rose-500/70">Total Difference</span>
          <Scale class="w-3.5 h-3.5 text-rose-500/50" />
        </div>
        <div class="text-xl font-bold text-rose-400 truncate">{{ fmtAmt(stats.totalDiff) }}</div>
        <div class="text-[11px] text-rose-500/60 mt-0.5">DZD unreconciled</div>
      </div>
    </div>

    <!-- ── Search ─────────────────────────────────────────────────────────── -->
    <div class="px-6 pb-3 flex-shrink-0 flex items-center gap-3">
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500" />
        <input v-model="search" type="text" placeholder="Search reconciliations..."
          class="h-8 w-64 pl-8 pr-3 rounded-lg bg-slate-900 border border-slate-700/60 text-sm text-slate-200 placeholder-slate-600 focus:outline-none focus:border-sky-500/60 focus:ring-1 focus:ring-sky-500/20 transition-all" />
      </div>
      <div class="ml-auto text-xs text-slate-600">{{ filtered.length }} records</div>
    </div>

    <!-- ── Cards list ─────────────────────────────────────────────────────── -->
    <div class="flex-1 overflow-auto px-6 pb-6">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <Loader2 class="w-7 h-7 text-sky-400 animate-spin" />
      </div>
      <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center py-24 gap-4">
        <div class="w-16 h-16 rounded-2xl bg-sky-500/10 border border-sky-500/20 flex items-center justify-center">
          <Landmark class="w-8 h-8 text-sky-400/50" />
        </div>
        <div class="text-center">
          <p class="text-slate-300 font-medium mb-1">No reconciliations yet</p>
          <p class="text-slate-600 text-sm">Create your first bank reconciliation to start matching records</p>
        </div>
        <button @click="showModal=true"
          class="h-9 px-5 rounded-lg bg-sky-600 hover:bg-sky-500 text-white text-sm font-medium inline-flex items-center gap-2 transition-all">
          <Plus class="w-4 h-4" />
          New Reconciliation
        </button>
      </div>
      <div v-else class="space-y-3">
        <div v-for="rec in filtered" :key="rec.id"
          class="rounded-xl border border-slate-800/60 bg-slate-900/40 hover:border-slate-700/80 transition-all overflow-hidden">
          <!-- Card header -->
          <div class="px-5 py-3 border-b border-slate-800/40 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-sky-500/10 border border-sky-500/20 flex items-center justify-center">
                <Landmark class="w-4 h-4 text-sky-400" />
              </div>
              <div>
                <div class="text-[13px] font-semibold text-slate-200">{{ monthName(rec.period_date) }}</div>
                <div class="text-[11px] text-slate-500">{{ fmtDate(rec.period_date) }} — {{ fmtDate(rec.period_date) }}</div>
              </div>
            </div>
            <span :class="['inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-semibold border', statusConfig[rec.status]?.bg, statusConfig[rec.status]?.color, statusConfig[rec.status]?.border]">
              <span class="w-1.5 h-1.5 rounded-full" :class="statusConfig[rec.status]?.dot" />
              {{ statusConfig[rec.status]?.label ?? rec.status }}
            </span>
          </div>

          <!-- Card body — 3 column balance display -->
          <div class="px-5 py-4 grid grid-cols-3 gap-0 divide-x divide-slate-800/40">
            <!-- Bank balance -->
            <div class="pr-5">
              <div class="text-[10px] font-semibold uppercase tracking-wider text-slate-600 mb-1.5 flex items-center gap-1.5">
                <TrendingUp class="w-3 h-3" />
                Bank Balance
              </div>
              <div class="text-xl font-mono font-bold text-slate-100">{{ fmtAmt(rec.statement_balance) }}</div>
              <div class="text-[11px] text-slate-600 mt-0.5">DZD — Statement</div>
            </div>

            <!-- Book balance -->
            <div class="px-5">
              <div class="text-[10px] font-semibold uppercase tracking-wider text-slate-600 mb-1.5 flex items-center gap-1.5">
                <TrendingDown class="w-3 h-3" />
                Book Balance
              </div>
              <div class="text-xl font-mono font-bold text-slate-100">{{ fmtAmt(rec.book_balance) }}</div>
              <div class="text-[11px] text-slate-600 mt-0.5">DZD — Ledger</div>
            </div>

            <!-- Difference -->
            <div class="pl-5">
              <div class="text-[10px] font-semibold uppercase tracking-wider text-slate-600 mb-1.5 flex items-center gap-1.5">
                <Scale class="w-3 h-3" />
                Difference
              </div>
              <div :class="['text-xl font-mono font-bold', Math.abs(Number(rec.difference)) < 0.01 ? 'text-emerald-400' : 'text-rose-400']">
                {{ fmtAmt(rec.difference) }}
              </div>
              <div :class="['text-[11px] mt-0.5', Math.abs(Number(rec.difference)) < 0.01 ? 'text-emerald-500/60' : 'text-rose-500/60']">
                {{ Math.abs(Number(rec.difference)) < 0.01 ? 'Fully reconciled' : 'Unreconciled amount' }}
              </div>
            </div>
          </div>
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
            <div v-if="showModal" class="relative w-full max-w-md bg-slate-900 border border-slate-700/60 rounded-2xl shadow-2xl overflow-hidden">
              <!-- Header -->
              <div class="px-6 py-4 border-b border-slate-800/60 flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-sky-500/15 border border-sky-500/25 flex items-center justify-center">
                    <Landmark class="w-4 h-4 text-sky-400" />
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-slate-100">New Bank Reconciliation</h3>
                    <p class="text-[11px] text-slate-500">Match bank statement to books</p>
                  </div>
                </div>
                <button @click="showModal=false" class="w-7 h-7 rounded-lg flex items-center justify-center text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-all">
                  <X class="w-4 h-4" />
                </button>
              </div>
              <!-- Body -->
              <div class="px-6 py-5 space-y-4">
                <!-- Bank account -->
                <div>
                  <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Bank Account Reference</label>
                  <input v-model="form.bank_account_id" type="text" placeholder="e.g. BNP-001 or account ID"
                    class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-sky-500/60 focus:ring-1 focus:ring-sky-500/20" />
                </div>
                <!-- Period -->
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Period Start *</label>
                    <input v-model="form.period_date" type="date"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-sky-500/60" />
                  </div>
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Period End *</label>
                    <input v-model="form.period_date" type="date"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-sky-500/60" />
                  </div>
                </div>
                <!-- Balances -->
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Bank Balance (DZD)</label>
                    <input v-model.number="form.statement_balance" type="number" step="0.01"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm font-mono text-slate-100 focus:outline-none focus:border-sky-500/60" />
                  </div>
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Book Balance (DZD)</label>
                    <input v-model.number="form.book_balance" type="number" step="0.01"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm font-mono text-slate-100 focus:outline-none focus:border-sky-500/60" />
                  </div>
                </div>
                <!-- Live difference -->
                <div :class="['rounded-lg px-4 py-3 flex items-center gap-3', Math.abs(liveDiff) < 0.01 ? 'bg-emerald-500/10 border border-emerald-500/20' : 'bg-amber-500/10 border border-amber-500/20']">
                  <CheckCircle v-if="Math.abs(liveDiff) < 0.01" class="w-4 h-4 text-emerald-400" />
                  <AlertCircle v-else class="w-4 h-4 text-amber-400" />
                  <div>
                    <div :class="['text-sm font-semibold', Math.abs(liveDiff) < 0.01 ? 'text-emerald-400' : 'text-amber-400']">
                      Difference: {{ fmtAmt(liveDiff) }} DZD
                    </div>
                    <div class="text-[11px] text-slate-500">
                      {{ Math.abs(liveDiff) < 0.01 ? 'Fully reconciled — bank and book match' : 'Unreconciled amount needs investigation' }}
                    </div>
                  </div>
                </div>
                <!-- Status -->
                <div>
                  <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Status</label>
                  <select v-model="form.status"
                    class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-sky-500/60">
                    <option value="draft">Draft</option>
                    <option value="in_progress">In Progress</option>
                    <option value="reconciled">Reconciled</option>
                  </select>
                </div>
              </div>
              <!-- Footer -->
              <div class="px-6 py-4 border-t border-slate-800/60 flex items-center justify-end gap-3">
                <button @click="showModal=false" class="h-9 px-4 rounded-lg border border-slate-700 bg-slate-800/40 text-slate-400 hover:text-slate-200 text-sm font-medium transition-all">Cancel</button>
                <button @click="create" :disabled="saving"
                  class="h-9 px-5 rounded-lg bg-sky-600 hover:bg-sky-500 disabled:opacity-50 text-white text-sm font-medium inline-flex items-center gap-2 transition-all shadow-lg shadow-sky-900/30">
                  <Loader2 v-if="saving" class="w-3.5 h-3.5 animate-spin" />
                  <CheckCircle v-else class="w-3.5 h-3.5" />
                  Create
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
