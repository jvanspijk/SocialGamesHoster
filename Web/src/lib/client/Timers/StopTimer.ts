import { createEndpoint } from '../api';
export type StopTimerRequest = void;
export const StopTimer = createEndpoint<StopTimerRequest, void>('/api/timers/stop', 'POST');
