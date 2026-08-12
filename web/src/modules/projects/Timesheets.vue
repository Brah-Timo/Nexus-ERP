<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Timesheets</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Track and manage time entries across projects</p>
      </div>
      <button @click="openCreate" class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm">
        <Plus class="w-4 h-4" />
        Log Time
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center flex-shrink-0">
          <Clock class="w-6 h-6 text-indigo-600 dark:text-indigo-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Total Hours</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ totalHours.toFixed(1) }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-emerald-100 dark:bg-emerald-900/40 flex items-center justify-center flex-shrink-0">
          <DollarSign class="w-6 h-6 text-emerald-600 dark:text-emerald-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Billable Hours</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ billableHours.toFixed(1) }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center flex-shrink-0">
          <CheckCircle class="w-6 h-6 text-violet-600 dark:text-violet-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Approved</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ approvedCount }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-amber-100 dark:bg-amber-900/40 flex items-center justify-center flex-shrink-0">
          <FileText class="w-6 h-6 text-amber-600 dark:text-amber-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Entries</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ timesheets.length }}</p>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
      <div class="flex flex-wrap gap-3">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input v-model="search" placeholder="Search employee, project..." class="w-full pl-9 pr-4 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white placeholder-gray-400" />
        </div>
        <div class="relative">
          <select v-model="filterProject" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="">All Projects</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
        <div class="relative">
          <select v-model="filterMonth" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="">All Months</option>
            <option v-for="m in months" :key="m.value" :value="m.value">{{ m.label }}</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
        <div class="relative">
          <select v-model="filterYear" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
        <button @click="load" class="inline-flex items-center gap-2 px-4 py-2 bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-200 text-sm font-medium rounded-lg transition-colors">
          <RefreshCw class="w-4 h-4" />
          Refresh
        </button>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center py-16">
        <Loader2 class="w-8 h-8 text-indigo-500 animate-spin" />
      </div>
      <div v-else-if="filteredTimesheets.length === 0" class="flex flex-col items-center justify-center py-16 text-gray-400">
        <Clock class="w-12 h-12 mb-3 opacity-30" />
        <p class="text-sm">No time entries found</p>
      </div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Date</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Employee</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Project</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Task</th>
            <th class="text-right px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Hours</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Billable</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Approved</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Description</th>
            <th class="px-4 py-3 w-20"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
          <tr v-for="ts in filteredTimesheets" :key="ts.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
            <td class="px-4 py-3 font-medium text-gray-900 dark:text-white whitespace-nowrap">{{ fmtDate(ts.date) }}</td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2">
                <div class="w-6 h-6 rounded-full bg-gradient-to-br from-indigo-400 to-violet-500 flex items-center justify-center text-white text-xs font-bold flex-shrink-0">
                  {{ (ts.employee_name || '?').charAt(0) }}
                </div>
                <span class="text-gray-700 dark:text-gray-200 truncate max-w-28">{{ ts.employee_name }}</span>
              </div>
            </td>
            <td class="px-4 py-3 text-gray-600 dark:text-gray-300 text-xs truncate max-w-32">{{ ts.project_name }}</td>
            <td class="px-4 py-3 text-gray-400 text-xs truncate max-w-28">{{ ts.task_title || '—' }}</td>
            <td class="px-4 py-3 text-right font-bold text-indigo-600 dark:text-indigo-400">{{ ts.hours?.toFixed(1) }}h</td>
            <td class="px-4 py-3">
              <span :class="ts.billable ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'" class="px-2 py-0.5 rounded-full text-xs font-semibold">{{ ts.billable ? 'Billable' : 'Internal' }}</span>
            </td>
            <td class="px-4 py-3">
              <span :class="ts.approved ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'" class="px-2 py-0.5 rounded-full text-xs font-semibold">{{ ts.approved ? 'Approved' : 'Pending' }}</span>
            </td>
            <td class="px-4 py-3 text-gray-400 text-xs max-w-40 truncate">{{ ts.description || '—' }}</td>
            <td class="px-4 py-3" @click.stop>
              <div class="flex gap-1">
                <button v-if="!ts.approved" @click="approveEntry(ts.id)" class="p-1.5 hover:bg-emerald-50 dark:hover:bg-emerald-900/30 rounded-lg text-gray-400 hover:text-emerald-600 transition-colors" title="Approve">
                  <CheckCircle class="w-3.5 h-3.5" />
                </button>
                <button @click="openEdit(ts)" class="p-1.5 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg text-gray-400 hover:text-indigo-600 transition-colors">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button @click="confirmDelete(ts)" class="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg text-gray-400 hover:text-red-600 transition-colors">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
        <tfoot>
          <tr class="border-t-2 border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/60">
            <td colspan="4" class="px-4 py-3 font-semibold text-gray-700 dark:text-gray-200 text-sm">Total</td>
            <td class="px-4 py-3 text-right font-bold text-indigo-600 dark:text-indigo-400 text-sm">{{ filteredTimesheets.reduce((s, t) => s + (t.hours || 0), 0).toFixed(1) }}h</td>
            <td colspan="4"></td>
          </tr>
        </tfoot>
      </table>
    </div>

    <!-- Log Time Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeModal" />
          <div class="relative w-full max-w-lg bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden">
            <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-indigo-600 to-violet-600">
              <div class="flex items-center gap-3 text-white">
                <Clock class="w-5 h-5" />
                <h2 class="font-semibold text-lg">{{ editingTs ? 'Edit Time Entry' : 'Log Time' }}</h2>
              </div>
              <button @click="closeModal" class="p-1 hover:bg-white/20 rounded-lg text-white transition-colors">
                <X class="w-5 h-5" />
              </button>
            </div>
            <form @submit.prevent="saveTimesheet" class="p-6 space-y-4">
              <div class="grid grid-cols-2 gap-4">
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
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Date *</label>
                  <input type="date" v-model="form.date" required class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Hours *</label>
                  <input type="number" v-model.number="form.hours" required min="0.1" max="24" step="0.5" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Hourly Rate (DZD)</label>
                  <input type="number" v-model.number="form.hourly_rate" min="0" step="100" placeholder="0" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div class="flex items-end pb-1 gap-6">
                  <label class="flex items-center gap-2 cursor-pointer">
                    <input type="checkbox" v-model="form.billable" class="w-4 h-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" />
                    <span class="text-sm font-medium text-gray-700 dark:text-gray-200">Billable</span>
                  </label>
                </div>
              </div>
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Description</label>
                <textarea v-model="form.description" rows="3" placeholder="What was worked on..." class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white resize-none" />
              </div>
              <div class="flex gap-3">
                <button type="button" @click="closeModal" class="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
                <button type="submit" :disabled="saving" class="flex-1 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center justify-center gap-2">
                  <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                  {{ saving ? 'Saving...' : (editingTs ? 'Update' : 'Log Time') }}
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
  Plus, Search, Loader2, Clock, DollarSign, CheckCircle, FileText,
  Pencil, Trash2, X, ChevronDown, RefreshCw
} from 'lucide-vue-next'
import { projectsAPI, hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

const loading = ref(true)
const saving = ref(false)
const timesheets = ref<any[]>([])
const projects = ref<any[]>([])
const employees = ref<any[]>([])
const search = ref('')
const filterProject = ref('')
const filterMonth = ref(new Date().getMonth() + 1 + '')
const filterYear = ref(new Date().getFullYear() + '')
const showModal = ref(false)
const editingTs = ref<any>(null)
const deletingTs = ref<any>(null)

const form = ref({
  employee_id: '', project_id: '', task_id: '', date: new Date().toISOString().slice(0, 10),
  hours: 0, hourly_rate: 0, description: '', billable: false
})

const months = [
  { value: '1', label: 'January' }, { value: '2', label: 'February' }, { value: '3', label: 'March' },
  { value: '4', label: 'April' }, { value: '5', label: 'May' }, { value: '6', label: 'June' },
  { value: '7', label: 'July' }, { value: '8', label: 'August' }, { value: '9', label: 'September' },
  { value: '10', label: 'October' }, { value: '11', label: 'November' }, { value: '12', label: 'December' }
]
const currentYear = new Date().getFullYear()
const years = [currentYear - 2, currentYear - 1, currentYear, currentYear + 1].map(String)

const totalHours = computed(() => timesheets.value.reduce((s, t) => s + (t.hours || 0), 0))
const billableHours = computed(() => timesheets.value.filter(t => t.billable).reduce((s, t) => s + (t.hours || 0), 0))
const approvedCount = computed(() => timesheets.value.filter(t => t.approved).length)

const filteredTimesheets = computed(() => {
  let list = [...timesheets.value]
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(t => (t.employee_name || '').toLowerCase().includes(q) || (t.project_name || '').toLowerCase().includes(q))
  }
  if (filterProject.value) list = list.filter(t => t.project_id === filterProject.value)
  return list
})

