<template>
  <div>
    <!-- Notification banner -->
    <Transition
      enter-active-class="transition-all duration-300"
      enter-from-class="opacity-0 -translate-y-2"
      enter-to-class="opacity-100 translate-y-0"
      leave-active-class="transition-all duration-200"
      leave-from-class="opacity-100 translate-y-0"
      leave-to-class="opacity-0 -translate-y-2"
    >
      <div
        v-if="banner"
        class="mb-6 flex items-start justify-between gap-3 rounded-lg border px-4 py-3 text-sm"
        :class="banner.type === 'success'
          ? 'border-emerald-800 bg-emerald-950 text-emerald-300'
          : 'border-border bg-card text-muted-foreground'"
      >
        <span>{{ banner.message }}</span>
        <button class="shrink-0 opacity-60 hover:opacity-100" @click="banner = null">✕</button>
      </div>
    </Transition>

    <!-- Hero -->
    <div class="mb-8">
      <h1 class="text-2xl font-bold tracking-tight text-foreground">AI News Feed</h1>
      <p class="mt-1 text-sm text-muted-foreground">
        The top AI stories from research papers, Hacker News, Reddit, and RSS — ranked by relevance and freshness.
      </p>
    </div>

    <!-- Category pills -->
    <div class="mb-4 flex gap-2 overflow-x-auto pb-1 scrollbar-none">
      <button
        v-for="cat in categories"
        :key="cat"
        class="shrink-0 rounded-full border px-3 py-1 text-xs font-medium transition-colors"
        :class="
          selectedCategory === cat
            ? 'border-primary bg-primary text-primary-foreground'
            : 'border-border text-muted-foreground hover:border-border/60 hover:text-foreground'
        "
        @click="toggleCategory(cat)"
      >
        {{ cat }}
      </button>
    </div>

    <!-- Search + source filter -->
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
      No articles found{{ selectedCategory !== 'All' ? ` in ${selectedCategory}` : '' }}.
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

    <!-- Subscribe -->
    <div class="mt-16 rounded-lg border border-border bg-card px-6 py-8 text-center">
      <p class="text-xs font-semibold uppercase tracking-widest text-primary">Daily digest</p>
      <h2 class="mt-2 text-lg font-semibold text-foreground">Stay ahead of AI</h2>
      <p class="mt-1 text-sm text-muted-foreground">
        Get the top-scored stories delivered to your inbox every morning.
      </p>

      <form
        v-if="!subscribeSubmitted"
        class="mt-5 flex flex-col gap-2 sm:flex-row sm:justify-center"
        @submit.prevent="submitSubscribe"
      >
        <input
          v-model="subscribeEmail"
          type="email"
          required
          placeholder="you@example.com"
          class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring sm:w-64"
        />
        <button
          type="submit"
          :disabled="subscribeLoading"
          class="h-9 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-50"
        >
          {{ subscribeLoading ? 'Sending…' : 'Subscribe' }}
        </button>
      </form>

      <p v-if="subscribeError" class="mt-2 text-xs text-red-400">
        Something went wrong — please try again.
      </p>

      <p v-if="subscribeSubmitted" class="mt-4 text-sm text-emerald-400">
        Check your inbox — we've sent you a confirmation link.
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Search } from 'lucide-vue-next'
import type { Article, Source } from '~/types/api'

const LIMIT = 20

// ── Notification banner ────────────────────────────────────────────────────

const route = useRoute()
const router = useRouter()

interface Banner { type: 'success' | 'info'; message: string }
const banner = ref<Banner | null>(null)

onMounted(() => {
  const { subscribed, unsubscribed } = route.query
  if (subscribed === '1') {
    banner.value = { type: 'success', message: '✓ You\'re on the list — check your inbox to confirm your subscription.' }
  } else if (subscribed === 'error') {
    banner.value = { type: 'info', message: 'That confirmation link is invalid or has already been used.' }
  } else if (unsubscribed === '1') {
    banner.value = { type: 'info', message: 'You\'ve been unsubscribed. You won\'t receive any more emails from us.' }
  } else if (unsubscribed === 'error') {
    banner.value = { type: 'info', message: 'That unsubscribe link is invalid or has already been used.' }
  }
  if (subscribed || unsubscribed) router.replace({ query: {} })
})

// ── Subscribe form ─────────────────────────────────────────────────────────

const subscribeEmail = ref('')
const subscribeLoading = ref(false)
const subscribeSubmitted = ref(false)
const subscribeError = ref(false)

async function submitSubscribe() {
  subscribeLoading.value = true
  subscribeError.value = false
  try {
    await $fetch('/api/subscribe', {
      method: 'POST',
      body: { email: subscribeEmail.value },
    })
    subscribeSubmitted.value = true
  } catch {
    subscribeError.value = true
  } finally {
    subscribeLoading.value = false
  }
}

const categories = [
  'All',
  'Claude',
  'GPT / OpenAI',
  'Frontier Models',
  'Open Source Models',
  'Agents & Automation',
  'Safety & Policy',
  'Research & Benchmarks',
  'Image & Video',
]

const searchInput = ref('')
const searchQuery = ref('')
const selectedSource = ref(0)
const selectedCategory = ref('All')
const page = ref(0)
const articles = ref<Article[]>([])
const hasMore = ref(true)
const loading = ref(false)
const error = ref(false)

function toggleCategory(cat: string) {
  selectedCategory.value = selectedCategory.value === cat ? 'All' : cat
}

// Debounce search input
let debounceTimer: ReturnType<typeof setTimeout>
watch(searchInput, (val) => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { searchQuery.value = val }, 300)
})

// Reset and reload when filters change
watch([searchQuery, selectedSource, selectedCategory], () => {
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
    if (selectedCategory.value !== 'All') params.category = selectedCategory.value

    const data = (await $fetch<Article[]>('/api/feed', { query: params })) ?? []
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
