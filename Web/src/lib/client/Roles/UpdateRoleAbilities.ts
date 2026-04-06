// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
import type { AbilityInfo } from './Common';

export type UpdateRoleAbilitiesRequest = {
	readonly id: number;
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
