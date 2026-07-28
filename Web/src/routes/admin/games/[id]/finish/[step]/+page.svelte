<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import {
		Award,
		Check,
		ChevronLeft,
		ChevronRight,
		Flag,
		RotateCcw,
		Trophy,
		Users
	} from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Dialog from '$lib/components/Dialog.svelte';
	import GameSummaryCard from '$lib/components/GameSummaryCard.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import type { GameSummary, Participant } from '$lib/api/types';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	const steps = [
		{ id: 'outcomes', label: 'Outcomes', icon: Users },
		{ id: 'achievements', label: 'Achievements', icon: Award },
		{ id: 'summary', label: 'Summary', icon: Flag }
	] as const;

	let outcomes = $state<Record<string, Participant['outcome']>>({});
	let summary = $state<GameSummary | null>(null);
	let busy = $state(false);
	let archiveConfirmOpen = $state(false);

	const view = $derived(gameState.admin);
	const step = $derived(page.params.step as (typeof steps)[number]['id']);
	const activePlayers = $derived(
		view?.participants.filter((player) => !['kicked', 'left'].includes(player.status)) ?? []
	);

	$effect(() => {
		if (view) {
			for (const player of activePlayers) {
				if (!(player.id in outcomes)) outcomes[player.id] = player.outcome;
			}
		}
	});

	$effect(() => {
		if (step === 'summary' && view && !summary) void loadSummary();
	});

	onMount(() => {
		if (!steps.some((candidate) => candidate.id === step)) {
			void goto(resolve(`/admin/games/${page.params.id}/finish/outcomes`), { replaceState: true });
		}
	});

	function playerName(player: Participant) {
		return player.gameAlias || player.displayNameSnapshot;
	}

	async function saveOutcomes() {
		if (!view) return false;
		busy = true;
		try {
			await api(`/games/${view.game.id}/outcomes`, {
				method: 'PUT',
				...jsonBody({
					outcomes: activePlayers.map((player) => ({
						participantId: player.id,
						outcome: outcomes[player.id]
					}))
				})
			});
			await gameState.refreshAdmin(view.game.id);
			toasts.success('Outcomes saved.');
			return true;
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'Outcomes could not be saved.');
			return false;
		} finally {
			busy = false;
		}
	}

	function setWinners(ids: string[]) {
		const winners = new Set(ids);
		outcomes = Object.fromEntries(
			activePlayers.map((player) => [player.id, winners.has(player.id) ? 'win' : 'loss'])
		) as Record<string, Participant['outcome']>;
	}

	function setAll(outcome: Participant['outcome']) {
		outcomes = Object.fromEntries(activePlayers.map((player) => [player.id, outcome])) as Record<
			string,
			Participant['outcome']
		>;
	}

	async function nextFromOutcomes() {
		if (await saveOutcomes())
			await goto(resolve(`/admin/games/${page.params.id}/finish/achievements`));
	}

	async function toggleAchievement(profileId: string, achievementId: string) {
		if (!view) return;
		const existing = view.awards.find(
			(award) => award.profileId === profileId && award.achievementId === achievementId
		);
		busy = true;
		try {
			if (existing) {
				await api(`/games/${view.game.id}/achievement-awards/${existing.id}`, { method: 'DELETE' });
			} else {
				await api(`/games/${view.game.id}/achievement-awards`, {
					method: 'POST',
					...jsonBody({ profileId, achievementId, note: '' })
				});
			}
			await gameState.refreshAdmin(view.game.id);
		} catch (caught) {
			toasts.error(
				caught instanceof Error ? caught.message : 'The achievement could not be updated.'
			);
		} finally {
			busy = false;
		}
	}

	async function loadSummary() {
		try {
			summary = await api<GameSummary>(`/games/${page.params.id}/summary`);
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'The summary could not be loaded.', {
				actionLabel: 'Retry',
				action: loadSummary,
				persistent: true
			});
		}
	}

	async function returnToGame() {
		if (!view) return;
		busy = true;
		try {
			await api(`/games/${view.game.id}/completion/cancel`, { method: 'POST', ...jsonBody({}) });
			await gameState.refreshAdmin(view.game.id);
			await goto(resolve(`/admin/games/${view.game.id}/overview`));
			toasts.success('Returned to the game.');
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'The game could not be restored.');
		} finally {
			busy = false;
		}
	}

	async function archive() {
		if (!view) return;
		busy = true;
		try {
			await api(`/games/${view.game.id}/archive`, {
				method: 'POST',
				...jsonBody({ confirmUnsetOutcomes: true })
			});
			await gameState.refreshAdmin(view.game.id);
			archiveConfirmOpen = false;
			await goto(resolve(`/admin/games/${view.game.id}/summary`));
			toasts.success('Game finished and archived.');
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'The game could not be archived.');
		} finally {
			busy = false;
		}
	}
</script>

