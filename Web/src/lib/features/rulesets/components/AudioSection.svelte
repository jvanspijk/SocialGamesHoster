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
	const audioAssets = () => assets.filter((asset) => asset.kind === 'audio');
	function addAudioCue() {
		definition.audioCues.push({
			id: nextID(
				'cue',
				definition.audioCues.map((item) => item.id)
			),
			name: 'New audio cue',
			assetKey: audioAssets()[0]?.assetKey ?? '',
			defaultAudience: 'all'
		});
	}
</script>

<ContentHeader
	density="dense"
	description="Named sounds a game master or phase can play for selected listeners."
>
	{#snippet title()}<h2>Audio cues</h2>{/snippet}
	{#snippet actions()}<Button variant="secondary" onclick={addAudioCue}>Add audio cue</Button
		>{/snippet}
</ContentHeader>
<div class="cards">
	{#each definition.audioCues as cue, index (cue.id)}
		<article class="item-card">
			<ContentHeader density="dense">
				{#snippet title()}<h3>{cue.name || 'Unnamed cue'}</h3>{/snippet}
				{#snippet actions()}<button
						class="remove"
						onclick={() => removeAt(definition.audioCues, index)}>Remove</button
					>{/snippet}
			</ContentHeader>
			<div class="form-grid">
				<Field label="Name" name={`cue-name-${index}`} bind:value={cue.name} required />
				<SelectField
					label="Audio file"
					name={`cue-audio-file-${index}`}
					bind:value={cue.assetKey}
					options={[
						{ value: '', label: 'Choose uploaded audio' },
						...audioAssets().map((asset) => ({ value: asset.assetKey, label: asset.assetKey }))
					]}
				/>
				<SelectField
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
			{#if cue.defaultAudience === 'team' || cue.defaultAudience === 'player'}<p class="hint">
					This cue needs a team or player chosen when a game master plays it. It cannot be used as
					an automatic phase sound.
				</p>{/if}
		</article>
	{:else}
		<p class="empty">Upload an audio file, then add a cue that uses it.</p>
	{/each}
</div>
