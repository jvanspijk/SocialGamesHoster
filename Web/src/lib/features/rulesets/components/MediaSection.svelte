<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import type { EditorSection } from '../editor-state';
	import type { AssetOption, MediaActions } from './definition-editor';
	import MediaLibrary from './MediaLibrary.svelte';

	let {
		definition = $bindable(),
		assets,
		media,
		onnavigate
	}: {
		definition: RulesetDefinition;
		assets: AssetOption[];
		media: MediaActions;
		onnavigate: (section: EditorSection, itemId?: string) => void;
	} = $props();
</script>

<MediaLibrary {definition} {assets} {media} {onnavigate} />
