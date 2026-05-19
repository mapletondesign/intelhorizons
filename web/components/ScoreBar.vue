<template>
  <div class="flex items-center gap-0.5" :title="`Score: ${score.toFixed(1)}`">
    <span
      v-for="i in 5"
      :key="i"
      class="h-2.5 w-1 rounded-sm"
      :class="i <= filled ? filledClass : 'bg-muted'"
    />
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{ score: number }>()

const filled = computed(() => Math.max(1, Math.round((props.score / 10) * 5)))

const filledClass = computed(() => {
  if (props.score >= 6.5) return 'bg-emerald-500'
  if (props.score >= 3.5) return 'bg-amber-500'
  return 'bg-sky-600'
})
</script>
