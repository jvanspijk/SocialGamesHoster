<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import Button from '$lib/components/Button.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import ContentHeader from '$lib/components/ContentHeader.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import { ArrowLeft, Plus } from '@lucide/svelte';
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

<main class="create-page">
	<header class="topbar">
		<a href={resolve('/admin/rulesets')}><ArrowLeft size={18} /> Rulesets</a>
	</header>
	<Panel variant="focal">
		<form class="create-form" onsubmit={create}>
			<div class="form-heading">
				<ContentHeader
					eyebrow="Rulesets"
					description="Choose a starting point, then add the game details in the editor."
				>
					{#snippet title()}<h1>Create ruleset</h1>{/snippet}
					{#snippet actions()}<Button type="submit" loading={busy}
							><Plus size={17} /> Create ruleset</Button
						>{/snippet}
				</ContentHeader>
			</div>
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
				<!-- Ruleset bundle import is being added in a later release and is not planned for v1.
		<label><input type="radio" bind:group={sourceKind} value="import" /> Import a ruleset bundle</label>
		-->
			</fieldset>
			{#if sourceKind === 'duplicate'}
				<SelectField
					label="Ruleset to duplicate"
					name="source-ruleset"
					bind:value={sourceRulesetId}
					required
					options={[
						{ value: '', label: 'Choose a ruleset', disabled: true },
						...sourceRulesets.map((ruleset) => ({ value: ruleset.id, label: ruleset.name }))
					]}
				/>
			{:else if sourceKind === 'import'}
				<label class="source-select"
					><span>Ruleset file</span><input
						bind:this={importInput}
						type="file"
						accept=".sghrules,application/vnd.socialgameshoster.ruleset+zip"
						required
					/></label
				>
			{/if}
			<Field
				label="Description"
				help="The main idea behind the game."
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
		</form>
	</Panel>
</main>

<style>
	.create-page {
		width: min(100%, 58rem);
		margin-inline: auto;
		padding: var(--space-5) max(var(--space-4), env(safe-area-inset-right)) var(--space-7)
			max(var(--space-4), env(safe-area-inset-left));
	}
	.topbar {
		margin-bottom: var(--space-6);
	}
	.topbar a {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		min-height: var(--target-size);
		color: var(--action-dark);
		text-decoration: none;
	}
	.create-form {
		display: grid;
		gap: var(--space-5);
		padding: var(--space-4);
	}
	.form-heading {
		border-bottom: var(--border-subtle);
		padding-bottom: var(--space-5);
		margin-bottom: var(--space-3);
	}
	h1 {
		margin: 0;
		font-size: 1.6rem;
	}
	fieldset {
		display: grid;
		gap: var(--space-2);
		border: 0;
		padding: 0;
		margin: 0;
	}
	legend {
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		padding: 0 0 var(--space-2);
	}
	fieldset label {
		min-height: var(--target-size);
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
	@media (max-width: 47.99rem) {
		.create-form {
			padding: 0;
		}
	}
	@media (max-width: 30rem) {
		.limits {
			grid-template-columns: 1fr;
		}
	}
</style>
