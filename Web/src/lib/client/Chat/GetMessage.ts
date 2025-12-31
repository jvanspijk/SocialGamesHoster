import { createEndpoint } from '../api';
export type GetMessageRequest = {
	readonly id: string;
};
export type GetMessageResponse = {
	readonly id: string;
	readonly content: string;
	readonly sentAt: string;
	readonly playerName: string;
	readonly isDeleted: boolean;
};
export const GetMessage = createEndpoint<GetMessageRequest, GetMessageResponse>(
	'/api/chat/{id}',
	'GET'
);
