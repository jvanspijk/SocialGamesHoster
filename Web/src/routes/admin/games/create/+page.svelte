<script lang="ts">
    import type { PageProps } from './$types';
    
    let { data } = $props(); 

    let newPlayerName = $state("");
    let players: string[] = $state([]);
    let isFinalized = $state(false);

    let selectedRulesetId: number | undefined = $state(undefined);

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
        selectedRulesetId = undefined;
    }

    const canAddPlayer = $derived(newPlayerName.trim().length > 0);
    const canFinalizeList = $derived(players.length > 0 && selectedRulesetId !== undefined);

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
            <h3>1. Select Ruleset</h3>
            <select bind:value={selectedRulesetId} name="selectedRulesetId" required>
                <option value={""} disabled selected>Choose a Ruleset...</option>
                {#each data.rulesets as ruleset (ruleset.id)}
                    <option value={ruleset.id}>{ruleset.name}</option>
                {/each}
            </select>

            {#if selectedRuleset}
                <p class="ruleset-description">
                    {selectedRuleset.description}
                </p>
            {/if}
        </div>

        {#if selectedRulesetId !== undefined}
            <h3>2. Add Players</h3>
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
                <h3>3. Roster ({players.length})</h3>
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
            {:else}
                <p>Start by typing a name and pressing Enter or "Add Player."</p>
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
        {:else}
            <p style="margin-top: 20px; color: #888;">Please select a ruleset above to begin adding players.</p>
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