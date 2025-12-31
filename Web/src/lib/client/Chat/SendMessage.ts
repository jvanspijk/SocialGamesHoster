import { createEndpoint } from '../api';
export type SendMessageRequest = {
	readonly channelId: string;
	readonly playerId: number;
	readonly message: string;
};
export const SendMessage = createEndpoint<SendMessageRequest, void>(
	'/api/chat/channels/{channelId}/send',
	'POST'
);
