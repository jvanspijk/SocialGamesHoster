import { createEndpoint } from "../api";
export type AdjustRoundTimeRequest = {
    readonly gameId: string;
};
export const AdjustRoundTime = createEndpoint<AdjustRoundTimeRequest, void>('/api/games/{gameId}/rounds/current/time', 'PATCH');
