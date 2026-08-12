<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { settingsAPI } from '@/api/client'
import {
  ScrollText, Search, RefreshCw, Download, ChevronLeft, ChevronRight,
  Filter, X, Eye, User, Clock, Globe, Monitor, AlertCircle
} from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(false)
const logs = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const limit = ref(50)
const search = ref('')
const actionFilter = ref('')
const entityFilter = ref('')
const showDetail = ref<any>(null)
const showFilters = ref(false)

const actions = [
  'create', 'update', 'delete', 'login', 'logout', 'approve',
  'reject', 'post', 'cancel', 'confirm', 'close', 'print', 'export'
]
const entities = [
  'company', 'user', 'role', 'fiscal_year', 'currency',
  'invoice', 'payment', 'journal_entry', 'purchase_order',
  'product', 'employee', 'payroll', 'project', 'asset', 'vehicle'
]

const totalPages = computed(() => Math.ceil(total.value / limit.value))
const from = computed(() => ((page.value - 1) * limit.value) + 1)
const to = computed(() => Math.min(page.value * limit.value, total.value))

async function load() {
  loading.value = true
  try {
    const params: Record<string, any> = { page: page.value, limit: limit.value }
    if (actionFilter.value) params.action = actionFilter.value
    if (entityFilter.value) params.entity_type = entityFilter.value
    if (search.value) params.search = search.value
    const r = await settingsAPI.getAuditLog(params)
    const d = r.data
    logs.value = Array.isArray(d.data) ? d.data : []
    total.value = d.total || 0
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error loading audit log', 'error')
  } finally {
    loading.value = false
  }
}

function clearFilters() {
  actionFilter.value = ''
  entityFilter.value = ''
  search.value = ''
  page.value = 1
}

watch([actionFilter, entityFilter], () => { page.value = 1; load() })

let searchTimer: ReturnType<typeof setTimeout>
watch(search, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => { page.value = 1; load() }, 400)
})

function prevPage() { if (page.value > 1) { page.value--; load() } }
function nextPage() { if (page.value < totalPages.value) { page.value++; load() } }

function exportCSV() {
  const rows = [
    ['ID', 'User', 'Action', 'Entity Type', 'Entity ID', 'IP Address', 'Date'],
    ...logs.value.map(l => [l.id, l.username || l.user_id, l.action, l.entity_type, l.entity_id, l.ip_address, l.created_at])
  ]
  const a = document.createElement('a')
  a.href = URL.createObjectURL(new Blob(['\ufeff' + rows.map(r => r.join(',')).join('\n')], { type: 'text/csv' }))
  a.download = `audit_log_${new Date().toISOString().split('T')[0]}.csv`
  a.click()
}

