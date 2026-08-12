<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { qualityAPI } from '@/api/client'
import {
  AlertOctagon, Plus, Search, RefreshCw,
  ChevronLeft, ChevronRight, Loader2, X,
  AlertTriangle, ShieldAlert, Info, Zap,
  Pencil, Trash2, ChevronDown, ArrowRight
} from 'lucide-vue-next'

const app = useAppStore()
const dk = (d: string, l: string) => app.darkMode ? d : l

// ─── state ────────────────────────────────────────────────────────────────────
const loading = ref(false)
const ncs     = ref<any[]>([])
const total   = ref(0)
const page    = ref(1)
const perPage = ref(20)

const filterStatus   = ref('')
const filterSeverity = ref('')
const search         = ref('')

const showCreate   = ref(false)
const showEdit     = ref(false)
const showStatus   = ref(false)
const editTarget   = ref<any>(null)
const statusTarget = ref<any>(null)
const saving       = ref(false)

const createForm = ref({
  title: '', description: '', severity: 'minor',
  source_type: '', source_ref: '', inspection_id: '',
  item_id: '', lot_number: '', qty_affected: '',
  department_id: '', process: '',
  detected_by: '', detected_date: new Date().toISOString().split('T')[0],
  assigned_to: '', target_date: '',
  root_cause: '', immediate_action: '',
})

const editForm = ref<any>({})

const statusForm = ref({
  status: '',
  closure_notes: '',
})

// ─── computed ─────────────────────────────────────────────────────────────────
const totalPages = computed(() => Math.ceil(total.value / perPage.value))

const statusOptions = [
  { value: '', label: 'All Statuses' },
  { value: 'open', label: 'Open' },
  { value: 'under_review', label: 'Under Review' },
  { value: 'corrective_action', label: 'Corrective Action' },
  { value: 'closed', label: 'Closed' },
  { value: 'cancelled', label: 'Cancelled' },
]

const severityOptions = [
  { value: '', label: 'All Severities' },
  { value: 'minor', label: 'Minor' },
  { value: 'major', label: 'Major' },
  { value: 'critical', label: 'Critical' },
  { value: 'critical_safety', label: 'Critical Safety' },
]

const statusTransitions: Record<string, string[]> = {
  open: ['under_review', 'cancelled'],
  under_review: ['corrective_action', 'closed', 'open'],
  corrective_action: ['closed'],
  closed: [],
  cancelled: ['open'],
}

// ─── load ─────────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const res = await qualityAPI.listNonConformities({
      page: page.value,
      per_page: perPage.value,
      status: filterStatus.value || undefined,
      severity: filterSeverity.value || undefined,
      search: search.value || undefined,
    })
    ncs.value   = res.data.non_conformities || []
    total.value = res.data.total || 0
  } catch {}
  loading.value = false
}

onMounted(load)

function applyFilter() { page.value = 1; load() }
function resetFilter() {
  filterStatus.value = ''; filterSeverity.value = ''; search.value = ''
  page.value = 1; load()
}
function prevPage() { if (page.value > 1) { page.value--; load() } }
function nextPage() { if (page.value < totalPages.value) { page.value++; load() } }

// ─── create ───────────────────────────────────────────────────────────────────
function openCreate() {
  createForm.value = {
    title: '', description: '', severity: 'minor',
    source_type: '', source_ref: '', inspection_id: '',
    item_id: '', lot_number: '', qty_affected: '',
    department_id: '', process: '',
    detected_by: '', detected_date: new Date().toISOString().split('T')[0],
    assigned_to: '', target_date: '',
    root_cause: '', immediate_action: '',
  }
  showCreate.value = true
}

async function submitCreate() {
  if (!createForm.value.title) { alert('Title is required.'); return }
  saving.value = true
  try {
    await qualityAPI.createNonConformity(createForm.value)
    showCreate.value = false
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error creating non-conformity')
  }
  saving.value = false
}

// ─── edit ─────────────────────────────────────────────────────────────────────
function openEdit(nc: any) {
  editTarget.value = nc
  editForm.value = { ...nc }
  showEdit.value = true
}

async function submitEdit() {
  saving.value = true
  try {
    await qualityAPI.updateNonConformity(editTarget.value.id, editForm.value)
    showEdit.value = false
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error updating')
  }
  saving.value = false
}

// ─── status ───────────────────────────────────────────────────────────────────
function openStatus(nc: any) {
  statusTarget.value = nc
  const transitions = statusTransitions[nc.status] || []
  statusForm.value = { status: transitions[0] || '', closure_notes: '' }
  showStatus.value = true
}

