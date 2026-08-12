<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { qualityAPI } from '@/api/client'
import {
  ClipboardList, Plus, Search, Filter, RefreshCw,
  ChevronLeft, ChevronRight, Play, CheckCircle2,
  Trash2, Eye, EyeOff, ChevronDown, ChevronUp,
  ClipboardCheck, XCircle, Clock, AlertCircle,
  Loader2, Check, X, Minus
} from 'lucide-vue-next'

const app = useAppStore()
const dk = (d: string, l: string) => app.darkMode ? d : l

// ─── state ────────────────────────────────────────────────────────────────────
const loading      = ref(false)
const inspections  = ref<any[]>([])
const total        = ref(0)
const page         = ref(1)
const perPage      = ref(20)

const filterStatus = ref('')
const filterType   = ref('')
const search       = ref('')

const expandedRow  = ref<string | null>(null)
const rowChecks    = ref<Record<string, any[]>>({})
const loadingChecks = ref<Record<string, boolean>>({})

// modal state
const showCreate   = ref(false)
const showComplete = ref(false)
const completeTarget = ref<any>(null)
const saving       = ref(false)

const createForm = ref({
  inspection_type: 'incoming',
  plan_id: '',
  item_id: '',
  lot_number: '',
  qty_to_inspect: '',
  source_type: '',
  source_ref: '',
  scheduled_date: '',
  inspector_id: '',
  notes: '',
})

const completeForm = ref({
  overall_result: 'pass',
  qty_passed: '',
  qty_failed: '',
  notes: '',
})

const recordingCheck = ref<string | null>(null)
const checkResultForm = ref({ result: 'pass', measured_value: '', notes: '' })

// ─── computed ─────────────────────────────────────────────────────────────────
const totalPages = computed(() => Math.ceil(total.value / perPage.value))

const statusOptions = [
  { value: '', label: 'All Statuses' },
  { value: 'pending', label: 'Pending' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'passed', label: 'Passed' },
  { value: 'failed', label: 'Failed' },
  { value: 'cancelled', label: 'Cancelled' },
]

const typeOptions = [
  { value: '', label: 'All Types' },
  { value: 'incoming', label: 'Incoming' },
  { value: 'in_process', label: 'In Process' },
  { value: 'final', label: 'Final' },
  { value: 'audit', label: 'Audit' },
  { value: 'periodic', label: 'Periodic' },
]

// ─── load ──────────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const res = await qualityAPI.listInspections({
      page: page.value,
      per_page: perPage.value,
      status: filterStatus.value || undefined,
      inspection_type: filterType.value || undefined,
      search: search.value || undefined,
    })
    inspections.value = res.data.inspections || []
    total.value       = res.data.total || 0
  } catch {}
  loading.value = false
}

onMounted(load)

function applyFilter() { page.value = 1; load() }
function resetFilter() {
  filterStatus.value = ''; filterType.value = ''; search.value = ''
  page.value = 1; load()
}
function prevPage() { if (page.value > 1) { page.value--; load() } }
function nextPage() { if (page.value < totalPages.value) { page.value++; load() } }

// ─── expand row checks ────────────────────────────────────────────────────────
async function toggleRow(id: string) {
  if (expandedRow.value === id) { expandedRow.value = null; return }
  expandedRow.value = id
  if (!rowChecks.value[id]) {
    loadingChecks.value[id] = true
    try {
      const res = await qualityAPI.getInspection(id)
      rowChecks.value[id] = res.data.checks || []
    } catch { rowChecks.value[id] = [] }
    loadingChecks.value[id] = false
  }
}

// ─── create ───────────────────────────────────────────────────────────────────
function openCreate() {
  createForm.value = {
    inspection_type: 'incoming', plan_id: '', item_id: '',
    lot_number: '', qty_to_inspect: '', source_type: '',
    source_ref: '', scheduled_date: '', inspector_id: '', notes: '',
  }
  showCreate.value = true
}

async function submitCreate() {
  saving.value = true
  try {
    await qualityAPI.createInspection(createForm.value)
    showCreate.value = false
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error creating inspection')
  }
  saving.value = false
}

// ─── start ────────────────────────────────────────────────────────────────────
async function startInspection(id: string) {
  if (!confirm('Start this inspection?')) return
  try {
    await qualityAPI.startInspection(id)
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error')
  }
}

