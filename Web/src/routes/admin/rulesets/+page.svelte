<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { Archive, CheckCircle2, FileUp, Plus, ScrollText, TriangleAlert } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import { api } from '$lib/api/client';
	import type { RulesetSummary } from '$lib/api/types';
	import { toasts } from '$lib/state/toasts.svelte';

	let rulesets = $state<RulesetSummary[]>([]);
	let loading = $state(true);
	let importing = $state(false);
	let importInput: HTMLInputElement;

	onMount(load);

	async function load() {
		try {
			rulesets = await api<RulesetSummary[]>('/rulesets');
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'Rulesets could not be loaded.', {
				actionLabel: 'Retry',
				action: load,
				persistent: true
			});
		} finally {
			loading = false;
		}
	}

	async function importRuleset() {
		const file = importInput.files?.[0];
		if (!file) return;
		importing = true;
		try {
			const created = await api<{ ruleset: RulesetSummary }>('/rulesets/import', {
				method: 'POST',
				body: file
			});
			rulesets = [created.ruleset, ...rulesets];
			toasts.success('Ruleset imported.');
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'The ruleset could not be imported.');
		} finally {
			importing = false;
			importInput.value = '';
		}
	}
</script>

<header class="page-heading">
	<div>
		<p class="eyebrow">Library</p>
		<h1>Rulesets</h1>
		<p>Only ready rulesets can be used for new games. Invalid work remains editable.</p>
	</div>
	<div class="heading-actions">
		<input
			class="sr-only"
			bind:this={importInput}
			type="file"
			accept=".sghrules,application/vnd.socialgameshoster.ruleset+zip"
			onchange={importRuleset}
		/>
		<Button variant="secondary" loading={importing} onclick={() => importInput.click()}>
			<FileUp size={18} /> Import
		</Button>
		<a class="primary-link" href={resolve('/admin/rulesets/new')}><Plus size={18} /> New ruleset</a>
	</div>
</header>

{#if loading}
	<p role="status">Loading rulesets…</p>
{:else if rulesets.length === 0}
	<section class="empty">
		<ScrollText size={42} strokeWidth={1.5} />
		<h2>No rulesets yet</h2>
		<p>Create a ruleset or import an existing bundle.</p>
		<a class="primary-link" href={resolve('/admin/rulesets/new')}>New ruleset</a>
	</section>
{:else}
	<div class="ruleset-grid">
		{#each rulesets as ruleset (ruleset.id)}
			<a
				class:archived={ruleset.archived}
				href={resolve(`/admin/rulesets/${ruleset.id}/edit/metadata`)}
			>
				<div class="cover" aria-hidden="true">
					<span>{ruleset.name.slice(0, 1).toUpperCase()}</span>
				</div>
				<div class="copy">
					<h2>{ruleset.name}</h2>
					{#if ruleset.archived}
						<p class="status archived-status"><Archive size={16} /> Archived</p>
					{:else if ruleset.latestPublishedVersion}
						<p class="status ready"><CheckCircle2 size={16} /> Ready</p>
					{:else}
						<p class="status invalid"><TriangleAlert size={16} /> Invalid</p>
					{/if}
					<p class="hint">Open editor</p>
				</div>
			</a>
		{/each}
	</div>
{/if}

<style>
	.page-heading {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: var(--space-4);
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

	.ruleset-grid > a.archived {
		opacity: 0.7;
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

	.archived-status,
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
		.page-heading {
			align-items: stretch;
			flex-direction: column;
		}

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
