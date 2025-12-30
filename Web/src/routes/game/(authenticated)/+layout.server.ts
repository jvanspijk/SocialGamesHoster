import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';
import { invalidate_player_cookies } from '$lib/cookies.svelte';

export const load = (async ({locals, cookies}) => {
    const gameId = cookies.get('game_id');
    if (!locals.user || !locals.user.id || !locals.user.name || !gameId) {        
        invalidate_player_cookies(cookies);
        throw redirect(303, '/game/lobby');
    }
}) satisfies LayoutServerLoad;