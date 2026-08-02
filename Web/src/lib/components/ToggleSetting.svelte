<script lang="ts">
	import CheckboxField from './CheckboxField.svelte';

	let {
		title,
		description,
		name,
		checked = $bindable(false),
		disabled = false,
		onchange
	}: {
		title: string;
		description: string;
		name: string;
		checked?: boolean;
		disabled?: boolean;
		onchange?: (checked: boolean) => void;
	} = $props();
</script>

<CheckboxField label={title} {description} {name} bind:checked {disabled} {onchange} />

<style>
	:global(label) {
		border-inline: 0;
		border-block-start: 0;
		padding-inline: 0;
	}

	:global(strong) {
		font-family: var(--font-display);
		font-size: 0.78rem;
		letter-spacing: 0.04em;
	}
</style>
