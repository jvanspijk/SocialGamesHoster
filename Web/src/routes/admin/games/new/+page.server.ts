import type { PageServerLoad } from './$types';
import { getAllRulesets, createGameSession } from '$lib/client';
import type { GetAllRulesetsResponse } from '$lib/client';
import type { Actions } from './$types';
import { fail, redirect } from '@sveltejs/kit';

export const actions: Actions = {
    start: async ({ request }) => {
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
            const options = {
                body: {
                    rulesetId: selectedRulesetId,
                    playerNames: [],
                }
            };

            const { data: response } = await createGameSession(options);

            if(!response) {
                console.error("No response");

                return fail(500, { 
                    success: false, 
                    message: 'Create failed due to a server or network error.' 
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

export const load = (async () => {
    const rulesetsResponse = await getAllRulesets();
    const rulesets: GetAllRulesetsResponse[] = (rulesetsResponse.data || []) as GetAllRulesetsResponse[];
    return {
        rulesets: rulesets
    };
}) satisfies PageServerLoad;