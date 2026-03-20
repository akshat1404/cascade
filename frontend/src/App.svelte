<script lang="ts">
  import { supabase } from "./lib/supabase";

  async function signIn() {
    const { error } = await supabase.auth.signInWithOAuth({
      provider: "google",
      options: {
        redirectTo: "http://localhost:8080/auth/callback",
      },
    });
    if (error) console.error(error);
  }
</script>

<main>
  <h1>Collaborent</h1>
  <button on:click={signIn}>Continue with Google</button>
</main>
