// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type UpdatePlayerRequest = {
	readonly id: number;
	readonly newName: string | null;
	readonly newRoleId: number | null;
};
export type UpdatePlayerResponse = {
	readonly id: number;
	readonly name: string;
	readonly roleId: number | null;
};
export const UpdatePlayer = createEndpoint<UpdatePlayerRequest, UpdatePlayerResponse>(
	'/api/players/{id}',
	'PATCH'
);
