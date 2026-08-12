<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Bank Accounts</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Manage all linked bank accounts and cash accounts</p>
      </div>
      <div class="flex items-center gap-2">
        <div class="flex items-center bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-1">
          <button v-for="t in tabs" :key="t.id" @click="activeTab = t.id"
            :class="activeTab === t.id ? 'bg-indigo-600 text-white shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'"
            class="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all">
            <component :is="t.icon" class="w-3.5 h-3.5" />
            {{ t.label }}
          </button>
        </div>
        <button @click="openModal()"
          class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-semibold transition-colors shadow-sm">
          <Plus class="w-4 h-4" />
          New Account
        </button>
      </div>
    </div>

    <!-- KPI row -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-9 h-9 bg-blue-100 dark:bg-blue-900/30 rounded-xl flex items-center justify-center">
            <Building2 class="w-5 h-5 text-blue-600 dark:text-blue-400" />
          </div>
          <span class="text-xs uppercase font-semibold text-gray-400 tracking-wide">Bank Total</span>
        </div>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(bankTotal) }}</p>
        <p class="text-xs text-gray-400 mt-1">{{ bankAccounts.length }} active accounts</p>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-9 h-9 bg-emerald-100 dark:bg-emerald-900/30 rounded-xl flex items-center justify-center">
            <Wallet class="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
          </div>
          <span class="text-xs uppercase font-semibold text-gray-400 tracking-wide">Cash Total</span>
        </div>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(cashTotal) }}</p>
        <p class="text-xs text-gray-400 mt-1">{{ cashAccounts.length }} cash accounts</p>
      </div>
      <div class="bg-gradient-to-br from-indigo-500 to-blue-600 rounded-xl p-5 text-white">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-9 h-9 bg-white/20 rounded-xl flex items-center justify-center">
            <Landmark class="w-5 h-5" />
          </div>
          <span class="text-xs uppercase font-semibold text-indigo-200 tracking-wide">Grand Total</span>
        </div>
        <p class="text-2xl font-bold">{{ fmtCurrency(bankTotal + cashTotal) }}</p>
        <p class="text-xs text-indigo-200 mt-1">All liquid assets combined</p>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <Loader2 class="w-8 h-8 text-indigo-500 animate-spin" />
    </div>

    <template v-else>
      <!-- Bank Accounts tab -->
      <div v-if="activeTab === 'bank'">
        <div v-if="!bankAccounts.length" class="flex flex-col items-center justify-center py-20 text-gray-400">
          <Building2 class="w-14 h-14 mb-3 opacity-30" />
          <p class="font-medium">No bank accounts</p>
          <p class="text-sm mt-1">Add your first bank account to get started</p>
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          <div v-for="acc in bankAccounts" :key="acc.id"
            class="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-5 hover:shadow-md transition-shadow group">
            <!-- Card header -->
            <div class="flex items-start justify-between mb-4">
              <div class="flex items-center gap-3">
                <div class="w-12 h-12 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-xl flex items-center justify-center text-white font-bold text-lg flex-shrink-0 shadow">
                  {{ acc.bank_name?.charAt(0) || 'B' }}
                </div>
                <div>
                  <p class="font-semibold text-gray-900 dark:text-white">{{ acc.bank_name }}</p>
                  <p class="text-xs text-gray-400">{{ acc.branch || 'Main Branch' }}</p>
                </div>
              </div>
              <span :class="acc.is_active ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'"
                class="text-xs font-semibold px-2.5 py-1 rounded-full">
                {{ acc.is_active ? 'Active' : 'Inactive' }}
              </span>
            </div>
            <!-- Account info -->
            <div class="space-y-2 mb-4">
              <div class="flex justify-between py-1.5 border-b border-gray-100 dark:border-gray-800">
                <span class="text-xs text-gray-400 font-medium">Account No.</span>
                <span class="text-xs font-mono text-gray-700 dark:text-gray-200">{{ acc.account_number }}</span>
              </div>
              <div v-if="acc.rib" class="flex justify-between py-1.5 border-b border-gray-100 dark:border-gray-800">
                <span class="text-xs text-gray-400 font-medium">RIB</span>
                <span class="text-xs font-mono text-gray-700 dark:text-gray-200">{{ acc.rib }}</span>
              </div>
              <div v-if="acc.swift_code" class="flex justify-between py-1.5 border-b border-gray-100 dark:border-gray-800">
                <span class="text-xs text-gray-400 font-medium">SWIFT</span>
                <span class="text-xs font-mono text-gray-700 dark:text-gray-200">{{ acc.swift_code }}</span>
              </div>
              <div class="flex justify-between py-1.5">
                <span class="text-xs text-gray-400 font-medium">Currency</span>
                <span class="text-xs font-semibold text-gray-700 dark:text-gray-200">{{ acc.currency }}</span>
              </div>
            </div>
            <!-- Balance footer -->
            <div class="bg-blue-50 dark:bg-blue-900/10 rounded-xl p-3 flex items-center justify-between">
              <span class="text-xs text-blue-600 dark:text-blue-400 font-medium">Current Balance</span>
              <span class="text-lg font-bold text-blue-700 dark:text-blue-400">{{ fmtCurrency(acc.balance) }}</span>
            </div>
            <!-- Actions -->
            <div class="flex gap-2 mt-3 opacity-0 group-hover:opacity-100 transition-opacity">
              <button @click="editAccount(acc, 'bank')"
                class="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs text-gray-600 dark:text-gray-300 bg-gray-50 dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors font-medium">
                <Pencil class="w-3.5 h-3.5" /> Edit
              </button>
              <button @click="viewMovements(acc.id)"
                class="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/20 hover:bg-indigo-100 dark:hover:bg-indigo-900/40 rounded-lg transition-colors font-medium">
                <Activity class="w-3.5 h-3.5" /> Movements
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Cash Accounts tab -->
      <div v-if="activeTab === 'cash'">
        <div v-if="!cashAccounts.length" class="flex flex-col items-center justify-center py-20 text-gray-400">
          <Wallet class="w-14 h-14 mb-3 opacity-30" />
          <p class="font-medium">No cash accounts</p>
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          <div v-for="acc in cashAccounts" :key="acc.id"
            class="bg-white dark:bg-gray-900 rounded-2xl border border-gray-200 dark:border-gray-800 p-5 hover:shadow-md transition-shadow group">
            <div class="flex items-start justify-between mb-4">
              <div class="flex items-center gap-3">
                <div class="w-12 h-12 bg-gradient-to-br from-emerald-500 to-teal-600 rounded-xl flex items-center justify-center text-white font-bold text-lg shadow">
                  {{ acc.name?.charAt(0) || 'C' }}
                </div>
                <div>
                  <p class="font-semibold text-gray-900 dark:text-white">{{ acc.name }}</p>
                  <p class="text-xs text-gray-400 capitalize">{{ (acc.account_type || 'petty_cash').replace('_', ' ') }}</p>
                </div>
              </div>
              <span :class="acc.is_active ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-gray-100 text-gray-600'"
                class="text-xs font-semibold px-2.5 py-1 rounded-full">
                {{ acc.is_active ? 'Active' : 'Inactive' }}
              </span>
            </div>
            <div class="space-y-2 mb-4">
              <div class="flex justify-between py-1.5 border-b border-gray-100 dark:border-gray-800">
                <span class="text-xs text-gray-400 font-medium">Currency</span>
                <span class="text-xs font-semibold text-gray-700 dark:text-gray-200">{{ acc.currency }}</span>
              </div>
              <div v-if="acc.account_number" class="flex justify-between py-1.5">
                <span class="text-xs text-gray-400 font-medium">Account No.</span>
                <span class="text-xs font-mono text-gray-700 dark:text-gray-200">{{ acc.account_number }}</span>
              </div>
            </div>
            <div class="bg-emerald-50 dark:bg-emerald-900/10 rounded-xl p-3 flex items-center justify-between">
              <span class="text-xs text-emerald-600 dark:text-emerald-400 font-medium">Balance</span>
              <span class="text-lg font-bold text-emerald-700 dark:text-emerald-400">{{ fmtCurrency(acc.balance) }}</span>
            </div>
            <div class="flex gap-2 mt-3 opacity-0 group-hover:opacity-100 transition-opacity">
              <button @click="editAccount(acc, 'cash')"
                class="flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs text-gray-600 dark:text-gray-300 bg-gray-50 dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors font-medium">
                <Pencil class="w-3.5 h-3.5" /> Edit
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeModal" />
        <div class="relative bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-lg border border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-800">
            <h3 class="font-bold text-gray-900 dark:text-white text-lg">
              {{ editMode ? 'Edit Account' : (activeTab === 'bank' ? 'New Bank Account' : 'New Cash Account') }}
            </h3>
            <button @click="closeModal" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
              <X class="w-5 h-5 text-gray-400" />
            </button>
          </div>

          <!-- Account type selector (new only) -->
          <div v-if="!editMode" class="px-6 pt-4 flex gap-3">
            <button v-for="t in ['bank', 'cash']" :key="t" @click="modalAccountType = t"
              :class="modalAccountType === t ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20 text-indigo-700 dark:text-indigo-400' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400'"
              class="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 border-2 rounded-xl text-sm font-semibold transition-all capitalize">
              <Building2 v-if="t === 'bank'" class="w-4 h-4" />
              <Wallet v-else class="w-4 h-4" />
              {{ t }} Account
            </button>
          </div>

          <div class="p-6 space-y-4">
            <!-- Bank fields -->
            <template v-if="modalAccountType === 'bank'">
              <div>
                <label class="label-sm">Bank Name *</label>
                <input v-model="form.bank_name" type="text" placeholder="BNA, CPA, BEA, BADR..."
                  class="input-field" />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="label-sm">Account Number *</label>
                  <input v-model="form.account_number" type="text" class="input-field" />
                </div>
                <div>
                  <label class="label-sm">RIB</label>
                  <input v-model="form.rib" type="text" class="input-field" />
                </div>
                <div>
                  <label class="label-sm">SWIFT Code</label>
                  <input v-model="form.swift_code" type="text" class="input-field" />
                </div>
                <div>
                  <label class="label-sm">Branch</label>
                  <input v-model="form.branch" type="text" class="input-field" />
                </div>
                <div>
                  <label class="label-sm">Currency</label>
                  <select v-model="form.currency" class="input-field">
                    <option>DZD</option><option>EUR</option><option>USD</option>
                  </select>
                </div>
                <div>
                  <label class="label-sm">Opening Balance</label>
                  <input v-model.number="form.opening_balance" type="number" min="0" step="0.01" class="input-field" />
                </div>
              </div>
              <div>
                <label class="label-sm">Notes</label>
                <textarea v-model="form.notes" rows="2" class="input-field resize-none" />
              </div>
            </template>
            <!-- Cash fields -->
            <template v-else>
              <div>
                <label class="label-sm">Account Name *</label>
                <input v-model="form.name" type="text" placeholder="Petty Cash - Head Office" class="input-field" />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="label-sm">Type</label>
                  <select v-model="form.account_type" class="input-field">
                    <option value="petty_cash">Petty Cash</option>
                    <option value="safe">Safe</option>
                    <option value="cashier">Cashier</option>
                  </select>
                </div>
                <div>
                  <label class="label-sm">Currency</label>
                  <select v-model="form.currency" class="input-field">
                    <option>DZD</option><option>EUR</option><option>USD</option>
                  </select>
                </div>
                <div class="col-span-2">
                  <label class="label-sm">Opening Balance</label>
                  <input v-model.number="form.opening_balance" type="number" min="0" step="0.01" class="input-field" />
                </div>
              </div>
              <div>
                <label class="label-sm">Notes</label>
                <textarea v-model="form.notes" rows="2" class="input-field resize-none" />
              </div>
            </template>
          </div>

          <div class="flex gap-3 px-6 py-4 border-t border-gray-100 dark:border-gray-800">
            <button @click="closeModal"
              class="flex-1 px-4 py-2 rounded-lg border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
              Cancel
            </button>
            <button @click="saveAccount" :disabled="saving"
              class="flex-1 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold transition-colors disabled:opacity-50">
              {{ saving ? 'Saving...' : (editMode ? 'Update' : 'Create Account') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Building2, Wallet, Landmark, Plus, X, Loader2,
  Pencil, Activity, RefreshCw
} from 'lucide-vue-next'
import { treasuryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import { useRouter } from 'vue-router'

const store = useAppStore()
const router = useRouter()
const loading = ref(true)
const saving = ref(false)
const showModal = ref(false)
const editMode = ref(false)
const activeTab = ref<'bank' | 'cash'>('bank')
const modalAccountType = ref<'bank' | 'cash'>('bank')
const editId = ref('')

const tabs = [
  { id: 'bank', label: 'Bank', icon: Building2 },
  { id: 'cash', label: 'Cash', icon: Wallet },
]

const bankAccounts = ref<any[]>([])
const cashAccounts = ref<any[]>([])

const bankTotal = computed(() => bankAccounts.value.reduce((s, a) => s + (a.balance || 0), 0))
const cashTotal = computed(() => cashAccounts.value.reduce((s, a) => s + (a.balance || 0), 0))

const form = ref<any>({
  bank_name: '', account_number: '', rib: '', swift_code: '', branch: '',
  currency: 'DZD', opening_balance: 0, notes: '',
  name: '', account_type: 'petty_cash'
})

async function loadAll() {
  loading.value = true
  try {
    const [bankRes, cashRes] = await Promise.all([
      treasuryAPI.getBankAccounts(),
      treasuryAPI.getCashAccounts()
    ])
    bankAccounts.value = bankRes.data || []
    cashAccounts.value = cashRes.data || []
  } catch (e: any) {
    store.addToast('Failed to load accounts', 'error')
  } finally {
    loading.value = false
  }
}

function resetForm() {
  form.value = {
    bank_name: '', account_number: '', rib: '', swift_code: '', branch: '',
    currency: 'DZD', opening_balance: 0, notes: '',
    name: '', account_type: 'petty_cash'
  }
}

function openModal() {
  editMode.value = false
  editId.value = ''
  modalAccountType.value = activeTab.value
  resetForm()
  showModal.value = true
}

function editAccount(acc: any, type: 'bank' | 'cash') {
  editMode.value = true
  editId.value = acc.id
  modalAccountType.value = type
  if (type === 'bank') {
    form.value = {
      bank_name: acc.bank_name, account_number: acc.account_number,
      rib: acc.rib || '', swift_code: acc.swift_code || '',
      branch: acc.branch || '', currency: acc.currency || 'DZD',
      notes: acc.notes || ''
    }
  } else {
    form.value = {
      name: acc.name, account_type: acc.account_type || 'petty_cash',
      currency: acc.currency || 'DZD', notes: acc.notes || ''
    }
  }
  showModal.value = true
}

function closeModal() { showModal.value = false }

async function saveAccount() {
  saving.value = true
  try {
    if (editMode.value) {
      if (modalAccountType.value === 'bank') await treasuryAPI.updateBankAccount(editId.value, form.value)
      else await treasuryAPI.updateCashAccount(editId.value, form.value)
      store.addToast('Account updated', 'success')
    } else {
      if (modalAccountType.value === 'bank') {
        if (!form.value.bank_name || !form.value.account_number) { store.addToast('Bank name and account number required', 'error'); return }
        await treasuryAPI.createBankAccount(form.value)
      } else {
        if (!form.value.name) { store.addToast('Account name required', 'error'); return }
        await treasuryAPI.createCashAccount(form.value)
      }
      store.addToast('Account created', 'success')
    }
    closeModal()
    await loadAll()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Operation failed', 'error')
  } finally {
    saving.value = false
  }
}

function viewMovements(id: string) {
  router.push(`/treasury/cash-position`)
}

function fmtCurrency(n?: number) {
  if (!n) return '0 DZD'
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(n) + ' DZD'
}

onMounted(loadAll)
</script>

<style scoped>
.label-sm { @apply block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5; }
.input-field { @apply w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition-colors; }
</style>
