import { GetTimerState } from '$lib/client/Timers/GetTimerState';
import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { Me } from '$lib/client/Auth/Me';
import { invalidate_player_cookies } from '$lib/cookies.svelte';


export const load = (async ({ fetch, locals, cookies }) => {
    if(!locals.user || !locals.user.id || !locals.user.name) {
        invalidate_player_cookies(cookies);
        redirect(302, '/game/lobby');
    }

    const gameId = cookies.get('game_id');
    if(!gameId) {
        invalidate_player_cookies(cookies);
        redirect(302, '/game/lobby');
    }

    const [playerResponse, timerResponse] = await Promise.all([
        Me(fetch),
        GetTimerState(fetch),
    ]);

    if(!playerResponse.ok) {
        invalidate_player_cookies(cookies);
        redirect(302, '/game/lobby');        
    }

    if (!timerResponse.ok) {
        throw error(timerResponse.error.status || 500, timerResponse.error.title || 'Timer failed');
    }   

    if(!playerResponse.data) {
        invalidate_player_cookies(cookies);
        redirect(302, '/game/lobby');
    }

    return {
        player: playerResponse.data,
        timer: timerResponse.data        
    };
}) satisfies PageServerLoad;


