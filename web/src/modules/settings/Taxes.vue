<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { settingsAPI } from '@/api/client'
import {
  Percent, Plus, Edit2, X, Save, RefreshCw, Trash2,
  CheckCircle, XCircle, Tag, AlertCircle
} from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(false)
const saving = ref(false)
const taxes = ref<any[]>([])
const showModal = ref(false)
const editingId = ref('')
const filterActive = ref<'all' | 'active' | 'inactive'>('all')

const form = ref({
  name: '', code: '', tax_type: 'percentage',
  rate: 0, description: '', is_active: true
})

const algerianPresets = [
  { name: 'TVA 19%',       code: 'TVA19',  tax_type: 'percentage', rate: 19,   description: 'Taxe sur la Valeur Ajoutée — taux normal' },
  { name: 'TVA 9%',        code: 'TVA9',   tax_type: 'percentage', rate: 9,    description: 'TVA taux réduit' },
  { name: 'TVA Exonéré',   code: 'TVA0',   tax_type: 'percentage', rate: 0,    description: 'TVA taux zéro / exonéré' },
  { name: 'TAP 2%',        code: 'TAP2',   tax_type: 'percentage', rate: 2,    description: 'Taxe sur l\'Activité Professionnelle' },
  { name: 'IRG 10%',       code: 'IRG10',  tax_type: 'percentage', rate: 10,   description: 'Impôt sur le Revenu Global' },
  { name: 'IBS 26%',       code: 'IBS26',  tax_type: 'percentage', rate: 26,   description: 'Impôt sur les Bénéfices des Sociétés' },
  { name: 'Retenue Source', code: 'RAS10', tax_type: 'percentage', rate: 10,   description: 'Retenue à la source 10%' },
]

const filtered = computed(() => {
  let list = taxes.value
  if (filterActive.value === 'active') list = list.filter(t => t.is_active)
  if (filterActive.value === 'inactive') list = list.filter(t => !t.is_active)
  return list
})

