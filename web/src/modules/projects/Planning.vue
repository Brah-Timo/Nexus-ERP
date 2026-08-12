<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-950 p-6 space-y-6">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Resource Planning</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Plan and visualize team capacity across projects</p>
      </div>
      <button @click="openCreate" class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg transition-colors shadow-sm">
        <Plus class="w-4 h-4" />
        Add Slot
      </button>
    </div>

    <!-- Controls -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
      <div class="flex flex-wrap gap-3 items-center">
        <div class="flex items-center gap-2">
          <button @click="prevWeek" class="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors text-gray-600 dark:text-gray-300">
            <ChevronLeft class="w-4 h-4" />
          </button>
          <span class="text-sm font-semibold text-gray-900 dark:text-white min-w-44 text-center">{{ weekRangeLabel }}</span>
          <button @click="nextWeek" class="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors text-gray-600 dark:text-gray-300">
            <ChevronRight class="w-4 h-4" />
          </button>
          <button @click="goToday" class="px-3 py-1.5 text-xs font-medium bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 rounded-lg hover:bg-indigo-200 dark:hover:bg-indigo-900/50 transition-colors">Today</button>
        </div>
        <div class="relative">
          <select v-model="filterEmployee" @change="load" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="">All Employees</option>
            <option v-for="e in employees" :key="e.id" :value="e.id">{{ e.full_name }}</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
        <div class="relative">
          <select v-model="filterProject" @change="load" class="appearance-none pl-4 pr-10 py-2 text-sm bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white">
            <option value="">All Projects</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <ChevronDown class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
        </div>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-20">
      <Loader2 class="w-10 h-10 text-indigo-500 animate-spin" />
    </div>

    <!-- Grid Calendar -->
    <div v-else class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
      <!-- Day headers -->
      <div class="grid border-b border-gray-200 dark:border-gray-800" :style="{ gridTemplateColumns: `200px repeat(7, 1fr)` }">
        <div class="px-4 py-3 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide bg-gray-50 dark:bg-gray-800/60">Employee</div>
        <div v-for="day in weekDays" :key="day.iso"
          :class="day.isToday ? 'bg-indigo-50 dark:bg-indigo-900/20' : 'bg-gray-50 dark:bg-gray-800/60'"
          class="px-3 py-3 text-center border-l border-gray-200 dark:border-gray-800">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ day.weekday }}</p>
          <p :class="day.isToday ? 'text-indigo-600 dark:text-indigo-400 font-bold' : 'text-gray-900 dark:text-white font-semibold'" class="text-sm">{{ day.dayNum }}</p>
        </div>
      </div>

      <!-- Employee rows -->
      <div v-if="groupedByEmployee.length === 0" class="flex flex-col items-center justify-center py-16 text-gray-400">
        <CalendarDays class="w-12 h-12 mb-3 opacity-30" />
        <p class="text-sm">No planning slots found for this week</p>
      </div>
      <div v-else>
        <div v-for="(empRow, idx) in groupedByEmployee" :key="empRow.employee_id"
          :class="idx % 2 === 0 ? 'bg-white dark:bg-gray-900' : 'bg-gray-50/50 dark:bg-gray-800/20'"
          class="grid border-b border-gray-100 dark:border-gray-800 hover:bg-indigo-50/20 dark:hover:bg-indigo-900/5 transition-colors"
          :style="{ gridTemplateColumns: `200px repeat(7, 1fr)` }">

          <!-- Employee cell -->
          <div class="px-4 py-3 flex items-center gap-2 border-r border-gray-100 dark:border-gray-800">
            <div class="w-7 h-7 rounded-full bg-gradient-to-br from-indigo-400 to-violet-500 flex items-center justify-center text-white text-xs font-bold flex-shrink-0">
              {{ empRow.employee_name.charAt(0) }}
            </div>
            <div class="min-w-0">
              <p class="text-sm font-medium text-gray-900 dark:text-white truncate">{{ empRow.employee_name }}</p>
              <p class="text-xs text-gray-400">{{ empRow.total_hours.toFixed(1) }}h / week</p>
            </div>
          </div>

          <!-- Day cells -->
          <div v-for="day in weekDays" :key="day.iso"
            :class="day.isToday ? 'bg-indigo-50/30 dark:bg-indigo-900/10' : ''"
            class="px-2 py-2 border-l border-gray-100 dark:border-gray-800 min-h-14">
            <div v-for="slot in slotsForCell(empRow.employee_id, day.iso)" :key="slot.id">
              <div :class="slotColor(slot.project_name)" class="rounded-md px-2 py-1 mb-1 cursor-pointer hover:opacity-80 transition-opacity group relative" @click="openEdit(slot)">
                <p class="text-xs font-semibold text-white truncate">{{ slot.project_name }}</p>
                <p class="text-xs text-white/80">{{ slot.planned_hours }}h</p>
                <button @click.stop="deleteSlot(slot.id)" class="absolute top-0.5 right-0.5 opacity-0 group-hover:opacity-100 w-4 h-4 bg-white/30 rounded flex items-center justify-center transition-opacity">
                  <X class="w-2.5 h-2.5 text-white" />
                </button>
              </div>
            </div>
            <!-- Add slot button -->
            <button @click="openCreateForCell(empRow.employee_id, day.iso)" class="w-full text-xs text-gray-300 dark:text-gray-600 hover:text-indigo-400 dark:hover:text-indigo-500 transition-colors py-0.5 opacity-0 hover:opacity-100 group-hover:opacity-100">
              <Plus class="w-3 h-3 mx-auto" />
            </button>
          </div>
        </div>
      </div>

      <!-- Capacity summary footer -->
      <div class="border-t border-gray-200 dark:border-gray-800 px-4 py-3 bg-gray-50 dark:bg-gray-800/40 flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
        <span>{{ planningSlots.length }} planning slots &bull; Week of {{ weekRangeLabel }}</span>
        <span>Total planned: <strong class="text-indigo-600 dark:text-indigo-400">{{ totalPlannedHours.toFixed(1) }}h</strong></span>
      </div>
    </div>

    <!-- Slot List Below Grid -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
      <div class="flex items-center justify-between px-5 py-4 border-b border-gray-100 dark:border-gray-800">
        <h3 class="font-semibold text-gray-900 dark:text-white text-sm">All Planning Slots</h3>
        <span class="text-xs text-gray-400">{{ planningSlots.length }} slots</span>
      </div>
      <table class="w-full text-sm">
        <thead>
          <tr class="border-b border-gray-100 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/60">
            <th class="text-left px-4 py-2.5 font-semibold text-gray-600 dark:text-gray-300">Employee</th>
            <th class="text-left px-4 py-2.5 font-semibold text-gray-600 dark:text-gray-300">Project</th>
            <th class="text-left px-4 py-2.5 font-semibold text-gray-600 dark:text-gray-300">Date</th>
            <th class="text-right px-4 py-2.5 font-semibold text-gray-600 dark:text-gray-300">Hours</th>
            <th class="text-left px-4 py-2.5 font-semibold text-gray-600 dark:text-gray-300">Note</th>
            <th class="px-4 py-2.5 w-16"></th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
          <tr v-for="slot in planningSlots" :key="slot.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/40 transition-colors">
            <td class="px-4 py-2.5 text-gray-900 dark:text-white font-medium">{{ slot.employee_name }}</td>
            <td class="px-4 py-2.5 text-gray-600 dark:text-gray-300">{{ slot.project_name }}</td>
            <td class="px-4 py-2.5 text-gray-600 dark:text-gray-300">{{ fmtDate(slot.slot_date) }}</td>
            <td class="px-4 py-2.5 text-right font-bold text-indigo-600 dark:text-indigo-400">{{ slot.planned_hours }}h</td>
            <td class="px-4 py-2.5 text-gray-400 text-xs truncate max-w-32">{{ slot.note || '—' }}</td>
            <td class="px-4 py-2.5" @click.stop>
              <div class="flex gap-1">
                <button @click="openEdit(slot)" class="p-1.5 hover:bg-indigo-50 dark:hover:bg-indigo-900/30 rounded-lg text-gray-400 hover:text-indigo-600 transition-colors">
                  <Pencil class="w-3.5 h-3.5" />
                </button>
                <button @click="deleteSlot(slot.id)" class="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg text-gray-400 hover:text-red-600 transition-colors">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div class="absolute inset-0 bg-black/60 backdrop-blur-sm" @click="closeModal" />
          <div class="relative w-full max-w-md bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden">
            <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-indigo-600 to-violet-600">
              <div class="flex items-center gap-3 text-white">
                <CalendarDays class="w-5 h-5" />
                <h2 class="font-semibold text-lg">{{ editingSlot ? 'Edit Slot' : 'Add Planning Slot' }}</h2>
              </div>
              <button @click="closeModal" class="p-1 hover:bg-white/20 rounded-lg text-white transition-colors">
                <X class="w-5 h-5" />
              </button>
            </div>
            <form @submit.prevent="saveSlot" class="p-6 space-y-4">
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Employee *</label>
                <div class="relative">
                  <select v-model="form.employee_id" required class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                    <option value="">— Select Employee —</option>
                    <option v-for="e in employees" :key="e.id" :value="e.id">{{ e.full_name }}</option>
                  </select>
                  <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                </div>
              </div>
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Project *</label>
                <div class="relative">
                  <select v-model="form.project_id" required class="w-full appearance-none px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white pr-8">
                    <option value="">— Select Project —</option>
                    <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
                  </select>
                  <ChevronDown class="absolute right-2.5 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400 pointer-events-none" />
                </div>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Date *</label>
                  <input type="date" v-model="form.slot_date" required class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Planned Hours *</label>
                  <input type="number" v-model.number="form.planned_hours" required min="0.5" max="24" step="0.5" class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
                </div>
              </div>
              <div>
                <label class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5">Note</label>
                <input v-model="form.note" placeholder="Optional note..." class="w-full px-3 py-2.5 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 text-gray-900 dark:text-white" />
              </div>
              <div class="flex gap-3 pt-2">
                <button type="button" @click="closeModal" class="flex-1 px-4 py-2.5 border border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-200 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">Cancel</button>
                <button type="submit" :disabled="saving" class="flex-1 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-medium transition-colors disabled:opacity-50 flex items-center justify-center gap-2">
                  <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                  {{ saving ? 'Saving...' : 'Save' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </Transition>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  Plus, Loader2, Pencil, Trash2, X, ChevronDown, ChevronLeft, ChevronRight, CalendarDays
} from 'lucide-vue-next'
import { projectsAPI, hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const store = useAppStore()

const loading = ref(true)
const saving = ref(false)
const planningSlots = ref<any[]>([])
const projects = ref<any[]>([])
const employees = ref<any[]>([])
const filterEmployee = ref('')
const filterProject = ref('')
const showModal = ref(false)
const editingSlot = ref<any>(null)
const currentMonday = ref(getMonday(new Date()))

const form = ref({ employee_id: '', project_id: '', slot_date: '', planned_hours: 8, note: '' })

function getMonday(d: Date) {
  const dt = new Date(d)
  const day = dt.getDay()
  const diff = dt.getDate() - day + (day === 0 ? -6 : 1)
  dt.setDate(diff)
  dt.setHours(0, 0, 0, 0)
  return dt
}

const weekDays = computed(() => {
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(currentMonday.value)
    d.setDate(d.getDate() + i)
    const today = new Date(); today.setHours(0, 0, 0, 0)
    return {
      iso: d.toISOString().slice(0, 10),
      weekday: d.toLocaleDateString('en-US', { weekday: 'short' }),
      dayNum: d.getDate(),
      isToday: d.getTime() === today.getTime()
    }
  })
})

