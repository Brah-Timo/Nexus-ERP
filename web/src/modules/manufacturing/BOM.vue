<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Search, Plus, RefreshCw, Edit2, Trash2, ChevronDown, ChevronUp,
  Layers3, Package, Wrench, X, Check, AlertCircle, Info,
  ChevronsUpDown, GripVertical, Copy
} from 'lucide-vue-next'
import { manufacturingAPI, inventoryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── State ────────────────────────────────────────────────────────────────────
interface BOMComponent {
  id?: string
  component_id: string
  component_code: string
  component_name: string
  quantity: number
  uom_id: string | null
  uom_code: string
  scrap_pct: number
  sort_order: number
}

interface BOMOperation {
  id?: string
  work_center_id: string
  work_center_name: string
  name: string
  duration_hours: number
  sort_order: number
}

interface BOM {
  id: string
  code: string
  product_id: string
  product_code: string
  product_name: string
  version: string
  quantity: number
  uom_id: string | null
  uom_code: string
  is_active: boolean
  notes: string | null
  component_count: number
  operation_count: number
  components?: BOMComponent[]
  operations?: BOMOperation[]
  created_at: string
  updated_at: string
}

const boms = ref<BOM[]>([])
const items = ref<any[]>([])
const workCenters = ref<any[]>([])
const units = ref<any[]>([])

const loading = ref(false)
const search = ref('')
const filterActive = ref('')
const sortField = ref('code')
const sortDir = ref<'asc' | 'desc'>('asc')

// Modal state
const showModal = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const saving = ref(false)

// Detail drawer
const drawerBOM = ref<BOM | null>(null)
const drawerLoading = ref(false)

// Delete confirm
const confirmDelete = ref<BOM | null>(null)
const deleting = ref(false)

// Form state
const form = ref({
  id: '',
  code: '',
  product_id: '',
  version: '1.0',
  quantity: 1,
  uom_id: null as string | null,
  is_active: true,
  notes: '',
  components: [] as BOMComponent[],
  operations: [] as BOMOperation[]
})

// ─── Computed ─────────────────────────────────────────────────────────────────
const filtered = computed(() => {
  let list = [...boms.value]
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(b =>
      b.code.toLowerCase().includes(q) ||
      b.product_name.toLowerCase().includes(q) ||
      b.product_code.toLowerCase().includes(q)
    )
  }
  if (filterActive.value === 'true') list = list.filter(b => b.is_active)
  if (filterActive.value === 'false') list = list.filter(b => !b.is_active)
  list.sort((a, b) => {
    const av = (a as any)[sortField.value] ?? ''
    const bv = (b as any)[sortField.value] ?? ''
    const cmp = String(av).localeCompare(String(bv))
    return sortDir.value === 'asc' ? cmp : -cmp
  })
  return list
})

const totalBOMs = computed(() => boms.value.length)
const activeBOMs = computed(() => boms.value.filter(b => b.is_active).length)
const totalComponents = computed(() => boms.value.reduce((s, b) => s + b.component_count, 0))
const totalOperations = computed(() => boms.value.reduce((s, b) => s + b.operation_count, 0))