async function load() {
  loading.value = true
  try {
    const r = await settingsAPI.getTaxes()
    taxes.value = Array.isArray(r.data) ? r.data : []
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error loading taxes', 'error')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  form.value = { name: '', code: '', tax_type: 'percentage', rate: 0, description: '', is_active: true }
  showModal.value = true
}

function applyPreset(p: any) {
  form.value.name = p.name
  form.value.code = p.code
  form.value.tax_type = p.tax_type
  form.value.rate = p.rate
  form.value.description = p.description
}

function openEdit(t: any) {
  editingId.value = t.id
  form.value = { name: t.name, code: t.code, tax_type: t.tax_type || 'percentage', rate: t.rate, description: t.description || '', is_active: t.is_active }
  showModal.value = true
}

async function save() {
  if (!form.value.name || !form.value.code) {
    app.addToast('Name and code are required', 'error'); return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await settingsAPI.updateTax(editingId.value, form.value)
      app.addToast('Tax updated', 'success')
    } else {
      await settingsAPI.createTax(form.value)
      app.addToast('Tax created', 'success')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error saving tax', 'error')
  } finally {
    saving.value = false
  }
}

async function deleteTax(t: any) {
  if (!confirm(`Delete tax "${t.name}"?`)) return
  try {
    await settingsAPI.deleteTax(t.id)
    app.addToast('Tax deleted', 'success')
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error', 'error')
  }
}

function rateDisplay(t: any) {
  if (t.tax_type === 'fixed') return `${Number(t.rate).toFixed(2)} DZD`
  return `${Number(t.rate).toFixed(2)}%`
}

function rateColor(rate: number) {
  if (rate === 0) return 'text-slate-400'
  if (rate <= 5) return 'text-emerald-400'
  if (rate <= 15) return 'text-sky-400'
  if (rate <= 25) return 'text-amber-400'
  return 'text-red-400'
}

const dk = (d: string, l: string) => app.darkMode ? d : l
onMounted(load)
</script>

<template>
  <div :class="dk('bg-slate-950 text-slate-100','bg-slate-50 text-slate-900')" class="min-h-screen">
    <!-- Header -->
    <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border-b px-6 py-4 sticky top-0 z-10">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-rose-600/20">
            <Percent class="w-5 h-5 text-rose-400" />
          </div>
          <div>
            <h1 class="text-lg font-bold">Taxes</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">Configure tax rates for Algeria and beyond</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
            class="p-2 rounded-lg transition-colors disabled:opacity-50">
            <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          </button>
          <button @click="openCreate"
            class="flex items-center gap-2 px-4 py-2 bg-rose-600 hover:bg-rose-500 text-white rounded-lg text-sm font-medium transition-colors">
            <Plus class="w-4 h-4" /> Add Tax
          </button>
        </div>
      </div>
      <!-- Filter -->
      <div class="mt-3 flex gap-2">
        <button v-for="f in ['all','active','inactive'] as const" :key="f"
          @click="filterActive=f"
          :class="filterActive===f
            ? 'bg-rose-600 text-white'
            : dk('bg-slate-800 text-slate-400 hover:text-slate-200','bg-slate-100 text-slate-600 hover:text-slate-900')"
          class="px-3 py-1.5 rounded-lg text-xs font-medium capitalize transition-colors">
          {{ f }} <span class="ml-1 opacity-70">{{ f==='all' ? taxes.length : f==='active' ? taxes.filter(t=>t.is_active).length : taxes.filter(t=>!t.is_active).length }}</span>
        </button>
      </div>
    </div>

    <!-- Content -->
    <div class="p-6 space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="animate-spin w-8 h-8 text-rose-400" />
      </div>

      <template v-else>
        <!-- Stats -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl p-4">
            <div class="text-2xl font-bold text-rose-400">{{ taxes.length }}</div>
            <div :class="dk('text-slate-500','text-slate-400')" class="text-xs mt-1">Total Taxes</div>
          </div>
          <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl p-4">
            <div class="text-2xl font-bold text-emerald-400">{{ taxes.filter(t=>t.is_active).length }}</div>
            <div :class="dk('text-slate-500','text-slate-400')" class="text-xs mt-1">Active</div>
          </div>
          <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl p-4">
            <div class="text-2xl font-bold text-sky-400">{{ taxes.filter(t=>t.tax_type==='percentage').length }}</div>
            <div :class="dk('text-slate-500','text-slate-400')" class="text-xs mt-1">Percentage</div>
          </div>
          <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl p-4">
            <div class="text-2xl font-bold text-amber-400">{{ taxes.filter(t=>t.tax_type==='fixed').length }}</div>
            <div :class="dk('text-slate-500','text-slate-400')" class="text-xs mt-1">Fixed Amount</div>
          </div>
        </div>

        <div v-if="filtered.length===0" :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
          class="border rounded-xl p-16 text-center">
          <Percent :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12 mx-auto mb-4" />
          <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">No taxes found</p>
        </div>

        <div v-else :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl overflow-hidden">
          <table class="w-full text-sm">
            <thead>
              <tr :class="dk('bg-slate-800/50 text-slate-400 border-slate-800','bg-slate-50 text-slate-500 border-slate-100')"
                class="border-b text-xs uppercase tracking-wider">
                <th class="px-5 py-3 text-left font-medium">Tax Name</th>
                <th class="px-5 py-3 text-left font-medium">Code</th>
                <th class="px-5 py-3 text-left font-medium">Type</th>
                <th class="px-5 py-3 text-right font-medium">Rate</th>
                <th class="px-5 py-3 text-left font-medium">Description</th>
                <th class="px-5 py-3 text-center font-medium">Status</th>
                <th class="px-5 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y" :class="dk('divide-slate-800','divide-slate-100')">
              <tr v-for="t in filtered" :key="t.id"
                :class="dk('hover:bg-slate-800/40','hover:bg-slate-50')" class="transition-colors">
                <td class="px-5 py-3.5">
                  <div class="flex items-center gap-2.5">
                    <div class="p-1.5 rounded-lg bg-rose-500/10">
                      <Percent class="w-3.5 h-3.5 text-rose-400" />
                    </div>
                    <span class="font-semibold">{{ t.name }}</span>
                  </div>
                </td>
                <td class="px-5 py-3.5">
                  <span :class="dk('bg-slate-800 text-slate-300','bg-slate-100 text-slate-600')"
                    class="font-mono text-xs px-2 py-0.5 rounded">{{ t.code }}</span>
                </td>
                <td class="px-5 py-3.5">
                  <span :class="t.tax_type==='percentage' ? 'bg-sky-500/10 text-sky-400' : 'bg-amber-500/10 text-amber-400'"
                    class="text-xs px-2 py-0.5 rounded-full font-medium capitalize">{{ t.tax_type }}</span>
                </td>
                <td class="px-5 py-3.5 text-right">
                  <span :class="['font-mono font-bold text-base', rateColor(t.rate)]">{{ rateDisplay(t) }}</span>
                </td>
                <td class="px-5 py-3.5">
                  <span :class="dk('text-slate-400','text-slate-500')" class="text-xs">{{ t.description || '-' }}</span>
                </td>
                <td class="px-5 py-3.5 text-center">
                  <span :class="t.is_active ? 'text-emerald-400' : 'text-red-400'">
                    <CheckCircle v-if="t.is_active" class="w-4 h-4 mx-auto" />
                    <XCircle v-else class="w-4 h-4 mx-auto" />
                  </span>
                </td>
                <td class="px-5 py-3.5">
                  <div class="flex items-center justify-end gap-1">
                    <button @click="openEdit(t)"
                      :class="dk('hover:bg-slate-700 text-slate-400 hover:text-white','hover:bg-slate-100 text-slate-500 hover:text-slate-900')"
                      class="p-1.5 rounded-lg transition-colors">
                      <Edit2 class="w-3.5 h-3.5" />
                    </button>
                    <button @click="deleteTax(t)"
                      class="p-1.5 rounded-lg hover:bg-red-500/10 text-red-400 transition-colors">
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4"
        style="background:rgba(0,0,0,0.7)" @click.self="showModal=false">
        <div :class="dk('bg-slate-900 border-slate-700','bg-white border-slate-200')"
          class="border rounded-2xl w-full max-w-lg max-h-[90vh] overflow-hidden flex flex-col shadow-2xl">
          <div :class="dk('border-slate-800','border-slate-100')" class="flex items-center justify-between px-6 py-4 border-b flex-shrink-0">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-xl bg-rose-600/20">
                <Percent class="w-4 h-4 text-rose-400" />
              </div>
              <h2 class="font-bold">{{ editingId ? 'Edit Tax' : 'New Tax' }}</h2>
            </div>
            <button @click="showModal=false" :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-900')"
              class="p-1.5 rounded-lg transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="overflow-y-auto flex-1 px-6 py-5 space-y-4">
            <!-- Presets -->
            <div v-if="!editingId">
              <label class="block text-xs font-medium mb-2" :class="dk('text-slate-300','text-slate-700')">
                Algerian Tax Presets
              </label>
              <div class="flex flex-wrap gap-2">
                <button v-for="p in algerianPresets" :key="p.code" type="button"
                  @click="applyPreset(p)"
                  :class="form.code===p.code
                    ? 'bg-rose-600/20 border-rose-500 text-rose-400'
                    : dk('bg-slate-800 border-slate-700 text-slate-400 hover:border-slate-500','bg-slate-50 border-slate-200 text-slate-600 hover:border-slate-300')"
                  class="px-2.5 py-1 border rounded-lg text-xs font-medium transition-all">
                  {{ p.code }}
                </button>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Name <span class="text-red-400">*</span></label>
                <input v-model="form.name" placeholder="TVA 19%"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-500" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Code <span class="text-red-400">*</span></label>
                <input v-model="form.code" placeholder="TVA19" :disabled="!!editingId"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100 font-mono','bg-white border-slate-300 text-slate-900 font-mono')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-500 uppercase disabled:opacity-50" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Type</label>
                <select v-model="form.tax_type"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-500">
                  <option value="percentage">Percentage (%)</option>
                  <option value="fixed">Fixed Amount (DZD)</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">
                  Rate {{ form.tax_type === 'percentage' ? '(%)' : '(DZD)' }}
                </label>
                <input v-model.number="form.rate" type="number" step="0.01" min="0"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-500" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Description</label>
              <textarea v-model="form.description" rows="2" placeholder="Tax description..."
                :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-rose-500 resize-none" />
            </div>
            <div class="flex items-center gap-3">
              <button type="button" @click="form.is_active = !form.is_active"
                :class="form.is_active ? 'bg-emerald-500' : 'bg-slate-600'"
                class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors">
                <span :class="form.is_active ? 'translate-x-6' : 'translate-x-1'"
                  class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
              </button>
              <span class="text-sm" :class="dk('text-slate-300','text-slate-700')">Tax is Active</span>
            </div>
          </div>
          <div :class="dk('border-slate-800 bg-slate-900/50','border-slate-100 bg-slate-50')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t flex-shrink-0">
            <button @click="showModal=false"
              :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
              class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="save" :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 bg-rose-600 hover:bg-rose-500 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors">
              <Save class="w-4 h-4" />
              {{ saving ? 'Saving...' : (editingId ? 'Update Tax' : 'Create Tax') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
