<script lang="ts">
    import { onMount } from "svelte";
    import { supabase } from "$lib/supabase";
    import Navbar from "$lib/components/Navbar.svelte";
    import DocumentsPage from "$lib/components/DocumentsPage.svelte";
    import LoadingScreen from "$lib/components/LoadingScreen.svelte";

    let user = $state<any>(null);
    let loading = $state(true);

    onMount(() => {
        supabase.auth.getSession().then(({ data }) => {
            user = data.session?.user ?? null;
            loading = false;
            if (!user) window.location.href = "/";
        });

        const {
            data: { subscription },
        } = supabase.auth.onAuthStateChange((_event, session) => {
            user = session?.user ?? null;
            loading = false;
            if (!user) window.location.href = "/";
        });

        return () => subscription.unsubscribe();
    });

    async function signOut() {
        await supabase.auth.signOut();
        window.location.href = "/";
    }

    function createDocument() {}
</script>

<svelte:head>
    <title>Collaborent — Dashboard</title>
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link
        rel="preconnect"
        href="https://fonts.gstatic.com"
        crossorigin="anonymous"
    />
    <link
        href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap"
        rel="stylesheet"
    />
</svelte:head>

{#if loading}
    <LoadingScreen />
{:else if user}
    <div class="page-wrapper">
        <Navbar onSignOut={signOut} onCreateDoc={createDocument} />
        <DocumentsPage />
    </div>
{:else}
    <LoadingScreen message="Redirecting..." />
{/if}

<style>
    :global(*, *::before, *::after) {
        box-sizing: border-box;
        margin: 0;
        padding: 0;
    }

    :global(body) {
        font-family: "Inter", sans-serif;
        background: #0f0a14;
        color: #f5e6f0;
        min-height: 100vh;
    }

    .page-wrapper {
        min-height: 100vh;
        display: flex;
        flex-direction: column;
        background: radial-gradient(
                ellipse 80% 50% at 20% -10%,
                rgba(244, 114, 182, 0.18) 0%,
                transparent 60%
            ),
            radial-gradient(
                ellipse 60% 40% at 80% 110%,
                rgba(168, 85, 247, 0.14) 0%,
                transparent 60%
            ),
            #0f0a14;
    }
</style>
