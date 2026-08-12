<template>
  <div class="flex flex-col h-full bg-gray-50 dark:bg-gray-950 min-h-screen">

    <!-- Header -->
    <div class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Inventory Counts</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Physical inventory counts with variance analysis</p>
        </div>
        <div class="flex items-center gap-3">
          <button @click="openCreate"
            class="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            <Plus :size="16" />
            New Count
          </button>
        </div>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="px-6 py-4 grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Total Counts</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ counts.length }}</p>
          </div>
          <div class="w-10 h-10 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg flex items-center justify-center">
            <ClipboardList :size="20" class="text-indigo-600 dark:text-indigo-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Draft</p>
            <p class="text-2xl font-bold text-amber-600 dark:text-amber-400 mt-1">{{ countsByStatus('draft') }}</p>
          </div>
          <div class="w-10 h-10 bg-amber-100 dark:bg-amber-900/40 rounded-lg flex items-center justify-center">
            <FileEdit :size="20" class="text-amber-600 dark:text-amber-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Validated</p>
            <p class="text-2xl font-bold text-emerald-600 dark:text-emerald-400 mt-1">{{ countsByStatus('validated') }}</p>
          </div>
          <div class="w-10 h-10 bg-emerald-100 dark:bg-emerald-900/40 rounded-lg flex items-center justify-center">
            <CheckCircle :size="20" class="text-emerald-600 dark:text-emerald-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">This Month</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ thisMonthCount }}</p>
          </div>
          <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900/40 rounded-lg flex items-center justify-center">
            <Calendar :size="20" class="text-blue-600 dark:text-blue-400" />
          </div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="px-6 pb-4 flex flex-wrap gap-3">
      <div class="relative flex-1 min-w-64">
        <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
        <input v-model="search" placeholder="Search by count number, warehouse..." type="text"
          class="w-full pl-9 pr-4 py-2 text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
      </div>
      <select v-model="filterStatus"
        class="text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
        <option value="">All Status</option>
        <option value="draft">Draft</option>
        <option value="validated">Validated</option>
      </select>
      <select v-model="filterWarehouse"
        class="text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
        <option value="">All Warehouses</option>
        <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
      </select>
    </div>

    <!-- List -->
    <div class="flex-1 px-6 pb-6">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div v-if="loading" class="flex items-center justify-center py-16">
          <Loader2 :size="28" class="animate-spin text-indigo-500" />
          <span class="ml-3 text-gray-500 dark:text-gray-400">Loading counts...</span>
        </div>
        <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center py-16">
          <ClipboardList :size="48" class="text-gray-300 dark:text-gray-600 mb-3" />
          <p class="text-gray-500 dark:text-gray-400 font-medium">No inventory counts found</p>
          <p class="text-sm text-gray-400 dark:text-gray-500 mt-1">Create your first count to start reconciliation</p>
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th v-for="col in columns" :key="col.key"
                  @click="sortBy(col.key)"
                  class="px-4 py-3 text-left font-semibold text-gray-600 dark:text-gray-300 cursor-pointer select-none hover:text-gray-900 dark:hover:text-white transition-colors whitespace-nowrap">
                  <div class="flex items-center gap-1">
                    {{ col.label }}
                    <component :is="sortIcon(col.key)" :size="12" class="text-gray-400" />
                  </div>
                </th>
                <th class="px-4 py-3 text-right font-semibold text-gray-600 dark:text-gray-300">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
              <tr v-for="count in paginated" :key="count.id"
                class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
                <td class="px-4 py-3">
                  <span class="font-mono text-xs font-semibold text-indigo-600 dark:text-indigo-400">{{ count.number }}</span>
                </td>
                <td class="px-4 py-3 text-gray-700 dark:text-gray-300">{{ count.date }}</td>
                <td class="px-4 py-3 text-gray-700 dark:text-gray-300">{{ count.warehouse_name || '-' }}</td>
                <td class="px-4 py-3">
                  <span :class="statusBadge(count.status)" class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium">
                    <component :is="statusIcon(count.status)" :size="10" />
                    {{ statusLabel(count.status) }}
                  </span>
                </td>
                <td class="px-4 py-3 text-xs text-gray-500 dark:text-gray-400">
                  {{ count.validated_at ? count.validated_at.slice(0, 10) : '-' }}
                </td>
                <td class="px-4 py-3 text-xs text-gray-500 dark:text-gray-400">
                  {{ count.created_at ? count.created_at.slice(0, 10) : '-' }}
                </td>
                <td class="px-4 py-3 text-xs text-gray-500 dark:text-gray-400 max-w-[160px] truncate">
                  {{ count.notes || '-' }}
                </td>
                <td class="px-4 py-3 text-right" @click.stop>
                  <div class="flex items-center justify-end gap-1">
                    <button @click="openDetail(count.id)"
                      class="inline-flex items-center gap-1 text-xs text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 px-2 py-1 rounded-lg hover:bg-indigo-50 dark:hover:bg-indigo-900/30 transition-colors font-medium">
                      <Eye :size="12" />
                      View
                    </button>
                    <button v-if="count.status === 'draft'" @click="validate(count)"
                      class="inline-flex items-center gap-1 text-xs text-emerald-600 dark:text-emerald-400 hover:text-emerald-700 px-2 py-1 rounded-lg hover:bg-emerald-50 dark:hover:bg-emerald-900/30 transition-colors font-medium">
                      <CheckCircle :size="12" />
                      Validate
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="!loading && filtered.length > 0" class="border-t border-gray-200 dark:border-gray-700 px-4 py-3 flex items-center justify-between">
          <p class="text-sm text-gray-500 dark:text-gray-400">
            Showing {{ (currentPage - 1) * pageSize + 1 }}–{{ Math.min(currentPage * pageSize, filtered.length) }} of {{ filtered.length }}
          </p>
          <div class="flex items-center gap-1">
            <button @click="currentPage--" :disabled="currentPage === 1"
              class="p-1.5 rounded-lg text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
              <ChevronLeft :size="16" />
            </button>
            <span class="text-sm text-gray-600 dark:text-gray-300 px-2">{{ currentPage }} / {{ totalPages }}</span>
            <button @click="currentPage++" :disabled="currentPage === totalPages"
              class="p-1.5 rounded-lg text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
              <ChevronRight :size="16" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Modal -->
    <Teleport to="body">
      <div v-if="showCreate" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showCreate = false"></div>
        <div class="relative bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-3xl max-h-[90vh] flex flex-col">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800">
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">New Inventory Count</h2>
            <button @click="showCreate = false" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 transition-colors">
              <X :size="18" />
            </button>
          </div>
          <div class="flex-1 overflow-y-auto px-6 py-4 space-y-5">
            <!-- Header -->
            <div class="grid grid-cols-3 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Date <span class="text-red-500">*</span></label>
                <input v-model="createForm.date" type="date"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Warehouse <span class="text-red-500">*</span></label>
                <select v-model="createForm.warehouse_id"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                  <option value="">Select warehouse...</option>
                  <option v-for="w in warehouses" :key="w.id" :value="w.id">{{ w.name }}</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Notes</label>
                <input v-model="createForm.notes" type="text" placeholder="Optional notes"
                  class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
            </div>

            <!-- Lines -->
            <div>
              <div class="flex items-center justify-between mb-3">
                <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Count Lines</h3>
                <button @click="addLine"
                  class="inline-flex items-center gap-1.5 text-xs text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 font-medium px-2 py-1 rounded-lg hover:bg-indigo-50 dark:hover:bg-indigo-900/30 transition-colors">
                  <Plus :size="12" />
                  Add Line
                </button>
              </div>
              <div class="rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
                <table class="w-full text-sm">
                  <thead class="bg-gray-50 dark:bg-gray-800">
                    <tr>
                      <th class="px-3 py-2 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">Item</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">Book Qty</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">Counted Qty</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">Unit Cost</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-gray-600 dark:text-gray-300 w-8"></th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                    <tr v-if="createForm.lines.length === 0">
                      <td colspan="5" class="px-3 py-6 text-center text-gray-400 dark:text-gray-500 text-xs">
                        No lines added — click "Add Line" to begin
                      </td>
                    </tr>
                    <tr v-for="(line, idx) in createForm.lines" :key="idx" class="bg-white dark:bg-gray-900">
                      <td class="px-3 py-2">
                        <select v-model="line.item_id" @change="onItemChange(line)"
                          class="w-full text-xs bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded px-2 py-1.5 text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-indigo-500">
                          <option value="">Select item...</option>
                          <option v-for="item in items" :key="item.id" :value="item.id">
                            {{ item.code }} – {{ item.name }}
                          </option>
                        </select>
                      </td>
                      <td class="px-3 py-2">
                        <input v-model.number="line.book_qty" type="number" min="0" step="0.001"
                          class="w-24 text-xs bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded px-2 py-1.5 text-right text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-indigo-500 ml-auto block" />
                      </td>
                      <td class="px-3 py-2">
                        <input v-model.number="line.counted_qty" type="number" min="0" step="0.001"
                          class="w-24 text-xs bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded px-2 py-1.5 text-right focus:outline-none focus:ring-1 focus:ring-indigo-500 ml-auto block"
                          :class="getDiffClass(line)" />
                      </td>
                      <td class="px-3 py-2">
                        <input v-model.number="line.unit_cost" type="number" min="0" step="0.0001"
                          class="w-24 text-xs bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded px-2 py-1.5 text-right text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-indigo-500 ml-auto block" />
                      </td>
                      <td class="px-3 py-2 text-right">
                        <button @click="removeLine(idx)" class="p-1 rounded text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors">
                          <Trash2 :size="12" />
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <div v-if="createForm.lines.length > 0" class="mt-2 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400 px-1">
                <span>{{ createForm.lines.length }} line(s)</span>
                <span>Variance lines: {{ varianceLines }}</span>
              </div>
            </div>
          </div>
          <div class="border-t border-gray-200 dark:border-gray-800 px-6 py-4 flex items-center justify-end gap-3">
            <button @click="showCreate = false" class="text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white px-4 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
              Cancel
            </button>
            <button @click="saveCount" :disabled="saving"
              class="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 text-white text-sm font-medium px-5 py-2 rounded-lg transition-colors">
              <Loader2 v-if="saving" :size="14" class="animate-spin" />
              <Save v-else :size="14" />
              Create Count
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Detail Drawer -->
    <Teleport to="body">
      <div v-if="detailCount" class="fixed inset-0 z-50 flex">
        <div class="flex-1 bg-black/40 backdrop-blur-sm" @click="detailCount = null"></div>
        <div class="w-full max-w-2xl bg-white dark:bg-gray-900 shadow-2xl overflow-y-auto">
          <div class="sticky top-0 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 px-6 py-4 flex items-center justify-between z-10">
            <div>
              <h2 class="text-lg font-bold text-gray-900 dark:text-white font-mono">{{ detailCount.number }}</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">{{ detailCount.warehouse_name }} — {{ detailCount.date }}</p>
            </div>
            <div class="flex items-center gap-2">
              <span :class="statusBadge(detailCount.status)" class="inline-flex items-center gap-1 px-3 py-1 rounded-lg text-sm font-medium">
                <component :is="statusIcon(detailCount.status)" :size="12" />
                {{ statusLabel(detailCount.status) }}
              </span>
              <button v-if="detailCount.status === 'draft'" @click="validate(detailCount)"
                class="inline-flex items-center gap-1.5 bg-emerald-600 hover:bg-emerald-700 text-white text-sm font-medium px-3 py-1.5 rounded-lg transition-colors">
                <CheckCircle :size="13" />
                Validate
              </button>
              <button @click="detailCount = null" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 transition-colors">
                <X :size="18" />
              </button>
            </div>
          </div>

          <div class="p-6 space-y-6">
            <!-- Summary -->
            <div class="grid grid-cols-3 gap-4">
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3 text-center">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Total Lines</p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">{{ detailCount.lines?.length || 0 }}</p>
              </div>
              <div class="bg-amber-50 dark:bg-amber-900/20 rounded-lg p-3 text-center">
                <p class="text-xs text-amber-600 dark:text-amber-400 mb-1">Variance Lines</p>
                <p class="text-xl font-bold text-amber-700 dark:text-amber-300">{{ detailVarianceLines }}</p>
              </div>
              <div class="rounded-lg p-3 text-center" :class="detailTotalVariance > 0 ? 'bg-emerald-50 dark:bg-emerald-900/20' : detailTotalVariance < 0 ? 'bg-red-50 dark:bg-red-900/20' : 'bg-gray-50 dark:bg-gray-800'">
                <p class="text-xs mb-1" :class="detailTotalVariance > 0 ? 'text-emerald-600 dark:text-emerald-400' : detailTotalVariance < 0 ? 'text-red-600 dark:text-red-400' : 'text-gray-500 dark:text-gray-400'">Total Variance</p>
                <p class="text-xl font-bold" :class="detailTotalVariance > 0 ? 'text-emerald-700 dark:text-emerald-300' : detailTotalVariance < 0 ? 'text-red-700 dark:text-red-300' : 'text-gray-900 dark:text-white'">
                  {{ detailTotalVariance > 0 ? '+' : '' }}{{ fmtQty(detailTotalVariance) }}
                </p>
              </div>
            </div>

            <!-- Lines Table -->
            <div v-if="detailCount.lines && detailCount.lines.length > 0">
              <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">Count Lines</h3>
              <div class="rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
                <table class="w-full text-sm">
                  <thead class="bg-gray-50 dark:bg-gray-800">
                    <tr>
                      <th class="px-3 py-2 text-left text-xs font-semibold text-gray-600 dark:text-gray-300">Item</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">Book Qty</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">Counted Qty</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">Difference</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">Unit Cost</th>
                      <th class="px-3 py-2 text-right text-xs font-semibold text-gray-600 dark:text-gray-300">Variance Value</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
                    <tr v-for="line in detailCount.lines" :key="line.id"
                      :class="line.difference !== 0 ? 'bg-amber-50/40 dark:bg-amber-900/10' : ''">
                      <td class="px-3 py-2">
                        <div class="font-mono text-xs text-indigo-600 dark:text-indigo-400 font-semibold">{{ line.item_code }}</div>
                        <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ line.item_name }}</div>
                      </td>
                      <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300 text-xs">{{ fmtQty(line.book_qty) }}</td>
                      <td class="px-3 py-2 text-right text-xs">
                        <span v-if="line.counted_qty !== null && line.counted_qty !== undefined" class="font-medium text-gray-900 dark:text-white">
                          {{ fmtQty(line.counted_qty) }}
                        </span>
                        <span v-else class="text-gray-400 dark:text-gray-500">—</span>
                      </td>
                      <td class="px-3 py-2 text-right text-xs font-semibold">
                        <span v-if="line.difference !== 0" :class="line.difference > 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-red-600 dark:text-red-400'">
                          {{ line.difference > 0 ? '+' : '' }}{{ fmtQty(line.difference) }}
                        </span>
                        <span v-else class="text-gray-400 dark:text-gray-500">0</span>
                      </td>
                      <td class="px-3 py-2 text-right text-gray-600 dark:text-gray-300 text-xs">{{ fmt(line.unit_cost) }}</td>
                      <td class="px-3 py-2 text-right text-xs font-medium">
                        <span :class="(line.difference * line.unit_cost) > 0 ? 'text-emerald-600 dark:text-emerald-400' : (line.difference * line.unit_cost) < 0 ? 'text-red-600 dark:text-red-400' : 'text-gray-400'">
                          {{ line.difference !== 0 ? (line.difference > 0 ? '+' : '') + fmt(line.difference * line.unit_cost) : '—' }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div v-if="detailCount.notes" class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3 text-sm text-gray-600 dark:text-gray-300">
              {{ detailCount.notes }}
            </div>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Plus, ClipboardList, FileEdit, CheckCircle, Calendar, Search, Loader2,
  X, Save, Eye, Trash2, ChevronLeft, ChevronRight, ChevronsUpDown, ChevronUp, ChevronDown,
  XCircle
} from 'lucide-vue-next'
import { inventoryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── Types ────────────────────────────────────────────────────────────────────

interface CountLine {
  id: string
  count_id: string
  item_id: string
  item_code: string
  item_name: string
  uom_code?: string
  location_id?: string
  book_qty: number
  counted_qty?: number | null
  difference: number
  unit_cost: number
}

interface InventoryCount {
  id: string
  company_id: string
  number: string
  date: string
  warehouse_id: string
  warehouse_name?: string
  status: string
  notes?: string
  validated_by?: string
  validated_at?: string
  created_by?: string
  created_at: string
  lines?: CountLine[]
}

interface SimpleItem { id: string; code: string; name: string; standard_cost: number; cmup_cost: number }
interface SimpleWarehouse { id: string; code: string; name: string }

// ─── State ────────────────────────────────────────────────────────────────────

const counts = ref<InventoryCount[]>([])
const items = ref<SimpleItem[]>([])
const warehouses = ref<SimpleWarehouse[]>([])
const loading = ref(false)
const saving = ref(false)
const validating = ref(false)

const search = ref('')
const filterStatus = ref('')
const filterWarehouse = ref('')

const sortKey = ref('date')
const sortDir = ref<'asc' | 'desc'>('desc')
const currentPage = ref(1)
const pageSize = 20

const showCreate = ref(false)
const detailCount = ref<InventoryCount | null>(null)
const detailLoading = ref(false)

interface CreateLine {
  item_id: string
  book_qty: number
  counted_qty: number | null
  unit_cost: number
}

const createForm = ref<{
  date: string
  warehouse_id: string
  notes: string
  lines: CreateLine[]
}>({
  date: new Date().toISOString().slice(0, 10),
  warehouse_id: '',
  notes: '',
  lines: [],
})

// ─── Columns ──────────────────────────────────────────────────────────────────

const columns = [
  { key: 'number', label: 'Number' },
  { key: 'date', label: 'Date' },
  { key: 'warehouse_name', label: 'Warehouse' },
  { key: 'status', label: 'Status' },
  { key: 'validated_at', label: 'Validated' },
  { key: 'created_at', label: 'Created' },
  { key: 'notes', label: 'Notes' },
]

// ─── Computed ─────────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = [...counts.value]
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(c =>
      c.number.toLowerCase().includes(q) ||
      (c.warehouse_name || '').toLowerCase().includes(q)
    )
  }
  if (filterStatus.value) list = list.filter(c => c.status === filterStatus.value)
  if (filterWarehouse.value) list = list.filter(c => c.warehouse_id === filterWarehouse.value)

  list.sort((a, b) => {
    const va = (a as Record<string, unknown>)[sortKey.value]
    const vb = (b as Record<string, unknown>)[sortKey.value]
    const cmp = String(va ?? '').localeCompare(String(vb ?? ''), undefined, { numeric: true })
    return sortDir.value === 'asc' ? cmp : -cmp
  })
  return list
})

