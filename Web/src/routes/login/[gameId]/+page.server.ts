import type { PageServerLoad } from './$types';
import { getGamePlayers } from '$lib/client';
import { playerLogin } from '$lib/client';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

export const actions: Actions = {
    login: async ({ request, cookies, params }) => {
        const formData = await request.formData();
        const selectedPlayerId = Number(formData.get('selectedPlayerId'));       

        
        if (!selectedPlayerId) {
            return fail(400, { 
                success: false, 
                message: 'Player selection is required.' 
            });
        }
        
        try {

            const playerName = 'TestPlayer';
            
            const options = {
                path: {
                    gameId: Number(params.gameId),
                    name: playerName,                 
                }
            };
            
            const { data: loginToken } = await playerLogin(options);

            if(!loginToken) {
                // TODO: return error
                return;
            }
            
            // 3. Set the cookie securely on the server
            cookies.set('auth_token', loginToken, {
                path: '/', 
                httpOnly: true, 
                maxAge: 60 * 60 * 24 * 1 
            });

            cookies.set('player_id', selectedPlayerId.toLocaleString(), {
                path: '/', 
                httpOnly: true, 
                maxAge: 60 * 60 * 24 * 1 
            });

            throw redirect(303, '/me');

        } catch (error) {
            console.error('Server Login Exception:', error);
            
            return fail(500, { 
                success: false, 
                message: 'Login failed due to a server or network error.' 
            });
        }
    }
};

export const load = (async ({params}) => {
    const options = {
        path: {
            gameId: Number(params.gameId),                 
        }
    };
    const response = await getGamePlayers(options);
    return {
        players: response.data,
        gameId: params.gameId
    };
}) satisfies PageServerLoad;