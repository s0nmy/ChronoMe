import type { Entry, Project, ProjectFormData, Tag, User } from '../types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';
const CSRF_COOKIE_NAME = 'chronome_csrf';
const CSRF_HEADER_NAME = 'X-CSRF-Token';

export class APIError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = 'APIError';
    this.status = status;
  }
}

interface RequestOptions extends RequestInit {
  skipCsrf?: boolean;
}

function requiresCsrf(method: string) {
  return !['GET', 'HEAD', 'OPTIONS'].includes(method.toUpperCase());
}

function getCsrfToken(): string | null {
  if (typeof document === 'undefined') {
    return null;
  }
  const cookie = document.cookie
    .split(';')
    .map((part) => part.trim())
    .find((part) => part.startsWith(`${CSRF_COOKIE_NAME}=`));
  if (!cookie) {
    return null;
  }
  return decodeURIComponent(cookie.split('=').at(1) ?? '');
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const method = (options.method ?? 'GET').toUpperCase();
  const headers = new Headers(options.headers ?? {});
  if (!headers.has('Accept')) {
    headers.set('Accept', 'application/json');
  }
  if (options.body && !(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }
  if (!options.skipCsrf && requiresCsrf(method)) {
    const token = getCsrfToken();
    if (!token) {
      throw new APIError('CSRF token is missing. Please log in again.', 401);
    }
    headers.set(CSRF_HEADER_NAME, token);
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    method,
    headers,
    credentials: 'include',
  });

  if (!response.ok) {
    let message = `Request failed with status ${response.status}`;
    try {
      const payload = await response.json();
      if (typeof payload?.error === 'string') {
        message = payload.error;
      } else if (typeof payload?.error?.message === 'string') {
        message = payload.error.message;
      }
    } catch {
      // 何もしない
    }
    throw new APIError(message, response.status);
  }

  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

function parseDate(value: string | Date | undefined | null): Date {
  if (!value) {
    return new Date();
  }
  return value instanceof Date ? value : new Date(value);
}

function mapUser(payload: any): User {
  return {
    id: (payload?.id ?? payload?.ID) as string,
    email: (payload?.email ?? payload?.Email) as string,
    displayName: payload?.display_name ?? payload?.DisplayName ?? undefined,
    timeZone: payload?.time_zone ?? payload?.TimeZone ?? undefined,
    createdAt: parseDate(payload?.created_at ?? payload?.CreatedAt),
    updatedAt: payload?.updated_at ? parseDate(payload.updated_at) : undefined,
  };
}

function mapProject(payload: any): Project {
  return {
    id: (payload?.id ?? payload?.ID) as string,
    userId: (payload?.user_id ?? payload?.UserID) as string,
    name: (payload?.name ?? payload?.Name) as string,
    description: payload?.description ?? payload?.Description ?? undefined,
    color: (payload?.color ?? payload?.Color ?? '#3B82F6') as string,
    isArchived: Boolean(payload?.is_archived ?? payload?.IsArchived),
    createdAt: parseDate(payload?.created_at ?? payload?.CreatedAt),
    updatedAt: parseDate(payload?.updated_at ?? payload?.UpdatedAt),
  };
}

function mapTag(payload: any): Tag {
  return {
    id: (payload?.id ?? payload?.ID) as string,
    name: (payload?.name ?? payload?.Name) as string,
    color: (payload?.color ?? payload?.Color ?? '#94a3b8') as string,
    userId: (payload?.user_id ?? payload?.UserID ?? '') as string,
    createdAt: parseDate(payload?.created_at ?? payload?.CreatedAt),
    updatedAt: parseDate(payload?.updated_at ?? payload?.UpdatedAt),
  };
}

function mapEntry(payload: any): Entry {
  const tagsRaw: any[] = Array.isArray(payload?.tags ?? payload?.Tags)
    ? (payload?.tags ?? payload?.Tags)
    : [];
  return {
    id: (payload?.id ?? payload?.ID) as string,
    projectId: (payload?.project_id ?? payload?.ProjectID) ?? undefined,
    title: (payload?.title ?? payload?.Title ?? '') as string,
    notes: (payload?.notes ?? payload?.Notes) ?? undefined,
    tags: tagsRaw.map((tag) => tag?.name ?? tag?.Name).filter(Boolean),
    tagIds: tagsRaw.map((tag) => tag?.id ?? tag?.ID).filter(Boolean),
    startedAt: parseDate(payload?.started_at ?? payload?.StartedAt),
    endedAt: payload?.ended_at ?? payload?.EndedAt ? parseDate(payload?.ended_at ?? payload?.EndedAt) : null,
    durationSec: Number(payload?.duration_sec ?? payload?.DurationSec ?? 0),
    ratio: Number(payload?.ratio ?? payload?.Ratio ?? 1),
    isBreak: Boolean(payload?.is_break ?? payload?.IsBreak),
    userId: (payload?.user_id ?? payload?.UserID ?? '') as string,
    createdAt: parseDate(payload?.created_at ?? payload?.CreatedAt),
    updatedAt: parseDate(payload?.updated_at ?? payload?.UpdatedAt),
  };
}

export async function signup(params: { email: string; password: string; displayName?: string; timeZone?: string }) {
  const response = await request<{ user: any }>('/api/auth/signup', {
    method: 'POST',
    skipCsrf: true,
    body: JSON.stringify({
      email: params.email,
      password: params.password,
      display_name: params.displayName,
      time_zone: params.timeZone,
    }),
  });
  return mapUser(response.user);
}

export async function login(email: string, password: string) {
  const response = await request<{ user: any }>('/api/auth/login', {
    method: 'POST',
    skipCsrf: true,
    body: JSON.stringify({ email, password }),
  });
  return mapUser(response.user);
}

export async function fetchCurrentUser(): Promise<User | null> {
  try {
    const response = await request<{ user: any }>('/api/auth/me');
    return mapUser(response.user);
  } catch (error) {
    if (error instanceof APIError && error.status === 401) {
      return null;
    }
    throw error;
  }
}

export async function logout() {
  await request('/api/auth/logout', { method: 'POST' });
}

export async function listProjects(): Promise<Project[]> {
  const response = await request<{ projects: any[] }>('/api/projects/');
  return Array.isArray(response.projects) ? response.projects.map(mapProject) : [];
}

export async function createProject(data: ProjectFormData): Promise<Project> {
  const response = await request<{ project: any }>('/api/projects/', {
    method: 'POST',
    body: JSON.stringify({
      name: data.name,
      color: data.color,
      description: data.description ?? '',
    }),
  });
  return mapProject(response.project ?? response);
}

export async function updateProject(
  id: string,
  data: Partial<ProjectFormData & { isArchived?: boolean }>,
): Promise<Project> {
  const response = await request<{ project: any }>(`/api/projects/${id}`, {
    method: 'PATCH',
    body: JSON.stringify({
      name: data.name,
      color: data.color,
      description: data.description,
      is_archived: data.isArchived,
    }),
  });
  return mapProject(response.project ?? response);
}

export async function deleteProject(id: string) {
  await request(`/api/projects/${id}`, { method: 'DELETE' });
}

export async function listTags(): Promise<Tag[]> {
  const response = await request<{ tags: any[] }>('/api/tags/');
  return Array.isArray(response.tags) ? response.tags.map(mapTag) : [];
}

export async function createTag(data: { name: string; color: string }): Promise<Tag> {
  const response = await request<{ tag: any }>('/api/tags/', {
    method: 'POST',
    body: JSON.stringify(data),
  });
  return mapTag(response.tag ?? response);
}

export async function listEntries(params?: { from?: string; to?: string }): Promise<Entry[]> {
  const search = new URLSearchParams();
  if (params?.from) {
    search.set('from', params.from);
  }
  if (params?.to) {
    search.set('to', params.to);
  }
  const query = search.toString();
  const response = await request<{ entries: any[] }>(`/api/entries/${query ? `?${query}` : ''}`);
  return Array.isArray(response.entries) ? response.entries.map(mapEntry) : [];
}

export type EntryCreatePayload = {
  title: string;
  notes?: string;
  project_id?: string | null;
  started_at?: string;
  ended_at?: string;
  is_break?: boolean;
  ratio?: number;
  tag_ids?: string[];
};

export async function createEntry(payload: EntryCreatePayload): Promise<Entry> {
  const response = await request<{ entry: any }>('/api/entries/', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
  return mapEntry(response.entry ?? response);
}

export type EntryUpdatePayload = {
  title?: string;
  notes?: string;
  project_id?: string | null;
  started_at?: string | null;
  ended_at?: string | null;
  is_break?: boolean;
  ratio?: number;
  tag_ids?: string[] | null;
};

export async function updateEntry(id: string, payload: EntryUpdatePayload): Promise<Entry> {
  const response = await request<{ entry: any }>(`/api/entries/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  });
  return mapEntry(response.entry ?? response);
}

export async function deleteEntry(id: string) {
  await request(`/api/entries/${id}`, { method: 'DELETE' });
}

export type BootstrapData = {
  user: User | null;
  projects: Project[];
  entries: Entry[];
  tags: Tag[];
};

export async function bootstrap(): Promise<BootstrapData> {
  const user = await fetchCurrentUser();
  if (!user) {
    return { user: null, projects: [], entries: [], tags: [] };
  }
  const [projects, entries, tags] = await Promise.all([listProjects(), listEntries(), listTags()]);
  return { user, projects, entries, tags };
}

export type ApiClient = {
  signup: typeof signup;
  login: typeof login;
  fetchCurrentUser: typeof fetchCurrentUser;
  logout: typeof logout;
  listProjects: typeof listProjects;
  createProject: typeof createProject;
  updateProject: typeof updateProject;
  deleteProject: typeof deleteProject;
  listEntries: typeof listEntries;
  createEntry: typeof createEntry;
  updateEntry: typeof updateEntry;
  deleteEntry: typeof deleteEntry;
  listTags: typeof listTags;
  createTag: typeof createTag;
};

export const api: ApiClient = {
  signup,
  login,
  fetchCurrentUser,
  logout,
  listProjects,
  createProject,
  updateProject,
  deleteProject,
  listEntries,
  createEntry,
  updateEntry,
  deleteEntry,
  listTags,
  createTag,
};
