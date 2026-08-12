<template>
  <div class="min-h-screen bg-gray-950 text-gray-100">
    <!-- Header -->
    <div class="bg-gray-900 border-b border-gray-800 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center shadow-lg">
            <Banknote :size="20" class="text-white" />
          </div>
          <div>
            <h1 class="text-xl font-bold text-white">Paiements Fiscaux</h1>
            <p class="text-xs text-gray-400 mt-0.5">Suivi des versements DGI — TVA, TAP, IBS, IRG</p>
          </div>
        </div>
        <button @click="openCreateModal" class="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-sm font-medium transition-colors shadow-lg shadow-emerald-900/30">
          <Plus :size="15" />
          Nouveau Paiement
        </button>
      </div>
    </div>

    <div class="p-6 space-y-6">
      <!-- Filter Bar -->
      <div class="bg-gray-900 rounded-xl border border-gray-800 p-4">
        <div class="flex flex-wrap items-center gap-4">
          <div class="flex items-center gap-2">
            <Calendar :size="16" class="text-gray-400" />
            <select v-model="filterYear" @change="loadPayments" class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500">
              <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}</option>
            </select>
          </div>
          <div class="flex items-center gap-2">
            <Filter :size="16" class="text-gray-400" />
            <div class="flex gap-1">
              <button v-for="s in statusFilters" :key="s.value"
                @click="filterStatus = s.value"
                :class="['px-3 py-1.5 rounded-lg text-xs font-medium transition-colors', filterStatus === s.value ? s.active : 'bg-gray-800 text-gray-400 hover:text-gray-200']">
                {{ s.label }}
              </button>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-gray-500">Type:</span>
            <select v-model="filterType" class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500">
              <option value="">Tous les impôts</option>
              <option value="TVA">TVA</option>
              <option value="TAP">TAP</option>
              <option value="IBS">IBS</option>
              <option value="IRG">IRG</option>
              <option value="TIMBRE">Timbre Fiscal</option>
            </select>
          </div>
          <div class="relative ml-auto">
            <Search :size="15" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
            <input v-model="searchQuery" type="text" placeholder="Rechercher..." class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg pl-9 pr-4 py-2 text-sm w-56 focus:outline-none focus:ring-2 focus:ring-emerald-500 placeholder-gray-500" />
          </div>
        </div>
      </div>

      <!-- KPI Cards -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-gray-500 font-medium uppercase tracking-wide">Total Dû</span>
            <div class="w-7 h-7 rounded-lg bg-blue-500/15 flex items-center justify-center">
              <Receipt :size="14" class="text-blue-400" />
            </div>
          </div>
          <p class="text-xl font-bold text-white">{{ fmtCur(kpi.totalDue) }}</p>
          <p class="text-xs text-gray-500 mt-1">{{ kpi.totalCount }} paiements</p>
        </div>

        <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-gray-500 font-medium uppercase tracking-wide">Payé</span>
            <div class="w-7 h-7 rounded-lg bg-emerald-500/15 flex items-center justify-center">
              <CheckCircle2 :size="14" class="text-emerald-400" />
            </div>
          </div>
          <p class="text-xl font-bold text-emerald-400">{{ fmtCur(kpi.totalPaid) }}</p>
          <p class="text-xs text-gray-500 mt-1">{{ kpi.paidCount }} versements</p>
        </div>

        <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-gray-500 font-medium uppercase tracking-wide">En Attente</span>
            <div class="w-7 h-7 rounded-lg bg-yellow-500/15 flex items-center justify-center">
              <Clock :size="14" class="text-yellow-400" />
            </div>
          </div>
          <p class="text-xl font-bold text-yellow-400">{{ fmtCur(kpi.totalPending) }}</p>
          <p class="text-xs text-gray-500 mt-1">{{ kpi.pendingCount }} en cours</p>
        </div>

        <div class="bg-gradient-to-br from-red-900/40 to-rose-900/20 border border-red-700/30 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-red-300 font-medium uppercase tracking-wide">En Retard</span>
            <div class="w-7 h-7 rounded-lg bg-red-500/25 flex items-center justify-center">
              <AlertTriangle :size="14" class="text-red-400" />
            </div>
          </div>
          <p class="text-xl font-bold text-red-400">{{ fmtCur(kpi.totalOverdue) }}</p>
          <p class="text-xs text-red-400/70 mt-1">{{ kpi.overdueCount }} dépassés</p>
        </div>
      </div>

      <!-- Aging Buckets -->
      <div v-if="kpi.overdueCount > 0" class="grid grid-cols-4 gap-3">
        <div v-for="bucket in agingBuckets" :key="bucket.label" class="bg-gray-900 border border-gray-800 rounded-xl p-3 text-center">
          <p class="text-xs text-gray-500 mb-1">{{ bucket.label }}</p>
          <p class="text-base font-bold" :class="bucket.color">{{ fmtCur(bucket.amount) }}</p>
          <p class="text-xs text-gray-600 mt-0.5">{{ bucket.count }} item{{ bucket.count !== 1 ? 's' : '' }}</p>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-16">
        <div class="flex flex-col items-center gap-3">
          <div class="w-10 h-10 border-2 border-emerald-500 border-t-transparent rounded-full animate-spin"></div>
          <span class="text-sm text-gray-400">Chargement paiements fiscaux...</span>
        </div>
      </div>

      <!-- Table -->
      <div v-else class="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
        <div class="px-5 py-4 border-b border-gray-800 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <Banknote :size="16" class="text-emerald-400" />
            <span class="text-sm font-semibold text-gray-200">Paiements Fiscaux</span>
            <span class="px-2 py-0.5 bg-gray-800 rounded-full text-xs text-gray-400">{{ filteredPayments.length }} entrées</span>
          </div>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-800/60">
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Référence</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Type Impôt</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Période</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Échéance</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-400 uppercase tracking-wide">Montant Dû</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-400 uppercase tracking-wide">Payé</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-400 uppercase tracking-wide">Solde</th>
                <th class="px-4 py-3 text-center text-xs font-semibold text-gray-400 uppercase tracking-wide">Statut</th>
                <th class="px-4 py-3 text-center text-xs font-semibold text-gray-400 uppercase tracking-wide">Actions</th>
              </tr>
            </thead>
            <tbody>
              <template v-if="filteredPayments.length > 0">
                <tr v-for="p in paginatedPayments" :key="p.id"
                  class="border-t border-gray-800 hover:bg-gray-800/40 transition-colors"
                  :class="isOverdue(p) ? 'bg-red-950/10' : ''">
                  <td class="px-4 py-3">
                    <span class="font-mono text-emerald-400 text-xs font-semibold">{{ p.reference }}</span>
                  </td>
                  <td class="px-4 py-3">
                    <span :class="['px-2 py-0.5 rounded-full text-xs font-bold', taxTypeBadge(p.tax_type)]">
                      {{ p.tax_type }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-gray-300">
                    <div>
                      <p>{{ fmtDate(p.period_start) }}</p>
                      <p v-if="p.period_end" class="text-xs text-gray-500">au {{ fmtDate(p.period_end) }}</p>
                    </div>
                  </td>
                  <td class="px-4 py-3">
                    <div>
                      <p class="text-gray-300">{{ fmtDate(p.due_date) }}</p>
                      <p v-if="isOverdue(p)" class="text-xs text-red-400 font-semibold">
                        {{ daysOverdue(p) }}j de retard
                      </p>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-right font-bold text-white">{{ fmtCur(p.amount_due) }}</td>
                  <td class="px-4 py-3 text-right text-emerald-400 font-semibold">{{ fmtCur(p.amount_paid || 0) }}</td>
                  <td class="px-4 py-3 text-right font-bold" :class="(p.balance || 0) > 0 ? 'text-red-400' : 'text-gray-500'">
                    {{ fmtCur(p.balance || 0) }}
                  </td>
                  <td class="px-4 py-3 text-center">
                    <div class="flex flex-col items-center gap-1">
                      <span :class="['px-2 py-0.5 rounded-full text-xs font-medium', paymentStatusBadge(p.status)]">
                        {{ paymentStatusLabel(p.status) }}
                      </span>
                      <p v-if="isOverdue(p)" class="text-xs text-red-400 font-bold">RETARD</p>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-center">
                    <div class="flex items-center justify-center gap-1">
                      <button @click="openPayModal(p)" title="Enregistrer paiement" class="p-1.5 rounded-lg hover:bg-gray-700 text-gray-400 hover:text-emerald-400 transition-colors">
                        <CreditCard :size="14" />
                      </button>
                      <button @click="openEditModal(p)" class="p-1.5 rounded-lg hover:bg-gray-700 text-gray-400 hover:text-blue-400 transition-colors">
                        <Edit3 :size="14" />
                      </button>
                      <button @click="deletePayment(p.id)" class="p-1.5 rounded-lg hover:bg-gray-700 text-gray-400 hover:text-red-400 transition-colors">
                        <Trash2 :size="14" />
                      </button>
                    </div>
                  </td>
                </tr>
              </template>
              <tr v-else>
                <td colspan="9" class="px-4 py-16 text-center">
                  <div class="flex flex-col items-center gap-3">
                    <div class="w-12 h-12 rounded-full bg-gray-800 flex items-center justify-center">
                      <Banknote :size="22" class="text-gray-600" />
                    </div>
                    <p class="text-gray-500 text-sm">Aucun paiement fiscal enregistré</p>
                    <button @click="openCreateModal" class="text-emerald-400 hover:text-emerald-300 text-xs underline">
                      Ajouter un paiement
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
            <tfoot v-if="filteredPayments.length > 0">
              <tr class="bg-gray-800/80 border-t-2 border-gray-700">
                <td colspan="4" class="px-4 py-3 text-sm font-bold text-gray-300">TOTAL</td>
                <td class="px-4 py-3 text-right text-sm font-bold text-white">{{ fmtCur(totals.due) }}</td>
                <td class="px-4 py-3 text-right text-sm font-bold text-emerald-400">{{ fmtCur(totals.paid) }}</td>
                <td class="px-4 py-3 text-right text-sm font-bold text-red-400">{{ fmtCur(totals.balance) }}</td>
                <td colspan="2"></td>
              </tr>
            </tfoot>
          </table>
        </div>

        <div v-if="totalPages > 1" class="px-5 py-3 border-t border-gray-800 flex items-center justify-between">
          <span class="text-xs text-gray-500">Page {{ currentPage }} / {{ totalPages }}</span>
          <div class="flex gap-2">
            <button @click="currentPage--" :disabled="currentPage === 1" class="px-3 py-1.5 rounded-lg bg-gray-800 border border-gray-700 text-sm text-gray-300 disabled:opacity-40 hover:bg-gray-700 transition-colors">
              <ChevronLeft :size="14" />
            </button>
            <button @click="currentPage++" :disabled="currentPage === totalPages" class="px-3 py-1.5 rounded-lg bg-gray-800 border border-gray-700 text-sm text-gray-300 disabled:opacity-40 hover:bg-gray-700 transition-colors">
              <ChevronRight :size="14" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
        <div class="bg-gray-900 border border-gray-700 rounded-2xl shadow-2xl w-full max-w-xl max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between p-6 border-b border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-emerald-600 flex items-center justify-center">
                <Banknote :size="18" class="text-white" />
              </div>
              <h2 class="text-lg font-bold text-white">{{ editMode ? 'Modifier Paiement' : 'Nouveau Paiement Fiscal' }}</h2>
            </div>
            <button @click="closeModal" class="p-2 rounded-lg hover:bg-gray-800 text-gray-400 transition-colors">
              <X :size="18" />
            </button>
          </div>

          <form @submit.prevent="submitModal" class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="lbl">Type d'impôt *</label>
                <select v-model="form.tax_type" required class="field">
                  <option value="TVA">TVA</option>
                  <option value="TAP">TAP</option>
                  <option value="IBS">IBS</option>
                  <option value="IRG">IRG</option>
                  <option value="IRG_SALAIRES">IRG Salaires</option>
                  <option value="TIMBRE">Timbre Fiscal</option>
                  <option value="AUTRE">Autre</option>
                </select>
              </div>
              <div>
                <label class="lbl">Méthode de Paiement</label>
                <select v-model="form.payment_method" class="field">
                  <option value="virement">Virement Bancaire</option>
                  <option value="cheque">Chèque</option>
                  <option value="especes">Espèces</option>
                  <option value="telepaiement">Télépaiement</option>
                </select>
              </div>
              <div>
                <label class="lbl">Période Début *</label>
                <input v-model="form.period_start" type="date" required class="field" />
              </div>
              <div>
                <label class="lbl">Période Fin</label>
                <input v-model="form.period_end" type="date" class="field" />
              </div>
              <div>
                <label class="lbl">Date d'Échéance *</label>
                <input v-model="form.due_date" type="date" required class="field" />
              </div>
              <div>
                <label class="lbl">Date de Paiement</label>
                <input v-model="form.payment_date" type="date" class="field" />
              </div>
              <div>
                <label class="lbl">Montant Dû *</label>
                <input v-model.number="form.amount_due" type="number" min="0" step="0.01" required class="field" />
              </div>
              <div>
                <label class="lbl">Montant Payé</label>
                <input v-model.number="form.amount_paid" type="number" min="0" step="0.01" class="field" />
              </div>
              <div>
                <label class="lbl">Référence de Paiement</label>
                <input v-model="form.payment_reference" type="text" placeholder="N° avis de débit / reçu" class="field" />
              </div>
              <div>
                <label class="lbl">Bureau des Impôts</label>
                <input v-model="form.tax_office" type="text" placeholder="CDI / CPI / DGE..." class="field" />
              </div>
              <div class="col-span-2">
                <label class="lbl">Notes</label>
                <textarea v-model="form.notes" rows="2" class="field resize-none" placeholder="Pénalités, majorations, remarques..."></textarea>
              </div>
            </div>

            <!-- Balance Preview -->
            <div class="bg-gray-800/60 border border-gray-700 rounded-xl p-4">
              <div class="flex justify-between text-sm mb-2"><span class="text-gray-400">Montant dû</span><span class="text-white font-bold">{{ fmtCur(form.amount_due || 0) }}</span></div>
              <div class="flex justify-between text-sm mb-2"><span class="text-gray-400">Montant payé</span><span class="text-emerald-400 font-bold">{{ fmtCur(form.amount_paid || 0) }}</span></div>
              <div class="flex justify-between text-base font-bold border-t border-gray-700 pt-2">
                <span class="text-gray-300">Solde restant</span>
                <span :class="((form.amount_due || 0) - (form.amount_paid || 0)) > 0 ? 'text-red-400' : 'text-emerald-400'">
                  {{ fmtCur(Math.max(0, (form.amount_due || 0) - (form.amount_paid || 0))) }}
                </span>
              </div>
            </div>

            <div class="flex gap-3 pt-2">
              <button type="button" @click="closeModal" class="flex-1 px-4 py-2.5 bg-gray-800 border border-gray-700 text-gray-200 rounded-xl text-sm font-medium hover:bg-gray-700 transition-colors">
                Annuler
              </button>
              <button type="submit" :disabled="submitting" class="flex-1 px-4 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-xl text-sm font-bold transition-colors disabled:opacity-60 flex items-center justify-center gap-2">
                <div v-if="submitting" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                <Save v-else :size="15" />
                {{ editMode ? 'Enregistrer' : 'Créer' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Pay Modal (quick payment) -->
    <Teleport to="body">
      <div v-if="showPayModal && payTarget" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
        <div class="bg-gray-900 border border-gray-700 rounded-2xl shadow-2xl w-full max-w-md">
          <div class="flex items-center justify-between p-6 border-b border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-emerald-700 flex items-center justify-center">
                <CreditCard :size="18" class="text-white" />
              </div>
              <div>
                <h2 class="text-lg font-bold text-white">Enregistrer Paiement</h2>
                <p class="text-xs text-gray-500">{{ payTarget.tax_type }} — {{ fmtDate(payTarget.period_start) }}</p>
              </div>
            </div>
            <button @click="showPayModal = false" class="p-2 rounded-lg hover:bg-gray-800 text-gray-400 transition-colors"><X :size="18" /></button>
          </div>
          <div class="p-6 space-y-4">
            <div class="bg-gray-800 rounded-xl p-4 space-y-2 text-sm">
              <div class="flex justify-between"><span class="text-gray-400">Montant Dû</span><span class="text-white font-bold">{{ fmtCur(payTarget.amount_due) }}</span></div>
              <div class="flex justify-between"><span class="text-gray-400">Déjà Payé</span><span class="text-emerald-400">{{ fmtCur(payTarget.amount_paid || 0) }}</span></div>
              <div class="flex justify-between border-t border-gray-700 pt-2"><span class="text-gray-400">Reste à Payer</span><span class="text-red-400 font-bold">{{ fmtCur(payTarget.balance || 0) }}</span></div>
            </div>
            <div>
              <label class="lbl">Montant Versé *</label>
              <input v-model.number="payAmount" type="number" min="0" :max="payTarget.balance" step="0.01" class="field" />
            </div>
            <div>
              <label class="lbl">Date de Paiement *</label>
              <input v-model="payDate" type="date" class="field" />
            </div>
            <div>
              <label class="lbl">Référence de Paiement</label>
              <input v-model="payRef" type="text" placeholder="N° avis de débit / reçu DGI" class="field" />
            </div>
            <div class="flex gap-3">
              <button @click="showPayModal = false" class="flex-1 px-4 py-2.5 bg-gray-800 border border-gray-700 text-gray-200 rounded-xl text-sm font-medium hover:bg-gray-700 transition-colors">
                Annuler
              </button>
              <button @click="recordPayment" :disabled="submitting" class="flex-1 px-4 py-2.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-xl text-sm font-bold transition-colors disabled:opacity-60 flex items-center justify-center gap-2">
                <div v-if="submitting" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                <CreditCard v-else :size="15" />
                Enregistrer
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  Banknote, Plus, Calendar, Filter, Search, Receipt, CheckCircle2,
  Clock, AlertTriangle, CreditCard, Edit3, Trash2, ChevronLeft,
  ChevronRight, X, Save
} from 'lucide-vue-next'
import { taxAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()
const loading = ref(false)
const submitting = ref(false)

const now = new Date()
const filterYear = ref(now.getFullYear())
const filterStatus = ref('all')
const filterType = ref('')
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = 20

const payments = ref<any[]>([])
const showModal = ref(false)
const editMode = ref(false)
const showPayModal = ref(false)
const payTarget = ref<any>(null)
const payAmount = ref(0)
const payDate = ref(now.toISOString().split('T')[0])
const payRef = ref('')

const statusFilters = [
  { value: 'all', label: 'Tous', active: 'bg-gray-600 text-white' },
  { value: 'pending', label: 'En Attente', active: 'bg-yellow-700 text-yellow-200' },
  { value: 'paid', label: 'Payé', active: 'bg-emerald-700 text-emerald-200' },
  { value: 'partial', label: 'Partiel', active: 'bg-blue-700 text-blue-200' },
  { value: 'overdue', label: 'Retard', active: 'bg-red-700 text-red-200' }
]

const yearOptions = computed(() => {
  const y = now.getFullYear()
  return [y + 1, y, y - 1, y - 2, y - 3]
})

const defaultForm = () => ({
  id: null as number | null,
  tax_type: 'TVA',
  payment_method: 'virement',
  period_start: '',
  period_end: '',
  due_date: '',
  payment_date: '',
  amount_due: 0,
  amount_paid: 0,
  payment_reference: '',
  tax_office: '',
  notes: ''
})
const form = ref(defaultForm())

async function loadPayments() {
  loading.value = true
  currentPage.value = 1
  try {
    const params: Record<string, any> = { year: filterYear.value }
    if (filterStatus.value !== 'all') params.status = filterStatus.value
    if (filterType.value) params.tax_type = filterType.value
    const res = await taxAPI.listTaxPayments(params)
    payments.value = res.data?.payments || res.data || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur chargement paiements', 'error')
    payments.value = []
  } finally {
    loading.value = false
  }
}

const filteredPayments = computed(() => {
  let list = payments.value
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(p =>
      (p.reference || '').toLowerCase().includes(q) ||
      (p.tax_type || '').toLowerCase().includes(q) ||
      (p.payment_reference || '').toLowerCase().includes(q)
    )
  }
  return list
})

const totalPages = computed(() => Math.ceil(filteredPayments.value.length / pageSize))
const paginatedPayments = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredPayments.value.slice(start, start + pageSize)
})

