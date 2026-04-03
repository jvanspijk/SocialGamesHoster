import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { CreatePlayer } from '$lib/client/Players/CreatePlayer';
import { PlayerLogin } from '$lib/client/Auth/PlayerLogin';
import { set_token, invalidate_session } from '$lib/cookie_utils';

export const load = (async ({ locals }) => {
    if (locals.user) {
        throw redirect(303, '/game/me');
    }
}) satisfies PageServerLoad;

export const actions: Actions = {
    login: async ({ request, cookies, params, getClientAddress, fetch }) => {
        const formData = await request.formData();
        const playerNameInput = formData.get('playerName')?.toString().trim();
        const gameId = params.gameId;

        if (!playerNameInput || !gameId) {
            return fail(400, {
                success: false,
                message: 'Player name and Game ID are required.'
            });
        }

        const createRes = await CreatePlayer(fetch, {
            gameId: gameId,
            name: playerNameInput,
        });

        if (!createRes.ok) {
            return fail(createRes.error.status || 400, {
                message: createRes.error.title || 'Could not create player.',
                errors: createRes.error.errors
            });
        }
        
        const playerId = createRes.data.id;

		try {
            const loginRes = await PlayerLogin(fetch, {
                gameId: Number(gameId),
                playerId: playerId,
                iPAddress: getClientAddress()
            });

            if (!loginRes.ok) {
                return fail(loginRes.error.status || 401, {
                    message: loginRes.error.title || 'Login failed.',
                    detail: loginRes.error.detail
                });
            }

            const token = loginRes.data?.token;
            if (!token) {
                return fail(500, { message: 'Server did not return a session token.' });
            }

            set_token(cookies, token);
        } catch (error) {
            console.error('Auth Exception:', error);
            invalidate_session(cookies);
            return fail(500, { message: 'A network error occurred during login.' });
        }

        throw redirect(303, '/game/me');
    }
};