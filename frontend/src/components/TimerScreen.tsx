import React, { useState } from "react";
import { ActiveEntryTimers } from "./ActiveEntryTimers";
import { ManualEntryDialog } from "./ManualEntryDialog";
import { DailyGanttChart } from "./DailyGanttChart";
import { ProjectDialog } from "./ProjectDialog";
import { Button } from "./ui/button";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "./ui/tabs";
import type {
  ActiveEntry,
  Entry,
  ManualEntryData,
  Project,
  ProjectFormData,
} from "../types";
import { Card, CardContent } from "./ui/card";
import { Timer, BarChart3, Plus } from "lucide-react";

interface TimerScreenProps {
  projects: Project[];
  activeEntries: ActiveEntry[];
  entries: Entry[];
  onStartEntry: (
    projectId: string,
    notes?: string,
    tags?: string[],
    isBreak?: boolean,
  ) => void;
  onPauseEntry: (timerId: string) => void;
  onResumeEntry: (timerId: string) => void;
  onStopEntry: (timerId: string) => Promise<void>;
  onUpdateActiveEntry: (
    timerId: string,
    updates: Partial<ActiveEntry>,
  ) => void;
  onCreateManualEntry: (data: ManualEntryData) => Promise<void>;
  onCreateProject: (data: ProjectFormData) => Promise<void>;
  getCurrentElapsedTime: (entry: ActiveEntry) => number;
}

export function TimerScreen({
  projects,
  activeEntries,
  entries,
  onStartEntry,
  onPauseEntry,
  onResumeEntry,
  onStopEntry,
  onUpdateActiveEntry,
  onCreateManualEntry,
  onCreateProject,
  getCurrentElapsedTime,
}: TimerScreenProps) {
  const [showManualDialog, setShowManualDialog] = useState(false);
  const [showProjectDialog, setShowProjectDialog] = useState(false);
  const [manualDefaults, setManualDefaults] =
    useState<Partial<ManualEntryData>>();

  const openManualDialog = (
    date: string,
    startTime: string = "09:00",
  ) => {
    setManualDefaults({ date, startTime });
    setShowManualDialog(true);
  };

  return (
    <div className="px-4 lg:px-8 pb-16">
      <div className="w-full max-w-6xl mx-auto space-y-6">
        <div className="text-center">
          <h1 className="text-xl font-semibold mb-2">
            時間記録
          </h1>
          <p className="text-sm text-muted-foreground">
            複数タイマー、ガントチャート、手動入力を組み合わせて ChronoMe
            エントリを管理します
          </p>
        </div>

        <Tabs defaultValue="timers" className="w-full">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="timers">
              <Timer className="w-4 h-4 mr-2" />
              タイマー
            </TabsTrigger>
            <TabsTrigger value="gantt">
              <BarChart3 className="w-4 h-4 mr-2" />
              ガント
            </TabsTrigger>
            <TabsTrigger value="manual">
              <Plus className="w-4 h-4 mr-2" />
              手動入力
            </TabsTrigger>
          </TabsList>

          <TabsContent value="timers" className="space-y-4">
            <ActiveEntryTimers
              activeEntries={activeEntries}
              projects={projects}
              onStartEntry={onStartEntry}
              onPauseEntry={onPauseEntry}
              onResumeEntry={onResumeEntry}
              onStopEntry={onStopEntry}
              onUpdateEntry={onUpdateActiveEntry}
              onCreateProject={() => setShowProjectDialog(true)}
              getCurrentElapsedTime={getCurrentElapsedTime}
            />
          </TabsContent>

          <TabsContent value="gantt">
            <DailyGanttChart
              entries={entries}
              projects={projects}
              onAddManualEntry={openManualDialog}
            />
          </TabsContent>

          <TabsContent value="manual">
            <Card>
              <CardContent className="p-6 text-center space-y-4">
                <Plus className="w-10 h-10 mx-auto text-muted-foreground" />
                <p className="text-sm text-muted-foreground">
                  過去のエントリを手入力する場合はこちらから
                </p>
                <Button onClick={() => setShowManualDialog(true)}>
                  <Plus className="w-4 h-4 mr-2" />
                  手動で追加
                </Button>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>

      <ManualEntryDialog
        open={showManualDialog}
        onOpenChange={setShowManualDialog}
        onSave={onCreateManualEntry}
        onCreateProject={() => setShowProjectDialog(true)}
        projects={projects}
        initialData={manualDefaults}
      />

      <ProjectDialog
        open={showProjectDialog}
        onOpenChange={setShowProjectDialog}
        onSave={onCreateProject}
      />
    </div>
  );
}
