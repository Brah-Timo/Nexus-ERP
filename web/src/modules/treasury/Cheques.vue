<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Cheques</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Track issued and received cheques</p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="loadCheques"
          class="inline-flex items-center gap-2 px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors shadow-sm">
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
          Refresh
        </button>
        <button @click="openModal()"
          class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-semibold transition-colors shadow-sm">
          <Plus class="w-4 h-4" />
          New Cheque
        </button>
      </div>
    </div>

    <!-- KPIs -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="kpi in kpis" :key="kpi.label"
        class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center gap-3 mb-2">
          <div :class="kpi.iconBg" class="w-8 h-8 rounded-lg flex items-center justify-center">
            <component :is="kpi.icon" :class="kpi.iconColor" class="w-4 h-4" />
          </div>
          <span class="text-xs uppercase font-semibold text-gray-400 tracking-wide">{{ kpi.label }}</span>
        </div>
        <p class="text-xl font-bold text-gray-900 dark:text-white">{{ kpi.value }}</p>
        <p class="text-xs text-gray-400 mt-0.5">{{ kpi.sub }}</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
      <div class="flex flex-wrap items-center gap-3">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input v-model="search" type="text" placeholder="Search by number, partner..."
            class="w-full pl-9 pr-4 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm text-gray-900 dark:text-white outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
        </div>
        <select v-model="filterType"
          class="px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm text-gray-700 dark:text-gray-200 outline-none focus:ring-2 focus:ring-indigo-500">
          <option value="">All Types</option>
          <option value="received">Received</option>
          <option value="issued">Issued</option>
        </select>
        <select v-model="filterStatus"
          class="px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm text-gray-700 dark:text-gray-200 outline-none focus:ring-2 focus:ring-indigo-500">
          <option value="">All Statuses</option>
          <option value="pending">Pending</option>
          <option value="deposited">Deposited</option>
          <option value="bounced">Bounced</option>
          <option value="cancelled">Cancelled</option>
        </select>
        <span v-if="filtered.length !== cheques.length" class="text-xs text-gray-400">
          {{ filtered.length }} of {{ cheques.length }}
        </span>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <Loader2 class="w-8 h-8 text-indigo-500 animate-spin" />
      </div>
      <div v-else-if="!filtered.length" class="flex flex-col items-center justify-center py-20 text-gray-400">
        <FileX class="w-14 h-14 mb-3 opacity-30" />
        <p class="font-medium">No cheques found</p>
        <p class="text-sm mt-1">Add your first cheque to get started</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Number</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Type</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Partner</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide hidden md:table-cell">Bank</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide hidden sm:table-cell">Due Date</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Amount</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Status</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr v-for="ch in filtered" :key="ch.id"
              class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
              <td class="px-4 py-3">
                <span class="font-mono text-sm font-semibold text-gray-900 dark:text-white">{{ ch.number }}</span>
              </td>
              <td class="px-4 py-3">
                <span :class="ch.type === 'received' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400'"
                  class="px-2 py-0.5 rounded-full text-xs font-semibold capitalize">
                  {{ ch.type }}
                </span>
              </td>
              <td class="px-4 py-3 text-gray-700 dark:text-gray-200">{{ ch.partner_name || '—' }}</td>
              <td class="px-4 py-3 text-gray-500 dark:text-gray-400 hidden md:table-cell text-xs">{{ ch.bank_name || '—' }}</td>
              <td class="px-4 py-3 hidden sm:table-cell">
                <span :class="isPastDue(ch) ? 'text-red-500 font-semibold' : 'text-gray-500 dark:text-gray-400'" class="text-xs">
                  {{ fmtDate(ch.due_date) }}
                </span>
              </td>
              <td class="px-4 py-3 text-right font-semibold text-gray-900 dark:text-white">
                {{ fmtCurrency(ch.amount) }}
              </td>
              <td class="px-4 py-3">
                <span :class="statusClass(ch.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold capitalize">
                  {{ ch.status }}
                </span>
              </td>
              <td class="px-4 py-3 text-right">
                <div class="flex items-center justify-end gap-1">
                  <button v-if="ch.status === 'pending'" @click="depositCheque(ch)"
                    class="p-1.5 rounded-lg bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-100 dark:hover:bg-emerald-900/40 transition-colors"
                    title="Deposit">
                    <ArrowDownToLine class="w-3.5 h-3.5" />
                  </button>
                  <button v-if="ch.status === 'pending'" @click="bounceCheque(ch)"
                    class="p-1.5 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/40 transition-colors"
                    title="Bounce">
                    <X class="w-3.5 h-3.5" />
                  </button>
                  <button @click="editCheque(ch)"
                    class="p-1.5 rounded-lg bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
                    title="Edit">
                    <Pencil class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showModal = false" />
        <div class="relative bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-lg border border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-100 dark:border-gray-800">
            <h3 class="font-bold text-gray-900 dark:text-white text-lg">{{ editMode ? 'Edit Cheque' : 'New Cheque' }}</h3>
            <button @click="showModal = false" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
              <X class="w-5 h-5 text-gray-400" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="label-sm">Number *</label>
                <input v-model="form.number" type="text" :disabled="editMode" class="input-field" />
              </div>
              <div>
                <label class="label-sm">Type *</label>
                <select v-model="form.type" class="input-field">
                  <option value="received">Received</option>
                  <option value="issued">Issued</option>
                </select>
              </div>
              <div class="col-span-2">
                <label class="label-sm">Partner Name</label>
                <input v-model="form.partner_name" type="text" placeholder="Customer / Supplier" class="input-field" />
              </div>
              <div>
                <label class="label-sm">Amount (DZD) *</label>
                <input v-model.number="form.amount" type="number" min="0" step="0.01" class="input-field" />
              </div>
              <div>
                <label class="label-sm">Bank Account</label>
                <select v-model="form.bank_account_id" class="input-field">
                  <option value="">— Select —</option>
                  <option v-for="ba in bankAccounts" :key="ba.id" :value="ba.id">{{ ba.bank_name }} - {{ ba.account_number }}</option>
                </select>
              </div>
              <div>
                <label class="label-sm">Issue Date</label>
                <input v-model="form.issue_date" type="date" class="input-field" />
              </div>
              <div>
                <label class="label-sm">Due Date</label>
                <input v-model="form.due_date" type="date" class="input-field" />
              </div>
            </div>
            <div>
              <label class="label-sm">Notes</label>
              <textarea v-model="form.notes" rows="2" class="input-field resize-none" />
            </div>
          </div>
          <div class="flex gap-3 px-6 py-4 border-t border-gray-100 dark:border-gray-800">
            <button @click="showModal = false"
              class="flex-1 px-4 py-2 rounded-lg border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
              Cancel
            </button>
            <button @click="saveCheque" :disabled="saving"
              class="flex-1 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold transition-colors disabled:opacity-50">
              {{ saving ? 'Saving...' : (editMode ? 'Update' : 'Create Cheque') }}
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
  Pencil, ArrowDownToLine, FileText, CheckCircle2, XCircle, Clock
} from 'lucide-vue-next'
import { treasuryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()
const loading = ref(true)
const saving = ref(false)
const showModal = ref(false)
const editMode = ref(false)
const editId = ref('')

const cheques = ref<any[]>([])
const bankAccounts = ref<any[]>([])
const search = ref('')
const filterType = ref('')
const filterStatus = ref('')

const filtered = computed(() => {
  let list = cheques.value
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(c => c.number?.toLowerCase().includes(q) || c.partner_name?.toLowerCase().includes(q))
  }
  if (filterType.value) list = list.filter(c => c.type === filterType.value)
  if (filterStatus.value) list = list.filter(c => c.status === filterStatus.value)
  return list
})

