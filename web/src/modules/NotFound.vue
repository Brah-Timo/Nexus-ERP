<script setup lang="ts">
</script>
<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-50">
    <div class="text-center">
      <p class="text-6xl mb-4">404</p>
      <h1 class="text-2xl font-bold text-slate-800 mb-2">Page Not Found</h1>
      <p class="text-slate-500 mb-6">The page you're looking for doesn't exist.</p>
      <router-link to="/dashboard" class="bg-indigo-600 text-white px-6 py-2.5 rounded-lg font-medium hover:bg-indigo-700">
        Go to Dashboard
      </router-link>
    </div>
  </div>
</template>
