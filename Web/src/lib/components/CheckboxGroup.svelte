<script module lang="ts">
	export type CheckboxOption<T extends string> = {
		value: T;
		label: string;
		disabled?: boolean;
	};
</script>

<script lang="ts" generics="T extends string">
	let {
		label,
		name,
		values = $bindable<T[]>([]),
		options,
		help = '',
		error = '',
		required = false,
		disabled = false
	}: {
		label: string;
		name: string;
		values?: T[];
		options: readonly CheckboxOption<T>[];
		help?: string;
		error?: string;
		required?: boolean;
		disabled?: boolean;
	} = $props();

	const instanceId = $props.id();
	const groupId = $derived(`checkbox-group-${name}-${instanceId}`);
	const descriptionId = $derived(`${groupId}-description`);
	let validationInputs = $state<HTMLInputElement[]>([]);

	$effect(() => {
		for (const input of validationInputs) input?.setCustomValidity('');
		const validationInput = validationInputs.find((input) => input && !input.disabled);
		validationInput?.setCustomValidity(
			required && !disabled && values.length === 0 ? 'Select at least one option.' : ''
		);
	});
</script>

<fieldset aria-describedby={help || error ? descriptionId : undefined}>
	<legend
		>{label}{#if required}<i aria-hidden="true"> *</i>{/if}</legend
	>
	<div class="choices">
		{#each options as option, index (option.value)}
			<label for={`${groupId}-${index}`}>
				<input
					id={`${groupId}-${index}`}
					{name}
					type="checkbox"
					value={option.value}
					bind:group={values}
					disabled={disabled || option.disabled}
					aria-invalid={error ? 'true' : undefined}
					bind:this={validationInputs[index]}
				/>
				{option.label}
			</label>
		{/each}
	</div>
	{#if error}
		<small id={descriptionId} class="error">{error}</small>
	{:else if help}
		<small id={descriptionId}>{help}</small>
	{/if}
</fieldset>

<style>
	fieldset {
		display: grid;
		gap: var(--space-2);
		min-width: 0;
		border: 0;
		margin: 0;
		padding: 0;
	}

	legend {
		font-family: var(--font-display);
		font-size: 0.67rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		padding: 0;
		text-transform: uppercase;
	}

	i {
		color: var(--danger);
		font-style: normal;
	}

	.choices {
		display: flex;
		flex-wrap: wrap;
		gap: 0.45rem 0.9rem;
	}

	.choices label {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: 0.4rem;
		padding-inline: var(--space-1);
		cursor: pointer;
	}

	.choices label:has(input:focus-visible) {
		outline: 2px solid var(--focus);
		outline-offset: 2px;
	}

	input {
		width: 1.35rem;
		height: 1.35rem;
		accent-color: var(--crimson);
	}

	input:disabled {
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
