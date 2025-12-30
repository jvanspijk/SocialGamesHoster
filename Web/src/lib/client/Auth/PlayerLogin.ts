import { createEndpoint } from '../api';
export type PlayerLoginRequest = {
	readonly playerId: number;
	readonly iPAddress: string;
	readonly gameId: number;
};
export type PlayerLoginResponse = {
	readonly token: string;
};
export const PlayerLogin = createEndpoint<PlayerLoginRequest, PlayerLoginResponse>(
	'/api/auth/player/login',
	'POST'
);
