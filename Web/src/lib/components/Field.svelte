<script lang="ts">
	let {
		label,
		name,
		value = $bindable(''),
		type = 'text',
		placeholder = '',
		required = false,
		autocomplete,
		multiline = false,
		help = ''
	}: {
		label: string;
		name: string;
		value?: string;
		type?: string;
		placeholder?: string;
		required?: boolean;
		autocomplete?: 'username' | 'new-password' | 'current-password' | 'nickname' | 'off';
		multiline?: boolean;
		help?: string;
	} = $props();
</script>

<label>
	<span>{label}</span>
	{#if multiline}
		<textarea {name} bind:value {placeholder} {required} rows="4"></textarea>
	{:else}
		<input {name} bind:value {type} {placeholder} {required} {autocomplete} />
	{/if}
	{#if help}<small>{help}</small>{/if}
</label>

<style>
	label {
		display: grid;
		gap: 0.3rem;
	}

	span {
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	input,
	textarea {
		width: 100%;
		min-height: 44px;
		border: 1px solid #8d7248;
		border-radius: 1px;
		background: var(--paper-light);
		color: var(--ink);
		padding: 0.65rem 0.75rem;
		transition: border-color var(--speed-fast) ease-out;
	}

	input:hover,
	textarea:hover,
	input:focus,
	textarea:focus {
		border-color: var(--crimson);
	}

	small {
		color: var(--ink-soft);
	}
</style>
