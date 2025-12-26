<script lang="ts">
    import type { GetAllGameSessionsResponse } from '$lib/client/GameSessions/GetAllGameSessions';
    import LedgerTable from '$lib/components/LedgerTable.svelte';
    import SecondaryButton from '$lib/components/SecondaryButton.svelte';
    import { goto } from '$app/navigation';
	import BackLink from '$lib/components/BackLink.svelte';

    let { data } = $props();
</script>

<div class="page-header">
    <BackLink href="/admin" pageName="Admin Overview"></BackLink>
    <h1>Game Session Overview</h1>
</div>

<LedgerTable 
    data={data.games} 
    enableFilter={true} 
    filterOptions={['Not started', 'Running', 'Paused', 'Stopped', 'Cancelled', 'Finished']}
>
    {#snippet columns()}
        <tr>
            <th>Id</th>
            <th>Ruleset</th>
            <th>Players</th>
            <th>Status</th>
            <th>Actions</th>
        </tr>
    {/snippet}

    {#snippet rows(game: GetAllGameSessionsResponse)}
        <tr>
            <td data-label="Id" style="font-weight: bold;">{game.id}</td>
            <td data-label="Ruleset">{game.rulesetName}</td>
            <td data-label="Players">👤 {game.participantIds.length}</td>
            <td data-label="Status"><span class="status-stamp">{game.status}</span></td>
            <td>
                <div class="actions-wrapper">
                    <SecondaryButton onclick={() => goto(`/admin/games/${game.id}`)}>Manage</SecondaryButton>
                    <SecondaryButton variant="danger" onclick={() => {}}>Stop Game</SecondaryButton>
                </div>
            </td>
        </tr>
    {/snippet}
</LedgerTable>

<style>
    :global(th) {
        text-align: left;
        font-family: 'Cinzel', serif;
        border-bottom: 2px solid #5b4a3c;
        padding-bottom: 10px;
    }
    :global(td) {
        padding: 12px 0;
        border-bottom: 1px solid rgba(91, 74, 60, 0.2);
    }
    .status-stamp {
        text-transform: uppercase;
        font-size: 0.8rem;
        font-weight: bold;
        color: #a62a2a;
    }

    .actions-wrapper {
        display: flex;
        gap: 0.8rem;
        align-items: center;
        justify-content: center;
    }

    /* mobile */
    @media(max-width: 650px) {
        .actions-wrapper {
            gap: 1rem;            
        }
        
    }
</style>