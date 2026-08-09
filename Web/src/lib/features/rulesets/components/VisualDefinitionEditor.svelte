<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
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
		assets
	}: {
		definition: RulesetDefinition;
		section: DefinitionEditorSection;
		assets: AssetOption[];
	} = $props();
</script>

<DefinitionEditorLayout>
	{#if section === 'teams'}
		<TeamsSection bind:definition {assets} />
	{:else if section === 'roles'}
		<RolesSection bind:definition {assets} />
	{:else if section === 'phases'}
		<PhasesSection bind:definition />
	{:else if section === 'composition'}
		<CompositionSection bind:definition />
	{:else if section === 'knowledge'}
		<KnowledgeSection bind:definition />
	{:else if section === 'chat'}
		<ChatSection bind:definition />
	{:else if section === 'achievements'}
		<AchievementsSection bind:definition {assets} />
	{:else}
		<AudioSection bind:definition {assets} />
	{/if}
</DefinitionEditorLayout>
