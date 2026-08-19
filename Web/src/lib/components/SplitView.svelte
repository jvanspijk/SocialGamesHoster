<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		rail,
		detail,
		detailOpen = false,
		compact = false
	}: {
		rail: Snippet;
		detail: Snippet;
		detailOpen?: boolean;
		compact?: boolean;
	} = $props();
</script>

<div class="split-view" class:detail-open={detailOpen} class:compact>
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

	.compact {
		height: auto;
		min-height: 28rem;
		max-height: min(46rem, calc(100dvh - 12rem));
		grid-template-columns: minmax(13rem, 0.32fr) minmax(0, 1fr);
	}

	.compact .rail,
	.compact .detail {
		overflow: auto;
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

		.split-view.compact {
			height: auto;
			min-height: 28rem;
			max-height: none;
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
