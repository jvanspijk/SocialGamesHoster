<script lang="ts">
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import Lock from '@lucide/svelte/icons/lock';
	import LockOpen from '@lucide/svelte/icons/lock-open';
	import Button from '$lib/components/Button.svelte';
	import IconButton from '$lib/components/IconButton.svelte';

	let {
		label,
		initial,
		typeLabel,
		playersCanPost,
		canModerate = false,
		archived = false,
		postingDisabled = false,
		onback,
		ontoggleposting
	}: {
		label: string;
		initial: string;
		typeLabel: string;
		playersCanPost: boolean;
		canModerate?: boolean;
		archived?: boolean;
		postingDisabled?: boolean;
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
	<div class="conversation-title">
		<h2>{label}</h2>
		<p>
			{typeLabel}
			<span class:concealed={playersCanPost} aria-hidden={playersCanPost}>· Players read-only</span>
		</p>
	</div>
	{#if canModerate && !archived}
		<div class="posting-action">
			<Button disabled={postingDisabled} onclick={() => ontoggleposting?.()}>
				{#if playersCanPost}<Lock size={18} />{:else}<LockOpen size={18} />{/if}
				<span class="posting-label">
					<span class:concealed={!playersCanPost} aria-hidden={!playersCanPost}>Lock channel</span>
					<span class:concealed={playersCanPost} aria-hidden={playersCanPost}>Unlock channel</span>
				</span>
			</Button>
		</div>
	{/if}
</header>

<style>
	.conversation-header {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		background: var(--surface-raised);
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
		font-size: var(--font-size-lg);
		overflow-wrap: anywhere;
	}

	.conversation-header p {
		color: var(--text-secondary);
		font-size: var(--font-size-sm);
	}

	.back {
		display: none;
	}

	.conversation-avatar {
		display: grid;
		width: 2.8rem;
		height: 2.8rem;
		place-items: center;
		border: 2px double var(--accent);
		border-radius: 50%;
		background: var(--action-dark);
		color: var(--accent-light);
		font-family: var(--font-display);
		font-weight: 700;
	}

	.conversation-title {
		min-width: 0;
	}

	.posting-label {
		display: grid;
	}

	.posting-label span {
		grid-area: 1 / 1;
	}

	.concealed {
		visibility: hidden;
	}

	@media (max-width: 47.99rem) {
		.conversation-header {
			grid-template-columns: auto auto minmax(0, 1fr);
			padding-inline: var(--space-2);
		}

		.back,
		.conversation-header .conversation-avatar {
			display: grid;
		}

		.posting-action {
			grid-column: 2 / -1;
			justify-self: start;
		}
	}
</style>
