<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { salesAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  Plus, X, Save, Search, Filter, RefreshCw, Edit2, Trash2,
  Building2, User, Phone, Mail, Globe, MapPin, FileText,
  CreditCard, DollarSign, ShieldCheck, ChevronDown, ChevronUp,
  CheckCircle, XCircle, MoreVertical, Eye
} from 'lucide-vue-next'

// ─── Types ─────────────────────────────────────────────────────────────────

interface Customer {
  id: string
  company_id: string
  code: string
  name: string
  type: string          // company | individual
  nif: string
  nis: string
  rc: string
  art: string
  tax_regime: string    // reel | forfait | exonere
  address: string
  city: string
  wilaya: string
  postal_code: string
  phone: string
  email: string
  website: string
  credit_limit: number
  balance: number
  payment_terms: number
  account_id?: string
  salesperson_id?: string
  is_active: boolean
  notes: string
  created_at: string
  updated_at: string
}

const EMPTY_CUSTOMER = (): Partial<Customer> => ({
  code: '', name: '', type: 'company', nif: '', nis: '', rc: '', art: '',
  tax_regime: 'reel', address: '', city: '', wilaya: '', postal_code: '',
  phone: '', email: '', website: '', credit_limit: 0, balance: 0,
  payment_terms: 30, is_active: true, notes: '',
})

const WILAYAS = [
  'Adrar','Chlef','Laghouat','Oum El Bouaghi','Batna','Béjaïa','Biskra','Béchar',
  'Blida','Bouira','Tamanrasset','Tébessa','Tlemcen','Tiaret','Tizi Ouzou','Alger',
  'Djelfa','Jijel','Sétif','Saïda','Skikda','Sidi Bel Abbès','Annaba','Guelma',
  'Constantine','Médéa','Mostaganem','Msila','Mascara','Ouargla','Oran','El Bayadh',
  'Illizi','Bordj Bou Arréridj','Boumerdès','El Tarf','Tindouf','Tissemsilt',
  'El Oued','Khenchela','Souk Ahras','Tipaza','Mila','Aïn Defla','Naâma',
  'Aïn Témouchent','Ghardaïa','Relizane','Timimoun','Bordj Badji Mokhtar',
  'Ouled Djellal','Béni Abbès','In Salah','In Guezzam','Touggourt','Djanet',
  'El MGhair','El Menia',
]

// ─── State ─────────────────────────────────────────────────────────────────

const app = useAppStore()
const customers = ref<Customer[]>([])
const loading = ref(true)
const saving = ref(false)

// Filters
const search = ref('')
const filterType = ref('')
const filterActive = ref('')

// Modal
const showModal = ref(false)
const showDetail = ref(false)
const isEdit = ref(false)
const editingCustomer = ref<Partial<Customer>>(EMPTY_CUSTOMER())
const detailCustomer = ref<Customer | null>(null)
const activeTab = ref<'info' | 'fiscal' | 'financial' | 'notes'>('info')

// Sort
const sortBy = ref<keyof Customer>('name')
const sortDir = ref<'asc' | 'desc'>('asc')

// ─── Computed ──────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = [...customers.value]
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(c =>
      c.name.toLowerCase().includes(q) ||
      c.code.toLowerCase().includes(q) ||
      c.email?.toLowerCase().includes(q) ||
      c.phone?.toLowerCase().includes(q) ||
      c.nif?.toLowerCase().includes(q)
    )
  }
  if (filterType.value) list = list.filter(c => c.type === filterType.value)
  if (filterActive.value !== '') list = list.filter(c => String(c.is_active) === filterActive.value)
  // Sort
  list.sort((a, b) => {
    const av = String(a[sortBy.value] ?? '')
    const bv = String(b[sortBy.value] ?? '')
    return sortDir.value === 'asc' ? av.localeCompare(bv) : bv.localeCompare(av)
  })
  return list
})

const totalBalance = computed(() => customers.value.reduce((s, c) => s + (c.balance || 0), 0))
const activeCount = computed(() => customers.value.filter(c => c.is_active).length)

