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
    <h1>Player Login 🎮</h1>
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
    .login-container {
        max-width: 400px;
        margin: 50px auto;
        padding: 20px;
        border: 1px solid #ddd;
        border-radius: 8px;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        text-align: center;
    }

    h1 {
        color: #333;
    }

    .form-group {
        margin-bottom: 15px;
        text-align: left;
    }

    button {
        width: 100%;
        padding: 10px;
        background-color: #007bff;
        color: white;
        border: none;
        border-radius: 4px;
        cursor: pointer;
        font-size: 16px;
        margin-top: 10px;
    }

    button:disabled {
        background-color: #a0c9ff;
        cursor: not-allowed;
    }

    .message {
        padding: 10px;
        border-radius: 4px;
        margin-bottom: 15px;
        font-weight: bold;
    }

    .success {
        background-color: #d4edda;
        color: #155724;
        border: 1px solid #c3e6cb;
    }

    .error {
        background-color: #f8d7da;
        color: #721c24;
        border: 1px solid #f5c6cb;
    }
</style>