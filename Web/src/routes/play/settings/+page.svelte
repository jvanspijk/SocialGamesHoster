<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { ArrowLeft, History, LogOut, UserRound, Wifi } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import DisplayPreferencesSettings from '$lib/features/settings/components/DisplayPreferencesSettings.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import ToggleSetting from '$lib/components/ToggleSetting.svelte';
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
		title="Settings"
		description="Accessibility and device preferences apply across every game."
		variant="flush"
	>
		{#snippet actions()}
			<nav aria-label="Account pages">
				<a href={resolve('/play')}><ArrowLeft size={18} /> Player home</a>
				<a href={resolve('/play/profile')}><UserRound size={18} /> Profile</a>
				<a href={resolve('/play/history')}><History size={18} /> History</a>
			</nav>
		{/snippet}
	</PageHeading>

	<Panel title="Sound">
		<ToggleSetting
			title="Game sounds"
			name="game-sounds"
			checked={sound.enabled}
			onchange={sound.set}
		/>
	</Panel>

	<Panel title="Display">
		<DisplayPreferencesSettings />
	</Panel>

	<Panel title="Connection">
		<div class="connection-row">
			<Wifi size={21} />
			<span>Connected to this game</span>
		</div>
	</Panel>

	<Panel title="Account">
		<p>
			Signing out removes this profile from the current device. Your game history stays with this
			game.
		</p>
		<Button variant="danger" onclick={signOut}><LogOut size={18} /> Sign out</Button>
	</Panel>
</div>

<style>
	.account-page {
		display: grid;
		width: min(100%, 36rem);
		gap: var(--space-4);
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
		color: var(--action-dark);
		font-family: var(--font-display);
		font-size: var(--font-size-sm);
		font-weight: 700;
		text-decoration: none;
	}

	.connection-row {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}
</style>
