import { createEndpoint } from "../api";
export type CancelCurrentRoundRequest = {
    readonly gameId: string;
};
export const CancelCurrentRound = createEndpoint<CancelCurrentRoundRequest, void>('/api/games/{gameId}/rounds/current/cancel', 'POST');
