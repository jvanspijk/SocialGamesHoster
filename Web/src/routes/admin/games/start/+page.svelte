<script lang="ts">
    import MainSelect from '$lib/components/MainSelect.svelte';
    import Description from '$lib/components/Description.svelte';    
    
    let { data } = $props(); 

    let isFinalized = $state(false);

    let selectedRulesetId: number | null = $state(null);
    
    function resetForm() {
        isFinalized = false;
        selectedRulesetId = null;
    }

    const selectedRuleset = $derived(
        data.rulesets.find(r => r.id === selectedRulesetId)
    );

    let isSubmitting = $state(false);
</script>

<div class="page-container">
    <h2>Start New Game</h2>

    <form method="POST" action="?/start">
    
    {#if !isFinalized}
        <div class="ruleset-select-area">
            <p>Select Ruleset</p>
            <MainSelect
                placeholder="Select a ruleset"
                options={data.rulesets.map((r) => ({
                    id: r.id,
                    label: r.name
                }))}
                bind:selectedValue={selectedRulesetId}
            />          

            <Description text={selectedRuleset?.description} isHidden={selectedRulesetId == null} />
        </div>    

        {#if selectedRulesetId !== null}
            <input type="hidden" name="selectedRulesetId" value={selectedRulesetId} />
            <button type="submit" disabled={selectedRulesetId == null}>
                {#if isSubmitting}
                    Submitting...
                {:else}
                    Start game
                {/if}
            </button>
        {/if}


    {:else}
        <div class="success-state">
            <h3>✅ Game Setup Complete!</h3>
            <p>Ready to start the game with **{selectedRuleset?.name || 'an unknown'}** ruleset.</p>
            <button type="button" onclick={resetForm}>Start Another Game</button>
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

    .ruleset-select-area {
        text-align: left;
        width: 100%;
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