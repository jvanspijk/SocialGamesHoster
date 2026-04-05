<script lang="ts">
	import type { Snippet } from 'svelte';

	type ButtonType = 'button' | 'submit' | 'reset';

	interface Props {
		children: Snippet;
		variant?: 'secondary' | 'primary' | 'danger';
		onclick?: (e: MouseEvent) => void;
		enabled?: boolean;
		type?: ButtonType;
	}

	let {
		children,
		variant = 'secondary',
		onclick,
		enabled = true,
		type = 'button'
	}: Props = $props();
</script>

<button
	{type}
	class="secondary-btn {variant}"
	{onclick}
	disabled={!enabled}
	aria-disabled={!enabled}
>
	{@render children()}
</button>

<style>
	.secondary-btn {
		font-family: var(--font-heading);
		font-weight: bold;
		padding: 2px 10px;
		cursor: pointer;
		border: 1px solid var(--color-border);
		background: var(--color-surface-soft);
		color: var(--color-border);
		text-transform: uppercase;
		font-size: 0.77rem;
		transition: all 0.1s ease;
		box-shadow: 1px 1px 0px var(--color-border);
		line-height: 2.4;
		height: fit-content;
	}

	.secondary-btn:not(:disabled):active {
		transform: translate(0.5px, 0.5px);
		box-shadow: inset 1px 1px 2px rgba(0, 0, 0, 0.1);
	}

	.secondary-btn:not(:disabled).primary:hover {
		background: var(--color-border);
		color: var(--color-surface);
	}

	.secondary-btn:not(:disabled).danger:hover {
		background: var(--color-accent-strong);
		color: white;
	}

	.danger {
		color: var(--color-accent-strong);
		border-color: var(--color-accent-strong);
	}

	.secondary-btn:disabled {
		cursor: not-allowed;
		filter: grayscale(0.8);
		opacity: 0.5;
		box-shadow: 1px 1px 0px var(--color-border);
		border-style: dashed;
	}
</style>
