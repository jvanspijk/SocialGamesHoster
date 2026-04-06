// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
import type { Participant } from './Common';

export type UpdateGameParticipantsRequest = {
	readonly gameId: number;
	readonly participantIds: number[];
};
export type UpdateGameParticipantsResponse = {
	readonly id: number;
	readonly participants: Participant[];
};
export const UpdateGameParticipants = createEndpoint<
	UpdateGameParticipantsRequest,
	UpdateGameParticipantsResponse
>('/api/games/{gameId}/players', 'PATCH');
