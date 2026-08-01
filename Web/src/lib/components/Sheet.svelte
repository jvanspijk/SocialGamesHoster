<script lang="ts">
	import type { Snippet } from 'svelte';
	import ModalOverlay from './ModalOverlay.svelte';

	let {
		open,
		title,
		description = '',
		close,
		children,
		actions
	}: {
		open: boolean;
		title: string;
		description?: string;
		close: () => void;
		children: Snippet;
		actions?: Snippet;
	} = $props();
</script>

<ModalOverlay {open} {title} {description} {close} presentation="sheet" {children} {actions} />
