<template>
    <div class="flex flex-col justify-center items-center min-h-screen bg-white">
      <SearchBar :loading="loading" @search="handleSearch" />
        <div v-if="loading" class="mt-4 text-gray-500">Searching...</div>
        <div v-if="data && !loading" class="p-6 bg-gray-100 rounded-lg shadow-lg mt-6 w-96">
        <p><strong>Definition : </strong>{{ data.word?.definitions?.[0]?.text || 'No definition available.' }}</p>
      </div>
        <div v-if="error && !loading" class="mt-4 text-red-500 font-semibold">
        {{ error }}
      </div>
    </div>
  </template>
  
  <script setup lang="ts">
  import { ref } from 'vue';
  import { searchWord } from '@/services/api/searchServices';
  
  const data = ref<object | null>(null);
  const error = ref<string | null>(null);
  const loading = ref<boolean>(false);
  
  const handleSearch = async (searchTerm: string, language: string) => {
    if (!searchTerm.trim()) {
      error.value = "Please enter a valid search term.";
      return;
    }
  
    try {
      error.value = null; 
      data.value = null;
      loading.value = true;
  
      const response = await searchWord(searchTerm, language);
  
      if (response) {
        data.value = response;
      } else {
        error.value = "Word not found.";
      }
    } catch (err) {
      error.value = (err as Error).message || "An unexpected error occurred.";
    } finally {
      loading.value = false;
    }
  };
  </script>
  