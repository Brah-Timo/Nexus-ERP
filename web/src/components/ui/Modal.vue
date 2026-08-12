<script setup lang="ts">
interface Props {
  title?: string
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  closable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  size: 'md',
  closable: true,
})

const emit = defineEmits<{
  close: []
}>()

const sizeClasses: Record<string, string> = {
  sm: 'max-w-md',
  md: 'max-w-lg',
  lg: 'max-w-2xl',
  xl: 'max-w-4xl',
  full: 'max-w-7xl',
}
</script>

<template>
  <teleport to="body">
    <div
      class="fixed inset-0 z-50 flex items-center justify-center p-4 overflow-y-auto"
      @click.self="closable && emit('close')"
    >
      <!-- Backdrop -->
      <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="closable && emit('close')"></div>

      <!-- Panel -->
      <div
        class="relative bg-white rounded-2xl shadow-2xl w-full my-4"
        :class="sizeClasses[size]"
        style="animation: fadeInScale 0.2s ease forwards"
      >
        <!-- Header -->
        <div v-if="title" class="flex items-center justify-between px-6 py-4 border-b border-slate-200">
          <h2 class="text-lg font-bold text-slate-800">{{ title }}</h2>
          <button
            v-if="closable"
            class="w-8 h-8 flex items-center justify-center rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors"
            @click="emit('close')"
          >
            ✕
          </button>
        </div>
        <button
          v-else-if="closable"
          class="absolute top-4 right-4 w-8 h-8 flex items-center justify-center rounded-lg text-slate-400 hover:text-slate-600 hover:bg-slate-100 transition-colors z-10"
          @click="emit('close')"
        >
          ✕
        </button>

        <!-- Content -->
        <div class="overflow-y-auto max-h-[80vh]">
          <slot />
        </div>

        <!-- Footer slot -->
        <slot name="footer" />
      </div>
    </div>
  </teleport>
</template>

<style scoped>
@keyframes fadeInScale {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-8px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}
</style>
