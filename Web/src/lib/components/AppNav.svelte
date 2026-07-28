<script lang="ts">
	import type { Component } from 'svelte';

	let {
		items,
		current,
		label
	}: {
		items: Array<{
			id: string;
			label: string;
			href: string;
			icon: Component<{ size?: number; strokeWidth?: number }>;
			attention?: boolean;
			disabled?: boolean;
			disabledDescription?: string;
		}>;
		current: string;
		label: string;
	} = $props();
</script>

<!-- Dynamic destinations are resolved by each shell before they reach this reusable primitive. -->
<!-- eslint-disable svelte/no-navigation-without-resolve -->
<nav aria-label={label} style={`--nav-count: ${items.length}`}>
	{#each items as item (item.id)}
		{@const Icon = item.icon}
		<a
			href={item.disabled ? undefined : item.href}
			class:active={current === item.id}
			class:disabled={item.disabled}
			aria-current={current === item.id ? 'page' : undefined}
			aria-disabled={item.disabled || undefined}
			aria-describedby={item.disabled ? `nav-description-${item.id}` : undefined}
			tabindex={item.disabled ? -1 : undefined}
		>
			<span class="icon"><Icon size={21} strokeWidth={1.8} /></span>
			<span>{item.label}</span>
			{#if item.attention}<i aria-label="New activity"></i>{/if}
			{#if item.disabled && item.disabledDescription}
				<em class="sr-only" id={`nav-description-${item.id}`}>{item.disabledDescription}</em>
			{/if}
		</a>
	{/each}
</nav>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	nav {
		position: fixed;
		z-index: var(--layer-navigation);
		inset-inline: 0;
		inset-block-end: 0;
		display: grid;
		grid-template-columns: repeat(var(--nav-count, 4), minmax(0, 1fr));
		border-block-start: 1px solid var(--gold-dark);
		background: linear-gradient(rgb(28 18 12 / 96%), rgb(20 12 8 / 98%)), var(--wood);
		box-shadow: 0 -0.35rem 1.25rem rgb(19 10 6 / 25%);
		padding-block-end: env(safe-area-inset-bottom);
	}

	a {
		position: relative;
		display: grid;
		min-width: 0;
		min-height: calc(var(--target-size) + var(--space-2));
		place-items: center;
		align-content: center;
		gap: 0.1rem;
		border-block-start: 3px solid transparent;
		color: var(--paper-muted);
		font-family: var(--font-display);
		font-size: 0.64rem;
		font-weight: 700;
		letter-spacing: 0.04em;
		text-decoration: none;
		text-transform: uppercase;
	}

	a.active {
		border-color: var(--crimson-light);
		background: rgb(145 36 38 / 18%);
		color: var(--gold-light);
	}

	a.disabled {
		cursor: not-allowed;
		opacity: 0.52;
	}

	.icon {
		display: grid;
		place-items: center;
	}

	i {
		position: absolute;
		inset-block-start: 0.45rem;
		inset-inline-end: calc(50% - 1.7rem);
		width: 0.55rem;
		height: 0.55rem;
		border: 2px solid #1c120c;
		border-radius: 50%;
		background: var(--crimson-light);
	}

	@media (min-width: 64rem) {
		nav {
			inset: 0 auto 0 0;
			width: 13rem;
			grid-template-columns: 1fr;
			grid-auto-rows: min-content;
			align-content: start;
			gap: var(--space-1);
			border-block-start: 0;
			border-inline-end: 1px solid var(--gold-dark);
			padding: var(--space-8) var(--space-2);
		}

		a {
			grid-template-columns: 2rem 1fr;
			min-height: var(--target-size);
			justify-items: start;
			align-content: center;
			gap: var(--space-2);
			border-block-start: 0;
			border-inline-start: 3px solid transparent;
			font-size: 0.72rem;
			padding-inline: var(--space-3);
		}

		i {
			inset-inline: auto var(--space-3);
			inset-block-start: 50%;
			transform: translateY(-50%);
		}
	}
</style>
