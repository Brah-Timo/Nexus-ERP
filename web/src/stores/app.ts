import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export interface Toast {
  id: string
  message: string
  type: 'success' | 'error' | 'warning' | 'info'
  duration?: number
}

export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const commandPaletteOpen = ref(false)
  const toasts = ref<Toast[]>([])
  const globalLoading = ref(false)
  const currentFiscalYear = ref<string | null>(null)
  const currentBranch = ref<string | null>(null)

  // ── Theme ──────────────────────────────────────────────────────────────────
  // Initialise from localStorage or system preference
  const savedTheme = localStorage.getItem('nexus-theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  const darkMode = ref<boolean>(
    savedTheme !== null ? savedTheme === 'dark' : prefersDark
  )

  function applyTheme(dark: boolean) {
    if (dark) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
    localStorage.setItem('nexus-theme', dark ? 'dark' : 'light')
  }

  // Apply on init
  applyTheme(darkMode.value)

  // Watch for changes
  watch(darkMode, (val) => applyTheme(val))

  function toggleTheme() {
    darkMode.value = !darkMode.value
  }

  // ── Sidebar ────────────────────────────────────────────────────────────────
  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function toggleCommandPalette() {
    commandPaletteOpen.value = !commandPaletteOpen.value
  }

  // ── Toasts ─────────────────────────────────────────────────────────────────
  function addToast(message: string, type: Toast['type'] = 'info', duration = 4000) {
    const id = Date.now().toString()
    toasts.value.push({ id, message, type, duration })
    if (duration > 0) {
      setTimeout(() => removeToast(id), duration)
    }
    return id
  }

  function removeToast(id: string) {
    const idx = toasts.value.findIndex(t => t.id === id)
    if (idx !== -1) toasts.value.splice(idx, 1)
  }

  function success(message: string) {
    return addToast(message, 'success')
  }

  function error(message: string) {
    return addToast(message, 'error', 6000)
  }

  function warning(message: string) {
    return addToast(message, 'warning')
  }

  return {
    sidebarCollapsed, commandPaletteOpen, toasts,
    globalLoading, currentFiscalYear, currentBranch, darkMode,
    toggleSidebar, toggleCommandPalette, toggleTheme,
    addToast, removeToast, success, error, warning
  }
})
