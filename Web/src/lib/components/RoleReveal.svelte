<script lang="ts">
	import { ArrowLeft, Eye, EyeOff, Shield } from '@lucide/svelte';
	import Button from './Button.svelte';
	import ProtectedMedia from './ProtectedMedia.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { PlayerGameView } from '$lib/api/types';
	import { gameState } from '$lib/state/game.svelte';
	import { toasts } from '$lib/state/toasts.svelte';

	let {
		view,
		revealed,
		reveal,
		hide,
		back
	}: {
		view: PlayerGameView;
		revealed: boolean;
		reveal: () => void;
		hide: () => void;
		back: () => void;
	} = $props();

	const role = $derived(view.role);
	const roleAsset = $derived(
		role?.imageAssetKey
			? view.assets.find((asset) => asset.kind === 'image' && asset.assetKey === role.imageAssetKey)
			: undefined
	);
	let abilityBusy = $state('');

	function choiceFor(abilityID: string) {
		return (view.abilityChoices ?? []).find((choice) => choice.abilityId === abilityID);
	}

	async function activateAbility(abilityID: string) {
		abilityBusy = abilityID;
		try {
			await api(`/games/${view.game.id}/abilities/${abilityID}/activate`, {
				method: 'POST',
				...jsonBody({})
			});
			await gameState.refreshPlayer();
			toasts.success('Ability activated.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The ability could not be activated.'));
		} finally {
			abilityBusy = '';
		}
	}

	async function undoAbility(abilityID: string) {
		abilityBusy = abilityID;
		try {
			await api(`/games/${view.game.id}/abilities/${abilityID}/activate`, { method: 'DELETE' });
			await gameState.refreshPlayer();
			toasts.success('Ability activation undone.');
		} catch (caught) {
			toasts.error(errorMessage(caught, 'The activation could not be undone.'));
		} finally {
			abilityBusy = '';
		}
	}

	function knowledgeText(item: Record<string, unknown>) {
		const name = String(item.displayName ?? `Seat ${item.seatNumber ?? ''}`).trim();
		if (item.role && typeof item.role === 'object') {
			const roleName = String((item.role as { name?: unknown }).name ?? 'role known');
			return `${name}: ${roleName}`;
		}
		if (item.teamId) return `${name}: team ${String(item.teamId)}`;
		if (item.status) return `${name}: ${String(item.status)}`;
		return name;
	}
</script>

