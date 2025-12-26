import { createEndpoint } from "../api";
import type { RoleInfo } from './common';

export type GetFullPlayerRequest = {
    readonly id: string;
};
export type GetFullPlayerResponse = {
    readonly id: number;
    readonly name: string;
    readonly role: RoleInfo | null;
    readonly isEliminated: boolean;
    readonly canSeeIds: number[];
    readonly canBeSeenByIds: number[];
};
export const GetFullPlayer = createEndpoint<GetFullPlayerRequest, GetFullPlayerResponse>('/api/players/{id}/full', 'GET');
