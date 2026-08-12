<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { FolderKanban, RefreshCw, Calendar, Target, Clock, DollarSign, CheckSquare, Download, BarChart3 } from 'lucide-vue-next'
import { reportsAPI } from '@/api/client'

const app = useAppStore()
const loading = ref(false)
const selectedYear = ref(new Date().getFullYear())
const years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - i)
const data = ref<any>(null)

async function load() {
  loading.value = true
  try { const r = await reportsAPI.getProjectReports(String(selectedYear.value)); data.value = r.data }
  catch { app.addToast('Erreur chargement rapports projets', 'error') }
  finally { loading.value = false }
}
onMounted(load)

const fmt = (n: number) => new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'
const fmtN = (n: number) => new Intl.NumberFormat('fr-DZ').format(n)
const fmtPct = (n: number) => (n || 0).toFixed(1) + '%'

const maxBudget = computed(() => data.value?.projects?.length ? Math.max(...data.value.projects.map((p: any) => p.budget), 1) : 1)

const statusColors: Record<string, string> = {
  in_progress: 'bg-indigo-500/10 text-indigo-500',
  completed: 'bg-emerald-500/10 text-emerald-500',
  planning: 'bg-sky-500/10 text-sky-500',
  on_hold: 'bg-amber-500/10 text-amber-500',
  cancelled: 'bg-red-500/10 text-red-500',
}

function exportCSV() {
  if (!data.value?.projects) return
  const rows = [['Code','Nom','Statut','Budget','Réel','Variance','Heures','H.Facturables','Tâches','%'],
    ...data.value.projects.map((p: any) => [p.code, p.name, p.status, p.budget, p.actual_cost, p.variance, p.total_hours, p.billable_hours, p.total_tasks, p.progress_pct])]
  const a = document.createElement('a')
  a.href = URL.createObjectURL(new Blob(['\ufeff' + rows.map((r: any[]) => r.join(',')).join('\n')], { type: 'text/csv' }))
  a.download = `project_report_${selectedYear.value}.csv`; a.click()
}
</script>

