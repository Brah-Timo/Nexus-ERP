<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { accountingAPI } from '@/api/client'
const route = useRoute()
const entry = ref<any>(null)
onMounted(async () => {
  const res = await accountingAPI.getJournalEntry(route.params.id as string)
  entry.value = res.data
})
</script>
<template>
  <div>
    <h1 class="text-xl font-bold text-slate-800 mb-4">Journal Entry Detail</h1>
    <div v-if="entry" class="bg-white rounded-xl border border-slate-200 p-6">
      <div class="grid grid-cols-3 gap-4 mb-6">
        <div><p class="text-xs text-slate-500">Number</p><p class="font-semibold">{{ entry.number }}</p></div>
        <div><p class="text-xs text-slate-500">Date</p><p>{{ new Date(entry.date).toLocaleDateString() }}</p></div>
        <div><p class="text-xs text-slate-500">Status</p><p class="font-semibold capitalize">{{ entry.status }}</p></div>
      </div>
      <table class="w-full">
        <thead class="bg-slate-50"><tr>
          <th class="text-left px-3 py-2 text-xs font-semibold text-slate-600">Account</th>
          <th class="text-right px-3 py-2 text-xs font-semibold text-slate-600">Debit</th>
          <th class="text-right px-3 py-2 text-xs font-semibold text-slate-600">Credit</th>
        </tr></thead>
        <tbody class="divide-y divide-slate-100">
          <tr v-for="l in entry.lines" :key="l.id">
            <td class="px-3 py-2 text-sm">{{ l.account_code }} — {{ l.account_name }}</td>
            <td class="px-3 py-2 text-sm text-right font-mono">{{ Number(l.debit).toLocaleString() }}</td>
            <td class="px-3 py-2 text-sm text-right font-mono">{{ Number(l.credit).toLocaleString() }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="text-center py-8 text-slate-400">Loading...</div>
  </div>
</template>
