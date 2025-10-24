import type { PageServerLoad } from './$types';
import { getGamePlayers } from '$lib/client';

export const load = (async ({params}) => {
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