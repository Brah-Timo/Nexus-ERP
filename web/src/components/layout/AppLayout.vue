<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import Sidebar from './Sidebar.vue'
import AppBar from './AppBar.vue'
import StatusBar from './StatusBar.vue'

const app = useAppStore()
const auth = useAuthStore()
</script>

<template>
  <div class="flex h-screen overflow-hidden transition-colors duration-200"
       :class="app.darkMode ? 'bg-slate-900' : 'bg-slate-50'">
    <!-- Sidebar -->
    <Sidebar :collapsed="app.sidebarCollapsed" />

    <!-- Main content area -->
    <div class="flex flex-col flex-1 min-w-0 overflow-hidden">
      <!-- Top bar -->
      <AppBar />

      <!-- Page content -->
      <main class="flex-1 overflow-y-auto">
        <div class="p-6">
          <slot />
        </div>
      </main>

      <!-- Status bar -->
      <StatusBar />
    </div>
  </div>
</template>
