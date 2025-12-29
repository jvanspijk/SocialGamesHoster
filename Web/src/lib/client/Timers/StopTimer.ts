import { createEndpoint } from '../api';

export const StopTimer = createEndpoint<void, void>(
    '/api/timers/stop',
    'POST'
);