<script lang="ts">
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { Room } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';
	import DirectMessageChatPage from './DirectMessageChatPage.svelte';
	import { directMessageRecipient } from './directMessageRecipients';

	let { roomId = '' }: { roomId?: string } = $props();
	const view = $derived(gameState.player);
	const recipients = $derived(
		view?.party
			.filter((player) => player.profileId !== auth.actor?.id && player.status === 'active')
			.map((player) =>
				directMessageRecipient(player.id, player.gameAlias || player.displayName, player.seatNumber)
			) ?? []
	);

	async function openConversation(participantId: string) {
		if (!view) return;
		try {
			const room = await api<Room>(`/games/${view.game.id}/rooms/player-dm`, {
				method: 'POST',
				...jsonBody({ participantId })
			});
			return room.id;
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The conversation could not be opened.'));
		}
	}
</script>

{#if view}
	<DirectMessageChatPage
		gameId={view.game.id}
		{roomId}
		archived={view.game.status === 'archived'}
		policyRevision={`${view.game.status}:${view.game.phaseKey}`}
		recipientEntries={recipients}
		roomPath={(selected) => (selected ? `/play/chat/${selected}` : '/play/chat')}
		{openConversation}
	/>
{/if}
