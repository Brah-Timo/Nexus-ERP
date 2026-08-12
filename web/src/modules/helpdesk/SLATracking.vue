<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { helpdeskAPI } from '@/api/client'
import { Timer, Plus, Edit, Trash2, RefreshCw, X, AlertTriangle, CheckCircle2, Clock, ShieldCheck } from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const activeTab = ref<'tracking'|'policies'>('tracking')

const summaries = ref<any[]>([])
const overdueTickets = ref<any[]>([])
const policies = ref<any[]>([])
const showForm = ref(false)
const editId = ref<string | null>(null)

const form = ref({
  name: '', description: '', priority: 'medium',
  first_response_hours: 4, resolution_hours: 24,
  business_hours_only: true, is_active: true
})

const priorityColor: Record<string, string> = {
  low: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
  medium: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  high: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300',
  critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [tracking, pol] = await Promise.all([
      helpdeskAPI.getSLATracking(),
      helpdeskAPI.listSLAPolicies()
    ])
    summaries.value = tracking.data.summaries || []
    overdueTickets.value = tracking.data.overdue_tickets || []
    policies.value = pol.data.policies || []
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to load'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editId.value = null
  form.value = { name: '', description: '', priority: 'medium', first_response_hours: 4, resolution_hours: 24, business_hours_only: true, is_active: true }
  showForm.value = true
}

function openEdit(p: any) {
  editId.value = p.id
  form.value = { ...p }
  showForm.value = true
}

async function save() {
  if (!form.value.name.trim()) { error.value = 'Name is required'; return }
  saving.value = true
  error.value = ''
  try {
    if (editId.value) {
      await helpdeskAPI.updateSLAPolicy(editId.value, form.value)
    } else {
      await helpdeskAPI.createSLAPolicy(form.value)
    }
    showForm.value = false
    await load()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to save'
  } finally {
    saving.value = false
  }
}

async function deletePolicy(id: string) {
  if (!confirm('Delete this SLA policy?')) return
  try {
    await helpdeskAPI.deleteSLAPolicy(id)
    await load()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to delete'
  }
}

onMounted(load)

const overallCompliance = () => {
  const total = summaries.value.reduce((s, p) => s + p.total_tickets, 0)
  const within = summaries.value.reduce((s, p) => s + p.within_sla, 0)
  return total > 0 ? Math.round(within / total * 100) : 100
}
</script>

