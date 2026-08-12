<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { helpdeskAPI } from '@/api/client'
import {
  Headphones, Ticket, Clock, CheckCircle2, AlertTriangle,
  TrendingUp, Users, Star, BarChart3, RefreshCw,
  ArrowUpRight, ArrowDownRight, Timer, XCircle
} from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(true)
const error = ref('')

const dashboard = ref<any>({
  kpi: {},
  priority_counts: [],
  status_counts: [],
  trend: [],
  agent_stats: [],
  recent_tickets: []
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await helpdeskAPI.getDashboard()
    dashboard.value = res.data
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to load dashboard'
  } finally {
    loading.value = false
  }
}

onMounted(load)

const statusColor: Record<string, string> = {
  open: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  pending: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300',
  in_progress: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
  resolved: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
  closed: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
  cancelled: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
}
const priorityColor: Record<string, string> = {
  low: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',
  medium: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  high: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300',
  critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
}

function csatStars(score: number) {
  return Math.round(score)
}
</script>

<template>
  <div :class="app.darkMode ? 'bg-slate-900 text-slate-100' : 'bg-slate-50 text-slate-900'" class="min-h-screen p-6">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <Headphones class="w-7 h-7 text-indigo-500" />
        <div>
          <h1 class="text-2xl font-bold">Support Dashboard</h1>
          <p class="text-sm text-slate-500">Helpdesk overview and key metrics</p>
        </div>
      </div>
      <button @click="load" :class="app.darkMode ? 'bg-slate-700 hover:bg-slate-600 text-slate-200' : 'bg-white hover:bg-slate-100 text-slate-700'"
        class="flex items-center gap-2 px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-600 text-sm font-medium transition-colors">
        <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
        Refresh
      </button>
    </div>

    <div v-if="error" class="mb-4 p-4 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg text-sm">
      {{ error }}
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-4 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">Open Tickets</span>
          <Ticket class="w-5 h-5 text-blue-500" />
        </div>
        <div class="text-2xl font-bold text-blue-600">{{ dashboard.kpi.total_open ?? 0 }}</div>
        <div class="text-xs text-slate-500 mt-1">+{{ dashboard.kpi.opened_today ?? 0 }} today</div>
      </div>

      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-4 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">In Progress</span>
          <Clock class="w-5 h-5 text-purple-500" />
        </div>
        <div class="text-2xl font-bold text-purple-600">{{ dashboard.kpi.total_in_progress ?? 0 }}</div>
        <div class="text-xs text-slate-500 mt-1">{{ dashboard.kpi.total_pending ?? 0 }} pending</div>
      </div>

      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-4 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">Resolved Today</span>
          <CheckCircle2 class="w-5 h-5 text-green-500" />
        </div>
        <div class="text-2xl font-bold text-green-600">{{ dashboard.kpi.resolved_today ?? 0 }}</div>
        <div class="text-xs text-slate-500 mt-1">{{ dashboard.kpi.total_resolved ?? 0 }} total resolved</div>
      </div>

      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-4 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">SLA Breached</span>
          <AlertTriangle class="w-5 h-5 text-red-500" />
        </div>
        <div class="text-2xl font-bold text-red-600">{{ dashboard.kpi.overdue_sla ?? 0 }}</div>
        <div class="text-xs text-slate-500 mt-1">overdue tickets</div>
      </div>
    </div>

    <!-- Second row KPIs -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-4 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">Avg Resolution</span>
          <Timer class="w-5 h-5 text-indigo-500" />
        </div>
        <div class="text-2xl font-bold">{{ (dashboard.kpi.avg_resolution_hrs ?? 0).toFixed(1) }}h</div>
        <div class="text-xs text-slate-500 mt-1">average resolution time</div>
      </div>

      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-4 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">CSAT Score</span>
          <Star class="w-5 h-5 text-yellow-500" />
        </div>
        <div class="text-2xl font-bold text-yellow-600">{{ (dashboard.kpi.csat_score ?? 0).toFixed(1) }}/5</div>
        <div class="text-xs text-slate-500 mt-1">{{ dashboard.kpi.csat_responses ?? 0 }} responses</div>
      </div>

      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-4 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">Active Agents</span>
          <Users class="w-5 h-5 text-teal-500" />
        </div>
        <div class="text-2xl font-bold text-teal-600">{{ dashboard.kpi.active_agents ?? 0 }}</div>
        <div class="text-xs text-slate-500 mt-1">available agents</div>
      </div>

      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-4 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">Total Closed</span>
          <XCircle class="w-5 h-5 text-slate-500" />
        </div>
        <div class="text-2xl font-bold">{{ dashboard.kpi.total_closed ?? 0 }}</div>
        <div class="text-xs text-slate-500 mt-1">closed tickets</div>
      </div>
    </div>

    <!-- Charts row -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
      <!-- Status distribution -->
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-5 shadow-sm">
        <h3 class="font-semibold mb-4 flex items-center gap-2">
          <BarChart3 class="w-4 h-4 text-indigo-500" /> Tickets by Status
        </h3>
        <div class="space-y-2">
          <div v-for="s in dashboard.status_counts" :key="s.status" class="flex items-center gap-2">
            <span :class="statusColor[s.status] || 'bg-slate-100 text-slate-600'"
              class="text-xs px-2 py-0.5 rounded-full font-medium capitalize w-24 text-center">{{ s.status.replace('_',' ') }}</span>
            <div class="flex-1 bg-slate-100 dark:bg-slate-700 rounded-full h-2">
              <div class="bg-indigo-500 h-2 rounded-full transition-all"
                :style="`width:${Math.min(100,(s.count/(dashboard.kpi.total_open+dashboard.kpi.total_in_progress+dashboard.kpi.total_pending+dashboard.kpi.total_resolved+dashboard.kpi.total_closed+1))*100)}%`"></div>
            </div>
            <span class="text-sm font-bold w-8 text-right">{{ s.count }}</span>
          </div>
        </div>
      </div>

      <!-- Priority distribution -->
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-5 shadow-sm">
        <h3 class="font-semibold mb-4 flex items-center gap-2">
          <AlertTriangle class="w-4 h-4 text-orange-500" /> Tickets by Priority
        </h3>
        <div class="space-y-2">
          <div v-for="p in dashboard.priority_counts" :key="p.priority" class="flex items-center gap-2">
            <span :class="priorityColor[p.priority] || 'bg-slate-100 text-slate-600'"
              class="text-xs px-2 py-0.5 rounded-full font-medium capitalize w-20 text-center">{{ p.priority }}</span>
            <div class="flex-1 bg-slate-100 dark:bg-slate-700 rounded-full h-2">
              <div class="h-2 rounded-full transition-all"
                :class="p.priority==='critical'?'bg-red-500':p.priority==='high'?'bg-orange-500':p.priority==='medium'?'bg-blue-500':'bg-slate-400'"
                :style="`width:${Math.min(100,(p.count/Math.max(...(dashboard.priority_counts||[{count:1}]).map((x:any)=>x.count),1))*100)}%`"></div>
            </div>
            <span class="text-sm font-bold w-8 text-right">{{ p.count }}</span>
          </div>
        </div>
        <div v-if="!dashboard.priority_counts?.length" class="text-sm text-slate-400 text-center py-4">No data</div>
      </div>

      <!-- Top agents -->
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-5 shadow-sm">
        <h3 class="font-semibold mb-4 flex items-center gap-2">
          <TrendingUp class="w-4 h-4 text-green-500" /> Top Agents
        </h3>
        <div class="space-y-2">
          <div v-for="a in (dashboard.agent_stats || []).slice(0,6)" :key="a.name"
            class="flex items-center justify-between py-1 border-b border-slate-100 dark:border-slate-700 last:border-0">
            <span class="text-sm font-medium truncate max-w-[130px]">{{ a.name }}</span>
            <div class="text-right">
              <div class="text-sm font-bold text-green-600">{{ a.resolved }} resolved</div>
              <div class="text-xs text-slate-400">{{ a.avg_resolution_hrs.toFixed(1) }}h avg</div>
            </div>
          </div>
        </div>
        <div v-if="!dashboard.agent_stats?.length" class="text-sm text-slate-400 text-center py-4">No agent data</div>
      </div>
    </div>

    <!-- 14-day trend -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border p-5 shadow-sm mb-6">
      <h3 class="font-semibold mb-4 flex items-center gap-2">
        <TrendingUp class="w-4 h-4 text-blue-500" /> 14-Day Ticket Trend
      </h3>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
              <th class="text-left pb-2">Date</th>
              <th class="text-center pb-2">Opened</th>
              <th class="text-center pb-2">Resolved</th>
              <th class="text-center pb-2">Net</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="d in dashboard.trend" :key="d.date"
              class="border-t border-slate-100 dark:border-slate-700">
              <td class="py-1.5 text-xs text-slate-500">{{ d.date?.substring(0,10) }}</td>
              <td class="py-1.5 text-center font-medium text-blue-600">{{ d.opened }}</td>
              <td class="py-1.5 text-center font-medium text-green-600">{{ d.resolved }}</td>
              <td class="py-1.5 text-center">
                <span :class="(d.opened-d.resolved)>0?'text-red-500':'text-green-500'" class="font-medium">
                  {{ d.opened - d.resolved > 0 ? '+' : '' }}{{ d.opened - d.resolved }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Recent tickets -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border p-5 shadow-sm">
      <h3 class="font-semibold mb-4 flex items-center gap-2">
        <Ticket class="w-4 h-4 text-indigo-500" /> Recent Tickets
      </h3>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'text-slate-400 border-slate-700' : 'text-slate-500 border-slate-200'"
              class="border-b text-xs uppercase tracking-wide">
              <th class="text-left pb-3 font-medium">Ticket #</th>
              <th class="text-left pb-3 font-medium">Subject</th>
              <th class="text-left pb-3 font-medium">Status</th>
              <th class="text-left pb-3 font-medium">Priority</th>
              <th class="text-left pb-3 font-medium">Requester</th>
              <th class="text-left pb-3 font-medium">Agent</th>
              <th class="text-left pb-3 font-medium">Created</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in dashboard.recent_tickets" :key="t.id"
              :class="app.darkMode ? 'border-slate-700 hover:bg-slate-700/50' : 'border-slate-100 hover:bg-slate-50'"
              class="border-b transition-colors">
              <td class="py-2.5 font-mono text-xs text-indigo-600">{{ t.ticket_number }}</td>
              <td class="py-2.5 max-w-[180px] truncate">{{ t.subject }}</td>
              <td class="py-2.5">
                <span :class="statusColor[t.status] || ''" class="text-xs px-2 py-0.5 rounded-full font-medium capitalize">
                  {{ t.status?.replace('_',' ') }}
                </span>
              </td>
              <td class="py-2.5">
                <span :class="priorityColor[t.priority] || ''" class="text-xs px-2 py-0.5 rounded-full font-medium capitalize">
                  {{ t.priority }}
                </span>
              </td>
              <td class="py-2.5 text-slate-500 text-xs">{{ t.requester_name || '-' }}</td>
              <td class="py-2.5 text-xs">{{ t.assigned_agent }}</td>
              <td class="py-2.5 text-xs text-slate-400">{{ t.created_at?.substring(0,10) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-if="!dashboard.recent_tickets?.length" class="text-center text-slate-400 py-8">No recent tickets</div>
      </div>
    </div>
  </div>
</template>
