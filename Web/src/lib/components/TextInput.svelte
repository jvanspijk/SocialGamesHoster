<script lang="ts">
	let {
		value = $bindable(''),
		placeholder = 'Enter text...',
		name = '',
		type = 'text',
		multiline = false,
		rows = 3,
		onkeydown = undefined
	} = $props();
</script>

{#if multiline}
	<textarea {name} {placeholder} {rows} bind:value {onkeydown}></textarea>
{:else}
	<input {type} {name} {placeholder} bind:value {onkeydown} />
{/if}

<style>
	input,
	textarea {
		font-family: var(--font-body);
		font-size: 1.2em;
		color: var(--color-text);
		padding: 10px 15px;
		width: 100%;
		max-width: 400px;
		background-color: var(--color-surface-soft);
		border: 2px solid var(--color-border);
		border-radius: 3px;
		box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.2);
		margin-top: 10px;
		margin-bottom: 10px;
		display: block;
		box-sizing: border-box;
	}

	textarea {
		resize: vertical;
	}

	input::placeholder,
	textarea::placeholder {
		color: var(--color-text-muted);
		opacity: 1;
	}

	input:focus,
	textarea:focus {
		outline: none;
		border-color: var(--color-accent);
		box-shadow: 0 0 5px rgba(140, 42, 62, 0.5);
	}
</style>