const totalPages = computed(() => Math.max(1, Math.ceil(filtered.value.length / pageSize)))
const paginated = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filtered.value.slice(start, start + pageSize)
})

const countsByStatus = (s: string) => counts.value.filter(c => c.status === s).length

const thisMonthCount = computed(() => {
  const now = new Date()
  const ym = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  return counts.value.filter(c => c.date.startsWith(ym)).length
})

const varianceLines = computed(() =>
  createForm.value.lines.filter(l =>
    l.counted_qty !== null && l.counted_qty !== undefined && l.counted_qty !== l.book_qty
  ).length
)

const detailVarianceLines = computed(() =>
  detailCount.value?.lines?.filter(l => l.difference !== 0).length || 0
)

const detailTotalVariance = computed(() =>
  detailCount.value?.lines?.reduce((s, l) => s + (l.difference || 0), 0) || 0
)

// ─── Helpers ──────────────────────────────────────────────────────────────────

const fmt = (v: number) =>
  new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0)

const fmtQty = (v: number) =>
  new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 0, maximumFractionDigits: 4 }).format(v || 0)

const statusLabel = (s: string) => {
  if (s === 'draft') return 'Draft'
  if (s === 'validated') return 'Validated'
  return s
}

const statusBadge = (s: string) => {
  if (s === 'draft') return 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400'
  if (s === 'validated') return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400'
  return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
}

