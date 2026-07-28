<script lang="ts">
	import { Award } from '@lucide/svelte';
	import type { GameSummary } from '$lib/api/types';

	let { summary }: { summary: GameSummary } = $props();

	function playerName(player: GameSummary['participants'][number]) {
		return player.gameAlias || player.displayNameSnapshot;
	}
</script>

<section class="summary-hero">
	<div class="summary-mark">{summary.ruleset.name.slice(0, 1).toUpperCase()}</div>
	<div>
		<p>{summary.ruleset.name}</p>
		<h2>{summary.game.name}</h2>
		<p>
			{Math.max(1, Math.round(summary.durationMs / 60_000))} minutes · {summary.participants.length}
			players
		</p>
	</div>
</section>
<div class="summary-list">
	{#each summary.participants as player (player.id)}
		<article>
			<div>
				<h3>{playerName(player)}</h3>
				<p>Seat {player.seatNumber}</p>
			</div>
			<strong class:win={player.outcome === 'win'}>{player.outcome}</strong>
			<div class="summary-awards">
				{#each player.achievements as achievement (achievement.id)}
					<span><Award size={14} /> {achievement.title}</span>
				{/each}
			</div>
		</article>
	{/each}
</div>

<style>
	.summary-hero {
		display: grid;
		grid-template-columns: 8rem 1fr;
		align-items: center;
		gap: var(--space-5);
		border: var(--border-strong);
		background: linear-gradient(100deg, rgb(22 13 8 / 92%), rgb(62 38 24 / 82%)), var(--wood);
		color: var(--paper-light);
		padding: var(--space-5);
	}

	.summary-mark {
		display: grid;
		width: 8rem;
		height: 8rem;
		place-items: center;
		border: 4px double var(--gold-light);
		border-radius: 50%;
		background: var(--crimson-dark);
		color: var(--gold-light);
		font-family: var(--font-display);
		font-size: 3.2rem;
	}

	.summary-hero h2,
	.summary-hero p {
		margin: 0;
		color: var(--paper-light);
	}

	.summary-list article {
		display: grid;
		grid-template-columns: 1fr auto minmax(10rem, auto);
		align-items: center;
		gap: var(--space-3);
		border-block-end: var(--border-subtle);
		padding: var(--space-3) 0;
	}

	.summary-list h3,
	.summary-list p {
		margin: 0;
	}

	.summary-list > article > strong {
		text-transform: capitalize;
	}

	.summary-list > article > strong.win {
		color: var(--success);
	}

	.summary-awards {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: var(--space-1);
	}

	.summary-awards span {
		display: inline-flex;
		align-items: center;
		gap: 0.2rem;
		border: 1px solid var(--gold-dark);
		font-size: 0.75rem;
		padding: 0.2rem 0.4rem;
	}

	@media (max-width: 47.99rem) {
		.summary-hero {
			grid-template-columns: 5rem 1fr;
			gap: var(--space-3);
			padding: var(--space-3);
		}

		.summary-mark {
			width: 5rem;
			height: 5rem;
			font-size: 2rem;
		}

		.summary-list article {
			grid-template-columns: 1fr auto;
		}

		.summary-awards {
			grid-column: 1 / -1;
			justify-content: flex-start;
		}
	}
</style>