async function submitStatus() {
  if (!statusForm.value.status) { alert('Select a new status.'); return }
  saving.value = true
  try {
    await qualityAPI.updateNCStatus(statusTarget.value.id, statusForm.value)
    showStatus.value = false
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error updating status')
  }
  saving.value = false
}

// ─── delete ───────────────────────────────────────────────────────────────────
async function deleteNC(id: string) {
  if (!confirm('Delete this non-conformity? Only open NCs can be deleted.')) return
  try {
    await qualityAPI.deleteNonConformity(id)
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Error')
  }
}

// ─── helpers ─────────────────────────────────────────────────────────────────
function severityBadge(s: string) {
  const m: Record<string, string> = {
    minor: 'bg-slate-100 text-slate-700 dark:bg-slate-700 dark:text-slate-300',
    major: 'bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300',
    critical: 'bg-orange-100 text-orange-800 dark:bg-orange-900/40 dark:text-orange-300',
    critical_safety: 'bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300',
  }
  return m[s] || 'bg-slate-100 text-slate-600'
}

function severityIcon(s: string) {
  if (s === 'critical_safety') return ShieldAlert
  if (s === 'critical')        return Zap
  if (s === 'major')           return AlertTriangle
  return Info
}

function severityIconColor(s: string) {
  if (s === 'critical_safety') return 'text-red-500'
  if (s === 'critical')        return 'text-orange-500'
  if (s === 'major')           return 'text-amber-500'
  return 'text-slate-400'
}

function statusBadge(s: string) {
  const m: Record<string, string> = {
    open: 'bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300',
    under_review: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/40 dark:text-indigo-300',
    corrective_action: 'bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-300',
    closed: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300',
    cancelled: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-400',
  }
  return m[s] || 'bg-slate-100 text-slate-600'
}

function statusLabel(s: string) {
  const m: Record<string, string> = {
    open: 'Open', under_review: 'Under Review',
    corrective_action: 'Corrective Action',
    closed: 'Closed', cancelled: 'Cancelled',
  }
  return m[s] || s
}

function fmtDate(d: any) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString()
}

// KPI summary from current page
const severitySummary = computed(() => {
  const counts: Record<string, number> = { minor: 0, major: 0, critical: 0, critical_safety: 0 }
  ncs.value.forEach(nc => {
    if (nc.severity in counts) counts[nc.severity]++
  })
  return counts
})
</script>