const statusIcon = (s: string) => {
  if (s === 'draft') return FileEdit
  if (s === 'validated') return CheckCircle
  return XCircle
}

const getDiffClass = (line: CreateLine) => {
  if (line.counted_qty === null || line.counted_qty === undefined) return 'text-gray-900 dark:text-white'
  if (line.counted_qty > line.book_qty) return 'text-emerald-700 dark:text-emerald-300'
  if (line.counted_qty < line.book_qty) return 'text-red-700 dark:text-red-300'
  return 'text-gray-900 dark:text-white'
}

const sortBy = (key: string) => {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'asc'
  }
  currentPage.value = 1
}

const sortIcon = (key: string) => {
  if (sortKey.value !== key) return ChevronsUpDown
  return sortDir.value === 'asc' ? ChevronUp : ChevronDown
}

// ─── Actions ──────────────────────────────────────────────────────────────────

const openCreate = () => {
  createForm.value = {
    date: new Date().toISOString().slice(0, 10),
    warehouse_id: '',
    notes: '',
    lines: [],
  }
  showCreate.value = true
}

const addLine = () => {
  createForm.value.lines.push({
    item_id: '', book_qty: 0, counted_qty: null, unit_cost: 0
  })
}

const removeLine = (idx: number) => {
  createForm.value.lines.splice(idx, 1)
}

