/**
 * アプリケーション全体で使用するType定義
 */

export interface User {
  id: string;
  email: string;
  displayName?: string;
  timeZone?: string;
  createdAt: Date;
  updatedAt?: Date;
}

export interface Project {
  id: string;
  name: string;
  description?: string;
  color: string;
  userId: string;
  isArchived: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export interface Tag {
  id: string;
  name: string;
  color: string;
  userId: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface Entry {
  id: string;
  projectId?: string | null;
  project?: Project;
  title: string;
  notes?: string;
  tags: string[];
  tagIds?: string[];
  startedAt: Date;
  endedAt: Date | null;
  durationSec: number; // seconds, calculated field
  ratio: number; // 1.0 = 100%
  isBreak: boolean;
  userId: string;
  createdAt: Date;
  updatedAt: Date;
}

export interface DailySummary {
  date: string; // YYYY-MM-DD
  totalDurationSec: number; // seconds
  entryCount: number;
  projects: Array<{
    projectId: string;
    projectName: string;
    durationSec: number;
  }>;
}

export interface WeeklySummary {
  weekStart: string; // YYYY-MM-DD (Monday)
  totalDurationSec: number;
  entryCount: number;
  dailyBreakdown: Array<{
    date: string;
    durationSec: number;
    entryCount: number;
  }>;
  projectBreakdown: Array<{
    projectId: string;
    projectName: string;
    durationSec: number;
  }>;
}

export interface MonthlySummary {
  month: string; // YYYY-MM
  totalDurationSec: number;
  entryCount: number;
  weeklyBreakdown: Array<{
    weekStart: string;
    durationSec: number;
  }>;
  projectBreakdown: Array<{
    projectId: string;
    projectName: string;
    durationSec: number;
  }>;
}

export interface EntryFormData {
  projectId: string;
  notes?: string;
  tags: string[];
  isBreak?: boolean;
  ratio?: number;
}

export interface ProjectFormData {
  name: string;
  description?: string;
  color: string;
}

// アクティブなエントリ（複数同時実行可能）
export interface ActiveEntry {
  timerId: string;
  projectId: string;
  projectName: string;
  projectColor: string;
  startedAt: Date;
  notes?: string;
  tags: string[];
  isPaused: boolean;
  pausedDurationSec: number; // 一時停止した時間の累計（秒）
  lastPausedAt?: Date;
  isBreak?: boolean;
}

// 手入力用のエントリデータ
export interface ManualEntryData {
  projectId: string;
  date: string; // YYYY-MM-DD
  startTime: string; // HH:MM
  endTime: string; // HH:MM
  title?: string;
  notes?: string;
  tags: string[];
  isBreak?: boolean;
  ratio?: number;
}

// エクスポート用のデータ形式
export interface ExportEntry {
  date: string;
  projectName: string;
  startTime: string;
  endTime: string;
  duration: string;
  notes: string;
  isBreak: string;
  ratio: string;
}
