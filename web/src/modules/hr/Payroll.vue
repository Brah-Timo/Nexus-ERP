<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  DollarSign, Plus, RefreshCw, X, ChevronUp, ChevronDown, Eye,
  CheckCircle, CreditCard, Users, TrendingUp, Play, ArrowUpDown,
  AlertCircle, Download, BarChart2, Banknote
} from 'lucide-vue-next'
import { hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const runs    = ref<any[]>([])
const payslips = ref<any[]>([])
const loading = ref(true)
const running = ref(false)

const selectedRun  = ref<any>(null)
const showPayslips = ref(false)
const showRunModal = ref(false)
const confirmApprove = ref<any>(null)
const confirmPay     = ref<any>(null)

const runForm = ref({ month: new Date().getMonth() + 1, year: new Date().getFullYear() })
const sortKey = ref('period_year')
const sortDir = ref<'asc'|'desc'>('desc')

async function load() {
  loading.value = true
  try {
    const res = await hrAPI.getPayrollRuns()
    runs.value = res.data || []
  } catch { app.addToast('Failed to load payroll runs', 'error') }
  finally { loading.value = false }
}
onMounted(load)

async function viewPayslips(run: any) {
  selectedRun.value  = run
  showPayslips.value = true
  payslips.value     = []
  try {
    const res = await hrAPI.getPayslips(run.id)
    payslips.value = res.data || []
  } catch { app.addToast('Failed to load payslips', 'error') }
}

async function runPayroll() {
  running.value = true
  try {
    const res = await hrAPI.runPayroll(runForm.value.month, runForm.value.year)
    app.addToast(`Payroll calculated for ${monthName(runForm.value.month)} ${runForm.value.year}`, 'success')
    showRunModal.value = false
    await load()
  } catch (e: any) {
    const msg = e?.response?.data?.error || 'Failed to run payroll'
    app.addToast(msg, 'error')
  } finally { running.value = false }
}

async function doApprove() {
  if (!confirmApprove.value) return
  try {
    await hrAPI.approvePayrollRun(confirmApprove.value.id)
    app.addToast('Payroll run approved', 'success')
    confirmApprove.value = null
    await load()
  } catch { app.addToast('Failed to approve', 'error') }
}

async function doPay() {
  if (!confirmPay.value) return
  try {
    await hrAPI.payPayrollRun(confirmPay.value.id)
    app.addToast('Payroll marked as paid', 'success')
    confirmPay.value = null
    await load()
  } catch { app.addToast('Failed to mark as paid', 'error') }
}

// KPIs
const kpis = computed(() => {
  const total     = runs.value.reduce((s, r) => s + (r.total_net || 0), 0)
  const totalEmp  = runs.value.reduce((s, r) => s + (r.total_employees || 0), 0)
  const paid      = runs.value.filter(r => r.status === 'paid').length
  const draft     = runs.value.filter(r => r.status === 'draft').length
  return [
    { label: 'Total Runs',    value: runs.value.length, icon: BarChart2,   color: 'blue'   },
    { label: 'Paid Runs',     value: paid,               icon: CheckCircle, color: 'green'  },
    { label: 'Draft',         value: draft,              icon: AlertCircle, color: 'amber'  },
    { label: 'Total Payouts', value: fmt(total),         icon: Banknote,    color: 'violet' },
  ]
})

const sorted = computed(() => {
  return [...runs.value].sort((a,b) => {
    if (sortKey.value === 'period') {
      const av = a.period_year * 100 + a.period_month
      const bv = b.period_year * 100 + b.period_month
      return sortDir.value === 'asc' ? av - bv : bv - av
    }
    const av = a[sortKey.value] ?? ''; const bv = b[sortKey.value] ?? ''
    return sortDir.value === 'asc' ? String(av).localeCompare(String(bv)) : String(bv).localeCompare(String(av))
  })
})
function sort(k: string) {
  if (sortKey.value===k) sortDir.value=sortDir.value==='asc'?'desc':'asc'
  else { sortKey.value=k; sortDir.value='asc' }
}

function fmt(n: number) {
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0 }).format(Math.round(n)) + ' DA'
}
function fmtDate(d: string) { return d ? new Date(d).toLocaleDateString('fr-FR') : '-' }
function monthName(m: number) {
  return ['Jan','Feb','Mar','Apr','May','Jun','Jul','Aug','Sep','Oct','Nov','Dec'][m-1]
}
function statusColor(s: string) {
  return s==='paid'     ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' :
         s==='approved' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' :
         s==='draft'    ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400' :
                          'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300'
}
function kpiBg(c: string) {
  return c==='blue'?'bg-blue-50 dark:bg-blue-900/20':
         c==='green'?'bg-emerald-50 dark:bg-emerald-900/20':
         c==='amber'?'bg-amber-50 dark:bg-amber-900/20':
         'bg-violet-50 dark:bg-violet-900/20'
}
function kpiText(c: string) {
  return c==='blue'?'text-blue-600 dark:text-blue-400':
         c==='green'?'text-emerald-600 dark:text-emerald-400':
         c==='amber'?'text-amber-600 dark:text-amber-400':
         'text-violet-600 dark:text-violet-400'
}

