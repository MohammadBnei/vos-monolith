// src/types/search.ts
export interface Definition {
  text: string;
}

export interface Word {
  text: string;
  definitions?: Definition[];
  etymology: string;
  synonyms?: string[];
}

export interface SearchResult {
  word?: Word;
}
