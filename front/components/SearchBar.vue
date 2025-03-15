<template>
    <div class="flex flex-col items-center p-4">
      <input
        v-model="searchTerm"
        placeholder="Entrez votre recherche"
        class="p-2 border border-gray-300 rounded-lg w-64 mb-4"
        @input="onInput"
      >
      <button
        class="bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-700"
        @click="search"
      >
        Rechercher
      </button>
    </div>
  </template>
  
  <script setup lang="ts">
  import { ref } from 'vue';
  
  const emit = defineEmits<{
    (event: 'search', searchTerm: string): void;
  }>();
  
  const searchTerm = ref('');
  
  const search = () => {
    if (searchTerm.value) {
      emit('search', searchTerm.value); 
    }
  };
  </script>
  