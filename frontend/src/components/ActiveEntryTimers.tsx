import React, { useState } from "react";
import { Play, Pause, Square, Plus } from "lucide-react";
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

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4">
          <div>
            <CardTitle className="text-base">
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
            <div className="grid gap-4 md:grid-cols-[240px,1fr,1fr,auto]">
              <div>
                <Label className="text-xs">プロジェクト</Label>
                <ProjectSelect
                  projects={projects}
                  value={selectedProject}
                  onChange={setSelectedProject}
                  onCreateProject={onCreateProject}
                  placeholder="プロジェクトを選択"
                />
              </div>
              <div>
                <Label className="text-xs">メモ（任意）</Label>
                <Input
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  placeholder="作業内容を簡単に入力"
                />
              </div>
              <div>
                <Label className="text-xs">
                  タグ（カンマ区切り）
                </Label>
                <Input
                  value={tagInput}
                  onChange={(e) => setTagInput(e.target.value)}
                  placeholder="例: 設計, レビュー"
                />
              </div>
              <div className="flex items-end gap-2">
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
            <Button
              variant="outline"
              onClick={() => setShowQuickStart(true)}
              className="w-full"
            >
              <Plus className="w-4 h-4 mr-2" />
              新しいタイマーを開始
            </Button>
          )}
        </CardContent>
      </Card>

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
