import { createEndpoint } from "../api";
export type GetTimerStateRequest = void;
export type GetTimerStateResponse = {
    readonly remainingSeconds: number;
    readonly totalSeconds: number;
    readonly isRunning: boolean;
};
export const GetTimerState = createEndpoint<GetTimerStateRequest, GetTimerStateResponse>('/api/timers', 'GET');
