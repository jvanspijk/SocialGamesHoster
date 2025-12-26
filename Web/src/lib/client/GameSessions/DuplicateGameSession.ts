import { createEndpoint } from "../api";
import type { Participant } from './common';

export type DuplicateGameSessionRequest = {
    readonly gameSessionId: number;
};
export type DuplicateGameSessionResponse = {
    readonly gameSessionId: number;
    readonly rulesetId: number;
    readonly participants: Participant[];
};
export const DuplicateGameSession = createEndpoint<DuplicateGameSessionRequest, DuplicateGameSessionResponse>('/api/games/duplicate', 'POST');
