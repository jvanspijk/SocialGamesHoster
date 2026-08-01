<script lang="ts">
	import { X } from '@lucide/svelte';
	import type { Snippet } from 'svelte';
	import IconButton from './IconButton.svelte';

	type Presentation = 'dialog' | 'sheet';

	let {
		open,
		title,
		description = '',
		close,
		presentation,
		children,
		actions
	}: {
		open: boolean;
		title: string;
		description?: string;
		close: () => void;
		presentation: Presentation;
		children: Snippet;
		actions?: Snippet;
	} = $props();

	const id = $props.id();
	const titleId = `${id}-title`;
	const descriptionId = `${id}-description`;

	let dialog: HTMLDialogElement;
	let closing = $state(false);
	let closeRequested = false;
	let closeTimer: ReturnType<typeof setTimeout> | undefined;

	$effect(() => {
		if (!dialog) return;

		if (open) {
			if (closing) cancelClose();
			else if (!dialog.open) dialog.showModal();
		} else if (dialog.open && !closing) {
			beginClose();
		}
	});

	function requestClose() {
		if (!dialog.open || closing) return;

		closeRequested = true;
		close();
		if (dialog.open) beginClose();
	}

	function handleCancel(event: Event) {
		event.preventDefault();
		requestClose();
	}

	function beginClose() {
		if (closing || !dialog.open) return;

		closing = true;
		if (window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false) {
			queueMicrotask(finishClose);
			return;
		}

		const duration = Number.parseFloat(getComputedStyle(dialog).getPropertyValue('--speed-fast'));
		closeTimer = setTimeout(finishClose, Number.isFinite(duration) ? duration + 50 : 150);
	}

	function cancelClose() {
		closing = false;
		if (closeTimer) clearTimeout(closeTimer);
		closeTimer = undefined;
	}

	function finishClose() {
		if (!closing || !dialog.open) return;

		cancelClose();
		dialog.close();
	}

	function handleTransitionEnd(event: TransitionEvent) {
		if (event.target !== event.currentTarget || event.propertyName !== 'opacity') return;
		finishClose();
	}

	function handleClose() {
		cancelClose();
		if (open && !closeRequested) close();
		closeRequested = false;
	}
</script>

<dialog
	bind:this={dialog}
	class:closing
	class:dialog-presentation={presentation === 'dialog'}
	class:sheet-presentation={presentation === 'sheet'}
	class="modal-overlay"
	data-presentation={presentation}
	aria-labelledby={titleId}
	aria-describedby={description ? descriptionId : undefined}
	onclose={handleClose}
	oncancel={handleCancel}
	ontransitionend={handleTransitionEnd}
>
	<header>
		<div>
			<h2 id={titleId}>{title}</h2>
			{#if description}<p id={descriptionId}>{description}</p>{/if}
		</div>
		<IconButton autofocus label={`Close ${title}`} variant="ghost" onclick={requestClose}>
			{#snippet icon()}<X size={presentation === 'dialog' ? 21 : 22} />{/snippet}
		</IconButton>
	</header>
	<div class="overlay-body">
		{@render children()}
	</div>
	{#if actions}
		<footer>{@render actions()}</footer>
	{/if}
</dialog>

<style>
	.modal-overlay {
		display: flex;
		flex-direction: column;
		max-width: none;
		border: var(--border-strong);
		background: var(--paper);
		box-shadow: var(--shadow);
		color: var(--ink);
		overflow: hidden;
		padding: 0;
		opacity: 1;
		transform: translateY(0);
		transition:
			opacity var(--speed-fast) ease-in,
			transform var(--speed-fast) ease-in;
	}

	.modal-overlay:not([open]) {
		display: none;
	}

	.modal-overlay::backdrop {
		background: rgb(17 10 6 / 72%);
		backdrop-filter: blur(2px);
		opacity: 1;
		transition: opacity var(--speed-fast) ease-in;
	}

	.modal-overlay.closing {
		opacity: 0;
		transform: translateY(var(--space-2));
	}

	.modal-overlay.closing::backdrop {
		opacity: 0;
	}

	.dialog-presentation {
		z-index: var(--layer-dialog);
		width: min(32rem, calc(100vw - 2rem));
		min-height: 0;
		max-height: min(44rem, calc(100dvh - 2rem));
		margin: auto;
	}

	.sheet-presentation {
		z-index: var(--layer-sheet);
		width: min(32rem, calc(100vw - 1rem));
		height: min(48rem, calc(100dvh - 1rem));
		max-height: none;
		margin: auto 0.5rem auto auto;
		transform: translateX(0);
	}

	.sheet-presentation.closing {
		transform: translateX(var(--space-3));
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

	.overlay-body {
		flex: 1 1 auto;
		min-height: 0;
		min-width: 0;
		overflow: auto;
		overscroll-behavior: contain;
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

	.sheet-presentation header {
		position: sticky;
		z-index: 1;
		inset-block-start: 0;
		align-items: center;
		background: var(--paper);
		padding: max(var(--space-3), env(safe-area-inset-top))
			max(var(--space-3), env(safe-area-inset-right)) var(--space-3) var(--space-4);
	}

	.sheet-presentation h2 {
		font-size: 1.15rem;
	}

	.sheet-presentation .overlay-body {
		padding-block-end: max(var(--space-5), env(safe-area-inset-bottom));
	}

	@media (max-width: 47.99rem) {
		.sheet-presentation {
			inset: 0;
			width: 100%;
			height: 100dvh;
			max-height: 100dvh;
			border: 0;
			margin: 0;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.modal-overlay,
		.modal-overlay::backdrop {
			transition-duration: 1ms;
		}
	}
</style>
