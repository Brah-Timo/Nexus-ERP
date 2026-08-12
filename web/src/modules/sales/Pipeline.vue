<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { salesAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  Plus, X, Save, RefreshCw, Target, TrendingUp, DollarSign,
  Users, ChevronRight, BarChart3, Percent, Edit2, Trash2,
  CircleDot, CheckCircle2, XCircle, MessageSquare, Calendar,
  User, Building2, MoveRight
} from 'lucide-vue-next'

// ─── Types ─────────────────────────────────────────────────────────────────

interface Opportunity {
  id: string
  company_id: string
  customer_id?: string
  customer_name?: string
  lead_id?: string
  name: string
  stage: string
  amount: number
  probability: number
  expected_close?: string
  salesperson_id?: string
  notes: string
  lost_reason: string
  created_at: string
  updated_at: string
}

interface PipelineSummary {
  stage: string
  count: number
  total_amount: number
  avg_probability: number
}

interface Customer { id: string; name: string; code: string }

// ─── Stage Config ──────────────────────────────────────────────────────────

const STAGES = [
  { key: 'lead',        label: 'Lead',         color: 'gray',   bgClass: 'bg-gray-100 dark:bg-gray-800',   textClass: 'text-gray-700 dark:text-gray-300',   borderClass: 'border-gray-300 dark:border-gray-600' },
  { key: 'qualified',   label: 'Qualified',    color: 'blue',   bgClass: 'bg-blue-50 dark:bg-blue-900/20', textClass: 'text-blue-700 dark:text-blue-300',    borderClass: 'border-blue-300 dark:border-blue-700' },
  { key: 'proposal',    label: 'Proposal',     color: 'indigo', bgClass: 'bg-indigo-50 dark:bg-indigo-900/20',textClass: 'text-indigo-700 dark:text-indigo-300',borderClass: 'border-indigo-300 dark:border-indigo-700' },
  { key: 'negotiation', label: 'Negotiation',  color: 'amber',  bgClass: 'bg-amber-50 dark:bg-amber-900/20',textClass: 'text-amber-700 dark:text-amber-300',  borderClass: 'border-amber-300 dark:border-amber-700' },
  { key: 'won',         label: 'Won',          color: 'green',  bgClass: 'bg-green-50 dark:bg-green-900/20',textClass: 'text-green-700 dark:text-green-300',  borderClass: 'border-green-300 dark:border-green-700' },
  { key: 'lost',        label: 'Lost',         color: 'red',    bgClass: 'bg-red-50 dark:bg-red-900/20',   textClass: 'text-red-700 dark:text-red-300',     borderClass: 'border-red-300 dark:border-red-700' },
]

const stageBadge: Record<string, string> = {
  lead:        'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  qualified:   'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300',
  proposal:    'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/40 dark:text-indigo-300',
  negotiation: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300',
  won:         'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300',
  lost:        'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300',
}

// ─── State ─────────────────────────────────────────────────────────────────

const app = useAppStore()
const opportunities = ref<Opportunity[]>([])
const pipelineSummary = ref<PipelineSummary[]>([])
const customers = ref<Customer[]>([])
const loading = ref(true)
const saving = ref(false)

// Modal state
const showModal = ref(false)
const editingOpp = ref<Partial<Opportunity> | null>(null)
const isEdit = ref(false)

// Drag-and-drop
const dragging = ref<string | null>(null)
const dragOver = ref<string | null>(null)

// ─── Computed ──────────────────────────────────────────────────────────────

const oppsByStage = computed(() => {
  const map: Record<string, Opportunity[]> = {}
  STAGES.forEach(s => { map[s.key] = [] })
  opportunities.value.forEach(o => {
    if (map[o.stage]) map[o.stage].push(o)
  })
  return map
})

const totalPipeline = computed(() =>
  pipelineSummary.value
    .filter(s => !['won', 'lost'].includes(s.stage))
    .reduce((sum, s) => sum + s.total_amount, 0)
)

const summaryMap = computed(() => {
  const m: Record<string, PipelineSummary> = {}
  pipelineSummary.value.forEach(s => { m[s.stage] = s })
  return m
})

// ─── Data Loading ──────────────────────────────────────────────────────────

async function loadData() {
  loading.value = true
  try {
    const [oppRes, summRes, custRes] = await Promise.allSettled([
      salesAPI.getOpportunities(),
      salesAPI.getPipelineSummary(),
      salesAPI.getCustomers(),
    ])
    if (oppRes.status === 'fulfilled') opportunities.value = oppRes.value.data || []
    if (summRes.status === 'fulfilled') pipelineSummary.value = summRes.value.data || []
    if (custRes.status === 'fulfilled') customers.value = custRes.value.data || []
  } catch {
    app.addToast('Failed to load pipeline', 'error')
  } finally {
    loading.value = false
  }
}

