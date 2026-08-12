<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Receipts</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Record and manage all incoming cash receipts</p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="loadReceipts"
          class="inline-flex items-center gap-2 px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors shadow-sm">
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
          Refresh
        </button>
        <button @click="openModal()"
          class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-semibold transition-colors shadow-sm">
          <Plus class="w-4 h-4" />
          New Receipt
        </button>
      </div>
    </div>

    <!-- KPIs -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-gradient-to-br from-emerald-500 to-teal-600 rounded-xl p-5 text-white shadow-sm">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-8 h-8 bg-white/20 rounded-lg flex items-center justify-center">
            <ArrowDownToLine class="w-4 h-4" />
          </div>
          <span class="text-xs uppercase font-semibold text-emerald-100 tracking-wide">Total Receipts</span>
        </div>
        <p class="text-2xl font-bold">{{ fmtCurrency(totalAmount) }}</p>
        <p class="text-xs text-emerald-100 mt-1">{{ receipts.length }} records</p>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
        <div class="flex items-center gap-3 mb-2">
          <div class="w-9 h-9 bg-blue-100 dark:bg-blue-900/30 rounded-xl flex items-center justify-center">
            <CheckCircle2 class="w-5 h-5 text-blue-600 dark:text-blue-400" />
          </div>
          <span class="text-xs uppercase font-semibold text-gray-400 tracking-wide">Confirmed</span>
        </div>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(confirmedTotal) }}</p>
        <p class="text-xs text-gray-400 mt-1">{{ receipts.filter(r => r.status === 'confirmed').length }} receipts</p>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
        <div class="flex items-center gap-3 mb-2">
          <div class="w-9 h-9 bg-amber-100 dark:bg-amber-900/30 rounded-xl flex items-center justify-center">
            <Clock class="w-5 h-5 text-amber-600 dark:text-amber-400" />
          </div>
          <span class="text-xs uppercase font-semibold text-gray-400 tracking-wide">Draft</span>
        </div>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(draftTotal) }}</p>
        <p class="text-xs text-gray-400 mt-1">{{ receipts.filter(r => r.status === 'draft').length }} pending</p>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
        <div class="flex items-center gap-3 mb-2">
          <div class="w-9 h-9 bg-violet-100 dark:bg-violet-900/30 rounded-xl flex items-center justify-center">
            <AlertCircle class="w-5 h-5 text-violet-600 dark:text-violet-400" />
          </div>
          <span class="text-xs uppercase font-semibold text-gray-400 tracking-wide">Unallocated</span>
        </div>
        <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(unallocatedTotal) }}</p>
        <p class="text-xs text-gray-400 mt-1">Needs matching</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
      <div class="flex flex-wrap items-center gap-3">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input v-model="search" type="text" placeholder="Search by number, partner, reference..."
            class="w-full pl-9 pr-4 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm text-gray-900 dark:text-white outline-none focus:ring-2 focus:ring-indigo-500" />
        </div>
        <select v-model="filterStatus"
          class="px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm text-gray-700 dark:text-gray-200 outline-none focus:ring-2 focus:ring-indigo-500">
          <option value="">All Statuses</option>
          <option value="draft">Draft</option>
          <option value="confirmed">Confirmed</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <select v-model="filterMethod"
          class="px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm text-gray-700 dark:text-gray-200 outline-none focus:ring-2 focus:ring-indigo-500">
          <option value="">All Methods</option>
          <option value="bank_transfer">Bank Transfer</option>
          <option value="cheque">Cheque</option>
          <option value="cash">Cash</option>
          <option value="card">Card</option>
        </select>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <Loader2 class="w-8 h-8 text-indigo-500 animate-spin" />
      </div>
      <div v-else-if="!filtered.length" class="flex flex-col items-center justify-center py-20 text-gray-400">
        <FileX class="w-14 h-14 mb-3 opacity-30" />
        <p class="font-medium">No receipts found</p>
        <p class="text-sm mt-1">Create your first receipt to get started</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Number</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Date</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Partner</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide hidden md:table-cell">Method</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide hidden md:table-cell">Bank</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Amount</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide hidden sm:table-cell">Unallocated</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Status</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr v-for="rec in filtered" :key="rec.id"
              class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
              <td class="px-4 py-3">
                <span class="font-mono text-sm font-semibold text-indigo-600 dark:text-indigo-400">{{ rec.number }}</span>
              </td>
              <td class="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">{{ fmtDate(rec.receipt_date) }}</td>
              <td class="px-4 py-3 font-medium text-gray-900 dark:text-white">{{ rec.partner_name || '—' }}</td>
              <td class="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs capitalize hidden md:table-cell">
                {{ (rec.payment_method || '—').replace('_', ' ') }}
              </td>
              <td class="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs hidden md:table-cell">{{ rec.bank_name || '—' }}</td>
              <td class="px-4 py-3 text-right font-semibold text-gray-900 dark:text-white">{{ fmtCurrency(rec.amount) }}</td>
              <td class="px-4 py-3 text-right text-xs hidden sm:table-cell"
                :class="rec.unallocated_amount > 0 ? 'text-amber-600 dark:text-amber-400 font-semibold' : 'text-gray-400'">
                {{ rec.unallocated_amount > 0 ? fmtCurrency(rec.unallocated_amount) : '—' }}
              </td>
              <td class="px-4 py-3">
                <span :class="statusClass(rec.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold capitalize">
                  {{ rec.status }}
                </span>
              </td>
              <td class="px-4 py-3 text-right">
                <div class="flex items-center justify-end gap-1">
                  <button v-if="rec.status === 'draft'" @click="confirmReceipt(rec)"
                    class="p-1.5 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-100 transition-colors"
                    title="Confirm">
                    <CheckCircle2 class="w-3.5 h-3.5" />
                  </button>
                  <button v-if="rec.status !== 'cancelled'" @click="editReceipt(rec)"
                    class="p-1.5 rounded-lg bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
                    title="Edit">
                    <Pencil class="w-3.5 h-3.5" />
                  </button>
                  <button v-if="rec.status === 'draft'" @click="deleteReceipt(rec)"
                    class="p-1.5 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 hover:bg-red-100 transition-colors"
                    title="Delete">
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="border-t-2 border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/40">
              <td colspan="5" class="px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">
                Totals ({{ filtered.length }})
              </td>
              <td class="px-4 py-3 text-right font-bold text-gray-900 dark:text-white">
                {{ fmtCurrency(filtered.reduce((s, r) => s + r.amount, 0)) }}
              </td>
              <td class="px-4 py-3 text-right font-semibold text-amber-600 dark:text-amber-400 hidden sm:table-cell">
                {{ fmtCurrency(filtered.reduce((s, r) => s + (r.unallocated_amount || 0), 0)) }}
              </td>
              <td colspan="2"></td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showModal = false" />
        <div class="relative bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-lg border border-gray-200 dark:border-gray-700 max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-800 sticky top-0 bg-white dark:bg-gray-900 z-10">
            <h3 class="font-bold text-gray-900 dark:text-white text-lg">{{ editMode ? 'Edit Receipt' : 'New Receipt' }}</h3>
            <button @click="showModal = false" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
              <X class="w-5 h-5 text-gray-400" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="label-sm">Receipt Type</label>
                <select v-model="form.receipt_type" class="input-field">
                  <option value="customer">Customer</option>
                  <option value="other">Other</option>
                </select>
              </div>
              <div>
                <label class="label-sm">Receipt Date *</label>
                <input v-model="form.receipt_date" type="date" class="input-field" />
              </div>
              <div class="col-span-2">
                <label class="label-sm">Partner Name</label>
                <input v-model="form.partner_name" type="text" placeholder="Customer name" class="input-field" />
              </div>
              <div>
                <label class="label-sm">Amount (DZD) *</label>
                <input v-model.number="form.amount" type="number" min="0" step="0.01" class="input-field" />
              </div>
              <div>
                <label class="label-sm">Currency</label>
                <select v-model="form.currency" class="input-field">
                  <option>DZD</option><option>EUR</option><option>USD</option>
                </select>
              </div>
              <div>
                <label class="label-sm">Payment Method</label>
                <select v-model="form.payment_method" class="input-field">
                  <option value="bank_transfer">Bank Transfer</option>
                  <option value="cheque">Cheque</option>
                  <option value="cash">Cash</option>
                  <option value="card">Card</option>
                </select>
              </div>
              <div>
                <label class="label-sm">Bank Account</label>
                <select v-model="form.bank_account_id" class="input-field">
                  <option value="">— None —</option>
                  <option v-for="ba in bankAccounts" :key="ba.id" :value="ba.id">
                    {{ ba.bank_name }} - {{ ba.account_number }}
                  </option>
                </select>
              </div>
              <div class="col-span-2">
                <label class="label-sm">Reference</label>
                <input v-model="form.reference" type="text" placeholder="Invoice number, cheque number..." class="input-field" />
              </div>
              <div>
                <label class="label-sm">Invoice Number</label>
                <input v-model="form.invoice_number" type="text" class="input-field" />
              </div>
            </div>
            <div>
              <label class="label-sm">Description</label>
              <textarea v-model="form.description" rows="2" class="input-field resize-none" />
            </div>
          </div>
          <div class="flex gap-3 px-6 py-4 border-t border-gray-100 dark:border-gray-800">
            <button @click="showModal = false"
              class="flex-1 px-4 py-2 rounded-lg border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
              Cancel
            </button>
            <button @click="saveReceipt" :disabled="saving"
              class="flex-1 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold transition-colors disabled:opacity-50">
              {{ saving ? 'Saving...' : (editMode ? 'Update' : 'Create Receipt') }}
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
  Plus, X, Loader2, FileX, Search, RefreshCw,
  ArrowDownToLine, CheckCircle2, Clock, AlertCircle,
  Pencil, Trash2
} from 'lucide-vue-next'
import { treasuryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()
const loading = ref(true)
const saving = ref(false)
const showModal = ref(false)
const editMode = ref(false)
const editId = ref('')

const receipts = ref<any[]>([])
const bankAccounts = ref<any[]>([])
const search = ref('')
const filterStatus = ref('')
const filterMethod = ref('')

const filtered = computed(() => {
  let list = receipts.value
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(r => r.number?.toLowerCase().includes(q) || r.partner_name?.toLowerCase().includes(q) || r.reference?.toLowerCase().includes(q))
  }
  if (filterStatus.value) list = list.filter(r => r.status === filterStatus.value)
  if (filterMethod.value) list = list.filter(r => r.payment_method === filterMethod.value)
  return list
})

