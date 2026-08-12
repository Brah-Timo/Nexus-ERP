<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Back -->
    <button @click="$router.push('/projects')" class="inline-flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors font-medium">
      <ArrowLeft class="w-4 h-4" />
      Back to Projects
    </button>

    <div v-if="loading" class="flex items-center justify-center py-20">
      <Loader2 class="w-10 h-10 text-indigo-500 animate-spin" />
    </div>

    <div v-else-if="!project" class="flex flex-col items-center justify-center py-20 text-gray-400">
      <FolderX class="w-16 h-16 mb-3 opacity-30" />
      <p>Project not found</p>
    </div>

    <template v-else>
      <!-- Hero -->
      <div class="relative bg-gradient-to-r from-indigo-600 via-violet-600 to-purple-700 rounded-2xl overflow-hidden shadow-xl">
        <div class="relative flex flex-col sm:flex-row items-start sm:items-end gap-4 p-6 pb-8">
          <div :class="projectColor(project.name)" class="w-16 h-16 rounded-2xl flex items-center justify-center text-white text-xl font-bold flex-shrink-0 shadow-lg border-2 border-white/30">
            {{ initials(project.name) }}
          </div>
          <div class="flex-1 min-w-0">
            <h1 class="text-2xl font-bold text-white truncate">{{ project.name }}</h1>
            <p class="text-indigo-200 text-sm mt-0.5">{{ project.code }} &bull; {{ project.customer_name || 'Internal' }}</p>
            <div class="flex flex-wrap items-center gap-2 mt-2">
              <span :class="statusBadge(project.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold">{{ project.status }}</span>
              <span class="text-white/70 text-sm">Manager: {{ project.manager_name || '—' }}</span>
            </div>
          </div>
          <button @click="openEdit" class="inline-flex items-center gap-2 px-4 py-2 bg-white/20 hover:bg-white/30 backdrop-blur border border-white/30 text-white text-sm font-medium rounded-lg transition-colors">
            <Pencil class="w-4 h-4" /> Edit
          </button>
        </div>
        <!-- Stats Bar -->
        <div class="grid grid-cols-2 sm:grid-cols-4 divide-x divide-white/20 border-t border-white/20 bg-black/10">
          <div class="px-5 py-3 text-center">
            <p class="text-xs text-indigo-200 font-medium uppercase tracking-wide">Progress</p>
            <p class="text-lg font-bold text-white">{{ project.progress_pct || 0 }}%</p>
          </div>
          <div class="px-5 py-3 text-center">
            <p class="text-xs text-indigo-200 font-medium uppercase tracking-wide">Budget</p>
            <p class="text-lg font-bold text-white">{{ fmtNum(project.budget) }}</p>
          </div>
          <div class="px-5 py-3 text-center">
            <p class="text-xs text-indigo-200 font-medium uppercase tracking-wide">End Date</p>
            <p class="text-lg font-bold text-white">{{ fmtDate(project.end_date) }}</p>
          </div>
          <div class="px-5 py-3 text-center">
            <p class="text-xs text-indigo-200 font-medium uppercase tracking-wide">Tasks</p>
            <p class="text-lg font-bold text-white">{{ project.task_count || 0 }}</p>
          </div>
        </div>
      </div>

      <!-- Progress bar -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex justify-between text-sm mb-2">
          <span class="font-medium text-gray-700 dark:text-gray-200">Overall Progress</span>
          <span class="font-bold text-indigo-600 dark:text-indigo-400">{{ project.progress_pct || 0 }}%</span>
        </div>
        <div class="h-3 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
          <div :style="{ width: `${project.progress_pct || 0}%` }" class="h-full bg-gradient-to-r from-indigo-500 to-violet-500 rounded-full transition-all" />
        </div>
      </div>

      <!-- Tabs -->
      <div class="flex gap-1 bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-1.5">
        <button v-for="tab in tabs" :key="tab.id" @click="activeTab = tab.id"
          :class="activeTab === tab.id ? 'bg-indigo-600 text-white shadow' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'"
          class="flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all">
          <component :is="tab.icon" class="w-4 h-4" />
          <span class="hidden sm:inline">{{ tab.label }}</span>
        </button>
      </div>

      <!-- Tasks Tab -->
      <div v-if="activeTab === 'tasks'">
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
          <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-800">
            <h3 class="font-semibold text-gray-900 dark:text-white">Tasks</h3>
            <span class="text-xs text-gray-400">{{ tasks.length }} tasks</span>
          </div>
          <div v-if="loadingTasks" class="flex items-center justify-center py-10"><Loader2 class="w-6 h-6 text-indigo-500 animate-spin" /></div>
          <div v-else-if="tasks.length === 0" class="flex items-center justify-center py-10 text-gray-400 text-sm">No tasks</div>
          <table v-else class="w-full text-sm">
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
              <tr v-for="task in tasks" :key="task.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
                <td class="px-4 py-3 font-medium text-gray-900 dark:text-white">{{ task.title }}</td>
                <td class="px-4 py-3"><span :class="priorBadge(task.priority)" class="px-2 py-0.5 rounded-full text-xs font-semibold">{{ task.priority }}</span></td>
                <td class="px-4 py-3"><span :class="taskStatBadge(task.status)" class="px-2 py-0.5 rounded-full text-xs font-semibold">{{ task.status?.replace('_', ' ') }}</span></td>
                <td class="px-4 py-3 text-gray-400 text-xs">{{ fmtDate(task.due_date) }}</td>
                <td class="px-4 py-3 text-gray-600 dark:text-gray-300 text-xs">{{ task.assignee_name || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Milestones Tab -->
      <div v-if="activeTab === 'milestones'">
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
          <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-800">
            <h3 class="font-semibold text-gray-900 dark:text-white">Milestones</h3>
          </div>
          <div v-if="loadingMs" class="flex items-center justify-center py-10"><Loader2 class="w-6 h-6 text-indigo-500 animate-spin" /></div>
          <div v-else-if="milestones.length === 0" class="flex items-center justify-center py-10 text-gray-400 text-sm">No milestones</div>
          <div v-else class="divide-y divide-gray-100 dark:divide-gray-800">
            <div v-for="ms in milestones" :key="ms.id" class="px-5 py-4 flex items-center gap-4">
              <div :class="ms.status === 'completed' ? 'bg-emerald-100 dark:bg-emerald-900/30' : 'bg-gray-100 dark:bg-gray-800'" class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0">
                <Flag :class="ms.status === 'completed' ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400'" class="w-4 h-4" />
              </div>
              <div class="flex-1 min-w-0">
                <p class="font-medium text-gray-900 dark:text-white">{{ ms.title }}</p>
                <p class="text-xs text-gray-400">Due: {{ fmtDate(ms.due_date) }}</p>
              </div>
              <span :class="msBadge(ms.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold">{{ ms.status?.replace('_', ' ') }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Overview Tab -->
      <div v-if="activeTab === 'overview'" class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 space-y-3">
          <h3 class="font-semibold text-gray-900 dark:text-white mb-3">Project Info</h3>
          <div v-for="row in projectDetails" :key="row.label" class="flex justify-between py-1 border-b border-gray-100 dark:border-gray-800">
            <span class="text-xs font-medium text-gray-400 uppercase tracking-wide">{{ row.label }}</span>
            <span class="text-sm text-gray-700 dark:text-gray-200">{{ row.value }}</span>
          </div>
        </div>
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
          <h3 class="font-semibold text-gray-900 dark:text-white mb-3">Description</h3>
          <p class="text-sm text-gray-600 dark:text-gray-300 whitespace-pre-wrap">{{ project.description || 'No description provided.' }}</p>
        </div>
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowLeft, Loader2, FolderX, Pencil, Flag, CheckSquare, CalendarDays, Info
} from 'lucide-vue-next'
import { projectsAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const route = useRoute()
const router = useRouter()
const store = useAppStore()

const loading = ref(true)
const loadingTasks = ref(false)
const loadingMs = ref(false)
const project = ref<any>(null)
const tasks = ref<any[]>([])
const milestones = ref<any[]>([])
const activeTab = ref('overview')

const tabs = [
  { id: 'overview', label: 'Overview', icon: Info },
  { id: 'tasks', label: 'Tasks', icon: CheckSquare },
  { id: 'milestones', label: 'Milestones', icon: Flag },
]

const projectDetails = computed(() => [
  { label: 'Code', value: project.value?.code },
  { label: 'Status', value: project.value?.status },
  { label: 'Manager', value: project.value?.manager_name || '—' },
  { label: 'Customer', value: project.value?.customer_name || '—' },
  { label: 'Start Date', value: fmtDate(project.value?.start_date) },
  { label: 'End Date', value: fmtDate(project.value?.end_date) },
  { label: 'Budget', value: fmtCurrency(project.value?.budget) },
  { label: 'Actual Cost', value: fmtCurrency(project.value?.actual_cost) },
])

const COLORS = ['bg-gradient-to-br from-indigo-500 to-blue-600', 'bg-gradient-to-br from-violet-500 to-purple-600', 'bg-gradient-to-br from-emerald-500 to-teal-600', 'bg-gradient-to-br from-amber-500 to-orange-500']
function projectColor(name: string) { return COLORS[name.charCodeAt(0) % COLORS.length] }
function initials(name: string) { return name?.split(' ').slice(0, 2).map(w => w[0]).join('').toUpperCase() }
function fmtDate(d?: string) { if (!d) return '—'; return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' }) }
function fmtNum(n?: number) { if (!n) return '0'; if (n >= 1e6) return (n/1e6).toFixed(1)+'M'; if (n >= 1e3) return (n/1e3).toFixed(0)+'K'; return n.toString() }
function fmtCurrency(n?: number) { if (!n) return '0 DZD'; return fmtNum(n) + ' DZD' }
function statusBadge(s?: string) {
  switch (s) {
    case 'active': return 'bg-emerald-500/20 text-emerald-200'
    case 'planning': return 'bg-blue-500/20 text-blue-200'
    case 'on_hold': return 'bg-amber-500/20 text-amber-200'
    case 'completed': return 'bg-violet-500/20 text-violet-200'
    default: return 'bg-gray-500/20 text-gray-200'
  }
}
function priorBadge(p?: string) {
  switch (p) {
    case 'critical': return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    case 'high': return 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400'
    case 'medium': return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    default: return 'bg-gray-100 text-gray-600'
  }
}
function taskStatBadge(s?: string) {
  switch (s) {
    case 'done': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'in_progress': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    default: return 'bg-gray-100 text-gray-600'
  }
}
function msBadge(s?: string) {
  switch (s) {
    case 'completed': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'in_progress': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    default: return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
  }
}

function openEdit() { router.push(`/projects?edit=${route.params.id}`) }

watch(activeTab, async (tab) => {
  const id = route.params.id as string
  if (tab === 'tasks' && tasks.value.length === 0) {
    loadingTasks.value = true
    try { tasks.value = (await projectsAPI.getTasks(id)).data || [] } catch { /* skip */ }
    finally { loadingTasks.value = false }
  }
  if (tab === 'milestones' && milestones.value.length === 0) {
    loadingMs.value = true
    try { milestones.value = (await projectsAPI.getMilestones(id)).data || [] } catch { /* skip */ }
    finally { loadingMs.value = false }
  }
})

onMounted(async () => {
  loading.value = true
  try {
    const res = await projectsAPI.getProject(route.params.id as string)
    project.value = res.data
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Project not found', 'error')
  } finally {
    loading.value = false
  }
})
</script>