const kpis = computed(() => {
  const pending = cheques.value.filter(c => c.status === 'pending')
  const deposited = cheques.value.filter(c => c.status === 'deposited')
  const bounced = cheques.value.filter(c => c.status === 'bounced')
  const pastDue = cheques.value.filter(c => c.status === 'pending' && isPastDue(c))
  return [
    { label: 'Pending', icon: Clock, iconBg: 'bg-amber-100 dark:bg-amber-900/30', iconColor: 'text-amber-600 dark:text-amber-400',
      value: pending.length, sub: fmtCurrency(pending.reduce((s, c) => s + c.amount, 0)) },
    { label: 'Deposited', icon: CheckCircle2, iconBg: 'bg-emerald-100 dark:bg-emerald-900/30', iconColor: 'text-emerald-600 dark:text-emerald-400',
      value: deposited.length, sub: fmtCurrency(deposited.reduce((s, c) => s + c.amount, 0)) },
    { label: 'Bounced', icon: XCircle, iconBg: 'bg-red-100 dark:bg-red-900/30', iconColor: 'text-red-600 dark:text-red-400',
      value: bounced.length, sub: fmtCurrency(bounced.reduce((s, c) => s + c.amount, 0)) },
    { label: 'Past Due', icon: FileText, iconBg: 'bg-orange-100 dark:bg-orange-900/30', iconColor: 'text-orange-600 dark:text-orange-400',
      value: pastDue.length, sub: 'Awaiting action' },
  ]
})