const totalAmount = computed(() => receipts.value.reduce((s, r) => s + (r.amount || 0), 0))
const confirmedTotal = computed(() => receipts.value.filter(r => r.status === 'confirmed').reduce((s, r) => s + (r.amount || 0), 0))
const draftTotal = computed(() => receipts.value.filter(r => r.status === 'draft').reduce((s, r) => s + (r.amount || 0), 0))
const unallocatedTotal = computed(() => receipts.value.reduce((s, r) => s + (r.unallocated_amount || 0), 0))

const form = ref<any>({
  receipt_type: 'customer', receipt_date: '', partner_name: '', amount: 0,
  currency: 'DZD', payment_method: 'bank_transfer', bank_account_id: '',
  reference: '', invoice_number: '', description: ''
})

async function loadReceipts() {
  loading.value = true
  try {
    const [recRes, baRes] = await Promise.all([
      treasuryAPI.getReceipts(),
      treasuryAPI.getBankAccounts()
    ])
    receipts.value = recRes.data || []
    bankAccounts.value = baRes.data || []
  } catch {
    store.addToast('Failed to load receipts', 'error')
  } finally {
    loading.value = false
  }
}

function openModal() {
  editMode.value = false
  editId.value = ''
  form.value = {
    receipt_type: 'customer', receipt_date: new Date().toISOString().split('T')[0],
    partner_name: '', amount: 0, currency: 'DZD',
    payment_method: 'bank_transfer', bank_account_id: '',
    reference: '', invoice_number: '', description: ''
  }
  showModal.value = true
}

