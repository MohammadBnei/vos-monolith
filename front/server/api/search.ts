export default defineEventHandler(async (event: Event) => {
  const config = useRuntimeConfig(event);
  const body = await readBody(event);

  try {
    const result = await $fetch(`/v1/words/search`, {
      method: "POST",
      baseURL: config.public.apiBase,
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });

    return result;
  } catch (err) {
    if (err.response && err.response.status === 404) {
      return {
        error: "Word not found",
      };
    }
    return {
      error: "An unexpected error occurred",
    };
  }
});
