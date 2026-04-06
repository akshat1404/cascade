<script lang="ts">
    import { onMount } from "svelte";
    import { supabase } from "$lib/supabase";
    import Navbar from "$lib/components/Navbar.svelte";
    import DocumentsPage from "$lib/components/DocumentsPage.svelte";
    import LoadingScreen from "$lib/components/LoadingScreen.svelte";
    import CreateDocModal from "$lib/components/CreateDocModal.svelte";

    let user = $state<any>(null);
    let loading = $state(true);
    let showModal = $state(false);
    let documents = $state<any[]>([]);
    let docsLoading = $state(false);

    async function fetchDocuments() {
        docsLoading = true;
        try {
            const {
                data: { session },
            } = await supabase.auth.getSession();
            const res = await fetch(`${import.meta.env.VITE_API_URL}/documents`, {
                headers: { Authorization: `Bearer ${session?.access_token}` },
            });
            if (res.ok) documents = await res.json();
        } finally {
            docsLoading = false;
        }
    }

    onMount(() => {
        supabase.auth.getSession().then(({ data }) => {
            user = data.session?.user ?? null;
            loading = false;
            if (!user) window.location.href = "/";
            else fetchDocuments();
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

    function createDocument() {
        showModal = true;
    }

    function handleCloseModal() {
        showModal = false;
    }

    async function handleCreate(title: string, content?: string) {
        if (!title) {
            alert("Title is required");
            return;
        }

        const {
            data: { session },
        } = await supabase.auth.getSession();

        const res = await fetch(`${import.meta.env.VITE_API_URL}/documents/create`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${session?.access_token}`,
            },
            body: JSON.stringify({ title }),
        });

        if (!res.ok) {
            console.error("Failed to create document");
            return;
        }

        const doc = await res.json();

        // if content is provided from import, we save it immediately
        if (content) {
            await fetch(`${import.meta.env.VITE_API_URL}/documents/${doc.id}`, {
                method: "PUT",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${session?.access_token}`,
                },
                body: JSON.stringify({
                    title,
                    content,
                }),
            });
        }

        showModal = false;
        window.location.href = `/document/${doc.id}`;
    }

    async function handleDocDelete(id: string, e?: Event) {
        if (e) {
            e.preventDefault();
            e.stopPropagation();
        }
        
        if (!confirm("Are you sure you want to delete this document?")) {
            return;
        }

        const {
            data: { session },
        } = await supabase.auth.getSession();

        const res = await fetch(`${import.meta.env.VITE_API_URL}/documents/${id}`, {
            method: "DELETE",
            headers: {
                Authorization: `Bearer ${session?.access_token}`,
            },
        });

        if (!res.ok) {
            alert("Failed to delete document");
            return;
        }

        // Optimistically remove from state
        documents = documents.filter(d => d.id !== id);
    }
</script>

<svelte:head>
    <title>Cascade — Dashboard</title>
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
        <DocumentsPage {documents} loading={docsLoading} onDelete={handleDocDelete} />
    </div>
{:else}
    <LoadingScreen message="Redirecting..." />
{/if}

{#if showModal}
    <CreateDocModal onClose={handleCloseModal} onCreate={handleCreate} />
{/if}

<style>
    :global(*, *::before, *::after) {
        box-sizing: border-box;
        margin: 0;
        padding: 0;
    }

    :global(body) {
        font-family: "Inter", sans-serif;
        background: #ffffff;
        color: #1a0a12;
        min-height: 100vh;
    }

    .page-wrapper {
        min-height: 100vh;
        display: flex;
        flex-direction: column;
        background: radial-gradient(
                ellipse 80% 50% at 20% -10%,
                rgba(244, 114, 182, 0.12) 0%,
                transparent 60%
            ),
            radial-gradient(
                ellipse 60% 40% at 80% 110%,
                rgba(168, 85, 247, 0.08) 0%,
                transparent 60%
            ),
            #ffffff;
    }
</style>
