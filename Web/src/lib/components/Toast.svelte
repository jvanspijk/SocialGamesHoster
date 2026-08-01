<script lang="ts">
	import { CircleCheck, Info, TriangleAlert, X } from '@lucide/svelte';
	import IconButton from './IconButton.svelte';

	let {
		tone,
		message,
		actionLabel,
		onaction,
		ondismiss,
		onmouseenter,
		onmouseleave,
		onfocusin,
		onfocusout
	}: {
		tone: 'error' | 'success' | 'info';
		message: string;
		actionLabel?: string;
		onaction?: () => void;
		ondismiss: () => void;
		onmouseenter?: (event: MouseEvent) => void;
		onmouseleave?: (event: MouseEvent) => void;
		onfocusin?: (event: FocusEvent) => void;
		onfocusout?: (event: FocusEvent) => void;
	} = $props();
</script>

<article
	class:error={tone === 'error'}
	class:success={tone === 'success'}
	aria-live="polite"
	role="status"
	{onmouseenter}
	{onmouseleave}
	{onfocusin}
	{onfocusout}
>
	<span class="status-icon" aria-hidden="true">
		{#if tone === 'error'}
			<TriangleAlert size={20} />
		{:else if tone === 'success'}
			<CircleCheck size={20} />
		{:else}
			<Info size={20} />
		{/if}
	</span>
	<p>{message}</p>
	{#if actionLabel && onaction}
		<button class="action" type="button" onclick={onaction}>{actionLabel}</button>
	{/if}
	<IconButton label="Dismiss notification" variant="ghost" onclick={ondismiss}>
		{#snippet icon()}<X size={18} />{/snippet}
	</IconButton>
</article>

<style>
	article {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto auto;
		align-items: center;
		gap: var(--space-2);
		border: 1px solid var(--information);
		border-inline-start-width: 4px;
		background: var(--ink);
		box-shadow: var(--shadow);
		color: var(--paper-light);
		padding: var(--space-3);
	}

	.status-icon {
		display: grid;
		width: 1.7rem;
		height: 1.7rem;
		place-items: center;
		border: 1px solid currentColor;
		border-radius: 50%;
	}

	article.error {
		border-color: var(--danger);
	}

	article.success {
		border-color: var(--success);
	}

	p {
		margin: 0;
		color: var(--paper-light);
	}

	button {
		min-height: var(--target-size);
		border: 0;
		background: transparent;
		color: inherit;
		cursor: pointer;
	}

	.action {
		color: var(--gold-light);
		font-family: var(--font-display);
		font-weight: 700;
	}
</style>
