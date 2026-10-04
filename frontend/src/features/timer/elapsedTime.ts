import type { ActiveEntry } from "../../types";

export const getCurrentElapsedTime = (entry: ActiveEntry): number => {
  if (entry.isPaused) {
    if (entry.lastPausedAt) {
      const elapsedBeforePause = Math.floor(
        (entry.lastPausedAt.getTime() - entry.startedAt.getTime()) / 1000,
      );
      return elapsedBeforePause - entry.pausedDurationSec;
    }
    return 0;
  }

  const totalElapsed = Math.floor((new Date().getTime() - entry.startedAt.getTime()) / 1000);
  return totalElapsed - entry.pausedDurationSec;
};
