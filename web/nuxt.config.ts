export default defineNuxtConfig({
  ssr: false,

  app: {
    head: {
      link: [{ rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' }],
    },
  },

  modules: [
    ['@nuxtjs/tailwindcss', { cssPath: '~/assets/css/main.css' }],
    'nuxt-gtag',
  ],

  gtag: {
    id: 'G-Z8W0CE0MCF',
  },

  vite: {
    server: {
      proxy: {
        '/api': 'http://localhost:8080',
      },
    },
  },

  nitro: {
    output: {
      publicDir: 'dist',
    },
  },

  compatibilityDate: '2025-05-18',
})
