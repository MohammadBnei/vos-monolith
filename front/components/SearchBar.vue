<template>
    <form class="flex items-center gap-2" @submit.prevent="search">
      <input
        v-model="searchTerm"
        placeholder="Research a word"
        class="input px-4 py-2 border rounded-lg"
      >
        <button type="submit" class="btn px-4 py-2 text-grey rounded-lg">
        Search
      </button>
    </form>
  </template>
  
  <script setup lang="ts">
  import { ref } from 'vue';
  
  const emit = defineEmits<{
    (event: 'search', searchTerm: string, language: string): void;
  }>();
  
  const searchTerm = ref('');
  const language = ref('fr');
  
  const search = () => {
    if (searchTerm.value.trim()) {
      emit('search', searchTerm.value.trim(), language.value);
    }
  };
  </script>
  