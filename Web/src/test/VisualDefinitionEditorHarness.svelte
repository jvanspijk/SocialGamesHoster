<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import VisualDefinitionEditor from '$lib/features/rulesets/components/VisualDefinitionEditor.svelte';
	import type { DefinitionEditorSection } from '$lib/features/rulesets/components/definition-editor';

	let {
		section = 'composition',
		onnavigate = () => {}
	}: {
		section?: DefinitionEditorSection;
		onnavigate?: (section: string, itemId?: string) => void;
	} = $props();

	let definition = $state<RulesetDefinition>({
		schemaVersion: 1,
		metadata: { name: 'Test ruleset', description: '', minPlayers: 3, maxPlayers: 12 },
		teams: [],
		categories: [],
		abilities: [],
		roles: [],
		phases: [],
		knowledgeRules: [
			{
				viewer: { roleIds: [], teamIds: [], categoryIds: [], tags: [] },
				target: { roleIds: [], teamIds: [], categoryIds: [], tags: [] },
				reveal: ['role']
			}
		],
		compositionBands: [
			{
				id: 'band-1',
				minPlayers: 3,
				maxPlayers: 8,
				slots: [
					{
						id: 'slot-1',
						label: 'Villager slot',
						count: 1,
						selector: { roleIds: [], teamIds: [], categoryIds: [], tags: [] }
					}
				]
			}
		],
		compositionModifiers: [
			{
				id: 'modifier-1',
				whenRolePresent: '',
				slotAdjustments: [{ slotId: 'slot-1', delta: 1 }],
				requiresRoleIds: [],
				excludesRoleIds: []
			}
		],
		chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
		achievements: [],
		audioCues: [],
		assetAccessibility: {}
	});
	let selectedItems = $state<Record<string, string>>({ compositionBands: 'band-1' });
</script>

<VisualDefinitionEditor bind:definition {section} assets={[]} {selectedItems} {onnavigate} />