// ─── Data ──────────────────────────────────────────────────────────────────

async function loadData() {
  loading.value = true
  try {
    const res = await salesAPI.getCustomers()
    customers.value = res.data || []
  } catch {
    app.addToast('Failed to load customers', 'error')
  } finally {
    loading.value = false
  }
}

// ─── Modal ─────────────────────────────────────────────────────────────────

function openCreate() {
  isEdit.value = false
  editingCustomer.value = EMPTY_CUSTOMER()
  activeTab.value = 'info'
  showModal.value = true
}

function openEdit(c: Customer) {
  isEdit.value = true
  editingCustomer.value = { ...c }
  activeTab.value = 'info'
  showModal.value = true
}

function openDetail(c: Customer) {
  detailCustomer.value = c
  showDetail.value = true
}

function closeModal() { showModal.value = false }
function closeDetail() { showDetail.value = false }

async function save() {
  const c = editingCustomer.value
  if (!c.name?.trim()) { app.addToast('Customer name is required', 'error'); return }
  if (!c.code?.trim()) { app.addToast('Customer code is required', 'error'); return }
  saving.value = true
  try {
    if (isEdit.value && c.id) {
      await salesAPI.updateCustomer(c.id, c)
      app.addToast('Customer updated', 'success')
    } else {
      await salesAPI.createCustomer(c)
      app.addToast('Customer created', 'success')
    }
    closeModal()
    await loadData()
  } catch {
    app.addToast('Failed to save customer', 'error')
  } finally {
    saving.value = false
  }
}

async function deleteCustomer(c: Customer) {
  if (!confirm(`Delete customer "${c.name}"?`)) return
  try {
    await salesAPI.deleteCustomer(c.id)
    app.addToast('Customer deleted', 'success')
    await loadData()
  } catch {
    app.addToast('Failed to delete customer', 'error')
  }
}

// ─── Sort ──────────────────────────────────────────────────────────────────

function toggleSort(col: keyof Customer) {
  if (sortBy.value === col) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortBy.value = col; sortDir.value = 'asc' }
}

// ─── Formatters ────────────────────────────────────────────────────────────

function fmtCurrency(n?: number) {
  if (n == null) return '—'
  return n.toLocaleString('fr-DZ', { minimumFractionDigits: 2 }) + ' DZD'
}

function typeBadge(t: string) {
  return t === 'company'
    ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300'
    : 'bg-purple-100 text-purple-700 dark:bg-purple-900/40 dark:text-purple-300'
}

const MODAL_TABS = [
  { key: 'info',      label: 'General Info' },
  { key: 'fiscal',    label: 'Fiscal IDs' },
  { key: 'financial', label: 'Financial' },
  { key: 'notes',     label: 'Notes' },
]

onMounted(loadData)
</script>

