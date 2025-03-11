-- Create words table
CREATE TABLE IF NOT EXISTS words (
    id UUID PRIMARY KEY,
    text TEXT NOT NULL,
    language TEXT NOT NULL,
    definitions JSONB,
    etymology TEXT,
    translations JSONB,
    synonyms TEXT[],
    antonyms TEXT[],
    search_terms TEXT[],
    lemma TEXT,
    usage_notes TEXT[],
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(text, language)
);

-- Create index on text and language
CREATE INDEX IF NOT EXISTS idx_words_text_language ON words(text, language);

-- Create index on search_terms (using GIN for array searching)
CREATE INDEX IF NOT EXISTS idx_words_search_terms ON words USING GIN(search_terms);

-- Create index on updated_at for sorting recent words
CREATE INDEX IF NOT EXISTS idx_words_updated_at ON words(updated_at DESC);

-- Add pg_trgm extension for text similarity search
CREATE EXTENSION IF NOT EXISTS pg_trgm;
