# **User Domain Documentation**

- [**User Domain Documentation**](#user-domain-documentation)
  - [**1. Purpose**](#1-purpose)
  - [**2. Core Responsibilities**](#2-core-responsibilities)
  - [**3. Entities and Value Objects**](#3-entities-and-value-objects)
    - [**Entity: User**](#entity-user)
      - [**Attributes**](#attributes)
      - [**Responsibilities**](#responsibilities)
    - [**Value Object: UserPreferences**](#value-object-userpreferences)
      - [**Attributes**](#attributes-1)
      - [**Responsibilities**](#responsibilities-1)
  - [**4. Domain Relationships**](#4-domain-relationships)
    - [**Saved Words Domain**](#saved-words-domain)
    - [**Gamification Domain**](#gamification-domain)
    - [**Authentication Layer**](#authentication-layer)
  - [**5. Invariants and Constraints**](#5-invariants-and-constraints)
    - [**1. Authentication Constraints**](#1-authentication-constraints)
    - [**2. Verification Constraints**](#2-verification-constraints)
    - [**3. Relationship Constraints**](#3-relationship-constraints)
  - [**6. Future Considerations**](#6-future-considerations)
    - [**1. Expanded Security Options**](#1-expanded-security-options)
    - [**2. UserPreferences Evolution**](#2-userpreferences-evolution)
    - [**3. Anonymous Session Support**](#3-anonymous-session-support)
    - [**4. Subscription Integration**](#4-subscription-integration)
    - [**7. Summary**](#7-summary)

---

## **1. Purpose**

The **User Domain** handles the management of user accounts, serving as the foundation for all user-specific system functionality. Its primary responsibility is to manage user data, ensure account security, and enable personalization through user preferences.

The User domain acts as a **self-contained, independent domain** providing identifiers for interaction with other domains such as Saved Words, Gamification, and others.

---

## **2. Core Responsibilities**

The User domain is responsible for:

1. **Identity and Authentication**:
   - Managing user account creation, login, and password storage.
   - Providing uniquely identifiable `userId`s for all interactions.
2. **Email Verification**:
   - Enabling and enforcing email validation during user registration.
   - Managing unverified-user restrictions (e.g., limited features until verification).
3. **Password Management**:
   - Securely storing passwords.
   - Enabling password reset functionality.
4. **Personalization**:
   - Providing user-customizable preferences like language or notification settings.
5. **Account-Level Security**:
   - Enforcing constraints on account creation (e.g., unique emails).
   - Locking sensitive workflows to authenticated users.

---

## **3. Entities and Value Objects**

### **Entity: User**

The **User** represents an individual account interacting with the system.

#### **Attributes**

- **`id`**: A globally unique identifier for the user.
- **`username`**: Publicly visible, unique nickname chosen by the user.
- **`email`**: The user's primary email address, used for authentication (must be unique).
- **`emailVerified`**: A flag indicating whether the email has been verified (`false` by default upon registration).
- **`passwordHash`**: Securely hashed password (plaintext is not stored).
- **`createdAt`**: The timestamp when the account was created.
- **`lastLoginAt`**: The timestamp of the user's most recent login.

#### **Responsibilities**

- Represent the user's identity and account in the system.
- Provide a secure mechanism for login, registration, and authentication.
- Track and enforce email verification status.

---

### **Value Object: UserPreferences**

The **UserPreferences** value object stores customizable user settings, providing flexibility for future evolution without impacting the core User entity.

#### **Attributes**

- **`preferredLanguage`**: The default language for multilingual content (e.g., "en" for English, "fr" for French).
- **`difficultySetting`**: Preferred level of quiz or learning difficulty (`EASY`, `MEDIUM`, `HARD`).
- **`notificationsEnabled`**: Whether the user has opted into app and/or email notifications.
- **`theme`** (Optional): User's interface preference (e.g., "dark" or "light").

#### **Responsibilities**

- Encapsulate user-specific preference options to simplify changes without directly modifying the User entity.
- Ensure that user preferences remain consistent and valid.

---

## **4. Domain Relationships**

To maintain loose coupling and separation of domains, the User domain interacts with other domains indirectly via its unique `id`. These relationships are outlined below:

### **Saved Words Domain**

- **Relationship**:
  - The Saved Words domain references the `userId` to track which users have saved words or organized word lists.
- **Directionality**: Saved Words references User but does not depend on its internal details.
- **Restrictions for Unverified Users**:
  - Saved Words enforces word-saving limits (e.g., a 10-word cap) for unverified users based on verification status retrieved from the User domain.

---

### **Gamification Domain**

- **Relationship**:
  - The Gamification domain references the `userId` to associate achievements, streaks, and rewards with individual users.
- **Directionality**: The User domain generates the identifier (`id`), but no gamification-specific logic resides in the User domain.

---

### **Authentication Layer**

- **Relationship**:
  - The User domain authenticates and generates session tokens for user interactions.
- **Directionality**: Authentication interacts exclusively through the User entity and does not interface with broader system behavior.

---

## **5. Invariants and Constraints**

### **1. Authentication Constraints**

- Passwords are never stored in plaintext. Only secure, hashed passwords can be stored (e.g., with bcrypt or other hashing mechanisms).
- Email and username must be unique across all users.
- Password strength rules must ensure minimum complexity, such as:
  - Minimum length of 8 characters, including at least one uppercase letter, one number, and one special character.

### **2. Verification Constraints**

- A user's `emailVerified` flag must always default to `false` at registration.
- Verification tokens expire after a set duration (e.g., 48 hours) and become invalid upon successful verification.

### **3. Relationship Constraints**

- The `id` field for users must always be globally unique and immutable.
- No other domain stores User domain-specific data (e.g., preferences or passwords).

---

## **6. Future Considerations**

### **1. Expanded Security Options**

- Implement **multi-factor authentication (MFA)** to enhance login security.
- Consider security measures for failed login attempts (e.g., temporary account lockout or CAPTCHA challenges).

### **2. UserPreferences Evolution**

- As features expand, preferences could grow to include:
  - Dashboard layout options (e.g., word graphs, gamification progress).
  - Advanced notification settings (e.g., daily reminders, weekly word reviews).
  - Collaborative features (e.g., saving lists shared with others).

### **3. Anonymous Session Support**

- If anonymous users are introduced for browsing or basic app features:
  - Introduce temporary session-based storage for non-authenticated users (e.g., temporary word saving with restrictions).

### **4. Subscription Integration**

- Include subscription or payment data:
  - A `Subscription` entity could join with `userId` to handle freemium/premium plans.

---

### **7. Summary**

The User Domain provides the foundational mechanisms for account management, authentication, and user-specific customization. It interacts loosely with Saved Words and Gamification while remaining decoupled and self-contained. The domain is optimized for scalability, enabling easy integration of future features like security upgrades, expanded personalization, and subscription services.
