<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		rail,
		detail,
		detailOpen = false,
		compact = false,
		actions
	}: {
		rail: Snippet;
		detail: Snippet;
		detailOpen?: boolean;
		compact?: boolean;
		actions?: Snippet;
	} = $props();

	let actionsHeight = $state(0);
</script>

{#if actions}
	<div class="split-actions" bind:clientHeight={actionsHeight}>{@render actions()}</div>
{/if}

<div
	class="split-view"
	class:detail-open={detailOpen}
	class:compact
	style:--split-actions-height={`${actions ? actionsHeight : 0}px`}
>
	<div class="rail">{@render rail()}</div>
	<section class="detail">{@render detail()}</section>
</div>

<style>
	.split-actions {
		display: flex;
		padding-block-end: var(--space-3);
	}

	.split-view {
		display: grid;
		height: calc(100dvh - 7.25rem - var(--split-actions-height));
		min-height: calc(32rem - var(--split-actions-height));
		grid-template-columns: minmax(18rem, 0.34fr) minmax(0, 1fr);
		overflow: hidden;
		border: 1px solid var(--accent-strong);
		background: var(--surface-canvas);
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
		border-inline-end: 1px solid var(--accent-strong);
	}

	.detail {
		display: flex;
		flex-direction: column;
	}

	@media (max-width: 47.99rem) {
		.split-view {
			height: calc(100dvh - 7.75rem - env(safe-area-inset-bottom) - var(--split-actions-height));
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