{#if !view.roleAvailable || !role}
	<section class="unavailable">
		<Shield size={42} strokeWidth={1.4} />
		<h1>Role unavailable</h1>
		<p>The game master has not made roles available.</p>
		<Button variant="secondary" onclick={back}><ArrowLeft size={18} /> Return to game</Button>
	</section>
{:else if !revealed}
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
	<article class="role-reveal">
		<header class:has-art={roleAsset}>
			{#if roleAsset}
				<div class="role-art"><ProtectedMedia src={roleAsset.preview} kind="image" alt="" /></div>
			{:else}
				<div class="role-fallback" aria-hidden="true">
					<span>{role.name.slice(0, 1).toUpperCase()}</span>
				</div>
			{/if}
			<div class="hero-gradient"></div>
			<div class="hero-copy">
				{#if role.team}<p>{role.team.name}</p>{/if}
				<h1>{role.name}</h1>
				<p class="description">{role.description}</p>
			</div>
		</header>

		<div class="role-content">
			<section>
				<h2>How to win</h2>
				<p>{role.winCondition || 'Follow the win condition provided by the game master.'}</p>
			</section>

			<section>
				<h2>Abilities</h2>
				{#if role.abilities.length === 0}
					<p>No special abilities.</p>
				{:else}
					<div class="ability-list">
						{#each role.abilities as ability (ability.id)}
							{@const choice = choiceFor(ability.id)}
							<article>
								<h3>{ability.name}</h3>
								<p>{ability.description}</p>
								{#if choice}
									<p class:finalized={choice.status === 'Finalized'} class="ability-state">
										{choice.status}
									</p>
									{#if choice.status === 'Activated'}
										<Button
											variant="secondary"
											loading={abilityBusy === ability.id}
											onclick={() => undoAbility(ability.id)}>Undo activation</Button
										>
									{/if}
								{:else if (ability.activationPhaseIds ?? []).includes(view.game.phaseKey) && !view.game.abilityPhaseLockedAt && view.participant.status === 'active'}
									<Button
										loading={abilityBusy === ability.id}
										onclick={() => activateAbility(ability.id)}>Activate</Button
									>
								{:else if view.game.abilityPhaseLockedAt && (ability.activationPhaseIds ?? []).includes(view.game.phaseKey)}
									<p class="ability-state finalized">Choices finalized</p>
								{:else}
									<p class="ability-unavailable">Not playable in this phase</p>
								{/if}
							</article>
						{/each}
					</div>
				{/if}
			</section>

			<section>
				<h2>Information you know</h2>
				{#if view.knowledge.length === 0}
					<p>No additional information.</p>
				{:else}
					<ul>
						{#each view.knowledge as item, index (`${item.participantId ?? item.seatNumber ?? index}`)}
							<li>{knowledgeText(item)}</li>
						{/each}
					</ul>
				{/if}
			</section>
		</div>
		<div class="hide-control">
			<Button onclick={hide}><EyeOff size={19} /> Hide role</Button>
		</div>
	</article>
{/if}

<style>
	.unavailable,
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

	.unavailable h1,
	.unavailable p,
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

	.role-reveal {
		width: min(100%, 52rem);
		min-height: calc(100dvh - 7.75rem);
		margin-inline: auto;
		border-inline: var(--border-strong);
		background: var(--paper);
		box-shadow: var(--shadow);
	}

	.role-reveal > header {
		position: relative;
		display: grid;
		min-height: clamp(20rem, 54vh, 34rem);
		overflow: hidden;
		background: linear-gradient(145deg, var(--navy), var(--ink));
		color: var(--paper-light);
		isolation: isolate;
	}

	.role-art,
	.role-fallback,
	.hero-gradient,
	.hero-copy {
		grid-area: 1 / 1;
	}

	.role-art {
		position: absolute;
		inset: 0;
	}

	.role-art :global(img) {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.role-fallback {
		display: grid;
		place-items: center;
		background:
			radial-gradient(circle at 50% 40%, rgb(223 189 101 / 20%), transparent 26%),
			linear-gradient(135deg, var(--navy), var(--crimson-dark) 62%, var(--ink));
	}

	.role-fallback::before,
	.role-fallback::after {
		position: absolute;
		inset: 2rem;
		border: 2px solid rgb(223 189 101 / 35%);
		content: '';
		transform: rotate(4deg);
	}

	.role-fallback::after {
		transform: rotate(-4deg);
	}

	.role-fallback span {
		color: var(--gold-light);
		font-family: var(--font-display);
		font-size: clamp(9rem, 35vw, 17rem);
		opacity: 0.55;
	}

	.hero-gradient {
		z-index: 1;
		background: linear-gradient(
			to bottom,
			transparent 25%,
			rgb(16 10 7 / 28%) 52%,
			rgb(16 10 7 / 94%) 100%
		);
	}

	.hero-copy {
		z-index: 2;
		align-self: end;
		padding: clamp(var(--space-5), 7vw, var(--space-7));
	}

	.hero-copy h1,
	.hero-copy p {
		margin: 0;
		color: var(--paper-light);
		text-shadow: 0 2px 6px rgb(0 0 0 / 70%);
	}

	.hero-copy h1 {
		font-size: clamp(2.5rem, 10vw, 5.5rem);
	}

	.hero-copy > p:first-child {
		color: var(--gold-light);
		font-family: var(--font-display);
		font-size: 0.78rem;
		font-weight: 700;
		letter-spacing: 0.15em;
		text-transform: uppercase;
	}

	.description {
		max-width: 40rem;
		font-size: clamp(1rem, 3vw, 1.25rem);
	}

	.role-content {
		display: grid;
		gap: var(--space-6);
		padding: clamp(var(--space-4), 6vw, var(--space-7));
		padding-block-end: 7rem;
	}

	.role-content > section {
		border-block-start: 2px solid var(--gold-dark);
		padding-block-start: var(--space-3);
	}

	.role-content h2,
	.role-content h3,
	.role-content p {
		margin-block-start: 0;
	}

	.ability-list {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 15rem), 1fr));
		gap: var(--space-4);
	}

	.ability-list article {
		border-inline-start: 3px solid var(--crimson);
		padding-inline-start: var(--space-3);
	}

	.ability-state {
		color: var(--crimson-dark);
		font-family: var(--font-display);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	.ability-state.finalized {
		color: var(--success);
	}

	.ability-unavailable {
		color: var(--ink-faint);
		font-size: 0.82rem;
	}

	.role-content li {
		margin-block-end: var(--space-2);
	}

	.hide-control {
		position: fixed;
		z-index: var(--layer-sticky);
		inset-inline: 0;
		inset-block-end: calc(4rem + env(safe-area-inset-bottom));
		display: flex;
		justify-content: center;
		background: linear-gradient(transparent, var(--paper) 38%);
		padding: var(--space-5) var(--space-3) var(--space-3);
	}

	.hide-control :global(button) {
		width: min(100%, 24rem);
	}

	@media (min-width: 64rem) {
		.hide-control {
			inset-inline-start: 13rem;
			inset-block-end: 0;
		}
	}
</style>
