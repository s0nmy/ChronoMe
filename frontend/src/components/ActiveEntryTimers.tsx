import React, { useMemo, useState } from "react";
import { Play, Pause, Square, Plus, Zap, Timer } from "lucide-react";
import type { ActiveEntry, Project } from "../types";
import { formatDuration } from "../utils/time";
import { Button } from "./ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Badge } from "./ui/badge";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { ProjectSelect } from "./ProjectSelect";
import { parseTagInput } from "../utils/tags";
import { cn } from "./ui/utils";

interface ActiveEntryTimersProps {
  activeEntries: ActiveEntry[];
  projects: Project[];
  onStartEntry: (
    projectId: string,
    notes?: string,
    tags?: string[],
    isBreak?: boolean,
  ) => void;
  onPauseEntry: (timerId: string) => void;
  onResumeEntry: (timerId: string) => void;
  onStopEntry: (timerId: string) => Promise<void>;
  onUpdateEntry: (
    timerId: string,
    updates: Partial<ActiveEntry>,
  ) => void;
  onCreateProject: () => void;
  getCurrentElapsedTime: (entry: ActiveEntry) => number;
}

export function ActiveEntryTimers({
  activeEntries,
  projects,
  onStartEntry,
  onPauseEntry,
  onResumeEntry,
  onStopEntry,
  onUpdateEntry,
  onCreateProject,
  getCurrentElapsedTime,
}: ActiveEntryTimersProps) {
  const [showQuickStart, setShowQuickStart] = useState(false);
  const [selectedProject, setSelectedProject] = useState("");
  const [notes, setNotes] = useState("");
  const [tagInput, setTagInput] = useState("");
  const [quickStartPulse, setQuickStartPulse] = useState<string | null>(null);

  const availableProjects = useMemo(
    () => projects.filter((project) => !project.isArchived),
    [projects],
  );

  const quickStartProjects = useMemo(
    () => availableProjects.slice(0, 6),
    [availableProjects],
  );

  const getProjectAccent = (color: string, alpha = 0.18) => {
    if (!color) return "rgba(15, 23, 42, 0.08)";
    let hex = color.replace("#", "");
    if (hex.length === 3) {
      hex = hex
        .split("")
        .map((char) => char + char)
        .join("");
    }
    const r = parseInt(hex.slice(0, 2), 16);
    const g = parseInt(hex.slice(2, 4), 16);
    const b = parseInt(hex.slice(4, 6), 16);
    return `rgba(${r}, ${g}, ${b}, ${alpha})`;
  };

  const runningCount = activeEntries.filter(
    (entry) => !entry.isPaused,
  ).length;

  const totalElapsed = activeEntries.reduce(
    (sum, entry) => sum + getCurrentElapsedTime(entry),
    0,
  );

  const handleStart = () => {
    if (!selectedProject) return;
    const tags = parseTagInput(tagInput);
    onStartEntry(
      selectedProject,
      notes || undefined,
      tags.length ? tags : undefined,
    );
    setSelectedProject("");
    setNotes("");
    setTagInput("");
    setShowQuickStart(false);
  };

  const handleOneTapStart = (projectId: string) => {
    onStartEntry(projectId);
    setQuickStartPulse(projectId);
    setTimeout(() => setQuickStartPulse(null), 500);
  };

  return (
    <div className="space-y-4">
      <div className="grid gap-4 ">
        <Card className="h-full">
          <CardHeader className="flex flex-row items-center justify-between gap-4">
            <div>
              <CardTitle className="flex items-center gap-2 text-base">
                <span className="inline-flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 text-primary">
                  <Timer className="w-4 h-4" />
                </span>
                同時並行タイマー
              </CardTitle>
              <p className="text-sm text-muted-foreground">
                プロジェクト横断で複数作業を同時に追跡します
              </p>
            </div>
          <div className="text-right">
            <div className="text-xs text-muted-foreground">
              実行中 {runningCount} 件
            </div>
            <div className="font-semibold font-mono">
              {formatDuration(totalElapsed)}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {showQuickStart ? (
            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
              <div className="sm:col-span-2 lg:col-span-3">
                <Label className="text-xs">プロジェクト</Label>
                <ProjectSelect
                  projects={projects}
                  value={selectedProject}
                  onChange={setSelectedProject}
                  onCreateProject={onCreateProject}
                  placeholder="プロジェクトを選択"
                />
              </div>
              <div className="sm:col-span-2 lg:col-span-2">
                <Label className="text-xs">メモ（任意）</Label>
                <Input
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  placeholder="作業内容を簡単に入力"
                />
              </div>
              <div className="sm:col-span-2 lg:col-span-1">
                <Label className="text-xs">タグ（カンマ区切り）</Label>
                <Input
                  value={tagInput}
                  onChange={(e) => setTagInput(e.target.value)}
                  placeholder="例: 設計, レビュー"
                />
              </div>
              <div className="sm:col-span-2 lg:col-span-3 flex items-end gap-2">
                <Button onClick={handleStart}>
                  <Play className="w-4 h-4 mr-1" />
                  計測開始
                </Button>
                <Button
                  variant="ghost"
                  onClick={() => setShowQuickStart(false)}
                >
                  キャンセル
                </Button>
              </div>
            </div>
          ) : (
            <>
              {quickStartProjects.length === 0 ? (
                <div className="text-center text-sm text-muted-foreground space-y-3">
                  <p>スタートできるプロジェクトがありません。</p>
                  <Button
                    variant="outline"
                    onClick={onCreateProject}
                    size="sm"
                  >
                    <Plus className="w-4 h-4 mr-2" />
                    プロジェクトを追加
                  </Button>
                </div>
              ) : (
                <div className="grid gap-3 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
                  <button
                    type="button"
                    onClick={() => setShowQuickStart(true)}
                    className="group relative overflow-hidden rounded-xl border bg-gradient-to-br from-primary/10 to-primary/5 p-4 text-left shadow-sm transition hover:-translate-y-0.5 hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary"
                  >
                    <div className="flex items-center justify-between gap-3">
                      <div className="flex items-center gap-3 min-w-0 flex-1">
                        <span className="inline-flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground text-xs font-bold shadow-inner">
                          <Plus className="w-4 h-4" />
                        </span>
                        <div className="min-w-0 flex-1">
                          <p className="font-medium leading-tight">
                            新規タイマー
                          </p>
                          <p className="text-xs text-muted-foreground">
                            詳細設定で開始
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-1 text-xs font-semibold text-foreground/70 flex-shrink-0">
                        <span>Start</span>
                        <Play className="w-3 h-3" />
                      </div>
                    </div>
                  </button>
                  {quickStartProjects.map((project) => (
                    <button
                      key={project.id}
                      type="button"
                      onClick={() => handleOneTapStart(project.id)}
                      className={cn(
                        "group relative overflow-hidden rounded-xl border bg-card/70 p-4 text-left shadow-sm transition hover:-translate-y-0.5 hover:shadow-lg focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary",
                        quickStartPulse === project.id &&
                          "ring-2 ring-offset-2 ring-primary/40",
                      )}
                      style={{
                        borderColor: getProjectAccent(project.color, 0.8),
                      }}
                    >
                      <div className="flex items-center justify-between gap-3">
                        <div className="flex items-center gap-3 min-w-0 flex-1">
                          <span
                            className="inline-flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full text-xs font-bold text-white shadow-inner"
                            style={{
                              backgroundColor: project.color,
                            }}
                          >
                            {project.name.slice(0, 2).toUpperCase()}
                          </span>
                          <div className="min-w-0 flex-1">
                            <p className="font-medium leading-tight truncate">
                              {project.name}
                            </p>
                            {project.description && (
                              <p className="text-xs text-muted-foreground line-clamp-1">
                                {project.description}
                              </p>
                            )}
                          </div>
                        </div>
                        <div className="flex items-center gap-1 text-xs font-semibold text-foreground/70 flex-shrink-0">
                          <span>Start</span>
                          <Play className="w-3 h-3" />
                        </div>
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </>
          )}
        </CardContent>
        </Card>

        
      </div>

      {activeEntries.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground text-sm">
            現在計測中のエントリはありません。上のボタンから新しいタイマーを開始できます。
          </CardContent>
        </Card>
      ) : (
        activeEntries.map((entry) => {
          const elapsed = getCurrentElapsedTime(entry);
          return (
            <Card
              key={entry.timerId}
              className={`border-l-4 ${
                entry.isPaused ? "opacity-75" : ""
              }`}
              style={{ borderColor: entry.projectColor }}
            >
              <CardHeader className="pb-3 flex flex-row items-center justify-between gap-4">
                <div>
                  <div className="font-medium">
                    {entry.projectName}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    開始 {entry.startedAt.toLocaleTimeString("ja-JP", {
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant="secondary">
                    {formatDuration(elapsed)}
                  </Badge>
                  {entry.isBreak && (
                    <Badge variant="outline" className="text-xs">
                      休憩
                    </Badge>
                  )}
                  {entry.isPaused && (
                    <Badge variant="outline" className="text-xs">
                      一時停止
                    </Badge>
                  )}
                </div>
              </CardHeader>
              <CardContent className="pt-0 space-y-3">
                <Input
                  value={entry.notes || ""}
                  onChange={(e) =>
                    onUpdateEntry(entry.timerId, {
                      notes: e.target.value,
                    })
                  }
                  placeholder="メモを入力"
                />
                {entry.tags?.length ? (
                  <div className="flex flex-wrap gap-2">
                    {entry.tags.map((tag) => (
                      <Badge
                        key={`${entry.timerId}-${tag}`}
                        variant="outline"
                        className="text-xs"
                      >
                        {tag}
                      </Badge>
                    ))}
                  </div>
                ) : null}
                <div className="grid grid-cols-3 gap-2">
                  {entry.isPaused ? (
                    <Button
                      size="sm"
                      onClick={() =>
                        onResumeEntry(entry.timerId)
                      }
                    >
                      <Play className="w-4 h-4 mr-1" />
                      再開
                    </Button>
                  ) : (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        onPauseEntry(entry.timerId)
                      }
                    >
                      <Pause className="w-4 h-4 mr-1" />
                      一時停止
                    </Button>
                  )}
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      onUpdateEntry(entry.timerId, {
                        isBreak: !entry.isBreak,
                      })
                    }
                  >
                    {entry.isBreak ? "作業に戻る" : "休憩に切替"}
                  </Button>
                  <Button
                    variant="destructive"
                    size="sm"
                        onClick={() => {
                          void onStopEntry(entry.timerId);
                        }}
                  >
                    <Square className="w-4 h-4 mr-1" />
                    計測終了
                  </Button>
                </div>
              </CardContent>
            </Card>
          );
        })
      )}
    </div>
  );
}
