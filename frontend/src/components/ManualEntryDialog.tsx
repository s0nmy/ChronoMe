import React, { useState } from "react";
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
import type { ManualEntryData, Project } from "../types";
import { ProjectSelect } from "./ProjectSelect";
import { formatTags, parseTagInput } from "../utils/tags";

interface ManualEntryDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (data: ManualEntryData) => Promise<void>;
  onCreateProject: () => void;
  projects: Project[];
  initialData?: Partial<ManualEntryData>;
}

export function ManualEntryDialog({
  open,
  onOpenChange,
  onSave,
  onCreateProject,
  projects,
  initialData,
}: ManualEntryDialogProps) {
  const [formData, setFormData] = useState<ManualEntryData>({
    projectId: initialData?.projectId || "",
    date:
      initialData?.date ||
      new Date().toISOString().split("T")[0],
    startTime: initialData?.startTime || "09:00",
    endTime: initialData?.endTime || "10:00",
    notes: initialData?.notes || "",
    tags: initialData?.tags || [],
    isBreak: initialData?.isBreak || false,
    ratio: initialData?.ratio || 1,
  });
  const [tagsInput, setTagsInput] = useState(
    formatTags(initialData?.tags || []),
  );
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  React.useEffect(() => {
    if (open && initialData) {
      setFormData((prev) => ({
        ...prev,
        ...initialData,
        date: initialData.date || prev.date,
        startTime: initialData.startTime || prev.startTime,
        tags: initialData.tags || [],
      }));
      setTagsInput(formatTags(initialData.tags || []));
    }
    if (open) {
      setError(null);
    }
  }, [initialData, open]);

  const handleSave = async () => {
    if (
      !formData.projectId ||
      !formData.date ||
      !formData.startTime ||
      !formData.endTime
    ) {
      return;
    }

    const startTime = new Date(
      `${formData.date}T${formData.startTime}`,
    );
    const endTime = new Date(
      `${formData.date}T${formData.endTime}`,
    );

    if (endTime <= startTime) {
      alert("終了時刻は開始時刻より後にしてください");
      return;
    }

    const parsedTags = parseTagInput(tagsInput);
    setIsSaving(true);
    setError(null);
    try {
      await onSave({ ...formData, tags: parsedTags });
      onOpenChange(false);
      setFormData({
        projectId: "",
        date: new Date().toISOString().split("T")[0],
        startTime: "09:00",
        endTime: "10:00",
        notes: "",
        tags: [],
        isBreak: false,
        ratio: 1,
      });
      setTagsInput("");
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "エントリの保存に失敗しました。もう一度お試しください。";
      setError(message);
    } finally {
      setIsSaving(false);
    }
  };

  const durationMinutes = (() => {
    const start = new Date(
      `${formData.date}T${formData.startTime}`,
    );
    const end = new Date(
      `${formData.date}T${formData.endTime}`,
    );
    return Math.max(
      0,
      (end.getTime() - start.getTime()) / 1000 / 60,
    );
  })();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>手動エントリ</DialogTitle>
          <DialogDescription>
            タイマーを使わずにエントリを作成します
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

          <div className="flex items-center justify-between text-sm text-muted-foreground">
            <span>
              所要時間: {Math.floor(durationMinutes / 60)}時間
              {Math.round(durationMinutes % 60)}分
            </span>
            <div className="flex items-center gap-2 text-xs">
              <Label htmlFor="isBreak" className="flex items-center gap-1">
                <input
                  id="isBreak"
                  type="checkbox"
                  checked={formData.isBreak}
                  onChange={(e) =>
                    setFormData((prev) => ({
                      ...prev,
                      isBreak: e.target.checked,
                    }))
                  }
                />
                休憩として記録
              </Label>
            </div>
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
              placeholder="作業内容や補足情報"
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

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isSaving}>
            キャンセル
          </Button>
          <Button
            onClick={handleSave}
            disabled={
              isSaving ||
              !formData.projectId ||
              !formData.date ||
              !formData.startTime ||
              !formData.endTime
            }
          >
            {isSaving ? "保存中…" : "保存"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
