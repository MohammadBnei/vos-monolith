# **Word Domain Documentation**

- [**Word Domain Documentation**](#word-domain-documentation)
  - [**1. Purpose**](#1-purpose)
  - [**2. Core Responsibilities**](#2-core-responsibilities)
  - [**3. Entities and Value Objects**](#3-entities-and-value-objects)
    - [**Entity: Word**](#entity-word)
      - [**Attributes**](#attributes)
      - [**Responsibilities**](#responsibilities)
    - [**Entity: Definition**](#entity-definition)
      - [**Attributes**](#attributes-1)
      - [**Responsibilities**](#responsibilities-1)
    - [**Value Object: Translation**](#value-object-translation)
      - [**Attributes**](#attributes-2)
      - [**Responsibilities**](#responsibilities-2)
  - [**4. Domain Relationships**](#4-domain-relationships)
    - [**Saved Words Domain**](#saved-words-domain)
    - [**Gamification Domain**](#gamification-domain)
  - [**5. Invariants and Constraints**](#5-invariants-and-constraints)
    - [**1. Data Integrity**](#1-data-integrity)
    - [**2. Translations**](#2-translations)
    - [**3. Pronunciation and Examples**](#3-pronunciation-and-examples)
  - [**6. Future Considerations**](#6-future-considerations)
    - [1. Enrichment Pipelines](#1-enrichment-pipelines)
    - [2. Community Contributions](#2-community-contributions)
    - [3. Hierarchical Definitions](#3-hierarchical-definitions)
    - [4. Cross-Linguistic Relations](#4-cross-linguistic-relations)
    - [**7. Summary**](#7-summary)

---

## **1. Purpose**

The **Word Domain** acts as the core domain for managing and storing vocabulary-related data. Its primary role is to represent the immutable linguistic properties of a word, such as its definitions, etymology, pronunciation, examples, and relationships (e.g., synonyms). This domain supports vocabulary functionality and ensures that the underlying data is consistent, accurate, and extensible.

The Word Domain is **agnostic of user interactions** (e.g., saving words, tagging words) to maintain separation from personalization or gamification concerns, which are handled in their respective domains.

---

## **2. Core Responsibilities**

The Word Domain is responsible for:

1. **Vocabulary Representation**:
   - Act as the canonical source of truth for all word-related data (e.g., definitions, translations, etymology, synonyms, antonyms).
   - Maintain a consistent structure for multilingual word definitions.
2. **Contextual Enrichment**:
   - Enable the addition of dynamic, contextual examples for words to help users understand their real-world usage.
3. **Protection of Data Integrity**:
   - Ensure that linguistic data maintains its consistency and invariants (e.g., no duplicate definitions, translations grouped by language).
4. **Support for Enrichment Pipelines**:
   - Integration with external sources for enriching word data over time (e.g., adding missing etymology, new examples).

---

## **3. Entities and Value Objects**

### **Entity: Word**

The **Word** entity represents a single unit of vocabulary in the system, encapsulating its core meaning and associated linguistic data. It serves as the **Aggregate Root** for related entities and value objects within the domain (e.g., `Definition`, `Translation`).

#### **Attributes**

- **`id`**: A globally unique identifier for the word.
- **`text`**: The canonical representation of the word (e.g., "ephemeral" or "marécage").
- **`language`**: ISO 639-1 code representing the word's language (e.g., "en" for English, "fr" for French).
- **`definitions`**: A collection of `Definition` entities describing the meanings of the word.
- **`pronunciation`**:
  - Phonetic Guide: A string-based representation of the word's pronunciation.
  - Audio Reference: Potentially a link to an external resource for spoken audio.
- **`etymology`**: Historical origin of the word (optional—may be missing for certain words).
- **`synonyms`** (List of Strings): Words with similar meanings.
- **`antonyms`** (List of Strings): Words with opposite meanings.
- **`translations`** (Map): Language-based translations for the word, stored as a map of language codes (`ISO 639-1`) to lists of translated terms.

#### **Responsibilities**

- Store the canonical representation of linguistic data for a word.
- Provide access points for related data (e.g., definitions, examples, synonyms).
- Ensure data consistency by enforcing domain-specific invariants.

---

### **Entity: Definition**

The **Definition** entity describes **contextual and grammatical meanings** of a word.

#### **Attributes**

- **`id`**: A globally unique identifier for the definition.
- **`text`**: The explanation or meaning of the word (e.g., "Short-lived; transient; lasting a short time").
- **`wordType`**: The grammatical type of the word (e.g., noun, verb, adjective).
- **`examples`** (List of Strings): Sentences showcasing the usage of the word in the given meaning.
- **`gender`** (Optional): Gender information (if applicable, e.g., for gendered languages like French or Spanish).
- **`pronunciation`** (Optional): The pronunciation of the word in its specific context (e.g., IPA format for phonetics).
- **`languageSpecifics`**: Additional linguistic details (e.g., plural forms, case sensitivity) specific to the word's language.

#### **Responsibilities**

- Provide detailed descriptions of a word's meaning.
- Support nuanced contexts through examples and language-specific data.

---

### **Value Object: Translation**

The **Translation** value object defines multilingual representations of a word, storing translations by target language.

#### **Attributes**

- **`language`**: The ISO code of the target language (e.g., "en", "de").
- **`terms`** (List of Strings): A list of translated terms in the target language.

#### **Responsibilities**

- Provide a clear and consistent structure for multilingual translations associated with the word.
- Group translations uniquely by target language, ensuring no duplicates.

---

## **4. Domain Relationships**

The Word Domain interacts with other domains in an **isolated manner**. Cross-domain interactions reference Word data through its immutable `id` or properties but do not directly depend on Word logic.

### **Saved Words Domain**

- **Relationship**:
  - The Saved Words domain references the Word domain by its `id`, allowing users to save, tag, and organize words.
- **Directionality**:
  - Saved Words consumes word data but does not modify it.
- **Impact**:
  - User-specific metadata (e.g., tags) is stored in Saved Words, leaving the Word domain untouched.

### **Gamification Domain**

- **Relationship**:
  - Gamification references Word entities via challenges or achievements (e.g., completing 10 questions about synonyms for "ephemeral").
- **Directionality**: Gamification queries Word data but does not alter it.

---

## **5. Invariants and Constraints**

### **1. Data Integrity**

- Words cannot share the same combination of `text` and `language` (e.g., no duplicate entries for "ephemeral" in English).
- Definitions for a word must be unique within the context of the word.

### **2. Translations**

- Translations must be grouped uniquely by language in the `translations` map (e.g., no duplicate entries for "de" → ["Sumpf"]).
- Language codes must adhere to ISO 639-1 standards.

### **3. Pronunciation and Examples**

- Pronunciation is optional but must conform to a standardized format (e.g., IPA for phonetics).
- Examples must be unique within the word's context (e.g., avoid repeated entries).

---

## **6. Future Considerations**

### 1. Enrichment Pipelines

- Connect the Word domain to external lexical sources (e.g., Wiktionary, Oxford API) for contextual enrichment. Data gaps like missing etymology or examples could be filled automatically.

### 2. Community Contributions

- Allow user-submitted examples, comments, or contextual meanings via a collaboration domain. Submissions would pass through moderation workflows.

### 3. Hierarchical Definitions

- Support hierarchical relationships between definitions (e.g., primary definition vs. extended meanings) to enhance display for complex words.

### 4. Cross-Linguistic Relations

- Extend synonyms/antonyms to cover cross-linguistic relationships (e.g., "Sumpf" in German is a synonym for "Swamp" in English).

---

### **7. Summary**

The Word Domain encapsulates and protects the linguistic data of the system, defining the structure and rules for managing vocabulary. The domain is tightly focused on ensuring the consistency and accuracy of word data while isolating itself from user-specific or gamification-related logic. This self-contained nature ensures scalability and seamless interaction with other domains.
