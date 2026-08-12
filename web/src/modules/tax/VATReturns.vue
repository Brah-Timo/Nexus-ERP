<template>
  <div class="min-h-screen bg-gray-950 text-gray-100">
    <!-- Header -->
    <div class="bg-gray-900 border-b border-gray-800 px-6 py-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shadow-lg">
            <FileCheck :size="20" class="text-white" />
          </div>
          <div>
            <h1 class="text-xl font-bold text-white">Déclarations TVA</h1>
            <p class="text-xs text-gray-400 mt-0.5">Retours TVA — Suivi &amp; Soumission</p>
          </div>
        </div>
        <button @click="openCreateModal" class="flex items-center gap-2 px-4 py-2 bg-violet-600 hover:bg-violet-500 text-white rounded-lg text-sm font-medium transition-colors shadow-lg shadow-violet-900/30">
          <Plus :size="15" />
          Nouvelle Déclaration
        </button>
      </div>
    </div>

    <div class="p-6 space-y-6">
      <!-- Year Filter -->
      <div class="bg-gray-900 rounded-xl border border-gray-800 p-4">
        <div class="flex flex-wrap items-center gap-4">
          <div class="flex items-center gap-2">
            <Calendar :size="16" class="text-gray-400" />
            <span class="text-sm text-gray-400 font-medium">Année :</span>
            <select v-model="filterYear" @change="loadReturns" class="bg-gray-800 border border-gray-700 text-gray-200 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500">
              <option v-for="y in yearOptions" :key="y" :value="y">{{ y }}</option>
            </select>
          </div>
          <div class="flex items-center gap-2 ml-2">
            <span class="text-sm text-gray-400 font-medium">Statut :</span>
            <div class="flex gap-1">
              <button v-for="s in statusFilters" :key="s.value"
                @click="filterStatus = s.value"
                :class="['px-3 py-1.5 rounded-lg text-xs font-medium transition-colors', filterStatus === s.value ? s.active : 'bg-gray-800 text-gray-400 hover:text-gray-200']">
                {{ s.label }}
              </button>
            </div>
          </div>
          <div class="ml-auto">
            <button @click="computeVATReturn" :disabled="computing" class="flex items-center gap-2 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-gray-200 rounded-lg text-sm font-medium transition-colors">
              <div v-if="computing" class="w-4 h-4 border-2 border-gray-400 border-t-white rounded-full animate-spin"></div>
              <Calculator v-else :size="15" />
              Calculer TVA
            </button>
          </div>
        </div>
      </div>

      <!-- KPI Row -->
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-gray-500 font-medium uppercase tracking-wide">Déclarations</span>
            <div class="w-7 h-7 rounded-lg bg-violet-500/15 flex items-center justify-center">
              <FileCheck :size="14" class="text-violet-400" />
            </div>
          </div>
          <p class="text-2xl font-bold text-white">{{ vatReturns.length }}</p>
          <p class="text-xs text-gray-500 mt-1">total cette année</p>
        </div>
        <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-gray-500 font-medium uppercase tracking-wide">TVA Collectée</span>
            <div class="w-7 h-7 rounded-lg bg-emerald-500/15 flex items-center justify-center">
              <ArrowUpCircle :size="14" class="text-emerald-400" />
            </div>
          </div>
          <p class="text-xl font-bold text-emerald-400">{{ fmtCur(annualKPI.totalCollected) }}</p>
          <p class="text-xs text-gray-500 mt-1">cumulé annuel</p>
        </div>
        <div class="bg-gray-900 border border-gray-800 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-gray-500 font-medium uppercase tracking-wide">TVA Déductible</span>
            <div class="w-7 h-7 rounded-lg bg-amber-500/15 flex items-center justify-center">
              <ArrowDownCircle :size="14" class="text-amber-400" />
            </div>
          </div>
          <p class="text-xl font-bold text-amber-400">{{ fmtCur(annualKPI.totalDeductible) }}</p>
          <p class="text-xs text-gray-500 mt-1">cumulé annuel</p>
        </div>
        <div class="bg-gradient-to-br from-violet-900/50 to-purple-900/30 border border-violet-700/40 rounded-xl p-4">
          <div class="flex items-center justify-between mb-2">
            <span class="text-xs text-violet-300 font-medium uppercase tracking-wide">Net Dû / Crédit</span>
            <div class="w-7 h-7 rounded-lg bg-violet-500/25 flex items-center justify-center">
              <Calculator :size="14" class="text-violet-400" />
            </div>
          </div>
          <p :class="['text-xl font-bold', annualKPI.netDue >= 0 ? 'text-violet-300' : 'text-emerald-400']">
            {{ fmtCur(Math.abs(annualKPI.netDue)) }}
          </p>
          <p class="text-xs mt-1" :class="annualKPI.netDue >= 0 ? 'text-violet-400' : 'text-emerald-400'">
            {{ annualKPI.netDue >= 0 ? 'A verser au DGI' : 'Credit a reporter' }}
          </p>
        </div>
      </div>

      <!-- Compute Result Panel -->
      <div v-if="computeResult" class="bg-gray-900 border border-violet-700/30 rounded-xl p-5">
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-2">
            <Calculator :size="16" class="text-violet-400" />
            <span class="text-sm font-semibold text-gray-200">Calcul TVA — Résultat</span>
            <span class="px-2 py-0.5 bg-violet-900/50 text-violet-300 rounded-full text-xs">{{ computeResult.period }}</span>
          </div>
          <button @click="computeResult = null" class="p-1.5 rounded-lg hover:bg-gray-800 text-gray-500 transition-colors"><X :size="14" /></button>
        </div>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-4">
          <div class="bg-gray-800 rounded-xl p-3">
            <p class="text-xs text-gray-500 mb-1">Ventes (Base)</p>
            <p class="text-sm font-bold text-white">{{ fmtCur(computeResult.sales_base || 0) }}</p>
          </div>
          <div class="bg-gray-800 rounded-xl p-3">
            <p class="text-xs text-gray-500 mb-1">TVA Collectée</p>
            <p class="text-sm font-bold text-emerald-400">{{ fmtCur(computeResult.tva_collected || 0) }}</p>
          </div>
          <div class="bg-gray-800 rounded-xl p-3">
            <p class="text-xs text-gray-500 mb-1">TVA Déductible</p>
            <p class="text-sm font-bold text-amber-400">{{ fmtCur(computeResult.tva_deductible || 0) }}</p>
          </div>
          <div class="bg-gray-800 rounded-xl p-3">
            <p class="text-xs text-gray-500 mb-1">Credit BF</p>
            <p class="text-sm font-bold text-violet-400">{{ fmtCur(computeResult.credit_bf || 0) }}</p>
          </div>
        </div>
        <!-- Rate breakdown -->
        <div v-if="computeResult.rates" class="space-y-2 mb-4">
          <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide">Ventilation par taux</p>
          <div class="grid grid-cols-3 gap-2">
            <div v-for="r in computeResult.rates" :key="r.rate" class="bg-gray-800 rounded-xl p-3">
              <p class="text-xs text-gray-500 mb-1">TVA {{ r.rate }}%</p>
              <p class="text-xs text-gray-400">Base: {{ fmtCur(r.base) }}</p>
              <p class="text-sm font-bold text-emerald-400">TVA: {{ fmtCur(r.vat) }}</p>
            </div>
          </div>
        </div>
        <div class="flex items-center justify-between bg-gradient-to-r from-violet-900/40 to-purple-900/20 border border-violet-700/30 rounded-xl p-4">
          <div>
            <p class="text-xs text-violet-300 font-medium">TVA Nette Due</p>
            <p :class="['text-2xl font-bold mt-1', (computeResult.tva_net_due || 0) >= 0 ? 'text-white' : 'text-emerald-400']">
              {{ fmtCur(Math.abs(computeResult.tva_net_due || 0)) }}
            </p>
            <p class="text-xs mt-0.5" :class="(computeResult.tva_net_due || 0) >= 0 ? 'text-violet-400' : 'text-emerald-400'">
              {{ (computeResult.tva_net_due || 0) >= 0 ? 'Montant a verser' : 'Credit a reporter' }}
            </p>
          </div>
          <button @click="prefillFromCompute" class="px-4 py-2 bg-violet-600 hover:bg-violet-500 text-white rounded-lg text-sm font-medium transition-colors">
            Utiliser ces valeurs
          </button>
        </div>
      </div>

      <!-- Returns List -->
      <div v-if="loading" class="flex items-center justify-center py-16">
        <div class="flex flex-col items-center gap-3">
          <div class="w-10 h-10 border-2 border-violet-500 border-t-transparent rounded-full animate-spin"></div>
          <span class="text-sm text-gray-400">Chargement déclarations TVA...</span>
        </div>
      </div>

      <div v-else class="space-y-3">
        <div v-if="filteredReturns.length === 0" class="bg-gray-900 border border-gray-800 rounded-xl py-16 text-center">
          <div class="flex flex-col items-center gap-3">
            <div class="w-14 h-14 rounded-full bg-gray-800 flex items-center justify-center">
              <FileCheck :size="26" class="text-gray-600" />
            </div>
            <p class="text-gray-400 font-medium">Aucune déclaration TVA pour {{ filterYear }}</p>
            <p class="text-gray-500 text-sm">Créez votre première déclaration TVA</p>
            <button @click="openCreateModal" class="mt-2 px-4 py-2 bg-violet-600 hover:bg-violet-500 text-white rounded-lg text-sm font-medium transition-colors">
              Nouvelle Déclaration
            </button>
          </div>
        </div>

        <div
          v-for="vr in filteredReturns"
          :key="vr.id"
          class="bg-gray-900 border border-gray-800 rounded-xl overflow-hidden hover:border-gray-700 transition-colors"
        >
          <div class="flex items-center justify-between p-5 cursor-pointer" @click="toggleExpand(vr.id)">
            <div class="flex items-center gap-4">
              <div :class="['w-10 h-10 rounded-xl flex items-center justify-center', statusIconBg(vr.status)]">
                <component :is="statusIcon(vr.status)" :size="18" class="text-white" />
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <p class="font-bold text-white">{{ periodLabel(vr.period_start, vr.period_end) }}</p>
                  <span :class="['px-2 py-0.5 rounded-full text-xs font-medium', statusBadge(vr.status)]">
                    {{ statusLabel(vr.status) }}
                  </span>
                </div>
                <p class="text-xs text-gray-500 mt-0.5">
                  Ref: {{ vr.reference || '—' }}
                  <span v-if="vr.filing_date" class="ml-2">• Soumis le {{ fmtDate(vr.filing_date) }}</span>
                </p>
              </div>
            </div>
            <div class="flex items-center gap-6">
              <div class="text-right">
                <p class="text-xs text-gray-500 mb-0.5">TVA Collectée</p>
                <p class="text-sm font-bold text-emerald-400">{{ fmtCur(vr.tva_collected) }}</p>
              </div>
              <div class="text-right">
                <p class="text-xs text-gray-500 mb-0.5">TVA Déductible</p>
                <p class="text-sm font-bold text-amber-400">{{ fmtCur(vr.tva_deductible) }}</p>
              </div>
              <div class="text-right">
                <p class="text-xs text-gray-500 mb-0.5">Credit BF</p>
                <p class="text-sm font-bold text-violet-400">{{ fmtCur(vr.credit_bf || 0) }}</p>
              </div>
              <div class="text-right min-w-[120px]">
                <p class="text-xs text-gray-500 mb-0.5">Net Dû</p>
                <p :class="['text-base font-bold', (vr.tva_net_due || 0) >= 0 ? 'text-white' : 'text-emerald-400']">
                  {{ fmtCur(Math.abs(vr.tva_net_due || 0)) }}
                </p>
                <p class="text-xs" :class="(vr.tva_net_due || 0) >= 0 ? 'text-gray-500' : 'text-emerald-500'">
                  {{ (vr.tva_net_due || 0) >= 0 ? 'a verser' : 'credit CF' }}
                </p>
              </div>
              <ChevronDown :size="16" class="text-gray-500 transition-transform" :class="expandedId === vr.id ? 'rotate-180' : ''" />
            </div>
          </div>

          <!-- Expanded Detail -->
          <div v-if="expandedId === vr.id" class="border-t border-gray-800 p-5 bg-gray-800/30">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <!-- Sales Breakdown -->
              <div>
                <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-3">Ventes par Taux</p>
                <div class="space-y-2">
                  <div class="flex justify-between items-center py-2 border-b border-gray-800">
                    <div class="flex items-center gap-2">
                      <span class="px-2 py-0.5 bg-blue-900/50 text-blue-300 rounded text-xs font-bold">19%</span>
                      <span class="text-sm text-gray-400">Base HT</span>
                    </div>
                    <div class="text-right">
                      <p class="text-sm text-gray-200">{{ fmtCur(vr.sales_base_19 || 0) }}</p>
                      <p class="text-xs text-emerald-400">TVA: {{ fmtCur(vr.sales_vat_19 || 0) }}</p>
                    </div>
                  </div>
                  <div class="flex justify-between items-center py-2 border-b border-gray-800">
                    <div class="flex items-center gap-2">
                      <span class="px-2 py-0.5 bg-amber-900/50 text-amber-300 rounded text-xs font-bold">9%</span>
                      <span class="text-sm text-gray-400">Base HT</span>
                    </div>
                    <div class="text-right">
                      <p class="text-sm text-gray-200">{{ fmtCur(vr.sales_base_9 || 0) }}</p>
                      <p class="text-xs text-emerald-400">TVA: {{ fmtCur(vr.sales_vat_9 || 0) }}</p>
                    </div>
                  </div>
                  <div class="flex justify-between items-center py-2">
                    <div class="flex items-center gap-2">
                      <span class="px-2 py-0.5 bg-gray-700 text-gray-400 rounded text-xs font-bold">0%</span>
                      <span class="text-sm text-gray-400">Exonéré</span>
                    </div>
                    <p class="text-sm text-gray-200">{{ fmtCur(vr.sales_base_0 || 0) }}</p>
                  </div>
                </div>
              </div>

              <!-- Purchase & Net -->
              <div>
                <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-3">Achats &amp; Calcul Net</p>
                <div class="space-y-2">
                  <div class="flex justify-between items-center py-2 border-b border-gray-800">
                    <span class="text-sm text-gray-400">TVA sur achats déductible</span>
                    <span class="text-sm font-semibold text-amber-400">{{ fmtCur(vr.tva_deductible || 0) }}</span>
                  </div>
                  <div class="flex justify-between items-center py-2 border-b border-gray-800">
                    <span class="text-sm text-gray-400">Credit TVA reporté (BF)</span>
                    <span class="text-sm font-semibold text-violet-400">{{ fmtCur(vr.credit_bf || 0) }}</span>
                  </div>
                  <div class="flex justify-between items-center py-2 border-b border-gray-800">
                    <span class="text-sm text-gray-400">Total déductions</span>
                    <span class="text-sm font-bold text-gray-200">
                      {{ fmtCur((vr.tva_deductible || 0) + (vr.credit_bf || 0)) }}
                    </span>
                  </div>
                  <div class="flex justify-between items-center py-2 bg-gray-800 rounded-xl px-3">
                    <span class="text-sm font-bold text-gray-200">TVA Nette Due</span>
                    <span :class="['text-base font-bold', (vr.tva_net_due || 0) >= 0 ? 'text-white' : 'text-emerald-400']">
                      {{ fmtCur(Math.abs(vr.tva_net_due || 0)) }}
                    </span>
                  </div>
                  <div v-if="(vr.credit_cf || 0) > 0" class="flex justify-between items-center py-2">
                    <span class="text-sm text-violet-400">Credit TVA a reporter (CF)</span>
                    <span class="text-sm font-bold text-violet-400">{{ fmtCur(vr.credit_cf || 0) }}</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-3 mt-4 pt-4 border-t border-gray-800">
              <button v-if="vr.status === 'draft'" @click="openEditModal(vr)" class="flex items-center gap-2 px-3 py-2 bg-gray-700 hover:bg-gray-600 text-gray-200 rounded-lg text-sm transition-colors">
                <Edit3 :size="14" />
                Modifier
              </button>
              <button v-if="vr.status === 'draft'" @click="submitReturn(vr.id)" :disabled="submitting" class="flex items-center gap-2 px-3 py-2 bg-violet-600 hover:bg-violet-500 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-60">
                <Send :size="14" />
                Soumettre
              </button>
              <div v-if="vr.status !== 'draft'" class="flex items-center gap-2 text-xs text-gray-500">
                <CheckCircle2 :size="14" class="text-emerald-400" />
                Soumis le {{ fmtDate(vr.filing_date) }}
              </div>
              <div v-if="vr.notes" class="ml-auto text-xs text-gray-500 italic">{{ vr.notes }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm">
        <div class="bg-gray-900 border border-gray-700 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
          <div class="flex items-center justify-between p-6 border-b border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-violet-600 flex items-center justify-center">
                <FileCheck :size="18" class="text-white" />
              </div>
              <h2 class="text-lg font-bold text-white">{{ editMode ? 'Modifier Déclaration' : 'Nouvelle Déclaration TVA' }}</h2>
            </div>
            <button @click="closeModal" class="p-2 rounded-lg hover:bg-gray-800 text-gray-400 transition-colors">
              <X :size="18" />
            </button>
          </div>

          <form @submit.prevent="submitModal" class="p-6 space-y-5">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="lbl">Période Début *</label>
                <input v-model="form.period_start" type="date" required class="field" />
              </div>
              <div>
                <label class="lbl">Période Fin *</label>
                <input v-model="form.period_end" type="date" required class="field" />
              </div>

              <!-- Sales Section -->
              <div class="col-span-2">
                <p class="text-xs font-bold text-gray-300 uppercase tracking-wide mb-3 flex items-center gap-2">
                  <TrendingUp :size="13" class="text-emerald-400" /> Ventes par taux
                </p>
              </div>
              <div>
                <label class="lbl">Base HT (TVA 19%)</label>
                <input v-model.number="form.sales_base_19" type="number" min="0" step="0.01" class="field" @input="recalc" />
              </div>
              <div>
                <label class="lbl">TVA 19% (calculé)</label>
                <div class="field bg-gray-950 text-emerald-400 cursor-not-allowed font-semibold">{{ fmtCur(form.sales_base_19 * 0.19) }}</div>
              </div>
              <div>
                <label class="lbl">Base HT (TVA 9%)</label>
                <input v-model.number="form.sales_base_9" type="number" min="0" step="0.01" class="field" @input="recalc" />
              </div>
              <div>
                <label class="lbl">TVA 9% (calculé)</label>
                <div class="field bg-gray-950 text-emerald-400 cursor-not-allowed font-semibold">{{ fmtCur(form.sales_base_9 * 0.09) }}</div>
              </div>
              <div>
                <label class="lbl">Base HT (Exonéré 0%)</label>
                <input v-model.number="form.sales_base_0" type="number" min="0" step="0.01" class="field" @input="recalc" />
              </div>
              <div>
                <label class="lbl">TVA Collectée Totale</label>
                <div class="field bg-gray-950 text-emerald-400 cursor-not-allowed font-bold">{{ fmtCur(computedCollected) }}</div>
              </div>

              <!-- Purchase / Deduction Section -->
              <div class="col-span-2 mt-2">
                <p class="text-xs font-bold text-gray-300 uppercase tracking-wide mb-3 flex items-center gap-2">
                  <ShoppingCart :size="13" class="text-amber-400" /> Déductions
                </p>
              </div>
              <div>
                <label class="lbl">TVA sur Achats Déductible</label>
                <input v-model.number="form.tva_deductible" type="number" min="0" step="0.01" class="field" @input="recalc" />
              </div>
              <div>
                <label class="lbl">Credit TVA BF (reporté)</label>
                <input v-model.number="form.credit_bf" type="number" min="0" step="0.01" class="field" @input="recalc" />
              </div>

              <div class="col-span-2">
                <label class="lbl">Notes</label>
                <textarea v-model="form.notes" rows="2" class="field resize-none" placeholder="Remarques, références DGI..."></textarea>
              </div>
            </div>

            <!-- Net Summary -->
            <div class="bg-gray-800/60 border border-gray-700 rounded-xl p-4 space-y-2">
              <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-3">Calcul TVA Nette</p>
              <div class="flex justify-between text-sm"><span class="text-gray-400">TVA collectée</span><span class="text-emerald-400 font-semibold">{{ fmtCur(computedCollected) }}</span></div>
              <div class="flex justify-between text-sm"><span class="text-gray-400">— TVA déductible</span><span class="text-amber-400 font-semibold">- {{ fmtCur(form.tva_deductible || 0) }}</span></div>
              <div class="flex justify-between text-sm"><span class="text-gray-400">— Credit BF</span><span class="text-violet-400 font-semibold">- {{ fmtCur(form.credit_bf || 0) }}</span></div>
              <div class="flex justify-between text-base font-bold border-t border-gray-700 pt-2">
                <span class="text-gray-200">TVA Nette Due</span>
                <span :class="computedNetDue >= 0 ? 'text-white' : 'text-emerald-400'">
                  {{ fmtCur(Math.abs(computedNetDue)) }}
                  <span class="text-xs font-normal ml-1">{{ computedNetDue >= 0 ? 'a verser' : 'credit CF' }}</span>
                </span>
              </div>
            </div>

            <div class="flex gap-3 pt-2">
              <button type="button" @click="closeModal" class="flex-1 px-4 py-2.5 bg-gray-800 border border-gray-700 text-gray-200 rounded-xl text-sm font-medium hover:bg-gray-700 transition-colors">
                Annuler
              </button>
              <button type="submit" :disabled="submitting" class="flex-1 px-4 py-2.5 bg-violet-600 hover:bg-violet-500 text-white rounded-xl text-sm font-bold transition-colors disabled:opacity-60 flex items-center justify-center gap-2">
                <div v-if="submitting" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                <Save v-else :size="15" />
                {{ editMode ? 'Enregistrer' : 'Créer' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  FileCheck, Plus, Calendar, ArrowUpCircle, ArrowDownCircle,
  Calculator, X, TrendingUp, ShoppingCart, ChevronDown,
  Edit3, Send, CheckCircle2, Save, Clock, AlertCircle
} from 'lucide-vue-next'
import { taxAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()
const loading = ref(false)
const submitting = ref(false)
const computing = ref(false)

const now = new Date()
const filterYear = ref(now.getFullYear())
const filterStatus = ref('all')
const vatReturns = ref<any[]>([])
const expandedId = ref<number | null>(null)
const computeResult = ref<any>(null)
const showModal = ref(false)
const editMode = ref(false)

const statusFilters = [
  { value: 'all', label: 'Tous', active: 'bg-gray-600 text-white' },
  { value: 'draft', label: 'Brouillon', active: 'bg-yellow-700 text-yellow-200' },
  { value: 'submitted', label: 'Soumis', active: 'bg-emerald-700 text-emerald-200' },
  { value: 'accepted', label: 'Accepté', active: 'bg-blue-700 text-blue-200' }
]

const yearOptions = computed(() => {
  const y = now.getFullYear()
  return [y + 1, y, y - 1, y - 2, y - 3]
})

const defaultForm = () => ({
  id: null as number | null,
  period_start: '',
  period_end: '',
  sales_base_19: 0,
  sales_base_9: 0,
  sales_base_0: 0,
  tva_deductible: 0,
  credit_bf: 0,
  notes: ''
})
const form = ref(defaultForm())

const computedCollected = computed(() => {
  return (form.value.sales_base_19 || 0) * 0.19 + (form.value.sales_base_9 || 0) * 0.09
})
const computedNetDue = computed(() => {
  return computedCollected.value - (form.value.tva_deductible || 0) - (form.value.credit_bf || 0)
})

function recalc() { /* computed handles it */ }

const filteredReturns = computed(() => {
  if (filterStatus.value === 'all') return vatReturns.value
  return vatReturns.value.filter(r => r.status === filterStatus.value)
})

const annualKPI = computed(() => {
  const all = vatReturns.value
  const totalCollected = all.reduce((s, r) => s + (r.tva_collected || 0), 0)
  const totalDeductible = all.reduce((s, r) => s + (r.tva_deductible || 0), 0)
  const netDue = all.reduce((s, r) => s + (r.tva_net_due || 0), 0)
  return { totalCollected, totalDeductible, netDue }
})

async function loadReturns() {
  loading.value = true
  try {
    const res = await taxAPI.listVATReturns({ year: filterYear.value })
    vatReturns.value = res.data?.returns || res.data || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur chargement déclarations', 'error')
    vatReturns.value = []
  } finally {
    loading.value = false
  }
}

async function computeVATReturn() {
  computing.value = true
  try {
    const month = now.getMonth() + 1
    const res = await taxAPI.computeVATReturn({ year: filterYear.value, month })
    computeResult.value = res.data
    store.addToast('Calcul TVA effectué', 'success')
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur calcul TVA', 'error')
  } finally {
    computing.value = false
  }
}

function prefillFromCompute() {
  if (!computeResult.value) return
  const m = computeResult.value
  form.value.sales_base_19 = m.rates?.find((r: any) => r.rate === 19)?.base || 0
  form.value.sales_base_9 = m.rates?.find((r: any) => r.rate === 9)?.base || 0
  form.value.sales_base_0 = m.rates?.find((r: any) => r.rate === 0)?.base || 0
  form.value.tva_deductible = m.tva_deductible || 0
  form.value.credit_bf = m.credit_bf || 0
  showModal.value = true
  editMode.value = false
}

function toggleExpand(id: number) {
  expandedId.value = expandedId.value === id ? null : id
}

function openCreateModal() {
  form.value = defaultForm()
  form.value.period_start = `${filterYear.value}-01-01`
  form.value.period_end = `${filterYear.value}-01-31`
  editMode.value = false
  showModal.value = true
}

function openEditModal(vr: any) {
  form.value = {
    id: vr.id,
    period_start: vr.period_start?.split('T')[0] || '',
    period_end: vr.period_end?.split('T')[0] || '',
    sales_base_19: vr.sales_base_19 || 0,
    sales_base_9: vr.sales_base_9 || 0,
    sales_base_0: vr.sales_base_0 || 0,
    tva_deductible: vr.tva_deductible || 0,
    credit_bf: vr.credit_bf || 0,
    notes: vr.notes || ''
  }
  editMode.value = true
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  form.value = defaultForm()
}

async function submitModal() {
  submitting.value = true
  try {
    const payload = {
      ...form.value,
      tva_collected: computedCollected.value,
      sales_vat_19: (form.value.sales_base_19 || 0) * 0.19,
      sales_vat_9: (form.value.sales_base_9 || 0) * 0.09
    }
    if (editMode.value && form.value.id) {
      await taxAPI.updateVATReturn(form.value.id, payload)
      store.addToast('Déclaration mise à jour', 'success')
    } else {
      await taxAPI.createVATReturn(payload)
      store.addToast('Déclaration créée', 'success')
    }
    closeModal()
    await loadReturns()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur sauvegarde', 'error')
  } finally {
    submitting.value = false
  }
}

async function submitReturn(id: number) {
  if (!confirm('Soumettre cette déclaration TVA au DGI ?')) return
  submitting.value = true
  try {
    await taxAPI.submitVATReturn(id)
    store.addToast('Déclaration TVA soumise avec succès', 'success')
    await loadReturns()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Erreur soumission', 'error')
  } finally {
    submitting.value = false
  }
}

function periodLabel(start: string, end: string) {
  if (!start) return '—'
  const s = new Date(start)
  const e = end ? new Date(end) : null
  const opts: Intl.DateTimeFormatOptions = { month: 'long', year: 'numeric' }
  if (e && s.getMonth() === e.getMonth() && s.getFullYear() === e.getFullYear()) {
    return s.toLocaleDateString('fr-DZ', opts)
  }
  return `${s.toLocaleDateString('fr-DZ', opts)} — ${e ? e.toLocaleDateString('fr-DZ', opts) : ''}`
}
function fmtCur(v: number) {
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0) + ' DZD'
}
function fmtDate(d: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('fr-DZ', { day: '2-digit', month: '2-digit', year: 'numeric' })
}
function statusBadge(s: string) {
  if (s === 'submitted') return 'bg-emerald-900/50 text-emerald-300'
  if (s === 'accepted') return 'bg-blue-900/50 text-blue-300'
  if (s === 'rejected') return 'bg-red-900/50 text-red-300'
  return 'bg-yellow-900/50 text-yellow-300'
}
function statusLabel(s: string) {
  const m: Record<string, string> = { draft: 'Brouillon', submitted: 'Soumis', accepted: 'Accepté', rejected: 'Rejeté' }
  return m[s] || s
}
function statusIconBg(s: string) {
  if (s === 'submitted' || s === 'accepted') return 'bg-emerald-600'
  if (s === 'rejected') return 'bg-red-600'
  return 'bg-yellow-600'
}
function statusIcon(s: string) {
  if (s === 'submitted' || s === 'accepted') return CheckCircle2
  if (s === 'rejected') return AlertCircle
  return Clock
}

onMounted(loadReturns)
</script>

<style scoped>
.field {
  @apply w-full bg-gray-800 border border-gray-700 text-gray-200 rounded-xl px-4 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-violet-500 transition-all placeholder-gray-600;
}
.lbl {
  @apply block text-xs font-semibold text-gray-400 mb-1.5 uppercase tracking-wide;
}
</style>
