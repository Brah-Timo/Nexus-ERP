<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Project Expenses</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Track and approve project-related expenses</p>
      </div>
      <button @click="openCreate" class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm">
        <Plus class="w-4 h-4" />
        New Expense
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center flex-shrink-0">
          <Receipt class="w-6 h-6 text-indigo-600 dark:text-indigo-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Total</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(totalAmount) }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-emerald-100 dark:bg-emerald-900/40 flex items-center justify-center flex-shrink-0">
          <CheckCircle class="w-6 h-6 text-emerald-600 dark:text-emerald-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Approved</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(approvedAmount) }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-amber-100 dark:bg-amber-900/40 flex items-center justify-center flex-shrink-0">
          <Clock class="w-6 h-6 text-amber-600 dark:text-amber-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Pending</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ pendingCount }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center flex-shrink-0">
          <Wallet class="w-6 h-6 text-violet-600 dark:text-violet-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Paid</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ fmtCurrency(paidAmount) }}</p>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
      <div class="flex flex-wrap gap-3">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input v-model="search" placeholder="Search expenses..." class="w-full pl-9 pr-4 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white placeholder-gray-400" />
        </div>
        <div class="relative">
          <select v-model="filterProject" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="">All Projects</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
        <div class="relative">
          <select v-model="filterStatus" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="">All Status</option>
            <option value="draft">Draft</option>
            <option value="submitted">Submitted</option>
            <option value="approved">Approved</option>
            <option value="rejected">Rejected</option>
            <option value="paid">Paid</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
        <div class="relative">
          <select v-model="filterCategory" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="">All Categories</option>
            <option v-for="cat in categories" :key="cat.value" :value="cat.value">{{ cat.label }}</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center py-16">
        <Loader2 class="w-8 h-8 text-indigo-500 animate-spin" />
      </div>
      <div v-else-if="filteredExpenses.length === 0" class="flex flex-col items-center justify-center py-16 text-gray-400">
        <Receipt class="w-12 h-12 mb-3 opacity-30" />
        <p class="text-sm">No expenses found</p>
      </div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Description</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Project</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Employee</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Category</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Date</th>
            <th class="text-right px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Amount</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Status</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Billable</th>
            <th class="px-4 py-3 w-24"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
          <tr v-for="exp in filteredExpenses" :key="exp.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
            <td class="px-4 py-3 font-medium text-gray-900 dark:text-white max-w-48 truncate">{{ exp.description }}</td>
            <td class="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs truncate max-w-28">{{ exp.project_name }}</td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <div class="w-6 h-6 rounded-full bg-gradient-to-br from-indigo-400 to-violet-500 flex items-center justify-center text-white text-xs font-bold flex-shrink-0">{{ (exp.employee_name || '?').charAt(0) }}</div>
                <span class="text-sm text-gray-600 dark:text-gray-300 truncate max-w-24">{{ exp.employee_name }}</span>
              </div>
            </td>
            <td class="px-4 py-3">
              <span :class="categoryBadge(exp.category)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold capitalize">{{ exp.category }}</span>
            </td>
            <td class="px-4 py-3 text-gray-600 dark:text-gray-300 whitespace-nowrap">{{ fmtDate(exp.expense_date) }}</td>
            <td class="px-4 py-3 text-right font-bold text-gray-900 dark:text-white">{{ fmtCurrency(exp.amount) }}</td>
            <td class="px-4 py-3">
              <span :class="expStatusBadge(exp.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold capitalize">{{ exp.status }}</span>
            </td>
            <td class="px-4 py-3">
              <span :class="exp.is_billable ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'" class="px-2 py-0.5 rounded-full text-xs font-semibold">
                {{ exp.is_billable ? 'Billable' : 'Internal' }}
              </span>
            </td>
            <td class="px-4 py-3" @click.stop>
              <div class="flex gap-1">
                <button v-if="exp.status === 'submitted'" @click="approveExpense(exp)" class="p-1.5 hover:bg-emerald-50 dark:hover:bg-emerald-900/30 rounded-lg text-gray-400 hover:text-emerald-600 transition-colors" title="Approve">
                  <CheckCircle class="w-3.5 h-3.5" />
                </button>
                <button v-if="exp.status === 'draft'" @click="submitExpense(exp)" class="p-1.5 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded-lg text-gray-400 hover:text-blue-600 transition-colors" title="Submit">
                  <Send class="w-3.5 h-3.5" />
                </button>
                <button @click="openEdit(exp)" class="p-1.5 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg text-gray-400 hover:text-indigo-600 transition-colors">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button @click="confirmDelete(exp)" class="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg text-gray-400 hover:text-red-600 transition-colors">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
        <tfoot>
          <tr class="border-t-2 border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/60">
            <td colspan="5" class="px-4 py-3 font-semibold text-gray-700 dark:text-gray-200 text-sm">Total</td>
            <td class="px-4 py-3 text-right font-bold text-indigo-600 dark:text-indigo-400 text-sm">{{ fmtCurrency(filteredExpenses.reduce((s, e) => s + (e.amount || 0), 0)) }}</td>
            <td colspan="3"></td>
          </tr>
        </tfoot>
      </table>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeModal" />
          <div class="relative w-full max-w-lg bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 max-h-[90vh] flex flex-col">
            <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-indigo-600 to-violet-600 rounded-t-2xl flex-shrink-0">
              <div class="flex items-center gap-3 text-white">
                <Receipt class="w-5 h-5" />
                <h2 class="font-semibold text-lg">{{ editingExp ? 'Edit Expense' : 'New Expense' }}</h2>
              </div>
              <button @click="closeModal" class="p-1 hover:bg-white/20 rounded-lg text-white transition-colors">
                <X class="w-5 h-5" />
              </button>
            </div>
            <form @submit.prevent="saveExpense" class="p-6 space-y-4 overflow-y-auto">
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Description *</label>
                <input v-model="form.description" required placeholder="Expense description" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Project *</label>
                  <div class="relative">
                    <select v-model="form.project_id" required class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="">— Select Project —</option>
                      <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Employee *</label>
                  <div class="relative">
                    <select v-model="form.employee_id" required class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="">— Select Employee —</option>
                      <option v-for="e in employees" :key="e.id" :value="e.id">{{ e.full_name }}</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Category *</label>
                  <div class="relative">
                    <select v-model="form.category" required class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option v-for="cat in categories" :key="cat.value" :value="cat.value">{{ cat.label }}</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Date *</label>
                  <input type="date" v-model="form.expense_date" required class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Amount (DZD) *</label>
                  <input type="number" v-model.number="form.amount" required min="0" step="100" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Status</label>
                  <div class="relative">
                    <select v-model="form.status" class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="draft">Draft</option>
                      <option value="submitted">Submitted</option>
                      <option value="approved">Approved</option>
                      <option value="rejected">Rejected</option>
                      <option value="paid">Paid</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div class="col-span-2 flex gap-6">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" v-model="form.is_billable" class="w-4 h-4 rounded border-gray-300 text-indigo-600" />
                    <span class="text-sm font-medium text-gray-700 dark:text-gray-200">Billable</span>
                  </label>
                </div>
              </div>
              <div class="flex gap-3">
                <button type="button" @click="closeModal" class="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
                <button type="submit" :disabled="saving" class="flex-1 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center justify-center gap-2">
                  <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                  {{ saving ? 'Saving...' : (editingExp ? 'Update' : 'Create') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Plus, Search, Loader2, Receipt, CheckCircle, Clock, Wallet,
  Pencil, Trash2, X, ChevronDown, Send
} from 'lucide-vue-next'
import { projectsAPI, hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

const loading = ref(true)
const saving = ref(false)
const expenses = ref<any[]>([])
const projects = ref<any[]>([])
const employees = ref<any[]>([])
const search = ref('')
const filterProject = ref('')
const filterStatus = ref('')
const filterCategory = ref('')
const showModal = ref(false)
const editingExp = ref<any>(null)

const categories = [
  { value: 'travel', label: 'Travel' }, { value: 'accommodation', label: 'Accommodation' },
  { value: 'equipment', label: 'Equipment' }, { value: 'software', label: 'Software' },
  { value: 'consulting', label: 'Consulting' }, { value: 'materials', label: 'Materials' },
  { value: 'utilities', label: 'Utilities' }, { value: 'communication', label: 'Communication' },
  { value: 'other', label: 'Other' }
]

const form = ref({
  project_id: '', employee_id: '', category: 'other', description: '',
  amount: 0, expense_date: new Date().toISOString().slice(0, 10),
  status: 'draft', is_billable: false
})

const totalAmount = computed(() => expenses.value.reduce((s, e) => s + (e.amount || 0), 0))
const approvedAmount = computed(() => expenses.value.filter(e => e.status === 'approved').reduce((s, e) => s + (e.amount || 0), 0))
const paidAmount = computed(() => expenses.value.filter(e => e.status === 'paid').reduce((s, e) => s + (e.amount || 0), 0))
const pendingCount = computed(() => expenses.value.filter(e => e.status === 'submitted' || e.status === 'draft').length)

const filteredExpenses = computed(() => {
  let list = [...expenses.value]
  if (search.value) { const q = search.value.toLowerCase(); list = list.filter(e => e.description?.toLowerCase().includes(q) || e.employee_name?.toLowerCase().includes(q)) }
  if (filterProject.value) list = list.filter(e => e.project_id === filterProject.value)
  if (filterStatus.value) list = list.filter(e => e.status === filterStatus.value)
  if (filterCategory.value) list = list.filter(e => e.category === filterCategory.value)
  return list
})

function fmtDate(d?: string) { if (!d) return '—'; return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' }) }
function fmtCurrency(n?: number) {
  if (!n) return '0 DZD'
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(2) + 'M DZD'
  if (n >= 1_000) return (n / 1_000).toFixed(0) + 'K DZD'
  return n.toLocaleString() + ' DZD'
}

function categoryBadge(cat?: string) {
  const colors: Record<string, string> = {
    travel: 'bg-sky-100 text-sky-700 dark:bg-sky-900/30 dark:text-sky-400',
    accommodation: 'bg-violet-100 text-violet-700 dark:bg-violet-900/30 dark:text-violet-400',
    equipment: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400',
    software: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    consulting: 'bg-pink-100 text-pink-700 dark:bg-pink-900/30 dark:text-pink-400',
    materials: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
    other: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
  }
  return colors[cat || 'other'] || colors.other
}
function expStatusBadge(s?: string) {
  switch (s) {
    case 'approved': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'submitted': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    case 'rejected': return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    case 'paid': return 'bg-violet-100 text-violet-700 dark:bg-violet-900/30 dark:text-violet-400'
    default: return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
  }
}

async function load() {
  loading.value = true
  try {
    const params: any = {}
    if (filterProject.value) params.project_id = filterProject.value
    if (filterStatus.value) params.status = filterStatus.value
    const res = await projectsAPI.getExpenses(params)
    expenses.value = res.data || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load expenses', 'error')
  } finally {
    loading.value = false
  }
}

async function loadDropdowns() {
  try {
    const [pRes, eRes] = await Promise.all([projectsAPI.getProjects(), hrAPI.getEmployees()])
    projects.value = pRes.data || []
    employees.value = eRes.data || []
  } catch { /* ignore */ }
}

function openCreate() {
  editingExp.value = null
  form.value = { project_id: '', employee_id: '', category: 'other', description: '', amount: 0, expense_date: new Date().toISOString().slice(0, 10), status: 'draft', is_billable: false }
  showModal.value = true
}
function openEdit(exp: any) {
  editingExp.value = exp
  form.value = { project_id: exp.project_id, employee_id: exp.employee_id, category: exp.category, description: exp.description, amount: exp.amount, expense_date: exp.expense_date?.slice(0, 10) || '', status: exp.status, is_billable: exp.is_billable }
  showModal.value = true
}
function closeModal() { showModal.value = false; editingExp.value = null }

async function saveExpense() {
  saving.value = true
  try {
    if (editingExp.value) {
      await projectsAPI.updateExpense(editingExp.value.id, form.value)
      store.addToast('Expense updated', 'success')
    } else {
      await projectsAPI.createExpense(form.value)
      store.addToast('Expense created', 'success')
    }
    closeModal(); await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

async function approveExpense(exp: any) {
  try {
    await projectsAPI.updateExpense(exp.id, { ...exp, status: 'approved' })
    store.addToast('Expense approved', 'success')
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Approve failed', 'error')
  }
}

async function submitExpense(exp: any) {
  try {
    await projectsAPI.updateExpense(exp.id, { ...exp, status: 'submitted' })
    store.addToast('Expense submitted', 'success')
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Submit failed', 'error')
  }
}

async function confirmDelete(exp: any) {
  if (!confirm(`Delete expense "${exp.description}"?`)) return
  try {
    await projectsAPI.deleteExpense(exp.id)
    store.addToast('Expense deleted', 'success')
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Delete failed', 'error')
  }
}

onMounted(() => { load(); loadDropdowns() })
</script>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
