<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
          <GitMerge class="w-6 h-6 text-indigo-500" />
          Bank Reconciliation
        </h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Match bank statement transactions with system records</p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="load" class="inline-flex items-center gap-2 px-3 py-2 border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 text-sm rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
          <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          Refresh
        </button>
        <button @click="openCreate" class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm">
          <Plus class="w-4 h-4" /> New Reconciliation
        </button>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 xl:grid-cols-4 gap-4">
      <div v-for="kpi in kpis" :key="kpi.label"
        class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5 flex items-start gap-4 hover:shadow-md transition-shadow">
        <div :class="kpi.bg" class="w-12 h-12 rounded-xl flex items-center justify-center flex-shrink-0">
          <component :is="kpi.icon" :class="kpi.color" class="w-6 h-6" />
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-gray-400 uppercase tracking-wide">{{ kpi.label }}</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ kpi.value }}</p>
          <p class="text-xs text-gray-400 truncate">{{ kpi.sub }}</p>
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <Loader2 class="w-10 h-10 text-indigo-500 animate-spin" />
    </div>

    <template v-else>
      <!-- List View -->
      <div v-if="!selected">

        <!-- Filter Bar -->
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4 flex flex-wrap items-center gap-3">
          <div class="relative flex-1 min-w-48">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input v-model="search" type="text" placeholder="Search reference or bank..."
              class="field pl-9" />
          </div>
          <select v-model="filterStatus" class="field w-44">
            <option value="">All statuses</option>
            <option value="draft">Draft</option>
            <option value="in_progress">In Progress</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
          </select>
          <div class="text-xs text-gray-400 ml-auto">
            {{ filteredRecons.length }} of {{ recons.length }} sessions
          </div>
        </div>

        <!-- Table Card -->
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden mt-4">
          <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
            <h3 class="font-semibold text-gray-900 dark:text-white">Reconciliation Sessions</h3>
            <span class="text-xs text-gray-400">{{ filteredRecons.length }} total</span>
          </div>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
                  <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Reference</th>
                  <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Bank Account</th>
                  <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Period</th>
                  <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Statement Bal.</th>
                  <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">System Bal.</th>
                  <th class="text-right px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Difference</th>
                  <th class="text-left px-4 py-3 text-xs font-semibold text-gray-500 uppercase tracking-wide">Status</th>
                  <th class="px-4 py-3"></th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                <tr v-if="!filteredRecons.length">
                  <td colspan="8" class="text-center py-16">
                    <div class="flex flex-col items-center gap-3 text-gray-400">
                      <GitMerge class="w-10 h-10 opacity-30" />
                      <p class="text-sm">No reconciliations found</p>
                      <button @click="openCreate" class="text-indigo-600 text-xs hover:underline">Create your first one</button>
                    </div>
                  </td>
                </tr>
                <tr v-for="r in filteredRecons" :key="r.id"
                  class="hover:bg-indigo-50/40 dark:hover:bg-indigo-900/10 transition-colors cursor-pointer"
                  @click="viewDetail(r)">
                  <td class="px-4 py-3">
                    <span class="font-mono font-semibold text-indigo-600 dark:text-indigo-400 text-xs">{{ r.reference }}</span>
                  </td>
                  <td class="px-4 py-3">
                    <div>
                      <p class="font-medium text-gray-900 dark:text-white text-xs">{{ r.bank_name }}</p>
                      <p class="text-xs text-gray-400 font-mono">{{ r.account_number }}</p>
                    </div>
                  </td>
                  <td class="px-4 py-3">
                    <div class="text-xs text-gray-500">
                      <p>{{ fmtDate(r.period_start) }}</p>
                      <p class="text-gray-400">→ {{ fmtDate(r.period_end) }}</p>
                    </div>
                  </td>
                  <td class="px-4 py-3 text-right font-medium text-gray-900 dark:text-white text-xs">{{ fmtCurrency(r.statement_balance) }}</td>
                  <td class="px-4 py-3 text-right font-medium text-gray-900 dark:text-white text-xs">{{ fmtCurrency(r.system_balance) }}</td>
                  <td class="px-4 py-3 text-right">
                    <span :class="Math.abs(r.difference || 0) < 0.01
                      ? 'text-emerald-600 dark:text-emerald-400 font-bold'
                      : 'text-red-500 font-bold'" class="text-xs">
                      {{ fmtCurrency(r.difference) }}
                    </span>
                  </td>
                  <td class="px-4 py-3">
                    <span :class="statusBadge(r.status)" class="px-2 py-0.5 rounded-full text-xs font-semibold capitalize">
                      {{ r.status?.replace('_', ' ') }}
                    </span>
                  </td>
                  <td class="px-4 py-3" @click.stop>
                    <div class="flex items-center gap-1 justify-end">
                      <button @click="viewDetail(r)"
                        class="p-1.5 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/20 rounded-lg transition-colors"
                        title="View detail">
                        <Eye class="w-4 h-4" />
                      </button>
                      <button v-if="r.status !== 'completed' && r.status !== 'cancelled'"
                        @click="complete(r.id)"
                        class="px-2 py-1 text-xs bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400 rounded-lg hover:bg-emerald-200 transition-colors font-medium whitespace-nowrap">
                        Complete
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <!-- Footer totals -->
          <div v-if="filteredRecons.length" class="px-5 py-3 bg-gray-50 dark:bg-gray-800/50 border-t border-gray-100 dark:border-gray-800 flex items-center justify-between text-xs text-gray-500">
            <span>{{ completedCount }} completed · {{ inProgressCount }} in progress · {{ draftCount }} draft</span>
            <span class="font-medium text-gray-700 dark:text-gray-200">{{ filteredRecons.length }} sessions</span>
          </div>
        </div>
      </div>

      <!-- Detail View -->
      <div v-else>
        <!-- Breadcrumb -->
        <div class="flex items-center gap-2 text-sm mb-4">
          <button @click="selected = null; selectedLines = []"
            class="inline-flex items-center gap-1.5 text-gray-500 hover:text-indigo-600 transition-colors font-medium">
            <ArrowLeft class="w-4 h-4" />
            Reconciliations
          </button>
          <ChevronRight class="w-3 h-3 text-gray-300 dark:text-gray-600" />
          <span class="font-mono font-semibold text-indigo-600 dark:text-indigo-400">{{ selected.reference }}</span>
          <span :class="statusBadge(selected.status)" class="ml-2 px-2 py-0.5 rounded-full text-xs font-semibold capitalize">
            {{ selected.status?.replace('_', ' ') }}
          </span>
          <div class="ml-auto flex items-center gap-2">
            <button v-if="selected.status !== 'completed' && selected.status !== 'cancelled'"
              @click="complete(selected.id)"
              class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-emerald-600 hover:bg-emerald-700 text-white text-xs font-medium rounded-lg transition-colors">
              <CheckCircle class="w-3.5 h-3.5" />
              Mark Complete
            </button>
          </div>
        </div>

        <!-- Summary Cards -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
          <!-- Info Card -->
          <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
            <h4 class="font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2 text-sm">
              <Info class="w-4 h-4 text-indigo-400" /> Session Info
            </h4>
            <div class="space-y-3">
              <div class="flex justify-between items-center">
                <span class="text-xs text-gray-400">Bank</span>
                <span class="font-medium text-gray-900 dark:text-white text-xs">{{ selected.bank_name }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-xs text-gray-400">Account #</span>
                <span class="font-mono text-gray-700 dark:text-gray-200 text-xs">{{ selected.account_number }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-xs text-gray-400">Period Start</span>
                <span class="text-gray-700 dark:text-gray-200 text-xs">{{ fmtDate(selected.period_start) }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-xs text-gray-400">Period End</span>
                <span class="text-gray-700 dark:text-gray-200 text-xs">{{ fmtDate(selected.period_end) }}</span>
              </div>
              <div v-if="selected.notes" class="pt-2 border-t border-gray-100 dark:border-gray-800">
                <p class="text-xs text-gray-400 mb-1">Notes</p>
                <p class="text-xs text-gray-600 dark:text-gray-300">{{ selected.notes }}</p>
              </div>
            </div>
          </div>

          <!-- Balances Card -->
          <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
            <h4 class="font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2 text-sm">
              <Scale class="w-4 h-4 text-indigo-400" /> Balances
            </h4>
            <div class="space-y-3">
              <div class="flex justify-between items-center">
                <span class="text-xs text-gray-400">Opening Balance</span>
                <span class="font-medium text-gray-700 dark:text-gray-200 text-sm">{{ fmtCurrency(selected.opening_balance) }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-xs text-gray-400">Bank Statement</span>
                <span class="font-bold text-blue-600 dark:text-blue-400 text-sm">{{ fmtCurrency(selected.statement_balance) }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-xs text-gray-400">System Balance</span>
                <span class="font-bold text-gray-900 dark:text-white text-sm">{{ fmtCurrency(selected.system_balance) }}</span>
              </div>
              <div class="pt-2 border-t border-gray-100 dark:border-gray-800 flex justify-between items-center">
                <span class="text-xs font-semibold text-gray-500">Difference</span>
                <span :class="Math.abs(selected.difference || 0) < 0.01
                  ? 'text-emerald-600 dark:text-emerald-400'
                  : 'text-red-500'" class="font-bold text-lg">
                  {{ fmtCurrency(selected.difference) }}
                </span>
              </div>
              <div v-if="Math.abs(selected.difference || 0) < 0.01" class="flex items-center gap-1.5 text-xs text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-900/20 rounded-lg px-3 py-2">
                <CheckCircle class="w-3.5 h-3.5 flex-shrink-0" />
                Balanced — no discrepancy
              </div>
              <div v-else class="flex items-center gap-1.5 text-xs text-red-500 bg-red-50 dark:bg-red-900/20 rounded-lg px-3 py-2">
                <AlertTriangle class="w-3.5 h-3.5 flex-shrink-0" />
                Discrepancy detected
              </div>
            </div>
          </div>

          <!-- Matching Card -->
          <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-5">
            <h4 class="font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2 text-sm">
              <Link2 class="w-4 h-4 text-indigo-400" /> Matching Progress
            </h4>
            <div class="space-y-3">
              <div class="flex justify-between items-center">
                <span class="text-xs text-gray-400">Total Lines</span>
                <span class="font-bold text-gray-900 dark:text-white">{{ selectedLines.length }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-xs text-gray-400">Matched</span>
                <span class="font-bold text-emerald-600 dark:text-emerald-400">{{ matchedLinesCount }}</span>
              </div>
              <div class="flex justify-between items-center">
                <span class="text-xs text-gray-400">Unmatched</span>
                <span class="font-bold text-amber-600 dark:text-amber-400">{{ unmatchedLinesCount }}</span>
              </div>
              <!-- Progress bar -->
              <div class="pt-2 border-t border-gray-100 dark:border-gray-800">
                <div class="flex justify-between text-xs text-gray-400 mb-1.5">
                  <span>Matching rate</span>
                  <span class="font-semibold text-gray-700 dark:text-gray-200">{{ matchRate }}%</span>
                </div>
                <div class="h-2.5 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
                  <div class="h-full rounded-full transition-all duration-500"
                    :class="matchRate === 100 ? 'bg-emerald-500' : matchRate > 50 ? 'bg-blue-500' : 'bg-amber-400'"
                    :style="{ width: `${matchRate}%` }"></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Lines Table -->
        <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
          <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-800">
            <div class="flex items-center gap-3">
              <h3 class="font-semibold text-gray-900 dark:text-white">Statement Lines</h3>
              <span class="bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 text-xs font-medium px-2 py-0.5 rounded-full">
                {{ selectedLines.length }}
              </span>
            </div>
            <div class="flex items-center gap-2">
              <!-- Line type filter -->
              <select v-model="lineTypeFilter" class="text-xs border border-gray-200 dark:border-gray-700 rounded-lg px-2 py-1.5 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 outline-none">
                <option value="">All types</option>
                <option value="bank_statement">Bank Statement</option>
                <option value="system_transaction">System Transaction</option>
              </select>
              <button v-if="selected.status !== 'completed' && selected.status !== 'cancelled'"
                @click="showAddLine = true"
                class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 rounded-lg hover:bg-indigo-100 dark:hover:bg-indigo-900/50 font-medium transition-colors">
                <Plus class="w-3.5 h-3.5" />
                Add Line
              </button>
            </div>
          </div>

          <div v-if="!filteredLines.length" class="flex items-center justify-center py-14">
            <div class="text-center">
              <FileSearch class="w-10 h-10 text-gray-300 dark:text-gray-600 mx-auto mb-2" />
              <p class="text-sm text-gray-400">No lines added yet</p>
              <button v-if="selected.status !== 'completed' && selected.status !== 'cancelled'"
                @click="showAddLine = true"
                class="mt-2 text-xs text-indigo-600 hover:underline">
                Add a statement line
              </button>
            </div>
          </div>

          <div v-else class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="bg-gray-50 dark:bg-gray-800/50 border-b border-gray-200 dark:border-gray-700">
                  <th class="text-left px-4 py-2.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Date</th>
                  <th class="text-left px-4 py-2.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Description</th>
                  <th class="text-left px-4 py-2.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Reference</th>
                  <th class="text-right px-4 py-2.5 text-xs font-semibold text-red-500 uppercase tracking-wide">Debit</th>
                  <th class="text-right px-4 py-2.5 text-xs font-semibold text-emerald-600 uppercase tracking-wide">Credit</th>
                  <th class="text-left px-4 py-2.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Type</th>
                  <th class="text-left px-4 py-2.5 text-xs font-semibold text-gray-500 uppercase tracking-wide">Status</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                <tr v-for="l in filteredLines" :key="l.id"
                  class="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition-colors">
                  <td class="px-4 py-3 text-xs text-gray-500">{{ fmtDate(l.transaction_date) }}</td>
                  <td class="px-4 py-3 text-gray-900 dark:text-white text-xs max-w-48 truncate" :title="l.description">
                    {{ l.description || '—' }}
                  </td>
                  <td class="px-4 py-3 text-xs font-mono text-gray-400">{{ l.reference || '—' }}</td>
                  <td class="px-4 py-3 text-right">
                    <span v-if="l.debit_amount" class="text-red-500 font-medium text-xs">{{ fmtCurrency(l.debit_amount) }}</span>
                    <span v-else class="text-gray-300 text-xs">—</span>
                  </td>
                  <td class="px-4 py-3 text-right">
                    <span v-if="l.credit_amount" class="text-emerald-600 dark:text-emerald-400 font-medium text-xs">{{ fmtCurrency(l.credit_amount) }}</span>
                    <span v-else class="text-gray-300 text-xs">—</span>
                  </td>
                  <td class="px-4 py-3">
                    <span class="text-xs text-gray-400 capitalize bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded">
                      {{ l.line_type?.replace('_', ' ') || '—' }}
                    </span>
                  </td>
                  <td class="px-4 py-3">
                    <span :class="l.status === 'matched'
                      ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                      : l.status === 'unmatched'
                        ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                        : 'bg-gray-100 text-gray-500'"
                      class="px-2 py-0.5 rounded-full text-xs font-semibold capitalize">
                      {{ l.status || 'pending' }}
                    </span>
                  </td>
                </tr>
              </tbody>
              <!-- Totals footer -->
              <tfoot>
                <tr class="bg-gray-50 dark:bg-gray-800/50 border-t-2 border-gray-200 dark:border-gray-700">
                  <td colspan="3" class="px-4 py-3 text-xs font-semibold text-gray-500 uppercase">Totals</td>
                  <td class="px-4 py-3 text-right text-xs font-bold text-red-500">{{ fmtCurrency(linesDebitTotal) }}</td>
                  <td class="px-4 py-3 text-right text-xs font-bold text-emerald-600 dark:text-emerald-400">{{ fmtCurrency(linesCreditTotal) }}</td>
                  <td colspan="2" class="px-4 py-3 text-right text-xs font-bold text-gray-900 dark:text-white">
                    Net: {{ fmtCurrency(linesCreditTotal - linesDebitTotal) }}
                  </td>
                </tr>
              </tfoot>
            </table>
          </div>
        </div>
      </div>
    </template>

    <!-- ─── Create Reconciliation Modal ─── -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
        <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-lg border border-gray-200 dark:border-gray-800">
          <!-- Modal Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-indigo-100 dark:bg-indigo-900/30 flex items-center justify-center">
                <GitMerge class="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
              </div>
              <h2 class="text-lg font-bold text-gray-900 dark:text-white">New Reconciliation</h2>
            </div>
            <button @click="closeModal" class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors">
              <X class="w-5 h-5 text-gray-400" />
            </button>
          </div>
          <!-- Modal Body -->
          <div class="p-6 space-y-4 max-h-[70vh] overflow-y-auto">
            <div class="col-span-2">
              <label class="lbl">Bank Account *</label>
              <select v-model="form.bank_account_id" class="field">
                <option value="">— Select bank account —</option>
                <option v-for="ba in bankAccounts" :key="ba.id" :value="ba.id">
                  {{ ba.bank_name }} – {{ ba.account_number }}
                </option>
              </select>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="lbl">Period Start *</label>
                <input v-model="form.period_start" type="date" class="field" />
              </div>
              <div>
                <label class="lbl">Period End *</label>
                <input v-model="form.period_end" type="date" class="field" />
              </div>
              <div>
                <label class="lbl">Opening Balance (DZD)</label>
                <input v-model.number="form.opening_balance" type="number" min="0" step="0.01" class="field" placeholder="0.00" />
              </div>
              <div>
                <label class="lbl">Statement Balance (DZD) *</label>
                <input v-model.number="form.statement_balance" type="number" step="0.01" class="field" placeholder="0.00" />
              </div>
            </div>
            <div>
              <label class="lbl">Notes</label>
              <textarea v-model="form.notes" rows="3" class="field resize-none"
                placeholder="Optional notes about this reconciliation session..."></textarea>
            </div>
          </div>
          <!-- Modal Footer -->
          <div class="flex justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50 rounded-b-2xl">
            <button @click="closeModal"
              class="px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors">
              Cancel
            </button>
            <button @click="save" :disabled="saving"
              class="px-5 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-50 text-white text-sm font-semibold rounded-lg flex items-center gap-2 transition-colors shadow-sm">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              <span>{{ saving ? 'Creating...' : 'Create Reconciliation' }}</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ─── Add Line Modal ─── -->
    <Teleport to="body">
      <div v-if="showAddLine" class="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
        <div class="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-lg border border-gray-200 dark:border-gray-800">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-xl bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center">
                <Plus class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              </div>
              <h2 class="text-lg font-bold text-gray-900 dark:text-white">Add Statement Line</h2>
            </div>
            <button @click="showAddLine = false" class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors">
              <X class="w-5 h-5 text-gray-400" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="lbl">Line Type *</label>
                <select v-model="lineForm.line_type" class="field">
                  <option value="bank_statement">Bank Statement</option>
                  <option value="system_transaction">System Transaction</option>
                </select>
              </div>
              <div>
                <label class="lbl">Transaction Date *</label>
                <input v-model="lineForm.transaction_date" type="date" class="field" />
              </div>
              <div class="col-span-2">
                <label class="lbl">Description</label>
                <input v-model="lineForm.description" type="text" class="field"
                  placeholder="e.g., Supplier payment, Bank transfer..." />
              </div>
              <div>
                <label class="lbl">Debit Amount (DZD)</label>
                <input v-model.number="lineForm.debit_amount" type="number" min="0" step="0.01" class="field" placeholder="0.00" />
              </div>
              <div>
                <label class="lbl">Credit Amount (DZD)</label>
                <input v-model.number="lineForm.credit_amount" type="number" min="0" step="0.01" class="field" placeholder="0.00" />
              </div>
              <div class="col-span-2">
                <label class="lbl">External Reference</label>
                <input v-model="lineForm.reference" type="text" class="field"
                  placeholder="e.g., TXN-20240801-001" />
              </div>
            </div>
            <!-- Quick debit/credit helper note -->
            <p class="text-xs text-gray-400 italic">Enter Debit for outflows (withdrawals) and Credit for inflows (deposits). Leave the other field as 0.</p>
          </div>
          <div class="flex justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50 rounded-b-2xl">
            <button @click="showAddLine = false"
              class="px-4 py-2 text-sm font-medium text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors">
              Cancel
            </button>
            <button @click="addLine" :disabled="saving"
              class="px-5 py-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-semibold rounded-lg flex items-center gap-2 transition-colors">
              <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
              <span>{{ saving ? 'Adding...' : 'Add Line' }}</span>
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
  Plus, Loader2, X, Eye, ArrowLeft, RefreshCw,
  GitMerge, CheckCircle, Clock, AlertTriangle,
  Info, Scale, Link2, FileSearch, Search,
  ChevronRight
} from 'lucide-vue-next'
import { treasuryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

// ── State ──────────────────────────────────────────────
const loading    = ref(true)
const saving     = ref(false)
const showModal  = ref(false)
const showAddLine = ref(false)
const recons     = ref<any[]>([])
const bankAccounts = ref<any[]>([])
const selected   = ref<any>(null)
const selectedLines = ref<any[]>([])
const search     = ref('')
const filterStatus = ref('')
const lineTypeFilter = ref('')

// ── Forms ──────────────────────────────────────────────
const emptyForm = () => ({
  bank_account_id: '',
  period_start: '',
  period_end: '',
  opening_balance: 0,
  statement_balance: 0,
  notes: '',
})
const form = ref(emptyForm())

const emptyLineForm = () => ({
  line_type: 'bank_statement' as string,
  transaction_date: new Date().toISOString().slice(0, 10),
  description: '',
  reference: '',
  debit_amount: 0,
  credit_amount: 0,
})
const lineForm = ref(emptyLineForm())

// ── Computed ───────────────────────────────────────────
const filteredRecons = computed(() => {
  let list = recons.value
  if (filterStatus.value) list = list.filter(r => r.status === filterStatus.value)
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(r =>
      r.reference?.toLowerCase().includes(q) ||
      r.bank_name?.toLowerCase().includes(q) ||
      r.account_number?.toLowerCase().includes(q)
    )
  }
  return list
})

const completedCount  = computed(() => recons.value.filter(r => r.status === 'completed').length)
const inProgressCount = computed(() => recons.value.filter(r => r.status === 'in_progress').length)
const draftCount      = computed(() => recons.value.filter(r => r.status === 'draft').length)
const diffCount       = computed(() => recons.value.filter(r => Math.abs(r.difference || 0) > 0.01).length)

const filteredLines = computed(() => {
  if (!lineTypeFilter.value) return selectedLines.value
  return selectedLines.value.filter(l => l.line_type === lineTypeFilter.value)
})

const matchedLinesCount   = computed(() => selectedLines.value.filter(l => l.status === 'matched').length)
const unmatchedLinesCount = computed(() => selectedLines.value.filter(l => l.status !== 'matched').length)
const matchRate = computed(() => {
  if (!selectedLines.value.length) return 0
  return Math.round((matchedLinesCount.value / selectedLines.value.length) * 100)
})

const linesDebitTotal  = computed(() => selectedLines.value.reduce((s, l) => s + (l.debit_amount || 0), 0))
const linesCreditTotal = computed(() => selectedLines.value.reduce((s, l) => s + (l.credit_amount || 0), 0))

const kpis = computed(() => [
  {
    label: 'Total Sessions',
    value: recons.value.length,
    sub: 'All time',
    icon: GitMerge,
    bg: 'bg-indigo-100 dark:bg-indigo-900/30',
    color: 'text-indigo-600 dark:text-indigo-400',
  },
  {
    label: 'Completed',
    value: completedCount.value,
    sub: 'Fully reconciled',
    icon: CheckCircle,
    bg: 'bg-emerald-100 dark:bg-emerald-900/30',
    color: 'text-emerald-600 dark:text-emerald-400',
  },
  {
    label: 'In Progress',
    value: inProgressCount.value + draftCount.value,
    sub: 'Pending review',
    icon: Clock,
    bg: 'bg-amber-100 dark:bg-amber-900/30',
    color: 'text-amber-600 dark:text-amber-400',
  },
  {
    label: 'With Differences',
    value: diffCount.value,
    sub: 'Require attention',
    icon: AlertTriangle,
    bg: 'bg-red-100 dark:bg-red-900/30',
    color: 'text-red-500',
  },
])

// ── Formatters ─────────────────────────────────────────
function fmtCurrency(n?: number | null) {
  if (n == null) return '—'
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(n) + ' DZD'
}
function fmtDate(d?: string) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}
function statusBadge(s?: string) {
  switch (s) {
    case 'completed':   return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'in_progress': return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    case 'draft':       return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    case 'cancelled':   return 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'
    default:            return 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'
  }
}

// ── Actions ────────────────────────────────────────────
function openCreate() { form.value = emptyForm(); showModal.value = true }
function closeModal()  { showModal.value = false }

async function viewDetail(r: any) {
  try {
    const res = await treasuryAPI.getReconciliation(r.id)
    const detail = res.data
    selected.value = detail
    selectedLines.value = detail.lines || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Error loading detail', 'error')
  }
}

async function complete(id: string) {
  if (!confirm('Mark this reconciliation as completed? This cannot be undone.')) return
  try {
    await treasuryAPI.completeReconciliation(id)
    store.addToast('Reconciliation completed successfully', 'success')
    if (selected.value?.id === id) {
      selected.value = { ...selected.value, status: 'completed' }
    }
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Error completing reconciliation', 'error')
  }
}

async function save() {
  if (!form.value.bank_account_id) {
    store.addToast('Please select a bank account', 'error'); return
  }
  if (!form.value.period_start || !form.value.period_end) {
    store.addToast('Period start and end are required', 'error'); return
  }
  saving.value = true
  try {
    await treasuryAPI.createReconciliation({
      ...form.value,
      period_start: form.value.period_start + 'T00:00:00Z',
      period_end:   form.value.period_end + 'T00:00:00Z',
    })
    store.addToast('Reconciliation session created', 'success')
    closeModal()
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Error creating reconciliation', 'error')
  } finally {
    saving.value = false
  }
}

async function addLine() {
  if (!selected.value) return
  if (!lineForm.value.transaction_date) {
    store.addToast('Transaction date is required', 'error'); return
  }
  saving.value = true
  try {
    await treasuryAPI.addReconciliationLine(selected.value.id, {
      ...lineForm.value,
      transaction_date: lineForm.value.transaction_date + 'T00:00:00Z',
    })
    store.addToast('Line added successfully', 'success')
    showAddLine.value = false
    lineForm.value = emptyLineForm()
    await viewDetail(selected.value)
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Error adding line', 'error')
  } finally {
    saving.value = false
  }
}

async function load() {
  loading.value = true
  try {
    const res = await treasuryAPI.getReconciliations()
    recons.value = res.data || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load reconciliations', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await load()
  try {
    const baRes = await treasuryAPI.getBankAccounts()
    bankAccounts.value = baRes.data || []
  } catch { /* non-fatal */ }
})
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