<template>
  <div :class="app.darkMode ? 'bg-slate-900 text-slate-100' : 'bg-slate-50 text-slate-900'" class="min-h-screen p-6">
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <Timer class="w-7 h-7 text-indigo-500" />
        <div>
          <h1 class="text-2xl font-bold">SLA Tracking</h1>
          <p class="text-sm text-slate-500">Service Level Agreement compliance</p>
        </div>
      </div>
      <div class="flex gap-2">
        <button @click="load" :class="app.darkMode ? 'bg-slate-700 text-slate-200' : 'bg-white text-slate-700'"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-600 hover:bg-slate-100 dark:hover:bg-slate-600 transition-colors">
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
        </button>
        <button @click="activeTab='policies'; openCreate()" class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" /> New Policy
        </button>
      </div>
    </div>

    <div v-if="error" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg text-sm">{{ error }}</div>

    <!-- Overall compliance banner -->
    <div :class="overallCompliance() >= 90 ? 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-700' : overallCompliance() >= 70 ? 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-700' : 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-700'"
      class="rounded-xl border p-5 mb-6 flex items-center justify-between shadow-sm">
      <div class="flex items-center gap-3">
        <ShieldCheck class="w-8 h-8" :class="overallCompliance() >= 90 ? 'text-green-600' : overallCompliance() >= 70 ? 'text-yellow-600' : 'text-red-600'" />
        <div>
          <div class="text-sm font-medium text-slate-600 dark:text-slate-300">Overall SLA Compliance</div>
          <div class="text-3xl font-bold" :class="overallCompliance() >= 90 ? 'text-green-600' : overallCompliance() >= 70 ? 'text-yellow-600' : 'text-red-600'">
            {{ overallCompliance() }}%
          </div>
        </div>
      </div>
      <div class="text-right">
        <div class="text-2xl font-bold text-red-600">{{ overdueTickets.length }}</div>
        <div class="text-sm text-slate-500">Overdue tickets</div>
      </div>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 mb-4" :class="app.darkMode ? 'bg-slate-800' : 'bg-white'" style="display:inline-flex; border-radius:0.75rem; padding:0.25rem; border:1px solid; border-color:var(--tw-border-color)">
      <button @click="activeTab='tracking'" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        :class="activeTab==='tracking' ? 'bg-indigo-600 text-white' : (app.darkMode ? 'text-slate-400 hover:text-slate-200' : 'text-slate-600 hover:text-slate-900')">
        SLA Tracking
      </button>
      <button @click="activeTab='policies'" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors"
        :class="activeTab==='policies' ? 'bg-indigo-600 text-white' : (app.darkMode ? 'text-slate-400 hover:text-slate-200' : 'text-slate-600 hover:text-slate-900')">
        Policies
      </button>
    </div>

    <!-- Tracking Tab -->
    <div v-if="activeTab === 'tracking'">
      <!-- Policy summaries -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        <div v-for="s in summaries" :key="s.policy_id"
          :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
          class="rounded-xl border shadow-sm p-5">
          <div class="flex items-center justify-between mb-3">
            <div>
              <div class="font-semibold">{{ s.policy_name }}</div>
              <span :class="priorityColor[s.priority] || ''" class="text-xs px-2 py-0.5 rounded-full font-medium capitalize">{{ s.priority }}</span>
            </div>
            <div class="text-right">
              <div class="text-xl font-bold" :class="s.compliance_rate >= 90 ? 'text-green-600' : s.compliance_rate >= 70 ? 'text-yellow-600' : 'text-red-600'">
                {{ s.compliance_rate.toFixed(1) }}%
              </div>
              <div class="text-xs text-slate-400">compliance</div>
            </div>
          </div>
          <!-- Compliance bar -->
          <div class="mb-3 bg-slate-100 dark:bg-slate-700 rounded-full h-2">
            <div :style="`width:${s.compliance_rate}%`"
              :class="s.compliance_rate >= 90 ? 'bg-green-500' : s.compliance_rate >= 70 ? 'bg-yellow-500' : 'bg-red-500'"
              class="h-2 rounded-full transition-all"></div>
          </div>
          <div class="grid grid-cols-4 gap-2 text-center text-xs">
            <div>
              <div class="font-bold text-slate-700 dark:text-slate-300">{{ s.total_tickets }}</div>
              <div class="text-slate-400">Total</div>
            </div>
            <div>
              <div class="font-bold text-green-600">{{ s.within_sla }}</div>
              <div class="text-slate-400">Within SLA</div>
            </div>
            <div>
              <div class="font-bold text-red-600">{{ s.breached }}</div>
              <div class="text-slate-400">Breached</div>
            </div>
            <div>
              <div class="font-bold text-indigo-600">{{ s.avg_resolution_hrs.toFixed(1) }}h</div>
              <div class="text-slate-400">Avg Res.</div>
            </div>
          </div>
          <div class="mt-3 text-xs text-slate-400 flex gap-4">
            <span>First response target: {{ s.first_response_hours_target }}h</span>
            <span>Resolution target: {{ s.resolution_hours_target }}h</span>
          </div>
        </div>
        <div v-if="!summaries.length" class="col-span-2 text-center py-8 text-slate-400">
          No SLA policies configured
        </div>
      </div>

      <!-- Overdue tickets table -->
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border shadow-sm overflow-hidden">
        <div class="px-5 py-4 border-b flex items-center gap-2" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <AlertTriangle class="w-4 h-4 text-red-500" />
          <h3 class="font-semibold">Overdue Tickets</h3>
          <span class="ml-1 text-xs bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300 px-2 py-0.5 rounded-full">{{ overdueTickets.length }}</span>
        </div>
        <div v-if="!overdueTickets.length" class="text-center py-8 text-slate-400">
          <CheckCircle2 class="w-10 h-10 mx-auto mb-2 text-green-400" />
          No overdue tickets
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr :class="app.darkMode ? 'bg-slate-700/50 text-slate-400' : 'bg-slate-50 text-slate-500'" class="text-xs uppercase">
                <th class="text-left px-4 py-3 font-medium">Ticket</th>
                <th class="text-left px-4 py-3 font-medium">Subject</th>
                <th class="text-left px-4 py-3 font-medium">Priority</th>
                <th class="text-left px-4 py-3 font-medium">SLA Policy</th>
                <th class="text-left px-4 py-3 font-medium">Due Date</th>
                <th class="text-left px-4 py-3 font-medium">Hours Overdue</th>
                <th class="text-left px-4 py-3 font-medium">Agent</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="t in overdueTickets" :key="t.id"
                :class="app.darkMode ? 'border-slate-700 hover:bg-slate-700/50' : 'border-slate-100 hover:bg-red-50'"
                class="border-b">
                <td class="px-4 py-2.5 font-mono text-xs text-red-600 font-bold">{{ t.ticket_number }}</td>
                <td class="px-4 py-2.5 max-w-[150px] truncate">{{ t.subject }}</td>
                <td class="px-4 py-2.5">
                  <span :class="priorityColor[t.priority] || ''" class="text-xs px-2 py-0.5 rounded-full font-medium capitalize">{{ t.priority }}</span>
                </td>
                <td class="px-4 py-2.5 text-xs">{{ t.policy_name }}</td>
                <td class="px-4 py-2.5 text-xs text-red-500">{{ t.due_date?.substring(0,16).replace('T',' ') }}</td>
                <td class="px-4 py-2.5">
                  <span class="text-red-600 font-bold text-sm">{{ t.hours_overdue.toFixed(1) }}h</span>
                </td>
                <td class="px-4 py-2.5 text-xs">{{ t.assigned_agent }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Policies Tab -->
    <div v-if="activeTab === 'policies'">
      <div v-if="loading" class="flex items-center justify-center py-16">
        <RefreshCw class="w-6 h-6 animate-spin text-indigo-500" />
      </div>
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="p in policies" :key="p.id"
          :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
          class="rounded-xl border shadow-sm p-5">
          <div class="flex items-center justify-between mb-3">
            <div>
              <div class="font-semibold">{{ p.name }}</div>
              <span :class="priorityColor[p.priority] || ''" class="text-xs px-2 py-0.5 rounded-full font-medium capitalize mt-1 inline-block">{{ p.priority }}</span>
            </div>
            <div class="flex gap-1">
              <button @click="openEdit(p)" class="p-1.5 text-slate-400 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors">
                <Edit class="w-4 h-4" />
              </button>
              <button @click="deletePolicy(p.id)" class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors">
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>
          <p v-if="p.description" class="text-xs text-slate-400 mb-3">{{ p.description }}</p>
          <div class="grid grid-cols-2 gap-3 text-center">
            <div :class="app.darkMode ? 'bg-slate-700' : 'bg-slate-50'" class="rounded-lg p-2">
              <div class="text-lg font-bold text-indigo-600">{{ p.first_response_hours }}h</div>
              <div class="text-xs text-slate-400">First Response</div>
            </div>
            <div :class="app.darkMode ? 'bg-slate-700' : 'bg-slate-50'" class="rounded-lg p-2">
              <div class="text-lg font-bold text-indigo-600">{{ p.resolution_hours }}h</div>
              <div class="text-xs text-slate-400">Resolution</div>
            </div>
          </div>
          <div class="mt-3 flex items-center justify-between text-xs">
            <span class="text-slate-400">{{ p.ticket_count }} tickets</span>
            <div class="flex items-center gap-2">
              <span v-if="p.business_hours_only" class="bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300 px-2 py-0.5 rounded-full">Business hours</span>
              <span :class="p.is_active ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-slate-100 text-slate-500'" class="px-2 py-0.5 rounded-full">
                {{ p.is_active ? 'Active' : 'Inactive' }}
              </span>
            </div>
          </div>
        </div>
        <div v-if="!policies.length" class="col-span-3 text-center py-16">
          <Timer class="w-12 h-12 text-slate-300 mx-auto mb-3" />
          <p class="text-slate-400">No SLA policies yet</p>
        </div>
      </div>
    </div>

    <!-- Form Modal -->
    <div v-if="showForm" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="w-full max-w-lg rounded-2xl border shadow-xl">
        <div class="flex items-center justify-between p-6 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold">{{ editId ? 'Edit SLA Policy' : 'New SLA Policy' }}</h2>
          <button @click="showForm = false" class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg"><X class="w-5 h-5" /></button>
        </div>
        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1">Name *</label>
            <input v-model="form.name" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Description</label>
            <textarea v-model="form.description" rows="2" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none resize-none" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium mb-1">Priority</label>
              <select v-model="form.priority" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">First Response (hours)</label>
              <input v-model.number="form.first_response_hours" type="number" min="1" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Resolution (hours)</label>
              <input v-model.number="form.resolution_hours" type="number" min="1" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
          </div>
          <div class="flex gap-6">
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="form.business_hours_only" class="rounded" />
              <span class="text-sm">Business hours only</span>
            </label>
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="form.is_active" class="rounded" />
              <span class="text-sm">Active</span>
            </label>
          </div>
        </div>
        <div class="flex justify-end gap-3 p-6 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showForm = false" :class="app.darkMode ? 'bg-slate-700 text-slate-200 hover:bg-slate-600' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
            class="px-4 py-2 rounded-lg text-sm font-medium">Cancel</button>
          <button @click="save" :disabled="saving" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded-lg text-sm font-medium flex items-center gap-2">
            <RefreshCw v-if="saving" class="w-4 h-4 animate-spin" />
            {{ editId ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
