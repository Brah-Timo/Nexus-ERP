<script setup lang="ts">
interface Column {
  key: string
  label: string
  align?: 'left' | 'right' | 'center'
  format?: (val: unknown) => string
  class?: string
}

interface Props {
  columns: Column[]
  rows: Record<string, unknown>[]
  loading?: boolean
  emptyText?: string
  onRowClick?: (row: Record<string, unknown>) => void
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  emptyText: 'No data found',
})

function getCellValue(row: Record<string, unknown>, col: Column): string {
  const val = row[col.key]
  if (col.format) return col.format(val)
  if (val === null || val === undefined) return '—'
  if (typeof val === 'boolean') return val ? 'Yes' : 'No'
  return String(val)
}

function alignClass(align?: string): string {
  if (align === 'right') return 'text-right'
  if (align === 'center') return 'text-center'
  return 'text-left'
}
</script>

<template>
  <div class="bg-white rounded-xl border border-slate-200 overflow-hidden">
    <div v-if="loading" class="flex items-center justify-center py-16">
      <div class="text-center">
        <p class="text-3xl animate-spin mb-2">⏳</p>
        <p class="text-sm text-slate-400">Loading...</p>
      </div>
    </div>

    <div v-else class="overflow-x-auto">
      <table class="w-full min-w-max">
        <thead class="bg-slate-50 border-b border-slate-200">
          <tr>
            <th
              v-for="col in columns"
              :key="col.key"
              class="px-4 py-3 text-xs font-semibold text-slate-600 uppercase tracking-wider"
              :class="alignClass(col.align)"
            >
              {{ col.label }}
            </th>
          </tr>
        </thead>

        <tbody class="divide-y divide-slate-100">
          <tr v-if="rows.length === 0">
            <td :colspan="columns.length" class="px-4 py-12 text-center text-sm text-slate-400">
              {{ emptyText }}
            </td>
          </tr>

          <tr
            v-for="(row, idx) in rows"
            :key="idx"
            class="transition-colors"
            :class="onRowClick ? 'hover:bg-slate-50 cursor-pointer' : ''"
            @click="onRowClick?.(row)"
          >
            <td
              v-for="col in columns"
              :key="col.key"
              class="px-4 py-3 text-sm"
              :class="[alignClass(col.align), col.class || 'text-slate-700']"
            >
              <slot :name="`cell-${col.key}`" :value="row[col.key]" :row="row">
                {{ getCellValue(row, col) }}
              </slot>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