function fmtDate(d?: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

async function load() {
  loading.value = true
  try {
    const params: any = {}
    if (filterProject.value) params.project_id = filterProject.value
    if (filterMonth.value) params.month = filterMonth.value
    if (filterYear.value) params.year = filterYear.value
    const res = await projectsAPI.getTimesheets(params)
    timesheets.value = res.data || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load timesheets', 'error')
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
  editingTs.value = null
  form.value = { employee_id: '', project_id: '', task_id: '', date: new Date().toISOString().slice(0, 10), hours: 0, hourly_rate: 0, description: '', billable: false }
  showModal.value = true
}
function openEdit(ts: any) {
  editingTs.value = ts
  form.value = {
    employee_id: ts.employee_id, project_id: ts.project_id, task_id: ts.task_id || '',
    date: ts.date?.slice(0, 10) || '', hours: ts.hours, hourly_rate: ts.hourly_rate || 0,
    description: ts.description || '', billable: ts.billable || false
  }
  showModal.value = true
}
function closeModal() { showModal.value = false; editingTs.value = null }

async function saveTimesheet() {
  saving.value = true
  try {
    const payload = { ...form.value, task_id: form.value.task_id || null }
    if (editingTs.value) {
      await projectsAPI.updateTimesheet(editingTs.value.id, payload)
      store.addToast('Time entry updated', 'success')
    } else {
      await projectsAPI.createTimesheet(payload)
      store.addToast('Time logged', 'success')
    }
    closeModal()
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

async function approveEntry(id: string) {
  try {
    await projectsAPI.updateTimesheet(id, { approved: true })
    store.addToast('Entry approved', 'success')
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Approve failed', 'error')
  }
}

async function confirmDelete(ts: any) {
  if (!confirm(`Delete time entry for ${ts.employee_name} on ${fmtDate(ts.date)}?`)) return
  try {
    await projectsAPI.deleteTimesheet(ts.id)
    store.addToast('Entry deleted', 'success')
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
