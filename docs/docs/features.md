---
sidebar_position: 3
---

# Editor & AI Features

Cascade provides a sophisticated text editing experience utilizing **TipTap** (built on top of ProseMirror) and deep API integrations with **Groq** for fast AI processing.

## The TipTap Editor
The core editor offers a rich text interface capable of handling multiple formats. It natively supports **Document Imports**, allowing users to upload `.docx` or `.pdf` files which are securely parsed into HTML and fed into the ProseMirror state.

## Floating AI Assistant
Cascade integrates a floating contextual AI helper. When text is highlighted in the editor, users are presented with quick-action AI tools alongside standard formatting tools (bold, italic, etc.).

### Capabilities:
- ✦ **Fix grammar**: Automatically catches and fixes typos or grammatical mismatches.
- ✦ **Translate to Hindi**: Uses the LLM to provide a high-quality translation of the selected text block.
- ✦ **Make a table**: Intelligently restructures selected prose data into markdown tables.
- ✦ **Summarize**: Condenses the selected text into 2-3 precise sentences.
- ✦ **Ask AI**: A custom prompt modal where the user can define exact instructions for the AI engine (e.g., *"Rewrite this in a more professional tone"*).

### How it Works Securely
All AI prompts are routed through the Go backend rather than executed client-side. The Go server:
1. Validates the user's JWT.
2. Formats the system prompt constraints depending on the requested AI action.
3. Rapidly proxies the data to the Groq `llama-3.1-8b-instant` endpoint and streams/returns the response.
