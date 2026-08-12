<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  CalendarOff, Plus, RefreshCw, X, Search, ChevronUp, ChevronDown,
  ArrowUpDown, CheckCircle, XCircle, Clock, AlertCircle, Save,
  Eye, Filter, Calendar
} from 'lucide-vue-next'
import { hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const requests    = ref<any[]>([])
const employees   = ref<any[]>([])
const leaveTypes  = ref<any[]>([])
const loading     = ref(true)
const saving      = ref(false)

const filterStatus = ref('')
const search       = ref('')
const sortKey      = ref('created_at')
const sortDir      = ref<'asc'|'desc'>('desc')
const showModal    = ref(false)
const showDetail   = ref(false)
const selected     = ref<any>(null)
const rejectModal  = ref<any>(null)
const rejectReason = ref('')

const form = ref({
  employee_id: '', leave_type_id: '',
  start_date: '', end_date: '', days_count: 1, reason: ''
})

async function load() {
  loading.value = true
  try {
    const params: Record<string,string> = {}
    if (filterStatus.value) params.status = filterStatus.value
    const [reqRes, empRes, ltRes] = await Promise.all([
      hrAPI.getLeaveRequests(params),
      hrAPI.getEmployees(),
      hrAPI.getLeaveTypes()
    ])
    requests.value   = reqRes.data || []
    employees.value  = empRes.data || []
    leaveTypes.value = ltRes.data  || []
  } catch { app.addToast('Failed to load leave requests', 'error') }
  finally { loading.value = false }
}
onMounted(load)

const kpis = computed(() => {
  const all      = requests.value
  const pending  = all.filter(r => r.status === 'pending').length
  const approved = all.filter(r => r.status === 'approved').length
  const rejected = all.filter(r => r.status === 'rejected').length
  const totalDays = all.filter(r => r.status === 'approved').reduce((s, r) => s + (r.days_count || 0), 0)
  return [
    { label: 'Total Requests', value: all.length,  icon: CalendarOff, color: 'blue'  },
    { label: 'Pending',        value: pending,      icon: Clock,       color: 'amber' },
    { label: 'Approved',       value: approved,     icon: CheckCircle, color: 'green' },
    { label: 'Total Days Off', value: totalDays+'d',icon: Calendar,    color: 'violet'},
  ]
})

const filtered = computed(() => {
  let d = [...requests.value]
  if (search.value)
    d = d.filter(r => (r.employee_name||'').toLowerCase().includes(search.value.toLowerCase()))
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
  form.value = { employee_id:'', leave_type_id:'', start_date:'', end_date:'', days_count:1, reason:'' }
  showModal.value = true
}

async function save() {
  if (!form.value.employee_id || !form.value.leave_type_id || !form.value.start_date || !form.value.end_date) {
    app.addToast('All required fields must be filled', 'error'); return
  }
  saving.value = true
  try {
    await hrAPI.createLeaveRequest(form.value)
    app.addToast('Leave request submitted', 'success')
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Failed to submit', 'error')
  } finally { saving.value = false }
}

async function approve(req: any) {
  try {
    await hrAPI.approveLeave(req.id)
    app.addToast('Leave request approved', 'success')
    await load()
  } catch { app.addToast('Failed to approve', 'error') }
}

async function doReject() {
  if (!rejectModal.value) return
  try {
    await hrAPI.rejectLeave(rejectModal.value.id, { reason: rejectReason.value })
    app.addToast('Leave request rejected', 'success')
    rejectModal.value = null; rejectReason.value = ''
    await load()
  } catch { app.addToast('Failed to reject', 'error') }
}

async function cancel(req: any) {
  try {
    await hrAPI.cancelLeave(req.id)
    app.addToast('Leave request cancelled', 'success')
    await load()
  } catch { app.addToast('Failed to cancel', 'error') }
}

// Auto-calculate days when dates change
function recalcDays() {
  if (form.value.start_date && form.value.end_date) {
    const d1 = new Date(form.value.start_date)
    const d2 = new Date(form.value.end_date)
    const diff = Math.floor((d2.getTime() - d1.getTime()) / 86400000) + 1
    form.value.days_count = diff > 0 ? diff : 1
  }
}

function fmtDate(d: string) { return d ? new Date(d).toLocaleDateString('fr-FR') : '-' }
function statusColor(s: string) {
  return s==='pending'  ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400' :
         s==='approved' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' :
         s==='rejected' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' :
                          'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300'
}
function kpiBg(c: string) {
  return c==='blue'?'bg-blue-50 dark:bg-blue-900/20':
         c==='amber'?'bg-amber-50 dark:bg-amber-900/20':
         c==='green'?'bg-emerald-50 dark:bg-emerald-900/20':
         'bg-violet-50 dark:bg-violet-900/20'
}
function kpiText(c: string) {
  return c==='blue'?'text-blue-600 dark:text-blue-400':
         c==='amber'?'text-amber-600 dark:text-amber-400':
         c==='green'?'text-emerald-600 dark:text-emerald-400':
         'text-violet-600 dark:text-violet-400'
}
</script>

<template>
  <div class="space-y-6 p-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white">Leave Requests</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Manage employee leave and absences</p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="load" class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-400 transition-colors">
          <RefreshCw class="w-4 h-4" :class="loading&&'animate-spin'" />
        </button>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors shadow-sm">
          <Plus class="w-4 h-4" /> New Request
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
      <div class="flex gap-1">
        <button v-for="s in [{v:'',l:'All'},{v:'pending',l:'Pending'},{v:'approved',l:'Approved'},{v:'rejected',l:'Rejected'}]"
          :key="s.v" @click="filterStatus=s.v; load()"
          class="px-3 py-1.5 rounded-lg text-xs font-semibold transition-colors"
          :class="filterStatus===s.v
            ? 'bg-blue-600 text-white shadow'
            : 'bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-200 dark:hover:bg-slate-600'">
          {{ s.l }}
        </button>
      </div>
      <div class="relative flex-1 min-w-48">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
        <input v-model="search" placeholder="Search employee…"
          class="w-full pl-9 pr-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-slate-50 dark:bg-slate-700 text-sm text-slate-900 dark:text-white placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500" />
      </div>
      <span class="text-xs text-slate-400 self-center ml-auto">{{ filtered.length }} request{{ filtered.length!==1?'s':'' }}</span>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="w-8 h-8 animate-spin text-blue-500" />
      </div>
      <div v-else-if="!filtered.length" class="flex flex-col items-center py-20 text-slate-400 dark:text-slate-500">
        <CalendarOff class="w-12 h-12 mb-3 opacity-40" />
        <p class="font-medium">No leave requests found</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-700">
            <tr>
              <th v-for="col in [{key:'employee_name',label:'Employee'},{key:'department_name',label:'Dept'},{key:'leave_type_name',label:'Leave Type'},{key:'start_date',label:'From'},{key:'end_date',label:'To'},{key:'days_count',label:'Days'},{key:'status',label:'Status'}]"
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
              <td class="px-4 py-3">
                <p class="font-semibold text-slate-900 dark:text-white">{{ r.employee_name }}</p>
                <p class="text-xs text-slate-400">{{ r.employee_number }}</p>
              </td>
              <td class="px-4 py-3 text-xs text-slate-600 dark:text-slate-400">{{ r.department_name }}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <span class="w-2.5 h-2.5 rounded-full flex-shrink-0" :style="{background: r.leave_color || '#6366f1'}" />
                  <span class="text-slate-700 dark:text-slate-300">{{ r.leave_type_name }}</span>
                  <span v-if="r.is_paid" class="text-xs text-emerald-600 dark:text-emerald-400 font-medium">(Paid)</span>
                </div>
              </td>
              <td class="px-4 py-3 text-slate-700 dark:text-slate-300 whitespace-nowrap">{{ fmtDate(r.start_date) }}</td>
              <td class="px-4 py-3 text-slate-700 dark:text-slate-300 whitespace-nowrap">{{ fmtDate(r.end_date) }}</td>
              <td class="px-4 py-3 font-bold text-slate-800 dark:text-slate-200 text-center">{{ r.days_count }}</td>
              <td class="px-4 py-3">
                <span class="px-2 py-1 rounded-full text-xs font-semibold capitalize" :class="statusColor(r.status)">{{ r.status }}</span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center justify-end gap-1">
                  <button @click="selected=r; showDetail=true" class="p-1.5 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 text-slate-400 hover:text-blue-600 transition-colors">
                    <Eye class="w-4 h-4" />
                  </button>
                  <template v-if="r.status==='pending'">
                    <button @click="approve(r)" class="p-1.5 rounded-lg hover:bg-emerald-50 dark:hover:bg-emerald-900/20 text-slate-400 hover:text-emerald-600 transition-colors" title="Approve">
                      <CheckCircle class="w-4 h-4" />
                    </button>
                    <button @click="rejectModal=r; rejectReason=''" class="p-1.5 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 text-slate-400 hover:text-red-600 transition-colors" title="Reject">
                      <XCircle class="w-4 h-4" />
                    </button>
                  </template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- New Request Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-lg">
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-lg bg-blue-100 dark:bg-blue-900/30">
                <CalendarOff class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              </div>
              <h2 class="text-lg font-bold text-slate-900 dark:text-white">New Leave Request</h2>
            </div>
            <button @click="showModal=false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-400 transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div>
              <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Employee <span class="text-red-500">*</span></label>
              <select v-model="form.employee_id" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="">— Select employee —</option>
                <option v-for="e in employees" :key="e.id" :value="e.id">{{ e.full_name }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Leave Type <span class="text-red-500">*</span></label>
              <select v-model="form.leave_type_id" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
                <option value="">— Select type —</option>
                <option v-for="lt in leaveTypes" :key="lt.id" :value="lt.id">{{ lt.name }} ({{ lt.days_allowed }}d max)</option>
              </select>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Start Date <span class="text-red-500">*</span></label>
                <input v-model="form.start_date" type="date" @change="recalcDays"
                  class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">End Date <span class="text-red-500">*</span></label>
                <input v-model="form.end_date" type="date" @change="recalcDays"
                  class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500" />
              </div>
            </div>
            <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50 flex items-center gap-3">
              <Calendar class="w-4 h-4 text-blue-500 flex-shrink-0" />
              <span class="text-sm text-slate-700 dark:text-slate-300">
                Duration: <strong class="text-blue-600 dark:text-blue-400">{{ form.days_count }} day{{ form.days_count!==1?'s':'' }}</strong>
              </span>
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Reason</label>
              <textarea v-model="form.reason" rows="2"
                class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none" />
            </div>
          </div>
          <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-700">
            <button @click="showModal=false" class="px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-600 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
            <button @click="save" :disabled="saving" class="flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors">
              <Save class="w-4 h-4" /> {{ saving ? 'Submitting…' : 'Submit Request' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Detail Drawer -->
    <Teleport to="body">
      <div v-if="showDetail && selected" class="fixed inset-0 z-50 flex">
        <div class="flex-1 bg-black/40 backdrop-blur-sm" @click="showDetail=false" />
        <div class="w-full max-w-sm bg-white dark:bg-slate-800 shadow-2xl overflow-y-auto">
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700 sticky top-0 bg-white dark:bg-slate-800">
            <h2 class="text-lg font-bold text-slate-900 dark:text-white">Leave Request Details</h2>
            <button @click="showDetail=false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-400 transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div class="p-4 rounded-xl" :style="{background: (selected.leave_color||'#6366f1')+'20', borderLeft: '4px solid '+(selected.leave_color||'#6366f1')}">
              <p class="text-xs font-semibold text-slate-500 uppercase mb-1">{{ selected.leave_type_name }}</p>
              <p class="text-xl font-bold text-slate-900 dark:text-white">{{ selected.days_count }} days</p>
              <p class="text-sm text-slate-600 dark:text-slate-400 mt-1">{{ fmtDate(selected.start_date) }} — {{ fmtDate(selected.end_date) }}</p>
            </div>
            <div class="grid grid-cols-1 gap-3">
              <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50">
                <p class="text-xs text-slate-400 mb-1">Employee</p>
                <p class="font-semibold text-slate-800 dark:text-slate-200">{{ selected.employee_name }}</p>
                <p class="text-xs text-slate-400">{{ selected.department_name }}</p>
              </div>
              <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50">
                <p class="text-xs text-slate-400 mb-1">Status</p>
                <span class="px-2 py-1 rounded-full text-xs font-semibold capitalize" :class="statusColor(selected.status)">{{ selected.status }}</span>
              </div>
              <div v-if="selected.reason" class="p-3 rounded-lg bg-slate-50 dark:bg-slate-700/50">
                <p class="text-xs text-slate-400 mb-1">Reason</p>
                <p class="text-sm text-slate-700 dark:text-slate-300">{{ selected.reason }}</p>
              </div>
              <div v-if="selected.rejection_reason" class="p-3 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800">
                <p class="text-xs text-red-600 dark:text-red-400 font-semibold mb-1">Rejection Reason</p>
                <p class="text-sm text-red-700 dark:text-red-300">{{ selected.rejection_reason }}</p>
              </div>
            </div>
            <div v-if="selected.status==='pending'" class="flex gap-2">
              <button @click="approve(selected); showDetail=false"
                class="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-medium rounded-lg transition-colors">
                <CheckCircle class="w-4 h-4" /> Approve
              </button>
              <button @click="rejectModal=selected; rejectReason=''; showDetail=false"
                class="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-red-600 hover:bg-red-700 text-white text-sm font-medium rounded-lg transition-colors">
                <XCircle class="w-4 h-4" /> Reject
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Reject Modal -->
    <Teleport to="body">
      <div v-if="rejectModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-sm p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="p-2 rounded-lg bg-red-100 dark:bg-red-900/30">
              <AlertCircle class="w-5 h-5 text-red-600 dark:text-red-400" />
            </div>
            <h3 class="text-lg font-bold text-slate-900 dark:text-white">Reject Request</h3>
          </div>
          <p class="text-sm text-slate-600 dark:text-slate-400 mb-3">
            Rejecting leave for <strong>{{ rejectModal.employee_name }}</strong>.
          </p>
          <textarea v-model="rejectReason" rows="3" placeholder="Rejection reason (optional)…"
            class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-red-500 resize-none mb-4" />
          <div class="flex gap-3">
            <button @click="rejectModal=null" class="flex-1 px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-600 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
            <button @click="doReject" class="flex-1 px-4 py-2 rounded-lg bg-red-600 hover:bg-red-700 text-white text-sm font-medium transition-colors">Reject</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