// ─── complete ─────────────────────────────────────────────────────────────────
function openComplete(insp: any) {
  completeTarget.value = insp
  completeForm.value = {
    overall_result: 'pass',
    qty_passed: insp.qty_to_inspect?.toString() || '',
    qty_failed: '0',
    notes: '',
  }
  showComplete.value = true
}

async function submitComplete() {
  saving.value = true
  try {
    await qualityAPI.completeInspection(completeTarget.value.id, completeForm.value)
    showComplete.value = false
    load()
    if (expandedRow.value === completeTarget.value.id) {
      delete rowChecks.value[completeTarget.value.id]
      await toggleRow(completeTarget.value.id)
    }
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error')
  }
  saving.value = false
}

// ─── delete ───────────────────────────────────────────────────────────────────
async function deleteInspection(id: string) {
  if (!confirm('Delete this pending inspection? This cannot be undone.')) return
  try {
    await qualityAPI.deleteInspection(id)
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error')
  }
}

// ─── record check result ──────────────────────────────────────────────────────
function openRecordCheck(checkId: string, currentResult: string) {
  recordingCheck.value = checkId
  checkResultForm.value = { result: currentResult || 'pass', measured_value: '', notes: '' }
}

async function submitCheckResult(inspectionId: string) {
  if (!recordingCheck.value) return
  try {
    await qualityAPI.recordCheckResult(recordingCheck.value, checkResultForm.value)
    recordingCheck.value = null
    delete rowChecks.value[inspectionId]
    await toggleRow(inspectionId)
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error')
  }
}

// ─── helpers ─────────────────────────────────────────────────────────────────
function statusColor(s: string) {
  const m: Record<string, string> = {
    pending: 'bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300',
    in_progress: 'bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300',
    passed: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300',
    failed: 'bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300',
    cancelled: 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400',
  }
  return m[s] || 'bg-slate-100 text-slate-600'
}

function typeColor(t: string) {
  const m: Record<string, string> = {
    incoming: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/40 dark:text-indigo-300',
    in_process: 'bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-300',
    final: 'bg-cyan-100 text-cyan-800 dark:bg-cyan-900/40 dark:text-cyan-300',
    audit: 'bg-orange-100 text-orange-800 dark:bg-orange-900/40 dark:text-orange-300',
    periodic: 'bg-teal-100 text-teal-800 dark:bg-teal-900/40 dark:text-teal-300',
  }
  return m[t] || 'bg-slate-100 text-slate-600'
}

function resultColor(r: string) {
  const m: Record<string, string> = {
    pass: 'text-emerald-600 dark:text-emerald-400',
    fail: 'text-red-600 dark:text-red-400',
    na: 'text-slate-500 dark:text-slate-400',
    observation: 'text-amber-600 dark:text-amber-400',
  }
  return m[r] || 'text-slate-500'
}

function fmtDate(d: any) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString()
}
</script>

