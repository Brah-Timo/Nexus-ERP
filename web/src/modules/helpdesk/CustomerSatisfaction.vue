<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { helpdeskAPI } from '@/api/client'
import { Star, Plus, RefreshCw, X, ThumbsUp, ThumbsDown, Smile, Meh, Frown, BarChart3 } from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(true)
const saving = ref(false)
const error = ref('')

const surveys = ref<any[]>([])
const aggregate = ref<any>({})
const agents = ref<any[]>([])
const tickets = ref<any[]>([])
const showForm = ref(false)
const filterRating = ref('')
const filterAgentId = ref('')

const form = ref({
  ticket_id: '', agent_id: '', rating: 'satisfied',
  comment: '', requester_name: '', requester_email: ''
})

const ratingOptions = ['very_dissatisfied','dissatisfied','neutral','satisfied','very_satisfied']
const ratingLabels: Record<string, string> = {
  very_dissatisfied: 'Very Dissatisfied',
  dissatisfied: 'Dissatisfied',
  neutral: 'Neutral',
  satisfied: 'Satisfied',
  very_satisfied: 'Very Satisfied',
}
const ratingColor: Record<string, string> = {
  very_dissatisfied: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
  dissatisfied: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300',
  neutral: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-300',
  satisfied: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  very_satisfied: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const params: Record<string, string> = {}
    if (filterRating.value) params.rating = filterRating.value
    if (filterAgentId.value) params.agent_id = filterAgentId.value
    const [s, a, t] = await Promise.all([
      helpdeskAPI.listCSAT(params),
      helpdeskAPI.listAgents(),
      helpdeskAPI.listTickets({ status: 'resolved' })
    ])
    surveys.value = s.data.surveys || []
    aggregate.value = s.data.aggregate || {}
    agents.value = a.data.agents || []
    tickets.value = t.data.tickets || []
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to load'
  } finally {
    loading.value = false
  }
}

async function submitSurvey() {
  if (!form.value.ticket_id) { error.value = 'Select a ticket'; return }
  saving.value = true
  error.value = ''
  try {
    await helpdeskAPI.createCSAT(form.value)
    showForm.value = false
    form.value = { ticket_id: '', agent_id: '', rating: 'satisfied', comment: '', requester_name: '', requester_email: '' }
    await load()
  } catch (e: any) {
    error.value = e?.response?.data?.error || 'Failed to submit'
  } finally {
    saving.value = false
  }
}

onMounted(load)

function ratingStars(rating: string) {
  const map: Record<string, number> = { very_dissatisfied: 1, dissatisfied: 2, neutral: 3, satisfied: 4, very_satisfied: 5 }
  return map[rating] || 0
}

function scoreColor(score: number) {
  if (score >= 4) return 'text-green-600'
  if (score >= 3) return 'text-yellow-600'
  return 'text-red-600'
}
</script>

