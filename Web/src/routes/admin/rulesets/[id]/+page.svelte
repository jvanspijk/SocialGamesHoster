<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { ArrowLeft, MoreHorizontal, Save, Trash2 } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import VisualDefinitionEditor from '$lib/features/rulesets/components/VisualDefinitionEditor.svelte';
	import { api, download, jsonBody } from '$lib/api/client';
	import { toFormError, type FormError } from '$lib/forms/errors';
	import type { RulesetDefinition, RulesetSummary } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	type Section =
		| 'metadata'
		| 'teams'
		| 'roles'
		| 'phases'
		| 'composition'
		| 'knowledge'
		| 'chat'
		| 'achievements'
		| 'audio';
	type Report = {
		errors: Array<{ path: string; message: string }>;
		warnings: Array<{ path: string; message: string }>;
	};
	type Detail = { ruleset: RulesetSummary; definition: RulesetDefinition; validation: Report };

	const blank: RulesetDefinition = {
		schemaVersion: 1,
		metadata: { name: '', description: '', minPlayers: 3, maxPlayers: 12 },
		teams: [],
		categories: [],
		abilities: [],
		roles: [],
		phases: [],
		knowledgeRules: [],
		compositionBands: [],
		compositionModifiers: [],
		chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
		achievements: [],
		audioCues: [],
		assetAccessibility: {}
	};
	const sections: Array<{ id: Section; label: string }> = [
		{ id: 'metadata', label: 'Basics' },
		{ id: 'teams', label: 'Teams' },
		{ id: 'roles', label: 'Roles and abilities' },
		{ id: 'composition', label: 'Player setup' },
		{ id: 'phases', label: 'Game flow' },
		{ id: 'knowledge', label: 'Information rules' },
		{ id: 'chat', label: 'Chat' },
		{ id: 'achievements', label: 'Rewards' },
		{ id: 'audio', label: 'Media and audio' }
	];

	let section = $state<Section>('metadata');
	let ruleset = $state<RulesetSummary | null>(null);
	let definition = $state<RulesetDefinition>(structuredClone(blank));
	let report = $state<Report>({ errors: [], warnings: [] });
	let savedDefinition = $state('');
	let dirty = $state(false);
	let saving = $state(false);
	let actionsOpen = $state(false);
	let deleteOpen = $state(false);
	let error = $state<FormError | null>(null);

	onMount(() => {
		if (!auth.isGameMaster) {
			void goto(resolve('/admin'));
			return;
		}
		const requested = page.params.section as Section | undefined;
		if (requested && sections.some((item) => item.id === requested)) section = requested;
		void load();
		const warnBeforeLeaving = (event: BeforeUnloadEvent) => {
			if (dirty) event.preventDefault();
		};
		window.addEventListener('beforeunload', warnBeforeLeaving);
		return () => window.removeEventListener('beforeunload', warnBeforeLeaving);
	});

	$effect(() => {
		dirty = savedDefinition !== '' && JSON.stringify(definition) !== savedDefinition;
	});

	async function load() {
		try {
			const detail = await api<Detail>(`/rulesets/${page.params.id}`);
			ruleset = detail.ruleset;
			definition = detail.definition;
			report = detail.validation;
			savedDefinition = JSON.stringify(detail.definition);
		} catch (caught) {
			error = toFormError(caught, 'The ruleset could not be loaded.');
		}
	}

	function selectSection(next: Section) {
		section = next;
		void goto(resolve(`/admin/rulesets/${page.params.id}/edit/${next}`), {
			replaceState: true,
			keepFocus: true,
			noScroll: true
		});
	}

	async function save() {
		saving = true;
		error = null;
		try {
			const saved = await api<{ validation: Report; availability: 'ready' | 'invalid' }>(
				`/rulesets/${page.params.id}/save`,
				{ method: 'POST', ...jsonBody({ definition }) }
			);
			report = saved.validation;
			toasts.success(
				saved.availability === 'ready' ? 'Ruleset saved as Valid.' : 'Ruleset saved as Invalid.'
			);
			await goto(resolve('/admin/rulesets'));
		} catch (caught) {
			error = toFormError(caught, 'Save failed. Try again.');
		} finally {
			saving = false;
		}
	}

	async function remove() {
		if (!ruleset) return;
		try {
			await api(`/rulesets/${ruleset.id}`, { method: 'DELETE' });
			await goto(resolve('/admin/rulesets'));
		} catch (caught) {
			error = toFormError(caught, 'The ruleset could not be deleted.');
		}
	}
	async function exportRuleset() {
		await download(`/rulesets/${page.params.id}/export`, `${ruleset?.name || 'ruleset'}.sghrules`);
	}
</script>