<template>
  <div :class="dk('bg-slate-900 text-slate-100', 'bg-slate-50 text-slate-900')" class="min-h-screen">
    <div class="max-w-screen-2xl mx-auto px-6 py-6">

      <!-- Header -->
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-indigo-600 flex items-center justify-center">
            <ClipboardList class="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 class="text-2xl font-bold">Quality Inspections</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">
              {{ total }} inspection{{ total !== 1 ? 's' : '' }} total
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
            class="flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white font-medium text-sm transition-colors">
            <Plus class="w-4 h-4" />
            New Inspection
          </button>
        </div>
      </div>

      <!-- Filters -->
      <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
        class="rounded-xl border p-4 mb-5 flex flex-wrap items-center gap-3">
        <div class="relative flex-1 min-w-[200px]">
          <Search :class="dk('text-slate-400','text-slate-400')" class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" />
          <input v-model="search" @keyup.enter="applyFilter"
            :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-slate-50 border-slate-300 text-slate-900 placeholder-slate-400')"
            class="w-full pl-9 pr-3 py-2 text-sm border rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none"
            placeholder="Search reference, item..." />
        </div>
        <select v-model="filterStatus" @change="applyFilter"
          :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
          class="px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500">
          <option v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <select v-model="filterType" @change="applyFilter"
          :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
          class="px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500">
          <option v-for="o in typeOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <button @click="resetFilter"
          :class="dk('text-slate-400 hover:text-white border-slate-600 hover:border-slate-400','text-slate-500 hover:text-slate-700 border-slate-300')"
          class="px-3 py-2 text-sm border rounded-lg transition-colors">
          Reset
        </button>
      </div>

      <!-- Table -->
      <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
        class="rounded-xl border overflow-hidden">

        <!-- Loading -->
        <div v-if="loading" class="flex items-center justify-center py-20">
          <Loader2 class="w-8 h-8 text-indigo-500 animate-spin" />
        </div>

        <!-- Empty -->
        <div v-else-if="inspections.length === 0" class="flex flex-col items-center justify-center py-20 gap-3">
          <ClipboardList :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12" />
          <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">No inspections found</p>
          <button @click="openCreate"
            class="mt-2 px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium">
            Create First Inspection
          </button>
        </div>

        <!-- Data -->
        <template v-else>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr :class="dk('bg-slate-900/60 text-slate-400 border-b border-slate-700','bg-slate-50 text-slate-500 border-b border-slate-200')">
                  <th class="px-4 py-3 text-left font-medium w-8"></th>
                  <th class="px-4 py-3 text-left font-medium">Reference</th>
                  <th class="px-4 py-3 text-left font-medium">Type</th>
                  <th class="px-4 py-3 text-left font-medium">Status</th>
                  <th class="px-4 py-3 text-left font-medium">Item / Plan</th>
                  <th class="px-4 py-3 text-left font-medium">Lot</th>
                  <th class="px-4 py-3 text-right font-medium">Qty</th>
                  <th class="px-4 py-3 text-right font-medium">Passed</th>
                  <th class="px-4 py-3 text-right font-medium">Failed</th>
                  <th class="px-4 py-3 text-left font-medium">Inspector</th>
                  <th class="px-4 py-3 text-left font-medium">Scheduled</th>
                  <th class="px-4 py-3 text-left font-medium">Result</th>
                  <th class="px-4 py-3 text-center font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                <template v-for="insp in inspections" :key="insp.id">
                  <!-- Main row -->
                  <tr :class="[
                    dk('border-b border-slate-700/60 hover:bg-slate-700/30','border-b border-slate-100 hover:bg-slate-50'),
                    expandedRow === insp.id ? dk('bg-slate-700/20','bg-indigo-50/30') : ''
                  ]" class="transition-colors">
                    <td class="px-4 py-3">
                      <button @click="toggleRow(insp.id)"
                        :class="dk('text-slate-400 hover:text-white','text-slate-400 hover:text-slate-700')"
                        class="transition-colors">
                        <ChevronDown v-if="expandedRow !== insp.id" class="w-4 h-4" />
                        <ChevronUp v-else class="w-4 h-4" />
                      </button>
                    </td>
                    <td class="px-4 py-3 font-mono font-medium" :class="dk('text-indigo-300','text-indigo-600')">
                      {{ insp.reference }}
                    </td>
                    <td class="px-4 py-3">
                      <span :class="typeColor(insp.inspection_type)"
                        class="px-2 py-0.5 rounded-full text-xs font-medium capitalize">
                        {{ insp.inspection_type.replace('_', ' ') }}
                      </span>
                    </td>
                    <td class="px-4 py-3">
                      <span :class="statusColor(insp.status)"
                        class="px-2 py-0.5 rounded-full text-xs font-medium capitalize">
                        {{ insp.status.replace('_', ' ') }}
                      </span>
                    </td>
                    <td class="px-4 py-3">
                      <div class="font-medium" :class="dk('text-slate-100','text-slate-800')">
                        {{ insp.item_name || '—' }}
                      </div>
                      <div v-if="insp.plan_name" class="text-xs" :class="dk('text-slate-400','text-slate-500')">
                        {{ insp.plan_name }}
                      </div>
                    </td>
                    <td class="px-4 py-3 font-mono text-xs" :class="dk('text-slate-300','text-slate-600')">
                      {{ insp.lot_number || '—' }}
                    </td>
                    <td class="px-4 py-3 text-right" :class="dk('text-slate-200','text-slate-700')">
                      {{ insp.qty_to_inspect }}
                    </td>
                    <td class="px-4 py-3 text-right text-emerald-500 font-medium">
                      {{ insp.qty_passed }}
                    </td>
                    <td class="px-4 py-3 text-right text-red-500 font-medium">
                      {{ insp.qty_failed }}
                    </td>
                    <td class="px-4 py-3 text-xs" :class="dk('text-slate-300','text-slate-600')">
                      {{ insp.inspector_name || '—' }}
                    </td>
                    <td class="px-4 py-3 text-xs" :class="dk('text-slate-300','text-slate-600')">
                      {{ fmtDate(insp.scheduled_date) }}
                    </td>
                    <td class="px-4 py-3">
                      <span v-if="insp.overall_result"
                        :class="resultColor(insp.overall_result)"
                        class="text-xs font-semibold uppercase">
                        {{ insp.overall_result }}
                      </span>
                      <span v-else :class="dk('text-slate-500','text-slate-400')" class="text-xs">—</span>
                    </td>
                    <td class="px-4 py-3">
                      <div class="flex items-center justify-center gap-1">
                        <!-- Start -->
                        <button v-if="insp.status === 'pending'"
                          @click="startInspection(insp.id)"
                          title="Start Inspection"
                          class="p-1.5 rounded-lg bg-blue-600/20 hover:bg-blue-600/40 text-blue-400 hover:text-blue-300 transition-colors">
                          <Play class="w-3.5 h-3.5" />
                        </button>
                        <!-- Complete -->
                        <button v-if="insp.status === 'in_progress'"
                          @click="openComplete(insp)"
                          title="Complete Inspection"
                          class="p-1.5 rounded-lg bg-emerald-600/20 hover:bg-emerald-600/40 text-emerald-400 hover:text-emerald-300 transition-colors">
                          <CheckCircle2 class="w-3.5 h-3.5" />
                        </button>
                        <!-- Delete -->
                        <button v-if="insp.status === 'pending'"
                          @click="deleteInspection(insp.id)"
                          title="Delete Inspection"
                          class="p-1.5 rounded-lg bg-red-600/20 hover:bg-red-600/40 text-red-400 hover:text-red-300 transition-colors">
                          <Trash2 class="w-3.5 h-3.5" />
                        </button>
                        <!-- Toggle checks -->
                        <button @click="toggleRow(insp.id)"
                          title="View Checks"
                          :class="expandedRow === insp.id
                            ? 'bg-indigo-600/40 text-indigo-300'
                            : dk('bg-slate-700 hover:bg-slate-600 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-600')"
                          class="p-1.5 rounded-lg transition-colors">
                          <ClipboardCheck class="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>

                  <!-- Expanded checks row -->
                  <tr v-if="expandedRow === insp.id" :key="insp.id + '-checks'"
                    :class="dk('bg-slate-800/80 border-b border-slate-700','bg-indigo-50/40 border-b border-slate-200')">
                    <td colspan="13" class="px-6 py-4">
                      <div v-if="loadingChecks[insp.id]" class="flex items-center gap-2 text-sm"
                        :class="dk('text-slate-400','text-slate-500')">
                        <Loader2 class="w-4 h-4 animate-spin" /> Loading checks...
                      </div>
                      <div v-else>
                        <h4 class="text-sm font-semibold mb-3" :class="dk('text-slate-200','text-slate-700')">
                          Quality Checks ({{ (rowChecks[insp.id] || []).length }})
                        </h4>
                        <div v-if="(rowChecks[insp.id] || []).length === 0"
                          :class="dk('text-slate-500','text-slate-400')" class="text-sm">
                          No checks defined for this inspection.
                        </div>
                        <table v-else class="w-full text-xs">
                          <thead>
                            <tr :class="dk('text-slate-400','text-slate-500')">
                              <th class="text-left py-1 pr-4 font-medium w-6">#</th>
                              <th class="text-left py-1 pr-4 font-medium">Check Name</th>
                              <th class="text-left py-1 pr-4 font-medium">Type</th>
                              <th class="text-left py-1 pr-4 font-medium">Limits</th>
                              <th class="text-left py-1 pr-4 font-medium">Measured</th>
                              <th class="text-left py-1 pr-4 font-medium">Result</th>
                              <th class="text-left py-1 pr-4 font-medium">Checked By</th>
                              <th class="text-left py-1 font-medium">Action</th>
                            </tr>
                          </thead>
                          <tbody>
                            <tr v-for="chk in rowChecks[insp.id]" :key="chk.id"
                              :class="dk('border-t border-slate-700/50','border-t border-slate-200')">
                              <td class="py-2 pr-4" :class="dk('text-slate-400','text-slate-500')">
                                {{ chk.sequence }}
                              </td>
                              <td class="py-2 pr-4 font-medium" :class="dk('text-slate-200','text-slate-700')">
                                {{ chk.name }}
                                <span v-if="chk.is_mandatory" class="ml-1 text-red-400">*</span>
                              </td>
                              <td class="py-2 pr-4 capitalize" :class="dk('text-slate-400','text-slate-500')">
                                {{ chk.check_type }}
                              </td>
                              <td class="py-2 pr-4 font-mono" :class="dk('text-slate-300','text-slate-600')">
                                <span v-if="chk.min_value !== null || chk.max_value !== null">
                                  {{ chk.min_value ?? '—' }} – {{ chk.max_value ?? '—' }}
                                  <span v-if="chk.unit" class="text-slate-500"> {{ chk.unit }}</span>
                                </span>
                                <span v-else :class="dk('text-slate-500','text-slate-400')">—</span>
                              </td>
                              <td class="py-2 pr-4 font-mono" :class="dk('text-slate-300','text-slate-600')">
                                {{ chk.measured_value !== null ? chk.measured_value : '—' }}
                              </td>
                              <td class="py-2 pr-4">
                                <span v-if="chk.result" :class="resultColor(chk.result)"
                                  class="font-semibold uppercase">
                                  {{ chk.result }}
                                </span>
                                <span v-else :class="dk('text-slate-500','text-slate-400')">Pending</span>
                              </td>
                              <td class="py-2 pr-4" :class="dk('text-slate-400','text-slate-500')">
                                {{ chk.checked_by || '—' }}
                              </td>
                              <td class="py-2">
                                <!-- Inline result form or record button -->
                                <template v-if="recordingCheck === chk.id">
                                  <div class="flex items-center gap-2 flex-wrap">
                                    <select v-model="checkResultForm.result"
                                      :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                                      class="px-2 py-1 text-xs border rounded outline-none">
                                      <option value="pass">Pass</option>
                                      <option value="fail">Fail</option>
                                      <option value="na">N/A</option>
                                      <option value="observation">Obs.</option>
                                    </select>
                                    <input v-model="checkResultForm.measured_value"
                                      :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                                      class="px-2 py-1 text-xs border rounded w-20 outline-none"
                                      placeholder="Measured" />
                                    <button @click="submitCheckResult(insp.id)"
                                      class="p-1 rounded bg-emerald-600 hover:bg-emerald-700 text-white transition-colors">
                                      <Check class="w-3 h-3" />
                                    </button>
                                    <button @click="recordingCheck = null"
                                      :class="dk('bg-slate-600 hover:bg-slate-500','bg-slate-200 hover:bg-slate-300')"
                                      class="p-1 rounded text-slate-400 transition-colors">
                                      <X class="w-3 h-3" />
                                    </button>
                                  </div>
                                </template>
                                <template v-else>
                                  <button
                                    v-if="insp.status === 'in_progress'"
                                    @click="openRecordCheck(chk.id, chk.result)"
                                    class="px-2 py-1 rounded text-xs bg-indigo-600/30 hover:bg-indigo-600/50 text-indigo-300 hover:text-indigo-200 transition-colors">
                                    Record
                                  </button>
                                </template>
                              </td>
                            </tr>
                          </tbody>
                        </table>
                      </div>
                    </td>
                  </tr>
                </template>
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

    <!-- ─── Create Modal ───────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showCreate"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="relative w-full max-w-2xl rounded-2xl border shadow-2xl max-h-[90vh] overflow-y-auto">
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-between px-6 py-4 border-b">
            <h2 class="text-lg font-semibold" :class="dk('text-white','text-slate-800')">
              New Quality Inspection
            </h2>
            <button @click="showCreate = false"
              :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-700')"
              class="transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5 grid grid-cols-2 gap-4">
            <!-- Inspection Type -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">
                Inspection Type <span class="text-red-400">*</span>
              </label>
              <select v-model="createForm.inspection_type"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="incoming">Incoming</option>
                <option value="in_process">In Process</option>
                <option value="final">Final</option>
                <option value="audit">Audit</option>
                <option value="periodic">Periodic</option>
              </select>
            </div>
            <!-- Item ID -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Item ID</label>
              <input v-model="createForm.item_id"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="Item UUID" />
            </div>
            <!-- Plan ID -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Control Plan ID</label>
              <input v-model="createForm.plan_id"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="Plan UUID (optional)" />
            </div>
            <!-- Lot Number -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Lot Number</label>
              <input v-model="createForm.lot_number"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="LOT-001" />
            </div>
            <!-- Qty to Inspect -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">
                Qty to Inspect <span class="text-red-400">*</span>
              </label>
              <input v-model="createForm.qty_to_inspect" type="number" min="0"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500" />
            </div>
            <!-- Scheduled Date -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Scheduled Date</label>
              <input v-model="createForm.scheduled_date" type="date"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500" />
            </div>
            <!-- Source Type -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Source Type</label>
              <select v-model="createForm.source_type"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="">— None —</option>
                <option value="purchase_order">Purchase Order</option>
                <option value="manufacturing_order">Manufacturing Order</option>
                <option value="stock_movement">Stock Movement</option>
                <option value="manual">Manual</option>
              </select>
            </div>
            <!-- Source Ref -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Source Reference</label>
              <input v-model="createForm.source_ref"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="PO-00123" />
            </div>
            <!-- Inspector ID -->
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Inspector ID</label>
              <input v-model="createForm.inspector_id"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500"
                placeholder="User UUID" />
            </div>
            <!-- Notes -->
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Notes</label>
              <textarea v-model="createForm.notes" rows="2"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
                placeholder="Optional notes..." />
            </div>
          </div>
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showCreate = false"
              :class="dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
              class="px-4 py-2 text-sm border rounded-lg transition-colors">
              Cancel
            </button>
            <button @click="submitCreate" :disabled="saving"
              class="flex items-center gap-2 px-4 py-2 text-sm bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              Create Inspection
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ─── Complete Modal ────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showComplete"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="relative w-full max-w-md rounded-2xl border shadow-2xl">
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-between px-6 py-4 border-b">
            <h2 class="text-lg font-semibold" :class="dk('text-white','text-slate-800')">
              Complete Inspection
            </h2>
            <button @click="showComplete = false"
              :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-700')">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5 space-y-4">
            <p class="text-sm" :class="dk('text-slate-300','text-slate-600')">
              Inspecting: <span class="font-semibold">{{ completeTarget?.reference }}</span>
            </p>
            <!-- Overall Result -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-2">
                Overall Result <span class="text-red-400">*</span>
              </label>
              <div class="flex gap-3">
                <button v-for="r in ['pass','fail','na','observation']" :key="r"
                  @click="completeForm.overall_result = r"
                  :class="completeForm.overall_result === r
                    ? r === 'pass' ? 'bg-emerald-600 text-white' : r === 'fail' ? 'bg-red-600 text-white' : 'bg-indigo-600 text-white'
                    : dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
                  class="px-3 py-1.5 text-sm border rounded-lg transition-colors capitalize">
                  {{ r }}
                </button>
              </div>
            </div>
            <!-- Qty Passed / Failed -->
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Qty Passed</label>
                <input v-model="completeForm.qty_passed" type="number" min="0"
                  :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                  class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-emerald-500" />
              </div>
              <div>
                <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Qty Failed</label>
                <input v-model="completeForm.qty_failed" type="number" min="0"
                  :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                  class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-red-500" />
              </div>
            </div>
            <!-- Notes -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Notes</label>
              <textarea v-model="completeForm.notes" rows="2"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-indigo-500 resize-none"
                placeholder="Completion notes..." />
            </div>
          </div>
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showComplete = false"
              :class="dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
              class="px-4 py-2 text-sm border rounded-lg transition-colors">
              Cancel
            </button>
            <button @click="submitComplete" :disabled="saving"
              class="flex items-center gap-2 px-4 py-2 text-sm bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              Complete
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
