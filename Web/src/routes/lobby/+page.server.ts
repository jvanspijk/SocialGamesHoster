import type { PageServerLoad } from './$types';
import { getActiveGames, getRulesetById } from '$lib/client';
import type {GetActiveGameSessionsResponse, GetRulesetResponse } from '$lib/client'

export const load = (async () => {
    const gamesResponse = await getActiveGames();
    const games: GetActiveGameSessionsResponse[] = (gamesResponse.data || []) as GetActiveGameSessionsResponse[];
    const rulesetsPromise = fetchAllRulesets(games);
    return {
        games: games,
        rulesets: rulesetsPromise
    };
}) satisfies PageServerLoad;


const fetchAllRulesets = async (games: GetActiveGameSessionsResponse[]): Promise<Record<number, GetRulesetResponse>> => {
    const uniqueRulesetIds = Array.from(new Set(
        games.map(game => game.rulesetId).filter((id): id is number => id != null)
    ));   

    const fetchPromises = uniqueRulesetIds.map(async (rulesetId) => {
        try {
            const options = { path: { rulesetId } };
            const response = await getRulesetById(options);
            
            if (!response.data) {
                return {}; 
            }
            
            return response.data;

        } catch (error) {
            console.error(`Failed to load ruleset ${rulesetId} during streaming:`, error);
            return {};
        }
    });

    const results = await Promise.all(fetchPromises) as GetRulesetResponse[];    
    const rulesetDictionary: Record<number, GetRulesetResponse> = {};
    
    games.forEach(game => {
        if (game.rulesetId != null) { 
            const rulesetResult = results.find(r => r.id === game.rulesetId);
            
            if (rulesetResult) {
                rulesetDictionary[game.id] = rulesetResult;
            }
        }
    });

    return rulesetDictionary;
}