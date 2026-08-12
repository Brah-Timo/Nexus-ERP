<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Departments</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Manage organizational departments and hierarchy</p>
      </div>
      <button @click="openCreate" class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm">
        <Plus class="w-4 h-4" />
        New Department
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center flex-shrink-0">
          <Building2 class="w-6 h-6 text-indigo-600 dark:text-indigo-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Total</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ kpis.total }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-emerald-100 dark:bg-emerald-900/40 flex items-center justify-center flex-shrink-0">
          <CheckCircle class="w-6 h-6 text-emerald-600 dark:text-emerald-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Active</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ kpis.active }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-violet-100 dark:bg-violet-900/40 flex items-center justify-center flex-shrink-0">
          <UserCheck class="w-6 h-6 text-violet-600 dark:text-violet-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Managed</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ kpis.managed }}</p>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-center gap-4">
        <div class="w-12 h-12 rounded-xl bg-amber-100 dark:bg-amber-900/40 flex items-center justify-center flex-shrink-0">
          <Users class="w-6 h-6 text-amber-600 dark:text-amber-400" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400 font-medium uppercase tracking-wide">Total Staff</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ kpis.headcount }}</p>
        </div>
      </div>
    </div>

    <!-- Filters & Search -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
      <div class="flex flex-col sm:flex-row gap-3">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input v-model="search" placeholder="Search departments..." class="w-full pl-9 pr-4 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white placeholder-gray-400" />
        </div>
        <div class="relative">
          <select v-model="filterStatus" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="">All Status</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
        <div class="relative">
          <select v-model="sortBy" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="name">Sort: Name</option>
            <option value="code">Sort: Code</option>
            <option value="headcount">Sort: Headcount</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <Loader2 class="w-8 h-8 text-indigo-500 animate-spin" />
      </div>
      <div v-else-if="filteredDepts.length === 0" class="flex flex-col items-center justify-center py-20 text-gray-400">
        <Building2 class="w-12 h-12 mb-3 opacity-30" />
        <p class="text-sm font-medium">No departments found</p>
        <p class="text-xs mt-1">Try adjusting your filters or create a new department</p>
      </div>
      <template v-else>
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
              <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300 w-8"></th>
              <th @click="setSort('name')" class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300 cursor-pointer select-none hover:text-indigo-600 dark:hover:text-indigo-400">
                <div class="flex items-center gap-1">
                  Department
                  <ArrowUpDown class="w-3 h-3 opacity-50" />
                </div>
              </th>
              <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Code</th>
              <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Parent</th>
              <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Manager</th>
              <th @click="setSort('headcount')" class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300 cursor-pointer select-none hover:text-indigo-600 dark:hover:text-indigo-400">
                <div class="flex items-center gap-1">
                  Staff
                  <ArrowUpDown class="w-3 h-3 opacity-50" />
                </div>
              </th>
              <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Status</th>
              <th class="px-4 py-3 w-16"></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <template v-for="dept in filteredDepts" :key="dept.id">
              <tr @click="toggleExpand(dept.id)" class="hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer transition-colors">
                <td class="px-4 py-3">
                  <ChevronRight class="w-4 h-4 text-gray-400 transition-transform" :class="{ 'rotate-90': expandedRow === dept.id }" />
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-3">
                    <div :class="avatarBg(dept.name)" class="w-9 h-9 rounded-lg flex items-center justify-center text-white text-xs font-bold flex-shrink-0">
                      {{ initials(dept.name) }}
                    </div>
                    <div>
                      <p class="font-semibold text-gray-900 dark:text-white">{{ dept.name }}</p>
                      <p v-if="dept.description" class="text-xs text-gray-400 truncate max-w-xs">{{ dept.description }}</p>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <span class="font-mono text-xs bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded text-gray-600 dark:text-gray-300">{{ dept.code }}</span>
                </td>
                <td class="px-4 py-3 text-gray-600 dark:text-gray-300 text-sm">{{ dept.parent_name || '—' }}</td>
                <td class="px-4 py-3">
                  <div v-if="dept.manager_name" class="flex items-center gap-2">
                    <div class="w-6 h-6 rounded-full bg-gradient-to-br from-violet-400 to-purple-600 flex items-center justify-center text-white text-xs font-bold">
                      {{ dept.manager_name.charAt(0) }}
                    </div>
                    <span class="text-sm text-gray-600 dark:text-gray-300">{{ dept.manager_name }}</span>
                  </div>
                  <span v-else class="text-gray-400 text-sm">—</span>
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2">
                    <div class="w-7 h-7 rounded-full bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center">
                      <Users class="w-3.5 h-3.5 text-amber-600 dark:text-amber-400" />
                    </div>
                    <span class="font-semibold text-gray-900 dark:text-white">{{ dept.headcount || 0 }}</span>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <span :class="dept.is_active ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'" class="px-2.5 py-1 rounded-full text-xs font-semibold">
                    {{ dept.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </td>
                <td class="px-4 py-3" @click.stop>
                  <div class="flex items-center gap-1">
                    <button @click="openEdit(dept)" class="p-1.5 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors">
                      <Pencil class="w-3.5 h-3.5" />
                    </button>
                    <button @click="confirmDeactivate(dept)" class="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors">
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </td>
              </tr>
              <!-- Expanded detail row -->
              <tr v-if="expandedRow === dept.id" class="bg-indigo-50/50 dark:bg-indigo-900/10">
                <td colspan="8" class="px-4 py-4">
                  <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                    <div>
                      <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-1">Department ID</p>
                      <p class="font-mono text-xs text-gray-600 dark:text-gray-300 truncate">{{ dept.id }}</p>
                    </div>
                    <div>
                      <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-1">Cost Center</p>
                      <p class="text-gray-700 dark:text-gray-200">{{ dept.cost_center || '—' }}</p>
                    </div>
                    <div>
                      <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-1">Created</p>
                      <p class="text-gray-700 dark:text-gray-200">{{ fmtDate(dept.created_at) }}</p>
                    </div>
                    <div class="flex items-center gap-2">
                      <button @click="openEdit(dept)" class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 text-white text-xs font-medium rounded-lg hover:bg-indigo-700 transition-colors">
                        <Pencil class="w-3 h-3" /> Edit
                      </button>
                      <button @click="confirmDeactivate(dept)" class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-red-600 text-white text-xs font-medium rounded-lg hover:bg-red-700 transition-colors">
                        <Trash2 class="w-3 h-3" /> Delete
                      </button>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
        <!-- Pagination -->
        <div class="flex items-center justify-between px-4 py-3 border-t border-gray-200 dark:border-gray-800 text-sm text-gray-500 dark:text-gray-400">
          <span>{{ filteredDepts.length }} department(s) shown</span>
          <div class="flex items-center gap-2">
            <span class="text-xs">Rows per page:</span>
            <select v-model="pageSize" class="text-xs bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded px-2 py-1">
              <option :value="20">20</option>
              <option :value="50">50</option>
              <option :value="100">100</option>
            </select>
          </div>
        </div>
      </template>
    </div>

    <!-- Create / Edit Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeModal" />
          <div class="relative w-full max-w-lg bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden">
            <!-- Modal Header -->
            <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-indigo-600 to-violet-600">
              <div class="flex items-center gap-3 text-white">
                <Building2 class="w-5 h-5" />
                <h2 class="font-semibold text-lg">{{ editingDept ? 'Edit Department' : 'New Department' }}</h2>
              </div>
              <button @click="closeModal" class="p-1 hover:bg-white/20 rounded-lg transition-colors text-white">
                <X class="w-5 h-5" />
              </button>
            </div>
            <!-- Modal Body -->
            <form @submit.prevent="saveDept" class="p-6 space-y-4">
              <div class="grid grid-cols-2 gap-4">
                <div class="col-span-2">
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Department Name *</label>
                  <input v-model="form.name" required placeholder="e.g. Engineering" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Code *</label>
                  <input v-model="form.code" required placeholder="ENG" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Cost Center</label>
                  <input v-model="form.cost_center" placeholder="CC-001" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Parent Department</label>
                  <div class="relative">
                    <select v-model="form.parent_id" class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="">— None —</option>
                      <option v-for="d in departments.filter(x => x.id !== editingDept?.id)" :key="d.id" :value="d.id">{{ d.name }}</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Manager</label>
                  <div class="relative">
                    <select v-model="form.manager_id" class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                      <option value="">— None —</option>
                      <option v-for="emp in employees" :key="emp.id" :value="emp.id">{{ emp.full_name }}</option>
                    </select>
                    <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                  </div>
                </div>
                <div class="col-span-2">
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Description</label>
                  <textarea v-model="form.description" rows="3" placeholder="Department description..." class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white resize-none" />
                </div>
                <div class="col-span-2 flex items-center gap-3">
                  <label class="relative inline-flex items-center cursor-pointer">
                    <input type="checkbox" v-model="form.is_active" class="sr-only peer" />
                    <div class="w-10 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-indigo-500 rounded-full peer peer-checked:bg-indigo-600 after:content-[''] after:absolute after:top-0.5 after:left-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:after:translate-x-5"></div>
                  </label>
                  <span class="text-sm text-gray-700 dark:text-gray-200 font-medium">Active</span>
                </div>
              </div>
              <div class="flex gap-3 pt-2">
                <button type="button" @click="closeModal" class="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
                <button type="submit" :disabled="saving" class="flex-1 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center justify-center gap-2">
                  <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                  {{ saving ? 'Saving...' : (editingDept ? 'Update' : 'Create') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- Deactivate Confirmation Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showDeleteModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showDeleteModal = false" />
          <div class="relative w-full max-w-md bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 p-6">
            <div class="flex items-center gap-4 mb-4">
              <div class="w-12 h-12 rounded-xl bg-red-100 dark:bg-red-900/30 flex items-center justify-center flex-shrink-0">
                <AlertTriangle class="w-6 h-6 text-red-600 dark:text-red-400" />
              </div>
              <div>
                <h3 class="font-semibold text-gray-900 dark:text-white">Deactivate Department</h3>
                <p class="text-sm text-gray-500 dark:text-gray-400">This action cannot be undone easily.</p>
              </div>
            </div>
            <p class="text-sm text-gray-600 dark:text-gray-300 mb-5">Are you sure you want to deactivate <strong class="text-gray-900 dark:text-white">{{ deletingDept?.name }}</strong>? Employees assigned to it may be affected.</p>
            <div class="flex gap-3">
              <button @click="showDeleteModal = false" class="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
              <button @click="deleteDept" :disabled="saving" class="flex-1 px-4 py-2.5 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2">
                <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                Deactivate
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
  Building2, Users, CheckCircle, UserCheck, Search, Plus, Pencil, Trash2,
  Loader2, X, ChevronDown, ChevronRight, ArrowUpDown, AlertTriangle
} from 'lucide-vue-next'
import { hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

interface Department {
  id: string
  code: string
  name: string
  description?: string
  parent_id?: string
  parent_name?: string
  manager_id?: string
  manager_name?: string
  cost_center?: string
  headcount?: number
  is_active: boolean
  created_at?: string
}

interface Employee {
  id: string
  full_name: string
}

const loading = ref(true)
const saving = ref(false)
const departments = ref<Department[]>([])
const employees = ref<Employee[]>([])

const search = ref('')
const filterStatus = ref('')
const sortBy = ref('name')
const pageSize = ref(20)
const expandedRow = ref<string | null>(null)

const showModal = ref(false)
const showDeleteModal = ref(false)
const editingDept = ref<Department | null>(null)
const deletingDept = ref<Department | null>(null)

const form = ref({
  name: '',
  code: '',
  description: '',
  parent_id: '',
  manager_id: '',
  cost_center: '',
  is_active: true
})

const kpis = computed(() => ({
  total: departments.value.length,
  active: departments.value.filter(d => d.is_active).length,
  managed: departments.value.filter(d => d.manager_id).length,
  headcount: departments.value.reduce((s, d) => s + (d.headcount || 0), 0)
}))

const filteredDepts = computed(() => {
  let list = [...departments.value]
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(d => d.name.toLowerCase().includes(q) || d.code.toLowerCase().includes(q))
  }
  if (filterStatus.value === 'active') list = list.filter(d => d.is_active)
  if (filterStatus.value === 'inactive') list = list.filter(d => !d.is_active)
  list.sort((a, b) => {
    if (sortBy.value === 'name') return a.name.localeCompare(b.name)
    if (sortBy.value === 'code') return a.code.localeCompare(b.code)
    if (sortBy.value === 'headcount') return (b.headcount || 0) - (a.headcount || 0)
    return 0
  })
  return list.slice(0, pageSize.value)
})

const COLORS = [
  'from-indigo-500 to-blue-600', 'from-violet-500 to-purple-600',
  'from-emerald-500 to-teal-600', 'from-amber-500 to-orange-600',
  'from-rose-500 to-pink-600', 'from-cyan-500 to-sky-600'
]
function avatarBg(name: string) {
  const idx = name.charCodeAt(0) % COLORS.length
  return `bg-gradient-to-br ${COLORS[idx]}`
}
function initials(name: string) {
  return name.split(' ').slice(0, 2).map(w => w[0]).join('').toUpperCase()
}
function fmtDate(d?: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function setSort(field: string) {
  sortBy.value = field
}

function toggleExpand(id: string) {
  expandedRow.value = expandedRow.value === id ? null : id
}

async function load() {
  loading.value = true
  try {
    const [dRes, eRes] = await Promise.all([
      hrAPI.getDepartments(),
      hrAPI.getEmployees()
    ])
    departments.value = dRes.data || []
    employees.value = (eRes.data || []).map((e: any) => ({ id: e.id, full_name: e.full_name }))
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load departments', 'error')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingDept.value = null
  form.value = { name: '', code: '', description: '', parent_id: '', manager_id: '', cost_center: '', is_active: true }
  showModal.value = true
}

function openEdit(dept: Department) {
  editingDept.value = dept
  form.value = {
    name: dept.name,
    code: dept.code,
    description: dept.description || '',
    parent_id: dept.parent_id || '',
    manager_id: dept.manager_id || '',
    cost_center: dept.cost_center || '',
    is_active: dept.is_active
  }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editingDept.value = null
}

async function saveDept() {
  saving.value = true
  try {
    const payload = {
      ...form.value,
      parent_id: form.value.parent_id || null,
      manager_id: form.value.manager_id || null
    }
    if (editingDept.value) {
      await hrAPI.updateDepartment(editingDept.value.id, payload)
      store.addToast('Department updated', 'success')
    } else {
      await hrAPI.createDepartment(payload)
      store.addToast('Department created', 'success')
    }
    closeModal()
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

function confirmDeactivate(dept: Department) {
  deletingDept.value = dept
  showDeleteModal.value = true
}

async function deleteDept() {
  if (!deletingDept.value) return
  saving.value = true
  try {
    await hrAPI.deleteDepartment(deletingDept.value.id)
    store.addToast('Department deactivated', 'success')
    showDeleteModal.value = false
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Delete failed', 'error')
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
