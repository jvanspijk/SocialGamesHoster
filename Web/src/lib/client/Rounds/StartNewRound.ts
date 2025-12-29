import { createEndpoint } from '../api';
export type StartNewRoundRequest = {
	readonly gameId: number;
};
export type StartNewRoundResponse = {
	readonly id: number;
};
export const StartNewRound = createEndpoint<StartNewRoundRequest, StartNewRoundResponse>(
	'/api/rounds',
	'POST'
);