const kpi = computed(() => {
  const all = payments.value
  const overdue = all.filter(p => isOverdue(p))
  const paid = all.filter(p => p.status === 'paid')
  const pending = all.filter(p => p.status === 'pending' || p.status === 'partial')
  return {
    totalDue: all.reduce((s, p) => s + (p.amount_due || 0), 0),
    totalPaid: all.reduce((s, p) => s + (p.amount_paid || 0), 0),
    totalPending: pending.reduce((s, p) => s + ((p.amount_due || 0) - (p.amount_paid || 0)), 0),
    totalOverdue: overdue.reduce((s, p) => s + (p.balance || 0), 0),
    totalCount: all.length,
    paidCount: paid.length,
    pendingCount: pending.length,
    overdueCount: overdue.length
  }
})

const agingBuckets = computed(() => {
  const overdue = payments.value.filter(p => isOverdue(p))
  const buckets = [
    { label: '1-30 jours', days: [1, 30], color: 'text-yellow-400', amount: 0, count: 0 },
    { label: '31-60 jours', days: [31, 60], color: 'text-orange-400', amount: 0, count: 0 },
    { label: '61-90 jours', days: [61, 90], color: 'text-red-400', amount: 0, count: 0 },
    { label: '> 90 jours', days: [91, Infinity], color: 'text-red-600', amount: 0, count: 0 }
  ]
  overdue.forEach(p => {
    const d = daysOverdue(p)
    const bucket = buckets.find(b => d >= b.days[0] && d <= b.days[1])
    if (bucket) { bucket.amount += (p.balance || 0); bucket.count++ }
  })
  return buckets
})

