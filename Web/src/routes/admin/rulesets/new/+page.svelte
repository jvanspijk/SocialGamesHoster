<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Button from '$lib/components/Button.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { fieldError, toFormError, type FormError } from '$lib/forms/errors';
	import type { RulesetSummary } from '$lib/api/types';

	let form = $state({ name: '', description: '', minPlayers: '3', maxPlayers: '12' });
	let sourceKind = $state<'blank' | 'duplicate' | 'import'>('blank');
	let sourceRulesetId = $state('');
	let sourceRulesets = $state<RulesetSummary[]>([]);
	let importInput = $state<HTMLInputElement>();
	let busy = $state(false);
	let error = $state<FormError | null>(null);

	async function create(event: SubmitEvent) {
		event.preventDefault();
		busy = true;
		error = null;
		try {
			const created =
				sourceKind === 'import'
					? await api<RulesetSummary>('/rulesets/import', {
							method: 'POST',
							body: importInput?.files?.[0]
						})
					: await api<RulesetSummary>('/rulesets', {
							method: 'POST',
							...jsonBody({
								...form,
								minPlayers: Number(form.minPlayers),
								maxPlayers: Number(form.maxPlayers),
								sourceRulesetId: sourceKind === 'duplicate' ? sourceRulesetId : ''
							})
						});
			await goto(resolve(`/admin/rulesets/${created.id}/edit/metadata`));
		} catch (caught) {
			error = toFormError(caught, 'The ruleset could not be created.');
		} finally {
			busy = false;
		}
	}

	onMount(async () => {
		try {
			sourceRulesets = await api<RulesetSummary[]>('/rulesets');
		} catch {
			// Creating a blank ruleset remains available if the source list cannot load.
		}
	});
</script>

<PageHeading
	eyebrow="Rulesets"
	title="Create ruleset"
	description="Start with the name and player range. You can add the game details next."
/>

<form class="create-form" onsubmit={create}>
	<ErrorNotice message={error?.message} traceId={error?.traceId} />
	<Field
		label="Ruleset name"
		name="ruleset-name"
		bind:value={form.name}
		error={fieldError(error, 'name')}
		required
	/>
	<fieldset>
		<legend>Starting point</legend>
		<label><input type="radio" bind:group={sourceKind} value="blank" /> Blank ruleset</label>
		<label
			><input type="radio" bind:group={sourceKind} value="duplicate" /> Duplicate a ruleset</label
		>
		<label
			><input type="radio" bind:group={sourceKind} value="import" /> Import a ruleset bundle</label
		>
	</fieldset>
	{#if sourceKind === 'duplicate'}
		<label class="source-select"
			><span>Ruleset to duplicate</span><select bind:value={sourceRulesetId} required
				><option value="" disabled>Choose a ruleset</option
				>{#each sourceRulesets as ruleset (ruleset.id)}<option value={ruleset.id}
						>{ruleset.name}</option
					>{/each}</select
			></label
		>
	{:else if sourceKind === 'import'}
		<label class="source-select"
			><span>Ruleset bundle</span><input
				bind:this={importInput}
				type="file"
				accept=".sghrules,application/vnd.socialgameshoster.ruleset+zip"
				required
			/></label
		>
	{/if}
	<Field
		label="Description (optional)"
		name="description"
		bind:value={form.description}
		multiline
	/>
	<div class="limits">
		<label
			><span>Minimum players</span><input
				type="number"
				min="1"
				max="30"
				bind:value={form.minPlayers}
				required
			/></label
		>
		<label
			><span>Maximum players</span><input
				type="number"
				min="1"
				max="30"
				bind:value={form.maxPlayers}
				required
			/></label
		>
	</div>
	<div class="actions">
		<a href={resolve('/admin/rulesets')}>Cancel</a>
		<Button type="submit" loading={busy}>Create ruleset</Button>
	</div>
</form>

<style>
	.create-form {
		display: grid;
		gap: var(--space-4);
		max-width: 42rem;
	}
	fieldset {
		display: grid;
		gap: var(--space-2);
		border: var(--border-subtle);
		padding: var(--space-3);
	}
	fieldset label,
	.source-select {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
	.source-select {
		display: grid;
	}
	.source-select span {
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}
	.source-select select,
	.source-select input {
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		color: var(--ink);
		padding: var(--space-2);
	}
	.limits {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: var(--space-3);
	}
	.limits label {
		display: grid;
		gap: var(--space-1);
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}
	.limits input {
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		color: var(--ink);
		padding: var(--space-2);
	}
	.actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-3);
	}
	.actions a {
		color: var(--crimson-dark);
	}
	@media (max-width: 30rem) {
		.limits {
			grid-template-columns: 1fr;
		}
	}
</style>
