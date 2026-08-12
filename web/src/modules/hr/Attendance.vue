<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Clock, Plus, RefreshCw, Search, X, Filter, ChevronUp, ChevronDown,
  ArrowUpDown, Edit, Save, CalendarCheck, CheckCircle, XCircle,
  AlertTriangle, Users, BarChart2
} from 'lucide-vue-next'
import { hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const records     = ref<any[]>([])
const summary     = ref<any[]>([])
const employees   = ref<any[]>([])
const loading     = ref(true)
const saving      = ref(false)

const viewMode    = ref<'records'|'summary'>('records')
const filterMonth = ref(String(new Date().getMonth() + 1).padStart(2,'0'))
const filterYear  = ref(String(new Date().getFullYear()))
const filterEmp   = ref('')
const search      = ref('')
const sortKey     = ref('date')
const sortDir     = ref<'asc'|'desc'>('desc')
const showModal   = ref(false)
const isEditing   = ref(false)
const selRec      = ref<any>(null)

const form = ref({
  employee_id: '', date: new Date().toISOString().slice(0,10),
  check_in: '', check_out: '', hours_worked: 0, overtime_hours: 0,
  status: 'present', notes: ''
})

async function load() {
  loading.value = true
  try {
    const params: Record<string,string> = {}
    if (filterMonth.value) params.month = filterMonth.value
    if (filterYear.value)  params.year  = filterYear.value
    if (filterEmp.value)   params.employee_id = filterEmp.value

    const [recRes, empRes] = await Promise.all([
      hrAPI.getAttendance(params),
      hrAPI.getEmployees()
    ])
    records.value   = recRes.data  || []
    employees.value = empRes.data  || []

    if (viewMode.value === 'summary') await loadSummary()
  } catch { app.addToast('Failed to load attendance', 'error') }
  finally { loading.value = false }
}

async function loadSummary() {
  try {
    const res = await hrAPI.getAttendanceSummary({ month: filterMonth.value, year: filterYear.value })
    summary.value = res.data || []
  } catch { summary.value = [] }
}

onMounted(load)

const kpis = computed(() => {
  const r       = records.value
  const present = r.filter(x => x.status === 'present').length
  const absent  = r.filter(x => x.status === 'absent').length
  const late    = r.filter(x => x.status === 'late').length
  const totalH  = r.reduce((s, x) => s + (x.hours_worked || 0), 0)
  return [
    { label: 'Present', value: present, icon: CheckCircle,  color: 'green'  },
    { label: 'Absent',  value: absent,  icon: XCircle,      color: 'red'    },
    { label: 'Late',    value: late,    icon: AlertTriangle, color: 'amber'  },
    { label: 'Total Hours', value: totalH.toFixed(1)+'h', icon: Clock, color: 'blue' },
  ]
})

const filtered = computed(() => {
  let d = [...records.value]
  if (search.value)
    d = d.filter(r => (r.employee_name||'').toLowerCase().includes(search.value.toLowerCase()) ||
                      (r.employee_number||'').toLowerCase().includes(search.value.toLowerCase()))
  d.sort((a,b) => {
    const av = a[sortKey.value]??''; const bv = b[sortKey.value]??''
    return sortDir.value==='asc' ? String(av).localeCompare(String(bv)) : String(bv).localeCompare(String(av))
  })
  return d
})

function sort(k: string) {
  if (sortKey.value===k) sortDir.value=sortDir.value==='asc'?'desc':'asc'
  else { sortKey.value=k; sortDir.value='asc' }
}

function openCreate() {
  isEditing.value = false; selRec.value = null
  form.value = { employee_id:'', date: new Date().toISOString().slice(0,10),
    check_in:'', check_out:'', hours_worked:0, overtime_hours:0, status:'present', notes:'' }
  showModal.value = true
}
function openEdit(r: any) {
  isEditing.value = true; selRec.value = r
  form.value = {
    employee_id: r.employee_id, date: r.date?.slice(0,10)||'',
    check_in: r.check_in ? new Date(r.check_in).toISOString().slice(11,16) : '',
    check_out: r.check_out ? new Date(r.check_out).toISOString().slice(11,16) : '',
    hours_worked: r.hours_worked||0, overtime_hours: r.overtime_hours||0,
    status: r.status||'present', notes: r.notes||''
  }
  showModal.value = true
}

async function save() {
  if (!form.value.employee_id || !form.value.date) {
    app.addToast('Employee and date are required', 'error'); return
  }
  saving.value = true
  try {
    const payload = {
      ...form.value,
      check_in:  form.value.check_in  ? `${form.value.date}T${form.value.check_in}:00` : null,
      check_out: form.value.check_out ? `${form.value.date}T${form.value.check_out}:00` : null,
    }
    if (isEditing.value && selRec.value) {
      await hrAPI.updateAttendance(selRec.value.id, payload)
      app.addToast('Attendance updated', 'success')
    } else {
      await hrAPI.recordAttendance(payload)
      app.addToast('Attendance recorded', 'success')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally { saving.value = false }
}

function fmtTime(t: string|null) {
  if (!t) return '-'
  return new Date(t).toLocaleTimeString('fr-FR', { hour:'2-digit', minute:'2-digit' })
}
function fmtDate(d: string) { return d ? new Date(d).toLocaleDateString('fr-FR') : '-' }
function statusColor(s: string) {
  return s==='present' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' :
         s==='absent'  ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' :
         s==='late'    ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400' :
         s==='half_day'? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' :
                         'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300'
}
function kpiBg(c: string) {
  return c==='green'?'bg-emerald-50 dark:bg-emerald-900/20':
         c==='red'?'bg-red-50 dark:bg-red-900/20':
         c==='amber'?'bg-amber-50 dark:bg-amber-900/20':
         'bg-blue-50 dark:bg-blue-900/20'
}
function kpiText(c: string) {
  return c==='green'?'text-emerald-600 dark:text-emerald-400':
         c==='red'?'text-red-600 dark:text-red-400':
         c==='amber'?'text-amber-600 dark:text-amber-400':
         'text-blue-600 dark:text-blue-400'
}

const months = [
  {value:'1',label:'January'},{value:'2',label:'February'},{value:'3',label:'March'},
  {value:'4',label:'April'},{value:'5',label:'May'},{value:'6',label:'June'},
  {value:'7',label:'July'},{value:'8',label:'August'},{value:'9',label:'September'},
  {value:'10',label:'October'},{value:'11',label:'November'},{value:'12',label:'December'},
]
</script>

<template>
  <div class="space-y-6 p-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white">Attendance</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Track employee presence and working hours</p>
      </div>
      <div class="flex items-center gap-2">
        <div class="flex bg-slate-100 dark:bg-slate-700 rounded-lg p-1 gap-1">
          <button @click="viewMode='records'; load()"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-colors"
            :class="viewMode==='records'?'bg-white dark:bg-slate-600 text-slate-900 dark:text-white shadow':'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200'">
            Records
          </button>
          <button @click="viewMode='summary'; loadSummary()"
            class="px-3 py-1.5 rounded-md text-xs font-semibold transition-colors"
            :class="viewMode==='summary'?'bg-white dark:bg-slate-600 text-slate-900 dark:text-white shadow':'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200'">
            Summary
          </button>
        </div>
        <button @click="load" class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-400 transition-colors">
          <RefreshCw class="w-4 h-4" :class="loading&&'animate-spin'" />
        </button>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors shadow-sm">
          <Plus class="w-4 h-4" /> Record Attendance
        </button>
      </div>
    </div>

    <!-- KPIs -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="k in kpis" :key="k.label" class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4 flex items-center gap-4">
        <div class="p-3 rounded-xl" :class="kpiBg(k.color)">
          <component :is="k.icon" class="w-6 h-6" :class="kpiText(k.color)" />
        </div>
        <div>
          <p class="text-xs text-slate-500 dark:text-slate-400 font-medium uppercase tracking-wide">{{ k.label }}</p>
          <p class="text-2xl font-bold text-slate-900 dark:text-white mt-0.5">{{ k.value }}</p>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4 flex flex-wrap gap-3">
      <select v-model="filterMonth" @change="load"
        class="px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700 text-sm text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-2 focus:ring-blue-500">
        <option v-for="m in months" :key="m.value" :value="m.value">{{ m.label }}</option>
      </select>
      <select v-model="filterYear" @change="load"
        class="px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700 text-sm text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-2 focus:ring-blue-500">
        <option v-for="y in [2024,2025,2026]" :key="y" :value="String(y)">{{ y }}</option>
      </select>
      <select v-model="filterEmp" @change="load"
        class="px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700 text-sm text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-2 focus:ring-blue-500 flex-1 min-w-40">
        <option value="">All Employees</option>
        <option v-for="e in employees" :key="e.id" :value="e.id">{{ e.full_name }}</option>
      </select>
      <div class="relative flex-1 min-w-48">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input v-model="search" placeholder="Search employee…"
          class="w-full pl-9 pr-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700 text-sm text-slate-900 dark:text-white placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500" />
      </div>
    </div>

    <!-- Records Table -->
    <div v-if="viewMode==='records'" class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="w-8 h-8 animate-spin text-blue-500" />
      </div>
      <div v-else-if="!filtered.length" class="flex flex-col items-center py-20 text-slate-400 dark:text-slate-500">
        <Clock class="w-12 h-12 mb-3 opacity-40" />
        <p class="font-medium">No attendance records found</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-700">
            <tr>
              <th v-for="col in [{key:'date',label:'Date'},{key:'employee_name',label:'Employee'},{key:'department_name',label:'Dept'},{key:'check_in',label:'Check In'},{key:'check_out',label:'Check Out'},{key:'hours_worked',label:'Hours'},{key:'overtime_hours',label:'OT'},{key:'status',label:'Status'}]"
                :key="col.key" @click="sort(col.key)"
                class="text-left px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide cursor-pointer hover:text-slate-900 dark:hover:text-white select-none whitespace-nowrap">
                <div class="flex items-center gap-1">
                  {{ col.label }}
                  <ChevronUp v-if="sortKey===col.key&&sortDir==='asc'" class="w-3 h-3 text-blue-500" />
                  <ChevronDown v-else-if="sortKey===col.key&&sortDir==='desc'" class="w-3 h-3 text-blue-500" />
                  <ArrowUpDown v-else class="w-3 h-3 opacity-30" />
                </div>
              </th>
              <th class="px-4 py-3 text-right text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700/50">
            <tr v-for="r in filtered" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
              <td class="px-4 py-3 font-medium text-slate-800 dark:text-slate-200 whitespace-nowrap">{{ fmtDate(r.date) }}</td>
              <td class="px-4 py-3">
                <div>
                  <p class="font-semibold text-slate-900 dark:text-white">{{ r.employee_name }}</p>
                  <p class="text-xs text-slate-400">{{ r.employee_number }}</p>
                </div>
              </td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400 text-xs">{{ r.department_name }}</td>
              <td class="px-4 py-3 font-mono text-slate-700 dark:text-slate-300">{{ fmtTime(r.check_in) }}</td>
              <td class="px-4 py-3 font-mono text-slate-700 dark:text-slate-300">{{ fmtTime(r.check_out) }}</td>
              <td class="px-4 py-3 font-semibold text-slate-800 dark:text-slate-200">{{ r.hours_worked?.toFixed(1) || '-' }}h</td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400">{{ r.overtime_hours > 0 ? r.overtime_hours?.toFixed(1)+'h' : '-' }}</td>
              <td class="px-4 py-3">
                <span class="px-2 py-1 rounded-full text-xs font-semibold capitalize" :class="statusColor(r.status)">
                  {{ r.status?.replace('_',' ') }}
                </span>
              </td>
              <td class="px-4 py-3 text-right">
                <button @click="openEdit(r)" class="p-1.5 rounded-lg hover:bg-amber-50 dark:hover:bg-amber-900/20 text-slate-400 hover:text-amber-600 transition-colors">
                  <Edit class="w-4 h-4" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Summary Table -->
    <div v-if="viewMode==='summary'" class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      <div class="px-4 py-3 border-b border-slate-200 dark:border-slate-700 flex items-center gap-2">
        <BarChart2 class="w-4 h-4 text-blue-500" />
        <span class="text-sm font-semibold text-slate-700 dark:text-slate-300">Monthly Summary — {{ months.find(m=>m.value===filterMonth)?.label }} {{ filterYear }}</span>
      </div>
      <div v-if="!summary.length" class="flex flex-col items-center py-16 text-slate-400">
        <Users class="w-10 h-10 mb-2 opacity-40" />
        <p class="text-sm">No summary data available</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-700">
            <tr>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">Employee</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">Dept</th>
              <th class="text-center px-3 py-3 text-xs font-semibold text-emerald-600 dark:text-emerald-400 uppercase tracking-wide">Present</th>
              <th class="text-center px-3 py-3 text-xs font-semibold text-red-600 dark:text-red-400 uppercase tracking-wide">Absent</th>
              <th class="text-center px-3 py-3 text-xs font-semibold text-amber-600 dark:text-amber-400 uppercase tracking-wide">Late</th>
              <th class="text-center px-3 py-3 text-xs font-semibold text-blue-600 dark:text-blue-400 uppercase tracking-wide">Half</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">Total Hrs</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-violet-600 dark:text-violet-400 uppercase tracking-wide">OT Hrs</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700/50">
            <tr v-for="s in summary" :key="s.employee_id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
              <td class="px-4 py-3">
                <p class="font-semibold text-slate-900 dark:text-white">{{ s.full_name }}</p>
                <p class="text-xs text-slate-400">{{ s.employee_number }}</p>
              </td>
              <td class="px-4 py-3 text-xs text-slate-600 dark:text-slate-400">{{ s.department_name }}</td>
              <td class="px-3 py-3 text-center font-bold text-emerald-600 dark:text-emerald-400">{{ s.present_days }}</td>
              <td class="px-3 py-3 text-center font-bold text-red-600 dark:text-red-400">{{ s.absent_days }}</td>
              <td class="px-3 py-3 text-center font-bold text-amber-600 dark:text-amber-400">{{ s.late_days }}</td>
              <td class="px-3 py-3 text-center font-bold text-blue-600 dark:text-blue-400">{{ s.half_days }}</td>
              <td class="px-4 py-3 text-right font-semibold text-slate-800 dark:text-slate-200">{{ Number(s.total_hours).toFixed(1) }}h</td>
              <td class="px-4 py-3 text-right font-semibold text-violet-600 dark:text-violet-400">{{ Number(s.overtime_hours).toFixed(1) }}h</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-lg">
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-lg bg-blue-100 dark:bg-blue-900/30">
                <CalendarCheck class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              </div>
              <h2 class="text-lg font-bold text-slate-900 dark:text-white">
                {{ isEditing ? 'Edit Attendance' : 'Record Attendance' }}
              </h2>
            </div>
            <button @click="showModal=false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-400 transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div>
              <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Employee <span class="text-red-500">*</span></label>
              <select v-model="form.employee_id" :disabled="isEditing"
                class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-60">
                <option value="">— Select employee —</option>
                <option v-for="e in employees" :key="e.id" :value="e.id">{{ e.full_name }}</option>
              </select>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Date <span class="text-red-500">*</span></label>
                <input v-model="form.date" type="date" :disabled="isEditing"
                  class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-60" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Status</label>
                <select v-model="form.status"
                  class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option value="present">Present</option>
                  <option value="absent">Absent</option>
                  <option value="late">Late</option>
                  <option value="half_day">Half Day</option>
                  <option value="holiday">Holiday</option>
                </select>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Check In</label>
                <input v-model="form.check_in" type="time"
                  class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Check Out</label>
                <input v-model="form.check_out" type="time"
                  class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Hours Worked</label>
                <input v-model.number="form.hours_worked" type="number" step="0.5" min="0" max="24"
                  class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Overtime Hours</label>
                <input v-model.number="form.overtime_hours" type="number" step="0.5" min="0" max="12"
                  class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Notes</label>
              <input v-model="form.notes"
                class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
            </div>
          </div>
          <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-700">
            <button @click="showModal=false" class="px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-600 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
            <button @click="save" :disabled="saving" class="flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors">
              <Save class="w-4 h-4" /> {{ saving ? 'Saving…' : (isEditing ? 'Update' : 'Record') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
