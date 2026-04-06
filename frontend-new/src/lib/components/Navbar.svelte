<script lang="ts">
    interface Props {
        onSignOut: () => void;
        onCreateDoc?: () => void;
    }

    let { onSignOut, onCreateDoc }: Props = $props();
</script>

<nav class="navbar">
    <span class="navbar-brand">Cascade</span>
    <div class="navbar-actions">
        <button class="btn-create" id="create-doc-btn" onclick={onCreateDoc}>
            <span class="btn-icon">+</span> Create
        </button>
        <button class="btn-signout" id="signout-btn" onclick={onSignOut}>
            Sign out
        </button>
    </div>
</nav>

<style>
    .navbar {
        position: sticky;
        top: 0;
        z-index: 100;
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0 2rem;
        height: 64px;
        background: rgba(255, 255, 255, 0.75);
        backdrop-filter: blur(18px);
        -webkit-backdrop-filter: blur(18px);
        border-bottom: 1px solid rgba(236, 72, 153, 0.2);
        box-shadow: 0 2px 24px rgba(236, 72, 153, 0.08);
    }

    .navbar-brand {
        display: inline-block;
        font-size: 1.35rem;
        font-weight: 700;
        letter-spacing: -0.5px;
        background: linear-gradient(135deg, #f9a8d4, #e879f9);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
    }

    .navbar-actions {
        display: flex;
        align-items: center;
        gap: 0.75rem;
    }

    .btn-create {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        padding: 0.45rem 1.1rem;
        border-radius: 999px;
        border: none;
        cursor: pointer;
        font-family: inherit;
        font-size: 0.875rem;
        font-weight: 600;
        background: linear-gradient(135deg, #ec4899, #a855f7);
        color: #fff;
        box-shadow: 0 0 18px rgba(236, 72, 153, 0.35);
        transition:
            transform 0.18s ease,
            box-shadow 0.18s ease;
    }

    .btn-create:hover {
        transform: translateY(-1px) scale(1.03);
        box-shadow: 0 0 28px rgba(236, 72, 153, 0.55);
    }

    .btn-create:active {
        transform: translateY(0) scale(0.98);
    }

    .btn-icon {
        font-size: 1rem;
        line-height: 1;
    }

    .btn-signout {
        padding: 0.45rem 1.1rem;
        border-radius: 999px;
        border: 1px solid rgba(236, 72, 153, 0.35);
        cursor: pointer;
        font-family: inherit;
        font-size: 0.875rem;
        font-weight: 500;
        background: rgba(236, 72, 153, 0.06);
        color: #be185d;
        backdrop-filter: blur(8px);
        transition:
            background 0.18s ease,
            border-color 0.18s ease,
            transform 0.18s ease;
    }

    .btn-signout:hover {
        background: rgba(236, 72, 153, 0.12);
        border-color: rgba(236, 72, 153, 0.6);
        transform: translateY(-1px);
    }

    .btn-signout:active {
        transform: translateY(0);
    }
</style>
