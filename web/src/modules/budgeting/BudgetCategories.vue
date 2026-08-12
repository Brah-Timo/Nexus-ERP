<template>
  <div class="p-6" :class="app.darkMode ? 'bg-gray-900 text-white' : 'bg-gray-50 text-gray-900'">
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <Tag class="w-7 h-7 text-indigo-500" />
          Budget Categories
        </h1>
        <p class="text-sm mt-1" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">
          Manage budget classification categories and hierarchy
        </p>
      </div>
      <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 text-sm font-medium">
        <Plus class="w-4 h-4" /> New Category
      </button>
    </div>

    <!-- Table -->
    <div class="rounded-xl border overflow-hidden" :class="app.darkMode ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'">
      <div class="p-4 border-b flex gap-3" :class="app.darkMode ? 'border-gray-700' : 'border-gray-100'">
        <div class="relative flex-1">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'" />
          <input v-model="search" placeholder="Search categories..."
            class="w-full pl-9 pr-3 py-2 rounded-lg border text-sm"
            :class="app.darkMode ? 'bg-gray-700 border-gray-600 text-white placeholder-gray-500' : 'bg-gray-50 border-gray-300 text-gray-900'" />
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'bg-gray-750 text-gray-400' : 'bg-gray-50 text-gray-500'">
              <th class="px-4 py-3 text-left font-medium">Code</th>
              <th class="px-4 py-3 text-left font-medium">Name</th>
              <th class="px-4 py-3 text-left font-medium">Parent</th>
              <th class="px-4 py-3 text-left font-medium">Description</th>
              <th class="px-4 py-3 text-center font-medium">Sort</th>
              <th class="px-4 py-3 text-center font-medium">Status</th>
              <th class="px-4 py-3 text-center font-medium">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y" :class="app.darkMode ? 'divide-gray-700' : 'divide-gray-100'">
            <tr v-for="cat in filtered" :key="cat.id"
              class="hover:bg-opacity-50 transition-colors"
              :class="app.darkMode ? 'hover:bg-gray-700' : 'hover:bg-gray-50'">
              <td class="px-4 py-3 font-mono font-medium text-indigo-500">{{ cat.code }}</td>
              <td class="px-4 py-3 font-medium">{{ cat.name }}</td>
              <td class="px-4 py-3 text-sm" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">
                {{ cat.parent_name || '—' }}
              </td>
              <td class="px-4 py-3 max-w-xs truncate" :class="app.darkMode ? 'text-gray-400' : 'text-gray-500'">
                {{ cat.description || '—' }}
              </td>
              <td class="px-4 py-3 text-center">{{ cat.sort_order }}</td>
              <td class="px-4 py-3 text-center">
                <span class="px-2 py-0.5 rounded-full text-xs font-medium"
                  :class="cat.is_active ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300' : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'">
                  {{ cat.is_active ? 'Active' : 'Inactive' }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center justify-center gap-2">
                  <button @click="openEdit(cat)" class="p-1.5 rounded hover:bg-indigo-100 dark:hover:bg-indigo-900 text-indigo-500 transition-colors">
                    <Pencil class="w-4 h-4" />
                  </button>
                  <button @click="doDelete(cat)" class="p-1.5 rounded hover:bg-red-100 dark:hover:bg-red-900 text-red-500 transition-colors">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="!filtered.length">
              <td colspan="7" class="px-4 py-12 text-center" :class="app.darkMode ? 'text-gray-500' : 'text-gray-400'">
                No categories found
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm" @click.self="showModal = false">
        <div class="w-full max-w-lg rounded-2xl shadow-2xl p-6" :class="app.darkMode ? 'bg-gray-800 text-white' : 'bg-white text-gray-900'">
          <h2 class="text-lg font-bold mb-5 flex items-center gap-2">
            <Tag class="w-5 h-5 text-indigo-500" />
            {{ editing ? 'Edit' : 'New' }} Budget Category
          </h2>
          <div class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Code *</label>
                <input v-model="form.code" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Sort Order</label>
                <input v-model.number="form.sort_order" type="number" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Name *</label>
              <input v-model="form.name" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Parent Category</label>
              <select v-model="form.parent_id" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass">
                <option value="">None (top-level)</option>
                <option v-for="c in categories.filter(x => x.id !== form.id)" :key="c.id" :value="c.id">
                  {{ c.code }} — {{ c.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1 uppercase tracking-wide">Description</label>
              <textarea v-model="form.description" rows="2" class="w-full rounded-lg border px-3 py-2 text-sm" :class="inputClass"></textarea>
            </div>
            <label class="flex items-center gap-2 cursor-pointer select-none">
              <input type="checkbox" v-model="form.is_active" class="rounded" />
              <span class="text-sm">Active</span>
            </label>
          </div>
          <div class="flex justify-end gap-3 mt-6">
            <button @click="showModal = false" class="px-4 py-2 rounded-lg text-sm border" :class="app.darkMode ? 'border-gray-600 text-gray-300 hover:bg-gray-700' : 'border-gray-300 text-gray-600 hover:bg-gray-50'">Cancel</button>
            <button @click="save" :disabled="saving" class="px-5 py-2 rounded-lg text-sm bg-indigo-600 text-white hover:bg-indigo-700 font-medium disabled:opacity-60">
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Tag, Plus, Search, Pencil, Trash2 } from 'lucide-vue-next'
import { budgetingAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const categories = ref<Record<string, unknown>[]>([])
const search = ref('')
const showModal = ref(false)
const saving = ref(false)
const editing = ref(false)

const blankForm = () => ({ id: '', code: '', name: '', description: '', parent_id: '', is_active: true, sort_order: 0 })
const form = ref(blankForm())

const filtered = computed(() => {
  const q = search.value.toLowerCase()
  return categories.value.filter(c =>
    !q || String(c.name).toLowerCase().includes(q) || String(c.code).toLowerCase().includes(q)
  )
})

const inputClass = computed(() => app.darkMode
  ? 'bg-gray-700 border-gray-600 text-white placeholder-gray-500 focus:border-indigo-500 focus:ring-indigo-500 focus:outline-none'
  : 'bg-white border-gray-300 text-gray-900 focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none')

function openCreate() {
  form.value = blankForm()
  editing.value = false
  showModal.value = true
}

function openEdit(cat: Record<string, unknown>) {
  form.value = { ...blankForm(), ...cat as Record<string, unknown> } as typeof form.value
  editing.value = true
  showModal.value = true
}

async function save() {
  saving.value = true
  try {
    if (editing.value) {
      await budgetingAPI.updateCategory(form.value.id, form.value)
    } else {
      await budgetingAPI.createCategory(form.value)
    }
    showModal.value = false
    await load()
  } catch (e) {
    console.error(e)
  } finally {
    saving.value = false
  }
}

async function doDelete(cat: Record<string, unknown>) {
  if (!confirm(`Delete category "${cat.name}"?`)) return
  await budgetingAPI.deleteCategory(cat.id as string)
  await load()
}

async function load() {
  const r = await budgetingAPI.listCategories()
  categories.value = r.data.data || []
}

onMounted(load)
</script>
