<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAppStore } from '@/stores/app'
import { settingsAPI } from '@/api/client'
import {
  Users, Plus, Search, Edit2, X, Save, RefreshCw, Shield,
  UserCheck, UserX, Key, Trash2, Mail, Phone, Lock, Eye, EyeOff,
  ChevronDown, CheckCircle, XCircle
} from 'lucide-vue-next'

const app = useAppStore()
const loading = ref(false)
const saving = ref(false)
const search = ref('')
const activeTab = ref<'users' | 'roles'>('users')
const users = ref<any[]>([])
const roles = ref<any[]>([])
const showUserModal = ref(false)
const showRoleModal = ref(false)
const editingUser = ref<any>(null)
const editingRole = ref<any>(null)
const showPassword = ref(false)

const systemRoles = ['super_admin','admin','accountant','hr_manager','sales_manager','purchase_manager','inventory_manager','project_manager','viewer','approver']

const userForm = ref({
  username: '', email: '', password: '', full_name: '',
  role: 'viewer', role_id: '', phone: '', avatar_url: '', is_active: true
})

const roleForm = ref({
  name: '', description: '',
  permissions: [] as string[], is_active: true
})

const modulePermissions = [
  { key: 'accounting', label: 'Accounting' },
  { key: 'sales', label: 'Sales & CRM' },
  { key: 'purchase', label: 'Purchase' },
  { key: 'inventory', label: 'Inventory' },
  { key: 'hr', label: 'Human Resources' },
  { key: 'manufacturing', label: 'Manufacturing' },
  { key: 'projects', label: 'Projects' },
  { key: 'treasury', label: 'Treasury' },
  { key: 'tax', label: 'Tax' },
  { key: 'reports', label: 'Reports' },
  { key: 'fleet', label: 'Fleet' },
  { key: 'maintenance', label: 'Maintenance' },
  { key: 'settings', label: 'Settings' },
]

const filteredUsers = computed(() => {
  if (!search.value) return users.value
  const q = search.value.toLowerCase()
  return users.value.filter(u =>
    u.full_name?.toLowerCase().includes(q) ||
    u.username?.toLowerCase().includes(q) ||
    u.email?.toLowerCase().includes(q) ||
    u.role?.toLowerCase().includes(q)
  )
})

const filteredRoles = computed(() => {
  if (!search.value) return roles.value
  const q = search.value.toLowerCase()
  return roles.value.filter(r => r.name?.toLowerCase().includes(q))
})

async function load() {
  loading.value = true
  try {
    const [ur, rr] = await Promise.all([settingsAPI.getUsers(), settingsAPI.getRoles()])
    users.value = Array.isArray(ur.data) ? ur.data : []
    roles.value = Array.isArray(rr.data) ? rr.data : []
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error loading users', 'error')
  } finally {
    loading.value = false
  }
}

function openCreateUser() {
  editingUser.value = null
  userForm.value = { username: '', email: '', password: '', full_name: '', role: 'viewer', role_id: '', phone: '', avatar_url: '', is_active: true }
  showPassword.value = false
  showUserModal.value = true
}

function openEditUser(u: any) {
  editingUser.value = u
  userForm.value = { ...u, password: '' }
  showPassword.value = false
  showUserModal.value = true
}

function openCreateRole() {
  editingRole.value = null
  roleForm.value = { name: '', description: '', permissions: [], is_active: true }
  showRoleModal.value = true
}

function openEditRole(r: any) {
  editingRole.value = r
  roleForm.value = {
    name: r.name, description: r.description || '',
    permissions: Array.isArray(r.permissions) ? r.permissions : [],
    is_active: r.is_active
  }
  showRoleModal.value = true
}

function togglePermission(key: string) {
  const idx = roleForm.value.permissions.indexOf(key)
  if (idx >= 0) roleForm.value.permissions.splice(idx, 1)
  else roleForm.value.permissions.push(key)
}

