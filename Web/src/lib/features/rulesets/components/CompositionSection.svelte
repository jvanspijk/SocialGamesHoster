<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import CheckboxGroup from '$lib/components/CheckboxGroup.svelte';
	import ContentHeader from '$lib/components/ContentHeader.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import type { EditorSection } from '../editor-state';
	import type { ValidationIssue } from '../editor-state';
	import CollectionEditor from './CollectionEditor.svelte';
	import {
		blankSelector,
		duplicateByID,
		incomingReferences,
		moveByID,
		nextID
	} from './definition-editor';
	import SelectorEditor from './SelectorEditor.svelte';
	let {
		definition = $bindable(),
		issues = [],
		selectedItems,
		onnavigate
	}: {
		definition: RulesetDefinition;
		issues?: ValidationIssue[];
		selectedItems: Record<string, string>;
		onnavigate: (section: EditorSection, itemId?: string) => void;
	} = $props();
	const bandEntries = $derived(
		definition.compositionBands.map((band) => ({
			id: band.id,
			label: `${band.minPlayers}–${band.maxPlayers} players`,
			supportingLabel: `${band.slots.length} slots`
		}))
	);
	const modifierEntries = $derived(
		definition.compositionModifiers.map((modifier, index) => ({
			id: modifier.id,
			label: `Conditional change ${index + 1}`,
			supportingLabel:
				definition.roles.find((role) => role.id === modifier.whenRolePresent)?.name ??
				'Choose a role'
		}))
	);
	function addBand() {
		const item = {
			id: nextID(
				'band',
				definition.compositionBands.map((value) => value.id)
			),
			minPlayers: definition.metadata.minPlayers,
			maxPlayers: definition.metadata.maxPlayers,
			slots: []
		};
		definition.compositionBands.push(item);
		selectedItems.compositionBands = item.id;
	}
	function duplicateBand(id: string) {
		const source = definition.compositionBands.find((item) => item.id === id);
		if (!source) return;
		const item = JSON.parse(JSON.stringify(source)) as typeof source;
		item.id = nextID(
			'band',
			definition.compositionBands.map((value) => value.id)
		);
		const used = definition.compositionBands.flatMap((band) => band.slots.map((slot) => slot.id));
		item.slots.forEach((slot) => {
			slot.id = nextID('slot', used);
			used.push(slot.id);
		});
		definition.compositionBands.splice(definition.compositionBands.indexOf(source) + 1, 0, item);
		selectedItems.compositionBands = item.id;
	}
	function addSlot(bandId: string) {
		const band = definition.compositionBands.find((item) => item.id === bandId);
		if (!band) return;
		band.slots.push({
			id: nextID(
				'slot',
				definition.compositionBands.flatMap((item) => item.slots.map((slot) => slot.id))
			),
			label: 'Role slot',
			count: 1,
			selector: blankSelector()
		});
	}
	function addModifier() {
		const item = {
			id: nextID(
				'modifier',
				definition.compositionModifiers.map((value) => value.id)
			),
			whenRolePresent: definition.roles[0]?.id ?? '',
			slotAdjustments: [],
			requiresRoleIds: [],
			excludesRoleIds: []
		};
		definition.compositionModifiers.push(item);
		selectedItems.compositionModifiers = item.id;
	}
	function addAdjustment(id: string) {
		definition.compositionModifiers
			.find((item) => item.id === id)
			?.slotAdjustments.push({
				slotId: definition.compositionBands[0]?.slots[0]?.id ?? '',
				delta: 1
			});
	}
	function bandUsages(id: string) {
		const band = definition.compositionBands.find((item) => item.id === id);
		const usages = (band?.slots ?? []).flatMap((slot) =>
			incomingReferences(definition, 'slot', slot.id)
		);
		return Array.from(
			new Map(
				usages.map((usage) => [
					`${usage.section}:${usage.itemId ?? usage.label}`,
					{
						label: usage.label,
						navigate: () => onnavigate(usage.section, usage.itemId)
					}
				])
			).values()
		);
	}
</script>

<CollectionEditor
	title="Player-count bands"
	description="Define how roles are filled for every supported party size."
	entries={bandEntries}
	selectedId={selectedItems.compositionBands ?? ''}
	onselect={(id) => (selectedItems.compositionBands = id)}
	onadd={addBand}
	addLabel="Add band"
	onduplicate={duplicateBand}
	onmove={(id, direction) => moveByID(definition.compositionBands, id, direction)}
	onremove={(id) =>
		definition.compositionBands.splice(
			definition.compositionBands.findIndex((item) => item.id === id),
			1
		)}
	usages={bandUsages}
	validationPath="compositionBands"
	itemPath={(id) =>
		`compositionBands[${definition.compositionBands.findIndex((item) => item.id === id)}]`}
	{issues}
	emptyDescription="Add a player-count band covering the supported player range."
