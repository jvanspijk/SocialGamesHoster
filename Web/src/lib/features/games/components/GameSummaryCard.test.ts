import { cleanup, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';
import GameSummaryCard from './GameSummaryCard.svelte';
import type { GameSummary } from '$lib/api/types';

const summary = {
	game: { id: 'game-1', name: 'The Longest Game Name' },
	ruleset: { name: 'Murder at Midnight' },
	durationMs: 89_000,
	participants: [
		{
			id: 'participant-1',
			gameAlias: 'A very long alias that should wrap cleanly on a phone',
			displayNameSnapshot: 'Alex',
			seatNumber: 3,
			outcome: 'win',
			achievements: [
				{ id: 'award-1', title: 'Sharp Eye', description: '', points: 2 },
				{ id: 'award-2', title: 'Lasting Legacy', description: '', points: 1 }
			]
		},
		{
			id: 'participant-2',
			gameAlias: '',
			displayNameSnapshot: 'Jordan',
			seatNumber: 4,
			outcome: 'loss',
			achievements: []
		}
	],
	immutable: true
} as GameSummary;

afterEach(cleanup);

describe('GameSummaryCard', () => {
	it('keeps game-summary interpretation in the feature while composing shared structures', () => {
		const view = render(GameSummaryCard, { props: { summary } });

		expect(screen.getByText('Murder at Midnight')).toBeVisible();
		expect(screen.getByRole('heading', { name: 'The Longest Game Name' })).toBeVisible();
		expect(screen.getByText('1 minutes · 2 players')).toBeVisible();
		expect(screen.getByText('A very long alias that should wrap cleanly on a phone')).toBeVisible();
		expect(screen.getByText('Jordan')).toBeVisible();
		expect(screen.getByText('Seat 3')).toBeVisible();
		expect(screen.getByText('Win')).toHaveClass('tone-success');
		expect(screen.getByText('Loss')).toHaveClass('tone-danger');
		expect(screen.getByText('Sharp Eye')).toBeVisible();
		expect(screen.getByText('Lasting Legacy')).toBeVisible();
		expect(view.container.querySelectorAll('.tag')).toHaveLength(2);
	});

	it.each([
		['unset', 'Unset', 'tone-warning'],
		['draw', 'Draw', 'tone-info']
	] as const)('presents %s outcomes with explicit status text and tone', (outcome, label, tone) => {
		render(GameSummaryCard, {
			props: { summary: { ...summary, participants: [{ ...summary.participants[0], outcome }] } }
		});

		expect(screen.getByText(label)).toHaveClass(tone);
	});
});
