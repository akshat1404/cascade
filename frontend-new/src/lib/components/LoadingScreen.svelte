<script lang="ts">
    interface Props {
        message?: string;
    }
    let { message = "" }: Props = $props();
</script>

<div class="loading-screen">
    {#if message}
        <p class="redirect-text">{message}</p>
    {:else}
        <div class="spinner"></div>
    {/if}
</div>

<style>
    .loading-screen {
        min-height: 100vh;
        display: flex;
        align-items: center;
        justify-content: center;
        background: #0f0a14;
    }

    .spinner {
        width: 40px;
        height: 40px;
        border: 3px solid rgba(244, 114, 182, 0.2);
        border-top-color: #ec4899;
        border-radius: 50%;
        animation: spin 0.75s linear infinite;
    }

    .redirect-text {
        color: rgba(240, 171, 207, 0.6);
        font-size: 0.95rem;
    }

    @keyframes spin {
        to { transform: rotate(360deg); }
    }
</style>
