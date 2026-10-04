import assert from "node:assert/strict";
import { test } from "node:test";
import { getCurrentElapsedTime } from "../src/features/timer/elapsedTime.ts";

const startedAt = new Date("2026-10-04T00:00:00.000Z");

test("running timers subtract accumulated pauses and round down to seconds", (t) => {
  t.mock.timers.enable({ apis: ["Date"], now: startedAt.getTime() + 90500 });
  assert.equal(getCurrentElapsedTime({ startedAt, isPaused: false, pausedDurationSec: 20 }), 70);
});

test("paused timers remain frozen regardless of the current time", (t) => {
  t.mock.timers.enable({ apis: ["Date"], now: startedAt.getTime() + 3600000 });
  const entry = {
    startedAt,
    isPaused: true,
    lastPausedAt: new Date(startedAt.getTime() + 90500),
    pausedDurationSec: 20,
  };
  assert.equal(getCurrentElapsedTime(entry), 70);
  t.mock.timers.tick(60000);
  assert.equal(getCurrentElapsedTime(entry), 70);
});

test("a paused timer without its pause timestamp retains the zero fallback", () => {
  assert.equal(getCurrentElapsedTime({ startedAt, isPaused: true, pausedDurationSec: 20 }), 0);
});
