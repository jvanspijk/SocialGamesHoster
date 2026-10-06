<script lang="ts">
	import { resolve } from '$app/paths';
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import History from '@lucide/svelte/icons/history';
	import UserRound from '@lucide/svelte/icons/user-round';
	import DisplayPreferencesSettings from '$lib/features/settings/components/DisplayPreferencesSettings.svelte';
	import VersionMention from '$lib/features/settings/components/VersionMention.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import ToggleSetting from '$lib/components/ToggleSetting.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import { sound } from '$lib/state/sound.svelte';
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

	<VersionMention />
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
</style>
