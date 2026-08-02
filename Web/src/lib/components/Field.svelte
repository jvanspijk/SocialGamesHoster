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
		help = '',
		error = '',
		disabled = false,
		onchange
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
		error?: string;
		disabled?: boolean;
		onchange?: (value: string) => void;
	} = $props();

	const inputId = $derived(`field-${name}`);
	const descriptionId = $derived(`${inputId}-description`);
</script>

<label>
	<span
		>{label}{#if required}<i aria-hidden="true"> *</i>{/if}</span
	>
	{#if multiline}
		<textarea
			id={inputId}
			{name}
			bind:value
			{placeholder}
			{required}
			{disabled}
			aria-invalid={error ? 'true' : undefined}
			aria-describedby={help || error ? descriptionId : undefined}
			rows="4"
			onchange={(event) => onchange?.((event.currentTarget as HTMLTextAreaElement).value)}
		></textarea>
	{:else}
		<input
			id={inputId}
			{name}
			bind:value
			{type}
			{placeholder}
			{required}
			{disabled}
			{autocomplete}
			onchange={(event) => onchange?.((event.currentTarget as HTMLInputElement).value)}
			aria-invalid={error ? 'true' : undefined}
			aria-describedby={help || error ? descriptionId : undefined}
		/>
	{/if}
	{#if error}
		<small id={descriptionId} class="error">{error}</small>
	{:else if help}
		<small id={descriptionId}>{help}</small>
	{/if}
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

	i {
		color: var(--danger);
		font-style: normal;
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

	input[aria-invalid='true'],
	textarea[aria-invalid='true'] {
		border-color: var(--danger);
		box-shadow: 0 0 0 1px var(--danger);
	}

	input:disabled,
	textarea:disabled {
		background: var(--paper-deep);
		cursor: not-allowed;
		opacity: 0.72;
	}

	small {
		color: var(--ink-soft);
	}

	small.error {
		color: var(--danger);
	}
</style>
