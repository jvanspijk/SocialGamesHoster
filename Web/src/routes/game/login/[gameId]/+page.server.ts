import type { PageServerLoad } from './$types';
import { CreatePlayer, type CreatePlayerRequest } from '$lib/client/Players/CreatePlayer';
import { PlayerLogin, type PlayerLoginRequest } from '$lib/client/Auth/PlayerLogin';
import { fail, redirect } from '@sveltejs/kit';
import { set_player_cookies, invalidate_player_cookies } from '$lib/cookies.svelte';
import type { Actions } from './$types';

export const actions: Actions = {
	login: async ({ request, cookies, params, getClientAddress, fetch }) => {
		const formData = await request.formData();
		const playerNameInput: string | undefined = formData.get('playerName')?.toString().trim();
		const gameId = params.gameId;

		if (!playerNameInput || playerNameInput === '') {
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
			};

			const res = await CreatePlayer(fetch, request);
			if (!res.ok) {
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
				gameId: Number(gameId),
				playerId: playerId,
				iPAddress: getClientAddress()
			};

			const response = await PlayerLogin(fetch, request);

			if (!response.ok) {
				return fail(response.error.status, {
					success: false,
					message: response.error.title || 'Login failed due to a server or network error.'
				});
			}

			if (!response.data || !response.data.token) {
				return fail(500, {
					success: false,
					message: 'Token missing from response.'
				});
			}

			set_player_cookies(cookies, response.data.token, playerId, gameId);
		} catch (error) {
			console.error('Server Login Exception:', error);
			invalidate_player_cookies(cookies);

			return fail(500, {
				success: false,
				message: 'Login failed due to a server or network error.'
			});
		}

		redirect(303, '/game/me');
	}
};

export const load = (async ({ locals }) => {
	if (locals.user?.id && locals.user?.name) {
		redirect(303, '/game/me');
	}
}) satisfies PageServerLoad;
