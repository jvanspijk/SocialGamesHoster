import type { LayoutServerLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const load = (async ({ locals }) => {
	if (!locals.user?.role || locals.user.role.toLocaleLowerCase() !== 'admin') {
		throw redirect(303, '/admin/login');
	}
}) satisfies LayoutServerLoad;
