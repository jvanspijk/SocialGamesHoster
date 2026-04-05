<script lang="ts" context="module">
	export type SelectOption<T = number | string> = {
		id: T;
		label: string;
	};
</script>

<script lang="ts" generics="T">
	export let options: { id: T; label: string }[] = [];
	export let placeholder: string = 'Select an option';
	export let name: string = '';
	export let selectedValue: T | null = null;
	export let enabled: boolean = true;
</script>

<select {name} bind:value={selectedValue} disabled={!enabled}>
	<option value={null} disabled selected={selectedValue === null}>
		{placeholder}
	</option>

	{#each options as option (option.id)}
		<option value={option.id}>{option.label}</option>
	{/each}
</select>

<style>
	select {
		font-family: var(--font-body);
		font-size: 1.2em;
		padding: 10px 15px;
		background-color: var(--color-surface-soft);
		border: 2px solid var(--color-border);
		border-radius: 3px;
		color: var(--color-text);
		display: block;
		width: 320px;
		max-width: 320px;
		box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.2);
		appearance: none;
		background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" fill="%235b4a3c" viewBox="0 0 10 10"><path d="M5 8l-5-5h10z"/></svg>');
		background-repeat: no-repeat;
		background-position: right 15px center;
		cursor: pointer;
		margin-top: 20px;
		margin-bottom: 20px;
		transition: all 0.2s ease;
	}

	select:disabled {
		background-color: var(--color-surface-alt);
		border-color: var(--color-border-soft);
		color: var(--color-text-muted);
		cursor: not-allowed;
		filter: grayscale(0.3);
		opacity: 0.7;
	}
</style>
