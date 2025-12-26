import type { PageServerLoad } from './$types';
// import { createPlayer } from '$lib/client';
// import { getGamePlayers } from '$lib/client';
// import { playerLogin } from '$lib/client';
import { CreatePlayer, type CreatePlayerRequest } from '$lib/client/Players/CreatePlayer';
import { GetPlayersFromGame } from '$lib/client/Players/GetPlayersFromGame';
import { PlayerLogin, type PlayerLoginRequest } from '$lib/client/Authentication/PlayerLogin';
import { fail, redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

export const actions: Actions = {
    login: async ({ request, cookies, params, getClientAddress, fetch }) => {
        const formData = await request.formData();
        const playerNameInput: string | undefined = formData.get('playerName')?.toString().trim(); 
        const gameId = params.gameId;
        
        if (!playerNameInput || playerNameInput === "") {
            return fail(400, { 
                success: false, 
                message: 'Player name is required.' 
            });
        }

        let playerId: number;
        try {
            const request: CreatePlayerRequest = {
                gameId: gameId,
                name: playerNameInput
            }

            const res = await CreatePlayer(fetch, request);
            if(!res.ok) {
                return fail(res.error.status, { 
                    success: false, 
                    message: res.error.title || 'Player creation failed.'
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
            const request: PlayerLoginRequest = {
                gameId: gameId,
                playerId: playerId,
                iPAddress: getClientAddress()
            }
            
            
            const response = await PlayerLogin(fetch, request);

            if(!response.ok) {
                return fail(response.error.status, { 
                    success: false, 
                    message: response.error.title || 'Login failed due to a server or network error.' 
                });
            }

            if(!response.data || !response.data.token) {
                return fail(500, { 
                    success: false, 
                    message: 'Token missing from response.' 
                });
            }

            const maxCookieAge: number = 60 * 60 * 12 // Half a day

            cookies.set('auth_token', response.data.token, {
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

        // Tell the player hub that a new player joined the game.
        redirect(303, '/me');
    }
};

export const load = (async ({fetch, params, cookies}) => {
    const playerId = cookies.get('player_id');
    if(playerId) {
        redirect(303, '/me');
    }

    // TODO: redirect if cookies are already set
    const request = {
        gameId: params.gameId
    }

    const response = await GetPlayersFromGame(fetch, request);
    if(!response.ok) {
        return fail(response.error.status, { 
            success: false, 
            message: response.error.title || 'Failed to load players.' 
        });
    }

    return {
        players: response.data,
        gameId: params.gameId
    };
}) satisfies PageServerLoad;