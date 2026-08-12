<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Projects</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Manage and track all company projects</p>
      </div>
      <button @click="openCreate" class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm">
        <Plus class="w-4 h-4" />
        New Project
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-5 gap-4">
      <div v-for="kpi in kpis" :key="kpi.label" class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4 flex items-center gap-3">
        <div :class="kpi.iconBg" class="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0">
          <component :is="kpi.icon" :class="kpi.iconColor" class="w-5 h-5" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium">{{ kpi.label }}</p>
          <p class="text-xl font-bold text-gray-900 dark:text-white">{{ kpi.value }}</p>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
      <div class="flex flex-wrap gap-3">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input v-model="search" placeholder="Search projects..." class="w-full pl-9 pr-4 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white placeholder-gray-400" />
        </div>
        <div class="relative">
          <select v-model="filterStatus" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="">All Status</option>
            <option value="planning">Planning</option>
            <option value="active">Active</option>
            <option value="on_hold">On Hold</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
        <div class="relative">
          <select v-model="viewMode" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="table">Table View</option>
            <option value="cards">Card View</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <Loader2 class="w-10 h-10 text-indigo-500 animate-spin" />
    </div>

    <div v-else-if="filteredProjects.length === 0" class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 flex flex-col items-center justify-center py-20 text-gray-400">
      <FolderKanban class="w-14 h-14 mb-3 opacity-30" />
      <p class="font-medium">No projects found</p>
      <p class="text-xs mt-1">Create your first project to get started</p>
    </div>

    <!-- Card View -->
    <div v-else-if="viewMode === 'cards'" class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      <div v-for="proj in filteredProjects" :key="proj.id"
        @click="$router.push(`/projects/${proj.id}`)"
        class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 hover:border-indigo-300 dark:hover:border-indigo-700 hover:shadow-lg cursor-pointer transition-all group">

        <div class="flex items-start justify-between mb-4">
          <div class="flex items-center gap-3">
            <div :class="projectColor(proj.name)" class="w-10 h-10 rounded-xl flex items-center justify-center text-white text-sm font-bold flex-shrink-0 shadow-sm">
              {{ initials(proj.name) }}
            </div>
            <div>
              <h3 class="font-semibold text-gray-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors line-clamp-1">{{ proj.name }}</h3>
              <p class="text-xs text-gray-400 font-mono">{{ proj.code }}</p>
            </div>
          </div>
          <span :class="statusClass(proj.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold flex-shrink-0">{{ proj.status }}</span>
        </div>

        <!-- Progress -->
        <div class="mb-3">
          <div class="flex justify-between text-xs text-gray-500 dark:text-gray-400 mb-1.5">
            <span>Progress</span>
            <span class="font-semibold">{{ proj.progress_pct || 0 }}%</span>
          </div>
          <div class="h-2 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
            <div :style="{ width: `${proj.progress_pct || 0}%` }" :class="progressColor(proj.progress_pct)" class="h-full rounded-full transition-all" />
          </div>
        </div>

        <!-- Stats Row -->
        <div class="grid grid-cols-3 gap-2 mb-4 text-center">
          <div class="bg-gray-50 dark:bg-gray-800 rounded-lg py-1.5">
            <p class="text-xs text-gray-400">Tasks</p>
            <p class="text-sm font-bold text-gray-900 dark:text-white">{{ proj.task_count || 0 }}</p>
          </div>
          <div class="bg-gray-50 dark:bg-gray-800 rounded-lg py-1.5">
            <p class="text-xs text-gray-400">Done</p>
            <p class="text-sm font-bold text-emerald-600 dark:text-emerald-400">{{ proj.done_count || 0 }}</p>
          </div>
          <div class="bg-gray-50 dark:bg-gray-800 rounded-lg py-1.5">
            <p class="text-xs text-gray-400">Overdue</p>
            <p class="text-sm font-bold text-red-500">{{ proj.overdue_count || 0 }}</p>
          </div>
        </div>

        <!-- Footer -->
        <div class="flex items-center justify-between text-xs text-gray-400">
          <div class="flex items-center gap-1">
            <CalendarDays class="w-3.5 h-3.5" />
            <span>{{ fmtDate(proj.end_date) }}</span>
          </div>
          <div class="flex items-center gap-1">
            <DollarSign class="w-3.5 h-3.5" />
            <span>{{ fmtNum(proj.actual_cost) }} / {{ fmtNum(proj.budget) }}</span>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex gap-2 mt-3 pt-3 border-t border-gray-100 dark:border-gray-800" @click.stop>
          <button @click="$router.push(`/projects/${proj.id}`)" class="flex-1 text-xs py-1.5 bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 rounded-lg font-medium hover:bg-indigo-100 dark:hover:bg-indigo-900/40 transition-colors">
            View Details
          </button>
          <button @click="openEdit(proj)" class="p-1.5 bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">
            <Pencil class="w-3.5 h-3.5" />
          </button>
          <button @click="confirmDelete(proj)" class="p-1.5 bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 hover:text-red-500 transition-colors">
            <Trash2 class="w-3.5 h-3.5" />
          </button>
        </div>
      </div>
    </div>

    <!-- Table View -->
    <div v-else class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Project</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Status</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Progress</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Manager</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Budget</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">End Date</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Tasks</th>
            <th class="px-4 py-3 w-20"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
          <tr v-for="proj in filteredProjects" :key="proj.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/40 cursor-pointer transition-colors" @click="$router.push(`/projects/${proj.id}`)">
            <td class="px-4 py-3">
              <div class="flex items-center gap-3">
                <div :class="projectColor(proj.name)" class="w-8 h-8 rounded-lg flex items-center justify-center text-white text-xs font-bold flex-shrink-0">
                  {{ initials(proj.name) }}
                </div>
                <div>
                  <p class="font-semibold text-gray-900 dark:text-white">{{ proj.name }}</p>
                  <p class="text-xs text-gray-400 font-mono">{{ proj.code }}</p>
                </div>
              </div>
            </td>
            <td class="px-4 py-3">
              <span :class="statusClass(proj.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold">{{ proj.status }}</span>
            </td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-2 w-32">
                <div class="flex-1 h-1.5 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
                  <div :style="{ width: `${proj.progress_pct || 0}%` }" :class="progressColor(proj.progress_pct)" class="h-full rounded-full" />
                </div>
                <span class="text-xs font-semibold text-gray-700 dark:text-gray-200 w-8">{{ proj.progress_pct || 0 }}%</span>
              </div>
            </td>
            <td class="px-4 py-3 text-gray-600 dark:text-gray-300">{{ proj.manager_name || '—' }}</td>
            <td class="px-4 py-3">
              <div class="text-sm">
                <span class="text-gray-900 dark:text-white font-medium">{{ fmtNum(proj.budget) }}</span>
                <span v-if="proj.actual_cost > 0" class="text-xs text-gray-400 block">Spent: {{ fmtNum(proj.actual_cost) }}</span>
              </div>
            </td>
            <td class="px-4 py-3 text-gray-600 dark:text-gray-300">{{ fmtDate(proj.end_date) }}</td>
            <td class="px-4 py-3">
              <div class="flex items-center gap-1.5 text-xs">
                <span class="font-semibold text-gray-900 dark:text-white">{{ proj.task_count || 0 }}</span>
                <span class="text-gray-400">total</span>
                <span class="text-emerald-500 font-semibold">{{ proj.done_count || 0 }}</span>
                <span class="text-gray-400">done</span>
              </div>
            </td>
            <td class="px-4 py-3" @click.stop>
              <div class="flex items-center gap-1">
                <button @click="openEdit(proj)" class="p-1.5 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg text-gray-400 hover:text-indigo-600 transition-colors">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button @click="confirmDelete(proj)" class="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg text-gray-400 hover:text-red-600 transition-colors">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-gray-800 text-sm text-gray-500 dark:text-gray-400">
        <span>{{ filteredProjects.length }} project(s)</span>
      </div>
    </div>

    <!-- Create / Edit Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeModal" />
          <div class="relative w-full max-w-2xl bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 max-h-[90vh] flex flex-col">
            <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-indigo-600 to-violet-600 rounded-t-2xl flex-shrink-0">
              <div class="flex items-center gap-3 text-white">
                <FolderKanban class="w-5 h-5" />
                <h2 class="font-semibold text-lg">{{ editingProj ? 'Edit Project' : 'New Project' }}</h2>
              </div>
              <button @click="closeModal" class="p-1 hover:bg-white/20 rounded-lg text-white transition-colors">
                <X class="w-5 h-5" />
              </button>
            </div>
            <form @submit.prevent="saveProject" class="p-6 space-y-4 overflow-y-auto">
              <div class="grid grid-cols-2 gap-4">
                <div class="col-span-2">
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Project Name *</label>
                  <input v-model="form.name" required placeholder="Project name" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Status</label>
                  <div class="relative">
                    <select v-model="form.status" class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="planning">Planning</option>
                      <option value="active">Active</option>
                      <option value="on_hold">On Hold</option>
                      <option value="completed">Completed</option>
                      <option value="cancelled">Cancelled</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Manager</label>
                  <div class="relative">
                    <select v-model="form.manager_id" class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="">— Select Manager —</option>
                      <option v-for="emp in employees" :key="emp.id" :value="emp.id">{{ emp.full_name }}</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Start Date</label>
                  <input type="date" v-model="form.start_date" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">End Date</label>
                  <input type="date" v-model="form.end_date" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Budget (DZD)</label>
                  <input type="number" v-model.number="form.budget" placeholder="0.00" min="0" step="1000" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Customer</label>
                  <div class="relative">
                    <select v-model="form.customer_id" class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="">— Select Customer —</option>
                      <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div class="col-span-2">
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Description</label>
                  <textarea v-model="form.description" rows="3" placeholder="Project description..." class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white resize-none" />
                </div>
              </div>
              <div class="flex gap-3 pt-2">
                <button type="button" @click="closeModal" class="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
                <button type="submit" :disabled="saving" class="flex-1 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center justify-center gap-2">
                  <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                  {{ saving ? 'Saving...' : (editingProj ? 'Update' : 'Create') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Delete Confirmation -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showDeleteModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showDeleteModal = false" />
          <div class="relative w-full max-w-md bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 p-6">
            <div class="flex items-center gap-4 mb-4">
              <div class="w-12 h-12 rounded-xl bg-red-100 dark:bg-red-900/30 flex items-center justify-center flex-shrink-0">
                <AlertTriangle class="w-6 h-6 text-red-600" />
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">Cancel Project</h3>
                <p class="text-sm text-gray-500 dark:text-gray-400">This will set the project status to cancelled.</p>
              </div>
            </div>
            <p class="text-sm text-gray-600 dark:text-gray-300 mb-5">Are you sure you want to cancel <strong class="text-gray-900 dark:text-white">{{ deletingProj?.name }}</strong>?</p>
            <div class="flex gap-3">
              <button @click="showDeleteModal = false" class="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
              <button @click="deleteProject" :disabled="saving" class="flex-1 px-4 py-2.5 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2">
                <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                Confirm Cancel
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Plus, Search, Loader2, FolderKanban, Pencil, Trash2, X, ChevronDown,
  CalendarDays, DollarSign, AlertTriangle, CheckCircle, Clock, PauseCircle, XCircle
} from 'lucide-vue-next'
import { projectsAPI, hrAPI, salesAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

interface Project {
  id: string; code: string; name: string; status: string
  start_date?: string; end_date?: string; budget?: number; actual_cost?: number
  manager_id?: string; manager_name?: string; customer_id?: string; customer_name?: string
  description?: string; progress_pct?: number; task_count?: number; done_count?: number
  overdue_count?: number; created_at?: string
}

const loading = ref(true)
const saving = ref(false)
const projects = ref<Project[]>([])
const employees = ref<any[]>([])
const customers = ref<any[]>([])
const search = ref('')
const filterStatus = ref('')
const viewMode = ref('table')
const showModal = ref(false)
const showDeleteModal = ref(false)
const editingProj = ref<Project | null>(null)
const deletingProj = ref<Project | null>(null)

const form = ref({
  name: '', status: 'planning', manager_id: '', customer_id: '',
  start_date: '', end_date: '', budget: 0, description: ''
})

const kpis = computed(() => {
  const total = projects.value.length
  const active = projects.value.filter(p => p.status === 'active').length
  const completed = projects.value.filter(p => p.status === 'completed').length
  const on_hold = projects.value.filter(p => p.status === 'on_hold').length
  const totalBudget = projects.value.reduce((s, p) => s + (p.budget || 0), 0)
  return [
    { label: 'Total', value: total, icon: FolderKanban, iconBg: 'bg-indigo-100 dark:bg-indigo-900/40', iconColor: 'text-indigo-600 dark:text-indigo-400' },
    { label: 'Active', value: active, icon: CheckCircle, iconBg: 'bg-emerald-100 dark:bg-emerald-900/40', iconColor: 'text-emerald-600 dark:text-emerald-400' },
    { label: 'Completed', value: completed, icon: CheckCircle, iconBg: 'bg-blue-100 dark:bg-blue-900/40', iconColor: 'text-blue-600 dark:text-blue-400' },
    { label: 'On Hold', value: on_hold, icon: PauseCircle, iconBg: 'bg-amber-100 dark:bg-amber-900/40', iconColor: 'text-amber-600 dark:text-amber-400' },
    { label: 'Budget (DZD)', value: fmtNum(totalBudget), icon: DollarSign, iconBg: 'bg-violet-100 dark:bg-violet-900/40', iconColor: 'text-violet-600 dark:text-violet-400' },
  ]
})

const filteredProjects = computed(() => {
  let list = [...projects.value]
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(p => p.name.toLowerCase().includes(q) || p.code.toLowerCase().includes(q))
  }
  if (filterStatus.value) list = list.filter(p => p.status === filterStatus.value)
  return list
})

const COLORS = [
  'bg-gradient-to-br from-indigo-500 to-blue-600',
  'bg-gradient-to-br from-violet-500 to-purple-600',
  'bg-gradient-to-br from-emerald-500 to-teal-600',
  'bg-gradient-to-br from-amber-500 to-orange-500',
  'bg-gradient-to-br from-rose-500 to-pink-600',
  'bg-gradient-to-br from-cyan-500 to-sky-600',
]
function projectColor(name: string) { return COLORS[name.charCodeAt(0) % COLORS.length] }
function initials(name: string) { return name.split(' ').slice(0, 2).map(w => w[0]).join('').toUpperCase() }
function fmtDate(d?: string) { if (!d) return '—'; return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' }) }
function fmtNum(n?: number) {
  if (!n) return '0'
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(0) + 'K'
  return n.toString()
}

function statusClass(s?: string) {
  switch (s) {
    case 'active': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'planning': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    case 'on_hold': return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    case 'completed': return 'bg-violet-100 text-violet-700 dark:bg-violet-900/30 dark:text-violet-400'
    case 'cancelled': return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default: return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
  }
}
function progressColor(pct?: number) {
  if (!pct) return 'bg-gray-300 dark:bg-gray-600'
  if (pct >= 80) return 'bg-emerald-500'
  if (pct >= 50) return 'bg-indigo-500'
  return 'bg-amber-500'
}

async function load() {
  loading.value = true
  try {
    const res = await projectsAPI.getProjects()
    projects.value = res.data || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load projects', 'error')
  } finally {
    loading.value = false
  }
}

async function loadDropdowns() {
  try {
    const [eRes, cRes] = await Promise.all([hrAPI.getEmployees(), salesAPI.getCustomers()])
    employees.value = eRes.data || []
    customers.value = cRes.data || []
  } catch { /* silently fail */ }
}

function openCreate() {
  editingProj.value = null
  form.value = { name: '', status: 'planning', manager_id: '', customer_id: '', start_date: '', end_date: '', budget: 0, description: '' }
  showModal.value = true
}
function openEdit(proj: Project) {
  editingProj.value = proj
  form.value = {
    name: proj.name,
    status: proj.status,
    manager_id: proj.manager_id || '',
    customer_id: proj.customer_id || '',
    start_date: proj.start_date?.slice(0, 10) || '',
    end_date: proj.end_date?.slice(0, 10) || '',
    budget: proj.budget || 0,
    description: proj.description || ''
  }
  showModal.value = true
}
function closeModal() { showModal.value = false; editingProj.value = null }

async function saveProject() {
  saving.value = true
  try {
    const payload = { ...form.value, manager_id: form.value.manager_id || null, customer_id: form.value.customer_id || null }
    if (editingProj.value) {
      await projectsAPI.updateProject(editingProj.value.id, payload)
      store.addToast('Project updated', 'success')
    } else {
      await projectsAPI.createProject(payload)
      store.addToast('Project created', 'success')
    }
    closeModal()
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

function confirmDelete(proj: Project) { deletingProj.value = proj; showDeleteModal.value = true }
async function deleteProject() {
  if (!deletingProj.value) return
  saving.value = true
  try {
    await projectsAPI.deleteProject(deletingProj.value.id)
    store.addToast('Project cancelled', 'success')
    showDeleteModal.value = false
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Delete failed', 'error')
  } finally {
    saving.value = false
  }
}

onMounted(() => { load(); loadDropdowns() })
</script>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
