import type { Game } from '$lib/api/types';

export const gameStatusLabels = {
	draft: 'Draft',
	lobby: 'Lobby open',
	running: 'Running',
	paused: 'Paused',
	review: 'Finishing',
	archived: 'Archived'
} satisfies Record<Game['status'], string>;

export function gameStatusLabel(status: Game['status']) {
	return gameStatusLabels[status];
}
