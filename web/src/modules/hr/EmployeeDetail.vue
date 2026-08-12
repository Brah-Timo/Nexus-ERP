<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950">

    <!-- Back + Breadcrumb -->
    <div class="px-6 pt-6 pb-4">
      <button @click="$router.push('/hr/employees')" class="inline-flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors font-medium mb-4">
        <ArrowLeft class="w-4 h-4" />
        Back to Employees
      </button>

      <div v-if="loading" class="flex items-center justify-center py-32">
        <Loader2 class="w-10 h-10 text-indigo-500 animate-spin" />
      </div>

      <div v-else-if="!employee" class="flex flex-col items-center justify-center py-32 text-gray-400">
        <UserX class="w-16 h-16 mb-4 opacity-30" />
        <p class="font-medium">Employee not found</p>
      </div>

      <template v-else>
        <!-- Hero Banner -->
        <div class="relative bg-gradient-to-r from-indigo-600 via-violet-600 to-purple-700 rounded-2xl overflow-hidden shadow-xl mb-0">
          <div class="absolute inset-0 opacity-10 bg-[url('data:image/svg+xml,%3Csvg width=60 height=60 viewBox=0 0 60 60 xmlns=http://www.w3.org/2000/svg%3E%3Cg fill=none fill-rule=evenodd%3E%3Cg fill=%23fff fill-opacity=1%3E%3Cpath d=M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z/%3E%3C/g%3E%3C/g%3E%3C/svg%3E')] bg-repeat" />
          <div class="relative flex flex-col sm:flex-row items-start sm:items-end gap-4 p-6 pb-8">
            <!-- Avatar -->
            <div class="w-20 h-20 rounded-2xl bg-white/20 backdrop-blur border-2 border-white/30 flex items-center justify-center text-white text-2xl font-bold flex-shrink-0 shadow-lg">
              {{ initials(employee.full_name) }}
            </div>
            <!-- Name & role -->
            <div class="flex-1 min-w-0">
              <h1 class="text-2xl font-bold text-white truncate">{{ employee.full_name }}</h1>
              <p class="text-indigo-200 text-sm mt-0.5">{{ employee.job_title || 'No title' }} &bull; {{ employee.department_name || 'No department' }}</p>
              <div class="flex flex-wrap items-center gap-2 mt-2">
                <span :class="statusBadge(employee.employment_status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold">{{ employee.employment_status || 'Active' }}</span>
                <span class="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-white/20 text-white">{{ employee.employment_type || '—' }}</span>
                <span class="px-2.5 py-0.5 rounded-full text-xs font-semibold bg-white/20 text-white">{{ employee.employee_number }}</span>
              </div>
            </div>
            <!-- Quick Actions -->
            <div class="flex gap-2 flex-shrink-0">
              <button @click="openEdit" class="inline-flex items-center gap-2 px-4 py-2 bg-white/20 hover:bg-white/30 backdrop-blur border border-white/30 text-white text-sm font-medium rounded-lg transition-colors">
                <Pencil class="w-4 h-4" /> Edit
              </button>
            </div>
          </div>

          <!-- Quick Stats Bar -->
          <div class="grid grid-cols-2 sm:grid-cols-4 divide-x divide-white/20 border-t border-white/20 bg-black/10">
            <div class="px-5 py-3 text-center">
              <p class="text-xs text-indigo-200 font-medium uppercase tracking-wide">Years of Service</p>
              <p class="text-lg font-bold text-white">{{ yearsOfService }}</p>
            </div>
            <div class="px-5 py-3 text-center">
              <p class="text-xs text-indigo-200 font-medium uppercase tracking-wide">Base Salary</p>
              <p class="text-lg font-bold text-white">{{ fmtCurrency(employee.base_salary) }}</p>
            </div>
            <div class="px-5 py-3 text-center">
              <p class="text-xs text-indigo-200 font-medium uppercase tracking-wide">Hire Date</p>
              <p class="text-lg font-bold text-white">{{ fmtDate(employee.hire_date) }}</p>
            </div>
            <div class="px-5 py-3 text-center">
              <p class="text-xs text-indigo-200 font-medium uppercase tracking-wide">Contract</p>
              <p class="text-lg font-bold text-white">{{ employee.employment_type || '—' }}</p>
            </div>
          </div>
        </div>

        <!-- Tabs -->
        <div class="flex gap-1 mt-4 mb-4 bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-1.5">
          <button v-for="tab in tabs" :key="tab.id" @click="activeTab = tab.id"
            :class="activeTab === tab.id ? 'bg-indigo-600 text-white shadow' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'"
            class="flex-1 flex items-center justify-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all">
            <component :is="tab.icon" class="w-4 h-4" />
            <span class="hidden sm:inline">{{ tab.label }}</span>
          </button>
        </div>

        <!-- Tab: Overview -->
        <div v-if="activeTab === 'overview'" class="grid grid-cols-1 lg:grid-cols-2 gap-4">

          <!-- Personal Info -->
          <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
            <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <User class="w-5 h-5 text-indigo-500" />
              <h3 class="font-semibold text-gray-900 dark:text-white">Personal Information</h3>
            </div>
            <div class="p-5 space-y-3">
              <DetailRow label="Full Name" :value="employee.full_name" />
              <DetailRow label="Date of Birth" :value="fmtDate(employee.date_of_birth)" />
              <DetailRow label="Gender" :value="employee.gender" />
              <DetailRow label="National ID" :value="employee.national_id" />
              <DetailRow label="Nationality" :value="employee.nationality" />
              <DetailRow label="Marital Status" :value="employee.marital_status" />
            </div>
          </div>

          <!-- Contact Info -->
          <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
            <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <Phone class="w-5 h-5 text-emerald-500" />
              <h3 class="font-semibold text-gray-900 dark:text-white">Contact Details</h3>
            </div>
            <div class="p-5 space-y-3">
              <DetailRow label="Email" :value="employee.email" />
              <DetailRow label="Phone" :value="employee.phone" />
              <DetailRow label="Address" :value="employee.address" />
              <DetailRow label="City" :value="employee.city" />
              <DetailRow label="Emergency Contact" :value="employee.emergency_contact_name" />
              <DetailRow label="Emergency Phone" :value="employee.emergency_contact_phone" />
            </div>
          </div>

          <!-- Employment Info -->
          <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
            <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <Briefcase class="w-5 h-5 text-violet-500" />
              <h3 class="font-semibold text-gray-900 dark:text-white">Employment</h3>
            </div>
            <div class="p-5 space-y-3">
              <DetailRow label="Employee Number" :value="employee.employee_number" mono />
              <DetailRow label="Department" :value="employee.department_name" />
              <DetailRow label="Job Title" :value="employee.job_title" />
              <DetailRow label="Manager" :value="employee.manager_name" />
              <DetailRow label="Employment Type" :value="employee.employment_type" />
              <DetailRow label="Hire Date" :value="fmtDate(employee.hire_date)" />
              <DetailRow label="Base Salary" :value="fmtCurrency(employee.base_salary)" />
              <DetailRow label="Status" :value="employee.employment_status" />
            </div>
          </div>

          <!-- Banking Info -->
          <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
            <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <CreditCard class="w-5 h-5 text-amber-500" />
              <h3 class="font-semibold text-gray-900 dark:text-white">Banking</h3>
            </div>
            <div class="p-5 space-y-3">
              <DetailRow label="Bank Name" :value="employee.bank_name" />
              <DetailRow label="IBAN" :value="employee.iban" mono />
              <DetailRow label="RIB" :value="employee.rib" mono />
            </div>
          </div>

        </div>

        <!-- Tab: Attendance -->
        <div v-if="activeTab === 'attendance'">
          <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
            <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <div class="flex items-center gap-3">
                <CalendarDays class="w-5 h-5 text-indigo-500" />
                <h3 class="font-semibold text-gray-900 dark:text-white">Attendance Records</h3>
              </div>
              <span class="text-xs text-gray-400">Last 30 days</span>
            </div>
            <div v-if="loadingAttendance" class="flex items-center justify-center py-12">
              <Loader2 class="w-6 h-6 text-indigo-500 animate-spin" />
            </div>
            <div v-else-if="attendance.length === 0" class="flex flex-col items-center justify-center py-12 text-gray-400">
              <CalendarDays class="w-10 h-10 mb-2 opacity-30" />
              <p class="text-sm">No attendance records found</p>
            </div>
            <div v-else class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Date</th>
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Check In</th>
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Check Out</th>
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Hours</th>
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Status</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                  <tr v-for="rec in attendance" :key="rec.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
                    <td class="px-4 py-3 font-medium text-gray-900 dark:text-white">{{ fmtDate(rec.date) }}</td>
                    <td class="px-4 py-3 text-gray-600 dark:text-gray-300">{{ fmtTime(rec.check_in) }}</td>
                    <td class="px-4 py-3 text-gray-600 dark:text-gray-300">{{ fmtTime(rec.check_out) }}</td>
                    <td class="px-4 py-3">
                      <span class="font-semibold text-gray-900 dark:text-white">{{ rec.hours_worked?.toFixed(1) || '—' }}h</span>
                    </td>
                    <td class="px-4 py-3">
                      <span :class="attendanceBadge(rec.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold">{{ rec.status || 'present' }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- Tab: Leaves -->
        <div v-if="activeTab === 'leaves'">
          <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
            <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <Plane class="w-5 h-5 text-sky-500" />
              <h3 class="font-semibold text-gray-900 dark:text-white">Leave Requests</h3>
            </div>
            <div v-if="loadingLeaves" class="flex items-center justify-center py-12">
              <Loader2 class="w-6 h-6 text-indigo-500 animate-spin" />
            </div>
            <div v-else-if="leaves.length === 0" class="flex flex-col items-center justify-center py-12 text-gray-400">
              <Plane class="w-10 h-10 mb-2 opacity-30" />
              <p class="text-sm">No leave requests found</p>
            </div>
            <div v-else class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Type</th>
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">From</th>
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">To</th>
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Days</th>
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Status</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                  <tr v-for="leave in leaves" :key="leave.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
                    <td class="px-4 py-3 font-medium text-gray-900 dark:text-white">{{ leave.leave_type_name || leave.leave_type_id }}</td>
                    <td class="px-4 py-3 text-gray-600 dark:text-gray-300">{{ fmtDate(leave.start_date) }}</td>
                    <td class="px-4 py-3 text-gray-600 dark:text-gray-300">{{ fmtDate(leave.end_date) }}</td>
                    <td class="px-4 py-3 font-semibold text-gray-900 dark:text-white">{{ leave.days_count }}</td>
                    <td class="px-4 py-3">
                      <span :class="leaveBadge(leave.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold">{{ leave.status }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- Tab: Payslips -->
        <div v-if="activeTab === 'payslips'">
          <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
            <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800">
              <FileText class="w-5 h-5 text-green-500" />
              <h3 class="font-semibold text-gray-900 dark:text-white">Payslips</h3>
            </div>
            <div v-if="loadingPayslips" class="flex items-center justify-center py-12">
              <Loader2 class="w-6 h-6 text-indigo-500 animate-spin" />
            </div>
            <div v-else-if="payslips.length === 0" class="flex flex-col items-center justify-center py-12 text-gray-400">
              <FileText class="w-10 h-10 mb-2 opacity-30" />
              <p class="text-sm">No payslips found</p>
            </div>
            <div v-else class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Period</th>
                    <th class="text-right px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Gross</th>
                    <th class="text-right px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">CNAS Emp.</th>
                    <th class="text-right px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">IRG</th>
                    <th class="text-right px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Net Pay</th>
                    <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Status</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                  <tr v-for="slip in payslips" :key="slip.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
                    <td class="px-4 py-3 font-medium text-gray-900 dark:text-white">{{ slip.period_month }}/{{ slip.period_year }}</td>
                    <td class="px-4 py-3 text-right text-gray-700 dark:text-gray-200">{{ fmtCurrency(slip.gross_salary) }}</td>
                    <td class="px-4 py-3 text-right text-red-600 dark:text-red-400">-{{ fmtCurrency(slip.total_cnas_employee) }}</td>
                    <td class="px-4 py-3 text-right text-red-600 dark:text-red-400">-{{ fmtCurrency(slip.irg_amount) }}</td>
                    <td class="px-4 py-3 text-right font-bold text-emerald-600 dark:text-emerald-400">{{ fmtCurrency(slip.net_salary) }}</td>
                    <td class="px-4 py-3">
                      <span :class="payslipBadge(slip.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold">{{ slip.status }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowLeft, Loader2, UserX, Pencil, User, Phone, Briefcase, CreditCard,
  CalendarDays, Plane, FileText
} from 'lucide-vue-next'
import { hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

// Inline component for detail rows
const DetailRow = {
  props: ['label', 'value', 'mono'],
  template: `
    <div class="flex justify-between items-start gap-2 py-1">
      <span class="text-xs font-medium text-gray-400 uppercase tracking-wide flex-shrink-0">{{ label }}</span>
      <span :class="mono ? 'font-mono text-xs' : 'text-sm text-right'" class="text-gray-700 dark:text-gray-200 break-all">{{ value || '—' }}</span>
    </div>
  `
}

const route = useRoute()
const router = useRouter()
const store = useAppStore()

const loading = ref(true)
const loadingAttendance = ref(false)
const loadingLeaves = ref(false)
const loadingPayslips = ref(false)
const employee = ref<any>(null)
const attendance = ref<any[]>([])
const leaves = ref<any[]>([])
const payslips = ref<any[]>([])
const activeTab = ref('overview')

const tabs = [
  { id: 'overview', label: 'Overview', icon: User },
  { id: 'attendance', label: 'Attendance', icon: CalendarDays },
  { id: 'leaves', label: 'Leaves', icon: Plane },
  { id: 'payslips', label: 'Payslips', icon: FileText }
]

const yearsOfService = computed(() => {
  if (!employee.value?.hire_date) return '—'
  const diff = (Date.now() - new Date(employee.value.hire_date).getTime()) / (1000 * 60 * 60 * 24 * 365.25)
  return diff.toFixed(1) + ' yrs'
})

function initials(name: string) {
  return name?.split(' ').slice(0, 2).map((w: string) => w[0]).join('').toUpperCase() || '?'
}
function fmtDate(d?: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}
function fmtTime(d?: string) {
  if (!d) return '—'
  return new Date(d).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
}
function fmtCurrency(n?: number) {
  if (n === undefined || n === null) return '—'
  return new Intl.NumberFormat('fr-DZ', { style: 'currency', currency: 'DZD', minimumFractionDigits: 0 }).format(n)
}

function statusBadge(s?: string) {
  switch (s) {
    case 'active': return 'bg-emerald-500/20 text-emerald-200'
    case 'inactive': return 'bg-red-500/20 text-red-200'
    default: return 'bg-emerald-500/20 text-emerald-200'
  }
}
function attendanceBadge(s?: string) {
  switch (s) {
    case 'present': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'absent': return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    case 'late': return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    default: return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
  }
}
function leaveBadge(s?: string) {
  switch (s) {
    case 'approved': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'rejected': return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    case 'pending': return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    default: return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
  }
}
function payslipBadge(s?: string) {
  switch (s) {
    case 'paid': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'draft': return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
    case 'validated': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    default: return 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'
  }
}

function openEdit() {
  router.push(`/hr/employees?edit=${route.params.id}`)
}

async function loadAttendance() {
  loadingAttendance.value = true
  try {
    const res = await hrAPI.getAttendance({ employee_id: route.params.id })
    attendance.value = res.data || []
  } catch { attendance.value = [] }
  finally { loadingAttendance.value = false }
}
async function loadLeaves() {
  loadingLeaves.value = true
  try {
    const res = await hrAPI.getLeaveRequests({ employee_id: route.params.id })
    leaves.value = res.data || []
  } catch { leaves.value = [] }
  finally { loadingLeaves.value = false }
}
async function loadPayslips() {
  loadingPayslips.value = true
  try {
    const res = await hrAPI.getPayslips({ employee_id: route.params.id })
    payslips.value = res.data || []
  } catch { payslips.value = [] }
  finally { loadingPayslips.value = false }
}

// Watch tab changes
import { watch } from 'vue'
watch(activeTab, (tab) => {
  if (tab === 'attendance' && attendance.value.length === 0) loadAttendance()
  if (tab === 'leaves' && leaves.value.length === 0) loadLeaves()
  if (tab === 'payslips' && payslips.value.length === 0) loadPayslips()
})

onMounted(async () => {
  loading.value = true
  try {
    const res = await hrAPI.getEmployee(route.params.id as string)
    employee.value = res.data
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Employee not found', 'error')
  } finally {
    loading.value = false
  }
})
</script>
