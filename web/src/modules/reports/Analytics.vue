<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { LineChart, RefreshCw, Calendar, TrendingUp, Target, Users, BarChart3 } from 'lucide-vue-next'
import { reportsAPI } from '@/api/client'

const app = useAppStore()
const loading = ref(false)
const selectedYear = ref(new Date().getFullYear())
const years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - i)
const data = ref<any>(null)

const months = ['Jan','Fév','Mar','Avr','Mai','Jun','Jul','Aoû','Sep','Oct','Nov','Déc']

async function load() {
  loading.value = true
  try { const r = await reportsAPI.getAnalytics(String(selectedYear.value)); data.value = r.data }
  catch { app.addToast('Erreur chargement analytics', 'error') }
  finally { loading.value = false }
}
onMounted(load)

const fmt = (n: number) => new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'
const fmtN = (n: number) => new Intl.NumberFormat('fr-DZ').format(n)
const fmtPct = (n: number) => (n || 0).toFixed(1) + '%'

const maxCurrent = computed(() => {
  if (!data.value?.monthly_comparison?.length) return 1
  return Math.max(...data.value.monthly_comparison.flatMap((m: any) => [m.current, m.previous]), 1)
})

const segmentColors = ['bg-indigo-500', 'bg-sky-500', 'bg-emerald-500', 'bg-amber-500', 'bg-violet-500']
const segmentTextColors = ['text-indigo-500', 'text-sky-500', 'text-emerald-500', 'text-amber-500', 'text-violet-500']

const totalSegmentRevenue = computed(() =>
  data.value?.customer_segments?.reduce((a: number, s: any) => a + s.revenue, 0) || 1
)
</script>

