<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { qualityAPI } from '@/api/client'
import {
  ShieldCheck, ClipboardList, AlertTriangle, Wrench,
  TrendingUp, TrendingDown, CheckCircle, XCircle,
  Clock, RefreshCw, BarChart3, Activity,
  AlertCircle, Target, Percent, ThumbsUp
} from 'lucide-vue-next'

const app = useAppStore()
const dk = (d: string, l: string) => app.darkMode ? d : l

const loading = ref(false)
const stats = ref<any>({})
const recentInspections = ref<any[]>([])
const recentNC = ref<any[]>([])
const monthlyTrend = ref<any[]>([])
const ncBySeverity = ref<any[]>([])

async function load() {
  loading.value = true
  try {
    const res = await qualityAPI.getDashboard()
    const d = res.data
    stats.value          = d.stats || {}
    recentInspections.value = d.recent_inspections || []
    recentNC.value       = d.recent_nc || []
    monthlyTrend.value   = d.monthly_trend || []
    ncBySeverity.value   = d.nc_by_severity || []
  } catch {}
  loading.value = false
}

onMounted(load)

const statCards = computed(() => [
  {
    label: 'Total Inspections',
    value: stats.value.total_inspections ?? 0,
    sub: `${stats.value.pending_inspections ?? 0} pending`,
    icon: ClipboardList,
    color: 'indigo',
    trend: null
  },
  {
    label: 'First Pass Rate',
    value: `${(stats.value.first_pass_rate ?? 0).toFixed(1)}%`,
    sub: `${stats.value.passed_inspections ?? 0} passed / ${stats.value.failed_inspections ?? 0} failed`,
    icon: Target,
    color: (stats.value.first_pass_rate ?? 0) >= 90 ? 'emerald' : (stats.value.first_pass_rate ?? 0) >= 70 ? 'amber' : 'red',
    trend: null
  },
  {
    label: 'Open Non-Conformities',
    value: stats.value.open_nc ?? 0,
    sub: `${stats.value.critical_nc ?? 0} critical`,
    icon: AlertTriangle,
    color: (stats.value.critical_nc ?? 0) > 0 ? 'red' : (stats.value.open_nc ?? 0) > 0 ? 'amber' : 'emerald',
    trend: null
  },
  {
    label: 'Open Corrective Actions',
    value: stats.value.open_ca ?? 0,
    sub: `${stats.value.overdue_ca ?? 0} overdue`,
    icon: Wrench,
    color: (stats.value.overdue_ca ?? 0) > 0 ? 'red' : (stats.value.open_ca ?? 0) > 0 ? 'amber' : 'emerald',
    trend: null
  },
  {
    label: 'Defect Rate',
    value: `${(stats.value.defect_rate ?? 0).toFixed(2)}%`,
    sub: `${stats.value.failed_checks ?? 0} / ${stats.value.total_checks ?? 0} checks failed`,
    icon: Percent,
    color: (stats.value.defect_rate ?? 0) > 5 ? 'red' : (stats.value.defect_rate ?? 0) > 2 ? 'amber' : 'emerald',
    trend: null
  },
  {
    label: 'Total NC',
    value: stats.value.total_nc ?? 0,
    sub: `${stats.value.total_ca ?? 0} corrective actions`,
    icon: Activity,
    color: 'slate',
    trend: null
  },
])

const colorMap: Record<string, string> = {
  indigo:  'bg-indigo-500',
  emerald: 'bg-emerald-500',
  amber:   'bg-amber-500',
  red:     'bg-red-500',
  slate:   'bg-slate-500',
  sky:     'bg-sky-500',
  violet:  'bg-violet-500',
}
const textColorMap: Record<string, string> = {
  indigo:  'text-indigo-400',
  emerald: 'text-emerald-400',
  amber:   'text-amber-400',
  red:     'text-red-400',
  slate:   'text-slate-400',
  sky:     'text-sky-400',
  violet:  'text-violet-400',
}
const bgLightMap: Record<string, string> = {
  indigo:  'bg-indigo-50',
  emerald: 'bg-emerald-50',
  amber:   'bg-amber-50',
  red:     'bg-red-50',
  slate:   'bg-slate-50',
  sky:     'bg-sky-50',
  violet:  'bg-violet-50',
}
const textLightMap: Record<string, string> = {
  indigo:  'text-indigo-600',
  emerald: 'text-emerald-600',
  amber:   'text-amber-600',
  red:     'text-red-600',
  slate:   'text-slate-600',
  sky:     'text-sky-600',
  violet:  'text-violet-600',
}

