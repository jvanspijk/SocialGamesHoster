<script lang="ts">
	import type { Snippet } from 'svelte';

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
		<header>
			<div>
				{#if title}<h2>{title}</h2>{/if}
				{#if description}<p>{description}</p>{/if}
			</div>
			{#if actions}<div class="actions">{@render actions()}</div>{/if}
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

	header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-3);
		margin-block-end: var(--space-3);
	}

	h2,
	p {
		margin: 0;
	}

	header p {
		margin-block-start: var(--space-1);
		color: var(--ink-soft);
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: var(--space-2);
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

	.dark header p {
		color: var(--paper-muted);
	}
</style>
