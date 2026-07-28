<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import {
		Activity,
		ArrowLeft,
		MessageCircle,
		Users,
		Volume2,
		VolumeX,
		Gauge
	} from '@lucide/svelte';
	import AppNav from '$lib/components/AppNav.svelte';
	import ConnectionBadge from '$lib/components/ConnectionBadge.svelte';
	import { api, pb } from '$lib/api/client';
	import type { ChatMessage, RealtimeEnvelope } from '$lib/api/types';
	import { gameState } from '$lib/state/game.svelte';
	import { auth } from '$lib/state/auth.svelte';
	import {
		chatReadMarkersChanged,
		countUnreadMessages,
		readMarkers
	} from '$lib/state/chatReadMarkers';
	import { sound } from '$lib/state/sound.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();
	let loading = $state(true);
	let unsubscribers: Array<() => void> = [];
	let unreadChatCount = $state(0);
	let unreadRequest = 0;

	const view = $derived(gameState.admin);
	const standalone = $derived(
		page.url.pathname.includes('/finish/') || page.url.pathname.endsWith('/summary')
	);
	const current = $derived.by(() => {
		const path = page.url.pathname;
		if (path.includes('/players')) return 'players';
		if (path.includes('/chat')) return 'chat';
		if (path.includes('/activity')) return 'activity';
		return 'overview';
	});
	const navigation = $derived([
		{
			id: 'overview',
			label: 'Overview',
			href: resolve(`/admin/games/${page.params.id}/overview`),
			icon: Gauge
		},
		{
			id: 'players',
			label: 'Players',
			href: resolve(`/admin/games/${page.params.id}/players`),
			icon: Users
		},
		{
			id: 'chat',
			label: 'Chat',
			href: resolve(`/admin/games/${page.params.id}/chat`),
			icon: MessageCircle,
			attention: unreadChatCount > 0,
			attentionCount: unreadChatCount,
			attentionLabel: unreadChatCount === 1 ? 'unread message' : 'unread messages'
		},
		{
			id: 'activity',
			label: 'Activity',
			href: resolve(`/admin/games/${page.params.id}/activity`),
			icon: Activity
		}
	]);

	onMount(() => {
		void initialize();
		window.addEventListener(chatReadMarkersChanged, refreshUnreadChatCount);
		return () => {
			for (const unsubscribe of unsubscribers) unsubscribe();
			window.removeEventListener(chatReadMarkersChanged, refreshUnreadChatCount);
		};
	});

	async function initialize() {
		loading = true;
		try {
			const loaded = await gameState.refreshAdmin(page.params.id ?? '');
			unsubscribers = await Promise.all([
				gameState.subscribe(`game:${loaded.game.id}:public`, () =>
					gameState.refreshAdmin(loaded.game.id)
				),
				gameState.subscribe(`game:${loaded.game.id}:game-masters`, () =>
					gameState.refreshAdmin(loaded.game.id)
				),
				...loaded.rooms.map((room) =>
					pb.realtime.subscribe(`room:${room.id}`, (raw) => {
						const event = raw as unknown as RealtimeEnvelope<ChatMessage>;
						if (event.kind === 'chat.message_created' || event.kind === 'chat.message_deleted') {
							void refreshUnreadChatCount();
						}
					})
				)
			]);
			await refreshUnreadChatCount();
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'The game could not be loaded.', {
				actionLabel: 'Retry',
				action: initialize,
				persistent: true
			});
		} finally {
			loading = false;
		}
	}

	async function refreshUnreadChatCount() {
		const currentView = view;
		if (!currentView || typeof localStorage === 'undefined') return;
		const request = ++unreadRequest;
		try {
			const total = await countUnreadMessages(
				currentView.rooms,
				readMarkers(auth.actor?.id ?? '', currentView.game.id),
				(roomId, cursor) =>
					api<{ items: ChatMessage[]; nextCursor: string }>(
						`/rooms/${roomId}/messages${cursor ? `?cursor=${encodeURIComponent(cursor)}` : ''}`
					)
			);
			if (request === unreadRequest) unreadChatCount = total;
		} catch {
			// The existing chat view will surface a read failure when the host becomes available again.
		}
	}

	function statusLabel(status: string) {
		return (
			(
				{
					draft: 'Draft',
					lobby: 'Lobby open',
					running: 'Running',
					paused: 'Paused',
					review: 'Finishing',
					archived: 'Archived'
				} as Record<string, string>
			)[status] ?? status
		);
	}
