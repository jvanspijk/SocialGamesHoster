// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type StartGameSessionRequest = {
	readonly gameId: number;
};
export type StartGameSessionResponse = {
	readonly id: number;
	readonly status: string;
};
export const StartGameSession = createEndpoint<StartGameSessionRequest, StartGameSessionResponse>(
	'/api/games/{gameId}/start',
	'POST'
);
