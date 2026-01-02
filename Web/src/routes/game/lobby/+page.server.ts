import type { PageServerLoad } from './$types';
import {
	GetActiveGameSessions,
	type GetActiveGameSessionsResponse
} from '$lib/client/GameSessions/GetActiveGameSessions';
import { GetRuleset, type GetRulesetResponse } from '$lib/client/Rulesets/GetRuleset';
import { error } from '@sveltejs/kit';

export const load = (async ({ fetch }) => {
	const result = await GetActiveGameSessions(fetch);
	if (!result.ok) {
		throw error(result.error.status, {
			message: result.error.title || 'Failed to load games'
		});
	}

	const games = result.data;

	return {
		games,
		streamed: {
			rulesets: fetchAllRulesets(fetch, games)
		}
	};
}) satisfies PageServerLoad;

/**
 * Helper to fetch unique rulesets and map them to Game IDs
 */
async function fetchAllRulesets(f: typeof fetch, games: GetActiveGameSessionsResponse[]) {
	const uniqueRulesetIds = Array.from(
		new Set(games.map((g) => g.rulesetId).filter((id) => id != null))
	);

	const rulesetResults = await Promise.all(
		uniqueRulesetIds.map(async (id) => {
			const res = await GetRuleset(f, { rulesetId: id.toString() });
			return res.ok ? res.data : null;
		})
	);

	const rulesetMap = new Map(
		rulesetResults.filter((r): r is NonNullable<typeof r> => r !== null).map((r) => [r.id, r])
	);

	const dictionary: Record<number, GetRulesetResponse> = {};
	for (const game of games) {
		const data = rulesetMap.get(game.rulesetId);
		if (data) {
			dictionary[game.id] = data;
		}
	}

	return dictionary;
}
