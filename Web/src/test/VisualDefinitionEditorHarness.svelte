<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import VisualDefinitionEditor from '$lib/features/rulesets/components/VisualDefinitionEditor.svelte';
	import type { DefinitionEditorSection } from '$lib/features/rulesets/components/definition-editor';

	let { section = 'composition' }: { section?: DefinitionEditorSection } = $props();

	let definition = $state<RulesetDefinition>({
		schemaVersion: 1,
		metadata: { name: 'Test ruleset', description: '', minPlayers: 3, maxPlayers: 12 },
		teams: [],
		categories: [],
		abilities: [],
		roles: [],
		phases: [],
		knowledgeRules: [],
		compositionBands: [
			{
				id: 'band-1',
				minPlayers: 3,
				maxPlayers: 8,
				slots: []
			}
		],
		compositionModifiers: [
			{
				id: 'modifier-1',
				whenRolePresent: '',
				slotAdjustments: [],
				requiresRoleIds: [],
				excludesRoleIds: []
			}
		],
		chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
		achievements: [],
		audioCues: [],
		assetAccessibility: {}
	});
</script>

<VisualDefinitionEditor bind:definition {section} assets={[]} />
