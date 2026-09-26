<script lang="ts">
	import { onMount, setContext } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import {
		Gamepad2,
		House,
		MessageCircle,
		Shield,
		Swords,
		UserCircle,
		Users,
		Volume2,
		VolumeX
	} from '@lucide/svelte';
	import AppNav from '$lib/components/AppNav.svelte';
	import AttentionCard from '$lib/features/play/components/AttentionCard.svelte';
	import {
		playerShellContextKey,
		type PlayerShellContext
	} from '$lib/features/play/playerShellContext';
	import Button from '$lib/components/Button.svelte';
	import { api, AppApiError, jsonBody, pb } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { ChatMessage, Game, PlayerGameView, RealtimeEnvelope } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import {
		chatReadMarkersChanged,
		cursorIsAfter,
		hasUnreadMessages,
		readMarkers
	} from '$lib/state/chatReadMarkers';
	import { gameState } from '$lib/state/game.svelte';
	import { sound } from '$lib/state/sound.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();
	let loading = $state(true);
	let availableLobby = $state<Game | null>(null);
	let liveGame = $state<Game | null>(null);
	let joiningLobby = $state(false);
	let loadError = $state('');
	let acknowledging = $state(false);
	let hasUnreadChat = $state(false);
	let unsubscribers: Array<() => void> = [];
	let unsubscribeLobbyOpened: (() => void) | null = null;

	const view = $derived(gameState.player);
	const homeRoute = $derived(page.url.pathname === '/play');
	const accountRoute = $derived(
		page.url.pathname.startsWith('/play/profile') ||
			page.url.pathname.startsWith('/play/history') ||
			page.url.pathname.startsWith('/play/settings')
	);
	const hubRoute = $derived(homeRoute || accountRoute);
	const current = $derived.by(() => {
		if (page.url.pathname.startsWith('/play/role')) return 'role';
		if (page.url.pathname.startsWith('/play/party')) return 'party';
		if (page.url.pathname.startsWith('/play/chat')) return '';
		if (page.url.pathname.startsWith('/play/game')) return 'game';
		return '';
	});
	const navigation = $derived([
		{ id: 'game', label: 'Game', href: resolve('/play/game'), icon: Gamepad2 },
		{
			id: 'role',
			label: 'Role',
			href: resolve('/play/role'),
			icon: Shield,
			disabled: !view?.roleAvailable,
			disabledDescription: 'Roles are hidden.'
		},
		{ id: 'party', label: 'Party', href: resolve('/play/party'), icon: Users }
	]);

	setContext<PlayerShellContext>(playerShellContextKey, {
		get loading() {
			return loading;
		},
		get availableLobby() {
			return availableLobby;
		},
		get liveGame() {
			return liveGame;
		},
		get joiningLobby() {
			return joiningLobby;
		},
		get loadError() {
			return loadError;
		},
		refresh: initialize,
		joinAvailableLobby
	});

	$effect(() => {
		void view;
		refreshUnreadChat();
	});

	onMount(() => {
		void initialize();
		window.addEventListener(chatReadMarkersChanged, refreshUnreadChat);
		return () => {
			for (const unsubscribe of unsubscribers) unsubscribe();
			unsubscribeLobbyOpened?.();
			window.removeEventListener(chatReadMarkersChanged, refreshUnreadChat);
		};
	});

	async function initialize() {
		if (!auth.isPlayer) {
			loading = false;
			return;
		}
		loading = true;
		loadError = '';
		availableLobby = null;
		liveGame = null;
		try {
			let loaded: PlayerGameView;
			try {
				loaded = await gameState.refreshPlayer();
			} catch (caught) {
				if (
					!(caught instanceof AppApiError) ||
					!['game.no_live_game', 'game.not_joined'].includes(caught.body.code)
				) {
					throw caught;
				}
				if (caught.body.code === 'game.not_joined') {
					const currentLiveGame = await api<Game>('/games/live');
					if (currentLiveGame.joiningOpen) {
						availableLobby = currentLiveGame;
					} else {
						liveGame = currentLiveGame;
					}
				}
				await subscribeToLobbyOpening();
				return;
			}
			unsubscribeLobbyOpened?.();
			unsubscribeLobbyOpened = null;
			unsubscribers = await Promise.all([
				gameState.subscribe(`game:${loaded.game.id}:public`, () => gameState.refreshPlayer()),
				gameState.subscribe(`participant:${loaded.participant.id}:private`, () =>
					gameState.refreshPlayer()
				),
				...loaded.rooms.map((room) =>
					pb.realtime.subscribe(`room:${room.id}`, (raw) => {
						const event = raw as unknown as RealtimeEnvelope<ChatMessage>;
						if (event.kind !== 'chat.message_created') return;
						const markers = readMarkers(auth.actor?.id ?? '', loaded.game.id);
						if (cursorIsAfter(event.payload, markers[room.id])) hasUnreadChat = true;
					})
				)
			]);
			refreshUnreadChat();
		} catch (caught) {
			loadError = errorMessage(caught, 'The game could not be loaded.');
			if (!hubRoute) {
				toasts.error(errorMessage(caught, 'The game could not be loaded.'), {
					actionLabel: 'Retry',
					action: initialize,
					persistent: true
				});
			}
		} finally {
			loading = false;
		}
	}

	function refreshUnreadChat() {
		if (!view || typeof localStorage === 'undefined') {
			hasUnreadChat = false;
			return;
		}
		hasUnreadChat = hasUnreadMessages(view.rooms, readMarkers(auth.actor?.id ?? '', view.game.id));
	}

	async function subscribeToLobbyOpening() {
		if (!auth.actor || unsubscribeLobbyOpened) return;
		unsubscribeLobbyOpened = await pb.realtime.subscribe(
			`profile:${auth.actor.id}`,
			async (raw) => {
				const event = raw as unknown as { kind?: string };
				if (event.kind === 'game.lobby_opened') await initialize();
			}
		);
	}

	async function joinAvailableLobby() {
		if (!availableLobby) return;
		joiningLobby = true;
		loadError = '';
		try {
			await api(`/games/${availableLobby.id}/join`, { method: 'POST', ...jsonBody({}) });
			await initialize();
		} catch (caught) {
			loadError = errorMessage(caught, 'The lobby could not be joined.');
			toasts.error(loadError);
		} finally {
			joiningLobby = false;
		}
	}

	async function acknowledgeAnnouncement() {
		if (!view || view.attentionItems.length === 0) return;
		acknowledging = true;
		try {
			await api(`/games/${view.game.id}/announcements/${view.attentionItems[0].id}/acknowledge`, {
				method: 'POST'
			});
			await gameState.refreshPlayer();
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The announcement could not be acknowledged.'));
		} finally {
			acknowledging = false;
		}
	}
