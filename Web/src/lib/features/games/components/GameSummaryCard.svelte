<script lang="ts">
	import { Award } from '@lucide/svelte';
	import Panel from '$lib/components/Panel.svelte';
	import RecordItem from '$lib/components/RecordItem.svelte';
	import RecordList from '$lib/components/RecordList.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import Tag from '$lib/components/Tag.svelte';
	import TagList from '$lib/components/TagList.svelte';
	import type { GameSummary } from '$lib/api/types';

	let { summary }: { summary: GameSummary } = $props();

	function playerName(player: GameSummary['participants'][number]) {
		return player.gameAlias || player.displayNameSnapshot;
	}

	function outcomePresentation(outcome: GameSummary['participants'][number]['outcome']) {
		switch (outcome) {
			case 'win':
				return { label: 'Win', tone: 'success' as const };
			case 'loss':
				return { label: 'Loss', tone: 'danger' as const };
			case 'draw':
				return { label: 'Draw', tone: 'info' as const };
			default:
				return { label: 'Unset', tone: 'warning' as const };
		}
	}
</script>

<Panel variant="dark">
	<div class="summary-hero">
		<div class="summary-mark">{summary.ruleset.name.slice(0, 1).toUpperCase()}</div>
		<div>
			<p>{summary.ruleset.name}</p>
			<h2>{summary.game.name}</h2>
			<p>
				{Math.max(1, Math.round(summary.durationMs / 60_000))} minutes · {summary.participants
					.length}
				players
			</p>
		</div>
	</div>
</Panel>
<RecordList>
	{#each summary.participants as player (player.id)}
		<RecordItem>
			{#snippet trailing()}
				<StatusBadge {...outcomePresentation(player.outcome)} />
			{/snippet}
			{#snippet supporting()}
				<TagList align="end">
					{#each player.achievements as achievement (achievement.id)}
						<Tag>
							{#snippet icon()}<Award size={14} />{/snippet}
							{achievement.title}
						</Tag>
					{/each}
				</TagList>
			{/snippet}
			<div class="participant">
				<h3>{playerName(player)}</h3>
				<p>Seat {player.seatNumber}</p>
			</div>
		</RecordItem>
	{/each}
</RecordList>

<style>
	.summary-hero {
		display: grid;
		grid-template-columns: 8rem minmax(0, 1fr);
		align-items: center;
		gap: var(--space-5);
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

	.participant h3,
	.participant p {
		margin: 0;
	}

	@media (max-width: 47.99rem) {
		.summary-hero {
			grid-template-columns: 5rem minmax(0, 1fr);
			gap: var(--space-3);
		}

		.summary-mark {
			width: 5rem;
			height: 5rem;
			font-size: 2rem;
		}
	}
</style>
