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
		--button-background: var(--action);
		--button-background-hover: var(--action-dark);
		--button-border: var(--action-dark);
		--button-foreground: var(--text-on-dark);
		--button-shadow: 0 3px 0 var(--action-dark);
		--button-shadow-active: 0 1px 0 var(--action-dark);
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		border: 1px solid var(--button-border);
		border-radius: var(--control-radius);
		background: var(--button-background);
		box-shadow: var(--button-shadow);
		color: var(--button-foreground);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: var(--font-size-sm);
		font-weight: 700;
		line-height: 1.1;
		padding: 0.72rem 1rem;
		transition:
			transform var(--speed-fast) ease-out,
			box-shadow var(--speed-fast) ease-out,
			background var(--speed-fast) ease-out;
	}

	button:hover:not(:disabled) {
		background: var(--button-background-hover);
		transform: translateY(-1px);
	}

	button:active:not(:disabled) {
		box-shadow: var(--button-shadow-active);
		transform: translateY(2px);
	}

	button:disabled {
		cursor: not-allowed;
		opacity: var(--action-disabled-opacity);
	}

	.secondary {
		--button-border: var(--text-primary);
		--button-background: var(--surface-raised);
		--button-foreground: var(--text-primary);
		--button-shadow: 0 3px 0 var(--control-secondary-shadow);
	}

	.ghost {
		--button-border: transparent;
		--button-background: transparent;
		--button-foreground: var(--action-dark);
		--button-shadow: none;
	}

	.danger {
		--button-background: var(--status-danger);
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
