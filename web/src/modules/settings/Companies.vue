<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { settingsAPI } from '@/api/client'
import {
  Building2, Plus, Search, Edit2, X, Save, Globe, Phone,
  Mail, MapPin, RefreshCw, CheckCircle, XCircle, ChevronDown
} from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(false)
const saving = ref(false)
const search = ref('')
const companies = ref<any[]>([])
const showModal = ref(false)
const editing = ref<any>(null)

const form = ref({
  name: '', legal_name: '', nif: '', nis: '', rc: '', art: '',
  address: '', city: '', wilaya: '', postal_code: '',
  phone: '', email: '', website: '', logo_url: '',
  country: 'DZ', timezone: 'Africa/Algiers', default_currency: 'DZD',
  is_active: true
})

const wilayas = [
  'Adrar','Chlef','Laghouat','Oum El Bouaghi','Batna','Béjaïa','Biskra','Béchar',
  'Blida','Bouira','Tamanrasset','Tébessa','Tlemcen','Tiaret','Tizi Ouzou','Alger',
  'Djelfa','Jijel','Sétif','Saïda','Skikda','Sidi Bel Abbès','Annaba','Guelma',
  'Constantine','Médéa','Mostaganem','M\'Sila','Mascara','Ouargla','Oran','El Bayadh',
  'Illizi','Bordj Bou Arréridj','Boumerdès','El Tarf','Tindouf','Tissemsilt',
  'El Oued','Khenchela','Souk Ahras','Tipaza','Mila','Aïn Defla','Naâma',
  'Aïn Témouchent','Ghardaïa','Relizane'
]

const timezones = [
  { value: 'Africa/Algiers', label: 'Africa/Algiers (UTC+1)' },
  { value: 'Europe/Paris', label: 'Europe/Paris (UTC+1/+2)' },
  { value: 'UTC', label: 'UTC' },
]

const currencies = [
  { value: 'DZD', label: 'DZD - Algerian Dinar' },
  { value: 'EUR', label: 'EUR - Euro' },
  { value: 'USD', label: 'USD - US Dollar' },
  { value: 'GBP', label: 'GBP - British Pound' },
]

const filtered = computed(() => {
  if (!search.value) return companies.value
  const q = search.value.toLowerCase()
  return companies.value.filter(c =>
    c.name?.toLowerCase().includes(q) ||
    c.city?.toLowerCase().includes(q) ||
    c.nif?.toLowerCase().includes(q) ||
    c.email?.toLowerCase().includes(q)
  )
})

