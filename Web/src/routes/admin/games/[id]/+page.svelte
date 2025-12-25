<script lang="ts">
    import { page } from '$app/stores';
    import SecondaryButton from '$lib/components/SecondaryButton.svelte';
    import { onMount } from 'svelte';
    import BackLink from '$lib/components/BackLink.svelte';

    // Mock data based on your GetAllGameSessionsResponse structure
    let game = $state({
        id: $page.params.id,
        rulesetName: "Town of Salem",
        status: "RUNNING",
        startTime: "14:22",
        participants: [
            { id: "1", name: "Giles Corey", role: "Investigator", status: "ALIVE" },
            { id: "2", name: "Cotton Mather", role: "Executioner", status: "DEAD" },
            { id: "3", name: "Mary Warren", role: "Jester", status: "ALIVE" }
        ]
    });
</script>

<div class="management-header">
    <BackLink href="/admin/games" pageName="Game Sessions Overview" />
    <div class="header-main">
        <h1>Session #{game.id}</h1>
        <span class="status-badge {game.status.toLowerCase()}">{game.status}</span>
    </div>
</div>

<div class="management-grid">
    <section class="meta-scroll">
        <h3>Game Particulars</h3>
        <dl class="ledger-details">
            <dt>Ruleset</dt>
            <dd>{game.rulesetName}</dd>
            
            <dt>Commenced At</dt>
            <dd>{game.startTime}</dd>
            
            <dt>Elapsed Time</dt>
            <dd>42 Minutes</dd>
        </dl>

        <div class="session-actions">
            <h4>Administrative Decrees</h4>
            <SecondaryButton onclick={() => {}}>Broadcast Message</SecondaryButton>
            <SecondaryButton onclick={() => {}}>Edit Ruleset</SecondaryButton>
            <SecondaryButton variant="danger" onclick={() => {}}>Terminate Session</SecondaryButton>
        </div>
    </section>

    <section class="player-registry">
        <h3>Participants ({game.participants.length})</h3>
        <table class="registry-table">
            <thead>
                <tr>
                    <th>Subject</th>
                    <th>Status</th>
                    <th>Action</th>
                </tr>
            </thead>
            <tbody>
                {#each game.participants as player}
                    <tr>
                        <td>
                            <div class="player-info">
                                <span class="player-name">{player.name}</span>
                                <span class="player-role">{player.role}</span>
                            </div>
                        </td>
                        <td>
                            <span class="status-dot {player.status.toLowerCase()}"></span>
                            {player.status}
                        </td>
                        <td>
                            <button class="icon-btn" title="Punish">⚖️</button>
                        </td>
                    </tr>
                {/each}
            </tbody>
        </table>
    </section>
</div>

<style>
    .management-header {
        margin-bottom: 2rem;
        border-bottom: 2px double #5b4a3c;
        padding-bottom: 1rem;
    }

    .back-link {
        font-family: 'Cinzel', serif;
        color: #8b7355;
        text-decoration: none;
        font-size: 0.9rem;
        transition: color 0.2s;
    }

    .back-link:hover { color: #5b4a3c; }

    .header-main {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: 0.5rem;
    }

    /* Grid Layout: Stacks on mobile, splits on desktop */
    .management-grid {
        display: grid;
        grid-template-columns: 1fr;
        gap: 2rem;
    }

    @media (min-width: 900px) {
        .management-grid {
            grid-template-columns: 300px 1fr;
        }
    }

    /* Meta Details (The Left Column) */
    .meta-scroll {
        background: rgba(91, 74, 60, 0.05);
        padding: 1.5rem;
        border-right: 1px solid rgba(91, 74, 60, 0.2);
    }

    .ledger-details dt {
        font-family: 'Cinzel', serif;
        font-size: 0.75rem;
        color: #8b7355;
        text-transform: uppercase;
        margin-top: 1rem;
    }

    .ledger-details dd {
        font-size: 1.1rem;
        font-weight: bold;
        margin-left: 0;
        color: #3e322b;
    }

    .session-actions {
        margin-top: 3rem;
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    /* Registry Table (The Right Column) */
    .registry-table {
        width: 100%;
        border-collapse: collapse;
    }

    .registry-table th {
        text-align: left;
        font-family: 'Cinzel', serif;
        border-bottom: 2px solid #5b4a3c;
        padding: 10px;
    }

    .registry-table td {
        padding: 15px 10px;
        border-bottom: 1px solid rgba(91, 74, 60, 0.1);
    }

    .player-info { display: flex; flex-direction: column; }
    .player-name { font-weight: bold; }
    .player-role { font-size: 0.8rem; opacity: 0.7; font-style: italic; }

    .status-dot {
        display: inline-block;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        margin-right: 5px;
    }
    .status-dot.alive { background: #387038; }
    .status-dot.dead { background: #a62a2a; }

    .icon-btn {
        background: none;
        border: none;
        cursor: pointer;
        font-size: 1.2rem;
        filter: sepia(1);
    }
</style>