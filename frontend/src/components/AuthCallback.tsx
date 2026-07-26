import { useEffect, useState } from 'react';
import { supabase } from '../lib/supabase';

export function AuthCallback() {
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void supabase.auth.getSession().then(({ data, error: sessionError }) => {
      if (sessionError || !data.session) {
        setError(sessionError?.message ?? '認証セッションを確認できませんでした。');
        return;
      }
      window.history.replaceState({}, document.title, '/');
      window.location.assign('/');
    });
  }, []);

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <div className="space-y-3 text-center">
        <p className="text-sm text-muted-foreground">
          {error ?? 'ログイン処理を完了しています…'}
        </p>
        {error && (
          <a className="text-sm text-primary underline" href="/">
            ログイン画面へ戻る
          </a>
        )}
      </div>
    </div>
  );
}