const form = ref<any>({ number: '', type: 'received', partner_name: '', amount: 0, bank_account_id: '', issue_date: '', due_date: '', notes: '' })

async function loadCheques() {
  loading.value = true
  try {
    const [chRes, baRes] = await Promise.all([
      treasuryAPI.getCheques(),
      treasuryAPI.getBankAccounts()
    ])
    cheques.value = chRes.data || []
    bankAccounts.value = baRes.data || []
  } catch {
    store.addToast('Failed to load cheques', 'error')
  } finally {
    loading.value = false
  }
}

function openModal() {
  editMode.value = false
  editId.value = ''
  form.value = { number: '', type: 'received', partner_name: '', amount: 0, bank_account_id: '', issue_date: '', due_date: '', notes: '' }
  showModal.value = true
}

function editCheque(ch: any) {
  editMode.value = true
  editId.value = ch.id
  form.value = {
    number: ch.number, type: ch.type, partner_name: ch.partner_name || '',
    amount: ch.amount, bank_account_id: ch.bank_account_id || '',
    issue_date: ch.issue_date ? ch.issue_date.split('T')[0] : '',
    due_date: ch.due_date ? ch.due_date.split('T')[0] : '',
    notes: ch.notes || ''
  }
  showModal.value = true
}

async function saveCheque() {
  if (!editMode.value && (!form.value.number || !form.value.amount)) { store.addToast('Number and amount required', 'error'); return }
  saving.value = true
  try {
    if (editMode.value) {
      await treasuryAPI.updateCheque(editId.value, form.value)
      store.addToast('Cheque updated', 'success')
    } else {
      await treasuryAPI.createCheque(form.value)
      store.addToast('Cheque created', 'success')
    }
    showModal.value = false
    await loadCheques()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Operation failed', 'error')
  } finally {
    saving.value = false
  }
}

async function depositCheque(ch: any) {
  if (!confirm(`Deposit cheque ${ch.number} (${fmtCurrency(ch.amount)})?`)) return
  try {
    await treasuryAPI.depositCheque(ch.id)
    store.addToast('Cheque deposited', 'success')
    await loadCheques()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to deposit', 'error')
  }
}

async function bounceCheque(ch: any) {
  if (!confirm(`Mark cheque ${ch.number} as bounced?`)) return
  try {
    await treasuryAPI.bounceCheque(ch.id)
    store.addToast('Cheque marked as bounced', 'warning')
    await loadCheques()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Operation failed', 'error')
  }
}

function statusClass(s: string) {
  switch (s) {
    case 'pending': return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    case 'deposited': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'bounced': return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    case 'cancelled': return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
    default: return 'bg-gray-100 text-gray-600'
  }
}

function isPastDue(ch: any): boolean {
  if (!ch.due_date || ch.status !== 'pending') return false
  return new Date(ch.due_date) < new Date()
}

function fmtDate(d?: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function fmtCurrency(n?: number) {
  if (!n) return '0 DZD'
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(n) + ' DZD'
}

onMounted(loadCheques)
</script>

<style scoped>
.label-sm { @apply block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5; }
.input-field { @apply w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition-colors; }
</style>
