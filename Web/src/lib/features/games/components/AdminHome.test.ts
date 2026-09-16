import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Game } from '$lib/api/types';
import AdminHome from './AdminHome.svelte';

afterEach(cleanup);

function game(status: Game['status']): Game {
	return {
		id: 'game-1',
		name: 'Friday Night',
		status,
		rulesetVersion: 'version-1',
		joiningOpen: status === 'lobby',
		playerCount: 7,
		rolesVisible: false,
		roleVisibilityRevision: 0,
		revision: 1,
		roundNumber: 0,
		phaseKey: ''
	};
}

describe('AdminHome', () => {
	it.each(['lobby', 'running', 'paused'] as const)('opens a %s game at its overview', (status) => {
		render(AdminHome, {
			props: { games: [game(status)], loading: false, loadError: '', retry: vi.fn() }
		});
		expect(screen.getByRole('link', { name: 'Open game' })).toHaveAttribute(
			'href',
			'/admin/games/game-1/overview'
		);
	});

	it('continues a game in review through the completion flow', () => {
		render(AdminHome, {
			props: { games: [game('review')], loading: false, loadError: '', retry: vi.fn() }
		});
		expect(screen.getByRole('link', { name: 'Continue finishing' })).toHaveAttribute(
			'href',
			'/admin/games/game-1/finish/outcomes'
		);
	});

	it('does not treat drafts or archived games as active', () => {
		render(AdminHome, {
			props: {
				games: [game('draft'), game('archived')],
				loading: false,
				loadError: '',
				retry: vi.fn()
			}
		});
		expect(screen.getByRole('heading', { name: 'No active game' })).toBeVisible();
	});

	it('shows a retry action when loading fails', async () => {
		const retry = vi.fn();
		render(AdminHome, {
			props: { games: [], loading: false, loadError: 'Host unavailable.', retry }
		});
		await fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
		expect(screen.getByRole('alert')).toHaveTextContent('Host unavailable.');
		expect(retry).toHaveBeenCalledOnce();
	});
});
