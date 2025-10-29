import type { PageServerLoad } from './$types';
import { createPlayer } from '$lib/client';
import { getGamePlayers } from '$lib/client';
import { playerLogin } from '$lib/client';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

export const actions: Actions = {
    login: async ({ request, cookies, params, getClientAddress }) => {
        const formData = await request.formData();
        const playerNameInput: string | undefined = formData.get('playerName')?.toString().trim(); 
        const gameId: number = Number(params.gameId);
        
        if (!playerNameInput || playerNameInput === "") {
            return fail(400, { 
                success: false, 
                message: 'Player name is required.' 
            });
        }

        let playerId: number;
        try {
            const options = {
                path: {                    
                    gameId: gameId                    
                },
                body: {
                    name: playerNameInput
                }
            }

            const res = await createPlayer(options);
            if(!res.data) {
                return fail(500, { 
                    success: false, 
                    message: 'Player creation failed.' 
                }); 
            }
            playerId = res.data.id;
        } catch (error) {
            console.error('Player creation Exception:', error);
            
            return fail(500, { 
                success: false, 
                message: 'Player creation failed.' 
            });     
        }        
        
        try {            
            const options = {
                path: {
                    gameId: gameId,            
                },
                body: {
                    playerId: playerId,
                    ipAddress: getClientAddress(),
                }
            };
            
            const { data: response } = await playerLogin(options);

            if(!response || !response.token) {
                return fail(500, { 
                    success: false, 
                    message: 'Login failed due to a server or network error.' 
                });
            }

            const maxCookieAge: number = 60 * 60 * 24 // One day

            cookies.set('auth_token', response.token, {
                path: '/', 
                httpOnly: false,
                secure: false, 
                maxAge: maxCookieAge 
            });

            cookies.set('player_id', playerId.toLocaleString(), {
                path: '/', 
                httpOnly: false,
                secure: false, 
                maxAge: maxCookieAge 
            });

            cookies.set('game_id', gameId.toLocaleString(), {
                path: '/', 
                httpOnly: false,
                secure: false,  
                maxAge: maxCookieAge
            });            
        } catch (error) {
            console.error('Server Login Exception:', error);
            
            return fail(500, { 
                success: false, 
                message: 'Login failed due to a server or network error.' 
            });
        }
        redirect(303, '/me');
    }
};

export const load = (async ({params}) => {
    // TODO: redirect if cookies are already set
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