// ─── Modal ─────────────────────────────────────────────────────────────────

function openCreate(stage = 'lead') {
  isEdit.value = false
  editingOpp.value = {
    name: '', stage, amount: 0, probability: 10, notes: '', lost_reason: '',
    customer_id: undefined, expected_close: undefined,
  }
  showModal.value = true
}

function openEdit(opp: Opportunity) {
  isEdit.value = true
  editingOpp.value = { ...opp }
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editingOpp.value = null
}

async function saveOpp() {
  if (!editingOpp.value?.name?.trim()) {
    app.addToast('Opportunity name is required', 'error'); return
  }
  saving.value = true
  try {
    if (isEdit.value && editingOpp.value?.id) {
      await salesAPI.updateOpportunity(editingOpp.value.id, editingOpp.value)
      app.addToast('Opportunity updated', 'success')
    } else {
      await salesAPI.createOpportunity(editingOpp.value)
      app.addToast('Opportunity created', 'success')
    }
    closeModal()
    await loadData()
  } catch {
    app.addToast('Failed to save opportunity', 'error')
  } finally {
    saving.value = false
  }
}

async function deleteOpp(opp: Opportunity) {
  if (!confirm(`Delete "${opp.name}"?`)) return
  try {
    await salesAPI.deleteOpportunity(opp.id)
    app.addToast('Opportunity deleted', 'success')
    await loadData()
  } catch {
    app.addToast('Failed to delete opportunity', 'error')
  }
}

// ─── Drag & Drop ───────────────────────────────────────────────────────────

function onDragStart(opp: Opportunity) {
  dragging.value = opp.id
}

function onDragOver(e: DragEvent, stageKey: string) {
  e.preventDefault()
  dragOver.value = stageKey
}

function onDragLeave() {
  dragOver.value = null
}

async function onDrop(stageKey: string) {
  if (!dragging.value || !dragOver.value) return
  const opp = opportunities.value.find(o => o.id === dragging.value)
  if (!opp || opp.stage === stageKey) { dragging.value = null; dragOver.value = null; return }
  // Optimistic update
  opp.stage = stageKey
  try {
    await salesAPI.updateOpportunity(opp.id, { ...opp, stage: stageKey })
    app.addToast(`Moved to ${STAGES.find(s => s.key === stageKey)?.label}`, 'success')
    await loadData()
  } catch {
    app.addToast('Failed to move opportunity', 'error')
    await loadData()
  } finally {
    dragging.value = null
    dragOver.value = null
  }
}

// ─── Formatters ────────────────────────────────────────────────────────────

function fmtCurrency(n: number): string {
  if (Math.abs(n) >= 1_000_000) return (n / 1_000_000).toFixed(2) + ' M'
  if (Math.abs(n) >= 1_000) return (n / 1_000).toFixed(1) + ' k'
  return n.toLocaleString('fr-DZ') + ' DZD'
}

function fmtDate(d?: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('fr-DZ')
}

onMounted(loadData)
</script>

