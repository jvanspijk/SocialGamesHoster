import { createEndpoint } from '../api';
export type FinishCurrentRoundRequest = void;
export const FinishCurrentRound = createEndpoint<FinishCurrentRoundRequest, void>(
	'/api/rounds/current/finish',
	'POST'
);
