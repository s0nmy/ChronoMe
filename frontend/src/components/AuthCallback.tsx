import { useEffect } from 'react';
import { supabase } from '../lib/supabase';

export function AuthCallback() {
  useEffect(() => {
    void supabase.auth.getSession().finally(() => {
      window.history.replaceState({}, document.title, '/');
      window.location.assign('/');
    });
  }, []);

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <p className="text-sm text-muted-foreground">ログイン処理を完了しています…</p>
    </div>
  );
}
