<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import ChatApp from './ChatApp.svelte';
	import Dialog from './Dialog.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import type { Room } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let { roomId = '' }: { roomId?: string } = $props();
	let newMessageOpen = $state(false);
	const view = $derived(gameState.player);

	async function selectRoom(selected: string) {
		await goto(resolve(selected ? `/play/chat/${selected}` : '/play/chat'));
	}

	async function choosePlayer(participantId: string) {
		if (!view) return;
		try {
			const room = await api<Room>(`/games/${view.game.id}/rooms/player-dm`, {
				method: 'POST',
				...jsonBody({ participantId })
			});
			newMessageOpen = false;
			await selectRoom(room.id);
		} catch (caught) {
			toasts.error(
				caught instanceof Error ? caught.message : 'The conversation could not be opened.'
			);
		}
	}
</script>

{#if view}
	<ChatApp
		gameId={view.game.id}
		selectedRoomId={roomId}
		archived={view.game.status === 'archived'}
		{selectRoom}
		newMessage={() => (newMessageOpen = true)}
	/>
{/if}

<Dialog
	open={newMessageOpen}
	title="New message"
	description="Choose a player to start or open a direct conversation."
	close={() => (newMessageOpen = false)}
>
	<div class="player-list">
		{#each view?.party.filter((player) => player.profileId !== auth.actor?.id) ?? [] as player (player.id)}
			<button type="button" onclick={() => choosePlayer(player.id)}>
				<span>{(player.gameAlias || player.displayName).slice(0, 1).toUpperCase()}</span>
				<strong>{player.gameAlias || player.displayName}</strong>
				<small>Seat {player.seatNumber}</small>
			</button>
		{/each}
	</div>
</Dialog>

<style>
	.player-list {
		display: grid;
	}

	.player-list button {
		display: grid;
		min-height: 4rem;
		grid-template-columns: auto 1fr auto;
		align-items: center;
		gap: var(--space-3);
		border: 0;
		border-block-end: var(--border-subtle);
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		text-align: start;
	}

	.player-list button > span {
		display: grid;
		width: 2.5rem;
		height: 2.5rem;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--ink);
		color: var(--gold-light);
	}

	.player-list small {
		color: var(--ink-soft);
	}
</style>
