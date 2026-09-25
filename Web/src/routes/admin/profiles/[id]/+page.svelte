<script lang="ts">
	import { onMount } from 'svelte';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { ArrowLeft, UserRound } from '@lucide/svelte';
	import LoadingState from '$lib/components/LoadingState.svelte';
	import PageHeading from '$lib/components/PageHeading.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import ProtectedMedia from '$lib/features/media/components/ProtectedMedia.svelte';
	import { api } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { Profile } from '$lib/api/types';

	type AdminProfile = Profile & {
		normalizedName: string;
		approvedAt: string;
	};

	type ProfileHistory = {
		profile: AdminProfile;
		games: Array<{
			id: string;
			name: string;
			rulesetName: string;
			roleName: string;
			outcome: string;
			endedAt?: string;
			achievements: Array<{ id: string; title: string }>;
		}>;
		statistics: { achievementCount: number; achievementPoints: number };
	};

	let detail = $state<ProfileHistory | null>(null);
	let loadError = $state('');

	onMount(() => {
		void load();
	});

	async function load() {
		loadError = '';
		try {
			detail = await api<ProfileHistory>(`/admin/profiles/${page.params.id}`);
		} catch (caught) {
			loadError = errorMessage(caught, 'The profile could not be loaded.');
		}
	}

	function formatDate(value: string) {
		return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(value));
	}
</script>

<div class="profile-layout">
	<PageHeading
		eyebrow="Player profile"
		title={detail?.profile.displayName ?? 'Profile'}
		description="Profile details and completed game history."
		variant="flush"
	>
		{#snippet actions()}
			<a class="back" href={resolve('/admin/approvals')}><ArrowLeft size={18} /> Back to profiles</a
			>
		{/snippet}
	</PageHeading>

	{#if detail}
		<div class="profile-page">
			<Panel title="Profile details" variant="focal">
				<div class="profile-overview">
					<div class="profile-summary">
						<div class="profile-header">
							<div class="avatar" aria-hidden="true">
								{#if detail.profile.avatar}
									<ProtectedMedia src={detail.profile.avatar} kind="image" alt="" />
								{:else}
									<UserRound size={34} />
								{/if}
							</div>
							<div>
								<h2>{detail.profile.displayName}</h2>
								<StatusBadge
									label={detail.profile.active ? 'Active' : 'Disabled'}
									tone={detail.profile.active ? 'success' : 'danger'}
								/>
							</div>
						</div>

						{#if detail.profile.bio}<p class="bio">{detail.profile.bio}</p>{/if}
					</div>

					<dl>
						<div>
							<dt>Created at</dt>
							<dd>{formatDate(detail.profile.approvedAt)}</dd>
						</div>
						<div>
							<dt>Profile name</dt>
							<dd>{detail.profile.normalizedName}</dd>
						</div>
					</dl>
				</div>
			</Panel>

			<Panel
				title="Game history"
				description={`${detail.statistics.achievementPoints} achievement points across ${detail.statistics.achievementCount} achievements`}
			>
				{#if detail.games.length === 0}
					<p class="empty">No completed games yet.</p>
				{:else}
					<div class="history">
						{#each detail.games as game (game.id)}
							<article>
								<div>
									<h3>{game.name}</h3>
									<p>{game.rulesetName} · {game.roleName || 'No role'}</p>
									{#if game.achievements.length > 0}
										<p class="achievements">
											{game.achievements.map((achievement) => achievement.title).join(' · ')}
										</p>
									{/if}
								</div>
								<strong>{game.outcome || 'No outcome'}</strong>
							</article>
						{/each}
					</div>
				{/if}
			</Panel>
		</div>
	{:else if loadError}
		<section class="load-failure" role="alert">
			<p>{loadError}</p>
			<button type="button" onclick={load}>Try again</button>
		</section>
	{:else}
		<LoadingState label="Loading profile…" />
	{/if}
</div>

<style>
	.profile-layout {
		width: min(100%, 70rem);
		margin-inline: auto;
	}

	.profile-page {
		display: grid;
		gap: var(--space-5);
		margin-block-start: var(--space-5);
	}

	.back {
		display: inline-flex;
		min-height: var(--target-size);
		align-items: center;
		gap: var(--space-1);
		color: var(--action-dark);
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		text-decoration: none;
	}

	.profile-header {
		display: flex;
		align-items: center;
		gap: var(--space-3);
	}

	.profile-overview {
		display: grid;
		grid-template-columns: minmax(0, 1.5fr) minmax(15rem, 0.85fr);
		gap: var(--space-6);
		align-items: start;
	}

	.avatar {
		display: grid;
		width: 4rem;
		height: 4rem;
		overflow: hidden;
		flex: 0 0 auto;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--ink);
		color: var(--gold-light);
	}

	.avatar :global(img) {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.profile-header h2,
	.bio,
	.history h3,
	.history p,
	.empty {
		margin: 0;
	}

	.bio {
		margin-block: var(--space-4);
		white-space: pre-wrap;
	}

	dl {
		display: grid;
		grid-template-columns: 1fr;
		gap: var(--space-3);
		margin: var(--space-4) 0 0;
	}

	dl div {
		border-block-start: var(--border-subtle);
		padding-block-start: var(--space-2);
	}

	dt {
		color: var(--ink-soft);
		font-family: var(--font-display);
		font-size: 0.68rem;
		font-weight: 700;
		letter-spacing: 0.08em;
	}

	dd {
		margin: var(--space-1) 0 0;
	}

	.history article {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding: var(--space-3) 0;
	}

	.history article:first-child {
		padding-block-start: 0;
	}

	.history p,
	.empty {
		color: var(--ink-soft);
	}

	.achievements {
		margin-block-start: var(--space-1) !important;
		font-size: 0.88rem;
	}

	.history strong {
		text-align: end;
		text-transform: capitalize;
	}

	.load-failure {
		padding: var(--space-5);
		text-align: center;
	}

	.load-failure p {
		margin: 0 0 var(--space-3);
	}

	.load-failure button {
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		color: var(--action-dark);
		cursor: pointer;
		font-family: var(--font-display);
		font-weight: 700;
		padding-inline: var(--space-3);
	}

	@media (max-width: 47.99rem) {
		.profile-overview {
			grid-template-columns: 1fr;
			gap: var(--space-4);
		}
	}
</style>
