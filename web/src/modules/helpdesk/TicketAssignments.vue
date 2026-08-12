<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { helpdeskAPI } from '@/api/client'
import { UserPlus, RefreshCw, Search, Ticket, User, ArrowRight, CheckCircle2, Clock } from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(true)
const saving = ref(false)
const error = ref('')

const assignments = ref<any[]>([])
const agents = ref<any[]>([])
const tickets = ref<any[]>([])

const filterAgentId = ref('')
const searchQuery = ref('')
const showAssignModal = ref(false)

const assignForm = ref({ ticket_id: '', agent_id: '', assigned_by: '', reason: '' })

async function load() {
  loading.value = true
  error.value = ''
  try {
    const params: Record<string, string> = {}
    if (filterAgentId.value) params.agent_id = filterAgentId.value
    const [asgn, ag, tk] = await Promise.all([
      helpdeskAPI.listAssignments(params),
      helpdeskAPI.listAgents(),
      helpdeskAPI.listTickets({ status: 'open' })
    ])
    assignments.value = asgn.data.assignments || []
    agents.value = ag.data.agents || []
    tickets.value = tk.data.tickets || []
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to load'
  } finally {
    loading.value = false
  }
}

async function assign() {
  if (!assignForm.value.ticket_id) { error.value = 'Select a ticket'; return }
  saving.value = true
  error.value = ''
  try {
    await helpdeskAPI.assignTicket(assignForm.value.ticket_id, {
      agent_id: assignForm.value.agent_id,
      assigned_by: assignForm.value.assigned_by,
      reason: assignForm.value.reason
    })
    showAssignModal.value = false
    assignForm.value = { ticket_id: '', agent_id: '', assigned_by: '', reason: '' }
    await load()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to assign'
  } finally {
    saving.value = false
  }
}

const filteredAssignments = () => {
  if (!searchQuery.value) return assignments.value
  const q = searchQuery.value.toLowerCase()
  return assignments.value.filter(a =>
    a.ticket_number?.toLowerCase().includes(q) ||
    a.subject?.toLowerCase().includes(q) ||
    a.agent_name?.toLowerCase().includes(q)
  )
}

onMounted(load)
</script>

