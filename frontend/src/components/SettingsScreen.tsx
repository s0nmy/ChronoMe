import React, { useState } from 'react';
import { Button } from './ui/button';
import { Card, CardContent, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { User, Mail, LogOut, Download, Trash2 } from 'lucide-react';
import type { User as UserType, Entry } from '../types';
import { formatDuration } from '../utils/time';

interface SettingsScreenProps {
  user: UserType;
  entries: Entry[];
  onLogout: () => Promise<void>;
  onExportData: () => void;
  onDeleteAllData: () => Promise<void>;
}

export function SettingsScreen({ 
  user, 
  entries,
  onLogout, 
  onExportData, 
  onDeleteAllData 
}: SettingsScreenProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  const [isLoggingOut, setIsLoggingOut] = useState(false);
  
  const persistedEntries = entries.filter(entry => entry.endedAt);
  const totalDuration = persistedEntries.reduce((sum, entry) => sum + entry.durationSec, 0);
  const totalEntries = persistedEntries.length;

  const handleDeleteAllData = async () => {
    if (!window.confirm('すべてのデータを削除しますか？この操作は元に戻せません。')) {
      return;
    }
    setIsDeleting(true);
    try {
      await onDeleteAllData();
    } catch (error) {
      alert(
        error instanceof Error
          ? error.message
          : 'データの削除に失敗しました。時間をおいて再度お試しください。',
      );
    } finally {
      setIsDeleting(false);
    }
  };

  const handleLogout = async () => {
    setIsLoggingOut(true);
    try {
      await onLogout();
    } catch (error) {
      alert(
        error instanceof Error
          ? error.message
          : 'ログアウトに失敗しました。時間をおいて再実行してください。',
      );
    } finally {
      setIsLoggingOut(false);
    }
  };

  return (
    <div className="px-4 lg:px-8 pb-16">
      <div className="w-full max-w-6xl mx-auto space-y-6">
        <div>
          <h1 className="text-2xl font-semibold">設定</h1>
          <p className="text-sm text-muted-foreground">
            アカウント設定とデータ管理
          </p>
        </div>

        <div className="space-y-6">
          {/* アカウント情報 */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <User className="w-5 h-5" />
                アカウント情報
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 bg-primary rounded-full flex items-center justify-center">
                  <User className="w-6 h-6 text-primary-foreground" />
                </div>
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <Mail className="w-4 h-4 text-muted-foreground" />
                    <span className="text-sm">{user.email}</span>
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">
                    登録日: {user.createdAt.toLocaleDateString('ja-JP')}
                  </p>
                </div>
              </div>
                            
              <div className="grid grid-cols-2 gap-4 text-center">
                <div>
                  <div className="text-2xl font-bold text-primary">
                    {totalEntries}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    総エントリ数
                  </div>
                </div>
                <div>
                  <div className="text-2xl font-bold text-primary">
                    {formatDuration(totalDuration)}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    総作業時間
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* データ管理 */}
          <Card>
            <CardHeader>
              <CardTitle>データ管理</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex flex-col sm:flex-row gap-4">
                <Button 
                  variant="outline" 
                  onClick={onExportData}
                  className="flex-1 justify-start"
                >
                  <Download className="w-4 h-4 mr-2" />
                  データをエクスポート
                </Button>
                
                <Button 
                  variant="outline" 
                  onClick={handleDeleteAllData}
                  className="flex-1 justify-start text-destructive hover:text-destructive"
                  disabled={isDeleting}
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  {isDeleting ? '削除中…' : 'すべてのデータを削除'}
                </Button>
              </div>
            </CardContent>
          </Card>

          {/* アプリ情報 */}
          <Card>
            <CardHeader>
              <CardTitle>アプリ情報</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex flex-col sm:flex-row gap-4">
                <div className="flex-1 flex justify-between items-center">
                  <span className="text-sm">バージョン</span>
                  <Badge variant="secondary">1.0.0</Badge>
                </div>
                <div className="flex-1 flex justify-between items-center">
                  <span className="text-sm">最終更新</span>
                  <span className="text-sm text-muted-foreground">
                    2025年11月13日
                  </span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* ログアウト */}
          <Card>
            <CardContent className="pt-6">
              <Button 
                variant="destructive" 
                onClick={handleLogout}
                className="w-full"
                disabled={isLoggingOut}
              >
                <LogOut className="w-4 h-4 mr-2" />
                {isLoggingOut ? 'ログアウト中…' : 'ログアウト'}
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
