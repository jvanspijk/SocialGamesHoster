import { createEndpoint } from "../api";
export type GetActiveGameSessionsRequest = void;
export type GetActiveGameSessionsResponse = {
    readonly id: number;
    readonly rulesetId: number;
    readonly status: string;
    readonly currentRoundNumber: number;
};
export const GetActiveGameSessions = createEndpoint<GetActiveGameSessionsRequest, GetActiveGameSessionsResponse[]>('/api/games/active', 'GET');
