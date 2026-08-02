<script module lang="ts">
	export interface SelectableListEntry {
		id: string;
		label: string;
		accessibleLabel: string;
		supportingLabel?: string;
		leadingText?: string;
	}
</script>

<script lang="ts">
	let {
		entries,
		onselect
	}: {
		entries: readonly SelectableListEntry[];
		onselect: (entryId: string) => void;
	} = $props();
</script>

<div class="selectable-list">
	{#each entries as entry (entry.id)}
		<button type="button" aria-label={entry.accessibleLabel} onclick={() => onselect(entry.id)}>
			{#if entry.leadingText}
				<span class="leading-text" aria-hidden="true">{entry.leadingText}</span>
			{/if}
			<strong>{entry.label}</strong>
			{#if entry.supportingLabel}<small>{entry.supportingLabel}</small>{/if}
		</button>
	{/each}
</div>

<style>
	.selectable-list {
		display: grid;
	}

	.selectable-list button {
		display: flex;
		width: 100%;
		min-height: var(--target-size);
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

	.selectable-list button strong {
		min-width: 0;
		overflow-wrap: anywhere;
	}

	.leading-text {
		display: grid;
		width: 2.5rem;
		height: 2.5rem;
		flex: 0 0 2.5rem;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--ink);
		color: var(--gold-light);
	}

	.selectable-list small {
		margin-inline-start: auto;
		color: var(--ink-soft);
		text-align: end;
	}
</style>
