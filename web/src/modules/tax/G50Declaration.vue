<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
          <ScrollText class="w-6 h-6 text-indigo-500" />
          G50 Tax Declaration
        </h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
          Monthly Algerian DGI tax declaration — TVA, TAP, IRG, Stamp Tax
        </p>
      </div>
      <div class="flex items-center gap-2">
        <!-- Period selector -->
        <div class="flex items-center gap-1.5 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2">
          <Calendar class="w-4 h-4 text-gray-400" />
          <select v-model="selectedMonth" @change="loadG50"
            class="bg-transparent text-sm text-gray-700 dark:text-gray-200 outline-none cursor-pointer">
            <option v-for="m in months" :key="m.value" :value="m.value">{{ m.label }}</option>
          </select>
          <select v-model="selectedYear" @change="loadG50"
            class="bg-transparent text-sm text-gray-700 dark:text-gray-200 outline-none cursor-pointer ml-1">
            <option v-for="y in years" :key="y" :value="y">{{ y }}</option>
          </select>
        </div>
        <button @click="loadG50"
          class="inline-flex items-center gap-2 px-3 py-2 border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 text-sm rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
          <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
        </button>
        <button @click="openSave"
          class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-semibold rounded-lg transition-colors shadow-sm">
          <Save class="w-4 h-4" /> Save Declaration
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-24">
      <Loader2 class="w-10 h-10 text-indigo-500 animate-spin" />
    </div>

    <template v-else-if="computed">

      <!-- ── Period Banner ── -->
      <div class="bg-gradient-to-r from-indigo-600 to-indigo-800 rounded-xl p-5 text-white flex items-center justify-between flex-wrap gap-4">
        <div class="flex items-center gap-4">
          <div class="w-14 h-14 bg-white/20 rounded-xl flex items-center justify-center">
            <ScrollText class="w-7 h-7" />
          </div>
          <div>
            <p class="text-indigo-200 text-xs font-medium uppercase tracking-wide">G50 Declaration</p>
            <p class="text-2xl font-bold mt-0.5">{{ monthName(selectedMonth) }} {{ selectedYear }}</p>
            <p class="text-indigo-200 text-sm">Auto-computed from invoices</p>
          </div>
        </div>
        <div class="text-right">
          <p class="text-indigo-200 text-xs">Total Tax Due</p>
          <p class="text-3xl font-bold">{{ fmtShort(computed.total_tax_due) }}</p>
          <p class="text-indigo-200 text-sm">{{ fmtCurrency(computed.total_tax_due) }}</p>
        </div>
      </div>

      <!-- ── Saved Declaration Status ── -->
      <div v-if="existingDecl"
        class="flex items-center gap-3 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-5 py-3">
        <div class="w-9 h-9 rounded-xl flex items-center justify-center"
          :class="existingDecl.status === 'submitted' || existingDecl.status === 'accepted'
            ? 'bg-emerald-100 dark:bg-emerald-900/30'
            : 'bg-amber-100 dark:bg-amber-900/30'">
          <component :is="existingDecl.status === 'submitted' || existingDecl.status === 'accepted' ? CheckCircle : Clock"
            :class="existingDecl.status === 'submitted' || existingDecl.status === 'accepted'
              ? 'text-emerald-600 dark:text-emerald-400'
              : 'text-amber-600 dark:text-amber-400'"
            class="w-5 h-5" />
        </div>
        <div class="flex-1">
          <p class="text-sm font-semibold text-gray-900 dark:text-white">
            Saved: <span class="font-mono text-indigo-600 dark:text-indigo-400">{{ existingDecl.reference }}</span>
          </p>
          <p class="text-xs text-gray-400">Status: <span :class="statusColor(existingDecl.status)" class="font-semibold capitalize">{{ existingDecl.status }}</span></p>
        </div>
        <span :class="statusBadge(existingDecl.status)" class="px-3 py-1 rounded-full text-xs font-bold capitalize">
          {{ existingDecl.status }}
        </span>
      </div>

      <!-- ── Main G50 Form Grid ── -->
      <div class="grid grid-cols-1 xl:grid-cols-2 gap-6">

        <!-- TVA Section -->
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
          <div class="bg-blue-50 dark:bg-blue-900/20 px-5 py-3 border-b border-blue-100 dark:border-blue-900/40 flex items-center gap-2">
            <Percent class="w-4 h-4 text-blue-600 dark:text-blue-400" />
            <h3 class="font-bold text-blue-800 dark:text-blue-300 text-sm">TVA — Taxe sur la Valeur Ajoutée</h3>
          </div>
          <div class="p-5 space-y-4">
            <div class="flex justify-between items-center py-2 border-b border-gray-100 dark:border-gray-800">
              <span class="text-sm text-gray-500">TVA Collected (Sales)</span>
              <span class="font-bold text-gray-900 dark:text-white">{{ fmtCurrency(form.tva_collected) }}</span>
            </div>
            <div class="flex justify-between items-center py-2 border-b border-gray-100 dark:border-gray-800">
              <span class="text-sm text-gray-500">TVA Deductible (Purchases)</span>
              <span class="font-bold text-gray-900 dark:text-white">{{ fmtCurrency(form.tva_deductible) }}</span>
            </div>
            <div class="flex justify-between items-center py-2 border-b border-gray-100 dark:border-gray-800">
              <span class="text-sm text-gray-500">Credit Brought Forward</span>
              <div class="flex items-center gap-2">
                <input v-model.number="form.tva_credit_bf" type="number" min="0" step="0.01"
                  class="w-32 text-right field text-xs" />
              </div>
            </div>
            <div class="flex justify-between items-center py-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg px-3">
              <span class="text-sm font-semibold text-blue-800 dark:text-blue-300">TVA Net Due</span>
              <span :class="form.tva_net_due >= 0 ? 'text-blue-700 dark:text-blue-400' : 'text-emerald-600'"
                class="font-bold text-lg">
                {{ fmtCurrency(form.tva_net_due) }}
              </span>
            </div>
            <div v-if="form.tva_credit_carry > 0"
              class="flex justify-between items-center py-2 bg-emerald-50 dark:bg-emerald-900/20 rounded-lg px-3">
              <span class="text-sm font-semibold text-emerald-700 dark:text-emerald-300">Credit Carried Forward</span>
              <span class="font-bold text-emerald-700 dark:text-emerald-400">{{ fmtCurrency(form.tva_credit_carry) }}</span>
            </div>
          </div>
        </div>

        <!-- TAP Section -->
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
          <div class="bg-orange-50 dark:bg-orange-900/20 px-5 py-3 border-b border-orange-100 dark:border-orange-900/40 flex items-center gap-2">
            <TrendingUp class="w-4 h-4 text-orange-600 dark:text-orange-400" />
            <h3 class="font-bold text-orange-800 dark:text-orange-300 text-sm">TAP — Taxe sur l'Activité Professionnelle</h3>
          </div>
          <div class="p-5 space-y-4">
            <div>
              <label class="lbl">TAP Base (Revenue)</label>
              <input v-model.number="form.tap_base" type="number" min="0" step="0.01" class="field" />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="lbl">TAP Rate</label>
                <select v-model.number="form.tap_rate" class="field">
                  <option :value="0.02">2% — Standard</option>
                  <option :value="0.00">0% — Export</option>
                  <option :value="0.01">1% — Reduced</option>
                  <option :value="0.03">3% — Other</option>
                </select>
              </div>
              <div>
                <label class="lbl">TAP Reduction</label>
                <input v-model.number="form.tap_reduction" type="number" min="0" step="0.01" class="field" />
              </div>
            </div>
            <div class="flex justify-between items-center py-2 bg-orange-50 dark:bg-orange-900/20 rounded-lg px-3">
              <span class="text-sm font-semibold text-orange-800 dark:text-orange-300">TAP Net Due</span>
              <span class="font-bold text-lg text-orange-700 dark:text-orange-400">
                {{ fmtCurrency(tapNetDue) }}
              </span>
            </div>
          </div>
        </div>

        <!-- Withholding IRG Section -->
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
          <div class="bg-violet-50 dark:bg-violet-900/20 px-5 py-3 border-b border-violet-100 dark:border-violet-900/40 flex items-center gap-2">
            <Users class="w-4 h-4 text-violet-600 dark:text-violet-400" />
            <h3 class="font-bold text-violet-800 dark:text-violet-300 text-sm">IRG — Retenues à la Source</h3>
          </div>
          <div class="p-5 space-y-4">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="lbl">IRG on Wages</label>
                <input v-model.number="form.irg_wages_amount" type="number" min="0" step="0.01" class="field" />
              </div>
              <div>
                <label class="lbl">IRG on Fees</label>
                <input v-model.number="form.irg_fees_amount" type="number" min="0" step="0.01" class="field" />
              </div>
            </div>
            <div class="flex justify-between items-center py-2 bg-violet-50 dark:bg-violet-900/20 rounded-lg px-3">
              <span class="text-sm font-semibold text-violet-800 dark:text-violet-300">IRG Total</span>
              <span class="font-bold text-lg text-violet-700 dark:text-violet-400">
                {{ fmtCurrency(form.irg_wages_amount + form.irg_fees_amount) }}
              </span>
            </div>
          </div>
        </div>

        <!-- Stamp Tax + Notes -->
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
          <div class="bg-red-50 dark:bg-red-900/20 px-5 py-3 border-b border-red-100 dark:border-red-900/40 flex items-center gap-2">
            <Stamp class="w-4 h-4 text-red-600 dark:text-red-400" />
            <h3 class="font-bold text-red-800 dark:text-red-300 text-sm">Stamp Tax &amp; Other Taxes</h3>
          </div>
          <div class="p-5 space-y-4">
            <div>
              <label class="lbl">Stamp Tax (Timbre Fiscal)</label>
              <input v-model.number="form.stamp_tax_amount" type="number" min="0" step="0.01" class="field" />
            </div>
            <div>
              <label class="lbl">Notes</label>
              <textarea v-model="form.notes" rows="3" class="field resize-none"
                placeholder="Optional notes for this declaration..."></textarea>
            </div>
            <div class="flex justify-between items-center py-2 bg-red-50 dark:bg-red-900/20 rounded-lg px-3">
              <span class="text-sm font-semibold text-red-700 dark:text-red-300">Stamp Tax Total</span>
              <span class="font-bold text-lg text-red-700 dark:text-red-400">
                {{ fmtCurrency(form.stamp_tax_amount) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- ── Grand Total Summary ── -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-800 flex items-center gap-2">
          <Calculator class="w-4 h-4 text-indigo-500" />
          <h3 class="font-bold text-gray-900 dark:text-white">G50 Summary — {{ monthName(selectedMonth) }} {{ selectedYear }}</h3>
        </div>
        <div class="p-5">
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-5">
            <div class="bg-blue-50 dark:bg-blue-900/20 rounded-xl p-4 text-center">
              <p class="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">TVA Net</p>
              <p class="text-xl font-bold text-blue-700 dark:text-blue-400">{{ fmtShort(form.tva_net_due) }}</p>
            </div>
            <div class="bg-orange-50 dark:bg-orange-900/20 rounded-xl p-4 text-center">
              <p class="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">TAP</p>
              <p class="text-xl font-bold text-orange-700 dark:text-orange-400">{{ fmtShort(tapNetDue) }}</p>
            </div>
            <div class="bg-violet-50 dark:bg-violet-900/20 rounded-xl p-4 text-center">
              <p class="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">IRG Total</p>
              <p class="text-xl font-bold text-violet-700 dark:text-violet-400">{{ fmtShort(form.irg_wages_amount + form.irg_fees_amount) }}</p>
            </div>
            <div class="bg-red-50 dark:bg-red-900/20 rounded-xl p-4 text-center">
              <p class="text-xs font-medium text-gray-500 uppercase tracking-wide mb-1">Stamp Tax</p>
              <p class="text-xl font-bold text-red-700 dark:text-red-400">{{ fmtShort(form.stamp_tax_amount) }}</p>
            </div>
          </div>
          <!-- Grand total row -->
          <div class="bg-gradient-to-r from-indigo-600 to-indigo-800 rounded-xl p-5 flex items-center justify-between text-white">
            <div>
              <p class="text-indigo-200 text-xs font-medium uppercase tracking-wide">Grand Total — Tax Due</p>
              <p class="text-xs text-indigo-300 mt-0.5">TVA + TAP + IRG + Stamp Tax</p>
            </div>
            <div class="text-right">
              <p class="text-3xl font-bold">{{ fmtShort(grandTotal) }}</p>
              <p class="text-indigo-200 text-sm">{{ fmtCurrency(grandTotal) }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- ── Past Declarations ── -->
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-800">
          <h3 class="font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <History class="w-4 h-4 text-gray-400" />
            Past Declarations
          </h3>
          <span class="text-xs text-gray-400">{{ declarations.length }} total</span>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
                <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Reference</th>
                <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Period</th>
                <th class="text-right px-4 py-3 text-xs font-semibold text-blue-600 uppercase tracking-wide">TVA</th>
                <th class="text-right px-4 py-3 text-xs font-semibold text-orange-600 uppercase tracking-wide">TAP</th>
                <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Total Due</th>
                <th class="text-right px-4 py-3 text-xs font-semibold text-emerald-600 uppercase tracking-wide">Paid</th>
                <th class="text-right px-4 py-3 text-xs font-semibold text-red-500 uppercase tracking-wide">Balance</th>
                <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Status</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
              <tr v-if="!declarations.length">
                <td colspan="8" class="text-center py-12 text-gray-400">No declarations recorded yet</td>
              </tr>
              <tr v-for="d in declarations" :key="d.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition-colors">
                <td class="px-4 py-3 font-mono font-semibold text-indigo-600 dark:text-indigo-400 text-xs">{{ d.reference }}</td>
                <td class="px-4 py-3 text-xs text-gray-600 dark:text-gray-300">
                  {{ monthName(String(d.period_month)) }} {{ d.period_year }}
                </td>
                <td class="px-4 py-3 text-right text-blue-600 dark:text-blue-400 font-medium text-xs">{{ fmtCurrency(d.tva_net_due) }}</td>
                <td class="px-4 py-3 text-right text-orange-600 dark:text-orange-400 font-medium text-xs">{{ fmtCurrency(d.tap_net_due) }}</td>
                <td class="px-4 py-3 text-right font-bold text-gray-900 dark:text-white text-xs">{{ fmtCurrency(d.total_tax_due) }}</td>
                <td class="px-4 py-3 text-right text-emerald-600 dark:text-emerald-400 font-medium text-xs">{{ fmtCurrency(d.total_paid) }}</td>
                <td class="px-4 py-3 text-right font-bold text-xs"
                  :class="(d.balance_due ?? 0) > 0 ? 'text-red-500' : 'text-emerald-600 dark:text-emerald-400'">
                  {{ fmtCurrency(d.balance_due) }}
                </td>
                <td class="px-4 py-3">
                  <span :class="statusBadge(d.status)" class="px-2 py-0.5 rounded-full text-xs font-semibold capitalize">
                    {{ d.status }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

    </template>

    <!-- ── Save/Submit Modal ── -->
    <Teleport to="body">
      <div v-if="showSaveModal" class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
        <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-md border border-gray-200 dark:border-gray-800">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800">
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">Save G50 Declaration</h2>
            <button @click="showSaveModal = false" class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg">
              <X class="w-5 h-5 text-gray-400" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div class="bg-indigo-50 dark:bg-indigo-900/20 rounded-xl p-4 text-sm">
              <div class="grid grid-cols-2 gap-2">
                <div class="flex justify-between col-span-2">
                  <span class="text-gray-500">Period</span>
                  <span class="font-semibold text-gray-900 dark:text-white">{{ monthName(selectedMonth) }} {{ selectedYear }}</span>
                </div>
                <div class="flex justify-between col-span-2 border-t border-indigo-100 dark:border-indigo-800 pt-2">
                  <span class="text-gray-500">TVA Net Due</span>
                  <span class="font-bold text-blue-600 dark:text-blue-400">{{ fmtCurrency(form.tva_net_due) }}</span>
                </div>
                <div class="flex justify-between col-span-2">
                  <span class="text-gray-500">TAP Net Due</span>
                  <span class="font-bold text-orange-600 dark:text-orange-400">{{ fmtCurrency(tapNetDue) }}</span>
                </div>
                <div class="flex justify-between col-span-2 border-t border-indigo-100 dark:border-indigo-800 pt-2">
                  <span class="font-bold text-indigo-700 dark:text-indigo-300">Grand Total</span>
                  <span class="font-bold text-indigo-700 dark:text-indigo-300 text-lg">{{ fmtCurrency(grandTotal) }}</span>
                </div>
              </div>
            </div>
            <div>
              <label class="lbl">Submission Reference (optional)</label>
              <input v-model="submissionRef" type="text" class="field" placeholder="DGI reference number..." />
            </div>
          </div>
          <div class="flex justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50 rounded-b-2xl">
            <button @click="showSaveModal = false"
              class="px-4 py-2 text-sm text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg">
              Cancel
            </button>
            <button @click="saveDeclaration" :disabled="saving"
              class="px-5 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-sm font-semibold rounded-lg flex items-center gap-2">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              {{ saving ? 'Saving...' : 'Save & Submit' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  ScrollText, Calendar, RefreshCw, Save, Loader2, X,
  Percent, TrendingUp, Users, Calculator, History,
  CheckCircle, Clock
} from 'lucide-vue-next'
// Stamp icon workaround — use SquareCheck
import { Stamp } from 'lucide-vue-next'
import { taxAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

// ── State ──────────────────────────────────────────────
const loading      = ref(true)
const saving       = ref(false)
const showSaveModal = ref(false)
const submissionRef = ref('')
const computed_data = ref<any>(null)
const declarations  = ref<any[]>([])
const existingDecl  = ref<any>(null)

const now = new Date()
const currentYear  = now.getFullYear()
const currentMonth = now.getMonth() + 1

const selectedYear  = ref(String(currentYear))
const selectedMonth = ref(String(currentMonth))

const years  = Array.from({ length: 6 }, (_, i) => String(currentYear - i))
const months = [
  { value: '1',  label: 'January' },  { value: '2',  label: 'February' },
  { value: '3',  label: 'March' },    { value: '4',  label: 'April' },
  { value: '5',  label: 'May' },      { value: '6',  label: 'June' },
  { value: '7',  label: 'July' },     { value: '8',  label: 'August' },
  { value: '9',  label: 'September'}, { value: '10', label: 'October' },
  { value: '11', label: 'November' }, { value: '12', label: 'December' },
]

const form = ref({
  tva_collected: 0, tva_deductible: 0, tva_credit_bf: 0,
  tva_net_due: 0, tva_credit_carry: 0,
  tap_base: 0, tap_rate: 0.02, tap_reduction: 0,
  irg_wages_amount: 0, irg_fees_amount: 0,
  stamp_tax_amount: 0, notes: '',
})

const computed = ref<any>(null)

const tapNetDue = computed(() => {
  const raw = form.value.tap_base * form.value.tap_rate - form.value.tap_reduction
  return Math.max(0, raw)
})

const grandTotal = computed(() => {
  return (
    Math.max(0, form.value.tva_net_due) +
    tapNetDue.value +
    form.value.irg_wages_amount +
    form.value.irg_fees_amount +
    form.value.stamp_tax_amount
  )
})

// ── Helpers ────────────────────────────────────────────
function monthName(m: string | number) {
  const idx = parseInt(String(m)) - 1
  return months[idx]?.label ?? String(m)
}
function fmtCurrency(n?: number | null) {
  if (n == null) return '—'
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n) + ' DZD'
}
function fmtShort(n?: number | null) {
  if (n == null) return '—'
  if (Math.abs(n) >= 1e9) return (n / 1e9).toFixed(2) + 'B'
  if (Math.abs(n) >= 1e6) return (n / 1e6).toFixed(2) + 'M'
  if (Math.abs(n) >= 1e3) return (n / 1e3).toFixed(1) + 'K'
  return n.toFixed(0)
}
function statusBadge(s?: string) {
  switch (s) {
    case 'submitted': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    case 'accepted':  return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'rejected':  return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    case 'amended':   return 'bg-violet-100 text-violet-700 dark:bg-violet-900/30 dark:text-violet-400'
    default:          return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
  }
}
function statusColor(s?: string) {
  switch (s) {
    case 'submitted': return 'text-blue-600 dark:text-blue-400'
    case 'accepted':  return 'text-emerald-600 dark:text-emerald-400'
    case 'rejected':  return 'text-red-500'
    default:          return 'text-amber-600 dark:text-amber-400'
  }
}

// ── Actions ────────────────────────────────────────────
async function loadG50() {
  loading.value = true
  try {
    const [g50Res, declRes] = await Promise.allSettled([
      taxAPI.getG50(selectedYear.value, selectedMonth.value),
      taxAPI.listDeclarations({ year: selectedYear.value, type: 'g50' }),
    ])

    if (g50Res.status === 'fulfilled') {
      const d = g50Res.value.data
      computed.value = d
      form.value = {
        tva_collected: d.tva_collected ?? 0,
        tva_deductible: d.tva_deductible ?? 0,
        tva_credit_bf: d.tva_credit_bf ?? 0,
        tva_net_due: d.tva_net_due ?? 0,
        tva_credit_carry: d.tva_credit_carry ?? 0,
        tap_base: d.tap_base ?? 0,
        tap_rate: d.tap_rate ?? 0.02,
        tap_reduction: 0,
        irg_wages_amount: 0,
        irg_fees_amount: 0,
        stamp_tax_amount: d.stamp_tax_amount ?? 0,
        notes: '',
      }
    }

    if (declRes.status === 'fulfilled') {
      declarations.value = declRes.value.data || []
      existingDecl.value = declarations.value.find(
        d => String(d.period_month) === selectedMonth.value
      ) ?? null
    }
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load G50 data', 'error')
  } finally {
    loading.value = false
  }
}

function openSave() {
  submissionRef.value = ''
  showSaveModal.value = true
}

async function saveDeclaration() {
  saving.value = true
  try {
    await taxAPI.submitG50({
      period_year: parseInt(selectedYear.value),
      period_month: parseInt(selectedMonth.value),
      tva_collected: form.value.tva_collected,
      tva_deductible: form.value.tva_deductible,
      tva_credit_bf: form.value.tva_credit_bf,
      tap_base: form.value.tap_base,
      tap_rate: form.value.tap_rate,
      stamp_tax_amount: form.value.stamp_tax_amount,
      irg_wages_amount: form.value.irg_wages_amount,
      irg_fees_amount: form.value.irg_fees_amount,
      notes: form.value.notes,
    })
    store.addToast('G50 declaration saved and submitted', 'success')
    showSaveModal.value = false
    await loadG50()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Error saving declaration', 'error')
  } finally {
    saving.value = false
  }
}

onMounted(loadG50)
</script>

<style scoped>
.field {
  @apply w-full px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-lg
         bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-sm
         focus:ring-2 focus:ring-indigo-500 outline-none transition-shadow;
}
.lbl {
  @apply block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1;
}
</style>
