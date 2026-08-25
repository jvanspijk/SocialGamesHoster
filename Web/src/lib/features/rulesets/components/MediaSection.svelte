<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import type { EditorSection } from '../editor-state';
	import type { ValidationIssue } from '../editor-state';
	import type { AssetOption, MediaActions } from './definition-editor';
	import AudioSection from './AudioSection.svelte';
	import MediaLibrary from './MediaLibrary.svelte';

	let {
		definition = $bindable(),
		assets,
		media,
		issues = [],
		selectedItems,
		onnavigate
	}: {
		definition: RulesetDefinition;
		assets: AssetOption[];
		media: MediaActions;
		issues?: ValidationIssue[];
		selectedItems: Record<string, string>;
		onnavigate: (section: EditorSection, itemId?: string) => void;
	} = $props();
</script>

<MediaLibrary {definition} {assets} {media} {onnavigate} />
<div class="cues">
	<AudioSection bind:definition {assets} {media} {issues} {selectedItems} {onnavigate} />
</div>

<style>
	.cues {
		border-top: var(--border-strong);
		padding-top: var(--space-5);
	}
</style>
