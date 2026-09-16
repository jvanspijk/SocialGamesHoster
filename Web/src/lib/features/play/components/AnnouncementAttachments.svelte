<script lang="ts">
	import ProtectedMedia from '$lib/features/media/components/ProtectedMedia.svelte';

	let {
		image,
		audio
	}: {
		image?: { url: string; description: string };
		audio?: { url: string; alternative: string };
	} = $props();
</script>

{#if image}
	<figure>
		<ProtectedMedia src={image.url} kind="image" alt={image.description} />
	</figure>
{/if}
{#if audio}
	<div class="audio-attachment">
		<ProtectedMedia src={audio.url} kind="audio" autoplay />
	</div>
{/if}

<style>
	figure {
		margin: 0 0 var(--space-4);
	}

	figure :global(img) {
		max-height: 18rem;
		object-fit: contain;
	}

	.audio-attachment {
		display: grid;
		gap: var(--space-2);
		margin-block-end: var(--space-4);
	}
</style>
