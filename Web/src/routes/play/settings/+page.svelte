<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { ArrowLeft, LogOut, UserRound, Volume2, VolumeX, Wifi } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import DisplayPreferencesSettings from '$lib/features/settings/components/DisplayPreferencesSettings.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import { api } from '$lib/api/client';
	import { auth } from '$lib/state/auth.svelte';
	import { gameState } from '$lib/state/game.svelte';
	import { sound } from '$lib/state/sound.svelte';

	async function signOut() {
		try {
			await api('/auth/logout', { method: 'POST' });
		} finally {
			gameState.clear();
			auth.clear();
			await goto(resolve('/'));
		}
	}
</script>

<div class="account-page">
	<PageHeading
		eyebrow="Player account"
		title="Settings"
		description="Accessibility and device preferences apply across every game."
		variant="flush"
	>
		{#snippet actions()}
			<nav aria-label="Account pages">
				{#if gameState.player}<a href={resolve('/play')}><ArrowLeft size={18} /> Return to game</a
					>{/if}
				<a href={resolve('/play/profile')}><UserRound size={18} /> Profile</a>
			</nav>
		{/snippet}
	</PageHeading>

	<Panel title="Sound">
		<button class="setting-row" type="button" onclick={() => sound.toggle()}>
			{#if sound.enabled}<Volume2 size={22} />{:else}<VolumeX size={22} />{/if}
			<span><strong>Game sounds</strong><small>{sound.enabled ? 'On' : 'Off'}</small></span>
		</button>
	</Panel>

	<Panel title="Display" description="These preferences are saved on this device.">
		<DisplayPreferencesSettings />
	</Panel>

	<Panel title="Connection">
		<div class="connection-row">
			<Wifi size={21} />
			<span>Connected to this game host</span>
		</div>
	</Panel>

	<Panel title="Account">
		<p>
			Signing out removes this profile from the current device. Your game history stays on the host.
		</p>
		<Button variant="danger" onclick={signOut}><LogOut size={18} /> Sign out</Button>
	</Panel>
</div>

<style>
	.account-page {
		display: grid;
		width: min(100%, 48rem);
		gap: var(--space-5);
		margin-inline: auto;
		padding: clamp(var(--space-4), 5vw, var(--space-6));
	}

	nav {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	nav a {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-1);
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		text-decoration: none;
		text-transform: uppercase;
	}

	.setting-row {
		display: grid;
		width: 100%;
		min-height: var(--target-size);
		grid-template-columns: auto 1fr;
		align-items: center;
		gap: var(--space-3);
		border: 0;
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		text-align: start;
	}

	.setting-row strong,
	.setting-row small {
		display: block;
	}

	.setting-row small {
		color: var(--ink-soft);
	}

	.connection-row {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
</style>
