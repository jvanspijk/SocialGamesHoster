<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ArrowDown, ArrowUp, Copy, Plus, Trash2 } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import ContentHeader from '$lib/components/ContentHeader.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import IconButton from '$lib/components/IconButton.svelte';
	import SearchField from '$lib/components/SearchField.svelte';
	import SelectableList from '$lib/components/SelectableList.svelte';
	import SplitView from '$lib/components/SplitView.svelte';
	import type { ValidationIssue } from '../editor-state';
	import InlineValidationMessages from './InlineValidationMessages.svelte';

	export type CollectionEntry = { id: string; label: string; supportingLabel?: string };
	export type ReferenceUsage = { label: string; navigate?: () => void };

	let {
		title,
		description,
		entries,
		selectedId,
		onselect,
		onadd,
		onduplicate,
		onmove,
		onremove,
		usages = () => [],
		addLabel = `Add ${title.toLocaleLowerCase()}`,
		emptyTitle = `No ${title.toLocaleLowerCase()}`,
		emptyDescription = `Add the first ${title.toLocaleLowerCase()} to get started.`,
		issues = [],
		validationPath,
		itemPath,
		editor
	}: {
		title: string;
		description: string;
		entries: CollectionEntry[];
		selectedId: string;
		onselect: (id: string) => void;
		onadd: () => void;
		onduplicate: (id: string) => void;
		onmove: (id: string, direction: -1 | 1) => void;
		onremove: (id: string) => void;
		usages?: (id: string) => ReferenceUsage[];
		addLabel?: string;
		emptyTitle?: string;
		emptyDescription?: string;
		issues?: ValidationIssue[];
		validationPath?: string;
		itemPath?: (id: string) => string;
		editor: Snippet<[id: string]>;
	} = $props();

	let search = $state('');
	let deleteOpen = $state(false);
	const filtered = $derived(
		entries.filter((entry) =>
			`${entry.label} ${entry.supportingLabel ?? ''}`
				.toLocaleLowerCase()
				.includes(search.trim().toLocaleLowerCase())
		)
	);
	const selectedIndex = $derived(entries.findIndex((entry) => entry.id === selectedId));
	const selected = $derived(entries[selectedIndex]);
	const incoming = $derived(selected ? usages(selected.id) : []);

	$effect(() => {
		if (entries.length && !entries.some((entry) => entry.id === selectedId))
			onselect(entries[0].id);
	});

	function keyboardNavigate(event: KeyboardEvent) {
		if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key) || filtered.length === 0)
			return;
		event.preventDefault();
		const current = filtered.findIndex((entry) => entry.id === selectedId);
		const index =
			event.key === 'Home'
				? 0
				: event.key === 'End'
					? filtered.length - 1
					: Math.max(
							0,
							Math.min(filtered.length - 1, current + (event.key === 'ArrowDown' ? 1 : -1))
						);
		onselect(filtered[index].id);
		queueMicrotask(() =>
			document
				.querySelector<HTMLElement>(`[data-entry-id="${CSS.escape(filtered[index].id)}"]`)
				?.focus()
		);
	}

	function confirmDelete() {
		if (!selected || incoming.length) return;
		onremove(selected.id);
		deleteOpen = false;
	}
</script>

<section class="collection-editor" aria-label={title}>
	<ContentHeader density="dense" {description}>
		{#snippet title()}<h2>{title}</h2>{/snippet}
		{#snippet actions()}{#if entries.length}<Button variant="secondary" onclick={onadd}
					><Plus size={16} /> {addLabel}</Button
				>{/if}{/snippet}
	</ContentHeader>
	{#if validationPath}<InlineValidationMessages {issues} path={validationPath} />{/if}
	{#if entries.length === 0}
		<EmptyState
			title={emptyTitle}
			description={emptyDescription}
			actionLabel={addLabel}
			onaction={onadd}
		/>
	{:else}
		<SplitView compact detailOpen={Boolean(selected)}>
			{#snippet rail()}
				<div class="rail-head">
					<SearchField
						label={`Search ${title}`}
						placeholder={`Search ${title.toLocaleLowerCase()}`}
						bind:value={search}
					/>
				</div>
				<div class="entry-list">
					<SelectableList
						entries={filtered.map((entry) => ({ ...entry, accessibleLabel: entry.label }))}
						{selectedId}
						{onselect}
						onkeydown={keyboardNavigate}
					/>
					{#if filtered.length === 0}<p class="no-results">No matching items.</p>{/if}
				</div>
			{/snippet}
			{#snippet detail()}
				{#if selected}
					{#if itemPath}<InlineValidationMessages {issues} path={itemPath(selected.id)} />{/if}
					<div class="detail-toolbar" aria-label={`${selected.label} actions`}>
						<IconButton
							label={`Move ${selected.label} up`}
							disabled={selectedIndex <= 0}
							onclick={() => onmove(selected.id, -1)}
							>{#snippet icon()}<ArrowUp size={18} />{/snippet}</IconButton
						>
						<IconButton
							label={`Move ${selected.label} down`}
							disabled={selectedIndex >= entries.length - 1}
							onclick={() => onmove(selected.id, 1)}
							>{#snippet icon()}<ArrowDown size={18} />{/snippet}</IconButton
						>
						<IconButton
							label={`Duplicate ${selected.label}`}
							onclick={() => onduplicate(selected.id)}
							>{#snippet icon()}<Copy size={18} />{/snippet}</IconButton
						>
						<IconButton
							label={`Delete ${selected.label}`}
							variant="danger"
							onclick={() => (deleteOpen = true)}
							>{#snippet icon()}<Trash2 size={18} />{/snippet}</IconButton
						>
					</div>
					<div class="detail-body">{@render editor(selected.id)}</div>
				{/if}
			{/snippet}
		</SplitView>
	{/if}
</section>

<Dialog
	open={deleteOpen}
	title={`Delete ${selected?.label ?? 'item'}?`}
	description={incoming.length
		? 'This item is still used elsewhere in the ruleset.'
		: 'This cannot be undone after you save.'}
	close={() => (deleteOpen = false)}
>
	{#if incoming.length}
		<p>Remove or change these uses first:</p>
		<ul>
			{#each incoming as usage (usage.label)}<li>
					{#if usage.navigate}<button
							class="usage"
							onclick={() => {
								deleteOpen = false;
								usage.navigate?.();
							}}>{usage.label}</button
						>{:else}{usage.label}{/if}
				</li>{/each}
		</ul>
	{:else}<p>The item will be removed from this working copy.</p>{/if}
	{#snippet actions()}<Button variant="ghost" onclick={() => (deleteOpen = false)}>Keep item</Button
		>{#if !incoming.length}<Button variant="danger" onclick={confirmDelete}>Delete</Button
			>{/if}{/snippet}
</Dialog>

<style>
	.collection-editor {
		display: grid;
		gap: var(--space-3);
	}
	.rail-head {
		padding-top: var(--space-3);
	}
	.detail-toolbar {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-1);
		border-block-end: var(--border-subtle);
		padding: var(--space-2);
	}
	.detail-body {
		display: grid;
		gap: var(--space-3);
		padding: var(--space-4);
	}
	.no-results {
		color: var(--ink-soft);
		padding: var(--space-3);
	}
	.usage {
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		text-decoration: underline;
	}
</style>
