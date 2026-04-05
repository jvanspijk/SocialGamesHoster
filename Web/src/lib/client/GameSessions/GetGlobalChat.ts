import { createEndpoint } from '../api';
export type GetGlobalChatRequest = {
	readonly gameId: number;
};
export type GetGlobalChatResponse = {
	readonly channelId: number;
	readonly channelName: string;
};
export const GetGlobalChat = createEndpoint<GetGlobalChatRequest, GetGlobalChatResponse>(
	'/api/games/{gameId}/chat/global',
	'GET'
);
