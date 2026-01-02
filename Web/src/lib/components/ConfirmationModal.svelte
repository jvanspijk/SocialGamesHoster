<script lang="ts">
	import { fade, scale } from 'svelte/transition';

	interface Props {
		isOpen: boolean;
		title?: string;
		message?: string;
		confirmText?: string;
		cancelText?: string;
		onConfirm: () => void;
		onCancel: () => void;
	}

	let {
		isOpen = $bindable(false),
		title = "Confirm Action",
		message = "Are you sure you want to proceed?",
		confirmText = "Yes",
		cancelText = "No",
		onConfirm,
		onCancel
	}: Props = $props();

	let dialogElement = $state<HTMLDialogElement | null>(null);

	$effect(() => {
		if (isOpen) {
			dialogElement?.showModal();
		} else {
			dialogElement?.close();
		}
	});

	function handleCancel(e: MouseEvent): void {
        if (e.target === e.currentTarget) {
            isOpen = false;
            onCancel();
        }
	}

	function handleConfirm(): void {
		isOpen = false;
		onConfirm();
	}

	function handleClose(e: Event): void {
		isOpen = false;
		onCancel();
	}
</script>

{#if isOpen}
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
{/if}

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
		background: rgba(0, 0, 0, 0.5);
	}

	.modal-inner {
		background: white;
		padding: 1.5rem;
		border-radius: 12px;
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
		width: 90vw;
		max-width: 400px;
		text-align: center;
	}

	h2 {
		margin-top: 0;
		font-size: 1.25rem;
		color: #111827;
	}

	p {
		color: #4b5563;
		margin-bottom: 2rem;
		line-height: 1.5;
	}

	.actions {
		display: flex;
		gap: 0.75rem;
	}

	button {
		flex: 1;
		padding: 0.75rem;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		font-size: 1rem;
		cursor: pointer;
		transition: filter 0.2s;
	}

	button:active {
		filter: brightness(0.9);
	}

	.btn-cancel {
		background: #f3f4f6;
		color: #374151;
	}

	.btn-confirm {
		background: #2563eb;
		color: white;
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