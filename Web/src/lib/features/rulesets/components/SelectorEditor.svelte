<script lang="ts">
	import type { RulesetCategory, RulesetRole, RulesetSelector, RulesetTeam } from '$lib/api/types';
	import CheckboxGroup from '$lib/components/CheckboxGroup.svelte';
	import Field from '$lib/components/Field.svelte';

	let {
		selector,
		roles,
		teams,
		categories,
		label = 'Who matches?'
	}: {
		selector: RulesetSelector;
		roles: RulesetRole[];
		teams: RulesetTeam[];
		categories: RulesetCategory[];
		label?: string;
	} = $props();

	function splitTags(value: string) {
		selector.tags = value
			.split(',')
			.map((tag) => tag.trim())
			.filter(Boolean);
	}
</script>

<fieldset class="selector">
	<legend>{label}</legend>
	<p class="muted">Leave every choice empty to include all roles.</p>
	{#if teams.length}
		<CheckboxGroup
			label="Teams"
			name="selector-teams"
			bind:values={selector.teamIds}
			options={teams.map((team) => ({ value: team.id, label: team.name }))}
		/>
	{/if}
	{#if categories.length}
		<CheckboxGroup
			label="Categories"
			name="selector-categories"
			bind:values={selector.categoryIds}
			options={categories.map((category) => ({ value: category.id, label: category.name }))}
		/>
	{/if}
	{#if roles.length}
		<CheckboxGroup
			label="Specific roles"
			name="selector-roles"
			bind:values={selector.roleIds}
			options={roles.map((role) => ({ value: role.id, label: role.name }))}
		/>
	{/if}
	<Field
		label="Tags (comma-separated)"
		name="selector-tags"
		value={selector.tags.join(', ')}
		onchange={splitTags}
		placeholder="investigative, unique"
	/>
</fieldset>

<style>
	.selector {
		display: grid;
		gap: 0.65rem;
		min-width: 0;
		border: 1px solid #b89b6d;
		padding: 0.75rem;
	}

	legend {
		font-family: var(--font-display);
		font-size: 0.67rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	p {
		margin: 0;
	}
</style>
