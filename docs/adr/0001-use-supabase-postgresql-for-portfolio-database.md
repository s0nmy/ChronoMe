# ADR 0001: ポートフォリオ用途のDBを Cloud SQL から Supabase PostgreSQL へ移行する

## ステータス

採用済み

## 日付

2026-07-14

## 背景

ChronoMe はデモ利用されることを想定している。学生や採用担当者が触る可能性があるため、データ永続化と一定の安定性は必要だが、現時点では商用本番の高可用性や大規模トラフィックは主要要件ではない。

移行前の構成は Cloud Run + Cloud SQL for PostgreSQL だった。Cloud SQL は GCP 内で完結し、Cloud Run との統合も容易だが、現在の利用規模に対して月額約6,000円の継続コストが発生していた。特に Cloud SQL インスタンス、SSD ストレージ、バックアップ、PITR 用ログが固定費として効いていた。

アプリケーション側は `DB_DRIVER=postgres` と `DB_DSN` で接続先を切り替えられるため、PostgreSQL 互換の外部DBへ移行してもアプリケーションコードの変更は最小限にできる。

また、将来的には独自認証から Google / GitHub などの SSO へ移行したい。DB 移行だけでなく、将来の認証基盤との統合しやすさも選定要素になった。

## 決定

ChronoMe のデータベースを GCP Cloud SQL for PostgreSQL から Supabase PostgreSQL へ移行する。

Cloud Run 上のアプリケーション実行構成は維持し、DB 接続先のみ Supabase の Transaction pooler に切り替える。

採用後の主要構成:

- アプリケーション実行基盤: Cloud Run
- サービス名: `chronome`
- GCP Project ID: `chronome-488908`
- Region: `asia-northeast1`
- DB: Supabase PostgreSQL
- DB 接続方式: Supabase Transaction pooler
- DB 設定:
  - `DB_DRIVER=postgres`
  - `DB_DSN=postgresql://...@aws-0-ap-northeast-1.pooler.supabase.com:6543/postgres`
- 認証: 当面は既存の独自認証を継続
- 将来方針: Supabase Auth と SSO への段階的移行を検討する

Cloud Run の Cloud SQL 接続設定は削除し、Cloud SQL 固有の Unix socket 接続から、TLS を前提とした PostgreSQL 接続文字列へ変更する。

## 検討した選択肢

### Cloud SQL を継続してバックアップ設定だけ最適化する

PITR 無効化、バックアップ保持期間短縮、ログ保持期間短縮により、月額コストを削減できる。ただし Cloud SQL インスタンス自体の固定費は残るため、デモ用途としてはまだコストが重い。

### Cloud SQL のインスタンスサイズを下げる

db-g1-small から db-f1-micro などへ縮小すればインスタンス料金は下がる。ただしメモリや接続数に余裕がなくなり、採用担当者が触るデモ環境として体感品質を落とすリスクがある。

### SQLite を Cloud Run で使う

コストは抑えられるが、Cloud Run のコンテナ再起動でローカルデータが失われる。デモ用途でも「触ったデータが消えない」ことは重要なため、採用しない。

### Neon PostgreSQL を使う

PostgreSQL 互換で、DB 専用サービスとして性能面やシンプルさに強みがある。無料枠も ChronoMe の規模には十分だった。一方で、将来的な SSO や BaaS 機能との統合を考えると、DB 専用の Neon より Supabase のほうが全体構成を単純にできる。

### Supabase PostgreSQL を使う

PostgreSQL 互換で、Cloud SQL からの移行コストが低い。無料枠を使えば現在のポートフォリオ用途ではコストを大幅に削減できる。さらに Supabase Auth、RLS、Storage、Realtime などに後から拡張でき、将来の SSO 移行と相性がよい。

## 採用理由

Supabase を採用した主な理由は次の通り。

- Cloud SQL の月額約6,000円の固定費を削減できる。
- PostgreSQL 互換のため、既存の Go バックエンドのDBアクセス設計を大きく変えずに移行できる。
- Cloud Run のアプリケーションホスティング構成を維持できる。
- Transaction pooler により、Cloud Run のようなサーバーレス環境でDB接続数を抑えやすい。
- Supabase Auth による将来の SSO 移行パスを作れる。
- デモとして、BaaS、コスト最適化、段階的移行の説明がしやすい。
- Supabase は知名度とコミュニティが大きく、学習・運用時の情報を得やすい。

## 結果

### 良い影響

- Cloud SQL の固定費を削減できる。
- アプリケーションコードの変更を最小限に抑えられる。
- Cloud Run + PostgreSQL という基本構成は維持できる。
- 将来的に Supabase Auth / SSO / RLS へ拡張しやすくなる。
- データベースを外部マネージドサービスに移すことで、デモ用途に対してコストと運用負荷のバランスがよくなる。

### 悪い影響・トレードオフ

- Cloud Run と DB が GCP 内部で完結しなくなる。
- Cloud Run から Supabase への外部ネットワーク接続になり、レイテンシや外部サービス依存が増える。
- 障害調査時に GCP と Supabase の両方を見る必要がある。
- Supabase Free プランの容量、接続、停止、バックアップ、サポート制約を受ける。
- 接続文字列にパスワードを含むため、GitHub Secrets や Cloud Run 環境変数での管理を徹底する必要がある。
- URL 形式の DSN では、パスワード内の記号を URL エンコードする必要がある。

## 運用上の注意

Cloud SQL は移行直後に削除しない。Supabase 構成で数日間主要操作とログを確認し、問題がないことを確認してから停止する。削除はバックアップ保持方針を決めた後に実施する。

移行後に確認する項目:

- Cloud Run の `DB_DRIVER` が `postgres` であること。
- Cloud Run の `DB_DSN` が Supabase Transaction pooler を向いていること。
- Cloud Run の Cloud SQL 接続アノテーションが削除されていること。
- `/cloudsql/` や `chronome_user` など旧 Cloud SQL 用の設定が残っていないこと。
- ログイン、プロジェクト一覧、時間記録の作成・更新・削除が動作すること。
- Cloud Run ログにDB接続エラーが出ていないこと。
- Supabase 側の主要テーブル件数が移行前後で一致していること。

## 再検討条件

次の条件を満たす場合は、この決定を再検討する。

- Supabase Free プランの制限を超える。
- デモではなく、商用本番として高可用性やSLAが必要になる。
- DB レイテンシがユーザー体験に影響するほど大きくなる。
- Supabase 障害や接続制限が運用上の問題になる。
- GCP にインフラを統一することが要件になる。
- 独自認証から Supabase Auth 以外の認証基盤へ移行することが決まる。
