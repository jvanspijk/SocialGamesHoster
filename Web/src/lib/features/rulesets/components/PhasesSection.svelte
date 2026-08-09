<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import Button from '$lib/components/Button.svelte';
	import CheckboxField from '$lib/components/CheckboxField.svelte';
	import ContentHeader from '$lib/components/ContentHeader.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import { nextID, removeAt } from './definition-editor';

	let { definition = $bindable() }: { definition: RulesetDefinition } = $props();

	function addPhase() {
		const order = Math.max(0, ...definition.phases.map((item) => item.order)) + 1;
		definition.phases.push({
			id: nextID(
				'phase',
				definition.phases.map((item) => item.id)
			),
			name: 'New phase',
			description: '',
			order,
			startsRound: false,
			suggestedDurationSeconds: 0
		});
	}
</script>

<ContentHeader density="dense" description="The ordered steps a game master advances through.">
	{#snippet title()}<h2>Phases</h2>{/snippet}
	{#snippet actions()}<Button variant="secondary" onclick={addPhase}>Add phase</Button>{/snippet}
</ContentHeader>
<div class="cards">
	{#each definition.phases as phase, index (phase.id)}
		<article class="item-card">
			<ContentHeader density="dense">
				{#snippet title()}<h3>{phase.order}. {phase.name || 'Unnamed phase'}</h3>{/snippet}
				{#snippet actions()}<button
						class="remove"
						onclick={() => removeAt(definition.phases, index)}>Remove</button
					>{/snippet}
			</ContentHeader>
			<div class="form-grid thirds">
				<Field label="Name" name={`phase-name-${index}`} bind:value={phase.name} required />
				<Field label="Stable ID" name={`phase-id-${index}`} bind:value={phase.id} required />
				<label><span>Order</span><input type="number" min="1" bind:value={phase.order} /></label>
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
						type="number"
						min="0"
						bind:value={phase.suggestedDurationSeconds}
					/></label
				>
				<CheckboxField
					label="Starts a new round"
					name={`phase-starts-round-${index}`}
					bind:checked={phase.startsRound}
				/>
				<SelectField
					label="Sound when phase starts"
					name={`phase-audio-${index}`}
					bind:value={phase.audioCueId}
					options={[
						{ value: '', label: 'No sound' },
						...definition.audioCues.map((cue) => ({ value: cue.id, label: cue.name }))
					]}
				/>
			</div>
		</article>
	{:else}
		<p class="empty">Add the phases used to run this game.</p>
	{/each}
</div>
