<template>
  <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ring-1 ring-inset" :class="classes">
    {{ label }}
  </span>
</template>

<script setup lang="ts">
const props = defineProps<{
  sourceName: string
  sourceType: 'arxiv' | 'hn' | 'reddit' | 'rss' | 'twitter'
}>()

const config: Record<string, { label: string; classes: string }> = {
  arxiv:   { label: 'arXiv',    classes: 'bg-violet-500/10 text-violet-400 ring-violet-500/20' },
  hn:      { label: 'HN',       classes: 'bg-orange-500/10 text-orange-400 ring-orange-500/20' },
  reddit:  { label: 'Reddit',   classes: 'bg-red-500/10    text-red-400    ring-red-500/20'    },
  twitter: { label: 'X',        classes: 'bg-sky-500/10    text-sky-400    ring-sky-500/20'    },
  rss:     { label: 'RSS',      classes: 'bg-indigo-500/10 text-indigo-400 ring-indigo-500/20' },
}

const label = computed(() => config[props.sourceType]?.label ?? props.sourceName)
const classes = computed(() => config[props.sourceType]?.classes ?? config.rss.classes)
</script>