const weekRangeLabel = computed(() => {
  const start = weekDays.value[0]
  const end = weekDays.value[6]
  const startD = new Date(start.iso)
  const endD = new Date(end.iso)
  return `${startD.toLocaleDateString('en-GB', { day: '2-digit', month: 'short' })} – ${endD.toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })}`
})

const totalPlannedHours = computed(() => planningSlots.value.reduce((s, sl) => s + (sl.planned_hours || 0), 0))

const groupedByEmployee = computed(() => {
  const map = new Map<string, { employee_id: string; employee_name: string; total_hours: number }>()
  planningSlots.value.forEach(s => {
    if (!map.has(s.employee_id)) {
      map.set(s.employee_id, { employee_id: s.employee_id, employee_name: s.employee_name || '?', total_hours: 0 })
    }
    map.get(s.employee_id)!.total_hours += s.planned_hours || 0
  })
  return [...map.values()]
})

function slotsForCell(empId: string, date: string) {
  return planningSlots.value.filter(s => s.employee_id === empId && s.slot_date?.slice(0, 10) === date)
}

const SLOT_COLORS = [
  'bg-indigo-500', 'bg-violet-500', 'bg-emerald-500', 'bg-amber-500',
  'bg-rose-500', 'bg-cyan-500', 'bg-pink-500', 'bg-teal-500'
]
const projectColorMap = new Map<string, string>()
function slotColor(projName: string) {
  if (!projectColorMap.has(projName)) {
    projectColorMap.set(projName, SLOT_COLORS[projectColorMap.size % SLOT_COLORS.length])
  }
  return projectColorMap.get(projName)!
}

