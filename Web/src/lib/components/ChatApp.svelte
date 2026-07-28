<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import {
		ArrowLeft,
		Hash,
		MessageCircle,
		Plus,
		Search,
		Send,
		Shield,
		Trash2,
		Users
	} from '@lucide/svelte';
	import Button from './Button.svelte';
	import { api, jsonBody, pb } from '$lib/api/client';
	import type { ChatMessage, MessageSummary, RealtimeEnvelope, Room } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { cursorIsAfter, readMarkerStorageKey } from '$lib/state/chatReadMarkers';
	import { toasts } from '$lib/state/toasts.svelte';

	let {
		gameId,
		selectedRoomId = '',
		canModerate = false,
		archived = false,
		policyRevision = '',
		selectRoom,
		newMessage
	}: {
		gameId: string;
		selectedRoomId?: string;
		canModerate?: boolean;
		archived?: boolean;
		policyRevision?: string;
		selectRoom: (roomId: string) => void;
		newMessage?: () => void;
	} = $props();

	type Marker = Pick<MessageSummary, 'id' | 'createdAt'>;

	let rooms = $state<Room[]>([]);
	let messages = $state<ChatMessage[]>([]);
	let search = $state('');
	let content = $state('');
	let nextCursor = $state('');
	let loadingRooms = $state(true);
	let loadingMessages = $state(false);
	let sending = $state(false);
	let firstUnreadId = $state('');
	let messagesElement = $state<HTMLDivElement>();
	let subscriptions: Array<() => void> = [];
	let markers = $state<Record<string, Marker>>({});
	let loadedPolicyRevision = '';
	const scrollPositions = new SvelteMap<string, number>();

	const selectedRoom = $derived(rooms.find((room) => room.id === selectedRoomId) ?? null);
	const filteredRooms = $derived.by(() => {
		const query = search.trim().toLocaleLowerCase();
		return rooms
			.filter((room) => !query || room.label.toLocaleLowerCase().includes(query))
			.toSorted((left, right) => {
				const leftTime = left.latestMessage ? new Date(left.latestMessage.createdAt).getTime() : 0;
				const rightTime = right.latestMessage
					? new Date(right.latestMessage.createdAt).getTime()
					: 0;
				return rightTime - leftTime || left.label.localeCompare(right.label);
			});
	});

	onMount(() => {
		loadMarkers();
		void loadRooms();
		return () => {
			for (const unsubscribe of subscriptions) unsubscribe();
		};
	});

	$effect(() => {
		if (selectedRoomId && rooms.some((room) => room.id === selectedRoomId)) {
			void openConversation(selectedRoomId);
		}
	});

	$effect(() => {
		if (!policyRevision || !loadedPolicyRevision) {
			loadedPolicyRevision = policyRevision;
			return;
		}
		if (policyRevision !== loadedPolicyRevision) {
			loadedPolicyRevision = policyRevision;
			void loadRooms();
		}
	});

	function loadMarkers() {
		try {
			markers = JSON.parse(
				localStorage.getItem(readMarkerStorageKey(auth.actor?.id ?? '', gameId)) ?? '{}'
			);
		} catch {
			markers = {};
		}
	}

	function saveMarkers() {
		localStorage.setItem(
			readMarkerStorageKey(auth.actor?.id ?? '', gameId),
			JSON.stringify(markers)
		);
	}

	async function loadRooms() {
		loadingRooms = true;
		try {
			rooms = await api<Room[]>(`/games/${gameId}/rooms`);
			if (selectedRoomId && !rooms.some((room) => room.id === selectedRoomId)) {
				selectRoom('');
			}
			await subscribeRooms();
			if (selectedRoomId) await openConversation(selectedRoomId);
		} catch (caught) {
			toasts.error(
				caught instanceof Error ? caught.message : 'Conversations could not be loaded.',
				{
					actionLabel: 'Retry',
					action: loadRooms,
					persistent: true
				}
			);
		} finally {
			loadingRooms = false;
		}
	}

	async function subscribeRooms() {
		for (const unsubscribe of subscriptions) unsubscribe();
		subscriptions = [];
		for (const room of rooms) {
			const unsubscribe = await pb.realtime.subscribe(`room:${room.id}`, (raw) => {
				const event = raw as unknown as RealtimeEnvelope<ChatMessage | Room>;
				if (event.kind === 'chat.message_created') {
					const message = event.payload as ChatMessage;
					if (
						message.roomId === selectedRoomId &&
						!messages.some((item) => item.id === message.id)
					) {
						messages = [...messages, message];
						void tick().then(() => scrollToBottom());
						markRead(room.id, { id: message.id, createdAt: message.createdAt });
					}
					updateRoomMessage(room.id, message);
				} else if (event.kind === 'chat.message_deleted') {
					const deleted = event.payload as ChatMessage;
					if (deleted.roomId === selectedRoomId) {
						messages = messages.map((item) => (item.id === deleted.id ? deleted : item));
					}
					updateRoomMessage(room.id, deleted);
				} else if (event.kind === 'chat.room_updated') {
					const updated = event.payload as Room;
					rooms = rooms.map((item) => (item.id === updated.id ? { ...item, ...updated } : item));
				}
			});
			subscriptions.push(unsubscribe);
		}
	}

	function updateRoomMessage(roomId: string, message: ChatMessage) {
		const summary: MessageSummary = {
			id: message.id,
			createdAt: message.createdAt,
			senderLabel: message.senderLabel,
			preview: message.deleted ? 'Message removed' : message.content.slice(0, 120)
		};
		rooms = rooms.map((room) => (room.id === roomId ? { ...room, latestMessage: summary } : room));
	}

	async function openConversation(roomId: string) {
		if (selectedRoomId && messagesElement) {
			scrollPositions.set(selectedRoomId, messagesElement.scrollTop);
		}
		const room = rooms.find((item) => item.id === roomId);
		if (!room) return;
		loadingMessages = true;
		firstUnreadId = '';
		try {
			const page = await api<{ items: ChatMessage[]; nextCursor: string }>(
				`/rooms/${roomId}/messages`
			);
			messages = page.items.reverse();
			nextCursor = page.nextCursor;
			const previousMarker = markers[roomId];
			firstUnreadId =
				messages.find((message) =>
					cursorIsAfter({ id: message.id, createdAt: message.createdAt }, previousMarker)
				)?.id ?? '';
			if (room.latestMessage) markRead(roomId, room.latestMessage);
			await tick();
			const savedPosition = scrollPositions.get(roomId);
			if (savedPosition !== undefined && messagesElement) {
				messagesElement.scrollTop = savedPosition;
			} else if (firstUnreadId) {
				document.getElementById(`unread-${firstUnreadId}`)?.scrollIntoView({ block: 'center' });
			} else {
				scrollToBottom();
			}
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'Messages could not be loaded.');
		} finally {
			loadingMessages = false;
		}
	}

	function markRead(roomId: string, marker: Marker) {
		markers = { ...markers, [roomId]: marker };
		saveMarkers();
	}

	function unread(room: Room) {
		return cursorIsAfter(room.latestMessage, markers[room.id]);
	}

	async function loadEarlier() {
		if (!selectedRoom || !nextCursor) return;
		const previousHeight = messagesElement?.scrollHeight ?? 0;
		const page = await api<{ items: ChatMessage[]; nextCursor: string }>(
			`/rooms/${selectedRoom.id}/messages?cursor=${encodeURIComponent(nextCursor)}`
		);
		messages = [...page.items.reverse(), ...messages];
		nextCursor = page.nextCursor;
		await tick();
		if (messagesElement) messagesElement.scrollTop += messagesElement.scrollHeight - previousHeight;
	}

	async function send(event: SubmitEvent) {
		event.preventDefault();
		if (!selectedRoom || !content.trim()) return;
		const outgoing = content.trim();
		content = '';
		sending = true;
		try {
			const message = await api<ChatMessage>(`/rooms/${selectedRoom.id}/messages`, {
				method: 'POST',
				...jsonBody({ content: outgoing })
			});
			message.isOwn = true;
			if (!messages.some((item) => item.id === message.id)) messages = [...messages, message];
			updateRoomMessage(selectedRoom.id, message);
			markRead(selectedRoom.id, { id: message.id, createdAt: message.createdAt });
			await tick();
			scrollToBottom();
		} catch (caught) {
			content = outgoing;
			toasts.error(caught instanceof Error ? caught.message : 'The message could not be sent.');
		} finally {
			sending = false;
		}
	}

	async function remove(message: ChatMessage) {
		if (!selectedRoom) return;
		try {
			const deleted = await api<ChatMessage>(`/rooms/${selectedRoom.id}/messages/${message.id}`, {
				method: 'DELETE'
			});
			messages = messages.map((item) => (item.id === deleted.id ? deleted : item));
			updateRoomMessage(selectedRoom.id, deleted);
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'The message could not be removed.');
		}
	}

	async function togglePosting() {
		if (!selectedRoom) return;
		try {
			const updated = await api<Room>(`/rooms/${selectedRoom.id}`, {
				method: 'PATCH',
				...jsonBody({ playersCanPost: !selectedRoom.playersCanPost })
			});
			rooms = rooms.map((room) => (room.id === updated.id ? updated : room));
			toasts.success(updated.playersCanPost ? 'Players can post.' : 'Players are read-only.');
		} catch (caught) {
			toasts.error(
				caught instanceof Error ? caught.message : 'Posting access could not be updated.'
			);
		}
	}

	function scrollToBottom() {
		if (messagesElement) messagesElement.scrollTop = messagesElement.scrollHeight;
	}

	function roomType(room: Room) {
		if (room.kind === 'gm_dm' || room.kind === 'player_dm') return 'Direct';
		if (room.kind === 'team') return 'Team';
		if (room.kind === 'custom') return 'Ruleset channel';
		return 'Game';
	}

	function timeLabel(value?: string) {
		if (!value) return '';
		const date = new Date(value);
		const today = new Date();
		return date.toDateString() === today.toDateString()
			? date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
			: date.toLocaleDateString([], { month: 'short', day: 'numeric' });
	}

	function dayLabel(value: string) {
		const date = new Date(value);
		const today = new Date();
		if (date.toDateString() === today.toDateString()) return 'Today';
		const yesterday = new Date(today.getTime() - 86_400_000);
		if (date.toDateString() === yesterday.toDateString()) return 'Yesterday';
		return date.toLocaleDateString([], { weekday: 'long', month: 'long', day: 'numeric' });
	}

	function startsNewDay(index: number) {
		if (index === 0) return true;
		return (
			new Date(messages[index - 1].createdAt).toDateString() !==
			new Date(messages[index].createdAt).toDateString()
		);
	}
