<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { helpdeskAPI } from '@/api/client'
import { BarChart3, RefreshCw, Calendar, TrendingUp, Users, Star, Timer, CheckCircle2, Tag, Zap } from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(true)
const error = ref('')

const stats = ref<any>({})
const byCategory = ref<any[]>([])
const byAgent = ref<any[]>([])
const bySource = ref<any[]>([])
const dailyTrend = ref<any[]>([])

const dateFrom = ref(new Date(Date.now() - 30 * 86400000).toISOString().substring(0,10))
const dateTo   = ref(new Date().toISOString().substring(0,10))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await helpdeskAPI.getReports({ date_from: dateFrom.value, date_to: dateTo.value })
    stats.value = res.data.stats || {}
    byCategory.value = res.data.by_category || []
    byAgent.value = res.data.by_agent || []
    bySource.value = res.data.by_source || []
    dailyTrend.value = res.data.daily_trend || []
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to load reports'
  } finally {
    loading.value = false
  }
}

onMounted(load)

function maxVal(arr: any[], key: string) {
  return Math.max(...arr.map(a => a[key] || 0), 1)
}

function pct(val: number, max: number) {
  return Math.min(100, Math.round(val / max * 100))
}

const sourceIcon: Record<string, string> = {
  email: 'Mail', phone: 'Phone', chat: 'MessageSquare',
  portal: 'Globe', api: 'Code', internal: 'Building'
}
</script>

