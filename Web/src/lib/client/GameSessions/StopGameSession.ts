import { createEndpoint } from '../api';
export type StopGameSessionRequest = {
	readonly gameId: string;
};
export type StopGameSessionResponse = {
	readonly id: number;
	readonly status: string;
};
export const StopGameSession = createEndpoint<StopGameSessionRequest, StopGameSessionResponse>(
	'/api/games/{gameId}/stop',
	'POST'
);
