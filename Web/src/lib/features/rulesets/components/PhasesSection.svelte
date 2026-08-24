<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import CheckboxField from '$lib/components/CheckboxField.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import type { EditorSection } from '../editor-state';
	import CollectionEditor from './CollectionEditor.svelte';
	import {
		duplicateByID,
		incomingReferences,
		moveByID,
		nextID,
		type MediaActions
	} from './definition-editor';
	let {
		definition = $bindable(),
		media,
		selectedItems,
		onnavigate
	}: {
		definition: RulesetDefinition;
		media: MediaActions;
		selectedItems: Record<string, string>;
		onnavigate: (section: EditorSection, itemId?: string) => void;
	} = $props();
	let soundName = $state('New sound');
	let soundAlternative = $state('');
	let addingSound = $state(false);
	let soundError = $state('');

	async function addSound(phaseIndex: number, event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;
		addingSound = true;
		soundError = '';
		try {
			const asset = await media.upload(
				file,
				'audio',
				soundName.trim() || file.name,
				soundAlternative.trim()
			);
			const cue = {
				id: nextID(
					'cue',
					definition.audioCues.map((item) => item.id)
				),
				name: soundName.trim() || file.name,
				assetKey: asset.assetKey,
				defaultAudience: 'all' as const
			};
			definition.audioCues.push(cue);
			definition.phases[phaseIndex].audioCueId = cue.id;
		} catch (caught) {
			soundError = caught instanceof Error ? caught.message : 'The sound could not be added.';
		} finally {
			addingSound = false;
			input.value = '';
		}
	}
	const entries = $derived(
		definition.phases.map((phase) => ({
			id: phase.id,
			label: phase.name || 'Unnamed phase',
			supportingLabel: `Order ${phase.order}`
		}))
	);
	function add() {
		const item = {
			id: nextID(
				'phase',
				definition.phases.map((value) => value.id)
			),
			name: 'New phase',
			description: '',
			order: Math.max(0, ...definition.phases.map((value) => value.order)) + 1,
			startsRound: false,
			suggestedDurationSeconds: 0
		};
		definition.phases.push(item);
		selectedItems.phases = item.id;
	}
	function movePhase(id: string, direction: -1 | 1) {
		moveByID(definition.phases, id, direction);
		definition.phases.forEach((phase, index) => (phase.order = index + 1));
	}
</script>

<CollectionEditor
	title="Phases"
	description="Optional ordered steps a game master can advance through."
	{entries}
	selectedId={selectedItems.phases ?? ''}
	onselect={(id) => (selectedItems.phases = id)}
	onadd={add}
	onduplicate={(id) => {
		const item = duplicateByID(definition.phases, id, 'phase');
		if (item) {
			item.order = Math.max(...definition.phases.map((value) => value.order)) + 1;
			selectedItems.phases = item.id;
		}
	}}
	onmove={movePhase}
	onremove={(id) =>
		definition.phases.splice(
			definition.phases.findIndex((item) => item.id === id),
			1
		)}
	usages={(id) =>
		incomingReferences(definition, 'phase', id).map((usage) => ({
			label: usage.label,
			navigate: () => onnavigate(usage.section, usage.itemId)
		}))}
	emptyDescription="Game flow is optional. Add a phase if the game follows an ordered sequence."
>
	{#snippet editor(id)}{@const index = definition.phases.findIndex(
			(item) => item.id === id
		)}{@const phase = definition.phases[index]}{#if phase}<h3>{phase.name || 'Unnamed phase'}</h3>
			<div class="form-grid thirds">
				<Field label="Name" name={`phase-name-${index}`} bind:value={phase.name} required /><label
					><span>Order</span><input type="number" min="1" bind:value={phase.order} /></label
				>
			</div>
			<Field
				label="Instructions"
				name={`phase-description-${index}`}
				bind:value={phase.description}
				multiline
			/>
			<div class="form-grid thirds">
				<label
					><span>Suggested seconds</span><input
						name={`phase-seconds-${index}`}
						type="number"
						min="0"
						bind:value={phase.suggestedDurationSeconds}
					/></label
				><CheckboxField
					label="Starts a new round"
					name={`phase-starts-round-${index}`}
					bind:checked={phase.startsRound}
				/><SelectField
					label="Sound when phase starts"
					name={`phase-audio-${index}`}
					bind:value={phase.audioCueId}
					options={[
						{ value: '', label: 'No sound' },
						...definition.audioCues.map((cue) => ({ value: cue.id, label: cue.name }))
					]}
				/>
			</div>
			<section class="add-sound" aria-labelledby={`add-sound-${index}`}>
				<h4 id={`add-sound-${index}`}>Add sound</h4>
				<p>Upload audio and create a cue for this phase without leaving Game flow.</p>
				<div class="form-grid">
					<Field label="Sound name" name={`phase-sound-name-${index}`} bind:value={soundName} />
					<Field
						label="Audio alternative"
						name={`phase-sound-alternative-${index}`}
						bind:value={soundAlternative}
					/>
				</div>
				<label class="file-label"
					><span>{addingSound ? 'Adding sound…' : 'Choose audio file'}</span><input
						type="file"
						accept="audio/mpeg,audio/mp4,audio/ogg,audio/wav"
						disabled={addingSound}
						onchange={(event) => addSound(index, event)}
					/></label
				>
				{#if soundError}<p class="error" role="alert">{soundError}</p>{/if}
			</section>
		{/if}{/snippet}
</CollectionEditor>

<style>
	h3 {
		margin: 0;
	}
	.add-sound {
		display: grid;
		gap: var(--space-2);
		border-top: var(--border-subtle);
		padding-top: var(--space-3);
	}
	.add-sound h4,
	.add-sound p {
		margin: 0;
	}
	.file-label input {
		padding: var(--space-2);
	}
	.error {
		color: var(--danger);
	}
</style>
