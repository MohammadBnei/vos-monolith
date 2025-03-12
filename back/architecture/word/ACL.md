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
    - [**3.2 Normalize, Transform, and Enrich Data**](#32-normalize-transform-and-enrich-data)
      - [Example Transformation Workflow](#example-transformation-workflow)
    - [**3.3 Handling Partial Data**](#33-handling-partial-data)
      - [Progressive Enrichment Workflow](#progressive-enrichment-workflow)
      - [Prioritized Field Fetching](#prioritized-field-fetching)
  - [**4. Components \& Services**](#4-components--services)
    - [**4.1 Scraper (Infrastructure Layer)**](#41-scraper-infrastructure-layer)
      - [Example Scraper Output](#example-scraper-output)
    - [**4.2 Anti-Corruption Layer (WiktionaryAdapter)**](#42-anti-corruption-layer-wiktionaryadapter)
      - [Updated Adapter Behavior for Partial Handling](#updated-adapter-behavior-for-partial-handling)
    - [**4.3 LookupService**](#43-lookupservice)
      - [Updated Example Methods](#updated-example-methods)
  - [**5. Intended Workflow**](#5-intended-workflow)
    - [Scenario 1: Searching and Creating a Word](#scenario-1-searching-and-creating-a-word)
    - [Scenario 2: Enriching Partial Fields](#scenario-2-enriching-partial-fields)
  - [**6. Error Handling**](#6-error-handling)
    - [Graceful Partial Returns](#graceful-partial-returns)
  - [**7. Future Extensions**](#7-future-extensions)
    - [A. Multi-Source Lookup](#a-multi-source-lookup)
    - [B. Caching](#b-caching)
  - [**8. Summary**](#8-summary)

---

## **1. Purpose**

### Overview

The **Word Lookup Service** is a supporting service responsible for **fetching, normalizing, and enriching vocabulary data**, even when only partial data is available. It acts as a **translator or adapter** between external systems (e.g., Wiktionary, other APIs) and the **Word Domain**.

### Core Functions

1. Query **external sources** to retrieve raw data about a word.
2. Transform and normalize the raw data into a **domain-specific format** while handling incomplete data.
3. Support **progressive enrichment** of `Word` objects by focusing on missing fields identified through `EnrichmentStatus`.
4. Continue to serve as an **Anti-Corruption Layer** (ACL) to hide external schema differences.

### Why Word Lookup Exists

- **Domain Purity**: The `Word` aggregate in the domain should remain decoupled from external source schemas.
- **Data Normalization**: The Word Lookup Service processes and enriches the data into consistent, domain-ready objects.
- **Progressive Enrichment**: Supports iterative updates to vocabulary data when only partial data is returned or available during subsequent lookups.

---

## **2. System Context**

### Placement in the Architecture

The Word Lookup Service interfaces between:

- **The Word Domain**, which manages the business logic.
- **The Infrastructure Layer**, which interacts directly with external systems.

#### Diagram

```
+------------------------+
|      Word Domain       | <-- Validated Domain `Word` Objects
+------------------------+
             ^
             |
   +-------------------+
   | Word Lookup Layer | <-- Fetch/Normalize Raw Data
   +-------------------+
             ^
             |
+----------------------------+
| External Systems (e.g. API)|
+----------------------------+
```

### Integration Goal

- **Progressive Fetching**: Fetch and enrich `Word` objects incrementally, especially for missing fields like translations, examples, or synonyms.
- **Domain Isolation**: The `Word` aggregate remains insulated from schema differences in external data.

---

## **3. Responsibilities**

### **3.1 Fetch Data From External Sources**

The Word Lookup Service fetches raw word-related data. These sources include:

- **Scrapers**: Extracting data from Wiktionary HTML pages or similar.
- **APIs**: Querying public dictionary services.

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

### **3.2 Normalize, Transform, and Enrich Data**

Once raw data is fetched, it is transformed into `domain.Word` objects. New behavior is added to support **partial updates** and enrichment workflows.

#### Example Transformation Workflow

1. Wiktionary raw type "nom" → Domain `WordType = "noun"`.
2. Missing translations marked as incomplete in `EnrichmentStatus`.
3. Merge synonyms from scraped data into existing Word records.

---

### **3.3 Handling Partial Data**

#### Progressive Enrichment Workflow

- First, existing `EnrichmentStatus` is checked.
- If specific fields (e.g., translations) are incomplete:
  - Lookup requests **fetch missing data only** from external sources.
  - Merge data with the existing domain object in the Word domain.

#### Prioritized Field Fetching

The Lookup Service focuses on **missing or incomplete fields**, moving through priorities (definitions > translations > synonyms) as requested.

Example Enrichment Call:

```go
func (s *LookupService) EnrichMissingFields(
    wordID string,
    enrichment EnrichmentStatus,
) (domain.Word, error)
```

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

#### Updated Adapter Behavior for Partial Handling

- Check for incomplete fields in `EnrichmentStatus` before processing raw responses.
- Return updated `Word` objects with only fetched fields.

---

### **4.3 LookupService**

#### Updated Example Methods

```go
type LookupService struct {
    adapter WiktionaryAdapter
}

// Fetch data for missing fields only
func (s *LookupService) EnrichMissingFields(
    wordID string,
    enrichment EnrichmentStatus,
) (domain.Word, error) {
    // Fetch data for incomplete fields only
    ...
}

// Full word fetch (creates new Words without enrichment context)
func (s *LookupService) GetWord(text string, language string) (domain.Word, error) {
    return s.adapter.FetchAndTransformWord(text, language)
}
```

---

## **5. Intended Workflow**

### Scenario 1: Searching and Creating a Word

1. User searches for "marécage."
2. `LookupService` fetches data and returns a **partial Word**, marking missing fields in `EnrichmentStatus`.

---

### Scenario 2: Enriching Partial Fields

1. System checks `EnrichmentStatus` metadata:
   - Missing translations trigger further Lookup requests.
2. The updated Word object merges new translations, maintaining consistency.

---

## **6. Error Handling**

### Graceful Partial Returns

- If external systems fail to supply complete data, the Lookup Service:
  - Returns valid but incomplete Word objects.
  - Fills `EnrichmentStatus` to track missing data.
  -  **API Failures**: Raise retry jobs for temporary external system failures.
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