function inspStatusBadge(status: string) {
  const map: Record<string, string> = {
    pending:     dk('bg-slate-700 text-slate-200','bg-slate-100 text-slate-600'),
    in_progress: dk('bg-sky-900 text-sky-200','bg-sky-100 text-sky-700'),
    passed:      dk('bg-emerald-900 text-emerald-200','bg-emerald-100 text-emerald-700'),
    failed:      dk('bg-red-900 text-red-200','bg-red-100 text-red-700'),
    cancelled:   dk('bg-zinc-800 text-zinc-400','bg-zinc-100 text-zinc-500'),
  }
  return map[status] ?? dk('bg-slate-700 text-slate-200','bg-slate-100 text-slate-600')
}

function ncSeverityBadge(sev: string) {
  const map: Record<string, string> = {
    minor:           dk('bg-sky-900 text-sky-200','bg-sky-100 text-sky-700'),
    major:           dk('bg-amber-900 text-amber-200','bg-amber-100 text-amber-700'),
    critical:        dk('bg-red-900 text-red-200','bg-red-100 text-red-700'),
    critical_safety: dk('bg-red-950 text-red-300','bg-red-200 text-red-800'),
  }
  return map[sev] ?? dk('bg-slate-700 text-slate-200','bg-slate-100 text-slate-600')
}

function fmtDate(v: string) {
  if (!v) return '-'
  return new Date(v).toLocaleDateString('en-GB', { day:'2-digit', month:'short', year:'numeric' })
}

function maxTrend(field: string) {
  return Math.max(...monthlyTrend.value.map((m: any) => m[field] || 0), 1)
}

const severityColors: Record<string, string> = {
  minor:           'bg-sky-500',
  major:           'bg-amber-500',
  critical:        'bg-red-500',
  critical_safety: 'bg-red-700',
}
</script>

