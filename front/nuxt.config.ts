// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({

  modules: [
    '@nuxt/content',
    '@nuxt/eslint',
    '@nuxt/icon',
    '@nuxt/fonts',
    '@nuxt/image',
    '@nuxt/ui',
    '@nuxt/scripts',
    '@nuxtjs/storybook',
  ],
  devtools: { enabled: true },
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL,
    },
  },
  compatibilityDate: '2024-11-01',
  vite: {
    plugins: [tailwindcss()],
  },

  eslint: {
    config: {
      standalone: true,
    },
  },
})
