<script lang="ts">
	import { X } from '@lucide/svelte';
	import type { Snippet } from 'svelte';
	import IconButton from './IconButton.svelte';

	let {
		open,
		title,
		description = '',
		close,
		children,
		actions
	}: {
		open: boolean;
		title: string;
		description?: string;
		close: () => void;
		children: Snippet;
		actions?: Snippet;
	} = $props();

	let dialog: HTMLDialogElement;
	let heading: HTMLHeadingElement;
	let returnFocus: HTMLElement | null = null;

	$effect(() => {
		if (!dialog) return;
		if (open && !dialog.open) {
			returnFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
			dialog.showModal();
			queueMicrotask(() => heading?.focus());
		} else if (!open && dialog.open) {
			dialog.close();
		}
	});

	function finishClose() {
		if (open) close();
		queueMicrotask(() => returnFocus?.focus());
	}
</script>

<dialog
	bind:this={dialog}
	class="app-dialog"
	aria-labelledby="dialog-title"
	aria-describedby={description ? 'dialog-description' : undefined}
	onclose={finishClose}
	oncancel={(event) => {
		event.preventDefault();
		close();
	}}
>
	<header>
		<div>
			<h2 id="dialog-title" tabindex="-1" bind:this={heading}>{title}</h2>
			{#if description}<p id="dialog-description">{description}</p>{/if}
		</div>
		<IconButton label={`Close ${title}`} variant="ghost" onclick={close}>
			{#snippet icon()}<X size={21} />{/snippet}
		</IconButton>
	</header>
	<div class="dialog-body">
		{@render children()}
	</div>
	{#if actions}
		<footer>{@render actions()}</footer>
	{/if}
</dialog>

<style>
	.app-dialog {
		z-index: var(--layer-dialog);
		width: min(32rem, calc(100vw - 2rem));
		max-width: none;
		max-height: min(44rem, calc(100dvh - 2rem));
		border: var(--border-strong);
		background: var(--paper);
		box-shadow: var(--shadow);
		color: var(--ink);
		margin: auto;
		padding: 0;
	}

	.app-dialog::backdrop {
		background: rgb(17 10 6 / 72%);
		backdrop-filter: blur(2px);
	}

	header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding: var(--space-4);
	}

	h2,
	p {
		margin: 0;
	}

	header p {
		margin-block-start: var(--space-1);
		color: var(--ink-soft);
	}

	.dialog-body {
		overflow: auto;
		padding: var(--space-4);
	}

	footer {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: var(--space-2);
		border-block-start: var(--border-subtle);
		padding: var(--space-3) var(--space-4);
	}
</style>
