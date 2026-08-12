<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Shield, Plus, Search, Edit2, Trash2, X, RefreshCw,
  CheckCircle, AlertTriangle, Clock, Calendar, Wrench,
  ChevronDown, ToggleLeft, ToggleRight, Settings
} from 'lucide-vue-next'
import { maintenanceAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── types ────────────────────────────────────────────────────────────────────
interface Plan {
  id: string
  name: string
  description?: string
  equipment_id?: string
  equipment_name?: string
  equipment_code?: string
  frequency_type: string
  frequency_value: number
  estimated_duration?: number
  estimated_cost?: number
  assigned_to?: string
  instructions?: string
  is_active: boolean
  last_performed?: string
  next_due?: string
  created_at: string
}

// ─── state ────────────────────────────────────────────────────────────────────
const plans     = ref<Plan[]>([])
const equipment = ref<{ id: string; name: string; code: string }[]>([])
const loading   = ref(false)
const search    = ref('')
const activeFilter = ref<'all'|'active'|'inactive'>('all')

const showCreate = ref(false)
const showEdit   = ref(false)
const showDelete = ref(false)
const saving     = ref(false)
const deleting   = ref(false)
const selected   = ref<Plan | null>(null)

const form = ref({
  name: '',
  description: '',
  equipment_id: '',
  frequency_type: 'monthly',
  frequency_value: 1,
  estimated_duration: '',
  estimated_cost: '',
  assigned_to: '',
  instructions: '',
  is_active: true,
})

// ─── computed ────────────────────────────────────────────────────────────────
const dk = (a: string, b: string) => app.darkMode ? a : b

const filtered = computed(() => {
  let list = plans.value
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(p =>
      p.name.toLowerCase().includes(q) ||
      (p.equipment_name || '').toLowerCase().includes(q) ||
      (p.assigned_to || '').toLowerCase().includes(q)
    )
  }
  if (activeFilter.value === 'active')   list = list.filter(p => p.is_active)
  if (activeFilter.value === 'inactive') list = list.filter(p => !p.is_active)
  return list
})

const kpis = computed(() => [
  { label: 'Total Plans',  value: plans.value.length,                         color: 'text-violet-400', bg: 'bg-violet-500/10', icon: Shield },
  { label: 'Active',       value: plans.value.filter(p => p.is_active).length, color: 'text-emerald-400',bg: 'bg-emerald-500/10',icon: CheckCircle },
  { label: 'Due Soon',     value: plans.value.filter(p => isDueSoon(p)).length, color: 'text-amber-400',  bg: 'bg-amber-500/10',  icon: Clock },
  { label: 'Overdue',      value: plans.value.filter(p => isOverdue(p)).length, color: 'text-rose-400',   bg: 'bg-rose-500/10',   icon: AlertTriangle },
])

// ─── helpers ─────────────────────────────────────────────────────────────────
const fmt = (n: number) =>
  new Intl.NumberFormat('fr-DZ', { maximumFractionDigits: 0 }).format(n) + ' DZD'

const fmtDate = (s?: string) => {
  if (!s) return '—'
  return new Date(s).toLocaleDateString('fr-DZ', { day: '2-digit', month: 'short', year: 'numeric' })
}

const isOverdue = (p: Plan) => {
  if (!p.is_active || !p.next_due) return false
  return new Date(p.next_due) < new Date()
}

const isDueSoon = (p: Plan) => {
  if (!p.is_active || !p.next_due) return false
  const d = new Date(p.next_due)
  const now = new Date()
  const diff = (d.getTime() - now.getTime()) / 86400000
  return diff >= 0 && diff <= 7
}

const frequencyLabel = (p: Plan) => {
  const v = p.frequency_value
  const labels: Record<string, string> = {
    daily: v === 1 ? 'Daily' : `Every ${v} days`,
    weekly: v === 1 ? 'Weekly' : `Every ${v} weeks`,
    monthly: v === 1 ? 'Monthly' : `Every ${v} months`,
    quarterly: 'Quarterly',
    biannual: 'Bi-annual',
    annual: 'Annual',
    custom: `Every ${v} days`,
  }
  return labels[p.frequency_type] ?? p.frequency_type
}

