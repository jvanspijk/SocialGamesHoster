<script lang="ts">
	import SelectionDialog, {
		type SelectionDialogEntry
	} from '$lib/components/SelectionDialog.svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import type { Pathname } from '$app/types';
	import ChatApp from './ChatApp.svelte';

	let {
		gameId,
		roomId = '',
		canModerate = false,
		archived = false,
		policyRevision = '',
		recipientEntries,
		roomPath,
		openConversation
	}: DirectMessageChatPageProps = $props();

	interface DirectMessageChatPageProps {
		gameId: string;
		roomId?: string;
		canModerate?: boolean;
		archived?: boolean;
		policyRevision?: string;
		recipientEntries: readonly SelectionDialogEntry[];
		roomPath: (roomId: string) => Pathname;
		openConversation: (participantId: string) => string | undefined | Promise<string | undefined>;
	}

	let newMessageOpen = $state(false);

	async function selectRoom(roomId: string) {
		await goto(resolve(roomPath(roomId)));
	}

	async function chooseRecipient(participantId: string) {
		const conversationId = await openConversation(participantId);
		if (!conversationId) return;
		newMessageOpen = false;
		await selectRoom(conversationId);
	}
</script>

<ChatApp
	{gameId}
	selectedRoomId={roomId}
	{canModerate}
	{archived}
	{policyRevision}
	{selectRoom}
	newMessage={() => (newMessageOpen = true)}
/>

<SelectionDialog
	open={newMessageOpen}
	title="New message"
	description="Choose a player to open their direct conversation."
	close={() => (newMessageOpen = false)}
	emptyState={{
		title: 'No players available',
		description: 'No players are available for a direct conversation.'
	}}
	entries={recipientEntries}
	onselect={chooseRecipient}
/>
