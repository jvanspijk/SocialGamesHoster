<script lang="ts">
	import { fade, scale } from 'svelte/transition';

	interface Props {
		title?: string;
		message?: string;
		confirmText?: string;
		cancelText?: string;
		onConfirm: () => void;
		onCancel: () => void;
	}

	let {
		title = "Confirm Action",
		message = "Are you sure you want to proceed?",
		confirmText = "Yes",
		cancelText = "No",
		onConfirm,
		onCancel
	}: Props = $props();

	let dialogElement = $state<HTMLDialogElement | null>(null);

	function handleCancel(e: MouseEvent): void {
        if (e.target === e.currentTarget) {
            onCancel();
        }
	}

	function handleConfirm(): void {
		onConfirm();
	}

	function handleClose(): void {
		onCancel();
	}
</script>


<dialog
	bind:this={dialogElement}
	onclose={handleClose}
	onclick={handleCancel}
	transition:fade={{ duration: 200 }}
>
	<div 
		class="modal-inner"
		transition:scale={{ duration: 200, start: 0.95 }}
	>
		<h2>{title}</h2>
		<p>{message}</p>

		<div class="actions">
			<button 
				type="button" 
				class="btn-cancel" 
				onclick={handleCancel}
			>
				{cancelText}
			</button>
			<button 
				type="button" 
				class="btn-confirm" 
				onclick={handleConfirm}
			>
				{confirmText}
			</button>
		</div>
	</div>
</dialog>


<style>
	dialog {
		border: none;
		padding: 0;
		background: transparent;
		max-width: 100vw;
		max-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	dialog::backdrop {
		background: rgba(34, 27, 22, 0.65);
	}

	.modal-inner {
		background: var(--color-surface-soft);
		color: var(--color-text);
		padding: 1.5rem;
		border-radius: 3px;
		border: 2px solid var(--color-border);
		box-shadow:
			0 8px 18px rgba(0, 0, 0, 0.45),
			inset 0 0 24px rgba(91, 74, 60, 0.15),
			2px 2px 0 rgba(91, 74, 60, 0.35);
		width: 90vw;
		max-width: 430px;
		text-align: center;
	}

	h2 {
		margin-top: 0;
		margin-bottom: 0.65rem;
		font-size: 1.35rem;
		font-family: var(--font-heading);
		color: var(--color-border);
		text-transform: uppercase;
		letter-spacing: 0.08em;
		border-bottom: 1px solid var(--color-border);
		padding-bottom: 0.4rem;
	}

	p {
		font-family: var(--font-body);
		color: var(--color-text);
		margin-bottom: 1.5rem;
		line-height: 1.5;
		font-size: 1.06rem;
	}

	.actions {
		display: flex;
		gap: 0.75rem;
	}

	button {
		font-family: var(--font-heading);
		flex: 1;
		padding: 0.65rem 0.85rem;
		border: 1px solid var(--color-border);
		border-radius: 3px;
		font-weight: 600;
		font-size: 0.9rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		cursor: pointer;
		transition: transform 0.12s ease, box-shadow 0.12s ease, background-color 0.12s ease;
		box-shadow: 1px 1px 0 rgba(91, 74, 60, 0.8);
	}

	button:hover {
		transform: translateY(-1px);
		box-shadow: 2px 2px 0 rgba(91, 74, 60, 0.8);
	}

	button:active {
		transform: translateY(1px);
		box-shadow: inset 1px 1px 2px rgba(0, 0, 0, 0.18);
	}

	button:focus-visible {
		outline: 3px solid var(--color-focus);
		outline-offset: 2px;
	}

	.btn-cancel {
		background: var(--color-surface);
		color: var(--color-border);
	}

	.btn-cancel:hover {
		background: var(--color-surface-alt);
	}

	.btn-confirm {
		background: var(--color-accent-strong);
		border-color: var(--color-accent);
		color: var(--color-on-accent);
	}

	.btn-confirm:hover {
		background: var(--color-accent);
	}

	@media (max-width: 480px) {
		.modal-inner {
			padding: 1.25rem;
		}
		
		.actions {
			flex-direction: column-reverse;
		}
	}
</style>
