import { createEndpoint } from "../api";
import type { RoleInfo } from './Common';

export type GetPlayerRequest = {
    readonly id: string;
};
export type GetPlayerResponse = {
    readonly id: number;
    readonly name: string;
    readonly role: RoleInfo | null;
};
export const GetPlayer = createEndpoint<GetPlayerRequest, GetPlayerResponse>('/api/players/{id}', 'GET');
