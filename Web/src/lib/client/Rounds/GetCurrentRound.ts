import { createEndpoint } from '../api';
export type GetCurrentRoundRequest = {
	readonly gameId: number;
};
export type GetCurrentRoundResponse = {
	readonly id: number;
	readonly startTime: string | null;
	readonly isFinished: boolean;
};
export const GetCurrentRound = createEndpoint<GetCurrentRoundRequest, GetCurrentRoundResponse>(
	'/api/rounds/current',
	'GET'
);
