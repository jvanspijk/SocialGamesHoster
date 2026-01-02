import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load = (async ({ locals }) => {
	if (!locals.admin) {
		throw redirect(303, '/admin/login');
	}
}) satisfies LayoutServerLoad;
