<script lang="ts">
    import type { PageProps } from './$types';
    import type { GetActiveGameSessionsResponse, GetRulesetResponse } from '$lib/client';
    import { goto } from '$app/navigation';

    let { data }: PageProps = $props();
    let games: GetActiveGameSessionsResponse[] = (data.games as GetActiveGameSessionsResponse[] | undefined) ?? [];
    let rulesetPromise: Promise<Record<number, GetRulesetResponse>> = data.rulesets as Promise<Record<number, GetRulesetResponse>>;

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

<div class="join-page-main">
    <div class="parchment-container">
        <h1>Join A Game</h1>
        
        <p class="role-title">Select a game session.</p>
        
        <div class="controls-section">
            
            {#await rulesetPromise then descriptions}        
                
                <select bind:value={selectedGameId}>
                    <option value={null} disabled selected={selectedGameId === null}>-- Select a Game Session --</option>
                    {#each games as game (game.id)}
                        {@const ruleset = descriptions[game.id]}
                        <option value={game.id}>
                            Game {game.id} 
                            {#if ruleset}
                                ({ruleset.name})
                            {:else}
                                (Details unavailable)
                            {/if}
                        </option>
                    {/each}
                </select>

                <button onclick={handleJoin} disabled={selectedGameId === null || isLoading}>
                    {isLoading ? 'Preparing to Enter...' : 'Enter Selected Game'}
                </button>
                
                {#if selectedGameId !== null}
                    {@const ruleset = descriptions[selectedGameId]}
                    <div class="ruleset-info">
                        <h2>
                            {ruleset?.name || 'Loading Details...'}
                        </h2>
                        <p>
                            {ruleset?.description || ""}
                        </p>
                    </div>
                {/if}

                <div class="message-bar">
                    {#if message}{message}{/if}
                </div>
            {:catch error}
                <div class="error-message">
                    ERROR: Could not retrieve game data.
                </div>
            {/await}
        </div>
    </div>
</div>

<style>    
    .controls-section {
        display: flex;
        flex-direction: column;
        gap: 20px;
        margin-bottom: 30px;
        align-items: center;
    }
    
    select {
        font-family: 'IM Fell English', serif;
        font-size: 1.2em;
        padding: 10px 15px;
        background-color: #fff9e6;
        border: 2px solid #5b4a3c;
        border-radius: 3px;
        color: #3e322b;
        width: 100%;
        max-width: 400px;
        box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.2);
        appearance: none;
        background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" fill="%235b4a3c" viewBox="0 0 10 10"><path d="M5 8l-5-5h10z"/></svg>');
        background-repeat: no-repeat;
        background-position: right 15px center;
        cursor: pointer;
    }

    button {
        font-family: 'Cinzel', serif;
        font-size: 1.2em;
        text-transform: uppercase;
        padding: 10px 20px;
        color: #f7e7c4;
        background-color: #a62a2a;
        border: 3px solid #5b4a3c;
        border-radius: 5px;
        cursor: pointer;
        transition: background-color 0.2s, box-shadow 0.2s;
        box-shadow: 2px 2px 5px rgba(0, 0, 0, 0.4);
        letter-spacing: 1px;
    }

    button:hover:enabled {
        background-color: #7a1d1d;
        box-shadow: 3px 3px 8px rgba(0, 0, 0, 0.6);
    }
    
    button:disabled {
        background-color: #999;
        border-color: #777;
        cursor: not-allowed;
    }
    
    .ruleset-info {
        margin-top: 20px;
    }

    .message-bar {
        text-align: center;
        font-family: 'Cinzel', serif;
        color: #a62a2a;
        font-size: 1.1em;
        height: 1.5em;
    }
</style>