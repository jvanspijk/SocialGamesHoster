<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import {
		Gamepad2,
		MessageCircle,
		Shield,
		UserCircle,
		Users,
		Volume2,
		VolumeX
	} from '@lucide/svelte';
	import AppNav from '$lib/components/AppNav.svelte';
	import AttentionCard from '$lib/components/AttentionCard.svelte';
	import Button from '$lib/components/Button.svelte';
	import ConnectionBadge from '$lib/components/ConnectionBadge.svelte';
	import { api, AppApiError, jsonBody, pb } from '$lib/api/client';
	import type { Game, PlayerGameView } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { cursorIsAfter, readMarkerStorageKey } from '$lib/state/chatReadMarkers';
	import { gameState } from '$lib/state/game.svelte';
	import { sound } from '$lib/state/sound.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();
	let loading = $state(true);
	let availableLobby = $state<Game | null>(null);
	let liveGame = $state<Game | null>(null);
	let joiningLobby = $state(false);
	let acknowledging = $state(false);
	let unsubscribers: Array<() => void> = [];
	let unsubscribeLobbyOpened: (() => void) | null = null;

	const view = $derived(gameState.player);
	const accountRoute = $derived(
		page.url.pathname.startsWith('/play/profile') || page.url.pathname.startsWith('/play/settings')
	);
	const current = $derived.by(() => {
		if (page.url.pathname.startsWith('/play/role')) return 'role';
		if (page.url.pathname.startsWith('/play/party')) return 'party';
		if (page.url.pathname.startsWith('/play/chat')) return '';
		return 'game';
	});
	const hasUnreadChat = $derived.by(() => {
		void page.url.pathname;
		if (!view || typeof localStorage === 'undefined') return false;
		let markers: Record<string, { id: string; createdAt: string }> = {};
		try {
			markers = JSON.parse(
				localStorage.getItem(readMarkerStorageKey(auth.actor?.id ?? '', view.game.id)) ?? '{}'
			);
		} catch {
			markers = {};
		}
		return view.rooms.some((room) => cursorIsAfter(room.latestMessage, markers[room.id]));
	});
	const navigation = $derived([
		{ id: 'game', label: 'Game', href: resolve('/play'), icon: Gamepad2 },
		{
			id: 'role',
			label: 'Role',
			href: resolve('/play/role'),
			icon: Shield,
			disabled: !view?.roleAvailable,
			disabledDescription: 'Role unavailable. The game master has not made roles available.'
		},
		{ id: 'party', label: 'Party', href: resolve('/play/party'), icon: Users }
	]);

	onMount(() => {
		void initialize();
		return () => {
			for (const unsubscribe of unsubscribers) unsubscribe();
			unsubscribeLobbyOpened?.();
		};
	});

	async function initialize() {
		if (!auth.isPlayer) {
			loading = false;
			return;
		}
		loading = true;
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
				)
			]);
		} catch (caught) {
			if (!accountRoute) {
				toasts.error(caught instanceof Error ? caught.message : 'The game could not be loaded.', {
					actionLabel: 'Retry',
					action: initialize,
					persistent: true
				});
			}
		} finally {
			loading = false;
		}
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
		try {
			await api(`/games/${availableLobby.id}/join`, { method: 'POST', ...jsonBody({}) });
			await initialize();
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'The lobby could not be joined.');
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
			toasts.error(
				caught instanceof Error ? caught.message : 'The announcement could not be acknowledged.'
			);
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
{:else}
	<div class:account={accountRoute} class="player-shell">
		{#if view && !accountRoute}
			<AppNav items={navigation} {current} label="Player" />
		{/if}
		<header class:account={accountRoute} class="player-header">
			{#if view && !accountRoute}
				<div class="game-name">
					<strong>{view.game.name}</strong><span>{view.game.status}</span>
				</div>
			{:else}
				<div class="game-name">
					<strong>Player account</strong><span>{auth.actor?.displayName}</span>
				</div>
			{/if}
			<div class="player-tools">
				<ConnectionBadge />
				<button
					class="sound"
					type="button"
					aria-pressed={sound.enabled}
					aria-label={sound.enabled ? 'Turn sound off' : 'Turn sound on'}
					onclick={() => sound.toggle()}
				>
					{#if sound.enabled}<Volume2 size={19} />{:else}<VolumeX size={19} />{/if}
				</button>
				{#if view && !accountRoute}
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
					class:active={accountRoute}
					class="account-action"
					href={resolve('/play/profile')}
					aria-label="Profile and settings"
					aria-current={accountRoute ? 'page' : undefined}
				>
					<UserCircle size={22} />
				</a>
			</div>
		</header>
		<main class:account={accountRoute} class="player-content">
			{#if accountRoute}
				{@render children()}
			{:else if loading && !view}
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
					<p>The game master is not accepting new players right now.</p>
					<a href={resolve('/')}>Return to join page</a>
				</section>
			{:else}
				<section class="unavailable">
					<h1>No game available</h1>
					<p>Wait for the game master to allow players to join, then return to the join page.</p>
					<a href={resolve('/')}>Return to join page</a>
				</section>
			{/if}
		</main>
		{#if view && view.attentionItems.length > 0 && page.url.pathname !== resolve('/play')}
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
		background: linear-gradient(rgb(28 18 12 / 96%), rgb(20 12 8 / 98%)), var(--wood);
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
		font-size: 0.72rem;
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

	.account-action.active {
		border-radius: 50%;
		background: var(--gold);
		color: var(--wood);
	}

	.chat-action i {
		position: absolute;
		inset-block-start: 0.25rem;
		inset-inline-end: 0.2rem;
		width: 0.65rem;
		height: 0.65rem;
		border: 2px solid var(--wood);
		border-radius: 50%;
		background: var(--crimson-light);
	}

	.player-content {
		min-height: calc(100dvh - 7.75rem);
	}

	.player-content.account {
		min-height: calc(100dvh - 3.75rem);
		background: var(--paper);
	}

	.attention-popup {
		position: fixed;
		z-index: var(--layer-dialog);
		inset: 0;
		display: grid;
		place-items: center;
		background: rgb(17 10 6 / 72%);
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
		.player-shell:not(.account) {
			padding-block-end: 0;
			padding-inline-start: 13rem;
		}
	}

	@media (max-width: 47.99rem) {
		.player-tools :global(> span) {
			display: none;
		}
	}
</style>
