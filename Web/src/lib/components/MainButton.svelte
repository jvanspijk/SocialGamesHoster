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
		padding: 10px 20px;
		color: #f7e7c4;
		background-color: #a62a2a;
		border: 3px solid #5b4a3c;
		border-radius: 5px;
		cursor: pointer;
		transition: background-color 0.2s, box-shadow 0.2s;
		box-shadow: 2px 2px 5px rgba(0, 0, 0, 0.4);
		letter-spacing: 1px;
		margin-top: 12px;
		margin-bottom: 12px;
	}

	button:hover:enabled {
		background-color: #7a1d1d;
		box-shadow: 3px 3px 8px rgba(0, 0, 0, 0.6);
	}

	button:disabled {
		background-color: #999;
		border-color: #777;
		cursor: not-allowed;
	}
</style>
