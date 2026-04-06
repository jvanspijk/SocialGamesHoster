// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type GetAllGameSessionsRequest = void;
export type GetAllGameSessionsResponse = {
	readonly id: number;
	readonly rulesetName: string;
	readonly participantIds: number[];
	readonly status: string;
};
export const GetAllGameSessions = createEndpoint<
	GetAllGameSessionsRequest,
	GetAllGameSessionsResponse[]
>('/api/games', 'GET');
