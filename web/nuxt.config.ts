export default defineNuxtConfig({
  ssr: false,

  modules: [
    ['@nuxtjs/tailwindcss', { cssPath: '~/assets/css/main.css' }],
  ],

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