const totals = computed(() => ({
  due: filteredPayments.value.reduce((s, p) => s + (p.amount_due || 0), 0),
  paid: filteredPayments.value.reduce((s, p) => s + (p.amount_paid || 0), 0),
  balance: filteredPayments.value.reduce((s, p) => s + (p.balance || 0), 0)
}))

watch([filterStatus, filterType], () => { currentPage.value = 1; loadPayments() })
watch(searchQuery, () => { currentPage.value = 1 })

function isOverdue(p: any) {
  if (!p.due_date || p.status === 'paid') return false
  return new Date(p.due_date) < new Date() && (p.balance || 0) > 0
}
function daysOverdue(p: any) {
  if (!p.due_date) return 0
  const diff = new Date().getTime() - new Date(p.due_date).getTime()
  return Math.max(0, Math.floor(diff / 86400000))
}

function openCreateModal() {
  form.value = defaultForm()
  form.value.period_start = `${filterYear.value}-01-01`
  form.value.due_date = `${filterYear.value}-02-20`
  editMode.value = false
  showModal.value = true
}
function openEditModal(p: any) {
  form.value = {
    id: p.id,
    tax_type: p.tax_type || 'TVA',
    payment_method: p.payment_method || 'virement',
    period_start: p.period_start?.split('T')[0] || '',
    period_end: p.period_end?.split('T')[0] || '',
    due_date: p.due_date?.split('T')[0] || '',
    payment_date: p.payment_date?.split('T')[0] || '',
    amount_due: p.amount_due || 0,
    amount_paid: p.amount_paid || 0,
    payment_reference: p.payment_reference || '',
    tax_office: p.tax_office || '',
    notes: p.notes || ''
  }
  editMode.value = true
  showModal.value = true
}
function closeModal() { showModal.value = false; form.value = defaultForm() }

