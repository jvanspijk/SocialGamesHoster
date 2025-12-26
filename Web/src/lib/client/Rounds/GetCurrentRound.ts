import { createEndpoint } from "../api";
export type GetCurrentRoundRequest = {
    readonly gameId: string;
};
export type GetCurrentRoundResponse = {
    readonly id: number;
    readonly startTime: string | null;
    readonly isFinished: boolean;
    readonly remainingSeconds: number;
    readonly isPaused: boolean;
};
export const GetCurrentRound = createEndpoint<GetCurrentRoundRequest, GetCurrentRoundResponse>('/api/games/{gameId}/rounds/current', 'GET');
