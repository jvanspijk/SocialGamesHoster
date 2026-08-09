<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import Button from '$lib/components/Button.svelte';
	import CheckboxField from '$lib/components/CheckboxField.svelte';
	import CheckboxGroup from '$lib/components/CheckboxGroup.svelte';
	import ContentHeader from '$lib/components/ContentHeader.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import { nextID, removeAt, type AssetOption } from './definition-editor';

	let {
		definition = $bindable(),
		assets
	}: { definition: RulesetDefinition; assets: AssetOption[] } = $props();
	const imageOptions = () => [
		{ value: '', label: 'No image' },
		...assets
			.filter((asset) => asset.kind === 'image')
			.map((asset) => ({ value: asset.assetKey, label: asset.assetKey }))
	];
	function addAbility() {
		definition.abilities.push({
			id: nextID(
				'ability',
				definition.abilities.map((item) => item.id)
			),
			name: 'New ability',
			description: '',
			activationPhaseIds: [],
			canCombineWithOtherAbilities: false
		});
	}
	function addRole() {
		definition.roles.push({
			id: nextID(
				'role',
				definition.roles.map((item) => item.id)
			),
			name: 'New role',
			description: '',
			teamId: definition.teams[0]?.id ?? '',
			categoryIds: [],
			tags: [],
			abilityIds: [],
			winCondition: '',
			maxCopies: 1
		});
	}
	function setTags(index: number, value: string) {
		definition.roles[index].tags = value
			.split(',')
			.map((tag) => tag.trim())
			.filter(Boolean);
	}
</script>

<ContentHeader density="dense" description="Reusable powers that can be assigned to roles.">
	{#snippet title()}<h2>Abilities</h2>{/snippet}
	{#snippet actions()}<Button variant="secondary" onclick={addAbility}>Add ability</Button
		>{/snippet}
</ContentHeader>
<div class="cards compact">
	{#each definition.abilities as ability, index (ability.id)}
		<article class="item-card">
			<ContentHeader density="dense">
				{#snippet title()}<h3>{ability.name || 'Unnamed ability'}</h3>{/snippet}
				{#snippet actions()}<button
						class="remove"
						onclick={() => removeAt(definition.abilities, index)}>Remove</button
					>{/snippet}
			</ContentHeader>
			<div class="form-grid">
				<Field
					label="Name"
					name={`ability-name-${index}`}
					bind:value={ability.name}
					required
				/><Field label="Stable ID" name={`ability-id-${index}`} bind:value={ability.id} required />
			</div>
			<Field
				label="Description"
				name={`ability-description-${index}`}
				bind:value={ability.description}
				multiline
			/>
			<SelectField
				label="Ability image (optional)"
				name={`ability-image-${index}`}
				bind:value={ability.imageAssetKey}
				options={imageOptions()}
			/>
			<CheckboxField
				label="May combine with other combinable abilities"
				name={`ability-combinable-${index}`}
				checked={ability.canCombineWithOtherAbilities ?? false}
				onchange={(checked) => (ability.canCombineWithOtherAbilities = checked)}
			/>
			<div class="choice-block">
				<CheckboxGroup
					label="Playable during phases"
					name={`ability-phases-${index}`}
					bind:values={ability.activationPhaseIds}
					options={definition.phases.map((phase) => ({ value: phase.id, label: phase.name }))}
				/>{#if definition.phases.length === 0}<p class="hint compact">
						Add phases before making this ability playable.
					</p>{/if}
			</div>
		</article>
	{/each}
</div>

<div class="subsection">
	<ContentHeader
		density="dense"
		description="What each player may be assigned and how that role wins."
	>
		{#snippet title()}<h2>Roles</h2>{/snippet}
		{#snippet actions()}<Button variant="secondary" onclick={addRole}>Add role</Button>{/snippet}
	</ContentHeader>
</div>
<div class="cards">
	{#each definition.roles as role, index (role.id)}
		<article class="item-card">
			<ContentHeader density="dense">
				{#snippet title()}<h3>{role.name || 'Unnamed role'}</h3>{/snippet}
				{#snippet actions()}<button class="remove" onclick={() => removeAt(definition.roles, index)}
						>Remove</button
					>{/snippet}
			</ContentHeader>
			<div class="form-grid thirds">
				<Field label="Name" name={`role-name-${index}`} bind:value={role.name} required /><Field
					label="Stable ID"
					name={`role-id-${index}`}
					bind:value={role.id}
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
			/>
			<Field
				label="Win condition"
				name={`role-win-${index}`}
				bind:value={role.winCondition}
				multiline
			/>
			<div class="form-grid">
				<label
					><span>Maximum copies</span><input
						type="number"
						min="1"
						max="30"
						bind:value={role.maxCopies}
					/></label
				><SelectField
					label="Role image (optional)"
					name={`role-image-${index}`}
					bind:value={role.imageAssetKey}
					options={imageOptions()}
				/>
			</div>
			<div class="choice-block">
				<CheckboxGroup
					label="Categories"
					name={`role-categories-${index}`}
					bind:values={role.categoryIds}
					options={definition.categories.map((category) => ({
						value: category.id,
						label: category.name
					}))}
				/>
			</div>
			<div class="choice-block">
				<CheckboxGroup
					label="Abilities"
					name={`role-abilities-${index}`}
					bind:values={role.abilityIds}
					options={definition.abilities.map((ability) => ({
						value: ability.id,
						label: ability.name
					}))}
				/>
			</div>
			<label
				><span>Tags (comma-separated)</span><input
					value={role.tags.join(', ')}
					onchange={(event) => setTags(index, event.currentTarget.value)}
					placeholder="investigative, unique"
				/></label
			>
		</article>
	{:else}
		<p class="empty">Add roles after defining at least one team.</p>
	{/each}
</div>
