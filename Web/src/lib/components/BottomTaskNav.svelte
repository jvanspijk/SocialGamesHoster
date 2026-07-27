<script lang="ts">
	let {
		items,
		current,
		select
	}: {
		items: Array<{ id: string; label: string; attention?: boolean }>;
		current: string;
		select: (id: string) => void;
	} = $props();
</script>

<nav aria-label="Player tasks">
	{#each items as item (item.id)}
		<button
			type="button"
			class:active={current === item.id}
			aria-current={current === item.id ? 'page' : undefined}
			onclick={() => select(item.id)}
		>
			<span>{item.label}</span>
			{#if item.attention}<i aria-label="New activity"></i>{/if}
		</button>
	{/each}
</nav>

<style>
	nav {
		position: fixed;
		z-index: var(--layer-navigation);
		inset-inline: 0;
		inset-block-end: 0;
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		border-block-start: var(--border-strong);
		background: var(--paper-light);
		box-shadow: 0 -0.4rem 1rem rgb(31 19 11 / 18%);
		padding-block-end: env(safe-area-inset-bottom);
	}

	button {
		position: relative;
		min-width: 0;
		min-height: calc(var(--target-size) + var(--space-2));
		border: 0;
		border-block-start: 3px solid transparent;
		background: transparent;
		color: var(--ink-soft);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.68rem;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}

	button.active {
		border-color: var(--crimson);
		color: var(--crimson-dark);
	}

	i {
		position: absolute;
		inset-block-start: 0.45rem;
		inset-inline-end: calc(50% - 2.2rem);
		width: 0.55rem;
		height: 0.55rem;
		border: 2px solid var(--paper-light);
		border-radius: 50%;
		background: var(--crimson);
	}

	@media (min-width: 64rem) {
		nav {
			inset-inline: 0 auto;
			inset-block: 0;
			width: 6rem;
			grid-template-columns: 1fr;
			grid-template-rows: repeat(4, min-content);
			border-block-start: 0;
			border-inline-end: var(--border-strong);
			padding-block: var(--space-8);
		}

		button {
			min-height: 4rem;
			border-block-start: 0;
			border-inline-start: 3px solid transparent;
		}

		button.active {
			border-color: var(--crimson);
		}
	}
</style>
