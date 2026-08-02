<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { MessageCircle } from '@lucide/svelte';
	import Composer from '$lib/components/Composer.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import MessageList, { type MessageListItem } from '$lib/components/MessageList.svelte';
	import SplitView from '$lib/components/SplitView.svelte';
	import type { SelectableListEntry } from '$lib/components/SelectableList.svelte';
	import { api, jsonBody, pb } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { ChatMessage, MessageSummary, RealtimeEnvelope, Room } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import {
		chatReadMarkersChanged,
		cursorIsAfter,
		readMarkerStorageKey,
		readMarkers
	} from '$lib/state/chatReadMarkers';
	import { toasts } from '$lib/state/toasts.svelte';
	import ChatRail from './ChatRail.svelte';
	import ChatReadOnlyBanner from './ChatReadOnlyBanner.svelte';
	import ConversationHeader from './ConversationHeader.svelte';

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
	let loadedPolicyRevision: string | undefined;
	let roomLoadRequest = 0;
	let openedRoomId = '';
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
	const conversationEntries = $derived<readonly SelectableListEntry[]>(
		filteredRooms.map((room) => ({
			id: room.id,
			label: room.label,
			accessibleLabel: `${room.label}${unread(room) ? ', New messages' : ''}`,
			description: room.latestMessage
				? `${room.latestMessage.senderLabel}: ${room.latestMessage.preview}`
				: 'No messages yet',
			supportingLabel: roomType(room),
			metaLabel: timeLabel(room.latestMessage?.createdAt),
			leadingText: room.label.slice(0, 1).toUpperCase(),
			leadingVariant: room.kind === 'team' ? 'people' : room.kind === 'general' ? 'hash' : 'text',
			unread: unread(room)
		}))
	);
	const messageItems = $derived<readonly MessageListItem[]>(
		messages.map((message) => ({
			id: message.id,
			senderLabel: message.senderLabel,
			timeLabel: timeLabel(message.createdAt),
			dayKey: calendarDayKey(message.createdAt),
			dayLabel: dayLabel(message.createdAt),
			content: message.content,
			isOwn: message.isOwn,
			deleted: message.deleted,
			canRemove: !message.deleted && (canModerate || Boolean(message.isOwn)),
			removeLabel: `Remove message from ${message.senderLabel}`
		}))
	);

	onMount(() => {
		loadMarkers();
		return () => {
			for (const unsubscribe of subscriptions) unsubscribe();
		};
	});

	$effect(() => {
		const hasSelectedRoom = selectedRoomId && rooms.some((room) => room.id === selectedRoomId);
		if (!hasSelectedRoom) {
			if (openedRoomId && messagesElement) {
				scrollPositions.set(openedRoomId, messagesElement.scrollTop);
			}
			openedRoomId = '';
			return;
		}
		if (openedRoomId === selectedRoomId) return;
		void openConversation(selectedRoomId);
	});

	$effect(() => {
		if (!policyRevision || policyRevision === loadedPolicyRevision) return;
		const refreshSelectedRoom = loadedPolicyRevision !== undefined;
		loadedPolicyRevision = policyRevision;
		void loadRooms(refreshSelectedRoom);
	});

	function loadMarkers() {
		markers = readMarkers(auth.actor?.id ?? '', gameId);
	}

	function saveMarkers() {
		localStorage.setItem(
			readMarkerStorageKey(auth.actor?.id ?? '', gameId),
			JSON.stringify(markers)
		);
		window.dispatchEvent(new Event(chatReadMarkersChanged));
	}

	async function loadRooms(refreshSelectedRoom = false) {
		const request = ++roomLoadRequest;
		loadingRooms = true;
		try {
			const loadedRooms = await api<Room[]>(`/games/${gameId}/rooms`);
			if (request !== roomLoadRequest) return;
			rooms = loadedRooms;
			if (selectedRoomId && !rooms.some((room) => room.id === selectedRoomId)) {
				selectRoom('');
			}
			await subscribeRooms();
			if (
				refreshSelectedRoom &&
				selectedRoomId &&
				rooms.some((room) => room.id === selectedRoomId)
			) {
				await openConversation(selectedRoomId);
			}
		} catch (caught) {
			if (request !== roomLoadRequest) return;
			toasts.error(errorMessage(caught, 'Conversations could not be loaded.'), {
				actionLabel: 'Retry',
				action: loadRooms,
				persistent: true
			});
		} finally {
			if (request === roomLoadRequest) loadingRooms = false;
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
		const room = rooms.find((item) => item.id === roomId);
		if (!room) return;
		if (openedRoomId && messagesElement) {
			scrollPositions.set(openedRoomId, messagesElement.scrollTop);
		}
		openedRoomId = roomId;
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
			toasts.error(errorMessage(caught, 'Messages could not be loaded.'));
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

	async function send() {
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
			toasts.error(errorMessage(caught, 'The message could not be sent.'));
		} finally {
			sending = false;
		}
	}

	async function remove(messageId: string) {
		if (!selectedRoom) return;
		try {
			const deleted = await api<ChatMessage>(`/rooms/${selectedRoom.id}/messages/${messageId}`, {
				method: 'DELETE'
			});
			messages = messages.map((item) => (item.id === deleted.id ? deleted : item));
			updateRoomMessage(selectedRoom.id, deleted);
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The message could not be removed.'));
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
			toasts.error(errorMessage(caught, 'Posting access could not be updated.'));
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

	function calendarDayKey(value: string) {
		const date = new Date(value);
		return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`;
	}
</script>

<SplitView detailOpen={Boolean(selectedRoom)}>
	{#snippet rail()}
		<ChatRail
			entries={conversationEntries}
			selectedId={selectedRoomId}
			bind:search
			loading={loadingRooms}
			onselect={selectRoom}
			onnewmessage={newMessage}
		/>
	{/snippet}
	{#snippet detail()}
		{#if selectedRoom}
			<div class="conversation">
				<ConversationHeader
					label={selectedRoom.label}
					initial={selectedRoom.label.slice(0, 1).toUpperCase()}
					typeLabel={roomType(selectedRoom)}
					playersCanPost={selectedRoom.playersCanPost}
					{canModerate}
					{archived}
					onback={() => selectRoom('')}
					ontoggleposting={togglePosting}
				/>
				<MessageList
					messages={messageItems}
					loading={loadingMessages}
					hasEarlierMessages={Boolean(nextCursor)}
					{firstUnreadId}
					bind:messageElement={messagesElement}
					emptyDescription={selectedRoom.sendable && !archived
						? 'Start the conversation.'
						: 'This conversation is read-only.'}
					onloadEarlier={loadEarlier}
					onremove={remove}
				>
					{#snippet emptyIcon()}<MessageCircle size={36} strokeWidth={1.4} />{/snippet}
				</MessageList>
				{#if !archived && selectedRoom.sendable}
					<Composer
						bind:value={content}
						restrictionLabel={selectedRoom.messageRestriction === 'emoji_only' ? 'Emoji only' : ''}
						placeholder={selectedRoom.messageRestriction === 'emoji_only'
							? 'Add emoji'
							: 'Write a message'}
						{sending}
						onsubmit={send}
					/>
				{:else}
					<ChatReadOnlyBanner
						message={archived
							? 'Archived chat is read-only'
							: 'You cannot post in this conversation'}
					/>
				{/if}
			</div>
		{:else}
			<div class="conversation-placeholder">
				<EmptyState
					title="Select a conversation"
					description="Choose a conversation from the list."
				>
					{#snippet icon()}<MessageCircle size={48} strokeWidth={1.3} />{/snippet}
				</EmptyState>
			</div>
		{/if}
	{/snippet}
</SplitView>

<style>
	.conversation {
		display: grid;
		height: 100%;
		min-width: 0;
		min-height: 0;
		grid-template-rows: auto minmax(0, 1fr) auto;
		background:
			radial-gradient(circle at 15% 22%, rgb(112 73 35 / 6%) 0 0.08rem, transparent 0.11rem),
			var(--paper);
		background-size:
			1.8rem 2rem,
			auto;
	}

	.conversation-placeholder {
		display: grid;
		height: 100%;
		place-items: center;
		padding: var(--space-6);
		text-align: center;
	}
</style>
