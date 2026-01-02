import { error, type Actions } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { GetGameSession } from '$lib/client/GameSessions/GetGameSession';
import { GetPlayersFromGame } from '$lib/client/Players/GetPlayersFromGame';
import { GetCurrentRound } from '$lib/client/Rounds/GetCurrentRound';
import { GetRoles } from '$lib/client/Roles/GetRoles';
import { UpdatePlayer } from '$lib/client/Players/UpdatePlayer';
import { AddWinners } from '$lib/client/GameSessions/AddWinners';
import { GetFullPlayer } from '$lib/client/Players/GetFullPlayer';
import { GetTimerState } from '$lib/client/Timers/GetTimerState';
import { PauseTimer } from '$lib/client/Timers/PauseTimer';
import { ResumeTimer } from '$lib/client/Timers/ResumeTimer';
import { StartTimer } from '$lib/client/Timers/StartTimer';
import { StopTimer } from '$lib/client/Timers/StopTimer';

interface RoleUpdate {
	id: number;
	roleId: number | null;
}

export const load = (async ({ fetch, params }) => {
	const gameId = params.id;
	const req = { gameId };

	const [gameRes, playersRes, roundRes, timerRes] = await Promise.all([
		GetGameSession(fetch, req),
		GetPlayersFromGame(fetch, req),
		GetCurrentRound(fetch, { gameId: Number(gameId) }),
		GetTimerState(fetch)
	]);

	if (!gameRes.ok) throw error(404, 'Session not found');

	const playersSummary = playersRes.ok ? (playersRes.data ?? []) : [];

	const rolesRes = await GetRoles(fetch, { rulesetId: gameRes.data.rulesetId.toString() });

	return {
		gameSession: gameRes.data,
		players: playersSummary,
		roles: rolesRes.ok ? rolesRes.data : [],
		currentRound: roundRes.ok ? roundRes.data : null,
		timer: timerRes.ok ? timerRes.data : null,
		streamed: {
			fullPlayers: Promise.all(
				playersSummary.map(async (p) => {
					const res = await GetFullPlayer(fetch, { id: p.id.toString() });
					return res.ok ? res.data : null;
				})
			)
		}
	};
}) satisfies PageServerLoad;

export const actions = {
	saveRoles: async ({ request, fetch }) => {
		const formData = await request.formData();
		const updatesJson = formData.get('updates');

		if (typeof updatesJson !== 'string') {
			throw error(400, 'Invalid updates format');
		}

		const updates = JSON.parse(updatesJson) as RoleUpdate[];
		const results = await Promise.all(
			updates.map((u) =>
				UpdatePlayer(fetch, {
					id: u.id.toString(),
					newName: null,
					newRoleId: u.roleId
				})
			)
		);

		return { success: results.every((r) => r.ok) };
	},
	declareWinners: async ({ request, fetch, params }) => {
		const formData = await request.formData();
		const playerIds = formData.getAll('winnerIds').map(Number);
		const res = await AddWinners(fetch, { gameId: params.id!, playerIds });
		return { success: res.ok };
	},
	pauseTimer: async ({ fetch }) => {
		const res = await PauseTimer(fetch);
		return { success: res.ok };
	},
	resumeTimer: async ({ fetch }) => {
		const res = await ResumeTimer(fetch);
		return { success: res.ok };
	},
	startTimer: async ({ request, fetch }) => {
		const formData = await request.formData();
		const seconds = Number(formData.get('seconds')) || 0;
		const res = await StartTimer(fetch, { durationSeconds: seconds });
		return { success: res.ok };
	},
	stopTimer: async ({ fetch }) => {
		const res = await StopTimer(fetch);
		return { success: res.ok };
	}
} satisfies Actions;
