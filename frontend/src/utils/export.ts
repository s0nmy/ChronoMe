/**
 * データエクスポート機能
 */

import type { Entry, ExportEntry, Project } from '../types';
import { formatDateShort, formatTimeShort, formatDuration } from './time';

/**
 * エントリデータをCSVエクスポート用に変換
 */
export function convertEntriesToExportData(entries: Entry[] = [], projects: Project[] = []): ExportEntry[] {
  const projectMap = new Map(projects.map(project => [project.id, project]));
  return entries
    .filter(entry => entry.endedAt) // 完了したエントリのみ
    .map(entry => ({
      date: formatDateShort(entry.startedAt),
      projectName:
        entry.project?.name ||
        (entry.projectId ? projectMap.get(entry.projectId)?.name : undefined) ||
        'Unknown Project',
      startTime: formatTimeShort(entry.startedAt),
      endTime: entry.endedAt ? formatTimeShort(entry.endedAt) : '',
      duration: formatDuration(entry.durationSec),
      notes: entry.notes || '',
      tags: Array.isArray(entry.tags) ? entry.tags.join(', ') : '',
      isBreak: entry.isBreak ? '休憩' : '作業',
      ratio: `${Math.round((entry.ratio ?? 1) * 100)}%`
    }))
    .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
}

/**
 * CSVファイルとしてダウンロード
 */
export function downloadAsCSV(data: ExportEntry[], filename: string = 'chronome_entries.csv') {
  const headers = ['日付', 'プロジェクト', '開始時刻', '終了時刻', '作業時間', 'メモ', 'タグ', '区分', '比率'];
  
  const csvContent = [
    headers.join(','),
    ...data.map(row => [
      row.date,
      `"${row.projectName}"`,
      row.startTime,
      row.endTime,
      row.duration,
      `"${row.notes}"`,
      `"${row.tags}"`,
      row.isBreak,
      row.ratio
    ].join(','))
  ].join('\n');

  // BOM付きUTF-8でエンコード（Excel対応）
  const bom = '\uFEFF';
  const blob = new Blob([bom + csvContent], { type: 'text/csv;charset=utf-8;' });
  
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  link.setAttribute('href', url);
  link.setAttribute('download', filename);
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

/**
 * JSON形式でエクスポート
 */
export function downloadAsJSON(entries: Entry[], filename: string = 'chronome_entries.json') {
  const jsonContent = JSON.stringify(entries, null, 2);
  const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' });
  
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  link.setAttribute('href', url);
  link.setAttribute('download', filename);
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

/**
 * 期間指定でファイル名を生成
 */
export function generateExportFilename(startDate?: Date, endDate?: Date): string {
  const today = new Date();
  
  if (!startDate && !endDate) {
    return `chronome_${today.getFullYear()}-${(today.getMonth() + 1).toString().padStart(2, '0')}-${today.getDate().toString().padStart(2, '0')}.csv`;
  }
  
  if (startDate && endDate) {
    const start = startDate.toISOString().split('T')[0];
    const end = endDate.toISOString().split('T')[0];
    return `chronome_${start}_to_${end}.csv`;
  }
  
  return 'chronome_export.csv';
}
