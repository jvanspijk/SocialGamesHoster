<script lang="ts">
	import { Eye, EyeOff, ShieldCheck } from '@lucide/svelte';
	import type { RoleProjection } from '$lib/api/types';
	import Button from './Button.svelte';
	import ProtectedMedia from './ProtectedMedia.svelte';

	let {
		role,
		knowledge,
		imageUrl = ''
	}: {
		role: RoleProjection | null;
		knowledge: Array<Record<string, unknown>>;
		imageUrl?: string;
	} = $props();
	let concealed = $state(false);

	function describeFact(fact: Record<string, unknown>) {
		const revealedRole = fact.role as { name?: string } | undefined;
		return String(revealedRole?.name ?? fact.teamId ?? fact.displayName ?? 'revealed');
	}
</script>

<article class:concealed class="role-card card--dark">
	<div class="role-head">
		<div>
			<p class="eyebrow">Your role</p>
			<h2>{role?.name ?? 'Awaiting assignment'}</h2>
		</div>
		<Button variant="ghost" onclick={() => (concealed = !concealed)}>
			{#if concealed}<Eye size={18} /> Reveal{:else}<EyeOff size={18} /> Conceal{/if}
		</Button>
	</div>
	<div class="secret">
		{#if role}
			{#if imageUrl}<div class="role-image">
					<ProtectedMedia src={imageUrl} kind="image" alt="" />
				</div>{/if}
			<p class="team">{role.team?.name ?? 'Independent'}</p>
			<p>{role.description}</p>
			<div class="rule">
				<ShieldCheck size={19} aria-hidden="true" />
				<div><strong>Win condition</strong><br />{role.winCondition}</div>
			</div>
			{#if role.abilities.length}
				<h3>Abilities</h3>
				<ul>
					{#each role.abilities as ability (ability.id)}
						<li><strong>{ability.name}</strong> — {ability.description}</li>
					{/each}
				</ul>
			{/if}
			{#if knowledge.length}
				<h3>Known information</h3>
				<ul>
					{#each knowledge as fact (String(fact.participantId))}
						<li>Player {String(fact.seatNumber)}: {describeFact(fact)}</li>
					{/each}
				</ul>
			{/if}
		{:else}
			<p>Your game master has not revealed a role yet.</p>
		{/if}
	</div>
	{#if concealed}
		<div class="cover" aria-live="polite">
			<EyeOff size={36} />
			<strong>Role concealed</strong>
			<span>Safe to pass the phone</span>
		</div>
	{/if}
</article>

<style>
	.role-card {
		position: relative;
		min-height: 21rem;
		overflow: hidden;
		border: 1px solid #1e1511;
		padding: 1.15rem;
	}

	.role-card::after {
		position: absolute;
		inset: 0.55rem;
		border: 1px solid rgb(247 231 196 / 22%);
		content: '';
		pointer-events: none;
	}

	.role-head {
		position: relative;
		z-index: 2;
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
	}

	.eyebrow,
	.team {
		margin: 0 0 0.25rem;
		color: var(--paper-deep);
		font-family: var(--font-display);
		font-size: 0.67rem;
		letter-spacing: 0.14em;
		text-transform: uppercase;
	}

	.secret {
		position: relative;
		z-index: 1;
		transition:
			opacity var(--speed-fast) ease-out,
			transform var(--speed-fast) ease-out;
	}

	.concealed .secret {
		opacity: 0;
		transform: scale(0.98);
	}

	.rule {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 0.65rem;
		border-block: 1px solid rgb(247 231 196 / 25%);
		margin-block: 1rem;
		padding-block: 0.8rem;
	}

	.role-image {
		float: inline-end;
		width: min(36%, 10rem);
		margin: 0 0 0.75rem 0.75rem;
	}

	.role-image :global(img) {
		width: 100%;
		border: 1px solid rgb(247 231 196 / 35%);
		object-fit: cover;
	}

	ul {
		padding-inline-start: 1.1rem;
	}

	.cover {
		position: absolute;
		z-index: 3;
		inset: 0;
		display: grid;
		place-content: center;
		place-items: center;
		gap: 0.5rem;
		background: var(--ink);
		color: var(--paper-light);
		text-align: center;
	}

	.cover strong {
		font-family: var(--font-display);
		letter-spacing: 0.1em;
		text-transform: uppercase;
	}

	.cover span {
		color: var(--paper-deep);
	}
</style>
