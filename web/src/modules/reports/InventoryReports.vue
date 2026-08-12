<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { Package, RefreshCw, AlertTriangle, Warehouse, BarChart3, Download } from 'lucide-vue-next'
import { reportsAPI } from '@/api/client'

const app = useAppStore()
const loading = ref(false)
const data = ref<any>(null)

async function load() {
  loading.value = true
  try { const r = await reportsAPI.getInventoryReports(); data.value = r.data }
  catch { app.addToast('Erreur chargement rapports inventaire', 'error') }
  finally { loading.value = false }
}
onMounted(load)

const fmt = (n: number) => new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'
const fmtN = (n: number) => new Intl.NumberFormat('fr-DZ').format(n)

function exportCSV() {
  if (!data.value?.low_stock_items) return
  const rows = [['Code','Nom','Stock','Seuil Réappro'],
    ...data.value.low_stock_items.map((i: any) => [i.code, i.name, i.quantity, i.reorder_point])]
  const a = document.createElement('a')
  a.href = URL.createObjectURL(new Blob(['\ufeff' + rows.map((r: any[]) => r.join(',')).join('\n')], { type: 'text/csv' }))
  a.download = 'inventory_low_stock.csv'; a.click()
}

const movtColors: Record<string, string> = {
  sale: 'text-emerald-500',
  purchase: 'text-indigo-500',
  transfer: 'text-sky-500',
  adjustment: 'text-amber-500',
  return: 'text-violet-500',
}
</script>

