import type { PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';
import { GetAllGameSessions } from '$lib/client/GameSessions/GetAllGameSessions';

export const load = (async ({ fetch }) => {
	const response = await GetAllGameSessions(fetch);
	if (!response.ok) {
		throw error(response.error.status, {
			message: response.error.title || 'Failed to load games.'
		});
	}
	return { games: response.data };
}) satisfies PageServerLoad;
