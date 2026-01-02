import { error } from '@sveltejs/kit';
import { GetAllRulesets } from '$lib/client/Rulesets/GetAllRulesets';
import type { PageServerLoad } from '../$types';

export const load = (async ({ fetch }) => {
	const response = await GetAllRulesets(fetch);
	if (!response.ok) {
		throw error(response.error.status, {
			message: response.error.title || 'Failed to load rulesets.'
		});
	}
	return { rulesets: response.data };
}) satisfies PageServerLoad;
