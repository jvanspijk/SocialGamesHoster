<script lang="ts">
    import { enhance } from '$app/forms';
    import LedgerTable from '$lib/components/LedgerTable.svelte';
    import SecondaryButton from '$lib/components/SecondaryButton.svelte';

    import type { GetPlayersFromGameResponse } from '$lib/client/Players/GetPlayersFromGame';
	import { StartNewRound } from '$lib/client/Rounds/StartNewRound.js';
	import { ResumeCurrentRound } from '$lib/client/Rounds/ResumeCurrentRound.js';
	import { PauseCurrentRound } from '$lib/client/Rounds/PauseCurrentRound.js';
	import { StartGameSession } from '$lib/client/GameSessions/StartGameSession.js';
	import { FinishGameSession } from '$lib/client/GameSessions/FinishGameSession.js';
	import { invalidateAll } from '$app/navigation';
	import HUDFooter from '$lib/components/HUDFooter.svelte';
	import TimeDisplay from '$lib/components/TimeDisplay.svelte';
    import BackLink from '$lib/components/BackLink.svelte';


    let { data } = $props();

    let pendingRoles = $state<Record<number, number | null>>({});
    let winnerIds = $state<Set<number>>(new Set());
    let isSaving = $state(false);

    const isDirty = $derived(Object.keys(pendingRoles).length > 0);
    const gameId = $derived(data.gameSession.id.toString());

    async function handleRound(action: 'pause' | 'resume' | 'start') {
        let res;
        if (action === 'start') {
            const duration = prompt("Duration (seconds):", "300");
            if (!duration) return;
            res = await StartNewRound(fetch, { gameId, durationInSeconds: Number(duration) });
        } else if (action === 'pause') {
            res = await PauseCurrentRound(fetch, { gameId });
        } else {
            res = await ResumeCurrentRound(fetch, { gameId });
        }

        if (res?.ok) {
            await invalidateAll();
        }
    }

    function toggleWinner(id: number) {
        if (winnerIds.has(id)) winnerIds.delete(id);
        else winnerIds.add(id);
    }
</script>

<div class="page-main">
    <div class="parchment-container">
        <header class="admin-header">
            <BackLink href="/admin/games" pageName="Games Overview"></BackLink>
            <h1>Game Management</h1>
        </header>

        <section class="player-management">
            <h2>Player Roster</h2>
            {#await data.streamed.fullPlayers}
                <div class="loading-overlay">Retrieving player details...</div>
            {:then fullPlayers}
                {@const playerLookup = new Map(fullPlayers.filter(p => p !== null).map(p => [p!.id, p!]))}
                
                <LedgerTable 
                    data={data.players} 
                    enableFilter={false}
                >
                    {#snippet columns()}
                        <tr>
                            <th>Name</th>
                            <th>Role</th>
                            <th>Actions</th>
                        </tr>
                    {/snippet}

                    {#snippet rows(player: GetPlayersFromGameResponse)}
                        {@const fullData = playerLookup.get(player.id)}
                        <tr>
                            <td data-label="Name">
                                <span class={winnerIds.has(player.id) ? 'winner-highlight' : ''}>
                                    {player.name}
                                </span>
                            </td>
                            <td data-label="Role">
                                {#if fullData}
                                    <select 
                                        class="ledger-select-inline"
                                        value={pendingRoles[player.id] ?? fullData.role?.id ?? ''} 
                                        onchange={(e) => pendingRoles[player.id] = Number(e.currentTarget.value)}
                                    >
                                        <option value="">No Role</option>
                                        {#each data.roles as role}
                                            <option value={role.id}>{role.name}</option>
                                        {/each}
                                    </select>
                                {:else}
                                    <span class="status-error">Load Error</span>
                                {/if}
                            </td>
                            <td data-label="Fate">
                                <div class="action-cell">
                                    <button 
                                        type="button" 
                                        class="action-btn {winnerIds.has(player.id) ? 'active' : ''}"
                                        title="Mark as Winner"
                                        onclick={() => toggleWinner(player.id)}
                                    >🏆</button>
                                </div>
                            </td>
                        </tr>
                    {/snippet}
                </LedgerTable>
            {/await}
        </section>

        <footer class="admin-footer">
            {#if isDirty}
                <div class="sticky-action-bar">
                    <p>You have unsaved role assignments.</p>
                    <form method="POST" action="?/saveRoles" use:enhance={() => {
                        isSaving = true;
                        return async ({ update }) => {
                            await update();
                            pendingRoles = {};
                            isSaving = false;
                        };
                    }}>
                        <input type="hidden" name="updates" value={JSON.stringify(
                            Object.entries(pendingRoles).map(([id, roleId]) => ({ id: Number(id), roleId }))
                        )} />
                        <SecondaryButton type="submit" enabled={!isSaving}>
                            {isSaving ? 'Saving...' : 'Save All Roles'}
                        </SecondaryButton>
                        <button type="button" class="text-link" onclick={() => pendingRoles = {}}>Discard</button>
                    </form>
                </div>
            {/if}

            {#if winnerIds.size > 0}
                <form method="POST" action="?/declareWinners" use:enhance>
                    {#each Array.from(winnerIds) as id}
                        <input type="hidden" name="winnerIds" value={id} />
                    {/each}
                    <button type="submit" class="btn-finalize">Declare {winnerIds.size} Winner(s)</button>
                </form>
            {/if}
        </footer>
    </div>

    <HUDFooter>
        <div class="round-controls-mini">
            <span>Round #{data.currentRound?.id}</span>
            
            <TimeDisplay 
                initialSeconds={data.currentRound?.remainingSeconds ?? 0} 
                isPaused={data.currentRound?.isPaused ?? true}
            />

            {#if data.currentRound && data.currentRound.isPaused}
                <button onclick={() => handleRound('resume')}>▶ Resume Round</button>
            {:else}
                <button onclick={() => handleRound('pause')}>⏸ Pause Round</button>
            {/if}
        </div>        
        <button onclick={() => handleRound('start')}>Start new Round</button>    
    </HUDFooter>
</div>

<style>
    .admin-header {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 20px;
    }   

    .loading-overlay {
        padding: 40px;
        text-align: center;
        font-style: italic;
        color: #7a634e;
    }

    /* Action Buttons */
    .action-btn {
        background: none;
        border: 1px solid #5b4a3c;
        cursor: pointer;
        font-size: 1.2rem;
        padding: 5px 10px;
        filter: grayscale(1);
    }
    .action-btn.active { filter: grayscale(0); background: #fcf5e5; }

    .winner-highlight { font-weight: bold; color: #856404; text-decoration: underline; }

    .sticky-action-bar {
        position: sticky;
        bottom: 0;
        background: #f7e7c4;
        border: 2px solid #5b4a3c;
        padding: 15px;
        margin-top: 20px;
        text-align: center;
        box-shadow: 0 -5px 15px rgba(0,0,0,0.1);
    }

    .btn-finalize {
        background: #5b4a3c;
        color: #f7e7c4;
        font-family: 'Cinzel', serif;
        padding: 15px 30px;
        border: none;
        cursor: pointer;
        width: 100%;
        margin-top: 20px;
    }

    .text-link {
        background: none;
        border: none;
        color: #a62a2a;
        text-decoration: underline;
        cursor: pointer;
        margin-left: 10px;
    }

    .ledger-select-inline {
        background: transparent;
        border: 1px solid #5b4a3c;
        font-family: inherit;
        padding: 2px;
    }
</style>