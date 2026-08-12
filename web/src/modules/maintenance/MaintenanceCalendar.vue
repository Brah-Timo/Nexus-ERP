<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Calendar, ChevronLeft, ChevronRight, RefreshCw,
  Wrench, Shield, Search, AlertTriangle, Clock, CheckCircle, X
} from 'lucide-vue-next'
import { maintenanceAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const app = useAppStore()

// ─── types ────────────────────────────────────────────────────────────────────
interface CalendarEvent {
  id: string
  title: string
  date: string
  type: string          // order_type or 'schedule'
  status: string
  equipment_name?: string
  order_number?: string
  color?: string
}

// ─── state ────────────────────────────────────────────────────────────────────
const today      = new Date()
const viewYear   = ref(today.getFullYear())
const viewMonth  = ref(today.getMonth())   // 0-based
const events     = ref<CalendarEvent[]>([])
const loading    = ref(false)
const selected   = ref<CalendarEvent | null>(null)
const showDetail = ref(false)
const typeFilter = ref('all')

// ─── computed ────────────────────────────────────────────────────────────────
const dk = (a: string, b: string) => app.darkMode ? a : b

const monthName = computed(() =>
  new Date(viewYear.value, viewMonth.value, 1)
    .toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
)

const firstDOW = computed(() =>
  new Date(viewYear.value, viewMonth.value, 1).getDay()   // 0=Sun
)

const daysInMonth = computed(() =>
  new Date(viewYear.value, viewMonth.value + 1, 0).getDate()
)

// Build a 6-row × 7-col grid (some cells are null = prev/next month padding)
const calendarGrid = computed(() => {
  const cells: (number | null)[] = []
  for (let i = 0; i < firstDOW.value; i++) cells.push(null)
  for (let d = 1; d <= daysInMonth.value; d++) cells.push(d)
  while (cells.length % 7 !== 0) cells.push(null)
  return cells
})

const filteredEvents = computed(() => {
  if (typeFilter.value === 'all') return events.value
  return events.value.filter(e => e.type === typeFilter.value)
})

const eventsByDay = computed(() => {
  const map: Record<number, CalendarEvent[]> = {}
  for (const ev of filteredEvents.value) {
    const d = new Date(ev.date)
    if (d.getFullYear() === viewYear.value && d.getMonth() === viewMonth.value) {
      const day = d.getDate()
      if (!map[day]) map[day] = []
      map[day].push(ev)
    }
  }
  return map
})

const typeCounts = computed(() => {
  const types = ['corrective','preventive','inspection','emergency','upgrade','schedule']
  return types.map(t => ({ type: t, count: events.value.filter(e => e.type === t).length }))
    .filter(t => t.count > 0)
})

// ─── helpers ─────────────────────────────────────────────────────────────────
const DAYS   = ['Sun','Mon','Tue','Wed','Thu','Fri','Sat']

const typeColor = (t: string) => ({
  corrective:'#3b82f6', preventive:'#10b981',
  inspection:'#8b5cf6', emergency:'#f43f5e',
  upgrade:'#f59e0b',    schedule:'#06b6d4',
}[t] ?? '#64748b')

const statusBadge = (s: string) => ({
  pending:'bg-slate-500/15 text-slate-400',
  in_progress:'bg-blue-500/15 text-blue-400',
  on_hold:'bg-amber-500/15 text-amber-400',
  completed:'bg-emerald-500/15 text-emerald-400',
  cancelled:'bg-rose-500/15 text-rose-400',
  scheduled:'bg-violet-500/15 text-violet-400',
}[s] ?? 'bg-slate-500/15 text-slate-400')

const statusLabel = (s: string) => ({
  pending:'Pending', in_progress:'In Progress', on_hold:'On Hold',
  completed:'Completed', cancelled:'Cancelled', scheduled:'Scheduled'
}[s] ?? s)

const typeLabel = (t: string) => ({
  corrective:'Corrective', preventive:'Preventive',
  inspection:'Inspection', emergency:'Emergency',
  upgrade:'Upgrade', schedule:'Schedule'
}[t] ?? t)

const fmtDate = (s: string) =>
  new Date(s).toLocaleDateString('fr-DZ', { weekday:'long', day:'2-digit', month:'long', year:'numeric' })

const isToday = (day: number) =>
  day === today.getDate() &&
  viewMonth.value === today.getMonth() &&
  viewYear.value === today.getFullYear()

const isPast = (day: number) => {
  const d = new Date(viewYear.value, viewMonth.value, day)
  d.setHours(0,0,0,0)
  const t = new Date(); t.setHours(0,0,0,0)
  return d < t
}

// ─── navigation ──────────────────────────────────────────────────────────────
const prevMonth = () => {
  if (viewMonth.value === 0) { viewMonth.value = 11; viewYear.value-- }
  else viewMonth.value--
  load()
}

const nextMonth = () => {
  if (viewMonth.value === 11) { viewMonth.value = 0; viewYear.value++ }
  else viewMonth.value++
  load()
}

const goToday = () => {
  viewYear.value  = today.getFullYear()
  viewMonth.value = today.getMonth()
  load()
}

// ─── data loading ─────────────────────────────────────────────────────────────
const load = async () => {
  loading.value = true
  try {
    const m = String(viewMonth.value + 1).padStart(2,'0')
    const res = await maintenanceAPI.getCalendar({
      year:  String(viewYear.value),
      month: m,
    })
    events.value = res.data.events ?? res.data ?? []
  } catch {
    app.addToast('Failed to load calendar', 'error')
  } finally {
    loading.value = false
  }
}

const openDetail = (ev: CalendarEvent) => {
  selected.value = ev
  showDetail.value = true
}

onMounted(load)
</script>

<template>
  <div :class="['min-h-screen p-6 space-y-6', dk('bg-slate-950 text-slate-100','bg-slate-50 text-slate-900')]">

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-cyan-500/15 flex items-center justify-center">
          <Calendar class="w-5 h-5 text-cyan-400" />
        </div>
        <div>
          <h1 class="text-xl font-bold">Maintenance Calendar</h1>
          <p :class="['text-sm', dk('text-slate-400','text-slate-500')]">Scheduled work orders and maintenance events</p>
        </div>
      </div>
      <!-- Month nav -->
      <div class="flex items-center gap-2">
        <button @click="goToday"
          :class="['px-3 py-1.5 rounded-lg border text-sm transition-colors',
            dk('bg-slate-900 border-slate-700 text-slate-300 hover:bg-slate-800',
               'bg-white border-slate-200 text-slate-700 hover:bg-slate-50')]">
          Today
        </button>
        <button @click="prevMonth"
          :class="['p-2 rounded-lg border transition-colors',
            dk('bg-slate-900 border-slate-700 text-slate-300 hover:bg-slate-800',
               'bg-white border-slate-200 text-slate-700 hover:bg-slate-50')]">
          <ChevronLeft class="w-4 h-4" />
        </button>
        <span class="text-sm font-semibold w-40 text-center">{{ monthName }}</span>
        <button @click="nextMonth"
          :class="['p-2 rounded-lg border transition-colors',
            dk('bg-slate-900 border-slate-700 text-slate-300 hover:bg-slate-800',
               'bg-white border-slate-200 text-slate-700 hover:bg-slate-50')]">
          <ChevronRight class="w-4 h-4" />
        </button>
        <button @click="load"
          :class="['p-2 rounded-lg border transition-colors',
            dk('bg-slate-900 border-slate-700 text-slate-300','bg-white border-slate-200 text-slate-700')]">
          <RefreshCw :class="['w-4 h-4', loading && 'animate-spin']" />
        </button>
      </div>
    </div>

    <!-- Type Filter Pills -->
    <div class="flex flex-wrap gap-2">
      <button @click="typeFilter='all'"
        :class="['px-3 py-1.5 rounded-lg text-xs font-medium border transition-colors',
          typeFilter==='all'
            ? 'bg-cyan-600 text-white border-cyan-600'
            : dk('bg-slate-900 border-slate-700 text-slate-400 hover:border-slate-500',
                 'bg-white border-slate-200 text-slate-500 hover:border-slate-400')]">
        All Events ({{ events.length }})
      </button>
      <button v-for="t in typeCounts" :key="t.type"
        @click="typeFilter=t.type"
        :class="['px-3 py-1.5 rounded-lg text-xs font-medium border transition-colors',
          typeFilter===t.type
            ? 'text-white border-transparent'
            : dk('bg-slate-900 border-slate-700 text-slate-400 hover:border-slate-500',
                 'bg-white border-slate-200 text-slate-500 hover:border-slate-400')]"
        :style="typeFilter===t.type ? { backgroundColor: typeColor(t.type) } : {}">
        {{ typeLabel(t.type) }} ({{ t.count }})
      </button>
    </div>

    <!-- Calendar Grid -->
    <div :class="['rounded-xl border overflow-hidden', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">

      <!-- Day-of-week headers -->
      <div class="grid grid-cols-7 border-b" :class="dk('border-slate-800 bg-slate-800/50','border-slate-100 bg-slate-50')">
        <div v-for="d in DAYS" :key="d"
          :class="['text-center py-3 text-xs font-semibold', dk('text-slate-400','text-slate-500')]">
          {{ d }}
        </div>
      </div>

      <!-- Loading overlay -->
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="w-8 h-8 animate-spin text-cyan-400" />
      </div>

      <!-- Grid body -->
      <div v-else class="grid grid-cols-7" style="min-height:520px">
        <div v-for="(day, idx) in calendarGrid" :key="idx"
          :class="['border-r border-b min-h-24 p-1.5 relative',
            dk('border-slate-800','border-slate-100'),
            !day ? dk('bg-slate-900/40','bg-slate-50/60') : '',
            day && isToday(day)
              ? dk('bg-cyan-950/30','bg-cyan-50')
              : '',
            day && isPast(day) && !isToday(day)
              ? dk('opacity-60','opacity-60')
              : '']">

          <!-- Day number -->
          <div v-if="day" class="mb-1 flex items-center justify-end">
            <span :class="['text-xs font-medium w-6 h-6 flex items-center justify-center rounded-full',
              isToday(day)
                ? 'bg-cyan-500 text-white font-bold'
                : dk('text-slate-400','text-slate-500')]">
              {{ day }}
            </span>
          </div>

          <!-- Events -->
          <template v-if="day && eventsByDay[day]">
            <div v-for="(ev, ei) in eventsByDay[day].slice(0, 3)" :key="ev.id"
              @click="openDetail(ev)"
              :class="['rounded px-1.5 py-0.5 text-xs cursor-pointer mb-0.5 truncate transition-opacity hover:opacity-80']"
              :style="{ backgroundColor: (ev.color || typeColor(ev.type)) + '25', color: ev.color || typeColor(ev.type), borderLeft: `2px solid ${ev.color || typeColor(ev.type)}` }">
              {{ ev.title }}
            </div>
            <div v-if="eventsByDay[day].length > 3"
              :class="['text-xs px-1 cursor-pointer', dk('text-slate-500','text-slate-400')]"
              @click="openDetail(eventsByDay[day][3])">
              +{{ eventsByDay[day].length - 3 }} more
            </div>
          </template>
        </div>
      </div>
    </div>

    <!-- Legend -->
    <div :class="['rounded-xl border p-4', dk('bg-slate-900 border-slate-800','bg-white border-slate-200')]">
      <div class="flex flex-wrap gap-4 text-xs">
        <div v-for="t in typeCounts" :key="t.type" class="flex items-center gap-1.5">
          <div class="w-3 h-3 rounded-sm" :style="{ backgroundColor: typeColor(t.type) }"></div>
          <span :class="dk('text-slate-400','text-slate-500')">{{ typeLabel(t.type) }}</span>
        </div>
      </div>
    </div>

    <!-- ── Event Detail Modal ──────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showDetail && selected"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
        @click.self="showDetail=false">
        <div :class="['w-full max-w-sm rounded-2xl border shadow-2xl', dk('bg-slate-900 border-slate-700','bg-white border-slate-200')]">
          <!-- color bar at top -->
          <div class="h-1.5 rounded-t-2xl" :style="{ backgroundColor: selected.color || typeColor(selected.type) }"></div>
          <div class="p-5 space-y-4">
            <div class="flex items-start justify-between gap-3">
              <div>
                <h2 class="font-semibold text-sm">{{ selected.title }}</h2>
                <p :class="['text-xs mt-0.5', dk('text-slate-400','text-slate-500')]">{{ fmtDate(selected.date) }}</p>
              </div>
              <button @click="showDetail=false"
                :class="['p-1.5 rounded-lg flex-shrink-0', dk('hover:bg-slate-800 text-slate-400','hover:bg-slate-100 text-slate-500')]">
                <X class="w-4 h-4" />
              </button>
            </div>
            <div class="flex flex-wrap gap-2">
              <span class="px-2.5 py-1 rounded-lg text-xs font-medium"
                :style="{ backgroundColor: (selected.color || typeColor(selected.type)) + '20', color: selected.color || typeColor(selected.type) }">
                {{ typeLabel(selected.type) }}
              </span>
              <span :class="['px-2.5 py-1 rounded-lg text-xs font-medium', statusBadge(selected.status)]">
                {{ statusLabel(selected.status) }}
              </span>
            </div>
            <div class="space-y-2 text-sm">
              <div v-if="selected.order_number" class="flex justify-between">
                <span :class="dk('text-slate-400','text-slate-500')">Order #</span>
                <span class="font-mono text-xs font-medium">{{ selected.order_number }}</span>
              </div>
              <div v-if="selected.equipment_name" class="flex justify-between">
                <span :class="dk('text-slate-400','text-slate-500')">Equipment</span>
                <span class="font-medium text-xs">{{ selected.equipment_name }}</span>
              </div>
            </div>
            <button @click="showDetail=false"
              :class="['w-full py-2 rounded-lg border text-sm transition-colors',
                dk('border-slate-700 text-slate-300 hover:bg-slate-800','border-slate-200 text-slate-600 hover:bg-slate-50')]">
              Close
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
