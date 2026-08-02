<script lang="ts">
	import { ArrowLeft } from '@lucide/svelte';
	import IconButton from '$lib/components/IconButton.svelte';

	let {
		label,
		initial,
		typeLabel,
		playersCanPost,
		canModerate = false,
		archived = false,
		onback,
		ontoggleposting
	}: {
		label: string;
		initial: string;
		typeLabel: string;
		playersCanPost: boolean;
		canModerate?: boolean;
		archived?: boolean;
		onback: () => void;
		ontoggleposting?: () => void;
	} = $props();
</script>

<header class="conversation-header">
	<div class="back">
		<IconButton label="Back to conversations" variant="ghost" onclick={onback}>
			{#snippet icon()}<ArrowLeft size={21} />{/snippet}
		</IconButton>
	</div>
	<span class="conversation-avatar">{initial}</span>
	<div>
		<h2>{label}</h2>
		<p>
			{typeLabel}
			{#if !playersCanPost}
				· Players read-only{/if}
		</p>
	</div>
	{#if canModerate && !archived}
		<label class="posting-toggle">
			<input type="checkbox" checked={playersCanPost} onchange={() => ontoggleposting?.()} />
			Players can post
		</label>
	{/if}
</header>

<style>
	.conversation-header {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		background: rgb(255 249 230 / 82%);
		padding: var(--space-3) var(--space-4);
	}

	.conversation-header .conversation-avatar {
		display: none;
	}

	.conversation-header h2,
	.conversation-header p {
		margin: 0;
	}

	.conversation-header h2 {
		font-size: 1.2rem;
	}

	.conversation-header p {
		color: var(--ink-soft);
		font-size: 0.8rem;
	}

	.back {
		display: none;
	}

	.conversation-avatar {
		display: grid;
		width: 2.8rem;
		height: 2.8rem;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--crimson-dark);
		color: var(--gold-light);
		font-family: var(--font-display);
		font-weight: 700;
	}

	.posting-toggle {
		display: flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-2);
		font-size: 0.82rem;
	}

	@media (max-width: 47.99rem) {
		.conversation-header {
			grid-template-columns: auto auto minmax(0, 1fr) auto;
			padding-inline: var(--space-2);
		}

		.back,
		.conversation-header .conversation-avatar {
			display: grid;
		}

		.posting-toggle {
			width: var(--target-size);
			overflow: hidden;
			font-size: 0;
		}
	}
</style>
