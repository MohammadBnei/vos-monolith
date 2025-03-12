# **Word Domain Documentation**

**Table of Content**

- [**Word Domain Documentation**](#word-domain-documentation)
  - [**1. Purpose**](#1-purpose)
    - [Overview](#overview)
  - [**2. Domain Concept**](#2-domain-concept)
    - [Vocabulary and Learning](#vocabulary-and-learning)
  - [**3. Domain Model**](#3-domain-model)
    - [**3.1 Word Aggregate**](#31-word-aggregate)
      - [What is the Word Aggregate Root?](#what-is-the-word-aggregate-root)
    - [**3.2 Entity: Word**](#32-entity-word)
      - [Definition](#definition)
      - [Properties](#properties)
      - [Responsibilities](#responsibilities)
    - [**3.3 Entity: Definition**](#33-entity-definition)
      - [Definition](#definition-1)
      - [Properties](#properties-1)
      - [Responsibilities](#responsibilities-1)
    - [**3.4 Value Object: Translation**](#34-value-object-translation)
      - [Definition](#definition-2)
      - [Properties](#properties-2)
      - [Example](#example)
    - [**3.5 Enrichment Handling**](#35-enrichment-handling)
    - [**3.6 Invariants (Consistency Rules)**](#36-invariants-consistency-rules)
      - [Word Invariants](#word-invariants)
      - [Definition Invariants](#definition-invariants)
  - [**4. Domain Services**](#4-domain-services)
    - [Key Operations](#key-operations)
      - [1. Add New Definition](#1-add-new-definition)
      - [2. Merge Translations](#2-merge-translations)
  - [**5. Relationships with Other Domains**](#5-relationships-with-other-domains)
  - [**6. Persistence**](#6-persistence)
  - [**7. Future Considerations**](#7-future-considerations)

---

## **1. Purpose**

### Overview

The **Word Domain** is the primary **core domain** of the application, encapsulating the concept of a "word" with all its associated information and relationships. It governs the **state** of vocabulary data, ensures data **consistency**, and enforces **business rules** around words while supporting **partial word data** for progressive enrichment.

The **Word domain** contains entities (e.g., `Word`, `Definition`) and value objects (e.g., `Translation`). It interacts with supporting services (e.g., **Word Lookup**) to retrieve external data and enrich the domain model over time.

---

## **2. Domain Concept**

### Vocabulary and Learning

The Word domain represents the "real-world vocabulary" comprising:

- Linguistic properties like its definitions, etymology, synonyms, and antonyms.
- Usage details provided for better contextual understanding.
- Multilingual aspects by including translations into other languages.
- Metadata to track the completeness of vocabulary data fields.

This domain ensures words are **progressively enriched** and structured in a way that supports learning, navigating saved words, contextual usage, and future features like quizzes and learning algorithms.

---

## **3. Domain Model**

### **3.1 Word Aggregate**

#### What is the Word Aggregate Root?

The `Word` is the **aggregate root** in the Word domain. It encapsulates:

- **Entities**: Relationships between `Word` and its associated `Definition`.
- **Value Objects**: Concepts like `Translation` that add logical meaning.

The `Word` aggregate ensures consistency within its own boundaries:

1. No duplicate definitions.
2. Translations grouped uniquely by language, ensuring no conflicts.

---

### **3.2 Entity: Word**

#### Definition

A `Word` is the central entity of the Word domain, representing a vocabulary word in its canonical form along with associated linguistic data, including partial data.

#### Properties

```go
type Word struct {
 ID            string              // Unique identifier for the word (UUID)
 Text          string              // Canonical lemma (e.g., "run", "marécage")
 Language      string              // ISO 639-1 language code
 Definitions   []Definition        // Array of enriched definitions
 Etymology     *string             // Historical origin (optional)
 Translations  map[string][]string // Language-based translations
 Synonyms      []string            // Synonyms for the word
 Antonyms      []string            // Antonyms for the word
 SearchTerms   []string            // Searchable variants of the word
 Lemma         string              // Base form of the word
 UsageNotes    []string            // Usage or idiomatic notes
 Enrichment    EnrichmentStatus    // Tracks completeness of word data
 CreatedAt     time.Time           // Created timestamp
 UpdatedAt     time.Time           // Updated timestamp
}
```

#### Responsibilities

The `Word` is responsible for:

1. Representing the core properties of a vocabulary word.
2. Validating business rules, e.g.:
   - At least one `Definition` must exist.
   - Translating external data structures into domain formats.
3. Tracking enrichment status via a separate field.

---

### **3.3 Entity: Definition**

#### Definition

A `Definition` provides the semantic explanation of a `Word`.

#### Properties

```go
type Definition struct {
 ID               string              // Unique identifier for the definition
 Text             string              // Explanation of the word
 WordType         string              // Grammatical type of the word
 Examples         []string            // Example sentences (optional)
 Gender           string              // Gender information
 Pronunciation    *string             // Pronunciation in IPA
 LanguageSpecifics map[string]string  // Plural forms, gender specifics, etc.
 CreatedAt        time.Time           // Created timestamp
 UpdatedAt        time.Time           // Updated timestamp
}
```

#### Responsibilities

1. Represent linguistic meaning and context for a `Word`.
2. Support hierarchical structure (e.g., relationships to word types).

---

### **3.4 Value Object: Translation**

#### Definition

Stores groupings of cross-language translations as a map.

#### Properties

```go
Translations map[string][]string // Map of language codes -> array of translated terms
```

#### Example

```json
{
  "translations": {
    "en": ["swamp", "marsh"],
    "de": ["Sumpf"],
    "fr": []
  }
}
```

---

### **3.5 Enrichment Handling**

Enrichment Status tracks missing fields across the `Word` object. Example:

```json
"enrichment": {
  "definitions": true,
  "translations": false,
  "etymology": true,
  "examples": false,
  "synonyms": true,
  "antonyms": false
}
```

---

### **3.6 Invariants (Consistency Rules)**

#### Word Invariants

1. A `Word` must always have:
   - Valid `Text` and `Language`.
   - At least one `Definition`.

#### Definition Invariants

1. Each `Example` for a `Definition` must be unique.

---

## **4. Domain Services**

### Key Operations

#### 1. Add New Definition

```go
func (w *Word) AddDefinition(def Definition) error {
  // Prevent duplicate definitions
  for _, existing := range w.Definitions {
    if existing.Text == def.Text {
      return errors.New("duplicate definition")
    }
  }
  w.Definitions = append(w.Definitions, def)
  w.UpdateEnrichmentStatus()
}
```

#### 2. Merge Translations

```go
func (w *Word) MergeTranslations(newTranslations map[string][]string) {
  for lang, words := range newTranslations {
    w.Translations[lang] = mergeUnique(w.Translations[lang], words)
  }
}
```

---

## **5. Relationships with Other Domains**

1. **Word Lookup** supplies external data to this domain for creation/enrichment.
2. **Saved Word**: Enables persisting user-saved words.

---

## **6. Persistence**

| Table        | Properties                                                               |
|--------------|--------------------------------------------------------------------------|
| `words`      | `id`, `text`, `language`, `created_at`, `updated_at`                     |
| `definitions`| `id`, `word_id`, `text`, `word_type`, `examples`, `created_at`, `updated_at`|

---

## **7. Future Considerations**

1. **User Contributions**: Community-provided examples and annotations.
2. **Enrichment Pipelines**: Automate periodic crawling for missing data.
