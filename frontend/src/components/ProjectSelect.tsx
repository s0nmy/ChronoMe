import React from 'react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { Plus } from 'lucide-react';
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
    <div className={` ${className}`}>
      <div className="flex items-center justify-between gap-2">
        {projects.length === 0 ? (
          <>
            <div className="flex-1 text-center py-2 text-muted-foreground">
              <p className="text-sm">プロジェクトがありません</p>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={onCreateProject}
              className="h-8 w-8 p-0 flex-shrink-0"
              title="新規プロジェクト作成"
            >
              <Plus className="w-4 h-4" />
            </Button>
          </>
        ) : (
          <>
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
            <Button
              variant="ghost"
              size="sm"
              onClick={onCreateProject}
              className="h-8 w-8 p-0 flex-shrink-0"
              title="新規プロジェクト作成"
            >
              <Plus className="w-4 h-4" />
            </Button>
          </>
        )}
      </div>
    </div>
  );
}
