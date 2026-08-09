<script lang="ts">
	import { gameState } from '$lib/state/game.svelte';
	import DirectMessageChatPage from './DirectMessageChatPage.svelte';
	import { directMessageRecipient } from './directMessageRecipients';

	let { roomId = '' }: { roomId?: string } = $props();
	const view = $derived(gameState.admin);
	const recipients = $derived(
		view?.participants
			.filter((player) => !['kicked', 'left'].includes(player.status))
			.map((player) =>
				directMessageRecipient(
					player.id,
					player.gameAlias || player.displayNameSnapshot,
					player.seatNumber
				)
			) ?? []
	);

	function openConversation(participantId: string) {
		return view?.rooms.find((candidate) => candidate.key === `gm:${participantId}`)?.id;
	}
</script>

{#if view}
	<DirectMessageChatPage
		gameId={view.game.id}
		{roomId}
		canModerate
		archived={view.game.status === 'archived'}
		policyRevision={`${view.game.status}:${view.game.phaseKey}`}
		recipientEntries={recipients}
		roomPath={(selected) =>
			selected
				? `/admin/games/${view.game.id}/chat/${selected}`
				: `/admin/games/${view.game.id}/chat`}
		{openConversation}
	/>
{/if}
