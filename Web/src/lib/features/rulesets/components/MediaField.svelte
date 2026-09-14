<script lang="ts">
	import { Plus } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Field from '$lib/components/Field.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import ProtectedMedia from '$lib/features/media/components/ProtectedMedia.svelte';
	import type { AssetOption, MediaActions } from './definition-editor';

	let {
		label,
		kind,
		value = $bindable<string | undefined>(),
		assets,
		media,
		name,
		compact = false,
		onuploadnew
	}: {
		label: string;
		kind: 'image' | 'audio';
		value?: string;
		assets: AssetOption[];
		media: MediaActions;
		name: string;
		compact?: boolean;
		onuploadnew?: () => void;
	} = $props();

	let displayName = $state('');
	let accessibilityText = $state('');
	let uploading = $state(false);
	let uploadError = $state('');
	let input = $state<HTMLInputElement>();
	const options = $derived(assets.filter((asset) => asset.kind === kind));
	const selected = $derived(options.find((asset) => asset.assetKey === value));
	const selectedValue = $derived(selected?.assetKey ?? '');

	$effect(() => {
		if (value && !selected) value = '';
	});

	function chooseUpload() {
		input?.click();
	}

	async function upload(event: Event) {
		const target = event.currentTarget as HTMLInputElement;
		const file = target.files?.[0];
		if (!file) return;
		uploading = true;
		uploadError = '';
		try {
			const asset = await media.upload(
				file,
				kind,
				displayName.trim() || file.name,
				accessibilityText.trim()
			);
			value = asset.assetKey;
		} catch (caught) {
			uploadError =
				caught instanceof Error ? caught.message : 'The asset file could not be uploaded.';
		} finally {
			uploading = false;
			target.value = '';
		}
	}
</script>

<section class="media-field" aria-label={label}>
	<SelectField
		label={`${label} — Choose existing file`}
		{name}
		value={selectedValue}
		onchange={(next) => (value = next)}
		options={[
			{ value: '', label: kind === 'image' ? 'No image' : 'No audio' },
			...options.map((asset) => ({ value: asset.assetKey, label: asset.displayName }))
		]}
	/>
	{#if selected}
		<div class="preview">
			<ProtectedMedia
				src={selected.preview}
				{kind}
				alt={selected.accessibilityText || selected.displayName}
				controls={kind === 'audio'}
			/>
			<div>
				<strong>{selected.displayName}</strong><small
					>{selected.accessibilityText || 'No accessibility text added'}</small
				>
			</div>
		</div>
	{/if}
	{#if !compact}{@render uploadDetails()}{/if}
	<input
		class="visually-hidden"
		bind:this={input}
		type="file"
		accept={kind === 'image'
			? 'image/jpeg,image/png,image/webp'
			: 'audio/mpeg,audio/mp4,audio/ogg,audio/wav'}
		onchange={upload}
	/>
	<div class="actions">
		{#if selected}
			<Button variant="secondary" loading={uploading} onclick={chooseUpload}>Replace</Button>
		{:else}
			<Button variant="secondary" loading={uploading} onclick={onuploadnew ?? chooseUpload}
				><Plus size={16} aria-hidden="true" /> Upload new</Button
			>
		{/if}
	</div>
	{#if uploadError}<p class="error" role="alert">{uploadError}</p>{/if}
</section>

{#snippet uploadDetails()}
	<div class="upload-details">
		<Field label="Asset name" name={`${name}-upload-name`} bind:value={displayName} />
		<Field
			label={kind === 'image' ? 'Image description' : 'Accessibility description'}
			name={`${name}-upload-accessibility`}
			bind:value={accessibilityText}
			help="Describe the image or sound for players who cannot see or hear it."
		/>
	</div>
{/snippet}

<style>
	.media-field {
		display: grid;
		gap: var(--space-3);
		border-block: var(--border-subtle);
		padding-block: var(--space-3);
	}
	.preview {
		display: grid;
		grid-template-columns: 6rem minmax(0, 1fr);
		align-items: center;
		gap: var(--space-3);
	}
	.preview :global(img) {
		width: 6rem;
		height: 4rem;
		object-fit: cover;
		border: var(--border-subtle);
	}
	.preview :global(audio) {
		width: min(100%, 18rem);
	}
	.preview div {
		display: grid;
		gap: var(--space-1);
	}
	.preview small {
		color: var(--ink-soft);
	}
	.upload-details {
		display: grid;
		gap: var(--space-3);
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
	@media (max-width: 47.99rem) {
		.preview {
			grid-template-columns: 1fr;
		}
	}
</style>
