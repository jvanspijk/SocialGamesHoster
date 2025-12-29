import { createEndpoint } from '../api';
export type PlayerLoginRequest = {
	readonly gameId: string;
	readonly playerId: number;
	readonly iPAddress: string;
};
export type PlayerLoginResponse = {
	readonly token: string;
};
export const PlayerLogin = createEndpoint<PlayerLoginRequest, PlayerLoginResponse>(
	'/api/games/{gameId}/login',
	'POST'
);
