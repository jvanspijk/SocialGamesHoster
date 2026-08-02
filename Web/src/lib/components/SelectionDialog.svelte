<script module lang="ts">
	export interface SelectionDialogEntry {
		id: string;
		label: string;
		accessibleLabel: string;
		supportingLabel?: string;
		leadingText?: string;
	}

	export interface SelectionDialogEmptyState {
		title: string;
		description: string;
	}
</script>

<script lang="ts">
	import Dialog from '$lib/components/Dialog.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import SelectableList from '$lib/components/SelectableList.svelte';

	let {
		open,
		title,
		description,
		entries,
		emptyState,
		onselect,
		close
	}: {
		open: boolean;
		title: string;
		description: string;
		entries: readonly SelectionDialogEntry[];
		emptyState: SelectionDialogEmptyState;
		onselect: (entryId: string) => void;
		close: () => void;
	} = $props();
</script>

<Dialog {open} {title} {description} {close}>
	{#if entries.length > 0}
		<SelectableList {entries} {onselect} />
	{:else}
		<EmptyState title={emptyState.title} description={emptyState.description} />
	{/if}
</Dialog>
