<script setup lang="ts">
interface Props {
  title: string
  value: string | number
  icon: string
  color?: 'green' | 'blue' | 'amber' | 'red' | 'purple' | 'indigo' | 'teal' | 'orange'
  trend?: number
  subtitle?: string
}

const props = withDefaults(defineProps<Props>(), {
  color: 'blue'
})

const colorStyles: Record<string, { card: string; icon: string }> = {
  green:  { card: 'border-green-200 bg-green-50',   icon: 'bg-green-100 text-green-700' },
  blue:   { card: 'border-blue-200 bg-blue-50',     icon: 'bg-blue-100 text-blue-700' },
  amber:  { card: 'border-amber-200 bg-amber-50',   icon: 'bg-amber-100 text-amber-700' },
  red:    { card: 'border-red-200 bg-red-50',       icon: 'bg-red-100 text-red-700' },
  purple: { card: 'border-purple-200 bg-purple-50', icon: 'bg-purple-100 text-purple-700' },
  indigo: { card: 'border-indigo-200 bg-indigo-50', icon: 'bg-indigo-100 text-indigo-700' },
  teal:   { card: 'border-teal-200 bg-teal-50',     icon: 'bg-teal-100 text-teal-700' },
  orange: { card: 'border-orange-200 bg-orange-50', icon: 'bg-orange-100 text-orange-700' },
}
</script>

<template>
  <div
    class="rounded-xl border p-4 transition-all hover:shadow-md cursor-pointer"
    :class="colorStyles[color].card"
  >
    <div class="flex items-start justify-between mb-3">
      <div class="w-10 h-10 rounded-lg flex items-center justify-center text-lg" :class="colorStyles[color].icon">
        {{ icon }}
      </div>
      <div v-if="trend !== undefined" class="flex items-center gap-1 text-xs font-medium" :class="trend >= 0 ? 'text-green-600' : 'text-red-600'">
        <span>{{ trend >= 0 ? '▲' : '▼' }}</span>
        <span>{{ Math.abs(trend) }}%</span>
      </div>
    </div>
    <p class="text-2xl font-bold text-slate-900 tabular-nums">{{ value }}</p>
    <p class="text-sm font-medium text-slate-600 mt-0.5">{{ title }}</p>
    <p v-if="subtitle" class="text-xs text-slate-400 mt-0.5">{{ subtitle }}</p>
  </div>
</template>