async function saveUser() {
  if (!userForm.value.username || !userForm.value.email) {
    app.addToast('Username and email are required', 'error'); return
  }
  if (!editingUser.value && !userForm.value.password) {
    app.addToast('Password is required for new users', 'error'); return
  }
  saving.value = true
  try {
    if (editingUser.value) {
      const payload: any = { ...userForm.value }
      if (!payload.password) delete payload.password
      await settingsAPI.updateUser(editingUser.value.id, payload)
      app.addToast('User updated', 'success')
    } else {
      await settingsAPI.createUser(userForm.value)
      app.addToast('User created', 'success')
    }
    showUserModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error saving user', 'error')
  } finally {
    saving.value = false
  }
}

async function deleteUser(u: any) {
  if (!confirm(`Deactivate user "${u.full_name || u.username}"?`)) return
  try {
    await settingsAPI.deleteUser(u.id)
    app.addToast('User deactivated', 'success')
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error', 'error')
  }
}

async function saveRole() {
  if (!roleForm.value.name.trim()) { app.addToast('Role name is required', 'error'); return }
  saving.value = true
  try {
    if (editingRole.value) {
      await settingsAPI.updateRole(editingRole.value.id, roleForm.value)
      app.addToast('Role updated', 'success')
    } else {
      await settingsAPI.createRole(roleForm.value)
      app.addToast('Role created', 'success')
    }
    showRoleModal.value = false
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error saving role', 'error')
  } finally {
    saving.value = false
  }
}

async function deleteRole(r: any) {
  if (r.is_system) { app.addToast('System roles cannot be deleted', 'error'); return }
  if (!confirm(`Delete role "${r.name}"?`)) return
  try {
    await settingsAPI.deleteRole(r.id)
    app.addToast('Role deleted', 'success')
    await load()
  } catch (e: any) {
    app.addToast(e?.response?.data?.error || 'Error', 'error')
  }
}

