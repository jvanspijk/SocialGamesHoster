<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { CheckCircle2, Plus, ScrollText, TriangleAlert } from '@lucide/svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
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

<PageHeading
	eyebrow="Library"
	title="Rulesets"
	description="Only ready rulesets can be used for new games. Invalid work remains editable."
	variant="spacious"
>
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
			<a href={resolve(`/admin/rulesets/${ruleset.id}/edit/metadata`)}>
				<div class="cover" aria-hidden="true">
					<span>{ruleset.name.slice(0, 1).toUpperCase()}</span>
				</div>
				<div class="copy">
					<h2>{ruleset.name}</h2>
					{#if ruleset.status === 'valid'}
						<p class="status ready"><CheckCircle2 size={16} /> Valid</p>
					{:else}
						<p class="status invalid">
							<TriangleAlert size={16} /> Invalid · {ruleset.issueCount} issues
						</p>
					{/if}
					<p class="hint">Open editor</p>
				</div>
			</a>
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
		border: 1px solid var(--crimson-dark);
		background: var(--crimson);
		box-shadow: 0 3px 0 var(--crimson-dark);
		color: var(--paper-light);
		font-family: var(--font-display);
		font-size: 0.76rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		padding: var(--space-2) var(--space-4);
		text-decoration: none;
		text-transform: uppercase;
	}

	.ruleset-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(min(100%, 18rem), 1fr));
		gap: var(--space-4);
	}

	.ruleset-grid > a {
		display: grid;
		grid-template-columns: 5.5rem minmax(0, 1fr);
		min-height: 8rem;
		border: var(--border-subtle);
		background: rgb(255 249 230 / 62%);
		color: var(--ink);
		text-decoration: none;
		transition:
			transform var(--speed-fast) ease-out,
			box-shadow var(--speed-fast) ease-out;
	}

	.ruleset-grid > a:hover {
		box-shadow: var(--shadow-small);
		transform: translateY(-2px);
	}

	.cover {
		display: grid;
		place-items: center;
		border-inline-end: 1px solid var(--gold-dark);
		background:
			radial-gradient(circle, transparent 36%, rgb(0 0 0 / 32%) 100%),
			linear-gradient(
				145deg,
				var(--crimson-dark),
				color-mix(in srgb, var(--crimson-dark) 45%, var(--wood))
			);
		color: var(--gold-light);
	}

	.cover span {
		font-family: var(--font-display);
		font-size: 2.4rem;
	}

	.copy {
		align-self: center;
		padding: var(--space-3);
	}

	h2,
	.copy p {
		margin: 0;
	}

	.status {
		display: flex;
		align-items: center;
		gap: var(--space-1);
		font-size: 0.82rem;
	}

	.ready {
		color: var(--success);
	}

	.invalid {
		color: var(--danger);
	}

	.hint {
		color: var(--ink-soft);
	}

	.hint {
		margin-block-start: var(--space-2) !important;
		font-size: 0.8rem;
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