<template>
  <div :class="app.darkMode ? 'bg-slate-900 text-slate-100' : 'bg-slate-50 text-slate-900'" class="min-h-screen p-6">
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <UserPlus class="w-7 h-7 text-indigo-500" />
        <div>
          <h1 class="text-2xl font-bold">Ticket Assignments</h1>
          <p class="text-sm text-slate-500">{{ assignments.length }} assignment records</p>
        </div>
      </div>
      <div class="flex gap-2">
        <button @click="load" :class="app.darkMode ? 'bg-slate-700 text-slate-200' : 'bg-white text-slate-700'"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-600 hover:bg-slate-100 dark:hover:bg-slate-600 transition-colors">
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
        </button>
        <button @click="showAssignModal = true" class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors">
          <UserPlus class="w-4 h-4" /> Assign Ticket
        </button>
      </div>
    </div>

    <div v-if="error" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg text-sm">{{ error }}</div>

    <!-- Filters -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border p-4 mb-4 shadow-sm flex gap-3">
      <div class="relative flex-1">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input v-model="searchQuery" placeholder="Search tickets or agents..."
          :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
          class="w-full pl-9 pr-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
      </div>
      <select v-model="filterAgentId" @change="load"
        :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
        class="px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
        <option value="">All Agents</option>
        <option v-for="a in agents" :key="a.id" :value="a.id">{{ a.name }}</option>
      </select>
    </div>

    <!-- Agent workload summary -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
      <div v-for="a in agents.slice(0,4)" :key="a.id"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-4 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm font-medium truncate">{{ a.name }}</span>
          <span class="text-xs" :class="a.status === 'active' ? 'text-green-500' : 'text-slate-400'">
            {{ a.status }}
          </span>
        </div>
        <div class="flex items-end gap-2">
          <div class="text-2xl font-bold text-indigo-600">{{ a.open_tickets }}</div>
          <div class="text-xs text-slate-400 mb-0.5">/ {{ a.max_tickets }}</div>
        </div>
        <div class="mt-2 bg-slate-100 dark:bg-slate-700 rounded-full h-1.5">
          <div :style="`width:${Math.min(100,a.open_tickets/Math.max(a.max_tickets,1)*100)}%`"
            :class="a.open_tickets >= a.max_tickets ? 'bg-red-500' : 'bg-indigo-500'"
            class="h-1.5 rounded-full"></div>
        </div>
      </div>
    </div>

    <!-- Assignments Table -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border shadow-sm overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center py-16">
        <RefreshCw class="w-6 h-6 animate-spin text-indigo-500" />
      </div>
      <div v-else-if="!filteredAssignments().length" class="text-center py-16">
        <UserPlus class="w-12 h-12 text-slate-300 mx-auto mb-3" />
        <p class="text-slate-400">No assignments found</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'bg-slate-700/50 text-slate-400' : 'bg-slate-50 text-slate-500'"
              class="text-xs uppercase tracking-wide">
              <th class="text-left px-4 py-3 font-medium">Ticket</th>
              <th class="text-left px-4 py-3 font-medium">Subject</th>
              <th class="text-left px-4 py-3 font-medium">Agent</th>
              <th class="text-left px-4 py-3 font-medium">Assigned By</th>
              <th class="text-left px-4 py-3 font-medium">Reason</th>
              <th class="text-left px-4 py-3 font-medium">Assigned At</th>
              <th class="text-left px-4 py-3 font-medium">Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="a in filteredAssignments()" :key="a.id"
              :class="app.darkMode ? 'border-slate-700 hover:bg-slate-700/50' : 'border-slate-100 hover:bg-slate-50'"
              class="border-b transition-colors">
              <td class="px-4 py-3 font-mono text-xs text-indigo-600 font-medium">{{ a.ticket_number }}</td>
              <td class="px-4 py-3 max-w-[160px] truncate">{{ a.subject }}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <div class="w-7 h-7 rounded-full bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 text-xs flex items-center justify-center font-bold">
                    {{ a.agent_name?.charAt(0) || '?' }}
                  </div>
                  <span>{{ a.agent_name }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-xs text-slate-500">{{ a.assigned_by || '-' }}</td>
              <td class="px-4 py-3 text-xs text-slate-500 max-w-[140px] truncate">{{ a.reason || '-' }}</td>
              <td class="px-4 py-3 text-xs text-slate-400">{{ a.assigned_at?.substring(0,16).replace('T',' ') }}</td>
              <td class="px-4 py-3">
                <span :class="a.is_current ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-slate-100 text-slate-500 dark:bg-slate-700'"
                  class="text-xs px-2 py-0.5 rounded-full font-medium">
                  {{ a.is_current ? 'Current' : 'History' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Assign Modal -->
    <div v-if="showAssignModal" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="w-full max-w-md rounded-2xl border shadow-xl">
        <div class="flex items-center justify-between p-6 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold">Assign Ticket</h2>
          <button @click="showAssignModal = false" class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg">
            <span class="text-xl leading-none">&times;</span>
          </button>
        </div>
        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1">Ticket *</label>
            <select v-model="assignForm.ticket_id" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
              <option value="">Select ticket...</option>
              <option v-for="t in tickets" :key="t.id" :value="t.id">{{ t.ticket_number }} — {{ t.subject }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Assign To</label>
            <select v-model="assignForm.agent_id" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
              <option value="">Unassigned</option>
              <option v-for="a in agents" :key="a.id" :value="a.id">{{ a.name }} ({{ a.open_tickets }}/{{ a.max_tickets }})</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Assigned By</label>
            <input v-model="assignForm.assigned_by" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Reason</label>
            <textarea v-model="assignForm.reason" rows="2" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none resize-none" />
          </div>
        </div>
        <div class="flex justify-end gap-3 p-6 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showAssignModal = false" :class="app.darkMode ? 'bg-slate-700 text-slate-200 hover:bg-slate-600' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
            class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
          <button @click="assign" :disabled="saving" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors flex items-center gap-2">
            <RefreshCw v-if="saving" class="w-4 h-4 animate-spin" />
            Assign
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
