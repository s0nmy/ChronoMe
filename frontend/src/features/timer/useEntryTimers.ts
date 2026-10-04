import { useEffect, useState } from "react";
import type { ActiveEntry, Project } from "../../types";
import { generateId } from "../../utils/time";

export function useEntryTimers(projects: Project[]) {
  const [activeEntries, setActiveEntries] = useState<ActiveEntry[]>([]);
  useEffect(() => {
    let interval: ReturnType<typeof setInterval> | null = null;
    if (activeEntries.length > 0) {
      interval = setInterval(() => {
        setActiveEntries((prev) => [...prev]);
      }, 1000);
    }
    return () => {
      if (interval) {
        clearInterval(interval);
      }
    };
  }, [activeEntries.length]);

  const handleStartEntryTimer = (
    projectId: string,
    notes?: string,
    tags: string[] = [],
    isBreak = false,
  ) => {
    const project = projects.find((p) => p.id === projectId);
    if (!project) return;

    const startedAt = new Date();
    const timerId = generateId();

    const newActiveEntry: ActiveEntry = {
      timerId,
      projectId,
      projectName: project.name,
      projectColor: project.color,
      startedAt,
      notes,
      tags,
      isBreak,
      isPaused: false,
      pausedDurationSec: 0,
    };

    setActiveEntries((prev) => [...prev, newActiveEntry]);
  };

  const handlePauseEntryTimer = (timerId: string) => {
    setActiveEntries((prev) =>
      prev.map((entry) =>
        entry.timerId === timerId
          ? {
              ...entry,
              isPaused: true,
              lastPausedAt: new Date(),
            }
          : entry,
      ),
    );
  };

  const handleResumeEntryTimer = (timerId: string) => {
    setActiveEntries((prev) =>
      prev.map((entry) => {
        if (entry.timerId === timerId && entry.isPaused && entry.lastPausedAt) {
          const pausedTime = new Date().getTime() - entry.lastPausedAt.getTime();
          return {
            ...entry,
            isPaused: false,
            pausedDurationSec: entry.pausedDurationSec + Math.floor(pausedTime / 1000),
            lastPausedAt: undefined,
          };
        }
        return entry;
      }),
    );
  };

  const handleUpdateActiveEntry = (timerId: string, updates: Partial<ActiveEntry>) => {
    setActiveEntries((prev) =>
      prev.map((entry) => (entry.timerId === timerId ? { ...entry, ...updates } : entry)),
    );
  };

  const removeTimer = (timerId: string) => {
    setActiveEntries((prev) => prev.filter((entry) => entry.timerId !== timerId));
  };
  const removeProjectTimers = (projectId: string) => {
    setActiveEntries((prev) => prev.filter((entry) => entry.projectId !== projectId));
  };
  const clearTimers = () => setActiveEntries([]);

  return {
    activeEntries,
    handleStartEntryTimer,
    handlePauseEntryTimer,
    handleResumeEntryTimer,
    handleUpdateActiveEntry,
    removeTimer,
    removeProjectTimers,
    clearTimers,
  };
}
