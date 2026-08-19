<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { beforeNavigate, goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { ArrowLeft, ListChecks, Menu, MoreHorizontal, Save, Trash2 } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import Sheet from '$lib/components/Sheet.svelte';
	import VisualDefinitionEditor from '$lib/features/rulesets/components/VisualDefinitionEditor.svelte';
	import {
		copyDefinition,
		humanIssueLocation,
		isEditorSection,
		issueControlName,
		itemNameForIssue,
		nextRequiredSection,
		normalizeReport,
		normalizedDefinition,
		parseRecovery,
		recoveryKey,
		sectionForPath,
		sectionStates,
		serializeRecovery,
		type EditorSection,
		type ValidationIssue,
		type ValidationReport
	} from '$lib/features/rulesets/editor-state';
	import { api, download, jsonBody } from '$lib/api/client';
	import { toFormError, type FormError } from '$lib/forms/errors';
	import type { RulesetDefinition, RulesetSummary } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	type Detail = {
		ruleset: RulesetSummary;
		definition: RulesetDefinition;
		validation: ValidationReport;
	};
	type SectionDefinition = { id: EditorSection; label: string; optional: boolean };
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
	const sections: SectionDefinition[] = [
		{ id: 'metadata', label: 'Basics', optional: false },
		{ id: 'teams', label: 'Teams', optional: false },
		{ id: 'roles', label: 'Roles and abilities', optional: false },
		{ id: 'composition', label: 'Player setup', optional: false },
		{ id: 'phases', label: 'Game flow', optional: true },
		{ id: 'knowledge', label: 'Information rules', optional: true },
		{ id: 'chat', label: 'Chat', optional: true },
		{ id: 'achievements', label: 'Rewards', optional: true },
		{ id: 'audio', label: 'Media', optional: true }
	];
	const labels = Object.fromEntries(sections.map((item) => [item.id, item.label])) as Record<
		EditorSection,
		string
	>;

	let section = $state<EditorSection>('metadata');
	let ruleset = $state<RulesetSummary | null>(null);
	let definition = $state<RulesetDefinition>(structuredClone(blank));
	let savedDefinition = $state<RulesetDefinition>(structuredClone(blank));
	let report = $state<ValidationReport>({ errors: [], warnings: [] });
	let selectedItems = $state<Record<string, string>>({});
	let loaded = $state(false);
	let recovered = $state(false);
	let saving = $state(false);
	let saveFailed = $state(false);
	let validating = $state(false);
	let actionsOpen = $state(false);
	let deleteOpen = $state(false);
	let leaveOpen = $state(false);
	let sectionMenuOpen = $state(false);
	let readinessOpen = $state(false);
	let pendingPath = $state('');
	let bypassNavigation = false;
	let error = $state<FormError | null>(null);
	let announcement = $state('');
	let validationSequence = 0;
	const dirty = $derived.by(
		() => loaded && normalizedDefinition(definition) !== normalizedDefinition(savedDefinition)
	);
	const states = $derived.by(() => sectionStates(definition, report));
	const issueCounts = $derived(
		Object.fromEntries(
			sections.map((item) => [
				item.id,
				report.errors.filter((issue) => sectionForPath(issue.path) === item.id).length
			])
		) as Record<EditorSection, number>
	);
	const sectionIssues = $derived(
		report.errors.filter((issue) => sectionForPath(issue.path) === section)
	);
	const status = $derived(
		saving
			? 'Saving…'
			: saveFailed
				? 'Save failed — retry'
				: recovered && dirty
					? 'Recovered changes'
					: dirty
						? 'Unsaved changes'
						: ruleset?.status === 'valid'
							? 'Saved · Valid'
							: 'Saved · Invalid'
	);

	onMount(() => {
		if (!auth.isGameMaster) {
			void goto(resolve('/admin'));
			return;
		}
		void load();
		const warn = (event: BeforeUnloadEvent) => {
			if (dirty) event.preventDefault();
		};
		window.addEventListener('beforeunload', warn);
		return () => window.removeEventListener('beforeunload', warn);
	});

	beforeNavigate((navigation) => {
		if (!dirty || bypassNavigation || !navigation.to) return;
		const pathname = navigation.to.url.pathname;
		const editorRoot = resolve(`/admin/rulesets/${page.params.id}`);
		if (pathname === editorRoot || pathname.startsWith(`${editorRoot}/edit/`)) return;
		navigation.cancel();
		pendingPath = `${pathname}${navigation.to.url.search}${navigation.to.url.hash}`;
		leaveOpen = true;
	});

	$effect(() => {
		if (!loaded) return;
		const snapshot = normalizedDefinition(definition);
		const activeSection = section;
		const itemState = JSON.stringify(selectedItems);
		const timer = setTimeout(() => {
			if (dirty)
				localStorage.setItem(
					recoveryKey(page.params.id ?? ''),
					serializeRecovery({
						definition: copyDefinition(definition),
						section: activeSection,
						selectedItems: JSON.parse(itemState)
					})
				);
			else localStorage.removeItem(recoveryKey(page.params.id ?? ''));
		}, 500);
		const validationTimer = setTimeout(() => void validateWorkingCopy(snapshot), 350);
		return () => {
			clearTimeout(timer);
			clearTimeout(validationTimer);
		};
	});

	async function load() {
		try {
			const detail = await api<Detail>(`/rulesets/${page.params.id}`);
			ruleset = detail.ruleset;
			savedDefinition = copyDefinition(detail.definition);
			definition = copyDefinition(detail.definition);
			report = normalizeReport(detail.validation);
			const restored = parseRecovery(localStorage.getItem(recoveryKey(page.params.id ?? '')));
			const requested = page.params.section;
			if (
				restored &&
				normalizedDefinition(restored.definition) !== normalizedDefinition(detail.definition)
			) {
				definition = copyDefinition(restored.definition);
				selectedItems = { ...restored.selectedItems };
				section = restored.section;
				recovered = true;
			} else if (requested && isEditorSection(requested)) section = requested;
			else section = nextRequiredSection(definition, report);
			loaded = true;
		} catch (caught) {
			error = toFormError(caught, 'The ruleset could not be loaded.');
		}
	}

	async function validateWorkingCopy(snapshot: string) {
		const sequence = ++validationSequence;
		validating = true;
		try {
			const next = await api<ValidationReport>(`/rulesets/${page.params.id}/validate`, {
				method: 'POST',
				...jsonBody({ definition: JSON.parse(snapshot) })
			});
			if (sequence === validationSequence) report = normalizeReport(next);
		} catch {
			/* Save remains authoritative; retain the last useful report. */
		} finally {
			if (sequence === validationSequence) validating = false;
		}
	}

	function optionalSummary(id: EditorSection) {
		if (id === 'phases')
			return definition.phases.length ? `${definition.phases.length} phases` : 'Not configured';
		if (id === 'knowledge')
			return definition.knowledgeRules.length
				? `${definition.knowledgeRules.length} information rules`
				: 'Not configured';
		if (id === 'chat')
			return definition.chat.channels.length
				? `${definition.chat.channels.length} chat channels`
				: 'Not configured';
		if (id === 'achievements')
			return definition.achievements.length
				? `${definition.achievements.length} achievements`
				: 'Not configured';
		return definition.audioCues.length
			? `${definition.audioCues.length} audio cues`
			: 'Not configured';
	}

	function selectSection(next: EditorSection, itemId?: string) {
		section = next;
		if (itemId) {
			const key: Partial<Record<EditorSection, string>> = {
				teams: 'teams',
				roles: 'roles',
				phases: 'phases',
				composition: itemId.startsWith('modifier') ? 'compositionModifiers' : 'compositionBands',
				chat: 'channels',
				achievements: 'achievements',
				audio: 'audioCues'
			};
			if (key[next]) selectedItems[key[next]!] = itemId;
		}
		sectionMenuOpen = false;
		readinessOpen = false;
		void goto(resolve(`/admin/rulesets/${page.params.id}/edit/${next}`), {
			replaceState: true,
			keepFocus: true,
			noScroll: true
		});
	}

	async function goToIssue(issue: ValidationIssue) {
		const next = sectionForPath(issue.path);
		const item = itemNameForIssue(definition, issue);
		const index = Number(/\[(\d+)\]/.exec(issue.path)?.[1] ?? -1);
		const collections: Partial<Record<EditorSection, Array<{ id: string }>>> = {
			teams: issue.path.startsWith('categories') ? definition.categories : definition.teams,
			roles: issue.path.startsWith('abilities') ? definition.abilities : definition.roles,
			phases: definition.phases,
			composition: issue.path.startsWith('compositionModifiers')
				? definition.compositionModifiers
				: definition.compositionBands,
			chat: definition.chat.channels,
			achievements: definition.achievements,
			audio: definition.audioCues
		};
		selectSection(next, index >= 0 ? collections[next]?.[index]?.id : undefined);
		await tick();
		await tick();
		const control = issueControlName(issue.path);
		if (control)
			document
				.querySelector<HTMLElement>(
					`[name="${CSS.escape(control)}"], #field-${CSS.escape(control)}`
				)
				?.focus({ preventScroll: false });
		announcement = `${humanIssueLocation(definition, issue, labels)}. ${issue.message}${item ? '' : ''}`;
	}

	async function save(destination: Parameters<typeof goto>[0] = resolve('/admin/rulesets')) {
		saving = true;
		saveFailed = false;
		error = null;
		try {
			const saved = await api<{ validation: ValidationReport; availability: 'ready' | 'invalid' }>(
				`/rulesets/${page.params.id}/save`,
				{ method: 'POST', ...jsonBody({ definition }) }
			);
			report = normalizeReport(saved.validation);
			savedDefinition = copyDefinition(definition);
			recovered = false;
			localStorage.removeItem(recoveryKey(page.params.id ?? ''));
			toasts.success(
				saved.availability === 'ready' ? 'Ruleset saved as Valid.' : 'Ruleset saved as Invalid.'
			);
			bypassNavigation = true;
			await goto(resolve(destination as '/admin/rulesets'));
		} catch (caught) {
			saveFailed = true;
			error = toFormError(caught, 'Save failed. Try again.');
			leaveOpen = false;
		} finally {
			saving = false;
		}
	}

	async function discardAndLeave() {
		localStorage.removeItem(recoveryKey(page.params.id ?? ''));
		bypassNavigation = true;
		await goto(resolve((pendingPath || '/admin/rulesets') as '/admin/rulesets'));
	}
	async function remove() {
		if (!ruleset) return;
		try {
			await api(`/rulesets/${ruleset.id}`, { method: 'DELETE' });
			localStorage.removeItem(recoveryKey(ruleset.id));
			bypassNavigation = true;
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
			<p class="status" aria-live="polite">{status}</p>
			<h1>{definition.metadata.name || 'Ruleset'}</h1>
		</div>
		<div class="actions">
			<Button loading={saving} onclick={() => save()}><Save size={17} /> Save</Button><Button
				variant="ghost"
				onclick={() => (actionsOpen = true)}><MoreHorizontal size={20} /> Actions</Button
			>
		</div>
	</header>
	<ErrorNotice message={error?.message} traceId={error?.traceId} />
	<div class="mobile-tools">
		<Button variant="secondary" onclick={() => (sectionMenuOpen = true)}
			><Menu size={18} /> Sections</Button
		><Button variant="secondary" onclick={() => (readinessOpen = true)}
			><ListChecks size={18} /> Readiness ({report.errors.length})</Button
		>
	</div>
	<div class="workspace">
		<nav class="section-rail" aria-label="Ruleset sections">
			<h2>Required foundation</h2>
			{#each sections.filter((item) => !item.optional) as item (item.id)}<button
					class:active={section === item.id}
					onclick={() => selectSection(item.id)}
					><span>{item.label}</span><small
						>{states[item.id]}{#if issueCounts[item.id]}
							· {issueCounts[item.id]} issues{/if}</small
					></button
				>{/each}
			<h2>Optional features</h2>
			{#each sections.filter((item) => item.optional) as item (item.id)}<button
					class:active={section === item.id}
					onclick={() => selectSection(item.id)}
					><span>{item.label}</span><small
						>{states[item.id] === 'Needs attention'
							? `${issueCounts[item.id]} issues`
							: optionalSummary(item.id)}</small
					></button
				>{/each}
		</nav>
		<main class="panel stack">
			{#if !loaded}<p role="status">Loading ruleset…</p>
			{:else if sectionIssues.length}<section
					class="inline-issues"
					aria-labelledby="section-issues"
				>
					<h2 id="section-issues">Needs attention</h2>
					{#each sectionIssues as issue (`${issue.path}:${issue.message}`)}<button
							onclick={() => goToIssue(issue)}>{issue.message}</button
						>{/each}
				</section>{/if}
			{#if loaded && section === 'metadata'}<h2>Basics</h2>
				<Field
					label="Name"
					name="name"
					bind:value={definition.metadata.name}
					required
					error={report.errors.find((issue) => issue.path === 'metadata.name')?.message}
				/><Field
					label="Description"
					name="description"
					bind:value={definition.metadata.description}
					multiline
				/>
				<div class="limits">
					<label
						><span>Minimum players</span><input
							name="minimum-players"
							type="number"
							min="1"
							max="30"
							bind:value={definition.metadata.minPlayers}
							required
						/></label
					><label
						><span>Maximum players</span><input
							name="maximum-players"
							type="number"
							min="1"
							max="30"
							bind:value={definition.metadata.maxPlayers}
							required
						/></label
					>
				</div>{:else if loaded}<VisualDefinitionEditor
					bind:definition
					section={section === 'metadata' ? 'teams' : section}
					assets={[]}
					{selectedItems}
					onnavigate={selectSection}
				/>{/if}
		</main>
		<aside class="readiness">{@render readiness()}</aside>
	</div>
</div>

{#snippet readiness()}<div class="readiness-content">
		<p class="eyebrow">Readiness</p>
		<h2>{report.errors.length ? `${report.errors.length} blocking issues` : 'Ready to save'}</h2>
		{#if validating}<small>Checking changes…</small>{/if}
		<dl>
			<div>
				<dt>Players</dt>
				<dd>{definition.metadata.minPlayers}–{definition.metadata.maxPlayers}</dd>
			</div>
			<div>
				<dt>Teams</dt>
				<dd>{definition.teams.length}</dd>
			</div>
			<div>
				<dt>Roles</dt>
				<dd>{definition.roles.length}</dd>
			</div>
			<div>
				<dt>Phases</dt>
				<dd>{definition.phases.length || 'Optional'}</dd>
			</div>
			<div>
				<dt>Player setup</dt>
				<dd>{definition.compositionBands.length} bands</dd>
			</div>
		</dl>
		{#if report.errors.length}<h3>Fix next</h3>
			<Button onclick={() => goToIssue(report.errors[0])}
				>Fix {humanIssueLocation(definition, report.errors[0], labels)}</Button
			>
			<ul>
				{#each report.errors as issue (`${issue.path}:${issue.message}`)}<li>
						<span
							><strong>{humanIssueLocation(definition, issue, labels)}</strong>{issue.message}</span
						><button onclick={() => goToIssue(issue)}>Go to issue</button>
					</li>{/each}
			</ul>{:else if definition.roles.length === 0}<Button onclick={() => selectSection('roles')}
				>Add the first role</Button
			>{:else if definition.compositionBands.length === 0}<Button
				onclick={() => selectSection('composition')}>Add player setup</Button
			>{:else}<p>No blocking issues in the working copy.</p>{/if}{#if report.warnings.length}<h3>
				Warnings
			</h3>
			<ul>
				{#each report.warnings as issue (`${issue.path}:${issue.message}`)}<li>
						<span
							><strong>{humanIssueLocation(definition, issue, labels)}</strong>{issue.message}</span
						><button onclick={() => goToIssue(issue)}>Review</button>
					</li>{/each}
			</ul>{/if}
		{#if definition.phases.length === 0}<h3>Optional recommendation</h3>
			<p>Game flow is not configured. Add phases if the game master follows an ordered sequence.</p>
		{/if}
	</div>{/snippet}

<Sheet open={sectionMenuOpen} title="Ruleset sections" close={() => (sectionMenuOpen = false)}
	><nav class="sheet-nav" aria-label="Ruleset sections">
		{#each sections as item (item.id)}<button
				class:active={section === item.id}
				onclick={() => selectSection(item.id)}
				><span>{item.label}</span><small
					>{item.optional ? optionalSummary(item.id) : states[item.id]}</small
				></button
			>{/each}
	</nav></Sheet
>
<Sheet open={readinessOpen} title="Ruleset readiness" close={() => (readinessOpen = false)}
	>{@render readiness()}</Sheet
>
<Dialog
	open={leaveOpen}
	title="Leave with unsaved changes?"
	description="Choose what happens to this working copy."
	close={() => (leaveOpen = false)}
	><p>Your changes have not been saved to the host.</p>
	{#snippet actions()}<Button variant="ghost" onclick={() => (leaveOpen = false)}
			>Keep editing</Button
		><Button variant="secondary" onclick={discardAndLeave}>Discard and leave</Button><Button
			onclick={() => save(pendingPath as Parameters<typeof goto>[0])}>Save and leave</Button
		>{/snippet}</Dialog
>
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
<p class="sr-only" aria-live="assertive">{announcement}</p>

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
	.status {
		margin: 0;
	}
	.status,
	small {
		color: var(--ink-soft);
		font-size: 0.8rem;
	}
	.actions,
	.mobile-tools {
		display: flex;
		gap: var(--space-2);
	}
	.workspace {
		display: grid;
		grid-template-columns: 14rem minmax(0, 1fr) minmax(16rem, 20rem);
		align-items: start;
		gap: var(--space-4);
	}
	.section-rail,
	.sheet-nav {
		display: grid;
		align-content: start;
		gap: var(--space-1);
	}
	.section-rail h2 {
		margin: var(--space-3) var(--space-2) var(--space-1);
		font-size: 0.72rem;
	}
	.section-rail button,
	.sheet-nav button {
		display: grid;
		min-height: var(--target-size);
		border: 0;
		border-inline-start: 3px solid transparent;
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		padding: var(--space-2) var(--space-3);
		text-align: start;
	}
	.section-rail button.active,
	.sheet-nav button.active {
		border-color: var(--crimson);
		background: rgb(255 249 230 / 70%);
	}
	.section-rail button span,
	.sheet-nav button span {
		font-weight: 700;
	}
	.panel {
		min-width: 0;
		border: var(--border-subtle);
		background: rgb(255 249 230 / 62%);
		padding: var(--space-4);
	}
	.readiness {
		position: sticky;
		top: var(--space-3);
		max-height: calc(100dvh - var(--space-6));
		overflow: auto;
		border: var(--border-subtle);
		background: var(--paper-light);
		padding: var(--space-4);
	}
	.readiness-content {
		display: grid;
		gap: var(--space-3);
	}
	.readiness-content h2,
	.readiness-content h3,
	.readiness-content p {
		margin: 0;
	}
	.eyebrow {
		font-family: var(--font-display);
		font-size: 0.7rem;
		text-transform: uppercase;
	}
	.readiness dl {
		display: grid;
		gap: var(--space-1);
		margin: 0;
	}
	.readiness dl div {
		display: flex;
		justify-content: space-between;
		gap: var(--space-2);
	}
	.readiness dt {
		color: var(--ink-soft);
	}
	.readiness dd {
		margin: 0;
		font-weight: 700;
	}
	.readiness ul {
		display: grid;
		gap: var(--space-2);
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.readiness li {
		display: grid;
		gap: var(--space-1);
		border-block-start: var(--border-subtle);
		padding-top: var(--space-2);
	}
	.readiness li span {
		display: grid;
	}
	.readiness li button,
	.inline-issues button {
		width: fit-content;
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		padding: 0;
		text-decoration: underline;
	}
	.inline-issues {
		display: grid;
		gap: var(--space-2);
		border-inline-start: 0.25rem solid var(--danger);
		background: color-mix(in srgb, var(--danger) 8%, var(--paper-light));
		padding: var(--space-3);
	}
	.inline-issues h2 {
		margin: 0;
		font-size: 1rem;
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
	.mobile-tools {
		display: none;
	}
	@media (max-width: 63.99rem) {
		header {
			grid-template-columns: 1fr;
		}
		.workspace {
			grid-template-columns: 1fr;
		}
		.section-rail,
		.readiness {
			display: none;
		}
		.mobile-tools {
			display: flex;
		}
		.mobile-tools :global(button) {
			flex: 1;
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
		.mobile-tools {
			display: grid;
		}
	}
</style>
