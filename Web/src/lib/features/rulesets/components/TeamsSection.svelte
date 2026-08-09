<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import Button from '$lib/components/Button.svelte';
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

	function addTeam() {
		definition.teams.push({
			id: nextID(
				'team',
				definition.teams.map((item) => item.id)
			),
			name: 'New team',
			description: ''
		});
	}

	function addCategory() {
		definition.categories.push({
			id: nextID(
				'category',
				definition.categories.map((item) => item.id)
			),
			name: 'New category',
			description: ''
		});
	}
</script>

<ContentHeader density="dense" description="The main sides or factions in the game.">
	{#snippet title()}<h2>Teams</h2>{/snippet}
	{#snippet actions()}<Button variant="secondary" onclick={addTeam}>Add team</Button>{/snippet}
</ContentHeader>
<div class="cards">
	{#each definition.teams as team, index (team.id)}
		<article class="item-card">
			<ContentHeader density="dense">
				{#snippet title()}<h3>{team.name || 'Unnamed team'}</h3>{/snippet}
				{#snippet actions()}<button class="remove" onclick={() => removeAt(definition.teams, index)}
						>Remove</button
					>{/snippet}
			</ContentHeader>
			<div class="form-grid">
				<Field label="Name" name={`team-name-${index}`} bind:value={team.name} required />
				<Field
					label="Stable ID"
					name={`team-id-${index}`}
					bind:value={team.id}
					help="Used when other rules refer to this team."
					required
				/>
			</div>
			<Field
				label="Description"
				name={`team-description-${index}`}
				bind:value={team.description}
				multiline
			/>
			<SelectField
				label="Team image (optional)"
				name={`team-image-${index}`}
				bind:value={team.imageAssetKey}
				options={imageOptions()}
			/>
		</article>
	{:else}
		<p class="empty">Add at least one team before creating roles.</p>
	{/each}
</div>

<div class="subsection">
	<ContentHeader density="dense" description="Optional labels such as Investigative or Support.">
		{#snippet title()}<h2>Categories</h2>{/snippet}
		{#snippet actions()}<Button variant="secondary" onclick={addCategory}>Add category</Button
			>{/snippet}
	</ContentHeader>
</div>
<div class="cards compact">
	{#each definition.categories as category, index (category.id)}
		<article class="item-card">
			<ContentHeader density="dense">
				{#snippet title()}<h3>{category.name || 'Unnamed category'}</h3>{/snippet}
				{#snippet actions()}<button
						class="remove"
						onclick={() => removeAt(definition.categories, index)}>Remove</button
					>{/snippet}
			</ContentHeader>
			<div class="form-grid">
				<Field label="Name" name={`category-name-${index}`} bind:value={category.name} required />
				<Field label="Stable ID" name={`category-id-${index}`} bind:value={category.id} required />
			</div>
			<Field
				label="Description"
				name={`category-description-${index}`}
				bind:value={category.description}
				multiline
			/>
		</article>
	{/each}
</div>
