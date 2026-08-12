<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Tasks</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">All tasks across all projects</p>
      </div>
      <button @click="openCreate" class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm">
        <Plus class="w-4 h-4" />
        New Task
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-5 gap-4">
      <div v-for="stat in taskStats" :key="stat.label" class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4 flex items-center gap-3">
        <div :class="stat.bg" class="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0">
          <component :is="stat.icon" :class="stat.color" class="w-5 h-5" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ stat.label }}</p>
          <p class="text-xl font-bold text-gray-900 dark:text-white">{{ stat.value }}</p>
        </div>
      </div>
    </div>

    <!-- Filters & View Toggle -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
      <div class="flex flex-wrap gap-3">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input v-model="search" placeholder="Search tasks..." class="w-full pl-9 pr-4 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white placeholder-gray-400" />
        </div>
        <div class="relative">
          <select v-model="filterProject" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="">All Projects</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
        <div class="relative">
          <select v-model="filterPriority" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="">All Priorities</option>
            <option value="critical">Critical</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
        <!-- View toggle -->
        <div class="flex bg-gray-100 dark:bg-gray-800 rounded-lg p-1">
          <button @click="viewMode = 'kanban'" :class="viewMode === 'kanban' ? 'bg-white dark:bg-gray-700 shadow text-indigo-600 dark:text-indigo-400' : 'text-gray-500 dark:text-gray-400'" class="px-3 py-1.5 rounded-md text-xs font-medium transition-all flex items-center gap-1.5">
            <LayoutGrid class="w-3.5 h-3.5" /> Kanban
          </button>
          <button @click="viewMode = 'list'" :class="viewMode === 'list' ? 'bg-white dark:bg-gray-700 shadow text-indigo-600 dark:text-indigo-400' : 'text-gray-500 dark:text-gray-400'" class="px-3 py-1.5 rounded-md text-xs font-medium transition-all flex items-center gap-1.5">
            <List class="w-3.5 h-3.5" /> List
          </button>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <Loader2 class="w-10 h-10 text-indigo-500 animate-spin" />
    </div>

    <!-- Kanban Board -->
    <div v-else-if="viewMode === 'kanban'" class="flex gap-4 overflow-x-auto pb-4">
      <div v-for="col in kanbanCols" :key="col.status" class="flex-shrink-0 w-72">
        <div class="flex items-center justify-between mb-3 px-1">
          <div class="flex items-center gap-2">
            <div :class="col.dotColor" class="w-2.5 h-2.5 rounded-full" />
            <span class="text-sm font-semibold text-gray-700 dark:text-gray-200">{{ col.label }}</span>
            <span class="text-xs bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-300 rounded-full px-2 py-0.5 font-medium">{{ tasksByStatus(col.status).length }}</span>
          </div>
        </div>
        <div class="space-y-2 min-h-24">
          <div v-for="task in tasksByStatus(col.status)" :key="task.id"
            class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-3.5 hover:border-indigo-300 dark:hover:border-indigo-700 hover:shadow-md cursor-pointer transition-all group">

            <!-- Priority badge -->
            <div class="flex items-start justify-between gap-2 mb-2">
              <span :class="priorityClass(task.priority)" class="px-2 py-0.5 rounded-full text-xs font-semibold flex-shrink-0">{{ task.priority }}</span>
              <div class="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <button @click.stop="openEdit(task)" class="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded text-gray-400 hover:text-indigo-600 transition-colors">
                  <Pencil class="w-3 h-3" />
                </button>
              </div>
            </div>

            <p class="text-sm font-semibold text-gray-900 dark:text-white mb-1 line-clamp-2">{{ task.title }}</p>
            <p class="text-xs text-gray-400 mb-2">{{ task.project_name }}</p>

            <!-- Due & hours -->
            <div class="flex items-center justify-between text-xs text-gray-400 mt-2 pt-2 border-t border-gray-100 dark:border-gray-800">
              <div class="flex items-center gap-1">
                <CalendarDays class="w-3 h-3" />
                <span :class="isOverdue(task.due_date) ? 'text-red-500 font-medium' : ''">{{ fmtDate(task.due_date) }}</span>
              </div>
              <div v-if="task.estimated_hours" class="flex items-center gap-1">
                <Clock class="w-3 h-3" />
                <span>{{ task.actual_hours || 0 }}/{{ task.estimated_hours }}h</span>
              </div>
            </div>

            <!-- Assignee -->
            <div v-if="task.assignee_name" class="flex items-center gap-1.5 mt-2">
              <div class="w-5 h-5 rounded-full bg-gradient-to-br from-indigo-400 to-violet-500 flex items-center justify-center text-white text-xs font-bold">
                {{ task.assignee_name.charAt(0) }}
              </div>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ task.assignee_name }}</span>
            </div>
          </div>
          <div v-if="tasksByStatus(col.status).length === 0" class="text-center py-6 text-xs text-gray-400 border-2 border-dashed border-gray-200 dark:border-gray-700 rounded-xl">
            No tasks
          </div>
        </div>
      </div>
    </div>

    <!-- List View -->
    <div v-else class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
      <div v-if="filteredTasks.length === 0" class="flex flex-col items-center justify-center py-16 text-gray-400">
        <CheckSquare class="w-12 h-12 mb-3 opacity-30" />
        <p class="text-sm">No tasks found</p>
      </div>
      <table v-else class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Task</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Project</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Priority</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Status</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Assignee</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Due Date</th>
            <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Hours</th>
            <th class="px-4 py-3 w-16"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
          <tr v-for="task in filteredTasks" :key="task.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
            <td class="px-4 py-3">
              <p class="font-medium text-gray-900 dark:text-white line-clamp-1">{{ task.title }}</p>
              <p v-if="task.description" class="text-xs text-gray-400 mt-0.5 line-clamp-1">{{ task.description }}</p>
            </td>
            <td class="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">{{ task.project_name }}</td>
            <td class="px-4 py-3">
              <span :class="priorityClass(task.priority)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold">{{ task.priority }}</span>
            </td>
            <td class="px-4 py-3">
              <span :class="taskStatusClass(task.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold">{{ task.status?.replace('_', ' ') }}</span>
            </td>
            <td class="px-4 py-3">
              <div v-if="task.assignee_name" class="flex items-center gap-2">
                <div class="w-6 h-6 rounded-full bg-gradient-to-br from-indigo-400 to-violet-500 flex items-center justify-center text-white text-xs font-bold">{{ task.assignee_name.charAt(0) }}</div>
                <span class="text-sm text-gray-600 dark:text-gray-300 truncate max-w-24">{{ task.assignee_name }}</span>
              </div>
              <span v-else class="text-gray-400">—</span>
            </td>
            <td class="px-4 py-3">
              <span :class="isOverdue(task.due_date) && task.status !== 'done' ? 'text-red-500 font-medium' : 'text-gray-600 dark:text-gray-300'">{{ fmtDate(task.due_date) }}</span>
            </td>
            <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-300">
              {{ task.actual_hours || 0 }}/{{ task.estimated_hours || 0 }}h
            </td>
            <td class="px-4 py-3" @click.stop>
              <div class="flex gap-1">
                <button @click="openEdit(task)" class="p-1.5 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg text-gray-400 hover:text-indigo-600 transition-colors">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button @click="confirmDelete(task)" class="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg text-gray-400 hover:text-red-600 transition-colors">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Task Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeModal" />
          <div class="relative w-full max-w-lg bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden">
            <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-indigo-600 to-violet-600">
              <div class="flex items-center gap-3 text-white">
                <CheckSquare class="w-5 h-5" />
                <h2 class="font-semibold text-lg">{{ editingTask ? 'Edit Task' : 'New Task' }}</h2>
              </div>
              <button @click="closeModal" class="p-1 hover:bg-white/20 rounded-lg text-white transition-colors">
                <X class="w-5 h-5" />
              </button>
            </div>
            <form @submit.prevent="saveTask" class="p-6 space-y-4">
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Task Title *</label>
                <input v-model="form.title" required placeholder="Task title" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
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
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Priority</label>
                  <div class="relative">
                    <select v-model="form.priority" class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="low">Low</option>
                      <option value="medium">Medium</option>
                      <option value="high">High</option>
                      <option value="critical">Critical</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Status</label>
                  <div class="relative">
                    <select v-model="form.status" class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="todo">To Do</option>
                      <option value="in_progress">In Progress</option>
                      <option value="review">Review</option>
                      <option value="done">Done</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Assignee</label>
                  <div class="relative">
                    <select v-model="form.assignee_id" class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="">— None —</option>
                      <option v-for="e in employees" :key="e.id" :value="e.id">{{ e.full_name }}</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Due Date</label>
                  <input type="date" v-model="form.due_date" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Est. Hours</label>
                  <input type="number" v-model.number="form.estimated_hours" placeholder="0" min="0" step="0.5" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
              </div>
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Description</label>
                <textarea v-model="form.description" rows="3" placeholder="Task description..." class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white resize-none" />
              </div>
              <div class="flex gap-3">
                <button type="button" @click="closeModal" class="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
                <button type="submit" :disabled="saving" class="flex-1 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center justify-center gap-2">
                  <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                  {{ saving ? 'Saving...' : (editingTask ? 'Update' : 'Create') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Delete Confirm -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showDeleteModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showDeleteModal = false" />
          <div class="relative w-full max-w-md bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 p-6">
            <div class="flex items-center gap-4 mb-4">
              <div class="w-12 h-12 rounded-xl bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
                <AlertTriangle class="w-6 h-6 text-red-600" />
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">Delete Task</h3>
                <p class="text-sm text-gray-400">Are you sure you want to delete this task?</p>
              </div>
            </div>
            <div class="flex gap-3">
              <button @click="showDeleteModal = false" class="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
              <button @click="deleteTask" :disabled="saving" class="flex-1 px-4 py-2.5 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2">
                <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                Delete
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
  Plus, Search, Loader2, Pencil, Trash2, X, ChevronDown,
  CheckSquare, CalendarDays, Clock, LayoutGrid, List, AlertTriangle
} from 'lucide-vue-next'
import { projectsAPI, hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

interface Task {
  id: string; title: string; description?: string; status: string; priority: string
  project_id: string; project_name?: string; assignee_id?: string; assignee_name?: string
  due_date?: string; estimated_hours?: number; actual_hours?: number
}

const loading = ref(true)
const saving = ref(false)
const tasks = ref<Task[]>([])
const projects = ref<any[]>([])
const employees = ref<any[]>([])
const search = ref('')
const filterProject = ref('')
const filterPriority = ref('')
const viewMode = ref<'kanban' | 'list'>('kanban')
const showModal = ref(false)
const showDeleteModal = ref(false)
const editingTask = ref<Task | null>(null)
const deletingTask = ref<Task | null>(null)

const form = ref({
  title: '', description: '', project_id: '', priority: 'medium', status: 'todo',
  assignee_id: '', due_date: '', estimated_hours: 0
})

const kanbanCols = [
  { status: 'todo', label: 'To Do', dotColor: 'bg-gray-400' },
  { status: 'in_progress', label: 'In Progress', dotColor: 'bg-blue-500' },
  { status: 'review', label: 'Review', dotColor: 'bg-amber-500' },
  { status: 'done', label: 'Done', dotColor: 'bg-emerald-500' },
]

const taskStats = computed(() => [
  { label: 'Total', value: tasks.value.length, icon: CheckSquare, bg: 'bg-indigo-100 dark:bg-indigo-900/40', color: 'text-indigo-600 dark:text-indigo-400' },
  { label: 'To Do', value: tasks.value.filter(t => t.status === 'todo').length, icon: CheckSquare, bg: 'bg-gray-100 dark:bg-gray-800', color: 'text-gray-600 dark:text-gray-400' },
  { label: 'In Progress', value: tasks.value.filter(t => t.status === 'in_progress').length, icon: Clock, bg: 'bg-blue-100 dark:bg-blue-900/40', color: 'text-blue-600 dark:text-blue-400' },
  { label: 'Review', value: tasks.value.filter(t => t.status === 'review').length, icon: CheckSquare, bg: 'bg-amber-100 dark:bg-amber-900/40', color: 'text-amber-600 dark:text-amber-400' },
  { label: 'Done', value: tasks.value.filter(t => t.status === 'done').length, icon: CheckSquare, bg: 'bg-emerald-100 dark:bg-emerald-900/40', color: 'text-emerald-600 dark:text-emerald-400' },
])

const filteredTasks = computed(() => {
  let list = [...tasks.value]
  if (search.value) { const q = search.value.toLowerCase(); list = list.filter(t => t.title.toLowerCase().includes(q)) }
  if (filterProject.value) list = list.filter(t => t.project_id === filterProject.value)
  if (filterPriority.value) list = list.filter(t => t.priority === filterPriority.value)
  return list
})

function tasksByStatus(status: string) {
  return filteredTasks.value.filter(t => t.status === status)
}

function fmtDate(d?: string) { if (!d) return '—'; return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short' }) }
function isOverdue(d?: string) { if (!d) return false; return new Date(d) < new Date() }

function priorityClass(p?: string) {
  switch (p) {
    case 'critical': return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    case 'high': return 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400'
    case 'medium': return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    case 'low': return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
    default: return 'bg-gray-100 text-gray-600'
  }
}
function taskStatusClass(s?: string) {
  switch (s) {
    case 'done': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'in_progress': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    case 'review': return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    case 'cancelled': return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default: return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
  }
}

async function load() {
  loading.value = true
  try {
    const res = await projectsAPI.getAllTasks()
    tasks.value = res.data || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load tasks', 'error')
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
  editingTask.value = null
  form.value = { title: '', description: '', project_id: filterProject.value || '', priority: 'medium', status: 'todo', assignee_id: '', due_date: '', estimated_hours: 0 }
  showModal.value = true
}
function openEdit(task: Task) {
  editingTask.value = task
  form.value = {
    title: task.title, description: task.description || '', project_id: task.project_id,
    priority: task.priority, status: task.status, assignee_id: task.assignee_id || '',
    due_date: task.due_date?.slice(0, 10) || '', estimated_hours: task.estimated_hours || 0
  }
  showModal.value = true
}
function closeModal() { showModal.value = false; editingTask.value = null }

async function saveTask() {
  saving.value = true
  try {
    const payload = { ...form.value, assignee_id: form.value.assignee_id || null }
    if (editingTask.value) {
      await projectsAPI.updateTask(editingTask.value.id, payload)
      store.addToast('Task updated', 'success')
    } else {
      await projectsAPI.createTask(form.value.project_id, payload)
      store.addToast('Task created', 'success')
    }
    closeModal()
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

function confirmDelete(task: Task) { deletingTask.value = task; showDeleteModal.value = true }
async function deleteTask() {
  if (!deletingTask.value) return
  saving.value = true
  try {
    await projectsAPI.deleteTask(deletingTask.value.id)
    store.addToast('Task deleted', 'success')
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
