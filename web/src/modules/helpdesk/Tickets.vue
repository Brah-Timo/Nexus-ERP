<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { helpdeskAPI } from '@/api/client'
import {
  Ticket, Plus, Search, Filter, RefreshCw, X, Edit, Trash2,
  MessageSquare, CheckCircle2, ChevronDown, User, Tag, Clock,
  AlertTriangle, Eye, Send
} from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(true)
const saving = ref(false)
const error = ref('')

const tickets = ref<any[]>([])
const categories = ref<any[]>([])
const agents = ref<any[]>([])
const slaPolicies = ref<any[]>([])

const showForm = ref(false)
const showDetail = ref(false)
const editId = ref<string | null>(null)
const selectedTicket = ref<any>(null)
const ticketComments = ref<any[]>([])
const newComment = ref({ body: '', is_internal: false, author_name: '' })
const addingComment = ref(false)

const filters = ref({ status: '', priority: '', category_id: '', agent_id: '', search: '' })

const form = ref({
  subject: '', description: '', status: 'open', priority: 'medium', source: 'portal',
  category_id: '', sla_policy_id: '', assigned_agent_id: '',
  requester_name: '', requester_email: '', requester_phone: '', company_name: ''
})

const statusOptions = ['open','pending','in_progress','resolved','closed','cancelled']
const priorityOptions = ['low','medium','high','critical']
const sourceOptions = ['email','phone','chat','portal','api','internal']

const statusColor: Record<string, string> = {
  open: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  pending: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300',
  in_progress: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
  resolved: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
  closed: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
  cancelled: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
}
const priorityColor: Record<string, string> = {
  low: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
  medium: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  high: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300',
  critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
}

async function loadAll() {
  loading.value = true
  error.value = ''
  try {
    const params: Record<string, string> = {}
    if (filters.value.status) params.status = filters.value.status
    if (filters.value.priority) params.priority = filters.value.priority
    if (filters.value.category_id) params.category_id = filters.value.category_id
    if (filters.value.agent_id) params.agent_id = filters.value.agent_id
    if (filters.value.search) params.search = filters.value.search
    const [t, c, a, s] = await Promise.all([
      helpdeskAPI.listTickets(params),
      helpdeskAPI.listCategories(),
      helpdeskAPI.listAgents(),
      helpdeskAPI.listSLAPolicies()
    ])
    tickets.value = t.data.tickets || []
    categories.value = c.data.categories || []
    agents.value = a.data.agents || []
    slaPolicies.value = s.data.policies || []
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to load tickets'
  } finally {
    loading.value = false
  }
}

async function openDetail(ticket: any) {
  try {
    const res = await helpdeskAPI.getTicket(ticket.id)
    selectedTicket.value = res.data.ticket
    ticketComments.value = res.data.comments || []
    showDetail.value = true
  } catch {}
}

function openCreate() {
  editId.value = null
  form.value = {
    subject: '', description: '', status: 'open', priority: 'medium', source: 'portal',
    category_id: '', sla_policy_id: '', assigned_agent_id: '',
    requester_name: '', requester_email: '', requester_phone: '', company_name: ''
  }
  showForm.value = true
}

function openEdit(t: any) {
  editId.value = t.id
  form.value = { ...t }
  showForm.value = true
}

async function save() {
  if (!form.value.subject.trim()) { error.value = 'Subject is required'; return }
  saving.value = true
  error.value = ''
  try {
    if (editId.value) {
      await helpdeskAPI.updateTicket(editId.value, form.value)
    } else {
      await helpdeskAPI.createTicket(form.value)
    }
    showForm.value = false
    await loadAll()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to save'
  } finally {
    saving.value = false
  }
}

async function deleteTicket(id: string) {
  if (!confirm('Delete this ticket?')) return
  try {
    await helpdeskAPI.deleteTicket(id)
    await loadAll()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to delete'
  }
}

async function updateStatus(id: string, status: string) {
  try {
    await helpdeskAPI.updateTicketStatus(id, { status })
    await loadAll()
    if (selectedTicket.value?.id === id) selectedTicket.value.status = status
  } catch {}
}

async function submitComment() {
  if (!newComment.value.body.trim()) return
  addingComment.value = true
  try {
    await helpdeskAPI.addComment(selectedTicket.value.id, newComment.value)
    newComment.value.body = ''
    const res = await helpdeskAPI.getTicket(selectedTicket.value.id)
    ticketComments.value = res.data.comments || []
  } catch {} finally {
    addingComment.value = false
  }
}