const statusClass = (p: Plan) => {
  if (!p.is_active) return dk('bg-slate-800 border-slate-700','bg-slate-100 border-slate-200')
  if (isOverdue(p)) return dk('bg-rose-950/30 border-rose-800/50','bg-rose-50 border-rose-200')
  if (isDueSoon(p)) return dk('bg-amber-950/30 border-amber-800/50','bg-amber-50 border-amber-200')
  return dk('bg-slate-900 border-slate-800','bg-white border-slate-200')
}

// ─── data loading ─────────────────────────────────────────────────────────────
const load = async () => {
  loading.value = true
  try {
    const res = await maintenanceAPI.listPreventivePlans()
    plans.value = res.data.plans ?? res.data ?? []
  } catch {
    app.addToast('Failed to load preventive plans', 'error')
  } finally {
    loading.value = false
  }
}

const loadEquipment = async () => {
  try {
    const res = await maintenanceAPI.listEquipment({ limit: '500' })
    equipment.value = (res.data.equipment ?? res.data ?? []).map((e: any) => ({
      id: e.id, name: e.name, code: e.code
    }))
  } catch { /* silent */ }
}

// ─── CRUD ─────────────────────────────────────────────────────────────────────
const openCreate = () => {
  form.value = {
    name: '', description: '', equipment_id: '',
    frequency_type: 'monthly', frequency_value: 1,
    estimated_duration: '', estimated_cost: '',
    assigned_to: '', instructions: '', is_active: true,
  }
  showCreate.value = true
}

const openEdit = (p: Plan) => {
  selected.value = p
  form.value = {
    name: p.name,
    description: p.description ?? '',
    equipment_id: p.equipment_id ?? '',
    frequency_type: p.frequency_type,
    frequency_value: p.frequency_value,
    estimated_duration: p.estimated_duration?.toString() ?? '',
    estimated_cost: p.estimated_cost?.toString() ?? '',
    assigned_to: p.assigned_to ?? '',
    instructions: p.instructions ?? '',
    is_active: p.is_active,
  }
  showEdit.value = true
}

const openDelete = (p: Plan) => {
  selected.value = p
  showDelete.value = true
}

const savePlan = async (isEdit: boolean) => {
  saving.value = true
  try {
    const payload = {
      ...form.value,
      frequency_value: Number(form.value.frequency_value),
      estimated_duration: form.value.estimated_duration ? parseFloat(form.value.estimated_duration) : null,
      estimated_cost: form.value.estimated_cost ? parseFloat(form.value.estimated_cost) : null,
    }
    if (isEdit && selected.value) {
      await maintenanceAPI.updatePreventivePlan(selected.value.id, payload)
      app.addToast('Plan updated successfully', 'success')
    } else {
      await maintenanceAPI.createPreventivePlan(payload)
      app.addToast('Plan created successfully', 'success')
    }
    showCreate.value = false
    showEdit.value = false
    await load()
  } catch {
    app.addToast('Failed to save plan', 'error')
  } finally {
    saving.value = false
  }
}

const deletePlan = async () => {
  if (!selected.value) return
  deleting.value = true
  try {
    await maintenanceAPI.deletePreventivePlan(selected.value.id)
    app.addToast('Plan deleted', 'success')
    showDelete.value = false
    await load()
  } catch {
    app.addToast('Failed to delete plan', 'error')
  } finally {
    deleting.value = false
  }
}

onMounted(() => { load(); loadEquipment() })
</script>

