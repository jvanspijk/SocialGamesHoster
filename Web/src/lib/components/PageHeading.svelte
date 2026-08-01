<script lang="ts">
	import type { Snippet } from 'svelte';
	import ContentHeader from './ContentHeader.svelte';

	let {
		eyebrow,
		title: pageTitle,
		description = '',
		variant = 'default',
		actions
	}: {
		eyebrow?: string;
		title: string;
		description?: string;
		variant?: 'default' | 'spacious' | 'compact' | 'flush';
		actions?: Snippet;
	} = $props();
</script>

<header
	class:spacious={variant === 'spacious'}
	class:compact={variant === 'compact'}
	class:flush={variant === 'flush'}
>
	<ContentHeader
		{eyebrow}
		{description}
		{actions}
		alignment="end"
		density={variant === 'default' || variant === 'spacious' ? 'regular' : variant}
	>
		{#snippet title()}<h1>{pageTitle}</h1>{/snippet}
	</ContentHeader>
</header>

<style>
	header {
		margin-block-end: var(--space-5);
	}

	header.spacious {
		margin-block-end: var(--space-6);
	}

	header.compact {
		margin-block-end: var(--space-4);
	}

	header.flush {
		margin-block-end: 0;
	}
</style>
