<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { Activity, CheckCheck, Megaphone } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import { api, pb } from '$lib/api/client';
	import type { ActivityItem, AdminAttentionSummary } from '$lib/api/types';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	type Page<T> = { items: T[]; nextCursor: string };

	let actions = $state<ActivityItem[]>([]);
	let announcements = $state<AdminAttentionSummary[]>([]);
	let activityCursor = $state('');
	let announcementCursor = $state('');
	let loading = $state(true);
	let unsubscribeRealtime = () => {};

	const view = $derived(gameState.admin);

	onMount(() => {
		void initialize();
		return () => unsubscribeRealtime();
	});

	async function initialize() {
		await load();
		const gameId = page.params.id;
		if (!gameId) return;
		unsubscribeRealtime = await pb.realtime.subscribe(`game:${gameId}:game-masters`, () => {
			void load();
		});
	}

	async function load() {
		if (!view) return;
		loading = true;
		try {
			const [activityPage, announcementPage] = await Promise.all([
				api<Page<ActivityItem>>(`/games/${view.game.id}/activity`),
				api<Page<AdminAttentionSummary>>(`/games/${view.game.id}/announcements`)
			]);
			actions = activityPage.items;
			announcements = announcementPage.items;
			activityCursor = activityPage.nextCursor;
			announcementCursor = announcementPage.nextCursor;
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'Activity could not be loaded.', {
				actionLabel: 'Retry',
				action: load,
				persistent: true
			});
		} finally {
			loading = false;
		}
	}

	async function loadMoreActions() {
		if (!view || !activityCursor) return;
		const page = await api<Page<ActivityItem>>(
			`/games/${view.game.id}/activity?cursor=${encodeURIComponent(activityCursor)}`
		);
		actions = [...actions, ...page.items];
		activityCursor = page.nextCursor;
	}

	async function loadMoreAnnouncements() {
		if (!view || !announcementCursor) return;
		const page = await api<Page<AdminAttentionSummary>>(
			`/games/${view.game.id}/announcements?cursor=${encodeURIComponent(announcementCursor)}`
		);
		announcements = [...announcements, ...page.items];
		announcementCursor = page.nextCursor;
	}

	function audienceLabel(item: AdminAttentionSummary) {
		if (item.audience === 'all') return 'Every player';
		if (item.audience === 'team') return 'One team';
		return 'One player';
	}
</script>

<header class="page-heading">
	<p class="eyebrow">Game record</p>
	<h1>Activity</h1>
	<p>
		Announcements and readable game-master actions. Chat and private role information are excluded.
	</p>
</header>

{#if loading}
	<p role="status">Loading activity…</p>
{:else}
	<div class="activity-grid">
		<section>
			<div class="section-heading">
				<Megaphone size={23} />
				<div>
					<h2>Announcements</h2>
					<p>Delivery and acknowledgement progress</p>
				</div>
			</div>
			{#if announcements.length === 0}
				<div class="empty">
					<Megaphone size={32} />
					<p>No announcements yet.</p>
				</div>
			{:else}
				<div class="timeline">
					{#each announcements as announcement (announcement.id)}
						<article>
							<div class="marker"><Megaphone size={16} /></div>
							<div>
								<div class="meta">
									<strong>{announcement.senderLabel}</strong>
									<time>{new Date(announcement.createdAt).toLocaleString()}</time>
								</div>
								<p>{announcement.content}</p>
								<div class="progress">
									<span>{audienceLabel(announcement)}</span>
									<span
										><CheckCheck size={15} />
										{announcement.acknowledgementCount} of {announcement.recipientTotal} read</span
									>
								</div>
								<progress
									value={announcement.acknowledgementCount}
									max={Math.max(announcement.recipientTotal, 1)}
									aria-label={`${announcement.acknowledgementCount} of ${announcement.recipientTotal} players acknowledged`}
								></progress>
							</div>
						</article>
					{/each}
				</div>
				{#if announcementCursor}<Button variant="ghost" onclick={loadMoreAnnouncements}
						>Load earlier announcements</Button
					>{/if}
			{/if}
		</section>

		<section>
			<div class="section-heading">
				<Activity size={23} />
				<div>
					<h2>Game-master activity</h2>
					<p>Actions that changed the game</p>
				</div>
			</div>
			{#if actions.length === 0}
				<div class="empty">
					<Activity size={32} />
					<p>No game-master activity yet.</p>
				</div>
			{:else}
				<div class="timeline">
					{#each actions as item (item.id)}
						<article>
							<div class="marker"><Activity size={16} /></div>
							<div>
								<p>{item.text}</p>
								<time>{new Date(item.createdAt).toLocaleString()}</time>
							</div>
						</article>
					{/each}
				</div>
				{#if activityCursor}<Button variant="ghost" onclick={loadMoreActions}
						>Load earlier activity</Button
					>{/if}
			{/if}
		</section>
	</div>
{/if}

<style>
	.page-heading {
		margin-block-end: var(--space-6);
	}

	.page-heading h1,
	.page-heading p {
		margin: 0;
	}

	.eyebrow {
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.14em;
		text-transform: uppercase;
	}

	.activity-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--space-7);
	}

	.section-heading {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		border-block-end: var(--border-strong);
		padding-block-end: var(--space-3);
	}

	.section-heading h2,
	.section-heading p {
		margin: 0;
	}

	.section-heading p {
		color: var(--ink-soft);
	}

	.timeline article {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding: var(--space-3) 0;
	}

	.marker {
		display: grid;
		width: 2.2rem;
		height: 2.2rem;
		place-items: center;
		border: 1px solid var(--gold-dark);
		border-radius: 50%;
		background: var(--ink);
		color: var(--gold-light);
	}

	.timeline p {
		margin: 0;
	}

	.timeline time,
	.meta time {
		color: var(--ink-soft);
		font-size: 0.75rem;
	}

	.meta,
	.progress {
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		gap: var(--space-2);
	}

	.progress {
		color: var(--ink-soft);
		font-size: 0.75rem;
	}

	.progress span {
		display: inline-flex;
		align-items: center;
		gap: 0.2rem;
	}

	progress {
		width: 100%;
		height: 0.35rem;
		accent-color: var(--crimson);
	}

	.empty {
		display: grid;
		place-items: center;
		color: var(--ink-soft);
		padding: var(--space-7);
		text-align: center;
	}

	@media (max-width: 63.99rem) {
		.activity-grid {
			grid-template-columns: 1fr;
			gap: var(--space-6);
		}
	}
</style>
