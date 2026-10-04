import type { Entry, Project } from "../../types";

export const attachProjectsToEntries = (list: Entry[], projects: Project[]): Entry[] => {
  const projectsById = new Map(projects.map((project) => [project.id, project]));
  return list.map((entry) => {
    const project = entry.projectId ? projectsById.get(entry.projectId) : undefined;
    if (entry.project === project) {
      return entry;
    }
    return { ...entry, project };
  });
};

export const deriveEntryTitle = (notes?: string, projectName?: string): string => {
  if (notes && notes.trim().length > 0) {
    return notes.trim();
  }
  if (projectName) {
    return `${projectName}の作業`;
  }
  return "作業ログ";
};