<template>
  <div :class="app.darkMode ? 'bg-slate-900 text-slate-100' : 'bg-slate-50 text-slate-900'" class="min-h-screen p-6">
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <BarChart3 class="w-7 h-7 text-indigo-500" />
        <div>
          <h1 class="text-2xl font-bold">Support Reports</h1>
          <p class="text-sm text-slate-500">Analytics and performance insights</p>
        </div>
      </div>
      <button @click="load" :class="app.darkMode ? 'bg-slate-700 text-slate-200' : 'bg-white text-slate-700'"
        class="p-2 rounded-lg border border-slate-200 dark:border-slate-600 hover:bg-slate-100 dark:hover:bg-slate-600 transition-colors">
        <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
      </button>
    </div>

    <div v-if="error" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg text-sm">{{ error }}</div>

    <!-- Date range -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border p-4 mb-6 shadow-sm flex items-center gap-4">
      <Calendar class="w-4 h-4 text-slate-400" />
      <div class="flex items-center gap-2">
        <span class="text-sm text-slate-500">From</span>
        <input type="date" v-model="dateFrom" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
          class="px-3 py-1.5 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
        <span class="text-sm text-slate-500">To</span>
        <input type="date" v-model="dateTo" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
          class="px-3 py-1.5 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
        <button @click="load" class="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors">Apply</button>
      </div>
    </div>

    <!-- Summary KPIs -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'" class="rounded-xl border p-5 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">Total Tickets</span>
          <Zap class="w-5 h-5 text-indigo-500" />
        </div>
        <div class="text-2xl font-bold">{{ stats.total_tickets || 0 }}</div>
        <div class="text-xs text-slate-400 mt-1">{{ stats.resolved || 0 }} resolved, {{ stats.closed || 0 }} closed</div>
      </div>
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'" class="rounded-xl border p-5 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">Avg Resolution</span>
          <Timer class="w-5 h-5 text-purple-500" />
        </div>
        <div class="text-2xl font-bold">{{ (stats.avg_resolution_hrs || 0).toFixed(1) }}h</div>
        <div class="text-xs text-slate-400 mt-1">avg first response: {{ (stats.avg_first_response_hrs || 0).toFixed(1) }}h</div>
      </div>
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'" class="rounded-xl border p-5 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">SLA Compliance</span>
          <CheckCircle2 class="w-5 h-5 text-green-500" />
        </div>
        <div class="text-2xl font-bold" :class="(stats.sla_compliance || 100) >= 90 ? 'text-green-600' : (stats.sla_compliance || 100) >= 70 ? 'text-yellow-600' : 'text-red-600'">
          {{ (stats.sla_compliance || 0).toFixed(1) }}%
        </div>
        <div class="text-xs text-slate-400 mt-1">{{ stats.sla_breached || 0 }} SLA breaches</div>
      </div>
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'" class="rounded-xl border p-5 shadow-sm">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm text-slate-500">CSAT Average</span>
          <Star class="w-5 h-5 text-yellow-500" />
        </div>
        <div class="text-2xl font-bold text-yellow-600">{{ (stats.csat_avg || 0).toFixed(2) }}/5</div>
        <div class="flex gap-0.5 mt-1">
          <Star v-for="i in 5" :key="i" class="w-3.5 h-3.5" :class="i <= Math.round(stats.csat_avg || 0) ? 'text-yellow-400 fill-yellow-400' : 'text-slate-300'" />
        </div>
      </div>
    </div>

    <!-- Charts grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
      <!-- By Category -->
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-5 shadow-sm">
        <h3 class="font-semibold mb-4 flex items-center gap-2">
          <Tag class="w-4 h-4 text-indigo-500" /> Tickets by Category
        </h3>
        <div class="space-y-2">
          <div v-for="c in byCategory" :key="c.category" class="flex items-center gap-2">
            <span class="text-xs text-slate-500 w-28 truncate flex-shrink-0">{{ c.category }}</span>
            <div class="flex-1 bg-slate-100 dark:bg-slate-700 rounded-full h-3">
              <div class="bg-indigo-500 h-3 rounded-full"
                :style="`width:${pct(c.total, maxVal(byCategory,'total'))}%`"></div>
            </div>
            <span class="text-xs font-bold w-8 text-right">{{ c.total }}</span>
            <span class="text-xs text-green-600 w-16 text-right">{{ c.avg_hrs.toFixed(1) }}h avg</span>
          </div>
          <div v-if="!byCategory.length" class="text-center text-slate-400 py-4 text-sm">No data</div>
        </div>
      </div>

      <!-- By Source -->
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="rounded-xl border p-5 shadow-sm">
        <h3 class="font-semibold mb-4 flex items-center gap-2">
          <Zap class="w-4 h-4 text-orange-500" /> Tickets by Source
        </h3>
        <div class="space-y-2">
          <div v-for="s in bySource" :key="s.source" class="flex items-center gap-2">
            <span class="text-xs text-slate-500 w-20 capitalize">{{ s.source }}</span>
            <div class="flex-1 bg-slate-100 dark:bg-slate-700 rounded-full h-3">
              <div class="bg-orange-500 h-3 rounded-full"
                :style="`width:${pct(s.count, maxVal(bySource,'count'))}%`"></div>
            </div>
            <span class="text-xs font-bold w-8 text-right">{{ s.count }}</span>
            <span class="text-xs text-slate-400 w-12 text-right">{{ stats.total_tickets ? (s.count/stats.total_tickets*100).toFixed(0) : 0 }}%</span>
          </div>
          <div v-if="!bySource.length" class="text-center text-slate-400 py-4 text-sm">No data</div>
        </div>
      </div>
    </div>

    <!-- Daily trend table -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border p-5 shadow-sm mb-6">
      <h3 class="font-semibold mb-4 flex items-center gap-2">
        <TrendingUp class="w-4 h-4 text-blue-500" /> Daily Trend
      </h3>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'" class="text-xs uppercase">
              <th class="text-left pb-2">Date</th>
              <th class="text-center pb-2">Opened</th>
              <th class="text-center pb-2">Resolved</th>
              <th class="text-center pb-2">Closed</th>
              <th class="text-center pb-2">Net</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="d in dailyTrend" :key="d.date"
              :class="app.darkMode ? 'border-slate-700' : 'border-slate-100'" class="border-b">
              <td class="py-1.5 text-xs text-slate-500">{{ d.date }}</td>
              <td class="py-1.5 text-center font-medium text-blue-600">{{ d.opened }}</td>
              <td class="py-1.5 text-center font-medium text-green-600">{{ d.resolved }}</td>
              <td class="py-1.5 text-center font-medium text-slate-500">{{ d.closed }}</td>
              <td class="py-1.5 text-center">
                <span :class="(d.opened-d.resolved)>0?'text-red-500':'text-green-500'" class="font-medium">
                  {{ d.opened-d.resolved > 0 ? '+' : '' }}{{ d.opened - d.resolved }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="!dailyTrend.length" class="text-center text-slate-400 py-4">No trend data</div>
      </div>
    </div>

    <!-- Agent performance -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border p-5 shadow-sm">
      <h3 class="font-semibold mb-4 flex items-center gap-2">
        <Users class="w-4 h-4 text-teal-500" /> Agent Performance
      </h3>
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'text-slate-400 border-slate-700' : 'text-slate-500 border-slate-200'"
              class="border-b text-xs uppercase tracking-wide">
              <th class="text-left pb-3 font-medium">Agent</th>
              <th class="text-center pb-3 font-medium">Total</th>
              <th class="text-center pb-3 font-medium">Resolved</th>
              <th class="text-center pb-3 font-medium">Resolution Rate</th>
              <th class="text-center pb-3 font-medium">Avg Resolution</th>
              <th class="text-center pb-3 font-medium">CSAT Avg</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="a in byAgent" :key="a.agent_name"
              :class="app.darkMode ? 'border-slate-700 hover:bg-slate-700/50' : 'border-slate-100 hover:bg-slate-50'"
              class="border-b transition-colors">
              <td class="py-2.5 font-medium">{{ a.agent_name }}</td>
              <td class="py-2.5 text-center">{{ a.total }}</td>
              <td class="py-2.5 text-center text-green-600 font-medium">{{ a.resolved }}</td>
              <td class="py-2.5 text-center">
                <span :class="a.total > 0 && a.resolved/a.total >= 0.8 ? 'text-green-600' : 'text-yellow-600'" class="font-medium">
                  {{ a.total > 0 ? (a.resolved/a.total*100).toFixed(0) : 0 }}%
                </span>
              </td>
              <td class="py-2.5 text-center">{{ a.avg_hrs.toFixed(1) }}h</td>
              <td class="py-2.5 text-center">
                <div class="flex items-center justify-center gap-1">
                  <Star class="w-3.5 h-3.5 text-yellow-400 fill-yellow-400" />
                  <span class="font-medium">{{ a.csat_avg.toFixed(2) }}</span>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="!byAgent.length" class="text-center py-8 text-slate-400">No agent data</div>
      </div>
    </div>
  </div>
</template>
