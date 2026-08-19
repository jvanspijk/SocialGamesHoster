<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import CheckboxField from '$lib/components/CheckboxField.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import CollectionEditor from './CollectionEditor.svelte';
	import { duplicateByID, moveByID, nextID, type AssetOption } from './definition-editor';
	let {
		definition = $bindable(),
		assets,
		selectedItems
	}: {
		definition: RulesetDefinition;
		assets: AssetOption[];
		selectedItems: Record<string, string>;
	} = $props();
	const entries = $derived(
		definition.achievements.map((item) => ({
			id: item.id,
			label: item.name || 'Unnamed achievement',
			supportingLabel: `${item.points} points`
		}))
	);
	const imageOptions = () => [
		{ value: '', label: 'No image' },
		...assets
			.filter((asset) => asset.kind === 'image')
			.map((asset) => ({ value: asset.assetKey, label: asset.assetKey }))
	];
	function add() {
		const item = {
			id: nextID(
				'achievement',
				definition.achievements.map((value) => value.id)
			),
			name: 'New achievement',
			description: '',
			points: 0,
			hiddenUntilGameCompleted: false
		};
		definition.achievements.push(item);
		selectedItems.achievements = item.id;
	}
</script>

<CollectionEditor
	title="Achievements"
	description="Optional awards a game master can give after a game."
	{entries}
	selectedId={selectedItems.achievements ?? ''}
	onselect={(id) => (selectedItems.achievements = id)}
	onadd={add}
	onduplicate={(id) => {
		const item = duplicateByID(definition.achievements, id, 'achievement');
		if (item) selectedItems.achievements = item.id;
	}}
	onmove={(id, direction) => moveByID(definition.achievements, id, direction)}
	onremove={(id) =>
		definition.achievements.splice(
			definition.achievements.findIndex((item) => item.id === id),
			1
		)}
>
	{#snippet editor(id)}{@const index = definition.achievements.findIndex(
			(item) => item.id === id
		)}{@const item = definition.achievements[index]}{#if item}<h3>
				{item.name || 'Unnamed achievement'}
			</h3>
			<Field
				label="Name"
				name={`achievement-name-${index}`}
				bind:value={item.name}
				required
			/><Field
				label="Description"
				name={`achievement-description-${index}`}
				bind:value={item.description}
				multiline
			/>
			<div class="form-grid">
				<label
					><span>Achievement points</span><input
						name={`achievement-points-${index}`}
						type="number"
						min="0"
						max="10000"
						bind:value={item.points}
					/></label
				><CheckboxField
					label="Hide from players until the game ends"
					name={`achievement-hidden-${index}`}
					bind:checked={item.hiddenUntilGameCompleted}
				/>
			</div>
			<SelectField
				label="Badge image (optional)"
				name={`achievement-image-${index}`}
				bind:value={item.imageAssetKey}
				options={imageOptions()}
			/>{/if}{/snippet}
</CollectionEditor>

<style>
	h3 {
		margin: 0;
	}
</style>
