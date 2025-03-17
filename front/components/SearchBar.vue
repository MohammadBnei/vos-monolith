<template>
  <div class="flex justify-center items-center min-h-screen">
    <div 
      :class="{
        'absolute left-8 top-1/2 transform -translate-y-1/2 w-96 p-4': !data,
        'absolute left-8 top-1/2 transform -translate-y-1/2 w-96 p-4': data
      }"
    >
      <form @submit.prevent="handleSearch">
        <input
          v-model="searchTerm"
          type="text"
          placeholder="Enter a word..."
          class="w-full p-2 border border-gray-300 rounded bg-gray-100 focus:outline-none"
        >
        <button
          type="submit"
          :disabled="loading"
          class="mt-2 w-full bg-gray-300 text-white p-2 rounded hover:bg-gray-600 disabled:opacity-50"
        >Search</button>
      </form>
    </div>

    <div v-if="data && !loading" class="p-6 bg-gray-100 rounded-lg shadow-lg mt-6 w-full max-w-4xl flex-1 ml-96">
      <div class="flex-1 mr-8 min-w-[300px]">
        <h2 class="text-xl font-semibold">Definition</h2>
        <p>{{ data.word?.definitions?.[0]?.text || "No definition available." }}</p>
      </div>

      <div class="flex-1">
        <div v-if="data.word?.definitions?.length > 1" class="mt-4">
          <h3 class="text-lg font-medium border-t pt-2">Other definitions</h3>
          <ul class="list-disc ml-4 space-y-1">
            <li v-for="(definition, index) in data.word.definitions.slice(1)" :key="index">
              {{ definition.text }}
            </li>
          </ul>
        </div>

        <div v-if="data.word?.etymology?.length > 1" class="mt-4">
          <h3 class="text-lg font-medium border-t pt-2">Etymology</h3>
          <p class="italic">{{ data.word?.etymology }}</p>
        </div>

        <div v-if="data.word?.synonyms?.length > 1" class="mt-4">
          <h3 class="text-lg font-medium border-t pt-2">Synonyms</h3>
          <ul class="list-disc ml-4 space-y-1">
            <li v-for="(synonym, index) in data.word?.synonyms" :key="index">
              {{ synonym }}
            </li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useSearch } from "@/composables/useSearch";
const { searchTerm, data, loading, handleSearch } = useSearch();
</script>
