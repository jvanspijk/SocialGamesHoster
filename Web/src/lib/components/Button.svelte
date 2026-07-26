<script lang="ts">
	let {
		children,
		variant = 'primary',
		type = 'button',
		disabled = false,
		loading = false,
		onclick
	}: {
		children: import('svelte').Snippet;
		variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
		type?: 'button' | 'submit';
		disabled?: boolean;
		loading?: boolean;
		onclick?: (event: MouseEvent) => void;
	} = $props();
</script>

<button
	class:loading
	class:danger={variant === 'danger'}
	class:ghost={variant === 'ghost'}
	class:secondary={variant === 'secondary'}
	{type}
	disabled={disabled || loading}
	{onclick}
>
	{#if loading}<span class="spinner" aria-hidden="true"></span>{/if}
	{@render children()}
</button>

<style>
	button {
		display: inline-flex;
		min-height: 44px;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		border: 1px solid var(--crimson-dark);
		border-radius: 1px;
		background: var(--crimson);
		box-shadow: 0 3px 0 #6e1c1c;
		color: var(--paper-light);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.78rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		line-height: 1.1;
		padding: 0.72rem 1rem;
		text-transform: uppercase;
		transition:
			transform var(--speed-fast) ease-out,
			box-shadow var(--speed-fast) ease-out,
			background var(--speed-fast) ease-out;
	}

	button:hover:not(:disabled) {
		background: var(--crimson-dark);
		transform: translateY(-1px);
	}

	button:active:not(:disabled) {
		box-shadow: 0 1px 0 #6e1c1c;
		transform: translateY(2px);
	}

	button:disabled {
		cursor: not-allowed;
		opacity: 0.56;
	}

	.secondary {
		border-color: var(--ink);
		background: var(--paper-light);
		box-shadow: 0 3px 0 #9c8257;
		color: var(--ink);
	}

	.ghost {
		border-color: transparent;
		background: transparent;
		box-shadow: none;
		color: var(--crimson-dark);
	}

	.danger {
		background: var(--danger);
	}

	.spinner {
		width: 0.9rem;
		height: 0.9rem;
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
</style>
