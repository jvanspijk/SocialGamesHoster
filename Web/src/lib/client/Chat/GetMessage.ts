import { createEndpoint } from '../api';
export type GetMessageRequest = {
	readonly id: string;
};
export type GetMessageResponse = {
	readonly id: string;
	readonly content: string;
	readonly sentAt: string;
	readonly senderName: string;
	readonly senderId: number | null;
	readonly isDeleted: boolean;
};
export const GetMessage = createEndpoint<GetMessageRequest, GetMessageResponse>(
	'/api/chat/{id}',
	'GET'
);
