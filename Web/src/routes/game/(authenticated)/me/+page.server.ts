import { GetTimerState } from '$lib/client/Timers/GetTimerState';
import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { Me } from '$lib/client/Auth/Me';
import { invalidate_session } from '$lib/tokens.svelte';

export const load = (async ({ fetch, locals, cookies }) => {
	if (!locals.user) {
        invalidate_session(cookies);
        throw redirect(302, '/game/lobby');
    }

    const [playerResponse, timerResponse] = await Promise.all([
        Me(fetch), 
        GetTimerState(fetch)
    ]);	

	if (!playerResponse.ok) {
        console.error('Player Data Error:', playerResponse.error);
        
        if (playerResponse.error?.status === 401 || playerResponse.error?.status === 404) {
            invalidate_session(cookies);
            throw redirect(302, '/game/lobby');
        }

        const message = playerResponse.error?.detail || 'Failed to load player';
        error(playerResponse.error?.status || 500, message);
    }

    if (!timerResponse.ok) {
        error(timerResponse.error?.status || 500, timerResponse.error?.title || 'Timer failed');
    }

    return {
        player: playerResponse.data,
        timer: timerResponse.data
    };
}) satisfies PageServerLoad;
