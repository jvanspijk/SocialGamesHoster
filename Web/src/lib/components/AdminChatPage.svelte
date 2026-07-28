<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import ChatApp from './ChatApp.svelte';
	import Dialog from './Dialog.svelte';
	import { gameState } from '$lib/state/game.svelte';

	let { roomId = '' }: { roomId?: string } = $props();
	let newMessageOpen = $state(false);
	const view = $derived(gameState.admin);

	async function selectRoom(selected: string) {
		if (!view) return;
		await goto(
			resolve(
				selected
					? `/admin/games/${view.game.id}/chat/${selected}`
					: `/admin/games/${view.game.id}/chat`
			)
		);
	}

	async function choosePlayer(participantId: string) {
		const room = view?.rooms.find((candidate) => candidate.key === `gm:${participantId}`);
		if (!room) return;
		newMessageOpen = false;
		await selectRoom(room.id);
	}
</script>

{#if view}
	<ChatApp
		gameId={view.game.id}
		selectedRoomId={roomId}
		canModerate
		archived={view.game.status === 'archived'}
		{selectRoom}
		newMessage={() => (newMessageOpen = true)}
	/>
{/if}

<Dialog
	open={newMessageOpen}
	title="New message"
	description="Choose a player to open their direct conversation."
	close={() => (newMessageOpen = false)}
>
	<div class="player-list">
		{#each view?.participants.filter((player) => !['kicked', 'left'].includes(player.status)) ?? [] as player (player.id)}
			<button type="button" onclick={() => choosePlayer(player.id)}>
				<span>{(player.gameAlias || player.displayNameSnapshot).slice(0, 1).toUpperCase()}</span>
				<strong>{player.gameAlias || player.displayNameSnapshot}</strong>
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