>
	{#snippet editor(id)}{@const index = definition.compositionBands.findIndex(
			(item) => item.id === id
		)}{@const band = definition.compositionBands[index]}{#if band}<h3>
				{band.minPlayers}–{band.maxPlayers} players
			</h3>
			<div class="form-grid thirds">
				<label
					><span>Minimum players</span><input
						name={`band-min-${index}`}
						type="number"
						min="1"
						max="30"
						bind:value={band.minPlayers}
					/></label
				><label
					><span>Maximum players</span><input
						name={`band-max-${index}`}
						type="number"
						min="1"
						max="30"
						bind:value={band.maxPlayers}
					/></label
				>
			</div>
			<ContentHeader density="dense"
				>{#snippet title()}<strong>Role slots</strong>{/snippet}{#snippet actions()}<button
						class="add-small"
						onclick={() => addSlot(id)}>Add slot</button
					>{/snippet}</ContentHeader
			>{#each band.slots as slot, slotIndex (slot.id)}<div class="nested">
					<ContentHeader density="dense"
						>{#snippet title()}<strong>{slot.label || 'Unnamed slot'}</strong
							>{/snippet}{#snippet actions()}{@const references = incomingReferences(
								definition,
								'slot',
								slot.id
							)}{#if references.length}<span class="hint compact"
									>Used by
									{#each references as usage, usageIndex (usage.label + usageIndex)}<button
											class="usage"
											onclick={() => onnavigate(usage.section, usage.itemId)}>{usage.label}</button
										>{/each}</span
								>{:else}<button class="remove" onclick={() => band.slots.splice(slotIndex, 1)}
									>Remove</button
								>{/if}{/snippet}</ContentHeader
					>
					<div class="form-grid thirds">
						<Field
							label="Label"
							name={`slot-label-${index}-${slotIndex}`}
							bind:value={slot.label}
						/><label
							><span>Number of players</span><input
								name={`slot-count-${index}-${slotIndex}`}
								type="number"
								min="0"
								max="30"
								bind:value={slot.count}
							/></label
						>
					</div>
					<SelectorEditor
						selector={slot.selector}
						roles={definition.roles}
						teams={definition.teams}
						categories={definition.categories}
						label="Roles allowed in this slot"
						namePrefix={`slot-selector-${index}-${slotIndex}`}
					/>
				</div>{:else}<p class="hint">
					Add a slot and choose which roles can fill it.
				</p>{/each}{/if}{/snippet}
</CollectionEditor>

<CollectionEditor
	title="Conditional changes"
	description="Optionally adjust slots when a particular role appears."
	entries={modifierEntries}
	selectedId={selectedItems.compositionModifiers ?? ''}
	onselect={(id) => (selectedItems.compositionModifiers = id)}
	onadd={addModifier}
	addLabel="Add condition"
	onduplicate={(id) => {
		const item = duplicateByID(definition.compositionModifiers, id, 'modifier');
		if (item) selectedItems.compositionModifiers = item.id;
	}}
	onmove={(id, direction) => moveByID(definition.compositionModifiers, id, direction)}
	onremove={(id) =>
		definition.compositionModifiers.splice(
			definition.compositionModifiers.findIndex((item) => item.id === id),
			1
		)}
	validationPath="compositionModifiers"
	itemPath={(id) =>
		`compositionModifiers[${definition.compositionModifiers.findIndex((item) => item.id === id)}]`}
	{issues}
>
	{#snippet editor(id)}{@const index = definition.compositionModifiers.findIndex(
			(item) => item.id === id
		)}{@const modifier = definition.compositionModifiers[index]}{#if modifier}<h3>
				Conditional change {index + 1}
			</h3>
			<SelectField
				label="When this role is present"
				name={`modifier-role-${index}`}
				bind:value={modifier.whenRolePresent}
				options={[
					{ value: '', label: 'Choose a role' },
					...definition.roles.map((role) => ({ value: role.id, label: role.name }))
				]}
			/><CheckboxGroup
				label="Also require these roles"
				name={`modifier-required-roles-${index}`}
				bind:values={modifier.requiresRoleIds}
				options={definition.roles.map((role) => ({ value: role.id, label: role.name }))}
			/><CheckboxGroup
				label="Do not apply with these roles"
				name={`modifier-excluded-roles-${index}`}
				bind:values={modifier.excludesRoleIds}
				options={definition.roles.map((role) => ({ value: role.id, label: role.name }))}
			/><ContentHeader density="dense"
				>{#snippet title()}<strong>Slot changes</strong>{/snippet}{#snippet actions()}<button
						class="add-small"
						onclick={() => addAdjustment(id)}>Add change</button
					>{/snippet}</ContentHeader
			>{#each modifier.slotAdjustments as adjustment, adjustmentIndex (adjustment)}<div
					class="inline-row"
				>
					<SelectField
						label="Slot"
						name={`modifier-slot-${index}-${adjustmentIndex}`}
						bind:value={adjustment.slotId}
						options={[
							{ value: '', label: 'Choose a slot' },
							...definition.compositionBands.map((band) => ({
								label: `${band.minPlayers}–${band.maxPlayers} players`,
								options: band.slots.map((slot) => ({ value: slot.id, label: slot.label }))
							}))
						]}
					/><label
						><span>Change count by</span><input
							type="number"
							bind:value={adjustment.delta}
						/></label
					><button
						class="remove"
						onclick={() => modifier.slotAdjustments.splice(adjustmentIndex, 1)}>Remove</button
					>
				</div>{/each}{/if}{/snippet}
</CollectionEditor>

<style>
	h3 {
		margin: 0;
	}
	.usage {
		border: 0;
		background: transparent;
		color: inherit;
		cursor: pointer;
		padding: 0 0 0 var(--space-1);
		text-decoration: underline;
	}
</style>
