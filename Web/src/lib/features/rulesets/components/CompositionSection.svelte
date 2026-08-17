<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import Button from '$lib/components/Button.svelte';
	import CheckboxGroup from '$lib/components/CheckboxGroup.svelte';
	import ContentHeader from '$lib/components/ContentHeader.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import { blankSelector, nextID, removeAt } from './definition-editor';
	import SelectorEditor from './SelectorEditor.svelte';

	let { definition = $bindable() }: { definition: RulesetDefinition } = $props();
	function addBand() {
		definition.compositionBands.push({
			id: nextID(
				'band',
				definition.compositionBands.map((item) => item.id)
			),
			minPlayers: definition.metadata.minPlayers,
			maxPlayers: definition.metadata.maxPlayers,
			slots: []
		});
	}
	function addSlot(bandIndex: number) {
		const band = definition.compositionBands[bandIndex];
		const used = definition.compositionBands.flatMap((item) => item.slots.map((slot) => slot.id));
		band.slots.push({
			id: nextID('slot', used),
			label: 'Role slot',
			count: 1,
			selector: blankSelector()
		});
	}
	function addModifier() {
		definition.compositionModifiers.push({
			id: nextID(
				'modifier',
				definition.compositionModifiers.map((item) => item.id)
			),
			whenRolePresent: definition.roles[0]?.id ?? '',
			slotAdjustments: [],
			requiresRoleIds: [],
			excludesRoleIds: []
		});
	}
	function addAdjustment(modifierIndex: number) {
		definition.compositionModifiers[modifierIndex].slotAdjustments.push({
			slotId: definition.compositionBands[0]?.slots[0]?.id ?? '',
			delta: 1
		});
	}
</script>

<ContentHeader
	density="dense"
	description="Define how many slots are filled for every supported party size."
>
	{#snippet title()}<h2>Player-count bands</h2>{/snippet}
	{#snippet actions()}<Button variant="secondary" onclick={addBand}>Add band</Button>{/snippet}
</ContentHeader>
<div class="cards">
	{#each definition.compositionBands as band, bandIndex (band.id)}
		<article class="item-card">
			<ContentHeader density="dense">
				{#snippet title()}<h3>{band.minPlayers}–{band.maxPlayers} players</h3>{/snippet}
				{#snippet actions()}<button
						class="remove"
						onclick={() => removeAt(definition.compositionBands, bandIndex)}>Remove</button
					>{/snippet}
			</ContentHeader>
			<div class="form-grid thirds">
				<label
					><span>Minimum players</span><input
						type="number"
						min="1"
						max="30"
						bind:value={band.minPlayers}
					/></label
				><label
					><span>Maximum players</span><input
						type="number"
						min="1"
						max="30"
						bind:value={band.maxPlayers}
					/></label
				>
			</div>
			<div class="nested-heading">
				<ContentHeader density="dense"
					>{#snippet title()}<strong>Role slots</strong>{/snippet}{#snippet actions()}<button
							class="add-small"
							onclick={() => addSlot(bandIndex)}>Add slot</button
						>{/snippet}</ContentHeader
				>
			</div>
			{#each band.slots as slot, slotIndex (slot.id)}
				<div class="nested">
					<ContentHeader density="dense"
						>{#snippet title()}<strong>{slot.label || 'Unnamed slot'}</strong
							>{/snippet}{#snippet actions()}<button
								class="remove"
								onclick={() => removeAt(band.slots, slotIndex)}>Remove</button
							>{/snippet}</ContentHeader
					>
					<div class="form-grid thirds">
						<Field
							label="Label"
							name={`slot-label-${bandIndex}-${slotIndex}`}
							bind:value={slot.label}
						/><label
							><span>Number of players</span><input
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
					/>
				</div>
			{/each}
		</article>
	{/each}
</div>

<div class="subsection">
	<ContentHeader density="dense" description="Adjust slots when a particular role appears."
		>{#snippet title()}<h2>Conditional changes</h2>{/snippet}{#snippet actions()}<Button
				variant="secondary"
				onclick={addModifier}>Add condition</Button
			>{/snippet}</ContentHeader
	>
</div>
<div class="cards">
	{#each definition.compositionModifiers as modifier, modifierIndex (modifier.id)}
		<article class="item-card">
			<ContentHeader density="dense"
				>{#snippet title()}<h3>Conditional change</h3>{/snippet}{#snippet actions()}<button
						class="remove"
						onclick={() => removeAt(definition.compositionModifiers, modifierIndex)}>Remove</button
					>{/snippet}</ContentHeader
			>
			<div class="form-grid">
				<SelectField
					label="When this role is present"
					name={`modifier-role-${modifierIndex}`}
					bind:value={modifier.whenRolePresent}
					options={[
						{ value: '', label: 'Choose a role' },
						...definition.roles.map((role) => ({ value: role.id, label: role.name }))
					]}
				/>
			</div>
			<div class="choice-block">
				<CheckboxGroup
					label="Also require these roles"
					name={`modifier-required-roles-${modifierIndex}`}
					bind:values={modifier.requiresRoleIds}
					options={definition.roles.map((role) => ({ value: role.id, label: role.name }))}
				/>
			</div>
			<div class="choice-block">
				<CheckboxGroup
					label="Do not apply with these roles"
					name={`modifier-excluded-roles-${modifierIndex}`}
					bind:values={modifier.excludesRoleIds}
					options={definition.roles.map((role) => ({ value: role.id, label: role.name }))}
				/>
			</div>
			<div class="nested-heading">
				<ContentHeader density="dense"
					>{#snippet title()}<strong>Slot changes</strong>{/snippet}{#snippet actions()}<button
							class="add-small"
							onclick={() => addAdjustment(modifierIndex)}>Add change</button
						>{/snippet}</ContentHeader
				>
			</div>
			{#each modifier.slotAdjustments as adjustment, adjustmentIndex (adjustment)}
				<div class="inline-row">
					<SelectField
						label="Slot"
						name={`modifier-slot-${modifierIndex}-${adjustmentIndex}`}
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
						onclick={() => removeAt(modifier.slotAdjustments, adjustmentIndex)}>Remove</button
					>
				</div>
			{/each}
		</article>
	{/each}
</div>
