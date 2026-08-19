<script lang="ts">
	import CollectionEditor from '$lib/features/rulesets/components/CollectionEditor.svelte';
	let entries = $state([
		{ id: 'one', label: 'One' },
		{ id: 'two', label: 'Two' }
	]);
	let selectedId = $state('one');
	function move(id: string, direction: -1 | 1) {
		const index = entries.findIndex((entry) => entry.id === id);
		const target = index + direction;
		if (target < 0 || target >= entries.length) return;
		[entries[index], entries[target]] = [entries[target], entries[index]];
	}
</script>

<CollectionEditor
	title="Items"
	description="Test items."
	{entries}
	{selectedId}
	onselect={(id) => (selectedId = id)}
	onadd={() => entries.push({ id: 'three', label: 'Three' })}
	onduplicate={(id) => {
		const source = entries.find((entry) => entry.id === id);
		if (source) entries.push({ id: `${id}-copy`, label: `${source.label} copy` });
	}}
	onmove={move}
	onremove={(id) =>
		entries.splice(
			entries.findIndex((entry) => entry.id === id),
			1
		)}
	usages={(id) => (id === 'one' ? [{ label: 'Role · Villager' }] : [])}
>
	{#snippet editor(id)}<label
			>Item name <input value={entries.find((entry) => entry.id === id)?.label} /></label
		>{/snippet}
</CollectionEditor>
