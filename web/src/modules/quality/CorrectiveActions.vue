<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { qualityAPI } from '@/api/client'
import {
  Wrench, Plus, Search, RefreshCw,
  ChevronLeft, ChevronRight, Loader2, X,
  AlertTriangle, Clock, CheckCircle2, ShieldCheck,
  ArrowRight, Pencil, Trash2, ListChecks, Star,
  ChevronDown, ChevronUp, Check
} from 'lucide-vue-next'

const app = useAppStore()
const dk = (d: string, l: string) => app.darkMode ? d : l

// ─── state ────────────────────────────────────────────────────────────────────
const loading = ref(false)
const cas     = ref<any[]>([])
const total   = ref(0)
const page    = ref(1)
const perPage = ref(20)

const filterStatus   = ref('')
const filterType     = ref('')
const filterPriority = ref('')
const search         = ref('')

const showCreate  = ref(false)
const showEdit    = ref(false)
const showStatus  = ref(false)
const showVerify  = ref(false)
const editTarget  = ref<any>(null)
const statusTarget = ref<any>(null)
const saving      = ref(false)

const expandedCA  = ref<string | null>(null)

const createForm = ref({
  title: '', description: '',
  ca_type: 'corrective', priority: 'medium',
  nc_id: '',
  root_cause: '', root_cause_method: '5why',
  proposed_action: '',
  responsible_id: '', department_id: '',
  due_date: '', estimated_cost: '',
})

const editForm = ref<any>({})

const statusForm = ref({ status: '', })

const verifyForm = ref({
  effectiveness_rating: 3,
  effectiveness_notes: '',
})

// ─── computed ─────────────────────────────────────────────────────────────────
const totalPages = computed(() => Math.ceil(total.value / perPage.value))

const statusOptions = [
  { value: '', label: 'All Statuses' },
  { value: 'open', label: 'Open' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'pending_verification', label: 'Pending Verification' },
  { value: 'verified', label: 'Verified' },
  { value: 'closed', label: 'Closed' },
  { value: 'cancelled', label: 'Cancelled' },
]

const typeOptions = [
  { value: '', label: 'All Types' },
  { value: 'corrective', label: 'Corrective' },
  { value: 'preventive', label: 'Preventive' },
  { value: 'improvement', label: 'Improvement' },
]

const priorityOptions = [
  { value: '', label: 'All Priorities' },
  { value: 'critical', label: 'Critical' },
  { value: 'high', label: 'High' },
  { value: 'medium', label: 'Medium' },
  { value: 'low', label: 'Low' },
]

const caStatusTransitions: Record<string, string[]> = {
  open: ['in_progress', 'cancelled'],
  in_progress: ['pending_verification', 'open'],
  pending_verification: ['verified', 'in_progress'],
  verified: ['closed'],
  closed: [],
  cancelled: ['open'],
}

// ─── load ─────────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const res = await qualityAPI.listCorrectiveActions({
      page: page.value, per_page: perPage.value,
      status: filterStatus.value || undefined,
      ca_type: filterType.value || undefined,
      priority: filterPriority.value || undefined,
      search: search.value || undefined,
    })
    cas.value   = res.data.corrective_actions || []
    total.value = res.data.total || 0
  } catch {}
  loading.value = false
}

onMounted(load)

function applyFilter() { page.value = 1; load() }
function resetFilter() {
  filterStatus.value = ''; filterType.value = ''; filterPriority.value = ''; search.value = ''
  page.value = 1; load()
}
function prevPage() { if (page.value > 1) { page.value--; load() } }
function nextPage() { if (page.value < totalPages.value) { page.value++; load() } }

// ─── expand tasks ─────────────────────────────────────────────────────────────
async function toggleCA(id: string) {
  expandedCA.value = expandedCA.value === id ? null : id
}

