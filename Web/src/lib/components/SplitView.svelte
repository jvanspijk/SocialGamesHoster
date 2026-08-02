<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		rail,
		detail,
		detailOpen = false
	}: {
		rail: Snippet;
		detail: Snippet;
		detailOpen?: boolean;
	} = $props();
</script>

<div class="split-view" class:detail-open={detailOpen}>
	<div class="rail">{@render rail()}</div>
	<section class="detail">{@render detail()}</section>
</div>

<style>
	.split-view {
		display: grid;
		height: calc(100dvh - 7.25rem);
		min-height: 32rem;
		grid-template-columns: minmax(18rem, 0.34fr) minmax(0, 1fr);
		overflow: hidden;
		border: 1px solid var(--gold-dark);
		background: var(--paper);
		box-shadow: var(--shadow-small);
	}

	.rail,
	.detail {
		min-width: 0;
		min-height: 0;
	}

	.rail {
		border-inline-end: 1px solid var(--gold-dark);
	}

	.detail {
		display: grid;
		grid-template-rows: minmax(0, 1fr);
	}

	@media (max-width: 47.99rem) {
		.split-view {
			height: calc(100dvh - 7.75rem - env(safe-area-inset-bottom));
			min-height: 0;
			grid-template-columns: 1fr;
			border-inline: 0;
		}

		.rail,
		.detail {
			grid-column: 1;
			grid-row: 1;
		}

		.split-view:not(.detail-open) .detail {
			display: none;
		}

		.split-view.detail-open .rail {
			display: none;
		}

		.rail {
			border: 0;
		}
	}
</style>