<template>
  <div :class="dk('bg-slate-900 text-slate-100','bg-slate-50 text-slate-900')" class="min-h-screen">
    <div class="max-w-screen-2xl mx-auto px-6 py-6">

      <!-- Header -->
      <div class="flex items-center justify-between mb-6">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-orange-600 flex items-center justify-center">
            <AlertOctagon class="w-5 h-5 text-white" />
          </div>
          <div>
            <h1 class="text-2xl font-bold">Non-Conformities</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">
              {{ total }} non-conformit{{ total !== 1 ? 'ies' : 'y' }} total
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
            class="flex items-center gap-2 px-4 py-2 rounded-lg bg-orange-600 hover:bg-orange-700 text-white font-medium text-sm transition-colors">
            <Plus class="w-4 h-4" />
            New NC
          </button>
        </div>
      </div>

      <!-- Severity Summary Cards -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6">
        <div v-for="sev in [
          { key: 'critical_safety', label: 'Critical Safety', icon: ShieldAlert, color: 'red' },
          { key: 'critical',        label: 'Critical',        icon: Zap,         color: 'orange' },
          { key: 'major',           label: 'Major',           icon: AlertTriangle, color: 'amber' },
          { key: 'minor',           label: 'Minor',           icon: Info,        color: 'slate' },
        ]" :key="sev.key"
          :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="rounded-xl border p-4 flex items-center gap-3">
          <div :class="`bg-${sev.color}-500/10 text-${sev.color}-500`"
            class="w-9 h-9 rounded-lg flex items-center justify-center">
            <component :is="sev.icon" class="w-5 h-5" />
          </div>
          <div>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">{{ sev.label }}</p>
            <p class="text-xl font-bold" :class="`text-${sev.color}-500`">
              {{ severitySummary[sev.key] }}
            </p>
          </div>
        </div>
      </div>

      <!-- Filters -->
      <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
        class="rounded-xl border p-4 mb-5 flex flex-wrap items-center gap-3">
        <div class="relative flex-1 min-w-[200px]">
          <Search :class="dk('text-slate-400','text-slate-400')" class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" />
          <input v-model="search" @keyup.enter="applyFilter"
            :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-slate-50 border-slate-300 text-slate-900 placeholder-slate-400')"
            class="w-full pl-9 pr-3 py-2 text-sm border rounded-lg focus:ring-2 focus:ring-orange-500 outline-none"
            placeholder="Search reference, title..." />
        </div>
        <select v-model="filterStatus" @change="applyFilter"
          :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
          class="px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500">
          <option v-for="o in statusOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <select v-model="filterSeverity" @change="applyFilter"
          :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
          class="px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500">
          <option v-for="o in severityOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
        </select>
        <button @click="resetFilter"
          :class="dk('text-slate-400 hover:text-white border-slate-600','text-slate-500 hover:text-slate-700 border-slate-300')"
          class="px-3 py-2 text-sm border rounded-lg transition-colors">
          Reset
        </button>
      </div>

      <!-- NC List -->
      <div class="space-y-3">
        <div v-if="loading" class="flex items-center justify-center py-20">
          <Loader2 class="w-8 h-8 text-orange-500 animate-spin" />
        </div>
        <div v-else-if="ncs.length === 0"
          :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
          class="rounded-xl border flex flex-col items-center justify-center py-20 gap-3">
          <AlertOctagon :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12" />
          <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">No non-conformities found</p>
          <button @click="openCreate"
            class="mt-2 px-4 py-2 rounded-lg bg-orange-600 hover:bg-orange-700 text-white text-sm font-medium">
            Create First NC
          </button>
        </div>

        <template v-else>
          <!-- Table view -->
          <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
            class="rounded-xl border overflow-hidden">
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead>
                  <tr :class="dk('bg-slate-900/60 text-slate-400 border-b border-slate-700','bg-slate-50 text-slate-500 border-b border-slate-200')">
                    <th class="px-4 py-3 text-left font-medium">Severity</th>
                    <th class="px-4 py-3 text-left font-medium">Reference</th>
                    <th class="px-4 py-3 text-left font-medium">Title</th>
                    <th class="px-4 py-3 text-left font-medium">Status</th>
                    <th class="px-4 py-3 text-left font-medium">Item</th>
                    <th class="px-4 py-3 text-left font-medium">Department</th>
                    <th class="px-4 py-3 text-left font-medium">Detected</th>
                    <th class="px-4 py-3 text-left font-medium">Detected By</th>
                    <th class="px-4 py-3 text-left font-medium">Assigned To</th>
                    <th class="px-4 py-3 text-left font-medium">Target</th>
                    <th class="px-4 py-3 text-right font-medium">CAs</th>
                    <th class="px-4 py-3 text-center font-medium">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="nc in ncs" :key="nc.id"
                    :class="[
                      dk('border-b border-slate-700/60 hover:bg-slate-700/30','border-b border-slate-100 hover:bg-slate-50'),
                      nc.severity === 'critical_safety' ? dk('bg-red-900/10','bg-red-50/50') :
                      nc.severity === 'critical' ? dk('bg-orange-900/10','bg-orange-50/30') : ''
                    ]"
                    class="transition-colors">
                    <td class="px-4 py-3">
                      <div class="flex items-center gap-1.5">
                        <component :is="severityIcon(nc.severity)" :class="severityIconColor(nc.severity)" class="w-4 h-4" />
                        <span :class="severityBadge(nc.severity)"
                          class="px-2 py-0.5 rounded-full text-xs font-semibold capitalize">
                          {{ nc.severity.replace('_', ' ') }}
                        </span>
                      </div>
                    </td>
                    <td class="px-4 py-3 font-mono font-medium" :class="dk('text-orange-300','text-orange-600')">
                      {{ nc.reference }}
                    </td>
                    <td class="px-4 py-3 max-w-xs">
                      <div class="font-medium truncate" :class="dk('text-slate-100','text-slate-800')" :title="nc.title">
                        {{ nc.title }}
                      </div>
                      <div v-if="nc.source_type" class="text-xs mt-0.5" :class="dk('text-slate-400','text-slate-500')">
                        {{ nc.source_type.replace('_', ' ') }}
                        <span v-if="nc.source_ref"> · {{ nc.source_ref }}</span>
                      </div>
                    </td>
                    <td class="px-4 py-3">
                      <span :class="statusBadge(nc.status)"
                        class="px-2 py-0.5 rounded-full text-xs font-medium">
                        {{ statusLabel(nc.status) }}
                      </span>
                    </td>
                    <td class="px-4 py-3 text-xs" :class="dk('text-slate-300','text-slate-600')">
                      {{ nc.item_name || '—' }}
                      <div v-if="nc.lot_number" :class="dk('text-slate-500','text-slate-400')">Lot: {{ nc.lot_number }}</div>
                    </td>
                    <td class="px-4 py-3 text-xs" :class="dk('text-slate-300','text-slate-600')">
                      {{ nc.department_name || '—' }}
                    </td>
                    <td class="px-4 py-3 text-xs" :class="dk('text-slate-300','text-slate-600')">
                      {{ fmtDate(nc.detected_date) }}
                    </td>
                    <td class="px-4 py-3 text-xs" :class="dk('text-slate-300','text-slate-600')">
                      {{ nc.detected_by_name || '—' }}
                    </td>
                    <td class="px-4 py-3 text-xs" :class="dk('text-slate-300','text-slate-600')">
                      {{ nc.assigned_to_name || '—' }}
                    </td>
                    <td class="px-4 py-3 text-xs"
                      :class="nc.target_date && new Date(nc.target_date) < new Date() && nc.status !== 'closed'
                        ? 'text-red-500 font-semibold' : dk('text-slate-300','text-slate-600')">
                      {{ fmtDate(nc.target_date) }}
                    </td>
                    <td class="px-4 py-3 text-right">
                      <span v-if="nc.ca_count > 0"
                        class="px-2 py-0.5 rounded-full text-xs bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300 font-medium">
                        {{ nc.ca_count }}
                      </span>
                      <span v-else :class="dk('text-slate-600','text-slate-400')" class="text-xs">0</span>
                    </td>
                    <td class="px-4 py-3">
                      <div class="flex items-center justify-center gap-1">
                        <!-- Status change -->
                        <button v-if="(statusTransitions[nc.status] || []).length > 0"
                          @click="openStatus(nc)" title="Change Status"
                          class="p-1.5 rounded-lg bg-blue-600/20 hover:bg-blue-600/40 text-blue-400 hover:text-blue-300 transition-colors">
                          <ArrowRight class="w-3.5 h-3.5" />
                        </button>
                        <!-- Edit -->
                        <button @click="openEdit(nc)" title="Edit"
                          class="p-1.5 rounded-lg bg-amber-600/20 hover:bg-amber-600/40 text-amber-400 hover:text-amber-300 transition-colors">
                          <Pencil class="w-3.5 h-3.5" />
                        </button>
                        <!-- Delete (open only) -->
                        <button v-if="nc.status === 'open'"
                          @click="deleteNC(nc.id)" title="Delete"
                          class="p-1.5 rounded-lg bg-red-600/20 hover:bg-red-600/40 text-red-400 hover:text-red-300 transition-colors">
                          <Trash2 class="w-3.5 h-3.5" />
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
          <div :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
            class="flex items-center justify-between px-6 py-4 border-b sticky top-0">
            <h2 class="text-lg font-semibold" :class="dk('text-white','text-slate-800')">New Non-Conformity</h2>
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
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="Non-conformity title" />
            </div>
            <!-- Description -->
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Description</label>
              <textarea v-model="createForm.description" rows="2"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500 resize-none"
                placeholder="Detailed description..." />
            </div>
            <!-- Severity -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">
                Severity <span class="text-red-400">*</span>
              </label>
              <select v-model="createForm.severity"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500">
                <option value="minor">Minor</option>
                <option value="major">Major</option>
                <option value="critical">Critical</option>
                <option value="critical_safety">Critical Safety</option>
              </select>
            </div>
            <!-- Source Type -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Source Type</label>
              <select v-model="createForm.source_type"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500">
                <option value="">— None —</option>
                <option value="inspection">Inspection</option>
                <option value="customer_complaint">Customer Complaint</option>
                <option value="internal_audit">Internal Audit</option>
                <option value="supplier">Supplier</option>
                <option value="production">Production</option>
              </select>
            </div>
            <!-- Source Ref -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Source Reference</label>
              <input v-model="createForm.source_ref"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="QI-2024-00001" />
            </div>
            <!-- Item ID -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Item ID</label>
              <input v-model="createForm.item_id"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="Item UUID" />
            </div>
            <!-- Lot Number -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Lot Number</label>
              <input v-model="createForm.lot_number"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="LOT-001" />
            </div>
            <!-- Qty Affected -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Qty Affected</label>
              <input v-model="createForm.qty_affected" type="number" min="0"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500" />
            </div>
            <!-- Department ID -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Department ID</label>
              <input v-model="createForm.department_id"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="Department UUID" />
            </div>
            <!-- Process -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Process</label>
              <input v-model="createForm.process"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="Assembly line, Welding..." />
            </div>
            <!-- Detected By -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Detected By (User ID)</label>
              <input v-model="createForm.detected_by"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="User UUID" />
            </div>
            <!-- Detected Date -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Detected Date</label>
              <input v-model="createForm.detected_date" type="date"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500" />
            </div>
            <!-- Assigned To -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Assigned To (User ID)</label>
              <input v-model="createForm.assigned_to"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500"
                placeholder="User UUID" />
            </div>
            <!-- Target Date -->
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Target Resolution Date</label>
              <input v-model="createForm.target_date" type="date"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500" />
            </div>
            <!-- Immediate Action -->
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Immediate Action</label>
              <textarea v-model="createForm.immediate_action" rows="2"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500 resize-none"
                placeholder="Containment action taken immediately..." />
            </div>
            <!-- Root Cause -->
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Root Cause (if known)</label>
              <textarea v-model="createForm.root_cause" rows="2"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500 resize-none"
                placeholder="Root cause analysis..." />
            </div>
          </div>
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showCreate = false"
              :class="dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
              class="px-4 py-2 text-sm border rounded-lg transition-colors">Cancel</button>
            <button @click="submitCreate" :disabled="saving"
              class="flex items-center gap-2 px-4 py-2 text-sm bg-orange-600 hover:bg-orange-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              Create NC
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
              Edit NC — {{ editTarget?.reference }}
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
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Severity</label>
              <select v-model="editForm.severity"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500">
                <option value="minor">Minor</option>
                <option value="major">Major</option>
                <option value="critical">Critical</option>
                <option value="critical_safety">Critical Safety</option>
              </select>
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Assigned To (User ID)</label>
              <input v-model="editForm.assigned_to"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Target Date</label>
              <input v-model="editForm.target_date" type="date"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500" />
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Process</label>
              <input v-model="editForm.process"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500" />
            </div>
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Root Cause</label>
              <textarea v-model="editForm.root_cause" rows="2"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500 resize-none" />
            </div>
            <div class="col-span-2">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">Immediate Action</label>
              <textarea v-model="editForm.immediate_action" rows="2"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500 resize-none" />
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
          class="relative w-full max-w-md rounded-2xl border shadow-2xl">
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-between px-6 py-4 border-b">
            <h2 class="text-lg font-semibold" :class="dk('text-white','text-slate-800')">Update NC Status</h2>
            <button @click="showStatus = false"
              :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-700')">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5 space-y-4">
            <div class="flex items-center gap-2 text-sm" :class="dk('text-slate-300','text-slate-600')">
              Current:
              <span :class="statusBadge(statusTarget?.status)"
                class="px-2 py-0.5 rounded-full text-xs font-medium">
                {{ statusLabel(statusTarget?.status) }}
              </span>
            </div>
            <div>
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-2">New Status</label>
              <div class="flex flex-wrap gap-2">
                <button v-for="s in (statusTransitions[statusTarget?.status] || [])" :key="s"
                  @click="statusForm.status = s"
                  :class="statusForm.status === s
                    ? 'bg-indigo-600 border-indigo-600 text-white'
                    : dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
                  class="px-3 py-1.5 text-sm border rounded-lg transition-colors">
                  {{ statusLabel(s) }}
                </button>
              </div>
            </div>
            <div v-if="statusForm.status === 'closed'">
              <label :class="dk('text-slate-300','text-slate-600')" class="block text-xs font-medium mb-1">
                Closure Notes <span class="text-red-400">*</span>
              </label>
              <textarea v-model="statusForm.closure_notes" rows="3"
                :class="dk('bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400','bg-white border-slate-300 text-slate-700')"
                class="w-full px-3 py-2 text-sm border rounded-lg outline-none focus:ring-2 focus:ring-orange-500 resize-none"
                placeholder="Describe how and why the NC was closed..." />
            </div>
          </div>
          <div :class="dk('border-slate-700','border-slate-200')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showStatus = false"
              :class="dk('border-slate-600 text-slate-300 hover:bg-slate-700','border-slate-300 text-slate-600 hover:bg-slate-50')"
              class="px-4 py-2 text-sm border rounded-lg transition-colors">Cancel</button>
            <button @click="submitStatus" :disabled="saving || !statusForm.status"
              class="flex items-center gap-2 px-4 py-2 text-sm bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              Update Status
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
