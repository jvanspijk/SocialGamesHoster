import { createEndpoint } from '../api';
export type GetAbilityRequest = {
	readonly id: string;
};
export type GetAbilityResponse = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
};
export const GetAbility = createEndpoint<GetAbilityRequest, GetAbilityResponse>(
	'/api/abilities/{id}',
	'GET'
);
