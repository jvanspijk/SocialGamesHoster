<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import ChatApp from './ChatApp.svelte';
	import DirectMessageChooser from './DirectMessageChooser.svelte';
	import { gameState } from '$lib/state/game.svelte';

	let { roomId = '' }: { roomId?: string } = $props();
	let newMessageOpen = $state(false);
	const view = $derived(gameState.admin);
	const recipients = $derived(
		view?.participants
			.filter((player) => !['kicked', 'left'].includes(player.status))
			.map((player) => {
				const displayLabel = player.gameAlias || player.displayNameSnapshot;
				return {
					id: player.id,
					displayLabel,
					supportingLabel: `Seat ${player.seatNumber}`,
					avatarText: displayLabel.slice(0, 1).toUpperCase()
				};
			}) ?? []
	);

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
		policyRevision={`${view.game.status}:${view.game.phaseKey}`}
		{selectRoom}
		newMessage={() => (newMessageOpen = true)}
	/>
{/if}

<DirectMessageChooser
	open={newMessageOpen}
	description="Choose a player to open their direct conversation."
	close={() => (newMessageOpen = false)}
	entries={recipients}
	onchoose={choosePlayer}
/>
