import React, { useCallback, useEffect, useState } from "react";
import { LoginPage } from "./components/LoginPage";
import { SidebarNavigation } from "./components/SidebarNavigation";
import type { TabType } from "./components/SidebarNavigation";
import { TimerScreen } from "./components/TimerScreen";
import { EntriesScreen } from "./components/EntriesScreen";
import { ProjectsScreen } from "./components/ProjectsScreen";
import { SettingsScreen } from "./components/SettingsScreen";
import type {
  ActiveEntry,
  Entry,
  ManualEntryData,
  Project,
  ProjectFormData,
  Tag,
  User,
} from "./types";
import { generateId } from "./utils/time";
import {
  convertEntriesToExportData,
  downloadAsCSV,
} from "./utils/export";
import {
  api,
  bootstrap,
  type EntryCreatePayload,
  type EntryUpdatePayload,
} from "./lib/api";

const TAG_COLOR_POOL = [
  "#ef4444",
  "#f97316",
  "#eab308",
  "#22c55e",
  "#3b82f6",
  "#6366f1",
  "#8b5cf6",
  "#ec4899",
  "#06b6d4",
  "#84cc16",
  "#94a3b8",
];

const attachProjectsToEntries = (
  list: Entry[],
  projects: Project[],
): Entry[] =>
  list.map((entry) => {
    const project = entry.projectId
      ? projects.find((candidate) => candidate.id === entry.projectId)
      : undefined;
    if (entry.project === project) {
      return entry;
    }
    return { ...entry, project };
  });

const deriveEntryTitle = (
  notes?: string,
  projectName?: string,
): string => {
  if (notes && notes.trim().length > 0) {
    return notes.trim();
  }
  if (projectName) {
    return `${projectName}の作業`;
  }
  return "作業ログ";
};

const hashString = (value: string): number => {
  let hash = 0;
  for (let i = 0; i < value.length; i += 1) {
    hash = (hash << 5) - hash + value.charCodeAt(i);
    hash |= 0;
  }
  return hash;
};

const pickTagColor = (name: string): string => {
  if (!name) {
    return TAG_COLOR_POOL[0];
  }
  const hash = hashString(name.toLowerCase());
  const index = Math.abs(hash) % TAG_COLOR_POOL.length;
  return TAG_COLOR_POOL[index];
};

