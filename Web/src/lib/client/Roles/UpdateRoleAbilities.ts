import { createEndpoint } from '../api';
import type { AbilityInfo } from './Common';

export type UpdateRoleAbilitiesRequest = {
	readonly id: string;
	readonly abilityIds: number[];
};
export type UpdateRoleAbilitiesResponse = {
	readonly id: number;
	readonly abilities: AbilityInfo[];
};
export const UpdateRoleAbilities = createEndpoint<
	UpdateRoleAbilitiesRequest,
	UpdateRoleAbilitiesResponse
>('/api/roles/{id}/abilities', 'PATCH');
