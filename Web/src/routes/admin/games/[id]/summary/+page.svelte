<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { ArrowLeft, MessageCircle } from '@lucide/svelte';
	import GameSummaryCard from '$lib/features/games/components/GameSummaryCard.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import { api } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { GameSummary } from '$lib/api/types';
	import { toasts } from '$lib/state/toasts.svelte';

	let summary = $state<GameSummary | null>(null);

	onMount(load);

	async function load() {
		try {
			summary = await api<GameSummary>(`/games/${page.params.id}/summary`);
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The game summary could not be loaded.'), {
				actionLabel: 'Retry',
				action: load,
				persistent: true
			});
		}
	}
</script>

<div class="summary-page">
	<PageHeading
		eyebrow="Archived game"
		title="Game summary"
		description="This record is permanent and read-only."
	>
		{#snippet actions()}
			<div class="summary-actions">
				<a href={resolve('/admin/games')}><ArrowLeft size={18} /> Back to Games</a>
				{#if summary}
					<a href={resolve(`/admin/games/${summary.game.id}/chat`)}
						><MessageCircle size={18} /> View chat</a
					>
				{/if}
			</div>
		{/snippet}
	</PageHeading>
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

	.summary-actions {
		display: flex;
		gap: var(--space-2);
	}

	.summary-actions a {
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
		.summary-actions {
			justify-content: space-between;
		}
	}
</style>
