<template>
  <div class="p-6" :class="app.darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <Handshake class="w-7 h-7 text-orange-500" />
          Commitments
        </h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">Manage budget commitments, encumbrances and pre-commitments</p>
      </div>
      <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 rounded-lg bg-orange-600 text-white hover:bg-orange-700 text-sm font-medium">
        <Plus class="w-4 h-4" /> New Commitment
      </button>
    </div>

    <!-- Filters -->
    <div class="flex flex-wrap gap-3 mb-4">
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'" />
        <input v-model="search" placeholder="Search commitments..." class="pl-9 pr-3 py-2 rounded-lg border text-sm w-56" :class="inputClass" />
      </div>
      <select v-model="filterBudget" @change="load" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option value="">All Budgets</option>
        <option v-for="b in annualBudgets" :key="b.id as string" :value="b.id">{{ b.budget_number }} — {{ b.name }}</option>
      </select>
      <select v-model="filterStatus" @change="load" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option value="">All Statuses</option>
        <option value="pending">Pending</option>
        <option value="approved">Approved</option>
        <option value="fulfilled">Fulfilled</option>
        <option value="cancelled">Cancelled</option>
      </select>
      <select v-model="filterType" @change="load" class="rounded-lg border px-3 py-2 text-sm" :class="inputClass">
        <option value="">All Types</option>
        <option value="purchase_order">Purchase Order</option>
        <option value="contract">Contract</option>
        <option value="reservation">Reservation</option>
        <option value="other">Other</option>
      </select>
    </div>

    <!-- Summary -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-5">
      <div v-for="s in summaryCards" :key="s.label" class="rounded-xl p-4 border" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
        <p class="text-xs uppercase tracking-wide mb-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ s.label }}</p>
        <p class="text-xl font-bold" :class="s.color">{{ s.value }}</p>
      </div>
    </div>

    <!-- Table -->
    <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'bg-gray-750 text-gray-400' : 'bg-gray-50 text-gray-500'">
              <th class="px-4 py-3 text-left font-medium">Commitment #</th>
              <th class="px-4 py-3 text-left font-medium">Vendor / Description</th>
              <th class="px-4 py-3 text-center font-medium">Type</th>
              <th class="px-4 py-3 text-right font-medium">Committed</th>
              <th class="px-4 py-3 text-right font-medium">Fulfilled</th>
              <th class="px-4 py-3 text-right font-medium">Remaining</th>
              <th class="px-4 py-3 text-center font-medium">Date</th>
              <th class="px-4 py-3 text-center font-medium">Expected</th>
              <th class="px-4 py-3 text-center font-medium">Status</th>
              <th class="px-4 py-3 text-center font-medium">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
            <tr v-for="c in filtered" :key="c.id as string"
              class="transition-colors" :class="app.darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-50'">
              <td class="px-4 py-3 font-mono font-medium text-orange-500">{{ c.commitment_number }}</td>
              <td class="px-4 py-3">
                <p class="font-medium truncate max-w-xs">{{ c.vendor_name || c.description }}</p>
                <p v-if="c.vendor_name" class="text-xs truncate max-w-xs" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ c.description }}</p>
                <p v-if="c.reference_number" class="text-xs" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">Ref: {{ c.reference_number }}</p>
              </td>
              <td class="px-4 py-3 text-center">
                <span class="px-2 py-0.5 rounded-full text-xs capitalize" :class="typeClass(c.commitment_type as string)">
                  {{ String(c.commitment_type).replace('_', ' ') }}
                </span>
              </td>
              <td class="px-4 py-3 text-right font-semibold">{{ fmt(c.committed_amount as number) }}</td>
              <td class="px-4 py-3 text-right text-emerald-500">{{ fmt(c.fulfilled_amount as number) }}</td>
              <td class="px-4 py-3 text-right" :class="(c.remaining_amount as number) > 0 ? 'text-amber-500' : 'text-gray-400'">
                {{ fmt(c.remaining_amount as number) }}
              </td>
              <td class="px-4 py-3 text-center text-xs">{{ c.commitment_date }}</td>
              <td class="px-4 py-3 text-center text-xs" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">{{ c.expected_fulfillment || '—' }}</td>
              <td class="px-4 py-3 text-center">
                <span class="px-2 py-0.5 rounded-full text-xs font-medium capitalize" :class="statusClass(c.status as string)">{{ c.status }}</span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center justify-center gap-1">
                  <button v-if="c.status === 'pending'" @click="doApprove(c)" class="p-1.5 rounded hover:bg-green-100 dark:hover:bg-green-900 text-green-500" title="Approve">
                    <CheckCircle2 class="w-4 h-4" />
                  </button>
                  <button v-if="c.status === 'approved'" @click="openFulfill(c)" class="p-1.5 rounded hover:bg-blue-100 dark:hover:bg-blue-900 text-blue-500" title="Fulfill">
                    <PackageCheck class="w-4 h-4" />
                  </button>
                  <button v-if="['pending','approved'].includes(c.status as string)" @click="doCancel(c)" class="p-1.5 rounded hover:bg-amber-100 dark:hover:bg-amber-900 text-amber-500" title="Cancel">
                    <Ban class="w-4 h-4" />
                  </button>
                  <button @click="openEdit(c)" class="p-1.5 rounded hover:bg-indigo-100 dark:hover:bg-indigo-900 text-indigo-500" title="Edit">
                    <Pencil class="w-4 h-4" />
                  </button>
                  <button @click="doDelete(c)" class="p-1.5 rounded hover:bg-red-100 dark:hover:bg-red-900 text-red-500" title="Delete">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!filtered.length">
              <td colspan="10" class="px-4 py-12 text-center" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">No commitments found</td>
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
            <Handshake class="w-5 h-5 text-orange-500" />
            {{ editing ? 'Edit' : 'New' }} Commitment
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
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Type</label>
                <select v-model="form.commitment_type" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass">
                  <option value="purchase_order">Purchase Order</option>
                  <option value="contract">Contract</option>
                  <option value="reservation">Reservation</option>
                  <option value="other">Other</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Reference Number</label>
                <input v-model="form.reference_number" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Vendor Name</label>
              <input v-model="form.vendor_name" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Description *</label>
              <textarea v-model="form.description" rows="2" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass"></textarea>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Committed Amount *</label>
                <input v-model.number="form.committed_amount" type="number" step="0.01" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Status</label>
                <select v-model="form.status" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass">
                  <option value="pending">Pending</option>
                  <option value="approved">Approved</option>
                  <option value="fulfilled">Fulfilled</option>
                  <option value="cancelled">Cancelled</option>
                </select>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Commitment Date *</label>
                <input v-model="form.commitment_date" type="date" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Expected Fulfillment</label>
                <input v-model="form.expected_fulfillment" type="date" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Notes</label>
              <textarea v-model="form.notes" rows="2" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass"></textarea>
            </div>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showModal = false" class="px-4 py-2 rounded-lg text-sm border" :class="app.darkMode ? 'border-gray-600 text-gray-300 hover:bg-gray-700' : 'border-gray-300 text-gray-600 hover:bg-gray-50'">Cancel</button>
            <button @click="save" :disabled="saving" class="px-5 py-2 rounded-lg text-sm bg-orange-600 text-white hover:bg-orange-700 font-medium disabled:opacity-60">
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Fulfill Modal -->
    <Teleport to="body">
      <div v-if="fulfillTarget" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" @click.self="fulfillTarget = null">
        <div class="w-full max-w-md rounded-2xl shadow-2xl p-6" :class="app.darkMode ? 'bg-gray-800 text-white' : 'bg-white text-gray-900'">
          <h2 class="text-lg font-bold mb-4 flex items-center gap-2">
            <PackageCheck class="w-5 h-5 text-blue-500" />
            Fulfill Commitment
          </h2>
          <p class="text-sm mb-4" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">
            Commitment: <span class="font-mono font-medium">{{ fulfillTarget.commitment_number }}</span><br/>
            Committed: <span class="font-semibold">{{ fmt(fulfillTarget.committed_amount as number) }}</span><br/>
            Already Fulfilled: <span class="font-semibold text-emerald-500">{{ fmt(fulfillTarget.fulfilled_amount as number) }}</span>
          </p>
          <div>
            <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Fulfilled Amount</label>
            <input v-model.number="fulfillAmount" type="number" step="0.01" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
          </div>
          <div class="flex justify-end gap-3 mt-5">
            <button @click="fulfillTarget = null" class="px-4 py-2 rounded-lg text-sm border" :class="app.darkMode ? 'border-gray-600 text-gray-300 hover:bg-gray-700' : 'border-gray-300 text-gray-600 hover:bg-gray-50'">Cancel</button>
            <button @click="doFulfill" class="px-5 py-2 rounded-lg text-sm bg-blue-600 text-white hover:bg-blue-700 font-medium">Confirm Fulfillment</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Handshake, Plus, Search, Pencil, Trash2, CheckCircle2, PackageCheck, Ban } from 'lucide-vue-next'
import { budgetingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const commitments = ref<Record<string, unknown>[]>([])
const annualBudgets = ref<Record<string, unknown>[]>([])
const search = ref('')
const filterBudget = ref('')
const filterStatus = ref('')
const filterType = ref('')
const showModal = ref(false)
const saving = ref(false)
const editing = ref(false)
const fulfillTarget = ref<Record<string, unknown> | null>(null)
const fulfillAmount = ref(0)

const today = new Date().toISOString().split('T')[0]
const blankForm = () => ({ id: '', annual_budget_id: '', commitment_type: 'purchase_order', status: 'pending', reference_number: '', vendor_name: '', description: '', committed_amount: 0, commitment_date: today, expected_fulfillment: '', notes: '' })
const form = ref(blankForm())

const fmt = (v: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(v || 0)

const filtered = computed(() => {
  let list = commitments.value
  if (search.value) list = list.filter(c => String(c.commitment_number).includes(search.value) || String(c.vendor_name).toLowerCase().includes(search.value.toLowerCase()) || String(c.description).toLowerCase().includes(search.value.toLowerCase()))
  return list
})

const summaryCards = computed(() => {
  const list = filtered.value
  return [
    { label: 'Total Committed', value: fmt(list.reduce((s, c) => s + (c.committed_amount as number), 0)), color: 'text-orange-500' },
    { label: 'Fulfilled', value: fmt(list.reduce((s, c) => s + (c.fulfilled_amount as number), 0)), color: 'text-emerald-500' },
    { label: 'Remaining', value: fmt(list.reduce((s, c) => s + (c.remaining_amount as number), 0)), color: 'text-amber-500' },
    { label: 'Open Count', value: list.filter(c => ['pending', 'approved'].includes(c.status as string)).length, color: 'text-blue-500' },
  ]
})

const inputClass = computed(() => app.darkMode
  ? 'bg-gray-700 border-gray-600 text-white placeholder-gray-500 focus:border-orange-500 focus:outline-none'
  : 'bg-white border-gray-300 text-gray-900 focus:border-orange-500 focus:ring-1 focus:ring-orange-500 focus:outline-none')

const statusClass = (s: string) => ({
  pending:   'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300',
  approved:  'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  fulfilled: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  cancelled: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
}[s] || 'bg-gray-100 text-gray-700')

const typeClass = (t: string) => ({
  purchase_order: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  contract:       'bg-violet-100 text-violet-700 dark:bg-violet-900 dark:text-violet-300',
  reservation:    'bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300',
  other:          'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
}[t] || 'bg-gray-100 text-gray-700')

function openCreate() { form.value = blankForm(); editing.value = false; showModal.value = true }
function openEdit(c: Record<string, unknown>) { form.value = { ...blankForm(), ...c } as typeof form.value; editing.value = true; showModal.value = true }
function openFulfill(c: Record<string, unknown>) { fulfillTarget.value = c; fulfillAmount.value = c.remaining_amount as number }

async function save() {
  saving.value = true
  try {
    if (editing.value) await budgetingAPI.updateCommitment(form.value.id, form.value)
    else await budgetingAPI.createCommitment(form.value)
    showModal.value = false
    await load()
  } catch (e) { console.error(e) } finally { saving.value = false }
}

async function doApprove(c: Record<string, unknown>) {
  if (!confirm(`Approve commitment ${c.commitment_number}?`)) return
  await budgetingAPI.approveCommitment(c.id as string)
  await load()
}

async function doFulfill() {
  if (!fulfillTarget.value) return
  await budgetingAPI.fulfillCommitment(fulfillTarget.value.id as string, { fulfilled_amount: fulfillAmount.value })
  fulfillTarget.value = null
  await load()
}

async function doCancel(c: Record<string, unknown>) {
  if (!confirm(`Cancel commitment ${c.commitment_number}?`)) return
  await budgetingAPI.cancelCommitment(c.id as string)
  await load()
}

async function doDelete(c: Record<string, unknown>) {
  if (!confirm(`Delete commitment ${c.commitment_number}?`)) return
  await budgetingAPI.deleteCommitment(c.id as string)
  await load()
}

async function load() {
  const params: Record<string, string> = {}
  if (filterBudget.value) params.annual_budget_id = filterBudget.value
  if (filterStatus.value) params.status = filterStatus.value
  if (filterType.value) params.commitment_type = filterType.value
  const [commRes, budgetRes] = await Promise.all([budgetingAPI.listCommitments(params), budgetingAPI.listAnnualBudgets()])
  commitments.value = commRes.data.data || []
  annualBudgets.value = budgetRes.data.data || []
}

onMounted(load)
</script>
