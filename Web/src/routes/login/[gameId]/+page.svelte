<script lang="ts">
    import type { GetPlayersFromGameResponse } from '$lib/client'; 
	import type { PageProps } from './$types';
    import MainSelect from '$lib/components/MainSelect.svelte';
    import MainButton from '$lib/components/MainButton.svelte';

    let { data }: PageProps = $props();
    let players: GetPlayersFromGameResponse[] = data.players ?? [];

    let selectedPlayerId: number | null = $state(null);
    
    let isLoading: boolean = $state(false);
    let message: { text: string; type: 'success' | 'error' | '' } = $state({ text: '', type: '' });

    const playerOptions = players.map((p) => ({
		id: p.id,
		label: p.name
	}));
</script>

<form method="POST" action="?/login" class="login-container"> 
    <h1>Player Login</h1>
    <p>Select your Player Name to join.</p>

    {#if message.text}
        <p class={`message ${message.type}`}>
            {message.text}
        </p>
    {/if}

    <MainSelect
		name="selectedPlayerId"
		placeholder="Select a Player"
		options={playerOptions}
		bind:selectedValue={selectedPlayerId}
	/>
    
    <MainButton
		label="Login"
		type="submit"
		disabled={isLoading || selectedPlayerId === null}
		isLoading={isLoading}
	/>
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