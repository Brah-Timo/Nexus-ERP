<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { settingsAPI } from '@/api/client'
import { Hash, Save, RefreshCw, Eye, RotateCcw, Info, ChevronDown, ChevronUp } from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(false)
const saving = ref(false)
const configs = ref<any[]>([])
const expandedDoc = ref<string | null>(null)

interface DocConfig {
  id?: string
  doc_type: string
  prefix: string
  suffix: string
  next_number: number
  padding: number
  reset_yearly: boolean
  _dirty?: boolean
}

const docTypeLabels: Record<string, string> = {
  sales_invoice:     'Sales Invoice',
  sales_order:       'Sales Order',
  quotation:         'Quotation / Devis',
  purchase_order:    'Purchase Order',
  purchase_invoice:  'Purchase Invoice',
  goods_receipt:     'Goods Receipt Note',
  payment:           'Payment Voucher',
  receipt:           'Receipt',
  journal_entry:     'Journal Entry',
  manufacturing_order: 'Manufacturing Order',
  stock_movement:    'Stock Movement',
  inventory_count:   'Inventory Count',
  expense_report:    'Expense Report',
  asset:             'Fixed Asset',
}

const docTypeIcons: Record<string, string> = {
  sales_invoice: 'text-emerald-400',
  sales_order: 'text-sky-400',
  quotation: 'text-blue-400',
  purchase_order: 'text-violet-400',
  purchase_invoice: 'text-purple-400',
  goods_receipt: 'text-amber-400',
  payment: 'text-rose-400',
  receipt: 'text-pink-400',
  journal_entry: 'text-indigo-400',
  manufacturing_order: 'text-orange-400',
  stock_movement: 'text-teal-400',
  inventory_count: 'text-cyan-400',
  expense_report: 'text-yellow-400',
  asset: 'text-slate-400',
}

const docGroups = [
  { label: 'Sales', keys: ['sales_invoice','sales_order','quotation'] },
  { label: 'Purchase', keys: ['purchase_order','purchase_invoice','goods_receipt'] },
  { label: 'Finance', keys: ['payment','receipt','journal_entry'] },
  { label: 'Operations', keys: ['manufacturing_order','stock_movement','inventory_count','expense_report','asset'] },
]

async function load() {
  loading.value = true
  try {
    const r = await settingsAPI.getNumbering()
    configs.value = Array.isArray(r.data) ? r.data : []
    // Add missing doc types with defaults
    for (const key of Object.keys(docTypeLabels)) {
      if (!configs.value.find(c => c.doc_type === key)) {
        configs.value.push({
          doc_type: key, prefix: defaultPrefix(key), suffix: '',
          next_number: 1, padding: 5, reset_yearly: true
        })
      }
    }
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error loading numbering config', 'error')
  } finally {
    loading.value = false
  }
}

function defaultPrefix(docType: string) {
  const map: Record<string, string> = {
    sales_invoice: 'INV-', sales_order: 'SO-', quotation: 'QUO-',
    purchase_order: 'PO-', purchase_invoice: 'PI-', goods_receipt: 'GRN-',
    payment: 'PAY-', receipt: 'REC-', journal_entry: 'JE-',
    manufacturing_order: 'MO-', stock_movement: 'SM-',
    inventory_count: 'IC-', expense_report: 'EXP-', asset: 'AST-',
  }
  return map[docType] || ''
}

function getConfig(key: string): DocConfig {
  return configs.value.find(c => c.doc_type === key) || {
    doc_type: key, prefix: defaultPrefix(key), suffix: '',
    next_number: 1, padding: 5, reset_yearly: true
  }
}

function preview(cfg: DocConfig): string {
  const year = new Date().getFullYear()
  const num = String(cfg.next_number || 1).padStart(cfg.padding || 5, '0')
  const yearPart = cfg.reset_yearly ? `${year}-` : ''
  return `${cfg.prefix || ''}${yearPart}${num}${cfg.suffix || ''}`
}

function markDirty(key: string) {
  const cfg = configs.value.find(c => c.doc_type === key)
  if (cfg) cfg._dirty = true
}

function resetDoc(key: string) {
  const cfg = configs.value.find(c => c.doc_type === key)
  if (cfg) {
    cfg.prefix = defaultPrefix(key)
    cfg.suffix = ''
    cfg.next_number = 1
    cfg.padding = 5
    cfg.reset_yearly = true
    cfg._dirty = false
  }
}

