<script lang="ts">
	export let label: string;
	export let disabled = false;
	export let isLoading = false;
	export let type: 'button' | 'submit' | 'reset' = 'button';
	export let onActivate: (() => void) | null = null;

	function handleClick(event: MouseEvent) {
		if (onActivate) {
			event.preventDefault();
			onActivate();
		}
	}
</script>

<button
	type={type}
	disabled={disabled || isLoading}
	on:click={handleClick}
>
	{#if isLoading}
		Loading...
	{:else}
		{label}
	{/if}
</button>

<style>
	button {
		font-family: var(--font-heading);
		font-size: 1.2em;
		text-transform: uppercase;
		letter-spacing: 1.5px;
		padding: 12px 24px;
		margin: 12px 0;

		color: var(--color-surface);
		background-color: var(--color-accent-strong);
		border: none;
		border-radius: 5px;

		cursor: pointer;
		transition: all 0.2s ease-out;

		box-shadow:
			0 4px 0 0 var(--color-accent),
			0 5px 8px rgba(0, 0, 0, 0.5);
	}

	button:hover:enabled {
		background-color: var(--color-accent-strong);
		transform: translateY(-1px);
		box-shadow:
			0 5px 0 0 var(--color-accent),
			0 7px 10px rgba(0, 0, 0, 0.6);
	}

	button:active:enabled {
		background-color: var(--color-accent);
		transform: translateY(3px);
		box-shadow:
			0 1px 0 0 var(--color-accent),
			0 2px 5px rgba(0, 0, 0, 0.4);
	}

	button:focus-visible {
		outline: 3px solid var(--color-surface);
		outline-offset: 2px;
	}

	button:disabled {
		background-color: var(--color-border);
		color: var(--color-text-disabled);
		cursor: not-allowed;
		transform: none;
		box-shadow:
			0 4px 0 0 var(--color-border-dark),
			0 2px 4px rgba(0, 0, 0, 0.3);
	}
</style>
