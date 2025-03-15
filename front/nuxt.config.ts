// https://nuxt.com/docs/api/configuration/nuxt-config
import tailwindcss from "@tailwindcss/vite";
export default defineNuxtConfig({
  compatibilityDate: "2024-11-01",
  devtools: { enabled: true },
  css: ["~/assets/styles/tailwind.css"],
  vite: {
    plugins: [tailwindcss()],
  },
  modules: [
    "@nuxt/content",
    "@nuxt/eslint",
    "@nuxt/icon",
    "@nuxt/fonts",
    "@nuxt/image",
    "@nuxt/ui",
    "@nuxt/scripts",
  ],
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL,
    },
  },
});
