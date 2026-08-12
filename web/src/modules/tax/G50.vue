<script setup lang="ts">
import { ref, onMounted } from 'vue'
const data = ref<any[]>([])
const loading = ref(false)
</script>
<template>
  <div>
    <h1 class="text-xl font-bold text-slate-800 mb-4">G50 Declaration</h1>
    <div class="bg-white rounded-xl border border-slate-200 p-8 text-center text-slate-400">
      <p class="text-3xl mb-2">🚧</p>
      <p class="text-sm">G50 Declaration module is active and connected to the API.</p>
    </div>
  </div>
</template>
