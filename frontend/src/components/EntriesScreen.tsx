import React, { useState } from 'react';
import { Button } from './ui/button';
import { Download } from 'lucide-react';
import { WorkEntryList } from './WorkEntryList';
import { WorkCharts } from './WorkCharts';
import { ExportDialog } from './ExportDialog';
import type { Entry, Project } from '../types';

interface EntriesScreenProps {
  entries: Entry[];
  projects: Project[];
  onUpdateEntry?: (entryId: string, updates: Partial<Entry>) => Promise<void>;
  onDeleteEntry?: (entryId: string) => Promise<void>;
  onCreateProject?: () => void;
}

export function EntriesScreen({ 
  entries, 
  projects, 
  onUpdateEntry, 
  onDeleteEntry, 
  onCreateProject 
}: EntriesScreenProps) {
  const [showExportDialog, setShowExportDialog] = useState(false);

  return (
    <div className="px-4 lg:px-8 pb-16">
      <div className="w-full max-w-6xl mx-auto">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-semibold">作業履歴</h1>
            <p className="text-sm text-muted-foreground">
              作業履歴と統計情報
            </p>
          </div>
          <Button 
            variant="outline" 
            size="sm" 
            onClick={() => setShowExportDialog(true)}
          >
            <Download className="w-4 h-4 mr-2" />
            エクスポート
          </Button>
        </div>

        <div className="space-y-6">
          {/* チャート */}
          <WorkCharts entries={entries} projects={projects} />
          
          {/* エントリ一覧 */}
          <WorkEntryList 
            entries={entries} 
            projects={projects}
            onUpdateEntry={onUpdateEntry}
            onDeleteEntry={onDeleteEntry}
            onCreateProject={onCreateProject}
          />
        </div>
      </div>

      <ExportDialog
        open={showExportDialog}
        onOpenChange={setShowExportDialog}
        entries={entries}
      />
    </div>
  );
}
