// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type FinishCurrentRoundRequest = {
	readonly gameId: number;
};
export const FinishCurrentRound = createEndpoint<FinishCurrentRoundRequest, void>(
	'/api/games/{gameId}/current/finish',
	'POST'
);
