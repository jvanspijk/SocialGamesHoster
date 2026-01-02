import type { Handle, HandleFetch } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.user = null;
	const playerToken = event.cookies.get('player_token');
	const adminToken = event.cookies.get('admin_token');
	if (playerToken) {
		event.locals.user = parseToken(playerToken);
	}
	if (adminToken) {
		event.locals.admin = parseToken(adminToken);
	}
	const endTotal = performance.now();
	const resol = await resolve(event);
	const resolEnd = performance.now();
	console.debug(`resolve ${(resolEnd - endTotal).toFixed(1)}ms`);
	return resol;
};

export const handleFetch: HandleFetch = async ({ event, request, fetch }) => {
	if (request.url.startsWith('http://localhost:9090')) {
		const cookie = event.request.headers.get('cookie');
		if (cookie) {
			request.headers.set('cookie', cookie);
		}
	}

	return fetch(request);
};

function parseToken(token: string) {
	try {
		const payloadBase64 = token.split('.')[1];
		const decodedPayload = JSON.parse(Buffer.from(payloadBase64, 'base64').toString());

		return {
			id: decodedPayload.sub || decodedPayload.nameid,
			name: decodedPayload.unique_name || decodedPayload.name,
			roleId: decodedPayload.role,
			isAdmin: false
		};
	} catch (e) {
		console.error('JWT Decode failed', e);
		return null;
	}
}
