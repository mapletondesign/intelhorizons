<template>
  <header class="sticky top-0 z-50 border-b border-border bg-background/90 backdrop-blur">
    <div class="mx-auto flex h-14 max-w-3xl items-center justify-between px-4">
      <NuxtLink to="/" class="flex items-center gap-2 font-semibold tracking-tight text-foreground">
        <Telescope class="h-4 w-4 text-primary" />
        IntelHorizons
      </NuxtLink>

      <nav class="flex items-center gap-1 text-sm">
        <NuxtLink
          v-for="link in links"
          :key="link.to"
          :to="link.to"
          class="rounded-md px-3 py-1.5 transition-colors"
          :class="
            route.path === link.to
              ? 'bg-secondary text-foreground'
              : 'text-muted-foreground hover:text-foreground'
          "
        >
          {{ link.label }}
        </NuxtLink>
      </nav>
    </div>
  </header>
</template>

<script setup lang="ts">
import { Telescope } from 'lucide-vue-next'

const route = useRoute()

const links = [
  { to: '/', label: 'Feed' },
  { to: '/digest', label: 'Digest' },
  { to: '/sources', label: 'Sources' },
]
</script>
