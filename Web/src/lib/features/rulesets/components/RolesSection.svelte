<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import CheckboxField from '$lib/components/CheckboxField.svelte';
	import CheckboxGroup from '$lib/components/CheckboxGroup.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
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
	const abilityEntries = $derived(
		definition.abilities.map((item) => ({
			id: item.id,
			label: item.name || 'Unnamed ability',
			supportingLabel: `${definition.roles.filter((role) => role.abilityIds.includes(item.id)).length} roles`
		}))
	);
	const roleEntries = $derived(
		definition.roles.map((item) => ({
			id: item.id,
			label: item.name || 'Unnamed role',
			supportingLabel: definition.teams.find((team) => team.id === item.teamId)?.name ?? 'No team'
		}))
	);
	function addAbility() {
		const item = {
			id: nextID(
				'ability',
				definition.abilities.map((value) => value.id)
			),
			name: 'New ability',
			description: '',
			activationPhaseIds: [],
			canCombineWithOtherAbilities: false
		};
		definition.abilities.push(item);
		selectedItems.abilities = item.id;
	}
	function addRole() {
		const item = {
			id: nextID(
				'role',
				definition.roles.map((value) => value.id)
			),
			name: 'New role',
			description: '',
			teamId: definition.teams[0]?.id ?? '',
			categoryIds: [],
			tags: [],
			abilityIds: [],
			winCondition: '',
			maxCopies: 1
		};
		definition.roles.push(item);
		selectedItems.roles = item.id;
	}
	function refs(kind: 'ability' | 'role', id: string) {
		return incomingReferences(definition, kind, id).map((usage) => ({
			label: usage.label,
			navigate: () => onnavigate(usage.section, usage.itemId)
		}));
	}
	function setTags(index: number, value: string) {
		definition.roles[index].tags = value
			.split(',')
			.map((tag) => tag.trim())
			.filter(Boolean);
	}
</script>

<CollectionEditor
	title="Abilities"
	description="Optional reusable powers assigned to roles."
	entries={abilityEntries}
	selectedId={selectedItems.abilities ?? ''}
	onselect={(id) => (selectedItems.abilities = id)}
	onadd={addAbility}
	onduplicate={(id) => {
		const item = duplicateByID(definition.abilities, id, 'ability');
		if (item) selectedItems.abilities = item.id;
	}}
	onmove={(id, direction) => moveByID(definition.abilities, id, direction)}
	onremove={(id) =>
		definition.abilities.splice(
			definition.abilities.findIndex((item) => item.id === id),
			1
		)}
	usages={(id) => refs('ability', id)}
>
	{#snippet editor(id)}{@const index = definition.abilities.findIndex(
			(item) => item.id === id
		)}{@const ability = definition.abilities[index]}{#if ability}<h3>
				{ability.name || 'Unnamed ability'}
			</h3>
			<Field label="Name" name={`ability-name-${index}`} bind:value={ability.name} required /><Field
				label="Description"
				name={`ability-description-${index}`}
				bind:value={ability.description}
				multiline
			/><MediaField
				label="Ability image"
				kind="image"
				name={`ability-image-${index}`}
				bind:value={ability.imageAssetKey}
				{assets}
				{media}
			/><CheckboxField
				label="May combine with other combinable abilities"
				name={`ability-combinable-${index}`}
				checked={ability.canCombineWithOtherAbilities ?? false}
				onchange={(checked) => (ability.canCombineWithOtherAbilities = checked)}
			/><CheckboxGroup
				label="Playable during phases"
				name={`ability-phases-${index}`}
				bind:values={ability.activationPhaseIds}
				options={definition.phases.map((phase) => ({ value: phase.id, label: phase.name }))}
			/>{/if}{/snippet}
</CollectionEditor>

{#if definition.teams.length === 0}
	<EmptyState
		title="Add a team first"
		description="Every role belongs to a team."
		actionLabel="Go to Teams"
		onaction={() => onnavigate('teams')}
	/>
{:else}
	<CollectionEditor
		title="Roles"
		description="What each player may be assigned and how that role wins."
		entries={roleEntries}
		selectedId={selectedItems.roles ?? ''}
		onselect={(id) => (selectedItems.roles = id)}
		onadd={addRole}
		onduplicate={(id) => {
			const item = duplicateByID(definition.roles, id, 'role');
			if (item) selectedItems.roles = item.id;
		}}
		onmove={(id, direction) => moveByID(definition.roles, id, direction)}
		onremove={(id) =>
			definition.roles.splice(
				definition.roles.findIndex((item) => item.id === id),
				1
			)}
		usages={(id) => refs('role', id)}
		emptyDescription="Add the first role players can receive."
	>
		{#snippet editor(id)}{@const index = definition.roles.findIndex(
				(item) => item.id === id
			)}{@const role = definition.roles[index]}{#if role}<h3>{role.name || 'Unnamed role'}</h3>
				<div class="form-grid thirds">
					<Field
						label="Name"
						name={`role-name-${index}`}
						bind:value={role.name}
						required
					/><SelectField
						label="Team"
						name={`role-team-${index}`}
						bind:value={role.teamId}
						options={[
							{ value: '', label: 'Choose a team' },
							...definition.teams.map((team) => ({ value: team.id, label: team.name }))
						]}
						required
					/>
				</div>
				<Field
					label="Description"
					name={`role-description-${index}`}
					bind:value={role.description}
					multiline
				/><Field
					label="Win condition"
					name={`role-win-${index}`}
					bind:value={role.winCondition}
					multiline
				/>
				<div class="form-grid">
					<label
						><span>Maximum copies</span><input
							name={`role-max-copies-${index}`}
							type="number"
							min="1"
							max="30"
							bind:value={role.maxCopies}
						/></label
					>
				</div>
				<MediaField
					label="Role image"
					kind="image"
					name={`role-image-${index}`}
					bind:value={role.imageAssetKey}
					{assets}
					{media}
				/>
				<CheckboxGroup
					label="Categories"
					name={`role-categories-${index}`}
					bind:values={role.categoryIds}
					options={definition.categories.map((category) => ({
						value: category.id,
						label: category.name
					}))}
				/><CheckboxGroup
					label="Abilities"
					name={`role-abilities-${index}`}
					bind:values={role.abilityIds}
					options={definition.abilities.map((ability) => ({
						value: ability.id,
						label: ability.name
					}))}
				/><label
					><span>Tags (comma-separated)</span><input
						value={role.tags.join(', ')}
						onchange={(event) => setTags(index, event.currentTarget.value)}
						placeholder="investigative, unique"
					/></label
				>{/if}{/snippet}
	</CollectionEditor>
{/if}

<style>
	h3 {
		margin: 0;
	}
</style>
