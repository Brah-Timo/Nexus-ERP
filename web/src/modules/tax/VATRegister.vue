<template>
  <div class="min-h-screen bg-gray-950 text-gray-100">
    <!-- Header -->
    <div class="bg-gray-900 border-b border-gray-800 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-600 flex items-center justify-center shadow-lg">
            <BookOpen :size="20" class="text-white" />
          </div>
          <div>
            <h1 class="text-xl font-bold text-white">VAT Register</h1>
            <p class="text-xs text-gray-400 mt-0.5">Registre TVA — Achats &amp; Ventes</p>
          </div>
        </div>
        <div class="flex items-center gap-3">
          <button @click="exportCSV" class="flex items-center gap-2 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-200 rounded-lg text-sm font-medium transition-colors">
            <Download :size="15" />
            Export CSV
          </button>
          <button @click="openCreateModal" class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-sm font-medium transition-colors shadow-lg shadow-blue-900/30">
            <Plus :size="15" />
            Nouvelle Entrée
          </button>
        </div>
      </div>
    </div>

    <div class="p-6 space-y-6">
      <!-- Period & Filter Controls -->
      <div class="bg-gray-900 rounded-xl border border-gray-800 p-4">
        <div class="flex flex-wrap items-center gap-4">
          <div class="flex items-center gap-2">
            <Calendar :size="16" class="text-gray-400" />
            <span class="text-sm text-gray-400 font-medium">Période :</span>
          </div>
          <div class="flex items-center gap-2">
            <select v-model="filterYear" @change="loadData" class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
              <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}</option>
            </select>
            <select v-model="filterMonth" @change="loadData" class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500">
              <option value="">Tous les mois</option>
              <option v-for="m in months" :key="m.value" :value="m.value">{{ m.label }}</option>
            </select>
          </div>

          <!-- Tab Toggle -->
          <div class="ml-auto flex bg-gray-800 rounded-lg p-1 gap-1">
            <button
              @click="activeTab = 'sales'"
              :class="['px-4 py-1.5 rounded-md text-sm font-medium transition-all', activeTab === 'sales' ? 'bg-blue-600 text-white shadow' : 'text-gray-400 hover:text-gray-200']"
            >
              <div class="flex items-center gap-2"><TrendingUp :size="14" /> Ventes</div>
            </button>
            <button
              @click="activeTab = 'purchases'"
              :class="['px-4 py-1.5 rounded-md text-sm font-medium transition-all', activeTab === 'purchases' ? 'bg-cyan-600 text-white shadow' : 'text-gray-400 hover:text-gray-200']"
            >
              <div class="flex items-center gap-2"><ShoppingCart :size="14" /> Achats</div>
            </button>
          </div>

          <!-- Search -->
          <div class="relative">
            <Search :size="15" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Rechercher partenaire, N° doc..."
              class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg pl-9 pr-4 py-2 text-sm w-64 focus:outline-none focus:ring-2 focus:ring-blue-500 placeholder-gray-500"
            />
          </div>
        </div>
      </div>

      <!-- KPI Row -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-gray-500 font-medium uppercase tracking-wide">Base HT Ventes</span>
            <div class="w-7 h-7 rounded-lg bg-blue-500/15 flex items-center justify-center">
              <TrendingUp :size="14" class="text-blue-400" />
            </div>
          </div>
          <p class="text-xl font-bold text-white">{{ fmtCur(kpi.totalSalesBase) }}</p>
          <p class="text-xs text-gray-500 mt-1">{{ kpi.salesCount }} factures</p>
        </div>

        <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-gray-500 font-medium uppercase tracking-wide">TVA Collectée</span>
            <div class="w-7 h-7 rounded-lg bg-emerald-500/15 flex items-center justify-center">
              <ArrowUpCircle :size="14" class="text-emerald-400" />
            </div>
          </div>
          <p class="text-xl font-bold text-emerald-400">{{ fmtCur(kpi.totalSalesVAT) }}</p>
          <p class="text-xs text-gray-500 mt-1">TVA sur ventes</p>
        </div>

        <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-gray-500 font-medium uppercase tracking-wide">TVA Déductible</span>
            <div class="w-7 h-7 rounded-lg bg-amber-500/15 flex items-center justify-center">
              <ArrowDownCircle :size="14" class="text-amber-400" />
            </div>
          </div>
          <p class="text-xl font-bold text-amber-400">{{ fmtCur(kpi.totalPurchaseVAT) }}</p>
          <p class="text-xs text-gray-500 mt-1">{{ kpi.purchaseCount }} achats</p>
        </div>

        <div class="bg-gradient-to-br from-blue-900/50 to-cyan-900/30 border border-blue-700/40 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-blue-300 font-medium uppercase tracking-wide">TVA Nette Due</span>
            <div class="w-7 h-7 rounded-lg bg-blue-500/25 flex items-center justify-center">
              <Calculator :size="14" class="text-blue-400" />
            </div>
          </div>
          <p :class="['text-xl font-bold', kpi.netVAT >= 0 ? 'text-blue-300' : 'text-emerald-400']">
            {{ fmtCur(Math.abs(kpi.netVAT)) }}
          </p>
          <p class="text-xs mt-1" :class="kpi.netVAT >= 0 ? 'text-blue-400' : 'text-emerald-400'">
            {{ kpi.netVAT >= 0 ? 'A verser' : 'Credit TVA' }}
          </p>
        </div>
      </div>

      <!-- VAT by Rate Summary (sales only) -->
      <div v-if="activeTab === 'sales'" class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div v-for="rate in vatRateSummary" :key="rate.rate" class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-3">
            <span class="px-2 py-0.5 rounded-full bg-blue-900/50 text-blue-300 text-xs font-bold">TVA {{ rate.rate }}%</span>
            <span class="text-xs text-gray-500">{{ rate.count }} ops</span>
          </div>
          <div class="space-y-2">
            <div class="flex justify-between text-sm">
              <span class="text-gray-400">Base HT</span>
              <span class="text-gray-200 font-medium">{{ fmtCur(rate.base) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-gray-400">TVA</span>
              <span class="text-emerald-400 font-bold">{{ fmtCur(rate.vat) }}</span>
            </div>
            <div class="pt-2 border-t border-gray-800 flex justify-between text-sm">
              <span class="text-gray-400">TTC</span>
              <span class="text-white font-semibold">{{ fmtCur(rate.ttc) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-16">
        <div class="flex flex-col items-center gap-3">
          <div class="w-10 h-10 border-2 border-blue-500 border-t-transparent rounded-full animate-spin"></div>
          <span class="text-sm text-gray-400">Chargement du registre TVA...</span>
        </div>
      </div>

      <!-- Table -->
      <div v-else class="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden">
        <div class="px-5 py-4 border-b border-gray-800 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <BookOpen :size="16" class="text-blue-400" />
            <span class="text-sm font-semibold text-gray-200">
              Registre TVA — {{ activeTab === 'sales' ? 'Ventes' : 'Achats' }}
            </span>
            <span class="px-2 py-0.5 bg-gray-800 rounded-full text-xs text-gray-400">{{ filteredEntries.length }} entrées</span>
          </div>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-gray-800/60">
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Date</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">N° Document</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">Partenaire</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-gray-400 uppercase tracking-wide">NIF</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-400 uppercase tracking-wide">Base HT</th>
                <th class="px-4 py-3 text-center text-xs font-semibold text-gray-400 uppercase tracking-wide">Taux</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-400 uppercase tracking-wide">TVA</th>
                <th class="px-4 py-3 text-right text-xs font-semibold text-gray-400 uppercase tracking-wide">TTC</th>
                <th class="px-4 py-3 text-center text-xs font-semibold text-gray-400 uppercase tracking-wide">Statut</th>
                <th class="px-4 py-3 text-center text-xs font-semibold text-gray-400 uppercase tracking-wide">Actions</th>
              </tr>
            </thead>
            <tbody>
              <template v-if="filteredEntries.length > 0">
                <tr
                  v-for="entry in paginatedEntries"
                  :key="entry.id"
                  class="border-t border-gray-800 hover:bg-gray-800/40 transition-colors"
                >
                  <td class="px-4 py-3 text-gray-300">{{ fmtDate(entry.transaction_date) }}</td>
                  <td class="px-4 py-3">
                    <span class="font-mono text-blue-400 text-xs font-semibold">{{ entry.document_number }}</span>
                  </td>
                  <td class="px-4 py-3">
                    <div>
                      <p class="text-gray-200 font-medium">{{ entry.partner_name || '—' }}</p>
                      <p v-if="entry.partner_nis" class="text-xs text-gray-500">NIS: {{ entry.partner_nis }}</p>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-gray-400 text-xs font-mono">{{ entry.partner_nif || '—' }}</td>
                  <td class="px-4 py-3 text-right text-gray-200 font-medium">{{ fmtCur(entry.base_amount) }}</td>
                  <td class="px-4 py-3 text-center">
                    <span class="px-2 py-0.5 rounded-full text-xs font-bold"
                      :class="vatRateBadge(entry.vat_rate)">
                      {{ entry.vat_rate }}%
                    </span>
                  </td>
                  <td class="px-4 py-3 text-right font-semibold"
                    :class="activeTab === 'sales' ? 'text-emerald-400' : 'text-amber-400'">
                    {{ fmtCur(entry.vat_amount) }}
                  </td>
                  <td class="px-4 py-3 text-right text-white font-bold">{{ fmtCur(entry.total_amount) }}</td>
                  <td class="px-4 py-3 text-center">
                    <span :class="['px-2 py-0.5 rounded-full text-xs font-medium', statusBadge(entry.status)]">
                      {{ statusLabel(entry.status) }}
                    </span>
                  </td>
                  <td class="px-4 py-3 text-center">
                    <div class="flex items-center justify-center gap-1">
                      <button @click="viewEntry(entry)" class="p-1.5 rounded-lg hover:bg-gray-700 text-gray-400 hover:text-blue-400 transition-colors">
                        <Eye :size="14" />
                      </button>
                    </div>
                  </td>
                </tr>
              </template>
              <tr v-else>
                <td colspan="10" class="px-4 py-16 text-center">
                  <div class="flex flex-col items-center gap-3">
                    <div class="w-12 h-12 rounded-full bg-gray-800 flex items-center justify-center">
                      <BookOpen :size="22" class="text-gray-600" />
                    </div>
                    <p class="text-gray-500 text-sm">Aucune entrée pour cette période</p>
                    <button @click="openCreateModal" class="text-blue-400 hover:text-blue-300 text-xs underline">
                      Ajouter une entrée manuellement
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
            <!-- Footer totals -->
            <tfoot v-if="filteredEntries.length > 0">
              <tr class="bg-gray-800/80 border-t-2 border-gray-700">
                <td colspan="4" class="px-4 py-3 text-sm font-bold text-gray-300">TOTAL</td>
                <td class="px-4 py-3 text-right text-sm font-bold text-white">{{ fmtCur(totals.base) }}</td>
                <td></td>
                <td class="px-4 py-3 text-right text-sm font-bold" :class="activeTab === 'sales' ? 'text-emerald-400' : 'text-amber-400'">{{ fmtCur(totals.vat) }}</td>
                <td class="px-4 py-3 text-right text-sm font-bold text-white">{{ fmtCur(totals.ttc) }}</td>
                <td colspan="2"></td>
              </tr>
            </tfoot>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="px-5 py-3 border-t border-gray-800 flex items-center justify-between">
          <span class="text-xs text-gray-500">Page {{ currentPage }} / {{ totalPages }} — {{ filteredEntries.length }} entrées</span>
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

    <!-- Create Entry Modal -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
        <div class="bg-gray-900 border border-gray-700 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between p-6 border-b border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-blue-600 flex items-center justify-center">
                <Plus :size="18" class="text-white" />
              </div>
              <h2 class="text-lg font-bold text-white">Nouvelle Entrée TVA</h2>
            </div>
            <button @click="closeCreateModal" class="p-2 rounded-lg hover:bg-gray-800 text-gray-400 transition-colors">
              <X :size="18" />
            </button>
          </div>

          <form @submit.prevent="submitCreate" class="p-6 space-y-5">
            <div class="grid grid-cols-2 gap-4">
              <div class="col-span-2">
                <label class="lbl">Type d'opération</label>
                <div class="flex gap-3">
                  <label class="flex-1 flex items-center gap-2 cursor-pointer bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 hover:border-blue-600 transition-colors" :class="createForm.entry_type === 'sales' ? 'border-blue-600 bg-blue-900/20' : ''">
                    <input type="radio" v-model="createForm.entry_type" value="sales" class="accent-blue-500" />
                    <TrendingUp :size="15" class="text-blue-400" />
                    <span class="text-sm text-gray-200">Vente</span>
                  </label>
                  <label class="flex-1 flex items-center gap-2 cursor-pointer bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 hover:border-cyan-600 transition-colors" :class="createForm.entry_type === 'purchases' ? 'border-cyan-600 bg-cyan-900/20' : ''">
                    <input type="radio" v-model="createForm.entry_type" value="purchases" class="accent-cyan-500" />
                    <ShoppingCart :size="15" class="text-cyan-400" />
                    <span class="text-sm text-gray-200">Achat</span>
                  </label>
                </div>
              </div>

              <div>
                <label class="lbl">Date de transaction *</label>
                <input v-model="createForm.transaction_date" type="date" required class="field" />
              </div>
              <div>
                <label class="lbl">N° Document *</label>
                <input v-model="createForm.document_number" type="text" required placeholder="FAC-2024-001" class="field" />
              </div>
              <div>
                <label class="lbl">Partenaire</label>
                <input v-model="createForm.partner_name" type="text" placeholder="Nom du client/fournisseur" class="field" />
              </div>
              <div>
                <label class="lbl">NIF Partenaire</label>
                <input v-model="createForm.partner_nif" type="text" placeholder="000000000000000" class="field" />
              </div>
              <div>
                <label class="lbl">NIS Partenaire</label>
                <input v-model="createForm.partner_nis" type="text" placeholder="000000000000000" class="field" />
              </div>
              <div>
                <label class="lbl">Taux TVA *</label>
                <select v-model.number="createForm.vat_rate" required class="field">
                  <option :value="0">0% — Exonéré</option>
                  <option :value="9">9% — Taux réduit</option>
                  <option :value="19">19% — Taux normal</option>
                </select>
              </div>
              <div>
                <label class="lbl">Base HT *</label>
                <input v-model.number="createForm.base_amount" type="number" min="0" step="0.01" required class="field" @input="recalcCreate" />
              </div>
              <div>
                <label class="lbl">Montant TVA (calculé)</label>
                <div class="field bg-gray-950 text-gray-400 cursor-not-allowed">{{ fmtCur(createVATAmount) }}</div>
              </div>
              <div class="col-span-2">
                <label class="lbl">Description</label>
                <input v-model="createForm.description" type="text" placeholder="Description de l'opération" class="field" />
              </div>
            </div>

            <!-- Preview -->
            <div class="bg-gray-800/60 border border-gray-700 rounded-xl p-4 space-y-2">
              <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-3">Récapitulatif</p>
              <div class="flex justify-between text-sm"><span class="text-gray-400">Base HT</span><span class="text-gray-200">{{ fmtCur(createForm.base_amount || 0) }}</span></div>
              <div class="flex justify-between text-sm"><span class="text-gray-400">TVA {{ createForm.vat_rate }}%</span><span class="text-emerald-400">+ {{ fmtCur(createVATAmount) }}</span></div>
              <div class="flex justify-between text-sm font-bold border-t border-gray-700 pt-2"><span class="text-gray-300">Total TTC</span><span class="text-white">{{ fmtCur((createForm.base_amount || 0) + createVATAmount) }}</span></div>
            </div>

            <div class="flex gap-3 pt-2">
              <button type="button" @click="closeCreateModal" class="flex-1 px-4 py-2.5 bg-gray-800 border border-gray-700 text-gray-200 rounded-xl text-sm font-medium hover:bg-gray-700 transition-colors">
                Annuler
              </button>
              <button type="submit" :disabled="submitting" class="flex-1 px-4 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-sm font-bold transition-colors disabled:opacity-60 flex items-center justify-center gap-2">
                <div v-if="submitting" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                <Plus v-else :size="15" />
                Enregistrer
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- View Entry Modal -->
    <Teleport to="body">
      <div v-if="showViewModal && selectedEntry" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
        <div class="bg-gray-900 border border-gray-700 rounded-2xl shadow-2xl w-full max-w-lg">
          <div class="flex items-center justify-between p-6 border-b border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-gray-700 flex items-center justify-center">
                <Eye :size="18" class="text-blue-400" />
              </div>
              <div>
                <h2 class="text-lg font-bold text-white">Détail Entrée TVA</h2>
                <p class="text-xs text-gray-500">{{ selectedEntry.document_number }}</p>
              </div>
            </div>
            <button @click="showViewModal = false" class="p-2 rounded-lg hover:bg-gray-800 text-gray-400 transition-colors">
              <X :size="18" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-3 text-sm">
              <div class="bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 mb-1">Type</p>
                <p class="font-semibold text-gray-200 capitalize">{{ selectedEntry.entry_type === 'sales' ? 'Vente' : 'Achat' }}</p>
              </div>
              <div class="bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 mb-1">Date</p>
                <p class="font-semibold text-gray-200">{{ fmtDate(selectedEntry.transaction_date) }}</p>
              </div>
              <div class="bg-gray-800 rounded-lg p-3 col-span-2">
                <p class="text-xs text-gray-500 mb-1">Partenaire</p>
                <p class="font-semibold text-gray-200">{{ selectedEntry.partner_name || '—' }}</p>
                <p v-if="selectedEntry.partner_nif" class="text-xs text-gray-400 mt-0.5">NIF: {{ selectedEntry.partner_nif }}</p>
              </div>
              <div class="bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 mb-1">Base HT</p>
                <p class="font-bold text-white">{{ fmtCur(selectedEntry.base_amount) }}</p>
              </div>
              <div class="bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 mb-1">Taux TVA</p>
                <p class="font-bold text-blue-400">{{ selectedEntry.vat_rate }}%</p>
              </div>
              <div class="bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 mb-1">Montant TVA</p>
                <p class="font-bold text-emerald-400">{{ fmtCur(selectedEntry.vat_amount) }}</p>
              </div>
              <div class="bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 mb-1">Total TTC</p>
                <p class="font-bold text-white">{{ fmtCur(selectedEntry.total_amount) }}</p>
              </div>
            </div>
            <button @click="showViewModal = false" class="w-full px-4 py-2.5 bg-gray-800 border border-gray-700 text-gray-200 rounded-xl text-sm font-medium hover:bg-gray-700 transition-colors">
              Fermer
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  BookOpen, Download, Plus, Calendar, TrendingUp, ShoppingCart,
  Search, ArrowUpCircle, ArrowDownCircle, Calculator, Eye, X,
  ChevronLeft, ChevronRight
} from 'lucide-vue-next'
import { taxAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

const loading = ref(false)
const submitting = ref(false)
const activeTab = ref<'sales' | 'purchases'>('sales')
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = 20

const now = new Date()
const filterYear = ref(now.getFullYear())
const filterMonth = ref<number | ''>(now.getMonth() + 1)

const entries = ref<any[]>([])
const showCreateModal = ref(false)
const showViewModal = ref(false)
const selectedEntry = ref<any>(null)

const months = [
  { value: 1, label: 'Janvier' }, { value: 2, label: 'Février' }, { value: 3, label: 'Mars' },
  { value: 4, label: 'Avril' }, { value: 5, label: 'Mai' }, { value: 6, label: 'Juin' },
  { value: 7, label: 'Juillet' }, { value: 8, label: 'Août' }, { value: 9, label: 'Septembre' },
  { value: 10, label: 'Octobre' }, { value: 11, label: 'Novembre' }, { value: 12, label: 'Décembre' }
]
const yearOptions = computed(() => {
  const y = now.getFullYear()
  return [y + 1, y, y - 1, y - 2, y - 3]
})

const createForm = ref({
  entry_type: 'sales' as 'sales' | 'purchases',
  transaction_date: '',
  document_number: '',
  partner_name: '',
  partner_nif: '',
  partner_nis: '',
  vat_rate: 19,
  base_amount: 0,
  description: ''
})

const createVATAmount = computed(() => {
  return Math.round((createForm.value.base_amount || 0) * (createForm.value.vat_rate / 100) * 100) / 100
})

function recalcCreate() { /* computed handles it */ }

async function loadData() {
  loading.value = true
  currentPage.value = 1
  try {
    const params: Record<string, any> = {
      year: filterYear.value,
      type: activeTab.value
    }
    if (filterMonth.value !== '') params.month = filterMonth.value
    const res = await taxAPI.getVATRegister(params)
    entries.value = res.data?.entries || res.data || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur chargement registre TVA', 'error')
    entries.value = []
  } finally {
    loading.value = false
  }
}

const filteredEntries = computed(() => {
  let list = entries.value
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(e =>
      (e.partner_name || '').toLowerCase().includes(q) ||
      (e.document_number || '').toLowerCase().includes(q) ||
      (e.partner_nif || '').toLowerCase().includes(q)
    )
  }
  return list
})

const totalPages = computed(() => Math.ceil(filteredEntries.value.length / pageSize))
const paginatedEntries = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredEntries.value.slice(start, start + pageSize)
})

const kpi = computed(() => {
  const allEntries = entries.value
  const sales = allEntries.filter(e => e.entry_type === 'sales')
  const purchases = allEntries.filter(e => e.entry_type === 'purchases')
  const totalSalesVAT = sales.reduce((s, e) => s + (e.vat_amount || 0), 0)
  const totalPurchaseVAT = purchases.reduce((s, e) => s + (e.vat_amount || 0), 0)
  return {
    totalSalesBase: sales.reduce((s, e) => s + (e.base_amount || 0), 0),
    totalSalesVAT,
    totalPurchaseVAT,
    netVAT: totalSalesVAT - totalPurchaseVAT,
    salesCount: sales.length,
    purchaseCount: purchases.length
  }
})

const vatRateSummary = computed(() => {
  const rates = [0, 9, 19]
  const salesEntries = entries.value.filter(e => e.entry_type === 'sales')
  return rates.map(rate => {
    const subset = salesEntries.filter(e => Number(e.vat_rate) === rate)
    const base = subset.reduce((s, e) => s + (e.base_amount || 0), 0)
    const vat = subset.reduce((s, e) => s + (e.vat_amount || 0), 0)
    return { rate, count: subset.length, base, vat, ttc: base + vat }
  }).filter(r => r.count > 0 || r.rate === 19)
})

const totals = computed(() => {
  const list = filteredEntries.value
  return {
    base: list.reduce((s, e) => s + (e.base_amount || 0), 0),
    vat: list.reduce((s, e) => s + (e.vat_amount || 0), 0),
    ttc: list.reduce((s, e) => s + (e.total_amount || 0), 0)
  }
})

watch(activeTab, () => { currentPage.value = 1; loadData() })
watch(searchQuery, () => { currentPage.value = 1 })

function openCreateModal() {
  createForm.value = {
    entry_type: activeTab.value,
    transaction_date: new Date().toISOString().split('T')[0],
    document_number: '',
    partner_name: '',
    partner_nif: '',
    partner_nis: '',
    vat_rate: 19,
    base_amount: 0,
    description: ''
  }
  showCreateModal.value = true
}

function closeCreateModal() {
  showCreateModal.value = false
}

function viewEntry(entry: any) {
  selectedEntry.value = entry
  showViewModal.value = true
}

async function submitCreate() {
  submitting.value = true
  try {
    const payload = {
      ...createForm.value,
      vat_amount: createVATAmount.value,
      total_amount: (createForm.value.base_amount || 0) + createVATAmount.value
    }
    await taxAPI.createVATEntry(payload)
    store.addToast('Entrée TVA enregistrée', 'success')
    closeCreateModal()
    await loadData()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur création entrée', 'error')
  } finally {
    submitting.value = false
  }
}

function exportCSV() {
  const headers = ['Date', 'N° Document', 'Partenaire', 'NIF', 'Base HT', 'Taux TVA', 'Montant TVA', 'Total TTC', 'Statut']
  const rows = filteredEntries.value.map(e => [
    e.transaction_date, e.document_number, e.partner_name || '',
    e.partner_nif || '', e.base_amount, e.vat_rate,
    e.vat_amount, e.total_amount, e.status || ''
  ])
  const csv = [headers, ...rows].map(r => r.join(';')).join('\n')
  const blob = new Blob(['\uFEFF' + csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `vat_register_${activeTab.value}_${filterYear.value}${filterMonth.value ? '_' + filterMonth.value : ''}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

function fmtCur(v: number) {
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0) + ' DZD'
}
function fmtDate(d: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('fr-DZ', { day: '2-digit', month: '2-digit', year: 'numeric' })
}
function vatRateBadge(rate: number) {
  const r = Number(rate)
  if (r === 19) return 'bg-blue-900/60 text-blue-300 border border-blue-700/40'
  if (r === 9) return 'bg-amber-900/60 text-amber-300 border border-amber-700/40'
  return 'bg-gray-800 text-gray-400 border border-gray-700'
}
function statusBadge(s: string) {
  if (s === 'validated') return 'bg-emerald-900/50 text-emerald-300'
  if (s === 'cancelled') return 'bg-red-900/50 text-red-300'
  return 'bg-yellow-900/50 text-yellow-300'
}
function statusLabel(s: string) {
  if (s === 'validated') return 'Validé'
  if (s === 'cancelled') return 'Annulé'
  return 'Brouillon'
}

onMounted(loadData)
</script>

<style scoped>
.field {
  @apply w-full bg-gray-800 border border-gray-700 text-gray-200 rounded-xl px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all placeholder-gray-600;
}
.lbl {
  @apply block text-xs font-semibold text-gray-400 mb-1.5 uppercase tracking-wide;
}
</style>
