<script module lang="ts">
	import type { Snippet } from 'svelte';

	export interface MessageListItem {
		id: string;
		senderLabel: string;
		timeLabel: string;
		dayKey: string;
		dayLabel: string;
		content: string;
		isOwn?: boolean;
		deleted?: boolean;
		deletedLabel?: string;
		canRemove?: boolean;
		removeLabel?: string;
	}
</script>

<script lang="ts">
	import EmptyState from './EmptyState.svelte';
	import LoadingState from './LoadingState.svelte';
	import DayDivider from './DayDivider.svelte';
	import MessageItem from './MessageItem.svelte';
	import UnreadDivider from './UnreadDivider.svelte';

	let {
		messages,
		loading = false,
		hasEarlierMessages = false,
		firstUnreadId = '',
		loadingLabel = 'Loading messages…',
		emptyTitle = 'No messages yet',
		emptyDescription = '',
		emptyIcon,
		messageElement = $bindable<HTMLDivElement>(),
		onloadEarlier,
		onremove
	}: {
		messages: readonly MessageListItem[];
		loading?: boolean;
		hasEarlierMessages?: boolean;
		firstUnreadId?: string;
		loadingLabel?: string;
		emptyTitle?: string;
		emptyDescription: string;
		emptyIcon?: Snippet;
		messageElement?: HTMLDivElement;
		onloadEarlier?: () => void;
		onremove?: (messageId: string) => void;
	} = $props();
</script>

<div
	class="message-list"
	bind:this={messageElement}
	role="log"
	aria-label="Messages"
	aria-live="polite"
>
	{#if hasEarlierMessages && onloadEarlier}
		<button class="older" type="button" onclick={onloadEarlier}>Load earlier messages</button>
	{/if}
	{#if loading}
		<div class="message-status"><LoadingState label={loadingLabel} /></div>
	{:else if messages.length === 0}
		<div class="message-empty">
			<EmptyState title={emptyTitle} description={emptyDescription} icon={emptyIcon} />
		</div>
	{:else}
		{#each messages as message, index (message.id)}
			{#if index === 0 || message.dayKey !== messages[index - 1].dayKey}
				<DayDivider label={message.dayLabel} />
			{/if}
			{#if message.id === firstUnreadId}
				<UnreadDivider id={`unread-${message.id}`} />
			{/if}
			<MessageItem
				senderLabel={message.senderLabel}
				timeLabel={message.timeLabel}
				content={message.content}
				isOwn={message.isOwn}
				deleted={message.deleted}
				deletedLabel={message.deletedLabel}
				removeLabel={message.removeLabel}
				onremove={message.canRemove && onremove ? () => onremove(message.id) : undefined}
			/>
		{/each}
	{/if}
</div>

<style>
	.message-list {
		display: flex;
		min-height: 0;
		overflow-y: auto;
		flex-direction: column;
		gap: var(--space-2);
		padding: var(--space-4);
		overscroll-behavior: contain;
	}

	.older {
		align-self: center;
		min-height: var(--target-size);
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		font-size: 0.7rem;
		padding: 0 var(--space-2);
		text-decoration: underline;
	}

	.message-empty {
		align-self: center;
		justify-self: center;
		padding: var(--space-6);
		text-align: center;
	}

	.message-status {
		margin: auto;
	}

	@media (max-width: 47.99rem) {
		.message-list {
			padding: var(--space-3);
		}
	}
</style>
