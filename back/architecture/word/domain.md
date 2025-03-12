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
      - [Relationships](#relationships)
    - [**3.4 Value Object: Translation**](#34-value-object-translation)
      - [Definition](#definition-2)
      - [Properties](#properties-2)
      - [Example](#example)
      - [Responsibilities](#responsibilities-2)
    - [**3.5 Invariants (Consistency Rules)**](#35-invariants-consistency-rules)
      - [1. Word Invariants](#1-word-invariants)
      - [2. Definition Invariants](#2-definition-invariants)
  - [**4. Domain Services**](#4-domain-services)
    - [**4.1 Word Factory**](#41-word-factory)
    - [**4.2 Key Operations**](#42-key-operations)
      - [1. Add a New Definition](#1-add-a-new-definition)
      - [2. Enrich Synonyms](#2-enrich-synonyms)
  - [**5. Relationships with Other Domains**](#5-relationships-with-other-domains)
  - [**6. Persistence**](#6-persistence)
    - [Tables](#tables)
  - [**7. Future Considerations**](#7-future-considerations)

## **1. Purpose**

### Overview

The **Word Domain** is the primary **core domain** of the application, encapsulating the concept of a "word" with all its associated information and relationships. It governs the **state** of vocabulary data, ensures data **consistency**, and enforces **business rules** around words.

The Word domain contains entities (e.g., `Word`, `Definition`) and value objects (e.g., `Translation`). It interacts with supporting services (e.g., **Word Lookup**) to retrieve external data for enriching the domain model.

---

## **2. Domain Concept**

### Vocabulary and Learning

The Word domain represents the "real-world vocabulary" comprising:

- Linguistic properties like its definitions, etymology, synonyms, and antonyms.
- Usage details, provided for better contextual understanding.
- Multilingual aspects by including translations into other languages.

This domain ensures words are organized and structured in a way that supports learning, navigating saved words, contextual usage, and future features like quizzes and learning algorithms.

---

## **3. Domain Model**

### **3.1 Word Aggregate**

#### What is the Word Aggregate Root?

The `Word` is the **aggregate root** of the Word domain. It encapsulates:

- **Entities**: Relationships between `Word` and its associated `Definition` entities.
- **Value Objects**: Concepts like `Translation` are immutable value objects that add logical meaning to `Word`.

The `Word` aggregate ensures consistency within its own boundaries:

- No duplicate definitions.
- No conflicting translations.

---

### **3.2 Entity: Word**

#### Definition

A `Word` is the central entity of the Word domain, representing a vocabulary word in its canonical form along with associated linguistic data.

#### Properties

```go
type Word struct {
 ID            string              // Unique identifier of the word (UUID)
 Text          string              // Canonical lemma (e.g., "run", "marécage")
 Language      string              // ISO 639-1 language code (e.g., "en", "fr")
 Definitions   []Definition        // Array of enriched definitions
 Etymology     string              // Historical origin of the word (optional)
 Translations  map[string][]string // Translations grouped by language (e.g., "fr" -> ["marais"])
 Synonyms      []string            // List of synonyms for the word
 Antonyms      []string            // List of antonyms for the word
 SearchTerms   []string            // Variants of the word for easier searching
 Lemma         string              // Base form of the word if `Text` is conjugated
 UsageNotes    []string            // Notes for contextual usage
 CreatedAt     time.Time           // Timestamp for word creation
 UpdatedAt     time.Time           // Timestamp for last word update
}
```

#### Responsibilities

1. Ensure consistency between related entities like `Definition` and `Translation`.
2. Provide a structure for multilingual data (e.g., synonyms, antonyms, and translations).
3. Represent real-world vocabulary accurately, encapsulating linguistic and contextual aspects.

---

### **3.3 Entity: Definition**

#### Definition

A `Definition` represents a **semantic explanation** of a `Word`. A single `Word` can have one or more definitions depending on its linguistic or contextual usage.

#### Properties

```go
type Definition struct {
 ID               string              // Unique identifier for the definition (UUID)
 Text             string              // The core definition text
 WordType         string              // Grammatical type (e.g., noun, verb)
 Examples         []string            // Examples that clarify usage
 Gender           string              // Gender of the word (e.g., "masculine", "feminine" in French)
 Pronunciation    string              // Phonetics or IPA notation (e.g., "/rəˈnæt/ for run")
 LanguageSpecifics map[string]string  // Additional language-based details (e.g., plural forms)
 Notes            []string            // Supplementary notes
 CreatedAt        time.Time           // Timestamp for when the definition was created
 UpdatedAt        time.Time           // Timestamp for when the definition was last updated
}
```

#### Responsibilities

1. Provide linguistic meaning for a `Word`.
2. Store grammatical, gender, and phonetic information.
3. Store contextual examples to enrich understanding of the meaning.

#### Relationships

- A `Word` can have **multiple Definitions**, often corresponding to different contexts or grammatical types.

---

### **3.4 Value Object: Translation**

#### Definition

A `Translation` represents the equivalence of a `Word` in another language. It is an **immutable value object**, meaning once created, it cannot be changed.

#### Properties

```go
Translations map[string][]string // Map of [language code] -> [array of translated terms]
```

#### Example

```json
{
    "translations": {
        "en": ["marsh", "swamp", "bog"],
        "de": ["Sumpf"],
        "es": ["pantano", "ciénaga"]
    }
}
```

- **Key**: Language code (e.g., "en", "de").
- **Value**: Array of translations in the target language.

#### Responsibilities

1. Encapsulate cross-context vocabulary relationships (e.g., synonyms in other languages).
2. Ensure consistent and clean representation of multilingual data.

---

### **3.5 Invariants (Consistency Rules)**

#### 1. Word Invariants

- A `Word` must have at least one `Definition`.
- `Text` must be unique when combined with `Language` (e.g., "run:en" and "run:fr" are distinct).
- No duplicate values in synonyms, antonyms, or translations.

#### 2. Definition Invariants

- `Examples` should be unique for each definition to avoid redundancy.
- Every `WordType` must map to known grammatical categories ("noun", "verb", etc.).

---

## **4. Domain Services**

The Word domain primarily **interacts** with external services (like **Word Lookup**) and internal services (like **Saved Words**) via domain services.

---

### **4.1 Word Factory**

A `WordFactory` service may be introduced to encapsulate logic for creating `Word` aggregates. It may:

- Build `Word` objects from external sources (like Wiktionary).
- Populate a `Word` with enriched data (e.g., synonyms, antonyms).

---

### **4.2 Key Operations**

#### 1. Add a New Definition

```go
func (w *Word) AddDefinition(def Definition) error {
    if w.HasDuplicateDefinition(def) {
        return errors.New("duplicate definition")
    }
    w.Definitions = append(w.Definitions, def)
    w.UpdatedAt = time.Now()
    return nil
}
```

#### 2. Enrich Synonyms

```go
func (w *Word) AddSynonyms(newSynonyms []string) {
    for _, synonym := range newSynonyms {
        if !w.HasSynonym(synonym) {
            w.Synonyms = append(w.Synonyms, synonym)
        }
    }
    w.UpdatedAt = time.Now()
}
```

---

## **5. Relationships with Other Domains**

1. **Word Lookup**:
   - **Purpose**: The Word Lookup service supplies external vocabulary data to the Word domain.
   - **Interaction**:
     - The **Anti-Corruption Layer (ACL)** converts external data into `Word` domain objects.
     - For example, raw data from Wiktionary is normalized and passed into the Word Factory or directly into the domain's repository.

2. **Word Saving and Management**:
   - The `SavedWord` entity (from the **Word Saving domain**) references the `Word` entity, enabling users to save and organize specific vocabulary items.

3. **Quizzing and Gamification** (Future):
   - The Word domain integrates with features like spaced-repetition algorithms and quizzes to provide a personalized learning experience.

---

## **6. Persistence**

### Tables

| Table        | Description                                           |
|--------------|-------------------------------------------------------|
| `words`      | Stores `Word` metadata (ID, text, language, etc.).    |
| `definitions`| Stores definitions linked to the `Word` entity.       |

---

## **7. Future Considerations**

1. **Data Enrichment**:
   - Enable user-provided examples or definitions to supplement the data.

2. **Contextual Word Relationships**:
   - Relate words through tree-like relationships (e.g., "run" → "ran", "running").