const months = [
  {v:1,l:'January'},{v:2,l:'February'},{v:3,l:'March'},{v:4,l:'April'},
  {v:5,l:'May'},{v:6,l:'June'},{v:7,l:'July'},{v:8,l:'August'},
  {v:9,l:'September'},{v:10,l:'October'},{v:11,l:'November'},{v:12,l:'December'},
]
</script>

<template>
  <div class="space-y-6 p-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white">Payroll</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Algerian payroll with CNAS & IRG calculations</p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="load" class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-400 transition-colors">
          <RefreshCw class="w-4 h-4" :class="loading&&'animate-spin'" />
        </button>
        <button @click="showRunModal=true" class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors shadow-sm">
          <Play class="w-4 h-4" /> Run Payroll
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

    <!-- Payroll Runs Table -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      <div class="px-4 py-3 border-b border-slate-200 dark:border-slate-700 flex items-center gap-2">
        <DollarSign class="w-4 h-4 text-blue-500" />
        <span class="text-sm font-semibold text-slate-700 dark:text-slate-300">Payroll Runs</span>
      </div>
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="w-8 h-8 animate-spin text-blue-500" />
      </div>
      <div v-else-if="!runs.length" class="flex flex-col items-center py-20 text-slate-400 dark:text-slate-500">
        <DollarSign class="w-12 h-12 mb-3 opacity-40" />
        <p class="font-medium">No payroll runs yet</p>
        <p class="text-sm mt-1">Click "Run Payroll" to calculate salaries</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-700">
            <tr>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide cursor-pointer" @click="sort('period')">
                <div class="flex items-center gap-1">Period
                  <ChevronUp v-if="sortKey==='period'&&sortDir==='asc'" class="w-3 h-3 text-blue-500" />
                  <ChevronDown v-else-if="sortKey==='period'&&sortDir==='desc'" class="w-3 h-3 text-blue-500" />
                  <ArrowUpDown v-else class="w-3 h-3 opacity-30" />
                </div>
              </th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">Employees</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">Gross</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">CNAS Emp.</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">IRG</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">Net</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">Status</th>
              <th class="px-4 py-3 text-right text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700/50">
            <tr v-for="r in sorted" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
              <td class="px-4 py-3">
                <p class="font-bold text-slate-900 dark:text-white">{{ monthName(r.period_month) }} {{ r.period_year }}</p>
                <p class="text-xs text-slate-400">Created {{ fmtDate(r.created_at) }}</p>
              </td>
              <td class="px-4 py-3 text-right">
                <div class="flex items-center justify-end gap-1">
                  <Users class="w-3.5 h-3.5 text-slate-400" />
                  <span class="font-semibold text-slate-800 dark:text-slate-200">{{ r.total_employees }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-right font-medium text-slate-800 dark:text-slate-200">{{ fmt(r.total_gross||0) }}</td>
              <td class="px-4 py-3 text-right text-red-600 dark:text-red-400 font-medium">-{{ fmt(r.total_cnas_employee||0) }}</td>
              <td class="px-4 py-3 text-right text-red-600 dark:text-red-400 font-medium">-{{ fmt(r.total_irg||0) }}</td>
              <td class="px-4 py-3 text-right font-bold text-emerald-600 dark:text-emerald-400">{{ fmt(r.total_net||0) }}</td>
              <td class="px-4 py-3">
                <span class="px-2 py-1 rounded-full text-xs font-semibold capitalize" :class="statusColor(r.status)">{{ r.status }}</span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center justify-end gap-1">
                  <button @click="viewPayslips(r)" class="p-1.5 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/20 text-slate-400 hover:text-blue-600 transition-colors" title="View payslips">
                    <Eye class="w-4 h-4" />
                  </button>
                  <button v-if="r.status==='draft'" @click="confirmApprove=r"
                    class="p-1.5 rounded-lg hover:bg-emerald-50 dark:hover:bg-emerald-900/20 text-slate-400 hover:text-emerald-600 transition-colors" title="Approve">
                    <CheckCircle class="w-4 h-4" />
                  </button>
                  <button v-if="r.status==='approved'" @click="confirmPay=r"
                    class="p-1.5 rounded-lg hover:bg-violet-50 dark:hover:bg-violet-900/20 text-slate-400 hover:text-violet-600 transition-colors" title="Mark as paid">
                    <CreditCard class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Run Payroll Modal -->
    <Teleport to="body">
      <div v-if="showRunModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-md">
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-lg bg-blue-100 dark:bg-blue-900/30">
                <Play class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              </div>
              <h2 class="text-lg font-bold text-slate-900 dark:text-white">Run Payroll</h2>
            </div>
            <button @click="showRunModal=false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-400 transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div class="p-4 rounded-xl bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800">
              <p class="text-xs font-semibold text-blue-600 dark:text-blue-400 uppercase mb-2">Algerian Payroll Rules</p>
              <ul class="text-xs text-blue-700 dark:text-blue-300 space-y-1">
                <li>• CNAS Employee: 9% of gross salary</li>
                <li>• CNAS Employer: 26% of gross salary</li>
                <li>• IRG: Progressive scale (0–35%)</li>
                <li>• Transport allowance: 3 000 DA / month</li>
                <li>• Meal allowance: 1 000 DA / month</li>
              </ul>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Month</label>
                <select v-model.number="runForm.month" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option v-for="m in months" :key="m.v" :value="m.v">{{ m.l }}</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1">Year</label>
                <select v-model.number="runForm.year" class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-sm text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500">
                  <option v-for="y in [2024,2025,2026]" :key="y" :value="y">{{ y }}</option>
                </select>
              </div>
            </div>
          </div>
          <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-700">
            <button @click="showRunModal=false" class="px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-600 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
            <button @click="runPayroll" :disabled="running" class="flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors">
              <Play class="w-4 h-4" /> {{ running ? 'Calculating…' : 'Calculate Payroll' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Payslips Drawer -->
    <Teleport to="body">
      <div v-if="showPayslips && selectedRun" class="fixed inset-0 z-50 flex">
        <div class="flex-1 bg-black/40 backdrop-blur-sm" @click="showPayslips=false" />
        <div class="w-full max-w-3xl bg-white dark:bg-slate-800 shadow-2xl overflow-y-auto">
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700 sticky top-0 bg-white dark:bg-slate-800">
            <div>
              <h2 class="text-lg font-bold text-slate-900 dark:text-white">
                Payslips — {{ monthName(selectedRun.period_month) }} {{ selectedRun.period_year }}
              </h2>
              <p class="text-xs text-slate-400">{{ payslips.length }} employees</p>
            </div>
            <button @click="showPayslips=false" class="p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-400 transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Run summary -->
          <div class="grid grid-cols-4 gap-3 p-4 border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/30">
            <div class="text-center">
              <p class="text-xs text-slate-400 uppercase font-semibold">Gross</p>
              <p class="font-bold text-slate-800 dark:text-slate-200">{{ fmt(selectedRun.total_gross||0) }}</p>
            </div>
            <div class="text-center">
              <p class="text-xs text-slate-400 uppercase font-semibold">CNAS</p>
              <p class="font-bold text-red-600">{{ fmt(selectedRun.total_cnas_employee||0) }}</p>
            </div>
            <div class="text-center">
              <p class="text-xs text-slate-400 uppercase font-semibold">IRG</p>
              <p class="font-bold text-red-600">{{ fmt(selectedRun.total_irg||0) }}</p>
            </div>
            <div class="text-center">
              <p class="text-xs text-slate-400 uppercase font-semibold">Net Total</p>
              <p class="font-bold text-emerald-600 text-lg">{{ fmt(selectedRun.total_net||0) }}</p>
            </div>
          </div>

          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead class="bg-slate-50 dark:bg-slate-900/50 border-b border-slate-200 dark:border-slate-700">
                <tr>
                  <th class="text-left px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase">Employee</th>
                  <th class="text-right px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase">Base</th>
                  <th class="text-right px-4 py-3 text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase">Gross</th>
                  <th class="text-right px-4 py-3 text-xs font-semibold text-red-600 dark:text-red-400 uppercase">CNAS</th>
                  <th class="text-right px-4 py-3 text-xs font-semibold text-red-600 dark:text-red-400 uppercase">IRG</th>
                  <th class="text-right px-4 py-3 text-xs font-semibold text-emerald-600 dark:text-emerald-400 uppercase">Net</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-700/50">
                <tr v-for="ps in payslips" :key="ps.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
                  <td class="px-4 py-3">
                    <p class="font-semibold text-slate-900 dark:text-white">{{ ps.employee_name }}</p>
                    <p class="text-xs text-slate-400">{{ ps.department_name }}</p>
                  </td>
                  <td class="px-4 py-3 text-right text-slate-700 dark:text-slate-300">{{ fmt(ps.base_salary||0) }}</td>
                  <td class="px-4 py-3 text-right text-slate-700 dark:text-slate-300">{{ fmt(ps.gross_salary||0) }}</td>
                  <td class="px-4 py-3 text-right text-red-600 dark:text-red-400 font-medium">-{{ fmt(ps.cnas_employee||0) }}</td>
                  <td class="px-4 py-3 text-right text-red-600 dark:text-red-400 font-medium">-{{ fmt(ps.irg_amount||0) }}</td>
                  <td class="px-4 py-3 text-right font-bold text-emerald-600 dark:text-emerald-400">{{ fmt(ps.net_salary||0) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Approve confirm -->
    <Teleport to="body">
      <div v-if="confirmApprove" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-sm p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="p-2 rounded-lg bg-emerald-100 dark:bg-emerald-900/30">
              <CheckCircle class="w-5 h-5 text-emerald-600 dark:text-emerald-400" />
            </div>
            <h3 class="text-lg font-bold text-slate-900 dark:text-white">Approve Payroll</h3>
          </div>
          <p class="text-sm text-slate-600 dark:text-slate-400 mb-6">
            Approve payroll run for <strong>{{ monthName(confirmApprove.period_month) }} {{ confirmApprove.period_year }}</strong>?
            Net total: <strong class="text-emerald-600">{{ fmt(confirmApprove.total_net||0) }}</strong>
          </p>
          <div class="flex gap-3">
            <button @click="confirmApprove=null" class="flex-1 px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-600 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
            <button @click="doApprove" class="flex-1 px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-medium transition-colors">Approve</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Pay confirm -->
    <Teleport to="body">
      <div v-if="confirmPay" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-sm p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="p-2 rounded-lg bg-violet-100 dark:bg-violet-900/30">
              <CreditCard class="w-5 h-5 text-violet-600 dark:text-violet-400" />
            </div>
            <h3 class="text-lg font-bold text-slate-900 dark:text-white">Mark as Paid</h3>
          </div>
          <p class="text-sm text-slate-600 dark:text-slate-400 mb-6">
            Confirm payment for <strong>{{ monthName(confirmPay.period_month) }} {{ confirmPay.period_year }}</strong>?
          </p>
          <div class="flex gap-3">
            <button @click="confirmPay=null" class="flex-1 px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-600 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">Cancel</button>
            <button @click="doPay" class="flex-1 px-4 py-2 rounded-lg bg-violet-600 hover:bg-violet-700 text-white text-sm font-medium transition-colors">Confirm Payment</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
