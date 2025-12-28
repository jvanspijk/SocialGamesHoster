import { createEndpoint } from "../api";
export type StartTimerRequest = {
    readonly durationSeconds: number;
};
export type StartTimerResponse = {
    readonly remainingSeconds: number;
};
export const StartTimer = createEndpoint<StartTimerRequest, StartTimerResponse>('/api/timers/start', 'POST');
