import { createEndpoint } from '../api';
export type CreateGameSessionRequest = {
	readonly rulesetId: number;
	readonly playerNames: string[];
};
export type CreateGameSessionResponse = {
	readonly id: number;
	readonly rulesetId: number;
};
export const CreateGameSession = createEndpoint<
	CreateGameSessionRequest,
	CreateGameSessionResponse
>('/api/games', 'POST');
