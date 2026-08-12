<template>
  <div class="p-6" :class="app.darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <Building2 class="w-7 h-7 text-violet-500" />
          Department Budgets
        </h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Budget allocations per department</p>
      </div>
      <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 rounded-lg bg-violet-600 text-white hover:bg-violet-700 text-sm font-medium">
        <Plus class="w-4 h-4" /> Allocate Department
      </button>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-4">
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'" />
        <input v-model="search" placeholder="Search departments..." class="pl-9 pr-3 py-2 rounded-lg border text-sm w-56" :class="inputClass" />
      </div>
      <select v-model="filterBudget" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option value="">All Budgets</option>
        <option v-for="b in annualBudgets" :key="b.id as string" :value="b.id">{{ b.budget_number }} — {{ b.name }}</option>
      </select>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-5">
      <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Total Allocated</p>
        <p class="text-xl font-bold text-violet-500">{{ fmt(totals.allocated) }}</p>
      </div>
      <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Total Spent</p>
        <p class="text-xl font-bold text-blue-500">{{ fmt(totals.spent) }}</p>
      </div>
      <div class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Available</p>
        <p class="text-xl font-bold text-emerald-500">{{ fmt(totals.available) }}</p>
      </div>
    </div>

    <!-- Table -->
    <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'bg-gray-750 text-gray-400' : 'bg-gray-50 text-gray-500'">
              <th class="px-4 py-3 text-left font-medium">Department</th>
              <th class="px-4 py-3 text-left font-medium">Budget</th>
              <th class="px-4 py-3 text-center font-medium">FY</th>
              <th class="px-4 py-3 text-right font-medium">Allocated</th>
              <th class="px-4 py-3 text-right font-medium">Spent</th>
              <th class="px-4 py-3 text-right font-medium">Committed</th>
              <th class="px-4 py-3 text-right font-medium">Available</th>
              <th class="px-4 py-3 text-center font-medium">Utilisation</th>
              <th class="px-4 py-3 text-center font-medium">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
            <tr v-for="d in filtered" :key="d.id as string"
              class="transition-colors" :class="app.darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-50'">
              <td class="px-4 py-3">
                <p class="font-medium">{{ d.department_name }}</p>
                <p v-if="d.department_code" class="text-xs" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ d.department_code }}</p>
              </td>
              <td class="px-4 py-3 text-xs" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ d.budget_number }}</td>
              <td class="px-4 py-3 text-center">{{ d.fiscal_year }}</td>
              <td class="px-4 py-3 text-right font-semibold">{{ fmt(d.allocated_amount as number) }}</td>
              <td class="px-4 py-3 text-right">{{ fmt(d.spent_amount as number) }}</td>
              <td class="px-4 py-3 text-right text-amber-500">{{ fmt(d.committed_amount as number) }}</td>
              <td class="px-4 py-3 text-right" :class="(d.available_amount as number) >= 0 ? 'text-emerald-500' : 'text-red-500'">
                {{ fmt(d.available_amount as number) }}
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <div class="flex-1 rounded-full h-2 overflow-hidden" :class="app.darkMode ? 'bg-gray-700' : 'bg-gray-200'">
                    <div class="h-2 rounded-full transition-all"
                      :class="(d.utilization_pct as number) > 90 ? 'bg-red-500' : (d.utilization_pct as number) > 75 ? 'bg-amber-500' : 'bg-emerald-500'"
                      :style="{ width: Math.min(d.utilization_pct as number, 100) + '%' }">
                    </div>
                  </div>
                  <span class="text-xs w-10 text-right">{{ (d.utilization_pct as number).toFixed(0) }}%</span>
                </div>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center justify-center gap-1">
                  <button @click="openEdit(d)" class="p-1.5 rounded hover:bg-indigo-100 dark:hover:bg-indigo-900 text-indigo-500">
                    <Pencil class="w-4 h-4" />
                  </button>
                  <button @click="doDelete(d)" class="p-1.5 rounded hover:bg-red-100 dark:hover:bg-red-900 text-red-500">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!filtered.length">
              <td colspan="9" class="px-4 py-12 text-center" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">No department budgets found</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" @click.self="showModal = false">
        <div class="w-full max-w-lg rounded-2xl shadow-2xl p-6" :class="app.darkMode ? 'bg-gray-800 text-white' : 'bg-white text-gray-900'">
          <h2 class="text-lg font-bold mb-5 flex items-center gap-2">
            <Building2 class="w-5 h-5 text-violet-500" />
            {{ editing ? 'Edit' : 'New' }} Department Budget
          </h2>
          <div class="space-y-4">
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Annual Budget *</label>
              <select v-model="form.annual_budget_id" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass">
                <option value="">Select budget...</option>
                <option v-for="b in annualBudgets" :key="b.id as string" :value="b.id">{{ b.budget_number }} — {{ b.name }}</option>
              </select>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Department Name *</label>
                <input v-model="form.department_name" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Department Code</label>
                <input v-model="form.department_code" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Allocated Amount</label>
              <input v-model.number="form.allocated_amount" type="number" step="0.01" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Notes</label>
              <textarea v-model="form.notes" rows="2" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass"></textarea>
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showModal = false" class="px-4 py-2 rounded-lg text-sm border" :class="app.darkMode ? 'border-gray-600 text-gray-300 hover:bg-gray-700' : 'border-gray-300 text-gray-600 hover:bg-gray-50'">Cancel</button>
            <button @click="save" :disabled="saving" class="px-5 py-2 rounded-lg text-sm bg-violet-600 text-white hover:bg-violet-700 font-medium disabled:opacity-60">
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Building2, Plus, Search, Pencil, Trash2 } from 'lucide-vue-next'
import { budgetingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const deptBudgets = ref<Record<string, unknown>[]>([])
const annualBudgets = ref<Record<string, unknown>[]>([])
const search = ref('')
const filterBudget = ref('')
const showModal = ref(false)
const saving = ref(false)
const editing = ref(false)

const blankForm = () => ({ id: '', annual_budget_id: '', department_name: '', department_code: '', allocated_amount: 0, notes: '' })
const form = ref(blankForm())

const fmt = (v: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(v || 0)

const filtered = computed(() => {
  let list = deptBudgets.value
  if (search.value) list = list.filter(d => String(d.department_name).toLowerCase().includes(search.value.toLowerCase()))
  if (filterBudget.value) list = list.filter(d => d.annual_budget_id === filterBudget.value)
  return list
})

const totals = computed(() => ({
  allocated: filtered.value.reduce((s, d) => s + (d.allocated_amount as number), 0),
  spent: filtered.value.reduce((s, d) => s + (d.spent_amount as number), 0),
  available: filtered.value.reduce((s, d) => s + (d.available_amount as number), 0),
}))

const inputClass = computed(() => app.darkMode
  ? 'bg-gray-700 border-gray-600 text-white placeholder-gray-500 focus:border-violet-500 focus:outline-none'
  : 'bg-white border-gray-300 text-gray-900 focus:border-violet-500 focus:ring-1 focus:ring-violet-500 focus:outline-none')

function openCreate() { form.value = blankForm(); editing.value = false; showModal.value = true }
function openEdit(d: Record<string, unknown>) { form.value = { ...blankForm(), ...d } as typeof form.value; editing.value = true; showModal.value = true }

async function save() {
  saving.value = true
  try {
    if (editing.value) await budgetingAPI.updateDepartmentBudget(form.value.id, form.value)
    else await budgetingAPI.createDepartmentBudget(form.value)
    showModal.value = false
    await load()
  } catch (e) { console.error(e) } finally { saving.value = false }
}

async function doDelete(d: Record<string, unknown>) {
  if (!confirm(`Delete department budget for "${d.department_name}"?`)) return
  await budgetingAPI.deleteDepartmentBudget(d.id as string)
  await load()
}

async function load() {
  const [deptRes, budgetRes] = await Promise.all([
    budgetingAPI.listDepartmentBudgets(filterBudget.value ? { annual_budget_id: filterBudget.value } : {}),
    budgetingAPI.listAnnualBudgets()
  ])
  deptBudgets.value = deptRes.data.data || []
  annualBudgets.value = budgetRes.data.data || []
}

onMounted(load)
</script>
