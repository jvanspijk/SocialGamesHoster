<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import Field from '$lib/components/Field.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import type { EditorSection } from '../editor-state';
	import CollectionEditor from './CollectionEditor.svelte';
	import {
		duplicateByID,
		incomingReferences,
		moveByID,
		nextID,
		type AssetOption
	} from './definition-editor';

	let {
		definition = $bindable(),
		assets,
		selectedItems,
		onnavigate
	}: {
		definition: RulesetDefinition;
		assets: AssetOption[];
		selectedItems: Record<string, string>;
		onnavigate: (section: EditorSection, itemId?: string) => void;
	} = $props();
	const imageOptions = () => [
		{ value: '', label: 'No image' },
		...assets
			.filter((asset) => asset.kind === 'image')
			.map((asset) => ({ value: asset.assetKey, label: asset.assetKey }))
	];
	const teamEntries = $derived(
		definition.teams.map((item) => ({
			id: item.id,
			label: item.name || 'Unnamed team',
			supportingLabel: `${definition.roles.filter((role) => role.teamId === item.id).length} roles`
		}))
	);
	const categoryEntries = $derived(
		definition.categories.map((item) => ({ id: item.id, label: item.name || 'Unnamed category' }))
	);
	function addTeam() {
		const item = {
			id: nextID(
				'team',
				definition.teams.map((value) => value.id)
			),
			name: 'New team',
			description: ''
		};
		definition.teams.push(item);
		selectedItems.teams = item.id;
	}
	function addCategory() {
		const item = {
			id: nextID(
				'category',
				definition.categories.map((value) => value.id)
			),
			name: 'New category',
			description: ''
		};
		definition.categories.push(item);
		selectedItems.categories = item.id;
	}
	function navigation(kind: 'team' | 'category', id: string) {
		return incomingReferences(definition, kind, id).map((usage) => ({
			label: usage.label,
			navigate: () => onnavigate(usage.section, usage.itemId)
		}));
	}
</script>

<CollectionEditor
	title="Teams"
	description="The main sides or factions in the game."
	entries={teamEntries}
	selectedId={selectedItems.teams ?? ''}
	onselect={(id) => (selectedItems.teams = id)}
	onadd={addTeam}
	onduplicate={(id) => {
		const item = duplicateByID(definition.teams, id, 'team');
		if (item) selectedItems.teams = item.id;
	}}
	onmove={(id, direction) => moveByID(definition.teams, id, direction)}
	onremove={(id) =>
		definition.teams.splice(
			definition.teams.findIndex((item) => item.id === id),
			1
		)}
	usages={(id) => navigation('team', id)}
	emptyDescription="Add a team before creating roles."
>
	{#snippet editor(id)}{@const index = definition.teams.findIndex(
			(item) => item.id === id
		)}{@const team = definition.teams[index]}
		{#if team}<h3>{team.name || 'Unnamed team'}</h3>
			<Field label="Name" name={`team-name-${index}`} bind:value={team.name} required /><Field
				label="Description"
				name={`team-description-${index}`}
				bind:value={team.description}
				multiline
			/><SelectField
				label="Team image (optional)"
				name={`team-image-${index}`}
				bind:value={team.imageAssetKey}
				options={imageOptions()}
			/>{/if}
	{/snippet}
</CollectionEditor>

<CollectionEditor
	title="Categories"
	description="Optional labels such as Investigative or Support."
	entries={categoryEntries}
	selectedId={selectedItems.categories ?? ''}
	onselect={(id) => (selectedItems.categories = id)}
	onadd={addCategory}
	onduplicate={(id) => {
		const item = duplicateByID(definition.categories, id, 'category');
		if (item) selectedItems.categories = item.id;
	}}
	onmove={(id, direction) => moveByID(definition.categories, id, direction)}
	onremove={(id) =>
		definition.categories.splice(
			definition.categories.findIndex((item) => item.id === id),
			1
		)}
	usages={(id) => navigation('category', id)}
	emptyDescription="Categories are optional. Add one to group similar roles."
>
	{#snippet editor(id)}{@const index = definition.categories.findIndex(
			(item) => item.id === id
		)}{@const category = definition.categories[index]}
		{#if category}<h3>{category.name || 'Unnamed category'}</h3>
			<Field
				label="Name"
				name={`category-name-${index}`}
				bind:value={category.name}
				required
			/><Field
				label="Description"
				name={`category-description-${index}`}
				bind:value={category.description}
				multiline
			/>{/if}
	{/snippet}
</CollectionEditor>

<style>
	h3 {
		margin: 0;
	}
</style>
