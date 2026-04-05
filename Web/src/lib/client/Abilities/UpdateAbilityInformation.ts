import { createEndpoint } from '../api';
export type UpdateAbilityInformationRequest = {
	readonly id: number;
	readonly newName: string | null;
	readonly newDescription: string | null;
};
export type UpdateAbilityInformationResponse = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
};
export const UpdateAbilityInformation = createEndpoint<
	UpdateAbilityInformationRequest,
	UpdateAbilityInformationResponse
>('/api/abilities/{id}', 'PATCH');
