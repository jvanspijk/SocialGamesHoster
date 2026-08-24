<script lang="ts">
	import Button from '$lib/components/Button.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import Field from '$lib/components/Field.svelte';
	import ProtectedMedia from '$lib/features/media/components/ProtectedMedia.svelte';
	import type { RulesetDefinition } from '$lib/api/types';
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
	const filtered = $derived(
		assets.filter((asset) =>
			`${asset.displayName} ${asset.kind}`.toLowerCase().includes(query.trim().toLowerCase())
		)
	);
	const selected = $derived(assets.find((asset) => asset.assetKey === selectedKey));
	const usages = $derived(selected ? assetUsages(definition, selected.assetKey) : []);

	function select(key: string) {
		selectedKey = key;
		const asset = assets.find((item) => item.assetKey === key);
		displayName = asset?.displayName ?? '';
		accessibilityText = asset?.accessibilityText ?? '';
		error = '';
	}

	async function saveDetails() {
		if (!selected) return;
		saving = true;
		error = '';
		try {
			await media.update(selected.assetKey, displayName, accessibilityText);
		} catch (caught) {
			error = caught instanceof Error ? caught.message : 'The media details could not be saved.';
		} finally {
			saving = false;
		}
	}

	async function replaceEverywhere(event: Event) {
		const file = (event.currentTarget as HTMLInputElement).files?.[0];
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
			error = caught instanceof Error ? caught.message : 'The media item could not be replaced.';
		} finally {
			saving = false;
			(event.currentTarget as HTMLInputElement).value = '';
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
			error = caught instanceof Error ? caught.message : 'The media item could not be deleted.';
		} finally {
			saving = false;
		}
	}
</script>

<section class="library" aria-labelledby="media-library-title">
	<header>
		<div>
			<h2 id="media-library-title">Reusable media</h2>
			<p>
				Changes here apply to this ruleset after Save. Replace everywhere keeps every usage
				connected.
			</p>
		</div>
		<Field label="Search media" name="media-search" bind:value={query} />
	</header>
	{#if assets.length === 0}
		<EmptyState
			title="No reusable media yet"
			description="Upload an image or sound from any place where it is used. It will appear here."
		/>
	{:else}
		<div class="master-detail">
			<ul class="asset-list" aria-label="Reusable media">
				{#each filtered as asset (asset.assetKey)}
					<li>
						<button
							aria-current={selectedKey === asset.assetKey ? 'true' : undefined}
							onclick={() => select(asset.assetKey)}
						>
							<strong>{asset.displayName}</strong><small
								>{asset.kind === 'image' ? 'Image' : 'Audio'} · {assetUsages(
									definition,
									asset.assetKey
								).length} usages</small
							>
						</button>
					</li>
				{/each}
			</ul>
			<div class="detail">
				{#if selected}
					<h3>{selected.displayName}</h3>
					<ProtectedMedia
						src={selected.preview}
						kind={selected.kind}
						alt={selected.accessibilityText || selected.displayName}
						controls={selected.kind === 'audio'}
					/>
					<p class="metadata">
						{selected.kind === 'image' && selected.metadata.width
							? `${selected.metadata.width} × ${selected.metadata.height} pixels`
							: selected.metadata.durationSeconds
								? `${Math.round(selected.metadata.durationSeconds)} seconds`
								: selected.mimeType}
					</p>
					<Field label="Display name" name="asset-display-name" bind:value={displayName} required />
					<Field
						label={selected.kind === 'image' ? 'Image description' : 'Audio alternative'}
						name="asset-accessibility"
						bind:value={accessibilityText}
						multiline
					/>
					<Button loading={saving} onclick={saveDetails}>Apply details</Button>
					<section class="usages" aria-labelledby="asset-usages-title">
						<h4 id="asset-usages-title">Used by</h4>
						{#if usages.length}
							{#each usages as usage (`${usage.section}:${usage.itemId ?? ''}:${usage.label}`)}
								<button onclick={() => onnavigate(usage.section, usage.itemId)}
									>{usage.label}</button
								>
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
					<div class="actions">
						<Button variant="secondary" loading={saving} onclick={() => fileInput?.click()}
							>Replace everywhere</Button
						>
						<Button variant="danger" disabled={usages.length > 0} loading={saving} onclick={remove}
							>Delete media item</Button
						>
					</div>
					{#if usages.length}<p class="hint">
							Remove every usage before deleting this media item.
						</p>{/if}
					{#if error}<p class="error" role="alert">{error}</p>{/if}
				{:else}<p>Select a media item to view details and usages.</p>{/if}
			</div>
		</div>
	{/if}
</section>

<style>
	.library,
	.detail,
	.usages {
		display: grid;
		gap: var(--space-3);
	}
	header {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(14rem, 20rem);
		gap: var(--space-4);
		align-items: end;
	}
	header p,
	h2,
	h3,
	h4,
	.metadata {
		margin: 0;
	}
	.master-detail {
		display: grid;
		grid-template-columns: minmax(13rem, 0.8fr) minmax(0, 1.2fr);
		border: var(--border-strong);
	}
	.asset-list {
		display: grid;
		align-content: start;
		margin: 0;
		border-inline-end: var(--border-subtle);
		padding: 0;
		list-style: none;
	}
	.asset-list button,
	.usages button {
		display: grid;
		gap: var(--space-1);
		border: 0;
		border-bottom: var(--border-subtle);
		background: transparent;
		color: var(--ink);
		padding: var(--space-3);
		text-align: start;
		cursor: pointer;
	}
	.asset-list button[aria-current='true'] {
		background: var(--paper-deep);
		box-shadow: inset 4px 0 var(--crimson);
	}
	.detail {
		padding: var(--space-4);
	}
	.detail :global(img) {
		width: min(100%, 24rem);
		max-height: 14rem;
		object-fit: contain;
		background: var(--paper-deep);
	}
	.detail :global(audio) {
		width: min(100%, 30rem);
	}
	.actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}
	.visually-hidden {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip: rect(0 0 0 0);
	}
	.error {
		color: var(--danger);
	}
	@media (max-width: 63.99rem) {
		header,
		.master-detail {
			grid-template-columns: 1fr;
		}
		.asset-list {
			border-inline-end: 0;
			border-bottom: var(--border-subtle);
		}
	}
</style>