function editReceipt(rec: any) {
  editMode.value = true
  editId.value = rec.id
  form.value = {
    receipt_type: rec.receipt_type || 'customer',
    receipt_date: rec.receipt_date ? rec.receipt_date.split('T')[0] : '',
    partner_name: rec.partner_name || '',
    amount: rec.amount, currency: rec.currency || 'DZD',
    payment_method: rec.payment_method || 'bank_transfer',
    bank_account_id: rec.bank_account_id || '',
    reference: rec.reference || '', invoice_number: rec.invoice_number || '',
    description: rec.description || ''
  }
  showModal.value = true
}

async function saveReceipt() {
  if (!form.value.amount) { store.addToast('Amount is required', 'error'); return }
  saving.value = true
  try {
    if (editMode.value) {
      await treasuryAPI.updateReceipt(editId.value, form.value)
      store.addToast('Receipt updated', 'success')
    } else {
      await treasuryAPI.createReceipt(form.value)
      store.addToast('Receipt created', 'success')
    }
    showModal.value = false
    await loadReceipts()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Operation failed', 'error')
  } finally {
    saving.value = false
  }
}

async function confirmReceipt(rec: any) {
  try {
    await treasuryAPI.confirmReceipt(rec.id)
    store.addToast('Receipt confirmed', 'success')
    await loadReceipts()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to confirm', 'error')
  }
}

async function deleteReceipt(rec: any) {
  if (!confirm(`Delete receipt ${rec.number}?`)) return
  try {
    await treasuryAPI.deleteReceipt(rec.id)
    store.addToast('Receipt deleted', 'success')
    await loadReceipts()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to delete', 'error')
  }
}

function statusClass(s: string) {
  switch (s) {
    case 'confirmed': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'draft': return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    case 'cancelled': return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
    default: return 'bg-gray-100 text-gray-600'
  }
}

function fmtDate(d?: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function fmtCurrency(n?: number) {
  if (!n) return '0 DZD'
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(n) + ' DZD'
}

onMounted(loadReceipts)
</script>

<style scoped>
.label-sm { @apply block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5; }
.input-field { @apply w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition-colors; }
</style>
