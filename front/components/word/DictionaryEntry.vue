<template>
  <div class="container mx-auto p-4 space-y-6">
    <!-- Header Section -->
    <header class="flex flex-col gap-2">
      <h1 class="text-3xl font-bold">
        {{ word.text }}
      </h1>
      <div class="flex items-center gap-2">
        <span class="px-2 py-1 bg-primary text-primary-foreground rounded-full text-sm">
          {{ word.language }}
        </span>
        <span class="italic text-muted-foreground">
          Prononciation: {{ definition.pronunciation }}
        </span>
      </div>
    </header>

    <!-- Definitions Section -->
    <section class="space-y-4">
      <h2 class="text-xl font-semibold">
        Définitions
      </h2>

      <div class="card shadow-lg">
        <div class="card-body space-y-4">
          <p class="text-base leading-relaxed">
            {{ definition.text }}
          </p>

          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div>
              <p class="text-sm font-medium mb-1">
                Genre
              </p>
              <p class="text-sm text-muted-foreground">
                {{ definition.gender }}
              </p>
            </div>

            <div>
              <p class="text-sm font-medium mb-1">
                Type
              </p>
              <p class="text-sm text-muted-foreground">
                {{ definition.word_type }}
              </p>
            </div>
          </div>

          <!-- Examples -->
          <div class="mt-4">
            <h3 class="text-lg font-semibold mb-2">
              Exemples
            </h3>
            <ul class="space-y-2">
              <li
                v-for="(example, index) in definition.examples"
                :key="index"
                class="pl-4 border-l-2 border-primary"
              >
                {{ example }}
              </li>
            </ul>
          </div>
        </div>
      </div>
    </section>

    <!-- Etymology Section -->
    <section class="space-y-2">
      <h3 class="text-lg font-semibold">
        Étymologie
      </h3>
      <p class="text-muted-foreground">
        {{ word.etymology }}
      </p>
    </section>

    <!-- Translations Section -->
    <section class="space-y-2">
      <h3 class="text-lg font-semibold">
        Traductions
      </h3>
      <div class="grid grid-cols-2 md:grid-cols-3 gap-4">
        <div
          v-for="(translation, lang) in word.translations"
          :key="lang"
          class="bg-secondary p-3 rounded-lg"
        >
          <p class="font-medium">
            {{ lang.toUpperCase() }}
          </p>
          <p class="text-sm text-muted-foreground">
            {{ translation }}
          </p>
        </div>
      </div>
    </section>

    <!-- Search Terms Section -->
    <section class="space-y-2">
      <h3 class="text-lg font-semibold">
        Termes de recherche
      </h3>
      <div class="flex flex-wrap gap-2">
        <span
          v-for="term in word.search_terms"
          :key="term"
          class="px-3 py-1 bg-secondary rounded-full text-sm"
        >
          {{ term }}
        </span>
      </div>
    </section>

    <!-- Timestamps Section -->
    <footer class="space-y-2 pt-4 border-t">
      <p class="text-sm text-muted-foreground">
        Créé le: {{ formatDate(word.created_at) }}
      </p>
      <p class="text-sm text-muted-foreground">
        Mis à jour le: {{ formatDate(word.updated_at) }}
      </p>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

// Props interface
interface WordData {
  id: string
  text: string
  language: string
  definitions: Array<{
    text: string
    word_type: string
    examples: string[]
    gender: string
    pronunciation: string
  }>
  etymology: string
  translations: Record<string, string>
  search_terms: string[]
  lemma: string
  created_at: string
  updated_at: string
}

const props = defineProps<{
  word: WordData
}>()

// Computed properties
const word = computed(() => props.word)
const definition = computed(() => word.value.definitions[0])

// Utility function
const formatDate = (dateString: string): string => {
  return new Date(dateString).toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  })
}
</script>
