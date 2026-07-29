<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import {
		ArrowLeft,
		Copy,
		FileImage,
		Menu,
		MoreHorizontal,
		Archive,
		Save,
		Trash2
	} from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import ProtectedMedia from '$lib/components/ProtectedMedia.svelte';
	import Sheet from '$lib/components/Sheet.svelte';
	import VisualDefinitionEditor from '$lib/components/rulesets/VisualDefinitionEditor.svelte';
	import { api, download, jsonBody } from '$lib/api/client';
	import { toFormError, type FormError } from '$lib/forms/errors';
	import type { RulesetDefinition, RulesetSummary } from '$lib/api/types';
	import { auth } from '$lib/state/auth.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	type Version = {
		id: string;
		rulesetId: string;
		versionNumber: number;
		state: 'draft' | 'published';
		definition: RulesetDefinition;
		definitionChecksum: string;
	};
	type Detail = { ruleset: RulesetSummary; versions: Version[] };
	type Section =
		| 'metadata'
		| 'roles'
		| 'phases'
		| 'composition'
		| 'knowledge'
		| 'chat'
		| 'achievements'
		| 'assets'
		| 'advanced';
	type Asset = {
		id: string;
		assetKey: string;
		kind: 'image' | 'audio';
		mimeType: string;
		metadata: Record<string, unknown>;
		preview: string;
	};

	const blank: RulesetDefinition = {
		schemaVersion: 1,
		metadata: { name: 'Untitled game', description: '', minPlayers: 3, maxPlayers: 12 },
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

	let section = $state<Section>('metadata');
	let logical = $state<RulesetSummary | null>(null);
	let version = $state<Version | null>(null);
	let definition = $state<RulesetDefinition>(structuredClone(blank));
	let slug = $state('');
	let text = $state('');
	let error = $state<FormError | null>(null);
	let report = $state<{
		errors: Array<{ path: string; message: string }>;
		warnings: Array<{ path: string; message: string }>;
	} | null>(null);
	let busy = $state('');
	let dirty = $state(false);
	let assets = $state<Asset[]>([]);
	let assetKey = $state('');
	let assetKind = $state<'image' | 'audio'>('image');
	let assetFile = $state<File | null>(null);
	let savedDefinition = $state('');
	let trackingChanges = $state(false);
	let advancedOpen = $state(false);
	let sectionMenuOpen = $state(false);
	let actionsOpen = $state(false);

	function normalizeDefinition(value: RulesetDefinition) {
		value.assetAccessibility ??= {};
		value.abilities ??= [];
		for (const ability of value.abilities) {
			ability.activationPhaseIds ??= [];
			ability.canCombineWithOtherAbilities ??= false;
		}
		return value;
	}

	function ensureAssetDescriptions() {
		definition.assetAccessibility ??= {};
		for (const asset of assets) {
			definition.assetAccessibility[asset.assetKey] ??= { description: '' };
		}
	}

	const sections: Array<{ id: Section; label: string }> = [
		{ id: 'metadata', label: 'Basics' },
		{ id: 'roles', label: 'Roles and teams' },
		{ id: 'phases', label: 'Game flow' },
		{ id: 'composition', label: 'Role setup' },
		{ id: 'knowledge', label: 'Information rules' },
		{ id: 'chat', label: 'Chat rules' },
		{ id: 'achievements', label: 'Rewards' },
		{ id: 'assets', label: 'Media' },
		{ id: 'advanced', label: 'Advanced JSON' }
	];

	onMount(() => {
		if (!auth.isGameMaster) {
			void goto(resolve('/admin'));
			return;
		}
		const requestedSection = page.params.section as Section | undefined;
		if (requestedSection && sections.some((item) => item.id === requestedSection)) {
			section = requestedSection;
		}
		if (page.params.id !== 'new' && !page.params.section) {
			void goto(resolve(`/admin/rulesets/${page.params.id}/edit/metadata`), {
				replaceState: true
			});
			return;
		}
		if (page.params.id !== 'new') void load();
		else {
			const stored = localStorage.getItem('sgh.new-ruleset');
			if (stored) {
				try {
					const parsed = JSON.parse(stored) as { slug: string; definition: RulesetDefinition };
					slug = parsed.slug;
					definition = normalizeDefinition(parsed.definition);
				} catch {
					localStorage.removeItem('sgh.new-ruleset');
				}
			}
			syncText();
			savedDefinition = JSON.stringify(definition);
			trackingChanges = true;
		}
		const protectUnsavedChanges = (event: BeforeUnloadEvent) => {
			if (!dirty) return;
			event.preventDefault();
		};
		window.addEventListener('beforeunload', protectUnsavedChanges);
		return () => window.removeEventListener('beforeunload', protectUnsavedChanges);
	});

	$effect(() => {
		const snapshot = JSON.stringify(definition);
		const currentSlug = slug;
		if (!trackingChanges) return;
		if (!logical) {
			localStorage.setItem(
				'sgh.new-ruleset',
				JSON.stringify({ slug: currentSlug, definition: JSON.parse(snapshot) })
			);
			return;
		}
		if (snapshot === savedDefinition) return;
		dirty = true;
		localStorage.setItem(
			`sgh.ruleset-recovery:${logical.id}`,
			JSON.stringify({ definition: JSON.parse(snapshot), savedAt: new Date().toISOString() })
		);
	});

	async function load() {
		try {
			const detail = await api<Detail>(`/rulesets/${page.params.id}`);
			logical = detail.ruleset;
			const selectedVersion =
				detail.versions.find((candidate) => candidate.state === 'draft') ??
				detail.versions[0] ??
				null;
			const loadedDefinition = normalizeDefinition(
				selectedVersion ? structuredClone(selectedVersion.definition) : structuredClone(blank)
			);
			version = selectedVersion;
			definition = loadedDefinition;
			if (version) assets = await api<Asset[]>(`/ruleset-versions/${version.id}/assets`);
			ensureAssetDescriptions();
			savedDefinition = JSON.stringify(loadedDefinition);
			const recovery = localStorage.getItem(`sgh.ruleset-recovery:${detail.ruleset.id}`);
			if (recovery) {
				try {
					definition = normalizeDefinition(
						(JSON.parse(recovery) as { definition: RulesetDefinition }).definition
					);
					ensureAssetDescriptions();
					dirty = JSON.stringify(definition) !== savedDefinition;
					if (dirty) toasts.info('Recovered unsaved ruleset changes from this browser.');
				} catch {
					localStorage.removeItem(`sgh.ruleset-recovery:${detail.ruleset.id}`);
				}
			}
			syncText();
			trackingChanges = true;
		} catch (caught) {
			setError(caught);
		}
	}

	function selectSection(next: Section) {
		section = next;
		sectionMenuOpen = false;
		advancedOpen = false;
		syncText();
		if (logical) {
			void goto(resolve(`/admin/rulesets/${logical.id}/edit/${next}`), {
				replaceState: true,
				keepFocus: true,
				noScroll: true
			});
		}
	}

	function keysForSection(): Array<keyof RulesetDefinition> {
		switch (section) {
			case 'roles':
				return ['teams', 'categories', 'roles', 'abilities'];
			case 'composition':
				return ['compositionBands', 'compositionModifiers'];
			case 'knowledge':
				return ['knowledgeRules'];
			case 'chat':
				return ['chat'];
			case 'achievements':
				return ['achievements'];
			case 'assets':
				return ['audioCues'];
			case 'phases':
				return ['phases'];
			default:
				return [];
		}
	}

	function syncText() {
		if (section === 'advanced') {
			text = JSON.stringify(definition, null, 2);
			return;
		}
		const keys = keysForSection();
		const value: Record<string, unknown> = {};
		for (const key of keys) value[key] = definition[key];
		text = JSON.stringify(value, null, 2);
	}

	function applyText() {
		if (section === 'metadata') return true;
		try {
			const parsed = JSON.parse(text) as Record<string, unknown>;
			if (section === 'advanced') {
				definition = parsed as unknown as RulesetDefinition;
				error = null;
				dirty = true;
				return true;
			}
			for (const key of keysForSection()) {
				if (!(key in parsed)) throw new Error(`Missing "${key}"`);
				(definition as unknown as Record<string, unknown>)[key] = parsed[key];
			}
			error = null;
			dirty = true;
			return true;
		} catch (caught) {
			error = {
				kind: 'validation',
				message: caught instanceof Error ? caught.message : 'This section is not valid JSON.',
				fieldErrors: {}
			};
			return false;
		}
	}

	async function uploadAsset(event: SubmitEvent) {
		event.preventDefault();
		if (!assetFile) return;
		if (!version) {
			await save();
		}
		if (!version) return;
		if (version.state === 'published' && logical) {
			version = await api<Version>(`/rulesets/${logical.id}/draft`, {
				method: 'POST',
				...jsonBody({})
			});
			assets = await api<Asset[]>(`/ruleset-versions/${version.id}/assets`);
			ensureAssetDescriptions();
		}
		const form = new FormData();
		form.append('assetKey', assetKey);
		form.append('kind', assetKind);
		form.append('file', assetFile);
		busy = 'asset';
		try {
			await api(`/ruleset-versions/${version.id}/assets`, { method: 'POST', body: form });
			assets = await api<Asset[]>(`/ruleset-versions/${version.id}/assets`);
			ensureAssetDescriptions();
			assetKey = '';
			assetFile = null;
		} catch (caught) {
			setError(caught);
		} finally {
			busy = '';
		}
	}

	async function removeAsset(asset: Asset) {
		if (!version || !window.confirm(`Delete the draft asset “${asset.assetKey}”?`)) return;
		try {
			delete definition.assetAccessibility![asset.assetKey];
			await save();
			await api(`/ruleset-versions/${version.id}/assets/${asset.id}`, { method: 'DELETE' });
			assets = assets.filter((candidate) => candidate.id !== asset.id);
		} catch (caught) {
			setError(caught);
		}
	}

	async function save() {
		busy = 'save';
		error = null;
		try {
			if (!logical) {
				const created = await api<{ ruleset: RulesetSummary; draft: Version }>('/rulesets', {
					method: 'POST',
					...jsonBody({ slug, definition })
				});
				logical = created.ruleset;
				version = created.draft;
				const saved = await api<{
					version: Version;
					validation: NonNullable<typeof report>;
					availability: 'ready' | 'invalid';
				}>(`/rulesets/${created.ruleset.id}/save`, {
					method: 'POST',
					...jsonBody({ definition })
				});
				version = saved.version;
				report = saved.validation;
				logical.latestPublishedVersion = saved.availability === 'ready' ? saved.version.id : '';
				await goto(resolve(`/admin/rulesets/${created.ruleset.id}/edit/metadata`), {
					replaceState: true
				});
				localStorage.removeItem('sgh.new-ruleset');
			} else {
				const saved = await api<{
					version: Version;
					validation: NonNullable<typeof report>;
					availability: 'ready' | 'invalid';
				}>(`/rulesets/${logical.id}/save`, {
					method: 'POST',
					...jsonBody({ definition })
				});
				version = saved.version;
				report = saved.validation;
				logical.latestPublishedVersion = saved.availability === 'ready' ? saved.version.id : '';
			}
			dirty = false;
			savedDefinition = JSON.stringify(definition);
			if (logical) localStorage.removeItem(`sgh.ruleset-recovery:${logical.id}`);
			toasts.success(
				report?.errors.length ? 'Ruleset saved with validation issues.' : 'Ruleset saved and ready.'
			);
		} catch (caught) {
			setError(caught);
		} finally {
			busy = '';
		}
	}

	async function exportVersion() {
		if (!version) return;
		await download(
			`/ruleset-versions/${version.id}/export`,
			`${logical?.slug ?? 'ruleset'}-v${version.versionNumber}.sghrules`
		);
	}

	async function duplicate() {
		if (!version) return;
		busy = 'duplicate';
		try {
			const created = await api<{ ruleset: RulesetSummary }>(
				`/ruleset-versions/${version.id}/duplicate`,
				{
					method: 'POST',
					...jsonBody({})
				}
			);
			await goto(resolve(`/admin/rulesets/${created.ruleset.id}/edit/metadata`));
		} catch (caught) {
			setError(caught);
		} finally {
			busy = '';
		}
	}

	async function archiveRuleset() {
		if (!logical || !window.confirm(`Archive “${logical.name}” so it is hidden from new games?`))
			return;
		try {
			await api(`/rulesets/${logical.id}/archive`, {
				method: 'POST',
				...jsonBody({})
			});
			await goto(resolve('/admin/rulesets'));
		} catch (caught) {
			setError(caught);
		}
	}

	async function deleteRuleset() {
		if (
			!logical ||
			window.prompt(
				`Permanently delete “${logical.name}” and every unreferenced version?\n\nType DELETE to continue.`
			) !== 'DELETE'
		)
			return;
		try {
			await api(`/rulesets/${logical.id}`, { method: 'DELETE' });
			await goto(resolve('/admin/rulesets'));
		} catch (caught) {
			setError(caught);
		}
	}

	function setError(caught: unknown) {
		const nextError = toFormError(caught, 'The ruleset could not be updated.');
		error = nextError;
		if (nextError.kind !== 'validation') toasts.error(nextError.message);
	}
</script>

<div class="editor stack">
	<header>
		<a href={resolve('/admin/rulesets')}><ArrowLeft size={18} /> Rulesets</a>
		<div>
			<p class="ornament">
				{dirty
					? 'Unsaved changes'
					: report?.errors.length
						? 'Saved · Invalid'
						: logical?.latestPublishedVersion
							? 'Saved · Ready'
							: logical
								? 'Saved · Invalid'
								: 'New ruleset'}
			</p>
			<h1>{definition.metadata.name}</h1>
		</div>
		<div class="actions">
			<Button loading={busy === 'save'} onclick={save}><Save size={17} /> Save</Button>
			<Button variant="ghost" onclick={() => (actionsOpen = true)}>
				<MoreHorizontal size={20} /> Actions
			</Button>
		</div>
	</header>

	<ErrorNotice {error} />
	{#if report}
		<div class:valid={!report.errors.length} class="report card">
			<strong
				>{report.errors.length
					? `${report.errors.length} validation issue(s)`
					: 'Definition is valid'}</strong
			>
			{#each report.errors as issue (`${issue.path}:${issue.message}`)}<p>
					<code>{issue.path}</code> — {issue.message}
				</p>{/each}
			{#each report.warnings as issue (`${issue.path}:${issue.message}`)}<p class="warning">
					<code>{issue.path}</code> — {issue.message}
				</p>{/each}
		</div>
	{/if}

	<div class="workspace">
		<button class="section-menu" type="button" onclick={() => (sectionMenuOpen = true)}>
			<Menu size={18} /> Sections · {sections.find((item) => item.id === section)?.label}
			{#if dirty}<span>Unsaved</span>{/if}
		</button>
		<nav aria-label="Ruleset sections">
			{#each sections as item (item.id)}
				<button class:active={section === item.id} onclick={() => selectSection(item.id)}
					>{item.label}</button
				>
			{/each}
		</nav>
		<section class="card panel stack">
			{#if section === 'metadata'}
				<h2>Overview and limits</h2>
				{#if !logical}<Field
						label="Stable slug"
						name="slug"
						bind:value={slug}
						help="Lowercase letters, numbers, hyphens, and underscores."
						required
					/>{/if}
				<Field label="Name" name="name" bind:value={definition.metadata.name} required />
				<Field
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
						/></label
					>
					<label
						><span>Maximum players</span><input
							type="number"
							min="1"
							max="30"
							bind:value={definition.metadata.maxPlayers}
						/></label
					>
				</div>
			{:else if section === 'assets'}
				<div class="section-heading">
					<div>
						<FileImage size={24} />
						<h2>Images and audio</h2>
					</div>
					<small>Draft assets only</small>
				</div>
				<p class="muted">
					Images may be JPEG, PNG, or WebP up to 2 MB. Audio may be MP3, M4A, Ogg, or WAV up to 5 MB
					and 60 seconds.
				</p>
				<VisualDefinitionEditor
					bind:definition
					section="audio"
					assets={assets.map(({ assetKey: key, kind }) => ({ assetKey: key, kind }))}
				/>
				<form class="asset-form" onsubmit={uploadAsset}>
					<Field
						label="Asset key"
						name="assetKey"
						bind:value={assetKey}
						help="Reference this stable key from roles, achievements, or audio cues."
						required
					/>
					<label>
						<span>Kind</span>
						<select bind:value={assetKind}>
							<option value="image">Image</option>
							<option value="audio">Audio</option>
						</select>
					</label>
					<label>
						<span>File</span>
						<input
							type="file"
							accept={assetKind === 'image'
								? 'image/jpeg,image/png,image/webp'
								: 'audio/mpeg,audio/mp4,audio/ogg,audio/wav'}
							onchange={(event) =>
								(assetFile = (event.currentTarget as HTMLInputElement).files?.[0] ?? null)}
							required
						/>
					</label>
					<Button type="submit" loading={busy === 'asset'} disabled={!assetFile}
						>Upload asset</Button
					>
				</form>
				<div class="asset-list">
					{#each assets as asset (asset.id)}
						<article>
							{#if asset.kind === 'image'}
								<ProtectedMedia src={asset.preview} kind="image" alt="" />
							{:else}
								<ProtectedMedia src={asset.preview} kind="audio" controls />
							{/if}
							<div>
								<strong>{asset.assetKey}</strong>
								<small>{asset.mimeType} · {JSON.stringify(asset.metadata)}</small>
								<Field
									label={asset.kind === 'image' ? 'Image description' : 'Audio alternative'}
									name={`asset-accessibility-${asset.id}`}
									bind:value={definition.assetAccessibility![asset.assetKey].description}
									help="Used when this asset is attached to an announcement."
								/>
							</div>
							<Button variant="danger" onclick={() => removeAsset(asset)}
								><Trash2 size={16} /> Delete</Button
							>
						</article>
					{:else}
						<p>No files in this draft.</p>
					{/each}
				</div>
			{:else if section === 'advanced'}
				<h2>Advanced JSON</h2>
				<p class="muted">
					For advanced authors and imported definitions. Invalid changes can still be saved, but the
					ruleset will not be available for new games.
				</p>
				<textarea bind:value={text} spellcheck="false" aria-label="Complete ruleset JSON"
				></textarea>
				<Button variant="secondary" onclick={applyText}>Apply JSON</Button>
			{:else if section === 'roles'}
				<VisualDefinitionEditor
					bind:definition
					section="teams"
					assets={assets.map(({ assetKey: key, kind }) => ({ assetKey: key, kind }))}
				/>
				<VisualDefinitionEditor
					bind:definition
					section="roles"
					assets={assets.map(({ assetKey: key, kind }) => ({ assetKey: key, kind }))}
				/>
				<details
					class="advanced"
					bind:open={advancedOpen}
					ontoggle={() => {
						if (advancedOpen) syncText();
					}}
				>
					<summary>Section JSON</summary>
					<textarea bind:value={text} spellcheck="false" aria-label={`${section} JSON`}></textarea>
					<Button variant="secondary" onclick={applyText}>Apply JSON to this section</Button>
				</details>
			{:else}
				<VisualDefinitionEditor
					bind:definition
					{section}
					assets={assets.map(({ assetKey: key, kind }) => ({ assetKey: key, kind }))}
				/>
				<details
					class="advanced"
					bind:open={advancedOpen}
					ontoggle={() => {
						if (advancedOpen) syncText();
					}}
				>
					<summary>Section JSON</summary>
					<p class="muted">
						For advanced authors and imported definitions. Changes apply only when you press the
						button.
					</p>
					<textarea bind:value={text} spellcheck="false" aria-label={`${section} JSON`}></textarea>
					<Button variant="secondary" onclick={applyText}>Apply JSON to this section</Button>
				</details>
			{/if}
			{#if dirty}<p class="dirty">Unsaved changes</p>{/if}
		</section>
	</div>
</div>

<Sheet open={sectionMenuOpen} title="Ruleset sections" close={() => (sectionMenuOpen = false)}>
	<div class="section-list">
		{#each sections as item (item.id)}
			<button
				type="button"
				class:active={section === item.id}
				onclick={() => selectSection(item.id)}
			>
				<span>{item.label}</span>
				{#if dirty && section === item.id}<small>Unsaved</small>{/if}
			</button>
		{/each}
	</div>
</Sheet>

<Dialog open={actionsOpen} title="Ruleset actions" close={() => (actionsOpen = false)}>
	<div class="overflow-actions">
		{#if version}<Button variant="secondary" onclick={duplicate}
				><Copy size={17} /> Duplicate ruleset</Button
			>{/if}
		{#if version}<Button variant="secondary" onclick={exportVersion}>Export ruleset</Button>{/if}
		{#if logical && !logical.archived}
			<Button variant="secondary" onclick={archiveRuleset}
				><Archive size={17} /> Archive ruleset</Button
			>
		{/if}
		{#if logical}<Button variant="danger" onclick={deleteRuleset}
				><Trash2 size={17} /> Delete ruleset</Button
			>{/if}
	</div>
</Dialog>

<style>
	.editor {
		max-width: 70rem;
		margin-inline: auto;
	}

	header {
		display: grid;
		grid-template-columns: auto 1fr auto;
		align-items: center;
		gap: 1rem;
	}

	header > a {
		display: inline-flex;
		min-height: 44px;
		align-items: center;
		gap: 0.4rem;
		font-family: var(--font-display);
		font-size: 0.68rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	.actions {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: 0.4rem;
	}

	.overflow-actions {
		display: grid;
		gap: var(--space-2);
	}

	.overflow-actions :global(button) {
		justify-content: flex-start;
		width: 100%;
	}

	.workspace {
		display: grid;
		grid-template-columns: 13rem minmax(0, 1fr);
		gap: 1rem;
	}

	.section-menu {
		display: none;
	}

	.section-list {
		display: grid;
	}

	.section-list button {
		display: flex;
		min-height: var(--target-size);
		align-items: center;
		justify-content: space-between;
		border: 0;
		border-block-end: var(--border-subtle);
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		padding: var(--space-3);
		text-align: start;
	}

	.section-list button.active {
		border-inline-start: 3px solid var(--crimson);
		color: var(--crimson-dark);
	}

	.workspace > nav {
		display: grid;
		align-self: start;
		border: 1px solid #9a7e51;
		background: rgb(255 249 230 / 45%);
		padding: 0.45rem;
	}

	.workspace > nav button {
		min-height: 44px;
		border: 0;
		border-inline-start: 3px solid transparent;
		background: transparent;
		color: var(--ink-soft);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.65rem;
		font-weight: 700;
		letter-spacing: 0.05em;
		padding: 0.5rem;
		text-align: start;
		text-transform: uppercase;
	}

	.workspace > nav button.active {
		border-color: var(--crimson);
		background: color-mix(in srgb, var(--crimson) 8%, transparent);
		color: var(--crimson-dark);
	}

	.panel {
		min-width: 0;
	}

	.section-heading,
	.section-heading > div {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.5rem;
	}

	.section-heading h2 {
		margin: 0;
	}

	.section-heading small {
		color: var(--ink-faint);
	}

	textarea {
		width: 100%;
		min-height: 34rem;
		resize: vertical;
		border: 1px solid #8d7248;
		background: #2b231f;
		color: #f4e7cf;
		font-family: Consolas, 'Courier New', monospace;
		font-size: 0.8rem;
		line-height: 1.5;
		padding: 0.85rem;
		tab-size: 2;
	}

	.limits {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 0.8rem;
	}

	.asset-form {
		display: grid;
		grid-template-columns: minmax(12rem, 1fr) minmax(8rem, 0.35fr);
		gap: 0.75rem;
	}

	.asset-form label {
		display: grid;
		gap: 0.3rem;
	}

	.asset-form span {
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	.asset-form input,
	.asset-form select {
		min-height: 44px;
		border: 1px solid #8d7248;
		background: var(--paper-light);
		padding: 0.6rem;
	}

	.asset-list article {
		display: grid;
		grid-template-columns: 5rem minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.75rem;
		border-top: 1px solid #b89b6d;
		padding-block: 0.7rem;
	}

	.asset-list :global(img) {
		width: 5rem;
		height: 4rem;
		object-fit: cover;
	}

	.asset-list :global(audio) {
		width: 5rem;
	}

	.asset-list div {
		display: grid;
		min-width: 0;
	}

	.asset-list small {
		overflow-wrap: anywhere;
	}

	.limits label {
		display: grid;
		gap: 0.3rem;
	}

	.limits span {
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}

	.limits input {
		min-height: 44px;
		border: 1px solid #8d7248;
		background: var(--paper-light);
		padding: 0.6rem;
	}

	.report {
		border-inline-start: 4px solid var(--danger);
	}

	.report.valid {
		border-inline-start-color: var(--success);
	}

	.report p {
		margin: 0.3rem 0;
	}

	.warning,
	.dirty {
		color: var(--gold);
	}

	.dirty {
		margin: 0;
		font-family: var(--font-display);
		font-size: 0.65rem;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	@media (max-width: 850px) {
		header {
			grid-template-columns: 1fr;
		}

		.actions {
			justify-content: flex-start;
		}

		.workspace {
			grid-template-columns: 1fr;
		}

		.workspace > nav {
			display: none;
		}

		.section-menu {
			display: flex;
			width: 100%;
			min-height: var(--target-size);
			align-items: center;
			justify-content: space-between;
			gap: var(--space-2);
			border: var(--border-strong);
			background: var(--paper-light);
			color: var(--ink);
			padding: var(--space-2) var(--space-3);
		}

		.actions {
			position: sticky;
			z-index: var(--layer-sticky);
			inset-block-end: 0;
			flex-wrap: wrap;
			background: var(--paper);
			padding-block: var(--space-2);
		}
	}
</style>
