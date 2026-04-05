import { createEndpoint } from '../api';
export type CreateGameSessionRequest = {
	readonly rulesetId: number;
};
export type CreateGameSessionResponse = {
	readonly id: number;
	readonly rulesetId: number;
};
export const CreateGameSession = createEndpoint<
	CreateGameSessionRequest,
	CreateGameSessionResponse
>('/api/games', 'POST');
