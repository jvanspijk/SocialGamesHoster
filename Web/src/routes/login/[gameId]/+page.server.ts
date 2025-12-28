import type { PageServerLoad } from './$types';
import { CreatePlayer, type CreatePlayerRequest } from '$lib/client/Players/CreatePlayer';
import { GetPlayersFromGame } from '$lib/client/Players/GetPlayersFromGame';
import { PlayerLogin, type PlayerLoginRequest } from '$lib/client/Authentication/PlayerLogin';
import { fail, redirect, type Cookies } from '@sveltejs/kit';
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

            set_cookies(cookies, response.data.token, playerId, gameId);
        } catch (error) {
            console.error('Server Login Exception:', error);
            invalidate_cookies(cookies)
            
            return fail(500, { 
                success: false, 
                message: 'Login failed due to a server or network error.' 
            });
        }

        redirect(303, '/me');
    }
};

const set_cookies = (cookies: Cookies, token: string, playerId: number, gameId: string, duration: number = 60 * 60 * 12) => {
    const cookieOptions = {
        path: '/', 
        httpOnly: true,
        secure: false, 
        maxAge: duration,
        sameSite: 'lax' as const
    };

    cookies.set('auth_token', token, cookieOptions);
    cookies.set('player_id', playerId.toString(), cookieOptions);
    cookies.set('game_id', gameId, cookieOptions);            
}

const invalidate_cookies = (cookies: Cookies) => {
    const options = { path: '/', secure: false };
    cookies.delete('auth_token', options);
    cookies.delete('player_id', options);
    cookies.delete('game_id', options);
}

export const load = (async ({fetch, params, cookies}) => {
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

    const playerId = cookies.get('player_id');
    if(playerId) {
        if(response.data.find(p => p.id === Number(playerId))) {
            redirect(303, '/me');
        }
        else {
            console.debug(`Player ${playerId} not found in game ${params.gameId}`);
            invalidate_cookies(cookies);
        }        
    }

    return {
        players: response.data,
        gameId: params.gameId
    };
}) satisfies PageServerLoad;