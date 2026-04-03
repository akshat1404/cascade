<script lang="ts">
    import { onMount } from "svelte";
    import * as mammoth from "mammoth";
    import * as pdfjsLib from "pdfjs-dist";

    pdfjsLib.GlobalWorkerOptions.workerSrc = `https://cdnjs.cloudflare.com/ajax/libs/pdf.js/${pdfjsLib.version}/pdf.worker.min.mjs`;

    interface Props {
        onClose: () => void;
        onCreate: (title: string, content?: string) => void;
    }

    let { onClose, onCreate }: Props = $props();

    let title = $state("");
    let inputEl: HTMLInputElement;
    let fileInputEl: HTMLInputElement;
    let isProcessing = $state(false);

    function handleCreate() {
        const trimmed = title.trim();
        if (!trimmed) return;
        onCreate(trimmed);
    }

    function handleKeydown(e: KeyboardEvent) {
        if (e.key === "Enter") handleCreate();
        if (e.key === "Escape") onClose();
    }

    function triggerFileSelect() {
        fileInputEl?.click();
    }

    async function handleFileSelect(event: Event) {
        const target = event.target as HTMLInputElement;
        const file = target.files?.[0];
        if (!file) return;

        isProcessing = true;

        // Auto-fill title from filename if empty
        if (!title.trim()) {
            title = file.name.replace(/\.[^/.]+$/, "");
        }

        try {
            const arrayBuffer = await file.arrayBuffer();
            let contentStr = "";

            if (file.name.toLowerCase().endsWith(".docx")) {
                const result = await mammoth.convertToHtml({ arrayBuffer });
                contentStr = result.value;
            } else if (file.name.toLowerCase().endsWith(".pdf")) {
                const loadingTask = pdfjsLib.getDocument({ data: arrayBuffer });
                const pdf = await loadingTask.promise;

                contentStr = "<div>";
                for (let i = 1; i <= pdf.numPages; i++) {
                    const page = await pdf.getPage(i);
                    const textContent = await page.getTextContent();
                    // Basic paragraphing logic based on string items
                    const textItems = textContent.items
                        .map((item: any) => item.str)
                        .join(" ");
                    contentStr += `<p>${textItems}</p>`;
                }
                contentStr += "</div>";
            } else {
                alert("Please select a valid .docx or .pdf file");
                isProcessing = false;
                return;
            }

            // Fire onCreate immediately with the compiled content
            onCreate(title.trim() || "Imported Document", contentStr);
        } catch (err) {
            console.error("Error processing file:", err);
            alert("Error importing the file structure");
        } finally {
            isProcessing = false;
        }
    }

    onMount(() => inputEl?.focus());
</script>

<div class="backdrop" role="presentation" onclick={onClose}></div>

<div class="modal" role="dialog" aria-modal="true" aria-label="Create document">
    <button class="close-btn" onclick={onClose} aria-label="Close">✕</button>

    <div class="modal-icon">✦</div>
    <h2 class="modal-title">New Document</h2>
    <p class="modal-subtitle">Enter a title to get started</p>

    <div class="input-wrapper">
        <input
            bind:this={inputEl}
            bind:value={title}
            onkeydown={handleKeydown}
            type="text"
            id="doc-title-input"
            class="doc-input"
            placeholder="Untitled document..."
            maxlength="120"
            autocomplete="off"
        />
        <div class="input-glow"></div>
    </div>

    <button
        class="create-btn"
        onclick={handleCreate}
        disabled={!title.trim() || isProcessing}
    >
        <span class="btn-sparkle">✦</span>
        {isProcessing ? "Processing..." : "Create"}
    </button>

    <div class="import-section">
        <div class="import-divider"><span>OR</span></div>
        <p class="import-text">Import an existing document</p>
        <button
            class="import-btn"
            onclick={triggerFileSelect}
            disabled={isProcessing}
        >
            Import Word / PDF
        </button>
        <input
            type="file"
            accept=".docx,.pdf"
            bind:this={fileInputEl}
            onchange={handleFileSelect}
            style="display:none;"
        />
    </div>
</div>

