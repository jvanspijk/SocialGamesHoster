import type { Actor } from '$lib/api/types';
import { clearAuth, pb, saveAuth } from '$lib/api/client';

let actor = $state<Actor | null>((pb.authStore.record as Actor | null) ?? null);

pb.authStore.onChange((_token, record) => {
	actor = record as Actor | null;
});

export const auth = {
	get actor() {
		return actor;
	},
	get authenticated() {
		return pb.authStore.isValid && actor !== null;
	},
	get isGameMaster() {
		return actor?.type === 'game_masters';
	},
	get isPlayer() {
		return actor?.type === 'player_profiles';
	},
	get isOwner() {
		return actor?.isOwner === true;
	},
	save: saveAuth,
	clear: clearAuth
};
