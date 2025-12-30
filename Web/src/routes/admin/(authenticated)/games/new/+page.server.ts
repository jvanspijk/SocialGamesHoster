import type { PageServerLoad } from './$types';
import { GetAllRulesets } from '$lib/client/Rulesets/GetAllRulesets';
import { CreateGameSession } from '$lib/client/GameSessions/CreateGameSession';
import type { GetAllRulesetsResponse } from '$lib/client/Rulesets/GetAllRulesets';
import type { Actions } from './$types';
import { fail, redirect, error } from '@sveltejs/kit';

export const actions: Actions = {
    start: async ({ request, fetch }) => {
        const formData = await request.formData();
        const selectedRulesetId: number = Number(formData.get('selectedRulesetId'));

        if (!selectedRulesetId) {
            console.error("selectedRulesetId not found");
            console.error(formData.get('selectedRulesetId'));
        }
        
        if (!selectedRulesetId) {
            return fail(400, { 
                success: false, 
                message: 'Ruleset selection is required.' 
            });
        }
        
        try {            
            const request = {
                rulesetId: selectedRulesetId,
                playerNames: [],
            };

            const response = await CreateGameSession(fetch, request);

            if(!response.ok) {
                return fail(response.error.status, { 
                    success: false, 
                    message: response.error.title || 'Create failed due to a server or network error.' 
                });
            }           
        } catch (error) {
            console.error('Server Create Exception:', error);
            
            return fail(500, { 
                success: false, 
                message: 'Create failed due to a server or network error.' 
            });
        }
        redirect(303, "/admin/games")        
    }
};

export const load = (async ({fetch}) => {
    const response = await GetAllRulesets(fetch);
    if(!response.ok) {
        throw error(response.error.status, {
            message: response.error.title || 'Failed to load rulesets.'
        });
    }
    const rulesets: GetAllRulesetsResponse[] = (response.data || []);
    return {
        rulesets: rulesets
    };
}) satisfies PageServerLoad;