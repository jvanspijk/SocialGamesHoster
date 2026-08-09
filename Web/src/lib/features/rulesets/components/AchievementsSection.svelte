<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import Button from '$lib/components/Button.svelte';
	import CheckboxField from '$lib/components/CheckboxField.svelte';
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
	function addAchievement() {
		definition.achievements.push({
			id: nextID(
				'achievement',
				definition.achievements.map((item) => item.id)
			),
			name: 'New achievement',
			description: '',
			points: 0,
			hiddenUntilGameCompleted: false
		});
	}
</script>

<ContentHeader density="dense" description="Awards a game master can give after a game.">
	{#snippet title()}<h2>Achievements</h2>{/snippet}
	{#snippet actions()}<Button variant="secondary" onclick={addAchievement}>Add achievement</Button
		>{/snippet}
</ContentHeader>
<div class="cards">
	{#each definition.achievements as achievement, index (achievement.id)}
		<article class="item-card">
			<ContentHeader density="dense">
				{#snippet title()}<h3>{achievement.name || 'Unnamed achievement'}</h3>{/snippet}
				{#snippet actions()}<button
						class="remove"
						onclick={() => removeAt(definition.achievements, index)}>Remove</button
					>{/snippet}
			</ContentHeader>
			<div class="form-grid">
				<Field
					label="Name"
					name={`achievement-name-${index}`}
					bind:value={achievement.name}
					required
				/><Field
					label="Stable ID"
					name={`achievement-id-${index}`}
					bind:value={achievement.id}
					required
				/>
			</div>
			<Field
				label="Description"
				name={`achievement-description-${index}`}
				bind:value={achievement.description}
				multiline
			/>
			<div class="form-grid">
				<label
					><span>Achievement points</span><input
						type="number"
						min="0"
						max="10000"
						bind:value={achievement.points}
					/></label
				><CheckboxField
					label="Hide from players until the game ends"
					name={`achievement-hidden-${index}`}
					bind:checked={achievement.hiddenUntilGameCompleted}
				/>
			</div>
			<SelectField
				label="Badge image (optional)"
				name={`achievement-image-${index}`}
				bind:value={achievement.imageAssetKey}
				options={imageOptions()}
			/>
		</article>
	{:else}
		<p class="empty">No achievements in this ruleset.</p>
	{/each}
</div>
