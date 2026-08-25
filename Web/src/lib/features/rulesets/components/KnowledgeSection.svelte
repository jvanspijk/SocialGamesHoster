<script lang="ts">
	import type { RulesetDefinition } from '$lib/api/types';
	import Button from '$lib/components/Button.svelte';
	import CheckboxGroup from '$lib/components/CheckboxGroup.svelte';
	import ContentHeader from '$lib/components/ContentHeader.svelte';
	import type { ValidationIssue } from '../editor-state';
	import { blankSelector, removeAt } from './definition-editor';
	import InlineValidationMessages from './InlineValidationMessages.svelte';
	import SelectorEditor from './SelectorEditor.svelte';

	let {
		definition = $bindable(),
		issues = []
	}: { definition: RulesetDefinition; issues?: ValidationIssue[] } = $props();
	function addKnowledge() {
		definition.knowledgeRules.push({
			viewer: blankSelector(),
			target: blankSelector(),
			reveal: ['role']
		});
	}
</script>

<ContentHeader density="dense" description="Choose what one group learns about another group.">
	{#snippet title()}<h2>Starting knowledge</h2>{/snippet}
	{#snippet actions()}<Button variant="secondary" onclick={addKnowledge}>Add knowledge rule</Button
		>{/snippet}
</ContentHeader>
<div class="cards">
	{#each definition.knowledgeRules as rule, index (rule)}
		<article class="item-card">
			<ContentHeader density="dense">
				{#snippet title()}<h3>Knowledge rule {index + 1}</h3>{/snippet}
				{#snippet actions()}<button
						class="remove"
						onclick={() => removeAt(definition.knowledgeRules, index)}>Remove</button
					>{/snippet}
			</ContentHeader>
			<InlineValidationMessages {issues} path={`knowledgeRules[${index}]`} />
			<div class="selector-grid">
				<SelectorEditor
					selector={rule.viewer}
					roles={definition.roles}
					teams={definition.teams}
					categories={definition.categories}
					label="Who receives the knowledge?"
					namePrefix={`knowledge-viewer-${index}`}
				/>
				<SelectorEditor
					selector={rule.target}
					roles={definition.roles}
					teams={definition.teams}
					categories={definition.categories}
					label="Who do they learn about?"
					namePrefix={`knowledge-target-${index}`}
				/>
			</div>
			<div class="choice-block">
				<CheckboxGroup
					label="Reveal"
					name={`knowledge-reveal-${index}`}
					bind:values={rule.reveal}
					options={[
						{ value: 'identity', label: 'Player identity' },
						{ value: 'role', label: 'Role' },
						{ value: 'team', label: 'Team' },
						{ value: 'elimination_state', label: 'Elimination state' }
					]}
				/>
			</div>
		</article>
	{:else}
		<p class="empty">
			No special starting knowledge. Add a rule if teammates or special roles should recognize
			anyone.
		</p>
	{/each}
</div>
