<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { settingsAPI } from '@/api/client'
import {
  CalendarDays, Plus, RefreshCw, X, Save, Lock, Unlock,
  CheckCircle, AlertCircle, Clock, ChevronRight
} from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(false)
const saving = ref(false)
const fiscalYears = ref<any[]>([])
const showModal = ref(false)
const closingId = ref('')

const form = ref({
  name: '', start_date: '', end_date: '', is_current: false
})

async function load() {
  loading.value = true
  try {
    const r = await settingsAPI.getFiscalYears()
    fiscalYears.value = Array.isArray(r.data) ? r.data : []
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error loading fiscal years', 'error')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  const now = new Date()
  const year = now.getFullYear()
  form.value = {
    name: `FY ${year}`,
    start_date: `${year}-01-01`,
    end_date: `${year}-12-31`,
    is_current: fiscalYears.value.filter(f => f.status !== 'closed').length === 0
  }
  showModal.value = true
}

async function save() {
  if (!form.value.name || !form.value.start_date || !form.value.end_date) {
    app.addToast('Name, start date and end date are required', 'error'); return
  }
  if (new Date(form.value.start_date) >= new Date(form.value.end_date)) {
    app.addToast('End date must be after start date', 'error'); return
  }
  saving.value = true
  try {
    await settingsAPI.createFiscalYear(form.value)
    app.addToast('Fiscal year created', 'success')
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error creating fiscal year', 'error')
  } finally {
    saving.value = false
  }
}

async function closeFiscalYear(fy: any) {
  if (!confirm(`Close fiscal year "${fy.name}"? This action cannot be undone.`)) return
  closingId.value = fy.id
  try {
    await settingsAPI.closeFiscalYear(fy.id)
    app.addToast(`Fiscal year "${fy.name}" closed`, 'success')
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error closing fiscal year', 'error')
  } finally {
    closingId.value = ''
  }
}

function statusBadge(fy: any) {
  if (fy.status === 'closed') return 'bg-red-500/10 text-red-400'
  if (fy.is_current) return 'bg-emerald-500/10 text-emerald-400'
  return 'bg-sky-500/10 text-sky-400'
}
function statusLabel(fy: any) {
  if (fy.status === 'closed') return 'Closed'
  if (fy.is_current) return 'Current'
  return 'Open'
}
function fmtDate(d: string) {
  if (!d) return '-'
  return new Date(d).toLocaleDateString('fr-DZ', { year: 'numeric', month: 'short', day: 'numeric' })
}
function duration(start: string, end: string) {
  if (!start || !end) return ''
  const ms = new Date(end).getTime() - new Date(start).getTime()
  const days = Math.round(ms / 86400000)
  return `${days} days`
}

const currentYear = computed(() => fiscalYears.value.find(f => f.is_current))
const openYears = computed(() => fiscalYears.value.filter(f => f.status !== 'closed' && !f.is_current))
const closedYears = computed(() => fiscalYears.value.filter(f => f.status === 'closed'))

const dk = (d: string, l: string) => app.darkMode ? d : l
onMounted(load)
</script>

<template>
  <div :class="dk('bg-slate-950 text-slate-100','bg-slate-50 text-slate-900')" class="min-h-screen">
    <!-- Header -->
    <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border-b px-6 py-4 sticky top-0 z-10">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-sky-600/20">
            <CalendarDays class="w-5 h-5 text-sky-400" />
          </div>
          <div>
            <h1 class="text-lg font-bold">Fiscal Years</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">Manage accounting periods and fiscal year settings</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
            class="p-2 rounded-lg transition-colors disabled:opacity-50">
            <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          </button>
          <button @click="openCreate"
            class="flex items-center gap-2 px-4 py-2 bg-sky-600 hover:bg-sky-500 text-white rounded-lg text-sm font-medium transition-colors">
            <Plus class="w-4 h-4" /> New Fiscal Year
          </button>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="p-6 space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="animate-spin w-8 h-8 text-sky-400" />
      </div>

      <template v-else>
        <!-- Summary cards -->
        <div class="grid grid-cols-3 gap-4">
          <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl p-4 text-center">
            <div class="text-2xl font-bold text-emerald-400">{{ fiscalYears.filter(f=>!f.is_closed && f.is_current).length }}</div>
            <div :class="dk('text-slate-400','text-slate-500')" class="text-xs mt-1">Current Year</div>
          </div>
          <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl p-4 text-center">
            <div class="text-2xl font-bold text-sky-400">{{ openYears.length }}</div>
            <div :class="dk('text-slate-400','text-slate-500')" class="text-xs mt-1">Open Years</div>
          </div>
          <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl p-4 text-center">
            <div class="text-2xl font-bold text-red-400">{{ closedYears.length }}</div>
            <div :class="dk('text-slate-400','text-slate-500')" class="text-xs mt-1">Closed Years</div>
          </div>
        </div>

        <div v-if="fiscalYears.length===0" :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
          class="border rounded-xl p-16 text-center">
          <CalendarDays :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12 mx-auto mb-4" />
          <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">No fiscal years found. Create your first one.</p>
        </div>

        <!-- Fiscal year list -->
        <div v-else :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl overflow-hidden">
          <div :class="dk('bg-slate-800/40 border-slate-800 text-slate-400','bg-slate-50 border-slate-100 text-slate-500')"
            class="border-b grid grid-cols-12 gap-4 px-5 py-2.5 text-xs uppercase tracking-wider font-medium">
            <div class="col-span-3">Name</div>
            <div class="col-span-2">Start Date</div>
            <div class="col-span-2">End Date</div>
            <div class="col-span-2">Duration</div>
            <div class="col-span-1 text-center">Status</div>
            <div class="col-span-1 text-center">Closed At</div>
            <div class="col-span-1 text-right">Actions</div>
          </div>

          <div class="divide-y" :class="dk('divide-slate-800','divide-slate-100')">
            <div v-for="fy in fiscalYears" :key="fy.id"
              :class="[dk('hover:bg-slate-800/40','hover:bg-slate-50'), fy.is_current ? dk('border-l-2 border-emerald-500','border-l-2 border-emerald-500') : '']"
              class="grid grid-cols-12 gap-4 px-5 py-4 items-center transition-colors">
              <!-- Name -->
              <div class="col-span-3">
                <div class="flex items-center gap-2">
                  <div :class="fy.status==='closed' ? 'bg-red-500/10' : fy.is_current ? 'bg-emerald-500/10' : 'bg-sky-500/10'"
                    class="p-1.5 rounded-lg">
                    <Lock v-if="fy.status==='closed'" :class="'w-3.5 h-3.5 text-red-400'" />
                    <CheckCircle v-else-if="fy.is_current" class="w-3.5 h-3.5 text-emerald-400" />
                    <CalendarDays v-else class="w-3.5 h-3.5 text-sky-400" />
                  </div>
                  <span class="font-semibold text-sm">{{ fy.name }}</span>
                </div>
              </div>
              <!-- Start -->
              <div class="col-span-2 text-sm" :class="dk('text-slate-300','text-slate-700')">{{ fmtDate(fy.start_date) }}</div>
              <!-- End -->
              <div class="col-span-2 text-sm" :class="dk('text-slate-300','text-slate-700')">{{ fmtDate(fy.end_date) }}</div>
              <!-- Duration -->
              <div class="col-span-2 text-xs" :class="dk('text-slate-500','text-slate-400')">
                {{ duration(fy.start_date, fy.end_date) }}
              </div>
              <!-- Status -->
              <div class="col-span-1 flex justify-center">
                <span :class="statusBadge(fy)" class="text-xs px-2 py-0.5 rounded-full font-medium">
                  {{ statusLabel(fy) }}
                </span>
              </div>
              <!-- Closed At -->
              <div class="col-span-1 text-xs text-center" :class="dk('text-slate-500','text-slate-400')">
                {{ fy.closed_at ? fmtDate(fy.closed_at) : '-' }}
              </div>
              <!-- Actions -->
              <div class="col-span-1 flex justify-end">
                <button v-if="fy.status !== 'closed'" @click="closeFiscalYear(fy)"
                  :disabled="closingId===fy.id"
                  class="flex items-center gap-1 px-2.5 py-1.5 bg-red-500/10 hover:bg-red-500/20 text-red-400 rounded-lg text-xs font-medium transition-colors disabled:opacity-50">
                  <Lock class="w-3 h-3" />
                  {{ closingId===fy.id ? '...' : 'Close' }}
                </button>
                <span v-else :class="dk('text-slate-600','text-slate-400')" class="text-xs px-2">Closed</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Info box -->
        <div :class="dk('bg-sky-900/20 border-sky-800','bg-sky-50 border-sky-200')"
          class="border rounded-xl p-4 flex items-start gap-3">
          <AlertCircle class="w-4 h-4 text-sky-400 flex-shrink-0 mt-0.5" />
          <div class="text-xs" :class="dk('text-sky-300','text-sky-700')">
            <strong>Important:</strong> Closing a fiscal year prevents any new postings to that period. Ensure all journal entries for the period are finalized before closing. This action cannot be reversed.
          </div>
        </div>
      </template>
    </div>

    <!-- Create Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4"
        style="background:rgba(0,0,0,0.7)" @click.self="showModal=false">
        <div :class="dk('bg-slate-900 border-slate-700','bg-white border-slate-200')"
          class="border rounded-2xl w-full max-w-md shadow-2xl">
          <div :class="dk('border-slate-800','border-slate-100')" class="flex items-center justify-between px-6 py-4 border-b">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-xl bg-sky-600/20">
                <CalendarDays class="w-4 h-4 text-sky-400" />
              </div>
              <h2 class="font-bold">New Fiscal Year</h2>
            </div>
            <button @click="showModal=false" :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-900')"
              class="p-1.5 rounded-lg transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="px-6 py-5 space-y-4">
            <div>
              <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">
                Year Name <span class="text-red-400">*</span>
              </label>
              <input v-model="form.name" placeholder="FY 2026"
                :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-sky-500" />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">
                  Start Date <span class="text-red-400">*</span>
                </label>
                <input v-model="form.start_date" type="date"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-sky-500" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">
                  End Date <span class="text-red-400">*</span>
                </label>
                <input v-model="form.end_date" type="date"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-sky-500" />
              </div>
            </div>
            <div class="flex items-center gap-3 pt-1">
              <button type="button" @click="form.is_current = !form.is_current"
                :class="form.is_current ? 'bg-emerald-500' : 'bg-slate-600'"
                class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors">
                <span :class="form.is_current ? 'translate-x-6' : 'translate-x-1'"
                  class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
              </button>
              <span class="text-sm" :class="dk('text-slate-300','text-slate-700')">Set as Current Year</span>
            </div>
          </div>
          <div :class="dk('border-slate-800 bg-slate-900/50','border-slate-100 bg-slate-50')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t">
            <button @click="showModal=false"
              :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
              class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="save" :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 bg-sky-600 hover:bg-sky-500 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors">
              <Save class="w-4 h-4" />
              {{ saving ? 'Creating...' : 'Create Fiscal Year' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
