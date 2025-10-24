import type { PageServerLoad } from './$types';
import { getActiveGames } from '$lib/client';

export const load = (async () => {
    const response = await getActiveGames();
    return {
        games: response.data,
    };
}) satisfies PageServerLoad;