<script lang="ts">
	import { Image as ImageIcon, Music2, Plus } from '@lucide/svelte';
	import type { RulesetDefinition } from '$lib/api/types';
	import Button from '$lib/components/Button.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Field from '$lib/components/Field.svelte';
	import ProtectedMedia from '$lib/features/media/components/ProtectedMedia.svelte';
	import type { EditorSection } from '../editor-state';
	import { assetUsages, type AssetOption, type MediaActions } from './definition-editor';

	let {
		definition,
		assets,
		media,
		onnavigate
	}: {
		definition: RulesetDefinition;
		assets: AssetOption[];
		media: MediaActions;
		onnavigate: (section: EditorSection, itemId?: string) => void;
	} = $props();
	let query = $state('');
	let selectedKey = $state('');
	let displayName = $state('');
	let accessibilityText = $state('');
	let error = $state('');
	let saving = $state(false);
	let fileInput = $state<HTMLInputElement>();
	let imageInput = $state<HTMLInputElement>();
	let soundInput = $state<HTMLInputElement>();
	let uploading = $state<'image' | 'audio' | ''>('');
	let uploadError = $state('');
	let uploadErrorKind = $state<'image' | 'audio' | ''>('');
	const normalizedQuery = $derived(query.trim().toLocaleLowerCase());
	const imageAssets = $derived(assets.filter((asset) => asset.kind === 'image'));
	const soundAssets = $derived(assets.filter((asset) => asset.kind === 'audio'));
	const filteredImages = $derived(imageAssets.filter(matchesQuery));
	const filteredSounds = $derived(soundAssets.filter(matchesQuery));
	const selected = $derived(assets.find((asset) => asset.assetKey === selectedKey));
	const usages = $derived(selected ? assetUsages(definition, selected.assetKey) : []);

	function matchesQuery(asset: AssetOption) {
		if (!normalizedQuery) return true;
		const kind = asset.kind === 'image' ? 'image' : 'audio sound';
		const usageLabels = assetUsages(definition, asset.assetKey)
			.map((usage) => usage.label)
			.join(' ');
		return `${asset.displayName} ${kind} ${usageLabels}`
			.toLocaleLowerCase()
			.includes(normalizedQuery);
	}

	function select(key: string) {
		if (selectedKey === key) {
			selectedKey = '';
			displayName = '';
			accessibilityText = '';
			error = '';
			return;
		}
		selectedKey = key;
		const asset = assets.find((item) => item.assetKey === key);
		displayName = asset?.displayName ?? '';
		accessibilityText = asset?.accessibilityText ?? '';
		error = '';
	}

	function chooseUpload(kind: 'image' | 'audio') {
		(kind === 'image' ? imageInput : soundInput)?.click();
	}

	async function upload(kind: 'image' | 'audio', event: Event) {
		const target = event.currentTarget as HTMLInputElement;
		const file = target.files?.[0];
		if (!file) return;
		uploading = kind;
		uploadError = '';
		uploadErrorKind = '';
		try {
			const asset = await media.upload(file, kind, file.name, '');
			selectedKey = asset.assetKey;
			displayName = asset.displayName;
			accessibilityText = asset.accessibilityText;
		} catch (caught) {
			uploadError =
				caught instanceof Error
					? caught.message
					: `The ${kind === 'image' ? 'image' : 'sound'} could not be uploaded.`;
			uploadErrorKind = kind;
		} finally {
			uploading = '';
			target.value = '';
		}
	}

	async function saveDetails() {
		if (!selected) return;
		saving = true;
		error = '';
		try {
			await media.update(selected.assetKey, displayName, accessibilityText);
		} catch (caught) {
			error = caught instanceof Error ? caught.message : 'The asset details could not be saved.';
		} finally {
			saving = false;
		}
	}

	async function replaceEverywhere(event: Event) {
		const target = event.currentTarget as HTMLInputElement;
		const file = target.files?.[0];
		if (!file || !selected) return;
		saving = true;
		error = '';
		try {
			await media.upload(
				file,
				selected.kind,
				displayName || selected.displayName,
				accessibilityText,
				selected.assetKey
			);
		} catch (caught) {
			error = caught instanceof Error ? caught.message : 'The Asset could not be replaced.';
		} finally {
			saving = false;
			target.value = '';
		}
	}

	async function remove() {
		if (!selected || usages.length) return;
		saving = true;
		error = '';
		try {
			await media.remove(selected.assetKey);
			selectedKey = '';
		} catch (caught) {
			error = caught instanceof Error ? caught.message : 'The Asset could not be deleted.';
		} finally {
			saving = false;
		}
	}
