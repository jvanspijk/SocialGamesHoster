import type { PageServerLoad } from './$types';
import { getActiveGames, getRulesetById } from '$lib/client';
import type {GetActiveGameSessionsResponse, GetRulesetResponse } from '$lib/client'

export const load = (async () => {
    const gamesResponse = await getActiveGames();
    const games = (gamesResponse.data || []) as GetActiveGameSessionsResponse[];

    return {
        games,
        streamed: {
            rulesets: fetchAllRulesets(games)
        }
    };
}) satisfies PageServerLoad;


const fetchAllRulesets = async (games: GetActiveGameSessionsResponse[]): Promise<Record<number, GetRulesetResponse>> => {
    const uniqueRulesetIds = Array.from(new Set(
        games.map(game => game.rulesetId).filter((id): id is number => id != null)
    ));   

    const results = await Promise.all(
        uniqueRulesetIds.map(id => 
            getRulesetById({ path: { rulesetId: id } })
            .catch(err => {
                console.error(`Error fetching ruleset ${id}:`, err);
                return { data: null };
            })
        )
    );

    const idToRuleset = new Map(
        results
            .filter(r => r.data) 
            .map(r => [r.data!.id, r.data!])
    );

    const dictionary: Record<number, GetRulesetResponse> = {};
    for (const game of games) {
        const data = idToRuleset.get(game.rulesetId);
        if (data) dictionary[game.id] = data;
    }

    return dictionary;   
}