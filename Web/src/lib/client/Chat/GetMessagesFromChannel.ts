import { createEndpoint } from '../api';
export type GetMessagesFromChannelRequest = {
	readonly channelId: number;
	readonly before: string | null;
	readonly after: string | null;
	readonly limit: number | null;
};
export type GetMessagesFromChannelResponse = {
	readonly id: string;
	readonly content: string;
	readonly sentAt: string;
	readonly senderId: number | null;
	readonly senderName: string | null;
};
export const GetMessagesFromChannel = createEndpoint<
	GetMessagesFromChannelRequest,
	GetMessagesFromChannelResponse[]
>('/api/chat/channels/{channelId}/messages', 'GET');
