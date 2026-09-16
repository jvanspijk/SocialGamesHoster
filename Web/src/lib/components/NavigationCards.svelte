<script lang="ts">
	import type { Component } from 'svelte';

	let {
		items
	}: {
		items: Array<{
			label: string;
			description: string;
			href?: string;
			onclick?: () => void | Promise<void>;
			disabled?: boolean;
			icon: Component<{ size?: number; strokeWidth?: number }>;
		}>;
	} = $props();
</script>

<nav aria-label="Home shortcuts">
	{#each items as item (`${item.label}:${item.href ?? 'action'}`)}
		{@const Icon = item.icon}
		{#if item.href}
			<!-- Destinations are resolved by the route that owns each shortcut. -->
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a href={item.href} aria-disabled={item.disabled || undefined} class:disabled={item.disabled}>
				<span class="icon" aria-hidden="true"><Icon size={25} strokeWidth={1.7} /></span>
				<span><strong>{item.label}</strong><small>{item.description}</small></span>
			</a>
		{:else}
			<button type="button" onclick={item.onclick} disabled={item.disabled}>
				<span class="icon" aria-hidden="true"><Icon size={25} strokeWidth={1.7} /></span>
				<span><strong>{item.label}</strong><small>{item.description}</small></span>
			</button>
		{/if}
	{/each}
</nav>

<style>
	nav {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--space-3);
	}

	a,
	button {
		display: grid;
		width: 100%;
		min-width: 0;
		min-height: 6.5rem;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: center;
		gap: var(--space-3);
		border: var(--border-subtle);
		background: rgb(255 249 230 / 62%);
		color: var(--ink);
		cursor: pointer;
		font: inherit;
		padding: var(--space-4);
		text-decoration: none;
		text-align: start;
		transition:
			transform var(--speed-fast) ease-out,
			box-shadow var(--speed-fast) ease-out;
	}

	a:hover,
	button:hover:not(:disabled) {
		box-shadow: var(--shadow-small);
		transform: translateY(-2px);
	}

	a:focus-visible,
	button:focus-visible {
		outline: var(--focus-ring);
		outline-offset: 3px;
	}

	.disabled,
	button:disabled {
		cursor: not-allowed;
		opacity: 0.58;
	}

	.icon {
		display: grid;
		width: 2.75rem;
		height: 2.75rem;
		place-items: center;
		border: 1px solid var(--gold-dark);
		color: var(--crimson-dark);
	}

	strong,
	small {
		display: block;
	}

	strong {
		font-family: var(--font-display);
		font-size: 0.8rem;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}

	small {
		color: var(--ink-soft);
		font-size: 0.88rem;
	}

	@media (max-width: 35rem) {
		nav {
			grid-template-columns: 1fr;
		}
	}
</style>
