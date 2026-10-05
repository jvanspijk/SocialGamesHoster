<script lang="ts">
	let { count, placement = 'corner' }: { count?: number; placement?: 'corner' | 'navigation' } =
		$props();
</script>

<span
	class="attention-badge"
	class:count={count !== undefined}
	class:navigation={placement === 'navigation'}
	aria-hidden="true">{count === 99 ? '99+' : (count ?? 'New')}</span
>

<style>
	.attention-badge {
		position: absolute;
		inset-inline-end: 0.2rem;
		inset-block-start: 0.2rem;
		display: grid;
		min-width: 1rem;
		height: 1rem;
		place-items: center;
		border: 2px solid var(--surface-surround);
		border-radius: 50%;
		background: var(--action-dark);
		color: var(--text-on-dark);
		font-family: var(--font-display);
		font-size: var(--font-size-xs);
		font-weight: 700;
		line-height: 1;
	}
	.count {
		min-width: 1.3rem;
		border-radius: 999px;
		padding-inline: 0.15rem;
		text-align: center;
	}
	.navigation {
		inset-inline-end: calc(50% - 1.7rem);
		inset-block-start: 0.45rem;
	}
	@media (min-width: 64rem) {
		.navigation {
			inset-inline: auto var(--space-3);
			inset-block-start: 50%;
			transform: translateY(-50%);
		}
	}
</style>
