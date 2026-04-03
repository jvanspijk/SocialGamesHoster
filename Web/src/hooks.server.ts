import type { Handle, HandleFetch } from '@sveltejs/kit';
import { decode_jwt, get_token as get_session_token } from '$lib/cookie_utils';

export const handle: Handle = async ({ event, resolve }) => {
    const start = performance.now();

    const token = get_session_token(event.cookies);
    if (!token || token === 'undefined') {
        event.locals.user = null;
    } else {
		const payload = decode_jwt(token);
		if (!payload) {
			event.locals.user = null;
		} else {
			const userId = parseInt(payload.sub);
			
			event.locals.user = {
				id: userId,
				name: payload.name ?? payload.unique_name ?? 'Unknown',
				role: payload.role
			};
		}
	}
	
    const response = await resolve(event);
    
    const end = performance.now();
    console.debug(`${event.request.method} ${event.url.pathname} - ${(end - start).toFixed(1)}ms`);
    
    return response;
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