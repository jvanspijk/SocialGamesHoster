// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type StartNewRoundRequest = {
	readonly gameId: number;
	readonly newPhaseId: number;
};
export type StartNewRoundResponse = {
	readonly id: number;
	readonly currentRound: number;
	readonly startedAt: string;
	readonly roundPhaseId: number | null;
};
export const StartNewRound = createEndpoint<StartNewRoundRequest, StartNewRoundResponse>(
	'/api/games/{gameId}/rounds',
	'POST'
);
