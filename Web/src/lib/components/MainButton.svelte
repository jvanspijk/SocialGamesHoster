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
		font-family: 'Cinzel', serif;
		font-size: 1.2em;
		text-transform: uppercase;
		letter-spacing: 1.5px;
		padding: 12px 24px;
		margin: 12px 0;

		color: #f7e7c4;
		background-color: #a62a2a;
		border: none;
		border-radius: 5px;

		cursor: pointer;
		transition: all 0.2s ease-out;

		box-shadow:
			0 4px 0 0 #7a1d1d,
			0 5px 8px rgba(0, 0, 0, 0.5);
	}

	button:hover:enabled {
		background-color: #b83a3a;
		transform: translateY(-1px);
		box-shadow:
			0 5px 0 0 #7a1d1d,
			0 7px 10px rgba(0, 0, 0, 0.6);
	}

	button:active:enabled {
		background-color: #7a1d1d;
		transform: translateY(3px);
		box-shadow:
			0 1px 0 0 #7a1d1d,
			0 2px 5px rgba(0, 0, 0, 0.4);
	}

	button:focus-visible {
		outline: 3px solid #f7e7c4;
		outline-offset: 2px;
	}

	button:disabled {
		background-color: #5b4a3c;
		color: #b3a598;
		cursor: not-allowed;
		transform: none;
		box-shadow:
			0 4px 0 0 #3e332a,
			0 2px 4px rgba(0, 0, 0, 0.3);
	}
</style>
