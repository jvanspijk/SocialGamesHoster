// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
import type { AbilityInfo } from './Common';

export type GetRoleRequest = {
	readonly id: number;
};
export type GetRoleResponse = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
	readonly abilities: AbilityInfo[];
};
export const GetRole = createEndpoint<GetRoleRequest, GetRoleResponse>('/api/roles/{id}', 'GET');