<template>
  <div :class="dk('bg-slate-900 text-slate-100','bg-slate-50 text-slate-900')" class="min-h-screen">
    <!-- Header -->
    <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
         class="border-b px-6 py-4 flex items-center justify-between sticky top-0 z-10">
      <div class="flex items-center gap-3">
        <div class="w-9 h-9 rounded-lg bg-indigo-600 flex items-center justify-center">
          <ShieldCheck class="w-5 h-5 text-white" />
        </div>
        <div>
          <h1 class="text-lg font-bold">Quality Dashboard</h1>
          <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">Quality Management Overview</p>
        </div>
      </div>
      <button @click="load"
        :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-200','bg-white hover:bg-slate-50 text-slate-700 border border-slate-200')"
        class="flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm transition-colors">
        <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" /> Refresh
      </button>
    </div>

    <div class="p-6 space-y-6">

      <!-- KPI Cards -->
      <div class="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-4">
        <div v-for="card in statCards" :key="card.label"
          :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="rounded-xl border p-4">
          <div class="flex items-center justify-between mb-3">
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs font-medium truncate">{{ card.label }}</p>
            <div :class="app.darkMode ? colorMap[card.color]+' bg-opacity-20' : bgLightMap[card.color]"
                 class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0">
              <component :is="card.icon" class="w-4 h-4"
                :class="app.darkMode ? textColorMap[card.color] : textLightMap[card.color]" />
            </div>
          </div>
          <p class="text-2xl font-bold tabular-nums">{{ card.value }}</p>
          <p :class="dk('text-slate-500','text-slate-400')" class="text-xs mt-1">{{ card.sub }}</p>
        </div>
      </div>

      <!-- Main Grid: Trend + NC Severity -->
      <div class="grid grid-cols-1 xl:grid-cols-3 gap-6">

        <!-- Monthly Trend Bar Chart -->
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
             class="xl:col-span-2 rounded-xl border p-5">
          <div class="flex items-center justify-between mb-5">
            <div class="flex items-center gap-2">
              <BarChart3 class="w-4 h-4 text-indigo-400" />
              <h2 class="text-sm font-semibold">Inspection Trend (Last 6 Months)</h2>
            </div>
            <div class="flex items-center gap-3 text-xs">
              <span class="flex items-center gap-1"><span class="w-2.5 h-2.5 rounded-sm bg-emerald-500 inline-block"></span>Passed</span>
              <span class="flex items-center gap-1"><span class="w-2.5 h-2.5 rounded-sm bg-red-500 inline-block"></span>Failed</span>
            </div>
          </div>
          <div v-if="monthlyTrend.length === 0"
               :class="dk('text-slate-500','text-slate-400')"
               class="text-center py-12 text-sm">No data for the last 6 months</div>
          <div v-else class="flex items-end gap-3 h-48">
            <div v-for="m in monthlyTrend" :key="m.month"
                 class="flex-1 flex flex-col items-center gap-1">
              <span :class="dk('text-slate-400','text-slate-500')" class="text-xs tabular-nums">
                {{ m.pass_rate ? m.pass_rate.toFixed(0)+'%' : '' }}
              </span>
              <div class="w-full flex flex-col gap-0.5 justify-end" style="height:120px">
                <div class="w-full bg-emerald-500 rounded-t"
                     :style="`height:${Math.round((m.passed/Math.max(m.inspections,1))*100)}%`"></div>
                <div class="w-full bg-red-500"
                     :style="`height:${Math.round((m.failed/Math.max(m.inspections,1))*100)}%`"></div>
              </div>
              <span :class="dk('text-slate-500','text-slate-400')" class="text-xs">{{ m.month.slice(5) }}</span>
              <span :class="dk('text-slate-400','text-slate-600')" class="text-xs font-medium">{{ m.inspections }}</span>
            </div>
          </div>
        </div>

        <!-- NC by Severity -->
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
             class="rounded-xl border p-5">
          <div class="flex items-center gap-2 mb-5">
            <AlertTriangle class="w-4 h-4 text-amber-400" />
            <h2 class="text-sm font-semibold">Open NC by Severity</h2>
          </div>
          <div v-if="ncBySeverity.length === 0"
               :class="dk('text-slate-500','text-slate-400')"
               class="text-center py-12 text-sm flex flex-col items-center gap-2">
            <CheckCircle class="w-8 h-8 text-emerald-500" />
            No open non-conformities
          </div>
          <div v-else class="space-y-4">
            <div v-for="sv in ncBySeverity" :key="sv.severity">
              <div class="flex items-center justify-between mb-1.5">
                <span class="text-sm capitalize">{{ sv.severity.replace('_',' ') }}</span>
                <span class="text-sm font-bold tabular-nums">{{ sv.count }}</span>
              </div>
              <div :class="dk('bg-slate-700','bg-slate-100')" class="h-2 rounded-full overflow-hidden">
                <div :class="severityColors[sv.severity] || 'bg-slate-500'"
                     class="h-full rounded-full transition-all duration-500"
                     :style="`width:${Math.min((sv.count/Math.max(...ncBySeverity.map((x:any)=>x.count),1))*100,100)}%`">
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Bottom Grid: Recent Inspections + Recent NC -->
      <div class="grid grid-cols-1 xl:grid-cols-2 gap-6">

        <!-- Recent Inspections -->
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
             class="rounded-xl border">
          <div class="flex items-center justify-between px-5 py-4 border-b"
               :class="dk('border-slate-700','border-slate-200')">
            <div class="flex items-center gap-2">
              <ClipboardList class="w-4 h-4 text-indigo-400" />
              <h2 class="text-sm font-semibold">Recent Inspections</h2>
            </div>
            <router-link to="/quality/inspections"
              :class="dk('text-indigo-400 hover:text-indigo-300','text-indigo-600 hover:text-indigo-700')"
              class="text-xs font-medium transition-colors">View all</router-link>
          </div>
          <div class="divide-y" :class="dk('divide-slate-700','divide-slate-100')">
            <div v-if="recentInspections.length === 0"
                 :class="dk('text-slate-500','text-slate-400')"
                 class="text-center py-8 text-sm">No recent inspections</div>
            <div v-for="ins in recentInspections" :key="ins.id"
                 class="px-5 py-3 flex items-center gap-3">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium">{{ ins.reference }}</span>
                  <span :class="inspStatusBadge(ins.status)"
                        class="px-1.5 py-0.5 rounded text-xs font-medium">
                    {{ ins.status.replace('_',' ') }}
                  </span>
                </div>
                <div class="flex items-center gap-3 mt-0.5">
                  <span :class="dk('text-slate-400','text-slate-500')" class="text-xs truncate">
                    {{ ins.item_name || ins.inspection_type }}
                  </span>
                  <span :class="dk('text-slate-500','text-slate-400')" class="text-xs">
                    {{ fmtDate(ins.created_at) }}
                  </span>
                </div>
              </div>
              <div class="text-right flex-shrink-0">
                <div class="flex items-center gap-1.5 text-xs">
                  <span class="text-emerald-500 font-medium">{{ ins.qty_passed }}</span>
                  <span :class="dk('text-slate-500','text-slate-400')">/</span>
                  <span :class="dk('text-slate-300','text-slate-600')">{{ ins.qty_to_inspect }}</span>
                </div>
                <span :class="dk('text-slate-500','text-slate-400')" class="text-xs">
                  {{ ins.inspector_name || '-' }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Critical / Open NCs -->
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
             class="rounded-xl border">
          <div class="flex items-center justify-between px-5 py-4 border-b"
               :class="dk('border-slate-700','border-slate-200')">
            <div class="flex items-center gap-2">
              <AlertTriangle class="w-4 h-4 text-amber-400" />
              <h2 class="text-sm font-semibold">Open Non-Conformities</h2>
            </div>
            <router-link to="/quality/non-conformities"
              :class="dk('text-indigo-400 hover:text-indigo-300','text-indigo-600 hover:text-indigo-700')"
              class="text-xs font-medium transition-colors">View all</router-link>
          </div>
          <div class="divide-y" :class="dk('divide-slate-700','divide-slate-100')">
            <div v-if="recentNC.length === 0"
                 :class="dk('text-slate-500','text-slate-400')"
                 class="text-center py-8 text-sm flex flex-col items-center gap-2">
              <ThumbsUp class="w-8 h-8 text-emerald-500" />
              No open non-conformities
            </div>
            <div v-for="nc in recentNC" :key="nc.id"
                 class="px-5 py-3">
              <div class="flex items-start justify-between gap-2">
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2 flex-wrap">
                    <span class="text-sm font-medium">{{ nc.reference }}</span>
                    <span :class="ncSeverityBadge(nc.severity)"
                          class="px-1.5 py-0.5 rounded text-xs font-medium capitalize">
                      {{ nc.severity.replace('_',' ') }}
                    </span>
                  </div>
                  <p :class="dk('text-slate-300','text-slate-600')" class="text-xs mt-0.5 truncate">{{ nc.title }}</p>
                  <div class="flex items-center gap-2 mt-1">
                    <span :class="dk('text-slate-500','text-slate-400')" class="text-xs">
                      {{ fmtDate(nc.detected_date) }}
                    </span>
                    <span v-if="nc.assignee_name" :class="dk('text-slate-500','text-slate-400')" class="text-xs">
                      — {{ nc.assignee_name }}
                    </span>
                  </div>
                </div>
                <span :class="dk('text-slate-400 border-slate-600','text-slate-500 border-slate-200')"
                      class="text-xs px-2 py-0.5 rounded border flex-shrink-0 capitalize">
                  {{ nc.status.replace('_',' ') }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Quick Links -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <router-link to="/quality/inspections"
          :class="dk('bg-slate-800 border-slate-700 hover:border-indigo-500','bg-white border-slate-200 hover:border-indigo-400')"
          class="border rounded-xl p-4 flex items-center gap-3 transition-colors group">
          <div class="w-10 h-10 rounded-lg bg-indigo-600 bg-opacity-20 flex items-center justify-center">
            <ClipboardList class="w-5 h-5 text-indigo-400" />
          </div>
          <div>
            <p class="text-sm font-semibold">Inspections</p>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">{{ stats.total_inspections ?? 0 }} total</p>
          </div>
        </router-link>
        <router-link to="/quality/checks"
          :class="dk('bg-slate-800 border-slate-700 hover:border-emerald-500','bg-white border-slate-200 hover:border-emerald-400')"
          class="border rounded-xl p-4 flex items-center gap-3 transition-colors group">
          <div class="w-10 h-10 rounded-lg bg-emerald-600 bg-opacity-20 flex items-center justify-center">
            <CheckCircle class="w-5 h-5 text-emerald-400" />
          </div>
          <div>
            <p class="text-sm font-semibold">Quality Checks</p>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">{{ stats.total_checks ?? 0 }} checks</p>
          </div>
        </router-link>
        <router-link to="/quality/non-conformities"
          :class="dk('bg-slate-800 border-slate-700 hover:border-amber-500','bg-white border-slate-200 hover:border-amber-400')"
          class="border rounded-xl p-4 flex items-center gap-3 transition-colors group">
          <div class="w-10 h-10 rounded-lg bg-amber-600 bg-opacity-20 flex items-center justify-center">
            <AlertTriangle class="w-5 h-5 text-amber-400" />
          </div>
          <div>
            <p class="text-sm font-semibold">Non-Conformities</p>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">{{ stats.open_nc ?? 0 }} open</p>
          </div>
        </router-link>
        <router-link to="/quality/corrective-actions"
          :class="dk('bg-slate-800 border-slate-700 hover:border-sky-500','bg-white border-slate-200 hover:border-sky-400')"
          class="border rounded-xl p-4 flex items-center gap-3 transition-colors group">
          <div class="w-10 h-10 rounded-lg bg-sky-600 bg-opacity-20 flex items-center justify-center">
            <Wrench class="w-5 h-5 text-sky-400" />
          </div>
          <div>
            <p class="text-sm font-semibold">Corrective Actions</p>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">{{ stats.open_ca ?? 0 }} open</p>
          </div>
        </router-link>
      </div>

    </div>
  </div>
</template>
