# **User's Word Domain Documentation**

- [**User's Word Domain Documentation**](#users-word-domain-documentation)
  - [**1. Purpose**](#1-purpose)
  - [**2. Core Responsibilities**](#2-core-responsibilities)
  - [**3. Entity: SavedWord**](#3-entity-savedword)
    - [**Attributes**](#attributes)
  - [**4. Invariants and Constraints**](#4-invariants-and-constraints)
  - [**5. Summary**](#5-summary)

## **1. Purpose**

The **User's Word Domain** represents the core relationship between users and the vocabulary words they save within the system. It is focused solely on recording which words have been saved by users and maintaining this association without embedding additional user-specific metadata or responsibilities.

This domain ensures that saved words are linked efficiently to their respective users and forms the foundation for user-specific interactions with vocabulary.

---

## **2. Core Responsibilities**

1. **Track Saved Words**:
   - Manage the association between users and the vocabulary words they save for personal use.
   - Persist the minimal data required to identify this relationship.

2. **Basic Metadata for Interactions**:
   - Record when a word was saved and when the user last interacted with it.

3. **Foundation for Other Features**:
   - Serve as the core building block for other user-specific vocabulary features, such as tagging, word lists, progress tracking, or analytics.

---

## **3. Entity: SavedWord**

The **SavedWord** entity represents a single instance of a word saved by a specific user. It is the core concept within the User's Word Domain.

### **Attributes**

- **`id`**: A unique identifier for the saved word.
- **`userId`**: The identifier of the user who saved the word.
- **`wordId`**: The identifier of the vocabulary word being saved.
- **`savedAt`**: The timestamp indicating when the word was saved.
- **`lastInteractedAt`**: The timestamp indicating the most recent interaction with the saved word.

---

## **4. Invariants and Constraints**

1. **Unique Saved Word**:
   - A user can only save a specific word once. The combination of `userId` and `wordId` must be unique.

2. **Valid References**:
   - Each `SavedWord` must reference:
     - A valid `userId` from the User Domain.
     - A valid `wordId` from the Word Domain.

3. **Timestamps**:
   - The `savedAt` timestamp is immutable once a SavedWord is created.
   - The `lastInteractedAt` timestamp must always reflect the most recent user-driven interaction.

---

## **5. Summary**

The **User's Word Domain** is a focused and minimal domain that manages the direct relationship between users and the words they save within the system. It tracks the essential association and metadata (like timestamps) while delegating additional features (such as tagging, lists, or gamification) to external domains.
