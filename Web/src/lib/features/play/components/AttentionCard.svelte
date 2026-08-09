<script lang="ts">
	import type { AttentionItem } from '$lib/api/types';
	import Button from '$lib/components/Button.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import AnnouncementAttachments from './AnnouncementAttachments.svelte';
	import AttentionQueuePosition from './AttentionQueuePosition.svelte';

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
	<Panel variant="focal">
		<AttentionQueuePosition {position} {total} />
		{#if item.kind === 'announcement'}
			<p class="sender">Announcement from {item.senderLabel}</p>
			<p class="content">{item.content}</p>
			<AnnouncementAttachments image={item.image} audio={item.audio} />
			<div class="actions">
				<Button disabled={busy} onclick={acknowledge}>
					{busy ? 'Acknowledging…' : 'Acknowledge'}
				</Button>
			</div>
		{:else}
			<p class="content">This event type is not available in this version.</p>
		{/if}
	</Panel>
</article>

<style>
	.attention-card {
		width: min(100%, 38rem);
		max-height: 100%;
		overflow: auto;
	}

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

	.actions {
		justify-self: start;
	}
</style>