async function submitModal() {
  submitting.value = true
  try {
    if (editMode.value && form.value.id) {
      await taxAPI.updateTaxPayment(form.value.id, form.value)
      store.addToast('Paiement mis à jour', 'success')
    } else {
      await taxAPI.createTaxPayment(form.value)
      store.addToast('Paiement créé', 'success')
    }
    closeModal()
    await loadPayments()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur sauvegarde', 'error')
  } finally {
    submitting.value = false
  }
}

function openPayModal(p: any) {
  payTarget.value = p
  payAmount.value = p.balance || 0
  payDate.value = now.toISOString().split('T')[0]
  payRef.value = ''
  showPayModal.value = true
}
async function recordPayment() {
  if (!payTarget.value || !payAmount.value) return
  submitting.value = true
  try {
    await taxAPI.updateTaxPayment(payTarget.value.id, {
      amount_paid: (payTarget.value.amount_paid || 0) + payAmount.value,
      payment_date: payDate.value,
      payment_reference: payRef.value || undefined
    })
    store.addToast('Paiement enregistré', 'success')
    showPayModal.value = false
    await loadPayments()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur', 'error')
  } finally {
    submitting.value = false
  }
}
async function deletePayment(id: number) {
  if (!confirm('Supprimer ce paiement fiscal ?')) return
  try {
    await taxAPI.deleteTaxPayment(id)
    store.addToast('Paiement supprimé', 'success')
    await loadPayments()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur suppression', 'error')
  }
}

