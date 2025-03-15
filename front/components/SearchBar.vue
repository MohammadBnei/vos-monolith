<template>
   <div class="flex items-center gap-2">
        <input
        v-model="searchTerm"
        placeholder="Entrez votre recherche"
        class="input"
        @input="onInput"
      >
      <button
        class= "btn"
        @click="search"
      > 
      Search
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
  