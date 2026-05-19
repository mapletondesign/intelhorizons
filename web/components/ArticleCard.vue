<template>
  <div class="group rounded-lg border border-border bg-card p-4 transition-colors hover:border-border/60">
    <div class="mb-2 flex items-center justify-between gap-3">
      <div class="flex min-w-0 items-center gap-2">
        <SourceBadge :source-name="article.SourceName" :source-type="sourceType" />
        <span v-if="article.Category" class="truncate text-xs text-muted-foreground">
          {{ article.Category }}
        </span>
      </div>
      <div class="flex shrink-0 items-center gap-2.5">
        <ScoreBar :score="article.Score" />
        <span class="text-xs text-muted-foreground">{{ timeAgo }}</span>
      </div>
    </div>

    <a
      :href="article.URL"
      target="_blank"
      rel="noopener noreferrer"
      class="mb-1.5 block font-medium leading-snug text-foreground transition-colors line-clamp-2 hover:text-primary"
    >
      {{ article.Title }}
    </a>

    <p v-if="article.Summary" class="text-sm leading-relaxed text-muted-foreground line-clamp-2">
      {{ article.Summary }}
    </p>
  </div>
</template>

<script setup lang="ts">
import type { Article } from '~/types/api'

const props = defineProps<{ article: Article }>()

const timeAgo = computed(() => {
  const ms = Date.now() - new Date(props.article.PublishedAt).getTime()
  const mins = Math.floor(ms / 60_000)
  if (mins < 60) return `${mins}m`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h`
  return `${Math.floor(hrs / 24)}d`
})

const sourceType = computed(() => {
  const name = props.article.SourceName.toLowerCase()
  if (name.includes('arxiv')) return 'arxiv'
  if (name.includes('hacker news')) return 'hn'
  if (name.includes('reddit')) return 'reddit'
  return 'rss'
})
</script>