async function save() {
  saving.value = true
  try {
    const payload = configs.value.map(c => ({
      doc_type: c.doc_type,
      prefix: c.prefix || '',
      suffix: c.suffix || '',
      next_number: c.next_number || 1,
      padding: c.padding || 5,
      reset_yearly: !!c.reset_yearly,
    }))
    await settingsAPI.updateNumbering(payload)
    configs.value.forEach(c => c._dirty = false)
    app.addToast('Numbering configuration saved', 'success')
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error saving config', 'error')
  } finally {
    saving.value = false
  }
}

const dirtyCount = computed(() => configs.value.filter(c => c._dirty).length)

const dk = (d: string, l: string) => app.darkMode ? d : l
onMounted(load)
</script>

<template>
  <div :class="dk('bg-slate-950 text-slate-100','bg-slate-50 text-slate-900')" class="min-h-screen">
    <!-- Header -->
    <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border-b px-6 py-4 sticky top-0 z-10">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-indigo-600/20">
            <Hash class="w-5 h-5 text-indigo-400" />
          </div>
          <div>
            <h1 class="text-lg font-bold">Document Numbering</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">Configure automatic numbering sequences for all document types</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
            class="p-2 rounded-lg transition-colors disabled:opacity-50">
            <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          </button>
          <button @click="save" :disabled="saving || dirtyCount===0"
            class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors">
            <Save class="w-4 h-4" />
            {{ saving ? 'Saving...' : `Save Changes${dirtyCount > 0 ? ` (${dirtyCount})` : ''}` }}
          </button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="p-6">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="animate-spin w-8 h-8 text-indigo-400" />
      </div>

      <div v-else class="space-y-6">
        <!-- Info banner -->
        <div :class="dk('bg-indigo-900/20 border-indigo-800','bg-indigo-50 border-indigo-200')"
          class="border rounded-xl p-4 flex items-start gap-3">
          <Info class="w-4 h-4 text-indigo-400 flex-shrink-0 mt-0.5" />
          <p :class="dk('text-indigo-300','text-indigo-700')" class="text-xs leading-relaxed">
            Document numbers are generated automatically using these patterns. With "Reset Yearly" enabled, the counter resets each year and the year is included in the number. Example: <code class="font-mono bg-indigo-500/20 px-1 rounded">INV-2026-00001</code>
          </p>
        </div>

        <!-- Groups -->
        <div v-for="group in docGroups" :key="group.label" :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
          class="border rounded-xl overflow-hidden">
          <!-- Group header -->
          <div :class="dk('bg-slate-800/50 border-slate-800','bg-slate-50 border-slate-100')"
            class="border-b px-5 py-3">
            <h3 class="font-semibold text-sm" :class="dk('text-slate-200','text-slate-800')">{{ group.label }}</h3>
          </div>

          <!-- Rows -->
          <div class="divide-y" :class="dk('divide-slate-800','divide-slate-100')">
            <div v-for="key in group.keys" :key="key" class="transition-all">
              <!-- Row header (clickable) -->
              <div @click="expandedDoc = expandedDoc===key ? null : key"
                :class="[dk('hover:bg-slate-800/40','hover:bg-slate-50'), getConfig(key)._dirty ? dk('border-l-2 border-indigo-500','border-l-2 border-indigo-400') : '']"
                class="flex items-center gap-4 px-5 py-3.5 cursor-pointer transition-colors">
                <div :class="docTypeIcons[key] || 'text-slate-400'" class="p-1.5 rounded-lg"
                  :style="{ background: 'currentColor', opacity: 0.1 }">
                </div>
                <div class="p-1.5 rounded-lg bg-current/10">
                  <Hash :class="docTypeIcons[key] || 'text-slate-400'" class="w-3.5 h-3.5" />
                </div>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <span class="font-medium text-sm">{{ docTypeLabels[key] || key }}</span>
                    <span v-if="getConfig(key)._dirty" class="w-1.5 h-1.5 rounded-full bg-indigo-400 flex-shrink-0"></span>
                  </div>
                  <div :class="dk('text-slate-400','text-slate-500')" class="text-xs font-mono mt-0.5">
                    {{ preview(getConfig(key)) }}
                  </div>
                </div>
                <div class="flex items-center gap-3">
                  <span v-if="getConfig(key).reset_yearly"
                    :class="dk('bg-sky-500/10 text-sky-400','bg-sky-50 text-sky-600')"
                    class="text-xs px-2 py-0.5 rounded-full">Yearly Reset</span>
                  <ChevronDown v-if="expandedDoc!==key" :class="dk('text-slate-500','text-slate-400')" class="w-4 h-4" />
                  <ChevronUp v-else :class="dk('text-slate-500','text-slate-400')" class="w-4 h-4" />
                </div>
              </div>

              <!-- Expanded editor -->
              <Transition name="expand">
                <div v-if="expandedDoc===key"
                  :class="dk('bg-slate-800/30 border-slate-800','bg-slate-50/80 border-slate-100')"
                  class="border-t px-5 py-4">
                  <div class="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-5 gap-4 items-end">
                    <div>
                      <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-400','text-slate-600')">Prefix</label>
                      <input :value="getConfig(key).prefix"
                        @input="e => { const c = getConfig(key); c.prefix = (e.target as HTMLInputElement).value; markDirty(key) }"
                        placeholder="INV-"
                        :class="dk('bg-slate-900 border-slate-700 text-slate-100 font-mono','bg-white border-slate-300 text-slate-900 font-mono')"
                        class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                    </div>
                    <div>
                      <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-400','text-slate-600')">Suffix</label>
                      <input :value="getConfig(key).suffix"
                        @input="e => { const c = getConfig(key); c.suffix = (e.target as HTMLInputElement).value; markDirty(key) }"
                        placeholder="-DZD"
                        :class="dk('bg-slate-900 border-slate-700 text-slate-100 font-mono','bg-white border-slate-300 text-slate-900 font-mono')"
                        class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                    </div>
                    <div>
                      <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-400','text-slate-600')">Next Number</label>
                      <input type="number" min="1" :value="getConfig(key).next_number"
                        @input="e => { const c = getConfig(key); c.next_number = parseInt((e.target as HTMLInputElement).value) || 1; markDirty(key) }"
                        :class="dk('bg-slate-900 border-slate-700 text-slate-100','bg-white border-slate-300 text-slate-900')"
                        class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                    </div>
                    <div>
                      <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-400','text-slate-600')">Zero Padding</label>
                      <select :value="getConfig(key).padding"
                        @change="e => { const c = getConfig(key); c.padding = parseInt((e.target as HTMLSelectElement).value); markDirty(key) }"
                        :class="dk('bg-slate-900 border-slate-700 text-slate-100','bg-white border-slate-300 text-slate-900')"
                        class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                        <option value="3">3 digits (001)</option>
                        <option value="4">4 digits (0001)</option>
                        <option value="5">5 digits (00001)</option>
                        <option value="6">6 digits (000001)</option>
                      </select>
                    </div>
                    <div class="flex items-end gap-3">
                      <div class="flex-1">
                        <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-400','text-slate-600')">Reset Yearly</label>
                        <button type="button"
                          @click="() => { const c = getConfig(key); c.reset_yearly = !c.reset_yearly; markDirty(key) }"
                          :class="getConfig(key).reset_yearly ? 'bg-indigo-500' : 'bg-slate-600'"
                          class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors">
                          <span :class="getConfig(key).reset_yearly ? 'translate-x-6' : 'translate-x-1'"
                            class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
                        </button>
                      </div>
                      <button @click="resetDoc(key)"
                        :class="dk('text-slate-400 hover:text-white hover:bg-slate-700','text-slate-500 hover:text-slate-900 hover:bg-slate-200')"
                        class="p-2 rounded-lg transition-colors flex-shrink-0" title="Reset to defaults">
                        <RotateCcw class="w-4 h-4" />
                      </button>
                    </div>
                  </div>

                  <!-- Preview -->
                  <div :class="dk('bg-slate-900 border-slate-700','bg-white border-slate-200')"
                    class="mt-3 border rounded-lg px-4 py-2.5 flex items-center gap-2">
                    <Eye class="w-3.5 h-3.5 text-indigo-400 flex-shrink-0" />
                    <span :class="dk('text-slate-400','text-slate-500')" class="text-xs">Preview: </span>
                    <span class="font-mono font-bold text-sm text-indigo-400">{{ preview(getConfig(key)) }}</span>
                  </div>
                </div>
              </Transition>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.expand-enter-active, .expand-leave-active { transition: all 0.2s ease; overflow: hidden; }
.expand-enter-from, .expand-leave-to { max-height: 0; opacity: 0; }
.expand-enter-to, .expand-leave-from { max-height: 400px; opacity: 1; }
</style>
