import { createEndpoint } from '../api';
import type { Participant } from './Common';

export type AddWinnersRequest = {
	readonly gameId: string;
	readonly playerIds: number[];
};
export type AddWinnersResponse = {
	readonly id: number;
	readonly winners: Participant[];
};
export const AddWinners = createEndpoint<AddWinnersRequest, AddWinnersResponse>(
	'/api/games/{gameId}/winners/add',
	'POST'
);
