<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Network, Plus, Search, RefreshCw, Loader2, CheckCircle, X,
  Pencil, Building2, FolderOpen, Tag, ArrowUpDown, Filter,
  GitBranch, Users, Briefcase, Activity, TrendingUp
} from 'lucide-vue-next'
import { accountingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ─────────────────────────────────────────────────────────────────────
interface CostCenter {
  id: string
  code: string
  name: string

  parent_id: string | null
  is_active: boolean
  created_at: string
}

// ─── State ─────────────────────────────────────────────────────────────────────
const centers  = ref<CostCenter[]>([])
const loading  = ref(true)
const saving   = ref(false)
const search   = ref('')
const typeFilter = ref('')
const showModal  = ref(false)
const editTarget = ref<CostCenter | null>(null)

const form = ref({
  code: '',
  name: '',
  
  parent_id: null as string | null,
  is_active: true
})

// ─── Type config ───────────────────────────────────────────────────────────────
const typeConfig: Record<string, { label: string; color: string; bg: string; border: string; icon: any }> = {
  department: { label: 'Department', color: 'text-teal-400',   bg: 'bg-teal-500/10',   border: 'border-teal-500/20',   icon: Building2 },
  project:    { label: 'Project',    color: 'text-violet-400', bg: 'bg-violet-500/10', border: 'border-violet-500/20', icon: FolderOpen },
  branch:     { label: 'Branch',     color: 'text-amber-400',  bg: 'bg-amber-500/10',  border: 'border-amber-500/20',  icon: GitBranch },
}

// ─── Computed ──────────────────────────────────────────────────────────────────
const stats = computed(() => {
  const total  = centers.value.length
  const active = centers.value.filter(c => c.is_active).length
  const dept   = centers.value.filter(c => c.type === 'department').length
  const proj   = centers.value.filter(c => c.type === 'project').length
  const branch = centers.value.filter(c => c.type === 'branch').length
  return { total, active, dept, proj, branch }
})

const filtered = computed(() => {
  let list = [...centers.value]
  if (typeFilter.value) list = list.filter(c => c.type === typeFilter.value)
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(c => c.code.toLowerCase().includes(q) || c.name.toLowerCase().includes(q))
  }
  return list.sort((a, b) => a.code.localeCompare(b.code))
})

const parentOptions = computed(() =>
  centers.value.filter(c => c.id !== editTarget.value?.id)
)

// ─── Methods ───────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const res = await accountingAPI.getCostCenters()
    centers.value = res.data ?? []
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Failed to load cost centers', 'error')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editTarget.value = null
  form.value = { code: '', name: '', type: 'department', parent_id: null, is_active: true }
  showModal.value = true
}

function openEdit(c: CostCenter) {
  editTarget.value = c
  form.value = { code: c.code, name: c.name, type: c.type, parent_id: c.parent_id, is_active: c.is_active }
  showModal.value = true
}

