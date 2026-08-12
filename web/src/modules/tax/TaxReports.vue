<template>
  <div class="min-h-screen bg-gray-950 text-gray-100">
    <!-- Header -->
    <div class="bg-gray-900 border-b border-gray-800 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-orange-500 to-red-600 flex items-center justify-center shadow-lg">
            <FileBarChart2 :size="20" class="text-white" />
          </div>
          <div>
            <h1 class="text-xl font-bold text-white">Rapports Fiscaux</h1>
            <p class="text-xs text-gray-400 mt-0.5">Synthèse annuelle — TVA, TAP, IBS, IRG, Timbre</p>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <select v-model="filterYear" @change="loadReport" class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-orange-500">
            <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}</option>
          </select>
          <button @click="exportReport" class="flex items-center gap-2 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-200 rounded-lg text-sm font-medium transition-colors">
            <Download :size="15" />
            Export CSV
          </button>
        </div>
      </div>
    </div>

    <div class="p-6 space-y-6">
      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-20">
        <div class="flex flex-col items-center gap-3">
          <div class="w-10 h-10 border-2 border-orange-500 border-t-transparent rounded-full animate-spin"></div>
          <span class="text-sm text-gray-400">Chargement rapports fiscaux...</span>
        </div>
      </div>

      <template v-else>
        <!-- Annual Summary Banner -->
        <div class="bg-gradient-to-r from-orange-950/60 via-red-950/40 to-gray-900 border border-orange-800/30 rounded-2xl p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-8 h-8 rounded-lg bg-orange-500/20 flex items-center justify-center">
              <TrendingUp :size="16" class="text-orange-400" />
            </div>
            <div>
              <p class="text-sm font-bold text-gray-200">Synthèse Fiscale Annuelle {{ filterYear }}</p>
              <p class="text-xs text-gray-500">Tous impôts confondus</p>
            </div>
          </div>
          <div class="grid grid-cols-2 md:grid-cols-5 gap-4">
            <div v-for="tax in taxSummaryCards" :key="tax.label" class="bg-black/20 rounded-xl p-3">
              <p class="text-xs text-gray-400 mb-1">{{ tax.label }}</p>
              <p :class="['text-lg font-bold', tax.color]">{{ fmtCur(tax.value) }}</p>
              <div v-if="tax.paid !== undefined" class="mt-1.5">
                <div class="h-1.5 bg-gray-800 rounded-full overflow-hidden">
                  <div class="h-full bg-emerald-500 rounded-full transition-all" :style="{ width: tax.value > 0 ? `${Math.min(100, (tax.paid / tax.value) * 100)}%` : '0%' }"></div>
                </div>
                <p class="text-xs text-gray-500 mt-0.5">Payé: {{ fmtCur(tax.paid) }}</p>
              </div>
            </div>
          </div>
          <div class="mt-4 pt-4 border-t border-orange-800/20 flex items-center justify-between">
            <div>
              <p class="text-xs text-gray-500">Total Impôts Dus</p>
              <p class="text-2xl font-black text-white">{{ fmtCur(report.total_tax_due || 0) }}</p>
            </div>
            <div class="text-right">
              <p class="text-xs text-gray-500">Total Payé</p>
              <p class="text-2xl font-black text-emerald-400">{{ fmtCur(report.total_paid || 0) }}</p>
            </div>
            <div class="text-right">
              <p class="text-xs text-gray-500">Solde Restant</p>
              <p :class="['text-2xl font-black', (report.total_balance || 0) > 0 ? 'text-red-400' : 'text-emerald-400']">
                {{ fmtCur(Math.abs(report.total_balance || 0)) }}
              </p>
            </div>
          </div>
        </div>

        <!-- Grid: Tax Breakdown + IBS -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Tax Type Breakdown Table -->
          <div class="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
            <div class="px-5 py-4 border-b border-gray-800 flex items-center gap-2">
              <BarChart3 :size="16" class="text-orange-400" />
              <span class="text-sm font-semibold text-gray-200">Ventilation par Nature d'Impôt</span>
            </div>
            <div class="p-4 space-y-3">
              <div v-for="item in taxBreakdown" :key="item.type" class="group">
                <div class="flex items-center justify-between mb-1">
                  <div class="flex items-center gap-2">
                    <span :class="['px-2 py-0.5 rounded text-xs font-bold', item.badge]">{{ item.type }}</span>
                    <span class="text-xs text-gray-500">{{ item.description }}</span>
                  </div>
                  <div class="text-right">
                    <p class="text-sm font-bold text-white">{{ fmtCur(item.due) }}</p>
                    <p class="text-xs text-emerald-400">payé: {{ fmtCur(item.paid) }}</p>
                  </div>
                </div>
                <div class="h-2 bg-gray-800 rounded-full overflow-hidden">
                  <div
                    :class="['h-full rounded-full transition-all', item.barColor]"
                    :style="{ width: item.due > 0 ? `${Math.min(100, (item.paid / item.due) * 100)}%` : '0%' }"
                  ></div>
                </div>
                <div class="flex justify-between mt-0.5">
                  <span class="text-xs text-gray-600">{{ item.due > 0 ? Math.round((item.paid / item.due) * 100) : 0 }}% payé</span>
                  <span class="text-xs" :class="(item.due - item.paid) > 0 ? 'text-red-400' : 'text-emerald-400'">
                    solde: {{ fmtCur(item.due - item.paid) }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- IBS Computation -->
          <div class="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
            <div class="px-5 py-4 border-b border-gray-800 flex items-center gap-2">
              <Building2 :size="16" class="text-amber-400" />
              <span class="text-sm font-semibold text-gray-200">Calcul IBS {{ filterYear }}</span>
            </div>
            <div class="p-4 space-y-4">
              <div class="grid grid-cols-2 gap-3">
                <div class="bg-gray-800 rounded-xl p-3">
                  <p class="text-xs text-gray-500 mb-1">Chiffre d'Affaires</p>
                  <p class="text-base font-bold text-white">{{ fmtCur(ibs.revenue || 0) }}</p>
                </div>
                <div class="bg-gray-800 rounded-xl p-3">
                  <p class="text-xs text-gray-500 mb-1">Charges Déductibles</p>
                  <p class="text-base font-bold text-amber-400">{{ fmtCur(ibs.expenses || 0) }}</p>
                </div>
                <div class="bg-gray-800 rounded-xl p-3">
                  <p class="text-xs text-gray-500 mb-1">Bénéfice Imposable</p>
                  <p class="text-base font-bold text-blue-400">{{ fmtCur(ibs.taxable_profit || 0) }}</p>
                </div>
                <div class="bg-gray-800 rounded-xl p-3">
                  <p class="text-xs text-gray-500 mb-1">Taux IBS</p>
                  <p class="text-base font-bold text-purple-400">{{ ibs.rate || 23 }}%</p>
                </div>
              </div>
              <div class="bg-gradient-to-r from-amber-950/50 to-orange-950/30 border border-amber-800/30 rounded-xl p-4">
                <div class="space-y-2 text-sm">
                  <div class="flex justify-between">
                    <span class="text-gray-400">Résultat fiscal brut</span>
                    <span class="text-blue-400 font-semibold">{{ fmtCur(ibs.taxable_profit || 0) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-400">IBS calculé ({{ ibs.rate || 23 }}%)</span>
                    <span class="text-amber-400 font-bold">{{ fmtCur(ibs.ibs_amount || 0) }}</span>
                  </div>
                  <div v-if="(ibs.ibs_minimum || 0) > 0" class="flex justify-between">
                    <span class="text-gray-400">IBS minimum (0.5% CA)</span>
                    <span class="text-gray-300">{{ fmtCur(ibs.ibs_minimum || 0) }}</span>
                  </div>
                  <div class="flex justify-between border-t border-amber-800/30 pt-2">
                    <span class="text-gray-200 font-bold">IBS Net Dû</span>
                    <span class="text-xl font-black text-white">{{ fmtCur(ibs.ibs_net_due || 0) }}</span>
                  </div>
                </div>
              </div>
              <p class="text-xs text-gray-600">Taux IBS: 23% (Production) / 26% (Autres activités) — Article 150 CIDTA</p>
            </div>
          </div>
        </div>

        <!-- Monthly Chart -->
        <div class="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
          <div class="px-5 py-4 border-b border-gray-800 flex items-center gap-2">
            <BarChart3 :size="16" class="text-blue-400" />
            <span class="text-sm font-semibold text-gray-200">Évolution Mensuelle — Impôts Dus vs Payés</span>
          </div>
          <div class="p-4">
            <div v-if="monthlyData.length === 0" class="text-center py-8 text-gray-500 text-sm">
              Aucune donnée mensuelle disponible
            </div>
            <div v-else class="overflow-x-auto">
              <div class="flex items-end gap-2 min-w-[600px] h-48">
                <div v-for="m in monthlyData" :key="m.month" class="flex-1 flex flex-col items-center gap-1">
                  <div class="w-full flex items-end gap-0.5 h-36">
                    <div class="flex-1 relative group">
                      <div
                        class="w-full bg-blue-600/70 hover:bg-blue-500 rounded-t transition-all cursor-pointer"
                        :style="{ height: `${maxMonthly > 0 ? (m.due / maxMonthly) * 100 : 0}%`, minHeight: m.due > 0 ? '4px' : '0' }"
                      ></div>
                      <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:block bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs text-gray-200 whitespace-nowrap z-10">
                        Dû: {{ fmtCur(m.due) }}
                      </div>
                    </div>
                    <div class="flex-1 relative group">
                      <div
                        class="w-full bg-emerald-600/70 hover:bg-emerald-500 rounded-t transition-all cursor-pointer"
                        :style="{ height: `${maxMonthly > 0 ? (m.paid / maxMonthly) * 100 : 0}%`, minHeight: m.paid > 0 ? '4px' : '0' }"
                      ></div>
                      <div class="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:block bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs text-gray-200 whitespace-nowrap z-10">
                        Payé: {{ fmtCur(m.paid) }}
                      </div>
                    </div>
                  </div>
                  <span class="text-xs text-gray-500">{{ m.label }}</span>
                </div>
              </div>
              <div class="flex items-center gap-4 mt-3 pl-2">
                <div class="flex items-center gap-1.5"><div class="w-3 h-3 bg-blue-600/70 rounded-sm"></div><span class="text-xs text-gray-400">Impôts Dus</span></div>
                <div class="flex items-center gap-1.5"><div class="w-3 h-3 bg-emerald-600/70 rounded-sm"></div><span class="text-xs text-gray-400">Impôts Payés</span></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Payment Status Breakdown -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <!-- Status donut-style breakdown -->
          <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
            <div class="flex items-center gap-2 mb-4">
              <PieChart :size="16" class="text-purple-400" />
              <span class="text-sm font-semibold text-gray-200">Statut des Paiements</span>
            </div>
            <div class="space-y-3">
              <div v-for="s in paymentStatusBreakdown" :key="s.label" class="flex items-center gap-3">
                <div class="w-2.5 h-2.5 rounded-full flex-shrink-0" :class="s.dot"></div>
                <div class="flex-1">
                  <div class="flex justify-between mb-0.5">
                    <span class="text-sm text-gray-300">{{ s.label }}</span>
                    <div class="flex items-center gap-3">
                      <span class="text-xs text-gray-500">{{ s.count }} pmt</span>
                      <span class="text-sm font-bold text-white">{{ fmtCur(s.amount) }}</span>
                    </div>
                  </div>
                  <div class="h-1.5 bg-gray-800 rounded-full overflow-hidden">
                    <div :class="['h-full rounded-full', s.bar]" :style="{ width: `${s.pct}%` }"></div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- TVA Quarterly Summary -->
          <div class="bg-gray-900 border border-gray-800 rounded-xl p-5">
            <div class="flex items-center gap-2 mb-4">
              <CalendarDays :size="16" class="text-blue-400" />
              <span class="text-sm font-semibold text-gray-200">TVA par Trimestre</span>
            </div>
            <div class="space-y-3">
              <div v-for="q in quarterlyVAT" :key="q.quarter" class="bg-gray-800 rounded-xl p-3">
                <div class="flex items-center justify-between mb-2">
                  <span class="text-xs font-bold text-gray-400 uppercase">{{ q.quarter }}</span>
                  <span class="text-xs text-gray-500">{{ q.declarations }} décl.</span>
                </div>
                <div class="grid grid-cols-3 gap-2 text-sm">
                  <div>
                    <p class="text-xs text-gray-500">Collectée</p>
                    <p class="font-semibold text-emerald-400">{{ fmtCur(q.collected) }}</p>
                  </div>
                  <div>
                    <p class="text-xs text-gray-500">Déductible</p>
                    <p class="font-semibold text-amber-400">{{ fmtCur(q.deductible) }}</p>
                  </div>
                  <div>
                    <p class="text-xs text-gray-500">Net Dû</p>
                    <p :class="['font-bold', q.net >= 0 ? 'text-white' : 'text-emerald-400']">{{ fmtCur(Math.abs(q.net)) }}</p>
                  </div>
                </div>
              </div>
              <div v-if="quarterlyVAT.length === 0" class="text-center py-6 text-gray-500 text-sm">
                Aucune donnée trimestrielle
              </div>
            </div>
          </div>
        </div>

        <!-- Detailed Monthly Table -->
        <div class="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
          <div class="px-5 py-4 border-b border-gray-800 flex items-center gap-2">
            <TableProperties :size="16" class="text-gray-400" />
            <span class="text-sm font-semibold text-gray-200">Tableau Récapitulatif Mensuel</span>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="bg-gray-800/60">
                  <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Mois</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold text-gray-400 uppercase tracking-wide">TVA Nette</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold text-gray-400 uppercase tracking-wide">TAP</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold text-gray-400 uppercase tracking-wide">IRG</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold text-gray-400 uppercase tracking-wide">Total Dû</th>
                  <th class="px-4 py-3 text-right text-xs font-semibold text-gray-400 uppercase tracking-wide">Payé</th>
                  <th class="px-4 py-3 text-center text-xs font-semibold text-gray-400 uppercase tracking-wide">Statut</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="m in monthlyTable" :key="m.month" class="border-t border-gray-800 hover:bg-gray-800/40 transition-colors">
                  <td class="px-4 py-3 text-gray-300 font-medium">{{ m.monthLabel }}</td>
                  <td class="px-4 py-3 text-right text-blue-400 font-semibold">{{ fmtCur(m.tva) }}</td>
                  <td class="px-4 py-3 text-right text-purple-400">{{ fmtCur(m.tap) }}</td>
                  <td class="px-4 py-3 text-right text-teal-400">{{ fmtCur(m.irg) }}</td>
                  <td class="px-4 py-3 text-right text-white font-bold">{{ fmtCur(m.due) }}</td>
                  <td class="px-4 py-3 text-right text-emerald-400 font-semibold">{{ fmtCur(m.paid) }}</td>
                  <td class="px-4 py-3 text-center">
                    <span v-if="m.due === 0" class="px-2 py-0.5 bg-gray-800 text-gray-500 rounded-full text-xs">—</span>
                    <span v-else-if="m.paid >= m.due" class="px-2 py-0.5 bg-emerald-900/50 text-emerald-300 rounded-full text-xs font-medium">Soldé</span>
                    <span v-else class="px-2 py-0.5 bg-yellow-900/50 text-yellow-300 rounded-full text-xs font-medium">Partiel</span>
                  </td>
                </tr>
              </tbody>
              <tfoot>
                <tr class="bg-gray-800/80 border-t-2 border-gray-700">
                  <td class="px-4 py-3 text-sm font-bold text-gray-300">TOTAL {{ filterYear }}</td>
                  <td class="px-4 py-3 text-right text-sm font-bold text-blue-400">{{ fmtCur(yearTotals.tva) }}</td>
                  <td class="px-4 py-3 text-right text-sm font-bold text-purple-400">{{ fmtCur(yearTotals.tap) }}</td>
                  <td class="px-4 py-3 text-right text-sm font-bold text-teal-400">{{ fmtCur(yearTotals.irg) }}</td>
                  <td class="px-4 py-3 text-right text-sm font-bold text-white">{{ fmtCur(yearTotals.due) }}</td>
                  <td class="px-4 py-3 text-right text-sm font-bold text-emerald-400">{{ fmtCur(yearTotals.paid) }}</td>
                  <td></td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  FileBarChart2, Download, TrendingUp, BarChart3, Building2,
  PieChart, CalendarDays, TableProperties
} from 'lucide-vue-next'
import { taxAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()
const loading = ref(false)

const now = new Date()
const filterYear = ref(now.getFullYear())
const report = ref<any>({})
const ibs = ref<any>({})
const payments = ref<any[]>([])
const vatReturns = ref<any[]>([])

const yearOptions = computed(() => {
  const y = now.getFullYear()
  return [y + 1, y, y - 1, y - 2, y - 3]
})

const monthLabels = ['Jan', 'Fév', 'Mar', 'Avr', 'Mai', 'Jun', 'Jul', 'Aoû', 'Sep', 'Oct', 'Nov', 'Déc']
const monthsFull = ['Janvier', 'Février', 'Mars', 'Avril', 'Mai', 'Juin', 'Juillet', 'Août', 'Septembre', 'Octobre', 'Novembre', 'Décembre']

async function loadReport() {
  loading.value = true
  try {
    const [reportRes, ibsRes] = await Promise.all([
      taxAPI.getTaxReport({ year: filterYear.value }),
      taxAPI.getIBS({ year: filterYear.value }).catch(() => ({ data: {} }))
    ])
    report.value = reportRes.data || {}
    ibs.value = ibsRes.data || {}

    // Also load payments and vat returns for derived computations
    const [pmtsRes, vatsRes] = await Promise.all([
      taxAPI.listTaxPayments({ year: filterYear.value }).catch(() => ({ data: [] })),
      taxAPI.listVATReturns({ year: filterYear.value }).catch(() => ({ data: [] }))
    ])
    payments.value = pmtsRes.data?.payments || pmtsRes.data || []
    vatReturns.value = vatsRes.data?.returns || vatsRes.data || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur chargement rapport', 'error')
  } finally {
    loading.value = false
  }
}

const taxSummaryCards = computed(() => {
  const r = report.value
  return [
    { label: 'TVA Nette', value: r.tva_net_due || 0, paid: r.tva_paid || 0, color: 'text-blue-400' },
    { label: 'TAP', value: r.tap_due || 0, paid: r.tap_paid || 0, color: 'text-purple-400' },
    { label: 'IBS', value: r.ibs_net_due || ibs.value.ibs_net_due || 0, paid: r.ibs_paid || 0, color: 'text-amber-400' },
    { label: 'IRG', value: r.irg_due || 0, paid: r.irg_paid || 0, color: 'text-teal-400' },
    { label: 'Timbre', value: r.stamp_due || 0, paid: r.stamp_paid || 0, color: 'text-rose-400' }
  ]
})

const taxBreakdown = computed(() => [
  {
    type: 'TVA', description: 'Taxe sur Valeur Ajoutée (19%/9%/0%)',
    due: report.value.tva_net_due || 0, paid: report.value.tva_paid || 0,
    badge: 'bg-blue-900/60 text-blue-300', barColor: 'bg-blue-500'
  },
  {
    type: 'TAP', description: 'Taxe sur Activité Professionnelle (2%)',
    due: report.value.tap_due || 0, paid: report.value.tap_paid || 0,
    badge: 'bg-purple-900/60 text-purple-300', barColor: 'bg-purple-500'
  },
  {
    type: 'IBS', description: `Impôt sur Bénéfice des Sociétés (${ibs.value.rate || 23}%)`,
    due: report.value.ibs_net_due || ibs.value.ibs_net_due || 0, paid: report.value.ibs_paid || 0,
    badge: 'bg-amber-900/60 text-amber-300', barColor: 'bg-amber-500'
  },
  {
    type: 'IRG', description: 'Impôt sur Revenu Global (Salaires)',
    due: report.value.irg_due || 0, paid: report.value.irg_paid || 0,
    badge: 'bg-teal-900/60 text-teal-300', barColor: 'bg-teal-500'
  },
  {
    type: 'TIMBRE', description: 'Timbre Fiscal',
    due: report.value.stamp_due || 0, paid: report.value.stamp_paid || 0,
    badge: 'bg-gray-800 text-gray-400', barColor: 'bg-gray-500'
  }
])

const monthlyData = computed(() => {
  const months = Array.from({ length: 12 }, (_, i) => {
    const pmts = payments.value.filter(p => {
      const d = p.due_date || p.period_start
      if (!d) return false
      return new Date(d).getMonth() === i && new Date(d).getFullYear() === filterYear.value
    })
    return {
      month: i + 1,
      label: monthLabels[i],
      due: pmts.reduce((s, p) => s + (p.amount_due || 0), 0),
      paid: pmts.reduce((s, p) => s + (p.amount_paid || 0), 0)
    }
  })
  return months
})

const maxMonthly = computed(() => Math.max(...monthlyData.value.map(m => Math.max(m.due, m.paid)), 1))

const paymentStatusBreakdown = computed(() => {
  const all = payments.value
  const total = all.reduce((s, p) => s + (p.amount_due || 0), 0)
  const groups = [
    { label: 'Payé intégralement', filter: (p: any) => p.status === 'paid', dot: 'bg-emerald-400', bar: 'bg-emerald-500' },
    { label: 'Paiement partiel', filter: (p: any) => p.status === 'partial', dot: 'bg-blue-400', bar: 'bg-blue-500' },
    { label: 'En attente', filter: (p: any) => p.status === 'pending' && new Date(p.due_date || '') >= new Date(), dot: 'bg-yellow-400', bar: 'bg-yellow-500' },
    { label: 'En retard', filter: (p: any) => p.status !== 'paid' && new Date(p.due_date || '') < new Date(), dot: 'bg-red-400', bar: 'bg-red-500' }
  ]
  return groups.map(g => {
    const subset = all.filter(g.filter)
    const amount = subset.reduce((s, p) => s + (p.amount_due || 0), 0)
    return { ...g, count: subset.length, amount, pct: total > 0 ? (amount / total) * 100 : 0 }
  })
})

const quarterlyVAT = computed(() => {
  const quarters = [
    { quarter: 'T1 (Jan-Mar)', months: [1, 2, 3] },
    { quarter: 'T2 (Avr-Jun)', months: [4, 5, 6] },
    { quarter: 'T3 (Jul-Sep)', months: [7, 8, 9] },
    { quarter: 'T4 (Oct-Déc)', months: [10, 11, 12] }
  ]
  return quarters.map(q => {
    const returns = vatReturns.value.filter(r => {
      if (!r.period_start) return false
      const m = new Date(r.period_start).getMonth() + 1
      return q.months.includes(m)
    })
    return {
      quarter: q.quarter,
      declarations: returns.length,
      collected: returns.reduce((s, r) => s + (r.tva_collected || 0), 0),
      deductible: returns.reduce((s, r) => s + (r.tva_deductible || 0), 0),
      net: returns.reduce((s, r) => s + (r.tva_net_due || 0), 0)
    }
  })
})

const monthlyTable = computed(() => {
  return Array.from({ length: 12 }, (_, i) => {
    const pmts = payments.value.filter(p => {
      const d = p.due_date || p.period_start
      if (!d) return false
      return new Date(d).getMonth() === i && new Date(d).getFullYear() === filterYear.value
    })
    const tva = pmts.filter(p => p.tax_type === 'TVA').reduce((s, p) => s + (p.amount_due || 0), 0)
    const tap = pmts.filter(p => p.tax_type === 'TAP').reduce((s, p) => s + (p.amount_due || 0), 0)
    const irg = pmts.filter(p => ['IRG', 'IRG_SALAIRES'].includes(p.tax_type || '')).reduce((s, p) => s + (p.amount_due || 0), 0)
    const due = pmts.reduce((s, p) => s + (p.amount_due || 0), 0)
    const paid = pmts.reduce((s, p) => s + (p.amount_paid || 0), 0)
    return { month: i + 1, monthLabel: monthsFull[i], tva, tap, irg, due, paid }
  })
})

const yearTotals = computed(() => ({
  tva: monthlyTable.value.reduce((s, m) => s + m.tva, 0),
  tap: monthlyTable.value.reduce((s, m) => s + m.tap, 0),
  irg: monthlyTable.value.reduce((s, m) => s + m.irg, 0),
  due: monthlyTable.value.reduce((s, m) => s + m.due, 0),
  paid: monthlyTable.value.reduce((s, m) => s + m.paid, 0)
}))

function exportReport() {
  const rows = [
    ['Mois', 'TVA Nette', 'TAP', 'IRG', 'Total Dû', 'Total Payé', 'Solde'],
    ...monthlyTable.value.map(m => [m.monthLabel, m.tva, m.tap, m.irg, m.due, m.paid, m.due - m.paid]),
    ['TOTAL', yearTotals.value.tva, yearTotals.value.tap, yearTotals.value.irg, yearTotals.value.due, yearTotals.value.paid, yearTotals.value.due - yearTotals.value.paid]
  ]
  const csv = rows.map(r => r.join(';')).join('\n')
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `rapport_fiscal_${filterYear.value}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

function fmtCur(v: number) {
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0) + ' DZD'
}

onMounted(loadReport)
</script>
