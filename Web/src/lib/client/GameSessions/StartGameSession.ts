import { createEndpoint } from "../api";
export type StartGameSessionRequest = {
    readonly gameId: string;
};
export type StartGameSessionResponse = {
    readonly id: number;
    readonly status: string;
};
export const StartGameSession = createEndpoint<StartGameSessionRequest, StartGameSessionResponse>('/api/games/{gameId}/start', 'POST');
