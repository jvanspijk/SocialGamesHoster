<script module lang="ts">
	export type SelectOption<T extends string | number> = {
		value: T;
		label: string;
		disabled?: boolean;
	};
	export type SelectOptionGroup<T extends string | number> = {
		label: string;
		options: readonly SelectOption<T>[];
	};
</script>

<script lang="ts" generics="T extends string | number">
	let {
		label,
		name,
		value = $bindable<T>(),
		options,
		required = false,
		disabled = false,
		help = '',
		error = '',
		accessibleLabel,
		onchange
	}: {
		label: string;
		name: string;
		value?: T;
		options: readonly (SelectOption<T> | SelectOptionGroup<T>)[];
		required?: boolean;
		disabled?: boolean;
		help?: string;
		error?: string;
		accessibleLabel?: string;
		onchange?: (value: T) => void;
	} = $props();

	const instanceId = $props.id();
	const inputId = $derived(`select-field-${name}-${instanceId}`);
	const descriptionId = $derived(`${inputId}-description`);

	function isGroup(option: SelectOption<T> | SelectOptionGroup<T>): option is SelectOptionGroup<T> {
		return 'options' in option;
	}
</script>

<label for={inputId}>
	<span
		>{label}{#if required}<i aria-hidden="true"> *</i>{/if}</span
	>
	<select
		id={inputId}
		{name}
		bind:value
		{required}
		{disabled}
		onchange={() => onchange?.(value)}
		aria-invalid={error ? 'true' : undefined}
		aria-label={accessibleLabel}
		aria-describedby={help || error ? descriptionId : undefined}
	>
		{#each options as option (isGroup(option) ? option.label : option.value)}
			{#if isGroup(option)}
				<optgroup label={option.label}>
					{#each option.options as child (child.value)}
						<option value={child.value} disabled={child.disabled}>{child.label}</option>
					{/each}
				</optgroup>
			{:else}
				<option value={option.value} disabled={option.disabled}>{option.label}</option>
			{/if}
		{/each}
	</select>
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

	select {
		width: 100%;
		min-height: var(--target-size);
		border: 1px solid #8d7248;
		border-radius: 1px;
		background: var(--paper-light);
		color: var(--ink);
		padding: 0.65rem 0.75rem;
		transition: border-color var(--speed-fast) ease-out;
	}

	select:hover,
	select:focus {
		border-color: var(--crimson);
	}

	select:focus-visible {
		outline: 2px solid var(--focus);
		outline-offset: 2px;
	}

	select[aria-invalid='true'] {
		border-color: var(--danger);
		box-shadow: 0 0 0 1px var(--danger);
	}

	select:disabled {
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