// ─── create ───────────────────────────────────────────────────────────────────
function openCreate() {
  createForm.value = {
    title: '', description: '', ca_type: 'corrective', priority: 'medium',
    nc_id: '', root_cause: '', root_cause_method: '5why',
    proposed_action: '', responsible_id: '', department_id: '',
    due_date: '', estimated_cost: '',
  }
  showCreate.value = true
}

async function submitCreate() {
  if (!createForm.value.title) { alert('Title is required.'); return }
  saving.value = true
  try {
    await qualityAPI.createCorrectiveAction(createForm.value)
    showCreate.value = false
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error creating corrective action')
  }
  saving.value = false
}

// ─── edit ─────────────────────────────────────────────────────────────────────
function openEdit(ca: any) {
  editTarget.value = ca
  editForm.value = { ...ca }
  showEdit.value = true
}

async function submitEdit() {
  saving.value = true
  try {
    await qualityAPI.updateCorrectiveAction(editTarget.value.id, editForm.value)
    showEdit.value = false
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error updating')
  }
  saving.value = false
}

// ─── status ───────────────────────────────────────────────────────────────────
function openStatus(ca: any) {
  statusTarget.value = ca
  const next = caStatusTransitions[ca.status] || []
  if (next.includes('verified')) {
    verifyForm.value = { effectiveness_rating: 3, effectiveness_notes: '' }
    showVerify.value = true
    return
  }
  statusForm.value = { status: next[0] || '' }
  showStatus.value = true
}

async function submitStatus() {
  if (!statusForm.value.status) { alert('Select a status.'); return }
  saving.value = true
  try {
    await qualityAPI.updateCAStatus(statusTarget.value.id, statusForm.value)
    showStatus.value = false
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error updating status')
  }
  saving.value = false
}

async function submitVerify() {
  saving.value = true
  try {
    await qualityAPI.updateCAStatus(statusTarget.value.id, {
      status: 'verified',
      effectiveness_rating: verifyForm.value.effectiveness_rating,
      effectiveness_notes: verifyForm.value.effectiveness_notes,
    })
    showVerify.value = false
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error verifying')
  }
  saving.value = false
}

// ─── delete ───────────────────────────────────────────────────────────────────
async function deleteCA(id: string) {
  if (!confirm('Delete this corrective action? Only open CAs can be deleted.')) return
  try {
    await qualityAPI.deleteCorrectiveAction(id)
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error')
  }
}

// ─── helpers ─────────────────────────────────────────────────────────────────
function priorityColor(p: string) {
  const m: Record<string, string> = {
    critical: 'bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300',
    high: 'bg-orange-100 text-orange-800 dark:bg-orange-900/40 dark:text-orange-300',
    medium: 'bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300',
    low: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400',
  }
  return m[p] || 'bg-slate-100 text-slate-600'
}

function statusColor(s: string) {
  const m: Record<string, string> = {
    open: 'bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300',
    in_progress: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/40 dark:text-indigo-300',
    pending_verification: 'bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-300',
    verified: 'bg-teal-100 text-teal-800 dark:bg-teal-900/40 dark:text-teal-300',
    closed: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300',
    cancelled: 'bg-slate-100 text-slate-500 dark:bg-slate-700 dark:text-slate-400',
  }
  return m[s] || 'bg-slate-100 text-slate-600'
}

function statusLabel(s: string) {
  const m: Record<string, string> = {
    open: 'Open', in_progress: 'In Progress',
    pending_verification: 'Pending Verification',
    verified: 'Verified', closed: 'Closed', cancelled: 'Cancelled',
  }
  return m[s] || s
}

function typeColor(t: string) {
  const m: Record<string, string> = {
    corrective: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
    preventive: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
    improvement: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300',
  }
  return m[t] || 'bg-slate-100 text-slate-600'
}

function methodLabel(m: string) {
  const labels: Record<string, string> = {
    '5why': '5 Why', fishbone: 'Fishbone',
    fmea: 'FMEA', brainstorming: 'Brainstorming', other: 'Other',
  }
  return labels[m] || m
}

