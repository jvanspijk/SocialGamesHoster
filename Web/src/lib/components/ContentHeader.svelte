<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		eyebrow,
		description = '',
		descriptionId,
		title,
		actions,
		density = 'regular',
		alignment = 'start'
	}: {
		eyebrow?: string;
		description?: string;
		descriptionId?: string;
		title?: Snippet;
		actions?: Snippet;
		density?: 'regular' | 'compact' | 'dense' | 'flush';
		alignment?: 'start' | 'center' | 'end';
	} = $props();
</script>

<div
	class="content-header"
	class:has-actions={actions}
	class:compact={density === 'compact'}
	class:dense={density === 'dense'}
	class:flush={density === 'flush'}
	class:align-center={alignment === 'center'}
	class:align-end={alignment === 'end'}
>
	{#if eyebrow || title || description}
		<div class="copy">
			{#if eyebrow}<p class="eyebrow">{eyebrow}</p>{/if}
			{#if title}{@render title()}{/if}
			{#if description}<p class="description" id={descriptionId}>{description}</p>{/if}
		</div>
	{/if}
	{#if actions}<div class="actions">{@render actions()}</div>{/if}
</div>

<style>
	.content-header {
		display: flex;
		min-width: 0;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-4);
	}

	.content-header.align-center {
		align-items: center;
	}

	.content-header.align-end {
		align-items: flex-end;
	}

	.content-header.compact,
	.content-header.dense,
	.content-header.flush {
		gap: var(--space-3);
	}

	.copy {
		min-width: 0;
	}

	.copy :global(h1),
	.copy :global(h2),
	.copy :global(h3),
	.copy p {
		margin: 0;
	}

	.description {
		margin-block-start: var(--space-1) !important;
		color: var(--ink-soft);
	}

	.actions {
		display: flex;
		min-width: 0;
		flex-wrap: wrap;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-2);
	}

	@media (max-width: 47.99rem) {
		.content-header.has-actions:not(.flush) {
			align-items: stretch;
			flex-direction: column;
		}

		.content-header.has-actions:not(.flush) .actions {
			width: 100%;
			justify-content: flex-start;
		}

		.content-header.has-actions:not(.flush) .actions > :global(*) {
			width: 100%;
		}
	}

	@media (max-width: 39.99rem) {
		.content-header.flush.has-actions {
			align-items: stretch;
			flex-direction: column;
		}

		.content-header.flush.has-actions .actions {
			width: 100%;
			justify-content: flex-start;
		}

		.content-header.flush.has-actions .actions > :global(*) {
			width: 100%;
		}
	}

	@media (min-width: 48rem) and (max-width: 63.99rem) {
		.content-header.compact.has-actions {
			align-items: stretch;
			flex-direction: column;
		}

		.content-header.compact.has-actions .actions {
			width: 100%;
			justify-content: flex-start;
		}

		.content-header.compact.has-actions .actions > :global(*) {
			width: 100%;
		}
	}

	@media (max-width: 45rem) {
		.content-header.dense.has-actions {
			align-items: stretch;
			flex-direction: column;
		}

		.content-header.dense.has-actions .actions {
			width: 100%;
			justify-content: flex-start;
		}

		.content-header.dense.has-actions .actions > :global(*) {
			width: 100%;
		}
	}
</style>
