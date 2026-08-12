<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Project Reports</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Budget vs actual, hours breakdown, and project performance</p>
      </div>
      <div class="flex items-center gap-3">
        <div class="relative">
          <select v-model="filterProject" @change="loadDetail" class="appearance-none pl-4 pr-10 py-2 text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white shadow-sm">
            <option value="">All Projects</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
        <button @click="load" class="inline-flex items-center gap-2 px-3 py-2 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 text-sm font-medium rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors shadow-sm">
          <RefreshCw class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- Summary KPIs -->
    <div class="grid grid-cols-2 lg:grid-cols-5 gap-4">
      <div v-for="kpi in summaryKpis" :key="kpi.label" class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4 flex items-center gap-3">
        <div :class="kpi.iconBg" class="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0">
          <component :is="kpi.icon" :class="kpi.iconColor" class="w-5 h-5" />
        </div>
        <div>
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ kpi.label }}</p>
          <p class="text-lg font-bold text-gray-900 dark:text-white">{{ kpi.value }}</p>
        </div>
      </div>
    </div>

    <!-- Two column: Budget vs Actual + By Status -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">

      <!-- Budget vs Actual Chart -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800">
          <BarChart3 class="w-5 h-5 text-indigo-500" />
          <h3 class="font-semibold text-gray-900 dark:text-white">Budget vs Actual</h3>
        </div>
        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 class="w-6 h-6 text-indigo-500 animate-spin" />
        </div>
        <div v-else-if="topProjects.length === 0" class="flex items-center justify-center py-12 text-gray-400 text-sm">
          No data
        </div>
        <div v-else class="p-5 space-y-4">
          <div v-for="proj in topProjects" :key="proj.id">
            <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 mb-1">
              <span class="font-medium text-gray-700 dark:text-gray-200 truncate max-w-32">{{ proj.name }}</span>
              <div class="text-right">
                <span class="text-gray-900 dark:text-white font-semibold">{{ fmtNum(proj.actual_cost) }}</span>
                <span class="text-gray-400"> / {{ fmtNum(proj.budget) }}</span>
              </div>
            </div>
            <div class="relative h-5 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
              <div :style="{ width: proj.budget > 0 ? '100%' : '0' }" class="absolute inset-y-0 left-0 bg-gray-200 dark:bg-gray-700 rounded-full" />
              <div
                :style="{ width: `${Math.min(100, budgetPct(proj.actual_cost, proj.budget))}%` }"
                :class="budgetColor(proj.actual_cost, proj.budget)"
                class="absolute inset-y-0 left-0 rounded-full transition-all"
              />
              <span class="absolute inset-y-0 right-2 flex items-center text-xs font-bold text-gray-700 dark:text-gray-200">
                {{ budgetPct(proj.actual_cost, proj.budget).toFixed(0) }}%
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Projects by Status -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800">
          <PieChart class="w-5 h-5 text-violet-500" />
          <h3 class="font-semibold text-gray-900 dark:text-white">Projects by Status</h3>
        </div>
        <div class="p-5 space-y-3">
          <div v-for="stat in statusBreakdown" :key="stat.status" class="flex items-center gap-3">
            <div :class="stat.dotColor" class="w-3 h-3 rounded-full flex-shrink-0"></div>
            <div class="flex-1">
              <div class="flex justify-between text-xs mb-1">
                <span class="font-medium text-gray-700 dark:text-gray-200 capitalize">{{ stat.status.replace('_', ' ') }}</span>
                <span class="text-gray-500 dark:text-gray-400">{{ stat.count }} ({{ stat.pct.toFixed(0) }}%)</span>
              </div>
              <div class="h-2 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
                <div :style="{ width: `${stat.pct}%` }" :class="stat.barColor" class="h-full rounded-full"></div>
              </div>
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- Detail: Project Costs (when single project selected) -->
    <div v-if="filterProject && projectCosts" class="grid grid-cols-1 lg:grid-cols-2 gap-4">

      <!-- By Employee Hours -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800">
          <Users class="w-5 h-5 text-emerald-500" />
          <h3 class="font-semibold text-gray-900 dark:text-white">Hours by Employee</h3>
        </div>
        <div v-if="loadingDetail" class="flex items-center justify-center py-8">
          <Loader2 class="w-5 h-5 text-indigo-500 animate-spin" />
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
                <th class="text-left px-4 py-2.5 font-semibold text-gray-600 dark:text-gray-300">Employee</th>
                <th class="text-right px-4 py-2.5 font-semibold text-gray-600 dark:text-gray-300">Hours</th>
                <th class="text-right px-4 py-2.5 font-semibold text-gray-600 dark:text-gray-300">Cost</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
              <tr v-for="emp in projectCosts.by_employee" :key="emp.employee_id" class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
                <td class="px-4 py-2.5 font-medium text-gray-900 dark:text-white">{{ emp.employee_name }}</td>
                <td class="px-4 py-2.5 text-right text-gray-700 dark:text-gray-200">{{ emp.total_hours?.toFixed(1) }}h</td>
                <td class="px-4 py-2.5 text-right text-indigo-600 dark:text-indigo-400 font-semibold">{{ fmtCurrency(emp.total_cost) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Budget Summary -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800">
          <DollarSign class="w-5 h-5 text-amber-500" />
          <h3 class="font-semibold text-gray-900 dark:text-white">Budget Summary</h3>
        </div>
        <div v-if="loadingDetail" class="flex items-center justify-center py-8">
          <Loader2 class="w-5 h-5 text-indigo-500 animate-spin" />
        </div>
        <div v-else-if="projectCosts" class="p-5 space-y-4">
          <div class="flex justify-between items-center py-2 border-b border-gray-100 dark:border-gray-800">
            <span class="text-sm text-gray-600 dark:text-gray-300">Total Budget</span>
            <span class="font-bold text-gray-900 dark:text-white">{{ fmtCurrency(projectCosts.budget) }}</span>
          </div>
          <div class="flex justify-between items-center py-2 border-b border-gray-100 dark:border-gray-800">
            <span class="text-sm text-gray-600 dark:text-gray-300">Actual Cost</span>
            <span :class="projectCosts.actual_cost > projectCosts.budget ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'" class="font-bold">{{ fmtCurrency(projectCosts.actual_cost) }}</span>
          </div>
          <div class="flex justify-between items-center py-2 border-b border-gray-100 dark:border-gray-800">
            <span class="text-sm text-gray-600 dark:text-gray-300">Labor Cost</span>
            <span class="font-bold text-gray-700 dark:text-gray-200">{{ fmtCurrency(projectCosts.labor_cost) }}</span>
          </div>
          <div class="flex justify-between items-center py-2 border-b border-gray-100 dark:border-gray-800">
            <span class="text-sm text-gray-600 dark:text-gray-300">Expense Cost</span>
            <span class="font-bold text-gray-700 dark:text-gray-200">{{ fmtCurrency(projectCosts.expense_cost) }}</span>
          </div>
          <div class="flex justify-between items-center py-2">
            <span class="text-sm font-semibold text-gray-700 dark:text-gray-200">Variance</span>
            <span :class="projectCosts.variance >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'" class="font-bold text-lg">
              {{ projectCosts.variance >= 0 ? '+' : '' }}{{ fmtCurrency(projectCosts.variance) }}
            </span>
          </div>
          <div class="pt-2">
            <div class="flex justify-between text-xs text-gray-400 mb-1.5">
              <span>Budget Utilization</span>
              <span class="font-semibold">{{ projectCosts.completion_pct?.toFixed(1) }}%</span>
            </div>
            <div class="h-3 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
              <div
                :style="{ width: `${Math.min(100, projectCosts.completion_pct || 0)}%` }"
                :class="(projectCosts.completion_pct || 0) > 100 ? 'bg-red-500' : (projectCosts.completion_pct || 0) > 80 ? 'bg-amber-500' : 'bg-emerald-500'"
                class="h-full rounded-full transition-all"
              ></div>
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- All Projects Table -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
      <div class="flex items-center gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800">
        <FileText class="w-5 h-5 text-indigo-500" />
        <h3 class="font-semibold text-gray-900 dark:text-white">All Projects Summary</h3>
      </div>
      <div v-if="loading" class="flex items-center justify-center py-10">
        <Loader2 class="w-6 h-6 text-indigo-500 animate-spin" />
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
              <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Project</th>
              <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Status</th>
              <th class="text-right px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Budget</th>
              <th class="text-right px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Actual Cost</th>
              <th class="text-right px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Variance</th>
              <th class="text-left px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Progress</th>
              <th class="text-right px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Tasks</th>
              <th class="text-right px-4 py-3 font-semibold text-gray-600 dark:text-gray-300">Hours</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr
              v-for="proj in reportData"
              :key="proj.id"
              class="hover:bg-gray-50 dark:hover:bg-gray-800/40 cursor-pointer transition-colors"
              @click="filterProject = proj.id; loadDetail()"
            >
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <div :class="projAvatarBg(proj.name)" class="w-7 h-7 rounded-lg flex items-center justify-center text-white text-xs font-bold flex-shrink-0">{{ initials(proj.name) }}</div>
                  <div>
                    <p class="font-semibold text-gray-900 dark:text-white">{{ proj.name }}</p>
                    <p class="text-xs font-mono text-gray-400">{{ proj.code }}</p>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3">
                <span :class="statusBadge(proj.status)" class="px-2.5 py-0.5 rounded-full text-xs font-semibold capitalize">{{ proj.status?.replace('_', ' ') }}</span>
              </td>
              <td class="px-4 py-3 text-right text-gray-700 dark:text-gray-200">{{ fmtCurrency(proj.budget) }}</td>
              <td class="px-4 py-3 text-right font-semibold text-gray-900 dark:text-white">{{ fmtCurrency(proj.actual_cost) }}</td>
              <td class="px-4 py-3 text-right" :class="(proj.budget - proj.actual_cost) >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
                <span class="font-semibold">{{ fmtCurrency(proj.budget - proj.actual_cost) }}</span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2 w-28">
                  <div class="flex-1 h-1.5 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
                    <div :style="{ width: `${proj.progress_pct || 0}%` }" :class="proj.progress_pct >= 80 ? 'bg-emerald-500' : 'bg-indigo-500'" class="h-full rounded-full"></div>
                  </div>
                  <span class="text-xs font-semibold text-gray-700 dark:text-gray-200">{{ proj.progress_pct || 0 }}%</span>
                </div>
              </td>
              <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-300">
                <span class="font-semibold text-gray-900 dark:text-white">{{ proj.task_count || 0 }}</span> / <span class="text-emerald-500">{{ proj.done_count || 0 }}</span>
              </td>
              <td class="px-4 py-3 text-right text-gray-600 dark:text-gray-300">{{ proj.total_hours?.toFixed(1) || '0' }}h</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  ChevronDown, Loader2, RefreshCw, BarChart3, PieChart, Users, DollarSign,
  FileText, FolderKanban, CheckCircle, TrendingUp, Clock
} from 'lucide-vue-next'
import { projectsAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

const loading = ref(true)
const loadingDetail = ref(false)
const projects = ref<any[]>([])
const reportData = ref<any[]>([])
const projectCosts = ref<any>(null)
const filterProject = ref('')

const summaryKpis = computed(() => {
  const total = reportData.value.length
  const active = reportData.value.filter(p => p.status === 'active').length
  const totalBudget = reportData.value.reduce((s, p) => s + (p.budget || 0), 0)
  const totalActual = reportData.value.reduce((s, p) => s + (p.actual_cost || 0), 0)
  const totalHours = reportData.value.reduce((s, p) => s + (p.total_hours || 0), 0)
  return [
    { label: 'Total Projects', value: total, icon: FolderKanban, iconBg: 'bg-indigo-100 dark:bg-indigo-900/40', iconColor: 'text-indigo-600 dark:text-indigo-400' },
    { label: 'Active', value: active, icon: CheckCircle, iconBg: 'bg-emerald-100 dark:bg-emerald-900/40', iconColor: 'text-emerald-600 dark:text-emerald-400' },
    { label: 'Total Budget', value: fmtCurrency(totalBudget), icon: DollarSign, iconBg: 'bg-amber-100 dark:bg-amber-900/40', iconColor: 'text-amber-600 dark:text-amber-400' },
    { label: 'Total Spent', value: fmtCurrency(totalActual), icon: TrendingUp, iconBg: 'bg-rose-100 dark:bg-rose-900/40', iconColor: 'text-rose-600 dark:text-rose-400' },
    { label: 'Total Hours', value: totalHours.toFixed(0) + 'h', icon: Clock, iconBg: 'bg-violet-100 dark:bg-violet-900/40', iconColor: 'text-violet-600 dark:text-violet-400' },
  ]
})

const topProjects = computed(() => {
  return [...reportData.value].sort((a, b) => (b.budget || 0) - (a.budget || 0)).slice(0, 6)
})

const statusBreakdown = computed(() => {
  const statuses = ['active', 'planning', 'completed', 'on_hold', 'cancelled']
  const total = reportData.value.length || 1
  const colors: Record<string, { dot: string; bar: string }> = {
    active: { dot: 'bg-emerald-500', bar: 'bg-emerald-500' },
    planning: { dot: 'bg-blue-500', bar: 'bg-blue-500' },
    completed: { dot: 'bg-violet-500', bar: 'bg-violet-500' },
    on_hold: { dot: 'bg-amber-500', bar: 'bg-amber-500' },
    cancelled: { dot: 'bg-red-400', bar: 'bg-red-400' },
  }
  return statuses.map(s => {
    const count = reportData.value.filter(p => p.status === s).length
    return { status: s, count, pct: (count / total) * 100, dotColor: colors[s].dot, barColor: colors[s].bar }
  }).filter(s => s.count > 0)
})

function budgetPct(actual?: number, budget?: number) {
  if (!budget || budget === 0) return 0
  return ((actual || 0) / budget) * 100
}
function budgetColor(actual?: number, budget?: number) {
  const pct = budgetPct(actual, budget)
  if (pct > 100) return 'bg-red-500'
  if (pct > 80) return 'bg-amber-500'
  return 'bg-emerald-500'
}

function fmtCurrency(n?: number) {
  if (!n) return '0 DZD'
  if (Math.abs(n) >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M DZD'
  if (Math.abs(n) >= 1_000) return (n / 1_000).toFixed(0) + 'K DZD'
  return n.toLocaleString() + ' DZD'
}
function fmtNum(n?: number) {
  if (!n) return '0'
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(0) + 'K'
  return n.toString()
}
const COLORS = ['from-indigo-500 to-blue-600', 'from-violet-500 to-purple-600', 'from-emerald-500 to-teal-600', 'from-amber-500 to-orange-500', 'from-rose-500 to-pink-600', 'from-cyan-500 to-sky-600']
function projAvatarBg(name: string) { return `bg-gradient-to-br ${COLORS[name.charCodeAt(0) % COLORS.length]}` }
function initials(name: string) { return name?.split(' ').slice(0, 2).map(w => w[0]).join('').toUpperCase() }
function statusBadge(s?: string) {
  switch (s) {
    case 'active': return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'planning': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    case 'on_hold': return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    case 'completed': return 'bg-violet-100 text-violet-700 dark:bg-violet-900/30 dark:text-violet-400'
    case 'cancelled': return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default: return 'bg-gray-100 text-gray-600'
  }
}

async function load() {
  loading.value = true
  try {
    const [rRes, pRes] = await Promise.all([projectsAPI.getProjectsReport(), projectsAPI.getProjects()])
    reportData.value = rRes.data || []
    projects.value = pRes.data || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load report', 'error')
  } finally {
    loading.value = false
  }
}

async function loadDetail() {
  if (!filterProject.value) { projectCosts.value = null; return }
  loadingDetail.value = true
  try {
    const res = await projectsAPI.getProjectCosts(filterProject.value)
    projectCosts.value = res.data
  } catch { projectCosts.value = null }
  finally { loadingDetail.value = false }
}

onMounted(load)
</script>
