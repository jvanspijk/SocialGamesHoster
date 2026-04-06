// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type SendMessageRequest = {
	readonly channelId: number;
	readonly playerId: number;
	readonly message: string;
};
export type SendMessageResponse = {
	readonly messageId: number;
	readonly playerId: number;
	readonly channelId: number;
	readonly message: string;
};
export const SendMessage = createEndpoint<SendMessageRequest, SendMessageResponse>(
	'/api/chat/channels/{channelId}/send',
	'POST'
);