<template>
  <div class="min-h-screen p-6 space-y-6" :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <FolderKanban class="w-6 h-6 text-indigo-500" /> Rapports Projets
        </h1>
        <p class="text-sm mt-0.5" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Budget vs réel, heures, tâches — {{ selectedYear }}</p>
      </div>
      <div class="flex items-center gap-2">
        <div class="flex items-center gap-1.5 rounded-lg px-3 py-2 border text-sm" :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'">
          <Calendar class="w-4 h-4 text-slate-400" />
          <select v-model="selectedYear" @change="load" class="bg-transparent outline-none cursor-pointer">
            <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
          </select>
        </div>
        <button @click="exportCSV" class="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-sm font-medium transition-colors" :class="app.darkMode ? 'bg-slate-900 border-slate-700 hover:bg-slate-800' : 'bg-white border-slate-200 hover:bg-slate-50'">
          <Download class="w-4 h-4" /> CSV
        </button>
        <button @click="load" class="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-sm font-medium transition-colors" :class="app.darkMode ? 'bg-slate-900 border-slate-700 hover:bg-slate-800' : 'bg-white border-slate-200 hover:bg-slate-50'">
          <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
        </button>
      </div>
    </div>

    <div v-if="loading" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="i in 4" :key="i" class="h-24 rounded-xl animate-pulse" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-200'" />
    </div>

    <template v-else-if="data">
      <!-- KPI Cards -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div v-for="kpi in [
          { label: 'Projets Actifs', value: fmtN(data.active_projects), icon: FolderKanban, color: 'text-indigo-500', bg: 'bg-indigo-500/10' },
          { label: 'Terminés', value: fmtN(data.completed_projects), icon: CheckSquare, color: 'text-emerald-500', bg: 'bg-emerald-500/10' },
          { label: 'Budget Total', value: fmt(data.total_budget), icon: DollarSign, color: 'text-sky-500', bg: 'bg-sky-500/10' },
          { label: 'Coût Réel', value: fmt(data.total_actual), icon: Target, color: data.total_actual > data.total_budget ? 'text-red-500' : 'text-emerald-500', bg: data.total_actual > data.total_budget ? 'bg-red-500/10' : 'bg-emerald-500/10' },
          { label: 'Heures Totales', value: fmtN(data.total_hours) + 'h', icon: Clock, color: 'text-violet-500', bg: 'bg-violet-500/10' },
          { label: 'H. Facturables', value: fmtN(data.billable_hours) + 'h', icon: BarChart3, color: 'text-teal-500', bg: 'bg-teal-500/10' },
          { label: 'Taux Facturation', value: fmtPct(data.billable_rate), icon: Target, color: 'text-amber-500', bg: 'bg-amber-500/10' },
          { label: 'Variance Budget', value: fmt(data.budget_variance), icon: BarChart3, color: data.budget_variance >= 0 ? 'text-emerald-500' : 'text-red-500', bg: data.budget_variance >= 0 ? 'bg-emerald-500/10' : 'bg-red-500/10' },
        ]" :key="kpi.label"
          class="rounded-xl border p-4 flex items-center gap-3" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div :class="[kpi.bg, 'w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0']">
            <component :is="kpi.icon" :class="[kpi.color, 'w-5 h-5']" />
          </div>
          <div>
            <p class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ kpi.label }}</p>
            <p class="text-base font-bold mt-0.5">{{ kpi.value }}</p>
          </div>
        </div>
      </div>

      <!-- Status Breakdown -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div class="rounded-xl border p-5" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <h2 class="font-semibold mb-4">Projets par Statut</h2>
          <div class="space-y-3">
            <div v-for="s in data.by_status" :key="s.status">
              <div class="flex items-center justify-between mb-1.5">
                <span :class="[statusColors[s.status] || 'bg-slate-500/10 text-slate-400', 'px-2 py-0.5 rounded-full text-xs font-medium capitalize']">
                  {{ s.status.replace('_',' ') }}
                </span>
                <span class="text-sm font-semibold">{{ s.count }}</span>
              </div>
              <div class="h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                <div class="h-full rounded-full bg-indigo-500" :style="{ width: ((s.count / (data.total_projects || 1)) * 100).toFixed(1) + '%' }" />
              </div>
            </div>
          </div>
        </div>

        <!-- Budget vs Actual Chart -->
        <div class="lg:col-span-2 rounded-xl border p-5" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <h2 class="font-semibold mb-4">Budget vs Réel — Top 8 Projets</h2>
          <div class="space-y-3">
            <div v-for="p in (data.projects || []).slice(0,8)" :key="p.id">
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm truncate max-w-[50%]" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">{{ p.code }} — {{ p.name }}</span>
                <div class="text-right text-xs">
                  <span :class="p.budget_used_pct > 100 ? 'text-red-500' : 'text-slate-400'">{{ fmtPct(p.budget_used_pct) }}</span>
                </div>
              </div>
              <!-- Budget bar -->
              <div class="h-2 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                <div class="h-full rounded-full transition-all duration-500"
                  :class="p.budget_used_pct > 100 ? 'bg-red-500' : p.budget_used_pct > 80 ? 'bg-amber-500' : 'bg-indigo-500'"
                  :style="{ width: Math.min(p.budget_used_pct, 100).toFixed(1) + '%' }" />
              </div>
            </div>
            <p v-if="!data.projects?.length" class="text-sm text-center text-slate-400 py-4">Aucun projet</p>
          </div>
        </div>
      </div>

      <!-- Projects Table -->
      <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <div class="px-5 py-4 border-b" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
          <h2 class="font-semibold">Détail Projets</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b" :class="app.darkMode ? 'border-slate-800 bg-slate-800/40' : 'border-slate-100 bg-slate-50'">
                <th class="text-left px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Projet</th>
                <th class="text-left px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Statut</th>
                <th class="text-right px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Budget</th>
                <th class="text-right px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Réel</th>
                <th class="text-right px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Heures</th>
                <th class="text-right px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Tâches</th>
                <th class="text-right px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Avancement</th>
              </tr>
            </thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
              <tr v-for="p in data.projects" :key="p.id" class="hover:bg-indigo-500/5 transition-colors">
                <td class="px-5 py-3">
                  <p class="font-medium">{{ p.name }}</p>
                  <p class="text-xs text-slate-400">{{ p.code }}</p>
                </td>
                <td class="px-4 py-3">
                  <span :class="[statusColors[p.status] || 'bg-slate-500/10 text-slate-400', 'px-2 py-0.5 rounded-full text-xs font-medium capitalize']">
                    {{ p.status.replace('_',' ') }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right">{{ fmt(p.budget) }}</td>
                <td class="px-4 py-3 text-right font-medium" :class="p.actual_cost > p.budget ? 'text-red-500' : 'text-emerald-500'">{{ fmt(p.actual_cost) }}</td>
                <td class="px-4 py-3 text-right">{{ fmtN(p.total_hours) }}h</td>
                <td class="px-4 py-3 text-right">{{ p.done_tasks }}/{{ p.total_tasks }}</td>
                <td class="px-5 py-3">
                  <div class="flex items-center gap-2 justify-end">
                    <div class="w-16 h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                      <div class="h-full rounded-full bg-indigo-500" :style="{ width: p.progress_pct + '%' }" />
                    </div>
                    <span class="text-xs font-medium">{{ p.progress_pct }}%</span>
                  </div>
                </td>
              </tr>
              <tr v-if="!data.projects?.length"><td colspan="7" class="px-5 py-8 text-center text-slate-400">Aucun projet</td></tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <div v-else-if="!loading" class="text-center py-20">
      <FolderKanban class="w-12 h-12 mx-auto text-slate-300 mb-3" />
      <p class="text-slate-400">Aucune donnée de projet</p>
    </div>
  </div>
</template>
