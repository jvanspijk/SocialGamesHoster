<script lang="ts">
	import { ArrowLeft, Shield } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import Panel from '$lib/components/Panel.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import PrivateDisclosure from './PrivateDisclosure.svelte';
	import RoleHero, { type RoleHeroPresentation } from './RoleHero.svelte';

	export type AbilityAction = 'activate' | 'undo';

	export type RoleAbilityPresentation = {
		id: string;
		name: string;
		description: string;
		status?: { label: string; tone: 'success' | 'info' };
		action?: { label: string; command: AbilityAction; variant?: 'primary' | 'secondary' };
		unavailableLabel?: string;
	};

	export type RoleKnowledgePresentation = { id: string; text: string };
	export type RoleRevealPresentation = RoleHeroPresentation & { winCondition: string };

	let {
		available,
		role,
		roleAsset,
		knowledge,
		abilities,
		revealed,
		busyAbilityId,
		reveal,
		hide,
		back,
		onAbilityAction
	}: {
		available: boolean;
		role: RoleRevealPresentation | null;
		roleAsset?: string;
		knowledge: RoleKnowledgePresentation[];
		abilities: RoleAbilityPresentation[];
		revealed: boolean;
		busyAbilityId: string;
		reveal: () => void;
		hide: () => void;
		back: () => void;
		onAbilityAction: (abilityID: string, action: AbilityAction) => void;
	} = $props();
</script>

{#if !available || !role}
	<section class="unavailable">
		<Panel variant="focal">
			<div class="unavailable-copy">
				<Shield size={42} strokeWidth={1.4} aria-hidden="true" />
				<h1>Role unavailable</h1>
				<p>The game master has not made roles available.</p>
				<Button variant="secondary" onclick={back}><ArrowLeft size={18} /> Return to game</Button>
			</div>
		</Panel>
	</section>
{:else}
	<PrivateDisclosure {revealed} {reveal} {hide} {back}>
		<article class="role-reveal">
			<RoleHero {role} asset={roleAsset} />

			<div class="role-content">
				<Panel title="How to win">
					<p>{role.winCondition || 'Follow the win condition provided by the game master.'}</p>
				</Panel>

				<Panel title="Abilities">
					{#if abilities.length === 0}
						<p>No special abilities.</p>
					{:else}
						<div class="ability-list">
							{#each abilities as ability (ability.id)}
								<article class="ability-card">
									<h3>{ability.name}</h3>
									<p>{ability.description}</p>
									{#if ability.status}
										<StatusBadge label={ability.status.label} tone={ability.status.tone} />
									{/if}
									{#if ability.action}
										<Button
											variant={ability.action.variant}
											loading={busyAbilityId === ability.id}
											onclick={() => onAbilityAction(ability.id, ability.action!.command)}
											>{ability.action.label}</Button
										>
									{:else if ability.unavailableLabel}
										<p class="ability-unavailable">{ability.unavailableLabel}</p>
									{/if}
								</article>
							{/each}
						</div>
					{/if}
				</Panel>

				<Panel title="Information you know">
					{#if knowledge.length === 0}
						<p>No additional information.</p>
					{:else}
						<ul>
							{#each knowledge as item (item.id)}
								<li>{item.text}</li>
							{/each}
						</ul>
					{/if}
				</Panel>
			</div>
		</article>
	</PrivateDisclosure>
{/if}

<style>
	.unavailable {
		display: grid;
		width: min(100%, 34rem);
		min-height: calc(100dvh - 7.75rem);
		place-content: center;
		margin-inline: auto;
		padding: var(--space-5);
	}

	.unavailable-copy {
		display: grid;
		justify-items: center;
		gap: var(--space-3);
		text-align: center;
	}

	.unavailable-copy h1,
	.unavailable-copy p {
		margin: 0;
	}

	.role-reveal {
		width: min(100%, 52rem);
		min-height: calc(100dvh - 7.75rem);
		margin-inline: auto;
		border-inline: var(--border-strong);
		background: var(--paper);
		box-shadow: var(--shadow);
	}

	.role-content {
		display: grid;
		gap: var(--space-6);
		padding: clamp(var(--space-4), 6vw, var(--space-7));
		padding-block-end: 7rem;
	}

	.role-content :global(h2),
	.role-content p {
		margin-block-start: 0;
	}

	.ability-list {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 15rem), 1fr));
		gap: var(--space-4);
	}

	.ability-card {
		border-inline-start: 3px solid var(--crimson);
		padding-inline-start: var(--space-3);
	}

	.ability-card h3,
	.ability-card p {
		margin-block-start: 0;
	}

	.ability-unavailable {
		color: var(--ink-faint);
		font-size: 0.82rem;
	}

	.role-content li {
		margin-block-end: var(--space-2);
	}
</style>
