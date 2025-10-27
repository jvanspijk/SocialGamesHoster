<script lang="ts">
    import type { PageProps } from './$types';
    import type { GetActiveGameSessionsResponse } from '$lib/client';
    import { goto } from '$app/navigation';

    let { data }: PageProps = $props();
    // response: List(int Id, int RulesetId, string Status, int CurrentRoundNumber)
    let games: GetActiveGameSessionsResponse[] = (data.games as GetActiveGameSessionsResponse[] | undefined) ?? [];

    let message = $state("");
    let isLoading = $state(false);
    let selectedGameId = $state(null);

    async function handleJoin() {
        if(selectedGameId == null) {
            return;
        }
        await goto(`/login/${selectedGameId}`)
    };    
</script>

<div class="join-container">
    <h1>Join game</h1>
    <p>Active games:</p>

    {#if message}
        <p class={`message ${message}`}>
            {message}
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