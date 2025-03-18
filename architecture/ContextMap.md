# **Context Map: Vocabulary Application**

- [**Context Map: Vocabulary Application**](#context-map-vocabulary-application)
  - [**1. Purpose**](#1-purpose)
  - [**2. Bounded Contexts**](#2-bounded-contexts)
    - [**Core Domains**](#core-domains)
    - [**Subdomains**](#subdomains)
  - [**3. Context Map Diagram**](#3-context-map-diagram)
    - [**Overview of Domain Interactions**](#overview-of-domain-interactions)
    - [**Explanation of the Diagram:**](#explanation-of-the-diagram)
  - [**4. Relationships Between Domains**](#4-relationships-between-domains)
    - [**4.1 User ↔ User's Word Domain**](#41-user--users-word-domain)
    - [**4.2 Word ↔ User's Word Domain**](#42-word--users-word-domain)
    - [**4.3 User's Word Domain ↔ Word List Domain**](#43-users-word-domain--word-list-domain)
    - [**4.4 Tagging Domain ↔ Saved Words / Word Lists**](#44-tagging-domain--saved-words--word-lists)
  - [**5. Integration Patterns**](#5-integration-patterns)
    - [**1. References (Loose Coupling)**](#1-references-loose-coupling)
    - [**2. Event-Driven Interactions**](#2-event-driven-interactions)
    - [**3. Separate Bounded Contexts**](#3-separate-bounded-contexts)
  - [**6. Challenges and Trade-Offs**](#6-challenges-and-trade-offs)
    - [**1. Modular Simplicity vs. Integration Complexity**](#1-modular-simplicity-vs-integration-complexity)
    - [**2. Expansion of Shared References**](#2-expansion-of-shared-references)
    - [**3. Performance Overhead**](#3-performance-overhead)
  - [**7. Summary**](#7-summary)

## **1. Purpose**

The **Context Map** provides a high-level view of the system architecture, showcasing the different bounded contexts (domains) and their relationships. It ensures:

- **Clarity in Responsibilities**: Identify what each domain owns and its role in the system.
- **Separation of Concerns**: Highlight boundaries between domains to avoid coupling them unnecessarily.
- **Collaboration Patterns**: Explain how domains interact via references, data flow, and integration patterns.

---

## **2. Bounded Contexts**

The system is divided into **core domains** (essential for delivering business value) and **subdomains** (support-specific behaviors related to the core domains).

### **Core Domains**

1. **User Domain**: Manages user accounts, authentication, and preferences.
2. **Word Domain**: Acts as the canonical source for vocabulary-related data, including definitions, synonyms, antonyms, and etymology.

### **Subdomains**

1. **User's Word Domain**: Tracks associations between users and words they've saved, acting as the foundational connection for user-vocabulary personalization.
2. **Tagging Domain**: Provides a reusable framework for applying user-defined tags to various objects (e.g., Saved Words or Word Lists).
3. **Word List Domain**: Responsible for organizing saved words into user-curated collections/themes.

---

## **3. Context Map Diagram**

### **Overview of Domain Interactions**

```plaintext
                       +---------------------+        +----------------------+
                       |    Word Domain      |        |     User Domain      |
                       |  (Vocabulary Data)  |        |  (Users & Accounts)  |
                       +---------------------+        +----------------------+
                                 |                            |
                    References wordId                References userId
                                 |                            |
                  +--------------+                            +------------------+
                  |                                                       |
                  V                                                       V
        +-----------------------+                      +-----------------------+
        |   User's Word Domain  |--------------------->|   Word List Domain     |
        |  (Saved Words Tracker)| References wordId   |  (Organized Collections)|
        +-----------------------+ Tracks saved words  +----------------------- +
                  |                        \                    /
          References userId                \                  /
                  |                         \                /
                  V                          \              /
        +-----------------------+             +----------------------+
        |    Tagging Domain     |------------->  Tracks tags on words |
        |  (User-defined Tags)  | <------------- Tracks lists/objects |
        +-----------------------+  Tracks tags          Tracks tags
```

### **Explanation of the Diagram:**

1. **Core Domains**:
   - **User Domain**:
     - Manages user accounts, identity, and ownership (`userId`).
     - Provides a reference to subdomains such as **User's Word Domain** and **Word List Domain** for associating user-specific data (e.g., saved words, lists).

   - **Word Domain**:
     - Maintains vocabulary as immutable entities (`wordId`).
     - Provides read-only access to vocab data (e.g., definitions, synonyms, examples) to subdomains like the **User's Word Domain**.

2. **User's Word Domain**:
   - Tracks saved words:
     - Associates `userId` (from **User Domain**) with `wordId` (from **Word Domain**) to record users' saved vocabulary.

3. **Word List Domain**:
   - Builds on `SavedWord` entities:
     - Organizes `SavedWord` references into collections (e.g., themed lists like "Philosophy Terms").

4. **Tagging Domain**:
   - Provides a reusable tagging system:
     - Tracks tags applied to `SavedWord` or **Word List** entities.
     - Enables users to manage their tags without coupling directly to the objects being tagged.

---

## **4. Relationships Between Domains**

### **4.1 User ↔ User's Word Domain**

- **Integration**:
  - **Reference**: `userId` (from the **User Domain**) is the key to associating a user's saved words.
- **Pattern**:
  - The **User's Word Domain** depends on the **User Domain** for validating associations or ownership (e.g., ensuring a saved word belongs to the correct user).

### **4.2 Word ↔ User's Word Domain**

- **Integration**:
  - **Reference**: `wordId` (from the **Word Domain**) connects saved words to immutable vocabulary data.
- **Pattern**:
  - **Read-only Reference**: The **Word Domain** provides read-only content (e.g., definitions, translations) while the **User's Word Domain** manages user-specific metadata.

### **4.3 User's Word Domain ↔ Word List Domain**

- **Integration**:
  - **Reference**: The **Word List Domain** references `SavedWord` entities (from the User's Word Domain) for organizing words into collections.
- **Pattern**:
  - **Dependency**: Word Lists depend on the existence of Saved Words.

### **4.4 Tagging Domain ↔ Saved Words / Word Lists**

- **Integration**:
  - **Generic Reference**: Tags are loosely coupled to objects using their `id` and `type` (`targetId` and `targetType`), supporting tagging for:
    - `SavedWord` (saved vocabulary).
    - `WordList` (curated collections).
- **Pattern**:
  - The **Tagging Domain** does not directly depend on Word List or User's Word. It is domain-agnostic and references objects dynamically.

---

## **5. Integration Patterns**

### **1. References (Loose Coupling)**

- Subdomains interact via **shared references** (e.g., `userId`, `wordId`) rather than direct ownership. This ensures modularity:
  - The **User's Word Domain** depends only on `wordId` to track saved words without duplicating or owning vocabulary data.
  - Tags and Word Lists depend on `SavedWord` from the **User's Word Domain** but do not directly manage word ownership.

### **2. Event-Driven Interactions**

In future iterations, interactions could evolve into event-based patterns for better scalability:

- Example: When a `SavedWord` is created:
  - The **Tagging Domain** could subscribe to a "SavedWordCreated" event to add default tags.
  - The **Word List Domain** could use events to maintain list consistency when words are tagged.

### **3. Separate Bounded Contexts**

Each domain adheres to **DDD's bounded context principle**:

- Core data and behavior stay inside the domain boundary.
- Cross-domain interactions occur only through **references** or **read-only queries**.

---

## **6. Challenges and Trade-Offs**

### **1. Modular Simplicity vs. Integration Complexity**

- Keeping domains separate increases modularity but requires REST-like or event-driven communication for cross-domain operations (e.g., querying tags for a Word List).

### **2. Expansion of Shared References**

- As domains grow, managing shared keys (`wordId`, `savedWordId`, etc.) requires consistency (e.g., using UUIDs globally).

### **3. Performance Overhead**

- Querying across domains (e.g., listing SavedWords in Word Lists with Tags attached) may require a **domain service layer** or a pre-aggregated solution.

---

## **7. Summary**

The **Context Map** showcases the clear and modular design of your system:

- **Core Domains (User and Word)** serve as the foundation.
- **Subdomains (User's Word, Tagging, Word Lists)** enable specialized behavior for user-vocabulary interactions.
By leveraging loose coupling and shared references (`userId`, `wordId`), the system remains highly extensible while adhering to separation of concerns. This structure allows for straightforward scaling (e.g., adding collaboration, gamification, or analytics).
