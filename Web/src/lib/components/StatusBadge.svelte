<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		label,
		tone = 'success',
		dot = true,
		icon
	}: {
		label: string;
		tone?: 'success' | 'warning' | 'danger' | 'info';
		dot?: boolean;
		icon?: Snippet;
	} = $props();
</script>

<span
	class="status-badge"
	class:tone-success={tone === 'success'}
	class:tone-warning={tone === 'warning'}
	class:tone-danger={tone === 'danger'}
	class:tone-info={tone === 'info'}
>
	{#if icon}<span class="icon" aria-hidden="true">{@render icon()}</span>{:else if dot}<i
			aria-hidden="true"
		></i>{/if}
	{label}
</span>

<style>
	.status-badge {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		color: var(--success);
		font-family: var(--font-display);
		font-size: 0.65rem;
		letter-spacing: 0.09em;
		text-transform: uppercase;
	}
	i {
		width: 0.5rem;
		height: 0.5rem;
		border-radius: 50%;
		background: currentColor;
	}
	.icon {
		display: inline-flex;
	}
	.tone-warning {
		color: var(--gold);
	}
	.tone-danger {
		color: var(--danger);
	}
	.tone-info {
		color: var(--information);
	}
</style>
