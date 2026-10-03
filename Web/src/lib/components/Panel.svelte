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
		--panel-border: 1px solid color-mix(in srgb, var(--accent-strong) 45%, transparent);
		--panel-background: var(--surface-raised);
		--panel-shadow: none;
		min-width: 0;
		border: var(--panel-border);
		background: var(--panel-background);
		box-shadow: var(--panel-shadow);
		padding: var(--space-4);
	}

	.panel-header {
		margin-block-end: var(--space-4);
	}

	h2 {
		color: inherit;
	}

	.focal {
		--panel-border: var(--border-strong);
		--panel-background: var(--surface-raised);
		--panel-shadow: var(--shadow-small);
	}

	.dark {
		--panel-border: 1px solid var(--accent-strong);
		--panel-background: var(--surface-dark);
		--panel-foreground: var(--text-on-dark);
		--header-description-color: var(--text-on-dark-muted);
		color: var(--panel-foreground);
	}
</style>