async function load() {
  loading.value = true
  try {
    const r = await settingsAPI.getCompanies()
    companies.value = Array.isArray(r.data) ? r.data : []
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error loading companies', 'error')
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editing.value = null
  form.value = {
    name: '', legal_name: '', nif: '', nis: '', rc: '', art: '',
    address: '', city: '', wilaya: '', postal_code: '',
    phone: '', email: '', website: '', logo_url: '',
    country: 'DZ', timezone: 'Africa/Algiers', default_currency: 'DZD',
    is_active: true
  }
  showModal.value = true
}

function openEdit(c: any) {
  editing.value = c
  form.value = { ...c }
  showModal.value = true
}

async function save() {
  if (!form.value.name.trim()) { app.addToast('Company name is required', 'error'); return }
  saving.value = true
  try {
    if (editing.value) {
      await settingsAPI.updateCompany(editing.value.id, form.value)
      app.addToast('Company updated successfully', 'success')
    } else {
      await settingsAPI.createCompany(form.value)
      app.addToast('Company created successfully', 'success')
    }
    showModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error saving company', 'error')
  } finally {
    saving.value = false
  }
}

const dk = (d: string, l: string) => app.darkMode ? d : l

onMounted(load)
</script>

<template>
  <div :class="dk('bg-slate-950 text-slate-100','bg-slate-50 text-slate-900')" class="min-h-screen">
    <!-- Header -->
    <div :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border-b px-6 py-4 sticky top-0 z-10">
      <div class="flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-3">
          <div class="p-2 rounded-xl bg-indigo-600/20">
            <Building2 class="w-5 h-5 text-indigo-400" />
          </div>
          <div>
            <h1 class="text-lg font-bold">Companies</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">Manage company profiles and settings</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
            class="p-2 rounded-lg transition-colors disabled:opacity-50">
            <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          </button>
          <button @click="openCreate"
            class="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-sm font-medium transition-colors">
            <Plus class="w-4 h-4" /> Add Company
          </button>
        </div>
      </div>
      <!-- Search -->
      <div class="mt-3 relative max-w-md">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
        <input v-model="search" placeholder="Search companies..."
          :class="dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500','bg-slate-100 border-slate-200 text-slate-900 placeholder-slate-400')"
          class="w-full pl-9 pr-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
      </div>
    </div>

    <!-- Content -->
    <div class="p-6">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="animate-spin w-8 h-8 text-indigo-400" />
      </div>

      <div v-else-if="filtered.length === 0" :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
        class="border rounded-xl p-16 text-center">
        <Building2 :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12 mx-auto mb-4" />
        <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">
          {{ search ? 'No companies match your search' : 'No companies found. Create one to get started.' }}
        </p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-5">
        <div v-for="c in filtered" :key="c.id"
          :class="dk('bg-slate-900 border-slate-800 hover:border-slate-700','bg-white border-slate-200 hover:border-slate-300')"
          class="border rounded-xl overflow-hidden transition-all hover:shadow-lg group">
          <!-- Card header -->
          <div class="px-5 pt-5 pb-4">
            <div class="flex items-start justify-between gap-2 mb-3">
              <div class="flex items-center gap-3 min-w-0">
                <div class="w-10 h-10 rounded-xl bg-indigo-600/20 flex items-center justify-center flex-shrink-0">
                  <Building2 class="w-5 h-5 text-indigo-400" />
                </div>
                <div class="min-w-0">
                  <h3 class="font-semibold text-sm truncate">{{ c.name }}</h3>
                  <p v-if="c.legal_name" :class="dk('text-slate-400','text-slate-500')" class="text-xs truncate">{{ c.legal_name }}</p>
                </div>
              </div>
              <span :class="c.is_active ? 'bg-emerald-500/10 text-emerald-500' : 'bg-red-500/10 text-red-500'"
                class="text-xs px-2 py-0.5 rounded-full font-medium flex-shrink-0 flex items-center gap-1">
                <CheckCircle v-if="c.is_active" class="w-3 h-3" />
                <XCircle v-else class="w-3 h-3" />
                {{ c.is_active ? 'Active' : 'Inactive' }}
              </span>
            </div>

            <!-- Info grid -->
            <div class="space-y-1.5 mt-3">
              <div v-if="c.email" class="flex items-center gap-2 text-xs" :class="dk('text-slate-400','text-slate-600')">
                <Mail class="w-3.5 h-3.5 flex-shrink-0" /> <span class="truncate">{{ c.email }}</span>
              </div>
              <div v-if="c.phone" class="flex items-center gap-2 text-xs" :class="dk('text-slate-400','text-slate-600')">
                <Phone class="w-3.5 h-3.5 flex-shrink-0" /> <span>{{ c.phone }}</span>
              </div>
              <div v-if="c.city || c.wilaya" class="flex items-center gap-2 text-xs" :class="dk('text-slate-400','text-slate-600')">
                <MapPin class="w-3.5 h-3.5 flex-shrink-0" />
                <span class="truncate">{{ [c.city, c.wilaya].filter(Boolean).join(', ') }}</span>
              </div>
              <div v-if="c.website" class="flex items-center gap-2 text-xs" :class="dk('text-slate-400','text-slate-600')">
                <Globe class="w-3.5 h-3.5 flex-shrink-0" /> <span class="truncate">{{ c.website }}</span>
              </div>
            </div>
          </div>

          <!-- Tax IDs row -->
          <div v-if="c.nif || c.nis || c.rc" :class="dk('bg-slate-800/50 border-slate-800','bg-slate-50 border-slate-100')"
            class="px-5 py-2.5 border-t flex items-center gap-4 flex-wrap">
            <div v-if="c.nif" class="text-xs">
              <span :class="dk('text-slate-500','text-slate-400')">NIF: </span>
              <span :class="dk('text-slate-300','text-slate-700')" class="font-mono">{{ c.nif }}</span>
            </div>
            <div v-if="c.nis" class="text-xs">
              <span :class="dk('text-slate-500','text-slate-400')">NIS: </span>
              <span :class="dk('text-slate-300','text-slate-700')" class="font-mono">{{ c.nis }}</span>
            </div>
            <div v-if="c.rc" class="text-xs">
              <span :class="dk('text-slate-500','text-slate-400')">RC: </span>
              <span :class="dk('text-slate-300','text-slate-700')" class="font-mono">{{ c.rc }}</span>
            </div>
          </div>

          <!-- Footer -->
          <div :class="dk('border-slate-800','border-slate-100')" class="px-5 py-3 border-t flex items-center justify-between">
            <div class="flex items-center gap-3 text-xs" :class="dk('text-slate-500','text-slate-400')">
              <span class="font-mono bg-indigo-500/10 text-indigo-400 px-2 py-0.5 rounded">{{ c.default_currency || 'DZD' }}</span>
              <span>{{ c.timezone?.replace('Africa/', '').replace('Europe/', '') || 'Algiers' }}</span>
            </div>
            <button @click="openEdit(c)"
              :class="dk('text-slate-400 hover:text-white hover:bg-slate-700','text-slate-500 hover:text-slate-900 hover:bg-slate-100')"
              class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors">
              <Edit2 class="w-3.5 h-3.5" /> Edit
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4"
        style="background:rgba(0,0,0,0.7)" @click.self="showModal=false">
        <div :class="dk('bg-slate-900 border-slate-700','bg-white border-slate-200')"
          class="border rounded-2xl w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col shadow-2xl">
          <!-- Modal header -->
          <div :class="dk('border-slate-800','border-slate-100')" class="flex items-center justify-between px-6 py-4 border-b flex-shrink-0">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-xl bg-indigo-600/20">
                <Building2 class="w-4 h-4 text-indigo-400" />
              </div>
              <h2 class="font-bold">{{ editing ? 'Edit Company' : 'New Company' }}</h2>
            </div>
            <button @click="showModal=false" :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-900')"
              class="p-1.5 rounded-lg transition-colors hover:bg-slate-800/50">
              <X class="w-5 h-5" />
            </button>
          </div>

          <!-- Modal body -->
          <div class="overflow-y-auto flex-1 px-6 py-5">
            <div class="space-y-5">
              <!-- Basic info -->
              <div>
                <h3 class="text-xs font-semibold uppercase tracking-wider text-indigo-400 mb-3">Basic Information</h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div class="sm:col-span-2">
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">
                      Company Name <span class="text-red-400">*</span>
                    </label>
                    <input v-model="form.name" placeholder="Acme Corporation"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Legal Name</label>
                    <input v-model="form.legal_name" placeholder="Acme Corporation EURL"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Default Currency</label>
                    <select v-model="form.default_currency"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                      <option v-for="c in currencies" :key="c.value" :value="c.value">{{ c.label }}</option>
                    </select>
                  </div>
                </div>
              </div>

              <!-- Algerian Tax IDs -->
              <div>
                <h3 class="text-xs font-semibold uppercase tracking-wider text-indigo-400 mb-3">Algerian Tax Identifiers</h3>
                <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">NIF</label>
                    <input v-model="form.nif" placeholder="000000000000000"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100 font-mono','bg-white border-slate-300 text-slate-900 font-mono')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">NIS</label>
                    <input v-model="form.nis" placeholder="000000000000000"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100 font-mono','bg-white border-slate-300 text-slate-900 font-mono')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">RC</label>
                    <input v-model="form.rc" placeholder="16/00-0000000"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100 font-mono','bg-white border-slate-300 text-slate-900 font-mono')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">ART</label>
                    <input v-model="form.art" placeholder="000000000"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100 font-mono','bg-white border-slate-300 text-slate-900 font-mono')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                </div>
              </div>

              <!-- Contact -->
              <div>
                <h3 class="text-xs font-semibold uppercase tracking-wider text-indigo-400 mb-3">Contact Information</h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Phone</label>
                    <input v-model="form.phone" placeholder="+213 21 000000"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Email</label>
                    <input v-model="form.email" type="email" placeholder="contact@company.com"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                  <div class="sm:col-span-2">
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Website</label>
                    <input v-model="form.website" placeholder="https://www.company.com"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                </div>
              </div>

              <!-- Address -->
              <div>
                <h3 class="text-xs font-semibold uppercase tracking-wider text-indigo-400 mb-3">Address</h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div class="sm:col-span-2">
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Street Address</label>
                    <input v-model="form.address" placeholder="123 Rue des Entrepreneurs"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">City</label>
                    <input v-model="form.city" placeholder="Algiers"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Wilaya</label>
                    <select v-model="form.wilaya"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                      <option value="">Select wilaya</option>
                      <option v-for="w in wilayas" :key="w" :value="w">{{ w }}</option>
                    </select>
                  </div>
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Postal Code</label>
                    <input v-model="form.postal_code" placeholder="16000"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                  <div>
                    <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Timezone</label>
                    <select v-model="form.timezone"
                      :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                      class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500">
                      <option v-for="tz in timezones" :key="tz.value" :value="tz.value">{{ tz.label }}</option>
                    </select>
                  </div>
                </div>
              </div>

              <!-- Status -->
              <div class="flex items-center gap-3 pt-1">
                <button type="button" @click="form.is_active = !form.is_active"
                  :class="form.is_active ? 'bg-emerald-500' : 'bg-slate-600'"
                  class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors">
                  <span :class="form.is_active ? 'translate-x-6' : 'translate-x-1'"
                    class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
                </button>
                <span class="text-sm" :class="dk('text-slate-300','text-slate-700')">
                  Company is {{ form.is_active ? 'Active' : 'Inactive' }}
                </span>
              </div>
            </div>
          </div>

          <!-- Modal footer -->
          <div :class="dk('border-slate-800 bg-slate-900/50','border-slate-100 bg-slate-50')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t flex-shrink-0">
            <button @click="showModal=false"
              :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
              class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="save" :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors">
              <Save class="w-4 h-4" />
              {{ saving ? 'Saving...' : (editing ? 'Update Company' : 'Create Company') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
