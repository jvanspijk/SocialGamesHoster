<script lang="ts">
	import { Send } from '@lucide/svelte';
	import Button from './Button.svelte';

	let {
		value = $bindable(''),
		inputId = 'chat-message',
		restrictionId = 'message-restriction',
		restrictionLabel = '',
		placeholder = 'Write a message',
		maxLength = 1000,
		disabled = false,
		sending = false,
		onsubmit
	}: {
		value?: string;
		inputId?: string;
		restrictionId?: string;
		restrictionLabel?: string;
		placeholder?: string;
		maxLength?: number;
		disabled?: boolean;
		sending?: boolean;
		onsubmit?: () => void;
	} = $props();

	function submit(event: SubmitEvent) {
		event.preventDefault();
		onsubmit?.();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			(event.currentTarget as HTMLTextAreaElement).form?.requestSubmit();
		}
	}
</script>

<form onsubmit={submit}>
	<label class="sr-only" for={inputId}>Message</label>
	{#if restrictionLabel}<p id={restrictionId} class="message-restriction">
			{restrictionLabel}
		</p>{/if}
	<textarea
		id={inputId}
		bind:value
		aria-describedby={restrictionLabel ? restrictionId : undefined}
		maxlength={maxLength}
		rows="1"
		{placeholder}
		{disabled}
		onkeydown={handleKeydown}
	></textarea>
	<Button type="submit" loading={sending} disabled={disabled || !value.trim()}>
		<Send size={18} /> Send
	</Button>
</form>

<style>
	form {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		gap: var(--space-2);
		border-block-start: var(--border-subtle);
		background: var(--paper-light);
		padding: var(--space-3);
		padding-block-end: max(var(--space-3), env(safe-area-inset-bottom));
	}

	textarea {
		min-width: 0;
		min-height: var(--target-size);
		max-height: 8rem;
		resize: vertical;
		border: var(--border-subtle);
		background: white;
		color: var(--ink);
		padding: var(--space-2);
	}

	.message-restriction {
		align-self: center;
		margin: 0;
		color: var(--ink-soft);
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}

	@media (max-width: 47.99rem) {
		form {
			position: sticky;
			inset-block-end: 0;
			padding: var(--space-2);
			padding-block-end: max(var(--space-2), env(safe-area-inset-bottom));
		}

		form :global(button) {
			width: var(--target-size);
			overflow: hidden;
			font-size: 0;
			padding: 0;
		}
	}
</style>
