<script lang="ts">
    import type { PageProps } from './$types';
    import type { GetActiveGameSessionsResponse } from '$lib/client/GameSessions/GetActiveGameSessions';
    import { goto } from '$app/navigation';
	import MainButton from '$lib/components/MainButton.svelte';
    import MainSelect from '$lib/components/MainSelect.svelte';

    let { data }: PageProps = $props();
    let games = $derived((data.games as GetActiveGameSessionsResponse[] | undefined) ?? []);

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
    <h1>Lobby</h1>
    
    <p>Select a game session.</p>
    
    <div class="controls-section">        
        {#await data.streamed.rulesets}
        <MainSelect
                placeholder="Loading game sessions..."
                options={games.map((g) => ({
                    id: g.id,
                    label: `Game ${g.id} (Loading details...)`
                }))}
                bind:selectedValue={selectedGameId}
                enabled={false} 
            />
        {:then descriptions}
            <MainSelect
                placeholder="Select a Game Session"
                options={games.map((g) => ({
                    id: g.id,
                    label: `Game ${g.id} (${descriptions[g.id]?.name ?? 'Details unavailable'})`
                }))}
                bind:selectedValue={selectedGameId}
            />

            <MainButton 
                disabled={selectedGameId === null || isLoading} 
                isLoading={isLoading} 
                onActivate={handleJoin}
                label={'Enter Selected Game'} 
            />
            
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

<style>  
    .join-page-main {
        display: flex;
        flex-direction: column;
        align-items: center;
        text-align: center;
        padding: 20px;
        gap: 5px;
        margin-top: 10px;
    }

    .controls-section {
        display: flex;
        flex-direction: column;
        margin-bottom: 30px;
        align-items: center;
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