function fmtDate(d?: string) { if (!d) return '—'; return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' }) }

function prevWeek() { const d = new Date(currentMonday.value); d.setDate(d.getDate() - 7); currentMonday.value = d; load() }
function nextWeek() { const d = new Date(currentMonday.value); d.setDate(d.getDate() + 7); currentMonday.value = d; load() }
function goToday() { currentMonday.value = getMonday(new Date()); load() }

async function load() {
  loading.value = true
  try {
    const params: any = {
      start_date: weekDays.value[0].iso,
      end_date: weekDays.value[6].iso
    }
    if (filterEmployee.value) params.employee_id = filterEmployee.value
    if (filterProject.value) params.project_id = filterProject.value
    const res = await projectsAPI.getPlanningSlots(params)
    planningSlots.value = res.data || []
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Failed to load planning', 'error')
  } finally {
    loading.value = false
  }
}

async function loadDropdowns() {
  try {
    const [pRes, eRes] = await Promise.all([projectsAPI.getProjects(), hrAPI.getEmployees()])
    projects.value = pRes.data || []
    employees.value = eRes.data || []
  } catch { /* ignore */ }
}

function openCreate() {
  editingSlot.value = null
  form.value = { employee_id: '', project_id: '', slot_date: weekDays.value[0].iso, planned_hours: 8, note: '' }
  showModal.value = true
}
function openCreateForCell(empId: string, date: string) {
  editingSlot.value = null
  form.value = { employee_id: empId, project_id: '', slot_date: date, planned_hours: 8, note: '' }
  showModal.value = true
}
function openEdit(slot: any) {
  editingSlot.value = slot
  form.value = { employee_id: slot.employee_id, project_id: slot.project_id, slot_date: slot.slot_date?.slice(0, 10) || '', planned_hours: slot.planned_hours, note: slot.note || '' }
  showModal.value = true
}
function closeModal() { showModal.value = false; editingSlot.value = null }

async function saveSlot() {
  saving.value = true
  try {
    await projectsAPI.upsertPlanningSlot({ ...form.value, id: editingSlot.value?.id })
    store.addToast(editingSlot.value ? 'Slot updated' : 'Slot added', 'success')
    closeModal()
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

async function deleteSlot(id: string) {
  if (!confirm('Delete this planning slot?')) return
  try {
    await projectsAPI.deletePlanningSlot(id)
    store.addToast('Slot deleted', 'success')
    await load()
  } catch (e: any) {
    store.addToast(e?.response?.data?.error || 'Delete failed', 'error')
  }
}

onMounted(() => { load(); loadDropdowns() })
</script>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
