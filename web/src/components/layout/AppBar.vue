<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { workflowAPI } from '@/api/client'
import {
  Search,
  CheckSquare,
  Settings,
  LogOut,
  ChevronDown,
  Mail,
  Sun,
  Moon
} from 'lucide-vue-next'

const router = useRouter()
const auth = useAuthStore()
const app = useAppStore()

const pendingApprovals = ref(0)
const userMenuOpen = ref(false)

async function loadApprovals() {
  try {
    const res = await workflowAPI.getApprovalInbox()
    pendingApprovals.value = Array.isArray(res.data) ? res.data.length : 0
  } catch { /* silent */ }
}

async function handleLogout() {
  await auth.logout()
  router.push({ name: 'Login' })
}

onMounted(() => { loadApprovals() })
</script>

<template>
  <header
    class="h-14 border-b flex items-center justify-between px-4 gap-4 flex-shrink-0 transition-colors duration-200"
    :class="app.darkMode
      ? 'bg-slate-900 border-slate-700 text-slate-100'
      : 'bg-white border-slate-200 text-slate-800'"
  >
    <!-- Left: Page title -->
    <div class="flex items-center gap-3">
      <h1 class="text-sm font-semibold truncate" :class="app.darkMode ? 'text-slate-100' : 'text-slate-700'">
        {{ $route.meta?.title ?? 'Nexus ERP' }}
      </h1>
    </div>

    <!-- Center: Command palette trigger -->
    <button
      class="hidden md:flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg border transition-colors"
      :class="app.darkMode
        ? 'text-slate-400 bg-slate-800 border-slate-700 hover:bg-slate-700'
        : 'text-slate-400 bg-slate-50 border-slate-200 hover:bg-slate-100'"
      @click="app.toggleCommandPalette()"
    >
      <Search class="w-4 h-4" />
      <span>Search or navigate...</span>
      <kbd
        class="text-xs rounded px-1.5 py-0.5 font-mono"
        :class="app.darkMode ? 'bg-slate-700 border border-slate-600' : 'bg-white border border-slate-200'"
      >Ctrl+K</kbd>
    </button>

    <!-- Right: Actions -->
    <div class="flex items-center gap-1">

      <!-- Theme toggle -->
      <button
        class="p-2 rounded-lg transition-colors"
        :class="app.darkMode
          ? 'text-amber-400 hover:bg-slate-800'
          : 'text-slate-500 hover:bg-slate-100 hover:text-amber-500'"
        :title="app.darkMode ? 'Switch to light mode' : 'Switch to dark mode'"
        @click="app.toggleTheme()"
      >
        <Sun v-if="app.darkMode" class="w-5 h-5" />
        <Moon v-else class="w-5 h-5" />
      </button>

      <!-- Approvals -->
      <button
        class="relative p-2 rounded-lg transition-colors"
        :class="app.darkMode
          ? 'text-slate-400 hover:bg-slate-800 hover:text-slate-200'
          : 'text-slate-500 hover:bg-slate-100 hover:text-slate-700'"
        title="Approval Inbox"
        @click="router.push('/settings/workflow')"
      >
        <CheckSquare class="w-5 h-5" />
        <span
          v-if="pendingApprovals > 0"
          class="absolute -top-0.5 -right-0.5 bg-red-500 text-white text-[10px] font-bold rounded-full w-4 h-4 flex items-center justify-center"
        >{{ pendingApprovals > 9 ? '9+' : pendingApprovals }}</span>
      </button>

      <!-- User menu -->
      <div class="relative">
        <button
          class="flex items-center gap-2 p-1.5 pr-3 rounded-lg transition-colors"
          :class="app.darkMode ? 'hover:bg-slate-800' : 'hover:bg-slate-100'"
          @click="userMenuOpen = !userMenuOpen"
        >
          <div class="w-7 h-7 bg-indigo-500 rounded-full flex items-center justify-center text-white text-xs font-bold">
            {{ (auth.userName || 'U').charAt(0).toUpperCase() }}
          </div>
          <span
            class="text-sm font-medium hidden md:block max-w-[120px] truncate"
            :class="app.darkMode ? 'text-slate-200' : 'text-slate-700'"
          >{{ auth.userName || 'User' }}</span>
          <ChevronDown class="w-4 h-4" :class="app.darkMode ? 'text-slate-400' : 'text-slate-400'" />
        </button>

        <div
          v-if="userMenuOpen"
          class="absolute right-0 top-full mt-1 w-48 rounded-xl shadow-lg py-1 z-50 border"
          :class="app.darkMode
            ? 'bg-slate-800 border-slate-700'
            : 'bg-white border-slate-200'"
          @mouseleave="userMenuOpen = false"
        >
          <div class="px-3 py-2 border-b" :class="app.darkMode ? 'border-slate-700' : 'border-slate-100'">
            <p class="text-sm font-medium" :class="app.darkMode ? 'text-slate-100' : 'text-slate-800'">{{ auth.userName }}</p>
            <p class="text-xs flex items-center gap-1" :class="app.darkMode ? 'text-slate-400' : 'text-slate-500'">
              <Mail class="w-3 h-3" />
              {{ auth.user?.email }}
            </p>
          </div>
          <button
            class="w-full text-left px-3 py-2 text-sm flex items-center gap-2 transition-colors"
            :class="app.darkMode
              ? 'text-slate-300 hover:bg-slate-700'
              : 'text-slate-700 hover:bg-slate-50'"
            @click="router.push('/settings'); userMenuOpen = false"
          >
            <Settings class="w-4 h-4" />
            Settings
          </button>
          <hr :class="app.darkMode ? 'border-slate-700' : 'border-slate-100'" />
          <button
            class="w-full text-left px-3 py-2 text-sm flex items-center gap-2 text-red-500 hover:bg-red-500/10 transition-colors"
            @click="handleLogout"
          >
            <LogOut class="w-4 h-4" />
            Sign out
          </button>
        </div>
      </div>
    </div>
  </header>
</template>
