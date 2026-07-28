<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { ArrowLeft, MessageCircle } from '@lucide/svelte';
	import GameSummaryCard from '$lib/components/GameSummaryCard.svelte';
	import { api } from '$lib/api/client';
	import type { GameSummary } from '$lib/api/types';
	import { toasts } from '$lib/state/toasts.svelte';

	let summary = $state<GameSummary | null>(null);

	onMount(load);

	async function load() {
		try {
			summary = await api<GameSummary>(`/games/${page.params.id}/summary`);
		} catch (caught) {
			toasts.error(
				caught instanceof Error ? caught.message : 'The game summary could not be loaded.',
				{
					actionLabel: 'Retry',
					action: load,
					persistent: true
				}
			);
		}
	}
</script>

<div class="summary-page">
	<header>
		<div>
			<p class="eyebrow">Archived game</p>
			<h1>Game summary</h1>
			<p>This record is permanent and read-only.</p>
		</div>
		<div class="actions">
			<a href={resolve('/admin/games')}><ArrowLeft size={18} /> Back to Games</a>
			{#if summary}
				<a href={resolve(`/admin/games/${summary.game.id}/chat`)}
					><MessageCircle size={18} /> View chat</a
				>
			{/if}
		</div>
	</header>
	{#if summary}
		<GameSummaryCard {summary} />
	{:else}
		<p role="status">Loading summary…</p>
	{/if}
</div>

<style>
	.summary-page {
		width: min(100%, 64rem);
		margin-inline: auto;
	}

	header {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: var(--space-4);
		margin-block-end: var(--space-5);
	}

	header h1,
	header p {
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

	.actions {
		display: flex;
		gap: var(--space-2);
	}

	.actions a {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-1);
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		text-decoration: none;
		text-transform: uppercase;
	}

	@media (max-width: 47.99rem) {
		header {
			align-items: stretch;
			flex-direction: column;
		}

		.actions {
			justify-content: space-between;
		}
	}
</style>
