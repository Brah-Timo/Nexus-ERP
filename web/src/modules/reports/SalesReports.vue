<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { TrendingUp, RefreshCw, Calendar, Users, ShoppingCart, Clock, Download, BarChart3, AlertCircle } from 'lucide-vue-next'
import { reportsAPI } from '@/api/client'

const app = useAppStore()
const loading = ref(false)
const selectedYear = ref(new Date().getFullYear())
const years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - i)

const data = ref<any>(null)
const months = ['Jan','Fév','Mar','Avr','Mai','Jun','Jul','Aoû','Sep','Oct','Nov','Déc']

async function load() {
  loading.value = true
  try {
    const r = await reportsAPI.getSalesReports(String(selectedYear.value))
    data.value = r.data
  } catch { app.addToast('Erreur chargement rapports ventes', 'error') }
  finally { loading.value = false }
}
onMounted(load)

const fmt = (n: number) =>
  new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(n) + ' DZD'
const fmtN = (n: number) => new Intl.NumberFormat('fr-DZ').format(n)
const fmtPct = (n: number) => (n || 0).toFixed(1) + '%'

const maxMonthlyRev = computed(() => {
  if (!data.value?.monthly?.length) return 1
  return Math.max(...data.value.monthly.map((m: any) => m.revenue), 1)
})
const maxCust = computed(() => {
  if (!data.value?.top_customers?.length) return 1
  return Math.max(...data.value.top_customers.map((c: any) => c.revenue), 1)
})

function exportCSV() {
  if (!data.value) return
  const rows = [['Mois', 'CA', 'Nb Factures'], ...data.value.monthly.map((m: any) => [months[m.month - 1], m.revenue, m.count])]
  const a = document.createElement('a')
  a.href = URL.createObjectURL(new Blob(['\ufeff' + rows.map((r: any[]) => r.join(',')).join('\n')], { type: 'text/csv' }))
  a.download = `sales_report_${selectedYear.value}.csv`
  a.click()
}

const statusColors: Record<string, string> = {
  paid: 'bg-emerald-500/10 text-emerald-500',
  confirmed: 'bg-indigo-500/10 text-indigo-500',
  partial: 'bg-amber-500/10 text-amber-500',
  cancelled: 'bg-red-500/10 text-red-500',
  draft: 'bg-slate-500/10 text-slate-400',
}
</script>

