<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { ShoppingBag, RefreshCw, Calendar, Users, Download, BarChart3, AlertCircle } from 'lucide-vue-next'
import { reportsAPI } from '@/api/client'

const app = useAppStore()
const loading = ref(false)
const selectedYear = ref(new Date().getFullYear())
const years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - i)
const data = ref<any>(null)
const months = ['Jan','Fév','Mar','Avr','Mai','Jun','Jul','Aoû','Sep','Oct','Nov','Déc']

async function load() {
  loading.value = true
  try { const r = await reportsAPI.getPurchaseReports(String(selectedYear.value)); data.value = r.data }
  catch { app.addToast('Erreur chargement rapports achats', 'error') }
  finally { loading.value = false }
}
onMounted(load)

const fmt = (n: number) => new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'
const fmtN = (n: number) => new Intl.NumberFormat('fr-DZ').format(n)
const maxMonthly = computed(() => data.value?.monthly?.length ? Math.max(...data.value.monthly.map((m: any) => m.spend), 1) : 1)
const maxSup = computed(() => data.value?.top_suppliers?.length ? Math.max(...data.value.top_suppliers.map((s: any) => s.spend), 1) : 1)

function exportCSV() {
  if (!data.value) return
  const rows = [['Mois','Achats','Nb'],
    ...data.value.monthly.map((m: any) => [months[m.month - 1], m.spend, m.count])]
  const a = document.createElement('a')
  a.href = URL.createObjectURL(new Blob(['\ufeff' + rows.map((r: any[]) => r.join(',')).join('\n')], { type: 'text/csv' }))
  a.download = `purchase_report_${selectedYear.value}.csv`; a.click()
}
</script>

<template>
  <div class="min-h-screen p-6 space-y-6" :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <ShoppingBag class="w-6 h-6 text-orange-500" /> Rapports Achats
        </h1>
        <p class="text-sm mt-0.5" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Dépenses, fournisseurs, articles achetés — {{ selectedYear }}</p>
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
          { label: 'Total Achats', value: fmt(data.total_spend), color: 'text-orange-500', bg: 'bg-orange-500/10', icon: ShoppingBag },
          { label: 'Nb Commandes', value: fmtN(data.order_count), color: 'text-indigo-500', bg: 'bg-indigo-500/10', icon: BarChart3 },
          { label: 'Fournisseurs Actifs', value: fmtN(data.supplier_count), color: 'text-teal-500', bg: 'bg-teal-500/10', icon: Users },
          { label: 'Dettes Frs.', value: fmt(data.total_ap), color: 'text-red-500', bg: 'bg-red-500/10', icon: AlertCircle },
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

      <!-- Monthly spend chart -->
      <div class="rounded-xl border p-5" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <h2 class="font-semibold mb-4">Dépenses Mensuelles {{ selectedYear }}</h2>
        <div class="flex items-end gap-1.5 h-52">
          <div v-for="(m, idx) in data.monthly" :key="idx" class="flex-1 flex flex-col items-center gap-1">
            <span class="text-[10px]" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ fmtN(m.spend / 1000) }}k</span>
            <div class="w-full rounded-t transition-all duration-500 bg-orange-500 min-h-1"
              :style="{ height: Math.max((m.spend / maxMonthly) * 180, 2) + 'px' }"
              :title="`${months[idx]}: ${fmt(m.spend)}`" />
            <span class="text-[10px]" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">{{ months[idx] }}</span>
          </div>
        </div>
      </div>

      <!-- Top Suppliers + Top Items -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- Top Suppliers -->
        <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="px-5 py-4 border-b" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
            <h2 class="font-semibold flex items-center gap-2"><Users class="w-4 h-4 text-orange-500" /> Top Fournisseurs</h2>
          </div>
          <div class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
            <div v-for="(s, i) in data.top_suppliers" :key="i" class="px-5 py-3 flex items-center gap-3">
              <span class="text-xs font-bold w-5 h-5 rounded bg-orange-500 text-white flex items-center justify-center flex-shrink-0">{{ i+1 }}</span>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium truncate">{{ s.name }}</p>
                <div class="mt-1 h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                  <div class="h-full rounded-full bg-orange-500" :style="{ width: ((s.spend / maxSup) * 100).toFixed(1) + '%' }" />
                </div>
              </div>
              <div class="text-right flex-shrink-0">
                <p class="text-sm font-semibold">{{ fmtN(s.spend) }}</p>
                <p class="text-xs text-slate-400">{{ s.invoice_count }} fac.</p>
              </div>
            </div>
            <p v-if="!data.top_suppliers?.length" class="px-5 py-6 text-sm text-center text-slate-400">Aucun fournisseur</p>
          </div>
        </div>

        <!-- Top Items -->
        <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="px-5 py-4 border-b" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
            <h2 class="font-semibold">Top Articles Achetés</h2>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b" :class="app.darkMode ? 'border-slate-800 bg-slate-800/40' : 'border-slate-100 bg-slate-50'">
                  <th class="text-left px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">#</th>
                  <th class="text-left px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Article</th>
                  <th class="text-right px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Qté</th>
                  <th class="text-right px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Montant</th>
                </tr>
              </thead>
              <tbody class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
                <tr v-for="(item, i) in data.top_items" :key="i" class="hover:bg-orange-500/5 transition-colors">
                  <td class="px-5 py-2.5 text-slate-400">{{ i + 1 }}</td>
                  <td class="px-4 py-2.5 font-medium truncate max-w-[180px]">{{ item.name }}</td>
                  <td class="px-4 py-2.5 text-right">{{ fmtN(item.quantity) }}</td>
                  <td class="px-5 py-2.5 text-right font-semibold text-orange-500">{{ fmtN(item.spend) }}</td>
                </tr>
                <tr v-if="!data.top_items?.length"><td colspan="4" class="px-5 py-6 text-center text-slate-400">Aucun article</td></tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Order Status -->
      <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <div class="px-5 py-4 border-b" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
          <h2 class="font-semibold">Commandes par Statut</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b" :class="app.darkMode ? 'border-slate-800 bg-slate-800/40' : 'border-slate-100 bg-slate-50'">
                <th class="text-left px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Statut</th>
                <th class="text-right px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Nb</th>
                <th class="text-right px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Montant Total</th>
              </tr>
            </thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
              <tr v-for="s in data.by_status" :key="s.status" class="hover:bg-orange-500/5 transition-colors">
                <td class="px-5 py-2.5">
                  <span class="px-2 py-0.5 rounded-full text-xs font-medium capitalize bg-slate-500/10 text-slate-400">{{ s.status }}</span>
                </td>
                <td class="px-4 py-2.5 text-right">{{ s.count }}</td>
                <td class="px-5 py-2.5 text-right font-semibold">{{ fmt(s.amount) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <div v-else-if="!loading" class="text-center py-20">
      <ShoppingBag class="w-12 h-12 mx-auto text-slate-300 mb-3" />
      <p class="text-slate-400">Aucune donnée d'achat</p>
    </div>
  </div>
</template>
