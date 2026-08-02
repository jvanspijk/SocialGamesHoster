<script lang="ts">
	import { Hash, MessageCircle, Plus, Users } from '@lucide/svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import IconButton from '$lib/components/IconButton.svelte';
	import ItemRail from '$lib/components/ItemRail.svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import SearchField from '$lib/components/SearchField.svelte';
	import SelectableList, { type SelectableListEntry } from '$lib/components/SelectableList.svelte';

	let {
		entries,
		selectedId = '',
		search = $bindable(''),
		loading = false,
		onselect,
		onnewmessage
	}: {
		entries: readonly SelectableListEntry[];
		selectedId?: string;
		search?: string;
		loading?: boolean;
		onselect: (entryId: string) => void;
		onnewmessage?: () => void;
	} = $props();
</script>

<ItemRail eyebrow="Game chat" title="Conversations">
	{#snippet actions()}
		{#if onnewmessage}
			<IconButton label="New message" variant="ghost" onclick={onnewmessage}>
				{#snippet icon()}<Plus size={22} />{/snippet}
			</IconButton>
		{/if}
	{/snippet}
	{#snippet toolbar()}
		<SearchField
			label="Search conversations"
			placeholder="Search conversations"
			variant="inverse"
			bind:value={search}
		/>
	{/snippet}
	{#if loading}
		<div class="rail-status"><LoadingState label="Loading conversations…" /></div>
	{:else if entries.length === 0}
		<div class="rail-empty">
			<EmptyState
				title="No conversations"
				description={search
					? 'No conversations match your search.'
					: 'Conversations appear when chat is available.'}
				actionLabel={onnewmessage && !search ? 'New message' : undefined}
				onaction={onnewmessage && !search ? onnewmessage : undefined}
			>
				{#snippet icon()}<MessageCircle size={30} />{/snippet}
			</EmptyState>
		</div>
	{:else}
		<SelectableList {entries} {selectedId} variant="rich" {onselect}>
			{#snippet leading(entry)}
				{#if entry.leadingVariant === 'people'}
					<Users size={20} />
				{:else if entry.leadingVariant === 'hash'}
					<Hash size={20} />
				{:else}
					<span>{entry.leadingText}</span>
				{/if}
			{/snippet}
		</SelectableList>
	{/if}
</ItemRail>

<style>
	.rail-status,
	.rail-empty {
		color: var(--paper-muted);
		padding: var(--space-5);
		text-align: center;
	}

	.rail-empty :global(.empty-state) {
		color: var(--paper-light);
	}

	.rail-empty :global(.empty-state p),
	.rail-empty :global(.empty-state .icon) {
		color: var(--paper-muted);
	}
</style>
