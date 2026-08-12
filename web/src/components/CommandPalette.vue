<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'

const app = useAppStore()
const router = useRouter()
const query = ref('')

interface CommandItem {
  label: string
  description?: string
  to?: string
  action?: () => void
  group: string
  icon: string
}

const allCommands: CommandItem[] = [
  { label: 'Dashboard', to: '/dashboard', group: 'Navigation', icon: '📊' },
  // Accounting
  { label: 'Chart of Accounts', to: '/accounting/chart-of-accounts', group: 'Accounting', icon: '📋' },
  { label: 'Journal Entries', to: '/accounting/journal-entries', group: 'Accounting', icon: '📖' },
  { label: 'Financial Reports', to: '/accounting/reports', group: 'Accounting', icon: '📈' },
  { label: 'Fixed Assets', to: '/accounting/fixed-assets', group: 'Accounting', icon: '🏢' },
  { label: 'Trial Balance', to: '/accounting/reports', group: 'Accounting', icon: '⚖️' },
  // Sales
  { label: 'Sales Pipeline', to: '/sales/pipeline', group: 'Sales', icon: '🔽' },
  { label: 'Customers', to: '/sales/customers', group: 'Sales', icon: '👥' },
  { label: 'Sales Invoices', to: '/sales/invoices', group: 'Sales', icon: '🧾' },
  { label: 'Customer Aging', to: '/sales/reports/aging', group: 'Sales', icon: '⏰' },
  // Purchase
  { label: 'Suppliers', to: '/purchase/suppliers', group: 'Purchase', icon: '🚚' },
  { label: 'Purchase Orders', to: '/purchase/orders', group: 'Purchase', icon: '🛒' },
  // Inventory
  { label: 'Products & Items', to: '/inventory/items', group: 'Inventory', icon: '🏷️' },
  { label: 'Stock Levels', to: '/inventory/stock-levels', group: 'Inventory', icon: '📊' },
  { label: 'Warehouses', to: '/inventory/warehouses', group: 'Inventory', icon: '🏭' },
  // HR
  { label: 'Employees', to: '/hr/employees', group: 'HR', icon: '🪪' },
  { label: 'Payroll', to: '/hr/payroll', group: 'HR', icon: '💳' },
  { label: 'Leave Requests', to: '/hr/leave-requests', group: 'HR', icon: '📅' },
  // Manufacturing
  { label: 'Bill of Materials', to: '/manufacturing/bom', group: 'Manufacturing', icon: '📐' },
  { label: 'Manufacturing Orders', to: '/manufacturing/orders', group: 'Manufacturing', icon: '⚙️' },
  { label: 'MRP Planning', to: '/manufacturing/mrp', group: 'Manufacturing', icon: '🗓️' },
  // Projects
  { label: 'Projects', to: '/projects', group: 'Projects', icon: '📁' },
  { label: 'Timesheets', to: '/projects/timesheets', group: 'Projects', icon: '🕒' },
  // Treasury
  { label: 'Cash Position', to: '/treasury/cash-position', group: 'Treasury', icon: '🏛️' },
  { label: 'Bank Accounts', to: '/treasury/bank-accounts', group: 'Treasury', icon: '💳' },
  { label: 'Cheques', to: '/treasury/cheques', group: 'Treasury', icon: '🧾' },
  // Tax
  { label: 'G50 Declaration', to: '/tax/g50', group: 'Tax', icon: '📋' },
  { label: 'VAT Register', to: '/tax/vat-register', group: 'Tax', icon: '📊' },
  // Settings
  { label: 'Settings', to: '/settings', group: 'System', icon: '⚙️' },
  { label: 'Users & Roles', to: '/settings/users', group: 'System', icon: '👥' },
  { label: 'Fiscal Years', to: '/settings/fiscal-years', group: 'System', icon: '📅' },
  { label: 'Audit Log', to: '/settings/audit-log', group: 'System', icon: '📜' },
]

const selectedIndex = ref(0)

const filtered = computed(() => {
  if (!query.value.trim()) return allCommands.slice(0, 10)
  const q = query.value.toLowerCase()
  return allCommands.filter(c =>
    c.label.toLowerCase().includes(q) || c.group.toLowerCase().includes(q)
  ).slice(0, 10)
})

watch(filtered, () => { selectedIndex.value = 0 })

function selectItem(item: CommandItem) {
  if (item.to) {
    router.push(item.to)
  } else if (item.action) {
    item.action()
  }
  app.commandPaletteOpen = false
  query.value = ''
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    selectedIndex.value = Math.min(selectedIndex.value + 1, filtered.value.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    if (filtered.value[selectedIndex.value]) selectItem(filtered.value[selectedIndex.value])
  }
}

const groupedCommands = computed(() => {
  const groups: Record<string, CommandItem[]> = {}
  for (const item of filtered.value) {
    if (!groups[item.group]) groups[item.group] = []
    groups[item.group].push(item)
  }
  return groups
})
</script>

<template>
  <teleport to="body">
    <transition name="palette">
      <div
        v-if="app.commandPaletteOpen"
        class="fixed inset-0 z-[9998] flex items-start justify-center pt-20 px-4"
        @click.self="app.commandPaletteOpen = false"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm"></div>

        <!-- Palette panel -->
        <div class="relative w-full max-w-lg bg-white rounded-2xl shadow-2xl overflow-hidden">
          <!-- Search input -->
          <div class="flex items-center gap-3 px-4 py-3 border-b border-slate-100">
            <span class="text-slate-400 text-lg">🔍</span>
            <input
              ref="searchInput"
              v-model="query"
              class="flex-1 text-slate-800 placeholder-slate-400 text-base outline-none bg-transparent"
              placeholder="Search or navigate to..."
              autofocus
              @keydown="handleKeydown"
            />
            <kbd class="text-xs bg-slate-100 text-slate-500 rounded px-1.5 py-0.5 font-mono">ESC</kbd>
          </div>

          <!-- Results -->
          <div class="max-h-80 overflow-y-auto py-2">
            <template v-if="filtered.length > 0">
              <template v-for="(items, group) in groupedCommands" :key="group">
                <div class="px-4 py-1 text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
                  {{ group }}
                </div>
                <button
                  v-for="(item, idx) in items"
                  :key="item.label"
                  class="w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-colors text-left"
                  :class="filtered.indexOf(item) === selectedIndex ? 'bg-indigo-50 text-indigo-800' : 'text-slate-700 hover:bg-slate-50'"
                  @click="selectItem(item)"
                  @mouseenter="selectedIndex = filtered.indexOf(item)"
                >
                  <span>{{ item.icon }}</span>
                  <span class="flex-1 font-medium">{{ item.label }}</span>
                  <span class="text-xs text-slate-400">{{ item.group }}</span>
                </button>
              </template>
            </template>
            <div v-else class="px-4 py-8 text-center text-sm text-slate-400">
              No results found for "{{ query }}"
            </div>
          </div>

          <!-- Footer shortcuts -->
          <div class="px-4 py-2 border-t border-slate-100 flex items-center gap-4 text-xs text-slate-400">
            <span><kbd class="bg-slate-100 px-1 rounded">↑↓</kbd> Navigate</span>
            <span><kbd class="bg-slate-100 px-1 rounded">↵</kbd> Select</span>
            <span><kbd class="bg-slate-100 px-1 rounded">ESC</kbd> Close</span>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<style scoped>
.palette-enter-active,
.palette-leave-active {
  transition: all 0.2s ease;
}
.palette-enter-from,
.palette-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>
