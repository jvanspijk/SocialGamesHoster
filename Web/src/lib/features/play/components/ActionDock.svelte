<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();
</script>

<div class="action-dock">
	{@render children()}
</div>

<style>
	.action-dock {
		position: fixed;
		z-index: var(--layer-sticky);
		inset-inline: 0;
		inset-block-end: calc(4rem + env(safe-area-inset-bottom));
		display: flex;
		justify-content: center;
		background: linear-gradient(transparent, var(--paper) 38%);
		padding: var(--space-5) var(--space-3) var(--space-3);
	}

	.action-dock :global(button) {
		width: min(100%, 24rem);
	}

	@media (min-width: 64rem) {
		.action-dock {
			inset-inline-start: 13rem;
			inset-block-end: 0;
		}
	}
</style>
