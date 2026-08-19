<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import type { EditorSection } from '../editor-state';
	import AchievementsSection from './AchievementsSection.svelte';
	import AudioSection from './AudioSection.svelte';
	import ChatSection from './ChatSection.svelte';
	import CompositionSection from './CompositionSection.svelte';
	import DefinitionEditorLayout from './DefinitionEditorLayout.svelte';
	import type { AssetOption, DefinitionEditorSection } from './definition-editor';
	import KnowledgeSection from './KnowledgeSection.svelte';
	import PhasesSection from './PhasesSection.svelte';
	import RolesSection from './RolesSection.svelte';
	import TeamsSection from './TeamsSection.svelte';

	let {
		definition = $bindable(),
		section,
		assets,
		selectedItems,
		onnavigate
	}: {
		definition: RulesetDefinition;
		section: DefinitionEditorSection;
		assets: AssetOption[];
		selectedItems: Record<string, string>;
		onnavigate: (section: EditorSection, itemId?: string) => void;
	} = $props();
</script>

<DefinitionEditorLayout>
	{#if section === 'teams'}
		<TeamsSection bind:definition {assets} {selectedItems} {onnavigate} />
	{:else if section === 'roles'}
		<RolesSection bind:definition {assets} {selectedItems} {onnavigate} />
	{:else if section === 'phases'}
		<PhasesSection bind:definition {selectedItems} {onnavigate} />
	{:else if section === 'composition'}
		<CompositionSection bind:definition {selectedItems} {onnavigate} />
	{:else if section === 'knowledge'}
		<KnowledgeSection bind:definition />
	{:else if section === 'chat'}
		<ChatSection bind:definition {selectedItems} />
	{:else if section === 'achievements'}
		<AchievementsSection bind:definition {assets} {selectedItems} />
	{:else}
		<AudioSection bind:definition {assets} {selectedItems} {onnavigate} />
	{/if}
</DefinitionEditorLayout>
