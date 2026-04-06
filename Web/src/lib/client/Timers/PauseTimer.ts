// GENERATED FILE - DO NOT EDIT - SEE SDK GENERATION
import { createEndpoint } from '../api';
export type PauseTimerRequest = void;
export type PauseTimerResponse = {
	readonly remainingSeconds: number;
};
export const PauseTimer = createEndpoint<PauseTimerRequest, PauseTimerResponse>(
	'/api/timers/pause',
	'POST'
);
