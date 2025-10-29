import type { PageServerLoad } from './$types';
import type { GetAllGameSessionsResponse } from '$lib/client';
import { getAllGames } from '$lib/client';

export const load = (async () => {
    const response = await getAllGames();
    const games: GetAllGameSessionsResponse[] = (response.data || []) as GetAllGameSessionsResponse[];
    return { games: games};
}) satisfies PageServerLoad;