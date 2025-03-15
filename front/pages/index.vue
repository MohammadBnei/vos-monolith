<template>
    <div class="flex flex-col justify-center items-center min-h-screen bg-white">
      <SearchBar @search="searchData" />
  
      <div v-if="data" class="p-6 bg-gray-100 rounded-lg shadow-lg mt-6 w-96">
        <p><strong>Definition : </strong>{{ data.word.definitions[0].text }}</p>
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
      const response = await $fetch('/api/search', {
        method: 'POST',
        body: { text: searchTerm, language: 'fr' },
      });
  
      data.value = response;
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
  