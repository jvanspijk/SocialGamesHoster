<script lang="ts">
	import type { AttentionItem } from '$lib/api/types';
	import Button from './Button.svelte';

	let {
		item,
		position,
		total,
		acknowledge,
		busy = false
	}: {
		item: AttentionItem;
		position: number;
		total: number;
		acknowledge: () => void;
		busy?: boolean;
	} = $props();
</script>

<article class="attention-card" aria-label={`Announcement ${position} of ${total}`}>
	<p class="queue">{position} of {total}</p>
	{#if item.kind === 'announcement'}
		<p class="sender">Announcement from {item.senderLabel}</p>
		<p class="content">{item.content}</p>
		<Button disabled={busy} onclick={acknowledge}>{busy ? 'Acknowledging…' : 'Acknowledge'}</Button>
	{:else}
		<p class="content">This event type is not available in this version.</p>
	{/if}
</article>

<style>
	.attention-card {
		display: grid;
		width: min(100%, 38rem);
		max-height: 100%;
		overflow: auto;
		border: 2px solid var(--crimson-dark);
		background: var(--paper-light);
		box-shadow: var(--shadow-small);
		padding: clamp(var(--space-4), 5vw, var(--space-7));
	}

	.queue,
	.sender {
		margin: 0;
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.09em;
		text-transform: uppercase;
	}

	.content {
		margin-block: var(--space-5);
		font-size: clamp(1.15rem, 4vw, 1.55rem);
		line-height: 1.4;
		white-space: pre-wrap;
	}

	:global(button) {
		justify-self: start;
	}
</style>
