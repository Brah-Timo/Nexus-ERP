<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import {
  TrendingUp, TrendingDown, DollarSign, ShoppingCart, Package,
  Users, AlertCircle, BarChart3, RefreshCw, Calendar, ArrowUpRight,
  ArrowDownRight, Wallet, CreditCard, Clock, Target
} from 'lucide-vue-next'
import { reportsAPI } from '@/api/client'

const app = useAppStore()

const loading = ref(false)
const selectedYear = ref(new Date().getFullYear())
const years = Array.from({ length: 5 }, (_, i) => new Date().getFullYear() - i)

interface DashboardData {
  year: number
  revenue: number
  expenses: number
  net_profit: number
  profit_margin: number
  accounts_receivable: number
  accounts_payable: number
  cash_balance: number
  bank_balance: number
  total_cash: number
  inventory_value: number
  active_orders: number
  overdue_invoices: number
  monthly_trend: Array<{ month: number; revenue: number; expenses: number }>
  top_customers: Array<{ name: string; revenue: number }>
  top_products: Array<{ name: string; revenue: number }>
}

const data = ref<DashboardData | null>(null)

const months = ['Jan','Fév','Mar','Avr','Mai','Jun','Jul','Aoû','Sep','Oct','Nov','Déc']

async function load() {
  loading.value = true
  try {
    const r = await reportsAPI.getBIDashboard(String(selectedYear.value))
    data.value = r.data
  } catch {
    app.addToast('Erreur lors du chargement du tableau de bord', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(load)

const fmt = (n: number) =>
  new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(n) + ' DZD'
const fmtPct = (n: number) => n.toFixed(1) + '%'
const fmtNum = (n: number) => new Intl.NumberFormat('fr-DZ').format(n)

// Chart helpers
const maxMonthlyRevenue = computed(() => {
  if (!data.value) return 1
  return Math.max(...data.value.monthly_trend.map(m => Math.max(m.revenue, m.expenses)), 1)
})

function barHeight(val: number): string {
  const pct = (val / maxMonthlyRevenue.value) * 100
  return Math.max(pct, 1).toFixed(1) + '%'
}

const maxTopCustomer = computed(() => {
  if (!data.value?.top_customers?.length) return 1
  return Math.max(...data.value.top_customers.map(c => c.revenue), 1)
})
const maxTopProduct = computed(() => {
  if (!data.value?.top_products?.length) return 1
  return Math.max(...data.value.top_products.map(p => p.revenue), 1)
})
</script>

<template>
  <div class="min-h-screen p-6 space-y-6"
    :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">

    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <BarChart3 class="w-6 h-6 text-indigo-500" />
          Tableau de Bord BI
        </h1>
        <p class="text-sm mt-0.5" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
          Vue d'ensemble consolidée — indicateurs financiers et opérationnels
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
        <button @click="load"
          class="inline-flex items-center gap-2 px-3 py-2 rounded-lg border text-sm font-medium transition-colors"
          :class="app.darkMode ? 'bg-slate-900 border-slate-700 hover:bg-slate-800 text-slate-200' : 'bg-white border-slate-200 hover:bg-slate-50 text-slate-700'">
          <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          Actualiser
        </button>
      </div>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="i in 8" :key="i" class="h-28 rounded-xl animate-pulse"
        :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-200'" />
    </div>

    <template v-else-if="data">
      <!-- KPI Row 1 — P&L -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- Revenue -->
        <div class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs font-medium uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Chiffre d'Affaires</p>
              <p class="text-2xl font-bold mt-1">{{ data ? fmt(data.revenue) : '—' }}</p>
            </div>
            <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-emerald-500/10">
              <TrendingUp class="w-5 h-5 text-emerald-500" />
            </div>
          </div>
          <p class="text-xs mt-2" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
            HT — factures confirmées {{ selectedYear }}
          </p>
        </div>

        <!-- Expenses -->
        <div class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs font-medium uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Achats / Charges</p>
              <p class="text-2xl font-bold mt-1">{{ fmt(data.expenses) }}</p>
            </div>
            <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-red-500/10">
              <TrendingDown class="w-5 h-5 text-red-500" />
            </div>
          </div>
          <p class="text-xs mt-2" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Factures fournisseurs</p>
        </div>

        <!-- Net Profit -->
        <div class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs font-medium uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Résultat Net</p>
              <p class="text-2xl font-bold mt-1" :class="data.net_profit >= 0 ? 'text-emerald-500' : 'text-red-500'">
                {{ fmt(data.net_profit) }}
              </p>
            </div>
            <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-indigo-500/10">
              <DollarSign class="w-5 h-5 text-indigo-500" />
            </div>
          </div>
          <p class="text-xs mt-2" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
            Marge : {{ fmtPct(data.profit_margin) }}
          </p>
        </div>

        <!-- Cash -->
        <div class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs font-medium uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Position de Trésorerie</p>
              <p class="text-2xl font-bold mt-1">{{ fmt(data.total_cash) }}</p>
            </div>
            <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-sky-500/10">
              <Wallet class="w-5 h-5 text-sky-500" />
            </div>
          </div>
          <p class="text-xs mt-2" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
            Caisse: {{ fmt(data.cash_balance) }} | Banque: {{ fmt(data.bank_balance) }}
          </p>
        </div>
      </div>

      <!-- KPI Row 2 — Operational -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs font-medium uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Créances Clients</p>
              <p class="text-2xl font-bold mt-1">{{ fmt(data.accounts_receivable) }}</p>
            </div>
            <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-amber-500/10">
              <ArrowUpRight class="w-5 h-5 text-amber-500" />
            </div>
          </div>
          <p class="text-xs mt-2 text-red-400" v-if="data.overdue_invoices > 0">
            {{ fmtNum(data.overdue_invoices) }} factures en retard
          </p>
        </div>
        <div class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs font-medium uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Dettes Fournisseurs</p>
              <p class="text-2xl font-bold mt-1">{{ fmt(data.accounts_payable) }}</p>
            </div>
            <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-orange-500/10">
              <ArrowDownRight class="w-5 h-5 text-orange-500" />
            </div>
          </div>
          <p class="text-xs mt-2" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Factures impayées</p>
        </div>
        <div class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs font-medium uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Stock</p>
              <p class="text-2xl font-bold mt-1">{{ fmt(data.inventory_value) }}</p>
            </div>
            <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-violet-500/10">
              <Package class="w-5 h-5 text-violet-500" />
            </div>
          </div>
          <p class="text-xs mt-2" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Valeur d'inventaire</p>
        </div>
        <div class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="flex items-start justify-between">
            <div>
              <p class="text-xs font-medium uppercase tracking-wide" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Commandes Actives</p>
              <p class="text-2xl font-bold mt-1">{{ fmtNum(data.active_orders) }}</p>
            </div>
            <div class="w-10 h-10 rounded-xl flex items-center justify-center bg-teal-500/10">
              <ShoppingCart class="w-5 h-5 text-teal-500" />
            </div>
          </div>
          <p class="text-xs mt-2" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">En cours de traitement</p>
        </div>
      </div>

      <!-- Chart + Top Tables -->
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">

        <!-- Monthly Revenue vs Expenses chart -->
        <div class="lg:col-span-2 rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <div class="flex items-center justify-between mb-4">
            <h2 class="font-semibold">Tendance mensuelle {{ selectedYear }}</h2>
            <div class="flex items-center gap-4 text-xs">
              <span class="flex items-center gap-1.5">
                <span class="w-3 h-3 rounded-sm bg-indigo-500 inline-block" />
                Revenus
              </span>
              <span class="flex items-center gap-1.5">
                <span class="w-3 h-3 rounded-sm bg-red-400 inline-block" />
                Charges
              </span>
            </div>
          </div>
          <div class="flex items-end gap-1 h-48">
            <template v-for="(m, idx) in data.monthly_trend" :key="idx">
              <div class="flex-1 flex flex-col items-center gap-0.5">
                <div class="w-full flex items-end gap-0.5 h-40">
                  <div class="flex-1 rounded-t transition-all duration-500 bg-indigo-500"
                    :style="{ height: barHeight(m.revenue) }"
                    :title="`Revenus: ${fmt(m.revenue)}`" />
                  <div class="flex-1 rounded-t transition-all duration-500 bg-red-400"
                    :style="{ height: barHeight(m.expenses) }"
                    :title="`Charges: ${fmt(m.expenses)}`" />
                </div>
                <span class="text-[10px] leading-none" :class="app.darkMode ? 'text-slate-500' : 'text-slate-400'">
                  {{ months[idx] }}
                </span>
              </div>
            </template>
          </div>
        </div>

        <!-- Top 5 Customers -->
        <div class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <h2 class="font-semibold mb-4">Top 5 Clients</h2>
          <div class="space-y-3">
            <div v-for="(c, i) in data.top_customers" :key="i">
              <div class="flex items-center justify-between text-sm mb-1">
                <span class="truncate max-w-[60%]" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">{{ c.name }}</span>
                <span class="font-medium text-xs" :class="app.darkMode ? 'text-slate-200' : 'text-slate-800'">
                  {{ new Intl.NumberFormat('fr-DZ').format(c.revenue) }}
                </span>
              </div>
              <div class="h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                <div class="h-full rounded-full bg-indigo-500 transition-all duration-500"
                  :style="{ width: ((c.revenue / maxTopCustomer) * 100).toFixed(1) + '%' }" />
              </div>
            </div>
            <p v-if="!data.top_customers?.length" class="text-sm text-slate-400 text-center py-4">Aucune donnée</p>
          </div>
        </div>
      </div>

      <!-- Top Products + Alerts -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">

        <!-- Top Products -->
        <div class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <h2 class="font-semibold mb-4">Top 5 Produits / Services</h2>
          <div class="space-y-3">
            <div v-for="(p, i) in data.top_products" :key="i">
              <div class="flex items-center justify-between text-sm mb-1">
                <span class="flex items-center gap-2">
                  <span class="text-xs font-bold w-5 h-5 rounded flex items-center justify-center bg-indigo-500 text-white">{{ i+1 }}</span>
                  <span class="truncate max-w-[55%]" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">{{ p.name }}</span>
                </span>
                <span class="font-medium text-xs">{{ new Intl.NumberFormat('fr-DZ').format(p.revenue) }}</span>
              </div>
              <div class="h-1.5 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                <div class="h-full rounded-full bg-emerald-500 transition-all duration-500"
                  :style="{ width: ((p.revenue / maxTopProduct) * 100).toFixed(1) + '%' }" />
              </div>
            </div>
            <p v-if="!data.top_products?.length" class="text-sm text-slate-400 text-center py-4">Aucune donnée</p>
          </div>
        </div>

        <!-- Financial Health Summary -->
        <div class="rounded-xl border p-5"
          :class="app.darkMode ? 'bg-slate-900 border-slate-800' : 'bg-white border-slate-200'">
          <h2 class="font-semibold mb-4 flex items-center gap-2">
            <Target class="w-4 h-4 text-indigo-500" />
            Santé Financière
          </h2>
          <div class="space-y-4">
            <!-- Profit margin gauge -->
            <div>
              <div class="flex justify-between text-sm mb-1.5">
                <span :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Marge nette</span>
                <span class="font-semibold" :class="data.profit_margin >= 10 ? 'text-emerald-500' : data.profit_margin >= 0 ? 'text-amber-500' : 'text-red-500'">
                  {{ fmtPct(data.profit_margin) }}
                </span>
              </div>
              <div class="h-2 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                <div class="h-full rounded-full transition-all duration-700"
                  :class="data.profit_margin >= 10 ? 'bg-emerald-500' : data.profit_margin >= 0 ? 'bg-amber-500' : 'bg-red-500'"
                  :style="{ width: Math.min(Math.max(data.profit_margin, 0), 100).toFixed(1) + '%' }" />
              </div>
            </div>
            <!-- AR/AP ratio -->
            <div>
              <div class="flex justify-between text-sm mb-1.5">
                <span :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Ratio Créances/Dettes</span>
                <span class="font-semibold">
                  {{ data.accounts_payable > 0 ? (data.accounts_receivable / data.accounts_payable).toFixed(2) : 'N/A' }}
                </span>
              </div>
              <div class="h-2 rounded-full" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-100'">
                <div class="h-full rounded-full bg-sky-500 transition-all duration-700"
                  :style="{ width: Math.min(data.accounts_payable > 0 ? (data.accounts_receivable / data.accounts_payable * 50) : 50, 100).toFixed(1) + '%' }" />
              </div>
            </div>
            <!-- Liquidity -->
            <div class="pt-2 border-t" :class="app.darkMode ? 'border-slate-800' : 'border-slate-100'">
              <div class="grid grid-cols-2 gap-3">
                <div class="rounded-lg p-3 text-center" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-50'">
                  <p class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Trésorerie disponible</p>
                  <p class="text-sm font-bold mt-1">{{ fmt(data.total_cash) }}</p>
                </div>
                <div class="rounded-lg p-3 text-center" :class="app.darkMode ? 'bg-slate-800' : 'bg-slate-50'">
                  <p class="text-xs" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">Factures en retard</p>
                  <p class="text-sm font-bold mt-1" :class="data.overdue_invoices > 0 ? 'text-red-500' : 'text-emerald-500'">
                    {{ fmtNum(data.overdue_invoices) }}
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Empty state -->
    <div v-else-if="!loading" class="text-center py-20">
      <BarChart3 class="w-12 h-12 mx-auto text-slate-300 mb-3" />
      <p class="text-slate-400">Aucune donnée disponible</p>
      <button @click="load" class="mt-3 text-sm text-indigo-500 hover:underline">Actualiser</button>
    </div>
  </div>
</template>
