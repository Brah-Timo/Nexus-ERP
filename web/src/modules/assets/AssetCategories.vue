<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { assetsAPI } from '@/api/client'
import { FolderOpen, Plus, Edit, Trash2, RefreshCw, XCircle, ChevronRight } from 'lucide-vue-next'

const app = useAppStore()
const categories = ref<any[]>([])
const loading = ref(false)
const showModal = ref(false)
const editId = ref('')
const form = ref<Record<string, any>>({})

const depMethods = [
  { value: 'straight_line', label: 'Straight Line' },
  { value: 'declining_balance', label: 'Declining Balance' },
  { value: 'double_declining', label: 'Double Declining' },
  { value: 'sum_of_years', label: 'Sum of Years' },
  { value: 'units_of_production', label: 'Units of Production' },
]

async function load() {
  loading.value = true
  try {
    const res = await assetsAPI.listCategories()
    categories.value = res.data
  } finally {
    loading.value = false
  }
}

onMounted(load)

function openCreate() {
  editId.value = ''
  form.value = {
    name: '', description: '', parent_id: '',
    depreciation_method: 'straight_line',
    useful_life_years: 5, depreciation_rate: 20,
    is_active: true,
  }
  showModal.value = true
}

function openEdit(cat: any) {
  editId.value = cat.id
  form.value = { ...cat }
  showModal.value = true
}

async function save() {
  if (editId.value) {
    await assetsAPI.updateCategory(editId.value, form.value)
  } else {
    await assetsAPI.createCategory(form.value)
  }
  showModal.value = false
  load()
}

async function remove(id: string) {
  if (!confirm('Delete this category?')) return
  try {
    await assetsAPI.deleteCategory(id)
    load()
  } catch (e: any) {
    alert(e?.response?.data?.error || 'Delete failed')
  }
}

const cardCls = computed(() =>
  app.darkMode ? 'bg-slate-800/60 border-slate-700' : 'bg-white border-slate-200 shadow-sm'
)
const inputCls = computed(() =>
  app.darkMode
    ? 'bg-slate-700 border-slate-600 text-slate-100 placeholder-slate-400 focus:border-indigo-500'
    : 'bg-white border-slate-300 text-slate-900 placeholder-slate-400 focus:border-indigo-500'
)
const thCls = computed(() =>
  app.darkMode ? 'text-slate-400 border-slate-700' : 'text-slate-500 border-slate-200'
)
const tdCls = computed(() =>
  app.darkMode ? 'text-slate-300 border-slate-700' : 'text-slate-700 border-slate-200'
)

function fmtCurrency(v: number) {
  return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0)
}

const topLevelCategories = computed(() => categories.value.filter(c => !c.parent_id))
const parentOptions = computed(() => categories.value.filter(c => c.id !== editId.value))
</script>