<template>
  <div :class="app.darkMode ? 'bg-slate-900 text-slate-100' : 'bg-slate-50 text-slate-900'" class="min-h-screen p-6">
    <div class="flex items-center justify-between mb-6">
      <div class="flex items-center gap-3">
        <Star class="w-7 h-7 text-yellow-500" />
        <div>
          <h1 class="text-2xl font-bold">Customer Satisfaction</h1>
          <p class="text-sm text-slate-500">CSAT surveys and ratings</p>
        </div>
      </div>
      <div class="flex gap-2">
        <button @click="load" :class="app.darkMode ? 'bg-slate-700 text-slate-200' : 'bg-white text-slate-700'"
          class="p-2 rounded-lg border border-slate-200 dark:border-slate-600 hover:bg-slate-100 dark:hover:bg-slate-600 transition-colors">
          <RefreshCw class="w-4 h-4" :class="loading ? 'animate-spin' : ''" />
        </button>
        <button @click="showForm = true" class="flex items-center gap-2 px-4 py-2 bg-yellow-500 hover:bg-yellow-600 text-white rounded-lg text-sm font-medium transition-colors">
          <Plus class="w-4 h-4" /> Add Survey
        </button>
      </div>
    </div>

    <div v-if="error" class="mb-4 p-3 bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 rounded-lg text-sm">{{ error }}</div>

    <!-- Aggregate KPIs -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'" class="rounded-xl border p-5 shadow-sm text-center">
        <div class="text-3xl font-bold" :class="scoreColor(aggregate.avg_score || 0)">
          {{ (aggregate.avg_score || 0).toFixed(2) }}
        </div>
        <div class="text-xs text-slate-400 mt-1">Avg Score (out of 5)</div>
        <div class="flex justify-center mt-2 gap-0.5">
          <Star v-for="i in 5" :key="i" class="w-4 h-4"
            :class="i <= Math.round(aggregate.avg_score || 0) ? 'text-yellow-400 fill-yellow-400' : 'text-slate-300'" />
        </div>
      </div>
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'" class="rounded-xl border p-5 shadow-sm text-center">
        <div class="text-3xl font-bold text-indigo-600">{{ aggregate.total_responses || 0 }}</div>
        <div class="text-xs text-slate-400 mt-1">Total Responses</div>
      </div>
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'" class="rounded-xl border p-5 shadow-sm text-center">
        <div class="text-3xl font-bold text-green-600">{{ (aggregate.satisfaction_rate || 0).toFixed(1) }}%</div>
        <div class="text-xs text-slate-400 mt-1">Satisfaction Rate</div>
        <div class="text-xs text-slate-400">(Satisfied + Very Satisfied)</div>
      </div>
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'" class="rounded-xl border p-5 shadow-sm text-center">
        <div class="text-3xl font-bold text-red-600">{{ (aggregate.very_dissatisfied || 0) + (aggregate.dissatisfied || 0) }}</div>
        <div class="text-xs text-slate-400 mt-1">Negative Reviews</div>
      </div>
    </div>

    <!-- Rating distribution bar chart -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border p-5 shadow-sm mb-6">
      <h3 class="font-semibold mb-4 flex items-center gap-2">
        <BarChart3 class="w-4 h-4 text-yellow-500" /> Rating Distribution
      </h3>
      <div class="space-y-2">
        <div v-for="(label, rating) in ratingLabels" :key="rating" class="flex items-center gap-3">
          <span class="text-xs w-28 text-right text-slate-500 flex-shrink-0">{{ label }}</span>
          <div class="flex-1 bg-slate-100 dark:bg-slate-700 rounded-full h-4 relative">
            <div :class="ratingColor[rating]"
              :style="`width:${aggregate.total_responses ? ((aggregate[rating.replace('very_','very')] || 0) / aggregate.total_responses * 100) : 0}%`"
              class="h-4 rounded-full transition-all flex items-center justify-end pr-2">
            </div>
          </div>
          <span class="text-sm font-bold w-8 text-right">
            {{ rating === 'very_dissatisfied' ? aggregate.very_dissatisfied :
               rating === 'dissatisfied' ? aggregate.dissatisfied :
               rating === 'neutral' ? aggregate.neutral :
               rating === 'satisfied' ? aggregate.satisfied :
               aggregate.very_satisfied || 0 }}
          </span>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border p-4 mb-4 shadow-sm flex gap-3 flex-wrap">
      <select v-model="filterRating" @change="load"
        :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
        class="px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
        <option value="">All Ratings</option>
        <option v-for="(label, rating) in ratingLabels" :key="rating" :value="rating">{{ label }}</option>
      </select>
      <select v-model="filterAgentId" @change="load"
        :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
        class="px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
        <option value="">All Agents</option>
        <option v-for="a in agents" :key="a.id" :value="a.id">{{ a.name }}</option>
      </select>
    </div>

    <!-- Surveys list -->
    <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
      class="rounded-xl border shadow-sm overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center py-16">
        <RefreshCw class="w-6 h-6 animate-spin text-indigo-500" />
      </div>
      <div v-else-if="!surveys.length" class="text-center py-16">
        <Star class="w-12 h-12 text-slate-300 mx-auto mb-3" />
        <p class="text-slate-400">No CSAT surveys yet</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead>
            <tr :class="app.darkMode ? 'bg-slate-700/50 text-slate-400' : 'bg-slate-50 text-slate-500'"
              class="text-xs uppercase tracking-wide">
              <th class="text-left px-4 py-3 font-medium">Ticket</th>
              <th class="text-left px-4 py-3 font-medium">Rating</th>
              <th class="text-left px-4 py-3 font-medium">Stars</th>
              <th class="text-left px-4 py-3 font-medium">Requester</th>
              <th class="text-left px-4 py-3 font-medium">Agent</th>
              <th class="text-left px-4 py-3 font-medium">Comment</th>
              <th class="text-left px-4 py-3 font-medium">Date</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="s in surveys" :key="s.id"
              :class="app.darkMode ? 'border-slate-700 hover:bg-slate-700/50' : 'border-slate-100 hover:bg-slate-50'"
              class="border-b transition-colors">
              <td class="px-4 py-3 font-mono text-xs text-indigo-600 font-medium">{{ s.ticket_number }}</td>
              <td class="px-4 py-3">
                <span :class="ratingColor[s.rating] || ''" class="text-xs px-2 py-0.5 rounded-full font-medium">
                  {{ ratingLabels[s.rating] || s.rating }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex gap-0.5">
                  <Star v-for="i in 5" :key="i" class="w-3.5 h-3.5"
                    :class="i <= ratingStars(s.rating) ? 'text-yellow-400 fill-yellow-400' : 'text-slate-300'" />
                </div>
              </td>
              <td class="px-4 py-3 text-xs">
                <div>{{ s.requester_name || '-' }}</div>
                <div class="text-slate-400">{{ s.requester_email }}</div>
              </td>
              <td class="px-4 py-3 text-xs">{{ s.agent_name || '-' }}</td>
              <td class="px-4 py-3 text-xs text-slate-500 max-w-[200px] truncate">{{ s.comment || '-' }}</td>
              <td class="px-4 py-3 text-xs text-slate-400">{{ s.submitted_at?.substring(0,10) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Form Modal -->
    <div v-if="showForm" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
      <div :class="app.darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-slate-200'"
        class="w-full max-w-lg rounded-2xl border shadow-xl">
        <div class="flex items-center justify-between p-6 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <h2 class="text-lg font-semibold">Submit CSAT Survey</h2>
          <button @click="showForm = false" class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg"><X class="w-5 h-5" /></button>
        </div>
        <div class="p-6 space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1">Ticket *</label>
            <select v-model="form.ticket_id" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
              <option value="">Select resolved ticket...</option>
              <option v-for="t in tickets" :key="t.id" :value="t.id">{{ t.ticket_number }} — {{ t.subject }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Rating *</label>
            <select v-model="form.rating" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
              <option v-for="(label, rating) in ratingLabels" :key="rating" :value="rating">{{ label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Agent</label>
            <select v-model="form.agent_id" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none">
              <option value="">No specific agent</option>
              <option v-for="a in agents" :key="a.id" :value="a.id">{{ a.name }}</option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium mb-1">Requester Name</label>
              <input v-model="form.requester_name" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Requester Email</label>
              <input v-model="form.requester_email" type="email" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none" />
            </div>
          </div>
          <div>
            <label class="block text-sm font-medium mb-1">Comment</label>
            <textarea v-model="form.comment" rows="3" :class="app.darkMode ? 'bg-slate-700 border-slate-600' : 'bg-white border-slate-200'"
              class="w-full px-3 py-2 border rounded-lg text-sm focus:ring-2 focus:ring-indigo-500 focus:outline-none resize-none" placeholder="Optional feedback..." />
          </div>
        </div>
        <div class="flex justify-end gap-3 p-6 border-t" :class="app.darkMode ? 'border-slate-700' : 'border-slate-200'">
          <button @click="showForm = false" :class="app.darkMode ? 'bg-slate-700 text-slate-200 hover:bg-slate-600' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'"
            class="px-4 py-2 rounded-lg text-sm font-medium">Cancel</button>
          <button @click="submitSurvey" :disabled="saving" class="px-4 py-2 bg-yellow-500 hover:bg-yellow-600 disabled:opacity-50 text-white rounded-lg text-sm font-medium flex items-center gap-2">
            <RefreshCw v-if="saving" class="w-4 h-4 animate-spin" />
            Submit Survey
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
