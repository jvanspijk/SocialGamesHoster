// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type GetGamePlayersRequest = {
	readonly gameId: number;
};
export type GetGamePlayersResponse = {
	readonly id: number;
	readonly name: string;
};
export const GetGamePlayers = createEndpoint<GetGamePlayersRequest, GetGamePlayersResponse[]>(
	'/api/games/{gameId}/players',
	'GET'
);
