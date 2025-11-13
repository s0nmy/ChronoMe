import React, { useState } from "react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Button } from "./ui/button";
import { Badge } from "./ui/badge";
import { Plus, Pencil, Trash2 } from "lucide-react";
import type { Entry, Project, ProjectFormData } from "../types";
import { ProjectDialog } from "./ProjectDialog";
import { formatDuration } from "../utils/time";

interface ProjectsScreenProps {
  projects: Project[];
  entries: Entry[];
  onCreateProject: (data: ProjectFormData) => Promise<void>;
  onUpdateProject: (id: string, data: ProjectFormData) => Promise<void>;
  onDeleteProject: (id: string) => Promise<void>;
}

export function ProjectsScreen({
  projects,
  entries,
  onCreateProject,
  onUpdateProject,
  onDeleteProject,
}: ProjectsScreenProps) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingProject, setEditingProject] =
    useState<Project | undefined>();

  const getStats = (projectId: string) => {
    const relatedEntries = entries.filter(
      (entry) =>
        entry.projectId === projectId && entry.endedAt,
    );
    const totalDuration = relatedEntries.reduce(
      (sum, entry) => sum + entry.durationSec,
      0,
    );
    return {
      count: relatedEntries.length,
      totalDuration,
    };
  };

  const openCreate = () => {
    setEditingProject(undefined);
    setDialogOpen(true);
  };

  const openEdit = (project: Project) => {
    setEditingProject(project);
    setDialogOpen(true);
  };

  const handleDeleteProject = async (id: string) => {
    if (
      window.confirm(
        "このプロジェクトを削除しますか？関連するエントリも削除されます。",
      )
    ) {
      try {
        await onDeleteProject(id);
      } catch (error) {
        alert(
          error instanceof Error
            ? error.message
            : "プロジェクトの削除に失敗しました。",
        );
      }
    }
  };

  return (
    <div className="px-4 lg:px-8 pb-16">
      <div className="w-full max-w-6xl mx-auto space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold">
              プロジェクト
            </h1>
            <p className="text-sm text-muted-foreground">
              ChronoMe の project エンティティをメンテナンス
            </p>
          </div>
          <Button onClick={openCreate}>
            <Plus className="w-4 h-4 mr-2" />
            追加
          </Button>
        </div>

        {projects.length === 0 ? (
          <Card>
            <CardContent className="py-16 text-center text-muted-foreground">
              プロジェクトがありません。まずは 1 件作成しましょう。
            </CardContent>
          </Card>
        ) : (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {projects.map((project) => {
              const stats = getStats(project.id);
              return (
                <Card key={project.id} className="relative">
                  <CardHeader className="pb-3">
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex items-center gap-2">
                        <span
                          className="w-3 h-3 rounded-full block"
                          style={{ backgroundColor: project.color }}
                        />
                        <div>
                          <CardTitle className="text-base">
                            {project.name}
                          </CardTitle>
                          {project.description && (
                            <p className="text-sm text-muted-foreground">
                              {project.description}
                            </p>
                          )}
                        </div>
                      </div>
                      <div className="flex gap-1">
                        <Button
                          size="icon"
                          variant="ghost"
                          onClick={() => openEdit(project)}
                        >
                          <Pencil className="w-4 h-4" />
                        </Button>
                        <Button
                          size="icon"
                          variant="ghost"
                          className="text-destructive"
                          onClick={() => void handleDeleteProject(project.id)}
                        >
                          <Trash2 className="w-4 h-4" />
                        </Button>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="flex items-center justify-between">
                    <div>
                      <p className="text-xs text-muted-foreground">
                        総時間
                      </p>
                      <p className="text-lg font-semibold">
                        {formatDuration(stats.totalDuration)}
                      </p>
                    </div>
                    <Badge variant="secondary">
                      {stats.count}件
                    </Badge>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        )}
      </div>

      <ProjectDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        onSave={(data) => {
          if (editingProject) {
            return onUpdateProject(editingProject.id, data);
          }
          return onCreateProject(data);
        }}
        project={editingProject}
      />
    </div>
  );
}
