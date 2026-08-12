import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI } from '@/api/client'

interface User {
  id: string
  username: string
  email: string
  full_name: string
  role_id: string
  company_id: string
  branch_id: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('nexus_token'))
  const refreshToken = ref<string | null>(localStorage.getItem('nexus_refresh_token'))
  const user = ref<User | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value)
  const userName = computed(() => user.value?.full_name ?? '')
  const userRole = computed(() => user.value?.role_id ?? '')
  const companyId = computed(() => user.value?.company_id ?? '')

  async function login(username: string, password: string, companyId?: string) {
    loading.value = true
    error.value = null
    try {
      const response = await authAPI.login(username, password, companyId)
      const data = response.data
      token.value = data.token
      refreshToken.value = data.refresh_token
      user.value = data.user
      localStorage.setItem('nexus_token', data.token)
      localStorage.setItem('nexus_refresh_token', data.refresh_token)
      return data
    } catch (err: unknown) {
      const e = err as { response?: { data?: { error?: string } } }
      error.value = e.response?.data?.error ?? 'Login failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  async function logout() {
    try {
      await authAPI.logout()
    } catch {
      // Ignore
    } finally {
      token.value = null
      refreshToken.value = null
      user.value = null
      localStorage.removeItem('nexus_token')
      localStorage.removeItem('nexus_refresh_token')
    }
  }

  async function refresh() {
    if (!refreshToken.value) throw new Error('No refresh token')
    const response = await authAPI.refresh(refreshToken.value)
    token.value = response.data.token
    localStorage.setItem('nexus_token', response.data.token)
  }

  async function fetchCurrentUser() {
    // In a real app, fetch /api/auth/me
    // For now, parse from stored user if needed
    const stored = localStorage.getItem('nexus_user')
    if (stored) {
      try {
        user.value = JSON.parse(stored)
      } catch {
        // Invalid stored data
      }
    }
  }

  function hasPermission(permission: string): boolean {
    // Simplified — admin has all permissions
    if (userRole.value === 'admin' || userRole.value === 'superadmin') return true
    // TODO: check against role permissions
    return true
  }

  return {
    token, refreshToken, user, loading, error,
    isAuthenticated, userName, userRole, companyId,
    login, logout, refresh, fetchCurrentUser, hasPermission
  }
}, {
  persist: false
})
