import { createEndpoint } from '../api';
export type CreateChannelRequest = {
	readonly channelId: string;
	readonly name: string;
	readonly gameId: number;
};
export type CreateChannelResponse = {
	readonly id: number;
	readonly name: string;
	readonly gameId: number;
};
export const CreateChannel = createEndpoint<CreateChannelRequest, CreateChannelResponse>(
	'/api/chat/channels/{channelId}',
	'POST'
);