function fmtDate(d: any) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString()
}

function starClass(rating: number, current: number) {
  return current <= rating
    ? 'text-amber-400 fill-amber-400'
    : dk('text-slate-600 fill-slate-600','text-slate-300 fill-slate-300')
}
</script>

<template>
  <div :class="dk('bg-slate-900 text-slate-100','bg-slate-50 text-slate-900')" class="min-h-screen">
    <div class="max-w-screen-2xl mx-auto px-6 py-6">

      <!-- Header -->
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-teal-600 flex items-center justify-center">
            <Wrench class="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 class="text-2xl font-bold">Corrective Actions</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">
              {{ total }} action{{ total !== 1 ? 's' : '' }} total
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
            class="flex items-center gap-2 px-4 py-2 rounded-lg bg-teal-600 hover:bg-teal-700 text-white font-medium text-sm transition-colors">
            <Plus class="w-4 h-4" />
            New CA
          </button>
        </div>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6">
        <div v-for="item in [
          { label: 'Total', value: total, color: 'teal' },
          { label: 'Open', value: cas.filter(c => ['open','in_progress'].includes(c.status)).length, color: 'blue' },
          { label: 'Overdue', value: cas.filter(c => c.is_overdue).length, color: 'red' },
          { label: 'Pending Tasks', value: cas.reduce((s, c) => s + (c.pending_tasks || 0), 0), color: 'amber' },
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
        <div class="relative flex-1 min-w-[200px]">
          <Search :class="dk('text-slate-400','text-slate-400')" class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" />
          <input v-model="search" @keyup.enter="applyFilter"
            :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-slate-50 border-slate-300 text-slate-900 placeholder-slate-400')"
            class="w-full pl-9 pr-3 py-2 text-sm border rounded-lg focus:ring-2 focus:ring-teal-500 outline-none"
            placeholder="Search reference, title..." />
        </div>
        <select v-model="filterStatus" @change="applyFilter"
          :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
          class="px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500">
          <option v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <select v-model="filterType" @change="applyFilter"
          :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
          class="px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500">
          <option v-for="o in typeOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <select v-model="filterPriority" @change="applyFilter"
          :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
          class="px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500">
          <option v-for="o in priorityOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
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
          <Loader2 class="w-8 h-8 text-teal-500 animate-spin" />
        </div>
        <div v-else-if="cas.length === 0" class="flex flex-col items-center justify-center py-20 gap-3">
          <Wrench :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12" />
          <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">No corrective actions found</p>
          <button @click="openCreate"
            class="mt-2 px-4 py-2 rounded-lg bg-teal-600 hover:bg-teal-700 text-white text-sm font-medium">
            Create First CA
          </button>
        </div>

        <template v-else>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr :class="dk('bg-slate-900/60 text-slate-400 border-b border-slate-700','bg-slate-50 text-slate-500 border-b border-slate-200')">
                  <th class="px-4 py-3 text-left font-medium w-8"></th>
                  <th class="px-4 py-3 text-left font-medium">Priority</th>
                  <th class="px-4 py-3 text-left font-medium">Reference</th>
                  <th class="px-4 py-3 text-left font-medium">Title</th>
                  <th class="px-4 py-3 text-left font-medium">Type</th>
                  <th class="px-4 py-3 text-left font-medium">Status</th>
                  <th class="px-4 py-3 text-left font-medium">NC</th>
                  <th class="px-4 py-3 text-left font-medium">Responsible</th>
                  <th class="px-4 py-3 text-left font-medium">Due Date</th>
                  <th class="px-4 py-3 text-left font-medium">Method</th>
                  <th class="px-4 py-3 text-right font-medium">Est. Cost</th>
                  <th class="px-4 py-3 text-left font-medium">Effectiveness</th>
                  <th class="px-4 py-3 text-center font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                <template v-for="ca in cas" :key="ca.id">
                  <tr :class="[
                    dk('border-b border-slate-700/60 hover:bg-slate-700/30','border-b border-slate-100 hover:bg-slate-50'),
                    ca.is_overdue ? dk('bg-red-900/10','bg-red-50/30') : '',
                    expandedCA === ca.id ? dk('bg-slate-700/20','bg-teal-50/30') : ''
                  ]" class="transition-colors">
                    <td class="px-4 py-3">
                      <button @click="toggleCA(ca.id)"
                        :class="dk('text-slate-400 hover:text-white','text-slate-400 hover:text-slate-700')">
                        <ChevronDown v-if="expandedCA !== ca.id" class="w-4 h-4" />
                        <ChevronUp v-else class="w-4 h-4" />
                      </button>
                    </td>
                    <td class="px-4 py-3">
                      <span :class="priorityColor(ca.priority)"
                        class="px-2 py-0.5 rounded-full text-xs font-semibold capitalize">
                        {{ ca.priority }}
                      </span>
                    </td>
                    <td class="px-4 py-3 font-mono font-medium" :class="dk('text-teal-300','text-teal-600')">
                      {{ ca.reference }}
                    </td>
                    <td class="px-4 py-3 max-w-xs">
                      <div class="font-medium truncate" :class="dk('text-slate-100','text-slate-800')" :title="ca.title">
                        {{ ca.title }}
                      </div>
                      <div v-if="ca.pending_tasks > 0" class="text-xs mt-0.5 text-amber-500">
                        {{ ca.pending_tasks }} pending task{{ ca.pending_tasks !== 1 ? 's' : '' }}
                      </div>
                    </td>
                    <td class="px-4 py-3">
                      <span :class="typeColor(ca.ca_type)"
                        class="px-2 py-0.5 rounded-full text-xs font-medium capitalize">
                        {{ ca.ca_type }}
                      </span>
                    </td>
                    <td class="px-4 py-3">
                      <span :class="statusColor(ca.status)"
                        class="px-2 py-0.5 rounded-full text-xs font-medium">
                        {{ statusLabel(ca.status) }}
                      </span>
                    </td>
                    <td class="px-4 py-3 text-xs">
                      <div v-if="ca.nc_ref" class="font-mono" :class="dk('text-orange-300','text-orange-600')">
                        {{ ca.nc_ref }}
                      </div>
                      <div v-if="ca.nc_severity" :class="dk('text-slate-400','text-slate-500')">{{ ca.nc_severity }}</div>
                      <span v-if="!ca.nc_ref" :class="dk('text-slate-600','text-slate-400')">—</span>
                    </td>
                    <td class="px-4 py-3 text-xs" :class="dk('text-slate-300','text-slate-600')">
                      {{ ca.responsible_name || '—' }}
                      <div v-if="ca.dept_name" :class="dk('text-slate-500','text-slate-400')">{{ ca.dept_name }}</div>
                    </td>
                    <td class="px-4 py-3 text-xs"
                      :class="ca.is_overdue ? 'text-red-500 font-semibold' : dk('text-slate-300','text-slate-600')">
                      {{ fmtDate(ca.due_date) }}
                      <div v-if="ca.is_overdue" class="text-red-400 font-normal text-xs">Overdue</div>
                    </td>
                    <td class="px-4 py-3 text-xs" :class="dk('text-slate-400','text-slate-500')">
                      {{ methodLabel(ca.root_cause_method || '5why') }}
                    </td>
                    <td class="px-4 py-3 text-right text-xs font-mono" :class="dk('text-slate-300','text-slate-600')">
                      {{ ca.estimated_cost > 0 ? ca.estimated_cost.toLocaleString() : '—' }}
                    </td>
                    <td class="px-4 py-3">
                      <div v-if="ca.effectiveness_rating" class="flex items-center gap-0.5">
                        <svg v-for="i in 5" :key="i" viewBox="0 0 20 20"
                          :class="starClass(ca.effectiveness_rating, i)"
                          class="w-3.5 h-3.5">
                          <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                        </svg>
                        <span class="ml-1 text-xs" :class="dk('text-slate-400','text-slate-500')">
                          {{ ca.effectiveness_rating }}/5
                        </span>
                      </div>
                      <span v-else :class="dk('text-slate-600','text-slate-400')" class="text-xs">—</span>
                    </td>
                    <td class="px-4 py-3">
                      <div class="flex items-center justify-center gap-1">
                        <button v-if="(caStatusTransitions[ca.status] || []).length > 0"
                          @click="openStatus(ca)" title="Advance Status"
                          class="p-1.5 rounded-lg bg-blue-600/20 hover:bg-blue-600/40 text-blue-400 hover:text-blue-300 transition-colors">
                          <ArrowRight class="w-3.5 h-3.5" />
                        </button>
                        <button @click="openEdit(ca)" title="Edit"
                          class="p-1.5 rounded-lg bg-amber-600/20 hover:bg-amber-600/40 text-amber-400 hover:text-amber-300 transition-colors">
                          <Pencil class="w-3.5 h-3.5" />
                        </button>
                        <button v-if="ca.status === 'open'"
                          @click="deleteCA(ca.id)" title="Delete"
                          class="p-1.5 rounded-lg bg-red-600/20 hover:bg-red-600/40 text-red-400 hover:text-red-300 transition-colors">
                          <Trash2 class="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </td>
                  </tr>

                  <!-- Expanded detail row -->
                  <tr v-if="expandedCA === ca.id" :key="ca.id + '-detail'"
                    :class="dk('bg-slate-800/80 border-b border-slate-700','bg-teal-50/30 border-b border-slate-200')">
                    <td colspan="13" class="px-6 py-5">
                      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
                        <!-- Root Cause -->
                        <div>
                          <h4 class="text-xs font-semibold uppercase tracking-wide mb-2"
                            :class="dk('text-slate-400','text-slate-500')">
                            Root Cause ({{ methodLabel(ca.root_cause_method || '5why') }})
                          </h4>
                          <p class="text-sm" :class="dk('text-slate-200','text-slate-700')">
                            {{ ca.proposed_action || ca.root_cause || '—' }}
                          </p>
                        </div>
                        <!-- Proposed Action -->
                        <div>
                          <h4 class="text-xs font-semibold uppercase tracking-wide mb-2"
                            :class="dk('text-slate-400','text-slate-500')">Proposed Action</h4>
                          <p class="text-sm" :class="dk('text-slate-200','text-slate-700')">
                            {{ ca.proposed_action || '—' }}
                          </p>
                        </div>
                        <!-- Implemented Action -->
                        <div>
                          <h4 class="text-xs font-semibold uppercase tracking-wide mb-2"
                            :class="dk('text-slate-400','text-slate-500')">Implemented Action</h4>
                          <p class="text-sm" :class="dk('text-slate-200','text-slate-700')">
                            {{ ca.implemented_action || '—' }}
                          </p>
                        </div>
                      </div>
                      <!-- Cost info -->
                      <div v-if="ca.estimated_cost > 0 || ca.actual_cost > 0"
                        class="mt-4 flex items-center gap-6 text-sm">
                        <div>
                          <span :class="dk('text-slate-400','text-slate-500')" class="text-xs">Estimated Cost: </span>
                          <span class="font-semibold font-mono" :class="dk('text-slate-200','text-slate-700')">
                            {{ ca.estimated_cost.toLocaleString() }}
                          </span>
                        </div>
                        <div>
                          <span :class="dk('text-slate-400','text-slate-500')" class="text-xs">Actual Cost: </span>
                          <span class="font-semibold font-mono" :class="dk('text-slate-200','text-slate-700')">
                            {{ ca.actual_cost.toLocaleString() }}
                          </span>
                        </div>
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
          class="relative w-full max-w-3xl rounded-2xl border shadow-2xl max-h-[90vh] overflow-y-auto">
          <div :class="dk('border-slate-700 bg-slate-800','border-slate-200 bg-white')"
            class="flex items-center justify-between px-6 py-4 border-b sticky top-0">
            <h2 class="text-lg font-semibold" :class="dk('text-white','text-slate-800')">New Corrective Action</h2>
            <button @click="showCreate = false"
              :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-700')">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5 grid grid-cols-2 gap-4">
            <!-- Title -->
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">
                Title <span class="text-red-400">*</span>
              </label>
              <input v-model="createForm.title"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500"
                placeholder="Corrective action title" />
            </div>
            <!-- Type -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Type</label>
              <select v-model="createForm.ca_type"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500">
                <option value="corrective">Corrective</option>
                <option value="preventive">Preventive</option>
                <option value="improvement">Improvement</option>
              </select>
            </div>
            <!-- Priority -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Priority</label>
              <select v-model="createForm.priority"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500">
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>
            <!-- NC ID -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Linked NC ID</label>
              <input v-model="createForm.nc_id"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500"
                placeholder="NC UUID (optional)" />
            </div>
            <!-- Root Cause Method -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Root Cause Method</label>
              <select v-model="createForm.root_cause_method"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500">
                <option value="5why">5 Why</option>
                <option value="fishbone">Fishbone (Ishikawa)</option>
                <option value="fmea">FMEA</option>
                <option value="brainstorming">Brainstorming</option>
                <option value="other">Other</option>
              </select>
            </div>
            <!-- Root Cause -->
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Root Cause Analysis</label>
              <textarea v-model="createForm.root_cause" rows="3"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500 resize-none"
                placeholder="Describe root cause analysis..." />
            </div>
            <!-- Proposed Action -->
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Proposed Action</label>
              <textarea v-model="createForm.proposed_action" rows="2"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500 resize-none"
                placeholder="Describe proposed corrective action..." />
            </div>
            <!-- Responsible -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Responsible (User ID)</label>
              <input v-model="createForm.responsible_id"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500"
                placeholder="User UUID" />
            </div>
            <!-- Department -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Department ID</label>
              <input v-model="createForm.department_id"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500"
                placeholder="Department UUID" />
            </div>
            <!-- Due Date -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Due Date</label>
              <input v-model="createForm.due_date" type="date"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500" />
            </div>
            <!-- Estimated Cost -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Estimated Cost</label>
              <input v-model="createForm.estimated_cost" type="number" min="0"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500" />
            </div>
          </div>
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showCreate = false"
              :class="dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
              class="px-4 py-2 text-sm border rounded-lg transition-colors">Cancel</button>
            <button @click="submitCreate" :disabled="saving"
              class="flex items-center gap-2 px-4 py-2 text-sm bg-teal-600 hover:bg-teal-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              Create CA
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ─── Edit Modal ─────────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showEdit"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="relative w-full max-w-2xl rounded-2xl border shadow-2xl max-h-[90vh] overflow-y-auto">
          <div :class="dk('border-slate-700 bg-slate-800','border-slate-200 bg-white')"
            class="flex items-center justify-between px-6 py-4 border-b sticky top-0">
            <h2 class="text-lg font-semibold" :class="dk('text-white','text-slate-800')">
              Edit CA — {{ editTarget?.reference }}
            </h2>
            <button @click="showEdit = false"
              :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-700')">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5 grid grid-cols-2 gap-4">
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Title</label>
              <input v-model="editForm.title"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Priority</label>
              <select v-model="editForm.priority"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500">
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Due Date</label>
              <input v-model="editForm.due_date" type="date"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Responsible ID</label>
              <input v-model="editForm.responsible_id"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Implementation Date</label>
              <input v-model="editForm.implementation_date" type="date"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500" />
            </div>
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Implemented Action</label>
              <textarea v-model="editForm.implemented_action" rows="3"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500 resize-none" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Estimated Cost</label>
              <input v-model="editForm.estimated_cost" type="number" min="0"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Actual Cost</label>
              <input v-model="editForm.actual_cost" type="number" min="0"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500" />
            </div>
          </div>
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showEdit = false"
              :class="dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
              class="px-4 py-2 text-sm border rounded-lg transition-colors">Cancel</button>
            <button @click="submitEdit" :disabled="saving"
              class="flex items-center gap-2 px-4 py-2 text-sm bg-amber-600 hover:bg-amber-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              Save Changes
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ─── Status Modal ──────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showStatus"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="relative w-full max-w-sm rounded-2xl border shadow-2xl">
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-between px-6 py-4 border-b">
            <h2 class="text-lg font-semibold" :class="dk('text-white','text-slate-800')">Update CA Status</h2>
            <button @click="showStatus = false"
              :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-700')">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5 space-y-4">
            <p class="text-sm" :class="dk('text-slate-300','text-slate-600')">
              Current: <span :class="statusColor(statusTarget?.status)"
                class="px-2 py-0.5 rounded-full text-xs font-medium ml-1">
                {{ statusLabel(statusTarget?.status) }}
              </span>
            </p>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-2">New Status</label>
              <div class="flex flex-wrap gap-2">
                <button v-for="s in (caStatusTransitions[statusTarget?.status] || [])" :key="s"
                  @click="statusForm.status = s"
                  :class="statusForm.status === s
                    ? 'bg-teal-600 border-teal-600 text-white'
                    : dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
                  class="px-3 py-1.5 text-sm border rounded-lg transition-colors">
                  {{ statusLabel(s) }}
                </button>
              </div>
            </div>
          </div>
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showStatus = false"
              :class="dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
              class="px-4 py-2 text-sm border rounded-lg transition-colors">Cancel</button>
            <button @click="submitStatus" :disabled="saving"
              class="flex items-center gap-2 px-4 py-2 text-sm bg-teal-600 hover:bg-teal-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              Update
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ─── Verify Modal ──────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showVerify"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
        <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="relative w-full max-w-md rounded-2xl border shadow-2xl">
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-between px-6 py-4 border-b">
            <h2 class="text-lg font-semibold" :class="dk('text-white','text-slate-800')">Verify Effectiveness</h2>
            <button @click="showVerify = false"
              :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-700')">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5 space-y-5">
            <p class="text-sm" :class="dk('text-slate-300','text-slate-600')">
              Verifying: <span class="font-semibold">{{ statusTarget?.reference }}</span>
            </p>
            <!-- Star rating -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-3">
                Effectiveness Rating (1 = Poor, 5 = Excellent)
              </label>
              <div class="flex items-center gap-2">
                <button v-for="i in 5" :key="i"
                  @click="verifyForm.effectiveness_rating = i"
                  class="transition-all hover:scale-110">
                  <svg viewBox="0 0 20 20"
                    :class="i <= verifyForm.effectiveness_rating ? 'text-amber-400 fill-amber-400' : dk('text-slate-600 fill-slate-600','text-slate-300 fill-slate-300')"
                    class="w-7 h-7">
                    <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                  </svg>
                </button>
                <span class="ml-2 text-sm font-semibold text-amber-400">{{ verifyForm.effectiveness_rating }}/5</span>
              </div>
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Effectiveness Notes</label>
              <textarea v-model="verifyForm.effectiveness_notes" rows="3"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-teal-500 resize-none"
                placeholder="Describe how effective the corrective action was..." />
            </div>
          </div>
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showVerify = false"
              :class="dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
              class="px-4 py-2 text-sm border rounded-lg transition-colors">Cancel</button>
            <button @click="submitVerify" :disabled="saving"
              class="flex items-center gap-2 px-4 py-2 text-sm bg-teal-600 hover:bg-teal-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              <ShieldCheck class="w-4 h-4" />
              Verify
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