function fmtDate(d: string) {
  if (!d) return '-'
  return new Date(d).toLocaleString('fr-DZ', { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function actionColor(action: string) {
  const map: Record<string, string> = {
    create: 'bg-emerald-500/10 text-emerald-400',
    update: 'bg-sky-500/10 text-sky-400',
    delete: 'bg-red-500/10 text-red-400',
    login: 'bg-violet-500/10 text-violet-400',
    logout: 'bg-slate-500/10 text-slate-400',
    approve: 'bg-amber-500/10 text-amber-400',
    reject: 'bg-red-500/10 text-red-400',
    post: 'bg-indigo-500/10 text-indigo-400',
    cancel: 'bg-orange-500/10 text-orange-400',
    close: 'bg-slate-500/10 text-slate-400',
    export: 'bg-teal-500/10 text-teal-400',
  }
  return map[action?.toLowerCase()] || 'bg-slate-500/10 text-slate-400'
}

function initials(name: string) {
  return name?.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2) || '?'
}

function pageNumbers() {
  const pages: (number | '...')[] = []
  const tp = totalPages.value
  const cp = page.value
  if (tp <= 7) {
    for (let i = 1; i <= tp; i++) pages.push(i)
  } else {
    pages.push(1)
    if (cp > 3) pages.push('...')
    for (let i = Math.max(2, cp - 1); i <= Math.min(tp - 1, cp + 1); i++) pages.push(i)
    if (cp < tp - 2) pages.push('...')
    pages.push(tp)
  }
  return pages
}

const activeFiltersCount = computed(() =>
  [actionFilter.value, entityFilter.value, search.value].filter(Boolean).length
)

const dk = (d: string, l: string) => app.darkMode ? d : l
onMounted(load)
</script>

<template>
  <div :class="dk('bg-slate-950 text-slate-100','bg-slate-50 text-slate-900')" class="min-h-screen">
    <!-- Header -->
    <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border-b px-6 py-4 sticky top-0 z-10">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-slate-600/20">
            <ScrollText class="w-5 h-5 text-slate-400" />
          </div>
          <div>
            <h1 class="text-lg font-bold">Audit Log</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">Complete record of all system activities</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="exportCSV" :disabled="logs.length===0"
            :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
            class="flex items-center gap-1.5 px-3 py-2 rounded-lg text-sm transition-colors disabled:opacity-50">
            <Download class="w-4 h-4" /> Export
          </button>
          <button @click="load" :disabled="loading"
            :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
            class="p-2 rounded-lg transition-colors disabled:opacity-50">
            <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- Search + Filters -->
      <div class="mt-3 flex items-center gap-3 flex-wrap">
        <div class="relative flex-1 min-w-48">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
          <input v-model="search" placeholder="Search by action, entity..."
            :class="dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500','bg-slate-100 border-slate-200 text-slate-900 placeholder-slate-400')"
            class="w-full pl-9 pr-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-slate-500" />
        </div>
        <button @click="showFilters=!showFilters"
          :class="[dk('bg-slate-800 text-slate-300 border-slate-700','bg-white text-slate-700 border-slate-200'), activeFiltersCount > 0 ? 'border-indigo-500 text-indigo-400' : '']"
          class="flex items-center gap-2 px-3 py-2 border rounded-lg text-sm transition-colors">
          <Filter class="w-4 h-4" />
          Filters
          <span v-if="activeFiltersCount > 0" class="w-4 h-4 bg-indigo-500 text-white rounded-full text-xs flex items-center justify-center">
            {{ activeFiltersCount }}
          </span>
        </button>
        <button v-if="activeFiltersCount > 0" @click="clearFilters"
          :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-900')"
          class="flex items-center gap-1 text-xs transition-colors">
          <X class="w-3.5 h-3.5" /> Clear
        </button>
      </div>

      <!-- Filter dropdowns -->
      <Transition name="expand">
        <div v-if="showFilters" class="mt-3 flex gap-3 flex-wrap">
          <select v-model="actionFilter"
            :class="dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')"
            class="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
            <option value="">All Actions</option>
            <option v-for="a in actions" :key="a" :value="a" class="capitalize">{{ a }}</option>
          </select>
          <select v-model="entityFilter"
            :class="dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')"
            class="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
            <option value="">All Entities</option>
            <option v-for="e in entities" :key="e" :value="e" class="capitalize">{{ e }}</option>
          </select>
          <select v-model.number="limit" @change="page=1; load()"
            :class="dk('bg-slate-800 border-slate-700 text-slate-100','bg-white border-slate-200 text-slate-900')"
            class="border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
            <option :value="25">25 per page</option>
            <option :value="50">50 per page</option>
            <option :value="100">100 per page</option>
            <option :value="200">200 per page</option>
          </select>
        </div>
      </Transition>
    </div>

    <!-- Content -->
    <div class="p-6">
      <div v-if="loading && logs.length===0" class="flex items-center justify-center py-20">
        <RefreshCw class="animate-spin w-8 h-8 text-slate-400" />
      </div>

      <template v-else>
        <!-- Total count bar -->
        <div class="flex items-center justify-between mb-4">
          <div :class="dk('text-slate-400','text-slate-500')" class="text-sm">
            <span v-if="total > 0">Showing {{ from }}–{{ to }} of {{ total.toLocaleString() }} entries</span>
            <span v-else>No entries found</span>
          </div>
          <div v-if="loading" class="flex items-center gap-1.5 text-xs text-slate-400">
            <RefreshCw class="animate-spin w-3.5 h-3.5" /> Loading...
          </div>
        </div>

        <div v-if="logs.length === 0" :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
          class="border rounded-xl p-16 text-center">
          <ScrollText :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12 mx-auto mb-4" />
          <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">No audit log entries found</p>
        </div>

        <div v-else :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
          class="border rounded-xl overflow-hidden">
          <!-- Table -->
          <div class="overflow-x-auto">
            <table class="w-full text-sm min-w-[700px]">
              <thead>
                <tr :class="dk('bg-slate-800/50 text-slate-400 border-slate-800','bg-slate-50 text-slate-500 border-slate-100')"
                  class="border-b text-xs uppercase tracking-wider">
                  <th class="px-4 py-3 text-left font-medium">User</th>
                  <th class="px-4 py-3 text-left font-medium">Action</th>
                  <th class="px-4 py-3 text-left font-medium">Entity Type</th>
                  <th class="px-4 py-3 text-left font-medium">Entity ID</th>
                  <th class="px-4 py-3 text-left font-medium">IP Address</th>
                  <th class="px-4 py-3 text-left font-medium">Date &amp; Time</th>
                  <th class="px-4 py-3 text-right font-medium">Detail</th>
                </tr>
              </thead>
              <tbody class="divide-y" :class="dk('divide-slate-800','divide-slate-100')">
                <tr v-for="log in logs" :key="log.id"
                  :class="dk('hover:bg-slate-800/40','hover:bg-slate-50')" class="transition-colors">
                  <!-- User -->
                  <td class="px-4 py-3">
                    <div class="flex items-center gap-2.5">
                      <div :class="dk('bg-slate-700','bg-slate-200')"
                        class="w-7 h-7 rounded-full flex items-center justify-center flex-shrink-0">
                        <span :class="dk('text-slate-300','text-slate-600')" class="text-xs font-medium">
                          {{ log.username ? initials(log.full_name || log.username) : '?' }}
                        </span>
                      </div>
                      <div>
                        <div class="text-xs font-medium">{{ log.full_name || log.username || 'System' }}</div>
                        <div v-if="log.username" :class="dk('text-slate-500','text-slate-400')" class="text-xs font-mono">@{{ log.username }}</div>
                      </div>
                    </div>
                  </td>
                  <!-- Action -->
                  <td class="px-4 py-3">
                    <span :class="actionColor(log.action)" class="text-xs px-2 py-0.5 rounded-full font-medium capitalize">
                      {{ log.action }}
                    </span>
                  </td>
                  <!-- Entity type -->
                  <td class="px-4 py-3">
                    <span :class="dk('text-slate-300 bg-slate-800','text-slate-600 bg-slate-100')"
                      class="text-xs px-2 py-0.5 rounded font-mono">{{ log.entity_type }}</span>
                  </td>
                  <!-- Entity ID -->
                  <td class="px-4 py-3">
                    <span :class="dk('text-slate-500','text-slate-400')" class="text-xs font-mono truncate max-w-[100px] block">
                      {{ log.entity_id ? log.entity_id.slice(0, 8) + '...' : '-' }}
                    </span>
                  </td>
                  <!-- IP -->
                  <td class="px-4 py-3">
                    <div class="flex items-center gap-1 text-xs" :class="dk('text-slate-400','text-slate-500')">
                      <Globe class="w-3 h-3 flex-shrink-0" />
                      <span class="font-mono">{{ log.ip_address || '-' }}</span>
                    </div>
                  </td>
                  <!-- Date -->
                  <td class="px-4 py-3">
                    <div class="flex items-center gap-1 text-xs" :class="dk('text-slate-400','text-slate-500')">
                      <Clock class="w-3 h-3 flex-shrink-0" />
                      <span>{{ fmtDate(log.created_at) }}</span>
                    </div>
                  </td>
                  <!-- Detail -->
                  <td class="px-4 py-3 text-right">
                    <button @click="showDetail=log"
                      :class="dk('hover:bg-slate-700 text-slate-400 hover:text-white','hover:bg-slate-100 text-slate-500 hover:text-slate-900')"
                      class="p-1.5 rounded-lg transition-colors">
                      <Eye class="w-3.5 h-3.5" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Pagination -->
          <div :class="dk('border-slate-800 bg-slate-900/50','border-slate-100 bg-slate-50')"
            class="border-t px-5 py-3 flex items-center justify-between gap-4 flex-wrap">
            <div :class="dk('text-slate-400','text-slate-500')" class="text-xs">
              Page {{ page }} of {{ totalPages || 1 }}
            </div>
            <div class="flex items-center gap-1">
              <button @click="prevPage" :disabled="page<=1"
                :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-30','bg-white hover:bg-slate-100 text-slate-700 border border-slate-200 disabled:opacity-30')"
                class="p-1.5 rounded-lg transition-colors">
                <ChevronLeft class="w-4 h-4" />
              </button>
              <template v-for="pg in pageNumbers()" :key="pg">
                <button v-if="pg!=='...'" @click="page=Number(pg); load()"
                  :class="page===Number(pg)
                    ? 'bg-indigo-600 text-white'
                    : dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-white hover:bg-slate-100 text-slate-700 border border-slate-200')"
                  class="w-8 h-8 rounded-lg text-xs font-medium transition-colors">{{ pg }}</button>
                <span v-else :class="dk('text-slate-600','text-slate-400')" class="px-1 text-xs">...</span>
              </template>
              <button @click="nextPage" :disabled="page>=totalPages"
                :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-30','bg-white hover:bg-slate-100 text-slate-700 border border-slate-200 disabled:opacity-30')"
                class="p-1.5 rounded-lg transition-colors">
                <ChevronRight class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Detail modal -->
    <Teleport to="body">
      <div v-if="showDetail" class="fixed inset-0 z-50 flex items-center justify-center p-4"
        style="background:rgba(0,0,0,0.7)" @click.self="showDetail=null">
        <div :class="dk('bg-slate-900 border-slate-700','bg-white border-slate-200')"
          class="border rounded-2xl w-full max-w-lg max-h-[85vh] overflow-hidden flex flex-col shadow-2xl">
          <div :class="dk('border-slate-800','border-slate-100')" class="flex items-center justify-between px-6 py-4 border-b flex-shrink-0">
            <h3 class="font-bold flex items-center gap-2">
              <ScrollText class="w-4 h-4 text-slate-400" /> Log Entry Detail
            </h3>
            <button @click="showDetail=null" :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-900')"
              class="p-1.5 rounded-lg transition-colors">
              <X class="w-4 h-4" />
            </button>
          </div>
          <div class="overflow-y-auto flex-1 px-6 py-5 space-y-3">
            <div v-for="[key, val] in Object.entries(showDetail).filter(([k,v]) => v !== null && v !== '')" :key="key"
              :class="dk('border-slate-800','border-slate-100')" class="border-b pb-3 last:border-0">
              <div :class="dk('text-slate-500','text-slate-400')" class="text-xs uppercase font-medium mb-1">{{ key }}</div>
              <div v-if="typeof val === 'object' && val !== null"
                :class="dk('bg-slate-800 text-slate-200','bg-slate-50 text-slate-800')"
                class="p-3 rounded-lg text-xs font-mono whitespace-pre-wrap break-all max-h-48 overflow-y-auto">
                {{ JSON.stringify(val, null, 2) }}
              </div>
              <div v-else :class="dk('text-slate-200','text-slate-800')" class="text-sm break-all">
                {{ val || '-' }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.expand-enter-active, .expand-leave-active { transition: all 0.2s ease; overflow: hidden; }
.expand-enter-from, .expand-leave-to { max-height: 0; opacity: 0; }
.expand-enter-to, .expand-leave-from { max-height: 200px; opacity: 1; }
</style>
