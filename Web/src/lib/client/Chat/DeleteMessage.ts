// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type DeleteMessageRequest = {
	readonly id: number;
};
export const DeleteMessage = createEndpoint<DeleteMessageRequest, void>('/api/chat/{id}', 'POST');
