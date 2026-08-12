<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { purchaseAPI } from '@/api/client'
import { useAppStore } from '@/stores/app'
import {
  Plus, X, Save, Search, RefreshCw, Edit2, Trash2,
  Building2, Phone, Mail, MapPin, User, Star,
  CreditCard, Hash, ChevronDown, ChevronUp, CheckCircle,
  AlertTriangle, ShieldCheck
} from 'lucide-vue-next'

// ─── Types ────────────────────────────────────────────────────────────────────

interface Supplier {
  id: string
  code: string
  name: string
  type: string
  nif: string
  nis: string
  rc: string
  art: string
  address: string
  city: string
  wilaya: string
  phone: string
  email: string
  contact_name: string
  payment_terms: number
  credit_limit: number
  balance: number
  rating: number
  is_active: boolean
  notes: string
  created_at: string
}

const EMPTY_SUPPLIER = (): Partial<Supplier> => ({
  code: '', name: '', type: 'company',
  nif: '', nis: '', rc: '', art: '',
  address: '', city: '', wilaya: '',
  phone: '', email: '', contact_name: '',
  payment_terms: 30, credit_limit: 0, rating: 3,
  is_active: true, notes: ''
})

const WILAYAS = [
  'Adrar','Chlef','Laghouat','Oum El Bouaghi','Batna','Béjaïa','Biskra','Béchar',
  'Blida','Bouira','Tamanrasset','Tébessa','Tlemcen','Tiaret','Tizi Ouzou','Alger',
  'Djelfa','Jijel','Sétif','Saïda','Skikda','Sidi Bel Abbès','Annaba','Guelma',
  'Constantine','Médéa','Mostaganem','MSila','Mascara','Ouargla','Oran','El Bayadh',
  'Illizi','Bordj Bou Arréridj','Boumerdès','El Tarf','Tindouf','Tissemsilt',
  'El Oued','Khenchela','Souk Ahras','Tipaza','Mila','Aïn Defla','Naâma',
  'Aïn Témouchent','Ghardaïa','Relizane','Timimoun','Bordj Badji Mokhtar',
  'Ouled Djellal','Béni Abbès','In Salah','In Guezzam','Touggourt','Djanet',
  'El MGhair','El Meniaa'
]

// ─── State ────────────────────────────────────────────────────────────────────

const app = useAppStore()
const suppliers = ref<Supplier[]>([])
const loading = ref(true)
const saving = ref(false)
const showForm = ref(false)
const showDetail = ref<Supplier | null>(null)
const isEdit = ref(false)
const searchQ = ref('')
const filterActive = ref<'all' | 'active' | 'inactive'>('all')
const sortField = ref<keyof Supplier>('name')
const sortDir = ref<'asc' | 'desc'>('asc')
const form = ref<Partial<Supplier>>(EMPTY_SUPPLIER())

// ─── Computed ─────────────────────────────────────────────────────────────────

const filtered = computed(() => {
  let list = suppliers.value
  if (searchQ.value.trim()) {
    const q = searchQ.value.toLowerCase()
    list = list.filter(s =>
      s.name.toLowerCase().includes(q) ||
      s.code.toLowerCase().includes(q) ||
      (s.nif || '').toLowerCase().includes(q) ||
      (s.email || '').toLowerCase().includes(q) ||
      (s.city || '').toLowerCase().includes(q)
    )
  }
  if (filterActive.value === 'active') list = list.filter(s => s.is_active)
  if (filterActive.value === 'inactive') list = list.filter(s => !s.is_active)
  return [...list].sort((a, b) => {
    const av = (a as any)[sortField.value] ?? ''
    const bv = (b as any)[sortField.value] ?? ''
    const r = String(av).localeCompare(String(bv))
    return sortDir.value === 'asc' ? r : -r
  })
})

const stats = computed(() => ({
  total: suppliers.value.length,
  active: suppliers.value.filter(s => s.is_active).length,
  totalBalance: suppliers.value.reduce((sum, s) => sum + (s.balance || 0), 0),
  avgRating: suppliers.value.length
    ? (suppliers.value.reduce((sum, s) => sum + (s.rating || 0), 0) / suppliers.value.length).toFixed(1)
    : '0.0'
}))

