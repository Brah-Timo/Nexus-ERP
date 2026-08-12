<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Milestones</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Track key project milestones and delivery targets</p>
      </div>
      <button @click="openCreate" class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm">
        <Plus class="w-4 h-4" />
        New Milestone
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center flex-shrink-0">
          <Flag class="w-6 h-6 text-indigo-600 dark:text-indigo-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Total</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ milestones.length }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-emerald-100 dark:bg-emerald-900/40 flex items-center justify-center flex-shrink-0">
          <CheckCircle class="w-6 h-6 text-emerald-600 dark:text-emerald-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Completed</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ milestones.filter(m => m.status === 'completed').length }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-amber-100 dark:bg-amber-900/40 flex items-center justify-center flex-shrink-0">
          <Clock class="w-6 h-6 text-amber-600 dark:text-amber-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">In Progress</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ milestones.filter(m => m.status === 'in_progress').length }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-red-100 dark:bg-red-900/40 flex items-center justify-center flex-shrink-0">
          <AlertTriangle class="w-6 h-6 text-red-600 dark:text-red-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Overdue</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ overdueCount }}</p>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
      <div class="flex flex-wrap gap-3">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input v-model="search" placeholder="Search milestones..." class="w-full pl-9 pr-4 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white placeholder-gray-400" />
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
            <option value="pending">Pending</option>
            <option value="in_progress">In Progress</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
      </div>
    </div>

    <!-- Milestones List -->
    <div v-if="loading" class="flex items-center justify-center py-16">
      <Loader2 class="w-10 h-10 text-indigo-500 animate-spin" />
    </div>

    <div v-else-if="filteredMilestones.length === 0" class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 flex flex-col items-center justify-center py-16 text-gray-400">
      <Flag class="w-12 h-12 mb-3 opacity-30" />
      <p class="text-sm">No milestones found</p>
    </div>

    <div v-else class="space-y-3">
      <div v-for="ms in filteredMilestones" :key="ms.id"
        class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden hover:border-indigo-200 dark:hover:border-indigo-700 transition-all group">

        <div class="p-5">
          <div class="flex items-start gap-4">
            <!-- Status icon -->
            <div :class="statusIconBg(ms.status)" class="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0 mt-0.5">
              <component :is="statusIcon(ms.status)" :class="statusIconColor(ms.status)" class="w-5 h-5" />
            </div>

            <div class="flex-1 min-w-0">
              <div class="flex items-start justify-between gap-2 mb-2">
                <div>
                  <h3 class="font-semibold text-gray-900 dark:text-white">{{ ms.title }}</h3>
                  <p class="text-xs text-gray-400 mt-0.5">{{ ms.project_name }}</p>
                </div>
                <div class="flex items-center gap-2 flex-shrink-0">
                  <span :class="msBadge(ms.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold">{{ ms.status?.replace('_', ' ') }}</span>
                  <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button @click="openEdit(ms)" class="p-1.5 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg text-gray-400 hover:text-indigo-600 transition-colors">
                      <Pencil class="w-3.5 h-3.5" />
                    </button>
                    <button @click="confirmDelete(ms)" class="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg text-gray-400 hover:text-red-600 transition-colors">
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
              </div>

              <!-- Progress bar -->
              <div class="mb-3">
                <div class="flex justify-between text-xs text-gray-400 mb-1">
                  <span>Progress</span>
                  <span class="font-semibold">{{ ms.progress_pct || 0 }}%</span>
                </div>
                <div class="h-2 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
                  <div :style="{ width: `${ms.progress_pct || 0}%` }" :class="ms.status === 'completed' ? 'bg-emerald-500' : 'bg-indigo-500'" class="h-full rounded-full transition-all" />
                </div>
              </div>

              <!-- Meta row -->
              <div class="flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
                <div class="flex items-center gap-1">
                  <CalendarDays class="w-3.5 h-3.5" />
                  <span :class="isOverdue(ms.due_date) && ms.status !== 'completed' ? 'text-red-500 font-medium' : ''">Due: {{ fmtDate(ms.due_date) }}</span>
                </div>
                <div v-if="ms.completed_at" class="flex items-center gap-1 text-emerald-600 dark:text-emerald-400">
                  <CheckCircle class="w-3.5 h-3.5" />
                  <span>Completed: {{ fmtDate(ms.completed_at) }}</span>
                </div>
                <div v-if="ms.owner_name" class="flex items-center gap-1">
                  <User class="w-3.5 h-3.5" />
                  <span>{{ ms.owner_name }}</span>
                </div>
              </div>

              <p v-if="ms.description" class="text-xs text-gray-400 mt-2 line-clamp-2">{{ ms.description }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeModal" />
          <div class="relative w-full max-w-lg bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden">
            <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-indigo-600 to-violet-600">
              <div class="flex items-center gap-3 text-white">
                <Flag class="w-5 h-5" />
                <h2 class="font-semibold text-lg">{{ editingMs ? 'Edit Milestone' : 'New Milestone' }}</h2>
              </div>
              <button @click="closeModal" class="p-1 hover:bg-white/20 rounded-lg text-white transition-colors">
                <X class="w-5 h-5" />
              </button>
            </div>
            <form @submit.prevent="saveMilestone" class="p-6 space-y-4">
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Title *</label>
                <input v-model="form.title" required placeholder="Milestone title" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
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
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Status</label>
                  <div class="relative">
                    <select v-model="form.status" class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="pending">Pending</option>
                      <option value="in_progress">In Progress</option>
                      <option value="completed">Completed</option>
                      <option value="cancelled">Cancelled</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Due Date *</label>
                  <input type="date" v-model="form.due_date" required class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Progress %</label>
                  <input type="number" v-model.number="form.progress_pct" min="0" max="100" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Owner</label>
                  <div class="relative">
                    <select v-model="form.owner_id" class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="">— None —</option>
                      <option v-for="e in employees" :key="e.id" :value="e.id">{{ e.full_name }}</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
              </div>
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Description</label>
                <textarea v-model="form.description" rows="3" placeholder="Milestone description..." class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white resize-none" />
              </div>
              <div class="flex gap-3">
                <button type="button" @click="closeModal" class="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
                <button type="submit" :disabled="saving" class="flex-1 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center justify-center gap-2">
                  <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                  {{ saving ? 'Saving...' : (editingMs ? 'Update' : 'Create') }}
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
  Plus, Search, Loader2, Flag, CheckCircle, Clock, AlertTriangle, CalendarDays, User,
  Pencil, Trash2, X, ChevronDown
} from 'lucide-vue-next'
import { projectsAPI, hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

const loading = ref(true)
const saving = ref(false)
const milestones = ref<any[]>([])
const projects = ref<any[]>([])
const employees = ref<any[]>([])
const search = ref('')
const filterProject = ref('')
const filterStatus = ref('')
const showModal = ref(false)
const editingMs = ref<any>(null)

const form = ref({
  title: '', description: '', project_id: '', status: 'pending',
  due_date: '', progress_pct: 0, owner_id: ''
})

const overdueCount = computed(() =>
  milestones.value.filter(m => m.status !== 'completed' && m.status !== 'cancelled' && isOverdue(m.due_date)).length
)

const filteredMilestones = computed(() => {
  let list = [...milestones.value]
  if (search.value) { const q = search.value.toLowerCase(); list = list.filter(m => m.title.toLowerCase().includes(q)) }
  if (filterProject.value) list = list.filter(m => m.project_id === filterProject.value)
  if (filterStatus.value) list = list.filter(m => m.status === filterStatus.value)
  return list.sort((a, b) => new Date(a.due_date).getTime() - new Date(b.due_date).getTime())
})

function fmtDate(d?: string) { if (!d) return '—'; return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' }) }
function isOverdue(d?: string) { if (!d) return false; return new Date(d) < new Date() }

function statusIcon(s?: string) {
  switch (s) {
    case 'completed': return CheckCircle
    case 'in_progress': return Clock
    case 'cancelled': return AlertTriangle
    default: return Flag
  }
}
function statusIconBg(s?: string) {
  switch (s) {
    case 'completed': return 'bg-emerald-100 dark:bg-emerald-900/40'
    case 'in_progress': return 'bg-blue-100 dark:bg-blue-900/40'
    case 'cancelled': return 'bg-gray-100 dark:bg-gray-800'
    default: return 'bg-indigo-100 dark:bg-indigo-900/40'
  }
}
function statusIconColor(s?: string) {
  switch (s) {
    case 'completed': return 'text-emerald-600 dark:text-emerald-400'
    case 'in_progress': return 'text-blue-600 dark:text-blue-400'
    case 'cancelled': return 'text-gray-400'
    default: return 'text-indigo-600 dark:text-indigo-400'
  }
}
function msBadge(s?: string) {
  switch (s) {
    case 'completed': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'in_progress': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    case 'cancelled': return 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'
    default: return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
  }
}

async function load() {
  loading.value = true
  try {
    // Load milestones from all projects
    const pRes = await projectsAPI.getProjects()
    projects.value = pRes.data || []
    const allMs: any[] = []
    await Promise.all((pRes.data || []).slice(0, 20).map(async (proj: any) => {
      try {
        const msRes = await projectsAPI.getMilestones(proj.id)
        ;(msRes.data || []).forEach((m: any) => { m.project_name = proj.name; allMs.push(m) })
      } catch { /* skip */ }
    }))
    milestones.value = allMs
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load milestones', 'error')
  } finally {
    loading.value = false
  }
}

async function loadEmployees() {
  try { employees.value = (await hrAPI.getEmployees()).data || [] } catch { /* ignore */ }
}

function openCreate() {
  editingMs.value = null
  form.value = { title: '', description: '', project_id: filterProject.value || '', status: 'pending', due_date: '', progress_pct: 0, owner_id: '' }
  showModal.value = true
}
function openEdit(ms: any) {
  editingMs.value = ms
  form.value = { title: ms.title, description: ms.description || '', project_id: ms.project_id, status: ms.status, due_date: ms.due_date?.slice(0, 10) || '', progress_pct: ms.progress_pct || 0, owner_id: ms.owner_id || '' }
  showModal.value = true
}
function closeModal() { showModal.value = false; editingMs.value = null }

async function saveMilestone() {
  saving.value = true
  try {
    const payload = { ...form.value, owner_id: form.value.owner_id || null }
    if (editingMs.value) {
      await projectsAPI.updateMilestone(editingMs.value.id, payload)
      store.addToast('Milestone updated', 'success')
    } else {
      await projectsAPI.createMilestone(form.value.project_id, payload)
      store.addToast('Milestone created', 'success')
    }
    closeModal()
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

async function confirmDelete(ms: any) {
  if (!confirm(`Delete milestone "${ms.title}"?`)) return
  try {
    await projectsAPI.deleteMilestone(ms.id)
    store.addToast('Milestone deleted', 'success')
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Delete failed', 'error')
  }
}

onMounted(() => { load(); loadEmployees() })
</script>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
