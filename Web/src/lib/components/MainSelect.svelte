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
</script>

<select name={name} bind:value={selectedValue}>
	<option value={null} disabled selected={selectedValue === null}>
		{placeholder}
	</option>

	{#each options as option (option.id)}
		<option value={option.id}>{option.label}</option>
	{/each}
</select>

<style>
	select {
		font-family: 'IM Fell English', serif;
		font-size: 1.2em;
		padding: 10px 15px;
		background-color: #fff9e6;
		border: 2px solid #5b4a3c;
		border-radius: 3px;
		color: #3e322b;
		width: 100%;
		max-width: 400px;
		box-shadow: inset 0 1px 3px rgba(0, 0, 0, 0.2);
		appearance: none;
		background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" fill="%235b4a3c" viewBox="0 0 10 10"><path d="M5 8l-5-5h10z"/></svg>');
		background-repeat: no-repeat;
		background-position: right 15px center;
		cursor: pointer;
		margin-top: 20px;
		margin-bottom: 20px;
	}
</style>