<template>
  <div class="p-6 space-y-5">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">Asset Categories</h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
          Define categories with default depreciation settings
        </p>
      </div>
      <div class="flex gap-2">
        <button @click="load" class="flex items-center gap-2 px-3 py-2 rounded-lg text-sm border transition-colors"
          :class="app.darkMode ? 'border-slate-600 text-slate-300 hover:bg-slate-700' : 'border-slate-300 text-slate-600 hover:bg-slate-50'">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white">
          <Plus class="w-4 h-4" /> New Category
        </button>
      </div>
    </div>

    <!-- Summary Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div v-for="stat in [
        { label: 'Total Categories', value: categories.length, color: 'text-indigo-400' },
        { label: 'Active', value: categories.filter(c=>c.is_active).length, color: 'text-emerald-400' },
        { label: 'Top-Level', value: topLevelCategories.length, color: 'text-amber-400' },
        { label: 'Total Assets', value: categories.reduce((s,c)=>s+c.asset_count,0), color: 'text-sky-400' },
      ]" :key="stat.label"
        class="rounded-xl border p-4"
        :class="cardCls">
        <div class="text-xs mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">{{ stat.label }}</div>
        <div class="text-2xl font-bold" :class="stat.color">{{ stat.value }}</div>
      </div>
    </div>

    <!-- Table -->
    <div class="rounded-xl border overflow-hidden" :class="cardCls">
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b" :class="thCls">
              <th class="text-left px-4 py-3 font-medium">Category</th>
              <th class="text-left px-4 py-3 font-medium">Parent</th>
              <th class="text-left px-4 py-3 font-medium">Dep. Method</th>
              <th class="text-right px-4 py-3 font-medium">Useful Life</th>
              <th class="text-right px-4 py-3 font-medium">Rate</th>
              <th class="text-right px-4 py-3 font-medium">Assets</th>
              <th class="text-right px-4 py-3 font-medium">Net Book Value</th>
              <th class="text-center px-4 py-3 font-medium">Active</th>
              <th class="text-right px-4 py-3 font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="9" class="text-center py-10 text-slate-400">Loading...</td>
            </tr>
            <tr v-else-if="categories.length === 0">
              <td colspan="9" class="text-center py-10 text-slate-400">No categories found</td>
            </tr>
            <tr v-for="cat in categories" :key="cat.id"
              class="border-t transition-colors"
              :class="[tdCls, app.darkMode ? 'hover:bg-slate-700/30' : 'hover:bg-slate-50']">
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <FolderOpen class="w-4 h-4 text-amber-400 flex-shrink-0" />
                  <div>
                    <div class="font-medium" :class="app.darkMode ? 'text-white' : 'text-slate-900'">{{ cat.name }}</div>
                    <div class="text-xs text-slate-400">{{ cat.description }}</div>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3">
                <span v-if="cat.parent_name" class="flex items-center gap-1 text-xs text-slate-400">
                  <ChevronRight class="w-3 h-3" /> {{ cat.parent_name }}
                </span>
                <span v-else class="text-xs text-slate-500">—</span>
              </td>
              <td class="px-4 py-3">
                <span class="text-xs px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300">
                  {{ cat.depreciation_method?.replace(/_/g,' ') }}
                </span>
              </td>
              <td class="px-4 py-3 text-right">{{ cat.useful_life_years }} yrs</td>
              <td class="px-4 py-3 text-right">{{ cat.depreciation_rate }}%</td>
              <td class="px-4 py-3 text-right font-medium">{{ cat.asset_count }}</td>
              <td class="px-4 py-3 text-right font-mono text-emerald-400">{{ fmtCurrency(cat.total_net_book_value) }}</td>
              <td class="px-4 py-3 text-center">
                <span class="px-2 py-0.5 rounded-full text-xs font-medium"
                  :class="cat.is_active ? 'bg-emerald-500/20 text-emerald-300' : 'bg-slate-500/20 text-slate-400'">
                  {{ cat.is_active ? 'Yes' : 'No' }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <button @click="openEdit(cat)" class="p-1.5 rounded hover:bg-amber-500/20 text-amber-400 transition-colors">
                    <Edit class="w-4 h-4" />
                  </button>
                  <button @click="remove(cat.id)" class="p-1.5 rounded hover:bg-red-500/20 text-red-400 transition-colors">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal -->
    <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60">
      <div class="w-full max-w-lg rounded-2xl border"
        :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'">
        <div class="flex items-center justify-between p-5 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold" :class="app.darkMode ? 'text-white' : 'text-slate-900'">
            {{ editId ? 'Edit Category' : 'New Category' }}
          </h2>
          <button @click="showModal = false" class="text-slate-400 hover:text-slate-200"><XCircle class="w-5 h-5" /></button>
        </div>
        <div class="p-5 space-y-4">
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Name *</label>
            <input v-model="form.name" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
          </div>
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Description</label>
            <textarea v-model="form.description" rows="2" class="w-full px-3 py-2 rounded-lg border text-sm outline-none resize-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
          </div>
          <div>
            <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Parent Category</label>
            <select v-model="form.parent_id" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
              <option value="">None (Top Level)</option>
              <option v-for="c in parentOptions" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Dep. Method</label>
              <select v-model="form.depreciation_method" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls">
                <option v-for="m in depMethods" :key="m.value" :value="m.value">{{ m.label }}</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Rate (%)</label>
              <input v-model.number="form.depreciation_rate" type="number" step="0.01" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-600'">Useful Life (yrs)</label>
              <input v-model.number="form.useful_life_years" type="number" step="0.5" class="w-full px-3 py-2 rounded-lg border text-sm outline-none focus:ring-1 focus:ring-indigo-500" :class="inputCls" />
            </div>
            <div class="flex items-end pb-2">
              <label class="flex items-center gap-2 cursor-pointer">
                <input v-model="form.is_active" type="checkbox" class="rounded" />
                <span class="text-sm" :class="app.darkMode ? 'text-slate-300' : 'text-slate-700'">Active</span>
              </label>
            </div>
          </div>
        </div>
        <div class="flex justify-end gap-3 p-5 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showModal = false" class="px-4 py-2 rounded-lg text-sm border" :class="app.darkMode ? 'border-slate-600 text-slate-300' : 'border-slate-300 text-slate-600'">Cancel</button>
          <button @click="save" class="px-4 py-2 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-700 text-white">
            {{ editId ? 'Save Changes' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
