<script lang="ts">
	import { onMount } from 'svelte';
	import { Send, Trash2 } from '@lucide/svelte';
	import { api, AppApiError, jsonBody, pb } from '$lib/api/client';
	import type { AppErrorBody, ChatMessage, RealtimeEnvelope, Room } from '$lib/api/types';
	import Button from './Button.svelte';
	import ErrorNotice from './ErrorNotice.svelte';

	let { room, canModerate = false }: { room: Room; canModerate?: boolean } = $props();
	let messages = $state<ChatMessage[]>([]);
	let content = $state('');
	let nextCursor = $state('');
	let error = $state<AppErrorBody | null>(null);
	let sending = $state(false);

	onMount(() => {
		let unsubscribe: (() => void) | undefined;
		let disposed = false;
		void load();
		void pb.realtime
			.subscribe(`room:${room.id}`, (raw) => {
				const event = raw as unknown as RealtimeEnvelope<ChatMessage>;
				if (event.kind === 'chat.message_created') {
					if (!messages.some((item) => item.id === event.payload.id))
						messages = [...messages, event.payload];
				}
				if (event.kind === 'chat.message_deleted') {
					messages = messages.map((item) => (item.id === event.payload.id ? event.payload : item));
				}
			})
			.then((cancel) => {
				if (disposed) cancel();
				else unsubscribe = cancel;
			});
		return () => {
			disposed = true;
			unsubscribe?.();
		};
	});

	async function load(cursor = '') {
		try {
			const page = await api<{ items: ChatMessage[]; nextCursor: string }>(
				`/rooms/${room.id}/messages${cursor ? `?cursor=${encodeURIComponent(cursor)}` : ''}`
			);
			messages = cursor ? [...page.items.reverse(), ...messages] : page.items.reverse();
			nextCursor = page.nextCursor;
		} catch (caught) {
			setError(caught);
		}
	}

	async function send(event: SubmitEvent) {
		event.preventDefault();
		const outgoing = content.trim();
		if (!outgoing) return;
		sending = true;
		error = null;
		content = '';
		try {
			const message = await api<ChatMessage>(`/rooms/${room.id}/messages`, {
				method: 'POST',
				...jsonBody({ content: outgoing })
			});
			if (!messages.some((item) => item.id === message.id)) messages = [...messages, message];
		} catch (caught) {
			content = outgoing;
			setError(caught);
		} finally {
			sending = false;
		}
	}

	async function remove(message: ChatMessage) {
		error = null;
		try {
			const deleted = await api<ChatMessage>(`/rooms/${room.id}/messages/${message.id}`, {
				method: 'DELETE'
			});
			messages = messages.map((item) => (item.id === deleted.id ? deleted : item));
		} catch (caught) {
			setError(caught);
		}
	}

	function setError(caught: unknown) {
		error =
			caught instanceof AppApiError
				? caught.body
				: { code: 'network.failed', message: 'Chat is temporarily unavailable.' };
	}
</script>

<section class="chat">
	<header>
		<div>
			<p class="kind">{room.kind.replace('_', ' ')}</p>
			<h2>{room.label}</h2>
		</div>
		{#if room.locked}<span class="locked">Read only</span>{/if}
	</header>
	<div class="messages" aria-live="polite">
		{#if nextCursor}
			<button class="older" onclick={() => load(nextCursor)}>Load earlier messages</button>
		{/if}
		{#each messages as message (message.id)}
			<article class:own={message.isOwn} class:deleted={message.deleted}>
				<div>
					<strong>{message.senderLabel}</strong><time
						>{new Date(message.createdAt).toLocaleTimeString([], {
							hour: '2-digit',
							minute: '2-digit'
						})}</time
					>
				</div>
				<p>{message.deleted ? 'Message removed' : message.content}</p>
				{#if !message.deleted && (canModerate || message.isOwn)}
					<button
						class="remove"
						aria-label={`Remove message from ${message.senderLabel}`}
						onclick={() => remove(message)}
					>
						<Trash2 size={14} /> Remove
					</button>
				{/if}
			</article>
		{:else}
			<p class="empty">No messages have been written here yet.</p>
		{/each}
	</div>
	<ErrorNotice {error} />
	{#if room.sendable && !room.locked}
		<form onsubmit={send}>
			<label class="sr-only" for="message">Message</label>
			<input
				id="message"
				bind:value={content}
				maxlength="1000"
				autocomplete="off"
				placeholder="Write a message…"
			/>
			<Button type="submit" loading={sending} disabled={!content.trim()}
				><Send size={17} /> Send</Button
			>
		</form>
	{:else}
		<p class="read-only">This room is read only.</p>
	{/if}
</section>

<style>
	.chat {
		display: grid;
		height: 100%;
		min-height: 20rem;
		grid-template-rows: auto minmax(12rem, 1fr) auto auto;
	}

	header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		border-bottom: 1px solid #b99b6c;
	}

	.kind {
		margin: 0;
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.62rem;
		letter-spacing: 0.12em;
		text-transform: uppercase;
	}

	.locked,
	.read-only {
		color: var(--ink-soft);
		font-size: 0.8rem;
	}

	.messages {
		display: flex;
		overflow-y: auto;
		flex-direction: column;
		gap: 0.65rem;
		padding-block: 0.9rem;
	}

	article {
		max-width: 88%;
		align-self: flex-start;
		border-inline-start: 3px solid #b69a6c;
		background: rgb(255 249 230 / 60%);
		padding: 0.5rem 0.65rem;
	}

	article.own {
		align-self: flex-end;
		border-inline-start: 0;
		border-inline-end: 3px solid var(--crimson);
	}

	article.deleted {
		opacity: 0.65;
		font-style: italic;
	}

	article div {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		font-size: 0.75rem;
	}

	article p {
		margin: 0.15rem 0 0;
		white-space: pre-wrap;
	}

	.remove {
		display: inline-flex;
		align-items: center;
		gap: 0.2rem;
		border: 0;
		background: transparent;
		color: var(--danger);
		cursor: pointer;
		font-size: 0.72rem;
		padding: 0.2rem 0 0;
		text-decoration: underline;
	}

	time {
		color: var(--ink-faint);
	}

	form {
		position: sticky;
		inset-block-end: 0;
		display: grid;
		grid-template-columns: 1fr auto;
		gap: 0.5rem;
		background: var(--paper);
		padding-block-end: max(0.25rem, env(safe-area-inset-bottom));
	}

	input {
		min-width: 0;
		min-height: 44px;
		border: 1px solid #8d7248;
		background: var(--paper-light);
		padding: 0.65rem;
	}

	.older {
		align-self: center;
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		text-decoration: underline;
	}

	.empty {
		margin: auto;
		color: var(--ink-faint);
		text-align: center;
	}
</style>