<template>
  <div :class="['min-h-screen p-6 space-y-6', dk('bg-slate-950 text-slate-100','bg-slate-50 text-slate-900')]">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-emerald-500/15 flex items-center justify-center">
          <Shield class="w-5 h-5 text-emerald-400" />
        </div>
        <div>
          <h1 class="text-xl font-bold">Preventive Maintenance</h1>
          <p :class="['text-sm', dk('text-slate-400','text-slate-500')]">Scheduled maintenance plans</p>
        </div>
      </div>
      <button @click="openCreate"
        class="flex items-center gap-2 px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-medium transition-colors">
        <Plus class="w-4 h-4" /> New Plan
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="k in kpis" :key="k.label"
        :class="['rounded-xl border p-4', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
        <div class="flex items-center justify-between mb-2">
          <span :class="['text-xs font-medium', dk('text-slate-400','text-slate-500')]">{{ k.label }}</span>
          <div :class="['w-8 h-8 rounded-lg flex items-center justify-center', k.bg]">
            <component :is="k.icon" :class="['w-4 h-4', k.color]" />
          </div>
        </div>
        <div class="text-2xl font-bold">{{ k.value }}</div>
      </div>
    </div>

    <!-- Filters -->
    <div :class="['rounded-xl border p-4', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
      <div class="flex flex-col sm:flex-row gap-3">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="search" placeholder="Search plans, equipment, assignee…"
            :class="['w-full pl-9 pr-4 py-2 rounded-lg border text-sm',
              dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500 focus:border-emerald-500',
                 'bg-slate-50 border-slate-200 text-slate-900 placeholder-slate-400 focus:border-emerald-500')]" />
        </div>
        <div class="flex gap-2">
          <button v-for="f in ['all','active','inactive']" :key="f"
            @click="activeFilter = f as any"
            :class="['px-3 py-2 rounded-lg border text-sm font-medium capitalize transition-colors',
              activeFilter === f
                ? 'bg-emerald-600 text-white border-emerald-600'
                : dk('bg-slate-800 border-slate-700 text-slate-400 hover:border-slate-500',
                     'bg-white border-slate-200 text-slate-600 hover:border-slate-400')]">
            {{ f }}
          </button>
        </div>
        <button @click="load"
          :class="['px-3 py-2 rounded-lg border text-sm flex items-center gap-2',
            dk('bg-slate-800 border-slate-700 text-slate-300','bg-slate-50 border-slate-200 text-slate-600')]">
          <RefreshCw :class="['w-4 h-4', loading && 'animate-spin']" />
        </button>
      </div>
    </div>

    <!-- Plans Grid -->
    <div v-if="loading" class="flex items-center justify-center py-16">
      <RefreshCw class="w-8 h-8 animate-spin text-emerald-400" />
    </div>

    <div v-else-if="!filtered.length" class="text-center py-16">
      <Shield class="w-12 h-12 mx-auto mb-3 text-slate-500" />
      <p :class="['font-medium text-lg', dk('text-slate-300','text-slate-700')]">No preventive plans found</p>
      <p :class="['text-sm mt-1', dk('text-slate-500','text-slate-400')]">Create a plan to schedule recurring maintenance</p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      <!-- Plan Card -->
      <div v-for="p in filtered" :key="p.id"
        :class="['rounded-xl border p-5 space-y-4 transition-shadow hover:shadow-lg', statusClass(p)]">
        <!-- Card Header -->
        <div class="flex items-start justify-between gap-3">
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-1">
              <div :class="['w-2 h-2 rounded-full flex-shrink-0', p.is_active ? 'bg-emerald-400' : 'bg-slate-500']"></div>
              <h3 class="font-semibold text-sm truncate">{{ p.name }}</h3>
            </div>
            <div v-if="p.equipment_name" :class="['text-xs flex items-center gap-1', dk('text-slate-400','text-slate-500')]">
              <Settings class="w-3 h-3" />
              {{ p.equipment_name }}
              <span v-if="p.equipment_code" class="opacity-60">({{ p.equipment_code }})</span>
            </div>
          </div>
          <div class="flex items-center gap-1 flex-shrink-0">
            <button @click="openEdit(p)"
              :class="['p-1.5 rounded-lg transition-colors', dk('hover:bg-slate-700 text-slate-400','hover:bg-slate-200 text-slate-500')]">
              <Edit2 class="w-3.5 h-3.5" />
            </button>
            <button @click="openDelete(p)"
              :class="['p-1.5 rounded-lg transition-colors', dk('hover:bg-rose-900/40 text-rose-400','hover:bg-rose-50 text-rose-500')]">
              <Trash2 class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <!-- Frequency -->
        <div :class="['flex items-center gap-2 px-3 py-2 rounded-lg text-sm', dk('bg-slate-800/60','bg-slate-100')]">
          <RefreshCw class="w-3.5 h-3.5 text-emerald-400 flex-shrink-0" />
          <span class="font-medium">{{ frequencyLabel(p) }}</span>
        </div>

        <!-- Due dates -->
        <div class="grid grid-cols-2 gap-3 text-xs">
          <div>
            <span :class="dk('text-slate-500','text-slate-400')">Last Performed</span>
            <p class="font-medium mt-0.5">{{ fmtDate(p.last_performed) }}</p>
          </div>
          <div>
            <span :class="dk('text-slate-500','text-slate-400')">Next Due</span>
            <p :class="['font-semibold mt-0.5', isOverdue(p) ? 'text-rose-400' : isDueSoon(p) ? 'text-amber-400' : '']">
              {{ fmtDate(p.next_due) }}
            </p>
          </div>
        </div>

        <!-- Status chips -->
        <div class="flex flex-wrap gap-1.5">
          <span v-if="isOverdue(p)"
            class="flex items-center gap-1 px-2 py-0.5 rounded-md text-xs bg-rose-500/15 text-rose-400">
            <AlertTriangle class="w-3 h-3" /> Overdue
          </span>
          <span v-else-if="isDueSoon(p)"
            class="flex items-center gap-1 px-2 py-0.5 rounded-md text-xs bg-amber-500/15 text-amber-400">
            <Clock class="w-3 h-3" /> Due Soon
          </span>
          <span v-if="p.assigned_to" :class="['px-2 py-0.5 rounded-md text-xs', dk('bg-slate-700 text-slate-300','bg-slate-200 text-slate-600')]">
            {{ p.assigned_to }}
          </span>
          <span v-if="p.estimated_cost" class="px-2 py-0.5 rounded-md text-xs bg-teal-500/15 text-teal-400">
            {{ fmt(p.estimated_cost) }}
          </span>
          <span v-if="p.estimated_duration" :class="['px-2 py-0.5 rounded-md text-xs', dk('bg-slate-700 text-slate-300','bg-slate-200 text-slate-600')]">
            {{ p.estimated_duration }}h
          </span>
        </div>

        <!-- Description -->
        <p v-if="p.description" :class="['text-xs line-clamp-2 leading-relaxed', dk('text-slate-400','text-slate-500')]">
          {{ p.description }}
        </p>

        <!-- Footer -->
        <div :class="['flex items-center justify-between pt-2 border-t text-xs', dk('border-slate-700','border-slate-200')]">
          <span :class="dk('text-slate-500','text-slate-400')">
            Created {{ fmtDate(p.created_at) }}
          </span>
          <span :class="['font-medium', p.is_active ? 'text-emerald-400' : dk('text-slate-500','text-slate-400')]">
            {{ p.is_active ? 'Active' : 'Inactive' }}
          </span>
        </div>
      </div>
    </div>

    <!-- ── Create / Edit Modal ─────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showCreate || showEdit"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
        @click.self="showCreate=false; showEdit=false">
        <div :class="['w-full max-w-2xl rounded-2xl border shadow-2xl', dk('bg-slate-900 border-slate-700','bg-white border-slate-200')]">
          <div :class="['flex items-center justify-between px-6 py-4 border-b', dk('border-slate-800','border-slate-200')]">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-emerald-500/15 flex items-center justify-center">
                <Shield class="w-4 h-4 text-emerald-400" />
              </div>
              <h2 class="font-semibold">{{ showEdit ? 'Edit Plan' : 'New Preventive Plan' }}</h2>
            </div>
            <button @click="showCreate=false; showEdit=false"
              :class="['p-1.5 rounded-lg', dk('hover:bg-slate-800 text-slate-400','hover:bg-slate-100 text-slate-500')]">
              <X class="w-4 h-4" />
            </button>
          </div>
          <div class="p-6 space-y-4 max-h-[70vh] overflow-y-auto">
            <!-- Name -->
            <div>
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Plan Name *</label>
              <input v-model="form.name" placeholder="e.g. Monthly Oil Change"
                :class="['w-full px-3 py-2 rounded-lg border text-sm',
                  dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500 focus:border-emerald-500',
                     'bg-white border-slate-200 text-slate-900 placeholder-slate-400 focus:border-emerald-500')]" />
            </div>
            <!-- Equipment -->
            <div>
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Equipment</label>
              <select v-model="form.equipment_id"
                :class="['w-full px-3 py-2 rounded-lg border text-sm',
                  dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')]">
                <option value="">— None —</option>
                <option v-for="e in equipment" :key="e.id" :value="e.id">{{ e.name }} ({{ e.code }})</option>
              </select>
            </div>
            <!-- Frequency -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Frequency Type</label>
                <select v-model="form.frequency_type"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')]">
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="monthly">Monthly</option>
                  <option value="quarterly">Quarterly</option>
                  <option value="biannual">Bi-Annual</option>
                  <option value="annual">Annual</option>
                  <option value="custom">Custom (days)</option>
                </select>
              </div>
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Every N Units</label>
                <input v-model.number="form.frequency_value" type="number" min="1"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')]" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Est. Duration (hours)</label>
                <input v-model="form.estimated_duration" type="number" step="0.5" min="0" placeholder="0"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                       'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
              </div>
              <div>
                <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Est. Cost (DZD)</label>
                <input v-model="form.estimated_cost" type="number" min="0" placeholder="0"
                  :class="['w-full px-3 py-2 rounded-lg border text-sm',
                    dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                       'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
              </div>
            </div>
            <div>
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Assigned To</label>
              <input v-model="form.assigned_to" placeholder="Technician or team name"
                :class="['w-full px-3 py-2 rounded-lg border text-sm',
                  dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                     'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
            </div>
            <!-- Active toggle -->
            <div class="flex items-center gap-3">
              <button @click="form.is_active = !form.is_active"
                :class="['p-1 rounded transition-colors', form.is_active ? 'text-emerald-400' : dk('text-slate-600','text-slate-400')]">
                <component :is="form.is_active ? ToggleRight : ToggleLeft" class="w-7 h-7" />
              </button>
              <span class="text-sm font-medium">{{ form.is_active ? 'Active' : 'Inactive' }}</span>
            </div>
            <div>
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Instructions</label>
              <textarea v-model="form.instructions" rows="3" placeholder="Step-by-step maintenance instructions…"
                :class="['w-full px-3 py-2 rounded-lg border text-sm resize-none',
                  dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                     'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
            </div>
            <div>
              <label :class="['block text-xs font-medium mb-1.5', dk('text-slate-400','text-slate-600')]">Description</label>
              <textarea v-model="form.description" rows="2" placeholder="Optional description…"
                :class="['w-full px-3 py-2 rounded-lg border text-sm resize-none',
                  dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500',
                     'bg-white border-slate-200 text-slate-900 placeholder-slate-400')]" />
            </div>
          </div>
          <div :class="['flex justify-end gap-3 px-6 py-4 border-t', dk('border-slate-800','border-slate-200')]">
            <button @click="showCreate=false; showEdit=false"
              :class="['px-4 py-2 rounded-lg border text-sm',
                dk('border-slate-700 text-slate-300 hover:bg-slate-800','border-slate-200 text-slate-600 hover:bg-slate-50')]">
              Cancel
            </button>
            <button @click="savePlan(showEdit)" :disabled="saving || !form.name"
              class="px-4 py-2 rounded-lg bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white text-sm font-medium transition-colors">
              {{ saving ? 'Saving…' : showEdit ? 'Save Changes' : 'Create Plan' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Delete Modal ────────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showDelete && selected"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
        @click.self="showDelete=false">
        <div :class="['w-full max-w-sm rounded-2xl border shadow-2xl p-6', dk('bg-slate-900 border-slate-700','bg-white border-slate-200')]">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 rounded-xl bg-rose-500/15 flex items-center justify-center">
              <Trash2 class="w-5 h-5 text-rose-400" />
            </div>
            <div>
              <h2 class="font-semibold">Delete Plan</h2>
              <p :class="['text-xs', dk('text-slate-400','text-slate-500')]">This cannot be undone</p>
            </div>
          </div>
          <p :class="['text-sm mb-5', dk('text-slate-300','text-slate-600')]">
            Delete plan <span class="font-semibold text-rose-400">{{ selected.name }}</span>?
          </p>
          <div class="flex gap-3">
            <button @click="showDelete=false"
              :class="['flex-1 py-2 rounded-lg border text-sm',
                dk('border-slate-700 text-slate-300 hover:bg-slate-800','border-slate-200 text-slate-600 hover:bg-slate-50')]">
              Cancel
            </button>
            <button @click="deletePlan" :disabled="deleting"
              class="flex-1 py-2 rounded-lg bg-rose-600 hover:bg-rose-500 disabled:opacity-50 text-white text-sm font-medium transition-colors">
              {{ deleting ? 'Deleting…' : 'Delete' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