function fmtCur(v: number) {
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0) + ' DZD'
}
function fmtDate(d: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('fr-DZ', { day: '2-digit', month: '2-digit', year: 'numeric' })
}
function taxTypeBadge(t: string) {
  const m: Record<string, string> = {
    TVA: 'bg-blue-900/60 text-blue-300 border border-blue-700/40',
    TAP: 'bg-purple-900/60 text-purple-300 border border-purple-700/40',
    IBS: 'bg-amber-900/60 text-amber-300 border border-amber-700/40',
    IRG: 'bg-teal-900/60 text-teal-300 border border-teal-700/40',
    TIMBRE: 'bg-gray-800 text-gray-400 border border-gray-700'
  }
  return m[t] || 'bg-gray-800 text-gray-400 border border-gray-700'
}
function paymentStatusBadge(s: string) {
  if (s === 'paid') return 'bg-emerald-900/50 text-emerald-300'
  if (s === 'partial') return 'bg-blue-900/50 text-blue-300'
  if (s === 'overdue') return 'bg-red-900/50 text-red-300'
  return 'bg-yellow-900/50 text-yellow-300'
}
function paymentStatusLabel(s: string) {
  const m: Record<string, string> = { paid: 'Payé', partial: 'Partiel', pending: 'En Attente', overdue: 'Retard' }
  return m[s] || s
}

onMounted(loadPayments)
</script>

<style scoped>
.field {
  @apply w-full bg-gray-800 border border-gray-700 text-gray-200 rounded-xl px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500 transition-all placeholder-gray-600;
}
.lbl {
  @apply block text-xs font-semibold text-gray-400 mb-1.5 uppercase tracking-wide;
}
</style>