const onItemChange = (line: CreateLine) => {
  const item = items.value.find(i => i.id === line.item_id)
  if (item) {
    line.unit_cost = item.cmup_cost || item.standard_cost || 0
  }
}

const saveCount = async () => {
  if (!createForm.value.warehouse_id) {
    app.addToast('Warehouse is required', 'error')
    return
  }
  saving.value = true
  try {
    await inventoryAPI.createInventoryCount({
      date: createForm.value.date,
      warehouse_id: createForm.value.warehouse_id,
      notes: createForm.value.notes || null,
      lines: createForm.value.lines.filter(l => l.item_id).map(l => ({
        item_id: l.item_id,
        book_qty: l.book_qty,
        counted_qty: l.counted_qty,
        unit_cost: l.unit_cost,
      })),
    })
    app.addToast('Inventory count created', 'success')
    showCreate.value = false
    await load()
  } catch {
    app.addToast('Failed to create inventory count', 'error')
  } finally {
    saving.value = false
  }
}

const openDetail = async (id: string) => {
  detailCount.value = null
  detailLoading.value = true
  try {
    const res = await inventoryAPI.getInventoryCount(id)
    detailCount.value = res.data
  } catch {
    app.addToast('Failed to load count details', 'error')
  } finally {
    detailLoading.value = false
  }
}

const validate = async (count: InventoryCount) => {
  if (!confirm(`Validate count ${count.number}? This will update stock levels based on counted quantities.`)) return
  validating.value = true
  try {
    await inventoryAPI.validateInventoryCount(count.id)
    app.addToast('Inventory count validated and stock updated', 'success')
    if (detailCount.value?.id === count.id) {
      await openDetail(count.id)
    }
    await load()
  } catch {
    app.addToast('Failed to validate count', 'error')
  } finally {
    validating.value = false
  }
}

// ─── Load ─────────────────────────────────────────────────────────────────────

const load = async () => {
  loading.value = true
  try {
    const [countsRes, itemsRes, whRes] = await Promise.all([
      inventoryAPI.getInventoryCounts(),
      inventoryAPI.getItems(),
      inventoryAPI.getWarehouses(),
    ])
    counts.value = countsRes.data || []
    items.value = itemsRes.data || []
    warehouses.value = whRes.data || []
    currentPage.value = 1
  } catch {
    app.addToast('Failed to load inventory counts', 'error')
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
