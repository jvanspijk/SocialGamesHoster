<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import SelectionDialog from '$lib/components/SelectionDialog.svelte';
	import ChatApp from './ChatApp.svelte';
	import { gameState } from '$lib/state/game.svelte';

	let { roomId = '' }: { roomId?: string } = $props();
	let newMessageOpen = $state(false);
	const view = $derived(gameState.admin);
	const recipients = $derived(
		view?.participants
			.filter((player) => !['kicked', 'left'].includes(player.status))
			.map((player) => {
				const displayLabel = player.gameAlias || player.displayNameSnapshot;
				const supportingLabel = `Seat ${player.seatNumber}`;
				const avatarText = displayLabel.slice(0, 1).toUpperCase();
				return {
					id: player.id,
					label: displayLabel,
					accessibleLabel: `${displayLabel}, ${supportingLabel}`,
					supportingLabel,
					leadingText: avatarText
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

<SelectionDialog
	open={newMessageOpen}
	title="New message"
	description="Choose a player to open their direct conversation."
	close={() => (newMessageOpen = false)}
	emptyState={{
		title: 'No players available',
		description: 'No players are available for a direct conversation.'
	}}
	entries={recipients}
	onselect={choosePlayer}
/>