<template>
  <div class="min-h-screen p-6 space-y-6" :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <Package class="w-6 h-6 text-violet-500" /> Rapports Inventaire
        </h1>
        <p class="text-sm mt-0.5" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
          Valorisation des stocks, alertes réapprovisionnement, mouvements
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="exportCSV" class="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-sm font-medium transition-colors" :class="app.darkMode ? 'bg-slate-900 border-slate-700 hover:bg-slate-800' : 'bg-white border-slate-200 hover:bg-slate-50'">
          <Download class="w-4 h-4" /> Stock Faible
        </button>
        <button @click="load" class="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-sm font-medium transition-colors" :class="app.darkMode ? 'bg-slate-900 border-slate-700 hover:bg-slate-800' : 'bg-white border-slate-200 hover:bg-slate-50'">
          <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
        </button>
      </div>
    </div>

    <div v-if="loading" class="grid grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="i in 6" :key="i" class="h-24 rounded-xl animate-pulse" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-200'" />
    </div>

    <template v-else-if="data">
      <!-- KPIs -->
      <div class="grid grid-cols-2 lg:grid-cols-3 gap-4">
        <div v-for="kpi in [
          { label: 'Articles Actifs', value: fmtN(data.total_items), icon: Package, color: 'text-violet-500', bg: 'bg-violet-500/10' },
          { label: 'Valeur Stock', value: fmt(data.total_value), icon: BarChart3, color: 'text-emerald-500', bg: 'bg-emerald-500/10' },
          { label: 'Stock Faible', value: fmtN(data.low_stock_count), icon: AlertTriangle, color: 'text-amber-500', bg: 'bg-amber-500/10' },
          { label: 'En Rupture', value: fmtN(data.out_of_stock), icon: AlertTriangle, color: 'text-red-500', bg: 'bg-red-500/10' },
          { label: 'Entrepôts', value: fmtN(data.warehouse_count), icon: Warehouse, color: 'text-sky-500', bg: 'bg-sky-500/10' },
          { label: 'Mouvements 30j', value: fmtN(data.movements_30d), icon: BarChart3, color: 'text-indigo-500', bg: 'bg-indigo-500/10' },
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

      <!-- By Category + By Warehouse -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <!-- By Category -->
        <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="px-5 py-4 border-b" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
            <h2 class="font-semibold">Stock par Catégorie</h2>
          </div>
          <div class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
            <div v-for="cat in data.by_category" :key="cat.category" class="px-5 py-3">
              <div class="flex items-center justify-between mb-1.5">
                <span class="text-sm font-medium">{{ cat.category }}</span>
                <div class="text-right">
                  <p class="text-sm font-semibold">{{ fmt(cat.value) }}</p>
                  <p class="text-xs text-slate-400">{{ cat.item_count }} articles, {{ fmtN(cat.quantity) }} unités</p>
                </div>
              </div>
              <div class="h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                <div class="h-full rounded-full bg-violet-500 transition-all duration-500"
                  :style="{ width: ((cat.value / (data.total_value || 1)) * 100).toFixed(1) + '%' }" />
              </div>
            </div>
            <p v-if="!data.by_category?.length" class="px-5 py-6 text-center text-sm text-slate-400">Aucune catégorie</p>
          </div>
        </div>

        <!-- By Warehouse -->
        <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="px-5 py-4 border-b" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
            <h2 class="font-semibold flex items-center gap-2"><Warehouse class="w-4 h-4 text-sky-500" /> Stock par Entrepôt</h2>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b" :class="app.darkMode ? 'border-slate-800 bg-slate-800/40' : 'border-slate-100 bg-slate-50'">
                  <th class="text-left px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Entrepôt</th>
                  <th class="text-right px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Articles</th>
                  <th class="text-right px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Unités</th>
                  <th class="text-right px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Valeur</th>
                </tr>
              </thead>
              <tbody class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
                <tr v-for="wh in data.by_warehouse" :key="wh.warehouse" class="hover:bg-sky-500/5 transition-colors">
                  <td class="px-5 py-2.5 font-medium">{{ wh.warehouse }}</td>
                  <td class="px-4 py-2.5 text-right">{{ wh.item_count }}</td>
                  <td class="px-4 py-2.5 text-right">{{ fmtN(wh.quantity) }}</td>
                  <td class="px-5 py-2.5 text-right font-semibold text-emerald-500">{{ fmt(wh.value) }}</td>
                </tr>
                <tr v-if="!data.by_warehouse?.length"><td colspan="4" class="px-5 py-6 text-center text-slate-400">Aucun entrepôt</td></tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Low Stock Alert -->
      <div v-if="data.low_stock_items?.length" class="rounded-xl border overflow-hidden border-amber-500/30"
        :class="app.darkMode ? 'bg-slate-900' : 'bg-white'">
        <div class="px-5 py-4 border-b border-amber-500/20 flex items-center gap-2 bg-amber-500/5">
          <AlertTriangle class="w-4 h-4 text-amber-500" />
          <h2 class="font-semibold text-amber-500">Articles en Stock Faible ({{ data.low_stock_items.length }})</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b" :class="app.darkMode ? 'border-slate-800 bg-slate-800/40' : 'border-slate-100 bg-slate-50'">
                <th class="text-left px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Code</th>
                <th class="text-left px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Désignation</th>
                <th class="text-right px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Stock Actuel</th>
                <th class="text-right px-5 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-400">Seuil Réappro</th>
              </tr>
            </thead>
            <tbody class="divide-y" :class="app.darkMode ? 'divide-slate-800' : 'divide-slate-100'">
              <tr v-for="item in data.low_stock_items" :key="item.code" class="hover:bg-amber-500/5 transition-colors">
                <td class="px-5 py-2.5 font-mono text-xs text-amber-500">{{ item.code }}</td>
                <td class="px-4 py-2.5 font-medium">{{ item.name }}</td>
                <td class="px-4 py-2.5 text-right font-bold" :class="item.quantity === 0 ? 'text-red-500' : 'text-amber-500'">{{ fmtN(item.quantity) }}</td>
                <td class="px-5 py-2.5 text-right text-slate-400">{{ fmtN(item.reorder_point) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Movement type breakdown -->
      <div class="rounded-xl border p-5" :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
        <h2 class="font-semibold mb-4">Mouvements par Type (30 derniers jours)</h2>
        <div class="grid grid-cols-2 lg:grid-cols-5 gap-3">
          <div v-for="mvt in data.by_movement_type" :key="mvt.type"
            class="rounded-lg p-3 text-center border" :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-slate-50 border-slate-200'">
            <p :class="[movtColors[mvt.type] || 'text-slate-400', 'text-lg font-bold']">{{ fmtN(mvt.count) }}</p>
            <p class="text-xs capitalize" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ mvt.type }}</p>
            <p class="text-xs font-medium mt-0.5" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">{{ fmtN(mvt.quantity) }} u.</p>
          </div>
        </div>
      </div>
    </template>

    <div v-else-if="!loading" class="text-center py-20">
      <Package class="w-12 h-12 mx-auto text-slate-300 mb-3" />
      <p class="text-slate-400">Aucune donnée d'inventaire</p>
    </div>
  </div>
</template>
