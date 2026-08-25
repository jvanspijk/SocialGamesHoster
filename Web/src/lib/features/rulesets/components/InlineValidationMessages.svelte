<script lang="ts">
	import type { ValidationIssue } from '../editor-state';

	let { issues = [], path }: { issues?: ValidationIssue[]; path: string } = $props();
	const matching = $derived(
		issues.filter((issue) => issue.path === path || issue.path.startsWith(`${path}.`))
	);
</script>

{#if matching.length}
	<section class="inline-validation" aria-label="Issues to fix">
		<strong>Needs attention</strong>
		<ul>
			{#each matching as issue (`${issue.path}:${issue.message}`)}
				<li>{issue.message}</li>
			{/each}
		</ul>
	</section>
{/if}

<style>
	.inline-validation {
		border-inline-start: 3px solid var(--danger);
		color: var(--danger);
		padding-inline-start: var(--space-2);
	}

	strong {
		font-family: var(--font-display);
		font-size: 0.72rem;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	ul {
		margin: var(--space-1) 0 0;
		padding-inline-start: var(--space-4);
	}
</style>
