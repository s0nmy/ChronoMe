import assert from "node:assert/strict";
import { test } from "node:test";
import { getCurrentElapsedTime, getTimerEndedAt } from "../src/features/timer/elapsedTime.ts";

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

test("saving 10 minutes work, 30 minutes pause, 10 minutes work records 20 minutes", () => {
  const entry = { startedAt, isPaused: false, pausedDurationSec: 30 * 60 };
  const stoppedAt = new Date(startedAt.getTime() + 50 * 60 * 1000);
  assert.equal(getTimerEndedAt(entry, stoppedAt).toISOString(), "2026-10-04T00:20:00.000Z");
});

test("stopping while paused excludes the ongoing pause and earlier pauses", () => {
  const entry = {
    startedAt,
    isPaused: true,
    pausedDurationSec: 30 * 60,
    lastPausedAt: new Date(startedAt.getTime() + 50 * 60 * 1000),
  };
  const stoppedAt = new Date(startedAt.getTime() + 90 * 60 * 1000);
  assert.equal(getTimerEndedAt(entry, stoppedAt).toISOString(), "2026-10-04T00:20:00.000Z");
});

test("saving excludes multiple accumulated pauses and matches displayed whole seconds", () => {
  const entry = { startedAt, isPaused: false, pausedDurationSec: 45 };
  const stoppedAt = new Date(startedAt.getTime() + 90500);
  assert.equal(getTimerEndedAt(entry, stoppedAt).getTime() - startedAt.getTime(), 45000);
});

test("saving without pauses retains the elapsed working time", () => {
  const entry = { startedAt, isPaused: false, pausedDurationSec: 0 };
  const stoppedAt = new Date(startedAt.getTime() + 600000);
  assert.equal(getTimerEndedAt(entry, stoppedAt).getTime(), stoppedAt.getTime());
});

test("saving uses the captured stop time even if tag creation takes time", (t) => {
  const entry = { startedAt, isPaused: false, pausedDurationSec: 20 };
  const stoppedAt = new Date(startedAt.getTime() + 90000);
  t.mock.timers.enable({ apis: ["Date"], now: startedAt.getTime() + 3600000 });
  assert.equal(getTimerEndedAt(entry, stoppedAt).getTime() - startedAt.getTime(), 70000);
});

test("zero working time is rejected instead of sending an invalid end timestamp", () => {
  const entry = { startedAt, isPaused: false, pausedDurationSec: 0 };
  assert.throws(() => getTimerEndedAt(entry, new Date(startedAt.getTime() + 500)), /1秒以上/);
});