</script>

<div class="chat-app" class:conversation-open={selectedRoom}>
	<aside class="conversation-rail">
		<header>
			<div>
				<p>Game chat</p>
				<h1>Conversations</h1>
			</div>
			{#if newMessage}
				<button type="button" aria-label="New message" onclick={newMessage}
					><Plus size={22} /></button
				>
			{/if}
		</header>
		<label class="search">
			<Search size={18} aria-hidden="true" />
			<span class="sr-only">Search conversations</span>
			<input bind:value={search} placeholder="Search conversations" />
		</label>
		<div class="conversation-list">
			{#if loadingRooms}
				<p class="rail-status" role="status">Loading conversations…</p>
			{:else if filteredRooms.length === 0}
				<div class="rail-empty">
					<MessageCircle size={30} />
					<h2>No conversations</h2>
					<p>
						{search
							? 'No conversations match your search.'
							: 'Conversations appear when chat is available.'}
					</p>
					{#if newMessage && !search}<button type="button" onclick={newMessage}>New message</button
						>{/if}
				</div>
			{:else}
				{#each filteredRooms as room (room.id)}
					<button
						type="button"
						class:selected={selectedRoomId === room.id}
						class:unread={unread(room)}
						aria-current={selectedRoomId === room.id ? 'page' : undefined}
						onclick={() => selectRoom(room.id)}
					>
						<span class="room-avatar">
							{#if room.kind === 'team'}<Users size={20} />{:else if room.kind === 'general'}<Hash
									size={20}
								/>{:else}<span>{room.label.slice(0, 1).toUpperCase()}</span>{/if}
						</span>
						<span class="room-copy">
							<span class="room-title">
								<strong>{room.label}</strong>
								<time>{timeLabel(room.latestMessage?.createdAt)}</time>
							</span>
							<span class="preview">
								{room.latestMessage
									? `${room.latestMessage.senderLabel}: ${room.latestMessage.preview}`
									: 'No messages yet'}
							</span>
							<small>{roomType(room)}</small>
						</span>
						{#if unread(room)}<i aria-label="New messages"></i>{/if}
					</button>
				{/each}
			{/if}
		</div>
	</aside>

	<section class="conversation">
		{#if selectedRoom}
			<header class="conversation-header">
				<button
					class="back"
					type="button"
					aria-label="Back to conversations"
					onclick={() => selectRoom('')}
				>
					<ArrowLeft size={21} />
				</button>
				<span class="conversation-avatar">{selectedRoom.label.slice(0, 1).toUpperCase()}</span>
				<div>
					<h2>{selectedRoom.label}</h2>
					<p>
						{roomType(selectedRoom)}
						{#if !selectedRoom.playersCanPost}
							· Players read-only{/if}
					</p>
				</div>
				{#if canModerate && !archived}
					<label class="posting-toggle">
						<input type="checkbox" checked={selectedRoom.playersCanPost} onchange={togglePosting} />
						Players can post
					</label>
				{/if}
			</header>
			<div class="messages" bind:this={messagesElement} aria-live="polite">
				{#if nextCursor}
					<button class="older" type="button" onclick={loadEarlier}>Load earlier messages</button>
				{/if}
				{#if loadingMessages}
					<p class="message-status" role="status">Loading messages…</p>
				{:else if messages.length === 0}
					<div class="message-empty">
						<MessageCircle size={36} strokeWidth={1.4} />
						<h3>No messages yet</h3>
						<p>
							{selectedRoom.sendable && !archived
								? 'Start the conversation.'
								: 'This conversation is read-only.'}
						</p>
					</div>
				{:else}
					{#each messages as message, index (message.id)}
						{#if startsNewDay(index)}
							<div class="day-divider"><span>{dayLabel(message.createdAt)}</span></div>
						{/if}
						{#if message.id === firstUnreadId}
							<div class="unread-divider" id={`unread-${message.id}`}>
								<span>New messages</span>
							</div>
						{/if}
						<article class:own={message.isOwn} class:deleted={message.deleted}>
							<div class="message-meta">
								<strong>{message.senderLabel}</strong>
								<time>{timeLabel(message.createdAt)}</time>
							</div>
							<p>{message.deleted ? 'Message removed' : message.content}</p>
							{#if !message.deleted && (canModerate || message.isOwn)}
								<button
									type="button"
									aria-label={`Remove message from ${message.senderLabel}`}
									onclick={() => remove(message)}
								>
									<Trash2 size={14} /> Remove
								</button>
							{/if}
						</article>
					{/each}
				{/if}
			</div>
			{#if !archived && selectedRoom.sendable}
				<form onsubmit={send}>
					<label class="sr-only" for="chat-message">Message</label>
					{#if selectedRoom.messageRestriction === 'emoji_only'}
						<p class="message-restriction" id="message-restriction">Emoji only</p>
					{/if}
					<textarea
						id="chat-message"
						bind:value={content}
						aria-describedby={selectedRoom.messageRestriction === 'emoji_only'
							? 'message-restriction'
							: undefined}
						maxlength="1000"
						rows="1"
						placeholder={selectedRoom.messageRestriction === 'emoji_only'
							? 'Add emoji'
							: 'Write a message'}
						onkeydown={(event) => {
							if (event.key === 'Enter' && !event.shiftKey) {
								event.preventDefault();
								event.currentTarget.form?.requestSubmit();
							}
						}}
					></textarea>
					<Button type="submit" loading={sending} disabled={!content.trim()}
						><Send size={18} /> Send</Button
					>
				</form>
			{:else}
				<div class="read-only">
					<Shield size={17} />
					{archived ? 'Archived chat is read-only' : 'You cannot post in this conversation'}
				</div>
			{/if}
		{:else}
			<div class="conversation-placeholder">
				<MessageCircle size={48} strokeWidth={1.3} />
				<h2>Select a conversation</h2>
				<p>Choose a conversation from the list.</p>
			</div>
		{/if}
	</section>
</div>

<style>
	.chat-app {
		display: grid;
		height: calc(100dvh - 7.25rem);
		min-height: 32rem;
		grid-template-columns: minmax(18rem, 0.34fr) minmax(0, 1fr);
		overflow: hidden;
		border: 1px solid var(--gold-dark);
		background: var(--paper);
		box-shadow: var(--shadow-small);
	}

	.conversation-rail {
		display: grid;
		min-width: 0;
		grid-template-rows: auto auto minmax(0, 1fr);
		border-inline-end: 1px solid var(--gold-dark);
		background: linear-gradient(rgb(27 18 12 / 96%), rgb(18 11 8 / 98%)), var(--wood);
		color: var(--paper-light);
	}

	.conversation-rail > header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--space-4);
	}

	.conversation-rail h1,
	.conversation-rail header p {
		margin: 0;
		color: var(--paper-light);
	}

	.conversation-rail h1 {
		font-size: 1.35rem;
	}

	.conversation-rail header p {
		color: var(--gold-light);
		font-family: var(--font-display);
		font-size: 0.64rem;
		letter-spacing: 0.12em;
		text-transform: uppercase;
	}

	.conversation-rail header button {
		display: grid;
		width: var(--target-size);
		height: var(--target-size);
		place-items: center;
		border: 1px solid var(--gold-dark);
		background: transparent;
		color: var(--gold-light);
		cursor: pointer;
	}

	.search {
		display: grid;
		grid-template-columns: auto 1fr;
		align-items: center;
		gap: var(--space-2);
		margin: 0 var(--space-3) var(--space-3);
		border: 1px solid #755d43;
		background: rgb(255 255 255 / 7%);
		padding-inline: var(--space-2);
	}

	.search input {
		min-width: 0;
		min-height: var(--target-size);
		border: 0;
		background: transparent;
		color: var(--paper-light);
		outline: 0;
	}

	.search input::placeholder {
		color: var(--paper-muted);
	}

	.conversation-list {
		overflow-y: auto;
	}

	.conversation-list > button {
		position: relative;
		display: grid;
		width: 100%;
		min-height: 5.25rem;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: center;
		gap: var(--space-3);
		border: 0;
		border-block-end: 1px solid rgb(223 189 101 / 18%);
		background: transparent;
		color: var(--paper-muted);
		cursor: pointer;
		padding: var(--space-3);
		text-align: start;
	}

	.conversation-list > button:hover,
	.conversation-list > button.selected {
		background: rgb(207 85 85 / 13%);
	}

	.conversation-list > button.selected {
		box-shadow: inset 3px 0 var(--gold-light);
	}

	.room-avatar,
	.conversation-avatar {
		display: grid;
		width: 2.8rem;
		height: 2.8rem;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--crimson-dark);
		color: var(--gold-light);
		font-family: var(--font-display);
		font-weight: 700;
	}

	.room-title {
		display: flex;
		justify-content: space-between;
		gap: var(--space-2);
	}

	.room-title strong {
		overflow: hidden;
		color: var(--paper-light);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.room-title time,
	.room-copy small {
		color: var(--paper-muted);
		font-size: 0.7rem;
	}

	.preview,
	.room-copy small {
		display: block;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.preview {
		margin-block: 0.1rem;
		font-size: 0.82rem;
	}

	.conversation-list > button.unread .room-title strong,
	.conversation-list > button.unread .preview {
		color: var(--gold-light);
		font-weight: 700;
	}

	.conversation-list i {
		position: absolute;
		inset-inline-end: var(--space-3);
		inset-block-end: var(--space-3);
		width: 0.55rem;
		height: 0.55rem;
		border-radius: 50%;
		background: var(--crimson-light);
	}

	.rail-status,
	.rail-empty {
		color: var(--paper-muted);
		padding: var(--space-5);
		text-align: center;
	}

	.rail-empty h2 {
		color: var(--paper-light);
		font-size: 1rem;
	}

	.rail-empty button {
		min-height: var(--target-size);
		border: 0;
		background: transparent;
		color: var(--gold-light);
		cursor: pointer;
		font-family: var(--font-display);
		font-weight: 700;
	}

	.conversation {
		display: grid;
		min-width: 0;
		grid-template-rows: auto minmax(0, 1fr) auto;
		background:
			radial-gradient(circle at 15% 22%, rgb(112 73 35 / 6%) 0 0.08rem, transparent 0.11rem),
			var(--paper);
		background-size:
			1.8rem 2rem,
			auto;
	}

	.conversation-header {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		background: rgb(255 249 230 / 82%);
		padding: var(--space-3) var(--space-4);
	}

	.conversation-header .conversation-avatar {
		display: none;
	}

	.conversation-header h2,
	.conversation-header p {
		margin: 0;
	}

	.conversation-header h2 {
		font-size: 1.2rem;
	}

	.conversation-header p {
		color: var(--ink-soft);
		font-size: 0.8rem;
	}

	.back {
		display: none;
		width: var(--target-size);
		height: var(--target-size);
		place-items: center;
		border: 0;
		background: transparent;
		color: var(--ink);
		cursor: pointer;
	}

	.posting-toggle {
		display: flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-2);
		font-size: 0.82rem;
	}

	.messages {
		display: flex;
		min-height: 0;
		overflow-y: auto;
		flex-direction: column;
		gap: var(--space-2);
		padding: var(--space-4);
		overscroll-behavior: contain;
	}

	.messages article {
		width: fit-content;
		max-width: min(82%, 42rem);
		align-self: flex-start;
		border: 1px solid #bda574;
		border-radius: 0 0.65rem 0.65rem 0.65rem;
		background: var(--paper-light);
		box-shadow: var(--shadow-small);
		padding: var(--space-2) var(--space-3);
	}

	.messages article.own {
		align-self: flex-end;
		border-color: #9c7740;
		border-radius: 0.65rem 0 0.65rem 0.65rem;
		background: #ead3a7;
	}

	.messages article.deleted {
		box-shadow: none;
		opacity: 0.68;
		font-style: italic;
	}

	.message-meta {
		display: flex;
		justify-content: space-between;
		gap: var(--space-4);
		font-size: 0.74rem;
	}

	.message-meta strong {
		color: var(--crimson-dark);
	}

	.message-meta time {
		color: var(--ink-faint);
	}

	.messages article p {
		margin: 0.15rem 0 0;
		white-space: pre-wrap;
	}

	.messages article button,
	.older {
		display: inline-flex;
		min-height: 2rem;
		align-items: center;
		gap: 0.2rem;
		border: 0;
		background: transparent;
		color: var(--danger);
		cursor: pointer;
		font-size: 0.7rem;
		padding: 0;
	}

	.older {
		align-self: center;
		min-height: var(--target-size);
		color: var(--crimson-dark);
		text-decoration: underline;
	}

	.day-divider,
	.unread-divider {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		color: var(--ink-soft);
		font-size: 0.72rem;
		text-align: center;
	}

	.day-divider::before,
	.day-divider::after,
	.unread-divider::before,
	.unread-divider::after {
		height: 1px;
		flex: 1;
		background: #b9a170;
		content: '';
	}

	.unread-divider {
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-weight: 700;
	}

	.unread-divider::before,
	.unread-divider::after {
		background: var(--crimson);
	}

	.conversation form {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		gap: var(--space-2);
		border-block-start: var(--border-subtle);
		background: var(--paper-light);
		padding: var(--space-3);
		padding-block-end: max(var(--space-3), env(safe-area-inset-bottom));
	}

	.conversation textarea {
		min-width: 0;
		min-height: var(--target-size);
		max-height: 8rem;
		resize: vertical;
		border: var(--border-subtle);
		background: white;
		color: var(--ink);
		padding: var(--space-2);
	}

	.read-only {
		display: flex;
		min-height: 3.5rem;
		align-items: center;
		justify-content: center;
		gap: var(--space-2);
		border-block-start: var(--border-subtle);
		background: var(--paper-deep);
		color: var(--ink-soft);
	}

	.message-restriction {
		align-self: center;
		margin: 0;
		color: var(--ink-soft);
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}

	.conversation-placeholder,
	.message-empty {
		align-self: center;
		justify-self: center;
		padding: var(--space-6);
		text-align: center;
	}

	.conversation-placeholder {
		grid-row: 1 / -1;
	}

	.conversation-placeholder h2,
	.conversation-placeholder p,
	.message-empty h3,
	.message-empty p {
		margin: 0;
	}

	.message-status {
		margin: auto;
	}

	@media (max-width: 47.99rem) {
		.chat-app {
			height: calc(100dvh - 7.75rem - env(safe-area-inset-bottom));
			min-height: 0;
			grid-template-columns: 1fr;
			border-inline: 0;
		}

		.conversation-rail,
		.conversation {
			grid-column: 1;
			grid-row: 1;
		}

		.chat-app:not(.conversation-open) .conversation {
			display: none;
		}

		.chat-app.conversation-open .conversation-rail {
			display: none;
		}

		.conversation-rail {
			border: 0;
		}

		.conversation-header {
			grid-template-columns: auto auto minmax(0, 1fr) auto;
			padding-inline: var(--space-2);
		}

		.back,
		.conversation-header .conversation-avatar {
			display: grid;
		}

		.posting-toggle {
			width: var(--target-size);
			overflow: hidden;
			font-size: 0;
		}

		.messages {
			padding: var(--space-3);
		}

		.messages article {
			max-width: 88%;
		}

		.conversation form {
			position: sticky;
			inset-block-end: 0;
			padding: var(--space-2);
			padding-block-end: max(var(--space-2), env(safe-area-inset-bottom));
		}

		.conversation form :global(button) {
			width: var(--target-size);
			overflow: hidden;
			font-size: 0;
			padding: 0;
		}
	}
</style>
