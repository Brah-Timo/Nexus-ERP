<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { helpdeskAPI } from '@/api/client'
import { Users, Plus, Edit, Trash2, RefreshCw, X, Search, Mail, Phone, Briefcase, Ticket } from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const agents = ref<any[]>([])
const showForm = ref(false)
const editId = ref<string | null>(null)
const searchQuery = ref('')
const filterStatus = ref('')

const statusOptions = ['active','inactive','on_leave','busy']
const form = ref({
  name: '', email: '', phone: '', department: '', specialization: '',
  status: 'active', max_tickets: 20, is_active: true
})

const statusColor: Record<string, string> = {
  active: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
  inactive: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
  on_leave: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300',
  busy: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300',
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const params: Record<string, string> = {}
    if (filterStatus.value) params.status = filterStatus.value
    if (searchQuery.value) params.search = searchQuery.value
    const res = await helpdeskAPI.listAgents(params)
    agents.value = res.data.agents || []
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to load agents'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editId.value = null
  form.value = { name: '', email: '', phone: '', department: '', specialization: '', status: 'active', max_tickets: 20, is_active: true }
  showForm.value = true
}

function openEdit(a: any) {
  editId.value = a.id
  form.value = { ...a }
  showForm.value = true
}

async function save() {
  if (!form.value.name.trim()) { error.value = 'Name is required'; return }
  if (!form.value.email.trim()) { error.value = 'Email is required'; return }
  saving.value = true
  error.value = ''
  try {
    if (editId.value) {
      await helpdeskAPI.updateAgent(editId.value, form.value)
    } else {
      await helpdeskAPI.createAgent(form.value)
    }
    showForm.value = false
    await load()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to save'
  } finally {
    saving.value = false
  }
}

async function deleteAgent(id: string) {
  if (!confirm('Delete this agent?')) return
  try {
    await helpdeskAPI.deleteAgent(id)
    await load()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to delete'
  }
}

onMounted(load)

function initials(name: string) {
  return name.split(' ').map(n => n[0]).join('').toUpperCase().substring(0, 2)
}
</script>

