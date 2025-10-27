import type { PageServerLoad } from './$types';
import { getPlayerById } from '$lib/client';

export const load = (async ({ cookies }) => {
    // Cookies can be tampered with client side so this should be done differently
    const playerId: number = Number(cookies.get('player_id'));

    const options = {
        path: {
            id: playerId,              
        }
    };
    const response = await getPlayerById(options);
    return {
        player: response.data       
    };
}) satisfies PageServerLoad;