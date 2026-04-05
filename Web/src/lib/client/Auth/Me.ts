import { createEndpoint } from '../api';
import type { RoleInfo } from './Common';

export type MeRequest = void;
export type MeResponse = {
	readonly id: number;
	readonly name: string;
	readonly gameId: number | null;
	readonly role: RoleInfo | null;
};
export const Me = createEndpoint<MeRequest, MeResponse>('/api/auth/me', 'GET');
