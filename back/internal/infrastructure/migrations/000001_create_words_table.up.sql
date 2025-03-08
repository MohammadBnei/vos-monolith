CREATE TABLE IF NOT EXISTS words (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    text VARCHAR(255) NOT NULL,
    language VARCHAR(10) NOT NULL,
    definitions JSONB NOT NULL DEFAULT '[]'::jsonb,
    examples TEXT[] NOT NULL DEFAULT '{}',
    pronunciation JSONB NOT NULL DEFAULT '{}'::jsonb,
    etymology TEXT,
    translations JSONB NOT NULL DEFAULT '{}'::jsonb,
    word_type TEXT,
    forms JSONB NOT NULL DEFAULT '[]'::jsonb,
    search_terms TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[],
    lemma TEXT,
    synonyms TEXT[] NOT NULL DEFAULT '{}',
    antonyms TEXT[] NOT NULL DEFAULT '{}',
    usage_notes TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(text, language)
);

CREATE INDEX idx_words_text_language ON words(text, language);
CREATE INDEX idx_words_language ON words(language);
CREATE INDEX idx_words_updated_at ON words(updated_at DESC);
CREATE INDEX idx_words_search_terms ON words USING GIN (search_terms);

COMMENT ON COLUMN words.definitions IS 'Word definitions with type information';
COMMENT ON COLUMN words.forms IS 'Different forms of the word with attributes';
COMMENT ON COLUMN words.search_terms IS 'All searchable forms of the word';
COMMENT ON COLUMN words.lemma IS 'Base form of the word';
COMMENT ON COLUMN words.pronunciation IS 'Different pronunciation formats (IPA, audio URL, etc.)';
COMMENT ON COLUMN words.synonyms IS 'List of synonyms for the word';
COMMENT ON COLUMN words.antonyms IS 'List of antonyms for the word';
COMMENT ON COLUMN words.usage_notes IS 'General usage information';
