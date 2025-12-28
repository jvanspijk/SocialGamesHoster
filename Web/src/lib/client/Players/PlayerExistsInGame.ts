import { createEndpoint } from "../api";
export type PlayerExistsInGameRequest = {
    readonly id: string;
    readonly gameId: string;
};
export type PlayerExistsInGameResponse = {
    readonly exists: boolean;
};
export const PlayerExistsInGame = createEndpoint<PlayerExistsInGameRequest, PlayerExistsInGameResponse>('/api/players/{id}/{gameId}/exists', 'GET');
