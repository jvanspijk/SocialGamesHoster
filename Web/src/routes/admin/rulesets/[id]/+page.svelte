<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { beforeNavigate, goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import {
		ArrowLeft,
		Eye,
		ListChecks,
		Menu,
		Save,
		Trash2,
		Check,
		Circle,
		CircleAlert
	} from '@lucide/svelte';
	import Panel from '$lib/components/Panel.svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import ErrorNotice from '$lib/components/ErrorNotice.svelte';
	import Field from '$lib/components/Field.svelte';
	import Sheet from '$lib/components/Sheet.svelte';
	import VisualDefinitionEditor from '$lib/features/rulesets/components/VisualDefinitionEditor.svelte';
	import InlineValidationMessages from '$lib/features/rulesets/components/InlineValidationMessages.svelte';
	import MediaField from '$lib/features/rulesets/components/MediaField.svelte';
	import RulesetPreview from '$lib/features/rulesets/components/RulesetPreview.svelte';
	import { assetUsages } from '$lib/features/rulesets/components/definition-editor';
	import {
		copyDefinition,
		humanIssueLocation,
		isEditorSection,
		issueControlName,
		itemTargetForIssue,
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
	import { api, jsonBody } from '$lib/api/client';
	import { toFormError, type FormError } from '$lib/forms/errors';
	import type {
		RulesetAsset,
		RulesetDefinition,
		RulesetEditSession,
		RulesetPreviewMode,
		RulesetPreviewRequest,
		RulesetPreviewResponse,
		RulesetSummary
	} from '$lib/api/types';
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
		chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
		achievements: [],
		audioCues: [],
		assetAccessibility: {}
	};
	const sections: SectionDefinition[] = [
		{ id: 'metadata', label: 'Basics', optional: false },
		{ id: 'teams', label: 'Teams', optional: false },
		{ id: 'roles', label: 'Roles and abilities', optional: false },
		{ id: 'phases', label: 'Game flow', optional: true },
		{ id: 'knowledge', label: 'Information rules', optional: true },
		{ id: 'chat', label: 'Chat', optional: true },
		{ id: 'achievements', label: 'Rewards', optional: true },
		{ id: 'assets', label: 'Assets', optional: true }
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
	let editSession = $state<RulesetEditSession | null>(null);
	let assets = $state<RulesetAsset[]>([]);
	let mediaDirty = $state(false);
	let reuploadNames = $state<string[]>([]);
	let loaded = $state(false);
	let saving = $state(false);
	let saveFailed = $state(false);
	let validating = $state(false);
	let deleteOpen = $state(false);
	let leaveOpen = $state(false);
	let sectionMenuOpen = $state(false);
	let overviewOpen = $state(false);
	let previewOpen = $state(false);
	let previewResults = $state<Partial<Record<RulesetPreviewMode, RulesetPreviewResponse>>>({});
	let previewSource = '';
	let pendingPath = $state('');
	let bypassNavigation = false;
	let error = $state<FormError | null>(null);
	let announcement = $state('');
	let validationSequence = 0;
	const dirty = $derived.by(
		() =>
			loaded &&
			(normalizedDefinition(definition) !== normalizedDefinition(savedDefinition) || mediaDirty)
	);
	const states = $derived.by(() => {
		const values = sectionStates(definition, report);
		if (assets.length && values.assets === 'Not started') values.assets = 'Complete';
		return values;
	});
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
				: dirty
					? 'Unsaved changes'
					: loaded
						? 'All changes saved'
						: 'Loading ruleset…'
	);
	const requiredComplete = $derived(
		sections.filter((item) => !item.optional && states[item.id] === 'Complete').length
	);
	const optionalConfigured = $derived(
		sections.filter((item) => item.optional && states[item.id] !== 'Not started').length
	);
	const orphanedAssets = $derived(
		assets.filter((asset) => assetUsages(definition, asset.assetKey).length === 0)
	);

	const previewGuidance = $derived.by(() => {
		const guidance: Array<{
			mode: RulesetPreviewMode;
			message: string;
			action?: string;
			section?: EditorSection;
		}> = [];
		const chat = previewResults.chat;
		if (chat?.rooms) {
			const readable = chat.rooms.filter((room) => room.readable).length;
			const sendable = chat.rooms.filter((room) => room.sendable).length;
			guidance.push({
				mode: 'chat',
				message: `${chat.audience} can read ${readable} and post in ${sendable} of ${chat.rooms.length} chat spaces during ${chat.phase}.`,
				action: sendable === 0 ? 'Review chat permissions' : undefined,
				section: sendable === 0 ? 'chat' : undefined
			});
		}
		const media = previewResults.media;
		if (media?.media) {
			const count = media.contexts?.length ?? 0;
			guidance.push({
				mode: 'media',
				message: count
					? `${media.media.displayName} was checked in ${count} ${count === 1 ? 'game context' : 'game contexts'}.`
					: `${media.media.displayName} is not currently used in the ruleset.`,
				action: count === 0 ? 'Review assets' : undefined,
				section: count === 0 ? 'assets' : undefined
			});
		}
		const phases = previewResults.phases;
		if (phases?.empty) {
			guidance.push({
				mode: 'phases',
				message: phases.message ?? 'No game flow is configured.',
				action: 'Add game flow',
				section: 'phases'
			});
		}
		return guidance;
	});

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
		const currentPreviewSource = `${snapshot}:${JSON.stringify(assets.map((asset) => [asset.assetKey, asset.checksum]))}`;
		if (previewSource && previewSource !== currentPreviewSource) {
			previewResults = {};
			previewSource = '';
		}
		const activeSection = section;
		const itemState = JSON.stringify(selectedItems);
		const activeSessionId = editSession?.id;
		const stagedAssetNames = assets
			.filter((asset) => asset.staged)
			.map((asset) => asset.displayName);
		const timer = setTimeout(() => {
			if (dirty)
				localStorage.setItem(
					recoveryKey(page.params.id ?? ''),
					serializeRecovery({
						definition: copyDefinition(definition),
						section: activeSection,
						selectedItems: JSON.parse(itemState),
						sessionId: activeSessionId,
						stagedAssetNames
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
			const opened = await api<RulesetEditSession>(`/rulesets/${page.params.id}/edit-session`, {
				method: 'POST'
			});
			editSession = opened;
			assets = await api<RulesetAsset[]>(
				`/rulesets/${page.params.id}/edit-session/${opened.id}/assets`
			);
			mediaDirty = opened.hasChanges;
			ruleset = detail.ruleset;
			savedDefinition = copyDefinition(detail.definition);
			definition = copyDefinition(detail.definition);
			report = normalizeReport(detail.validation);
			const restored = parseRecovery(localStorage.getItem(recoveryKey(page.params.id ?? '')));
			if (
				restored?.sessionId &&
				restored.sessionId !== opened.id &&
				restored.stagedAssetNames?.length
			) {
				reuploadNames = restored.stagedAssetNames;
			}
			const requested = page.params.section;
			if (
				restored &&
				(normalizedDefinition(restored.definition) !== normalizedDefinition(detail.definition) ||
					(restored.sessionId === opened.id && opened.hasChanges))
			) {
				definition = copyDefinition(restored.definition);
				selectedItems = { ...restored.selectedItems };
				section = restored.section;
			} else if (requested && isEditorSection(requested)) section = requested;
			else section = nextRequiredSection(definition, report);
			if (requested === 'audio' || requested === 'media') {
				section = 'assets';
				void goto(resolve(`/admin/rulesets/${page.params.id}/edit/assets`), { replaceState: true });
			}
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
				...jsonBody({ definition: JSON.parse(snapshot), sessionId: editSession?.id })
			});
			if (sequence === validationSequence) report = normalizeReport(next);
		} catch {
			/* Save remains authoritative; retain the last useful report. */
		} finally {
			if (sequence === validationSequence) validating = false;
		}
	}

	function sectionSummary(item: SectionDefinition) {
		if (issueCounts[item.id])
			return `${issueCounts[item.id]} ${issueCounts[item.id] === 1 ? 'issue' : 'issues'}`;
		if (section === item.id) return 'Editing';
		if (states[item.id] === 'Not started') return item.optional ? 'Not configured' : 'Not started';
		const count = (n: number, label: string) => `${n} ${label}${n === 1 ? '' : 's'}`;
		switch (item.id) {
			case 'metadata':
				return states.metadata;
			case 'teams':
				return count(definition.teams.length, 'team');
			case 'roles':
				return count(definition.roles.length, 'role');
			case 'phases':
				return count(definition.phases.length, 'phase');
			case 'knowledge':
				return count(definition.knowledgeRules.length, 'rule');
			case 'chat':
				return definition.chat.channels.length
					? count(definition.chat.channels.length, 'channel')
					: 'Configured';
			case 'achievements':
				return count(definition.achievements.length, 'reward');
			case 'assets':
				return assets.length ? count(assets.length, 'file') : 'Configured';
		}
	}

	function selectSection(next: EditorSection, itemId?: string, itemKey?: string) {
		section = next;
		if (itemId) {
			const key =
				itemKey ??
				(
					{
						teams: definition.categories.some((item) => item.id === itemId)
							? 'categories'
							: 'teams',
						roles: definition.abilities.some((item) => item.id === itemId) ? 'abilities' : 'roles',
						phases: 'phases',
						chat: 'channels',
						achievements: 'achievements',
						assets: 'audioCues'
					} as Partial<Record<EditorSection, string>>
				)[next];
			if (key) selectedItems[key] = itemId;
		}
		sectionMenuOpen = false;
		overviewOpen = false;
		void goto(resolve(`/admin/rulesets/${page.params.id}/edit/${next}`), {
			replaceState: true,
			keepFocus: true,
			noScroll: true
		});
	}

	async function goToIssue(issue: ValidationIssue) {
		const next = sectionForPath(issue.path);
		const item = itemNameForIssue(definition, issue);
		const target = itemTargetForIssue(definition, issue);
		selectSection(next, target?.id, target?.key);
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
				{ method: 'POST', ...jsonBody({ definition, sessionId: editSession?.id }) }
			);
			report = normalizeReport(saved.validation);
			savedDefinition = copyDefinition(definition);
			localStorage.removeItem(recoveryKey(page.params.id ?? ''));
			toasts.success(
				saved.availability === 'ready'
					? 'Ruleset saved and ready to use.'
					: 'Ruleset saved, but needs attention.'
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
		if (editSession) {
			try {
				await api(`/rulesets/${page.params.id}/edit-session/${editSession.id}`, {
					method: 'DELETE'
				});
			} catch {
				toasts.info(
					'Your unsaved changes were discarded. Some uploaded assets could not be cleaned up and will expire automatically.'
				);
			}
		}
		localStorage.removeItem(recoveryKey(page.params.id ?? ''));
		bypassNavigation = true;
		await goto(resolve((pendingPath || '/admin/rulesets') as '/admin/rulesets'));
	}

	async function refreshAssets() {
		if (!editSession) return;
		assets = await api<RulesetAsset[]>(
			`/rulesets/${page.params.id}/edit-session/${editSession.id}/assets`
		);
	}

	async function uploadMedia(
		file: File,
		kind: 'image' | 'audio',
		displayName: string,
		accessibilityText: string,
		replaceAssetKey?: string
	) {
		if (!editSession) throw new Error('The asset editing session is not ready.');
		const body = new FormData();
		body.set('file', file);
		body.set('kind', kind);
		body.set('displayName', displayName);
		body.set('accessibilityText', accessibilityText);
		body.set('mode', replaceAssetKey ? 'replace' : 'add');
		if (replaceAssetKey) body.set('assetKey', replaceAssetKey);
		const asset = await api<RulesetAsset>(
			`/rulesets/${page.params.id}/edit-session/${editSession.id}/assets`,
			{ method: 'POST', body }
		);
		mediaDirty = true;
		await refreshAssets();
		return assets.find((item) => item.assetKey === asset.assetKey) ?? asset;
	}

	async function updateMedia(assetKey: string, displayName: string, accessibilityText: string) {
		if (!editSession) throw new Error('The asset editing session is not ready.');
		await api(
			`/rulesets/${page.params.id}/edit-session/${editSession.id}/assets/${encodeURIComponent(assetKey)}`,
			{ method: 'PATCH', ...jsonBody({ displayName, accessibilityText }) }
		);
		mediaDirty = true;
		await refreshAssets();
	}

	async function removeMedia(assetKey: string) {
		if (!editSession) throw new Error('The asset editing session is not ready.');
		await api(
			`/rulesets/${page.params.id}/edit-session/${editSession.id}/assets/${encodeURIComponent(assetKey)}`,
			{ method: 'DELETE', ...jsonBody({ definition }) }
		);
		mediaDirty = true;
		await refreshAssets();
	}

	const media = { upload: uploadMedia, update: updateMedia, remove: removeMedia };
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
	async function loadPreview(request: RulesetPreviewRequest) {
		return api<RulesetPreviewResponse>(`/rulesets/${page.params.id}/preview`, {
			method: 'POST',
			...jsonBody({ ...request, definition, sessionId: editSession?.id })
		});
	}
	function recordPreviewResult(response: RulesetPreviewResponse) {
		previewSource = `${normalizedDefinition(definition)}:${JSON.stringify(assets.map((asset) => [asset.assetKey, asset.checksum]))}`;
		previewResults = { ...previewResults, [response.mode]: response };
	}
</script>

<div class="editor stack">
	<header class="editor-topbar">
		<a href={resolve('/admin/rulesets')}><ArrowLeft size={18} /> Rulesets</a>
		<div class="topbar-actions">
			<p class="status" aria-live="polite">{status}</p>
			<Button variant="danger" disabled={!loaded} onclick={() => (deleteOpen = true)}
				><Trash2 size={18} /> Delete ruleset</Button
			>
		</div>
	</header>
	<ErrorNotice message={error?.message} traceId={error?.traceId} />
	{#if reuploadNames.length}
		<section class="reupload-warning" role="status">
			<strong>Some recovered asset files must be uploaded again.</strong>
			<p>{reuploadNames.join(', ')}</p>
		</section>
	{/if}
	<div class="mobile-tools">
		<Button variant="secondary" onclick={() => (sectionMenuOpen = true)}
			><Menu size={18} /> Sections</Button
		><Button variant="secondary" onclick={() => (overviewOpen = true)}
			><ListChecks size={18} /> Overview ({report.errors.length})</Button
		>
	</div>
	<div class="workspace">
		<div class="section-rail">{@render sectionNavigation()}</div>
		<section class="editor-section" aria-label="Ruleset editor section">
			<Panel variant="focal">
				<div class="editor-body">
					<div class="section-heading">
						<div>
							<p class="eyebrow">{definition.metadata.name || 'Ruleset'}</p>
							<h1>{labels[section]}</h1>
						</div>
						<div class="actions">
							<Button variant="secondary" disabled={!loaded} onclick={() => (previewOpen = true)}
								><Eye size={17} /> Preview</Button
							>
							<Button loading={saving} disabled={!loaded} onclick={() => save()}
								><Save size={17} /> Save</Button
							>
						</div>
					</div>
					<div class="section-fields">
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
						{#if loaded && section === 'metadata'}
							<Field
								label="Name"
								name="name"
								bind:value={definition.metadata.name}
								required
								error={report.errors.find((issue) => issue.path === 'metadata.name')?.message}
							/><Field
								label="Description"
								name="description"
								help="The main idea behind the game."
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
							</div>
							<Panel title="Cover image (Optional)">
								<MediaField
									label="Ruleset cover"
									kind="image"
									name="ruleset-cover"
									compact
									bind:value={definition.metadata.coverAssetKey}
									{assets}
									{media}
									onuploadnew={() => selectSection('assets')}
								/>
							</Panel>
							<InlineValidationMessages issues={report.errors} path="metadata" />
						{:else if loaded}<VisualDefinitionEditor
								bind:definition
								section={section === 'metadata' ? 'teams' : section}
								{assets}
								{media}
								issues={report.errors}
								bind:selectedItems
								onnavigate={selectSection}
							/>{/if}
					</div>
				</div></Panel
			>
		</section>
		<aside class="overview">{@render overview()}</aside>
	</div>
</div>

{#snippet overview()}<div class="overview-content">
		<h2>Overview</h2>
		<label class="setup-progress"
			>Required steps: {requiredComplete} of 3 complete<progress max="3" value={requiredComplete}
			></progress></label
		>
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
			>{/if}{#if report.warnings.length || orphanedAssets.length}<h3 class="warning-heading">
				<CircleAlert size={18} aria-hidden="true" /> Warnings
			</h3>
			<ul>
				{#each report.warnings as issue (`${issue.path}:${issue.message}`)}<li>
						<span
							><strong>{humanIssueLocation(definition, issue, labels)}</strong>{issue.message}</span
						><button onclick={() => goToIssue(issue)}>Review</button>
					</li>{/each}
				{#each orphanedAssets as asset (asset.assetKey)}
					<li>
						<span>{asset.displayName} is not used anywhere.</span>
						<button onclick={() => selectSection('assets')}>Review</button>
					</li>
				{/each}
			</ul>
		{/if}
		{#if previewGuidance.length}<h3>Preview checks</h3>
			<ul class="preview-guidance">
				{#each previewGuidance as item (item.mode)}<li>
						<span>{item.message}</span>
						{#if item.action && item.section}<button onclick={() => selectSection(item.section!)}
								>{item.action}</button
							>{/if}
					</li>{/each}
			</ul>
		{/if}
		{#if loaded}<p
				class="readiness"
				class:ready={requiredComplete === 3 && report.errors.length === 0}
				role="status"
			>
				{#if requiredComplete < 3}<CircleAlert size={18} /> Missing required steps
				{:else if report.errors.length}<CircleAlert size={18} /> Resolve issues
				{:else}<Check size={18} /> Ready to use{/if}
			</p>{/if}
	</div>{/snippet}

{#snippet sectionNavigation()}
	<nav class="section-nav" aria-label="Ruleset sections">
		{#each [false, true] as optional (optional)}
			<div class="nav-group">
				<h2>
					{optional ? 'Optional' : 'Required'}
					<span
						>· {optional
							? `${optionalConfigured} configured`
							: `${requiredComplete} of 3 complete`}</span
					>
				</h2>
				{#each sections.filter((item) => item.optional === optional) as item (item.id)}
					<button
						class:active={section === item.id}
						aria-current={section === item.id ? 'page' : undefined}
						onclick={() => selectSection(item.id)}
					>
						<span
							class="step-icon"
							class:complete={states[item.id] === 'Complete'}
							class:attention={issueCounts[item.id] > 0}
							aria-hidden="true"
						>
							{#if issueCounts[item.id]}<CircleAlert
									size={18}
								/>{:else if states[item.id] === 'Complete'}<Check size={18} />{:else}<Circle
									size={18}
								/>{/if}
						</span>
						<span class="step-label">{item.label}</span><small>{sectionSummary(item)}</small>
					</button>
				{/each}
			</div>
		{/each}
	</nav>
{/snippet}
<Sheet open={sectionMenuOpen} title="Ruleset sections" close={() => (sectionMenuOpen = false)}
	>{@render sectionNavigation()}</Sheet
>
<Sheet open={overviewOpen} title="Ruleset overview" close={() => (overviewOpen = false)}
	>{@render overview()}</Sheet
>
<RulesetPreview
	open={previewOpen}
	close={() => (previewOpen = false)}
	{definition}
	{assets}
	{dirty}
	{loadPreview}
	onresult={recordPreviewResult}
/>
<Dialog
	open={leaveOpen}
	title="Leave with unsaved changes?"
	description="Choose what happens to your unsaved changes."
	close={() => (leaveOpen = false)}
	><p>Your changes have not been saved.</p>
	{#snippet actions()}<Button variant="ghost" onclick={() => (leaveOpen = false)}
			>Keep editing</Button
		><Button variant="secondary" onclick={discardAndLeave}>Discard and leave</Button><Button
			onclick={() => save(pendingPath as Parameters<typeof goto>[0])}>Save and leave</Button
		>{/snippet}</Dialog
>
<Dialog
	open={deleteOpen}
	title="Delete ruleset?"
	description="This removes the ruleset from the library and from new-game selection."
	close={() => (deleteOpen = false)}
	><p>This has no effect on existing games.</p>
	{#snippet actions()}<Button variant="ghost" onclick={() => (deleteOpen = false)}>Cancel</Button
		><Button variant="danger" onclick={remove}>Delete ruleset</Button>{/snippet}</Dialog
>
<p class="sr-only" aria-live="assertive">{announcement}</p>

<style>
	.editor {
		width: min(100%, 100rem);
		margin-inline: auto;
		padding: var(--space-5) max(var(--space-4), env(safe-area-inset-right)) var(--space-7)
			max(var(--space-4), env(safe-area-inset-left));
		gap: var(--space-6);
	}
	.reupload-warning {
		border: 1px solid var(--warning);
		background: var(--paper-deep);
		padding: var(--space-3);
	}
	.reupload-warning p {
		margin: var(--space-1) 0 0;
	}
	header {
		display: grid;
		grid-template-columns: 1fr auto;
		align-items: center;
		gap: var(--space-4);
	}
	header a {
		display: inline-flex;
		gap: var(--space-1);
		color: var(--action-dark);
		text-decoration: none;
	}
	.section-heading h1,
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
		flex-wrap: wrap;
		display: flex;
		gap: var(--space-2);
	}
	.workspace {
		display: grid;
		grid-template-columns: 17rem minmax(0, 1fr) 18rem;
		align-items: start;
		gap: var(--space-5);
	}
	.topbar-actions {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-5);
	}
	.section-rail {
		background: var(--paper-light);
		border: var(--border-subtle);
		padding: var(--space-4) var(--space-2);
	}
	.section-nav {
		display: grid;
		gap: var(--space-5);
	}
	.nav-group + .nav-group {
		border-top: var(--border-subtle);
		padding-top: var(--space-5);
	}
	.nav-group h2 {
		margin: 0 var(--space-3) var(--space-3);
		font-size: 0.72rem;
		line-height: 1.6;
	}
	.nav-group h2 span {
		color: var(--ink-soft);
	}
	.section-nav button {
		width: 100%;
		display: grid;
		grid-template-columns: 1.25rem minmax(0, 1fr) auto;
		gap: var(--space-2);
		align-items: center;
		min-height: var(--target-size);
		margin-block: var(--space-1);
		padding: var(--space-3) var(--space-2);
		border: 0;
		border-inline-start: 3px solid transparent;
		background: transparent;
		color: var(--ink);
		text-align: start;
		cursor: pointer;
	}
	.section-nav button.active {
		border-inline-start-color: var(--action);
		background: color-mix(in srgb, var(--action) 12%, var(--paper-light));
		color: var(--action-dark);
	}
	.section-nav button:hover {
		background: color-mix(in srgb, var(--action) 8%, var(--paper-light));
	}
	.section-nav button:focus-visible {
		outline: var(--focus-ring);
		outline-offset: 2px;
	}
	.step-icon {
		display: flex;
		color: var(--ink-soft);
	}
	.step-icon.complete {
		color: var(--success);
	}
	.step-icon.attention {
		color: var(--danger);
	}
	.step-label {
		font-weight: 700;
	}
	.section-nav small {
		text-align: end;
		max-width: 6rem;
	}
	.editor-body {
		padding: var(--space-4);
	}
	.section-fields {
		display: grid;
		gap: var(--space-5);
	}
	.readiness {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		border-top: var(--border-subtle);
		padding-top: var(--space-4);
		color: var(--danger);
		font-size: 1rem;
	}
	.readiness.ready {
		color: var(--success);
	}
	.editor-section {
		min-width: 0;
	}
	.section-heading {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		padding-bottom: var(--space-4);
		border-bottom: var(--border-subtle);
		margin-bottom: var(--space-6);
	}
	.section-heading h1 {
		font-size: 1.6rem;
	}
	.section-heading .eyebrow {
		margin: 0 0 var(--space-1);
		overflow-wrap: anywhere;
	}
	.setup-progress {
		display: grid;
		gap: var(--space-2);
	}
	.setup-progress progress {
		width: 100%;
		height: 0.5rem;
		accent-color: var(--success);
	}
	.overview {
		min-width: 0;
		position: sticky;
		top: var(--space-3);
		max-height: calc(100dvh - var(--space-6));
		overflow: auto;
		border: var(--border-subtle);
		background: var(--paper-light);
		padding: var(--space-5);
	}
	.overview-content {
		display: grid;
		gap: var(--space-3);
		min-width: 0;
	}
	.overview-content h2 {
		font-size: 1.2rem;
	}
	.overview-content h2,
	.overview-content h3,
	.overview-content p {
		margin: 0;
	}
	.warning-heading {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		color: var(--warning);
	}
	.eyebrow {
		font-family: var(--font-display);
		font-size: 0.7rem;
	}
	.overview-content dl {
		display: grid;
		gap: var(--space-1);
		margin: 0;
	}
	.overview-content dl div {
		display: flex;
		justify-content: space-between;
		gap: var(--space-2);
		padding-block: var(--space-3);
		border-bottom: var(--border-subtle);
	}
	.overview-content dt {
		color: var(--ink-soft);
	}
	.overview-content dd {
		margin: 0;
		font-weight: 700;
	}
	.overview-content ul {
		display: grid;
		gap: var(--space-2);
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.overview-content li {
		display: grid;
		gap: var(--space-1);
		min-width: 0;
		border-block-start: var(--border-subtle);
		padding-top: var(--space-2);
	}
	.overview-content li span {
		display: grid;
		min-width: 0;
		overflow-wrap: anywhere;
	}
	.overview-content li button,
	.inline-issues button {
		width: fit-content;
		border: 0;
		background: transparent;
		color: var(--action-dark);
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
	@media (min-width: 64rem) and (max-width: 89.99rem) {
		.workspace {
			grid-template-columns: 17rem minmax(0, 1fr);
		}
		.overview {
			grid-column: 2;
			position: static;
			max-height: none;
		}
	}
	@media (max-width: 63.99rem) {
		.editor-body {
			padding: 0;
		}
		header {
			grid-template-columns: 1fr;
		}
		.workspace {
			grid-template-columns: 1fr;
		}
		.section-rail,
		.overview {
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
