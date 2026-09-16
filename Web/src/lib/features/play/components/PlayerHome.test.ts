import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Game, PlayerGameView } from '$lib/api/types';
import PlayerHome from './PlayerHome.svelte';

afterEach(cleanup);

const game: Game = {
	id: 'game-1',
	name: 'Friday Night',
	status: 'lobby',
	rulesetVersion: 'version-1',
	joiningOpen: true,
	rolesVisible: false,
	roleVisibilityRevision: 0,
	revision: 1,
	roundNumber: 0,
	phaseKey: ''
};

const view = {
	game,
	participant: {},
	ruleset: {},
	roleAvailable: false,
	roleRevision: 0,
	role: null,
	knowledge: [],
	rooms: [],
	attentionItems: [],
	assets: [],
	party: []
} as unknown as PlayerGameView;

type PlayerHomeProps = {
	displayName: string;
	loading: boolean;
	view: PlayerGameView | null;
	availableLobby: Game | null;
	liveGame: Game | null;
	joiningLobby: boolean;
	loadError: string;
	join: () => void | Promise<void>;
	retry: () => void | Promise<void>;
};

function renderHome(overrides: Partial<PlayerHomeProps> = {}) {
	return render(PlayerHome, {
		props: {
			displayName: 'Rowan',
			loading: false,
			view: null,
			availableLobby: null,
			liveGame: null,
			joiningLobby: false,
			loadError: '',
			join: vi.fn(),
			retry: vi.fn(),
			...overrides
		}
	});
}

describe('PlayerHome', () => {
	it('links a joined player to the current game', () => {
		renderHome({ view });
		expect(
			screen.getByRole('link', { name: /Current game.*Friday Night.*Lobby open/ })
		).toHaveAttribute('href', '/play/game');
	});

	it('offers to join an open lobby', async () => {
		const join = vi.fn();
		renderHome({ availableLobby: game, join });
		await fireEvent.click(screen.getByRole('button', { name: /Join game.*Friday Night/ }));
		expect(join).toHaveBeenCalledOnce();
	});

	it('explains when a live game is closed to the player', () => {
		renderHome({ liveGame: { ...game, joiningOpen: false, status: 'running' } });
		expect(screen.getByText('Friday Night is not accepting players.')).toBeVisible();
	});

	it('shows the no-game and loading states', () => {
		const rendered = renderHome();
		expect(screen.getByText('No lobby is open right now.')).toBeVisible();
		rendered.unmount();
		renderHome({ loading: true });
		expect(screen.getByText('Checking for your current game…')).toBeVisible();
		expect(screen.getByText('Checking for an open lobby…')).toBeVisible();
	});

	it('shows a recoverable join or load failure', async () => {
		const retry = vi.fn();
		renderHome({ loadError: 'The lobby could not be joined.', retry });
		expect(screen.getByRole('alert')).toHaveTextContent('The lobby could not be joined.');
		await fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
		expect(retry).toHaveBeenCalledOnce();
	});
});
