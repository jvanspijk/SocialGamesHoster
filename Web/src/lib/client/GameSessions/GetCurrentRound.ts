// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
import type { Phase } from './Common';

export type GetCurrentRoundRequest = {
	readonly gameId: number;
};
export type GetCurrentRoundResponse = {
	readonly id: number;
	readonly roundNumber: number;
	readonly startTime: string | null;
	readonly currentPhase: Phase | null;
};
export const GetCurrentRound = createEndpoint<GetCurrentRoundRequest, GetCurrentRoundResponse>(
	'/api/games/{gameId}/current',
	'GET'
);
