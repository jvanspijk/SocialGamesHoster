// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type GetAbilityRequest = {
	readonly id: number;
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
