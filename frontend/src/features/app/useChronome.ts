import { useCallback, useEffect, useMemo, useState } from "react";
import { useAuth } from "../../contexts/AuthContext";
import type { TabType } from "../../components/SidebarNavigation";
import type { Entry, ManualEntryData, Project, ProjectFormData, User } from "../../types";
import { convertEntriesToExportData, downloadAsCSV } from "../../utils/export";
import { api, bootstrap, type EntryCreatePayload, type EntryUpdatePayload } from "../../lib/api";
import { attachProjectsToEntries, deriveEntryTitle } from "../entries/entryHelpers";
import { useTags } from "../tags/useTags";
import { useEntryTimers } from "../timer/useEntryTimers";
import { getCurrentElapsedTime, getTimerEndedAt } from "../timer/elapsedTime";

export function useChronome() {
  const {
    supabaseUser,
    isAuthLoading,
    signInWithPassword,
    signUpWithPassword,
    signInWithOAuth,
    signOut,
  } = useAuth();
  const [user, setUser] = useState<User | null>(null);
  const [storedEntries, setEntries] = useState<Entry[]>([]);
  const [projects, setProjects] = useState<Project[]>([]);
  const { setTags, ensureTagsForNames } = useTags();
  const {
    activeEntries,
    handleStartEntryTimer,
    handlePauseEntryTimer,
    handleResumeEntryTimer,
    handleUpdateActiveEntry,
    removeTimer,
    removeProjectTimers,
    clearTimers,
  } = useEntryTimers(projects);
  const entries = useMemo(
    () => attachProjectsToEntries(storedEntries, projects),
    [storedEntries, projects],
  );
  const [activeTab, setActiveTab] = useState<TabType>("timer");
  const [initializing, setInitializing] = useState(true);

  const fetchCollections = useCallback(async () => {
    const [projectList, entryList, tagList] = await Promise.all([
      api.listProjects(),
      api.listEntries(),
      api.listTags(),
    ]);
    setProjects(projectList);
    setEntries(entryList);
    setTags(tagList);
  }, [setTags]);

  useEffect(() => {
    const init = async () => {
      if (isAuthLoading) {
        return;
      }
      if (!supabaseUser) {
        setUser(null);
        setProjects([]);
        setEntries([]);
        setTags([]);
        setInitializing(false);
        return;
      }
      try {
        setInitializing(true);
        const result = await bootstrap();
        if (result.user) {
          setUser(result.user);
          setProjects(result.projects);
          setEntries(result.entries);
          setTags(result.tags);
        }
      } catch (error) {
        console.error("Failed to initialize ChronoMe", error);
      } finally {
        setInitializing(false);
      }
    };
    void init();
  }, [isAuthLoading, setTags, supabaseUser]);

  const handleLogin = async (email: string, password: string) => {
    try {
      await signInWithPassword(email, password);
      const loggedIn = await api.fetchCurrentUser();
      if (!loggedIn) {
        throw new Error("ログインユーザーを取得できませんでした。");
      }
      setUser(loggedIn);
      await fetchCollections();
    } catch (error) {
      console.error(error);
      const message =
        error instanceof Error
          ? error.message
          : "ログインに失敗しました。時間をおいて再実行してください。";
      throw new Error(message);
    }
  };

  const handleSignup = async (email: string, password: string) => {
    try {
      await signUpWithPassword(email, password);
      const registered = await api.fetchCurrentUser();
      if (!registered) {
        throw new Error("確認メールを送信しました。メール確認後にログインしてください。");
      }
      setUser(registered);
      await fetchCollections();
    } catch (error) {
      console.error(error);
      const message =
        error instanceof Error
          ? error.message
          : "サインアップに失敗しました。時間をおいて再実行してください。";
      throw new Error(message);
    }
  };

  const handleOAuthLogin = async (provider: "google" | "github" | "apple") => {
    await signInWithOAuth(provider);
  };

  const handleLogout = async () => {
    try {
      await signOut();
    } catch (error) {
      console.error("Failed to logout cleanly", error);
      throw error instanceof Error ? error : new Error("ログアウトに失敗しました。");
    } finally {
      setUser(null);
      setEntries([]);
      setProjects([]);
      setTags([]);
      clearTimers();
      setActiveTab("timer");
    }
  };

  const handleStopEntryTimer = async (timerId: string) => {
    if (!user) return;

    const activeEntry = activeEntries.find((entry) => entry.timerId === timerId);
    if (!activeEntry) return;

    const stoppedAt = new Date();
    try {
      const endedAt = getTimerEndedAt(activeEntry, stoppedAt);
      const tagEntities = await ensureTagsForNames(activeEntry.tags || []);
      const payload: EntryCreatePayload = {
        title: deriveEntryTitle(activeEntry.notes, activeEntry.projectName),
        notes: activeEntry.notes,
        project_id: activeEntry.projectId,
        started_at: activeEntry.startedAt.toISOString(),
        ended_at: endedAt.toISOString(),
        is_break: Boolean(activeEntry.isBreak),
        ratio: 1,
        tag_ids: tagEntities.map((tag) => tag.id),
      };
      const created = await api.createEntry(payload);
      setEntries((prev) => [created, ...prev]);
      removeTimer(timerId);
    } catch (error) {
      console.error(error);
      alert(
        error instanceof Error
          ? error.message
          : "エントリの保存に失敗しました。もう一度お試しください。",
      );
    }
  };

  const handleCreateManualEntry = async (data: ManualEntryData) => {
    if (!user) {
      throw new Error("ログインが必要です。");
    }

    const project = projects.find((p) => p.id === data.projectId);
    if (!project) {
      throw new Error("有効なプロジェクトを選択してください。");
    }

    const startedAt = new Date(`${data.date}T${data.startTime}`);
    const endedAt = new Date(`${data.date}T${data.endTime}`);
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
    setEntries((prev) => [created, ...prev]);
  };

  const handleCreateProject = async (data: ProjectFormData) => {
    try {
      const created = await api.createProject(data);
      setProjects((prev) => [...prev, created]);
    } catch (error) {
      console.error(error);
      throw error instanceof Error ? error : new Error("プロジェクトの作成に失敗しました。");
    }
  };

  const handleUpdateProject = async (id: string, data: ProjectFormData) => {
    try {
      const updated = await api.updateProject(id, data);
      setProjects((prev) => prev.map((project) => (project.id === id ? updated : project)));
    } catch (error) {
      console.error(error);
      throw error instanceof Error ? error : new Error("プロジェクトの更新に失敗しました。");
    }
  };

  const handleDeleteProject = async (id: string) => {
    try {
      await api.deleteProject(id);
      setProjects((prev) => prev.filter((project) => project.id !== id));
      setEntries((prev) => prev.filter((entry) => entry.projectId !== id));
      removeProjectTimers(id);
    } catch (error) {
      console.error(error);
      throw error instanceof Error ? error : new Error("プロジェクトの削除に失敗しました。");
    }
  };

  const handleExportData = () => {
    const exportData = convertEntriesToExportData(entries, projects);
    downloadAsCSV(exportData, "chronome_entries.csv");
  };

  const handleUpdateEntry = async (entryId: string, updates: Partial<Entry>) => {
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
      setEntries((prev) => prev.map((entry) => (entry.id === entryId ? updated : entry)));
    } catch (error) {
      console.error(error);
      throw error instanceof Error ? error : new Error("エントリの更新に失敗しました。");
    }
  };

  const handleDeleteEntry = async (entryId: string) => {
    try {
      await api.deleteEntry(entryId);
      setEntries((prev) => prev.filter((entry) => entry.id !== entryId));
    } catch (error) {
      console.error(error);
      throw error instanceof Error ? error : new Error("エントリの削除に失敗しました。");
    }
  };

  const handleDeleteAllData = async () => {
    try {
      await Promise.all(entries.map((entry) => api.deleteEntry(entry.id)));
      await Promise.all(projects.map((project) => api.deleteProject(project.id)));
      setEntries([]);
      setProjects([]);
      setTags([]);
      clearTimers();
    } catch (error) {
      console.error(error);
      throw error instanceof Error ? error : new Error("データの削除に失敗しました。");
    }
  };

  return {
    user,
    entries,
    projects,
    activeEntries,
    activeTab,
    setActiveTab,
    initializing,
    handleLogin,
    handleSignup,
    handleOAuthLogin,
    handleLogout,
    handleStartEntryTimer,
    handlePauseEntryTimer,
    handleResumeEntryTimer,
    handleUpdateActiveEntry,
    getCurrentElapsedTime,
    handleStopEntryTimer,
    handleCreateManualEntry,
    handleCreateProject,
    handleUpdateProject,
    handleDeleteProject,
    handleExportData,
    handleUpdateEntry,
    handleDeleteEntry,
    handleDeleteAllData,
  };
}