{#if view}
	<div class="completion">
		<header class="completion-heading">
			<p class="eyebrow">Finish game</p>
			<h1>{view.game.name}</h1>
			<p>Review the record before archiving it permanently.</p>
		</header>

		<nav aria-label="Completion steps">
			{#each steps as item, index (item.id)}
				{@const Icon = item.icon}
				<a
					class:active={step === item.id}
					aria-current={step === item.id ? 'step' : undefined}
					href={resolve(`/admin/games/${view.game.id}/finish/${item.id}`)}
				>
					<span>{index + 1}</span><Icon size={18} />
					{item.label}
				</a>
			{/each}
		</nav>

		<main>
			{#if step === 'outcomes'}
				<div class="step-heading">
					<div>
						<h2>Outcomes</h2>
						<p>Outcomes are independent from elimination status.</p>
					</div>
					<div class="bulk-actions">
						<button type="button" onclick={() => setAll('draw')}>Mark all draw</button>
						<button type="button" onclick={() => setAll('unset')}>Clear results</button>
					</div>
				</div>
				<div class="outcome-list">
					{#each activePlayers as player (player.id)}
						<article>
							<div class="avatar">{playerName(player).slice(0, 1).toUpperCase()}</div>
							<div>
								<h3>{playerName(player)}</h3>
								<p>Seat {player.seatNumber}</p>
							</div>
							<div class="outcome-choices">
								{#each ['win', 'loss', 'draw', 'unset'] as outcome (outcome)}
									<button
										type="button"
										class:active={outcomes[player.id] === outcome}
										aria-pressed={outcomes[player.id] === outcome}
										onclick={() => (outcomes[player.id] = outcome as Participant['outcome'])}
									>
										{outcome === 'unset' ? 'Unset' : outcome}
									</button>
								{/each}
							</div>
						</article>
					{/each}
				</div>
				<div class="winner-shortcut">
					<Trophy size={19} />
					<span>Quick winners:</span>
					{#each activePlayers as player (player.id)}
						<button type="button" onclick={() => setWinners([player.id])}
							>{playerName(player)}</button
						>
					{/each}
				</div>
			{:else if step === 'achievements'}
				<div class="step-heading">
					<div>
						<h2>Achievements</h2>
						<p>Award or revoke achievements without leaving the flow.</p>
					</div>
				</div>
				{#if view.ruleset.achievements.length === 0}
					<div class="empty">
						<Award size={38} />
						<h3>No achievements in this ruleset</h3>
					</div>
				{:else}
					<div class="award-players">
						{#each activePlayers as player (player.id)}
							<section>
								<header>
									<span class="avatar">{playerName(player).slice(0, 1).toUpperCase()}</span>
									<h3>{playerName(player)}</h3>
								</header>
								<div>
									{#each view.ruleset.achievements as achievement (achievement.id)}
										{@const awarded = view.awards.some(
											(award) =>
												award.profileId === player.profileId &&
												award.achievementId === achievement.id
										)}
										<button
											type="button"
											class:awarded
											onclick={() => toggleAchievement(player.profileId, achievement.id)}
										>
											<span class="check"
												>{#if awarded}<Check size={16} />{/if}</span
											>
											<span
												><strong>{achievement.name}</strong><small>{achievement.description}</small
												></span
											>
										</button>
									{/each}
								</div>
							</section>
						{/each}
					</div>
				{/if}
			{:else}
				<div class="step-heading">
					<div>
						<h2>Summary</h2>
						<p>Confirm the final game record.</p>
					</div>
				</div>
				{#if !summary}
					<p role="status">Loading summary…</p>
				{:else}
					<GameSummaryCard {summary} />
				{/if}
			{/if}
		</main>

		<footer class="completion-actions">
			<div>
				{#if step === 'outcomes'}
					<Button variant="ghost" loading={busy} onclick={returnToGame}
						><RotateCcw size={18} /> Return to game</Button
					>
				{:else}
					<Button
						variant="ghost"
						onclick={() =>
							goto(
								resolve(
									`/admin/games/${view.game.id}/finish/${step === 'summary' ? 'achievements' : 'outcomes'}`
								)
							)}
					>
						<ChevronLeft size={18} /> Back
					</Button>
					<Button variant="ghost" loading={busy} onclick={returnToGame}
						><RotateCcw size={18} /> Return to game</Button
					>
				{/if}
			</div>
			{#if step === 'outcomes'}
				<Button loading={busy} onclick={nextFromOutcomes}
					>Save and continue <ChevronRight size={18} /></Button
				>
			{:else if step === 'achievements'}
				<Button onclick={() => goto(resolve(`/admin/games/${view.game.id}/finish/summary`))}>
					Review summary <ChevronRight size={18} />
				</Button>
			{:else}
				<Button onclick={() => (archiveConfirmOpen = true)}
					><Flag size={18} /> Finish and archive</Button
				>
			{/if}
		</footer>
	</div>
{/if}

<Dialog
	open={archiveConfirmOpen}
	title="Finish and archive?"
	description="The game summary will remain viewable, but the game and chat will become read-only."
	close={() => (archiveConfirmOpen = false)}
>
	<p>Check outcomes and achievements before continuing.</p>
	{#snippet actions()}
		<Button variant="ghost" onclick={() => (archiveConfirmOpen = false)}>Cancel</Button>
		<Button loading={busy} onclick={archive}>Finish and archive</Button>
	{/snippet}
</Dialog>

<style>
	.completion {
		width: min(100%, 70rem);
		margin-inline: auto;
	}

	.completion-heading {
		margin-block-end: var(--space-5);
		text-align: center;
	}

	.completion-heading h1,
	.completion-heading p {
		margin: 0;
	}

	.eyebrow {
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
		letter-spacing: 0.14em;
		text-transform: uppercase;
	}

	nav {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		border-block: var(--border-strong);
	}

	nav a {
		display: flex;
		min-height: 3.5rem;
		align-items: center;
		justify-content: center;
		gap: var(--space-2);
		border-block-end: 3px solid transparent;
		color: var(--ink-soft);
		font-family: var(--font-display);
		font-size: 0.74rem;
		font-weight: 700;
		text-decoration: none;
		text-transform: uppercase;
	}

	nav a.active {
		border-color: var(--crimson);
		color: var(--crimson-dark);
	}

	nav a > span {
		display: grid;
		width: 1.5rem;
		height: 1.5rem;
		place-items: center;
		border: 1px solid currentColor;
		border-radius: 50%;
	}

	main {
		min-height: 26rem;
		padding-block: var(--space-5);
	}

	.step-heading {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-end: var(--border-strong);
		padding-block-end: var(--space-3);
	}

	.step-heading h2,
	.step-heading p {
		margin: 0;
	}

	.bulk-actions {
		display: flex;
		gap: var(--space-2);
	}

	.bulk-actions button,
	.winner-shortcut button {
		min-height: var(--target-size);
		border: 0;
		background: transparent;
		color: var(--crimson-dark);
		cursor: pointer;
		font-family: var(--font-display);
		font-size: 0.7rem;
		font-weight: 700;
	}

	.outcome-list article {
		display: grid;
		grid-template-columns: auto minmax(10rem, 1fr) minmax(18rem, auto);
		align-items: center;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding: var(--space-3) 0;
	}

	.avatar {
		display: grid;
		width: 2.7rem;
		height: 2.7rem;
		place-items: center;
		border: 2px double var(--gold);
		border-radius: 50%;
		background: var(--ink);
		color: var(--gold-light);
		font-family: var(--font-display);
		font-weight: 700;
	}

	.outcome-list h3,
	.outcome-list p {
		margin: 0;
	}

	.outcome-choices {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: var(--space-1);
	}

	.outcome-choices button {
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		color: var(--ink);
		cursor: pointer;
		text-transform: capitalize;
	}

	.outcome-choices button.active {
		border: 2px solid var(--crimson-dark);
		background: color-mix(in srgb, var(--crimson) 10%, transparent);
		font-weight: 700;
	}

	.winner-shortcut {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-1);
		margin-block-start: var(--space-3);
	}

	.award-players {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 20rem), 1fr));
		gap: var(--space-4);
		margin-block-start: var(--space-4);
	}

	.award-players > section {
		border: var(--border-subtle);
		background: rgb(255 249 230 / 58%);
	}

	.award-players header {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		border-block-end: var(--border-subtle);
		padding: var(--space-3);
	}

	.award-players header h3 {
		margin: 0;
	}

	.award-players button {
		display: grid;
		width: 100%;
		min-height: 4rem;
		grid-template-columns: auto 1fr;
		align-items: center;
		gap: var(--space-2);
		border: 0;
		border-block-end: var(--border-subtle);
		background: transparent;
		color: var(--ink);
		cursor: pointer;
		padding: var(--space-2) var(--space-3);
		text-align: start;
	}

	.award-players button.awarded {
		background: rgb(49 91 58 / 8%);
		color: var(--success);
	}

	.check {
		display: grid;
		width: 1.5rem;
		height: 1.5rem;
		place-items: center;
		border: 1px solid currentColor;
	}

	.award-players strong,
	.award-players small {
		display: block;
	}

	.award-players small {
		color: var(--ink-soft);
	}

	.completion-actions {
		position: sticky;
		z-index: var(--layer-sticky);
		inset-block-end: 0;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		border-block-start: var(--border-strong);
		background: rgb(247 231 196 / 96%);
		padding: var(--space-3) 0 max(var(--space-3), env(safe-area-inset-bottom));
	}

	.completion-actions > div {
		display: flex;
		gap: var(--space-1);
	}

	.empty {
		padding: var(--space-7);
		text-align: center;
	}

	@media (max-width: 47.99rem) {
		nav a {
			flex-direction: column;
			gap: 0.1rem;
			font-size: 0.62rem;
		}

		nav a :global(svg) {
			display: none;
		}

		.step-heading {
			align-items: stretch;
			flex-direction: column;
		}

		.outcome-list article {
			grid-template-columns: auto 1fr;
		}

		.outcome-choices {
			grid-column: 1 / -1;
		}

		.completion-actions {
			align-items: stretch;
			flex-direction: column-reverse;
		}

		.completion-actions > :global(button),
		.completion-actions > div {
			width: 100%;
		}

		.completion-actions > div :global(button) {
			flex: 1;
		}
	}
</style>
