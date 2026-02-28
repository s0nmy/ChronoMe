import React, { useEffect, useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from './ui/dialog';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';
import { Textarea } from './ui/textarea';
import type { Project, ProjectFormData } from '../types';

interface ProjectDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (data: ProjectFormData) => Promise<void>;
  project?: Project;
}

const PROJECT_COLORS = [
  '#ef4444', // 赤
  '#f97316', // オレンジ
  '#eab308', // 黄色
  '#22c55e', // 緑
  '#3b82f6', // 青
  '#6366f1', // 藍
  '#8b5cf6', // 紫
  '#ec4899', // ピンク
  '#06b6d4', // シアン
  '#84cc16', // ライム
];

export function ProjectDialog({ open, onOpenChange, onSave, project }: ProjectDialogProps) {
  const [formData, setFormData] = useState<ProjectFormData>({
    name: project?.name || '',
    description: project?.description || '',
    color: project?.color || PROJECT_COLORS[0],
  });
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (open) {
      setFormData({
        name: project?.name || '',
        description: project?.description || '',
        color: project?.color || PROJECT_COLORS[0],
      });
      setError(null);
    }
  }, [project, open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.name.trim()) {
      return;
    }
    setIsSaving(true);
    setError(null);
    try {
      await onSave(formData);
      if (!project) {
        setFormData({
          name: '',
          description: '',
          color: PROJECT_COLORS[0],
        });
      }
      onOpenChange(false);
    } catch (err) {
      const message =
        err instanceof Error ? err.message : 'プロジェクトの保存に失敗しました。やり直してください。';
      setError(message);
    } finally {
      setIsSaving(false);
    }
  };

  const handleClose = () => {
    if (!isSaving) {
      onOpenChange(false);
    }
    if (!project) {
      setFormData({
        name: '',
        description: '',
        color: PROJECT_COLORS[0],
      });
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {project ? 'プロジェクトを編集' : '新しいプロジェクトを作成'}
          </DialogTitle>
          <DialogDescription>
            {project ? 'プロジェクトの詳細を変更できます' : 'プロジェクトの詳細を入力してください'}
          </DialogDescription>
        </DialogHeader>
        
        <form onSubmit={handleSubmit}>
          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">プロジェクト名 *</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                placeholder="例: Webサイト制作"
                required
              />
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="description">説明（任意）</Label>
              <Textarea
                id="description"
                value={formData.description}
                onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
                placeholder="プロジェクトの詳細説明..."
                rows={3}
              />
            </div>
            
            <div className="space-y-2">
              <Label>カラー</Label>
              <div className="flex flex-wrap gap-2">
                {PROJECT_COLORS.map((color) => (
                  <button
                    key={color}
                    type="button"
                    className={`w-8 h-8 rounded-full border-2 transition-all ${
                      formData.color === color
                        ? 'border-foreground scale-110'
                        : 'border-border hover:scale-105'
                    }`}
                    style={{ backgroundColor: color }}
                    onClick={() => setFormData(prev => ({ ...prev, color }))}
                  />
                ))}
              </div>
            </div>
          </div>
          {error && (
            <p className="text-sm text-destructive" role="alert">
              {error}
            </p>
          )}
          
          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose} disabled={isSaving}>
              キャンセル
            </Button>
            <Button type="submit" disabled={!formData.name.trim() || isSaving}>
              {isSaving ? '保存中…' : project ? '更新' : '作成'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
