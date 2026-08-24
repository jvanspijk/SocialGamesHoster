<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import type { EditorSection } from '../editor-state';
	import AchievementsSection from './AchievementsSection.svelte';
	import ChatSection from './ChatSection.svelte';
	import CompositionSection from './CompositionSection.svelte';
	import DefinitionEditorLayout from './DefinitionEditorLayout.svelte';
	import type { AssetOption, DefinitionEditorSection, MediaActions } from './definition-editor';
	import KnowledgeSection from './KnowledgeSection.svelte';
	import MediaSection from './MediaSection.svelte';
	import PhasesSection from './PhasesSection.svelte';
	import RolesSection from './RolesSection.svelte';
	import TeamsSection from './TeamsSection.svelte';

	let {
		definition = $bindable(),
		section,
		assets,
		media,
		selectedItems,
		onnavigate
	}: {
		definition: RulesetDefinition;
		section: DefinitionEditorSection;
		assets: AssetOption[];
		media: MediaActions;
		selectedItems: Record<string, string>;
		onnavigate: (section: EditorSection, itemId?: string) => void;
	} = $props();
</script>

<DefinitionEditorLayout>
	{#if section === 'teams'}
		<TeamsSection bind:definition {assets} {media} {selectedItems} {onnavigate} />
	{:else if section === 'roles'}
		<RolesSection bind:definition {assets} {media} {selectedItems} {onnavigate} />
	{:else if section === 'phases'}
		<PhasesSection bind:definition {media} {selectedItems} {onnavigate} />
	{:else if section === 'composition'}
		<CompositionSection bind:definition {selectedItems} {onnavigate} />
	{:else if section === 'knowledge'}
		<KnowledgeSection bind:definition />
	{:else if section === 'chat'}
		<ChatSection bind:definition {selectedItems} />
	{:else if section === 'achievements'}
		<AchievementsSection bind:definition {assets} {media} {selectedItems} />
	{:else}
		<MediaSection bind:definition {assets} {media} {selectedItems} {onnavigate} />
	{/if}
</DefinitionEditorLayout>
