<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Users, Plus, Search, Filter, X, Edit, Trash2, Eye, UserCheck, UserX,
  ChevronUp, ChevronDown, Download, RefreshCw, Phone, Mail, Building2,
  Briefcase, Calendar, CreditCard, MapPin, ChevronLeft, ChevronRight,
  ArrowUpDown, Save, AlertCircle, User
} from 'lucide-vue-next'
import { hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ── State ────────────────────────────────────────────────────────────────────
const employees   = ref<any[]>([])
const departments = ref<any[]>([])
const positions   = ref<any[]>([])
const loading     = ref(true)
const saving      = ref(false)

const search       = ref('')
const filterDept   = ref('')
const filterStatus = ref('')
const sortKey      = ref('last_name')
const sortDir      = ref<'asc'|'desc'>('asc')
const page         = ref(1)
const pageSize     = ref(20)

const showModal    = ref(false)
const showDetail   = ref(false)
const isEditing    = ref(false)
const selected     = ref<any>(null)
const confirmDeactivate = ref<any>(null)

const form = ref({
  first_name: '', last_name: '', email: '', phone: '', gender: '',
  birth_date: '', hire_date: '', national_id: '', cnas_number: '', nif: '',
  department_id: '', position_id: '', manager_id: '', employment_type: 'permanent',
  base_salary: 0, bank_name: '', bank_account: '',
  address: '', city: '', wilaya: '', notes: '', status: 'active'
})

// ── Load ─────────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const [empRes, deptRes, posRes] = await Promise.all([
      hrAPI.getEmployees(),
      hrAPI.getDepartments(),
      hrAPI.getPositions()
    ])
    employees.value   = empRes.data   || []
    departments.value = deptRes.data  || []
    positions.value   = posRes.data   || []
  } catch { app.addToast('Failed to load employees', 'error') }
  finally { loading.value = false }
}
onMounted(load)

// ── KPIs ─────────────────────────────────────────────────────────────────────
const kpis = computed(() => {
  const all     = employees.value
  const active  = all.filter(e => e.status === 'active').length
  const onLeave = all.filter(e => e.status === 'on_leave').length
  const avgSal  = all.length ? all.reduce((s, e) => s + (e.base_salary || 0), 0) / all.length : 0
  return [
    { label: 'Total Employees', value: all.length,  icon: Users,     color: 'blue'   },
    { label: 'Active',          value: active,       icon: UserCheck, color: 'green'  },
    { label: 'On Leave',        value: onLeave,      icon: UserX,     color: 'amber'  },
    { label: 'Avg. Salary',     value: fmt(avgSal),  icon: CreditCard,color: 'violet' },
  ]
})

// ── Computed table ────────────────────────────────────────────────────────────
const filtered = computed(() => {
  let data = [...employees.value]
  if (search.value)
    data = data.filter(e =>
      (e.full_name||'').toLowerCase().includes(search.value.toLowerCase()) ||
      (e.employee_number||'').toLowerCase().includes(search.value.toLowerCase()) ||
      (e.email||'').toLowerCase().includes(search.value.toLowerCase()))
  if (filterDept.value)
    data = data.filter(e => e.department_id === filterDept.value)
  if (filterStatus.value)
    data = data.filter(e => e.status === filterStatus.value)
  data.sort((a, b) => {
    const av = a[sortKey.value] ?? ''
    const bv = b[sortKey.value] ?? ''
    return sortDir.value === 'asc'
      ? String(av).localeCompare(String(bv))
      : String(bv).localeCompare(String(av))
  })
  return data
})
const totalPages = computed(() => Math.ceil(filtered.value.length / pageSize.value) || 1)
const paginated  = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return filtered.value.slice(start, start + pageSize.value)
})

function sort(key: string) {
  if (sortKey.value === key) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortKey.value = key; sortDir.value = 'asc' }
}