// ─── Methods ─────────────────────────────────────────────────────────────────

async function load() {
  loading.value = true
  try {
    const res = await purchaseAPI.getSuppliers()
    suppliers.value = res.data || []
  } catch {
    app.addToast('Failed to load suppliers', 'error')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.value = EMPTY_SUPPLIER()
  isEdit.value = false
  showForm.value = true
}

function openEdit(s: Supplier) {
  form.value = { ...s }
  isEdit.value = true
  showForm.value = true
}

async function save() {
  if (!form.value.name?.trim()) {
    app.addToast('Supplier name is required', 'error')
    return
  }
  saving.value = true
  try {
    if (isEdit.value && form.value.id) {
      await purchaseAPI.updateSupplier(form.value.id, form.value)
      app.addToast('Supplier updated', 'success')
    } else {
      await purchaseAPI.createSupplier(form.value)
      app.addToast('Supplier created', 'success')
    }
    showForm.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Save failed', 'error')
  } finally {
    saving.value = false
  }
}

async function deactivate(s: Supplier) {
  if (!confirm(`Deactivate supplier "${s.name}"?`)) return
  try {
    await purchaseAPI.deleteSupplier(s.id)
    app.addToast('Supplier deactivated', 'success')
    await load()
  } catch {
    app.addToast('Failed to deactivate', 'error')
  }
}

function toggleSort(field: keyof Supplier) {
  if (sortField.value === field) sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  else { sortField.value = field; sortDir.value = 'asc' }
}

function ratingColor(r: number) {
  if (r >= 4) return 'text-green-600 dark:text-green-400'
  if (r >= 3) return 'text-amber-600 dark:text-amber-400'
  return 'text-red-600 dark:text-red-400'
}

function fmt(n: number) {
  return new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2 }).format(n)
}

onMounted(load)
</script>

