import React, { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "./ui/dialog";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Textarea } from "./ui/textarea";
import type { Entry, Project } from "../types";
import { ProjectSelect } from "./ProjectSelect";
import { formatTags, parseTagInput } from "../utils/tags";

interface EntryEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (entryId: string, updates: Partial<Entry>) => Promise<void>;
  onDelete: (entryId: string) => Promise<void>;
  onCreateProject: () => void;
  entry: Entry | null;
  projects: Project[];
}

export function EntryEditDialog({
  open,
  onOpenChange,
  onSave,
  onDelete,
  onCreateProject,
  entry,
  projects,
}: EntryEditDialogProps) {
  const [formData, setFormData] = useState({
    projectId: "",
    date: "",
    startTime: "",
    endTime: "",
    notes: "",
    tags: [] as string[],
    isBreak: false,
    ratio: 1,
  });
  const [tagsInput, setTagsInput] = useState("");
  const [isSaving, setIsSaving] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (entry && open) {
      const startedAt = entry.startedAt;
      const endedAt = entry.endedAt || entry.startedAt;
      setFormData({
        projectId: entry.projectId,
        date: startedAt.toISOString().split("T")[0],
        startTime: startedAt.toTimeString().slice(0, 5),
        endTime: endedAt.toTimeString().slice(0, 5),
        notes: entry.notes || "",
        tags: entry.tags || [],
        isBreak: entry.isBreak,
        ratio: entry.ratio,
      });
      setTagsInput(formatTags(entry.tags || []));
    }
    if (open) {
      setError(null);
      setIsSaving(false);
      setIsDeleting(false);
    }
  }, [entry, open]);

  if (!entry) return null;

  const handleSave = async () => {
    if (
      !formData.projectId ||
      !formData.date ||
      !formData.startTime ||
      !formData.endTime
    ) {
      return;
    }

    const startedAt = new Date(
      `${formData.date}T${formData.startTime}`,
    );
    const endedAt = new Date(
      `${formData.date}T${formData.endTime}`,
    );
    if (endedAt <= startedAt) {
      alert("終了時刻は開始時刻より後にしてください");
      return;
    }

    const durationSec = Math.floor(
      (endedAt.getTime() - startedAt.getTime()) / 1000,
    );

    const project = projects.find(
      (proj) => proj.id === formData.projectId,
    );

    setIsSaving(true);
    setError(null);
    try {
      await onSave(entry.id, {
        projectId: formData.projectId,
        project,
        startedAt,
        endedAt,
        durationSec,
        notes: formData.notes,
        tags: parseTagInput(tagsInput),
        isBreak: formData.isBreak,
        ratio: formData.ratio,
        updatedAt: new Date(),
      });
      onOpenChange(false);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "エントリの更新に失敗しました。やり直してください。";
      setError(message);
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!window.confirm("このエントリを削除しますか？")) {
      return;
    }
    setIsDeleting(true);
    setError(null);
    try {
      await onDelete(entry.id);
      onOpenChange(false);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "エントリの削除に失敗しました。時間をおいて再試行してください。";
      setError(message);
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>エントリを編集</DialogTitle>
          <DialogDescription>
            ChronoMe のエントリ属性に沿って編集します
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div>
            <Label>プロジェクト</Label>
            <ProjectSelect
              projects={projects}
              value={formData.projectId}
              onChange={(projectId) =>
                setFormData((prev) => ({ ...prev, projectId }))
              }
              onCreateProject={onCreateProject}
              placeholder="プロジェクトを選択"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>日付</Label>
              <Input
                type="date"
                value={formData.date}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    date: e.target.value,
                  }))
                }
              />
            </div>
            <div>
              <Label>比率</Label>
              <Input
                type="number"
                step="0.1"
                min="0.1"
                max="1"
                value={formData.ratio}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    ratio: Number(e.target.value),
                  }))
                }
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label>開始時刻</Label>
              <Input
                type="time"
                value={formData.startTime}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    startTime: e.target.value,
                  }))
                }
              />
            </div>
            <div>
              <Label>終了時刻</Label>
              <Input
                type="time"
                value={formData.endTime}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    endTime: e.target.value,
                  }))
                }
              />
            </div>
          </div>

          <div className="flex items-center gap-2 text-sm">
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={formData.isBreak}
                onChange={(e) =>
                  setFormData((prev) => ({
                    ...prev,
                    isBreak: e.target.checked,
                  }))
                }
              />
              休憩エントリとして扱う
            </label>
          </div>

          <div>
            <Label>メモ</Label>
            <Textarea
              value={formData.notes}
              onChange={(e) =>
                setFormData((prev) => ({
                  ...prev,
                  notes: e.target.value,
                }))
              }
              placeholder="作業内容の詳細"
              rows={3}
            />
          </div>
          <div>
            <Label>タグ（カンマ区切り）</Label>
            <Input
              value={tagsInput}
              onChange={(e) => {
                const value = e.target.value;
                setTagsInput(value);
                setFormData((prev) => ({
                  ...prev,
                  tags: parseTagInput(value),
                }));
              }}
              placeholder="例: 設計, レビュー"
            />
          </div>

        </div>

        {error && (
          <p className="text-sm text-destructive" role="alert">
            {error}
          </p>
        )}
        <DialogFooter className="gap-2">
          <Button
            variant="destructive"
            onClick={handleDelete}
            className="mr-auto"
            disabled={isDeleting || isSaving}
          >
            {isDeleting ? "削除中…" : "削除"}
          </Button>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isSaving || isDeleting}>
            キャンセル
          </Button>
          <Button onClick={handleSave} disabled={isSaving || isDeleting}>
            {isSaving ? "保存中…" : "保存"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
