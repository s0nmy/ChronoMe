import assert from "node:assert/strict";
import { test } from "node:test";
import { formatLocalDateInput } from "../src/utils/time.ts";

const cases = [
  ["Asia/Tokyo", "2026-10-05T23:00:00Z", "2026-10-06", "08:00"],
  ["Asia/Tokyo", "2025-12-31T15:00:00Z", "2026-01-01", "00:00"],
  ["America/Los_Angeles", "2026-10-06T02:00:00Z", "2026-10-05", "19:00"],
  ["UTC", "2026-10-06T08:00:00Z", "2026-10-06", "08:00"],
];

for (const [timeZone, timestamp, expectedDate, expectedTime] of cases) {
  test(`editing preserves the date and instant in ${timeZone}: ${timestamp}`, () => {
    const previousTZ = process.env.TZ;
    process.env.TZ = timeZone;
    try {
      const startedAt = new Date(timestamp);
      const endedAt = new Date(startedAt.getTime() + 30 * 60 * 1000);
      const date = formatLocalDateInput(startedAt);
      const startTime = startedAt.toTimeString().slice(0, 5);
      const endTime = endedAt.toTimeString().slice(0, 5);

      assert.equal(date, expectedDate);
      assert.equal(startTime, expectedTime);
      assert.equal(new Date(`${date}T${startTime}`).getTime(), startedAt.getTime());
      assert.equal(new Date(`${date}T${endTime}`).getTime(), endedAt.getTime());
    } finally {
      if (previousTZ === undefined) delete process.env.TZ;
      else process.env.TZ = previousTZ;
    }
  });
}
