<script lang="ts">
    import type { PageProps } from './$types';

    const { data }: PageProps = $props();
    const games = $derived(data.games);
</script>

<div class="games-table-container">
    <h2>Game Administration Overview</h2>
    
    {#if games && games.length > 0}
        <table class="games-table">
            <thead>
                <tr>
                    <th>ID</th>
                    <th>Ruleset</th>
                    <th>Participants</th>
                    <th>Status</th>
                    <th>Actions</th>
                </tr>
            </thead>
            <tbody>
                {#each games as game (game.id)}
                    <tr class="game-row game-status-{game.status.toLowerCase()}">
                        <td>{game.id}</td>
                        <td>{game.rulesetName}</td>
                        <td>{game.participantIds.length}</td>
                        <td class="status-cell">
                            <span class="status-indicator" aria-label="Status: {game.status}"></span>
                            {game.status}
                        </td>
                        <td>
                            <button onclick={() => alert(`Viewing Game ${game.id}`)}>View</button>
                            <button class="action-danger" onclick={() => confirm(`Are you sure you want to delete Game ${game.id}?`)}>Delete</button>
                        </td>
                    </tr>
                {/each}
            </tbody>
        </table>
    {:else}
        <p class="no-games">No games found in the database. Start a new one!</p>
    {/if}
</div>

<style>
    .games-table-container {
        padding: 20px;
        max-width: 1200px;
        margin: 0 auto;
        font-family: sans-serif;
    }

    h2 {
        color: #333;
        margin-bottom: 1.5em;
        border-bottom: 2px solid #eee;
        padding-bottom: 10px;
    }

    .games-table {
        width: 100%;
        border-collapse: collapse;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    }

    .games-table thead tr {
        background-color: #f4f4f4;
        color: #333;
        text-align: left;
    }

    .games-table th, .games-table td {
        padding: 12px 15px;
        border: 1px solid #ddd;
    }
    
    .games-table tbody tr:hover {
        background-color: #f9f9f9;
    }

    .status-cell {
        display: flex;
        align-items: center;
        gap: 8px;
    }

    .status-indicator {
        display: inline-block;
        width: 10px;
        height: 10px;
        border-radius: 50%;
    }

    .game-status-in_progress .status-indicator {
        background-color: orange;
    }
    .game-status-completed .status-indicator {
        background-color: green;
    }
    .game-status-pending .status-indicator {
        background-color: blue;
    }

    .games-table button {
        padding: 6px 10px;
        margin-right: 5px;
        border: none;
        border-radius: 4px;
        cursor: pointer;
        transition: background-color 0.2s;
    }

    .games-table button:not(.action-danger) {
        background-color: #5cb85c;
        color: white;
    }
    .games-table button:not(.action-danger):hover {
        background-color: #4cae4c;
    }

    .action-danger {
        background-color: #d9534f;
        color: white;
    }
    .action-danger:hover {
        background-color: #c9302c;
    }

    .no-games {
        text-align: center;
        padding: 50px;
        color: #999;
        font-size: 1.2em;
        border: 1px dashed #ddd;
        margin-top: 20px;
    }
</style>