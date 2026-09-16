import type { Game } from '$lib/api/types';

export type PlayerShellContext = {
	readonly loading: boolean;
	readonly availableLobby: Game | null;
	readonly liveGame: Game | null;
	readonly joiningLobby: boolean;
	readonly loadError: string;
	refresh: () => Promise<void>;
	joinAvailableLobby: () => Promise<void>;
};

export const playerShellContextKey = Symbol('player-shell');