<style>
    .backdrop {
        position: fixed;
        inset: 0;
        z-index: 1000;
        background: rgba(10, 0, 20, 0.45);
        backdrop-filter: blur(6px);
        -webkit-backdrop-filter: blur(6px);
        animation: fade-in 0.18s ease;
    }

    .modal {
        position: fixed;
        inset: 0;
        margin: auto;
        z-index: 1001;
        width: min(92vw, 420px);
        height: fit-content;

        background: rgba(255, 255, 255, 0.82);
        backdrop-filter: blur(28px);
        -webkit-backdrop-filter: blur(28px);
        border: 1px solid rgba(236, 72, 153, 0.22);
        border-radius: 22px;
        padding: 2.2rem 2rem 2rem;

        box-shadow:
            0 0 0 1px rgba(255, 255, 255, 0.6) inset,
            0 8px 40px rgba(236, 72, 153, 0.14),
            0 2px 8px rgba(0, 0, 0, 0.06);

        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 0.65rem;

        animation: modal-in 0.26s cubic-bezier(0.34, 1.4, 0.64, 1);
    }

    .close-btn {
        position: absolute;
        top: 1rem;
        right: 1rem;
        width: 30px;
        height: 30px;
        border-radius: 50%;
        border: none;
        background: rgba(236, 72, 153, 0.08);
        color: #be185d;
        font-size: 0.8rem;
        display: flex;
        align-items: center;
        justify-content: center;
        transition:
            background 0.15s ease,
            transform 0.15s ease;
        font-family: inherit;
    }
    .close-btn:hover {
        background: rgba(236, 72, 153, 0.16);
        transform: scale(1.1);
    }

    .modal-icon {
        font-size: 1.6rem;
        background: linear-gradient(135deg, #f9a8d4, #a855f7);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
        margin-bottom: 0.1rem;
        filter: drop-shadow(0 0 8px rgba(236, 72, 153, 0.4));
    }

    .modal-title {
        font-size: 1.35rem;
        font-weight: 700;
        color: #1a0a12;
        letter-spacing: -0.4px;
        margin: 0;
    }
    .modal-subtitle {
        font-size: 0.82rem;
        color: #9d6b8a;
        margin: 0 0 0.5rem;
    }

    .input-wrapper {
        position: relative;
        width: 100%;
    }
    .doc-input {
        width: 100%;
        padding: 0.75rem 1rem;
        border-radius: 12px;
        border: 1.5px solid rgba(236, 72, 153, 0.25);
        background: rgba(255, 255, 255, 0.9);
        font-family: inherit;
        font-size: 0.95rem;
        color: #1a0a12;
        outline: none;
        transition:
            border-color 0.18s ease,
            box-shadow 0.18s ease;
        position: relative;
        z-index: 1;
    }
    .doc-input::placeholder {
        color: #c4a0b8;
    }
    .doc-input:focus {
        border-color: rgba(236, 72, 153, 0.6);
        box-shadow:
            0 0 0 3px rgba(236, 72, 153, 0.1),
            0 0 14px rgba(236, 72, 153, 0.12);
    }

    .create-btn {
        margin-top: 0.4rem;
        width: 100%;
        padding: 0.72rem 1rem;
        border-radius: 12px;
        border: none;
        background: linear-gradient(135deg, #ec4899, #a855f7);
        color: #fff;
        font-family: inherit;
        font-size: 0.95rem;
        font-weight: 600;
        letter-spacing: 0.01em;
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 0.45rem;
        box-shadow:
            0 4px 18px rgba(236, 72, 153, 0.38),
            0 1px 4px rgba(0, 0, 0, 0.08);
        transition:
            opacity 0.15s ease,
            transform 0.15s ease,
            box-shadow 0.15s ease;
    }
    .create-btn:hover:not(:disabled) {
        transform: translateY(-1px);
        box-shadow:
            0 6px 26px rgba(236, 72, 153, 0.52),
            0 2px 6px rgba(0, 0, 0, 0.1);
    }
    .create-btn:active:not(:disabled) {
        transform: translateY(0);
    }
    .create-btn:disabled {
        opacity: 0.45;
    }
    .btn-sparkle {
        font-size: 0.8rem;
        opacity: 0.85;
    }

    /* ── Import section ── */
    .import-section {
        width: 100%;
        display: flex;
        flex-direction: column;
        align-items: center;
        margin-top: 0.2rem;
    }

    .import-divider {
        display: flex;
        align-items: center;
        text-align: center;
        width: 100%;
        margin: 0.5rem 0 1rem;
    }
    .import-divider::before,
    .import-divider::after {
        content: "";
        flex: 1;
        border-bottom: 1px solid rgba(236, 72, 153, 0.15);
    }
    .import-divider span {
        padding: 0 10px;
        color: #c4a0b8;
        font-size: 0.75rem;
        font-weight: 600;
        letter-spacing: 0.05em;
    }

    .import-text {
        font-size: 0.85rem;
        color: #9d6b8a;
        margin: 0 0 0.6rem;
    }

    .import-btn {
        width: 100%;
        padding: 0.7rem 1rem;
        border-radius: 12px;
        border: 1px dashed rgba(236, 72, 153, 0.4);
        background: rgba(253, 242, 248, 0.5);
        color: #be185d;
        font-family: inherit;
        font-size: 0.9rem;
        font-weight: 600;
        cursor: pointer;
        transition: all 0.15s ease;
    }
    .import-btn:hover:not(:disabled) {
        background: rgba(253, 242, 248, 1);
        border-color: #ec4899;
        transform: translateY(-1px);
        box-shadow: 0 4px 12px rgba(236, 72, 153, 0.1);
    }
    .import-btn:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }

    @keyframes fade-in {
        from {
            opacity: 0;
        }
        to {
            opacity: 1;
        }
    }

    @keyframes modal-in {
        from {
            opacity: 0;
            transform: scale(0.88) translateY(12px);
        }
        to {
            opacity: 1;
            transform: scale(1) translateY(0);
        }
    }
</style>