</script>

<div class:standalone class="live-shell">
	{#if view && !standalone}
		<AppNav items={navigation} {current} label="Live game" />
	{/if}
	<header class="live-header">
		<a class="back" href={resolve('/admin/games')}
			><ArrowLeft size={18} /> <span>Back to Games</span></a
		>
		{#if view}
			<div class="game-identity">
				<strong>{view.game.name}</strong>
				<span>{statusLabel(view.game.status)}</span>
			</div>
		{/if}
		<div class="live-tools">
			<ConnectionBadge />
			<button
				type="button"
				class="sound"
				aria-pressed={sound.enabled}
				onclick={() => sound.toggle()}
			>
				{#if sound.enabled}<Volume2 size={18} /> Sound on{:else}<VolumeX size={18} /> Sound off{/if}
			</button>
		</div>
	</header>
	<main class="live-content">
		{#if loading && !view}
			<p role="status">Loading game…</p>
		{:else if view}
			{@render children()}
		{:else}
			<section class="load-failure">
				<h1>Game unavailable</h1>
				<p>Return to Games and choose another game.</p>
			</section>
		{/if}
	</main>
</div>

<style>
	.live-shell {
		min-height: 100dvh;
		padding-block-end: calc(4rem + env(safe-area-inset-bottom));
	}

	.live-header {
		position: sticky;
		z-index: var(--layer-sticky);
		inset-block-start: 0;
		display: grid;
		grid-template-columns: minmax(8rem, 1fr) minmax(0, 2fr) minmax(15rem, 1fr);
		min-height: 4.25rem;
		align-items: center;
		gap: var(--space-3);
		border-block-end: 1px solid var(--gold-dark);
		background: linear-gradient(rgb(28 18 12 / 96%), rgb(20 12 8 / 98%)), var(--wood);
		color: var(--paper-light);
		padding: var(--space-2) max(var(--space-4), env(safe-area-inset-right)) var(--space-2)
			max(var(--space-4), env(safe-area-inset-left));
	}

	.back {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-1);
		color: var(--paper-muted);
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		text-decoration: none;
		text-transform: uppercase;
	}

	.game-identity {
		min-width: 0;
		text-align: center;
	}

	.game-identity strong {
		display: block;
		overflow: hidden;
		color: var(--gold-light);
		font-family: var(--font-display);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.game-identity span {
		color: var(--paper-muted);
		font-size: 0.78rem;
	}

	.live-tools {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-2);
	}

	.sound {
		display: inline-flex;
		width: 7.5rem;
		min-height: var(--target-size);
		align-items: center;
		justify-content: center;
		gap: var(--space-1);
		border: 1px solid var(--gold-dark);
		background: rgb(255 255 255 / 4%);
		color: var(--paper-light);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.68rem;
		font-weight: 700;
		text-transform: uppercase;
	}

	.live-content {
		width: min(100%, 84rem);
		margin-inline: auto;
		padding: clamp(var(--space-3), 3vw, var(--space-6));
	}

	.load-failure {
		padding: var(--space-7);
		text-align: center;
	}

	@media (min-width: 64rem) {
		.live-shell {
			padding-block-end: 0;
			padding-inline-start: 13rem;
		}

		.live-shell.standalone {
			padding-inline-start: 0;
		}
	}

	@media (max-width: 47.99rem) {
		.live-header {
			grid-template-columns: auto minmax(0, 1fr) auto;
			min-height: 3.75rem;
			padding-inline: max(var(--space-2), env(safe-area-inset-left))
				max(var(--space-2), env(safe-area-inset-right));
		}

		.back {
			width: var(--target-size);
			justify-content: center;
		}

		.back span,
		.live-tools :global(.connection-label) {
			display: none;
		}

		.live-tools :global(> span) {
			display: none;
		}

		.sound {
			width: var(--target-size);
			font-size: 0;
		}
	}
</style>
