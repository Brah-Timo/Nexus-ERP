<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import AppLayout from '@/components/layout/AppLayout.vue'
import CommandPalette from '@/components/CommandPalette.vue'

const route = useRoute()
const auth = useAuthStore()
const app = useAppStore()

const isAuthPage = computed(() => route.meta?.public === true)

// Global keyboard shortcuts
function handleKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    app.toggleCommandPalette()
  }
  if (e.key === 'Escape') {
    app.commandPaletteOpen = false
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
  if (auth.token) {
    auth.fetchCurrentUser()
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div id="nexus-app" class="font-sans antialiased">
    <!-- Auth pages (login, forgot password) -->
    <template v-if="isAuthPage">
      <router-view />
    </template>

    <!-- Main application layout -->
    <template v-else>
      <AppLayout>
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" :key="route.path" />
          </transition>
        </router-view>
      </AppLayout>
    </template>

    <!-- Global Command Palette -->
    <CommandPalette v-if="!isAuthPage" />

    <!-- Global toast notifications -->
    <teleport to="body">
      <div class="fixed top-4 right-4 z-[9999] flex flex-col gap-2 pointer-events-none">
        <transition-group name="toast">
          <div
            v-for="toast in app.toasts"
            :key="toast.id"
            class="pointer-events-auto flex items-center gap-3 px-4 py-3 rounded-lg shadow-lg text-sm font-medium max-w-sm"
            :class="{
              'bg-green-50 text-green-800 border border-green-200': toast.type === 'success',
              'bg-red-50 text-red-800 border border-red-200': toast.type === 'error',
              'bg-amber-50 text-amber-800 border border-amber-200': toast.type === 'warning',
              'bg-blue-50 text-blue-800 border border-blue-200': toast.type === 'info',
            }"
          >
            <span>{{ toast.message }}</span>
            <button @click="app.removeToast(toast.id)" class="ml-auto text-current opacity-60 hover:opacity-100">✕</button>
          </div>
        </transition-group>
      </div>
    </teleport>
  </div>
</template>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.toast-enter-active {
  transition: all 0.3s ease;
}
.toast-leave-active {
  transition: all 0.3s ease;
}
.toast-enter-from {
  transform: translateX(100%);
  opacity: 0;
}
.toast-leave-to {
  transform: translateX(100%);
  opacity: 0;
}
</style>