<template>
  <div class="space-y-5">

    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <h1 class="text-2xl font-bold text-slate-900 dark:text-white">Suppliers</h1>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Manage your supplier directory and fiscal information
        </p>
      </div>
      <button @click="openCreate"
        class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary-600 text-white
               text-sm font-semibold hover:bg-primary-700 transition-colors shadow-sm">
        <Plus :size="16" /> New Supplier
      </button>
    </div>

    <!-- KPI Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total</p>
        <p class="text-2xl font-bold text-slate-900 dark:text-white mt-1">{{ stats.total }}</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Active</p>
        <p class="text-2xl font-bold text-green-600 dark:text-green-400 mt-1">{{ stats.active }}</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Total Balance</p>
        <p class="text-2xl font-bold text-red-600 dark:text-red-400 mt-1">{{ fmt(stats.totalBalance) }}</p>
      </div>
      <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <p class="text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wide">Avg Rating</p>
        <p class="text-2xl font-bold text-amber-600 dark:text-amber-400 mt-1">{{ stats.avgRating }} / 5</p>
      </div>
    </div>

    <!-- Filters -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
      <div class="flex flex-wrap gap-3 items-center">
        <div class="relative flex-1 min-w-[200px]">
          <Search :size="16" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input v-model="searchQ" placeholder="Search name, code, NIF, city..."
            class="w-full pl-9 pr-4 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                   bg-slate-50 dark:bg-slate-700/50 text-slate-900 dark:text-white
                   focus:outline-none focus:ring-2 focus:ring-primary-500" />
        </div>
        <div class="flex rounded-lg overflow-hidden border border-slate-200 dark:border-slate-600">
          <button v-for="opt in (['all','active','inactive'] as const)" :key="opt"
            @click="filterActive = opt"
            :class="['px-3 py-2 text-sm font-medium transition-colors capitalize',
              filterActive === opt
                ? 'bg-primary-600 text-white'
                : 'text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700']">
            {{ opt }}
          </button>
        </div>
        <button @click="load" class="p-2 rounded-lg border border-slate-200 dark:border-slate-600
               text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
          <RefreshCw :size="16" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>
    </div>

    <!-- Table -->
    <div class="bg-white dark:bg-slate-800 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      <div v-if="loading" class="flex items-center justify-center h-48">
        <RefreshCw :size="24" class="animate-spin text-primary-500" />
      </div>
      <div v-else-if="filtered.length === 0" class="flex flex-col items-center justify-center h-48 text-slate-400">
        <Building2 :size="40" class="mb-3 opacity-30" />
        <p class="font-medium">No suppliers found</p>
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead class="bg-slate-50 dark:bg-slate-700/50 border-b border-slate-200 dark:border-slate-600">
            <tr>
              <th v-for="col in [['code','Code'],['name','Name'],['city','City/Wilaya'],['nif','NIF'],['payment_terms','Terms'],['balance','Balance'],['rating','Rating'],['is_active','Status']]"
                  :key="col[0]"
                  @click="toggleSort(col[0] as keyof Supplier)"
                  class="px-4 py-3 text-left text-xs font-semibold text-slate-500 dark:text-slate-400
                         uppercase tracking-wide cursor-pointer hover:text-slate-900 dark:hover:text-white select-none">
                <span class="inline-flex items-center gap-1">{{ col[1] }}
                  <span v-if="sortField === col[0]">
                    <ChevronUp v-if="sortDir === 'asc'" :size="12" />
                    <ChevronDown v-else :size="12" />
                  </span>
                </span>
              </th>
              <th class="px-4 py-3 text-right text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wide">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-for="s in filtered" :key="s.id"
              class="hover:bg-slate-50 dark:hover:bg-slate-700/30 transition-colors">
              <td class="px-4 py-3 font-mono text-xs text-slate-600 dark:text-slate-400">{{ s.code }}</td>
              <td class="px-4 py-3">
                <div class="font-semibold text-slate-900 dark:text-white">{{ s.name }}</div>
                <div v-if="s.contact_name" class="text-xs text-slate-500 dark:text-slate-400">{{ s.contact_name }}</div>
              </td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400">
                <div>{{ s.city || '—' }}</div>
                <div v-if="s.wilaya" class="text-xs text-slate-400">{{ s.wilaya }}</div>
              </td>
              <td class="px-4 py-3 font-mono text-xs text-slate-600 dark:text-slate-400">{{ s.nif || '—' }}</td>
              <td class="px-4 py-3 text-slate-600 dark:text-slate-400">{{ s.payment_terms }}d</td>
              <td class="px-4 py-3 font-semibold" :class="s.balance > 0 ? 'text-red-600 dark:text-red-400' : 'text-slate-600 dark:text-slate-400'">
                {{ fmt(s.balance || 0) }}
              </td>
              <td class="px-4 py-3">
                <span class="inline-flex items-center gap-1 font-semibold" :class="ratingColor(s.rating)">
                  <Star :size="12" />{{ s.rating || 3 }}/5
                </span>
              </td>
              <td class="px-4 py-3">
                <span :class="['inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium',
                  s.is_active
                    ? 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300'
                    : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400']">
                  <CheckCircle v-if="s.is_active" :size="10" />
                  <AlertTriangle v-else :size="10" />
                  {{ s.is_active ? 'Active' : 'Inactive' }}
                </span>
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-1">
                  <button @click="showDetail = s"
                    class="p-1.5 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors"
                    title="View details">
                    <ShieldCheck :size="15" />
                  </button>
                  <button @click="openEdit(s)"
                    class="p-1.5 rounded-lg text-slate-500 hover:bg-primary-50 dark:hover:bg-primary-900/30 hover:text-primary-600 transition-colors"
                    title="Edit">
                    <Edit2 :size="15" />
                  </button>
                  <button v-if="s.is_active" @click="deactivate(s)"
                    class="p-1.5 rounded-lg text-slate-500 hover:bg-red-50 dark:hover:bg-red-900/30 hover:text-red-600 transition-colors"
                    title="Deactivate">
                    <Trash2 :size="15" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-if="!loading" class="px-4 py-2 border-t border-slate-100 dark:border-slate-700 text-xs text-slate-500 dark:text-slate-400">
        {{ filtered.length }} of {{ suppliers.length }} suppliers
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showForm" class="fixed inset-0 z-50 flex items-start justify-center bg-black/50 backdrop-blur-sm p-4 overflow-y-auto">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-2xl my-8">
          <!-- Modal Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700">
            <h2 class="text-lg font-bold text-slate-900 dark:text-white">
              {{ isEdit ? 'Edit Supplier' : 'New Supplier' }}
            </h2>
            <button @click="showForm = false" class="p-2 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors">
              <X :size="18" />
            </button>
          </div>

          <div class="p-6 space-y-5">
            <!-- Basic Info -->
            <div>
              <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3 pb-1 border-b border-slate-100 dark:border-slate-700">
                Basic Information
              </h3>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Code <span class="text-red-500">*</span></label>
                  <input v-model="form.code" placeholder="SUP-001"
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white
                           focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Type</label>
                  <select v-model="form.type"
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500">
                    <option value="company">Company</option>
                    <option value="individual">Individual</option>
                  </select>
                </div>
                <div class="col-span-2">
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Name <span class="text-red-500">*</span></label>
                  <input v-model="form.name" placeholder="Supplier company name"
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white
                           focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Contact Name</label>
                  <input v-model="form.contact_name" placeholder="Contact person"
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white
                           focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Rating</label>
                  <select v-model.number="form.rating"
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500">
                    <option v-for="r in [1,2,3,4,5]" :key="r" :value="r">{{ r }} / 5</option>
                  </select>
                </div>
              </div>
            </div>

            <!-- Fiscal IDs (Algerian) -->
            <div>
              <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3 pb-1 border-b border-slate-100 dark:border-slate-700">
                Fiscal Identifiers
              </h3>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">NIF</label>
                  <input v-model="form.nif" placeholder="Numéro d'Identification Fiscale"
                    class="w-full px-3 py-2 text-sm font-mono rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">NIS</label>
                  <input v-model="form.nis" placeholder="Numéro d'Identification Statistique"
                    class="w-full px-3 py-2 text-sm font-mono rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">RC</label>
                  <input v-model="form.rc" placeholder="Registre de Commerce"
                    class="w-full px-3 py-2 text-sm font-mono rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">ART</label>
                  <input v-model="form.art" placeholder="Article"
                    class="w-full px-3 py-2 text-sm font-mono rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
              </div>
            </div>

            <!-- Contact -->
            <div>
              <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3 pb-1 border-b border-slate-100 dark:border-slate-700">
                Contact & Address
              </h3>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Phone</label>
                  <input v-model="form.phone" placeholder="+213 ..."
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Email</label>
                  <input v-model="form.email" type="email" placeholder="supplier@example.com"
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">City</label>
                  <input v-model="form.city" placeholder="City"
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Wilaya</label>
                  <select v-model="form.wilaya"
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500">
                    <option value="">— Select Wilaya —</option>
                    <option v-for="w in WILAYAS" :key="w" :value="w">{{ w }}</option>
                  </select>
                </div>
                <div class="col-span-2">
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Address</label>
                  <textarea v-model="form.address" rows="2" placeholder="Full address"
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none" />
                </div>
              </div>
            </div>

            <!-- Financial -->
            <div>
              <h3 class="text-sm font-semibold text-slate-700 dark:text-slate-300 mb-3 pb-1 border-b border-slate-100 dark:border-slate-700">
                Financial Terms
              </h3>
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Payment Terms (days)</label>
                  <input v-model.number="form.payment_terms" type="number" min="0"
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
                <div>
                  <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Credit Limit (DZD)</label>
                  <input v-model.number="form.credit_limit" type="number" min="0"
                    class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                           bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500" />
                </div>
              </div>
            </div>

            <!-- Notes -->
            <div>
              <label class="block text-xs font-medium text-slate-600 dark:text-slate-400 mb-1">Notes</label>
              <textarea v-model="form.notes" rows="2"
                class="w-full px-3 py-2 text-sm rounded-lg border border-slate-200 dark:border-slate-600
                       bg-white dark:bg-slate-700 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary-500 resize-none" />
            </div>
          </div>

          <!-- Modal Footer -->
          <div class="flex justify-end gap-3 px-6 py-4 border-t border-slate-200 dark:border-slate-700">
            <button @click="showForm = false"
              class="px-4 py-2 text-sm font-medium rounded-lg border border-slate-200 dark:border-slate-600
                     text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors">
              Cancel
            </button>
            <button @click="save" :disabled="saving"
              class="inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-lg
                     bg-primary-600 text-white hover:bg-primary-700 disabled:opacity-50 transition-colors shadow-sm">
              <RefreshCw v-if="saving" :size="15" class="animate-spin" />
              <Save v-else :size="15" />
              {{ saving ? 'Saving...' : 'Save Supplier' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Detail Drawer -->
    <Teleport to="body">
      <div v-if="showDetail" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
        <div class="bg-white dark:bg-slate-800 rounded-2xl shadow-2xl w-full max-w-lg">
          <div class="flex items-center justify-between px-6 py-4 border-b border-slate-200 dark:border-slate-700">
            <h2 class="text-lg font-bold text-slate-900 dark:text-white">{{ showDetail.name }}</h2>
            <button @click="showDetail = null" class="p-2 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700">
              <X :size="18" />
            </button>
          </div>
          <div class="p-6 space-y-4">
            <!-- Fiscal IDs grid -->
            <div class="grid grid-cols-2 gap-3">
              <div v-for="field in [['NIF', showDetail.nif],['NIS', showDetail.nis],['RC', showDetail.rc],['ART', showDetail.art]]"
                   :key="field[0]"
                   class="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-3">
                <p class="text-xs font-medium text-slate-500 dark:text-slate-400">{{ field[0] }}</p>
                <p class="font-mono font-semibold text-slate-900 dark:text-white text-sm mt-0.5">{{ field[1] || '—' }}</p>
              </div>
            </div>
            <!-- Contact -->
            <div class="space-y-2">
              <div v-if="showDetail.phone" class="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
                <Phone :size="14" class="text-slate-400 flex-shrink-0" />{{ showDetail.phone }}
              </div>
              <div v-if="showDetail.email" class="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
                <Mail :size="14" class="text-slate-400 flex-shrink-0" />{{ showDetail.email }}
              </div>
              <div v-if="showDetail.address" class="flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
                <MapPin :size="14" class="text-slate-400 flex-shrink-0" />{{ showDetail.address }}, {{ showDetail.city }}
              </div>
            </div>
            <!-- Terms -->
            <div class="grid grid-cols-3 gap-3 text-center">
              <div class="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-3">
                <p class="text-xs text-slate-500 dark:text-slate-400">Payment Terms</p>
                <p class="font-bold text-slate-900 dark:text-white">{{ showDetail.payment_terms }}d</p>
              </div>
              <div class="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-3">
                <p class="text-xs text-slate-500 dark:text-slate-400">Balance</p>
                <p class="font-bold" :class="(showDetail.balance||0) > 0 ? 'text-red-600 dark:text-red-400' : 'text-slate-900 dark:text-white'">
                  {{ fmt(showDetail.balance || 0) }}
                </p>
              </div>
              <div class="bg-slate-50 dark:bg-slate-700/50 rounded-lg p-3">
                <p class="text-xs text-slate-500 dark:text-slate-400">Rating</p>
                <p class="font-bold" :class="ratingColor(showDetail.rating)">{{ showDetail.rating }}/5</p>
              </div>
            </div>
          </div>
          <div class="flex justify-end gap-2 px-6 py-4 border-t border-slate-200 dark:border-slate-700">
            <button @click="openEdit(showDetail!); showDetail = null"
              class="inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold rounded-lg
                     bg-primary-600 text-white hover:bg-primary-700 transition-colors">
              <Edit2 :size="14" /> Edit
            </button>
          </div>
        </div>
      </div>
    </Teleport>

  </div>
</template>
