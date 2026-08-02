<script lang="ts">
	import { Trash2 } from '@lucide/svelte';

	let {
		senderLabel,
		timeLabel,
		content,
		isOwn = false,
		deleted = false,
		deletedLabel = 'Message removed',
		removeLabel,
		onremove
	}: {
		senderLabel: string;
		timeLabel: string;
		content: string;
		isOwn?: boolean;
		deleted?: boolean;
		deletedLabel?: string;
		removeLabel?: string;
		onremove?: () => void;
	} = $props();
</script>

<article class:own={isOwn} class:deleted>
	<div class="message-meta">
		<strong>{senderLabel}</strong>
		<time>{timeLabel}</time>
	</div>
	<p>{deleted ? deletedLabel : content}</p>
	{#if !deleted && onremove}
		<button type="button" aria-label={removeLabel ?? 'Remove message'} onclick={() => onremove?.()}>
			<Trash2 size={14} /> Remove
		</button>
	{/if}
</article>

<style>
	article {
		width: fit-content;
		max-width: min(82%, 42rem);
		align-self: flex-start;
		border: 1px solid #bda574;
		border-radius: 0 0.65rem 0.65rem 0.65rem;
		background: var(--paper-light);
		box-shadow: var(--shadow-small);
		padding: var(--space-2) var(--space-3);
	}

	article.own {
		align-self: flex-end;
		border-color: #9c7740;
		border-radius: 0.65rem 0 0.65rem 0.65rem;
		background: #ead3a7;
	}

	article.deleted {
		box-shadow: none;
		opacity: 0.68;
		font-style: italic;
	}

	.message-meta {
		display: flex;
		justify-content: space-between;
		gap: var(--space-4);
		font-size: 0.74rem;
	}

	.message-meta strong {
		color: var(--crimson-dark);
	}

	.message-meta time {
		color: var(--ink-faint);
	}

	p {
		margin: 0.15rem 0 0;
		white-space: pre-wrap;
	}

	button {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: 0.2rem;
		border: 0;
		background: transparent;
		color: var(--danger);
		cursor: pointer;
		font-size: 0.7rem;
		padding: 0;
	}

	@media (max-width: 47.99rem) {
		article {
			max-width: 88%;
		}
	}
</style>
