<script module lang="ts">
	export interface DirectMessageRecipient {
		id: string;
		displayLabel: string;
		supportingLabel?: string;
		avatarText: string;
	}
</script>

<script lang="ts">
	import Dialog from '$lib/components/Dialog.svelte';

	let {
		open,
		description,
		entries,
		onchoose,
		close
	}: {
		open: boolean;
		description: string;
		entries: DirectMessageRecipient[];
		onchoose: (recipientId: string) => void;
		close: () => void;
	} = $props();
</script>

<Dialog {open} title="New message" {description} {close}>
	{#if entries.length > 0}
		<div class="player-list">
			{#each entries as entry (entry.id)}
				<button type="button" onclick={() => onchoose(entry.id)}>
					<span>{entry.avatarText}</span>
					<strong>{entry.displayLabel}</strong>
					{#if entry.supportingLabel}<small>{entry.supportingLabel}</small>{/if}
				</button>
			{/each}
		</div>
	{:else}
		<p class="empty-state">No players are available for a direct conversation.</p>
	{/if}
</Dialog>

<style>
	.player-list {
		display: grid;
	}

	.player-list button {
		display: grid;
		width: 100%;
		min-height: var(--target-size);
		grid-template-columns: auto 1fr auto;
		align-items: center;
		gap: var(--space-3);
		border: 0;
		border-block-end: var(--border-subtle);
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		padding: var(--space-2) 0;
		text-align: start;
	}

	.player-list button > span {
		display: grid;
		width: 2.5rem;
		height: 2.5rem;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--ink);
		color: var(--gold-light);
	}

	.player-list small,
	.empty-state {
		color: var(--ink-soft);
	}

	.empty-state {
		margin: 0;
	}
</style>
