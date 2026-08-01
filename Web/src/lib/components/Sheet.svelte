<script lang="ts">
	import type { Snippet } from 'svelte';
	import ModalOverlay from './ModalOverlay.svelte';

	let {
		open,
		title,
		close,
		children
	}: {
		open: boolean;
		title: string;
		close: () => void;
		children: Snippet;
	} = $props();
</script>

<ModalOverlay {open} {title} {close} presentation="sheet" {children} />
