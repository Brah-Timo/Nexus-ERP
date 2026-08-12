<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { settingsAPI } from '@/api/client'
import {
  DollarSign, Plus, Edit2, X, Save, RefreshCw, Trash2,
  TrendingUp, Star, CheckCircle, XCircle, AlertCircle
} from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(false)
const saving = ref(false)
const currencies = ref<any[]>([])
const showModal = ref(false)
const editingId = ref('')

const form = ref({
  code: '', name: '', symbol: '',
  exchange_rate: 1, is_base: false, is_active: true
})

const globalCurrencies = [
  { code: 'DZD', name: 'Algerian Dinar', symbol: 'DA' },
  { code: 'EUR', name: 'Euro', symbol: '€' },
  { code: 'USD', name: 'US Dollar', symbol: '$' },
  { code: 'GBP', name: 'British Pound', symbol: '£' },
  { code: 'SAR', name: 'Saudi Riyal', symbol: 'SR' },
  { code: 'MAD', name: 'Moroccan Dirham', symbol: 'MAD' },
  { code: 'TND', name: 'Tunisian Dinar', symbol: 'TND' },
  { code: 'XOF', name: 'CFA Franc', symbol: 'FCFA' },
  { code: 'JPY', name: 'Japanese Yen', symbol: '¥' },
  { code: 'CNY', name: 'Chinese Yuan', symbol: '¥' },
  { code: 'CHF', name: 'Swiss Franc', symbol: 'CHF' },
  { code: 'CAD', name: 'Canadian Dollar', symbol: 'CA$' },
]

const baseCurrency = computed(() => currencies.value.find(c => c.is_base))

