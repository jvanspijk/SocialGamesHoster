<script lang="ts">
	import { CircleAlert, CircleCheck, Info, TriangleAlert } from '@lucide/svelte';
	import Button from './Button.svelte';

	let {
		tone = 'info',
		title,
		message,
		actionLabel,
		onaction
	}: {
		tone?: 'error' | 'success' | 'warning' | 'info';
		title?: string;
		message: string;
		actionLabel?: string;
		onaction?: () => void;
	} = $props();
</script>

<section
	class:danger={tone === 'error'}
	class:success={tone === 'success'}
	class:warning={tone === 'warning'}
	class:information={tone === 'info'}
	aria-live={tone === 'error' ? 'assertive' : 'polite'}
	role={tone === 'error' ? 'alert' : 'status'}
>
	<span class="icon" aria-hidden="true">
		{#if tone === 'error'}
			<CircleAlert size={22} />
		{:else if tone === 'success'}
			<CircleCheck size={22} />
		{:else if tone === 'warning'}
			<TriangleAlert size={22} />
		{:else}
			<Info size={22} />
		{/if}
	</span>
	<div class="copy">
		{#if title}<strong>{title}</strong>{/if}
		<p>{message}</p>
	</div>
	{#if actionLabel && onaction}
		<Button variant="secondary" onclick={onaction}>{actionLabel}</Button>
	{/if}
</section>

<style>
	section {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-3);
		border: 1px solid var(--status-info);
		border-inline-start-width: 4px;
		background: color-mix(in srgb, var(--status-info) 9%, var(--surface-raised));
		color: var(--text-primary);
		padding: var(--space-3);
	}

	.icon {
		display: grid;
		width: 2rem;
		height: 2rem;
		place-items: center;
		border: 1px solid currentColor;
		border-radius: 50%;
	}

	.copy {
		min-width: 0;
	}

	strong,
	p {
		margin: 0;
	}

	p {
		color: var(--text-secondary);
	}

	.danger {
		border-color: var(--status-danger);
		background: color-mix(in srgb, var(--status-danger) 9%, var(--surface-raised));
		color: var(--status-danger);
	}

	.success {
		border-color: var(--status-success);
		background: color-mix(in srgb, var(--status-success) 10%, var(--surface-raised));
		color: var(--status-success);
	}

	.warning {
		border-color: var(--status-warning);
		background: color-mix(in srgb, var(--status-warning) 13%, var(--surface-raised));
		color: var(--status-warning);
	}

	.information {
		color: var(--status-info);
	}

	.copy p {
		color: var(--text-primary);
	}

	@media (max-width: 31rem) {
		section {
			grid-template-columns: auto minmax(0, 1fr);
		}

		section :global(button) {
			grid-column: 1 / -1;
		}
	}
</style>
