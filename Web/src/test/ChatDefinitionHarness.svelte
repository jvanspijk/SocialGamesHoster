<script lang="ts">
	import VisualDefinitionEditor from '$lib/components/rulesets/VisualDefinitionEditor.svelte';
	import type { RulesetDefinition } from '$lib/api/types';

	let definition = $state<RulesetDefinition>({
		schemaVersion: 1,
		metadata: { name: 'Test', description: '', minPlayers: 3, maxPlayers: 12 },
		teams: [{ id: 'town', name: 'Town', description: '' }],
		categories: [],
		abilities: [],
		roles: [
			{
				id: 'villager',
				name: 'Villager',
				description: '',
				teamId: 'town',
				categoryIds: [],
				tags: [],
				abilityIds: [],
				winCondition: '',
				maxCopies: 12
			}
		],
		phases: [
			{
				id: 'night',
				name: 'Night',
				description: '',
				order: 1,
				startsRound: true
			}
		],
		knowledgeRules: [],
		compositionBands: [],
		compositionModifiers: [],
		chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
		achievements: [],
		audioCues: []
	});
</script>

<VisualDefinitionEditor bind:definition section="chat" assets={[]} />
<output aria-label="Custom channel count">{definition.chat.channels.length}</output>
<output aria-label="Custom channel restriction">
	{definition.chat.channels[0]?.messageRestriction ?? 'none'}
</output>
<output aria-label="Custom channel night visibility">
	{definition.chat.channels[0]?.phaseOverrides.night?.visible === false ? 'hidden' : 'not hidden'}
</output>
