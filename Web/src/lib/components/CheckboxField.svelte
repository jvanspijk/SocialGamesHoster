<script lang="ts">
	let {
		label,
		name,
		checked = $bindable(false),
		description = '',
		required = false,
		disabled = false,
		error = '',
		onchange
	}: {
		label: string;
		name: string;
		checked?: boolean;
		description?: string;
		required?: boolean;
		disabled?: boolean;
		error?: string;
		onchange?: (checked: boolean) => void;
	} = $props();

	const instanceId = $props.id();
	const inputId = $derived(`checkbox-field-${name}-${instanceId}`);
	const descriptionId = $derived(`${inputId}-description`);
</script>

<label for={inputId}>
	<input
		id={inputId}
		{name}
		type="checkbox"
		bind:checked
		{required}
		{disabled}
		onchange={() => onchange?.(checked)}
		aria-invalid={error ? 'true' : undefined}
		aria-describedby={description || error ? descriptionId : undefined}
	/>
	<span>
		<strong
			>{label}{#if required}<i aria-hidden="true"> *</i>{/if}</strong
		>
		{#if error}
			<small id={descriptionId} class="error">{error}</small>
		{:else if description}
			<small id={descriptionId}>{description}</small>
		{/if}
	</span>
</label>

<style>
	label {
		display: grid;
		grid-template-columns: var(--target-size) minmax(0, 1fr);
		align-items: start;
		gap: var(--space-2);
		min-height: var(--target-size);
		border: 1px solid #b89b6d;
		padding: var(--space-2);
		cursor: pointer;
	}

	label:has(input:focus-visible) {
		outline: 2px solid var(--focus);
		outline-offset: 2px;
	}

	input {
		width: 1.35rem;
		height: 1.35rem;
		margin: 0.15rem auto;
		accent-color: var(--crimson);
	}

	input:disabled {
		cursor: not-allowed;
		opacity: 0.72;
	}

	span {
		display: grid;
		gap: var(--space-1);
	}

	strong {
		font: inherit;
	}

	i {
		color: var(--danger);
		font-style: normal;
	}

	small {
		color: var(--ink-soft);
		line-height: 1.4;
	}

	small.error {
		color: var(--danger);
	}
</style>
