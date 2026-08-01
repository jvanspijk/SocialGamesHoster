<script lang="ts">
	import { Inbox } from '@lucide/svelte';
	import type { Snippet } from 'svelte';
	import Button from './Button.svelte';

	let {
		title,
		description,
		actionLabel,
		onaction,
		icon
	}: {
		title: string;
		description: string;
		actionLabel?: string;
		onaction?: () => void;
		icon?: Snippet;
	} = $props();
</script>

<section class="empty-state" aria-label={title}>
	<div class="icon" aria-hidden="true">
		{#if icon}
			{@render icon()}
		{:else}
			<Inbox size={34} strokeWidth={1.5} />
		{/if}
	</div>
	<h2>{title}</h2>
	<p>{description}</p>
	{#if actionLabel && onaction}
		<Button onclick={onaction}>{actionLabel}</Button>
	{/if}
</section>

<style>
	.empty-state {
		display: grid;
		justify-items: center;
		gap: var(--space-2);
		padding: var(--space-5);
		text-align: center;
	}

	.icon {
		display: grid;
		width: 3rem;
		height: 3rem;
		place-items: center;
		border: 1px solid var(--gold-dark);
		border-radius: 50%;
		color: var(--ink-soft);
	}

	h2,
	p {
		margin: 0;
	}

	h2 {
		font-size: 1.1rem;
	}

	p {
		max-width: 34rem;
		color: var(--ink-soft);
	}
</style>
