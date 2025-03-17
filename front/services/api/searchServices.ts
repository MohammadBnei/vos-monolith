export const searchWord = async (searchTerm: string, language: string) => {
  if (!searchTerm) return null

  return $fetch('/api/search', {
    method: 'POST',
    body: { text: searchTerm, language },
  })
}
