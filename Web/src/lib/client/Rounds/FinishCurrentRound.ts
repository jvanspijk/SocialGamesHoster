import { createEndpoint } from "../api";
export type FinishCurrentRoundRequest = {
    readonly gameId: string;
};
export const FinishCurrentRound = createEndpoint<FinishCurrentRoundRequest, void>('/api/games/{gameId}/rounds/current/finish', 'POST');
