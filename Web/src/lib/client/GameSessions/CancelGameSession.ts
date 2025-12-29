import { createEndpoint } from '../api';
export type CancelGameSessionRequest = {
	readonly gameId: string;
};
export const CancelGameSession = createEndpoint<CancelGameSessionRequest, void>(
	'/api/games/{gameId}/cancel',
	'POST'
);
