import { createEndpoint } from '../api';
export type DeleteGameSessionRequest = {
	readonly gameId: string;
};
export const DeleteGameSession = createEndpoint<DeleteGameSessionRequest, void>(
	'/api/games/{gameId}/delete',
	'POST'
);
