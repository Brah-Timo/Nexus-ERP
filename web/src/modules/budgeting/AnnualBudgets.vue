<template>
  <div class="p-6" :class="app.darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <CalendarDays class="w-7 h-7 text-blue-500" />
          Annual Budgets
        </h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Create and manage annual budget plans</p>
      </div>
      <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-600 text-white hover:bg-blue-700 text-sm font-medium">
        <Plus class="w-4 h-4" /> New Budget
      </button>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-4">
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'" />
        <input v-model="search" placeholder="Search..."
          class="pl-9 pr-3 py-2 rounded-lg border text-sm w-56"
          :class="inputClass" />
      </div>
      <select v-model="filterYear" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option value="">All Years</option>
        <option v-for="y in yearOptions" :key="y" :value="y">FY {{ y }}</option>
      </select>
      <select v-model="filterStatus" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option value="">All Statuses</option>
        <option value="draft">Draft</option>
        <option value="active">Active</option>
        <option value="locked">Locked</option>
        <option value="closed">Closed</option>
        <option value="cancelled">Cancelled</option>
      </select>
    </div>

    <!-- Table -->
    <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'bg-gray-750 text-gray-400' : 'bg-gray-50 text-gray-500'">
              <th class="px-4 py-3 text-left font-medium">Budget #</th>
              <th class="px-4 py-3 text-left font-medium">Name</th>
              <th class="px-4 py-3 text-center font-medium">FY</th>
              <th class="px-4 py-3 text-center font-medium">Type</th>
              <th class="px-4 py-3 text-right font-medium">Total Amount</th>
              <th class="px-4 py-3 text-center font-medium">Period</th>
              <th class="px-4 py-3 text-center font-medium">Status</th>
              <th class="px-4 py-3 text-center font-medium">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
            <tr v-for="b in filtered" :key="b.id"
              class="hover:bg-opacity-50 transition-colors"
              :class="app.darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-50'">
              <td class="px-4 py-3 font-mono font-medium text-blue-500">{{ b.budget_number }}</td>
              <td class="px-4 py-3 font-medium max-w-xs truncate">{{ b.name }}</td>
              <td class="px-4 py-3 text-center">{{ b.fiscal_year }}</td>
              <td class="px-4 py-3 text-center">
                <span class="px-2 py-0.5 rounded-full text-xs capitalize" :class="typeClass(b.budget_type as string)">{{ b.budget_type }}</span>
              </td>
              <td class="px-4 py-3 text-right font-semibold">{{ fmt(b.total_amount as number) }}</td>
              <td class="px-4 py-3 text-center text-xs" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">
                {{ b.start_date }} → {{ b.end_date }}
              </td>
              <td class="px-4 py-3 text-center">
                <span class="px-2 py-0.5 rounded-full text-xs font-medium capitalize" :class="statusClass(b.status as string)">{{ b.status }}</span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center justify-center gap-1">
                  <button @click="openDetail(b)" class="p-1.5 rounded hover:bg-blue-100 dark:hover:bg-blue-900 text-blue-500" title="View Details">
                    <Eye class="w-4 h-4" />
                  </button>
                  <button v-if="b.status === 'draft'" @click="doApprove(b)" class="p-1.5 rounded hover:bg-green-100 dark:hover:bg-green-900 text-green-500" title="Approve">
                    <CheckCircle2 class="w-4 h-4" />
                  </button>
                  <button v-if="b.status === 'active'" @click="doLock(b)" class="p-1.5 rounded hover:bg-amber-100 dark:hover:bg-amber-900 text-amber-500" title="Lock">
                    <Lock class="w-4 h-4" />
                  </button>
                  <button @click="openEdit(b)" class="p-1.5 rounded hover:bg-indigo-100 dark:hover:bg-indigo-900 text-indigo-500" title="Edit">
                    <Pencil class="w-4 h-4" />
                  </button>
                  <button @click="doDelete(b)" class="p-1.5 rounded hover:bg-red-100 dark:hover:bg-red-900 text-red-500" title="Delete">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!filtered.length">
              <td colspan="8" class="px-4 py-12 text-center" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">No budgets found</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" @click.self="showModal = false">
        <div class="w-full max-w-2xl rounded-2xl shadow-2xl p-6 max-h-[90vh] overflow-y-auto" :class="app.darkMode ? 'bg-gray-800 text-white' : 'bg-white text-gray-900'">
          <h2 class="text-lg font-bold mb-5 flex items-center gap-2">
            <CalendarDays class="w-5 h-5 text-blue-500" />
            {{ editing ? 'Edit' : 'New' }} Annual Budget
          </h2>
          <div class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Fiscal Year *</label>
                <input v-model.number="form.fiscal_year" type="number" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Budget Type</label>
                <select v-model="form.budget_type" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass">
                  <option value="operational">Operational</option>
                  <option value="capital">Capital</option>
                  <option value="project">Project</option>
                  <option value="department">Department</option>
                </select>
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Name *</label>
              <input v-model="form.name" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Start Date *</label>
                <input v-model="form.start_date" type="date" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">End Date *</label>
                <input v-model="form.end_date" type="date" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Total Amount</label>
              <input v-model.number="form.total_amount" type="number" step="0.01" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Status</label>
              <select v-model="form.status" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass">
                <option value="draft">Draft</option>
                <option value="active">Active</option>
                <option value="locked">Locked</option>
                <option value="closed">Closed</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Description</label>
              <textarea v-model="form.description" rows="2" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass"></textarea>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Notes</label>
              <textarea v-model="form.notes" rows="2" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass"></textarea>
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showModal = false" class="px-4 py-2 rounded-lg text-sm border" :class="app.darkMode ? 'border-gray-600 text-gray-300 hover:bg-gray-700' : 'border-gray-300 text-gray-600 hover:bg-gray-50'">Cancel</button>
            <button @click="save" :disabled="saving" class="px-5 py-2 rounded-lg text-sm bg-blue-600 text-white hover:bg-blue-700 font-medium disabled:opacity-60">
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Detail Modal -->
    <Teleport to="body">
      <div v-if="detailBudget" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" @click.self="detailBudget = null">
        <div class="w-full max-w-4xl rounded-2xl shadow-2xl p-6 max-h-[90vh] overflow-y-auto" :class="app.darkMode ? 'bg-gray-800 text-white' : 'bg-white text-gray-900'">
          <div class="flex items-center justify-between mb-5">
            <h2 class="text-lg font-bold flex items-center gap-2">
              <CalendarDays class="w-5 h-5 text-blue-500" />
              {{ detailBudget.budget_number }} — {{ detailBudget.name }}
            </h2>
            <button @click="detailBudget = null" class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700">
              <X class="w-5 h-5" />
            </button>
          </div>

          <div class="grid grid-cols-3 gap-4 mb-6">
            <div class="rounded-lg p-3 text-center" :class="app.darkMode ? 'bg-gray-700' : 'bg-gray-50'">
              <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Total Amount</p>
              <p class="text-lg font-bold text-blue-500">{{ fmt(detailBudget.total_amount as number) }}</p>
            </div>
            <div class="rounded-lg p-3 text-center" :class="app.darkMode ? 'bg-gray-700' : 'bg-gray-50'">
              <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Status</p>
              <span class="px-2 py-0.5 rounded-full text-sm capitalize font-medium" :class="statusClass(detailBudget.status as string)">{{ detailBudget.status }}</span>
            </div>
            <div class="rounded-lg p-3 text-center" :class="app.darkMode ? 'bg-gray-700' : 'bg-gray-50'">
              <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Period</p>
              <p class="text-sm font-medium">{{ detailBudget.start_date }} → {{ detailBudget.end_date }}</p>
            </div>
          </div>

          <h3 class="font-semibold mb-3 flex items-center gap-2">
            <List class="w-4 h-4 text-blue-500" />
            Line Items
          </h3>
          <div class="overflow-x-auto rounded-lg border" :class="app.darkMode ? 'border-gray-700' : 'border-gray-200'">
            <table class="w-full text-sm">
              <thead>
                <tr :class="app.darkMode ? 'bg-gray-700 text-gray-400' : 'bg-gray-50 text-gray-500'">
                  <th class="px-3 py-2 text-left font-medium">Category</th>
                  <th class="px-3 py-2 text-left font-medium">Account</th>
                  <th class="px-3 py-2 text-right font-medium">Budget</th>
                  <th class="px-3 py-2 text-right font-medium">Actual</th>
                  <th class="px-3 py-2 text-right font-medium">Variance</th>
                </tr>
              </thead>
              <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
                <tr v-for="li in detailBudget.line_items as Record<string, unknown>[]" :key="li.id as string">
                  <td class="px-3 py-2">{{ li.category_name }}</td>
                  <td class="px-3 py-2 text-xs" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ li.account_code }} {{ li.account_name }}</td>
                  <td class="px-3 py-2 text-right">{{ fmt(li.budget_amount as number) }}</td>
                  <td class="px-3 py-2 text-right">{{ fmt(li.actual_amount as number) }}</td>
                  <td class="px-3 py-2 text-right" :class="(li.budget_amount as number) >= (li.actual_amount as number) ? 'text-green-500' : 'text-red-500'">
                    {{ fmt((li.budget_amount as number) - (li.actual_amount as number)) }}
                  </td>
                </tr>
                <tr v-if="!(detailBudget.line_items as unknown[])?.length">
                  <td colspan="5" class="px-3 py-6 text-center" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">No line items</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { CalendarDays, Plus, Search, Eye, Pencil, Trash2, CheckCircle2, Lock, X, List } from 'lucide-vue-next'
import { budgetingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const budgets = ref<Record<string, unknown>[]>([])
const search = ref('')
const filterYear = ref('')
const filterStatus = ref('')
const showModal = ref(false)
const saving = ref(false)
const editing = ref(false)
const detailBudget = ref<Record<string, unknown> | null>(null)
const currentYear = new Date().getFullYear()
const yearOptions = Array.from({ length: 5 }, (_, i) => currentYear - 2 + i)

const blankForm = () => ({
  id: '', budget_number: '', fiscal_year: currentYear,
  name: '', description: '', budget_type: 'operational',
  status: 'draft', start_date: `${currentYear}-01-01`,
  end_date: `${currentYear}-12-31`, total_amount: 0, notes: ''
})
const form = ref(blankForm())

const fmt = (v: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(v || 0)

const filtered = computed(() => budgets.value.filter(b => {
  if (search.value && !String(b.name).toLowerCase().includes(search.value.toLowerCase()) &&
      !String(b.budget_number).toLowerCase().includes(search.value.toLowerCase())) return false
  if (filterYear.value && b.fiscal_year !== Number(filterYear.value)) return false
  if (filterStatus.value && b.status !== filterStatus.value) return false
  return true
}))

const inputClass = computed(() => app.darkMode
  ? 'bg-gray-700 border-gray-600 text-white placeholder-gray-500 focus:border-blue-500 focus:outline-none'
  : 'bg-white border-gray-300 text-gray-900 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none')

const statusClass = (s: string) => ({
  draft:     'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
  active:    'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  locked:    'bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300',
  closed:    'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  cancelled: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
}[s] || 'bg-gray-100 text-gray-700')

const typeClass = (t: string) => ({
  operational: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  capital:     'bg-violet-100 text-violet-700 dark:bg-violet-900 dark:text-violet-300',
  project:     'bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300',
  department:  'bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300',
}[t] || 'bg-gray-100 text-gray-700')

function openCreate() { form.value = blankForm(); editing.value = false; showModal.value = true }
function openEdit(b: Record<string, unknown>) { form.value = { ...blankForm(), ...b } as typeof form.value; editing.value = true; showModal.value = true }

async function openDetail(b: Record<string, unknown>) {
  const r = await budgetingAPI.getAnnualBudget(b.id as string)
  detailBudget.value = r.data
}

async function save() {
  saving.value = true
  try {
    if (editing.value) {
      await budgetingAPI.updateAnnualBudget(form.value.id, form.value)
    } else {
      await budgetingAPI.createAnnualBudget(form.value)
    }
    showModal.value = false
    await load()
  } catch (e) { console.error(e) } finally { saving.value = false }
}

async function doApprove(b: Record<string, unknown>) {
  if (!confirm(`Approve budget "${b.budget_number}"?`)) return
  await budgetingAPI.approveBudget(b.id as string)
  await load()
}

async function doLock(b: Record<string, unknown>) {
  if (!confirm(`Lock budget "${b.budget_number}"?`)) return
  await budgetingAPI.lockBudget(b.id as string)
  await load()
}

async function doDelete(b: Record<string, unknown>) {
  if (!confirm(`Delete budget "${b.budget_number}"?`)) return
  await budgetingAPI.deleteAnnualBudget(b.id as string)
  await load()
}

async function load() {
  const params: Record<string, string> = {}
  if (filterYear.value) params.fiscal_year = filterYear.value
  if (filterStatus.value) params.status = filterStatus.value
  const r = await budgetingAPI.listAnnualBudgets(params)
  budgets.value = r.data.data || []
}

onMounted(load)
</script>
