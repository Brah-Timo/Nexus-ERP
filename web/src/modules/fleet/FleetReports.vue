<template>
  <div :class="dk('min-h-screen bg-gray-950 text-gray-100','min-h-screen bg-gray-50 text-gray-900')">
    <!-- Header -->
    <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border-b px-6 py-4'">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-lg bg-cyan-600/20">
            <BarChart3 class="w-5 h-5 text-cyan-400" />
          </div>
          <div>
            <h1 class="text-lg font-semibold">Fleet Reports</h1>
            <p :class="dk('text-gray-400','text-gray-500')" class="text-xs">Fleet performance analytics and insights</p>
          </div>
        </div>
        <button @click="load" :disabled="loading" :class="dk('bg-gray-800 hover:bg-gray-700 text-gray-300','bg-gray-100 hover:bg-gray-200 text-gray-700')+ ' flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors disabled:opacity-50'">
          <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          Refresh
        </button>
      </div>
    </div>

    <div class="p-6 space-y-6">
      <!-- KPI Grid -->
      <div class="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-4">
        <div v-for="kpi in kpis" :key="kpi.label" :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-4'">
          <div class="flex items-center justify-between mb-3">
            <span :class="dk('text-gray-400','text-gray-500')" class="text-xs font-medium">{{ kpi.label }}</span>
            <div :class="'p-1.5 rounded-lg ' + kpi.bg">
              <component :is="kpi.icon" :class="'w-4 h-4 ' + kpi.color" />
            </div>
          </div>
          <div class="text-2xl font-bold">{{ kpi.value }}</div>
          <div v-if="kpi.sub" :class="'text-xs mt-1 ' + kpi.subColor">{{ kpi.sub }}</div>
        </div>
      </div>

      <!-- Fleet Status + Vehicle Distribution -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Vehicle Status -->
        <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-5'">
          <h3 class="text-sm font-semibold mb-4 flex items-center gap-2">
            <Truck class="w-4 h-4 text-blue-400" />Vehicle Status Distribution
          </h3>
          <div class="space-y-3">
            <div v-for="s in vehicleStatusData" :key="s.label" class="flex items-center gap-3">
              <span :class="dk('text-gray-400','text-gray-500')" class="text-xs w-24 shrink-0">{{ s.label }}</span>
              <div class="flex-1 h-5 rounded-full overflow-hidden" :class="dk('bg-gray-800','bg-gray-100')">
                <div :class="'h-full rounded-full transition-all ' + s.barColor" :style="{ width: totalVehicles > 0 ? (s.count / totalVehicles * 100) + '%' : '0%' }" />
              </div>
              <span class="text-sm font-bold w-8 text-right">{{ s.count }}</span>
            </div>
          </div>
        </div>

        <!-- Expense by Type -->
        <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-5'">
          <h3 class="text-sm font-semibold mb-4 flex items-center gap-2">
            <PieChart class="w-4 h-4 text-rose-400" />Expense Breakdown
          </h3>
          <div class="space-y-3">
            <div v-for="e in expenseBreakdown" :key="e.type" class="flex items-center gap-3">
              <span :class="dk('text-gray-400','text-gray-500')" class="text-xs w-24 shrink-0">{{ formatEnum(e.type) }}</span>
              <div class="flex-1 h-5 rounded-full overflow-hidden" :class="dk('bg-gray-800','bg-gray-100')">
                <div class="h-full rounded-full transition-all bg-rose-500" :style="{ width: totalExpenses > 0 ? (e.amount / totalExpenses * 100) + '%' : '0%' }" />
              </div>
              <span :class="dk('text-gray-300','text-gray-700')" class="text-xs font-medium w-24 text-right">{{ fmtDZD(e.amount) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Monthly Cost Chart -->
      <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-5'">
        <h3 class="text-sm font-semibold mb-6 flex items-center gap-2">
          <TrendingUp class="w-4 h-4 text-cyan-400" />Monthly Fleet Costs (Last 12 Months)
        </h3>
        <div class="flex items-end gap-2 h-40" v-if="monthlyData.length > 0">
          <div v-for="m in monthlyData" :key="m.month" class="flex-1 flex flex-col items-center gap-1">
            <span :class="dk('text-gray-400','text-gray-500')" class="text-xs">{{ fmtK(m.total) }}</span>
            <div class="w-full rounded-t-sm transition-all bg-cyan-500/70 hover:bg-cyan-500"
              :style="{ height: maxMonthly > 0 ? Math.max(4, (m.total / maxMonthly) * 120) + 'px' : '4px' }"
              :title="m.month + ': ' + fmtDZD(m.total)" />
            <span :class="dk('text-gray-500','text-gray-400')" class="text-xs">{{ m.month.slice(5) }}</span>
          </div>
        </div>
        <div v-else class="h-40 flex items-center justify-center text-gray-400 text-sm">No monthly data available</div>
      </div>

      <!-- Top Vehicles by Cost -->
      <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-5'">
        <h3 class="text-sm font-semibold mb-4 flex items-center gap-2">
          <Trophy class="w-4 h-4 text-amber-400" />Top Vehicles by Total Cost
        </h3>
        <div class="space-y-3">
          <div v-if="topVehicles.length === 0" class="text-center text-gray-400 py-6 text-sm">No vehicle cost data</div>
          <div v-for="(v, idx) in topVehicles" :key="v.vehicle_id" class="flex items-center gap-4">
            <span :class="['font-bold text-sm w-6', idx === 0 ? 'text-amber-400' : idx === 1 ? 'text-gray-400' : idx === 2 ? 'text-orange-600' : dk('text-gray-500','text-gray-400')]">
              #{{ idx + 1 }}
            </span>
            <div class="flex-1">
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm font-medium">{{ v.vehicle_plate || v.vehicle_id }}</span>
                <span class="text-sm font-bold text-amber-400">{{ fmtDZD(v.total_cost) }}</span>
              </div>
              <div class="h-1.5 rounded-full overflow-hidden" :class="dk('bg-gray-800','bg-gray-100')">
                <div class="h-full rounded-full bg-amber-500/70"
                  :style="{ width: topVehicles[0]?.total_cost > 0 ? (v.total_cost / topVehicles[0].total_cost * 100) + '%' : '0%' }" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Recent Assignments + Driver Stats -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- Driver Rankings -->
        <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-5'">
          <h3 class="text-sm font-semibold mb-4 flex items-center gap-2">
            <UserCheck class="w-4 h-4 text-indigo-400" />Driver Status Overview
          </h3>
          <div class="space-y-2">
            <div v-for="d in driverStatusData" :key="d.status" class="flex items-center justify-between p-2 rounded-lg" :class="dk('bg-gray-800/50','bg-gray-50')">
              <div class="flex items-center gap-2">
                <span :class="d.dot + ' w-2 h-2 rounded-full'" />
                <span class="text-sm">{{ formatEnum(d.status) }}</span>
              </div>
              <span class="font-bold text-sm">{{ d.count }}</span>
            </div>
          </div>
        </div>

        <!-- Fuel Summary -->
        <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-5'">
          <h3 class="text-sm font-semibold mb-4 flex items-center gap-2">
            <Fuel class="w-4 h-4 text-yellow-400" />Fuel Summary
          </h3>
          <div class="space-y-3">
            <div :class="dk('bg-gray-800/50','bg-gray-50')+ ' rounded-lg p-3 flex items-center justify-between'">
              <span :class="dk('text-gray-400','text-gray-500')" class="text-sm">Total Liters Consumed</span>
              <span class="font-bold text-yellow-400">{{ report.total_fuel_liters?.toFixed(0) || 0 }} L</span>
            </div>
            <div :class="dk('bg-gray-800/50','bg-gray-50')+ ' rounded-lg p-3 flex items-center justify-between'">
              <span :class="dk('text-gray-400','text-gray-500')" class="text-sm">Total Fuel Cost</span>
              <span class="font-bold text-yellow-400">{{ fmtDZD(report.total_fuel_cost || 0) }}</span>
            </div>
            <div :class="dk('bg-gray-800/50','bg-gray-50')+ ' rounded-lg p-3 flex items-center justify-between'">
              <span :class="dk('text-gray-400','text-gray-500')" class="text-sm">Avg Price per Liter</span>
              <span class="font-bold">{{ report.avg_fuel_price?.toFixed(2) || '—' }} DZD</span>
            </div>
            <div :class="dk('bg-gray-800/50','bg-gray-50')+ ' rounded-lg p-3 flex items-center justify-between'">
              <span :class="dk('text-gray-400','text-gray-500')" class="text-sm">Total Fill Records</span>
              <span class="font-bold">{{ report.total_fuel_records || 0 }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  BarChart3, RefreshCw, Truck, PieChart, TrendingUp, Trophy, UserCheck, Fuel,
  Car, Wrench, DollarSign, Users, Activity, ShieldCheck
} from 'lucide-vue-next'
import { fleetAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const dk = (dark: string, light: string) => app.darkMode ? dark : light

const loading = ref(false)
const report = ref<any>({})
const vehicles = ref<any[]>([])
const drivers = ref<any[]>([])
const expenses = ref<any[]>([])
const fuelLogs = ref<any[]>([])
const assignments = ref<any[]>([])

const kpis = computed(() => [
  { label: 'Total Vehicles', value: vehicles.value.length, icon: Truck, bg: 'bg-blue-500/15', color: 'text-blue-400', sub: null, subColor: '' },
  { label: 'Active Vehicles', value: vehicles.value.filter(v => v.status === 'active' || v.status === 'in_use').length, icon: Car, bg: 'bg-green-500/15', color: 'text-green-400', sub: null, subColor: '' },
  { label: 'Total Drivers', value: drivers.value.length, icon: Users, bg: 'bg-indigo-500/15', color: 'text-indigo-400', sub: null, subColor: '' },
  { label: 'In Service', value: vehicles.value.filter(v => v.status === 'maintenance').length, icon: Wrench, bg: 'bg-amber-500/15', color: 'text-amber-400', sub: null, subColor: '' },
  { label: 'Total Expenses', value: fmtDZD(totalExpenses.value), icon: DollarSign, bg: 'bg-rose-500/15', color: 'text-rose-400', sub: null, subColor: '' },
  { label: 'Active Assignments', value: assignments.value.filter(a => a.status === 'active').length, icon: Activity, bg: 'bg-cyan-500/15', color: 'text-cyan-400', sub: null, subColor: '' },
])

const totalVehicles = computed(() => vehicles.value.length)

const vehicleStatusData = computed(() => [
  { label: 'Active', count: vehicles.value.filter(v => v.status === 'active').length, barColor: 'bg-green-500' },
  { label: 'In Use', count: vehicles.value.filter(v => v.status === 'in_use').length, barColor: 'bg-blue-500' },
  { label: 'Maintenance', count: vehicles.value.filter(v => v.status === 'maintenance').length, barColor: 'bg-amber-500' },
  { label: 'Inactive', count: vehicles.value.filter(v => v.status === 'inactive').length, barColor: 'bg-gray-500' },
  { label: 'Retired', count: vehicles.value.filter(v => v.status === 'retired').length, barColor: 'bg-red-500' },
].filter(s => s.count > 0))

const totalExpenses = computed(() => expenses.value.reduce((s, e) => s + (e.amount || 0), 0))

const expenseBreakdown = computed(() => {
  const m: Record<string, number> = {}
  for (const e of expenses.value) { m[e.expense_type] = (m[e.expense_type] || 0) + (e.amount || 0) }
  return Object.entries(m).map(([type, amount]) => ({ type, amount })).sort((a, b) => b.amount - a.amount).slice(0, 6)
})

const monthlyData = computed(() => {
  if (report.value.monthly_costs) return report.value.monthly_costs
  // Build from expenses
  const m: Record<string, number> = {}
  for (const e of expenses.value) {
    const mo = e.expense_date?.slice(0, 7)
    if (mo) m[mo] = (m[mo] || 0) + (e.amount || 0)
  }
  for (const f of fuelLogs.value) {
    const mo = f.fill_date?.slice(0, 7)
    if (mo) m[mo] = (m[mo] || 0) + (f.total_cost || 0)
  }
  return Object.entries(m).map(([month, total]) => ({ month, total })).sort((a, b) => a.month.localeCompare(b.month)).slice(-12)
})

const maxMonthly = computed(() => Math.max(1, ...monthlyData.value.map((m: any) => m.total)))

const topVehicles = computed(() => {
  const m: Record<string, { vehicle_plate: string; vehicle_id: string; total_cost: number }> = {}
  for (const e of expenses.value) {
    if (!m[e.vehicle_id]) m[e.vehicle_id] = { vehicle_plate: e.vehicle_plate || e.vehicle_id, vehicle_id: e.vehicle_id, total_cost: 0 }
    m[e.vehicle_id].total_cost += e.amount || 0
  }
  for (const f of fuelLogs.value) {
    if (!m[f.vehicle_id]) m[f.vehicle_id] = { vehicle_plate: f.vehicle_plate || f.vehicle_id, vehicle_id: f.vehicle_id, total_cost: 0 }
    m[f.vehicle_id].total_cost += f.total_cost || 0
  }
  return Object.values(m).sort((a, b) => b.total_cost - a.total_cost).slice(0, 8)
})

const driverStatusData = computed(() => {
  const m: Record<string, number> = {}
  for (const d of drivers.value) { m[d.status] = (m[d.status] || 0) + 1 }
  const dots: Record<string, string> = { available: 'bg-green-400', on_duty: 'bg-blue-400', off_duty: 'bg-gray-400', on_leave: 'bg-amber-400', terminated: 'bg-red-400' }
  return Object.entries(m).map(([status, count]) => ({ status, count, dot: dots[status] || 'bg-gray-400' }))
})

function formatEnum(s: string) { return s ? s.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()) : '—' }
function fmtDZD(n: number) { return new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n || 0) + ' DZD' }
function fmtK(n: number) { return n >= 1000000 ? (n / 1000000).toFixed(1) + 'M' : n >= 1000 ? (n / 1000).toFixed(0) + 'K' : String(Math.round(n)) }

async function load() {
  loading.value = true
  try {
    const [rDash, rVeh, rDrv, rExp, rFuel, rAsgn] = await Promise.allSettled([
      fleetAPI.getDashboard(),
      fleetAPI.listVehicles(),
      fleetAPI.listDrivers(),
      fleetAPI.listExpenses(),
      fleetAPI.listFuelLogs(),
      fleetAPI.listAssignments(),
    ])
    if (rDash.status === 'fulfilled') report.value = rDash.value.data || {}
    if (rVeh.status === 'fulfilled') vehicles.value = rVeh.value.data.items || rVeh.value.data || []
    if (rDrv.status === 'fulfilled') drivers.value = rDrv.value.data.items || rDrv.value.data || []
    if (rExp.status === 'fulfilled') expenses.value = rExp.value.data.items || rExp.value.data || []
    if (rFuel.status === 'fulfilled') fuelLogs.value = rFuel.value.data.items || rFuel.value.data || []
    if (rAsgn.status === 'fulfilled') assignments.value = rAsgn.value.data.items || rAsgn.value.data || []
  } catch { app.addToast('Failed to load fleet reports', 'error') }
  finally { loading.value = false }
}

onMounted(load)
</script>
