# **Word Lookup Documentation**

**Table of Contents**

- [**Word Lookup Documentation**](#word-lookup-documentation)
  - [**1. Purpose**](#1-purpose)
    - [Overview](#overview)
    - [Core Functions](#core-functions)
    - [Why Word Lookup Exists](#why-word-lookup-exists)
  - [**2. System Context**](#2-system-context)
    - [Placement in the Architecture](#placement-in-the-architecture)
      - [Diagram](#diagram)
    - [Integration Goal](#integration-goal)
  - [**3. Responsibilities**](#3-responsibilities)
    - [**3.1 Fetch Data From External Sources**](#31-fetch-data-from-external-sources)
      - [Expected Response Format (Raw Data)](#expected-response-format-raw-data)
    - [**3.2 Normalize and Transform Data**](#32-normalize-and-transform-data)
      - [Example Transformation Workflow](#example-transformation-workflow)
  - [**4. Components \& Services**](#4-components--services)
    - [**4.1 Scraper (Infrastructure Layer)**](#41-scraper-infrastructure-layer)
      - [Example Scraper Output](#example-scraper-output)
    - [**4.2 Anti-Corruption Layer (WiktionaryAdapter)**](#42-anti-corruption-layer-wiktionaryadapter)
      - [WiktionaryAdapter](#wiktionaryadapter)
    - [**4.3 LookupService**](#43-lookupservice)
      - [Example Methods](#example-methods)
  - [**5. Intended Workflow**](#5-intended-workflow)
    - [Scenario 1: Searching for a Word](#scenario-1-searching-for-a-word)
  - [**6. Error Handling**](#6-error-handling)
  - [**7. Future Extensions**](#7-future-extensions)
    - [A. Multi-Source Lookup](#a-multi-source-lookup)
    - [B. Caching](#b-caching)
  - [**8. Summary**](#8-summary)

## **1. Purpose**

### Overview

The **Word Lookup Service** is a supporting service responsible for **fetching, normalizing, and enriching vocabulary data**. It acts as a **translator or adapter** between external sources (e.g., Wiktionary, other APIs) and the **Word Domain**. Its primary responsibility is to supply normalized, domain-ready data that aligns with the structure of the `Word` aggregate.

### Core Functions

1. Query **external sources** to retrieve raw data about a word.
2. Transform and normalize the raw data into a **domain-specific format**.
3. Serve as an **Anti-Corruption Layer** (ACL) to isolate the domain from external API changes.
4. Enrich existing `Word` data with new information (e.g., adding synonyms, definitions, translations).

### Why Word Lookup Exists

- The domain (Word Aggregate) should remain **pure and unaware of external data sources or integration concerns**.
- The Word Domain only works with **complete, validated `Word` aggregates**, which the Word Lookup Service helps generate.

---

## **2. System Context**

### Placement in the Architecture

The Word Lookup Service interfaces between:

- The **Word Domain**, where the business logic resides, and
- The **Infrastructure Layer**, where external source integration occurs (e.g., scraping Wiktionary).

#### Diagram

```
+------------------------+
|      Word Domain       | <-- Validated Word Objects
+------------------------+
             ^
             |
   +-------------------+
   | Word Lookup Layer | <-- Fetch/Translate Raw Data
   +-------------------+
             ^
             |
+----------------------------+
| External Sources (e.g. API)|
+----------------------------+
```

### Integration Goal

- **Domain-Ready Data**: The Word Lookup Service transforms **raw external data** into normalized aggregates (`Word`) usable by the domain.
- **Domain Independence**: The `Word` Aggregate is isolated from the schema and inconsistencies of external systems.

---

## **3. Responsibilities**

The Word Lookup Service fulfills two key responsibilities:

### **3.1 Fetch Data From External Sources**

The Word Lookup fetches raw word-related data through:

- **Scrapers**: For example, extracting definitions and metadata from Wiktionary HTML pages.
- **APIs**: Using REST or GraphQL APIs from dictionary or language services (if integrated in future phases).

#### Expected Response Format (Raw Data)

```json
{
  "word": "marécage",
  "language": "fr",
  "definitions": [
    { "text": "Swampy land...", "type": "noun" },
    { "text": "Unhealthy metaphorical place", "type": "noun" }
  ],
  "etymology": "From old French marais",
  "translations": {
    "en": ["marsh", "swamp"],
    "de": ["Sumpf"]
  },
  "examples": [
    "Tout ce pays-là n'est qu'un grand marécage."
  ]
}
```

---

### **3.2 Normalize and Transform Data**

Once raw data is fetched, the Word Lookup Service transforms it into a structure matching the **Word Aggregate** in the domain:

- **Translate raw schemas into domain concepts** (e.g., turn Wiktionary's "type" field into the domain-relevant `WordType` entity).
- **Fill in missing details** through enrichment or defaults.
- **Validate raw data** to ensure it meets domain constraints before passing it to the Word Domain.

#### Example Transformation Workflow

- Wiktionary "nom" → Domain `WordType: "noun"`.
- Merge multiple synonyms from scraped data into a clean list of unique synonyms.
- Normalize translations into a consistent `map[string][]string`.

---

## **4. Components & Services**

### **4.1 Scraper (Infrastructure Layer)**

The Scraper is the lowest-level component responsible for communicating with an external system. This component:

- Fetches raw HTML data (or JSON via APIs).
- Handles **source-specific logic** (e.g., HTML structure parsing, API authentication).
- Provides the raw `WiktionaryResponse`.

#### Example Scraper Output

```json
{
  "word": "marécage",
  "definitions": [
    { "text": "Swampy land...", "type": "nom" },
    { "text": "Unhealthy metaphorical location", "type": "nom" }
  ],
  "translations": {
    "en": ["marsh", "swamp"],
    "de": ["Sumpf"]
  }
}
```

---

### **4.2 Anti-Corruption Layer (WiktionaryAdapter)**

A **translator** that serves as an Anti-Corruption Layer (ACL), bridging the raw scraper data with domain logic. It serves to:

1. Handle **schema mismatches**: If Wiktionary or other APIs change formats, the adapter hides these changes from the domain.
2. Normalize multi-source inputs into domain-ready objects (`domain.Word`).

#### WiktionaryAdapter

```go
type WiktionaryAdapter struct {
    scraper Scraper
}

func (w *WiktionaryAdapter) FetchAndTransformWord(text string, language string) (domain.Word, error) {
    rawResponse, err := w.scraper.FetchFromWiktionary(text, language)
    if err != nil {
        return domain.Word{}, err
    }

    definitions := parseDefinitions(rawResponse.Definitions)
    translations := normalizeTranslations(rawResponse.Translations)

    return domain.Word{
        ID:           generateUUID(),
        Text:         rawResponse.Word,
        Language:     rawResponse.Language,
        Definitions:  definitions,
        Translations: translations,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }, nil
}
```

---

### **4.3 LookupService**

The publicly accessible service that interacts with application features like search, quiz creation, or word management:

- Abstracts interactions with the `WiktionaryAdapter`.
- Provides the main entry point for retrieving and refreshing word data.

#### Example Methods

```go
type LookupService struct {
    adapter WiktionaryAdapter
}

// Retrieves a Word object by fetching or updating from external sources
func (s *LookupService) GetWord(text string, language string) (domain.Word, error) {
    return s.adapter.FetchAndTransformWord(text, language)
}

// Refreshes and updates a word (e.g., add missing translations or synonyms)
func (s *LookupService) RefreshWord(wordID string) (domain.Word, error) {
    // Fetch new data and merge with existing domain data
}
```

---

## **5. Intended Workflow**

### Scenario 1: Searching for a Word

1. A user searches for **marécage**.
2. **LookupService** receives the input and queries the database:
   - If the word exists, the `Word` object is returned.
   - If not, the **adapter** fetches it from the external source.
3. Scraper retrieves raw data from Wiktionary.
4. The Adapter translates `WiktionaryResponse` into a `domain.Word`.
5. The LookupService stores and returns the new domain `Word`.

---

## **6. Error Handling**

- **Missing Data**: If external sources don’t have enough data for a word:
  - Return a partial `Word` object with missing fields (e.g., no synonyms).
- **API Failures**: Raise retry jobs for temporary external system failures.
- **Invalid Data**: Raise errors if domain invariants (e.g., unique definitions) are violated.

---

## **7. Future Extensions**

### A. Multi-Source Lookup

Support other dictionary APIs to complement Wiktionary, such as:

- Dictionary.com, Oxford API, or proprietary datasets.

### B. Caching

Use Redis or a similar tool to cache raw query results or domain-transformed words.

---

## **8. Summary**

The **Word Lookup Service**:

- Fetches raw external vocabulary data.
- Normalizes and translates external schemas into `Word` domain aggregates.
- Acts as a supporting service for both the Word domain and user-facing features.
- Decouples external API concerns from the core domain logic.
