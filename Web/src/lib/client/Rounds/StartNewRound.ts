import { createEndpoint } from "../api";
export type StartNewRoundRequest = {
    readonly gameId: string;
    readonly durationInSeconds: number;
};
export type StartNewRoundResponse = {
    readonly id: number;
    readonly startTime: string | null;
};
export const StartNewRound = createEndpoint<StartNewRoundRequest, StartNewRoundResponse>('/api/games/{gameId}/rounds', 'POST');
