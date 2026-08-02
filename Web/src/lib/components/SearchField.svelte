<script lang="ts">
	import { Search } from '@lucide/svelte';

	let {
		label,
		placeholder = '',
		value = $bindable(''),
		variant = 'default'
	}: {
		label: string;
		placeholder?: string;
		value?: string;
		variant?: 'default' | 'inverse';
	} = $props();
</script>

<label class="search-field" class:inverse={variant === 'inverse'}>
	<Search size={18} aria-hidden="true" />
	<span class="sr-only">{label}</span>
	<input type="search" bind:value {placeholder} aria-label={label} />
</label>

<style>
	.search-field {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: center;
		gap: var(--space-2);
		margin: 0 var(--space-3) var(--space-3);
		border: 1px solid #8d7248;
		background: var(--paper-light);
		color: var(--ink);
		padding-inline: var(--space-2);
	}

	.search-field input {
		min-width: 0;
		min-height: var(--target-size);
		border: 0;
		background: transparent;
		color: inherit;
		outline: 0;
	}

	.search-field input::placeholder {
		color: var(--ink-soft);
	}

	.search-field.inverse {
		border-color: #755d43;
		background: rgb(255 255 255 / 7%);
		color: var(--paper-light);
	}

	.search-field.inverse input::placeholder {
		color: var(--paper-muted);
	}
</style>
