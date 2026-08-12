<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { helpdeskAPI } from '@/api/client'
import { AlertTriangle, Plus, RefreshCw, X, Search, CheckCircle2, Clock, Edit } from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(true)
const saving = ref(false)
const error = ref('')

const escalations = ref<any[]>([])
const tickets = ref<any[]>([])
const showForm = ref(false)
const showResolveModal = ref(false)
const selectedId = ref<string | null>(null)
const filterStatus = ref('')
const searchQuery = ref('')

const form = ref({ ticket_id: '', escalated_by: '', escalated_to: '', reason: '' })
const resolveForm = ref({ status: 'resolved', resolution_note: '' })

const statusColor: Record<string, string> = {
  pending: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300',
  active: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
  resolved: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
  closed: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const params: Record<string, string> = {}
    if (filterStatus.value) params.status = filterStatus.value
    const [esc, tk] = await Promise.all([
      helpdeskAPI.listEscalations(params),
      helpdeskAPI.listTickets()
    ])
    escalations.value = esc.data.escalations || []
    tickets.value = tk.data.tickets || []
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to load'
  } finally {
    loading.value = false
  }
}

async function createEscalation() {
  if (!form.value.ticket_id) { error.value = 'Select a ticket'; return }
  saving.value = true
  error.value = ''
  try {
    await helpdeskAPI.createEscalation(form.value)
    showForm.value = false
    form.value = { ticket_id: '', escalated_by: '', escalated_to: '', reason: '' }
    await load()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to create'
  } finally {
    saving.value = false
  }
}

async function resolveEscalation() {
  if (!selectedId.value) return
  saving.value = true
  try {
    await helpdeskAPI.updateEscalationStatus(selectedId.value, resolveForm.value)
    showResolveModal.value = false
    await load()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to update'
  } finally {
    saving.value = false
  }
}

async function deleteEscalation(id: string) {
  if (!confirm('Delete this escalation?')) return
  try {
    await helpdeskAPI.deleteEscalation(id)
    await load()
  } catch {}
}

function openResolve(id: string) {
  selectedId.value = id
  resolveForm.value = { status: 'resolved', resolution_note: '' }
  showResolveModal.value = true
}

const filtered = () => {
  if (!searchQuery.value) return escalations.value
  const q = searchQuery.value.toLowerCase()
  return escalations.value.filter(e =>
    e.ticket_number?.toLowerCase().includes(q) ||
    e.subject?.toLowerCase().includes(q) ||
    e.escalated_by?.toLowerCase().includes(q)
  )
}

onMounted(load)

// KPIs
const kpiActive = () => escalations.value.filter(e => e.status === 'active').length
const kpiPending = () => escalations.value.filter(e => e.status === 'pending').length
const kpiResolved = () => escalations.value.filter(e => e.status === 'resolved').length
</script>

