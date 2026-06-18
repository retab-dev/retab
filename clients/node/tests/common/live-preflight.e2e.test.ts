import { expect, test } from 'bun:test';

import { LIVE, LIVE_FATAL_REASON } from '../live.js';

test.skipIf(!LIVE && !LIVE_FATAL_REASON)(
  'live e2e preflight accepts the configured server and API key',
  () => {
    expect(LIVE_FATAL_REASON).toBe('');
  }
);
