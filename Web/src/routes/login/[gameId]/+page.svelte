<script lang="ts">
    import { client } from '$lib/client/client.gen';
    import { playerLogin } from '$lib/client'; 
	import { goto } from '$app/navigation';

    let { data }: PageProps = $props();
    let players = data.players;
    let gameId = data.gameId;

    let selectedPlayerId: number = $state(null);
    
    // State variables for UI feedback
    let isLoading: boolean = $state(false);
    let message: { text: string; type: 'success' | 'error' | '' } = $state({ text: '', type: '' });

    async function handleLogin() {
        message = { text: '', type: '' };
        isLoading = true;

        const selectedPlayer = players.find(p => p.id == selectedPlayerId);

        if (!gameId || !selectedPlayer) {
            message = { text: 'Both Game ID and Player Name are required.', type: 'error' };
            isLoading = false;
            return;
        }

        const playerName = selectedPlayer.Name;

        try {         
            const options = {
                path: {
                    gameId: gameId, 
                    name: playerName
                }
            };
            const { data: loginToken } = await playerLogin(options);
            // Handle success based on what your API returns
            console.log('Login Successful! Token:', loginToken);
            message = { 
                text: `Login successful for ${playerName}!`, 
                type: 'success' 
            };
            await goto('/me')
        } catch (error) {
            const errorDetails = (error as any).message || 'An unknown API error occurred.';
            const statusCode = (error as any).status ? ` (Status: ${(error as any).status})` : '';

            console.error('Login Exception:', error);
            message = { 
                text: `Login failed${statusCode}. ${errorDetails}`, 
                type: 'error' 
            };
        } finally {
            isLoading = false;
        }
    }
</script>

<div class="login-container">
    <h1>Player Login 🎮</h1>
    <p>Select your Player Name to join.</p>

    {#if message.text}
        <p class={`message ${message.type}`}>
            {message.text}
        </p>
    {/if}

    <select bind:value={selectedPlayerId}>
        <option value={null} disabled selected>Select a Player</option>
        {#each players as player (player.id)}
            <option value={player.id}>
                {player.name}
            </option>
        {/each}
    </select>

    <button onclick={handleLogin} disabled={isLoading || selectedPlayerId === null}>
        {#if isLoading}
            Logging in...
        {:else}
            Login
        {/if}
    </button>
</div>

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