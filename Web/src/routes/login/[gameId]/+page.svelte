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

<form method="POST" action="?/login" class="container"> 
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