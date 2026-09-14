<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import VisualDefinitionEditor from '$lib/features/rulesets/components/VisualDefinitionEditor.svelte';
	import type { DefinitionEditorSection } from '$lib/features/rulesets/components/definition-editor';
	import type { ValidationIssue } from '$lib/features/rulesets/editor-state';

	let {
		section = 'roles',
		issues = [],
		onnavigate = () => {}
	}: {
		section?: DefinitionEditorSection;
		issues?: ValidationIssue[];
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
		chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
		achievements: [],
		audioCues: [],
		assetAccessibility: {}
	});
	let selectedItems = $state<Record<string, string>>({});
	const media = {
		upload: async () => {
			throw new Error('Media upload is not available in this harness.');
		},
		update: async () => {},
		remove: async () => {}
	};
</script>

<VisualDefinitionEditor
	bind:definition
	{section}
	assets={[]}
	{media}
	{issues}
	bind:selectedItems
	{onnavigate}
/>
