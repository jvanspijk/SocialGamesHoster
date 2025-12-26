import { createEndpoint } from "../api";
export type CreatePlayerRequest = {
    readonly gameId: string;
    readonly name: string;
};
export type CreatePlayerResponse = {
    readonly id: number;
    readonly name: string;
};
export const CreatePlayer = createEndpoint<CreatePlayerRequest, CreatePlayerResponse>('/api/games/{gameId}/players', 'POST');