</script>

{#if !auth.isPlayer}
	<main class="unavailable">
		<h1>Join a game</h1>
		<p>Choose or create your player profile first.</p>
		<a href={resolve('/')}>Return to join page</a>
	</main>
{:else if hubRoute}
	<div class="account-shell">
		<header class="account-header">
			<a class="account-product" href={resolve('/play')}>
				<Swords size={22} /> Player home
			</a>
			<div class="account-tools">
				<a href={resolve('/play/profile')} aria-label="Player profile">
					<UserCircle size={22} />
				</a>
			</div>
		</header>
		<main class="account-content">{@render children()}</main>
	</div>
{:else}
	<div class="player-shell">
		{#if view}<AppNav items={navigation} {current} label="Player" />{/if}
		<header class="player-header">
			<div class="game-name">
				<strong>{view?.game.name ?? 'Current game'}</strong><span
					>{view?.game.status ?? 'Unavailable'}</span
				>
			</div>
			<div class="player-tools">
				<button
					class="sound"
					type="button"
					aria-pressed={sound.enabled}
					aria-label={sound.enabled ? 'Turn sound off' : 'Turn sound on'}
					onclick={() => sound.toggle()}
				>
					{#if sound.enabled}<Volume2 size={19} />{:else}<VolumeX size={19} />{/if}
				</button>
				{#if view}
					<a
						class="chat-action"
						href={resolve('/play/chat')}
						aria-label={hasUnreadChat ? 'Chat, new messages' : 'Chat'}
					>
						<MessageCircle size={21} />
						{#if hasUnreadChat}<i></i>{/if}
					</a>
				{/if}
				<a
					class="account-action"
					href={resolve('/play')}
					aria-label="Leave game and return to Player home"
				>
					<House size={22} />
				</a>
			</div>
		</header>
		<main class="player-content">
			{#if loading && !view}
				<p role="status">Loading game…</p>
			{:else if view}
				{@render children()}
			{:else if availableLobby}
				<section class="unavailable">
					<h1>Game accepting players</h1>
					<p>{availableLobby.name} is ready for you to join.</p>
					<Button loading={joiningLobby} onclick={joinAvailableLobby}>Join game</Button>
				</section>
			{:else if liveGame}
				<section class="unavailable">
					<h1>{liveGame.name} has started</h1>
					<p>The Game Master is not accepting new players right now.</p>
					<a href={resolve('/play')}>Return to Player home</a>
				</section>
			{:else}
				<section class="unavailable">
					<h1>No game available</h1>
					<p>Wait for the Game Master to allow players to join.</p>
					<a href={resolve('/play')}>Return to Player home</a>
				</section>
			{/if}
		</main>
		{#if view && view.attentionItems.length > 0 && page.url.pathname !== resolve('/play/game')}
			<section class="attention-popup" aria-live="assertive" aria-label="New announcement">
				<AttentionCard
					item={view.attentionItems[0]}
					position={1}
					total={view.attentionItems.length}
					acknowledge={acknowledgeAnnouncement}
					busy={acknowledging}
				/>
			</section>
		{/if}
	</div>
{/if}

<style>
	.player-shell {
		min-height: 100dvh;
		padding-block-end: calc(4rem + env(safe-area-inset-bottom));
	}

	.account-shell {
		min-height: 100dvh;
		background: var(--paper);
	}

	.account-header {
		position: sticky;
		z-index: var(--layer-sticky);
		inset-block-start: 0;
		display: flex;
		min-height: 4rem;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		background: var(--surface-shell);
		padding: var(--space-2) max(var(--space-4), env(safe-area-inset-right)) var(--space-2)
			max(var(--space-4), env(safe-area-inset-left));
	}

	.account-product {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		color: var(--ink);
		font-family: var(--font-display);
		font-size: var(--font-size-sm);
		font-weight: 700;
		text-decoration: none;
	}

	.account-tools {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.account-tools > a {
		display: grid;
		width: var(--target-size);
		height: var(--target-size);
		place-items: center;
		color: var(--ink);
	}

	.player-header {
		position: sticky;
		z-index: var(--layer-sticky);
		inset-block-start: 0;
		display: flex;
		min-height: 3.75rem;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: 1px solid var(--gold-dark);
		background: var(--surface-nav);
		color: var(--paper-light);
		padding: var(--space-2) max(var(--space-3), env(safe-area-inset-right)) var(--space-2)
			max(var(--space-3), env(safe-area-inset-left));
	}

	.game-name {
		min-width: 0;
	}

	.game-name strong,
	.game-name span {
		display: block;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.game-name strong {
		color: var(--gold-light);
		font-family: var(--font-display);
	}

	.game-name span {
		color: var(--paper-muted);
		font-size: var(--font-size-sm);
		text-transform: capitalize;
	}

	.player-tools {
		display: flex;
		align-items: center;
		gap: var(--space-1);
	}

	.sound,
	.chat-action,
	.account-action {
		position: relative;
		display: grid;
		width: var(--target-size);
		height: var(--target-size);
		place-items: center;
		border: 0;
		background: transparent;
		color: var(--paper-light);
		cursor: pointer;
	}

	.chat-action i {
		position: absolute;
		inset-block-start: 0.25rem;
		inset-inline-end: 0.2rem;
		width: 0.65rem;
		height: 0.65rem;
		border: 2px solid var(--wood);
		border-radius: 50%;
		background: var(--action-light);
	}

	.player-content {
		min-height: calc(100dvh - 7.75rem);
	}

	.attention-popup {
		position: fixed;
		z-index: var(--layer-dialog);
		inset: 0;
		display: grid;
		place-items: center;
		background: var(--surface-backdrop);
		padding: max(var(--space-4), env(safe-area-inset-top))
			max(var(--space-4), env(safe-area-inset-right))
			max(var(--space-4), env(safe-area-inset-bottom))
			max(var(--space-4), env(safe-area-inset-left));
	}

	.unavailable {
		display: grid;
		min-height: 100dvh;
		place-content: center;
		padding: var(--space-5);
		text-align: center;
	}

	.unavailable h1,
	.unavailable p {
		margin: 0;
	}

	.unavailable a {
		margin-block-start: var(--space-3);
	}

	@media (min-width: 64rem) {
		.player-shell {
			padding-block-end: 0;
			padding-inline-start: 13rem;
		}
	}
</style>
