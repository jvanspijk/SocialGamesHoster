// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type GetRolesRequest = {
	readonly rulesetId: number;
};
export type GetRolesResponse = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
};
export const GetRoles = createEndpoint<GetRolesRequest, GetRolesResponse[]>(
	'/api/rulesets/{rulesetId}/roles',
	'GET'
);