<div class="editor stack">
	<header>
		<a href={resolve('/admin/rulesets')}><ArrowLeft size={18} /> Rulesets</a>
		<div>
			<p class="ornament">
				{dirty
					? 'Unsaved changes'
					: ruleset?.status === 'valid'
						? 'Saved · Valid'
						: 'Saved · Invalid'}
			</p>
			<h1>{definition.metadata.name || 'Ruleset'}</h1>
		</div>
		<div class="actions">
			<Button loading={saving} onclick={save}><Save size={17} /> Save</Button><Button
				variant="ghost"
				onclick={() => (actionsOpen = true)}><MoreHorizontal size={20} /> Actions</Button
			>
		</div>
	</header>
	<ErrorNotice message={error?.message} traceId={error?.traceId} />
	{#if report.errors.length}<section class="report" aria-labelledby="issues">
			<h2 id="issues">Needs attention</h2>
			{#each report.errors as issue (`${issue.path}:${issue.message}`)}<p>{issue.message}</p>{/each}
		</section>{/if}
	<div class="workspace">
		<nav aria-label="Ruleset sections">
			{#each sections as item (item.id)}<button
					class:active={section === item.id}
					onclick={() => selectSection(item.id)}>{item.label}</button
				>{/each}
		</nav>
		<section class="panel stack">
			{#if section === 'metadata'}
				<h2>Basics</h2>
				<Field label="Name" name="name" bind:value={definition.metadata.name} required /><Field
					label="Description"
					name="description"
					bind:value={definition.metadata.description}
					multiline
				/>
				<div class="limits">
					<label
						><span>Minimum players</span><input
							type="number"
							min="1"
							max="30"
							bind:value={definition.metadata.minPlayers}
							required
						/></label
					><label
						><span>Maximum players</span><input
							type="number"
							min="1"
							max="30"
							bind:value={definition.metadata.maxPlayers}
							required
						/></label
					>
				</div>
			{:else}
				<VisualDefinitionEditor bind:definition {section} assets={[]} />
			{/if}
		</section>
	</div>
</div>

<Dialog
	open={actionsOpen}
	title="Ruleset actions"
	description="Utilities and deletion."
	close={() => (actionsOpen = false)}
	><Button variant="secondary" onclick={exportRuleset}>Export ruleset</Button><Button
		variant="danger"
		onclick={() => {
			actionsOpen = false;
			deleteOpen = true;
		}}><Trash2 size={17} /> Delete ruleset</Button
	>{#snippet actions()}<Button variant="ghost" onclick={() => (actionsOpen = false)}>Close</Button
		>{/snippet}</Dialog
>
<Dialog
	open={deleteOpen}
	title="Delete ruleset?"
	description="This removes the ruleset from the library and from new-game selection."
	close={() => (deleteOpen = false)}
	><p>Games already using this ruleset keep their saved copy.</p>
	{#snippet actions()}<Button variant="ghost" onclick={() => (deleteOpen = false)}>Cancel</Button
		><Button variant="danger" onclick={remove}>Delete ruleset</Button>{/snippet}</Dialog
>

<style>
	.editor {
		max-width: 100rem;
	}
	header {
		display: grid;
		grid-template-columns: minmax(10rem, 1fr) minmax(0, 2fr) auto;
		align-items: center;
		gap: var(--space-4);
	}
	header a {
		display: inline-flex;
		gap: var(--space-1);
		color: var(--crimson-dark);
		text-decoration: none;
	}
	header h1,
	.ornament {
		margin: 0;
	}
	.ornament {
		color: var(--ink-soft);
		font-size: 0.8rem;
	}
	.actions {
		display: flex;
		gap: var(--space-2);
	}
	.report {
		border-inline-start: 0.25rem solid var(--danger);
		background: color-mix(in srgb, var(--danger) 8%, var(--paper-light));
		padding: var(--space-3);
	}
	.report h2,
	.report p {
		margin: 0;
	}
	.report p + p {
		margin-top: var(--space-2);
	}
	.workspace {
		display: grid;
		grid-template-columns: 15rem minmax(0, 1fr);
		gap: var(--space-5);
	}
	nav {
		display: grid;
		align-content: start;
		gap: var(--space-1);
	}
	nav button {
		min-height: var(--target-size);
		border: 0;
		border-inline-start: 3px solid transparent;
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		font: inherit;
		text-align: start;
		padding-inline: var(--space-3);
	}
	nav button.active {
		border-color: var(--crimson);
		background: rgb(255 249 230 / 70%);
		font-weight: 700;
	}
	.panel {
		min-width: 0;
		border: var(--border-subtle);
		background: rgb(255 249 230 / 62%);
		padding: var(--space-4);
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
	@media (max-width: 63.99rem) {
		header {
			grid-template-columns: 1fr;
		}
		.workspace {
			grid-template-columns: 1fr;
		}
		nav {
			grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
		}
	}
	@media (max-width: 30rem) {
		.limits {
			grid-template-columns: 1fr;
		}
		.actions,
		.actions :global(button) {
			width: 100%;
		}
		.actions {
			display: grid;
		}
	}
</style>
