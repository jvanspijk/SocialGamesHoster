<script lang="ts">
	import type { RulesetCategory, RulesetRole, RulesetSelector, RulesetTeam } from '$lib/api/types';

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
		<div>
			<strong>Teams</strong>
			<div class="choices">
				{#each teams as team (team.id)}
					<label
						><input type="checkbox" value={team.id} bind:group={selector.teamIds} />
						{team.name}</label
					>
				{/each}
			</div>
		</div>
	{/if}
	{#if categories.length}
		<div>
			<strong>Categories</strong>
			<div class="choices">
				{#each categories as category (category.id)}
					<label
						><input type="checkbox" value={category.id} bind:group={selector.categoryIds} />
						{category.name}</label
					>
				{/each}
			</div>
		</div>
	{/if}
	{#if roles.length}
		<div>
			<strong>Specific roles</strong>
			<div class="choices">
				{#each roles as role (role.id)}
					<label
						><input type="checkbox" value={role.id} bind:group={selector.roleIds} />
						{role.name}</label
					>
				{/each}
			</div>
		</div>
	{/if}
	<label class="tags">
		<span>Tags (comma-separated)</span>
		<input
			value={selector.tags.join(', ')}
			onchange={(event) => splitTags(event.currentTarget.value)}
			placeholder="investigative, unique"
		/>
	</label>
</fieldset>

<style>
	.selector {
		display: grid;
		gap: 0.65rem;
		min-width: 0;
		border: 1px solid #b89b6d;
		padding: 0.75rem;
	}

	legend,
	strong,
	.tags span {
		font-family: var(--font-display);
		font-size: 0.67rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	p {
		margin: 0;
	}

	.choices {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem 0.8rem;
		margin-top: 0.3rem;
	}

	.choices label {
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
	}

	.tags {
		display: grid;
		gap: 0.25rem;
	}

	.tags input {
		min-height: 44px;
		border: 1px solid #8d7248;
		background: var(--paper-light);
		padding: 0.55rem;
	}
</style>
