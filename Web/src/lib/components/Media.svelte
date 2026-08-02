<script lang="ts">
	export type PreparedMediaSource = string | Blob;
	export type MediaLoader = () => Promise<PreparedMediaSource>;

	let {
		source,
		loader,
		kind,
		alt = '',
		controls = false,
		autoplay = false
	}: {
		source?: PreparedMediaSource;
		loader?: MediaLoader;
		kind: 'image' | 'audio';
		alt?: string;
		controls?: boolean;
		autoplay?: boolean;
	} = $props();

	let url = $state('');
	let loading = $state(true);
	let failed = $state(false);

	function asUrl(prepared: PreparedMediaSource): { url: string; revoke?: () => void } {
		if (prepared instanceof Blob) {
			const objectUrl = URL.createObjectURL(prepared);
			return { url: objectUrl, revoke: () => URL.revokeObjectURL(objectUrl) };
		}
		return { url: prepared };
	}

	$effect(() => {
		let disposed = false;
		let revoke: (() => void) | undefined;

		function show(prepared: PreparedMediaSource) {
			const resolved = asUrl(prepared);
			if (disposed) {
				resolved.revoke?.();
				return;
			}
			url = resolved.url;
			revoke = resolved.revoke;
			loading = false;
		}

		url = '';
		loading = true;
		failed = false;
		if (source !== undefined) show(source);
		else if (loader) {
			void loader()
				.then(show)
				.catch(() => {
					if (!disposed) {
						failed = true;
						loading = false;
					}
				});
		} else {
			failed = true;
			loading = false;
		}

		return () => {
			disposed = true;
			revoke?.();
		};
	});
</script>

{#if failed}
	<span class="unavailable">Media unavailable</span>
{:else if loading}
	<span class="loading">Loading media…</span>
{:else if kind === 'image'}
	<img src={url} {alt} decoding="async" />
{:else}
	<audio src={url} {controls} {autoplay} preload={autoplay ? 'auto' : 'none'}>
		<track kind="captions" />
	</audio>
{/if}

<style>
	img,
	audio {
		max-width: 100%;
	}
	.loading,
	.unavailable {
		color: var(--ink-faint);
		font-size: 0.8rem;
	}
</style>
