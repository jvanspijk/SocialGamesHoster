import { createEndpoint } from "../api";
export type GetMessagesFromChannelRequest = {
    readonly channelId: number;
    readonly null: DateTime? After =;
    readonly 50: int Limit =;
};
export type GetMessagesFromChannelResponse = {
    readonly id: string;
    readonly content: string;
    readonly sentAt: string;
    readonly senderId: number | null;
    readonly senderName: string | null;
};
export const GetMessagesFromChannel = createEndpoint<GetMessagesFromChannelRequest, GetMessagesFromChannelResponse[]>('/api/chat/channels/{channelId}/messages', 'GET');
