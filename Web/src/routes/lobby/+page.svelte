<script lang="ts">
    import type { PageProps } from './$types';
    import { goto } from '$app/navigation';

    let { data }: PageProps = $props();
    let games = data.games;
    let message = $state("");
    let isLoading = $state(false);
    let selectedGameId = $state(null);

    async function handleJoin() {
        await goto(`/login/${selectedGameId}`)
    };    
</script>

<div class="join-container">
    <h1>Join game</h1>
    <p>Select an activate game to join.</p>

    {#if message.text}
        <p class={`message ${message.type}`}>
            {message.text}
        </p>
    {/if}

    <select bind:value={selectedGameId}>
        <option value={null} disabled selected>Select a Game</option>
        {#each games as game (game.id)}
            <option value={game.id}>
                {game.id}
            </option>
        {/each}
    </select>

    <button onclick={handleJoin} disabled={isLoading || selectedGameId === null}>
        {#if isLoading}
            Joining game...
        {:else}
            Join game
        {/if}
    </button>
</div>