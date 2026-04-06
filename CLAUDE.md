# Cascade — Project Context

## Stack
- Frontend: SvelteKit with Svelte 5 (use $state, $derived runes — NOT writable stores)
- Backend: Go 1.22 on localhost:8080
- Database: MongoDB (via Go backend)
- Auth: Supabase Auth — session token at `session.access_token`
- Rich text: TipTap via @tiptap/core (no official Svelte wrapper — use onMount)

## Project Structure
src/
├── routes/
│   ├── dashboard/+page.svelte      ← document list, create doc
│   └── document/[id]/+page.svelte  ← rich text editor (YOU ARE BUILDING THIS)
├── lib/
│   ├── supabase.ts                 ← supabase client
│   └── components/
│       ├── Navbar.svelte
│       ├── DocumentsPage.svelte
│       └── CreateDocModal.svelte

## API Endpoints (Go backend)
POST   /documents/create    → creates doc, returns { id, title, content, owner_id }
GET    /documents/:id       → returns full document
PUT    /documents/:id       → saves document (body: { title, content })

All endpoints require: Authorization: Bearer <token>

## Document Data Shape
{
  id: string            // MongoDB ObjectID as hex string
  title: string
  content: object       // ProseMirror JSON — pass directly to TipTap as content
  owner_id: string      // Supabase user ID
  created_at: string
  updated_at: string
}

## Auth Pattern
Always get token like this — never use user.access_token:
  const { data: { session } } = await supabase.auth.getSession()
  session.access_token

## TipTap Setup Pattern (Svelte)
- Import from @tiptap/core, NOT @tiptap/react
- Initialize inside onMount
- Bind to a div with bind:this
- Destroy in onDestroy
- Use editor.getJSON() to get ProseMirror JSON for saving

## Autosave
- Debounce 1500ms after last keystroke
- PATCH/PUT to /documents/:id with full content JSON
- Show save status: "saving..." → "saved"

## Styling Conventions
- Font: Inter
- No Tailwind — plain CSS with scoped <style> blocks
- Color palette: pink (#f472b4) and purple (#a855f7) accents on white