<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { settingsAPI } from '@/api/client'
import {
  GitBranch, Plus, Edit2, X, Save, RefreshCw, Trash2, Play,
  ChevronDown, ChevronRight, CheckCircle, XCircle, AlertTriangle,
  ArrowRight, Settings2, Zap
} from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(false)
const saving = ref(false)
const rules = ref<any[]>([])
const showModal = ref(false)
const editingId = ref('')
const expanded = ref<Set<string>>(new Set())

const docTypes = [
  { value: 'sales_invoice',      label: 'Sales Invoice' },
  { value: 'sales_order',        label: 'Sales Order' },
  { value: 'quotation',          label: 'Quotation' },
  { value: 'purchase_order',     label: 'Purchase Order' },
  { value: 'purchase_invoice',   label: 'Purchase Invoice' },
  { value: 'goods_receipt',      label: 'Goods Receipt' },
  { value: 'payment',            label: 'Payment' },
  { value: 'journal_entry',      label: 'Journal Entry' },
  { value: 'manufacturing_order','label': 'Manufacturing Order' },
  { value: 'expense_report',     label: 'Expense Report' },
]

const triggerEvents = [
  { value: 'on_create',   label: 'On Create' },
  { value: 'on_update',   label: 'On Update' },
  { value: 'on_approve',  label: 'On Approval' },
  { value: 'on_post',     label: 'On Post' },
  { value: 'on_cancel',   label: 'On Cancel' },
  { value: 'on_validate', label: 'On Validate' },
]

const conditionFields = ['amount', 'total', 'status', 'customer_type', 'department', 'cost_center']
const conditionOps = ['>', '<', '>=', '<=', '=', '!=', 'contains']
const actionTypes = [
  { value: 'send_email',       label: 'Send Email' },
  { value: 'notify_user',      label: 'Notify User' },
  { value: 'require_approval', label: 'Require Approval' },
  { value: 'set_status',       label: 'Set Status' },
  { value: 'assign_to',        label: 'Assign To' },
  { value: 'webhook',          label: 'Webhook' },
]

interface Condition { field: string; op: string; value: string }
interface Action { type: string; value: string; label?: string }

const form = ref({
  name: '',
  doc_type: 'sales_invoice',
  trigger_event: 'on_create',
  conditions: [] as Condition[],
  actions: [] as Action[],
  priority: 10,
  is_active: true,
})

function addCondition() {
  form.value.conditions.push({ field: 'amount', op: '>', value: '' })
}
function removeCondition(i: number) {
  form.value.conditions.splice(i, 1)
}
function addAction() {
  form.value.actions.push({ type: 'send_email', value: '', label: '' })
}
function removeAction(i: number) {
  form.value.actions.splice(i, 1)
}

function toggleExpand(id: string) {
  if (expanded.value.has(id)) expanded.value.delete(id)
  else expanded.value.add(id)
}

async function load() {
  loading.value = true
  try {
    const r = await settingsAPI.getWorkflowRules()
    rules.value = Array.isArray(r.data) ? r.data : []
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error loading workflow rules', 'error')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  form.value = {
    name: '', doc_type: 'sales_invoice', trigger_event: 'on_create',
    conditions: [], actions: [], priority: 10, is_active: true,
  }
  showModal.value = true
}

function openEdit(r: any) {
  editingId.value = r.id
  form.value = {
    name: r.name, doc_type: r.doc_type || 'sales_invoice',
    trigger_event: r.trigger_event || 'on_create',
    conditions: Array.isArray(r.conditions) ? JSON.parse(JSON.stringify(r.conditions)) : [],
    actions: Array.isArray(r.actions) ? JSON.parse(JSON.stringify(r.actions)) : [],
    priority: r.priority || 10,
    is_active: r.is_active !== false,
  }
  showModal.value = true
}

