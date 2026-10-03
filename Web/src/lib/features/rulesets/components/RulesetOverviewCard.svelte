<script lang="ts">
	import { resolve } from '$app/paths';
	import CheckCircle2 from '@lucide/svelte/icons/check-circle-2';
	import TriangleAlert from '@lucide/svelte/icons/triangle-alert';
	import type { RulesetSummary } from '$lib/api/types';

	let { ruleset }: { ruleset: RulesetSummary } = $props();
</script>

<a href={resolve(`/admin/rulesets/${ruleset.id}/edit/metadata`)}>
	<div class="cover" aria-hidden="true">
		<span>{ruleset.name.slice(0, 1).toUpperCase()}</span>
	</div>
	<div class="copy">
		<h2>{ruleset.name}</h2>
		{#if ruleset.status === 'valid'}
			<p class="status ready"><CheckCircle2 size={16} /> Ready to use</p>
		{:else}
			<p class="status invalid">
				<TriangleAlert size={16} /> Needs attention · {ruleset.issueCount} issues
			</p>
		{/if}
		<p class="hint">Open editor</p>
	</div>
</a>

<style>
	a {
		display: grid;
		grid-template-columns: 5.5rem minmax(0, 1fr);
		min-height: 8rem;
		border: var(--border-subtle);
		background: var(--surface-raised);
		color: var(--text-primary);
		text-decoration: none;
		transition:
			transform var(--speed-fast) ease-out,
			box-shadow var(--speed-fast) ease-out;
	}

	a:hover {
		box-shadow: var(--shadow-small);
		transform: translateY(-2px);
	}

	a:focus-visible {
		outline: var(--focus-ring);
		outline-offset: var(--focus-offset);
	}

	.cover {
		display: grid;
		place-items: center;
		border-inline-end: 1px solid var(--accent-strong);
		background: var(--action-dark);
		color: var(--accent-light);
	}

	.cover span {
		font-family: var(--font-display);
		font-size: var(--font-size-2xl);
	}

	.copy {
		align-self: center;
		padding: var(--space-3);
	}

	h2,
	p {
		margin: 0;
	}

	h2 {
		font-weight: 700;
	}

	.status {
		display: flex;
		align-items: center;
		gap: var(--space-1);
		font-size: var(--font-size-sm);
	}

	.ready {
		color: var(--status-success);
	}

	.invalid {
		color: var(--status-danger);
	}

	.hint {
		color: var(--text-secondary);
		font-size: var(--font-size-sm);
		margin-block-start: var(--space-2);
	}
</style>
