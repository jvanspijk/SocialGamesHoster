import { createEndpoint } from '../api';
export type GetGameSessionRequest = {
	readonly gameId: string;
};
export type GetGameSessionResponse = {
	readonly id: number;
	readonly rulesetId: number;
	readonly status: string;
};
export const GetGameSession = createEndpoint<GetGameSessionRequest, GetGameSessionResponse>(
	'/api/games/{gameId}',
	'GET'
);