export default function App() {
  const [user, setUser] = useState<User | null>(null);
  const [entries, setEntries] = useState<Entry[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const [tags, setTags] = useState<Tag[]>([]);
  const [activeEntries, setActiveEntries] = useState<ActiveEntry[]>([]);
  const [activeTab, setActiveTab] = useState<TabType>("timer");
  const [initializing, setInitializing] = useState(true);

  const fetchCollections = useCallback(async () => {
    const [projectList, entryList, tagList] = await Promise.all([
      api.listProjects(),
      api.listEntries(),
      api.listTags(),
    ]);
    setProjects(projectList);
    setEntries(attachProjectsToEntries(entryList, projectList));
    setTags(tagList);
  }, []);

  useEffect(() => {
    const init = async () => {
      try {
        const result = await bootstrap();
        if (result.user) {
          setUser(result.user);
          setProjects(result.projects);
          setEntries(
            attachProjectsToEntries(result.entries, result.projects),
          );
          setTags(result.tags);
        }
      } catch (error) {
        console.error("Failed to initialize ChronoMe", error);
      } finally {
        setInitializing(false);
      }
    };
    void init();
  }, []);

  useEffect(() => {
    setEntries((prev) => attachProjectsToEntries(prev, projects));
  }, [projects]);

  useEffect(() => {
    let interval: NodeJS.Timeout | null = null;
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

  const ensureTagsForNames = useCallback(
    async (tagNames: string[]) => {
      const normalized = Array.from(
        new Set(tagNames.map((tag) => tag.trim()).filter(Boolean)),
      );
      if (!normalized.length) {
        return [];
      }
      const existingMap = new Map(
        tags.map((tag) => [tag.name.toLowerCase(), tag]),
      );
      const resolved: Tag[] = [];
      const created: Tag[] = [];
      for (const name of normalized) {
        const key = name.toLowerCase();
        const existing = existingMap.get(key);
        if (existing) {
          resolved.push(existing);
          continue;
        }
        const createdTag = await api.createTag({
          name,
          color: pickTagColor(name),
        });
        existingMap.set(key, createdTag);
        created.push(createdTag);
        resolved.push(createdTag);
      }
      if (created.length) {
        setTags((prev) => [...prev, ...created]);
      }
      return resolved;
    },
    [tags],
  );

  const handleLogin = async (email: string, password: string) => {
    try {
      const loggedIn = await api.login(email, password);
      setUser(loggedIn);
      await fetchCollections();
    } catch (error) {
      console.error(error);
      const message =
        error instanceof Error ? error.message : "ログインに失敗しました。時間をおいて再実行してください。";
      throw new Error(message);
    }
  };

  const handleSignup = async (email: string, password: string) => {
    try {
      const registered = await api.signup({ email, password });
      setUser(registered);
      await fetchCollections();
    } catch (error) {
      console.error(error);
      const message =
        error instanceof Error ? error.message : "サインアップに失敗しました。時間をおいて再実行してください。";
      throw new Error(message);
    }
  };

  const handleLogout = async () => {
    try {
      await api.logout();
    } catch (error) {
      console.error("Failed to logout cleanly", error);
      throw error instanceof Error
        ? error
        : new Error("ログアウトに失敗しました。");
    } finally {
      setUser(null);
      setEntries([]);
      setProjects([]);
      setTags([]);
      setActiveEntries([]);
      setActiveTab("timer");
    }
  };

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
        if (
          entry.timerId === timerId &&
          entry.isPaused &&
          entry.lastPausedAt
        ) {
          const pausedTime =
            new Date().getTime() -
            entry.lastPausedAt.getTime();
          return {
            ...entry,
            isPaused: false,
            pausedDurationSec:
              entry.pausedDurationSec +
              Math.floor(pausedTime / 1000),
            lastPausedAt: undefined,
          };
        }
        return entry;
      }),
    );
  };

  const handleUpdateActiveEntry = (
    timerId: string,
    updates: Partial<ActiveEntry>,
  ) => {
    setActiveEntries((prev) =>
      prev.map((entry) =>
        entry.timerId === timerId
          ? { ...entry, ...updates }
          : entry,
      ),
    );
  };

  const getCurrentElapsedTime = (
    entry: ActiveEntry,
  ): number => {
    if (entry.isPaused) {
      if (entry.lastPausedAt) {
        const elapsedBeforePause = Math.floor(
          (entry.lastPausedAt.getTime() -
            entry.startedAt.getTime()) /
            1000,
        );
        return (
          elapsedBeforePause - entry.pausedDurationSec
        );
      }
      return 0;
    }

    const totalElapsed = Math.floor(
      (new Date().getTime() - entry.startedAt.getTime()) /
        1000,
    );
    return totalElapsed - entry.pausedDurationSec;
  };

  const handleStopEntryTimer = async (timerId: string) => {
    if (!user) return;

    const activeEntry = activeEntries.find(
      (entry) => entry.timerId === timerId,
    );
    if (!activeEntry) return;

    const endedAt = new Date();
    try {
      const tagEntities = await ensureTagsForNames(
        activeEntry.tags || [],
      );
      const payload: EntryCreatePayload = {
        title: deriveEntryTitle(
          activeEntry.notes,
          activeEntry.projectName,
        ),
        notes: activeEntry.notes,
        project_id: activeEntry.projectId,
        started_at: activeEntry.startedAt.toISOString(),
        ended_at: endedAt.toISOString(),
        is_break: Boolean(activeEntry.isBreak),
        ratio: 1,
        tag_ids: tagEntities.map((tag) => tag.id),
      };
      const created = await api.createEntry(payload);
      const hydrated = attachProjectsToEntries(
        [created],
        projects,
      )[0];
      setEntries((prev) => [hydrated, ...prev]);
      setActiveEntries((prev) =>
        prev.filter((timer) => timer.timerId !== timerId),
      );
    } catch (error) {
      console.error(error);
      alert(
        error instanceof Error
          ? error.message
          : "エントリの保存に失敗しました。もう一度お試しください。",
      );
    }
  };

  const handleCreateManualEntry = async (
    data: ManualEntryData,
  ) => {
    if (!user) {
      throw new Error("ログインが必要です。");
    }

    const project = projects.find(
      (p) => p.id === data.projectId,
    );
    if (!project) {
      throw new Error("有効なプロジェクトを選択してください。");
    }

    const startedAt = new Date(
      `${data.date}T${data.startTime}`,
    );
    const endedAt = new Date(
      `${data.date}T${data.endTime}`,
    );
    if (endedAt <= startedAt) {
      throw new Error("終了時刻は開始時刻より後にしてください。");
    }

    const tagEntities = await ensureTagsForNames(data.tags || []);
    const payload: EntryCreatePayload = {
      title: deriveEntryTitle(data.notes, project.name),
      notes: data.notes,
      project_id: data.projectId,
      started_at: startedAt.toISOString(),
      ended_at: endedAt.toISOString(),
      is_break: Boolean(data.isBreak),
      ratio: data.ratio ?? 1,
      tag_ids: tagEntities.map((tag) => tag.id),
    };
    const created = await api.createEntry(payload);
    const hydrated = attachProjectsToEntries(
      [created],
      projects,
    )[0];
    setEntries((prev) => [hydrated, ...prev]);
  };

  const handleCreateProject = async (data: ProjectFormData) => {
    try {
      const created = await api.createProject(data);
      setProjects((prev) => [...prev, created]);
    } catch (error) {
      console.error(error);
      throw error instanceof Error
        ? error
        : new Error("プロジェクトの作成に失敗しました。");
    }
  };

  const handleUpdateProject = async (
    id: string,
    data: ProjectFormData,
  ) => {
    try {
      const updated = await api.updateProject(id, data);
      setProjects((prev) =>
        prev.map((project) =>
          project.id === id ? updated : project,
        ),
      );
      setEntries((prev) =>
        prev.map((entry) =>
          entry.projectId === id
            ? { ...entry, project: updated }
            : entry,
        ),
      );
    } catch (error) {
      console.error(error);
      throw error instanceof Error
        ? error
        : new Error("プロジェクトの更新に失敗しました。");
    }
  };

  const handleDeleteProject = async (id: string) => {
    try {
      await api.deleteProject(id);
      setProjects((prev) =>
        prev.filter((project) => project.id !== id),
      );
      setEntries((prev) =>
        prev.filter((entry) => entry.projectId !== id),
      );
      setActiveEntries((prev) =>
        prev.filter((entry) => entry.projectId !== id),
      );
    } catch (error) {
      console.error(error);
      throw error instanceof Error
        ? error
        : new Error("プロジェクトの削除に失敗しました。");
    }
  };

  const handleExportData = () => {
    const exportData = convertEntriesToExportData(entries, projects);
    downloadAsCSV(exportData, "chronome_entries.csv");
  };

  const handleUpdateEntry = async (
    entryId: string,
    updates: Partial<Entry>,
  ) => {
    try {
      const payload: EntryUpdatePayload = {};
      if (typeof updates.title === "string") {
        payload.title = updates.title;
      }
      if (typeof updates.notes === "string") {
        payload.notes = updates.notes;
      }
      if (typeof updates.projectId !== "undefined") {
        payload.project_id = updates.projectId;
      }
      if (updates.startedAt instanceof Date) {
        payload.started_at = updates.startedAt.toISOString();
      }
      if (updates.endedAt instanceof Date) {
        payload.ended_at = updates.endedAt.toISOString();
      } else if (updates.endedAt === null) {
        payload.ended_at = null;
      }
      if (typeof updates.isBreak === "boolean") {
        payload.is_break = updates.isBreak;
      }
      if (typeof updates.ratio === "number") {
        payload.ratio = updates.ratio;
      }
      if (updates.tags) {
        const tagEntities = await ensureTagsForNames(updates.tags);
        payload.tag_ids = tagEntities.map((tag) => tag.id);
      }
      const updated = await api.updateEntry(entryId, payload);
      const hydrated = attachProjectsToEntries(
        [updated],
        projects,
      )[0];
      setEntries((prev) =>
        prev.map((entry) =>
          entry.id === entryId ? hydrated : entry,
        ),
      );
    } catch (error) {
      console.error(error);
      throw error instanceof Error
        ? error
        : new Error("エントリの更新に失敗しました。");
    }
  };

  const handleDeleteEntry = async (entryId: string) => {
    try {
      await api.deleteEntry(entryId);
      setEntries((prev) =>
        prev.filter((entry) => entry.id !== entryId),
      );
    } catch (error) {
      console.error(error);
      throw error instanceof Error
        ? error
        : new Error("エントリの削除に失敗しました。");
    }
  };

  const handleDeleteAllData = async () => {
    try {
      await Promise.all(
        entries.map((entry) => api.deleteEntry(entry.id)),
      );
      await Promise.all(
        projects.map((project) => api.deleteProject(project.id)),
      );
      setEntries([]);
      setProjects([]);
      setTags([]);
      setActiveEntries([]);
    } catch (error) {
      console.error(error);
      throw error instanceof Error
        ? error
        : new Error("データの削除に失敗しました。");
    }
  };

  if (initializing) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <p className="text-sm text-muted-foreground">
          ChronoMe を読み込み中です…
        </p>
      </div>
    );
  }

  if (!user) {
    return (
      <LoginPage
        onLogin={handleLogin}
        onSignup={handleSignup}
      />
    );
  }

  const renderScreen = () => {
    switch (activeTab) {
      case "timer":
        return (
          <TimerScreen
            projects={projects}
            activeEntries={activeEntries}
            entries={entries}
            onStartEntry={handleStartEntryTimer}
            onPauseEntry={handlePauseEntryTimer}
            onResumeEntry={handleResumeEntryTimer}
            onStopEntry={handleStopEntryTimer}
            onUpdateActiveEntry={handleUpdateActiveEntry}
            onCreateManualEntry={handleCreateManualEntry}
            onCreateProject={handleCreateProject}
            getCurrentElapsedTime={getCurrentElapsedTime}
            onUpdateEntry={handleUpdateEntry}
            onDeleteEntry={handleDeleteEntry}
          />
        );
      case "entries":
        return (
          <EntriesScreen
            entries={entries}
            projects={projects}
            onUpdateEntry={handleUpdateEntry}
            onDeleteEntry={handleDeleteEntry}
            onCreateProject={() => setActiveTab("projects")}
          />
        );
      case "projects":
        return (
          <ProjectsScreen
            projects={projects}
            entries={entries}
            onCreateProject={handleCreateProject}
            onUpdateProject={handleUpdateProject}
            onDeleteProject={handleDeleteProject}
          />
        );
      case "settings":
        return (
          <SettingsScreen
            user={user}
            entries={entries}
            onLogout={handleLogout}
            onExportData={handleExportData}
            onDeleteAllData={handleDeleteAllData}
          />
        );
      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <SidebarNavigation
        activeTab={activeTab}
        onTabChange={setActiveTab}
      />
      <main className="flex-1 overflow-y-auto">
        <div className="max-w-6xl mx-auto p-4 md:p-6">
          {renderScreen()}
        </div>
      </main>
    </div>
  );
}
