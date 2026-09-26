import { cleanup, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { PlayerHistoryView } from '$lib/api/types';
import PlayerHistory from './PlayerHistory.svelte';

afterEach(cleanup);

const history: PlayerHistoryView = {
	profile: {
		id: 'profile-1',
		displayName: 'Rowan',
		avatar: '',
		bio: '',
		active: true
	},
	games: [
		{
			id: 'game-1',
			name: 'Friday Night',
			rulesetName: 'Secret Village',
			roleName: 'Seer',
			outcome: 'win',
			achievements: [{ id: 'achievement-1', title: 'Sharp Eye' }]
		}
	],
	statistics: { achievementCount: 1, achievementPoints: 10 }
};

describe('PlayerHistory', () => {
	it('shows completed games, statistics, and achievements', () => {
		render(PlayerHistory, {
			props: { history, loading: false, loadError: '', retry: vi.fn() }
		});
		expect(screen.getByText('10 achievement points across 1 achievements')).toBeVisible();
		expect(screen.getByRole('heading', { name: 'Friday Night' })).toBeVisible();
		expect(screen.getByText('Sharp Eye')).toBeVisible();
		expect(screen.getByRole('link', { name: 'Player home' })).toHaveAttribute('href', '/play');
	});

	it('shows the empty-history state', () => {
		render(PlayerHistory, {
			props: {
				history: {
					...history,
					games: [],
					statistics: { achievementCount: 0, achievementPoints: 0 }
				},
				loading: false,
				loadError: '',
				retry: vi.fn()
			}
		});
		expect(screen.getByText('No completed games yet.')).toBeVisible();
		expect(screen.getByRole('link', { name: 'Player home' })).toHaveAttribute('href', '/play');
	});
});
