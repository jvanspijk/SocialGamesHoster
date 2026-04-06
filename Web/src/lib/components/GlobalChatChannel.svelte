<script lang="ts">
	import { GetGlobalChat } from '$lib/client/GameSessions/GetGlobalChat';
	import { GetMessagesFromChannel } from '$lib/client/Chat/GetMessagesFromChannel';
	import ChatChannel, { type Message } from './ChatChannel.svelte';

	let {
		gameId,
		readerId,
		isOpen = false,
		onClose
	} = $props<{
		gameId: number;
		readerId: number;
		isOpen?: boolean;
		onClose: () => void;
	}>();

	let channelId = $state<number | null>(null);
	let channelName = $state('Global Chat');
	let initialMessages = $state<Message[]>([]);
	let isReady = $state(false);

	function getAnonymizedSender(senderId: number | null): string {
		if (senderId === null) return 'Unknown';
		if (senderId === readerId) return 'You';

		return `Player ${senderId}`;
	}

	function transformSender(
		senderId: number | null,
		_senderName: string | null,
		isAdmin?: boolean
	): string {
		if (isAdmin) return 'Admin';
		return getAnonymizedSender(senderId);
	}

	async function loadChatData() {
		if (!gameId) return;

		const chatRes = await GetGlobalChat(fetch, { gameId });
		if (!chatRes.ok) {
			console.error('Failed to get global chat:', chatRes.error);
			return;
		}

		channelId = chatRes.data.channelId;
		channelName = chatRes.data.channelName;

		if (channelId) {
			const messagesRes = await GetMessagesFromChannel(fetch, {
				channelId: channelId,
				before: null,
				after: null,
				maxMessages: 50
			});

			if (messagesRes.ok) {
				// Transform messages with anonymized senders
				initialMessages = messagesRes.data.map((m) => ({
					id: m.id,
					text: m.content,
					sender: transformSender(m.senderId, m.senderName, m.isAdmin),
					isMe: readerId === 0 ? Boolean(m.isAdmin) : m.senderId === readerId,
					sentAt: m.sentAt,
					isAdmin: m.isAdmin,
					isDeleted: m.isDeleted
				}));
			}
		}

		isReady = true;
	}

	$effect.pre(() => {
		loadChatData();
	});
</script>

{#if isReady && channelId !== null}
	<ChatChannel
		{channelId}
		{channelName}
		{readerId}
		{isOpen}
		{onClose}
		{initialMessages}
		{transformSender}
	/>
{/if}
