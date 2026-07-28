<script lang="ts">
	import VisualDefinitionEditor from '$lib/components/rulesets/VisualDefinitionEditor.svelte';
	import type { RulesetDefinition } from '$lib/api/types';

	let definition = $state<RulesetDefinition>({
		schemaVersion: 1,
		metadata: { name: 'Test', description: '', minPlayers: 1, maxPlayers: 2 },
		teams: [{ id: 'town', name: 'Town', description: '' }],
		categories: [],
		abilities: [],
		roles: [],
		phases: [{ id: 'night', name: 'Night', description: '', order: 1, startsRound: false }],
		knowledgeRules: [],
		compositionBands: [],
		compositionModifiers: [],
		chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
		achievements: [],
		audioCues: [],
		assetAccessibility: {}
	});
</script>

<VisualDefinitionEditor bind:definition section="roles" assets={[]} />
<output aria-label="Ability phases"
	>{definition.abilities[0]?.activationPhaseIds?.join(',') ?? ''}</output
>
<output aria-label="Ability combination">
	{definition.abilities[0]?.canCombineWithOtherAbilities ? 'combinable' : 'exclusive'}
</output>
