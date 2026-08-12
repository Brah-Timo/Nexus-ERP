<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '@/api/client'

const router = useRouter()
const email = ref('')
const loading = ref(false)
const sent = ref(false)
const error = ref('')

async function submit() {
  if (!email.value) return
  loading.value = true
  try {
    await authAPI.forgotPassword(email.value)
    sent.value = true
  } catch {
    error.value = 'Failed to send reset email'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-900">
    <div class="bg-white rounded-2xl shadow-xl p-10 w-[420px]">
      <router-link to="/login" class="text-sm text-indigo-600 flex items-center gap-1 mb-6">
        ← Back to login
      </router-link>
      <h1 class="text-xl font-bold text-slate-800 mb-2">Reset Password</h1>
      <p class="text-sm text-slate-500 mb-6">Enter your email and we'll send you a reset link.</p>

      <div v-if="sent" class="bg-green-50 border border-green-200 text-green-700 rounded-xl px-4 py-3 text-sm">
        ✅ Reset link sent. Please check your email.
      </div>
      <template v-else>
        <div v-if="error" class="mb-4 bg-red-50 border border-red-200 text-red-700 rounded-xl px-4 py-3 text-sm">{{ error }}</div>
        <input
          v-model="email"
          type="email"
          placeholder="your@email.com"
          class="w-full border border-slate-300 rounded-xl px-4 py-2.5 mb-4 focus:ring-2 focus:ring-indigo-500 focus:outline-none"
        />
        <button
          class="w-full bg-indigo-600 text-white rounded-xl py-2.5 font-semibold disabled:opacity-60"
          :disabled="loading || !email"
          @click="submit"
        >
          {{ loading ? 'Sending...' : 'Send Reset Link' }}
        </button>
      </template>
    </div>
  </div>
</template>