<template>
  <div class="space-y-5">

    <!-- ─── Header ──────────────────────────────────────────────────────────── -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Sales Pipeline</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Drag cards to advance stage · click to edit</p>
      </div>
      <div class="flex items-center gap-2">
        <button
          @click="loadData"
          :disabled="loading"
          class="p-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors disabled:opacity-50"
        >
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button
          @click="openCreate()"
          class="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg shadow-sm transition-colors"
        >
          <Plus class="w-4 h-4" />
          New Opportunity
        </button>
      </div>
    </div>

    <!-- ─── Pipeline KPI bar ───────────────────────────────────────────────── -->
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
      <div
        v-for="stage in STAGES"
        :key="stage.key"
        class="rounded-xl border p-3 bg-white dark:bg-gray-900 shadow-sm"
        :class="stage.borderClass"
      >
        <div class="flex items-center justify-between mb-1">
          <span class="text-xs font-semibold" :class="stage.textClass">{{ stage.label }}</span>
          <span class="text-xs font-bold text-gray-700 dark:text-gray-300">
            {{ summaryMap[stage.key]?.count ?? 0 }}
          </span>
        </div>
        <p class="text-sm font-bold text-gray-900 dark:text-white truncate">
          {{ summaryMap[stage.key] ? fmtCurrency(summaryMap[stage.key].total_amount) : '0 DZD' }}
        </p>
        <div class="flex items-center gap-1 mt-1">
          <div class="flex-1 h-1 rounded-full bg-gray-200 dark:bg-gray-700">
            <div
              class="h-1 rounded-full transition-all duration-500"
              :class="stage.bgClass.replace('bg-', 'bg-').replace('dark:bg-', 'dark:bg-')"
              :style="{ width: (summaryMap[stage.key]?.avg_probability ?? 0) + '%' }"
            />
          </div>
          <span class="text-xs text-gray-400 dark:text-gray-500">{{ summaryMap[stage.key]?.avg_probability ?? 0 }}%</span>
        </div>
      </div>
    </div>

    <!-- ─── Total Pipeline bar ────────────────────────────────────────────── -->
    <div class="flex items-center gap-3 px-4 py-3 rounded-xl bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800">
      <BarChart3 class="w-5 h-5 text-blue-600 dark:text-blue-400 flex-shrink-0" />
      <div class="flex items-center gap-2">
        <span class="text-sm text-blue-700 dark:text-blue-300 font-medium">Active Pipeline:</span>
        <span class="text-sm font-bold text-blue-800 dark:text-blue-200">{{ fmtCurrency(totalPipeline) }}</span>
      </div>
      <div class="ml-auto text-xs text-blue-600 dark:text-blue-400">
        {{ opportunities.filter(o => !['won','lost'].includes(o.stage)).length }} open opportunities
      </div>
    </div>

    <!-- ─── Loading ───────────────────────────────────────────────────────── -->
    <div v-if="loading" class="flex items-center justify-center h-64">
      <RefreshCw class="w-8 h-8 text-gray-400 animate-spin" />
    </div>

    <!-- ─── Kanban Board ───────────────────────────────────────────────────── -->
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4 items-start">

      <div
        v-for="stage in STAGES"
        :key="stage.key"
        class="flex flex-col gap-2 min-h-32"
        @dragover.prevent="onDragOver($event, stage.key)"
        @dragleave="onDragLeave"
        @drop="onDrop(stage.key)"
      >
        <!-- Column Header -->
        <div
          class="flex items-center justify-between px-3 py-2 rounded-lg border font-semibold text-sm"
          :class="[stage.bgClass, stage.borderClass, stage.textClass]"
        >
          <span>{{ stage.label }}</span>
          <div class="flex items-center gap-1.5">
            <span class="text-xs font-bold px-1.5 py-0.5 rounded-full bg-white/60 dark:bg-black/20">
              {{ oppsByStage[stage.key]?.length ?? 0 }}
            </span>
            <button
              @click="openCreate(stage.key)"
              class="opacity-70 hover:opacity-100 transition-opacity"
            >
              <Plus class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- Drop zone highlight -->
        <div
          v-if="dragOver === stage.key"
          class="h-2 rounded-full bg-blue-400 dark:bg-blue-500 transition-all"
        />

        <!-- Cards -->
        <div
          v-for="opp in oppsByStage[stage.key]"
          :key="opp.id"
          draggable="true"
          @dragstart="onDragStart(opp)"
          class="relative group flex flex-col gap-2 p-3 rounded-lg border border-gray-200 dark:border-gray-700
                 bg-white dark:bg-gray-900 shadow-sm hover:shadow-md cursor-grab active:cursor-grabbing
                 transition-all duration-150 select-none"
          :class="{ 'opacity-50 scale-95': dragging === opp.id }"
        >
          <!-- Card actions -->
          <div class="absolute top-2 right-2 hidden group-hover:flex items-center gap-1 z-10">
            <button
              @click.stop="openEdit(opp)"
              class="p-1 rounded bg-blue-50 dark:bg-blue-900/40 text-blue-600 dark:text-blue-400 hover:bg-blue-100 dark:hover:bg-blue-900/60 transition-colors"
            >
              <Edit2 class="w-3 h-3" />
            </button>
            <button
              @click.stop="deleteOpp(opp)"
              class="p-1 rounded bg-red-50 dark:bg-red-900/40 text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/60 transition-colors"
            >
              <Trash2 class="w-3 h-3" />
            </button>
          </div>

          <!-- Name -->
          <p class="text-sm font-semibold text-gray-900 dark:text-white pr-12 leading-tight">{{ opp.name }}</p>

          <!-- Customer -->
          <div v-if="opp.customer_name" class="flex items-center gap-1 text-xs text-gray-500 dark:text-gray-400">
            <Building2 class="w-3 h-3 flex-shrink-0" />
            <span class="truncate">{{ opp.customer_name }}</span>
          </div>

          <!-- Amount + Probability -->
          <div class="flex items-center justify-between gap-2">
            <span class="text-sm font-bold text-gray-800 dark:text-gray-200">{{ fmtCurrency(opp.amount) }}</span>
            <span
              class="text-xs px-1.5 py-0.5 rounded-full font-medium"
              :class="opp.probability >= 75 ? 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-400' :
                       opp.probability >= 50 ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-400' :
                                               'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'"
            >
              {{ opp.probability }}%
            </span>
          </div>

          <!-- Expected close -->
          <div v-if="opp.expected_close" class="flex items-center gap-1 text-xs text-gray-400 dark:text-gray-500">
            <Calendar class="w-3 h-3 flex-shrink-0" />
            <span>{{ fmtDate(opp.expected_close) }}</span>
          </div>

          <!-- Notes -->
          <div v-if="opp.notes" class="flex items-start gap-1 text-xs text-gray-400 dark:text-gray-500">
            <MessageSquare class="w-3 h-3 flex-shrink-0 mt-0.5" />
            <span class="line-clamp-2">{{ opp.notes }}</span>
          </div>

          <!-- Move arrows -->
          <div class="flex justify-end gap-1 pt-1 border-t border-gray-100 dark:border-gray-800">
            <button
              v-for="s in STAGES.filter(s => s.key !== opp.stage)"
              :key="s.key"
              @click.stop="salesAPI.updateOpportunity(opp.id, { ...opp, stage: s.key }).then(() => loadData())"
              class="text-xs px-1.5 py-0.5 rounded hover:opacity-80 transition-opacity"
              :class="stageBadge[s.key]"
              :title="'Move to ' + s.label"
            >
              {{ s.label[0] }}
            </button>
          </div>
        </div>

        <!-- Empty column hint -->
        <div
          v-if="oppsByStage[stage.key]?.length === 0"
          class="flex items-center justify-center h-16 border-2 border-dashed border-gray-200 dark:border-gray-700 rounded-lg text-xs text-gray-400 dark:text-gray-600"
        >
          Drop here
        </div>

      </div>
    </div>

    <!-- ─── Create / Edit Modal ─────────────────────────────────────────────── -->
    <Teleport to="body">
      <div
        v-if="showModal"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm"
        @click.self="closeModal"
      >
        <div class="w-full max-w-lg bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 flex flex-col max-h-[90vh]">

          <!-- Modal header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div class="flex items-center gap-2">
              <Target class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ isEdit ? 'Edit Opportunity' : 'New Opportunity' }}
              </h2>
            </div>
            <button @click="closeModal" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Modal body -->
          <div class="flex-1 overflow-y-auto p-6 space-y-4">
            <div v-if="editingOpp">
              <!-- Name -->
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">
                  Opportunity Name <span class="text-red-500">*</span>
                </label>
                <input
                  v-model="editingOpp.name"
                  class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500 outline-none"
                  placeholder="e.g. Software deployment at SONATRACH"
                />
              </div>

              <!-- Customer -->
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Customer</label>
                <select
                  v-model="editingOpp.customer_id"
                  class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500 outline-none"
                >
                  <option value="">— None —</option>
                  <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
                </select>
              </div>

              <!-- Stage + Probability -->
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Stage</label>
                  <select
                    v-model="editingOpp.stage"
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500 outline-none"
                  >
                    <option v-for="s in STAGES" :key="s.key" :value="s.key">{{ s.label }}</option>
                  </select>
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">
                    Probability (%)
                  </label>
                  <input
                    v-model.number="editingOpp.probability"
                    type="number" min="0" max="100"
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500 outline-none"
                  />
                </div>
              </div>

              <!-- Amount + Expected Close -->
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Amount (DZD)</label>
                  <input
                    v-model.number="editingOpp.amount"
                    type="number" min="0" step="1000"
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500 outline-none"
                  />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Expected Close</label>
                  <input
                    v-model="editingOpp.expected_close"
                    type="date"
                    class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500 outline-none"
                  />
                </div>
              </div>

              <!-- Notes -->
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Notes</label>
                <textarea
                  v-model="editingOpp.notes"
                  rows="3"
                  class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500 outline-none resize-none"
                  placeholder="Additional notes..."
                />
              </div>

              <!-- Lost reason (only if stage = lost) -->
              <div v-if="editingOpp.stage === 'lost'">
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Lost Reason</label>
                <textarea
                  v-model="editingOpp.lost_reason"
                  rows="2"
                  class="w-full px-3 py-2 rounded-lg border border-red-300 dark:border-red-700 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-red-500 outline-none resize-none"
                  placeholder="Why was this opportunity lost?"
                />
              </div>
            </div>
          </div>

          <!-- Modal footer -->
          <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
            <button
              @click="closeModal"
              class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="saveOpp"
              :disabled="saving"
              class="inline-flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg shadow-sm transition-colors disabled:opacity-60"
            >
              <Save class="w-4 h-4" />
              {{ saving ? 'Saving…' : (isEdit ? 'Update' : 'Create') }}
            </button>
          </div>

        </div>
      </div>
    </Teleport>

  </div>
</template>
