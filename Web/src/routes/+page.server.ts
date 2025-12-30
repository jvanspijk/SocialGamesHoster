import type { PageServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load = (async () => {
    throw redirect(302, '/game/lobby');
}) satisfies PageServerLoad;