<template>
  <div>
    <div class="mb-6 border-b border-border pb-6">
      <h1 class="text-2xl font-semibold tracking-tight text-foreground">Sources</h1>
      <p class="mt-1 text-sm text-muted-foreground">
        {{ activeSources.length }} active sources across {{ typeCount }} types
      </p>
    </div>

    <div v-if="pending" class="space-y-3">
      <div v-for="i in 4" :key="i" class="h-10 animate-pulse rounded-lg bg-card" />
    </div>

    <div v-else-if="error" class="py-16 text-center text-muted-foreground">
      Failed to load sources.
    </div>

    <div v-else class="space-y-8">
      <section v-for="(group, type) in grouped" :key="type">
        <div class="mb-3 flex items-center gap-3">
          <SourceBadge :source-name="String(type)" :source-type="(type as any)" />
          <span class="text-sm text-muted-foreground">{{ group.length }} source{{ group.length !== 1 ? 's' : '' }}</span>
        </div>
        <div class="rounded-lg border border-border divide-y divide-border overflow-hidden">
          <div
            v-for="s in group"
            :key="s.ID"
            class="flex items-center justify-between bg-card px-4 py-3"
          >
            <span class="text-sm text-foreground">{{ s.Name }}</span>
            <span class="text-xs text-muted-foreground">every {{ interval(s.FetchIntervalMins) }}</span>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Source } from '~/types/api'

const { data: sources, pending, error } = await useFetch<Source[]>('/api/sources', {
  default: () => [],
})

const activeSources = computed(() => sources.value.filter(s => s.Active))

const grouped = computed(() => {
  const order = ['rss', 'arxiv', 'hn', 'reddit', 'twitter']
  const map: Record<string, Source[]> = {}
  for (const s of activeSources.value) {
    if (!map[s.Type]) map[s.Type] = []
    map[s.Type].push(s)
  }
  const sorted: Record<string, Source[]> = {}
  for (const key of order) {
    if (map[key]) sorted[key] = map[key]
  }
  return sorted
})

const typeCount = computed(() => Object.keys(grouped.value).length)

function interval(mins: number) {
  if (mins < 60) return `${mins}m`
  const hrs = mins / 60
  return hrs === 1 ? '1h' : `${hrs}h`
}
</script>
