# **Tagging Domain Documentation**

- [**Tagging Domain Documentation**](#tagging-domain-documentation)
  - [**1. Purpose**](#1-purpose)
  - [**2. Core Responsibilities**](#2-core-responsibilities)
  - [**3. Entities**](#3-entities)
  - [**Entity: Tag**](#entity-tag)
    - [**Attributes**](#attributes)
    - [**Responsibilities**](#responsibilities)
  - [**Entity: TagAssignment**](#entity-tagassignment)
    - [**Attributes**](#attributes-1)
    - [**Responsibilities**](#responsibilities-1)
  - [**4. Invariants and Constraints**](#4-invariants-and-constraints)
    - [**Tag Rules**](#tag-rules)
    - [**TagAssignment Rules**](#tagassignment-rules)
  - [**5. Future Considerations**](#5-future-considerations)
  - [**6. Summary**](#6-summary)

## **1. Purpose**

The **Tagging Domain** is responsible for managing the assignment of user-specific tags to various objects in the system. It provides a flexible and extensible framework for categorizing and organizing entities like Saved Words, Word Lists, or other future objects.

This domain is designed to remain independent yet integrate seamlessly with other domains by supporting object tagging while maintaining referential integrity.

---

## **2. Core Responsibilities**

1. **Tag Management**:
   - Allow users to create, assign, and manage custom tags for organizing and contextualizing objects across different domains.

2. **Generic Tag Assignment**:
   - Enable the tagging of any supported object type (e.g., Saved Words, Word Lists) by linking tags to target IDs and object types.

3. **Reusability and Extensibility**:
   - Ensure the tagging system is reusable across multiple domains without coupling tightly to specific logic or workflows.

4. **User Ownership**:
   - Maintain user-level ownership of private tags while ensuring no conflicts occur across multiple users.

---

## **3. Entities**

## **Entity: Tag**

The **Tag** entity represents a user-defined label that can be applied to various types of objects.

### **Attributes**

- **`id`**: A unique identifier for the tag.
- **`userId`**: The identifier of the user who owns the tag (from the **User Domain**).
- **`name`**: The user-defined label or name of the tag (must be unique within the context of a single user).
- **`createdAt`**: The timestamp indicating when the tag was created.
- **`updatedAt`**: The timestamp indicating the last time the tag was modified.

### **Responsibilities**

- Represent a reusable user-specific label.
- Ensure that tags remain unique per user (e.g., two tags with the name "review-later" cannot exist for the same user).
- Allow future extensibility for common/global tags in shared contexts.

---

## **Entity: TagAssignment**

The **TagAssignment** entity tracks the relationship between a tag and the object it is applied to (e.g., a Saved Word or a Word List).

### **Attributes**

- **`id`**: A unique identifier for the tag assignment.
- **`tagId`**: Reference to the `Tag` being assigned.
- **`targetId`**: The identifier of the object being tagged (e.g., a Saved Word or Word List).
- **`targetType`**: The type of object being tagged (e.g., "SavedWord," "WordList").
- **`createdAt`**: Timestamp indicating when the tag was assigned.

### **Responsibilities**

- Track the relationship between tags and objects in a clean and generic way.
- Ensure extensibility by allowing tagging for any object type in the system.

---

## **4. Invariants and Constraints**

### **Tag Rules**

1. **Unique Tag Names**:
   - Tag names must be unique within the scope of a single user's account (e.g., user A and user B may both create the tag "review-later," but user A cannot create another "review-later" tag).

2. **Ownership**:
   - Each `Tag` must reference a valid `userId` from the User Domain.

### **TagAssignment Rules**

1. **Unique Assignment**:
   - A single object (`targetId`) cannot have the same tag (`tagId`) assigned multiple times.
   - Uniqueness is enforced on the combination of `tagId`, `targetId`, and `targetType`.

2. **Target Object Validation**:
   - A `TagAssignment` must reference an existing `tagId` and a valid `targetId` in the context of the specified `targetType`.

---

## **5. Future Considerations**

1. **Domain-Agnostic Tags**:
   - Allow global, systemwide tags for public or shared content (e.g., enabling shared tags that span all users for collaborative projects or community voting systems).

2. **Cross-Object Tagging**:
   - Support tagging objects from other future domains (e.g., tagging games, collaborative lists, or challenges).

3. **Tag Hierarchies**:
   - Introduce parent-child relationships between tags to support hierarchical categorization (e.g., "Food" → "Asian Cuisine" → "Japanese Food").

4. **Tag Recommendation System**:
   - Enable support for smart tag recommendations based on user behavior or machine learning models, offering suggested tags when users save or organize objects.

---

## **6. Summary**

The **Tagging Domain** provides a robust but simple framework for managing and assigning user-specific tags across the system. With its focus on reusability and loose coupling, the domain enables scalable categorization and organization without bloating core logic in other domains. By representing tags and assignments independently, it lays the groundwork for extensibility into shared or global tags in future iterations.
