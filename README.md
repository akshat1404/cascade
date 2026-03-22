# Collaborent

> A collaborative document editor — built with SvelteKit, Go, Supabase, and MongoDB.

---

## ✨ Overview

Collaborent is a real-time collaborative document editing platform. Users authenticate securely via Google OAuth (powered by Supabase), and their sessions are validated server-side by a Go backend before any data is persisted to MongoDB.

The project is intentionally kept full-stack and minimal at its core — authentication and user management are rock-solid before document features are layered on top.

---

## 🏗️ Architecture

```
┌──────────────────────┐        ┌────────────────────────┐        ┌─────────────┐
│   SvelteKit Frontend │ ──────▶│     Go REST Backend    │ ──────▶│   MongoDB   │
│  (Vite · TypeScript) │        │  JWT verified via JWKS │        │  (Atlas)    │
└──────────────────────┘        └────────────────────────┘        └─────────────┘
           │                                  ▲
           │       Supabase Auth              │
           └──────────────────────────────────┘
                  Google OAuth 2.0
```

- The **frontend** handles login and session management via `@supabase/supabase-js`.
- On sign-in, the frontend sends the Supabase JWT to the **Go backend**.
- The backend validates the token by fetching Supabase's JWKS endpoint, then upserts the user record into **MongoDB**.

---

## 🛠️ Tech Stack

| Layer | Technology |
|---|---|
| Frontend | [SvelteKit](https://kit.svelte.dev/) (Svelte 5, TypeScript, Vite) |
| Backend | [Go](https://go.dev/) 1.22 — standard `net/http` |
| Auth | [Supabase](https://supabase.com/) + Google OAuth 2.0 |
| JWT Validation | `golang-jwt/jwt` + `MicahParks/keyfunc` (JWKS) |
| Database | [MongoDB](https://www.mongodb.com/) (via `mongo-driver/v2`) |
| Styling | Vanilla CSS — glassmorphism, pinkish theme |

---

## 📁 Project Structure

```
collaborent/
├── backend/
│   ├── main.go          # Go HTTP server — auth callback, health check
│   ├── go.mod
│   └── .env             # SUPABASE_URL, MONGODB_URI, MONGODB_DB
│
└── frontend-new/
    ├── src/
    │   ├── lib/
    │   │   ├── supabase.ts              # Supabase client initialisation
    │   │   └── components/
    │   │       ├── Navbar.svelte        # Top nav — brand, Create+, Sign out
    │   │       ├── DocumentsPage.svelte # Document list / empty state
    │   │       └── LoadingScreen.svelte # Spinner / redirect fallback
    │   └── routes/
    │       ├── +page.svelte             # Landing / login page
    │       └── dashboard/
    │           └── +page.svelte         # Auth-gated dashboard (orchestrator)
    ├── package.json
    └── .env             # PUBLIC_SUPABASE_URL, PUBLIC_SUPABASE_ANON_KEY
```

---

## 🔐 Authentication Flow

1. User clicks **"Continue with Google"** on the landing page.
2. Supabase handles the OAuth redirect and issues a signed **JWT**.
3. The SvelteKit dashboard calls `GET /auth/callback` on the Go backend, passing the JWT as a `Bearer` token.
4. The Go server fetches Supabase's **JWKS** and cryptographically verifies the token.
5. The verified user is **upserted** into MongoDB (creates on first login, updates `lastLoginAt` on subsequent logins).

---

## 🚀 Getting Started

### Prerequisites

- **Go** ≥ 1.22
- **Node.js** ≥ 18
- A **Supabase** project with Google OAuth enabled
- A **MongoDB** cluster (Atlas free tier works fine)

---

### Backend

```bash
cd backend

# Copy and fill in your environment variables
cp .env.example .env

# Run the server (listens on :8080)
go run main.go
```

**`.env` keys required:**

```env
SUPABASE_URL=https://<your-project>.supabase.co
MONGODB_URI=mongodb+srv://<user>:<pass>@cluster.mongodb.net/
MONGODB_DB=collaborent
```

---

### Frontend

```bash
cd frontend-new

# Install dependencies
npm install

# Copy and fill in your environment variables
cp .env.example .env

# Start the dev server (http://localhost:5173)
npm run dev
```

**`.env` keys required:**

```env
PUBLIC_SUPABASE_URL=https://<your-project>.supabase.co
PUBLIC_SUPABASE_ANON_KEY=<your-anon-key>
```

---

## 📡 Backend API

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Health check — returns `{ "status": "ok" }` |
| `POST` | `/auth/callback` | Validates JWT and upserts user into MongoDB |

All endpoints use CORS middleware permitting `http://localhost:5173` during development.

---

## ✍️ Behind the Build — Articles on Medium

I document the engineering decisions and concepts behind this project on Medium.

**[@akshatmtiwari on Medium](https://medium.com/@akshatmtiwari)**

| # | Article |
|---|---|
| 1 | [**Google OAuth Explained** — From Cookies to "Continue with Google"](https://medium.com/@akshatmtiwari/google-oauth-explained-from-cookies-to-continue-with-google-3d88141590f3) |
| 2 | [**Building Google OAuth** with Supabase, SvelteKit and Go](https://medium.com/@akshatmtiwari/building-google-oauth-with-supabase-sveltekit-and-go-1a0402de408f) |

---

## 🗺️ Roadmap

- [x] Google OAuth authentication (Supabase)
- [x] JWT validation in Go backend
- [x] User upsert into MongoDB
- [x] Dashboard UI — component-based (Navbar, DocumentsPage, LoadingScreen)
- [ ] Document creation & persistence
- [ ] Real-time collaborative editing
- [ ] Document sharing & permissions

---

## 📄 License

MIT — feel free to use this as a reference for your own full-stack SvelteKit + Go projects.
