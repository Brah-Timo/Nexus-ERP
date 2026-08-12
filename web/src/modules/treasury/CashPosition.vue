<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Cash Position</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
          Real-time overview of all cash &amp; bank balances
          <span class="ml-2 text-xs text-gray-400 dark:text-gray-500">as of {{ today }}</span>
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="loadAll"
          class="inline-flex items-center gap-2 px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors shadow-sm">
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
          Refresh
        </button>
        <button @click="openAddCash"
          class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-semibold transition-colors shadow-sm">
          <Plus class="w-4 h-4" />
          Add Account
        </button>
      </div>
    </div>

    <!-- KPI Banner -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-gradient-to-br from-indigo-500 to-blue-600 rounded-2xl p-5 text-white shadow-lg">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-semibold uppercase tracking-wider text-indigo-200">Total Position</span>
          <div class="w-9 h-9 bg-white/20 rounded-xl flex items-center justify-center">
            <Landmark class="w-5 h-5" />
          </div>
        </div>
        <p class="text-3xl font-bold">{{ fmtCurrency(pos.total_position) }}</p>
        <p class="text-xs text-indigo-200 mt-1">Cash + Bank combined</p>
      </div>

      <div class="bg-white dark:bg-gray-900 rounded-2xl p-5 border border-gray-200 dark:border-gray-800 shadow-sm">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-semibold uppercase tracking-wider text-gray-400">Cash Accounts</span>
          <div class="w-9 h-9 bg-emerald-100 dark:bg-emerald-900/30 rounded-xl flex items-center justify-center">
            <Wallet class="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
          </div>
        </div>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(pos.total_cash) }}</p>
        <p class="text-xs text-gray-400 mt-1">{{ pos.cash_accounts?.length || 0 }} accounts</p>
      </div>

      <div class="bg-white dark:bg-gray-900 rounded-2xl p-5 border border-gray-200 dark:border-gray-800 shadow-sm">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-semibold uppercase tracking-wider text-gray-400">Bank Accounts</span>
          <div class="w-9 h-9 bg-blue-100 dark:bg-blue-900/30 rounded-xl flex items-center justify-center">
            <Building2 class="w-5 h-5 text-blue-600 dark:text-blue-400" />
          </div>
        </div>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(pos.total_bank) }}</p>
        <p class="text-xs text-gray-400 mt-1">{{ pos.bank_accounts?.length || 0 }} accounts</p>
      </div>

      <div class="bg-white dark:bg-gray-900 rounded-2xl p-5 border border-gray-200 dark:border-gray-800 shadow-sm">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-semibold uppercase tracking-wider text-gray-400">Pending Cheques</span>
          <div class="w-9 h-9 bg-amber-100 dark:bg-amber-900/30 rounded-xl flex items-center justify-center">
            <FileText class="w-5 h-5 text-amber-600 dark:text-amber-400" />
          </div>
        </div>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(pos.pending_cheques) }}</p>
        <p class="text-xs text-gray-400 mt-1">Awaiting deposit</p>
      </div>
    </div>

    <!-- Month Flow Row -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 bg-emerald-100 dark:bg-emerald-900/30 rounded-xl flex items-center justify-center flex-shrink-0">
          <TrendingUp class="w-6 h-6 text-emerald-600 dark:text-emerald-400" />
        </div>
        <div>
          <p class="text-xs text-gray-400 uppercase font-semibold tracking-wide">This Month Inflow</p>
          <p class="text-xl font-bold text-emerald-600 dark:text-emerald-400">+{{ fmtCurrency(pos.month_inflow) }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 bg-red-100 dark:bg-red-900/30 rounded-xl flex items-center justify-center flex-shrink-0">
          <TrendingDown class="w-6 h-6 text-red-500 dark:text-red-400" />
        </div>
        <div>
          <p class="text-xs text-gray-400 uppercase font-semibold tracking-wide">This Month Outflow</p>
          <p class="text-xl font-bold text-red-500 dark:text-red-400">-{{ fmtCurrency(pos.month_outflow) }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div :class="pos.net_month >= 0 ? 'bg-indigo-100 dark:bg-indigo-900/30' : 'bg-orange-100 dark:bg-orange-900/30'"
          class="w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0">
          <ArrowLeftRight :class="pos.net_month >= 0 ? 'text-indigo-600 dark:text-indigo-400' : 'text-orange-600 dark:text-orange-400'"
            class="w-6 h-6" />
        </div>
        <div>
          <p class="text-xs text-gray-400 uppercase font-semibold tracking-wide">Net Cash Flow</p>
          <p class="text-xl font-bold" :class="pos.net_month >= 0 ? 'text-indigo-600 dark:text-indigo-400' : 'text-orange-600 dark:text-orange-400'">
            {{ pos.net_month >= 0 ? '+' : '' }}{{ fmtCurrency(pos.net_month) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Two columns: Cash + Bank Accounts -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

      <!-- Cash Accounts -->
      <div class="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
        <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-800">
          <div class="flex items-center gap-2">
            <Wallet class="w-5 h-5 text-emerald-500" />
            <h2 class="font-semibold text-gray-900 dark:text-white">Cash Accounts</h2>
            <span class="ml-1 text-xs text-gray-400">({{ pos.cash_accounts?.length || 0 }})</span>
          </div>
          <button @click="openAddCash('cash')"
            class="inline-flex items-center gap-1 text-xs px-3 py-1.5 bg-emerald-50 dark:bg-emerald-900/20 text-emerald-700 dark:text-emerald-400 rounded-lg hover:bg-emerald-100 dark:hover:bg-emerald-900/40 transition-colors font-medium">
            <Plus class="w-3.5 h-3.5" /> Add
          </button>
        </div>
        <div v-if="loading" class="flex items-center justify-center py-10">
          <Loader2 class="w-6 h-6 text-indigo-500 animate-spin" />
        </div>
        <div v-else-if="!pos.cash_accounts?.length" class="flex flex-col items-center justify-center py-12 text-gray-400">
          <Wallet class="w-10 h-10 mb-2 opacity-30" />
          <p class="text-sm">No cash accounts configured</p>
        </div>
        <div v-else class="divide-y divide-gray-100 dark:divide-gray-800">
          <div v-for="acc in pos.cash_accounts" :key="acc.id"
            class="flex items-center justify-between px-5 py-4 hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors group">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 bg-emerald-100 dark:bg-emerald-900/30 rounded-xl flex items-center justify-center flex-shrink-0">
                <Wallet class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
              </div>
              <div>
                <p class="font-medium text-gray-900 dark:text-white text-sm">{{ acc.name }}</p>
                <p class="text-xs text-gray-400">{{ acc.currency || 'DZD' }}</p>
              </div>
            </div>
            <div class="text-right">
              <p class="font-bold text-gray-900 dark:text-white">{{ fmtCurrency(acc.balance) }}</p>
              <div class="mt-1 h-1.5 w-24 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
                <div class="h-full bg-emerald-500 rounded-full"
                  :style="{ width: totalForBar(acc.balance, pos.total_position) + '%' }" />
              </div>
            </div>
          </div>
        </div>
        <div v-if="pos.cash_accounts?.length" class="px-5 py-3 bg-emerald-50 dark:bg-emerald-900/10 border-t border-gray-100 dark:border-gray-800 flex justify-between items-center">
          <span class="text-xs text-gray-500 dark:text-gray-400 font-medium">Total Cash</span>
          <span class="font-bold text-emerald-700 dark:text-emerald-400">{{ fmtCurrency(pos.total_cash) }}</span>
        </div>
      </div>

      <!-- Bank Accounts -->
      <div class="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
        <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-800">
          <div class="flex items-center gap-2">
            <Building2 class="w-5 h-5 text-blue-500" />
            <h2 class="font-semibold text-gray-900 dark:text-white">Bank Accounts</h2>
            <span class="ml-1 text-xs text-gray-400">({{ pos.bank_accounts?.length || 0 }})</span>
          </div>
          <button @click="openAddCash('bank')"
            class="inline-flex items-center gap-1 text-xs px-3 py-1.5 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-400 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/40 transition-colors font-medium">
            <Plus class="w-3.5 h-3.5" /> Add
          </button>
        </div>
        <div v-if="loading" class="flex items-center justify-center py-10">
          <Loader2 class="w-6 h-6 text-indigo-500 animate-spin" />
        </div>
        <div v-else-if="!pos.bank_accounts?.length" class="flex flex-col items-center justify-center py-12 text-gray-400">
          <Building2 class="w-10 h-10 mb-2 opacity-30" />
          <p class="text-sm">No bank accounts configured</p>
        </div>
        <div v-else class="divide-y divide-gray-100 dark:divide-gray-800">
          <div v-for="acc in pos.bank_accounts" :key="acc.id"
            class="flex items-center justify-between px-5 py-4 hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 bg-blue-100 dark:bg-blue-900/30 rounded-xl flex items-center justify-center flex-shrink-0">
                <Building2 class="w-4 h-4 text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <p class="font-medium text-gray-900 dark:text-white text-sm">{{ acc.bank_name }}</p>
                <p class="text-xs text-gray-400">{{ acc.account_number }} &bull; {{ acc.currency || 'DZD' }}</p>
              </div>
            </div>
            <div class="text-right">
              <p class="font-bold text-gray-900 dark:text-white">{{ fmtCurrency(acc.balance) }}</p>
              <div class="mt-1 h-1.5 w-24 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
                <div class="h-full bg-blue-500 rounded-full"
                  :style="{ width: totalForBar(acc.balance, pos.total_position) + '%' }" />
              </div>
            </div>
          </div>
        </div>
        <div v-if="pos.bank_accounts?.length" class="px-5 py-3 bg-blue-50 dark:bg-blue-900/10 border-t border-gray-100 dark:border-gray-800 flex justify-between items-center">
          <span class="text-xs text-gray-500 dark:text-gray-400 font-medium">Total Bank</span>
          <span class="font-bold text-blue-700 dark:text-blue-400">{{ fmtCurrency(pos.total_bank) }}</span>
        </div>
      </div>
    </div>

    <!-- Recent Movements -->
    <div class="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
      <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-800">
        <div class="flex items-center gap-2">
          <Activity class="w-5 h-5 text-violet-500" />
          <h2 class="font-semibold text-gray-900 dark:text-white">Recent Movements</h2>
        </div>
        <router-link to="/treasury/bank-accounts"
          class="text-xs text-indigo-600 dark:text-indigo-400 hover:underline font-medium">View all</router-link>
      </div>
      <div v-if="!pos.recent_movements?.length" class="flex items-center justify-center py-10 text-gray-400 text-sm">
        No recent movements
      </div>
      <div v-else class="divide-y divide-gray-100 dark:divide-gray-800">
        <div v-for="(mov, idx) in pos.recent_movements" :key="idx"
          class="flex items-center gap-4 px-5 py-3 hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
          <div :class="mov.type === 'inflow' ? 'bg-emerald-100 dark:bg-emerald-900/30' : 'bg-red-100 dark:bg-red-900/30'"
            class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0">
            <ArrowUpRight v-if="mov.type === 'inflow'" class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
            <ArrowDownLeft v-else class="w-4 h-4 text-red-500 dark:text-red-400" />
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-gray-900 dark:text-white truncate">{{ mov.notes || mov.reference || mov.type }}</p>
            <p class="text-xs text-gray-400">{{ fmtDate(mov.date) }}</p>
          </div>
          <p class="font-semibold text-sm flex-shrink-0"
            :class="mov.type === 'inflow' ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-500 dark:text-red-400'">
            {{ mov.type === 'inflow' ? '+' : '-' }}{{ fmtCurrency(mov.amount) }}
          </p>
        </div>
      </div>
    </div>

    <!-- Add Account Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showModal = false" />
        <div class="relative bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-md border border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-800">
            <h3 class="font-bold text-gray-900 dark:text-white text-lg">
              {{ modalType === 'cash' ? 'New Cash Account' : 'New Bank Account' }}
            </h3>
            <button @click="showModal = false" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
              <X class="w-5 h-5 text-gray-400" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <template v-if="modalType === 'cash'">
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Account Name *</label>
                <input v-model="form.name" type="text" placeholder="e.g. Petty Cash - Head Office"
                  class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none" />
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Type</label>
                  <select v-model="form.account_type"
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none">
                    <option value="petty_cash">Petty Cash</option>
                    <option value="safe">Safe</option>
                    <option value="cashier">Cashier</option>
                  </select>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Currency</label>
                  <select v-model="form.currency"
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none">
                    <option>DZD</option><option>EUR</option><option>USD</option>
                  </select>
                </div>
              </div>
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Opening Balance</label>
                <input v-model.number="form.opening_balance" type="number" min="0" step="0.01"
                  class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none" />
              </div>
            </template>

            <template v-else>
              <div class="grid grid-cols-2 gap-4">
                <div class="col-span-2">
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Bank Name *</label>
                  <input v-model="form.bank_name" type="text" placeholder="e.g. BNA, CPA, BEA..."
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Account Number *</label>
                  <input v-model="form.account_number" type="text"
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">RIB</label>
                  <input v-model="form.rib" type="text"
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Currency</label>
                  <select v-model="form.currency"
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none">
                    <option>DZD</option><option>EUR</option><option>USD</option>
                  </select>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Opening Balance</label>
                  <input v-model.number="form.opening_balance" type="number" min="0" step="0.01"
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none" />
                </div>
              </div>
            </template>
          </div>
          <div class="flex gap-3 px-6 py-4 border-t border-gray-100 dark:border-gray-800">
            <button @click="showModal = false"
              class="flex-1 px-4 py-2 rounded-lg border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
              Cancel
            </button>
            <button @click="saveAccount" :disabled="saving"
              class="flex-1 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold transition-colors disabled:opacity-50">
              {{ saving ? 'Saving...' : 'Save Account' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  Landmark, Wallet, Building2, FileText, TrendingUp, TrendingDown,
  ArrowLeftRight, ArrowUpRight, ArrowDownLeft, Activity,
  RefreshCw, Plus, X, Loader2
} from 'lucide-vue-next'
import { treasuryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()
const loading = ref(true)
const saving = ref(false)
const showModal = ref(false)
const modalType = ref<'cash' | 'bank'>('cash')

const today = new Date().toLocaleDateString('en-GB', { day: '2-digit', month: 'long', year: 'numeric' })

interface CashPos {
  cash_accounts: any[]
  bank_accounts: any[]
  total_cash: number
  total_bank: number
  total_position: number
  pending_cheques: number
  month_inflow: number
  month_outflow: number
  net_month: number
  recent_movements: any[]
}

const pos = ref<CashPos>({
  cash_accounts: [], bank_accounts: [],
  total_cash: 0, total_bank: 0, total_position: 0,
  pending_cheques: 0, month_inflow: 0, month_outflow: 0, net_month: 0,
  recent_movements: []
})

const form = ref<any>({
  name: '', account_type: 'petty_cash', currency: 'DZD', opening_balance: 0,
  bank_name: '', account_number: '', rib: ''
})

async function loadAll() {
  loading.value = true
  try {
    const res = await treasuryAPI.getCashPosition()
    pos.value = res.data
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load cash position', 'error')
  } finally {
    loading.value = false
  }
}

function openAddCash(type: 'cash' | 'bank' = 'cash') {
  modalType.value = type
  form.value = { name: '', account_type: 'petty_cash', currency: 'DZD', opening_balance: 0, bank_name: '', account_number: '', rib: '' }
  showModal.value = true
}

async function saveAccount() {
  saving.value = true
  try {
    if (modalType.value === 'cash') {
      if (!form.value.name) { store.addToast('Account name required', 'error'); return }
      await treasuryAPI.createCashAccount(form.value)
    } else {
      if (!form.value.bank_name || !form.value.account_number) { store.addToast('Bank name and account number required', 'error'); return }
      await treasuryAPI.createBankAccount(form.value)
    }
    store.addToast('Account created', 'success')
    showModal.value = false
    await loadAll()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to create account', 'error')
  } finally {
    saving.value = false
  }
}

function fmtCurrency(n?: number) {
  if (!n) return '0 DZD'
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(n) + ' DZD'
}

function fmtDate(d?: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function totalForBar(val: number, total: number): number {
  if (!total || !val) return 0
  return Math.min(Math.round((val / total) * 100), 100)
}

onMounted(loadAll)
</script>