// ─── Load data ────────────────────────────────────────────────────────────────
async function load() {
  loading.value = true
  try {
    const [bomsRes, itemsRes, wcRes, unitsRes] = await Promise.all([
      manufacturingAPI.getBOMs(),
      inventoryAPI.getItems(),
      manufacturingAPI.getWorkCenters({ active: 'true' }),
      inventoryAPI.getUnits()
    ])
    boms.value = bomsRes.data || []
    items.value = itemsRes.data || []
    workCenters.value = wcRes.data || []
    units.value = unitsRes.data || []
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to load BOMs', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(load)

// ─── Sort ─────────────────────────────────────────────────────────────────────
function setSort(field: string) {
  if (sortField.value === field) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortField.value = field
    sortDir.value = 'asc'
  }
}

// ─── Modal helpers ────────────────────────────────────────────────────────────
function openCreate() {
  modalMode.value = 'create'
  form.value = {
    id: '', code: '', product_id: '', version: '1.0',
    quantity: 1, uom_id: null, is_active: true, notes: '',
    components: [], operations: []
  }
  showModal.value = true
}

function openEdit(bom: BOM) {
  modalMode.value = 'edit'
  form.value = {
    id: bom.id,
    code: bom.code,
    product_id: bom.product_id,
    version: bom.version,
    quantity: bom.quantity,
    uom_id: bom.uom_id,
    is_active: bom.is_active,
    notes: bom.notes || '',
    components: bom.components ? [...bom.components] : [],
    operations: bom.operations ? [...bom.operations] : []
  }
  // Load detail if not already loaded
  if (!bom.components) loadBOMDetail(bom.id, true)
  showModal.value = true
}

async function loadBOMDetail(id: string, forEdit = false) {
  if (!forEdit) drawerLoading.value = true
  try {
    const res = await manufacturingAPI.getBOM(id)
    if (forEdit) {
      form.value.components = res.data.components || []
      form.value.operations = res.data.operations || []
    } else {
      drawerBOM.value = res.data
    }
  } catch {
    // ignore
  } finally {
    drawerLoading.value = false
  }
}

function openDrawer(bom: BOM) {
  drawerBOM.value = bom
  loadBOMDetail(bom.id)
}

function closeModal() {
  showModal.value = false
}

// ─── Component line management ────────────────────────────────────────────────
function addComponent() {
  form.value.components.push({
    component_id: '', component_code: '', component_name: '',
    quantity: 1, uom_id: null, uom_code: '', scrap_pct: 0,
    sort_order: form.value.components.length
  })
}

function removeComponent(idx: number) {
  form.value.components.splice(idx, 1)
}

function onComponentSelect(idx: number, itemId: string) {
  const item = items.value.find(i => i.id === itemId)
  if (item) {
    form.value.components[idx].component_id = item.id
    form.value.components[idx].component_code = item.code
    form.value.components[idx].component_name = item.name
    form.value.components[idx].uom_id = item.uom_id || null
    // find uom code
    const uom = units.value.find(u => u.id === item.uom_id)
    form.value.components[idx].uom_code = uom?.code || ''
  }
}

// ─── Operation line management ────────────────────────────────────────────────
function addOperation() {
  form.value.operations.push({
    work_center_id: '', work_center_name: '',
    name: '', duration_hours: 1,
    sort_order: form.value.operations.length
  })
}

function removeOperation(idx: number) {
  form.value.operations.splice(idx, 1)
}

function onWorkCenterSelect(idx: number, wcId: string) {
  const wc = workCenters.value.find(w => w.id === wcId)
  if (wc) {
    form.value.operations[idx].work_center_id = wc.id
    form.value.operations[idx].work_center_name = wc.name
  }
}

// ─── Save ─────────────────────────────────────────────────────────────────────
async function save() {
  if (!form.value.code || !form.value.product_id) {
    app.addToast('BOM code and product are required', 'error')
    return
  }
  saving.value = true
  try {
    const payload = {
      code: form.value.code,
      product_id: form.value.product_id,
      version: form.value.version,
      quantity: form.value.quantity,
      uom_id: form.value.uom_id,
      is_active: form.value.is_active,
      notes: form.value.notes || null,
      components: form.value.components,
      operations: form.value.operations
    }
    if (modalMode.value === 'create') {
      await manufacturingAPI.createBOM(payload)
      app.addToast('BOM created successfully', 'success')
    } else {
      await manufacturingAPI.updateBOM(form.value.id, payload)
      app.addToast('BOM updated successfully', 'success')
    }
    closeModal()
    await load()
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to save BOM', 'error')
  } finally {
    saving.value = false
  }
}

// ─── Delete / Deactivate ──────────────────────────────────────────────────────
async function deactivate() {
  if (!confirmDelete.value) return
  deleting.value = true
  try {
    await manufacturingAPI.deleteBOM(confirmDelete.value.id)
    app.addToast('BOM deactivated', 'success')
    confirmDelete.value = null
    await load()
  } catch (e: any) {
    app.addToast(e.response?.data?.error || 'Failed to deactivate BOM', 'error')
  } finally {
    deleting.value = false
  }
}

// ─── Helpers ──────────────────────────────────────────────────────────────────
function fmtDate(d: string) {
  return d ? new Date(d).toLocaleDateString('en-GB') : '—'
}

function getProductName(id: string) {
  const item = items.value.find(i => i.id === id)
  return item ? `${item.code} — ${item.name}` : id
}
</script>

<template>
  <div class="flex flex-col h-full gap-4 p-4 bg-slate-50 dark:bg-slate-950 min-h-screen">

    <!-- ── Header ──────────────────────────────────────────────────────────── -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
          <Layers3 class="w-6 h-6 text-violet-600" />
          Bill of Materials
        </h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Manage product structures, components, and operations
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="load" :disabled="loading"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300
                 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button @click="openCreate"
          class="flex items-center gap-2 px-4 py-2 bg-violet-600 hover:bg-violet-700 text-white
                 rounded-lg text-sm font-medium transition-colors shadow-sm">
          <Plus class="w-4 h-4" />
          New BOM
        </button>
      </div>
    </div>

    <!-- ── KPI Cards ───────────────────────────────────────────────────────── -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total BOMs</p>
        <p class="text-2xl font-bold text-slate-900 dark:text-white mt-1">{{ totalBOMs }}</p>
        <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">All versions</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Active BOMs</p>
        <p class="text-2xl font-bold text-emerald-600 mt-1">{{ activeBOMs }}</p>
        <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">In production</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total Components</p>
        <p class="text-2xl font-bold text-blue-600 mt-1">{{ totalComponents }}</p>
        <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">Across all BOMs</p>
      </div>
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Operations</p>
        <p class="text-2xl font-bold text-amber-600 mt-1">{{ totalOperations }}</p>
        <p class="text-xs text-slate-400 dark:text-slate-500 mt-0.5">Work center ops</p>
      </div>
    </div>

    <!-- ── Filters + Table ─────────────────────────────────────────────────── -->
    <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 flex-1 flex flex-col overflow-hidden">

      <!-- Toolbar -->
      <div class="flex flex-wrap items-center gap-3 p-4 border-b border-slate-200 dark:border-slate-700">
        <div class="relative flex-1 min-w-[200px]">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input v-model="search" type="text" placeholder="Search code, product..."
            class="w-full pl-9 pr-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                   bg-white dark:bg-slate-800 text-slate-900 dark:text-white
                   focus:ring-2 focus:ring-violet-500 focus:border-transparent outline-none" />
        </div>
        <select v-model="filterActive"
          class="px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-violet-500 outline-none">
          <option value="">All Status</option>
          <option value="true">Active</option>
          <option value="false">Inactive</option>
        </select>
      </div>

      <!-- Table -->
      <div class="flex-1 overflow-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/50">
              <th v-for="col in [
                { key: 'code', label: 'BOM Code' },
                { key: 'product_name', label: 'Product' },
                { key: 'version', label: 'Version' },
                { key: 'quantity', label: 'Qty' },
                { key: 'component_count', label: 'Components' },
                { key: 'operation_count', label: 'Operations' },
                { key: 'is_active', label: 'Status' },
                { key: '', label: '' }
              ]" :key="col.key"
                class="px-4 py-3 text-left font-medium text-slate-500 dark:text-slate-400 whitespace-nowrap"
                :class="col.key ? 'cursor-pointer hover:text-slate-700 dark:hover:text-slate-200 select-none' : ''"
                @click="col.key ? setSort(col.key) : null">
                <span class="flex items-center gap-1">
                  {{ col.label }}
                  <template v-if="col.key && sortField === col.key">
                    <ChevronUp v-if="sortDir === 'asc'" class="w-3 h-3" />
                    <ChevronDown v-else class="w-3 h-3" />
                  </template>
                </span>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="8" class="px-4 py-12 text-center text-slate-400 dark:text-slate-500">
                <RefreshCw class="w-5 h-5 animate-spin mx-auto mb-2" />
                Loading...
              </td>
            </tr>
            <tr v-else-if="filtered.length === 0">
              <td colspan="8" class="px-4 py-12 text-center text-slate-400 dark:text-slate-500">
                <Layers3 class="w-8 h-8 mx-auto mb-2 opacity-30" />
                No BOMs found
              </td>
            </tr>
            <tr v-for="bom in filtered" :key="bom.id"
              class="border-b border-slate-100 dark:border-slate-800 hover:bg-slate-50 dark:hover:bg-slate-800/50
                     cursor-pointer transition-colors"
              @click="openDrawer(bom)">
              <td class="px-4 py-3">
                <span class="font-mono font-semibold text-violet-700 dark:text-violet-400">{{ bom.code }}</span>
              </td>
              <td class="px-4 py-3">
                <div>
                  <p class="font-medium text-slate-900 dark:text-white">{{ bom.product_name }}</p>
                  <p class="text-xs text-slate-400">{{ bom.product_code }}</p>
                </div>
              </td>
              <td class="px-4 py-3">
                <span class="px-2 py-0.5 bg-slate-100 dark:bg-slate-700 text-slate-600 dark:text-slate-300 rounded text-xs font-mono">
                  v{{ bom.version }}
                </span>
              </td>
              <td class="px-4 py-3 text-slate-700 dark:text-slate-300 font-mono">
                {{ bom.quantity }} {{ bom.uom_code || '' }}
              </td>
              <td class="px-4 py-3">
                <span class="flex items-center gap-1 text-blue-600 dark:text-blue-400">
                  <Package class="w-3.5 h-3.5" />
                  {{ bom.component_count }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span class="flex items-center gap-1 text-amber-600 dark:text-amber-400">
                  <Wrench class="w-3.5 h-3.5" />
                  {{ bom.operation_count }}
                </span>
              </td>
              <td class="px-4 py-3">
                <span :class="bom.is_active
                  ? 'bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400'
                  : 'bg-slate-100 dark:bg-slate-700 text-slate-500'"
                  class="px-2 py-0.5 rounded-full text-xs font-medium">
                  {{ bom.is_active ? 'Active' : 'Inactive' }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1" @click.stop>
                  <button @click="openEdit(bom)"
                    class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-500 hover:text-slate-700 dark:hover:text-white transition-colors">
                    <Edit2 class="w-3.5 h-3.5" />
                  </button>
                  <button @click="confirmDelete = bom"
                    class="p-1.5 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 text-slate-500 hover:text-red-600 transition-colors">
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="px-4 py-2 border-t border-slate-200 dark:border-slate-700 text-xs text-slate-500 dark:text-slate-400">
        {{ filtered.length }} of {{ boms.length }} BOMs
      </div>
    </div>

    <!-- ── Detail Drawer ────────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="drawerBOM" class="fixed inset-0 z-40 flex justify-end">
        <div class="absolute inset-0 bg-black/30 dark:bg-black/50" @click="drawerBOM = null" />
        <div class="relative z-10 w-full max-w-xl bg-white dark:bg-slate-900 shadow-2xl flex flex-col overflow-hidden">
          <!-- Drawer header -->
          <div class="flex items-center justify-between p-5 border-b border-slate-200 dark:border-slate-700">
            <div>
              <h2 class="text-lg font-bold text-slate-900 dark:text-white font-mono">{{ drawerBOM.code }}</h2>
              <p class="text-sm text-slate-500 dark:text-slate-400">v{{ drawerBOM.version }} — {{ drawerBOM.product_name }}</p>
            </div>
            <div class="flex items-center gap-2">
              <button @click="openEdit(drawerBOM)"
                class="flex items-center gap-1.5 px-3 py-1.5 text-sm bg-violet-600 hover:bg-violet-700 text-white rounded-lg transition-colors">
                <Edit2 class="w-3.5 h-3.5" /> Edit
              </button>
              <button @click="drawerBOM = null" class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500">
                <X class="w-4 h-4" />
              </button>
            </div>
          </div>

          <div class="flex-1 overflow-y-auto p-5 space-y-6">
            <div v-if="drawerLoading" class="flex justify-center py-8">
              <RefreshCw class="w-5 h-5 animate-spin text-violet-600" />
            </div>
            <template v-else>
              <!-- Info grid -->
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Product</p>
                  <p class="text-sm font-medium text-slate-900 dark:text-white mt-0.5">{{ drawerBOM.product_name }}</p>
                </div>
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Code</p>
                  <p class="text-sm font-mono text-violet-700 dark:text-violet-400 mt-0.5">{{ drawerBOM.product_code }}</p>
                </div>
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Quantity</p>
                  <p class="text-sm font-medium text-slate-900 dark:text-white mt-0.5">{{ drawerBOM.quantity }} {{ drawerBOM.uom_code }}</p>
                </div>
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Status</p>
                  <span :class="drawerBOM.is_active
                    ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                    : 'bg-slate-100 text-slate-500 dark:bg-slate-700'"
                    class="inline-block px-2 py-0.5 rounded-full text-xs font-medium mt-0.5">
                    {{ drawerBOM.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </div>
                <div v-if="drawerBOM.notes" class="col-span-2">
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Notes</p>
                  <p class="text-sm text-slate-600 dark:text-slate-300 mt-0.5">{{ drawerBOM.notes }}</p>
                </div>
              </div>

              <!-- Components -->
              <div>
                <h3 class="text-sm font-semibold text-slate-900 dark:text-white flex items-center gap-2 mb-3">
                  <Package class="w-4 h-4 text-blue-600" />
                  Components ({{ drawerBOM.components?.length ?? drawerBOM.component_count }})
                </h3>
                <div v-if="drawerBOM.components && drawerBOM.components.length > 0"
                  class="rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
                  <table class="w-full text-xs">
                    <thead class="bg-slate-50 dark:bg-slate-800">
                      <tr>
                        <th class="px-3 py-2 text-left text-slate-500 dark:text-slate-400 font-medium">Component</th>
                        <th class="px-3 py-2 text-right text-slate-500 dark:text-slate-400 font-medium">Qty</th>
                        <th class="px-3 py-2 text-right text-slate-500 dark:text-slate-400 font-medium">Scrap%</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="comp in drawerBOM.components" :key="comp.id"
                        class="border-t border-slate-100 dark:border-slate-800">
                        <td class="px-3 py-2">
                          <p class="font-medium text-slate-900 dark:text-white">{{ comp.component_name }}</p>
                          <p class="text-slate-400">{{ comp.component_code }}</p>
                        </td>
                        <td class="px-3 py-2 text-right font-mono text-slate-700 dark:text-slate-300">
                          {{ comp.quantity }} {{ comp.uom_code }}
                        </td>
                        <td class="px-3 py-2 text-right text-slate-500 dark:text-slate-400">
                          {{ comp.scrap_pct }}%
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <p v-else class="text-xs text-slate-400 italic">No components defined</p>
              </div>

              <!-- Operations -->
              <div>
                <h3 class="text-sm font-semibold text-slate-900 dark:text-white flex items-center gap-2 mb-3">
                  <Wrench class="w-4 h-4 text-amber-600" />
                  Operations ({{ drawerBOM.operations?.length ?? drawerBOM.operation_count }})
                </h3>
                <div v-if="drawerBOM.operations && drawerBOM.operations.length > 0"
                  class="rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
                  <table class="w-full text-xs">
                    <thead class="bg-slate-50 dark:bg-slate-800">
                      <tr>
                        <th class="px-3 py-2 text-left text-slate-500 dark:text-slate-400 font-medium">Operation</th>
                        <th class="px-3 py-2 text-left text-slate-500 dark:text-slate-400 font-medium">Work Center</th>
                        <th class="px-3 py-2 text-right text-slate-500 dark:text-slate-400 font-medium">Hours</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="op in drawerBOM.operations" :key="op.id"
                        class="border-t border-slate-100 dark:border-slate-800">
                        <td class="px-3 py-2 font-medium text-slate-900 dark:text-white">{{ op.name }}</td>
                        <td class="px-3 py-2 text-slate-500 dark:text-slate-400">{{ op.work_center_name }}</td>
                        <td class="px-3 py-2 text-right font-mono text-slate-700 dark:text-slate-300">{{ op.duration_hours }}h</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <p v-else class="text-xs text-slate-400 italic">No operations defined</p>
              </div>

              <!-- Timestamps -->
              <div class="grid grid-cols-2 gap-4 pt-2 border-t border-slate-100 dark:border-slate-800">
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Created</p>
                  <p class="text-xs text-slate-600 dark:text-slate-400 mt-0.5">{{ fmtDate(drawerBOM.created_at) }}</p>
                </div>
                <div>
                  <p class="text-xs text-slate-400 uppercase tracking-wide">Updated</p>
                  <p class="text-xs text-slate-600 dark:text-slate-400 mt-0.5">{{ fmtDate(drawerBOM.updated_at) }}</p>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Create / Edit Modal ─────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-start justify-center p-4 pt-12 overflow-y-auto">
        <div class="absolute inset-0 bg-black/40 dark:bg-black/60" @click="closeModal" />
        <div class="relative z-10 w-full max-w-3xl bg-white dark:bg-slate-900 rounded-2xl shadow-2xl">
          <!-- Modal header -->
          <div class="flex items-center justify-between p-6 border-b border-slate-200 dark:border-slate-700">
            <h2 class="text-lg font-bold text-slate-900 dark:text-white">
              {{ modalMode === 'create' ? 'New Bill of Materials' : 'Edit BOM' }}
            </h2>
            <button @click="closeModal" class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500">
              <X class="w-5 h-5" />
            </button>
          </div>

          <div class="p-6 overflow-y-auto max-h-[75vh] space-y-6">
            <!-- Header fields -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                  BOM Code <span class="text-red-500">*</span>
                </label>
                <input v-model="form.code" type="text" placeholder="BOM-001"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-violet-500 outline-none" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">
                  Product <span class="text-red-500">*</span>
                </label>
                <select v-model="form.product_id"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-violet-500 outline-none">
                  <option value="">Select product...</option>
                  <option v-for="item in items.filter(i => i.item_type !== 'service')" :key="item.id" :value="item.id">
                    {{ item.code }} — {{ item.name }}
                  </option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Version</label>
                <input v-model="form.version" type="text" placeholder="1.0"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-violet-500 outline-none" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Produced Quantity</label>
                <input v-model.number="form.quantity" type="number" min="0.0001" step="0.0001"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-violet-500 outline-none" />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Unit of Measure</label>
                <select v-model="form.uom_id"
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-violet-500 outline-none">
                  <option :value="null">None</option>
                  <option v-for="u in units" :key="u.id" :value="u.id">{{ u.code }} — {{ u.name }}</option>
                </select>
              </div>
              <div class="flex items-end pb-0.5">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" v-model="form.is_active" class="w-4 h-4 rounded border-slate-300 text-violet-600 focus:ring-violet-500" />
                  <span class="text-sm text-slate-700 dark:text-slate-300">Active</span>
                </label>
              </div>
              <div class="col-span-2">
                <label class="block text-xs font-medium text-slate-700 dark:text-slate-300 mb-1.5">Notes</label>
                <textarea v-model="form.notes" rows="2" placeholder="Internal notes..."
                  class="w-full px-3 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg
                         bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-2 focus:ring-violet-500 outline-none resize-none" />
              </div>
            </div>

            <!-- Components section -->
            <div>
              <div class="flex items-center justify-between mb-3">
                <h3 class="text-sm font-semibold text-slate-900 dark:text-white flex items-center gap-2">
                  <Package class="w-4 h-4 text-blue-600" />
                  Components
                </h3>
                <button @click="addComponent" type="button"
                  class="flex items-center gap-1 px-2.5 py-1.5 text-xs bg-blue-50 dark:bg-blue-900/20 hover:bg-blue-100 dark:hover:bg-blue-900/40
                         text-blue-700 dark:text-blue-400 rounded-lg transition-colors">
                  <Plus class="w-3.5 h-3.5" /> Add Component
                </button>
              </div>

              <div v-if="form.components.length === 0"
                class="text-xs text-slate-400 italic text-center py-4 border-2 border-dashed border-slate-200 dark:border-slate-700 rounded-lg">
                No components added. Click "Add Component" to start.
              </div>

              <div v-else class="rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
                <table class="w-full text-xs">
                  <thead class="bg-slate-50 dark:bg-slate-800">
                    <tr>
                      <th class="px-3 py-2 text-left text-slate-500 dark:text-slate-400 font-medium">Component</th>
                      <th class="px-3 py-2 text-left text-slate-500 dark:text-slate-400 font-medium w-24">Qty</th>
                      <th class="px-3 py-2 text-left text-slate-500 dark:text-slate-400 font-medium w-20">Scrap%</th>
                      <th class="px-3 py-2 w-10"></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(comp, idx) in form.components" :key="idx"
                      class="border-t border-slate-100 dark:border-slate-800">
                      <td class="px-3 py-2">
                        <select :value="comp.component_id"
                          @change="onComponentSelect(idx, ($event.target as HTMLSelectElement).value)"
                          class="w-full px-2 py-1 border border-slate-200 dark:border-slate-600 rounded
                                 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-1 focus:ring-violet-500 outline-none text-xs">
                          <option value="">Select item...</option>
                          <option v-for="item in items" :key="item.id" :value="item.id">
                            {{ item.code }} — {{ item.name }}
                          </option>
                        </select>
                      </td>
                      <td class="px-3 py-2">
                        <input v-model.number="comp.quantity" type="number" min="0.0001" step="0.0001"
                          class="w-full px-2 py-1 border border-slate-200 dark:border-slate-600 rounded
                                 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-1 focus:ring-violet-500 outline-none text-xs" />
                      </td>
                      <td class="px-3 py-2">
                        <input v-model.number="comp.scrap_pct" type="number" min="0" max="100" step="0.01"
                          class="w-full px-2 py-1 border border-slate-200 dark:border-slate-600 rounded
                                 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-1 focus:ring-violet-500 outline-none text-xs" />
                      </td>
                      <td class="px-3 py-2">
                        <button @click="removeComponent(idx)" type="button"
                          class="p-1 rounded hover:bg-red-50 dark:hover:bg-red-900/20 text-slate-400 hover:text-red-600 transition-colors">
                          <X class="w-3.5 h-3.5" />
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <!-- Operations section -->
            <div>
              <div class="flex items-center justify-between mb-3">
                <h3 class="text-sm font-semibold text-slate-900 dark:text-white flex items-center gap-2">
                  <Wrench class="w-4 h-4 text-amber-600" />
                  Operations
                </h3>
                <button @click="addOperation" type="button"
                  class="flex items-center gap-1 px-2.5 py-1.5 text-xs bg-amber-50 dark:bg-amber-900/20 hover:bg-amber-100 dark:hover:bg-amber-900/40
                         text-amber-700 dark:text-amber-400 rounded-lg transition-colors">
                  <Plus class="w-3.5 h-3.5" /> Add Operation
                </button>
              </div>

              <div v-if="form.operations.length === 0"
                class="text-xs text-slate-400 italic text-center py-4 border-2 border-dashed border-slate-200 dark:border-slate-700 rounded-lg">
                No operations added. Click "Add Operation" to start.
              </div>

              <div v-else class="rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
                <table class="w-full text-xs">
                  <thead class="bg-slate-50 dark:bg-slate-800">
                    <tr>
                      <th class="px-3 py-2 text-left text-slate-500 dark:text-slate-400 font-medium">Work Center</th>
                      <th class="px-3 py-2 text-left text-slate-500 dark:text-slate-400 font-medium">Operation Name</th>
                      <th class="px-3 py-2 text-left text-slate-500 dark:text-slate-400 font-medium w-24">Duration (h)</th>
                      <th class="px-3 py-2 w-10"></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="(op, idx) in form.operations" :key="idx"
                      class="border-t border-slate-100 dark:border-slate-800">
                      <td class="px-3 py-2">
                        <select :value="op.work_center_id"
                          @change="onWorkCenterSelect(idx, ($event.target as HTMLSelectElement).value)"
                          class="w-full px-2 py-1 border border-slate-200 dark:border-slate-600 rounded
                                 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-1 focus:ring-violet-500 outline-none text-xs">
                          <option value="">Select work center...</option>
                          <option v-for="wc in workCenters" :key="wc.id" :value="wc.id">{{ wc.name }}</option>
                        </select>
                      </td>
                      <td class="px-3 py-2">
                        <input v-model="op.name" type="text" placeholder="Operation name..."
                          class="w-full px-2 py-1 border border-slate-200 dark:border-slate-600 rounded
                                 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-1 focus:ring-violet-500 outline-none text-xs" />
                      </td>
                      <td class="px-3 py-2">
                        <input v-model.number="op.duration_hours" type="number" min="0.01" step="0.01"
                          class="w-full px-2 py-1 border border-slate-200 dark:border-slate-600 rounded
                                 bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:ring-1 focus:ring-violet-500 outline-none text-xs" />
                      </td>
                      <td class="px-3 py-2">
                        <button @click="removeOperation(idx)" type="button"
                          class="p-1 rounded hover:bg-red-50 dark:hover:bg-red-900/20 text-slate-400 hover:text-red-600 transition-colors">
                          <X class="w-3.5 h-3.5" />
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <!-- Modal footer -->
          <div class="flex items-center justify-end gap-3 p-6 border-t border-slate-200 dark:border-slate-700">
            <button @click="closeModal" type="button"
              class="px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-600
                     rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
              Cancel
            </button>
            <button @click="save" :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 text-sm font-medium bg-violet-600 hover:bg-violet-700
                     disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-lg transition-colors">
              <RefreshCw v-if="saving" class="w-3.5 h-3.5 animate-spin" />
              <Check v-else class="w-3.5 h-3.5" />
              {{ modalMode === 'create' ? 'Create BOM' : 'Save Changes' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ── Delete Confirm ──────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="confirmDelete" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/40" @click="confirmDelete = null" />
        <div class="relative z-10 w-full max-w-sm bg-white dark:bg-slate-900 rounded-xl shadow-2xl p-6">
          <div class="flex items-start gap-4 mb-4">
            <div class="p-2 bg-red-100 dark:bg-red-900/30 rounded-lg">
              <AlertCircle class="w-5 h-5 text-red-600 dark:text-red-400" />
            </div>
            <div>
              <h3 class="font-semibold text-slate-900 dark:text-white">Deactivate BOM?</h3>
              <p class="text-sm text-slate-500 dark:text-slate-400 mt-1">
                BOM <strong>{{ confirmDelete.code }}</strong> will be marked inactive.
              </p>
            </div>
          </div>
          <div class="flex gap-3 justify-end">
            <button @click="confirmDelete = null"
              class="px-4 py-2 text-sm text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-600 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800 transition-colors">
              Cancel
            </button>
            <button @click="deactivate" :disabled="deleting"
              class="px-4 py-2 text-sm font-medium bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white rounded-lg transition-colors">
              Deactivate
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
