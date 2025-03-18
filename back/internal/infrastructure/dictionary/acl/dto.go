package acl

// WiktionaryResponse represents the raw data structure returned by Wiktionary
type WiktionaryResponse struct {
	Word          string                      `json:"word"`
	Language      string                      `json:"language"`
	Definitions   []WiktionaryDefinition      `json:"definitions,omitempty"`
	Etymology     string                      `json:"etymology,omitempty"`
	Translations  map[string][]string         `json:"translations,omitempty"`
	Synonyms      []string                    `json:"synonyms,omitempty"`
	Antonyms      []string                    `json:"antonyms,omitempty"`
	Pronunciation string                      `json:"pronunciation,omitempty"`
	SearchTerms   []string                    `json:"search_terms,omitempty"`
	Lemma         string                      `json:"lemma,omitempty"`
}

// WiktionaryDefinition represents a raw definition from Wiktionary
type WiktionaryDefinition struct {
	Text              string            `json:"text"`
	Type              string            `json:"type,omitempty"` // noun, verb, etc.
	Examples          []string          `json:"examples,omitempty"`
	Gender            string            `json:"gender,omitempty"`
	Pronunciation     string            `json:"pronunciation,omitempty"`
	LanguageSpecifics map[string]string `json:"language_specifics,omitempty"`
	Notes             []string          `json:"notes,omitempty"`
}

// WiktionaryRelatedResponse represents related words data from Wiktionary
type WiktionaryRelatedResponse struct {
	SourceWord string   `json:"source_word"`
	Language   string   `json:"language"`
	Synonyms   []string `json:"synonyms,omitempty"`
	Antonyms   []string `json:"antonyms,omitempty"`
}

// EnrichmentStatus tracks which fields need to be enriched
type EnrichmentStatus struct {
	NeedsDefinitions  bool `json:"needs_definitions"`
	NeedsEtymology    bool `json:"needs_etymology"`
	NeedsTranslations bool `json:"needs_translations"`
	NeedsSynonyms     bool `json:"needs_synonyms"`
	NeedsAntonyms     bool `json:"needs_antonyms"`
	NeedsPronunciation bool `json:"needs_pronunciation"`
}
