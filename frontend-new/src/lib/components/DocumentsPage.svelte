<script lang="ts">
    type Document = {
        id: string;
        title: string;
        updated_at: string;
        created_at: string;
    };

    let { documents = [], loading = false }: { documents: Document[]; loading: boolean } = $props();

    function formatDate(iso: string): string {
        if (!iso) return "";
        const d = new Date(iso);
        return d.toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
            year: "numeric",
        });
    }

    function excerpt(doc: Document): string {
        // No content preview available without loading full doc — show title as fallback
        return "Open to view content";
    }
</script>

<div class="documents-page">
    {#if loading}
        <div class="loading-grid">
            {#each [1, 2, 3] as _}
                <div class="skeleton-card"></div>
            {/each}
        </div>
    {:else if documents.length === 0}
        <div class="empty-state">
            <div class="empty-icon">
                <svg
                    width="64"
                    height="64"
                    viewBox="0 0 64 64"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                >
                    <rect
                        x="12"
                        y="8"
                        width="36"
                        height="46"
                        rx="4"
                        stroke="url(#grad1)"
                        stroke-width="2.5"
                        fill="none"
                    />
                    <line x1="20" y1="22" x2="44" y2="22" stroke="url(#grad1)" stroke-width="2" stroke-linecap="round" />
                    <line x1="20" y1="30" x2="44" y2="30" stroke="url(#grad1)" stroke-width="2" stroke-linecap="round" />
                    <line x1="20" y1="38" x2="34" y2="38" stroke="url(#grad1)" stroke-width="2" stroke-linecap="round" />
                    <defs>
                        <linearGradient id="grad1" x1="12" y1="8" x2="48" y2="54" gradientUnits="userSpaceOnUse">
                            <stop stop-color="#f472b6" />
                            <stop offset="1" stop-color="#a855f7" />
                        </linearGradient>
                    </defs>
                </svg>
            </div>
            <p class="empty-title">You don't have any documents</p>
            <p class="empty-subtitle">
                Click <strong>Create +</strong> to get started
            </p>
        </div>
    {:else}
        <div class="section-header">
            <h2 class="section-title">Your Documents</h2>
            <span class="doc-count">{documents.length} document{documents.length === 1 ? "" : "s"}</span>
        </div>
        <div class="docs-grid">
            {#each documents as doc (doc.id)}
                <a href="/document/{doc.id}" class="doc-card">
                    <div class="card-accent"></div>
                    <div class="card-body">
                        <div class="doc-icon">
                            <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
                                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="url(#cardGrad)" stroke-width="1.8" stroke-linejoin="round" fill="none"/>
                                <polyline points="14 2 14 8 20 8" stroke="url(#cardGrad)" stroke-width="1.8" stroke-linejoin="round"/>
                                <line x1="8" y1="13" x2="16" y2="13" stroke="url(#cardGrad)" stroke-width="1.5" stroke-linecap="round"/>
                                <line x1="8" y1="17" x2="13" y2="17" stroke="url(#cardGrad)" stroke-width="1.5" stroke-linecap="round"/>
                                <defs>
                                    <linearGradient id="cardGrad" x1="4" y1="2" x2="22" y2="24" gradientUnits="userSpaceOnUse">
                                        <stop stop-color="#f472b6"/>
                                        <stop offset="1" stop-color="#a855f7"/>
                                    </linearGradient>
                                </defs>
                            </svg>
                        </div>
                        <h3 class="doc-title">{doc.title || "Untitled"}</h3>
                        <p class="doc-excerpt">{excerpt(doc)}</p>
                    </div>
                    <div class="card-footer">
                        <span class="doc-date">Edited {formatDate(doc.updated_at)}</span>
                        <span class="open-arrow">→</span>
                    </div>
                </a>
            {/each}
        </div>
    {/if}
</div>

<style>
    .documents-page {
        flex: 1;
        padding: 2.5rem 2rem;
        max-width: 1100px;
        width: 100%;
        margin: 0 auto;
    }

    /* ── Section header ── */
    .section-header {
        display: flex;
        align-items: baseline;
        gap: 0.75rem;
        margin-bottom: 1.75rem;
    }

    .section-title {
        font-size: 1.1rem;
        font-weight: 700;
        color: #1a0a12;
        letter-spacing: 0.01em;
    }

    .doc-count {
        font-size: 0.8rem;
        font-weight: 500;
        color: #b0a0bb;
    }

    /* ── Document grid ── */
    .docs-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
        gap: 1.25rem;
    }

    /* ── Document card ── */
    .doc-card {
        display: flex;
        flex-direction: column;
        border-radius: 16px;
        background: rgba(255, 255, 255, 0.85);
        border: 1px solid rgba(236, 72, 153, 0.12);
        box-shadow: 0 2px 12px rgba(168, 85, 247, 0.06);
        text-decoration: none;
        color: inherit;
        overflow: hidden;
        transition:
            transform 0.18s ease,
            box-shadow 0.18s ease,
            border-color 0.18s ease;
        position: relative;
        animation: fadeUp 0.35s ease both;
        backdrop-filter: blur(8px);
        -webkit-backdrop-filter: blur(8px);
    }

    .doc-card:hover {
        transform: translateY(-3px);
        box-shadow:
            0 8px 28px rgba(168, 85, 247, 0.14),
            0 2px 8px rgba(244, 114, 182, 0.1);
        border-color: rgba(168, 85, 247, 0.3);
    }

    /* Pink top accent bar */
    .card-accent {
        height: 3px;
        background: linear-gradient(90deg, #f472b4 0%, #a855f7 100%);
        opacity: 0;
        transition: opacity 0.18s ease;
    }
    .doc-card:hover .card-accent {
        opacity: 1;
    }

    .card-body {
        padding: 1.25rem 1.25rem 0.75rem;
        flex: 1;
    }

    .doc-icon {
        margin-bottom: 0.75rem;
        display: flex;
    }

    .doc-title {
        font-size: 0.95rem;
        font-weight: 600;
        color: #1a0a12;
        margin: 0 0 0.35rem;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        line-height: 1.3;
    }

    .doc-excerpt {
        font-size: 0.78rem;
        color: #9d84b0;
        line-height: 1.5;
        display: -webkit-box;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }

    .card-footer {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.6rem 1.25rem 1rem;
        border-top: 1px solid rgba(236, 72, 153, 0.07);
        margin-top: 0.5rem;
    }

    .doc-date {
        font-size: 0.73rem;
        color: #b8a8c8;
        font-weight: 400;
    }

    .open-arrow {
        font-size: 0.85rem;
        color: #d8b4fe;
        transition: transform 0.15s ease, color 0.15s ease;
    }
    .doc-card:hover .open-arrow {
        color: #a855f7;
        transform: translateX(3px);
    }

    /* ── Skeleton loading ── */
    .loading-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
        gap: 1.25rem;
    }

    .skeleton-card {
        height: 160px;
        border-radius: 16px;
        background: linear-gradient(
            90deg,
            rgba(244, 114, 182, 0.06) 0%,
            rgba(168, 85, 247, 0.1) 50%,
            rgba(244, 114, 182, 0.06) 100%
        );
        background-size: 200% 100%;
        animation: shimmer 1.5s infinite;
    }

    /* ── Empty state ── */
    .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 1rem;
        padding: 4rem 2rem;
        margin-top: 3rem;
        border-radius: 20px;
        background: rgba(255, 255, 255, 0.72);
        border: 1px dashed rgba(236, 72, 153, 0.3);
        backdrop-filter: blur(12px);
        -webkit-backdrop-filter: blur(12px);
        box-shadow: 0 4px 32px rgba(236, 72, 153, 0.07);
        text-align: center;
        animation: fadeUp 0.5s ease both;
    }

    .empty-icon {
        opacity: 0.75;
        animation: float 3.5s ease-in-out infinite;
    }

    .empty-title {
        font-size: 1.15rem;
        font-weight: 600;
        color: #9d174d;
    }

    .empty-subtitle {
        font-size: 0.875rem;
        color: rgba(157, 23, 77, 0.5);
    }

    .empty-subtitle strong {
        color: #be185d;
        font-weight: 600;
    }

    /* ── Animations ── */
    @keyframes float {
        0%, 100% { transform: translateY(0); }
        50%       { transform: translateY(-8px); }
    }

    @keyframes fadeUp {
        from { opacity: 0; transform: translateY(16px); }
        to   { opacity: 1; transform: translateY(0); }
    }

    @keyframes shimmer {
        0%   { background-position: -200% 0; }
        100% { background-position:  200% 0; }
    }
</style>
