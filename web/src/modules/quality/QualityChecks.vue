<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { qualityAPI } from '@/api/client'
import {
  ClipboardCheck, Plus, Search, RefreshCw,
  ChevronLeft, ChevronRight, Loader2, X, Check,
  AlertCircle, CheckCircle2, Minus, HelpCircle, Eye
} from 'lucide-vue-next'

const app = useAppStore()
const dk = (d: string, l: string) => app.darkMode ? d : l

// ─── state ────────────────────────────────────────────────────────────────────
const loading     = ref(false)
const checks      = ref<any[]>([])
const total       = ref(0)
const page        = ref(1)
const perPage     = ref(25)

const filterResult       = ref('')
const filterInspectionId = ref('')
const search             = ref('')

// Create check modal
const showCreate  = ref(false)
const saving      = ref(false)
const createForm  = ref({
  inspection_id: '',
  sequence: 1,
  name: '',
  description: '',
  check_type: 'visual',
  unit: '',
  min_value: '',
  max_value: '',
  norm_reference: '',
  instructions: '',
  is_mandatory: true,
})

// Record result modal
const showRecord  = ref(false)
const recordTarget = ref<any>(null)
const recordForm  = ref({
  result: 'pass',
  measured_value: '',
  notes: '',
})

// ─── computed ─────────────────────────────────────────────────────────────────
const totalPages = computed(() => Math.ceil(total.value / perPage.value))

const resultOptions = [
  { value: '', label: 'All Results' },
  { value: 'pass', label: 'Pass' },
  { value: 'fail', label: 'Fail' },
  { value: 'na', label: 'N/A' },
  { value: 'observation', label: 'Observation' },
]

// ─── load ─────────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const res = await qualityAPI.listChecks({
      page: page.value,
      per_page: perPage.value,
      result: filterResult.value || undefined,
      inspection_id: filterInspectionId.value || undefined,
    })
    checks.value = res.data.checks || []
    total.value  = res.data.total || 0
  } catch {}
  loading.value = false
}

onMounted(load)

function applyFilter() { page.value = 1; load() }
function resetFilter() {
  filterResult.value = ''; filterInspectionId.value = ''; search.value = ''
  page.value = 1; load()
}
function prevPage() { if (page.value > 1) { page.value--; load() } }
function nextPage() { if (page.value < totalPages.value) { page.value++; load() } }

// ─── create ───────────────────────────────────────────────────────────────────
function openCreate() {
  createForm.value = {
    inspection_id: '', sequence: 1, name: '', description: '',
    check_type: 'visual', unit: '', min_value: '', max_value: '',
    norm_reference: '', instructions: '', is_mandatory: true,
  }
  showCreate.value = true
}

async function submitCreate() {
  if (!createForm.value.inspection_id || !createForm.value.name) {
    alert('Inspection ID and Check Name are required.')
    return
  }
  saving.value = true
  try {
    await qualityAPI.createCheck(createForm.value)
    showCreate.value = false
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error creating check')
  }
  saving.value = false
}

// ─── record result ────────────────────────────────────────────────────────────
function openRecord(chk: any) {
  recordTarget.value = chk
  recordForm.value = {
    result: chk.result || 'pass',
    measured_value: chk.measured_value?.toString() || '',
    notes: chk.notes || '',
  }
  showRecord.value = true
}

async function submitRecord() {
  saving.value = true
  try {
    await qualityAPI.recordCheckResult(recordTarget.value.id, recordForm.value)
    showRecord.value = false
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error recording result')
  }
  saving.value = false
}

// ─── delete ───────────────────────────────────────────────────────────────────
async function deleteCheck(id: string) {
  if (!confirm('Delete this check?')) return
  try {
    await qualityAPI.deleteCheck(id)
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error')
  }
}

// ─── helpers ──────────────────────────────────────────────────────────────────
function resultBadge(r: string) {
  const m: Record<string, string> = {
    pass: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300',
    fail: 'bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300',
    na: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400',
    observation: 'bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300',
  }
  return m[r] || 'bg-slate-100 text-slate-600'
}

