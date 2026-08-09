<script lang="ts">
	import { ArrowLeft, Eye, EyeOff } from '@lucide/svelte';
	import type { Snippet } from 'svelte';
	import Button from '$lib/components/Button.svelte';
	import ActionDock from './ActionDock.svelte';

	let {
		revealed,
		reveal,
		hide,
		back,
		children
	}: {
		revealed: boolean;
		reveal: () => void;
		hide: () => void;
		back: () => void;
		children: Snippet;
	} = $props();
</script>

{#if !revealed}
	<section class="concealed">
		<div class="conceal-mark" aria-hidden="true"><EyeOff size={46} strokeWidth={1.4} /></div>
		<p class="eyebrow">Private screen</p>
		<h1>Your role is hidden</h1>
		<p>Make sure only you can see the screen.</p>
		<div class="conceal-actions">
			<Button onclick={reveal}><Eye size={19} /> Reveal role</Button>
			<Button variant="ghost" onclick={back}><ArrowLeft size={18} /> Return to game</Button>
		</div>
	</section>
{:else}
	{@render children()}
	<ActionDock>
		<Button onclick={hide}><EyeOff size={19} /> Hide role</Button>
	</ActionDock>
{/if}

<style>
	.concealed {
		display: grid;
		width: min(100%, 34rem);
		min-height: calc(100dvh - 7.75rem);
		place-content: center;
		justify-items: center;
		margin-inline: auto;
		padding: var(--space-5);
		text-align: center;
	}

	.concealed h1,
	.concealed p {
		margin: 0;
	}

	.conceal-mark {
		display: grid;
		width: 7rem;
		height: 7rem;
		place-items: center;
		border: 4px double var(--gold-light);
		border-radius: 50%;
		background:
			radial-gradient(circle at 35% 28%, rgb(255 255 255 / 13%), transparent 30%),
			var(--crimson-dark);
		box-shadow: var(--shadow);
		color: var(--paper-light);
		margin-block-end: var(--space-4);
	}

	.conceal-actions {
		display: grid;
		width: min(100%, 20rem);
		gap: var(--space-2);
		margin-block-start: var(--space-5);
	}
</style>
