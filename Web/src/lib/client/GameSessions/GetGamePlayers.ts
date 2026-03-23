import { createEndpoint } from '../api';
export type GetGamePlayersRequest = {
	readonly gameId: string;
};
export type GetGamePlayersResponse = {
	readonly id: number;
	readonly name: string;
};
export const GetGamePlayers = createEndpoint<GetGamePlayersRequest, GetGamePlayersResponse[]>(
	'/api/games/{gameId}/players',
	'GET'
);
