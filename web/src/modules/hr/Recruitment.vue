<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Briefcase, Users, TrendingUp, CheckCircle, Plus, Search, Filter,
  ChevronDown, ChevronUp, X, Edit2, Trash2, Eye, ArrowRight,
  MapPin, Calendar, Clock, Building2, User, Mail, Phone,
  FileText, DollarSign, ChevronLeft, ChevronRight, MoreHorizontal,
  AlertCircle, RefreshCw, Award, Loader2, UserCheck, UserX
} from 'lucide-vue-next'
import { hrAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

// ─── Types ──────────────────────────────────────────────────────────────────
interface JobPosting {
  id: string
  company_id: string
  department_id: string | null
  position_id: string | null
  title: string
  description: string
  requirements: string
  location: string
  employment_type: string
  vacancies: number
  status: 'draft' | 'open' | 'closed' | 'on_hold'
  published_at: string | null
  deadline_date: string | null
  created_by: string | null
  department_name?: string
  position_title?: string
  application_count?: number
  created_at: string
}

interface Application {
  id: string
  job_posting_id: string
  company_id: string
  first_name: string
  last_name: string
  email: string
  phone: string
  cv_url: string
  cover_letter: string
  source: string
  status: 'new' | 'screening' | 'interview' | 'offer' | 'hired' | 'rejected' | 'withdrawn'
  expected_salary: number | null
  interview_date: string | null
  interview_notes: string
  rejection_reason: string
  hired_as_employee_id: string | null
  job_title?: string
  created_at: string
}

interface Department { id: string; name: string; code: string }
interface Position { id: string; title: string; department_id: string | null }

// ─── State ──────────────────────────────────────────────────────────────────
const activeTab = ref<'jobs' | 'applications'>('jobs')

// Jobs state
const jobs = ref<JobPosting[]>([])
const jobsLoading = ref(false)
const jobsPage = ref(1)
const jobsPerPage = 15
const jobsSearch = ref('')
const jobsStatusFilter = ref('all')
const jobsSortCol = ref('created_at')
const jobsSortDir = ref<'asc' | 'desc'>('desc')

// Applications state
const applications = ref<Application[]>([])
const appsLoading = ref(false)
const appsJobFilter = ref('all')
const appsStatusFilter = ref('all')
const appsSearch = ref('')

// Supporting data
const departments = ref<Department[]>([])
const positions = ref<Position[]>([])

// Modals
const showJobModal = ref(false)
const showDeleteJobConfirm = ref(false)
const showJobDetail = ref(false)
const showAppModal = ref(false)
const showStatusModal = ref(false)

const editingJob = ref<JobPosting | null>(null)
const selectedJob = ref<JobPosting | null>(null)
const deletingJobId = ref<string | null>(null)
const selectedApp = ref<Application | null>(null)

const jobSaving = ref(false)
const appSaving = ref(false)
const statusSaving = ref(false)
const deletingJob = ref(false)

// ─── Job Form ────────────────────────────────────────────────────────────────
const jobForm = ref({
  title: '',
  department_id: '',
  position_id: '',
  employment_type: 'full_time',
  location: '',
  vacancies: 1,
  status: 'draft' as JobPosting['status'],
  published_at: '',
  deadline_date: '',
  description: '',
  requirements: ''
})

// ─── Application Form ────────────────────────────────────────────────────────
const appForm = ref({
  job_posting_id: '',
  first_name: '',
  last_name: '',
  email: '',
  phone: '',
  source: 'direct',
  expected_salary: '' as string | number,
  cover_letter: '',
  cv_url: ''
})

// ─── Status Update Form ───────────────────────────────────────────────────────
const statusForm = ref({
  status: '' as Application['status'],
  interview_date: '',
  interview_notes: '',
  rejection_reason: ''
})

// ─── Computed — KPI ──────────────────────────────────────────────────────────
const kpiTotalJobs = computed(() => jobs.value.length)
const kpiOpenJobs = computed(() => jobs.value.filter(j => j.status === 'open').length)
const kpiTotalVacancies = computed(() => jobs.value.filter(j => j.status === 'open').reduce((s, j) => s + (j.vacancies || 0), 0))
const kpiTotalApps = computed(() => applications.value.length)
const kpiHired = computed(() => applications.value.filter(a => a.status === 'hired').length)

// ─── Computed — Jobs table ────────────────────────────────────────────────────
const filteredJobs = computed(() => {
  let list = [...jobs.value]
  if (jobsStatusFilter.value !== 'all') list = list.filter(j => j.status === jobsStatusFilter.value)
  if (jobsSearch.value.trim()) {
    const q = jobsSearch.value.toLowerCase()
    list = list.filter(j =>
      j.title.toLowerCase().includes(q) ||
      (j.department_name || '').toLowerCase().includes(q) ||
      (j.location || '').toLowerCase().includes(q)
    )
  }
  list.sort((a, b) => {
    let av: string | number = (a as any)[jobsSortCol.value] ?? ''
    let bv: string | number = (b as any)[jobsSortCol.value] ?? ''
    if (typeof av === 'string') av = av.toLowerCase()
    if (typeof bv === 'string') bv = bv.toLowerCase()
    if (av < bv) return jobsSortDir.value === 'asc' ? -1 : 1
    if (av > bv) return jobsSortDir.value === 'asc' ? 1 : -1
    return 0
  })
  return list
})

const jobsPageCount = computed(() => Math.max(1, Math.ceil(filteredJobs.value.length / jobsPerPage)))
const pagedJobs = computed(() => {
  const start = (jobsPage.value - 1) * jobsPerPage
  return filteredJobs.value.slice(start, start + jobsPerPage)
})

// ─── Computed — Applications table / pipeline ────────────────────────────────
const filteredApps = computed(() => {
  let list = [...applications.value]
  if (appsJobFilter.value !== 'all') list = list.filter(a => a.job_posting_id === appsJobFilter.value)
  if (appsStatusFilter.value !== 'all') list = list.filter(a => a.status === appsStatusFilter.value)
  if (appsSearch.value.trim()) {
    const q = appsSearch.value.toLowerCase()
    list = list.filter(a =>
      `${a.first_name} ${a.last_name}`.toLowerCase().includes(q) ||
      a.email.toLowerCase().includes(q) ||
      (a.job_title || '').toLowerCase().includes(q)
    )
  }
  return list
})

const appsByStatus = computed(() => {
  const stages: Application['status'][] = ['new', 'screening', 'interview', 'offer', 'hired', 'rejected']
  const result: Record<string, Application[]> = {}
  for (const s of stages) result[s] = filteredApps.value.filter(a => a.status === s)
  return result
})

const filteredPositions = computed(() => {
  if (!jobForm.value.department_id) return positions.value
  return positions.value.filter(p => !p.department_id || p.department_id === jobForm.value.department_id)
})

// ─── Helpers ─────────────────────────────────────────────────────────────────
function sortJobs(col: string) {
  if (jobsSortCol.value === col) {
    jobsSortDir.value = jobsSortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    jobsSortCol.value = col
    jobsSortDir.value = 'asc'
  }
  jobsPage.value = 1
}

const jobStatusBadge: Record<string, { label: string; classes: string }> = {
  draft:   { label: 'Draft',   classes: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300' },
  open:    { label: 'Open',    classes: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-400' },
  closed:  { label: 'Closed',  classes: 'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-400' },
  on_hold: { label: 'On Hold', classes: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-400' }
}

const appStatusConfig: Record<string, { label: string; color: string; dot: string }> = {
  new:        { label: 'New',        color: 'bg-sky-100 text-sky-700 dark:bg-sky-900/40 dark:text-sky-400',         dot: 'bg-sky-500' },
  screening:  { label: 'Screening',  color: 'bg-violet-100 text-violet-700 dark:bg-violet-900/40 dark:text-violet-400', dot: 'bg-violet-500' },
  interview:  { label: 'Interview',  color: 'bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-400', dot: 'bg-amber-500' },
  offer:      { label: 'Offer',      color: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/40 dark:text-indigo-400', dot: 'bg-indigo-500' },
  hired:      { label: 'Hired',      color: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-400', dot: 'bg-emerald-500' },
  rejected:   { label: 'Rejected',   color: 'bg-rose-100 text-rose-700 dark:bg-rose-900/40 dark:text-rose-400',   dot: 'bg-rose-500' },
  withdrawn:  { label: 'Withdrawn',  color: 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300',   dot: 'bg-slate-400' }
}

const employmentTypeLabel: Record<string, string> = {
  full_time:  'Full-Time',
  part_time:  'Part-Time',
  contract:   'Contract',
  intern:     'Intern',
  consultant: 'Consultant',
  temporary:  'Temporary'
}

const sourceLabel: Record<string, string> = {
  direct:    'Direct',
  linkedin:  'LinkedIn',
  indeed:    'Indeed',
  referral:  'Referral',
  website:   'Website',
  agency:    'Agency',
  other:     'Other'
}

function fmtDate(d: string | null) {
  if (!d) return '—'
  return new Date(d).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })
}

function fmtSalary(v: number | null) {
  if (!v) return '—'
  return Number(v).toLocaleString('fr-DZ') + ' DZD'
}

function applicationsForJob(jobId: string) {
  return applications.value.filter(a => a.job_posting_id === jobId)
}

function isDeadlinePast(d: string | null) {
  if (!d) return false
  return new Date(d) < new Date()
}

// ─── Data Loading ────────────────────────────────────────────────────────────
async function loadJobs() {
  jobsLoading.value = true
  try {
    const r = await hrAPI.getJobPostings()
    jobs.value = r.data || []
  } catch {
    appStore.addToast('Failed to load job postings', 'error')
  } finally {
    jobsLoading.value = false
  }
}

async function loadApplications() {
  appsLoading.value = true
  try {
    const r = await hrAPI.getApplications()
    applications.value = r.data || []
  } catch {
    appStore.addToast('Failed to load applications', 'error')
  } finally {
    appsLoading.value = false
  }
}

async function loadDepartments() {
  try {
    const r = await hrAPI.getDepartments()
    departments.value = r.data || []
  } catch { /* silent */ }
}

async function loadPositions() {
  try {
    const r = await hrAPI.getPositions()
    positions.value = r.data || []
  } catch { /* silent */ }
}

// ─── Job CRUD ─────────────────────────────────────────────────────────────────
function openCreateJob() {
  editingJob.value = null
  jobForm.value = {
    title: '', department_id: '', position_id: '', employment_type: 'full_time',
    location: '', vacancies: 1, status: 'draft', published_at: '', deadline_date: '',
    description: '', requirements: ''
  }
  showJobModal.value = true
}

function openEditJob(job: JobPosting) {
  editingJob.value = job
  jobForm.value = {
    title: job.title,
    department_id: job.department_id || '',
    position_id: job.position_id || '',
    employment_type: job.employment_type || 'full_time',
    location: job.location || '',
    vacancies: job.vacancies || 1,
    status: job.status,
    published_at: job.published_at ? job.published_at.substring(0, 10) : '',
    deadline_date: job.deadline_date ? job.deadline_date.substring(0, 10) : '',
    description: job.description || '',
    requirements: job.requirements || ''
  }
  showJobModal.value = true
}

function openJobDetail(job: JobPosting) {
  selectedJob.value = job
  showJobDetail.value = true
}

function confirmDeleteJob(id: string) {
  deletingJobId.value = id
  showDeleteJobConfirm.value = true
}

async function saveJob() {
  if (!jobForm.value.title.trim()) {
    appStore.addToast('Job title is required', 'error')
    return
  }
  jobSaving.value = true
  try {
    const payload: any = {
      title: jobForm.value.title.trim(),
      department_id: jobForm.value.department_id || null,
      position_id: jobForm.value.position_id || null,
      employment_type: jobForm.value.employment_type,
      location: jobForm.value.location.trim(),
      vacancies: Number(jobForm.value.vacancies) || 1,
      status: jobForm.value.status,
      published_at: jobForm.value.published_at || null,
      deadline_date: jobForm.value.deadline_date || null,
      description: jobForm.value.description.trim(),
      requirements: jobForm.value.requirements.trim()
    }
    if (editingJob.value) {
      await hrAPI.updateJobPosting(editingJob.value.id, payload)
      appStore.addToast('Job posting updated', 'success')
    } else {
      await hrAPI.createJobPosting(payload)
      appStore.addToast('Job posting created', 'success')
    }
    showJobModal.value = false
    await loadJobs()
  } catch (e: any) {
    appStore.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally {
    jobSaving.value = false
  }
}

async function deleteJob() {
  if (!deletingJobId.value) return
  deletingJob.value = true
  try {
    await hrAPI.deleteJobPosting(deletingJobId.value)
    appStore.addToast('Job posting deleted', 'success')
    showDeleteJobConfirm.value = false
    deletingJobId.value = null
    await loadJobs()
  } catch (e: any) {
    appStore.addToast(e?.response?.data?.error || 'Delete failed', 'error')
  } finally {
    deletingJob.value = false
  }
}

async function quickStatusJob(job: JobPosting, status: JobPosting['status']) {
  try {
    await hrAPI.updateJobPosting(job.id, { ...job, status, department_id: job.department_id || null, position_id: job.position_id || null })
    appStore.addToast(`Job ${status === 'open' ? 'published' : status}`, 'success')
    await loadJobs()
  } catch (e: any) {
    appStore.addToast(e?.response?.data?.error || 'Update failed', 'error')
  }
}

// ─── Application CRUD ─────────────────────────────────────────────────────────
function openCreateApp(jobId?: string) {
  appForm.value = {
    job_posting_id: jobId || (jobs.value.find(j => j.status === 'open')?.id || ''),
    first_name: '', last_name: '', email: '', phone: '',
    source: 'direct', expected_salary: '', cover_letter: '', cv_url: ''
  }
  showAppModal.value = true
}

async function saveApplication() {
  if (!appForm.value.first_name.trim() || !appForm.value.last_name.trim() || !appForm.value.email.trim()) {
    appStore.addToast('First name, last name and email are required', 'error')
    return
  }
  if (!appForm.value.job_posting_id) {
    appStore.addToast('Please select a job posting', 'error')
    return
  }
  appSaving.value = true
  try {
    const payload = {
      job_posting_id: appForm.value.job_posting_id,
      first_name: appForm.value.first_name.trim(),
      last_name: appForm.value.last_name.trim(),
      email: appForm.value.email.trim(),
      phone: appForm.value.phone.trim(),
      source: appForm.value.source,
      expected_salary: appForm.value.expected_salary ? Number(appForm.value.expected_salary) : null,
      cover_letter: appForm.value.cover_letter.trim(),
      cv_url: appForm.value.cv_url.trim()
    }
    await hrAPI.createApplication(payload)
    appStore.addToast('Application submitted', 'success')
    showAppModal.value = false
    await loadApplications()
  } catch (e: any) {
    appStore.addToast(e?.response?.data?.error || 'Submit failed', 'error')
  } finally {
    appSaving.value = false
  }
}

function openStatusUpdate(app: Application) {
  selectedApp.value = app
  statusForm.value = {
    status: app.status,
    interview_date: app.interview_date ? app.interview_date.substring(0, 16) : '',
    interview_notes: app.interview_notes || '',
    rejection_reason: app.rejection_reason || ''
  }
  showStatusModal.value = true
}

async function saveStatus() {
  if (!selectedApp.value) return
  statusSaving.value = true
  try {
    const payload: any = { status: statusForm.value.status }
    if (statusForm.value.interview_date) payload.interview_date = statusForm.value.interview_date
    if (statusForm.value.interview_notes) payload.interview_notes = statusForm.value.interview_notes
    if (statusForm.value.rejection_reason) payload.rejection_reason = statusForm.value.rejection_reason
    await hrAPI.updateApplicationStatus(selectedApp.value.id, payload)
    appStore.addToast('Application status updated', 'success')
    showStatusModal.value = false
    await loadApplications()
  } catch (e: any) {
    appStore.addToast(e?.response?.data?.error || 'Update failed', 'error')
  } finally {
    statusSaving.value = false
  }
}

// ─── Init ─────────────────────────────────────────────────────────────────────
onMounted(async () => {
  await Promise.all([loadJobs(), loadApplications(), loadDepartments(), loadPositions()])
})
</script>

<template>
  <div class="space-y-6">

    <!-- ── Page Header ─────────────────────────────────────────────────── -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">Recruitment</h1>
        <p class="mt-0.5 text-sm text-slate-500 dark:text-slate-400">Manage job postings and candidate pipeline</p>
      </div>
      <div class="flex items-center gap-2">
        <button
          @click="openCreateApp()"
          class="inline-flex items-center gap-2 px-4 py-2 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
        >
          <User class="w-4 h-4" />
          Add Application
        </button>
        <button
          @click="openCreateJob()"
          class="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg text-sm font-semibold transition-colors shadow-sm"
        >
          <Plus class="w-4 h-4" />
          Post Job
        </button>
      </div>
    </div>

    <!-- ── KPI Cards ──────────────────────────────────────────────────── -->
    <div class="grid grid-cols-2 lg:grid-cols-5 gap-4">
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total Jobs</span>
          <div class="w-8 h-8 bg-slate-100 dark:bg-slate-700 rounded-lg flex items-center justify-center">
            <Briefcase class="w-4 h-4 text-slate-600 dark:text-slate-300" />
          </div>
        </div>
        <p class="text-2xl font-bold text-slate-900 dark:text-slate-100">{{ kpiTotalJobs }}</p>
        <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">All postings</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Open Jobs</span>
          <div class="w-8 h-8 bg-emerald-100 dark:bg-emerald-900/40 rounded-lg flex items-center justify-center">
            <TrendingUp class="w-4 h-4 text-emerald-600 dark:text-emerald-400" />
          </div>
        </div>
        <p class="text-2xl font-bold text-emerald-600 dark:text-emerald-400">{{ kpiOpenJobs }}</p>
        <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">Active listings</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Open Vacancies</span>
          <div class="w-8 h-8 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg flex items-center justify-center">
            <Award class="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
          </div>
        </div>
        <p class="text-2xl font-bold text-indigo-600 dark:text-indigo-400">{{ kpiTotalVacancies }}</p>
        <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">Positions available</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Applications</span>
          <div class="w-8 h-8 bg-amber-100 dark:bg-amber-900/40 rounded-lg flex items-center justify-center">
            <Users class="w-4 h-4 text-amber-600 dark:text-amber-400" />
          </div>
        </div>
        <p class="text-2xl font-bold text-amber-600 dark:text-amber-400">{{ kpiTotalApps }}</p>
        <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">Total candidates</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Hired</span>
          <div class="w-8 h-8 bg-rose-100 dark:bg-rose-900/40 rounded-lg flex items-center justify-center">
            <CheckCircle class="w-4 h-4 text-rose-600 dark:text-rose-400" />
          </div>
        </div>
        <p class="text-2xl font-bold text-rose-600 dark:text-rose-400">{{ kpiHired }}</p>
        <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">Successfully hired</p>
      </div>
    </div>

    <!-- ── Tab Switcher ────────────────────────────────────────────────── -->
    <div class="flex items-center gap-1 bg-slate-100 dark:bg-slate-800 rounded-xl p-1 w-fit">
      <button
        @click="activeTab = 'jobs'"
        :class="[
          'px-4 py-2 rounded-lg text-sm font-medium transition-all',
          activeTab === 'jobs'
            ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm'
            : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300'
        ]"
      >
        <div class="flex items-center gap-2">
          <Briefcase class="w-4 h-4" />
          Job Postings
          <span class="bg-slate-200 dark:bg-slate-600 text-slate-600 dark:text-slate-300 text-xs px-1.5 py-0.5 rounded-full font-semibold">{{ kpiTotalJobs }}</span>
        </div>
      </button>
      <button
        @click="activeTab = 'applications'"
        :class="[
          'px-4 py-2 rounded-lg text-sm font-medium transition-all',
          activeTab === 'applications'
            ? 'bg-white dark:bg-slate-700 text-slate-900 dark:text-slate-100 shadow-sm'
            : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300'
        ]"
      >
        <div class="flex items-center gap-2">
          <Users class="w-4 h-4" />
          Applications
          <span class="bg-slate-200 dark:bg-slate-600 text-slate-600 dark:text-slate-300 text-xs px-1.5 py-0.5 rounded-full font-semibold">{{ kpiTotalApps }}</span>
        </div>
      </button>
    </div>

    <!-- ══════════════════════════════════════════════════════════════════ -->
    <!-- TAB: JOB POSTINGS                                                  -->
    <!-- ══════════════════════════════════════════════════════════════════ -->
    <template v-if="activeTab === 'jobs'">

      <!-- Filters bar -->
      <div class="flex flex-col sm:flex-row items-start sm:items-center gap-3">
        <div class="relative flex-1 min-w-0">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="jobsSearch"
            type="text"
            placeholder="Search jobs by title, department or location..."
            class="w-full pl-9 pr-4 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            @input="jobsPage = 1"
          />
        </div>
        <div class="flex items-center gap-2 flex-shrink-0">
          <div class="flex items-center gap-1 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg p-1">
            <button v-for="s in ['all','draft','open','on_hold','closed']" :key="s"
              @click="jobsStatusFilter = s; jobsPage = 1"
              :class="[
                'px-3 py-1 rounded-md text-xs font-medium transition-all capitalize',
                jobsStatusFilter === s
                  ? 'bg-indigo-600 text-white shadow-sm'
                  : 'text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300'
              ]"
            >{{ s === 'all' ? 'All' : s === 'on_hold' ? 'On Hold' : s.charAt(0).toUpperCase() + s.slice(1) }}</button>
          </div>
          <button @click="loadJobs()" class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-500 hover:text-indigo-600 transition-colors">
            <RefreshCw class="w-4 h-4" :class="jobsLoading && 'animate-spin'" />
          </button>
        </div>
      </div>

      <!-- Jobs Table -->
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm">
        <div v-if="jobsLoading" class="flex items-center justify-center py-20">
          <Loader2 class="w-8 h-8 animate-spin text-indigo-500" />
        </div>
        <table v-else class="w-full">
          <thead class="bg-slate-50 dark:bg-slate-700/50 border-b border-slate-200 dark:border-slate-700">
            <tr>
              <th class="text-left px-4 py-3">
                <button @click="sortJobs('title')" class="flex items-center gap-1 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hover:text-slate-700 dark:hover:text-slate-200">
                  Job Title
                  <component :is="jobsSortCol === 'title' && jobsSortDir === 'asc' ? ChevronUp : ChevronDown" class="w-3 h-3" />
                </button>
              </th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hidden md:table-cell">Department</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hidden lg:table-cell">Type</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hidden lg:table-cell">Location</th>
              <th class="text-center px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Vacancies</th>
              <th class="text-center px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hidden md:table-cell">Applications</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hidden xl:table-cell">Deadline</th>
              <th class="text-center px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Status</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="pagedJobs.length === 0">
              <td colspan="9" class="px-4 py-16 text-center">
                <Briefcase class="w-10 h-10 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
                <p class="text-sm font-medium text-slate-500 dark:text-slate-400">No job postings found</p>
                <p class="text-xs text-slate-400 dark:text-slate-500 mt-1">Create your first job posting to start recruiting</p>
              </td>
            </tr>
            <tr v-for="job in pagedJobs" :key="job.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
              <!-- Title + position -->
              <td class="px-4 py-3">
                <div class="flex items-center gap-3">
                  <div class="w-9 h-9 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg flex items-center justify-center flex-shrink-0">
                    <Briefcase class="w-4 h-4 text-indigo-600 dark:text-indigo-400" />
                  </div>
                  <div>
                    <p class="text-sm font-semibold text-slate-800 dark:text-slate-100">{{ job.title }}</p>
                    <p class="text-xs text-slate-400 dark:text-slate-500">{{ job.position_title || '—' }}</p>
                  </div>
                </div>
              </td>
              <!-- Department -->
              <td class="px-4 py-3 hidden md:table-cell">
                <span class="text-sm text-slate-600 dark:text-slate-300">{{ job.department_name || '—' }}</span>
              </td>
              <!-- Type -->
              <td class="px-4 py-3 hidden lg:table-cell">
                <span class="text-xs font-medium text-slate-600 dark:text-slate-300 bg-slate-100 dark:bg-slate-700 px-2 py-1 rounded">
                  {{ employmentTypeLabel[job.employment_type] || job.employment_type }}
                </span>
              </td>
              <!-- Location -->
              <td class="px-4 py-3 hidden lg:table-cell">
                <div class="flex items-center gap-1 text-sm text-slate-600 dark:text-slate-300">
                  <MapPin class="w-3.5 h-3.5 text-slate-400 flex-shrink-0" />
                  {{ job.location || '—' }}
                </div>
              </td>
              <!-- Vacancies -->
              <td class="px-4 py-3 text-center">
                <span class="text-sm font-semibold text-slate-800 dark:text-slate-100">{{ job.vacancies }}</span>
              </td>
              <!-- Applications count -->
              <td class="px-4 py-3 text-center hidden md:table-cell">
                <button
                  @click="activeTab = 'applications'; appsJobFilter = job.id"
                  class="inline-flex items-center gap-1 text-sm font-semibold text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 transition-colors"
                >
                  <Users class="w-3.5 h-3.5" />
                  {{ applicationsForJob(job.id).length }}
                </button>
              </td>
              <!-- Deadline -->
              <td class="px-4 py-3 hidden xl:table-cell">
                <span v-if="job.deadline_date" :class="['text-xs font-medium', isDeadlinePast(job.deadline_date) ? 'text-rose-600 dark:text-rose-400' : 'text-slate-600 dark:text-slate-300']">
                  {{ fmtDate(job.deadline_date) }}
                  <span v-if="isDeadlinePast(job.deadline_date)" class="ml-1 text-rose-500">(expired)</span>
                </span>
                <span v-else class="text-xs text-slate-400">—</span>
              </td>
              <!-- Status badge -->
              <td class="px-4 py-3 text-center">
                <span :class="['inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold', jobStatusBadge[job.status]?.classes]">
                  {{ jobStatusBadge[job.status]?.label }}
                </span>
              </td>
              <!-- Actions -->
              <td class="px-4 py-3">
                <div class="flex items-center justify-end gap-1">
                  <button @click="openJobDetail(job)" title="View" class="p-1.5 text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
                    <Eye class="w-4 h-4" />
                  </button>
                  <button @click="openEditJob(job)" title="Edit" class="p-1.5 text-slate-400 hover:text-amber-600 dark:hover:text-amber-400 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
                    <Edit2 class="w-4 h-4" />
                  </button>
                  <button v-if="job.status !== 'open'" @click="quickStatusJob(job, 'open')" title="Publish" class="p-1.5 text-slate-400 hover:text-emerald-600 dark:hover:text-emerald-400 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
                    <ArrowRight class="w-4 h-4" />
                  </button>
                  <button v-if="job.status === 'open'" @click="quickStatusJob(job, 'closed')" title="Close" class="p-1.5 text-slate-400 hover:text-rose-600 dark:hover:text-rose-400 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
                    <X class="w-4 h-4" />
                  </button>
                  <button @click="confirmDeleteJob(job.id)" title="Delete" class="p-1.5 text-slate-400 hover:text-rose-600 dark:hover:text-rose-400 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Pagination -->
        <div v-if="filteredJobs.length > jobsPerPage" class="flex items-center justify-between px-4 py-3 border-t border-slate-100 dark:border-slate-700 bg-slate-50 dark:bg-slate-700/30">
          <p class="text-xs text-slate-500 dark:text-slate-400">
            Showing {{ (jobsPage - 1) * jobsPerPage + 1 }}–{{ Math.min(jobsPage * jobsPerPage, filteredJobs.length) }} of {{ filteredJobs.length }}
          </p>
          <div class="flex items-center gap-1">
            <button @click="jobsPage--" :disabled="jobsPage === 1"
              class="p-1.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-500 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
              <ChevronLeft class="w-4 h-4" />
            </button>
            <span class="px-3 py-1 text-xs font-medium text-slate-700 dark:text-slate-300">{{ jobsPage }} / {{ jobsPageCount }}</span>
            <button @click="jobsPage++" :disabled="jobsPage === jobsPageCount"
              class="p-1.5 rounded-lg border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-500 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
              <ChevronRight class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </template>

    <!-- ══════════════════════════════════════════════════════════════════ -->
    <!-- TAB: APPLICATIONS PIPELINE                                         -->
    <!-- ══════════════════════════════════════════════════════════════════ -->
    <template v-if="activeTab === 'applications'">

      <!-- Filters bar -->
      <div class="flex flex-col sm:flex-row items-start sm:items-center gap-3">
        <div class="relative flex-1 min-w-0">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
          <input
            v-model="appsSearch"
            type="text"
            placeholder="Search by name, email or job title..."
            class="w-full pl-9 pr-4 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
          />
        </div>
        <div class="flex items-center gap-2 flex-shrink-0 flex-wrap">
          <!-- Job filter -->
          <div class="relative">
            <select
              v-model="appsJobFilter"
              class="appearance-none pl-3 pr-8 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="all">All Jobs</option>
              <option v-for="j in jobs" :key="j.id" :value="j.id">{{ j.title }}</option>
            </select>
            <ChevronDown class="absolute right-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400 pointer-events-none" />
          </div>
          <!-- Status filter -->
          <div class="relative">
            <select
              v-model="appsStatusFilter"
              class="appearance-none pl-3 pr-8 py-2 text-sm bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-700 dark:text-slate-300 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
              <option value="all">All Stages</option>
              <option v-for="(cfg, key) in appStatusConfig" :key="key" :value="key">{{ cfg.label }}</option>
            </select>
            <ChevronDown class="absolute right-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400 pointer-events-none" />
          </div>
          <button @click="loadApplications()" class="p-2 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-500 hover:text-indigo-600 transition-colors">
            <RefreshCw class="w-4 h-4" :class="appsLoading && 'animate-spin'" />
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="appsLoading" class="flex items-center justify-center py-20">
        <Loader2 class="w-8 h-8 animate-spin text-indigo-500" />
      </div>

      <!-- Kanban Pipeline (shown when no status filter) -->
      <div v-else-if="appsStatusFilter === 'all' && appsSearch === '' && appsJobFilter !== 'all' || (appsStatusFilter === 'all' && appsSearch === '')" class="overflow-x-auto pb-2">
        <div class="flex gap-4 min-w-max">
          <div
            v-for="stage in ['new','screening','interview','offer','hired','rejected']"
            :key="stage"
            class="w-72 flex-shrink-0"
          >
            <!-- Stage header -->
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2">
                <span :class="['w-2.5 h-2.5 rounded-full flex-shrink-0', appStatusConfig[stage]?.dot]"></span>
                <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">{{ appStatusConfig[stage]?.label }}</span>
              </div>
              <span class="text-xs font-semibold text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-700 px-2 py-0.5 rounded-full">
                {{ (appsByStatus[stage] || []).length }}
              </span>
            </div>

            <!-- Cards -->
            <div class="space-y-2 min-h-24">
              <div
                v-for="app in (appsByStatus[stage] || [])"
                :key="app.id"
                class="bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl p-3 shadow-sm hover:shadow-md hover:border-indigo-300 dark:hover:border-indigo-600 transition-all cursor-pointer group"
                @click="openStatusUpdate(app)"
              >
                <!-- Candidate name -->
                <div class="flex items-center gap-2 mb-2">
                  <div class="w-7 h-7 rounded-full bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center text-xs font-bold text-indigo-600 dark:text-indigo-400 flex-shrink-0">
                    {{ app.first_name.charAt(0) }}{{ app.last_name.charAt(0) }}
                  </div>
                  <div class="min-w-0">
                    <p class="text-sm font-semibold text-slate-800 dark:text-slate-100 truncate">{{ app.first_name }} {{ app.last_name }}</p>
                    <p class="text-xs text-slate-400 dark:text-slate-500 truncate">{{ app.job_title || '—' }}</p>
                  </div>
                </div>
                <!-- Meta row -->
                <div class="flex items-center justify-between text-xs text-slate-400 dark:text-slate-500">
                  <div class="flex items-center gap-1">
                    <Calendar class="w-3 h-3" />
                    {{ fmtDate(app.created_at) }}
                  </div>
                  <div class="flex items-center gap-1">
                    <DollarSign class="w-3 h-3" />
                    {{ app.expected_salary ? (Number(app.expected_salary) / 1000).toFixed(0) + 'K' : '—' }}
                  </div>
                </div>
                <!-- Source tag -->
                <div class="mt-2 flex items-center gap-1">
                  <span class="text-xs bg-slate-100 dark:bg-slate-700 text-slate-500 dark:text-slate-400 px-1.5 py-0.5 rounded font-medium">
                    {{ sourceLabel[app.source] || app.source }}
                  </span>
                  <span v-if="app.interview_date" class="text-xs bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-400 px-1.5 py-0.5 rounded font-medium flex items-center gap-0.5">
                    <Clock class="w-2.5 h-2.5" />
                    {{ fmtDate(app.interview_date) }}
                  </span>
                </div>
              </div>

              <!-- Empty stage -->
              <div v-if="!(appsByStatus[stage] || []).length" class="border-2 border-dashed border-slate-200 dark:border-slate-700 rounded-xl h-20 flex items-center justify-center">
                <p class="text-xs text-slate-400 dark:text-slate-500">No candidates</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- List view (when filtered by status or search) -->
      <div v-else class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden shadow-sm">
        <table class="w-full">
          <thead class="bg-slate-50 dark:bg-slate-700/50 border-b border-slate-200 dark:border-slate-700">
            <tr>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Candidate</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hidden md:table-cell">Job Posting</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hidden lg:table-cell">Source</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hidden lg:table-cell">Expected Salary</th>
              <th class="text-left px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide hidden xl:table-cell">Applied</th>
              <th class="text-center px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Stage</th>
              <th class="text-right px-4 py-3 text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="filteredApps.length === 0">
              <td colspan="7" class="px-4 py-16 text-center">
                <Users class="w-10 h-10 text-slate-300 dark:text-slate-600 mx-auto mb-3" />
                <p class="text-sm font-medium text-slate-500 dark:text-slate-400">No applications found</p>
              </td>
            </tr>
            <tr v-for="app in filteredApps" :key="app.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
              <!-- Candidate -->
              <td class="px-4 py-3">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center text-xs font-bold text-indigo-600 dark:text-indigo-400 flex-shrink-0">
                    {{ app.first_name.charAt(0) }}{{ app.last_name.charAt(0) }}
                  </div>
                  <div>
                    <p class="text-sm font-semibold text-slate-800 dark:text-slate-100">{{ app.first_name }} {{ app.last_name }}</p>
                    <p class="text-xs text-slate-400 dark:text-slate-500">{{ app.email }}</p>
                  </div>
                </div>
              </td>
              <!-- Job -->
              <td class="px-4 py-3 hidden md:table-cell">
                <span class="text-sm text-slate-600 dark:text-slate-300">{{ app.job_title || '—' }}</span>
              </td>
              <!-- Source -->
              <td class="px-4 py-3 hidden lg:table-cell">
                <span class="text-xs font-medium text-slate-600 dark:text-slate-300 bg-slate-100 dark:bg-slate-700 px-2 py-1 rounded">
                  {{ sourceLabel[app.source] || app.source }}
                </span>
              </td>
              <!-- Salary -->
              <td class="px-4 py-3 text-right hidden lg:table-cell">
                <span class="text-sm font-mono text-slate-700 dark:text-slate-300">{{ fmtSalary(app.expected_salary) }}</span>
              </td>
              <!-- Applied date -->
              <td class="px-4 py-3 hidden xl:table-cell">
                <span class="text-xs text-slate-500 dark:text-slate-400">{{ fmtDate(app.created_at) }}</span>
              </td>
              <!-- Stage badge -->
              <td class="px-4 py-3 text-center">
                <span :class="['inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold', appStatusConfig[app.status]?.color]">
                  <span :class="['w-1.5 h-1.5 rounded-full', appStatusConfig[app.status]?.dot]"></span>
                  {{ appStatusConfig[app.status]?.label }}
                </span>
              </td>
              <!-- Actions -->
              <td class="px-4 py-3">
                <div class="flex items-center justify-end gap-1">
                  <button @click="openStatusUpdate(app)" title="Update Stage" class="p-1.5 text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
                    <ArrowRight class="w-4 h-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>


    <!-- ══════════════════════════════════════════════════════════════════ -->
    <!-- MODAL: Create / Edit Job Posting                                   -->
    <!-- ══════════════════════════════════════════════════════════════════ -->
    <Teleport to="body">
      <div v-if="showJobModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showJobModal = false"></div>
        <div class="relative bg-white dark:bg-slate-900 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">

          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700 flex-shrink-0">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 bg-indigo-100 dark:bg-indigo-900/40 rounded-lg flex items-center justify-center">
                <Briefcase class="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
              </div>
              <div>
                <h2 class="text-base font-bold text-slate-900 dark:text-slate-100">
                  {{ editingJob ? 'Edit Job Posting' : 'New Job Posting' }}
                </h2>
                <p class="text-xs text-slate-500 dark:text-slate-400">Fill in the position details below</p>
              </div>
            </div>
            <button @click="showJobModal = false" class="p-2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Body -->
          <div class="overflow-y-auto flex-1 px-6 py-5 space-y-5">

            <!-- Row: Title + Status -->
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div class="sm:col-span-2">
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Job Title <span class="text-rose-500">*</span></label>
                <input v-model="jobForm.title" type="text" placeholder="e.g. Senior Software Engineer"
                  class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Status</label>
                <div class="relative">
                  <select v-model="jobForm.status"
                    class="w-full appearance-none pl-3 pr-8 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500">
                    <option value="draft">Draft</option>
                    <option value="open">Open</option>
                    <option value="on_hold">On Hold</option>
                    <option value="closed">Closed</option>
                  </select>
                  <ChevronDown class="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
                </div>
              </div>
            </div>

            <!-- Row: Department + Position -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Department</label>
                <div class="relative">
                  <select v-model="jobForm.department_id"
                    class="w-full appearance-none pl-3 pr-8 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500">
                    <option value="">No Department</option>
                    <option v-for="d in departments" :key="d.id" :value="d.id">{{ d.name }}</option>
                  </select>
                  <ChevronDown class="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
                </div>
              </div>
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Position</label>
                <div class="relative">
                  <select v-model="jobForm.position_id"
                    class="w-full appearance-none pl-3 pr-8 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500">
                    <option value="">No Position</option>
                    <option v-for="p in filteredPositions" :key="p.id" :value="p.id">{{ p.title }}</option>
                  </select>
                  <ChevronDown class="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
                </div>
              </div>
            </div>

            <!-- Row: Employment Type + Location + Vacancies -->
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Employment Type</label>
                <div class="relative">
                  <select v-model="jobForm.employment_type"
                    class="w-full appearance-none pl-3 pr-8 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500">
                    <option v-for="(lbl, val) in employmentTypeLabel" :key="val" :value="val">{{ lbl }}</option>
                  </select>
                  <ChevronDown class="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
                </div>
              </div>
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Location</label>
                <input v-model="jobForm.location" type="text" placeholder="e.g. Algiers, Remote"
                  class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Vacancies</label>
                <input v-model.number="jobForm.vacancies" type="number" min="1"
                  class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
            </div>

            <!-- Row: Published + Deadline -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Published Date</label>
                <input v-model="jobForm.published_at" type="date"
                  class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Application Deadline</label>
                <input v-model="jobForm.deadline_date" type="date"
                  class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
            </div>

            <!-- Description -->
            <div>
              <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Job Description</label>
              <textarea v-model="jobForm.description" rows="4" placeholder="Describe the role, responsibilities and team..."
                class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"></textarea>
            </div>

            <!-- Requirements -->
            <div>
              <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Requirements & Qualifications</label>
              <textarea v-model="jobForm.requirements" rows="4" placeholder="List skills, experience, and education requirements..."
                class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"></textarea>
            </div>
          </div>

          <!-- Footer -->
          <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/50 flex-shrink-0">
            <button @click="showJobModal = false"
              class="px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
              Cancel
            </button>
            <button @click="saveJob()" :disabled="jobSaving"
              class="px-5 py-2 text-sm font-semibold text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 disabled:cursor-not-allowed rounded-lg transition-colors flex items-center gap-2">
              <Loader2 v-if="jobSaving" class="w-4 h-4 animate-spin" />
              {{ editingJob ? 'Save Changes' : 'Create Posting' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>


    <!-- ══════════════════════════════════════════════════════════════════ -->
    <!-- DRAWER: Job Detail                                                 -->
    <!-- ══════════════════════════════════════════════════════════════════ -->
    <Teleport to="body">
      <div v-if="showJobDetail && selectedJob" class="fixed inset-0 z-50 flex justify-end">
        <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="showJobDetail = false"></div>
        <div class="relative bg-white dark:bg-slate-900 w-full max-w-lg h-full overflow-y-auto shadow-2xl flex flex-col">

          <!-- Header -->
          <div class="bg-gradient-to-r from-indigo-600 to-violet-600 px-6 py-5 flex-shrink-0">
            <div class="flex items-start justify-between">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2 mb-2">
                  <span :class="['inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-white/20 text-white']">
                    {{ jobStatusBadge[selectedJob.status]?.label }}
                  </span>
                  <span class="text-xs text-indigo-200">{{ employmentTypeLabel[selectedJob.employment_type] || selectedJob.employment_type }}</span>
                </div>
                <h2 class="text-xl font-bold text-white">{{ selectedJob.title }}</h2>
                <div class="flex items-center gap-3 mt-2 flex-wrap">
                  <div class="flex items-center gap-1 text-indigo-200 text-sm">
                    <Building2 class="w-3.5 h-3.5" />
                    {{ selectedJob.department_name || 'No department' }}
                  </div>
                  <div v-if="selectedJob.location" class="flex items-center gap-1 text-indigo-200 text-sm">
                    <MapPin class="w-3.5 h-3.5" />
                    {{ selectedJob.location }}
                  </div>
                </div>
              </div>
              <button @click="showJobDetail = false" class="p-2 text-indigo-200 hover:text-white rounded-lg hover:bg-white/10 transition-colors ml-2 flex-shrink-0">
                <X class="w-5 h-5" />
              </button>
            </div>
          </div>

          <!-- Stats bar -->
          <div class="grid grid-cols-3 border-b border-slate-200 dark:border-slate-700 flex-shrink-0">
            <div class="flex flex-col items-center py-4 border-r border-slate-200 dark:border-slate-700">
              <p class="text-xl font-bold text-slate-900 dark:text-slate-100">{{ selectedJob.vacancies }}</p>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Vacancies</p>
            </div>
            <div class="flex flex-col items-center py-4 border-r border-slate-200 dark:border-slate-700">
              <p class="text-xl font-bold text-indigo-600 dark:text-indigo-400">{{ applicationsForJob(selectedJob.id).length }}</p>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Applications</p>
            </div>
            <div class="flex flex-col items-center py-4">
              <p class="text-xl font-bold text-emerald-600 dark:text-emerald-400">{{ applicationsForJob(selectedJob.id).filter(a => a.status === 'hired').length }}</p>
              <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">Hired</p>
            </div>
          </div>

          <!-- Content -->
          <div class="flex-1 overflow-y-auto px-6 py-5 space-y-6">

            <!-- Dates -->
            <div class="grid grid-cols-2 gap-4">
              <div class="bg-slate-50 dark:bg-slate-800 rounded-xl p-3">
                <p class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1">Published</p>
                <p class="text-sm font-semibold text-slate-800 dark:text-slate-100">{{ fmtDate(selectedJob.published_at) }}</p>
              </div>
              <div class="bg-slate-50 dark:bg-slate-800 rounded-xl p-3">
                <p class="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide mb-1">Deadline</p>
                <p :class="['text-sm font-semibold', isDeadlinePast(selectedJob.deadline_date) ? 'text-rose-600 dark:text-rose-400' : 'text-slate-800 dark:text-slate-100']">
                  {{ fmtDate(selectedJob.deadline_date) }}
                </p>
              </div>
            </div>

            <!-- Description -->
            <div v-if="selectedJob.description">
              <h3 class="text-sm font-bold text-slate-700 dark:text-slate-300 mb-2 flex items-center gap-2">
                <FileText class="w-4 h-4 text-slate-400" />
                Job Description
              </h3>
              <p class="text-sm text-slate-600 dark:text-slate-400 leading-relaxed whitespace-pre-line">{{ selectedJob.description }}</p>
            </div>

            <!-- Requirements -->
            <div v-if="selectedJob.requirements">
              <h3 class="text-sm font-bold text-slate-700 dark:text-slate-300 mb-2 flex items-center gap-2">
                <CheckCircle class="w-4 h-4 text-slate-400" />
                Requirements
              </h3>
              <p class="text-sm text-slate-600 dark:text-slate-400 leading-relaxed whitespace-pre-line">{{ selectedJob.requirements }}</p>
            </div>

            <!-- Application pipeline mini -->
            <div>
              <h3 class="text-sm font-bold text-slate-700 dark:text-slate-300 mb-3 flex items-center gap-2">
                <Users class="w-4 h-4 text-slate-400" />
                Pipeline Summary
              </h3>
              <div class="space-y-2">
                <div v-for="stage in ['new','screening','interview','offer','hired','rejected']" :key="stage"
                  class="flex items-center gap-3">
                  <span :class="['w-2 h-2 rounded-full flex-shrink-0', appStatusConfig[stage]?.dot]"></span>
                  <span class="text-sm text-slate-600 dark:text-slate-400 flex-1">{{ appStatusConfig[stage]?.label }}</span>
                  <span class="text-sm font-semibold text-slate-800 dark:text-slate-100">
                    {{ applicationsForJob(selectedJob.id).filter(a => a.status === stage).length }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer actions -->
          <div class="flex items-center gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-700 flex-shrink-0">
            <button @click="openCreateApp(selectedJob.id); showJobDetail = false"
              class="flex-1 py-2 text-sm font-semibold text-indigo-600 dark:text-indigo-400 bg-indigo-50 dark:bg-indigo-900/30 rounded-lg hover:bg-indigo-100 dark:hover:bg-indigo-900/50 transition-colors flex items-center justify-center gap-2">
              <User class="w-4 h-4" />
              Add Candidate
            </button>
            <button @click="openEditJob(selectedJob); showJobDetail = false"
              class="flex-1 py-2 text-sm font-semibold text-white bg-indigo-600 hover:bg-indigo-700 rounded-lg transition-colors flex items-center justify-center gap-2">
              <Edit2 class="w-4 h-4" />
              Edit Posting
            </button>
          </div>
        </div>
      </div>
    </Teleport>


    <!-- ══════════════════════════════════════════════════════════════════ -->
    <!-- MODAL: Add Application                                             -->
    <!-- ══════════════════════════════════════════════════════════════════ -->
    <Teleport to="body">
      <div v-if="showAppModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showAppModal = false"></div>
        <div class="relative bg-white dark:bg-slate-900 rounded-2xl shadow-2xl w-full max-w-xl max-h-[90vh] overflow-hidden flex flex-col">

          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700 flex-shrink-0">
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 bg-amber-100 dark:bg-amber-900/40 rounded-lg flex items-center justify-center">
                <User class="w-5 h-5 text-amber-600 dark:text-amber-400" />
              </div>
              <div>
                <h2 class="text-base font-bold text-slate-900 dark:text-slate-100">Add Application</h2>
                <p class="text-xs text-slate-500 dark:text-slate-400">Register a new candidate</p>
              </div>
            </div>
            <button @click="showAppModal = false" class="p-2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Body -->
          <div class="overflow-y-auto flex-1 px-6 py-5 space-y-4">

            <!-- Job Posting -->
            <div>
              <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Job Posting <span class="text-rose-500">*</span></label>
              <div class="relative">
                <select v-model="appForm.job_posting_id"
                  class="w-full appearance-none pl-3 pr-8 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500">
                  <option value="">Select job posting...</option>
                  <option v-for="j in jobs" :key="j.id" :value="j.id">{{ j.title }}</option>
                </select>
                <ChevronDown class="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
              </div>
            </div>

            <!-- Name row -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">First Name <span class="text-rose-500">*</span></label>
                <input v-model="appForm.first_name" type="text" placeholder="First name"
                  class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Last Name <span class="text-rose-500">*</span></label>
                <input v-model="appForm.last_name" type="text" placeholder="Last name"
                  class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
            </div>

            <!-- Contact row -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Email <span class="text-rose-500">*</span></label>
                <div class="relative">
                  <Mail class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
                  <input v-model="appForm.email" type="email" placeholder="email@example.com"
                    class="w-full pl-9 pr-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
              </div>
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Phone</label>
                <div class="relative">
                  <Phone class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
                  <input v-model="appForm.phone" type="tel" placeholder="+213 ..."
                    class="w-full pl-9 pr-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
              </div>
            </div>

            <!-- Source + Salary row -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Source</label>
                <div class="relative">
                  <select v-model="appForm.source"
                    class="w-full appearance-none pl-3 pr-8 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500">
                    <option v-for="(lbl, val) in sourceLabel" :key="val" :value="val">{{ lbl }}</option>
                  </select>
                  <ChevronDown class="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
                </div>
              </div>
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Expected Salary (DZD)</label>
                <input v-model="appForm.expected_salary" type="number" min="0" step="1000" placeholder="0"
                  class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
            </div>

            <!-- CV URL -->
            <div>
              <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">CV / Resume URL</label>
              <input v-model="appForm.cv_url" type="url" placeholder="https://..."
                class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
            </div>

            <!-- Cover letter -->
            <div>
              <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Cover Letter</label>
              <textarea v-model="appForm.cover_letter" rows="3" placeholder="Optional cover letter text..."
                class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"></textarea>
            </div>
          </div>

          <!-- Footer -->
          <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/50 flex-shrink-0">
            <button @click="showAppModal = false"
              class="px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
              Cancel
            </button>
            <button @click="saveApplication()" :disabled="appSaving"
              class="px-5 py-2 text-sm font-semibold text-white bg-amber-500 hover:bg-amber-600 disabled:opacity-60 disabled:cursor-not-allowed rounded-lg transition-colors flex items-center gap-2">
              <Loader2 v-if="appSaving" class="w-4 h-4 animate-spin" />
              Submit Application
            </button>
          </div>
        </div>
      </div>
    </Teleport>


    <!-- ══════════════════════════════════════════════════════════════════ -->
    <!-- MODAL: Update Application Status                                   -->
    <!-- ══════════════════════════════════════════════════════════════════ -->
    <Teleport to="body">
      <div v-if="showStatusModal && selectedApp" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showStatusModal = false"></div>
        <div class="relative bg-white dark:bg-slate-900 rounded-2xl shadow-2xl w-full max-w-md overflow-hidden flex flex-col">

          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700 flex-shrink-0">
            <div>
              <h2 class="text-base font-bold text-slate-900 dark:text-slate-100">Update Stage</h2>
              <p class="text-xs text-slate-500 dark:text-slate-400">{{ selectedApp.first_name }} {{ selectedApp.last_name }}</p>
            </div>
            <button @click="showStatusModal = false" class="p-2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Body -->
          <div class="px-6 py-5 space-y-4">

            <!-- Candidate info bar -->
            <div class="bg-slate-50 dark:bg-slate-800 rounded-xl p-3 flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-indigo-100 dark:bg-indigo-900/40 flex items-center justify-center text-sm font-bold text-indigo-600 dark:text-indigo-400 flex-shrink-0">
                {{ selectedApp.first_name.charAt(0) }}{{ selectedApp.last_name.charAt(0) }}
              </div>
              <div class="min-w-0">
                <p class="text-sm font-semibold text-slate-800 dark:text-slate-100">{{ selectedApp.first_name }} {{ selectedApp.last_name }}</p>
                <p class="text-xs text-slate-500 dark:text-slate-400 truncate">{{ selectedApp.email }} · {{ selectedApp.job_title || 'Unknown Job' }}</p>
              </div>
              <span :class="['ml-auto inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold flex-shrink-0', appStatusConfig[selectedApp.status]?.color]">
                <span :class="['w-1.5 h-1.5 rounded-full', appStatusConfig[selectedApp.status]?.dot]"></span>
                {{ appStatusConfig[selectedApp.status]?.label }}
              </span>
            </div>

            <!-- Stage pipeline visual -->
            <div>
              <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-2">Move to Stage</label>
              <div class="grid grid-cols-2 gap-2">
                <button
                  v-for="stage in ['new','screening','interview','offer','hired','rejected','withdrawn']"
                  :key="stage"
                  @click="statusForm.status = stage as Application['status']"
                  :class="[
                    'flex items-center gap-2 px-3 py-2.5 rounded-xl border-2 text-sm font-medium transition-all',
                    statusForm.status === stage
                      ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-300'
                      : 'border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-600'
                  ]"
                >
                  <span :class="['w-2 h-2 rounded-full flex-shrink-0', appStatusConfig[stage]?.dot]"></span>
                  {{ appStatusConfig[stage]?.label }}
                </button>
              </div>
            </div>

            <!-- Interview fields -->
            <template v-if="statusForm.status === 'interview'">
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Interview Date & Time</label>
                <input v-model="statusForm.interview_date" type="datetime-local"
                  class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500" />
              </div>
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Interview Notes</label>
                <textarea v-model="statusForm.interview_notes" rows="2" placeholder="Notes about the interview..."
                  class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"></textarea>
              </div>
            </template>

            <!-- Rejection reason -->
            <template v-if="statusForm.status === 'rejected'">
              <div>
                <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wide mb-1.5">Rejection Reason</label>
                <textarea v-model="statusForm.rejection_reason" rows="2" placeholder="Optional reason for rejection..."
                  class="w-full px-3 py-2 text-sm bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 resize-none"></textarea>
              </div>
            </template>

            <!-- Hired success message -->
            <div v-if="statusForm.status === 'hired'" class="flex items-center gap-3 bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-xl p-3">
              <UserCheck class="w-5 h-5 text-emerald-600 dark:text-emerald-400 flex-shrink-0" />
              <p class="text-sm text-emerald-700 dark:text-emerald-300">
                This candidate will be marked as <strong>Hired</strong>. Create their employee record in the Employees module.
              </p>
            </div>
          </div>

          <!-- Footer -->
          <div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800/50 flex-shrink-0">
            <button @click="showStatusModal = false"
              class="px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
              Cancel
            </button>
            <button @click="saveStatus()" :disabled="statusSaving"
              class="px-5 py-2 text-sm font-semibold text-white bg-indigo-600 hover:bg-indigo-700 disabled:opacity-60 disabled:cursor-not-allowed rounded-lg transition-colors flex items-center gap-2">
              <Loader2 v-if="statusSaving" class="w-4 h-4 animate-spin" />
              Update Stage
            </button>
          </div>
        </div>
      </div>
    </Teleport>


    <!-- ══════════════════════════════════════════════════════════════════ -->
    <!-- MODAL: Delete Job Confirmation                                     -->
    <!-- ══════════════════════════════════════════════════════════════════ -->
    <Teleport to="body">
      <div v-if="showDeleteJobConfirm" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="showDeleteJobConfirm = false"></div>
        <div class="relative bg-white dark:bg-slate-900 rounded-2xl shadow-2xl w-full max-w-sm overflow-hidden">
          <div class="p-6 text-center">
            <div class="w-14 h-14 bg-rose-100 dark:bg-rose-900/30 rounded-full flex items-center justify-center mx-auto mb-4">
              <AlertCircle class="w-7 h-7 text-rose-600 dark:text-rose-400" />
            </div>
            <h3 class="text-lg font-bold text-slate-900 dark:text-slate-100 mb-2">Delete Job Posting</h3>
            <p class="text-sm text-slate-500 dark:text-slate-400">
              This will permanently delete the job posting and all associated applications. This action cannot be undone.
            </p>
          </div>
          <div class="flex gap-3 px-6 pb-6">
            <button @click="showDeleteJobConfirm = false; deletingJobId = null"
              class="flex-1 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
              Cancel
            </button>
            <button @click="deleteJob()" :disabled="deletingJob"
              class="flex-1 py-2 text-sm font-semibold text-white bg-rose-600 hover:bg-rose-700 disabled:opacity-60 disabled:cursor-not-allowed rounded-lg transition-colors flex items-center justify-center gap-2">
              <Loader2 v-if="deletingJob" class="w-4 h-4 animate-spin" />
              Delete
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
