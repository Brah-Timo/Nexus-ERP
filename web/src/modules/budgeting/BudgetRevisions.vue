<template>
  <div class="p-6" :class="app.darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <FilePen class="w-7 h-7 text-amber-500" />
          Budget Revisions
        </h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Track and approve budget adjustments and reallocations</p>
      </div>
      <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 rounded-lg bg-amber-600 text-white hover:bg-amber-700 text-sm font-medium">
        <Plus class="w-4 h-4" /> New Revision
      </button>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-4">
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'" />
        <input v-model="search" placeholder="Search revisions..." class="pl-9 pr-3 py-2 rounded-lg border text-sm w-56" :class="inputClass" />
      </div>
      <select v-model="filterBudget" @change="load" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option value="">All Budgets</option>
        <option v-for="b in annualBudgets" :key="b.id as string" :value="b.id">{{ b.budget_number }} — {{ b.name }}</option>
      </select>
      <select v-model="filterStatus" @change="load" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option value="">All Statuses</option>
        <option value="draft">Draft</option>
        <option value="active">Approved</option>
        <option value="cancelled">Cancelled</option>
      </select>
    </div>

    <!-- Table -->
    <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'bg-gray-750 text-gray-400' : 'bg-gray-50 text-gray-500'">
              <th class="px-4 py-3 text-left font-medium">Revision #</th>
              <th class="px-4 py-3 text-left font-medium">Budget</th>
              <th class="px-4 py-3 text-center font-medium">Type</th>
              <th class="px-4 py-3 text-right font-medium">Original</th>
              <th class="px-4 py-3 text-right font-medium">Revised</th>
              <th class="px-4 py-3 text-right font-medium">Change</th>
              <th class="px-4 py-3 text-left font-medium">Reason</th>
              <th class="px-4 py-3 text-center font-medium">Eff. Date</th>
              <th class="px-4 py-3 text-center font-medium">Status</th>
              <th class="px-4 py-3 text-center font-medium">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
            <tr v-for="r in filtered" :key="r.id as string"
              class="transition-colors" :class="app.darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-50'">
              <td class="px-4 py-3 font-mono font-medium text-amber-500">{{ r.revision_number }}</td>
              <td class="px-4 py-3 text-xs" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">
                <p class="font-medium text-sm">{{ r.budget_number }}</p>
                <p>{{ r.budget_name }}</p>
              </td>
              <td class="px-4 py-3 text-center">
                <span class="px-2 py-0.5 rounded-full text-xs capitalize" :class="revTypeClass(r.revision_type as string)">{{ r.revision_type }}</span>
              </td>
              <td class="px-4 py-3 text-right">{{ fmt(r.original_amount as number) }}</td>
              <td class="px-4 py-3 text-right font-semibold">{{ fmt(r.revised_amount as number) }}</td>
              <td class="px-4 py-3 text-right font-semibold" :class="(r.change_amount as number) >= 0 ? 'text-green-500' : 'text-red-500'">
                {{ (r.change_amount as number) >= 0 ? '+' : '' }}{{ fmt(r.change_amount as number) }}
              </td>
              <td class="px-4 py-3 max-w-xs truncate" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ r.reason }}</td>
              <td class="px-4 py-3 text-center text-xs">{{ r.effective_date || '—' }}</td>
              <td class="px-4 py-3 text-center">
                <span class="px-2 py-0.5 rounded-full text-xs font-medium capitalize" :class="statusClass(r.status as string)">
                  {{ r.status === 'active' ? 'Approved' : r.status }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center justify-center gap-1">
                  <button v-if="r.status === 'draft'" @click="doApprove(r)" class="p-1.5 rounded hover:bg-green-100 dark:hover:bg-green-900 text-green-500" title="Approve">
                    <CheckCircle2 class="w-4 h-4" />
                  </button>
                  <button v-if="r.status === 'draft'" @click="openEdit(r)" class="p-1.5 rounded hover:bg-indigo-100 dark:hover:bg-indigo-900 text-indigo-500" title="Edit">
                    <Pencil class="w-4 h-4" />
                  </button>
                  <button @click="doDelete(r)" class="p-1.5 rounded hover:bg-red-100 dark:hover:bg-red-900 text-red-500" title="Delete">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!filtered.length">
              <td colspan="10" class="px-4 py-12 text-center" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">No revisions found</td>
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
            <FilePen class="w-5 h-5 text-amber-500" />
            {{ editing ? 'Edit' : 'New' }} Budget Revision
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
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Revision Type</label>
                <select v-model="form.revision_type" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass">
                  <option value="increase">Increase</option>
                  <option value="decrease">Decrease</option>
                  <option value="reallocation">Reallocation</option>
                  <option value="correction">Correction</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Effective Date</label>
                <input v-model="form.effective_date" type="date" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Original Amount</label>
                <input v-model.number="form.original_amount" type="number" step="0.01" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Revised Amount</label>
                <input v-model.number="form.revised_amount" type="number" step="0.01" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
            </div>
            <div class="rounded-lg p-3" :class="app.darkMode ? 'bg-gray-700' : 'bg-gray-50'">
              <p class="text-xs" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Change Amount</p>
              <p class="text-lg font-bold" :class="(form.revised_amount - form.original_amount) >= 0 ? 'text-green-500' : 'text-red-500'">
                {{ (form.revised_amount - form.original_amount) >= 0 ? '+' : '' }}{{ fmt(form.revised_amount - form.original_amount) }}
              </p>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Reason *</label>
              <textarea v-model="form.reason" rows="3" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" placeholder="Explain the reason for this revision..."></textarea>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Status</label>
              <select v-model="form.status" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass">
                <option value="draft">Draft</option>
                <option value="active">Approved</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Notes</label>
              <textarea v-model="form.notes" rows="2" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass"></textarea>
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showModal = false" class="px-4 py-2 rounded-lg text-sm border" :class="app.darkMode ? 'border-gray-600 text-gray-300 hover:bg-gray-700' : 'border-gray-300 text-gray-600 hover:bg-gray-50'">Cancel</button>
            <button @click="save" :disabled="saving" class="px-5 py-2 rounded-lg text-sm bg-amber-600 text-white hover:bg-amber-700 font-medium disabled:opacity-60">
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
import { FilePen, Plus, Search, Pencil, Trash2, CheckCircle2 } from 'lucide-vue-next'
import { budgetingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const revisions = ref<Record<string, unknown>[]>([])
const annualBudgets = ref<Record<string, unknown>[]>([])
const search = ref('')
const filterBudget = ref('')
const filterStatus = ref('')
const showModal = ref(false)
const saving = ref(false)
const editing = ref(false)

const blankForm = () => ({ id: '', annual_budget_id: '', revision_type: 'increase', original_amount: 0, revised_amount: 0, reason: '', status: 'draft', effective_date: '', notes: '' })
const form = ref(blankForm())

const fmt = (v: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(v || 0)

const filtered = computed(() => {
  let list = revisions.value
  if (search.value) list = list.filter(r => String(r.revision_number).includes(search.value) || String(r.reason).toLowerCase().includes(search.value.toLowerCase()))
  return list
})

const inputClass = computed(() => app.darkMode
  ? 'bg-gray-700 border-gray-600 text-white placeholder-gray-500 focus:border-amber-500 focus:outline-none'
  : 'bg-white border-gray-300 text-gray-900 focus:border-amber-500 focus:ring-1 focus:ring-amber-500 focus:outline-none')

const statusClass = (s: string) => ({
  draft:     'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
  active:    'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  cancelled: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
}[s] || 'bg-gray-100 text-gray-700')

const revTypeClass = (t: string) => ({
  increase:     'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  decrease:     'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
  reallocation: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  correction:   'bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300',
}[t] || 'bg-gray-100 text-gray-700')

function openCreate() { form.value = blankForm(); editing.value = false; showModal.value = true }
function openEdit(r: Record<string, unknown>) { form.value = { ...blankForm(), ...r } as typeof form.value; editing.value = true; showModal.value = true }

async function save() {
  saving.value = true
  try {
    if (editing.value) await budgetingAPI.updateRevision(form.value.id, form.value)
    else await budgetingAPI.createRevision(form.value)
    showModal.value = false
    await load()
  } catch (e) { console.error(e) } finally { saving.value = false }
}

async function doApprove(r: Record<string, unknown>) {
  if (!confirm(`Approve revision ${r.revision_number}? This will update the budget line item.`)) return
  await budgetingAPI.approveRevision(r.id as string)
  await load()
}

async function doDelete(r: Record<string, unknown>) {
  if (!confirm(`Delete revision ${r.revision_number}?`)) return
  await budgetingAPI.deleteRevision(r.id as string)
  await load()
}

async function load() {
  const params: Record<string, string> = {}
  if (filterBudget.value) params.annual_budget_id = filterBudget.value
  if (filterStatus.value) params.status = filterStatus.value
  const [revRes, budgetRes] = await Promise.all([budgetingAPI.listRevisions(params), budgetingAPI.listAnnualBudgets()])
  revisions.value = revRes.data.data || []
  annualBudgets.value = budgetRes.data.data || []
}

onMounted(load)
</script>
