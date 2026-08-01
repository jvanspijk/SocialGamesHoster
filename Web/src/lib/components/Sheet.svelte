<script lang="ts">
	import { X } from '@lucide/svelte';
	import type { Snippet } from 'svelte';
	import IconButton from './IconButton.svelte';

	let {
		open,
		title,
		close,
		children
	}: {
		open: boolean;
		title: string;
		close: () => void;
		children: Snippet;
	} = $props();

	let dialog: HTMLDialogElement;
	let heading: HTMLHeadingElement;
	let returnFocus: HTMLElement | null = null;
	let shown = false;

	$effect(() => {
		if (!dialog) return;
		if (open && !dialog.open) {
			returnFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
			dialog.showModal();
			shown = true;
			queueMicrotask(() => heading?.focus());
		} else if (!open && shown) {
			if (dialog.open) dialog.close();
			restoreTriggerFocus();
		}
	});

	function restoreTriggerFocus() {
		if (shown) {
			shown = false;
			queueMicrotask(() => returnFocus?.focus());
		}
	}

	function handleClose() {
		restoreTriggerFocus();
		if (open) close();
	}
</script>

<dialog bind:this={dialog} class="app-sheet" aria-label={title} onclose={handleClose}>
	<header>
		<h2 tabindex="-1" bind:this={heading}>{title}</h2>
		<IconButton label={`Close ${title}`} variant="ghost" onclick={close}>
			{#snippet icon()}<X size={22} />{/snippet}
		</IconButton>
	</header>
	<div class="sheet-content">
		{@render children()}
	</div>
</dialog>

<style>
	.app-sheet {
		z-index: var(--layer-sheet);
		width: min(32rem, calc(100% - 1rem));
		max-width: none;
		height: min(48rem, calc(100dvh - 1rem));
		max-height: none;
		border: var(--border-strong);
		background: var(--paper);
		color: var(--ink);
		box-shadow: var(--shadow);
		margin: auto 0.5rem auto auto;
		padding: 0;
	}

	.app-sheet::backdrop {
		background: rgb(18 12 8 / 68%);
	}

	header {
		position: sticky;
		z-index: 1;
		inset-block-start: 0;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		background: var(--paper);
		padding: max(var(--space-3), env(safe-area-inset-top))
			max(var(--space-3), env(safe-area-inset-right)) var(--space-3) var(--space-4);
	}

	h2 {
		margin: 0;
		font-size: 1.15rem;
	}

	.sheet-content {
		min-width: 0;
		height: calc(100% - 4.25rem);
		overflow: auto;
		overscroll-behavior: contain;
		padding: var(--space-4);
		padding-block-end: max(var(--space-5), env(safe-area-inset-bottom));
	}

	@media (max-width: 47.99rem) {
		.app-sheet {
			inset: 0;
			width: 100%;
			height: 100dvh;
			max-height: 100dvh;
			border: 0;
			margin: 0;
		}
	}
</style>