<template>
  <div :class="app.darkMode ? 'bg-slate-900 text-slate-100' : 'bg-slate-50 text-slate-900'" class="min-h-screen p-6">
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <Users class="w-7 h-7 text-indigo-500" />
        <div>
          <h1 class="text-2xl font-bold">Support Agents</h1>
          <p class="text-sm text-slate-500">{{ agents.length }} agents</p>
        </div>
      </div>
      <div class="flex gap-2">
        <button @click="load" :class="app.darkMode ? 'bg-slate-700 text-slate-200' : 'bg-white text-slate-700'"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-600 hover:bg-slate-100 dark:hover:bg-slate-600 transition-colors">
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
        </button>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" /> New Agent
        </button>
      </div>
    </div>

    <div v-if="error" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg text-sm">{{ error }}</div>

    <!-- Filters -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border p-4 mb-4 shadow-sm flex gap-3">
      <div class="relative flex-1">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input v-model="searchQuery" @input="load" placeholder="Search agents..."
          :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
          class="w-full pl-9 pr-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
      </div>
      <select v-model="filterStatus" @change="load"
        :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
        class="px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
        <option value="">All Status</option>
        <option v-for="s in statusOptions" :key="s" :value="s">{{ s.replace('_',' ') }}</option>
      </select>
    </div>

    <!-- Agents Grid -->
    <div v-if="loading" class="flex items-center justify-center py-16">
      <RefreshCw class="w-6 h-6 animate-spin text-indigo-500" />
    </div>
    <div v-else-if="!agents.length" class="text-center py-16">
      <Users class="w-12 h-12 text-slate-300 mx-auto mb-3" />
      <p class="text-slate-400">No agents found</p>
    </div>
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="a in agents" :key="a.id"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700 hover:border-indigo-500' : 'bg-white border-slate-200 hover:border-indigo-400'"
        class="rounded-xl border shadow-sm p-5 transition-all">
        <div class="flex items-start justify-between mb-4">
          <div class="flex items-center gap-3">
            <!-- Avatar initials -->
            <div class="w-10 h-10 rounded-full bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 flex items-center justify-center font-bold text-sm">
              {{ initials(a.name) }}
            </div>
            <div>
              <div class="font-semibold">{{ a.name }}</div>
              <span :class="statusColor[a.status] || ''" class="text-xs px-2 py-0.5 rounded-full font-medium capitalize">
                {{ a.status.replace('_',' ') }}
              </span>
            </div>
          </div>
          <div class="flex gap-1">
            <button @click="openEdit(a)" class="p-1.5 text-slate-400 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors">
              <Edit class="w-4 h-4" />
            </button>
            <button @click="deleteAgent(a.id)" class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors">
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>

        <div class="space-y-2 text-sm">
          <div class="flex items-center gap-2 text-slate-500">
            <Mail class="w-3.5 h-3.5 flex-shrink-0" />
            <span class="truncate">{{ a.email }}</span>
          </div>
          <div v-if="a.phone" class="flex items-center gap-2 text-slate-500">
            <Phone class="w-3.5 h-3.5 flex-shrink-0" />
            <span>{{ a.phone }}</span>
          </div>
          <div v-if="a.department" class="flex items-center gap-2 text-slate-500">
            <Briefcase class="w-3.5 h-3.5 flex-shrink-0" />
            <span>{{ a.department }}</span>
          </div>
        </div>

        <div class="mt-4 pt-4 border-t flex items-center justify-between" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <div class="text-center">
            <div class="text-lg font-bold text-orange-600">{{ a.open_tickets }}</div>
            <div class="text-xs text-slate-400">Open</div>
          </div>
          <div class="text-center">
            <div class="text-lg font-bold text-green-600">{{ a.resolved_tickets }}</div>
            <div class="text-xs text-slate-400">Resolved</div>
          </div>
          <div class="text-center">
            <div class="text-lg font-bold text-indigo-600">{{ a.max_tickets }}</div>
            <div class="text-xs text-slate-400">Max Cap</div>
          </div>
          <div class="text-center">
            <div class="text-sm font-bold" :class="a.open_tickets >= a.max_tickets ? 'text-red-500' : 'text-slate-500'">
              {{ Math.min(100, Math.round(a.open_tickets / Math.max(a.max_tickets,1) * 100)) }}%
            </div>
            <div class="text-xs text-slate-400">Load</div>
          </div>
        </div>
        <!-- Load bar -->
        <div class="mt-2 bg-slate-100 dark:bg-slate-700 rounded-full h-1.5">
          <div :style="`width:${Math.min(100,a.open_tickets/Math.max(a.max_tickets,1)*100)}%`"
            :class="a.open_tickets >= a.max_tickets ? 'bg-red-500' : 'bg-indigo-500'"
            class="h-1.5 rounded-full transition-all"></div>
        </div>
      </div>
    </div>

    <!-- Form Modal -->
    <div v-if="showForm" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="w-full max-w-lg rounded-2xl border shadow-xl">
        <div class="flex items-center justify-between p-6 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold">{{ editId ? 'Edit Agent' : 'New Agent' }}</h2>
          <button @click="showForm = false" class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-6 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="col-span-2">
              <label class="block text-sm font-medium mb-1">Name *</label>
              <input v-model="form.name" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Email *</label>
              <input v-model="form.email" type="email" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Phone</label>
              <input v-model="form.phone" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Department</label>
              <input v-model="form.department" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Specialization</label>
              <input v-model="form.specialization" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Status</label>
              <select v-model="form.status" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
                <option v-for="s in statusOptions" :key="s" :value="s">{{ s.replace('_',' ') }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Max Tickets</label>
              <input v-model.number="form.max_tickets" type="number" min="1" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
          </div>
          <label class="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" v-model="form.is_active" class="rounded" />
            <span class="text-sm font-medium">Active</span>
          </label>
        </div>
        <div class="flex justify-end gap-3 p-6 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showForm = false" :class="app.darkMode ? 'bg-slate-700 text-slate-200 hover:bg-slate-600' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
            class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
          <button @click="save" :disabled="saving" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors flex items-center gap-2">
            <RefreshCw v-if="saving" class="w-4 h-4 animate-spin" />
            {{ editId ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
