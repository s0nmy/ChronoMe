import React from 'react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { Plus, Folder } from 'lucide-react';
import type { Project } from '../types';

interface ProjectSelectProps {
  projects: Project[];
  value?: string;
  selectedProjectId?: string;
  onChange?: (value: string) => void;
  onProjectSelect?: (projectId: string) => void;
  onCreateProject: () => void;
  placeholder?: string;
  className?: string;
}

export function ProjectSelect({
  projects,
  value,
  selectedProjectId,
  onChange,
  onProjectSelect,
  onCreateProject,
  placeholder = "プロジェクトを選択",
  className
}: ProjectSelectProps) {
  const currentValue = value ?? selectedProjectId ?? undefined;
  const handleChange = onChange || onProjectSelect;
  const selectedProject = projects.find(p => p.id === currentValue);

  return (
    <div className={`space-y-2 ${className}`}>
      <div className="flex items-center justify-between">
        <label className="text-sm font-medium">プロジェクト</label>
        <Button
          variant="ghost"
          size="sm"
          onClick={onCreateProject}
          className="h-auto p-1 text-xs"
        >
          <Plus className="w-3 h-3 mr-1" />
          新規作成
        </Button>
      </div>
      
      {projects.length === 0 ? (
        <div className="text-center py-8 text-muted-foreground">
          <Folder className="w-8 h-8 mx-auto mb-2 opacity-50" />
          <p className="text-sm">プロジェクトがありません</p>
          <Button
            variant="outline"
            size="sm"
            onClick={onCreateProject}
            className="mt-2"
          >
            <Plus className="w-4 h-4 mr-1" />
            最初のプロジェクトを作成
          </Button>
        </div>
      ) : (
        <Select value={currentValue} onValueChange={handleChange}>
          <SelectTrigger>
            <SelectValue placeholder={placeholder}>
              {selectedProject && (
                <div className="flex items-center gap-2">
                  <div
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: selectedProject.color }}
                  />
                  <span>{selectedProject.name}</span>
                </div>
              )}
            </SelectValue>
          </SelectTrigger>
          <SelectContent>
            {projects.map((project) => (
              <SelectItem key={project.id} value={project.id}>
                <div className="flex items-center gap-2">
                  <div
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: project.color }}
                  />
                  <span>{project.name}</span>
                  {project.description && (
                    <Badge variant="secondary" className="text-xs">
                      {project.description}
                    </Badge>
                  )}
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}
    </div>
  );
}