const roleColorMap: Record<string, string> = {
  super_admin: 'bg-red-500/10 text-red-400',
  admin: 'bg-orange-500/10 text-orange-400',
  accountant: 'bg-blue-500/10 text-blue-400',
  hr_manager: 'bg-purple-500/10 text-purple-400',
  sales_manager: 'bg-emerald-500/10 text-emerald-400',
  viewer: 'bg-slate-500/10 text-slate-400',
}
function roleColor(role: string) {
  return roleColorMap[role] || 'bg-indigo-500/10 text-indigo-400'
}
function initials(name: string) {
  return name?.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2) || '?'
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
          <div class="p-2 rounded-xl bg-violet-600/20">
            <Users class="w-5 h-5 text-violet-400" />
          </div>
          <div>
            <h1 class="text-lg font-bold">Users &amp; Roles</h1>
            <p :class="dk('text-slate-400','text-slate-500')" class="text-xs">Manage system users and access roles</p>
          </div>
        </div>
        <div class="flex items-center gap-2">
          <button @click="load" :disabled="loading"
            :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
            class="p-2 rounded-lg transition-colors disabled:opacity-50">
            <RefreshCw :class="loading ? 'animate-spin' : ''" class="w-4 h-4" />
          </button>
          <button v-if="activeTab==='users'" @click="openCreateUser"
            class="flex items-center gap-2 px-4 py-2 bg-violet-600 hover:bg-violet-500 text-white rounded-lg text-sm font-medium transition-colors">
            <Plus class="w-4 h-4" /> Add User
          </button>
          <button v-else @click="openCreateRole"
            class="flex items-center gap-2 px-4 py-2 bg-violet-600 hover:bg-violet-500 text-white rounded-lg text-sm font-medium transition-colors">
            <Plus class="w-4 h-4" /> Add Role
          </button>
        </div>
      </div>

      <!-- Tabs + Search -->
      <div class="mt-3 flex items-center gap-4">
        <div :class="dk('bg-slate-800','bg-slate-100')" class="flex rounded-lg p-1 gap-1">
          <button v-for="tab in ['users','roles'] as const" :key="tab"
            @click="activeTab=tab; search=''"
            :class="activeTab===tab
              ? 'bg-violet-600 text-white shadow'
              : dk('text-slate-400 hover:text-slate-200','text-slate-500 hover:text-slate-700')"
            class="px-4 py-1.5 rounded-md text-sm font-medium transition-all capitalize">
            {{ tab }}
            <span class="ml-1.5 text-xs opacity-70">{{ tab==='users' ? users.length : roles.length }}</span>
          </button>
        </div>
        <div class="relative flex-1 max-w-sm">
          <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none" />
          <input v-model="search" :placeholder="`Search ${activeTab}...`"
            :class="dk('bg-slate-800 border-slate-700 text-slate-100 placeholder-slate-500','bg-slate-100 border-slate-200 text-slate-900 placeholder-slate-400')"
            class="w-full pl-9 pr-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" />
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="p-6">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <RefreshCw class="animate-spin w-8 h-8 text-violet-400" />
      </div>

      <!-- Users Table -->
      <template v-else-if="activeTab==='users'">
        <div v-if="filteredUsers.length===0" :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
          class="border rounded-xl p-16 text-center">
          <Users :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12 mx-auto mb-4" />
          <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">No users found</p>
        </div>
        <div v-else :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')" class="border rounded-xl overflow-hidden">
          <table class="w-full text-sm">
            <thead>
              <tr :class="dk('bg-slate-800/50 text-slate-400 border-slate-800','bg-slate-50 text-slate-500 border-slate-100')" class="border-b text-xs uppercase tracking-wider">
                <th class="px-4 py-3 text-left font-medium">User</th>
                <th class="px-4 py-3 text-left font-medium">Email / Phone</th>
                <th class="px-4 py-3 text-left font-medium">System Role</th>
                <th class="px-4 py-3 text-left font-medium">Custom Role</th>
                <th class="px-4 py-3 text-center font-medium">Status</th>
                <th class="px-4 py-3 text-left font-medium">Last Login</th>
                <th class="px-4 py-3 text-right font-medium">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y" :class="dk('divide-slate-800','divide-slate-100')">
              <tr v-for="u in filteredUsers" :key="u.id"
                :class="dk('hover:bg-slate-800/50','hover:bg-slate-50')" class="transition-colors">
                <td class="px-4 py-3">
                  <div class="flex items-center gap-3">
                    <div class="w-9 h-9 rounded-full bg-violet-600/20 flex items-center justify-center flex-shrink-0">
                      <span class="text-xs font-bold text-violet-400">{{ initials(u.full_name || u.username) }}</span>
                    </div>
                    <div>
                      <div class="font-medium text-sm">{{ u.full_name || u.username }}</div>
                      <div :class="dk('text-slate-500','text-slate-400')" class="text-xs font-mono">@{{ u.username }}</div>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <div :class="dk('text-slate-300','text-slate-700')" class="text-xs">{{ u.email }}</div>
                  <div v-if="u.phone" :class="dk('text-slate-500','text-slate-400')" class="text-xs mt-0.5">{{ u.phone }}</div>
                </td>
                <td class="px-4 py-3">
                  <span :class="roleColor(u.role)" class="text-xs px-2 py-0.5 rounded-full font-medium capitalize">
                    {{ u.role?.replace('_', ' ') || '-' }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span v-if="u.role_name" class="text-xs px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 font-medium">
                    {{ u.role_name }}
                  </span>
                  <span v-else :class="dk('text-slate-600','text-slate-400')" class="text-xs">-</span>
                </td>
                <td class="px-4 py-3 text-center">
                  <span :class="u.is_active ? 'text-emerald-400' : 'text-red-400'">
                    <CheckCircle v-if="u.is_active" class="w-4 h-4 mx-auto" />
                    <XCircle v-else class="w-4 h-4 mx-auto" />
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span :class="dk('text-slate-500','text-slate-400')" class="text-xs">
                    {{ u.last_login ? new Date(u.last_login).toLocaleDateString('fr-DZ') : 'Never' }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center justify-end gap-1">
                    <button @click="openEditUser(u)"
                      :class="dk('hover:bg-slate-700 text-slate-400 hover:text-white','hover:bg-slate-100 text-slate-500 hover:text-slate-900')"
                      class="p-1.5 rounded-lg transition-colors" title="Edit">
                      <Edit2 class="w-3.5 h-3.5" />
                    </button>
                    <button @click="deleteUser(u)"
                      class="p-1.5 rounded-lg hover:bg-red-500/10 text-red-400 transition-colors" title="Deactivate">
                      <UserX class="w-3.5 h-3.5" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>

      <!-- Roles Grid -->
      <template v-else>
        <div v-if="filteredRoles.length===0" :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
          class="border rounded-xl p-16 text-center">
          <Shield :class="dk('text-slate-600','text-slate-300')" class="w-12 h-12 mx-auto mb-4" />
          <p :class="dk('text-slate-400','text-slate-500')" class="text-sm">No roles found</p>
        </div>
        <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <div v-for="r in filteredRoles" :key="r.id"
            :class="dk('bg-slate-900 border-slate-800','bg-white border-slate-200')"
            class="border rounded-xl p-5">
            <div class="flex items-start justify-between mb-3">
              <div class="flex items-center gap-2.5">
                <div class="p-2 rounded-xl" :class="r.is_system ? 'bg-amber-500/10' : 'bg-indigo-500/10'">
                  <Shield :class="r.is_system ? 'text-amber-400' : 'text-indigo-400'" class="w-4 h-4" />
                </div>
                <div>
                  <div class="font-semibold text-sm">{{ r.name }}</div>
                  <div v-if="r.is_system" class="text-xs text-amber-400 mt-0.5">System Role</div>
                </div>
              </div>
              <div class="flex gap-1">
                <button v-if="!r.is_system" @click="openEditRole(r)"
                  :class="dk('hover:bg-slate-700 text-slate-400 hover:text-white','hover:bg-slate-100 text-slate-500 hover:text-slate-900')"
                  class="p-1.5 rounded-lg transition-colors">
                  <Edit2 class="w-3.5 h-3.5" />
                </button>
                <button v-if="!r.is_system" @click="deleteRole(r)"
                  class="p-1.5 rounded-lg hover:bg-red-500/10 text-red-400 transition-colors">
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
            <p v-if="r.description" :class="dk('text-slate-400','text-slate-500')" class="text-xs mb-3 leading-relaxed">
              {{ r.description }}
            </p>
            <div class="flex flex-wrap gap-1.5 mt-2">
              <span v-for="perm in (Array.isArray(r.permissions) ? r.permissions.slice(0,6) : [])" :key="perm"
                :class="dk('bg-slate-800 text-slate-300','bg-slate-100 text-slate-600')"
                class="text-xs px-2 py-0.5 rounded font-mono">{{ perm }}</span>
              <span v-if="Array.isArray(r.permissions) && r.permissions.length > 6"
                :class="dk('text-slate-500','text-slate-400')" class="text-xs px-2 py-0.5">
                +{{ r.permissions.length - 6 }} more
              </span>
            </div>
            <div class="mt-3 pt-3 border-t" :class="dk('border-slate-800','border-slate-100')">
              <span :class="r.is_active ? 'text-emerald-400' : 'text-red-400'" class="text-xs flex items-center gap-1">
                <CheckCircle v-if="r.is_active" class="w-3.5 h-3.5" />
                <XCircle v-else class="w-3.5 h-3.5" />
                {{ r.is_active ? 'Active' : 'Inactive' }}
              </span>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- User Modal -->
    <Teleport to="body">
      <div v-if="showUserModal" class="fixed inset-0 z-50 flex items-center justify-center p-4"
        style="background:rgba(0,0,0,0.7)" @click.self="showUserModal=false">
        <div :class="dk('bg-slate-900 border-slate-700','bg-white border-slate-200')"
          class="border rounded-2xl w-full max-w-lg max-h-[90vh] overflow-hidden flex flex-col shadow-2xl">
          <div :class="dk('border-slate-800','border-slate-100')" class="flex items-center justify-between px-6 py-4 border-b flex-shrink-0">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-xl bg-violet-600/20">
                <Users class="w-4 h-4 text-violet-400" />
              </div>
              <h2 class="font-bold">{{ editingUser ? 'Edit User' : 'New User' }}</h2>
            </div>
            <button @click="showUserModal=false" :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-900')"
              class="p-1.5 rounded-lg transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="overflow-y-auto flex-1 px-6 py-5 space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Username <span class="text-red-400">*</span></label>
                <input v-model="userForm.username" :disabled="!!editingUser" placeholder="john.doe"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500 disabled:opacity-50" />
              </div>
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Full Name</label>
                <input v-model="userForm.full_name" placeholder="John Doe"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" />
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Email <span class="text-red-400">*</span></label>
              <input v-model="userForm.email" type="email" placeholder="john@company.com"
                :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">
                {{ editingUser ? 'New Password (leave blank to keep)' : 'Password' }}
                <span v-if="!editingUser" class="text-red-400">*</span>
              </label>
              <div class="relative">
                <input v-model="userForm.password" :type="showPassword ? 'text' : 'password'" placeholder="Min. 8 characters"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 pr-10 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" />
                <button type="button" @click="showPassword=!showPassword"
                  :class="dk('text-slate-400 hover:text-slate-200','text-slate-400 hover:text-slate-700')"
                  class="absolute right-3 top-1/2 -translate-y-1/2">
                  <Eye v-if="!showPassword" class="w-4 h-4" />
                  <EyeOff v-else class="w-4 h-4" />
                </button>
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">System Role</label>
                <select v-model="userForm.role"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500">
                  <option v-for="r in systemRoles" :key="r" :value="r">{{ r.replace(/_/g,' ') }}</option>
                </select>
              </div>
              <div>
                <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Custom Role</label>
                <select v-model="userForm.role_id"
                  :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                  class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500">
                  <option value="">None</option>
                  <option v-for="r in roles" :key="r.id" :value="r.id">{{ r.name }}</option>
                </select>
              </div>
            </div>
            <div>
              <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Phone</label>
              <input v-model="userForm.phone" placeholder="+213 555 000000"
                :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-violet-500" />
            </div>
            <div class="flex items-center gap-3">
              <button type="button" @click="userForm.is_active = !userForm.is_active"
                :class="userForm.is_active ? 'bg-emerald-500' : 'bg-slate-600'"
                class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors">
                <span :class="userForm.is_active ? 'translate-x-6' : 'translate-x-1'"
                  class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
              </button>
              <span class="text-sm" :class="dk('text-slate-300','text-slate-700')">User is Active</span>
            </div>
          </div>
          <div :class="dk('border-slate-800 bg-slate-900/50','border-slate-100 bg-slate-50')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t flex-shrink-0">
            <button @click="showUserModal=false"
              :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
              class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="saveUser" :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 bg-violet-600 hover:bg-violet-500 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors">
              <Save class="w-4 h-4" />
              {{ saving ? 'Saving...' : (editingUser ? 'Update User' : 'Create User') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Role Modal -->
    <Teleport to="body">
      <div v-if="showRoleModal" class="fixed inset-0 z-50 flex items-center justify-center p-4"
        style="background:rgba(0,0,0,0.7)" @click.self="showRoleModal=false">
        <div :class="dk('bg-slate-900 border-slate-700','bg-white border-slate-200')"
          class="border rounded-2xl w-full max-w-lg max-h-[90vh] overflow-hidden flex flex-col shadow-2xl">
          <div :class="dk('border-slate-800','border-slate-100')" class="flex items-center justify-between px-6 py-4 border-b flex-shrink-0">
            <div class="flex items-center gap-3">
              <div class="p-2 rounded-xl bg-amber-600/20">
                <Shield class="w-4 h-4 text-amber-400" />
              </div>
              <h2 class="font-bold">{{ editingRole ? 'Edit Role' : 'New Role' }}</h2>
            </div>
            <button @click="showRoleModal=false" :class="dk('text-slate-400 hover:text-white','text-slate-500 hover:text-slate-900')"
              class="p-1.5 rounded-lg transition-colors">
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="overflow-y-auto flex-1 px-6 py-5 space-y-4">
            <div>
              <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Role Name <span class="text-red-400">*</span></label>
              <input v-model="roleForm.name" placeholder="Finance Manager"
                :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amber-500" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-1.5" :class="dk('text-slate-300','text-slate-700')">Description</label>
              <textarea v-model="roleForm.description" rows="2" placeholder="Role description..."
                :class="dk('bg-slate-800 border-slate-600 text-slate-100','bg-white border-slate-300 text-slate-900')"
                class="w-full px-3 py-2 border rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-amber-500 resize-none" />
            </div>
            <div>
              <label class="block text-xs font-medium mb-2" :class="dk('text-slate-300','text-slate-700')">Module Permissions</label>
              <div class="grid grid-cols-2 gap-2">
                <button v-for="m in modulePermissions" :key="m.key" type="button"
                  @click="togglePermission(m.key)"
                  :class="roleForm.permissions.includes(m.key)
                    ? 'bg-amber-600/20 border-amber-500 text-amber-400'
                    : dk('bg-slate-800 border-slate-700 text-slate-400 hover:border-slate-500','bg-slate-50 border-slate-200 text-slate-500 hover:border-slate-300')"
                  class="flex items-center gap-2 px-3 py-2 border rounded-lg text-xs font-medium transition-all text-left">
                  <div :class="roleForm.permissions.includes(m.key) ? 'bg-amber-500' : dk('bg-slate-700','bg-slate-200')"
                    class="w-3.5 h-3.5 rounded-sm flex items-center justify-center flex-shrink-0 transition-colors">
                    <svg v-if="roleForm.permissions.includes(m.key)" class="w-2.5 h-2.5 text-white" fill="currentColor" viewBox="0 0 12 12">
                      <path d="M10 3L5 8L2 5" stroke="currentColor" stroke-width="1.5" fill="none" />
                    </svg>
                  </div>
                  {{ m.label }}
                </button>
              </div>
            </div>
            <div class="flex items-center gap-3">
              <button type="button" @click="roleForm.is_active = !roleForm.is_active"
                :class="roleForm.is_active ? 'bg-emerald-500' : 'bg-slate-600'"
                class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors">
                <span :class="roleForm.is_active ? 'translate-x-6' : 'translate-x-1'"
                  class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform" />
              </button>
              <span class="text-sm" :class="dk('text-slate-300','text-slate-700')">Role is Active</span>
            </div>
          </div>
          <div :class="dk('border-slate-800 bg-slate-900/50','border-slate-100 bg-slate-50')"
            class="flex items-center justify-end gap-3 px-6 py-4 border-t flex-shrink-0">
            <button @click="showRoleModal=false"
              :class="dk('bg-slate-800 hover:bg-slate-700 text-slate-300','bg-slate-100 hover:bg-slate-200 text-slate-700')"
              class="px-4 py-2 rounded-lg text-sm font-medium transition-colors">Cancel</button>
            <button @click="saveRole" :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 bg-amber-600 hover:bg-amber-500 disabled:opacity-50 text-white rounded-lg text-sm font-medium transition-colors">
              <Save class="w-4 h-4" />
              {{ saving ? 'Saving...' : (editingRole ? 'Update Role' : 'Create Role') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