onMounted(loadAll)
</script>

<template>
  <div :class="app.darkMode ? 'bg-slate-900 text-slate-100' : 'bg-slate-50 text-slate-900'" class="min-h-screen p-6">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <Ticket class="w-7 h-7 text-indigo-500" />
        <div>
          <h1 class="text-2xl font-bold">Tickets</h1>
          <p class="text-sm text-slate-500">{{ tickets.length }} tickets</p>
        </div>
      </div>
      <div class="flex gap-2">
        <button @click="loadAll" :class="app.darkMode ? 'bg-slate-700 text-slate-200' : 'bg-white text-slate-700'"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-600 hover:bg-slate-100 dark:hover:bg-slate-600 transition-colors">
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
        </button>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" /> New Ticket
        </button>
      </div>
    </div>

    <div v-if="error" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg text-sm">{{ error }}</div>

    <!-- Filters -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border p-4 mb-4 shadow-sm">
      <div class="grid grid-cols-2 md:grid-cols-5 gap-3">
        <div class="relative col-span-2 md:col-span-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="filters.search" @input="loadAll" placeholder="Search tickets..."
            :class="app.darkMode ? 'bg-slate-700 border-slate-600 text-slate-100' : 'bg-white border-slate-200 text-slate-900'"
            class="w-full pl-9 pr-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
        </div>
        <select v-model="filters.status" @change="loadAll"
          :class="app.darkMode ? 'bg-slate-700 border-slate-600 text-slate-100' : 'bg-white border-slate-200 text-slate-900'"
          class="px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
          <option value="">All Status</option>
          <option v-for="s in statusOptions" :key="s" :value="s">{{ s.replace('_',' ') }}</option>
        </select>
        <select v-model="filters.priority" @change="loadAll"
          :class="app.darkMode ? 'bg-slate-700 border-slate-600 text-slate-100' : 'bg-white border-slate-200 text-slate-900'"
          class="px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
          <option value="">All Priority</option>
          <option v-for="p in priorityOptions" :key="p" :value="p">{{ p }}</option>
        </select>
        <select v-model="filters.category_id" @change="loadAll"
          :class="app.darkMode ? 'bg-slate-700 border-slate-600 text-slate-100' : 'bg-white border-slate-200 text-slate-900'"
          class="px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
          <option value="">All Categories</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <select v-model="filters.agent_id" @change="loadAll"
          :class="app.darkMode ? 'bg-slate-700 border-slate-600 text-slate-100' : 'bg-white border-slate-200 text-slate-900'"
          class="px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
          <option value="">All Agents</option>
          <option v-for="a in agents" :key="a.id" :value="a.id">{{ a.name }}</option>
        </select>
      </div>
    </div>

    <!-- Tickets Table -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border shadow-sm overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center py-16">
        <RefreshCw class="w-6 h-6 animate-spin text-indigo-500" />
      </div>
      <div v-else-if="!tickets.length" class="text-center py-16">
        <Ticket class="w-12 h-12 text-slate-300 mx-auto mb-3" />
        <p class="text-slate-400">No tickets found</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'bg-slate-700/50 text-slate-400' : 'bg-slate-50 text-slate-500'"
              class="text-xs uppercase tracking-wide">
              <th class="text-left px-4 py-3 font-medium">Ticket #</th>
              <th class="text-left px-4 py-3 font-medium">Subject</th>
              <th class="text-left px-4 py-3 font-medium">Status</th>
              <th class="text-left px-4 py-3 font-medium">Priority</th>
              <th class="text-left px-4 py-3 font-medium">Category</th>
              <th class="text-left px-4 py-3 font-medium">Requester</th>
              <th class="text-left px-4 py-3 font-medium">Agent</th>
              <th class="text-left px-4 py-3 font-medium">Created</th>
              <th class="text-left px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in tickets" :key="t.id"
              :class="app.darkMode ? 'border-slate-700 hover:bg-slate-700/50' : 'border-slate-100 hover:bg-slate-50'"
              class="border-b transition-colors">
              <td class="px-4 py-3 font-mono text-xs text-indigo-600 font-medium">{{ t.ticket_number }}</td>
              <td class="px-4 py-3 max-w-[200px]">
                <div class="truncate font-medium">{{ t.subject }}</div>
                <div class="text-xs text-slate-400 capitalize">{{ t.source }}</div>
              </td>
              <td class="px-4 py-3">
                <select :value="t.status" @change="updateStatus(t.id, ($event.target as HTMLSelectElement).value)"
                  :class="statusColor[t.status] || ''"
                  class="text-xs px-2 py-1 rounded-full font-medium border-0 focus:ring-1 focus:ring-indigo-500 cursor-pointer bg-transparent capitalize">
                  <option v-for="s in statusOptions" :key="s" :value="s">{{ s.replace('_',' ') }}</option>
                </select>
              </td>
              <td class="px-4 py-3">
                <span :class="priorityColor[t.priority] || ''" class="text-xs px-2 py-0.5 rounded-full font-medium capitalize">
                  {{ t.priority }}
                </span>
              </td>
              <td class="px-4 py-3 text-xs text-slate-500">{{ t.category || '-' }}</td>
              <td class="px-4 py-3 text-xs">
                <div>{{ t.requester_name || '-' }}</div>
                <div class="text-slate-400">{{ t.requester_email }}</div>
              </td>
              <td class="px-4 py-3 text-xs">{{ t.assigned_agent }}</td>
              <td class="px-4 py-3 text-xs text-slate-400">{{ t.created_at?.substring(0,10) }}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1">
                  <button @click="openDetail(t)" class="p-1.5 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded transition-colors" title="View">
                    <Eye class="w-4 h-4" />
                  </button>
                  <button @click="openEdit(t)" class="p-1.5 text-slate-400 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors" title="Edit">
                    <Edit class="w-4 h-4" />
                  </button>
                  <button @click="deleteTicket(t.id)" class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors" title="Delete">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showForm" class="fixed inset-0 bg-black/50 z-50 flex items-start justify-center p-4 overflow-y-auto">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="w-full max-w-2xl rounded-2xl border shadow-xl mt-4">
        <div class="flex items-center justify-between p-6 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold">{{ editId ? 'Edit Ticket' : 'New Ticket' }}</h2>
          <button @click="showForm = false" class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-6 space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div class="md:col-span-2">
              <label class="block text-sm font-medium mb-1">Subject *</label>
              <input v-model="form.subject" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-sm font-medium mb-1">Description</label>
              <textarea v-model="form.description" rows="3" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none resize-none" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Status</label>
              <select v-model="form.status" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option v-for="s in statusOptions" :key="s" :value="s">{{ s.replace('_',' ') }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Priority</label>
              <select v-model="form.priority" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option v-for="p in priorityOptions" :key="p" :value="p">{{ p }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Source</label>
              <select v-model="form.source" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option v-for="s in sourceOptions" :key="s" :value="s">{{ s }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Category</label>
              <select v-model="form.category_id" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option value="">Select Category</option>
                <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Assigned Agent</label>
              <select v-model="form.assigned_agent_id" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option value="">Unassigned</option>
                <option v-for="a in agents" :key="a.id" :value="a.id">{{ a.name }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">SLA Policy</label>
              <select v-model="form.sla_policy_id" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option value="">No SLA</option>
                <option v-for="s in slaPolicies" :key="s.id" :value="s.id">{{ s.name }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Requester Name</label>
              <input v-model="form.requester_name" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Requester Email</label>
              <input v-model="form.requester_email" type="email" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Requester Phone</label>
              <input v-model="form.requester_phone" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Company</label>
              <input v-model="form.company_name" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
          </div>
        </div>
        <div class="flex justify-end gap-3 p-6 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showForm = false" :class="app.darkMode ? 'bg-slate-700 text-slate-200 hover:bg-slate-600' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
            class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
          <button @click="save" :disabled="saving" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors flex items-center gap-2">
            <RefreshCw v-if="saving" class="w-4 h-4 animate-spin" />
            {{ editId ? 'Update' : 'Create' }} Ticket
          </button>
        </div>
      </div>
    </div>

    <!-- Detail Panel -->
    <div v-if="showDetail && selectedTicket" class="fixed inset-0 bg-black/50 z-50 flex justify-end">
      <div :class="app.darkMode ? 'bg-slate-800' : 'bg-white'" class="w-full max-w-xl h-full overflow-y-auto shadow-2xl">
        <div class="flex items-center justify-between p-5 border-b sticky top-0 z-10" :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'">
          <div>
            <span class="font-mono text-xs text-indigo-600 font-bold">{{ selectedTicket.ticket_number }}</span>
            <h2 class="font-semibold">{{ selectedTicket.subject }}</h2>
          </div>
          <button @click="showDetail = false" class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-5 space-y-4">
          <!-- Status/Priority badges -->
          <div class="flex gap-2 flex-wrap">
            <span :class="statusColor[selectedTicket.status] || ''" class="text-xs px-3 py-1 rounded-full font-medium capitalize">{{ selectedTicket.status?.replace('_',' ') }}</span>
            <span :class="priorityColor[selectedTicket.priority] || ''" class="text-xs px-3 py-1 rounded-full font-medium capitalize">{{ selectedTicket.priority }}</span>
            <span class="text-xs px-3 py-1 rounded-full bg-slate-100 dark:bg-slate-700 font-medium capitalize">{{ selectedTicket.source }}</span>
          </div>
          <!-- Description -->
          <div :class="app.darkMode ? 'bg-slate-700' : 'bg-slate-50'" class="rounded-lg p-4 text-sm">
            {{ selectedTicket.description || 'No description' }}
          </div>
          <!-- Meta info -->
          <div class="grid grid-cols-2 gap-3 text-sm">
            <div>
              <div class="text-xs text-slate-400 mb-0.5">Requester</div>
              <div class="font-medium">{{ selectedTicket.requester_name || '-' }}</div>
              <div class="text-xs text-slate-400">{{ selectedTicket.requester_email }}</div>
            </div>
            <div>
              <div class="text-xs text-slate-400 mb-0.5">Assigned Agent</div>
              <div class="font-medium">{{ selectedTicket.assigned_agent || 'Unassigned' }}</div>
            </div>
            <div>
              <div class="text-xs text-slate-400 mb-0.5">Category</div>
              <div class="font-medium">{{ selectedTicket.category || '-' }}</div>
            </div>
            <div>
              <div class="text-xs text-slate-400 mb-0.5">SLA Policy</div>
              <div class="font-medium">{{ selectedTicket.sla_policy || '-' }}</div>
            </div>
            <div>
              <div class="text-xs text-slate-400 mb-0.5">Created</div>
              <div class="font-medium">{{ selectedTicket.created_at?.substring(0,16).replace('T',' ') }}</div>
            </div>
            <div>
              <div class="text-xs text-slate-400 mb-0.5">Due Date</div>
              <div class="font-medium" :class="selectedTicket.due_date && new Date(selectedTicket.due_date) < new Date() ? 'text-red-500' : ''">
                {{ selectedTicket.due_date?.substring(0,16).replace('T',' ') || '-' }}
              </div>
            </div>
          </div>
          <!-- Comments -->
          <div class="border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
            <h3 class="font-semibold py-3 flex items-center gap-2">
              <MessageSquare class="w-4 h-4" /> Comments ({{ ticketComments.length }})
            </h3>
            <div class="space-y-3 mb-4">
              <div v-for="cm in ticketComments" :key="cm.id"
                :class="[cm.is_internal ? (app.darkMode ? 'bg-yellow-900/20 border-yellow-700' : 'bg-yellow-50 border-yellow-200') : (app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-slate-50 border-slate-200')]"
                class="rounded-lg p-3 border text-sm">
                <div class="flex items-center justify-between mb-1">
                  <span class="font-medium text-xs">{{ cm.author_name || 'System' }}</span>
                  <div class="flex items-center gap-2">
                    <span v-if="cm.is_internal" class="text-xs text-yellow-600 dark:text-yellow-400">Internal</span>
                    <span class="text-xs text-slate-400">{{ cm.created_at?.substring(0,16).replace('T',' ') }}</span>
                  </div>
                </div>
                <p class="text-slate-700 dark:text-slate-300">{{ cm.body }}</p>
              </div>
              <p v-if="!ticketComments.length" class="text-sm text-slate-400 text-center py-2">No comments yet</p>
            </div>
            <!-- Add comment -->
            <div :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'" class="border-t pt-3">
              <textarea v-model="newComment.body" placeholder="Add a comment..." rows="3"
                :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none resize-none mb-2" />
              <div class="flex items-center justify-between">
                <label class="flex items-center gap-2 text-sm cursor-pointer">
                  <input type="checkbox" v-model="newComment.is_internal" class="rounded" />
                  Internal note
                </label>
                <button @click="submitComment" :disabled="addingComment || !newComment.body.trim()"
                  class="flex items-center gap-2 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded-lg text-xs font-medium transition-colors">
                  <Send class="w-3 h-3" />
                  {{ addingComment ? 'Sending...' : 'Send' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
