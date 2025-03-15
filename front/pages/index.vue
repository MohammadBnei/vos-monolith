<template>
    <div class="container mx-auto p-8">
      <h1 class="text-4xl font-bold mb-6 text-center">Recherche API</h1>
      <SearchBar @search="searchData" />
  
      <div v-if="data" class="mt-6">
        <h2 class="text-2xl font-semibold">Résultat :</h2>
        <p class="mt-2 text-lg"><strong>Mot : </strong>{{ data.word }}</p>
        <p class="mt-2 text-lg"><strong>Définition : </strong>{{ data.word.definitions }}</p>
      </div>
  
      <div v-if="error" class="error mt-4 text-red-500">
        Une erreur est survenue lors de la recherche.
      </div>
    </div>
  </template>
  
  <script setup lang="ts">
  import { ref } from 'vue';
  
  const data = ref<any>(null);  
  const error = ref<string | null>(null);
  
  const searchData = async (searchTerm: string) => {
    if (!searchTerm) return;
  
    try {
      // Appel API
      const response = await $fetch('/api/search', {
        method: 'POST',
        body: { text: searchTerm, language: 'fr' },
      });
  
      data.value = response;  // Stocke le résultat dans `data`
    } catch (err) {
      error.value = 'Une erreur est survenue lors de la recherche';
    }
  };
  </script>
  
  <style scoped>
  .error {
    color: red;
    font-weight: bold;
  }
  </style>
  