<script lang="ts">
	import { UserRound } from '@lucide/svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import ProtectedMedia from '$lib/features/media/components/ProtectedMedia.svelte';
	import { api } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { Profile } from '$lib/api/types';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	type Achievement = {
		id: string;
		title: string;
		description: string;
		points: number;
		awardedAt: string;
	};
	type PartyProfile = Profile & {
		statistics: Record<string, number>;
		achievements: Achievement[];
	};

	let profile = $state<PartyProfile | null>(null);
	const view = $derived(gameState.player);

	async function openProfile(profileId: string) {
		try {
			profile = await api<PartyProfile>(`/profiles/${profileId}/summary`);
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The player profile could not be loaded.'));
		}
	}
</script>

{#if view}
	<div class="party-page">
		<header>
			<p class="eyebrow">Current game</p>
			<h1>Party</h1>
			<p>{view.party.length} players</p>
		</header>
		<div class="party-list">
			{#each view.party as member (member.id)}
				<button type="button" onclick={() => openProfile(member.profileId)}>
					<span class="avatar"
						>{(member.gameAlias || member.displayName).slice(0, 1).toUpperCase()}</span
					>
					<span>
						<strong>{member.gameAlias || member.displayName}</strong>
						<small>{member.gameAlias ? member.displayName : `Seat ${member.seatNumber}`}</small>
					</span>
					<i class:eliminated={member.status === 'eliminated'}>{member.status}</i>
				</button>
			{/each}
		</div>
	</div>
{/if}

<Dialog
	open={profile !== null}
	title={profile?.displayName ?? 'Player profile'}
	close={() => (profile = null)}
>
	{#if profile}
		<div class="profile">
			<div class="profile-avatar">
				{#if profile.avatar}
					<ProtectedMedia src={profile.avatar} kind="image" alt="" />
				{:else}
					<UserRound size={38} />
				{/if}
			</div>
			{#if profile.bio}<p>{profile.bio}</p>{/if}
			<dl>
				{#each Object.entries(profile.statistics) as [label, value] (label)}
					<div>
						<dt>{label.replaceAll('_', ' ')}</dt>
						<dd>{value}</dd>
					</div>
				{/each}
			</dl>
			{#if profile.achievements.length > 0}
				<section>
					<h3>Achievements</h3>
					{#each profile.achievements as achievement (achievement.id)}
						<article>
							<strong>{achievement.title}</strong>
							<p>{achievement.description}</p>
						</article>
					{/each}
				</section>
			{/if}
		</div>
	{/if}
</Dialog>

<style>
	.party-page {
		width: min(100%, 48rem);
		margin-inline: auto;
		padding: clamp(var(--space-4), 5vw, var(--space-6));
	}

	header {
		border-block-end: var(--border-strong);
		padding-block-end: var(--space-3);
	}

	header h1,
	header p {
		margin: 0;
	}

	.party-list > button {
		display: grid;
		width: 100%;
		min-height: 4.5rem;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: var(--space-3);
		border: 0;
		border-block-end: var(--border-subtle);
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		text-align: start;
	}

	.avatar {
		display: grid;
		width: 3rem;
		height: 3rem;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--ink);
		color: var(--gold-light);
		font-family: var(--font-display);
		font-weight: 700;
	}

	.party-list strong,
	.party-list small {
		display: block;
	}

	.party-list small {
		color: var(--ink-soft);
	}

	.party-list i {
		border: 1px solid var(--success);
		color: var(--success);
		font-size: 0.72rem;
		font-style: normal;
		padding: 0.15rem 0.4rem;
		text-transform: capitalize;
	}

	.party-list i.eliminated {
		border-color: var(--danger);
		color: var(--danger);
	}

	.profile {
		display: grid;
		gap: var(--space-4);
	}

	.profile-avatar {
		display: grid;
		width: 6rem;
		height: 6rem;
		overflow: hidden;
		place-items: center;
		border: 3px double var(--gold);
		border-radius: 50%;
		background: var(--ink);
		color: var(--gold-light);
		margin-inline: auto;
	}

	.profile-avatar :global(img) {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	dl div {
		display: flex;
		justify-content: space-between;
		border-block-end: var(--border-subtle);
		padding: var(--space-2) 0;
	}

	dt {
		text-transform: capitalize;
	}

	dd {
		margin: 0;
		font-weight: 700;
	}

	.profile article {
		border-inline-start: 3px solid var(--gold);
		padding-inline-start: var(--space-3);
	}

	.profile article p {
		margin: 0;
	}
</style>
