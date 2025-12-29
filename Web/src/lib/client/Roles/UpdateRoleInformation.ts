import { createEndpoint } from '../api';
export type UpdateRoleInformationRequest = {
	readonly id: string;
	readonly name: string | null;
	readonly description: string | null;
};
export type UpdateRoleInformationResponse = {
	readonly id: number;
	readonly name: string;
	readonly description: string;
};
export const UpdateRoleInformation = createEndpoint<
	UpdateRoleInformationRequest,
	UpdateRoleInformationResponse
>('/api/roles/{id}', 'PATCH');
