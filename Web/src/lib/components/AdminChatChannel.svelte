<script lang="ts">
	import { onMount } from 'svelte';
	import { GetGlobalChat } from '$lib/client/GameSessions/GetGlobalChat';
	import { GetMessagesFromChannel } from '$lib/client/Chat/GetMessagesFromChannel';
	import ChatChannel, { type Message } from './ChatChannel.svelte';

	/**
	 * Admin chat channel wrapper for viewing the global chat.
	 * Unlike GlobalChatChannel, this shows real player names (no anonymization).
	 * Use readerId=0 for admin viewers who are not participants.
	 */
	let {
		gameId,
		readerId = 0,
		isOpen = false,
		onClose
	} = $props<{
		gameId: number;
		readerId?: number;
		isOpen?: boolean;
		onClose: () => void;
	}>();

	let channelId = $state<number | null>(null);
	let channelName = $state('Global Chat');
	let initialMessages = $state<Message[]>([]);
	let isReady = $state(false);

	function formatTime(dateStr?: string) {
		const date = dateStr ? new Date(dateStr) : new Date();
		return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}

	// No transformation - show real sender names
	function transformSender(_senderId: number | null, senderName: string | null): string {
		return senderName ?? 'Unknown' + '(' + _senderId + ')';
	}

	onMount(async () => {
		// 1. Get global chat channel info
		const chatRes = await GetGlobalChat(fetch, { gameId });
		if (!chatRes.ok) {
			console.error('Failed to get global chat:', chatRes.error);
			return;
		}

		channelId = chatRes.data.channelId;
		channelName = chatRes.data.channelName;

		// 2. Get recent messages
		const messagesRes = await GetMessagesFromChannel(fetch, {
			channelId: channelId,
			before: null,
			after: null,
			maxMessages: 50
		});

		if (messagesRes.ok) {
			// Transform messages with real sender names
			initialMessages = messagesRes.data.map((m) => ({
				id: m.id,
				text: m.content,
				sender: m.senderName ?? 'Unknown',
				isMe: m.senderId === readerId,
				time: formatTime(m.sentAt)
			}));
		}

		isReady = true;
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
