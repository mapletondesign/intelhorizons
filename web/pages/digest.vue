<template>
  <div>
    <!-- Header -->
    <div class="mb-6 border-b border-border pb-6">
      <h1 class="text-2xl font-semibold tracking-tight text-foreground">Today's Digest</h1>
      <p class="mt-1 text-sm text-muted-foreground">
        {{ today }} · top {{ articles.length }} AI stories from the last 24 hours
      </p>
    </div>

    <!-- Loading skeleton -->
    <div v-if="pending" class="space-y-px">
      <div
        v-for="i in 10"
        :key="i"
        class="h-16 animate-pulse rounded-lg bg-card"
      />
    </div>

    <!-- Error -->
    <div v-else-if="error" class="py-16 text-center text-muted-foreground">
      Failed to load digest.
    </div>

    <!-- Empty -->
    <div v-else-if="!articles.length" class="py-16 text-center text-muted-foreground">
      No stories yet today — check back soon.
    </div>

    <!-- Ranked list -->
    <ol v-else class="space-y-px">
      <li
        v-for="(article, i) in articles"
        :key="article.ID"
        class="group flex gap-4 rounded-lg p-4 transition-colors hover:bg-card"
      >
        <span class="mt-0.5 w-6 shrink-0 text-right text-sm font-medium tabular-nums text-muted-foreground">
          {{ i + 1 }}
        </span>
        <div class="min-w-0 flex-1">
          <div class="mb-1 flex items-center gap-2">
            <SourceBadge :source-name="article.SourceName" :source-type="sourceType(article)" />
            <span class="text-xs text-muted-foreground">{{ timeAgo(article.PublishedAt) }}</span>
            <ScoreBar :score="article.Score" />
          </div>
          <a
            :href="article.URL"
            target="_blank"
            rel="noopener noreferrer"
            class="block font-medium leading-snug text-foreground transition-colors line-clamp-2 hover:text-primary"
          >
            {{ article.Title }}
          </a>
          <p v-if="article.Summary" class="mt-1 text-sm text-muted-foreground line-clamp-1">
            {{ article.Summary }}
          </p>
        </div>
      </li>
    </ol>
  </div>
</template>

<script setup lang="ts">
import type { Article } from '~/types/api'

const { data: articles, pending, error } = await useFetch<Article[]>('/api/digest', {
  default: () => [],
})

const today = new Intl.DateTimeFormat('en-US', {
  weekday: 'long', year: 'numeric', month: 'long', day: 'numeric',
}).format(new Date())

function timeAgo(iso: string) {
  const ms = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(ms / 60_000)
  if (mins < 60) return `${mins}m`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h`
  return `${Math.floor(hrs / 24)}d`
}

function sourceType(article: Article): 'arxiv' | 'hn' | 'reddit' | 'rss' | 'twitter' {
  const name = article.SourceName.toLowerCase()
  if (name.includes('arxiv')) return 'arxiv'
  if (name.includes('hacker news')) return 'hn'
  if (name.includes('reddit')) return 'reddit'
  return 'rss'
}
</script>
