import type { PageServerLoad } from './$types';
import type { GetAllGameSessionsResponse } from '$lib/client';
import { getAllGameSessions } from '$lib/client';

export const load = (async () => {
    const response = await getAllGameSessions();
    const games: GetAllGameSessionsResponse[] = (response.data || []) as GetAllGameSessionsResponse[];
    return { games: games};
}) satisfies PageServerLoad;