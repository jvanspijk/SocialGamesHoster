import { createEndpoint } from "../api";
export type ResumeCurrentRoundRequest = {
    readonly gameId: string;
};
export const ResumeCurrentRound = createEndpoint<ResumeCurrentRoundRequest, void>('/api/games/{gameId}/rounds/current/resume', 'POST');
