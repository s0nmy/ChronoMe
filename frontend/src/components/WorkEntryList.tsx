import React, { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "./ui/table";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";
import { Calendar, Edit3 } from "lucide-react";
import type { Entry, Project } from "../types";
import {
  formatDuration,
  formatTimeShort,
  isSameDay,
} from "../utils/time";
import { EntryEditDialog } from "./EntryEditDialog";

interface WorkEntryListProps {
  entries: Entry[];
  projects: Project[];
  onUpdateEntry?: (
    entryId: string,
    updates: Partial<Entry>,
  ) => Promise<void>;
  onDeleteEntry?: (entryId: string) => Promise<void>;
  onCreateProject?: () => void;
}

export function WorkEntryList({
  entries,
  projects,
  onUpdateEntry,
  onDeleteEntry,
  onCreateProject,
}: WorkEntryListProps) {
  const [editingEntry, setEditingEntry] =
    useState<Entry | null>(null);
  const todayEntries = entries.filter((entry) =>
    isSameDay(entry.startedAt, new Date()),
  );
  const totalDuration = todayEntries.reduce(
    (sum, entry) => sum + entry.durationSec,
    0,
  );

  const getProject = (id: string) =>
    projects.find((project) => project.id === id);

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Calendar className="w-5 h-5" />
              今日のエントリ
            </CardTitle>
            <p className="text-sm text-muted-foreground">
              合計 {formatDuration(totalDuration)} /{" "}
              {todayEntries.length}件
            </p>
          </div>
        </CardHeader>
        <CardContent>
          {todayEntries.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground text-sm">
              今日はまだエントリがありません
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>プロジェクト</TableHead>
                  <TableHead>メモ / タグ</TableHead>
                  <TableHead>開始</TableHead>
                  <TableHead>終了</TableHead>
                  <TableHead className="text-right">
                    時間
                  </TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {todayEntries.map((entry) => {
                  const project = getProject(entry.projectId);
                  return (
                    <TableRow key={entry.id}>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <div
                            className="w-3 h-3 rounded-full"
                            style={{
                              backgroundColor:
                                project?.color || "#94a3b8",
                            }}
                          />
                          <span className="font-medium">
                            {project?.name ?? "未分類"}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="space-y-1">
                          <p className="text-sm">
                            {entry.notes || "メモなし"}
                          </p>
                          {entry.tags?.length ? (
                            <div className="flex flex-wrap gap-1">
                              {entry.tags.map((tag) => (
                                <Badge
                                  key={`${entry.id}-${tag}`}
                                  variant="secondary"
                                  className="text-xs"
                                >
                                  {tag}
                                </Badge>
                              ))}
                            </div>
                          ) : null}
                          {entry.isBreak && (
                            <Badge variant="outline" className="text-xs">
                              休憩
                            </Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        {formatTimeShort(entry.startedAt)}
                      </TableCell>
                      <TableCell>
                        {entry.endedAt
                          ? formatTimeShort(entry.endedAt)
                          : "-"}
                      </TableCell>
                      <TableCell className="text-right font-mono font-semibold">
                        {formatDuration(entry.durationSec)}
                      </TableCell>
                      <TableCell className="text-right">
                        {(onUpdateEntry || onDeleteEntry) && (
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => setEditingEntry(entry)}
                          >
                            <Edit3 className="w-4 h-4" />
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {editingEntry && (
        <EntryEditDialog
          open
          entry={editingEntry}
          projects={projects}
          onOpenChange={(open) =>
            !open && setEditingEntry(null)
          }
          onSave={
            onUpdateEntry
              ? (entryId, updates) => onUpdateEntry(entryId, updates)
              : async () => {}
          }
          onDelete={onDeleteEntry}
          onCreateProject={
            onCreateProject || (() => void 0)
          }
        />
      )}
    </>
  );
}
