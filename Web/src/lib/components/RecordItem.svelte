<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		children,
		trailing,
		supporting
	}: {
		children: Snippet;
		trailing?: Snippet;
		supporting?: Snippet;
	} = $props();
</script>

<li class="record-item" class:has-trailing={trailing} class:has-supporting={supporting}>
	<div class="main">{@render children()}</div>
	{#if trailing}<div class="trailing">{@render trailing()}</div>{/if}
	{#if supporting}<div class="supporting">{@render supporting()}</div>{/if}
</li>

<style>
	.record-item {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		align-items: center;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding-block: var(--space-3);
	}

	.record-item.has-trailing {
		grid-template-columns: minmax(0, 1fr) auto;
	}

	.record-item.has-supporting {
		grid-template-columns: minmax(0, 1fr) minmax(10rem, auto);
	}

	.record-item.has-trailing.has-supporting {
		grid-template-columns: minmax(0, 1fr) auto minmax(10rem, auto);
	}

	.main,
	.trailing,
	.supporting {
		min-width: 0;
	}

	.trailing,
	.supporting {
		display: flex;
		justify-content: flex-end;
	}

	@media (max-width: 47.99rem) {
		.record-item.has-supporting,
		.record-item.has-trailing.has-supporting {
			grid-template-columns: minmax(0, 1fr) auto;
		}

		.record-item.has-supporting .supporting {
			grid-column: 1 / -1;
			justify-content: flex-start;
		}
	}
</style>
