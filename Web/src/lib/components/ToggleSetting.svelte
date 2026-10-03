<script lang="ts">
	import CheckboxField from './CheckboxField.svelte';

	let {
		title,
		description = '',
		name,
		checked = $bindable(false),
		disabled = false,
		onchange
	}: {
		title: string;
		description?: string;
		name: string;
		checked?: boolean;
		disabled?: boolean;
		onchange?: (checked: boolean) => void;
	} = $props();
</script>

<CheckboxField
	label={title}
	{description}
	{name}
	bind:checked
	{disabled}
	{onchange}
	variant="setting"
/>
