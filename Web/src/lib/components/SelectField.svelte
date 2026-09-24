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

<label class="form-field" for={inputId}>
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
