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

<label class="form-field">
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
