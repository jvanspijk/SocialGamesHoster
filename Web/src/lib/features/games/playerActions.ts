import type { AdminGameView, Participant } from '$lib/api/types';

export type PlayerAction = 'eliminate' | 'reinstate' | 'achievement' | 'outcome';

export const playerActionLabels: Record<PlayerAction, string> = {
	eliminate: 'Eliminate player',
	reinstate: 'Reinstate player',
	achievement: 'Award or revoke achievement',
	outcome: 'Set outcome'
};

export function playerActionOptions(view: AdminGameView, player: Participant) {
	const actions: PlayerAction[] = [];
	if (view.game.status === 'archived' || ['kicked', 'left'].includes(player.status)) return [];
	if (player.status === 'active' && ['running', 'paused'].includes(view.game.status))
		actions.push('eliminate');
	if (player.status === 'eliminated') actions.push('reinstate');
	if (['running', 'paused', 'review'].includes(view.game.status)) {
		if (view.ruleset.achievements.length) actions.push('achievement');
		actions.push('outcome');
	}
	return actions.map((value) => ({ value, label: playerActionLabels[value] }));
}