async function save() {
  if (!form.value.name) { app.addToast('Name is required', 'error'); return }
  saving.value = true
  try {
    if (editingId.value) {
      await settingsAPI.updateWorkflowRule(editingId.value, form.value)
      app.addToast('Workflow rule updated', 'success')
    } else {
      await settingsAPI.createWorkflowRule(form.value)
      app.addToast('Workflow rule created', 'success')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error saving rule', 'error')
  } finally {
    saving.value = false
  }
}

async function deleteRule(r: any) {
  if (!confirm(`Delete workflow rule "${r.name}"?`)) return
  try {
    await settingsAPI.deleteWorkflowRule(r.id)
    app.addToast('Rule deleted', 'success')
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error', 'error')
  }
}

async function toggleActive(r: any) {
  try {
    await settingsAPI.updateWorkflowRule(r.id, { ...r, is_active: !r.is_active })
    await load()
  } catch (e: any) {
    app.addToast('Error toggling rule', 'error')
  }
}

function docTypeLabel(v: string) {
  return docTypes.find(d => d.value === v)?.label || v
}
function triggerLabel(v: string) {
  return triggerEvents.find(e => e.value === v)?.label || v
}
function actionLabel(v: string) {
  return actionTypes.find(a => a.value === v)?.label || v
}

const byDocType = computed(() => {
  const g: Record<string, any[]> = {}
  for (const r of rules.value) {
    const k = r.doc_type || 'general'
    if (!g[k]) g[k] = []
    g[k].push(r)
  }
  return g
})

const dk = (d: string, l: string) => app.darkMode ? d : l
onMounted(load)
</script>

<template>
  <div :class="dk('bg-slate-950 text-slate-100','bg-slate-50 text-slate-900')" class="min-h-screen">
    <!-- Header -->
    <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border-b px-6 py-4 sticky top-0 z-10">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-violet-600/20">
            <GitBranch class="w-5 h-5 text-violet-400" />
          </div>
          <div>
            <h1 class="text-lg font-bold">Workflow Rules</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">Automate business processes with conditional rules</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
            class="p-2 rounded-lg transition-colors disabled:opacity-50">
            <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          </button>
          <button @click="openCreate"
            class="flex items-center gap-2 px-4 py-2 bg-violet-600 hover:bg-violet-500 text-white rounded-lg text-sm font-medium transition-colors">
            <Plus class="w-4 h-4" /> New Rule
          </button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="p-6">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="animate-spin w-8 h-8 text-violet-400" />
      </div>

      <template v-else>
        <!-- Stats -->
        <div class="grid grid-cols-3 gap-4 mb-6">
          <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl p-4">
            <div class="text-2xl font-bold text-violet-400">{{ rules.length }}</div>
            <div :class="dk('text-slate-500','text-slate-400')" class="text-xs mt-1">Total Rules</div>
          </div>
          <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl p-4">
            <div class="text-2xl font-bold text-emerald-400">{{ rules.filter(r=>r.is_active).length }}</div>
            <div :class="dk('text-slate-500','text-slate-400')" class="text-xs mt-1">Active</div>
          </div>
          <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl p-4">
            <div class="text-2xl font-bold text-sky-400">{{ Object.keys(byDocType).length }}</div>
            <div :class="dk('text-slate-500','text-slate-400')" class="text-xs mt-1">Document Types</div>
          </div>
        </div>

        <div v-if="rules.length === 0" :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
          class="border rounded-xl p-16 text-center">
          <GitBranch :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12 mx-auto mb-4" />
          <p :class="dk('text-slate-400','text-slate-500')" class="text-sm mb-4">No workflow rules defined</p>
          <button @click="openCreate" class="px-4 py-2 bg-violet-600 hover:bg-violet-500 text-white rounded-lg text-sm font-medium transition-colors">
            Create First Rule
          </button>
        </div>

        <!-- Rules grouped by doc type -->
        <div v-else class="space-y-4">
          <div v-for="(groupRules, docType) in byDocType" :key="docType"
            :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
            class="border rounded-xl overflow-hidden">
            <!-- Group header -->
            <div :class="dk('bg-slate-800/40 border-slate-800','bg-slate-50 border-slate-100')"
              class="border-b px-5 py-3 flex items-center gap-2">
              <div class="p-1.5 rounded-lg bg-violet-500/10">
                <GitBranch class="w-3.5 h-3.5 text-violet-400" />
              </div>
              <span class="font-semibold text-sm">{{ docTypeLabel(docType) }}</span>
              <span :class="dk('bg-slate-700 text-slate-400','bg-slate-200 text-slate-500')"
                class="text-xs px-2 py-0.5 rounded-full ml-1">{{ groupRules.length }}</span>
            </div>

            <div class="divide-y" :class="dk('divide-slate-800','divide-slate-100')">
              <div v-for="rule in groupRules" :key="rule.id">
                <!-- Rule row -->
                <div @click="toggleExpand(rule.id)"
                  :class="dk('hover:bg-slate-800/40','hover:bg-slate-50')"
                  class="flex items-center gap-4 px-5 py-4 cursor-pointer transition-colors">
                  <div :class="rule.is_active ? 'bg-emerald-500/10' : 'bg-red-500/10'" class="p-1.5 rounded-lg flex-shrink-0">
                    <Zap :class="rule.is_active ? 'text-emerald-400' : 'text-red-400'" class="w-3.5 h-3.5" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2 mb-0.5">
                      <span class="font-semibold text-sm">{{ rule.name }}</span>
                      <span :class="dk('bg-slate-800 text-slate-400','bg-slate-100 text-slate-500')"
                        class="text-xs px-1.5 py-0.5 rounded font-medium">P{{ rule.priority || 10 }}</span>
                    </div>
                    <div class="flex items-center gap-2 text-xs" :class="dk('text-slate-500','text-slate-400')">
                      <span class="text-sky-400 font-medium">{{ triggerLabel(rule.trigger_event) }}</span>
                      <ArrowRight class="w-3 h-3" />
                      <span>{{ Array.isArray(rule.conditions) ? rule.conditions.length : 0 }} conditions</span>
                      <ArrowRight class="w-3 h-3" />
                      <span>{{ Array.isArray(rule.actions) ? rule.actions.length : 0 }} actions</span>
                    </div>
                  </div>
                  <div class="flex items-center gap-2 flex-shrink-0">
                    <button @click.stop="toggleActive(rule)"
                      :class="rule.is_active ? 'bg-emerald-500' : 'bg-slate-600'"
                      class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors">
                      <span :class="rule.is_active ? 'translate-x-4.5' : 'translate-x-1'"
                        class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform" />
                    </button>
                    <button @click.stop="openEdit(rule)"
                      :class="dk('hover:bg-slate-700 text-slate-400 hover:text-white','hover:bg-slate-100 text-slate-500 hover:text-slate-900')"
                      class="p-1.5 rounded-lg transition-colors">
                      <Edit2 class="w-3.5 h-3.5" />
                    </button>
                    <button @click.stop="deleteRule(rule)"
                      class="p-1.5 rounded-lg hover:bg-red-500/10 text-red-400 transition-colors">
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                    <ChevronDown v-if="!expanded.has(rule.id)" :class="dk('text-slate-600','text-slate-300')" class="w-4 h-4" />
                    <ChevronRight v-else :class="dk('text-slate-600','text-slate-300')" class="w-4 h-4 rotate-90" />
                  </div>
                </div>

                <!-- Expanded detail -->
                <Transition name="expand">
                  <div v-if="expanded.has(rule.id)"
                    :class="dk('bg-slate-800/30 border-slate-800','bg-slate-50/80 border-slate-100')"
                    class="border-t px-5 py-4 space-y-3">
                    <!-- Conditions -->
                    <div v-if="Array.isArray(rule.conditions) && rule.conditions.length > 0">
                      <div class="text-xs font-semibold uppercase tracking-wider text-amber-400 mb-2">Conditions</div>
                      <div class="space-y-1.5">
                        <div v-for="(cond, i) in rule.conditions" :key="i"
                          :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
                          class="border rounded-lg px-3 py-2 flex items-center gap-2 text-xs">
                          <span :class="dk('text-slate-300','text-slate-700')" class="font-mono font-medium">{{ cond.field }}</span>
                          <span class="text-amber-400 font-mono">{{ cond.op }}</span>
                          <span :class="dk('text-slate-300','text-slate-700')" class="font-mono">{{ cond.value }}</span>
                        </div>
                      </div>
                    </div>
                    <!-- Actions -->
                    <div v-if="Array.isArray(rule.actions) && rule.actions.length > 0">
                      <div class="text-xs font-semibold uppercase tracking-wider text-violet-400 mb-2">Actions</div>
                      <div class="space-y-1.5">
                        <div v-for="(act, i) in rule.actions" :key="i"
                          :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
                          class="border rounded-lg px-3 py-2 flex items-center gap-2 text-xs">
                          <Play class="w-3 h-3 text-violet-400" />
                          <span class="text-violet-400 font-medium">{{ actionLabel(act.type) }}</span>
                          <span v-if="act.value" :class="dk('text-slate-400','text-slate-500')">: {{ act.value }}</span>
                        </div>
                      </div>
                    </div>
                    <div v-if="(!rule.conditions?.length) && (!rule.actions?.length)"
                      :class="dk('text-slate-600','text-slate-400')" class="text-xs italic">
                      No conditions or actions configured
                    </div>
                  </div>
                </Transition>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4"
        style="background:rgba(0,0,0,0.7)" @click.self="showModal=false">
        <div :class="dk('bg-slate-900 border-slate-700','bg-white border-slate-200')"
          class="border rounded-2xl w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col shadow-2xl">
          <div :class="dk('border-slate-800','border-slate-100')" class="flex items-center justify-between px-6 py-4 border-b flex-shrink-0">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-xl bg-violet-600/20">
                <GitBranch class="w-4 h-4 text-violet-400" />
              </div>
              <h2 class="font-bold">{{ editingId ? 'Edit Rule' : 'New Workflow Rule' }}</h2>
            </div>
            <button @click="showModal=false" :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-900')"
              class="p-1.5 rounded-lg transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="overflow-y-auto flex-1 px-6 py-5 space-y-5">
            <!-- Basic info -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div class="sm:col-span-2">
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Rule Name <span class="text-red-400">*</span></label>
                <input v-model="form.name" placeholder="High Value Invoice Approval"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Document Type</label>
                <select v-model="form.doc_type"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500">
                  <option v-for="d in docTypes" :key="d.value" :value="d.value">{{ d.label }}</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Trigger Event</label>
                <select v-model="form.trigger_event"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500">
                  <option v-for="e in triggerEvents" :key="e.value" :value="e.value">{{ e.label }}</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Priority</label>
                <input v-model.number="form.priority" type="number" min="1" max="100"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" />
              </div>
            </div>

            <!-- Conditions -->
            <div>
              <div class="flex items-center justify-between mb-2">
                <label class="text-xs font-semibold uppercase tracking-wider text-amber-400">Conditions</label>
                <button @click="addCondition" type="button"
                  :class="dk('bg-slate-800 hover:bg-slate-700 text-amber-400','bg-amber-50 hover:bg-amber-100 text-amber-600')"
                  class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium transition-colors">
                  <Plus class="w-3.5 h-3.5" /> Add
                </button>
              </div>
              <div v-if="form.conditions.length===0"
                :class="dk('bg-slate-800/50 border-slate-700','bg-slate-50 border-slate-200')"
                class="border rounded-lg px-4 py-3 text-xs text-center">
                No conditions — rule will always trigger
              </div>
              <div class="space-y-2">
                <div v-for="(cond, i) in form.conditions" :key="i"
                  :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
                  class="border rounded-lg px-3 py-2.5 flex items-center gap-2">
                  <select v-model="cond.field"
                    :class="dk('bg-slate-700 border-slate-600 text-slate-200','bg-slate-50 border-slate-200 text-slate-700')"
                    class="flex-1 border rounded px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-violet-500">
                    <option v-for="f in conditionFields" :key="f" :value="f">{{ f }}</option>
                  </select>
                  <select v-model="cond.op"
                    :class="dk('bg-slate-700 border-slate-600 text-amber-400','bg-slate-50 border-slate-200 text-amber-600')"
                    class="w-16 border rounded px-2 py-1.5 text-xs font-mono focus:outline-none focus:ring-1 focus:ring-violet-500">
                    <option v-for="op in conditionOps" :key="op" :value="op">{{ op }}</option>
                  </select>
                  <input v-model="cond.value" placeholder="value"
                    :class="dk('bg-slate-700 border-slate-600 text-slate-200','bg-white border-slate-200 text-slate-700')"
                    class="flex-1 border rounded px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-violet-500" />
                  <button @click="removeCondition(i)" class="text-red-400 hover:text-red-300 flex-shrink-0">
                    <X class="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Actions -->
            <div>
              <div class="flex items-center justify-between mb-2">
                <label class="text-xs font-semibold uppercase tracking-wider text-violet-400">Actions</label>
                <button @click="addAction" type="button"
                  :class="dk('bg-slate-800 hover:bg-slate-700 text-violet-400','bg-violet-50 hover:bg-violet-100 text-violet-600')"
                  class="flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium transition-colors">
                  <Plus class="w-3.5 h-3.5" /> Add
                </button>
              </div>
              <div class="space-y-2">
                <div v-for="(act, i) in form.actions" :key="i"
                  :class="dk('bg-slate-800 border-slate-700','bg-white border-slate-200')"
                  class="border rounded-lg px-3 py-2.5 flex items-center gap-2">
                  <select v-model="act.type"
                    :class="dk('bg-slate-700 border-slate-600 text-violet-300','bg-slate-50 border-slate-200 text-violet-700')"
                    class="flex-1 border rounded px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-violet-500">
                    <option v-for="a in actionTypes" :key="a.value" :value="a.value">{{ a.label }}</option>
                  </select>
                  <input v-model="act.value" placeholder="value / recipient"
                    :class="dk('bg-slate-700 border-slate-600 text-slate-200','bg-white border-slate-200 text-slate-700')"
                    class="flex-1 border rounded px-2 py-1.5 text-xs focus:outline-none focus:ring-1 focus:ring-violet-500" />
                  <button @click="removeAction(i)" class="text-red-400 hover:text-red-300 flex-shrink-0">
                    <X class="w-3.5 h-3.5" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Status -->
            <div class="flex items-center gap-3">
              <button type="button" @click="form.is_active = !form.is_active"
                :class="form.is_active ? 'bg-emerald-500' : 'bg-slate-600'"
                class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors">
                <span :class="form.is_active ? 'translate-x-6' : 'translate-x-1'"
                  class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
              </button>
              <span class="text-sm" :class="dk('text-slate-300','text-slate-700')">Rule is Active</span>
            </div>
          </div>
          <div :class="dk('border-slate-800 bg-slate-900/50','border-slate-100 bg-slate-50')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t flex-shrink-0">
            <button @click="showModal=false"
              :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
              class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="save" :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 bg-violet-600 hover:bg-violet-500 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors">
              <Save class="w-4 h-4" />
              {{ saving ? 'Saving...' : (editingId ? 'Update Rule' : 'Create Rule') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.expand-enter-active, .expand-leave-active { transition: all 0.2s ease; overflow: hidden; }
.expand-enter-from, .expand-leave-to { max-height: 0; opacity: 0; }
.expand-enter-to, .expand-leave-from { max-height: 600px; opacity: 1; }
</style>
