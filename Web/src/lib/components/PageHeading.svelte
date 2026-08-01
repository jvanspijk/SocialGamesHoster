<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		eyebrow,
		title,
		description = '',
		variant = 'default',
		actions
	}: {
		eyebrow?: string;
		title: string;
		description?: string;
		variant?: 'default' | 'spacious' | 'compact' | 'flush';
		actions?: Snippet;
	} = $props();
</script>

<header
	class:has-actions={actions}
	class:spacious={variant === 'spacious'}
	class:compact={variant === 'compact'}
	class:flush={variant === 'flush'}
>
	<div class="copy">
		{#if eyebrow}<p class="eyebrow">{eyebrow}</p>{/if}
		<h1>{title}</h1>
		{#if description}<p>{description}</p>{/if}
	</div>
	{#if actions}<div class="actions">{@render actions()}</div>{/if}
</header>

<style>
	header {
		display: flex;
		min-width: 0;
		align-items: flex-end;
		justify-content: space-between;
		gap: var(--space-4);
		margin-block-end: var(--space-5);
	}

	header.spacious {
		margin-block-end: var(--space-6);
	}

	header.compact {
		margin-block-end: var(--space-4);
	}

	header.flush {
		margin-block-end: 0;
	}

	.copy {
		min-width: 0;
	}

	h1,
	p {
		margin: 0;
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
		header.has-actions:not(.flush) {
			align-items: stretch;
			flex-direction: column;
		}

		header:not(.flush) .actions {
			width: 100%;
			justify-content: flex-start;
		}

		header:not(.flush) .actions > :global(*) {
			width: 100%;
		}
	}

	@media (max-width: 39.99rem) {
		header.flush.has-actions {
			align-items: stretch;
			flex-direction: column;
		}

		header.flush .actions {
			width: 100%;
			justify-content: flex-start;
		}

		header.flush .actions > :global(*) {
			width: 100%;
		}
	}

	@media (min-width: 48rem) and (max-width: 63.99rem) {
		header.compact.has-actions {
			align-items: stretch;
			flex-direction: column;
		}

		header.compact .actions {
			width: 100%;
			justify-content: flex-start;
		}

		header.compact .actions > :global(*) {
			width: 100%;
		}
	}
</style>
