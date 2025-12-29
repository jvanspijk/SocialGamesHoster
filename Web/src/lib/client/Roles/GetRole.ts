import { createEndpoint } from '../api';
import type { AbilityInfo } from './Common';

export type GetRoleRequest = {
	readonly id: string;
};
export type GetRoleResponse = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
	readonly abilities: AbilityInfo[];
};
export const GetRole = createEndpoint<GetRoleRequest, GetRoleResponse>('/api/roles/{id}', 'GET');
