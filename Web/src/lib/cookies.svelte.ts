import type { Cookies } from '@sveltejs/kit';

const player_cookie_duration = 60 * 60 * 12; // 12 hours
const admin_cookie_duration = 60 * 60 * 24 * 7; // 1 week

export const set_player_cookies = (
	cookies: Cookies,
	token: string,
	playerId: number,
	gameId: string
) => {
	const cookieOptions = {
		path: '/',
		httpOnly: true,
		secure: false,
		maxAge: player_cookie_duration,
		sameSite: 'lax' as const
	};

	cookies.set('player_token', token, cookieOptions);
	cookies.set('player_id', playerId.toString(), cookieOptions);
	cookies.set('game_id', gameId, cookieOptions);
};

export const invalidate_player_cookies = (cookies: Cookies) => {
	const options = { path: '/', secure: false };
	cookies.delete('player_token', options);
	cookies.delete('player_id', options);
	cookies.delete('game_id', options);
};

export const set_admin_cookies = (cookies: Cookies, token: string) => {
	const cookieOptions = {
		path: '/',
		httpOnly: true,
		secure: false,
		maxAge: admin_cookie_duration,
		sameSite: 'lax' as const
	};

	cookies.set('admin_token', token, cookieOptions);
};

export const invalidate_admin_cookies = (cookies: Cookies) => {
	const options = { path: '/', secure: false };
	cookies.delete('admin_token', options);
};
