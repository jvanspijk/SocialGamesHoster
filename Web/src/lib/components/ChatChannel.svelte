<script lang="ts">
	import TextInput from './TextInput.svelte';
	import MainButton from './MainButton.svelte';
	import { chatHub } from '$lib/client/Chat/ChatHub.svelte';
	import { GetMessage } from '$lib/client/Chat/GetMessage';
	import { GetMessagesFromChannel } from '$lib/client/Chat/GetMessagesFromChannel';
	import { onMount } from 'svelte';
	import { SendMessage } from '$lib/client/Chat/SendMessage';

	export interface Message {
		id: number;
		text: string;
		sender: string;
		isMe: boolean;
		time: string;
	}

	let {
		channelId,
		channelName = 'Global Chat',
		readerId,
		isOpen = false,
		onClose,
		initialMessages = [],
		transformSender
	} = $props<{
		channelId: number;
		channelName?: string;
		readerId: number;
		isOpen?: boolean;
		onClose: () => void;
		initialMessages?: Message[];
		transformSender?: (senderId: number | null, senderName: string | null) => string;
	}>();

	let newMessage = $state('');
	let messages: Message[] = $state([]);
	let scrollContainer = $state<HTMLElement | null>(null);
	let isLoading = $state(true);

	function formatTime(dateStr?: string) {
		const date = dateStr ? new Date(dateStr) : new Date();
		return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}

	function getSenderName(senderId: number | null, senderName: string | null): string {
		if (transformSender) {
			return transformSender(senderId, senderName);
		}
		return senderName ?? 'Unknown';
	}

	onMount(async () => {
		// Load initial messages if provided, otherwise fetch from API
		if (initialMessages.length > 0) {
			messages = [...initialMessages];
			isLoading = false;
		} else {
			const response = await GetMessagesFromChannel(fetch, {
				channelId,
				before: null,
				after: null,
				maxMessages: 50
			});

			if (response.ok) {
				messages = response.data.map((m) => ({
					id: m.id,
					text: m.content,
					sender: getSenderName(m.senderId, m.senderName),
					isMe: m.senderId === readerId,
					time: formatTime(m.sentAt)
				}));
			}
			isLoading = false;
		}

		// Subscribe to real-time messages
		const unsubscribe = chatHub.onEvent('MessageSent', async (event) => {
			console.debug(`Message received: ${JSON.stringify(event)}`);
			if (event.channelId !== channelId) return;
			if (event.senderId === readerId) return;

			const response = await GetMessage(fetch, { id: event.messageId });

			if (response.ok) {
				const data = response.data;
				if (!messages.some((m) => m.id === data.id)) {
					messages.push({
						id: data.id,
						text: data.content,
						sender: getSenderName(data.senderId, data.senderName),
						isMe: data.senderId === readerId,
						time: formatTime(data.sentAt)
					});
				}
			}
		});

		return unsubscribe;
	});

	async function sendMessage() {
		if (!newMessage.trim()) return;

		const textToSend = newMessage;
		const timestamp = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

		try {
			const res = await SendMessage(fetch, { channelId, playerId: readerId, message: textToSend });
			if (!res.ok) {
				throw new Error(res.error?.detail);
			}
			messages.push({
				id: res.data.messageId,
				text: textToSend,
				sender: 'You',
				isMe: true,
				time: timestamp
			});
			newMessage = '';
		} catch (err) {
			console.error('Failed to send message:', err);
		}
	}

	$effect(() => {
		if (isOpen || messages.length) {
			if (scrollContainer) {
				scrollContainer.scrollTo({
					top: scrollContainer.scrollHeight,
					behavior: 'smooth'
				});
			}
		}
	});

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') onClose();
	}
</script>

<svelte:window on:keydown={handleKeydown} />

