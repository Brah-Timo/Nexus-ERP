<template>
  <div :class="dk('min-h-screen bg-gray-950 text-gray-100','min-h-screen bg-gray-50 text-gray-900')">
    <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border-b px-6 py-4'">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-lg bg-rose-600/20">
            <Receipt class="w-5 h-5 text-rose-400" />
          </div>
          <div>
            <h1 class="text-lg font-semibold">Vehicle Expenses</h1>
            <p :class="dk('text-gray-400','text-gray-500')" class="text-xs">Track all fleet-related expenses</p>
          </div>
        </div>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 bg-rose-600 hover:bg-rose-700 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" />
          Add Expense
        </button>
      </div>
    </div>

    <div class="p-6 space-y-6">
      <!-- KPI Cards -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div v-for="kpi in kpis" :key="kpi.label" :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-4'">
          <div class="flex items-center justify-between mb-2">
            <span :class="dk('text-gray-400','text-gray-500')" class="text-xs">{{ kpi.label }}</span>
            <component :is="kpi.icon" :class="'w-4 h-4 ' + kpi.color" />
          </div>
          <div class="text-xl font-bold">{{ kpi.value }}</div>
        </div>
      </div>

      <!-- Expense Type Breakdown -->
      <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-5'">
        <h3 class="text-sm font-semibold mb-4">Expense Breakdown by Type</h3>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
          <div v-for="(item, type) in expenseByType" :key="type" :class="dk('bg-gray-800/60','bg-gray-50')+ ' rounded-lg p-3'">
            <div :class="dk('text-gray-400','text-gray-500')" class="text-xs mb-1">{{ formatEnum(String(type)) }}</div>
            <div class="font-bold text-sm">{{ fmtDZD(item.total) }}</div>
            <div :class="dk('text-gray-500','text-gray-400')" class="text-xs">{{ item.count }} records</div>
            <div class="mt-2 h-1 rounded-full bg-gray-700">
              <div class="h-1 rounded-full bg-rose-500" :style="{ width: Math.min(100, (item.total / maxTypeTotal) * 100) + '%' }" />
            </div>
          </div>
        </div>
      </div>

      <!-- Filters -->
      <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl p-4 flex flex-wrap gap-3 items-center'">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input v-model="search" placeholder="Search vehicle, vendor, description…"
            :class="dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500','bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400')+ ' border rounded-lg pl-9 pr-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-rose-500'" />
        </div>
        <select v-model="typeFilter" :class="dk('bg-gray-800 border-gray-700 text-gray-100','bg-gray-50 border-gray-300 text-gray-900')+ ' border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-rose-500'">
          <option value="">All Types</option>
          <option v-for="t in expenseTypes" :key="t" :value="t">{{ formatEnum(t) }}</option>
        </select>
        <div class="flex gap-2">
          <input v-model="dateFrom" type="date" :class="dk('bg-gray-800 border-gray-700 text-gray-100','bg-gray-50 border-gray-300 text-gray-900')+ ' border rounded-lg px-3 py-2 text-sm focus:outline-none'" />
          <input v-model="dateTo" type="date" :class="dk('bg-gray-800 border-gray-700 text-gray-100','bg-gray-50 border-gray-300 text-gray-900')+ ' border rounded-lg px-3 py-2 text-sm focus:outline-none'" />
        </div>
      </div>

      <!-- Table -->
      <div :class="dk('bg-gray-900 border-gray-800','bg-white border-gray-200')+ ' border rounded-xl overflow-hidden'">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead :class="dk('bg-gray-800/50 text-gray-400','bg-gray-50 text-gray-500')">
              <tr>
                <th class="px-4 py-3 text-left font-medium">Date</th>
                <th class="px-4 py-3 text-left font-medium">Vehicle</th>
                <th class="px-4 py-3 text-left font-medium">Type</th>
                <th class="px-4 py-3 text-left font-medium">Description</th>
                <th class="px-4 py-3 text-left font-medium">Vendor</th>
                <th class="px-4 py-3 text-right font-medium">Amount</th>
                <th class="px-4 py-3 text-left font-medium">Receipt</th>
                <th class="px-4 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody :class="dk('divide-gray-800','divide-gray-100')+ ' divide-y'">
              <tr v-if="loading">
                <td colspan="8" class="py-12 text-center text-gray-400">
                  <Loader2 class="w-6 h-6 animate-spin mx-auto mb-2" />Loading…
                </td>
              </tr>
              <tr v-else-if="filtered.length === 0">
                <td colspan="8" class="py-12 text-center text-gray-400">
                  <Receipt class="w-8 h-8 mx-auto mb-2 opacity-30" />No expenses found
                </td>
              </tr>
              <tr v-for="e in filtered" :key="e.id" :class="dk('hover:bg-gray-800/50','hover:bg-gray-50')" class="transition-colors">
                <td class="px-4 py-3 text-xs">{{ e.expense_date }}</td>
                <td class="px-4 py-3 font-medium text-sm">{{ e.vehicle_plate || e.vehicle_id }}</td>
                <td class="px-4 py-3">
                  <span :class="typeColor(e.expense_type)" class="px-2 py-0.5 rounded text-xs">{{ formatEnum(e.expense_type) }}</span>
                </td>
                <td class="px-4 py-3 text-xs max-w-40 truncate">{{ e.description || '—' }}</td>
                <td class="px-4 py-3 text-xs">{{ e.vendor_name || '—' }}</td>
                <td class="px-4 py-3 text-right font-semibold text-rose-400">{{ fmtDZD(e.amount) }}</td>
                <td class="px-4 py-3 text-xs">{{ e.receipt_number || '—' }}</td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-end gap-1">
                    <button @click="openEdit(e)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors">
                      <Pencil class="w-4 h-4 text-amber-400" />
                    </button>
                    <button @click="openDelete(e)" :class="dk('hover:bg-gray-700','hover:bg-gray-100')" class="p-1.5 rounded transition-colors">
                      <Trash2 class="w-4 h-4 text-red-400" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
            <tfoot v-if="filtered.length > 0" :class="dk('bg-gray-800/30 text-gray-300','bg-gray-50 text-gray-700')">
              <tr>
                <td colspan="5" class="px-4 py-3 font-medium text-sm">Total ({{ filtered.length }} expenses)</td>
                <td class="px-4 py-3 text-right font-bold text-rose-400">{{ fmtDZD(filteredTotal) }}</td>
                <td colspan="2" />
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showForm" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showForm = false" />
        <div :class="dk('bg-gray-900 border-gray-700','bg-white border-gray-200')+ ' relative border rounded-2xl w-full max-w-xl max-h-[90vh] overflow-y-auto shadow-2xl'">
          <div :class="dk('border-gray-800','border-gray-100')+ ' flex items-center justify-between p-6 border-b'">
            <h2 class="text-lg font-semibold">{{ editing ? 'Edit Expense' : 'Add Expense' }}</h2>
            <button @click="showForm = false" :class="dk('hover:bg-gray-800','hover:bg-gray-100')" class="p-2 rounded-lg transition-colors"><X class="w-5 h-5" /></button>
          </div>
          <div class="p-6 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Vehicle ID <span class="text-red-400">*</span></label>
              <input v-model="form.vehicle_id" :class="inputCls" placeholder="Vehicle UUID" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Expense Date <span class="text-red-400">*</span></label>
              <input v-model="form.expense_date" type="date" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Expense Type <span class="text-red-400">*</span></label>
              <select v-model="form.expense_type" :class="inputCls">
                <option v-for="t in expenseTypes" :key="t" :value="t">{{ formatEnum(t) }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Amount (DZD) <span class="text-red-400">*</span></label>
              <input v-model.number="form.amount" type="number" step="0.01" :class="inputCls" placeholder="0.00" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium mb-1 text-gray-400">Description</label>
              <input v-model="form.description" :class="inputCls" placeholder="Brief description" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Vendor Name</label>
              <input v-model="form.vendor_name" :class="inputCls" placeholder="Supplier / vendor" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Receipt Number</label>
              <input v-model="form.receipt_number" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 text-gray-400">Approved By</label>
              <input v-model="form.approved_by" :class="inputCls" />
            </div>
            <div class="md:col-span-2">
              <label class="block text-xs font-medium mb-1 text-gray-400">Notes</label>
              <textarea v-model="form.notes" :class="inputCls" rows="2" />
            </div>
          </div>
          <div :class="dk('border-gray-800 bg-gray-900/50','border-gray-100 bg-gray-50')+ ' flex justify-end gap-3 p-6 border-t'">
            <button @click="showForm = false" :class="dk('bg-gray-800 hover:bg-gray-700 text-gray-300','bg-gray-100 hover:bg-gray-200 text-gray-700')" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="save" :disabled="saving" class="px-4 py-2 bg-rose-600 hover:bg-rose-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              {{ editing ? 'Save Changes' : 'Add Expense' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Delete Modal -->
    <Teleport to="body">
      <div v-if="showDelete && deleteItem" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="showDelete = false" />
        <div :class="dk('bg-gray-900 border-gray-700','bg-white border-gray-200')+ ' relative border rounded-2xl w-full max-w-md shadow-2xl p-6'">
          <div class="flex items-center gap-3 mb-4">
            <div class="p-2 rounded-full bg-red-500/20"><AlertTriangle class="w-5 h-5 text-red-400" /></div>
            <h2 class="text-lg font-semibold">Delete Expense</h2>
          </div>
          <p :class="dk('text-gray-400','text-gray-600')" class="text-sm mb-6">Delete expense <strong>{{ fmtDZD(deleteItem.amount) }}</strong> on {{ deleteItem.expense_date }}?</p>
          <div class="flex justify-end gap-3">
            <button @click="showDelete = false" :class="dk('bg-gray-800 hover:bg-gray-700 text-gray-300','bg-gray-100 hover:bg-gray-200 text-gray-700')" class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="confirmDelete" :disabled="saving" class="px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center gap-2">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />Delete
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Receipt, Plus, Search, Pencil, Trash2, X, Loader2, AlertTriangle, DollarSign, TrendingDown, LayoutList, Car } from 'lucide-vue-next'
import { fleetAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const dk = (dark: string, light: string) => app.darkMode ? dark : light

const expenses = ref<any[]>([])
const loading = ref(false)
const saving = ref(false)
const search = ref('')
const typeFilter = ref('')
const dateFrom = ref('')
const dateTo = ref('')
const showForm = ref(false)
const showDelete = ref(false)
const editing = ref<any>(null)
const deleteItem = ref<any>(null)

const expenseTypes = ['fuel','maintenance','insurance','registration','toll','parking','repair','cleaning','other']

const defaultForm = () => ({
  vehicle_id: '', expense_date: new Date().toISOString().split('T')[0],
  expense_type: 'fuel', amount: null, description: '',
  vendor_name: '', receipt_number: '', approved_by: '', notes: '',
})
const form = ref(defaultForm())

const inputCls = computed(() =>
  dk('bg-gray-800 border-gray-700 text-gray-100 placeholder-gray-500', 'bg-gray-50 border-gray-300 text-gray-900 placeholder-gray-400') +
  ' border rounded-lg px-3 py-2 text-sm w-full focus:outline-none focus:ring-2 focus:ring-rose-500')

const filtered = computed(() => {
  let list = expenses.value
  if (typeFilter.value) list = list.filter(e => e.expense_type === typeFilter.value)
  if (dateFrom.value) list = list.filter(e => e.expense_date >= dateFrom.value)
  if (dateTo.value) list = list.filter(e => e.expense_date <= dateTo.value)
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(e => [e.vehicle_plate, e.vendor_name, e.description, e.receipt_number].some(f => f?.toLowerCase().includes(q)))
  }
  return list
})

