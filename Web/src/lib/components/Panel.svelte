<script lang="ts">
	import type { Snippet } from 'svelte';
	import ContentHeader from './ContentHeader.svelte';

	let {
		title,
		description = '',
		variant = 'quiet',
		children,
		actions
	}: {
		title?: string;
		description?: string;
		variant?: 'quiet' | 'focal' | 'dark';
		children: Snippet;
		actions?: Snippet;
	} = $props();
</script>

<section class:focal={variant === 'focal'} class:dark={variant === 'dark'}>
	{#if title || actions}
		<header class="panel-header">
			<ContentHeader {description} {actions} density="compact">
				{#snippet title()}
					{#if title}<h2>{title}</h2>{/if}
				{/snippet}
			</ContentHeader>
		</header>
	{/if}
	<div class="body">
		{@render children()}
	</div>
</section>

<style>
	section {
		min-width: 0;
		border-block-start: 1px solid color-mix(in srgb, var(--gold-dark) 45%, transparent);
		padding-block: var(--space-4);
	}

	.panel-header {
		margin-block-end: var(--space-3);
	}

	.focal {
		border: var(--border-strong);
		background:
			linear-gradient(135deg, rgb(255 255 255 / 20%), transparent 42%), var(--paper-light);
		box-shadow: var(--shadow);
		padding: var(--space-4);
	}

	.dark {
		border: 1px solid var(--gold-dark);
		background: var(--ink);
		color: var(--paper-light);
		padding: var(--space-4);
	}

	.dark :global(.description) {
		color: var(--paper-muted);
	}
</style>
