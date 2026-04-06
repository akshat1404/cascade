<script lang="ts">
    import { onMount } from "svelte";
    import { supabase } from "$lib/supabase";

    let session: any = $state(null);
    let loading = $state(true);

    async function handleSession(currentSession: any) {
        if (!currentSession) return;

        try {
            const response = await fetch(
                `${import.meta.env.VITE_API_URL}/auth/callback`,
                {
                    method: "POST",
                    headers: {
                        Authorization: `Bearer ${currentSession.access_token}`,
                        "Content-Type": "application/json",
                    },
                },
            );

            await response.json();

            window.location.href = "/dashboard";
        } catch (err) {
            console.error("Callback error:", err);
        }
    }

    onMount(() => {
        init();
    });

    async function init() {
        try {
            const { data, error } = await supabase.auth.getSession();

            if (error) {
                console.error("Session error:", error);
            }

            session = data?.session ?? null;

            if (session) {
                await handleSession(session);
            }
        } catch (err) {
            console.error("init error:", err);
        } finally {
            loading = false;
        }

        const { data: listener } = supabase.auth.onAuthStateChange(
            async (_event, newSession) => {
                console.log("Auth change:", _event);

                session = newSession;

                if (newSession) {
                    await handleSession(newSession);
                }
            },
        );

        return () => {
            listener.subscription.unsubscribe();
        };
    }

    async function signIn() {
        const { error } = await supabase.auth.signInWithOAuth({
            provider: "google",
            options: {
                redirectTo: `${import.meta.env.VITE_SVELTE_DOMAIN}`,
            },
        });

        if (error) {
            console.error("Google Sign-In Error:", error);
        }
    }
</script>

<main>
    <h1>Cascade</h1>

    {#if loading}
        <p>Loading...</p>
    {:else if !session}
        <button onclick={signIn}>Continue with Google</button>
    {/if}
</main>

<style>
    main {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100vh;
        font-family: system-ui, sans-serif;
    }

    h1 {
        margin-bottom: 20px;
    }

    button {
        padding: 10px 20px;
        font-size: 16px;
        border: none;
        border-radius: 8px;
        cursor: pointer;
        background-color: black;
        color: white;
    }

    button:hover {
        opacity: 0.9;
    }
</style>
