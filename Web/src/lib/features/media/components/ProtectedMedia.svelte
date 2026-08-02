<script lang="ts">
	import Media from '$lib/components/Media.svelte';
	import { loadProtectedMedia } from '../media';

	let {
		src,
		kind,
		alt = '',
		controls = false,
		autoplay = false
	}: {
		src: string;
		kind: 'image' | 'audio';
		alt?: string;
		controls?: boolean;
		autoplay?: boolean;
	} = $props();

	const loader = $derived.by(() => {
		const endpoint = src;
		return () => loadProtectedMedia(endpoint);
	});
</script>

<Media {loader} {kind} {alt} {controls} {autoplay} />
