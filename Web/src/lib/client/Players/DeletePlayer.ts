import { createEndpoint } from '../api';
export type DeletePlayerRequest = {
	readonly id: string;
};
export const DeletePlayer = createEndpoint<DeletePlayerRequest, void>(
	'/api/players/{id}',
	'DELETE'
);