const filteredTotal = computed(() => filtered.value.reduce((s, e) => s + (e.amount || 0), 0))

const expenseByType = computed(() => {
  const result: Record<string, { total: number; count: number }> = {}
  for (const e of expenses.value) {
    if (!result[e.expense_type]) result[e.expense_type] = { total: 0, count: 0 }
    result[e.expense_type].total += e.amount || 0
    result[e.expense_type].count++
  }
  return result
})

const maxTypeTotal = computed(() => Math.max(1, ...Object.values(expenseByType.value).map(v => v.total)))

const kpis = computed(() => {
  const total = expenses.value.reduce((s, e) => s + (e.amount || 0), 0)
  const thisMonth = expenses.value.filter(e => {
    const d = new Date(e.expense_date)
    const n = new Date()
    return d.getMonth() === n.getMonth() && d.getFullYear() === n.getFullYear()
  }).reduce((s, e) => s + (e.amount || 0), 0)
  return [
    { label: 'Total Expenses', value: fmtDZD(total), icon: DollarSign, color: 'text-rose-400' },
    { label: 'This Month', value: fmtDZD(thisMonth), icon: TrendingDown, color: 'text-amber-400' },
    { label: 'Records', value: expenses.value.length, icon: LayoutList, color: 'text-blue-400' },
    { label: 'Vehicles', value: new Set(expenses.value.map(e => e.vehicle_id)).size, icon: Car, color: 'text-green-400' },
  ]
})

