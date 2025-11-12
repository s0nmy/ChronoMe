/**
 * 時間処理ユーティリティ
 * UTC保存、ローカル表示を基本とする
 */

/**
 * 秒数を「X時間Y分」形式にフォーマット
 */
export function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  
  if (hours === 0) {
    return `${minutes}分`;
  }
  
  if (minutes === 0) {
    return `${hours}時間`;
  }
  
  return `${hours}時間${minutes}分`;
}

/**
 * 秒数を「HH:MM:SS」形式にフォーマット
 */
export function formatTimeElapsed(seconds: number): string {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remainingSeconds = seconds % 60;
  
  return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${remainingSeconds.toString().padStart(2, '0')}`;
}

/**
 * DateをローカルのHH:MM形式にフォーマット
 */
export function formatTimeShort(date: Date): string {
  return date.toLocaleTimeString('ja-JP', {
    hour: '2-digit',
    minute: '2-digit',
    timeZone: 'Asia/Tokyo'
  });
}

/**
 * DateをローカルのYYYY/MM/DD形式にフォーマット
 */
export function formatDateShort(date: Date): string {
  return date.toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    timeZone: 'Asia/Tokyo'
  });
}

/**
 * DateをローカルのYYYY/MM/DD HH:MM形式にフォーマット
 */
export function formatDateTime(date: Date): string {
  return date.toLocaleString('ja-JP', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    timeZone: 'Asia/Tokyo'
  });
}

/**
 * 今日の日付範囲を取得（ローカル時間基準）
 */
export function getTodayRange(): { start: Date; end: Date } {
  const now = new Date();
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const end = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1);
  return { start, end };
}

/**
 * 今週の日付範囲を取得（月曜日開始）
 */
export function getThisWeekRange(): { start: Date; end: Date } {
  const now = new Date();
  const currentDay = now.getDay(); // 0 = Sunday, 1 = Monday, etc.
  const mondayOffset = currentDay === 0 ? -6 : 1 - currentDay;
  
  const start = new Date(now);
  start.setDate(now.getDate() + mondayOffset);
  start.setHours(0, 0, 0, 0);
  
  const end = new Date(start);
  end.setDate(start.getDate() + 7);
  
  return { start, end };
}

/**
 * 今月の日付範囲を取得
 */
export function getThisMonthRange(): { start: Date; end: Date } {
  const now = new Date();
  const start = new Date(now.getFullYear(), now.getMonth(), 1);
  const end = new Date(now.getFullYear(), now.getMonth() + 1, 1);
  return { start, end };
}

/**
 * 2つの日付が同じ日かチェック（ローカル時間基準）
 */
export function isSameDay(date1: Date, date2: Date): boolean {
  return (
    date1.getFullYear() === date2.getFullYear() &&
    date1.getMonth() === date2.getMonth() &&
    date1.getDate() === date2.getDate()
  );
}

/**
 * エントリの経過時間を計算（秒）
 */
export function calculateElapsedTime(startTime: Date, endTime?: Date): number {
  const end = endTime || new Date();
  return Math.floor((end.getTime() - startTime.getTime()) / 1000);
}

/**
 * ULIDライクなIDを生成（簡易版）
 */
export function generateId(): string {
  const timestamp = Date.now().toString(36);
  const randomPart = Math.random().toString(36).substring(2, 15);
  return `${timestamp}_${randomPart}`;
}
