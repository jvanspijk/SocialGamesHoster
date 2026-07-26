<script lang="ts">
	import { fetchBlob } from '$lib/api/client';

	let {
		src,
		kind,
		alt = '',
		controls = false
	}: {
		src: string;
		kind: 'image' | 'audio';
		alt?: string;
		controls?: boolean;
	} = $props();

	let objectUrl = $state('');
	let failed = $state(false);
	let currentUrl = '';

	$effect(() => {
		let disposed = false;
		const path = src;
		if (currentUrl) URL.revokeObjectURL(currentUrl);
		currentUrl = '';
		objectUrl = '';
		failed = false;
		void fetchBlob(path.replace('/api/app/v1', ''))
			.then((blob) => {
				if (disposed) return;
				currentUrl = URL.createObjectURL(blob);
				objectUrl = currentUrl;
			})
			.catch(() => {
				failed = true;
			});
		return () => {
			disposed = true;
			if (currentUrl) URL.revokeObjectURL(currentUrl);
			currentUrl = '';
		};
	});
</script>

{#if failed}
	<span class="unavailable">Media unavailable</span>
{:else if !objectUrl}
	<span class="loading">Loading media…</span>
{:else if kind === 'image'}
	<img src={objectUrl} {alt} />
{:else}
	<audio src={objectUrl} {controls} preload="none">
		<track kind="captions" />
	</audio>
{/if}

<style>
	img {
		max-width: 100%;
	}

	audio {
		max-width: 100%;
	}

	.loading,
	.unavailable {
		color: var(--ink-faint);
		font-size: 0.8rem;
	}
</style>
