<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Eye, EyeOff, LogIn, Loader2, AlertCircle, Lock, User } from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const form = ref({ username: '', password: '', rememberMe: false })
const showPassword = ref(false)

async function handleLogin() {
  if (!form.value.username || !form.value.password) return
  auth.error = null
  try {
    await auth.login(form.value.username, form.value.password)
    const redirect = route.query.redirect as string
    router.push(redirect || '/dashboard')
  } catch {
    // error handled in store
  }
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') handleLogin()
}
</script>

<template>
  <div class="min-h-screen relative flex items-center justify-center overflow-hidden"
    style="background: linear-gradient(135deg, #0f172a 0%, #1e1b4b 40%, #312e81 70%, #1e1b4b 100%);">

    <!-- Animated background blobs -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none">
      <div class="absolute -top-40 -left-40 w-96 h-96 rounded-full opacity-20"
        style="background: radial-gradient(circle, #6366f1, transparent 70%); animation: pulse 4s ease-in-out infinite;"></div>
      <div class="absolute -bottom-40 -right-40 w-96 h-96 rounded-full opacity-20"
        style="background: radial-gradient(circle, #818cf8, transparent 70%); animation: pulse 4s ease-in-out infinite 2s;"></div>
      <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] rounded-full opacity-5"
        style="background: radial-gradient(circle, #a5b4fc, transparent 70%);"></div>
      <!-- Grid pattern -->
      <svg class="absolute inset-0 w-full h-full opacity-5" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
            <path d="M 40 0 L 0 0 0 40" fill="none" stroke="white" stroke-width="0.5"/>
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid)" />
      </svg>
    </div>

    <!-- Main card -->
    <div class="relative w-full max-w-md mx-4 z-10">
      <div class="rounded-3xl overflow-hidden"
        style="background: rgba(255,255,255,0.03); backdrop-filter: blur(24px); border: 1px solid rgba(255,255,255,0.12); box-shadow: 0 32px 64px rgba(0,0,0,0.5), 0 0 0 1px rgba(255,255,255,0.05);">

        <!-- Top gradient banner -->
        <div class="relative px-8 py-8 overflow-hidden"
          style="background: linear-gradient(135deg, rgba(99,102,241,0.9) 0%, rgba(139,92,246,0.9) 100%);">
          <!-- Banner shimmer -->
          <div class="absolute inset-0 opacity-30"
            style="background: linear-gradient(90deg, transparent, rgba(255,255,255,0.15), transparent); background-size: 200% 100%; animation: shimmer 3s infinite;"></div>
          <div class="relative flex items-center gap-4">
            <!-- Logo -->
            <div class="w-14 h-14 rounded-2xl flex items-center justify-center flex-shrink-0"
              style="background: rgba(255,255,255,0.15); border: 1px solid rgba(255,255,255,0.25); box-shadow: 0 4px 12px rgba(0,0,0,0.2);">
              <span class="text-white font-black text-2xl tracking-tight">N</span>
            </div>
            <div>
              <h1 class="text-white font-bold text-xl tracking-tight">Nexus ERP</h1>
              <p class="text-indigo-200 text-xs font-medium mt-0.5">Enterprise Resource Planning — Algeria</p>
            </div>
          </div>
          <!-- Decorative circles -->
          <div class="absolute -right-6 -top-6 w-24 h-24 rounded-full opacity-20"
            style="border: 2px solid white;"></div>
          <div class="absolute -right-2 -top-2 w-12 h-12 rounded-full opacity-10"
            style="background: white;"></div>
        </div>

        <!-- Form body -->
        <div class="px-8 py-8">
          <div class="mb-6">
            <h2 class="font-bold text-white text-xl tracking-tight">Welcome back</h2>
            <p class="text-slate-400 text-sm mt-1">Sign in to your workspace</p>
          </div>

          <!-- Error alert -->
          <transition name="slide-down">
            <div v-if="auth.error"
              class="mb-5 rounded-xl px-4 py-3 flex items-center gap-3 text-sm font-medium"
              style="background: rgba(239,68,68,0.15); border: 1px solid rgba(239,68,68,0.3); color: #fca5a5;">
              <AlertCircle class="w-4 h-4 flex-shrink-0 text-red-400" />
              <span>{{ auth.error }}</span>
            </div>
          </transition>

          <div class="space-y-4">
            <!-- Username field -->
            <div>
              <label class="block text-xs font-semibold mb-1.5" style="color: #94a3b8;">Username</label>
              <div class="relative group">
                <User class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 transition-colors"
                  :style="form.username ? 'color:#818cf8' : 'color:#475569'" />
                <input
                  v-model="form.username"
                  type="text"
                  autocomplete="username"
                  placeholder="Enter your username"
                  class="w-full rounded-xl text-sm transition-all outline-none pl-10 pr-4 py-3"
                  style="background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.1); color: #f1f5f9; placeholder-color: #475569;"
                  :style="form.username
                    ? 'background: rgba(99,102,241,0.1); border-color: rgba(99,102,241,0.5); box-shadow: 0 0 0 3px rgba(99,102,241,0.15);'
                    : ''"
                  @focus="$event.target.style.cssText += '; background: rgba(99,102,241,0.08); border-color: rgba(99,102,241,0.4); box-shadow: 0 0 0 3px rgba(99,102,241,0.12);'"
                  @blur="$event.target.style.cssText = form.username
                    ? 'background: rgba(99,102,241,0.1); border: 1px solid rgba(99,102,241,0.5); color: #f1f5f9;'
                    : 'background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.1); color: #f1f5f9;'"
                  @keydown="handleKeydown"
                />
              </div>
            </div>

            <!-- Password field -->
            <div>
              <label class="block text-xs font-semibold mb-1.5" style="color: #94a3b8;">Password</label>
              <div class="relative">
                <Lock class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 transition-colors"
                  :style="form.password ? 'color:#818cf8' : 'color:#475569'" />
                <input
                  v-model="form.password"
                  :type="showPassword ? 'text' : 'password'"
                  autocomplete="current-password"
                  placeholder="Enter your password"
                  class="w-full rounded-xl text-sm transition-all outline-none pl-10 pr-11 py-3"
                  style="background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.1); color: #f1f5f9;"
                  @focus="$event.target.style.cssText += '; background: rgba(99,102,241,0.08); border-color: rgba(99,102,241,0.4); box-shadow: 0 0 0 3px rgba(99,102,241,0.12);'"
                  @blur="$event.target.style.cssText = form.password
                    ? 'background: rgba(99,102,241,0.1); border: 1px solid rgba(99,102,241,0.5); color: #f1f5f9;'
                    : 'background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.1); color: #f1f5f9;'"
                  @keydown="handleKeydown"
                />
                <button
                  type="button"
                  class="absolute right-3.5 top-1/2 -translate-y-1/2 transition-colors hover:opacity-100 opacity-60"
                  style="color: #94a3b8;"
                  @click="showPassword = !showPassword"
                >
                  <EyeOff v-if="showPassword" class="w-4 h-4" />
                  <Eye v-else class="w-4 h-4" />
                </button>
              </div>
            </div>

            <!-- Remember / Forgot row -->
            <div class="flex items-center justify-between">
              <label class="flex items-center gap-2 cursor-pointer select-none group">
                <div class="relative w-4 h-4 flex-shrink-0">
                  <input type="checkbox" v-model="form.rememberMe" class="sr-only" />
                  <div class="w-4 h-4 rounded transition-all"
                    :style="form.rememberMe
                      ? 'background: #6366f1; border: 1px solid #6366f1;'
                      : 'background: transparent; border: 1px solid rgba(255,255,255,0.2);'">
                    <svg v-if="form.rememberMe" class="w-full h-full p-0.5" fill="none" stroke="white" stroke-width="3" viewBox="0 0 16 16">
                      <path d="M3 8l3.5 3.5 6.5-6.5" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                  </div>
                </div>
                <span class="text-xs" style="color: #94a3b8;">Remember me</span>
              </label>
              <router-link to="/forgot-password"
                class="text-xs font-medium transition-colors hover:opacity-100 opacity-80"
                style="color: #818cf8;">
                Forgot password?
              </router-link>
            </div>

            <!-- Sign in button -->
            <button
              class="w-full rounded-xl py-3 text-sm font-semibold flex items-center justify-center gap-2.5 transition-all relative overflow-hidden"
              :disabled="auth.loading || !form.username || !form.password"
              :style="(auth.loading || !form.username || !form.password)
                ? 'background: rgba(99,102,241,0.4); color: rgba(255,255,255,0.5); cursor: not-allowed;'
                : 'background: linear-gradient(135deg, #6366f1, #8b5cf6); color: white; box-shadow: 0 4px 20px rgba(99,102,241,0.4); cursor: pointer;'"
              @click="handleLogin"
            >
              <!-- Shimmer on hover -->
              <div class="absolute inset-0 opacity-0 hover:opacity-100 transition-opacity"
                style="background: linear-gradient(90deg, transparent, rgba(255,255,255,0.1), transparent); background-size: 200% 100%; animation: shimmer 2s infinite;"></div>
              <Loader2 v-if="auth.loading" class="w-4 h-4 animate-spin relative z-10" />
              <LogIn v-else class="w-4 h-4 relative z-10" />
              <span class="relative z-10">{{ auth.loading ? 'Signing in...' : 'Sign In' }}</span>
            </button>
          </div>

          <!-- Credentials hint card -->
          <div class="mt-6 rounded-xl p-4"
            style="background: rgba(99,102,241,0.08); border: 1px solid rgba(99,102,241,0.2);">
            <div class="flex items-center gap-2 mb-2">
              <div class="w-1.5 h-1.5 rounded-full" style="background: #6366f1;"></div>
              <p class="text-xs font-semibold" style="color: #818cf8;">Default Credentials</p>
            </div>
            <div class="space-y-1">
              <div class="flex items-center gap-2">
                <User class="w-3 h-3 flex-shrink-0" style="color:#6366f1" />
                <p class="text-xs" style="color: #94a3b8;">
                  Username: <code class="font-mono font-bold" style="color:#a5b4fc;">admin</code>
                </p>
              </div>
              <div class="flex items-center gap-2">
                <Lock class="w-3 h-3 flex-shrink-0" style="color:#6366f1" />
                <p class="text-xs" style="color: #94a3b8;">
                  Password: <code class="font-mono font-bold" style="color:#a5b4fc;">Admin@123456</code>
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="text-center mt-5 flex items-center justify-center gap-3">
        <div class="h-px flex-1" style="background: rgba(255,255,255,0.08);"></div>
        <p class="text-xs flex-shrink-0" style="color: #475569;">
          Nexus ERP v1.1.0 &copy; {{ new Date().getFullYear() }}
        </p>
        <div class="h-px flex-1" style="background: rgba(255,255,255,0.08);"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
input::placeholder {
  color: #475569;
}
@keyframes shimmer {
  0%   { background-position: -200% 0; }
  100% { background-position:  200% 0; }
}
.slide-down-enter-active { transition: all 0.25s ease; }
.slide-down-leave-active { transition: all 0.2s ease; }
.slide-down-enter-from   { opacity: 0; transform: translateY(-8px); }
.slide-down-leave-to     { opacity: 0; transform: translateY(-4px); }
</style>
