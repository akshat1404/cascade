<div class="documents-page">
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
                <line
                    x1="20"
                    y1="22"
                    x2="44"
                    y2="22"
                    stroke="url(#grad1)"
                    stroke-width="2"
                    stroke-linecap="round"
                />
                <line
                    x1="20"
                    y1="30"
                    x2="44"
                    y2="30"
                    stroke="url(#grad1)"
                    stroke-width="2"
                    stroke-linecap="round"
                />
                <line
                    x1="20"
                    y1="38"
                    x2="34"
                    y2="38"
                    stroke="url(#grad1)"
                    stroke-width="2"
                    stroke-linecap="round"
                />
                <defs>
                    <linearGradient
                        id="grad1"
                        x1="12"
                        y1="8"
                        x2="48"
                        y2="54"
                        gradientUnits="userSpaceOnUse"
                    >
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
</div>

<style>
    .documents-page {
        flex: 1;
        padding: 2.5rem 2rem;
        max-width: 1100px;
        width: 100%;
        margin: 0 auto;
    }

    .section-header {
        margin-bottom: 2rem;
    }

    .section-title {
        font-size: 1.1rem;
        font-weight: 600;
        color: #f5d0ea;
        letter-spacing: 0.02em;
    }

    .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 1rem;
        padding: 4rem 2rem;
        margin-top: 3rem;
        border-radius: 20px;
        background: rgba(255, 255, 255, 0.03);
        border: 1px dashed rgba(244, 114, 182, 0.25);
        backdrop-filter: blur(12px);
        -webkit-backdrop-filter: blur(12px);
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
        color: #f0abcf;
    }

    .empty-subtitle {
        font-size: 0.875rem;
        color: rgba(240, 171, 207, 0.55);
    }

    .empty-subtitle strong {
        color: rgba(240, 171, 207, 0.85);
        font-weight: 600;
    }

    @keyframes float {
        0%,
        100% {
            transform: translateY(0);
        }
        50% {
            transform: translateY(-8px);
        }
    }

    @keyframes fadeUp {
        from {
            opacity: 0;
            transform: translateY(16px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
</style>