<template>
  <div class="space-y-5">

    <!-- ─── Header ──────────────────────────────────────────────────────────── -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Customers</h1>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
          {{ activeCount }} active · {{ customers.length }} total · Receivables: {{ fmtCurrency(totalBalance) }}
        </p>
      </div>
      <div class="flex items-center gap-2">
        <button @click="loadData" :disabled="loading" class="p-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors disabled:opacity-50">
          <RefreshCw class="w-4 h-4" :class="{ 'animate-spin': loading }" />
        </button>
        <button @click="openCreate" class="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg shadow-sm transition-colors">
          <Plus class="w-4 h-4" />
          New Customer
        </button>
      </div>
    </div>

    <!-- ─── Filters ─────────────────────────────────────────────────────────── -->
    <div class="flex flex-wrap items-center gap-3">
      <div class="relative flex-1 min-w-48">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
        <input
          v-model="search"
          placeholder="Search name, code, email, NIF…"
          class="w-full pl-9 pr-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white placeholder-gray-400 focus:ring-2 focus:ring-blue-500 outline-none"
        />
      </div>
      <select v-model="filterType" class="px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 outline-none">
        <option value="">All Types</option>
        <option value="company">Company</option>
        <option value="individual">Individual</option>
      </select>
      <select v-model="filterActive" class="px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 outline-none">
        <option value="">All Status</option>
        <option value="true">Active</option>
        <option value="false">Inactive</option>
      </select>
    </div>

    <!-- ─── Table ─────────────────────────────────────────────────────────────── -->
    <div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center h-40">
        <RefreshCw class="w-6 h-6 text-gray-400 animate-spin" />
      </div>
      <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center h-40 gap-2 text-gray-400 dark:text-gray-600">
        <Building2 class="w-8 h-8" />
        <p class="text-sm">No customers found</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <tr>
              <th @click="toggleSort('code')" class="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 whitespace-nowrap">
                <div class="flex items-center gap-1">
                  Code
                  <ChevronUp v-if="sortBy==='code' && sortDir==='asc'" class="w-3 h-3" />
                  <ChevronDown v-else-if="sortBy==='code'" class="w-3 h-3" />
                </div>
              </th>
              <th @click="toggleSort('name')" class="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide cursor-pointer hover:text-gray-700 dark:hover:text-gray-200">
                <div class="flex items-center gap-1">
                  Name
                  <ChevronUp v-if="sortBy==='name' && sortDir==='asc'" class="w-3 h-3" />
                  <ChevronDown v-else-if="sortBy==='name'" class="w-3 h-3" />
                </div>
              </th>
              <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide whitespace-nowrap">Type</th>
              <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide whitespace-nowrap">NIF</th>
              <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">City / Wilaya</th>
              <th class="px-4 py-3 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Contact</th>
              <th @click="toggleSort('balance')" class="px-4 py-3 text-right text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide cursor-pointer hover:text-gray-700 dark:hover:text-gray-200 whitespace-nowrap">
                <div class="flex items-center justify-end gap-1">
                  Balance
                  <ChevronUp v-if="sortBy==='balance' && sortDir==='asc'" class="w-3 h-3" />
                  <ChevronDown v-else-if="sortBy==='balance'" class="w-3 h-3" />
                </div>
              </th>
              <th class="px-4 py-3 text-center text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Status</th>
              <th class="px-4 py-3" />
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-100 dark:divide-gray-800">
            <tr
              v-for="c in filtered"
              :key="c.id"
              class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
            >
              <td class="px-4 py-3 text-xs font-mono font-medium text-gray-500 dark:text-gray-400">{{ c.code }}</td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-2">
                  <div class="w-7 h-7 rounded-full bg-blue-100 dark:bg-blue-900/40 flex items-center justify-center flex-shrink-0">
                    <Building2 v-if="c.type==='company'" class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" />
                    <User v-else class="w-3.5 h-3.5 text-purple-600 dark:text-purple-400" />
                  </div>
                  <div>
                    <p class="font-semibold text-gray-900 dark:text-white">{{ c.name }}</p>
                    <p v-if="c.website" class="text-xs text-gray-400 truncate max-w-40">{{ c.website }}</p>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3">
                <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium capitalize" :class="typeBadge(c.type)">
                  {{ c.type }}
                </span>
              </td>
              <td class="px-4 py-3 text-xs font-mono text-gray-700 dark:text-gray-300">{{ c.nif || '—' }}</td>
              <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                {{ [c.city, c.wilaya].filter(Boolean).join(', ') || '—' }}
              </td>
              <td class="px-4 py-3">
                <div class="space-y-0.5">
                  <div v-if="c.phone" class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-400">
                    <Phone class="w-3 h-3 flex-shrink-0" />
                    <span>{{ c.phone }}</span>
                  </div>
                  <div v-if="c.email" class="flex items-center gap-1 text-xs text-gray-600 dark:text-gray-400">
                    <Mail class="w-3 h-3 flex-shrink-0" />
                    <span class="truncate max-w-40">{{ c.email }}</span>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3 text-right">
                <span :class="c.balance > 0 ? 'text-blue-700 dark:text-blue-400 font-semibold' : 'text-gray-500 dark:text-gray-400'">
                  {{ fmtCurrency(c.balance) }}
                </span>
              </td>
              <td class="px-4 py-3 text-center">
                <CheckCircle v-if="c.is_active" class="w-4 h-4 text-green-500 inline" />
                <XCircle v-else class="w-4 h-4 text-red-400 inline" />
              </td>
              <td class="px-4 py-3">
                <div class="flex items-center gap-1 justify-end">
                  <button @click="openDetail(c)" class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 transition-colors" title="View">
                    <Eye class="w-3.5 h-3.5" />
                  </button>
                  <button @click="openEdit(c)" class="p-1.5 rounded-lg hover:bg-blue-50 dark:hover:bg-blue-900/40 text-blue-600 dark:text-blue-400 transition-colors" title="Edit">
                    <Edit2 class="w-3.5 h-3.5" />
                  </button>
                  <button @click="deleteCustomer(c)" class="p-1.5 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/40 text-red-500 dark:text-red-400 transition-colors" title="Delete">
                    <Trash2 class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ─── Create / Edit Modal ─────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm" @click.self="closeModal">
        <div class="w-full max-w-2xl bg-white dark:bg-gray-900 rounded-2xl shadow-2xl border border-gray-200 dark:border-gray-700 flex flex-col max-h-[92vh]">

          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div class="flex items-center gap-2">
              <Building2 class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ isEdit ? 'Edit Customer' : 'New Customer' }}
              </h2>
            </div>
            <button @click="closeModal" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition-colors"><X class="w-5 h-5" /></button>
          </div>

          <!-- Tabs -->
          <div class="flex border-b border-gray-200 dark:border-gray-700 px-6">
            <button
              v-for="tab in MODAL_TABS"
              :key="tab.key"
              @click="activeTab = tab.key as any"
              class="px-4 py-3 text-sm font-medium border-b-2 transition-colors"
              :class="activeTab === tab.key
                ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'"
            >
              {{ tab.label }}
            </button>
          </div>

          <!-- Body -->
          <div class="flex-1 overflow-y-auto p-6">
            <div v-if="editingCustomer">

              <!-- ── General Info ── -->
              <div v-show="activeTab === 'info'" class="grid grid-cols-2 gap-4">
                <div>
                  <label class="label">Code <span class="text-red-500">*</span></label>
                  <input v-model="editingCustomer.code" class="input" placeholder="CLI001" />
                </div>
                <div>
                  <label class="label">Type</label>
                  <select v-model="editingCustomer.type" class="input">
                    <option value="company">Company</option>
                    <option value="individual">Individual</option>
                  </select>
                </div>
                <div class="col-span-2">
                  <label class="label">Name <span class="text-red-500">*</span></label>
                  <input v-model="editingCustomer.name" class="input" placeholder="Company name or full name" />
                </div>
                <div>
                  <label class="label">Phone</label>
                  <input v-model="editingCustomer.phone" class="input" placeholder="+213 xx xx xx xx" />
                </div>
                <div>
                  <label class="label">Email</label>
                  <input v-model="editingCustomer.email" type="email" class="input" placeholder="contact@company.dz" />
                </div>
                <div class="col-span-2">
                  <label class="label">Website</label>
                  <input v-model="editingCustomer.website" class="input" placeholder="https://www.company.dz" />
                </div>
                <div class="col-span-2">
                  <label class="label">Address</label>
                  <input v-model="editingCustomer.address" class="input" placeholder="Street address" />
                </div>
                <div>
                  <label class="label">City</label>
                  <input v-model="editingCustomer.city" class="input" placeholder="City" />
                </div>
                <div>
                  <label class="label">Wilaya</label>
                  <select v-model="editingCustomer.wilaya" class="input">
                    <option value="">— Select —</option>
                    <option v-for="w in WILAYAS" :key="w" :value="w">{{ w }}</option>
                  </select>
                </div>
                <div>
                  <label class="label">Postal Code</label>
                  <input v-model="editingCustomer.postal_code" class="input" placeholder="16000" />
                </div>
                <div>
                  <label class="label">Status</label>
                  <select v-model="editingCustomer.is_active" class="input">
                    <option :value="true">Active</option>
                    <option :value="false">Inactive</option>
                  </select>
                </div>
              </div>

              <!-- ── Fiscal IDs ── -->
              <div v-show="activeTab === 'fiscal'" class="grid grid-cols-2 gap-4">
                <div>
                  <label class="label">NIF (Numéro d'Identification Fiscale)</label>
                  <input v-model="editingCustomer.nif" class="input font-mono" placeholder="000123456789012" />
                </div>
                <div>
                  <label class="label">NIS (Numéro d'Identification Statistique)</label>
                  <input v-model="editingCustomer.nis" class="input font-mono" placeholder="123456789012345" />
                </div>
                <div>
                  <label class="label">RC (Registre de Commerce)</label>
                  <input v-model="editingCustomer.rc" class="input font-mono" placeholder="16/00-1234567B12" />
                </div>
                <div>
                  <label class="label">Article d'Imposition (ART)</label>
                  <input v-model="editingCustomer.art" class="input font-mono" placeholder="12345678901" />
                </div>
                <div class="col-span-2">
                  <label class="label">Tax Regime</label>
                  <select v-model="editingCustomer.tax_regime" class="input">
                    <option value="reel">Régime Réel</option>
                    <option value="forfait">Régime Forfaitaire</option>
                    <option value="exonere">Exonéré</option>
                  </select>
                </div>
              </div>

              <!-- ── Financial ── -->
              <div v-show="activeTab === 'financial'" class="grid grid-cols-2 gap-4">
                <div>
                  <label class="label">Credit Limit (DZD)</label>
                  <input v-model.number="editingCustomer.credit_limit" type="number" min="0" step="10000" class="input" />
                </div>
                <div>
                  <label class="label">Payment Terms (days)</label>
                  <input v-model.number="editingCustomer.payment_terms" type="number" min="0" max="365" class="input" />
                </div>
                <div>
                  <label class="label">Current Balance (DZD)</label>
                  <input :value="editingCustomer.balance" readonly class="input opacity-60 cursor-not-allowed bg-gray-50 dark:bg-gray-800" />
                </div>
              </div>

              <!-- ── Notes ── -->
              <div v-show="activeTab === 'notes'">
                <label class="label">Internal Notes</label>
                <textarea v-model="editingCustomer.notes" rows="8" class="input resize-none" placeholder="Internal notes about this customer…" />
              </div>

            </div>
          </div>

          <!-- Footer -->
          <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-3">
            <button @click="closeModal" class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors">
              Cancel
            </button>
            <button @click="save" :disabled="saving" class="inline-flex items-center gap-2 px-5 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg shadow-sm transition-colors disabled:opacity-60">
              <Save class="w-4 h-4" />
              {{ saving ? 'Saving…' : (isEdit ? 'Update' : 'Create') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- ─── Detail Drawer ─────────────────────────────────────────────────────── -->
    <Teleport to="body">
      <div v-if="showDetail && detailCustomer" class="fixed inset-0 z-50 flex" @click.self="closeDetail">
        <div class="ml-auto w-full max-w-md bg-white dark:bg-gray-900 border-l border-gray-200 dark:border-gray-700 shadow-2xl flex flex-col max-h-screen overflow-y-auto">
          <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 sticky top-0 bg-white dark:bg-gray-900 z-10">
            <div class="flex items-center gap-2">
              <Building2 class="w-5 h-5 text-blue-600 dark:text-blue-400" />
              <h2 class="text-base font-semibold text-gray-900 dark:text-white truncate">{{ detailCustomer.name }}</h2>
            </div>
            <button @click="closeDetail" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"><X class="w-5 h-5" /></button>
          </div>
          <div class="p-6 space-y-6">
            <!-- Code / Type / Status -->
            <div class="flex items-center gap-3 flex-wrap">
              <span class="text-xs font-mono bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 px-2 py-1 rounded">{{ detailCustomer.code }}</span>
              <span class="text-xs px-2 py-1 rounded-full font-medium capitalize" :class="typeBadge(detailCustomer.type)">{{ detailCustomer.type }}</span>
              <span :class="detailCustomer.is_active ? 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-400' : 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-400'" class="text-xs px-2 py-1 rounded-full font-medium">
                {{ detailCustomer.is_active ? 'Active' : 'Inactive' }}
              </span>
            </div>
            <!-- Contact section -->
            <div class="space-y-2">
              <h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide">Contact</h3>
              <div v-if="detailCustomer.phone" class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300"><Phone class="w-4 h-4 text-gray-400 flex-shrink-0" />{{ detailCustomer.phone }}</div>
              <div v-if="detailCustomer.email" class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300"><Mail class="w-4 h-4 text-gray-400 flex-shrink-0" />{{ detailCustomer.email }}</div>
              <div v-if="detailCustomer.website" class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300"><Globe class="w-4 h-4 text-gray-400 flex-shrink-0" />{{ detailCustomer.website }}</div>
              <div v-if="detailCustomer.address || detailCustomer.city" class="flex items-start gap-2 text-sm text-gray-700 dark:text-gray-300"><MapPin class="w-4 h-4 text-gray-400 flex-shrink-0 mt-0.5" /><span>{{ [detailCustomer.address, detailCustomer.city, detailCustomer.wilaya, detailCustomer.postal_code].filter(Boolean).join(', ') }}</span></div>
            </div>
            <!-- Fiscal -->
            <div class="space-y-2">
              <h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide">Fiscal Identifiers</h3>
              <div class="grid grid-cols-2 gap-2 text-xs">
                <div v-for="[k, v] in [['NIF', detailCustomer.nif], ['NIS', detailCustomer.nis], ['RC', detailCustomer.rc], ['ART', detailCustomer.art]]" :key="k" class="bg-gray-50 dark:bg-gray-800 rounded-lg p-2">
                  <p class="text-gray-400 dark:text-gray-500 mb-0.5">{{ k }}</p>
                  <p class="font-mono font-medium text-gray-800 dark:text-gray-200">{{ v || '—' }}</p>
                </div>
              </div>
              <p class="text-xs text-gray-500 dark:text-gray-400 capitalize">Régime: {{ detailCustomer.tax_regime }}</p>
            </div>
            <!-- Financial -->
            <div class="space-y-2">
              <h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide">Financial</h3>
              <div class="grid grid-cols-3 gap-2 text-center">
                <div class="bg-blue-50 dark:bg-blue-900/20 rounded-lg p-2">
                  <p class="text-xs text-blue-500 dark:text-blue-400">Balance</p>
                  <p class="text-sm font-bold text-blue-700 dark:text-blue-300">{{ fmtCurrency(detailCustomer.balance) }}</p>
                </div>
                <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-2">
                  <p class="text-xs text-gray-400 dark:text-gray-500">Credit Limit</p>
                  <p class="text-sm font-bold text-gray-700 dark:text-gray-300">{{ fmtCurrency(detailCustomer.credit_limit) }}</p>
                </div>
                <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-2">
                  <p class="text-xs text-gray-400 dark:text-gray-500">Payment Terms</p>
                  <p class="text-sm font-bold text-gray-700 dark:text-gray-300">{{ detailCustomer.payment_terms }} d</p>
                </div>
              </div>
            </div>
            <!-- Notes -->
            <div v-if="detailCustomer.notes" class="space-y-2">
              <h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wide">Notes</h3>
              <p class="text-sm text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{{ detailCustomer.notes }}</p>
            </div>
            <!-- Actions -->
            <div class="flex gap-2 pt-2 border-t border-gray-200 dark:border-gray-700">
              <button @click="openEdit(detailCustomer); closeDetail()" class="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium rounded-lg transition-colors">
                <Edit2 class="w-4 h-4" />Edit
              </button>
              <button @click="deleteCustomer(detailCustomer); closeDetail()" class="px-4 py-2 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 text-sm font-medium rounded-lg transition-colors">
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<style scoped>
.label {
  @apply block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1.5;
}
.input {
  @apply w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800
         text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500 outline-none transition-shadow;
}
</style>
