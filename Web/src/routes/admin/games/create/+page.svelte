<script lang="ts">
    import MainSelect from '$lib/components/MainSelect.svelte';
    
    let { data } = $props(); 

    let newPlayerName = $state("");
    let players: string[] = $state([]);
    let isFinalized = $state(false);

    let selectedRulesetId: number | null = $state(null);

    function addPlayer() {
        const name = newPlayerName.trim();
        if (name.length > 0) {
            if (!players.includes(name)) {
                players = [...players, name]; 
            }
            newPlayerName = "";
        }
    }

    function removePlayer(index: number) {
        players = players.filter((_, i) => i !== index);
    }    
    
    function resetForm() {
        newPlayerName = "";
        players = [];
        isFinalized = false;
        selectedRulesetId = null;
    }

    const canAddPlayer = $derived(newPlayerName.trim().length > 0);
    const canFinalizeList = $derived(players.length > 0 && selectedRulesetId !== null);

    const selectedRuleset = $derived(
        data.rulesets.find(r => r.id === selectedRulesetId)
    );

    let isSubmitting = $state(false);
</script>

<div class="page-container">
    <h2>Create New Game</h2>

    <form method="POST" action="?/create">
    
    {#if !isFinalized}
        <div class="ruleset-select-area">
            <p>1. Select Ruleset</p>
            <MainSelect
                placeholder="Select a ruleset"
                options={data.rulesets.map((r) => ({
                    id: r.id,
                    label: r.name
                }))}
                bind:selectedValue={selectedRulesetId}
            />          

            {#if selectedRuleset}
                <p class="ruleset-description">
                    {selectedRuleset.description}
                </p>
            {/if}
        </div>

        {#if selectedRulesetId !== null}
            <p>2. Add Players</p>
            <div class="input-area">
                <input
                    type="text"
                    bind:value={newPlayerName}
                    onkeydown={(e) => {
                        if (e.key === 'Enter') {
                            e.preventDefault();
                            addPlayer(); 
                        }
                    }}
                    placeholder="Enter player name..."
                />
                <button type="button" onclick={addPlayer} disabled={!canAddPlayer}>
                    Add Player
                </button>
            </div>

            {#if players.length > 0}
                <p>3. Roster ({players.length})</p>
                <ol class="player-list">
                    {#each players as name, i}
                        <li class="player-item">
                            <span>{name}</span>
                            <button 
                                type="button"
                                onclick={() => removePlayer(i)} 
                                class="remove-button"
                                title="Remove {name}"
                            >
                                &times; 
                            </button>
                        </li>
                        
                        <input type="hidden" name="participant" value={name} />
                        
                    {/each}
                </ol>
            {/if}

            {#if players.length > 0}
                <button type="submit" disabled={!canFinalizeList || isSubmitting}>
                    {#if isSubmitting}
                        Submitting...
                    {:else}
                        Finalize Game
                    {/if}
                </button>
            {/if}
        {/if}

    {:else}
        <div class="success-state">
            <h3>✅ Game Setup Complete!</h3>
            <p>Ready to start the game with **{players.length}** players and **{selectedRuleset?.name || 'an unknown'}** ruleset.</p>
            <button type="button" onclick={resetForm}>Create Another Game</button>
        </div>
    {/if}
    
    </form>
</div>

<style>    
    .page-container {
        max-width: 600px;
        margin: 0 auto;
        padding: 20px;
    }

    h2 {
        font-family: 'Cinzel', serif;
        color: #5b4a3c;
        text-transform: uppercase;
        letter-spacing: 2px;
        font-size: 2.5em;
        text-align: center;
        border-bottom: 3px double #5b4a3c;
        padding-bottom: 15px;
        margin-bottom: 30px;
        text-shadow: 1px 1px 1px rgba(255, 255, 255, 0.5);
    }
    
    p:not(.ruleset-description) {
        font-family: 'Cinzel', serif;
        font-size: 1.4em;
        color: #8c2a3e;
        text-align: left;
        margin-top: 30px;
        margin-bottom: 10px;
        border-bottom: 1px solid #c9b48c;
        padding-bottom: 5px;
    }

    .ruleset-description {
        background-color: #e3d8b8;
        border: 1px solid #c9b48c;
        padding: 15px;
        margin-top: 15px;
        border-radius: 3px;
        line-height: 1.5;
        font-size: 1.1em;
        text-align: left;
        color: #3e322b;
    }

    .ruleset-select-area {
        text-align: left;
        width: 100%;
    }

    .input-area {
        display: flex;
        gap: 10px;
        align-items: center;
        margin-bottom: 25px;
    }

    .input-area input[type="text"] {
        flex-grow: 1;
        font-family: 'IM Fell English', serif;
        font-size: 1.2em;
        padding: 10px 15px;
        background-color: #fff9e6; 
        border: 2px solid #5b4a3c;
        border-radius: 3px;
        color: #3e322b;
        box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.2);
        max-width: none;
    }

    button {
        font-family: 'Cinzel', serif;
        text-transform: uppercase;
        letter-spacing: 1px;
        padding: 10px 20px;
        border: none;
        border-radius: 3px;
        cursor: pointer;
        transition: background-color 0.2s, box-shadow 0.2s;
        
        background-color: #8c2a3e;
        color: #f7e7c4;
        border: 2px solid #5b4a3c;
        box-shadow: 2px 2px 4px rgba(0, 0, 0, 0.4);
        font-weight: 700;
        white-space: nowrap;
    }

    button:hover:not(:disabled) {
        background-color: #a6384a;
        box-shadow: 3px 3px 6px rgba(0, 0, 0, 0.5);
    }

    button:disabled {
        background-color: #5b4a3c;
        color: #c9b48c;
        cursor: not-allowed;
        box-shadow: none;
    }

    h3 {
        font-family: 'Cinzel', serif;
        color: #5b4a3c;
        font-size: 1.5em;
        text-align: left;
        margin-top: 30px;
        margin-bottom: 10px;
    }

    .player-list {
        list-style: none;
        padding: 0;
        margin: 0;
        border: 2px solid #5b4a3c;
        border-radius: 3px;
        overflow: hidden;
        margin-bottom: 30px;
    }

    .player-item {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 10px 15px;
        background-color: #fff9e6;
        border-bottom: 1px solid #5b4a3c;
        font-family: 'IM Fell English', serif;
        font-size: 1.2em;
        color: #3e322b;
    }

    .player-item:last-child {
        border-bottom: none;
    }

    .remove-button {
        background: none;
        color: #8c2a3e;
        border: none;
        padding: 0 5px;
        line-height: 1;
        font-size: 1.5em;
        transition: color 0.2s;
        box-shadow: none;
        margin-left: 10px;
        text-shadow: 1px 1px 1px rgba(255, 255, 255, 0.5);
    }

    .remove-button:hover {
        color: #a6384a;
        transform: scale(1.1);
    }

    .success-state {
        padding: 40px 20px;
        background-color: #e7fce7;
        border: 3px solid #387038;
        border-radius: 5px;
        text-align: center;
        margin-top: 50px;
    }

    .success-state h3 {
        font-size: 2em;
        color: #387038;
        margin-bottom: 15px;
    }
    
    .success-state p {
        font-family: 'IM Fell English', serif;
        font-size: 1.3em;
        color: #3e322b;
        margin-bottom: 30px;
        text-align: center;
    }
</style>