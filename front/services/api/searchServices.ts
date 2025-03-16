export const searchWord = async (searchTerm: string, language: string) => {
  if (!searchTerm) return null;

  try {
    const response = await $fetch("/api/search", {
      method: "POST",
      body: { text: searchTerm, language: language },
    });

    return response;
  } catch (error) {
    console.error("Search issue", error);
    throw new Error("An error occurs for research");
  }
};
