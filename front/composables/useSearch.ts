import { ref } from "vue";
import { searchWord } from "@/services/api/searchServices";
import type { SearchResult } from "@/types/search";

export function useSearch() {
  const searchTerm = ref<string>("");
  const language = ref("fr");
  const data = ref<SearchResult | null>(null);
  const error = ref<string | null>(null);
  const loading = ref<boolean>(false);

  const handleSearch = async () => {
    if (!searchTerm.value.trim()) return;

    error.value = null;
    data.value = null;
    loading.value = true;

    try {
      const response = await searchWord(searchTerm.value, language.value);
      data.value = response || null;
      if (!response) error.value = "Word not found.";
    } catch (err) {
      error.value =
        err instanceof Error ? err.message : "An unexpected error occurred.";
    } finally {
      loading.value = false;
    }
  };

  return { searchTerm, language, data, loading, handleSearch };
}
