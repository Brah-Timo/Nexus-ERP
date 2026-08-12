<template>
  <div class="flex flex-col h-full bg-gray-50 dark:bg-gray-950 min-h-screen">

    <!-- Header -->
    <div class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Warehouses</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Manage storage facilities and locations</p>
        </div>
        <div class="flex items-center gap-3">
          <button @click="openCreateWarehouse"
            class="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            <Plus :size="16" />
            New Warehouse
          </button>
        </div>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="px-6 py-4 grid grid-cols-2 lg:grid-cols-3 gap-4">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Total Warehouses</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ warehouses.length }}</p>
          </div>
          <div class="w-10 h-10 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg flex items-center justify-center">
            <Warehouse :size="20" class="text-indigo-600 dark:text-indigo-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Active</p>
            <p class="text-2xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">{{ warehouses.filter(w => w.is_active).length }}</p>
          </div>
          <div class="w-10 h-10 bg-emerald-100 dark:bg-emerald-900/40 rounded-lg flex items-center justify-center">
            <CheckCircle :size="20" class="text-emerald-600 dark:text-emerald-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Locations</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ allLocations.length }}</p>
          </div>
          <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900/40 rounded-lg flex items-center justify-center">
            <MapPin :size="20" class="text-blue-600 dark:text-blue-400" />
          </div>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="flex-1 px-6 pb-6 grid grid-cols-1 lg:grid-cols-3 gap-6">

      <!-- Warehouse List -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden flex flex-col">
        <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <h2 class="text-sm font-semibold text-gray-900 dark:text-white">Warehouses</h2>
          <span class="text-xs text-gray-500 dark:text-gray-400">{{ warehouses.length }} total</span>
        </div>
        <div v-if="loading" class="flex items-center justify-center py-12">
          <Loader2 :size="24" class="animate-spin text-indigo-500" />
        </div>
        <div v-else-if="warehouses.length === 0" class="flex flex-col items-center justify-center py-12 px-4">
          <Warehouse :size="36" class="text-gray-300 dark:text-gray-600 mb-2" />
          <p class="text-sm text-gray-500 dark:text-gray-400 text-center">No warehouses yet</p>
        </div>
        <div v-else class="flex-1 overflow-y-auto divide-y divide-gray-100 dark:divide-gray-800">
          <div v-for="w in warehouses" :key="w.id"
            @click="selectWarehouse(w)"
            :class="selectedWarehouse?.id === w.id ? 'bg-indigo-50 dark:bg-indigo-900/20 border-l-2 border-indigo-500' : 'hover:bg-gray-50 dark:hover:bg-gray-800/50'"
            class="flex items-center justify-between px-4 py-3 cursor-pointer transition-colors">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <span class="font-mono text-xs text-indigo-600 dark:text-indigo-400 font-semibold">{{ w.code }}</span>
                <span :class="w.is_active ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'"
                  class="text-xs px-1.5 py-0.5 rounded font-medium">
                  {{ w.is_active ? 'Active' : 'Inactive' }}
                </span>
              </div>
              <p class="text-sm font-medium text-gray-900 dark:text-white mt-0.5 truncate">{{ w.name }}</p>
              <p v-if="w.address" class="text-xs text-gray-400 dark:text-gray-500 mt-0.5 truncate">{{ w.address }}</p>
            </div>
            <div class="flex items-center gap-1 ml-2 flex-shrink-0">
              <button @click.stop="openEditWarehouse(w)"
                class="p-1.5 rounded-lg text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 transition-colors">
                <Pencil :size="13" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Location Tree -->
      <div class="lg:col-span-2 bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden flex flex-col">
        <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <div>
            <h2 class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ selectedWarehouse ? selectedWarehouse.name + ' — Locations' : 'Select a Warehouse' }}
            </h2>
            <p v-if="selectedWarehouse" class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              {{ warehouseLocations.length }} location(s)
            </p>
          </div>
          <button v-if="selectedWarehouse" @click="openCreateLocation"
            class="inline-flex items-center gap-1.5 text-xs text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 font-medium px-2 py-1 rounded-lg hover:bg-indigo-50 dark:hover:bg-indigo-900/30 transition-colors">
            <Plus :size="12" />
            Add Location
          </button>
        </div>

        <div v-if="!selectedWarehouse" class="flex flex-col items-center justify-center flex-1 py-16">
          <MapPin :size="40" class="text-gray-300 dark:text-gray-600 mb-3" />
          <p class="text-gray-500 dark:text-gray-400">Select a warehouse to view its locations</p>
        </div>

        <div v-else-if="locationsLoading" class="flex items-center justify-center flex-1 py-12">
          <Loader2 :size="24" class="animate-spin text-indigo-500" />
        </div>

        <div v-else-if="warehouseLocations.length === 0" class="flex flex-col items-center justify-center flex-1 py-12">
          <FolderOpen :size="40" class="text-gray-300 dark:text-gray-600 mb-3" />
          <p class="text-gray-500 dark:text-gray-400 font-medium">No locations configured</p>
          <p class="text-sm text-gray-400 dark:text-gray-500 mt-1">Add bin locations, aisles, or zones</p>
          <button @click="openCreateLocation"
            class="mt-4 inline-flex items-center gap-2 text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 font-medium">
            <Plus :size="14" />
            Add first location
          </button>
        </div>

        <div v-else class="flex-1 overflow-y-auto">
          <div class="p-4">
            <!-- Root locations -->
            <div v-for="loc in rootLocations" :key="loc.id" class="mb-2">
              <LocationNode :location="loc" :all-locations="warehouseLocations" :depth="0"
                @add-child="openCreateChildLocation" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Warehouse Modal -->
    <Teleport to="body">
      <div v-if="showWarehouseModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="closeWarehouseModal"></div>
        <div class="relative bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-md">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800">
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">
              {{ editingWarehouse ? 'Edit Warehouse' : 'New Warehouse' }}
            </h2>
            <button @click="closeWarehouseModal" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 transition-colors">
              <X :size="18" />
            </button>
          </div>
          <div class="px-6 py-4 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Code <span class="text-red-500">*</span></label>
                <input v-model="whForm.code" type="text" placeholder="WH-01"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name <span class="text-red-500">*</span></label>
                <input v-model="whForm.name" type="text" placeholder="Main Warehouse"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div class="col-span-2">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Address</label>
                <textarea v-model="whForm.address" rows="2" placeholder="Physical address"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"></textarea>
              </div>
              <div v-if="editingWarehouse" class="col-span-2">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" v-model="whForm.is_active"
                    class="w-4 h-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" />
                  <span class="text-sm text-gray-700 dark:text-gray-300">Active</span>
                </label>
              </div>
            </div>
          </div>
          <div class="border-t border-gray-200 dark:border-gray-800 px-6 py-4 flex items-center justify-end gap-3">
            <button @click="closeWarehouseModal" class="text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white px-4 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
              Cancel
            </button>
            <button @click="saveWarehouse" :disabled="saving"
              class="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 text-white text-sm font-medium px-5 py-2 rounded-lg transition-colors">
              <Loader2 v-if="saving" :size="14" class="animate-spin" />
              <Save v-else :size="14" />
              {{ editingWarehouse ? 'Update' : 'Create' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Location Modal -->
    <Teleport to="body">
      <div v-if="showLocationModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showLocationModal = false"></div>
        <div class="relative bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-md">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800">
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">Add Location</h2>
            <button @click="showLocationModal = false" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 transition-colors">
              <X :size="18" />
            </button>
          </div>
          <div class="px-6 py-4 space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Code <span class="text-red-500">*</span></label>
              <input v-model="locForm.code" type="text" placeholder="A-01-01"
                class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name <span class="text-red-500">*</span></label>
              <input v-model="locForm.name" type="text" placeholder="Aisle A, Row 1, Bin 1"
                class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Parent Location</label>
              <select v-model="locForm.parent_id"
                class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                <option value="">No parent (root)</option>
                <option v-for="l in warehouseLocations" :key="l.id" :value="l.id">
                  {{ l.code }} – {{ l.name }}
                </option>
              </select>
            </div>
          </div>
          <div class="border-t border-gray-200 dark:border-gray-800 px-6 py-4 flex items-center justify-end gap-3">
            <button @click="showLocationModal = false" class="text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white px-4 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
              Cancel
            </button>
            <button @click="saveLocation" :disabled="saving"
              class="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 text-white text-sm font-medium px-5 py-2 rounded-lg transition-colors">
              <Loader2 v-if="saving" :size="14" class="animate-spin" />
              <Save v-else :size="14" />
              Add Location
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, defineComponent, h } from 'vue'
import {
  Plus, Warehouse, CheckCircle, MapPin, Pencil, X, Save, Loader2,
  FolderOpen, FolderTree, ChevronRight
} from 'lucide-vue-next'
import { inventoryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ────────────────────────────────────────────────────────────────────

interface WarehouseItem {
  id: string
  company_id: string
  code: string
  name: string
  address?: string
  is_active: boolean
  created_at?: string
}

interface Location {
  id: string
  warehouse_id: string
  code: string
  name: string
  parent_id?: string
}

// ─── State ────────────────────────────────────────────────────────────────────

const warehouses = ref<WarehouseItem[]>([])
const allLocations = ref<Location[]>([])
const warehouseLocations = ref<Location[]>([])
const loading = ref(false)
const locationsLoading = ref(false)
const saving = ref(false)

const selectedWarehouse = ref<WarehouseItem | null>(null)

const showWarehouseModal = ref(false)
const editingWarehouse = ref<WarehouseItem | null>(null)
const whForm = ref({ code: '', name: '', address: '', is_active: true })

const showLocationModal = ref(false)
const locForm = ref({ code: '', name: '', parent_id: '' })

// ─── Computed ─────────────────────────────────────────────────────────────────

const rootLocations = computed(() =>
  warehouseLocations.value.filter(l => !l.parent_id)
)

// ─── Location Tree Node Component ─────────────────────────────────────────────

const LocationNode = defineComponent({
  name: 'LocationNode',
  props: {
    location: { type: Object as () => Location, required: true },
    allLocations: { type: Array as () => Location[], required: true },
    depth: { type: Number, default: 0 },
  },
  emits: ['add-child'],
  setup(props, { emit }) {
    const children = computed(() =>
      props.allLocations.filter(l => l.parent_id === props.location.id)
    )
    const expanded = ref(true)

    return () => h('div', { class: 'select-none' }, [
      h('div', {
        class: `flex items-center gap-2 px-2 py-1.5 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 group`,
        style: { paddingLeft: `${props.depth * 16 + 8}px` },
      }, [
        h('div', {
          class: `w-4 h-4 flex items-center justify-center cursor-pointer ${children.value.length === 0 ? 'invisible' : ''}`,
          onClick: () => { expanded.value = !expanded.value }
        }, [
          h(ChevronRight, {
            size: 12,
            class: `text-gray-400 transition-transform ${expanded.value ? 'rotate-90' : ''}`
          })
        ]),
        h(FolderTree, { size: 14, class: 'text-amber-500 flex-shrink-0' }),
        h('div', { class: 'flex-1 min-w-0' }, [
          h('span', { class: 'font-mono text-xs text-indigo-600 dark:text-indigo-400 font-semibold mr-2' }, props.location.code),
          h('span', { class: 'text-sm text-gray-700 dark:text-gray-300' }, props.location.name),
        ]),
        h('button', {
          class: 'opacity-0 group-hover:opacity-100 p-1 rounded text-gray-400 hover:text-indigo-600 transition-all',
          onClick: (e: Event) => { e.stopPropagation(); emit('add-child', props.location) }
        }, [h(Plus, { size: 11 })])
      ]),
      expanded.value && children.value.length > 0
        ? h('div', {}, children.value.map(child =>
            h(LocationNode, {
              key: child.id,
              location: child,
              allLocations: props.allLocations,
              depth: props.depth + 1,
              onAddChild: (loc: Location) => emit('add-child', loc)
            })
          ))
        : null
    ])
  }
})

// ─── Actions ──────────────────────────────────────────────────────────────────

const selectWarehouse = async (w: WarehouseItem) => {
  selectedWarehouse.value = w
  locationsLoading.value = true
  try {
    const res = await inventoryAPI.getLocations(w.id)
    warehouseLocations.value = res.data || []
  } catch {
    app.addToast('Failed to load locations', 'error')
  } finally {
    locationsLoading.value = false
  }
}

const openCreateWarehouse = () => {
  editingWarehouse.value = null
  whForm.value = { code: '', name: '', address: '', is_active: true }
  showWarehouseModal.value = true
}

const openEditWarehouse = (w: WarehouseItem) => {
  editingWarehouse.value = w
  whForm.value = { code: w.code, name: w.name, address: w.address || '', is_active: w.is_active }
  showWarehouseModal.value = true
}

const closeWarehouseModal = () => {
  showWarehouseModal.value = false
  editingWarehouse.value = null
}

const saveWarehouse = async () => {
  if (!whForm.value.code || !whForm.value.name) {
    app.addToast('Code and Name are required', 'error')
    return
  }
  saving.value = true
  try {
    const payload = {
      code: whForm.value.code,
      name: whForm.value.name,
      address: whForm.value.address || null,
      is_active: whForm.value.is_active,
    }
    if (editingWarehouse.value) {
      await inventoryAPI.updateWarehouse(editingWarehouse.value.id, payload)
      app.addToast('Warehouse updated', 'success')
    } else {
      await inventoryAPI.createWarehouse(payload)
      app.addToast('Warehouse created', 'success')
    }
    closeWarehouseModal()
    await load()
  } catch {
    app.addToast('Failed to save warehouse', 'error')
  } finally {
    saving.value = false
  }
}

const openCreateLocation = () => {
  locForm.value = { code: '', name: '', parent_id: '' }
  showLocationModal.value = true
}

const openCreateChildLocation = (parent: Location) => {
  locForm.value = { code: '', name: '', parent_id: parent.id }
  showLocationModal.value = true
}

const saveLocation = async () => {
  if (!locForm.value.code || !locForm.value.name) {
    app.addToast('Code and Name are required', 'error')
    return
  }
  if (!selectedWarehouse.value) return
  saving.value = true
  try {
    await inventoryAPI.createLocation({
      warehouse_id: selectedWarehouse.value.id,
      code: locForm.value.code,
      name: locForm.value.name,
      parent_id: locForm.value.parent_id || null,
    })
    app.addToast('Location added', 'success')
    showLocationModal.value = false
    await selectWarehouse(selectedWarehouse.value)
  } catch {
    app.addToast('Failed to add location', 'error')
  } finally {
    saving.value = false
  }
}

// ─── Load ─────────────────────────────────────────────────────────────────────

const load = async () => {
  loading.value = true
  try {
    const [whRes, locRes] = await Promise.all([
      inventoryAPI.getWarehouses(),
      inventoryAPI.getLocations(),
    ])
    warehouses.value = whRes.data || []
    allLocations.value = locRes.data || []
  } catch {
    app.addToast('Failed to load warehouses', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
