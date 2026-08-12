<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { helpdeskAPI } from '@/api/client'
import { Tag, Plus, Edit, Trash2, RefreshCw, X, ChevronRight, FolderOpen } from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const categories = ref<any[]>([])
const showForm = ref(false)
const editId = ref<string | null>(null)

const form = ref({
  name: '', description: '', color: '#6366f1', parent_id: '', is_active: true, sort_order: 0
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const res = await helpdeskAPI.listCategories()
    categories.value = res.data.categories || []
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to load'
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editId.value = null
  form.value = { name: '', description: '', color: '#6366f1', parent_id: '', is_active: true, sort_order: 0 }
  showForm.value = true
}

function openEdit(cat: any) {
  editId.value = cat.id
  form.value = { ...cat }
  showForm.value = true
}

async function save() {
  if (!form.value.name.trim()) { error.value = 'Name is required'; return }
  saving.value = true
  error.value = ''
  try {
    if (editId.value) {
      await helpdeskAPI.updateCategory(editId.value, form.value)
    } else {
      await helpdeskAPI.createCategory(form.value)
    }
    showForm.value = false
    await load()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to save'
  } finally {
    saving.value = false
  }
}

async function deleteCategory(id: string) {
  if (!confirm('Delete this category?')) return
  try {
    await helpdeskAPI.deleteCategory(id)
    await load()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to delete'
  }
}

onMounted(load)

// Group by parent
const topLevel = () => categories.value.filter(c => !c.parent_id)
const children = (parentId: string) => categories.value.filter(c => c.parent_id === parentId)
</script>

<template>
  <div :class="app.darkMode ? 'bg-slate-900 text-slate-100' : 'bg-slate-50 text-slate-900'" class="min-h-screen p-6">
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <Tag class="w-7 h-7 text-indigo-500" />
        <div>
          <h1 class="text-2xl font-bold">Ticket Categories</h1>
          <p class="text-sm text-slate-500">{{ categories.length }} categories</p>
        </div>
      </div>
      <div class="flex gap-2">
        <button @click="load" :class="app.darkMode ? 'bg-slate-700 text-slate-200' : 'bg-white text-slate-700'"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-600 hover:bg-slate-100 dark:hover:bg-slate-600 transition-colors">
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
        </button>
        <button @click="openCreate" class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" /> New Category
        </button>
      </div>
    </div>

    <div v-if="error" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg text-sm">{{ error }}</div>

    <!-- Categories Grid -->
    <div v-if="loading" class="flex items-center justify-center py-16">
      <RefreshCw class="w-6 h-6 animate-spin text-indigo-500" />
    </div>
    <div v-else-if="!categories.length" class="text-center py-16">
      <Tag class="w-12 h-12 text-slate-300 mx-auto mb-3" />
      <p class="text-slate-400">No categories yet</p>
    </div>
    <div v-else class="space-y-3">
      <div v-for="cat in topLevel()" :key="cat.id">
        <!-- Parent category -->
        <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
          class="rounded-xl border shadow-sm">
          <div class="flex items-center justify-between p-4">
            <div class="flex items-center gap-3">
              <div class="w-4 h-4 rounded-full flex-shrink-0" :style="`background-color:${cat.color}`"></div>
              <FolderOpen class="w-5 h-5 text-indigo-500" />
              <div>
                <div class="font-semibold">{{ cat.name }}</div>
                <div class="text-xs text-slate-400">{{ cat.description || 'No description' }}</div>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <span class="text-xs text-slate-400">{{ cat.ticket_count }} tickets</span>
              <span :class="cat.is_active ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-slate-100 text-slate-500 dark:bg-slate-700'"
                class="text-xs px-2 py-0.5 rounded-full">{{ cat.is_active ? 'Active' : 'Inactive' }}</span>
              <div class="flex gap-1">
                <button @click="openEdit(cat)" class="p-1.5 text-slate-400 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors">
                  <Edit class="w-4 h-4" />
                </button>
                <button @click="deleteCategory(cat.id)" class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors">
                  <Trash2 class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
          <!-- Children -->
          <div v-if="children(cat.id).length" class="border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
            <div v-for="child in children(cat.id)" :key="child.id"
              class="flex items-center justify-between px-4 py-3"
              :class="app.darkMode ? 'hover:bg-slate-700/50' : 'hover:bg-slate-50'">
              <div class="flex items-center gap-3 ml-8">
                <ChevronRight class="w-3 h-3 text-slate-400" />
                <div class="w-3 h-3 rounded-full" :style="`background-color:${child.color}`"></div>
                <div>
                  <div class="text-sm font-medium">{{ child.name }}</div>
                  <div class="text-xs text-slate-400">{{ child.description }}</div>
                </div>
              </div>
              <div class="flex items-center gap-3">
                <span class="text-xs text-slate-400">{{ child.ticket_count }} tickets</span>
                <span :class="child.is_active ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' : 'bg-slate-100 text-slate-500 dark:bg-slate-700'"
                  class="text-xs px-2 py-0.5 rounded-full">{{ child.is_active ? 'Active' : 'Inactive' }}</span>
                <div class="flex gap-1">
                  <button @click="openEdit(child)" class="p-1.5 text-slate-400 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors">
                    <Edit class="w-4 h-4" />
                  </button>
                  <button @click="deleteCategory(child.id)" class="p-1.5 text-slate-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Form Modal -->
    <div v-if="showForm" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="w-full max-w-lg rounded-2xl border shadow-xl">
        <div class="flex items-center justify-between p-6 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold">{{ editId ? 'Edit Category' : 'New Category' }}</h2>
          <button @click="showForm = false" class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1">Name *</label>
            <input v-model="form.name" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Description</label>
            <textarea v-model="form.description" rows="2" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none resize-none" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium mb-1">Color</label>
              <input v-model="form.color" type="color"
                class="w-full h-10 rounded-lg border border-slate-200 dark:border-slate-600 cursor-pointer" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Sort Order</label>
              <input v-model.number="form.sort_order" type="number" min="0" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Parent Category</label>
            <select v-model="form.parent_id" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
              <option value="">None (top level)</option>
              <option v-for="c in categories.filter(c => c.id !== editId && !c.parent_id)" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
          </div>
          <label class="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" v-model="form.is_active" class="rounded" />
            <span class="text-sm font-medium">Active</span>
          </label>
        </div>
        <div class="flex justify-end gap-3 p-6 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showForm = false" :class="app.darkMode ? 'bg-slate-700 text-slate-200 hover:bg-slate-600' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
            class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
          <button @click="save" :disabled="saving" class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors flex items-center gap-2">
            <RefreshCw v-if="saving" class="w-4 h-4 animate-spin" />
            {{ editId ? 'Update' : 'Create' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
