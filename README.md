# Go-Auth-API
![Go](https://img.shields.io/badge/Go-1.23+-blue.svg)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-14+-blue.svg)
![Docker](https://img.shields.io/badge/Docker-24.0+-blue.svg)
![GitHub](https://img.shields.io/badge/GitHub-1.0+-blue.svg)

![Image](go-auth-api.png)

A **Production-Ready Authentication System** built with Go & PostgreSQL.

This project is a high-performance, stateless REST API designed to handle user registration, secure login, and session management using Bearer Tokens. Built with a focus on clean architecture and security best practices, it serves as a robust foundation for modern web applications.

---

## Project Organogram & Architecture

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

## Project Directory Structure

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

## Security Implementation Details

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
## Database Initialization 

To run your SQL schema directly from the terminal, you need to use the PostgreSQL Command Line Interface (psql).

Here are the exact steps to get your CLI ready, connect to your database, and execute your schema.

## Getting Started

### Clone & Install

```bash
git clone https://github.com/Emeditweb/go-auth-api.git
go mod tidy
```

---

### Configure `.env`

Create a `.env` file in the root directory to manage your environment-specific variables.

```bash
PORT=8080
ENV=development
DB_DSN=postgres://postgres:yourpassword@localhost/go_auth?sslmode=disable

// The duration for which a token is valid (e.g., 24 hours)
TOKEN_EXPIRY=24h
```
---

### Step 1: Database Setup and Access the PostgreSQL CLI
First, you need to enter the Postgres environment. Open your terminal and run:

```bash
# Enter the Postgres CLI as the default superuser
sudo -u postgres psql
```
#### Step 2: Setup database username and password
```sql
CREATE USER postgres WITH PASSWORD 'yourpassword';
```
#### Step 3: Create and Connect to the Database
Once you see the postgres=# prompt, run these commands to prepare the "house" for your tables:

#### Step 1: Create the database
```sql
CREATE DATABASE go_auth;
```

#### Step 2: Connect to the new database
```sql
\c go_auth
```
The prompt should change from postgres=# to go_auth=#.

#### Step 4: Run the Schema (Copy & Paste)
Now, simply paste this entire block into the terminal and hit Enter.

```sql
-- Create Users Table
CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    username text NOT NULL,
    email text UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    role text NOT NULL DEFAULT 'user'
);

-- Create Tokens Table
CREATE TABLE IF NOT EXISTS tokens (
    hash bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL,
    scope text NOT NULL
);
```
#### Step 5: Verify the Tables
To make sure the tables were created successfully, use the "describe" command:

##### List all tables in the current database
```sql
\dt
```
Expected Output:
```
         List of relations
 Schema |  Name   | Type  |  Owner   
--------+---------+-------+----------
 public | tokens  | table | postgres
 public | users   | table | postgres
```
#### Step 6: Exit the CLI
When you are finished, type:

```bash
\q
```

#### Pro-Tip: Running the file directly (The "Automated" Way)
If you saved your SQL code in a file named schema.sql, you don't even need to open the prompt. you can just run this one-liner from your project root:

```bash
psql -U postgres -d go_auth -f path/to/schema.sql
```
---
### Run the Application

```bash
go run ./cmd/api
```
---
## API Testing (CURL)

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

## 👤 Author
**[Emmanuel Itighise](https://github.com/Emeditweb)**  
Microbiologist | Data Analyst | AI-Native Software Engineering Fellow