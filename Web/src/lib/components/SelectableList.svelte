<script module lang="ts">
	import type { Snippet } from 'svelte';

	export interface SelectableListEntry {
		id: string;
		label: string;
		accessibleLabel: string;
		supportingLabel?: string;
		leadingText?: string;
		description?: string;
		metaLabel?: string;
		unread?: boolean;
		leadingVariant?: string;
	}
</script>

<script lang="ts">
	let {
		entries,
		onselect,
		onkeydown,
		selectedId = '',
		variant = 'compact',
		leading
	}: {
		entries: readonly SelectableListEntry[];
		onselect: (entryId: string) => void;
		onkeydown?: (event: KeyboardEvent) => void;
		selectedId?: string;
		variant?: 'compact' | 'rich';
		leading?: Snippet<[entry: SelectableListEntry]>;
	} = $props();
</script>

<div class="selectable-list" class:rich={variant === 'rich'}>
	{#each entries as entry (entry.id)}
		<button
			data-entry-id={entry.id}
			{onkeydown}
			type="button"
			class:selected={selectedId === entry.id}
			class:unread={entry.unread}
			aria-label={entry.accessibleLabel}
			aria-current={selectedId === entry.id ? 'page' : undefined}
			onclick={() => onselect(entry.id)}
		>
			{#if variant === 'rich'}
				<span class="leading" aria-hidden="true">
					{#if leading}
						{@render leading(entry)}
					{:else if entry.leadingText}
						{entry.leadingText}
					{/if}
				</span>
				<span class="entry-copy">
					<span class="entry-title">
						<strong>{entry.label}</strong>
						{#if entry.metaLabel}<time>{entry.metaLabel}</time>{/if}
					</span>
					{#if entry.description}<span class="description">{entry.description}</span>{/if}
					{#if entry.supportingLabel}<small>{entry.supportingLabel}</small>{/if}
				</span>
				{#if entry.unread}<i aria-label="New messages"></i>{/if}
			{:else}
				{#if entry.leadingText}
					<span class="leading-text" aria-hidden="true">{entry.leadingText}</span>
				{/if}
				<strong>{entry.label}</strong>
				{#if entry.supportingLabel}<small>{entry.supportingLabel}</small>{/if}
			{/if}
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

	.rich {
		margin: 0;
	}

	.rich button {
		position: relative;
		display: grid;
		width: 100%;
		min-height: 5.25rem;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: center;
		gap: var(--space-3);
		border: 0;
		border-block-end: 1px solid rgb(223 189 101 / 18%);
		background: transparent;
		color: var(--paper-muted);
		cursor: pointer;
		padding: var(--space-3);
		text-align: start;
	}

	.rich button:hover,
	.rich button.selected {
		background: color-mix(in srgb, var(--crimson-light) 13%, transparent);
	}

	.rich button.selected {
		box-shadow: inset 3px 0 var(--gold-light);
	}

	.rich .leading {
		display: grid;
		width: 2.8rem;
		height: 2.8rem;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--crimson-dark);
		color: var(--gold-light);
		font-family: var(--font-display);
		font-weight: 700;
	}

	.rich .entry-copy {
		min-width: 0;
	}

	.rich .entry-title {
		display: flex;
		justify-content: space-between;
		gap: var(--space-2);
	}

	.rich .entry-title strong {
		overflow: hidden;
		color: var(--paper-light);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.rich .entry-title time,
	.rich .entry-copy small {
		color: var(--paper-muted);
		font-size: 0.7rem;
	}

	.rich .description,
	.rich .entry-copy small {
		display: block;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.rich .description {
		margin-block: 0.1rem;
		font-size: 0.82rem;
	}

	.rich button.unread .entry-title strong,
	.rich button.unread .description {
		color: var(--gold-light);
		font-weight: 700;
	}

	.rich button > i {
		position: absolute;
		inset-inline-end: var(--space-3);
		inset-block-end: var(--space-3);
		width: 0.55rem;
		height: 0.55rem;
		border-radius: 50%;
		background: var(--crimson-light);
	}
</style>