function formatEnum(s: string) { return s ? s.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase()) : '—' }
function fmtDZD(n: number) { return new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n || 0) + ' DZD' }
function typeColor(t: string) {
  const m: Record<string, string> = {
    fuel: 'bg-yellow-500/15 text-yellow-400',
    maintenance: 'bg-orange-500/15 text-orange-400',
    insurance: 'bg-blue-500/15 text-blue-400',
    registration: 'bg-purple-500/15 text-purple-400',
    repair: 'bg-red-500/15 text-red-400',
    toll: 'bg-green-500/15 text-green-400',
    parking: 'bg-indigo-500/15 text-indigo-400',
  }
  return m[t] || 'bg-gray-500/15 text-gray-400'
}

async function load() {
  loading.value = true
  try {
    const r = await fleetAPI.listExpenses()
    expenses.value = r.data.items || r.data || []
  } catch { app.addToast('Failed to load expenses', 'error') }
  finally { loading.value = false }
}

function openCreate() { editing.value = null; form.value = defaultForm(); showForm.value = true }
function openEdit(e: any) { editing.value = e; form.value = { ...defaultForm(), ...e }; showForm.value = true }
function openDelete(e: any) { deleteItem.value = e; showDelete.value = true }

async function save() {
  if (!form.value.vehicle_id || !form.value.amount) { app.addToast('Vehicle and amount are required', 'error'); return }
  saving.value = true
  try {
    if (editing.value) {
      await fleetAPI.updateExpense(editing.value.id, form.value)
      app.addToast('Expense updated', 'success')
    } else {
      await fleetAPI.createExpense(form.value)
      app.addToast('Expense added', 'success')
    }
    showForm.value = false; await load()
  } catch (e: any) { app.addToast(e?.response?.data?.error || 'Save failed', 'error') }
  finally { saving.value = false }
}

async function confirmDelete() {
  if (!deleteItem.value) return
  saving.value = true
  try {
    await fleetAPI.deleteExpense(deleteItem.value.id)
    app.addToast('Expense deleted', 'success')
    showDelete.value = false; await load()
  } catch { app.addToast('Delete failed', 'error') }
  finally { saving.value = false }
}

onMounted(load)
</script>
