<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import SelectionDialog from '$lib/components/SelectionDialog.svelte';
	import ChatApp from './ChatApp.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { Room } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let { roomId = '' }: { roomId?: string } = $props();
	let newMessageOpen = $state(false);
	const view = $derived(gameState.player);
	const recipients = $derived(
		view?.party
			.filter((player) => player.profileId !== auth.actor?.id)
			.map((player) => {
				const displayLabel = player.gameAlias || player.displayName;
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
			toasts.error(errorMessage(caught, 'The conversation could not be opened.'));
		}
	}
</script>

{#if view}
	<ChatApp
		gameId={view.game.id}
		selectedRoomId={roomId}
		archived={view.game.status === 'archived'}
		policyRevision={`${view.game.status}:${view.game.phaseKey}`}
		{selectRoom}
		newMessage={() => (newMessageOpen = true)}
	/>
{/if}

<SelectionDialog
	open={newMessageOpen}
	title="New message"
	description="Choose a player to start or open a direct conversation."
	close={() => (newMessageOpen = false)}
	emptyState={{
		title: 'No players available',
		description: 'No players are available for a direct conversation.'
	}}
	entries={recipients}
	onselect={choosePlayer}
/>