function resultIcon(r: string) {
  if (r === 'pass') return CheckCircle2
  if (r === 'fail') return AlertCircle
  if (r === 'na')   return Minus
  return HelpCircle
}

function checkTypeLabel(t: string) {
  const m: Record<string, string> = {
    visual: 'Visual', measurement: 'Measurement',
    functional: 'Functional', document: 'Document', count: 'Count',
  }
  return m[t] || t
}

function checkTypeColor(t: string) {
  const m: Record<string, string> = {
    visual: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
    measurement: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
    functional: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300',
    document: 'bg-teal-100 text-teal-700 dark:bg-teal-900/30 dark:text-teal-300',
    count: 'bg-pink-100 text-pink-700 dark:bg-pink-900/30 dark:text-pink-300',
  }
  return m[t] || 'bg-slate-100 text-slate-600'
}

function fmtDateTime(d: any) {
  if (!d) return '—'
  return new Date(d).toLocaleString()
}
</script>

<template>
  <div :class="dk('bg-slate-900 text-slate-100','bg-slate-50 text-slate-900')" class="min-h-screen">
    <div class="max-w-screen-2xl mx-auto px-6 py-6">

      <!-- Header -->
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-purple-600 flex items-center justify-center">
            <ClipboardCheck class="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 class="text-2xl font-bold">Quality Checks</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">
              {{ total }} check{{ total !== 1 ? 's' : '' }} total
            </p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load"
            :class="dk('text-slate-400 hover:text-white hover:bg-slate-700','text-slate-500 hover:text-slate-700 hover:bg-slate-200')"
            class="p-2 rounded-lg transition-colors">
            <RefreshCw :class="{'animate-spin': loading}" class="w-5 h-5" />
          </button>
          <button @click="openCreate"
            class="flex items-center gap-2 px-4 py-2 rounded-lg bg-purple-600 hover:bg-purple-700 text-white font-medium text-sm transition-colors">
            <Plus class="w-4 h-4" />
            Add Check
          </button>
        </div>
      </div>

      <!-- Summary Cards -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6">
        <div v-for="item in [
          { label: 'Total Checks', value: total, color: 'indigo' },
          { label: 'Passed', value: checks.filter(c => c.result === 'pass').length, color: 'emerald' },
          { label: 'Failed', value: checks.filter(c => c.result === 'fail').length, color: 'red' },
          { label: 'Pending', value: checks.filter(c => !c.result).length, color: 'amber' },
        ]" :key="item.label"
          :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="rounded-xl border p-4">
          <p :class="dk('text-slate-400','text-slate-500')" class="text-xs font-medium mb-1">{{ item.label }}</p>
          <p class="text-2xl font-bold" :class="`text-${item.color}-500`">{{ item.value }}</p>
        </div>
      </div>

      <!-- Filters -->
      <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
        class="rounded-xl border p-4 mb-5 flex flex-wrap items-center gap-3">
        <div class="relative flex-1 min-w-[220px]">
          <Search :class="dk('text-slate-400','text-slate-400')" class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" />
          <input v-model="filterInspectionId" @keyup.enter="applyFilter"
            :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-slate-50 border-slate-300 text-slate-900 placeholder-slate-400')"
            class="w-full pl-9 pr-3 py-2 text-sm border rounded-lg focus:ring-2 focus:ring-purple-500 outline-none"
            placeholder="Filter by Inspection ID..." />
        </div>
        <select v-model="filterResult" @change="applyFilter"
          :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
          class="px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500">
          <option v-for="o in resultOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <button @click="resetFilter"
          :class="dk('text-slate-400 hover:text-white border-slate-600','text-slate-500 hover:text-slate-700 border-slate-300')"
          class="px-3 py-2 text-sm border rounded-lg transition-colors">
          Reset
        </button>
      </div>

      <!-- Table -->
      <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
        class="rounded-xl border overflow-hidden">
        <div v-if="loading" class="flex items-center justify-center py-20">
          <Loader2 class="w-8 h-8 text-purple-500 animate-spin" />
        </div>
        <div v-else-if="checks.length === 0" class="flex flex-col items-center justify-center py-20 gap-3">
          <ClipboardCheck :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12" />
          <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">No quality checks found</p>
        </div>
        <template v-else>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr :class="dk('bg-slate-900/60 text-slate-400 border-b border-slate-700','bg-slate-50 text-slate-500 border-b border-slate-200')">
                  <th class="px-4 py-3 text-left font-medium">#</th>
                  <th class="px-4 py-3 text-left font-medium">Inspection</th>
                  <th class="px-4 py-3 text-left font-medium">Check Name</th>
                  <th class="px-4 py-3 text-left font-medium">Type</th>
                  <th class="px-4 py-3 text-left font-medium">Spec (Min – Max)</th>
                  <th class="px-4 py-3 text-left font-medium">Measured</th>
                  <th class="px-4 py-3 text-left font-medium">Result</th>
                  <th class="px-4 py-3 text-left font-medium">Mandatory</th>
                  <th class="px-4 py-3 text-left font-medium">Checked By</th>
                  <th class="px-4 py-3 text-left font-medium">Checked At</th>
                  <th class="px-4 py-3 text-center font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="chk in checks" :key="chk.id"
                  :class="dk('border-b border-slate-700/60 hover:bg-slate-700/30','border-b border-slate-100 hover:bg-slate-50')"
                  class="transition-colors">
                  <td class="px-4 py-3" :class="dk('text-slate-400','text-slate-500')">{{ chk.sequence }}</td>
                  <td class="px-4 py-3">
                    <div class="font-mono text-xs" :class="dk('text-indigo-300','text-indigo-600')">
                      {{ chk.inspection_ref }}
                    </div>
                  </td>
                  <td class="px-4 py-3 font-medium" :class="dk('text-slate-100','text-slate-800')">
                    {{ chk.name }}
                    <div v-if="chk.norm_reference" class="text-xs font-normal mt-0.5"
                      :class="dk('text-slate-400','text-slate-500')">
                      Ref: {{ chk.norm_reference }}
                    </div>
                  </td>
                  <td class="px-4 py-3">
                    <span :class="checkTypeColor(chk.check_type)"
                      class="px-2 py-0.5 rounded-full text-xs font-medium">
                      {{ checkTypeLabel(chk.check_type) }}
                    </span>
                  </td>
                  <td class="px-4 py-3 font-mono text-xs" :class="dk('text-slate-300','text-slate-600')">
                    <span v-if="chk.min_value !== null || chk.max_value !== null">
                      {{ chk.min_value ?? '—' }} – {{ chk.max_value ?? '—' }}
                      <span v-if="chk.unit" class="ml-1" :class="dk('text-slate-500','text-slate-400')">{{ chk.unit }}</span>
                    </span>
                    <span v-else :class="dk('text-slate-600','text-slate-400')">—</span>
                  </td>
                  <td class="px-4 py-3">
                    <span v-if="chk.measured_value !== null && chk.measured_value !== undefined"
                      class="font-mono text-sm font-semibold"
                      :class="chk.result === 'fail' ? 'text-red-500' : chk.result === 'pass' ? 'text-emerald-500' : dk('text-slate-200','text-slate-700')">
                      {{ chk.measured_value }}
                      <span v-if="chk.unit" class="text-xs font-normal ml-0.5" :class="dk('text-slate-400','text-slate-500')">
                        {{ chk.unit }}
                      </span>
                    </span>
                    <span v-else :class="dk('text-slate-500','text-slate-400')" class="text-xs">—</span>
                  </td>
                  <td class="px-4 py-3">
                    <div v-if="chk.result" class="flex items-center gap-1.5">
                      <component :is="resultIcon(chk.result)"
                        :class="resultBadge(chk.result).includes('emerald') ? 'text-emerald-500' : resultBadge(chk.result).includes('red') ? 'text-red-500' : 'text-slate-500'"
                        class="w-4 h-4" />
                      <span :class="resultBadge(chk.result)"
                        class="px-2 py-0.5 rounded-full text-xs font-semibold uppercase">
                        {{ chk.result }}
                      </span>
                    </div>
                    <span v-else
                      :class="dk('bg-slate-700 text-slate-400','bg-slate-100 text-slate-500')"
                      class="px-2 py-0.5 rounded-full text-xs">
                      Pending
                    </span>
                  </td>
                  <td class="px-4 py-3">
                    <span v-if="chk.is_mandatory"
                      class="px-2 py-0.5 rounded-full text-xs bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300 font-medium">
                      Yes
                    </span>
                    <span v-else
                      :class="dk('text-slate-500','text-slate-400')" class="text-xs">No</span>
                  </td>
                  <td class="px-4 py-3 text-xs" :class="dk('text-slate-300','text-slate-600')">
                    {{ chk.checked_by || '—' }}
                  </td>
                  <td class="px-4 py-3 text-xs" :class="dk('text-slate-400','text-slate-500')">
                    {{ fmtDateTime(chk.checked_at) }}
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex items-center justify-center gap-1">
                      <button @click="openRecord(chk)" title="Record Result"
                        class="p-1.5 rounded-lg bg-purple-600/20 hover:bg-purple-600/40 text-purple-400 hover:text-purple-300 transition-colors">
                        <Check class="w-3.5 h-3.5" />
                      </button>
                      <button @click="deleteCheck(chk.id)" title="Delete"
                        class="p-1.5 rounded-lg bg-red-600/20 hover:bg-red-600/40 text-red-400 hover:text-red-300 transition-colors">
                        <X class="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <!-- Pagination -->
          <div :class="dk('border-t border-slate-700 bg-slate-800/50','border-t border-slate-200 bg-slate-50')"
            class="flex items-center justify-between px-4 py-3">
            <span class="text-xs" :class="dk('text-slate-400','text-slate-500')">
              {{ (page - 1) * perPage + 1 }}–{{ Math.min(page * perPage, total) }} of {{ total }}
            </span>
            <div class="flex items-center gap-1">
              <button @click="prevPage" :disabled="page <= 1"
                :class="page <= 1 ? 'opacity-40 cursor-not-allowed' : dk('hover:bg-slate-700','hover:bg-slate-200')"
                class="p-1.5 rounded-lg transition-colors">
                <ChevronLeft :class="dk('text-slate-300','text-slate-600')" class="w-4 h-4" />
              </button>
              <span class="px-3 text-sm" :class="dk('text-slate-300','text-slate-600')">
                {{ page }} / {{ totalPages || 1 }}
              </span>
              <button @click="nextPage" :disabled="page >= totalPages"
                :class="page >= totalPages ? 'opacity-40 cursor-not-allowed' : dk('hover:bg-slate-700','hover:bg-slate-200')"
                class="p-1.5 rounded-lg transition-colors">
                <ChevronRight :class="dk('text-slate-300','text-slate-600')" class="w-4 h-4" />
              </button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- ─── Create Check Modal ──────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showCreate"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="relative w-full max-w-2xl rounded-2xl border shadow-2xl max-h-[90vh] overflow-y-auto">
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-between px-6 py-4 border-b">
            <h2 class="text-lg font-semibold" :class="dk('text-white','text-slate-800')">Add Quality Check</h2>
            <button @click="showCreate = false"
              :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-700')">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5 grid grid-cols-2 gap-4">
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">
                Inspection ID <span class="text-red-400">*</span>
              </label>
              <input v-model="createForm.inspection_id"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500"
                placeholder="Inspection UUID" />
            </div>
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">
                Check Name <span class="text-red-400">*</span>
              </label>
              <input v-model="createForm.name"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500"
                placeholder="e.g. Visual Surface Inspection" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Check Type</label>
              <select v-model="createForm.check_type"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500">
                <option value="visual">Visual</option>
                <option value="measurement">Measurement</option>
                <option value="functional">Functional</option>
                <option value="document">Document</option>
                <option value="count">Count</option>
              </select>
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Sequence</label>
              <input v-model.number="createForm.sequence" type="number" min="1"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Min Value</label>
              <input v-model="createForm.min_value" type="number"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Max Value</label>
              <input v-model="createForm.max_value" type="number"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Unit</label>
              <input v-model="createForm.unit"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500"
                placeholder="mm, kg, ..." />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Norm Reference</label>
              <input v-model="createForm.norm_reference"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500"
                placeholder="ISO 9001, ..." />
            </div>
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Instructions</label>
              <textarea v-model="createForm.instructions" rows="2"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500 resize-none"
                placeholder="Step-by-step instructions..." />
            </div>
            <div class="col-span-2 flex items-center gap-2">
              <input id="mandatory" v-model="createForm.is_mandatory" type="checkbox"
                class="w-4 h-4 rounded border-slate-500 bg-slate-700 text-purple-600 focus:ring-purple-500" />
              <label for="mandatory" :class="dk('text-slate-300','text-slate-600')" class="text-sm">
                Mandatory check
              </label>
            </div>
          </div>
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showCreate = false"
              :class="dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
              class="px-4 py-2 text-sm border rounded-lg transition-colors">Cancel</button>
            <button @click="submitCreate" :disabled="saving"
              class="flex items-center gap-2 px-4 py-2 text-sm bg-purple-600 hover:bg-purple-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              Add Check
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ─── Record Result Modal ────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showRecord"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="relative w-full max-w-md rounded-2xl border shadow-2xl">
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-between px-6 py-4 border-b">
            <h2 class="text-lg font-semibold" :class="dk('text-white','text-slate-800')">Record Check Result</h2>
            <button @click="showRecord = false"
              :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-700')">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5 space-y-4">
            <p class="text-sm" :class="dk('text-slate-300','text-slate-600')">
              Check: <span class="font-semibold">{{ recordTarget?.name }}</span>
            </p>
            <!-- Spec info -->
            <div v-if="recordTarget?.min_value !== null || recordTarget?.max_value !== null"
              :class="dk('bg-slate-700/50 border-slate-600','bg-slate-50 border-slate-200')"
              class="rounded-lg border px-4 py-3 text-xs">
              <span :class="dk('text-slate-400','text-slate-500')">Acceptable range: </span>
              <span class="font-mono font-semibold" :class="dk('text-slate-200','text-slate-700')">
                {{ recordTarget?.min_value ?? '—' }} – {{ recordTarget?.max_value ?? '—' }}
                <span v-if="recordTarget?.unit" class="font-normal ml-1">{{ recordTarget?.unit }}</span>
              </span>
            </div>
            <!-- Result selection -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-2">
                Result <span class="text-red-400">*</span>
              </label>
              <div class="grid grid-cols-2 gap-2">
                <button v-for="r in ['pass','fail','na','observation']" :key="r"
                  @click="recordForm.result = r"
                  :class="recordForm.result === r
                    ? r === 'pass' ? 'bg-emerald-600 border-emerald-600 text-white'
                      : r === 'fail' ? 'bg-red-600 border-red-600 text-white'
                      : r === 'observation' ? 'bg-amber-600 border-amber-600 text-white'
                      : 'bg-slate-500 border-slate-500 text-white'
                    : dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
                  class="px-3 py-2 text-sm border rounded-lg capitalize transition-colors font-medium">
                  {{ r === 'na' ? 'N/A' : r.charAt(0).toUpperCase() + r.slice(1) }}
                </button>
              </div>
            </div>
            <!-- Measured value -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">
                Measured Value
                <span v-if="recordTarget?.unit" :class="dk('text-slate-400','text-slate-500')">({{ recordTarget?.unit }})</span>
              </label>
              <input v-model="recordForm.measured_value" type="number"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500"
                placeholder="0.00" />
            </div>
            <!-- Notes -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Notes</label>
              <textarea v-model="recordForm.notes" rows="2"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-purple-500 resize-none"
                placeholder="Observations..." />
            </div>
          </div>
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showRecord = false"
              :class="dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
              class="px-4 py-2 text-sm border rounded-lg transition-colors">Cancel</button>
            <button @click="submitRecord" :disabled="saving"
              :class="recordForm.result === 'pass' ? 'bg-emerald-600 hover:bg-emerald-700' : recordForm.result === 'fail' ? 'bg-red-600 hover:bg-red-700' : 'bg-purple-600 hover:bg-purple-700'"
              class="flex items-center gap-2 px-4 py-2 text-sm text-white rounded-lg font-medium transition-colors disabled:opacity-50">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              Save Result
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
