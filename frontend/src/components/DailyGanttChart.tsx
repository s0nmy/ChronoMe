import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Button } from './ui/button';
import { ChevronLeft, ChevronRight, Calendar, Plus } from 'lucide-react';
import type { Entry, Project } from '../types';
import { formatDuration } from '../utils/time';
import { EntryEditDialog } from './EntryEditDialog';

const getLocalDateKey = (date: Date) => {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
};

interface DailyGanttChartProps {
  entries: Entry[];
  projects: Project[];
  selectedDate?: Date;
  onDateChange?: (date: Date) => void;
  onAddManualEntry?: (date: string, startTime?: string) => void;
  onUpdateEntry: (entryId: string, updates: Partial<Entry>) => Promise<void>;
  onDeleteEntry: (entryId: string) => Promise<void>;
  onCreateProject?: () => void;
}

interface GanttBlock {
  entry: Entry;
  project: Project;
  startHour: number;
  startMinute: number;
  durationMinutes: number;
  left: number; // パーセンテージ
  width: number; // パーセンテージ
  laneIndex: number;
}

export function DailyGanttChart({ 
  entries, 
  projects, 
  selectedDate = new Date(),
  onDateChange,
  onAddManualEntry,
  onUpdateEntry,
  onDeleteEntry,
  onCreateProject
}: DailyGanttChartProps) {
  const [currentDate, setCurrentDate] = useState(selectedDate);
  const [editingEntry, setEditingEntry] = useState<Entry | null>(null);

  const handleDateChange = (newDate: Date) => {
    setCurrentDate(newDate);
    onDateChange?.(newDate);
  };

  const goToPreviousDay = () => {
    const prevDay = new Date(currentDate);
    prevDay.setDate(prevDay.getDate() - 1);
    handleDateChange(prevDay);
  };

  const goToNextDay = () => {
    const nextDay = new Date(currentDate);
    nextDay.setDate(nextDay.getDate() + 1);
    handleDateChange(nextDay);
  };

  const goToToday = () => {
    handleDateChange(new Date());
  };

  // 選択された日のエントリを取得（ローカル日付で比較）
  const getDaySessions = () => {
    const dateKey = getLocalDateKey(currentDate);
    return entries.filter(entry => getLocalDateKey(entry.startedAt) === dateKey);
  };

  // ガントチャート用のブロック生成
  const generateGanttBlocks = (): { blocks: GanttBlock[]; laneCount: number } => {
    const daySessions = getDaySessions().sort((a, b) => a.startedAt.getTime() - b.startedAt.getTime());
    const laneEndTimes: number[] = [];

    const blocks = daySessions
      .map(entry => {
        const project = projects.find(p => p.id === entry.projectId);
        if (!project) return null;

        const startMs = entry.startedAt.getTime();
        const explicitEndMs = entry.endedAt?.getTime();
        const fallbackDurationMs =
          typeof entry.durationSec === 'number' && entry.durationSec > 0
            ? entry.durationSec * 1000
            : 5 * 60 * 1000; // デフォルト 5 分
        const endMs = explicitEndMs ?? startMs + fallbackDurationMs;

        const durationMinutes = Math.max(1, Math.round((endMs - startMs) / (60 * 1000)));

        let laneIndex = laneEndTimes.findIndex(end => end <= startMs);
        if (laneIndex === -1) {
          laneIndex = laneEndTimes.length;
          laneEndTimes.push(endMs);
        } else {
          laneEndTimes[laneIndex] = endMs;
        }

        // 0時から24時を100%として計算
        const minutesFromMidnight =
          entry.startedAt.getHours() * 60 + entry.startedAt.getMinutes() + entry.startedAt.getSeconds() / 60;
        const left = (minutesFromMidnight / (24 * 60)) * 100;
        const width = (durationMinutes / (24 * 60)) * 100;

        return {
          entry,
          project,
          startHour: entry.startedAt.getHours(),
          startMinute: entry.startedAt.getMinutes(),
          durationMinutes,
          left,
          width,
          laneIndex
        };
      })
      .filter(Boolean) as GanttBlock[];

    return {
      blocks,
      laneCount: Math.max(laneEndTimes.length, 1)
    };
  };

  // 時間軸の目盛り生成
  const generateTimeScale = () => {
    const hours = [];
    for (let i = 0; i <= 24; i += 2) {
      hours.push(i);
    }
    return hours;
  };

  const { blocks: ganttBlocks, laneCount } = generateGanttBlocks();
  const laneHeight = 28;
  const laneGap = 4;
  const timelineHeight = Math.max(64, laneCount * (laneHeight + laneGap) + laneGap);
  const timeScale = generateTimeScale();
  const daySessions = getDaySessions();
  const totalDuration = daySessions.reduce((sum, entry) => sum + entry.durationSec, 0);

  const handleTimelineClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!onAddManualEntry) return;
    
    const rect = e.currentTarget.getBoundingClientRect();
    const clickX = e.clientX - rect.left;
    const percentage = clickX / rect.width;
    const totalMinutes = percentage * 24 * 60;
    const hour = Math.floor(totalMinutes / 60);
    const minute = Math.floor(totalMinutes % 60);
    
    const timeString = `${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`;
    const dateString = getLocalDateKey(currentDate);
    
    onAddManualEntry(dateString, timeString);
  };

  return (
    <>
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Calendar className="w-5 h-5" />
            1日のガントチャート
          </CardTitle>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={goToPreviousDay}>
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <Button variant="outline" size="sm" onClick={goToToday}>
              今日
            </Button>
            <Button variant="outline" size="sm" onClick={goToNextDay}>
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <span>
            {currentDate.toLocaleDateString('ja-JP', { 
              year: 'numeric', 
              month: 'long', 
              day: 'numeric',
              weekday: 'long'
            })}
          </span>
          <span>
            合計: {formatDuration(totalDuration)} ({daySessions.length}件)
          </span>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {/* 時間軸 */}
        <div className="relative">
          <div className="flex justify-between text-xs text-muted-foreground mb-2">
            {timeScale.map(hour => (
              <span key={hour} className="min-w-0">
                {hour}:00
              </span>
            ))}
          </div>
          
          {/* タイムライン */}
          <div 
            className="relative bg-gray-100 rounded cursor-pointer hover:bg-gray-200 transition-colors"
            onClick={handleTimelineClick}
            title="クリックして手動エントリを追加"
            style={{
              height: `${timelineHeight}px`
            }}
          >
            {/* 時間の境界線 */}
            {timeScale.map(hour => (
              <div
                key={hour}
                className="absolute top-0 bottom-0 w-px bg-gray-300"
                style={{ left: `${(hour / 24) * 100}%` }}
              />
            ))}
            
          {/* エントリブロック */}
            {ganttBlocks.map((block, index) => (
              <div
                key={`${block.entry.id}-${index}`}
                className="absolute rounded px-2 py-1 text-white text-xs font-medium cursor-pointer hover:opacity-80 transition-opacity overflow-hidden flex flex-col justify-center"
                style={{
                  left: `${block.left}%`,
                  width: `${Math.max(block.width, 0.5)}%`, // 最小幅を設定
                  backgroundColor: block.project.color,
                  top: `${block.laneIndex * (laneHeight + laneGap) + laneGap}px`,
                  height: `${laneHeight}px`
                }}
                title={`${block.project.name}\n${block.startHour}:${block.startMinute
                  .toString()
                  .padStart(2, "0")} - ${formatDuration(block.entry.durationSec)}\n${
                  block.entry.notes || ""
                }`}
                onClick={(e) => {
                  e.stopPropagation();
                  setEditingEntry(block.entry);
                }}
              >
                <div className="truncate">
                  {block.project.name}
                </div>
                {laneCount === 1 && (
                  <div className="text-xs opacity-90">
                    {formatDuration(block.entry.durationSec)}
                  </div>
                )}
              </div>
            ))}
            
            {/* 現在時刻の線（今日の場合） */}
            {currentDate.toDateString() === new Date().toDateString() && (
              <div
                className="absolute top-0 bottom-0 w-0.5 bg-red-500 z-10"
                style={{
                  left: `${((new Date().getHours() * 60 + new Date().getMinutes()) / (24 * 60)) * 100}%`
                }}
              />
            )}
          </div>
          
          {onAddManualEntry && (
            <div className="mt-2 text-xs text-muted-foreground text-center">
              タイムライン上をクリックして手動エントリを追加
            </div>
          )}
        </div>

        {/* エントリ詳細 */}
        {daySessions.length > 0 ? (
          <div className="space-y-2">
            <h4 className="font-medium">エントリ詳細</h4>
            <div className="space-y-2">
              {daySessions
                .sort((a, b) => a.startedAt.getTime() - b.startedAt.getTime())
                .map(entry => {
                  const project = projects.find(p => p.id === entry.projectId);
                  if (!project) return null;
                  
                  return (
                    <div
                      key={entry.id}
                      className="flex items-center gap-3 p-2 bg-gray-50 rounded text-sm"
                    >
                      <div
                        className="w-3 h-3 rounded-full flex-shrink-0"
                        style={{ backgroundColor: project.color }}
                      />
                      <div className="flex-1">
                        <div className="font-medium">{project.name}</div>
                        <div className="text-muted-foreground">
                          {entry.notes || '作業内容なし'}
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="font-medium">
                          {formatDuration(entry.durationSec)}
                        </div>
                        <div className="text-muted-foreground">
                          {entry.startedAt.toLocaleTimeString('ja-JP', { 
                            hour: '2-digit', 
                            minute: '2-digit' 
                          })} - {entry.endedAt?.toLocaleTimeString('ja-JP', { 
                            hour: '2-digit', 
                            minute: '2-digit' 
                          })}
                        </div>
                      </div>
                    </div>
                  );
                })}
            </div>
          </div>
        ) : (
          <div className="text-center py-8 text-muted-foreground">
            <Calendar className="w-12 h-12 mx-auto mb-4 opacity-50" />
            <p>この日のエントリはありません</p>
            {onAddManualEntry && (
              <Button
                variant="outline"
                size="sm"
                className="mt-2"
                onClick={() => onAddManualEntry(getLocalDateKey(currentDate))}
              >
                <Plus className="w-4 h-4 mr-2" />
                手動でエントリを追加
              </Button>
            )}
          </div>
        )}
      </CardContent>
    </Card>

    {editingEntry && (
      <EntryEditDialog
        open
        entry={editingEntry}
        projects={projects}
        onOpenChange={(open) => {
          if (!open) {
            setEditingEntry(null);
          }
        }}
        onSave={onUpdateEntry}
        onDelete={onDeleteEntry}
        onCreateProject={onCreateProject || (() => undefined)}
      />
    )}
    </>
  );
}
