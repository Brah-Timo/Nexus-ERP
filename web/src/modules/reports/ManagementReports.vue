<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { BarChart3, RefreshCw, Calendar, Users, TrendingUp, TrendingDown, DollarSign, Target } from 'lucide-vue-next'
import { reportsAPI } from '@/api/client'

const app = useAppStore()
const loading = ref(false)
const selectedYear = ref(new Date().getFullYear())
const years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - i)
const data = ref<any>(null)

async function load() {
  loading.value = true
  try { const r = await reportsAPI.getManagementReports(String(selectedYear.value)); data.value = r.data }
  catch { app.addToast('Erreur chargement rapport direction', 'error') }
  finally { loading.value = false }
}
onMounted(load)

const fmt = (n: number) => new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'
const fmtN = (n: number) => new Intl.NumberFormat('fr-DZ').format(n)
const fmtPct = (n: number) => (n || 0).toFixed(1) + '%'
const fmtX = (n: number) => (n || 0).toFixed(1)
</script>

<template>
  <div class="min-h-screen p-6 space-y-6" :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <BarChart3 class="w-6 h-6 text-sky-500" /> Rapport Direction
        </h1>
        <p class="text-sm mt-0.5" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
          Tableau de bord exécutif, P&L trimestriel, KPIs stratégiques — {{ selectedYear }}
        </p>
      </div>
      <div class="flex items-center gap-2">
        <div class="flex items-center gap-1.5 rounded-lg px-3 py-2 border text-sm" :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'">
          <Calendar class="w-4 h-4 text-slate-400" />
          <select v-model="selectedYear" @change="load" class="bg-transparent outline-none cursor-pointer">
            <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
          </select>
        </div>
        <button @click="load" class="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-sm font-medium transition-colors" :class="app.darkMode ? 'bg-slate-900 border-slate-700 hover:bg-slate-800' : 'bg-white border-slate-200 hover:bg-slate-50'">
          <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
        </button>
      </div>
    </div>

    <div v-if="loading" class="space-y-4">
      <div v-for="i in 3" :key="i" class="h-32 rounded-xl animate-pulse" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-200'" />
    </div>

    <template v-else-if="data">

      <!-- Executive Summary Banner -->
      <div class="rounded-xl p-6 bg-gradient-to-r from-sky-600 to-indigo-700 text-white">
        <div class="flex items-center justify-between flex-wrap gap-4">
          <div>
            <p class="text-sky-200 text-sm uppercase tracking-wide font-medium">Performance {{ selectedYear }}</p>
            <p class="text-4xl font-bold mt-1">{{ fmt(data.revenue) }}</p>
            <p class="text-sky-200 text-sm mt-1">Chiffre d'affaires annuel</p>
          </div>
          <div class="grid grid-cols-3 gap-6 text-center">
            <div>
              <p class="text-sky-200 text-xs uppercase tracking-wide">Croissance</p>
              <p class="text-2xl font-bold mt-0.5" :class="data.revenue_growth_pct >= 0 ? 'text-emerald-300' : 'text-red-300'">
                {{ data.revenue_growth_pct >= 0 ? '+' : '' }}{{ fmtPct(data.revenue_growth_pct) }}
              </p>
              <p class="text-xs text-sky-200">vs {{ selectedYear - 1 }}</p>
            </div>
            <div>
              <p class="text-sky-200 text-xs uppercase tracking-wide">Marge Nette</p>
              <p class="text-2xl font-bold mt-0.5" :class="data.net_margin_pct >= 0 ? 'text-emerald-300' : 'text-red-300'">
                {{ fmtPct(data.net_margin_pct) }}
              </p>
            </div>
            <div>
              <p class="text-sky-200 text-xs uppercase tracking-wide">Résultat Net</p>
              <p class="text-2xl font-bold mt-0.5" :class="data.net_profit >= 0 ? 'text-emerald-300' : 'text-red-300'">
                {{ new Intl.NumberFormat('fr-DZ', { notation: 'compact', maximumSignificantDigits: 3 }).format(data.net_profit) }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Key Metrics Grid -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div v-for="kpi in [
          { label: 'Effectif', value: fmtN(data.employees), sub: 'Employés actifs', icon: Users, color: 'text-sky-500', bg: 'bg-sky-500/10' },
          { label: 'Masse Salariale', value: fmt(data.payroll), sub: 'Coût total annuel', icon: DollarSign, color: 'text-orange-500', bg: 'bg-orange-500/10' },
          { label: 'Revenu/Employé', value: fmt(data.revenue_per_employee), sub: 'Productivité', icon: TrendingUp, color: 'text-emerald-500', bg: 'bg-emerald-500/10' },
          { label: 'Charges', value: fmt(data.expenses), sub: 'Achats + charges', icon: TrendingDown, color: 'text-red-500', bg: 'bg-red-500/10' },
        ]" :key="kpi.label"
          class="rounded-xl border p-4 flex items-center gap-3" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div :class="[kpi.bg, 'w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0']">
            <component :is="kpi.icon" :class="[kpi.color, 'w-5 h-5']" />
          </div>
          <div>
            <p class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ kpi.label }}</p>
            <p class="text-sm font-bold mt-0.5">{{ kpi.value }}</p>
            <p class="text-xs" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ kpi.sub }}</p>
          </div>
        </div>
      </div>

      <!-- Quarterly P&L table -->
      <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <div class="px-5 py-4 border-b flex items-center gap-2" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
          <BarChart3 class="w-4 h-4 text-sky-500" />
          <h2 class="font-semibold">Compte de Résultat Trimestriel {{ selectedYear }}</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b" :class="app.darkMode ? 'border-slate-800 bg-slate-800/40' : 'border-slate-100 bg-slate-50'">
                <th class="text-left px-5 py-3 text-xs font-semibold uppercase tracking-wide text-slate-400">Trimestre</th>
                <th class="text-right px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-400">Revenus</th>
                <th class="text-right px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-400">Charges</th>
                <th class="text-right px-4 py-3 text-xs font-semibold uppercase tracking-wide text-slate-400">Résultat Net</th>
                <th class="text-right px-5 py-3 text-xs font-semibold uppercase tracking-wide text-slate-400">Marge</th>
              </tr>
            </thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
              <tr v-for="q in data.quarterly_pl" :key="q.quarter" class="hover:bg-sky-500/5 transition-colors">
                <td class="px-5 py-3 font-semibold">T{{ q.quarter }} {{ selectedYear }}</td>
                <td class="px-4 py-3 text-right font-semibold text-emerald-500">{{ fmt(q.revenue) }}</td>
                <td class="px-4 py-3 text-right text-red-400">{{ fmt(q.expenses) }}</td>
                <td class="px-4 py-3 text-right font-bold" :class="q.net_profit >= 0 ? 'text-emerald-500' : 'text-red-500'">{{ fmt(q.net_profit) }}</td>
                <td class="px-5 py-3 text-right">
                  <span class="px-2 py-0.5 rounded-full text-xs font-medium"
                    :class="q.margin_pct >= 15 ? 'bg-emerald-500/10 text-emerald-500' : q.margin_pct >= 0 ? 'bg-amber-500/10 text-amber-500' : 'bg-red-500/10 text-red-500'">
                    {{ fmtPct(q.margin_pct) }}
                  </span>
                </td>
              </tr>
            </tbody>
            <tfoot>
              <tr class="border-t font-bold" :class="app.darkMode ? 'border-slate-700 bg-slate-800/50' : 'border-slate-200 bg-slate-50'">
                <td class="px-5 py-3">Total {{ selectedYear }}</td>
                <td class="px-4 py-3 text-right text-emerald-500">{{ fmt(data.revenue) }}</td>
                <td class="px-4 py-3 text-right text-red-400">{{ fmt(data.expenses) }}</td>
                <td class="px-4 py-3 text-right" :class="data.net_profit >= 0 ? 'text-emerald-500' : 'text-red-500'">{{ fmt(data.net_profit) }}</td>
                <td class="px-5 py-3 text-right">{{ fmtPct(data.net_margin_pct) }}</td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>

      <!-- Scorecard + Department Costs -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

        <!-- Operational Scorecard -->
        <div class="rounded-xl border p-5" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <h2 class="font-semibold mb-4 flex items-center gap-2">
            <Target class="w-4 h-4 text-sky-500" /> Tableau de Bord Opérationnel
          </h2>
          <div class="space-y-4">
            <div v-for="m in [
              { label: 'DSO — Délai Recouvrement Client', value: fmtX(data.ar_days) + ' jours', good: data.ar_days <= 45, target: '≤ 45j' },
              { label: 'DPO — Délai Paiement Fournisseur', value: fmtX(data.ap_days) + ' jours', good: data.ap_days >= 30, target: '30-60j' },
              { label: 'Rotation des Stocks', value: fmtX(data.inventory_turnover) + 'x/an', good: data.inventory_turnover >= 4, target: '≥ 4x' },
              { label: 'Croissance CA annuelle', value: (data.revenue_growth_pct >= 0 ? '+' : '') + fmtPct(data.revenue_growth_pct), good: data.revenue_growth_pct >= 5, target: '≥ 5%' },
            ]" :key="m.label" class="flex items-center justify-between py-2 border-b" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
              <div>
                <p class="text-sm font-medium" :class="app.darkMode ? 'text-slate-200' : 'text-slate-700'">{{ m.label }}</p>
                <p class="text-xs" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">Cible: {{ m.target }}</p>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-sm font-bold" :class="m.good ? 'text-emerald-500' : 'text-amber-500'">{{ m.value }}</span>
                <span class="w-5 h-5 rounded-full flex items-center justify-center text-xs"
                  :class="m.good ? 'bg-emerald-500/10 text-emerald-500' : 'bg-amber-500/10 text-amber-500'">
                  {{ m.good ? '✓' : '!' }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Department Costs -->
        <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="px-5 py-4 border-b" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
            <h2 class="font-semibold flex items-center gap-2">
              <Users class="w-4 h-4 text-sky-500" /> Coûts par Département
            </h2>
          </div>
          <div class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
            <div v-for="d in data.department_costs" :key="d.department" class="px-5 py-3">
              <div class="flex items-center justify-between mb-1.5">
                <span class="text-sm font-medium">{{ d.department }}</span>
                <div class="text-right">
                  <p class="text-sm font-semibold">{{ fmt(d.payroll_cost) }}</p>
                  <p class="text-xs text-slate-400">{{ d.headcount }} pers.</p>
                </div>
              </div>
              <div class="h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                <div class="h-full rounded-full bg-sky-500 transition-all duration-500"
                  :style="{ width: ((d.payroll_cost / (data.payroll || 1)) * 100).toFixed(1) + '%' }" />
              </div>
            </div>
            <p v-if="!data.department_costs?.length" class="px-5 py-6 text-center text-sm text-slate-400">Aucun département</p>
          </div>
        </div>
      </div>
    </template>

    <div v-else-if="!loading" class="text-center py-20">
      <BarChart3 class="w-12 h-12 mx-auto text-slate-300 mb-3" />
      <p class="text-slate-400">Aucune donnée disponible</p>
    </div>
  </div>
</template>
