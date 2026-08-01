<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		label,
		icon,
		variant = 'default',
		type = 'button',
		autofocus = false,
		disabled = false,
		loading = false,
		onclick
	}: {
		label: string;
		icon: Snippet;
		variant?: 'default' | 'ghost' | 'danger';
		type?: 'button' | 'submit';
		autofocus?: boolean;
		disabled?: boolean;
		loading?: boolean;
		onclick?: (event: MouseEvent) => void;
	} = $props();
</script>

<!-- svelte-ignore a11y_autofocus -->
<button
	aria-busy={loading || undefined}
	aria-label={loading ? `Loading ${label}` : label}
	class:danger={variant === 'danger'}
	class:ghost={variant === 'ghost'}
	class:loading
	data-variant={variant}
	{type}
	{autofocus}
	disabled={disabled || loading}
	{onclick}
>
	{#if loading}
		<span class="spinner" aria-hidden="true"></span>
	{:else}
		{@render icon()}
	{/if}
</button>

<style>
	button {
		display: inline-grid;
		box-sizing: border-box;
		width: var(--target-size);
		height: var(--target-size);
		flex: 0 0 auto;
		place-items: center;
		border: var(--border-subtle);
		border-radius: 1px;
		background: var(--paper-light);
		color: var(--ink);
		cursor: pointer;
		transition:
			background var(--speed-fast) ease-out,
			box-shadow var(--speed-fast) ease-out,
			transform var(--speed-fast) ease-out;
	}

	button:hover:not(:disabled) {
		background: var(--paper-deep);
		transform: translateY(-1px);
	}

	button:active:not(:disabled) {
		transform: translateY(1px);
	}

	button:focus-visible {
		outline: 3px solid var(--focus);
		outline-offset: 2px;
	}

	button:disabled {
		cursor: not-allowed;
		opacity: 0.56;
	}

	.ghost {
		border-color: transparent;
		background: transparent;
		color: inherit;
	}

	.ghost:hover:not(:disabled) {
		background: color-mix(in srgb, currentColor 12%, transparent);
	}

	.danger {
		border-color: var(--danger);
		background: var(--danger);
		color: var(--paper-light);
	}

	.danger:hover:not(:disabled) {
		background: color-mix(in srgb, var(--danger) 82%, var(--ink));
	}

	.spinner {
		width: 1rem;
		height: 1rem;
		border: 2px solid currentColor;
		border-right-color: transparent;
		border-radius: 50%;
		animation: spin 700ms linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(1turn);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.spinner {
			animation-duration: 1ms;
		}
	}
</style>
