import type { PageServerLoad } from './$types';
import { GetPlayer } from '$lib/client/Players/GetPlayer';
import { GetTimerState } from '$lib/client/Timers/GetTimerState';
import { error } from '@sveltejs/kit';

export const load = (async ({ fetch, cookies }) => {
    // Cookies can be tampered with client side so this should be done differently
    const playerId = cookies.get('player_id');
    if(!playerId) {
        throw error(404, {
            message: 'Login cookie missing.'
        });
    }
    
    const playerRequest = { id: playerId };
    const playerResponse = await GetPlayer(fetch, playerRequest);

    if(!playerResponse.ok) {
        throw error(playerResponse.error.status || 500, {
            message: playerResponse.error.title || 'Failed to load player.'
        });
    }

    const timerResponse = await GetTimerState(fetch);

    if(!timerResponse.ok) {
        throw error(timerResponse.error.status || 500, {
            message: timerResponse.error.title || 'Failed to load timer.'
        });
    }
    
    return {
        player: playerResponse.data,
        timer: timerResponse.data        
    };
}) satisfies PageServerLoad;