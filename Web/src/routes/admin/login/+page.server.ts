import { AdminLogin } from '$lib/client/Auth/AdminLogin';
import { fail, redirect } from '@sveltejs/kit';
import { invalidate_session, set_token } from '$lib/cookie_utils';
import type { Actions } from './$types';

export const actions: Actions = {
	login: async ({ request, cookies, fetch }) => {
		const formData = await request.formData();
		const name: string | undefined = formData.get('name')?.toString().trim();
		const password: string | undefined = formData.get('password')?.toString().trim();

		if (!name || name === '') {
			return fail(400, {
				success: false,
				message: 'Name is required.'
			});
		}

		if (!password || password === '') {
			return fail(400, {
				success: false,
				message: 'Password is required.'
			});
		}

		try {
			const res = await AdminLogin(fetch, { username: name, passwordHash: password });
			if (!res.ok) {
				return fail(res.error.status, {
					success: false,
					message: res.error.title || 'Admin login failed.'
				});
			}
			set_token(cookies, res.data.token);
		} catch (error) {
			console.error('Admin login Exception:', error);
			invalidate_session(cookies);

			return fail(500, {
				success: false,
				message: 'Admin login failed.'
			});
		}

		redirect(303, '/admin');
	}
};

export async function load({ locals }) {
	if (locals.user?.role.toLocaleLowerCase() === 'admin') {
		redirect(303, '/admin');
	}
}