async function load() {
  loading.value = true
  try {
    const r = await settingsAPI.getCurrencies()
    currencies.value = Array.isArray(r.data) ? r.data : []
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error loading currencies', 'error')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = ''
  form.value = { code: '', name: '', symbol: '', exchange_rate: 1, is_base: false, is_active: true }
  showModal.value = true
}

function openEdit(c: any) {
  editingId.value = c.id
  form.value = { code: c.code, name: c.name, symbol: c.symbol, exchange_rate: c.exchange_rate, is_base: c.is_base, is_active: c.is_active }
  showModal.value = true
}

function selectGlobal(g: any) {
  form.value.code = g.code
  form.value.name = g.name
  form.value.symbol = g.symbol
}

async function save() {
  if (!form.value.code || !form.value.name) {
    app.addToast('Code and name are required', 'error'); return
  }
  saving.value = true
  try {
    if (editingId.value) {
      await settingsAPI.updateCurrency(editingId.value, form.value)
      app.addToast('Currency updated', 'success')
    } else {
      await settingsAPI.createCurrency(form.value)
      app.addToast('Currency added', 'success')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error saving currency', 'error')
  } finally {
    saving.value = false
  }
}

async function deleteCurrency(c: any) {
  if (c.is_base) { app.addToast('Cannot delete base currency', 'error'); return }
  if (!confirm(`Remove currency "${c.code}"?`)) return
  try {
    await settingsAPI.deleteCurrency(c.id)
    app.addToast('Currency removed', 'success')
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error', 'error')
  }
}

function rateColor(rate: number) {
  if (rate > 100) return 'text-amber-400'
  if (rate > 1) return 'text-sky-400'
  return 'text-emerald-400'
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
          <div class="p-2 rounded-xl bg-emerald-600/20">
            <DollarSign class="w-5 h-5 text-emerald-400" />
          </div>
          <div>
            <h1 class="text-lg font-bold">Currencies</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">Manage currencies and exchange rates</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
            class="p-2 rounded-lg transition-colors disabled:opacity-50">
            <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          </button>
          <button @click="openCreate"
            class="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-sm font-medium transition-colors">
            <Plus class="w-4 h-4" /> Add Currency
          </button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="p-6 space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="animate-spin w-8 h-8 text-emerald-400" />
      </div>

      <template v-else>
        <!-- Base currency card -->
        <div v-if="baseCurrency" :class="dk('bg-emerald-900/20 border-emerald-800','bg-emerald-50 border-emerald-200')"
          class="border rounded-xl p-5 flex items-center gap-5">
          <div class="p-3 rounded-2xl bg-emerald-600/20">
            <Star class="w-6 h-6 text-emerald-400" />
          </div>
          <div class="flex-1">
            <div class="flex items-center gap-2 mb-1">
              <span class="font-bold text-lg">{{ baseCurrency.code }}</span>
              <span :class="dk('text-slate-300','text-slate-700')">{{ baseCurrency.name }}</span>
              <span class="text-xs px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 font-medium">Base Currency</span>
            </div>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">
              All exchange rates are defined relative to {{ baseCurrency.code }} ({{ baseCurrency.symbol }})
            </p>
          </div>
          <div class="text-3xl font-mono font-bold text-emerald-400">{{ baseCurrency.symbol }}</div>
        </div>

        <!-- Currency grid -->
        <div v-if="currencies.length === 0" :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
          class="border rounded-xl p-16 text-center">
          <DollarSign :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12 mx-auto mb-4" />
          <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">No currencies configured</p>
        </div>

        <div v-else :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl overflow-hidden">
          <table class="w-full text-sm">
            <thead>
              <tr :class="dk('bg-slate-800/50 text-slate-400 border-slate-800','bg-slate-50 text-slate-500 border-slate-100')"
                class="border-b text-xs uppercase tracking-wider">
                <th class="px-5 py-3 text-left font-medium">Currency</th>
                <th class="px-5 py-3 text-left font-medium">Name</th>
                <th class="px-5 py-3 text-right font-medium">Exchange Rate</th>
                <th class="px-5 py-3 text-center font-medium">Type</th>
                <th class="px-5 py-3 text-center font-medium">Status</th>
                <th class="px-5 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y" :class="dk('divide-slate-800','divide-slate-100')">
              <tr v-for="c in currencies" :key="c.id"
                :class="dk('hover:bg-slate-800/40','hover:bg-slate-50')" class="transition-colors">
                <td class="px-5 py-3.5">
                  <div class="flex items-center gap-3">
                    <div :class="dk('bg-slate-800 border-slate-700','bg-slate-100 border-slate-200')"
                      class="w-10 h-10 rounded-xl border flex items-center justify-center font-bold text-sm">
                      {{ c.symbol || c.code.slice(0,2) }}
                    </div>
                    <span class="font-mono font-bold text-base">{{ c.code }}</span>
                  </div>
                </td>
                <td class="px-5 py-3.5" :class="dk('text-slate-300','text-slate-700')">{{ c.name }}</td>
                <td class="px-5 py-3.5 text-right">
                  <span :class="[rateColor(c.exchange_rate), 'font-mono font-semibold text-sm']">
                    {{ c.is_base ? '1.000000' : c.exchange_rate?.toFixed(6) }}
                  </span>
                  <span v-if="!c.is_base" :class="dk('text-slate-500','text-slate-400')" class="text-xs ml-1">
                    per {{ baseCurrency?.code || 'DZD' }}
                  </span>
                </td>
                <td class="px-5 py-3.5 text-center">
                  <span v-if="c.is_base" class="text-xs px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 font-medium">
                    Base
                  </span>
                  <span v-else class="text-xs px-2 py-0.5 rounded-full bg-sky-500/10 text-sky-400 font-medium">
                    Foreign
                  </span>
                </td>
                <td class="px-5 py-3.5 text-center">
                  <span :class="c.is_active ? 'text-emerald-400' : 'text-red-400'">
                    <CheckCircle v-if="c.is_active" class="w-4 h-4 mx-auto" />
                    <XCircle v-else class="w-4 h-4 mx-auto" />
                  </span>
                </td>
                <td class="px-5 py-3.5">
                  <div class="flex items-center justify-end gap-1">
                    <button @click="openEdit(c)"
                      :class="dk('hover:bg-slate-700 text-slate-400 hover:text-white','hover:bg-slate-100 text-slate-500 hover:text-slate-900')"
                      class="p-1.5 rounded-lg transition-colors" title="Edit rate">
                      <Edit2 class="w-3.5 h-3.5" />
                    </button>
                    <button v-if="!c.is_base" @click="deleteCurrency(c)"
                      class="p-1.5 rounded-lg hover:bg-red-500/10 text-red-400 transition-colors" title="Remove">
                      <Trash2 class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Info -->
        <div :class="dk('bg-slate-900/50 border-slate-800','bg-slate-100 border-slate-200')"
          class="border rounded-xl p-4 flex items-start gap-3">
          <AlertCircle class="w-4 h-4 text-amber-400 flex-shrink-0 mt-0.5" />
          <p :class="dk('text-slate-400','text-slate-600')" class="text-xs leading-relaxed">
            Exchange rates are used in multi-currency transactions. The base currency rate is always 1.0. Update rates regularly to ensure accurate financial reporting.
          </p>
        </div>
      </template>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4"
        style="background:rgba(0,0,0,0.7)" @click.self="showModal=false">
        <div :class="dk('bg-slate-900 border-slate-700','bg-white border-slate-200')"
          class="border rounded-2xl w-full max-w-md shadow-2xl">
          <div :class="dk('border-slate-800','border-slate-100')" class="flex items-center justify-between px-6 py-4 border-b">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-xl bg-emerald-600/20">
                <DollarSign class="w-4 h-4 text-emerald-400" />
              </div>
              <h2 class="font-bold">{{ editingId ? 'Edit Currency' : 'Add Currency' }}</h2>
            </div>
            <button @click="showModal=false" :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-900')"
              class="p-1.5 rounded-lg transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5 space-y-4">
            <!-- Quick select -->
            <div v-if="!editingId">
              <label class="block text-xs font-medium mb-2" :class="dk('text-slate-300','text-slate-700')">Quick Select</label>
              <div class="flex flex-wrap gap-2">
                <button v-for="g in globalCurrencies" :key="g.code" type="button"
                  @click="selectGlobal(g)"
                  :class="form.code===g.code
                    ? 'bg-emerald-600/20 border-emerald-500 text-emerald-400'
                    : dk('bg-slate-800 border-slate-700 text-slate-400 hover:border-slate-500','bg-slate-50 border-slate-200 text-slate-600 hover:border-slate-400')"
                  class="px-2.5 py-1 border rounded-lg text-xs font-mono font-medium transition-all">
                  {{ g.code }}
                </button>
              </div>
            </div>
            <div class="grid grid-cols-3 gap-3">
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Code <span class="text-red-400">*</span></label>
                <input v-model="form.code" :disabled="!!editingId" placeholder="USD" maxlength="10"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100 font-mono','bg-white border-slate-300 text-slate-900 font-mono')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 uppercase disabled:opacity-50" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Symbol</label>
                <input v-model="form.symbol" placeholder="$"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Rate</label>
                <input v-model.number="form.exchange_rate" type="number" step="0.000001" min="0"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Name <span class="text-red-400">*</span></label>
              <input v-model="form.name" placeholder="US Dollar"
                :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500" />
            </div>
            <div class="flex items-center gap-4 pt-1">
              <div class="flex items-center gap-2.5">
                <button type="button" @click="form.is_active = !form.is_active"
                  :class="form.is_active ? 'bg-emerald-500' : 'bg-slate-600'"
                  class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors flex-shrink-0">
                  <span :class="form.is_active ? 'translate-x-4.5' : 'translate-x-1'"
                    class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform" />
                </button>
                <span class="text-xs" :class="dk('text-slate-300','text-slate-700')">Active</span>
              </div>
              <div class="flex items-center gap-2.5">
                <button type="button" @click="form.is_base = !form.is_base"
                  :class="form.is_base ? 'bg-amber-500' : 'bg-slate-600'"
                  class="relative inline-flex h-5 w-9 items-center rounded-full transition-colors flex-shrink-0">
                  <span :class="form.is_base ? 'translate-x-4.5' : 'translate-x-1'"
                    class="inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform" />
                </button>
                <span class="text-xs" :class="dk('text-slate-300','text-slate-700')">Base Currency</span>
              </div>
            </div>
          </div>
          <div :class="dk('border-slate-800 bg-slate-900/50','border-slate-100 bg-slate-50')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showModal=false"
              :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
              class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="save" :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors">
              <Save class="w-4 h-4" />
              {{ saving ? 'Saving...' : (editingId ? 'Update' : 'Add Currency') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