<template>
  <div :class="app.darkMode ? 'bg-slate-900 text-slate-100' : 'bg-slate-50 text-slate-900'" class="min-h-screen p-6">
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <AlertTriangle class="w-7 h-7 text-red-500" />
        <div>
          <h1 class="text-2xl font-bold">Escalations</h1>
          <p class="text-sm text-slate-500">{{ escalations.length }} escalations</p>
        </div>
      </div>
      <div class="flex gap-2">
        <button @click="load" :class="app.darkMode ? 'bg-slate-700 text-slate-200' : 'bg-white text-slate-700'"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-600 hover:bg-slate-100 dark:hover:bg-slate-600 transition-colors">
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
        </button>
        <button @click="showForm = true" class="flex items-center gap-2 px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" /> New Escalation
        </button>
      </div>
    </div>

    <div v-if="error" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg text-sm">{{ error }}</div>

    <!-- KPIs -->
    <div class="grid grid-cols-3 gap-4 mb-6">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'" class="rounded-xl border p-4 shadow-sm text-center">
        <div class="text-2xl font-bold text-red-600">{{ kpiActive() }}</div>
        <div class="text-sm text-slate-500">Active</div>
      </div>
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'" class="rounded-xl border p-4 shadow-sm text-center">
        <div class="text-2xl font-bold text-yellow-600">{{ kpiPending() }}</div>
        <div class="text-sm text-slate-500">Pending</div>
      </div>
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'" class="rounded-xl border p-4 shadow-sm text-center">
        <div class="text-2xl font-bold text-green-600">{{ kpiResolved() }}</div>
        <div class="text-sm text-slate-500">Resolved</div>
      </div>
    </div>

    <!-- Filters -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border p-4 mb-4 shadow-sm flex gap-3">
      <div class="relative flex-1">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input v-model="searchQuery" placeholder="Search escalations..."
          :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
          class="w-full pl-9 pr-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
      </div>
      <select v-model="filterStatus" @change="load"
        :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
        class="px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
        <option value="">All Status</option>
        <option value="pending">Pending</option>
        <option value="active">Active</option>
        <option value="resolved">Resolved</option>
        <option value="closed">Closed</option>
      </select>
    </div>

    <!-- Table -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border shadow-sm overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center py-16">
        <RefreshCw class="w-6 h-6 animate-spin text-indigo-500" />
      </div>
      <div v-else-if="!filtered().length" class="text-center py-16">
        <AlertTriangle class="w-12 h-12 text-slate-300 mx-auto mb-3" />
        <p class="text-slate-400">No escalations found</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'bg-slate-700/50 text-slate-400' : 'bg-slate-50 text-slate-500'"
              class="text-xs uppercase tracking-wide">
              <th class="text-left px-4 py-3 font-medium">Ticket</th>
              <th class="text-left px-4 py-3 font-medium">Subject</th>
              <th class="text-left px-4 py-3 font-medium">Priority</th>
              <th class="text-left px-4 py-3 font-medium">Escalated By</th>
              <th class="text-left px-4 py-3 font-medium">Escalated To</th>
              <th class="text-left px-4 py-3 font-medium">Reason</th>
              <th class="text-left px-4 py-3 font-medium">Status</th>
              <th class="text-left px-4 py-3 font-medium">Date</th>
              <th class="text-left px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="e in filtered()" :key="e.id"
              :class="app.darkMode ? 'border-slate-700 hover:bg-slate-700/50' : 'border-slate-100 hover:bg-slate-50'"
              class="border-b transition-colors">
              <td class="px-4 py-3 font-mono text-xs text-indigo-600 font-medium">{{ e.ticket_number }}</td>
              <td class="px-4 py-3 max-w-[160px] truncate">{{ e.subject }}</td>
              <td class="px-4 py-3">
                <span :class="e.priority === 'critical' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300' : 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300'"
                  class="text-xs px-2 py-0.5 rounded-full font-medium capitalize">{{ e.priority }}</span>
              </td>
              <td class="px-4 py-3 text-xs">{{ e.escalated_by || '-' }}</td>
              <td class="px-4 py-3 text-xs">{{ e.escalated_to || '-' }}</td>
              <td class="px-4 py-3 text-xs text-slate-500 max-w-[140px] truncate">{{ e.reason || '-' }}</td>
              <td class="px-4 py-3">
                <span :class="statusColor[e.status] || ''" class="text-xs px-2 py-0.5 rounded-full font-medium capitalize">{{ e.status }}</span>
              </td>
              <td class="px-4 py-3 text-xs text-slate-400">{{ e.escalated_at?.substring(0,10) }}</td>
              <td class="px-4 py-3">
                <div class="flex gap-1">
                  <button v-if="e.status === 'active' || e.status === 'pending'"
                    @click="openResolve(e.id)"
                    class="p-1.5 text-slate-400 hover:text-green-600 hover:bg-green-50 dark:hover:bg-green-900/20 rounded transition-colors" title="Resolve">
                    <CheckCircle2 class="w-4 h-4" />
                  </button>
                  <button @click="deleteEscalation(e.id)"
                    class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors" title="Delete">
                    <X class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create Modal -->
    <div v-if="showForm" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="w-full max-w-md rounded-2xl border shadow-xl">
        <div class="flex items-center justify-between p-6 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold">New Escalation</h2>
          <button @click="showForm = false" class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg"><X class="w-5 h-5" /></button>
        </div>
        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1">Ticket *</label>
            <select v-model="form.ticket_id" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
              <option value="">Select ticket...</option>
              <option v-for="t in tickets" :key="t.id" :value="t.id">{{ t.ticket_number }} — {{ t.subject }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Escalated By</label>
            <input v-model="form.escalated_by" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Escalated To</label>
            <input v-model="form.escalated_to" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Reason *</label>
            <textarea v-model="form.reason" rows="3" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none resize-none" />
          </div>
        </div>
        <div class="flex justify-end gap-3 p-6 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showForm = false" :class="app.darkMode ? 'bg-slate-700 text-slate-200 hover:bg-slate-600' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
            class="px-4 py-2 rounded-lg text-sm font-medium">Cancel</button>
          <button @click="createEscalation" :disabled="saving" class="px-4 py-2 bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white rounded-lg text-sm font-medium flex items-center gap-2">
            <RefreshCw v-if="saving" class="w-4 h-4 animate-spin" />
            Escalate
          </button>
        </div>
      </div>
    </div>

    <!-- Resolve Modal -->
    <div v-if="showResolveModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="w-full max-w-md rounded-2xl border shadow-xl">
        <div class="flex items-center justify-between p-6 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold">Resolve Escalation</h2>
          <button @click="showResolveModal = false" class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg"><X class="w-5 h-5" /></button>
        </div>
        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1">Status</label>
            <select v-model="resolveForm.status" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
              <option value="resolved">Resolved</option>
              <option value="closed">Closed</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Resolution Note</label>
            <textarea v-model="resolveForm.resolution_note" rows="3" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none resize-none" />
          </div>
        </div>
        <div class="flex justify-end gap-3 p-6 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showResolveModal = false" :class="app.darkMode ? 'bg-slate-700 text-slate-200 hover:bg-slate-600' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
            class="px-4 py-2 rounded-lg text-sm font-medium">Cancel</button>
          <button @click="resolveEscalation" :disabled="saving" class="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white rounded-lg text-sm font-medium flex items-center gap-2">
            <RefreshCw v-if="saving" class="w-4 h-4 animate-spin" />
            Confirm
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
