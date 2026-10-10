import type { ActiveEntry } from "../../types";

export const getCurrentElapsedTime = (entry: ActiveEntry, now: Date = new Date()): number => {
  if (entry.isPaused) {
    if (entry.lastPausedAt) {
      const elapsedBeforePause = Math.floor(
        (entry.lastPausedAt.getTime() - entry.startedAt.getTime()) / 1000,
      );
      return elapsedBeforePause - entry.pausedDurationSec;
    }
    return 0;
  }

  const totalElapsed = Math.floor((now.getTime() - entry.startedAt.getTime()) / 1000);
  return totalElapsed - entry.pausedDurationSec;
};

// Persist the same whole-second duration shown by the timer, including when paused.
export const getTimerEndedAt = (entry: ActiveEntry, stoppedAt: Date): Date => {
  const elapsedSec = getCurrentElapsedTime(entry, stoppedAt);
  if (elapsedSec <= 0) {
    throw new Error("作業時間が1秒以上になってから保存してください。");
  }
  return new Date(entry.startedAt.getTime() + elapsedSec * 1000);
};
