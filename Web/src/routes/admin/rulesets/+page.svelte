<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { Plus, ScrollText } from '@lucide/svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import RulesetOverviewCard from '$lib/features/rulesets/components/RulesetOverviewCard.svelte';
	import { api } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { RulesetSummary } from '$lib/api/types';
	import { toasts } from '$lib/state/toasts.svelte';

	let rulesets = $state<RulesetSummary[]>([]);
	let loading = $state(true);

	onMount(load);

	async function load() {
		try {
			rulesets = await api<RulesetSummary[]>('/rulesets');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'Rulesets could not be loaded.'), {
				actionLabel: 'Retry',
				action: load,
				persistent: true
			});
		} finally {
			loading = false;
		}
	}
</script>

<PageHeading title="Rulesets" variant="spacious">
	{#snippet actions()}
		<div class="heading-actions">
			<a class="primary-link" href={resolve('/admin/rulesets/new')}
				><Plus size={18} /> Create ruleset</a
			>
		</div>
	{/snippet}
</PageHeading>

{#if loading}
	<p role="status">Loading rulesets…</p>
{:else if rulesets.length === 0}
	<section class="empty">
		<ScrollText size={42} strokeWidth={1.5} />
		<h2>No rulesets yet</h2>
		<p>Create a ruleset from scratch, a copy, or a bundle.</p>
		<a class="primary-link" href={resolve('/admin/rulesets/new')}>Create ruleset</a>
	</section>
{:else}
	<div class="ruleset-grid">
		{#each rulesets as ruleset (ruleset.id)}
			<RulesetOverviewCard {ruleset} />
		{/each}
	</div>
{/if}

<style>
	.heading-actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	.primary-link {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		justify-content: center;
		gap: var(--space-2);
		border: 1px solid var(--action-dark);
		background: var(--action);
		box-shadow: 0 3px 0 var(--action-dark);
		color: var(--text-on-dark);
		font-family: var(--font-display);
		font-size: var(--font-size-sm);
		font-weight: 700;
		padding: var(--space-2) var(--space-4);
		text-decoration: none;
	}

	.ruleset-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 18rem), 1fr));
		gap: var(--space-4);
	}

	.empty {
		border: var(--border-strong);
		padding: var(--space-7);
		text-align: center;
	}

	@media (max-width: 47.99rem) {
		.heading-actions {
			display: grid;
			grid-template-columns: 1fr 1fr;
		}

		.heading-actions :global(button),
		.primary-link {
			width: 100%;
		}
	}
</style>
