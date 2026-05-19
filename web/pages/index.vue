<template>
  <div>
    <!-- Filters -->
    <div class="mb-6 flex flex-col gap-3 sm:flex-row">
      <div class="relative flex-1">
        <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          v-model="searchInput"
          type="text"
          placeholder="Search AI news..."
          class="h-9 w-full rounded-md border border-input bg-transparent pl-9 pr-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring"
        />
      </div>
      <select
        v-model="selectedSource"
        class="h-9 rounded-md border border-input bg-transparent px-3 text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring sm:w-48"
      >
        <option :value="0">All sources</option>
        <option v-for="s in sources" :key="s.ID" :value="s.ID">{{ s.Name }}</option>
      </select>
    </div>

    <!-- Loading skeleton -->
    <div v-if="loading && articles.length === 0" class="space-y-3">
      <div
        v-for="i in 5"
        :key="i"
        class="h-28 animate-pulse rounded-lg border border-border bg-card"
      />
    </div>

    <!-- Error -->
    <div v-else-if="error" class="rounded-lg border border-border bg-card p-6 text-center text-muted-foreground">
      Failed to load feed. Is the server running?
    </div>

    <!-- Empty -->
    <div v-else-if="!loading && articles.length === 0" class="py-16 text-center text-muted-foreground">
      No articles found.
    </div>

    <!-- Articles -->
    <div v-else class="space-y-3">
      <ArticleCard v-for="a in articles" :key="a.ID" :article="a" />
    </div>

    <!-- Load more -->
    <div v-if="articles.length > 0 && hasMore" class="mt-6 flex justify-center">
      <button
        :disabled="loading"
        class="rounded-md border border-border px-5 py-2 text-sm text-muted-foreground transition-colors hover:border-border/60 hover:text-foreground disabled:opacity-50"
        @click="loadMore"
      >
        {{ loading ? 'Loading…' : 'Load more' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Search } from 'lucide-vue-next'
import type { Article, Source } from '~/types/api'

const LIMIT = 20

const searchInput = ref('')
const searchQuery = ref('')
const selectedSource = ref(0)
const page = ref(0)
const articles = ref<Article[]>([])
const hasMore = ref(true)
const loading = ref(false)
const error = ref(false)

// Debounce search input
let debounceTimer: ReturnType<typeof setTimeout>
watch(searchInput, (val) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { searchQuery.value = val }, 300)
})

// Reset and reload when filters change
watch([searchQuery, selectedSource], () => {
  page.value = 0
  articles.value = []
  hasMore.value = true
  fetchArticles()
})

const { data: sources } = await useFetch<Source[]>('/api/sources', { default: () => [] })

async function fetchArticles() {
  if (loading.value) return
  loading.value = true
  error.value = false
  try {
    const params: Record<string, string | number> = {
      limit: LIMIT,
      offset: page.value * LIMIT,
    }
    if (selectedSource.value) params.source = selectedSource.value
    if (searchQuery.value) params.q = searchQuery.value

    const data = await $fetch<Article[]>('/api/feed', { query: params })
    articles.value = page.value === 0 ? data : [...articles.value, ...data]
    hasMore.value = data.length === LIMIT
    page.value++
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

function loadMore() {
  fetchArticles()
}

await fetchArticles()
</script>