async function save() {
  if (!form.value.code || !form.value.name) {
    app.addToast('Code and name are required', 'error')
    return
  }
  saving.value = true
  try {
    await accountingAPI.createCostCenter({ ...form.value })
    app.addToast(editTarget.value ? 'Cost center updated' : 'Cost center created', 'success')
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error ?? 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

function getParentName(parentId: string | null) {
  if (!parentId) return null
  const p = centers.value.find(c => c.id === parentId)
  return p ? `${p.code} — ${p.name}` : null
}

onMounted(load)
</script>

<template>
  <div class="flex flex-col h-full transition-colors duration-200"
     :class="app.darkMode ? 'bg-slate-950 text-slate-100' : 'bg-slate-50 text-slate-900'">

    <!-- ── Header ─────────────────────────────────────────────────────────── -->
    <div class="border-b border-slate-800/60 bg-slate-900/50 backdrop-blur-sm px-6 py-4 flex-shrink-0">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-lg bg-teal-500/15 border border-teal-500/25 flex items-center justify-center">
            <Network class="w-4.5 h-4.5 text-teal-400" />
          </div>
          <div>
            <h1 class="text-[15px] font-semibold text-slate-100">Cost Centers</h1>
            <p class="text-[11px] text-slate-500">Organizational units for cost allocation</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            class="h-8 px-3 rounded-lg border border-slate-700/60 bg-slate-800/60 text-slate-400 hover:text-slate-200 hover:border-slate-600 text-xs font-medium inline-flex items-center gap-1.5 transition-all disabled:opacity-50">
            <RefreshCw class="w-3.5 h-3.5" :class="loading ? 'animate-spin' : ''" />
            Refresh
          </button>
          <button @click="openCreate"
            class="h-8 px-3 rounded-lg bg-teal-600 hover:bg-teal-500 text-white text-xs font-medium inline-flex items-center gap-1.5 transition-all shadow-lg shadow-teal-900/30">
            <Plus class="w-3.5 h-3.5" />
            New Cost Center
          </button>
        </div>
      </div>
    </div>

    <!-- ── Stats ──────────────────────────────────────────────────────────── -->
    <div class="px-6 py-4 flex-shrink-0 grid grid-cols-5 gap-3">
      <div class="rounded-xl bg-slate-900/70 border border-slate-800/50 p-4 col-span-1">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider text-slate-500">Total</span>
          <Network class="w-3.5 h-3.5 text-teal-500/50" />
        </div>
        <div class="text-2xl font-bold text-slate-100">{{ stats.total }}</div>
        <div class="text-[11px] text-teal-400/70 mt-0.5">{{ stats.active }} active</div>
      </div>
      <button v-for="(cfg, type) in typeConfig" :key="type"
        @click="typeFilter = typeFilter === type ? '' : type"
        :class="['rounded-xl border p-4 transition-all text-left', typeFilter === type ? `${cfg.bg} ${cfg.border} border-2` : 'bg-slate-900/70 border-slate-800/50 hover:border-slate-700']">
        <div class="flex items-center justify-between mb-2">
          <span class="text-[10px] font-semibold uppercase tracking-wider" :class="cfg.color">{{ cfg.label }}</span>
          <component :is="cfg.icon" class="w-3.5 h-3.5 opacity-50" :class="cfg.color" />
        </div>
        <div class="text-xl font-bold text-slate-100">{{ type === 'department' ? stats.dept : type === 'project' ? stats.proj : stats.branch }}</div>
        <div v-if="typeFilter === type" class="text-[10px] mt-0.5" :class="cfg.color">Filtered</div>
      </button>
    </div>

    <!-- ── Toolbar ────────────────────────────────────────────────────────── -->
    <div class="px-6 pb-3 flex-shrink-0 flex items-center gap-3">
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-500" />
        <input v-model="search" type="text" placeholder="Search code, name..."
          class="h-8 w-64 pl-8 pr-3 rounded-lg bg-slate-900 border border-slate-700/60 text-sm text-slate-200 placeholder-slate-600 focus:outline-none focus:border-teal-500/60 focus:ring-1 focus:ring-teal-500/20 transition-all" />
        <button v-if="search" @click="search=''" class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-600 hover:text-slate-400">
          <X class="w-3 h-3" />
        </button>
      </div>
      <div v-if="typeFilter" class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[11px] font-medium border"
        :class="[typeConfig[typeFilter]?.bg, typeConfig[typeFilter]?.color, typeConfig[typeFilter]?.border]">
        {{ typeConfig[typeFilter]?.label }}
        <button @click="typeFilter=''" class="ml-0.5 hover:opacity-70"><X class="w-2.5 h-2.5" /></button>
      </div>
      <div class="ml-auto text-xs text-slate-600">{{ filtered.length }} cost centers</div>
    </div>

    <!-- ── Table ──────────────────────────────────────────────────────────── -->
    <div class="flex-1 overflow-hidden px-6 pb-6">
      <div class="h-full rounded-xl border border-slate-800/60 overflow-hidden bg-slate-900/40">
        <div v-if="loading" class="flex items-center justify-center h-full gap-3">
          <Loader2 class="w-7 h-7 text-teal-400 animate-spin" />
          <span class="text-sm text-slate-500">Loading...</span>
        </div>
        <div v-else class="overflow-auto h-full">
          <table class="w-full text-sm border-collapse">
            <thead class="sticky top-0 z-10">
              <tr class="bg-slate-900/90 backdrop-blur border-b border-slate-800/60">
                <th class="text-left px-4 py-3 w-28 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Code</th>
                <th class="text-left px-4 py-3 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Name</th>
                <th class="text-left px-4 py-3 w-32 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Type</th>
                <th class="text-left px-4 py-3 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Parent</th>
                <th class="text-left px-4 py-3 w-20 text-[11px] font-semibold uppercase tracking-wider text-slate-500">Status</th>
                <th class="px-4 py-3 w-16 text-right text-[11px] font-semibold uppercase tracking-wider text-slate-500">Edit</th>
              </tr>
            </thead>
            <tbody>
              <template v-if="filtered.length === 0">
                <tr>
                  <td colspan="6" class="py-20 text-center">
                    <Network class="w-10 h-10 mx-auto mb-3 text-slate-700" />
                    <p class="text-sm text-slate-600">No cost centers found</p>
                    <button @click="openCreate" class="mt-3 px-4 py-1.5 rounded-lg bg-teal-600 text-white text-xs">Create first cost center</button>
                  </td>
                </tr>
              </template>
              <tr v-for="(cc, i) in filtered" :key="cc.id"
                :class="['border-b border-slate-800/30 hover:bg-slate-800/30 transition-colors group', i%2===0?'':'bg-slate-900/20']">
                <!-- Code -->
                <td class="px-4 py-3">
                  <span class="font-mono text-[13px] font-semibold text-teal-300">{{ cc.code }}</span>
                </td>
                <!-- Name -->
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2">
                    <component :is="typeConfig[cc.type]?.icon ?? Building2" class="w-3.5 h-3.5 flex-shrink-0 opacity-50" :class="typeConfig[cc.type]?.color" />
                    <span class="text-[13px] text-slate-200 font-medium">{{ cc.name }}</span>
                  </div>
                </td>
                <!-- Type -->
                <td class="px-4 py-3">
                  <span :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold border', typeConfig[cc.type]?.bg, typeConfig[cc.type]?.color, typeConfig[cc.type]?.border]">
                    {{ typeConfig[cc.type]?.label ?? cc.type }}
                  </span>
                </td>
                <!-- Parent -->
                <td class="px-4 py-3">
                  <span v-if="getParentName(cc.parent_id)" class="text-[12px] text-slate-500">{{ getParentName(cc.parent_id) }}</span>
                  <span v-else class="text-slate-700 text-[12px]">—</span>
                </td>
                <!-- Status -->
                <td class="px-4 py-3">
                  <span :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold', cc.is_active ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-slate-500/10 text-slate-500 border border-slate-600/20']">
                    <span class="w-1.5 h-1.5 rounded-full" :class="cc.is_active ? 'bg-emerald-400' : 'bg-slate-500'" />
                    {{ cc.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </td>
                <!-- Edit -->
                <td class="px-4 py-3 text-right">
                  <button @click="openEdit(cc)"
                    class="opacity-0 group-hover:opacity-100 inline-flex items-center justify-center w-7 h-7 rounded-lg bg-slate-700/60 hover:bg-teal-500/20 hover:text-teal-400 text-slate-400 transition-all">
                    <Pencil class="w-3.5 h-3.5" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ── Create Modal ──────────────────────────────────────────────────── -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-all duration-200"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-all duration-150"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" @click="showModal=false" />
          <Transition
            enter-active-class="transition-all duration-200"
            enter-from-class="opacity-0 scale-95"
            enter-to-class="opacity-100 scale-100"
            leave-active-class="transition-all duration-150"
            leave-from-class="opacity-100 scale-100"
            leave-to-class="opacity-0 scale-95">
            <div v-if="showModal" class="relative w-full max-w-md bg-slate-900 border border-slate-700/60 rounded-2xl shadow-2xl overflow-hidden">
              <!-- Header -->
              <div class="px-6 py-4 border-b border-slate-800/60 flex items-center justify-between">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-lg bg-teal-500/15 border border-teal-500/25 flex items-center justify-center">
                    <Network class="w-4 h-4 text-teal-400" />
                  </div>
                  <div>
                    <h3 class="text-sm font-semibold text-slate-100">{{ editTarget ? 'Edit Cost Center' : 'New Cost Center' }}</h3>
                    <p class="text-[11px] text-slate-500">{{ editTarget ? `Editing ${editTarget.code}` : 'Define organizational unit' }}</p>
                  </div>
                </div>
                <button @click="showModal=false" class="w-7 h-7 rounded-lg flex items-center justify-center text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition-all">
                  <X class="w-4 h-4" />
                </button>
              </div>
              <!-- Body -->
              <div class="px-6 py-5 space-y-4">
                <div class="grid grid-cols-2 gap-4">
                  <div>
                    <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Code *</label>
                    <input v-model="form.code" type="text" placeholder="e.g. CC-001"
                      class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-teal-500/60 focus:ring-1 focus:ring-teal-500/20 font-mono" />
                  </div>
                </div>
                <div>
                  <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Name *</label>
                  <input v-model="form.name" type="text" placeholder="e.g. Sales Department"
                    class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-teal-500/60 focus:ring-1 focus:ring-teal-500/20" />
                </div>
                <div>
                  <label class="block text-[11px] font-semibold uppercase tracking-wider text-slate-500 mb-1.5">Parent Cost Center</label>
                  <select v-model="form.parent_id"
                    class="w-full h-9 px-3 rounded-lg bg-slate-800/60 border border-slate-700/60 text-sm text-slate-100 focus:outline-none focus:border-teal-500/60 focus:ring-1 focus:ring-teal-500/20">
                    <option :value="null">-- None (root) --</option>
                    <option v-for="p in parentOptions" :key="p.id" :value="p.id">{{ p.code }} — {{ p.name }}</option>
                  </select>
                </div>
                <label class="flex items-center gap-2.5 cursor-pointer">
                  <button type="button" @click="form.is_active = !form.is_active"
                    :class="['w-9 h-5 rounded-full relative transition-all', form.is_active ? 'bg-teal-600' : 'bg-slate-700']">
                    <span :class="['absolute top-0.5 w-4 h-4 rounded-full bg-white shadow transition-all', form.is_active ? 'left-4.5' : 'left-0.5']" />
                  </button>
                  <span class="text-xs text-slate-400">Active</span>
                </label>
              </div>
              <!-- Footer -->
              <div class="px-6 py-4 border-t border-slate-800/60 flex items-center justify-end gap-3">
                <button @click="showModal=false" class="h-9 px-4 rounded-lg border border-slate-700 bg-slate-800/40 text-slate-400 hover:text-slate-200 text-sm font-medium transition-all">Cancel</button>
                <button @click="save" :disabled="saving"
                  class="h-9 px-5 rounded-lg bg-teal-600 hover:bg-teal-500 disabled:opacity-50 text-white text-sm font-medium inline-flex items-center gap-2 transition-all shadow-lg shadow-teal-900/30">
                  <Loader2 v-if="saving" class="w-3.5 h-3.5 animate-spin" />
                  <CheckCircle v-else class="w-3.5 h-3.5" />
                  {{ editTarget ? 'Update' : 'Create' }}
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
