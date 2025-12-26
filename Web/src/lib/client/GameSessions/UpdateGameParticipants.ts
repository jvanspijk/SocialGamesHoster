import { createEndpoint } from "../api";
import type { Participant } from './common';

export type UpdateGameParticipantsRequest = {
    readonly gameId: string;
    readonly participantIds: number[];
};
export type UpdateGameParticipantsResponse = {
    readonly id: number;
    readonly participants: Participant[];
};
export const UpdateGameParticipants = createEndpoint<UpdateGameParticipantsRequest, UpdateGameParticipantsResponse>('/api/games/{gameId}/players', 'PATCH');
