import type { PageServerLoad } from './$types';
import { GetPlayer } from '$lib/client/Players/GetPlayer';
import { error } from '@sveltejs/kit';

export const load = (async ({ fetch, cookies }) => {
    // Cookies can be tampered with client side so this should be done differently
    const playerId = cookies.get('player_id');
    if(!playerId) {
        throw error(404, {
            message: 'Login cookie missing.'
        });
    }

    const request = { id: playerId };
    const response = await GetPlayer(fetch, request);

    if(!response.ok) {
        throw error(response.error.status || 500, {
            message: response.error.title || 'Failed to load player.'
        });
    }
    
    return {
        player: response.data       
    };
}) satisfies PageServerLoad;