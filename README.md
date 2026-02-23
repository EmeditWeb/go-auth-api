# Go-Auth-API

A **Production-Ready Authentication System** built with Go & PostgreSQL.

This project is a high-performance, stateless REST API designed to handle user registration, secure login, and session management using Bearer Tokens. Built with a focus on clean architecture and security best practices, it serves as a robust foundation for modern web applications.

---

## 🏗️ Project Organogram & Architecture

### System Logic Flow

This diagram illustrates how the application initializes and how requests flow through the security layers to reach the data.

```plaintext
        ┌─────────────────────────────────────────────────────────┐
        │                      MAIN.GO (Entry)                    │
        │  1. Loads .env Config      3. Initializes Data Models   │
        │  2. Connects Postgres      4. Starts HTTP Server        │
        └──────────────┬───────────────────────────▲──────────────┘
                       │                           │
               [Mounts Routes]              [Dependency Injection]
                       │                           │
        ┌──────────────▼──────────────┐    ┌───────┴──────────────┐
        │        MIDDLEWARE.GO        │    │      HANDLERS.GO      │
        │       (The Guardian)        │    │    (The Task Logic)   │
        │                             │    │                      │
        │ Extracts Bearer Tokens,     │    │ Processes Requests,  │
        │ validates via DB, & injects │    │ interacts with models│
        │ User into Request Context.  │    │ and returns JSON.    │
        └─────────────────────────────┘    └──────────────────────┘
```

---

## 📁 Project Directory Structure

```plaintext
go-auth-api/
├── cmd/
│   └── api/
│       ├── context.go      # Request context management (User hand-off)
│       ├── handlers.go     # HTTP logic (Login, Register, Me, Logout)
│       ├── main.go         # Application ignition & DB setup
│       ├── middleware.go   # Authentication & Bearer token verification
│       ├── routes.go       # Route mapping & method routing
│       └── helpers.go      # JSON & Error response utilities
├── internal/
│   ├── data/
│   │   ├── models.go       # Database interaction logic (Users & Tokens)
│   │   └── tokens.go       # Cryptographic token generation
│   └── validator/
│       └── validator.go    # Data integrity & input validation engine
├── .env                    # Secure Environment variables
└── go.mod                  # Dependency management
```

---

## 🛡️ Security Implementation Details

### 1. Anonymous Structs (Payload Protection)

In `handlers.go`, we use **Anonymous Structs** to parse incoming JSON.

**The Benefit:**  
By defining exactly what fields we accept (e.g., `Username`, `Email`), we prevent **Mass Assignment attacks**. This ensures users cannot inject unauthorized fields, such as `"roles"` or `"IDs"`, directly into the database during registration.

---

### 2. Cryptographic Hashing

**Passwords:**  
Hashed with **Bcrypt (Cost Factor 12)** to protect against brute-force attacks.

**Tokens:**  
Users receive a **Base32 plaintext token** for authentication.  
The database stores only the **SHA-256 hash** of that token, ensuring active sessions remain secure even if database access is compromised.

---

## 🚀 API Testing (CURL)

### User Registration

```bash
curl -i -X POST -d '{
    "username": "your_username",
    "email": "user@example.com",
    "password": "<YOUR_PASSWORD>"
}' http://localhost:8080/v1/users
```

---

### User Login (Get Token)

```bash
curl -i -X POST -d '{
    "email": "user@example.com",
    "password": "<YOUR_PASSWORD>"
}' http://localhost:8080/v1/tokens/authentication
```

---

### Show My Profile

```bash
curl -i -H "Authorization: Bearer <YOUR_TOKEN_HERE>" \
http://localhost:8080/v1/users/me
```

---

### Logout (Revoke Token)

```bash
curl -i -X DELETE \
-H "Authorization: Bearer <YOUR_TOKEN_HERE>" \
http://localhost:8080/v1/tokens/logout
```

---

## 🚦 Getting Started

### Clone & Install

```bash
git clone https://github.com/Emeditweb/go-auth-api.git
go mod tidy
```

---

### Configure `.env`

Create a `.env` file in the root directory to manage your environment-specific variables.

---

### Run the Application

```bash
go run ./cmd/api
```

---

## 👤 Author

**Emmanuel Itighise**  
Microbiologist | Data Analyst | AI-Native Software Engineering Fellow