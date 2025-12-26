import { createEndpoint } from "../api";
export type PauseCurrentRoundRequest = {
    readonly gameId: string;
};
export const PauseCurrentRound = createEndpoint<PauseCurrentRoundRequest, void>('/api/games/{gameId}/rounds/current/pause', 'POST');