{#if isOpen}
	<div class="modal-overlay" onclick={(e) => e.target === e.currentTarget && onClose()} role="none">
		<div class="chat-modal">
			<header>
				<div class="channel-info">
					<span class="status-dot"></span>
					<h2>{channelName}</h2>
				</div>
				<button class="close-btn" onclick={onClose}>✕</button>
			</header>

			<div class="message-container" bind:this={scrollContainer}>
				{#if isLoading}
					<div class="loading-overlay">Loading messages...</div>
				{:else}"
					{#each messages as msg (msg.id)}
						<div class="message-wrapper {msg.isMe ? 'me' : 'them'}">
							<div class="bubble">
								{#if !msg.isMe}<span class="sender">{msg.sender}</span>{/if}
								<p>{msg.text}</p>
								<span class="timestamp">{msg.time}</span>
							</div>
						</div>
					{/each}
				{/if}
			</div>

			<footer>
				<div class="input-wrapper">
					<TextInput
						bind:value={newMessage}
						placeholder="Type a message..."
						onkeydown={(e: KeyboardEvent) => e.key === 'Enter' && sendMessage()}
					/>
				</div>
				<div class="send-wrapper">
					<MainButton label="Send" onActivate={sendMessage} />
				</div>
			</footer>
		</div>
	</div>
{/if}

<style>
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100vw;
		height: 100vh;
		background: rgba(0, 0, 0, 0.85);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 2000;
		backdrop-filter: blur(4px);
	}

	.chat-modal {
		width: 95vw;
		max-width: 600px;
		height: 85vh;
		background-color: var(--color-surface-soft); /* Parchment Color */
		border: 4px solid var(--color-border);
		border-radius: 8px;
		display: flex;
		flex-direction: column;
		overflow: hidden;
		box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
	}

	header {
		background: var(--color-text);
		padding: 15px 20px;
		display: flex;
		justify-content: space-between;
		align-items: center;
		border-bottom: 3px solid var(--color-border);
		color: var(--color-surface);
		font-family: var(--font-heading);
	}

	.channel-info {
		display: flex;
		align-items: center;
		gap: 12px;
	}

	.status-dot {
		width: 10px;
		height: 10px;
		background: var(--color-success);
		border-radius: 50%;
		box-shadow: 0 0 8px var(--color-success);
	}

	h2 {
		margin: 0;
		font-size: 1.2rem;
		text-transform: uppercase;
		letter-spacing: 1px;
	}

	.close-btn {
		background: none;
		border: none;
		color: var(--color-surface);
		font-size: 1.5rem;
		cursor: pointer;
		transition: color 0.2s;
	}

	.close-btn:hover {
		color: var(--color-accent);
	}

	.message-container {
		flex: 1;
		overflow-y: auto;
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 6px;
		background-image: radial-gradient(var(--color-border) 0.5px, transparent 0.5px);
		background-size: 20px 20px;
		background-color: var(--color-surface-soft);
	}

	.message-wrapper {
		display: flex;
		width: 100%;
	}

	.message-wrapper.me {
		justify-content: flex-end;
	}
	.message-wrapper.them {
		justify-content: flex-start;
	}

	.bubble {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		max-width: 85%;
		padding: 4px 12px;
		border-radius: 12px;
		position: relative;
		font-family: var(--font-body);
		font-size: 1rem;
		border: 2px solid var(--color-border);
		box-shadow: 2px 2px 0px rgba(0, 0, 0, 0.1);
		line-height: 1.25;
	}

	.bubble p {
		margin: 0;
	}

	.me .bubble {
		background: var(--color-surface);
		color: var(--color-text);
		border-bottom-right-radius: 2px;
	}

	.them .bubble {
		background: var(--color-on-accent);
		color: var(--color-text);
		border-bottom-left-radius: 2px;
	}

	.sender {
		display: block;
		font-size: 0.9rem;
		font-weight: bold;
		color: var(--color-accent);
		margin-bottom: 2px;
		text-transform: uppercase;
	}

	.timestamp {
		display: block;
		font-size: 1rem;
		text-align: right;
		margin-top: 1px;
		opacity: 0.7;
	}

	footer {
		padding: 15px;
		background: var(--color-text);
		border-top: 3px solid var(--color-border);
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.input-wrapper {
		flex: 1;
	}

	:global(.input-wrapper input) {
		margin: 0 !important;
	}

	.send-wrapper :global(button) {
		margin: 0 !important;
		padding: 8px 20px !important;
	}
</style>
