# **Word List Domain Documentation**

## **1. Purpose**

The **Word List Domain** is responsible for enabling users to organize their saved words into collections, providing a flexible framework for categorizing and managing vocabulary. Word Lists are user-specific and customizable, allowing users to group and review words in ways that suit their learning needs (e.g., theme-based lists such as "Philosophy Terms" or "Cooking Vocabulary").

By focusing only on the management of word groupings, this domain remains independent of the Saved Words Domain while leveraging it as a foundational reference.

---

## **2. Core Responsibilities**

1. **Custom Word Grouping**:
   - Allow users to create and manage customizable lists of words.
   - Provide a structure for thematically organizing vocabulary (e.g., grouping Saved Words by topic or language).

2. **User-Specific Ownership**:
   - Ensure that Word Lists are private to the individual user unless explicitly shared in the future.

3. **Basic Metadata**:
   - Record timestamps and tracking data for Word Lists (e.g., when they were created or last updated).
   - Maintain references to the Saved Words included in each Word List.

4. **Support Future Collaboration**:
   - Provide the foundation for additional features, such as sharing Word Lists or creating collaborative/communal lists.

---

## **3. Entities**

## **Entity: WordList**

The **WordList** entity represents a single, user-defined collection of saved words.

### **Attributes**

- **`id`**: A globally unique identifier for the word list.
- **`userId`**: Reference to the user who owns the word list (from the **User Domain**).
- **`name`**: The name of the word list (e.g., "Philosophy Terms").
- **`description`** (Optional): A brief user-provided description of the list's purpose or contents.
- **`savedWordIds`**: A list of references to `SavedWord` entities included in the word list (from the **User's Word Domain**).
- **`createdAt`**: The timestamp indicating when the word list was created.
- **`updatedAt`**: The timestamp indicating the last time the word list was modified.

### **Responsibilities**

- Manage collections of words grouped by users for personal organization.
- Ensure references to Saved Words are valid and user-specific.
- Provide a clear structure for thematic or topical word organization.

---

## **4. Invariants and Constraints**

### **WordList Rules**

1. **Ownership**
   - Each `WordList` must reference a valid `userId` from the **User Domain**.
   - Word Lists are private by default and only accessible to their creator.

2. **Name Uniqueness**:
   - The `name` of each Word List must be unique within the scope of the same user (e.g., a user cannot create two lists called "Philosophy Terms").

3. **Valid Saved Words**:
   - All `savedWordIds` in a `WordList` must reference valid `SavedWord` entities (from the **User's Word Domain**) that belong to the same `userId`.

4. **Timestamps**:
   - `createdAt` is immutable once set.
   - `updatedAt` must always reflect the most recent change to the Word List's properties or contents.

---

## **5. Future Considerations**

1. **Collaborative Lists**:
   - Enable sharing Word Lists with other users (e.g., inviting others to view, comment on, or contribute to a list).
   - Introduce permissions or privacy settings for shared lists (e.g., public vs. private vs. shared-with-specific-users).

2. **Global or Public Word Lists**:
   - Allow Word Lists to be made publicly available for other users to browse or duplicate into their own collections.
   - Support community-driven Word Lists that enable contributions or votes for additions (e.g., "Most Important Vocabulary for IELTS").

3. **List Metadata Expansion**:
   - Add richer metadata tracking for Word Lists, such as:
     - Statistics (e.g., number of words mastered in the list).
     - Progress tracking for linked vocabulary.

4. **Gamification Integration**:
   - Reward users for creating and maintaining Word Lists (e.g., earning badges for curating lists or reaching milestones with associated words).

---

## **6. Summary**

The **Word List Domain** provides the framework for users to organize their saved vocabulary into user-defined thematic lists. By focusing solely on the structuring and management of these lists, the domain remains independent and can evolve to support collaboration, communal lists, or advanced organizational features in the future.

This separation ensures scalability and loose coupling with other domains like Saved Words, Tags, and User interactions.
