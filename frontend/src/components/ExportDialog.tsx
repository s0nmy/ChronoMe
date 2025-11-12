import React, { useState } from 'react';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from './ui/dialog';
import { Button } from './ui/button';
import { Label } from './ui/label';
import { RadioGroup, RadioGroupItem } from './ui/radio-group';
import { Calendar } from './ui/calendar';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';
import { CalendarIcon, Download, FileText, FileJson } from 'lucide-react';
import { format } from 'date-fns';
import { ja } from 'date-fns/locale';
import type { DateRange } from 'react-day-picker';
import type { Entry } from '../types';
import { convertEntriesToExportData, downloadAsCSV, downloadAsJSON, generateExportFilename } from '../utils/export';
import { getTodayRange, getThisWeekRange, getThisMonthRange } from '../utils/time';

interface ExportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  entries: Entry[];
}

type ExportPeriod = 'all' | 'today' | 'week' | 'month' | 'custom';
type ExportFormat = 'csv' | 'json';

export function ExportDialog({ open, onOpenChange, entries }: ExportDialogProps) {
  const [period, setPeriod] = useState<ExportPeriod>('today');
  const [fileFormat, setFileFormat] = useState<ExportFormat>('csv');
  const [dateRange, setDateRange] = useState<DateRange | undefined>();
  const [isDatePickerOpen, setIsDatePickerOpen] = useState(false);

  const getFilteredEntries = (): Entry[] => {
    let startDate: Date | undefined;
    let endDate: Date | undefined;

    switch (period) {
      case 'today':
        ({ start: startDate, end: endDate } = getTodayRange());
        break;
      case 'week':
        ({ start: startDate, end: endDate } = getThisWeekRange());
        break;
      case 'month':
        ({ start: startDate, end: endDate } = getThisMonthRange());
        break;
      case 'custom':
        startDate = dateRange?.from;
        endDate = dateRange?.to;
        break;
      case 'all':
      default:
        return entries.filter(entry => entry.endedAt); // 完了したエントリのみ
    }

    if (!startDate) return entries.filter(entry => entry.endedAt);

    return entries.filter(entry => {
      if (!entry.endedAt) return false;
      const entryDate = entry.startedAt;
      if (!endDate) return entryDate >= startDate;
      return entryDate >= startDate && entryDate < endDate;
    });
  };

  const handleExport = () => {
    const filteredEntries = getFilteredEntries();
    
    if (filteredEntries.length === 0) {
      alert('エクスポートするデータがありません。');
      return;
    }

    const filename = generateExportFilename(dateRange?.from, dateRange?.to);

    if (fileFormat === 'csv') {
      const exportData = convertEntriesToExportData(filteredEntries);
      downloadAsCSV(exportData, filename);
    } else {
      downloadAsJSON(filteredEntries, filename.replace('.csv', '.json'));
    }

    onOpenChange(false);
  };

  const filteredCount = getFilteredEntries().length;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Download className="w-5 h-5" />
            データエクスポート
          </DialogTitle>
          <DialogDescription>
            作業記録をCSVファイルとしてダウンロードできます
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* 期間選択 */}
          <div className="space-y-3">
            <Label>期間を選択</Label>
            <RadioGroup value={period} onValueChange={(value) => setPeriod(value as ExportPeriod)}>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="today" id="today" />
                <Label htmlFor="today">今日</Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="week" id="week" />
                <Label htmlFor="week">今週</Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="month" id="month" />
                <Label htmlFor="month">今月</Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="custom" id="custom" />
                <Label htmlFor="custom">カスタム期間</Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="all" id="all" />
                <Label htmlFor="all">全期間</Label>
              </div>
            </RadioGroup>

            {period === 'custom' && (
              <Popover open={isDatePickerOpen} onOpenChange={setIsDatePickerOpen}>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    className="w-full justify-start text-left font-normal"
                  >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {dateRange?.from ? (
                      dateRange.to ? (
                        <>
                          {format(dateRange.from, 'yyyy/MM/dd', { locale: ja })} -{' '}
                          {format(dateRange.to, 'yyyy/MM/dd', { locale: ja })}
                        </>
                      ) : (
                        format(dateRange.from, 'yyyy/MM/dd', { locale: ja })
                      )
                    ) : (
                      '期間を選択してください'
                    )}
                  </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                  <Calendar
                    initialFocus
                    mode="range"
                    defaultMonth={dateRange?.from}
                    selected={dateRange}
                    onSelect={setDateRange}
                    numberOfMonths={2}
                    locale={ja}
                  />
                </PopoverContent>
              </Popover>
            )}
          </div>

          {/* フォーマット選択 */}
          <div className="space-y-3">
            <Label>ファイル形式</Label>
            <RadioGroup value={fileFormat} onValueChange={(value) => setFileFormat(value as ExportFormat)}>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="csv" id="csv" />
                <Label htmlFor="csv" className="flex items-center gap-2">
                  <FileText className="w-4 h-4" />
                  CSV（Excel対応）
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="json" id="json" />
                <Label htmlFor="json" className="flex items-center gap-2">
                  <FileJson className="w-4 h-4" />
                  JSON（データバックアップ）
                </Label>
              </div>
            </RadioGroup>
          </div>

          {/* プレビュー情報 */}
          <div className="p-3 bg-muted rounded-lg">
            <p className="text-sm text-muted-foreground">
              エクスポート対象: <span className="font-medium text-foreground">{filteredCount}件</span>の完了したエントリ
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            キャンセル
          </Button>
          <Button onClick={handleExport} disabled={filteredCount === 0}>
            <Download className="w-4 h-4 mr-2" />
            エクスポート
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
