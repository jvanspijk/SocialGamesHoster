import { createEndpoint } from "../api";
export type AdjustTimerRequest = {
    readonly deltaSeconds: number;
};
export type AdjustTimerResponse = {
    readonly remainingSeconds: number;
};
export const AdjustTimer = createEndpoint<AdjustTimerRequest, AdjustTimerResponse>('/api/timers/adjust', 'PUT');