<template>
  <div class="min-h-screen p-6 space-y-6" :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <LineChart class="w-6 h-6 text-violet-500" /> Analytics Avancés
        </h1>
        <p class="text-sm mt-0.5" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
          Comparaison N/N-1, segmentation clients, performance produits, charge fiscale
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
      <div v-for="i in 4" :key="i" class="h-48 rounded-xl animate-pulse" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-200'" />
    </div>

    <template v-else-if="data">

      <!-- YoY Revenue (last 3 years) -->
      <div class="rounded-xl border p-5" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <h2 class="font-semibold mb-4 flex items-center gap-2">
          <BarChart3 class="w-4 h-4 text-violet-500" /> Évolution CA sur 3 ans
        </h2>
        <div class="flex items-end gap-8 h-32">
          <div v-for="y in data.yoy_revenue" :key="y.year" class="flex-1 flex flex-col items-center gap-2">
            <span class="text-sm font-semibold" :class="app.darkMode ? 'text-slate-200' : 'text-slate-700'">
              {{ new Intl.NumberFormat('fr-DZ', { notation: 'compact', maximumSignificantDigits: 3 }).format(y.revenue) }}
            </span>
            <div class="w-full rounded-t transition-all duration-700 bg-violet-500"
              :style="{ height: Math.max((y.revenue / Math.max(...(data.yoy_revenue || []).map((r: any) => r.revenue), 1)) * 90, 4) + 'px' }" />
            <span class="text-sm font-bold" :class="app.year === selectedYear ? 'text-violet-500' : (app.darkMode ? 'text-slate-400' : 'text-slate-500')">{{ y.year }}</span>
          </div>
          <p v-if="!data.yoy_revenue?.length" class="text-slate-400 text-sm">Aucune donnée</p>
        </div>
      </div>

      <!-- Monthly Comparison N vs N-1 -->
      <div class="rounded-xl border p-5" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <div class="flex items-center justify-between mb-4">
          <h2 class="font-semibold">Comparaison Mensuelle {{ selectedYear }} vs {{ selectedYear - 1 }}</h2>
          <div class="flex items-center gap-4 text-xs">
            <span class="flex items-center gap-1.5"><span class="w-3 h-3 rounded-sm bg-violet-500 inline-block"/>{{ selectedYear }}</span>
            <span class="flex items-center gap-1.5"><span class="w-3 h-3 rounded-sm bg-slate-400 inline-block"/>{{ selectedYear - 1 }}</span>
          </div>
        </div>
        <div class="flex items-end gap-1 h-48">
          <div v-for="(m, idx) in data.monthly_comparison" :key="idx" class="flex-1 flex flex-col items-center gap-0.5">
            <div class="w-full flex items-end gap-0.5 h-40">
              <div class="flex-1 rounded-t bg-violet-500 transition-all duration-500"
                :style="{ height: Math.max((m.current / maxCurrent) * 150, 2) + 'px' }"
                :title="`${months[idx]} ${selectedYear}: ${fmt(m.current)}`" />
              <div class="flex-1 rounded-t bg-slate-400/60 transition-all duration-500"
                :style="{ height: Math.max((m.previous / maxCurrent) * 150, 2) + 'px' }"
                :title="`${months[idx]} ${selectedYear - 1}: ${fmt(m.previous)}`" />
            </div>
            <span class="text-[10px]" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ months[idx] }}</span>
          </div>
        </div>
        <!-- Growth indicators -->
        <div class="mt-3 flex gap-1 overflow-x-auto pb-1">
          <div v-for="(m, idx) in data.monthly_comparison" :key="idx" class="flex-1 text-center min-w-[40px]">
            <span class="text-[10px] font-medium" :class="m.growth_pct >= 0 ? 'text-emerald-500' : 'text-red-400'">
              {{ m.growth_pct >= 0 ? '+' : '' }}{{ (m.growth_pct || 0).toFixed(0) }}%
            </span>
          </div>
        </div>
      </div>

      <!-- Customer Segmentation + Product Performance -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

        <!-- Customer Segments -->
        <div class="rounded-xl border p-5" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <h2 class="font-semibold mb-4 flex items-center gap-2">
            <Users class="w-4 h-4 text-violet-500" /> Segmentation Clients
          </h2>
          <div class="space-y-4">
            <div v-for="(seg, i) in data.customer_segments" :key="seg.segment">
              <div class="flex items-center justify-between mb-1.5">
                <div class="flex items-center gap-2">
                  <span :class="[segmentColors[i % segmentColors.length], 'w-3 h-3 rounded-sm inline-block']" />
                  <span class="text-sm font-medium">{{ seg.segment }}</span>
                </div>
                <div class="text-right">
                  <p :class="[segmentTextColors[i % segmentTextColors.length], 'text-sm font-semibold']">{{ fmtPct(seg.revenue / totalSegmentRevenue * 100) }}</p>
                  <p class="text-xs text-slate-400">{{ seg.customers }} clients</p>
                </div>
              </div>
              <div class="h-2 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                <div :class="[segmentColors[i % segmentColors.length], 'h-full rounded-full transition-all duration-500']"
                  :style="{ width: ((seg.revenue / totalSegmentRevenue) * 100).toFixed(1) + '%' }" />
              </div>
              <p class="text-xs text-right mt-0.5 text-slate-400">{{ fmt(seg.revenue) }}</p>
            </div>
            <p v-if="!data.customer_segments?.length" class="text-sm text-center text-slate-400 py-4">Aucune donnée</p>
          </div>
        </div>

        <!-- Product Performance -->
        <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="px-5 py-4 border-b" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
            <h2 class="font-semibold flex items-center gap-2">
              <Target class="w-4 h-4 text-violet-500" /> Performance Produits (Pareto)
            </h2>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b" :class="app.darkMode ? 'border-slate-800 bg-slate-800/40' : 'border-slate-100 bg-slate-50'">
                  <th class="text-left px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">#</th>
                  <th class="text-left px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Produit</th>
                  <th class="text-right px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">CA</th>
                  <th class="text-right px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Part</th>
                </tr>
              </thead>
              <tbody class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
                <tr v-for="(p, i) in data.product_performance" :key="i" class="hover:bg-violet-500/5 transition-colors">
                  <td class="px-5 py-2.5">
                    <span class="text-xs font-bold w-5 h-5 rounded bg-violet-500 text-white flex items-center justify-center">{{ i+1 }}</span>
                  </td>
                  <td class="px-4 py-2.5 font-medium truncate max-w-[150px]">{{ p.product }}</td>
                  <td class="px-4 py-2.5 text-right font-semibold text-emerald-500">{{ fmtN(p.revenue) }}</td>
                  <td class="px-5 py-2.5 text-right">
                    <span class="px-2 py-0.5 rounded-full text-xs font-medium bg-violet-500/10 text-violet-500">
                      {{ fmtPct(p.contribution_pct) }}
                    </span>
                  </td>
                </tr>
                <tr v-if="!data.product_performance?.length">
                  <td colspan="4" class="px-5 py-6 text-center text-slate-400">Aucun produit</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Tax Burden -->
      <div class="rounded-xl border p-5" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <h2 class="font-semibold mb-4 flex items-center gap-2">
          <TrendingUp class="w-4 h-4 text-violet-500" /> Charge Fiscale {{ selectedYear }}
        </h2>
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div class="text-center rounded-xl p-5" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-50'">
            <p class="text-xs uppercase tracking-wide mb-2" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Impôts dus</p>
            <p class="text-3xl font-bold text-red-500">{{ fmt(data.total_tax_due) }}</p>
          </div>
          <div class="text-center rounded-xl p-5" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-50'">
            <p class="text-xs uppercase tracking-wide mb-2" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Impôts payés</p>
            <p class="text-3xl font-bold text-emerald-500">{{ fmt(data.total_tax_paid) }}</p>
          </div>
          <div class="text-center rounded-xl p-5" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-50'">
            <p class="text-xs uppercase tracking-wide mb-2" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Taux fiscal effectif</p>
            <p class="text-3xl font-bold" :class="data.effective_tax_rate <= 25 ? 'text-emerald-500' : 'text-amber-500'">
              {{ fmtPct(data.effective_tax_rate) }}
            </p>
            <p class="text-xs mt-1 text-slate-400">% du CA HT</p>
          </div>
        </div>
        <div class="mt-4">
          <div class="flex justify-between text-sm mb-1.5">
            <span :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Taux de règlement fiscal</span>
            <span class="font-semibold" :class="data.total_tax_due > 0 ? (data.total_tax_paid / data.total_tax_due >= 0.9 ? 'text-emerald-500' : 'text-amber-500') : 'text-slate-400'">
              {{ data.total_tax_due > 0 ? fmtPct((data.total_tax_paid / data.total_tax_due) * 100) : '—' }}
            </span>
          </div>
          <div class="h-2 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
            <div class="h-full rounded-full transition-all duration-700 bg-violet-500"
              :style="{ width: data.total_tax_due > 0 ? Math.min((data.total_tax_paid / data.total_tax_due) * 100, 100).toFixed(1) + '%' : '0%' }" />
          </div>
        </div>
      </div>
    </template>

    <div v-else-if="!loading" class="text-center py-20">
      <LineChart class="w-12 h-12 mx-auto text-slate-300 mb-3" />
      <p class="text-slate-400">Aucune donnée d'analyse</p>
    </div>
  </div>
</template>
