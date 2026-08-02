<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		eyebrow = '',
		title,
		actions,
		toolbar,
		children
	}: {
		eyebrow?: string;
		title: string;
		actions?: Snippet;
		toolbar?: Snippet;
		children: Snippet;
	} = $props();
</script>

<aside class="item-rail">
	<header>
		<div>
			{#if eyebrow}<p>{eyebrow}</p>{/if}
			<h1>{title}</h1>
		</div>
		{#if actions}{@render actions()}{/if}
	</header>
	{#if toolbar}<div class="toolbar">{@render toolbar()}</div>{/if}
	<div class="body">{@render children()}</div>
</aside>

<style>
	.item-rail {
		display: grid;
		height: 100%;
		min-width: 0;
		min-height: 0;
		grid-template-rows: auto auto minmax(0, 1fr);
		background: linear-gradient(rgb(27 18 12 / 96%), rgb(18 11 8 / 98%)), var(--wood);
		color: var(--paper-light);
	}

	header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--space-4);
	}

	h1,
	header p {
		margin: 0;
		color: var(--paper-light);
	}

	h1 {
		font-size: 1.35rem;
	}

	header p {
		color: var(--gold-light);
		font-family: var(--font-display);
		font-size: 0.64rem;
		letter-spacing: 0.12em;
		text-transform: uppercase;
	}

	.toolbar {
		min-width: 0;
	}

	.body {
		min-width: 0;
		min-height: 0;
		overflow-y: auto;
		overflow-x: hidden;
		overscroll-behavior: contain;
	}
</style>
