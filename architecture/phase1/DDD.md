# **Domain-Driven Design Document**

**Project Name**: Voc on Steroids  
**Phase**: Phase 1 (Core Functionality)

---

**Table of Contents**

- [**Domain-Driven Design Document**](#domain-driven-design-document)
  - [**1. High-Level Overview**](#1-high-level-overview)
    - [Purpose](#purpose)
    - [Core Domains in Phase 1](#core-domains-in-phase-1)
  - [**2. Domain Models**](#2-domain-models)
    - [**2.1 Word Lookup Domain**](#21-word-lookup-domain)
      - [Purpose](#purpose-1)
      - [Models](#models)
        - [**Word**](#word)
        - [**Definition**](#definition)
      - [Key Use Cases](#key-use-cases)
    - [**2.2 User Authentication Domain**](#22-user-authentication-domain)
      - [Purpose](#purpose-2)
      - [Models](#models-1)
        - [**User**](#user)
      - [Key Use Cases](#key-use-cases-1)
    - [**2.3 Word Saving and Management Domain**](#23-word-saving-and-management-domain)
      - [Purpose](#purpose-3)
      - [Models](#models-2)
        - [**SavedWord**](#savedword)
        - [**List**](#list)
      - [Key Use Cases](#key-use-cases-2)
    - [**2.4 User Preferences Domain**](#24-user-preferences-domain)
      - [Purpose](#purpose-4)
      - [Models](#models-3)
        - [**UserPreferences**](#userpreferences)
      - [Key Use Cases](#key-use-cases-3)
  - [**3. Domain Relationships**](#3-domain-relationships)
    - [Key Interactions](#key-interactions)
    - [High-Level Diagram](#high-level-diagram)
  - [**4. Services**](#4-services)
    - [**WordSavingService**](#wordsavingservice)
      - [Key Operations](#key-operations)
    - [**UserPreferenceService**](#userpreferenceservice)
      - [Key Operations](#key-operations-1)
  - [**5. Future Considerations**](#5-future-considerations)

---

## **1. High-Level Overview**

### Purpose

The application aims to enhance vocabulary learning through personalized word lookup, saving, and organization features.

---

### Core Domains in Phase 1

1. **Word Lookup**: The core functionality enabling the retrieval of detailed word information, including definitions, examples, pronunciation, synonym/antonym relations, and etymology.
2. **User Authentication**: Handles secure access to user-specific features, such as saving and organizing words.
3. **Word Saving and Management**: Enables users to save words into personalized lists, along with tagging capabilities for contextual organization.
4. **User Preferences**: A centralized object for managing user-specific settings, including their default list.

---

## **2. Domain Models**

### **2.1 Word Lookup Domain**

#### Purpose

Supports search and retrieval of word-related information from dynamic sources (e.g., Wiktionary scraping). Provides consolidated data to other domains.

#### Models

##### **Word**

Represents a core vocabulary word, aggregating its definitions, metadata, and usage.

```go
type Word struct {
  ID            string            `json:"id"`            // Unique identifier (UUID)
  Text          string            `json:"text"`          // Canonical form of the word
  Language      string            `json:"language"`      // ISO language code (e.g., "fr")
  Definitions   []Definition      `json:"definitions"`   // Array of definitions
  Etymology     string            `json:"etymology,omitempty"` // Word origin
  Translations  map[string]string `json:"translations,omitempty"` // Translations to other languages
  Synonyms      []string          `json:"synonyms,omitempty"`     // Related synonyms
  Antonyms      []string          `json:"antonyms,omitempty"`     // Related antonyms
  SearchTerms   []string          `json:"search_terms"`     // Variants for search
  Lemma         string            `json:"lemma,omitempty"`  // Root/base form
  UsageNotes    []string          `json:"usage_notes,omitempty"` // Notes on usage
  CreatedAt     time.Time         `json:"created_at"`
  UpdatedAt     time.Time         `json:"updated_at"`
}
```

##### **Definition**

Represents the meaning and usage details of a word. Supports multilingual and grammatical data.

```go
type Definition struct {
  ID               string            `json:"id"`               // Unique ID for this definition
  Text             string            `json:"text"`             // The actual definition text
  WordType         string            `json:"word_type"`        // Word's grammatical type (noun, verb, etc.)
  Examples         []string          `json:"examples"`         // Example sentences
  Gender           string            `json:"gender,omitempty"` // Gender info for gendered languages
  Pronunciation    string            `json:"pronunciation"`    // Phonetic/IPA representation
  LanguageSpecifics map[string]string `json:"language_specifics,omitempty"` // E.g., plural forms
  Notes            []string          `json:"notes,omitempty"`  // Additional usage metadata
  CreatedAt        time.Time         `json:"created_at"`       // Timestamp for when the definition was created
  UpdatedAt        time.Time         `json:"updated_at"`       // Timestamp for the last update to the definition
}
```

#### Key Use Cases

- **Search a Word**: Retrieve details for a specific word based on user input.
- **Integration**: Provide word data to downstream systems (e.g., for saving into lists).

---

### **2.2 User Authentication Domain**

#### Purpose

Provides secure, personalized access to app features via registration and authentication.

#### Models

##### **User**

Represents a system user with credentials and roles.

```go
type User struct {
  ID            string `json:"id"`
  Email         string `json:"email"`
  PasswordHash  string `json:"password_hash"`
  Role          string `json:"role"` // Role-based authorization (e.g., "admin", "learner")
  CreatedAt     time.Time `json:"created_at"`
  UpdatedAt     time.Time `json:"updated_at"`
}
```

#### Key Use Cases

- **Sign Up**: Register a new user.
- **Log In**: Authenticate a user and provide an access token.

---

### **2.3 Word Saving and Management Domain**

#### Purpose

Provides the ability to save, organize, and tag words for personalized vocabulary management.

#### Models

##### **SavedWord**

Represents a saved word, linked to a `Word` and containing user-specific metadata.

```go
type SavedWord struct {
  ID          string    `json:"id"`
  UserID      string    `json:"user_id"`
  WordID      string    `json:"word_id"` 
  ListID      string    `json:"list_id"` 
  Tags        []string  `json:"tags"`    
  DateSaved   time.Time `json:"date_saved"`
  LastReviewed time.Time `json:"last_reviewed,omitempty"`  
}
```

##### **List**

Represents a user-defined collection of saved words.

```go
type List struct {
  ID        string    `json:"id"`
  UserID    string    `json:"user_id"`
  Name      string    `json:"name"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}
```

#### Key Use Cases

- **Save Word**: Save a word into a specific list.
- **Organize Words**: Create and manage lists.
- **Tag Saved Words**: Add tags for meta-organization of saved words.

---

### **2.4 User Preferences Domain**

#### Purpose

Centralized object for managing user-specific preferences, including the default list.

#### Models

##### **UserPreferences**

Encapsulates configurable settings for a user.

```go
type UserPreferences struct {
  UserID       string    `json:"user_id"`
  DefaultList  string    `json:"default_list"`  // ID of the default list
  DefaultLang  string    `json:"default_lang"`  // Default search language
  NotificationsEnabled bool `json:"notifications_enabled"`
  CreatedAt    time.Time `json:"created_at"`
  UpdatedAt    time.Time `json:"updated_at"`
  }
```

#### Key Use Cases

- Set a user's default list for saving words.
- Update preferences when required.

---

## **3. Domain Relationships**

### Key Interactions

Here's how the domains interconnect:

1. **User → User Preferences**
    - A user owns their preferences, which define their default behaviors (e.g., default list, default language).

2. **User → Lists**
    - A user owns multiple lists for organizing saved words.

3. **List → SavedWord**
    - Each `SavedWord` belongs to one list.

4. **SavedWord → Word**
    - A saved word is linked to the core `Word` entity from the Lookup domain.

---

### High-Level Diagram

```
User
 ├── Owns → Preferences
 |                      ├── DefaultList → List
 |                      └── DefaultLang
 ├── Owns → Lists               ─── Contains → SavedWord
 ├── Saves → SavedWord        │
 └── Uses → Word Lookup ───→ Word (definitions, translations, metadata)
```

---

## **4. Services**

### **WordSavingService**

Handles word saving and management logic.

#### Key Operations

- `SaveWord(userID, wordID, listID, tags)`: Saves a word to a list.
- `GetSavedWords(userID, listID)`: Retrieves saved words for a given list.
- `CreateList(userID, name)`: Creates a new list.
- `SetDefaultList(userID, listID)`: Updates the default list in User Preferences.

---

### **UserPreferenceService**

Handles user preferences.

#### Key Operations

- `GetPreferences(userID)`: Retrieves user preferences.
- `UpdatePreferences(userID, preferences)`: Updates the user's preferences.

---

## **5. Future Considerations**

1. **Phase 2**:
    - Integrate spaced repetition algorithms into the saved words system.
    - Introduce gamification features (e.g., leaderboards, badges).

2. **Data Enrichment**:
    - Support fetching richer contextual examples (e.g., pulling example sentences from books and news).

3. **Multi-Language Support**:
    - Enable searching and saving words across multiple languages.
