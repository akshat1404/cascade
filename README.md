# Cascade

> A rich text editor — built with SvelteKit, Go, Supabase, and MongoDB.
> <a href="https://collaborent.vercel.app" target="_blank">Access Link</a> | <a href="https://cascade-docs-seven.vercel.app/" target="_blank">Interactive Docs</a>

---

**If you find this project useful or find it interesting enough to clone and make it your own, consider starring the repo — it genuinely helps!**

---

## Overview

Cascade is a real-time collaborative document editing platform. Users authenticate securely via Google OAuth (powered by Supabase), and their sessions are validated server-side by a Go backend before any data is persisted to MongoDB.

The project is intentionally kept full-stack and minimal at its core — authentication and user management are rock-solid before document features are layered on top.

---

## Architecture

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

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | [SvelteKit](https://kit.svelte.dev/) (Svelte 5, TypeScript, Vite) |
| Backend | [Go](https://go.dev/) 1.22 — standard `net/http` |
| Auth | [Supabase](https://supabase.com/) + Google OAuth 2.0 |
| Database | [MongoDB](https://www.mongodb.com/) (via `mongo-driver/v2`) |
| Editor | [TipTap](https://tiptap.dev/) (ProseMirror based rich-text editor) |
| AI Inference | [Groq](https://groq.com/) API (llama-3.1-8b-instant model) |
| Styling | Vanilla CSS — glassmorphism, pinkish theme |

---

## Project Structure

```
cascade/
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

## Authentication Flow

1. User clicks **"Continue with Google"** on the landing page.
2. Supabase handles the OAuth redirect and issues a signed **JWT**.
3. The SvelteKit dashboard calls `GET /auth/callback` on the Go backend, passing the JWT as a `Bearer` token.
4. The Go server fetches Supabase's **JWKS** and cryptographically verifies the token.
5. The verified user is **upserted** into MongoDB (creates on first login, updates `lastLoginAt` on subsequent logins).

---

## Getting Started

### Prerequisites

- **Go** >= 1.22
- **Node.js** >= 18
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
MONGODB_DB=cascade
GROQ_API_KEY=<your-groq-api-key>
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

## Backend API

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Health check — returns `{ "status": "ok" }` |
| `POST` | `/auth/callback` | Validates JWT and upserts user into MongoDB |
| `GET` | `/documents` | Lists all documents belonging to the user |
| `POST` | `/documents/create` | Creates a new document |
| `GET`, `PUT`, `DELETE` | `/documents/{id}` | Fetches, updates, or deletes a specific document |
| `POST` | `/ai/process` | Processes text using the Groq AI API (`llama-3.1`) |

All endpoints use CORS middleware permitting `http://localhost:5173` during development.

---

## Editor & AI Features

The core document editor runs on **TipTap**. It includes standard rich text formatting and introduces native AI intelligence directly in the editor.

When text is selected, a floating AI Bubble appears above the formatting tools, allowing the user to seamlessly use Groq's fast inference to edit content:

- ✦ **Fix grammar**: Automatically fixes grammatical errors.
- ✦ **Translate to Hindi**: Quickly translates the selection.
- ✦ **Make a table**: Intelligently formats selected text into a markdown table.
- ✦ **Summarise**: Condenses the selected text into 2-3 sentences.
- ✦ **Ask AI**: Opens a modal where the user can type free-form instructions (e.g. *"This is code, format it properly"*), which instructs the AI exactly how to rewrite the selected text.

The Go backend safely takes the user request, verifies the user token, securely constructs the prompts, and proxies the request to the Groq API. Content returned by Groq is rendered from Markdown into native ProseMirror HTML directly in the editor.

Cascade also natively supports **Document Imports**. You can upload `.docx` or `.pdf` files from the Dashboard, which are locally parsed into text/HTML and converted into a new Cascade document.

---

## Behind the Build — Articles on Medium

I document the engineering decisions and concepts behind this project on Medium.

[@akshatmtiwari on Medium](https://medium.com/@akshatmtiwari)

| # | Article |
|---|---|
| 1 | [**Google OAuth Explained** — From Cookies to "Continue with Google"](https://medium.com/@akshatmtiwari/google-oauth-explained-from-cookies-to-continue-with-google-3d88141590f3) |
| 2 | [**Building Google OAuth** with Supabase, SvelteKit and Go](https://medium.com/@akshatmtiwari/building-google-oauth-with-supabase-sveltekit-and-go-1a0402de408f) |
| 3 | [**OAuth vs SSO**](https://medium.com/@akshatmtiwari/i-reverse-engineered-my-gmail-login-and-found-something-unexpected-6d3d326f7cc6) |
| 4 | [**The Headache of Maintaining Documentations in Production**](https://medium.com/@akshatmtiwari/the-headache-of-maintaining-documentations-in-production-3552d853a923) |

---

## Roadmap

- [x] Google OAuth authentication (Supabase)
- [x] JWT validation in Go backend
- [x] User upsert into MongoDB
- [x] Dashboard UI — component-based (Navbar, DocumentsPage, LoadingScreen)
- [x] Document creation, persistence, and deletion
- [x] Rich Text Editor implementation (TipTap)
- [x] Word Document (`.docx`) and PDF Import functionality
- [x] Floating AI Assistant (Groq `llama-3.1-8b-instant`) — Grammar, Translate, Summarize, Custom Prompts
- [ ] Real-time collaborative editing using WebSockets / Yjs
- [ ] Document sharing & permissions

---

## License

MIT — feel free to use this as a reference for your own full-stack SvelteKit + Go projects.
