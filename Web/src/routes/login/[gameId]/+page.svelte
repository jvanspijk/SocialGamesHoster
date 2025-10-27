<script lang="ts">
    import type { GetPlayersFromGameResponse } from '$lib/client'; 
	import type { PageProps } from './$types';

    let { data }: PageProps = $props();
    let players: GetPlayersFromGameResponse[] = data.players ?? [];

    let selectedPlayerId: number | null = $state(null);
    
    let isLoading: boolean = $state(false);
    let message: { text: string; type: 'success' | 'error' | '' } = $state({ text: '', type: '' });
</script>

<form method="POST" action="?/login" class="login-container"> 
    <h1>Player Login</h1>
    <p>Select your Player Name to join.</p>

    {#if message.text}
        <p class={`message ${message.type}`}>
            {message.text}
        </p>
    {/if}

    <select name="selectedPlayerId" bind:value={selectedPlayerId}>
        <option value={null} disabled selected>Select a Player</option>
        {#each players as player (player.id)}
            <option value={player.id}>
                {player.name}
            </option>
        {/each}
    </select>
    
    <button type="submit" disabled={isLoading || selectedPlayerId === null}>
        {#if isLoading}
            Logging in...
        {:else}
            Login
        {/if}
    </button>
</form>

<style>
    form {
        display: flex;
        flex-direction: column;
        align-items: center;
        text-align: center;
        padding: 20px;
    }

    .form-content {
        display: flex;
        flex-direction: column;
        gap: 25px;
        width: 100%;
        max-width: 400px;
        margin-top: 10px;
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
    
    /* Message Bar Styling */
    .message-bar {
        font-family: 'IM Fell English', serif;
        padding: 10px;
        border: 2px solid;
        border-radius: 3px;
        text-align: center;
        margin: 0;
    }
    
    .error-message {
        color: #a62a2a;
        border-color: #a62a2a;
        background-color: #fce7e7;
    }

    .success-message {
        color: #387038;
        border-color: #387038;
        background-color: #e7fce7;
    }

</style>