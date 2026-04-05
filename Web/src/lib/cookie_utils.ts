import type { Cookies } from '@sveltejs/kit';

// We'll use a long duration as the max,
// since the JWT's internal 'exp' claim will handle actual expiration.
const SESSION_DURATION = 60 * 60 * 24 * 30; // 30 days

export interface TokenContent {
	sub: string; // User ID
	name?: string; // Player Name
	unique_name?: string; // Admin Name
	role: string; // "admin" or roleId string
	jti: string; // Unique token ID
	exp: number; // Expiration timestamp
}

const COOKIE_NAME = 'session_token';

const COOKIE_OPTIONS = {
	path: '/',
	httpOnly: true,
	secure: false, // For local network setup
	sameSite: 'lax' as const
};

export const set_token = (cookies: Cookies, token: string): void => {
	cookies.set(COOKIE_NAME, token, { ...COOKIE_OPTIONS, maxAge: SESSION_DURATION });
};

export const get_token = (cookies: Cookies): string | undefined => {
	return cookies.get(COOKIE_NAME);
};

export const invalidate_session = (cookies: Cookies): void => {
	cookies.delete(COOKIE_NAME, COOKIE_OPTIONS);
};

/**
 * Decodes a JWT payload without any external libraries.
 * @param token The raw JWT string
 * @returns The decoded JSON object or null if invalid
 */
export const decode_jwt = (token: string): TokenContent | null => {
	try {
		const base64Url = token.split('.')[1];
		const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');

		const jsonPayload = decodeURIComponent(
			atob(base64)
				.split('')
				.map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
				.join('')
		);

		return JSON.parse(jsonPayload) as TokenContent;
	} catch {
		console.error('JWT decode failed.');
		return null;
	}
};