</script>

<section class="library" aria-label="Asset library">
	<input
		class="visually-hidden"
		bind:this={imageInput}
		type="file"
		accept="image/jpeg,image/png,image/webp"
		aria-label="Choose image to upload"
		onchange={(event) => upload('image', event)}
	/>
	<input
		class="visually-hidden"
		bind:this={soundInput}
		type="file"
		accept="audio/mpeg,audio/mp4,audio/ogg,audio/wav"
		aria-label="Choose sound to upload"
		onchange={(event) => upload('audio', event)}
	/>

	<div class="media-workspace">
		<div class="collection">
			<div class="search">
				<Field
					label="Search assets"
					name="asset-search"
					placeholder="Filter by name, type, or usage…"
					bind:value={query}
				/>
			</div>
			{@render assetSection('image', 'Images', imageAssets, filteredImages)}
			{@render assetSection('audio', 'Sounds', soundAssets, filteredSounds)}
		</div>

		<aside class="details" aria-labelledby="item-details-title">
			<h2 id="item-details-title">Item details</h2>
			{#if selected}
				<div class="detail-content">
					<header class="detail-heading">
						<p>{selected.kind === 'image' ? 'Image' : 'Sound'}</p>
						<h3>{selected.displayName}</h3>
					</header>
					<div class:audio-preview={selected.kind === 'audio'} class="preview">
						<ProtectedMedia
							src={selected.preview}
							kind={selected.kind}
							alt={selected.accessibilityText || selected.displayName}
							controls={selected.kind === 'audio'}
						/>
					</div>
					<dl class="metadata">
						<div>
							<dt>File type</dt>
							<dd>{selected.mimeType}</dd>
						</div>
						<div>
							<dt>{selected.kind === 'image' ? 'Dimensions' : 'Duration'}</dt>
							<dd>
								{#if selected.kind === 'image' && selected.metadata.width}
									{selected.metadata.width} × {selected.metadata.height} px
								{:else if selected.metadata.durationSeconds}
									{Math.round(selected.metadata.durationSeconds)} seconds
								{:else}Not available{/if}
							</dd>
						</div>
						<div>
							<dt>Used in</dt>
							<dd>{usages.length} {usages.length === 1 ? 'place' : 'places'}</dd>
						</div>
					</dl>

					<div class="detail-form">
						<Field
							label="Display name"
							name="asset-display-name"
							bind:value={displayName}
							required
						/>
						<Field
							label="Description"
							name="asset-accessibility"
							bind:value={accessibilityText}
							multiline
						/>
						<Button loading={saving} onclick={saveDetails}>Save</Button>
					</div>

					<section class="usages" aria-labelledby="asset-usages-title">
						<h4 id="asset-usages-title">Used by</h4>
						{#if usages.length}
							{#each usages as usage (`${usage.section}:${usage.itemId ?? ''}:${usage.label}`)}
								<button onclick={() => onnavigate(usage.section, usage.itemId)}>
									{usage.label}
								</button>
							{/each}
						{:else}<p>Not currently used.</p>{/if}
					</section>

					<input
						class="visually-hidden"
						bind:this={fileInput}
						type="file"
						accept={selected.kind === 'image'
							? 'image/jpeg,image/png,image/webp'
							: 'audio/mpeg,audio/mp4,audio/ogg,audio/wav'}
						onchange={replaceEverywhere}
					/>
					<div class="detail-actions">
						<Button variant="secondary" loading={saving} onclick={() => fileInput?.click()}>
							Replace
						</Button>
						<Button variant="danger" disabled={usages.length > 0} loading={saving} onclick={remove}
							>Delete</Button
						>
					</div>
					{#if usages.length}
						<p class="hint">This asset cannot be deleted since it's still being used.</p>
					{/if}
					{#if error}<p class="error" role="alert">{error}</p>{/if}
				</div>
			{:else}
				<div class="detail-empty">
					<div aria-hidden="true"><ImageIcon size={30} strokeWidth={1.5} /></div>
					<h3>Select a Asset</h3>
					<p>Choose an image or sound to preview it, edit its details, or see where it is used.</p>
				</div>
			{/if}
		</aside>
	</div>
</section>

{#snippet assetSection(
	kind: 'image' | 'audio',
	title: string,
	kindAssets: AssetOption[],
	filteredAssets: AssetOption[]
)}
	<section class="asset-section" aria-labelledby={`${kind}-library-title`}>
		<div class="section-heading">
			<h2 id={`${kind}-library-title`}>{title}</h2>
			<Button variant="secondary" loading={uploading === kind} onclick={() => chooseUpload(kind)}>
				<Plus size={16} aria-hidden="true" /> Upload {kind === 'image' ? 'image' : 'sound'}
			</Button>
		</div>
		{#if kindAssets.length === 0}
			<div class="empty-frame">
				<EmptyState
					title={`No ${title.toLocaleLowerCase()} yet`}
					description={`Upload a ${kind === 'image' ? 'JPEG, PNG, or WebP image' : 'sound file'} to use in this ruleset.`}
				/>
			</div>
		{:else if filteredAssets.length === 0}
			<p class="no-results">No matching {title.toLocaleLowerCase()}.</p>
		{:else}
			<ul class="asset-list" aria-label={title}>
				{#each filteredAssets as asset (asset.assetKey)}
					<li>
						<button
							aria-current={selectedKey === asset.assetKey ? 'true' : undefined}
							onclick={() => select(asset.assetKey)}
						>
							<span class="asset-icon" aria-hidden="true">
								{#if asset.kind === 'image'}
									<ImageIcon size={22} strokeWidth={1.5} />
								{:else}
									<Music2 size={22} strokeWidth={1.5} />
								{/if}
							</span>
							<span class="asset-copy">
								<strong>{asset.displayName}</strong>
								<small>
									{asset.kind === 'image' ? 'Image' : 'Sound'} · {assetUsages(
										definition,
										asset.assetKey
									).length}
									{assetUsages(definition, asset.assetKey).length === 1 ? 'usage' : 'usages'}
								</small>
							</span>
							{#if selectedKey === asset.assetKey}<span class="selection-label">Selected</span>{/if}
						</button>
					</li>
				{/each}
			</ul>
		{/if}
		{#if uploadErrorKind === kind}<p class="error" role="alert">{uploadError}</p>{/if}
	</section>
{/snippet}

<style>
	.library,
	.collection,
	.asset-section,
	.detail-content,
	.detail-form,
	.usages {
		display: grid;
		gap: var(--space-3);
	}
	.media-workspace {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		align-items: start;
		gap: var(--space-5);
	}
	.collection {
		min-width: 0;
		gap: var(--space-5);
	}
	.search {
		width: min(100%, 32rem);
	}
	.section-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border-bottom: var(--border-subtle);
		padding-bottom: var(--space-2);
	}
	.section-heading :global(button) {
		gap: 0.35rem;
		padding-inline: 0.75rem;
	}
	.asset-section + .asset-section {
		border-top: var(--border-strong);
		padding-top: var(--space-5);
	}
	h2,
	h3,
	h4,
	.detail-heading p,
	.detail-empty p,
	.usages p,
	.hint,
	.error,
	.no-results {
		margin: 0;
	}
	.section-heading h2,
	.details > h2 {
		font-size: var(--font-size-lg);
	}
	.asset-list {
		display: grid;
		gap: var(--space-2);
		margin: 0;
		padding: 0;
		list-style: none;
	}
	.asset-list button {
		display: grid;
		width: 100%;
		min-height: var(--target-size);
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-3);
		border: var(--border-subtle);
		background: color-mix(in srgb, var(--paper-light) 72%, transparent);
		color: var(--ink);
		padding: var(--space-3);
		text-align: start;
		cursor: pointer;
		transition:
			background var(--speed-fast) ease-out,
			box-shadow var(--speed-fast) ease-out;
	}
	.asset-list button:hover {
		background: color-mix(in srgb, var(--action) 8%, var(--paper-light));
	}
	.asset-list button:focus-visible {
		outline: var(--focus-ring);
		outline-offset: 2px;
	}
	.asset-list button[aria-current='true'] {
		background: color-mix(in srgb, var(--action) 10%, var(--paper-light));
		box-shadow:
			inset 4px 0 var(--action),
			inset 0 0 0 1px var(--action);
	}
	.asset-icon {
		display: grid;
		width: 3.25rem;
		height: 3.25rem;
		place-items: center;
		border: 1px solid var(--gold-dark);
		background: var(--paper-deep);
		color: var(--ink-soft);
	}
	.asset-copy {
		display: grid;
		min-width: 0;
		gap: var(--space-1);
	}
	.asset-copy strong,
	.asset-copy small {
		overflow-wrap: anywhere;
	}
	.asset-copy small,
	.detail-heading p {
		color: var(--ink-soft);
	}
	.selection-label,
	.detail-heading p {
		font-family: var(--font-display);
		font-size: var(--font-size-xs);
		font-weight: 700;
	}
	.selection-label {
		color: var(--action-dark);
	}
	.empty-frame,
	.no-results {
		border: 1px dashed var(--gold-dark);
		background: var(--surface-paper-veil);
	}
	.no-results {
		color: var(--ink-soft);
		padding: var(--space-5);
		text-align: center;
	}
	.details {
		min-width: 0;
		border: var(--border-subtle);
		background: color-mix(in srgb, var(--paper-light) 78%, transparent);
		padding: var(--space-4);
	}
	.details > h2 {
		border-bottom: var(--border-subtle);
		padding-bottom: var(--space-3);
	}
	.detail-content {
		gap: var(--space-4);
		padding-top: var(--space-4);
	}
	.detail-heading {
		display: grid;
		gap: var(--space-1);
	}
	.detail-heading h3 {
		overflow-wrap: anywhere;
	}
	.preview {
		display: grid;
		min-height: 12rem;
		place-items: center;
		border: var(--border-subtle);
		background: var(--paper-deep);
		padding: var(--space-2);
	}
	.preview.audio-preview {
		min-height: 7rem;
	}
	.preview :global(img) {
		width: 100%;
		max-height: 18rem;
		object-fit: contain;
	}
	.preview :global(audio) {
		width: 100%;
	}
	.metadata {
		display: grid;
		gap: var(--space-2);
		margin: 0;
	}
	.metadata div {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1.3fr);
		gap: var(--space-2);
	}
	.metadata dt {
		color: var(--ink-soft);
	}
	.metadata dd {
		margin: 0;
		font-weight: 700;
		text-align: end;
		overflow-wrap: anywhere;
	}
	.detail-form,
	.usages,
	.detail-actions {
		border-top: var(--border-subtle);
		padding-top: var(--space-4);
	}
	.detail-form :global(button),
	.detail-actions :global(button) {
		width: 100%;
	}
	.usages button {
		width: fit-content;
		min-height: var(--target-size);
		border: 0;
		background: transparent;
		color: var(--action-dark);
		padding: 0;
		text-align: start;
		text-decoration: underline;
		cursor: pointer;
	}
	.detail-actions {
		display: grid;
		gap: var(--space-2);
	}
	.detail-empty {
		display: grid;
		min-height: 20rem;
		place-content: center;
		justify-items: center;
		gap: var(--space-2);
		color: var(--ink-soft);
		padding: var(--space-5);
		text-align: center;
	}
	.detail-empty div {
		display: grid;
		width: 3rem;
		height: 3rem;
		place-items: center;
		border: 1px solid var(--gold-dark);
		border-radius: 50%;
	}
	.detail-empty h3 {
		color: var(--ink);
	}
	.hint {
		color: var(--ink-soft);
		font-size: var(--font-size-sm);
	}
	.error {
		color: var(--danger);
	}
	.visually-hidden {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
	}
	@media (min-width: 64rem) {
		.media-workspace {
			grid-template-columns: minmax(0, 1fr) minmax(17rem, 0.42fr);
		}
		.details {
			position: sticky;
			top: var(--space-3);
			max-height: calc(100dvh - var(--space-6));
			overflow: auto;
		}
	}
	@media (max-width: 47.99rem) {
		.section-heading {
			align-items: stretch;
			flex-direction: column;
		}
		.section-heading :global(button) {
			width: 100%;
		}
		.asset-list button {
			grid-template-columns: auto minmax(0, 1fr);
		}
		.selection-label {
			grid-column: 2;
		}
		.details {
			padding: var(--space-3);
		}
	}
</style>
