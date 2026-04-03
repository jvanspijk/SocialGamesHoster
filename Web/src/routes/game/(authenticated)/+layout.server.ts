import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';
import { invalidate_session } from '$lib/cookie_utils';

export const load = (async ({ locals, cookies }) => {
	if (!locals.user || !locals.user.id || !locals.user.name) {
		invalidate_session(cookies);
		throw redirect(303, '/game/lobby');
	}
}) satisfies LayoutServerLoad;
