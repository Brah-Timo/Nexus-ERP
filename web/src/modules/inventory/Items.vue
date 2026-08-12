<template>
  <div class="flex flex-col h-full bg-gray-50 dark:bg-gray-950 min-h-screen">

    <!-- Header -->
    <div class="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 px-6 py-4">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Products &amp; Items</h1>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Manage your product catalog, costs, and stock settings</p>
        </div>
        <div class="flex items-center gap-3">
          <button @click="openCreate" class="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors">
            <Plus :size="16" />
            New Item
          </button>
        </div>
      </div>
    </div>

    <!-- KPI Cards -->
    <div class="px-6 py-4 grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Total Items</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ items.length }}</p>
          </div>
          <div class="w-10 h-10 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg flex items-center justify-center">
            <Package :size="20" class="text-indigo-600 dark:text-indigo-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Storable</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ countByType('storable') }}</p>
          </div>
          <div class="w-10 h-10 bg-emerald-100 dark:bg-emerald-900/40 rounded-lg flex items-center justify-center">
            <Boxes :size="20" class="text-emerald-600 dark:text-emerald-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Consumable</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ countByType('consumable') }}</p>
          </div>
          <div class="w-10 h-10 bg-amber-100 dark:bg-amber-900/40 rounded-lg flex items-center justify-center">
            <FlaskConical :size="20" class="text-amber-600 dark:text-amber-400" />
          </div>
        </div>
      </div>
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">Services</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ countByType('service') }}</p>
          </div>
          <div class="w-10 h-10 bg-blue-100 dark:bg-blue-900/40 rounded-lg flex items-center justify-center">
            <Wrench :size="20" class="text-blue-600 dark:text-blue-400" />
          </div>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="px-6 pb-4 flex flex-wrap gap-3">
      <div class="relative flex-1 min-w-64">
        <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
        <input v-model="search" placeholder="Search by code, name, barcode..." type="text"
          class="w-full pl-9 pr-4 py-2 text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
      </div>
      <select v-model="filterType"
        class="text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
        <option value="">All Types</option>
        <option value="storable">Storable</option>
        <option value="consumable">Consumable</option>
        <option value="service">Service</option>
      </select>
      <select v-model="filterCategory"
        class="text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
        <option value="">All Categories</option>
        <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
      </select>
      <select v-model="filterActive"
        class="text-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
        <option value="">All Status</option>
        <option value="true">Active</option>
        <option value="false">Inactive</option>
      </select>
    </div>

    <!-- Table -->
    <div class="flex-1 px-6 pb-6">
      <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
        <div v-if="loading" class="flex items-center justify-center py-16">
          <Loader2 :size="28" class="animate-spin text-indigo-500" />
          <span class="ml-3 text-gray-500 dark:text-gray-400">Loading items...</span>
        </div>
        <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center py-16">
          <PackageOpen :size="48" class="text-gray-300 dark:text-gray-600 mb-3" />
          <p class="text-gray-500 dark:text-gray-400 font-medium">No items found</p>
          <p class="text-sm text-gray-400 dark:text-gray-500 mt-1">Try adjusting filters or add a new item</p>
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th v-for="col in columns" :key="col.key"
                  class="px-4 py-3 text-left font-semibold text-gray-600 dark:text-gray-300 cursor-pointer select-none hover:text-gray-900 dark:hover:text-white transition-colors"
                  @click="sortBy(col.key)">
                  <div class="flex items-center gap-1">
                    {{ col.label }}
                    <component :is="sortIcon(col.key)" :size="12" class="text-gray-400" />
                  </div>
                </th>
                <th class="px-4 py-3 text-right font-semibold text-gray-600 dark:text-gray-300">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
              <tr v-for="item in paginated" :key="item.id"
                class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors cursor-pointer"
                @click="openDetail(item)">
                <td class="px-4 py-3">
                  <div class="font-mono text-xs font-semibold text-indigo-600 dark:text-indigo-400">{{ item.code }}</div>
                  <div v-if="item.internal_ref" class="text-xs text-gray-400 mt-0.5">Ref: {{ item.internal_ref }}</div>
                </td>
                <td class="px-4 py-3">
                  <div class="font-medium text-gray-900 dark:text-white">{{ item.name }}</div>
                  <div v-if="item.barcode" class="text-xs text-gray-400 mt-0.5">{{ item.barcode }}</div>
                </td>
                <td class="px-4 py-3">
                  <span :class="typeBadge(item.item_type)" class="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium">
                    {{ typeLabel(item.item_type) }}
                  </span>
                </td>
                <td class="px-4 py-3 text-gray-600 dark:text-gray-300">{{ item.category_name || '-' }}</td>
                <td class="px-4 py-3 text-gray-600 dark:text-gray-300">{{ item.uom_code || '-' }}</td>
                <td class="px-4 py-3 text-right">
                  <div class="font-medium text-gray-900 dark:text-white">{{ fmt(item.sale_price) }}</div>
                  <div class="text-xs text-gray-400 mt-0.5">Cost: {{ fmt(item.standard_cost) }}</div>
                </td>
                <td class="px-4 py-3 text-right">
                  <div class="text-gray-700 dark:text-gray-300">{{ fmt(item.cmup_cost) }}</div>
                </td>
                <td class="px-4 py-3">
                  <span :class="item.is_active ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'"
                    class="inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium">
                    {{ item.is_active ? 'Active' : 'Inactive' }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right" @click.stop>
                  <div class="flex items-center justify-end gap-1">
                    <button @click="openEdit(item)" class="p-1.5 rounded-lg text-gray-400 hover:text-indigo-600 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 transition-colors">
                      <Pencil :size="14" />
                    </button>
                    <button @click="deactivate(item)" class="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 transition-colors">
                      <Trash2 :size="14" />
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

    <!-- Detail Drawer -->
    <Teleport to="body">
      <div v-if="detailItem" class="fixed inset-0 z-50 flex">
        <div class="flex-1 bg-black/40 backdrop-blur-sm" @click="detailItem = null"></div>
        <div class="w-full max-w-lg bg-white dark:bg-gray-900 shadow-2xl overflow-y-auto">
          <div class="sticky top-0 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 px-6 py-4 flex items-center justify-between z-10">
            <div>
              <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ detailItem.name }}</h2>
              <p class="text-sm font-mono text-indigo-600 dark:text-indigo-400">{{ detailItem.code }}</p>
            </div>
            <div class="flex items-center gap-2">
              <button @click="openEdit(detailItem)" class="inline-flex items-center gap-1.5 text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 font-medium">
                <Pencil :size="14" /> Edit
              </button>
              <button @click="detailItem = null" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 transition-colors">
                <X :size="18" />
              </button>
            </div>
          </div>
          <div class="p-6 space-y-6">
            <div class="flex flex-wrap gap-2">
              <span :class="typeBadge(detailItem.item_type)" class="inline-flex items-center px-3 py-1 rounded-lg text-sm font-medium">
                {{ typeLabel(detailItem.item_type) }}
              </span>
              <span :class="detailItem.is_active ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'"
                class="inline-flex items-center px-3 py-1 rounded-lg text-sm font-medium">
                {{ detailItem.is_active ? 'Active' : 'Inactive' }}
              </span>
            </div>

            <div v-if="detailItem.description" class="text-sm text-gray-600 dark:text-gray-300 bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
              {{ detailItem.description }}
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Category</p>
                <p class="font-medium text-gray-900 dark:text-white">{{ detailItem.category_name || 'Uncategorized' }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Unit of Measure</p>
                <p class="font-medium text-gray-900 dark:text-white">{{ detailItem.uom_code || '-' }} <span v-if="detailItem.uom_name" class="text-gray-500 text-xs">({{ detailItem.uom_name }})</span></p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Sale Price</p>
                <p class="font-semibold text-gray-900 dark:text-white">{{ fmt(detailItem.sale_price) }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Standard Cost</p>
                <p class="font-semibold text-gray-900 dark:text-white">{{ fmt(detailItem.standard_cost) }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">CMUP Cost</p>
                <p class="font-semibold text-gray-900 dark:text-white">{{ fmt(detailItem.cmup_cost) }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">TVA Rate</p>
                <p class="font-semibold text-gray-900 dark:text-white">{{ detailItem.tva_rate }}%</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Min Stock</p>
                <p class="font-medium text-gray-900 dark:text-white">{{ detailItem.min_stock_qty }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Max Stock</p>
                <p class="font-medium text-gray-900 dark:text-white">{{ detailItem.max_stock_qty }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Reorder Qty</p>
                <p class="font-medium text-gray-900 dark:text-white">{{ detailItem.reorder_qty }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-3">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Cost Method</p>
                <p class="font-medium text-gray-900 dark:text-white uppercase">{{ detailItem.cost_method }}</p>
              </div>
            </div>

            <div class="space-y-1 text-sm">
              <div v-if="detailItem.barcode" class="flex justify-between py-1 border-b border-gray-100 dark:border-gray-800">
                <span class="text-gray-500 dark:text-gray-400">Barcode</span>
                <span class="font-mono text-gray-900 dark:text-white">{{ detailItem.barcode }}</span>
              </div>
              <div v-if="detailItem.internal_ref" class="flex justify-between py-1 border-b border-gray-100 dark:border-gray-800">
                <span class="text-gray-500 dark:text-gray-400">Internal Ref</span>
                <span class="font-mono text-gray-900 dark:text-white">{{ detailItem.internal_ref }}</span>
              </div>
              <div v-if="detailItem.hs_code" class="flex justify-between py-1 border-b border-gray-100 dark:border-gray-800">
                <span class="text-gray-500 dark:text-gray-400">HS Code</span>
                <span class="font-mono text-gray-900 dark:text-white">{{ detailItem.hs_code }}</span>
              </div>
              <div class="flex justify-between py-1 border-b border-gray-100 dark:border-gray-800">
                <span class="text-gray-500 dark:text-gray-400">Track Inventory</span>
                <span class="text-gray-900 dark:text-white">{{ detailItem.track_inventory ? 'Yes' : 'No' }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Create / Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="closeModal"></div>
        <div class="relative bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] flex flex-col">
          <!-- Modal Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800">
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ editingItem ? 'Edit Item' : 'New Item' }}</h2>
            <button @click="closeModal" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 text-gray-400 transition-colors">
              <X :size="18" />
            </button>
          </div>

          <!-- Modal Body -->
          <div class="flex-1 overflow-y-auto px-6 py-4 space-y-5">
            <!-- Basic Info -->
            <div>
              <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">Basic Information</h3>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Code <span class="text-red-500">*</span></label>
                  <input v-model="form.code" type="text" placeholder="ITEM-001"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Item Type</label>
                  <select v-model="form.item_type"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                    <option value="storable">Storable</option>
                    <option value="consumable">Consumable</option>
                    <option value="service">Service</option>
                  </select>
                </div>
                <div class="col-span-2">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Name <span class="text-red-500">*</span></label>
                  <input v-model="form.name" type="text" placeholder="Product name"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div class="col-span-2">
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
                  <textarea v-model="form.description" rows="2" placeholder="Optional description"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"></textarea>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Category</label>
                  <select v-model="form.category_id"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                    <option value="">No category</option>
                    <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
                  </select>
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Unit of Measure</label>
                  <select v-model="form.uom_id"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                    <option value="">No unit</option>
                    <option v-for="u in units" :key="u.id" :value="u.id">{{ u.code }} – {{ u.name }}</option>
                  </select>
                </div>
              </div>
            </div>

            <!-- Pricing & Costs -->
            <div>
              <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">Pricing &amp; Costs</h3>
              <div class="grid grid-cols-3 gap-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Sale Price</label>
                  <input v-model.number="form.sale_price" type="number" min="0" step="0.01"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Standard Cost</label>
                  <input v-model.number="form.standard_cost" type="number" min="0" step="0.01"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">TVA Rate (%)</label>
                  <input v-model.number="form.tva_rate" type="number" min="0" max="100" step="0.01"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Cost Method</label>
                  <select v-model="form.cost_method"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500">
                    <option value="cmup">CMUP</option>
                    <option value="fifo">FIFO</option>
                    <option value="lifo">LIFO</option>
                  </select>
                </div>
              </div>
            </div>

            <!-- Stock Settings -->
            <div>
              <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">Stock Settings</h3>
              <div class="grid grid-cols-3 gap-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Min Stock Qty</label>
                  <input v-model.number="form.min_stock_qty" type="number" min="0" step="0.01"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Max Stock Qty</label>
                  <input v-model.number="form.max_stock_qty" type="number" min="0" step="0.01"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Reorder Qty</label>
                  <input v-model.number="form.reorder_qty" type="number" min="0" step="0.01"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
              </div>
            </div>

            <!-- Extra Identifiers -->
            <div>
              <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-3">Identifiers</h3>
              <div class="grid grid-cols-3 gap-4">
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Barcode</label>
                  <input v-model="form.barcode" type="text"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Internal Ref</label>
                  <input v-model="form.internal_ref" type="text"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">HS Code</label>
                  <input v-model="form.hs_code" type="text"
                    class="w-full text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-2 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
              </div>
              <div class="mt-3 flex items-center gap-3">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" v-model="form.track_inventory"
                    class="w-4 h-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" />
                  <span class="text-sm text-gray-700 dark:text-gray-300">Track Inventory</span>
                </label>
                <label v-if="editingItem" class="flex items-center gap-2 cursor-pointer">
                  <input type="checkbox" v-model="form.is_active"
                    class="w-4 h-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500" />
                  <span class="text-sm text-gray-700 dark:text-gray-300">Active</span>
                </label>
              </div>
            </div>
          </div>

          <!-- Modal Footer -->
          <div class="border-t border-gray-200 dark:border-gray-800 px-6 py-4 flex items-center justify-end gap-3">
            <button @click="closeModal" class="text-sm font-medium text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white px-4 py-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">
              Cancel
            </button>
            <button @click="save" :disabled="saving"
              class="inline-flex items-center gap-2 bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 text-white text-sm font-medium px-5 py-2 rounded-lg transition-colors">
              <Loader2 v-if="saving" :size="14" class="animate-spin" />
              <Save v-else :size="14" />
              {{ editingItem ? 'Update Item' : 'Create Item' }}
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
  Plus, Package, Boxes, FlaskConical, Wrench, Search, Loader2, PackageOpen,
  Pencil, Trash2, X, Save, ChevronLeft, ChevronRight, ChevronsUpDown, ChevronUp, ChevronDown
} from 'lucide-vue-next'
import { inventoryAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── State ────────────────────────────────────────────────────────────────────

interface Item {
  id: string
  code: string
  name: string
  description?: string
  category_id?: string
  category_name?: string
  uom_id?: string
  uom_code?: string
  uom_name?: string
  item_type: string
  track_inventory: boolean
  tva_rate: number
  cost_method: string
  standard_cost: number
  cmup_cost: number
  sale_price: number
  min_stock_qty: number
  reorder_qty: number
  max_stock_qty: number
  barcode?: string
  internal_ref?: string
  hs_code?: string
  is_active: boolean
  created_at?: string
  updated_at?: string
}

interface Category { id: string; name: string; code: string }
interface Unit { id: string; code: string; name: string; category?: string; factor: number }

const items = ref<Item[]>([])
const categories = ref<Category[]>([])
const units = ref<Unit[]>([])
const loading = ref(false)
const saving = ref(false)

const search = ref('')
const filterType = ref('')
const filterCategory = ref('')
const filterActive = ref('')

const sortKey = ref('code')
const sortDir = ref<'asc' | 'desc'>('asc')
const currentPage = ref(1)
const pageSize = 20

const showModal = ref(false)
const editingItem = ref<Item | null>(null)
const detailItem = ref<Item | null>(null)

const defaultForm = () => ({
  code: '',
  name: '',
  description: '',
  category_id: '',
  uom_id: '',
  item_type: 'storable',
  track_inventory: true,
  tva_rate: 19,
  cost_method: 'cmup',
  standard_cost: 0,
  sale_price: 0,
  min_stock_qty: 0,
  reorder_qty: 0,
  max_stock_qty: 0,
  barcode: '',
  internal_ref: '',
  hs_code: '',
  is_active: true,
})

const form = ref(defaultForm())

// ─── Columns ──────────────────────────────────────────────────────────────────

const columns = [
  { key: 'code', label: 'Code' },
  { key: 'name', label: 'Name' },
  { key: 'item_type', label: 'Type' },
  { key: 'category_name', label: 'Category' },
  { key: 'uom_code', label: 'UOM' },
  { key: 'sale_price', label: 'Price / Cost' },
  { key: 'cmup_cost', label: 'CMUP' },
  { key: 'is_active', label: 'Status' },
]

// ─── Computed ─────────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = [...items.value]
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(i =>
      i.code.toLowerCase().includes(q) ||
      i.name.toLowerCase().includes(q) ||
      (i.barcode || '').toLowerCase().includes(q) ||
      (i.internal_ref || '').toLowerCase().includes(q)
    )
  }
  if (filterType.value) list = list.filter(i => i.item_type === filterType.value)
  if (filterCategory.value) list = list.filter(i => i.category_id === filterCategory.value)
  if (filterActive.value !== '') {
    const isActive = filterActive.value === 'true'
    list = list.filter(i => i.is_active === isActive)
  }

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

// ─── Helpers ──────────────────────────────────────────────────────────────────

const fmt = (v: number) =>
  new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v || 0)

const countByType = (t: string) => items.value.filter(i => i.item_type === t && i.is_active).length

const typeLabel = (t: string) => {
  if (t === 'storable') return 'Storable'
  if (t === 'consumable') return 'Consumable'
  if (t === 'service') return 'Service'
  return t
}

const typeBadge = (t: string) => {
  if (t === 'storable') return 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400'
  if (t === 'consumable') return 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400'
  if (t === 'service') return 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400'
  return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
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

// ─── Load Data ────────────────────────────────────────────────────────────────

const load = async () => {
  loading.value = true
  try {
    const [itemsRes, catsRes, unitsRes] = await Promise.all([
      inventoryAPI.getItems(),
      inventoryAPI.getCategories(),
      inventoryAPI.getUnits(),
    ])
    items.value = itemsRes.data || []
    categories.value = catsRes.data || []
    units.value = unitsRes.data || []
  } catch {
    app.addToast('Failed to load inventory items', 'error')
  } finally {
    loading.value = false
  }
}

// ─── Actions ──────────────────────────────────────────────────────────────────

const openCreate = () => {
  editingItem.value = null
  form.value = defaultForm()
  showModal.value = true
}

const openEdit = (item: Item) => {
  detailItem.value = null
  editingItem.value = item
  form.value = {
    code: item.code,
    name: item.name,
    description: item.description || '',
    category_id: item.category_id || '',
    uom_id: item.uom_id || '',
    item_type: item.item_type,
    track_inventory: item.track_inventory,
    tva_rate: item.tva_rate,
    cost_method: item.cost_method,
    standard_cost: item.standard_cost,
    sale_price: item.sale_price,
    min_stock_qty: item.min_stock_qty,
    reorder_qty: item.reorder_qty,
    max_stock_qty: item.max_stock_qty,
    barcode: item.barcode || '',
    internal_ref: item.internal_ref || '',
    hs_code: item.hs_code || '',
    is_active: item.is_active,
  }
  showModal.value = true
}

const openDetail = (item: Item) => {
  detailItem.value = item
}

const closeModal = () => {
  showModal.value = false
  editingItem.value = null
}

const save = async () => {
  if (!form.value.code || !form.value.name) {
    app.addToast('Code and Name are required', 'error')
    return
  }
  saving.value = true
  try {
    const payload = {
      ...form.value,
      category_id: form.value.category_id || null,
      uom_id: form.value.uom_id || null,
      description: form.value.description || null,
      barcode: form.value.barcode || null,
      internal_ref: form.value.internal_ref || null,
      hs_code: form.value.hs_code || null,
    }
    if (editingItem.value) {
      await inventoryAPI.updateItem(editingItem.value.id, payload)
      app.addToast('Item updated successfully', 'success')
    } else {
      await inventoryAPI.createItem(payload)
      app.addToast('Item created successfully', 'success')
    }
    closeModal()
    await load()
  } catch {
    app.addToast('Failed to save item', 'error')
  } finally {
    saving.value = false
  }
}

const deactivate = async (item: Item) => {
  if (!confirm(`Deactivate item "${item.name}"?`)) return
  try {
    await inventoryAPI.deleteItem(item.id)
    app.addToast('Item deactivated', 'success')
    await load()
  } catch {
    app.addToast('Failed to deactivate item', 'error')
  }
}

// ─── Lifecycle ────────────────────────────────────────────────────────────────

onMounted(load)
</script>
