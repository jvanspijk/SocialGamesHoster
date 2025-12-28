import { createEndpoint } from "../api";
import type { RoleInfo } from './Common';

export type GetPlayerFromGameRequest = {
    readonly gameId: string;
    readonly playerId: string;
};
export type GetPlayerFromGameResponse = {
    readonly id: number;
    readonly name: string;
    readonly role: RoleInfo | null;
};
export const GetPlayerFromGame = createEndpoint<GetPlayerFromGameRequest, GetPlayerFromGameResponse>('/api/games/{gameId}/players/{playerId}', 'GET');
