<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import {
  FileBarChart, RefreshCw, Calendar, TrendingUp, TrendingDown,
  DollarSign, Scale, PiggyBank, BarChart3, Download
} from 'lucide-vue-next'
import { reportsAPI } from '@/api/client'

const app = useAppStore()
const loading = ref(false)
const selectedYear = ref(new Date().getFullYear())
const years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - i)

interface FinancialData {
  year: number
  revenue: number; tva_collected: number; cogs: number
  gross_profit: number; gross_margin_pct: number; opex: number
  ebitda: number; net_profit: number; net_margin_pct: number
  total_assets: number; total_liabilities: number; equity: number
  current_ratio: number; debt_ratio: number; roe: number; roa: number
  quarters: Array<{ quarter: number; revenue: number; cost: number; gross_profit: number; margin_pct: number }>
  cash_flow: Array<{ month: number; inflow: number; outflow: number; net: number }>
}

const data = ref<FinancialData | null>(null)

const months = ['Jan','Fév','Mar','Avr','Mai','Jun','Jul','Aoû','Sep','Oct','Nov','Déc']

async function load() {
  loading.value = true
  try {
    const r = await reportsAPI.getFinancialReports(String(selectedYear.value))
    data.value = r.data
  } catch {
    app.addToast('Erreur chargement rapports financiers', 'error')
  } finally {
    loading.value = false
  }
}
onMounted(load)

const fmt = (n: number) =>
  new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(n) + ' DZD'
const fmtPct = (n: number) => (n || 0).toFixed(1) + '%'
const fmtX = (n: number) => (n || 0).toFixed(2) + 'x'

// Cash flow chart
const maxCF = computed(() => {
  if (!data.value) return 1
  return Math.max(...data.value.cash_flow.flatMap(m => [m.inflow, m.outflow]), 1)
})
function cfBarH(val: number) {
  return Math.max((val / maxCF.value) * 100, 1).toFixed(1) + '%'
}

function exportCSV() {
  if (!data.value) return
  const rows = [
    ['Indicateur','Valeur'],
    ['CA HT', data.value.revenue],
    ['TVA Collectée', data.value.tva_collected],
    ['Achats', data.value.cogs],
    ['Marge Brute', data.value.gross_profit],
    ['Marge Brute %', data.value.gross_margin_pct],
    ['Charges Exploitation', data.value.opex],
    ['EBITDA', data.value.ebitda],
    ['Résultat Net', data.value.net_profit],
    ['Marge Nette %', data.value.net_margin_pct],
    ['Total Actif', data.value.total_assets],
    ['Total Passif', data.value.total_liabilities],
    ['Capitaux Propres', data.value.equity],
    ['Ratio Courant', data.value.current_ratio],
    ['Ratio Endettement %', data.value.debt_ratio],
    ['ROE %', data.value.roe],
    ['ROA %', data.value.roa],
  ]
  const csv = rows.map(r => r.join(',')).join('\n')
  const a = document.createElement('a')
  a.href = URL.createObjectURL(new Blob(['\ufeff' + csv], { type: 'text/csv' }))
  a.download = `financial_report_${selectedYear.value}.csv`
  a.click()
}
</script>