// ── Modal helpers ─────────────────────────────────────────────────────────────
function openCreate() {
  isEditing.value = false
  form.value = {
    first_name: '', last_name: '', email: '', phone: '', gender: 'male',
    birth_date: '', hire_date: new Date().toISOString().slice(0,10),
    national_id: '', cnas_number: '', nif: '',
    department_id: '', position_id: '', manager_id: '', employment_type: 'permanent',
    base_salary: 0, bank_name: '', bank_account: '',
    address: '', city: '', wilaya: '', notes: '', status: 'active'
  }
  showModal.value = true
}
function openEdit(emp: any) {
  isEditing.value = true
  form.value = {
    first_name: emp.first_name, last_name: emp.last_name,
    email: emp.email || '', phone: emp.phone || '', gender: emp.gender || 'male',
    birth_date: emp.birth_date ? emp.birth_date.slice(0,10) : '',
    hire_date: emp.hire_date ? emp.hire_date.slice(0,10) : '',
    national_id: emp.national_id || '', cnas_number: emp.cnas_number || '', nif: emp.nif || '',
    department_id: emp.department_id || '', position_id: emp.position_id || '',
    manager_id: emp.manager_id || '', employment_type: emp.employment_type || 'permanent',
    base_salary: emp.base_salary || 0,
    bank_name: emp.bank_name || '', bank_account: emp.bank_account || '',
    address: emp.address || '', city: emp.city || '', wilaya: emp.wilaya || '',
    notes: emp.notes || '', status: emp.status || 'active'
  }
  selected.value = emp
  showModal.value = true
}
function openDetail(emp: any) { selected.value = emp; showDetail.value = true }

async function save() {
  if (!form.value.first_name || !form.value.last_name || !form.value.hire_date) {
    app.addToast('First name, last name and hire date are required', 'error'); return
  }
  saving.value = true
  try {
    const payload = { ...form.value,
      department_id: form.value.department_id || null,
      position_id:   form.value.position_id   || null,
      manager_id:    form.value.manager_id     || null,
      birth_date:    form.value.birth_date     || null,
    }
    if (isEditing.value) {
      await hrAPI.updateEmployee(selected.value.id, payload)
      app.addToast('Employee updated', 'success')
    } else {
      await hrAPI.createEmployee(payload)
      app.addToast('Employee created', 'success')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally { saving.value = false }
}

async function confirmDel(emp: any) { confirmDeactivate.value = emp }
async function doDeactivate() {
  if (!confirmDeactivate.value) return
  try {
    await hrAPI.deleteEmployee(confirmDeactivate.value.id)
    app.addToast('Employee deactivated', 'success')
    confirmDeactivate.value = null
    await load()
  } catch { app.addToast('Deactivation failed', 'error') }
}

// ── Helpers ───────────────────────────────────────────────────────────────────
function fmt(n: number) {
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0 }).format(Math.round(n)) + ' DA'
}
function fmtDate(d: string) {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('fr-FR')
}
function statusColor(s: string) {
  return s === 'active'   ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' :
         s === 'on_leave' ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400' :
         s === 'inactive' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' :
                            'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300'
}
function kpiColor(c: string) {
  return c === 'blue'   ? 'bg-blue-500'   :
         c === 'green'  ? 'bg-emerald-500' :
         c === 'amber'  ? 'bg-amber-500'  :
         c === 'violet' ? 'bg-violet-500' : 'bg-slate-500'
}
function kpiBg(c: string) {
  return c === 'blue'   ? 'bg-blue-50 dark:bg-blue-900/20'   :
         c === 'green'  ? 'bg-emerald-50 dark:bg-emerald-900/20' :
         c === 'amber'  ? 'bg-amber-50 dark:bg-amber-900/20'  :
         c === 'violet' ? 'bg-violet-50 dark:bg-violet-900/20' : ''
}
const columns = [
  { key:'employee_number', label:'#'           },
  { key:'full_name',       label:'Name'        },
  { key:'department_name', label:'Department'  },
  { key:'position_title',  label:'Position'    },
  { key:'employment_type', label:'Type'        },
  { key:'base_salary',     label:'Salary'      },
  { key:'hire_date',       label:'Hire Date'   },
  { key:'status',          label:'Status'      },
]
</script>

