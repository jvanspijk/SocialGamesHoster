// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type PlayerLogoutRequest = {
	readonly playerId: number;
	readonly token: string;
};
export const PlayerLogout = createEndpoint<PlayerLogoutRequest, void>(
	'/api/auth/player/logout',
	'POST'
);
