import { createEndpoint } from "../api";
export type FinishGameSessionRequest = {
    readonly gameId: string;
};
export const FinishGameSession = createEndpoint<FinishGameSessionRequest, void>('/api/games/{gameId}/finish', 'POST');