<template>
  <div class="space-y-6 p-6">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white">Employees</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Manage your workforce — {{ employees.length }} total
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="load"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-400 transition-colors">
          <RefreshCw class="w-4 h-4" :class="loading && 'animate-spin'" />
        </button>
        <button @click="openCreate"
          class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors shadow-sm">
          <Plus class="w-4 h-4" />
          New Employee
        </button>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="k in kpis" :key="k.label"
        class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4 flex items-center gap-4">
        <div class="p-3 rounded-xl" :class="kpiBg(k.color)">
          <component :is="k.icon" class="w-6 h-6" :class="'text-'+k.color+'-600 dark:text-'+k.color+'-400'" />
        </div>
        <div>
          <p class="text-xs text-slate-500 dark:text-slate-400 font-medium uppercase tracking-wide">{{ k.label }}</p>
          <p class="text-2xl font-bold text-slate-900 dark:text-white mt-0.5">{{ k.value }}</p>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
      <div class="flex flex-wrap gap-3">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="search" placeholder="Search by name, number, email…"
            class="w-full pl-9 pr-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700 text-sm text-slate-900 dark:text-white placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>
        <select v-model="filterDept"
          class="px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700 text-sm text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-2 focus:ring-blue-500 min-w-40">
          <option value="">All Departments</option>
          <option v-for="d in departments" :key="d.id" :value="d.id">{{ d.name }}</option>
        </select>
        <select v-model="filterStatus"
          class="px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700 text-sm text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-2 focus:ring-blue-500">
          <option value="">All Status</option>
          <option value="active">Active</option>
          <option value="on_leave">On Leave</option>
          <option value="inactive">Inactive</option>
          <option value="terminated">Terminated</option>
        </select>
        <button v-if="search||filterDept||filterStatus"
          @click="search='';filterDept='';filterStatus=''"
          class="flex items-center gap-1 px-3 py-2 text-sm text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors">
          <X class="w-3 h-3" /> Clear
        </button>
        <span class="ml-auto text-xs text-slate-400 dark:text-slate-500 self-center">
          {{ filtered.length }} result{{ filtered.length !== 1 ? 's' : '' }}
        </span>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="w-8 h-8 animate-spin text-blue-500" />
      </div>
      <div v-else-if="!paginated.length" class="flex flex-col items-center justify-center py-20 text-slate-400 dark:text-slate-500">
        <Users class="w-12 h-12 mb-3 opacity-40" />
        <p class="font-medium">No employees found</p>
        <p class="text-sm mt-1">Try adjusting your search or filters</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-700">
            <tr>
              <th v-for="col in columns" :key="col.key"
                @click="sort(col.key)"
                class="text-left px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide cursor-pointer hover:text-slate-900 dark:hover:text-white select-none whitespace-nowrap">
                <div class="flex items-center gap-1">
                  {{ col.label }}
                  <ChevronUp v-if="sortKey===col.key && sortDir==='asc'" class="w-3 h-3 text-blue-500" />
                  <ChevronDown v-else-if="sortKey===col.key && sortDir==='desc'" class="w-3 h-3 text-blue-500" />
                  <ArrowUpDown v-else class="w-3 h-3 opacity-30" />
                </div>
              </th>
              <th class="px-4 py-3 text-right text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700/50">
            <tr v-for="emp in paginated" :key="emp.id"
              class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
              <td class="px-4 py-3 font-mono text-xs text-slate-500 dark:text-slate-400">{{ emp.employee_number }}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-gradient-to-br from-blue-500 to-violet-600 flex items-center justify-center text-white text-xs font-bold flex-shrink-0">
                    {{ (emp.first_name?.[0]||'') + (emp.last_name?.[0]||'') }}
                  </div>
                  <div>
                    <p class="font-semibold text-slate-900 dark:text-white">{{ emp.full_name }}</p>
                    <p class="text-xs text-slate-400">{{ emp.email }}</p>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3 text-slate-700 dark:text-slate-300">{{ emp.department_name || '-' }}</td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400">{{ emp.position_title || '-' }}</td>
              <td class="px-4 py-3">
                <span class="px-2 py-0.5 rounded text-xs font-medium bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300 capitalize">
                  {{ emp.employment_type?.replace('_',' ') }}
                </span>
              </td>
              <td class="px-4 py-3 font-medium text-slate-800 dark:text-slate-200">{{ fmt(emp.base_salary||0) }}</td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400">{{ fmtDate(emp.hire_date) }}</td>
              <td class="px-4 py-3">
                <span class="px-2 py-1 rounded-full text-xs font-semibold capitalize" :class="statusColor(emp.status)">
                  {{ emp.status?.replace('_',' ') }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center justify-end gap-1">
                  <button @click="openDetail(emp)" title="View"
                    class="p-1.5 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
                    <Eye class="w-4 h-4" />
                  </button>
                  <button @click="openEdit(emp)" title="Edit"
                    class="p-1.5 rounded-lg hover:bg-amber-50 dark:hover:bg-amber-900/20 text-slate-400 hover:text-amber-600 dark:hover:text-amber-400 transition-colors">
                    <Edit class="w-4 h-4" />
                  </button>
                  <button @click="confirmDel(emp)" title="Deactivate"
                    class="p-1.5 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 text-slate-400 hover:text-red-600 dark:hover:text-red-400 transition-colors">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="!loading && filtered.length > pageSize"
        class="flex items-center justify-between px-4 py-3 border-t border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/30">
        <p class="text-xs text-slate-500 dark:text-slate-400">
          Showing {{ (page-1)*pageSize+1 }}–{{ Math.min(page*pageSize, filtered.length) }} of {{ filtered.length }}
        </p>
        <div class="flex items-center gap-1">
          <button @click="page=Math.max(1,page-1)" :disabled="page===1"
            class="p-1.5 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed transition-colors">
            <ChevronLeft class="w-4 h-4 text-slate-600 dark:text-slate-400" />
          </button>
          <span class="px-3 py-1 text-xs font-medium text-slate-700 dark:text-slate-300">
            {{ page }} / {{ totalPages }}
          </span>
          <button @click="page=Math.min(totalPages,page+1)" :disabled="page===totalPages"
            class="p-1.5 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed transition-colors">
            <ChevronRight class="w-4 h-4 text-slate-600 dark:text-slate-400" />
          </button>
        </div>
      </div>
    </div>

    <!-- Create / Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-3xl max-h-[90vh] flex flex-col">
          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-lg bg-blue-100 dark:bg-blue-900/30">
                <User class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              </div>
              <h2 class="text-lg font-bold text-slate-900 dark:text-white">
                {{ isEditing ? 'Edit Employee' : 'New Employee' }}
              </h2>
            </div>
            <button @click="showModal=false"
              class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <!-- Body -->
          <div class="overflow-y-auto p-6 space-y-5">
            <!-- Personal -->
            <div>
              <h3 class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-3">Personal Information</h3>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">First Name <span class="text-red-500">*</span></label>
                  <input v-model="form.first_name" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Last Name <span class="text-red-500">*</span></label>
                  <input v-model="form.last_name" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Gender</label>
                  <select v-model="form.gender" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="male">Male</option>
                    <option value="female">Female</option>
                  </select>
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Date of Birth</label>
                  <input v-model="form.birth_date" type="date" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">National ID</label>
                  <input v-model="form.national_id" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">CNAS Number</label>
                  <input v-model="form.cnas_number" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
              </div>
            </div>
            <!-- Contact -->
            <div>
              <h3 class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-3">Contact</h3>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Email</label>
                  <input v-model="form.email" type="email" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Phone</label>
                  <input v-model="form.phone" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">City</label>
                  <input v-model="form.city" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Wilaya</label>
                  <input v-model="form.wilaya" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
              </div>
            </div>
            <!-- Employment -->
            <div>
              <h3 class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-3">Employment</h3>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Hire Date <span class="text-red-500">*</span></label>
                  <input v-model="form.hire_date" type="date" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Employment Type</label>
                  <select v-model="form.employment_type" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="permanent">Permanent</option>
                    <option value="contract">Contract</option>
                    <option value="part_time">Part-time</option>
                    <option value="intern">Intern</option>
                  </select>
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Department</label>
                  <select v-model="form.department_id" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">— Select —</option>
                    <option v-for="d in departments" :key="d.id" :value="d.id">{{ d.name }}</option>
                  </select>
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Position</label>
                  <select v-model="form.position_id" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="">— Select —</option>
                    <option v-for="p in positions" :key="p.id" :value="p.id">{{ p.title }}</option>
                  </select>
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Base Salary (DA)</label>
                  <input v-model.number="form.base_salary" type="number" min="0" step="1000" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div v-if="isEditing">
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Status</label>
                  <select v-model="form.status" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
                    <option value="active">Active</option>
                    <option value="on_leave">On Leave</option>
                    <option value="inactive">Inactive</option>
                    <option value="terminated">Terminated</option>
                  </select>
                </div>
              </div>
            </div>
            <!-- Bank -->
            <div>
              <h3 class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-3">Banking</h3>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Bank Name</label>
                  <input v-model="form.bank_name" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">RIB / Account</label>
                  <input v-model="form.bank_account" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>
              </div>
            </div>
            <!-- Notes -->
            <div>
              <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Notes</label>
              <textarea v-model="form.notes" rows="2" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none" />
            </div>
          </div>
          <!-- Footer -->
          <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-700">
            <button @click="showModal=false"
              class="px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-600 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
              Cancel
            </button>
            <button @click="save" :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors">
              <Save class="w-4 h-4" />
              {{ saving ? 'Saving…' : (isEditing ? 'Update' : 'Create') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Detail Drawer -->
    <Teleport to="body">
      <div v-if="showDetail && selected" class="fixed inset-0 z-50 flex">
        <div class="flex-1 bg-black/40 backdrop-blur-sm" @click="showDetail=false" />
        <div class="w-full max-w-md bg-white dark:bg-slate-800 shadow-2xl overflow-y-auto">
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700 sticky top-0 bg-white dark:bg-slate-800">
            <h2 class="text-lg font-bold text-slate-900 dark:text-white">Employee Profile</h2>
            <button @click="showDetail=false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-400 transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="p-6 space-y-6">
            <!-- Avatar block -->
            <div class="flex items-center gap-4">
              <div class="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-violet-600 flex items-center justify-center text-white text-2xl font-bold">
                {{ (selected.first_name?.[0]||'') + (selected.last_name?.[0]||'') }}
              </div>
              <div>
                <p class="text-xl font-bold text-slate-900 dark:text-white">{{ selected.full_name }}</p>
                <p class="text-sm text-slate-500">{{ selected.employee_number }}</p>
                <span class="inline-block mt-1 px-2 py-0.5 rounded-full text-xs font-semibold capitalize" :class="statusColor(selected.status)">{{ selected.status?.replace('_',' ') }}</span>
              </div>
            </div>

            <div class="grid grid-cols-1 gap-3">
              <div class="flex items-center gap-3 p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50">
                <Building2 class="w-4 h-4 text-slate-400 flex-shrink-0" />
                <div>
                  <p class="text-xs text-slate-400">Department / Position</p>
                  <p class="text-sm font-medium text-slate-800 dark:text-slate-200">
                    {{ selected.department_name || '-' }} / {{ selected.position_title || '-' }}
                  </p>
                </div>
              </div>
              <div class="flex items-center gap-3 p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50">
                <Mail class="w-4 h-4 text-slate-400 flex-shrink-0" />
                <div>
                  <p class="text-xs text-slate-400">Email</p>
                  <p class="text-sm font-medium text-slate-800 dark:text-slate-200">{{ selected.email || '-' }}</p>
                </div>
              </div>
              <div class="flex items-center gap-3 p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50">
                <Phone class="w-4 h-4 text-slate-400 flex-shrink-0" />
                <div>
                  <p class="text-xs text-slate-400">Phone</p>
                  <p class="text-sm font-medium text-slate-800 dark:text-slate-200">{{ selected.phone || '-' }}</p>
                </div>
              </div>
              <div class="flex items-center gap-3 p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50">
                <Calendar class="w-4 h-4 text-slate-400 flex-shrink-0" />
                <div>
                  <p class="text-xs text-slate-400">Hire Date</p>
                  <p class="text-sm font-medium text-slate-800 dark:text-slate-200">{{ fmtDate(selected.hire_date) }}</p>
                </div>
              </div>
              <div class="flex items-center gap-3 p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50">
                <CreditCard class="w-4 h-4 text-slate-400 flex-shrink-0" />
                <div>
                  <p class="text-xs text-slate-400">Base Salary</p>
                  <p class="text-sm font-bold text-slate-800 dark:text-slate-200">{{ fmt(selected.base_salary||0) }}</p>
                </div>
              </div>
              <div class="flex items-center gap-3 p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50">
                <Briefcase class="w-4 h-4 text-slate-400 flex-shrink-0" />
                <div>
                  <p class="text-xs text-slate-400">Employment Type</p>
                  <p class="text-sm font-medium text-slate-800 dark:text-slate-200 capitalize">{{ selected.employment_type?.replace('_',' ') }}</p>
                </div>
              </div>
              <div v-if="selected.national_id || selected.cnas_number" class="flex items-center gap-3 p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50">
                <User class="w-4 h-4 text-slate-400 flex-shrink-0" />
                <div>
                  <p class="text-xs text-slate-400">National ID / CNAS</p>
                  <p class="text-sm font-medium text-slate-800 dark:text-slate-200">
                    {{ selected.national_id || '-' }} / {{ selected.cnas_number || '-' }}
                  </p>
                </div>
              </div>
              <div v-if="selected.bank_name" class="flex items-center gap-3 p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50">
                <CreditCard class="w-4 h-4 text-slate-400 flex-shrink-0" />
                <div>
                  <p class="text-xs text-slate-400">Bank / RIB</p>
                  <p class="text-sm font-medium text-slate-800 dark:text-slate-200">{{ selected.bank_name }} — {{ selected.bank_account }}</p>
                </div>
              </div>
            </div>

            <div class="flex gap-2 pt-2">
              <button @click="openEdit(selected); showDetail=false"
                class="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors">
                <Edit class="w-4 h-4" /> Edit
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Deactivate Confirm -->
    <Teleport to="body">
      <div v-if="confirmDeactivate" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-sm p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="p-2 rounded-lg bg-red-100 dark:bg-red-900/30">
              <AlertCircle class="w-5 h-5 text-red-600 dark:text-red-400" />
            </div>
            <h3 class="text-lg font-bold text-slate-900 dark:text-white">Deactivate Employee</h3>
          </div>
          <p class="text-sm text-slate-600 dark:text-slate-400 mb-6">
            Are you sure you want to deactivate <strong>{{ confirmDeactivate.full_name }}</strong>?
            Their record will be preserved but they will no longer appear as active.
          </p>
          <div class="flex gap-3">
            <button @click="confirmDeactivate=null"
              class="flex-1 px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-600 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
              Cancel
            </button>
            <button @click="doDeactivate"
              class="flex-1 px-4 py-2 rounded-lg bg-red-600 hover:bg-red-700 text-white text-sm font-medium transition-colors">
              Deactivate
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
