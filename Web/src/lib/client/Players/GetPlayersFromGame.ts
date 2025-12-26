import { createEndpoint } from "../api";
export type GetPlayersFromGameRequest = {
    readonly gameId: string;
};
export type GetPlayersFromGameResponse = {
    readonly id: number;
    readonly name: string;
};
export const GetPlayersFromGame = createEndpoint<GetPlayersFromGameRequest, GetPlayersFromGameResponse[]>('/api/games/{gameId}/players', 'GET');
