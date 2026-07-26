<script lang="ts">
	import VisualDefinitionEditor from '$lib/components/rulesets/VisualDefinitionEditor.svelte';
	import type { RulesetDefinition } from '$lib/api/types';

	let definition = $state<RulesetDefinition>({
		schemaVersion: 1,
		metadata: { name: 'Test', description: '', minPlayers: 3, maxPlayers: 12 },
		teams: [],
		categories: [],
		abilities: [],
		roles: [],
		phases: [],
		knowledgeRules: [],
		compositionBands: [],
		compositionModifiers: [],
		chat: { defaultPolicy: { teams: {} }, phaseOverrides: {} },
		achievements: [],
		audioCues: []
	});
</script>

<VisualDefinitionEditor {definition} section="achievements" assets={[]} />
<output aria-label="Achievement count">{definition.achievements.length}</output>
<output aria-label="Saved achievement points">{definition.achievements[0]?.points ?? 'none'}</output
>
<output aria-label="Achievement spoiler visibility">
	{definition.achievements[0]?.hiddenUntilGameCompleted ? 'hidden until complete' : 'visible live'}
</output>