<template>
  <div class="min-h-screen p-6 space-y-6" :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">

    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <TrendingUp class="w-6 h-6 text-emerald-500" /> Rapports Ventes
        </h1>
        <p class="text-sm mt-0.5" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
          Analyse des ventes, clients, produits — {{ selectedYear }}
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
        <button @click="exportCSV" class="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-sm font-medium transition-colors"
          :class="app.darkMode ? 'bg-slate-900 border-slate-700 hover:bg-slate-800' : 'bg-white border-slate-200 hover:bg-slate-50'">
          <Download class="w-4 h-4" /> CSV
        </button>
        <button @click="load" class="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-sm font-medium transition-colors"
          :class="app.darkMode ? 'bg-slate-900 border-slate-700 hover:bg-slate-800' : 'bg-white border-slate-200 hover:bg-slate-50'">
          <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
        </button>
      </div>
    </div>

    <div v-if="loading" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="i in 8" :key="i" class="h-24 rounded-xl animate-pulse" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-200'" />
    </div>

    <template v-else-if="data">

      <!-- KPI Cards -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div v-for="kpi in [
          { label: 'CA Total', value: fmt(data.total_revenue), icon: TrendingUp, color: 'text-emerald-500', bg: 'bg-emerald-500/10' },
          { label: 'Commandes', value: fmtN(data.total_orders), icon: ShoppingCart, color: 'text-indigo-500', bg: 'bg-indigo-500/10' },
          { label: 'Créances', value: fmt(data.total_ar), icon: Clock, color: 'text-amber-500', bg: 'bg-amber-500/10' },
          { label: 'Panier Moyen', value: fmt(data.avg_order_value), icon: BarChart3, color: 'text-sky-500', bg: 'bg-sky-500/10' },
          { label: 'Nb Factures', value: fmtN(data.invoice_count), icon: Users, color: 'text-violet-500', bg: 'bg-violet-500/10' },
          { label: 'Taux Recouvrement', value: fmtPct(data.collection_rate), icon: TrendingUp, color: 'text-emerald-500', bg: 'bg-emerald-500/10' },
          { label: 'Créances Échues', value: fmt(data.overdue_ar), icon: AlertCircle, color: 'text-red-500', bg: 'bg-red-500/10' },
          { label: 'Pipeline CRM', value: fmt(data.pipeline_total), icon: BarChart3, color: 'text-teal-500', bg: 'bg-teal-500/10' },
        ]" :key="kpi.label"
          class="rounded-xl border p-4 flex items-center gap-3"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div :class="[kpi.bg, 'w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0']">
            <component :is="kpi.icon" :class="[kpi.color, 'w-5 h-5']" />
          </div>
          <div>
            <p class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ kpi.label }}</p>
            <p class="text-base font-bold mt-0.5">{{ kpi.value }}</p>
          </div>
        </div>
      </div>

      <!-- Monthly Revenue Chart -->
      <div class="rounded-xl border p-5"
        :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <h2 class="font-semibold mb-4">Chiffre d'Affaires Mensuel {{ selectedYear }}</h2>
        <div class="flex items-end gap-1.5 h-52">
          <div v-for="(m, idx) in data.monthly" :key="idx" class="flex-1 flex flex-col items-center gap-1">
            <span class="text-[10px]" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
              {{ fmtN(m.revenue / 1000) }}k
            </span>
            <div class="w-full rounded-t transition-all duration-500 bg-emerald-500 min-h-1"
              :style="{ height: Math.max((m.revenue / maxMonthlyRev) * 180, 2) + 'px' }"
              :title="`${months[idx]}: ${fmt(m.revenue)} (${m.count} factures)`" />
            <span class="text-[10px]" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ months[idx] }}</span>
          </div>
        </div>
      </div>

      <!-- Top Customers + By Status -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

        <!-- Top Customers -->
        <div class="rounded-xl border overflow-hidden"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="px-5 py-4 border-b flex items-center gap-2"
            :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
            <Users class="w-4 h-4 text-emerald-500" />
            <h2 class="font-semibold">Top Clients</h2>
          </div>
          <div class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
            <div v-for="(c, i) in data.top_customers" :key="i" class="px-5 py-3 flex items-center gap-3">
              <span class="text-xs font-bold w-5 h-5 rounded bg-emerald-500 text-white flex items-center justify-center flex-shrink-0">{{ i+1 }}</span>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium truncate">{{ c.name }}</p>
                <div class="mt-1 h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                  <div class="h-full rounded-full bg-emerald-500" :style="{ width: ((c.revenue / maxCust) * 100).toFixed(1) + '%' }" />
                </div>
              </div>
              <div class="text-right flex-shrink-0">
                <p class="text-sm font-semibold">{{ fmtN(c.revenue) }}</p>
                <p class="text-xs text-slate-400">{{ c.invoice_count }} factures</p>
              </div>
            </div>
            <p v-if="!data.top_customers?.length" class="px-5 py-6 text-sm text-center text-slate-400">Aucun client</p>
          </div>
        </div>

        <!-- Status breakdown -->
        <div class="rounded-xl border overflow-hidden"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="px-5 py-4 border-b flex items-center gap-2"
            :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
            <BarChart3 class="w-4 h-4 text-emerald-500" />
            <h2 class="font-semibold">Factures par Statut</h2>
          </div>
          <div class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
            <div v-for="s in data.by_status" :key="s.status" class="px-5 py-3 flex items-center justify-between gap-3">
              <span :class="[statusColors[s.status] || 'bg-slate-500/10 text-slate-400', 'px-2 py-0.5 rounded-full text-xs font-medium capitalize']">
                {{ s.status }}
              </span>
              <div class="flex-1 mx-3">
                <div class="h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                  <div class="h-full rounded-full bg-emerald-500" :style="{ width: ((s.amount / (data.total_revenue || 1)) * 100).toFixed(1) + '%' }" />
                </div>
              </div>
              <div class="text-right flex-shrink-0">
                <p class="text-sm font-semibold">{{ fmtN(s.amount) }}</p>
                <p class="text-xs text-slate-400">{{ s.count }} doc.</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Top Products table -->
      <div class="rounded-xl border overflow-hidden"
        :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <div class="px-5 py-4 border-b"
          :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
          <h2 class="font-semibold">Top Produits / Services</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b" :class="app.darkMode ? 'border-slate-800 bg-slate-800/40' : 'border-slate-100 bg-slate-50'">
                <th class="text-left px-5 py-3 text-xs font-semibold uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">#</th>
                <th class="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Produit / Service</th>
                <th class="text-right px-4 py-3 text-xs font-semibold uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Quantité</th>
                <th class="text-right px-5 py-3 text-xs font-semibold uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">CA</th>
              </tr>
            </thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
              <tr v-for="(p, i) in data.top_products" :key="i" class="hover:bg-indigo-500/5 transition-colors">
                <td class="px-5 py-3 text-slate-400">{{ i + 1 }}</td>
                <td class="px-4 py-3 font-medium">{{ p.name }}</td>
                <td class="px-4 py-3 text-right">{{ fmtN(p.quantity) }}</td>
                <td class="px-5 py-3 text-right font-semibold text-emerald-500">{{ fmt(p.revenue) }}</td>
              </tr>
              <tr v-if="!data.top_products?.length">
                <td colspan="4" class="px-5 py-8 text-center text-slate-400">Aucun produit</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <div v-else-if="!loading" class="text-center py-20">
      <TrendingUp class="w-12 h-12 mx-auto text-slate-300 mb-3" />
      <p class="text-slate-400">Aucune donnée de vente</p>
    </div>
  </div>
</template>
