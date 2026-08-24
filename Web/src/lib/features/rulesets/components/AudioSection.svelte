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
		type AssetOption,
		type MediaActions
	} from './definition-editor';
	import MediaField from './MediaField.svelte';
	let {
		definition = $bindable(),
		assets,
		media,
		selectedItems,
		onnavigate
	}: {
		definition: RulesetDefinition;
		assets: AssetOption[];
		media: MediaActions;
		selectedItems: Record<string, string>;
		onnavigate: (section: EditorSection, itemId?: string) => void;
	} = $props();
	const audioAssets = () => assets.filter((asset) => asset.kind === 'audio');
	const entries = $derived(
		definition.audioCues.map((item) => ({
			id: item.id,
			label: item.name || 'Unnamed audio cue',
			supportingLabel:
				item.defaultAudience === 'all' ? 'All players' : item.defaultAudience.replace('_', ' ')
		}))
	);
	function add() {
		const item = {
			id: nextID(
				'cue',
				definition.audioCues.map((value) => value.id)
			),
			name: 'New audio cue',
			assetKey: audioAssets()[0]?.assetKey ?? '',
			defaultAudience: 'all' as const
		};
		definition.audioCues.push(item);
		selectedItems.audioCues = item.id;
	}
</script>

<CollectionEditor
	title="Audio cues"
	description="Optional named sounds for selected listeners."
	{entries}
	selectedId={selectedItems.audioCues ?? ''}
	onselect={(id) => (selectedItems.audioCues = id)}
	onadd={add}
	onduplicate={(id) => {
		const item = duplicateByID(definition.audioCues, id, 'cue');
		if (item) selectedItems.audioCues = item.id;
	}}
	onmove={(id, direction) => moveByID(definition.audioCues, id, direction)}
	onremove={(id) =>
		definition.audioCues.splice(
			definition.audioCues.findIndex((item) => item.id === id),
			1
		)}
	usages={(id) =>
		incomingReferences(definition, 'audioCue', id).map((usage) => ({
			label: usage.label,
			navigate: () => onnavigate(usage.section, usage.itemId)
		}))}
	emptyDescription="Add a named sound for phases or game-master playback."
>
	{#snippet editor(id)}{@const index = definition.audioCues.findIndex(
			(item) => item.id === id
		)}{@const cue = definition.audioCues[index]}{#if cue}<h3>{cue.name || 'Unnamed audio cue'}</h3>
			<div class="form-grid">
				<Field label="Name" name={`cue-name-${index}`} bind:value={cue.name} required /><SelectField
					label="Normal audience"
					name={`cue-audience-${index}`}
					bind:value={cue.defaultAudience}
					options={[
						{ value: 'all', label: 'All players' },
						{ value: 'game_masters', label: 'Game masters' },
						{ value: 'team', label: 'A selected team' },
						{ value: 'player', label: 'A selected player' }
					]}
				/>
			</div>
			<MediaField
				label="Audio file"
				kind="audio"
				name={`cue-audio-file-${index}`}
				bind:value={cue.assetKey}
				assets={audioAssets()}
				{media}
			/>
			{#if cue.defaultAudience === 'team' || cue.defaultAudience === 'player'}<p class="hint">
					Choose the target when playing this cue. It cannot start automatically with a phase.
				</p>{/if}{/if}{/snippet}
</CollectionEditor>

<style>
	h3 {
		margin: 0;
	}
</style>