<template>
  <div class="min-h-screen p-6 space-y-6"
    :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">

    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <FileBarChart class="w-6 h-6 text-indigo-500" />
          Rapports Financiers
        </h1>
        <p class="text-sm mt-0.5" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
          Compte de résultat, bilan, ratios financiers
        </p>
      </div>
      <div class="flex items-center gap-2">
        <div class="flex items-center gap-1.5 rounded-lg px-3 py-2 border text-sm"
          :class="app.darkMode ? 'bg-slate-900 border-slate-700' : 'bg-white border-slate-200'">
          <Calendar class="w-4 h-4 text-slate-400" />
          <select v-model="selectedYear" @change="load" class="bg-transparent outline-none cursor-pointer">
            <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
          </select>
        </div>
        <button @click="exportCSV"
          class="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-sm font-medium transition-colors"
          :class="app.darkMode ? 'bg-slate-900 border-slate-700 hover:bg-slate-800' : 'bg-white border-slate-200 hover:bg-slate-50'">
          <Download class="w-4 h-4" /> Export CSV
        </button>
        <button @click="load"
          class="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-sm font-medium transition-colors"
          :class="app.darkMode ? 'bg-slate-900 border-slate-700 hover:bg-slate-800' : 'bg-white border-slate-200 hover:bg-slate-50'">
          <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
        </button>
      </div>
    </div>

    <div v-if="loading" class="space-y-4">
      <div v-for="i in 3" :key="i" class="h-32 rounded-xl animate-pulse"
        :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-200'" />
    </div>

    <template v-else-if="data">

      <!-- Income Statement -->
      <div class="rounded-xl border overflow-hidden"
        :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <div class="px-5 py-4 border-b flex items-center gap-2"
          :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
          <TrendingUp class="w-4 h-4 text-indigo-500" />
          <h2 class="font-semibold">Compte de Résultat {{ selectedYear }}</h2>
        </div>
        <div class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
          <!-- CA HT -->
          <div class="flex items-center justify-between px-5 py-3">
            <span class="text-sm" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Chiffre d'affaires HT</span>
            <span class="font-semibold text-emerald-500">{{ fmt(data.revenue) }}</span>
          </div>
          <!-- TVA -->
          <div class="flex items-center justify-between px-5 py-3">
            <span class="text-sm pl-4" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">dont TVA collectée</span>
            <span class="text-sm text-slate-400">{{ fmt(data.tva_collected) }}</span>
          </div>
          <!-- COGS -->
          <div class="flex items-center justify-between px-5 py-3">
            <span class="text-sm" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Coût des achats (COGS)</span>
            <span class="font-medium text-red-400">- {{ fmt(data.cogs) }}</span>
          </div>
          <!-- Gross Profit -->
          <div class="flex items-center justify-between px-5 py-3" :class="app.darkMode ? 'bg-slate-800/50' : 'bg-slate-50'">
            <span class="text-sm font-semibold">Marge Brute</span>
            <span class="font-bold" :class="data.gross_profit >= 0 ? 'text-emerald-500' : 'text-red-500'">
              {{ fmt(data.gross_profit) }}
              <span class="ml-2 text-xs font-normal text-slate-400">({{ fmtPct(data.gross_margin_pct) }})</span>
            </span>
          </div>
          <!-- OpEx -->
          <div class="flex items-center justify-between px-5 py-3">
            <span class="text-sm" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Charges d'exploitation (masse salariale)</span>
            <span class="font-medium text-red-400">- {{ fmt(data.opex) }}</span>
          </div>
          <!-- EBITDA -->
          <div class="flex items-center justify-between px-5 py-3" :class="app.darkMode ? 'bg-slate-800/50' : 'bg-slate-50'">
            <span class="text-sm font-semibold">EBITDA</span>
            <span class="font-bold" :class="data.ebitda >= 0 ? 'text-emerald-500' : 'text-red-500'">
              {{ fmt(data.ebitda) }}
            </span>
          </div>
          <!-- Net -->
          <div class="flex items-center justify-between px-5 py-3 bg-indigo-500/5">
            <span class="text-sm font-bold">Résultat Net</span>
            <span class="text-lg font-bold" :class="data.net_profit >= 0 ? 'text-emerald-500' : 'text-red-500'">
              {{ fmt(data.net_profit) }}
              <span class="ml-2 text-xs font-normal text-slate-400">({{ fmtPct(data.net_margin_pct) }})</span>
            </span>
          </div>
        </div>
      </div>

      <!-- Balance Sheet + Ratios -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

        <!-- Balance Sheet -->
        <div class="rounded-xl border overflow-hidden"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="px-5 py-4 border-b flex items-center gap-2"
            :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
            <Scale class="w-4 h-4 text-indigo-500" />
            <h2 class="font-semibold">Bilan Simplifié</h2>
          </div>
          <div class="p-5 space-y-4">
            <div>
              <p class="text-xs font-semibold uppercase tracking-wide mb-2 text-emerald-500">Actif</p>
              <div class="flex justify-between text-sm py-1.5">
                <span :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Total Actif</span>
                <span class="font-semibold">{{ fmt(data.total_assets) }}</span>
              </div>
            </div>
            <div class="border-t" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'" />
            <div>
              <p class="text-xs font-semibold uppercase tracking-wide mb-2 text-red-400">Passif</p>
              <div class="flex justify-between text-sm py-1.5">
                <span :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Total Dettes</span>
                <span class="font-semibold text-red-400">{{ fmt(data.total_liabilities) }}</span>
              </div>
            </div>
            <div class="border-t" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'" />
            <div class="flex justify-between text-sm py-1.5 font-bold">
              <span>Capitaux Propres</span>
              <span :class="data.equity >= 0 ? 'text-emerald-500' : 'text-red-500'">{{ fmt(data.equity) }}</span>
            </div>
          </div>
        </div>

        <!-- Financial Ratios -->
        <div class="rounded-xl border overflow-hidden"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="px-5 py-4 border-b flex items-center gap-2"
            :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
            <PiggyBank class="w-4 h-4 text-indigo-500" />
            <h2 class="font-semibold">Ratios Financiers</h2>
          </div>
          <div class="p-5 space-y-4">
            <div v-for="ratio in [
              { label: 'Ratio de liquidité', value: fmtX(data.current_ratio), desc: 'Actif / Passif', good: data.current_ratio >= 1.5 },
              { label: 'Ratio d\'endettement', value: fmtPct(data.debt_ratio), desc: 'Dettes / Actif total', good: data.debt_ratio <= 50 },
              { label: 'ROE — Rendement capitaux propres', value: fmtPct(data.roe), desc: 'Résultat net / Capitaux propres', good: data.roe >= 10 },
              { label: 'ROA — Rendement des actifs', value: fmtPct(data.roa), desc: 'Résultat net / Total actif', good: data.roa >= 5 },
            ]" :key="ratio.label">
              <div class="flex items-start justify-between">
                <div>
                  <p class="text-sm font-medium" :class="app.darkMode ? 'text-slate-200' : 'text-slate-800'">{{ ratio.label }}</p>
                  <p class="text-xs mt-0.5" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ ratio.desc }}</p>
                </div>
                <span class="text-lg font-bold" :class="ratio.good ? 'text-emerald-500' : 'text-amber-500'">{{ ratio.value }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Quarterly breakdown -->
      <div class="rounded-xl border overflow-hidden"
        :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <div class="px-5 py-4 border-b flex items-center gap-2"
          :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
          <BarChart3 class="w-4 h-4 text-indigo-500" />
          <h2 class="font-semibold">Décomposition Trimestrielle</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b" :class="app.darkMode ? 'border-slate-800 bg-slate-800/40' : 'border-slate-100 bg-slate-50'">
                <th class="text-left px-5 py-3 font-semibold text-xs uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Trimestre</th>
                <th class="text-right px-4 py-3 font-semibold text-xs uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">CA HT</th>
                <th class="text-right px-4 py-3 font-semibold text-xs uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Charges</th>
                <th class="text-right px-4 py-3 font-semibold text-xs uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Marge Brute</th>
                <th class="text-right px-5 py-3 font-semibold text-xs uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Marge %</th>
              </tr>
            </thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
              <tr v-for="q in data.quarters" :key="q.quarter" class="hover:bg-indigo-500/5 transition-colors">
                <td class="px-5 py-3 font-medium">T{{ q.quarter }} {{ selectedYear }}</td>
                <td class="px-4 py-3 text-right font-semibold text-emerald-500">{{ fmt(q.revenue) }}</td>
                <td class="px-4 py-3 text-right text-red-400">{{ fmt(q.cost) }}</td>
                <td class="px-4 py-3 text-right font-semibold" :class="q.gross_profit >= 0 ? 'text-emerald-500' : 'text-red-500'">{{ fmt(q.gross_profit) }}</td>
                <td class="px-5 py-3 text-right">
                  <span class="px-2 py-0.5 rounded-full text-xs font-medium"
                    :class="q.margin_pct >= 20 ? 'bg-emerald-500/10 text-emerald-500' : q.margin_pct >= 0 ? 'bg-amber-500/10 text-amber-500' : 'bg-red-500/10 text-red-500'">
                    {{ fmtPct(q.margin_pct) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Cash Flow chart -->
      <div class="rounded-xl border p-5"
        :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <div class="flex items-center justify-between mb-4">
          <h2 class="font-semibold flex items-center gap-2">
            <TrendingDown class="w-4 h-4 text-sky-500" />
            Flux de Trésorerie Mensuel
          </h2>
          <div class="flex items-center gap-4 text-xs">
            <span class="flex items-center gap-1.5"><span class="w-3 h-3 rounded-sm bg-sky-500 inline-block" />Entrées</span>
            <span class="flex items-center gap-1.5"><span class="w-3 h-3 rounded-sm bg-orange-400 inline-block" />Sorties</span>
          </div>
        </div>
        <div class="flex items-end gap-1 h-48">
          <template v-for="(m, idx) in data.cash_flow" :key="idx">
            <div class="flex-1 flex flex-col items-center gap-0.5">
              <div class="w-full flex items-end gap-0.5 h-40">
                <div class="flex-1 rounded-t bg-sky-500/80 transition-all duration-500"
                  :style="{ height: cfBarH(m.inflow) }" :title="`Entrées: ${fmt(m.inflow)}`" />
                <div class="flex-1 rounded-t bg-orange-400/80 transition-all duration-500"
                  :style="{ height: cfBarH(m.outflow) }" :title="`Sorties: ${fmt(m.outflow)}`" />
              </div>
              <span class="text-[10px]" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ months[idx] }}</span>
            </div>
          </template>
        </div>
      </div>

    </template>

    <div v-else-if="!loading" class="text-center py-20">
      <FileBarChart class="w-12 h-12 mx-auto text-slate-300 mb-3" />
      <p class="text-slate-400">Aucune donnée financière disponible</p>
    </div>
  </div>
</template>
